import { createHash, createHmac, randomUUID } from "node:crypto";
import { createServer } from "node:http";
import net from "node:net";
import tls from "node:tls";
import { createRealtimeConfig } from "./runtime.mjs";

function writeJson(res, statusCode, payload) {
  res.writeHead(statusCode, {
    "Content-Type": "application/json; charset=utf-8",
    "X-Content-Type-Options": "nosniff"
  });
  res.end(JSON.stringify(payload));
}

function readJson(req, maxBytes = 128 * 1024) {
  return new Promise((resolve, reject) => {
    let body = "";
    req.setEncoding("utf8");
    req.on("data", (chunk) => {
      body += chunk;
      if (body.length > maxBytes) {
        reject(new Error("request body too large"));
        req.destroy();
      }
    });
    req.on("end", () => {
      if (!body.trim()) {
        resolve({});
        return;
      }
      try {
        resolve(JSON.parse(body));
      } catch {
        reject(new Error("invalid json body"));
      }
    });
    req.on("error", reject);
  });
}

function bearerToken(value = "") {
  const text = String(value || "").trim();
  return text.startsWith("Bearer ") ? text.slice(7).trim() : text;
}

const knownRealtimeRoles = new Set([
  "admin",
  "dispatch_admin",
  "finance_admin",
  "merchant",
  "ops_admin",
  "rider",
  "security_auditor",
  "station_manager",
  "super_admin",
  "support_admin",
  "user"
]);

function authorizeInternalPublish(req, token) {
  if (!token) return true;
  return bearerToken(req.headers.authorization) === token;
}

