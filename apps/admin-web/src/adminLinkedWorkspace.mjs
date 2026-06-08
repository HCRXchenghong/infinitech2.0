import { buildAdminRequest, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";

const DEFAULT_LINKED_RESULT_LIMIT = 4;
const LINKED_RESULT_GROUP_LABELS = Object.freeze({
  all: "全部",
  order: "订单",
  "after-sales": "售后",
  finance: "退款",
  support: "客服",
  dispatch: "派单",
  audit: "审计",
  other: "其他"
});

function safeObject(value) {
  return value && typeof value === "object" && !Array.isArray(value) ? value : {};
}

function safeArray(value) {
  return Array.isArray(value) ? value : [];
}

function hasText(value) {
  return String(value || "").trim().length > 0;
}

function linkedResultFallbackId(operationKey, values = {}) {
  return `${operationKey}:${JSON.stringify(values)}`;
}

export function canInlinePreviewAction(action) {
  const operation = getAdminOperation(action?.operationKey);
  return Boolean(operation && operation.method === "GET");
}

export function linkedResultEntryId(action) {
  const operation = getAdminOperation(action?.operationKey);
  if (!operation) {
    return "";
  }
  try {
    const request = buildAdminRequest(operation, action?.values || {});
    return `${operation.key}:${request.url}`;
  } catch {
    return linkedResultFallbackId(operation.key, action?.values || {});
  }
}

export function buildLinkedResultEntry(action, result, status = "ready") {
  const operation = getAdminOperation(action?.operationKey);
  const request = operation ? buildAdminRequest(operation, action?.values || {}) : null;
  return {
    id: linkedResultEntryId(action),
    title: action?.label || operation?.title || "联动结果",
    operationKey: operation?.key || action?.operationKey || "",
    requestUrl: request?.url || "",
    status,
    updatedAt: new Date().toISOString(),
    values: { ...(action?.values || {}) },
    result
  };
}

export function linkedResultErrorMessage(entry) {
  const payload = safeObject(entry?.result?.payload);
  const payloadErrors = safeArray(payload.errors)
    .map((item) => {
      if (!item) {
        return "";
      }
      if (typeof item === "string") {
        return item.trim();
      }
      if (typeof item === "object") {
        return firstNonEmpty(item.message, item.error, item.detail, item.reason);
      }
      return String(item).trim();
    })
    .filter(Boolean)
    .join(" / ");
  const status = Number(entry?.result?.status || 0);
  return firstNonEmpty(
    payload.error,
    payload.message,
    payload.detail,
    payload.reason,
    payloadErrors,
    status > 0 ? `请求失败（HTTP ${status}）` : "",
    "请求失败，请稍后重试。"
  );
}

function summarizedValue(value) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item ?? "").trim()).filter(Boolean).join(",");
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  if (value === null || value === undefined) {
    return "";
  }
  return String(value).trim();
}

function summarizedParamEntries(values) {
  return Object.entries(safeObject(values))
    .map(([key, value]) => [String(key || "").trim(), summarizedValue(value)])
    .filter(([key, value]) => key && value)
    .map(([key, value]) => `${key}=${value}`);
}

export function linkedResultFailureFacts(entry) {
  const request = safeObject(entry?.result?.request);
  const status = Number(entry?.result?.status || 0);
  const requestUrl = firstNonEmpty(entry?.requestUrl, request.url, "-");
  const requestMethod = firstNonEmpty(request.method, "GET").toUpperCase();
  const params = summarizedParamEntries(entry?.values).slice(0, 5);
  return {
    statusLabel: status > 0 ? `HTTP ${status}` : "请求失败",
    requestLabel: `${requestMethod} ${requestUrl}`,
    paramsLabel: params.length > 0 ? params.join(" / ") : "无附加参数"
  };
}

export function linkedResultRetrySummary(entry, context) {
  const action = linkedResultSyncAction(entry, context);
  if (!action) {
    return "";
  }
  const currentEntries = new Set(summarizedParamEntries(entry?.values));
  const nextEntries = summarizedParamEntries(action.values);
  const changedEntries = nextEntries.filter((item) => !currentEntries.has(item));
  return changedEntries.length > 0 ? changedEntries.join(" / ") : "沿用当前参数";
}

