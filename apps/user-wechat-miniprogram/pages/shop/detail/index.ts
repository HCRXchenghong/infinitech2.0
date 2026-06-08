import { createGroupbuyOrder, ensurePreviewAuth, getChatThreadMembership, getShopDetail, getShopGroupbuyDeals, getShopProducts, joinChatThread, setDevAuthToken, upsertCartItem } from "../../../utils/api";
import { generatedImages, mediaUrl, productFallbackImage } from "../../../utils/media";

Page({
  data: {
    shopId: "shop_1",
    shop: {
      name: "蓝海餐厅",
      rating: "4.8",
      sales: "月售 2381",
      delivery: "约 32 分钟送达",
      announcement: "招牌牛肉饭热卖，到店套餐扫码验券。",
      qualification: "资质已审核",
      coverUrl: generatedImages.shopDetailCover,
      logoUrl: "/assets/brand/logo.jpg"
    },
    tabs: ["点餐", "团购券", "评价", "商家"],
    activeTab: "点餐",
    categories: ["热销", "主食", "饮品", "小食"],
    activeCategory: "热销",
    activities: ["满 45 减 8", "新人立减 5", "团购可用"],
    merchantGroup: {
      threadId: "merchant_blue_sea",
      summary: "326 人已加入 · 新用户默认静音",
      notice: "进群可领外卖专享券，优惠和拼单通知都会同步到群内。",
      joined: false,
      canJoin: true,
      buttonText: "加入商户群领券"
    },
    reviewSummary: {
      averageRating: "4.8",
      reviewCount: 3,
      positiveRate: "97% 好评",
      highlightTags: ["出餐快", "分量足", "服务细致", "包装完整"]
    },
    reviews: [
      {
        reviewId: "shop_1_seed_review_1",
        userName: "林同学",
        avatarText: "林",
        starsText: "★★★★★",
        content: "牛肉饭分量很稳，晚高峰也没有撒汤，出餐比想象中快。",
        imageURLs: [],
        imageTiles: [],
        imageOverflowCount: 0,
        itemHighlights: ["招牌牛肉饭 5分", "柠檬茶 4分"],
        riderStarsText: "★★★★★",
        tags: ["出餐快", "包装完整", "分量足"],
        replyText: "谢谢支持，晚饭时段也会尽量保证出餐速度。",
        createdText: "今天 18:20"
      },
      {
        reviewId: "shop_1_seed_review_2",
        userName: "阿杰",
        avatarText: "阿",
        starsText: "★★★★☆",
        content: "柠檬茶清爽，套餐券到店核销也挺顺，适合中午拼单。",
        imageURLs: [],
        imageTiles: [],
        imageOverflowCount: 0,
        itemHighlights: ["双人套餐 4分"],
        riderStarsText: "★★★★☆",
        tags: ["适合拼单", "饮品不错", "团购方便"],
        replyText: "团购券工作日和周末都能用，欢迎常来。",
        createdText: "昨天 12:08"
      }
    ],
    merchantInfo: {
      merchantName: "蓝海餐厅",
      qualificationText: "资质已审核",
      businessHours: "09:30-22:30",
      contactPhone: "13800000001",
      address: "大学城生活区 2 号门西侧 18 米",
      serviceCommitments: ["后厨明档公示", "准时出餐提醒", "支持到店自取"],
      qualificationItems: ["营业执照已公示", "健康证在有效期内", "平台保证金已缴纳"],
      supportBulletins: ["遇到缺货会先电话确认再处理订单。", "团购券支持到店扫码验券，不支持与到店红包叠加。"]
    },
    products: [
      {
        id: "prod_beef_rice",
        name: "招牌牛肉饭",
        description: "牛肉、米饭、时蔬，配料表可由商户后台维护。",
        imageUrl: generatedImages.productBeefRice,
        price: "25.99",
        sales: "月售 928",
        category: "热销",
        thumb: "饭",
        count: 1
      },
      {
        id: "prod_tea",
        name: "柠檬茶",
        description: "清爽柠檬茶，适合随餐搭配。",
        imageUrl: generatedImages.productLemonTea,
        price: "9.00",
        sales: "月售 420",
        category: "饮品",
        thumb: "茶",
        count: 0
      },
      {
        id: "prod_wings",
        name: "烤鸡翅",
        description: "外焦里嫩，建议趁热食用。",
        imageUrl: generatedImages.productChickenWings,
        price: "18.00",
        sales: "月售 316",
        category: "小食",
        thumb: "翅",
        count: 0
      },
      {
        id: "prod_noodle",
        name: "番茄鸡蛋面",
        description: "番茄汤底，鸡蛋和青菜。",
        imageUrl: generatedImages.productTomatoNoodle,
        price: "16.00",
        sales: "月售 268",
        category: "主食",
        thumb: "面",
        count: 0
      }
    ],
    groupbuyDeals: [
      {
        id: "deal_two_person_set",
        title: "双人到店套餐券",
        price: "39.99",
        note: "购买后生成二维码，商户端扫码验券。"
      }
    ],
    cartSelectedCount: 1,
    cartSubtitle: "商品 ¥25.99 · 配送 ¥3 · 打包 ¥1"
  },
  onLoad(query) {
    const shopId = String(query?.id || "shop_1");
    this.setData({ shopId });
    this.loadShopDetail(shopId);
    this.loadMerchantGroupMembership();
    this.loadProducts(shopId);
    this.loadGroupbuyDeals(shopId);
  },
  onShow() {
    this.loadShopDetail(String(this.data.shopId || "shop_1"));
    this.loadMerchantGroupMembership();
  },
  async loadShopDetail(shopId) {
    try {
      const detail = await getShopDetail(shopId) as Record<string, any>;
      const reviewSummary = (detail.review_summary || {}) as Record<string, any>;
      const merchantInfo = (detail.merchant_info || {}) as Record<string, any>;
      const coverUrl = String(detail.cover_url || "").trim();
      const logoUrl = String(detail.logo_url || "").trim();
      this.setData({
        shop: {
          ...this.data.shop,
          name: String(detail.name || this.data.shop.name),
          rating: String(detail.rating_text || this.data.shop.rating),
          sales: String(detail.sales_text || this.data.shop.sales),
          delivery: String(detail.delivery_text || this.data.shop.delivery),
          announcement: String(detail.announcement || this.data.shop.announcement),
          qualification: String(detail.qualification_text || this.data.shop.qualification),
          coverUrl: mediaUrl(coverUrl, this.data.shop.coverUrl),
          logoUrl: logoUrl && !/\.svg$/i.test(logoUrl) ? logoUrl : this.data.shop.logoUrl
        },
        activities: Array.isArray(detail.activity_tags) && detail.activity_tags.length
          ? detail.activity_tags.map((tag) => String(tag))
          : this.data.activities,
        reviewSummary: {
          averageRating: String(reviewSummary.average_rating || this.data.reviewSummary.averageRating),
          reviewCount: Number(reviewSummary.review_count || this.data.reviewSummary.reviewCount || 0),
          positiveRate: String(reviewSummary.positive_rate || this.data.reviewSummary.positiveRate),
          highlightTags: Array.isArray(reviewSummary.highlight_tags) && reviewSummary.highlight_tags.length
            ? reviewSummary.highlight_tags.map((tag) => String(tag))
            : this.data.reviewSummary.highlightTags
        },
        reviews: Array.isArray(detail.reviews) && detail.reviews.length
          ? detail.reviews.map((review) => {
              const imageURLs = Array.isArray(review.image_urls) ? review.image_urls.map((url) => String(url)).filter(Boolean) : [];
              const imageTiles = buildReviewImageTiles(imageURLs);
              return {
                reviewId: String(review.review_id || ""),
                userName: String(review.user_name || "匿名用户"),
                avatarText: String(review.avatar_text || "匿"),
                starsText: String(review.stars_text || "★★★★★"),
                content: String(review.content || ""),
                imageURLs,
                imageTiles,
                imageOverflowCount: Math.max(0, imageURLs.length - imageTiles.length),
                itemHighlights: Array.isArray(review.item_highlights) ? review.item_highlights.map((tag) => String(tag)).filter(Boolean) : [],
                riderStarsText: String(review.rider_stars_text || ""),
                tags: Array.isArray(review.tags) ? review.tags.map((tag) => String(tag)) : [],
                replyText: String(review.reply_text || ""),
                createdText: String(review.created_text || "")
              };
            })
          : this.data.reviews,
        merchantInfo: {
          merchantName: String(merchantInfo.merchant_name || this.data.merchantInfo.merchantName),
          qualificationText: String(merchantInfo.qualification_text || detail.qualification_text || this.data.merchantInfo.qualificationText),
          businessHours: String(merchantInfo.business_hours || this.data.merchantInfo.businessHours),
          contactPhone: String(merchantInfo.contact_phone || this.data.merchantInfo.contactPhone),
          address: String(merchantInfo.address || this.data.merchantInfo.address),
          serviceCommitments: Array.isArray(merchantInfo.service_commitments) && merchantInfo.service_commitments.length
            ? merchantInfo.service_commitments.map((item) => String(item))
            : this.data.merchantInfo.serviceCommitments,
          qualificationItems: Array.isArray(merchantInfo.qualification_items) && merchantInfo.qualification_items.length
            ? merchantInfo.qualification_items.map((item) => String(item))
            : this.data.merchantInfo.qualificationItems,
          supportBulletins: Array.isArray(merchantInfo.support_bulletins) && merchantInfo.support_bulletins.length
            ? merchantInfo.support_bulletins.map((item) => String(item))
            : this.data.merchantInfo.supportBulletins
        }
      });
    } catch (_error) {
      // 保留页面兜底数据，便于无后端时继续预览参考图。
    }
  },
  async loadMerchantGroupMembership() {
    ensurePreviewAuth();
    try {
      const membership = await getChatThreadMembership("merchant_blue_sea") as Record<string, unknown>;
      const joined = Boolean(membership.joined);
      this.setData({
        merchantGroup: {
          ...this.data.merchantGroup,
          summary: String(membership.summary || this.data.merchantGroup.summary),
          joined,
          canJoin: Boolean(membership.can_join),
          buttonText: joined ? "去群里看看" : "加入商户群领券"
        }
      });
    } catch (_error) {
      // Keep seeded preview copy.
    }
  },
  async loadProducts(shopId) {
    try {
      const products = await getShopProducts(shopId);
      if (Array.isArray(products) && products.length > 0) {
        this.setData({
          products: products.map((product) => ({
            id: String(product.id || ""),
            name: product.name,
            description: product.description,
            imageUrl: mediaUrl(product.image_url, productFallbackImage(String(product.id || ""))),
            price: (Number(product.price_fen || 0) / 100).toFixed(2),
            sales: "月售 928",
            category: "热销",
            thumb: "饭",
            count: 0
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  async loadGroupbuyDeals(shopId) {
    try {
      const deals = await getShopGroupbuyDeals(shopId);
      if (Array.isArray(deals) && deals.length > 0) {
        this.setData({
          groupbuyDeals: deals.map((deal) => ({
            id: deal.id,
            title: deal.name,
            note: deal.description,
            price: (Number(deal.price_fen || 0) / 100).toFixed(2)
          }))
        });
      }
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
  },
  handleTabTap(event) {
    this.setData({ activeTab: String(event.currentTarget.dataset.tab || "点餐") });
  },
  handleCategoryTap(event) {
    this.setData({ activeCategory: String(event.currentTarget.dataset.category || "热销") });
  },
  handleCallMerchant() {
    const phoneNumber = String(this.data.merchantInfo.contactPhone || "").trim();
    if (!phoneNumber) {
      wx.showToast({ title: "暂无联系电话", icon: "none" });
      return;
    }
    wx.makePhoneCall({
      phoneNumber,
      fail: () => {
        wx.showToast({ title: "拨号失败", icon: "none" });
      }
    });
  },
  handleCopyAddress() {
    const address = String(this.data.merchantInfo.address || "").trim();
    if (!address) {
      wx.showToast({ title: "暂无地址信息", icon: "none" });
      return;
    }
    wx.setClipboardData({
      data: address,
      success: () => {
        wx.showToast({ title: "地址已复制", icon: "success" });
      }
    });
  },
  handleReviewImagePreview(event) {
    const reviewIndex = Number(event.currentTarget.dataset.reviewIndex || -1);
    const imageIndex = Number(event.currentTarget.dataset.imageIndex || 0);
    const review = Array.isArray(this.data.reviews) ? this.data.reviews[reviewIndex] : null;
    const urls = Array.isArray(review?.imageURLs) ? review.imageURLs.filter((url) => Boolean(url)) : [];
    if (!urls.length) {
      wx.showToast({ title: "暂无可预览图片", icon: "none" });
      return;
    }
    wx.previewImage({
      current: urls[imageIndex] || urls[0],
      urls
    });
  },
  async handleAddTap(event) {
    const productId = String(event.currentTarget.dataset.id || "");
    if (!productId) {
      return;
    }
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      await upsertCartItem(String(this.data.shopId), productId, 1);
      const product = this.data.products.find((item) => item.id === productId);
      this.setData({
        cartSelectedCount: Number(this.data.cartSelectedCount || 0) + 1,
        cartSubtitle: product ? `商品 ¥${product.price} · 配送 ¥3 · 打包 ¥1` : this.data.cartSubtitle
      });
      wx.showToast({ title: "已加入购物车", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "已加入预览购物车", icon: "none" });
    }
  },
  async handleGroupbuyTap(event) {
    const dealId = String(event.currentTarget.dataset.id || "");
    if (!dealId) {
      return;
    }
    if (!wx.getStorageSync("authToken")) {
      setDevAuthToken("user_1");
    }
    try {
      const order = await createGroupbuyOrder(String(this.data.shopId), dealId, 1);
      wx.navigateTo({ url: `/pages/order/detail/index?id=${order.id}` });
    } catch (_error) {
      wx.showToast({ title: "团购券已加入预览订单", icon: "none" });
    }
  },
  async handleMerchantGroupAction() {
    ensurePreviewAuth();
    if (this.data.merchantGroup.joined) {
      wx.navigateTo({ url: `/pages/messages/merchant-group/index?thread_id=${encodeURIComponent(this.data.merchantGroup.threadId)}` });
      return;
    }
    try {
      const membership = await joinChatThread(this.data.merchantGroup.threadId) as Record<string, unknown>;
      this.setData({
        merchantGroup: {
          ...this.data.merchantGroup,
          joined: Boolean(membership.joined),
          canJoin: Boolean(membership.can_join),
          summary: String(membership.summary || this.data.merchantGroup.summary),
          buttonText: "去群里看看"
        }
      });
      wx.showToast({ title: "已加入商户群", icon: "success" });
      setTimeout(() => {
        wx.navigateTo({ url: `/pages/messages/merchant-group/index?thread_id=${encodeURIComponent(this.data.merchantGroup.threadId)}` });
      }, 320);
    } catch (_error) {
      wx.showToast({ title: "加入失败", icon: "none" });
    }
  },
  handleCartTap() {
    wx.navigateTo({ url: "/pages/cart/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/shop/list/index" }) });
  }
});

function buildReviewImageTiles(imageURLs = []) {
  return (Array.isArray(imageURLs) ? imageURLs : [])
    .filter((url) => Boolean(url))
    .slice(0, 3)
    .map((url, index) => ({
      id: `${index}_${url}`,
      url,
      imageIndex: index
    }));
}
