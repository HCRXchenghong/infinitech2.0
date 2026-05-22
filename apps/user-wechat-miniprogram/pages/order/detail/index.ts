import { getOrderDetail, setDevAuthToken } from "../../../utils/api";

Page({
  data: {
    orderId: "ord_preview",
    status: "待骑手接单",
    amount: "55.98",
    events: [
      { type: "order.checkout_created", message: "订单已创建" },
      { type: "order.payment.success", message: "余额支付成功，订单进入待调度" }
    ],
    items: [
      { product_name: "招牌牛肉饭", quantity: 2, price: "25.99" }
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
      this.setData({
        status: statusText(order.status),
        amount: (Number(order.amount_fen || 0) / 100).toFixed(2),
        events: order.events || [],
        items: Array.isArray(order.items) ? order.items.map((item) => ({
          ...item,
          price: (Number(item.unit_price_fen || 0) / 100).toFixed(2)
        })) : []
      });
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  }
});

function statusText(status: string) {
  const map: Record<string, string> = {
    pending_payment: "待支付",
    dispatching: "待骑手接单",
    rider_assigned: "骑手已接单",
    completed: "已完成",
    cancelled: "已取消",
    refunded: "已退款"
  };
  return map[status] || status || "订单处理中";
}
