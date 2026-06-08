<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view class="title-block">
        <text class="title">通知偏好</text>
        <text class="subtitle">控制外部触达渠道和静默时段，站内信始终保留。</text>
      </view>
      <button class="refresh-button" :disabled="loading" @click="refresh">{{ loading ? "刷新中" : "刷新" }}</button>
    </view>

    <view class="summary-strip">
      <view>
        <text class="summary-value">{{ enabledSummary }}</text>
        <text class="summary-label">已启用外部渠道</text>
      </view>
      <view>
        <text class="summary-value">{{ quietSummary }}</text>
        <text class="summary-label">静默规则</text>
      </view>
    </view>

    <view v-for="type in notificationTypes" :key="type.value" class="preference-card">
      <view class="card-head">
        <view>
          <text class="card-title">{{ type.label }}</text>
          <text class="card-sub">{{ type.description }}</text>
        </view>
        <text class="status-badge" :class="{ warn: disabledCount(type.value) > 0 }">
          {{ statusText(type.value) }}
        </text>
      </view>

      <view class="field-block">
        <text class="field-title">外部渠道</text>
        <checkbox-group class="channel-grid" @change="onEnabledChannelsChange(type.value, $event)">
          <label v-for="channel in channels" :key="channel.value" class="check-row">
            <checkbox :value="channel.value" :checked="isChannelEnabled(type.value, channel.value)" />
            <text>{{ channel.label }}</text>
          </label>
        </checkbox-group>
      </view>

      <view class="quiet-row">
        <view>
          <text class="field-title">静默时段</text>
          <text class="field-sub">静默内不触达外部渠道，到期后进入可靠重投。</text>
        </view>
        <switch :checked="settingsFor(type.value).quietEnabled" @change="onQuietEnabledChange(type.value, $event)" />
      </view>

      <view v-if="settingsFor(type.value).quietEnabled" class="quiet-panel">
        <view class="time-grid">
          <label class="input-field">
            <text>开始</text>
            <input :value="settingsFor(type.value).quietStart" placeholder="22:00" @input="onFieldChange(type.value, 'quietStart', $event)" />
          </label>
          <label class="input-field">
            <text>结束</text>
            <input :value="settingsFor(type.value).quietEnd" placeholder="08:00" @input="onFieldChange(type.value, 'quietEnd', $event)" />
          </label>
          <label class="input-field">
            <text>时区</text>
            <input :value="settingsFor(type.value).timezoneOffset" placeholder="+08:00" @input="onFieldChange(type.value, 'timezoneOffset', $event)" />
          </label>
        </view>
        <text class="field-title quiet-channel-title">静默渠道</text>
        <checkbox-group class="channel-grid" @change="onQuietChannelsChange(type.value, $event)">
          <label v-for="channel in channels" :key="type.value + ':' + channel.value + ':quiet'" class="check-row">
            <checkbox :value="channel.value" :checked="isQuietChannel(type.value, channel.value)" />
            <text>{{ channel.label }}</text>
          </label>
        </checkbox-group>
      </view>

      <button class="primary-button" :disabled="savingType === type.value" @click="save(type.value)">
        {{ savingType === type.value ? "保存中" : "保存偏好" }}
      </button>
    </view>

    <view v-if="message" class="toast-line">
      <text>{{ message }}</text>
    </view>
  </view>
</template>

<script>
import {
  getMerchantNotificationPreferences,
  saveMerchantNotificationPreference
} from "../../utils/api.js";

const channels = [
  { value: "wechat_subscribe", label: "微信订阅" },
  { value: "sms", label: "短信" },
  { value: "enterprise_wechat", label: "企业微信" },
  { value: "push", label: "端内 Push" }
];

const notificationTypes = [
  {
    value: "order.status_changed",
    label: "订单状态",
    description: "新订单、接单、出餐、配送等订单流转通知。"
  },
  {
    value: "merchant.qualification_reviewed",
    label: "资质审核",
    description: "营业执照、健康证、保证金和准入结果通知。"
  }
];

