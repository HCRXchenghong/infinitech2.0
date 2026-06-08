<template>
  <view class="page">
    <view class="topbar">
      <view>
        <text class="eyebrow">今日经营</text>
        <text class="title">{{ shopName }}</text>
      </view>
      <view class="status-pill" :class="{ warn: !canAcceptOrders }">
        <text>{{ canAcceptOrders ? "可接单" : "待补资料" }}</text>
      </view>
    </view>

    <view class="metric-grid">
      <view class="metric">
        <text class="metric-value">{{ pendingCount }}</text>
        <text class="metric-label">待接单</text>
      </view>
      <view class="metric">
        <text class="metric-value">{{ preparingCount }}</text>
        <text class="metric-label">备餐中</text>
      </view>
      <view class="metric">
        <text class="metric-value">{{ dispatchingCount }}</text>
        <text class="metric-label">待骑手</text>
      </view>
    </view>

    <view class="actions">
      <button class="primary-action" @click="goOrders">订单处理</button>
      <button class="ghost-action" @click="goProducts">商品</button>
      <button class="ghost-action" @click="goGroupbuy">团购</button>
      <button class="ghost-action" @click="goCompliance">资料</button>
      <button class="ghost-action" @click="goNotificationPreferences">通知</button>
      <button class="ghost-action">钱包</button>
    </view>

    <view class="section-head">
      <text>待处理订单</text>
      <button class="link-action" @click="goOrders">查看全部</button>
    </view>

    <view v-if="visibleOrders.length === 0" class="empty">
      <text>暂无待处理订单</text>
    </view>
    <view v-for="order in visibleOrders" :key="order.id" class="order-card">
      <view class="order-main">
        <text class="order-id">#{{ order.id }}</text>
        <text class="order-status">{{ statusLabel(order.status) }}</text>
      </view>
      <text class="order-sub">{{ order.items && order.items.length ? order.items[0].product_name : "外卖订单" }}</text>
      <view class="order-foot">
        <text class="amount">¥{{ fenToYuan(order.amount_fen) }}</text>
        <button v-if="order.status === 'merchant_pending'" class="small-button" @click="accept(order.id)">接单</button>
        <button v-else-if="order.status === 'preparing'" class="small-button" @click="ready(order.id)">出餐</button>
      </view>
    </view>
  </view>
</template>

<script>
import { acceptMerchantOrder, getMerchantOrders, getMerchantProfile, markMerchantOrderReady } from "../../utils/api.js";

const fallbackOrders = [
  {
    id: "ord_demo_1",
    status: "merchant_pending",
    amount_fen: 5598,
    items: [{ product_name: "招牌牛肉饭 x2" }]
  },
  {
    id: "ord_demo_2",
    status: "preparing",
    amount_fen: 3899,
    items: [{ product_name: "团购套餐核销备餐" }]
  }
];

export default {
  data() {
    return {
      shopName: "蓝海餐厅",
      canAcceptOrders: true,
      orders: []
    };
  },
  computed: {
    pendingCount() {
      return this.orders.filter((item) => item.status === "merchant_pending").length;
    },
    preparingCount() {
      return this.orders.filter((item) => item.status === "preparing").length;
    },
    dispatchingCount() {
      return this.orders.filter((item) => item.status === "dispatching").length;
    },
    visibleOrders() {
      return this.orders
        .filter((item) => ["merchant_pending", "preparing", "dispatching"].includes(item.status))
        .slice(0, 3);
    }
  },
  onShow() {
    this.refresh();
  },
  methods: {
    async refresh() {
      try {
        const profile = await getMerchantProfile();
        this.shopName = profile?.account?.display_name || this.shopName;
        this.canAcceptOrders = Boolean(profile?.can_accept_orders);
        this.orders = await getMerchantOrders();
      } catch (_error) {
        this.orders = fallbackOrders;
      }
    },
    goOrders() {
      uni.navigateTo({ url: "/pages/orders/index" });
    },
    goProducts() {
      uni.navigateTo({ url: "/pages/products/index" });
    },
    goGroupbuy() {
      uni.navigateTo({ url: "/pages/groupbuy/index" });
    },
    goCompliance() {
      uni.navigateTo({ url: "/pages/compliance/index" });
    },
    goNotificationPreferences() {
      uni.navigateTo({ url: "/pages/notification-preferences/index" });
    },
    async accept(orderId) {
      await acceptMerchantOrder(orderId);
      await this.refresh();
    },
    async ready(orderId) {
      await markMerchantOrderReady(orderId);
      await this.refresh();
    },
    statusLabel(status) {
      const labels = {
        merchant_pending: "待接单",
        preparing: "备餐中",
        dispatching: "待骑手",
        rider_assigned: "已接骑手"
      };
      return labels[status] || status;
    },
    fenToYuan(value) {
      return (Number(value || 0) / 100).toFixed(2);
    }
  }
};
</script>

<style>
.page {
  min-height: 100vh;
  padding: 32rpx 28rpx 48rpx;
  background: #f4f7fb;
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
}
.eyebrow {
  display: block;
  color: #667085;
  font-size: 24rpx;
}
.title {
  display: block;
  margin-top: 8rpx;
  color: #101828;
  font-size: 42rpx;
  font-weight: 700;
}
.status-pill {
  padding: 12rpx 18rpx;
  border-radius: 8rpx;
  background: #e5f5ff;
  color: #0077c8;
  font-size: 24rpx;
  font-weight: 600;
}
.status-pill.warn {
  background: #fff4e5;
  color: #b45309;
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16rpx;
  margin-top: 32rpx;
}
.metric,
.order-card,
.empty {
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.metric {
  padding: 22rpx 18rpx;
}
.metric-value {
  display: block;
  color: #009bf5;
  font-size: 42rpx;
  font-weight: 700;
}
.metric-label {
  display: block;
  margin-top: 6rpx;
  color: #667085;
  font-size: 23rpx;
}
.actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14rpx;
  margin-top: 28rpx;
}
button {
  margin: 0;
  border-radius: 8rpx;
  font-size: 24rpx;
  line-height: 72rpx;
}
button::after {
  border: 0;
}
.primary-action {
  background: #009bf5;
  color: #ffffff;
}
.ghost-action {
  background: #ffffff;
  color: #1f2937;
  border: 1rpx solid #d0d5dd;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 38rpx 0 18rpx;
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.link-action {
  padding: 0;
  background: transparent;
  color: #009bf5;
  font-size: 24rpx;
  line-height: 48rpx;
}
.empty {
  padding: 48rpx 24rpx;
  color: #98a2b3;
  text-align: center;
}
.order-card {
  margin-bottom: 16rpx;
  padding: 24rpx;
}
.order-main,
.order-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}
.order-id {
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.order-status {
  color: #009bf5;
  font-size: 24rpx;
  font-weight: 600;
}
.order-sub {
  display: block;
  margin-top: 12rpx;
  color: #667085;
  font-size: 24rpx;
}
.order-foot {
  margin-top: 18rpx;
}
.amount {
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.small-button {
  width: 136rpx;
  background: #009bf5;
  color: #ffffff;
  font-size: 24rpx;
  line-height: 56rpx;
}
</style>
