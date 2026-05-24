import assert from "node:assert/strict";
import test from "node:test";
import {
  createLeaseRenewalLoop,
  createKafkaRestPublisher,
  createOutboxApiClient,
  defaultLeaseSeconds,
  defaultLeaseRenewIntervalMs,
  defaultMaxAttempts,
  defaultRelayTopics,
  normalizeRelayEvent,
  parseRelayTopics,
  relayOutboxBatch,
  runRelayLoop
} from "./index.mjs";

test("outbox relay declares critical platform topics", () => {
  assert.ok(defaultRelayTopics.includes("order.paid"));
  assert.ok(defaultRelayTopics.includes("dispatch.assigned"));
  assert.ok(defaultRelayTopics.includes("order.completed"));
  assert.ok(defaultRelayTopics.includes("audit.retention_alerts"));
  assert.ok(defaultRelayTopics.includes("audit.archive_requested"));
});

test("normalize relay event preserves idempotency key and payload", () => {
  const event = normalizeRelayEvent({
    id: "obe_1",
    topic: "order.paid",
    aggregate_type: "order",
    aggregate_id: "ord_1",
    event_type: "order.payment.success",
    idempotency_key: "order_event:ord_1:paid",
    payload: { order_id: "ord_1" },
    attempts: 2
  });
  assert.equal(event.key, "order_event:ord_1:paid");
  assert.equal(event.aggregate_id, "ord_1");
  assert.deepEqual(event.payload, { order_id: "ord_1" });
  assert.equal(event.attempts, 2);
});

test("relay topics parse comma separated env and fall back to defaults", () => {
  assert.deepEqual(parseRelayTopics("order.paid, dispatch.assigned ,,order.completed"), ["order.paid", "dispatch.assigned", "order.completed"]);
  assert.deepEqual(parseRelayTopics("", ["fallback.topic"]), ["fallback.topic"]);
});

test("relay publishes events and marks them published", async () => {
  const calls = [];
  const client = {
    async pendingEvents(request) {
      calls.push(["pending", request]);
      return [{
        id: "obe_1",
        topic: "order.paid",
        aggregate_id: "ord_1",
        event_type: "order.payment.success",
        idempotency_key: "order_event:ord_1:paid",
        payload: { order_id: "ord_1" }
      }];
    },
    async markPublished(eventID, publishedAt) {
      calls.push(["published", eventID, publishedAt.toISOString()]);
    },
    async markFailed(eventID, error, retryAfterSeconds) {
      calls.push(["failed", eventID, error, retryAfterSeconds]);
    }
  };
  const publishedMessages = [];
  const publisher = {
    async publish(message) {
      publishedMessages.push(message);
    }
  };

  const result = await relayOutboxBatch({
    client,
    publisher,
    topics: ["order.paid"],
    now: new Date("2026-05-22T12:00:00Z"),
    publishedAt: new Date("2026-05-22T12:00:01Z")
  });

  assert.deepEqual(result, { scanned: 1, published: 1, failed: 0, errors: [] });
  assert.equal(publishedMessages[0].topic, "order.paid");
  assert.equal(publishedMessages[0].key, "order_event:ord_1:paid");
  assert.equal(calls[0][0], "pending");
  assert.equal(calls[0][1].topic, "order.paid");
  assert.deepEqual(calls[1], ["published", "obe_1", "2026-05-22T12:00:01.000Z"]);
});

test("relay claims leased events before publishing", async () => {
  const calls = [];
  const client = {
    async claimEvents(request) {
      calls.push(["claim", request]);
      return {
        claimed: 1,
        lease_owner: request.leaseOwner,
        events: [{
          id: "obe_claimed",
          topic: "order.paid",
          aggregate_id: "ord_claimed",
          event_type: "order.payment.success",
          idempotency_key: "order_event:ord_claimed:paid",
          payload: { order_id: "ord_claimed" }
        }]
      };
    },
    async markPublished(eventID, publishedAt) {
      calls.push(["published", eventID, publishedAt.toISOString()]);
    },
    async markFailed(eventID, error, retryAfterSeconds) {
      calls.push(["failed", eventID, error, retryAfterSeconds]);
    }
  };
  const publishedMessages = [];
  const publisher = {
    async publish(message) {
      publishedMessages.push(message);
    }
  };

  const result = await relayOutboxBatch({
    client,
    publisher,
    topics: ["order.paid"],
    limit: 5,
    leaseOwner: "relay-a",
    leaseSeconds: 45,
    now: new Date("2026-05-22T12:00:00Z"),
    publishedAt: new Date("2026-05-22T12:00:01Z")
  });

  assert.deepEqual(result, { scanned: 1, published: 1, failed: 0, errors: [] });
  assert.deepEqual(calls[0], ["claim", {
    topic: "order.paid",
    limit: 5,
    leaseOwner: "relay-a",
    leaseSeconds: 45,
    now: new Date("2026-05-22T12:00:00Z")
  }]);
  assert.equal(publishedMessages[0].topic, "order.paid");
  assert.equal(publishedMessages[0].key, "order_event:ord_claimed:paid");
  assert.deepEqual(calls[1], ["published", "obe_claimed", "2026-05-22T12:00:01.000Z"]);
});

