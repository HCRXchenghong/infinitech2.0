export function normalizeMoneyFen(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return 0;
  return Math.max(0, Math.trunc(numeric));
}

export function normalizeOutboxConsumerEvent(event = {}, consumerName = "", options = {}) {
  const payload = event && typeof event.payload === "object" && event.payload !== null && !Array.isArray(event.payload)
    ? event.payload
    : {};
  const topic = String(event.topic || payload.topic || "").trim();
  const outboxEventId = String(event.id || event.event_id || event.outbox_event_id || payload.outbox_event_id || "").trim();
  const aggregateType = String(event.aggregate_type || event.aggregateType || payload.aggregate_type || payload.aggregateType || "").trim();
  const aggregateId = String(event.aggregate_id || event.aggregateId || payload.aggregate_id || payload.aggregateId || "").trim();
  const eventType = String(event.event_type || event.eventType || event.type || payload.event_type || payload.eventType || payload.type || "").trim();
  const explicitIdempotencyKey = String(event.idempotency_key || event.idempotencyKey || payload.idempotency_key || payload.idempotencyKey || "").trim();
  const fallbackKey = [topic, aggregateType, aggregateId, eventType].filter(Boolean).join(":");
  const idempotencyKey = explicitIdempotencyKey || outboxEventId || fallbackKey;
  const consumer = String(consumerName || event.consumer_name || event.consumerName || "").trim();
  const now = options.now instanceof Date ? options.now : new Date();
  return {
    consumer_name: consumer,
    consumer_event_key: consumer && idempotencyKey ? `${consumer}:${idempotencyKey}` : "",
    idempotency_key: idempotencyKey,
    outbox_event_id: outboxEventId,
    topic,
    aggregate_type: aggregateType,
    aggregate_id: aggregateId,
    event_type: eventType,
    first_seen_at: now.toISOString(),
    status: "processing",
    attempts: 0,
    last_error: ""
  };
}

export function createConsumedEventLedger(initialRecords = []) {
  const records = new Map();
  for (const record of Array.isArray(initialRecords) ? initialRecords : []) {
    const key = String(record.consumer_event_key || "").trim();
    if (key) {
      records.set(key, { ...record });
    }
  }
  return {
    reserve(record = {}, now = new Date()) {
      const key = String(record.consumer_event_key || "").trim();
      if (!key || !record.consumer_name || !record.idempotency_key) {
        return { ok: false, duplicate: false, code: "INVALID_CONSUMED_EVENT", record: { ...record } };
      }
      const current = records.get(key);
      if (current && (current.status === "processing" || current.status === "processed")) {
        return { ok: false, duplicate: true, code: "DUPLICATE_CONSUMED_EVENT", record: { ...current } };
      }
      const timestamp = now.toISOString();
      const next = {
        ...record,
        status: "processing",
        attempts: Number(current?.attempts || 0) + 1,
        first_seen_at: current?.first_seen_at || record.first_seen_at || timestamp,
        processed_at: "",
        updated_at: timestamp,
        last_error: ""
      };
      records.set(key, next);
      return { ok: true, duplicate: false, code: "", record: { ...next } };
    },
    complete(record = {}, now = new Date()) {
      const key = String(record.consumer_event_key || "").trim();
      const current = records.get(key) || record;
      const timestamp = now.toISOString();
      const next = {
        ...current,
        status: "processed",
        processed_at: timestamp,
        updated_at: timestamp,
        last_error: ""
      };
      records.set(key, next);
      return { ok: true, record: { ...next } };
    },
    fail(record = {}, error = new Error("consumer failed"), now = new Date()) {
      const key = String(record.consumer_event_key || "").trim();
      const current = records.get(key) || record;
      const timestamp = now.toISOString();
      const next = {
        ...current,
        status: "failed",
        processed_at: "",
        updated_at: timestamp,
        last_error: String(error?.message || error || "consumer failed")
      };
      records.set(key, next);
      return { ok: true, record: { ...next } };
    },
    snapshot() {
      return [...records.values()].map((record) => ({ ...record }));
    }
  };
}

export function createIdempotentConsumer(options = {}) {
  const consumerName = String(options.consumerName || options.consumer_name || "").trim();
  const handler = typeof options.handler === "function" ? options.handler : async () => ({ ok: true });
  const ledger = options.ledger || createConsumedEventLedger();
  const clock = typeof options.clock === "function" ? options.clock : () => new Date();
  const consume = async (event = {}) => {
    const record = normalizeOutboxConsumerEvent(event, consumerName, { now: clock() });
    const reservation = await ledger.reserve(record, clock());
    if (!reservation.ok) {
      return {
        status: reservation.duplicate ? "duplicate" : "invalid",
        consumed: false,
        duplicate: reservation.duplicate,
        record: reservation.record,
        result: null
      };
    }
    try {
      const result = await handler(event, reservation.record);
      const completion = await ledger.complete(reservation.record, clock());
      return {
        status: "processed",
        consumed: true,
        duplicate: false,
        record: completion.record,
        result
      };
    } catch (error) {
      await ledger.fail(reservation.record, error, clock());
      throw error;
    }
  };
  consume.ledger = ledger;
  return consume;
}

export function buildHomeModules(input = []) {
  return [...input]
    .filter((item) => item && item.enabled !== false)
    .map((item) => ({
      key: String(item.key || "").trim(),
      title: String(item.title || "").trim(),
      route: String(item.route || "").trim(),
      icon: String(item.icon || "").trim(),
      sort_order: Number.isFinite(Number(item.sort_order)) ? Number(item.sort_order) : 0,
      scene: String(item.scene || "home").trim() || "home"
    }))
    .filter((item) => item.key && item.title && item.route)
    .sort((left, right) => left.sort_order - right.sort_order || left.key.localeCompare(right.key));
}

const HOME_CARD_TYPES = new Set(["product", "shop", "groupbuy_deal", "coupon", "circle_post"]);

export function normalizeHomeCard(card = {}) {
  return {
    id: String(card.id || "").trim(),
    type: HOME_CARD_TYPES.has(String(card.type || "").trim()) ? String(card.type || "").trim() : "product",
    title: String(card.title || "").trim(),
    subtitle: String(card.subtitle || "").trim(),
    image_url: String(card.image_url || card.imageUrl || "").trim(),
    target_id: String(card.target_id || card.targetId || "").trim(),
    shop_id: String(card.shop_id || card.shopId || "").trim(),
    price_fen: normalizeMoneyFen(card.price_fen ?? card.priceFen),
    enabled: card.enabled !== false,
    sort_order: Number.isFinite(Number(card.sort_order ?? card.sortOrder)) ? Number(card.sort_order ?? card.sortOrder) : 0,
    starts_at: String(card.starts_at || card.startsAt || "").trim(),
    ends_at: String(card.ends_at || card.endsAt || "").trim()
  };
}

