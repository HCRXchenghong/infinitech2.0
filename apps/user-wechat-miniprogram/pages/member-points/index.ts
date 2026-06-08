import { checkInPoints, ensurePreviewAuth, getUserPointsSummary } from "../../utils/api";

Page({
  data: {
    points: 2680,
    nickname: "张三",
    levelName: "美食达人",
    growthText: "距 V3 还需 1320 成长值",
    growthPercent: 67,
    benefits: [
      { icon: "券", title: "专属优惠券", status: "已解锁", lockedClass: "" },
      { icon: "抵", title: "积分抵扣", status: "已解锁", lockedClass: "" },
      { icon: "礼", title: "生日礼", status: "V3 解锁", lockedClass: "locked" },
      { icon: "客", title: "优先客服", status: "V3 解锁", lockedClass: "locked" }
    ],
    tasks: [
      { id: "order", icon: "单", title: "完成订单", subtitle: "每完成一笔订单", reward: "+30 积分", actionText: "去下单", route: "/pages/index/index" },
      { id: "review", icon: "评", title: "评价订单", subtitle: "每完成一笔评价", reward: "+10 积分", actionText: "去评价", route: "/pages/order/list/index" },
      { id: "invite", icon: "邀", title: "邀请好友", subtitle: "好友下单后可得奖励", reward: "+100 积分", actionText: "去邀请", route: "/pages/invite-friends/index" },
      { id: "checkin", icon: "签", title: "每日签到", subtitle: "连续签到积分更多", reward: "+5 积分", actionText: "签到", route: "" }
    ],
    rewards: [
      { id: "coupon_5", amount: "¥5", title: "500 积分", subtitle: "兑 ¥5 优惠券", accentClass: "blue" },
      { id: "delivery_12", amount: "配", title: "1200 积分", subtitle: "兑换配送券", accentClass: "orange" }
    ],
    records: [
      { id: "pt_order", icon: "+", title: "订单完成奖励", time: "今天 12:52", pointsText: "+30", typeClass: "income" },
      { id: "pt_review", icon: "+", title: "评价订单奖励", time: "昨天 14:40", pointsText: "+10", typeClass: "income" },
      { id: "pt_redeem", icon: "-", title: "兑换优惠券", time: "周一", pointsText: "-500", typeClass: "expense" }
    ],
    filters: [
      { id: "all", title: "全部", active: "active" },
      { id: "earn", title: "收入", active: "" },
      { id: "redeem", title: "支出", active: "" }
    ]
  },
  onShow() {
    this.loadPoints();
  },
  async loadPoints() {
    ensurePreviewAuth();
    try {
      const summary = await getUserPointsSummary() as Record<string, unknown>;
      const benefits = (summary.benefits as Array<Record<string, unknown>> || []).map((item) => ({
        icon: String(item.icon || "分"),
        title: String(item.title || "会员权益"),
        status: String(item.status || ""),
        lockedClass: item.unlocked ? "" : "locked"
      }));
      const tasks = (summary.tasks as Array<Record<string, unknown>> || []).map((item) => ({
        id: String(item.id || ""),
        icon: taskIcon(String(item.id || "")),
        title: String(item.title || ""),
        subtitle: String(item.subtitle || ""),
        reward: `+${item.reward || 0} 积分`,
        actionText: String(item.action_text || "去完成"),
        route: String(item.route || "")
      }));
      const rewards = (summary.rewards as Array<Record<string, unknown>> || []).map((item) => ({
        id: String(item.id || ""),
        amount: item.amount_fen ? `¥${(Number(item.amount_fen) / 100).toFixed(0)}` : "券",
        title: `${item.points || 0} 积分`,
        subtitle: String(item.title || ""),
        accentClass: String(item.accent_color || "").includes("ff") ? "orange" : "blue"
      }));
      const records = (summary.transactions as Array<Record<string, unknown>> || []).map(pointsRecordFromApi);
      this.setData({
        points: Number(summary.points || 0),
        nickname: String(summary.nickname || "张三"),
        levelName: String(summary.level_name || "美食达人"),
        growthText: `距 V3 还需 ${summary.next_level_growth || 1320} 成长值`,
        growthPercent: 67,
        benefits: benefits.length ? benefits : this.data.benefits,
        tasks: tasks.length ? tasks : this.data.tasks,
        rewards: rewards.length ? rewards : this.data.rewards,
        records: records.length ? records : this.data.records
      });
    } catch (_error) {
      wx.showToast({ title: "积分缓存已加载", icon: "none" });
    }
  },
  async handleTaskTap(event) {
    const id = String(event.currentTarget.dataset.id || "");
    const route = String(event.currentTarget.dataset.route || "");
    if (id === "checkin") {
      try {
        const summary = await checkInPoints() as Record<string, unknown>;
        this.setData({ points: Number(summary.points || this.data.points) });
        await this.loadPoints();
        wx.showToast({ title: "签到成功", icon: "success" });
      } catch (_error) {
        wx.showToast({ title: "今日已签到", icon: "none" });
      }
      return;
    }
    wx.navigateTo({ url: route || "/pages/index/index" });
  },
  handleRewardTap() {
    wx.navigateTo({ url: "/pages/coupons/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});

function pointsRecordFromApi(item: Record<string, unknown>) {
  const points = Number(item.points || 0);
  return {
    id: String(item.id || Date.now()),
    icon: points >= 0 ? "+" : "-",
    title: String(item.title || "积分流水"),
    time: String(item.created_at || "").slice(0, 10) || "刚刚",
    pointsText: `${points >= 0 ? "+" : ""}${points}`,
    typeClass: points >= 0 ? "income" : "expense"
  };
}

function taskIcon(id: string) {
  switch (id) {
    case "order":
      return "单";
    case "review":
      return "评";
    case "invite":
      return "邀";
    case "checkin":
      return "签";
    default:
      return "分";
  }
}
