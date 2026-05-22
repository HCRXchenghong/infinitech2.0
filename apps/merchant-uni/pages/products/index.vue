<template>
  <view class="page">
    <view class="topbar">
      <button class="icon-button" @click="back">‹</button>
      <view>
        <text class="title">商品管理</text>
        <text class="subtitle">{{ products.length }} 个商品</text>
      </view>
      <button class="create-button" @click="createDemo">新增</button>
    </view>

    <scroll-view scroll-y class="product-list">
      <view v-for="product in products" :key="product.id" class="product-card">
        <view class="media" :class="{ empty: !product.image_url }">
          <image v-if="product.image_url" :src="product.image_url" mode="aspectFill" />
          <text v-else>{{ product.name.slice(0, 1) }}</text>
        </view>
        <view class="content">
          <view class="head">
            <text class="name">{{ product.name }}</text>
            <text class="status" :class="product.status">{{ statusLabel(product.status) }}</text>
          </view>
          <text class="desc">{{ product.description || "暂无描述" }}</text>
          <text class="ingredients">{{ (product.ingredient_list || []).join(" / ") }}</text>
          <view class="meta">
            <text class="price">¥{{ fenToYuan(product.price_fen) }}</text>
            <text class="stock">库存 {{ product.stock_count }}</text>
          </view>
          <view class="actions">
            <button v-if="product.status !== 'active'" class="primary-button" @click="setStatus(product.id, 'active')">上架</button>
            <button v-if="product.status === 'active'" class="ghost-button" @click="setStatus(product.id, 'sold_out')">售罄</button>
            <button v-if="product.status !== 'removed'" class="danger-button" @click="setStatus(product.id, 'removed')">下架</button>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { getMerchantProducts, saveMerchantProduct, setMerchantProductStatus } from "../../utils/api.js";

const fallbackProducts = [
  {
    id: "prod_beef_rice",
    shop_id: "shop_1",
    name: "招牌牛肉饭",
    image_url: "/assets/mock/beef-rice.jpg",
    description: "牛肉、米饭、时蔬，适合作为外卖闭环样例。",
    ingredient_list: ["牛肉", "米饭", "青菜"],
    price_fen: 2599,
    stock_count: 50,
    status: "active"
  },
  {
    id: "prod_soup",
    shop_id: "shop_1",
    name: "每日例汤",
    image_url: "/assets/mock/soup.jpg",
    description: "随餐热汤。",
    ingredient_list: ["汤底", "蔬菜"],
    price_fen: 599,
    stock_count: 80,
    status: "active"
  }
];

export default {
  data() {
    return {
      products: []
    };
  },
  onLoad() {
    this.refresh();
  },
  methods: {
    async refresh() {
      try {
        this.products = await getMerchantProducts("shop_1");
      } catch (_error) {
        this.products = fallbackProducts;
      }
    },
    async createDemo() {
      const product = {
        shop_id: "shop_1",
        name: "轻食鸡胸饭",
        image_url: "/assets/mock/chicken-rice.jpg",
        description: "鸡胸肉、糙米、蔬菜。",
        ingredient_list: ["鸡胸肉", "糙米", "蔬菜"],
        price_fen: 2299,
        stock_count: 20,
        status: "active"
      };
      try {
        await saveMerchantProduct(product);
        await this.refresh();
      } catch (_error) {
        this.products = [{ id: `local_${Date.now()}`, ...product }, ...this.products];
      }
    },
    async setStatus(productId, status) {
      try {
        await setMerchantProductStatus(productId, status);
        await this.refresh();
      } catch (_error) {
        this.products = this.products.map((item) => (item.id === productId ? { ...item, status } : item));
      }
    },
    back() {
      uni.navigateBack({ delta: 1 });
    },
    statusLabel(status) {
      const labels = {
        active: "上架",
        sold_out: "售罄",
        removed: "下架"
      };
      return labels[status] || status;
    },
    fenToYuan(value) {
      return (Number(value || 0) / 100).toFixed(2);
    }
  }
};
</script>

<style>
.page {
  min-height: 100vh;
  padding: 28rpx 28rpx 0;
  background: #f4f7fb;
}
.topbar {
  display: grid;
  grid-template-columns: 72rpx 1fr 112rpx;
  align-items: center;
  gap: 16rpx;
}
.title {
  display: block;
  color: #101828;
  font-size: 38rpx;
  font-weight: 700;
}
.subtitle {
  display: block;
  margin-top: 4rpx;
  color: #667085;
  font-size: 23rpx;
}
button {
  margin: 0;
  border-radius: 8rpx;
  font-size: 24rpx;
}
button::after {
  border: 0;
}
.icon-button {
  width: 72rpx;
  height: 72rpx;
  background: #ffffff;
  color: #101828;
  font-size: 42rpx;
  line-height: 64rpx;
}
.create-button {
  background: #009bf5;
  color: #ffffff;
  line-height: 64rpx;
}
.product-list {
  height: calc(100vh - 116rpx);
  margin-top: 24rpx;
  padding-bottom: 36rpx;
}
.product-card {
  display: grid;
  grid-template-columns: 144rpx 1fr;
  gap: 20rpx;
  margin-bottom: 16rpx;
  padding: 20rpx;
  border: 1rpx solid #e5e7eb;
  border-radius: 8rpx;
  background: #ffffff;
}
.media,
.media image {
  width: 144rpx;
  height: 144rpx;
  border-radius: 8rpx;
}
.media {
  overflow: hidden;
  background: #e5f5ff;
}
.media.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #009bf5;
  font-size: 48rpx;
  font-weight: 700;
}
.content {
  min-width: 0;
}
.head,
.meta,
.actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14rpx;
}
.name {
  min-width: 0;
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.status {
  flex-shrink: 0;
  color: #009bf5;
  font-size: 23rpx;
  font-weight: 700;
}
.status.sold_out {
  color: #b45309;
}
.status.removed {
  color: #98a2b3;
}
.desc,
.ingredients {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 23rpx;
  line-height: 34rpx;
}
.ingredients {
  color: #98a2b3;
}
.meta {
  justify-content: flex-start;
  margin-top: 12rpx;
}
.price {
  color: #101828;
  font-size: 30rpx;
  font-weight: 700;
}
.stock {
  color: #667085;
  font-size: 23rpx;
}
.actions {
  justify-content: flex-end;
  margin-top: 16rpx;
}
.primary-button,
.ghost-button,
.danger-button {
  width: 112rpx;
  line-height: 52rpx;
}
.primary-button {
  background: #009bf5;
  color: #ffffff;
}
.ghost-button {
  border: 1rpx solid #d0d5dd;
  background: #ffffff;
  color: #344054;
}
.danger-button {
  background: #fff1f3;
  color: #b42318;
}
</style>
