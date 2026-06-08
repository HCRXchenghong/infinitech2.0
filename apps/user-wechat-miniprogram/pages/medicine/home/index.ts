import { ensurePreviewAuth, getMedicineHome } from "../../../utils/api";
import { generatedImages, mediaUrl, medicineFallbackImage } from "../../../utils/media";

const previewMedicineProducts = [
  productFromApi({ id: "med_cooling_patch", name: "退热贴", subtitle: "医用退热贴 · 适用于发热物理降温", category: "感冒发热", image_url: generatedImages.medicineCoolingPatch, price_fen: 1290, stock_count: 26, selected_quantity: 1 }),
  productFromApi({ id: "med_amoxicillin", name: "阿莫西林胶囊", subtitle: "处方药 · 凭处方与校医审核购买", category: "处方药", image_url: generatedImages.medicineCapsules, price_fen: 1880, stock_count: 12, requires_prescription: true }),
  productFromApi({ id: "med_swab", name: "碘伏棉签", subtitle: "外伤消毒 · 独立包装", category: "外伤消毒", image_url: generatedImages.medicineFirstAid, price_fen: 690, stock_count: 38 }),
  productFromApi({ id: "med_bandage", name: "创可贴", subtitle: "防水透气 · 10 片装", category: "医用耗材", image_url: generatedImages.medicineFirstAid, price_fen: 550, stock_count: 52 })
];

Page({
  data: {
    statusText: "已同步",
    clinic: {
      name: "校医务室",
      location: "综合楼一层",
      coverUrl: generatedImages.medicineClinicCover,
      businessTime: "今日 08:30-20:30",
      deliveryText: "校内配送约 20-30 分钟"
    },
    categories: [
      { title: "全部", active: "active" },
      { title: "感冒发热", active: "" },
      { title: "处方药", active: "" },
      { title: "外伤消毒", active: "" },
      { title: "医用耗材", active: "" }
    ],
    selectedCategory: "全部",
    products: previewMedicineProducts,
    allProducts: previewMedicineProducts,
    cartCount: 1,
    cartAmount: "12.90"
  },
  onLoad() {
    this.loadMedicineHome();
  },
  async loadMedicineHome() {
    ensurePreviewAuth();
    try {
      const home = await getMedicineHome() as Record<string, unknown>;
      const products = (home.products as Array<Record<string, unknown>> || []).map(productFromApi);
      const clinic = clinicFromApi(home.clinic as Record<string, unknown>, this.data.clinic);
      this.setData({
        clinic,
        categories: (home.categories as string[] || ["全部", "感冒发热", "处方药", "外伤消毒", "医用耗材"]).map((title) => ({ title, active: title === this.data.selectedCategory ? "active" : "" })),
        allProducts: products.length ? products : previewMedicineProducts,
        products: products.length ? products : previewMedicineProducts,
        cartCount: Number(home.cart_count || 0),
        cartAmount: formatFen(Number(home.cart_amount_fen || 0)),
        statusText: "已同步"
      });
    } catch (_error) {
      this.setData({ statusText: "离线缓存" });
    }
  },
  handleCategoryTap(event) {
    const selectedCategory = String(event.currentTarget.dataset.title || "全部");
    const allProducts = this.data.allProducts.length ? this.data.allProducts : this.data.products;
    this.setData({
      selectedCategory,
      categories: this.data.categories.map((item) => ({ ...item, active: item.title === selectedCategory ? "active" : "" })),
      products: selectedCategory === "全部" ? allProducts : allProducts.filter((item) => item.category === selectedCategory)
    });
  },
  handleSearchTap() {
    wx.navigateTo({ url: "/pages/search/index?keyword=药" });
  },
  handleConsultTap() {
    wx.navigateTo({ url: "/pages/customer-service/chat/index?type=medicine" });
  },
  handlePrescriptionTap() {
    wx.navigateTo({ url: "/pages/prescription/upload/index" });
  },
  handleAddProduct(event) {
    const productId = String(event.currentTarget.dataset.id || "");
    const product = this.data.products.find((item) => item.id === productId);
    if (product?.requiresPrescription) {
      wx.navigateTo({ url: "/pages/prescription/upload/index" });
      return;
    }
    const allProducts = this.data.allProducts.map((item) => item.id === productId ? { ...item, selectedQuantity: Number(item.selectedQuantity || 0) + 1 } : item);
    const cartCount = Number(this.data.cartCount || 0) + 1;
    const cartAmountFen = Math.round(Number(this.data.cartAmount || "0") * 100) + Number(product?.priceFen || 0);
    this.setData({
      allProducts,
      products: this.data.selectedCategory === "全部" ? allProducts : allProducts.filter((item) => item.category === this.data.selectedCategory),
      cartCount,
      cartAmount: formatFen(cartAmountFen)
    });
  },
  handleCheckout() {
    wx.navigateTo({ url: "/pages/medicine/order-confirm/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function productFromApi(item: Record<string, unknown>) {
  return {
    id: String(item.id || Date.now()),
    name: String(item.name || "药品"),
    subtitle: String(item.subtitle || ""),
    category: String(item.category || "全部"),
    imageUrl: mediaUrl(item.image_url, medicineFallbackImage(String(item.id || ""))),
    priceText: `¥${formatFen(Number(item.price_fen || 0))}`,
    priceFen: Number(item.price_fen || 0),
    stockText: `库存 ${item.stock_count || 0}`,
    requiresPrescription: Boolean(item.requires_prescription),
    prescriptionText: item.requires_prescription ? "待上传处方" : "",
    selectedQuantity: Number(item.selected_quantity || 0)
  };
}

function clinicFromApi(item: Record<string, unknown> = {}, fallback: Record<string, unknown>) {
  return {
    name: String(item.name || fallback.name || "校医务室"),
    location: String(item.location || fallback.location || "综合楼一层"),
    coverUrl: mediaUrl(item.cover_url, fallback.coverUrl || generatedImages.medicineClinicCover),
    businessTime: String(item.business_time || fallback.businessTime || "今日 08:30-20:30"),
    deliveryText: String(item.delivery_text || fallback.deliveryText || "校内配送约 20-30 分钟")
  };
}

function formatFen(amountFen: number) {
  return (amountFen / 100).toFixed(2);
}
