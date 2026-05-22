import assert from "node:assert/strict";
import { once } from "node:events";
import http from "node:http";
import test from "node:test";
import { createRealtimeConfig } from "./runtime.mjs";
import { createRealtimeServer } from "./server.mjs";

test("realtime config is websocket only and capacity gated", () => {
  const config = createRealtimeConfig();
  assert.equal(config.transportMode, "websocket-only");
  assert.deepEqual(config.capacityGate.requiredStages, ["10k", "30k", "60k", "100k"]);
  assert.equal(config.capacityGate.claimAllowedWithoutReports, false);
  assert.ok(config.namespaces.includes("/rtc"));
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

function getStatus(url) {
  return new Promise((resolve, reject) => {
    http.get(url, (res) => {
      res.resume();
      res.on("end", () => resolve(res.statusCode));
    }).on("error", reject);
  });
}

