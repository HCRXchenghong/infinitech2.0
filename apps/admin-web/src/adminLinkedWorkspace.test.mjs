import assert from "node:assert/strict";
import test from "node:test";
import {
  buildLinkedResultEntry,
  canInlinePreviewAction,
  linkedResultEntryId,
  linkedResultErrorMessage,
  linkedResultFailureFacts,
  linkedResultGroup,
  linkedResultMatchesFilter,
  linkedResultMatchesFocus,
  linkedResultAttemptTrail,
  linkedResultPrefillAction,
  linkedResultRetrySummary,
  linkedOperationPrefillValues,
  linkedResultSyncState,
  linkedWorkspaceSyncActions,
  linkedWorkspaceSyncActionGroups,
  linkedWorkspaceSyncEntries,
  linkedWorkspaceSyncOverview,
  linkedWorkspaceFailedActions,
  linkedWorkspaceFailedGroups,
  linkedResultTokens,
  linkedResultsFilterOptions,
  linkedWorkspaceSyncGroups,
  linkedWorkspaceContextCandidates,
  linkedWorkspaceContextEquals,
  linkedWorkspacePrimaryFocusKey,
  normalizeLinkedWorkspaceContext,
  syncLinkedResultValues,
  linkedWorkspaceBundles,
  linkedWorkspaceContext,
  linkedWorkspacePrimaryActions,
  upsertLinkedResult
} from "./adminLinkedWorkspace.mjs";

test("canInlinePreviewAction only auto-runs GET operations", () => {
  assert.equal(canInlinePreviewAction({ operationKey: "order-detail" }), true);
  assert.equal(canInlinePreviewAction({ operationKey: "order-refund" }), false);
  assert.equal(canInlinePreviewAction({ operationKey: "unknown-operation" }), false);
});

test("linkedResultEntryId uses operation url for dedupe", () => {
  const action = {
    operationKey: "refund-transactions",
    values: { order_id: "ord_9", limit: 20 }
  };
  assert.equal(linkedResultEntryId(action), "refund-transactions:/api/admin/refunds?order_id=ord_9&limit=20");
});

test("upsertLinkedResult keeps latest entry and enforces limit", () => {
  const first = buildLinkedResultEntry({
    operationKey: "order-detail",
    label: "订单总览",
    values: { order_id: "ord_1" }
  }, { ok: true, request: { url: "/api/admin/orders/ord_1" } });
  const second = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    label: "退款流水",
    values: { order_id: "ord_1", limit: 20 }
  }, { ok: true, request: { url: "/api/admin/refunds?order_id=ord_1&limit=20" } });
  const newerFirst = {
    ...first,
    status: "error"
  };

  const seeded = upsertLinkedResult([second], first, 2);
  const combined = upsertLinkedResult(seeded, newerFirst, 2);
  assert.equal(combined.length, 2);
  assert.equal(combined[0].id, first.id);
  assert.equal(combined[0].status, "error");
  assert.equal(combined[1].id, second.id);
  assert.equal(combined[0].lastFailure?.message, "请求失败，请稍后重试。");
  assert.equal(combined[0].lastSuccess?.status, "ready");
});

test("upsertLinkedResult preserves failure and success traces across retries", () => {
  const failed = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    values: { order_id: "ord_2", limit: 20 }
  }, {
    status: 502,
    request: { method: "GET", url: "/api/admin/refunds?order_id=ord_2&limit=20" },
    payload: { error: "bad gateway" }
  }, "error");
  const ready = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    values: { order_id: "ord_2", limit: 20 }
  }, {
    status: 200,
    request: { method: "GET", url: "/api/admin/refunds?order_id=ord_2&limit=20" },
    payload: { data: [] }
  }, "ready");

  const mergedFailure = upsertLinkedResult([], failed, 2);
  const mergedSuccess = upsertLinkedResult(mergedFailure, ready, 2);
  assert.equal(mergedSuccess[0].lastFailure?.statusLabel, "HTTP 502");
  assert.equal(mergedSuccess[0].lastSuccess?.statusLabel, "HTTP 200");
  assert.deepEqual(
    linkedResultAttemptTrail(mergedSuccess[0]).map((item) => item.key),
    ["failure", "success"]
  );
});

