const AFTER_SALES_ACTION_LABELS = Object.freeze({
  created: "用户提交售后",
  evidence_uploaded: "补充凭证",
  merchant_reply: "商户回复",
  customer_care: "客服跟进",
  internal_note: "内部备注",
  review_approved: "审核通过",
  review_rejected: "审核驳回",
  review_escalated: "升级复核",
  refunded: "已退款"
});

const ACTOR_ROLE_LABELS = Object.freeze({
  admin: "平台",
  merchant: "商户",
  rider: "骑手",
  station_manager: "站长",
  support: "客服",
  user: "用户"
});

const DISPATCH_TYPE_LABELS = Object.freeze({
  "dispatch.auto_assign": "自动派单",
  "dispatch.manual_assign": "手动派单",
  "dispatch.accepted": "骑手接单",
  "dispatch.rejected": "骑手拒单",
  "dispatch.timeout": "派单超时",
  "dispatch.completed": "配送完成"
});

const DISPATCH_MODE_LABELS = Object.freeze({
  auto_assign: "自动",
  manual_assign: "手动",
  system: "系统"
});

const EVIDENCE_STATUS_LABELS = Object.freeze({
  uploaded: "已确认",
  pending_confirm: "待确认",
  pending_scan: "待扫描",
  rejected: "已拦截",
  blocked: "已冻结"
});

const AFTER_SALES_STATUS_LABELS = Object.freeze({
  pending_merchant: "商户待处理",
  admin_review: "平台仲裁",
  approved: "审核通过",
  rejected: "审核驳回",
  refunded: "已退款"
});

const REFUND_DESTINATION_LABELS = Object.freeze({
  balance: "退平台余额",
  original_route: "原路返回"
});

const REFUND_STATUS_LABELS = Object.freeze({
  success: "成功",
  pending: "处理中",
  failed: "失败"
});

const SERVICE_TICKET_STATUS_LABELS = Object.freeze({
  processing: "处理中",
  waiting_confirm: "待用户确认",
  resolved: "已解决",
  closed: "已关闭"
});

const SERVICE_TICKET_EVENT_STATUS_LABELS = Object.freeze({
  done: "已完成",
  active: "处理中",
  pending: "待处理"
});

const ORDER_STATUS_LABELS = Object.freeze({
  pending_payment: "待支付",
  merchant_pending: "商户待处理",
  preparing: "备餐中",
  dispatching: "待配送",
  rider_assigned: "骑手已接单",
  completed: "已完成",
  cancelled: "已取消"
});

const ORDER_TYPE_LABELS = Object.freeze({
  takeout: "外卖",
  groupbuy: "团购",
  errand: "跑腿",
  medicine: "买药"
});

const PAYMENT_METHOD_LABELS = Object.freeze({
  balance: "余额支付",
  wechat: "微信支付"
});

const AUDIT_ACTION_LABELS = Object.freeze({
  "after_sales.reviewed": "售后审核",
  "admin.order.refunded": "订单退款",
  "admin.service_ticket.assigned": "工单分派",
  "admin.service_ticket.escalated": "工单升级",
  "admin.service_ticket.resolved": "工单处理方案",
  "admin.service_ticket.quality_reviewed": "工单质检"
});

function safeArray(value) {
  return Array.isArray(value) ? value : [];
}

function safeObject(value) {
  return value && typeof value === "object" && !Array.isArray(value) ? value : {};
}

function sumBy(items, selector) {
  return items.reduce((total, item) => total + selector(item), 0);
}

function formatCount(value, unit) {
  return `${value}${unit}`;
}

function formatFen(value) {
  const amount = Number(value);
  if (!Number.isFinite(amount)) {
    return "¥0.00";
  }
  return `¥${(amount / 100).toFixed(2)}`;
}

