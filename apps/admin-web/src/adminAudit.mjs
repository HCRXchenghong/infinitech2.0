const SENSITIVE_KEY_PATTERN = /(password|secret|token|authorization|openid|session|credential|certificate|phone|mobile|email|id_card|identity|file_url|object_key|signature|pay_sign|nonce)/i;
const PAYLOAD_ALLOWLIST = Object.freeze([
  "action_filter",
  "actor_id",
  "actor_type",
  "after",
  "amount_fen",
  "applied_scopes",
  "archive_id",
  "archive_id_matched",
  "actual_bytes",
  "actual_content_hash",
  "before",
  "bytes",
  "bytes_matched",
  "change_request_id",
  "content_hash",
  "content_hash_matched",
  "current_scopes",
  "decision",
  "default_refund_strategy",
  "error_code",
  "error_message",
  "event_id",
  "expected_bytes",
  "expected_content_hash",
  "export_format",
  "generated_at",
  "header_log_count",
  "idempotency_key",
  "limit",
  "log_count_matched",
  "manifest_algorithm",
  "manifest_entry_count",
  "manifest_hash",
  "manifest_hash_matched",
  "max_attempts",
  "policy_version",
  "previous_scopes",
  "reason",
  "refund_id",
  "requested_scopes",
  "retry_after_seconds",
  "role",
  "rollback_from_scopes",
  "rollback_to_scopes",
  "row_count",
  "status",
  "storage_key",
  "topic",
  "type",
  "verified_at"
]);

export const AUDIT_FILTER_DEFAULTS = Object.freeze({
  actor_type: "",
  actor_id: "",
  action: "",
  target_type: "",
  target_id: "",
  after: "",
  before: "",
  limit: 20
});

const AUDIT_TARGET_ROUTES = Object.freeze({
  after_sales: { module: "after-sales", operation: "after-sales-list", label: "售后审核" },
  merchant: { module: "merchants", operation: "merchant-invite", label: "商户资质" },
  merchant_invite: { module: "merchants", operation: "merchant-invite", label: "商户资质" },
  admin_rbac_role: { module: "permissions", operation: "rbac-policy", label: "权限治理" },
  admin_rbac_change_request: { module: "permissions", operation: "rbac-change-requests", label: "权限申请" },
  object_storage_ticket: { module: "after-sales", operation: "object-cleanup-candidates", label: "对象清理" },
  order: { module: "orders", operation: "order-compensate", label: "订单监控" },
  outbox_event: { module: "dashboard", operation: "outbox-events", label: "Outbox 事件" },
  outbox_topic: { module: "dashboard", operation: "outbox-stats", label: "Outbox 健康" },
  refund_settings: { module: "refund-settings", operation: "refund-settings-read", label: "退款策略" },
  rider: { module: "riders", operation: "station-riders", label: "骑手/站长" },
  rider_invite: { module: "riders", operation: "rider-invite", label: "骑手/站长" },
  station_manager: { module: "riders", operation: "station-riders", label: "骑手/站长" }
});

function compact(value, fallback = "-") {
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return String(value);
}

function trim(value) {
  return String(value ?? "").trim();
}

function formatTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  });
}

function maskScalar(value) {
  const text = compact(value, "");
  if (!text) {
    return "";
  }
  if (text.length <= 6) {
    return "***";
  }
  return `${text.slice(0, 3)}***${text.slice(-2)}`;
}

function clampLimit(value) {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return AUDIT_FILTER_DEFAULTS.limit;
  }
  return Math.min(500, Math.max(1, Math.trunc(numberValue)));
}

function hashText(text) {
  let hash = 0;
  for (let index = 0; index < text.length; index += 1) {
    hash = (hash * 31 + text.charCodeAt(index)) >>> 0;
  }
  return hash.toString(36);
}

function auditIntegrityState(log = {}) {
  const algorithm = trim(log.integrity_algorithm);
  const hash = trim(log.integrity_hash);
  if (log.integrity_verified === true) {
    return {
      label: "已验证",
      tone: "ok",
      algorithm: algorithm || "未返回",
      hash,
      hashShort: hash ? `${hash.slice(0, 12)}...` : "-"
    };
  }
  if (algorithm || hash) {
    return {
      label: "未通过",
      tone: "danger",
      algorithm: algorithm || "未返回",
      hash,
      hashShort: hash ? `${hash.slice(0, 12)}...` : "-"
    };
  }
  return {
    label: "未签封",
    tone: "muted",
    algorithm: "未返回",
    hash: "",
    hashShort: "-"
  };
}

export function redactAuditValue(key, value) {
  if (SENSITIVE_KEY_PATTERN.test(String(key || ""))) {
    return maskScalar(value);
  }
  if (Array.isArray(value)) {
    return value.map((item) => {
      if (item && typeof item === "object") {
        return redactAuditPayload(item);
      }
      return item;
    });
  }
  if (value && typeof value === "object") {
    return redactAuditPayload(value);
  }
  return value;
}

export function redactAuditPayload(payload = {}) {
  const redacted = {};
  for (const [key, value] of Object.entries(payload || {})) {
    redacted[key] = redactAuditValue(key, value);
  }
  return redacted;
}

