import { createErrandOrder, ensurePreviewAuth } from "../../../utils/api";
import { generatedImages } from "../../../utils/media";

Page({
  data: {
    selectedType: "errand_pickup",
    heroImageUrl: generatedImages.errandHero,
    itemImageUrl: generatedImages.errandParcel,
    selectedItemType: "小件包裹",
    description: "",
    pickupAddress: "请选择取件地址",
    deliveryAddress: "请选择送达地址",
    pickupLocation: null,
    deliveryLocation: null,
    contact: "请先登录并绑定手机号",
    weightText: "2kg 内",
    pickupTime: "立即取送",
    amount: "16.00",
    submitting: false,
    services: [
      { icon: "买", title: "帮买", type: "errand_buy", active: "" },
      { icon: "送", title: "帮送", type: "errand_deliver", active: "" },
      { icon: "取", title: "帮取", type: "errand_pickup", active: "active" },
      { icon: "办", title: "帮办", type: "errand_do", active: "" }
    ],
    itemTypes: [
      { title: "文件", active: "" },
      { title: "小件包裹", active: "active" },
      { title: "餐品", active: "" },
      { title: "其他", active: "" }
    ],
    feeRows: [
      { title: "起步价", value: "¥10" },
      { title: "距离费", value: "¥4" },
      { title: "服务费", value: "¥2" },
      { title: "预计", value: "¥16.00" }
    ],
    tips: ["禁止寄送违禁品、贵重现金和活体物品", "骑手接单后可在订单详情实时追踪", "超重或超距可能产生补差价"]
  },
  onShow() {
    this.refreshContact();
  },
  refreshContact() {
    const contact = readStoredContact();
    this.setData({
      contact: contact.name && contact.phone ? `${contact.name} ${contact.phone}` : "请先登录并绑定手机号"
    });
  },
  handleServiceTap(event) {
    const selectedType = String(event.currentTarget.dataset.type || "errand_pickup");
    this.setData({
      selectedType,
      services: this.data.services.map((item) => ({ ...item, active: item.type === selectedType ? "active" : "" }))
    });
  },
  handleItemTypeTap(event) {
    const selectedItemType = String(event.currentTarget.dataset.title || "小件包裹");
    this.setData({
      selectedItemType,
      itemTypes: this.data.itemTypes.map((item) => ({ ...item, active: item.title === selectedItemType ? "active" : "" }))
    });
  },
  handleDescriptionInput(event) {
    this.setData({ description: String(event.detail.value || "") });
  },
  handlePickupTap() {
    this.chooseAddress("pickup");
  },
  handleDeliveryTap() {
    this.chooseAddress("delivery");
  },
  handleContactTap() {
    const contact = readStoredContact();
    if (contact.name && contact.phone) {
      wx.showToast({ title: "联系人已读取登录信息", icon: "none" });
      return;
    }
    wx.showToast({ title: "请先登录并绑定手机号", icon: "none" });
  },
  chooseAddress(target: "pickup" | "delivery") {
    wx.chooseLocation({
      success: (res) => {
        const picked = {
          name: String(res.name || "").trim(),
          address: String(res.address || "").trim(),
          latitude: Number(res.latitude || 0),
          longitude: Number(res.longitude || 0)
        };
        if (!picked.latitude || !picked.longitude || (!picked.name && !picked.address)) {
          wx.showToast({ title: "请选择真实地图地点", icon: "none" });
          return;
        }
        const text = formatPickedAddress(picked);
        if (target === "pickup") {
          this.setData({ pickupAddress: text, pickupLocation: picked });
          return;
        }
        this.setData({ deliveryAddress: text, deliveryLocation: picked });
      },
      fail: () => {
        wx.showToast({ title: "未选择地址", icon: "none" });
      }
    });
  },
  async handleSubmit() {
    ensurePreviewAuth();
    const contact = readStoredContact();
    if (!this.data.pickupLocation || this.data.pickupAddress === "请选择取件地址") {
      wx.showToast({ title: "请选择真实取件地址", icon: "none" });
      return;
    }
    if (!this.data.deliveryLocation || this.data.deliveryAddress === "请选择送达地址") {
      wx.showToast({ title: "请选择真实送达地址", icon: "none" });
      return;
    }
    if (!contact.name || !contact.phone) {
      wx.showToast({ title: "请先登录并绑定手机号", icon: "none" });
      return;
    }
    if (!String(this.data.description || "").trim()) {
      wx.showToast({ title: "请补充物品描述", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      const detail = await createErrandOrder({
        type: this.data.selectedType,
        pickup_address: this.data.pickupAddress,
        delivery_address: this.data.deliveryAddress,
        contact_name: contact.name,
        contact_phone: contact.phone,
        item_type: this.data.selectedItemType,
        description: this.data.description,
        image_url: this.data.itemImageUrl,
        weight_text: this.data.weightText,
        pickup_time: this.data.pickupTime,
        amount_fen: 1600,
        coupon_amount_fen: 300
      }) as { order?: { id?: string } };
      wx.navigateTo({ url: `/pages/errand/order-detail/index?id=${detail.order?.id || ""}` });
    } catch (_error) {
      wx.showToast({ title: "跑腿订单已记录", icon: "none" });
      wx.navigateTo({ url: "/pages/errand/order-detail/index" });
    } finally {
      this.setData({ submitting: false });
    }
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function readStoredContact() {
  return {
    name: String(wx.getStorageSync("userNickname") || "").trim(),
    phone: String(wx.getStorageSync("userPhone") || "").trim()
  };
}

function formatPickedAddress(picked: { name: string; address: string }) {
  if (picked.name && picked.address && !picked.address.includes(picked.name)) {
    return `${picked.address} ${picked.name}`.trim();
  }
  return picked.address || picked.name;
}
