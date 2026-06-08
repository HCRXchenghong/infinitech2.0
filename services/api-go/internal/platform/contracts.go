package platform

import (
	"strings"
	"time"
)

const (
	OrderTypeTakeout       = "takeout"
	OrderTypeGroupbuy      = "groupbuy"
	OrderTypeMedicine      = "medicine"
	OrderTypeCourier       = "courier"
	OrderTypeErrandBuy     = "errand_buy"
	OrderTypeErrandDeliver = "errand_deliver"
	OrderTypeErrandPickup  = "errand_pickup"
	OrderTypeErrandDo      = "errand_do"

	StatusPendingPayment  = "pending_payment"
	StatusMerchantPending = "merchant_pending"
	StatusPreparing       = "preparing"
	StatusDispatching     = "dispatching"
	StatusRiderAssigned   = "rider_assigned"
	StatusPickedUp        = "picked_up"
	StatusDelivering      = "delivering"
	StatusVoucherIssued   = "voucher_issued"
	StatusCancelled       = "cancelled"
	StatusCompleted       = "completed"
	StatusRefundPending   = "refund_pending"
	StatusRefunded        = "refunded"

	PaymentBalance = "balance"
	PaymentWechat  = "wechat_pay"

	OrderCancelActorSystem         = "system"
	OrderOptionRemark              = "remark"
	OrderOptionTableware           = "tableware"
	OrderOptionInvoice             = "invoice"
	OrderOptionContactlessDelivery = "contactless_delivery"

	MerchantAccountStandard        = "standard"
	MerchantAccountPharmacy        = "pharmacy"
	MerchantAccountClinic          = "clinic"
	MerchantAccountPlatformService = "platform_service"

	ShopCapabilityTakeout  = "takeout"
	ShopCapabilityGroupbuy = "groupbuy"
	ShopCapabilityMedicine = "medicine"
	ShopServiceOpen        = "open"
	ShopServiceBusy        = "busy"
	ShopServiceClosed      = "closed"

	FulfillmentRiderDelivery    = "rider_delivery"
	FulfillmentInStoreRedeem    = "in_store_redemption"
	FulfillmentPlatformErrand   = "platform_errand"
	GroupbuyRedemptionMethodQR  = "qr_scan"
	GroupbuyVoucherStatusIssued = "issued"
	GroupbuyVoucherRedeemed     = "redeemed"
	GroupbuyVoucherRefunded     = "refunded"

	MerchantRegistrationAdminInviteOnly = "admin_invite_only"
	OnboardingInviteMerchant            = "merchant"
	OnboardingInviteActive              = "active"
	OnboardingInviteUsed                = "used"
	OnboardingInviteExpired             = "expired"
	QualificationBusinessLicense        = "business_license"
	QualificationHealthCertificate      = "health_certificate"
	QualificationSupplementalDocument   = "supplemental_document"
	QualificationStatusPendingReview    = "pending_review"
	QualificationStatusApproved         = "approved"
	QualificationStatusRejected         = "rejected"
	ShopStatusActive                    = "active"
	ShopStatusQualificationExpired      = "qualification_expired"
	MerchantStaffActive                 = "active"

	RiderAccountStationManager = "station_manager"
	RiderAccountRider          = "rider"
	RiderInviteOnly            = "admin_or_station_invite_only"

	DepositStatusUnpaid               = "unpaid"
	DepositStatusPaid                 = "paid"
	DepositStatusWechatExemptApproved = "wechat_exempt_approved"
	DepositStatusRefundPending        = "refund_pending"
	DepositStatusRefunded             = "refunded"
	DepositStatusDisputeHold          = "dispute_hold"
	RiderDepositAmountFen             = int64(5000)
	MerchantDepositAmountFen          = int64(5000)

	DispatchModeGrabHall             = "grab_hall"
	DispatchModeAutoAssign           = "auto_assign"
	DispatchModeManualAssign         = "manual_assign"
	DispatchGrabHallSeconds          = 600
	DispatchAssignmentTimeoutSeconds = 60

	RiderLevelS        = "S"
	RiderLevelA        = "A"
	RiderLevelB        = "B"
	RiderLevelC        = "C"
	RiderAppealPending = "pending"

	ProductStatusActive  = "active"
	ProductStatusSoldOut = "sold_out"
	ProductStatusRemoved = "removed"

	RefundDestinationBalance         = "balance"
	RefundDestinationOriginalRoute   = "original_route"
	RefundStrategyBalanceFirst       = "balance_first"
	RefundStrategyOriginalFirst      = "original_route_first"
	RefundStatusSuccess              = "success"
	RefundStatusPendingOriginal      = "pending_original_route"
	AfterSalesRefundOnly             = "refund_only"
	AfterSalesPartialRefund          = "partial_refund"
	AfterSalesFoodSafety             = "food_safety"
	AfterSalesPendingMerchant        = "pending_merchant"
	AfterSalesAdminReview            = "admin_review"
	AfterSalesApproved               = "approved"
	AfterSalesRejected               = "rejected"
	AfterSalesRefunded               = "refunded"
	AfterSalesDecisionApprove        = "approve"
	AfterSalesDecisionReject         = "reject"
	AfterSalesDecisionEscalate       = "escalate"
	AfterSalesActionCreated          = "created"
	AfterSalesActionUserSupplement   = "user_supplement"
	AfterSalesActionMerchantReply    = "merchant_reply"
	AfterSalesActionCustomerCare     = "customer_service_intervention"
	AfterSalesActionArbitration      = "arbitration_opened"
	AfterSalesActionInternalNote     = "internal_note"
	AfterSalesActionEvidenceUploaded = "evidence_uploaded"
	AfterSalesActionReviewApproved   = "review_approved"
	AfterSalesActionReviewRejected   = "review_rejected"
	AfterSalesActionEscalated        = "escalated"
	AfterSalesEvidenceUploaded       = "uploaded"
	AfterSalesUploadTicketIssued     = "issued"
	AfterSalesUploadTicketUploaded   = "uploaded"
	AfterSalesUploadTicketConfirmed  = "confirmed"
	AfterSalesUploadTicketDeleted    = "deleted"
	AfterSalesUploadScanNotRequired  = "not_required"
	AfterSalesUploadScanPending      = "pending"
	AfterSalesUploadScanPassed       = "passed"
	AfterSalesUploadScanRejected     = "rejected"
	AfterSalesObjectCleanupExpired   = "expired_unconfirmed"
	AfterSalesObjectCleanupRejected  = "scan_rejected"
	AfterSalesEvidenceMaxBytes       = int64(10 * 1024 * 1024)
	DeliveryPromiseOnTime            = "on_time"
	DeliveryPromiseTimeout           = "timeout"
	DeliveryPromiseExempt            = "exempt"

	WalletPaymentPasswordUnset  = "unset"
	WalletPaymentPasswordSet    = "set"
	WalletPaymentPasswordLocked = "locked"

	GroupChatOfficial = "official"
	GroupChatMerchant = "merchant"
	GroupNotifyMuted  = "muted"

	MessageRiskPassed                   = "passed"
	MessageRiskFlagged                  = "flagged"
	MessageRiskBlocked                  = "blocked"
	MessageRiskPaymentPasswordDisclosed = "payment_password_disclosed"
	MessageRiskVerificationCodeShared   = "verification_code_shared"
	MessageRiskBankCardShared           = "bank_card_shared"
	MessageRiskSensitiveMention         = "sensitive_mention"

	CouponRequirementNone            = "none"
	CouponRequirementGroupMembership = "group_membership"
	CouponIssuerMerchant             = "merchant"
	CouponIssuerPlatform             = "platform"
	CouponCostBearerMerchant         = "merchant"
	CouponCostBearerPlatform         = "platform"
	CouponScopeSingleShop            = "single_shop"
	CouponScopeParticipatingShops    = "participating_shops"
	CouponActivityPending            = "pending"
	CouponActivityAccepted           = "accepted"
	CouponActivityRejected           = "rejected"

	RedPacketSceneGroupChat     = "group_chat"
	RedPacketSceneDirectMessage = "direct_message"
	RedPacketTypeFixed          = "fixed"
	RedPacketTypeRandom         = "random"
	RedPacketStatusCreated      = "created"
	RedPacketStatusFinished     = "finished"
	RedPacketStatusRefunded     = "refunded"
	RedPacketStatusExpired      = "expired_refunded"
	RedPacketRiskPassed         = "passed"
	RedPacketRiskBlocked        = "blocked"
	RedPacketRiskFrequencyLimit = "claim_frequency_limit"
	RedPacketRiskAmountLimit    = "claim_amount_limit"
	RedPacketRiskSender         = "sender_cannot_claim"

	ServiceTicketStatusProcessing     = "processing"
	ServiceTicketStatusWaitingConfirm = "waiting_confirm"
	ServiceTicketStatusResolved       = "resolved"
	ServiceTicketStatusClosed         = "closed"
	ServiceTicketSLAStatusNormal      = "normal"
	ServiceTicketSLAStatusDueSoon     = "due_soon"
	ServiceTicketSLAStatusOverdue     = "overdue"
	ServiceTicketSLAStatusEscalated   = "escalated"
	ServiceTicketSLAStatusCompleted   = "completed"
	ServiceTicketQualityPassed        = "passed"
	ServiceTicketQualityNeedsCoaching = "needs_coaching"
	ServiceTicketQualityCritical      = "critical"
	ServiceTicketEventDone            = "done"
	ServiceTicketEventActive          = "active"
	ServiceTicketEventPending         = "pending"

	PrescriptionReviewPending  = "pending"
	PrescriptionReviewApproved = "approved"
	PrescriptionReviewRejected = "rejected"
	PrescriptionOCRMatched     = "matched"
	PrescriptionOCRNeedReview  = "need_review"

	HomeModuleCircle  = "circle"
	HomeModuleCharity = "charity"
	HomeCardProduct   = "product"
	HomeCardShop      = "shop"
	HomeCardCoupon    = "coupon"
	HomeCardCircle    = "circle_post"

	CircleFeatureDisabled            = "disabled"
	CircleFeatureWallOnly            = "wall_only"
	CircleFeatureCircleAndMealMatch  = "circle_and_meal_match"
	CirclePostText                   = "text"
	CirclePostImage                  = "image"
	CirclePostFoodInvite             = "food_invite"
	CirclePostPendingReview          = "pending_review"
	CirclePostPublished              = "published"
	MealMatchModerationReported      = "reported"
	MealMatchModerationBlocked       = "blocked"
	MealMatchModerationProfileReview = "profile_review"
	MealMatchModerationPending       = "pending_review"
	MealMatchModerationApproved      = "approved"
	MealMatchModerationRejected      = "rejected"
	MealMatchModerationActive        = "active"
	MealMatchPrivacySameSchool       = "same_school"
	MealMatchPrivacySameBuilding     = "same_building"
	MealMatchLocationCampusOnly      = "campus_only"
	MealMatchLocationBuildingOnly    = "building_only"
	MealMatchDeviceRiskPassed        = "passed"
	MealMatchDeviceRiskReview        = "review"
	MealMatchDeviceRiskBlocked       = "blocked"
	MealMatchDeviceRiskMissing       = "device_missing"
	MealMatchDeviceRiskSharedDevice  = "shared_device"
	MealMatchDeviceRiskKnownBlocked  = "known_blocked_device"

	AddressTagHome                = "home"
	FavoriteTargetShop            = "shop"
	ReviewTargetOrder             = "order"
	ReviewTargetShop              = "shop"
	ReviewTargetRider             = "rider"
	ReviewPublished               = "published"
	PointsTransactionEarn         = "earn"
	PointsTransactionRedeem       = "redeem"
	PointsTransactionRefund       = "refund_deduct"
	MembershipNone                = "none"
	MembershipSilver              = "silver"
	MembershipGold                = "gold"
	MembershipBlackGold           = "black_gold"
	NotificationChannelInApp      = "in_app"
	NotificationWechatSubscribe   = "wechat_subscribe"
	NotificationSMS               = "sms"
	NotificationEnterpriseWechat  = "enterprise_wechat"
	NotificationPush              = "push"
	NotificationStatusUnread      = "unread"
	NotificationStatusRead        = "read"
	NotificationDeliveryQueued    = "queued"
	NotificationDeliveryDelivered = "delivered"
	NotificationDeliveryFailed    = "failed"
	PushStatusQueued              = "queued"
	PushStatusAcked               = "acked"
	RiskEventAbnormalOrdering     = "abnormal_ordering"
	RiskEventMaliciousRefund      = "malicious_refund"
	RiskEventFakeTransaction      = "fake_transaction"
	DataManagementFullBundle      = "full_bundle"

	OutboxStatusPending    = "pending"
	OutboxStatusPublished  = "published"
	OutboxStatusFailed     = "failed"
	OutboxStatusDeadLetter = "dead_letter"
)

type HomeModule struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Route     string `json:"route"`
	Icon      string `json:"icon"`
	IconURL   string `json:"icon_url,omitempty"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
	Scene     string `json:"scene"`
}

type HomeCard struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	ImageURL  string `json:"image_url"`
	TargetID  string `json:"target_id"`
	ShopID    string `json:"shop_id"`
	PriceFen  int64  `json:"price_fen"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

type AppUser struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone,omitempty"`
	AvatarURL string    `json:"avatar_url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WechatMiniLoginRequest struct {
	Code            string `json:"code"`
	Nickname        string `json:"nickname"`
	AvatarURL       string `json:"avatar_url"`
	ProviderOpenID  string `json:"-"`
	ProviderUnionID string `json:"-"`
}

type WechatMiniLoginResult struct {
	User            AppUser `json:"user"`
	Provider        string  `json:"provider"`
	ProviderOpenID  string  `json:"provider_open_id"`
	ProviderUnionID string  `json:"provider_union_id,omitempty"`
	IsNewUser       bool    `json:"is_new_user"`
}

type SendPhoneVerificationCodeRequest struct {
	Phone   string `json:"phone"`
	Purpose string `json:"purpose"`
}

type PhoneVerificationCodeTicket struct {
	Phone             string    `json:"phone"`
	Purpose           string    `json:"purpose"`
	MaskedPhone       string    `json:"masked_phone"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	CooldownSeconds   int64     `json:"cooldown_seconds"`
	DeliveryProvider  string    `json:"delivery_provider,omitempty"`
	DeliveryStatus    string    `json:"delivery_status,omitempty"`
	DeliveryRequestID string    `json:"delivery_request_id,omitempty"`
	DevCode           string    `json:"dev_code,omitempty"`
}

type PhoneVerificationConfig struct {
	Mode            string
	Provider        string
	TemplateID      string
	Cooldown        time.Duration
	ExpiresIn       time.Duration
	MaxPerPhoneHour int
	MaxPerPhoneDay  int
	ReturnDevCode   bool
	Dispatcher      PhoneVerificationDispatcher
}

type PhoneVerificationDispatchRequest struct {
	Phone       string    `json:"phone"`
	MaskedPhone string    `json:"masked_phone"`
	Purpose     string    `json:"purpose"`
	Code        string    `json:"code"`
	TemplateID  string    `json:"template_id,omitempty"`
	Provider    string    `json:"provider,omitempty"`
	RequestID   string    `json:"request_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type PhoneVerificationDispatchResult struct {
	Provider  string    `json:"provider,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	SentAt    time.Time `json:"sent_at,omitempty"`
}

type PhoneVerificationDispatcher interface {
	DispatchPhoneVerificationCode(req PhoneVerificationDispatchRequest) (*PhoneVerificationDispatchResult, error)
}

type PhoneLoginRequest struct {
	Phone    string `json:"phone"`
	Code     string `json:"code,omitempty"`
	Password string `json:"password,omitempty"`
	Mode     string `json:"mode"`
}

type PhoneRegisterRequest struct {
	Phone             string `json:"phone"`
	Code              string `json:"code"`
	Password          string `json:"password"`
	Nickname          string `json:"nickname"`
	InviteCode        string `json:"invite_code,omitempty"`
	AcceptedAgreement bool   `json:"accepted_agreement"`
}

type PhoneAuthResult struct {
	User      AppUser `json:"user"`
	Provider  string  `json:"provider"`
	IsNewUser bool    `json:"is_new_user"`
}

type UserAddress struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	ContactName  string   `json:"contact_name"`
	ContactPhone string   `json:"contact_phone"`
	City         string   `json:"city"`
	Detail       string   `json:"detail"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Tag          string   `json:"tag"`
	IsDefault    bool     `json:"is_default"`
}

type OrderOptions struct {
	Remark              string `json:"remark"`
	TablewareCount      int    `json:"tableware_count"`
	ContactlessDelivery bool   `json:"contactless_delivery"`
	InvoiceRequested    bool   `json:"invoice_requested"`
}

type MerchantAccount struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	DisplayName   string `json:"display_name"`
	Status        string `json:"status"`
	DepositStatus string `json:"deposit_status"`
}

type Shop struct {
	ID             string   `json:"id"`
	MerchantID     string   `json:"merchant_id"`
	StationID      string   `json:"station_id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	AccountType    string   `json:"account_type"`
	Status         string   `json:"status"`
	Capabilities   []string `json:"capabilities"`
	Qualifications []string `json:"qualifications"`
	CoverURL       string   `json:"cover_url"`
	LogoURL        string   `json:"logo_url"`
	Announcement   string   `json:"announcement"`
}