export function normalizeUserAddress(address = {}) {
  return {
    id: String(address.id || "").trim(),
    user_id: String(address.user_id || address.userId || "").trim(),
    contact_name: String(address.contact_name || address.contactName || "").trim(),
    contact_phone: String(address.contact_phone || address.contactPhone || "").trim(),
    province: String(address.province || "").trim(),
    city: String(address.city || "").trim(),
    district: String(address.district || "").trim(),
    street: String(address.street || "").trim(),
    detail: String(address.detail || "").trim(),
    latitude: Number.isFinite(Number(address.latitude)) ? Number(address.latitude) : null,
    longitude: Number.isFinite(Number(address.longitude)) ? Number(address.longitude) : null,
    tag: String(address.tag || "other").trim() || "other",
    is_default: address.is_default === true || address.isDefault === true
  };
}

export function canUseAddressForDelivery(address = {}) {
  const normalized = normalizeUserAddress(address);
  const missing = [];
  if (!normalized.contact_name) missing.push("contact_name");
  if (!normalized.contact_phone) missing.push("contact_phone");
  if (!normalized.city) missing.push("city");
  if (!normalized.detail) missing.push("detail");
  if (normalized.latitude === null || normalized.longitude === null) missing.push("geo_location");
  return {
    ok: missing.length === 0,
    missing_fields: missing
  };
}

export function normalizeOrderOptions(options = {}) {
  return {
    remark: String(options.remark || "").trim().slice(0, 120),
    tableware_count: Number.isFinite(Number(options.tableware_count ?? options.tablewareCount))
      ? Math.max(0, Math.min(99, Math.trunc(Number(options.tableware_count ?? options.tablewareCount))))
      : 0,
    contactless_delivery: options.contactless_delivery === true || options.contactlessDelivery === true,
    invoice_requested: options.invoice_requested === true || options.invoiceRequested === true
  };
}

export function calculateOrderPayable(amounts = {}) {
  const items = normalizeMoneyFen(amounts.items_total_fen ?? amounts.itemsTotalFen);
  const delivery = normalizeMoneyFen(amounts.delivery_fee_fen ?? amounts.deliveryFeeFen);
  const packaging = normalizeMoneyFen(amounts.packaging_fee_fen ?? amounts.packagingFeeFen);
  const discount = normalizeMoneyFen(amounts.discount_fen ?? amounts.discountFen);
  const payable = Math.max(0, items + delivery + packaging - discount);
  return {
    items_total_fen: items,
    delivery_fee_fen: delivery,
    packaging_fee_fen: packaging,
    discount_fen: discount,
    payable_fen: payable
  };
}

export function normalizeCartItem(item = {}) {
  return {
    user_id: String(item.user_id || item.userId || "").trim(),
    shop_id: String(item.shop_id || item.shopId || "").trim(),
    product_id: String(item.product_id || item.productId || "").trim(),
    product_name: String(item.product_name || item.productName || "").trim(),
    unit_price_fen: normalizeMoneyFen(item.unit_price_fen ?? item.unitPriceFen),
    quantity: Number.isFinite(Number(item.quantity)) ? Math.max(0, Math.trunc(Number(item.quantity))) : 0,
    selected: item.selected !== false,
    status: String(item.status || "active").trim() || "active"
  };
}

export function upsertCartItem(cartItems = [], nextItem = {}) {
  const normalized = normalizeCartItem(nextItem);
  if (!normalized.user_id || !normalized.shop_id || !normalized.product_id) {
    return Array.isArray(cartItems) ? cartItems.map(normalizeCartItem) : [];
  }
  const current = Array.isArray(cartItems) ? cartItems.map(normalizeCartItem) : [];
  const withoutExisting = current.filter((item) => item.product_id !== normalized.product_id);
  if (normalized.quantity <= 0) {
    return withoutExisting;
  }
  return [...withoutExisting, normalized].sort((left, right) => left.product_id.localeCompare(right.product_id));
}

export function buildCartSummary(cartItems = [], amounts = {}) {
  const items = (Array.isArray(cartItems) ? cartItems : [])
    .map(normalizeCartItem)
    .filter((item) => item.selected && item.status === "active" && item.quantity > 0 && item.unit_price_fen > 0);
  const itemsTotalFen = items.reduce((sum, item) => sum + item.unit_price_fen * item.quantity, 0);
  const deliveryFeeFen = normalizeMoneyFen(amounts.delivery_fee_fen ?? amounts.deliveryFeeFen);
  const packagingFeeFen = normalizeMoneyFen(amounts.packaging_fee_fen ?? amounts.packagingFeeFen);
  const discountFen = normalizeMoneyFen(amounts.discount_fen ?? amounts.discountFen);
  return {
    user_id: items[0]?.user_id || String(amounts.user_id || amounts.userId || "").trim(),
    shop_id: items[0]?.shop_id || String(amounts.shop_id || amounts.shopId || "").trim(),
    items,
    ...calculateOrderPayable({
      items_total_fen: itemsTotalFen,
      delivery_fee_fen: deliveryFeeFen,
      packaging_fee_fen: packagingFeeFen,
      discount_fen: discountFen
    })
  };
}

export function canCheckoutCart(summary = {}, address = {}) {
  const addressState = canUseAddressForDelivery(address);
  const items = Array.isArray(summary.items) ? summary.items : [];
  const payable = normalizeMoneyFen(summary.payable_fen ?? summary.payableFen);
  const missing = [];
  if (items.length === 0) missing.push("cart_items");
  if (payable <= 0) missing.push("payable_fen");
  return {
    ok: missing.length === 0 && addressState.ok,
    missing_fields: [...missing, ...addressState.missing_fields]
  };
}

export function buildControlledHomeCards(cards = [], now = new Date()) {
  return (Array.isArray(cards) ? cards : [])
    .map(normalizeHomeCard)
    .filter((card) => card.enabled && card.id && card.title && card.target_id)
    .filter((card) => isWithinSchedule(card, now))
    .sort((left, right) => left.sort_order - right.sort_order || left.id.localeCompare(right.id));
}

function isWithinSchedule(item = {}, now = new Date()) {
  const startsAt = Date.parse(item.starts_at || "");
  const endsAt = Date.parse(item.ends_at || "");
  if (Number.isFinite(startsAt) && startsAt > now.getTime()) return false;
  if (Number.isFinite(endsAt) && endsAt <= now.getTime()) return false;
  return true;
}

