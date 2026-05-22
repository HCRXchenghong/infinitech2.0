import { setDevAuthToken, setPaymentPassword } from "../../../utils/api";

Page({
  data: {
    password: "",
    confirmPassword: ""
  },
  handlePasswordInput(event) {
    this.setData({ password: String(event.detail.value || "").slice(0, 6) });
  },
  handleConfirmInput(event) {
    this.setData({ confirmPassword: String(event.detail.value || "").slice(0, 6) });
  },
  async handleSubmitTap() {
    const password = String(this.data.password || "");
    const confirmPassword = String(this.data.confirmPassword || "");
    if (!/^\d{6}$/.test(password) || password !== confirmPassword) {
      wx.showToast({ title: "请输入一致的 6 位数字", icon: "none" });
      return;
    }
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      await setPaymentPassword(password);
      wx.showToast({ title: "已设置", icon: "success" });
      wx.navigateBack();
    } catch (_error) {
      wx.showToast({ title: "预览已保存", icon: "none" });
    }
  }
});
