import assert from "node:assert/strict";
import test from "node:test";
import {
  archiveAuditEvent,
  archiveOutboxBatch,
  buildAuditArchiveCompletion,
  buildAuditArchiveUploadURL,
  computeArchiveManifestHash,
  createArchiveLoop,
  createAuditArchiveApiClient,
  createAuditArchiveConsumer,
  createAuditArchiveUploader,
  normalizeArchiveRequest,
  serializeAuditArchive,
  subscribedTopics,
  workerName
} from "./index.mjs";

function sampleArchive(overrides = {}) {
  const archive = {
    archive_id: "audit_archive_demo",
    status: "requested",
    storage_prefix: "worm://audit-logs",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_demo.jsonl",
    hot_days: 180,
    cold_archive_cutoff: "2025-11-25T00:00:00Z",
    log_count: 1,
    integrity_failures: 0,
    manifest_algorithm: "sha256:v1",
    manifest_entries: [{
      id: "aud_1",
      created_at: "2025-10-01T08:30:00Z",
      action: "admin.order.refunded",
      target_type: "order",
      target_id: "ord_1",
      integrity_algorithm: "sha256:v1",
      integrity_hash: "hash_1",
      integrity_verified: true
    }],
    requested_at: "2026-05-24T00:00:00Z",
    idempotency_key: "audit_archive:2026-05-24:demo",
    ...overrides
  };
  return {
    ...archive,
    manifest_hash: overrides.manifest_hash || computeArchiveManifestHash(archive)
  };
}

test("audit archive worker declares the archive request topic", () => {
  assert.equal(workerName, "audit-archive-worker");
  assert.deepEqual(subscribedTopics, ["audit.archive_requested"]);
});

test("normalize archive request preserves manifest entries and idempotency", () => {
  const archive = normalizeArchiveRequest({
    id: "obe_1",
    payload: sampleArchive({ idempotency_key: "archive_key" })
  });
  assert.equal(archive.outbox_event_id, "obe_1");
  assert.equal(archive.idempotency_key, "archive_key");
  assert.equal(archive.manifest_entries[0].target_type, "order");
  assert.equal(archive.manifest_entries[0].integrity_verified, true);
});

test("manifest hash stays compatible with backend archive hash contract", () => {
  const archive = sampleArchive();
  assert.equal(computeArchiveManifestHash(archive), archive.manifest_hash);

  const changed = {
    ...archive,
    manifest_entries: [{ ...archive.manifest_entries[0], target_id: "ord_2" }]
  };
  assert.notEqual(computeArchiveManifestHash(changed), archive.manifest_hash);
});

test("serialize audit archive writes manifest header and entry JSONL", () => {
  const archive = sampleArchive();
  const serialized = serializeAuditArchive(archive);
  const lines = serialized.trim().split("\n").map((line) => JSON.parse(line));
  assert.equal(lines[0].type, "audit_archive_manifest");
  assert.equal(lines[0].manifest_hash, archive.manifest_hash);
  assert.equal(lines[1].type, "audit_log_manifest_entry");
  assert.equal(lines[1].archive_id, archive.archive_id);
});

test("upload URL normalizes WORM object keys", () => {
  assert.equal(
    buildAuditArchiveUploadURL(sampleArchive(), { uploadBaseUrl: "https://storage.example.test" }),
    "https://storage.example.test/audit-logs/2026/05/24/audit_archive_demo.jsonl"
  );
});

test("archive event verifies manifest hash and uploads immutable JSONL", async () => {
  const archive = sampleArchive();
  const calls = [];
  const result = await archiveAuditEvent({ id: "obe_1", payload: archive }, {
    uploadBaseUrl: "https://storage.example.test",
    uploadToken: "archive-token",
    now: new Date("2026-05-24T00:00:00Z"),
    retentionDays: 7,
    fetchImpl: async (url, request) => {
      calls.push({ url, request });
      return { ok: true, status: 200, text: async () => "" };
    }
  });

  assert.equal(calls[0].url, "https://storage.example.test/audit-logs/2026/05/24/audit_archive_demo.jsonl");
  assert.equal(calls[0].request.method, "PUT");
  assert.equal(calls[0].request.headers.Authorization, "Bearer archive-token");
  assert.equal(calls[0].request.headers["x-amz-object-lock-mode"], "COMPLIANCE");
  assert.equal(calls[0].request.headers["x-amz-object-lock-retain-until-date"], "2026-05-31T00:00:00.000Z");
  assert.match(calls[0].request.body, /audit_archive_manifest/);
  assert.equal(result.archive_id, archive.archive_id);
  assert.equal(result.manifest_hash, archive.manifest_hash);
  assert.equal(result.object_lock_mode, "COMPLIANCE");
  assert.equal(result.retain_until, "2026-05-31T00:00:00.000Z");
  assert.ok(result.content_hash);
});

