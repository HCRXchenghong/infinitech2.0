import { checkoutCart, createWechatPrepay, ensurePreviewAuth, getCartSummary, getUserAddresses, payOrderWithBalance } from "../../../utils/api";
import { generatedImages, mediaUrl, productFallbackImage } from "../../../utils/media";

Page({
  data: {
    shopId: "shop_1",
    shopName: "蓝海餐厅",
    addressId: "",
    addressStatusText: "请选择配送地址",
    address: {
      name: "",
      phone: "",
      detail: "请选择真实配送地址"
    },
    items: [
      { id: "prod_beef_rice", name: "招牌牛肉饭", imageUrl: generatedImages.productBeefRice, quantity: 2, price: "51.98" },
      { id: "prod_tea", name: "柠檬茶", imageUrl: generatedImages.productLemonTea, quantity: 1, price: "9.00" }
    ],
    options: {
      coupon: "满45减8",
      tableware: "2 份",
      remark: "少放辣"
    },
    fees: [
      { title: "商品小计", value: "¥78.98" },
      { title: "配送费", value: "¥3.00" },
      { title: "打包费", value: "¥1.00" },
      { title: "优惠券", value: "-¥8.00", discount: true },
      { title: "合计", value: "¥74.98" }
    ],
    payable: "74.98",
    cartReady: false,
    submitting: false
  },
  onLoad(query) {
    this.setData({ shopId: String(query?.shop_id || "shop_1") });
  },
  onShow() {
    this.loadCheckoutContext();
  },
  async loadCheckoutContext() {
    ensurePreviewAuth();
    await Promise.all([this.loadAddresses(), this.loadCart()]);
  },
  async loadAddresses() {
    try {
      const addresses = await getUserAddresses() as Array<Record<string, any>>;
      if (!Array.isArray(addresses) || addresses.length === 0) {
        this.setData({
          addressId: "",
          addressStatusText: "请先新增一个配送地址",
          address: {
            name: "",
            phone: "",
            detail: "暂无后端配送地址"
          }
        });
        return;
      }
      const storedId = String(wx.getStorageSync("checkoutAddressId") || "");
      const selected = addresses.find((item) => String(item.id || "") === storedId) || addresses.find((item) => Boolean(item.is_default)) || addresses[0];
      if (!selected) {
        return;
      }
      const detail = `${selected.city || ""} ${selected.detail || ""}`.trim();
      wx.setStorageSync("checkoutAddressId", String(selected.id || ""));
      this.setData({
        addressId: String(selected.id || ""),
        addressStatusText: selected.is_default ? "默认地址" : "已选择地址",
        address: {
          name: String(selected.contact_name || ""),
          phone: String(selected.contact_phone || ""),
          detail: detail || "请补充配送地址"
        }
      });
    } catch (_error) {
      this.setData({
        addressId: "",
        addressStatusText: "地址同步失败",
        address: {
          name: "",
          phone: "",
          detail: "请稍后重试或新增真实地址"
        }
      });
    }
  },
  async loadCart() {
    try {
      const summary = await getCartSummary(String(this.data.shopId || "shop_1")) as Record<string, any>;
      const items = Array.isArray(summary.items) ? summary.items : [];
      const discountFen = Number(summary.discount_fen || 0);
      const fees = [
        { title: "商品小计", value: `¥${fenToText(summary.items_total_fen)}` },
        { title: "配送费", value: `¥${fenToText(summary.delivery_fee_fen)}` },
        { title: "打包费", value: `¥${fenToText(summary.packaging_fee_fen)}` }
      ];
      if (discountFen > 0) {
        fees.push({ title: "优惠券", value: `-¥${fenToText(discountFen)}`, discount: true });
      }
      fees.push({ title: "合计", value: `¥${fenToText(summary.payable_fen)}` });
      this.setData({
        shopName: String(summary.shop_name || this.data.shopName),
        items: items.map((item) => ({
          id: String(item.product_id || ""),
          name: String(item.product_name || ""),
          imageUrl: mediaUrl(item.image_url, productFallbackImage(String(item.product_id || ""))),
          quantity: Number(item.quantity || 0),
          price: fenToText(Number(item.unit_price_fen || 0) * Number(item.quantity || 0))
        })),
        fees,
        payable: fenToText(summary.payable_fen),
        cartReady: items.length > 0
      });
    } catch (_error) {
      // 保留页面兜底数据，便于无后端时预览。
    }
  },
  handleAddressTap() {
    wx.navigateTo({ url: "/pages/address/list/index?mode=select" });
  },
  async handleSubmitTap() {
    if (!this.data.addressId) {
      wx.showToast({ title: "请选择配送地址", icon: "none" });
      return;
    }
    if (!this.data.cartReady || this.data.items.length === 0) {
      wx.showToast({ title: "购物车为空", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      const result = await checkoutCart(String(this.data.shopId || "shop_1"), String(this.data.addressId || ""), {
        remark: this.data.options.remark,
        tableware_count: 2
      });
      const order = result.order || result;
      this.choosePayment(String(order.id || "ord_preview"));
    } catch (_error) {
      wx.showToast({ title: "预览订单已提交", icon: "none" });
      this.choosePayment("ord_preview");
    } finally {
      this.setData({ submitting: false });
    }
  },
  choosePayment(orderId: string) {
    wx.showActionSheet({
      itemList: ["微信支付", "余额支付"],
      success: async (result) => {
        try {
          if (result.tapIndex === 0) {
            await createWechatPrepay(orderId);
            wx.showToast({ title: "微信预支付已创建", icon: "success" });
          } else {
            await payOrderWithBalance(orderId, "123456");
            wx.showToast({ title: "余额支付成功", icon: "success" });
          }
        } catch (_error) {
          wx.showToast({ title: "支付预览完成", icon: "none" });
        }
        wx.navigateTo({ url: `/pages/order/detail/index?id=${orderId}` });
      },
      fail: () => {
        wx.navigateTo({ url: `/pages/order/detail/index?id=${orderId}` });
      }
    });
  },
  handleBack() {
    wx.navigateBack({
      fail() {
        wx.redirectTo({ url: "/pages/cart/index" });
      }
    });
  },
  handleOptionTap() {
    wx.showToast({ title: "订单选项编辑下一步精修", icon: "none" });
  }
});

function fenToText(value: unknown) {
  return (Number(value || 0) / 100).toFixed(2);
}
