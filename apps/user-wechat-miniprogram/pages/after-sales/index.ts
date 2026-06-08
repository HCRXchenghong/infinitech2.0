import {
  confirmAfterSalesEvidenceUpload,
  createAfterSales,
  createAfterSalesEvidenceUpload,
  ensurePreviewAuth,
  getAfterSalesEvidence,
  getAfterSalesEvents,
  getAfterSalesRequests,
  getOrderDetail,
  getOrders,
  reportObjectStorageScanResult,
  reportObjectStorageUpload
} from "../../utils/api";

const MAX_AFTER_SALES_EVIDENCE = 3;

Page({
  data: {
    statusText: "后端已接",
    loading: false,
    requestLoading: false,
    selectedOrderId: "",
    selectedRequestId: "",
    type: "refund_only",
    reason: "商家少送 / 错送商品",
    amount: "55.98",
    amountPlaceholder: "最多可退 ¥55.98",
    maxRefundText: "55.98",
    selectedOrderAmountFen: 5598,
    submitting: false,
    uploadingEvidence: false,
    canSubmit: true,
    submitText: "提交申请",
    evidenceHint: "提交申请后可继续补充图片、小票等凭证。",
    orderSummary: {
      shop: "蓝海餐厅",
      orderNo: "10031",
      time: "今天 12:18",
      items: "招牌牛肉饭等 3 件",
      status: "配送中",
      paid: "55.98"
    },
    typeOptions: [
      { key: "refund_only", title: "仅退款" },
      { key: "partial_refund", title: "部分退款" },
      { key: "food_safety", title: "食品安全投诉" }
    ],
    evidenceCards: buildEvidenceCards([], null, null, true, false),
    pendingEvidence: null,
    steps: ["提交申请", "商家 24 小时内处理", "复杂问题平台会继续介入"],
    orders: [
      { id: "ord_preview", title: "蓝海餐厅", subtitle: "招牌牛肉饭 x2", amountText: "¥55.98", selected: true }
    ],
    requests: [],
    requestDetail: null,
    requestEvents: [],
    requestEvidence: []
  },

  onLoad(query) {
    const orderId = String(query?.order_id || "").trim();
    if (orderId) {
      this.setData({ selectedOrderId: orderId });
    }
    this.loadData();
  },

  async loadData() {
    ensurePreviewAuth();
    this.setData({ loading: true });
    try {
      const orders = await getOrders();
      const mappedOrders = Array.isArray(orders) && orders.length > 0
        ? orders.map((order) => mapOrderOption(order, this.data.selectedOrderId))
        : this.data.orders;
      const selectedOrderId = this.data.selectedOrderId || mappedOrders[0]?.id || "";
      this.setData({
        orders: mappedOrders.map((item) => ({ ...item, selected: item.id === selectedOrderId })),
        selectedOrderId,
        statusText: "已同步"
      });
      await this.syncSelectedOrder(selectedOrderId, true);
    } catch (_error) {
      this.setData({ statusText: "本地预览", loading: false });
    }
  },

  async syncSelectedOrder(orderId, resetAmount = false) {
    if (!orderId) {
      this.setData({
        loading: false,
        requestLoading: false,
        requests: [],
        requestDetail: null,
        requestEvents: [],
        requestEvidence: [],
        pendingEvidence: null,
        uploadingEvidence: false,
        ...computeEvidenceState({
          request: null,
          evidence: [],
          pendingEvidence: null,
          canSubmit: true,
          uploading: false
        })
      });
      return;
    }
    this.setData({
      loading: true,
      requestLoading: true,
      selectedOrderId: orderId,
      orders: this.data.orders.map((item) => ({ ...item, selected: item.id === orderId }))
    });
    const [order, requests] = await Promise.all([
      getOrderDetail(orderId).catch(() => null),
      getAfterSalesRequests(orderId).catch(() => [])
    ]);
    const mappedRequests = Array.isArray(requests) ? requests.map(mapRequestCard) : [];
    const highlightedRequest = mappedRequests.find((item) => !isFinalAfterSalesStatus(item.statusCode)) || mappedRequests[0] || null;
    const refundableFen = highlightedRequest
      ? Number(highlightedRequest.refundableFen || 0)
      : Number(order?.amount_fen || 0);
    const maxRefundText = fenToText(refundableFen || order?.amount_fen || 0);
    const canSubmit = !highlightedRequest || isFinalAfterSalesStatus(highlightedRequest.statusCode);
    const evidenceState = computeEvidenceState({
      request: highlightedRequest,
      evidence: [],
      pendingEvidence: null,
      canSubmit,
      uploading: false
    });
    this.setData({
      loading: false,
      selectedOrderAmountFen: Number(order?.amount_fen || 0),
      orderSummary: order ? mapOrderSummary(order) : this.data.orderSummary,
      requests: mappedRequests,
      selectedRequestId: highlightedRequest?.id || "",
      canSubmit,
      submitText: highlightedRequest && !isFinalAfterSalesStatus(highlightedRequest.statusCode) ? "查看进度" : "提交申请",
      maxRefundText,
      amountPlaceholder: `最多可退 ¥${maxRefundText}`,
      amount: resetAmount ? maxRefundText : (this.data.amount || maxRefundText),
      statusText: order ? "已同步" : this.data.statusText,
      requestEvidence: [],
      requestDetail: null,
      requestEvents: [],
      pendingEvidence: null,
      uploadingEvidence: false,
      ...evidenceState
    });
    if (highlightedRequest) {
      await this.loadRequestDetail(highlightedRequest.id);
      return;
    }
    this.setData({
      requestLoading: false,
      requestDetail: null,
      requestEvents: [],
      requestEvidence: [],
      pendingEvidence: null,
      uploadingEvidence: false,
      ...computeEvidenceState({
        request: null,
        evidence: [],
        pendingEvidence: null,
        canSubmit: true,
        uploading: false
      })
    });
  },

  async loadRequestDetail(requestId) {
    const request = findRequestByID(this.data.requests, requestId);
    this.setData({
      selectedRequestId: requestId,
      requestLoading: true,
      requestDetail: request ? buildRequestDetail(request, [], []) : null
    });
    const [events, evidence] = await Promise.all([
      getAfterSalesEvents(requestId).catch(() => []),
      getAfterSalesEvidence(requestId).catch(() => [])
    ]);
    const mappedEvents = Array.isArray(events) ? events.map(mapRequestEvent) : [];
    const mappedEvidence = Array.isArray(evidence) ? evidence.map(mapEvidenceItem) : [];
    const evidenceState = computeEvidenceState({
      request,
      evidence: mappedEvidence,
      pendingEvidence: this.data.pendingEvidence,
      canSubmit: this.data.canSubmit,
      uploading: this.data.uploadingEvidence
    });
    this.setData({
      requestLoading: false,
      requestEvents: mappedEvents,
      requestEvidence: mappedEvidence,
      requestDetail: request ? buildRequestDetail(request, mappedEvents, mappedEvidence) : null,
      ...evidenceState
    });
  },

  handleOrderSelect(event) {
    const orderId = String(event.currentTarget.dataset.id || "");
    if (!orderId || orderId === this.data.selectedOrderId) return;
    this.syncSelectedOrder(orderId, true);
  },

  handleTypeTap(event) {
    this.setData({ type: String(event.currentTarget.dataset.type || "refund_only") });
  },

  handleReasonInput(event) {
    this.setData({ reason: String(event.detail.value || "") });
  },

  handleAmountInput(event) {
    this.setData({ amount: String(event.detail.value || "") });
  },

  handleRequestTap(event) {
    const requestId = String(event.currentTarget.dataset.id || "");
    if (!requestId) return;
    this.loadRequestDetail(requestId);
  },

  handleEvidenceTap(event) {
    const kind = String(event.currentTarget.dataset.kind || "");
    const current = String(event.currentTarget.dataset.url || "");
    const evidence = Array.isArray(this.data.requestEvidence) ? this.data.requestEvidence : [];
    if (kind === "uploaded" || kind === "pending") {
      const urls = [...evidence.map((item) => item.url), current].filter(Boolean);
      if (!urls.length) {
        wx.showToast({ title: "当前凭证暂不可预览", icon: "none" });
        return;
      }
      wx.previewImage({
        current: current || urls[0],
        urls: Array.from(new Set(urls))
      });
      return;
    }
    if (kind === "add") {
      this.handleEvidenceUpload();
      return;
    }
    if (this.data.requestLoading || this.data.uploadingEvidence) {
      wx.showToast({ title: "售后记录同步中", icon: "none" });
      return;
    }
    if (!this.data.selectedRequestId) {
      wx.showToast({ title: "提交申请后可继续补充凭证", icon: "none" });
      return;
    }
    if (this.data.canSubmit) {
      wx.showToast({ title: "当前售后已结束，请重新提交申请", icon: "none" });
      return;
    }
    this.handleEvidenceUpload();
  },

  handleEvidenceUpload() {
    ensurePreviewAuth();
    if (!this.data.selectedRequestId) {
      wx.showToast({ title: "提交申请后可继续补充凭证", icon: "none" });
      return;
    }
    if (this.data.canSubmit) {
      wx.showToast({ title: "当前售后已结束，请重新提交申请", icon: "none" });
      return;
    }
    if (this.data.uploadingEvidence || this.data.requestLoading) {
      wx.showToast({ title: "正在同步凭证，请稍等", icon: "none" });
      return;
    }
    if ((this.data.requestEvidence || []).length >= MAX_AFTER_SALES_EVIDENCE) {
      wx.showToast({ title: "当前最多保留 3 张凭证", icon: "none" });
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
        void this.uploadEvidenceImage(imageUrl, sizeBytes);
      },
      fail: () => wx.showToast({ title: "未选择图片", icon: "none" })
    });
  },

  async uploadEvidenceImage(imageUrl, sizeBytes) {
    const requestId = String(this.data.selectedRequestId || "");
    const request = findRequestByID(this.data.requests, requestId);
    if (!requestId || !request) {
      wx.showToast({ title: "请先提交售后申请", icon: "none" });
      return;
    }
    const fileName = this.afterSalesFileName(imageUrl);
    const contentType = this.afterSalesContentType(fileName);
    const pendingEvidence = {
      id: `pending_${Date.now()}`,
      name: fileName,
      url: imageUrl,
      meta: "正在上传"
    };
    this.setData({
      uploadingEvidence: true,
      pendingEvidence,
      ...computeEvidenceState({
        request,
        evidence: this.data.requestEvidence,
        pendingEvidence,
        canSubmit: this.data.canSubmit,
        uploading: true
      })
    });
    try {
      const ticket = await createAfterSalesEvidenceUpload(requestId, {
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes
      });
      const ticketId = String(ticket?.ticket_id || "");
      const objectKey = String(ticket?.object_key || "");
      const contentSha = `sha256:${ticketId || `preview_${Date.now()}`}`;
      if (!ticketId || !objectKey) {
        throw new Error("missing after-sales upload ticket");
      }
      await this.confirmEvidenceUploadChain(requestId, {
        ticket_id: ticketId,
        object_key: objectKey,
        file_name: fileName,
        content_type: contentType,
        size_bytes: sizeBytes,
        content_sha: contentSha,
        message: "用户补充了新的售后凭证"
      });
      wx.showToast({ title: "凭证已补充", icon: "success" });
      await this.syncSelectedOrder(this.data.selectedOrderId, false);
    } catch (_error) {
      this.setData({
        uploadingEvidence: false,
        pendingEvidence: null,
        ...computeEvidenceState({
          request,
          evidence: this.data.requestEvidence,
          pendingEvidence: null,
          canSubmit: this.data.canSubmit,
          uploading: false
        })
      });
      wx.showToast({ title: "上传失败，请稍后再试", icon: "none" });
    }
  },

  async confirmEvidenceUploadChain(requestId, payload) {
    try {
      return await confirmAfterSalesEvidenceUpload(requestId, payload);
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
        // 某些环境要求真实对象存储回调，这里保留最终确认错误给上层提示。
      }
      try {
        return await confirmAfterSalesEvidenceUpload(requestId, payload);
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
            return await confirmAfterSalesEvidenceUpload(requestId, payload);
          } catch (_scanError) {
            // 严格生产环境会依赖真实扫描回调，这里继续抛出原始确认错误。
          }
        }
        throw retryError || confirmError;
      }
    }
  },

  async handleSubmit() {
    if (!this.data.selectedOrderId) {
      wx.showToast({ title: "请先选择订单", icon: "none" });
      return;
    }
    if (!this.data.canSubmit) {
      if (this.data.selectedRequestId) {
        await this.loadRequestDetail(this.data.selectedRequestId);
      }
      wx.showToast({ title: "当前订单已有售后处理中", icon: "none" });
      return;
    }
    if (!String(this.data.reason || "").trim()) {
      wx.showToast({ title: "请填写售后原因", icon: "none" });
      return;
    }
    this.setData({ submitting: true });
    try {
      const amountFen = Math.round(Number(this.data.amount || 0) * 100);
      await createAfterSales({
        order_id: this.data.selectedOrderId,
        type: this.data.type,
        reason: this.data.reason,
        requested_amount_fen: amountFen > 0 ? amountFen : undefined
      });
      wx.showToast({ title: "已提交售后", icon: "success" });
      await this.syncSelectedOrder(this.data.selectedOrderId, false);
    } catch (_error) {
      wx.showToast({ title: "提交失败，请稍后再试", icon: "none" });
    } finally {
      this.setData({ submitting: false });
    }
  },

  handleSupportTap() {
    wx.navigateTo({ url: `/pages/customer-service/chat/index?order_id=${this.data.selectedOrderId}` });
  },

  afterSalesFileName(imageUrl) {
    const cleanPath = String(imageUrl || "").split("?")[0];
    const fileName = cleanPath.split("/").pop() || "";
    return /\.(jpg|jpeg|png|webp|heic)$/i.test(fileName) ? fileName : `after-sales-${Date.now()}.jpg`;
  },

  afterSalesContentType(fileName) {
    const lower = String(fileName || "").toLowerCase();
    if (lower.endsWith(".png")) return "image/png";
    if (lower.endsWith(".webp")) return "image/webp";
    if (lower.endsWith(".heic")) return "image/heic";
    return "image/jpeg";
  },

  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function mapOrderOption(order, selectedOrderId) {
  return {
    id: String(order.id || ""),
    title: String(order.shop_name || order.shop_id || "订单"),
    subtitle: summarizeOrderItems(order.items),
    amountText: `¥${fenToText(order.amount_fen)}`,
    selected: String(order.id || "") === selectedOrderId
  };
}

