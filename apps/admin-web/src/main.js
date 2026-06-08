import { ADMIN_API_OPERATIONS, DEFAULT_BFF_BASE_URL, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import {
  AUDIT_FILTER_DEFAULTS,
  auditDataFromResult,
  auditArchiveVerificationsFromResult,
  auditSearchValuesFromFilters,
  buildAuditArchiveVerificationRows,
  buildAuditRows,
  makeAuditFilterPreset,
  nextAuditBefore,
  normalizeAuditFilters,
  upsertAuditFilterPreset
} from "./adminAudit.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS } from "./config.mjs";
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
import {
  buildLinkedResultEntry,
  canInlinePreviewAction,
  linkedContextTokens,
  linkedOperationPrefillValues,
  linkedResultContext,
  linkedResultErrorMessage,
  linkedResultFailureFacts,
  linkedResultGroup,
  linkedResultGroupLabel,
  linkedResultMatchesFilter,
  linkedResultMatchesFocus,
  linkedResultAttemptTrail,
  linkedResultPrefillAction,
  linkedResultRetrySummary,
  linkedResultSyncState,
  linkedResultTokens,
  linkedResultsFilterOptions,
  linkedWorkspaceSyncActions,
  linkedWorkspaceSyncActionGroups,
  linkedWorkspaceSyncEntries,
  linkedWorkspaceFailedActions,
  linkedWorkspaceFailedGroups,
  linkedWorkspaceSyncGroups,
  linkedWorkspaceSyncOverview,
  linkedWorkspaceContextCandidates,
  linkedWorkspaceContextEquals,
  linkedWorkspacePrimaryFocusKey,
  linkedWorkspaceBundles,
  linkedWorkspaceContext,
  linkedWorkspacePrimaryActions,
  normalizeLinkedWorkspaceContext,
  upsertLinkedResult
} from "./adminLinkedWorkspace.mjs";
import { previewAdminResult } from "./adminResultPreview.mjs";
import { VIEW_PAGE_SIZE_OPTIONS, buildViewPage, normalizeViewFilter } from "./adminTable.mjs";
import { getAdminView } from "./adminViews.mjs";
import { applySnapshotToAdminView, buildSnapshotKpis, buildSnapshotQueues, snapshotDataFromResult } from "./adminSnapshot.mjs";

const STORAGE_KEY = "infinitech.admin-web";
const root = document.getElementById("app");

const state = {
  activeModule: "dashboard",
  activeOperation: "refund-settings-read",
  baseUrl: DEFAULT_BFF_BASE_URL,
  token: "",
  lastResult: null,
  snapshot: null,
  snapshotStatus: "idle",
  snapshotError: "",
  busy: false,
  linkedResults: [],
  linkedResultsFilter: "all",
  linkedResultsFocus: "",
  linkedWorkspaceContextOverride: null,
  linkedWorkspaceSyncActivity: {
    running: false,
    kind: "sync",
    mode: "",
    groupKey: "",
    requestedCount: 0,
    updatedCount: 0,
    failedCount: 0,
    groupCounts: {},
    failedGroupCounts: {},
    message: "",
    finishedAt: ""
  },
  snapshotBusy: false,
  auditBusy: false,
  auditError: "",
  auditLogs: [],
  auditNextBefore: "",
  auditFilters: { ...AUDIT_FILTER_DEFAULTS },
  auditFilterPresets: [],
  auditSelectedId: "",
  archiveVerificationBusy: false,
  archiveVerificationError: "",
  archiveVerificationFilters: { archive_id: "", status: "", limit: 20 },
  archiveVerifications: [],
  businessDetail: null,
  viewFilters: {},
  pendingOperation: null,
  operationHistory: [],
  values: {}
};

function restoreState() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
    state.baseUrl = saved.baseUrl || state.baseUrl;
    state.token = saved.token || state.token;
    state.auditFilters = normalizeAuditFilters(saved.auditFilters || state.auditFilters);
    state.auditFilterPresets = Array.isArray(saved.auditFilterPresets) ? saved.auditFilterPresets.slice(0, 8) : [];
    state.archiveVerificationFilters = {
      ...state.archiveVerificationFilters,
      ...(saved.archiveVerificationFilters || {})
    };
    state.viewFilters = Object.entries(saved.viewFilters || {}).reduce((filters, [key, value]) => {
      filters[key] = normalizeViewFilter(value);
      return filters;
    }, {});
    state.operationHistory = trimOperationHistory(saved.operationHistory || []);
  } catch {
    localStorage.removeItem(STORAGE_KEY);
  }
}

function persistState() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({
    baseUrl: state.baseUrl,
    token: state.token,
    auditFilters: normalizeAuditFilters(state.auditFilters),
    auditFilterPresets: state.auditFilterPresets,
    archiveVerificationFilters: state.archiveVerificationFilters,
    viewFilters: state.viewFilters,
    operationHistory: trimOperationHistory(state.operationHistory)
  }));
}

function formatJson(value) {
  if (!value) {
    return "{\n  \"status\": \"waiting\"\n}";
  }
  return JSON.stringify(value, null, 2);
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function safeHref(value) {
  const normalized = String(value ?? "").trim();
  return /^https?:\/\//i.test(normalized) ? normalized : "";
}

function applyPreviewAction(action) {
  if (!action?.operationKey) {
    return;
  }
  state.activeOperation = action.operationKey;
  state.values = { ...(action.values || {}) };
  state.pendingOperation = null;
  render();
}

function formatTraceTime(value) {
  const normalized = String(value || "").trim();
  if (!normalized) {
    return "";
  }
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) {
    return normalized;
  }
  return date.toLocaleString("zh-CN", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  });
}

function linkedResultTrailText(trail) {
  if (!trail) {
    return "";
  }
  return [
    trail.statusLabel || "",
    formatTraceTime(trail.at),
    trail.message || "",
    trail.paramsLabel || ""
  ].filter(Boolean).join(" · ");
}

function renderPreviewSection(result, scope) {
  const preview = previewAdminResult(result);
  if (!preview) {
    return "";
  }
  return `
    <section class="result-preview" aria-label="${escapeHtml(preview.title)}">
      <div class="result-preview-head">
        <div>
          <h3>${escapeHtml(preview.title)}</h3>
          <p>${escapeHtml(preview.subtitle)}</p>
        </div>
        <span class="badge">Preview</span>
      </div>
      <div class="result-preview-stats">
        ${preview.stats.map((stat) => `
          <div class="result-stat ${escapeHtml(stat.tone || "slate")}">
            <span>${escapeHtml(stat.label)}</span>
            <strong>${escapeHtml(stat.value)}</strong>
          </div>
        `).join("")}
      </div>
      <div class="result-preview-list">
        ${preview.items.length > 0 ? preview.items.map((item, itemIndex) => `
          <article class="result-item ${escapeHtml(item.tone || "slate")}">
            <div class="result-item-head">
              <div>
                <h4>${escapeHtml(item.title)}</h4>
                <div class="result-meta">
                  ${item.meta.map((meta) => `<span>${escapeHtml(meta)}</span>`).join("")}
                </div>
              </div>
              <span class="result-badge ${escapeHtml(item.tone || "slate")}">${escapeHtml(item.badge || "详情")}</span>
            </div>
            <p class="result-body">${escapeHtml(item.body || "")}</p>
            ${item.note ? `<p class="result-note">${escapeHtml(item.note)}</p>` : ""}
            ${item.previewImageUrl && safeHref(item.previewImageUrl) ? `
              <a class="result-image-link" href="${escapeHtml(safeHref(item.previewImageUrl))}" target="_blank" rel="noreferrer">
                <img class="result-image" src="${escapeHtml(safeHref(item.previewImageUrl))}" alt="${escapeHtml(item.title)}" loading="lazy" />
              </a>
            ` : ""}
            ${item.chips.length > 0 ? `
              <div class="result-chips">
                ${item.chips.map((chip) => `<span class="result-chip">${escapeHtml(chip)}</span>`).join("")}
              </div>
            ` : ""}
            ${Array.isArray(item.actions) && item.actions.length > 0 ? `
              <div class="result-item-actions">
                ${item.actions.map((action, actionIndex) => `
                  <button
                    type="button"
                    class="link-button"
                    data-preview-action="${escapeHtml(`${scope}:${itemIndex}:${actionIndex}`)}"
                  >${escapeHtml(action.label || action.operationKey)}</button>
                `).join("")}
              </div>
            ` : ""}
            ${item.links.length > 0 ? `
              <div class="result-links">
                ${item.links.map((link) => {
                  const href = safeHref(link.href);
                  if (!href) {
                    return "";
                  }
                  return `<a class="result-link" href="${escapeHtml(href)}" target="_blank" rel="noreferrer">${escapeHtml(link.label)}</a>`;
                }).join("")}
              </div>
            ` : ""}
          </article>
        `).join("") : `<div class="empty-state">${escapeHtml(preview.emptyMessage || "暂无可预览内容。")}</div>`}
      </div>
    </section>
  `;
}

function renderResultPreview(result) {
  return renderPreviewSection(result, "main");
}

function linkedWorkspaceBaseContext() {
  return normalizeLinkedWorkspaceContext(linkedWorkspaceContext(state.lastResult));
}

function linkedWorkspaceSyncMode({ groupKey = "", filterKey = "all", focusKey = "", origin = "manual" } = {}) {
  if (origin === "auto") {
    return "auto";
  }
  if (String(groupKey || "").trim()) {
    return "group";
  }
  if ((String(filterKey || "all").trim() || "all") === "all" && !String(focusKey || "").trim()) {
    return "all";
  }
  return "visible";
}

function linkedWorkspaceSyncScopeLabel({ mode = "", groupKey = "" } = {}) {
  if (mode === "group") {
    return linkedResultGroupLabel(groupKey);
  }
  if (mode === "all") {
    return "全部";
  }
  return "当前筛选";
}

function linkedWorkspaceSyncMessage(activity) {
  if (activity?.message) {
    return activity.message;
  }
  if (!activity?.running) {
    return "";
  }
  const scopeLabel = linkedWorkspaceSyncScopeLabel(activity);
  if (activity.kind === "retry") {
    return `正在重试${scopeLabel}失败项 ${activity.requestedCount} 张。`;
  }
  if (activity.mode === "auto") {
    return `正在自动刷新${scopeLabel} ${activity.requestedCount} 张卡。`;
  }
  return `正在刷新${scopeLabel} ${activity.requestedCount} 张卡。`;
}

function linkedWorkspaceSyncFinishedText(activity) {
  const finishedAt = String(activity?.finishedAt || "").trim();
  const groupCounts = activity?.groupCounts || {};
  const failedGroupCounts = activity?.failedGroupCounts || {};
  const groupSummary = Object.entries(groupCounts)
    .filter(([, count]) => Number(count) > 0)
    .map(([groupKey, count]) => `${linkedResultGroupLabel(groupKey)} ${count}`)
    .join(" / ");
  const failedSummary = Object.entries(failedGroupCounts)
    .filter(([, count]) => Number(count) > 0)
    .map(([groupKey, count]) => `${linkedResultGroupLabel(groupKey)} ${count}`)
    .join(" / ");
  if (!finishedAt && !groupSummary) {
    return "";
  }
  const timeText = finishedAt
    ? new Date(finishedAt).toLocaleTimeString("zh-CN", { hour12: false })
    : "";
  const segments = [timeText, groupSummary ? `已刷新 ${groupSummary}` : "", failedSummary ? `失败 ${failedSummary}` : ""]
    .filter(Boolean);
  if (segments.length > 0) {
    return segments.join(" · ");
  }
  return "";
}