test("linked results expose failed filter and readable error summary", () => {
  const failedRefund = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    values: { order_id: "ord_1", limit: 20 }
  }, {
    status: 504,
    request: { url: "/api/admin/refunds?order_id=ord_1&limit=20" },
    payload: { error: "refund gateway timeout" }
  }, "error");
  const storedFailedRefund = upsertLinkedResult([], failedRefund, 2)[0];
  const readySupport = buildLinkedResultEntry({
    operationKey: "support-tickets",
    values: { related_order_id: "ord_1", user_id: "user_1", limit: 20 }
  }, {
    status: 200,
    request: { url: "/api/admin/service-tickets?related_order_id=ord_1&user_id=user_1&limit=20" },
    payload: { data: [] }
  }, "ready");

  assert.deepEqual(
    linkedResultsFilterOptions([storedFailedRefund, readySupport]).map((option) => `${option.key}:${option.count}`),
    ["all:2", "failed:1", "finance:1", "support:1"]
  );
  assert.equal(linkedResultMatchesFilter(storedFailedRefund, "failed"), true);
  assert.equal(linkedResultMatchesFilter(readySupport, "failed"), false);
  assert.equal(linkedResultErrorMessage(storedFailedRefund), "refund gateway timeout");
  assert.deepEqual(linkedResultFailureFacts(storedFailedRefund), {
    statusLabel: "HTTP 504",
    requestLabel: "GET /api/admin/refunds?order_id=ord_1&limit=20",
    paramsLabel: "order_id=ord_1 / limit=20"
  });
  assert.equal(linkedResultRetrySummary(storedFailedRefund, { orderId: "ord_9" }), "order_id=ord_9");
  assert.equal(linkedResultRetrySummary(storedFailedRefund, { orderId: "ord_1" }), "沿用当前参数");
  assert.deepEqual(linkedResultPrefillAction(storedFailedRefund, { orderId: "ord_9" }), {
    label: "退款流水",
    operationKey: "refund-transactions",
    values: { order_id: "ord_9", limit: 20 }
  });
  assert.deepEqual(
    linkedResultAttemptTrail(storedFailedRefund).map((item) => item.key),
    ["failure"]
  );
});

test("linkedWorkspaceContext extracts order and request facts from aggregated detail results", () => {
  const orderContext = linkedWorkspaceContext({
    operation: { key: "order-detail" },
    request: { url: "/api/admin/orders/ord_9" },
    payload: {
      data: {
        order: {
          id: "ord_9",
          user_id: "user_9"
        },
        after_sales_requests: [
          { id: "asr_9" }
        ]
      }
    }
  });
  assert.deepEqual(orderContext, {
    operationKey: "order-detail",
    orderId: "ord_9",
    requestId: "asr_9",
    ticketId: "",
    userId: "user_9"
  });

  const afterSalesContext = linkedWorkspaceContext({
    operation: { key: "after-sales-detail" },
    request: { url: "/api/admin/after-sales/asr_9" },
    payload: {
      data: {
        request: {
          id: "asr_9",
          order_id: "ord_9",
          user_id: "user_9"
        },
        service_tickets: [
          { id: "st_9" }
        ]
      }
    }
  });
  assert.deepEqual(afterSalesContext, {
    operationKey: "after-sales-detail",
    orderId: "ord_9",
    requestId: "asr_9",
    ticketId: "st_9",
    userId: "user_9"
  });
});

