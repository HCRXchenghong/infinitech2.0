export const ORDER_TYPES = Object.freeze({
  TAKEOUT: "takeout",
  GROUPBUY: "groupbuy",
  MEDICINE: "medicine",
  COURIER: "courier",
  ERRAND_BUY: "errand_buy",
  ERRAND_DELIVER: "errand_deliver",
  ERRAND_PICKUP: "errand_pickup",
  ERRAND_DO: "errand_do"
});

export const ORDER_STATUSES = Object.freeze({
  CREATED: "created",
  PENDING_PAYMENT: "pending_payment",
  PAID: "paid",
  MERCHANT_PENDING: "merchant_pending",
  PREPARING: "preparing",
  DISPATCHING: "dispatching",
  RIDER_ASSIGNED: "rider_assigned",
  RIDER_ARRIVED_STORE: "rider_arrived_store",
  PICKED_UP: "picked_up",
  DELIVERING: "delivering",
  COMPLETED: "completed",
  CANCELLED: "cancelled",
  REFUND_PENDING: "refund_pending",
  REFUNDED: "refunded"
});

export const ORDER_CANCEL_ACTORS = Object.freeze({
  USER: "user",
  MERCHANT: "merchant",
  RIDER: "rider",
  SYSTEM: "system",
  ADMIN: "admin"
});

export const ORDER_OPTION_TYPES = Object.freeze({
  REMARK: "remark",
  TABLEWARE: "tableware",
  INVOICE: "invoice",
  CONTACTLESS_DELIVERY: "contactless_delivery"
});

export const PAYMENT_METHODS = Object.freeze({
  WECHAT_PAY: "wechat_pay",
  BALANCE: "balance",
  MIXED: "mixed"
});

export const MERCHANT_ACCOUNT_TYPES = Object.freeze({
  STANDARD: "standard",
  PHARMACY: "pharmacy",
  CLINIC: "clinic",
  PLATFORM_SERVICE: "platform_service"
});

export const MERCHANT_REGISTRATION_MODES = Object.freeze({
  ADMIN_INVITE_ONLY: "admin_invite_only"
});

export const ONBOARDING_INVITE_TYPES = Object.freeze({
  MERCHANT: "merchant",
  RIDER: "rider",
  STATION_MANAGER: "station_manager",
  OLD_USER: "old_user"
});

export const ONBOARDING_INVITE_STATUSES = Object.freeze({
  ACTIVE: "active",
  USED: "used",
  EXPIRED: "expired",
  REVOKED: "revoked"
});

export const SHOP_CATEGORIES = Object.freeze({
  RESTAURANT: "restaurant",
  RETAIL: "retail",
  PHARMACY: "pharmacy",
  CLINIC: "clinic",
  SERVICE: "service"
});

export const SHOP_CAPABILITIES = Object.freeze({
  TAKEOUT: "takeout",
  GROUPBUY: "groupbuy",
  MEDICINE: "medicine"
});

export const SHOP_OPERATION_STATUSES = Object.freeze({
  ACTIVE: "active",
  PENDING_REVIEW: "pending_review",
  TEMPORARILY_CLOSED: "temporarily_closed",
  QUALIFICATION_EXPIRED: "qualification_expired"
});

export const SHOP_SERVICE_STATES = Object.freeze({
  OPEN: "open",
  CLOSED: "closed",
  BUSY: "busy",
  RESTING: "resting",
  SOLD_OUT: "sold_out"
});

export const MERCHANT_QUALIFICATION_TYPES = Object.freeze({
  BUSINESS_LICENSE: "business_license",
  HEALTH_CERTIFICATE: "health_certificate",
  SUPPLEMENTAL_DOCUMENT: "supplemental_document"
});

export const MERCHANT_STAFF_STATUSES = Object.freeze({
  ACTIVE: "active",
  INACTIVE: "inactive",
  LEFT: "left"
});

export const RIDER_ACCOUNT_TYPES = Object.freeze({
  STATION_MANAGER: "station_manager",
  RIDER: "rider"
});

