import { getShops } from "../../../utils/api";

Page({
  data: {
    activeTab: "takeout",
    tabs: [
      { key: "takeout", title: "外卖" },
      { key: "groupbuy", title: "团购" },
      { key: "medicine", title: "买药" },
      { key: "courier", title: "跑腿" }
    ],
    shops: [
      {
        id: "shop_1",
        name: "蓝海餐厅",
        category: "简餐 · 团购",
        rating: "4.8",
        monthlySales: "月售 1200+",
        delivery: "约 32 分钟 · 配送 ¥3",
        distance: "1.8km",
        tags: ["外卖", "团购", "健康证已审核"],
        announcement: "招牌牛肉饭热卖，团购套餐到店扫码验券。"
      }
    ]
  },
  onLoad(query) {
    const tab = String(query?.tab || "takeout");
    this.setData({ activeTab: tab });
    this.loadShops();
  },
  async loadShops() {
    try {
      const shops = await getShops();
      if (Array.isArray(shops) && shops.length > 0) {
        this.setData({
          shops: shops.map((shop) => ({
            id: shop.id,
            name: shop.name,
            category: shop.category || "本地生活",
            rating: String(shop.rating || "4.8"),
            monthlySales: "月售 1200+",
            delivery: "约 32 分钟 · 配送 ¥3",
            distance: "1.8km",
            tags: shop.capabilities || ["外卖"],
            announcement: shop.announcement || ""
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleTabTap(event) {
    this.setData({ activeTab: String(event.currentTarget.dataset.key || "takeout") });
  },
  handleShopTap(event) {
    wx.navigateTo({ url: `/pages/shop/detail/index?id=${event.currentTarget.dataset.id}` });
  }
});