test("linked workspace exposes direct actions and bundles for shared order context", () => {
  const context = {
    orderId: "ord_9",
    requestId: "asr_9",
    ticketId: "st_9",
    userId: "user_9"
  };
  const actions = linkedWorkspacePrimaryActions(context);
  const bundles = linkedWorkspaceBundles(context);

  assert.ok(actions.some((action) => action.operationKey === "refund-transactions" && action.values.order_id === "ord_9"));
  assert.ok(actions.some((action) => action.operationKey === "support-tickets" && action.values.related_order_id === "ord_9"));
  assert.ok(actions.some((action) => action.operationKey === "after-sales-events" && action.values.request_id === "asr_9"));
  assert.ok(actions.some((action) => action.operationKey === "support-ticket-detail" && action.values.ticket_id === "st_9"));

  const orderWorkspace = bundles.find((bundle) => bundle.key === "order-workspace");
  assert.ok(orderWorkspace);
  assert.equal(orderWorkspace.actions.length, 4);
  assert.equal(orderWorkspace.actions[0].operationKey, "refund-transactions");

  const afterSalesWorkspace = bundles.find((bundle) => bundle.key === "after-sales-workspace");
  assert.ok(afterSalesWorkspace);
  assert.ok(afterSalesWorkspace.actions.some((action) => action.operationKey === "after-sales-evidence"));
});

test("linked results expose group filters and focus tokens", () => {
  const refundEntry = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    values: { order_id: "ord_9", limit: 20 }
  }, {
    operation: { key: "refund-transactions" },
    request: { url: "/api/admin/refunds?order_id=ord_9&limit=20" },
    payload: { data: [{ order_id: "ord_9", user_id: "user_9" }] }
  });
  const supportEntry = buildLinkedResultEntry({
    operationKey: "support-tickets",
    values: { related_order_id: "ord_9", user_id: "user_9", limit: 20 }
  }, {
    operation: { key: "support-tickets" },
    request: { url: "/api/admin/service-tickets?related_order_id=ord_9&user_id=user_9&limit=20" },
    payload: { data: [{ id: "st_9", related_order_id: "ord_9", user_id: "user_9" }] }
  });
  const auditEntry = buildLinkedResultEntry({
    operationKey: "audit-logs",
    values: { target_type: "order", target_id: "ord_9", limit: 20 }
  }, {
    operation: { key: "audit-logs" },
    request: { url: "/api/admin/audit-logs?target_type=order&target_id=ord_9&limit=20" },
    payload: { data: [{ target_type: "order", target_id: "ord_9" }] }
  });

  assert.equal(linkedResultGroup(refundEntry.operationKey), "finance");
  assert.equal(linkedResultGroup(supportEntry.operationKey), "support");

  const options = linkedResultsFilterOptions([refundEntry, supportEntry, auditEntry]);
  assert.deepEqual(options.map((option) => option.key), ["all", "finance", "support", "audit"]);
  assert.equal(options.find((option) => option.key === "support")?.count, 1);

  const tokens = linkedResultTokens(supportEntry);
  assert.ok(tokens.some((token) => token.focusKey === "order:ord_9"));
  assert.ok(tokens.some((token) => token.focusKey === "user:user_9"));

  assert.equal(linkedResultMatchesFilter(refundEntry, "finance"), true);
  assert.equal(linkedResultMatchesFilter(refundEntry, "support"), false);
  assert.equal(linkedResultMatchesFocus(auditEntry, "order:ord_9"), true);
  assert.equal(linkedResultMatchesFocus(auditEntry, "service_ticket:st_9"), false);
});