test("lease renewal loop renews claimed event leases and clears timer", async () => {
  const renewals = [];
  const errors = [];
  let scheduled;
  let cleared;
  const loop = createLeaseRenewalLoop({
    client: {
      async renewLease(eventID, request) {
        renewals.push({ eventID, request });
      }
    },
    eventID: "obe_claimed",
    leaseOwner: "relay-a",
    leaseSeconds: 45,
    intervalMs: 15000,
    nowFactory: () => new Date("2026-05-22T12:00:15Z"),
    setIntervalImpl(callback, intervalMs) {
      scheduled = { callback, intervalMs };
      return "timer-1";
    },
    clearIntervalImpl(timer) {
      cleared = timer;
    },
    onError(error) {
      errors.push(error);
    }
  });

  assert.equal(scheduled.intervalMs, 15000);
  scheduled.callback();
  await Promise.resolve();
  loop.stop();

  assert.deepEqual(renewals, [{
    eventID: "obe_claimed",
    request: {
      leaseOwner: "relay-a",
      leaseSeconds: 45,
      now: new Date("2026-05-22T12:00:15Z")
    }
  }]);
  assert.equal(cleared, "timer-1");
  assert.deepEqual(errors, []);
});

test("relay renews leases while publishing slow claimed events", async () => {
  const calls = [];
  const timers = [];
  const client = {
    async claimEvents(request) {
      calls.push(["claim", request]);
      return {
        claimed: 1,
        events: [{
          id: "obe_slow",
          topic: "order.paid",
          idempotency_key: "order_event:ord_slow:paid",
          payload: { order_id: "ord_slow" }
        }]
      };
    },
    async renewLease(eventID, request) {
      calls.push(["renew", eventID, request]);
    },
    async markPublished(eventID, publishedAt) {
      calls.push(["published", eventID, publishedAt.toISOString()]);
    },
    async markFailed(eventID, error) {
      calls.push(["failed", eventID, error]);
    }
  };
  const publisher = {
    async publish() {
      timers[0].callback();
      await Promise.resolve();
    }
  };
  const cleared = [];

  const result = await relayOutboxBatch({
    client,
    publisher,
    topics: ["order.paid"],
    leaseOwner: "relay-a",
    leaseSeconds: 45,
    leaseRenewIntervalMs: 15000,
    now: new Date("2026-05-22T12:00:00Z"),
    publishedAt: new Date("2026-05-22T12:00:40Z"),
    nowFactory: () => new Date("2026-05-22T12:00:15Z"),
    setIntervalImpl(callback, intervalMs) {
      timers.push({ callback, intervalMs });
      return `timer-${timers.length}`;
    },
    clearIntervalImpl(timer) {
      cleared.push(timer);
    }
  });

  assert.deepEqual(result, { scanned: 1, published: 1, failed: 0, errors: [] });
  assert.equal(timers[0].intervalMs, 15000);
  assert.deepEqual(calls[1], ["renew", "obe_slow", {
    leaseOwner: "relay-a",
    leaseSeconds: 45,
    now: new Date("2026-05-22T12:00:15Z")
  }]);
  assert.deepEqual(calls[2], ["published", "obe_slow", "2026-05-22T12:00:40.000Z"]);
  assert.deepEqual(cleared, ["timer-1"]);
});