export const RIDER_REGISTRATION_MODES = Object.freeze({
  ADMIN_OR_STATION_INVITE_ONLY: "admin_or_station_invite_only"
});

export const DEPOSIT_SUBJECT_TYPES = Object.freeze({
  RIDER: "rider",
  MERCHANT: "merchant"
});

export const DEPOSIT_STATUSES = Object.freeze({
  UNPAID: "unpaid",
  PAID: "paid",
  WECHAT_EXEMPT_APPROVED: "wechat_exempt_approved",
  REFUND_PENDING: "refund_pending",
  REFUNDED: "refunded",
  DISPUTE_HOLD: "dispute_hold"
});

export const DEPOSIT_AMOUNTS_FEN = Object.freeze({
  RIDER: 5000,
  MERCHANT: 5000
});

export const DISPATCH_MODES = Object.freeze({
  GRAB_HALL: "grab_hall",
  AUTO_ASSIGN: "auto_assign",
  MANUAL_ASSIGN: "manual_assign"
});

export const DISPATCH_POLICY = Object.freeze({
  GRAB_HALL_SECONDS: 600
});

export const RIDER_LEVELS = Object.freeze({
  S: "S",
  A: "A",
  B: "B",
  C: "C"
});

export const RIDER_DAILY_QUOTA_EXEMPTION = Object.freeze({
  AFTER_FIXED_ORDER_COUNT: "after_fixed_order_count"
});

export const RIDER_APPEAL_STATUSES = Object.freeze({
  PENDING: "pending",
  APPROVED: "approved",
  REJECTED: "rejected"
});

export const PRODUCT_STATUSES = Object.freeze({
  ACTIVE: "active",
  SOLD_OUT: "sold_out",
  REMOVED: "removed"
});

export const REFUND_DESTINATIONS = Object.freeze({
  BALANCE: "balance",
  ORIGINAL_ROUTE: "original_route"
});

export const REFUND_DEFAULT_STRATEGIES = Object.freeze({
  BALANCE_FIRST: "balance_first",
  ORIGINAL_ROUTE_FIRST: "original_route_first"
});

export const AFTER_SALES_TYPES = Object.freeze({
  REFUND_ONLY: "refund_only",
  PARTIAL_REFUND: "partial_refund",
  COMPLAINT: "complaint",
  FOOD_SAFETY: "food_safety"
});

export const AFTER_SALES_STATUSES = Object.freeze({
  PENDING_MERCHANT: "pending_merchant",
  MERCHANT_APPROVED: "merchant_approved",
  MERCHANT_REJECTED: "merchant_rejected",
  ADMIN_REVIEW: "admin_review",
  REFUNDED: "refunded",
  CLOSED: "closed"
});

export const AFTER_SALES_ACTIONS = Object.freeze({
  CREATED: "created",
  USER_SUPPLEMENT: "user_supplement",
  MERCHANT_REPLY: "merchant_reply",
  CUSTOMER_SERVICE_INTERVENTION: "customer_service_intervention",
  ARBITRATION_OPENED: "arbitration_opened",
  INTERNAL_NOTE: "internal_note",
  EVIDENCE_UPLOADED: "evidence_uploaded",
  REVIEW_APPROVED: "review_approved",
  REVIEW_REJECTED: "review_rejected",
  ESCALATED: "escalated"
});

export const AFTER_SALES_EVIDENCE_STATUSES = Object.freeze({
  UPLOADED: "uploaded"
});

export const DELIVERY_PROMISE_STATUSES = Object.freeze({
  ON_TIME: "on_time",
  TIMEOUT: "timeout",
  EXEMPT: "exempt"
});

export const WALLET_PAYMENT_PASSWORD_STATUSES = Object.freeze({
  UNSET: "unset",
  SET: "set",
  LOCKED: "locked"
});

export const GROUP_CHAT_TYPES = Object.freeze({
  OFFICIAL: "official",
  MERCHANT: "merchant",
  ORDER: "order"
});

