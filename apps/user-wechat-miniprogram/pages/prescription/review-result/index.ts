import { ensurePreviewAuth, getPrescriptionReview } from "../../../utils/api";

Page({
  data: {
    reviewId: "rx_preview",
    status: "校医审核通过",
    statusText: "处方信息已确认，可加入购物车购买",
    reviewedAt: "14:28",
    product: {
      name: "阿莫西林胶囊",
      price: "18.80",
      quantity: 1,
      doctor: "王医生"
    },
    patientName: "张三",
    ocr: {
      status: "已识别",
      confidence: 96,
      dosage: "0.5g 每日 3 次，遵医嘱"
    },
    archive: {
      id: "rxa_preview",
      retention: "校内处方留档 6 年"
    },
    steps: [
      { title: "处方上传", subtitle: "照片清晰度与格式校验", status: "done" },
      { title: "校医审核", subtitle: "药品、剂量、有效期检查", status: "done" },
      { title: "订单履约", subtitle: "审核通过后进入药房备货", status: "active" }
    ]
  },
  onLoad(query) {
    const reviewId = String(query?.id || "rx_preview");
    this.setData({ reviewId });
    this.loadReview(reviewId);
  },
  async loadReview(reviewId) {
    ensurePreviewAuth();
    try {
      const review = await getPrescriptionReview(reviewId) as Record<string, unknown>;
      const statusValue = String(review.status || "");
      const approved = statusValue === "approved";
      const rejected = statusValue === "rejected";
      const ocr = (review.ocr_result || {}) as Record<string, unknown>;
      const archive = (review.archive || {}) as Record<string, unknown>;
      this.setData({
        status: approved ? "药师审核通过" : (rejected ? "处方审核未通过" : "处方审核中"),
        statusText: String(review.review_text || "处方信息已确认，可加入购物车购买"),
        reviewedAt: String(review.reviewed_at || review.updated_at || "").slice(11, 16) || "14:28",
        product: {
          name: String(review.product_name || "阿莫西林胶囊"),
          price: (Number(review.price_fen || 1880) / 100).toFixed(2),
          quantity: Number(review.quantity || 1),
          doctor: String(review.doctor_name || "王医生")
        },
        patientName: String(review.patient_name || "张三"),
        ocr: {
          status: String(ocr.status || "") === "need_review" ? "待人工复核" : "已识别",
          confidence: Number(ocr.confidence || 0),
          dosage: String(ocr.dosage_text || "按校医建议使用")
        },
        archive: {
          id: String(archive.archive_id || "rxa_preview"),
          retention: String(archive.retention_text || "校内处方留档 6 年")
        },
        steps: Array.isArray(review.steps) ? review.steps.map((item) => ({
          title: String(item.title || ""),
          subtitle: String(item.subtitle || ""),
          status: String(item.status || "pending")
        })) : this.data.steps
      });
    } catch (_error) {
      // Keep preview review result.
    }
  },
  handleContinue() {
    wx.navigateTo({ url: `/pages/medicine/order-confirm/index?prescription_id=${this.data.reviewId}` });
  },
  handleBackList() {
    wx.redirectTo({ url: "/pages/medicine/home/index" });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/medicine/home/index" }) });
  }
});
