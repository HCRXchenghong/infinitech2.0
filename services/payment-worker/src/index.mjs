import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "payment-worker";
export const subscribedTopics = ["payment.wechat.callback", "payment.refund.requested", "wallet.recharge.callback", "payment.reconcile.tick"];

export function normalizePaymentCallback(callback = {}) {
  const transactionId = String(callback.transaction_id || callback.transactionId || "").trim();
  const outTradeNo = String(callback.out_trade_no || callback.outTradeNo || "").trim();
  return {
    type: "wechat_callback",
    transaction_id: transactionId,
    out_trade_no: outTradeNo,
    amount_fen: Number(callback.amount_fen || callback.amountFen || 0),
    status: String(callback.status || "").trim().toLowerCase() || "unknown",
    idempotency_key: `wechat:${transactionId || outTradeNo}`
  };
}

export function normalizePaymentRefundRequest(event = {}) {
  const payload = event.payload || event.refund || event;
  const orderId = String(payload.order_id || payload.orderId || event.aggregate_id || event.aggregateId || "").trim();
  const refundId = String(payload.refund_id || payload.refundId || event.id || "").trim();
  const idempotencyKey = String(payload.idempotency_key || payload.idempotencyKey || event.idempotency_key || event.idempotencyKey || "").trim()
    || `refund:${orderId || refundId}`;
  return {
    type: "refund_requested",
    refund_id: refundId || `refund_${orderId}`,
    order_id: orderId,
    user_id: String(payload.user_id || payload.userId || "").trim(),
    amount_fen: Number(payload.amount_fen || payload.amountFen || 0),
    destination: String(payload.destination || payload.refund_destination || payload.refundDestination || "original_route").trim() || "original_route",
    status: "pending_original_route",
    idempotency_key: idempotencyKey
  };
}

export function normalizePaymentEvent(event = {}) {
  if (event.topic === "payment.refund.requested") {
    return normalizePaymentRefundRequest(event);
  }
  return normalizePaymentCallback(event.payload || event.callback || event);
}

export function createPaymentConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || normalizePaymentEvent
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
