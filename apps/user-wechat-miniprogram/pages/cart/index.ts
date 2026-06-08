import { getCartSummary, setDevAuthToken, upsertCartItem } from "../../utils/api";
import { generatedImages, mediaUrl, productFallbackImage } from "../../utils/media";

Page({
  data: {
    shopId: "shop_1",
    shopName: "蓝海餐厅",
    items: [
      { id: "prod_beef_rice", name: "招牌牛肉饭", imageUrl: generatedImages.productBeefRice, price: "51.98", quantity: 2, selected: true, thumb: "饭" },
      { id: "prod_tea", name: "柠檬茶", imageUrl: generatedImages.productLemonTea, price: "9.00", quantity: 1, selected: true, thumb: "茶" },
      { id: "prod_wings", name: "烤鸡翅", imageUrl: generatedImages.productChickenWings, price: "18.00", quantity: 1, selected: true, thumb: "翅" }
    ],
    summary: {
      itemsTotal: "78.98",
      deliveryFee: "3.00",
      packagingFee: "1.00",
      discount: "8.00",
      payable: "74.98"
    },
    discountNote: "商品 ¥78.98 · 配送 ¥3.00 · 打包 ¥1.00 · 已优惠 ¥8.00",
    checkedAll: true,
    loading: false,
    updating: false
  },
  onLoad(query) {
    this.setData({ shopId: String(query?.shop_id || "shop_1") });
    this.loadCart();
  },
  async loadCart() {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    this.setData({ loading: true });
    try {
      const summary = await getCartSummary(String(this.data.shopId || "shop_1"));
      const data = summary as {
        shop_name?: string;
        items?: Array<{ product_id: string; product_name: string; image_url?: string; unit_price_fen: number; quantity: number; selected: boolean }>;
        items_total_fen?: number;
        delivery_fee_fen?: number;
        packaging_fee_fen?: number;
        discount_fen?: number;
        payable_fen?: number;
      };
      const items = Array.isArray(data?.items) ? data.items : [];
      const nextSummary = {
        itemsTotal: (Number(data?.items_total_fen || 0) / 100).toFixed(2),
        deliveryFee: (Number(data?.delivery_fee_fen || 0) / 100).toFixed(2),
        packagingFee: (Number(data?.packaging_fee_fen || 0) / 100).toFixed(2),
        discount: (Number(data?.discount_fen || 0) / 100).toFixed(2),
        payable: (Number(data?.payable_fen || 0) / 100).toFixed(2)
      };
      this.setData({
        shopName: String(data?.shop_name || this.data.shopName),
        items: items.map((item) => ({
          id: item.product_id,
          name: item.product_name,
          imageUrl: mediaUrl(item.image_url, productFallbackImage(String(item.product_id || ""))),
          price: (Number(item.unit_price_fen || 0) / 100).toFixed(2),
          quantity: item.quantity,
          selected: item.selected,
          thumb: item.product_name ? String(item.product_name).slice(0, 1) : "餐"
        })),
        summary: nextSummary,
        discountNote: formatCartNote(nextSummary),
        checkedAll: items.length > 0 && items.every((item) => Boolean(item.selected))
      });
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    } finally {
      this.setData({ loading: false });
    }
  },
  handleCloseTap() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: `/pages/shop/detail/index?id=${encodeURIComponent(String(this.data.shopId || "shop_1"))}` }) });
  },
  async handleClearTap() {
    if (this.data.updating || this.data.items.length === 0) {
      return;
    }
    this.setData({ updating: true });
    try {
      await Promise.all(this.data.items.map((item) => upsertCartItem(String(this.data.shopId || "shop_1"), item.id, 0)));
      wx.showToast({ title: "购物车已清空", icon: "success" });
      this.loadCart();
    } catch (_error) {
      const nextSummary = { ...this.data.summary, payable: "0.00", itemsTotal: "0.00", packagingFee: "0.00", deliveryFee: "0.00", discount: "0.00" };
      this.setData({ items: [], summary: nextSummary, discountNote: formatCartNote(nextSummary), checkedAll: false });
    } finally {
      this.setData({ updating: false });
    }
  },
  async handleQuantityTap(event) {
    if (this.data.updating) {
      return;
    }
    const id = String(event.currentTarget.dataset.id || "");
    const delta = Number(event.currentTarget.dataset.delta || 0);
    const current = this.data.items.find((item) => item.id === id);
    if (!id || !delta || !current) {
      return;
    }
    const nextQuantity = Math.max(0, Number(current.quantity || 0) + delta);
    this.setData({ updating: true });
    try {
      await upsertCartItem(String(this.data.shopId || "shop_1"), id, nextQuantity);
      this.loadCart();
    } catch (_error) {
      wx.showToast({ title: delta > 0 ? "加购失败" : "更新失败", icon: "none" });
    } finally {
      this.setData({ updating: false });
    }
  },
  handleCheckoutTap() {
    if (this.data.items.length === 0) {
      wx.showToast({ title: "先选点商品", icon: "none" });
      return;
    }
    wx.navigateTo({ url: `/pages/order/confirm/index?shop_id=${encodeURIComponent(String(this.data.shopId || "shop_1"))}` });
  }
});

function formatCartNote(summary: { itemsTotal: string; deliveryFee: string; packagingFee: string; discount: string }) {
  const parts = [
    `商品 ¥${summary.itemsTotal}`,
    `配送 ¥${summary.deliveryFee}`,
    `打包 ¥${summary.packagingFee}`
  ];
  if (Number(summary.discount || 0) > 0) {
    parts.push(`已优惠 ¥${summary.discount}`);
  }
  return parts.join(" · ");
}
