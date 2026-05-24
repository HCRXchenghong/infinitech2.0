import { ADMIN_API_OPERATIONS, DEFAULT_BFF_BASE_URL, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import {
  AUDIT_FILTER_DEFAULTS,
  auditDataFromResult,
  auditSearchValuesFromFilters,
  buildAuditRows,
  makeAuditFilterPreset,
  nextAuditBefore,
  normalizeAuditFilters,
  upsertAuditFilterPreset
} from "./adminAudit.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS } from "./config.mjs";
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
  snapshotBusy: false,
  auditBusy: false,
  auditError: "",
  auditLogs: [],
  auditNextBefore: "",
  auditFilters: { ...AUDIT_FILTER_DEFAULTS },
  auditFilterPresets: [],
  auditSelectedId: "",
  values: {}
};

function restoreState() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
    state.baseUrl = saved.baseUrl || state.baseUrl;
    state.token = saved.token || state.token;
    state.auditFilters = normalizeAuditFilters(saved.auditFilters || state.auditFilters);
    state.auditFilterPresets = Array.isArray(saved.auditFilterPresets) ? saved.auditFilterPresets.slice(0, 8) : [];
  } catch {
    localStorage.removeItem(STORAGE_KEY);
  }
}

function persistState() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({
    baseUrl: state.baseUrl,
    token: state.token,
    auditFilters: normalizeAuditFilters(state.auditFilters),
    auditFilterPresets: state.auditFilterPresets
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

function renderFields(operation) {
  const fields = fieldsForOperation(operation);
  if (fields.length === 0) {
    return `<div class="empty-state">此操作不需要额外参数。</div>`;
  }
  return fields.map((field) => {
    const value = state.values[field.key] ?? field.defaultValue ?? "";
    if (field.type === "select") {
      const options = (field.options || []).map((option) => `<option value="${option}" ${String(value) === option ? "selected" : ""}>${option}</option>`).join("");
      return `
        <label class="field">
          <span>${field.label}</span>
          <select data-field="${field.key}">${options}</select>
        </label>
      `;
    }
    return `
      <label class="field">
        <span>${field.label}</span>
        <input data-field="${field.key}" type="${field.type || "text"}" value="${String(value)}" ${field.required ? "required" : ""} />
      </label>
    `;
  }).join("");
}

function renderModuleView(view) {
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
      <div class="table-wrap">
        <table>
          <thead>
            <tr>${view.columns.map((column) => `<th>${escapeHtml(column)}</th>`).join("")}</tr>
          </thead>
          <tbody>
            ${view.rows.map((row) => `
              <tr>${row.map((cell) => `<td>${escapeHtml(cell)}</td>`).join("")}</tr>
            `).join("")}
          </tbody>
        </table>
      </div>
      <div class="safeguards">
        ${view.safeguards.map((item) => `<span>${escapeHtml(item)}</span>`).join("")}
      </div>
    </article>
  `;
}

function renderAuditCenter(view) {
  const rows = buildAuditRows(state.auditLogs);
  const filters = state.auditFilters;
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
  const activeModule = ADMIN_WEB_MODULES.find((module) => module.key === state.activeModule) || ADMIN_WEB_MODULES[0];
  const activeView = applySnapshotToAdminView(getAdminView(activeModule.key), state.snapshot);
  const kpis = buildSnapshotKpis(state.snapshot, ADMIN_WEB_KPIS);
  const queues = buildSnapshotQueues(state.snapshot, ADMIN_WEB_QUEUES);
  root.innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <div class="brand">
          <img src="../../assets/brand/logo.svg" alt="Infinitech" />
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
                <button type="submit" ${state.busy ? "disabled" : ""}>${state.busy ? "执行中" : "执行"}</button>
                <button type="button" id="fill-login">填入登录返回 Token</button>
              </div>
            </form>
            <pre class="result">${formatJson(state.lastResult)}</pre>
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
      render();
    });
  });
  document.querySelectorAll("[data-operation]").forEach((button) => {
    button.addEventListener("click", () => {
      state.activeOperation = button.getAttribute("data-operation");
      state.values = {};
      render();
    });
  });
  document.getElementById("operation-select")?.addEventListener("change", (event) => {
    state.activeOperation = event.target.value;
    state.values = {};
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
  state.busy = true;
  state.lastResult = { status: "running", operation: operation.title };
  render();
  try {
    state.lastResult = await executeAdminOperation({
      baseUrl: state.baseUrl || DEFAULT_BFF_BASE_URL,
      token: state.token,
      operationKey: operation.key,
      values: state.values
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
  } catch (error) {
    state.lastResult = {
      ok: false,
      error: error instanceof Error ? error.message : String(error),
      operation: operation.title
    };
  } finally {
    state.busy = false;
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