export const GROUP_NOTIFICATION_DEFAULTS = Object.freeze({
  MUTED: "muted",
  NORMAL: "normal"
});

export const COUPON_CLAIM_REQUIREMENTS = Object.freeze({
  NONE: "none",
  GROUP_MEMBERSHIP: "group_membership"
});

export const COUPON_ISSUER_TYPES = Object.freeze({
  MERCHANT: "merchant",
  PLATFORM: "platform"
});

export const COUPON_COST_BEARERS = Object.freeze({
  MERCHANT: "merchant",
  PLATFORM: "platform"
});

export const COUPON_SCOPE_TYPES = Object.freeze({
  SINGLE_SHOP: "single_shop",
  PARTICIPATING_SHOPS: "participating_shops"
});

export const COUPON_ACTIVITY_ACCEPTANCE_STATUSES = Object.freeze({
  PENDING: "pending",
  ACCEPTED: "accepted",
  REJECTED: "rejected"
});

export const RED_PACKET_SCENES = Object.freeze({
  GROUP_CHAT: "group_chat",
  DIRECT_MESSAGE: "direct_message"
});

export const RED_PACKET_TYPES = Object.freeze({
  FIXED: "fixed",
  RANDOM: "random"
});

export const RED_PACKET_STATUSES = Object.freeze({
  CREATED: "created",
  CLAIMING: "claiming",
  FINISHED: "finished",
  REFUNDED: "refunded"
});

export const FULFILLMENT_MODES = Object.freeze({
  RIDER_DELIVERY: "rider_delivery",
  IN_STORE_REDEMPTION: "in_store_redemption",
  PLATFORM_ERRAND: "platform_errand"
});

export const GROUPBUY_VOUCHER_STATUSES = Object.freeze({
  ISSUED: "issued",
  REDEEMED: "redeemed",
  EXPIRED: "expired",
  REFUND_PENDING: "refund_pending",
  REFUNDED: "refunded"
});

export const ADDRESS_TAGS = Object.freeze({
  HOME: "home",
  COMPANY: "company",
  SCHOOL: "school",
  OTHER: "other"
});

export const FAVORITE_TARGET_TYPES = Object.freeze({
  SHOP: "shop",
  PRODUCT: "product",
  GROUPBUY_DEAL: "groupbuy_deal"
});

export const REVIEW_TARGET_TYPES = Object.freeze({
  ORDER: "order",
  SHOP: "shop",
  RIDER: "rider",
  PRODUCT: "product"
});

export const REVIEW_STATUSES = Object.freeze({
  PUBLISHED: "published",
  HIDDEN: "hidden",
  DELETED: "deleted"
});

export const POINTS_TRANSACTION_TYPES = Object.freeze({
  EARN: "earn",
  REDEEM: "redeem",
  REFUND_DEDUCT: "refund_deduct",
  ADMIN_ADJUST: "admin_adjust"
});

export const MEMBERSHIP_TIERS = Object.freeze({
  NONE: "none",
  SILVER: "silver",
  GOLD: "gold",
  BLACK_GOLD: "black_gold"
});

export const NOTIFICATION_CHANNELS = Object.freeze({
  IN_APP: "in_app",
  WECHAT_SUBSCRIBE: "wechat_subscribe",
  SMS: "sms",
  PUSH: "push"
});

export const PUSH_DELIVERY_STATUSES = Object.freeze({
  QUEUED: "queued",
  SENT: "sent",
  FAILED: "failed",
  ACKED: "acked"
});

export const RISK_CONTROL_EVENTS = Object.freeze({
  ABNORMAL_ORDERING: "abnormal_ordering",
  MALICIOUS_REFUND: "malicious_refund",
  FAKE_TRANSACTION: "fake_transaction",
  CONTACT_ABUSE: "contact_abuse"
});

