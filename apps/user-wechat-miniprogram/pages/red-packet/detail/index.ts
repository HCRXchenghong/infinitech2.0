import { claimRedPacket, ensurePreviewAuth, getRedPacketDetail, refundRedPacket } from "../../../utils/api";

Page({
  data: {
    packetId: "",
    title: "蓝海餐厅商户群红包",
    subtitle: "小林发出的拼手气红包",
    message: "一起拼单，吃得开心！",
    amount: "30.00",
    quantity: 6,
    claimedCount: 4,
    remainingCount: 2,
    claimedAmount: "20.86",
    pendingAmount: "9.14",
    progressWidth: "67%",
    statusText: "领取中",
    timerText: "剩余 2 个 · 23:18:42 后退回",
    riskState: "passed",
    riskTip: "领取前会校验频次、金额和红包状态",
    records: [
      { key: "1", initial: "小", name: "小林", amount: "6.18", time: "14:08", best: true },
      { key: "2", initial: "阿", name: "阿杰", amount: "4.52", time: "14:09", best: false },
      { key: "3", initial: "M", name: "Mia", amount: "5.36", time: "14:11", best: false },
      { key: "4", initial: "张", name: "张三", amount: "4.80", time: "14:13", best: false },
      { key: "5", initial: "?", name: "等待领取", amount: "--", time: "--", best: false }
    ]
  },
  onLoad(options) {
    const packetId = String(options?.id || options?.packet_id || "");
    this.setData({ packetId });
    if (packetId) {
      this.loadPacket(packetId);
    }
  },
  async loadPacket(packetId) {
    ensurePreviewAuth();
    try {
      const detail = await getRedPacketDetail(packetId) as { packet?: Record<string, unknown>; shares?: Array<Record<string, unknown>>; risk?: Record<string, unknown> };
      this.applyPacketDetail(detail);
    } catch (_error) {
      this.setData({ statusText: "领取中" });
    }
  },
  applyPacketDetail(detail) {
    const packet = detail?.packet || {};
    const shares = Array.isArray(detail?.shares) ? detail.shares : [];
    const risk = (detail?.risk || {}) as Record<string, unknown>;
    const total = Number(packet.total_amount_fen || 0);
    const claimedShares = shares.filter((item) => item.claimed_at);
    const claimed = claimedShares.reduce((sum, item) => sum + Number(item.amount_fen || 0), 0);
    const quantity = Number(packet.quantity || shares.length || this.data.quantity || 1);
    const progress = quantity > 0 ? Math.round((claimedShares.length / quantity) * 100) : 0;
    const status = String(packet.status || "created");
    const remaining = Math.max(0, quantity - claimedShares.length);
    const pending = status === "refunded" || status === "expired_refunded" ? 0 : Math.max(0, total - claimed);
    this.setData({
      title: "蓝海餐厅商户群红包",
      subtitle: `${String(packet.sender_id || "小林")}发出的${packet.type === "fixed" ? "普通红包" : "拼手气红包"}`,
      message: String(packet.message || "一起拼单，吃得开心！"),
      amount: (total / 100 || Number(this.data.amount)).toFixed(2),
      quantity,
      claimedCount: claimedShares.length,
      remainingCount: remaining,
      claimedAmount: (claimed / 100).toFixed(2),
      pendingAmount: (pending / 100).toFixed(2),
      progressWidth: `${Math.max(8, progress)}%`,
      statusText: statusText(status),
      timerText: timerText(status, remaining, String(packet.expires_at || "")),
      riskState: String(risk.state || "passed"),
      riskTip: String(risk.reason || "领取前会校验频次、金额和红包状态"),
      records: shares.map((item, index) => {
        const user = String(item.user_id || "");
        return {
          key: `${user || "pending"}_${index}`,
          initial: user ? user.slice(0, 1).toUpperCase() : "?",
          name: user || "等待领取",
          amount: item.claimed_at ? (Number(item.amount_fen || 0) / 100).toFixed(2) : "--",
          time: item.claimed_at ? String(item.claimed_at).slice(11, 16) : "--",
          best: index === 0 && Boolean(item.claimed_at)
        };
      })
    });
  },
  async handleClaim() {
    if (!this.data.packetId) {
      wx.showToast({ title: "预览红包已领取", icon: "none" });
      return;
    }
    ensurePreviewAuth();
    try {
      const result = await claimRedPacket(this.data.packetId) as { detail?: Record<string, unknown>; risk?: Record<string, unknown> };
      this.applyPacketDetail(result.detail || result);
      wx.showToast({ title: "领取成功", icon: "success" });
    } catch (error) {
      const tip = redPacketRiskTip(error);
      this.setData({ riskState: "blocked", riskTip: tip });
      wx.showToast({ title: tip.slice(0, 18), icon: "none" });
    }
  },
  async handleRefund() {
    if (!this.data.packetId) {
      wx.showToast({ title: "预览余额已退回", icon: "none" });
      return;
    }
    ensurePreviewAuth();
    try {
      const detail = await refundRedPacket(this.data.packetId);
      this.applyPacketDetail(detail);
      wx.showToast({ title: "未领余额已退回", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "仅发红包人可退回", icon: "none" });
    }
  },
  handleContinue() {
    wx.navigateTo({ url: "/pages/red-packet/send/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/messages/merchant-group/index" }) });
  }
});

function statusText(status: string) {
  switch (status) {
    case "finished":
      return "已领完";
    case "expired_refunded":
      return "已退回";
    case "refunded":
      return "已退回";
    default:
      return "领取中";
  }
}

function redPacketRiskTip(error: unknown) {
  if (error instanceof Error && error.message) {
    return error.message.replace(/^risk control rejected:\s*/i, "");
  }
  return "红包暂不可领取，请稍后再试";
}

function timerText(status: string, remaining: number, expiresAt: string) {
  if (status === "finished") return "已全部领取 · 金额已进入余额";
  if (status === "refunded" || status === "expired_refunded") return "未领金额已退回余额";
  const expires = expiresAt ? new Date(expiresAt).getTime() : 0;
  const seconds = Math.max(0, Math.floor((expires - Date.now()) / 1000));
  const hours = String(Math.floor(seconds / 3600)).padStart(2, "0");
  const minutes = String(Math.floor((seconds % 3600) / 60)).padStart(2, "0");
  const secs = String(seconds % 60).padStart(2, "0");
  return `剩余 ${remaining} 个 · ${hours}:${minutes}:${secs} 后退回`;
}
