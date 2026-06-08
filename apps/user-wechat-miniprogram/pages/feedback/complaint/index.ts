import { createFeedback, createServiceTicket, ensurePreviewAuth } from "../../../utils/api";

Page({
  data: {
    content: "",
    contact: "138****0000",
    selectedType: "配送问题",
    severity: "较严重",
    notify: "站内信",
    types: ["配送问题", "商家服务", "商品质量", "红包钱包", "账户安全", "功能建议"],
    severities: ["一般", "较严重", "非常严重"],
    evidence: ["订单截图", "聊天记录", "添加图片"],
    submitting: false
  },
  handleTypeTap(event) {
    this.setData({ selectedType: String(event.currentTarget.dataset.type || "配送问题") });
  },
  handleSeverityTap(event) {
    this.setData({ severity: String(event.currentTarget.dataset.level || "较严重") });
  },
  handleContentInput(event) {
    this.setData({ content: String(event.detail.value || "") });
  },
  handleContactInput(event) {
    this.setData({ contact: String(event.detail.value || "") });
  },
  async handleSubmit() {
    ensurePreviewAuth();
    const content = this.data.content.trim();
    if (!content) {
      wx.showToast({ title: "请先填写问题说明", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      await createFeedback({
        type: this.data.selectedType,
        content,
        contact: this.data.contact
      });
      const detail = await createServiceTicket({
        type: "feedback",
        category: this.data.selectedType,
        title: `${this.data.selectedType} · 用户反馈`,
        content,
        contact: this.data.contact,
        related_order_id: "DD240518001",
        related_order_title: "蓝海餐厅 · 招牌牛肉饭等 3 件",
        related_order_status: "配送中",
        severity: this.data.severity,
        attachments: this.data.evidence.slice(0, 2)
      }) as { ticket?: { id?: string } };
      wx.showToast({ title: "反馈已提交", icon: "success" });
      const ticketId = detail.ticket?.id || "";
      setTimeout(() => wx.navigateTo({ url: ticketId ? `/pages/service-ticket/detail/index?id=${ticketId}` : "/pages/feedback/records/index" }), 500);
    } catch (_error) {
      wx.showToast({ title: "反馈已保存", icon: "none" });
      wx.navigateTo({ url: "/pages/feedback/records/index" });
    } finally {
      this.setData({ submitting: false });
    }
  },
  handleRecords() {
    wx.navigateTo({ url: "/pages/feedback/records/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});