function effectiveLinkedWorkspaceContext() {
  const baseContext = linkedWorkspaceBaseContext();
  const overrideContext = state.linkedWorkspaceContextOverride
    ? normalizeLinkedWorkspaceContext(state.linkedWorkspaceContextOverride)
    : null;
  if (!overrideContext) {
    return baseContext;
  }
  const hasOverrideTokens = linkedContextTokens(overrideContext).length > 0;
  return hasOverrideTokens ? overrideContext : baseContext;
}

function setLinkedWorkspaceContext(nextContext, { syncFocus = true } = {}) {
  const baseContext = linkedWorkspaceBaseContext();
  const normalizedContext = normalizeLinkedWorkspaceContext(nextContext);
  if (linkedWorkspaceContextEquals(baseContext, normalizedContext) || linkedContextTokens(normalizedContext).length === 0) {
    state.linkedWorkspaceContextOverride = null;
    if (syncFocus) {
      state.linkedResultsFocus = linkedWorkspacePrimaryFocusKey(baseContext);
    }
    state.values = linkedOperationPrefillValues(state.activeOperation, state.values, effectiveLinkedWorkspaceContext());
    return;
  }
  state.linkedWorkspaceContextOverride = normalizedContext;
  if (syncFocus) {
    state.linkedResultsFocus = linkedWorkspacePrimaryFocusKey(normalizedContext);
  }
  state.values = linkedOperationPrefillValues(state.activeOperation, state.values, effectiveLinkedWorkspaceContext());
}

async function setLinkedWorkspaceContextAndSync(nextContext, { syncFocus = true, origin = "auto" } = {}) {
  setLinkedWorkspaceContext(nextContext, { syncFocus });
  render();
  await runLinkedWorkspaceSync({
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus,
    origin
  });
}

function renderLinkedResults() {
  if (state.linkedResults.length === 0) {
    return "";
  }
  const filterOptions = linkedResultsFilterOptions(state.linkedResults);
  const workspaceContext = effectiveLinkedWorkspaceContext();
  const visibleEntries = state.linkedResults.filter((entry) =>
    linkedResultMatchesFilter(entry, state.linkedResultsFilter)
  );
  return `
    <section class="linked-results" aria-label="联动结果区">
      <div class="linked-results-head">
        <div>
          <h3>联动结果区</h3>
          <p>把退款、工单、审计这类只读结果先挂在旁边，排查上下文不会丢。</p>
        </div>
        <div class="linked-results-head-actions">
          ${filterOptions.map((option) => `
            <button
              type="button"
              class="link-button ${state.linkedResultsFilter === option.key ? "active" : ""}"
              data-linked-results-filter="${option.key}"
            >${escapeHtml(`${option.label} ${option.count}`)}</button>
          `).join("")}
          <button type="button" class="link-button" id="linked-results-clear">清空</button>
        </div>
      </div>
      <div class="linked-results-grid">
        ${visibleEntries.length > 0 ? visibleEntries.map((entry) => {
          const entryIndex = state.linkedResults.findIndex((candidate) => candidate.id === entry.id);
          const contextTokens = linkedResultTokens(entry);
          const focusMatched = linkedResultMatchesFocus(entry, state.linkedResultsFocus);
          const syncState = linkedResultSyncState(entry, workspaceContext);
          const errorMessage = entry.status === "error" ? linkedResultErrorMessage(entry) : "";
          const errorFacts = entry.status === "error" ? linkedResultFailureFacts(entry) : null;
          const retrySummary = entry.status === "error" ? linkedResultRetrySummary(entry, workspaceContext) : "";
          const attemptTrail = entry.status === "error" ? linkedResultAttemptTrail(entry) : [];
          const entryContext = linkedResultContext(entry);
          const canSwitchContext = !linkedWorkspaceContextEquals(entryContext, workspaceContext)
            && linkedContextTokens(entryContext).length > 0;
          const prefillAction = linkedResultPrefillAction(entry, workspaceContext);
          return `
          <article class="linked-result-card ${escapeHtml(entry.status || "ready")} ${state.linkedResultsFocus ? (focusMatched ? "focus-match" : "focus-dimmed") : ""}">
            <div class="linked-result-card-head">
              <div>
                <strong>${escapeHtml(entry.title || entry.result?.operation?.title || "联动结果")}</strong>
                <span>${escapeHtml(entry.requestUrl || entry.result?.request?.url || "")}</span>
              </div>
              <div class="linked-result-card-actions">
                ${canSwitchContext ? `<button type="button" class="inline-action" data-linked-context-switch="${entryIndex}">切到此上下文</button>` : ""}
                ${syncState.stale ? `<button type="button" class="inline-action" data-linked-sync="${entryIndex}">同步当前上下文</button>` : ""}
                ${entry.status === "error" && prefillAction ? `<button type="button" class="inline-action" data-linked-prefill-card="${entryIndex}">回填到操作台</button>` : ""}
                ${entry.status === "error" && syncState.action ? `<button type="button" class="inline-action" data-linked-retry-card="${entryIndex}">重试此卡</button>` : ""}
                <button type="button" class="inline-action" data-linked-promote="${entryIndex}">设为主结果</button>
                <button type="button" class="icon-button" data-linked-close="${entryIndex}" aria-label="关闭联动结果">×</button>
              </div>
            </div>
            <div class="linked-result-card-meta">
              ${entry.status === "error" ? `<span class="result-badge red">失败项</span>` : ""}
              <span class="result-badge slate">${escapeHtml(linkedResultGroupLabel(linkedResultGroup(entry.operationKey)))}</span>
              <span class="result-badge ${syncState.stale ? "amber" : "green"}">${escapeHtml(syncState.stale ? "待同步" : "已对齐")}</span>
              ${contextTokens.map((token) => `
                <button
                  type="button"
                  class="linked-context-chip ${state.linkedResultsFocus === token.focusKey ? "active" : ""}"
                  data-linked-workspace-focus="${token.focusKey}"
                >${escapeHtml(token.label)}</button>
              `).join("")}
            </div>
            ${syncState.stale && syncState.targetTokens.length > 0 ? `
              <p class="linked-sync-hint">当前主结果会把这张卡同步到：${escapeHtml(syncState.targetTokens.map((token) => token.label).join(" / "))}</p>
            ` : ""}
            ${entry.status === "error" ? `
              <div class="linked-result-error-panel">
                ${errorMessage ? `<p class="linked-result-error">失败原因：${escapeHtml(errorMessage)}</p>` : ""}
                ${errorFacts ? `
                  <div class="linked-result-error-facts">
                    <span class="result-badge red">${escapeHtml(errorFacts.statusLabel)}</span>
                    <span class="result-chip">${escapeHtml(`参数 ${errorFacts.paramsLabel}`)}</span>
                  </div>
                  <p class="linked-result-error-meta"><strong>请求：</strong>${escapeHtml(errorFacts.requestLabel)}</p>
                  <p class="linked-result-error-meta"><strong>重试：</strong>${escapeHtml(retrySummary || "沿用当前参数")}</p>
                  ${attemptTrail.length > 0 ? `
                    <div class="linked-result-attempts">
                      ${attemptTrail.map((trail) => `
                        <div class="linked-result-attempt ${escapeHtml(trail.key)}">
                          <div class="linked-result-attempt-head">
                            <span class="result-badge ${escapeHtml(trail.key === "failure" ? "red" : "green")}">${escapeHtml(trail.key === "failure" ? "最近失败" : "最近成功")}</span>
                            <small>${escapeHtml(formatTraceTime(trail.at) || "刚刚")}</small>
                          </div>
                          <p class="linked-result-error-meta">${escapeHtml(linkedResultTrailText(trail))}</p>
                        </div>
                      `).join("")}
                    </div>
                  ` : ""}
                ` : ""}
              </div>
            ` : ""}
            ${entry.status === "loading" ? `<div class="empty-state compact">正在读取 ${escapeHtml(entry.title || "联动结果")}...</div>` : ""}
            ${renderPreviewSection(entry.result, `linked-${entryIndex}`)}
            <details class="result-raw">
              <summary>查看原始返回</summary>
              <pre class="result embedded">${formatJson(entry.result)}</pre>
            </details>
          </article>
        `;
        }).join("") : `<div class="empty-state">当前筛选下还没有联动面板。</div>`}
      </div>
    </section>
  `;
}