function allChannelValues() {
  return channels.map((channel) => channel.value);
}

function defaultSettings() {
  return {
    disabledChannels: [],
    quietEnabled: false,
    quietStart: "22:00",
    quietEnd: "08:00",
    timezoneOffset: "+08:00",
    quietChannels: ["wechat_subscribe", "push"]
  };
}

function initialSettings() {
  return notificationTypes.reduce((settings, type) => {
    settings[type.value] = defaultSettings();
    return settings;
  }, {});
}

function normalizeList(value) {
  return Array.isArray(value) ? value.map((item) => String(item || "").trim()).filter(Boolean) : [];
}

function settingsFromPreference(preference) {
  const quiet = preference?.quiet_hours || {};
  return {
    disabledChannels: normalizeList(preference?.disabled_channels),
    quietEnabled: Boolean(quiet.enabled),
    quietStart: quiet.start || "22:00",
    quietEnd: quiet.end || "08:00",
    timezoneOffset: quiet.timezone_offset || "+08:00",
    quietChannels: normalizeList(quiet.channels).length ? normalizeList(quiet.channels) : ["wechat_subscribe", "push"]
  };
}

export default {
  data() {
    return {
      channels,
      notificationTypes,
      settings: initialSettings(),
      loading: false,
      savingType: "",
      message: ""
    };
  },
  computed: {
    enabledSummary() {
      const total = notificationTypes.length * channels.length;
      const disabled = notificationTypes.reduce((sum, type) => sum + this.settingsFor(type.value).disabledChannels.length, 0);
      return `${total - disabled}/${total}`;
    },
    quietSummary() {
      const enabled = notificationTypes.filter((type) => this.settingsFor(type.value).quietEnabled).length;
      return `${enabled}/${notificationTypes.length}`;
    }
  },
  onLoad() {
    this.refresh();
  },
  methods: {
    settingsFor(type) {
      return this.settings[type] || defaultSettings();
    },
    setSettings(type, patch) {
      this.settings = {
        ...this.settings,
        [type]: {
          ...this.settingsFor(type),
          ...patch
        }
      };
    },
    async refresh() {
      this.loading = true;
      this.message = "";
      try {
        const next = initialSettings();
        await Promise.all(notificationTypes.map(async (type) => {
          const preferences = await getMerchantNotificationPreferences(type.value, 1);
          const preference = Array.isArray(preferences) ? preferences[0] : null;
          if (preference) {
            next[type.value] = settingsFromPreference(preference);
          }
        }));
        this.settings = next;
        this.message = "通知偏好已同步";
      } catch (error) {
        this.message = error.message || "通知偏好同步失败，已保留本地预览值";
      } finally {
        this.loading = false;
      }
    },
    async save(type) {
      const current = this.settingsFor(type);
      this.savingType = type;
      this.message = "";
      try {
        await saveMerchantNotificationPreference({
          notification_type: type,
          disabled_channels: current.disabledChannels,
          quiet_hours: {
            enabled: current.quietEnabled,
            start: current.quietStart,
            end: current.quietEnd,
            timezone_offset: current.timezoneOffset,
            channels: current.quietEnabled ? current.quietChannels : []
          },
          updated_at: new Date().toISOString()
        });
        this.message = `${this.typeLabel(type)}偏好已保存`;
      } catch (error) {
        this.message = error.message || "通知偏好保存失败";
      } finally {
        this.savingType = "";
      }
    },
    onEnabledChannelsChange(type, event) {
      const enabled = normalizeList(event.detail?.value);
      const disabled = allChannelValues().filter((channel) => !enabled.includes(channel));
      this.setSettings(type, { disabledChannels: disabled });
    },
    onQuietEnabledChange(type, event) {
      this.setSettings(type, { quietEnabled: Boolean(event.detail?.value) });
    },
    onQuietChannelsChange(type, event) {
      this.setSettings(type, { quietChannels: normalizeList(event.detail?.value) });
    },
    onFieldChange(type, field, event) {
      this.setSettings(type, { [field]: event.detail?.value || "" });
    },
    isChannelEnabled(type, channel) {
      return !this.settingsFor(type).disabledChannels.includes(channel);
    },
    isQuietChannel(type, channel) {
      return this.settingsFor(type).quietChannels.includes(channel);
    },
    disabledCount(type) {
      return this.settingsFor(type).disabledChannels.length;
    },
    statusText(type) {
      const count = this.disabledCount(type);
      return count > 0 ? `关闭 ${count} 个` : "全部开启";
    },
    typeLabel(type) {
      const item = notificationTypes.find((candidate) => candidate.value === type);
      return item ? item.label : type;
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
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.icon-button,
.refresh-button {
  margin: 0;
  border-radius: 8rpx;
  background: #ffffff;
  color: #1f2937;
  font-size: 24rpx;
  line-height: 64rpx;
}
.icon-button {
  width: 64rpx;
  padding: 0;
}
.refresh-button {
  padding: 0 22rpx;
}
button::after {
  border: 0;
}
.title-block {
  flex: 1;
}
.title {
  display: block;
  color: #101828;
  font-size: 40rpx;
  font-weight: 700;
}
.subtitle {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 23rpx;
  line-height: 34rpx;
}
.summary-strip {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16rpx;
  margin-top: 28rpx;
}
.summary-strip > view,
.preference-card {
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.summary-strip > view {
  padding: 22rpx;
}
.summary-value {
  display: block;
  color: #0064a8;
  font-size: 38rpx;
  font-weight: 700;
}
.summary-label {
  display: block;
  margin-top: 6rpx;
  color: #667085;
  font-size: 23rpx;
}
.preference-card {
  margin-top: 22rpx;
  padding: 24rpx;
}
.card-head,
.quiet-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}
.card-title,
.field-title {
  display: block;
  color: #111827;
  font-size: 28rpx;
  font-weight: 700;
}
.card-sub,
.field-sub {
  display: block;
  margin-top: 6rpx;
  color: #667085;
  font-size: 23rpx;
  line-height: 34rpx;
}
.status-badge {
  flex-shrink: 0;
  padding: 10rpx 16rpx;
  border-radius: 8rpx;
  background: #e8f7ef;
  color: #027a48;
  font-size: 22rpx;
  font-weight: 700;
}
.status-badge.warn {
  background: #fff4e5;
  color: #b45309;
}
.field-block,
.quiet-row,
.quiet-panel {
  margin-top: 24rpx;
}
.channel-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12rpx;
  margin-top: 14rpx;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-height: 62rpx;
  padding: 0 14rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  background: #f9fafb;
  color: #344054;
  font-size: 23rpx;
}
.quiet-panel {
  padding: 20rpx;
  border-radius: 8rpx;
  background: #f8fafc;
}
.time-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12rpx;
}
.input-field {
  display: block;
}
.input-field text {
  display: block;
  color: #667085;
  font-size: 22rpx;
}
.input-field input {
  height: 64rpx;
  margin-top: 8rpx;
  padding: 0 14rpx;
  border: 1rpx solid #d0d5dd;
  border-radius: 8rpx;
  background: #ffffff;
  color: #101828;
  font-size: 24rpx;
}
.quiet-channel-title {
  margin-top: 18rpx;
}
.primary-button {
  margin: 24rpx 0 0;
  border-radius: 8rpx;
  background: #009bf5;
  color: #ffffff;
  font-size: 25rpx;
  font-weight: 700;
  line-height: 76rpx;
}
.primary-button[disabled],
.refresh-button[disabled] {
  opacity: 0.55;
}
.toast-line {
  margin-top: 22rpx;
  padding: 18rpx 20rpx;
  border-radius: 8rpx;
  background: #eef6ff;
  color: #005b99;
  font-size: 24rpx;
}
</style>
