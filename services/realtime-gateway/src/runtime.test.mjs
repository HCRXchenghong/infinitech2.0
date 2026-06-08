import assert from "node:assert/strict";
import { createHmac } from "node:crypto";
import { once } from "node:events";
import http from "node:http";
import net from "node:net";
import test from "node:test";
import { createRealtimeConfig } from "./runtime.mjs";
import { createRealtimeServer } from "./server.mjs";

test("realtime config is websocket only and capacity gated", () => {
  const config = createRealtimeConfig();
  assert.equal(config.transportMode, "websocket-only");
  assert.equal(config.websocketAuthRequired, false);
  assert.equal(config.membershipAuthConfigured, false);
  assert.equal(config.redisAdapter, "disabled");
  assert.deepEqual(config.capacityGate.requiredStages, ["10k", "30k", "60k", "100k"]);
  assert.equal(config.capacityGate.claimAllowedWithoutReports, false);
  assert.ok(config.namespaces.includes("/rtc"));
});

test("websocket upgrade requires a signed user token when auth is enabled", async () => {
  const server = createRealtimeServer({
    env: {
      REALTIME_WS_AUTH_REQUIRED: "true",
      AUTH_TOKEN_SECRET: "unit-secret"
    }
  });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const unauthorized = await openRawWebSocketUpgrade(port, "/ws?thread_id=merchant_blue_sea&user_id=user_1");
  const token = issueRealtimeToken({ sub: "user_1", role: "user" }, "unit-secret");
  const authorized = await openWebSocket(port, "/ws?thread_id=merchant_blue_sea&user_id=user_1", {
    Authorization: `Bearer ${token}`
  });

  assert.match(unauthorized, /401 Unauthorized/);
  assert.equal(authorized.messages.length, 0);

  authorized.socket.destroy();
  server.close();
});

test("websocket upgrade rejects user id spoofing when auth is enabled", async () => {
  const server = createRealtimeServer({
    env: {
      REALTIME_WS_AUTH_REQUIRED: "true",
      AUTH_TOKEN_SECRET: "unit-secret"
    }
  });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const token = issueRealtimeToken({ sub: "user_1", role: "user" }, "unit-secret");
  const response = await openRawWebSocketUpgrade(port, "/ws?thread_id=merchant_blue_sea&user_id=user_2", {
    Authorization: `Bearer ${token}`
  });

  server.close();
  assert.match(response, /403 Forbidden/);
});

test("websocket upgrade asks backend membership before joining a thread", async () => {
  const membershipCalls = [];
  const server = createRealtimeServer({
    env: {
      REALTIME_WS_AUTH_REQUIRED: "true",
      AUTH_TOKEN_SECRET: "unit-secret"
    },
    membershipAuthorizer: async (request) => {
      membershipCalls.push(request);
      return request.threadID === "merchant_blue_sea" && request.principal.id === "user_1";
    }
  });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const token = issueRealtimeToken({ sub: "user_1", role: "user" }, "unit-secret");
  const rejected = await openRawWebSocketUpgrade(port, "/ws?thread_id=rider_zhang&user_id=user_1", {
    Authorization: `Bearer ${token}`
  });
  const accepted = await openWebSocket(port, "/ws?thread_id=merchant_blue_sea&user_id=user_1", {
    Authorization: `Bearer ${token}`
  });

  assert.match(rejected, /403 Forbidden/);
  assert.equal(membershipCalls.length, 2);
  assert.deepEqual(membershipCalls.map((call) => call.threadID), ["rider_zhang", "merchant_blue_sea"]);

  accepted.socket.destroy();
  server.close();
});

test("realtime config enables redis pubsub adapter when configured", () => {
  const config = createRealtimeConfig({
    REALTIME_MEMBERSHIP_AUTH_URL: "http://api-go:1029/internal/realtime/authorize",
    REALTIME_REDIS_URL: "redis://redis:6379/0",
    REALTIME_REDIS_CHANNEL: "infinitech:test:realtime"
  });
  assert.equal(config.membershipAuthConfigured, true);
  assert.equal(config.redisAdapter, "redis-pubsub");
  assert.equal(config.redisConfigured, true);
  assert.equal(config.redisChannel, "infinitech:test:realtime");
});

test("http requests are rejected with websocket required outside health endpoints", async () => {
  const server = createRealtimeServer();
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const statusCode = await getStatus(`http://127.0.0.1:${port}/im`);
  server.close();
  assert.equal(statusCode, 426);
});