export function linkedResultPrefillAction(entry, context) {
  const syncAction = linkedResultSyncAction(entry, context);
  if (syncAction) {
    return syncAction;
  }
  if (!entry?.operationKey) {
    return null;
  }
  return {
    operationKey: entry.operationKey,
    values: { ...(entry.values || {}) }
  };
}

function buildLinkedResultAttemptSnapshot(entry) {
  const status = String(entry?.status || "").trim();
  if (!status || status === "loading") {
    return null;
  }
  const request = safeObject(entry?.result?.request);
  const requestMethod = firstNonEmpty(request.method, "GET").toUpperCase();
  const requestUrl = firstNonEmpty(entry?.requestUrl, request.url, "-");
  const statusCode = Number(entry?.result?.status || 0);
  const params = summarizedParamEntries(entry?.values).slice(0, 5);
  return {
    status,
    at: String(entry?.updatedAt || "").trim() || new Date().toISOString(),
    statusLabel: status === "ready"
      ? (statusCode > 0 ? `HTTP ${statusCode}` : "请求成功")
      : (statusCode > 0 ? `HTTP ${statusCode}` : "请求失败"),
    requestLabel: `${requestMethod} ${requestUrl}`,
    paramsLabel: params.length > 0 ? params.join(" / ") : "无附加参数",
    message: status === "error" ? linkedResultErrorMessage(entry) : ""
  };
}

function mergeLinkedResultEntry(previousEntry, nextEntry) {
  const previous = previousEntry && typeof previousEntry === "object" ? previousEntry : null;
  const next = nextEntry && typeof nextEntry === "object" ? nextEntry : null;
  if (!next) {
    return null;
  }
  if (!previous) {
    const snapshot = buildLinkedResultAttemptSnapshot(next);
    return {
      ...next,
      lastSuccess: next.status === "ready" ? snapshot : null,
      lastFailure: next.status === "error" ? snapshot : null
    };
  }
  const snapshot = buildLinkedResultAttemptSnapshot(next);
  return {
    ...previous,
    ...next,
    lastSuccess: next.status === "ready"
      ? snapshot
      : (previous.lastSuccess || null),
    lastFailure: next.status === "error"
      ? snapshot
      : (previous.lastFailure || null)
  };
}

export function linkedResultAttemptTrail(entry) {
  const normalizedEntry = entry && typeof entry === "object" ? entry : null;
  if (!normalizedEntry) {
    return [];
  }
  return [
    normalizedEntry.lastFailure ? { key: "failure", ...normalizedEntry.lastFailure } : null,
    normalizedEntry.lastSuccess ? { key: "success", ...normalizedEntry.lastSuccess } : null
  ].filter(Boolean);
}

export function upsertLinkedResult(entries, entry, limit = DEFAULT_LINKED_RESULT_LIMIT) {
  const nextEntry = entry && typeof entry === "object" ? entry : null;
  if (!nextEntry?.id) {
    return Array.isArray(entries) ? entries.slice(0, limit) : [];
  }
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  const previousEntry = normalizedEntries.find((item) => item.id === nextEntry.id) || null;
  const mergedEntry = mergeLinkedResultEntry(previousEntry, nextEntry);
  return [mergedEntry, ...normalizedEntries.filter((item) => item.id !== nextEntry.id)].slice(0, limit);
}

function requestUrl(result) {
  return String(result?.request?.url || "");
}

function requestParam(result, key) {
  const url = requestUrl(result);
  if (!url) {
    return "";
  }
  try {
    const parsed = new URL(url, "http://admin.local");
    return parsed.searchParams.get(key) || "";
  } catch {
    return "";
  }
}

function firstNonEmpty(...values) {
  for (const value of values) {
    const normalized = String(value || "").trim();
    if (normalized) {
      return normalized;
    }
  }
  return "";
}

function setIfPresent(target, key, value) {
  if (hasText(value)) {
    target[key] = String(value).trim();
  }
}

function dedupeActions(actions) {
  const seen = new Set();
  return actions.filter((action) => {
    const id = linkedResultEntryId(action);
    if (!id || seen.has(id)) {
      return false;
    }
    seen.add(id);
    return true;
  });
}

