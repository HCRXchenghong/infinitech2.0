import { ensurePreviewAuth, getInviteSummary } from "../../utils/api";

Page({
  data: {
    inviteCode: "YXES-8K29",
    rewardText: "好友首单完成后，双方各得一张优惠券",
    shareTitle: "来悦享e食一起点餐",
    sharePath: "/pages/auth/register/index?invite=YXES-8K29",
    records: [
      { id: "inv_1", friendName: "李四", status: "首单待完成", rewardText: "待发放" },
      { id: "inv_2", friendName: "王五", status: "奖励已到账", rewardText: "+1 张优惠券" }
    ],
    riskTip: "同一手机号仅计一次 · 异常订单不参与奖励"
  },
  onShow() {
    this.loadInviteSummary();
  },
  async loadInviteSummary() {
    ensurePreviewAuth();
    try {
      const summary = await getInviteSummary() as Record<string, unknown>;
      this.setData({
        inviteCode: String(summary.invite_code || "YXES-8K29"),
        rewardText: String(summary.reward_text || this.data.rewardText),
        shareTitle: String(summary.share_title || this.data.shareTitle),
        sharePath: String(summary.share_path || this.data.sharePath),
        riskTip: String(summary.abuse_risk_tip || this.data.riskTip),
        records: (summary.records as Array<Record<string, unknown>> || []).map((item) => ({
          id: String(item.id || Date.now()),
          friendName: String(item.friend_name || "好友"),
          status: String(item.status || "进行中"),
          rewardText: String(item.reward_text || "待发放")
        }))
      });
    } catch (_error) {
      wx.showToast({ title: "邀请缓存已加载", icon: "none" });
    }
  },
  onShareAppMessage() {
    return {
      title: this.data.shareTitle,
      path: this.data.sharePath
    };
  },
  handleCopy() {
    wx.setClipboardData({ data: this.data.inviteCode });
  },
  handleRecordTap() {
    wx.showToast({ title: "邀请记录已同步", icon: "none" });
  },
  handlePosterTap() {
    wx.showToast({ title: "分享图生成中", icon: "none" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});