test("websocket clients receive message.sent events for their subscribed thread", async () => {
  const server = createRealtimeServer({ env: { REALTIME_INTERNAL_TOKEN: "realtime-secret" } });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const merchantClient = await openWebSocket(port, "/ws?thread_id=merchant_blue_sea&user_id=user_2");
  const officialClient = await openWebSocket(port, "/ws?thread_id=official&user_id=user_2");

  const response = await postJSON(`http://127.0.0.1:${port}/internal/realtime/publish`, {
    topic: "message.sent",
    key: "message.sent:msg_1",
    payload: {
      id: "msg_1",
      thread_id: "merchant_blue_sea",
      sender: "蓝海餐厅",
      content: "今晚群内券可用",
      created_at: "2026-05-27T12:00:00Z"
    }
  }, "Bearer realtime-secret");
  const message = JSON.parse(await merchantClient.nextMessage());

  assert.equal(response.data.delivered, 1);
  assert.equal(message.topic, "message.sent");
  assert.equal(message.payload.thread_id, "merchant_blue_sea");
  assert.equal(message.payload.content, "今晚群内券可用");
  assert.equal(officialClient.messages.length, 0);

  merchantClient.socket.destroy();
  officialClient.socket.destroy();
  server.close();
});

test("redis adapter fanout delivers message.sent across gateway replicas", async () => {
  const bus = createMemoryRealtimeBus();
  const serverA = createRealtimeServer({
    env: { REALTIME_INTERNAL_TOKEN: "realtime-secret" },
    instanceID: "gateway-a",
    clusterChannel: "infinitech:test:realtime",
    clusterAdapter: bus.createAdapter()
  });
  const serverB = createRealtimeServer({
    env: { REALTIME_INTERNAL_TOKEN: "realtime-secret" },
    instanceID: "gateway-b",
    clusterChannel: "infinitech:test:realtime",
    clusterAdapter: bus.createAdapter()
  });
  serverA.listen(0);
  serverB.listen(0);
  await Promise.all([once(serverA, "listening"), once(serverB, "listening")]);
  const portA = serverA.address().port;
  const portB = serverB.address().port;
  const remoteClient = await openWebSocket(portB, "/ws?thread_id=merchant_blue_sea&user_id=user_2");

  const response = await postJSON(`http://127.0.0.1:${portA}/internal/realtime/publish`, {
    topic: "message.sent",
    key: "message.sent:msg_cluster_1",
    payload: {
      id: "msg_cluster_1",
      thread_id: "merchant_blue_sea",
      sender: "蓝海餐厅",
      content: "跨副本群消息",
      created_at: "2026-05-29T12:00:00Z"
    }
  }, "Bearer realtime-secret");
  const message = JSON.parse(await remoteClient.nextMessage());

  assert.equal(response.data.delivered, 0);
  assert.equal(response.data.cluster.published, true);
  assert.equal(response.data.cluster.subscribers, 2);
  assert.equal(message.topic, "message.sent");
  assert.equal(message.payload.thread_id, "merchant_blue_sea");
  assert.equal(message.payload.content, "跨副本群消息");

  remoteClient.socket.destroy();
  serverA.close();
  serverB.close();
});

test("internal realtime publish endpoint requires configured token", async () => {
  const server = createRealtimeServer({ env: { REALTIME_INTERNAL_TOKEN: "realtime-secret" } });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const response = await postJSON(`http://127.0.0.1:${port}/internal/realtime/publish`, { topic: "message.sent" });
  server.close();
  assert.equal(response.statusCode, 401);
  assert.equal(response.code, "UNAUTHORIZED");
});

function getStatus(url) {
  return new Promise((resolve, reject) => {
    http.get(url, (res) => {
      res.resume();
      res.on("end", () => resolve(res.statusCode));
    }).on("error", reject);
  });
}

function createMemoryRealtimeBus() {
  const subscribers = new Map();
  return {
    createAdapter() {
      return {
        mode: "redis-pubsub",
        async publish(channel, envelope) {
          const handlers = subscribers.get(channel) || new Set();
          for (const handler of handlers) {
            handler(envelope);
          }
          return { subscribers: handlers.size };
        },
        subscribe(channel, handler) {
          if (!subscribers.has(channel)) {
            subscribers.set(channel, new Set());
          }
          subscribers.get(channel).add(handler);
        },
        close() {}
      };
    }
  };
}

