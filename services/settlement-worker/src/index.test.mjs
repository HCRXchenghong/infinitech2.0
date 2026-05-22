import assert from "node:assert/strict";
import test from "node:test";
import { calculateSettlement, createSettlementConsumer, subscribedTopics } from "./index.mjs";

test("settlement worker watches completed orders", () => {
  assert.ok(subscribedTopics.includes("order.completed"));
});

test("settlement calculation is integer-fen based", () => {
  assert.deepEqual(calculateSettlement({ id: "ord_1", amount_fen: 999 }, { platform_commission_bps: 1250, version: 2 }), {
    order_id: "ord_1",
    amount_fen: 999,
    platform_commission_fen: 124,
    merchant_amount_fen: 875,
    idempotency_key: "settlement:ord_1:2"
  });
});

test("settlement consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createSettlementConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { settled: true };
    }
  });
  const event = { id: "obe_settlement_1", topic: "order.completed", idempotency_key: "settlement:ord_1:2" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_settlement_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "settlement-worker");
});
