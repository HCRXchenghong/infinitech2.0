import { claimUserCoupon, ensurePreviewAuth, getAuthToken, getChatMessages, getChatThreadMembers, getChatThreadOverview, getChatThreadPreference, getCurrentUserId, getRealtimeSocketUrl, getUserCoupons, markChatThreadRead, saveChatThreadPreference, sendChatMessage, syncChatMessages } from "../../../utils/api";

Page({
  data: {
    threadId: "merchant_blue_sea",
    title: "蓝海餐厅商户群",
    icon: "店",
    groupSummary: "326 人已加入 · 新用户默认静音",
    groupNotice: "群内优惠每日 10:00 更新，重要通知会保留在消息中心。",
    groupSettingsText: "群设置",
    realtimeStatus: "实时连接中",
    muted: false,
    draft: "",
    lastMessageId: "",
    memberCount: 326,
    memberPreview: [],
    messages: [
      { id: "m1", side: "left", avatar: "店", sender: "蓝海餐厅", time: "14:08", content: "今晚 20:00 前可领商户群券，下单可叠加平台满减。" },
      { id: "m2", side: "left", avatar: "小", sender: "小林", time: "14:10", content: "午饭拼单还差一份，有人一起吗？" },
      { id: "m3", side: "left", avatar: "阿", sender: "阿杰", time: "14:11", content: "我也拼一份，12:20 前下单可以吗？" }
    ],
    couponCard: {
      amountText: "¥8",
      thresholdText: "满30可用",
      title: "蓝海餐厅商户群券",
      subtitle: "有效期至 今天 20:00",
      buttonText: "立即领取"
    },
    couponClaimed: false,
    systemTip: "进群后可领取 1 张商户群券，下单时可与平台满减叠加。",
    orderCard: {
      shop: "蓝海餐厅",
      items: "招牌牛肉饭等 3 件",
      status: "待提交"
    },
    redPacket: {
      title: "商户群红包",
      subtitle: "限 3 分钟领取"
    }
  },
  onLoad(options) {
    const threadId = String(options?.thread_id || "merchant_blue_sea");
    const isOfficial = String(options?.type || "") === "official";
    this.setData({
      threadId,
      icon: isOfficial ? "官" : "店",
      title: isOfficial ? "悦享e食官方群" : "蓝海餐厅商户群",
      groupSummary: isOfficial ? "1286 人已加入 · 新用户默认静音" : "326 人已加入 · 新用户默认静音",
      groupNotice: isOfficial ? "重要通知会同步保留在消息中心，常规讨论默认静默。" : "群内优惠每日 10:00 更新，重要通知会保留在消息中心。",
      muted: isOfficial,
      memberCount: isOfficial ? 1286 : 326
    });
    this.loadThreadOverview(threadId);
    this.loadThreadPreference(threadId);
    this.loadMessages(threadId);
    this.connectRealtime(threadId);
  },
  onShow() {
    if (!this.data.threadId) return;
    this.loadThreadOverview(this.data.threadId);
    this.loadThreadPreference(this.data.threadId);
    this.syncCouponState();
  },
  onUnload() {
    const socketTask = (this as any).realtimeSocket;
    if (socketTask && typeof socketTask.close === "function") {
      socketTask.close({});
    }
    (this as any).realtimeSocket = null;
  },
  async loadThreadOverview(threadId) {
    ensurePreviewAuth();
    try {
      const overview = await getChatThreadOverview(threadId) as Record<string, unknown>;
      const memberPayload = await getChatThreadMembers(threadId) as Record<string, unknown>;
      const members = Array.isArray(memberPayload?.members) ? memberPayload.members.slice(0, 5) : [];
      this.setData({
        title: String(overview.title || this.data.title),
        icon: String(overview.icon || this.data.icon),
        groupSummary: String(overview.summary || this.data.groupSummary),
        groupNotice: String(overview.announcement || this.data.groupNotice),
        groupSettingsText: String(overview.settings_text || "群设置"),
        memberCount: Number(overview.member_count || this.data.memberCount),
        memberPreview: members,
        muted: typeof overview.muted === "boolean" ? Boolean(overview.muted) : this.data.muted
      });
    } catch (_error) {
      // Keep seeded preview content.
    }
  },
  async loadMessages(threadId) {
    ensurePreviewAuth();
    try {
      const sync = await syncChatMessages(threadId, this.data.lastMessageId, true) as Record<string, unknown>;
      const messages = Array.isArray(sync.messages) ? sync.messages as Array<Record<string, unknown>> : [];
      if (!messages.length) return;
      const baseMessages = this.data.lastMessageId ? this.data.messages : [];
      const merged = this.mergeMessages(baseMessages, messages.map((item, index) => this.mapMessage(item, index)));
      this.setData({
        messages: merged,
        lastMessageId: String(sync.next_cursor || sync.last_message_id || messages[messages.length - 1]?.id || "")
      });
    } catch (_error) {
      this.loadLegacyMessages(threadId);
    }
  },
  async syncCouponState() {
    ensurePreviewAuth();
    try {
      const summary = await getUserCoupons() as Record<string, unknown>;
      const coupons = Array.isArray(summary.coupons) ? summary.coupons as Array<Record<string, unknown>> : [];
      const claimed = coupons.some((item) => String(item.id || "").startsWith("coupon_group_8_"));
      this.setData({
        couponClaimed: claimed,
        systemTip: claimed ? "你已领取 1 张商户群券，可在红包优惠中查看。" : "进群后可领取 1 张商户群券，下单时可与平台满减叠加。",
        couponCard: {
          ...this.data.couponCard,
          buttonText: claimed ? "已领取" : "立即领取"
        }
      });
    } catch (_error) {
      // Keep seeded preview state.
    }
  },
  async loadThreadPreference(threadId) {
    ensurePreviewAuth();
    try {
      const preference = await getChatThreadPreference(threadId) as Record<string, unknown>;
      this.applyThreadPreference(preference);
    } catch (_error) {
      this.applyThreadPreference({ muted: this.data.muted });
    }
  },
  async loadLegacyMessages(threadId) {
    try {
      const messages = await getChatMessages(threadId) as Array<Record<string, unknown>>;
      if (!messages.length) return;
      const mapped = messages.map((item, index) => this.mapMessage(item, index));
      const lastMessageId = String(messages[messages.length - 1]?.id || "");
      this.setData({ messages: mapped, lastMessageId });
      if (lastMessageId) {
        await markChatThreadRead(threadId, lastMessageId);
      }
    } catch (_error) {
      this.applyThreadPreference({ muted: this.data.muted });
    }
  },
  applyThreadPreference(preference) {
    const muted = Boolean(preference?.muted);
    this.setData({
      muted
    });
  },
  mapMessage(item, index) {
    const sender = String(item.sender || "群成员");
    return {
      id: String(item.id || `msg_${index}`),
      side: sender === "我" ? "right" : "left",
      avatar: avatarInitial(sender),
      sender,
      time: String(item.created_at || "").slice(11, 16) || "刚刚",
      content: String(item.content || "")
    };
  },
  mergeMessages(current, incoming) {
    const byId = new Map();
    current.concat(incoming).forEach((item) => {
      byId.set(item.id, item);
    });
    return Array.from(byId.values());
  },
  connectRealtime(threadId) {
    const wxAny = wx as unknown as { connectSocket?: (options: Record<string, unknown>) => unknown };
    if (typeof wxAny.connectSocket !== "function") {
      this.setData({ realtimeStatus: "实时待接入" });
      return;
    }
    const token = getAuthToken();
    const userId = getCurrentUserId();
    const socketTask = wxAny.connectSocket({
      url: getRealtimeSocketUrl("/ws", { thread_id: threadId, user_id: userId }),
      header: token ? { Authorization: `Bearer ${token}` } : {}
    }) as {
      onOpen?: (callback: () => void) => void;
      onMessage?: (callback: (event: { data: string | ArrayBuffer }) => void) => void;
      onClose?: (callback: () => void) => void;
      onError?: (callback: () => void) => void;
      close?: (options: Record<string, unknown>) => void;
    };
    (this as any).realtimeSocket = socketTask;
    socketTask.onOpen?.(() => this.setData({ realtimeStatus: "实时在线" }));
    socketTask.onClose?.(() => this.setData({ realtimeStatus: "实时断开" }));
    socketTask.onError?.(() => this.setData({ realtimeStatus: "实时重连中" }));
    socketTask.onMessage?.((event) => {
      const text = typeof event.data === "string" ? event.data : "";
      this.handleRealtimeMessage(text);
    });
  },
  handleRealtimeMessage(text) {
    let event;
    try {
      event = JSON.parse(text || "{}");
    } catch (_error) {
      return;
    }
    if (event.topic !== "message.sent" || event.payload?.thread_id !== this.data.threadId) return;
    const mapped = this.mapMessage(event.payload, this.data.messages.length);
    const merged = this.mergeMessages(this.data.messages, [mapped]);
    this.setData({
      messages: merged,
      lastMessageId: mapped.id || this.data.lastMessageId,
      realtimeStatus: "实时在线"
    });
  },
  handleDraftInput(event) {
    this.setData({ draft: String(event.detail.value || "") });
  },
  async handleSend() {
    const content = String(this.data.draft || "").trim();
    if (!content) return;
    ensurePreviewAuth();
    const localMessage = {
      id: `local_${Date.now()}`,
      side: "right",
      avatar: "我",
      sender: "我",
      time: "刚刚",
      content
    };
    this.setData({ draft: "", messages: this.data.messages.concat([localMessage]) });
    try {
      const sent = await sendChatMessage(this.data.threadId, { sender: "我", content, message_type: "text" }) as Record<string, unknown>;
      this.setData({
        lastMessageId: String(sent.id || this.data.lastMessageId)
      });
    } catch (_error) {
      wx.showToast({ title: "消息已暂存本地", icon: "none" });
    }
  },
  async handleToggleMute() {
    ensurePreviewAuth();
    const nextMuted = !this.data.muted;
    try {
      const preference = await saveChatThreadPreference(this.data.threadId, { muted: nextMuted }) as Record<string, unknown>;
      this.applyThreadPreference(preference);
      wx.showToast({ title: nextMuted ? "已设为免打扰" : "已恢复提醒", icon: "none" });
    } catch (_error) {
      wx.showToast({ title: "设置未保存", icon: "none" });
    }
  },
  handleOpenGroupSettings() {
    wx.navigateTo({ url: `/pages/messages/group-settings/index?thread_id=${encodeURIComponent(this.data.threadId)}` });
  },
  async handleClaimCoupon() {
    if (this.data.couponClaimed) {
      wx.navigateTo({ url: "/pages/coupons/index" });
      return;
    }
    ensurePreviewAuth();
    try {
      await claimUserCoupon("GROUP8");
      this.setData({
        couponClaimed: true,
        systemTip: "你已领取 1 张商户群券，可在红包优惠中查看。",
        couponCard: {
          ...this.data.couponCard,
          buttonText: "已领取"
        }
      });
      wx.showToast({ title: "商户群券已到账", icon: "success" });
    } catch (error) {
      const message = String((error as Error)?.message || "");
      wx.showToast({ title: message.includes("membership") ? "请先保持商户群成员身份" : "领取失败", icon: "none" });
    }
  },
  handleSendRedPacket() {
    wx.navigateTo({ url: "/pages/red-packet/send/index" });
  },
  handleOpenRedPacket() {
    wx.navigateTo({ url: "/pages/red-packet/detail/index" });
  },
  handleGoOrder() {
    wx.navigateTo({ url: "/pages/order/confirm/index" });
  },
  handleCoupon() {
    wx.navigateTo({ url: "/pages/coupons/index" });
  },
  handleImage() {
    wx.showToast({ title: "可选择聊天图片", icon: "none" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/messages/index" }) });
  }
});

function avatarInitial(value: string) {
  const text = String(value || "").trim();
  return text ? text.slice(0, 1) : "群";
}