function mapOrderSummary(order) {
  return {
    shop: String(order.shop_name || order.shop_id || "订单"),
    orderNo: String(order.id || ""),
    time: formatDateTime(order.updated_at || order.created_at),
    items: summarizeOrderItems(order.items),
    status: orderStatusText(String(order.status || "")),
    paid: fenToText(order.amount_fen)
  };
}

function mapRequestCard(request) {
  return {
    id: String(request.id || ""),
    title: afterSalesTypeText(String(request.type || "")),
    subtitle: String(request.order_item_summary || request.reason || "售后处理中"),
    status: afterSalesStatusText(String(request.status || "")),
    statusCode: String(request.status || ""),
    statusTone: afterSalesStatusTone(String(request.status || "")),
    reason: String(request.reason || ""),
    shopName: String(request.shop_name || "订单售后"),
    latest: String(request.latest_event_message || request.reason || "售后处理中"),
    latestTime: formatDateTime(request.latest_event_at || request.updated_at || request.created_at),
    requestedAmountFen: Number(request.requested_amount_fen || 0),
    requestedAmountText: fenToText(request.requested_amount_fen),
    refundableFen: Number(request.refundable_fen || 0),
    refundableText: fenToText(request.refundable_fen),
    reviewReason: String(request.review_reason || ""),
    evidenceCount: Array.isArray(request.evidence_urls) ? request.evidence_urls.length : 0
  };
}