test("relay marks failed events with retry backoff and continues", async () => {
  const failed = [];
  const client = {
    async pendingEvents() {
      return [
        { id: "obe_bad", topic: "dispatch.assigned", idempotency_key: "dispatch:bad", payload: { order_id: "ord_bad" } },
        { id: "obe_good", topic: "dispatch.assigned", idempotency_key: "dispatch:good", payload: { order_id: "ord_good" } }
      ];
    },
    async markPublished(eventID) {
      failed.push(["published", eventID]);
    },
    async markFailed(eventID, error, retryAfterSeconds, now, maxAttempts) {
      failed.push(["failed", eventID, error, retryAfterSeconds, now.toISOString(), maxAttempts]);
    }
  };
  const publisher = {
    async publish(message) {
      if (message.event.id === "obe_bad") {
        throw new Error("kafka unavailable");
      }
    }
  };

  const result = await relayOutboxBatch({
    client,
    publisher,
    topics: ["dispatch.assigned"],
    retryAfterSeconds: 180,
    maxAttempts: 7,
    now: new Date("2026-05-22T12:00:00Z")
  });

  assert.equal(result.scanned, 2);
  assert.equal(result.published, 1);
  assert.equal(result.failed, 1);
  assert.deepEqual(failed[0], ["failed", "obe_bad", "kafka unavailable", 180, "2026-05-22T12:00:00.000Z", 7]);
  assert.deepEqual(failed[1], ["published", "obe_good"]);
});

test("outbox api client forwards authorization and uses admin endpoints", async () => {
  const requests = [];
  const fetchImpl = async (url, options) => {
    requests.push({ url, options });
    if (url.endsWith("/api/admin/outbox/events/claim")) {
      return jsonResponse({ success: true, data: { topic: "order.paid", claimed: 1, events: [{ id: "obe_1" }] } });
    }
	  if (url.endsWith("/api/admin/outbox/events/replay")) {
	    return jsonResponse({ success: true, data: { topic: "order.paid", replayed: 1, events: [{ id: "obe_1" }] } });
	  }
    if (url.includes("/lease/renew")) {
      return jsonResponse({ success: true, data: { id: "obe_1", status: "pending", lease_owner: "relay-a" } });
    }
	  if (url.includes("/published")) {
	    return jsonResponse({ success: true, data: { id: "obe_1", status: "published" } });
	  }
    if (url.includes("/failed")) {
      return jsonResponse({ success: true, data: { id: "obe_1", status: "failed" } });
    }
    if (url.includes("/replay")) {
      return jsonResponse({ success: true, data: { id: "obe_1", status: "pending" } });
    }
    return jsonResponse({ success: true, data: [{ id: "obe_1", topic: "order.paid" }] });
  };
  const client = createOutboxApiClient({
    apiBaseUrl: "http://api.test",
    token: "admin-token",
    fetchImpl
  });

  await client.pendingEvents({ topic: "order.paid", limit: 5, now: new Date("2026-05-22T12:00:00Z") });
  await client.claimEvents({ topic: "order.paid", limit: 5, leaseOwner: "relay-a", leaseSeconds: 45, now: new Date("2026-05-22T12:00:01Z") });
  await client.renewLease("obe_1", { leaseOwner: "relay-a", leaseSeconds: 45, now: new Date("2026-05-22T12:00:02Z") });
  await client.markPublished("obe_1", new Date("2026-05-22T12:00:03Z"));
  await client.markFailed("obe_1", "relay down", 90, new Date("2026-05-22T12:00:04Z"), 10);
  await client.replayEvent("obe_1", new Date("2026-05-22T12:00:05Z"));
  await client.replayEvents({ topic: "order.paid", limit: 10, now: new Date("2026-05-22T12:00:06Z") });

  assert.equal(requests[0].url, "http://api.test/api/admin/outbox/events?status=pending&topic=order.paid&limit=5&now=2026-05-22T12%3A00%3A00.000Z");
  assert.equal(requests[0].options.headers.Authorization, "Bearer admin-token");
  assert.equal(requests[1].url, "http://api.test/api/admin/outbox/events/claim");
  assert.deepEqual(JSON.parse(requests[1].options.body), {
    topic: "order.paid",
    limit: 5,
    lease_owner: "relay-a",
    lease_seconds: 45,
    now: "2026-05-22T12:00:01.000Z"
  });
  assert.equal(requests[2].url, "http://api.test/api/admin/outbox/events/obe_1/lease/renew");
  assert.deepEqual(JSON.parse(requests[2].options.body), {
    lease_owner: "relay-a",
    lease_seconds: 45,
    now: "2026-05-22T12:00:02.000Z"
  });
  assert.equal(requests[3].url, "http://api.test/api/admin/outbox/events/obe_1/published");
  assert.equal(JSON.parse(requests[3].options.body).published_at, "2026-05-22T12:00:03.000Z");
  assert.equal(requests[4].url, "http://api.test/api/admin/outbox/events/obe_1/failed");
  assert.equal(JSON.parse(requests[4].options.body).retry_after_seconds, 90);
  assert.equal(JSON.parse(requests[4].options.body).max_attempts, 10);
  assert.equal(requests[5].url, "http://api.test/api/admin/outbox/events/obe_1/replay");
  assert.equal(JSON.parse(requests[5].options.body).now, "2026-05-22T12:00:05.000Z");
  assert.equal(requests[6].url, "http://api.test/api/admin/outbox/events/replay");
  assert.deepEqual(JSON.parse(requests[6].options.body), {
    topic: "order.paid",
    limit: 10,
    now: "2026-05-22T12:00:06.000Z"
  });
});

