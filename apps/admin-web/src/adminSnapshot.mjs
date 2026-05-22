const ORDER_TYPE_LABELS = Object.freeze({
  takeout: "外卖",
  groupbuy: "团购",
  medicine: "买药",
  courier: "快递",
  errand_buy: "跑腿代买",
  errand_deliver: "跑腿代送",
  errand_pickup: "跑腿代取",
  errand_do: "跑腿办事"
});

const ORDER_STATUS_LABELS = Object.freeze({
  pending_payment: "待支付",
  merchant_pending: "待商家接单",
  preparing: "备货中",
  dispatching: "待派单",
  rider_assigned: "骑手已接单",
  picked_up: "已取货",
  delivering: "配送中",
  voucher_issued: "已发券",
  cancelled: "已取消",
  completed: "已完成",
  refund_pending: "退款中",
  refunded: "已退款"
});

const AFTER_SALES_STATUS_LABELS = Object.freeze({
  pending_merchant: "商户待处理",
  admin_review: "平台审核",
  approved: "已通过",
  rejected: "已驳回",
  refunded: "已退款"
});

const REFUND_STRATEGY_LABELS = Object.freeze({
  balance_first: "余额优先",
  original_route_first: "原路优先"
});

function compact(value, fallback = "-") {
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return String(value);
}

