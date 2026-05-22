import assert from "node:assert/strict";
import test from "node:test";
import {
  ADDRESS_TAGS,
  AFTER_SALES_ACTIONS,
  AFTER_SALES_EVIDENCE_STATUSES,
  AFTER_SALES_STATUSES,
  AFTER_SALES_TYPES,
  DEPOSIT_AMOUNTS_FEN,
  DEPOSIT_STATUSES,
  DEPOSIT_SUBJECT_TYPES,
  CIRCLE_FEATURE_MODES,
  CIRCLE_POST_STATUSES,
  CIRCLE_POST_TYPES,
  COUPON_CLAIM_REQUIREMENTS,
  COUPON_ACTIVITY_ACCEPTANCE_STATUSES,
  COUPON_COST_BEARERS,
  COUPON_ISSUER_TYPES,
  COUPON_SCOPE_TYPES,
  DATA_MANAGEMENT_SCOPES,
  DELIVERY_PROMISE_STATUSES,
  DISPATCH_MODES,
  DISPATCH_POLICY,
  FAVORITE_TARGET_TYPES,
  FULFILLMENT_MODES,
  GROUP_CHAT_TYPES,
  GROUP_NOTIFICATION_DEFAULTS,
  GROUPBUY_REDEMPTION_METHODS,
  GROUPBUY_VOUCHER_STATUSES,
  HOME_CARD_TYPES,
  HOME_MODULE_KEYS,
  MEAL_MATCH_PROFILE_REQUIREMENTS,
  MEMBERSHIP_TIERS,
  MERCHANT_ACCOUNT_TYPES,
  MERCHANT_QUALIFICATION_TYPES,
  MERCHANT_REGISTRATION_MODES,
  MERCHANT_STAFF_STATUSES,
  NOTIFICATION_CHANNELS,
  ONBOARDING_INVITE_STATUSES,
  ONBOARDING_INVITE_TYPES,
  ORDER_CANCEL_ACTORS,
  ORDER_OPTION_TYPES,
  ORDER_TYPES,
  PAYMENT_METHODS,
  POINTS_TRANSACTION_TYPES,
  PRODUCT_STATUSES,
  PUSH_DELIVERY_STATUSES,
  RED_PACKET_SCENES,
  RED_PACKET_STATUSES,
  RED_PACKET_TYPES,
  REFUND_DEFAULT_STRATEGIES,
  REFUND_DESTINATIONS,
  REVIEW_STATUSES,
  REVIEW_TARGET_TYPES,
  RIDER_ACCOUNT_TYPES,
  RIDER_APPEAL_STATUSES,
  RIDER_DAILY_QUOTA_EXEMPTION,
  RIDER_LEVELS,
  RIDER_REGISTRATION_MODES,
  SHOP_CAPABILITIES,
  SHOP_CATEGORIES,
  SHOP_OPERATION_STATUSES,
  SHOP_SERVICE_STATES,
  RISK_CONTROL_EVENTS,
  WALLET_PAYMENT_PASSWORD_STATUSES,
  errorEnvelope,
  isAfterSalesType,
  isFulfillmentMode,
  isMerchantQualificationType,
  isDepositStatus,
  isOrderType,
  isPaymentMethod,
  isRiderAccountType,
  isCouponCostBearer,
  isHomeCardType,
  isReviewTargetType,
  isRedPacketType,
  isRefundDestination,
  isShopCapability,
  successEnvelope
} from "./index.mjs";