function formatBytes(value) {
  const bytes = Number(value);
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return "0 B";
  }
  if (bytes >= 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(bytes >= 10 * 1024 * 1024 ? 0 : 1)} MB`;
  }
  if (bytes >= 1024) {
    return `${Math.round(bytes / 1024)} KB`;
  }
  return `${bytes} B`;
}

function pad(value) {
  return String(value).padStart(2, "0");
}

function formatLocalTime(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return `${pad(date.getMonth() + 1)}/${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function actorRoleLabel(role) {
  return ACTOR_ROLE_LABELS[role] || role || "未知角色";
}

function shortHash(value) {
  const normalized = String(value || "").trim();
  if (!normalized) {
    return "";
  }
  if (normalized.length <= 18) {
    return normalized;
  }
  return `${normalized.slice(0, 10)}...${normalized.slice(-6)}`;
}

function previewActions(...items) {
  return items.filter(Boolean);
}

function previewAction(label, operationKey, values, enabled = true) {
  if (!enabled) {
    return null;
  }
  return {
    label,
    operationKey,
    values: safeObject(values)
  };
}

function extractIdFromUrl(url, segment) {
  const normalized = String(url || "");
  const match = normalized.match(new RegExp(`/${segment}/([^/?#]+)/`, "i"));
  return match?.[1] || "";
}

function dispatchTypeTone(type) {
  if (type === "dispatch.completed" || type === "dispatch.accepted") {
    return "green";
  }
  if (type === "dispatch.timeout") {
    return "amber";
  }
  if (type === "dispatch.rejected") {
    return "red";
  }
  if (type === "dispatch.manual_assign") {
    return "slate";
  }
  return "blue";
}

function afterSalesEventTone(action) {
  if (action === "review_approved" || action === "evidence_uploaded" || action === "refunded") {
    return "green";
  }
  if (action === "review_rejected") {
    return "red";
  }
  if (action === "customer_care" || action === "merchant_reply" || action === "review_escalated") {
    return "amber";
  }
  if (action === "internal_note") {
    return "slate";
  }
  return "blue";
}

function evidenceTone(status) {
  if (status === "uploaded") {
    return "green";
  }
  if (status === "rejected" || status === "blocked") {
    return "red";
  }
  if (status === "pending_scan" || status === "pending_confirm") {
    return "amber";
  }
  return "blue";
}

function dispatchTypeLabel(type) {
  return DISPATCH_TYPE_LABELS[type] || type || "派单事件";
}

function dispatchModeLabel(mode) {
  return DISPATCH_MODE_LABELS[mode] || mode || "未标记";
}

function afterSalesActionLabel(action) {
  return AFTER_SALES_ACTION_LABELS[action] || action || "售后事件";
}

function evidenceStatusLabel(status) {
  return EVIDENCE_STATUS_LABELS[status] || status || "未知状态";
}

function afterSalesStatusLabel(status) {
  return AFTER_SALES_STATUS_LABELS[status] || status || "未知状态";
}

function refundDestinationLabel(destination) {
  return REFUND_DESTINATION_LABELS[destination] || destination || "未标记";
}

function refundStatusLabel(status) {
  return REFUND_STATUS_LABELS[status] || status || "未知状态";
}

function refundTone(status) {
  if (status === "success") {
    return "green";
  }
  if (status === "failed") {
    return "red";
  }
  if (status === "pending") {
    return "amber";
  }
  return "slate";
}

function serviceTicketStatusLabel(status) {
  return SERVICE_TICKET_STATUS_LABELS[status] || status || "未知状态";
}

function serviceTicketEventStatusLabel(status) {
  return SERVICE_TICKET_EVENT_STATUS_LABELS[status] || status || "未知阶段";
}

function auditActionLabel(action) {
  return AUDIT_ACTION_LABELS[action] || action || "审计事件";
}

function orderStatusLabel(status) {
  return ORDER_STATUS_LABELS[status] || status || "未知状态";
}

function orderTypeLabel(type) {
  return ORDER_TYPE_LABELS[type] || type || "订单";
}

function paymentMethodLabel(method) {
  return PAYMENT_METHOD_LABELS[method] || method || "未标记";
}

function orderStatusTone(status) {
  if (status === "completed") {
    return "green";
  }
  if (status === "cancelled") {
    return "red";
  }
  if (status === "merchant_pending" || status === "preparing" || status === "dispatching" || status === "rider_assigned") {
    return "amber";
  }
  return "blue";
}

function serviceTicketTone(status, slaStatus = "") {
  if (status === "closed" || status === "resolved") {
    return "green";
  }
  if (slaStatus === "overdue") {
    return "red";
  }
  if (slaStatus === "escalated" || slaStatus === "due_soon") {
    return "amber";
  }
  if (status === "processing" || status === "waiting_confirm") {
    return "blue";
  }
  return "slate";
}

function afterSalesRequestTone(status) {
  if (status === "refunded" || status === "approved") {
    return "green";
  }
  if (status === "rejected") {
    return "red";
  }
  if (status === "admin_review" || status === "pending_merchant") {
    return "amber";
  }
  return "blue";
}

function buildAfterSalesEventPreview(items, requestUrl) {
  const requestID = items[0]?.request_id || extractIdFromUrl(requestUrl, "after-sales");
  const attachmentCount = sumBy(items, (item) => safeArray(item.attachments).length);
  const visibleCount = items.filter((item) => item.visible_to_user).length;
  return {
    key: "after-sales-events",
    title: "售后时间线",
    subtitle: `${requestID ? `工单 ${requestID}` : "当前工单"} · ${formatCount(items.length, " 条事件")}`,
    stats: [
      { label: "事件", value: formatCount(items.length, " 条"), tone: "blue" },
      { label: "用户可见", value: formatCount(visibleCount, " 条"), tone: "green" },
      { label: "内部备注", value: formatCount(items.length - visibleCount, " 条"), tone: "slate" },
      { label: "附件", value: formatCount(attachmentCount, " 个"), tone: "amber" }
    ],
    emptyMessage: "当前工单还没有售后时间线。",
    items: items.map((item) => ({
      title: afterSalesActionLabel(item.action),
      badge: item.visible_to_user ? "用户可见" : "内部备注",
      tone: afterSalesEventTone(item.action),
      meta: [
        formatLocalTime(item.created_at),
        `${actorRoleLabel(item.actor_role)} ${item.actor_id || "-"}`,
        item.action || "-"
      ],
      body: item.message || "无补充说明",
      note: item.id ? `事件 ${item.id}` : "",
      chips: safeArray(item.attachments).length > 0 ? [`附件 ${safeArray(item.attachments).length}`] : [],
      links: safeArray(item.attachments).map((href, index) => ({
        href,
        label: `附件 ${index + 1}`
      }))
    }))
  };
}

function buildAfterSalesEvidencePreview(items, requestUrl) {
  const requestID = items[0]?.request_id || extractIdFromUrl(requestUrl, "after-sales");
  const imageCount = items.filter((item) => String(item.content_type || "").startsWith("image/")).length;
  const totalBytes = sumBy(items, (item) => Number(item.size_bytes) || 0);
  return {
    key: "after-sales-evidence",
    title: "售后凭证",
    subtitle: `${requestID ? `工单 ${requestID}` : "当前工单"} · ${formatCount(items.length, " 份凭证")}`,
    stats: [
      { label: "凭证", value: formatCount(items.length, " 份"), tone: "blue" },
      { label: "图片", value: formatCount(imageCount, " 张"), tone: "green" },
      { label: "总大小", value: formatBytes(totalBytes), tone: "slate" },
      { label: "已确认", value: formatCount(items.filter((item) => item.confirmed_at).length, " 份"), tone: "amber" }
    ],
    emptyMessage: "当前工单还没有售后凭证。",
    items: items.map((item) => ({
      title: item.file_name || item.id || "未命名凭证",
      badge: evidenceStatusLabel(item.status),
      tone: evidenceTone(item.status),
      meta: [
        formatLocalTime(item.confirmed_at || item.created_at),
        formatBytes(item.size_bytes),
        item.content_type || "-"
      ],
      body: `上传人 ${actorRoleLabel(item.uploaded_by_role)} ${item.uploaded_by_id || "-"}`,
      note: item.object_key || "",
      chips: [
        ...(item.content_sha ? [`SHA ${shortHash(item.content_sha)}`] : []),
        ...(item.id ? [item.id] : [])
      ],
      links: item.public_url ? [{ href: item.public_url, label: "打开凭证" }] : [],
      previewImageUrl: String(item.content_type || "").startsWith("image/") ? item.public_url : ""
    }))
  };
}

function buildDispatchEventPreview(items, requestUrl) {
  const orderID = items[0]?.order_id || extractIdFromUrl(requestUrl, "orders");
  const timeoutCount = items.filter((item) => item.type === "dispatch.timeout").length;
  const rejectCount = items.filter((item) => item.type === "dispatch.rejected").length;
  const autoAssignCount = items.filter((item) => item.mode === "auto_assign").length;
  return {
    key: "dispatch-order-events",
    title: "订单派单事件",
    subtitle: `${orderID ? `订单 ${orderID}` : "当前订单"} · ${formatCount(items.length, " 条派单记录")}`,
    stats: [
      { label: "事件", value: formatCount(items.length, " 条"), tone: "blue" },
      { label: "自动派单", value: formatCount(autoAssignCount, " 条"), tone: "green" },
      { label: "拒单", value: formatCount(rejectCount, " 次"), tone: "red" },
      { label: "超时", value: formatCount(timeoutCount, " 次"), tone: "amber" }
    ],
    emptyMessage: "当前订单还没有派单事件。",
    items: items.map((item) => ({
      title: dispatchTypeLabel(item.type),
      badge: dispatchModeLabel(item.mode),
      tone: dispatchTypeTone(item.type),
      meta: [
        formatLocalTime(item.created_at),
        item.rider_id ? `骑手 ${item.rider_id}` : "未落骑手",
        item.actor_id ? `操作人 ${item.actor_id}` : `候选 ${item.online_candidate_size || 0}`
      ],
      body: item.reason || "无额外原因说明",
      note: item.idempotency_key ? `幂等键 ${item.idempotency_key}` : "",
      chips: [
        ...(item.can_decline_without_penalty ? ["免责拒派"] : []),
        ...(safeArray(item.rejected_rider_ids).length > 0 ? [`已拒 ${safeArray(item.rejected_rider_ids).length}`] : []),
        ...(item.station_id ? [`站点 ${item.station_id}`] : [])
      ],
      links: [],
      previewImageUrl: ""
    }))
  };
}

function buildRefundTransactionPreview(items, requestUrl) {
  const orderID = items[0]?.order_id || extractIdFromUrl(requestUrl, "refunds");
  const totalAmountFen = sumBy(items, (item) => Number(item.amount_fen) || 0);
  const balanceCount = items.filter((item) => item.destination === "balance").length;
  const originalRouteCount = items.filter((item) => item.destination === "original_route").length;
  return {
    key: "refund-transactions",
    title: "退款流水",
    subtitle: `${orderID ? `订单 ${orderID}` : "当前筛选"} · ${formatCount(items.length, " 笔退款")}`,
    stats: [
      { label: "退款", value: formatCount(items.length, " 笔"), tone: "blue" },
      { label: "累计金额", value: formatFen(totalAmountFen), tone: "green" },
      { label: "退余额", value: formatCount(balanceCount, " 笔"), tone: "amber" },
      { label: "原路退", value: formatCount(originalRouteCount, " 笔"), tone: "slate" }
    ],
    emptyMessage: "当前筛选下还没有退款流水。",
    items: items.map((item) => ({
      title: item.id || "退款记录",
      badge: refundStatusLabel(item.status),
      tone: refundTone(item.status),
      meta: [
        item.order_id ? `订单 ${item.order_id}` : "订单 -",
        item.user_id ? `用户 ${item.user_id}` : "用户 -",
        item.created_at ? formatLocalTime(item.created_at) : "暂无时间"
      ],
      body: `${formatFen(item.amount_fen)} · ${refundDestinationLabel(item.destination)}`,
      note: item.reason || "",
      chips: [
        ...(item.idempotency_key ? [`幂等 ${item.idempotency_key}`] : []),
        ...(item.destination ? [item.destination] : [])
      ],
      actions: previewActions(
        previewAction("退款审计", "audit-logs", {
          target_type: "order",
          target_id: item.order_id || orderID,
          action: "admin.order.refunded",
          limit: 20
        }, Boolean(item.order_id || orderID)),
        previewAction("订单总览", "order-detail", {
          order_id: item.order_id || orderID
        }, Boolean(item.order_id || orderID))
      ),
      links: [],
      previewImageUrl: ""
    }))
  };
}

function buildAdminOrderDetailPreview(detail) {
  const order = safeObject(detail?.order);
  const afterSalesRequests = safeArray(detail?.after_sales_requests);
  const refunds = safeArray(detail?.refunds);
  const serviceTickets = safeArray(detail?.service_tickets);
  const dispatchEvents = safeArray(detail?.dispatch_events);
  const relatedAudits = safeArray(detail?.related_audits);
  const afterSalesSummary = safeObject(detail?.after_sales_summary);
  const refundSummary = safeObject(detail?.refund_summary);
  const serviceTicketSummary = safeObject(detail?.service_ticket_summary);
  const dispatchSummary = safeObject(detail?.dispatch_summary);
  const auditSummary = safeObject(detail?.audit_summary);
  const latestAfterSales = afterSalesRequests[0] || null;
  const latestRefund = refunds.length > 0 ? refunds[refunds.length - 1] : null;
  const latestTicket = serviceTickets.length > 0 ? serviceTickets[0] : null;
  const latestAudit = relatedAudits.length > 0 ? relatedAudits[0] : null;
  const latestDispatchLabel = dispatchSummary.latest_type ? dispatchTypeLabel(dispatchSummary.latest_type) : "暂无";
  const itemSummary = safeArray(order.items)
    .slice(0, 3)
    .map((item) => `${item.product_name || item.productName || "商品"} x ${item.quantity || 0}`)
    .join("、");
  return {
    key: "order-detail",
    title: "订单聚合详情",
    subtitle: `${order.id ? `订单 ${order.id}` : "当前订单"}${order.shop_name || order.shopName ? ` · ${order.shop_name || order.shopName}` : ""}`,
    stats: [
      { label: "订单状态", value: orderStatusLabel(order.status), tone: orderStatusTone(order.status) },
      { label: "订单金额", value: formatFen(order.amount_fen), tone: "green" },
      { label: "售后记录", value: formatCount(Number(afterSalesSummary.total) || 0, " 单"), tone: "amber" },
      { label: "退款记录", value: formatCount(Number(refundSummary.total) || 0, " 笔"), tone: "blue" },
      { label: "客服工单", value: formatCount(Number(serviceTicketSummary.total) || 0, " 单"), tone: "slate" },
      { label: "派单关联", value: formatCount(Number(dispatchSummary.total) || 0, " 条"), tone: "amber" },
      { label: "关联审计", value: formatCount(Number(auditSummary.total) || 0, " 条"), tone: "slate" }
    ],
    emptyMessage: "当前订单暂无聚合详情。",
    items: [
      {
        title: order.shop_name || order.shopName || orderTypeLabel(order.type),
        badge: orderStatusLabel(order.status),
        tone: orderStatusTone(order.status),
        meta: [
          order.user_id ? `用户 ${order.user_id}` : "用户 -",
          orderTypeLabel(order.type),
          order.created_at ? formatLocalTime(order.created_at) : "暂无时间"
        ],
        body: itemSummary || "当前订单暂无商品明细。",
        note: order.address_snapshot?.detail ? `配送地址 ${order.address_snapshot.detail}` : "",
        chips: [
          `支付 ${paymentMethodLabel(order.payment_method)}`,
          `实付 ${formatFen(order.amount_fen)}`,
          ...(order.rider_id ? [`骑手 ${order.rider_id}`] : []),
          ...(order.reviewed ? ["已评价"] : ["待评价"])
        ],
        actions: previewActions(
          previewAction("查看退款流水", "refund-transactions", { order_id: order.id, limit: 20 }, Boolean(order.id)),
          previewAction("查看售后列表", "after-sales-list", { order_id: order.id, status: "" }, Boolean(order.id)),
          previewAction("查看工单列表", "support-tickets", { related_order_id: order.id, status: "", sla_status: "", assigned_support_id: "", limit: 20 }, Boolean(order.id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "售后概览",
        badge: `${Number(afterSalesSummary.total) || 0} 单`,
        tone: (Number(afterSalesSummary.open_count) || 0) > 0 ? "amber" : (Number(afterSalesSummary.total) || 0) > 0 ? "green" : "slate",
        meta: [
          `处理中 ${Number(afterSalesSummary.open_count) || 0} 单`,
          `已退款 ${Number(afterSalesSummary.refunded_count) || 0} 单`,
          afterSalesSummary.latest_updated_at ? formatLocalTime(afterSalesSummary.latest_updated_at) : "暂无售后"
        ],
        body: latestAfterSales ? `${afterSalesStatusLabel(latestAfterSales.status)} · ${latestAfterSales.reason || latestAfterSales.latest_event_message || "售后单"}` : "当前订单还没有关联售后单。",
        note: latestAfterSales?.id ? `工单 ${latestAfterSales.id}` : "",
        chips: latestAfterSales?.requested_amount_fen ? [`申请 ${formatFen(latestAfterSales.requested_amount_fen)}`] : [],
        actions: previewActions(
          previewAction("最新售后详情", "after-sales-detail", { request_id: latestAfterSales?.id || "" }, Boolean(latestAfterSales?.id)),
          previewAction("查看售后列表", "after-sales-list", { order_id: order.id, status: "" }, Boolean(order.id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "退款概览",
        badge: `${Number(refundSummary.total) || 0} 笔`,
        tone: (Number(refundSummary.success_count) || 0) > 0 ? "green" : "slate",
        meta: [
          `成功 ${Number(refundSummary.success_count) || 0} 笔`,
          `累计 ${formatFen(refundSummary.total_amount_fen)}`,
          refundSummary.latest_created_at ? formatLocalTime(refundSummary.latest_created_at) : "暂无退款"
        ],
        body: latestRefund ? `最近退款 ${formatFen(latestRefund.amount_fen)}，${refundDestinationLabel(refundSummary.latest_destination || latestRefund.destination)}。` : "当前订单还没有退款记录。",
        note: latestRefund?.id ? `退款 ${latestRefund.id}` : "",
        chips: latestRefund?.status ? [latestRefund.status] : [],
        actions: previewActions(
          previewAction("查看退款流水", "refund-transactions", { order_id: order.id, limit: 20 }, Boolean(order.id)),
          previewAction("退款审计", "audit-logs", { target_type: "order", target_id: order.id, action: "admin.order.refunded", limit: 20 }, Boolean(order.id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "客服工单",
        badge: `${Number(serviceTicketSummary.total) || 0} 单`,
        tone: (Number(serviceTicketSummary.escalated_count) || 0) > 0 ? "amber" : (Number(serviceTicketSummary.total) || 0) > 0 ? "blue" : "slate",
        meta: [
          `处理中 ${Number(serviceTicketSummary.open_count) || 0} 单`,
          `升级 ${Number(serviceTicketSummary.escalated_count) || 0} 单`,
          serviceTicketSummary.latest_updated_at ? formatLocalTime(serviceTicketSummary.latest_updated_at) : "暂无工单"
        ],
        body: latestTicket ? `${serviceTicketStatusLabel(latestTicket.status)} · ${latestTicket.title || latestTicket.category || "客服工单"}` : "当前订单还没有客服工单。",
        note: latestTicket?.assigned_support_name ? `当前客服 ${latestTicket.assigned_support_name}` : "",
        chips: latestTicket?.id ? [latestTicket.id] : [],
        actions: previewActions(
          previewAction("工单详情", "support-ticket-detail", { ticket_id: latestTicket?.id || "" }, Boolean(latestTicket?.id)),
          previewAction("工单列表", "support-tickets", { related_order_id: order.id, status: "", sla_status: "", assigned_support_id: "", limit: 20 }, Boolean(order.id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "派单关联",
        badge: `${Number(dispatchSummary.total) || 0} 条`,
        tone: (Number(dispatchSummary.timeout_count) || 0) > 0 || (Number(dispatchSummary.reject_count) || 0) > 0 ? "amber" : "blue",
        meta: [
          `最近事件 ${latestDispatchLabel}`,
          dispatchSummary.latest_event_at ? formatLocalTime(dispatchSummary.latest_event_at) : "暂无派单记录",
          `自动 ${Number(dispatchSummary.auto_assign_count) || 0} / 手动 ${Number(dispatchSummary.manual_assign_count) || 0}`
        ],
        body: `拒单 ${Number(dispatchSummary.reject_count) || 0} 次，超时 ${Number(dispatchSummary.timeout_count) || 0} 次。`,
        note: "",
        chips: [],
        actions: previewActions(
          previewAction("查看派单事件", "dispatch-order-events", { order_id: order.id }, Boolean(order.id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "审计概览",
        badge: `${Number(auditSummary.total) || 0} 条`,
        tone: (Number(auditSummary.total) || 0) === 0 ? "slate" : (Number(auditSummary.verified_count) || 0) === (Number(auditSummary.total) || 0) ? "green" : "amber",
        meta: [
          `已验 ${Number(auditSummary.verified_count) || 0} / ${Number(auditSummary.total) || 0}`,
          `订单 ${Number(auditSummary.order_count) || 0} / 售后 ${Number(auditSummary.after_sales_count) || 0} / 工单 ${Number(auditSummary.service_ticket_count) || 0}`,
          auditSummary.latest_created_at ? formatLocalTime(auditSummary.latest_created_at) : "暂无审计"
        ],
        body: latestAudit ? `最近审计 ${auditActionLabel(latestAudit.action)}，目标 ${latestAudit.target_type || "-"}:${latestAudit.target_id || "-"}` : "当前订单还没有关联审计。",
        note: latestAudit?.request_id ? `请求 ${latestAudit.request_id}` : "",
        chips: latestAudit?.id ? [latestAudit.id] : [],
        actions: previewActions(
          previewAction("订单审计", "audit-logs", { target_type: "order", target_id: order.id, limit: 20 }, Boolean(order.id)),
          previewAction("退款审计", "audit-logs", { target_type: "order", target_id: order.id, action: "admin.order.refunded", limit: 20 }, Boolean(order.id)),
          previewAction("售后审计", "audit-logs", { target_type: "after_sales", target_id: latestAfterSales?.id || "", limit: 20 }, Boolean(latestAfterSales?.id))
        ),
        links: [],
        previewImageUrl: ""
      }
    ]
  };
}

function buildAdminAfterSalesDetailPreview(detail) {
  const request = safeObject(detail?.request);
  const eventSummary = safeObject(detail?.event_summary);
  const evidenceSummary = safeObject(detail?.evidence_summary);
  const dispatchSummary = safeObject(detail?.dispatch_summary);
  const refundSummary = safeObject(detail?.refund_summary);
  const serviceTicketSummary = safeObject(detail?.service_ticket_summary);
  const auditSummary = safeObject(detail?.audit_summary);
  const refunds = safeArray(detail?.refunds);
  const serviceTickets = safeArray(detail?.service_tickets);
  const relatedAudits = safeArray(detail?.related_audits);
  const evidenceItems = safeArray(detail?.evidence);
  const firstImage = evidenceItems.find((item) => String(item?.content_type || "").startsWith("image/"));
  const latestRefund = refunds.length > 0 ? refunds[refunds.length - 1] : null;
  const latestServiceTicket = serviceTickets.length > 0 ? serviceTickets[0] : null;
  const latestAudit = relatedAudits.length > 0 ? relatedAudits[0] : null;
  const latestEventLabel = eventSummary.latest_action ? afterSalesActionLabel(eventSummary.latest_action) : "暂无";
  const latestDispatchLabel = dispatchSummary.latest_type ? dispatchTypeLabel(dispatchSummary.latest_type) : "暂无";
  const requestID = String(request.id || "");
  const orderID = String(request.order_id || latestRefund?.order_id || "");
  return {
    key: "after-sales-detail",
    title: "售后聚合详情",
    subtitle: `${request.id ? `工单 ${request.id}` : "当前工单"}${request.order_id ? ` · 订单 ${request.order_id}` : ""}`,
    stats: [
      { label: "工单状态", value: afterSalesStatusLabel(request.status), tone: afterSalesRequestTone(request.status) },
      { label: "事件", value: formatCount(Number(eventSummary.total) || 0, " 条"), tone: "blue" },
      { label: "凭证", value: formatCount(Number(evidenceSummary.total) || 0, " 份"), tone: "green" },
      { label: "派单关联", value: formatCount(Number(dispatchSummary.total) || 0, " 条"), tone: "amber" },
      { label: "退款记录", value: formatCount(Number(refundSummary.total) || 0, " 笔"), tone: "blue" },
      { label: "客服工单", value: formatCount(Number(serviceTicketSummary.total) || 0, " 单"), tone: "slate" },
      { label: "关联审计", value: formatCount(Number(auditSummary.total) || 0, " 条"), tone: "slate" }
    ],
    emptyMessage: "当前工单暂无聚合详情。",
    items: [
      {
        title: request.shop_name || request.order_item_summary || "售后工单",
        badge: afterSalesStatusLabel(request.status),
        tone: afterSalesRequestTone(request.status),
        meta: [
          request.user_id ? `用户 ${request.user_id}` : "用户 -",
          request.order_status ? `订单 ${request.order_status}` : "订单状态 -",
          request.latest_event_at ? formatLocalTime(request.latest_event_at) : "暂无最新时间"
        ],
        body: request.reason || request.latest_event_message || "暂无补充说明",
        note: request.order_item_summary || "",
        chips: [
          `申请 ${formatFen(request.requested_amount_fen)}`,
          `可退 ${formatFen(request.refundable_fen)}`,
          `已退 ${formatFen(request.refunded_amount_fen)}`
        ],
        actions: previewActions(
          previewAction("工单列表", "after-sales-list", {
            request_id: requestID,
            order_id: orderID,
            status: request.status || ""
          }, Boolean(requestID || orderID)),
          previewAction("订单总览", "order-detail", {
            order_id: orderID
          }, Boolean(orderID))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "时间线概览",
        badge: `${Number(eventSummary.total) || 0} 条`,
        tone: "blue",
        meta: [
          `最近动作 ${latestEventLabel}`,
          eventSummary.latest_event_at ? formatLocalTime(eventSummary.latest_event_at) : "暂无时间线",
          `附件 ${Number(eventSummary.attachment_count) || 0}`
        ],
        body: `用户可见 ${Number(eventSummary.user_visible) || 0} 条，内部备注 ${Number(eventSummary.internal_only) || 0} 条。`,
        note: request.latest_event_message || "",
        actions: previewActions(
          previewAction("查看时间线", "after-sales-events", { request_id: requestID }, Boolean(requestID)),
          previewAction("售后审计", "audit-logs", {
            target_type: "after_sales",
            target_id: requestID,
            limit: 20
          }, Boolean(requestID))
        ),
        chips: [],
        links: [],
        previewImageUrl: ""
      },
      {
        title: "凭证概览",
        badge: `${Number(evidenceSummary.total) || 0} 份`,
        tone: (Number(evidenceSummary.total) || 0) > 0 ? "green" : "slate",
        meta: [
          `图片 ${Number(evidenceSummary.image_count) || 0} 张`,
          `已确认 ${Number(evidenceSummary.confirmed_count) || 0} 份`,
          formatBytes(evidenceSummary.total_size_bytes)
        ],
        body: firstImage?.file_name ? `首张凭证 ${firstImage.file_name}` : "当前工单暂无已确认图片凭证。",
        note: evidenceSummary.latest_confirmed_at ? `最近确认 ${formatLocalTime(evidenceSummary.latest_confirmed_at)}` : "",
        actions: previewActions(
          previewAction("查看凭证", "after-sales-evidence", { request_id: requestID }, Boolean(requestID))
        ),
        chips: [],
        links: firstImage?.public_url ? [{ href: firstImage.public_url, label: "打开首张凭证" }] : [],
        previewImageUrl: String(firstImage?.public_url || "")
      },
      {
        title: "派单关联",
        badge: `${Number(dispatchSummary.total) || 0} 条`,
        tone: (Number(dispatchSummary.timeout_count) || 0) > 0 || (Number(dispatchSummary.reject_count) || 0) > 0 ? "amber" : "blue",
        meta: [
          `最近事件 ${latestDispatchLabel}`,
          dispatchSummary.latest_event_at ? formatLocalTime(dispatchSummary.latest_event_at) : "暂无派单记录",
          `自动 ${Number(dispatchSummary.auto_assign_count) || 0} / 手动 ${Number(dispatchSummary.manual_assign_count) || 0}`
        ],
        body: `拒单 ${Number(dispatchSummary.reject_count) || 0} 次，超时 ${Number(dispatchSummary.timeout_count) || 0} 次。`,
        note: "",
        actions: previewActions(
          previewAction("查看派单事件", "dispatch-order-events", { order_id: orderID }, Boolean(orderID)),
          previewAction("订单审计", "audit-logs", {
            target_type: "order",
            target_id: orderID,
            limit: 20
          }, Boolean(orderID))
        ),
        chips: [],
        links: [],
        previewImageUrl: ""
      },
      {
        title: "退款概览",
        badge: `${Number(refundSummary.total) || 0} 笔`,
        tone: (Number(refundSummary.success_count) || 0) > 0 ? "green" : "slate",
        meta: [
          `成功 ${Number(refundSummary.success_count) || 0} 笔`,
          `累计 ${formatFen(refundSummary.total_amount_fen)}`,
          refundSummary.latest_created_at ? formatLocalTime(refundSummary.latest_created_at) : "暂无退款"
        ],
        body: latestRefund ? `最近退款 ${formatFen(latestRefund.amount_fen)}，${refundDestinationLabel(refundSummary.latest_destination || latestRefund.destination)}。` : "当前工单还没有关联退款记录。",
        note: latestRefund?.id ? `退款 ${latestRefund.id}` : "",
        actions: previewActions(
          previewAction("查看退款流水", "refund-transactions", {
            order_id: orderID,
            limit: 20
          }, Boolean(orderID)),
          previewAction("退款审计", "audit-logs", {
            target_type: "order",
            target_id: orderID,
            action: "admin.order.refunded",
            limit: 20
          }, Boolean(orderID))
        ),
        chips: latestRefund?.status ? [latestRefund.status] : [],
        links: [],
        previewImageUrl: ""
      },
      {
        title: "客服工单",
        badge: `${Number(serviceTicketSummary.total) || 0} 单`,
        tone: (Number(serviceTicketSummary.escalated_count) || 0) > 0 ? "amber" : (Number(serviceTicketSummary.total) || 0) > 0 ? "blue" : "slate",
        meta: [
          `处理中 ${Number(serviceTicketSummary.open_count) || 0} 单`,
          `升级 ${Number(serviceTicketSummary.escalated_count) || 0} 单`,
          serviceTicketSummary.latest_updated_at ? formatLocalTime(serviceTicketSummary.latest_updated_at) : "暂无工单"
        ],
        body: latestServiceTicket ? `${serviceTicketStatusLabel(latestServiceTicket.status)} · ${latestServiceTicket.title || latestServiceTicket.category || "客服工单"}` : "当前订单还没有关联客服工单。",
        note: latestServiceTicket?.assigned_support_name ? `当前客服 ${latestServiceTicket.assigned_support_name}` : "",
        actions: previewActions(
          previewAction("工单详情", "support-ticket-detail", {
            ticket_id: latestServiceTicket?.id || ""
          }, Boolean(latestServiceTicket?.id)),
          previewAction("工单列表", "support-tickets", {
            related_order_id: orderID,
            status: latestServiceTicket?.status || "",
            limit: 20
          }, Boolean(orderID))
        ),
        chips: latestServiceTicket?.id ? [latestServiceTicket.id] : [],
        links: [],
        previewImageUrl: ""
      },
      {
        title: "审计概览",
        badge: `${Number(auditSummary.total) || 0} 条`,
        tone: (Number(auditSummary.total) || 0) === 0 ? "slate" : (Number(auditSummary.verified_count) || 0) === (Number(auditSummary.total) || 0) ? "green" : "amber",
        meta: [
          `已验 ${Number(auditSummary.verified_count) || 0} / ${Number(auditSummary.total) || 0}`,
          `订单 ${Number(auditSummary.order_count) || 0} / 售后 ${Number(auditSummary.after_sales_count) || 0} / 工单 ${Number(auditSummary.service_ticket_count) || 0}`,
          auditSummary.latest_created_at ? formatLocalTime(auditSummary.latest_created_at) : "暂无审计"
        ],
        body: latestAudit ? `最近审计 ${auditActionLabel(latestAudit.action)}，目标 ${latestAudit.target_type || "-"}:${latestAudit.target_id || "-"}` : "当前工单还没有关联审计记录。",
        note: latestAudit?.request_id ? `请求 ${latestAudit.request_id}` : "",
        actions: previewActions(
          previewAction("订单审计", "audit-logs", {
            target_type: "order",
            target_id: orderID,
            limit: 20
          }, Boolean(orderID)),
          previewAction("售后审计", "audit-logs", {
            target_type: "after_sales",
            target_id: requestID,
            limit: 20
          }, Boolean(requestID)),
          previewAction("退款审计", "audit-logs", {
            target_type: "order",
            target_id: orderID,
            action: "admin.order.refunded",
            limit: 20
          }, Boolean(orderID && Number(refundSummary.total) > 0))
        ),
        chips: latestAudit?.id ? [latestAudit.id] : [],
        links: [],
        previewImageUrl: ""
      }
    ]
  };
}

function buildSupportTicketDetailPreview(detail) {
  const ticket = safeObject(detail?.ticket);
  const events = safeArray(detail?.events);
  const latestEvent = events.length > 0 ? events[events.length - 1] : null;
  const attachmentCount = sumBy(events, (item) => safeArray(item.attachments).length) + safeArray(ticket.attachments).length;
  return {
    key: "support-ticket-detail",
    title: "客服工单详情",
    subtitle: `${ticket.id ? `工单 ${ticket.id}` : "当前工单"}${ticket.related_order_id ? ` · 订单 ${ticket.related_order_id}` : ""}`,
    stats: [
      { label: "工单状态", value: serviceTicketStatusLabel(ticket.status), tone: serviceTicketTone(ticket.status, ticket.sla_status) },
      { label: "SLA", value: ticket.sla_status || "未标记", tone: serviceTicketTone(ticket.status, ticket.sla_status) },
      { label: "事件", value: formatCount(events.length, " 条"), tone: "blue" },
      { label: "附件", value: formatCount(attachmentCount, " 个"), tone: "slate" }
    ],
    emptyMessage: "当前工单暂无详情。",
    items: [
      {
        title: ticket.title || ticket.category || "客服工单",
        badge: serviceTicketStatusLabel(ticket.status),
        tone: serviceTicketTone(ticket.status, ticket.sla_status),
        meta: [
          ticket.user_id ? `用户 ${ticket.user_id}` : "用户 -",
          ticket.related_order_id ? `订单 ${ticket.related_order_id}` : "未关联订单",
          ticket.updated_at ? formatLocalTime(ticket.updated_at) : "暂无更新时间"
        ],
        body: ticket.content || "暂无问题描述",
        note: ticket.solution || ticket.escalation_reason || "",
        chips: [
          ...(ticket.category ? [ticket.category] : []),
          ...(ticket.assigned_support_name ? [`客服 ${ticket.assigned_support_name}`] : []),
          ...(ticket.escalation_level ? [`升级 ${ticket.escalation_level}`] : [])
        ],
        actions: previewActions(
          previewAction("订单总览", "order-detail", {
            order_id: ticket.related_order_id || ""
          }, Boolean(ticket.related_order_id))
        ),
        links: [],
        previewImageUrl: ""
      },
      {
        title: "SLA / 责任人",
        badge: ticket.sla_status || "未标记",
        tone: serviceTicketTone(ticket.status, ticket.sla_status),
        meta: [
          ticket.reply_due_at ? `首响截至 ${formatLocalTime(ticket.reply_due_at)}` : "未返回首响时间",
          ticket.assigned_at ? `分派 ${formatLocalTime(ticket.assigned_at)}` : "未分派",
          ticket.escalated_at ? `升级 ${formatLocalTime(ticket.escalated_at)}` : "未升级"
        ],
        body: ticket.assigned_support_name ? `当前由 ${ticket.assigned_support_name} 跟进。` : "当前还没有指定客服责任人。",
        note: ticket.severity ? `严重级别 ${ticket.severity}` : "",
        chips: [],
        links: [],
        previewImageUrl: ""
      },
      {
        title: "最新进展",
        badge: latestEvent ? serviceTicketEventStatusLabel(latestEvent.status) : "暂无事件",
        tone: latestEvent ? serviceTicketTone(ticket.status, ticket.sla_status) : "slate",
        meta: [
          latestEvent?.created_at ? formatLocalTime(latestEvent.created_at) : "暂无时间",
          latestEvent?.actor_role ? actorRoleLabel(latestEvent.actor_role) : "未知角色",
          latestEvent?.actor_id || "-"
        ],
        body: latestEvent?.message || "当前工单还没有补充处理进展。",
        note: latestEvent?.title || "",
        chips: safeArray(latestEvent?.attachments).length > 0 ? [`附件 ${safeArray(latestEvent?.attachments).length}`] : [],
        links: safeArray(latestEvent?.attachments).map((href, index) => ({ href, label: `附件 ${index + 1}` })),
        previewImageUrl: ""
      },
      {
        title: "时间线概览",
        badge: `${events.length} 条`,
        tone: events.length > 0 ? "blue" : "slate",
        meta: [
          `已完成 ${events.filter((item) => item.status === "done").length} 条`,
          `处理中 ${events.filter((item) => item.status === "active").length} 条`,
          `待处理 ${events.filter((item) => item.status === "pending").length} 条`
        ],
        body: events.length > 0 ? `首条 ${events[0]?.title || "已提交"}，最新 ${latestEvent?.title || "处理中"}。` : "当前工单还没有时间线事件。",
        note: ticket.follow_up_comment || "",
        chips: ticket.follow_up_rating ? [`回访 ${ticket.follow_up_rating} 分`] : [],
        links: [],
        previewImageUrl: ""
      }
    ]
  };
}

export function previewAdminResult(result) {
  if (!result?.ok || result?.payload?.success === false) {
    return null;
  }
  const data = result?.payload?.data;
  const items = safeArray(data);
  const requestUrl = result?.request?.url || "";
  switch (result?.operation?.key) {
    case "order-detail":
      return buildAdminOrderDetailPreview(safeObject(data));
    case "after-sales-detail":
      return buildAdminAfterSalesDetailPreview(safeObject(data));
    case "refund-transactions":
      return buildRefundTransactionPreview(items, requestUrl);
    case "support-ticket-detail":
      return buildSupportTicketDetailPreview(safeObject(data));
    case "after-sales-events":
      return buildAfterSalesEventPreview(items, requestUrl);
    case "after-sales-evidence":
      return buildAfterSalesEvidencePreview(items, requestUrl);
    case "dispatch-order-events":
      return buildDispatchEventPreview(items, requestUrl);
    default:
      return null;
  }
}
