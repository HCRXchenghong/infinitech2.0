import { createRedPacket, ensurePreviewAuth } from "../../../utils/api";

Page({
  data: {
    type: "random",
    amount: "30.00",
    count: "6",
    message: "一起拼单，吃得开心！",
    averageText: "¥5.00",
    totalText: "30.00",
    submitting: false
  },
  handleTypeTap(event) {
    this.setData({ type: String(event.currentTarget.dataset.type || "random") });
  },
  handleAmountInput(event) {
    this.setData({ amount: String(event.detail.value || "") });
    this.refreshTotal();
  },
  handleCountInput(event) {
    this.setData({ count: String(event.detail.value || "") });
    this.refreshTotal();
  },
  handleMessageInput(event) {
    this.setData({ message: String(event.detail.value || "") });
  },
  refreshTotal() {
    const amount = Number(this.data.amount || 0);
    const count = Math.max(1, Number(this.data.count || 1));
    this.setData({
      totalText: amount.toFixed(2),
      averageText: `¥${(amount / count).toFixed(2)}`
    });
  },
  async handleSubmit() {
    ensurePreviewAuth();
    const amountFen = Math.round(Number(this.data.amount || 0) * 100);
    const quantity = Math.max(1, Number(this.data.count || 1));
    if (amountFen <= 0) {
      wx.showToast({ title: "请输入红包金额", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      const detail = await createRedPacket({
        scene: "group_chat",
        target_id: "merchant_blue_sea",
        type: this.data.type,
        total_amount_fen: amountFen,
        quantity,
        payment_method: "balance",
        message: this.data.message
      }) as { packet?: { id?: string } };
      const packetId = detail.packet?.id || "";
      wx.showToast({ title: "红包已发出", icon: "success" });
      setTimeout(() => wx.navigateTo({ url: `/pages/red-packet/detail/index?id=${packetId}` }), 500);
    } catch (error) {
      const message = String((error as Error)?.message || "");
      if (message.includes("insufficient")) {
        wx.showToast({ title: "余额不足，请先充值", icon: "none" });
        return;
      }
      wx.showToast({ title: "红包预览已生成", icon: "none" });
      wx.navigateTo({ url: "/pages/red-packet/detail/index" });
    } finally {
      this.setData({ submitting: false });
    }
  },
  handleRecords() {
    wx.navigateTo({ url: "/pages/red-packet/detail/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/messages/merchant-group/index" }) });
  }
});
