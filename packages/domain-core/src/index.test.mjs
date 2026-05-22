import assert from "node:assert/strict";
import test from "node:test";
import {
  buildMerchantSupplementPayload,
  allocateRedPacketShares,
  buildCartSummary,
  buildControlledHomeCards,
  buildShopDisplayPage,
  buildHomeModules,
  buildOfficialGroupMembership,
  createConsumedEventLedger,
  createIdempotentConsumer,
  canUseMerchantInvite,
  canUseRiderInvite,
  canUseFreeDispatchCancel,
  canClaimCouponWithGroup,
  canMerchantAcceptOrders,
  canRiderAcceptOrders,
  canStationManagerManualDispatch,
  canDeclineDispatchWithoutPenalty,
  canCouponApplyToShop,
  canCreateAfterSalesRequest,
  canCheckoutCart,
  canUseAddressForDelivery,
  canUseBalancePaymentPassword,
  canUseMealMatch,
  calculateDeliveryPromise,
  calculateOrderPayable,
  consumeFreeDispatchCancel,
  evaluateRiderLevel,
  applyPointsBalance,
  isQualificationValid,
  normalizeAfterSalesEvidence,
  normalizeAfterSalesRequest,
  normalizeCartItem,
  normalizeFavorite,
  normalizeCircleFeatureConfig,
  normalizeCirclePost,
  normalizeCouponPolicy,
  normalizeHomeCard,
  normalizeMoneyFen,
  normalizeMerchantQualification,
  normalizeMerchantProduct,
  normalizeMerchantStaff,
  normalizeGroupChat,
  normalizeOrderOptions,
  normalizeOutboxConsumerEvent,
  normalizePushDelivery,
  normalizeRedPacket,
  normalizeReview,
  normalizeShopProfile,
  normalizeUserAddress,
  redeemGroupbuyVoucher,
  rankRidersForPriorityDispatch,
  rankMealBuddyCandidates,
  resolveGroupbuyUnavailableRefund,
  resolveDispatchMode,
  resolveRiderDepositRefund,
  resolveRefundDestination,
  resolveMembershipTier,
  resolveRiskDecision,
  resolveShopOperationState,
  resolveFulfillmentMode,
  normalizeStationTaskConfig,
  selectNextOnlineRider,
  shopSupportsOrderType,
  upsertCartItem
} from "./index.mjs";

test("money is normalized to non-negative integer fen", () => {
  assert.equal(normalizeMoneyFen(100.9), 100);
  assert.equal(normalizeMoneyFen(-1), 0);
  assert.equal(normalizeMoneyFen("bad"), 0);
});

test("outbox consumer events normalize durable idempotency keys", () => {
  const normalized = normalizeOutboxConsumerEvent({
    id: "obe_1",
    topic: "order.paid",
    aggregate_type: "order",
    aggregate_id: "ord_1",
    event_type: "order.paid",
    idempotency_key: "order:ord_1:paid"
  }, "dispatch-worker", { now: new Date("2026-05-22T12:00:00.000Z") });
  assert.deepEqual(normalized, {
    consumer_name: "dispatch-worker",
    consumer_event_key: "dispatch-worker:order:ord_1:paid",
    idempotency_key: "order:ord_1:paid",
    outbox_event_id: "obe_1",
    topic: "order.paid",
    aggregate_type: "order",
    aggregate_id: "ord_1",
    event_type: "order.paid",
    first_seen_at: "2026-05-22T12:00:00.000Z",
    status: "processing",
    attempts: 0,
    last_error: ""
  });
  assert.equal(normalizeOutboxConsumerEvent({ topic: "order.paid", aggregate_id: "ord_2", event_type: "order.paid" }, "notification-worker").consumer_event_key, "notification-worker:order.paid:ord_2:order.paid");
});

test("idempotent consumer processes each outbox event once per worker", async () => {
  let handled = 0;
  const ledger = createConsumedEventLedger();
  const consumer = createIdempotentConsumer({
    consumerName: "payment-worker",
    ledger,
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { transaction_id: "wx_1" };
    }
  });
  const event = { id: "obe_1", topic: "payment.wechat.callback", idempotency_key: "wechat:wx_1" };
  const first = await consumer(event);
  const duplicate = await consumer({ ...event, id: "obe_1_replayed" });
  assert.equal(first.status, "processed");
  assert.equal(first.consumed, true);
  assert.equal(duplicate.status, "duplicate");
  assert.equal(duplicate.consumed, false);
  assert.equal(handled, 1);
  assert.equal(ledger.snapshot().length, 1);
  assert.equal(ledger.snapshot()[0].status, "processed");
});