test("syncLinkedResultValues rebases shared order context across operations", () => {
  assert.deepEqual(
    syncLinkedResultValues("refund-transactions", { order_id: "ord_9", limit: 20 }, { orderId: "ord_10" }),
    { order_id: "ord_10", limit: 20 }
  );

  assert.deepEqual(
    syncLinkedResultValues(
      "support-tickets",
      { related_order_id: "ord_9", user_id: "user_9", limit: 20 },
      { orderId: "ord_10", userId: "user_10" }
    ),
    { related_order_id: "ord_10", user_id: "user_10", limit: 20 }
  );

  assert.deepEqual(
    syncLinkedResultValues(
      "audit-logs",
      { target_type: "after_sales", target_id: "asr_9", limit: 20 },
      { requestId: "asr_10" }
    ),
    { target_type: "after_sales", target_id: "asr_10", limit: 20 }
  );
});

test("linkedResultSyncState marks stale linked cards and provides target context tokens", () => {
  const refundEntry = buildLinkedResultEntry({
    operationKey: "refund-transactions",
    label: "退款流水",
    values: { order_id: "ord_9", limit: 20 }
  }, {
    operation: { key: "refund-transactions" },
    request: { url: "/api/admin/refunds?order_id=ord_9&limit=20" },
    payload: { data: [{ order_id: "ord_9", user_id: "user_9" }] }
  });

  const syncState = linkedResultSyncState(refundEntry, { orderId: "ord_10", userId: "user_10" });
  assert.equal(syncState.stale, true);
  assert.equal(syncState.action?.values.order_id, "ord_10");
  assert.ok(syncState.targetTokens.some((token) => token.focusKey === "order:ord_10"));

  const settledSyncState = linkedResultSyncState(refundEntry, { orderId: "ord_9" });
  assert.equal(settledSyncState.stale, false);
});

test("linked workspace context helpers normalize compare and expose context candidates", () => {
  const normalized = normalizeLinkedWorkspaceContext({ orderId: " ord_9 ", requestId: "asr_9" });
  assert.deepEqual(normalized, {
    operationKey: "",
    orderId: "ord_9",
    requestId: "asr_9",
    ticketId: "",
    userId: ""
  });

  assert.equal(
    linkedWorkspaceContextEquals(
      { orderId: "ord_9", requestId: "asr_9" },
      { orderId: "ord_9", requestId: "asr_9", ticketId: "", userId: "" }
    ),
    true
  );
  assert.equal(linkedWorkspacePrimaryFocusKey({ orderId: "ord_9", requestId: "asr_9" }), "order:ord_9");

  const entries = [
    buildLinkedResultEntry({
      operationKey: "refund-transactions",
      title: "退款流水",
      values: { order_id: "ord_9", limit: 20 }
    }, {
      operation: { key: "refund-transactions" },
      request: { url: "/api/admin/refunds?order_id=ord_9&limit=20" },
      payload: { data: [{ order_id: "ord_9", user_id: "user_9" }] }
    }),
    buildLinkedResultEntry({
      operationKey: "after-sales-detail",
      title: "售后详情",
      values: { request_id: "asr_9" }
    }, {
      operation: { key: "after-sales-detail" },
      request: { url: "/api/admin/after-sales/asr_9" },
      payload: { data: { request: { id: "asr_9", order_id: "ord_9", user_id: "user_9" }, service_tickets: [{ id: "st_9" }] } }
    })
  ];

  const candidates = linkedWorkspaceContextCandidates(entries, { orderId: "ord_9", userId: "user_9" });
  assert.equal(candidates.length, 1);
  assert.equal(candidates[0].context.requestId, "asr_9");
  assert.equal(candidates[0].context.ticketId, "st_9");
  assert.deepEqual(candidates[0].sourceTitles, ["售后聚合详情"]);
});

