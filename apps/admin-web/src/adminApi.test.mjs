import assert from "node:assert/strict";
import test from "node:test";
import { ADMIN_API_OPERATIONS, buildAdminRequest, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import {
  auditDataFromResult,
  auditArchiveRequestFromResult,
  auditArchiveRecordsFromResult,
  auditArchiveVerificationFromResult,
  auditArchiveVerificationsFromResult,
  auditExportDataFromResult,
  auditRetentionAlertEmissionFromResult,
  auditRetentionReportFromResult,
  auditSearchValuesFromFilters,
  auditTargetRoute,
  buildAuditRows,
  makeAuditFilterPreset,
  nextAuditBefore,
  normalizeAuditFilters,
  redactAuditPayload,
  summarizeAuditPayload,
  upsertAuditFilterPreset
} from "./adminAudit.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS, getAdminWebModule } from "./config.mjs";
import { ADMIN_WEB_VIEWS, getAdminView } from "./adminViews.mjs";
import { applySnapshotToAdminView, buildSnapshotKpis, buildSnapshotQueues, snapshotDataFromResult } from "./adminSnapshot.mjs";

test("admin web exposes the first operable control-center modules", () => {
  for (const key of ["orders", "after-sales", "merchants", "riders", "dispatch", "audit-logs", "refund-settings", "payment", "support", "rtc", "integrations", "permissions"]) {
    assert.ok(getAdminWebModule(key), `missing ${key}`);
  }
  assert.ok(ADMIN_WEB_SECTIONS.length >= 4);
  assert.ok(ADMIN_WEB_MODULES.length >= 25);
  assert.ok(ADMIN_WEB_KPIS.some((item) => item.key === "outbox"));
  assert.ok(ADMIN_WEB_QUEUES.some((item) => item.operationKey === "object-cleanup-stats"));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "finance_admin" && item.scopes.includes("refund:write")));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "security_auditor" && item.scopes.includes("audit:read")));
  assert.ok(ADMIN_WEB_RBAC.every((item) => item.scopes.includes("*") || item.scopes.includes("rbac:read")));
});

test("admin web operation catalog covers shipped admin API surfaces", () => {
  for (const key of [
    "admin-login",
    "merchant-invite",
    "station-manager-invite",
    "rider-invite",
    "refund-settings-read",
    "refund-settings-save",
    "after-sales-list",
    "operations-snapshot",
    "audit-logs",
    "audit-logs-export",
    "audit-retention-report",
    "audit-retention-alert-emit",
    "audit-archive-request",
    "audit-archive-records",
    "audit-archive-verify",
    "audit-archive-verifications",
    "rbac-policy",
    "rbac-change-requests",
    "rbac-change-request",
    "rbac-review-request",
    "rbac-apply-request",
    "rbac-rollback-request",
    "object-cleanup-stats",
    "object-cleanup-candidates",
    "outbox-stats",
    "outbox-events",
    "outbox-replay-batch",
    "order-compensate",
    "station-riders",
    "station-orders",
    "station-performance",
    "station-task-config"
  ]) {
    assert.ok(getAdminOperation(key), `missing ${key}`);
  }
  assert.ok(ADMIN_API_OPERATIONS.every((operation) => operation.method && operation.path && operation.title));
});

test("admin web ships P0 business views with actions and safeguards", () => {
  for (const key of ["orders", "after-sales", "merchants", "riders", "rider-performance", "dispatch", "audit-logs", "refund-settings", "permissions"]) {
    const view = getAdminView(key);
    assert.equal(view.key, key);
    assert.ok(view.metrics.length >= 4, `missing metrics for ${key}`);
    assert.ok(view.columns.length >= 4, `missing columns for ${key}`);
    assert.ok(view.rows.length >= 4, `missing rows for ${key}`);
    assert.ok(view.safeguards.length >= 3, `missing safeguards for ${key}`);
  }
  assert.ok(ADMIN_WEB_VIEWS.orders.actions.includes("order-compensate"));
  assert.ok(ADMIN_WEB_VIEWS.merchants.actions.includes("merchant-invite"));
  assert.ok(ADMIN_WEB_VIEWS.riders.actions.includes("station-riders"));
  assert.ok(ADMIN_WEB_VIEWS.dispatch.actions.includes("station-orders"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-logs-export"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-retention-report"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-retention-alert-emit"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-request"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-records"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-verify"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-verifications"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-change-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-review-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-apply-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-rollback-request"));
});