test("home modules are configurable and ordered", () => {
  assert.deepEqual(buildHomeModules([
    { key: "circle", title: "圈子", route: "/circle", sort_order: 25 },
    { key: "medicine", title: "买药", route: "/medicine", sort_order: 30 },
    { key: "takeout", title: "外卖", route: "/takeout", sort_order: 10 },
    { key: "hidden", title: "隐藏", route: "/hidden", enabled: false }
  ]).map((item) => item.key), ["takeout", "circle", "medicine"]);
});

test("homepage cards are controlled by backend config and schedule", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  const card = normalizeHomeCard({
    id: "card_1",
    type: "product",
    targetId: "product_1",
    title: "招牌牛肉饭",
    priceFen: 2599
  });
  assert.equal(card.type, "product");
  assert.equal(card.price_fen, 2599);
  const cards = buildControlledHomeCards([
    card,
    { id: "future", type: "coupon", target_id: "coupon_1", title: "未开始", starts_at: "2026-05-22T00:00:00.000Z" },
    { id: "disabled", type: "shop", target_id: "shop_1", title: "隐藏店", enabled: false }
  ], now);
  assert.deepEqual(cards.map((item) => item.id), ["card_1"]);
});

test("legacy delivery address order options and payable amount are normalized", () => {
  const address = normalizeUserAddress({
    userId: "user_1",
    contactName: " 张三 ",
    contactPhone: "13800000000",
    city: "北京",
    detail: "望京SOHO",
    latitude: 39.99,
    longitude: 116.48,
    tag: "company",
    isDefault: true
  });
  assert.equal(address.contact_name, "张三");
  assert.equal(canUseAddressForDelivery(address).ok, true);
  assert.deepEqual(canUseAddressForDelivery({ contact_name: "李四" }).missing_fields, ["contact_phone", "city", "detail", "geo_location"]);

  assert.deepEqual(normalizeOrderOptions({
    remark: "少放辣",
    tablewareCount: 200,
    contactlessDelivery: true,
    invoiceRequested: true
  }), {
    remark: "少放辣",
    tableware_count: 99,
    contactless_delivery: true,
    invoice_requested: true
  });

  assert.deepEqual(calculateOrderPayable({
    itemsTotalFen: 3000,
    deliveryFeeFen: 500,
    packagingFeeFen: 200,
    discountFen: 800
  }), {
    items_total_fen: 3000,
    delivery_fee_fen: 500,
    packaging_fee_fen: 200,
    discount_fen: 800,
    payable_fen: 2900
  });
});

test("cart items can be upserted summarized and checked out with address", () => {
  const item = normalizeCartItem({
    userId: "user_1",
    shopId: "shop_1",
    productId: "prod_beef",
    productName: "牛肉饭",
    unitPriceFen: 2599,
    quantity: 2
  });
  assert.equal(item.unit_price_fen, 2599);
  const cart = upsertCartItem([], item);
  assert.equal(cart.length, 1);
  const updated = upsertCartItem(cart, { ...item, quantity: 3 });
  assert.equal(updated[0].quantity, 3);
  const summary = buildCartSummary(updated, { deliveryFeeFen: 300, packagingFeeFen: 100 });
  assert.equal(summary.items_total_fen, 7797);
  assert.equal(summary.payable_fen, 8197);
  assert.equal(canCheckoutCart(summary, {
    contact_name: "张三",
    contact_phone: "13800000000",
    city: "北京",
    detail: "望京SOHO",
    latitude: 39.99,
    longitude: 116.48
  }).ok, true);
  assert.deepEqual(canCheckoutCart({ items: [], payable_fen: 0 }, {}).missing_fields, [
    "cart_items",
    "payable_fen",
    "contact_name",
    "contact_phone",
    "city",
    "detail",
    "geo_location"
  ]);
});

test("dispatch free cancel is allowed once per day", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  assert.equal(canUseFreeDispatchCancel({}, now), true);
  const used = consumeFreeDispatchCancel({}, now);
  assert.deepEqual(used, { allowed: true, used_on: "2026-05-21" });
  assert.equal(canUseFreeDispatchCancel(used, now), false);
});