test("core contracts expose fixed platform enums", () => {
  assert.equal(ORDER_TYPES.TAKEOUT, "takeout");
  assert.equal(ORDER_TYPES.GROUPBUY, "groupbuy");
  assert.equal(ORDER_TYPES.MEDICINE, "medicine");
  assert.equal(ORDER_TYPES.COURIER, "courier");
  assert.equal(ORDER_CANCEL_ACTORS.SYSTEM, "system");
  assert.equal(ORDER_OPTION_TYPES.TABLEWARE, "tableware");
  assert.equal(PAYMENT_METHODS.WECHAT_PAY, "wechat_pay");
  assert.equal(PAYMENT_METHODS.BALANCE, "balance");
  assert.equal(HOME_MODULE_KEYS.CIRCLE, "circle");
  assert.equal(HOME_MODULE_KEYS.CHARITY, "charity");
  assert.equal(HOME_MODULE_KEYS.SOCIAL, "social");
  assert.equal(MERCHANT_ACCOUNT_TYPES.CLINIC, "clinic");
  assert.equal(MERCHANT_REGISTRATION_MODES.ADMIN_INVITE_ONLY, "admin_invite_only");
  assert.equal(ONBOARDING_INVITE_TYPES.MERCHANT, "merchant");
  assert.equal(ONBOARDING_INVITE_TYPES.STATION_MANAGER, "station_manager");
  assert.equal(ONBOARDING_INVITE_STATUSES.ACTIVE, "active");
  assert.equal(SHOP_CATEGORIES.PHARMACY, "pharmacy");
  assert.equal(SHOP_CAPABILITIES.GROUPBUY, "groupbuy");
  assert.equal(SHOP_OPERATION_STATUSES.QUALIFICATION_EXPIRED, "qualification_expired");
  assert.equal(SHOP_SERVICE_STATES.BUSY, "busy");
  assert.equal(MERCHANT_QUALIFICATION_TYPES.BUSINESS_LICENSE, "business_license");
  assert.equal(MERCHANT_QUALIFICATION_TYPES.HEALTH_CERTIFICATE, "health_certificate");
  assert.equal(MERCHANT_STAFF_STATUSES.ACTIVE, "active");
  assert.equal(RIDER_ACCOUNT_TYPES.STATION_MANAGER, "station_manager");
  assert.equal(RIDER_REGISTRATION_MODES.ADMIN_OR_STATION_INVITE_ONLY, "admin_or_station_invite_only");
  assert.equal(DEPOSIT_SUBJECT_TYPES.RIDER, "rider");
  assert.equal(DEPOSIT_STATUSES.WECHAT_EXEMPT_APPROVED, "wechat_exempt_approved");
  assert.equal(DEPOSIT_AMOUNTS_FEN.RIDER, 5000);
  assert.equal(DEPOSIT_AMOUNTS_FEN.MERCHANT, 5000);
  assert.equal(DISPATCH_MODES.GRAB_HALL, "grab_hall");
  assert.equal(DISPATCH_POLICY.GRAB_HALL_SECONDS, 600);
  assert.equal(RIDER_LEVELS.S, "S");
  assert.equal(RIDER_DAILY_QUOTA_EXEMPTION.AFTER_FIXED_ORDER_COUNT, "after_fixed_order_count");
  assert.equal(RIDER_APPEAL_STATUSES.PENDING, "pending");
  assert.equal(PRODUCT_STATUSES.SOLD_OUT, "sold_out");
  assert.equal(REFUND_DESTINATIONS.BALANCE, "balance");
  assert.equal(REFUND_DEFAULT_STRATEGIES.BALANCE_FIRST, "balance_first");
  assert.equal(AFTER_SALES_TYPES.PARTIAL_REFUND, "partial_refund");
  assert.equal(AFTER_SALES_TYPES.FOOD_SAFETY, "food_safety");
  assert.equal(AFTER_SALES_ACTIONS.CUSTOMER_SERVICE_INTERVENTION, "customer_service_intervention");
  assert.equal(AFTER_SALES_ACTIONS.EVIDENCE_UPLOADED, "evidence_uploaded");
  assert.equal(AFTER_SALES_EVIDENCE_STATUSES.UPLOADED, "uploaded");
  assert.equal(AFTER_SALES_STATUSES.PENDING_MERCHANT, "pending_merchant");
  assert.equal(DELIVERY_PROMISE_STATUSES.TIMEOUT, "timeout");
  assert.equal(WALLET_PAYMENT_PASSWORD_STATUSES.SET, "set");
  assert.equal(GROUP_CHAT_TYPES.OFFICIAL, "official");
  assert.equal(GROUP_NOTIFICATION_DEFAULTS.MUTED, "muted");
  assert.equal(COUPON_CLAIM_REQUIREMENTS.GROUP_MEMBERSHIP, "group_membership");
  assert.equal(COUPON_ISSUER_TYPES.PLATFORM, "platform");
  assert.equal(COUPON_COST_BEARERS.MERCHANT, "merchant");
  assert.equal(COUPON_SCOPE_TYPES.PARTICIPATING_SHOPS, "participating_shops");
  assert.equal(COUPON_ACTIVITY_ACCEPTANCE_STATUSES.ACCEPTED, "accepted");
  assert.equal(CIRCLE_FEATURE_MODES.WALL_ONLY, "wall_only");
  assert.equal(CIRCLE_POST_TYPES.FOOD_INVITE, "food_invite");
  assert.equal(CIRCLE_POST_STATUSES.PUBLISHED, "published");
  assert.equal(MEAL_MATCH_PROFILE_REQUIREMENTS.QUESTIONNAIRE_COMPLETED, "questionnaire_completed");
  assert.equal(HOME_CARD_TYPES.PRODUCT, "product");
  assert.equal(ADDRESS_TAGS.HOME, "home");
  assert.equal(FAVORITE_TARGET_TYPES.SHOP, "shop");
  assert.equal(REVIEW_TARGET_TYPES.RIDER, "rider");
  assert.equal(REVIEW_STATUSES.PUBLISHED, "published");
  assert.equal(POINTS_TRANSACTION_TYPES.REFUND_DEDUCT, "refund_deduct");
  assert.equal(MEMBERSHIP_TIERS.BLACK_GOLD, "black_gold");
  assert.equal(NOTIFICATION_CHANNELS.WECHAT_SUBSCRIBE, "wechat_subscribe");
  assert.equal(PUSH_DELIVERY_STATUSES.ACKED, "acked");
  assert.equal(RISK_CONTROL_EVENTS.FAKE_TRANSACTION, "fake_transaction");
  assert.equal(DATA_MANAGEMENT_SCOPES.FULL_BUNDLE, "full_bundle");
  assert.equal(RED_PACKET_SCENES.DIRECT_MESSAGE, "direct_message");
  assert.equal(RED_PACKET_TYPES.RANDOM, "random");
  assert.equal(RED_PACKET_STATUSES.CREATED, "created");
  assert.equal(FULFILLMENT_MODES.IN_STORE_REDEMPTION, "in_store_redemption");
  assert.equal(GROUPBUY_VOUCHER_STATUSES.REDEEMED, "redeemed");
  assert.equal(GROUPBUY_REDEMPTION_METHODS.QR_SCAN, "qr_scan");
});

