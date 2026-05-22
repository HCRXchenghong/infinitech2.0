<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view class="title-block">
        <text class="title">资质资料</text>
        <text class="subtitle">{{ canAcceptOrders ? "店铺可接单" : "补齐资质和保证金后可接单" }}</text>
      </view>
      <button class="refresh-button" @click="refresh">刷新</button>
    </view>

    <view class="status-panel" :class="{ warn: !canAcceptOrders }">
      <view>
        <text class="panel-title">{{ shopName }}</text>
        <text class="status-line">{{ statusText }}</text>
      </view>
      <button v-if="depositStatus !== 'paid'" class="primary-button small" @click="payDeposit">缴保证金</button>
    </view>

    <view class="section">
      <view class="section-head">
        <text>主体资质</text>
        <text class="muted">{{ missingText }}</text>
      </view>
      <view class="two-actions">
        <button class="ghost-button" @click="submitQualification('business_license')">补营业执照</button>
        <button class="ghost-button" @click="submitQualification('health_certificate')">补健康证</button>
      </view>
    </view>

    <view class="section">
      <view class="section-head">
        <text>员工健康证</text>
        <text class="muted">{{ staff.length }} 人</text>
      </view>
      <view v-for="item in staff" :key="item.id" class="list-row">
        <view>
          <text class="row-title">{{ item.name }}</text>
          <text class="row-sub">{{ item.role || "staff" }} · {{ item.status }}</text>
        </view>
        <text class="row-date">{{ shortDate(item.health_certificate_expires_at) }}</text>
      </view>
      <view class="form-grid">
        <input class="input" v-model="staffForm.name" placeholder="员工姓名" />
        <input class="input" v-model="staffForm.phone" placeholder="手机号" />
        <input class="input" v-model="staffForm.health_certificate_url" placeholder="健康证文件 URL" />
        <input class="input" v-model="staffForm.health_certificate_expires_at" placeholder="失效日期 2027-05-22T00:00:00Z" />
      </view>
      <button class="primary-button" @click="saveStaff">保存员工</button>
    </view>

    <view class="section">
      <view class="section-head">
        <text>补充资料</text>
        <text class="muted">{{ materials.length }} 份</text>
      </view>
      <view v-for="item in materials" :key="item.id" class="list-row">
        <view>
          <text class="row-title">{{ materialLabel(item.type) }}</text>
          <text class="row-sub">{{ item.description || item.status }}</text>
        </view>
        <text class="row-date">{{ shortDate(item.expires_at) }}</text>
      </view>
      <view class="form-grid">
        <input class="input" v-model="materialForm.type" placeholder="资料类型 kitchen_photo" />
        <input class="input" v-model="materialForm.file_url" placeholder="文件 URL" />
        <input class="input" v-model="materialForm.description" placeholder="描述" />
        <input class="input" v-model="materialForm.expires_at" placeholder="失效日期，可留空" />
      </view>
      <button class="primary-button" @click="saveMaterial">保存资料</button>
    </view>

    <view v-if="message" class="toast-line">
      <text>{{ message }}</text>
    </view>
  </view>
</template>

<script>
import {
  getMerchantDeposit,
  getMerchantMaterials,
  getMerchantProfile,
  getMerchantStaff,
  payMerchantDeposit,
  saveMerchantMaterial,
  saveMerchantQualification,
  saveMerchantStaff
} from "../../utils/api.js";

const nextYear = "2027-05-22T00:00:00Z";