export function linkedWorkspaceContext(result) {
  const operationKey = String(result?.operation?.key || "");
  const data = safeObject(result?.payload?.data);
  const order = safeObject(data.order);
  const request = safeObject(data.request);
  const ticket = safeObject(data.ticket);
  const firstRefund = safeObject(safeArray(data.refunds)[0]);
  const firstTicket = safeObject(safeArray(data.service_tickets)[0]);
  const firstAfterSales = safeObject(safeArray(data.after_sales_requests)[0]);
  const firstTimelineItem = safeObject(safeArray(result?.payload?.data)[0]);

  const orderId = firstNonEmpty(
    order.id,
    request.order_id,
    ticket.related_order_id,
    firstRefund.order_id,
    firstAfterSales.order_id,
    firstTimelineItem.target_type === "order" ? firstTimelineItem.target_id : "",
    firstTimelineItem.order_id,
    requestParam(result, "related_order_id"),
    requestParam(result, "order_id")
  );
  const requestId = firstNonEmpty(
    request.id,
    firstAfterSales.id,
    firstTimelineItem.target_type === "after_sales" ? firstTimelineItem.target_id : "",
    firstTimelineItem.request_id,
    requestParam(result, "request_id")
  );
  const ticketId = firstNonEmpty(
    ticket.id,
    firstTicket.id,
    firstTimelineItem.target_type === "service_ticket" ? firstTimelineItem.target_id : "",
    requestParam(result, "ticket_id")
  );
  const userId = firstNonEmpty(
    order.user_id,
    request.user_id,
    ticket.user_id,
    firstRefund.user_id,
    firstTimelineItem.user_id,
    requestParam(result, "user_id")
  );

  return {
    operationKey,
    orderId,
    requestId,
    ticketId,
    userId
  };
}

export function normalizeLinkedWorkspaceContext(context) {
  return {
    operationKey: String(context?.operationKey || "").trim(),
    orderId: String(context?.orderId || "").trim(),
    requestId: String(context?.requestId || "").trim(),
    ticketId: String(context?.ticketId || "").trim(),
    userId: String(context?.userId || "").trim()
  };
}

export function linkedResultContext(entry) {
  const resultContext = linkedWorkspaceContext(entry?.result);
  const values = safeObject(entry?.values);
  return normalizeLinkedWorkspaceContext({
    operationKey: firstNonEmpty(entry?.operationKey, resultContext.operationKey),
    orderId: firstNonEmpty(resultContext.orderId, values.order_id, values.related_order_id, values.target_type === "order" ? values.target_id : ""),
    requestId: firstNonEmpty(resultContext.requestId, values.request_id, values.target_type === "after_sales" ? values.target_id : ""),
    ticketId: firstNonEmpty(resultContext.ticketId, values.ticket_id, values.target_type === "service_ticket" ? values.target_id : ""),
    userId: firstNonEmpty(resultContext.userId, values.user_id)
  });
}

export function linkedResultGroup(operationKey) {
  const normalized = String(operationKey || "").trim();
  if (!normalized) {
    return "other";
  }
  if (normalized === "order-detail") {
    return "order";
  }
  if (normalized === "refund-transactions") {
    return "finance";
  }
  if (normalized === "support-tickets" || normalized === "support-ticket-detail") {
    return "support";
  }
  if (normalized === "dispatch-order-events") {
    return "dispatch";
  }
  if (normalized === "audit-logs") {
    return "audit";
  }
  if (normalized.startsWith("after-sales")) {
    return "after-sales";
  }
  return "other";
}

export function linkedResultGroupLabel(groupKey) {
  return LINKED_RESULT_GROUP_LABELS[groupKey] || LINKED_RESULT_GROUP_LABELS.other;
}

function focusKey(type, value) {
  const normalizedValue = String(value || "").trim();
  return normalizedValue ? `${type}:${normalizedValue}` : "";
}

export function linkedContextTokens(context) {
  return [
    { type: "order", value: String(context?.orderId || "").trim(), label: context?.orderId ? `订单 ${context.orderId}` : "" },
    { type: "after_sales", value: String(context?.requestId || "").trim(), label: context?.requestId ? `售后 ${context.requestId}` : "" },
    { type: "service_ticket", value: String(context?.ticketId || "").trim(), label: context?.ticketId ? `工单 ${context.ticketId}` : "" },
    { type: "user", value: String(context?.userId || "").trim(), label: context?.userId ? `用户 ${context.userId}` : "" }
  ].filter((token) => token.value && token.label).map((token) => ({ ...token, focusKey: focusKey(token.type, token.value) }));
}

export function linkedResultTokens(entry) {
  return linkedContextTokens(linkedResultContext(entry));
}

