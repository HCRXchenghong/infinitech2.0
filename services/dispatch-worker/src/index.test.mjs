import assert from "node:assert/strict";
import test from "node:test";
import { createDispatchConsumer, createDispatchDecision, resolveDispatchMode, subscribedTopics } from "./index.mjs";

test("dispatch worker subscribes to order and rider events", () => {
  assert.ok(subscribedTopics.includes("order.paid"));
  assert.ok(subscribedTopics.includes("rider.location_updated"));
});

test("dispatch decision picks nearest qualified online rider", () => {
  const decision = createDispatchDecision({ id: "ord_1", version: 3, created_at: "2026-05-21T08:00:00.000Z" }, [
    { id: "rider_far", online: true, capacity: 1, distance_meters: 5000 },
    { id: "rider_offline", online: false, capacity: 10, distance_meters: 10 },
    { id: "rider_near", online: true, capacity: 1, distance_meters: 300 }
  ], { now: new Date("2026-05-21T08:10:00.000Z") });
  assert.equal(decision.candidate_rider_id, "rider_near");
  assert.equal(decision.mode, "auto_assign");
  assert.equal(decision.idempotency_key, "dispatch:ord_1:3");
});

test("dispatch stays in grab hall for first ten minutes and skips rejected riders", () => {
  assert.equal(resolveDispatchMode({ created_at: "2026-05-21T08:00:01.000Z" }, new Date("2026-05-21T08:10:00.000Z")), "grab_hall");
  const decision = createDispatchDecision({ id: "ord_2", version: 1, created_at: "2026-05-21T08:00:00.000Z" }, [
    { id: "rider_1", online: true, capacity: 1, distance_meters: 100, dispatch_score: 99 },
    { id: "rider_2", online: true, capacity: 1, distance_meters: 200, dispatch_score: 80 }
  ], { now: new Date("2026-05-21T08:10:00.000Z"), rejectedRiderIds: ["rider_1"] });
  assert.equal(decision.mode, "auto_assign");
  assert.equal(decision.candidate_rider_id, "rider_2");
});

test("dispatch priority uses rider level priority and daily quota exemption", () => {
  const decision = createDispatchDecision({ id: "ord_3", version: 1, created_at: "2026-05-21T08:00:00.000Z", completed_order_count: 12, fixed_daily_order_count: 12 }, [
    { id: "rider_b", online: true, capacity: 1, distance_meters: 10, dispatch_priority: 200, average_accept_seconds: 5 },
    { id: "rider_s", online: true, capacity: 1, distance_meters: 500, dispatch_priority: 400, average_accept_seconds: 20 }
  ], { now: new Date("2026-05-21T08:10:00.000Z") });
  assert.equal(decision.candidate_rider_id, "rider_s");
  assert.equal(decision.can_decline_without_penalty, true);
});

test("dispatch consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createDispatchConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { assigned: true };
    }
  });
  const event = { id: "obe_dispatch_1", topic: "order.paid", idempotency_key: "order:ord_1:paid" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_dispatch_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "dispatch-worker");
});
