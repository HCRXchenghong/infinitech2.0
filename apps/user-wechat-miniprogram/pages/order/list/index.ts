import { getAuthToken, getOrders } from "../../../utils/api";
import { mediaUrl, productFallbackImage } from "../../../utils/media";

type OrderCard = {
  id: string;
  shopName: string;
  imageUrl: string;
  status: string;
  statusKey: string;
  statusTone: string;
  amount: string;
  subtitle: string;
  time: string;
  eta: string;
  primaryAction: string;
  primaryActionKey: string;
  secondaryAction: string;
  secondaryActionKey: string;
};

Page({
  data: {
    keyword: "",
    activeStatus: "all",
    tabs: [
      { key: "all", title: "全部" },
      { key: "pending", title: "待支付" },
      { key: "running", title: "进行中" },
      { key: "review", title: "待评价" },
      { key: "after_sales", title: "退款/售后" }
    ],
    orders: [] as OrderCard[],
    visibleOrders: [] as OrderCard[],
    emptyText: ""
  },
  onLoad(options) {
    this.applyStatusIntent(options?.status);
    this.applyOrderFilter(this.data.orders);
    this.loadOrders();
  },
  onShow() {
    const status = String(wx.getStorageSync("orderListStatus") || "");
    if (status) {
      wx.removeStorageSync("orderListStatus");
      this.applyStatusIntent(status);
    }
  },
  async loadOrders() {
    if (!getAuthToken()) {
      this.setData({
        orders: [],
        visibleOrders: [],
        emptyText: "请先登录后查看真实订单。"
      });
      return;
    }
    try {
      const orders = await getOrders();
      if (Array.isArray(orders)) {
        const mappedOrders = orders.map(orderFromApi);
        this.setData({ orders: mappedOrders });
        this.applyOrderFilter(mappedOrders);
      }
    } catch (error) {
      this.setData({
        orders: [],
        visibleOrders: [],
        emptyText: String((error as any)?.message || "订单接口连接失败，请稍后重试。")
      });
    }
  },
  handleSearchInput(event) {
    const keyword = String(event.detail.value || "").trim();
    this.setData({ keyword });
    this.applyOrderFilter(this.data.orders, this.data.activeStatus, keyword);
  },
  handleClearSearch() {
    this.setData({ keyword: "" });
    this.applyOrderFilter(this.data.orders, this.data.activeStatus, "");
  },
  handleTabTap(event) {
    const activeStatus = String(event.currentTarget.dataset.key || "all");
    this.setData({ activeStatus });
    this.applyOrderFilter(this.data.orders, activeStatus, this.data.keyword);
  },
  handleOrderTap(event) {
    wx.navigateTo({ url: `/pages/order/detail/index?id=${event.currentTarget.dataset.id}` });
  },
  handleActionTap(event) {
    const orderId = String(event.currentTarget.dataset.id || "");
    const key = String(event.currentTarget.dataset.key || "detail");
    if (!orderId) return;
    if (key === "review") {
      wx.navigateTo({ url: `/pages/order/review/index?order_id=${orderId}` });
      return;
    }
    if (key === "after_sales") {
      wx.navigateTo({ url: `/pages/after-sales/index?order_id=${orderId}` });
      return;
    }
    if (key === "rider") {
      wx.navigateTo({ url: "/pages/customer-service/chat/index" });
      return;
    }
    if (key === "reorder") {
      wx.switchTab({ url: "/pages/index/index" });
      return;
    }
    wx.navigateTo({ url: `/pages/order/detail/index?id=${orderId}` });
  },
  applyStatusIntent(status) {
    const key = normalizeOrderStatus(String(status || ""));
    if (key) {
      this.setData({ activeStatus: key });
      this.applyOrderFilter(this.data.orders, key);
    }
  },
  applyOrderFilter(
    orders = this.data.orders,
    activeStatus = String(this.data.activeStatus || "all"),
    keyword = String(this.data.keyword || "")
  ) {
    const list = Array.isArray(orders) ? orders : [];
    const searchText = String(keyword || "").trim();
    const visibleOrders = list.filter((order) => {
      const normalized = normalizeOrderStatus(order.statusKey || order.status || order.primaryActionKey);
      const statusMatched = activeStatus === "all" || normalized === activeStatus;
      const searchMatched = !searchText || `${order.shopName}${order.subtitle}${order.status}${order.time}`.includes(searchText);
      return statusMatched && searchMatched;
    });
    this.setData({
      visibleOrders,
      emptyText: visibleOrders.length ? "" : emptyTextForStatus(activeStatus, Boolean(searchText))
    });
  }
});

