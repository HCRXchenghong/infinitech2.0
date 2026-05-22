Page({
  data: {
    modules: [
      { key: "takeout", title: "外卖", primary: true, route: "/pages/shop/list/index" },
      { key: "groupbuy", title: "团购", primary: true, route: "/pages/shop/list/index?tab=groupbuy" },
      { key: "medicine", title: "买药", primary: false, route: "/pages/shop/list/index?tab=medicine" },
      { key: "courier", title: "快递跑腿", primary: false, route: "/pages/shop/list/index?tab=courier" },
      { key: "circle", title: "圈子", primary: false, route: "/pages/circle/index" }
    ],
    homeCards: [
      {
        id: "card_takeout_sample",
        title: "后台推荐商品位",
        subtitle: "商品、店铺、团购券和圈子动态都由后台控制"
      },
      {
        id: "card_circle_sample",
        title: "圈子小微墙",
        subtitle: "发动态，找饭搭，查看附近活动"
      }
    ]
  },
  handleModuleTap(event: WechatMiniprogram.BaseEvent) {
    const route = String(event.currentTarget.dataset.route || "");
    if (route) {
      wx.navigateTo({ url: route });
    }
  }
});