export function linkedWorkspaceContextId(context) {
  const normalized = normalizeLinkedWorkspaceContext(context);
  return ["orderId", "requestId", "ticketId", "userId"]
    .map((key) => `${key}:${normalized[key] || ""}`)
    .join("|");
}

export function linkedWorkspaceContextEquals(left, right) {
  return linkedWorkspaceContextId(left) === linkedWorkspaceContextId(right);
}

export function linkedWorkspacePrimaryFocusKey(context) {
  return linkedContextTokens(normalizeLinkedWorkspaceContext(context))[0]?.focusKey || "";
}

export function linkedWorkspaceContextCandidates(entries, activeContext) {
  const currentContext = normalizeLinkedWorkspaceContext(activeContext);
  const currentContextId = linkedWorkspaceContextId(currentContext);
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  const candidates = new Map();

  normalizedEntries.forEach((entry) => {
    const context = linkedResultContext(entry);
    const tokens = linkedContextTokens(context);
    if (tokens.length === 0) {
      return;
    }
    const id = linkedWorkspaceContextId(context);
    if (!id || id === currentContextId) {
      return;
    }
    if (!candidates.has(id)) {
      candidates.set(id, {
        id,
        context,
        tokens,
        sourceTitles: []
      });
    }
    const candidate = candidates.get(id);
    const sourceTitle = String(entry?.title || entry?.operationKey || "").trim();
    if (sourceTitle && !candidate.sourceTitles.includes(sourceTitle)) {
      candidate.sourceTitles.push(sourceTitle);
    }
  });

  return Array.from(candidates.values());
}

export function syncLinkedResultValues(operationKey, values, context) {
  const normalizedOperationKey = String(operationKey || "").trim();
  const nextValues = {
    ...safeObject(values)
  };
  const orderId = String(context?.orderId || "").trim();
  const requestId = String(context?.requestId || "").trim();
  const ticketId = String(context?.ticketId || "").trim();
  const userId = String(context?.userId || "").trim();

  switch (normalizedOperationKey) {
    case "order-detail":
    case "dispatch-order-events":
    case "refund-transactions":
      setIfPresent(nextValues, "order_id", orderId);
      break;
    case "support-tickets":
      setIfPresent(nextValues, "related_order_id", orderId);
      setIfPresent(nextValues, "user_id", userId);
      break;
    case "support-ticket-detail":
      setIfPresent(nextValues, "ticket_id", ticketId);
      break;
    case "after-sales-detail":
    case "after-sales-events":
    case "after-sales-evidence":
      setIfPresent(nextValues, "request_id", requestId);
      break;
    case "after-sales-list":
      setIfPresent(nextValues, "order_id", orderId);
      setIfPresent(nextValues, "request_id", requestId);
      break;
    default:
      break;
  }

  if (hasText(nextValues.order_id)) {
    setIfPresent(nextValues, "order_id", orderId);
  }
  if (hasText(nextValues.related_order_id)) {
    setIfPresent(nextValues, "related_order_id", orderId);
  }
  if (hasText(nextValues.request_id)) {
    setIfPresent(nextValues, "request_id", requestId);
  }
  if (hasText(nextValues.ticket_id)) {
    setIfPresent(nextValues, "ticket_id", ticketId);
  }
  if (hasText(nextValues.user_id)) {
    setIfPresent(nextValues, "user_id", userId);
  }
  if (nextValues.target_type === "order") {
    setIfPresent(nextValues, "target_id", orderId);
  }
  if (nextValues.target_type === "after_sales") {
    setIfPresent(nextValues, "target_id", requestId);
  }
  if (nextValues.target_type === "service_ticket") {
    setIfPresent(nextValues, "target_id", ticketId);
  }

  return nextValues;
}

export function linkedResultSyncAction(entry, context) {
  const operationKey = String(entry?.operationKey || "").trim();
  if (!operationKey) {
    return null;
  }
  return {
    label: entry?.title || getAdminOperation(operationKey)?.title || "联动结果",
    operationKey,
    values: syncLinkedResultValues(operationKey, entry?.values, context)
  };
}

export function linkedResultSyncState(entry, context) {
  const action = linkedResultSyncAction(entry, context);
  if (!action) {
    return {
      action: null,
      stale: false,
      targetTokens: []
    };
  }
  const nextId = linkedResultEntryId(action);
  const targetTokens = linkedResultTokens({
    operationKey: action.operationKey,
    values: action.values
  });
  return {
    action,
    stale: nextId !== String(entry?.id || ""),
    targetTokens
  };
}

