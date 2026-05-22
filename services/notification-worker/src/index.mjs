import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "notification-worker";
export const subscribedTopics = ["order.status_changed", "message.created", "dispatch.assigned", "rtc.call_status_changed"];

export function buildNotification(event = {}) {
  const type = String(event.type || "").trim();
  const target = event.target || {};
  return {
    type,
    target_role: String(target.role || "").trim(),
    target_id: String(target.id || "").trim(),
    title: event.title || "Infinitech 通知",
    body: event.body || "",
    idempotency_key: `notify:${type}:${event.id || event.order_id || event.message_id || ""}`
  };
}

export function createNotificationConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event = {}) => buildNotification(event.payload || event.notification || event))
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
