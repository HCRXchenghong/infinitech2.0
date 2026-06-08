const TAB_ROUTES = [
  "/pages/index/index",
  "/pages/order/list/index",
  "/pages/messages/index",
  "/pages/profile/index"
];

export function normalizeRoute(url: string) {
  const text = String(url || "");
  return text.split("?")[0];
}

export function isTabRoute(url: string) {
  return TAB_ROUTES.includes(normalizeRoute(url));
}

export function switchTabRoute(url = "/pages/index/index") {
  const target = normalizeRoute(url);
  wx.switchTab({
    url: target,
    fail() {
      wx.reLaunch({ url: target });
    }
  });
}

export function openRoute(url = "/pages/index/index") {
  const target = String(url || "/pages/index/index");
  const path = normalizeRoute(target);

  if (isTabRoute(path)) {
    rememberTabIntent(target);
    switchTabRoute(path);
    return;
  }

  wx.navigateTo({ url: target });
}

export function replaceRoute(url = "/pages/index/index") {
  const target = String(url || "/pages/index/index");
  const path = normalizeRoute(target);

  if (isTabRoute(path)) {
    rememberTabIntent(target);
    switchTabRoute(path);
    return;
  }

  wx.redirectTo({ url: target });
}

export function backOrSwitchTab(url = "/pages/index/index") {
  wx.navigateBack({
    fail() {
      switchTabRoute(url);
    }
  });
}

function rememberTabIntent(url: string) {
  const [path, query = ""] = String(url || "").split("?");
  if (!query) return;

  wx.setStorageSync(`tabIntent:${path}`, query);

  if (path === "/pages/order/list/index") {
    const status = getQueryValue(query, "status");
    if (status) {
      wx.setStorageSync("orderListStatus", status);
    }
  }
}

function getQueryValue(query: string, key: string) {
  const target = `${key}=`;
  const pair = query.split("&").find((item) => item.indexOf(target) === 0);
  if (!pair) return "";

  try {
    return decodeURIComponent(pair.slice(target.length));
  } catch (_error) {
    return pair.slice(target.length);
  }
}
