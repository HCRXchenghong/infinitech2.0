<template>
  <view class="page">
    <view class="topbar">
      <view>
        <text class="eyebrow">Infinitech 骑手</text>
        <text class="title">抢单大厅</text>
      </view>
      <view class="online-pill" :class="{ off: !online }">
        <text>{{ online ? "在线" : "离线" }}</text>
      </view>
    </view>

    <view class="metric-grid">
      <view class="metric">
        <text class="metric-value">{{ capacity }}</text>
        <text class="metric-label">可接容量</text>
      </view>
      <view class="metric">
        <text class="metric-value">{{ distanceMeters }}m</text>
        <text class="metric-label">上报距离</text>
      </view>
      <view class="metric">
        <text class="metric-value">{{ freeCancelLabel }}</text>
        <text class="metric-label">免责拒单</text>
      </view>
    </view>

    <view class="control-row">
      <button class="primary-button" @click="toggleOnline">{{ online ? "下线" : "上线" }}</button>
      <button class="ghost-button" @click="consumeCancel">免责取消</button>
      <button class="ghost-button" @click="goStation">站长台</button>
    </view>

    <view class="deposit-panel">
      <view>
        <text class="panel-title">保证金</text>
        <text class="deposit-status">{{ depositStatusLabel }}</text>
      </view>
      <view class="deposit-actions">
        <button class="ghost-button" @click="payDeposit">缴纳</button>
        <button class="ghost-button" @click="wechatExempt">免押</button>
        <button class="ghost-button" @click="refundDeposit">退押</button>
      </view>
    </view>

    <view class="panel">
      <view class="panel-head">
        <text class="panel-title">当前派单</text>
        <text class="status">{{ currentOrderId ? "待处理" : "暂无" }}</text>
      </view>
      <input class="order-input" v-model="currentOrderId" placeholder="订单 ID" />
      <view class="panel-foot action-grid">
        <button class="primary-button" :disabled="!currentOrderId" @click="pickupOrder">确认取货</button>
        <button class="primary-button" :disabled="!currentOrderId" @click="deliverOrder">送达完成</button>
        <button class="danger-button" :disabled="!currentOrderId" @click="rejectOrder">拒绝派单</button>
      </view>
    </view>

    <view v-if="message" class="toast-line">
      <text>{{ message }}</text>
    </view>
  </view>
</template>

<script>
import { applyWechatDepositExemption, consumeFreeCancel, getRiderDeposit, markOrderDelivered, markOrderPickedUp, payRiderDeposit, rejectAssignedOrder, requestRiderDepositRefund, setRiderOnline } from "../../utils/api.js";

