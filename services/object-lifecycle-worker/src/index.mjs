export const workerName = "object-lifecycle-worker";
export const defaultBatchLimit = 100;
export const defaultGraceSeconds = 3600;
export const defaultPollIntervalMs = 60000;

function clean(value) {
  return String(value || "").trim();
}

function positiveInteger(value, fallback) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return fallback;
  return Math.trunc(numeric);
}

function encodeObjectPath(value) {
  return clean(value)
    .split("/")
    .map((part) => clean(part))
    .filter(Boolean)
    .map((part) => encodeURIComponent(part))
    .join("/");
}

export function normalizeCleanupCandidate(candidate = {}) {
  return {
    ticket_id: clean(candidate.ticket_id || candidate.ticketId),
    request_id: clean(candidate.request_id || candidate.requestId),
    order_id: clean(candidate.order_id || candidate.orderId),
    provider: clean(candidate.provider || "minio"),
    bucket: clean(candidate.bucket),
    object_key: clean(candidate.object_key || candidate.objectKey),
    public_url: clean(candidate.public_url || candidate.publicUrl),
    status: clean(candidate.status),
    scan_status: clean(candidate.scan_status || candidate.scanStatus),
    reason: clean(candidate.reason),
    retain_until: clean(candidate.retain_until || candidate.retainUntil),
    expires_at: clean(candidate.expires_at || candidate.expiresAt),
    cleanup_attempts: positiveInteger(candidate.cleanup_attempts || candidate.cleanupAttempts, 0),
    last_cleanup_error: clean(candidate.last_cleanup_error || candidate.lastCleanupError),
    last_cleanup_failed_at: clean(candidate.last_cleanup_failed_at || candidate.lastCleanupFailedAt)
  };
}

export function buildObjectDeleteURL(candidate = {}, options = {}) {
  const normalized = normalizeCleanupCandidate(candidate);
  const directURL = clean(candidate.delete_url || candidate.deleteUrl);
  if (directURL) return directURL;
  const baseUrl = clean(options.deleteBaseUrl || options.objectStorageDeleteBaseUrl || process.env.OBJECT_STORAGE_DELETE_BASE_URL || process.env.OBJECT_STORAGE_DOWNLOAD_BASE_URL).replace(/\/+$/, "");
  if (!baseUrl || !normalized.object_key) return "";
  const path = [encodeObjectPath(normalized.bucket), encodeObjectPath(normalized.object_key)].filter(Boolean).join("/");
  return path ? `${baseUrl}/${path}` : "";
}

async function parseJSONResponse(response, fallbackMessage = "api request failed") {
  const text = typeof response.text === "function" ? await response.text() : "";
  const body = text ? JSON.parse(text) : {};
  if (!response.ok || body.success === false) {
    throw new Error(body.message || `${fallbackMessage}: ${response.status}`);
  }
  return body.data ?? body;
}

export function createObjectLifecycleApiClient(options = {}) {
  const apiBaseUrl = clean(options.apiBaseUrl || process.env.API_BASE_URL || "http://127.0.0.1:1029").replace(/\/+$/, "");
  const token = clean(options.token || process.env.OBJECT_LIFECYCLE_WORKER_TOKEN || "");
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  function headers(extra = {}) {
    const output = { "Content-Type": "application/json", ...extra };
    if (token) output.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    return output;
  }
  return {
    async cleanupCandidates({ limit = defaultBatchLimit, graceSeconds = defaultGraceSeconds, now = new Date() } = {}) {
      const params = new URLSearchParams();
      params.set("limit", String(positiveInteger(limit, defaultBatchLimit)));
      params.set("grace_seconds", String(positiveInteger(graceSeconds, defaultGraceSeconds)));
      if (now) params.set("now", now instanceof Date ? now.toISOString() : String(now));
      const response = await fetchImpl(`${apiBaseUrl}/api/admin/object-storage/cleanup-candidates?${params.toString()}`, {
        method: "GET",
        headers: headers()
      });
      const data = await parseJSONResponse(response, "object cleanup candidates request failed");
      return Array.isArray(data) ? data.map(normalizeCleanupCandidate) : [];
    },
    async completeCleanup(candidate = {}, deletedAt = new Date()) {
      const normalized = normalizeCleanupCandidate(candidate);
      const response = await fetchImpl(`${apiBaseUrl}/api/admin/object-storage/cleanup-complete`, {
        method: "POST",
        headers: headers(),
        body: JSON.stringify({
          ticket_id: normalized.ticket_id,
          object_key: normalized.object_key,
          reason: normalized.reason,
          deleted_at: deletedAt instanceof Date ? deletedAt.toISOString() : deletedAt
        })
      });
      return parseJSONResponse(response, "object cleanup completion request failed");
    },
    async recordCleanupFailure(candidate = {}, error = "", failedAt = new Date()) {
      const normalized = normalizeCleanupCandidate(candidate);
      const response = await fetchImpl(`${apiBaseUrl}/api/admin/object-storage/cleanup-failed`, {
        method: "POST",
        headers: headers(),
        body: JSON.stringify({
          ticket_id: normalized.ticket_id,
          object_key: normalized.object_key,
          reason: normalized.reason,
          error: clean(error),
          failed_at: failedAt instanceof Date ? failedAt.toISOString() : failedAt
        })
      });
      return parseJSONResponse(response, "object cleanup failure request failed");
    }
  };
}