test("linked workspace sync groups and operation prefills follow active context", () => {
  const entries = [
    buildLinkedResultEntry({
      operationKey: "refund-transactions",
      values: { order_id: "ord_9", limit: 20 }
    }, {
      operation: { key: "refund-transactions" },
      request: { url: "/api/admin/refunds?order_id=ord_9&limit=20" },
      payload: { data: [{ order_id: "ord_9", user_id: "user_9" }] }
    }),
    buildLinkedResultEntry({
      operationKey: "support-tickets",
      values: { related_order_id: "ord_9", user_id: "user_9", limit: 20 }
    }, {
      operation: { key: "support-tickets" },
      request: { url: "/api/admin/service-tickets?related_order_id=ord_9&user_id=user_9&limit=20" },
      payload: { data: [{ id: "st_9", related_order_id: "ord_9", user_id: "user_9" }] }
    })
  ];

  const syncGroups = linkedWorkspaceSyncGroups(entries, { orderId: "ord_10", userId: "user_10" });
  assert.deepEqual(syncGroups.map((group) => `${group.key}:${group.count}`), ["finance:1", "support:1"]);
  assert.equal(syncGroups[0].actions[0].values.order_id, "ord_10");
  assert.equal(syncGroups[1].actions[0].values.related_order_id, "ord_10");
  assert.equal(
    linkedWorkspaceSyncEntries(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "support", focusKey: "order:ord_10" }).length,
    1
  );
  assert.deepEqual(
    linkedWorkspaceSyncActions(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "all", focusKey: "order:ord_10", groupKey: "finance" })
      .map((action) => action.operationKey),
    ["refund-transactions"]
  );
  assert.deepEqual(
    linkedWorkspaceSyncOverview(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "all", focusKey: "order:ord_10" })
      .map((group) => `${group.key}:${group.stale}/${group.total}`),
    ["finance:1/1", "support:1/1"]
  );
  assert.deepEqual(
    linkedWorkspaceSyncActionGroups(
      linkedWorkspaceSyncActions(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "all", focusKey: "order:ord_10" })
    ).map((group) => `${group.key}:${group.count}`),
    ["finance:1", "support:1"]
  );

  assert.deepEqual(
    linkedOperationPrefillValues("support-tickets", { limit: 20 }, { orderId: "ord_10", userId: "user_10" }),
    { limit: 20, related_order_id: "ord_10", user_id: "user_10" }
  );
  assert.deepEqual(
    linkedOperationPrefillValues("audit-logs", { limit: 20 }, { requestId: "asr_10", orderId: "ord_10" }),
    { limit: 20, target_type: "after_sales", target_id: "asr_10" }
  );
});

test("linked workspace failed actions and groups stay scoped to current context", () => {
  const entries = [
    buildLinkedResultEntry({
      operationKey: "refund-transactions",
      values: { order_id: "ord_9", limit: 20 }
    }, {
      operation: { key: "refund-transactions" },
      request: { url: "/api/admin/refunds?order_id=ord_9&limit=20" },
      payload: { error: "timeout" }
    }, "error"),
    buildLinkedResultEntry({
      operationKey: "support-tickets",
      values: { related_order_id: "ord_9", user_id: "user_9", limit: 20 }
    }, {
      operation: { key: "support-tickets" },
      request: { url: "/api/admin/service-tickets?related_order_id=ord_9&user_id=user_9&limit=20" },
      payload: { error: "gateway" }
    }, "error"),
    buildLinkedResultEntry({
      operationKey: "dispatch-order-events",
      values: { order_id: "ord_9" }
    }, {
      operation: { key: "dispatch-order-events" },
      request: { url: "/api/admin/dispatch/events?order_id=ord_9" },
      payload: { data: [] }
    }, "ready")
  ];

  assert.deepEqual(
    linkedWorkspaceFailedActions(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "all", focusKey: "order:ord_10" })
      .map((action) => `${action.operationKey}:${action.values.order_id || action.values.related_order_id}`),
    ["refund-transactions:ord_10", "support-tickets:ord_10"]
  );
  assert.deepEqual(
    linkedWorkspaceFailedGroups(entries, { orderId: "ord_10", userId: "user_10" }, { filterKey: "all", focusKey: "order:ord_10" })
      .map((group) => `${group.key}:${group.count}`),
    ["finance:1", "support:1"]
  );
});