function renderLinkedWorkspaceToolbar() {
  const baseContext = linkedWorkspaceBaseContext();
  const context = effectiveLinkedWorkspaceContext();
  const actions = linkedWorkspacePrimaryActions(context);
  const bundles = linkedWorkspaceBundles(context);
  const chips = linkedContextTokens(context);
  const baseChips = linkedContextTokens(baseContext);
  const activeOperation = getAdminOperation(state.activeOperation);
  const contextCandidates = linkedWorkspaceContextCandidates(state.linkedResults, context);
  const syncGroups = linkedWorkspaceSyncGroups(state.linkedResults, context, {
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus
  });
  const syncOverview = linkedWorkspaceSyncOverview(state.linkedResults, context, {
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus
  });
  const failedVisibleActions = linkedWorkspaceFailedActions(state.linkedResults, context, {
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus
  });
  const failedVisibleGroups = linkedWorkspaceFailedGroups(state.linkedResults, context, {
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus
  });
  const syncAllEntries = linkedWorkspaceSyncEntries(state.linkedResults, context, { filterKey: "all", focusKey: "" });
  const syncVisibleEntries = linkedWorkspaceSyncEntries(state.linkedResults, context, {
    filterKey: state.linkedResultsFilter,
    focusKey: state.linkedResultsFocus
  });
  const syncActivity = state.linkedWorkspaceSyncActivity;
  const syncStatusText = linkedWorkspaceSyncMessage(syncActivity);
  const syncFinishedText = linkedWorkspaceSyncFinishedText(syncActivity);
  const isSyncRunning = Boolean(syncActivity?.running);
  if (chips.length === 0 || (actions.length === 0 && bundles.length === 0)) {
    return "";
  }
  return `
    <section class="linked-workspace-toolbar" aria-label="联动工作区">
      <div class="linked-workspace-head">
        <div>
          <h3>同单工作区</h3>
          <p>围绕当前结果的订单/售后上下文，直接把常用排查面板挂起来。</p>
        </div>
        <div class="linked-workspace-head-actions">
          ${state.linkedResultsFocus ? `<button type="button" class="link-button" id="linked-workspace-focus-clear">取消定位</button>` : ""}
          ${state.linkedWorkspaceContextOverride ? `<button type="button" class="link-button" id="linked-workspace-follow-main">跟随主结果</button>` : ""}
        </div>
      </div>
      ${state.linkedWorkspaceContextOverride ? `
        <div class="linked-workspace-context-note">
          <span>当前工作区已切到联动面板上下文。</span>
          <small>主结果仍保持 ${escapeHtml(baseChips.map((chip) => chip.label).join(" / ") || "原上下文")}</small>
        </div>
      ` : ""}
      <div class="linked-workspace-context-note subtle">
        <span>操作台参数会跟随当前工作区上下文自动预填。</span>
        <small>当前操作 ${escapeHtml(activeOperation?.title || "未选择")}；切换上下文时会自动刷新当前筛选下待同步卡片。</small>
      </div>
      ${state.linkedResults.length > 0 ? `
        <div class="linked-workspace-sync">
          <span>已挂 ${state.linkedResults.length} 张面板，待同步 ${syncAllEntries.length} 张。</span>
          <div class="linked-workspace-sync-actions">
            <button
              type="button"
              class="link-button"
              id="linked-workspace-sync-visible"
              ${(syncVisibleEntries.length === 0 || isSyncRunning) ? "disabled" : ""}
            >${escapeHtml(syncActivity.running && (syncActivity.mode === "visible" || syncActivity.mode === "auto") ? `刷新中 ${syncActivity.requestedCount}` : `同步当前筛选 ${syncVisibleEntries.length}`)}</button>
            <button
              type="button"
              class="link-button"
              id="linked-workspace-sync-all"
              ${(syncAllEntries.length === 0 || isSyncRunning) ? "disabled" : ""}
            >${escapeHtml(syncActivity.running && syncActivity.mode === "all" ? `刷新中 ${syncActivity.requestedCount}` : `同步全部 ${syncAllEntries.length}`)}</button>
            ${failedVisibleActions.length > 0 ? `
              <button
                type="button"
                class="link-button linked-workspace-retry-group"
                id="linked-workspace-retry-visible"
                ${isSyncRunning ? "disabled" : ""}
              >${escapeHtml(syncActivity.running && syncActivity.kind === "retry" && syncActivity.mode !== "group" ? `重试中 ${syncActivity.requestedCount}` : `重试失败项 ${failedVisibleActions.length}`)}</button>
            ` : ""}
            ${failedVisibleActions.length > 0 && state.linkedResultsFilter !== "failed" ? `
              <button
                type="button"
                class="link-button"
                id="linked-workspace-filter-failed"
              >仅看失败项</button>
            ` : ""}
          </div>
        </div>
      ` : ""}
      ${syncStatusText ? `
        <div class="linked-workspace-sync-status">
          <span class="result-badge ${escapeHtml(syncActivity.running ? "blue" : "slate")}">${escapeHtml(syncActivity.running ? "同步中" : "同步反馈")}</span>
          <div>
            <span>${escapeHtml(syncStatusText)}</span>
            ${syncFinishedText ? `<small>${escapeHtml(syncFinishedText)}</small>` : ""}
          </div>
        </div>
      ` : ""}
      ${syncOverview.length > 0 ? `
        <div class="linked-workspace-sync-summary">
          ${syncOverview.map((group) => {
            const completedCount = Number(syncActivity.groupCounts?.[group.key] || 0);
            const running = syncActivity.running && (
              syncActivity.mode === "auto"
              || syncActivity.mode === "visible"
              || syncActivity.mode === "all"
              || (syncActivity.mode === "group" && syncActivity.groupKey === group.key)
            );
            const failedCount = Number(syncActivity.failedGroupCounts?.[group.key] || 0);
            const recentlyCompleted = !syncActivity.running && completedCount > 0 && Boolean(syncActivity.finishedAt);
            const tone = running
              ? "blue"
              : (failedCount > 0 ? "red" : (recentlyCompleted ? "slate" : (group.stale > 0 ? "amber" : "green")));
            const summary = running
              ? `刷新中 ${Math.max(group.stale, 1)}`
              : (failedCount > 0
                ? `失败 ${failedCount}`
                : (recentlyCompleted ? `刚刷新 ${completedCount}` : (group.stale > 0 ? `待同步 ${group.stale}/${group.total}` : `已对齐 ${group.total}`)));
            return `<span class="result-badge ${escapeHtml(tone)}">${escapeHtml(`${group.label} ${summary}`)}</span>`;
          }).join("")}
        </div>
      ` : ""}
      ${syncGroups.length > 0 ? `
        <div class="linked-workspace-sync-groups">
          ${syncGroups.map((group) => `
            <button
              type="button"
              class="link-button linked-workspace-sync-group"
              data-linked-workspace-sync-group="${group.key}"
              ${isSyncRunning ? "disabled" : ""}
            >${escapeHtml(syncActivity.running && syncActivity.mode === "group" && syncActivity.groupKey === group.key ? `刷新中 ${syncActivity.requestedCount}` : `同步${group.label} ${group.count}`)}</button>
          `).join("")}
        </div>
      ` : ""}
      ${failedVisibleGroups.length > 0 ? `
        <div class="linked-workspace-retry-groups">
          ${failedVisibleGroups.map((group) => `
            <button
              type="button"
              class="link-button linked-workspace-retry-group"
              data-linked-workspace-retry-group="${group.key}"
              ${isSyncRunning ? "disabled" : ""}
            >${escapeHtml(syncActivity.running && syncActivity.kind === "retry" && syncActivity.mode === "group" && syncActivity.groupKey === group.key ? `重试中 ${syncActivity.requestedCount}` : `重试${group.label}失败 ${group.count}`)}</button>
          `).join("")}
        </div>
      ` : ""}
      ${contextCandidates.length > 0 ? `
        <div class="linked-workspace-contexts">
          ${contextCandidates.map((candidate, candidateIndex) => `
            <button
              type="button"
              class="link-button linked-workspace-context-option"
              data-linked-workspace-context="${candidateIndex}"
            >
              <strong>${escapeHtml(candidate.tokens.map((token) => token.label).join(" / "))}</strong>
              <small>${escapeHtml(candidate.sourceTitles.join(" · "))}</small>
            </button>
          `).join("")}
        </div>
      ` : ""}
      <div class="linked-workspace-chips">
        ${chips.map((chip) => `
          <button
            type="button"
            class="linked-context-chip ${state.linkedResultsFocus === chip.focusKey ? "active" : ""}"
            data-linked-workspace-focus="${chip.focusKey}"
          >${escapeHtml(chip.label)}</button>
        `).join("")}
      </div>
      ${bundles.length > 0 ? `
        <div class="linked-workspace-bundles">
          ${bundles.map((bundle, bundleIndex) => `
            <button type="button" class="link-button" data-linked-workspace-bundle="${bundleIndex}">${escapeHtml(bundle.label)}</button>
          `).join("")}
        </div>
      ` : ""}
      <div class="linked-workspace-actions">
        ${actions.map((action, actionIndex) => `
          <button type="button" class="link-button" data-linked-workspace-action="${actionIndex}">${escapeHtml(action.label)}</button>
        `).join("")}
      </div>
    </section>
  `;
}

function renderOperationResult(result) {
  if (result?.status === "linked_prefilled") {
    return renderLinkedPrefillResult(result);
  }
  const preview = renderResultPreview(result);
  if (!preview) {
    return `<pre class="result">${formatJson(result)}</pre>`;
  }
  return `
    ${preview}
    <details class="result-raw">
      <summary>查看原始返回</summary>
      <pre class="result embedded">${formatJson(result)}</pre>
    </details>
  `;
}

function previewResultForScope(scope) {
  if (scope === "main") {
    return state.lastResult;
  }
  if (!scope?.startsWith("linked-")) {
    return null;
  }
  const linkedIndex = Number(scope.slice("linked-".length));
  return Number.isInteger(linkedIndex) ? state.linkedResults[linkedIndex]?.result || null : null;
}

function removeLinkedResultAt(index) {
  state.linkedResults = state.linkedResults.filter((_, entryIndex) => entryIndex !== index);
}

async function runLinkedPreviewAction(action, { pinOperation = true } = {}) {
  const operation = getAdminOperation(action?.operationKey);
  if (!action?.operationKey || !operation) {
    return;
  }
  if (pinOperation) {
    state.activeOperation = action.operationKey;
    state.values = { ...(action.values || {}) };
    state.pendingOperation = null;
  }
  if (!canInlinePreviewAction(action)) {
    render();
    return;
  }

  const loadingEntry = buildLinkedResultEntry(action, {
    status: "running",
    operation,
    request: { method: operation.method, url: "" },
    payload: { status: "loading" }
  }, "loading");
  state.linkedResults = upsertLinkedResult(state.linkedResults, loadingEntry);
  render();

  try {
    const result = await executeAdminOperation({
      baseUrl: state.baseUrl,
      token: state.token,
      operationKey: action.operationKey,
      values: action.values || {}
    });
    state.linkedResults = upsertLinkedResult(state.linkedResults, buildLinkedResultEntry(action, result, result.ok ? "ready" : "error"));
  } catch (error) {
    state.linkedResults = upsertLinkedResult(state.linkedResults, buildLinkedResultEntry(action, {
      ok: false,
      status: 0,
      operation,
      request: { method: operation.method, url: loadingEntry.requestUrl || "" },
      payload: { error: error instanceof Error ? error.message : String(error) }
    }, "error"));
  }
  render();
}

async function runLinkedPreviewActions(actions, { pinOperation = true } = {}) {
  const inlineActions = actions.filter((action) => canInlinePreviewAction(action));
  if (inlineActions.length === 0) {
    return [];
  }
  if (pinOperation) {
    const firstAction = inlineActions[0];
    state.activeOperation = firstAction.operationKey;
    state.values = { ...(firstAction.values || {}) };
    state.pendingOperation = null;
  }
  let nextEntries = state.linkedResults;
  for (const action of inlineActions.slice().reverse()) {
    nextEntries = upsertLinkedResult(nextEntries, buildLinkedResultEntry(action, {
      status: "running",
      operation: getAdminOperation(action.operationKey),
      request: { method: "GET", url: "" },
      payload: { status: "loading" }
    }, "loading"));
  }
  state.linkedResults = nextEntries;
  render();

  const resolved = await Promise.all(inlineActions.map(async (action) => {
    try {
      const result = await executeAdminOperation({
        baseUrl: state.baseUrl,
        token: state.token,
        operationKey: action.operationKey,
        values: action.values || {}
      });
      return { action, result, status: result.ok ? "ready" : "error" };
    } catch (error) {
      return {
        action,
        result: {
          ok: false,
          status: 0,
          operation: getAdminOperation(action.operationKey),
          request: { method: "GET", url: "" },
          payload: { error: error instanceof Error ? error.message : String(error) }
        },
        status: "error"
      };
    }
  }));

  let resolvedEntries = state.linkedResults;
  for (const entry of resolved.slice().reverse()) {
    resolvedEntries = upsertLinkedResult(resolvedEntries, buildLinkedResultEntry(entry.action, entry.result, entry.status));
  }
  state.linkedResults = resolvedEntries;
  render();
  return resolved;
}