test("contract guards accept only known values", () => {
  assert.equal(isOrderType("takeout"), true);
  assert.equal(isOrderType("unknown"), false);
  assert.equal(isPaymentMethod("balance"), true);
  assert.equal(isPaymentMethod("cash"), false);
  assert.equal(isShopCapability("groupbuy"), true);
  assert.equal(isShopCapability("courier"), false);
  assert.equal(isFulfillmentMode("rider_delivery"), true);
  assert.equal(isFulfillmentMode("legacy"), false);
  assert.equal(isMerchantQualificationType("business_license"), true);
  assert.equal(isMerchantQualificationType("random_file"), false);
  assert.equal(isRiderAccountType("station_manager"), true);
  assert.equal(isRiderAccountType("captain"), false);
  assert.equal(isDepositStatus("paid"), true);
  assert.equal(isDepositStatus("unknown"), false);
  assert.equal(isCouponCostBearer("merchant"), true);
  assert.equal(isCouponCostBearer("user"), false);
  assert.equal(isHomeCardType("product"), true);
  assert.equal(isHomeCardType("banner"), false);
  assert.equal(isAfterSalesType("refund_only"), true);
  assert.equal(isAfterSalesType("exchange"), false);
  assert.equal(isReviewTargetType("shop"), true);
  assert.equal(isReviewTargetType("driver"), false);
  assert.equal(isRefundDestination("balance"), true);
  assert.equal(isRefundDestination("cash"), false);
  assert.equal(isRedPacketType("random"), true);
  assert.equal(isRedPacketType("lucky"), false);
});

test("response envelopes stay stable", () => {
  assert.deepEqual(successEnvelope({ ok: true }, "done", "req-1"), {
    success: true,
    message: "done",
    data: { ok: true },
    request_id: "req-1"
  });
  assert.deepEqual(errorEnvelope("BAD_REQUEST", "invalid"), {
    success: false,
    code: "BAD_REQUEST",
    message: "invalid",
    details: undefined,
    request_id: undefined
  });
});