function base64UrlEncode(buffer) {
  return Buffer.from(buffer)
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

function base64UrlDecode(value = "") {
  const normalized = String(value).replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(normalized.length + (4 - normalized.length % 4) % 4, "=");
  return Buffer.from(padded, "base64");
}

function hmacRealtimeTokenPayload(payload, secret) {
  return base64UrlEncode(createHmac("sha256", secret).update(payload).digest());
}

function verifySignedRealtimeToken(token, secret, nowSeconds = Math.floor(Date.now() / 1000)) {
  const value = String(token || "").trim();
  const tokenSecret = String(secret || "").trim();
  if (!value || !tokenSecret) return null;
  const parts = value.split(".");
  if (parts.length !== 2) return null;
  const expected = hmacRealtimeTokenPayload(parts[0], tokenSecret);
  if (expected !== parts[1]) return null;
  let claims;
  try {
    claims = JSON.parse(base64UrlDecode(parts[0]).toString("utf8"));
  } catch {
    return null;
  }
  const id = String(claims.sub || "").trim();
  const role = String(claims.role || "").trim();
  const exp = Number(claims.exp || 0);
  if (!id || !knownRealtimeRoles.has(role) || nowSeconds >= exp) return null;
  return { id, role, session_id: String(claims.sid || "").trim() };
}

function verifyDevRealtimeToken(token, allowDevTokens) {
  if (!allowDevTokens) return null;
  const parts = String(token || "").trim().split(":");
  if (parts.length !== 2) return null;
  const role = parts[0].trim();
  const id = parts[1].trim();
  if (!id || !knownRealtimeRoles.has(role)) return null;
  return { id, role, session_id: "" };
}

function realtimeAccessToken(req, url) {
  const headerToken = bearerToken(req.headers.authorization);
  if (headerToken) return headerToken;
  return String(url.searchParams.get("access_token") || url.searchParams.get("token") || "").trim();
}

function canUseRequestedRealtimeUser(principal, requestedUserID) {
  const userID = String(requestedUserID || "").trim();
  if (!userID) return true;
  if (!principal?.id) return false;
  if (principal.role === "admin" || principal.role === "super_admin" || principal.role === "support_admin") return true;
  return principal.id === userID;
}

function websocketStatusText(statusCode) {
  if (statusCode === 401) return "Unauthorized";
  if (statusCode === 403) return "Forbidden";
  if (statusCode === 503) return "Service Unavailable";
  return "Bad Request";
}

function createHTTPMembershipAuthorizer(options = {}) {
  const endpoint = String(options.url || "").trim();
  if (!endpoint) return null;
  const token = String(options.token || "").trim();
  return async function authorizeHTTPMembership(request = {}) {
    const principal = request.principal || {};
    const headers = { "Content-Type": "application/json" };
    if (token) headers.Authorization = `Bearer ${token}`;
    const response = await fetch(endpoint, {
      method: "POST",
      headers,
      body: JSON.stringify({
        thread_id: request.threadID,
        subject_type: principal.role,
        subject_id: principal.id,
        role: principal.role
      })
    });
    if (!response.ok) return false;
    const body = await response.json().catch(() => null);
    return body?.success === true && body?.data?.allowed === true;
  };
}

async function authorizeWebSocketUpgrade(req, url, options = {}) {
  if (!options.required) return { principal: null };
  const token = realtimeAccessToken(req, url);
  const principal = verifySignedRealtimeToken(token, options.secret) || verifyDevRealtimeToken(token, options.allowDevTokens);
  if (!principal) return { statusCode: 401, reason: "invalid websocket token" };
  if (!canUseRequestedRealtimeUser(principal, url.searchParams.get("user_id"))) {
    return { statusCode: 403, reason: "websocket user mismatch" };
  }
  const threadID = String(url.searchParams.get("thread_id") || "").trim();
  if (threadID && options.membershipAuthorizer) {
    let allowed = false;
    try {
      allowed = await options.membershipAuthorizer({
        threadID,
        principal,
        requestedUserID: String(url.searchParams.get("user_id") || "").trim()
      });
    } catch {
      return { statusCode: 503, reason: "websocket membership authorization unavailable" };
    }
    if (!allowed) return { statusCode: 403, reason: "websocket thread membership denied" };
  }
  return { principal };
}

function normalizeRealtimeMessage(message = {}) {
  const payload = message.payload && typeof message.payload === "object" ? message.payload : {};
  const event = message.event && typeof message.event === "object" ? message.event : {};
  return {
    topic: String(message.topic || event.topic || "").trim(),
    key: String(message.key || event.key || "").trim(),
    payload,
    event,
    delivered_at: new Date().toISOString()
  };
}

function normalizeClusterEnvelope(envelope = {}) {
  const data = envelope && typeof envelope === "object" ? envelope : {};
  return {
    source_id: String(data.source_id || "").trim(),
    message: normalizeRealtimeMessage(data.message && typeof data.message === "object" ? data.message : data),
    published_at: String(data.published_at || "").trim()
  };
}

function encodeRedisCommand(args = []) {
  const chunks = [`*${args.length}\r\n`];
  for (const arg of args) {
    const value = String(arg);
    chunks.push(`$${Buffer.byteLength(value)}\r\n${value}\r\n`);
  }
  return Buffer.from(chunks.join(""), "utf8");
}

function parseRespValue(buffer, offset = 0) {
  if (offset >= buffer.length) return null;
  const marker = String.fromCharCode(buffer[offset]);
  const lineEnd = buffer.indexOf("\r\n", offset + 1);
  if (lineEnd === -1) return null;
  const line = buffer.subarray(offset + 1, lineEnd).toString("utf8");
  const nextOffset = lineEnd + 2;
  if (marker === "+") return { value: line, offset: nextOffset };
  if (marker === "-") return { value: new Error(line), offset: nextOffset };
  if (marker === ":") return { value: Number(line), offset: nextOffset };
  if (marker === "$") {
    const length = Number(line);
    if (length < 0) return { value: null, offset: nextOffset };
    const end = nextOffset + length;
    if (buffer.length < end + 2) return null;
    return { value: buffer.subarray(nextOffset, end).toString("utf8"), offset: end + 2 };
  }
  if (marker === "*") {
    const length = Number(line);
    const values = [];
    let cursor = nextOffset;
    for (let index = 0; index < length; index++) {
      const parsed = parseRespValue(buffer, cursor);
      if (!parsed) return null;
      if (parsed.value instanceof Error) throw parsed.value;
      values.push(parsed.value);
      cursor = parsed.offset;
    }
    return { value: values, offset: cursor };
  }
  throw new Error("unsupported redis response");
}

function createRespReader(socket) {
  let buffer = Buffer.alloc(0);
  const waiters = [];
  function flush(error) {
    while (waiters.length > 0) {
      waiters.shift().reject(error);
    }
  }
  function pump() {
    while (waiters.length > 0) {
      let parsed;
      try {
        parsed = parseRespValue(buffer, 0);
      } catch (error) {
        flush(error);
        return;
      }
      if (!parsed) return;
      buffer = buffer.subarray(parsed.offset);
      if (parsed.value instanceof Error) {
        waiters.shift().reject(parsed.value);
      } else {
        waiters.shift().resolve(parsed.value);
      }
    }
  }
  socket.on("data", (chunk) => {
    buffer = Buffer.concat([buffer, chunk]);
    pump();
  });
  socket.on("error", (error) => flush(error));
  socket.on("close", () => flush(new Error("redis connection closed")));
  return {
    next() {
      return new Promise((resolve, reject) => {
        waiters.push({ resolve, reject });
        pump();
      });
    }
  };
}

function redisConnectionOptions(redisUrl) {
  const target = new URL(redisUrl);
  const useTLS = target.protocol === "rediss:";
  const database = String(target.pathname || "").replace(/^\//, "");
  return {
    host: target.hostname || "127.0.0.1",
    port: Number(target.port || 6379),
    tls: useTLS,
    username: decodeURIComponent(target.username || ""),
    password: decodeURIComponent(target.password || ""),
    database
  };
}

function connectRedisSocket(options) {
  return new Promise((resolve, reject) => {
    const socket = options.tls
      ? tls.connect({ host: options.host, port: options.port, servername: options.host })
      : net.connect({ host: options.host, port: options.port });
    socket.setKeepAlive(true);
    socket.once("connect", () => resolve(socket));
    socket.once("error", reject);
  });
}

async function sendRedisCommand(socket, reader, args) {
  socket.write(encodeRedisCommand(args));
  return reader.next();
}

async function prepareRedisConnection(socket, reader, options) {
  if (options.password) {
    if (options.username) {
      await sendRedisCommand(socket, reader, ["AUTH", options.username, options.password]);
    } else {
      await sendRedisCommand(socket, reader, ["AUTH", options.password]);
    }
  }
  if (options.database) {
    await sendRedisCommand(socket, reader, ["SELECT", options.database]);
  }
}

async function redisPublish(redisUrl, channel, payload) {
  const options = redisConnectionOptions(redisUrl);
  const socket = await connectRedisSocket(options);
  const reader = createRespReader(socket);
  try {
    await prepareRedisConnection(socket, reader, options);
    return await sendRedisCommand(socket, reader, ["PUBLISH", channel, payload]);
  } finally {
    socket.end();
  }
}

function createRedisSubscriber(redisUrl, channel, handler, onError) {
  let active = true;
  let socket;
  let reconnectTimer;
  async function connect() {
    if (!active) return;
    try {
      const options = redisConnectionOptions(redisUrl);
      socket = await connectRedisSocket(options);
      const reader = createRespReader(socket);
      await prepareRedisConnection(socket, reader, options);
      socket.write(encodeRedisCommand(["SUBSCRIBE", channel]));
      socket.on("close", () => {
        if (!active) return;
        reconnectTimer = setTimeout(connect, 1000);
      });
      while (active) {
        const message = await reader.next();
        if (!Array.isArray(message) || String(message[0]) !== "message") continue;
        try {
          handler(JSON.parse(String(message[2] || "{}")));
        } catch (error) {
          onError?.(error);
        }
      }
    } catch (error) {
      onError?.(error);
      if (active) {
        reconnectTimer = setTimeout(connect, 1000);
      }
    }
  }
  return {
    start() {
      connect();
    },
    close() {
      active = false;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      if (socket) socket.destroy();
    }
  };
}

export function createRedisRealtimeAdapter(options = {}) {
  const redisUrl = String(options.redisUrl || options.url || "").trim();
  if (!redisUrl) return null;
  let subscriber;
  return {
    mode: "redis-pubsub",
    async publish(channel, envelope) {
      const payload = JSON.stringify(envelope);
      const subscribers = await redisPublish(redisUrl, channel, payload);
      return { subscribers };
    },
    subscribe(channel, handler, onError) {
      subscriber = createRedisSubscriber(redisUrl, channel, handler, onError);
      subscriber.start();
      return subscriber;
    },
    close() {
      subscriber?.close();
    }
  };
}

function encodeWebSocketFrame(text) {
  const payload = Buffer.from(String(text), "utf8");
  if (payload.length < 126) {
    return Buffer.concat([Buffer.from([0x81, payload.length]), payload]);
  }
  if (payload.length <= 0xffff) {
    const header = Buffer.alloc(4);
    header[0] = 0x81;
    header[1] = 126;
    header.writeUInt16BE(payload.length, 2);
    return Buffer.concat([header, payload]);
  }
  const header = Buffer.alloc(10);
  header[0] = 0x81;
  header[1] = 127;
  header.writeBigUInt64BE(BigInt(payload.length), 2);
  return Buffer.concat([header, payload]);
}

function decodeClientFrames(buffer) {
  const frames = [];
  let offset = 0;
  while (buffer.length - offset >= 2) {
    const first = buffer[offset];
    const second = buffer[offset + 1];
    const opcode = first & 0x0f;
    const masked = (second & 0x80) !== 0;
    let length = second & 0x7f;
    let headerLength = 2;
    if (length === 126) {
      if (buffer.length - offset < 4) break;
      length = buffer.readUInt16BE(offset + 2);
      headerLength = 4;
    } else if (length === 127) {
      if (buffer.length - offset < 10) break;
      const bigLength = buffer.readBigUInt64BE(offset + 2);
      if (bigLength > BigInt(Number.MAX_SAFE_INTEGER)) {
        throw new Error("websocket frame too large");
      }
      length = Number(bigLength);
      headerLength = 10;
    }
    const maskLength = masked ? 4 : 0;
    const frameLength = headerLength + maskLength + length;
    if (buffer.length - offset < frameLength) break;
    let payload = buffer.subarray(offset + headerLength + maskLength, offset + frameLength);
    if (masked) {
      const mask = buffer.subarray(offset + headerLength, offset + headerLength + 4);
      payload = Buffer.from(payload.map((byte, index) => byte ^ mask[index % 4]));
    }
    frames.push({ opcode, payload });
    offset += frameLength;
  }
  return { frames, remaining: buffer.subarray(offset) };
}

function sendWebSocketJson(client, message) {
  if (client.socket.destroyed) return false;
  client.socket.write(encodeWebSocketFrame(JSON.stringify(message)));
  return true;
}

function shouldDeliver(client, message) {
  if (message.topic === "message.sent") {
    const threadID = String(message.payload.thread_id || "").trim();
    return !client.threadID || client.threadID === threadID;
  }
  return true;
}

function handleClientFrame(client, frame) {
  if (frame.opcode === 0x8) {
    client.socket.end();
    return;
  }
  if (frame.opcode === 0x9) {
    client.socket.write(Buffer.concat([Buffer.from([0x8a, frame.payload.length]), frame.payload]));
    return;
  }
  if (frame.opcode !== 0x1) return;
  let data;
  try {
    data = JSON.parse(frame.payload.toString("utf8"));
  } catch {
    return;
  }
  if (data.type === "subscribe" && data.thread_id) {
    client.threadID = String(data.thread_id).trim();
  }
}

export function publishRealtimeMessage(clients, rawMessage = {}) {
  const message = normalizeRealtimeMessage(rawMessage);
  if (!message.topic) {
    throw new Error("realtime topic is required");
  }
  return publishNormalizedRealtimeMessage(clients, message);
}

function publishNormalizedRealtimeMessage(clients, message) {
  let delivered = 0;
  for (const client of clients.values()) {
    if (!shouldDeliver(client, message)) continue;
    if (sendWebSocketJson(client, message)) {
      delivered++;
    }
  }
  return { ...message, delivered };
}

export function createRealtimeServer(options = {}) {
  const env = options.env || process.env;
  const config = createRealtimeConfig(env);
  const internalToken = String(options.internalToken || env.REALTIME_INTERNAL_TOKEN || "").trim();
  const websocketAuth = {
    required: options.websocketAuthRequired ?? config.websocketAuthRequired,
    allowDevTokens: options.websocketDevTokensAllowed ?? config.websocketDevTokensAllowed,
    secret: String(options.authTokenSecret || env.REALTIME_AUTH_TOKEN_SECRET || env.AUTH_TOKEN_SECRET || "infinitech-dev-secret-change-me").trim(),
    membershipAuthorizer: options.membershipAuthorizer === undefined
      ? createHTTPMembershipAuthorizer({
        url: options.membershipAuthURL || config.membershipAuthUrl,
        token: internalToken
      })
      : options.membershipAuthorizer
  };
  const instanceID = String(options.instanceID || config.instanceID || randomUUID()).trim();
  const clusterChannel = String(options.clusterChannel || config.redisChannel || "infinitech:realtime:events").trim();
  const clusterAdapter = options.clusterAdapter === undefined
    ? createRedisRealtimeAdapter({ redisUrl: env.REALTIME_REDIS_URL })
    : options.clusterAdapter;
  const clusterState = {
    adapter: clusterAdapter?.mode || config.redisAdapter || "disabled",
    channel: clusterChannel,
    instance_id: instanceID,
    published: 0,
    received: 0,
    last_error: ""
  };
  const clients = new Map();
  function recordClusterError(error) {
    clusterState.last_error = error?.message || String(error || "");
  }
  function deliverClusterEnvelope(envelope) {
    try {
      const normalized = normalizeClusterEnvelope(envelope);
      if (!normalized.message.topic || normalized.source_id === instanceID) return;
      publishNormalizedRealtimeMessage(clients, normalized.message);
      clusterState.received++;
    } catch (error) {
      recordClusterError(error);
    }
  }
  if (clusterAdapter?.subscribe) {
    try {
      clusterAdapter.subscribe(clusterChannel, deliverClusterEnvelope, recordClusterError);
    } catch (error) {
      recordClusterError(error);
    }
  }
  async function publishWithCluster(rawMessage = {}) {
    const message = normalizeRealtimeMessage(rawMessage);
    if (!message.topic) {
      throw new Error("realtime topic is required");
    }
    const local = publishNormalizedRealtimeMessage(clients, message);
    const cluster = {
      adapter: clusterState.adapter,
      channel: clusterChannel,
      published: false,
      subscribers: 0
    };
    if (clusterAdapter?.publish) {
      try {
        const result = await clusterAdapter.publish(clusterChannel, {
          source_id: instanceID,
          published_at: new Date().toISOString(),
          message
        });
        cluster.published = true;
        cluster.subscribers = Number(result?.subscribers || 0);
        clusterState.published++;
      } catch (error) {
        recordClusterError(error);
        throw error;
      }
    }
    return { ...local, cluster };
  }
  const server = createServer(async (req, res) => {
    const url = new URL(req.url || "/", `http://${req.headers.host || "127.0.0.1"}`);
    if (req.method === "GET" && url.pathname === "/healthz") {
      writeJson(res, 200, { success: true, message: "ok", data: { status: "ok", service: "realtime-gateway" } });
      return;
    }
    if (req.method === "GET" && url.pathname === "/readyz") {
      writeJson(res, 200, { success: true, message: "ok", data: { ...config, instanceID, connections: clients.size, cluster: clusterState } });
      return;
    }
    if (req.method === "POST" && url.pathname === config.internalPublishPath) {
      if (!authorizeInternalPublish(req, internalToken)) {
        writeJson(res, 401, { success: false, code: "UNAUTHORIZED", message: "invalid realtime publish token" });
        return;
      }
      try {
        const message = await readJson(req);
        const delivered = await publishWithCluster(message);
        writeJson(res, 200, { success: true, message: "ok", data: { connections: clients.size, ...delivered } });
      } catch (error) {
        writeJson(res, 400, { success: false, code: "BAD_REALTIME_EVENT", message: error.message });
      }
      return;
    }
    writeJson(res, 426, { success: false, code: "WEBSOCKET_REQUIRED", message: "realtime gateway requires websocket upgrade" });
  });

  server.on("upgrade", async (req, socket) => {
    const url = new URL(req.url || "/", `http://${req.headers.host || "127.0.0.1"}`);
    if (url.pathname !== "/ws" || String(req.headers.upgrade || "").toLowerCase() !== "websocket") {
      socket.end("HTTP/1.1 404 Not Found\r\n\r\n");
      return;
    }
    const key = req.headers["sec-websocket-key"];
    if (!key) {
      socket.end("HTTP/1.1 400 Bad Request\r\n\r\n");
      return;
    }
    const auth = await authorizeWebSocketUpgrade(req, url, websocketAuth);
    if (auth.statusCode) {
      socket.end(`HTTP/1.1 ${auth.statusCode} ${websocketStatusText(auth.statusCode)}\r\n\r\n${auth.reason || ""}`);
      return;
    }
    const accept = createHash("sha1")
      .update(`${key}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
      .digest("base64");
    socket.write([
      "HTTP/1.1 101 Switching Protocols",
      "Upgrade: websocket",
      "Connection: Upgrade",
      `Sec-WebSocket-Accept: ${accept}`,
      "\r\n"
    ].join("\r\n"));
    const client = {
      id: randomUUID(),
      socket,
      userID: auth.principal?.id || String(url.searchParams.get("user_id") || "").trim(),
      role: auth.principal?.role || "",
      threadID: String(url.searchParams.get("thread_id") || "").trim(),
      buffer: Buffer.alloc(0)
    };
    clients.set(client.id, client);
    socket.on("data", (chunk) => {
      try {
        const decoded = decodeClientFrames(Buffer.concat([client.buffer, chunk]));
        client.buffer = decoded.remaining;
        decoded.frames.forEach((frame) => handleClientFrame(client, frame));
      } catch {
        socket.destroy();
      }
    });
    socket.on("close", () => clients.delete(client.id));
    socket.on("error", () => clients.delete(client.id));
  });

  server.realtime = {
    clients,
    publish(message) {
      return publishWithCluster(message);
    }
  };
  server.publishRealtimeMessage = server.realtime.publish;
  server.on("close", () => {
    for (const client of clients.values()) {
      client.socket.destroy();
    }
    clients.clear();
    clusterAdapter?.close?.();
  });

  return server;
}
