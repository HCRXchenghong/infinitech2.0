export function createRuntimeConfig(env = process.env) {
  return {
    service: "bff",
    apiBaseUrl: env.API_BASE_URL || "http://127.0.0.1:1029",
    realtimeUrl: env.REALTIME_URL || "ws://127.0.0.1:9898/ws",
    homeModulesMode: "configurable",
    homeCardsMode: "admin_configurable",
    circleFeatureMode: env.CIRCLE_FEATURE_MODE || "wall_only",
    clientKinds: ["wechat-miniprogram", "merchant-uni", "rider-uni", "admin-web", "admin-uni"],
    brand: {
      primaryColor: "#009bf5",
      logo: "/assets/brand/logo.svg"
    }
  };
}

export function defaultHomeModules() {
  return [
    { key: "takeout", title: "外卖", route: "/pages/shop/list/index", icon: "takeout", icon_url: "/assets/generated/category-takeout.png", enabled: true, sort_order: 10, scene: "home" },
    { key: "groupbuy", title: "团购", route: "/pages/shop/list/index?tab=groupbuy", icon: "groupbuy", icon_url: "/assets/generated/category-groupbuy.png", enabled: true, sort_order: 20, scene: "home" },
    { key: "medicine", title: "买药", route: "/pages/medicine/home/index", icon: "medicine", icon_url: "/assets/generated/category-medicine.png", enabled: true, sort_order: 30, scene: "home" },
    { key: "courier", title: "快递跑腿", route: "/pages/errand/home/index", icon: "courier", icon_url: "/assets/generated/category-courier.png", enabled: true, sort_order: 40, scene: "home" },
    { key: "circle", title: "圈子", route: "/pages/circle/index", icon: "circle", icon_url: "/assets/generated/category-circle.png", enabled: true, sort_order: 50, scene: "home" },
    { key: "charity", title: "公益", route: "/pages/charity/index", icon: "charity", enabled: false, sort_order: 60, scene: "home" },
    { key: "social", title: "交友", route: "/pages/social/index", icon: "social", enabled: false, sort_order: 70, scene: "home" },
    { key: "meal-match", title: "找饭搭", route: "/pages/meal-match/index", icon: "meal-match", icon_url: "/assets/generated/category-meal-match.png", enabled: true, sort_order: 80, scene: "home" },
    { key: "coupons", title: "红包优惠", route: "/pages/coupons/index", icon: "coupons", icon_url: "/assets/generated/category-coupons.png", enabled: true, sort_order: 90, scene: "home" },
    { key: "points", title: "会员积分", route: "/pages/member-points/index", icon: "points", icon_url: "/assets/generated/category-points.png", enabled: true, sort_order: 100, scene: "home" }
  ];
}

export function defaultHomeCards() {
  return [
    {
      id: "card_takeout_sample",
      type: "product",
      title: "后台推荐商品位",
      subtitle: "商品、店铺、团购和圈子内容都由后台控制",
      target_id: "product_placeholder",
      image_url: "/assets/generated/home-recommend-restaurant.jpg",
      price_fen: 0,
      enabled: true,
      sort_order: 10
    },
    {
      id: "card_circle_sample",
      type: "circle_post",
      title: "圈子小微墙",
      subtitle: "轻量动态和饭搭入口",
      target_id: "circle_micro_wall",
      image_url: "/assets/generated/home-meal-match.jpg",
      price_fen: 0,
      enabled: true,
      sort_order: 20
    }
  ];
}
