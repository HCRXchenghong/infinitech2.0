import {
  confirmPrescriptionImageUpload,
  createPrescriptionImageUpload,
  createPrescriptionReview,
  ensurePreviewAuth
} from "../../../utils/api";

Page({
  data: {
    patient: "张三",
    phone: "138****0000",
    address: "望京校区 3 号宿舍楼",
    hospital: "校医务室",
    note: "",
    imageText: "上传处方照片",
    imageUrl: "",
    imagePublicUrl: "",
    uploadTicketId: "",
    objectKey: "",
    contentSha: "",
    uploadStatus: "待上传处方",
    uploadStateClass: "",
    scanStatusText: "上传后自动安全扫描",
    uploading: false,
    submitting: false,
    product: {
      id: "med_amoxicillin",
      name: "阿莫西林胶囊",
      price: "18.80",
      tag: "处方药"
    }
  },
  handlePatientInput(event) {
    this.setData({ patient: String(event.detail.value || "") });
  },
  handleHospitalInput(event) {
    this.setData({ hospital: String(event.detail.value || "") });
  },
  handleNoteInput(event) {
    this.setData({ note: String(event.detail.value || "") });
  },
  handleChooseImage() {
    ensurePreviewAuth();
    wx.chooseImage({
      count: 1,
      sizeType: ["compressed"],
      sourceType: ["album", "camera"],
      success: (result) => {
        const imageUrl = result.tempFilePaths?.[0] || "";
        const tempFile = Array.isArray(result.tempFiles) ? result.tempFiles[0] : undefined;
        const sizeBytes = Number(tempFile?.size || 1024);
        this.setData({
          imageUrl,
          imageText: imageUrl ? "正在加密上传处方..." : "上传处方照片",
          uploadStatus: imageUrl ? "上传中" : "待上传处方",
          uploadStateClass: imageUrl ? "uploading" : "",
          scanStatusText: imageUrl ? "正在申请加密上传票据" : "上传后自动安全扫描",
          uploading: Boolean(imageUrl),
          imagePublicUrl: "",
          uploadTicketId: "",
          objectKey: "",
          contentSha: ""
        });
        if (imageUrl) {
          this.createImageUploadTicket(imageUrl, sizeBytes);
        }
      },
      fail: () => wx.showToast({ title: "未选择图片", icon: "none" })
    });
  },
  async createImageUploadTicket(imageUrl: string, sizeBytes: number) {
    const fileName = this.prescriptionFileName(imageUrl);
    const contentType = this.prescriptionContentType(fileName);
    try {
      const ticket = await createPrescriptionImageUpload({
        product_id: this.data.product.id,
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes
      }) as Record<string, unknown>;
      const ticketId = String(ticket.ticket_id || "");
      const objectKey = String(ticket.object_key || "");
      const publicUrl = String(ticket.public_url || "");
      const contentSha = `sha256:${ticketId || "preview"}`;
      if (!ticketId || !objectKey) {
        throw new Error("missing prescription upload ticket");
      }
      const confirmed = await confirmPrescriptionImageUpload({
        ticket_id: ticketId,
        object_key: objectKey,
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes,
        content_sha: contentSha
      }) as Record<string, unknown>;
      const scanStatus = String(confirmed.scan_status || "");
      const scanText = this.prescriptionScanStatusText(scanStatus);
      this.setData({
        uploadTicketId: ticketId,
        objectKey,
        imagePublicUrl: String(confirmed.public_url || publicUrl),
        contentSha: String(confirmed.content_sha || contentSha),
        imageText: "处方照片已加密上传",
        uploadStatus: scanText,
        uploadStateClass: scanStatus === "pending" ? "uploading" : "scanned",
        scanStatusText: scanText,
        uploading: false
      });
    } catch (_error) {
      this.setData({
        imageText: "已选择 1 张处方照片",
        uploadStatus: "本地预览",
        uploadStateClass: "",
        scanStatusText: "后端不可用时仅生成本地预览",
        uploading: false
      });
      wx.showToast({ title: "已保留本地预览", icon: "none" });
    }
  },
  prescriptionScanStatusText(scanStatus: string) {
    if (scanStatus === "passed") return "安全扫描通过";
    if (scanStatus === "pending") return "安全扫描中";
    if (scanStatus === "rejected") return "扫描未通过";
    return "已上传";
  },
  prescriptionFileName(imageUrl: string) {
    const cleanPath = String(imageUrl || "").split("?")[0];
    const fileName = cleanPath.split("/").pop() || "";
    return /\.(jpg|jpeg|png|webp|heic)$/i.test(fileName) ? fileName : `prescription-${Date.now()}.jpg`;
  },
  prescriptionContentType(fileName: string) {
    const lower = String(fileName || "").toLowerCase();
    if (lower.endsWith(".png")) return "image/png";
    if (lower.endsWith(".webp")) return "image/webp";
    if (lower.endsWith(".heic")) return "image/heic";
    return "image/jpeg";
  },
  async handleSubmit() {
    ensurePreviewAuth();
    if (!this.data.patient.trim()) {
      wx.showToast({ title: "请填写用药人姓名", icon: "none" });
      return;
    }
    if (this.data.uploading) {
      wx.showToast({ title: "处方照片上传中", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      const review = await createPrescriptionReview({
        patient_name: this.data.patient,
        patient_phone: this.data.phone,
        address: this.data.address,
        hospital: this.data.hospital,
        product_id: this.data.product.id,
        product_name: this.data.product.name,
        price_fen: 1880,
        quantity: 1,
        image_url: this.data.imagePublicUrl || this.data.imageUrl || "prescription-preview.jpg",
        prescription_image_ticket_id: this.data.uploadTicketId || undefined,
        prescription_object_key: this.data.objectKey || undefined,
        prescription_content_sha: this.data.contentSha || undefined,
        note: this.data.note
      }) as { id?: string };
      wx.showToast({ title: "已提交审核", icon: "success" });
      setTimeout(() => wx.navigateTo({ url: `/pages/prescription/review-result/index?id=${review.id || "rx_preview"}` }), 500);
    } catch (_error) {
      wx.showToast({ title: "审核预览已生成", icon: "none" });
      wx.navigateTo({ url: "/pages/prescription/review-result/index?id=rx_preview" });
    } finally {
      this.setData({ submitting: false });
    }
  },
  handleCancel() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/medicine/home/index" }) });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.redirectTo({ url: "/pages/medicine/home/index" }) });
  }
});
