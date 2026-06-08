import { ensurePreviewAuth, setPaymentPassword } from "../../../utils/api";

function dots(value: string) {
  return Array.from({ length: 6 }).map((_, index) => ({ active: index < value.length }));
}

Page({
  data: {
    step: "first",
    password: "",
    confirmPassword: "",
    passwordDots: dots(""),
    confirmDots: dots("")
  },
  handlePasswordInput(event) {
    const password = String(event.detail.value || "").replace(/\D/g, "").slice(0, 6);
    this.setData({ password, passwordDots: dots(password) });
  },
  handleConfirmInput(event) {
    const confirmPassword = String(event.detail.value || "").replace(/\D/g, "").slice(0, 6);
    this.setData({ confirmPassword, confirmDots: dots(confirmPassword) });
  },
  handleNextTap() {
    const password = String(this.data.password || "");
    if (!/^\d{6}$/.test(password)) {
      wx.showToast({ title: "请输入 6 位数字", icon: "none" });
      return;
    }
    if (/^(\d)\1{5}$/.test(password) || "0123456789".includes(password) || "9876543210".includes(password)) {
      wx.showModal({
        title: "安全提示",
        content: "请勿使用连续或重复数字。平台不会向任何人索要支付密码，连续输错将临时锁定余额支付。",
        showCancel: false
      });
      return;
    }
    this.setData({ step: "confirm" });
  },
  async handleSubmitTap() {
    const password = String(this.data.password || "");
    const confirmPassword = String(this.data.confirmPassword || "");
    if (!/^\d{6}$/.test(confirmPassword) || password !== confirmPassword) {
      wx.showToast({ title: "两次密码不一致", icon: "none" });
      return;
    }
    ensurePreviewAuth();
    try {
      await setPaymentPassword(password);
      wx.showToast({ title: "已设置", icon: "success" });
      wx.navigateBack();
    } catch (_error) {
      wx.showToast({ title: "预览已保存", icon: "none" });
    }
  },
  handleSafetyTap() {
    wx.showModal({
      title: "支付安全说明",
      content: "余额支付前会校验 6 位数字密码；平台不会向你索要验证码或支付密码，异常输错会临时锁定余额支付。",
      showCancel: false
    });
  },
  handleBack() {
    if (this.data.step === "confirm") {
      this.setData({ step: "first", confirmPassword: "", confirmDots: dots("") });
      return;
    }
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/wallet/index" }) });
  }
});