test("admin request builder normalizes auth query path and body", () => {
  const operation = getAdminOperation("order-compensate");
  const request = buildAdminRequest(operation, { order_id: "ord 1" }, "admin.token");
  assert.equal(request.method, "POST");
  assert.equal(request.url, "/api/admin/orders/ord%201/state/compensate");
  assert.equal(request.headers.Authorization, "Bearer admin.token");
  assert.equal(request.body, "{}");

  const statsRequest = buildAdminRequest(getAdminOperation("outbox-stats"), { topic: "order.paid", lease_expiring_within_seconds: 30 }, "Bearer admin.token");
  assert.equal(statsRequest.url, "/api/admin/outbox/stats?topic=order.paid&lease_expiring_within_seconds=30");

  const snapshotRequest = buildAdminRequest(getAdminOperation("operations-snapshot"), { limit: 10, lease_expiring_within_seconds: 45 }, "admin.token");
  assert.equal(snapshotRequest.url, "/api/admin/operations/snapshot?limit=10&lease_expiring_within_seconds=45&object_cleanup_grace_seconds=3600");

  const auditRequest = buildAdminRequest(getAdminOperation("audit-logs"), { actor_type: "admin", actor_id: "admin_1", target_type: "order", after: "2026-05-22T00:00:00Z", before: "2026-05-22T12:00:00Z", limit: 5 }, "Bearer admin.token");
  assert.equal(auditRequest.url, "/api/admin/audit-logs?actor_type=admin&actor_id=admin_1&target_type=order&after=2026-05-22T00%3A00%3A00Z&before=2026-05-22T12%3A00%3A00Z&limit=5");
  assert.equal(auditRequest.headers.Authorization, "Bearer admin.token");

  const auditExportRequest = buildAdminRequest(getAdminOperation("audit-logs-export"), { actor_type: "admin", target_type: "order", action: "admin.order.refunded", limit: 100 }, "Bearer admin.token");
  assert.equal(auditExportRequest.url, "/api/admin/audit-logs/export?actor_type=admin&target_type=order&action=admin.order.refunded&limit=100");
  assert.equal(auditExportRequest.headers.Authorization, "Bearer admin.token");

  const auditRetentionRequest = buildAdminRequest(getAdminOperation("audit-retention-report"), { retention_days: 2555, hot_days: 180, integrity_sample_limit: 500 }, "Bearer admin.token");
  assert.equal(auditRetentionRequest.url, "/api/admin/audit-logs/retention-report?retention_days=2555&hot_days=180&integrity_sample_limit=500");
  assert.equal(auditRetentionRequest.headers.Authorization, "Bearer admin.token");

  const auditRetentionAlertRequest = buildAdminRequest(getAdminOperation("audit-retention-alert-emit"), { retention_days: 2555, hot_days: 180, integrity_sample_limit: 500 }, "Bearer admin.token");
  assert.equal(auditRetentionAlertRequest.method, "POST");
  assert.equal(auditRetentionAlertRequest.url, "/api/admin/audit-logs/retention-alerts/emit");
  assert.deepEqual(JSON.parse(auditRetentionAlertRequest.body), { retention_days: 2555, hot_days: 180, integrity_sample_limit: 500 });

  const auditArchiveRequest = buildAdminRequest(getAdminOperation("audit-archive-request"), { hot_days: 180, limit: 500, storage_prefix: "worm://audit-logs" }, "Bearer admin.token");
  assert.equal(auditArchiveRequest.method, "POST");
  assert.equal(auditArchiveRequest.url, "/api/admin/audit-logs/archive/request");
  assert.deepEqual(JSON.parse(auditArchiveRequest.body), { hot_days: 180, limit: 500, storage_prefix: "worm://audit-logs" });

  const auditArchiveRecordsRequest = buildAdminRequest(getAdminOperation("audit-archive-records"), { archive_id: "audit_archive_1", limit: 5 }, "Bearer admin.token");
  assert.equal(auditArchiveRecordsRequest.method, "GET");
  assert.equal(auditArchiveRecordsRequest.url, "/api/admin/audit-logs/archive/records?archive_id=audit_archive_1&limit=5");

  const auditArchiveVerifyRequest = buildAdminRequest(getAdminOperation("audit-archive-verify"), { archive_id: "audit_archive_1" }, "Bearer admin.token");
  assert.equal(auditArchiveVerifyRequest.method, "POST");
  assert.equal(auditArchiveVerifyRequest.url, "/api/admin/audit-logs/archive/verify");
  assert.deepEqual(JSON.parse(auditArchiveVerifyRequest.body), { archive_id: "audit_archive_1" });

  const auditArchiveVerificationsRequest = buildAdminRequest(getAdminOperation("audit-archive-verifications"), { archive_id: "audit_archive_1", status: "verified", limit: 5 }, "Bearer admin.token");
  assert.equal(auditArchiveVerificationsRequest.method, "GET");
  assert.equal(auditArchiveVerificationsRequest.url, "/api/admin/audit-logs/archive/verifications?archive_id=audit_archive_1&status=verified&limit=5");

  const rbacRequest = buildAdminRequest(getAdminOperation("rbac-change-request"), { role: "support_admin", requested_scopes: "after_sales:read, rbac:read, rbac:read", reason: "support recertification" }, "admin.token");
  assert.equal(rbacRequest.url, "/api/admin/rbac/change-requests");
  assert.deepEqual(JSON.parse(rbacRequest.body), { role: "support_admin", requested_scopes: ["after_sales:read", "rbac:read"], reason: "support recertification" });

  const rbacList = buildAdminRequest(getAdminOperation("rbac-change-requests"), { status: "pending_approval", limit: 5 }, "admin.token");
  assert.equal(rbacList.url, "/api/admin/rbac/change-requests?status=pending_approval&limit=5");

  const rbacReview = buildAdminRequest(getAdminOperation("rbac-review-request"), { change_request_id: "rbac change 1", decision: "reject", reason: "scope too broad" }, "admin.token");
  assert.equal(rbacReview.url, "/api/admin/rbac/change-requests/rbac%20change%201/review");
  assert.deepEqual(JSON.parse(rbacReview.body), { decision: "reject", reason: "scope too broad" });

  const rbacApply = buildAdminRequest(getAdminOperation("rbac-apply-request"), { change_request_id: "rbac change 1", reason: "approved runtime apply" }, "admin.token");
  assert.equal(rbacApply.url, "/api/admin/rbac/change-requests/rbac%20change%201/apply");
  assert.deepEqual(JSON.parse(rbacApply.body), { reason: "approved runtime apply" });

  const rbacRollback = buildAdminRequest(getAdminOperation("rbac-rollback-request"), { change_request_id: "rbac change 1", reason: "restore runtime policy" }, "admin.token");
  assert.equal(rbacRollback.url, "/api/admin/rbac/change-requests/rbac%20change%201/rollback");
  assert.deepEqual(JSON.parse(rbacRollback.body), { reason: "restore runtime policy" });

  const loginFields = fieldsForOperation(getAdminOperation("admin-login"));
  assert.deepEqual(loginFields.map((field) => field.key), ["account_id", "password"]);
});

