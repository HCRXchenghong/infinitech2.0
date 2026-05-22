const DEFAULT_API_BASE_URL = "http://127.0.0.1:25500";

type RequestOptions = {
  method?: "GET" | "POST";
  data?: Record<string, unknown>;
  auth?: boolean;
};

export function getApiBaseUrl() {
  return String(wx.getStorageSync("apiBaseUrl") || DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export function getAuthToken() {
  return String(wx.getStorageSync("authToken") || "");
}

export function setDevAuthToken(userId = "user_1") {
  wx.setStorageSync("authToken", `user:${userId}`);
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
  }) as { access_token?: string };
  if (data?.access_token) {
    wx.setStorageSync("authToken", data.access_token);
  }
  return data;
}

export function request(path: string, options: RequestOptions = {}) {
  const method = options.method || "GET";
  const token = getAuthToken();
  const header: Record<string, string> = {
    "Content-Type": "application/json"
  };
  if (options.auth !== false && token) {
    header.Authorization = `Bearer ${token}`;
  }
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${getApiBaseUrl()}${path}`,
      method,
      data: options.data || {},
      header,
      success(response) {
        const body = response.data as { success?: boolean; data?: unknown; message?: string };
        if (response.statusCode >= 200 && response.statusCode < 300 && body?.success) {
          resolve(body.data);
          return;
        }
        reject(new Error(body?.message || `request failed: ${response.statusCode}`));
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

export function getShops() {
  return request("/api/shops", { auth: false });
}

export function getShopProducts(shopId: string) {
  return request(`/api/shops/${shopId}/products`, { auth: false });
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

export function getOrderDetail(orderId: string) {
  return request(`/api/orders/${encodeURIComponent(orderId)}`);
}

export function setPaymentPassword(password: string) {
  return request("/api/wallet/payment-password", {
    method: "POST",
    data: { password }
  });
}

export function createWechatPrepay(orderId: string) {
  return request("/api/payments/wechat/prepay", {
    method: "POST",
    data: { order_id: orderId }
  });
}
