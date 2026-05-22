<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view class="title-block">
        <text class="title">站长工作台</text>
        <text class="subtitle">{{ riders.length }} 名骑手 · {{ orders.length }} 笔调度单</text>
      </view>
      <button class="refresh-button" @click="refresh">刷新</button>
    </view>

    <view class="task-panel">
      <view class="task-field">
        <text class="field-label">任务时长</text>
        <input class="task-input" type="number" v-model.number="taskDurationMinutes" />
      </view>
      <view class="task-field">
        <text class="field-label">固定单量</text>
        <input class="task-input" type="number" v-model.number="fixedOrderCount" />
      </view>
      <button class="save-button" @click="saveTask">保存</button>
    </view>

    <view class="section-head">
      <text>在线骑手</text>
      <text class="muted">优先级派单</text>
    </view>
    <scroll-view scroll-x class="rider-strip">
      <view
        v-for="rider in riders"
        :key="rider.id"
        class="rider-chip"
        :class="{ active: selectedRiderId === rider.id }"
        @click="selectedRiderId = rider.id"
      >
        <text class="rider-name">{{ rider.id }}</text>
        <text class="rider-meta">{{ rider.online ? "在线" : "离线" }} · {{ rider.capacity }} 单</text>
        <text class="rider-level">L{{ performanceFor(rider.id).level || "-" }} · {{ performanceFor(rider.id).average_accept_seconds || 0 }}s</text>
      </view>
    </scroll-view>

    <view class="section-head">
      <text>待调度订单</text>
      <text class="muted">{{ selectedRiderId || "未选骑手" }}</text>
    </view>

    <scroll-view scroll-y class="order-list">
      <view v-if="orders.length === 0" class="empty">
        <text>暂无调度订单</text>
      </view>
      <view v-for="order in orders" :key="order.id" class="order-card">
        <view class="order-head">
          <view>
            <text class="order-id">#{{ order.id }}</text>
            <text class="order-type">{{ typeLabel(order.type) }}</text>
          </view>
          <text class="status">{{ statusLabel(order.status) }}</text>
        </view>
        <view class="order-line">
          <text>{{ order.items && order.items.length ? order.items[0].product_name : "跑腿/外卖订单" }}</text>
        </view>
        <view class="order-foot">
          <text class="amount">¥{{ fenToYuan(order.amount_fen) }}</text>
          <button class="primary-button" :disabled="!selectedRiderId" @click="assign(order.id)">派给所选骑手</button>
        </view>
      </view>
    </scroll-view>

    <view v-if="message" class="toast-line">
      <text>{{ message }}</text>
    </view>
  </view>
</template>

<script>
import { getStationOrders, getStationRiderPerformance, getStationRiders, getStationTaskConfig, manualAssignOrder, saveStationTaskConfig } from "../../utils/api.js";

const fallbackRiders = [
  { id: "rider_1", online: true, capacity: 3, dispatch_priority: 300 },
  { id: "rider_2", online: true, capacity: 2, dispatch_priority: 200 }
];

const fallbackOrders = [
  { id: "ord_demo_1", type: "takeout", status: "dispatching", amount_fen: 3599, items: [{ product_name: "招牌牛肉饭" }] },
  { id: "ord_demo_2", type: "courier", status: "dispatching", amount_fen: 1200, items: [] }
];

const fallbackTaskConfig = {
  daily_task_duration_minutes: 480,
  daily_fixed_order_count: 30
};

const fallbackPerformance = [
  { rider_id: "rider_1", level: "B", dispatch_priority: 200, average_accept_seconds: 18 },
  { rider_id: "rider_2", level: "A", dispatch_priority: 300, average_accept_seconds: 12 }
];