test("shop profile separates merchant account from shop capabilities", () => {
  const shop = normalizeShopProfile({
    id: " shop_1 ",
    merchantId: " merchant_1 ",
    name: " 蓝海餐厅 ",
    category: "restaurant",
    accountType: "standard",
    capabilities: ["groupbuy", "takeout", "takeout"],
    display: { coverUrl: "https://cdn/cover.jpg" }
  });
  assert.deepEqual(shop.capabilities, ["groupbuy", "takeout"]);
  assert.equal(shopSupportsOrderType(shop, "takeout"), true);
  assert.equal(shopSupportsOrderType(shop, "groupbuy"), true);
  assert.equal(shopSupportsOrderType(shop, "medicine"), false);
});

test("medicine shops require pharmacy or clinic attributes", () => {
  assert.equal(shopSupportsOrderType({
    account_type: "standard",
    category: "restaurant",
    capabilities: ["medicine"]
  }, "medicine"), false);
  assert.equal(shopSupportsOrderType({
    account_type: "clinic",
    category: "clinic",
    capabilities: ["medicine", "groupbuy"]
  }, "medicine"), true);
});

test("fulfillment modes match business flow", () => {
  assert.equal(resolveFulfillmentMode("takeout"), "rider_delivery");
  assert.equal(resolveFulfillmentMode("medicine"), "rider_delivery");
  assert.equal(resolveFulfillmentMode("courier"), "platform_errand");
  assert.equal(resolveFulfillmentMode("errand_buy"), "platform_errand");
  assert.equal(resolveFulfillmentMode("groupbuy"), "in_store_redemption");
});

test("shop display page exposes takeout and groupbuy tabs", () => {
  const page = buildShopDisplayPage(
    { id: "s1", merchant_id: "m1", name: "店铺", capabilities: ["takeout", "groupbuy"] },
    [{ id: "p1" }],
    [{ id: "d1" }]
  );
  assert.equal(page.takeout_enabled, true);
  assert.equal(page.groupbuy_enabled, true);
  assert.equal(page.products.length, 1);
  assert.equal(page.groupbuy_deals.length, 1);
});

test("groupbuy voucher redemption requires in-store qr scan", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  assert.deepEqual(redeemGroupbuyVoucher({ status: "issued", shop_id: "s1" }, { method: "manual", shop_id: "s1" }, now), {
    ok: false,
    code: "QR_SCAN_REQUIRED"
  });
  const result = redeemGroupbuyVoucher({ id: "v1", status: "issued", shop_id: "s1" }, { method: "qr_scan", shop_id: "s1" }, now);
  assert.equal(result.ok, true);
  assert.equal(result.voucher.status, "redeemed");
  assert.equal(result.voucher.redemption_method, "qr_scan");
});

test("merchant registration requires active admin invite", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  assert.equal(canUseMerchantInvite({
    token: "invite_1",
    type: "merchant",
    status: "active",
    expires_at: "2026-05-22T00:00:00.000Z",
    created_by_admin_id: "admin_1"
  }, now), true);
  assert.equal(canUseMerchantInvite({
    token: "invite_1",
    type: "merchant",
    status: "active",
    expires_at: "2026-05-20T00:00:00.000Z",
    created_by_admin_id: "admin_1"
  }, now), false);
  assert.equal(canUseMerchantInvite({
    token: "invite_1",
    type: "merchant",
    status: "active"
  }, now), false);
});

test("merchant qualifications require files, expiry dates and approval", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  const qualification = normalizeMerchantQualification({
    type: "business_license",
    fileUrl: "https://cdn/license.jpg",
    expiresAt: "2027-01-01T00:00:00.000Z",
    status: "approved"
  });
  assert.equal(qualification.file_url, "https://cdn/license.jpg");
  assert.equal(isQualificationValid(qualification, now), true);
  assert.equal(isQualificationValid({ ...qualification, expires_at: "2026-01-01T00:00:00.000Z" }, now), false);
  assert.equal(isQualificationValid({ ...qualification, status: "pending_review" }, now), false);
});

