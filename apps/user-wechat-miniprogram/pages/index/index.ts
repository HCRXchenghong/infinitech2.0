import { getHomeModules } from "../../utils/api";
import { HOME_LOCATION_KEY } from "../../utils/location";
import { categoryIconImage, generatedImages, mediaUrl } from "../../utils/media";
import { openRoute } from "../../utils/navigation";

Page({
  data: {
    currentHeroIndex: 0,
    locationName: "定位中",
    locationAddress: "",
    locationReady: false,
    locationPoint: null,
    capabilityRoutes: {
      notificationPreferences: "/pages/notification-preferences/index"
    },
    heroSlides: [
      {
        id: "hero_takeout",
        title: "悦享e食",
        subtitle: "外卖 · 团购 · 买药 · 快递跑腿 · 圈子",
        note: "好服务 触手可及",
        imageUrl: generatedImages.homeHero,
        route: "/pages/shop/list/index"
      },
      {
        id: "hero_groupbuy",
        title: "今日团购",
        subtitle: "精选套餐 · 到店立减 · 学生友好",
        note: "低至 6.4 折",
        imageUrl: generatedImages.shopClaypot,
        route: "/pages/shop/list/index?tab=groupbuy"
      },
      {
        id: "hero_medicine",
        title: "买药即达",
        subtitle: "校医务室 · 常备药 · 急用速配",
        note: "合规药房 透明履约",
        imageUrl: generatedImages.medicineClinicCover,
        route: "/pages/medicine/home/index"
      }
    ],
    modules: [
      { key: "takeout", title: "外卖", iconGlyph: "外", iconImageUrl: categoryIconImage("takeout"), iconTone: "tone-blue", route: "/pages/shop/list/index" },
      { key: "groupbuy", title: "团购", iconGlyph: "团", iconImageUrl: categoryIconImage("groupbuy"), iconTone: "tone-orange", route: "/pages/shop/list/index?tab=groupbuy" },
      { key: "medicine", title: "买药", iconGlyph: "药", iconImageUrl: categoryIconImage("medicine"), iconTone: "tone-green", route: "/pages/medicine/home/index" },
      { key: "courier", title: "快递跑腿", iconGlyph: "跑", iconImageUrl: categoryIconImage("courier"), iconTone: "tone-sky", route: "/pages/errand/home/index" },
      { key: "circle", title: "圈子", iconGlyph: "圈", iconImageUrl: categoryIconImage("circle"), iconTone: "tone-purple", route: "/pages/circle/index" },
      { key: "meal-match", title: "找饭搭", iconGlyph: "搭", iconImageUrl: categoryIconImage("meal-match"), iconTone: "tone-pink", route: "/pages/meal-match/index" },
      { key: "coupons", title: "红包优惠", iconGlyph: "券", iconImageUrl: categoryIconImage("coupons"), iconTone: "tone-red", route: "/pages/coupons/index" },
      { key: "points", title: "会员积分", iconGlyph: "分", iconImageUrl: categoryIconImage("points"), iconTone: "tone-gold", route: "/pages/member-points/index" }
    ],
    groupDeals: [
      {
        id: "deal_hotpot",
        title: "呷哺呷哺 2-3人套餐",
        discountText: "6.4折",
        price: "99",
        originPrice: "155",
        soldText: "已售 2.1万+",
        imageUrl: generatedImages.shopClaypot,
        route: "/pages/shop/list/index?tab=groupbuy"
      },
      {
        id: "deal_coffee",
        title: "瑞幸咖啡 生椰拿铁 2杯",
        discountText: "5.7折",
        price: "28.9",
        originPrice: "50",
        soldText: "已售 3.4万+",
        imageUrl: generatedImages.shopCoffee,
        route: "/pages/shop/list/index?tab=groupbuy"
      },
      {
        id: "deal_ktv",
        title: "唱吧麦颂 KTV 欢唱套餐",
        discountText: "3.8折",
        price: "128",
        originPrice: "338",
        soldText: "已售 9800+",
        imageUrl: generatedImages.shopHomeCooking,
        route: "/pages/shop/list/index?tab=groupbuy"
      },
      {
        id: "deal_skewer",
        title: "木屋烧烤 100元券",
        discountText: "7.5折",
        price: "75",
        originPrice: "100",
        soldText: "已售 1.6万+",
        imageUrl: generatedImages.productChickenWings,
        route: "/pages/shop/list/index?tab=groupbuy"
      }
    ],
    takeoutPicks: [
      {
        id: "takeout_noodle",
        title: "牛肉面馆",
        subtitle: "招牌红烧牛肉面 · 热汤现做",
        ratingText: "★ 4.8",
        salesText: "月售 2000+",
        deliveryTime: "30分钟",
        deliveryFee: "配送 ¥3",
        tagText: "满30减6",
        price: "18.8",
        imageUrl: generatedImages.productTomatoNoodle,
        route: "/pages/shop/detail/index?id=shop_1"
      },
      {
        id: "takeout_spicy",
        title: "川味小馆",
        subtitle: "麻辣香锅 · 米饭套餐 · 夜宵",
        ratingText: "★ 4.7",
        salesText: "月售 1500+",
        deliveryTime: "35分钟",
        deliveryFee: "配送 ¥2",
        tagText: "第二份半价",
        price: "22.9",
        imageUrl: generatedImages.homeFeaturedDish,
        route: "/pages/shop/list/index"
      },
      {
        id: "takeout_light",
        title: "轻食主义",
        subtitle: "鸡胸肉沙拉 · 全麦能量碗",
        ratingText: "★ 4.9",
        salesText: "月售 1800+",
        deliveryTime: "25分钟",
        deliveryFee: "配送 ¥2",
        tagText: "新客立减",
        price: "19.9",
        imageUrl: generatedImages.productBeefRice,
        route: "/pages/shop/list/index"
      },
      {
        id: "takeout_burger",
        title: "快乐汉堡",
        subtitle: "双层牛堡 · 炸鸡 · 冰饮套餐",
        ratingText: "★ 4.6",
        salesText: "月售 1200+",
        deliveryTime: "30分钟",
        deliveryFee: "配送 ¥3",
        tagText: "满45减8",
        price: "26.8",
        imageUrl: generatedImages.shopHomeCooking,
        route: "/pages/shop/list/index"
      }
    ]
  },

  onLoad() {
    this.loadCachedLocation();
    this.loadUserLocation();
    this.loadHomeConfig();
  },

  onShow() {
    this.loadCachedLocation();
  },

  loadCachedLocation() {
    const cached = wx.getStorageSync(HOME_LOCATION_KEY);
    if (cached && cached.name) {
      this.setData({
        locationName: String(cached.name),
        locationAddress: String(cached.address || ""),
        locationReady: true,
        locationPoint: {
          latitude: Number(cached.latitude || 0),
          longitude: Number(cached.longitude || 0)
        }
      });
    }
  },

  loadUserLocation(showFailToast = false) {
    wx.getLocation({
      type: "gcj02",
      success: (res) => {
        if (this.data.locationReady) return;
        this.applyHomeLocation({
          name: "当前位置",
          address: "已获取真实定位，点击可选择具体地点",
          latitude: res.latitude,
          longitude: res.longitude,
          source: "gps"
        });
      },
      fail: () => {
        if (!this.data.locationReady) {
          this.setData({ locationName: "选择位置", locationAddress: "", locationReady: false });
        }
        if (showFailToast) {
          wx.showToast({ title: "请授权定位或手动选择位置", icon: "none" });
        }
      }
    });
  },

  handleLocationTap() {
    openRoute("/pages/location/select/index");
  },

  applyHomeLocation(location) {
    const normalized = {
      name: String(location.name || "当前位置"),
      address: String(location.address || ""),
      latitude: Number(location.latitude || 0),
      longitude: Number(location.longitude || 0),
      source: String(location.source || "gps"),
      updatedAt: Date.now()
    };
    wx.setStorageSync(HOME_LOCATION_KEY, normalized);
    this.setData({
      locationName: normalized.name,
      locationAddress: normalized.address,
      locationReady: true,
      locationPoint: {
        latitude: normalized.latitude,
        longitude: normalized.longitude
      }
    });
  },

  async loadHomeConfig() {
    try {
      const modules = await getHomeModules();
      const remoteModules = Array.isArray(modules)
        ? modules
            .filter((item) => item.enabled !== false)
            .sort((left, right) => Number(left.sort_order || 0) - Number(right.sort_order || 0))
            .map((item) => {
              const visual = moduleVisual(String(item.key || ""));
              return {
                key: String(item.key || ""),
                title: String(item.title || visual.title),
                iconGlyph: visual.iconGlyph,
                iconImageUrl: mediaUrl(item.icon_url, visual.iconImageUrl),
                iconTone: visual.iconTone,
                route: fallbackModuleRoute(String(item.key || ""), String(item.route || ""))
              };
            })
        : [];

      this.setData({
        modules: remoteModules.length > 0 ? remoteModules.slice(0, 8) : this.data.modules
      });
    } catch (_error) {
      // keep local preview data
    }
  },

  handleHeroChange(event: WechatMiniprogram.CustomEvent) {
    const detail = event.detail as { current?: number };
    this.setData({ currentHeroIndex: Number(detail.current || 0) });
  },

  handleHeroTap() {
    const slide = this.data.heroSlides[this.data.currentHeroIndex] || this.data.heroSlides[0];
    if (slide.route) {
      openRoute(slide.route);
    }
  },

  handleModuleTap(event: WechatMiniprogram.BaseEvent) {
    const route = String(event.currentTarget.dataset.route || "");
    if (route) {
      openRoute(route);
    }
  },

  handleSearchTap() {
    openRoute("/pages/search/index");
  },

  handleRecommendTap(event: WechatMiniprogram.BaseEvent) {
    const route = String(event.currentTarget.dataset.route || "");
    if (route) {
      openRoute(route);
    }
  }
});