export default {
  data() {
    return {
      online: true,
      capacity: 2,
      distanceMeters: 500,
      currentOrderId: "ord_demo_1",
      depositStatus: "paid",
      freeCancelUsed: false,
      message: ""
    };
  },
  computed: {
    freeCancelLabel() {
      return this.freeCancelUsed ? "已用" : "可用";
    },
    depositStatusLabel() {
      const labels = {
        unpaid: "未缴纳",
        paid: "已缴纳",
        wechat_exempt_approved: "微信免押",
        refund_pending: "退押处理中",
        refunded: "已退还",
        dispute_hold: "纠纷冻结"
      };
      return labels[this.depositStatus] || this.depositStatus;
    }
  },
  onShow() {
    this.refreshDeposit();
  },
  methods: {
    async refreshDeposit() {
      try {
        const deposit = await getRiderDeposit();
        this.depositStatus = deposit.status || this.depositStatus;
      } catch (_error) {
        this.depositStatus = this.depositStatus || "unpaid";
      }
    },
    async toggleOnline() {
      try {
        const nextOnline = !this.online;
        const rider = await setRiderOnline({
          online: nextOnline,
          capacity: this.capacity,
          distance_meters: this.distanceMeters
        });
        this.online = Boolean(rider.online);
        this.message = this.online ? "已上线接单" : "已下线";
      } catch (error) {
        this.message = error.message || "状态更新失败";
      }
    },
    async consumeCancel() {
      try {
        const result = await consumeFreeCancel();
        this.freeCancelUsed = true;
        this.message = result.allowed ? "今日免责取消已使用" : "今日免责取消已用完";
      } catch (error) {
        this.message = error.message || "免责取消失败";
      }
    },
    async payDeposit() {
      try {
        const deposit = await payRiderDeposit();
        this.depositStatus = deposit.status;
        this.message = "保证金已缴纳";
      } catch (error) {
        this.message = error.message || "保证金缴纳失败";
      }
    },
    async wechatExempt() {
      try {
        const result = await applyWechatDepositExemption();
        this.depositStatus = result?.deposit?.status || "wechat_exempt_approved";
        this.message = "微信免押已通过";
      } catch (error) {
        this.message = error.message || "微信免押申请失败";
      }
    },
    async refundDeposit() {
      try {
        const result = await requestRiderDepositRefund();
        this.depositStatus = result?.deposit?.status || "refund_pending";
        this.message = "退押申请已提交";
      } catch (error) {
        this.message = error.message || "退押申请失败";
      }
    },
    async rejectOrder() {
      try {
        const result = await rejectAssignedOrder(this.currentOrderId);
        const nextRider = result?.decision?.candidate_rider_id || "暂无";
        const decision = result?.decision || {};
        const quotaText = decision.can_decline_without_penalty
          ? `，已完成${decision.daily_completed_order_count || 0}/${decision.daily_fixed_order_count || 0}单，本次免责`
          : "";
        this.message = `已拒绝，下一位：${nextRider}${quotaText}`;
        this.currentOrderId = result?.order?.id || "";
      } catch (error) {
        this.message = error.message || "拒绝派单失败";
      }
    },
    async pickupOrder() {
      try {
        const order = await markOrderPickedUp(this.currentOrderId);
        this.message = `订单 ${order.id} 已取货`;
      } catch (error) {
        this.message = error.message || "确认取货失败";
      }
    },
    async deliverOrder() {
      try {
        const order = await markOrderDelivered(this.currentOrderId);
        this.message = `订单 ${order.id} 已完成`;
        this.currentOrderId = "";
      } catch (error) {
        this.message = error.message || "送达完成失败";
      }
    },
    goStation() {
      uni.navigateTo({ url: "/pages/station/index" });
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
.online-pill {
  padding: 12rpx 18rpx;
  border-radius: 8rpx;
  background: #e5f5ff;
  color: #0077c8;
  font-size: 24rpx;
  font-weight: 600;
}
.online-pill.off {
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
.panel,
.toast-line {
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
  font-size: 36rpx;
  font-weight: 700;
}
.metric-label {
  display: block;
  margin-top: 6rpx;
  color: #667085;
  font-size: 23rpx;
}
.control-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14rpx;
  margin-top: 28rpx;
}
button {
  margin: 0;
  border-radius: 8rpx;
  font-size: 24rpx;
}
button::after {
  border: 0;
}
.primary-button {
  background: #009bf5;
  color: #ffffff;
  line-height: 72rpx;
}
.ghost-button {
  background: #ffffff;
  color: #0077c8;
  line-height: 72rpx;
}
.danger-button {
  width: 100%;
  background: #fee2e2;
  color: #b42318;
  line-height: 72rpx;
}
.action-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12rpx;
}
.danger-button[disabled] {
  color: #98a2b3;
  background: #f2f4f7;
}
.panel {
  margin-top: 28rpx;
  padding: 24rpx;
}
.deposit-panel {
  display: grid;
  grid-template-columns: 1fr 330rpx;
  align-items: center;
  gap: 16rpx;
  margin-top: 24rpx;
  padding: 22rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.deposit-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10rpx;
}
.deposit-actions .ghost-button {
  line-height: 58rpx;
}
.deposit-status {
  display: block;
  margin-top: 8rpx;
  color: #0077c8;
  font-size: 24rpx;
  font-weight: 700;
}
.panel-head,
.panel-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.panel-title {
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.status {
  color: #079455;
  font-size: 24rpx;
}
.order-input {
  height: 76rpx;
  margin: 24rpx 0;
  padding: 0 20rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #101828;
  font-size: 26rpx;
}
.toast-line {
  margin-top: 20rpx;
  padding: 18rpx 20rpx;
  color: #344054;
  font-size: 24rpx;
}
</style>