test("expired or missing merchant qualifications close the shop and require popup", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  const state = resolveShopOperationState(
    { id: "s1", capabilities: ["takeout"] },
    [{ type: "business_license", file_url: "https://cdn/license.jpg", expires_at: "2027-01-01T00:00:00.000Z", status: "approved" }],
    now
  );
  assert.equal(state.status, "qualification_expired");
  assert.equal(state.shop_closed, true);
  assert.equal(state.popup_required, true);
  assert.deepEqual(state.missing_qualification_types, ["health_certificate"]);

  const active = resolveShopOperationState(
    { id: "s1", capabilities: ["takeout"] },
    [
      { type: "business_license", file_url: "https://cdn/license.jpg", expires_at: "2027-01-01T00:00:00.000Z", status: "approved" },
      { type: "health_certificate", file_url: "https://cdn/health.jpg", expires_at: "2027-01-01T00:00:00.000Z", status: "approved" }
    ],
    now
  );
  assert.equal(active.status, "active");
  assert.equal(active.shop_closed, false);
});

test("merchant staff and supplemental materials are normalized", () => {
  const staff = normalizeMerchantStaff({
    merchantId: "m1",
    shopId: "s1",
    name: " 张三 ",
    phone: " 13800000000 ",
    healthCertificateExpiresAt: "2027-01-01"
  });
  assert.equal(staff.name, "张三");
  assert.equal(staff.health_certificate_expires_at, "2027-01-01");

  const payload = buildMerchantSupplementPayload({
    merchantId: "m1",
    shopId: "s1",
    staff: [staff, { name: "", phone: "1" }],
    qualifications: [{ type: "business_license", fileUrl: "https://cdn/license.jpg", expiresAt: "2027-01-01" }],
    supplementalDocuments: [{ title: "门头照", fileUrl: "https://cdn/front.jpg" }]
  });
  assert.equal(payload.staff.length, 1);
  assert.equal(payload.qualifications[0].file_url, "https://cdn/license.jpg");
  assert.equal(payload.supplemental_documents[0].title, "门头照");
});

test("rider and station manager accounts are invite-only", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  assert.equal(canUseRiderInvite({
    token: "r1",
    type: "rider",
    status: "active",
    created_by_admin_id: "station_owner",
    expires_at: "2026-05-22T00:00:00.000Z"
  }, now), true);
  assert.equal(canUseRiderInvite({
    token: "s1",
    type: "station_manager",
    status: "active",
    created_by_admin_id: "admin_1",
    expires_at: "2026-05-22T00:00:00.000Z"
  }, now), true);
  assert.equal(canUseRiderInvite({ token: "bad", type: "rider", status: "active" }, now), false);
});

test("station manager can manually dispatch and configure task duration", () => {
  assert.equal(canStationManagerManualDispatch({ account_type: "station_manager", status: "active" }), true);
  assert.equal(canStationManagerManualDispatch({ account_type: "rider", status: "active" }), false);
  assert.deepEqual(normalizeStationTaskConfig({
    stationId: "station_1",
    configuredByStationManagerId: "manager_1",
    dailyTaskDurationMinutes: 360
  }), {
    station_id: "station_1",
    configured_by_station_manager_id: "manager_1",
    daily_task_duration_minutes: 360
  });
});

test("deposits gate rider and merchant order acceptance", () => {
  assert.equal(canRiderAcceptOrders(
    { account_type: "rider", status: "active", online: true },
    { subject_type: "rider", amount_fen: 5000, status: "paid" }
  ), true);
  assert.equal(canRiderAcceptOrders(
    { account_type: "rider", status: "active", online: true },
    { subject_type: "rider", amount_fen: 0, status: "wechat_exempt_approved" }
  ), true);
  assert.equal(canMerchantAcceptOrders(
    { status: "active" },
    { subject_type: "merchant", amount_fen: 5000, status: "wechat_exempt_approved" }
  ), false);
  assert.equal(canMerchantAcceptOrders(
    { status: "active" },
    { subject_type: "merchant", amount_fen: 5000, status: "paid" }
  ), true);
});

test("dispatch mode switches from grab hall to auto assign after ten minutes", () => {
  const now = new Date("2026-05-21T08:10:00.000Z");
  assert.equal(resolveDispatchMode({ created_at: "2026-05-21T08:00:01.000Z" }, now), "grab_hall");
  assert.equal(resolveDispatchMode({ created_at: "2026-05-21T08:00:00.000Z" }, now), "auto_assign");
});

test("rejected dispatch is assigned to next online rider", () => {
  const next = selectNextOnlineRider([
    { id: "rider_1", online: true, capacity: 1, distance_meters: 100, dispatch_score: 99 },
    { id: "rider_2", online: true, capacity: 1, distance_meters: 200, dispatch_score: 90 },
    { id: "rider_3", online: false, capacity: 1, distance_meters: 10, dispatch_score: 100 }
  ], ["rider_1"]);
  assert.equal(next.id, "rider_2");
});