async function runLinkedWorkspaceSync({ filterKey = "all", focusKey = "", groupKey = "", origin = "manual" } = {}) {
  const context = effectiveLinkedWorkspaceContext();
  const mode = linkedWorkspaceSyncMode({ filterKey, focusKey, groupKey, origin });
  const actions = linkedWorkspaceSyncActions(state.linkedResults, context, { filterKey, focusKey, groupKey });
  const actionGroups = linkedWorkspaceSyncActionGroups(actions);
  const groupCounts = Object.fromEntries(actionGroups.map((group) => [group.key, group.count]));
  if (actions.length === 0) {
    state.linkedWorkspaceSyncActivity = {
      running: false,
      kind: "sync",
      mode,
      groupKey,
      requestedCount: 0,
      updatedCount: 0,
      failedCount: 0,
      groupCounts: {},
      failedGroupCounts: {},
      message: `${linkedWorkspaceSyncScopeLabel({ mode, groupKey })} 已对齐，无需刷新。`,
      finishedAt: new Date().toISOString()
    };
    render();
    return;
  }
  state.linkedWorkspaceSyncActivity = {
    running: true,
    kind: "sync",
    mode,
    groupKey,
    requestedCount: actions.length,
    updatedCount: 0,
    failedCount: 0,
    groupCounts,
    failedGroupCounts: {},
    message: "",
    finishedAt: ""
  };
  render();
  const resolved = await runLinkedPreviewActions(actions, { pinOperation: false });
  const failedActions = (Array.isArray(resolved) ? resolved : [])
    .filter((entry) => entry?.status === "error")
    .map((entry) => entry.action)
    .filter(Boolean);
  const failedActionGroups = linkedWorkspaceSyncActionGroups(failedActions);
  const failedGroupCounts = Object.fromEntries(failedActionGroups.map((group) => [group.key, group.count]));
  const failedCount = failedActions.length;
  const updatedCount = actions.length - failedCount;
  state.linkedWorkspaceSyncActivity = {
    running: false,
    kind: "sync",
    mode,
    groupKey,
    requestedCount: actions.length,
    updatedCount,
    failedCount,
    groupCounts,
    failedGroupCounts,
    message: failedCount > 0
      ? `${origin === "auto" ? "已自动刷新" : "已刷新"}${linkedWorkspaceSyncScopeLabel({ mode, groupKey })} ${actions.length} 张卡，其中失败 ${failedCount} 张。`
      : `${origin === "auto" ? "已自动刷新" : "已刷新"}${linkedWorkspaceSyncScopeLabel({ mode, groupKey })} ${actions.length} 张卡。`,
    finishedAt: new Date().toISOString()
  };
  render();
}

async function runLinkedWorkspaceRetry({ filterKey = "all", focusKey = "", groupKey = "" } = {}) {
  const context = effectiveLinkedWorkspaceContext();
  const mode = linkedWorkspaceSyncMode({ filterKey, focusKey, groupKey, origin: "manual" });
  const actions = linkedWorkspaceFailedActions(state.linkedResults, context, { filterKey, focusKey, groupKey });
  const actionGroups = linkedWorkspaceSyncActionGroups(actions);
  const groupCounts = Object.fromEntries(actionGroups.map((group) => [group.key, group.count]));
  if (actions.length === 0) {
    state.linkedWorkspaceSyncActivity = {
      running: false,
      kind: "retry",
      mode,
      groupKey,
      requestedCount: 0,
      updatedCount: 0,
      failedCount: 0,
      groupCounts: {},
      failedGroupCounts: {},
      message: `${linkedWorkspaceSyncScopeLabel({ mode, groupKey })} 没有可重试的失败项。`,
      finishedAt: new Date().toISOString()
    };
    render();
    return;
  }
  state.linkedWorkspaceSyncActivity = {
    running: true,
    kind: "retry",
    mode,
    groupKey,
    requestedCount: actions.length,
    updatedCount: 0,
    failedCount: 0,
    groupCounts,
    failedGroupCounts: {},
    message: "",
    finishedAt: ""
  };
  render();
  const resolved = await runLinkedPreviewActions(actions, { pinOperation: false });
  const failedActions = (Array.isArray(resolved) ? resolved : [])
    .filter((entry) => entry?.status === "error")
    .map((entry) => entry.action)
    .filter(Boolean);
  const failedActionGroups = linkedWorkspaceSyncActionGroups(failedActions);
  const failedGroupCounts = Object.fromEntries(failedActionGroups.map((group) => [group.key, group.count]));
  const failedCount = failedActions.length;
  const updatedCount = actions.length - failedCount;
  state.linkedWorkspaceSyncActivity = {
    running: false,
    kind: "retry",
    mode,
    groupKey,
    requestedCount: actions.length,
    updatedCount,
    failedCount,
    groupCounts,
    failedGroupCounts,
    message: failedCount > 0
      ? `已重试${linkedWorkspaceSyncScopeLabel({ mode, groupKey })}失败项 ${actions.length} 张，仍失败 ${failedCount} 张。`
      : `已重试${linkedWorkspaceSyncScopeLabel({ mode, groupKey })}失败项 ${actions.length} 张。`,
    finishedAt: new Date().toISOString()
  };
  render();
}

async function runLinkedEntryRetry(entryIndex) {
  const entry = state.linkedResults[entryIndex];
  const action = linkedResultSyncState(entry, effectiveLinkedWorkspaceContext()).action;
  if (!entry || !action) {
    return;
  }
  const groupKey = linkedResultGroup(action.operationKey);
  state.linkedWorkspaceSyncActivity = {
    running: true,
    kind: "retry",
    mode: "group",
    groupKey,
    requestedCount: 1,
    updatedCount: 0,
    failedCount: 0,
    groupCounts: { [groupKey]: 1 },
    failedGroupCounts: {},
    message: `正在重试 ${entry.title || "当前卡片"}。`,
    finishedAt: ""
  };
  render();
  const resolved = await runLinkedPreviewActions([action], { pinOperation: false });
  const failedCount = (Array.isArray(resolved) ? resolved : []).filter((item) => item?.status === "error").length;
  state.linkedWorkspaceSyncActivity = {
    running: false,
    kind: "retry",
    mode: "group",
    groupKey,
    requestedCount: 1,
    updatedCount: failedCount > 0 ? 0 : 1,
    failedCount,
    groupCounts: { [groupKey]: failedCount > 0 ? 0 : 1 },
    failedGroupCounts: failedCount > 0 ? { [groupKey]: failedCount } : {},
    message: failedCount > 0 ? `已重试 ${entry.title || "当前卡片"}，仍失败。` : `已重试 ${entry.title || "当前卡片"}。`,
    finishedAt: new Date().toISOString()
  };
  render();
}

function snapshotStatusText() {
  if (state.snapshotBusy) return "正在刷新运营快照";
  if (state.snapshotStatus === "ready" && state.snapshot?.generated_at) return `已同步 ${new Date(state.snapshot.generated_at).toLocaleString("zh-CN")}`;
  if (state.snapshotStatus === "error") return state.snapshotError || "快照同步失败";
  if (!state.token) return "填入管理员 Token 后刷新快照";
  return "尚未同步运营快照";
}

function statusLabel(status) {
  const labels = {
    ready: "已成型",
    wired: "已接 API",
    planned: "待实装"
  };
  return labels[status] || status;
}

function operationOptions() {
  return ADMIN_API_OPERATIONS.map((operation) => `<option value="${operation.key}" ${operation.key === state.activeOperation ? "selected" : ""}>${operation.title}</option>`).join("");
}

function toneClass(tone) {
  return ["blue", "green", "red", "amber", "slate"].includes(tone) ? tone : "slate";
}

function selectedAuditRow(rows) {
  return rows.find((row) => row.id === state.auditSelectedId) || null;
}

function findAuditRow(rowId) {
  return buildAuditRows(state.auditLogs).find((row) => row.id === rowId) || null;
}

function valuesForAuditJump(row) {
  if (!row) {
    return {};
  }
  if (row.targetOperation === "order-compensate" && row.targetId) {
    return { order_id: row.targetId };
  }
  if (row.targetOperation === "outbox-events") {
    return { topic: row.payload?.topic || "", status: "pending", limit: 20 };
  }
  if (row.targetOperation === "outbox-stats") {
    return { topic: row.payload?.topic || "", lease_expiring_within_seconds: 60 };
  }
  if (row.targetOperation === "object-cleanup-candidates") {
    return { limit: 20, grace_seconds: 3600 };
  }
  return {};
}

function jumpFromAuditRow(row) {
  if (!row) {
    return;
  }
  state.activeModule = row.targetModule;
  state.activeOperation = row.targetOperation;
  state.values = valuesForAuditJump(row);
  state.auditSelectedId = "";
  render();
}

function currentActiveModule() {
  return ADMIN_WEB_MODULES.find((module) => module.key === state.activeModule) || ADMIN_WEB_MODULES[0];
}

function currentActiveView() {
  return applySnapshotToAdminView(getAdminView(currentActiveModule().key), state.snapshot);
}

function selectedBusinessDetail(view) {
  if (!state.businessDetail || state.businessDetail.moduleKey !== view.key) {
    return null;
  }
  return buildAdminBusinessDetail(view, state.businessDetail.rowIndex);
}

function currentViewFilter(viewKey) {
  return normalizeViewFilter(state.viewFilters[viewKey]);
}

function setViewFilter(viewKey, nextFilter) {
  if (!viewKey) {
    return;
  }
  state.viewFilters = {
    ...state.viewFilters,
    [viewKey]: normalizeViewFilter({
      ...currentViewFilter(viewKey),
      ...nextFilter
    })
  };
  persistState();
}

function renderBusinessDetail(detail) {
  if (!detail) {
    return "";
  }
  return `
    <aside class="business-detail" aria-label="业务详情">
      <div class="business-detail-head">
        <div>
          <span class="badge">${escapeHtml(detail.moduleKey)}</span>
          <h3>${escapeHtml(detail.title)}</h3>
          <p>${escapeHtml(detail.subtitle)}</p>
        </div>
        <button type="button" class="icon-button" id="business-detail-close" aria-label="关闭">×</button>
      </div>
      <div class="business-detail-grid">
        ${detail.facts.map((fact) => `
          <div>
            <span>${escapeHtml(fact.label)}</span>
            <strong>${escapeHtml(fact.value)}</strong>
          </div>
        `).join("")}
      </div>
      <div class="business-detail-actions">
        ${detail.actions.map((action, actionIndex) => {
          const operation = getAdminOperation(action.operationKey);
          return operation ? `<button type="button" class="link-button" data-detail-operation="${actionIndex}">${escapeHtml(action.label || operation.title)}</button>` : "";
        }).join("")}
      </div>
      <div class="business-detail-checklist">
        ${detail.checklist.map((item) => `<span>${escapeHtml(item)}</span>`).join("")}
      </div>
    </aside>
  `;
}

function renderFields(operation) {
  const fields = fieldsForOperation(operation);
  if (fields.length === 0) {
    return `<div class="empty-state">此操作不需要额外参数。</div>`;
  }
  return fields.map((field) => {
    const value = state.values[field.key] ?? field.defaultValue ?? "";
    if (field.type === "select" || field.type === "boolean") {
      const fieldOptions = field.type === "boolean" ? ["false", "true"] : (field.options || []);
      const options = fieldOptions.map((option) => `<option value="${escapeHtml(option)}" ${String(value) === option ? "selected" : ""}>${escapeHtml(option)}</option>`).join("");
      return `
        <label class="field">
          <span>${escapeHtml(field.label)}</span>
          <select data-field="${escapeHtml(field.key)}">${options}</select>
        </label>
      `;
    }
    const inputType = field.type === "csv" || field.type === "json" ? "text" : (field.type || "text");
    return `
      <label class="field">
        <span>${escapeHtml(field.label)}</span>
        <input data-field="${escapeHtml(field.key)}" type="${escapeHtml(inputType)}" value="${escapeHtml(value)}" ${field.required ? "required" : ""} />
      </label>
    `;
  }).join("");
}

