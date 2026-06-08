import { createMedicineOrder, ensurePreviewAuth, getUserAddresses } from "../../../utils/api";
import { generatedImages } from "../../../utils/media";

Page({
  data: {
    prescriptionId: "",
    addressId: "",
    addressStatusText: "请选择配送地址",
    payable: "40.60",
    submitting: false,
    address: {
      detail: "请选择真实配送地址",
      contact: "请先选择后端保存地址",
      contactName: "",
      contactPhone: ""
    },
    clinic: {
      name: "校医务室",
      location: "综合楼一层 · 今日 08:30-20:30"
    },
    items: [
      { product_id: "med_cooling_patch", name: "退热贴", imageUrl: generatedImages.medicineCoolingPatch, quantity: 1, tag: "校医务室", price_fen: 1290, price: "12.90", requires_prescription: false },
      { product_id: "med_amoxicillin", name: "阿莫西林胶囊", imageUrl: generatedImages.medicineCapsules, quantity: 1, tag: "处方药", approved: true, price_fen: 1880, price: "18.80", requires_prescription: true },
      { product_id: "med_swab", name: "碘伏棉签", imageUrl: generatedImages.medicineFirstAid, quantity: 1, tag: "外伤消毒", price_fen: 690, price: "6.90", requires_prescription: false }
    ],
    fees: [
      { title: "配送时间", value: "立即配送，约 20-30 分钟" },
      { title: "配送费", value: "¥2.00" },
      { title: "优惠券", value: "暂无可用" },
      { title: "支付方式", value: "余额支付" }
    ]
  },
  onLoad(query) {
    this.setData({ prescriptionId: String(query?.prescription_id || query?.rx_id || "") });
  },
  onShow() {
    this.loadAddress();
  },
  async loadAddress() {
    ensurePreviewAuth();
    try {
      const addresses = await getUserAddresses() as Array<Record<string, any>>;
      if (!Array.isArray(addresses) || addresses.length === 0) {
        this.setData({
          addressId: "",
          addressStatusText: "请先新增一个配送地址",
          address: {
            detail: "暂无后端配送地址",
            contact: "请先新增真实收货地址",
            contactName: "",
            contactPhone: ""
          }
        });
        return;
      }
      const storedId = String(wx.getStorageSync("checkoutAddressId") || "");
      const selected = addresses.find((item) => String(item.id || "") === storedId) || addresses.find((item) => Boolean(item.is_default)) || addresses[0];
      const detail = `${selected.city || ""} ${selected.detail || ""}`.trim();
      wx.setStorageSync("checkoutAddressId", String(selected.id || ""));
      this.setData({
        addressId: String(selected.id || ""),
        addressStatusText: selected.is_default ? "默认地址" : "已选择地址",
        address: {
          detail: detail || "请补充配送地址",
          contact: `${String(selected.contact_name || "")} ${String(selected.contact_phone || "")}`.trim(),
          contactName: String(selected.contact_name || ""),
          contactPhone: String(selected.contact_phone || "")
        }
      });
    } catch (_error) {
      this.setData({
        addressId: "",
        addressStatusText: "地址同步失败",
        address: {
          detail: "请稍后重试或新增真实地址",
          contact: "后端地址暂未同步",
          contactName: "",
          contactPhone: ""
        }
      });
    }
  },
  handleAddressTap() {
    wx.navigateTo({ url: "/pages/address/list/index?mode=select" });
  },
  async handleSubmit() {
    if (!this.data.addressId || !this.data.address.contactName || !this.data.address.contactPhone) {
      wx.showToast({ title: "请选择真实配送地址", icon: "none" });
      return;
    }
    ensurePreviewAuth();
    this.setData({ submitting: true });
    try {
      const detail = await createMedicineOrder({
        prescription_id: this.data.prescriptionId,
        address: this.data.address.detail,
        contact_name: this.data.address.contactName,
        contact_phone: this.data.address.contactPhone,
        clinic_name: this.data.clinic.name,
        delivery_fee_fen: 200,
        payment_method: "balance",
        items: this.data.items.map((item) => ({
          product_id: item.product_id,
          name: item.name,
          category: item.tag,
          image_url: item.imageUrl,
          price_fen: item.price_fen,
          quantity: item.quantity,
          requires_prescription: item.requires_prescription
        }))
      }) as { order?: { id?: string } };
      const orderId = detail.order?.id || "";
      wx.showToast({ title: "药品订单已提交", icon: "success" });
      setTimeout(() => wx.navigateTo({ url: `/pages/medicine/order-detail/index?id=${orderId}` }), 500);
    } catch (_error) {
      const message = _error instanceof Error ? _error.message : "";
      if (message.includes("stock") || message.includes("库存")) {
        wx.showToast({ title: "库存不足，请返回调整药品", icon: "none" });
        return;
      }
      wx.showToast({ title: "药品订单提交失败", icon: "none" });
    } finally {
      this.setData({ submitting: false });
    }
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/medicine/home/index" }) });
  }
});
