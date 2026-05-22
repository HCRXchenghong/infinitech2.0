const DEFAULT_API_BASE_URL = "http://127.0.0.1:25500";

export function getApiBaseUrl() {
  return String(uni.getStorageSync("apiBaseUrl") || DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export function getRiderToken() {
  return String(uni.getStorageSync("riderToken") || "rider:rider_1");
}

export function getStationManagerToken() {
  return String(uni.getStorageSync("stationManagerToken") || "station_manager:station_manager_1");
}

export function request(path, options = {}) {
  const method = options.method || "GET";
  const token = Object.prototype.hasOwnProperty.call(options, "token") ? options.token : getRiderToken();
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

function persistRiderAuth(data) {
  if (!data || !data.access_token || !data.rider) {
    return data;
  }
  if (data.rider.type === "station_manager") {
    uni.setStorageSync("stationManagerToken", data.access_token);
  } else {
    uni.setStorageSync("riderToken", data.access_token);
  }
  return data;
}

export function setRiderOnline(payload) {
  return request("/api/rider/online", {
    method: "POST",
    data: payload
  });
}

export function getRiderDeposit() {
  return request("/api/rider/deposit");
}

export function payRiderDeposit() {
  return request("/api/rider/deposit/pay", {
    method: "POST",
    data: { amount_fen: 5000 }
  });
}

export function applyWechatDepositExemption() {
  return request("/api/rider/deposit/wechat-exempt", {
    method: "POST",
    data: { application_id: `wx_exempt_${Date.now()}` }
  });
}

export function requestRiderDepositRefund() {
  return request("/api/rider/deposit/refund-request", {
    method: "POST",
    data: {}
  });
}

export function acceptRiderInvite(token, password) {
  return request("/api/auth/rider/invite-register", {
    method: "POST",
    token: "",
    data: { token, password }
  }).then(persistRiderAuth);
}

export function loginRider(accountId, password) {
  return request("/api/auth/rider/login", {
    method: "POST",
    token: "",
    data: {
      account_id: accountId,
      password
    }
  }).then(persistRiderAuth);
}

export function createStationRiderInvite(stationId) {
  return request("/api/station-manager/rider-invites", {
    method: "POST",
    token: getStationManagerToken(),
    data: {
      type: "rider",
      station_id: stationId
    }
  });
}

export function rejectAssignedOrder(orderId) {
  return request(`/api/rider/orders/${encodeURIComponent(orderId)}/reject-assignment`, {
    method: "POST"
  });
}

export function markOrderPickedUp(orderId) {
  return request(`/api/rider/orders/${encodeURIComponent(orderId)}/pickup`, {
    method: "POST"
  });
}

export function markOrderDelivered(orderId) {
  return request(`/api/rider/orders/${encodeURIComponent(orderId)}/delivered`, {
    method: "POST"
  });
}

export function consumeFreeCancel() {
  return request("/api/rider/dispatch/cancel-free", {
    method: "POST"
  });
}

export function getStationRiders() {
  return request("/api/station-manager/riders", {
    token: getStationManagerToken()
  });
}

export function getStationOrders() {
  return request("/api/station-manager/orders", {
    token: getStationManagerToken()
  });
}

export function manualAssignOrder(orderId, riderId) {
  return request(`/api/station-manager/dispatch/${encodeURIComponent(orderId)}/manual-assign`, {
    method: "POST",
    token: getStationManagerToken(),
    data: { rider_id: riderId }
  });
}

export function getStationTaskConfig() {
  return request("/api/station-manager/task-duration", {
    token: getStationManagerToken()
  });
}

export function saveStationTaskConfig(payload) {
  return request("/api/station-manager/task-duration", {
    method: "PUT",
    token: getStationManagerToken(),
    data: payload
  });
}

export function getStationRiderPerformance() {
  return request("/api/station-manager/rider-performance", {
    token: getStationManagerToken()
  });
}