function orderFromApi(order: Record<string, any>): OrderCard {
  const statusKey = normalizeOrderStatus(String(order.status || ""), Boolean(order.reviewed));
  const status = statusText(String(order.status || ""), Boolean(order.reviewed));
  return {
    id: String(order.id || ""),
    shopName: String(order.shop_name || order.shop_id || "订单商家"),
    imageUrl: firstOrderImage(order.items),
    status,
    statusKey,
    statusTone: statusTone(statusKey),
    amount: (Number(order.amount_fen || 0) / 100).toFixed(2),
    subtitle: summarizeOrderItems(order.items),
    time: formatOrderTime(order.updated_at || order.created_at),
    eta: etaText(String(order.status || "")),
    primaryAction: primaryActionText(String(order.status || ""), Boolean(order.reviewed)),
    primaryActionKey: primaryActionKey(String(order.status || ""), Boolean(order.reviewed)),
    secondaryAction: secondaryActionText(String(order.status || ""), Boolean(order.reviewed)),
    secondaryActionKey: secondaryActionKey(String(order.status || ""), Boolean(order.reviewed))
  };
}

function normalizeOrderStatus(status: string, reviewed = false) {
  const map: Record<string, string> = {
    all: "all",
    pending: "pending",
    pending_payment: "pending",
    merchant_pending: "running",
    preparing: "running",
    dispatching: "running",
    rider_assigned: "running",
    in_progress: "running",
    running: "running",
    review: "review",
    pending_review: "review",
    after_sales: "after_sales",
    refunded: "after_sales"
  };
  if (status === "completed" && !reviewed) return "review";
  if (status === "completed" && reviewed) return "completed";
  return map[status] || "";
}

function emptyTextForStatus(status: string, searching = false) {
  if (searching) return "没有搜到相关订单。";
  const map: Record<string, string> = {
    all: "还没有订单，去首页看看今天想吃什么。",
    pending: "没有待支付订单。",
    running: "没有进行中的订单。",
    review: "没有待评价订单。",
    after_sales: "没有退款或售后订单。"
  };
  return map[status] || "没有符合条件的订单。";
}

function statusText(status: string, reviewed = false) {
  const map: Record<string, string> = {
    pending_payment: "待支付",
    merchant_pending: "待商家接单",
    preparing: "商家备餐中",
    dispatching: "待骑手接单",
    rider_assigned: "配送中",
    completed: "已完成",
    cancelled: "已取消",
    refunded: "已退款"
  };
  if (status === "completed" && !reviewed) {
    return "待评价";
  }
  return map[status] || status || "订单处理中";
}

function statusTone(statusKey: string) {
  if (statusKey === "running") return "blue";
  if (statusKey === "review") return "orange";
  if (statusKey === "pending") return "red";
  if (statusKey === "after_sales") return "gray";
  return "gray";
}

function primaryActionText(status: string, reviewed = false) {
  if (status === "pending_payment") return "去支付";
  if (status === "completed" && !reviewed) return "去评价";
  if (status === "refunded") return "售后详情";
  return "查看详情";
}

function primaryActionKey(status: string, reviewed = false) {
  if (status === "completed" && !reviewed) return "review";
  if (status === "refunded") return "after_sales";
  return "detail";
}

function secondaryActionText(status: string, reviewed = false) {
  if (status === "rider_assigned" || status === "dispatching" || status === "preparing") return "联系骑手";
  if (status === "completed") return reviewed ? "再来一单" : "再来一单";
  return "";
}

function secondaryActionKey(status: string, reviewed = false) {
  if (status === "rider_assigned" || status === "dispatching" || status === "preparing") return "rider";
  if (status === "completed") return reviewed ? "reorder" : "reorder";
  return "";
}

function etaText(status: string) {
  if (status === "rider_assigned") return "预计 12:52 送达";
  if (status === "dispatching") return "正在匹配骑手";
  if (status === "preparing") return "商家正在备餐";
  if (status === "merchant_pending") return "等待商家接单";
  return "";
}

function summarizeOrderItems(items: Array<{ product_name?: string; quantity?: number }> = []) {
  if (!Array.isArray(items) || items.length === 0) {
    return "订单详情";
  }
  const [first, ...rest] = items;
  if (rest.length === 0) return `${first.product_name || "商品"} ${first.quantity || 1} 件`;
  const totalCount = items.reduce((sum, item) => sum + Number(item.quantity || 0), 0);
  return `${first.product_name || "商品"}等 ${totalCount} 件`;
}

function firstOrderImage(items: Array<{ product_id?: string; image_url?: string }> = []) {
  const first = Array.isArray(items) && items.length > 0 ? items[0] : {};
  return mediaUrl(first.image_url, productFallbackImage(String(first.product_id || "")));
}

function formatOrderTime(value: string) {
  const text = String(value || "").trim();
  if (!text) return "刚刚";
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) return "刚刚";
  const now = new Date();
  const sameDay = date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate();
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  if (sameDay) return `今天 ${hour}:${minute}`;
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${month}-${day} ${hour}:${minute}`;
}