test("kafka rest publisher posts keyed records with outbox metadata", async () => {
  const requests = [];
  const publisher = createKafkaRestPublisher({
    kafkaRestUrl: "http://kafka-rest.test/",
    token: "kafka-token",
    fetchImpl: async (url, options) => {
      requests.push({ url, options });
      return jsonResponse({});
    }
  });

  await publisher.publish({
    topic: "order.paid",
    key: "order_event:ord_1:paid",
    payload: { order_id: "ord_1" },
    event: {
      id: "obe_1",
      aggregate_type: "order",
      aggregate_id: "ord_1",
      event_type: "order.payment.success"
    }
  });

  assert.equal(requests[0].url, "http://kafka-rest.test/topics/order.paid");
  assert.equal(requests[0].options.headers.Authorization, "Bearer kafka-token");
  const body = JSON.parse(requests[0].options.body);
  assert.equal(body.records[0].key, "order_event:ord_1:paid");
  assert.equal(body.records[0].value.order_id, "ord_1");
  assert.equal(body.records[0].value._meta.outbox_event_id, "obe_1");
});

test("relay loop polls repeatedly and aggregates tick results", async () => {
  let claimCalls = 0;
  const slept = [];
  const client = {
    async claimEvents(request) {
      claimCalls++;
      assert.equal(request.leaseOwner, "relay-loop");
      assert.equal(request.leaseSeconds, 45);
      assert.equal(request.leaseRenewIntervalMs, undefined);
      return [{
        id: `obe_${claimCalls}`,
        topic: "order.paid",
        idempotency_key: `order_event:ord_${claimCalls}:paid`,
        payload: { order_id: `ord_${claimCalls}` }
      }];
    },
    async markPublished() {},
    async markFailed() {}
  };
  const publisher = {
    async publish() {}
  };

  const summary = await runRelayLoop({
    client,
    publisher,
    topics: ["order.paid"],
    leaseOwner: "relay-loop",
    leaseSeconds: 45,
    leaseRenewIntervalMs: defaultLeaseRenewIntervalMs,
    maxIterations: 2,
    intervalMs: 25,
    sleepImpl: async (ms) => {
      slept.push(ms);
    },
    nowFactory: () => new Date("2026-05-22T12:00:00Z")
  });

  assert.equal(summary.iterations, 2);
  assert.equal(summary.scanned, 2);
  assert.equal(summary.published, 2);
  assert.equal(summary.failed, 0);
  assert.deepEqual(slept, [25]);
});

test("relay loop forwards default max attempts to failed markers", async () => {
  const failed = [];
  const client = {
    async claimEvents(request) {
      assert.equal(request.leaseSeconds, defaultLeaseSeconds);
      return {
        claimed: 1,
        events: [{
          id: "obe_bad",
          topic: "order.paid",
          idempotency_key: "order_event:ord_bad:paid",
          payload: { order_id: "ord_bad" }
        }]
      };
    },
    async markPublished() {},
    async markFailed(eventID, error, retryAfterSeconds, now, maxAttempts) {
      failed.push({ eventID, error, retryAfterSeconds, now: now.toISOString(), maxAttempts });
    }
  };
  const publisher = {
    async publish() {
      throw new Error("poison payload");
    }
  };

  const summary = await runRelayLoop({
    client,
    publisher,
    topics: ["order.paid"],
    maxIterations: 1,
    retryAfterSeconds: 120,
    nowFactory: () => new Date("2026-05-22T12:00:00Z")
  });

  assert.equal(summary.failed, 1);
  assert.equal(failed[0].eventID, "obe_bad");
  assert.equal(failed[0].maxAttempts, defaultMaxAttempts);
});

function jsonResponse(body, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    async text() {
      return JSON.stringify(body);
    }
  };
}