export function canUseFreeDispatchCancel(cancelState = {}, now = new Date()) {
  const usedOn = String(cancelState.used_on || "").slice(0, 10);
  const today = now.toISOString().slice(0, 10);
  return usedOn !== today;
}

export function consumeFreeDispatchCancel(cancelState = {}, now = new Date()) {
  if (!canUseFreeDispatchCancel(cancelState, now)) {
    return { allowed: false, used_on: String(cancelState.used_on || "").slice(0, 10) };
  }
  return { allowed: true, used_on: now.toISOString().slice(0, 10) };
}

const CAPABILITY_BY_ORDER_TYPE = Object.freeze({
  takeout: "takeout",
  groupbuy: "groupbuy",
  medicine: "medicine"
});

export function normalizeShopProfile(shop = {}) {
  const capabilities = Array.isArray(shop.capabilities) ? shop.capabilities : [];
  return {
    id: String(shop.id || "").trim(),
    merchant_id: String(shop.merchant_id || shop.merchantId || "").trim(),
    name: String(shop.name || "").trim(),
    category: String(shop.category || "restaurant").trim() || "restaurant",
    account_type: String(shop.account_type || shop.accountType || "standard").trim() || "standard",
    capabilities: [...new Set(capabilities.map((item) => String(item).trim()).filter(Boolean))].sort(),
    qualifications: Array.isArray(shop.qualifications)
      ? shop.qualifications.map((item) => String(item).trim()).filter(Boolean)
      : [],
    display: {
      cover_url: String(shop.display?.cover_url || shop.display?.coverUrl || "").trim(),
      logo_url: String(shop.display?.logo_url || shop.display?.logoUrl || "").trim(),
      announcement: String(shop.display?.announcement || "").trim()
    }
  };
}

export function normalizeMerchantInvite(invite = {}) {
  return {
    token: String(invite.token || "").trim(),
    type: String(invite.type || "").trim(),
    status: String(invite.status || "").trim() || "active",
    merchant_id: String(invite.merchant_id || invite.merchantId || "").trim(),
    expires_at: String(invite.expires_at || invite.expiresAt || "").trim(),
    created_by_admin_id: String(invite.created_by_admin_id || invite.createdByAdminId || "").trim()
  };
}

export function canUseMerchantInvite(invite = {}, now = new Date()) {
  const normalized = normalizeMerchantInvite(invite);
  if (!normalized.token || normalized.type !== "merchant" || normalized.status !== "active") {
    return false;
  }
  if (!normalized.created_by_admin_id) {
    return false;
  }
  if (!normalized.expires_at) {
    return true;
  }
  const expiresAt = Date.parse(normalized.expires_at);
  return Number.isFinite(expiresAt) && expiresAt > now.getTime();
}

export function canUseRiderInvite(invite = {}, now = new Date()) {
  const normalized = normalizeMerchantInvite(invite);
  if (!normalized.token || !["rider", "station_manager"].includes(normalized.type) || normalized.status !== "active") {
    return false;
  }
  if (!normalized.created_by_admin_id) {
    return false;
  }
  if (!normalized.expires_at) {
    return true;
  }
  const expiresAt = Date.parse(normalized.expires_at);
  return Number.isFinite(expiresAt) && expiresAt > now.getTime();
}

export function normalizeStationTaskConfig(config = {}) {
  return {
    station_id: String(config.station_id || config.stationId || "").trim(),
    configured_by_station_manager_id: String(config.configured_by_station_manager_id || config.configuredByStationManagerId || "").trim(),
    daily_task_duration_minutes: Number.isFinite(Number(config.daily_task_duration_minutes ?? config.dailyTaskDurationMinutes))
      ? Math.max(0, Math.trunc(Number(config.daily_task_duration_minutes ?? config.dailyTaskDurationMinutes)))
      : 480
  };
}

export function canStationManagerManualDispatch(account = {}) {
  return String(account.account_type || account.accountType || "").trim() === "station_manager"
    && String(account.status || "active").trim() === "active";
}

export function normalizeDeposit(deposit = {}) {
  return {
    subject_type: String(deposit.subject_type || deposit.subjectType || "").trim(),
    subject_id: String(deposit.subject_id || deposit.subjectId || "").trim(),
    amount_fen: normalizeMoneyFen(deposit.amount_fen ?? deposit.amountFen),
    status: String(deposit.status || "unpaid").trim() || "unpaid",
    wechat_exempt_application_id: String(deposit.wechat_exempt_application_id || deposit.wechatExemptApplicationId || "").trim(),
    last_order_completed_at: String(deposit.last_order_completed_at || deposit.lastOrderCompletedAt || "").trim(),
    resignation_submitted_at: String(deposit.resignation_submitted_at || deposit.resignationSubmittedAt || "").trim(),
    dispute_closed_at: String(deposit.dispute_closed_at || deposit.disputeClosedAt || "").trim(),
    dispute_open: deposit.dispute_open === true || deposit.disputeOpen === true
  };
}

export function canRiderAcceptOrders(rider = {}, deposit = {}) {
  const accountType = String(rider.account_type || rider.accountType || "rider").trim();
  const status = String(rider.status || "active").trim();
  const normalizedDeposit = normalizeDeposit(deposit);
  return accountType === "rider"
    && status === "active"
    && rider.online === true
    && normalizedDeposit.subject_type === "rider"
    && (
      (normalizedDeposit.status === "paid" && normalizedDeposit.amount_fen >= 5000)
      || normalizedDeposit.status === "wechat_exempt_approved"
    );
}

export function canMerchantAcceptOrders(merchant = {}, deposit = {}) {
  const status = String(merchant.status || "active").trim();
  const normalizedDeposit = normalizeDeposit(deposit);
  return status === "active"
    && normalizedDeposit.subject_type === "merchant"
    && normalizedDeposit.status === "paid"
    && normalizedDeposit.amount_fen >= 5000;
}

export function resolveDispatchMode(order = {}, now = new Date(), policy = {}) {
  const createdAt = Date.parse(order.created_at || order.createdAt || "");
  const grabHallSeconds = Number.isFinite(Number(policy.grab_hall_seconds ?? policy.grabHallSeconds))
    ? Number(policy.grab_hall_seconds ?? policy.grabHallSeconds)
    : 600;
  if (!Number.isFinite(createdAt)) {
    return "grab_hall";
  }
  const elapsedSeconds = Math.max(0, Math.floor((now.getTime() - createdAt) / 1000));
  return elapsedSeconds >= grabHallSeconds ? "auto_assign" : "grab_hall";
}

