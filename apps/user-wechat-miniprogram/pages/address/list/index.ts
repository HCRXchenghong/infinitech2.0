import { ensurePreviewAuth, getUserAddresses, saveUserAddress } from "../../../utils/api";

type AddressRow = {
  id: string;
  name: string;
  phone: string;
  tag: string;
  tagValue: string;
  city: string;
  detail: string;
  rawDetail: string;
  latitude: number;
  longitude: number;
  defaulted: boolean;
};

Page({
  data: {
    mode: "manage",
    statusText: "地址已接",
    emptyText: "",
    selectedAddressId: "",
    addresses: [] as AddressRow[],
    saving: false
  },
  onLoad(query) {
    this.setData({
      mode: String(query?.mode || "manage"),
      selectedAddressId: String(wx.getStorageSync("checkoutAddressId") || "")
    });
    this.loadAddresses();
  },
  onShow() {
    this.setData({ selectedAddressId: String(wx.getStorageSync("checkoutAddressId") || "") });
    this.loadAddresses();
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/profile/index" }) });
  },
  async loadAddresses() {
    ensurePreviewAuth();
    try {
      const addresses = await getUserAddresses();
      const rows = (Array.isArray(addresses) ? addresses : [])
        .map(addressFromApi)
        .filter(Boolean) as AddressRow[];
      this.setData({
        addresses: rows,
        statusText: rows.length ? "已同步" : "暂无地址",
        emptyText: rows.length ? "" : "暂无后端保存地址，请使用微信地图新增真实收货地址"
      });
    } catch (_error) {
      this.setData({
        addresses: [],
        statusText: "同步失败",
        emptyText: "后端地址暂未同步成功，请稍后重试"
      });
    }
  },
  async handleAddTap() {
    const contact = readStoredContact();
    if (!contact.name || !contact.phone) {
      wx.showToast({ title: "请先登录并绑定手机号", icon: "none" });
      return;
    }

    this.setData({ saving: true });
    try {
      const picked = await chooseWechatLocation();
      const city = inferCity(picked.address);
      if (!city) {
        wx.showToast({ title: "未识别城市，请重新选点", icon: "none" });
        return;
      }
      const detail = normalizeDetail(picked.address, city, picked.name);
      if (!detail || !picked.latitude || !picked.longitude) {
        wx.showToast({ title: "请选择带门牌和坐标的位置", icon: "none" });
        return;
      }

      ensurePreviewAuth();
      await saveUserAddress({
        contact_name: contact.name,
        contact_phone: contact.phone,
        city,
        detail,
        latitude: picked.latitude,
        longitude: picked.longitude,
        tag: "other",
        is_default: this.data.addresses.length === 0
      });
      wx.showToast({ title: "地址已保存", icon: "success" });
      this.loadAddresses();
    } catch (error) {
      const message = String((error as any)?.message || (error as any)?.errMsg || "");
      if (!message.includes("cancel")) {
        wx.showToast({ title: "地址保存失败", icon: "none" });
      }
    } finally {
      this.setData({ saving: false });
    }
  },
  async handleSetDefault(event) {
    const id = String(event.currentTarget.dataset.id || "");
    const current = this.data.addresses.find((address) => address.id === id);
    if (!current) {
      return;
    }
    ensurePreviewAuth();
    this.setData({ saving: true });
    try {
      await saveUserAddress({
        id: current.id,
        contact_name: current.name,
        contact_phone: current.phone,
        city: current.city,
        detail: current.rawDetail,
        latitude: current.latitude,
        longitude: current.longitude,
        tag: current.tagValue,
        is_default: true
      });
      wx.showToast({ title: "默认地址已更新", icon: "success" });
      this.loadAddresses();
    } catch (_error) {
      wx.showToast({ title: "默认地址更新失败", icon: "none" });
    } finally {
      this.setData({ saving: false });
    }
  },
  handleEdit() {
    wx.showToast({ title: "地址编辑表单待接后端字段", icon: "none" });
  },
  handleDelete() {
    wx.showToast({ title: "删除地址接口待接入", icon: "none" });
  },
  handleAddressCardTap(event) {
    if (this.data.mode !== "select") {
      return;
    }
    const id = String(event.currentTarget.dataset.id || "");
    const current = this.data.addresses.find((address) => address.id === id);
    if (!id || !current) {
      return;
    }
    wx.setStorageSync("checkoutAddressId", id);
    wx.setStorageSync("checkoutAddressSnapshot", current);
    this.setData({ selectedAddressId: id });
    wx.showToast({ title: "已选择配送地址", icon: "success" });
    setTimeout(() => {
      wx.navigateBack();
    }, 260);
  }
});

function addressFromApi(address: Record<string, any>): AddressRow | null {
  const id = String(address.id || "").trim();
  const city = String(address.city || "").trim();
  const rawDetail = String(address.detail || "").trim();
  const latitude = Number(address.latitude || 0);
  const longitude = Number(address.longitude || 0);
  if (!id || !city || !rawDetail || !Number.isFinite(latitude) || !Number.isFinite(longitude) || !latitude || !longitude) {
    return null;
  }
  return {
    id,
    name: String(address.contact_name || "").trim(),
    phone: String(address.contact_phone || "").trim(),
    tag: tagText(String(address.tag || "")),
    tagValue: String(address.tag || "other"),
    city,
    detail: `${city} ${rawDetail}`.trim(),
    rawDetail,
    latitude,
    longitude,
    defaulted: Boolean(address.is_default)
  };
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

function readStoredContact() {
  return {
    name: String(wx.getStorageSync("userNickname") || "").trim(),
    phone: String(wx.getStorageSync("userPhone") || "").trim()
  };
}

function chooseWechatLocation() {
  return new Promise<{ name: string; address: string; latitude: number; longitude: number }>((resolve, reject) => {
    wx.chooseLocation({
      success(res) {
        resolve({
          name: String(res.name || "").trim(),
          address: String(res.address || "").trim(),
          latitude: Number(res.latitude || 0),
          longitude: Number(res.longitude || 0)
        });
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

function inferCity(address: string) {
  const text = String(address || "").trim();
  const direct = text.match(/([\u4e00-\u9fa5]{2,}(?:市|地区|自治州|盟))/);
  if (direct) return direct[1];
  const county = text.match(/([\u4e00-\u9fa5]{2,}(?:县|区))/);
  return county ? county[1] : "";
}

function normalizeDetail(address: string, city: string, name: string) {
  const normalizedAddress = String(address || "").trim();
  const normalizedCity = String(city || "").trim();
  const withoutCity = normalizedCity && normalizedAddress.startsWith(normalizedCity)
    ? normalizedAddress.slice(normalizedCity.length).trim()
    : normalizedAddress;
  const placeName = String(name || "").trim();
  if (withoutCity && placeName && !withoutCity.includes(placeName)) {
    return `${withoutCity} ${placeName}`.trim();
  }
  return withoutCity || placeName || normalizedAddress;
}