export default {
  data() {
    return {
      shopName: "蓝海餐厅",
      canAcceptOrders: false,
      missingQualifications: ["business_license", "health_certificate"],
      depositStatus: "unpaid",
      staff: [],
      materials: [],
      staffForm: {
        shop_id: "shop_1",
        name: "李四",
        phone: "13900000000",
        role: "kitchen",
        health_certificate_url: "https://cdn.example.test/staff-health.jpg",
        health_certificate_expires_at: nextYear
      },
      materialForm: {
        shop_id: "shop_1",
        type: "kitchen_photo",
        file_url: "https://cdn.example.test/kitchen.jpg",
        description: "后厨照",
        expires_at: nextYear
      },
      message: ""
    };
  },
  computed: {
    statusText() {
      const deposit = this.depositStatus === "paid" ? "保证金已缴" : "保证金未缴";
      const qualification = this.missingQualifications.length === 0 ? "资质完整" : "资质待补";
      return `${qualification} · ${deposit}`;
    },
    missingText() {
      if (this.missingQualifications.length === 0) {
        return "已补齐";
      }
      return this.missingQualifications.map((item) => this.qualificationLabel(item)).join(" / ");
    }
  },
  onLoad() {
    this.refresh();
  },
  methods: {
    async refresh() {
      try {
        const [profile, staff, materials, deposit] = await Promise.all([
          getMerchantProfile(),
          getMerchantStaff(),
          getMerchantMaterials(),
          getMerchantDeposit()
        ]);
        this.shopName = profile?.account?.display_name || this.shopName;
        this.canAcceptOrders = Boolean(profile?.can_accept_orders);
        this.missingQualifications = profile?.missing_qualifications || [];
        this.staff = Array.isArray(staff) ? staff : [];
        this.materials = Array.isArray(materials) ? materials : [];
        this.depositStatus = deposit?.status || this.depositStatus;
      } catch (_error) {
        this.staff = this.staff.length ? this.staff : [{
          id: "staff_demo",
          name: "张三",
          role: "负责人",
          status: "active",
          health_certificate_expires_at: nextYear
        }];
        this.materials = this.materials.length ? this.materials : [{
          id: "material_demo",
          type: "storefront_photo",
          description: "门头照",
          status: "submitted",
          expires_at: nextYear
        }];
      }
    },
    async submitQualification(type) {
      try {
        const profile = await saveMerchantQualification({
          type,
          file_url: `https://cdn.example.test/${type}.jpg`,
          expires_at: nextYear
        });
        this.missingQualifications = profile?.missing_qualifications || [];
        this.canAcceptOrders = Boolean(profile?.can_accept_orders);
        this.message = `${this.qualificationLabel(type)}已提交`;
      } catch (error) {
        this.message = error.message || "资质提交失败";
      }
    },
    async saveStaff() {
      try {
        await saveMerchantStaff({ ...this.staffForm });
        this.message = "员工资料已保存";
        await this.refresh();
      } catch (error) {
        this.message = error.message || "员工资料保存失败";
      }
    },
    async saveMaterial() {
      try {
        await saveMerchantMaterial({ ...this.materialForm });
        this.message = "补充资料已保存";
        await this.refresh();
      } catch (error) {
        this.message = error.message || "补充资料保存失败";
      }
    },
    async payDeposit() {
      try {
        const deposit = await payMerchantDeposit();
        this.depositStatus = deposit?.status || "paid";
        this.message = "保证金已缴纳";
        await this.refresh();
      } catch (error) {
        this.message = error.message || "保证金缴纳失败";
      }
    },
    qualificationLabel(type) {
      const labels = {
        business_license: "营业执照",
        health_certificate: "健康证",
        supplemental_document: "补充资料"
      };
      return labels[type] || type;
    },
    materialLabel(type) {
      const labels = {
        storefront_photo: "门头照",
        kitchen_photo: "后厨照",
        permit: "许可证"
      };
      return labels[type] || type;
    },
    shortDate(value) {
      return String(value || "").slice(0, 10) || "无失效日";
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
  padding: 28rpx 28rpx 48rpx;
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
.status-panel,
.section,
.toast-line {
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.status-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin-top: 28rpx;
  padding: 24rpx;
}
.status-panel.warn {
  border-color: #fed7aa;
  background: #fff7ed;
}
.panel-title {
  display: block;
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.status-line {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 24rpx;
}
.section {
  margin-top: 22rpx;
  padding: 24rpx;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  color: #101828;
  font-size: 29rpx;
  font-weight: 700;
}
.muted {
  color: #667085;
  font-size: 23rpx;
  font-weight: 400;
}
.two-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14rpx;
  margin-top: 18rpx;
}
.list-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-top: 16rpx;
  padding: 18rpx;
  border: 1rpx solid #eef2f7;
  border-radius: 8rpx;
  background: #f8fafc;
}
.row-title {
  display: block;
  color: #101828;
  font-size: 27rpx;
  font-weight: 700;
}
.row-sub,
.row-date {
  color: #667085;
  font-size: 23rpx;
}
.row-sub {
  display: block;
  margin-top: 6rpx;
}
.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12rpx;
  margin-top: 18rpx;
}
.input {
  box-sizing: border-box;
  width: 100%;
  height: 72rpx;
  padding: 0 18rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #101828;
  font-size: 24rpx;
}
.primary-button,
.ghost-button {
  line-height: 68rpx;
}
.primary-button {
  background: #009bf5;
  color: #ffffff;
}
.primary-button.small {
  width: 150rpx;
  flex-shrink: 0;
}
.ghost-button {
  border: 1rpx solid #d0d5dd;
  background: #ffffff;
  color: #0077c8;
}
.section > .primary-button {
  margin-top: 16rpx;
}
.toast-line {
  position: fixed;
  right: 28rpx;
  bottom: 28rpx;
  max-width: 560rpx;
  padding: 18rpx 22rpx;
  color: #0077c8;
  font-size: 24rpx;
  box-shadow: 0 16rpx 40rpx rgba(15, 23, 42, 0.12);
}
</style>
