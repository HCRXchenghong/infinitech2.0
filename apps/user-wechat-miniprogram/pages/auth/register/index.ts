import { activatePreviewAuth, getWechatLoginCode, loginWithWechatMini, registerWithPhone, sendPhoneVerificationCode } from "../../../utils/api";
import { switchTabRoute } from "../../../utils/navigation";

Page({
  data: {
    nickname: "",
    phone: "",
    inviteCode: "",
    code: "",
    password: "",
    confirmPassword: "",
    agreementAccepted: true,
    submitting: false,
    codeSending: false,
    wechatLoading: false,
    codeStatusText: "",
    wechatStatusText: ""
  },

  handleNicknameInput(event: any) {
    this.setData({ nickname: String(event.detail.value || "") });
  },

  handlePhoneInput(event: any) {
    this.setData({ phone: String(event.detail.value || "").slice(0, 11) });
  },

  handleInviteInput(event: any) {
    this.setData({ inviteCode: String(event.detail.value || "") });
  },

  handleCodeInput(event: any) {
    this.setData({ code: String(event.detail.value || "").slice(0, 6) });
  },

  handlePasswordInput(event: any) {
    this.setData({ password: String(event.detail.value || "") });
  },

  handleConfirmInput(event: any) {
    this.setData({ confirmPassword: String(event.detail.value || "") });
  },

  handleAgreementToggle() {
    this.setData({ agreementAccepted: !this.data.agreementAccepted });
  },

  async handleSendCode() {
    const phone = String(this.data.phone || "");
    if (!/^1\d{10}$/.test(phone)) {
      wx.showToast({ title: "请输入正确手机号", icon: "none" });
      return;
    }
    this.setData({ codeSending: true });
    try {
      const ticket = await sendPhoneVerificationCode(phone, "register");
      if (ticket?.dev_code) {
        this.setData({ code: String(ticket.dev_code).slice(0, 6) });
      }
      const cooldown = Number(ticket?.cooldown_seconds || 60);
      const statusText = ticket?.delivery_status === "dev_returned"
        ? `开发验证码已回填，${cooldown} 秒后可重发`
        : `验证码已发送，${cooldown} 秒后可重发`;
      this.setData({ codeStatusText: statusText });
      wx.showToast({ title: `验证码已发送至${ticket?.masked_phone || "手机"}`, icon: "none" });
    } catch (error) {
      const message = String((error as Error)?.message || "");
      const statusText = message.includes("rate") || message.includes("频") ? "发送太频繁，请稍后再试" : "验证码发送失败，请稍后重试";
      this.setData({ codeStatusText: statusText });
      wx.showToast({ title: statusText, icon: "none" });
    } finally {
      this.setData({ codeSending: false });
    }
  },

  async handleRegister() {
    const phone = String(this.data.phone || "");
    const password = String(this.data.password || "");
    if (!/^1\d{10}$/.test(phone)) {
      wx.showToast({ title: "请输入正确手机号", icon: "none" });
      return;
    }
    if (!/^\d{6}$/.test(String(this.data.code || ""))) {
      wx.showToast({ title: "请输入 6 位验证码", icon: "none" });
      return;
    }
    if (password.length < 6 || password !== this.data.confirmPassword) {
      wx.showToast({ title: "请确认密码一致", icon: "none" });
      return;
    }
    if (!this.data.agreementAccepted) {
      wx.showToast({ title: "请先同意用户协议", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      await registerWithPhone({
        phone,
        code: this.data.code,
        password,
        nickname: this.data.nickname || "悦享用户",
        invite_code: this.data.inviteCode,
        accepted_agreement: this.data.agreementAccepted
      });
      switchTabRoute("/pages/index/index");
    } catch (error) {
      if (activatePreviewAuth("user_1")) {
        wx.setStorageSync("authMode", "user");
        wx.setStorageSync("userNickname", this.data.nickname || "悦享用户");
        wx.setStorageSync("userPhone", phone);
        wx.showToast({ title: "已进入本地预览注册", icon: "none" });
        switchTabRoute("/pages/index/index");
        return;
      }
      wx.showToast({ title: authErrorText(error, "注册失败，请检查验证码"), icon: "none" });
    } finally {
      this.setData({ submitting: false });
    }
  },

  async handleWechatRegister() {
    this.setData({ wechatLoading: true });
    try {
      const code = await getWechatLoginCode();
      await loginWithWechatMini(code, {
        nickname: this.data.nickname || "悦享用户"
      });
      wx.setStorageSync("authMode", "user");
      this.setData({ wechatStatusText: "微信注册/登录成功" });
      switchTabRoute("/pages/index/index");
    } catch (error) {
      if (activatePreviewAuth("user_1")) {
        wx.setStorageSync("authMode", "user");
        this.setData({ wechatStatusText: "当前为本地 API，已进入预览注册" });
        wx.showToast({ title: "已进入本地预览注册", icon: "none" });
        switchTabRoute("/pages/index/index");
        return;
      }
      const message = authErrorText(error, "微信注册失败，请稍后重试");
      this.setData({ wechatStatusText: message });
      wx.showToast({ title: message, icon: "none" });
    } finally {
      this.setData({ wechatLoading: false });
    }
  },

  goLogin() {
    wx.navigateBack({
      fail() {
        wx.navigateTo({ url: "/pages/auth/login/index" });
      }
    });
  }
});

function authErrorText(error: unknown, fallback: string) {
  const message = String((error as Error)?.message || "").trim();
  if (!message) return fallback;
  if (message.includes("WECHAT_LOGIN_REJECTED") || message.includes("invalid code")) return "微信登录凭证已失效，请重试";
  if (message.includes("INVALID_CREDENTIALS")) return "验证码不正确或已过期";
  if (message.length > 18) return fallback;
  return message;
}
