import { ensurePreviewAuth, getFeedbackTickets, getServiceTickets } from "../../../utils/api";

Page({
  data: {
    activeTab: "全部",
    tabs: ["全部", "处理中", "待确认", "已完成", "已关闭"],
    stats: [
      { title: "全部", value: 5 },
      { title: "处理中", value: 2 },
      { title: "已完成", value: 3 }
    ],
    keyword: "",
    records: [
      { id: "st_preview_delivery", icon: "送", title: "配送问题 · 预计送达未更新", status: "处理中", order: "蓝海餐厅 #DD240518001", progress: "商家反馈：补做菜品，预计 8 分钟后出餐", time: "今天 14:18", primary: "查看详情", secondary: "补充说明" },
      { id: "st_preview_quality", icon: "质", title: "商品质量 · 少送小菜", status: "待确认", order: "川味小馆 #DD240516006", progress: "客服已提出补偿方案，待你确认", time: "05-16 12:42", primary: "确认结果", secondary: "继续沟通" },
      { id: "st_preview_red_packet", icon: "包", title: "红包钱包 · 红包未到账", status: "已完成", order: "GD2405130036", progress: "已退回余额 ¥9.14", time: "05-13 20:08", primary: "查看详情", secondary: "" }
    ]
  },
  onShow() {
    this.loadRecords();
  },
  async loadRecords() {
    ensurePreviewAuth();
    try {
      const [serviceTickets, feedbackTickets] = await Promise.all([
        getServiceTickets(),
        getFeedbackTickets()
      ]) as [Array<Record<string, unknown>>, Array<Record<string, unknown>>];
      const mapped = serviceTickets.map((item) => ({
        id: String(item.id || ""),
        icon: iconForType(String(item.type || item.category || "")),
        title: String(item.title || item.category || "用户反馈"),
        status: statusText(String(item.status || "processing")),
        order: String(item.related_order_title || item.related_order_id || "客服工单"),
        progress: String(item.solution || item.content || "客服正在处理"),
        time: formatTime(String(item.updated_at || item.created_at || "")),
        primary: "查看详情",
        secondary: String(item.status || "") === "waiting_confirm" ? "确认结果" : "补充说明"
      }));
      const feedbackMapped = feedbackTickets.map((item) => ({
        id: String(item.id || ""),
        icon: "议",
        title: `${String(item.type || "功能建议")} · ${String(item.content || "用户反馈").slice(0, 10)}`,
        status: statusText(String(item.status || "processing")),
        order: "平台反馈",
        progress: String(item.content || "建议已收录，感谢反馈"),
        time: formatTime(String(item.updated_at || item.created_at || "")),
        primary: "再次反馈",
        secondary: ""
      }));
      const records = mapped.concat(feedbackMapped);
      if (records.length) {
        this.setData({ records, stats: buildStats(records) });
      }
    } catch (_error) {
      // Keep seeded preview records when the API service is not running.
    }
  },
  handleTab(event) {
    this.setData({ activeTab: String(event.currentTarget.dataset.tab || "全部") });
  },
  handleDetail(event) {
    const id = String(event.currentTarget.dataset.id || "");
    if (id.startsWith("st_")) {
      wx.navigateTo({ url: `/pages/service-ticket/detail/index?id=${id}` });
      return;
    }
    wx.navigateTo({ url: "/pages/feedback/complaint/index" });
  },
  handleNew() {
    wx.navigateTo({ url: "/pages/feedback/complaint/index" });
  },
  handleService() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});

function statusText(status: string) {
  switch (status) {
    case "resolved":
      return "已完成";
    case "waiting_confirm":
      return "待确认";
    case "closed":
      return "已关闭";
    default:
      return "处理中";
  }
}

function iconForType(type: string) {
  if (type.includes("red")) return "包";
  if (type.includes("quality")) return "质";
  if (type.includes("delivery")) return "送";
  return "议";
}

function formatTime(value: string) {
  if (!value) return "今天 14:18";
  return value.slice(5, 16).replace("T", " ");
}

function buildStats(records: Array<{ status: string }>) {
  return [
    { title: "全部", value: records.length },
    { title: "处理中", value: records.filter((item) => item.status === "处理中").length },
    { title: "已完成", value: records.filter((item) => item.status === "已完成").length }
  ];
}
