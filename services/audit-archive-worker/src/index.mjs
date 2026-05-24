import { createHash } from "node:crypto";
import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "audit-archive-worker";
export const subscribedTopics = ["audit.archive_requested"];
export const manifestVersion = "audit_archive_manifest:v1";
export const defaultBatchLimit = 20;
export const defaultPollIntervalMs = 30000;
export const defaultRetryAfterSeconds = 300;
export const defaultMaxAttempts = 10;
export const defaultLeaseSeconds = 300;
export const defaultObjectLockMode = "COMPLIANCE";
export const defaultRetentionDays = 2555;

function clean(value) {
  return String(value ?? "").trim();
}

function positiveInteger(value, fallback) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return fallback;
  return Math.trunc(numeric);
}

function normalizeTopicList(value, fallback = subscribedTopics) {
  const topics = Array.isArray(value)
    ? value.map(clean).filter(Boolean)
    : clean(value).split(",").map(clean).filter(Boolean);
  return topics.length > 0 ? topics : [...fallback];
}

function normalizeArchivePayload(event = {}) {
  if (event.payload && typeof event.payload === "object" && !Array.isArray(event.payload)) {
    return event.payload;
  }
  if (event.archive && typeof event.archive === "object" && !Array.isArray(event.archive)) {
    return event.archive;
  }
  return event;
}

export function normalizeArchiveManifestEntry(entry = {}) {
  return {
    id: clean(entry.id || entry.ID),
    created_at: clean(entry.created_at || entry.createdAt || entry.CreatedAt),
    action: clean(entry.action || entry.Action),
    target_type: clean(entry.target_type || entry.targetType || entry.TargetType),
    target_id: clean(entry.target_id || entry.targetId || entry.TargetID),
    integrity_algorithm: clean(entry.integrity_algorithm || entry.integrityAlgorithm || entry.IntegrityAlgorithm),
    integrity_hash: clean(entry.integrity_hash || entry.integrityHash || entry.IntegrityHash),
    integrity_verified: entry.integrity_verified === true || entry.integrityVerified === true || entry.IntegrityVerified === true
  };
}

export function normalizeArchiveRequest(event = {}) {
  const payload = normalizeArchivePayload(event);
  const entries = Array.isArray(payload.manifest_entries || payload.manifestEntries)
    ? (payload.manifest_entries || payload.manifestEntries).map(normalizeArchiveManifestEntry)
    : [];
  return {
    archive_id: clean(payload.archive_id || payload.archiveId),
    status: clean(payload.status || "requested"),
    storage_prefix: clean(payload.storage_prefix || payload.storagePrefix),
    storage_key: clean(payload.storage_key || payload.storageKey),
    hot_days: positiveInteger(payload.hot_days ?? payload.hotDays, 180),
    cold_archive_cutoff: clean(payload.cold_archive_cutoff || payload.coldArchiveCutoff),
    log_count: positiveInteger(payload.log_count ?? payload.logCount ?? entries.length, entries.length),
    integrity_failures: positiveInteger(payload.integrity_failures ?? payload.integrityFailures, 0),
    manifest_algorithm: clean(payload.manifest_algorithm || payload.manifestAlgorithm || "sha256:v1"),
    manifest_hash: clean(payload.manifest_hash || payload.manifestHash),
    manifest_entries: entries,
    requested_at: clean(payload.requested_at || payload.requestedAt),
    idempotency_key: clean(payload.idempotency_key || payload.idempotencyKey || event.idempotency_key || event.idempotencyKey),
    upload_url: clean(payload.upload_url || payload.uploadUrl),
    outbox_event_id: clean(event.id || event.outbox_event_id || event.outboxEventId || payload.outbox_event_id || payload.outboxEventId)
  };
}

export function archiveManifestHashPayload(archive = {}) {
  const normalized = normalizeArchiveRequest(archive);
  return {
    cold_archive_cutoff: normalized.cold_archive_cutoff,
    entries: normalized.manifest_entries.map((entry) => ({
      id: entry.id,
      created_at: entry.created_at,
      action: entry.action,
      target_type: entry.target_type,
      target_id: entry.target_id,
      integrity_algorithm: entry.integrity_algorithm,
      integrity_hash: entry.integrity_hash,
      integrity_verified: entry.integrity_verified
    })),
    hot_days: normalized.hot_days,
    integrity_failures: normalized.integrity_failures,
    log_count: normalized.log_count,
    manifest_version: manifestVersion,
    requested_at: normalized.requested_at
  };
}

export function computeArchiveManifestHash(archive = {}) {
  const payload = archiveManifestHashPayload(archive);
  return createHash("sha256").update(JSON.stringify(payload)).digest("hex");
}

