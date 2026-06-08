const HIGH_RISK_OPERATIONS = Object.freeze({
  "merchant-invite": "创建商户准入入口，可能影响平台开户主体",
  "merchant-qualification-review": "审核商户资质会改变商户接单资格和准入风控状态",
  "station-manager-invite": "创建站长准入入口，可能影响站点调度权限",
  "rider-invite": "创建骑手准入入口，可能影响履约队伍",
  "refund-settings-save": "修改资金退款默认策略，影响后续退款去向",
  "order-refund": "发起订单退款会改变资金账本和订单退款状态",
  "after-sales-review": "审核售后可能触发退款、驳回或升级平台仲裁",
  "support-ticket-assign": "分派客服工单会改变用户问题的处理责任人和 SLA 归属",
  "support-ticket-escalate": "升级客服工单会改变 SLA 归属并触发主管处理",
  "support-ticket-resolve": "提交客服处理方案会进入用户确认流程，影响用户侧工单状态",
  "support-quality-review": "客服质检结果会影响客服绩效和辅导记录",
  "audit-logs-export": "导出审计数据，需确认筛选范围和用途",
  "audit-retention-alert-emit": "投递审计告警，会进入通知和事故处理链路",
  "audit-archive-request": "请求 WORM 冷归档，会生成归档事件和审计证据",
  "audit-archive-verify": "触发归档对象回查，会读取归档对象并写入校验审计",
  "notification-delivery-record": "补录通知投递回执会影响运营对送达、失败和渠道质量的判断",
  "notification-failure-alert-emit": "投递通知失败告警，会进入 outbox 和后续事故处理链路",
  "notification-delivery-retry-schedule": "安排通知投递重试，会在 provider 退避后重新进入可靠投递链路",
  "notification-quiet-window-retry-schedule": "扫描静默窗口到期回执，会把 queued 通知重新放回可靠投递链路",
  "notification-preference-save": "保存通知偏好会改变外部触达渠道和静默窗口，可能影响商户接收关键通知",
  "notification-preference-batch-save": "批量保存通知偏好会同时改变多条外部触达策略，需确认变更范围和原因",
  "notification-preference-change-request": "提交通知偏好变更申请，会固化灰度范围并进入审批台账等待另一名管理员复核",
  "notification-preference-change-review": "审批通知偏好变更会影响后续是否允许应用到触达策略",
  "notification-preference-change-apply": "应用已审批通知偏好变更会按灰度范围修改外部触达策略并触发缓存失效",
  "rbac-change-request": "提交权限变更申请，可能改变后台权限治理流程",
  "rbac-review-request": "审批或驳回权限申请，影响高危权限变更状态",
  "rbac-apply-request": "应用已审批权限，会修改运行时 RBAC 策略",
  "rbac-rollback-request": "回滚已应用权限，会恢复运行时 RBAC 策略",
  "outbox-claim-events": "领取 outbox 租约，会影响 relay 副本对事件的可见性",
  "outbox-renew-lease": "续租 outbox 租约，错误 owner 可能造成重复投递或阻塞",
  "outbox-replay-event": "恢复单个 outbox 事件，可能触发一次可靠投递重试",
  "outbox-release-dead-letter": "解封 outbox dead-letter 事件，可能重新触发可靠投递",
  "outbox-mark-failed": "标记 outbox 事件失败，可能进入 backoff 或 dead-letter",
  "outbox-mark-published": "标记 outbox 事件已发布，会结束该事件的后续投递",
  "outbox-replay-batch": "批量恢复 outbox 事件，可能触发重复投递风险",
  "order-compensate": "补偿订单状态，可能影响订单履约和资金后续处理"
});

function compact(value, fallback = "-") {
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return String(value);
}

function operationTitle(operation) {
  return operation?.title || operation?.key || "未知操作";
}

export function operationRiskProfile(operation) {
  const reason = HIGH_RISK_OPERATIONS[operation?.key] || "";
  return {
    requiresConfirmation: Boolean(reason),
    severity: reason ? "high" : "normal",
    reason
  };
}

export function operationValuesSnapshot(operation, values = {}) {
  const fields = [
    ...(operation?.pathFields || []),
    ...(operation?.queryFields || []),
    ...(operation?.fields || [])
  ];
  return fields.reduce((snapshot, field) => {
    const value = values[field.key] ?? field.defaultValue ?? "";
    snapshot[field.key] = Array.isArray(value) ? [...value] : value;
    return snapshot;
  }, {});
}

export function buildPendingOperation(operation, values = {}, at = new Date().toISOString()) {
  const risk = operationRiskProfile(operation);
  return {
    id: `pending_${operation?.key || "operation"}_${Date.parse(at) || Date.now()}`,
    operationKey: operation?.key || "",
    title: operationTitle(operation),
    method: operation?.method || "",
    path: operation?.path || "",
    area: operation?.area || "",
    reason: risk.reason,
    values: operationValuesSnapshot(operation, values),
    createdAt: at
  };
}

export function buildOperationHistoryEntry(result, values = {}, at = new Date().toISOString()) {
  const operation = result?.operation || {};
  const ok = Boolean(result?.ok && result?.payload?.success !== false);
  const message = ok
    ? (result?.payload?.message || "执行成功")
    : (result?.payload?.message || result?.error || "执行失败");
  return {
    id: `op_${operation?.key || "unknown"}_${Date.parse(at) || Date.now()}`,
    operationKey: operation?.key || "",
    title: operationTitle(operation),
    ok,
    status: compact(result?.status, "error"),
    method: result?.request?.method || operation?.method || "",
    url: result?.request?.url || operation?.path || "",
    message,
    at,
    values: operationValuesSnapshot(operation, values)
  };
}

export function trimOperationHistory(history, limit = 12) {
  return (Array.isArray(history) ? history : []).slice(0, limit);
}

export function canReplayOperationHistoryEntry(entry) {
  return Boolean(entry?.operationKey && entry.ok === false);
}

export function operationHistoryReplayValues(entry) {
  const values = entry?.values && typeof entry.values === "object" ? entry.values : {};
  return Object.entries(values).reduce((snapshot, [key, value]) => {
    snapshot[key] = Array.isArray(value) ? [...value] : value;
    return snapshot;
  }, {});
}
