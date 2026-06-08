const DEFAULT_API_BASE_URL = "http://127.0.0.1:25500";

type RequestOptions = {
  method?: "GET" | "POST" | "PUT";
  data?: Record<string, unknown>;
  auth?: boolean;
};

export function getApiBaseUrl() {
  return String(wx.getStorageSync("apiBaseUrl") || DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export function getRealtimeBaseUrl() {
  const stored = String(wx.getStorageSync("realtimeBaseUrl") || "").trim();
  if (stored) return stored.replace(/\/$/, "");
  return getApiBaseUrl()
    .replace(/^https:/, "wss:")
    .replace(/^http:/, "ws:")
    .replace(/:25500$/, ":9898");
}

export function getRealtimeSocketUrl(path = "/ws", query: Record<string, string> = {}) {
  const params = Object.keys(query)
    .filter((key) => query[key])
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(query[key])}`)
    .join("&");
  return `${getRealtimeBaseUrl()}${path}${params ? `?${params}` : ""}`;
}

export function getAuthToken() {
  return String(wx.getStorageSync("authToken") || "");
}

export function getCurrentUserId(defaultUserId = "user_1") {
  return String(wx.getStorageSync("userId") || defaultUserId);
}

export function setDevAuthToken(userId = "user_1") {
  wx.setStorageSync("authToken", `user:${userId}`);
  wx.setStorageSync("userId", userId);
}

export function isPreviewAuthAllowed() {
  const override = String(wx.getStorageSync("allowPreviewAuth") || "").trim();
  if (override === "true") return true;
  if (override === "false") return false;
  return /^https?:\/\/(127\.0\.0\.1|localhost|\[::1\])(?::|\/|$)/.test(getApiBaseUrl());
}

export function activatePreviewAuth(userId = "user_1") {
  if (!isPreviewAuthAllowed()) return false;
  setDevAuthToken(userId);
  wx.setStorageSync("authPreview", "true");
  return true;
}

export function ensurePreviewAuth(userId = "user_1") {
  if (!getAuthToken()) {
    activatePreviewAuth(userId);
  }
}

export function getWechatLoginCode() {
  return new Promise<string>((resolve, reject) => {
    wx.login({
      success(result) {
        const code = String(result.code || "").trim();
        if (code) {
          resolve(code);
          return;
        }
        reject(new Error("未获取到微信登录凭证"));
      },
      fail(error) {
        reject(new Error(String((error as any)?.errMsg || "微信登录失败")));
      }
    });
  });
}

export async function loginWithWechatMini(code: string, profile: Record<string, unknown> = {}) {
  const data = await request("/api/auth/wechat-mini/login", {
    method: "POST",
    auth: false,
    data: {
      code,
      nickname: profile.nickname || "",
      avatar_url: profile.avatar_url || ""
    }
  }) as { access_token?: string; user?: Record<string, unknown> };
  return persistAuthResult(data);
}

function persistAuthResult(data: { access_token?: string; user?: Record<string, unknown> }) {
  if (data?.access_token) {
    wx.setStorageSync("authToken", data.access_token);
    wx.setStorageSync("authPreview", "false");
  }
  if (data?.user) {
    wx.setStorageSync("authMode", "user");
    if (data.user.id) wx.setStorageSync("userId", String(data.user.id));
    if (data.user.phone) wx.setStorageSync("userPhone", String(data.user.phone));
    if (data.user.nickname) wx.setStorageSync("userNickname", String(data.user.nickname));
  }
  return data;
}

export async function sendPhoneVerificationCode(phone: string, purpose = "login") {
  return request("/api/auth/phone/code", {
    method: "POST",
    auth: false,
    data: { phone, purpose }
  }) as Promise<{ dev_code?: string; masked_phone?: string; cooldown_seconds?: number; delivery_status?: string; delivery_provider?: string; delivery_request_id?: string }>;
}

export async function loginWithPhone(payload: Record<string, unknown>) {
  const data = await request("/api/auth/phone/login", {
    method: "POST",
    auth: false,
    data: payload
  }) as { access_token?: string; user?: Record<string, unknown> };
  return persistAuthResult(data);
}

export async function registerWithPhone(payload: Record<string, unknown>) {
  const data = await request("/api/auth/phone/register", {
    method: "POST",
    auth: false,
    data: payload
  }) as { access_token?: string; user?: Record<string, unknown> };
  return persistAuthResult(data);
}

export function request(path: string, options: RequestOptions = {}) {
  const method = options.method || "GET";
  const token = getAuthToken();
  const baseUrl = getApiBaseUrl();
  const header: Record<string, string> = {
    "Content-Type": "application/json"
  };
  if (options.auth !== false && token) {
    header.Authorization = `Bearer ${token}`;
  }
  const requestOnce = (url: string) => new Promise((resolve, reject) => {
    wx.request({
      url,
      method,
      data: options.data || {},
      header,
      success(response) {
        const body = response.data as { success?: boolean; data?: unknown; message?: string };
        if (response.statusCode >= 200 && response.statusCode < 300 && body?.success) {
          resolve(body.data);
          return;
        }
        reject(new Error(body?.message || `request failed: ${response.statusCode} · ${url}`));
      },
      fail(error) {
        reject(new Error(`${String((error as any)?.errMsg || "request failed")} · ${url}`));
      }
    });
  });
  return new Promise((resolve, reject) => {
    requestOnce(`${baseUrl}${path}`)
      .then(resolve)
      .catch((error) => {
        if (baseUrl !== DEFAULT_API_BASE_URL && canFallbackToDefaultApiBase(baseUrl)) {
          requestOnce(`${DEFAULT_API_BASE_URL}${path}`)
            .then((data) => {
              wx.setStorageSync("apiBaseUrl", DEFAULT_API_BASE_URL);
              resolve(data);
            })
            .catch((fallbackError) => {
              reject(new Error(`${String((error as any)?.message || error)}；fallback failed: ${String((fallbackError as any)?.message || fallbackError)}`));
            });
          return;
        }
        reject(error);
      });
  });
}

function canFallbackToDefaultApiBase(baseUrl: string) {
  return /^https?:\/\/(127\.0\.0\.1|localhost|\[::1\]|192\.168\.|10\.|172\.(1[6-9]|2\d|3[0-1])\.)/.test(baseUrl);
}

export function getShops() {
  return request("/api/shops", { auth: false });
}

export function getHomeModules() {
  return request("/api/home/modules", { auth: false });
}

export function getHomeCards() {
  return request("/api/home/cards", { auth: false });
}

export function getShopProducts(shopId: string) {
  return request(`/api/shops/${shopId}/products`, { auth: false });
}

export function getShopDetail(shopId: string) {
  return request(`/api/shops/${encodeURIComponent(shopId)}/detail`, { auth: false });
}

export function getShopGroupbuyDeals(shopId: string) {
  return request(`/api/shops/${shopId}/groupbuy-deals`, { auth: false });
}

export function createGroupbuyOrder(shopId: string, dealId: string, quantity = 1) {
  return request("/api/groupbuy/orders", {
    method: "POST",
    data: { shop_id: shopId, deal_id: dealId, quantity }
  });
}

export function getGroupbuyVouchers() {
  return request("/api/groupbuy/vouchers");
}

export function getUserAddresses() {
  return request("/api/user/addresses");
}

export function saveUserAddress(payload: Record<string, unknown>) {
  return request("/api/user/addresses", {
    method: "POST",
    data: payload
  });
}

export function getUserProfileOverview() {
  return request("/api/user/profile");
}

export function getUserCoupons() {
  return request("/api/user/coupons");
}

export function claimUserCoupon(code: string) {
  return request("/api/user/coupons/claim", {
    method: "POST",
    data: { code }
  });
}

export function getUserPointsSummary() {
  return request("/api/user/points");
}

export function checkInPoints() {
  return request("/api/user/points/check-in", {
    method: "POST",
    data: {}
  });
}

export function getInviteSummary() {
  return request("/api/user/invite-summary");
}

export function searchCatalog(keyword: string, category = "all") {
  return request(`/api/search?keyword=${encodeURIComponent(keyword)}&category=${encodeURIComponent(category)}`, { auth: false });
}

export function getMedicineHome() {
  return request("/api/medicine/home");
}

export function createPrescriptionImageUpload(payload: Record<string, unknown>) {
  return request("/api/prescriptions/upload-ticket", {
    method: "POST",
    data: payload
  });
}

export function confirmPrescriptionImageUpload(payload: Record<string, unknown>) {
  return request("/api/prescriptions/upload-confirm", {
    method: "POST",
    data: payload
  });
}

export function createPrescriptionReview(payload: Record<string, unknown>) {
  return request("/api/prescriptions", {
    method: "POST",
    data: payload
  });
}

export function getPrescriptionReview(reviewId: string) {
  return request(`/api/prescriptions/${encodeURIComponent(reviewId)}`);
}

export function createMedicineOrder(payload: Record<string, unknown>) {
  return request("/api/medicine/orders", {
    method: "POST",
    data: payload
  });
}

export function getMedicineOrderDetail(orderId: string) {
  return request(`/api/medicine/orders/${encodeURIComponent(orderId)}`);
}

export function createErrandOrder(payload: Record<string, unknown>) {
  return request("/api/errand/orders", {
    method: "POST",
    data: payload
  });
}

export function getErrandOrderDetail(orderId: string) {
  return request(`/api/errand/orders/${encodeURIComponent(orderId)}`);
}

export function getCartSummary(shopId: string) {
  return request(`/api/cart?shop_id=${encodeURIComponent(shopId)}`);
}

export function upsertCartItem(shopId: string, productId: string, quantity: number) {
  return request("/api/cart/items", {
    method: "POST",
    data: { shop_id: shopId, product_id: productId, quantity }
  });
}

export function checkoutCart(shopId: string, addressId: string, options: Record<string, unknown>) {
  return request("/api/orders/checkout", {
    method: "POST",
    data: { shop_id: shopId, address_id: addressId, options }
  });
}

export function getOrders() {
  return request("/api/orders");
}

export function createOrder(type: string, amountFen: number) {
  return request("/api/orders", {
    method: "POST",
    data: { type, amount_fen: amountFen }
  });
}

export function requestOrderRefund(orderId: string, payload: Record<string, unknown>) {
  return request(`/api/orders/${encodeURIComponent(orderId)}/refund`, {
    method: "POST",
    data: payload
  });
}

export function createReview(payload: Record<string, unknown>) {
  return request("/api/reviews", {
    method: "POST",
    data: payload
  });
}

export function createReviewImageUpload(payload: Record<string, unknown>) {
  return request("/api/reviews/upload-ticket", {
    method: "POST",
    data: payload
  });
}

export function confirmReviewImageUpload(payload: Record<string, unknown>) {
  return request("/api/reviews/upload-confirm", {
    method: "POST",
    data: payload
  });
}

export function getUserReviews(orderId = "") {
  const query = orderId ? `?order_id=${encodeURIComponent(orderId)}` : "";
  return request(`/api/reviews${query}`);
}

export function getAfterSalesRequests(orderId = "") {
  const query = orderId ? `?order_id=${encodeURIComponent(orderId)}` : "";
  return request(`/api/after-sales${query}`);
}

export function createAfterSales(payload: Record<string, unknown>) {
  return request("/api/after-sales", {
    method: "POST",
    data: payload
  });
}

export function getAfterSalesEvents(requestId: string) {
  return request(`/api/after-sales/${encodeURIComponent(requestId)}/events`);
}

export function addAfterSalesEvent(requestId: string, payload: Record<string, unknown>) {
  return request(`/api/after-sales/${encodeURIComponent(requestId)}/events`, {
    method: "POST",
    data: payload
  });
}

export function getAfterSalesEvidence(requestId: string) {
  return request(`/api/after-sales/${encodeURIComponent(requestId)}/evidence`);
}

export function createAfterSalesEvidenceUpload(requestId: string, payload: Record<string, unknown>) {
  return request(`/api/after-sales/${encodeURIComponent(requestId)}/evidence/upload-ticket`, {
    method: "POST",
    data: payload
  });
}

export function confirmAfterSalesEvidenceUpload(requestId: string, payload: Record<string, unknown>) {
  return request(`/api/after-sales/${encodeURIComponent(requestId)}/evidence/confirm`, {
    method: "POST",
    data: payload
  });
}

export function reportObjectStorageUpload(payload: Record<string, unknown>) {
  return request("/api/object-storage/upload-callback", {
    method: "POST",
    auth: false,
    data: payload
  });
}

export function reportObjectStorageScanResult(payload: Record<string, unknown>) {
  return request("/api/object-storage/scan-result", {
    method: "POST",
    auth: false,
    data: payload
  });
}

export function getUserNotificationPreferences(notificationType = "", limit = 20) {
  const query = [
    notificationType ? `notification_type=${encodeURIComponent(notificationType)}` : "",
    `limit=${encodeURIComponent(String(limit))}`
  ].filter(Boolean).join("&");
  return request(`/api/user/notification-preferences${query ? `?${query}` : ""}`);
}

export function saveUserNotificationPreference(payload: Record<string, unknown>) {
  return request("/api/user/notification-preferences", {
    method: "PUT",
    data: payload
  });
}

export function getFeedbackTickets() {
  return request("/api/feedback");
}

export function createFeedback(payload: Record<string, unknown>) {
  return request("/api/feedback", {
    method: "POST",
    data: payload
  });
}

export function getServiceTickets() {
  return request("/api/service-tickets");
}

export function createServiceTicket(payload: Record<string, unknown>) {
  return request("/api/service-tickets", {
    method: "POST",
    data: payload
  });
}

export function getServiceTicketDetail(ticketId: string) {
  return request(`/api/service-tickets/${encodeURIComponent(ticketId)}`);
}

export function addServiceTicketEvent(ticketId: string, payload: Record<string, unknown>) {
  return request(`/api/service-tickets/${encodeURIComponent(ticketId)}/events`, {
    method: "POST",
    data: payload
  });
}

export function closeServiceTicket(ticketId: string, payload: Record<string, unknown> = {}) {
  return request(`/api/service-tickets/${encodeURIComponent(ticketId)}/close`, {
    method: "POST",
    data: payload
  });
}

export function followUpServiceTicket(ticketId: string, payload: Record<string, unknown>) {
  return request(`/api/service-tickets/${encodeURIComponent(ticketId)}/follow-up`, {
    method: "POST",
    data: payload
  });
}

export function getCirclePosts() {
  return request("/api/circle/posts");
}

export function createCirclePost(payload: Record<string, unknown>) {
  return request("/api/circle/posts", {
    method: "POST",
    data: payload
  });
}

export function getMealMatchProfile() {
  return request("/api/meal-match/profile");
}

export function saveMealMatchProfile(payload: Record<string, unknown>) {
  return request("/api/meal-match/profile", {
    method: "PUT",
    data: payload
  });
}

export function getMealMatchCandidates() {
  return request("/api/meal-match/candidates");
}

export function reportMealMatchCandidate(payload: Record<string, unknown>) {
  return request("/api/meal-match/reports", {
    method: "POST",
    data: payload
  });
}

export function blockMealMatchCandidate(payload: Record<string, unknown>) {
  return request("/api/meal-match/blocks", {
    method: "POST",
    data: payload
  });
}

export function createRedPacket(payload: Record<string, unknown>) {
  return request("/api/red-packets", {
    method: "POST",
    data: payload
  });
}

export function getRedPacketDetail(packetId: string) {
  return request(`/api/red-packets/${encodeURIComponent(packetId)}`);
}

export function claimRedPacket(packetId: string) {
  return request(`/api/red-packets/${encodeURIComponent(packetId)}/claim`, {
    method: "POST",
    data: {}
  });
}

export function refundRedPacket(packetId: string) {
  return request(`/api/red-packets/${encodeURIComponent(packetId)}/refund`, {
    method: "POST",
    data: {}
  });
}

export function getMessageThreads() {
  return request("/api/messages/threads");
}

export function getChatThreadOverview(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/overview`);
}

