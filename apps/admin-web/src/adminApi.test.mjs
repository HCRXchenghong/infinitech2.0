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
  buildAuditArchiveVerificationRows,
  buildAuditRows,
  makeAuditFilterPreset,
  nextAuditBefore,
  normalizeAuditFilters,
  redactAuditPayload,
  summarizeAuditPayload,
  upsertAuditFilterPreset
} from "./adminAudit.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS, getAdminWebModule } from "./config.mjs";
import { buildAdminBusinessDetail } from "./adminDetails.mjs";
import {
  buildOperationHistoryEntry,
  buildPendingOperation,
  canReplayOperationHistoryEntry,
  operationHistoryReplayValues,
  operationRiskProfile,
  operationValuesSnapshot,
  trimOperationHistory
} from "./adminOperations.mjs";
import { VIEW_PAGE_SIZE_OPTIONS, buildFilteredRows, buildViewPage, normalizeViewFilter } from "./adminTable.mjs";
import { ADMIN_WEB_VIEWS, getAdminView } from "./adminViews.mjs";
import { applySnapshotToAdminView, buildSnapshotKpis, buildSnapshotQueues, snapshotDataFromResult } from "./adminSnapshot.mjs";

test("admin web exposes the first operable control-center modules", () => {
  for (const key of ["orders", "after-sales", "merchants", "riders", "dispatch", "audit-logs", "refund-settings", "payment", "support", "notifications", "rtc", "integrations", "permissions"]) {
    assert.ok(getAdminWebModule(key), `missing ${key}`);
  }
  assert.ok(ADMIN_WEB_SECTIONS.length >= 4);
  assert.ok(ADMIN_WEB_MODULES.length >= 25);
  assert.ok(ADMIN_WEB_KPIS.some((item) => item.key === "outbox"));
  assert.ok(ADMIN_WEB_QUEUES.some((item) => item.operationKey === "object-cleanup-stats"));
  assert.ok(ADMIN_WEB_QUEUES.some((item) => item.operationKey === "notification-deliveries"));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "ops_admin" && item.scopes.includes("merchant:qualification_review")));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "ops_admin" && item.scopes.includes("notification:write")));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "support_admin" && item.scopes.includes("notification:read")));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "finance_admin" && item.scopes.includes("refund:write")));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "security_auditor" && item.scopes.includes("audit:read")));
  assert.ok(ADMIN_WEB_RBAC.every((item) => item.scopes.includes("*") || item.scopes.includes("rbac:read")));
});

