export const workerName = "outbox-relay-worker";
export const defaultRelayTopics = [
  "order.paid",
  "order.status_changed",
  "order.completed",
  "dispatch.assigned",
  "dispatch.timeout",
  "dispatch.status_changed"
];
export const defaultPollIntervalMs = 5000;
export const defaultBatchLimit = 100;
export const defaultRetryAfterSeconds = 60;
export const defaultMaxAttempts = 10;
export const defaultLeaseSeconds = 60;
export const defaultLeaseRenewIntervalMs = 30000;

export function parseRelayTopics(value, fallback = defaultRelayTopics) {
  if (Array.isArray(value)) {
    const topics = value.map((topic) => String(topic).trim()).filter(Boolean);
    return topics.length > 0 ? topics : [...fallback];
  }
  const topics = String(value || "").split(",").map((topic) => topic.trim()).filter(Boolean);
  return topics.length > 0 ? topics : [...fallback];
}

export function normalizeRelayEvent(event = {}) {
  return {
    id: String(event.id || "").trim(),
    topic: String(event.topic || "").trim(),
    key: String(event.idempotency_key || event.idempotencyKey || event.id || "").trim(),
    aggregate_type: String(event.aggregate_type || event.aggregateType || "").trim(),
    aggregate_id: String(event.aggregate_id || event.aggregateId || "").trim(),
    event_type: String(event.event_type || event.eventType || "").trim(),
    payload: event.payload && typeof event.payload === "object" ? event.payload : {},
    attempts: Number(event.attempts || 0)
  };
}

export function createConsolePublisher(log = console.log) {
  return {
    async publish(message) {
      log(JSON.stringify({ topic: message.topic, key: message.key, payload: message.payload }));
    }
  };
}

export function createKafkaRestPublisher(options = {}) {
  const kafkaRestUrl = String(options.kafkaRestUrl || process.env.KAFKA_REST_URL || "").replace(/\/+$/, "");
  const token = String(options.token || process.env.KAFKA_REST_TOKEN || "").trim();
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (!kafkaRestUrl) {
    throw new Error("KAFKA_REST_URL is required for kafka rest publisher");
  }
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  return {
    async publish(message = {}) {
      const topic = String(message.topic || "").trim();
      if (!topic) {
        throw new Error("kafka topic is required");
      }
      const event = message.event || {};
      const headers = {
        "Content-Type": "application/vnd.kafka.json.v2+json",
        Accept: "application/vnd.kafka.v2+json"
      };
      if (token) {
        headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
      }
      const response = await fetchImpl(`${kafkaRestUrl}/topics/${encodeURIComponent(topic)}`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          records: [{
            key: message.key || event.id || "",
            value: {
              ...message.payload,
              _meta: {
                outbox_event_id: event.id || "",
                aggregate_type: event.aggregate_type || "",
                aggregate_id: event.aggregate_id || "",
                event_type: event.event_type || ""
              }
            }
          }]
        })
      });
      if (!response.ok) {
        const text = typeof response.text === "function" ? await response.text() : "";
        throw new Error(text || `kafka rest publish failed: ${response.status}`);
      }
      return true;
    }
  };
}

