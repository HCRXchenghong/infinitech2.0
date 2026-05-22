import { getCartSummary, setDevAuthToken } from "../../utils/api";

Page({
  data: {
    shopName: "蓝海餐厅",
    items: [
      { id: "prod_beef_rice", name: "招牌牛肉饭", price: "25.99", quantity: 2, selected: true }
    ],
    summary: {
      itemsTotal: "51.98",
      deliveryFee: "3.00",
      packagingFee: "1.00",
      payable: "55.98"
    }
  },
  onLoad() {
    this.loadCart();
  },
  async loadCart() {
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      const summary = await getCartSummary("shop_1");
      const data = summary as {
        items?: Array<{ product_id: string; product_name: string; unit_price_fen: number; quantity: number; selected: boolean }>;
        items_total_fen?: number;
        delivery_fee_fen?: number;
        packaging_fee_fen?: number;
        payable_fen?: number;
      };
      if (data?.items && data.items.length > 0) {
        this.setData({
          items: data.items.map((item) => ({
            id: item.product_id,
            name: item.product_name,
            price: (Number(item.unit_price_fen || 0) / 100).toFixed(2),
            quantity: item.quantity,
            selected: item.selected
          })),
          summary: {
            itemsTotal: (Number(data.items_total_fen || 0) / 100).toFixed(2),
            deliveryFee: (Number(data.delivery_fee_fen || 0) / 100).toFixed(2),
            packagingFee: (Number(data.packaging_fee_fen || 0) / 100).toFixed(2),
            payable: (Number(data.payable_fen || 0) / 100).toFixed(2)
          }
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleCheckoutTap() {
    wx.navigateTo({ url: "/pages/order/confirm/index" });
  }
});
