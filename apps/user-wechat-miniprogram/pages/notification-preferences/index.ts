import {
  getUserNotificationPreferences,
  saveUserNotificationPreference,
  setDevAuthToken
} from "../../utils/api";

const CHANNELS = [
  { key: "wechat_subscribe", label: "微信订阅", hint: "订单、售后关键状态优先使用" },
  { key: "sms", label: "短信", hint: "账号与履约异常兜底触达" },
  { key: "push", label: "App Push", hint: "活动、优惠和状态提醒" }
];

const DEFAULT_TYPES = [
  {
    value: "order.status_changed",
    label: "订单状态",
    description: "支付、接单、配送、完成和退款状态。"
  },
  {
    value: "after_sales.updated",
    label: "售后进度",
    description: "商户处理、平台仲裁、退款到账和补充证据。"
  },
  {
    value: "coupon.campaign",
    label: "优惠活动",
    description: "优惠券、团购活动、商户群权益和平台补贴。"
  }
];

type ChannelState = {
  key: string;
  label: string;
  hint: string;
  enabled: boolean;
  quiet: boolean;
};

type PreferenceState = {
  value: string;
  label: string;
  description: string;
  statusText: string;
  channels: ChannelState[];
  quiet: {
    enabled: boolean;
    start: string;
    end: string;
    timezone_offset: string;
  };
  saving: boolean;
};

function createPreferenceState(type: typeof DEFAULT_TYPES[number]): PreferenceState {
  return {
    ...type,
    statusText: type.value === "coupon.campaign" ? "部分开启" : "已开启",
    channels: CHANNELS.map((channel) => ({
      ...channel,
      enabled: true,
      quiet: channel.key !== "sms"
    })),
    quiet: {
      enabled: false,
      start: "22:00",
      end: "08:00",
      timezone_offset: "+08:00"
    },
    saving: false
  };
}

function applyPreferenceRecord(type: PreferenceState, record: Record<string, any> | undefined): PreferenceState {
  if (!record) {
    return type;
  }
  const enabledChannels = Array.isArray(record.enabled_channels) ? record.enabled_channels : [];
  const disabledChannels = Array.isArray(record.disabled_channels) ? record.disabled_channels : [];
  const quietHours = record.quiet_hours && typeof record.quiet_hours === "object" ? record.quiet_hours : {};
  const quietChannels = Array.isArray(quietHours.channels) ? quietHours.channels : [];
  return {
    ...type,
    statusText: "已保存",
    channels: type.channels.map((channel) => ({
      ...channel,
      enabled: enabledChannels.length > 0 ? enabledChannels.includes(channel.key) : !disabledChannels.includes(channel.key),
      quiet: quietChannels.includes(channel.key)
    })),
    quiet: {
      enabled: Boolean(quietHours.enabled),
      start: quietHours.start || type.quiet.start,
      end: quietHours.end || type.quiet.end,
      timezone_offset: quietHours.timezone_offset || "+08:00"
    }
  };
}

Page({
  data: {
    types: DEFAULT_TYPES.map(createPreferenceState)
  },
  onLoad() {
    this.loadPreferences();
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  },
  async loadPreferences() {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    const nextTypes = [...this.data.types];
    await Promise.all(nextTypes.map(async (type, index) => {
      try {
        const preferences = await getUserNotificationPreferences(type.value, 1);
        const record = Array.isArray(preferences) ? preferences[0] : undefined;
        nextTypes[index] = applyPreferenceRecord(type, record);
      } catch (_error) {
        nextTypes[index] = { ...type, statusText: "本地预览" };
      }
    }));
    this.setData({ types: nextTypes });
  },
  handleChannelChange(event: any) {
    const typeIndex = Number(event.currentTarget.dataset.typeIndex);
    const channelKey = String(event.currentTarget.dataset.channel || "");
    const checked = Boolean(event.detail.value);
    this.updateType(typeIndex, (type) => ({
      ...type,
      statusText: "待保存",
      channels: type.channels.map((channel) => channel.key === channelKey ? { ...channel, enabled: checked } : channel)
    }));
  },
  handleQuietEnabledChange(event: any) {
    const typeIndex = Number(event.currentTarget.dataset.typeIndex);
    const checked = Boolean(event.detail.value);
    this.updateType(typeIndex, (type) => ({
      ...type,
      statusText: "待保存",
      quiet: { ...type.quiet, enabled: checked }
    }));
  },
  handleQuietInput(event: any) {
    const typeIndex = Number(event.currentTarget.dataset.typeIndex);
    const field = String(event.currentTarget.dataset.field || "");
    const value = String(event.detail.value || "");
    if (field !== "start" && field !== "end") {
      return;
    }
    this.updateType(typeIndex, (type) => ({
      ...type,
      statusText: "待保存",
      quiet: { ...type.quiet, [field]: value }
    }));
  },
  handleQuietChannelToggle(event: WechatMiniprogram.BaseEvent) {
    const typeIndex = Number(event.currentTarget.dataset.typeIndex);
    const channelKey = String(event.currentTarget.dataset.channel || "");
    this.updateType(typeIndex, (type) => ({
      ...type,
      statusText: "待保存",
      channels: type.channels.map((channel) => channel.key === channelKey ? { ...channel, quiet: !channel.quiet } : channel)
    }));
  },
  async handleSave(event: WechatMiniprogram.BaseEvent) {
    const typeIndex = Number(event.currentTarget.dataset.typeIndex);
    const type = this.data.types[typeIndex] as PreferenceState | undefined;
    if (!type) {
      return;
    }
    this.updateType(typeIndex, (current) => ({ ...current, saving: true }));
    const enabledChannels = type.channels.filter((channel) => channel.enabled).map((channel) => channel.key);
    const disabledChannels = type.channels.filter((channel) => !channel.enabled).map((channel) => channel.key);
    const quietChannels = type.channels.filter((channel) => channel.quiet).map((channel) => channel.key);
    try {
      await saveUserNotificationPreference({
        notification_type: type.value,
        enabled_channels: enabledChannels,
        disabled_channels: disabledChannels,
        quiet_hours: {
          enabled: type.quiet.enabled,
          start: type.quiet.start,
          end: type.quiet.end,
          timezone_offset: type.quiet.timezone_offset,
          channels: quietChannels
        },
        updated_at: new Date().toISOString()
      });
      this.updateType(typeIndex, (current) => ({ ...current, statusText: "已保存", saving: false }));
      wx.showToast({ title: "已保存", icon: "success" });
    } catch (_error) {
      this.updateType(typeIndex, (current) => ({ ...current, statusText: "保存失败", saving: false }));
      wx.showToast({ title: "保存失败", icon: "none" });
    }
  },
  updateType(typeIndex: number, updater: (type: PreferenceState) => PreferenceState) {
    const types = [...this.data.types] as PreferenceState[];
    const current = types[typeIndex];
    if (!current) {
      return;
    }
    types[typeIndex] = updater(current);
    this.setData({ types });
  }
});