function renderPendingOperation() {
  const pending = state.pendingOperation;
  if (!pending) {
    return "";
  }
  return `
    <section class="confirm-panel" aria-label="高风险操作确认">
      <div class="confirm-panel-head">
        <div>
          <span class="badge warn">需二次确认</span>
          <h3>${escapeHtml(pending.title)}</h3>
          <p>${escapeHtml(pending.reason || "该操作会改变关键运营状态。")}</p>
        </div>
      </div>
      <div class="confirm-grid">
        <div><span>方法</span><strong>${escapeHtml(pending.method)}</strong></div>
        <div><span>路径</span><strong>${escapeHtml(pending.path)}</strong></div>
        <div><span>区域</span><strong>${escapeHtml(pending.area)}</strong></div>
      </div>
      <pre class="confirm-values">${escapeHtml(JSON.stringify(pending.values, null, 2))}</pre>
      <div class="confirm-actions">
        <button type="button" id="confirm-operation" ${state.busy ? "disabled" : ""}>确认执行</button>
        <button type="button" id="cancel-operation" ${state.busy ? "disabled" : ""}>取消</button>
      </div>
    </section>
  `;
}

function renderLinkedPrefillResult(result) {
  const operationTitle = String(result?.operation || "当前查询");
  const sourceTitle = String(result?.source || "联动结果");
  const retrySummary = String(result?.retry_summary || "").trim();
  const failureFacts = result?.failure_facts || null;
  const attemptTrail = Array.isArray(result?.attempt_trail) ? result.attempt_trail : [];
  const values = result?.values && typeof result.values === "object" ? result.values : {};
  return `
    <section class="confirm-panel linked-prefill-panel" aria-label="联动卡回填确认">
      <div class="confirm-panel-head">
        <div>
          <span class="badge">已回填</span>
          <h3>${escapeHtml(operationTitle)}</h3>
          <p>参数已经从 ${escapeHtml(sourceTitle)} 带回操作台。可以直接执行当前查询，也可以先微调再跑。</p>
        </div>
      </div>
      <div class="confirm-grid linked-prefill-grid">
        <div><span>来源</span><strong>${escapeHtml(sourceTitle)}</strong></div>
        <div><span>操作</span><strong>${escapeHtml(operationTitle)}</strong></div>
        <div><span>当前参数</span><strong>${escapeHtml(retrySummary || "沿用当前参数")}</strong></div>
      </div>
      ${failureFacts ? `
        <div class="linked-prefill-facts">
          <span class="result-badge red">${escapeHtml(failureFacts.statusLabel)}</span>
          <span class="result-chip">${escapeHtml(`请求 ${failureFacts.requestLabel}`)}</span>
          <span class="result-chip">${escapeHtml(`原参数 ${failureFacts.paramsLabel}`)}</span>
        </div>
      ` : ""}
      ${attemptTrail.length > 0 ? `
        <div class="linked-prefill-attempts">
          ${attemptTrail.map((trail) => `
            <div class="linked-prefill-attempt ${escapeHtml(trail.key)}">
              <div class="linked-prefill-attempt-head">
                <span class="result-badge ${escapeHtml(trail.key === "failure" ? "red" : "green")}">${escapeHtml(trail.key === "failure" ? "最近失败" : "最近成功")}</span>
                <small>${escapeHtml(formatTraceTime(trail.at) || "刚刚")}</small>
              </div>
              <p>${escapeHtml(trail.message || trail.requestLabel || trail.statusLabel || "")}</p>
              <small>${escapeHtml(trail.paramsLabel || "无附加参数")}</small>
            </div>
          `).join("")}
        </div>
      ` : ""}
      <pre class="confirm-values">${escapeHtml(JSON.stringify(values, null, 2))}</pre>
      <div class="confirm-actions">
        <button type="button" id="linked-prefill-run" ${state.busy ? "disabled" : ""}>执行当前查询</button>
        <button type="button" id="linked-prefill-dismiss" ${state.busy ? "disabled" : ""}>收起提示</button>
      </div>
    </section>
  `;
}

function renderOperationHistory() {
  return `
    <section class="operation-history" aria-label="操作结果追踪">
      <div class="operation-history-head">
        <h3>操作结果追踪</h3>
        <span>${state.operationHistory.length} 条</span>
      </div>
      ${state.operationHistory.length > 0 ? `
        <div class="operation-history-list">
          ${state.operationHistory.map((entry, entryIndex) => `
            <div class="operation-history-row ${entry.ok ? "ok" : "failed"}">
              <div>
                <strong>${escapeHtml(entry.title)}</strong>
                <span>${escapeHtml(entry.method)} ${escapeHtml(entry.url)}</span>
              </div>
              <div>
                <span class="integrity-pill ${entry.ok ? "ok" : "danger"}">${entry.ok ? "成功" : "失败"}</span>
                <small>${escapeHtml(entry.status)} · ${escapeHtml(entry.at)}</small>
              </div>
              <p>${escapeHtml(entry.message)}</p>
              ${canReplayOperationHistoryEntry(entry) ? `
                <div class="operation-history-actions">
                  <button type="button" class="inline-action" data-replay-operation="${entryIndex}">重试</button>
                </div>
              ` : ""}
            </div>
          `).join("")}
        </div>
      ` : `<div class="empty-state">暂无操作结果。高风险动作确认后会在这里保留最近记录。</div>`}
    </section>
  `;
}

function renderModuleView(view) {
  const detail = selectedBusinessDetail(view);
  const page = buildViewPage(view, currentViewFilter(view.key));
  return `
    <article class="panel wide module-view">
      <div class="panel-head">
        <div>
          <h2>${escapeHtml(view.title)}</h2>
          <p>${escapeHtml(view.subtitle)}</p>
        </div>
        <span class="badge">${escapeHtml(view.key)}</span>
      </div>
      <div class="mini-metrics">
        ${view.metrics.map((metric) => `
          <div class="mini-metric ${toneClass(metric.tone)}">
            <span>${escapeHtml(metric.label)}</span>
            <strong>${escapeHtml(metric.value)}</strong>
          </div>
        `).join("")}
      </div>
      <div class="view-actions">
        ${view.actions.length > 0 ? view.actions.map((operationKey) => {
          const operation = getAdminOperation(operationKey);
          return operation ? `<button class="link-button" data-operation="${operation.key}">${operation.title}</button>` : "";
        }).join("") : `<span class="empty-state compact">暂无直连操作</span>`}
      </div>
      <form class="view-controls" data-view-filter-form="${escapeHtml(view.key)}">
        <label class="field">
          <span>筛选</span>
          <input data-view-filter-field="${escapeHtml(view.key)}" value="${escapeHtml(page.query)}" placeholder="${escapeHtml(view.columns.slice(0, 3).join(" / "))}" />
        </label>
        <label class="field">
          <span>每页</span>
          <select data-view-page-size="${escapeHtml(view.key)}">
            ${VIEW_PAGE_SIZE_OPTIONS.map((size) => `<option value="${size}" ${page.pageSize === size ? "selected" : ""}>${size}</option>`).join("")}
          </select>
        </label>
        <div class="view-filter-actions">
          <button type="submit" class="link-button">筛选</button>
          <button type="button" class="link-button" data-view-filter-clear="${escapeHtml(view.key)}">清除</button>
        </div>
        <div class="view-pagination">
          <button type="button" class="inline-action" data-view-page="${escapeHtml(view.key)}:prev" ${page.page <= 1 ? "disabled" : ""}>上一页</button>
          <span>第 ${page.page}/${page.totalPages} 页 · ${page.totalRows} 条</span>
          <button type="button" class="inline-action" data-view-page="${escapeHtml(view.key)}:next" ${page.page >= page.totalPages ? "disabled" : ""}>下一页</button>
        </div>
      </form>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>${view.columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}<th>动作</th></tr>
          </thead>
          <tbody>
            ${page.rows.length > 0 ? page.rows.map(({ row, rowIndex }) => `
                <tr class="${detail?.rowIndex === rowIndex ? "selected" : ""}">
                  ${row.map((cell) => `<td>${escapeHtml(cell)}</td>`).join("")}
                  <td><button type="button" class="inline-action" data-business-detail="${rowIndex}">详情</button></td>
                </tr>
              `).join("") : `
                <tr>
                  <td colspan="${view.columns.length + 1}">
                    <div class="empty-state">暂无匹配记录。</div>
                  </td>
                </tr>
              `}
          </tbody>
        </table>
      </div>
      ${renderBusinessDetail(detail)}
      <div class="safeguards">
        ${view.safeguards.map((item) => `<span>${escapeHtml(item)}</span>`).join("")}
      </div>
    </article>
  `;
}