export function createOutboxApiClient(options = {}) {
  const apiBaseUrl = String(options.apiBaseUrl || process.env.API_BASE_URL || "http://127.0.0.1:8080").replace(/\/+$/, "");
  const token = String(options.token || process.env.OUTBOX_RELAY_TOKEN || "").trim();
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }

  async function request(path, requestOptions = {}) {
    const headers = {
      "Content-Type": "application/json",
      ...requestOptions.headers
    };
    if (token) {
      headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    }
    const response = await fetchImpl(`${apiBaseUrl}${path}`, {
      method: requestOptions.method || "GET",
      headers,
      body: requestOptions.body === undefined ? undefined : JSON.stringify(requestOptions.body)
    });
    const text = await response.text();
    const body = text ? JSON.parse(text) : {};
    if (!response.ok || body.success === false) {
      throw new Error(body.message || `outbox api request failed: ${response.status}`);
    }
    return body.data;
  }

  return {
    async pendingEvents({ topic = "", limit = 100, now = "" } = {}) {
      const params = new URLSearchParams();
      params.set("status", "pending");
      if (topic) params.set("topic", topic);
      if (limit > 0) params.set("limit", String(limit));
      if (now) params.set("now", now instanceof Date ? now.toISOString() : String(now));
      return request(`/api/admin/outbox/events?${params.toString()}`);
    },
    async claimEvents({ topic = "", limit = 100, leaseOwner = "outbox-relay", leaseSeconds = defaultLeaseSeconds, now = new Date() } = {}) {
      return request("/api/admin/outbox/events/claim", {
        method: "POST",
        body: {
          topic,
          limit,
          lease_owner: leaseOwner,
          lease_seconds: leaseSeconds,
          now: now instanceof Date ? now.toISOString() : now
        }
      });
    },
    async renewLease(eventID, { leaseOwner = "outbox-relay", leaseSeconds = defaultLeaseSeconds, now = new Date() } = {}) {
      return request(`/api/admin/outbox/events/${encodeURIComponent(eventID)}/lease/renew`, {
        method: "POST",
        body: {
          lease_owner: leaseOwner,
          lease_seconds: leaseSeconds,
          now: now instanceof Date ? now.toISOString() : now
        }
      });
    },
    async markPublished(eventID, publishedAt = new Date()) {
      return request(`/api/admin/outbox/events/${encodeURIComponent(eventID)}/published`, {
        method: "POST",
        body: { published_at: publishedAt instanceof Date ? publishedAt.toISOString() : publishedAt }
      });
    },
    async markFailed(eventID, error, retryAfterSeconds = 60, now = new Date(), maxAttempts = 0) {
      const body = {
        error: String(error || "").slice(0, 500),
        retry_after_seconds: retryAfterSeconds,
        now: now instanceof Date ? now.toISOString() : now
      };
      if (Number(maxAttempts) > 0) {
        body.max_attempts = Number(maxAttempts);
      }
      return request(`/api/admin/outbox/events/${encodeURIComponent(eventID)}/failed`, {
        method: "POST",
        body
      });
    },
    async replayEvent(eventID, now = new Date()) {
      return request(`/api/admin/outbox/events/${encodeURIComponent(eventID)}/replay`, {
        method: "POST",
        body: { now: now instanceof Date ? now.toISOString() : now }
      });
    },
    async replayEvents({ topic = "", limit = 100, now = new Date() } = {}) {
      return request("/api/admin/outbox/events/replay", {
        method: "POST",
        body: {
          topic,
          limit,
          now: now instanceof Date ? now.toISOString() : now
        }
      });
    }
  };
}

export function createLeaseRenewalLoop(options = {}) {
  const client = options.client;
  const eventID = String(options.eventID || "").trim();
  const leaseOwner = String(options.leaseOwner || "outbox-relay").trim() || "outbox-relay";
  const leaseSeconds = Number.isFinite(Number(options.leaseSeconds)) ? Number(options.leaseSeconds) : defaultLeaseSeconds;
  const intervalMs = Number.isFinite(Number(options.intervalMs)) ? Number(options.intervalMs) : defaultLeaseRenewIntervalMs;
  const nowFactory = options.nowFactory || (() => new Date());
  const setIntervalImpl = options.setIntervalImpl || setInterval;
  const clearIntervalImpl = options.clearIntervalImpl || clearInterval;
  const onError = options.onError;
  if (!eventID || intervalMs <= 0 || !client || typeof client.renewLease !== "function") {
    return { stop() {} };
  }
  const timer = setIntervalImpl(() => {
    Promise.resolve(client.renewLease(eventID, { leaseOwner, leaseSeconds, now: nowFactory() }))
      .catch((error) => {
        if (typeof onError === "function") {
          onError(error);
        }
      });
  }, intervalMs);
  return {
    stop() {
      clearIntervalImpl(timer);
    }
  };
}

export async function publishWithLeaseRenewal(options = {}) {
  const publisher = options.publisher;
  const event = options.event;
  if (!publisher || typeof publisher.publish !== "function") {
    throw new Error("publisher.publish is required");
  }
  const loop = createLeaseRenewalLoop({
    client: options.client,
    eventID: event?.id,
    leaseOwner: options.leaseOwner,
    leaseSeconds: options.leaseSeconds,
    intervalMs: options.leaseRenewIntervalMs,
    nowFactory: options.nowFactory,
    setIntervalImpl: options.setIntervalImpl,
    clearIntervalImpl: options.clearIntervalImpl,
    onError: options.onRenewError
  });
  try {
    return await publisher.publish(options.message);
  } finally {
    loop.stop();
  }
}

