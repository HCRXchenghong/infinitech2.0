import { getOrders, setDevAuthToken } from "../../../utils/api";

Page({
  data: {
    orders: [
      {
        id: "ord_preview",
        shopName: "蓝海餐厅",
        status: "待骑手接单",
        amount: "55.98",
        subtitle: "招牌牛肉饭 x 2",
        time: "今天 12:30"
      }
    ]
  },
  onLoad() {
    this.loadOrders();
  },
  async loadOrders() {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      const orders = await getOrders();
      if (Array.isArray(orders)) {
        this.setData({
          orders: orders.map((order) => ({
            id: order.id,
            shopName: order.shop_id || "蓝海餐厅",
            status: statusText(order.status),
            amount: (Number(order.amount_fen || 0) / 100).toFixed(2),
            subtitle: Array.isArray(order.items) && order.items.length > 0 ? `${order.items[0].product_name} x ${order.items[0].quantity}` : "订单详情",
            time: "刚刚"
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleOrderTap(event) {
    wx.navigateTo({ url: `/pages/order/detail/index?id=${event.currentTarget.dataset.id}` });
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