function numberValue(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatCount(value) {
  return new Intl.NumberFormat("zh-CN").format(numberValue(value));
}

function formatFen(value) {
  return `${(numberValue(value) / 100).toFixed(2)} 元`;
}

function formatPercent(value) {
  const parsed = numberValue(value);
  if (parsed <= 0) {
    return "0%";
  }
  return `${(parsed * 100).toFixed(1)}%`;
}

function formatSeconds(value) {
  const parsed = numberValue(value);
  if (parsed <= 0) {
    return "-";
  }
  return `${Math.round(parsed)}s`;
}

function formatShortTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function labelOrderType(type) {
  return ORDER_TYPE_LABELS[type] || compact(type);
}

function labelOrderStatus(status) {
  return ORDER_STATUS_LABELS[status] || compact(status);
}

function labelAfterSalesStatus(status) {
  return AFTER_SALES_STATUS_LABELS[status] || compact(status);
}

function labelRefundStrategy(strategy) {
  return REFUND_STRATEGY_LABELS[strategy] || compact(strategy);
}

function orderRisk(order) {
  if (!order) {
    return "-";
  }
  if (order.status === "refund_pending" || order.status === "cancelled") {
    return "需客服复核";
  }
  if (order.status === "dispatching") {
    return "派单关注";
  }
  if (order.status === "rider_assigned") {
    return "履约关注";
  }
  return "无";
}

function merchantShopNames(merchant) {
  return (merchant?.shops || []).map((shop) => shop.name).filter(Boolean).join(" / ") || "-";
}

function merchantCapabilities(merchant) {
  const capabilities = new Set();
  for (const shop of merchant?.shops || []) {
    for (const capability of shop.capabilities || []) {
      capabilities.add(capability);
    }
  }
  return Array.from(capabilities).join(" / ") || merchant?.account?.type || "-";
}

function merchantQualificationText(merchant) {
  const missing = merchant?.missing_qualifications || [];
  if (missing.length > 0) {
    return `缺失：${missing.join(" / ")}`;
  }
  return "齐全";
}

function merchantQualificationExpiry(merchant) {
  const expiries = (merchant?.qualifications || [])
    .map((item) => item.expires_at)
    .filter(Boolean)
    .sort();
  return formatShortTime(expiries[0]);
}

function depositLabel(status) {
  if (status === "paid") {
    return "已缴";
  }
  if (status === "wechat_exempt_approved") {
    return "微信免押";
  }
  if (status === "refund_pending") {
    return "退押中";
  }
  if (status === "refunded") {
    return "已退";
  }
  return "未缴";
}

function riderStatus(rider) {
  if (rider?.type === "station_manager") {
    return rider.online ? "站长在线" : "站长";
  }
  return rider?.online ? "在线" : "离线";
}

function dispatchEventStage(event) {
  const type = event?.type || "";
  if (type.includes("timeout")) {
    return "确认超时";
  }
  if (type.includes("rejected")) {
    return "骑手拒单";
  }
  if (type.includes("assigned")) {
    return "已派单";
  }
  if (type.includes("grab")) {
    return "抢单成功";
  }
  if (type.includes("no_candidate")) {
    return "无候选";
  }
  return compact(type, "派单事件");
}

function limitedRows(rows, columns) {
  if (rows.length > 0) {
    return rows;
  }
  return [["暂无快照数据", ...Array(Math.max(0, columns.length - 1)).fill("-")]];
}

export function snapshotDataFromResult(result) {
  return result?.payload?.data || null;
}

export function buildSnapshotKpis(snapshot, fallbackKpis) {
  if (!snapshot?.counts) {
    return fallbackKpis;
  }
  const counts = snapshot.counts;
  return [
    { key: "totalOrders", title: "总订单", value: formatCount(counts.total_orders), trend: "快照口径", tone: "blue" },
    { key: "afterSales", title: "待售后处理", value: formatCount(numberValue(counts.after_sales_pending) + numberValue(counts.after_sales_admin_review)), trend: "商户/平台", tone: "red" },
    { key: "riderOnline", title: "在线骑手", value: formatCount(counts.online_riders), trend: `${formatCount(counts.total_riders)} 名骑手`, tone: "green" },
    { key: "merchantRisk", title: "商户风险", value: formatCount(numberValue(counts.merchant_qualification_risks) + numberValue(counts.merchant_deposit_missing)), trend: "资质/保证金", tone: "amber" },
    { key: "outbox", title: "Outbox 待处理", value: formatCount(numberValue(counts.outbox_ready) + numberValue(counts.outbox_blocked)), trend: `阻塞 ${formatCount(counts.outbox_blocked)}`, tone: counts.outbox_blocked > 0 ? "red" : "slate" },
    { key: "objectCleanup", title: "对象清理失败", value: formatCount(counts.object_cleanup_failed), trend: `候选 ${formatCount(counts.object_cleanup_total_candidate)}`, tone: counts.object_cleanup_failed > 0 ? "red" : "slate" }
  ];
}

export function buildSnapshotQueues(snapshot, fallbackQueues) {
  if (!snapshot?.counts) {
    return fallbackQueues;
  }
  const counts = snapshot.counts;
  return [
    { key: "after-sales-list", title: "售后审核", level: "P0", target: `${formatCount(numberValue(counts.after_sales_pending) + numberValue(counts.after_sales_admin_review))} 个待处理`, operationKey: "after-sales-list" },
    { key: "merchant-risk", title: "商户资质/保证金", level: "P0", target: `${formatCount(numberValue(counts.merchant_qualification_risks) + numberValue(counts.merchant_deposit_missing))} 个风险`, operationKey: "operations-snapshot" },
    { key: "rider-risk", title: "骑手准入", level: "P0", target: `${formatCount(counts.rider_deposit_missing)} 个未满足保证金`, operationKey: "operations-snapshot" },
    { key: "dispatch", title: "派单审计", level: "P0", target: `${formatCount(counts.dispatch_event_count)} 条事件`, operationKey: "station-orders" },
    { key: "outbox-stats", title: "事件队列健康", level: "P0", target: `Ready ${formatCount(counts.outbox_ready)} / Blocked ${formatCount(counts.outbox_blocked)}`, operationKey: "outbox-stats" },
    { key: "object-cleanup-stats", title: "对象清理", level: "P1", target: `${formatCount(counts.object_cleanup_failed)} 个失败`, operationKey: "object-cleanup-stats" }
  ];
}

export function applySnapshotToAdminView(view, snapshot) {
  if (!snapshot?.counts || !view) {
    return view;
  }
  const counts = snapshot.counts;
  const next = { ...view };
  if (view.key === "dashboard") {
    next.metrics = [
      { label: "待售后", value: formatCount(numberValue(counts.after_sales_pending) + numberValue(counts.after_sales_admin_review)), tone: "red" },
      { label: "商户风险", value: formatCount(numberValue(counts.merchant_qualification_risks) + numberValue(counts.merchant_deposit_missing)), tone: "amber" },
      { label: "在线骑手", value: formatCount(counts.online_riders), tone: "green" },
      { label: "队列待处理", value: formatCount(numberValue(counts.outbox_ready) + numberValue(counts.outbox_blocked)), tone: counts.outbox_blocked > 0 ? "red" : "blue" }
    ];
    next.rows = [
      ["售后审核", "客服", `${formatCount(numberValue(counts.after_sales_pending) + numberValue(counts.after_sales_admin_review))} 个待处理`, "售后列表"],
      ["商户资质", "招商", `${formatCount(counts.merchant_qualification_risks)} 个资质风险`, "资质台"],
      ["骑手调度", "配送", `${formatCount(counts.online_riders)} 名在线骑手`, "派单台"],
      ["Outbox", "运维", `Ready ${formatCount(counts.outbox_ready)} / Blocked ${formatCount(counts.outbox_blocked)}`, "事件台"]
    ];
    return next;
  }
  if (view.key === "orders") {
    next.metrics = [
      { label: "待商家接单", value: formatCount(counts.pending_merchant_orders), tone: "amber" },
      { label: "待派单", value: formatCount(counts.dispatching_orders), tone: "red" },
      { label: "骑手履约中", value: formatCount(counts.rider_assigned_orders), tone: "blue" },
      { label: "异常单", value: formatCount(counts.exception_orders), tone: counts.exception_orders > 0 ? "red" : "green" }
    ];
    next.rows = limitedRows((snapshot.orders || []).map((order) => [
      compact(order.id),
      labelOrderType(order.type),
      labelOrderStatus(order.status),
      compact(order.shop_id, "平台/未知"),
      compact(order.rider_id, "未分配"),
      orderRisk(order)
    ]), next.columns);
    return next;
  }
  if (view.key === "after-sales") {
    next.metrics = [
      { label: "商户待处理", value: formatCount(counts.after_sales_pending), tone: "red" },
      { label: "平台审核", value: formatCount(counts.after_sales_admin_review), tone: "amber" },
      { label: "对象清理失败", value: formatCount(counts.object_cleanup_failed), tone: counts.object_cleanup_failed > 0 ? "red" : "green" },
      { label: "可清理候选", value: formatCount(counts.object_cleanup_total_candidate), tone: "blue" }
    ];
    next.rows = limitedRows((snapshot.after_sales || []).map((item) => [
      compact(item.id),
      compact(item.order_id),
      compact(item.user_id),
      labelAfterSalesStatus(item.status),
      formatFen(item.refundable_fen ?? item.requested_amount_fen),
      `${(item.evidence_urls || []).length} 个附件`
    ]), next.columns);
    return next;
  }
  if (view.key === "merchants") {
    next.metrics = [
      { label: "商户总数", value: formatCount(counts.total_merchants), tone: "blue" },
      { label: "资质风险", value: formatCount(counts.merchant_qualification_risks), tone: counts.merchant_qualification_risks > 0 ? "red" : "green" },
      { label: "未缴保证金", value: formatCount(counts.merchant_deposit_missing), tone: counts.merchant_deposit_missing > 0 ? "red" : "green" },
      { label: "快照商户", value: formatCount((snapshot.merchants || []).length), tone: "slate" }
    ];
    next.rows = limitedRows((snapshot.merchants || []).map((merchant) => [
      compact(merchant.account?.display_name || merchant.account?.id),
      merchantShopNames(merchant),
      merchantCapabilities(merchant),
      depositLabel(merchant.deposit?.status || merchant.account?.deposit_status),
      merchantQualificationText(merchant),
      merchantQualificationExpiry(merchant)
    ]), next.columns);
    return next;
  }
  if (view.key === "riders") {
    next.metrics = [
      { label: "在线骑手", value: formatCount(counts.online_riders), tone: "green" },
      { label: "骑手总数", value: formatCount(counts.total_riders), tone: "blue" },
      { label: "站长账号", value: formatCount(counts.station_managers), tone: "slate" },
      { label: "待缴保证金", value: formatCount(counts.rider_deposit_missing), tone: counts.rider_deposit_missing > 0 ? "amber" : "green" }
    ];
    next.rows = limitedRows((snapshot.riders || []).map((rider) => [
      compact(rider.id),
      compact(rider.station_id),
      riderStatus(rider),
      rider.type === "station_manager" ? "站长" : depositLabel(rider.deposit_status),
      rider.type === "station_manager" ? "-" : `优先级 ${formatCount(rider.dispatch_priority)}`,
      rider.type === "station_manager" ? "全部" : `容量 ${formatCount(rider.capacity)}`
    ]), next.columns);
    return next;
  }
  if (view.key === "rider-performance") {
    const sLevelCount = (snapshot.rider_performance || []).filter((item) => item.level === "S").length;
    next.metrics = [
      { label: "绩效样本", value: formatCount((snapshot.rider_performance || []).length), tone: "blue" },
      { label: "S 级骑手", value: formatCount(sLevelCount), tone: "green" },
      { label: "在线骑手", value: formatCount(counts.online_riders), tone: "green" },
      { label: "待缴保证金", value: formatCount(counts.rider_deposit_missing), tone: counts.rider_deposit_missing > 0 ? "amber" : "green" }
    ];
    next.rows = limitedRows((snapshot.rider_performance || []).map((item) => [
      compact(item.rider_id),
      formatSeconds(item.average_accept_seconds),
      formatPercent(item.completion_rate),
      "-",
      compact(item.level),
      formatCount(item.dispatch_priority)
    ]), next.columns);
    return next;
  }
  if (view.key === "dispatch") {
    next.metrics = [
      { label: "派单事件", value: formatCount(counts.dispatch_event_count), tone: "blue" },
      { label: "待派单", value: formatCount(counts.dispatching_orders), tone: counts.dispatching_orders > 0 ? "amber" : "green" },
      { label: "骑手履约中", value: formatCount(counts.rider_assigned_orders), tone: "blue" },
      { label: "异常单", value: formatCount(counts.exception_orders), tone: counts.exception_orders > 0 ? "red" : "green" }
    ];
    next.rows = limitedRows((snapshot.dispatch_events || []).map((event) => [
      compact(event.order_id),
      dispatchEventStage(event),
      compact(event.online_candidate_size, "0"),
      compact(event.mode),
      compact(event.rider_id || event.reason, "等待"),
      formatShortTime(event.created_at)
    ]), next.columns);
    return next;
  }
  if (view.key === "refund-settings") {
    const strategy = snapshot.refund_settings?.default_refund_strategy;
    const outbox = snapshot.outbox_stats || {};
    next.metrics = [
      { label: "默认策略", value: labelRefundStrategy(strategy), tone: "blue" },
      { label: "Outbox Ready", value: formatCount(outbox.ready), tone: outbox.ready > 0 ? "amber" : "green" },
      { label: "Outbox Blocked", value: formatCount(outbox.blocked), tone: outbox.blocked > 0 ? "red" : "green" },
      { label: "死信", value: formatCount(outbox.dead_letter), tone: outbox.dead_letter > 0 ? "red" : "green" }
    ];
    next.rows = [
      [compact(strategy), strategy === "original_route_first" ? "微信原路优先" : "平台余额优先", "后台默认策略", "退款设置", "需财务审计"],
      ["outbox_ready", formatCount(outbox.ready), "可立即投递", "Outbox 事件", outbox.ready > 0 ? "需 relay 消费" : "无"],
      ["outbox_blocked", formatCount(outbox.blocked), "租约/重试中", "Outbox 事件", outbox.blocked > 0 ? "需运维关注" : "无"],
      ["dead_letter", formatCount(outbox.dead_letter), "人工解封", "Outbox 事件", outbox.dead_letter > 0 ? "高风险" : "无"]
    ];
    return next;
  }
  return next;
}
