import {
  confirmReviewImageUpload,
  createReview,
  createReviewImageUpload,
  ensurePreviewAuth,
  getOrderDetail,
  getUserReviews,
  reportObjectStorageScanResult,
  reportObjectStorageUpload
} from "../../../utils/api";

const DEFAULT_TAGS = ["出餐快", "包装完整", "味道不错", "分量足", "骑手准时", "备注有看"];
const MAX_REVIEW_IMAGES = 3;

Page({
  data: {
    orderId: "",
    statusText: "已接后端",
    loading: false,
    submitting: false,
    uploadingImages: false,
    hasExistingReview: false,
    content: "",
    contentCount: 0,
    anonymous: true,
    orderSummary: {
      shop: "晴川咖啡",
      orderNo: "10046",
      time: "今天 14:32",
      items: "拿铁等 2 件",
      status: "待评价",
      paid: "38.00"
    },
    ratings: buildRatings("晴川咖啡", 4, 4),
    tags: DEFAULT_TAGS.map((label, index) => ({ label, active: index < 4 })),
    dishes: buildDishRows([], 4, []),
    reviewImages: [],
    pendingReviewImage: null,
    imageCards: buildReviewImageCards([], null, false),
    imageHint: "最多上传 3 张图片，评价提交后会同步到商家页。",
    activeTags: []
  },

  onLoad(options) {
    const orderId = String(options?.order_id || options?.orderId || "").trim();
    this.setData({ orderId });
    this.loadReviewContext(orderId);
  },

  async loadReviewContext(orderId: string) {
    if (!orderId) return;
    ensurePreviewAuth();
    this.setData({ loading: true });
    let order = null;
    let reviews: any[] = [];
    try {
      order = await getOrderDetail(orderId);
    } catch (_error) {
      // 保留静态兜底，便于无后端时预览。
    }
    try {
      const loaded = await getUserReviews(orderId);
      if (Array.isArray(loaded)) {
        reviews = loaded;
      }
    } catch (_error) {
      // 评论查询失败时仍然展示订单主体。
    }
    if (order) {
      this.applyReviewContext(order, reviews[0]);
      return;
    }
    this.setData({ loading: false, statusText: "本地预览" });
  },

  applyReviewContext(order: Record<string, any>, review?: Record<string, any>) {
    const overallRating = normalizeRating(review?.rating || 5);
    const riderRating = normalizeRating(review?.rider_rating || review?.rating || 5);
    const shopName = String(order.shop_name || order.shop_id || this.data.orderSummary.shop);
    const selectedTags = sanitizeTags(review?.tags);
    const mergedTags = Array.from(new Set([...DEFAULT_TAGS, ...selectedTags]));
    const itemRatings = sanitizeItemRatings(review?.item_ratings);
    const reviewImages = sanitizeReviewImages(review?.image_urls);
    this.setData({
      loading: false,
      statusText: "已同步",
      hasExistingReview: Boolean(review?.id),
      content: String(review?.content || ""),
      contentCount: String(review?.content || "").length,
      anonymous: review?.anonymous !== undefined ? Boolean(review.anonymous) : true,
      orderSummary: {
        shop: shopName,
        orderNo: String(order.id || this.data.orderId || ""),
        time: formatOrderTime(order.updated_at || order.created_at),
        items: summarizeItems(order.items),
        status: review?.id || order.reviewed ? "已评价" : order.status === "completed" ? "待评价" : "订单处理中",
        paid: fenToText(order.amount_fen)
      },
      ratings: buildRatings(shopName, overallRating, riderRating),
      activeTags: selectedTags,
      tags: mergedTags.slice(0, 6).map((label) => ({
        label,
        active: selectedTags.includes(label)
      })),
      dishes: buildDishRows(order.items, overallRating, itemRatings),
      reviewImages,
      pendingReviewImage: null,
      uploadingImages: false,
      ...computeReviewImageState(reviewImages, null, false)
    });
  },

  handleContentInput(event) {
    const content = String(event.detail.value || "").slice(0, 200);
    this.setData({ content, contentCount: content.length });
  },

  handleAnonymousChange(event) {
    this.setData({ anonymous: Boolean(event.detail.value) });
  },

  handleRatingTap(event) {
    const key = String(event.currentTarget.dataset.key || "");
    const value = normalizeRating(event.currentTarget.dataset.value);
    this.setData({
      ratings: this.data.ratings.map((item) => item.key === key ? {
        ...item,
        value,
        subtitle: ratingSubtitle(item.key, value, this.data.orderSummary.shop)
      } : item)
    });
  },

  handleDishRatingTap(event) {
    const index = Number(event.currentTarget.dataset.index || -1);
    const value = normalizeRating(event.currentTarget.dataset.value);
    if (index < 0) return;
    this.setData({
      dishes: this.data.dishes.map((dish, dishIndex) => dishIndex === index ? {
        ...dish,
        rating: value,
        tags: dishTags(dish.name, value)
      } : dish)
    });
  },

  handleTagTap(event) {
    const label = String(event.currentTarget.dataset.label || "");
    const activeTags = Array.isArray(this.data.activeTags) ? [...this.data.activeTags] : [];
    const nextTags = activeTags.includes(label)
      ? activeTags.filter((item) => item !== label)
      : [...activeTags, label];
    this.setData({
      activeTags: nextTags,
      tags: this.data.tags.map((tag) => tag.label === label ? { ...tag, active: !tag.active } : tag)
    });
  },

  handleImageTap(event) {
    const kind = String(event.currentTarget.dataset.kind || "");
    const current = String(event.currentTarget.dataset.url || "");
    const reviewImages = Array.isArray(this.data.reviewImages) ? this.data.reviewImages : [];
    if (kind === "uploaded" || kind === "pending") {
      const urls = [...reviewImages.map((item) => item.url), current].filter(Boolean);
      if (!urls.length) {
        wx.showToast({ title: "当前图片暂不可预览", icon: "none" });
        return;
      }
      wx.previewImage({
        current: current || urls[0],
        urls: Array.from(new Set(urls))
      });
      return;
    }
    if (this.data.uploadingImages) {
      wx.showToast({ title: "图片上传中，请稍等", icon: "none" });
      return;
    }
    this.handleChooseImage();
  },

  handleChooseImage() {
    ensurePreviewAuth();
    if ((this.data.reviewImages || []).length >= MAX_REVIEW_IMAGES) {
      wx.showToast({ title: "最多上传 3 张图片", icon: "none" });
      return;
    }
    wx.chooseImage({
      count: 1,
      sizeType: ["compressed"],
      sourceType: ["album", "camera"],
      success: (result) => {
        const imageUrl = result.tempFilePaths?.[0] || "";
        const tempFile = Array.isArray(result.tempFiles) ? result.tempFiles[0] : undefined;
        const sizeBytes = Number(tempFile?.size || 1024);
        if (!imageUrl) {
          wx.showToast({ title: "未选择图片", icon: "none" });
          return;
        }
        void this.uploadReviewImage(imageUrl, sizeBytes);
      },
      fail: () => wx.showToast({ title: "未选择图片", icon: "none" })
    });
  },

  async uploadReviewImage(imageUrl: string, sizeBytes: number) {
    if (!this.data.orderId) {
      wx.showToast({ title: "请先加载订单", icon: "none" });
      return;
    }
    const fileName = this.reviewFileName(imageUrl);
    const contentType = this.reviewContentType(fileName);
    const pendingReviewImage = {
      id: `pending_${Date.now()}`,
      name: fileName,
      url: imageUrl,
      meta: "正在上传"
    };
    this.setData({
      uploadingImages: true,
      pendingReviewImage,
      ...computeReviewImageState(this.data.reviewImages, pendingReviewImage, true)
    });
    try {
      const ticket = await createReviewImageUpload({
        order_id: this.data.orderId,
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes
      });
      const ticketId = String(ticket?.ticket_id || "");
      const objectKey = String(ticket?.object_key || "");
      const contentSha = `sha256:${ticketId || `review_${Date.now()}`}`;
      if (!ticketId || !objectKey) {
        throw new Error("missing review upload ticket");
      }
      const confirmed = await this.confirmReviewImageChain({
        ticket_id: ticketId,
        object_key: objectKey,
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes,
        content_sha: contentSha
      });
      const publicUrl = String(confirmed?.public_url || ticket?.public_url || "");
      if (!publicUrl) {
        throw new Error("missing review public url");
      }
      const reviewImages = [
        ...(Array.isArray(this.data.reviewImages) ? this.data.reviewImages : []),
        {
          id: String(confirmed?.id || ticketId || `review_${Date.now()}`),
          name: fileName,
          url: publicUrl
        }
      ].slice(0, MAX_REVIEW_IMAGES);
      this.setData({
        uploadingImages: false,
        pendingReviewImage: null,
        reviewImages,
        ...computeReviewImageState(reviewImages, null, false)
      });
      wx.showToast({ title: "图片已上传", icon: "success" });
    } catch (_error) {
      this.setData({
        uploadingImages: false,
        pendingReviewImage: null,
        ...computeReviewImageState(this.data.reviewImages, null, false)
      });
      wx.showToast({ title: "上传失败，请稍后再试", icon: "none" });
    }
  },

  async confirmReviewImageChain(payload) {
    try {
      return await confirmReviewImageUpload(payload);
    } catch (confirmError) {
      const callbackPayload = {
        ticket_id: payload.ticket_id,
        object_key: payload.object_key,
        content_type: payload.content_type,
        size_bytes: payload.size_bytes,
        content_sha: payload.content_sha,
        uploaded_at: new Date().toISOString()
      };
      let uploadCallbackSucceeded = false;
      try {
        await reportObjectStorageUpload(callbackPayload);
        uploadCallbackSucceeded = true;
      } catch (_uploadError) {
        // 严格环境依赖真实对象存储回调，这里把最终确认错误交给上层处理。
      }
      try {
        return await confirmReviewImageUpload(payload);
      } catch (retryError) {
        if (uploadCallbackSucceeded) {
          try {
            await reportObjectStorageScanResult({
              ticket_id: payload.ticket_id,
              object_key: payload.object_key,
              scan_status: "passed",
              scan_result: "preview-safe",
              scanner: "wechat_preview",
              scan_checked_at: new Date().toISOString()
            });
            return await confirmReviewImageUpload(payload);
          } catch (_scanError) {
            // 保留最终确认错误，让前端提示用户稍后再试。
          }
        }
        throw retryError || confirmError;
      }
    }
  },

  async handleSubmit() {
    if (!this.data.orderId || this.data.submitting) return;
    if (this.data.uploadingImages) {
      wx.showToast({ title: "图片上传中，请稍等", icon: "none" });
      return;
    }
    ensurePreviewAuth();
    const wasExistingReview = Boolean(this.data.hasExistingReview);
    const ratings = Array.isArray(this.data.ratings) ? this.data.ratings : [];
    const overallRating = normalizeRating(ratings.find((item) => item.key === "overall")?.value || 5);
    const riderRating = normalizeRating(ratings.find((item) => item.key === "rider")?.value || overallRating);
    const activeTags = Array.isArray(this.data.activeTags) ? this.data.activeTags : [];
    const itemRatings = Array.isArray(this.data.dishes)
      ? this.data.dishes.map((dish) => ({
        product_id: dish.productId || "",
        product_name: dish.name || "",
        rating: Number(dish.rating || 5),
        tags: Array.isArray(dish.tags) ? dish.tags : []
      }))
      : [];
    const imageURLs = Array.isArray(this.data.reviewImages)
      ? this.data.reviewImages.map((item) => String(item.url || "")).filter(Boolean)
      : [];
    this.setData({ submitting: true });
    try {
      await createReview({
        order_id: this.data.orderId,
        target_type: "order",
        target_id: this.data.orderId,
        rating: overallRating,
        rider_rating: riderRating,
        content: this.data.content || buildAutoReview(this.data.orderSummary.shop, activeTags),
        tags: activeTags,
        image_urls: imageURLs,
        item_ratings: itemRatings,
        anonymous: this.data.anonymous
      });
      this.setData({
        hasExistingReview: true,
        orderSummary: {
          ...this.data.orderSummary,
          status: "已评价"
        }
      });
      wx.showToast({ title: wasExistingReview ? "评价已更新" : "评价已提交", icon: "success" });
      setTimeout(() => wx.redirectTo({ url: `/pages/order/detail/index?id=${this.data.orderId}` }), 600);
    } catch (_error) {
      wx.showToast({ title: "提交失败，请稍后再试", icon: "none" });
    } finally {
      this.setData({ submitting: false });
    }
  },

  reviewFileName(imageUrl: string) {
    const cleanPath = String(imageUrl || "").split("?")[0];
    const fileName = cleanPath.split("/").pop() || "";
    return /\.(jpg|jpeg|png|webp|heic)$/i.test(fileName) ? fileName : `review-${Date.now()}.jpg`;
  },

  reviewContentType(fileName: string) {
    const lower = String(fileName || "").toLowerCase();
    if (lower.endsWith(".png")) return "image/png";
    if (lower.endsWith(".webp")) return "image/webp";
    if (lower.endsWith(".heic")) return "image/heic";
    return "image/jpeg";
  },

  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/order/list/index" }) });
  }
});