export function getChatThreadMembers(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/members`);
}

export function getChatThreadMembership(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/membership`);
}

export function joinChatThread(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/join`, {
    method: "POST",
    data: {}
  });
}

export function leaveChatThread(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/leave`, {
    method: "POST",
    data: {}
  });
}

export function getChatMessages(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}`);
}

export function syncChatMessages(threadId: string, sinceId = "", markRead = true) {
  const query = [
    sinceId ? `since_id=${encodeURIComponent(sinceId)}` : "",
    `mark_read=${markRead ? "true" : "false"}`
  ].filter(Boolean).join("&");
  return request(`/api/messages/${encodeURIComponent(threadId)}/sync${query ? `?${query}` : ""}`);
}

export function getChatThreadPreference(threadId: string) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/preference`);
}

export function saveChatThreadPreference(threadId: string, payload: Record<string, unknown>) {
  return request(`/api/messages/${encodeURIComponent(threadId)}/preference`, {
    method: "PUT",
    data: payload
  });
}

export function markChatThreadRead(threadId: string, lastMessageId = "") {
  return request(`/api/messages/${encodeURIComponent(threadId)}/read`, {
    method: "POST",
    data: lastMessageId ? { last_message_id: lastMessageId } : {}
  });
}

export function sendChatMessage(threadId: string, payload: Record<string, unknown>) {
  return request(`/api/messages/${encodeURIComponent(threadId)}`, {
    method: "POST",
    data: payload
  });
}

