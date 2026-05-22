import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "dispatch-worker";
export const subscribedTopics = ["order.paid", "dispatch.timeout", "rider.location_updated", "dispatch.cancel_requested"];
export const grabHallSeconds = 600;

export function resolveDispatchMode(order, now = new Date()) {
  const createdAt = Date.parse(order?.created_at || order?.createdAt || "");
  if (!Number.isFinite(createdAt)) return "grab_hall";
  const elapsedSeconds = Math.max(0, Math.floor((now.getTime() - createdAt) / 1000));
  return elapsedSeconds >= grabHallSeconds ? "auto_assign" : "grab_hall";
}

export function createDispatchDecision(order, riders = [], options = {}) {
  const rejectedRiderIds = new Set((options.rejectedRiderIds || order?.rejected_rider_ids || []).map((id) => String(id)));
  const candidates = riders
    .filter((rider) => rider.online && rider.capacity > 0 && rider.qualified !== false)
    .filter((rider) => !rejectedRiderIds.has(String(rider.id || "")))
    .sort((left, right) => {
      const leftScore = Number(left.dispatch_priority ?? left.level_priority ?? left.dispatch_score ?? left.score ?? 0);
      const rightScore = Number(right.dispatch_priority ?? right.level_priority ?? right.dispatch_score ?? right.score ?? 0);
      if (leftScore !== rightScore) return rightScore - leftScore;
      const leftAccept = Number(left.average_accept_seconds ?? Infinity);
      const rightAccept = Number(right.average_accept_seconds ?? Infinity);
      if (leftAccept !== rightAccept) return leftAccept - rightAccept;
      return (left.distance_meters ?? Infinity) - (right.distance_meters ?? Infinity);
    });
  const mode = options.manual === true ? "manual_assign" : resolveDispatchMode(order, options.now || new Date());
  const completedCount = Number(options.completedOrderCount ?? order?.completed_order_count ?? 0);
  const fixedDailyOrderCount = Number(options.fixedDailyOrderCount ?? order?.fixed_daily_order_count ?? 0);
  return {
    order_id: order?.id || "",
    mode,
    candidate_rider_id: candidates[0]?.id || "",
    rejected_rider_ids: [...rejectedRiderIds],
    can_decline_without_penalty: fixedDailyOrderCount > 0 && completedCount >= fixedDailyOrderCount,
    idempotency_key: `dispatch:${order?.id || "unknown"}:${order?.version || 0}`
  };
}

export function createDispatchConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event = {}) => {
      const payload = event.payload || {};
      const order = payload.order || event.order || payload;
      const riders = payload.riders || event.riders || [];
      const decisionOptions = {
        ...(payload.options || {}),
        ...(event.options || {})
      };
      if (payload.now || event.now) {
        decisionOptions.now = new Date(payload.now || event.now);
      }
      return createDispatchDecision(order, riders, decisionOptions);
    })
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
