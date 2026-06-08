import { addServiceTicketEvent, createServiceTicket, ensurePreviewAuth, getServiceTickets } from "../../../utils/api";

function detectCustomerServiceMessageRisk(content: string) {
  const text = String(content || "").trim().toLowerCase();
  const hasSensitiveWord = /支付密码|付款密码|验证码|校验码|短信码|银行卡|银行卡号|卡号|payment password|pay password|verification code|sms code|bank card|card number/.test(text);
  if (!hasSensitiveWord) return { blocked: false, warning: "" };
  const hasSecret = /\d{4,}/.test(text) || /发给你|发给客服|告诉你|告诉客服|提供给你|提供给客服|报给你|报给客服|给你看|给客服看|my code is|my password is|my card number is/.test(text);
  if (hasSecret) {
    return {
      blocked: true,
      warning: "消息包含支付密码、验证码或银行卡等敏感信息，已为你拦截。客服不会索要这些信息。"
    };
  }
  return {
    blocked: false,
    warning: "已检测到敏感词，发送前请确认不要包含支付密码、验证码或银行卡号。"
  };
}

function isRiskControlError(error: unknown) {
  const message = String((error as { message?: string })?.message || error || "");
  return /risk control|支付密码|付款密码|验证码|银行卡|敏感信息/.test(message);
}

Page({
  data: {
    ticketId: "",
    draft: "",
    statusText: "工单",
    riskTip: "",
    activeScene: "订单售后",
    scenes: ["订单售后", "退款进度", "配送问题", "红包钱包", "账户安全"],
    orderCard: {
      id: "DD240518001",
      shop: "蓝海餐厅",
      items: "招牌牛肉饭等 3 件",
      amount: "46.90",
      status: "配送中"
    },
    messages: [
      { id: "m1", side: "left", avatar: "客", sender: "客服助手", content: "你好，我是悦享 e 食客服。请描述你遇到的问题，也可以直接选择下方快捷入口。", time: "刚刚" },
      { id: "m2", side: "right", avatar: "我", sender: "我", content: "骑手到店很久了，预计送达时间一直没变化。", time: "14:18" },
      { id: "m3", side: "left", avatar: "客", sender: "客服助手", content: "我已查看订单，商家正在补做一份菜品。预计 8 分钟后出餐，会同步通知骑手。", time: "14:19" }
    ],
    suggestions: [
      { title: "继续等待：预计 14:35 前送达" },
      { title: "申请补偿：订单完成后可领取延误券" }
    ],
    quickActions: ["催骑手", "申请退款", "联系商家", "投诉建议"]
  },
  onLoad(query) {
    const ticketId = String(query?.ticket_id || query?.request_id || "");
    this.setData({ ticketId });
    this.bootstrapTicket(ticketId);
  },
  async bootstrapTicket(ticketId) {
    ensurePreviewAuth();
    if (ticketId) {
      this.setData({ statusText: "已同步" });
      return;
    }
    try {
      const tickets = await getServiceTickets() as Array<Record<string, unknown>>;
      const first = tickets?.[0];
      if (first?.id) {
        this.setData({ ticketId: String(first.id), statusText: "已同步" });
      }
    } catch (_error) {
      this.setData({ statusText: "本地缓存" });
    }
  },
  handleSceneTap(event) {
    this.setData({ activeScene: String(event.currentTarget.dataset.scene || "订单售后") });
  },
  handleDraftInput(event) {
    const draft = String(event.detail.value || "");
    this.setData({ draft, riskTip: draft ? this.data.riskTip : "" });
  },
  async handleSend() {
    const message = String(this.data.draft || "").trim();
    if (!message) return;
    ensurePreviewAuth();
    const risk = detectCustomerServiceMessageRisk(message);
    if (risk.blocked) {
      this.setData({ draft: "", riskTip: risk.warning, statusText: "已拦截" });
      wx.showToast({ title: "敏感信息已拦截", icon: "none" });
      return;
    }
    const local = { id: `local_${Date.now()}`, side: "right", avatar: "我", sender: "我", content: message, time: "刚刚" };
    this.setData({ draft: "", riskTip: risk.warning, messages: this.data.messages.concat([local]) });
    try {
      let ticketId = this.data.ticketId;
      if (!ticketId) {
        const detail = await createServiceTicket({
          type: "delivery",
          category: this.data.activeScene,
          title: `${this.data.activeScene} · 在线客服`,
          content: message,
          related_order_id: this.data.orderCard.id,
          related_order_title: `${this.data.orderCard.shop} · ${this.data.orderCard.items}`,
          related_order_status: this.data.orderCard.status,
          severity: "较严重"
        }) as { ticket?: { id?: string } };
        ticketId = detail.ticket?.id || "";
        this.setData({ ticketId });
      } else {
        await addServiceTicketEvent(ticketId, { title: "用户补充", message, status: "active" });
      }
      this.setData({ statusText: "已同步" });
    } catch (error) {
      if (isRiskControlError(error)) {
        this.setData({
          riskTip: "消息触发客服安全风控，已停止同步。请删除支付密码、验证码或银行卡号后再发送。",
          statusText: "已拦截"
        });
        wx.showToast({ title: "敏感信息已拦截", icon: "none" });
        return;
      }
      this.setData({ statusText: "本地缓存" });
    }
  },
  handleQuick(event) {
    const text = String(event.currentTarget.dataset.text || "");
    if (text === "投诉建议") {
      wx.navigateTo({ url: "/pages/feedback/complaint/index" });
      return;
    }
    this.setData({ draft: text });
  },
  handleTicket() {
    const id = this.data.ticketId || "st_preview_delivery";
    wx.navigateTo({ url: `/pages/service-ticket/detail/index?id=${id}` });
  },
  handleOrder() {
    wx.navigateTo({ url: "/pages/order/detail/index?id=ord_preview" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/messages/index" }) });
  }
});
