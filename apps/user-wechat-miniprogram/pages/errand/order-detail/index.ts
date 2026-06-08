import { ensurePreviewAuth, getErrandOrderDetail } from "../../../utils/api";
import { errandFallbackImage, generatedImages, mediaUrl } from "../../../utils/media";

Page({
  data: {
    orderId: "errand_preview",
    status: "骑手已接单",
    orderNo: "PT10086",
    estimateText: "预计 14:25 送达",
    imageUrl: generatedImages.errandParcel,
    serviceTitle: "帮取 · 小件包裹",
    mapStatus: "骑手正在前往取件地址",
    pickupAddress: "取件地址待同步",
    deliveryAddress: "送达地址待同步",
    rider: {
      name: "张师傅",
      ratingText: "4.9",
      vehicle: "电动车",
      distanceText: "距取件地 600m"
    },
    infoRows: [
      { icon: "取", title: "取件地址", value: "取件地址待同步" },
      { icon: "送", title: "送达地址", value: "送达地址待同步" },
      { icon: "人", title: "联系人", value: "联系人待同步" },
      { icon: "注", title: "任务备注", value: "任务备注待同步" }
    ],
    feeRows: [
      { title: "起步价", value: "¥10" },
      { title: "距离费", value: "¥4" },
      { title: "服务费", value: "¥2" },
      { title: "优惠券", value: "-¥3" },
      { title: "实付", value: "¥13.00", strong: "strong" }
    ],
    timeline: [
      { title: "订单已创建", subtitle: "", time: "14:02", statusClass: "done" },
      { title: "骑手已接单", subtitle: "", time: "14:04", statusClass: "done" },
      { title: "前往取件", subtitle: "骑手正在前往取件地址", time: "进行中", statusClass: "active" },
      { title: "已取件", subtitle: "待骑手取件", time: "", statusClass: "pending" },
      { title: "送达完成", subtitle: "请确认收货并评价", time: "", statusClass: "pending" }
    ],
    tips: ["如物品超重或超距，骑手会发起补差价", "取消订单可能产生跑腿空驶费", "可在订单完成后评价骑手服务"]
  },
  onLoad(query) {
    const orderId = String(query?.id || "errand_preview");
    this.setData({ orderId });
    this.loadOrder(orderId);
  },
  async loadOrder(orderId) {
    if (orderId === "errand_preview") return;
    ensurePreviewAuth();
    try {
      const detail = await getErrandOrderDetail(orderId) as Record<string, unknown>;
      const order = detail.order as Record<string, unknown> || {};
      this.setData({
        status: statusText(String(order.status || "")),
        orderNo: String(order.id || orderId).replace("ord_", "PT1008"),
        estimateText: String(detail.estimate_text || this.data.estimateText),
        imageUrl: mediaUrl(detail.image_url, errandFallbackImage()),
        serviceTitle: `${detail.service_title || "帮取"} · ${detail.item_type || "小件包裹"}`,
        mapStatus: String(detail.map_status || this.data.mapStatus),
        pickupAddress: String(detail.pickup_address || this.data.pickupAddress),
        deliveryAddress: String(detail.delivery_address || this.data.deliveryAddress),
        rider: detail.rider || this.data.rider,
        infoRows: [
          { icon: "取", title: "取件地址", value: String(detail.pickup_address || "") },
          { icon: "送", title: "送达地址", value: String(detail.delivery_address || "") },
          { icon: "人", title: "联系人", value: `${detail.contact_name || ""} ${detail.contact_phone || ""}`.trim() || "联系人待同步" },
          { icon: "注", title: "任务备注", value: String(detail.description || "任务备注待同步") }
        ],
        feeRows: (detail.fee_rows as Array<Record<string, unknown>> || []).map(feeRowFromApi).concat([{ title: "实付", value: `¥${formatFen(Number(order.amount_fen || 0))}`, strong: "strong" }]),
        timeline: (detail.timeline as Array<Record<string, unknown>> || []).map((item) => ({
          title: String(item.title || ""),
          subtitle: String(item.subtitle || ""),
          time: String(item.time || ""),
          statusClass: String(item.status || "pending")
        }))
      });
    } catch (_error) {
      this.setData({ status: "离线缓存" });
    }
  },
  handleCancel() {
    wx.showToast({ title: "取消申请已记录", icon: "none" });
  },
  handleReminder() {
    wx.showToast({ title: "已提醒骑手", icon: "none" });
  },
  handleContactRider() {
    wx.navigateTo({ url: "/pages/messages/merchant-group/index?thread_id=rider_zhang" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/errand/home/index" }) });
  }
});

function statusText(status: string) {
  switch (status) {
    case "rider_assigned":
      return "骑手已接单";
    case "delivering":
      return "配送中";
    case "completed":
      return "已送达";
    default:
      return "订单处理中";
  }
}

function feeRowFromApi(item: Record<string, unknown>) {
  const amountFen = Number(item.amount_fen || 0);
  return {
    title: String(item.title || ""),
    value: `${amountFen < 0 ? "-" : ""}¥${formatFen(Math.abs(amountFen))}`,
    strong: ""
  };
}

function formatFen(amountFen: number) {
  return (amountFen / 100).toFixed(2);
}
