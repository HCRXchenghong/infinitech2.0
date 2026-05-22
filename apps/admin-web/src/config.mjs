export const ADMIN_WEB_SECTIONS = Object.freeze([
  {
    key: "operate",
    title: "运营中枢",
    modules: ["dashboard", "orders", "after-sales", "merchants", "riders", "rider-performance", "dispatch"]
  },
  {
    key: "finance",
    title: "资金与风控",
    modules: ["refund-settings", "payment", "wallet", "pricing", "risk-control", "data-management"]
  },
  {
    key: "growth",
    title: "营销内容",
    modules: ["home-cards", "featured-products", "home-campaigns", "coupons", "circle", "groups", "reviews"]
  },
  {
    key: "service",
    title: "服务与集成",
    modules: ["support", "notifications", "rtc", "contact-audits", "integrations", "content-settings", "settings"]
  }
]);

export const ADMIN_WEB_MODULES = Object.freeze([
  { key: "dashboard", title: "运营首页", status: "ready", priority: "P0", owner: "运营" },
  { key: "orders", title: "订单监控", status: "wired", priority: "P0", owner: "交易" },
  { key: "after-sales", title: "售后审核", status: "wired", priority: "P0", owner: "客服" },
  { key: "merchants", title: "商户资质", status: "wired", priority: "P0", owner: "招商" },
  { key: "riders", title: "骑手/站长", status: "wired", priority: "P0", owner: "配送" },
  { key: "rider-performance", title: "骑手绩效", status: "wired", priority: "P1", owner: "配送" },
  { key: "dispatch", title: "派单审计", status: "wired", priority: "P0", owner: "配送" },
  { key: "refund-settings", title: "退款策略", status: "wired", priority: "P0", owner: "财务" },
  { key: "payment", title: "支付中心", status: "planned", priority: "P0", owner: "财务" },
  { key: "wallet", title: "钱包财务", status: "planned", priority: "P0", owner: "财务" },
  { key: "pricing", title: "骑手计价", status: "planned", priority: "P1", owner: "财务" },
  { key: "risk-control", title: "风控中心", status: "planned", priority: "P0", owner: "安全" },
  { key: "data-management", title: "数据备份", status: "planned", priority: "P0", owner: "运维" },
  { key: "home-cards", title: "首页卡片", status: "planned", priority: "P1", owner: "运营" },
  { key: "featured-products", title: "精选商品", status: "planned", priority: "P1", owner: "运营" },
  { key: "home-campaigns", title: "首页活动", status: "planned", priority: "P1", owner: "运营" },
  { key: "coupons", title: "优惠券", status: "planned", priority: "P0", owner: "营销" },
  { key: "circle", title: "圈子饭搭", status: "planned", priority: "P1", owner: "内容" },
  { key: "groups", title: "群聊红包", status: "planned", priority: "P1", owner: "社群" },
  { key: "reviews", title: "评价管理", status: "planned", priority: "P1", owner: "内容" },
  { key: "support", title: "客服工作台", status: "planned", priority: "P0", owner: "客服" },
  { key: "notifications", title: "通知推送", status: "planned", priority: "P1", owner: "运营" },
  { key: "rtc", title: "RTC 审计", status: "planned", priority: "P1", owner: "质检" },
  { key: "contact-audits", title: "电话审计", status: "planned", priority: "P1", owner: "质检" },
  { key: "integrations", title: "OAuth/API", status: "planned", priority: "P0", owner: "开放平台" },
  { key: "content-settings", title: "内容设置", status: "planned", priority: "P1", owner: "内容" },
  { key: "settings", title: "系统设置", status: "planned", priority: "P0", owner: "运维" }
]);

export const ADMIN_WEB_KPIS = Object.freeze([
  { key: "paidOrders", title: "今日已支付", value: "1,284", trend: "+8.6%", tone: "blue" },
  { key: "afterSales", title: "待售后审核", value: "37", trend: "P0", tone: "red" },
  { key: "riderOnline", title: "在线骑手", value: "426", trend: "87% 活跃", tone: "green" },
  { key: "merchantRisk", title: "资质临期/过期", value: "19", trend: "需关店复核", tone: "amber" },
  { key: "outbox", title: "Outbox 待处理", value: "0", trend: "以接口为准", tone: "slate" },
  { key: "objectCleanup", title: "对象清理失败", value: "0", trend: "以接口为准", tone: "slate" }
]);

export const ADMIN_WEB_QUEUES = Object.freeze([
  { key: "after-sales-list", title: "售后审核", level: "P0", target: "30 分钟内首响", operationKey: "after-sales-list" },
  { key: "merchant-invite", title: "商户邀约", level: "P0", target: "管理员创建链接", operationKey: "merchant-invite" },
  { key: "station-manager-invite", title: "站长邀约", level: "P0", target: "站点实名准入", operationKey: "station-manager-invite" },
  { key: "refund-settings-read", title: "退款策略", level: "P0", target: "余额/原路策略", operationKey: "refund-settings-read" },
  { key: "outbox-stats", title: "事件队列健康", level: "P0", target: "阻塞和租约预警", operationKey: "outbox-stats" },
  { key: "object-cleanup-stats", title: "对象清理", level: "P1", target: "过期证据和失败账本", operationKey: "object-cleanup-stats" }
]);

export const ADMIN_WEB_RBAC = Object.freeze([
  { role: "super_admin", name: "超级管理员", scopes: ["*"] },
  { role: "ops_admin", name: "运营管理员", scopes: ["orders:read", "after_sales:review", "merchant:review", "home:write"] },
  { role: "finance_admin", name: "财务管理员", scopes: ["refund:write", "wallet:read", "settlement:read"] },
  { role: "dispatch_admin", name: "调度管理员", scopes: ["dispatch:read", "dispatch:manual_assign", "rider:read"] },
  { role: "support_admin", name: "客服管理员", scopes: ["support:read", "after_sales:event", "rtc:audit"] }
]);

export function getAdminWebModule(key) {
  return ADMIN_WEB_MODULES.find((module) => module.key === key) || null;
}
