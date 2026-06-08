const DEFAULT_API_BASE_URL = "http://127.0.0.1:25500";

export function getApiBaseUrl() {
  return String(uni.getStorageSync("apiBaseUrl") || DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export function getMerchantToken() {
  return String(uni.getStorageSync("merchantToken") || "merchant:merchant_1");
}

export function request(path, options = {}) {
  const method = options.method || "GET";
  const token = Object.prototype.hasOwnProperty.call(options, "token") ? options.token : getMerchantToken();
  const header = {
    "Content-Type": "application/json"
  };
  if (token) {
    header.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
  }
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${getApiBaseUrl()}${path}`,
      method,
      data: options.data || {},
      header,
      success(response) {
        const body = response.data || {};
        if (response.statusCode >= 200 && response.statusCode < 300 && body.success) {
          resolve(body.data);
          return;
        }
        reject(new Error(body.message || `request failed: ${response.statusCode}`));
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

function persistMerchantAuth(data) {
  if (data && data.access_token) {
    uni.setStorageSync("merchantToken", data.access_token);
  }
  return data;
}

export function acceptMerchantInvite(token, displayName, accountType, password) {
  return request("/api/auth/merchant/invite-register", {
    method: "POST",
    token: "",
    data: {
      token,
      display_name: displayName,
      account_type: accountType,
      password
    }
  }).then(persistMerchantAuth);
}

export function loginMerchant(accountId, password) {
  return request("/api/auth/merchant/login", {
    method: "POST",
    token: "",
    data: {
      account_id: accountId,
      password
    }
  }).then(persistMerchantAuth);
}

export function getMerchantProfile() {
  return request("/api/merchant/me");
}

export function getMerchantNotificationPreferences(notificationType = "", limit = 20) {
  const params = [];
  if (notificationType) params.push(`notification_type=${encodeURIComponent(notificationType)}`);
  if (limit) params.push(`limit=${encodeURIComponent(String(limit))}`);
  const query = params.join("&");
  return request(`/api/merchant/notification-preferences${query ? `?${query}` : ""}`);
}

export function saveMerchantNotificationPreference(payload) {
  return request("/api/merchant/notification-preferences", {
    method: "PUT",
    data: payload
  });
}

export function saveMerchantQualification(payload) {
  return request("/api/merchant/qualifications", {
    method: "POST",
    data: payload
  });
}

export function getMerchantStaff() {
  return request("/api/merchant/staff");
}

export function saveMerchantStaff(payload) {
  return request("/api/merchant/staff", {
    method: "POST",
    data: payload
  });
}

export function getMerchantMaterials() {
  return request("/api/merchant/materials");
}

export function saveMerchantMaterial(payload) {
  return request("/api/merchant/materials", {
    method: "POST",
    data: payload
  });
}

export function getMerchantOrders() {
  return request("/api/merchant/orders");
}

export function getMerchantDeposit() {
  return request("/api/merchant/deposit");
}

export function payMerchantDeposit() {
  return request("/api/merchant/deposit/pay", {
    method: "POST",
    data: { amount_fen: 5000 }
  });
}

export function acceptMerchantOrder(orderId) {
  return request(`/api/merchant/orders/${encodeURIComponent(orderId)}/accept`, {
    method: "POST"
  });
}

export function markMerchantOrderReady(orderId) {
  return request(`/api/merchant/orders/${encodeURIComponent(orderId)}/ready`, {
    method: "POST"
  });
}

export function getMerchantProducts(shopId = "shop_1") {
  return request(`/api/merchant/products?shop_id=${encodeURIComponent(shopId)}`);
}

export function saveMerchantProduct(product) {
  return request("/api/merchant/products", {
    method: "POST",
    data: product
  });
}

export function setMerchantProductStatus(productId, status) {
  return request(`/api/merchant/products/${encodeURIComponent(productId)}/status`, {
    method: "POST",
    data: { status }
  });
}

export function scanGroupbuyVoucher(payload) {
  return request("/api/merchant/groupbuy/vouchers/scan", {
    method: "POST",
    data: {
      method: "qr_scan",
      qr_payload: payload
    }
  });
}