export async function relayOutboxBatch(options = {}) {
  const client = options.client;
  const publisher = options.publisher;
  if (!client || (typeof client.claimEvents !== "function" && typeof client.pendingEvents !== "function") || typeof client.markPublished !== "function" || typeof client.markFailed !== "function") {
    throw new Error("outbox client with claimEvents or pendingEvents, markPublished and markFailed is required");
  }
  if (!publisher || typeof publisher.publish !== "function") {
    throw new Error("publisher.publish is required");
  }

  const topics = Array.isArray(options.topics) && options.topics.length > 0 ? options.topics : [""];
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 100;
  const retryAfterSeconds = Number.isFinite(Number(options.retryAfterSeconds)) ? Number(options.retryAfterSeconds) : 60;
  const maxAttempts = Number.isFinite(Number(options.maxAttempts)) ? Number(options.maxAttempts) : defaultMaxAttempts;
  const leaseOwner = String(options.leaseOwner || process.env.OUTBOX_RELAY_WORKER_ID || `${workerName}-${process.pid}`).trim() || "outbox-relay";
  const leaseSeconds = Number.isFinite(Number(options.leaseSeconds)) ? Number(options.leaseSeconds) : defaultLeaseSeconds;
  const leaseRenewIntervalMs = Number.isFinite(Number(options.leaseRenewIntervalMs))
    ? Number(options.leaseRenewIntervalMs)
    : Math.max(1000, Math.floor(leaseSeconds * 1000 / 2));
  const nowFactory = options.nowFactory || (() => new Date());
  const now = options.now || new Date();
  const publishedAt = options.publishedAt || now;
  const result = {
    scanned: 0,
    published: 0,
    failed: 0,
    errors: []
  };

  for (const topic of topics) {
    const claimResult = typeof client.claimEvents === "function"
      ? await client.claimEvents({ topic, limit, leaseOwner, leaseSeconds, now })
      : { events: await client.pendingEvents({ topic, limit, now }) };
    const events = Array.isArray(claimResult) ? claimResult : claimResult.events || [];
    for (const rawEvent of events || []) {
      const event = normalizeRelayEvent(rawEvent);
      if (!event.id || !event.topic) {
        result.failed++;
        result.errors.push({ id: event.id, message: "invalid outbox event" });
        continue;
      }
      result.scanned++;
      const renewErrors = [];
      try {
        await publishWithLeaseRenewal({
          client,
          publisher,
          event,
          leaseOwner,
          leaseSeconds,
          leaseRenewIntervalMs,
          nowFactory,
          setIntervalImpl: options.setIntervalImpl,
          clearIntervalImpl: options.clearIntervalImpl,
          onRenewError(error) {
            const message = error instanceof Error ? error.message : String(error);
            renewErrors.push(message);
            result.errors.push({ id: event.id, message: `lease renewal failed: ${message}` });
          },
          message: {
            topic: event.topic,
            key: event.key,
            payload: event.payload,
            event
          }
        });
        await client.markPublished(event.id, publishedAt);
        result.published++;
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        await client.markFailed(event.id, message, retryAfterSeconds, now, maxAttempts);
        result.failed++;
        result.errors.push({ id: event.id, message });
      }
      if (renewErrors.length > 0) {
        result.leaseRenewalFailed = (result.leaseRenewalFailed || 0) + renewErrors.length;
      }
    }
  }
  return result;
}

export function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, Math.max(0, ms)));
}