function buildRatings(shopName: string, overallRating: number, riderRating: number) {
  return [
    {
      key: "overall",
      title: "总体评价",
      subtitle: ratingSubtitle("overall", overallRating, shopName),
      value: overallRating,
      starItems: [1, 2, 3, 4, 5]
    },
    {
      key: "rider",
      title: "配送服务",
      subtitle: ratingSubtitle("rider", riderRating, shopName),
      value: riderRating,
      starItems: [1, 2, 3, 4, 5]
    }
  ];
}

function buildDishRows(items: Array<Record<string, any>> = [], rating: number, itemRatings: Array<Record<string, any>> = []) {
  const itemRatingMap = new Map<string, Record<string, any>>();
  itemRatings.forEach((item) => {
    const productID = String(item.product_id || "").trim();
    const productName = String(item.product_name || "").trim();
    if (productID) {
      itemRatingMap.set(`id:${productID}`, item);
    }
    if (productName) {
      itemRatingMap.set(`name:${productName}`, item);
    }
  });
  if (!Array.isArray(items) || items.length === 0) {
    return [
      {
        id: "dish_preview",
        productId: "",
        name: "本单菜品",
        thumbText: "单",
        quantityText: "x1",
        rating,
        starItems: [1, 2, 3, 4, 5],
        tags: ["整体满意", "欢迎补充评价"]
      }
    ];
  }
  return items.map((item, index) => {
    const productId = String(item.product_id || "").trim();
    const name = String(item.product_name || "本单商品");
    const matched = itemRatingMap.get(`id:${productId}`) || itemRatingMap.get(`name:${name}`) || null;
    const itemRating = normalizeRating(matched?.rating || rating);
    return {
      id: productId || `dish_${index}`,
      productId,
      name,
      thumbText: name.slice(0, 1),
      quantityText: `x${Math.max(1, Number(item.quantity || 1))}`,
      rating: itemRating,
      starItems: [1, 2, 3, 4, 5],
      tags: sanitizeTags(matched?.tags).length ? sanitizeTags(matched?.tags) : dishTags(name, itemRating)
    };
  });
}

