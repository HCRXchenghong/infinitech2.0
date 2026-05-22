import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "settlement-worker";
export const subscribedTopics = ["order.completed", "settlement.tick", "rider.earning.adjusted", "merchant.payout.requested"];

export function calculateSettlement(order = {}, rule = {}) {
  const amount = Number(order.amount_fen || 0);
  const commissionBps = Number(rule.platform_commission_bps ?? 0);
  const commission = Math.max(0, Math.trunc((amount * commissionBps) / 10000));
  const merchantAmount = Math.max(0, amount - commission);
  return {
    order_id: String(order.id || ""),
    amount_fen: amount,
    platform_commission_fen: commission,
    merchant_amount_fen: merchantAmount,
    idempotency_key: `settlement:${order.id || "unknown"}:${rule.version || 0}`
  };
}

export function createSettlementConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event = {}) => {
      const payload = event.payload || {};
      return calculateSettlement(payload.order || event.order || payload, payload.rule || event.rule || {});
    })
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