function postJSON(url, payload, authorization = "") {
  return new Promise((resolve, reject) => {
    const target = new URL(url);
    const body = JSON.stringify(payload);
    const req = http.request({
      hostname: target.hostname,
      port: target.port,
      path: target.pathname + target.search,
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Content-Length": Buffer.byteLength(body),
        ...(authorization ? { Authorization: authorization } : {})
      }
    }, (res) => {
      let text = "";
      res.setEncoding("utf8");
      res.on("data", (chunk) => {
        text += chunk;
      });
      res.on("end", () => {
        resolve({ statusCode: res.statusCode, ...(text ? JSON.parse(text) : {}) });
      });
    });
    req.on("error", reject);
    req.end(body);
  });
}

function openWebSocket(port, path, headers = {}) {
  return new Promise((resolve, reject) => {
    const socket = net.connect(port, "127.0.0.1");
    const messages = [];
    const waiters = [];
    let buffer = Buffer.alloc(0);
    let handshakeDone = false;
    socket.on("connect", () => {
      socket.write([
        `GET ${path} HTTP/1.1`,
        `Host: 127.0.0.1:${port}`,
        "Upgrade: websocket",
        "Connection: Upgrade",
        `Sec-WebSocket-Key: ${Buffer.from("0123456789abcdef").toString("base64")}`,
        "Sec-WebSocket-Version: 13",
        ...Object.entries(headers).map(([key, value]) => `${key}: ${value}`),
        "\r\n"
      ].join("\r\n"));
    });
    socket.on("data", (chunk) => {
      buffer = Buffer.concat([buffer, chunk]);
      if (!handshakeDone) {
        const marker = buffer.indexOf("\r\n\r\n");
        if (marker === -1) return;
        const head = buffer.subarray(0, marker).toString("utf8");
        if (!head.includes("101 Switching Protocols")) {
          reject(new Error(head));
          socket.destroy();
          return;
        }
        handshakeDone = true;
        buffer = buffer.subarray(marker + 4);
        resolve({
          socket,
          messages,
          nextMessage() {
            if (messages.length > 0) return Promise.resolve(messages.shift());
            return new Promise((done) => waiters.push(done));
          }
        });
      }
      const decoded = decodeServerFrames(buffer);
      buffer = decoded.remaining;
      decoded.messages.forEach((message) => {
        if (waiters.length > 0) {
          waiters.shift()(message);
        } else {
          messages.push(message);
        }
      });
    });
    socket.on("error", reject);
  });
}

function openRawWebSocketUpgrade(port, path, headers = {}) {
  return new Promise((resolve, reject) => {
    const socket = net.connect(port, "127.0.0.1");
    let buffer = Buffer.alloc(0);
    socket.on("connect", () => {
      socket.write([
        `GET ${path} HTTP/1.1`,
        `Host: 127.0.0.1:${port}`,
        "Upgrade: websocket",
        "Connection: Upgrade",
        `Sec-WebSocket-Key: ${Buffer.from("0123456789abcdef").toString("base64")}`,
        "Sec-WebSocket-Version: 13",
        ...Object.entries(headers).map(([key, value]) => `${key}: ${value}`),
        "\r\n"
      ].join("\r\n"));
    });
    socket.on("data", (chunk) => {
      buffer = Buffer.concat([buffer, chunk]);
      if (buffer.includes("\r\n\r\n")) {
        resolve(buffer.toString("utf8"));
        socket.destroy();
      }
    });
    socket.on("error", reject);
  });
}

function base64UrlEncode(buffer) {
  return Buffer.from(buffer)
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

function issueRealtimeToken(claims, secret) {
  const payload = base64UrlEncode(Buffer.from(JSON.stringify({
    sid: "sess_test",
    exp: Math.floor(Date.now() / 1000) + 3600,
    ...claims
  }), "utf8"));
  const signature = base64UrlEncode(createHmac("sha256", secret).update(payload).digest());
  return `${payload}.${signature}`;
}

function decodeServerFrames(buffer) {
  const messages = [];
  let offset = 0;
  while (buffer.length - offset >= 2) {
    const opcode = buffer[offset] & 0x0f;
    let length = buffer[offset + 1] & 0x7f;
    let headerLength = 2;
    if (length === 126) {
      if (buffer.length - offset < 4) break;
      length = buffer.readUInt16BE(offset + 2);
      headerLength = 4;
    } else if (length === 127) {
      if (buffer.length - offset < 10) break;
      length = Number(buffer.readBigUInt64BE(offset + 2));
      headerLength = 10;
    }
    if (buffer.length - offset < headerLength + length) break;
    if (opcode === 0x1) {
      messages.push(buffer.subarray(offset + headerLength, offset + headerLength + length).toString("utf8"));
    }
    offset += headerLength + length;
  }
  return { messages, remaining: buffer.subarray(offset) };
}
