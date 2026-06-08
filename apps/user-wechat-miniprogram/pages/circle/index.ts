import { createCirclePost, ensurePreviewAuth, getCirclePosts } from "../../utils/api";
import { backOrSwitchTab } from "../../utils/navigation";

Page({
  data: {
    activeTab: "recommend",
    tabs: [
      { key: "recommend", title: "推荐" },
      { key: "nearby", title: "附近" },
      { key: "merchant", title: "商户群" },
      { key: "official", title: "官方" }
    ],
    posts: [
      { id: "post_1", author: "小林", initial: "小", tag: "附近 500m", title: "午饭拼单有人吗", content: "蓝海餐厅满减差一份，12:20 前下单", meta: "评论 3 · 点赞 8" },
      { id: "post_2", author: "悦享e食官方群", initial: "官", tag: "官方", title: "新用户入群默认静音", content: "重要通知会通过站内信保留", meta: "评论 5 · 点赞 23" },
      { id: "post_3", author: "蓝海餐厅商户群", initial: "店", tag: "商户群", title: "群内团购券限时领取", content: "加入商户群后可领取指定优惠券", meta: "评论 2 · 点赞 16" }
    ]
  },
  onShow() {
    this.loadPosts();
  },
  async loadPosts() {
    ensurePreviewAuth();
    try {
      const posts = await getCirclePosts() as Array<Record<string, unknown>>;
      if (posts.length > 0) {
        this.setData({
          posts: posts.map((post) => ({
            id: String(post.id || ""),
            author: String(post.author_name || "附近用户"),
            initial: String(post.author_name || "附").slice(0, 1),
            tag: String(post.distance_text || (Array.isArray(post.tags) ? post.tags[0] : "") || "推荐"),
            title: String(post.title || "圈子动态"),
            content: String(post.content || ""),
            meta: `评论 ${Number(post.comment_count || 0)} · 点赞 ${Number(post.like_count || 0)}`
          }))
        });
      }
    } catch (_error) {
      // Keep seeded preview posts.
    }
  },
  handleTabTap(event) {
    this.setData({ activeTab: String(event.currentTarget.dataset.key || "recommend") });
  },
  async handlePublishTap() {
    ensurePreviewAuth();
    try {
      await createCirclePost({
        circle_id: "nearby",
        type: "food_invite",
        title: "今天想拼单",
        content: "蓝海餐厅满减还差一份，有附近同事一起吗？",
        tags: ["附近", "拼单"]
      });
      wx.showToast({ title: "已发布", icon: "success" });
      this.loadPosts();
    } catch (_error) {
      wx.showToast({ title: "预览发布成功", icon: "none" });
    }
  },
  handleMerchantGroupTap() {
    wx.navigateTo({ url: "/pages/messages/merchant-group/index?thread_id=merchant_blue_sea" });
  },
  handleBack() {
    backOrSwitchTab("/pages/index/index");
  }
});