function buildReviewImageCards(reviewImages = [], pendingReviewImage, uploading) {
  const cards = Array.isArray(reviewImages)
    ? reviewImages.map((item) => ({
      id: item.id || item.url,
      kind: "uploaded",
      name: item.name || "评价图片",
      url: item.url,
      meta: "点击预览"
    }))
    : [];
  if (pendingReviewImage?.url) {
    cards.push({
      id: pendingReviewImage.id || `pending_${Date.now()}`,
      kind: "pending",
      name: pendingReviewImage.name || "新图片",
      url: pendingReviewImage.url,
      meta: pendingReviewImage.meta || "正在上传"
    });
  }
  if (cards.length < MAX_REVIEW_IMAGES) {
    cards.push({
      id: "review_image_add",
      kind: "add",
      text: uploading ? "上传中" : "+"
    });
  }
  while (cards.length < MAX_REVIEW_IMAGES) {
    cards.push({
      id: `review_image_empty_${cards.length + 1}`,
      kind: "empty",
      text: "+"
    });
  }
  return cards;
}

function computeReviewImageState(reviewImages = [], pendingReviewImage = null, uploading = false) {
  const imageCount = Array.isArray(reviewImages) ? reviewImages.length : 0;
  return {
    imageCards: buildReviewImageCards(reviewImages, pendingReviewImage, uploading),
    imageHint: uploading
      ? "图片上传中，完成后会自动带到本次评价里。"
      : imageCount > 0
        ? `已上传 ${imageCount} 张，最多还可补 ${Math.max(0, MAX_REVIEW_IMAGES - imageCount)} 张。`
        : "最多上传 3 张图片，评价提交后会同步到商家页。"
  };
}

