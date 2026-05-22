import assert from "node:assert/strict";
import test from "node:test";
import { ADMIN_API_OPERATIONS, buildAdminRequest, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS, getAdminWebModule } from "./config.mjs";
import { ADMIN_WEB_VIEWS, getAdminView } from "./adminViews.mjs";
import { applySnapshotToAdminView, buildSnapshotKpis, buildSnapshotQueues, snapshotDataFromResult } from "./adminSnapshot.mjs";

test("admin web exposes the first operable control-center modules", () => {
  for (const key of ["orders", "after-sales", "merchants", "riders", "dispatch", "refund-settings", "payment", "support", "rtc", "integrations"]) {
    assert.ok(getAdminWebModule(key), `missing ${key}`);
  }
  assert.ok(ADMIN_WEB_SECTIONS.length >= 4);
  assert.ok(ADMIN_WEB_MODULES.length >= 25);
  assert.ok(ADMIN_WEB_KPIS.some((item) => item.key === "outbox"));
  assert.ok(ADMIN_WEB_QUEUES.some((item) => item.operationKey === "object-cleanup-stats"));
  assert.ok(ADMIN_WEB_RBAC.some((item) => item.role === "finance_admin" && item.scopes.includes("refund:write")));
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
  for (const key of ["orders", "after-sales", "merchants", "riders", "rider-performance", "dispatch", "refund-settings"]) {
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

  const auditRequest = buildAdminRequest(getAdminOperation("audit-logs"), { target_type: "order", limit: 5 }, "Bearer admin.token");
  assert.equal(auditRequest.url, "/api/admin/audit-logs?target_type=order&limit=5");
  assert.equal(auditRequest.headers.Authorization, "Bearer admin.token");

  const loginFields = fieldsForOperation(getAdminOperation("admin-login"));
  assert.deepEqual(loginFields.map((field) => field.key), ["account_id", "password"]);
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