export const DATA_MANAGEMENT_SCOPES = Object.freeze({
  USERS: "users",
  ORDERS: "orders",
  MERCHANTS: "merchants",
  RIDERS: "riders",
  SYSTEM_CONFIG: "system_config",
  FULL_BUNDLE: "full_bundle"
});

export const GROUPBUY_REDEMPTION_METHODS = Object.freeze({
  QR_SCAN: "qr_scan"
});

export const REALTIME_EVENTS = Object.freeze({
  MESSAGE_CREATED: "message.created",
  ORDER_STATUS_CHANGED: "order.status_changed",
  RIDER_LOCATION_UPDATED: "rider.location_updated",
  DISPATCH_ASSIGNED: "dispatch.assigned",
  RTC_SIGNAL: "rtc.signal"
});

export const HOME_MODULE_KEYS = Object.freeze({
  TAKEOUT: "takeout",
  GROUPBUY: "groupbuy",
  MEDICINE: "medicine",
  COURIER: "courier",
  CIRCLE: "circle",
  CHARITY: "charity",
  SOCIAL: "social"
});

export const CIRCLE_FEATURE_MODES = Object.freeze({
  DISABLED: "disabled",
  WALL_ONLY: "wall_only",
  CIRCLE_AND_MEAL_MATCH: "circle_and_meal_match"
});

export const CIRCLE_POST_TYPES = Object.freeze({
  TEXT: "text",
  IMAGE: "image",
  FOOD_INVITE: "food_invite"
});

export const CIRCLE_POST_STATUSES = Object.freeze({
  PENDING_REVIEW: "pending_review",
  PUBLISHED: "published",
  HIDDEN: "hidden",
  DELETED: "deleted"
});

export const MEAL_MATCH_PROFILE_REQUIREMENTS = Object.freeze({
  GENDER: "gender",
  IDENTITY_TRUTH_SIGNED: "identity_truth_signed",
  PLATFORM_LIABILITY_RELEASE_SIGNED: "platform_liability_release_signed",
  QUESTIONNAIRE_COMPLETED: "questionnaire_completed"
});

export const HOME_CARD_TYPES = Object.freeze({
  PRODUCT: "product",
  SHOP: "shop",
  GROUPBUY_DEAL: "groupbuy_deal",
  COUPON: "coupon",
  CIRCLE_POST: "circle_post"
});

export function successEnvelope(data, message = "ok", requestId = "") {
  return {
    success: true,
    message,
    data,
    request_id: requestId || undefined
  };
}

export function errorEnvelope(code, message, requestId = "", details = undefined) {
  return {
    success: false,
    code,
    message,
    details,
    request_id: requestId || undefined
  };
}

export function isOrderType(value) {
  return Object.values(ORDER_TYPES).includes(value);
}

export function isPaymentMethod(value) {
  return Object.values(PAYMENT_METHODS).includes(value);
}

export function isShopCapability(value) {
  return Object.values(SHOP_CAPABILITIES).includes(value);
}

export function isFulfillmentMode(value) {
  return Object.values(FULFILLMENT_MODES).includes(value);
}

export function isMerchantQualificationType(value) {
  return Object.values(MERCHANT_QUALIFICATION_TYPES).includes(value);
}

export function isRiderAccountType(value) {
  return Object.values(RIDER_ACCOUNT_TYPES).includes(value);
}

export function isDepositStatus(value) {
  return Object.values(DEPOSIT_STATUSES).includes(value);
}

export function isRefundDestination(value) {
  return Object.values(REFUND_DESTINATIONS).includes(value);
}

export function isRedPacketType(value) {
  return Object.values(RED_PACKET_TYPES).includes(value);
}

export function isCouponCostBearer(value) {
  return Object.values(COUPON_COST_BEARERS).includes(value);
}

export function isHomeCardType(value) {
  return Object.values(HOME_CARD_TYPES).includes(value);
}

export function isAfterSalesType(value) {
  return Object.values(AFTER_SALES_TYPES).includes(value);
}

export function isReviewTargetType(value) {
  return Object.values(REVIEW_TARGET_TYPES).includes(value);
}