export function summarizeAuditPayload(payload = {}) {
  const redacted = redactAuditPayload(payload);
  const parts = [];
  for (const key of PAYLOAD_ALLOWLIST) {
    if (Object.prototype.hasOwnProperty.call(redacted, key)) {
      parts.push(`${key}: ${compact(redacted[key])}`);
    }
  }
  if (parts.length === 0) {
    const visibleKeys = Object.keys(redacted).filter((key) => !SENSITIVE_KEY_PATTERN.test(key)).slice(0, 3);
    for (const key of visibleKeys) {
      parts.push(`${key}: ${compact(redacted[key])}`);
    }
  }
  return parts.join(" / ") || "无公开摘要";
}

export function normalizeAuditFilters(filters = {}) {
  return {
    actor_type: trim(filters.actor_type),
    actor_id: trim(filters.actor_id),
    action: trim(filters.action),
    target_type: trim(filters.target_type),
    target_id: trim(filters.target_id),
    after: trim(filters.after),
    before: trim(filters.before),
    limit: clampLimit(filters.limit ?? AUDIT_FILTER_DEFAULTS.limit)
  };
}

export function auditSearchValuesFromFilters(filters = {}, { beforeOverride = "" } = {}) {
  const normalized = normalizeAuditFilters(filters);
  return {
    ...normalized,
    before: beforeOverride || normalized.before
  };
}

export function describeAuditFilters(filters = {}) {
  const normalized = normalizeAuditFilters(filters);
  const parts = [];
  if (normalized.actor_type) parts.push(`actor:${normalized.actor_type}`);
  if (normalized.actor_id) parts.push(`actor_id:${normalized.actor_id}`);
  if (normalized.action) parts.push(`action:${normalized.action}`);
  if (normalized.target_type) parts.push(`target:${normalized.target_type}`);
  if (normalized.target_id) parts.push(`target_id:${normalized.target_id}`);
  if (normalized.after) parts.push(`after:${normalized.after}`);
  if (normalized.before) parts.push(`before:${normalized.before}`);
  return parts.slice(0, 3).join(" / ") || "全部审计";
}

export function makeAuditFilterPreset(filters = {}, createdAt = new Date().toISOString()) {
  const normalized = normalizeAuditFilters(filters);
  const id = `audit_filter_${hashText(JSON.stringify(normalized))}`;
  return {
    id,
    name: describeAuditFilters(normalized),
    filters: normalized,
    created_at: createdAt
  };
}

export function upsertAuditFilterPreset(presets = [], preset, maxPresets = 8) {
  const next = [preset, ...presets.filter((item) => item?.id !== preset.id)];
  return next.slice(0, maxPresets);
}

export function auditTargetRoute(log = {}) {
  const targetType = trim(log.target_type);
  if (AUDIT_TARGET_ROUTES[targetType]) {
    return AUDIT_TARGET_ROUTES[targetType];
  }
  const action = trim(log.action);
  if (action.includes("after_sales")) return AUDIT_TARGET_ROUTES.after_sales;
  if (action.includes("rbac")) return AUDIT_TARGET_ROUTES.admin_rbac_role;
  if (action.includes("outbox")) return AUDIT_TARGET_ROUTES.outbox_event;
  if (action.includes("refund")) return AUDIT_TARGET_ROUTES.refund_settings;
  if (action.includes("rider") || action.includes("station")) return AUDIT_TARGET_ROUTES.rider;
  if (action.includes("merchant")) return AUDIT_TARGET_ROUTES.merchant;
  return { module: "audit-logs", operation: "audit-logs", label: "审计检索" };
}

export function buildAuditRows(logs = []) {
  return logs.map((log) => {
    const route = auditTargetRoute(log);
    const integrity = auditIntegrityState(log);
    return {
      id: compact(log.id),
      actorType: compact(log.actor_type, ""),
      actorId: compact(log.actor_id, ""),
      actor: `${compact(log.actor_type)}:${compact(log.actor_id)}`,
      action: compact(log.action),
      targetType: compact(log.target_type, ""),
      targetId: compact(log.target_id, ""),
      target: `${compact(log.target_type)}:${compact(log.target_id)}`,
      requestId: compact(log.request_id, ""),
      ipHash: compact(log.ip_hash, ""),
      request: compact(log.request_id || log.ip_hash),
      createdAt: formatTime(log.created_at),
      createdRaw: log.created_at || "",
      before: log.created_at || "",
      payloadSummary: summarizeAuditPayload(log.payload || {}),
      payload: redactAuditPayload(log.payload || {}),
      integrityLabel: integrity.label,
      integrityTone: integrity.tone,
      integrityAlgorithm: integrity.algorithm,
      integrityHash: integrity.hash,
      integrityHashShort: integrity.hashShort,
      targetModule: route.module,
      targetOperation: route.operation,
      targetLabel: route.label
    };
  });
}

export function nextAuditBefore(logs = []) {
  const last = logs[logs.length - 1];
  return last?.created_at || "";
}