export default {
  data() {
    return {
      riders: [],
      orders: [],
      performanceMap: {},
      selectedRiderId: "",
      taskDurationMinutes: 480,
      fixedOrderCount: 30,
      message: ""
    };
  },
  onLoad() {
    this.refresh();
  },
  methods: {
    async refresh() {
      try {
        const [riders, orders, taskConfig, performance] = await Promise.all([getStationRiders(), getStationOrders(), getStationTaskConfig(), getStationRiderPerformance()]);
        this.riders = riders;
        this.orders = orders;
        this.performanceMap = this.toPerformanceMap(performance);
        this.taskDurationMinutes = Number(taskConfig.daily_task_duration_minutes || 0);
        this.fixedOrderCount = Number(taskConfig.daily_fixed_order_count || 0);
        if (!this.selectedRiderId && riders.length > 0) {
          this.selectedRiderId = riders[0].id;
        }
      } catch (_error) {
        this.riders = fallbackRiders;
        this.orders = fallbackOrders;
        this.performanceMap = this.toPerformanceMap(fallbackPerformance);
        this.taskDurationMinutes = fallbackTaskConfig.daily_task_duration_minutes;
        this.fixedOrderCount = fallbackTaskConfig.daily_fixed_order_count;
        this.selectedRiderId = this.selectedRiderId || "rider_1";
      }
    },
    toPerformanceMap(performance) {
      return (performance || []).reduce((acc, item) => {
        acc[item.rider_id] = item;
        return acc;
      }, {});
    },
    performanceFor(riderId) {
      return this.performanceMap[riderId] || {};
    },
    async saveTask() {
      try {
        const config = await saveStationTaskConfig({
          daily_task_duration_minutes: Number(this.taskDurationMinutes || 0),
          daily_fixed_order_count: Number(this.fixedOrderCount || 0)
        });
        this.taskDurationMinutes = Number(config.daily_task_duration_minutes || 0);
        this.fixedOrderCount = Number(config.daily_fixed_order_count || 0);
        this.message = "任务配置已保存";
      } catch (error) {
        this.message = error.message || "任务配置保存失败";
      }
    },
    async assign(orderId) {
      try {
        const result = await manualAssignOrder(orderId, this.selectedRiderId);
        this.message = `已派给 ${result?.order?.rider_id || this.selectedRiderId}`;
        await this.refresh();
      } catch (error) {
        this.message = error.message || "手动派单失败";
      }
    },
    back() {
      uni.navigateBack({ delta: 1 });
    },
    statusLabel(status) {
      const labels = {
        dispatching: "待派单",
        rider_assigned: "已派单"
      };
      return labels[status] || status;
    },
    typeLabel(type) {
      const labels = {
        takeout: "外卖",
        medicine: "买药",
        courier: "快递",
        errand_buy: "代买",
        errand_deliver: "代送"
      };
      return labels[type] || "订单";
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
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 30rpx;
  color: #101828;
  font-size: 28rpx;
  font-weight: 700;
}
.task-panel {
  display: grid;
  grid-template-columns: 1fr 1fr 128rpx;
  gap: 14rpx;
  margin-top: 28rpx;
  padding: 18rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.task-field {
  min-width: 0;
}
.field-label {
  display: block;
  color: #667085;
  font-size: 22rpx;
}
.task-input {
  height: 58rpx;
  margin-top: 8rpx;
  padding: 0 14rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  color: #101828;
  font-size: 25rpx;
}
.save-button {
  align-self: end;
  background: #e5f5ff;
  color: #0077c8;
  line-height: 58rpx;
}
.muted {
  color: #667085;
  font-size: 23rpx;
  font-weight: 400;
}
.rider-strip {
  margin-top: 16rpx;
  white-space: nowrap;
}
.rider-chip {
  display: inline-flex;
  flex-direction: column;
  width: 220rpx;
  margin-right: 14rpx;
  padding: 18rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.rider-chip.active {
  border-color: #009bf5;
  background: #e5f5ff;
}
.rider-name {
  color: #101828;
  font-size: 26rpx;
  font-weight: 700;
}
.rider-meta {
  margin-top: 8rpx;
  color: #667085;
  font-size: 22rpx;
}
.rider-level {
  margin-top: 8rpx;
  color: #0077c8;
  font-size: 22rpx;
  font-weight: 700;
}
.order-list {
  height: calc(100vh - 360rpx);
  margin-top: 16rpx;
}
.order-card,
.empty,
.toast-line {
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.order-card {
  margin-bottom: 16rpx;
  padding: 22rpx;
}
.order-head,
.order-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}
.order-id {
  display: block;
  color: #101828;
  font-size: 28rpx;
  font-weight: 700;
}
.order-type,
.order-line {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 23rpx;
}
.status {
  color: #079455;
  font-size: 24rpx;
}
.amount {
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.primary-button {
  min-width: 210rpx;
  background: #009bf5;
  color: #ffffff;
  line-height: 68rpx;
}
.primary-button[disabled] {
  color: #98a2b3;
  background: #f2f4f7;
}
.empty {
  padding: 48rpx 24rpx;
  color: #667085;
  text-align: center;
}
.toast-line {
  position: fixed;
  right: 28rpx;
  bottom: 28rpx;
  left: 28rpx;
  padding: 18rpx 20rpx;
  color: #344054;
  font-size: 24rpx;
}
</style>
