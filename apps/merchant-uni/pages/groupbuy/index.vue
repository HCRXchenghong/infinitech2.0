<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view>
        <text class="title">团购核销</text>
        <text class="subtitle">到店扫码验券</text>
      </view>
      <button class="scan-button" @click="scan">扫码</button>
    </view>

    <view class="panel">
      <text class="label">券码或二维码内容</text>
      <input class="input" v-model="manualPayload" placeholder="粘贴用户团购券二维码内容" />
      <button class="primary-button" @click="submitManual">核销</button>
    </view>

    <view v-if="lastResult" class="result-card">
      <text class="result-title">核销成功</text>
      <text class="result-line">券码：{{ lastResult.voucher.voucher_code || lastResult.voucher.id }}</text>
      <text class="result-line">套餐：{{ lastResult.voucher.deal_name || "团购套餐" }}</text>
      <text class="result-line">订单：{{ lastResult.order.id }}</text>
    </view>

    <view v-if="errorMessage" class="error-card">
      <text>{{ errorMessage }}</text>
    </view>
  </view>
</template>

<script>
import { scanGroupbuyVoucher } from "../../utils/api.js";

export default {
  data() {
    return {
      manualPayload: "",
      lastResult: null,
      errorMessage: ""
    };
  },
  methods: {
    scan() {
      uni.scanCode({
        onlyFromCamera: true,
        success: async (result) => {
          await this.redeem(result.result);
        },
        fail: () => {
          this.errorMessage = "未读取到二维码";
        }
      });
    },
    async submitManual() {
      await this.redeem(this.manualPayload);
    },
    async redeem(payload) {
      this.errorMessage = "";
      const normalized = String(payload || "").trim();
      if (!normalized) {
        this.errorMessage = "请输入券码";
        return;
      }
      try {
        this.lastResult = await scanGroupbuyVoucher(normalized);
        this.manualPayload = "";
      } catch (error) {
        this.errorMessage = error?.message || "核销失败";
      }
    },
    back() {
      uni.navigateBack({ delta: 1 });
    }
  }
};
</script>

<style>
.page {
  min-height: 100vh;
  padding: 28rpx;
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
.scan-button,
.primary-button {
  background: #009bf5;
  color: #ffffff;
}
.scan-button {
  line-height: 64rpx;
}
.panel,
.result-card,
.error-card {
  margin-top: 24rpx;
  padding: 24rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.label {
  display: block;
  color: #344054;
  font-size: 25rpx;
  font-weight: 700;
}
.input {
  box-sizing: border-box;
  width: 100%;
  height: 80rpx;
  margin-top: 16rpx;
  padding: 0 20rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #101828;
  font-size: 25rpx;
}
.primary-button {
  margin-top: 18rpx;
  line-height: 72rpx;
}
.result-title {
  display: block;
  color: #009bf5;
  font-size: 32rpx;
  font-weight: 700;
}
.result-line {
  display: block;
  margin-top: 10rpx;
  color: #344054;
  font-size: 25rpx;
}
.error-card {
  color: #b42318;
  background: #fff1f3;
}
</style>