test("archive event rejects manifest hash mismatch", async () => {
  const archive = sampleArchive({ manifest_hash: "bad" });
  await assert.rejects(
    () => archiveAuditEvent({ payload: archive }, { uploader: async () => ({}) }),
    /audit archive manifest hash mismatch/
  );
});

test("audit archive uploader reports storage errors", async () => {
  const upload = createAuditArchiveUploader({
    uploadBaseUrl: "https://storage.example.test",
    fetchImpl: async () => ({ ok: false, status: 503, text: async () => "storage unavailable" })
  });
  await assert.rejects(() => upload(sampleArchive()), /storage unavailable/);
});

test("audit archive consumer reports duplicates once", async () => {
  const handled = [];
  const consumer = createAuditArchiveConsumer({
    handler: async (event) => {
      handled.push(event.id);
      return { ok: true };
    },
    clock: () => new Date("2026-05-24T00:00:00Z")
  });
  const event = {
    id: "obe_archive_1",
    topic: "audit.archive_requested",
    idempotency_key: "archive_key_1",
    payload: sampleArchive()
  };

  const first = await consumer(event);
  const duplicate = await consumer(event);

  assert.equal(first.status, "processed");
  assert.equal(duplicate.status, "duplicate");
  assert.deepEqual(handled, ["obe_archive_1"]);
});

test("archive completion payload preserves storage proof fields", () => {
  const archive = sampleArchive();
  const completion = buildAuditArchiveCompletion(
    { id: "obe_archive_1", payload: archive },
    {
      content_hash: "content_hash_1",
      bytes: 512,
      object_lock_mode: "COMPLIANCE",
      retain_until: "2033-05-24T00:00:00Z"
    },
    { uploadedAt: new Date("2026-05-24T00:00:01Z") }
  );
  assert.deepEqual(completion, {
    archive_id: archive.archive_id,
    storage_key: archive.storage_key,
    manifest_algorithm: archive.manifest_algorithm,
    manifest_hash: archive.manifest_hash,
    content_hash: "content_hash_1",
    bytes: 512,
    object_lock_mode: "COMPLIANCE",
    retain_until: "2033-05-24T00:00:00Z",
    outbox_event_id: "obe_archive_1",
    uploaded_at: "2026-05-24T00:00:01.000Z"
  });
});

test("archive outbox batch publishes success and marks failures for retry", async () => {
  const okArchive = sampleArchive();
  const badArchive = sampleArchive({ manifest_hash: "wrong" });
  const calls = [];
  const client = {
    async claimEvents(request) {
      calls.push(["claim", request]);
      return {
        events: [
          { id: "obe_ok", topic: "audit.archive_requested", idempotency_key: "ok", payload: okArchive },
          { id: "obe_bad", topic: "audit.archive_requested", idempotency_key: "bad", payload: badArchive }
        ]
      };
    },
    async markPublished(eventID, publishedAt) {
      calls.push(["published", eventID, publishedAt.toISOString()]);
    },
    async completeArchive(completion) {
      calls.push(["complete", completion]);
    },
    async markFailed(eventID, error, retryAfterSeconds, now, maxAttempts) {
      calls.push(["failed", eventID, error, retryAfterSeconds, now.toISOString(), maxAttempts]);
    }
  };

  const result = await archiveOutboxBatch({
    client,
    uploader: async (archive) => ({
      archive_id: archive.archive_id,
      content_hash: "content_hash_ok",
      bytes: 256,
      object_lock_mode: "COMPLIANCE",
      retain_until: "2033-05-24T00:00:00Z"
    }),
    now: new Date("2026-05-24T00:00:00Z"),
    publishedAt: new Date("2026-05-24T00:00:01Z"),
    retryAfterSeconds: 120,
    maxAttempts: 4,
    leaseOwner: "archive-a",
    leaseSeconds: 90
  });

  assert.equal(result.claimed, 2);
  assert.equal(result.archived, 1);
  assert.equal(result.failed, 1);
  assert.equal(calls[0][0], "claim");
  assert.equal(calls[0][1].topic, "audit.archive_requested");
  assert.equal(calls[0][1].leaseOwner, "archive-a");
  assert.equal(calls[1][0], "complete");
  assert.equal(calls[1][1].archive_id, okArchive.archive_id);
  assert.equal(calls[1][1].content_hash, "content_hash_ok");
  assert.deepEqual(calls[2], ["published", "obe_ok", "2026-05-24T00:00:01.000Z"]);
  assert.equal(calls[3][0], "failed");
  assert.equal(calls[3][1], "obe_bad");
  assert.match(calls[3][2], /manifest hash mismatch/);
  assert.equal(calls[3][3], 120);
  assert.equal(calls[3][5], 4);
});

