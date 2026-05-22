import { ADMIN_API_OPERATIONS, DEFAULT_BFF_BASE_URL, executeAdminOperation, fieldsForOperation, getAdminOperation } from "./adminApi.mjs";
import { ADMIN_WEB_KPIS, ADMIN_WEB_MODULES, ADMIN_WEB_QUEUES, ADMIN_WEB_RBAC, ADMIN_WEB_SECTIONS } from "./config.mjs";

const STORAGE_KEY = "infinitech.admin-web";
const root = document.getElementById("app");

const state = {
  activeModule: "dashboard",
  activeOperation: "refund-settings-read",
  baseUrl: DEFAULT_BFF_BASE_URL,
  token: "",
  lastResult: null,
  busy: false,
  values: {}
};

function restoreState() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
    state.baseUrl = saved.baseUrl || state.baseUrl;
    state.token = saved.token || state.token;
  } catch {
    localStorage.removeItem(STORAGE_KEY);
  }
}

function persistState() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ baseUrl: state.baseUrl, token: state.token }));
}

function formatJson(value) {
  if (!value) {
    return "{\n  \"status\": \"waiting\"\n}";
  }
  return JSON.stringify(value, null, 2);
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

function render() {
  if (!root) return;
  const activeOperation = getAdminOperation(state.activeOperation) || ADMIN_API_OPERATIONS[0];
  const activeModule = ADMIN_WEB_MODULES.find((module) => module.key === state.activeModule) || ADMIN_WEB_MODULES[0];
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
            <h1>${activeModule.title}</h1>
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

        <section class="kpis">
          ${ADMIN_WEB_KPIS.map((kpi) => `
            <article class="kpi ${kpi.tone}">
              <span>${kpi.title}</span>
              <strong>${kpi.value}</strong>
              <small>${kpi.trend}</small>
            </article>
          `).join("")}
        </section>

        <section class="grid">
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
                ${ADMIN_WEB_QUEUES.map((queue) => `
                  <tr>
                    <td>${queue.title}</td>
                    <td>${queue.target}</td>
                    <td><span class="pill">${queue.level}</span></td>
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
                  <strong>${role.name}</strong>
                  <span>${role.scopes.slice(0, 3).join(" / ")}</span>
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
                  <span>${module.title}</span>
                  <small>${statusLabel(module.status)} · ${module.owner}</small>
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
    persistState();
  });
  document.getElementById("admin-token")?.addEventListener("input", (event) => {
    state.token = event.target.value.trim();
    persistState();
  });
  document.querySelectorAll("[data-field]").forEach((field) => {
    field.addEventListener("input", () => {
      state.values[field.getAttribute("data-field")] = field.value;
    });
  });
  document.getElementById("fill-login")?.addEventListener("click", () => {
    const token = state.lastResult?.payload?.data?.access_token;
    if (token) {
      state.token = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
      persistState();
      render();
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

restoreState();
render();
