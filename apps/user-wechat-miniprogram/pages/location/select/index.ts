import { ensurePreviewAuth, getUserAddresses } from "../../../utils/api";
import { HOME_LOCATION_KEY, readCachedHomeLocation } from "../../../utils/location";

type LocationSource = "gps" | "map" | "address";

type LocationRow = {
  id: string;
  name: string;
  address: string;
  distanceText: string;
  latitude: number;
  longitude: number;
  source: LocationSource;
  contact?: string;
  tag?: string;
};

const EMPTY_LOCATION: LocationRow = {
  id: "current",
  name: "正在获取真实定位",
  address: "授权定位后可使用当前位置，或打开微信地图精确选点",
  distanceText: "当前",
  latitude: 0,
  longitude: 0,
  source: "gps"
};

Page({
  data: {
    keyword: "",
    locating: false,
    loadingAddresses: false,
    selectedId: "current",
    selectedLocation: EMPTY_LOCATION,
    currentLocation: EMPTY_LOCATION,
    savedAddresses: [] as LocationRow[],
    visiblePlaces: [] as LocationRow[],
    frequentPlaces: [] as LocationRow[],
    emptyText: "暂无后端保存地址，请使用微信地图选择真实地点"
  },

  onLoad() {
    this.loadCachedLocation();
    this.handleRelocate();
    this.loadSavedAddresses();
  },

  onShow() {
    this.loadSavedAddresses();
  },

  loadCachedLocation() {
    const cached = readCachedHomeLocation();
    if (!cached) {
      this.setData({ selectedLocation: this.data.currentLocation });
      return;
    }

    const currentLocation: LocationRow = {
      id: "current",
      name: cached.name || "当前位置",
      address: cached.address || "已读取上次选择的真实位置",
      distanceText: "当前",
      latitude: Number(cached.latitude || 0),
      longitude: Number(cached.longitude || 0),
      source: String(cached.source || "") === "backend_address" ? "address" : "gps"
    };
    this.setData({
      currentLocation,
      selectedId: "current",
      selectedLocation: currentLocation
    });
  },

  async loadSavedAddresses() {
    ensurePreviewAuth();
    this.setData({ loadingAddresses: true });
    try {
      const addresses = await getUserAddresses() as Array<Record<string, any>>;
      const rows = (Array.isArray(addresses) ? addresses : [])
        .map(addressToLocationRow)
        .filter(Boolean) as LocationRow[];
      this.setData({
        loadingAddresses: false,
        savedAddresses: rows,
        visiblePlaces: filterLocations(String(this.data.keyword || ""), rows),
        frequentPlaces: rows.slice(0, 2),
        emptyText: rows.length ? "" : "暂无后端保存地址，请使用微信地图选择真实地点"
      });
    } catch (_error) {
      this.setData({
        loadingAddresses: false,
        savedAddresses: [],
        visiblePlaces: [],
        frequentPlaces: [],
        emptyText: "后端地址暂未同步成功，请稍后重试或使用微信地图选点"
      });
    }
  },

  handleKeywordInput(event) {
    const keyword = String(event.detail.value || "").trim();
    this.setData({
      keyword,
      visiblePlaces: filterLocations(keyword, this.data.savedAddresses)
    });
  },

  handleClearKeyword() {
    this.setData({
      keyword: "",
      visiblePlaces: this.data.savedAddresses
    });
  },

  handleBack() {
    wx.navigateBack({
      fail() {
        wx.switchTab({ url: "/pages/index/index" });
      }
    });
  },

  handleSearchTap() {
    this.openWechatLocationPicker();
  },

  handleManualPick() {
    this.openWechatLocationPicker();
  },

  handleRelocate() {
    this.setData({ locating: true });
    wx.getLocation({
      type: "gcj02",
      success: (res) => {
        const currentLocation: LocationRow = {
          id: "current",
          name: "当前位置",
          address: "已获取真实坐标，建议用微信地图补全具体门牌",
          distanceText: "当前",
          latitude: res.latitude,
          longitude: res.longitude,
          source: "gps"
        };
        this.setData({
          locating: false,
          currentLocation,
          selectedId: "current",
          selectedLocation: currentLocation
        });
      },
      fail: () => {
        this.setData({ locating: false });
        wx.showToast({ title: "请授权定位或手动选择", icon: "none" });
      }
    });
  },

  openWechatLocationPicker() {
    wx.chooseLocation({
      success: (res) => {
        const picked: LocationRow = {
          id: "picked",
          name: String(res.name || res.address || "已选真实位置"),
          address: String(res.address || ""),
          distanceText: "地图",
          latitude: Number(res.latitude || 0),
          longitude: Number(res.longitude || 0),
          source: "map"
        };
        this.setData({
          selectedId: picked.id,
          selectedLocation: picked,
          currentLocation: picked
        });
      },
      fail: () => {
        wx.showToast({ title: "未选择位置", icon: "none" });
      }
    });
  },

  handlePlaceTap(event) {
    const id = String(event.currentTarget.dataset.id || "");
    const place = this.data.savedAddresses.find((item) => item.id === id);
    if (!place) return;

    this.setData({
      selectedId: id,
      selectedLocation: place
    });
  },

  handleCurrentTap() {
    this.setData({
      selectedId: "current",
      selectedLocation: this.data.currentLocation
    });
  },

  handleFrequentTap(event) {
    this.handlePlaceTap(event);
  },

  handleAddAddress() {
    wx.navigateTo({ url: "/pages/address/list/index" });
  },

  handleConfirm() {
    const selected = this.data.selectedLocation || this.data.currentLocation;
    if (!selected || !selected.name) {
      wx.showToast({ title: "请先选择位置", icon: "none" });
      return;
    }
    if (!selected.latitude || !selected.longitude) {
      wx.showToast({ title: "请选择带坐标的位置", icon: "none" });
      return;
    }

    wx.setStorageSync(HOME_LOCATION_KEY, {
      name: selected.name,
      address: selected.address,
      latitude: selected.latitude,
      longitude: selected.longitude,
      source: selected.source === "address" ? "backend_address" : selected.source,
      updatedAt: Date.now()
    });
    wx.showToast({ title: "位置已确认", icon: "success" });
    setTimeout(() => {
      wx.navigateBack({
        fail() {
          wx.switchTab({ url: "/pages/index/index" });
        }
      });
    }, 220);
  }
});

function addressToLocationRow(address: Record<string, any>): LocationRow | null {
  const id = String(address.id || "").trim();
  const city = String(address.city || "").trim();
  const detail = String(address.detail || "").trim();
  const latitude = Number(address.latitude || 0);
  const longitude = Number(address.longitude || 0);
  if (!id || !city || !detail || !Number.isFinite(latitude) || !Number.isFinite(longitude) || !latitude || !longitude) {
    return null;
  }
  const contact = `${String(address.contact_name || "").trim()} ${String(address.contact_phone || "").trim()}`.trim();
  const tag = tagText(String(address.tag || ""));
  return {
    id,
    name: `${city} ${detail}`.trim(),
    address: contact ? `${tag} · ${contact}` : tag,
    distanceText: "已保存",
    latitude,
    longitude,
    source: "address",
    contact,
    tag
  };
}

function filterLocations(keyword: string, rows: LocationRow[]) {
  const text = String(keyword || "").trim();
  if (!text) return rows;
  return rows.filter((item) => `${item.name}${item.address}${item.contact || ""}${item.tag || ""}`.includes(text));
}

function tagText(tag: string) {
  const map: Record<string, string> = {
    home: "家",
    company: "公司",
    school: "学校",
    other: "地址"
  };
  return map[tag] || tag || "地址";
}