test("audit archive api client claims events with worker authorization", async () => {
  const calls = [];
  const client = createAuditArchiveApiClient({
    apiBaseUrl: "https://api.example.test",
    token: "worker-token",
    fetchImpl: async (url, request) => {
      calls.push({ url, request });
      return { ok: true, text: async () => JSON.stringify({ success: true, data: { events: [] } }) };
    }
  });

  const result = await client.claimEvents({
    topic: "audit.archive_requested",
    limit: 3,
    leaseOwner: "archive-a",
    leaseSeconds: 90,
    now: new Date("2026-05-24T00:00:00Z")
  });

  assert.deepEqual(result, { events: [] });
  assert.equal(calls[0].url, "https://api.example.test/api/admin/outbox/events/claim");
  assert.equal(calls[0].request.headers.Authorization, "Bearer worker-token");
  assert.deepEqual(JSON.parse(calls[0].request.body), {
    topic: "audit.archive_requested",
    limit: 3,
    lease_owner: "archive-a",
    lease_seconds: 90,
    now: "2026-05-24T00:00:00.000Z"
  });
});

test("audit archive api client records completion evidence", async () => {
  const calls = [];
  const client = createAuditArchiveApiClient({
    apiBaseUrl: "https://api.example.test",
    token: "worker-token",
    fetchImpl: async (url, request) => {
      calls.push({ url, request });
      return { ok: true, text: async () => JSON.stringify({ success: true, data: { archive_id: "audit_archive_demo" } }) };
    }
  });

  await client.completeArchive({
    archive_id: "audit_archive_demo",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_demo.jsonl",
    manifest_algorithm: "sha256:v1",
    manifest_hash: "manifest_hash_1",
    content_hash: "content_hash_1",
    bytes: 1024,
    object_lock_mode: "COMPLIANCE",
    retain_until: "2033-05-24T00:00:00Z",
    outbox_event_id: "obe_archive_1",
    uploaded_at: "2026-05-24T00:00:01Z"
  });

  assert.equal(calls[0].url, "https://api.example.test/api/admin/audit-logs/archive/complete");
  assert.equal(calls[0].request.headers.Authorization, "Bearer worker-token");
  assert.deepEqual(JSON.parse(calls[0].request.body), {
    archive_id: "audit_archive_demo",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_demo.jsonl",
    manifest_algorithm: "sha256:v1",
    manifest_hash: "manifest_hash_1",
    content_hash: "content_hash_1",
    bytes: 1024,
    object_lock_mode: "COMPLIANCE",
    retain_until: "2033-05-24T00:00:00Z",
    outbox_event_id: "obe_archive_1",
    uploaded_at: "2026-05-24T00:00:01Z"
  });
});

test("archive loop exposes tick and stop", async () => {
  let scheduled;
  let cleared;
  const loop = createArchiveLoop({
    client: {
      async claimEvents() {
        return { events: [] };
      },
      async markPublished() {},
      async markFailed() {}
    },
    intervalMs: 15000,
    setIntervalImpl(callback, intervalMs) {
      scheduled = { callback, intervalMs };
      return "timer-1";
    },
    clearIntervalImpl(timer) {
      cleared = timer;
    }
  });

  assert.equal(scheduled.intervalMs, 15000);
  assert.deepEqual(await loop.tick(), { claimed: 0, archived: 0, failed: 0, results: [], errors: [] });
  loop.stop();
  assert.equal(cleared, "timer-1");
});
