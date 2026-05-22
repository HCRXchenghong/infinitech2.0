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
    { key: "takeout", title: "外卖", route: "/pages/shop/list/index", icon: "takeout", enabled: true, sort_order: 10, scene: "home" },
    { key: "groupbuy", title: "团购", route: "/pages/shop/list/index?tab=groupbuy", icon: "groupbuy", enabled: true, sort_order: 20, scene: "home" },
    { key: "medicine", title: "买药", route: "/pages/shop/list/index?tab=medicine", icon: "medicine", enabled: true, sort_order: 30, scene: "home" },
    { key: "courier", title: "快递跑腿", route: "/pages/shop/list/index?tab=courier", icon: "courier", enabled: true, sort_order: 40, scene: "home" },
    { key: "circle", title: "圈子", route: "/pages/circle/index", icon: "circle", enabled: true, sort_order: 50, scene: "home" },
    { key: "charity", title: "公益", route: "/pages/charity/index", icon: "charity", enabled: false, sort_order: 60, scene: "home" },
    { key: "social", title: "交友", route: "/pages/social/index", icon: "social", enabled: false, sort_order: 70, scene: "home" }
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
      image_url: "",
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
      image_url: "",
      price_fen: 0,
      enabled: true,
      sort_order: 20
    }
  ];
}