type ShopReviewEntry struct {
	ReviewID       string    `json:"review_id"`
	UserName       string    `json:"user_name"`
	AvatarText     string    `json:"avatar_text"`
	Rating         int       `json:"rating"`
	StarsText      string    `json:"stars_text"`
	Content        string    `json:"content"`
	ImageURLs      []string  `json:"image_urls,omitempty"`
	ItemHighlights []string  `json:"item_highlights,omitempty"`
	RiderRating    int       `json:"rider_rating,omitempty"`
	RiderStarsText string    `json:"rider_stars_text,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	ReplyText      string    `json:"reply_text,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	CreatedText    string    `json:"created_text,omitempty"`
}

type ShopReviewSummary struct {
	AverageRating string   `json:"average_rating"`
	ReviewCount   int      `json:"review_count"`
	PositiveRate  string   `json:"positive_rate"`
	HighlightTags []string `json:"highlight_tags,omitempty"`
}

type ShopMerchantInfo struct {
	MerchantName       string   `json:"merchant_name"`
	QualificationText  string   `json:"qualification_text"`
	BusinessHours      string   `json:"business_hours"`
	ContactPhone       string   `json:"contact_phone"`
	Address            string   `json:"address"`
	ServiceCommitments []string `json:"service_commitments,omitempty"`
	QualificationItems []string `json:"qualification_items,omitempty"`
	SupportBulletins   []string `json:"support_bulletins,omitempty"`
}

type ShopDetail struct {
	ShopID            string            `json:"shop_id"`
	MerchantID        string            `json:"merchant_id"`
	Name              string            `json:"name"`
	Category          string            `json:"category"`
	CoverURL          string            `json:"cover_url,omitempty"`
	LogoURL           string            `json:"logo_url,omitempty"`
	Announcement      string            `json:"announcement"`
	RatingText        string            `json:"rating_text"`
	SalesText         string            `json:"sales_text"`
	DeliveryText      string            `json:"delivery_text"`
	QualificationText string            `json:"qualification_text"`
	ActivityTags      []string          `json:"activity_tags,omitempty"`
	ReviewSummary     ShopReviewSummary `json:"review_summary"`
	Reviews           []ShopReviewEntry `json:"reviews,omitempty"`
	MerchantInfo      ShopMerchantInfo  `json:"merchant_info"`
}

type MerchantOnboardingInvite struct {
	Token                string    `json:"token"`
	Type                 string    `json:"type"`
	Status               string    `json:"status"`
	CreatedByAdminID     string    `json:"created_by_admin_id"`
	CreatedBySubjectType string    `json:"created_by_subject_type"`
	CreatedBySubjectID   string    `json:"created_by_subject_id"`
	StationID            string    `json:"station_id,omitempty"`
	ExpiresAt            time.Time `json:"expires_at"`
}

type MerchantQualification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	FileURL   string    `json:"file_url"`
	ExpiresAt time.Time `json:"expires_at"`
	Status    string    `json:"status"`
}

type MerchantProfile struct {
	Account                    MerchantAccount                `json:"account"`
	Qualifications             []MerchantQualification        `json:"qualifications"`
	MissingQualifications      []string                       `json:"missing_qualifications"`
	CanAcceptOrders            bool                           `json:"can_accept_orders"`
	QualificationPopupRequired bool                           `json:"qualification_popup_required"`
	QualificationPopupCode     string                         `json:"qualification_popup_code"`
	Staff                      []MerchantStaff                `json:"staff"`
	SupplementalMaterials      []MerchantSupplementalMaterial `json:"supplemental_materials"`
}

