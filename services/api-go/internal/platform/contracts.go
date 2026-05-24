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

	HomeModuleCircle  = "circle"
	HomeModuleCharity = "charity"
	HomeCardProduct   = "product"
	HomeCardShop      = "shop"
	HomeCardCoupon    = "coupon"
	HomeCardCircle    = "circle_post"

	CircleFeatureDisabled           = "disabled"
	CircleFeatureWallOnly           = "wall_only"
	CircleFeatureCircleAndMealMatch = "circle_and_meal_match"
	CirclePostText                  = "text"
	CirclePostImage                 = "image"
	CirclePostFoodInvite            = "food_invite"
	CirclePostPendingReview         = "pending_review"
	CirclePostPublished             = "published"

	AddressTagHome              = "home"
	FavoriteTargetShop          = "shop"
	ReviewTargetShop            = "shop"
	ReviewTargetRider           = "rider"
	ReviewPublished             = "published"
	PointsTransactionEarn       = "earn"
	PointsTransactionRedeem     = "redeem"
	PointsTransactionRefund     = "refund_deduct"
	MembershipNone              = "none"
	MembershipSilver            = "silver"
	MembershipGold              = "gold"
	MembershipBlackGold         = "black_gold"
	NotificationWechatSubscribe = "wechat_subscribe"
	PushStatusQueued            = "queued"
	PushStatusAcked             = "acked"
	RiskEventAbnormalOrdering   = "abnormal_ordering"
	RiskEventMaliciousRefund    = "malicious_refund"
	RiskEventFakeTransaction    = "fake_transaction"
	DataManagementFullBundle    = "full_bundle"

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

type RiderPerformance struct {
	RiderID              string  `json:"rider_id"`
	StationID            string  `json:"station_id"`
	AverageAcceptSeconds float64 `json:"average_accept_seconds"`
	AverageDailyOrders   float64 `json:"average_daily_orders"`
	CompletionRate       float64 `json:"completion_rate"`
	Score                int     `json:"score"`
	Level                string  `json:"level"`
	DispatchPriority     int     `json:"dispatch_priority"`
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
	UnitPriceFen int64  `json:"unit_price_fen"`
	Quantity     int    `json:"quantity"`
	Selected     bool   `json:"selected"`
}

type CartSummary struct {
	UserID          string     `json:"user_id"`
	ShopID          string     `json:"shop_id"`
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
	ID             string `json:"id"`
	SenderID       string `json:"sender_id"`
	SenderRole     string `json:"sender_role"`
	Scene          string `json:"scene"`
	TargetID       string `json:"target_id"`
	Type           string `json:"type"`
	TotalAmountFen int64  `json:"total_amount_fen"`
	Quantity       int    `json:"quantity"`
	PaymentMethod  string `json:"payment_method"`
}

type AfterSalesRequest struct {
	ID                 string    `json:"id"`
	OrderID            string    `json:"order_id"`
	UserID             string    `json:"user_id"`
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

type Review struct {
	ID         string   `json:"id"`
	TargetType string   `json:"target_type"`
	TargetID   string   `json:"target_id"`
	UserID     string   `json:"user_id"`
	Rating     int      `json:"rating"`
	Content    string   `json:"content"`
	ImageURLs  []string `json:"image_urls"`
	Status     string   `json:"status"`
}

type Favorite struct {
	UserID     string `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
}

type PointsTransaction struct {
	UserID   string `json:"user_id"`
	Type     string `json:"type"`
	Points   int    `json:"points"`
	SourceID string `json:"source_id"`
}

type PushDelivery struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Channel     string `json:"channel"`
	TemplateKey string `json:"template_key"`
	Status      string `json:"status"`
	RetryCount  int    `json:"retry_count"`
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
	ID           string   `json:"id"`
	AuthorUserID string   `json:"author_user_id"`
	CircleID     string   `json:"circle_id"`
	Type         string   `json:"type"`
	Content      string   `json:"content"`
	ImageURLs    []string `json:"image_urls"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
}

type MealMatchProfile struct {
	UserID                         string   `json:"user_id"`
	Gender                         string   `json:"gender"`
	IdentityTruthSigned            bool     `json:"identity_truth_signed"`
	PlatformLiabilityReleaseSigned bool     `json:"platform_liability_release_signed"`
	QuestionnaireCompleted         bool     `json:"questionnaire_completed"`
	PersonalityTraits              []string `json:"personality_traits"`
	DietaryHabits                  []string `json:"dietary_habits"`
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
	AddressID       string       `json:"address_id,omitempty"`
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

type OrderItem struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
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
		{Key: "takeout", Title: "外卖", Route: "/pages/shop/list/index", Icon: "takeout", Enabled: true, SortOrder: 10, Scene: "home"},
		{Key: "groupbuy", Title: "团购", Route: "/pages/shop/list/index?tab=groupbuy", Icon: "groupbuy", Enabled: true, SortOrder: 20, Scene: "home"},
		{Key: "medicine", Title: "买药", Route: "/pages/shop/list/index?tab=medicine", Icon: "medicine", Enabled: true, SortOrder: 30, Scene: "home"},
		{Key: "courier", Title: "快递跑腿", Route: "/pages/shop/list/index?tab=courier", Icon: "courier", Enabled: true, SortOrder: 40, Scene: "home"},
		{Key: HomeModuleCircle, Title: "圈子", Route: "/pages/circle/index", Icon: "circle", Enabled: true, SortOrder: 50, Scene: "home"},
		{Key: HomeModuleCharity, Title: "公益", Route: "/pages/charity/index", Icon: "charity", Enabled: false, SortOrder: 60, Scene: "home"},
		{Key: "social", Title: "交友", Route: "/pages/social/index", Icon: "social", Enabled: false, SortOrder: 70, Scene: "home"},
	}
}

func DefaultHomeCards() []HomeCard {
	return []HomeCard{
		{ID: "card_takeout_sample", Type: HomeCardProduct, Title: "后台推荐商品位", Subtitle: "商品、店铺、团购和圈子内容都由后台控制", TargetID: "product_placeholder", Enabled: true, SortOrder: 10},
		{ID: "card_circle_sample", Type: HomeCardCircle, Title: "圈子小微墙", Subtitle: "轻量动态和饭搭入口", TargetID: "circle_micro_wall", Enabled: true, SortOrder: 20},
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
	missing := []string{}
	if strings.TrimSpace(profile.Gender) == "" {
		missing = append(missing, "gender")
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