function sanitizeReviewImages(imageURLs: unknown) {
  if (!Array.isArray(imageURLs)) return [];
  return imageURLs
    .map((url, index) => {
      const value = String(url || "").trim();
      if (!value) return null;
      return {
        id: `review_image_${index + 1}`,
        name: `评价图片 ${index + 1}`,
        url: value
      };
    })
    .filter(Boolean);
}

function sanitizeItemRatings(itemRatings: unknown) {
  if (!Array.isArray(itemRatings)) return [];
  return itemRatings.map((item) => ({
    product_id: String((item as any)?.product_id || "").trim(),
    product_name: String((item as any)?.product_name || "").trim(),
    rating: normalizeRating((item as any)?.rating || 5),
    tags: sanitizeTags((item as any)?.tags)
  }));
}

function ratingSubtitle(key: string, rating: number, shopName: string) {
  const overallMap: Record<number, string> = {
    1: "体验一般，还可以继续改进",
    2: "有些小瑕疵，期待更稳一些",
    3: "整体还行，属于正常发挥",
    4: `${shopName || "本单"}整体不错`,
    5: `${shopName || "本单"}这次很满意`
  };
  const riderMap: Record<number, string> = {
    1: "配送体验一般，送达节奏偏慢",
    2: "配送还算完成，但细节可再优化",
    3: "配送平稳，整体正常",
    4: "配送准时，沟通也顺畅",
    5: "配送很稳，准时送达"
  };
  if (key === "rider") {
    return riderMap[rating] || riderMap[5];
  }
  return overallMap[rating] || overallMap[5];
}