type CreateMerchantInviteRequest struct {
	AdminID   string    `json:"admin_id"`
	Type      string    `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AcceptMerchantInviteRequest struct {
	Token       string `json:"token"`
	DisplayName string `json:"display_name"`
	AccountType string `json:"account_type"`
	Password    string `json:"password"`
}

type MerchantLoginRequest struct {
	AccountID string `json:"account_id"`
	Password  string `json:"password"`
}

type AdminLoginRequest struct {
	AccountID string `json:"account_id"`
	Password  string `json:"password"`
}

type CreateRiderInviteRequest struct {
	CreatedByID   string    `json:"created_by_id"`
	CreatedByRole string    `json:"created_by_role"`
	Type          string    `json:"type"`
	StationID     string    `json:"station_id"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type AcceptRiderInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type RiderLoginRequest struct {
	AccountID string `json:"account_id"`
	Password  string `json:"password"`
}

type UploadMerchantQualificationRequest struct {
	MerchantID string    `json:"merchant_id"`
	Type       string    `json:"type"`
	FileURL    string    `json:"file_url"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type ReviewMerchantQualificationRequest struct {
	MerchantID      string    `json:"merchant_id"`
	QualificationID string    `json:"qualification_id"`
	Decision        string    `json:"decision"`
	Reason          string    `json:"reason"`
	ReviewedAt      time.Time `json:"reviewed_at"`
}

type AdminRecommendedOperation struct {
	Key    string         `json:"key"`
	Title  string         `json:"title"`
	Reason string         `json:"reason"`
	Values map[string]any `json:"values"`
}

type AdminAuditFilter struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Action     string `json:"action,omitempty"`
	Limit      int    `json:"limit"`
}

type AdminMerchantQualificationListRequest struct {
	Status     string    `json:"status"`
	MerchantID string    `json:"merchant_id"`
	Type       string    `json:"type"`
	Limit      int       `json:"limit"`
	Now        time.Time `json:"now"`
}

type AdminMerchantQualificationDetailRequest struct {
	QualificationID string    `json:"qualification_id"`
	Now             time.Time `json:"now"`
	AuditLimit      int       `json:"audit_limit"`
}

type AdminMerchantQualificationCounts struct {
	Total         int `json:"total"`
	PendingReview int `json:"pending_review"`
	Approved      int `json:"approved"`
	Rejected      int `json:"rejected"`
	Expired       int `json:"expired"`
}

type AdminMerchantQualificationCase struct {
	Qualification              MerchantQualification     `json:"qualification"`
	Merchant                   MerchantAccount           `json:"merchant"`
	Shops                      []Shop                    `json:"shops"`
	Deposit                    *DepositAccount           `json:"deposit,omitempty"`
	MissingQualifications      []string                  `json:"missing_qualifications"`
	CanAcceptOrders            bool                      `json:"can_accept_orders"`
	QualificationPopupRequired bool                      `json:"qualification_popup_required"`
	QualificationPopupCode     string                    `json:"qualification_popup_code,omitempty"`
	IncidentCode               string                    `json:"incident_code"`
	IncidentSeverity           string                    `json:"incident_severity"`
	ExpiresInSeconds           int64                     `json:"expires_in_seconds"`
	ReviewSLAHours             int                       `json:"review_sla_hours"`
	RecommendedOperation       AdminRecommendedOperation `json:"recommended_operation"`
}

type AdminMerchantQualificationList struct {
	GeneratedAt    time.Time                        `json:"generated_at"`
	Status         string                           `json:"status,omitempty"`
	MerchantID     string                           `json:"merchant_id,omitempty"`
	Type           string                           `json:"type,omitempty"`
	Counts         AdminMerchantQualificationCounts `json:"counts"`
	Qualifications []AdminMerchantQualificationCase `json:"qualifications"`
}

type AdminMerchantQualificationDetail struct {
	GeneratedAt                time.Time                 `json:"generated_at"`
	Qualification              MerchantQualification     `json:"qualification"`
	Merchant                   MerchantAccount           `json:"merchant"`
	Shops                      []Shop                    `json:"shops"`
	Deposit                    *DepositAccount           `json:"deposit,omitempty"`
	MissingQualifications      []string                  `json:"missing_qualifications"`
	CanAcceptOrders            bool                      `json:"can_accept_orders"`
	QualificationPopupRequired bool                      `json:"qualification_popup_required"`
	QualificationPopupCode     string                    `json:"qualification_popup_code,omitempty"`
	IncidentCode               string                    `json:"incident_code"`
	IncidentSeverity           string                    `json:"incident_severity"`
	ExpiresInSeconds           int64                     `json:"expires_in_seconds"`
	ReviewSLAHours             int                       `json:"review_sla_hours"`
	AuditFilters               []AdminAuditFilter        `json:"audit_filters"`
	RecentAudits               []AuditLog                `json:"recent_audits"`
	RecommendedOperation       AdminRecommendedOperation `json:"recommended_operation"`
	Checklist                  []string                  `json:"checklist"`
}

type MerchantStaff struct {
	ID                         string    `json:"id"`
	MerchantID                 string    `json:"merchant_id"`
	ShopID                     string    `json:"shop_id"`
	Name                       string    `json:"name"`
	Phone                      string    `json:"phone"`
	Role                       string    `json:"role"`
	Status                     string    `json:"status"`
	HealthCertificateURL       string    `json:"health_certificate_url"`
	HealthCertificateExpiresAt time.Time `json:"health_certificate_expires_at"`
}

type MerchantSupplementalMaterial struct {
	ID          string    `json:"id"`
	MerchantID  string    `json:"merchant_id"`
	ShopID      string    `json:"shop_id"`
	Type        string    `json:"type"`
	FileURL     string    `json:"file_url"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type UpsertMerchantStaffRequest struct {
	MerchantID                 string    `json:"merchant_id"`
	StaffID                    string    `json:"staff_id"`
	ShopID                     string    `json:"shop_id"`
	Name                       string    `json:"name"`
	Phone                      string    `json:"phone"`
	Role                       string    `json:"role"`
	Status                     string    `json:"status"`
	HealthCertificateURL       string    `json:"health_certificate_url"`
	HealthCertificateExpiresAt time.Time `json:"health_certificate_expires_at"`
}

type UploadMerchantSupplementalMaterialRequest struct {
	MerchantID  string    `json:"merchant_id"`
	MaterialID  string    `json:"material_id"`
	ShopID      string    `json:"shop_id"`
	Type        string    `json:"type"`
	FileURL     string    `json:"file_url"`
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type RiderAccount struct {
	ID                   string  `json:"id"`
	StationID            string  `json:"station_id"`
	Type                 string  `json:"type"`
	Status               string  `json:"status"`
	Online               bool    `json:"online"`
	DepositStatus        string  `json:"deposit_status"`
	Capacity             int     `json:"capacity"`
	DispatchPriority     int     `json:"dispatch_priority"`
	AverageAcceptSeconds float64 `json:"average_accept_seconds"`
	AverageDailyOrders   float64 `json:"average_daily_orders"`
	CompletionRate       float64 `json:"completion_rate"`
	DistanceMeters       int     `json:"distance_meters"`
}

type StationTaskConfig struct {
	StationID                    string `json:"station_id"`
	ConfiguredByStationManagerID string `json:"configured_by_station_manager_id"`
	DailyTaskDurationMinutes     int    `json:"daily_task_duration_minutes"`
	DailyFixedOrderCount         int    `json:"daily_fixed_order_count"`
}

type StationServiceArea struct {
	StationID string   `json:"station_id"`
	ShopIDs   []string `json:"shop_ids"`
}

type DepositAccount struct {
	SubjectType               string    `json:"subject_type"`
	SubjectID                 string    `json:"subject_id"`
	AmountFen                 int64     `json:"amount_fen"`
	Status                    string    `json:"status"`
	LastOrderCompletedAt      time.Time `json:"last_order_completed_at"`
	ResignationSubmittedAt    time.Time `json:"resignation_submitted_at"`
	DisputeClosedAt           time.Time `json:"dispute_closed_at"`
	WechatExemptApplicationID string    `json:"wechat_exempt_application_id"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type PayDepositRequest struct {
	SubjectType    string `json:"subject_type"`
	SubjectID      string `json:"subject_id"`
	AmountFen      int64  `json:"amount_fen"`
	IdempotencyKey string `json:"idempotency_key"`
}

type RiderWechatExemptionRequest struct {
	RiderID       string `json:"rider_id"`
	ApplicationID string `json:"application_id"`
}

type RiderDepositRefundRequest struct {
	RiderID                string    `json:"rider_id"`
	ResignationSubmittedAt time.Time `json:"resignation_submitted_at"`
	DisputeClosedAt        time.Time `json:"dispute_closed_at"`
}

type RiderPerformanceScoreBreakdown struct {
	AcceptScore              float64 `json:"accept_score"`
	OrderVolumeScore         float64 `json:"order_volume_score"`
	CompletionScore          float64 `json:"completion_score"`
	RatingScore              float64 `json:"rating_score"`
	RatingConfidence         float64 `json:"rating_confidence,omitempty"`
	TeamAverageAcceptSeconds float64 `json:"team_average_accept_seconds"`
	TeamAverageDailyOrders   float64 `json:"team_average_daily_orders"`
}

type RiderPerformanceTrendPoint struct {
	Date            string  `json:"date"`
	Score           int     `json:"score"`
	CompletedOrders int     `json:"completed_orders"`
	AverageRating   float64 `json:"average_rating,omitempty"`
	TimeoutCount    int     `json:"timeout_count,omitempty"`
	RejectCount     int     `json:"reject_count,omitempty"`
}

type RiderPerformanceReviewExcerpt struct {
	ReviewID    string    `json:"review_id"`
	OrderID     string    `json:"order_id,omitempty"`
	Rating      int       `json:"rating"`
	RiderRating int       `json:"rider_rating,omitempty"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type RiderPerformanceExceptionSummary struct {
	DispatchTimeoutCount int       `json:"dispatch_timeout_count"`
	DispatchRejectCount  int       `json:"dispatch_reject_count"`
	AfterSalesCount      int       `json:"after_sales_count"`
	LowRatingCount       int       `json:"low_rating_count"`
	LastEventAt          time.Time `json:"last_event_at,omitempty"`
}

type RiderPerformanceExceptionDetail struct {
	Kind                string    `json:"kind"`
	Label               string    `json:"label"`
	OrderID             string    `json:"order_id,omitempty"`
	DispatchEventID     string    `json:"dispatch_event_id,omitempty"`
	AfterSalesRequestID string    `json:"after_sales_request_id,omitempty"`
	ReviewID            string    `json:"review_id,omitempty"`
	Status              string    `json:"status,omitempty"`
	Message             string    `json:"message"`
	CreatedAt           time.Time `json:"created_at"`
}

type RiderPerformance struct {
	RiderID              string                            `json:"rider_id"`
	StationID            string                            `json:"station_id"`
	AverageAcceptSeconds float64                           `json:"average_accept_seconds"`
	AverageDailyOrders   float64                           `json:"average_daily_orders"`
	CompletionRate       float64                           `json:"completion_rate"`
	RiderAverageRating   float64                           `json:"rider_average_rating,omitempty"`
	RiderReviewCount     int                               `json:"rider_review_count,omitempty"`
	Score                int                               `json:"score"`
	ScoreBreakdown       RiderPerformanceScoreBreakdown    `json:"score_breakdown"`
	RecentTrend          []RiderPerformanceTrendPoint      `json:"recent_trend,omitempty"`
	RecentReviews        []RiderPerformanceReviewExcerpt   `json:"recent_reviews,omitempty"`
	ExceptionSummary     RiderPerformanceExceptionSummary  `json:"exception_summary"`
	ExceptionDetails     []RiderPerformanceExceptionDetail `json:"exception_details,omitempty"`
	Level                string                            `json:"level"`
	DispatchPriority     int                               `json:"dispatch_priority"`
}

type SetRiderOnlineStatusRequest struct {
	RiderID        string `json:"rider_id"`
	Online         bool   `json:"online"`
	Capacity       int    `json:"capacity"`
	DistanceMeters int    `json:"distance_meters"`
}

type AutoAssignOrderRequest struct {
	OrderID string    `json:"order_id"`
	Now     time.Time `json:"now"`
}

type RejectRiderAssignmentRequest struct {
	OrderID string    `json:"order_id"`
	RiderID string    `json:"rider_id"`
	Now     time.Time `json:"now"`
}

type TimeoutReassignOrderRequest struct {
	OrderID          string    `json:"order_id"`
	RiderID          string    `json:"rider_id"`
	StationManagerID string    `json:"station_manager_id"`
	TimeoutSeconds   int       `json:"timeout_seconds"`
	Now              time.Time `json:"now"`
}

type CompensateOrderStateRequest struct {
	OrderID string    `json:"order_id"`
	ActorID string    `json:"actor_id"`
	Now     time.Time `json:"now"`
}

type CompensateOrderStateResult struct {
	Order            *Order   `json:"order"`
	Changed          bool     `json:"changed"`
	PreviousStatus   string   `json:"previous_status"`
	ExpectedStatus   string   `json:"expected_status"`
	PreviousRiderID  string   `json:"previous_rider_id,omitempty"`
	ExpectedRiderID  string   `json:"expected_rider_id,omitempty"`
	Evidence         []string `json:"evidence"`
	CompensationType string   `json:"compensation_type,omitempty"`
}

type ManualAssignOrderRequest struct {
	OrderID          string `json:"order_id"`
	RiderID          string `json:"rider_id"`
	StationManagerID string `json:"station_manager_id"`
}

type SaveStationTaskConfigRequest struct {
	StationManagerID         string `json:"station_manager_id"`
	DailyTaskDurationMinutes int    `json:"daily_task_duration_minutes"`
	DailyFixedOrderCount     int    `json:"daily_fixed_order_count"`
}

type DispatchDecision struct {
	OrderID                      string   `json:"order_id"`
	Mode                         string   `json:"mode"`
	StationID                    string   `json:"station_id"`
	CandidateRiderID             string   `json:"candidate_rider_id"`
	RejectedRiderIDs             []string `json:"rejected_rider_ids"`
	CanDeclineWithoutPenalty     bool     `json:"can_decline_without_penalty"`
	DailyCompletedOrderCount     int      `json:"daily_completed_order_count"`
	DailyFixedOrderCount         int      `json:"daily_fixed_order_count"`
	Reason                       string   `json:"reason,omitempty"`
	IdempotencyKey               string   `json:"idempotency_key"`
	RemainingOnlineCandidateSize int      `json:"remaining_online_candidate_size"`
}

type DispatchEvent struct {
	ID                       string    `json:"id"`
	OrderID                  string    `json:"order_id"`
	StationID                string    `json:"station_id"`
	Mode                     string    `json:"mode"`
	Type                     string    `json:"type"`
	RiderID                  string    `json:"rider_id,omitempty"`
	ActorID                  string    `json:"actor_id,omitempty"`
	Reason                   string    `json:"reason,omitempty"`
	IdempotencyKey           string    `json:"idempotency_key"`
	OnlineCandidateSize      int       `json:"online_candidate_size"`
	RejectedRiderIDs         []string  `json:"rejected_rider_ids"`
	CanDeclineWithoutPenalty bool      `json:"can_decline_without_penalty"`
	CreatedAt                time.Time `json:"created_at"`
}

type OutboxEvent struct {
	ID             string         `json:"id"`
	Topic          string         `json:"topic"`
	AggregateType  string         `json:"aggregate_type"`
	AggregateID    string         `json:"aggregate_id"`
	EventType      string         `json:"event_type"`
	IdempotencyKey string         `json:"idempotency_key"`
	Payload        map[string]any `json:"payload"`
	Status         string         `json:"status"`
	Attempts       int            `json:"attempts"`
	LastError      string         `json:"last_error,omitempty"`
	AvailableAt    time.Time      `json:"available_at"`
	LeaseOwner     string         `json:"lease_owner,omitempty"`
	LeaseExpiresAt time.Time      `json:"lease_expires_at,omitempty"`
	PublishedAt    time.Time      `json:"published_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type OutboxEventsRequest struct {
	Status string    `json:"status"`
	Topic  string    `json:"topic"`
	Limit  int       `json:"limit"`
	Now    time.Time `json:"now"`
}

type OutboxStatsRequest struct {
	Topic                      string    `json:"topic"`
	Now                        time.Time `json:"now"`
	LeaseExpiringWithinSeconds int       `json:"lease_expiring_within_seconds"`
}

type OutboxLeaseOwnerStats struct {
	Owner                     string    `json:"owner"`
	Leased                    int       `json:"leased"`
	LeaseExpiringSoon         int       `json:"lease_expiring_soon"`
	NextLeaseExpiresAt        time.Time `json:"next_lease_expires_at,omitempty"`
	NextLeaseExpiresInSeconds int64     `json:"next_lease_expires_in_seconds"`
}

type OutboxTopicStats struct {
	Topic                     string    `json:"topic"`
	Total                     int       `json:"total"`
	Pending                   int       `json:"pending"`
	Failed                    int       `json:"failed"`
	DeadLetter                int       `json:"dead_letter"`
	Published                 int       `json:"published"`
	Leased                    int       `json:"leased"`
	LeaseExpiringSoon         int       `json:"lease_expiring_soon"`
	Ready                     int       `json:"ready"`
	Blocked                   int       `json:"blocked"`
	OldestReadyLagSeconds     int64     `json:"oldest_ready_lag_seconds"`
	NextLeaseExpiresAt        time.Time `json:"next_lease_expires_at,omitempty"`
	NextLeaseExpiresInSeconds int64     `json:"next_lease_expires_in_seconds"`
}

type OutboxStats struct {
	GeneratedAt                time.Time               `json:"generated_at"`
	Topic                      string                  `json:"topic,omitempty"`
	Total                      int                     `json:"total"`
	Pending                    int                     `json:"pending"`
	Failed                     int                     `json:"failed"`
	DeadLetter                 int                     `json:"dead_letter"`
	Published                  int                     `json:"published"`
	Leased                     int                     `json:"leased"`
	LeaseExpiringWithinSeconds int                     `json:"lease_expiring_within_seconds"`
	LeaseExpiringSoon          int                     `json:"lease_expiring_soon"`
	Ready                      int                     `json:"ready"`
	Blocked                    int                     `json:"blocked"`
	OldestReadyAt              time.Time               `json:"oldest_ready_at,omitempty"`
	OldestReadyLagSeconds      int64                   `json:"oldest_ready_lag_seconds"`
	NextLeaseExpiresAt         time.Time               `json:"next_lease_expires_at,omitempty"`
	NextLeaseExpiresInSeconds  int64                   `json:"next_lease_expires_in_seconds"`
	LeaseOwners                []OutboxLeaseOwnerStats `json:"lease_owners"`
	Topics                     []OutboxTopicStats      `json:"topics"`
}

type OutboxEventDetailRequest struct {
	EventID    string    `json:"event_id"`
	Now        time.Time `json:"now"`
	AuditLimit int       `json:"audit_limit"`
}

type OutboxPayloadField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OutboxRelatedTarget struct {
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	Source       string `json:"source"`
	OperationKey string `json:"operation_key,omitempty"`
}

type OutboxRecommendedOperation struct {
	Key    string         `json:"key"`
	Title  string         `json:"title"`
	Reason string         `json:"reason"`
	Values map[string]any `json:"values"`
}

type OutboxEventAuditFilter struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Action     string `json:"action,omitempty"`
	Limit      int    `json:"limit"`
}

type OutboxEventDetail struct {
	GeneratedAt             time.Time                  `json:"generated_at"`
	Event                   OutboxEvent                `json:"event"`
	IncidentCode            string                     `json:"incident_code"`
	IncidentSeverity        string                     `json:"incident_severity"`
	Ready                   bool                       `json:"ready"`
	Blocked                 bool                       `json:"blocked"`
	LeaseActive             bool                       `json:"lease_active"`
	RetryAvailableInSeconds int64                      `json:"retry_available_in_seconds"`
	LeaseExpiresInSeconds   int64                      `json:"lease_expires_in_seconds"`
	PayloadSummary          []OutboxPayloadField       `json:"payload_summary"`
	RelatedTargets          []OutboxRelatedTarget      `json:"related_targets"`
	AuditFilters            []OutboxEventAuditFilter   `json:"audit_filters"`
	RecentAudits            []AuditLog                 `json:"recent_audits"`
	RecommendedOperation    OutboxRecommendedOperation `json:"recommended_operation"`
	Checklist               []string                   `json:"checklist"`
}

type MarkOutboxEventPublishedRequest struct {
	EventID     string    `json:"event_id"`
	PublishedAt time.Time `json:"published_at"`
}

type MarkOutboxEventFailedRequest struct {
	EventID           string    `json:"event_id"`
	Error             string    `json:"error"`
	RetryAfterSeconds int       `json:"retry_after_seconds"`
	MaxAttempts       int       `json:"max_attempts"`
	Now               time.Time `json:"now"`
}

type ClaimOutboxEventsRequest struct {
	Topic        string    `json:"topic"`
	Limit        int       `json:"limit"`
	LeaseOwner   string    `json:"lease_owner"`
	LeaseSeconds int       `json:"lease_seconds"`
	Now          time.Time `json:"now"`
}

type ClaimOutboxEventsResult struct {
	Topic          string        `json:"topic,omitempty"`
	Limit          int           `json:"limit"`
	LeaseOwner     string        `json:"lease_owner"`
	LeaseExpiresAt time.Time     `json:"lease_expires_at"`
	Claimed        int           `json:"claimed"`
	Events         []OutboxEvent `json:"events"`
}

type RenewOutboxEventLeaseRequest struct {
	EventID      string    `json:"event_id"`
	LeaseOwner   string    `json:"lease_owner"`
	LeaseSeconds int       `json:"lease_seconds"`
	Now          time.Time `json:"now"`
}

type ReplayOutboxEventRequest struct {
	EventID string    `json:"event_id"`
	Now     time.Time `json:"now"`
}

type ReplayOutboxEventsRequest struct {
	Topic string    `json:"topic"`
	Limit int       `json:"limit"`
	Now   time.Time `json:"now"`
}

type ReplayOutboxEventsResult struct {
	Topic    string        `json:"topic,omitempty"`
	Limit    int           `json:"limit"`
	Replayed int           `json:"replayed"`
	Events   []OutboxEvent `json:"events"`
}

type MerchantProduct struct {
	ID             string   `json:"id"`
	ShopID         string   `json:"shop_id"`
	Name           string   `json:"name"`
	ImageURL       string   `json:"image_url"`
	Description    string   `json:"description"`
	IngredientList []string `json:"ingredient_list"`
	PriceFen       int64    `json:"price_fen"`
	StockCount     int      `json:"stock_count"`
	Status         string   `json:"status"`
}

type UpsertMerchantProductRequest struct {
	MerchantID     string   `json:"merchant_id"`
	ProductID      string   `json:"product_id"`
	ShopID         string   `json:"shop_id"`
	Name           string   `json:"name"`
	ImageURL       string   `json:"image_url"`
	Description    string   `json:"description"`
	IngredientList []string `json:"ingredient_list"`
	PriceFen       int64    `json:"price_fen"`
	StockCount     int      `json:"stock_count"`
	Status         string   `json:"status"`
}

type SetMerchantProductStatusRequest struct {
	MerchantID string `json:"merchant_id"`
	ProductID  string `json:"product_id"`
	Status     string `json:"status"`
}

type CartItem struct {
	UserID       string `json:"user_id"`
	ShopID       string `json:"shop_id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	ImageURL     string `json:"image_url,omitempty"`
	UnitPriceFen int64  `json:"unit_price_fen"`
	Quantity     int    `json:"quantity"`
	Selected     bool   `json:"selected"`
}

type CartSummary struct {
	UserID          string     `json:"user_id"`
	ShopID          string     `json:"shop_id"`
	ShopName        string     `json:"shop_name,omitempty"`
	Items           []CartItem `json:"items"`
	ItemsTotalFen   int64      `json:"items_total_fen"`
	DeliveryFeeFen  int64      `json:"delivery_fee_fen"`
	PackagingFeeFen int64      `json:"packaging_fee_fen"`
	DiscountFen     int64      `json:"discount_fen"`
	PayableFen      int64      `json:"payable_fen"`
}

type CreateGroupbuyOrderRequest struct {
	UserID   string `json:"user_id"`
	ShopID   string `json:"shop_id"`
	DealID   string `json:"deal_id"`
	Quantity int    `json:"quantity"`
}

type GroupbuyVoucher struct {
	ID                   string    `json:"id"`
	VoucherCode          string    `json:"voucher_code"`
	QRPayload            string    `json:"qr_payload"`
	OrderID              string    `json:"order_id"`
	UserID               string    `json:"user_id"`
	ShopID               string    `json:"shop_id"`
	DealID               string    `json:"deal_id"`
	DealName             string    `json:"deal_name"`
	Status               string    `json:"status"`
	RedemptionMethod     string    `json:"redemption_method,omitempty"`
	RedeemedByMerchantID string    `json:"redeemed_by_merchant_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	RedeemedAt           time.Time `json:"redeemed_at,omitempty"`
	ExpiresAt            time.Time `json:"expires_at,omitempty"`
}

type RedeemGroupbuyVoucherRequest struct {
	MerchantID  string `json:"merchant_id"`
	VoucherCode string `json:"voucher_code"`
	QRPayload   string `json:"qr_payload"`
	Method      string `json:"method"`
}

type WalletPaymentPasswordState struct {
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"`
	FailedCount int       `json:"failed_count"`
	LockedUntil time.Time `json:"locked_until"`
}

type GroupChat struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	OwnerID             string `json:"owner_id"`
	OwnerRole           string `json:"owner_role"`
	Name                string `json:"name"`
	NotificationDefault string `json:"notification_default"`
}

type RedPacket struct {
	ID                string    `json:"id"`
	SenderID          string    `json:"sender_id"`
	SenderRole        string    `json:"sender_role"`
	Scene             string    `json:"scene"`
	TargetID          string    `json:"target_id"`
	Type              string    `json:"type"`
	TotalAmountFen    int64     `json:"total_amount_fen"`
	Quantity          int       `json:"quantity"`
	PaymentMethod     string    `json:"payment_method"`
	Message           string    `json:"message,omitempty"`
	Status            string    `json:"status,omitempty"`
	ClaimedAmountFen  int64     `json:"claimed_amount_fen,omitempty"`
	RefundedAmountFen int64     `json:"refunded_amount_fen,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	ExpiresAt         time.Time `json:"expires_at,omitempty"`
	RefundedAt        time.Time `json:"refunded_at,omitempty"`
}

type AfterSalesRequest struct {
	ID                 string    `json:"id"`
	OrderID            string    `json:"order_id"`
	UserID             string    `json:"user_id"`
	ShopName           string    `json:"shop_name,omitempty"`
	OrderStatus        string    `json:"order_status,omitempty"`
	OrderItemSummary   string    `json:"order_item_summary,omitempty"`
	LatestEventMessage string    `json:"latest_event_message,omitempty"`
	LatestEventAt      time.Time `json:"latest_event_at,omitempty"`
	Type               string    `json:"type"`
	Reason             string    `json:"reason"`
	RequestedAmountFen int64     `json:"requested_amount_fen"`
	OrderAmountFen     int64     `json:"order_amount_fen"`
	RefundedAmountFen  int64     `json:"refunded_amount_fen"`
	RefundableFen      int64     `json:"refundable_fen"`
	EvidenceURLs       []string  `json:"evidence_urls"`
	Status             string    `json:"status"`
	ReviewReason       string    `json:"review_reason,omitempty"`
	ReviewerID         string    `json:"reviewer_id,omitempty"`
	ReviewerRole       string    `json:"reviewer_role,omitempty"`
	RefundID           string    `json:"refund_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	ReviewedAt         time.Time `json:"reviewed_at,omitempty"`
}

type AfterSalesEvent struct {
	ID            string    `json:"id"`
	RequestID     string    `json:"request_id"`
	OrderID       string    `json:"order_id"`
	ActorID       string    `json:"actor_id"`
	ActorRole     string    `json:"actor_role"`
	Action        string    `json:"action"`
	Message       string    `json:"message"`
	Attachments   []string  `json:"attachments,omitempty"`
	VisibleToUser bool      `json:"visible_to_user"`
	CreatedAt     time.Time `json:"created_at"`
}

type AfterSalesEvidence struct {
	ID             string    `json:"id"`
	RequestID      string    `json:"request_id"`
	OrderID        string    `json:"order_id"`
	ObjectKey      string    `json:"object_key"`
	PublicURL      string    `json:"public_url"`
	FileName       string    `json:"file_name"`
	ContentType    string    `json:"content_type"`
	SizeBytes      int64     `json:"size_bytes"`
	ContentSHA     string    `json:"content_sha,omitempty"`
	UploadedByID   string    `json:"uploaded_by_id"`
	UploadedByRole string    `json:"uploaded_by_role"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ConfirmedAt    time.Time `json:"confirmed_at"`
}

type AfterSalesEvidenceUploadTicket struct {
	ID                  string    `json:"id"`
	RequestID           string    `json:"request_id"`
	OrderID             string    `json:"order_id"`
	Provider            string    `json:"provider"`
	Bucket              string    `json:"bucket"`
	ObjectKey           string    `json:"object_key"`
	PublicURL           string    `json:"public_url"`
	FileName            string    `json:"file_name"`
	ContentType         string    `json:"content_type"`
	SizeBytes           int64     `json:"size_bytes"`
	MaxSizeBytes        int64     `json:"max_size_bytes"`
	ContentSHA          string    `json:"content_sha,omitempty"`
	UploadedByID        string    `json:"uploaded_by_id"`
	UploadedByRole      string    `json:"uploaded_by_role"`
	Status              string    `json:"status"`
	ScanStatus          string    `json:"scan_status"`
	ScanResult          string    `json:"scan_result,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	ExpiresAt           time.Time `json:"expires_at"`
	UploadedAt          time.Time `json:"uploaded_at,omitempty"`
	ConfirmedAt         time.Time `json:"confirmed_at,omitempty"`
	ScanCheckedAt       time.Time `json:"scan_checked_at,omitempty"`
	CleanupReason       string    `json:"cleanup_reason,omitempty"`
	DeletedAt           time.Time `json:"deleted_at,omitempty"`
	CleanupAttempts     int       `json:"cleanup_attempts,omitempty"`
	LastCleanupError    string    `json:"last_cleanup_error,omitempty"`
	LastCleanupFailedAt time.Time `json:"last_cleanup_failed_at,omitempty"`
}

type CreateAfterSalesRequest struct {
	OrderID            string   `json:"order_id"`
	UserID             string   `json:"user_id,omitempty"`
	Type               string   `json:"type"`
	Reason             string   `json:"reason"`
	RequestedAmountFen int64    `json:"requested_amount_fen,omitempty"`
	EvidenceURLs       []string `json:"evidence_urls,omitempty"`
}

type AfterSalesListRequest struct {
	UserID    string `json:"user_id,omitempty"`
	OrderID   string `json:"order_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

type AfterSalesEventSummary struct {
	Total           int       `json:"total"`
	UserVisible     int       `json:"user_visible"`
	InternalOnly    int       `json:"internal_only"`
	AttachmentCount int       `json:"attachment_count"`
	LatestAction    string    `json:"latest_action,omitempty"`
	LatestEventAt   time.Time `json:"latest_event_at,omitempty"`
}

type AfterSalesEvidenceSummary struct {
	Total             int       `json:"total"`
	ImageCount        int       `json:"image_count"`
	ConfirmedCount    int       `json:"confirmed_count"`
	TotalSizeBytes    int64     `json:"total_size_bytes"`
	LatestConfirmedAt time.Time `json:"latest_confirmed_at,omitempty"`
}

type AfterSalesDispatchSummary struct {
	Total             int       `json:"total"`
	AutoAssignCount   int       `json:"auto_assign_count"`
	ManualAssignCount int       `json:"manual_assign_count"`
	RejectCount       int       `json:"reject_count"`
	TimeoutCount      int       `json:"timeout_count"`
	LatestType        string    `json:"latest_type,omitempty"`
	LatestEventAt     time.Time `json:"latest_event_at,omitempty"`
}

type AfterSalesRefundSummary struct {
	Total             int       `json:"total"`
	SuccessCount      int       `json:"success_count"`
	TotalAmountFen    int64     `json:"total_amount_fen"`
	LatestDestination string    `json:"latest_destination,omitempty"`
	LatestCreatedAt   time.Time `json:"latest_created_at,omitempty"`
}

type AfterSalesServiceTicketSummary struct {
	Total           int       `json:"total"`
	OpenCount       int       `json:"open_count"`
	EscalatedCount  int       `json:"escalated_count"`
	LatestStatus    string    `json:"latest_status,omitempty"`
	LatestUpdatedAt time.Time `json:"latest_updated_at,omitempty"`
}

type AfterSalesAuditSummary struct {
	Total              int       `json:"total"`
	VerifiedCount      int       `json:"verified_count"`
	OrderCount         int       `json:"order_count"`
	AfterSalesCount    int       `json:"after_sales_count"`
	ServiceTicketCount int       `json:"service_ticket_count"`
	LatestAction       string    `json:"latest_action,omitempty"`
	LatestCreatedAt    time.Time `json:"latest_created_at,omitempty"`
}

type AdminOrderAfterSalesSummary struct {
	Total           int       `json:"total"`
	OpenCount       int       `json:"open_count"`
	RefundedCount   int       `json:"refunded_count"`
	LatestStatus    string    `json:"latest_status,omitempty"`
	LatestUpdatedAt time.Time `json:"latest_updated_at,omitempty"`
}

type AdminOrderRefundSummary struct {
	Total             int       `json:"total"`
	SuccessCount      int       `json:"success_count"`
	TotalAmountFen    int64     `json:"total_amount_fen"`
	LatestDestination string    `json:"latest_destination,omitempty"`
	LatestCreatedAt   time.Time `json:"latest_created_at,omitempty"`
}

type AdminOrderServiceTicketSummary struct {
	Total           int       `json:"total"`
	OpenCount       int       `json:"open_count"`
	EscalatedCount  int       `json:"escalated_count"`
	LatestStatus    string    `json:"latest_status,omitempty"`
	LatestUpdatedAt time.Time `json:"latest_updated_at,omitempty"`
}

type AdminOrderDispatchSummary struct {
	Total             int       `json:"total"`
	AutoAssignCount   int       `json:"auto_assign_count"`
	ManualAssignCount int       `json:"manual_assign_count"`
	RejectCount       int       `json:"reject_count"`
	TimeoutCount      int       `json:"timeout_count"`
	LatestType        string    `json:"latest_type,omitempty"`
	LatestEventAt     time.Time `json:"latest_event_at,omitempty"`
}

type AdminOrderAuditSummary struct {
	Total              int       `json:"total"`
	VerifiedCount      int       `json:"verified_count"`
	OrderCount         int       `json:"order_count"`
	AfterSalesCount    int       `json:"after_sales_count"`
	ServiceTicketCount int       `json:"service_ticket_count"`
	LatestAction       string    `json:"latest_action,omitempty"`
	LatestCreatedAt    time.Time `json:"latest_created_at,omitempty"`
}

type AdminOrderDetail struct {
	Order                Order                          `json:"order"`
	AfterSalesRequests   []AfterSalesRequest            `json:"after_sales_requests"`
	Refunds              []RefundTransaction            `json:"refunds"`
	ServiceTickets       []ServiceTicket                `json:"service_tickets"`
	DispatchEvents       []DispatchEvent                `json:"dispatch_events"`
	RelatedAudits        []AuditLog                     `json:"related_audits"`
	AfterSalesSummary    AdminOrderAfterSalesSummary    `json:"after_sales_summary"`
	RefundSummary        AdminOrderRefundSummary        `json:"refund_summary"`
	ServiceTicketSummary AdminOrderServiceTicketSummary `json:"service_ticket_summary"`
	DispatchSummary      AdminOrderDispatchSummary      `json:"dispatch_summary"`
	AuditSummary         AdminOrderAuditSummary         `json:"audit_summary"`
}

type AdminAfterSalesDetail struct {
	Request              AfterSalesRequest              `json:"request"`
	Events               []AfterSalesEvent              `json:"events"`
	Evidence             []AfterSalesEvidence           `json:"evidence"`
	DispatchEvents       []DispatchEvent                `json:"dispatch_events"`
	Refunds              []RefundTransaction            `json:"refunds"`
	ServiceTickets       []ServiceTicket                `json:"service_tickets"`
	RelatedAudits        []AuditLog                     `json:"related_audits"`
	EventSummary         AfterSalesEventSummary         `json:"event_summary"`
	EvidenceSummary      AfterSalesEvidenceSummary      `json:"evidence_summary"`
	DispatchSummary      AfterSalesDispatchSummary      `json:"dispatch_summary"`
	RefundSummary        AfterSalesRefundSummary        `json:"refund_summary"`
	ServiceTicketSummary AfterSalesServiceTicketSummary `json:"service_ticket_summary"`
	AuditSummary         AfterSalesAuditSummary         `json:"audit_summary"`
}

type AddAfterSalesEventRequest struct {
	RequestID     string   `json:"request_id"`
	ActorID       string   `json:"actor_id,omitempty"`
	ActorRole     string   `json:"actor_role,omitempty"`
	Action        string   `json:"action"`
	Message       string   `json:"message"`
	Attachments   []string `json:"attachments,omitempty"`
	VisibleToUser *bool    `json:"visible_to_user,omitempty"`
}

type CreateAfterSalesEvidenceUploadRequest struct {
	RequestID   string `json:"request_id"`
	ActorID     string `json:"actor_id,omitempty"`
	ActorRole   string `json:"actor_role,omitempty"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type ConfirmAfterSalesEvidenceUploadRequest struct {
	RequestID   string `json:"request_id"`
	ActorID     string `json:"actor_id,omitempty"`
	ActorRole   string `json:"actor_role,omitempty"`
	TicketID    string `json:"ticket_id,omitempty"`
	ObjectKey   string `json:"object_key"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentSHA  string `json:"content_sha,omitempty"`
	Message     string `json:"message,omitempty"`
}

type ObjectStorageUploadCallbackRequest struct {
	TicketID    string    `json:"ticket_id"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	ContentSHA  string    `json:"content_sha,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at,omitempty"`
	Signature   string    `json:"signature,omitempty"`
}

type ObjectStorageScanResultRequest struct {
	TicketID      string    `json:"ticket_id"`
	ObjectKey     string    `json:"object_key"`
	ScanStatus    string    `json:"scan_status"`
	ScanResult    string    `json:"scan_result,omitempty"`
	Scanner       string    `json:"scanner,omitempty"`
	ScanCheckedAt time.Time `json:"scan_checked_at,omitempty"`
	Signature     string    `json:"signature,omitempty"`
}

type ObjectStorageCleanupCandidatesRequest struct {
	Limit        int       `json:"limit,omitempty"`
	GraceSeconds int64     `json:"grace_seconds,omitempty"`
	Now          time.Time `json:"now,omitempty"`
}

type ObjectStorageCleanupCompleteRequest struct {
	TicketID  string    `json:"ticket_id"`
	ObjectKey string    `json:"object_key"`
	Reason    string    `json:"reason"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
}

type ObjectStorageCleanupFailureRequest struct {
	TicketID  string    `json:"ticket_id"`
	ObjectKey string    `json:"object_key"`
	Reason    string    `json:"reason"`
	Error     string    `json:"error"`
	FailedAt  time.Time `json:"failed_at,omitempty"`
}

type ObjectStorageCleanupCandidate struct {
	TicketID            string    `json:"ticket_id"`
	RequestID           string    `json:"request_id"`
	OrderID             string    `json:"order_id"`
	Provider            string    `json:"provider"`
	Bucket              string    `json:"bucket"`
	ObjectKey           string    `json:"object_key"`
	PublicURL           string    `json:"public_url"`
	Status              string    `json:"status"`
	ScanStatus          string    `json:"scan_status"`
	Reason              string    `json:"reason"`
	RetainUntil         time.Time `json:"retain_until"`
	CreatedAt           time.Time `json:"created_at"`
	ExpiresAt           time.Time `json:"expires_at"`
	UploadedAt          time.Time `json:"uploaded_at,omitempty"`
	ScanCheckedAt       time.Time `json:"scan_checked_at,omitempty"`
	CleanupAttempts     int       `json:"cleanup_attempts,omitempty"`
	LastCleanupError    string    `json:"last_cleanup_error,omitempty"`
	LastCleanupFailedAt time.Time `json:"last_cleanup_failed_at,omitempty"`
}

type ObjectStorageCleanupStats struct {
	Pending             int       `json:"pending"`
	ExpiredUnconfirmed  int       `json:"expired_unconfirmed"`
	ScanRejected        int       `json:"scan_rejected"`
	Failed              int       `json:"failed"`
	Deleted             int       `json:"deleted"`
	CleanupAttempts     int       `json:"cleanup_attempts"`
	LastCleanupFailedAt time.Time `json:"last_cleanup_failed_at,omitempty"`
	LastDeletedAt       time.Time `json:"last_deleted_at,omitempty"`
}

type AdminOperationsSnapshotRequest struct {
	Now                        time.Time `json:"now"`
	Limit                      int       `json:"limit"`
	StationManagerID           string    `json:"station_manager_id,omitempty"`
	LeaseExpiringWithinSeconds int       `json:"lease_expiring_within_seconds"`
	ObjectCleanupGraceSeconds  int       `json:"object_cleanup_grace_seconds"`
}

type AdminOperationsCounts struct {
	TotalOrders                 int `json:"total_orders"`
	PendingMerchantOrders       int `json:"pending_merchant_orders"`
	DispatchingOrders           int `json:"dispatching_orders"`
	RiderAssignedOrders         int `json:"rider_assigned_orders"`
	CompletedOrders             int `json:"completed_orders"`
	RefundedOrders              int `json:"refunded_orders"`
	ExceptionOrders             int `json:"exception_orders"`
	TotalMerchants              int `json:"total_merchants"`
	MerchantQualificationRisks  int `json:"merchant_qualification_risks"`
	MerchantDepositMissing      int `json:"merchant_deposit_missing"`
	TotalRiders                 int `json:"total_riders"`
	OnlineRiders                int `json:"online_riders"`
	RiderDepositMissing         int `json:"rider_deposit_missing"`
	StationManagers             int `json:"station_managers"`
	AfterSalesPending           int `json:"after_sales_pending"`
	AfterSalesAdminReview       int `json:"after_sales_admin_review"`
	DispatchEventCount          int `json:"dispatch_event_count"`
	OutboxReady                 int `json:"outbox_ready"`
	OutboxBlocked               int `json:"outbox_blocked"`
	ObjectCleanupFailed         int `json:"object_cleanup_failed"`
	ObjectCleanupTotalCandidate int `json:"object_cleanup_total_candidate"`
}

type AdminMerchantSnapshot struct {
	Account                    MerchantAccount                `json:"account"`
	Shops                      []Shop                         `json:"shops"`
	Qualifications             []MerchantQualification        `json:"qualifications"`
	MissingQualifications      []string                       `json:"missing_qualifications"`
	StaffCount                 int                            `json:"staff_count"`
	SupplementalMaterialCount  int                            `json:"supplemental_material_count"`
	Deposit                    *DepositAccount                `json:"deposit,omitempty"`
	CanAcceptOrders            bool                           `json:"can_accept_orders"`
	QualificationPopupRequired bool                           `json:"qualification_popup_required"`
	QualificationPopupCode     string                         `json:"qualification_popup_code,omitempty"`
	SupplementalMaterials      []MerchantSupplementalMaterial `json:"supplemental_materials,omitempty"`
}

type AdminOperationsSnapshot struct {
	GeneratedAt               time.Time                 `json:"generated_at"`
	Counts                    AdminOperationsCounts     `json:"counts"`
	Orders                    []Order                   `json:"orders"`
	Merchants                 []AdminMerchantSnapshot   `json:"merchants"`
	Riders                    []RiderAccount            `json:"riders"`
	RiderPerformance          []RiderPerformance        `json:"rider_performance"`
	AfterSales                []AfterSalesRequest       `json:"after_sales"`
	DispatchEvents            []DispatchEvent           `json:"dispatch_events"`
	RefundSettings            RefundSettings            `json:"refund_settings"`
	OutboxStats               OutboxStats               `json:"outbox_stats"`
	ObjectStorageCleanupStats ObjectStorageCleanupStats `json:"object_storage_cleanup_stats"`
}

type ObjectUploadTicket struct {
	TicketID     string            `json:"ticket_id"`
	Provider     string            `json:"provider"`
	Bucket       string            `json:"bucket"`
	ObjectKey    string            `json:"object_key"`
	UploadURL    string            `json:"upload_url"`
	PublicURL    string            `json:"public_url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	ExpiresAt    time.Time         `json:"expires_at"`
	MaxSizeBytes int64             `json:"max_size_bytes"`
}

type ReviewAfterSalesRequest struct {
	RequestID            string `json:"request_id"`
	Decision             string `json:"decision"`
	Reason               string `json:"reason"`
	ActorID              string `json:"actor_id,omitempty"`
	ActorRole            string `json:"actor_role,omitempty"`
	RefundDestination    string `json:"refund_destination,omitempty"`
	RefundIdempotencyKey string `json:"refund_idempotency_key,omitempty"`
}

type RefundSettings struct {
	DefaultStrategy string `json:"default_refund_strategy"`
}

type SaveRefundSettingsRequest struct {
	DefaultStrategy string `json:"default_refund_strategy"`
}

type RefundOrderRequest struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id,omitempty"`
	AmountFen      int64  `json:"amount_fen,omitempty"`
	Destination    string `json:"destination,omitempty"`
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotency_key"`
	ActorID        string `json:"actor_id,omitempty"`
	ActorRole      string `json:"actor_role,omitempty"`
}

type RefundTransaction struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	AmountFen      int64     `json:"amount_fen"`
	Destination    string    `json:"destination"`
	Status         string    `json:"status"`
	Reason         string    `json:"reason"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type RefundTransactionListRequest struct {
	OrderID     string `json:"order_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Destination string `json:"destination,omitempty"`
	Status      string `json:"status,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type Review struct {
	ID          string             `json:"id"`
	OrderID     string             `json:"order_id,omitempty"`
	TargetType  string             `json:"target_type"`
	TargetID    string             `json:"target_id"`
	UserID      string             `json:"user_id"`
	Rating      int                `json:"rating"`
	RiderRating int                `json:"rider_rating,omitempty"`
	Content     string             `json:"content"`
	ImageURLs   []string           `json:"image_urls"`
	ItemRatings []ReviewItemRating `json:"item_ratings,omitempty"`
	Anonymous   bool               `json:"anonymous"`
	Status      string             `json:"status"`
	Tags        []string           `json:"tags,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
}

type ReviewItemRating struct {
	ProductID   string   `json:"product_id,omitempty"`
	ProductName string   `json:"product_name,omitempty"`
	Rating      int      `json:"rating"`
	Tags        []string `json:"tags,omitempty"`
}

type ReviewListRequest struct {
	UserID  string `json:"user_id,omitempty"`
	OrderID string `json:"order_id,omitempty"`
}

type ReviewImageUploadTicket struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	OrderID       string    `json:"order_id,omitempty"`
	Provider      string    `json:"provider"`
	Bucket        string    `json:"bucket"`
	ObjectKey     string    `json:"object_key"`
	PublicURL     string    `json:"public_url"`
	FileName      string    `json:"file_name"`
	ContentType   string    `json:"content_type"`
	SizeBytes     int64     `json:"size_bytes"`
	MaxSizeBytes  int64     `json:"max_size_bytes"`
	ContentSHA    string    `json:"content_sha,omitempty"`
	Status        string    `json:"status"`
	ScanStatus    string    `json:"scan_status"`
	ScanResult    string    `json:"scan_result,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	UploadedAt    time.Time `json:"uploaded_at,omitempty"`
	ConfirmedAt   time.Time `json:"confirmed_at,omitempty"`
	ScanCheckedAt time.Time `json:"scan_checked_at,omitempty"`
}

type CreateReviewImageUploadRequest struct {
	UserID      string `json:"user_id,omitempty"`
	OrderID     string `json:"order_id,omitempty"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type ConfirmReviewImageUploadRequest struct {
	UserID      string `json:"user_id,omitempty"`
	TicketID    string `json:"ticket_id"`
	ObjectKey   string `json:"object_key"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentSHA  string `json:"content_sha,omitempty"`
}

type FeedbackTicket struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Contact   string    `json:"contact,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServiceTicket struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	Type                string    `json:"type"`
	Category            string    `json:"category"`
	Title               string    `json:"title"`
	Content             string    `json:"content"`
	Contact             string    `json:"contact,omitempty"`
	RelatedOrderID      string    `json:"related_order_id,omitempty"`
	RelatedOrderTitle   string    `json:"related_order_title,omitempty"`
	RelatedOrderStatus  string    `json:"related_order_status,omitempty"`
	Severity            string    `json:"severity,omitempty"`
	Status              string    `json:"status"`
	SLAStatus           string    `json:"sla_status,omitempty"`
	Solution            string    `json:"solution,omitempty"`
	AssignedSupportID   string    `json:"assigned_support_id,omitempty"`
	AssignedSupportName string    `json:"assigned_support_name,omitempty"`
	EscalationLevel     string    `json:"escalation_level,omitempty"`
	EscalationReason    string    `json:"escalation_reason,omitempty"`
	Attachments         []string  `json:"attachments,omitempty"`
	RiskState           string    `json:"risk_state,omitempty"`
	RiskReasonCode      string    `json:"risk_reason_code,omitempty"`
	RiskReason          string    `json:"risk_reason,omitempty"`
	RiskCheckedAt       time.Time `json:"risk_checked_at,omitempty"`
	ReplyDueAt          time.Time `json:"reply_due_at,omitempty"`
	AssignedAt          time.Time `json:"assigned_at,omitempty"`
	EscalatedAt         time.Time `json:"escalated_at,omitempty"`
	ResolvedAt          time.Time `json:"resolved_at,omitempty"`
	ClosedAt            time.Time `json:"closed_at,omitempty"`
	FollowUpRating      int       `json:"follow_up_rating,omitempty"`
	FollowUpComment     string    `json:"follow_up_comment,omitempty"`
	FollowUpAt          time.Time `json:"follow_up_at,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ServiceTicketEvent struct {
	ID             string    `json:"id"`
	TicketID       string    `json:"ticket_id"`
	ActorID        string    `json:"actor_id,omitempty"`
	ActorRole      string    `json:"actor_role,omitempty"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	Status         string    `json:"status"`
	Attachments    []string  `json:"attachments,omitempty"`
	RiskState      string    `json:"risk_state,omitempty"`
	RiskReasonCode string    `json:"risk_reason_code,omitempty"`
	RiskReason     string    `json:"risk_reason,omitempty"`
	RiskCheckedAt  time.Time `json:"risk_checked_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateServiceTicketRequest struct {
	UserID             string   `json:"user_id,omitempty"`
	Type               string   `json:"type"`
	Category           string   `json:"category"`
	Title              string   `json:"title"`
	Content            string   `json:"content"`
	Contact            string   `json:"contact,omitempty"`
	RelatedOrderID     string   `json:"related_order_id,omitempty"`
	RelatedOrderTitle  string   `json:"related_order_title,omitempty"`
	RelatedOrderStatus string   `json:"related_order_status,omitempty"`
	Severity           string   `json:"severity,omitempty"`
	Attachments        []string `json:"attachments,omitempty"`
}

type AddServiceTicketEventRequest struct {
	TicketID    string   `json:"ticket_id"`
	ActorID     string   `json:"actor_id,omitempty"`
	ActorRole   string   `json:"actor_role,omitempty"`
	Title       string   `json:"title"`
	Message     string   `json:"message"`
	Status      string   `json:"status,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type ServiceTicketListRequest struct {
	UserID            string `json:"user_id,omitempty"`
	RelatedOrderID    string `json:"related_order_id,omitempty"`
	Status            string `json:"status,omitempty"`
	SLAStatus         string `json:"sla_status,omitempty"`
	AssignedSupportID string `json:"assigned_support_id,omitempty"`
	Limit             int    `json:"limit,omitempty"`
	Now               time.Time
}

type AssignServiceTicketRequest struct {
	TicketID    string `json:"ticket_id"`
	SupportID   string `json:"support_id"`
	SupportName string `json:"support_name,omitempty"`
	ActorID     string `json:"actor_id,omitempty"`
}

type ResolveServiceTicketRequest struct {
	TicketID string `json:"ticket_id"`
	ActorID  string `json:"actor_id,omitempty"`
	Solution string `json:"solution"`
}

type EscalateServiceTicketRequest struct {
	TicketID        string    `json:"ticket_id"`
	ActorID         string    `json:"actor_id,omitempty"`
	EscalationLevel string    `json:"escalation_level,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	Now             time.Time `json:"now,omitempty"`
}

type ServiceTicketQualityReview struct {
	ID                string    `json:"id"`
	TicketID          string    `json:"ticket_id"`
	SupportID         string    `json:"support_id,omitempty"`
	SupportName       string    `json:"support_name,omitempty"`
	ReviewerID        string    `json:"reviewer_id,omitempty"`
	ReviewerName      string    `json:"reviewer_name,omitempty"`
	Score             int       `json:"score"`
	Result            string    `json:"result"`
	Notes             string    `json:"notes,omitempty"`
	CoachingRequired  bool      `json:"coaching_required,omitempty"`
	TicketTitle       string    `json:"ticket_title,omitempty"`
	TicketCategory    string    `json:"ticket_category,omitempty"`
	TicketSLAStatus   string    `json:"ticket_sla_status,omitempty"`
	TicketFollowUp    int       `json:"ticket_follow_up,omitempty"`
	TicketResolvedAt  time.Time `json:"ticket_resolved_at,omitempty"`
	TicketEscalatedAt time.Time `json:"ticket_escalated_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type ServiceTicketQualityReviewRequest struct {
	TicketID         string    `json:"ticket_id"`
	ReviewerID       string    `json:"reviewer_id,omitempty"`
	ReviewerName     string    `json:"reviewer_name,omitempty"`
	Score            int       `json:"score"`
	Result           string    `json:"result,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	CoachingRequired bool      `json:"coaching_required,omitempty"`
	Now              time.Time `json:"now,omitempty"`
}

type ServiceTicketQualityReviewListRequest struct {
	TicketID         string `json:"ticket_id,omitempty"`
	SupportID        string `json:"support_id,omitempty"`
	ReviewerID       string `json:"reviewer_id,omitempty"`
	Result           string `json:"result,omitempty"`
	CoachingRequired string `json:"coaching_required,omitempty"`
	Limit            int    `json:"limit,omitempty"`
}

type ServiceTicketPerformanceRequest struct {
	SupportID string    `json:"support_id,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Now       time.Time `json:"now,omitempty"`
}

type ServiceTicketPerformanceSummary struct {
	SupportID             string  `json:"support_id"`
	SupportName           string  `json:"support_name,omitempty"`
	AssignedTickets       int     `json:"assigned_tickets"`
	ResolvedTickets       int     `json:"resolved_tickets"`
	ClosedTickets         int     `json:"closed_tickets"`
	EscalatedTickets      int     `json:"escalated_tickets"`
	OverdueTickets        int     `json:"overdue_tickets"`
	AverageFollowUpRating float64 `json:"average_follow_up_rating"`
	QualityReviewCount    int     `json:"quality_review_count"`
	QualityAverageScore   float64 `json:"quality_average_score"`
	CoachingRequiredCount int     `json:"coaching_required_count"`
	SLAComplianceRate     float64 `json:"sla_compliance_rate"`
	RiskLevel             string  `json:"risk_level"`
}

type CloseServiceTicketRequest struct {
	TicketID string `json:"ticket_id"`
	UserID   string `json:"user_id,omitempty"`
	ActorID  string `json:"actor_id,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type FollowUpServiceTicketRequest struct {
	TicketID string `json:"ticket_id"`
	UserID   string `json:"user_id,omitempty"`
	Rating   int    `json:"rating"`
	Comment  string `json:"comment,omitempty"`
}

type ServiceTicketDetail struct {
	Ticket ServiceTicket        `json:"ticket"`
	Events []ServiceTicketEvent `json:"events"`
}

type RedPacketShare struct {
	UserID    string    `json:"user_id"`
	AmountFen int64     `json:"amount_fen"`
	ClaimedAt time.Time `json:"claimed_at,omitempty"`
}

type RedPacketDetail struct {
	Packet RedPacket           `json:"packet"`
	Shares []RedPacketShare    `json:"shares"`
	Risk   *RedPacketRiskCheck `json:"risk,omitempty"`
}

type RedPacketClaimResult struct {
	Detail RedPacketDetail     `json:"detail"`
	Share  RedPacketShare      `json:"share"`
	Risk   *RedPacketRiskCheck `json:"risk,omitempty"`
}

type RedPacketRiskCheck struct {
	State          string    `json:"state"`
	ReasonCode     string    `json:"reason_code,omitempty"`
	Reason         string    `json:"reason,omitempty"`
	WindowSeconds  int       `json:"window_seconds,omitempty"`
	ClaimCount     int       `json:"claim_count,omitempty"`
	ClaimLimit     int       `json:"claim_limit,omitempty"`
	AmountFen      int64     `json:"amount_fen,omitempty"`
	AmountLimitFen int64     `json:"amount_limit_fen,omitempty"`
	CheckedAt      time.Time `json:"checked_at,omitempty"`
}

type Favorite struct {
	UserID     string `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
}

type PointsTransaction struct {
	ID        string    `json:"id,omitempty"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title,omitempty"`
	Points    int       `json:"points"`
	SourceID  string    `json:"source_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type PushDelivery struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Channel     string `json:"channel"`
	TemplateKey string `json:"template_key"`
	Status      string `json:"status"`
	RetryCount  int    `json:"retry_count"`
}

type PlatformNotification struct {
	ID             string    `json:"id"`
	TargetRole     string    `json:"target_role"`
	TargetID       string    `json:"target_id"`
	Type           string    `json:"type"`
	Channel        string    `json:"channel"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Status         string    `json:"status"`
	SourceTopic    string    `json:"source_topic,omitempty"`
	SourceEventID  string    `json:"source_event_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	ReadAt         time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateNotificationRequest struct {
	TargetRole     string    `json:"target_role"`
	TargetID       string    `json:"target_id"`
	Type           string    `json:"type"`
	Channel        string    `json:"channel,omitempty"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	SourceTopic    string    `json:"source_topic,omitempty"`
	SourceEventID  string    `json:"source_event_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

type NotificationListRequest struct {
	TargetRole    string `json:"target_role"`
	TargetID      string `json:"target_id"`
	Status        string `json:"status,omitempty"`
	SourceTopic   string `json:"source_topic,omitempty"`
	SourceEventID string `json:"source_event_id,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

type MarkNotificationReadRequest struct {
	NotificationID string    `json:"notification_id"`
	TargetRole     string    `json:"target_role"`
	TargetID       string    `json:"target_id"`
	ReadAt         time.Time `json:"read_at,omitempty"`
}

type PlatformNotificationDelivery struct {
	ID                string    `json:"id"`
	NotificationID    string    `json:"notification_id"`
	TargetRole        string    `json:"target_role"`
	TargetID          string    `json:"target_id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	Status            string    `json:"status"`
	ProviderMessageID string    `json:"provider_message_id,omitempty"`
	ErrorCode         string    `json:"error_code,omitempty"`
	ErrorMessage      string    `json:"error_message,omitempty"`
	IdempotencyKey    string    `json:"idempotency_key"`
	AttemptedAt       time.Time `json:"attempted_at"`
	DeliveredAt       time.Time `json:"delivered_at,omitempty"`
	RetryAt           time.Time `json:"retry_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RecordNotificationDeliveryRequest struct {
	NotificationID    string    `json:"notification_id"`
	Channel           string    `json:"channel,omitempty"`
	Provider          string    `json:"provider,omitempty"`
	Status            string    `json:"status"`
	ProviderMessageID string    `json:"provider_message_id,omitempty"`
	ErrorCode         string    `json:"error_code,omitempty"`
	ErrorMessage      string    `json:"error_message,omitempty"`
	IdempotencyKey    string    `json:"idempotency_key"`
	AttemptedAt       time.Time `json:"attempted_at,omitempty"`
	DeliveredAt       time.Time `json:"delivered_at,omitempty"`
	RetryAt           time.Time `json:"retry_at,omitempty"`
}

type NotificationDeliveryListRequest struct {
	NotificationID string    `json:"notification_id,omitempty"`
	TargetRole     string    `json:"target_role,omitempty"`
	TargetID       string    `json:"target_id,omitempty"`
	Channel        string    `json:"channel,omitempty"`
	Provider       string    `json:"provider,omitempty"`
	Status         string    `json:"status,omitempty"`
	ErrorCode      string    `json:"error_code,omitempty"`
	RetryAtBefore  time.Time `json:"retry_at_before,omitempty"`
	Limit          int       `json:"limit,omitempty"`
}

type NotificationFailureAlertEmissionRequest struct {
	TargetRole string    `json:"target_role,omitempty"`
	TargetID   string    `json:"target_id,omitempty"`
	Channel    string    `json:"channel,omitempty"`
	Provider   string    `json:"provider,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Now        time.Time `json:"now,omitempty"`
}

type NotificationFailureAlertEmission struct {
	Status         string                         `json:"status"`
	FailedCount    int                            `json:"failed_count"`
	Topic          string                         `json:"topic"`
	OutboxEventID  string                         `json:"outbox_event_id"`
	IdempotencyKey string                         `json:"idempotency_key"`
	TargetRole     string                         `json:"target_role,omitempty"`
	TargetID       string                         `json:"target_id,omitempty"`
	Channel        string                         `json:"channel,omitempty"`
	Provider       string                         `json:"provider,omitempty"`
	EmittedAt      time.Time                      `json:"emitted_at"`
	Deliveries     []PlatformNotificationDelivery `json:"deliveries"`
}

type NotificationDeliveryRetryScheduleRequest struct {
	TargetRole          string    `json:"target_role,omitempty"`
	TargetID            string    `json:"target_id,omitempty"`
	Channel             string    `json:"channel,omitempty"`
	Provider            string    `json:"provider,omitempty"`
	Status              string    `json:"status,omitempty"`
	ErrorCode           string    `json:"error_code,omitempty"`
	Limit               int       `json:"limit,omitempty"`
	RetryAfterSeconds   int       `json:"retry_after_seconds,omitempty"`
	RetryAt             time.Time `json:"retry_at,omitempty"`
	SourceRetryAtBefore time.Time `json:"source_retry_at_before,omitempty"`
	Now                 time.Time `json:"now,omitempty"`
}

type NotificationQuietWindowRetryScheduleRequest struct {
	TargetRole        string    `json:"target_role,omitempty"`
	TargetID          string    `json:"target_id,omitempty"`
	Channel           string    `json:"channel,omitempty"`
	Provider          string    `json:"provider,omitempty"`
	Limit             int       `json:"limit,omitempty"`
	RetryAfterSeconds int       `json:"retry_after_seconds,omitempty"`
	Now               time.Time `json:"now,omitempty"`
}

type NotificationDeliveryRetrySchedule struct {
	Status            string                         `json:"status"`
	ScheduledCount    int                            `json:"scheduled_count"`
	Topic             string                         `json:"topic"`
	OutboxEventID     string                         `json:"outbox_event_id"`
	IdempotencyKey    string                         `json:"idempotency_key"`
	TargetRole        string                         `json:"target_role,omitempty"`
	TargetID          string                         `json:"target_id,omitempty"`
	Channel           string                         `json:"channel,omitempty"`
	Provider          string                         `json:"provider,omitempty"`
	DeliveryStatus    string                         `json:"delivery_status"`
	ErrorCode         string                         `json:"error_code,omitempty"`
	RetryPolicy       string                         `json:"retry_policy"`
	RetryAfterSeconds int                            `json:"retry_after_seconds"`
	RetryAt           time.Time                      `json:"retry_at"`
	ScheduledAt       time.Time                      `json:"scheduled_at"`
	Deliveries        []PlatformNotificationDelivery `json:"deliveries"`
	Notifications     []PlatformNotification         `json:"notifications,omitempty"`
}

type NotificationQuietHours struct {
	Enabled        bool     `json:"enabled"`
	Start          string   `json:"start,omitempty"`
	End            string   `json:"end,omitempty"`
	TimezoneOffset string   `json:"timezone_offset,omitempty"`
	Channels       []string `json:"channels,omitempty"`
	ExemptTypes    []string `json:"exempt_types,omitempty"`
	Status         string   `json:"status,omitempty"`
}

type NotificationPreference struct {
	ID               string                 `json:"id"`
	PreferenceKey    string                 `json:"preference_key"`
	TargetRole       string                 `json:"target_role,omitempty"`
	TargetID         string                 `json:"target_id,omitempty"`
	NotificationType string                 `json:"notification_type,omitempty"`
	EnabledChannels  []string               `json:"enabled_channels,omitempty"`
	DisabledChannels []string               `json:"disabled_channels,omitempty"`
	QuietHours       NotificationQuietHours `json:"quiet_hours"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type SaveNotificationPreferenceRequest struct {
	TargetRole       string                 `json:"target_role,omitempty"`
	TargetID         string                 `json:"target_id,omitempty"`
	NotificationType string                 `json:"notification_type,omitempty"`
	EnabledChannels  []string               `json:"enabled_channels,omitempty"`
	DisabledChannels []string               `json:"disabled_channels,omitempty"`
	QuietHours       NotificationQuietHours `json:"quiet_hours,omitempty"`
	UpdatedAt        time.Time              `json:"updated_at,omitempty"`
}

type SaveNotificationPreferenceBatchRequest struct {
	Preferences []SaveNotificationPreferenceRequest `json:"preferences"`
	Reason      string                              `json:"reason,omitempty"`
	UpdatedAt   time.Time                           `json:"updated_at,omitempty"`
}

type NotificationPreferenceBatchSaveResult struct {
	BatchID        string                   `json:"batch_id"`
	Preferences    []NotificationPreference `json:"preferences"`
	Saved          int                      `json:"saved"`
	PreferenceKeys []string                 `json:"preference_keys"`
	Reason         string                   `json:"reason,omitempty"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type NotificationPreferenceListRequest struct {
	PreferenceKey    string `json:"preference_key,omitempty"`
	TargetRole       string `json:"target_role,omitempty"`
	TargetID         string `json:"target_id,omitempty"`
	NotificationType string `json:"notification_type,omitempty"`
	Limit            int    `json:"limit,omitempty"`
}

type RiskEvent struct {
	Type string `json:"type"`
}

type CircleFeatureConfig struct {
	Enabled            bool   `json:"enabled"`
	Mode               string `json:"mode"`
	HomeModuleEnabled  bool   `json:"home_module_enabled"`
	MealMatchEnabled   bool   `json:"meal_match_enabled"`
	ModerationRequired bool   `json:"moderation_required"`
}

type CirclePost struct {
	ID           string    `json:"id"`
	AuthorUserID string    `json:"author_user_id"`
	AuthorName   string    `json:"author_name,omitempty"`
	CircleID     string    `json:"circle_id"`
	Type         string    `json:"type"`
	Title        string    `json:"title,omitempty"`
	Content      string    `json:"content"`
	ImageURLs    []string  `json:"image_urls"`
	Status       string    `json:"status"`
	Tags         []string  `json:"tags"`
	DistanceText string    `json:"distance_text,omitempty"`
	LikeCount    int       `json:"like_count,omitempty"`
	CommentCount int       `json:"comment_count,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

type MealMatchProfile struct {
	UserID                         string    `json:"user_id"`
	Gender                         string    `json:"gender"`
	SchoolID                       string    `json:"school_id,omitempty"`
	SchoolName                     string    `json:"school_name,omitempty"`
	CampusName                     string    `json:"campus_name,omitempty"`
	BuildingID                     string    `json:"building_id,omitempty"`
	BuildingName                   string    `json:"building_name,omitempty"`
	PrivacyScope                   string    `json:"privacy_scope,omitempty"`
	LocationPrecision              string    `json:"location_precision,omitempty"`
	IdentityTruthSigned            bool      `json:"identity_truth_signed"`
	PlatformLiabilityReleaseSigned bool      `json:"platform_liability_release_signed"`
	QuestionnaireCompleted         bool      `json:"questionnaire_completed"`
	PersonalityTraits              []string  `json:"personality_traits"`
	DietaryHabits                  []string  `json:"dietary_habits"`
	DeviceID                       string    `json:"device_id,omitempty"`
	DeviceRiskState                string    `json:"device_risk_state,omitempty"`
	DeviceRiskReasonCode           string    `json:"device_risk_reason_code,omitempty"`
	DeviceRiskReason               string    `json:"device_risk_reason,omitempty"`
	DeviceRiskCheckedAt            time.Time `json:"device_risk_checked_at,omitempty"`
	ModerationStatus               string    `json:"moderation_status,omitempty"`
	ModerationReason               string    `json:"moderation_reason,omitempty"`
	ModerationRecordID             string    `json:"moderation_record_id,omitempty"`
	ModerationReviewerID           string    `json:"moderation_reviewer_id,omitempty"`
	ModerationReviewedAt           string    `json:"moderation_reviewed_at,omitempty"`
}

type MealMatchCandidate struct {
	UserID                   string   `json:"user_id"`
	DisplayName              string   `json:"display_name"`
	AvatarInitial            string   `json:"avatar_initial"`
	Gender                   string   `json:"gender"`
	DistanceText             string   `json:"distance_text"`
	SchoolName               string   `json:"school_name,omitempty"`
	CampusName               string   `json:"campus_name,omitempty"`
	BuildingName             string   `json:"building_name,omitempty"`
	SameSchool               bool     `json:"same_school"`
	SameBuilding             bool     `json:"same_building"`
	PrivacyScope             string   `json:"privacy_scope,omitempty"`
	LocationPrecision        string   `json:"location_precision,omitempty"`
	MatchScore               int      `json:"match_score"`
	MatchedPersonalityTraits []string `json:"matched_personality_traits"`
	MatchedDietaryHabits     []string `json:"matched_dietary_habits"`
	PersonalityTraits        []string `json:"personality_traits"`
	DietaryHabits            []string `json:"dietary_habits"`
	SafetyBadges             []string `json:"safety_badges"`
	PrivacyNotice            string   `json:"privacy_notice"`
}

type MealMatchCandidateList struct {
	Profile              MealMatchProfile     `json:"profile"`
	CanUse               bool                 `json:"can_use"`
	Missing              []string             `json:"missing"`
	PrivacyScope         string               `json:"privacy_scope,omitempty"`
	LocationPrecision    string               `json:"location_precision,omitempty"`
	PrivacyNotice        string               `json:"privacy_notice,omitempty"`
	DeviceRiskState      string               `json:"device_risk_state,omitempty"`
	DeviceRiskReasonCode string               `json:"device_risk_reason_code,omitempty"`
	DeviceRiskReason     string               `json:"device_risk_reason,omitempty"`
	DeviceRiskCheckedAt  time.Time            `json:"device_risk_checked_at,omitempty"`
	ModerationStatus     string               `json:"moderation_status,omitempty"`
	ReviewRequired       bool                 `json:"review_required"`
	Candidates           []MealMatchCandidate `json:"candidates"`
}

type MealMatchDeviceRiskCheck struct {
	State      string    `json:"state"`
	ReasonCode string    `json:"reason_code,omitempty"`
	Reason     string    `json:"reason,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
}

type MealMatchReportRequest struct {
	ReporterUserID string `json:"reporter_user_id"`
	TargetUserID   string `json:"target_user_id"`
	Reason         string `json:"reason"`
	Description    string `json:"description"`
}

type MealMatchBlockRequest struct {
	UserID       string `json:"user_id"`
	TargetUserID string `json:"target_user_id"`
	Reason       string `json:"reason"`
}

type MealMatchModerationRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	TargetUserID string    `json:"target_user_id"`
	Action       string    `json:"action"`
	Reason       string    `json:"reason"`
	Description  string    `json:"description,omitempty"`
	Status       string    `json:"status"`
	ReviewerID   string    `json:"reviewer_id,omitempty"`
	ReviewNote   string    `json:"review_note,omitempty"`
	ReviewedAt   time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type MealMatchModerationListRequest struct {
	Status       string `json:"status,omitempty"`
	Action       string `json:"action,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	TargetUserID string `json:"target_user_id,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

type MealMatchModerationReviewRequest struct {
	RecordID   string `json:"record_id"`
	Decision   string `json:"decision"`
	ReviewerID string `json:"reviewer_id,omitempty"`
	ReviewNote string `json:"review_note,omitempty"`
}

type ChatThread struct {
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	Title             string    `json:"title"`
	Subtitle          string    `json:"subtitle"`
	Icon              string    `json:"icon"`
	Route             string    `json:"route,omitempty"`
	LastMessageID     string    `json:"last_message_id,omitempty"`
	LastReadMessageID string    `json:"last_read_message_id,omitempty"`
	UnreadCount       int       `json:"unread_count"`
	Muted             bool      `json:"muted"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
	LastReadAt        time.Time `json:"last_read_at,omitempty"`
}

type ChatThreadMember struct {
	ThreadID    string    `json:"thread_id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Muted       bool      `json:"muted"`
	JoinedAt    time.Time `json:"joined_at,omitempty"`
}

type ChatThreadMemberProfile struct {
	ThreadID    string    `json:"thread_id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	DisplayName string    `json:"display_name"`
	AvatarText  string    `json:"avatar_text"`
	RoleLabel   string    `json:"role_label,omitempty"`
	IsSelf      bool      `json:"is_self,omitempty"`
	JoinedAt    time.Time `json:"joined_at,omitempty"`
}

type ChatMessage struct {
	ID             string    `json:"id"`
	ThreadID       string    `json:"thread_id"`
	SenderID       string    `json:"sender_id"`
	Sender         string    `json:"sender"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	RiskState      string    `json:"risk_state,omitempty"`
	RiskReasonCode string    `json:"risk_reason_code,omitempty"`
	RiskReason     string    `json:"risk_reason,omitempty"`
	RiskCheckedAt  time.Time `json:"risk_checked_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type MessageRiskCheck struct {
	State      string    `json:"state"`
	ReasonCode string    `json:"reason_code,omitempty"`
	Reason     string    `json:"reason,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
}

type ChatReadState struct {
	UserID            string    `json:"user_id"`
	ThreadID          string    `json:"thread_id"`
	LastReadMessageID string    `json:"last_read_message_id,omitempty"`
	ReadAt            time.Time `json:"read_at"`
	UnreadCount       int       `json:"unread_count"`
}

type ChatMessageSyncRequest struct {
	UserID   string `json:"user_id,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`
	SinceID  string `json:"since_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	MarkRead bool   `json:"mark_read,omitempty"`
}

type ChatMessageSyncResult struct {
	ThreadID      string         `json:"thread_id"`
	Messages      []ChatMessage  `json:"messages"`
	LastMessageID string         `json:"last_message_id,omitempty"`
	NextCursor    string         `json:"next_cursor,omitempty"`
	UnreadCount   int            `json:"unread_count"`
	ReadState     *ChatReadState `json:"read_state,omitempty"`
}

type ChatThreadPreference struct {
	ThreadID string `json:"thread_id"`
	Muted    bool   `json:"muted"`
}

type ChatThreadOverview struct {
	ThreadID      string                    `json:"thread_id"`
	Type          string                    `json:"type"`
	Title         string                    `json:"title"`
	Icon          string                    `json:"icon"`
	Summary       string                    `json:"summary"`
	Announcement  string                    `json:"announcement"`
	SettingsText  string                    `json:"settings_text,omitempty"`
	MemberCount   int                       `json:"member_count"`
	Muted         bool                      `json:"muted"`
	MemberPreview []ChatThreadMemberProfile `json:"member_preview,omitempty"`
}

type ChatThreadMembership struct {
	ThreadID          string    `json:"thread_id"`
	Joined            bool      `json:"joined"`
	Muted             bool      `json:"muted"`
	JoinedAt          time.Time `json:"joined_at,omitempty"`
	CanJoin           bool      `json:"can_join"`
	CanLeave          bool      `json:"can_leave"`
	MemberCount       int       `json:"member_count"`
	Summary           string    `json:"summary"`
	CouponRequirement string    `json:"coupon_requirement,omitempty"`
	CouponCode        string    `json:"coupon_code,omitempty"`
}

type ChatThreadJoinRequest struct {
	UserID   string `json:"user_id,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`
}

type ChatThreadLeaveRequest struct {
	UserID   string `json:"user_id,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`
}

type UpdateChatThreadPreferenceRequest struct {
	UserID   string `json:"user_id,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`
	Muted    bool   `json:"muted"`
}

type MarkChatThreadReadRequest struct {
	UserID        string `json:"user_id,omitempty"`
	ThreadID      string `json:"thread_id,omitempty"`
	LastMessageID string `json:"last_message_id,omitempty"`
}

type ChatThreadAccessRequest struct {
	ThreadID    string `json:"thread_id,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	Role        string `json:"role,omitempty"`
}

type ChatThreadAccessResult struct {
	ThreadID    string `json:"thread_id"`
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Allowed     bool   `json:"allowed"`
	Muted       bool   `json:"muted"`
	Reason      string `json:"reason,omitempty"`
}

type UserProfileOverview struct {
	UserID                string         `json:"user_id"`
	Nickname              string         `json:"nickname"`
	Phone                 string         `json:"phone"`
	AvatarInitial         string         `json:"avatar_initial"`
	MembershipLevel       string         `json:"membership_level"`
	MembershipTitle       string         `json:"membership_title"`
	CreditText            string         `json:"credit_text"`
	Verified              bool           `json:"verified"`
	GrowthValue           int            `json:"growth_value"`
	NextLevelGrowth       int            `json:"next_level_growth"`
	SavingsFen            int64          `json:"savings_fen"`
	WalletBalanceFen      int64          `json:"wallet_balance_fen"`
	PendingReceivableFen  int64          `json:"pending_receivable_fen"`
	CouponCount           int            `json:"coupon_count"`
	RedPacketCount        int            `json:"red_packet_count"`
	Points                int            `json:"points"`
	FavoriteShopCount     int            `json:"favorite_shop_count"`
	PaymentPasswordStatus string         `json:"payment_password_status"`
	OrderStats            map[string]int `json:"order_stats"`
}

type WalletOverview struct {
	Account               WalletAccount           `json:"account"`
	BalanceFen            int64                   `json:"balance_fen"`
	PendingReceivableFen  int64                   `json:"pending_receivable_fen"`
	CouponCount           int                     `json:"coupon_count"`
	RedPacketCount        int                     `json:"red_packet_count"`
	Points                int                     `json:"points"`
	PaymentPasswordStatus string                  `json:"payment_password_status"`
	Transactions          []WalletTransaction     `json:"transactions"`
	Withdrawals           []WalletWithdrawRequest `json:"withdrawals"`
}

type WalletWithdrawRequest struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	AmountFen   int64     `json:"amount_fen"`
	Channel     string    `json:"channel"`
	AccountName string    `json:"account_name,omitempty"`
	AccountNo   string    `json:"account_no,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserCoupon struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Kind         string    `json:"kind"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	Scope        string    `json:"scope"`
	Source       string    `json:"source"`
	Status       string    `json:"status"`
	ButtonText   string    `json:"button_text"`
	AccentColor  string    `json:"accent_color"`
	AmountFen    int64     `json:"amount_fen"`
	ThresholdFen int64     `json:"threshold_fen"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

type UserCouponsSummary struct {
	AvailableCount int          `json:"available_count"`
	RedPacketCount int          `json:"red_packet_count"`
	ExpiringCount  int          `json:"expiring_count"`
	Coupons        []UserCoupon `json:"coupons"`
}

type PointsBenefit struct {
	Icon     string `json:"icon"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Unlocked bool   `json:"unlocked"`
}

type PointsTask struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Reward     int    `json:"reward"`
	ActionText string `json:"action_text"`
	Route      string `json:"route,omitempty"`
	Done       bool   `json:"done"`
}

type PointsReward struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Points       int    `json:"points"`
	AmountFen    int64  `json:"amount_fen"`
	AccentColor  string `json:"accent_color"`
	RedeemText   string `json:"redeem_text"`
	ThresholdFen int64  `json:"threshold_fen"`
}

type PointsSummary struct {
	UserID          string              `json:"user_id"`
	Nickname        string              `json:"nickname"`
	MembershipLevel string              `json:"membership_level"`
	LevelName       string              `json:"level_name"`
	Verified        bool                `json:"verified"`
	Points          int                 `json:"points"`
	GrowthValue     int                 `json:"growth_value"`
	NextLevelGrowth int                 `json:"next_level_growth"`
	Benefits        []PointsBenefit     `json:"benefits"`
	Tasks           []PointsTask        `json:"tasks"`
	Rewards         []PointsReward      `json:"rewards"`
	Transactions    []PointsTransaction `json:"transactions"`
}

type InviteRecord struct {
	ID         string    `json:"id"`
	FriendName string    `json:"friend_name"`
	Status     string    `json:"status"`
	RewardText string    `json:"reward_text"`
	CreatedAt  time.Time `json:"created_at"`
}

type InviteSummary struct {
	UserID       string         `json:"user_id"`
	InviteCode   string         `json:"invite_code"`
	RewardText   string         `json:"reward_text"`
	ShareTitle   string         `json:"share_title"`
	SharePath    string         `json:"share_path"`
	Records      []InviteRecord `json:"records"`
	AbuseRiskTip string         `json:"abuse_risk_tip"`
}

type SearchResult struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	PriceFen     int64  `json:"price_fen,omitempty"`
	DistanceText string `json:"distance_text,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	Badge        string `json:"badge,omitempty"`
	ButtonText   string `json:"button_text"`
	Route        string `json:"route"`
}

type SearchCatalog struct {
	Keyword     string         `json:"keyword"`
	Category    string         `json:"category"`
	Total       int            `json:"total"`
	Suggestions []string       `json:"suggestions"`
	Results     []SearchResult `json:"results"`
}

type MedicineClinic struct {
	Name         string   `json:"name"`
	Location     string   `json:"location"`
	CoverURL     string   `json:"cover_url,omitempty"`
	BusinessTime string   `json:"business_time"`
	Tags         []string `json:"tags"`
	DeliveryText string   `json:"delivery_text"`
}

type MedicineProduct struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Subtitle             string `json:"subtitle"`
	Category             string `json:"category"`
	ImageURL             string `json:"image_url,omitempty"`
	PriceFen             int64  `json:"price_fen"`
	StockCount           int    `json:"stock_count"`
	RequiresPrescription bool   `json:"requires_prescription"`
	SelectedQuantity     int    `json:"selected_quantity"`
}

type MedicineHome struct {
	Clinic        MedicineClinic    `json:"clinic"`
	Categories    []string          `json:"categories"`
	Products      []MedicineProduct `json:"products"`
	CartCount     int               `json:"cart_count"`
	CartAmountFen int64             `json:"cart_amount_fen"`
}

type PrescriptionReviewStep struct {
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle"`
	Status   string    `json:"status"`
	At       time.Time `json:"at,omitempty"`
}

type PrescriptionOCRResult struct {
	Status             string   `json:"status"`
	Provider           string   `json:"provider"`
	Confidence         int      `json:"confidence"`
	MatchedProductID   string   `json:"matched_product_id,omitempty"`
	MatchedProductName string   `json:"matched_product_name,omitempty"`
	DosageText         string   `json:"dosage_text,omitempty"`
	RawText            string   `json:"raw_text,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

type PrescriptionArchiveRecord struct {
	ArchiveID     string    `json:"archive_id"`
	ObjectKey     string    `json:"object_key,omitempty"`
	ContentSHA    string    `json:"content_sha,omitempty"`
	RetainUntil   time.Time `json:"retain_until"`
	ArchivedAt    time.Time `json:"archived_at"`
	RetentionText string    `json:"retention_text"`
}

type PrescriptionReview struct {
	ID                  string                     `json:"id"`
	UserID              string                     `json:"user_id"`
	PatientName         string                     `json:"patient_name"`
	PatientPhone        string                     `json:"patient_phone,omitempty"`
	Address             string                     `json:"address,omitempty"`
	Hospital            string                     `json:"hospital,omitempty"`
	ProductID           string                     `json:"product_id"`
	ProductName         string                     `json:"product_name"`
	ProductImage        string                     `json:"product_image,omitempty"`
	PriceFen            int64                      `json:"price_fen"`
	Quantity            int                        `json:"quantity"`
	ImageURL            string                     `json:"image_url,omitempty"`
	ImageUploadTicketID string                     `json:"image_upload_ticket_id,omitempty"`
	ImageObjectKey      string                     `json:"image_object_key,omitempty"`
	ImageContentSHA     string                     `json:"image_content_sha,omitempty"`
	Note                string                     `json:"note,omitempty"`
	Status              string                     `json:"status"`
	DoctorName          string                     `json:"doctor_name,omitempty"`
	ReviewText          string                     `json:"review_text,omitempty"`
	OCRResult           *PrescriptionOCRResult     `json:"ocr_result,omitempty"`
	Archive             *PrescriptionArchiveRecord `json:"archive,omitempty"`
	Steps               []PrescriptionReviewStep   `json:"steps"`
	CreatedAt           time.Time                  `json:"created_at"`
	ReviewedAt          time.Time                  `json:"reviewed_at,omitempty"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

type CreatePrescriptionReviewRequest struct {
	UserID                    string `json:"user_id,omitempty"`
	PatientName               string `json:"patient_name"`
	PatientPhone              string `json:"patient_phone,omitempty"`
	Address                   string `json:"address,omitempty"`
	Hospital                  string `json:"hospital,omitempty"`
	ProductID                 string `json:"product_id"`
	ProductName               string `json:"product_name,omitempty"`
	ProductImage              string `json:"product_image,omitempty"`
	PriceFen                  int64  `json:"price_fen,omitempty"`
	Quantity                  int    `json:"quantity,omitempty"`
	ImageURL                  string `json:"image_url,omitempty"`
	PrescriptionImageTicketID string `json:"prescription_image_ticket_id,omitempty"`
	PrescriptionObjectKey     string `json:"prescription_object_key,omitempty"`
	PrescriptionContentSHA    string `json:"prescription_content_sha,omitempty"`
	Note                      string `json:"note,omitempty"`
}

type PrescriptionReviewListRequest struct {
	Status    string `json:"status,omitempty"`
	ProductID string `json:"product_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type ReviewPrescriptionRequest struct {
	ReviewID     string `json:"review_id,omitempty"`
	ReviewerID   string `json:"reviewer_id,omitempty"`
	ReviewerName string `json:"reviewer_name,omitempty"`
	Decision     string `json:"decision"`
	ReviewText   string `json:"review_text,omitempty"`
}

type PrescriptionImageUploadTicket struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ProductID     string    `json:"product_id,omitempty"`
	Provider      string    `json:"provider"`
	Bucket        string    `json:"bucket"`
	ObjectKey     string    `json:"object_key"`
	PublicURL     string    `json:"public_url"`
	FileName      string    `json:"file_name"`
	ContentType   string    `json:"content_type"`
	SizeBytes     int64     `json:"size_bytes"`
	MaxSizeBytes  int64     `json:"max_size_bytes"`
	ContentSHA    string    `json:"content_sha,omitempty"`
	Status        string    `json:"status"`
	ScanStatus    string    `json:"scan_status"`
	ScanResult    string    `json:"scan_result,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	UploadedAt    time.Time `json:"uploaded_at,omitempty"`
	ConfirmedAt   time.Time `json:"confirmed_at,omitempty"`
	ScanCheckedAt time.Time `json:"scan_checked_at,omitempty"`
}

type CreatePrescriptionImageUploadRequest struct {
	UserID      string `json:"user_id,omitempty"`
	ProductID   string `json:"product_id,omitempty"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type ConfirmPrescriptionImageUploadRequest struct {
	UserID      string `json:"user_id,omitempty"`
	TicketID    string `json:"ticket_id"`
	ObjectKey   string `json:"object_key"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentSHA  string `json:"content_sha,omitempty"`
}

type MedicineOrderItemRequest struct {
	ProductID            string `json:"product_id"`
	Name                 string `json:"name"`
	Category             string `json:"category,omitempty"`
	ImageURL             string `json:"image_url,omitempty"`
	PriceFen             int64  `json:"price_fen"`
	Quantity             int    `json:"quantity"`
	RequiresPrescription bool   `json:"requires_prescription,omitempty"`
}

type MedicineOrderRequest struct {
	UserID          string                     `json:"user_id,omitempty"`
	Address         string                     `json:"address,omitempty"`
	ContactName     string                     `json:"contact_name,omitempty"`
	ContactPhone    string                     `json:"contact_phone,omitempty"`
	ClinicName      string                     `json:"clinic_name,omitempty"`
	PrescriptionID  string                     `json:"prescription_id,omitempty"`
	Items           []MedicineOrderItemRequest `json:"items,omitempty"`
	DeliveryFeeFen  int64                      `json:"delivery_fee_fen,omitempty"`
	CouponAmountFen int64                      `json:"coupon_amount_fen,omitempty"`
	PaymentMethod   string                     `json:"payment_method,omitempty"`
	Remark          string                     `json:"remark,omitempty"`
}

type MedicineOrderItem struct {
	ProductID            string `json:"product_id"`
	Name                 string `json:"name"`
	Category             string `json:"category,omitempty"`
	ImageURL             string `json:"image_url,omitempty"`
	PriceFen             int64  `json:"price_fen"`
	Quantity             int    `json:"quantity"`
	RequiresPrescription bool   `json:"requires_prescription,omitempty"`
	PrescriptionApproved bool   `json:"prescription_approved,omitempty"`
	StockLocked          bool   `json:"stock_locked,omitempty"`
	StockRemaining       int    `json:"stock_remaining"`
}

type MedicineFeeRow struct {
	Title     string `json:"title"`
	AmountFen int64  `json:"amount_fen"`
}

type MedicineTimelineItem struct {
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle,omitempty"`
	Status   string    `json:"status"`
	Time     string    `json:"time,omitempty"`
	At       time.Time `json:"at,omitempty"`
}

type MedicineOrderDetail struct {
	Order              Order                  `json:"order"`
	Address            string                 `json:"address"`
	ContactName        string                 `json:"contact_name"`
	ContactPhone       string                 `json:"contact_phone"`
	ClinicName         string                 `json:"clinic_name"`
	ClinicLocation     string                 `json:"clinic_location"`
	DeliveryText       string                 `json:"delivery_text"`
	PrescriptionID     string                 `json:"prescription_id,omitempty"`
	PrescriptionStatus string                 `json:"prescription_status,omitempty"`
	DoctorName         string                 `json:"doctor_name,omitempty"`
	Advice             string                 `json:"advice"`
	Items              []MedicineOrderItem    `json:"items"`
	FeeRows            []MedicineFeeRow       `json:"fee_rows"`
	Timeline           []MedicineTimelineItem `json:"timeline"`
}

type ErrandOrderRequest struct {
	UserID          string `json:"user_id"`
	Type            string `json:"type"`
	PickupAddress   string `json:"pickup_address"`
	DeliveryAddress string `json:"delivery_address"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone"`
	ItemType        string `json:"item_type"`
	Description     string `json:"description"`
	ImageURL        string `json:"image_url,omitempty"`
	WeightText      string `json:"weight_text"`
	PickupTime      string `json:"pickup_time"`
	AmountFen       int64  `json:"amount_fen"`
	CouponAmountFen int64  `json:"coupon_amount_fen"`
}

type ErrandFeeRow struct {
	Title     string `json:"title"`
	AmountFen int64  `json:"amount_fen"`
}

type ErrandTimelineItem struct {
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle,omitempty"`
	Status   string    `json:"status"`
	Time     string    `json:"time,omitempty"`
	At       time.Time `json:"at,omitempty"`
}

type ErrandRider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RatingText   string `json:"rating_text"`
	Vehicle      string `json:"vehicle"`
	DistanceText string `json:"distance_text"`
}

type ErrandOrderDetail struct {
	Order           Order                `json:"order"`
	ServiceType     string               `json:"service_type"`
	ServiceTitle    string               `json:"service_title"`
	PickupAddress   string               `json:"pickup_address"`
	DeliveryAddress string               `json:"delivery_address"`
	ContactName     string               `json:"contact_name"`
	ContactPhone    string               `json:"contact_phone"`
	ItemType        string               `json:"item_type"`
	Description     string               `json:"description"`
	ImageURL        string               `json:"image_url,omitempty"`
	WeightText      string               `json:"weight_text"`
	PickupTime      string               `json:"pickup_time"`
	EstimateText    string               `json:"estimate_text"`
	MapStatus       string               `json:"map_status"`
	Rider           ErrandRider          `json:"rider"`
	FeeRows         []ErrandFeeRow       `json:"fee_rows"`
	Timeline        []ErrandTimelineItem `json:"timeline"`
}

type CouponPolicy struct {
	ID                                 string   `json:"id"`
	IssuerType                         string   `json:"issuer_type"`
	CostBearer                         string   `json:"cost_bearer"`
	SubsidySettlementRequired          bool     `json:"subsidy_settlement_required"`
	MerchantActivityAcceptanceRequired bool     `json:"merchant_activity_acceptance_required"`
	MerchantAcceptanceStatus           string   `json:"merchant_acceptance_status"`
	ScopeType                          string   `json:"scope_type"`
	ShopID                             string   `json:"shop_id"`
	ParticipatingShopIDs               []string `json:"participating_shop_ids"`
	AmountFen                          int64    `json:"amount_fen"`
}

type Order struct {
	ID              string       `json:"id"`
	UserID          string       `json:"user_id"`
	ShopID          string       `json:"shop_id,omitempty"`
	ShopName        string       `json:"shop_name,omitempty"`
	Reviewed        bool         `json:"reviewed,omitempty"`
	AddressID       string       `json:"address_id,omitempty"`
	AddressSnapshot OrderAddress `json:"address_snapshot,omitempty"`
	Type            string       `json:"type"`
	Status          string       `json:"status"`
	AmountFen       int64        `json:"amount_fen"`
	ItemsTotalFen   int64        `json:"items_total_fen"`
	DeliveryFeeFen  int64        `json:"delivery_fee_fen"`
	PackagingFeeFen int64        `json:"packaging_fee_fen"`
	DiscountFen     int64        `json:"discount_fen"`
	PaymentMethod   string       `json:"payment_method"`
	RiderID         string       `json:"rider_id,omitempty"`
	Items           []OrderItem  `json:"items"`
	Options         OrderOptions `json:"options"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Events          []OrderEvent `json:"events"`
}

type OrderAddress struct {
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	City         string `json:"city"`
	Detail       string `json:"detail"`
	Tag          string `json:"tag,omitempty"`
}

type OrderItem struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	ImageURL     string `json:"image_url,omitempty"`
	UnitPriceFen int64  `json:"unit_price_fen"`
	Quantity     int    `json:"quantity"`
}

type OrderEvent struct {
	Type      string    `json:"type"`
	ActorID   string    `json:"actor_id,omitempty"`
	Message   string    `json:"message"`
	AmountFen int64     `json:"amount_fen,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletAccount struct {
	UserID    string `json:"user_id"`
	Balance   int64  `json:"balance_fen"`
	Frozen    int64  `json:"frozen_fen"`
	Version   int64  `json:"version"`
	RiskState string `json:"risk_state"`
}

type WalletTransaction struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	OrderID        string    `json:"order_id,omitempty"`
	Type           string    `json:"type"`
	AmountFen      int64     `json:"amount_fen"`
	PaymentMethod  string    `json:"payment_method"`
	IdempotencyKey string    `json:"idempotency_key"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaymentTransaction struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	Method         string    `json:"method"`
	AmountFen      int64     `json:"amount_fen"`
	Status         string    `json:"status"`
	OutTradeNo     string    `json:"out_trade_no"`
	TransactionID  string    `json:"transaction_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	AmountFen int64  `json:"amount_fen"`
}

type CreditWalletRequest struct {
	UserID         string `json:"user_id"`
	AmountFen      int64  `json:"amount_fen"`
	IdempotencyKey string `json:"idempotency_key"`
}

type BalancePayRequest struct {
	OrderID         string `json:"order_id"`
	UserID          string `json:"user_id"`
	PaymentPassword string `json:"payment_password"`
	IdempotencyKey  string `json:"idempotency_key"`
}

type UpsertCartItemRequest struct {
	UserID    string `json:"user_id"`
	ShopID    string `json:"shop_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Selected  *bool  `json:"selected,omitempty"`
}

type CheckoutCartRequest struct {
	UserID    string       `json:"user_id"`
	ShopID    string       `json:"shop_id"`
	AddressID string       `json:"address_id"`
	Options   OrderOptions `json:"options"`
}

type SetWalletPaymentPasswordRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

type WechatPrepayRequest struct {
	UserID   string `json:"user_id"`
	OrderID  string `json:"order_id"`
	ClientIP string `json:"client_ip"`
}

type WechatPrepayResponse struct {
	AppID      string `json:"app_id"`
	PrepayID   string `json:"prepay_id"`
	OutTradeNo string `json:"out_trade_no"`
	TimeStamp  string `json:"time_stamp"`
	NonceStr   string `json:"nonce_str"`
	Package    string `json:"package"`
	SignType   string `json:"sign_type"`
	PaySign    string `json:"pay_sign"`
	AmountFen  int64  `json:"amount_fen"`
}

type WechatPaymentCallbackRequest struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	AmountFen     int64  `json:"amount_fen"`
}

type AuditLog struct {
	ID                 string         `json:"id"`
	ActorType          string         `json:"actor_type"`
	ActorID            string         `json:"actor_id"`
	Action             string         `json:"action"`
	TargetType         string         `json:"target_type"`
	TargetID           string         `json:"target_id"`
	RequestID          string         `json:"request_id"`
	IPHash             string         `json:"ip_hash"`
	Payload            map[string]any `json:"payload"`
	IntegrityAlgorithm string         `json:"integrity_algorithm"`
	IntegrityHash      string         `json:"integrity_hash"`
	IntegrityVerified  bool           `json:"integrity_verified"`
	CreatedAt          time.Time      `json:"created_at"`
}

type RecordAuditLogRequest struct {
	ActorType  string         `json:"actor_type"`
	ActorID    string         `json:"actor_id"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	RequestID  string         `json:"request_id"`
	IPHash     string         `json:"ip_hash"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

type AuditLogsRequest struct {
	ActorType  string    `json:"actor_type"`
	ActorID    string    `json:"actor_id"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Limit      int       `json:"limit"`
	After      time.Time `json:"after"`
	Before     time.Time `json:"before"`
}

type AuditRetentionReportRequest struct {
	RetentionDays        int       `json:"retention_days"`
	HotDays              int       `json:"hot_days"`
	IntegritySampleLimit int       `json:"integrity_sample_limit"`
	CriticalActions      []string  `json:"critical_actions"`
	Now                  time.Time `json:"now"`
}

type AuditActionCoverage struct {
	Action        string    `json:"action"`
	Count         int       `json:"count"`
	LastCreatedAt time.Time `json:"last_created_at"`
}

type AuditRetentionAlert struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Count    int    `json:"count"`
}

type AuditRetentionReport struct {
	Status                 string                `json:"status"`
	GeneratedAt            time.Time             `json:"generated_at"`
	RetentionDays          int                   `json:"retention_days"`
	HotDays                int                   `json:"hot_days"`
	RetentionCutoff        time.Time             `json:"retention_cutoff"`
	ColdArchiveCutoff      time.Time             `json:"cold_archive_cutoff"`
	TotalLogs              int                   `json:"total_logs"`
	OldestCreatedAt        time.Time             `json:"oldest_created_at"`
	NewestCreatedAt        time.Time             `json:"newest_created_at"`
	ExpiredLogs            int                   `json:"expired_logs"`
	ColdArchiveDueLogs     int                   `json:"cold_archive_due_logs"`
	IntegritySampleSize    int                   `json:"integrity_sample_size"`
	IntegrityFailures      int                   `json:"integrity_failures"`
	ExportEvents           int                   `json:"export_events"`
	CriticalActionCoverage []AuditActionCoverage `json:"critical_action_coverage"`
	MissingCriticalActions []string              `json:"missing_critical_actions"`
	Alerts                 []AuditRetentionAlert `json:"alerts"`
}

type AuditRetentionAlertEmissionRequest struct {
	RetentionDays        int       `json:"retention_days"`
	HotDays              int       `json:"hot_days"`
	IntegritySampleLimit int       `json:"integrity_sample_limit"`
	Now                  time.Time `json:"now"`
}

type AuditRetentionAlertEmission struct {
	Status         string                `json:"status"`
	ReportStatus   string                `json:"report_status"`
	AlertCount     int                   `json:"alert_count"`
	CriticalCount  int                   `json:"critical_count"`
	WarningCount   int                   `json:"warning_count"`
	Topic          string                `json:"topic"`
	OutboxEventID  string                `json:"outbox_event_id"`
	IdempotencyKey string                `json:"idempotency_key"`
	EmittedAt      time.Time             `json:"emitted_at"`
	Alerts         []AuditRetentionAlert `json:"alerts"`
	Report         *AuditRetentionReport `json:"report"`
}

type AuditArchiveRequest struct {
	HotDays       int       `json:"hot_days"`
	Limit         int       `json:"limit"`
	StoragePrefix string    `json:"storage_prefix"`
	Now           time.Time `json:"now"`
}

type AuditArchiveManifestEntry struct {
	ID                 string    `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	Action             string    `json:"action"`
	TargetType         string    `json:"target_type"`
	TargetID           string    `json:"target_id"`
	IntegrityAlgorithm string    `json:"integrity_algorithm"`
	IntegrityHash      string    `json:"integrity_hash"`
	IntegrityVerified  bool      `json:"integrity_verified"`
}

type AuditArchiveRequestResult struct {
	ArchiveID         string                      `json:"archive_id"`
	Status            string                      `json:"status"`
	Topic             string                      `json:"topic"`
	StoragePrefix     string                      `json:"storage_prefix"`
	StorageKey        string                      `json:"storage_key"`
	HotDays           int                         `json:"hot_days"`
	ColdArchiveCutoff time.Time                   `json:"cold_archive_cutoff"`
	LogCount          int                         `json:"log_count"`
	IntegrityFailures int                         `json:"integrity_failures"`
	OldestCreatedAt   time.Time                   `json:"oldest_created_at"`
	NewestCreatedAt   time.Time                   `json:"newest_created_at"`
	ManifestAlgorithm string                      `json:"manifest_algorithm"`
	ManifestHash      string                      `json:"manifest_hash"`
	ManifestEntries   []AuditArchiveManifestEntry `json:"manifest_entries"`
	OutboxEventID     string                      `json:"outbox_event_id"`
	IdempotencyKey    string                      `json:"idempotency_key"`
	RequestedAt       time.Time                   `json:"requested_at"`
}

type AuditArchiveCompletionRequest struct {
	ArchiveID         string    `json:"archive_id"`
	StorageKey        string    `json:"storage_key"`
	ManifestAlgorithm string    `json:"manifest_algorithm"`
	ManifestHash      string    `json:"manifest_hash"`
	ContentHash       string    `json:"content_hash"`
	Bytes             int64     `json:"bytes"`
	ObjectLockMode    string    `json:"object_lock_mode"`
	RetainUntil       time.Time `json:"retain_until"`
	OutboxEventID     string    `json:"outbox_event_id"`
	UploadedAt        time.Time `json:"uploaded_at"`
}

type AuditArchiveCompletion struct {
	ArchiveID         string    `json:"archive_id"`
	Status            string    `json:"status"`
	StorageKey        string    `json:"storage_key"`
	ManifestAlgorithm string    `json:"manifest_algorithm"`
	ManifestHash      string    `json:"manifest_hash"`
	ContentHash       string    `json:"content_hash"`
	Bytes             int64     `json:"bytes"`
	ObjectLockMode    string    `json:"object_lock_mode"`
	RetainUntil       time.Time `json:"retain_until,omitempty"`
	OutboxEventID     string    `json:"outbox_event_id"`
	UploadedAt        time.Time `json:"uploaded_at"`
	CompletedAt       time.Time `json:"completed_at"`
}

type AuditArchiveVerifyRequest struct {
	ArchiveID string    `json:"archive_id"`
	Now       time.Time `json:"now"`
}

type AuditArchiveVerification struct {
	ArchiveID           string    `json:"archive_id"`
	Status              string    `json:"status"`
	StorageKey          string    `json:"storage_key"`
	ManifestAlgorithm   string    `json:"manifest_algorithm"`
	ManifestHash        string    `json:"manifest_hash"`
	ExpectedContentHash string    `json:"expected_content_hash"`
	ActualContentHash   string    `json:"actual_content_hash"`
	ExpectedBytes       int64     `json:"expected_bytes"`
	ActualBytes         int64     `json:"actual_bytes"`
	ArchiveIDMatched    bool      `json:"archive_id_matched"`
	ManifestHashMatched bool      `json:"manifest_hash_matched"`
	ContentHashMatched  bool      `json:"content_hash_matched"`
	BytesMatched        bool      `json:"bytes_matched"`
	LogCountMatched     bool      `json:"log_count_matched"`
	HeaderLogCount      int       `json:"header_log_count"`
	ManifestEntryCount  int       `json:"manifest_entry_count"`
	ErrorCode           string    `json:"error_code,omitempty"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	VerifiedAt          time.Time `json:"verified_at"`
}

type AuditArchiveListRequest struct {
	ArchiveID string    `json:"archive_id"`
	Limit     int       `json:"limit"`
	After     time.Time `json:"after"`
	Before    time.Time `json:"before"`
}

type AuditArchiveVerificationListRequest struct {
	ArchiveID string    `json:"archive_id"`
	Status    string    `json:"status"`
	Limit     int       `json:"limit"`
	After     time.Time `json:"after"`
	Before    time.Time `json:"before"`
}

func DefaultHomeModules() []HomeModule {
	return []HomeModule{
		{Key: "takeout", Title: "外卖", Route: "/pages/shop/list/index", Icon: "takeout", IconURL: "/assets/generated/category-takeout.png", Enabled: true, SortOrder: 10, Scene: "home"},
		{Key: "groupbuy", Title: "团购", Route: "/pages/shop/list/index?tab=groupbuy", Icon: "groupbuy", IconURL: "/assets/generated/category-groupbuy.png", Enabled: true, SortOrder: 20, Scene: "home"},
		{Key: "medicine", Title: "买药", Route: "/pages/medicine/home/index", Icon: "medicine", IconURL: "/assets/generated/category-medicine.png", Enabled: true, SortOrder: 30, Scene: "home"},
		{Key: "courier", Title: "快递跑腿", Route: "/pages/errand/home/index", Icon: "courier", IconURL: "/assets/generated/category-courier.png", Enabled: true, SortOrder: 40, Scene: "home"},
		{Key: HomeModuleCircle, Title: "圈子", Route: "/pages/circle/index", Icon: "circle", IconURL: "/assets/generated/category-circle.png", Enabled: true, SortOrder: 50, Scene: "home"},
		{Key: HomeModuleCharity, Title: "公益", Route: "/pages/charity/index", Icon: "charity", Enabled: false, SortOrder: 60, Scene: "home"},
		{Key: "social", Title: "交友", Route: "/pages/social/index", Icon: "social", Enabled: false, SortOrder: 70, Scene: "home"},
		{Key: "meal-match", Title: "找饭搭", Route: "/pages/meal-match/index", Icon: "meal-match", IconURL: "/assets/generated/category-meal-match.png", Enabled: true, SortOrder: 80, Scene: "home"},
		{Key: "coupons", Title: "红包优惠", Route: "/pages/coupons/index", Icon: "coupons", IconURL: "/assets/generated/category-coupons.png", Enabled: true, SortOrder: 90, Scene: "home"},
		{Key: "points", Title: "会员积分", Route: "/pages/member-points/index", Icon: "points", IconURL: "/assets/generated/category-points.png", Enabled: true, SortOrder: 100, Scene: "home"},
	}
}

func DefaultHomeCards() []HomeCard {
	return []HomeCard{
		{ID: "card_takeout_sample", Type: HomeCardProduct, Title: "后台推荐商品位", Subtitle: "商品、店铺、团购和圈子内容都由后台控制", TargetID: "product_placeholder", ImageURL: "/assets/generated/home-recommend-restaurant.jpg", Enabled: true, SortOrder: 10},
		{ID: "card_circle_sample", Type: HomeCardCircle, Title: "圈子小微墙", Subtitle: "轻量动态和饭搭入口", TargetID: "circle_micro_wall", ImageURL: "/assets/generated/home-meal-match.jpg", Enabled: true, SortOrder: 20},
	}
}

func IsOrderType(value string) bool {
	switch value {
	case OrderTypeTakeout, OrderTypeGroupbuy, OrderTypeMedicine, OrderTypeCourier, OrderTypeErrandBuy, OrderTypeErrandDeliver, OrderTypeErrandPickup, OrderTypeErrandDo:
		return true
	default:
		return false
	}
}

func IsOutboxStatus(value string) bool {
	switch value {
	case OutboxStatusPending, OutboxStatusPublished, OutboxStatusFailed, OutboxStatusDeadLetter:
		return true
	default:
		return false
	}
}

func FulfillmentModeForOrderType(value string) string {
	switch value {
	case OrderTypeGroupbuy:
		return FulfillmentInStoreRedeem
	case OrderTypeCourier, OrderTypeErrandBuy, OrderTypeErrandDeliver, OrderTypeErrandPickup, OrderTypeErrandDo:
		return FulfillmentPlatformErrand
	default:
		return FulfillmentRiderDelivery
	}
}

func DispatchModeForOrderAgeSeconds(ageSeconds int) string {
	if ageSeconds >= DispatchGrabHallSeconds {
		return DispatchModeAutoAssign
	}
	return DispatchModeGrabHall
}

func RiderCanAcceptOrders(account RiderAccount, deposit DepositAccount) bool {
	if account.Type != RiderAccountRider || account.Status != "active" || !account.Online {
		return false
	}
	if deposit.SubjectType != "rider" {
		return false
	}
	if deposit.Status == DepositStatusWechatExemptApproved {
		return true
	}
	return deposit.Status == DepositStatusPaid && deposit.AmountFen >= RiderDepositAmountFen
}

func MerchantCanAcceptOrders(deposit DepositAccount) bool {
	return deposit.SubjectType == "merchant" && deposit.Status == DepositStatusPaid && deposit.AmountFen >= MerchantDepositAmountFen
}

func RiderDispatchPriority(level string) int {
	switch level {
	case RiderLevelS:
		return 400
	case RiderLevelA:
		return 300
	case RiderLevelB:
		return 200
	default:
		return 100
	}
}

func RiderCanDeclineDispatchWithoutPenalty(completedOrderCount int, fixedDailyOrderCount int) bool {
	return fixedDailyOrderCount > 0 && completedOrderCount >= fixedDailyOrderCount
}

func RefundDestinationForStrategy(defaultStrategy string, requestedDestination string) string {
	switch requestedDestination {
	case RefundDestinationBalance, RefundDestinationOriginalRoute:
		return requestedDestination
	}
	if defaultStrategy == RefundStrategyOriginalFirst {
		return RefundDestinationOriginalRoute
	}
	return RefundDestinationBalance
}

func NormalizeRefundStrategy(strategy string) string {
	if strings.TrimSpace(strategy) == RefundStrategyOriginalFirst {
		return RefundStrategyOriginalFirst
	}
	return RefundStrategyBalanceFirst
}

func GroupbuyUnavailableRefund(product MerchantProduct, voucherStatus string, defaultStrategy string) (bool, string) {
	unavailable := product.Status == ProductStatusSoldOut || product.Status == ProductStatusRemoved || product.StockCount <= 0
	if !unavailable || voucherStatus == GroupbuyVoucherRedeemed || voucherStatus == GroupbuyVoucherRefunded {
		return false, ""
	}
	return true, RefundDestinationForStrategy(defaultStrategy, "")
}

func CanUseMealMatch(profile MealMatchProfile) (bool, []string) {
	ok, missing := mealMatchProfilePrerequisites(profile)
	if !ok {
		return false, missing
	}
	if normalizeMealMatchDeviceRiskState(profile.DeviceRiskState) != MealMatchDeviceRiskPassed {
		return false, []string{"device_risk_review"}
	}
	status := normalizeMealMatchModerationStatus(profile.ModerationStatus)
	switch status {
	case MealMatchModerationApproved:
		return true, []string{}
	case MealMatchModerationRejected:
		return false, []string{"moderation_rejected"}
	default:
		return false, []string{"moderation_pending"}
	}
}

func mealMatchProfilePrerequisites(profile MealMatchProfile) (bool, []string) {
	missing := []string{}
	if strings.TrimSpace(profile.Gender) == "" {
		missing = append(missing, "gender")
	}
	if strings.TrimSpace(profile.SchoolID) == "" {
		missing = append(missing, "school")
	}
	if strings.TrimSpace(profile.DeviceID) == "" {
		missing = append(missing, "device_id")
	}
	if !profile.IdentityTruthSigned {
		missing = append(missing, "identity_truth_signed")
	}
	if !profile.PlatformLiabilityReleaseSigned {
		missing = append(missing, "platform_liability_release_signed")
	}
	if !profile.QuestionnaireCompleted || len(profile.PersonalityTraits) == 0 || len(profile.DietaryHabits) == 0 {
		missing = append(missing, "questionnaire_completed")
	}
	return len(missing) == 0, missing
}

func normalizeMealMatchModerationStatus(status string) string {
	switch strings.TrimSpace(status) {
	case MealMatchModerationApproved:
		return MealMatchModerationApproved
	case MealMatchModerationRejected:
		return MealMatchModerationRejected
	case MealMatchModerationPending:
		return MealMatchModerationPending
	default:
		return ""
	}
}

func normalizeMealMatchPrivacyScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case MealMatchPrivacySameBuilding:
		return MealMatchPrivacySameBuilding
	default:
		return MealMatchPrivacySameSchool
	}
}