export function linkedWorkspaceSyncGroups(entries, context, { filterKey = "all", focusKey = "" } = {}) {
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  const groups = new Map();

  normalizedEntries
    .filter((entry) => linkedResultMatchesFilter(entry, filterKey))
    .map((entry) => ({
      entry,
      syncState: linkedResultSyncState(entry, context)
    }))
    .filter(({ entry, syncState }) =>
      syncState.stale
      && syncState.action
      && (linkedResultMatchesFocus(entry, focusKey) || linkedTokensMatchFocus(syncState.targetTokens, focusKey))
    )
    .forEach(({ entry, syncState }) => {
      if (!syncState.action) {
        return;
      }
      const groupKey = linkedResultGroup(entry?.operationKey);
      if (!groups.has(groupKey)) {
        groups.set(groupKey, {
          key: groupKey,
          label: linkedResultGroupLabel(groupKey),
          count: 0,
          actions: []
        });
      }
      const group = groups.get(groupKey);
      group.count += 1;
      group.actions.push(syncState.action);
    });

  return Object.keys(LINKED_RESULT_GROUP_LABELS)
    .filter((groupKey) => groups.has(groupKey))
    .map((groupKey) => groups.get(groupKey));
}

export function linkedWorkspaceSyncEntries(entries, context, { filterKey = "all", focusKey = "" } = {}) {
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  return normalizedEntries
    .filter((entry) => linkedResultMatchesFilter(entry, filterKey))
    .map((entry) => ({
      entry,
      syncState: linkedResultSyncState(entry, context)
    }))
    .filter(({ entry, syncState }) =>
      syncState.stale
      && syncState.action
      && (linkedResultMatchesFocus(entry, focusKey) || linkedTokensMatchFocus(syncState.targetTokens, focusKey))
    );
}

export function linkedWorkspaceSyncActions(entries, context, { filterKey = "all", focusKey = "", groupKey = "" } = {}) {
  if (hasText(groupKey)) {
    return linkedWorkspaceSyncGroups(entries, context, { filterKey, focusKey })
      .find((group) => group.key === String(groupKey).trim())?.actions || [];
  }
  return linkedWorkspaceSyncEntries(entries, context, { filterKey, focusKey })
    .map(({ syncState }) => syncState.action)
    .filter(Boolean);
}

export function linkedWorkspaceFailedEntries(entries, context, { filterKey = "all", focusKey = "", groupKey = "" } = {}) {
  const normalizedGroupKey = String(groupKey || "").trim();
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  return normalizedEntries
    .filter((entry) => linkedResultMatchesFilter(entry, filterKey))
    .filter((entry) => !normalizedGroupKey || linkedResultGroup(entry?.operationKey) === normalizedGroupKey)
    .filter((entry) => String(entry?.status || "").trim() === "error")
    .map((entry) => ({
      entry,
      syncState: linkedResultSyncState(entry, context)
    }))
    .filter(({ entry, syncState }) =>
      syncState.action
      && (linkedResultMatchesFocus(entry, focusKey) || linkedTokensMatchFocus(syncState.targetTokens, focusKey))
    );
}

export function linkedWorkspaceFailedActions(entries, context, { filterKey = "all", focusKey = "", groupKey = "" } = {}) {
  return dedupeActions(
    linkedWorkspaceFailedEntries(entries, context, { filterKey, focusKey, groupKey })
      .map(({ syncState }) => syncState.action)
      .filter(Boolean)
  );
}

export function linkedWorkspaceSyncOverview(entries, context, { filterKey = "all", focusKey = "" } = {}) {
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  const groups = new Map();

  normalizedEntries
    .filter((entry) => linkedResultMatchesFilter(entry, filterKey))
    .map((entry) => ({
      entry,
      syncState: linkedResultSyncState(entry, context)
    }))
    .filter(({ entry, syncState }) =>
      linkedResultMatchesFocus(entry, focusKey) || linkedTokensMatchFocus(syncState.targetTokens, focusKey)
    )
    .forEach(({ entry, syncState }) => {
      const groupKey = linkedResultGroup(entry?.operationKey);
      if (!groups.has(groupKey)) {
        groups.set(groupKey, {
          key: groupKey,
          label: linkedResultGroupLabel(groupKey),
          total: 0,
          stale: 0
        });
      }
      const group = groups.get(groupKey);
      group.total += 1;
      if (syncState.stale) {
        group.stale += 1;
      }
    });

  return Object.keys(LINKED_RESULT_GROUP_LABELS)
    .filter((groupKey) => groups.has(groupKey))
    .map((groupKey) => groups.get(groupKey));
}