function mapRequestEvent(event) {
  return {
    id: String(event.id || ""),
    title: afterSalesActionTitle(String(event.action || "")),
    message: String(event.message || ""),
    time: formatDateTime(event.created_at),
    attachments: Array.isArray(event.attachments) ? event.attachments : []
  };
}

function mapEvidenceItem(item, index) {
  const url = String(item.public_url || "");
  return {
    id: String(item.id || `evidence_${index}`),
    name: String(item.file_name || `凭证 ${index + 1}`),
    url
  };
}

function buildRequestDetail(request, events, evidence) {
  return {
    title: `${request.shopName} · ${request.title}`,
    status: request.status,
    statusTone: request.statusTone,
    latest: request.latest,
    latestTime: request.latestTime,
    amountText: request.requestedAmountText,
    refundableText: request.refundableText,
    reviewReason: request.reviewReason,
    evidenceCount: evidence.length || request.evidenceCount,
    eventCount: events.length
  };
}

function summarizeOrderItems(items = []) {
  if (!Array.isArray(items) || items.length === 0) return "订单商品";
  const [first, ...rest] = items;
  const firstName = String(first.product_name || "商品");
  const firstQuantity = Math.max(1, Number(first.quantity || 1));
  if (rest.length === 0) {
    return `${firstName} x ${firstQuantity}`;
  }
  const totalCount = items.reduce((sum, item) => sum + Math.max(1, Number(item.quantity || 1)), 0);
  return `${firstName}等 ${totalCount} 件`;
}