function fallbackModuleRoute(key: string, route: string) {
  if (route) return route;
  const routeMap: Record<string, string> = {
    takeout: "/pages/shop/list/index",
    groupbuy: "/pages/shop/list/index?tab=groupbuy",
    medicine: "/pages/medicine/home/index",
    courier: "/pages/errand/home/index",
    circle: "/pages/circle/index",
    "meal-match": "/pages/meal-match/index",
    coupons: "/pages/coupons/index",
    points: "/pages/member-points/index"
  };
  return routeMap[key] || "/pages/shop/list/index";
}

function moduleVisual(key: string) {
  const map: Record<string, { title: string; iconGlyph: string; iconImageUrl: string; iconTone: string }> = {
    takeout: { title: "外卖", iconGlyph: "外", iconImageUrl: categoryIconImage("takeout"), iconTone: "tone-blue" },
    groupbuy: { title: "团购", iconGlyph: "团", iconImageUrl: categoryIconImage("groupbuy"), iconTone: "tone-orange" },
    medicine: { title: "买药", iconGlyph: "药", iconImageUrl: categoryIconImage("medicine"), iconTone: "tone-green" },
    courier: { title: "快递跑腿", iconGlyph: "跑", iconImageUrl: categoryIconImage("courier"), iconTone: "tone-sky" },
    circle: { title: "圈子", iconGlyph: "圈", iconImageUrl: categoryIconImage("circle"), iconTone: "tone-purple" },
    "meal-match": { title: "找饭搭", iconGlyph: "搭", iconImageUrl: categoryIconImage("meal-match"), iconTone: "tone-pink" },
    coupons: { title: "红包优惠", iconGlyph: "券", iconImageUrl: categoryIconImage("coupons"), iconTone: "tone-red" },
    points: { title: "会员积分", iconGlyph: "分", iconImageUrl: categoryIconImage("points"), iconTone: "tone-gold" }
  };
  return map[key] || { title: "功能", iconGlyph: "功", iconImageUrl: categoryIconImage("takeout"), iconTone: "tone-blue" };
}