export function linkedWorkspaceSyncActionGroups(actions) {
  const normalizedActions = Array.isArray(actions) ? actions.filter(Boolean) : [];
  const groups = new Map();

  normalizedActions.forEach((action) => {
    const groupKey = linkedResultGroup(action?.operationKey);
    if (!groups.has(groupKey)) {
      groups.set(groupKey, {
        key: groupKey,
        label: linkedResultGroupLabel(groupKey),
        count: 0
      });
    }
    groups.get(groupKey).count += 1;
  });

  return Object.keys(LINKED_RESULT_GROUP_LABELS)
    .filter((groupKey) => groups.has(groupKey))
    .map((groupKey) => groups.get(groupKey));
}

export function linkedWorkspaceFailedGroups(entries, context, { filterKey = "all", focusKey = "" } = {}) {
  return linkedWorkspaceSyncActionGroups(
    linkedWorkspaceFailedActions(entries, context, { filterKey, focusKey })
  );
}

function inferTargetTypeFromContext(context) {
  if (hasText(context?.requestId)) {
    return "after_sales";
  }
  if (hasText(context?.ticketId)) {
    return "service_ticket";
  }
  if (hasText(context?.orderId)) {
    return "order";
  }
  return "";
}

export function linkedOperationPrefillValues(operationKey, values, context) {
  const operation = getAdminOperation(operationKey);
  if (!operation) {
    return safeObject(values);
  }
  const nextValues = syncLinkedResultValues(operationKey, values, context);
  const fieldKeys = new Set(fieldsForOperation(operation).map((field) => field.key));
  const normalizedContext = normalizeLinkedWorkspaceContext(context);

  if (fieldKeys.has("order_id")) {
    setIfPresent(nextValues, "order_id", normalizedContext.orderId);
  }
  if (fieldKeys.has("related_order_id")) {
    setIfPresent(nextValues, "related_order_id", normalizedContext.orderId);
  }
  if (fieldKeys.has("request_id")) {
    setIfPresent(nextValues, "request_id", normalizedContext.requestId);
  }
  if (fieldKeys.has("ticket_id")) {
    setIfPresent(nextValues, "ticket_id", normalizedContext.ticketId);
  }
  if (fieldKeys.has("user_id")) {
    setIfPresent(nextValues, "user_id", normalizedContext.userId);
  }

  if (fieldKeys.has("target_type")) {
    const inferredTargetType = inferTargetTypeFromContext(normalizedContext);
    if (!hasText(nextValues.target_type)) {
      setIfPresent(nextValues, "target_type", inferredTargetType);
    }
  }
  if (fieldKeys.has("target_id")) {
    const targetType = String(nextValues.target_type || "").trim();
    if (targetType === "order") {
      setIfPresent(nextValues, "target_id", normalizedContext.orderId);
    }
    if (targetType === "after_sales") {
      setIfPresent(nextValues, "target_id", normalizedContext.requestId);
    }
    if (targetType === "service_ticket") {
      setIfPresent(nextValues, "target_id", normalizedContext.ticketId);
    }
  }

  return nextValues;
}

export function linkedResultsFilterOptions(entries) {
  const normalizedEntries = Array.isArray(entries) ? entries.filter(Boolean) : [];
  const counts = normalizedEntries.reduce((summary, entry) => {
    const group = linkedResultGroup(entry?.operationKey);
    summary.all += 1;
    if (String(entry?.status || "").trim() === "error") {
      summary.failed += 1;
    }
    summary[group] = (summary[group] || 0) + 1;
    return summary;
  }, { all: 0, failed: 0 });
  const optionKeys = ["all", "failed", ...Object.keys(LINKED_RESULT_GROUP_LABELS).filter((key) => key !== "all")];
  return optionKeys
    .filter((key) => key === "all" || (counts[key] || 0) > 0)
    .map((key) => ({
      key,
      label: key === "failed" ? "失败" : linkedResultGroupLabel(key),
      count: counts[key] || 0
    }));
}