export function serializeAuditArchive(archive = {}) {
  const normalized = normalizeArchiveRequest(archive);
  const manifestHash = normalized.manifest_hash || computeArchiveManifestHash(normalized);
  const header = {
    type: "audit_archive_manifest",
    manifest_version: manifestVersion,
    archive_id: normalized.archive_id,
    status: normalized.status,
    storage_prefix: normalized.storage_prefix,
    storage_key: normalized.storage_key,
    hot_days: normalized.hot_days,
    cold_archive_cutoff: normalized.cold_archive_cutoff,
    requested_at: normalized.requested_at,
    log_count: normalized.log_count,
    integrity_failures: normalized.integrity_failures,
    manifest_algorithm: normalized.manifest_algorithm,
    manifest_hash: manifestHash,
    idempotency_key: normalized.idempotency_key
  };
  const lines = [
    JSON.stringify(header),
    ...normalized.manifest_entries.map((entry) => JSON.stringify({
      type: "audit_log_manifest_entry",
      archive_id: normalized.archive_id,
      ...entry
    }))
  ];
  return `${lines.join("\n")}\n`;
}

function normalizeStoragePath(value) {
  const raw = clean(value)
    .replace(/^worm:\/\//, "")
    .replace(/^s3:\/\//, "")
    .replace(/^minio:\/\//, "");
  return raw
    .split("/")
    .map(clean)
    .filter(Boolean)
    .map((part) => encodeURIComponent(part))
    .join("/");
}

export function buildAuditArchiveUploadURL(archive = {}, options = {}) {
  const normalized = normalizeArchiveRequest(archive);
  if (normalized.upload_url) return normalized.upload_url;
  const baseUrl = clean(options.uploadBaseUrl || process.env.AUDIT_ARCHIVE_UPLOAD_BASE_URL).replace(/\/+$/, "");
  const path = normalizeStoragePath(normalized.storage_key);
  return baseUrl && path ? `${baseUrl}/${path}` : "";
}

function retentionDate(days = defaultRetentionDays, now = new Date()) {
  const normalizedDays = positiveInteger(days, defaultRetentionDays);
  const base = now instanceof Date ? new Date(now.getTime()) : new Date(now);
  if (Number.isNaN(base.getTime())) return "";
  base.setUTCDate(base.getUTCDate() + normalizedDays);
  return base.toISOString();
}

export function createAuditArchiveUploader(options = {}) {
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  return async function uploadAuditArchive(archive = {}) {
    const normalized = normalizeArchiveRequest(archive);
    const body = serializeAuditArchive(normalized);
    const contentHash = createHash("sha256").update(body).digest("hex");
    const uploadURL = buildAuditArchiveUploadURL(normalized, options);
    if (!uploadURL) {
      throw new Error("audit archive upload URL is required");
    }
    const token = clean(options.uploadToken || process.env.AUDIT_ARCHIVE_UPLOAD_TOKEN);
    const authorization = clean(options.uploadAuthorization || process.env.AUDIT_ARCHIVE_UPLOAD_AUTHORIZATION);
    const objectLockMode = clean(options.objectLockMode ?? process.env.AUDIT_ARCHIVE_OBJECT_LOCK_MODE ?? defaultObjectLockMode);
    const retainUntil = clean(options.retainUntil || process.env.AUDIT_ARCHIVE_RETAIN_UNTIL_DATE)
      || retentionDate(options.retentionDays || process.env.AUDIT_ARCHIVE_RETENTION_DAYS || defaultRetentionDays, options.now || new Date());
    const headers = {
      "Content-Type": "application/x-ndjson",
      "x-amz-meta-audit-archive-id": normalized.archive_id,
      "x-amz-meta-audit-manifest-sha256": normalized.manifest_hash,
      "x-amz-meta-audit-content-sha256": contentHash
    };
    if (objectLockMode) headers["x-amz-object-lock-mode"] = objectLockMode;
    if (retainUntil) headers["x-amz-object-lock-retain-until-date"] = retainUntil;
    if (authorization) {
      headers.Authorization = authorization;
    } else if (token) {
      headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    }
    const response = await fetchImpl(uploadURL, { method: "PUT", headers, body });
    if (!response || response.ok === false) {
      const text = response && typeof response.text === "function" ? await response.text() : "";
      throw new Error(text || `audit archive upload failed: ${response?.status || 0}`);
    }
    return {
      archive_id: normalized.archive_id,
      storage_key: normalized.storage_key,
      upload_url: uploadURL,
      manifest_hash: normalized.manifest_hash,
      content_hash: contentHash,
      object_lock_mode: objectLockMode,
      retain_until: retainUntil,
      bytes: Buffer.byteLength(body)
    };
  };
}

export async function archiveAuditEvent(event = {}, options = {}) {
  const archive = normalizeArchiveRequest(event);
  if (!archive.archive_id || !archive.storage_key || !archive.manifest_hash) {
    throw new Error("audit archive event must include archive_id, storage_key and manifest_hash");
  }
  if (archive.manifest_entries.length !== archive.log_count) {
    throw new Error("audit archive manifest entry count mismatch");
  }
  const computedHash = computeArchiveManifestHash(archive);
  if (computedHash !== archive.manifest_hash) {
    throw new Error("audit archive manifest hash mismatch");
  }
  const uploader = options.uploader || createAuditArchiveUploader(options);
  return uploader(archive);
}

export function createAuditArchiveConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event) => archiveAuditEvent(event, options))
  });
}