export function getOrderDetail(orderId: string) {
  return request(`/api/orders/${encodeURIComponent(orderId)}`);
}

export function setPaymentPassword(password: string) {
  return request("/api/wallet/payment-password", {
    method: "POST",
    data: { password }
  });
}

export function creditWallet(amountFen: number) {
  return request("/api/wallet/credit", {
    method: "POST",
    data: {
      amount_fen: amountFen,
      idempotency_key: `wechat_wallet_credit_${Date.now()}`
    }
  });
}

export function getWalletTransactions() {
  return request("/api/wallet/transactions");
}

export function getWalletOverview() {
  return request("/api/wallet/overview");
}

export function requestWalletWithdraw(amountFen: number) {
  return request("/api/wallet/withdraw", {
    method: "POST",
    data: {
      amount_fen: amountFen,
      channel: "wechat_change"
    }
  });
}

export function payOrderWithBalance(orderId: string, paymentPassword: string) {
  return request("/api/wallet/pay", {
    method: "POST",
    data: {
      order_id: orderId,
      payment_password: paymentPassword,
      idempotency_key: `wechat_balance_pay_${orderId}_${Date.now()}`
    }
  });
}

export function createWechatPrepay(orderId: string) {
  return request("/api/payments/wechat/prepay", {
    method: "POST",
    data: { order_id: orderId }
  });
}
