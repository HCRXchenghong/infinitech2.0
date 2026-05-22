import assert from "node:assert/strict";
import test from "node:test";
import { createIntegrationConsumer, normalizeProviderConfig, subscribedTopics } from "./index.mjs";

test("integration worker covers oauth and api sync events", () => {
  assert.ok(subscribedTopics.includes("oauth.callback"));
  assert.ok(subscribedTopics.includes("provider.sync.requested"));
});

test("provider config is normalized", () => {
  assert.deepEqual(normalizeProviderConfig({ provider: " WeChat ", enabled: true, scopes: [" snsapi_base ", ""] }), {
    provider: "wechat",
    enabled: true,
    scopes: ["snsapi_base"],
    rate_limit_per_minute: 60
  });
});

test("integration consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createIntegrationConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { synced: true };
    }
  });
  const event = { id: "obe_integration_1", topic: "provider.sync.requested", idempotency_key: "provider:wechat:sync:20260522" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_integration_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "integration-worker");
});
