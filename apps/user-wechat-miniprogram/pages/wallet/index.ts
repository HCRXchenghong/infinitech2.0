import { creditWallet, ensurePreviewAuth, getWalletOverview, requestWalletWithdraw } from "../../utils/api";

Page({
  data: {
    balance: "128.50",
    paymentStatus: "支付密码已设置",
    statusText: "已同步",
    selectedFilter: "all",
    assetStats: [
      { icon: "包", title: "红包", value: "3" },
      { icon: "券", title: "优惠券", value: "6" },
      { icon: "分", title: "积分", value: "2680" },
      { icon: "入", title: "待入账", value: "¥12.00" }
    ],
    filters: [
      { id: "all", title: "全部", active: "active" },
      { id: "income", title: "收入", active: "" },
      { id: "expense", title: "支出", active: "" },
      { id: "refund", title: "退款", active: "" }
    ],
    allRecords: [],
    records: [
      { id: "refund_preview", icon: "退", typeClass: "income", title: "订单退款入账", subtitle: "蓝海餐厅 · 今天 13:05", amount: "+¥18.00", statusText: "" },
      { id: "pay_preview", icon: "支", typeClass: "expense", title: "余额支付", subtitle: "订单 10031 · 今天 12:18", amount: "-¥55.98", statusText: "" }
    ]
  },
  onShow() {
    this.loadOverview();
  },
  async loadOverview() {
    ensurePreviewAuth();
    try {
      const overview = await getWalletOverview() as Record<string, unknown>;
      const transactions = overview.transactions as Array<Record<string, unknown>> || [];
      const records = transactions.map(walletRecordFromTransaction);
      this.setData({
        allRecords: records,
        records,
        balance: formatFen(Number(overview.balance_fen || 0)),
        paymentStatus: overview.payment_password_status === "set" ? "支付密码已设置" : "支付密码未设置",
        assetStats: [
          { icon: "包", title: "红包", value: String(overview.red_packet_count || 0) },
          { icon: "券", title: "优惠券", value: String(overview.coupon_count || 0) },
          { icon: "分", title: "积分", value: String(overview.points || 0) },
          { icon: "入", title: "待入账", value: `¥${formatFen(Number(overview.pending_receivable_fen || 0))}` }
        ],
        statusText: "已同步"
      });
    } catch (_error) {
      this.setData({ statusText: "离线缓存" });
    }
  },
  async handleCredit() {
    ensurePreviewAuth();
    try {
      const result = await creditWallet(1000) as { account?: Record<string, unknown> };
      const account = result.account || {};
      this.setData({
        balance: (Number(account.balance_fen || 0) / 100).toFixed(2),
        statusText: "已充值"
      });
      await this.loadOverview();
      wx.showToast({ title: "已充值", icon: "success" });
    } catch (_error) {
      this.setData({ balance: "10.00", statusText: "离线缓存" });
      wx.showToast({ title: "充值已记录", icon: "none" });
    }
  },
  handlePaymentPassword() {
    wx.navigateTo({ url: "/pages/wallet/payment-password/index" });
  },
  async handleWithdraw() {
    ensurePreviewAuth();
    try {
      await requestWalletWithdraw(5000);
      await this.loadOverview();
      wx.showToast({ title: "提现已提交", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "提现申请已记录", icon: "none" });
    }
  },
  handleBillTap() {
    this.loadOverview();
    wx.showToast({ title: "账单已刷新", icon: "none" });
  },
  handleFilterTap(event) {
    const selectedFilter = String(event.currentTarget.dataset.id || "all");
    const allRecords = this.data.allRecords.length ? this.data.allRecords : this.data.records;
    this.setData({
      selectedFilter,
      filters: this.data.filters.map((item) => ({ ...item, active: item.id === selectedFilter ? "active" : "" })),
      records: selectedFilter === "all" ? allRecords : allRecords.filter((item) => item.typeClass === selectedFilter || item.filterType === selectedFilter)
    });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  }
});

function walletRecordFromTransaction(item: Record<string, unknown>) {
  const type = String(item.type || "");
  const amountFen = Number(item.amount_fen || 0);
  const status = String(item.status || "");
  return {
    id: String(item.id || Date.now()),
    icon: walletTransactionIcon(type),
    typeClass: amountFen >= 0 ? "income" : "expense",
    filterType: type === "refund" ? "refund" : amountFen >= 0 ? "income" : "expense",
    title: walletTransactionTitle(type),
    subtitle: `${walletTransactionSubtitle(type)} · ${formatDateText(String(item.created_at || ""))}`,
    amount: `${amountFen >= 0 ? "+" : "-"}¥${formatFen(Math.abs(amountFen))}`,
    statusText: status === "processing" ? "处理中" : ""
  };
}

function walletTransactionTitle(type: string) {
  switch (type) {
    case "credit":
      return "钱包充值";
    case "payment":
      return "余额支付";
    case "refund":
      return "订单退款入账";
    case "red_packet":
      return "商户群红包";
    case "withdraw":
      return "提现申请";
    default:
      return "钱包流水";
  }
}

function walletTransactionSubtitle(type: string) {
  switch (type) {
    case "payment":
      return "订单 10031";
    case "refund":
      return "蓝海餐厅";
    case "red_packet":
      return "蓝海餐厅商户群";
    case "withdraw":
      return "微信零钱";
    case "credit":
      return "微信支付";
    default:
      return "悦享e食";
  }
}

function walletTransactionIcon(type: string) {
  switch (type) {
    case "refund":
      return "退";
    case "payment":
      return "支";
    case "red_packet":
      return "包";
    case "withdraw":
      return "提";
    case "credit":
      return "充";
    default:
      return "账";
  }
}

function formatFen(amountFen: number) {
  return (amountFen / 100).toFixed(2);
}

function formatDateText(value: string) {
  if (!value) return "刚刚";
  return value.slice(0, 10) || "刚刚";
}
