import { checkoutCart, setDevAuthToken } from "../../../utils/api";

Page({
  data: {
    address: {
      name: "张三",
      phone: "13800000000",
      detail: "北京 望京SOHO"
    },
    items: [
      { name: "招牌牛肉饭", quantity: 2, price: "51.98" }
    ],
    options: {
      tableware: "2 份",
      remark: "少放辣"
    },
    payable: "55.98"
  },
  handleAddressTap() {
    wx.navigateTo({ url: "/pages/address/list/index" });
  },
  async handleSubmitTap() {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      await checkoutCart("shop_1", "addr_1", {
        remark: this.data.options.remark,
        tableware_count: 2
      });
      wx.showToast({ title: "订单已创建", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "预览订单已提交", icon: "none" });
    }
  }
});
