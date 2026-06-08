export const HOME_LOCATION_KEY = "homeLocation";

export function readCachedHomeLocation() {
  const cached = wx.getStorageSync(HOME_LOCATION_KEY);
  if (!cached || !cached.name) return null;

  return {
    name: String(cached.name || ""),
    address: String(cached.address || ""),
    latitude: Number(cached.latitude || 0),
    longitude: Number(cached.longitude || 0),
    source: String(cached.source || "")
  };
}

export function cachedHomeAddress(fallback = "请补充详细地址") {
  const location = readCachedHomeLocation();
  if (!location) return fallback;

  return location.address || location.name || fallback;
}