test("rider deposit refund waits one week after resignation last order or dispute close", () => {
  const early = resolveRiderDepositRefund({
    last_order_completed_at: "2026-05-21T08:00:00.000Z",
    resignation_submitted_at: "2026-05-22T08:00:00.000Z"
  }, new Date("2026-05-28T07:59:59.000Z"));
  assert.equal(early.refundable, false);
  assert.equal(early.eligible_at, "2026-05-29T08:00:00.000Z");

  const disputed = resolveRiderDepositRefund({
    last_order_completed_at: "2026-05-21T08:00:00.000Z",
    resignation_submitted_at: "2026-05-22T08:00:00.000Z",
    dispute_closed_at: "2026-06-01T08:00:00.000Z"
  }, new Date("2026-06-08T08:00:00.000Z"));
  assert.equal(disputed.refundable, true);
  assert.equal(disputed.eligible_at, "2026-06-08T08:00:00.000Z");

  assert.equal(resolveRiderDepositRefund({ dispute_open: true }, new Date("2026-06-08T08:00:00.000Z")).reason, "DISPUTE_OPEN");
});

test("rider level is evaluated against team performance and used for priority dispatch", () => {
  const team = [
    { rider_id: "r1", station_id: "s1", accepted_order_count: 30, completed_order_count: 30, active_days: 3, average_accept_seconds: 20 },
    { rider_id: "r2", station_id: "s1", accepted_order_count: 18, completed_order_count: 15, active_days: 3, average_accept_seconds: 60 },
    { rider_id: "r3", station_id: "s1", accepted_order_count: 10, completed_order_count: 8, active_days: 3, average_accept_seconds: 90 }
  ];
  const level = evaluateRiderLevel(team[0], team);
  assert.equal(level.level, "S");
  assert.equal(level.dispatch_priority, 400);
  assert.equal(level.average_daily_orders, 10);

  const ranked = rankRidersForPriorityDispatch([
    { id: "r1", online: true, dispatch_priority: 400, average_accept_seconds: 20, distance_meters: 500 },
    { id: "r2", online: true, dispatch_priority: 300, average_accept_seconds: 15, distance_meters: 100 },
    { id: "r3", online: true, dispatch_priority: 400, average_accept_seconds: 30, distance_meters: 50 }
  ]);
  assert.deepEqual(ranked.map((rider) => rider.id), ["r1", "r3", "r2"]);
});

test("daily fixed order quota allows penalty-free decline after completion", () => {
  assert.equal(canDeclineDispatchWithoutPenalty({ completed_order_count: 12 }, { fixed_daily_order_count: 12 }), true);
  assert.equal(canDeclineDispatchWithoutPenalty({ completed_order_count: 11 }, { fixed_daily_order_count: 12 }), false);
  assert.equal(canDeclineDispatchWithoutPenalty({ completed_order_count: 99 }, { fixed_daily_order_count: 0 }), false);
});

test("merchant product details include Meituan-like content fields", () => {
  const product = normalizeMerchantProduct({
    id: "p1",
    shopId: "s1",
    name: "牛肉饭",
    imageUrl: "https://cdn/beef.jpg",
    description: "招牌套餐",
    ingredientList: [" 牛肉 ", "米饭", ""],
    priceFen: 2599,
    stockCount: 3
  });
  assert.deepEqual(product.ingredient_list, ["牛肉", "米饭"]);
  assert.equal(product.image_url, "https://cdn/beef.jpg");
  assert.equal(product.description, "招牌套餐");
});

test("refund destination is configurable and groupbuy unavailable refunds to balance by default", () => {
  assert.equal(resolveRefundDestination({}, {}), "balance");
  assert.equal(resolveRefundDestination({ default_refund_strategy: "original_route_first" }, {}), "original_route");
  assert.equal(resolveRefundDestination({ default_refund_strategy: "balance_first" }, { destination: "original_route" }), "original_route");

  assert.deepEqual(resolveGroupbuyUnavailableRefund(
    { id: "deal_1", status: "sold_out", stock_count: 0 },
    { id: "voucher_1", status: "issued" },
    {}
  ), {
    refund_required: true,
    destination: "balance",
    reason: "GROUPBUY_PRODUCT_UNAVAILABLE"
  });
});