function renderAuditCenter(view) {
  const rows = buildAuditRows(state.auditLogs);
  const archiveRows = buildAuditArchiveVerificationRows(state.archiveVerifications);
  const filters = state.auditFilters;
  const archiveFilters = state.archiveVerificationFilters;
  const selectedRow = selectedAuditRow(rows);
  return `
    <article class="panel wide audit-center">
      <div class="panel-head">
        <div>
          <h2>${escapeHtml(view.title)}</h2>
          <p>${escapeHtml(view.subtitle)}</p>
        </div>
        <span class="badge">P0</span>
      </div>
      <div class="mini-metrics">
        ${view.metrics.map((metric) => `
          <div class="mini-metric ${toneClass(metric.tone)}">
            <span>${escapeHtml(metric.label)}</span>
            <strong>${escapeHtml(metric.value)}</strong>
          </div>
        `).join("")}
      </div>
      <div class="audit-presets">
        <div class="audit-preset-list">
          ${state.auditFilterPresets.length > 0 ? state.auditFilterPresets.map((preset) => `
            <button type="button" class="link-button" data-audit-preset="${escapeHtml(preset.id)}">${escapeHtml(preset.name)}</button>
          `).join("") : `<span class="empty-state compact">暂无保存筛选</span>`}
        </div>
        <div class="audit-preset-actions">
          <button type="button" class="link-button" id="audit-save-filter">保存筛选</button>
          <button type="button" class="link-button" id="audit-reset-filter">重置</button>
        </div>
      </div>
      <form id="audit-search-form" class="audit-controls">
        <label class="field">
          <span>操作者类型</span>
          <select data-audit-field="actor_type">
            ${["", "admin", "merchant", "station_manager", "rider"].map((value) => `<option value="${value}" ${filters.actor_type === value ? "selected" : ""}>${value || "全部"}</option>`).join("")}
          </select>
        </label>
        <label class="field">
          <span>操作者 ID</span>
          <input data-audit-field="actor_id" value="${escapeHtml(filters.actor_id)}" placeholder="admin_1" />
        </label>
        <label class="field">
          <span>动作</span>
          <input data-audit-field="action" value="${escapeHtml(filters.action)}" placeholder="admin.order.refunded" />
        </label>
        <label class="field">
          <span>目标类型</span>
          <input data-audit-field="target_type" value="${escapeHtml(filters.target_type)}" placeholder="order" />
        </label>
        <label class="field">
          <span>目标 ID</span>
          <input data-audit-field="target_id" value="${escapeHtml(filters.target_id)}" placeholder="ord_1" />
        </label>
        <label class="field">
          <span>晚于时间</span>
          <input data-audit-field="after" value="${escapeHtml(filters.after)}" placeholder="2026-05-22T00:00:00Z" />
        </label>
        <label class="field">
          <span>早于时间</span>
          <input data-audit-field="before" value="${escapeHtml(filters.before)}" placeholder="2026-05-22T12:00:00Z" />
        </label>
        <label class="field">
          <span>条数</span>
          <input data-audit-field="limit" type="number" min="1" max="500" value="${escapeHtml(filters.limit)}" />
        </label>
        <div class="audit-actions">
          <button type="submit" ${state.auditBusy || !state.token ? "disabled" : ""}>${state.auditBusy ? "查询中" : "查询"}</button>
          <button type="button" id="audit-next-page" ${state.auditBusy || !state.token || !state.auditNextBefore ? "disabled" : ""}>下一页</button>
        </div>
      </form>
      ${state.auditError ? `<div class="empty-state audit-error">${escapeHtml(state.auditError)}</div>` : ""}
      <div class="table-wrap audit-table">
        <table>
          <thead>
            <tr>${view.columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr>
          </thead>
          <tbody>
            ${rows.length > 0 ? rows.map((row) => `
              <tr class="${row.id === state.auditSelectedId ? "selected" : ""}">
                <td>${escapeHtml(row.createdAt)}</td>
                <td>${escapeHtml(row.actor)}</td>
                <td>${escapeHtml(row.action)}</td>
                <td>
                  <span>${escapeHtml(row.target)}</span>
                  <button type="button" class="inline-action" data-audit-jump="${escapeHtml(row.id)}">${escapeHtml(row.targetLabel)}</button>
                </td>
                <td>${escapeHtml(row.request)}</td>
                <td>
                  <span>${escapeHtml(row.payloadSummary)}</span>
                  <button type="button" class="inline-action" data-audit-detail="${escapeHtml(row.id)}">详情</button>
                </td>
                <td>
                  <span class="integrity-pill ${escapeHtml(row.integrityTone)}">${escapeHtml(row.integrityLabel)}</span>
                  <small>${escapeHtml(row.integrityHashShort)}</small>
                </td>
              </tr>
            `).join("") : `
              <tr>
                <td colspan="${view.columns.length}">
                  <div class="empty-state">${state.token ? "暂无审计记录，调整筛选条件后重试。" : "填入管理员 Token 后可查询审计记录。"}</div>
                </td>
              </tr>
            `}
          </tbody>
        </table>
      </div>
      <section class="archive-verification-panel">
        <div class="archive-verification-head">
          <div>
            <h3>归档校验历史</h3>
            <p>回看每一次 WORM 归档对象校验结果、匹配项和异常码。</p>
          </div>
          <span class="badge">archive verify</span>
        </div>
        <form id="archive-verification-form" class="archive-verification-controls">
          <label class="field">
            <span>归档 ID</span>
            <input data-archive-verification-field="archive_id" value="${escapeHtml(archiveFilters.archive_id)}" placeholder="audit_archive_1" />
          </label>
          <label class="field">
            <span>状态</span>
            <select data-archive-verification-field="status">
              ${["", "verified", "failed"].map((value) => `<option value="${value}" ${archiveFilters.status === value ? "selected" : ""}>${value || "全部"}</option>`).join("")}
            </select>
          </label>
          <label class="field">
            <span>条数</span>
            <input data-archive-verification-field="limit" type="number" min="1" max="1000" value="${escapeHtml(archiveFilters.limit)}" />
          </label>
          <div class="audit-actions">
            <button type="submit" ${state.archiveVerificationBusy || !state.token ? "disabled" : ""}>${state.archiveVerificationBusy ? "查询中" : "查询历史"}</button>
          </div>
        </form>
        ${state.archiveVerificationError ? `<div class="empty-state audit-error">${escapeHtml(state.archiveVerificationError)}</div>` : ""}
        <div class="table-wrap archive-verification-table">
          <table>
            <thead>
              <tr>
                <th>归档</th>
                <th>状态</th>
                <th>校验时间</th>
                <th>Hash</th>
                <th>字节/条目</th>
                <th>详情</th>
              </tr>
            </thead>
            <tbody>
              ${archiveRows.length > 0 ? archiveRows.map((row) => `
                <tr>
                  <td>
                    <strong>${escapeHtml(row.archiveId || "-")}</strong>
                    <small>${escapeHtml(row.storageKey || "-")}</small>
                  </td>
                  <td><span class="integrity-pill ${escapeHtml(row.statusTone)}">${escapeHtml(row.status)}</span></td>
                  <td>${escapeHtml(row.verifiedAt || "-")}</td>
                  <td>
                    <span>manifest ${escapeHtml(row.manifestHashShort)}</span>
                    <small>content ${escapeHtml(row.actualContentHashShort)} / ${escapeHtml(row.expectedContentHashShort)}</small>
                  </td>
                  <td>
                    <span>bytes ${escapeHtml(row.bytesLabel)}</span>
                    <small>logs ${escapeHtml(row.logCountLabel)}</small>
                  </td>
                  <td>
                    <span>${escapeHtml(row.errorLabel || row.matchSummary)}</span>
                    <details>
                      <summary>详情</summary>
                      <pre class="audit-payload">${escapeHtml(JSON.stringify(row.raw, null, 2))}</pre>
                    </details>
                  </td>
                </tr>
              `).join("") : `
                <tr>
                  <td colspan="6">
                    <div class="empty-state">${state.token ? "暂无归档校验历史，输入归档 ID 后查询。" : "填入管理员 Token 后可查询归档校验历史。"}</div>
                  </td>
                </tr>
              `}
            </tbody>
          </table>
        </div>
      </section>
      <div class="safeguards">
        ${view.safeguards.map((item) => `<span>${escapeHtml(item)}</span>`).join("")}
      </div>
      ${selectedRow ? `
        <aside class="audit-detail" aria-label="审计详情">
          <div class="audit-detail-head">
            <div>
              <span class="badge">${escapeHtml(selectedRow.id)}</span>
              <h3>${escapeHtml(selectedRow.action)}</h3>
            </div>
            <button type="button" class="icon-button" id="audit-close-detail" aria-label="关闭">×</button>
          </div>
          <div class="audit-detail-grid">
            <div><span>时间</span><strong>${escapeHtml(selectedRow.createdRaw || selectedRow.createdAt)}</strong></div>
            <div><span>操作者</span><strong>${escapeHtml(selectedRow.actor)}</strong></div>
            <div><span>目标</span><strong>${escapeHtml(selectedRow.target)}</strong></div>
            <div><span>请求</span><strong>${escapeHtml(selectedRow.request)}</strong></div>
            <div><span>完整性</span><strong>${escapeHtml(selectedRow.integrityLabel)}</strong></div>
            <div><span>算法</span><strong>${escapeHtml(selectedRow.integrityAlgorithm)}</strong></div>
            <div><span>哈希</span><strong>${escapeHtml(selectedRow.integrityHash || "-")}</strong></div>
          </div>
          <div class="audit-detail-actions">
            <button type="button" class="link-button" data-audit-jump="${escapeHtml(selectedRow.id)}">跳到${escapeHtml(selectedRow.targetLabel)}</button>
            <button type="button" class="link-button" data-audit-target-filter="${escapeHtml(selectedRow.id)}">按此目标筛选</button>
          </div>
          <pre class="audit-payload detail">${escapeHtml(JSON.stringify(selectedRow.payload, null, 2))}</pre>
        </aside>
      ` : ""}
    </article>
  `;
}