export async function runRelayLoop(options = {}) {
  const topics = parseRelayTopics(options.topics);
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : defaultBatchLimit;
  const retryAfterSeconds = Number.isFinite(Number(options.retryAfterSeconds)) ? Number(options.retryAfterSeconds) : defaultRetryAfterSeconds;
  const maxAttempts = Number.isFinite(Number(options.maxAttempts)) ? Number(options.maxAttempts) : defaultMaxAttempts;
  const leaseOwner = String(options.leaseOwner || process.env.OUTBOX_RELAY_WORKER_ID || `${workerName}-${process.pid}`).trim() || "outbox-relay";
  const leaseSeconds = Number.isFinite(Number(options.leaseSeconds)) ? Number(options.leaseSeconds) : defaultLeaseSeconds;
  const leaseRenewIntervalMs = Number.isFinite(Number(options.leaseRenewIntervalMs)) ? Number(options.leaseRenewIntervalMs) : defaultLeaseRenewIntervalMs;
  const intervalMs = Number.isFinite(Number(options.intervalMs)) ? Number(options.intervalMs) : defaultPollIntervalMs;
  const maxIterations = Number.isFinite(Number(options.maxIterations)) ? Number(options.maxIterations) : Infinity;
  const sleepImpl = options.sleepImpl || sleep;
  const nowFactory = options.nowFactory || (() => new Date());
  const shouldStop = options.shouldStop || (() => false);
  const summary = {
    iterations: 0,
    scanned: 0,
    published: 0,
    failed: 0,
    errors: []
  };

  while (summary.iterations < maxIterations && !shouldStop()) {
    const now = nowFactory();
    try {
      const result = await relayOutboxBatch({
        client: options.client,
        publisher: options.publisher,
        topics,
        limit,
        retryAfterSeconds,
        maxAttempts,
        leaseOwner,
        leaseSeconds,
        leaseRenewIntervalMs,
        now,
        publishedAt: now,
        nowFactory,
        setIntervalImpl: options.setIntervalImpl,
        clearIntervalImpl: options.clearIntervalImpl
      });
      summary.scanned += result.scanned;
      summary.published += result.published;
      summary.failed += result.failed;
      summary.errors.push(...result.errors);
      if (typeof options.onResult === "function") {
        options.onResult(result);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      summary.failed++;
      summary.errors.push({ message });
      if (typeof options.onError === "function") {
        options.onError(error);
      }
      if (options.stopOnError === true) {
        throw error;
      }
    }
    summary.iterations++;
    if (summary.iterations >= maxIterations || shouldStop()) {
      break;
    }
    if (intervalMs > 0) {
      await sleepImpl(intervalMs);
    }
  }
  return summary;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  let stopping = false;
  process.once("SIGINT", () => {
    stopping = true;
  });
  process.once("SIGTERM", () => {
    stopping = true;
  });
  const client = createOutboxApiClient();
  const publisher = process.env.KAFKA_REST_URL ? createKafkaRestPublisher() : createConsolePublisher();
  runRelayLoop({
    client,
    publisher,
    topics: parseRelayTopics(process.env.OUTBOX_RELAY_TOPICS),
    limit: Number(process.env.OUTBOX_RELAY_BATCH_LIMIT || defaultBatchLimit),
    retryAfterSeconds: Number(process.env.OUTBOX_RELAY_RETRY_AFTER_SECONDS || defaultRetryAfterSeconds),
    maxAttempts: Number(process.env.OUTBOX_RELAY_MAX_ATTEMPTS || defaultMaxAttempts),
    leaseOwner: process.env.OUTBOX_RELAY_WORKER_ID || `${workerName}-${process.pid}`,
    leaseSeconds: Number(process.env.OUTBOX_RELAY_LEASE_SECONDS || defaultLeaseSeconds),
    leaseRenewIntervalMs: Number(process.env.OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS || defaultLeaseRenewIntervalMs),
    intervalMs: Number(process.env.OUTBOX_RELAY_INTERVAL_MS || defaultPollIntervalMs),
    maxIterations: process.env.OUTBOX_RELAY_MAX_ITERATIONS ? Number(process.env.OUTBOX_RELAY_MAX_ITERATIONS) : Infinity,
    shouldStop: () => stopping,
    onResult(result) {
      console.log(`${workerName} relay tick; scanned=${result.scanned} published=${result.published} failed=${result.failed}`);
    },
    onError(error) {
      console.error(`${workerName} tick failed: ${error.message}`);
    }
  })
    .then((summary) => {
      console.log(`${workerName} stopped; iterations=${summary.iterations} scanned=${summary.scanned} published=${summary.published} failed=${summary.failed}`);
    })
    .catch((error) => {
      console.error(`${workerName} failed: ${error.message}`);
      process.exitCode = 1;
    });
}
