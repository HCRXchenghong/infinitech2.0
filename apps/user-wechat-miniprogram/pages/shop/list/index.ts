import { getShops } from "../../../utils/api";
import { generatedImages, mediaUrl, shopFallbackImage } from "../../../utils/media";

Page({
  data: {
    activeTab: "takeout",
    filters: [
      { key: "default", title: "综合排序" },
      { key: "sales", title: "销量最高" },
      { key: "distance", title: "距离最近" },
      { key: "filter", title: "筛选" }
    ],
    shops: [
      {
        id: "shop_1",
        initial: "蓝",
        name: "蓝海餐厅",
        imageUrl: generatedImages.homeRecommendRestaurant,
        category: "简餐 · 团购",
        rating: "4.8",
        monthlySales: "月售 2381",
        delivery: "约 32 分钟 · 配送 ¥3",
        distance: "1.8km",
        tags: ["满减", "准时达", "可团购"],
        announcement: "招牌牛肉饭热卖，团购套餐到店扫码验券。"
      },
      {
        id: "shop_coffee_preview",
        initial: "咖",
        name: "晴川咖啡",
        imageUrl: generatedImages.shopCoffee,
        category: "咖啡 · 下午茶",
        rating: "4.9",
        monthlySales: "月售 980",
        delivery: "约 25 分钟 · 配送 ¥2",
        distance: "900m",
        tags: ["下午茶", "新人券"],
        announcement: "拿铁第二杯半价，早餐时段可预约。"
      },
      {
        id: "shop_medicine_preview",
        initial: "药",
        name: "安心药房",
        imageUrl: generatedImages.shopPharmacy,
        category: "药房 · 买药",
        rating: "4.8",
        monthlySales: "月售 620",
        delivery: "约 38 分钟 · 配送 ¥5",
        distance: "2.4km",
        tags: ["买药", "夜间可送", "资质已审核"],
        announcement: "处方药需上传处方并通过药师审核。"
      },
      {
        id: "shop_kitchen_preview",
        initial: "厨",
        name: "邻里小厨",
        imageUrl: generatedImages.shopHomeCooking,
        category: "家常菜",
        rating: "4.7",
        monthlySales: "月售 1560",
        delivery: "约 28 分钟 · 配送 ¥3",
        distance: "1.2km",
        tags: ["家常菜", "满减", "热卖"],
        announcement: "午餐套餐持续供应，支持少油少辣备注。"
      },
      {
        id: "shop_pot_preview",
        initial: "砂",
        name: "好味砂锅",
        imageUrl: generatedImages.shopClaypot,
        category: "砂锅 · 热汤",
        rating: "4.6",
        monthlySales: "月售 1320",
        delivery: "约 30 分钟 · 配送 ¥3",
        distance: "1.6km",
        tags: ["砂锅", "热汤", "新人券"],
        announcement: "砂锅粉和热汤销量靠前。"
      }
    ],
    activeFilter: "default"
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
          shops: shops.map((shop, index) => ({
            id: String(shop.id || `shop_${index}`),
            initial: String(shop.name || "店").slice(0, 1),
            name: String(shop.name || "本地商家"),
            imageUrl: mediaUrl(shop.cover_url || shop.logo_url || shop.image_url, shopFallbackImage(String(shop.id || ""), index)),
            category: shop.category || "本地生活",
            rating: String(shop.rating || "4.8"),
            monthlySales: "月售 2381",
            delivery: "约 32 分钟 · 配送 ¥3",
            distance: "1.8km",
            tags: normalizeTags(shop.capabilities),
            announcement: shop.announcement || ""
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleFilterTap(event) {
    this.setData({ activeFilter: String(event.currentTarget.dataset.key || "default") });
  },
  handleShopTap(event) {
    wx.navigateTo({ url: `/pages/shop/detail/index?id=${event.currentTarget.dataset.id}` });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function normalizeTags(capabilities: string[] = []) {
  const tags = capabilities.map((item) => {
    const map: Record<string, string> = {
      takeout: "外卖",
      groupbuy: "可团购",
      medicine: "买药",
      courier: "跑腿"
    };
    return map[item] || item;
  });
  return tags.length > 0 ? tags.concat(["准时达"]).slice(0, 3) : ["满减", "准时达"];
}