function render() {
  if (!root) return;
  const activeOperation = getAdminOperation(state.activeOperation) || ADMIN_API_OPERATIONS[0];
  const activeModule = currentActiveModule();
  const activeView = currentActiveView();
  const kpis = buildSnapshotKpis(state.snapshot, ADMIN_WEB_KPIS);
  const queues = buildSnapshotQueues(state.snapshot, ADMIN_WEB_QUEUES);
  root.innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <div class="brand">
          <img src="./assets/brand/logo.svg" alt="Infinitech" />
          <div>
            <strong>Infinitech</strong>
            <span>运营后台</span>
          </div>
        </div>
        <nav class="nav">
          ${ADMIN_WEB_SECTIONS.map((section) => `
            <section>
              <h2>${section.title}</h2>
              ${section.modules.map((moduleKey) => {
                const module = ADMIN_WEB_MODULES.find((item) => item.key === moduleKey);
                if (!module) return "";
                return `<button class="nav-item ${state.activeModule === module.key ? "active" : ""}" data-module="${module.key}">
                  <span>${module.title}</span>
                  <small>${module.priority}</small>
                </button>`;
              }).join("")}
            </section>
          `).join("")}
        </nav>
      </aside>

      <main class="workspace">
        <header class="topbar">
          <div>
            <p class="eyebrow">桌面 Web 管理端</p>
            <h1>${escapeHtml(activeModule.title)}</h1>
          </div>
          <div class="connection">
            <label>
              <span>BFF</span>
              <input id="base-url" value="${state.baseUrl}" />
            </label>
            <label>
              <span>Token</span>
              <input id="admin-token" value="${state.token}" placeholder="Bearer ..." />
            </label>
          </div>
        </header>

        <section class="snapshot-strip ${state.snapshotStatus === "error" ? "error" : ""}">
          <div>
            <strong>运营快照</strong>
            <span>${escapeHtml(snapshotStatusText())}</span>
          </div>
          <button id="refresh-snapshot" ${state.snapshotBusy || !state.token ? "disabled" : ""}>${state.snapshotBusy ? "刷新中" : "刷新快照"}</button>
        </section>

        <section class="kpis">
          ${kpis.map((kpi) => `
            <article class="kpi ${kpi.tone}">
              <span>${escapeHtml(kpi.title)}</span>
              <strong>${escapeHtml(kpi.value)}</strong>
              <small>${escapeHtml(kpi.trend)}</small>
            </article>
          `).join("")}
        </section>

        <section class="grid">
          ${activeView.key === "audit-logs" ? renderAuditCenter(activeView) : renderModuleView(activeView)}

          <article class="panel wide">
            <div class="panel-head">
              <div>
                <h2>今日必须盯住</h2>
                <p>P0 队列：订单、资质、售后、资金、异步任务。</p>
              </div>
              <span class="badge">P0</span>
            </div>
            <table>
              <thead>
                <tr>
                  <th>队列</th>
                  <th>目标</th>
                  <th>级别</th>
                  <th>动作</th>
                </tr>
              </thead>
              <tbody>
                ${queues.map((queue) => `
                  <tr>
                    <td>${escapeHtml(queue.title)}</td>
                    <td>${escapeHtml(queue.target)}</td>
                    <td><span class="pill">${escapeHtml(queue.level)}</span></td>
                    <td><button class="link-button" data-operation="${queue.operationKey}">打开</button></td>
                  </tr>
                `).join("")}
              </tbody>
            </table>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <h2>权限边界</h2>
                <p>服务端策略：待细化。</p>
              </div>
            </div>
            <div class="role-list">
              ${ADMIN_WEB_RBAC.map((role) => `
                <div class="role-row">
                  <strong>${escapeHtml(role.name)}</strong>
                  <span>${escapeHtml(role.scopes.slice(0, 3).join(" / "))}</span>
                </div>
              `).join("")}
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <h2>模块状态</h2>
                <p>真实页面优先级。</p>
              </div>
            </div>
            <div class="module-list">
              ${ADMIN_WEB_MODULES.map((module) => `
                <button class="module-row ${state.activeModule === module.key ? "active" : ""}" data-module="${module.key}">
                  <span>${escapeHtml(module.title)}</span>
                  <small>${escapeHtml(`${statusLabel(module.status)} · ${module.owner}`)}</small>
                </button>
              `).join("")}
            </div>
          </article>

          <article class="panel operations">
            <div class="panel-head">
              <div>
                <h2>接口操作台</h2>
                <p>当前连接：${state.baseUrl}</p>
              </div>
              <span class="badge ${activeOperation.authRequired ? "warn" : ""}">${activeOperation.authRequired ? "需登录" : "公开"}</span>
            </div>
            <label class="field">
              <span>操作</span>
              <select id="operation-select">${operationOptions()}</select>
            </label>
            <form id="operation-form">
              ${renderFields(activeOperation)}
              <div class="actions">
                <button type="submit" ${state.busy ? "disabled" : ""}>${state.busy ? "执行中" : operationRiskProfile(activeOperation).requiresConfirmation ? "进入确认" : "执行"}</button>
                <button type="button" id="fill-login">填入登录返回 Token</button>
              </div>
            </form>
            ${renderPendingOperation()}
            ${renderOperationResult(state.lastResult)}
            ${renderLinkedWorkspaceToolbar()}
            ${renderLinkedResults()}
            ${renderOperationHistory()}
          </article>
        </section>
      </main>
    </div>
  `;
  bindEvents();
}

function bindEvents() {
  document.querySelectorAll("[data-module]").forEach((button) => {
    button.addEventListener("click", () => {
      state.activeModule = button.getAttribute("data-module");
      state.businessDetail = null;
      render();
    });
  });
  document.querySelectorAll("[data-operation]").forEach((button) => {
    button.addEventListener("click", () => {
      state.activeOperation = button.getAttribute("data-operation");
      state.values = linkedOperationPrefillValues(state.activeOperation, {}, effectiveLinkedWorkspaceContext());
      state.pendingOperation = null;
      render();
    });
  });
  document.getElementById("operation-select")?.addEventListener("change", (event) => {
    state.activeOperation = event.target.value;
    state.values = linkedOperationPrefillValues(state.activeOperation, {}, effectiveLinkedWorkspaceContext());
    state.pendingOperation = null;
    render();
  });
  document.getElementById("base-url")?.addEventListener("input", (event) => {
    state.baseUrl = event.target.value.trim();
    state.snapshotStatus = "idle";
    state.snapshotError = "";
    persistState();
  });
  document.getElementById("admin-token")?.addEventListener("input", (event) => {
    state.token = event.target.value.trim();
    state.snapshotStatus = "idle";
    state.snapshotError = "";
    persistState();
  });
  document.querySelectorAll("[data-field]").forEach((field) => {
    field.addEventListener("input", () => {
      state.values[field.getAttribute("data-field")] = field.value;
    });
  });
  document.getElementById("refresh-snapshot")?.addEventListener("click", async () => {
    await refreshOperationsSnapshot();
  });
  document.querySelectorAll("[data-audit-field]").forEach((field) => {
    field.addEventListener("input", () => {
      state.auditFilters[field.getAttribute("data-audit-field")] = field.value;
    });
  });
  document.getElementById("audit-search-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAuditSearch();
  });
  document.getElementById("audit-next-page")?.addEventListener("click", async () => {
    await runAuditSearch({ useNextPage: true });
  });
  document.querySelectorAll("[data-archive-verification-field]").forEach((field) => {
    field.addEventListener("input", () => {
      state.archiveVerificationFilters[field.getAttribute("data-archive-verification-field")] = field.value;
    });
  });
  document.getElementById("archive-verification-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runArchiveVerificationSearch();
  });
  document.getElementById("audit-save-filter")?.addEventListener("click", () => {
    const preset = makeAuditFilterPreset(state.auditFilters);
    state.auditFilterPresets = upsertAuditFilterPreset(state.auditFilterPresets, preset);
    state.auditFilters = preset.filters;
    persistState();
    render();
  });
  document.getElementById("audit-reset-filter")?.addEventListener("click", () => {
    state.auditFilters = { ...AUDIT_FILTER_DEFAULTS };
    state.auditNextBefore = "";
    state.auditSelectedId = "";
    persistState();
    render();
  });
  document.querySelectorAll("[data-audit-preset]").forEach((button) => {
    button.addEventListener("click", () => {
      const preset = state.auditFilterPresets.find((item) => item.id === button.getAttribute("data-audit-preset"));
      if (!preset) return;
      state.auditFilters = normalizeAuditFilters(preset.filters);
      state.auditNextBefore = "";
      state.auditSelectedId = "";
      persistState();
      render();
    });
  });
  document.querySelectorAll("[data-audit-detail]").forEach((button) => {
    button.addEventListener("click", () => {
      state.auditSelectedId = button.getAttribute("data-audit-detail") || "";
      render();
    });
  });
  document.getElementById("audit-close-detail")?.addEventListener("click", () => {
    state.auditSelectedId = "";
    render();
  });
  document.querySelectorAll("[data-audit-jump]").forEach((button) => {
    button.addEventListener("click", () => {
      jumpFromAuditRow(findAuditRow(button.getAttribute("data-audit-jump")));
    });
  });
  document.querySelectorAll("[data-audit-target-filter]").forEach((button) => {
    button.addEventListener("click", async () => {
      const row = findAuditRow(button.getAttribute("data-audit-target-filter"));
      if (!row) return;
      state.auditFilters = normalizeAuditFilters({
        ...state.auditFilters,
        target_type: row.targetType,
        target_id: row.targetId,
        before: ""
      });
      state.auditNextBefore = "";
      state.auditSelectedId = "";
      persistState();
      await runAuditSearch();
    });
  });
  document.querySelectorAll("[data-business-detail]").forEach((button) => {
    button.addEventListener("click", () => {
      const rowIndex = Number(button.getAttribute("data-business-detail"));
      const activeView = currentActiveView();
      state.businessDetail = { moduleKey: activeView.key, rowIndex };
      render();
    });
  });
  document.querySelectorAll("[data-view-filter-form]").forEach((form) => {
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      const viewKey = form.getAttribute("data-view-filter-form");
      const input = form.querySelector("[data-view-filter-field]");
      setViewFilter(viewKey, { query: input?.value || "", page: 1 });
      state.businessDetail = null;
      render();
    });
  });
  document.querySelectorAll("[data-view-page-size]").forEach((field) => {
    field.addEventListener("change", () => {
      setViewFilter(field.getAttribute("data-view-page-size"), { pageSize: field.value, page: 1 });
      state.businessDetail = null;
      render();
    });
  });
  document.querySelectorAll("[data-view-page]").forEach((button) => {
    button.addEventListener("click", () => {
      const [viewKey, direction] = String(button.getAttribute("data-view-page") || "").split(":");
      const filter = currentViewFilter(viewKey);
      setViewFilter(viewKey, { page: direction === "next" ? filter.page + 1 : filter.page - 1 });
      state.businessDetail = null;
      render();
    });
  });
  document.querySelectorAll("[data-view-filter-clear]").forEach((button) => {
    button.addEventListener("click", () => {
      setViewFilter(button.getAttribute("data-view-filter-clear"), { query: "", page: 1 });
      state.businessDetail = null;
      render();
    });
  });
  document.getElementById("business-detail-close")?.addEventListener("click", () => {
    state.businessDetail = null;
    render();
  });
  document.querySelectorAll("[data-detail-operation]").forEach((button) => {
    button.addEventListener("click", () => {
      const detail = selectedBusinessDetail(currentActiveView());
      const action = detail?.actions[Number(button.getAttribute("data-detail-operation"))];
      if (!action) {
        return;
      }
      applyPreviewAction(action);
    });
  });
  document.querySelectorAll("[data-preview-action]").forEach((button) => {
    button.addEventListener("click", async () => {
      const [scope, itemIndexText, actionIndexText] = String(button.getAttribute("data-preview-action") || "").split(":");
      const preview = previewAdminResult(previewResultForScope(scope));
      const itemIndex = Number(itemIndexText);
      const actionIndex = Number(actionIndexText);
      const action = preview?.items?.[itemIndex]?.actions?.[actionIndex];
      if (!action) {
        return;
      }
      if (canInlinePreviewAction(action)) {
        await runLinkedPreviewAction(action);
        return;
      }
      applyPreviewAction(action);
    });
  });
  document.querySelectorAll("[data-linked-workspace-action]").forEach((button) => {
    button.addEventListener("click", async () => {
      const actions = linkedWorkspacePrimaryActions(effectiveLinkedWorkspaceContext());
      const action = actions[Number(button.getAttribute("data-linked-workspace-action"))];
      if (!action) {
        return;
      }
      await runLinkedPreviewAction(action);
    });
  });
  document.querySelectorAll("[data-linked-workspace-bundle]").forEach((button) => {
    button.addEventListener("click", async () => {
      const bundles = linkedWorkspaceBundles(effectiveLinkedWorkspaceContext());
      const bundle = bundles[Number(button.getAttribute("data-linked-workspace-bundle"))];
      if (!bundle) {
        return;
      }
      await runLinkedPreviewActions(bundle.actions);
    });
  });
  document.getElementById("linked-workspace-sync-visible")?.addEventListener("click", async () => {
    await runLinkedWorkspaceSync({
      filterKey: state.linkedResultsFilter,
      focusKey: state.linkedResultsFocus
    });
  });
  document.getElementById("linked-workspace-sync-all")?.addEventListener("click", async () => {
    await runLinkedWorkspaceSync();
  });
  document.getElementById("linked-workspace-retry-visible")?.addEventListener("click", async () => {
    await runLinkedWorkspaceRetry({
      filterKey: state.linkedResultsFilter,
      focusKey: state.linkedResultsFocus
    });
  });
  document.getElementById("linked-workspace-filter-failed")?.addEventListener("click", () => {
    state.linkedResultsFilter = "failed";
    render();
  });
  document.querySelectorAll("[data-linked-workspace-sync-group]").forEach((button) => {
    button.addEventListener("click", async () => {
      await runLinkedWorkspaceSync({
        filterKey: state.linkedResultsFilter,
        focusKey: state.linkedResultsFocus,
        groupKey: String(button.getAttribute("data-linked-workspace-sync-group") || "")
      });
    });
  });
  document.querySelectorAll("[data-linked-workspace-retry-group]").forEach((button) => {
    button.addEventListener("click", async () => {
      await runLinkedWorkspaceRetry({
        filterKey: state.linkedResultsFilter,
        focusKey: state.linkedResultsFocus,
        groupKey: String(button.getAttribute("data-linked-workspace-retry-group") || "")
      });
    });
  });
  document.querySelectorAll("[data-linked-workspace-context]").forEach((button) => {
    button.addEventListener("click", async () => {
      const candidates = linkedWorkspaceContextCandidates(state.linkedResults, effectiveLinkedWorkspaceContext());
      const candidate = candidates[Number(button.getAttribute("data-linked-workspace-context"))];
      if (!candidate) {
        return;
      }
      await setLinkedWorkspaceContextAndSync(candidate.context);
    });
  });
  document.querySelectorAll("[data-linked-workspace-focus]").forEach((button) => {
    button.addEventListener("click", () => {
      const nextFocus = String(button.getAttribute("data-linked-workspace-focus") || "");
      state.linkedResultsFocus = state.linkedResultsFocus === nextFocus ? "" : nextFocus;
      render();
    });
  });
  document.getElementById("linked-workspace-follow-main")?.addEventListener("click", async () => {
    await setLinkedWorkspaceContextAndSync(linkedWorkspaceBaseContext());
  });
  document.getElementById("linked-workspace-focus-clear")?.addEventListener("click", () => {
    state.linkedResultsFocus = "";
    render();
  });
  document.querySelectorAll("[data-linked-results-filter]").forEach((button) => {
    button.addEventListener("click", () => {
      state.linkedResultsFilter = String(button.getAttribute("data-linked-results-filter") || "all") || "all";
      render();
    });
  });
  document.getElementById("linked-results-clear")?.addEventListener("click", () => {
      state.linkedResults = [];
      state.linkedResultsFilter = "all";
      state.linkedResultsFocus = "";
      state.linkedWorkspaceContextOverride = null;
      state.linkedWorkspaceSyncActivity = {
        running: false,
        kind: "sync",
        mode: "",
        groupKey: "",
        requestedCount: 0,
        updatedCount: 0,
        failedCount: 0,
        groupCounts: {},
        failedGroupCounts: {},
        message: "",
        finishedAt: ""
      };
      render();
  });
  document.querySelectorAll("[data-linked-close]").forEach((button) => {
    button.addEventListener("click", () => {
      removeLinkedResultAt(Number(button.getAttribute("data-linked-close")));
      render();
    });
  });
  document.querySelectorAll("[data-linked-context-switch]").forEach((button) => {
    button.addEventListener("click", async () => {
      const entryIndex = Number(button.getAttribute("data-linked-context-switch"));
      const entry = state.linkedResults[entryIndex];
      if (!entry) {
        return;
      }
      await setLinkedWorkspaceContextAndSync(linkedResultContext(entry));
    });
  });
  document.querySelectorAll("[data-linked-sync]").forEach((button) => {
    button.addEventListener("click", async () => {
      const entryIndex = Number(button.getAttribute("data-linked-sync"));
      const entry = state.linkedResults[entryIndex];
      const syncState = linkedResultSyncState(entry, effectiveLinkedWorkspaceContext());
      if (!syncState.stale || !syncState.action) {
        return;
      }
      await runLinkedPreviewAction(syncState.action, { pinOperation: false });
    });
  });
  document.querySelectorAll("[data-linked-retry-card]").forEach((button) => {
    button.addEventListener("click", async () => {
      await runLinkedEntryRetry(Number(button.getAttribute("data-linked-retry-card")));
    });
  });
  document.querySelectorAll("[data-linked-prefill-card]").forEach((button) => {
    button.addEventListener("click", () => {
      const entryIndex = Number(button.getAttribute("data-linked-prefill-card"));
      const entry = state.linkedResults[entryIndex];
      const workspaceContext = effectiveLinkedWorkspaceContext();
      const action = linkedResultPrefillAction(entry, workspaceContext);
      if (!action) {
        return;
      }
      applyPreviewAction(action);
      state.lastResult = {
        status: "linked_prefilled",
        operation: getAdminOperation(action.operationKey)?.title || action.operationKey,
        source: entry?.title || "联动结果",
        retry_summary: linkedResultRetrySummary(entry, workspaceContext),
        failure_facts: linkedResultFailureFacts(entry),
        attempt_trail: linkedResultAttemptTrail(entry),
        values: { ...(action.values || {}) }
      };
      render();
    });
  });
  document.getElementById("linked-prefill-run")?.addEventListener("click", async () => {
    await runActiveOperation();
  });
  document.getElementById("linked-prefill-dismiss")?.addEventListener("click", () => {
    state.lastResult = null;
    render();
  });
  document.querySelectorAll("[data-linked-promote]").forEach((button) => {
    button.addEventListener("click", async () => {
      const entryIndex = Number(button.getAttribute("data-linked-promote"));
      const entry = state.linkedResults[entryIndex];
      if (!entry?.result) {
        return;
      }
      state.lastResult = entry.result;
      state.activeOperation = entry.operationKey || state.activeOperation;
      state.values = linkedOperationPrefillValues(state.activeOperation, { ...(entry.values || {}) }, linkedWorkspaceBaseContext());
      state.linkedResultsFilter = "all";
      state.linkedResultsFocus = "";
      state.linkedWorkspaceContextOverride = null;
      removeLinkedResultAt(entryIndex);
      render();
      await runLinkedWorkspaceSync({ origin: "auto" });
    });
  });
  document.getElementById("confirm-operation")?.addEventListener("click", async () => {
    await runConfirmedOperation();
  });
  document.getElementById("cancel-operation")?.addEventListener("click", () => {
    state.pendingOperation = null;
    state.lastResult = { status: "cancelled", operation: "高风险操作确认" };
    render();
  });
  document.querySelectorAll("[data-replay-operation]").forEach((button) => {
    button.addEventListener("click", () => {
      const entry = state.operationHistory[Number(button.getAttribute("data-replay-operation"))];
      const operation = getAdminOperation(entry?.operationKey);
      if (!operation || !canReplayOperationHistoryEntry(entry)) {
        return;
      }
      const values = operationHistoryReplayValues(entry);
      state.activeOperation = operation.key;
      state.values = values;
      if (operationRiskProfile(operation).requiresConfirmation) {
        state.pendingOperation = buildPendingOperation(operation, values);
        state.lastResult = {
          status: "pending_confirmation",
          operation: operation.title,
          reason: state.pendingOperation.reason,
          replay_of: entry.id
        };
      } else {
        state.pendingOperation = null;
        state.lastResult = {
          status: "ready_to_replay",
          operation: operation.title,
          replay_of: entry.id
        };
      }
      render();
    });
  });
  document.getElementById("fill-login")?.addEventListener("click", async () => {
    const token = state.lastResult?.payload?.data?.access_token;
    if (token) {
      state.token = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
      persistState();
      render();
      await refreshOperationsSnapshot({ silent: true });
    }
  });
  document.getElementById("operation-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runActiveOperation();
  });
}

async function runActiveOperation() {
  const operation = getAdminOperation(state.activeOperation);
  if (!operation) return;
  const values = operationValuesSnapshot(operation, state.values);
  if (operationRiskProfile(operation).requiresConfirmation) {
    state.pendingOperation = buildPendingOperation(operation, values);
    state.lastResult = {
      status: "pending_confirmation",
      operation: operation.title,
      reason: state.pendingOperation.reason
    };
    render();
    return;
  }
  await executeOperationNow(operation, values);
}

async function runConfirmedOperation() {
  const pending = state.pendingOperation;
  const operation = getAdminOperation(pending?.operationKey);
  if (!pending || !operation) {
    return;
  }
  state.pendingOperation = null;
  await executeOperationNow(operation, pending.values);
}

async function executeOperationNow(operation, values) {
  state.busy = true;
  state.linkedResults = [];
  state.linkedResultsFilter = "all";
  state.linkedResultsFocus = "";
  state.linkedWorkspaceContextOverride = null;
  state.lastResult = { status: "running", operation: operation.title };
  render();
  try {
    state.lastResult = await executeAdminOperation({
      baseUrl: state.baseUrl || DEFAULT_BFF_BASE_URL,
      token: state.token,
      operationKey: operation.key,
      values
    });
    if (operation.key === "operations-snapshot" && state.lastResult.ok) {
      const snapshot = snapshotDataFromResult(state.lastResult);
      if (snapshot) {
        state.snapshot = snapshot;
        state.snapshotStatus = "ready";
        state.snapshotError = "";
      }
    }
    if (operation.key === "audit-logs" && state.lastResult.ok) {
      state.auditLogs = auditDataFromResult(state.lastResult);
      state.auditNextBefore = nextAuditBefore(state.auditLogs);
      state.auditError = "";
    }
    if (operation.key === "audit-archive-verifications" && state.lastResult.ok) {
      state.archiveVerifications = auditArchiveVerificationsFromResult(state.lastResult);
      state.archiveVerificationError = "";
    }
    state.operationHistory = trimOperationHistory([
      buildOperationHistoryEntry(state.lastResult, values),
      ...state.operationHistory
    ]);
  } catch (error) {
    state.lastResult = {
      ok: false,
      error: error instanceof Error ? error.message : String(error),
      operation,
      request: { method: operation.method, url: operation.path },
      status: "error"
    };
    state.operationHistory = trimOperationHistory([
      buildOperationHistoryEntry(state.lastResult, values),
      ...state.operationHistory
    ]);
  } finally {
    state.busy = false;
    persistState();
    render();
  }
}

async function runArchiveVerificationSearch() {
  if (!state.token || state.archiveVerificationBusy) {
    return;
  }
  state.archiveVerificationBusy = true;
  state.archiveVerificationError = "";
  state.linkedResults = [];
  state.linkedResultsFilter = "all";
  state.linkedResultsFocus = "";
  state.linkedWorkspaceContextOverride = null;
  render();
  try {
    const result = await executeAdminOperation({
      baseUrl: state.baseUrl || DEFAULT_BFF_BASE_URL,
      token: state.token,
      operationKey: "audit-archive-verifications",
      values: state.archiveVerificationFilters
    });
    if (!result.ok || result.payload?.success === false) {
      throw new Error(result.payload?.message || `HTTP ${result.status}`);
    }
    state.lastResult = result;
    state.archiveVerifications = auditArchiveVerificationsFromResult(result);
  } catch (error) {
    state.archiveVerificationError = error instanceof Error ? error.message : String(error);
  } finally {
    state.archiveVerificationBusy = false;
    persistState();
    render();
  }
}

function auditSearchValues({ useNextPage = false } = {}) {
  return auditSearchValuesFromFilters(state.auditFilters, {
    beforeOverride: useNextPage ? state.auditNextBefore : ""
  });
}

async function runAuditSearch({ useNextPage = false } = {}) {
  if (!state.token || state.auditBusy) {
    return;
  }
  state.auditBusy = true;
  state.auditError = "";
  state.linkedResults = [];
  state.linkedResultsFilter = "all";
  state.linkedResultsFocus = "";
  state.linkedWorkspaceContextOverride = null;
  render();
  try {
    const values = auditSearchValues({ useNextPage });
    const result = await executeAdminOperation({
      baseUrl: state.baseUrl || DEFAULT_BFF_BASE_URL,
      token: state.token,
      operationKey: "audit-logs",
      values
    });
    if (!result.ok || result.payload?.success === false) {
      throw new Error(result.payload?.message || `HTTP ${result.status}`);
    }
    state.lastResult = result;
    state.auditLogs = auditDataFromResult(result);
    state.auditNextBefore = nextAuditBefore(state.auditLogs);
    state.auditFilters = normalizeAuditFilters({ ...state.auditFilters, before: values.before || "" });
    state.auditSelectedId = "";
  } catch (error) {
    state.auditError = error instanceof Error ? error.message : String(error);
  } finally {
    state.auditBusy = false;
    persistState();
    render();
  }
}

async function refreshOperationsSnapshot({ silent = false } = {}) {
  if (!state.token || state.snapshotBusy) {
    return;
  }
  state.snapshotBusy = true;
  state.snapshotStatus = state.snapshot ? state.snapshotStatus : "loading";
  state.snapshotError = "";
  if (!silent) {
    render();
  }
  try {
    const result = await executeAdminOperation({
      baseUrl: state.baseUrl || DEFAULT_BFF_BASE_URL,
      token: state.token,
      operationKey: "operations-snapshot",
      values: { limit: 20, lease_expiring_within_seconds: 60, object_cleanup_grace_seconds: 3600 }
    });
    if (!result.ok || result.payload?.success === false) {
      throw new Error(result.payload?.message || `HTTP ${result.status}`);
    }
    const snapshot = snapshotDataFromResult(result);
    if (!snapshot) {
      throw new Error("快照响应缺少 data");
    }
    state.snapshot = snapshot;
    state.snapshotStatus = "ready";
    state.snapshotError = "";
  } catch (error) {
    state.snapshotStatus = "error";
    state.snapshotError = error instanceof Error ? error.message : String(error);
  } finally {
    state.snapshotBusy = false;
    persistState();
    render();
  }
}

restoreState();
render();
if (state.token) {
  void refreshOperationsSnapshot({ silent: true });
}
