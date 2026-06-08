import { ensurePreviewAuth, getUserProfileOverview } from "../../utils/api";
import { backOrSwitchTab, openRoute } from "../../utils/navigation";

Page({
  data: {
    nickname: "张三",
    phone: "13800000000",
    avatarInitial: "张三",
    authText: "已登录",
    memberLevel: "V2",
    memberTitle: "美食达人",
    creditText: "信用良好",
    growthText: "距 V3 还需 1320 成长值",
    growthPercent: 67,
    topStats: [
      { icon: "券", title: "本月已省", value: "¥36" },
      { icon: "票", title: "专属券", value: "2 张" },
      { icon: "速", title: "积分加速", value: "1.2x" }
    ],
    assetStats: [
      { icon: "钱", title: "余额", value: "¥128.50", route: "/pages/wallet/index" },
      { icon: "券", title: "优惠券", value: "6", route: "/pages/coupons/index" },
      { icon: "分", title: "积分", value: "2680", route: "/pages/member-points/index" },
      { icon: "包", title: "红包", value: "3", route: "/pages/coupons/index?tab=red_packet" }
    ],
    orderShortcuts: [
      { icon: "付", title: "待支付", badge: "", route: "/pages/order/list/index?status=pending_payment" },
      { icon: "送", title: "进行中", badge: "2", route: "/pages/order/list/index?status=in_progress" },
      { icon: "评", title: "待评价", badge: "3", route: "/pages/order/list/index?status=review" },
      { icon: "售", title: "售后", badge: "", route: "/pages/after-sales/index" }
    ],
    serviceGrid: [
      { icon: "址", title: "收货地址", route: "/pages/address/list/index" },
      { icon: "铃", title: "通知偏好", route: "/pages/notification-preferences/index" },
      { icon: "星", title: "我的评价", route: "/pages/order/review/index" },
      { icon: "藏", title: "收藏店铺", route: "/pages/shop/list/index" },
      { icon: "邀", title: "邀请好友", route: "/pages/invite-friends/index" },
      { icon: "分", title: "会员积分", route: "/pages/member-points/index" },
      { icon: "客", title: "客服与反馈", route: "/pages/customer-service/chat/index" },
      { icon: "合", title: "商家合作", route: "/pages/feedback/complaint/index?type=merchant" }
    ],
    securityRows: [
      { icon: "密", title: "支付密码", value: "已设置", route: "/pages/wallet/payment-password/index" },
      { icon: "认", title: "实名认证", value: "已完成", route: "/pages/profile/index" }
    ]
  },
  onShow() {
    this.loadProfile();
  },
  async loadProfile() {
    ensurePreviewAuth();
    try {
      const overview = await getUserProfileOverview() as Record<string, unknown>;
      const orderStats = overview.order_stats as Record<string, number> || {};
      const nextLevelGrowth = Number(overview.next_level_growth || 0);
      const growthValue = Number(overview.growth_value || 0);
      const totalGrowth = growthValue + nextLevelGrowth;
      this.setData({
        nickname: String(overview.nickname || "张三"),
        phone: String(overview.phone || "13800000000"),
        avatarInitial: String(overview.avatar_initial || "张三"),
        authText: "已同步",
        memberLevel: String(overview.membership_level || "V2").replace("silver", "V2"),
        memberTitle: String(overview.membership_title || "美食达人"),
        creditText: String(overview.credit_text || "信用良好"),
        growthText: `距 V3 还需 ${nextLevelGrowth || 1320} 成长值`,
        growthPercent: totalGrowth > 0 ? Math.round((growthValue / totalGrowth) * 100) : 67,
        topStats: [
          { icon: "券", title: "本月已省", value: `¥${(Number(overview.savings_fen || 0) / 100).toFixed(0)}` },
          { icon: "票", title: "专属券", value: "2 张" },
          { icon: "速", title: "积分加速", value: "1.2x" }
        ],
        assetStats: [
          { icon: "钱", title: "余额", value: `¥${(Number(overview.wallet_balance_fen || 0) / 100).toFixed(2)}`, route: "/pages/wallet/index" },
          { icon: "券", title: "优惠券", value: String(overview.coupon_count || 0), route: "/pages/coupons/index" },
          { icon: "分", title: "积分", value: String(overview.points || 0), route: "/pages/member-points/index" },
          { icon: "包", title: "红包", value: String(overview.red_packet_count || 0), route: "/pages/coupons/index?tab=red_packet" }
        ],
        orderShortcuts: [
          { icon: "付", title: "待支付", badge: orderStats.pending_payment ? String(orderStats.pending_payment) : "", route: "/pages/order/list/index?status=pending_payment" },
          { icon: "送", title: "进行中", badge: orderStats.in_progress ? String(orderStats.in_progress) : "2", route: "/pages/order/list/index?status=in_progress" },
          { icon: "评", title: "待评价", badge: orderStats.pending_review ? String(orderStats.pending_review) : "3", route: "/pages/order/list/index?status=review" },
          { icon: "售", title: "售后", badge: orderStats.after_sales ? String(orderStats.after_sales) : "", route: "/pages/after-sales/index" }
        ],
        securityRows: [
          { icon: "密", title: "支付密码", value: overview.payment_password_status === "set" ? "已设置" : "未设置", route: "/pages/wallet/payment-password/index" },
          { icon: "认", title: "实名认证", value: overview.verified ? "已完成" : "待完善", route: "/pages/profile/index" }
        ]
      });
    } catch (_error) {
      this.setData({ authText: wx.getStorageSync("authToken") ? "已登录" : "游客" });
    }
  },
  handleNavigate(event) {
    openRoute(String(event.currentTarget.dataset.route || "/pages/index/index"));
  },
  handleBack() {
    backOrSwitchTab("/pages/index/index");
  }
});
