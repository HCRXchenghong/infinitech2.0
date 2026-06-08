import { ensurePreviewAuth, getChatThreadMembers, getChatThreadMembership, getChatThreadOverview, leaveChatThread, saveChatThreadPreference } from "../../../utils/api";

Page({
  data: {
    threadId: "merchant_blue_sea",
    title: "群设置",
    icon: "店",
    groupTitle: "蓝海餐厅商户群",
    summary: "326 人已加入 · 新用户默认静音",
    announcement: "群内优惠每日 10:00 更新，重要通知会保留在消息中心。",
    memberCount: 326,
    muted: false,
    joined: true,
    canLeave: true,
    members: []
  },
  onLoad(options) {
    this.setData({
      threadId: String(options?.thread_id || "merchant_blue_sea")
    });
  },
  onShow() {
    this.loadData();
  },
  async loadData() {
    ensurePreviewAuth();
    try {
      const membership = await getChatThreadMembership(this.data.threadId) as Record<string, unknown>;
      const overview = await getChatThreadOverview(this.data.threadId) as Record<string, unknown>;
      const memberPayload = await getChatThreadMembers(this.data.threadId) as Record<string, unknown>;
      const members = Array.isArray(memberPayload?.members) ? memberPayload.members : [];
      this.setData({
        groupTitle: String(overview.title || this.data.groupTitle),
        icon: String(overview.icon || this.data.icon),
        summary: String(membership.summary || overview.summary || this.data.summary),
        announcement: String(overview.announcement || this.data.announcement),
        memberCount: Number(membership.member_count || overview.member_count || this.data.memberCount),
        muted: typeof membership.muted === "boolean" ? Boolean(membership.muted) : Boolean(overview.muted),
        joined: Boolean(membership.joined),
        canLeave: Boolean(membership.can_leave),
        members
      });
    } catch (_error) {
      wx.showToast({ title: "群资料加载失败", icon: "none" });
    }
  },
  async handleMuteSwitch(event) {
    const muted = Boolean(event.detail.value);
    await this.saveMutePreference(muted);
  },
  async handleToggleMute() {
    await this.saveMutePreference(!this.data.muted);
  },
  async saveMutePreference(muted) {
    ensurePreviewAuth();
    try {
      const preference = await saveChatThreadPreference(this.data.threadId, { muted }) as Record<string, unknown>;
      this.setData({ muted: Boolean(preference?.muted) });
      wx.showToast({ title: muted ? "已设为免打扰" : "已恢复提醒", icon: "none" });
    } catch (_error) {
      wx.showToast({ title: "设置未保存", icon: "none" });
    }
  },
  async handleLeaveGroup() {
    if (!this.data.canLeave) return;
    ensurePreviewAuth();
    try {
      await leaveChatThread(this.data.threadId);
      wx.showToast({ title: "已退出商户群", icon: "none" });
      setTimeout(() => {
        wx.redirectTo({ url: "/pages/shop/detail/index?id=shop_1" });
      }, 320);
    } catch (_error) {
      wx.showToast({ title: "退出失败", icon: "none" });
    }
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: `/pages/messages/merchant-group/index?thread_id=${encodeURIComponent(this.data.threadId)}` }) });
  }
});