export function auditDataFromResult(result) {
  const data = result?.payload?.data;
  return Array.isArray(data) ? data : [];
}

export function auditExportDataFromResult(result) {
  const data = result?.payload?.data;
  if (!data || typeof data !== "object" || Array.isArray(data)) {
    return null;
  }
  return {
    format: compact(data.format, ""),
    filename: compact(data.filename, ""),
    contentType: compact(data.content_type, "text/csv; charset=utf-8"),
    rowCount: Number(data.row_count || 0),
    csv: compact(data.csv, ""),
    generatedAt: compact(data.generated_at, "")
  };
}

export function auditRetentionReportFromResult(result) {
  const data = result?.payload?.data;
  if (!data || typeof data !== "object" || Array.isArray(data)) {
    return null;
  }
  const alerts = Array.isArray(data.alerts) ? data.alerts : [];
  const missingCriticalActions = Array.isArray(data.missing_critical_actions) ? data.missing_critical_actions : [];
  return {
    status: compact(data.status, "unknown"),
    retentionDays: Number(data.retention_days || 0),
    hotDays: Number(data.hot_days || 0),
    totalLogs: Number(data.total_logs || 0),
    expiredLogs: Number(data.expired_logs || 0),
    coldArchiveDueLogs: Number(data.cold_archive_due_logs || 0),
    integritySampleSize: Number(data.integrity_sample_size || 0),
    integrityFailures: Number(data.integrity_failures || 0),
    exportEvents: Number(data.export_events || 0),
    missingCriticalActions,
    alerts
  };
}

export function auditRetentionAlertEmissionFromResult(result) {
  const data = result?.payload?.data;
  const emission = data?.emission;
  if (!emission || typeof emission !== "object" || Array.isArray(emission)) {
    return null;
  }
  return {
    status: compact(emission.status, "unknown"),
    reportStatus: compact(emission.report_status, "unknown"),
    alertCount: Number(emission.alert_count || 0),
    criticalCount: Number(emission.critical_count || 0),
    warningCount: Number(emission.warning_count || 0),
    topic: compact(emission.topic, ""),
    outboxEventId: compact(emission.outbox_event_id, ""),
    auditLogId: compact(data.audit_log?.id, "")
  };
}

export function auditArchiveRequestFromResult(result) {
  const data = result?.payload?.data;
  const archive = data?.archive;
  if (!archive || typeof archive !== "object" || Array.isArray(archive)) {
    return null;
  }
  return {
    archiveId: compact(archive.archive_id, ""),
    status: compact(archive.status, "unknown"),
    topic: compact(archive.topic, ""),
    storageKey: compact(archive.storage_key, ""),
    logCount: Number(archive.log_count || 0),
    integrityFailures: Number(archive.integrity_failures || 0),
    manifestAlgorithm: compact(archive.manifest_algorithm, ""),
    manifestHash: compact(archive.manifest_hash, ""),
    outboxEventId: compact(archive.outbox_event_id, ""),
    auditLogId: compact(data.audit_log?.id, "")
  };
}

export function auditArchiveRecordsFromResult(result) {
  const data = result?.payload?.data;
  if (!Array.isArray(data)) {
    return [];
  }
  return data
    .filter((item) => item && typeof item === "object" && !Array.isArray(item))
    .map((item) => ({
      archiveId: compact(item.archive_id, ""),
      status: compact(item.status, "unknown"),
      storageKey: compact(item.storage_key, ""),
      manifestHash: compact(item.manifest_hash, ""),
      contentHash: compact(item.content_hash, ""),
      bytes: Number(item.bytes || 0),
      objectLockMode: compact(item.object_lock_mode, ""),
      retainUntil: compact(item.retain_until, ""),
      uploadedAt: compact(item.uploaded_at, ""),
      completedAt: compact(item.completed_at, ""),
      outboxEventId: compact(item.outbox_event_id, "")
    }));
}

export function auditArchiveVerificationFromResult(result) {
  const data = result?.payload?.data;
  const verification = data?.verification;
  if (!verification || typeof verification !== "object" || Array.isArray(verification)) {
    return null;
  }
  return {
    archiveId: compact(verification.archive_id, ""),
    status: compact(verification.status, "unknown"),
    storageKey: compact(verification.storage_key, ""),
    manifestHash: compact(verification.manifest_hash, ""),
    expectedContentHash: compact(verification.expected_content_hash, ""),
    actualContentHash: compact(verification.actual_content_hash, ""),
    expectedBytes: Number(verification.expected_bytes || 0),
    actualBytes: Number(verification.actual_bytes || 0),
    archiveIdMatched: verification.archive_id_matched === true,
    manifestHashMatched: verification.manifest_hash_matched === true,
    contentHashMatched: verification.content_hash_matched === true,
    bytesMatched: verification.bytes_matched === true,
    logCountMatched: verification.log_count_matched === true,
    headerLogCount: Number(verification.header_log_count || 0),
    manifestEntryCount: Number(verification.manifest_entry_count || 0),
    errorCode: compact(verification.error_code, ""),
    verifiedAt: compact(verification.verified_at, ""),
    auditLogId: compact(data.audit_log?.id, "")
  };
}