func normalizeMealMatchLocationPrecision(precision string) string {
	switch strings.TrimSpace(precision) {
	case MealMatchLocationBuildingOnly:
		return MealMatchLocationBuildingOnly
	default:
		return MealMatchLocationCampusOnly
	}
}

func normalizeMealMatchDeviceRiskState(state string) string {
	switch strings.TrimSpace(state) {
	case MealMatchDeviceRiskPassed:
		return MealMatchDeviceRiskPassed
	case MealMatchDeviceRiskReview:
		return MealMatchDeviceRiskReview
	case MealMatchDeviceRiskBlocked:
		return MealMatchDeviceRiskBlocked
	default:
		return MealMatchDeviceRiskReview
	}
}

func UserAddressReady(address UserAddress) bool {
	return strings.TrimSpace(address.ContactName) != "" &&
		strings.TrimSpace(address.ContactPhone) != "" &&
		strings.TrimSpace(address.City) != "" &&
		strings.TrimSpace(address.Detail) != "" &&
		address.Latitude != nil &&
		address.Longitude != nil
}

func OrderPayableFen(itemsTotalFen int64, deliveryFeeFen int64, packagingFeeFen int64, discountFen int64) int64 {
	total := maxInt64(0, itemsTotalFen) + maxInt64(0, deliveryFeeFen) + maxInt64(0, packagingFeeFen) - maxInt64(0, discountFen)
	return maxInt64(0, total)
}