async function parseJSONResponse(response, fallbackMessage = "api request failed") {
  const text = typeof response.text === "function" ? await response.text() : "";
  const body = text ? JSON.parse(text) : {};
  if (!response.ok || body.success === false) {
    throw new Error(body.message || `${fallbackMessage}: ${response.status}`);
  }
  return body.data ?? body;
}

export function createAuditArchiveApiClient(options = {}) {
  const apiBaseUrl = clean(options.apiBaseUrl || process.env.API_BASE_URL || "http://127.0.0.1:1029").replace(/\/+$/, "");
  const token = clean(options.token || process.env.AUDIT_ARCHIVE_WORKER_TOKEN || "");
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  function headers(extra = {}) {
    const output = { "Content-Type": "application/json", ...extra };
    if (token) output.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    return output;
  }
  function request(path, requestOptions = {}) {
    return fetchImpl(`${apiBaseUrl}${path}`, {
      method: requestOptions.method || "GET",
      headers: headers(requestOptions.headers),
      body: requestOptions.body === undefined ? undefined : JSON.stringify(requestOptions.body)
    }).then((response) => parseJSONResponse(response, "audit archive api request failed"));
  }
  return {
    async claimEvents({ topic = subscribedTopics[0], limit = defaultBatchLimit, leaseOwner = workerName, leaseSeconds = defaultLeaseSeconds, now = new Date() } = {}) {
      return request("/api/admin/outbox/events/claim", {
        method: "POST",
        body: {
          topic,
          limit: positiveInteger(limit, defaultBatchLimit),
          lease_owner: leaseOwner,
          lease_seconds: positiveInteger(leaseSeconds, defaultLeaseSeconds),
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
    async completeArchive(completion = {}) {
      return request("/api/admin/audit-logs/archive/complete", {
        method: "POST",
        body: {
          archive_id: clean(completion.archive_id || completion.archiveId),
          storage_key: clean(completion.storage_key || completion.storageKey),
          manifest_algorithm: clean(completion.manifest_algorithm || completion.manifestAlgorithm || "sha256:v1"),
          manifest_hash: clean(completion.manifest_hash || completion.manifestHash),
          content_hash: clean(completion.content_hash || completion.contentHash),
          bytes: Number(completion.bytes || 0),
          object_lock_mode: clean(completion.object_lock_mode || completion.objectLockMode),
          retain_until: clean(completion.retain_until || completion.retainUntil),
          outbox_event_id: clean(completion.outbox_event_id || completion.outboxEventId),
          uploaded_at: completion.uploaded_at instanceof Date ? completion.uploaded_at.toISOString() : clean(completion.uploaded_at || completion.uploadedAt)
        }
      });
    },
    async markFailed(eventID, error, retryAfterSeconds = defaultRetryAfterSeconds, now = new Date(), maxAttempts = defaultMaxAttempts) {
      const body = {
        error: clean(error).slice(0, 500),
        retry_after_seconds: positiveInteger(retryAfterSeconds, defaultRetryAfterSeconds),
        now: now instanceof Date ? now.toISOString() : now
      };
      if (Number(maxAttempts) > 0) {
        body.max_attempts = positiveInteger(maxAttempts, defaultMaxAttempts);
      }
      return request(`/api/admin/outbox/events/${encodeURIComponent(eventID)}/failed`, {
        method: "POST",
        body
      });
    }
  };
}

export function buildAuditArchiveCompletion(event = {}, archived = {}, options = {}) {
  const archive = normalizeArchiveRequest(event);
  const uploadedAt = options.uploadedAt || options.publishedAt || options.now || new Date();
  return {
    archive_id: archive.archive_id,
    storage_key: archive.storage_key,
    manifest_algorithm: archive.manifest_algorithm,
    manifest_hash: archive.manifest_hash,
    content_hash: clean(archived.content_hash || archived.contentHash),
    bytes: Number(archived.bytes || 0),
    object_lock_mode: clean(archived.object_lock_mode || archived.objectLockMode || options.objectLockMode || process.env.AUDIT_ARCHIVE_OBJECT_LOCK_MODE || defaultObjectLockMode),
    retain_until: clean(archived.retain_until || archived.retainUntil || options.retainUntil || process.env.AUDIT_ARCHIVE_RETAIN_UNTIL_DATE),
    outbox_event_id: clean(event.id || event.outbox_event_id || event.outboxEventId || archive.outbox_event_id),
    uploaded_at: uploadedAt instanceof Date ? uploadedAt.toISOString() : clean(uploadedAt)
  };
}

export async function archiveOutboxBatch(options = {}) {
  const client = options.client || createAuditArchiveApiClient(options);
  if (!client || typeof client.claimEvents !== "function" || typeof client.markPublished !== "function" || typeof client.markFailed !== "function") {
    throw new Error("audit archive client with claimEvents, markPublished and markFailed is required");
  }
  const topics = normalizeTopicList(options.topics || process.env.AUDIT_ARCHIVE_TOPICS, subscribedTopics);
  const limit = positiveInteger(options.limit || process.env.AUDIT_ARCHIVE_BATCH_LIMIT, defaultBatchLimit);
  const retryAfterSeconds = positiveInteger(options.retryAfterSeconds || process.env.AUDIT_ARCHIVE_RETRY_AFTER_SECONDS, defaultRetryAfterSeconds);
  const maxAttempts = positiveInteger(options.maxAttempts || process.env.AUDIT_ARCHIVE_MAX_ATTEMPTS, defaultMaxAttempts);
  const leaseOwner = clean(options.leaseOwner || process.env.AUDIT_ARCHIVE_WORKER_ID || `${workerName}-${process.pid}`) || workerName;
  const leaseSeconds = positiveInteger(options.leaseSeconds || process.env.AUDIT_ARCHIVE_LEASE_SECONDS, defaultLeaseSeconds);
  const now = options.now || new Date();
  const publishedAt = options.publishedAt || now;
  const result = {
    claimed: 0,
    archived: 0,
    failed: 0,
    results: [],
    errors: []
  };
  for (const topic of topics) {
    const claimResult = await client.claimEvents({ topic, limit, leaseOwner, leaseSeconds, now });
    const events = Array.isArray(claimResult) ? claimResult : claimResult.events || [];
    result.claimed += events.length;
    for (const event of events) {
      const eventID = clean(event.id);
      try {
        const archived = await archiveAuditEvent(event, options);
        if (typeof client.completeArchive === "function") {
          await client.completeArchive(buildAuditArchiveCompletion(event, archived, {
            ...options,
            uploadedAt: options.uploadedAt || publishedAt
          }));
        }
        await client.markPublished(eventID, publishedAt);
        result.archived++;
        result.results.push({ event_id: eventID, archived });
      } catch (error) {
        const message = String(error?.message || error);
        if (eventID) {
          await client.markFailed(eventID, message, retryAfterSeconds, now, maxAttempts);
        }
        result.failed++;
        result.errors.push({ id: eventID, message });
      }
    }
  }
  return result;
}

export function createArchiveLoop(options = {}) {
  const intervalMs = positiveInteger(options.intervalMs || process.env.AUDIT_ARCHIVE_INTERVAL_MS, defaultPollIntervalMs);
  const setIntervalImpl = options.setIntervalImpl || setInterval;
  const clearIntervalImpl = options.clearIntervalImpl || clearInterval;
  const onResult = options.onResult;
  const onError = options.onError;
  const tick = () => archiveOutboxBatch(options)
    .then((result) => {
      if (typeof onResult === "function") onResult(result);
      return result;
    })
    .catch((error) => {
      if (typeof onError === "function") onError(error);
      return { claimed: 0, archived: 0, failed: 1, results: [], errors: [{ message: String(error?.message || error) }] };
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
  const loop = createArchiveLoop({
    intervalMs: process.env.AUDIT_ARCHIVE_INTERVAL_MS,
    onResult(result) {
      console.log(`${workerName} archive tick; claimed=${result.claimed} archived=${result.archived} failed=${result.failed}`);
    },
    onError(error) {
      console.error(`${workerName} archive tick failed: ${error.message}`);
    }
  });
  loop.tick();
  process.once("SIGINT", () => loop.stop());
  process.once("SIGTERM", () => loop.stop());
}