test("balance payment requires configured payment password", () => {
  const now = new Date("2026-05-21T08:00:00.000Z");
  assert.equal(canUseBalancePaymentPassword({ status: "unset" }, now), false);
  assert.equal(canUseBalancePaymentPassword({ status: "set" }, now), true);
  assert.equal(canUseBalancePaymentPassword({ status: "set", locked_until: "2026-05-21T09:00:00.000Z" }, now), false);
});

test("official group auto-joins muted and coupons can require group membership", () => {
  const membership = buildOfficialGroupMembership({ id: "user_1" }, { id: "official_1" });
  assert.deepEqual(membership, {
    group_id: "official_1",
    user_id: "user_1",
    role: "member",
    notification: "muted",
    joined_reason: "auto_join_on_registration"
  });
  assert.equal(canClaimCouponWithGroup({ claim_requirement: "group_membership", required_group_id: "g1" }, ["g1"]), true);
  assert.equal(canClaimCouponWithGroup({ claim_requirement: "group_membership", required_group_id: "g1" }, ["g2"]), false);
  assert.equal(normalizeGroupChat({ id: "g1", ownerId: "m1", ownerRole: "merchant", name: "粉丝群" }).type, "merchant");
});

test("balance red packets support fixed and random shares in group or direct message", () => {
  const packet = normalizeRedPacket({
    id: "rp1",
    senderId: "merchant_1",
    senderRole: "merchant",
    scene: "group_chat",
    targetId: "group_1",
    type: "fixed",
    totalAmountFen: 100,
    quantity: 3
  });
  assert.equal(packet.payment_method, "balance");
  assert.deepEqual(allocateRedPacketShares(packet).shares, [34, 33, 33]);

  const random = allocateRedPacketShares({
    id: "rp_random",
    sender_id: "user_1",
    sender_role: "user",
    scene: "direct_message",
    target_id: "rider_1",
    type: "random",
    total_amount_fen: 100,
    quantity: 4
  });
  assert.equal(random.ok, true);
  assert.equal(random.shares.length, 4);
  assert.equal(random.shares.reduce((sum, value) => sum + value, 0), 100);
});

test("circle feature is backend controlled and posts are lightweight micro wall items", () => {
  assert.deepEqual(normalizeCircleFeatureConfig({ enabled: true, mode: "circle_and_meal_match" }), {
    enabled: true,
    mode: "circle_and_meal_match",
    home_module_enabled: true,
    meal_match_enabled: true,
    moderation_required: true,
    source_reference: "InfiniLink-reference"
  });
  const post = normalizeCirclePost({
    id: "post_1",
    authorUserId: "user_1",
    type: "food_invite",
    content: "今晚找饭搭",
    imageUrls: [" https://cdn/food.jpg "],
    tags: ["火锅", ""]
  });
  assert.equal(post.circle_id, "micro_wall");
  assert.equal(post.status, "pending_review");
  assert.deepEqual(post.tags, ["火锅"]);
});

test("meal buddy matching requires gender truth agreement release and questionnaire", () => {
  assert.deepEqual(canUseMealMatch({ gender: "female" }), {
    ok: false,
    missing_requirements: [
      "identity_truth_signed",
      "platform_liability_release_signed",
      "questionnaire_completed"
    ]
  });

  const profile = {
    user_id: "u1",
    gender: "female",
    identity_truth_signed: true,
    platform_liability_release_signed: true,
    questionnaire_completed: true,
    personality_traits: ["安静", "准时"],
    dietary_habits: ["不吃辣", "爱面食"]
  };
  assert.equal(canUseMealMatch(profile).ok, true);
  const ranked = rankMealBuddyCandidates(profile, [
    {
      user_id: "u2",
      gender: "male",
      identity_truth_signed: true,
      platform_liability_release_signed: true,
      questionnaire_completed: true,
      personality_traits: ["安静"],
      dietary_habits: ["不吃辣", "爱面食"]
    },
    {
      user_id: "u3",
      gender: "female",
      identity_truth_signed: true,
      platform_liability_release_signed: true,
      questionnaire_completed: true,
      personality_traits: ["外向"],
      dietary_habits: ["爱甜品"]
    }
  ]);
  assert.equal(ranked[0].user_id, "u2");
  assert.equal(ranked[0].match_score, 160);
});

