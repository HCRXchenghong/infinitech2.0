import assert from "node:assert/strict";
import test from "node:test";
import { createPaymentConsumer, normalizePaymentCallback, normalizePaymentEvent, normalizePaymentRefundRequest, subscribedTopics } from "./index.mjs";

test("payment worker covers callback and reconcile topics", () => {
  assert.ok(subscribedTopics.includes("payment.wechat.callback"));
  assert.ok(subscribedTopics.includes("payment.refund.requested"));
  assert.ok(subscribedTopics.includes("payment.reconcile.tick"));
});

test("payment callback normalization creates stable idempotency key", () => {
  assert.deepEqual(normalizePaymentCallback({
    transaction_id: " wx_1 ",
    out_trade_no: "ord_1",
    amount_fen: 100,
    status: " SUCCESS "
  }), {
    type: "wechat_callback",
    transaction_id: "wx_1",
    out_trade_no: "ord_1",
    amount_fen: 100,
    status: "success",
    idempotency_key: "wechat:wx_1"
  });
});

test("payment refund request normalization creates original-route task payload", () => {
  const event = {
    id: "obe_refund_1",
    topic: "payment.refund.requested",
    aggregate_id: "ord_1",
    idempotency_key: "order_event:ord_1:order.refund.requested:2026-05-22T12:00:00Z",
    payload: {
      order_id: "ord_1",
      user_id: "user_1",
      amount_fen: 1200
    }
  };
  assert.deepEqual(normalizePaymentRefundRequest(event), {
    type: "refund_requested",
    refund_id: "obe_refund_1",
    order_id: "ord_1",
    user_id: "user_1",
    amount_fen: 1200,
    destination: "original_route",
    status: "pending_original_route",
    idempotency_key: "order_event:ord_1:order.refund.requested:2026-05-22T12:00:00Z"
  });
  assert.equal(normalizePaymentEvent(event).type, "refund_requested");
});

test("payment consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createPaymentConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { transaction_id: "wx_1" };
    }
  });
  const event = { id: "obe_payment_1", topic: "payment.wechat.callback", idempotency_key: "wechat:wx_1" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_payment_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "payment-worker");
});