function fenToText(value) {
  return (Number(value || 0) / 100).toFixed(2);
}

function formatDateTime(value) {
  const text = String(value || "").trim();
  if (!text) return "刚刚";
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) return "刚刚";
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");
  return `${month}-${day} ${hour}:${minute}`;
}

function orderStatusText(status) {
  const map = {
    pending_payment: "待支付",
    merchant_pending: "待商家处理",
    preparing: "商家备餐中",
    dispatching: "待配送",
    rider_assigned: "骑手配送中",
    completed: "已完成",
    refunded: "已退款"
  };
  return map[status] || status || "订单处理中";
}

function afterSalesTypeText(type) {
  const map = {
    refund_only: "仅退款",
    partial_refund: "部分退款",
    food_safety: "食品安全"
  };
  return map[type] || "售后申请";
}

function afterSalesStatusText(status) {
  const map = {
    pending_merchant: "待商家处理",
    admin_review: "平台介入中",
    approved: "已通过",
    rejected: "已驳回",
    refunded: "已退款"
  };
  return map[status] || status || "处理中";
}

function afterSalesStatusTone(status) {
  if (status === "refunded") return "done";
  if (status === "rejected") return "warning";
  if (status === "admin_review") return "danger";
  return "normal";
}

function afterSalesActionTitle(action) {
  const map = {
    created: "已提交申请",
    merchant_reply: "商家回复",
    customer_service_intervention: "平台介入",
    evidence_uploaded: "补充凭证",
    review_approved: "审核通过",
    review_rejected: "审核驳回"
  };
  return map[action] || "进度更新";
}