func CanCreateAfterSales(order Order, request AfterSalesRequest) bool {
	if strings.TrimSpace(request.OrderID) == "" || strings.TrimSpace(request.Reason) == "" || request.RequestedAmountFen <= 0 {
		return false
	}
	if order.AmountFen > 0 && request.RequestedAmountFen > order.AmountFen {
		return false
	}
	switch order.Status {
	case StatusPendingPayment, StatusCancelled, StatusRefundPending, StatusRefunded:
		return false
	default:
		return true
	}
}

func DeliveryPromiseStatus(deadline time.Time, deliveredAt time.Time, timeoutExempt bool) string {
	if timeoutExempt {
		return DeliveryPromiseExempt
	}
	if deadline.IsZero() || deliveredAt.IsZero() || !deliveredAt.After(deadline) {
		return DeliveryPromiseOnTime
	}
	return DeliveryPromiseTimeout
}

func ResolveMembershipTier(growthValue int) string {
	switch {
	case growthValue >= 10000:
		return MembershipBlackGold
	case growthValue >= 3000:
		return MembershipGold
	case growthValue >= 500:
		return MembershipSilver
	default:
		return MembershipNone
	}
}

func RiskDecisionBlocked(events []RiskEvent, abnormalOrderLimit int, maliciousRefundLimit int) bool {
	if abnormalOrderLimit <= 0 {
		abnormalOrderLimit = 7
	}
	if maliciousRefundLimit <= 0 {
		maliciousRefundLimit = 3
	}
	abnormalOrders := 0
	maliciousRefunds := 0
	for _, event := range events {
		switch event.Type {
		case RiskEventFakeTransaction:
			return true
		case RiskEventAbnormalOrdering:
			abnormalOrders++
		case RiskEventMaliciousRefund:
			maliciousRefunds++
		}
	}
	return abnormalOrders >= abnormalOrderLimit || maliciousRefunds >= maliciousRefundLimit
}