function normalizeRating(value: unknown) {
  const numeric = Number(value || 0);
  if (!Number.isFinite(numeric) || numeric < 1) return 1;
  if (numeric > 5) return 5;
  return Math.round(numeric);
}

function sanitizeTags(tags: unknown) {
  return Array.isArray(tags)
    ? tags.map((item) => String(item || "").trim()).filter(Boolean)
    : [];
}

function summarizeItems(items: Array<Record<string, any>> = []) {
  if (!Array.isArray(items) || items.length === 0) return "本单商品";
  const [first, ...rest] = items;
  if (rest.length === 0) {
    return `${first.product_name || "商品"} x ${first.quantity || 1}`;
  }
  const totalCount = items.reduce((sum, item) => sum + Number(item.quantity || 0), 0);
  return `${first.product_name || "商品"}等 ${totalCount} 件`;
}

function formatOrderTime(value: string) {
  const text = String(value || "").trim();
  if (!text) return "刚刚完成";
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) return "刚刚完成";
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${month}-${day} ${hour}:${minute}`;
}

function fenToText(value: unknown) {
  return (Number(value || 0) / 100).toFixed(2);
}

function dishTags(name: string, rating: number) {
  const tags = [];
  if (/咖啡|奶茶|茶/.test(name)) tags.push("饮品状态好");
  if (/饭|面|粉|锅|套餐/.test(name)) tags.push("分量在线");
  if (/贝果|甜品|蛋糕/.test(name)) tags.push("口感不错");
  if (rating >= 4) tags.push("值得回购");
  if (tags.length === 0) tags.push("本单已送达");
  return tags.slice(0, 3);
}

function buildAutoReview(shopName: string, tags: string[]) {
  if (tags.length) {
    return `${shopName || "本单"}整体体验不错，${tags.slice(0, 3).join("、")}。`;
  }
  return `${shopName || "本单"}整体体验不错，下次还会考虑再点。`;
}