test("admin audit adapter redacts sensitive payload and builds cursor rows", () => {
  const logs = [
    {
      id: "aud_2",
      actor_type: "admin",
      actor_id: "admin_1",
      action: "admin.order.refunded",
      target_type: "order",
      target_id: "ord_1",
      request_id: "req_1",
      ip_hash: "ip_hash",
      payload: {
        amount_fen: 1200,
        idempotency_key: "refund_ord_1",
        password: "PlainTextPassword",
        token: "secret-token-value",
        object_key: "after-sales/asr_1/private/evidence.jpg",
        nested: { phone: "13900000000", reason: "商品售罄" }
      },
      integrity_algorithm: "hmac-sha256:v1",
      integrity_hash: "abcdef1234567890abcdef1234567890",
      integrity_verified: true,
      created_at: "2026-05-22T12:00:00Z"
    }
  ];
  const redacted = redactAuditPayload(logs[0].payload);
  assert.equal(redacted.password, "Pla***rd");
  assert.equal(redacted.token, "sec***ue");
  assert.equal(redacted.object_key, "aft***pg");
  assert.equal(redacted.nested.phone, "139***00");
  assert.equal(redacted.nested.reason, "商品售罄");

  const summary = summarizeAuditPayload(logs[0].payload);
  assert.match(summary, /amount_fen: 1200/);
  assert.match(summary, /idempotency_key: refund_ord_1/);
  assert.doesNotMatch(summary, /PlainTextPassword|secret-token-value|private\/evidence/);

  const rows = buildAuditRows(logs);
  assert.equal(rows[0].actor, "admin:admin_1");
  assert.equal(rows[0].target, "order:ord_1");
  assert.equal(rows[0].targetModule, "orders");
  assert.equal(rows[0].targetOperation, "order-compensate");
  assert.equal(rows[0].targetLabel, "订单监控");
  assert.equal(rows[0].before, logs[0].created_at);
  assert.equal(rows[0].integrityLabel, "已验证");
  assert.equal(rows[0].integrityTone, "ok");
  assert.equal(rows[0].integrityAlgorithm, "hmac-sha256:v1");
  assert.equal(rows[0].integrityHashShort, "abcdef123456...");
  assert.equal(nextAuditBefore(logs), logs[0].created_at);
  assert.deepEqual(auditDataFromResult({ payload: { data: logs } }), logs);
  assert.deepEqual(auditDataFromResult({ payload: { data: { nope: true } } }), []);
  assert.deepEqual(auditExportDataFromResult({ payload: { data: { format: "csv", filename: "audit.csv", row_count: 1, csv: "id\\naud_1\\n", generated_at: "2026-05-22T12:00:00Z" } } }), {
    format: "csv",
    filename: "audit.csv",
    contentType: "text/csv; charset=utf-8",
    rowCount: 1,
    csv: "id\\naud_1\\n",
    generatedAt: "2026-05-22T12:00:00Z"
  });
  assert.equal(auditExportDataFromResult({ payload: { data: [] } }), null);
  assert.deepEqual(auditRetentionReportFromResult({ payload: { data: { status: "warning", retention_days: 2555, hot_days: 180, total_logs: 4, expired_logs: 0, cold_archive_due_logs: 1, integrity_sample_size: 4, integrity_failures: 0, export_events: 1, missing_critical_actions: ["after_sales.reviewed"], alerts: [{ code: "audit.missing_critical_action", severity: "warning" }] } } }), {
    status: "warning",
    retentionDays: 2555,
    hotDays: 180,
    totalLogs: 4,
    expiredLogs: 0,
    coldArchiveDueLogs: 1,
    integritySampleSize: 4,
    integrityFailures: 0,
    exportEvents: 1,
    missingCriticalActions: ["after_sales.reviewed"],
    alerts: [{ code: "audit.missing_critical_action", severity: "warning" }]
  });
  assert.equal(auditRetentionReportFromResult({ payload: { data: [] } }), null);
  assert.deepEqual(auditRetentionAlertEmissionFromResult({ payload: { data: { emission: { status: "emitted", report_status: "warning", alert_count: 1, critical_count: 0, warning_count: 1, topic: "audit.retention_alerts", outbox_event_id: "obe_1" }, audit_log: { id: "aud_1" } } } }), {
    status: "emitted",
    reportStatus: "warning",
    alertCount: 1,
    criticalCount: 0,
    warningCount: 1,
    topic: "audit.retention_alerts",
    outboxEventId: "obe_1",
    auditLogId: "aud_1"
  });
  assert.equal(auditRetentionAlertEmissionFromResult({ payload: { data: { emission: [] } } }), null);
  assert.deepEqual(auditArchiveRequestFromResult({ payload: { data: { archive: { archive_id: "audit_archive_1", status: "requested", topic: "audit.archive_requested", storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl", log_count: 2, integrity_failures: 0, manifest_algorithm: "sha256:v1", manifest_hash: "abc", outbox_event_id: "obe_archive_1" }, audit_log: { id: "aud_archive_1" } } } }), {
    archiveId: "audit_archive_1",
    status: "requested",
    topic: "audit.archive_requested",
    storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    logCount: 2,
    integrityFailures: 0,
    manifestAlgorithm: "sha256:v1",
    manifestHash: "abc",
    outboxEventId: "obe_archive_1",
    auditLogId: "aud_archive_1"
  });
  assert.equal(auditArchiveRequestFromResult({ payload: { data: { archive: [] } } }), null);
  assert.deepEqual(auditArchiveRecordsFromResult({ payload: { data: [{
    archive_id: "audit_archive_1",
    status: "archived",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifest_hash: "abc",
    content_hash: "content",
    bytes: 1024,
    object_lock_mode: "COMPLIANCE",
    retain_until: "2033-05-24T00:00:00Z",
    uploaded_at: "2026-05-24T00:00:01Z",
    completed_at: "2026-05-24T00:00:02Z",
    outbox_event_id: "obe_archive_1"
  }] } }), [{
    archiveId: "audit_archive_1",
    status: "archived",
    storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifestHash: "abc",
    contentHash: "content",
    bytes: 1024,
    objectLockMode: "COMPLIANCE",
    retainUntil: "2033-05-24T00:00:00Z",
    uploadedAt: "2026-05-24T00:00:01Z",
    completedAt: "2026-05-24T00:00:02Z",
    outboxEventId: "obe_archive_1"
  }]);
  assert.deepEqual(auditArchiveRecordsFromResult({ payload: { data: {} } }), []);
  assert.deepEqual(auditArchiveVerificationFromResult({ payload: { data: { verification: {
    archive_id: "audit_archive_1",
    status: "verified",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifest_hash: "abc",
    expected_content_hash: "content",
    actual_content_hash: "content",
    expected_bytes: 1024,
    actual_bytes: 1024,
    archive_id_matched: true,
    manifest_hash_matched: true,
    content_hash_matched: true,
    bytes_matched: true,
    log_count_matched: true,
    header_log_count: 1,
    manifest_entry_count: 1,
    verified_at: "2026-05-24T00:00:03Z"
  }, audit_log: { id: "aud_archive_verify_1" } } } }), {
    archiveId: "audit_archive_1",
    status: "verified",
    storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifestHash: "abc",
    expectedContentHash: "content",
    actualContentHash: "content",
    expectedBytes: 1024,
    actualBytes: 1024,
    archiveIdMatched: true,
    manifestHashMatched: true,
    contentHashMatched: true,
    bytesMatched: true,
    logCountMatched: true,
    headerLogCount: 1,
    manifestEntryCount: 1,
    errorCode: "",
    verifiedAt: "2026-05-24T00:00:03Z",
    auditLogId: "aud_archive_verify_1"
  });
  assert.equal(auditArchiveVerificationFromResult({ payload: { data: { verification: [] } } }), null);
  assert.deepEqual(auditArchiveVerificationsFromResult({ payload: { data: [{
    archive_id: "audit_archive_1",
    status: "verified",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifest_hash: "abc",
    expected_content_hash: "content",
    actual_content_hash: "content",
    expected_bytes: 1024,
    actual_bytes: 1024,
    archive_id_matched: true,
    manifest_hash_matched: true,
    content_hash_matched: true,
    bytes_matched: true,
    log_count_matched: true,
    header_log_count: 1,
    manifest_entry_count: 1,
    verified_at: "2026-05-24T00:00:03Z"
  }] } }), [{
    archiveId: "audit_archive_1",
    status: "verified",
    storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifestHash: "abc",
    expectedContentHash: "content",
    actualContentHash: "content",
    expectedBytes: 1024,
    actualBytes: 1024,
    archiveIdMatched: true,
    manifestHashMatched: true,
    contentHashMatched: true,
    bytesMatched: true,
    logCountMatched: true,
    headerLogCount: 1,
    manifestEntryCount: 1,
    errorCode: "",
    verifiedAt: "2026-05-24T00:00:03Z",
    auditLogId: ""
  }]);
  assert.deepEqual(auditArchiveVerificationsFromResult({ payload: { data: {} } }), []);

  const tamperedRows = buildAuditRows([{ ...logs[0], integrity_verified: false }]);
  assert.equal(tamperedRows[0].integrityLabel, "未通过");
  assert.equal(tamperedRows[0].integrityTone, "danger");
});

test("admin audit filters normalize ranges presets and target routes", () => {
  const normalized = normalizeAuditFilters({
    actor_type: " admin ",
    actor_id: " admin_1 ",
    action: " admin.outbox.replayed ",
    target_type: " outbox_event ",
    target_id: " obe_1 ",
    after: " 2026-05-22T00:00:00Z ",
    before: " 2026-05-22T12:00:00Z ",
    limit: 999
  });
  assert.deepEqual(normalized, {
    actor_type: "admin",
    actor_id: "admin_1",
    action: "admin.outbox.replayed",
    target_type: "outbox_event",
    target_id: "obe_1",
    after: "2026-05-22T00:00:00Z",
    before: "2026-05-22T12:00:00Z",
    limit: 500
  });
  assert.equal(auditSearchValuesFromFilters(normalized, { beforeOverride: "2026-05-22T08:00:00Z" }).before, "2026-05-22T08:00:00Z");

  const preset = makeAuditFilterPreset(normalized, "2026-05-22T13:00:00Z");
  assert.match(preset.id, /^audit_filter_/);
  assert.match(preset.name, /actor:admin/);
  assert.equal(preset.filters.limit, 500);
  assert.equal(upsertAuditFilterPreset([preset], preset).length, 1);

  const outboxRoute = auditTargetRoute({ target_type: "outbox_event", action: "admin.outbox.replayed" });
  assert.deepEqual(outboxRoute, { module: "dashboard", operation: "outbox-events", label: "Outbox 事件" });
  const rbacRoute = auditTargetRoute({ target_type: "admin_rbac_role", action: "admin.rbac.change_requested" });
  assert.deepEqual(rbacRoute, { module: "permissions", operation: "rbac-policy", label: "权限治理" });
  const rbacReviewRoute = auditTargetRoute({ target_type: "admin_rbac_change_request", action: "admin.rbac.change_reviewed" });
  assert.deepEqual(rbacReviewRoute, { module: "permissions", operation: "rbac-change-requests", label: "权限申请" });
  const fallbackRoute = auditTargetRoute({ target_type: "unknown", action: "admin.merchant_invite.created" });
  assert.equal(fallbackRoute.module, "merchants");
});

test("admin operation executor returns response metadata and payload", async () => {
  const result = await executeAdminOperation({
    baseUrl: "https://bff.local",
    operationKey: "refund-settings-save",
    token: "admin.token",
    values: { default_refund_strategy: "original_route_first" },
    fetchImpl: async (url, request) => {
      assert.equal(url, "https://bff.local/api/admin/refund-settings");
      assert.equal(request.method, "PUT");
      assert.equal(request.headers.Authorization, "Bearer admin.token");
      assert.equal(request.body, JSON.stringify({ default_refund_strategy: "original_route_first" }));
      return {
        ok: true,
        status: 200,
        text: async () => JSON.stringify({ success: true, data: { default_refund_strategy: "original_route_first" } })
      };
    }
  });
  assert.equal(result.ok, true);
  assert.equal(result.status, 200);
  assert.equal(result.payload.data.default_refund_strategy, "original_route_first");
});

test("admin snapshot adapter binds backend data into P0 views", () => {
  const snapshot = {
    generated_at: "2026-05-22T12:00:00Z",
    counts: {
      total_orders: 3,
      pending_merchant_orders: 1,
      dispatching_orders: 1,
      rider_assigned_orders: 1,
      exception_orders: 1,
      total_merchants: 2,
      merchant_qualification_risks: 1,
      merchant_deposit_missing: 1,
      total_riders: 3,
      online_riders: 2,
      rider_deposit_missing: 1,
      station_managers: 1,
      after_sales_pending: 1,
      after_sales_admin_review: 1,
      dispatch_event_count: 2,
      outbox_ready: 2,
      outbox_blocked: 1,
      object_cleanup_failed: 1,
      object_cleanup_total_candidate: 4
    },
    orders: [
      { id: "ord_1", type: "takeout", status: "dispatching", shop_id: "shop_1", rider_id: "" }
    ],
    merchants: [
      {
        account: { id: "merchant_1", display_name: "蓝湾轻食", deposit_status: "unpaid" },
        shops: [{ name: "蓝湾轻食望京店", capabilities: ["takeout", "groupbuy"] }],
        missing_qualifications: ["health_certificate"],
        qualifications: [{ expires_at: "2026-11-30T00:00:00Z" }],
        deposit: { status: "unpaid" },
        can_accept_orders: false
      }
    ],
    riders: [
      { id: "rider_1", type: "rider", station_id: "station_1", online: true, deposit_status: "paid", dispatch_priority: 40, capacity: 2 }
    ],
    rider_performance: [
      { rider_id: "rider_1", average_accept_seconds: 18.4, completion_rate: 0.98, level: "S", dispatch_priority: 40 }
    ],
    after_sales: [
      { id: "asr_1", order_id: "ord_1", user_id: "user_1", status: "admin_review", refundable_fen: 1200, evidence_urls: ["https://cdn.test/evidence.jpg"] }
    ],
    dispatch_events: [
      { order_id: "ord_1", type: "dispatch.assigned", mode: "auto_assign", rider_id: "rider_1", online_candidate_size: 2, created_at: "2026-05-22T12:01:00Z" }
    ],
    refund_settings: { default_refund_strategy: "balance_first" },
    outbox_stats: { ready: 2, blocked: 1, dead_letter: 0 }
  };

  const orderView = applySnapshotToAdminView(getAdminView("orders"), snapshot);
  assert.equal(orderView.metrics[1].value, "1");
  assert.equal(orderView.rows[0][0], "ord_1");
  assert.equal(orderView.rows[0][2], "待派单");

  const merchantView = applySnapshotToAdminView(getAdminView("merchants"), snapshot);
  assert.equal(merchantView.rows[0][0], "蓝湾轻食");
  assert.equal(merchantView.rows[0][3], "未缴");

  const performanceView = applySnapshotToAdminView(getAdminView("rider-performance"), snapshot);
  assert.equal(performanceView.rows[0][1], "18s");
  assert.equal(performanceView.rows[0][2], "98.0%");

  const kpis = buildSnapshotKpis(snapshot, ADMIN_WEB_KPIS);
  assert.equal(kpis[0].value, "3");
  assert.equal(kpis[3].value, "2");

  const queues = buildSnapshotQueues(snapshot, ADMIN_WEB_QUEUES);
  assert.equal(queues[1].title, "商户资质/保证金");
  assert.equal(queues[4].target, "Ready 2 / Blocked 1");

  assert.equal(snapshotDataFromResult({ payload: { data: snapshot } }), snapshot);
});