test("admin web operation catalog covers shipped admin API surfaces", () => {
  for (const key of [
    "admin-login",
    "merchant-invite",
    "merchant-qualifications",
    "merchant-qualification-detail",
    "merchant-qualification-review",
    "station-manager-invite",
    "rider-invite",
    "refund-settings-read",
    "refund-settings-save",
    "order-refund",
    "order-detail",
    "refund-transactions",
    "after-sales-list",
    "after-sales-detail",
    "after-sales-events",
    "after-sales-evidence",
    "after-sales-review",
    "support-tickets",
    "support-ticket-detail",
    "support-ticket-assign",
    "support-ticket-escalate",
    "support-ticket-resolve",
    "support-quality-review",
    "support-quality-reviews",
    "support-performance",
    "operations-snapshot",
    "audit-logs",
    "audit-logs-export",
    "audit-retention-report",
    "audit-retention-alert-emit",
    "audit-archive-request",
    "audit-archive-records",
    "audit-archive-verify",
    "audit-archive-verifications",
    "notifications",
    "notification-deliveries",
    "notification-preferences",
    "notification-preference-save",
    "notification-preference-batch-save",
    "notification-preference-change-requests",
    "notification-preference-change-request",
    "notification-preference-change-review",
    "notification-preference-change-apply",
    "notification-delivery-record",
    "notification-failure-alert-emit",
    "notification-delivery-retry-schedule",
    "notification-quiet-window-retry-schedule",
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
    "outbox-event-detail",
    "outbox-dead-letter-triage",
    "outbox-replay-batch",
    "outbox-claim-events",
    "outbox-renew-lease",
    "outbox-release-dead-letter",
    "outbox-replay-event",
    "outbox-mark-failed",
    "outbox-mark-published",
    "order-compensate",
    "dispatch-order-events",
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
  for (const key of ["orders", "after-sales", "merchants", "riders", "rider-performance", "dispatch", "audit-logs", "refund-settings", "support", "notifications", "permissions"]) {
    const view = getAdminView(key);
    assert.equal(view.key, key);
    assert.ok(view.metrics.length >= 4, `missing metrics for ${key}`);
    assert.ok(view.columns.length >= 4, `missing columns for ${key}`);
    assert.ok(view.rows.length >= 4, `missing rows for ${key}`);
    assert.ok(view.safeguards.length >= 3, `missing safeguards for ${key}`);
  }
  assert.ok(ADMIN_WEB_VIEWS.orders.actions.includes("order-compensate"));
  assert.ok(ADMIN_WEB_VIEWS.orders.actions.includes("order-refund"));
  assert.ok(ADMIN_WEB_VIEWS.orders.actions.includes("order-detail"));
  assert.ok(ADMIN_WEB_VIEWS.orders.actions.includes("refund-transactions"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-claim-events"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-event-detail"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-renew-lease"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-dead-letter-triage"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-release-dead-letter"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-replay-event"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-mark-failed"));
  assert.ok(ADMIN_WEB_VIEWS.dashboard.actions.includes("outbox-mark-published"));
  assert.ok(ADMIN_WEB_VIEWS["after-sales"].actions.includes("after-sales-review"));
  assert.ok(ADMIN_WEB_VIEWS["after-sales"].actions.includes("after-sales-detail"));
  assert.ok(ADMIN_WEB_VIEWS["after-sales"].actions.includes("refund-transactions"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-tickets"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-ticket-detail"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-ticket-assign"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-ticket-escalate"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-ticket-resolve"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-quality-review"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-quality-reviews"));
  assert.ok(ADMIN_WEB_VIEWS.support.actions.includes("support-performance"));
  assert.ok(ADMIN_WEB_VIEWS.merchants.actions.includes("merchant-invite"));
  assert.ok(ADMIN_WEB_VIEWS.merchants.actions.includes("merchant-qualifications"));
  assert.ok(ADMIN_WEB_VIEWS.merchants.actions.includes("merchant-qualification-detail"));
  assert.ok(ADMIN_WEB_VIEWS.merchants.actions.includes("merchant-qualification-review"));
  assert.ok(ADMIN_WEB_VIEWS.riders.actions.includes("station-riders"));
  assert.ok(ADMIN_WEB_VIEWS.dispatch.actions.includes("station-orders"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-logs-export"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-retention-report"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-retention-alert-emit"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-request"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-records"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-verify"));
  assert.ok(ADMIN_WEB_VIEWS["audit-logs"].actions.includes("audit-archive-verifications"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notifications"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-deliveries"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preferences"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-save"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-batch-save"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-change-requests"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-change-request"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-change-review"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-preference-change-apply"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-delivery-record"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-failure-alert-emit"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-delivery-retry-schedule"));
  assert.ok(ADMIN_WEB_VIEWS.notifications.actions.includes("notification-quiet-window-retry-schedule"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-change-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-review-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-apply-request"));
  assert.ok(ADMIN_WEB_VIEWS.permissions.actions.includes("rbac-rollback-request"));
});

test("admin web builds P0 business detail drawers with audited next actions", () => {
  const orderDetail = buildAdminBusinessDetail(getAdminView("orders"), 0);
  assert.equal(orderDetail.title, "订单 ord_10031");
  assert.deepEqual(orderDetail.facts.map((fact) => fact.label), ["订单", "类型", "状态", "商户", "骑手", "风险"]);
  assert.ok(orderDetail.actions.some((action) =>
    action.operationKey === "order-detail" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(orderDetail.actions.some((action) =>
    action.operationKey === "order-refund" &&
    action.values.order_id === "ord_10031" &&
    action.values.idempotency_key === "refund_ord_10031"
  ));
  assert.ok(orderDetail.actions.some((action) =>
    action.operationKey === "refund-transactions" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(orderDetail.actions.some((action) => action.operationKey === "order-compensate" && action.values.order_id === "ord_10031"));
  assert.ok(orderDetail.actions.some((action) => action.operationKey === "audit-logs" && action.values.target_type === "order"));
  assert.ok(orderDetail.checklist.some((item) => item.includes("支付流水")));

  const dashboardDetail = buildAdminBusinessDetail(getAdminView("dashboard"), 3);
  assert.equal(dashboardDetail.title, "队列 Outbox");
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "merchant-qualifications" && action.values.status === "pending_review"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "merchant-qualification-detail" && action.values.audit_limit === 20));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-dead-letter-triage" && action.values.status === "dead_letter"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-event-detail" && action.values.event_id === "obe_1"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-claim-events" && action.values.lease_owner === "relay-admin"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-renew-lease" && action.values.event_id === "obe_1"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-release-dead-letter" && action.values.event_id === "obe_dead_1"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-replay-event" && action.values.event_id === "obe_1"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-mark-failed" && action.values.error === "relay down"));
  assert.ok(dashboardDetail.actions.some((action) => action.operationKey === "outbox-mark-published" && action.values.event_id === "obe_1"));
  assert.ok(dashboardDetail.checklist.some((item) => item.includes("dead letter")));

  const afterSalesDetail = buildAdminBusinessDetail(getAdminView("after-sales"), 0);
  assert.equal(afterSalesDetail.title, "售后 asr_231");
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "order-detail" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "after-sales-review" &&
    action.values.request_id === "asr_231" &&
    action.values.refund_idempotency_key === "after_sales:asr_231"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "after-sales-detail" &&
    action.values.request_id === "asr_231"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "after-sales-list" &&
    action.values.request_id === "asr_231" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "refund-transactions" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "support-tickets" &&
    action.values.related_order_id === "ord_10031"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "after-sales-events" &&
    action.values.request_id === "asr_231"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "after-sales-evidence" &&
    action.values.request_id === "asr_231"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "dispatch-order-events" &&
    action.values.order_id === "ord_10031"
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "audit-logs" &&
    action.values.target_type === "order" &&
    action.values.target_id === "ord_10031" &&
    !("action" in action.values)
  ));
  assert.ok(afterSalesDetail.actions.some((action) =>
    action.operationKey === "audit-logs" &&
    action.values.target_type === "order" &&
    action.values.target_id === "ord_10031" &&
    action.values.action === "admin.order.refunded"
  ));
  assert.ok(afterSalesDetail.actions.some((action) => action.operationKey === "object-cleanup-candidates"));
  assert.ok(afterSalesDetail.checklist.some((item) => item.includes("扫描门禁")));

  const supportDetail = buildAdminBusinessDetail(getAdminView("support"), 0);
  assert.equal(supportDetail.title, "客服工单 st_1");
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-tickets" &&
    action.values.status === "processing"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-ticket-detail" &&
    action.values.ticket_id === "st_1"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-ticket-assign" &&
    action.values.ticket_id === "st_1" &&
    action.values.support_name === "客服小悦"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-ticket-escalate" &&
    action.values.ticket_id === "st_1" &&
    action.values.escalation_level === "support_lead"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-ticket-resolve" &&
    action.values.ticket_id === "st_1" &&
    action.values.solution.includes("补偿方案")
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-quality-review" &&
    action.values.ticket_id === "st_1" &&
    action.values.result === "needs_coaching"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-quality-reviews" &&
    action.values.ticket_id === "st_1"
  ));
  assert.ok(supportDetail.actions.some((action) =>
    action.operationKey === "support-performance" &&
    action.values.limit === 20
  ));
  assert.ok(supportDetail.checklist.some((item) => item.includes("SLA")));

  const riderPerformanceDetail = buildAdminBusinessDetail(getAdminView("rider-performance"), 0);
  assert.equal(riderPerformanceDetail.title, "绩效 rider_71");
  assert.deepEqual(riderPerformanceDetail.facts.map((fact) => fact.label), ["骑手", "平均接单", "完成率", "配送评分", "派单分", "评分拆解", "等级", "派单优先级"]);
  assert.ok(riderPerformanceDetail.checklist.some((item) => item.includes("派单分拆解")));

  const merchantDetail = buildAdminBusinessDetail(getAdminView("merchants"), 1);
  assert.equal(merchantDetail.title, "商户 merchant_19");
  assert.ok(merchantDetail.actions.some((action) =>
    action.operationKey === "merchant-qualifications" &&
    action.values.merchant_id === "merchant_19" &&
    action.values.status === "pending_review"
  ));
  assert.ok(merchantDetail.actions.some((action) =>
    action.operationKey === "merchant-qualification-detail" &&
    action.values.qualification_id === "mq_merchant_19_health_certificate"
  ));
  assert.ok(merchantDetail.actions.some((action) =>
    action.operationKey === "merchant-qualification-review" &&
    action.values.merchant_id === "merchant_19" &&
    action.values.qualification_id === "mq_merchant_19_health_certificate"
  ));
  assert.ok(merchantDetail.actions.some((action) => action.operationKey === "audit-logs" && action.values.target_type === "merchant_account"));

  const permissionDetail = buildAdminBusinessDetail(getAdminView("permissions"), 0);
  assert.equal(permissionDetail.title, "权限 super_admin");
  assert.ok(permissionDetail.actions.some((action) => action.operationKey === "rbac-policy"));
  assert.ok(permissionDetail.checklist.some((item) => item.includes("两人参与")));

  const notificationDetail = buildAdminBusinessDetail(getAdminView("notifications"), 1);
  assert.equal(notificationDetail.title, "通知 ntf_2");
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notifications" &&
    action.values.target_id === "merchant_19" &&
    action.values.source_topic === "merchant.qualification_reviewed"
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-deliveries" &&
    action.values.notification_id === "ntf_2" &&
    action.values.status === "failed"
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-delivery-record" &&
    action.values.idempotency_key === "delivery:manual:ntf_2:wechat_subscribe" &&
    action.values.error_code === "invalid_openid"
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-failure-alert-emit" &&
    action.values.channel === "wechat_subscribe" &&
    action.values.target_id === "merchant_19"
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-delivery-retry-schedule" &&
    action.values.channel === "wechat_subscribe" &&
    action.values.retry_after_seconds === 300
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-preferences" &&
    action.values.target_id === "merchant_19" &&
    action.values.notification_type === "merchant.qualification_reviewed"
  ));
  assert.ok(notificationDetail.actions.some((action) =>
    action.operationKey === "notification-preference-save" &&
    action.values.target_role === "merchant" &&
    action.values.target_id === "merchant_19" &&
    action.values.notification_type === "merchant.qualification_reviewed" &&
    action.values.disabled_channels === "wechat_subscribe" &&
    action.values.quiet_hours.includes("timezone_offset")
  ));
  assert.ok(notificationDetail.actions.some((action) => action.operationKey === "outbox-events"));
  assert.ok(notificationDetail.checklist.some((item) => item.includes("退避窗口")));

  assert.equal(buildAdminBusinessDetail({ key: "orders", columns: ["订单"], rows: [["暂无快照数据"]] }, 0), null);
  assert.equal(buildAdminBusinessDetail(getAdminView("orders"), 999), null);
});

test("admin web requires confirmation and tracks high-risk operation results", () => {
  const refundOperation = getAdminOperation("refund-settings-save");
  const merchantQualificationOperation = getAdminOperation("merchant-qualification-review");
  const orderRefundOperation = getAdminOperation("order-refund");
  const afterSalesOperation = getAdminOperation("after-sales-review");
  const supportAssignOperation = getAdminOperation("support-ticket-assign");
  const supportEscalateOperation = getAdminOperation("support-ticket-escalate");
  const supportResolveOperation = getAdminOperation("support-ticket-resolve");
  const supportQualityOperation = getAdminOperation("support-quality-review");
  const outboxClaimOperation = getAdminOperation("outbox-claim-events");
  const outboxRenewOperation = getAdminOperation("outbox-renew-lease");
  const outboxDeadLetterTriageOperation = getAdminOperation("outbox-dead-letter-triage");
  const outboxDeadLetterReleaseOperation = getAdminOperation("outbox-release-dead-letter");
  const outboxReplayOperation = getAdminOperation("outbox-replay-event");
  const outboxFailedOperation = getAdminOperation("outbox-mark-failed");
  const outboxPublishedOperation = getAdminOperation("outbox-mark-published");
  const notificationDeliveryRecordOperation = getAdminOperation("notification-delivery-record");
  const notificationFailureAlertOperation = getAdminOperation("notification-failure-alert-emit");
  const notificationRetryScheduleOperation = getAdminOperation("notification-delivery-retry-schedule");
  const notificationQuietRetryScheduleOperation = getAdminOperation("notification-quiet-window-retry-schedule");
  const notificationPreferenceSaveOperation = getAdminOperation("notification-preference-save");
  const readOperation = getAdminOperation("refund-settings-read");
  assert.equal(operationRiskProfile(refundOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(refundOperation).reason, /退款默认策略/);
  assert.equal(operationRiskProfile(merchantQualificationOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(merchantQualificationOperation).reason, /商户资质|接单资格/);
  assert.equal(operationRiskProfile(orderRefundOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(orderRefundOperation).reason, /订单退款/);
  assert.equal(operationRiskProfile(afterSalesOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(afterSalesOperation).reason, /审核售后/);
  assert.equal(operationRiskProfile(supportAssignOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(supportAssignOperation).reason, /客服工单|SLA/);
  assert.equal(operationRiskProfile(supportEscalateOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(supportEscalateOperation).reason, /升级客服工单|SLA|主管/);
  assert.equal(operationRiskProfile(supportResolveOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(supportResolveOperation).reason, /处理方案|用户确认|工单状态/);
  assert.equal(operationRiskProfile(supportQualityOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(supportQualityOperation).reason, /质检|绩效|辅导/);
  assert.equal(operationRiskProfile(outboxClaimOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxClaimOperation).reason, /租约/);
  assert.equal(operationRiskProfile(outboxRenewOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxRenewOperation).reason, /续租/);
  assert.equal(operationRiskProfile(outboxDeadLetterTriageOperation).requiresConfirmation, false);
  assert.equal(operationRiskProfile(outboxDeadLetterReleaseOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxDeadLetterReleaseOperation).reason, /dead-letter|解封/);
  assert.equal(operationRiskProfile(outboxReplayOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxReplayOperation).reason, /单个 outbox/);
  assert.equal(operationRiskProfile(outboxFailedOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxFailedOperation).reason, /backoff|dead-letter/);
  assert.equal(operationRiskProfile(outboxPublishedOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(outboxPublishedOperation).reason, /已发布/);
  assert.equal(operationRiskProfile(notificationDeliveryRecordOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(notificationDeliveryRecordOperation).reason, /投递回执|送达/);
  assert.equal(operationRiskProfile(notificationFailureAlertOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(notificationFailureAlertOperation).reason, /失败告警|outbox/);
  assert.equal(operationRiskProfile(notificationRetryScheduleOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(notificationRetryScheduleOperation).reason, /重试|退避/);
  assert.equal(operationRiskProfile(notificationQuietRetryScheduleOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(notificationQuietRetryScheduleOperation).reason, /静默|queued/);
  assert.equal(operationRiskProfile(notificationPreferenceSaveOperation).requiresConfirmation, true);
  assert.match(operationRiskProfile(notificationPreferenceSaveOperation).reason, /通知偏好|静默|渠道/);
  assert.equal(operationRiskProfile(readOperation).requiresConfirmation, false);

  const snapshot = operationValuesSnapshot(refundOperation, { default_refund_strategy: "original_route_first" });
  assert.deepEqual(snapshot, { default_refund_strategy: "original_route_first" });

  const pending = buildPendingOperation(refundOperation, snapshot, "2026-05-24T12:00:00Z");
  assert.equal(pending.operationKey, "refund-settings-save");
  assert.equal(pending.title, "保存退款策略");
  assert.equal(pending.values.default_refund_strategy, "original_route_first");

  const successEntry = buildOperationHistoryEntry({
    ok: true,
    status: 200,
    operation: refundOperation,
    request: { method: "PUT", url: "/api/admin/refund-settings" },
    payload: { success: true }
  }, snapshot, "2026-05-24T12:00:01Z");
  assert.equal(successEntry.ok, true);
  assert.equal(successEntry.title, "保存退款策略");
  assert.equal(successEntry.url, "/api/admin/refund-settings");

  const failedEntry = buildOperationHistoryEntry({
    ok: false,
    status: 403,
    operation: refundOperation,
    request: { method: "PUT", url: "/api/admin/refund-settings" },
    payload: { success: false, message: "forbidden" }
  }, snapshot, "2026-05-24T12:00:02Z");
  assert.equal(failedEntry.ok, false);
  assert.equal(failedEntry.message, "forbidden");
  assert.equal(trimOperationHistory([successEntry, failedEntry], 1).length, 1);
  assert.equal(canReplayOperationHistoryEntry(successEntry), false);
  assert.equal(canReplayOperationHistoryEntry(failedEntry), true);
  assert.deepEqual(operationHistoryReplayValues(failedEntry), snapshot);
});

test("admin web filters and paginates P0 business view rows", () => {
  const view = getAdminView("orders");
  assert.deepEqual(VIEW_PAGE_SIZE_OPTIONS, [4, 10, 20]);
  assert.deepEqual(normalizeViewFilter({ query: "  派单 ", page: "2", pageSize: "10" }), {
    query: "派单",
    page: 2,
    pageSize: 10
  });
  assert.equal(buildFilteredRows(view, { query: "康宁" }).length, 1);
  assert.equal(buildFilteredRows(view, { query: "ORD_10031" })[0].rowIndex, 0);

  const firstPage = buildViewPage(view, { pageSize: 4 });
  assert.equal(firstPage.totalRows, 4);
  assert.equal(firstPage.totalPages, 1);
  assert.equal(firstPage.rows[0].row[0], "ord_10031");

  const emptyPage = buildViewPage(view, { query: "不存在", page: 99, pageSize: 4 });
  assert.equal(emptyPage.page, 1);
  assert.equal(emptyPage.totalRows, 0);
  assert.deepEqual(emptyPage.rows, []);
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

  const orderRefund = buildAdminRequest(getAdminOperation("order-refund"), {
    order_id: "ord 10031",
    reason: "客服确认退款",
    idempotency_key: "refund_ord_10031",
    amount_fen: "1200",
    destination: "balance"
  }, "admin.token");
  assert.equal(orderRefund.method, "POST");
  assert.equal(orderRefund.url, "/api/orders/ord%2010031/refund");
  assert.deepEqual(JSON.parse(orderRefund.body), {
    reason: "客服确认退款",
    idempotency_key: "refund_ord_10031",
    amount_fen: 1200,
    destination: "balance"
  });

  const orderDetailRequest = buildAdminRequest(getAdminOperation("order-detail"), {
    order_id: "ord 10031"
  }, "admin.token");
  assert.equal(orderDetailRequest.method, "GET");
  assert.equal(orderDetailRequest.url, "/api/admin/orders/ord%2010031");

  const refundTransactions = buildAdminRequest(getAdminOperation("refund-transactions"), {
    order_id: "ord 10031",
    user_id: "user 1",
    destination: "balance",
    status: "success",
    limit: "5"
  }, "admin.token");
  assert.equal(refundTransactions.method, "GET");
  assert.equal(refundTransactions.url, "/api/admin/refunds?order_id=ord+10031&user_id=user+1&destination=balance&status=success&limit=5");

  const merchantQualifications = buildAdminRequest(getAdminOperation("merchant-qualifications"), {
    status: "pending_review",
    merchant_id: "merchant 19",
    type: "health_certificate",
    limit: "5",
    now: "2026-05-25T12:00:00Z"
  }, "admin.token");
  assert.equal(merchantQualifications.method, "GET");
  assert.equal(merchantQualifications.url, "/api/admin/merchant-qualifications?status=pending_review&merchant_id=merchant+19&type=health_certificate&limit=5&now=2026-05-25T12%3A00%3A00Z");

  const merchantQualificationDetail = buildAdminRequest(getAdminOperation("merchant-qualification-detail"), {
    qualification_id: "mq merchant health",
    audit_limit: "5",
    now: "2026-05-25T12:00:00Z"
  }, "admin.token");
  assert.equal(merchantQualificationDetail.method, "GET");
  assert.equal(merchantQualificationDetail.url, "/api/admin/merchant-qualifications/mq%20merchant%20health?audit_limit=5&now=2026-05-25T12%3A00%3A00Z");

  const merchantQualificationReview = buildAdminRequest(getAdminOperation("merchant-qualification-review"), {
    qualification_id: "mq merchant health",
    merchant_id: "merchant_19",
    decision: "reject",
    reason: "健康证照片模糊",
    reviewed_at: "2026-05-25T12:00:00Z"
  }, "admin.token");
  assert.equal(merchantQualificationReview.method, "POST");
  assert.equal(merchantQualificationReview.url, "/api/admin/merchant-qualifications/mq%20merchant%20health/review");
  assert.deepEqual(JSON.parse(merchantQualificationReview.body), {
    merchant_id: "merchant_19",
    decision: "reject",
    reason: "健康证照片模糊",
    reviewed_at: "2026-05-25T12:00:00Z"
  });

  const outboxReplay = buildAdminRequest(getAdminOperation("outbox-replay-event"), {
    event_id: "obe 1",
    now: "2026-05-22T12:01:30Z"
  }, "admin.token");
  assert.equal(outboxReplay.method, "POST");
  assert.equal(outboxReplay.url, "/api/admin/outbox/events/obe%201/replay");
  assert.deepEqual(JSON.parse(outboxReplay.body), { now: "2026-05-22T12:01:30Z" });

  const outboxDetail = buildAdminRequest(getAdminOperation("outbox-event-detail"), {
    event_id: "obe 1",
    now: "2026-05-22T12:01:30Z",
    audit_limit: "5"
  }, "admin.token");
  assert.equal(outboxDetail.method, "GET");
  assert.equal(outboxDetail.url, "/api/admin/outbox/events/obe%201?now=2026-05-22T12%3A01%3A30Z&audit_limit=5");

  const outboxDeadLetters = buildAdminRequest(getAdminOperation("outbox-dead-letter-triage"), {
    topic: "order.paid",
    limit: "5",
    now: "2026-05-22T12:05:00Z"
  }, "admin.token");
  assert.equal(outboxDeadLetters.method, "GET");
  assert.equal(outboxDeadLetters.url, "/api/admin/outbox/events?topic=order.paid&status=dead_letter&limit=5&now=2026-05-22T12%3A05%3A00Z");

  const outboxDeadLetterRelease = buildAdminRequest(getAdminOperation("outbox-release-dead-letter"), {
    event_id: "obe dead 1",
    now: "2026-05-22T12:06:00Z"
  }, "admin.token");
  assert.equal(outboxDeadLetterRelease.method, "POST");
  assert.equal(outboxDeadLetterRelease.url, "/api/admin/outbox/events/obe%20dead%201/replay");
  assert.deepEqual(JSON.parse(outboxDeadLetterRelease.body), { now: "2026-05-22T12:06:00Z" });

  const outboxClaim = buildAdminRequest(getAdminOperation("outbox-claim-events"), {
    topic: "order.paid",
    limit: "2",
    lease_owner: "relay-admin",
    lease_seconds: "90",
    now: "2026-05-22T12:00:00Z"
  }, "admin.token");
  assert.equal(outboxClaim.method, "POST");
  assert.equal(outboxClaim.url, "/api/admin/outbox/events/claim");
  assert.deepEqual(JSON.parse(outboxClaim.body), {
    topic: "order.paid",
    limit: 2,
    lease_owner: "relay-admin",
    lease_seconds: 90,
    now: "2026-05-22T12:00:00Z"
  });

  const outboxRenew = buildAdminRequest(getAdminOperation("outbox-renew-lease"), {
    event_id: "obe 1",
    lease_owner: "relay-admin",
    lease_seconds: "90",
    now: "2026-05-22T12:00:30Z"
  }, "admin.token");
  assert.equal(outboxRenew.method, "POST");
  assert.equal(outboxRenew.url, "/api/admin/outbox/events/obe%201/lease/renew");
  assert.deepEqual(JSON.parse(outboxRenew.body), {
    lease_owner: "relay-admin",
    lease_seconds: 90,
    now: "2026-05-22T12:00:30Z"
  });

  const outboxFailed = buildAdminRequest(getAdminOperation("outbox-mark-failed"), {
    event_id: "obe 1",
    error: "relay down",
    retry_after_seconds: "120",
    max_attempts: "10",
    now: "2026-05-22T12:03:00Z"
  }, "admin.token");
  assert.equal(outboxFailed.method, "POST");
  assert.equal(outboxFailed.url, "/api/admin/outbox/events/obe%201/failed");
  assert.deepEqual(JSON.parse(outboxFailed.body), {
    error: "relay down",
    retry_after_seconds: 120,
    max_attempts: 10,
    now: "2026-05-22T12:03:00Z"
  });

  const outboxPublished = buildAdminRequest(getAdminOperation("outbox-mark-published"), {
    event_id: "obe 1",
    published_at: "2026-05-22T12:04:00Z"
  }, "admin.token");
  assert.equal(outboxPublished.method, "POST");
  assert.equal(outboxPublished.url, "/api/admin/outbox/events/obe%201/published");
  assert.deepEqual(JSON.parse(outboxPublished.body), { published_at: "2026-05-22T12:04:00Z" });

  const notifications = buildAdminRequest(getAdminOperation("notifications"), {
    target_role: "merchant",
    target_id: "merchant 19",
    status: "unread",
    source_topic: "merchant.qualification_reviewed",
    source_event_id: "obe mq 1",
    limit: "5"
  }, "admin.token");
  assert.equal(notifications.method, "GET");
  assert.equal(notifications.url, "/api/admin/notifications?target_role=merchant&target_id=merchant+19&status=unread&source_topic=merchant.qualification_reviewed&source_event_id=obe+mq+1&limit=5");

  const notificationDeliveries = buildAdminRequest(getAdminOperation("notification-deliveries"), {
    notification_id: "ntf 1",
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    status: "failed",
    limit: "5"
  }, "admin.token");
  assert.equal(notificationDeliveries.method, "GET");
  assert.equal(notificationDeliveries.url, "/api/admin/notification-deliveries?notification_id=ntf+1&target_role=merchant&target_id=merchant+19&channel=wechat_subscribe&provider=wechat_subscribe&status=failed&limit=5");

  const notificationPreferences = buildAdminRequest(getAdminOperation("notification-preferences"), {
    preference_key: "merchant:merchant 19:merchant.qualification_reviewed",
    target_role: "merchant",
    target_id: "merchant 19",
    notification_type: "merchant.qualification_reviewed",
    limit: "5"
  }, "admin.token");
  assert.equal(notificationPreferences.method, "GET");
  assert.equal(notificationPreferences.url, "/api/admin/notification-preferences?preference_key=merchant%3Amerchant+19%3Amerchant.qualification_reviewed&target_role=merchant&target_id=merchant+19&notification_type=merchant.qualification_reviewed&limit=5");

  const notificationPreferenceSave = buildAdminRequest(getAdminOperation("notification-preference-save"), {
    target_role: "merchant",
    target_id: "merchant 19",
    notification_type: "merchant.qualification_reviewed",
    enabled_channels: "wechat_subscribe,push",
    disabled_channels: "sms",
    quiet_hours: "{\"enabled\":true,\"start\":\"22:00\",\"end\":\"08:00\",\"timezone_offset\":\"+08:00\",\"channels\":[\"wechat_subscribe\",\"push\"]}",
    updated_at: "2026-05-25T12:15:00Z"
  }, "admin.token");
  assert.equal(notificationPreferenceSave.method, "PUT");
  assert.equal(notificationPreferenceSave.url, "/api/admin/notification-preferences");
  assert.deepEqual(JSON.parse(notificationPreferenceSave.body), {
    target_role: "merchant",
    target_id: "merchant 19",
    notification_type: "merchant.qualification_reviewed",
    enabled_channels: ["wechat_subscribe", "push"],
    disabled_channels: ["sms"],
    quiet_hours: {
      enabled: true,
      start: "22:00",
      end: "08:00",
      timezone_offset: "+08:00",
      channels: ["wechat_subscribe", "push"]
    },
    updated_at: "2026-05-25T12:15:00Z"
  });
  const notificationPreferenceBatchSave = buildAdminRequest(getAdminOperation("notification-preference-batch-save"), {
    preferences: "[{\"target_role\":\"merchant\",\"target_id\":\"merchant 19\",\"notification_type\":\"merchant.qualification_reviewed\",\"disabled_channels\":[\"sms\"]},{\"target_role\":\"merchant\",\"target_id\":\"merchant 19\",\"notification_type\":\"order.status_changed\",\"disabled_channels\":[\"push\"]}]",
    reason: "bulk rollout",
    updated_at: "2026-05-25T12:16:00Z"
  }, "admin.token");
  assert.equal(notificationPreferenceBatchSave.method, "POST");
  assert.equal(notificationPreferenceBatchSave.url, "/api/admin/notification-preferences/batch");
  assert.deepEqual(JSON.parse(notificationPreferenceBatchSave.body), {
    preferences: [
      {
        target_role: "merchant",
        target_id: "merchant 19",
        notification_type: "merchant.qualification_reviewed",
        disabled_channels: ["sms"]
      },
      {
        target_role: "merchant",
        target_id: "merchant 19",
        notification_type: "order.status_changed",
        disabled_channels: ["push"]
      }
    ],
    reason: "bulk rollout",
    updated_at: "2026-05-25T12:16:00Z"
  });
  const notificationPreferenceChangeRequests = buildAdminRequest(getAdminOperation("notification-preference-change-requests"), {
    status: "pending_approval",
    limit: "5"
  }, "admin.token");
  assert.equal(notificationPreferenceChangeRequests.method, "GET");
  assert.equal(notificationPreferenceChangeRequests.url, "/api/admin/notification-preferences/change-requests?status=pending_approval&limit=5");

  const notificationPreferenceChangeRequest = buildAdminRequest(getAdminOperation("notification-preference-change-request"), {
    preferences: "[{\"target_role\":\"merchant\",\"target_id\":\"merchant 19\",\"notification_type\":\"order.status_changed\",\"disabled_channels\":[\"sms\"]}]",
    rollout: "{\"mode\":\"target_ids\",\"target_ids\":[\"merchant 19\"],\"max_targets\":5}",
    reason: "approval needed",
    updated_at: "2026-05-25T12:17:00Z"
  }, "admin.token");
  assert.equal(notificationPreferenceChangeRequest.method, "POST");
  assert.equal(notificationPreferenceChangeRequest.url, "/api/admin/notification-preferences/change-requests");
  assert.deepEqual(JSON.parse(notificationPreferenceChangeRequest.body), {
    preferences: [
      {
        target_role: "merchant",
        target_id: "merchant 19",
        notification_type: "order.status_changed",
        disabled_channels: ["sms"]
      }
    ],
    rollout: {
      mode: "target_ids",
      target_ids: ["merchant 19"],
      max_targets: 5
    },
    reason: "approval needed",
    updated_at: "2026-05-25T12:17:00Z"
  });

  const notificationPreferenceChangeReview = buildAdminRequest(getAdminOperation("notification-preference-change-review"), {
    change_request_id: "ntfp change 1",
    decision: "approve",
    reason: "reviewed"
  }, "admin.token");
  assert.equal(notificationPreferenceChangeReview.method, "POST");
  assert.equal(notificationPreferenceChangeReview.url, "/api/admin/notification-preferences/change-requests/ntfp%20change%201/review");

  const notificationPreferenceChangeApply = buildAdminRequest(getAdminOperation("notification-preference-change-apply"), {
    change_request_id: "ntfp change 1",
    reason: "apply approved policy",
    updated_at: "2026-05-25T12:18:00Z"
  }, "admin.token");
  assert.equal(notificationPreferenceChangeApply.method, "POST");
  assert.equal(notificationPreferenceChangeApply.url, "/api/admin/notification-preferences/change-requests/ntfp%20change%201/apply");
  assert.deepEqual(JSON.parse(notificationPreferenceChangeApply.body), {
    reason: "apply approved policy",
    updated_at: "2026-05-25T12:18:00Z"
  });
  assert.throws(() => buildAdminRequest(getAdminOperation("notification-preference-save"), {
    target_role: "merchant",
    target_id: "merchant 19",
    notification_type: "merchant.qualification_reviewed",
    quiet_hours: "{bad"
  }, "admin.token"), /有效 JSON/);

  const notificationDeliveryRecord = buildAdminRequest(getAdminOperation("notification-delivery-record"), {
    notification_id: "ntf 1",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    status: "failed",
    provider_message_id: "provider msg 1",
    error_code: "invalid_openid",
    error_message: "openid missing",
    idempotency_key: "delivery:manual:ntf_1:wechat",
    attempted_at: "2026-05-25T12:00:10Z"
  }, "admin.token");
  assert.equal(notificationDeliveryRecord.method, "POST");
  assert.equal(notificationDeliveryRecord.url, "/api/notifications/ntf%201/deliveries");
  assert.deepEqual(JSON.parse(notificationDeliveryRecord.body), {
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    status: "failed",
    provider_message_id: "provider msg 1",
    error_code: "invalid_openid",
    error_message: "openid missing",
    idempotency_key: "delivery:manual:ntf_1:wechat",
    attempted_at: "2026-05-25T12:00:10Z"
  });

  const notificationFailureAlert = buildAdminRequest(getAdminOperation("notification-failure-alert-emit"), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    limit: "5",
    now: "2026-05-25T12:01:00Z"
  }, "admin.token");
  assert.equal(notificationFailureAlert.method, "POST");
  assert.equal(notificationFailureAlert.url, "/api/admin/notification-deliveries/failure-alerts/emit");
  assert.deepEqual(JSON.parse(notificationFailureAlert.body), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    limit: 5,
    now: "2026-05-25T12:01:00Z"
  });

  const notificationRetrySchedule = buildAdminRequest(getAdminOperation("notification-delivery-retry-schedule"), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    limit: "5",
    retry_after_seconds: "300",
    now: "2026-05-25T12:02:00Z"
  }, "admin.token");
  assert.equal(notificationRetrySchedule.method, "POST");
  assert.equal(notificationRetrySchedule.url, "/api/admin/notification-deliveries/retries/schedule");
  assert.deepEqual(JSON.parse(notificationRetrySchedule.body), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    status: "failed",
    limit: 5,
    retry_after_seconds: 300,
    now: "2026-05-25T12:02:00Z"
  });

  const quietRetrySchedule = buildAdminRequest(getAdminOperation("notification-delivery-retry-schedule"), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "push",
    provider: "push",
    status: "queued",
    error_code: "notification_quiet_window",
    limit: "5",
    retry_at: "2026-05-25T12:12:00Z",
    now: "2026-05-25T12:03:00Z"
  }, "admin.token");
  assert.deepEqual(JSON.parse(quietRetrySchedule.body), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "push",
    provider: "push",
    status: "queued",
    error_code: "notification_quiet_window",
    limit: 5,
    retry_after_seconds: 300,
    retry_at: "2026-05-25T12:12:00Z",
    now: "2026-05-25T12:03:00Z"
  });

  const quietRetryAutoSchedule = buildAdminRequest(getAdminOperation("notification-quiet-window-retry-schedule"), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "push",
    provider: "push",
    limit: "10",
    retry_after_seconds: "0",
    now: "2026-05-25T12:10:00Z"
  }, "admin.token");
  assert.equal(quietRetryAutoSchedule.method, "POST");
  assert.equal(quietRetryAutoSchedule.url, "/api/admin/notification-deliveries/quiet-window-retries/schedule");
  assert.deepEqual(JSON.parse(quietRetryAutoSchedule.body), {
    target_role: "merchant",
    target_id: "merchant 19",
    channel: "push",
    provider: "push",
    limit: 10,
    retry_after_seconds: 0,
    now: "2026-05-25T12:10:00Z"
  });

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

  const afterSalesReview = buildAdminRequest(getAdminOperation("after-sales-review"), {
    request_id: "asr 231",
    decision: "approve",
    reason: "证据核验通过",
    refund_destination: "balance",
    refund_idempotency_key: "after_sales:asr_231"
  }, "admin.token");
  assert.equal(afterSalesReview.method, "POST");
  assert.equal(afterSalesReview.url, "/api/after-sales/asr%20231/review");
  assert.deepEqual(JSON.parse(afterSalesReview.body), {
    decision: "approve",
    reason: "证据核验通过",
    refund_destination: "balance",
    refund_idempotency_key: "after_sales:asr_231"
  });

  const supportTickets = buildAdminRequest(getAdminOperation("support-tickets"), {
    user_id: "user 1",
    status: "processing",
    sla_status: "overdue",
    assigned_support_id: "support 1",
    limit: "5",
    now: "2026-05-28T10:30:00Z"
  }, "admin.token");
  assert.equal(supportTickets.method, "GET");
  assert.equal(supportTickets.url, "/api/admin/service-tickets?user_id=user+1&status=processing&sla_status=overdue&assigned_support_id=support+1&limit=5&now=2026-05-28T10%3A30%3A00Z");

  const supportTicketDetail = buildAdminRequest(getAdminOperation("support-ticket-detail"), {
    ticket_id: "st 1"
  }, "admin.token");
  assert.equal(supportTicketDetail.method, "GET");
  assert.equal(supportTicketDetail.url, "/api/admin/service-tickets/st%201");

  const supportAssign = buildAdminRequest(getAdminOperation("support-ticket-assign"), {
    ticket_id: "st 1",
    support_id: "support_1",
    support_name: "客服小悦",
    actor_id: "ops_1"
  }, "admin.token");
  assert.equal(supportAssign.method, "POST");
  assert.equal(supportAssign.url, "/api/admin/service-tickets/st%201/assign");
  assert.deepEqual(JSON.parse(supportAssign.body), {
    support_id: "support_1",
    support_name: "客服小悦",
    actor_id: "ops_1"
  });

  const supportEscalate = buildAdminRequest(getAdminOperation("support-ticket-escalate"), {
    ticket_id: "st 1",
    reason: "超过 10 分钟未更新",
    escalation_level: "support_lead",
    actor_id: "ops_1"
  }, "admin.token");
  assert.equal(supportEscalate.method, "POST");
  assert.equal(supportEscalate.url, "/api/admin/service-tickets/st%201/escalate");
  assert.deepEqual(JSON.parse(supportEscalate.body), {
    reason: "超过 10 分钟未更新",
    escalation_level: "support_lead",
    actor_id: "ops_1"
  });

	  const supportResolve = buildAdminRequest(getAdminOperation("support-ticket-resolve"), {
	    ticket_id: "st 1",
	    solution: "已发放补偿券",
	    actor_id: "support_1"
	  }, "admin.token");
  assert.equal(supportResolve.method, "POST");
  assert.equal(supportResolve.url, "/api/admin/service-tickets/st%201/resolve");
	  assert.deepEqual(JSON.parse(supportResolve.body), {
	    solution: "已发放补偿券",
	    actor_id: "support_1"
	  });

	  const supportQualityReview = buildAdminRequest(getAdminOperation("support-quality-review"), {
	    ticket_id: "st 1",
	    score: "74",
	    result: "needs_coaching",
	    notes: "首响超时后补偿方案完整，需复盘主动同步话术",
	    coaching_required: "true",
	    reviewer_id: "quality_1",
	    reviewer_name: "质检主管"
	  }, "admin.token");
	  assert.equal(supportQualityReview.method, "POST");
	  assert.equal(supportQualityReview.url, "/api/admin/service-tickets/st%201/quality-review");
	  assert.deepEqual(JSON.parse(supportQualityReview.body), {
	    score: 74,
	    result: "needs_coaching",
	    notes: "首响超时后补偿方案完整，需复盘主动同步话术",
	    coaching_required: true,
	    reviewer_id: "quality_1",
	    reviewer_name: "质检主管"
	  });

	  const supportQualityReviews = buildAdminRequest(getAdminOperation("support-quality-reviews"), {
	    ticket_id: "st 1",
	    support_id: "support 1",
	    result: "needs_coaching",
	    coaching_required: "true",
	    limit: "5"
	  }, "admin.token");
	  assert.equal(supportQualityReviews.method, "GET");
	  assert.equal(supportQualityReviews.url, "/api/admin/service-ticket-quality-reviews?ticket_id=st+1&support_id=support+1&result=needs_coaching&coaching_required=true&limit=5");

	  const supportPerformance = buildAdminRequest(getAdminOperation("support-performance"), {
	    support_id: "support 1",
	    limit: "5",
	    now: "2026-05-28T10:30:00Z"
	  }, "admin.token");
	  assert.equal(supportPerformance.method, "GET");
	  assert.equal(supportPerformance.url, "/api/admin/service-ticket-performance?support_id=support+1&limit=5&now=2026-05-28T10%3A30%3A00Z");

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
  assert.deepEqual(buildAuditArchiveVerificationRows(auditArchiveVerificationsFromResult({ payload: { data: [{
    archive_id: "audit_archive_1",
    status: "verified",
    storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    manifest_hash: "abcdef1234567890",
    expected_content_hash: "content_expected_hash",
    actual_content_hash: "content_actual_hash",
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
  }] } })), [{
    id: "audit_archive_1:2026-05-24T00:00:03Z:0",
    archiveId: "audit_archive_1",
    status: "verified",
    statusTone: "ok",
    storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
    verifiedAt: "2026-05-24T00:00:03Z",
    manifestHashShort: "abcdef123456...",
    expectedContentHashShort: "content_expe...",
    actualContentHashShort: "content_actu...",
    bytesLabel: "1024/1024",
    logCountLabel: "1/1",
    matchSummary: "archive / manifest / content / bytes / log count",
    errorLabel: "",
    raw: {
      archiveId: "audit_archive_1",
      status: "verified",
      storageKey: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
      manifestHash: "abcdef1234567890",
      expectedContentHash: "content_expected_hash",
      actualContentHash: "content_actual_hash",
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
    }
  }]);

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
  const notificationRoute = auditTargetRoute({ target_type: "platform_notification", action: "admin.notification.delivery_recorded" });
  assert.deepEqual(notificationRoute, { module: "notifications", operation: "notifications", label: "通知运营" });
  const notificationAlertRoute = auditTargetRoute({ target_type: "notification_delivery_alerts", action: "admin.notification_delivery_failure_alerts.emitted" });
  assert.deepEqual(notificationAlertRoute, { module: "notifications", operation: "notification-failure-alert-emit", label: "通知失败告警" });
  const notificationRetryRoute = auditTargetRoute({ target_type: "notification_delivery_retries", action: "admin.notification_delivery_retries.scheduled" });
  assert.deepEqual(notificationRetryRoute, { module: "notifications", operation: "notification-delivery-retry-schedule", label: "通知重试计划" });
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
      {
        rider_id: "rider_1",
        average_accept_seconds: 18.4,
        completion_rate: 0.98,
        rider_average_rating: 4.9,
        rider_review_count: 12,
        score: 118,
        score_breakdown: { accept_score: 44.8, order_volume_score: 33.6, completion_score: 14.7, rating_score: 11.8 },
        recent_trend: [
          { date: "2026-05-20", score: 112, completed_orders: 18, average_rating: 4.7, timeout_count: 0, reject_count: 1 },
          { date: "2026-05-21", score: 118, completed_orders: 21, average_rating: 5, timeout_count: 0, reject_count: 0 },
          { date: "2026-05-22", score: 109, completed_orders: 16, average_rating: 4.5, timeout_count: 1, reject_count: 0 }
        ],
        recent_reviews: [
          { review_id: "rev_1", order_id: "ord_1", rider_rating: 5, content: "送达很快，态度也好", created_at: "2026-05-22T12:20:00Z" },
          { review_id: "rev_2", order_id: "ord_2", rider_rating: 4, content: "包装完整，就是高峰期稍慢", created_at: "2026-05-21T11:10:00Z" }
        ],
        exception_summary: {
          dispatch_timeout_count: 1,
          dispatch_reject_count: 0,
          after_sales_count: 1,
          low_rating_count: 0,
          last_event_at: "2026-05-22T12:25:00Z"
        },
        exception_details: [
          { kind: "dispatch_timeout", label: "派单超时", order_id: "ord_1", dispatch_event_id: "dpe_1", status: "dispatch.timeout", message: "assignment_timeout", created_at: "2026-05-22T12:05:00Z" },
          { kind: "after_sales", label: "售后介入", order_id: "ord_1", after_sales_request_id: "asr_1", status: "admin_review", message: "平台已介入核实", created_at: "2026-05-22T12:25:00Z" },
          { kind: "low_rating", label: "低分评价", order_id: "ord_2", review_id: "rev_9", status: "2 星", message: "配送评分 2 星：高峰期晚到了十分钟", created_at: "2026-05-22T11:45:00Z" }
        ],
        level: "S",
        dispatch_priority: 40
      }
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
  assert.equal(performanceView.rows[0][3], "4.9 / 12");
  assert.equal(performanceView.rows[0][4], "118");
  assert.equal(performanceView.rows[0][5], "接单 45 / 单量 34 / 履约 15 / 评分 12");
  assert.equal(performanceView.detailRows[0].facts[0].value, "05-20 112分 / 18单 / 4.7星 | 05-21 118分 / 21单 / 5.0星 | 05-22 109分 / 16单 / 4.5星");
  assert.equal(performanceView.detailRows[0].facts[1].value, "5星 送达很快，态度也好 | 4星 包装完整，就是高峰期稍慢");
  assert.equal(performanceView.detailRows[0].facts[2].value, "超时 1 / 拒单 0 / 售后 1 / 低分 0 / 最近 05/22 20:25");
  assert.equal(performanceView.detailRows[0].facts[3].value, "派单超时 / ord_1 / dispatch.timeout / assignment_timeout / 05/22 20:05");
  assert.equal(performanceView.detailRows[0].facts[4].value, "售后介入 / ord_1 / admin_review / 平台已介入核实 / 05/22 20:25");
  assert.equal(performanceView.detailRows[0].facts[5].value, "低分评价 / ord_2 / 2 星 / 配送评分 2 星：高峰期晚到了十分钟 / 05/22 19:45");

  const performanceDetail = buildAdminBusinessDetail(performanceView, 0);
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "dispatch-order-events" && action.values.order_id === "ord_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "after-sales-detail" && action.values.request_id === "asr_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "after-sales-events" && action.values.request_id === "asr_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "after-sales-evidence" && action.values.request_id === "asr_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "after-sales-list" && action.values.request_id === "asr_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "refund-transactions" && action.values.order_id === "ord_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "support-tickets" && action.values.related_order_id === "ord_1"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "audit-logs" && action.values.target_id === "ord_1" && action.values.action === ""));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "audit-logs" && action.values.target_id === "ord_1" && action.values.action === "admin.order.refunded"));
  assert.ok(performanceDetail.actions.some((action) => action.operationKey === "audit-logs" && action.values.target_id === "ord_2"));

  const kpis = buildSnapshotKpis(snapshot, ADMIN_WEB_KPIS);
  assert.equal(kpis[0].value, "3");
  assert.equal(kpis[3].value, "2");

  const queues = buildSnapshotQueues(snapshot, ADMIN_WEB_QUEUES);
  assert.equal(queues[1].title, "商户资质/保证金");
  assert.equal(queues[1].operationKey, "merchant-qualifications");
  assert.equal(queues[4].target, "Ready 2 / Blocked 1");

  assert.equal(snapshotDataFromResult({ payload: { data: snapshot } }), snapshot);
});