export function linkedResultMatchesFilter(entry, filterKey) {
  const normalized = String(filterKey || "all").trim() || "all";
  if (normalized === "all") {
    return true;
  }
  if (normalized === "failed") {
    return String(entry?.status || "").trim() === "error";
  }
  return linkedResultGroup(entry?.operationKey) === normalized;
}

export function linkedResultMatchesFocus(entry, activeFocusKey) {
  const normalizedFocus = String(activeFocusKey || "").trim();
  if (!normalizedFocus) {
    return true;
  }
  return linkedResultTokens(entry).some((token) => token.focusKey === normalizedFocus);
}

function linkedTokensMatchFocus(tokens, activeFocusKey) {
  const normalizedFocus = String(activeFocusKey || "").trim();
  if (!normalizedFocus) {
    return true;
  }
  return (Array.isArray(tokens) ? tokens : []).some((token) => token?.focusKey === normalizedFocus);
}

function action(label, operationKey, values) {
  return { label, operationKey, values };
}

export function linkedWorkspacePrimaryActions(context) {
  const orderId = String(context?.orderId || "");
  const requestId = String(context?.requestId || "");
  const ticketId = String(context?.ticketId || "");
  const userId = String(context?.userId || "");
  return dedupeActions([
    orderId ? action("挂退款流水", "refund-transactions", { order_id: orderId, limit: 20 }) : null,
    orderId ? action("挂客服工单", "support-tickets", { related_order_id: orderId, user_id: userId, status: "", sla_status: "", assigned_support_id: "", limit: 20, now: "" }) : null,
    orderId ? action("挂派单事件", "dispatch-order-events", { order_id: orderId, station_manager_id: "" }) : null,
    orderId ? action("挂订单审计", "audit-logs", { target_type: "order", target_id: orderId, action: "", actor_type: "", actor_id: "", after: "", before: "", limit: 20 }) : null,
    orderId ? action("挂退款审计", "audit-logs", { target_type: "order", target_id: orderId, action: "admin.order.refunded", actor_type: "", actor_id: "", after: "", before: "", limit: 20 }) : null,
    orderId ? action("挂售后列表", "after-sales-list", { order_id: orderId, request_id: requestId, status: "" }) : null,
    requestId ? action("挂售后时间线", "after-sales-events", { request_id: requestId }) : null,
    requestId ? action("挂售后凭证", "after-sales-evidence", { request_id: requestId }) : null,
    requestId ? action("挂售后审计", "audit-logs", { target_type: "after_sales", target_id: requestId, action: "", actor_type: "", actor_id: "", after: "", before: "", limit: 20 }) : null,
    ticketId ? action("挂工单详情", "support-ticket-detail", { ticket_id: ticketId }) : null
  ].filter(Boolean));
}

export function linkedWorkspaceBundles(context) {
  const orderId = String(context?.orderId || "");
  const requestId = String(context?.requestId || "");
  const bundles = [];
  if (orderId) {
    bundles.push({
      key: "order-workspace",
      label: "打开订单工作区",
      actions: dedupeActions([
        action("退款流水", "refund-transactions", { order_id: orderId, limit: 20 }),
        action("客服工单", "support-tickets", { related_order_id: orderId, user_id: String(context?.userId || ""), status: "", sla_status: "", assigned_support_id: "", limit: 20, now: "" }),
        action("派单事件", "dispatch-order-events", { order_id: orderId, station_manager_id: "" }),
        action("订单审计", "audit-logs", { target_type: "order", target_id: orderId, action: "", actor_type: "", actor_id: "", after: "", before: "", limit: 20 })
      ])
    });
  }
  if (requestId) {
    bundles.push({
      key: "after-sales-workspace",
      label: "打开售后工作区",
      actions: dedupeActions([
        action("售后时间线", "after-sales-events", { request_id: requestId }),
        action("售后凭证", "after-sales-evidence", { request_id: requestId }),
        action("售后审计", "audit-logs", { target_type: "after_sales", target_id: requestId, action: "", actor_type: "", actor_id: "", after: "", before: "", limit: 20 }),
        ...(orderId ? [action("退款流水", "refund-transactions", { order_id: orderId, limit: 20 })] : [])
      ])
    });
  }
  return bundles.filter((bundle) => bundle.actions.length > 0);
}
