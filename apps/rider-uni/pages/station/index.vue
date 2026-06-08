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
        <text class="rider-score">{{ riderRatingText(performanceFor(rider.id)) }}</text>
      </view>
    </scroll-view>

    <view class="section-head">
      <text>绩效拆解</text>
      <text class="muted">{{ selectedRiderId || "未选骑手" }}</text>
    </view>
    <view v-if="selectedRiderId" class="performance-panel">
      <view class="performance-summary">
        <view>
          <text class="performance-score">{{ selectedPerformance.score || 0 }}</text>
          <text class="performance-score-label">当前派单分</text>
        </view>
        <view class="performance-baseline">
          <text>团队接单均值 {{ formatDecimal(selectedPerformance.score_breakdown?.team_average_accept_seconds, 1) }}s</text>
          <text>团队日均单量 {{ formatDecimal(selectedPerformance.score_breakdown?.team_average_daily_orders, 1) }}</text>
        </view>
      </view>
      <view class="performance-grid">
        <view v-for="item in performanceBreakdownItems" :key="item.label" class="performance-item">
          <text class="performance-item-label">{{ item.label }}</text>
          <text class="performance-item-value">{{ item.value }}</text>
        </view>
      </view>
      <view class="performance-extra">
        <view class="performance-block">
          <text class="performance-block-title">近 3 日趋势</text>
          <view class="performance-trend-list">
            <text v-for="item in recentTrendItems" :key="item.label" class="performance-trend-chip">{{ item.label }}</text>
          </view>
        </view>
        <view class="performance-block">
          <text class="performance-block-title">最近评价</text>
          <view v-if="recentReviewItems.length > 0" class="performance-review-list">
            <text v-for="item in recentReviewItems" :key="item.id" class="performance-review-item">{{ item.text }}</text>
          </view>
          <text v-else class="performance-review-empty">最近暂无骑手评价</text>
        </view>
        <view class="performance-block">
          <text class="performance-block-title">异常履约</text>
          <view class="performance-trend-list">
            <text v-for="item in exceptionItems" :key="item.label" class="performance-trend-chip">{{ item.label }}</text>
          </view>
          <view v-if="exceptionDetailItems.length > 0" class="performance-review-list">
            <text v-for="item in exceptionDetailItems" :key="item.id" class="performance-review-item">{{ item.text }}</text>
          </view>
          <text v-else class="performance-review-empty">最近暂无异常履约明细</text>
        </view>
      </view>
    </view>

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
  {
    rider_id: "rider_1",
    level: "B",
    dispatch_priority: 200,
    average_accept_seconds: 18,
    rider_average_rating: 4.7,
    rider_review_count: 18,
      score: 92,
      score_breakdown: {
        accept_score: 35.4,
        order_volume_score: 31.2,
        completion_score: 14.2,
        rating_score: 11.3,
        rating_confidence: 1,
        team_average_accept_seconds: 15,
        team_average_daily_orders: 24
      },
      recent_trend: [
        { date: "2026-05-30", score: 88, completed_orders: 2, average_rating: 4.5 },
        { date: "2026-05-31", score: 90, completed_orders: 3, average_rating: 4.7, reject_count: 1 },
        { date: "2026-06-01", score: 92, completed_orders: 2, average_rating: 4.8 }
      ],
      recent_reviews: [
        { review_id: "rev_demo_1", rider_rating: 5, content: "送达准时，态度很好。", created_at: "2026-06-01T09:20:00Z" }
      ],
      exception_summary: { dispatch_timeout_count: 0, dispatch_reject_count: 1, after_sales_count: 0, low_rating_count: 0, last_event_at: "2026-05-31T20:10:00Z" },
      exception_details: [
        { kind: "dispatch_reject", label: "骑手拒单", order_id: "ord_demo_2", status: "dispatch.rejected", message: "高峰期路线冲突，系统已顺延下一位", created_at: "2026-05-31T20:10:00Z" }
      ]
    },
  {
    rider_id: "rider_2",
    level: "A",
    dispatch_priority: 300,
    average_accept_seconds: 12,
    rider_average_rating: 4.9,
    rider_review_count: 32,
      score: 116,
      score_breakdown: {
        accept_score: 44.8,
        order_volume_score: 33.6,
        completion_score: 14.8,
        rating_score: 11.8,
        rating_confidence: 1,
        team_average_accept_seconds: 15,
        team_average_daily_orders: 24
      },
      recent_trend: [
        { date: "2026-05-30", score: 110, completed_orders: 4, average_rating: 4.8 },
        { date: "2026-05-31", score: 113, completed_orders: 5, average_rating: 4.9 },
        { date: "2026-06-01", score: 116, completed_orders: 4, average_rating: 4.9 }
      ],
      recent_reviews: [
        { review_id: "rev_demo_2", rider_rating: 5, content: "配送很快，包装也完整。", created_at: "2026-06-01T08:40:00Z" },
        { review_id: "rev_demo_3", rider_rating: 4, content: "雨天也按时送到。", created_at: "2026-05-31T19:30:00Z" }
      ],
      exception_summary: { dispatch_timeout_count: 0, dispatch_reject_count: 0, after_sales_count: 0, low_rating_count: 0 },
      exception_details: []
    }
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
  computed: {
    selectedPerformance() {
      return this.performanceFor(this.selectedRiderId);
    },
    performanceBreakdownItems() {
      const breakdown = this.selectedPerformance?.score_breakdown || {};
      return [
        { label: "接单加分", value: this.formatScorePart(breakdown.accept_score) },
        { label: "单量加分", value: this.formatScorePart(breakdown.order_volume_score) },
        { label: "履约加分", value: this.formatScorePart(breakdown.completion_score) },
        { label: "评分加分", value: this.formatScorePart(breakdown.rating_score) },
        { label: "评分置信度", value: this.formatPercentValue(breakdown.rating_confidence) }
      ];
    },
    recentTrendItems() {
      const trend = this.selectedPerformance?.recent_trend || [];
      if (trend.length === 0) {
        return [{ label: "最近 3 天暂无趋势样本" }];
      }
      return trend.map((item) => {
        const dateLabel = String(item.date || "").slice(5) || "--";
        const rating = Number(item.average_rating || 0) > 0 ? `${Number(item.average_rating).toFixed(1)}星` : "无评分";
        return {
          label: `${dateLabel} ${Math.round(Number(item.score || 0))}分 · ${Number(item.completed_orders || 0)}单 · ${rating}`
        };
      });
    },
    recentReviewItems() {
      const reviews = this.selectedPerformance?.recent_reviews || [];
      return reviews.map((item) => {
        const rating = Number(item.rider_rating || item.rating || 0);
        return {
          id: item.review_id || item.created_at || item.content,
          text: `${rating > 0 ? `${rating}星` : "无评分"} · ${item.content || "无评价内容"}`
        };
      });
    },
    exceptionItems() {
      const summary = this.selectedPerformance?.exception_summary || {};
      return [
        { label: `超时 ${Number(summary.dispatch_timeout_count || 0)}` },
        { label: `拒单 ${Number(summary.dispatch_reject_count || 0)}` },
        { label: `售后 ${Number(summary.after_sales_count || 0)}` },
        { label: `低分 ${Number(summary.low_rating_count || 0)}` }
      ];
    },
    exceptionDetailItems() {
      const details = this.selectedPerformance?.exception_details || [];
      return details.map((item) => {
        const createdAt = this.formatShortTime(item.created_at);
        const orderID = item.order_id ? ` · ${item.order_id}` : "";
        const status = item.status ? ` · ${item.status}` : "";
        return {
          id: item.dispatch_event_id || item.after_sales_request_id || item.review_id || item.created_at || item.message,
          text: `${item.label || "异常履约"}${orderID}${status} · ${item.message || "待补详情"} · ${createdAt}`
        };
      });
    }
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
    riderRatingText(performance) {
      const rating = Number(performance?.rider_average_rating || 0);
      const reviewCount = Number(performance?.rider_review_count || 0);
      const dispatchScore = Number(performance?.score || 0);
      if (rating <= 0 || reviewCount <= 0) {
        return "暂无配送评分";
      }
      const scoreText = dispatchScore > 0 ? ` · 派单分 ${dispatchScore}` : "";
      return `配送评分 ${rating.toFixed(1)} · ${reviewCount} 条${scoreText}`;
    },
    formatDecimal(value, precision = 0) {
      const parsed = Number(value || 0);
      if (parsed <= 0) {
        return "0";
      }
      return parsed.toFixed(precision);
    },
    formatScorePart(value) {
      const parsed = Number(value || 0);
      if (parsed <= 0) {
        return "0";
      }
      return `${Math.round(parsed)} 分`;
    },
    formatPercentValue(value) {
      const parsed = Number(value || 0);
      if (parsed <= 0) {
        return "0%";
      }
      return `${Math.round(parsed * 100)}%`;
    },
    formatShortTime(value) {
      if (!value) {
        return "-";
      }
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) {
        return "-";
      }
      const month = `${date.getMonth() + 1}`.padStart(2, "0");
      const day = `${date.getDate()}`.padStart(2, "0");
      const hours = `${date.getHours()}`.padStart(2, "0");
      const minutes = `${date.getMinutes()}`.padStart(2, "0");
      return `${month}/${day} ${hours}:${minutes}`;
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
.rider-score {
  margin-top: 8rpx;
  color: #f59e0b;
  font-size: 21rpx;
}
.performance-panel {
  margin-top: 16rpx;
  padding: 24rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.performance-summary {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}
.performance-score {
  display: block;
  color: #101828;
  font-size: 50rpx;
  font-weight: 700;
}
.performance-score-label {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 22rpx;
}
.performance-baseline {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  color: #475467;
  font-size: 22rpx;
  text-align: right;
}
.performance-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-top: 22rpx;
}
.performance-extra {
  margin-top: 22rpx;
}
.performance-block + .performance-block {
  margin-top: 20rpx;
}
.performance-block-title {
  display: block;
  color: #475467;
  font-size: 22rpx;
  font-weight: 700;
}
.performance-trend-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 12rpx;
}
.performance-trend-chip {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #eef4ff;
  color: #344054;
  font-size: 20rpx;
}
.performance-review-list {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  margin-top: 12rpx;
}
.performance-review-item,
.performance-review-empty {
  color: #344054;
  font-size: 22rpx;
  line-height: 1.5;
}
.performance-item {
  padding: 18rpx;
  border-radius: 8rpx;
  background: #f8fafc;
}
.performance-item-label {
  display: block;
  color: #667085;
  font-size: 22rpx;
}
.performance-item-value {
  display: block;
  margin-top: 8rpx;
  color: #101828;
  font-size: 28rpx;
  font-weight: 700;
}
.order-list {
  height: calc(100vh - 620rpx);
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