func CouponPolicyFromInput(input CouponPolicy) CouponPolicy {
	output := input
	if output.IssuerType != CouponIssuerPlatform {
		output.IssuerType = CouponIssuerMerchant
	}
	if output.IssuerType == CouponIssuerPlatform && output.CostBearer == CouponCostBearerPlatform {
		output.CostBearer = CouponCostBearerPlatform
	} else {
		output.CostBearer = CouponCostBearerMerchant
	}
	if output.ScopeType != CouponScopeParticipatingShops {
		output.ScopeType = CouponScopeSingleShop
	}
	if output.MerchantAcceptanceStatus == "" {
		output.MerchantAcceptanceStatus = CouponActivityPending
	}
	output.SubsidySettlementRequired = output.IssuerType == CouponIssuerPlatform && output.CostBearer == CouponCostBearerPlatform
	output.MerchantActivityAcceptanceRequired = output.IssuerType == CouponIssuerPlatform && output.CostBearer == CouponCostBearerMerchant
	return output
}

func CouponCanApplyToShop(input CouponPolicy, shopID string) bool {
	policy := CouponPolicyFromInput(input)
	if policy.MerchantActivityAcceptanceRequired && policy.MerchantAcceptanceStatus != CouponActivityAccepted {
		return false
	}
	if policy.ScopeType == CouponScopeSingleShop {
		return policy.ShopID != "" && policy.ShopID == shopID
	}
	return stringSliceContains(policy.ParticipatingShopIDs, shopID)
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func maxInt64(left int64, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
