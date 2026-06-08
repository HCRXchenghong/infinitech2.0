import { getOrderDetail, setDevAuthToken } from "../../../utils/api";
import { generatedImages, mediaUrl, productFallbackImage } from "../../../utils/media";

Page({
  data: {
    orderId: "ord_preview",
    status: "配送中",
    reviewed: false,
    amount: "55.98",
    eta: "预计 12:52 送达",
    mapText: "骑手正在前往商家",
    shopName: "蓝海餐厅",
    address: {
      detail: "配送地址待同步",
      contact: "联系人待同步"
    },
    events: [
      { type: "order.checkout_created", message: "订单已创建", createdText: "刚刚" },
      { type: "order.payment.success", message: "余额支付成功，订单进入待调度", createdText: "刚刚" },
      { type: "merchant.accepted", message: "商家已接单", createdText: "刚刚" },
      { type: "delivery.picked_up", message: "骑手取餐中", createdText: "刚刚" },
      { type: "delivery.delivering", message: "配送中", createdText: "刚刚" }
    ],
    items: [
      { product_id: "prod_beef_rice", product_name: "招牌牛肉饭", imageUrl: generatedImages.productBeefRice, quantity: 2, price: "51.98" },
      { product_id: "prod_tea", product_name: "柠檬茶", imageUrl: generatedImages.productLemonTea, quantity: 1, price: "9.00" }
    ]
  },
  onLoad(query) {
    const orderId = String(query?.id || "ord_preview");
    this.setData({ orderId });
    this.loadOrder(orderId);
  },
  async loadOrder(orderId: string) {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      const order = await getOrderDetail(orderId);
      const addressSnapshot = order.address_snapshot || {};
      this.setData({
        status: statusText(order.status, Boolean(order.reviewed)),
        reviewed: Boolean(order.reviewed),
        amount: (Number(order.amount_fen || 0) / 100).toFixed(2),
        eta: etaText(order.status),
        mapText: mapText(order.status),
        shopName: String(order.shop_name || this.data.shopName),
        address: {
          detail: `${addressSnapshot.city || ""} ${addressSnapshot.detail || ""}`.trim() || "配送地址待同步",
          contact: `${addressSnapshot.contact_name || ""} ${addressSnapshot.contact_phone || ""}`.trim() || this.data.address.contact
        },
        events: Array.isArray(order.events) ? order.events.map((item) => ({
          ...item,
          createdText: formatEventTime(item.created_at)
        })) : [],
        items: Array.isArray(order.items) ? order.items.map((item) => ({
          ...item,
          imageUrl: mediaUrl(item.image_url, productFallbackImage(String(item.product_id || ""))),
          price: (Number(item.unit_price_fen || 0) / 100).toFixed(2)
        })) : []
      });
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleAfterSalesTap() {
    wx.navigateTo({ url: `/pages/after-sales/index?order_id=${this.data.orderId}` });
  },
  handleReviewTap() {
    wx.navigateTo({ url: `/pages/order/review/index?order_id=${this.data.orderId}` });
  },
  handleServiceTap() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/order/list/index" }) });
  },
  handleMerchantTap() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index" });
  },
  handleRiderTap() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index" });
  }
});

function statusText(status: string, reviewed = false) {
  const map: Record<string, string> = {
    pending_payment: "待支付",
    merchant_pending: "等待商家接单",
    preparing: "商家备餐中",
    dispatching: "待骑手接单",
    rider_assigned: "骑手已接单",
    completed: "已完成",
    cancelled: "已取消",
    refunded: "已退款"
  };
  if (status === "completed" && !reviewed) {
    return "待评价";
  }
  return map[status] || status || "订单处理中";
}

function etaText(status: string) {
  if (status === "completed") return "已送达";
  if (status === "pending_payment") return "等待支付";
  if (status === "merchant_pending") return "商家即将确认订单";
  if (status === "preparing") return "商家正在备餐";
  return "预计 12:52 送达";
}

function mapText(status: string) {
  if (status === "completed") return "订单已完成，欢迎给商家一个评价。";
  if (status === "pending_payment") return "支付完成后会立即进入商家接单流程。";
  if (status === "merchant_pending") return "订单已提交，正在等待商家确认。";
  if (status === "preparing") return "商家正在备餐，完成后会推送骑手取餐。";
  if (status === "dispatching") return "商家已出餐，系统正在为你匹配骑手。";
  if (status === "rider_assigned") return "骑手已接单，正在赶往取餐点。";
  return "骑手正在前往商家";
}

function formatEventTime(value: string) {
  const text = String(value || "").trim();
  if (!text) return "刚刚";
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) return "刚刚";
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${month}-${day} ${hour}:${minute}`;
}
