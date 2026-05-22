import assert from "node:assert/strict";
import test from "node:test";
import { ADMIN_API_OPERATIONS, buildAdminRequest, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS, getAdminWebModule } from "./config.mjs";

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
    "object-cleanup-stats",
    "object-cleanup-candidates",
    "outbox-stats",
    "outbox-events",
    "outbox-replay-batch",
    "order-compensate"
  ]) {
    assert.ok(getAdminOperation(key), `missing ${key}`);
  }
  assert.ok(ADMIN_API_OPERATIONS.every((operation) => operation.method && operation.path && operation.title));
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
