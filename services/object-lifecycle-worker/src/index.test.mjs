import assert from "node:assert/strict";
import test from "node:test";
import {
  buildObjectDeleteURL,
  cleanupObjectBatch,
  createLifecycleLoop,
  createObjectLifecycleApiClient,
  deleteObject,
  normalizeCleanupCandidate,
  workerName
} from "./index.mjs";

test("object lifecycle worker normalizes cleanup candidates", () => {
  assert.equal(workerName, "object-lifecycle-worker");
  assert.deepEqual(normalizeCleanupCandidate({
    ticketId: " aset_1 ",
    requestId: " asr_1 ",
    objectKey: " after-sales/asr_1/sig/evidence.jpg ",
    reason: " expired_unconfirmed "
  }), {
    ticket_id: "aset_1",
    request_id: "asr_1",
    order_id: "",
    provider: "minio",
    bucket: "",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    public_url: "",
    status: "",
    scan_status: "",
    reason: "expired_unconfirmed",
    retain_until: "",
    expires_at: "",
    cleanup_attempts: 0,
    last_cleanup_error: "",
    last_cleanup_failed_at: ""
  });
});

test("object delete URL prefers direct URL and encodes storage paths", () => {
  assert.equal(buildObjectDeleteURL({
    delete_url: " https://objects.test/signed-delete "
  }, {
    deleteBaseUrl: "https://objects.test"
  }), "https://objects.test/signed-delete");
  assert.equal(buildObjectDeleteURL({
    bucket: "after sales",
    object_key: "asr_1/食品 证据.jpg"
  }, {
    deleteBaseUrl: "https://objects.test/"
  }), "https://objects.test/after%20sales/asr_1/%E9%A3%9F%E5%93%81%20%E8%AF%81%E6%8D%AE.jpg");
});

test("object lifecycle api client lists candidates, marks complete and records failures", async () => {
  const calls = [];
  const client = createObjectLifecycleApiClient({
    apiBaseUrl: "https://api.test/",
    token: "admin-token",
    fetchImpl: async (url, init) => {
      calls.push({ url, init, body: init.body ? JSON.parse(init.body) : null });
      if (url.includes("/cleanup-candidates")) {
        return {
          ok: true,
          text: async () => JSON.stringify({
            success: true,
            data: [{
              ticket_id: "aset_1",
              bucket: "after-sales",
              object_key: "asr_1/evidence.jpg",
              reason: "expired_unconfirmed"
            }]
          })
        };
      }
      if (url.includes("/cleanup-failed")) {
        return {
          ok: true,
          text: async () => JSON.stringify({
            success: true,
            data: { ticket_id: "aset_1", cleanup_attempts: 1, last_cleanup_error: JSON.parse(init.body).error }
          })
        };
      }
      return {
        ok: true,
        text: async () => JSON.stringify({
          success: true,
          data: { ticket_id: "aset_1", status: "deleted" }
        })
      };
    }
  });
  const candidates = await client.cleanupCandidates({ limit: 5, graceSeconds: 60, now: "2026-05-22T12:00:00Z" });
  const completed = await client.completeCleanup(candidates[0], "2026-05-22T12:01:00Z");
  const failed = await client.recordCleanupFailure(candidates[0], "delete denied", "2026-05-22T12:02:00Z");
  assert.equal(calls[0].url, "https://api.test/api/admin/object-storage/cleanup-candidates?limit=5&grace_seconds=60&now=2026-05-22T12%3A00%3A00Z");
  assert.equal(calls[0].init.headers.Authorization, "Bearer admin-token");
  assert.equal(calls[1].body.reason, "expired_unconfirmed");
  assert.equal(completed.status, "deleted");
  assert.equal(calls[2].url, "https://api.test/api/admin/object-storage/cleanup-failed");
  assert.equal(calls[2].body.error, "delete denied");
  assert.equal(failed.cleanup_attempts, 1);
});

test("delete object treats missing object as idempotent success", async () => {
  const calls = [];
  const result = await deleteObject({
    ticket_id: "aset_1",
    bucket: "after-sales",
    object_key: "asr_1/evidence.jpg",
    reason: "expired_unconfirmed"
  }, {
    deleteBaseUrl: "https://objects.test",
    deleteToken: "delete-token",
    fetchImpl: async (url, init) => {
      calls.push({ url, init });
      return { ok: false, status: 404 };
    }
  });
  assert.deepEqual(result, {
    ticket_id: "aset_1",
    object_key: "asr_1/evidence.jpg",
    reason: "expired_unconfirmed",
    deleted: true
  });
  assert.equal(calls[0].url, "https://objects.test/after-sales/asr_1/evidence.jpg");
  assert.equal(calls[0].init.method, "DELETE");
  assert.equal(calls[0].init.headers.Authorization, "Bearer delete-token");
});

test("cleanup batch deletes every candidate and records failed deletions", async () => {
  const completed = [];
  const failures = [];
  const result = await cleanupObjectBatch({
    now: new Date("2026-05-22T12:00:00Z"),
    client: {
      async cleanupCandidates() {
        return [
          { ticket_id: "aset_1", object_key: "ok.jpg", reason: "expired_unconfirmed" },
          { ticket_id: "aset_2", object_key: "fail.jpg", reason: "scan_rejected" }
        ];
      },
      async completeCleanup(candidate, deletedAt) {
        completed.push({ candidate, deletedAt });
        return { ticket_id: candidate.ticket_id, status: "deleted" };
      },
      async recordCleanupFailure(candidate, error, failedAt) {
        failures.push({ candidate, error, failedAt });
        return { ticket_id: candidate.ticket_id, cleanup_attempts: 1 };
      }
    },
    deleter: async (candidate) => {
      if (candidate.ticket_id === "aset_2") throw new Error("delete denied");
      return { deleted: true };
    }
  });
  assert.equal(result.checked, 2);
  assert.equal(result.deleted, 1);
  assert.equal(result.failed, 1);
  assert.equal(completed.length, 1);
  assert.equal(failures.length, 1);
  assert.equal(failures[0].error, "delete denied");
  assert.equal(result.failed_items[0].candidate.ticket_id, "aset_2");
  assert.equal(result.failed_items[0].reported, true);
});

test("lifecycle loop exposes manual tick and stoppable timer", async () => {
  let cleared = false;
  let intervalMs = 0;
  const loop = createLifecycleLoop({
    intervalMs: 1234,
    setIntervalImpl: (_fn, ms) => {
      intervalMs = ms;
      return "timer";
    },
    clearIntervalImpl: (timer) => {
      if (timer === "timer") cleared = true;
    },
    client: {
      async cleanupCandidates() {
        return [];
      },
      async completeCleanup() {
        throw new Error("not called");
      }
    }
  });
  assert.equal(intervalMs, 1234);
  assert.deepEqual(await loop.tick(), { checked: 0, deleted: 0, failed: 0, deleted_items: [], failed_items: [] });
  loop.stop();
  assert.equal(cleared, true);
});
