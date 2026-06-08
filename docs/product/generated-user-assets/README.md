# 用户端生图资产留痕

日期：2026-06-04

本目录保存本轮用内置生图能力生成的用户端视觉资产高清 PNG。小程序运行包内另有压缩 JPG 兜底版本，位置为 `apps/user-wechat-miniprogram/assets/generated/`。

## 接入原则

- 页面不直接依赖参考图裁片，所有业务图片先按页面需要生图。
- 小程序页面统一消费 `imageUrl`、`image_url`、`cover_url`、`logo_url` 这类链接字段。
- 后台或 BFF/API 返回图片链接时优先使用后端值；没有配置时才使用小程序内压缩兜底图。
- 后续上线 CDN 时，只需要把后台配置值改成 `https://cdn...`，前端页面结构不用再改。
- 高清 PNG 用于设计留痕和二次压缩，不放入小程序主包。

## 本轮资产

| 文件 | 用途 | 小程序兜底 |
| --- | --- | --- |
| `home-hero.png` | 首页品牌主视觉 | `/assets/generated/home-hero.jpg` |
| `home-featured-dish.png` | 首页精选团购横幅 | `/assets/generated/home-featured-dish.jpg` |
| `home-recommend-restaurant.png` | 首页推荐餐厅 / 蓝海餐厅缩略图 | `/assets/generated/home-recommend-restaurant.jpg` |
| `home-recommend-courier.png` | 首页附近跑腿缩略图 | `/assets/generated/home-recommend-courier.jpg` |
| `home-meal-match.png` | 首页找饭搭缩略图 | `/assets/generated/home-meal-match.jpg` |
| `shop-coffee.png` | 商家列表咖啡店缩略图 | `/assets/generated/shop-coffee.jpg` |
| `shop-pharmacy.png` | 商家列表药房缩略图 | `/assets/generated/shop-pharmacy.jpg` |
| `shop-home-cooking.png` | 商家列表家常菜缩略图 | `/assets/generated/shop-home-cooking.jpg` |
| `shop-claypot.png` | 商家列表砂锅缩略图 | `/assets/generated/shop-claypot.jpg` |
| `shop-detail-cover.png` | 蓝海餐厅店铺详情头图 | `/assets/generated/shop-detail-cover.jpg` |
| `product-beef-rice.png` | 招牌牛肉饭商品图 | `/assets/generated/product-beef-rice.jpg` |
| `product-lemon-tea.png` | 柠檬茶商品图 | `/assets/generated/product-lemon-tea.jpg` |
| `product-chicken-wings.png` | 烤鸡翅商品图 | `/assets/generated/product-chicken-wings.jpg` |
| `product-tomato-noodle.png` | 番茄鸡蛋面商品图 | `/assets/generated/product-tomato-noodle.jpg` |
| `medicine-clinic-cover.png` | 买药首页校医务室封面 | `/assets/generated/medicine-clinic-cover.jpg` |
| `medicine-cooling-patch.png` | 退热贴药品图 | `/assets/generated/medicine-cooling-patch.jpg` |
| `medicine-capsules.png` | 处方胶囊药品图 | `/assets/generated/medicine-capsules.jpg` |
| `medicine-first-aid.png` | 碘伏棉签 / 创可贴急救用品图 | `/assets/generated/medicine-first-aid.jpg` |
| `errand-hero.png` | 快递跑腿首页封面 | `/assets/generated/errand-hero.jpg` |
| `errand-parcel.png` | 跑腿订单包裹缩略图 | `/assets/generated/errand-parcel.jpg` |
| `category-takeout-icon.png` | 首页分类图标：外卖 | `/assets/generated/category-takeout.png` |
| `category-groupbuy-icon.png` | 首页分类图标：团购 | `/assets/generated/category-groupbuy.png` |
| `category-medicine-icon.png` | 首页分类图标：买药 | `/assets/generated/category-medicine.png` |
| `category-courier-icon.png` | 首页分类图标：快递跑腿 | `/assets/generated/category-courier.png` |
| `category-circle-icon.png` | 首页分类图标：圈子 | `/assets/generated/category-circle.png` |
| `category-meal-match-icon.png` | 首页分类图标：找饭搭 | `/assets/generated/category-meal-match.png` |
| `category-coupons-icon.png` | 首页分类图标：红包优惠 | `/assets/generated/category-coupons.png` |
| `category-points-icon.png` | 首页分类图标：会员积分 | `/assets/generated/category-points.png` |
| `category-takeout-icon-v2.png` | 首页分类图标统一重制版：外卖 | `/assets/generated/category-takeout.png` |
| `category-groupbuy-icon-v2.png` | 首页分类图标统一重制版：团购 | `/assets/generated/category-groupbuy.png` |
| `category-medicine-icon-v2.png` | 首页分类图标统一重制版：买药 | `/assets/generated/category-medicine.png` |
| `category-courier-icon-v2.png` | 首页分类图标统一重制版：快递跑腿 | `/assets/generated/category-courier.png` |
| `category-circle-icon-v2.png` | 首页分类图标统一重制版：圈子 | `/assets/generated/category-circle.png` |
| `category-meal-match-icon-v2.png` | 首页分类图标统一重制版：找饭搭 | `/assets/generated/category-meal-match.png` |
| `category-coupons-icon-v2.png` | 首页分类图标统一重制版：红包优惠 | `/assets/generated/category-coupons.png` |
| `category-points-icon-v2.png` | 首页分类图标统一重制版：会员积分 | `/assets/generated/category-points.png` |
| `category-icons-v2-contact-sheet.png` | 首页分类图标统一重制版拼图预览 | 未直接接入 |
| `homepage-redesign-preview-2026-06-05.png` | 首页新版效果图历史稿 | 未直接接入 |
| `homepage-redesign-preview-single-carousel-2026-06-05.png` | 首页新版效果图确认稿：已按此方向落地首页代码，仅保留上方主轮播 | 未直接接入 |

## 已接字段

- 首页：`heroImageUrl`、`modules[].iconImageUrl`、`featuredDeal.imageUrl`、`recommendCards[].imageUrl`、`mealMatchCard.imageUrl`、`homeCards[].imageUrl`
- 商家列表：`shops[].imageUrl`，后端优先读取 `cover_url`、`logo_url`、`image_url`
- 店铺详情：`shop.coverUrl`、`products[].imageUrl`，后端优先读取 `cover_url`、`logo_url`、`image_url`
- 买药首页/确认/详情：`clinic.coverUrl`、`products[].imageUrl`、`items[].imageUrl`，后端优先读取 `cover_url`、`image_url`
- 跑腿首页/详情：`heroImageUrl`、`itemImageUrl`、`imageUrl`，后端优先读取 `image_url`
- BFF 默认首页卡片：`GET /api/home/cards` 返回 `image_url`
- API-Go 默认首页卡片、店铺、商品：`HomeCard.ImageURL`、`Shop.CoverURL`、`MerchantProduct.ImageURL`
