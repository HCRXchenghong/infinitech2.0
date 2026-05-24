import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "notification-worker";
export const subscribedTopics = ["order.status_changed", "message.created", "dispatch.assigned", "rtc.call_status_changed", "audit.retention_alerts"];

export function buildNotification(event = {}) {
  const type = String(event.type || "").trim();
  const payload = event.payload || {};
  if (event.topic === "audit.retention_alerts" || type === "audit.retention_alerts.emitted") {
    const alertCount = Number(payload.alert_count || event.alert_count || 0);
    const criticalCount = Number(payload.critical_count || event.critical_count || 0);
    return {
      type: "audit.retention_alerts.emitted",
      target_role: "security",
      target_id: "audit_retention",
      title: "审计留存告警",
      body: `critical=${criticalCount}; total=${alertCount}`,
      idempotency_key: `notify:audit.retention_alerts:${event.id || event.idempotency_key || payload.idempotency_key || ""}`
    };
  }
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
