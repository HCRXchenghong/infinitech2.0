export const DEFAULT_BFF_BASE_URL = "http://localhost:25500";

export const ADMIN_API_OPERATIONS = Object.freeze([
  {
    key: "admin-login",
    title: "管理员登录",
    method: "POST",
    path: "/api/auth/admin/login",
    authRequired: false,
    area: "auth",
    fields: [
      { key: "account_id", label: "账号", type: "text", defaultValue: "admin_1", required: true },
      { key: "password", label: "密码", type: "password", required: true }
    ]
  },
  {
    key: "merchant-invite",
    title: "创建商户邀请",
    method: "POST",
    path: "/api/admin/merchant-invites",
    authRequired: true,
    area: "merchant",
    fields: [
      { key: "expires_in_hours", label: "有效小时", type: "number", defaultValue: 72 },
      { key: "note", label: "备注", type: "text", defaultValue: "merchant onboarding" }
    ]
  },
  {
    key: "station-manager-invite",
    title: "创建站长邀请",
    method: "POST",
    path: "/api/admin/rider-invites",
    authRequired: true,
    area: "rider",
    fields: [
      { key: "type", label: "类型", type: "select", defaultValue: "station_manager", options: ["station_manager"] },
      { key: "station_id", label: "站点", type: "text", defaultValue: "station_1", required: true }
    ]
  },
  {
    key: "rider-invite",
    title: "创建骑手邀请",
    method: "POST",
    path: "/api/admin/rider-invites",
    authRequired: true,
    area: "rider",
    fields: [
      { key: "type", label: "类型", type: "select", defaultValue: "rider", options: ["rider"] },
      { key: "station_id", label: "站点", type: "text", defaultValue: "station_1", required: true }
    ]
  },
  {
    key: "refund-settings-read",
    title: "读取退款策略",
    method: "GET",
    path: "/api/admin/refund-settings",
    authRequired: true,
    area: "finance",
    fields: []
  },
  {
    key: "refund-settings-save",
    title: "保存退款策略",
    method: "PUT",
    path: "/api/admin/refund-settings",
    authRequired: true,
    area: "finance",
    fields: [
      { key: "default_refund_strategy", label: "默认策略", type: "select", defaultValue: "balance_first", options: ["balance_first", "original_route_first"] }
    ]
  },
  {
    key: "after-sales-list",
    title: "售后审核列表",
    method: "GET",
    path: "/api/admin/after-sales",
    authRequired: true,
    area: "support",
    fields: []
  },
  {
    key: "operations-snapshot",
    title: "运营快照",
    method: "GET",
    path: "/api/admin/operations/snapshot",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "lease_expiring_within_seconds", label: "租约预警秒", type: "number", defaultValue: 60 },
      { key: "object_cleanup_grace_seconds", label: "对象清理宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "audit-logs",
    title: "操作审计",
    method: "GET",
    path: "/api/admin/audit-logs",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "actor_type", label: "操作者类型", type: "select", defaultValue: "", options: ["", "admin", "merchant", "station_manager", "rider"] },
      { key: "actor_id", label: "操作者 ID", type: "text", defaultValue: "" },
      { key: "target_type", label: "目标类型", type: "text", defaultValue: "" },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "" },
      { key: "action", label: "动作", type: "text", defaultValue: "" },
      { key: "after", label: "晚于时间", type: "text", defaultValue: "" },
      { key: "before", label: "早于时间", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "object-cleanup-stats",
    title: "对象清理统计",
    method: "GET",
    path: "/api/admin/object-storage/cleanup-stats",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "grace_seconds", label: "宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "object-cleanup-candidates",
    title: "对象清理候选",
    method: "GET",
    path: "/api/admin/object-storage/cleanup-candidates",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "grace_seconds", label: "宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "outbox-stats",
    title: "Outbox 健康",
    method: "GET",
    path: "/api/admin/outbox/stats",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid" },
      { key: "lease_expiring_within_seconds", label: "租约预警秒", type: "number", defaultValue: 60 }
    ],
    fields: []
  },
  {
    key: "outbox-events",
    title: "Outbox 事件",
    method: "GET",
    path: "/api/admin/outbox/events",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid" },
      { key: "status", label: "状态", type: "select", defaultValue: "pending", options: ["pending", "failed", "dead_letter", "published"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "outbox-replay-batch",
    title: "批量恢复 Outbox",
    method: "POST",
    path: "/api/admin/outbox/events/replay",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid", required: true },
      { key: "limit", label: "条数", type: "number", defaultValue: 10 }
    ]
  },
  {
    key: "order-compensate",
    title: "订单状态补偿",
    method: "POST",
    path: "/api/admin/orders/:order_id/state/compensate",
    authRequired: true,
    area: "order",
    pathFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "ord_1", required: true }
    ],
    fields: []
  },
  {
    key: "station-riders",
    title: "站点骑手列表",
    method: "GET",
    path: "/api/station-manager/riders",
    authRequired: true,
    area: "rider",
    fields: []
  },
  {
    key: "station-orders",
    title: "站点待调度订单",
    method: "GET",
    path: "/api/station-manager/orders",
    authRequired: true,
    area: "dispatch",
    fields: []
  },
  {
    key: "station-performance",
    title: "站点骑手绩效",
    method: "GET",
    path: "/api/station-manager/rider-performance",
    authRequired: true,
    area: "rider",
    fields: []
  },
  {
    key: "station-task-config",
    title: "站点任务配置",
    method: "GET",
    path: "/api/station-manager/task-duration",
    authRequired: true,
    area: "dispatch",
    fields: []
  }
]);

