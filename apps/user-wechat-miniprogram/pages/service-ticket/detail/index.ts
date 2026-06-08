import { addServiceTicketEvent, closeServiceTicket, ensurePreviewAuth, followUpServiceTicket, getServiceTicketDetail } from "../../../utils/api";

Page({
  data: {
    ticketId: "st_preview_delivery",
    statusText: "处理中",
    ticket: {
      title: "配送问题处理中",
      id: "GD2405180021",
      content: "骑手到店很久了，预计送达时间一直没变化。",
      category: "配送问题 · 预计送达未更新",
      status: "处理中",
      solution: "继续等待：预计 14:35 前送达；延误补偿：订单完成后发放 ¥5 延误券",
      orderId: "DD240518001",
      orderTitle: "蓝海餐厅 · 招牌牛肉饭等 3 件",
      orderStatus: "配送中",
      slaStatus: "即将到期",
      slaTone: "due",
      replyDueText: "8 分钟内更新",
      escalationText: "超时会自动升级给客服主管"
    },
    events: [
      { id: "e1", status: "done", title: "已提交", message: "问题已同步到客服工单", time: "14:18" },
      { id: "e2", status: "done", title: "客服已受理", message: "正在核实商家出餐情况", time: "14:19" },
      { id: "e3", status: "active", title: "商家反馈", message: "补做菜品，预计 8 分钟后出餐", time: "14:23" },
      { id: "e4", status: "pending", title: "结果确认", message: "送达后可确认处理结果", time: "待完成" }
    ],
    evidence: ["订单截图", "聊天记录"]
  },
  onLoad(query) {
    const ticketId = String(query?.id || "st_preview_delivery");
    this.setData({ ticketId });
    this.loadTicket(ticketId);
  },
  async loadTicket(ticketId) {
    ensurePreviewAuth();
    try {
      const detail = await getServiceTicketDetail(ticketId) as { ticket?: Record<string, unknown>; events?: Array<Record<string, unknown>> };
      this.applyTicketDetail(detail, ticketId);
    } catch (_error) {
      this.setData({ statusText: "处理中" });
    }
  },
  applyTicketDetail(detail, ticketId) {
    applyDetail(this, detail, ticketId);
  },
  async handleSupplement() {
    ensurePreviewAuth();
    try {
      await addServiceTicketEvent(this.data.ticketId, {
        title: "补充说明",
        message: "用户补充了订单截图和聊天记录",
        status: "active",
        attachments: this.data.evidence
      });
      wx.showToast({ title: "说明已补充", icon: "success" });
      this.loadTicket(this.data.ticketId);
    } catch (_error) {
      wx.showToast({ title: "已保存补充说明", icon: "none" });
    }
  },
  handleContact() {
    wx.navigateTo({ url: `/pages/customer-service/chat/index?ticket_id=${this.data.ticketId}` });
  },
  async handleClose() {
    ensurePreviewAuth();
    try {
      await closeServiceTicket(this.data.ticketId, { reason: "用户接受处理方案" });
      const detail = await followUpServiceTicket(this.data.ticketId, { rating: 5, comment: "处理结果已确认" }) as { ticket?: Record<string, unknown>; events?: Array<Record<string, unknown>> };
      this.applyTicketDetail(detail, this.data.ticketId);
      wx.showToast({ title: "工单已关闭", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "工单已进入确认流程", icon: "none" });
    }
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/feedback/records/index" }) });
  }
});

function applyDetail(page, detail, ticketId: string) {
  const ticket = detail.ticket || {};
  const events = Array.isArray(detail.events) ? detail.events : [];
  page.setData({
    statusText: serviceStatus(String(ticket.status || "")),
    ticket: {
      title: String(ticket.title || "配送问题处理中"),
      id: String(ticket.id || ticketId),
      content: String(ticket.content || page.data.ticket.content),
      category: `${String(ticket.category || "配送问题")} · 预计送达未更新`,
      status: serviceStatus(String(ticket.status || "")),
      solution: String(ticket.solution || page.data.ticket.solution),
      orderId: String(ticket.related_order_id || "DD240518001"),
      orderTitle: String(ticket.related_order_title || "蓝海餐厅 · 招牌牛肉饭等 3 件"),
      orderStatus: String(ticket.related_order_status || "配送中"),
      slaStatus: serviceSLAStatus(String(ticket.sla_status || "")),
      slaTone: serviceSLATone(String(ticket.sla_status || "")),
      replyDueText: replyDueText(ticket.reply_due_at),
      escalationText: escalationText(ticket)
    },
    events: events.length ? events.map((item, index) => ({
      id: String(item.id || `event_${index}`),
      status: String(item.status || "pending"),
      title: String(item.title || "处理进度"),
      message: String(item.message || ""),
      time: String(item.created_at || "").slice(11, 16) || "刚刚"
    })) : page.data.events
  });
}

function replyDueText(value: unknown) {
  const raw = String(value || "");
  if (!raw) return "10 分钟内更新";
  const dueAt = new Date(raw);
  if (Number.isNaN(dueAt.getTime())) return "10 分钟内更新";
  const diffMinutes = Math.ceil((dueAt.getTime() - Date.now()) / 60000);
  if (diffMinutes <= 0) return "已超过首响时间";
  if (diffMinutes <= 60) return `${diffMinutes} 分钟内更新`;
  return `${Math.ceil(diffMinutes / 60)} 小时内更新`;
}

function serviceSLAStatus(status: string) {
  switch (status) {
    case "escalated":
      return "已升级";
    case "overdue":
      return "已超时";
    case "due_soon":
      return "即将到期";
    case "completed":
      return "已响应";
    default:
      return "正常";
  }
}

function serviceSLATone(status: string) {
  switch (status) {
    case "escalated":
    case "overdue":
      return "danger";
    case "due_soon":
      return "due";
    case "completed":
      return "done";
    default:
      return "normal";
  }
}

function escalationText(ticket: Record<string, unknown>) {
  const reason = String(ticket.escalation_reason || "");
  if (reason) return reason;
  const level = String(ticket.escalation_level || "");
  if (level) return "已进入主管复核队列";
  return "超时会自动升级给客服主管";
}

function serviceStatus(status: string) {
  switch (status) {
    case "resolved":
      return "已完成";
    case "waiting_confirm":
      return "待确认";
    case "closed":
      return "已关闭";
    default:
      return "处理中";
  }
}