export function selectNextOnlineRider(riders = [], rejectedRiderIds = []) {
  const rejected = new Set(rejectedRiderIds.map((id) => String(id)));
  const candidates = riders
    .filter((rider) => rider.online === true)
    .filter((rider) => !rejected.has(String(rider.id || "")))
    .filter((rider) => rider.capacity === undefined || Number(rider.capacity) > 0)
    .filter((rider) => rider.qualified !== false)
    .sort((left, right) => {
      const leftScore = Number(left.dispatch_score ?? left.score ?? 0);
      const rightScore = Number(right.dispatch_score ?? right.score ?? 0);
      if (leftScore !== rightScore) return rightScore - leftScore;
      return (left.distance_meters ?? Infinity) - (right.distance_meters ?? Infinity);
    });
  return candidates[0] || null;
}

export function normalizeRiderPerformance(performance = {}) {
  const accepted = normalizeMoneyFen(performance.accepted_order_count ?? performance.acceptedOrderCount);
  const completed = normalizeMoneyFen(performance.completed_order_count ?? performance.completedOrderCount);
  const onlineMinutes = normalizeMoneyFen(performance.online_minutes ?? performance.onlineMinutes);
  const activeDays = Math.max(1, normalizeMoneyFen(performance.active_days ?? performance.activeDays) || 1);
  const averageAcceptSeconds = Number.isFinite(Number(performance.average_accept_seconds ?? performance.averageAcceptSeconds))
    ? Math.max(0, Number(performance.average_accept_seconds ?? performance.averageAcceptSeconds))
    : 0;
  return {
    rider_id: String(performance.rider_id || performance.riderId || performance.id || "").trim(),
    station_id: String(performance.station_id || performance.stationId || "").trim(),
    accepted_order_count: accepted,
    completed_order_count: completed,
    online_minutes: onlineMinutes,
    active_days: activeDays,
    average_accept_seconds: averageAcceptSeconds,
    average_daily_orders: completed / activeDays,
    completion_rate: accepted > 0 ? completed / accepted : 0
  };
}

export function evaluateRiderLevel(riderPerformance = {}, teamPerformances = []) {
  const rider = normalizeRiderPerformance(riderPerformance);
  const team = teamPerformances.length > 0 ? teamPerformances.map(normalizeRiderPerformance) : [rider];
  const teamAverageAcceptSeconds = average(team.map((item) => item.average_accept_seconds).filter((value) => value > 0)) || rider.average_accept_seconds || 1;
  const teamAverageDailyOrders = average(team.map((item) => item.average_daily_orders).filter((value) => value > 0)) || rider.average_daily_orders || 1;
  const acceptScore = rider.average_accept_seconds > 0 ? (teamAverageAcceptSeconds / rider.average_accept_seconds) * 50 : 0;
  const orderScore = teamAverageDailyOrders > 0 ? (rider.average_daily_orders / teamAverageDailyOrders) * 35 : 0;
  const completionScore = rider.completion_rate * 15;
  const score = Math.max(0, Math.round(acceptScore + orderScore + completionScore));
  const level = score >= 120 ? "S" : score >= 100 ? "A" : score >= 80 ? "B" : "C";
  const priority = { S: 400, A: 300, B: 200, C: 100 }[level];
  return {
    rider_id: rider.rider_id,
    station_id: rider.station_id,
    score,
    level,
    dispatch_priority: priority,
    average_accept_seconds: rider.average_accept_seconds,
    average_daily_orders: Number(rider.average_daily_orders.toFixed(2)),
    completion_rate: Number(rider.completion_rate.toFixed(4))
  };
}

export function rankRidersForPriorityDispatch(riders = [], rejectedRiderIds = []) {
  const rejected = new Set(rejectedRiderIds.map((id) => String(id)));
  return [...riders]
    .filter((rider) => rider.online === true)
    .filter((rider) => !rejected.has(String(rider.id || rider.rider_id || "")))
    .filter((rider) => rider.capacity === undefined || Number(rider.capacity) > 0)
    .sort((left, right) => {
      const leftPriority = Number(left.dispatch_priority ?? left.priority ?? 0);
      const rightPriority = Number(right.dispatch_priority ?? right.priority ?? 0);
      if (leftPriority !== rightPriority) return rightPriority - leftPriority;
      const leftAccept = Number(left.average_accept_seconds ?? Infinity);
      const rightAccept = Number(right.average_accept_seconds ?? Infinity);
      if (leftAccept !== rightAccept) return leftAccept - rightAccept;
      return (left.distance_meters ?? Infinity) - (right.distance_meters ?? Infinity);
    });
}

export function canDeclineDispatchWithoutPenalty(progress = {}, config = {}) {
  const completed = normalizeMoneyFen(progress.completed_order_count ?? progress.completedOrderCount);
  const fixedCount = normalizeMoneyFen(config.fixed_daily_order_count ?? config.fixedDailyOrderCount);
  if (fixedCount <= 0) return false;
  return completed >= fixedCount;
}

export function normalizeAfterSalesRequest(request = {}) {
  return {
    id: String(request.id || "").trim(),
    order_id: String(request.order_id || request.orderId || "").trim(),
    user_id: String(request.user_id || request.userId || "").trim(),
    type: String(request.type || "refund_only").trim() || "refund_only",
    reason: String(request.reason || "").trim(),
    requested_amount_fen: normalizeMoneyFen(request.requested_amount_fen ?? request.requestedAmountFen),
    order_amount_fen: normalizeMoneyFen(request.order_amount_fen ?? request.orderAmountFen),
    refunded_amount_fen: normalizeMoneyFen(request.refunded_amount_fen ?? request.refundedAmountFen),
    refundable_fen: normalizeMoneyFen(request.refundable_fen ?? request.refundableFen),
    evidence_urls: normalizeStringList(request.evidence_urls || request.evidenceUrls),
    status: String(request.status || "pending_merchant").trim() || "pending_merchant"
  };
}

