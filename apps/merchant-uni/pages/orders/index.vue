<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view class="title-block">
        <text class="title">订单处理</text>
        <text class="subtitle">{{ orders.length }} 笔订单</text>
      </view>
      <button class="refresh-button" @click="refresh">刷新</button>
    </view>

    <view class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        class="tab"
        :class="{ active: currentTab === tab.value }"
        @click="currentTab = tab.value"
      >
        {{ tab.label }}
      </button>
    </view>

    <scroll-view scroll-y class="order-list">
      <view v-if="filteredOrders.length === 0" class="empty">
        <text>暂无订单</text>
      </view>
      <view v-for="order in filteredOrders" :key="order.id" class="order-card">
        <view class="order-head">
          <view>
            <text class="order-id">#{{ order.id }}</text>
            <text class="order-type">{{ order.type || "takeout" }}</text>
          </view>
          <text class="status">{{ statusLabel(order.status) }}</text>
        </view>

        <view class="items">
          <text v-for="item in order.items || []" :key="item.product_id" class="item-line">
            {{ item.product_name }} x{{ item.quantity || 1 }}
          </text>
          <text v-if="!order.items || order.items.length === 0" class="item-line">外卖订单</text>
        </view>

        <view class="meta-row">
          <text>实付 ¥{{ fenToYuan(order.amount_fen) }}</text>
          <text>{{ order.payment_method || "待支付" }}</text>
        </view>

        <view class="timeline">
          <text v-for="event in (order.events || []).slice(-3)" :key="event.type + event.created_at" class="event">
            {{ event.message }}
          </text>
        </view>

        <view class="actions">
          <button v-if="order.status === 'merchant_pending'" class="primary-button" @click="accept(order.id)">接单</button>
          <button v-if="order.status === 'preparing'" class="primary-button" @click="ready(order.id)">出餐</button>
          <button class="ghost-button">联系用户</button>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { acceptMerchantOrder, getMerchantOrders, markMerchantOrderReady } from "../../utils/api.js";

const fallbackOrders = [
  {
    id: "ord_demo_1",
    type: "takeout",
    status: "merchant_pending",
    amount_fen: 5598,
    payment_method: "balance",
    items: [{ product_id: "prod_beef_rice", product_name: "招牌牛肉饭", quantity: 2 }],
    events: [{ type: "order.payment.success", message: "支付成功，订单进入商户待接单", created_at: "2026-05-21T08:00:00Z" }]
  },
  {
    id: "ord_demo_2",
    type: "takeout",
    status: "preparing",
    amount_fen: 3199,
    payment_method: "wechat_pay",
    items: [{ product_id: "prod_soup", product_name: "每日例汤", quantity: 1 }],
    events: [{ type: "merchant.accepted", message: "商户已接单，开始备餐", created_at: "2026-05-21T08:01:00Z" }]
  },
  {
    id: "ord_demo_3",
    type: "takeout",
    status: "dispatching",
    amount_fen: 4599,
    payment_method: "balance",
    items: [{ product_id: "prod_set", product_name: "工作餐套餐", quantity: 1 }],
    events: [{ type: "merchant.ready_for_pickup", message: "商户已出餐，订单进入骑手调度", created_at: "2026-05-21T08:03:00Z" }]
  }
];

export default {
  data() {
    return {
      currentTab: "active",
      tabs: [
        { label: "处理中", value: "active" },
        { label: "待接单", value: "merchant_pending" },
        { label: "备餐中", value: "preparing" },
        { label: "待骑手", value: "dispatching" }
      ],
      orders: []
    };
  },
  computed: {
    filteredOrders() {
      if (this.currentTab === "active") {
        return this.orders.filter((item) => ["merchant_pending", "preparing", "dispatching"].includes(item.status));
      }
      return this.orders.filter((item) => item.status === this.currentTab);
    }
  },
  onLoad() {
    this.refresh();
  },
  methods: {
    async refresh() {
      try {
        this.orders = await getMerchantOrders();
      } catch (_error) {
        this.orders = fallbackOrders;
      }
    },
    async accept(orderId) {
      await acceptMerchantOrder(orderId);
      await this.refresh();
    },
    async ready(orderId) {
      await markMerchantOrderReady(orderId);
      await this.refresh();
    },
    back() {
      uni.navigateBack({ delta: 1 });
    },
    statusLabel(status) {
      const labels = {
        merchant_pending: "待接单",
        preparing: "备餐中",
        dispatching: "待骑手",
        rider_assigned: "骑手已接",
        completed: "已完成",
        cancelled: "已取消"
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
  padding: 28rpx 28rpx 0;
  background: #f4f7fb;
}
.topbar {
  display: grid;
  grid-template-columns: 72rpx 1fr 112rpx;
  align-items: center;
  gap: 16rpx;
}
.title {
  display: block;
  color: #101828;
  font-size: 38rpx;
  font-weight: 700;
}
.subtitle {
  display: block;
  margin-top: 4rpx;
  color: #667085;
  font-size: 23rpx;
}
button {
  margin: 0;
  border-radius: 8rpx;
  font-size: 24rpx;
}
button::after {
  border: 0;
}
.icon-button {
  width: 72rpx;
  height: 72rpx;
  background: #ffffff;
  color: #101828;
  font-size: 42rpx;
  line-height: 64rpx;
}
.refresh-button {
  background: #e5f5ff;
  color: #0077c8;
  line-height: 64rpx;
}
.tabs {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12rpx;
  margin-top: 28rpx;
}
.tab {
  background: #ffffff;
  color: #475467;
  line-height: 64rpx;
}
.tab.active {
  background: #009bf5;
  color: #ffffff;
  font-weight: 700;
}
.order-list {
  height: calc(100vh - 180rpx);
  margin-top: 20rpx;
  padding-bottom: 36rpx;
}
.empty,
.order-card {
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.empty {
  padding: 60rpx 24rpx;
  color: #98a2b3;
  text-align: center;
}
.order-card {
  margin-bottom: 16rpx;
  padding: 24rpx;
}
.order-head,
.meta-row,
.actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.order-id {
  display: block;
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.order-type {
  display: block;
  margin-top: 4rpx;
  color: #98a2b3;
  font-size: 22rpx;
}
.status {
  color: #009bf5;
  font-size: 24rpx;
  font-weight: 700;
}
.items {
  margin-top: 18rpx;
}
.item-line {
  display: block;
  color: #344054;
  font-size: 25rpx;
  line-height: 40rpx;
}
.meta-row {
  margin-top: 18rpx;
  color: #101828;
  font-size: 25rpx;
  font-weight: 600;
}
.timeline {
  margin-top: 18rpx;
  padding: 14rpx 16rpx;
  border-radius: 8rpx;
  background: #f8fafc;
}
.event {
  display: block;
  color: #667085;
  font-size: 22rpx;
  line-height: 34rpx;
}
.actions {
  justify-content: flex-end;
  margin-top: 20rpx;
}
.primary-button,
.ghost-button {
  width: 150rpx;
  line-height: 58rpx;
}
.primary-button {
  background: #009bf5;
  color: #ffffff;
}
.ghost-button {
  border: 1rpx solid #d0d5dd;
  background: #ffffff;
  color: #344054;
}
</style>