test("coupon funding separates merchant coupons platform subsidy and merchant-borne campaigns", () => {
  const merchantCoupon = normalizeCouponPolicy({
    id: "c1",
    issuerType: "merchant",
    shopId: "shop_1",
    amountFen: 500
  });
  assert.equal(merchantCoupon.cost_bearer, "merchant");
  assert.equal(canCouponApplyToShop(merchantCoupon, "shop_1"), true);

  const platformCoupon = normalizeCouponPolicy({
    id: "c2",
    issuerType: "platform",
    costBearer: "platform",
    scopeType: "participating_shops",
    participatingShopIds: ["shop_1", "shop_2"]
  });
  assert.equal(platformCoupon.subsidy_settlement_required, true);
  assert.equal(canCouponApplyToShop(platformCoupon, "shop_2"), true);

  const merchantBorneCampaign = normalizeCouponPolicy({
    id: "c3",
    issuerType: "platform",
    costBearer: "merchant",
    scopeType: "participating_shops",
    merchantAcceptanceStatus: "pending",
    participatingShopIds: ["shop_1"]
  });
  assert.equal(merchantBorneCampaign.merchant_activity_acceptance_required, true);
  assert.equal(canCouponApplyToShop(merchantBorneCampaign, "shop_1"), false);
  assert.equal(canCouponApplyToShop({ ...merchantBorneCampaign, merchant_acceptance_status: "accepted" }, "shop_1"), true);
});

test("after sales reviews favorites points membership push and risk rules cover legacy closure", () => {
  const afterSales = normalizeAfterSalesRequest({
    orderId: "order_1",
    userId: "user_1",
    type: "food_safety",
    reason: "餐品异常",
	    requestedAmountFen: 1200,
	    orderAmountFen: 1500,
	    refundableFen: 1500,
	    evidenceUrls: ["https://cdn/evidence.jpg"]
	  });
	  assert.equal(afterSales.status, "pending_merchant");
	  assert.equal(afterSales.order_amount_fen, 1500);
	  assert.equal(afterSales.refundable_fen, 1500);
	  const evidence = normalizeAfterSalesEvidence({
	    requestId: "asr_1",
	    objectKey: "after-sales/asr_1/sig/evidence.jpg",
	    publicUrl: "https://cdn/evidence.jpg",
	    fileName: "evidence.jpg",
	    contentType: "image/jpeg",
	    sizeBytes: 1024,
	    uploadedById: "user_1",
	    uploadedByRole: "user"
	  });
	  assert.equal(evidence.status, "uploaded");
	  assert.equal(evidence.object_key, "after-sales/asr_1/sig/evidence.jpg");
  assert.deepEqual(canCreateAfterSalesRequest(
    { id: "order_1", status: "completed", amount_fen: 1500 },
    afterSales
  ), {
    ok: true,
    missing_fields: [],
    order_status_allowed: true
  });
  assert.equal(canCreateAfterSalesRequest(
    { id: "order_1", status: "refunded", amount_fen: 1500 },
    afterSales
  ).ok, false);

  const review = normalizeReview({ targetType: "rider", targetId: "r1", userId: "u1", rating: 8, content: "很好" });
  assert.equal(review.rating, 5);
  assert.deepEqual(normalizeFavorite({ userId: "u1", targetType: "shop", targetId: "s1" }), {
    user_id: "u1",
    target_type: "shop",
    target_id: "s1",
    created_at: ""
  });

  assert.deepEqual(calculateDeliveryPromise({
    promiseDeadlineAt: "2026-05-21T08:30:00.000Z",
    timeoutCompensationFen: 300
  }, new Date("2026-05-21T08:35:30.000Z")), {
    status: "timeout",
    timeout_seconds: 330,
    compensation_fen: 300
  });

  assert.deepEqual(applyPointsBalance(100, { userId: "u1", type: "refund_deduct", points: -30, sourceId: "order_1" }).balance, 70);
  assert.equal(resolveMembershipTier(12000), "black_gold");
  assert.deepEqual(normalizePushDelivery({ userId: "u1", channel: "wechat_subscribe", templateKey: "order_paid" }), {
    id: "",
    user_id: "u1",
    channel: "wechat_subscribe",
    template_key: "order_paid",
    status: "queued",
    retry_count: 0
  });
  assert.deepEqual(resolveRiskDecision([{ type: "abnormal_ordering" }, { type: "abnormal_ordering" }], { abnormal_order_limit: 2 }), {
    blocked: true,
    reasons: ["abnormal_ordering"]
  });
});