export function normalizeAfterSalesEvidence(evidence = {}) {
  return {
    id: String(evidence.id || "").trim(),
    request_id: String(evidence.request_id || evidence.requestId || "").trim(),
    order_id: String(evidence.order_id || evidence.orderId || "").trim(),
    object_key: String(evidence.object_key || evidence.objectKey || "").trim(),
    public_url: String(evidence.public_url || evidence.publicUrl || "").trim(),
    file_name: String(evidence.file_name || evidence.fileName || "").trim(),
    content_type: String(evidence.content_type || evidence.contentType || "").trim(),
    size_bytes: normalizeMoneyFen(evidence.size_bytes ?? evidence.sizeBytes),
    content_sha: String(evidence.content_sha || evidence.contentSha || "").trim(),
    uploaded_by_id: String(evidence.uploaded_by_id || evidence.uploadedById || "").trim(),
    uploaded_by_role: String(evidence.uploaded_by_role || evidence.uploadedByRole || "").trim(),
    status: String(evidence.status || "uploaded").trim() || "uploaded",
    created_at: String(evidence.created_at || evidence.createdAt || "").trim(),
    confirmed_at: String(evidence.confirmed_at || evidence.confirmedAt || "").trim()
  };
}

export function canCreateAfterSalesRequest(order = {}, request = {}) {
  const normalized = normalizeAfterSalesRequest(request);
  const status = String(order.status || "").trim();
  const amount = normalizeMoneyFen(order.amount_fen ?? order.amountFen);
  const allowedStatuses = new Set(["paid", "merchant_pending", "preparing", "dispatching", "rider_assigned", "picked_up", "delivering", "completed"]);
  const missing = [];
  if (!normalized.order_id && !order.id) missing.push("order_id");
  if (!normalized.reason) missing.push("reason");
  if (normalized.requested_amount_fen <= 0) missing.push("requested_amount_fen");
  if (normalized.requested_amount_fen > amount && amount > 0) missing.push("requested_amount_exceeds_order");
  return {
    ok: missing.length === 0 && allowedStatuses.has(status),
    missing_fields: missing,
    order_status_allowed: allowedStatuses.has(status)
  };
}

export function normalizeReview(review = {}) {
  return {
    id: String(review.id || "").trim(),
    target_type: String(review.target_type || review.targetType || "order").trim() || "order",
    target_id: String(review.target_id || review.targetId || "").trim(),
    user_id: String(review.user_id || review.userId || "").trim(),
    rating: Number.isFinite(Number(review.rating)) ? Math.max(1, Math.min(5, Math.trunc(Number(review.rating)))) : 5,
    content: String(review.content || "").trim().slice(0, 500),
    image_urls: normalizeStringList(review.image_urls || review.imageUrls),
    status: String(review.status || "published").trim() || "published"
  };
}

export function normalizeFavorite(favorite = {}) {
  return {
    user_id: String(favorite.user_id || favorite.userId || "").trim(),
    target_type: String(favorite.target_type || favorite.targetType || "shop").trim() || "shop",
    target_id: String(favorite.target_id || favorite.targetId || "").trim(),
    created_at: String(favorite.created_at || favorite.createdAt || "").trim()
  };
}

export function calculateDeliveryPromise(order = {}, deliveredAt = new Date()) {
  const deadline = Date.parse(order.promise_deadline_at || order.promiseDeadlineAt || "");
  const delivered = deliveredAt instanceof Date ? deliveredAt.getTime() : Date.parse(deliveredAt);
  if (order.timeout_exempt === true || order.timeoutExempt === true) {
    return { status: "exempt", timeout_seconds: 0, compensation_fen: 0 };
  }
  if (!Number.isFinite(deadline) || !Number.isFinite(delivered) || delivered <= deadline) {
    return { status: "on_time", timeout_seconds: 0, compensation_fen: 0 };
  }
  const timeoutSeconds = Math.floor((delivered - deadline) / 1000);
  const compensationFen = normalizeMoneyFen(order.timeout_compensation_fen ?? order.timeoutCompensationFen);
  return {
    status: "timeout",
    timeout_seconds: timeoutSeconds,
    compensation_fen: compensationFen
  };
}

export function normalizePointsTransaction(transaction = {}) {
  return {
    user_id: String(transaction.user_id || transaction.userId || "").trim(),
    type: String(transaction.type || "earn").trim() || "earn",
    points: Math.trunc(Number(transaction.points || 0)),
    source_id: String(transaction.source_id || transaction.sourceId || "").trim()
  };
}

export function applyPointsBalance(currentPoints = 0, transaction = {}) {
  const normalized = normalizePointsTransaction(transaction);
  const next = Math.max(0, Math.trunc(Number(currentPoints || 0)) + normalized.points);
  return {
    user_id: normalized.user_id,
    balance: next,
    transaction: normalized
  };
}

export function resolveMembershipTier(growthValue = 0) {
  const value = Math.max(0, Math.trunc(Number(growthValue || 0)));
  if (value >= 10000) return "black_gold";
  if (value >= 3000) return "gold";
  if (value >= 500) return "silver";
  return "none";
}

export function normalizePushDelivery(delivery = {}) {
  return {
    id: String(delivery.id || "").trim(),
    user_id: String(delivery.user_id || delivery.userId || "").trim(),
    channel: String(delivery.channel || "in_app").trim() || "in_app",
    template_key: String(delivery.template_key || delivery.templateKey || "").trim(),
    status: String(delivery.status || "queued").trim() || "queued",
    retry_count: normalizeMoneyFen(delivery.retry_count ?? delivery.retryCount)
  };
}

export function resolveRiskDecision(events = [], policy = {}) {
  const counts = new Map();
  for (const event of Array.isArray(events) ? events : []) {
    const type = String(event.type || "").trim();
    if (!type) continue;
    counts.set(type, (counts.get(type) || 0) + 1);
  }
  const abnormalOrderLimit = Number.isFinite(Number(policy.abnormal_order_limit ?? policy.abnormalOrderLimit))
    ? Number(policy.abnormal_order_limit ?? policy.abnormalOrderLimit)
    : 7;
  const maliciousRefundLimit = Number.isFinite(Number(policy.malicious_refund_limit ?? policy.maliciousRefundLimit))
    ? Number(policy.malicious_refund_limit ?? policy.maliciousRefundLimit)
    : 3;
  const blocked = (counts.get("abnormal_ordering") || 0) >= abnormalOrderLimit
    || (counts.get("malicious_refund") || 0) >= maliciousRefundLimit
    || (counts.get("fake_transaction") || 0) > 0;
  return {
    blocked,
    reasons: [...counts.entries()].filter(([, count]) => count > 0).map(([type]) => type)
  };
}