export function getAdminOperation(key) {
  return ADMIN_API_OPERATIONS.find((operation) => operation.key === key) || null;
}

export function fieldsForOperation(operation) {
  return [
    ...(operation.pathFields || []),
    ...(operation.queryFields || []),
    ...(operation.fields || [])
  ];
}

function normalizeValue(field, value) {
  if (field.type === "number") {
    const numberValue = Number(value);
    return Number.isFinite(numberValue) ? numberValue : undefined;
  }
  if (typeof value === "string") {
    return value.trim();
  }
  return value;
}

export function buildAdminRequest(operation, values = {}, token = "") {
  if (!operation) {
    throw new Error("unknown admin operation");
  }
  let path = operation.path;
  for (const field of operation.pathFields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (field.required && !value) {
      throw new Error(`${field.label}不能为空`);
    }
    path = path.replace(`:${field.key}`, encodeURIComponent(String(value)));
  }

  const query = new URLSearchParams();
  for (const field of operation.queryFields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (value !== undefined && value !== "") {
      query.set(field.key, String(value));
    }
  }
  const body = {};
  for (const field of operation.fields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (field.required && !value) {
      throw new Error(`${field.label}不能为空`);
    }
    if (value !== undefined && value !== "") {
      body[field.key] = value;
    }
  }

  const headers = { "Content-Type": "application/json", "X-Client-Kind": "admin-web" };
  if (operation.authRequired && token) {
    headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
  }

  const url = `${path}${query.toString() ? `?${query.toString()}` : ""}`;
  return {
    method: operation.method,
    url,
    headers,
    body: operation.method === "GET" ? undefined : JSON.stringify(body)
  };
}

export async function executeAdminOperation({ baseUrl = DEFAULT_BFF_BASE_URL, token = "", operationKey, values = {}, fetchImpl = fetch }) {
  const operation = getAdminOperation(operationKey);
  const request = buildAdminRequest(operation, values, token);
  const response = await fetchImpl(`${baseUrl.replace(/\/$/, "")}${request.url}`, request);
  const text = await response.text();
  let payload;
  try {
    payload = text ? JSON.parse(text) : null;
  } catch {
    payload = { raw: text };
  }
  return {
    ok: response.ok,
    status: response.status,
    operation,
    request: { method: request.method, url: request.url },
    payload
  };
}