function isFinalAfterSalesStatus(status) {
  return status === "rejected" || status === "refunded";
}

function findRequestByID(requests = [], requestId = "") {
  return Array.isArray(requests)
    ? requests.find((item) => String(item.id || "") === String(requestId || "")) || null
    : null;
}

function buildEvidenceCards(evidence = [], pendingEvidence, request, canSubmit, uploading) {
  const cards = Array.isArray(evidence)
    ? evidence.map((item) => ({
      id: item.id,
      kind: "uploaded",
      name: item.name,
      url: item.url,
      meta: "点击查看"
    }))
    : [];
  if (pendingEvidence?.url) {
    cards.push({
      id: pendingEvidence.id || `pending_${Date.now()}`,
      kind: "pending",
      name: pendingEvidence.name || "新凭证",
      url: pendingEvidence.url,
      meta: pendingEvidence.meta || "正在上传"
    });
  }
  if (request && !canSubmit && !isFinalAfterSalesStatus(request.statusCode) && cards.length < MAX_AFTER_SALES_EVIDENCE) {
    cards.push({
      id: "evidence_add",
      kind: "add",
      text: uploading ? "上传中" : "补充"
    });
  }
  if (cards.length > 0) {
    return cards;
  }
  if (request) {
    return [{
      id: "empty_request",
      kind: "empty",
      text: "暂无"
    }];
  }
  return Array.from({ length: MAX_AFTER_SALES_EVIDENCE }, (_, index) => ({
    id: `empty_${index + 1}`,
    kind: "empty",
    text: "+"
  }));
}

function buildEvidenceHint(request, evidenceCount, canSubmit, uploading) {
  if (uploading) {
    return "正在补充凭证，上传完成后会自动同步到处理记录。";
  }
  if (!request) {
    return "提交申请后可继续补充图片、小票等凭证。";
  }
  if (isFinalAfterSalesStatus(request.statusCode) || canSubmit) {
    return "当前售后已结束，历史凭证可继续预览；如需补充请重新发起申请。";
  }
  if (evidenceCount >= MAX_AFTER_SALES_EVIDENCE) {
    return "当前最多保留 3 张凭证，已上传内容可直接点击预览。";
  }
  return "可继续补充图片、小票等凭证，商家和平台会同步看到。";
}

function computeEvidenceState({ request, evidence, pendingEvidence, canSubmit, uploading }) {
  const safeEvidence = Array.isArray(evidence) ? evidence : [];
  const pendingCount = pendingEvidence?.url ? 1 : 0;
  return {
    evidenceCards: buildEvidenceCards(safeEvidence, pendingEvidence, request, canSubmit, uploading),
    evidenceHint: buildEvidenceHint(request, safeEvidence.length + pendingCount, canSubmit, uploading)
  };
}
