import { ensurePreviewAuth, getMessageThreads } from "../../utils/api";

type MessageRow = {
  key: string;
  iconPath: string;
  tone: string;
  title: string;
  subtitle: string;
  time: string;
  badge: string;
  unreadDot: boolean;
  muted: boolean;
  route: string;
};

const DEFAULT_ROWS: MessageRow[] = [
  {
    key: "merchant",
    iconPath: "/assets/generated/messages/message-merchant.png",
    tone: "merchant",
    title: "商家消息",
    subtitle: "蓝海餐厅：您的订单已接单",
    time: "刚刚",
    badge: "",
    unreadDot: true,
    muted: false,
    route: "/pages/messages/merchant-group/index?thread_id=merchant_blue_sea"
  },
  {
    key: "rider",
    iconPath: "/assets/generated/messages/message-rider.png",
    tone: "rider",
    title: "骑手消息",
    subtitle: "李师傅：我已到取餐点",
    time: "2分钟前",
    badge: "",
    unreadDot: true,
    muted: false,
    route: "/pages/messages/merchant-group/index?thread_id=rider_zhang"
  },
  {
    key: "group",
    iconPath: "/assets/generated/messages/message-user.png",
    tone: "group",
    title: "群聊",
    subtitle: "望京SOHO拼饭群：今晚一起拼单？",
    time: "8分钟前",
    badge: "3",
    unreadDot: false,
    muted: false,
    route: "/pages/messages/merchant-group/index?thread_id=official&type=official"
  },
  {
    key: "service",
    iconPath: "/assets/generated/messages/message-service-logo.jpg",
    tone: "service",
    title: "官方客服",
    subtitle: "退款进度已更新",
    time: "10分钟前",
    badge: "",
    unreadDot: false,
    muted: false,
    route: "/pages/customer-service/chat/index"
  }
];

Page({
  data: {
    keyword: "",
    rows: DEFAULT_ROWS,
    visibleRows: DEFAULT_ROWS
  },
  onLoad() {
    this.applyKeywordFilter();
  },
  onShow() {
    this.loadThreads();
  },
  async loadThreads() {
    ensurePreviewAuth();
    try {
      const threads = await getMessageThreads() as Array<Record<string, unknown>>;
      if (!Array.isArray(threads) || threads.length === 0) {
        this.applyKeywordFilter();
        return;
      }
      const rows = mergeThreadRows(threads);
      this.setData({ rows });
      this.applyKeywordFilter(rows);
    } catch (_error) {
      this.applyKeywordFilter();
    }
  },
  handleKeywordInput(event) {
    const keyword = String(event.detail.value || "").trim();
    this.setData({ keyword });
    this.applyKeywordFilter(this.data.rows, keyword);
  },
  handleClearKeyword() {
    this.setData({ keyword: "" });
    this.applyKeywordFilter(this.data.rows, "");
  },
  handleRowTap(event) {
    wx.navigateTo({ url: String(event.currentTarget.dataset.route || "/pages/messages/merchant-group/index") });
  },
  applyKeywordFilter(rows = this.data.rows, keyword = this.data.keyword) {
    const text = String(keyword || "").trim();
    const sourceRows = Array.isArray(rows) ? rows : DEFAULT_ROWS;
    const visibleRows = text
      ? sourceRows.filter((row) => `${row.title}${row.subtitle}`.includes(text))
      : sourceRows;
    this.setData({ visibleRows });
  }
});

function mergeThreadRows(threads: Array<Record<string, unknown>>) {
  const merchant = pickThread(threads, isMerchantThread);
  const rider = pickThread(threads, isRiderThread);
  const group = pickThread(threads, isGroupThread);
  const service = pickThread(threads, isServiceThread);

  return DEFAULT_ROWS.map((row) => {
    const thread = { merchant, rider, group, service }[row.key as "merchant" | "rider" | "group" | "service"];
    if (!thread) return row;
    return rowFromThread(row, thread);
  });
}

function rowFromThread(row: MessageRow, thread: Record<string, unknown>) {
  const unreadCount = Number(thread.unread_count || 0);
  const route = String(thread.route || row.route);
  const subtitle = compactSubtitle(thread, row.subtitle);
  return {
    ...row,
    subtitle,
    time: formatThreadTime(String(thread.updated_at || "")) || row.time,
    badge: row.key === "group" && unreadCount > 0 ? String(unreadCount) : "",
    unreadDot: row.key !== "group" && unreadCount > 0,
    muted: Boolean(thread.muted),
    route
  };
}

function pickThread(threads: Array<Record<string, unknown>>, matcher: (thread: Record<string, unknown>) => boolean) {
  return threads.find(matcher);
}

function isMerchantThread(thread: Record<string, unknown>) {
  const type = String(thread.type || "");
  const title = String(thread.title || "");
  const id = String(thread.id || "");
  return type === "merchant" || /商家|商户|餐厅|咖啡|店/.test(title) || /merchant/.test(id);
}

function isRiderThread(thread: Record<string, unknown>) {
  const type = String(thread.type || "");
  const title = String(thread.title || "");
  const id = String(thread.id || "");
  return type === "rider" || /骑手|配送|师傅/.test(title) || /rider/.test(id);
}

function isGroupThread(thread: Record<string, unknown>) {
  const type = String(thread.type || "");
  const title = String(thread.title || "");
  return type === "official" || type === "group" || /群|拼饭/.test(title);
}

function isServiceThread(thread: Record<string, unknown>) {
  const type = String(thread.type || "");
  const title = String(thread.title || "");
  const id = String(thread.id || "");
  return type === "customer_service" || /客服|售后/.test(title) || /service/.test(id);
}

function compactSubtitle(thread: Record<string, unknown>, fallback: string) {
  const rawTitle = cleanThreadTitle(String(thread.title || ""));
  const rawSubtitle = String(thread.subtitle || "").trim();
  if (!rawTitle && !rawSubtitle) return fallback;
  if (!rawTitle) return rawSubtitle;
  if (!rawSubtitle) return rawTitle;
  return `${rawTitle}：${rawSubtitle}`;
}

function cleanThreadTitle(title: string) {
  return title
    .replace(/商户群|官方群|官方群|商家群/g, "")
    .replace(/\s+/g, "")
    .trim();
}

function formatThreadTime(value: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  if (diffMs >= 0 && diffMs < 60 * 1000) return "刚刚";
  if (diffMs >= 0 && diffMs < 60 * 60 * 1000) return `${Math.max(1, Math.floor(diffMs / 60000))}分钟前`;
  if (date.toDateString() === now.toDateString()) {
    return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
  }
  return "昨天";
}
