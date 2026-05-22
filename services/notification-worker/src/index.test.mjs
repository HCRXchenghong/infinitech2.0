import assert from "node:assert/strict";
import test from "node:test";
import { buildNotification, createNotificationConsumer, subscribedTopics } from "./index.mjs";

test("notification worker watches order and message events", () => {
  assert.ok(subscribedTopics.includes("order.status_changed"));
  assert.ok(subscribedTopics.includes("message.created"));
});

test("notification payload is idempotent", () => {
  const notification = buildNotification({ type: "order.status_changed", order_id: "ord_1", target: { role: "user", id: "u1" }, body: "已接单" });
  assert.equal(notification.idempotency_key, "notify:order.status_changed:ord_1");
  assert.equal(notification.target_role, "user");
});

test("notification consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createNotificationConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { pushed: true };
    }
  });
  const event = { id: "obe_notify_1", topic: "order.status_changed", idempotency_key: "notify:order.status_changed:ord_1" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_notify_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "notification-worker");
});
