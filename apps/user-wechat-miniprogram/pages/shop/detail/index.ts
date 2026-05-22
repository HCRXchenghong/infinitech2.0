import { createGroupbuyOrder, getShopGroupbuyDeals, getShopProducts, setDevAuthToken, upsertCartItem } from "../../../utils/api";

Page({
  data: {
    shopId: "shop_1",
    shop: {
      name: "蓝海餐厅",
      rating: "4.8",
      delivery: "约 32 分钟送达",
      announcement: "营业执照和健康证已审核。团购券到店扫码即可核销。"
    },
    products: [
      {
        id: "prod_beef_rice",
        name: "招牌牛肉饭",
        description: "牛肉、米饭、时蔬，配料表可由商户后台维护。",
        price: "25.99",
        count: 1
      },
      {
        id: "prod_soup",
        name: "每日例汤",
        description: "随餐热汤，库存不足时可触发自动退款规则。",
        price: "5.99",
        count: 0
      }
    ],
    groupbuyDeals: [
      {
        id: "deal_two_person_set",
        title: "双人到店套餐券",
        price: "39.99",
        note: "购买后生成二维码，商户端扫码验券。"
      }
    ]
  },
  onLoad(query) {
    const shopId = String(query?.id || "shop_1");
    this.setData({ shopId });
    this.loadProducts(shopId);
    this.loadGroupbuyDeals(shopId);
  },
  async loadProducts(shopId) {
    try {
      const products = await getShopProducts(shopId);
      if (Array.isArray(products) && products.length > 0) {
        this.setData({
          products: products.map((product) => ({
            id: product.id,
            name: product.name,
            description: product.description,
            price: (Number(product.price_fen || 0) / 100).toFixed(2),
            count: 0
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  async loadGroupbuyDeals(shopId) {
    try {
      const deals = await getShopGroupbuyDeals(shopId);
      if (Array.isArray(deals) && deals.length > 0) {
        this.setData({
          groupbuyDeals: deals.map((deal) => ({
            id: deal.id,
            title: deal.name,
            note: deal.description,
            price: (Number(deal.price_fen || 0) / 100).toFixed(2)
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  async handleAddTap(event) {
    const productId = String(event.currentTarget.dataset.id || "");
    if (!productId) {
      return;
    }
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      await upsertCartItem(String(this.data.shopId), productId, 1);
      wx.showToast({ title: "已加入购物车", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "已加入预览购物车", icon: "none" });
    }
  },
  async handleGroupbuyTap(event) {
    const dealId = String(event.currentTarget.dataset.id || "");
    if (!dealId) {
      return;
    }
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      const order = await createGroupbuyOrder(String(this.data.shopId), dealId, 1);
      wx.navigateTo({ url: `/pages/order/detail/index?id=${order.id}` });
    } catch (_error) {
      wx.showToast({ title: "团购券已加入预览订单", icon: "none" });
    }
  },
  handleCartTap() {
    wx.navigateTo({ url: "/pages/cart/index" });
  }
});
