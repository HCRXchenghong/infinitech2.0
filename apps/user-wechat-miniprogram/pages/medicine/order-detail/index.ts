import { ensurePreviewAuth, getMedicineOrderDetail } from "../../../utils/api";
import { generatedImages, mediaUrl, medicineFallbackImage } from "../../../utils/media";

Page({
  data: {
    orderId: "medicine_preview",
    status: "骑手已取药",
    eta: "预计 15:20 送达",
    address: "配送地址待同步",
    contact: "联系人待同步",
    deliveryText: "校内骑手配送",
    payable: "40.60",
    advice: "请按校医指导用药；如症状加重请及时就医。",
    items: [
      { product_id: "med_cooling_patch", name: "退热贴", imageUrl: generatedImages.medicineCoolingPatch, category: "校医务室", price: "12.90", quantity: 1, approved: false },
      { product_id: "med_amoxicillin", name: "阿莫西林胶囊", imageUrl: generatedImages.medicineCapsules, category: "处方药", price: "18.80", quantity: 1, approved: true },
      { product_id: "med_swab", name: "碘伏棉签", imageUrl: generatedImages.medicineFirstAid, category: "外伤消毒", price: "6.90", quantity: 1, approved: false }
    ],
    timeline: [
      { title: "订单已提交", time: "14:36", status: "done" },
      { title: "校医出药", time: "14:40", status: "done" },
      { title: "骑手已取药", time: "14:48", status: "active" },
      { title: "送达完成", time: "待完成", status: "pending" }
    ],
    fees: [
      { title: "商品金额", value: "¥38.60" },
      { title: "配送费", value: "¥2.00" },
      { title: "实付", value: "¥40.60" }
    ]
  },
  onLoad(query) {
    const orderId = String(query?.id || "medicine_preview");
    this.setData({ orderId });
    this.loadOrder(orderId);
  },
  async loadOrder(orderId) {
    if (orderId === "medicine_preview") return;
    ensurePreviewAuth();
    try {
      const detail = await getMedicineOrderDetail(orderId) as Record<string, unknown>;
      const order = detail.order as Record<string, unknown> || {};
      const items = Array.isArray(detail.items) ? detail.items : [];
      const timeline = Array.isArray(detail.timeline) ? detail.timeline : [];
      const fees = Array.isArray(detail.fee_rows) ? detail.fee_rows : [];
      this.setData({
        status: statusText(String(order.status || "")),
        eta: "预计 15:20 送达",
        address: String(detail.address || "配送地址待同步"),
        contact: `${String(detail.contact_name || "")}  ${String(detail.contact_phone || "")}`.trim() || "联系人待同步",
        deliveryText: String(detail.delivery_text || this.data.deliveryText),
        advice: String(detail.advice || this.data.advice),
        payable: (Number(order.amount_fen || 4060) / 100).toFixed(2),
        items: items.length ? items.map((item) => ({
          product_id: String(item.product_id || item.name || ""),
          name: String(item.name || ""),
          imageUrl: mediaUrl(item.image_url, medicineFallbackImage(String(item.product_id || ""))),
          category: String(item.category || ""),
          price: (Number(item.price_fen || 0) / 100).toFixed(2),
          quantity: Number(item.quantity || 1),
          approved: Boolean(item.prescription_approved)
        })) : this.data.items,
        timeline: timeline.length ? timeline.map((item) => ({
          title: String(item.title || ""),
          time: String(item.time || ""),
          status: String(item.status || "pending")
        })) : this.data.timeline,
        fees: fees.length ? fees.map((item) => ({
          title: String(item.title || ""),
          value: `¥${(Number(item.amount_fen || 0) / 100).toFixed(2)}`
        })) : this.data.fees
      });
    } catch (_error) {
      this.setData({ status: "本地预览" });
    }
  },
  handleService() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index" });
  },
  handleConfirm() {
    wx.showToast({ title: "已确认收药", icon: "success" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/medicine/home/index" }) });
  }
});

function statusText(status: string) {
  switch (status) {
    case "picked_up":
      return "骑手已取药";
    case "delivering":
      return "配送中";
    case "completed":
      return "已完成";
    case "pending_payment":
      return "待支付";
    default:
      return "骑手已取药";
  }
}