export async function deleteObject(candidate = {}, options = {}) {
  const normalized = normalizeCleanupCandidate(candidate);
  if (!normalized.ticket_id || !normalized.object_key || !normalized.reason) {
    throw new Error("cleanup candidate must include ticket_id, object_key and reason");
  }
  const url = buildObjectDeleteURL(normalized, options);
  if (!url) {
    throw new Error("object delete URL is required");
  }
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  const token = clean(options.deleteToken || process.env.OBJECT_STORAGE_DELETE_TOKEN || "");
  const authorization = clean(options.deleteAuthorization || process.env.OBJECT_STORAGE_DELETE_AUTHORIZATION || "");
  const headers = {};
  if (authorization) {
    headers.Authorization = authorization;
  } else if (token) {
    headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
  }
  const response = await fetchImpl(url, { method: "DELETE", headers });
  if (!response || (response.ok === false && response.status !== 404)) {
    throw new Error(`object delete failed: ${response?.status || 0}`);
  }
  return { ticket_id: normalized.ticket_id, object_key: normalized.object_key, reason: normalized.reason, deleted: true };
}

export async function cleanupObjectBatch(options = {}) {
  const now = options.now || new Date();
  const client = options.client || createObjectLifecycleApiClient(options);
  const deleter = options.deleter || ((candidate) => deleteObject(candidate, options));
  const candidates = await client.cleanupCandidates({
    limit: options.limit || process.env.OBJECT_LIFECYCLE_BATCH_LIMIT,
    graceSeconds: options.graceSeconds || process.env.OBJECT_LIFECYCLE_GRACE_SECONDS,
    now
  });
  const deleted = [];
  const failed = [];
  for (const candidate of candidates) {
    try {
      const deletion = await deleter(candidate);
      const completed = await client.completeCleanup(candidate, options.deletedAt || now);
      deleted.push({ candidate, deletion, completed });
    } catch (error) {
      const message = String(error?.message || error);
      const failure = { candidate, error: message, reported: false };
      if (typeof client.recordCleanupFailure === "function") {
        try {
          failure.record = await client.recordCleanupFailure(candidate, message, options.failedAt || now);
          failure.reported = true;
        } catch (reportError) {
          failure.report_error = String(reportError?.message || reportError);
        }
      }
      failed.push(failure);
    }
  }
  return {
    checked: candidates.length,
    deleted: deleted.length,
    failed: failed.length,
    deleted_items: deleted,
    failed_items: failed
  };
}

export function createLifecycleLoop(options = {}) {
  const intervalMs = positiveInteger(options.intervalMs || process.env.OBJECT_LIFECYCLE_INTERVAL_MS, defaultPollIntervalMs);
  const setIntervalImpl = options.setIntervalImpl || setInterval;
  const clearIntervalImpl = options.clearIntervalImpl || clearInterval;
  const onResult = options.onResult;
  const onError = options.onError;
  const tick = () => cleanupObjectBatch(options)
    .then((result) => {
      if (typeof onResult === "function") onResult(result);
      return result;
    })
    .catch((error) => {
      if (typeof onError === "function") onError(error);
      return { checked: 0, deleted: 0, failed: 1, error: String(error?.message || error) };
    });
  const timer = setIntervalImpl(tick, intervalMs);
  return {
    tick,
    stop() {
      clearIntervalImpl(timer);
    }
  };
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const loop = createLifecycleLoop({
    intervalMs: process.env.OBJECT_LIFECYCLE_INTERVAL_MS,
    onResult(result) {
      console.log(`${workerName} cleanup tick; checked=${result.checked} deleted=${result.deleted} failed=${result.failed}`);
    },
    onError(error) {
      console.error(`${workerName} cleanup tick failed: ${error.message}`);
    }
  });
  loop.tick();
  process.once("SIGINT", () => loop.stop());
  process.once("SIGTERM", () => loop.stop());
}
