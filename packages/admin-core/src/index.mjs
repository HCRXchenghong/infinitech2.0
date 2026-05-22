export const ADMIN_MODULES = Object.freeze([
  { key: "dashboard", title: "仪表盘", route: "/dashboard" },
  { key: "orders", title: "订单中心", route: "/orders" },
  { key: "after-sales", title: "售后与投诉", route: "/after-sales" },
  { key: "merchants", title: "商户管理", route: "/merchants" },
  { key: "riders", title: "骑手管理", route: "/riders" },
  { key: "rider-performance", title: "骑手绩效", route: "/rider-performance" },
  { key: "dispatch", title: "调度配置", route: "/dispatch" },
  { key: "pricing", title: "骑手计价", route: "/rider-pricing" },
  { key: "refund-settings", title: "退款策略", route: "/refund-settings" },
  { key: "home-cards", title: "首页卡片", route: "/home-cards" },
  { key: "featured-products", title: "精选商品", route: "/featured-products" },
  { key: "home-campaigns", title: "首页活动", route: "/home-campaigns" },
  { key: "coupons", title: "优惠券", route: "/coupons" },
  { key: "circle", title: "圈子与饭搭", route: "/circle" },
  { key: "reviews", title: "评价管理", route: "/reviews" },
  { key: "groups", title: "群聊与红包", route: "/groups" },
  { key: "payment", title: "支付中心", route: "/payment" },
  { key: "wallet", title: "钱包财务", route: "/wallet" },
  { key: "points-membership", title: "积分会员", route: "/points-membership" },
  { key: "notifications", title: "通知推送", route: "/notifications" },
  { key: "support", title: "客服工作台", route: "/support" },
  { key: "rtc", title: "RTC 审计", route: "/rtc" },
  { key: "contact-audits", title: "电话联系审计", route: "/contact-audits" },
  { key: "integrations", title: "OAuth/API", route: "/integrations" },
  { key: "risk-control", title: "风控中心", route: "/risk-control" },
  { key: "data-management", title: "数据备份恢复", route: "/data-management" },
  { key: "content-settings", title: "内容设置", route: "/content-settings" },
  { key: "settings", title: "系统设置", route: "/settings" }
]);

export function getAdminModule(key) {
  return ADMIN_MODULES.find((item) => item.key === key) || null;
}
