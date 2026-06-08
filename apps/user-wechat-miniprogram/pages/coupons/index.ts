import { claimUserCoupon, ensurePreviewAuth, getChatThreadMembership, getUserCoupons, joinChatThread } from "../../utils/api";

Page({
  data: {
    statusText: "已同步",
    availableCount: 6,
    redPacketCount: 3,
    expiringCount: 2,
    selectedTab: "coupon",
    tabs: [
      { id: "coupon", title: "优惠券", active: "active" },
      { id: "red_packet", title: "红包", active: "" },
      { id: "used", title: "已使用", active: "" },
      { id: "expired", title: "已过期", active: "" }
    ],
    filters: [
      { id: "all", title: "全部", active: "active" },
      { id: "外卖", title: "外卖", active: "" },
      { id: "团购", title: "团购", active: "" },
      { id: "买药", title: "买药", active: "" },
      { id: "跑腿", title: "跑腿", active: "" }
    ],
    groupCouponEntry: {
      joined: false,
      claimed: false,
      summary: "加入商户群后可领取外卖专享券",
      buttonText: "加入商户群领券"
    },
    allCoupons: [],
    coupons: [
      { id: "coupon_platform_15", kind: "platform", amount: "¥15", threshold: "满50可用", title: "平台外卖通用券", subtitle: "蓝海餐厅、晴川咖啡等可用", expireText: "明天 23:59 到期", source: "平台券", accentClass: "blue", buttonText: "去使用" }
    ]
  },
  onLoad(query) {
    const tab = String(query?.tab || "coupon");
    this.setData({ selectedTab: tab });
    this.loadCoupons();
  },
  onShow() {
    this.loadGroupCouponState();
  },
  async loadCoupons() {
    ensurePreviewAuth();
    try {
      const summary = await getUserCoupons() as Record<string, unknown>;
      const coupons = (summary.coupons as Array<Record<string, unknown>> || []).map(couponFromApi);
      const claimedGroupCoupon = coupons.some((item) => String(item.id || "").startsWith("coupon_group_8_"));
      this.setData({
        availableCount: Number(summary.available_count || coupons.length),
        redPacketCount: Number(summary.red_packet_count || 0),
        expiringCount: Number(summary.expiring_count || 0),
        allCoupons: coupons,
        coupons,
        statusText: "已同步",
        groupCouponEntry: {
          ...this.data.groupCouponEntry,
          claimed: claimedGroupCoupon,
          buttonText: claimedGroupCoupon ? "去使用" : this.data.groupCouponEntry.buttonText
        },
        tabs: this.data.tabs.map((item) => ({ ...item, active: item.id === this.data.selectedTab ? "active" : "" }))
      });
      await this.loadGroupCouponState(claimedGroupCoupon);
    } catch (_error) {
      this.setData({ statusText: "离线缓存" });
    }
  },
  async loadGroupCouponState(claimedOverride?: boolean) {
    ensurePreviewAuth();
    try {
      const membership = await getChatThreadMembership("merchant_blue_sea") as Record<string, unknown>;
      const joined = Boolean(membership.joined);
      const claimed = typeof claimedOverride === "boolean" ? claimedOverride : Boolean(this.data.groupCouponEntry.claimed);
      this.setData({
        groupCouponEntry: {
          joined,
          claimed,
          summary: claimed ? "商户群券已到账，可在下单时直接使用。" : joined ? "已在商户群内，可立即领取外卖专享券。" : String(membership.summary || "加入商户群后可领取外卖专享券"),
          buttonText: claimed ? "去使用" : joined ? "立即领取" : "加入商户群领券"
        }
      });
    } catch (_error) {
      // Keep preview copy.
    }
  },
  handleTabTap(event) {
    const selectedTab = String(event.currentTarget.dataset.id || "coupon");
    this.setData({
      selectedTab,
      tabs: this.data.tabs.map((item) => ({ ...item, active: item.id === selectedTab ? "active" : "" }))
    });
  },
  handleFilterTap(event) {
    const filter = String(event.currentTarget.dataset.id || "all");
    const allCoupons = this.data.allCoupons.length ? this.data.allCoupons : this.data.coupons;
    this.setData({
      filters: this.data.filters.map((item) => ({ ...item, active: item.id === filter ? "active" : "" })),
      coupons: filter === "all" ? allCoupons : allCoupons.filter((item) => item.scope === filter)
    });
  },
  async handleRedeemCode() {
    ensurePreviewAuth();
    try {
      await claimUserCoupon("YXES2026");
      await this.loadCoupons();
      wx.showToast({ title: "兑换成功", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "兑换码已记录", icon: "none" });
    }
  },
  async handleGroupCouponAction() {
    ensurePreviewAuth();
    try {
      const entry = this.data.groupCouponEntry;
      if (entry.claimed) {
        wx.navigateTo({ url: "/pages/shop/detail/index?id=shop_1" });
        return;
      }
      if (!entry.joined) {
        await joinChatThread("merchant_blue_sea");
      }
      await claimUserCoupon("GROUP8");
      await this.loadCoupons();
      wx.showToast({ title: entry.joined ? "商户群券已到账" : "已入群并领取商户券", icon: "success" });
    } catch (error) {
      const message = String((error as Error)?.message || "");
      wx.showToast({ title: message.includes("membership") ? "请先加入商户群" : "领取失败", icon: "none" });
    }
  },
  handleUseTap(event) {
    const route = String(event.currentTarget.dataset.route || "/pages/search/index");
    wx.navigateTo({ url: route });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});

function couponFromApi(coupon: Record<string, unknown>) {
  const amountFen = Number(coupon.amount_fen || 0);
  const thresholdFen = Number(coupon.threshold_fen || 0);
  const scope = String(coupon.scope || "外卖");
  const accent = String(coupon.accent_color || "#007aff");
  return {
    id: String(coupon.id || Date.now()),
    kind: String(coupon.kind || "platform"),
    scope,
    amount: `¥${(amountFen / 100).toFixed(0)}`,
    threshold: thresholdFen > 0 ? `满${(thresholdFen / 100).toFixed(0)}可用` : "无门槛",
    title: String(coupon.title || "优惠券"),
    subtitle: String(coupon.subtitle || "悦享e食可用"),
    expireText: expiryText(String(coupon.expires_at || "")),
    source: String(coupon.source || "平台券"),
    accentClass: accent.includes("ff") || scope === "团购" ? "orange" : scope === "买药" ? "green" : "blue",
    buttonText: String(coupon.button_text || "去使用"),
    route: scope === "买药" ? "/pages/medicine/home/index" : scope === "团购" ? "/pages/shop/detail/index?id=shop_1" : "/pages/search/index"
  };
}

function expiryText(value: string) {
  if (!value) return "本周内有效";
  return `${value.slice(5, 10)} 到期`;
}