export function resolveRiderDepositRefund(deposit = {}, now = new Date()) {
  const normalized = normalizeDeposit(deposit);
  if (normalized.dispute_open) {
    return {
      refundable: false,
      reason: "DISPUTE_OPEN",
      eligible_at: ""
    };
  }
  const baseDates = [
    normalized.dispute_closed_at,
    normalized.last_order_completed_at,
    normalized.resignation_submitted_at
  ]
    .map((value) => Date.parse(value))
    .filter((value) => Number.isFinite(value));
  if (baseDates.length === 0) {
    return {
      refundable: false,
      reason: "MISSING_REFUND_BASE_TIME",
      eligible_at: ""
    };
  }
  const baseTime = Math.max(...baseDates);
  const eligibleAt = new Date(baseTime + 7 * 24 * 60 * 60 * 1000);
  return {
    refundable: now.getTime() >= eligibleAt.getTime(),
    reason: now.getTime() >= eligibleAt.getTime() ? "REFUNDABLE" : "WAITING_7_DAYS",
    eligible_at: eligibleAt.toISOString()
  };
}

function average(values = []) {
  if (values.length === 0) return 0;
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

export function normalizeMerchantQualification(qualification = {}) {
  return {
    id: String(qualification.id || "").trim(),
    type: String(qualification.type || "").trim(),
    file_url: String(qualification.file_url || qualification.fileUrl || "").trim(),
    expires_at: String(qualification.expires_at || qualification.expiresAt || "").trim(),
    status: String(qualification.status || "pending_review").trim() || "pending_review"
  };
}

export function isQualificationValid(qualification = {}, now = new Date()) {
  const normalized = normalizeMerchantQualification(qualification);
  if (!normalized.type || !normalized.file_url || !normalized.expires_at) {
    return false;
  }
  const expiresAt = Date.parse(normalized.expires_at);
  if (!Number.isFinite(expiresAt) || expiresAt <= now.getTime()) {
    return false;
  }
  return normalized.status === "approved";
}

export function resolveShopOperationState(shop = {}, qualifications = [], now = new Date()) {
  const profile = normalizeShopProfile(shop);
  const requiredTypes = ["business_license", "health_certificate"];
  const validTypes = new Set(
    qualifications
      .filter((qualification) => isQualificationValid(qualification, now))
      .map((qualification) => normalizeMerchantQualification(qualification).type)
  );
  const missingTypes = requiredTypes.filter((type) => !validTypes.has(type));
  if (missingTypes.length > 0) {
    return {
      status: "qualification_expired",
      shop_closed: true,
      popup_required: true,
      popup_code: "MERCHANT_QUALIFICATION_EXPIRED",
      missing_qualification_types: missingTypes,
      message: "店铺资质已过期或未审核通过，请补充营业执照和健康证有效期后提交审核。"
    };
  }
  if (profile.capabilities.length === 0) {
    return {
      status: "pending_review",
      shop_closed: true,
      popup_required: true,
      popup_code: "SHOP_CAPABILITY_REQUIRED",
      missing_qualification_types: [],
      message: "请先配置店铺经营能力。"
    };
  }
  return {
    status: "active",
    shop_closed: false,
    popup_required: false,
    popup_code: "",
    missing_qualification_types: [],
    message: ""
  };
}

export function normalizeMerchantStaff(staff = {}) {
  return {
    id: String(staff.id || "").trim(),
    merchant_id: String(staff.merchant_id || staff.merchantId || "").trim(),
    shop_id: String(staff.shop_id || staff.shopId || "").trim(),
    name: String(staff.name || "").trim(),
    phone: String(staff.phone || "").trim(),
    role: String(staff.role || "staff").trim() || "staff",
    status: String(staff.status || "active").trim() || "active",
    health_certificate_url: String(staff.health_certificate_url || staff.healthCertificateUrl || "").trim(),
    health_certificate_expires_at: String(staff.health_certificate_expires_at || staff.healthCertificateExpiresAt || "").trim()
  };
}

export function buildMerchantSupplementPayload(input = {}) {
  const staff = Array.isArray(input.staff) ? input.staff.map(normalizeMerchantStaff).filter((item) => item.name && item.phone) : [];
  const qualifications = Array.isArray(input.qualifications) ? input.qualifications.map(normalizeMerchantQualification) : [];
  return {
    merchant_id: String(input.merchant_id || input.merchantId || "").trim(),
    shop_id: String(input.shop_id || input.shopId || "").trim(),
    staff,
    qualifications,
    supplemental_documents: Array.isArray(input.supplemental_documents || input.supplementalDocuments)
      ? (input.supplemental_documents || input.supplementalDocuments).map((item) => ({
        title: String(item.title || "").trim(),
        file_url: String(item.file_url || item.fileUrl || "").trim(),
        expires_at: String(item.expires_at || item.expiresAt || "").trim()
      })).filter((item) => item.title && item.file_url)
      : []
  };
}

export function shopSupportsOrderType(shop = {}, orderType = "") {
  const profile = normalizeShopProfile(shop);
  const normalizedOrderType = String(orderType || "").trim();
  const requiredCapability = CAPABILITY_BY_ORDER_TYPE[normalizedOrderType];
  if (!requiredCapability) {
    return normalizedOrderType === "courier"
      || normalizedOrderType === "errand_buy"
      || normalizedOrderType === "errand_deliver"
      || normalizedOrderType === "errand_pickup"
      || normalizedOrderType === "errand_do";
  }
  if (!profile.capabilities.includes(requiredCapability)) {
    return false;
  }
  if (requiredCapability === "medicine") {
    return profile.account_type === "pharmacy"
      || profile.account_type === "clinic"
      || profile.category === "pharmacy"
      || profile.category === "clinic";
  }
  return true;
}

export function resolveFulfillmentMode(orderType = "") {
  const normalizedOrderType = String(orderType || "").trim();
  if (normalizedOrderType === "groupbuy") {
    return "in_store_redemption";
  }
  if (normalizedOrderType === "courier" || normalizedOrderType.startsWith("errand_")) {
    return "platform_errand";
  }
  return "rider_delivery";
}

export function buildShopDisplayPage(shop = {}, products = [], deals = []) {
  const profile = normalizeShopProfile(shop);
  return {
    id: profile.id,
    merchant_id: profile.merchant_id,
    name: profile.name,
    category: profile.category,
    capabilities: profile.capabilities,
    display: profile.display,
    takeout_enabled: profile.capabilities.includes("takeout"),
    groupbuy_enabled: profile.capabilities.includes("groupbuy"),
    medicine_enabled: profile.capabilities.includes("medicine"),
    products: Array.isArray(products) ? products : [],
    groupbuy_deals: Array.isArray(deals) ? deals : []
  };
}

export function normalizeMerchantProduct(product = {}) {
  const ingredients = Array.isArray(product.ingredients ?? product.ingredient_list ?? product.ingredientList)
    ? (product.ingredients ?? product.ingredient_list ?? product.ingredientList).map((item) => String(item).trim()).filter(Boolean)
    : [];
  return {
    id: String(product.id || "").trim(),
    shop_id: String(product.shop_id || product.shopId || "").trim(),
    name: String(product.name || "").trim(),
    image_url: String(product.image_url || product.imageUrl || "").trim(),
    description: String(product.description || "").trim(),
    ingredient_list: ingredients,
    price_fen: normalizeMoneyFen(product.price_fen ?? product.priceFen),
    stock_count: Number.isFinite(Number(product.stock_count ?? product.stockCount))
      ? Math.max(0, Math.trunc(Number(product.stock_count ?? product.stockCount)))
      : 0,
    status: String(product.status || "active").trim() || "active"
  };
}

export function resolveRefundDestination(config = {}, request = {}) {
  const requested = String(request.destination || request.refund_destination || request.refundDestination || "").trim();
  if (requested === "balance" || requested === "original_route") {
    return requested;
  }
  const strategy = String(config.default_refund_strategy || config.defaultRefundStrategy || "balance_first").trim();
  return strategy === "original_route_first" ? "original_route" : "balance";
}

export function resolveGroupbuyUnavailableRefund(deal = {}, voucher = {}, config = {}) {
  const product = normalizeMerchantProduct(deal);
  const voucherStatus = String(voucher.status || "").trim();
  const unavailable = product.status === "sold_out" || product.status === "removed" || product.stock_count <= 0;
  if (!unavailable || voucherStatus === "redeemed" || voucherStatus === "refunded") {
    return {
      refund_required: false,
      destination: "",
      reason: ""
    };
  }
  return {
    refund_required: true,
    destination: resolveRefundDestination(config),
    reason: "GROUPBUY_PRODUCT_UNAVAILABLE"
  };
}

export function normalizeWalletPaymentPasswordState(state = {}) {
  return {
    user_id: String(state.user_id || state.userId || "").trim(),
    status: String(state.status || "unset").trim() || "unset",
    failed_attempts: normalizeMoneyFen(state.failed_attempts ?? state.failedAttempts),
    locked_until: String(state.locked_until || state.lockedUntil || "").trim()
  };
}

export function canUseBalancePaymentPassword(state = {}, now = new Date()) {
  const normalized = normalizeWalletPaymentPasswordState(state);
  if (normalized.status !== "set") return false;
  if (!normalized.locked_until) return true;
  const lockedUntil = Date.parse(normalized.locked_until);
  return Number.isFinite(lockedUntil) ? lockedUntil <= now.getTime() : true;
}

export function redeemGroupbuyVoucher(voucher = {}, input = {}, now = new Date()) {
  const status = String(voucher.status || "").trim();
  const method = String(input.method || "").trim();
  if (status !== "issued") {
    return { ok: false, code: "VOUCHER_NOT_REDEEMABLE" };
  }
  if (method !== "qr_scan") {
    return { ok: false, code: "QR_SCAN_REQUIRED" };
  }
  const shopId = String(voucher.shop_id || voucher.shopId || "").trim();
  const scanShopId = String(input.shop_id || input.shopId || "").trim();
  if (shopId && scanShopId && shopId !== scanShopId) {
    return { ok: false, code: "SHOP_MISMATCH" };
  }
  return {
    ok: true,
    voucher: {
      ...voucher,
      status: "redeemed",
      redeemed_at: now.toISOString(),
      redemption_method: "qr_scan"
    }
  };
}

export function buildOfficialGroupMembership(user = {}, group = {}) {
  return {
    group_id: String(group.id || group.group_id || group.groupId || "official").trim(),
    user_id: String(user.id || user.user_id || user.userId || "").trim(),
    role: "member",
    notification: "muted",
    joined_reason: "auto_join_on_registration"
  };
}

export function canClaimCouponWithGroup(coupon = {}, userGroupIds = []) {
  const requirement = String(coupon.claim_requirement || coupon.claimRequirement || "none").trim();
  if (requirement !== "group_membership") {
    return true;
  }
  const requiredGroupId = String(coupon.required_group_id || coupon.requiredGroupId || "").trim();
  return requiredGroupId !== "" && userGroupIds.map(String).includes(requiredGroupId);
}

export function normalizeGroupChat(group = {}) {
  return {
    id: String(group.id || "").trim(),
    owner_id: String(group.owner_id || group.ownerId || "").trim(),
    owner_role: String(group.owner_role || group.ownerRole || "").trim(),
    type: String(group.type || "merchant").trim() || "merchant",
    name: String(group.name || "").trim(),
    notification_default: String(group.notification_default || group.notificationDefault || "normal").trim() || "normal"
  };
}

export function normalizeRedPacket(packet = {}) {
  return {
    id: String(packet.id || "").trim(),
    sender_id: String(packet.sender_id || packet.senderId || "").trim(),
    sender_role: String(packet.sender_role || packet.senderRole || "").trim(),
    scene: String(packet.scene || "group_chat").trim() || "group_chat",
    target_id: String(packet.target_id || packet.targetId || "").trim(),
    type: String(packet.type || "fixed").trim() || "fixed",
    total_amount_fen: normalizeMoneyFen(packet.total_amount_fen ?? packet.totalAmountFen),
    quantity: Math.max(0, Math.trunc(Number(packet.quantity ?? 0))),
    payment_method: "balance"
  };
}

export function normalizeCircleFeatureConfig(config = {}) {
  const enabled = config.enabled !== false;
  const rawMode = String(config.mode || (enabled ? "wall_only" : "disabled")).trim();
  const mode = enabled && ["wall_only", "circle_and_meal_match"].includes(rawMode) ? rawMode : "disabled";
  return {
    enabled: mode !== "disabled",
    mode,
    home_module_enabled: config.home_module_enabled !== false && mode !== "disabled",
    meal_match_enabled: mode === "circle_and_meal_match" && config.meal_match_enabled !== false,
    moderation_required: config.moderation_required !== false,
    source_reference: String(config.source_reference || config.sourceReference || "InfiniLink-reference").trim()
  };
}

export function normalizeCirclePost(post = {}) {
  const type = String(post.type || "text").trim();
  return {
    id: String(post.id || "").trim(),
    author_user_id: String(post.author_user_id || post.authorUserId || "").trim(),
    circle_id: String(post.circle_id || post.circleId || "micro_wall").trim() || "micro_wall",
    type: ["text", "image", "food_invite"].includes(type) ? type : "text",
    content: String(post.content || "").trim(),
    image_urls: Array.isArray(post.image_urls || post.imageUrls)
      ? (post.image_urls || post.imageUrls).map((item) => String(item).trim()).filter(Boolean)
      : [],
    status: String(post.status || "pending_review").trim() || "pending_review",
    tags: Array.isArray(post.tags) ? post.tags.map((item) => String(item).trim()).filter(Boolean) : [],
    created_at: String(post.created_at || post.createdAt || "").trim()
  };
}

export function normalizeMealMatchProfile(profile = {}) {
  return {
    user_id: String(profile.user_id || profile.userId || profile.id || "").trim(),
    gender: String(profile.gender || "").trim(),
    identity_truth_signed: profile.identity_truth_signed === true || profile.identityTruthSigned === true,
    platform_liability_release_signed: profile.platform_liability_release_signed === true || profile.platformLiabilityReleaseSigned === true,
    questionnaire_completed: profile.questionnaire_completed === true || profile.questionnaireCompleted === true,
    personality_traits: normalizeStringList(profile.personality_traits || profile.personalityTraits),
    dietary_habits: normalizeStringList(profile.dietary_habits || profile.dietaryHabits)
  };
}

export function canUseMealMatch(profile = {}) {
  const normalized = normalizeMealMatchProfile(profile);
  const missing = [];
  if (!normalized.gender) missing.push("gender");
  if (!normalized.identity_truth_signed) missing.push("identity_truth_signed");
  if (!normalized.platform_liability_release_signed) missing.push("platform_liability_release_signed");
  if (!normalized.questionnaire_completed || normalized.personality_traits.length === 0 || normalized.dietary_habits.length === 0) {
    missing.push("questionnaire_completed");
  }
  return {
    ok: missing.length === 0,
    missing_requirements: missing
  };
}

export function rankMealBuddyCandidates(profile = {}, candidates = []) {
  const user = normalizeMealMatchProfile(profile);
  if (!canUseMealMatch(user).ok) return [];
  const personality = new Set(user.personality_traits);
  const dietary = new Set(user.dietary_habits);
  return (Array.isArray(candidates) ? candidates : [])
    .map(normalizeMealMatchProfile)
    .filter((candidate) => candidate.user_id && candidate.user_id !== user.user_id && canUseMealMatch(candidate).ok)
    .map((candidate) => {
      const personalityMatches = candidate.personality_traits.filter((item) => personality.has(item)).length;
      const dietaryMatches = candidate.dietary_habits.filter((item) => dietary.has(item)).length;
      return {
        ...candidate,
        match_score: dietaryMatches * 60 + personalityMatches * 40,
        matched_personality_traits: candidate.personality_traits.filter((item) => personality.has(item)),
        matched_dietary_habits: candidate.dietary_habits.filter((item) => dietary.has(item))
      };
    })
    .sort((left, right) => right.match_score - left.match_score || left.user_id.localeCompare(right.user_id));
}

export function normalizeCouponPolicy(coupon = {}) {
  const issuer = String(coupon.issuer_type || coupon.issuerType || "merchant").trim();
  const requestedBearer = String(coupon.cost_bearer || coupon.costBearer || "").trim();
  const scope = String(coupon.scope_type || coupon.scopeType || "single_shop").trim();
  const issuerType = issuer === "platform" ? "platform" : "merchant";
  const costBearer = requestedBearer === "platform" && issuerType === "platform" ? "platform" : "merchant";
  return {
    id: String(coupon.id || "").trim(),
    issuer_type: issuerType,
    cost_bearer: costBearer,
    subsidy_settlement_required: issuerType === "platform" && costBearer === "platform",
    merchant_activity_acceptance_required: issuerType === "platform" && costBearer === "merchant",
    merchant_acceptance_status: String(coupon.merchant_acceptance_status || coupon.merchantAcceptanceStatus || "pending").trim() || "pending",
    scope_type: scope === "participating_shops" ? "participating_shops" : "single_shop",
    shop_id: String(coupon.shop_id || coupon.shopId || "").trim(),
    participating_shop_ids: normalizeStringList(coupon.participating_shop_ids || coupon.participatingShopIds),
    amount_fen: normalizeMoneyFen(coupon.amount_fen ?? coupon.amountFen)
  };
}

export function canCouponApplyToShop(coupon = {}, shopId = "") {
  const normalized = normalizeCouponPolicy(coupon);
  const targetShopId = String(shopId || "").trim();
  if (normalized.merchant_activity_acceptance_required && normalized.merchant_acceptance_status !== "accepted") {
    return false;
  }
  if (normalized.scope_type === "single_shop") {
    return normalized.shop_id !== "" && normalized.shop_id === targetShopId;
  }
  return normalized.participating_shop_ids.includes(targetShopId);
}

function normalizeStringList(input = []) {
  return Array.isArray(input) ? input.map((item) => String(item).trim()).filter(Boolean) : [];
}

export function allocateRedPacketShares(packet = {}) {
  const normalized = normalizeRedPacket(packet);
  if (!normalized.sender_id || !normalized.target_id || normalized.total_amount_fen <= 0 || normalized.quantity <= 0) {
    return { ok: false, code: "INVALID_RED_PACKET", shares: [] };
  }
  if (normalized.quantity > normalized.total_amount_fen) {
    return { ok: false, code: "AMOUNT_TOO_SMALL", shares: [] };
  }
  if (normalized.type === "fixed") {
    const base = Math.floor(normalized.total_amount_fen / normalized.quantity);
    const remainder = normalized.total_amount_fen - base * normalized.quantity;
    return {
      ok: true,
      code: "",
      shares: Array.from({ length: normalized.quantity }, (_, index) => base + (index < remainder ? 1 : 0))
    };
  }
  return {
    ok: true,
    code: "",
    shares: allocateRandomRedPacketShares(normalized.total_amount_fen, normalized.quantity, normalized.id || `${normalized.sender_id}:${normalized.target_id}`)
  };
}

function allocateRandomRedPacketShares(totalAmountFen, quantity, seedText) {
  let remaining = totalAmountFen;
  let seed = 0;
  for (const char of seedText) {
    seed = (seed * 31 + char.charCodeAt(0)) >>> 0;
  }
  const shares = [];
  for (let index = 0; index < quantity - 1; index += 1) {
    const remainingRecipients = quantity - index;
    const max = remaining - (remainingRecipients - 1);
    seed = (seed * 1664525 + 1013904223) >>> 0;
    const amount = 1 + (seed % Math.max(1, max));
    shares.push(amount);
    remaining -= amount;
  }
  shares.push(remaining);
  return shares;
}
