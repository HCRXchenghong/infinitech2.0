package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"infinitech2/services/api-go/internal/platform"

	"golang.org/x/crypto/bcrypt"
)

type Router struct {
	store                              platform.Repository
	mux                                *http.ServeMux
	authVerifier                       AuthVerifier
	tokenSigner                        TokenSigner
	authSessions                       AuthSessionStore
	adminPasswordHash                  map[string]string
	allowDevAuth                       bool
	wechatPay                          WechatPayVerifier
	wechatMini                         WechatMiniSessionResolver
	notificationProviderCallbackSecret string
	realtimeInternalToken              string
}

type RouterOption func(*Router)

func WithWechatMiniSessionResolver(resolver WechatMiniSessionResolver) RouterOption {
	return func(router *Router) {
		if resolver != nil {
			router.wechatMini = resolver
		}
	}
}

func WithAuthSessionStore(store AuthSessionStore) RouterOption {
	return func(router *Router) {
		if store != nil {
			router.authSessions = store
		}
	}
}

func WithAdminLoginCredential(accountID string, password string) RouterOption {
	return func(router *Router) {
		accountID = strings.TrimSpace(accountID)
		password = strings.TrimSpace(password)
		if accountID == "" || len(password) < 8 || len(password) > 72 {
			return
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return
		}
		if router.adminPasswordHash == nil {
			router.adminPasswordHash = map[string]string{}
		}
		router.adminPasswordHash[accountID] = string(hashed)
	}
}

func WithDevBearerAuth(enabled bool) RouterOption {
	return func(router *Router) {
		router.allowDevAuth = enabled
	}
}

func WithNotificationProviderCallbackSecret(secret string) RouterOption {
	return func(router *Router) {
		router.notificationProviderCallbackSecret = strings.TrimSpace(secret)
	}
}

func WithRealtimeInternalToken(token string) RouterOption {
	return func(router *Router) {
		router.realtimeInternalToken = strings.TrimSpace(token)
	}
}

func NewRouter(store platform.Repository, options ...RouterOption) http.Handler {
	tokenSigner := NewTokenSigner(os.Getenv("AUTH_TOKEN_SECRET"))
	authSessions := NewMemoryAuthSessionStore()
	router := &Router{
		store:                              store,
		mux:                                http.NewServeMux(),
		tokenSigner:                        tokenSigner,
		authSessions:                       authSessions,
		adminPasswordHash:                  map[string]string{},
		allowDevAuth:                       true,
		wechatPay:                          NewWechatPayVerifier(os.Getenv("WECHAT_PAY_CALLBACK_SECRET")),
		wechatMini:                         DevWechatMiniSessionResolver{},
		notificationProviderCallbackSecret: strings.TrimSpace(os.Getenv("NOTIFICATION_PROVIDER_CALLBACK_SECRET")),
		realtimeInternalToken:              strings.TrimSpace(os.Getenv("REALTIME_INTERNAL_TOKEN")),
	}
	for _, option := range options {
		option(router)
	}
	router.restoreAdminRBACAppliedPolicyFromAudit()
	router.rebuildAuthVerifier()
	router.routes()
	return router
}

func (r *Router) rebuildAuthVerifier() {
	verifiers := ChainedVerifier{NewSessionAuthVerifier(r.tokenSigner, r.authSessions)}
	if r.allowDevAuth {
		verifiers = append(verifiers, DevBearerVerifier{})
	}
	r.authVerifier = verifiers
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	r.mux.ServeHTTP(w, req)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /healthz", r.handleHealth)
	r.mux.HandleFunc("GET /readyz", r.handleReady)
	r.mux.HandleFunc("POST /internal/realtime/authorize", r.handleInternalRealtimeAuthorize)
	r.mux.HandleFunc("POST /api/auth/wechat-mini/login", r.handleWechatMiniLogin)
	r.mux.HandleFunc("POST /api/auth/phone/code", r.handleSendPhoneVerificationCode)
	r.mux.HandleFunc("POST /api/auth/phone/login", r.handlePhoneLogin)
	r.mux.HandleFunc("POST /api/auth/phone/register", r.handlePhoneRegister)
	r.mux.HandleFunc("POST /api/auth/logout", r.handleLogout)
	r.mux.HandleFunc("POST /api/admin/merchant-invites", r.handleCreateMerchantInvite)
	r.mux.HandleFunc("POST /api/admin/rider-invites", r.handleCreateRiderInvite)
	r.mux.HandleFunc("GET /api/admin/refund-settings", r.handleAdminRefundSettings)
	r.mux.HandleFunc("PUT /api/admin/refund-settings", r.handleAdminSaveRefundSettings)
	r.mux.HandleFunc("GET /api/admin/refunds", r.handleAdminRefundTransactions)
	r.mux.HandleFunc("GET /api/admin/orders/{orderID}", r.handleAdminOrderDetail)
	r.mux.HandleFunc("GET /api/admin/after-sales", r.handleAdminAfterSales)
	r.mux.HandleFunc("GET /api/admin/after-sales/{requestID}", r.handleAdminAfterSalesDetail)
	r.mux.HandleFunc("GET /api/admin/operations/snapshot", r.handleAdminOperationsSnapshot)
	r.mux.HandleFunc("GET /api/admin/audit-logs", r.handleAdminAuditLogs)
	r.mux.HandleFunc("GET /api/admin/audit-logs/export", r.handleAdminAuditLogsExport)
	r.mux.HandleFunc("GET /api/admin/audit-logs/retention-report", r.handleAdminAuditRetentionReport)
	r.mux.HandleFunc("POST /api/admin/audit-logs/retention-alerts/emit", r.handleAdminEmitAuditRetentionAlerts)
	r.mux.HandleFunc("POST /api/admin/audit-logs/archive/request", r.handleAdminRequestAuditArchive)
	r.mux.HandleFunc("GET /api/admin/audit-logs/archive/records", r.handleAdminAuditArchiveRecords)
	r.mux.HandleFunc("GET /api/admin/audit-logs/archive/verifications", r.handleAdminAuditArchiveVerifications)
	r.mux.HandleFunc("POST /api/admin/audit-logs/archive/complete", r.handleAdminCompleteAuditArchive)
	r.mux.HandleFunc("POST /api/admin/audit-logs/archive/verify", r.handleAdminVerifyAuditArchive)
	r.mux.HandleFunc("GET /api/admin/rbac/policy", r.handleAdminRBACPolicy)
	r.mux.HandleFunc("GET /api/admin/rbac/change-requests", r.handleAdminRBACChangeRequests)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests", r.handleAdminRBACChangeRequest)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/review", r.handleAdminRBACChangeRequestReview)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/apply", r.handleAdminRBACChangeRequestApply)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/rollback", r.handleAdminRBACChangeRequestRollback)
	r.mux.HandleFunc("GET /api/admin/merchant-qualifications", r.handleAdminMerchantQualifications)
	r.mux.HandleFunc("GET /api/admin/merchant-qualifications/{qualificationID}", r.handleAdminMerchantQualificationDetail)
	r.mux.HandleFunc("POST /api/admin/merchant-qualifications/{qualificationID}/review", r.handleAdminReviewMerchantQualification)
	r.mux.HandleFunc("GET /api/admin/notifications", r.handleAdminNotifications)
	r.mux.HandleFunc("GET /api/admin/notification-deliveries", r.handleAdminNotificationDeliveries)
	r.mux.HandleFunc("POST /api/admin/notification-deliveries/failure-alerts/emit", r.handleAdminEmitNotificationFailureAlerts)
	r.mux.HandleFunc("POST /api/admin/notification-deliveries/retries/schedule", r.handleAdminScheduleNotificationDeliveryRetries)
	r.mux.HandleFunc("POST /api/admin/notification-deliveries/quiet-window-retries/schedule", r.handleAdminScheduleNotificationQuietWindowRetries)
	r.mux.HandleFunc("GET /api/admin/notification-preferences", r.handleAdminNotificationPreferences)
	r.mux.HandleFunc("PUT /api/admin/notification-preferences", r.handleAdminSaveNotificationPreference)
	r.mux.HandleFunc("POST /api/admin/notification-preferences/batch", r.handleAdminSaveNotificationPreferenceBatch)
	r.mux.HandleFunc("GET /api/admin/notification-preferences/change-requests", r.handleAdminNotificationPreferenceChangeRequests)
	r.mux.HandleFunc("POST /api/admin/notification-preferences/change-requests", r.handleAdminNotificationPreferenceChangeRequest)
	r.mux.HandleFunc("POST /api/admin/notification-preferences/change-requests/{changeRequestID}/review", r.handleAdminNotificationPreferenceChangeRequestReview)
	r.mux.HandleFunc("POST /api/admin/notification-preferences/change-requests/{changeRequestID}/apply", r.handleAdminNotificationPreferenceChangeRequestApply)
	r.mux.HandleFunc("GET /api/admin/object-storage/cleanup-candidates", r.handleAdminObjectStorageCleanupCandidates)
	r.mux.HandleFunc("GET /api/admin/object-storage/cleanup-stats", r.handleAdminObjectStorageCleanupStats)
	r.mux.HandleFunc("POST /api/admin/object-storage/cleanup-complete", r.handleAdminObjectStorageCleanupComplete)
	r.mux.HandleFunc("POST /api/admin/object-storage/cleanup-failed", r.handleAdminObjectStorageCleanupFailed)
	r.mux.HandleFunc("POST /api/admin/orders/{orderID}/state/compensate", r.handleAdminCompensateOrderState)
	r.mux.HandleFunc("GET /api/admin/outbox/events", r.handleAdminOutboxEvents)
	r.mux.HandleFunc("GET /api/admin/outbox/events/{eventID}", r.handleAdminOutboxEventDetail)
	r.mux.HandleFunc("GET /api/admin/outbox/stats", r.handleAdminOutboxStats)
	r.mux.HandleFunc("POST /api/admin/outbox/events/claim", r.handleAdminClaimOutboxEvents)
	r.mux.HandleFunc("POST /api/admin/outbox/events/{eventID}/lease/renew", r.handleAdminRenewOutboxEventLease)
	r.mux.HandleFunc("POST /api/admin/outbox/events/{eventID}/published", r.handleAdminMarkOutboxEventPublished)
	r.mux.HandleFunc("POST /api/admin/outbox/events/{eventID}/failed", r.handleAdminMarkOutboxEventFailed)
	r.mux.HandleFunc("POST /api/admin/outbox/events/replay", r.handleAdminReplayOutboxEvents)
	r.mux.HandleFunc("POST /api/admin/outbox/events/{eventID}/replay", r.handleAdminReplayOutboxEvent)
	r.mux.HandleFunc("POST /api/station-manager/rider-invites", r.handleCreateRiderInvite)
	r.mux.HandleFunc("POST /api/auth/merchant/invite-register", r.handleAcceptMerchantInvite)
	r.mux.HandleFunc("POST /api/auth/merchant/login", r.handleMerchantLogin)
	r.mux.HandleFunc("POST /api/auth/rider/invite-register", r.handleAcceptRiderInvite)
	r.mux.HandleFunc("POST /api/auth/rider/login", r.handleRiderLogin)
	r.mux.HandleFunc("POST /api/auth/admin/login", r.handleAdminLogin)
	r.mux.HandleFunc("POST /api/notifications", r.handleCreateNotification)
	r.mux.HandleFunc("POST /api/notifications/provider-callback", r.handleNotificationProviderCallback)
	r.mux.HandleFunc("POST /api/notifications/{notificationID}/deliveries", r.handleRecordNotificationDelivery)
	r.mux.HandleFunc("GET /api/merchant/me", r.handleMerchantMe)
	r.mux.HandleFunc("GET /api/merchant/notifications", r.handleMerchantNotifications)
	r.mux.HandleFunc("POST /api/merchant/notifications/{notificationID}/read", r.handleMarkMerchantNotificationRead)
	r.mux.HandleFunc("GET /api/merchant/notification-preferences", r.handleMerchantNotificationPreferences)
	r.mux.HandleFunc("PUT /api/merchant/notification-preferences", r.handleSaveMerchantNotificationPreference)
	r.mux.HandleFunc("POST /api/merchant/qualifications", r.handleMerchantQualification)
	r.mux.HandleFunc("GET /api/merchant/staff", r.handleMerchantStaff)
	r.mux.HandleFunc("POST /api/merchant/staff", r.handleSaveMerchantStaff)
	r.mux.HandleFunc("GET /api/merchant/materials", r.handleMerchantMaterials)
	r.mux.HandleFunc("POST /api/merchant/materials", r.handleSaveMerchantMaterial)
	r.mux.HandleFunc("GET /api/merchant/orders", r.handleMerchantOrders)
	r.mux.HandleFunc("GET /api/merchant/after-sales", r.handleMerchantAfterSales)
	r.mux.HandleFunc("GET /api/merchant/deposit", r.handleMerchantDeposit)
	r.mux.HandleFunc("POST /api/merchant/deposit/pay", r.handlePayMerchantDeposit)
	r.mux.HandleFunc("POST /api/merchant/orders/{orderID}/accept", r.handleMerchantAcceptOrder)
	r.mux.HandleFunc("POST /api/merchant/orders/{orderID}/ready", r.handleMerchantOrderReady)
	r.mux.HandleFunc("GET /api/merchant/products", r.handleMerchantProducts)
	r.mux.HandleFunc("POST /api/merchant/products", r.handleUpsertMerchantProduct)
	r.mux.HandleFunc("POST /api/merchant/products/{productID}/status", r.handleMerchantProductStatus)
	r.mux.HandleFunc("GET /api/home/modules", r.handleHomeModules)
	r.mux.HandleFunc("GET /api/home/cards", r.handleHomeCards)
	r.mux.HandleFunc("GET /api/shops", r.handleShops)
	r.mux.HandleFunc("GET /api/shops/{shopID}/detail", r.handleShopDetail)
	r.mux.HandleFunc("GET /api/shops/{shopID}/products", r.handleShopProducts)
	r.mux.HandleFunc("GET /api/shops/{shopID}/groupbuy-deals", r.handleShopGroupbuyDeals)
	r.mux.HandleFunc("GET /api/user/addresses", r.handleUserAddresses)
	r.mux.HandleFunc("POST /api/user/addresses", r.handleSaveAddress)
	r.mux.HandleFunc("GET /api/user/profile", r.handleUserProfileOverview)
	r.mux.HandleFunc("GET /api/user/notification-preferences", r.handleUserNotificationPreferences)
	r.mux.HandleFunc("PUT /api/user/notification-preferences", r.handleSaveUserNotificationPreference)
	r.mux.HandleFunc("GET /api/user/coupons", r.handleUserCoupons)
	r.mux.HandleFunc("POST /api/user/coupons/claim", r.handleClaimUserCoupon)
	r.mux.HandleFunc("GET /api/user/points", r.handleUserPointsSummary)
	r.mux.HandleFunc("POST /api/user/points/check-in", r.handleCheckInPoints)
	r.mux.HandleFunc("GET /api/user/invite-summary", r.handleInviteSummary)
	r.mux.HandleFunc("GET /api/search", r.handleUserSearch)
	r.mux.HandleFunc("GET /api/medicine/home", r.handleMedicineHome)
	r.mux.HandleFunc("POST /api/prescriptions/upload-ticket", r.handleCreatePrescriptionImageUpload)
	r.mux.HandleFunc("POST /api/prescriptions/upload-confirm", r.handleConfirmPrescriptionImageUpload)
	r.mux.HandleFunc("POST /api/prescriptions", r.handleCreatePrescriptionReview)
	r.mux.HandleFunc("GET /api/prescriptions/{reviewID}", r.handlePrescriptionReview)
	r.mux.HandleFunc("POST /api/medicine/orders", r.handleCreateMedicineOrder)
	r.mux.HandleFunc("GET /api/medicine/orders/{orderID}", r.handleMedicineOrderDetail)
	r.mux.HandleFunc("POST /api/errand/orders", r.handleCreateErrandOrder)
	r.mux.HandleFunc("GET /api/errand/orders/{orderID}", r.handleErrandOrderDetail)
	r.mux.HandleFunc("GET /api/cart", r.handleCartSummary)
	r.mux.HandleFunc("POST /api/cart/items", r.handleUpsertCartItem)
	r.mux.HandleFunc("POST /api/groupbuy/orders", r.handleCreateGroupbuyOrder)
	r.mux.HandleFunc("GET /api/groupbuy/vouchers", r.handleUserGroupbuyVouchers)
	r.mux.HandleFunc("POST /api/merchant/groupbuy/vouchers/scan", r.handleMerchantScanGroupbuyVoucher)
	r.mux.HandleFunc("GET /api/orders", r.handleUserOrders)
	r.mux.HandleFunc("GET /api/orders/{orderID}", r.handleOrderDetail)
	r.mux.HandleFunc("POST /api/orders", r.handleCreateOrder)
	r.mux.HandleFunc("POST /api/orders/checkout", r.handleCheckoutCart)
	r.mux.HandleFunc("POST /api/orders/{orderID}/refund", r.handleAdminRefundOrder)
	r.mux.HandleFunc("GET /api/reviews", r.handleUserReviews)
	r.mux.HandleFunc("POST /api/reviews", r.handleCreateReview)
	r.mux.HandleFunc("POST /api/reviews/upload-ticket", r.handleCreateReviewImageUpload)
	r.mux.HandleFunc("POST /api/reviews/upload-confirm", r.handleConfirmReviewImageUpload)
	r.mux.HandleFunc("GET /api/after-sales", r.handleUserAfterSales)
	r.mux.HandleFunc("POST /api/after-sales", r.handleCreateAfterSales)
	r.mux.HandleFunc("GET /api/after-sales/{requestID}/events", r.handleAfterSalesEvents)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/events", r.handleAddAfterSalesEvent)
	r.mux.HandleFunc("GET /api/after-sales/{requestID}/evidence", r.handleAfterSalesEvidence)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/evidence/upload-ticket", r.handleCreateAfterSalesEvidenceUpload)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/evidence/confirm", r.handleConfirmAfterSalesEvidenceUpload)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/review", r.handleReviewAfterSales)
	r.mux.HandleFunc("GET /api/feedback", r.handleUserFeedbackTickets)
	r.mux.HandleFunc("POST /api/feedback", r.handleCreateFeedback)
	r.mux.HandleFunc("GET /api/service-tickets", r.handleUserServiceTickets)
	r.mux.HandleFunc("POST /api/service-tickets", r.handleCreateServiceTicket)
	r.mux.HandleFunc("GET /api/service-tickets/{ticketID}", r.handleServiceTicketDetail)
	r.mux.HandleFunc("POST /api/service-tickets/{ticketID}/events", r.handleAddServiceTicketEvent)
	r.mux.HandleFunc("POST /api/service-tickets/{ticketID}/close", r.handleCloseServiceTicket)
	r.mux.HandleFunc("POST /api/service-tickets/{ticketID}/follow-up", r.handleFollowUpServiceTicket)
	r.mux.HandleFunc("GET /api/admin/service-tickets", r.handleAdminServiceTickets)
	r.mux.HandleFunc("GET /api/admin/service-tickets/{ticketID}", r.handleAdminServiceTicketDetail)
	r.mux.HandleFunc("GET /api/admin/service-ticket-quality-reviews", r.handleServiceTicketQualityReviews)
	r.mux.HandleFunc("GET /api/admin/service-ticket-performance", r.handleServiceTicketPerformance)
	r.mux.HandleFunc("POST /api/admin/service-tickets/{ticketID}/assign", r.handleAssignServiceTicket)
	r.mux.HandleFunc("POST /api/admin/service-tickets/{ticketID}/resolve", r.handleResolveServiceTicket)
	r.mux.HandleFunc("POST /api/admin/service-tickets/{ticketID}/escalate", r.handleEscalateServiceTicket)
	r.mux.HandleFunc("POST /api/admin/service-tickets/{ticketID}/quality-review", r.handleReviewServiceTicketQuality)
	r.mux.HandleFunc("GET /api/admin/prescriptions", r.handleAdminPrescriptionReviews)
	r.mux.HandleFunc("POST /api/admin/prescriptions/{reviewID}/review", r.handleAdminReviewPrescription)
	r.mux.HandleFunc("GET /api/circle/posts", r.handleCirclePosts)
	r.mux.HandleFunc("POST /api/circle/posts", r.handleCreateCirclePost)
	r.mux.HandleFunc("GET /api/meal-match/profile", r.handleMealMatchProfile)
	r.mux.HandleFunc("PUT /api/meal-match/profile", r.handleSaveMealMatchProfile)
	r.mux.HandleFunc("GET /api/meal-match/candidates", r.handleMealMatchCandidates)
	r.mux.HandleFunc("POST /api/meal-match/reports", r.handleReportMealMatchCandidate)
	r.mux.HandleFunc("POST /api/meal-match/blocks", r.handleBlockMealMatchCandidate)
	r.mux.HandleFunc("GET /api/admin/meal-match/moderation", r.handleAdminMealMatchModerationRecords)
	r.mux.HandleFunc("POST /api/admin/meal-match/moderation/{recordID}/review", r.handleAdminReviewMealMatchModeration)
	r.mux.HandleFunc("POST /api/red-packets", r.handleCreateRedPacket)
	r.mux.HandleFunc("GET /api/red-packets/{packetID}", r.handleRedPacketDetail)
	r.mux.HandleFunc("POST /api/red-packets/{packetID}/claim", r.handleClaimRedPacket)
	r.mux.HandleFunc("POST /api/red-packets/{packetID}/refund", r.handleRefundRedPacket)
	r.mux.HandleFunc("POST /api/admin/red-packets/expire", r.handleAutoRefundExpiredRedPackets)
	r.mux.HandleFunc("GET /api/messages/threads", r.handleMessageThreads)
	r.mux.HandleFunc("GET /api/messages/{threadID}/overview", r.handleChatThreadOverview)
	r.mux.HandleFunc("GET /api/messages/{threadID}/members", r.handleChatThreadMembers)
	r.mux.HandleFunc("GET /api/messages/{threadID}/membership", r.handleChatThreadMembership)
	r.mux.HandleFunc("POST /api/messages/{threadID}/join", r.handleJoinChatThread)
	r.mux.HandleFunc("POST /api/messages/{threadID}/leave", r.handleLeaveChatThread)
	r.mux.HandleFunc("GET /api/messages/{threadID}/sync", r.handleChatMessageSync)
	r.mux.HandleFunc("GET /api/messages/{threadID}/preference", r.handleChatThreadPreference)
	r.mux.HandleFunc("PUT /api/messages/{threadID}/preference", r.handleUpdateChatThreadPreference)
	r.mux.HandleFunc("POST /api/messages/{threadID}/read", r.handleMarkChatThreadRead)
	r.mux.HandleFunc("GET /api/messages/{threadID}", r.handleChatMessages)
	r.mux.HandleFunc("POST /api/messages/{threadID}", r.handleSendChatMessage)
	r.mux.HandleFunc("GET /api/wallet/transactions", r.handleWalletTransactions)
	r.mux.HandleFunc("GET /api/wallet/overview", r.handleWalletOverview)
	r.mux.HandleFunc("POST /api/wallet/credit", r.handleCreditWallet)
	r.mux.HandleFunc("POST /api/wallet/withdraw", r.handleWalletWithdraw)
	r.mux.HandleFunc("POST /api/wallet/payment-password", r.handleSetPaymentPassword)
	r.mux.HandleFunc("POST /api/wallet/pay", r.handleBalancePay)
	r.mux.HandleFunc("POST /api/payments/wechat/prepay", r.handleWechatPrepay)
	r.mux.HandleFunc("POST /api/payments/wechat/callback", r.handleWechatCallback)
	r.mux.HandleFunc("POST /api/object-storage/upload-callback", r.handleObjectStorageUploadCallback)
	r.mux.HandleFunc("POST /api/object-storage/scan-result", r.handleObjectStorageScanResult)
	r.mux.HandleFunc("GET /api/rider/deposit", r.handleRiderDeposit)
	r.mux.HandleFunc("POST /api/rider/deposit/pay", r.handlePayRiderDeposit)
	r.mux.HandleFunc("POST /api/rider/deposit/wechat-exempt", r.handleRiderWechatExemption)
	r.mux.HandleFunc("POST /api/rider/deposit/refund-request", r.handleRiderDepositRefund)
	r.mux.HandleFunc("POST /api/rider/online", r.handleRiderOnline)
	r.mux.HandleFunc("POST /api/dispatch/orders/{orderID}/auto-assign", r.handleAutoAssignOrder)
	r.mux.HandleFunc("POST /api/dispatch/orders/{orderID}/timeout-reassign", r.handleTimeoutReassignOrder)
	r.mux.HandleFunc("GET /api/dispatch/orders/{orderID}/events", r.handleDispatchEvents)
	r.mux.HandleFunc("POST /api/rider/orders/{orderID}/reject-assignment", r.handleRejectRiderAssignment)
	r.mux.HandleFunc("GET /api/station-manager/riders", r.handleStationManagerRiders)
	r.mux.HandleFunc("GET /api/station-manager/orders", r.handleStationManagerOrders)
	r.mux.HandleFunc("POST /api/station-manager/dispatch/{orderID}/manual-assign", r.handleStationManagerManualAssign)
	r.mux.HandleFunc("GET /api/station-manager/task-duration", r.handleStationManagerTaskConfig)
	r.mux.HandleFunc("PUT /api/station-manager/task-duration", r.handleSaveStationManagerTaskConfig)
	r.mux.HandleFunc("GET /api/station-manager/rider-performance", r.handleStationManagerRiderPerformance)
	r.mux.HandleFunc("POST /api/rider/dispatch/cancel-free", r.handleFreeDispatchCancel)
	r.mux.HandleFunc("POST /api/rider/orders/", r.handleRiderOrderAction)
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, map[string]string{"status": "ok", "service": "api-go"})
}

func (r *Router) handleReady(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, map[string]string{"status": "ready", "service": "api-go"})
}

func (r *Router) handleWechatMiniLogin(w http.ResponseWriter, req *http.Request) {
	var payload platform.WechatMiniLoginRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	session, err := r.wechatMini.Resolve(req.Context(), payload.Code)
	if err != nil {
		writeWechatMiniLoginError(w, err)
		return
	}
	payload.ProviderOpenID = session.OpenID
	payload.ProviderUnionID = session.UnionID
	result, err := r.store.LoginWechatMini(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: result.User.ID, Role: RoleUser}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"user":         result.User,
		"provider":     result.Provider,
		"is_new_user":  result.IsNewUser,
	})
}

func (r *Router) handleSendPhoneVerificationCode(w http.ResponseWriter, req *http.Request) {
	var payload platform.SendPhoneVerificationCodeRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	ticket, err := r.store.SendPhoneVerificationCode(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, ticket)
}

func (r *Router) handlePhoneLogin(w http.ResponseWriter, req *http.Request) {
	var payload platform.PhoneLoginRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	result, err := r.store.LoginWithPhone(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	r.writePhoneAuthResult(w, req, result)
}

func (r *Router) handlePhoneRegister(w http.ResponseWriter, req *http.Request) {
	var payload platform.PhoneRegisterRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	result, err := r.store.RegisterWithPhone(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	r.writePhoneAuthResult(w, req, result)
}

func (r *Router) writePhoneAuthResult(w http.ResponseWriter, req *http.Request, result *platform.PhoneAuthResult) {
	if result == nil {
		writePlatformError(w, platform.ErrInvalidCredentials)
		return
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: result.User.ID, Role: RoleUser}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"user":         result.User,
		"provider":     result.Provider,
		"is_new_user":  result.IsNewUser,
	})
}

func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	token, claims, err := r.tokenSigner.verifyRequestClaims(req)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	principal := Principal{ID: claims.SubjectID, Role: claims.Role}
	sessionID := strings.TrimSpace(claims.SessionID)
	if sessionID == "" || r.authSessions == nil {
		writeAuthError(w, errUnauthorized)
		return
	}
	if err := r.authSessions.Revoke(req.Context(), sessionID, tokenHash(token), principal, time.Now().UTC()); err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]bool{"revoked": true})
}

func (r *Router) handleCreateMerchantInvite(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateMerchantInviteRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageInvites() {
		writeAuthError(w, errForbidden)
		return
	}
	if strings.TrimSpace(payload.AdminID) == "" {
		payload.AdminID = principal.ID
	}
	invite, _, err := r.store.CreateMerchantInviteWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.merchant_invite.created",
		TargetType: "merchant_invite",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, invite)
}

func (r *Router) handleCreateRiderInvite(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateRiderInviteRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageInvites() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	payload.CreatedByID = principal.ID
	payload.CreatedByRole = principal.Role
	if principal.Role == RoleStationManager {
		if requestedType := strings.TrimSpace(payload.Type); requestedType != "" && requestedType != platform.RiderAccountRider {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		payload.Type = platform.RiderAccountRider
	} else if strings.TrimSpace(payload.StationID) == "" {
		payload.StationID = "station_1"
	}
	invite, _, err := r.store.CreateRiderInviteWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.rider_invite.created",
		TargetType: "rider_invite",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, invite)
}

func (r *Router) handleAcceptMerchantInvite(w http.ResponseWriter, req *http.Request) {
	var payload platform.AcceptMerchantInviteRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	profile, err := r.store.AcceptMerchantInvite(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: profile.Account.ID, Role: RoleMerchant}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"profile":      profile,
	})
}

func (r *Router) handleMerchantLogin(w http.ResponseWriter, req *http.Request) {
	var payload platform.MerchantLoginRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	profile, err := r.store.LoginMerchant(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: profile.Account.ID, Role: RoleMerchant}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"profile":      profile,
	})
}

func (r *Router) handleAcceptRiderInvite(w http.ResponseWriter, req *http.Request) {
	var payload platform.AcceptRiderInviteRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	rider, err := r.store.AcceptRiderInvite(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	role := RoleRider
	if rider.Type == platform.RiderAccountStationManager {
		role = RoleStationManager
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: rider.ID, Role: role}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"rider":        rider,
	})
}

func (r *Router) handleRiderLogin(w http.ResponseWriter, req *http.Request) {
	var payload platform.RiderLoginRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	rider, err := r.store.LoginRider(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	role := RoleRider
	if rider.Type == platform.RiderAccountStationManager {
		role = RoleStationManager
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: rider.ID, Role: role}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"rider":        rider,
	})
}

func (r *Router) handleAdminLogin(w http.ResponseWriter, req *http.Request) {
	var payload platform.AdminLoginRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	accountID := strings.TrimSpace(payload.AccountID)
	password := strings.TrimSpace(payload.Password)
	if accountID == "" || password == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	passwordHash := ""
	if r.adminPasswordHash != nil {
		passwordHash = r.adminPasswordHash[accountID]
	}
	if passwordHash == "" {
		writePlatformError(w, platform.ErrInvalidCredentials)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		writePlatformError(w, platform.ErrInvalidCredentials)
		return
	}
	token, expiresAt, err := r.issueAccessToken(req, Principal{ID: accountID, Role: RoleAdmin}, 30*24*time.Hour)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt,
		"admin": map[string]string{
			"id":   accountID,
			"role": RoleAdmin,
		},
	})
}

func (r *Router) handleMerchantMe(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleMerchant && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	merchantID := principal.ID
	if principal.IsAdmin() {
		merchantID = req.URL.Query().Get("merchant_id")
	}
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	profile, err := r.store.MerchantProfile(merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, profile)
}

func (r *Router) handleCreateNotification(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateNotificationRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	notification, err := r.store.CreateNotification(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, notification)
}

func (r *Router) handleAdminNotifications(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	notifications, err := r.store.Notifications(platform.NotificationListRequest{
		TargetRole:    req.URL.Query().Get("target_role"),
		TargetID:      req.URL.Query().Get("target_id"),
		Status:        req.URL.Query().Get("status"),
		SourceTopic:   req.URL.Query().Get("source_topic"),
		SourceEventID: req.URL.Query().Get("source_event_id"),
		Limit:         limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, notifications)
}

func (r *Router) handleRecordNotificationDelivery(w http.ResponseWriter, req *http.Request) {
	var payload platform.RecordNotificationDeliveryRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.NotificationID = req.PathValue("notificationID")
	delivery, err := r.store.RecordNotificationDelivery(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, delivery)
}

type notificationProviderCallbackPayload struct {
	NotificationID    string    `json:"notification_id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	Status            string    `json:"status"`
	ProviderMessageID string    `json:"provider_message_id"`
	ErrorCode         string    `json:"error_code"`
	ErrorMessage      string    `json:"error_message"`
	IdempotencyKey    string    `json:"idempotency_key"`
	AttemptedAt       time.Time `json:"attempted_at"`
	DeliveredAt       time.Time `json:"delivered_at"`
	CallbackAt        time.Time `json:"callback_at"`
	Signature         string    `json:"signature"`
}

func (r *Router) handleNotificationProviderCallback(w http.ResponseWriter, req *http.Request) {
	var payload notificationProviderCallbackPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	signatureVerified, err := verifyNotificationProviderCallbackSignature(payload, r.notificationProviderCallbackSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_NOTIFICATION_PROVIDER_SIGNATURE", "invalid notification provider signature")
		return
	}
	deliveryReq := payload.toRecordNotificationDeliveryRequest(time.Now().UTC())
	delivery, err := r.store.RecordNotificationDelivery(deliveryReq)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{
		"delivery":           delivery,
		"signature_verified": signatureVerified,
		"provider":           delivery.Provider,
		"channel":            delivery.Channel,
	})
}

func (payload notificationProviderCallbackPayload) toRecordNotificationDeliveryRequest(now time.Time) platform.RecordNotificationDeliveryRequest {
	payload.NotificationID = strings.TrimSpace(payload.NotificationID)
	payload.Channel = strings.TrimSpace(payload.Channel)
	payload.Provider = strings.TrimSpace(payload.Provider)
	payload.Status = strings.ToLower(strings.TrimSpace(payload.Status))
	payload.ProviderMessageID = strings.TrimSpace(payload.ProviderMessageID)
	payload.ErrorCode = strings.TrimSpace(payload.ErrorCode)
	payload.ErrorMessage = strings.TrimSpace(payload.ErrorMessage)
	payload.IdempotencyKey = strings.TrimSpace(payload.IdempotencyKey)
	if payload.Provider == "" {
		payload.Provider = payload.Channel
	}
	if payload.Channel == "" {
		payload.Channel = payload.Provider
	}
	if payload.Status == "" {
		payload.Status = platform.NotificationDeliveryDelivered
	}
	if payload.CallbackAt.IsZero() {
		payload.CallbackAt = now
	} else {
		payload.CallbackAt = payload.CallbackAt.UTC()
	}
	if payload.AttemptedAt.IsZero() {
		payload.AttemptedAt = payload.CallbackAt
	} else {
		payload.AttemptedAt = payload.AttemptedAt.UTC()
	}
	if !payload.DeliveredAt.IsZero() {
		payload.DeliveredAt = payload.DeliveredAt.UTC()
	}
	if payload.Status == platform.NotificationDeliveryDelivered && payload.DeliveredAt.IsZero() {
		payload.DeliveredAt = payload.AttemptedAt
	}
	if payload.Status == platform.NotificationDeliveryFailed {
		if payload.ErrorCode == "" {
			payload.ErrorCode = "provider_failed"
		}
		if payload.ErrorMessage == "" {
			payload.ErrorMessage = "provider delivery failed"
		}
	}
	if payload.IdempotencyKey == "" {
		messageID := payload.ProviderMessageID
		if messageID != "" {
			payload.IdempotencyKey = "provider_callback:" + payload.Provider + ":" + messageID
		} else {
			payload.IdempotencyKey = "provider_callback:" + payload.Provider + ":" + payload.NotificationID + ":" + strconv.FormatInt(payload.CallbackAt.Unix(), 10)
		}
	}
	return platform.RecordNotificationDeliveryRequest{
		NotificationID:    payload.NotificationID,
		Channel:           payload.Channel,
		Provider:          payload.Provider,
		Status:            payload.Status,
		ProviderMessageID: payload.ProviderMessageID,
		ErrorCode:         payload.ErrorCode,
		ErrorMessage:      payload.ErrorMessage,
		IdempotencyKey:    payload.IdempotencyKey,
		AttemptedAt:       payload.AttemptedAt,
		DeliveredAt:       payload.DeliveredAt,
	}
}

func verifyNotificationProviderCallbackSignature(payload notificationProviderCallbackPayload, secret string) (bool, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return false, nil
	}
	signature := strings.ToLower(strings.TrimSpace(payload.Signature))
	if signature == "" {
		return false, platform.ErrInvalidArgument
	}
	expected := signNotificationProviderCallback(payload, secret)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return false, platform.ErrInvalidArgument
	}
	return true, nil
}

func signNotificationProviderCallback(payload notificationProviderCallbackPayload, secret string) string {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	_, _ = mac.Write([]byte(strings.Join(notificationProviderCallbackCanonicalLines(payload), "\n")))
	return hex.EncodeToString(mac.Sum(nil))
}

func notificationProviderCallbackCanonicalLines(payload notificationProviderCallbackPayload) []string {
	return []string{
		strings.TrimSpace(payload.NotificationID),
		strings.TrimSpace(payload.Channel),
		strings.TrimSpace(payload.Provider),
		strings.TrimSpace(payload.Status),
		strings.TrimSpace(payload.ProviderMessageID),
		strings.TrimSpace(payload.ErrorCode),
		strings.TrimSpace(payload.ErrorMessage),
		strings.TrimSpace(payload.IdempotencyKey),
		strconv.FormatInt(notificationProviderCallbackUnix(payload.AttemptedAt), 10),
		strconv.FormatInt(notificationProviderCallbackUnix(payload.DeliveredAt), 10),
		strconv.FormatInt(notificationProviderCallbackUnix(payload.CallbackAt), 10),
	}
}

func notificationProviderCallbackUnix(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UTC().Unix()
}

func (r *Router) handleAdminNotificationDeliveries(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	retryAtBefore, ok := parseOptionalTimeQuery(w, req.URL.Query().Get("retry_at_before"))
	if !ok {
		return
	}
	deliveries, err := r.store.NotificationDeliveries(platform.NotificationDeliveryListRequest{
		NotificationID: req.URL.Query().Get("notification_id"),
		TargetRole:     req.URL.Query().Get("target_role"),
		TargetID:       req.URL.Query().Get("target_id"),
		Channel:        req.URL.Query().Get("channel"),
		Provider:       req.URL.Query().Get("provider"),
		Status:         req.URL.Query().Get("status"),
		ErrorCode:      req.URL.Query().Get("error_code"),
		RetryAtBefore:  retryAtBefore,
		Limit:          limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deliveries)
}

type adminNotificationFailureAlertPayload struct {
	TargetRole string    `json:"target_role"`
	TargetID   string    `json:"target_id"`
	Channel    string    `json:"channel"`
	Provider   string    `json:"provider"`
	Limit      int       `json:"limit"`
	Now        time.Time `json:"now"`
}

type adminNotificationDeliveryRetryPayload struct {
	TargetRole        string    `json:"target_role"`
	TargetID          string    `json:"target_id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	Status            string    `json:"status"`
	ErrorCode         string    `json:"error_code"`
	Limit             int       `json:"limit"`
	RetryAfterSeconds int       `json:"retry_after_seconds"`
	RetryAt           time.Time `json:"retry_at"`
	Now               time.Time `json:"now"`
}

type adminNotificationQuietWindowRetryPayload struct {
	TargetRole        string    `json:"target_role"`
	TargetID          string    `json:"target_id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	Limit             int       `json:"limit"`
	RetryAfterSeconds int       `json:"retry_after_seconds"`
	Now               time.Time `json:"now"`
}

func (r *Router) handleAdminEmitNotificationFailureAlerts(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload adminNotificationFailureAlertPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.Limit < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	emission, event, audit, err := r.store.EmitNotificationFailureAlerts(platform.NotificationFailureAlertEmissionRequest{
		TargetRole: payload.TargetRole,
		TargetID:   payload.TargetID,
		Channel:    payload.Channel,
		Provider:   payload.Provider,
		Limit:      payload.Limit,
		Now:        payload.Now,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_delivery_failure_alerts.emitted",
		TargetType: "notification_delivery_alerts",
		TargetID:   "failed",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"emission":     emission,
		"outbox_event": event,
		"audit_log":    audit,
	})
}

func (r *Router) handleAdminScheduleNotificationDeliveryRetries(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload adminNotificationDeliveryRetryPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.Limit < 0 || payload.RetryAfterSeconds < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	targetStatus := strings.ToLower(strings.TrimSpace(payload.Status))
	if targetStatus == "" {
		targetStatus = platform.NotificationDeliveryFailed
	}
	schedule, event, audit, err := r.store.ScheduleNotificationDeliveryRetries(platform.NotificationDeliveryRetryScheduleRequest{
		TargetRole:        payload.TargetRole,
		TargetID:          payload.TargetID,
		Channel:           payload.Channel,
		Provider:          payload.Provider,
		Status:            targetStatus,
		ErrorCode:         payload.ErrorCode,
		Limit:             payload.Limit,
		RetryAfterSeconds: payload.RetryAfterSeconds,
		RetryAt:           payload.RetryAt,
		Now:               payload.Now,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_delivery_retries.scheduled",
		TargetType: "notification_delivery_retries",
		TargetID:   targetStatus,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"schedule":     schedule,
		"outbox_event": event,
		"audit_log":    audit,
	})
}

func (r *Router) handleAdminScheduleNotificationQuietWindowRetries(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload adminNotificationQuietWindowRetryPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.Limit < 0 || payload.RetryAfterSeconds < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	schedule, event, audit, err := r.store.ScheduleNotificationQuietWindowRetries(platform.NotificationQuietWindowRetryScheduleRequest{
		TargetRole:        payload.TargetRole,
		TargetID:          payload.TargetID,
		Channel:           payload.Channel,
		Provider:          payload.Provider,
		Limit:             payload.Limit,
		RetryAfterSeconds: payload.RetryAfterSeconds,
		Now:               payload.Now,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_delivery_retries.scheduled",
		TargetType: "notification_delivery_retries",
		TargetID:   platform.NotificationDeliveryQueued,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"schedule":     schedule,
		"outbox_event": event,
		"audit_log":    audit,
	})
}

func (r *Router) handleAdminNotificationPreferences(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	preferences, err := r.store.NotificationPreferences(platform.NotificationPreferenceListRequest{
		PreferenceKey:    req.URL.Query().Get("preference_key"),
		TargetRole:       req.URL.Query().Get("target_role"),
		TargetID:         req.URL.Query().Get("target_id"),
		NotificationType: req.URL.Query().Get("notification_type"),
		Limit:            limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preferences)
}

func (r *Router) handleAdminSaveNotificationPreference(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload platform.SaveNotificationPreferenceRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	preference, audit, err := r.store.SaveNotificationPreferenceWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_preferences.saved",
		TargetType: "notification_preference",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.UpdatedAt,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"preference": preference,
		"audit_log":  audit,
	})
}

func (r *Router) handleAdminSaveNotificationPreferenceBatch(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload platform.SaveNotificationPreferenceBatchRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	result, audit, err := r.store.SaveNotificationPreferenceBatchWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_preferences.batch_saved",
		TargetType: "notification_preference_batch",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.UpdatedAt,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"batch":     result,
		"audit_log": audit,
	})
}

type adminNotificationPreferenceChangeRequestPayload struct {
	Preferences []platform.SaveNotificationPreferenceRequest `json:"preferences"`
	Reason      string                                       `json:"reason"`
	Rollout     adminNotificationPreferenceRolloutPolicy     `json:"rollout"`
	UpdatedAt   time.Time                                    `json:"updated_at"`
}

type adminNotificationPreferenceRolloutPolicy struct {
	Mode       string   `json:"mode"`
	Percentage int      `json:"percentage,omitempty"`
	TargetIDs  []string `json:"target_ids,omitempty"`
	MaxTargets int      `json:"max_targets,omitempty"`
}

type adminNotificationPreferenceChangeReviewPayload struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type adminNotificationPreferenceChangeApplyPayload struct {
	Reason    string    `json:"reason"`
	UpdatedAt time.Time `json:"updated_at"`
}

type adminNotificationPreferenceChangeRequestRecord struct {
	ID               string                                       `json:"id"`
	Status           string                                       `json:"status"`
	Reason           string                                       `json:"reason"`
	Preferences      []platform.SaveNotificationPreferenceRequest `json:"preferences"`
	PreferenceKeys   []string                                     `json:"preference_keys"`
	Rollout          adminNotificationPreferenceRolloutPolicy     `json:"rollout"`
	RequestedCount   int                                          `json:"requested_count"`
	ApprovalRequired bool                                         `json:"approval_required"`
	AutoApplied      bool                                         `json:"auto_applied"`
	RequestedByRole  string                                       `json:"requested_by_role"`
	RequestedByAdmin string                                       `json:"requested_by_admin"`
	RequestedAt      time.Time                                    `json:"requested_at"`
	RequestAuditID   string                                       `json:"request_audit_id"`
	UpdatedAt        time.Time                                    `json:"updated_at"`
	ReviewDecision   string                                       `json:"review_decision,omitempty"`
	ReviewReason     string                                       `json:"review_reason,omitempty"`
	ReviewedByRole   string                                       `json:"reviewed_by_role,omitempty"`
	ReviewedByAdmin  string                                       `json:"reviewed_by_admin,omitempty"`
	ReviewedAt       *time.Time                                   `json:"reviewed_at,omitempty"`
	ReviewAuditID    string                                       `json:"review_audit_id,omitempty"`
	Applied          bool                                         `json:"applied"`
	ApplyReason      string                                       `json:"apply_reason,omitempty"`
	AppliedByRole    string                                       `json:"applied_by_role,omitempty"`
	AppliedByAdmin   string                                       `json:"applied_by_admin,omitempty"`
	AppliedAt        *time.Time                                   `json:"applied_at,omitempty"`
	ApplyAuditID     string                                       `json:"apply_audit_id,omitempty"`
	BatchID          string                                       `json:"batch_id,omitempty"`
	BatchSaved       int                                          `json:"batch_saved,omitempty"`
	AppliedKeys      []string                                     `json:"applied_preference_keys,omitempty"`
	SkippedKeys      []string                                     `json:"skipped_preference_keys,omitempty"`
	SkippedCount     int                                          `json:"skipped_count,omitempty"`
}

const (
	adminNotificationPreferenceChangeStatusPending  = "pending_approval"
	adminNotificationPreferenceChangeStatusApproved = "approved"
	adminNotificationPreferenceChangeStatusRejected = "rejected"
	adminNotificationPreferenceChangeStatusApplied  = "applied"
	adminNotificationPreferenceReviewApprove        = "approve"
	adminNotificationPreferenceReviewReject         = "reject"
	adminNotificationPreferenceRolloutAll           = "all"
	adminNotificationPreferenceRolloutTargetIDs     = "target_ids"
	adminNotificationPreferenceRolloutPercentage    = "percentage"
)

func (r *Router) handleAdminNotificationPreferenceChangeRequests(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	status := strings.TrimSpace(query.Get("status"))
	if status != "" && !isAdminNotificationPreferenceChangeStatus(status) {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	requests, err := r.adminNotificationPreferenceChangeRequestLedger()
	if err != nil {
		writePlatformError(w, err)
		return
	}
	filtered := make([]adminNotificationPreferenceChangeRequestRecord, 0, len(requests))
	counts := map[string]int{
		adminNotificationPreferenceChangeStatusPending:  0,
		adminNotificationPreferenceChangeStatusApproved: 0,
		adminNotificationPreferenceChangeStatusRejected: 0,
		adminNotificationPreferenceChangeStatusApplied:  0,
	}
	for _, item := range requests {
		if _, ok := counts[item.Status]; ok {
			counts[item.Status]++
		}
		if status == "" || item.Status == status {
			filtered = append(filtered, item)
		}
	}
	filteredTotal := len(filtered)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	writeSuccess(w, map[string]any{
		"items":          filtered,
		"total":          len(requests),
		"filtered_total": filteredTotal,
		"pending_count":  counts[adminNotificationPreferenceChangeStatusPending],
		"approved_count": counts[adminNotificationPreferenceChangeStatusApproved],
		"rejected_count": counts[adminNotificationPreferenceChangeStatusRejected],
		"applied_count":  counts[adminNotificationPreferenceChangeStatusApplied],
		"auto_apply":     false,
		"manual_apply":   true,
	})
}

func (r *Router) handleAdminNotificationPreferenceChangeRequest(w http.ResponseWriter, req *http.Request) {
	var payload adminNotificationPreferenceChangeRequestPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	batch, preferenceKeys, err := normalizeAdminNotificationPreferenceChangeBatch(payload.Preferences, payload.Reason, payload.UpdatedAt)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	rollout, err := normalizeAdminNotificationPreferenceRolloutPolicy(payload.Rollout, batch.Preferences)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	changeRequestID := "ntfp_change_" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_preferences.change_requested",
		TargetType: "notification_preference_change_request",
		TargetID:   changeRequestID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"change_request_id":   changeRequestID,
			"preference_keys":     preferenceKeys,
			"preference_requests": batch.Preferences,
			"reason":              batch.Reason,
			"requested_count":     len(batch.Preferences),
			"rollout":             rollout,
			"rollout_mode":        rollout.Mode,
			"rollout_percentage":  rollout.Percentage,
			"rollout_target_ids":  rollout.TargetIDs,
			"rollout_max_targets": rollout.MaxTargets,
			"status":              adminNotificationPreferenceChangeStatusPending,
			"updated_at":          batch.UpdatedAt.Format(time.RFC3339Nano),
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{
		"id":                 changeRequestID,
		"status":             adminNotificationPreferenceChangeStatusPending,
		"reason":             batch.Reason,
		"preferences":        batch.Preferences,
		"preference_keys":    preferenceKeys,
		"rollout":            rollout,
		"requested_count":    len(batch.Preferences),
		"approval_required":  true,
		"auto_applied":       false,
		"audit_log":          audit,
		"requested_by_role":  principal.Role,
		"requested_by_admin": principal.ID,
		"updated_at":         batch.UpdatedAt,
	})
}

func (r *Router) handleAdminNotificationPreferenceChangeRequestReview(w http.ResponseWriter, req *http.Request) {
	var payload adminNotificationPreferenceChangeReviewPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	changeRequestID := strings.TrimSpace(req.PathValue("changeRequestID"))
	if changeRequestID == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	decision, status, ok := normalizeAdminNotificationPreferenceReviewDecision(payload.Decision)
	if !ok {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	record, err := r.adminNotificationPreferenceChangeRequestByID(changeRequestID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if record.Status != adminNotificationPreferenceChangeStatusPending {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	if record.RequestedByAdmin == principal.ID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_preferences.change_reviewed",
		TargetType: "notification_preference_change_request",
		TargetID:   changeRequestID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"change_request_id": changeRequestID,
			"decision":          decision,
			"preference_keys":   record.PreferenceKeys,
			"reason":            reason,
			"requested_count":   record.RequestedCount,
			"rollout":           record.Rollout,
			"rollout_mode":      record.Rollout.Mode,
			"status":            status,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	reviewedAt := audit.CreatedAt
	record.Status = status
	record.ReviewDecision = decision
	record.ReviewReason = reason
	record.ReviewedByRole = principal.Role
	record.ReviewedByAdmin = principal.ID
	record.ReviewedAt = &reviewedAt
	record.ReviewAuditID = audit.ID
	writeSuccess(w, map[string]any{
		"change_request": record,
		"audit_log":      audit,
		"auto_applied":   false,
	})
}

func (r *Router) handleAdminNotificationPreferenceChangeRequestApply(w http.ResponseWriter, req *http.Request) {
	var payload adminNotificationPreferenceChangeApplyPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanWriteNotifications() {
		writeAuthError(w, errForbidden)
		return
	}
	changeRequestID := strings.TrimSpace(req.PathValue("changeRequestID"))
	if changeRequestID == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	record, err := r.adminNotificationPreferenceChangeRequestByID(changeRequestID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if record.Status != adminNotificationPreferenceChangeStatusApproved {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	if record.RequestedByAdmin == principal.ID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	rolledPreferences, appliedPreferenceKeys, skippedPreferenceKeys, err := applyAdminNotificationPreferenceRollout(record.Preferences, record.Rollout)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	batch, audit, err := r.store.SaveNotificationPreferenceBatchWithAudit(platform.SaveNotificationPreferenceBatchRequest{
		Preferences: rolledPreferences,
		Reason:      record.Reason,
		UpdatedAt:   payload.UpdatedAt,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.notification_preferences.change_applied",
		TargetType: "notification_preference_batch",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.UpdatedAt,
		Payload: map[string]any{
			"apply_reason":            reason,
			"applied_count":           len(appliedPreferenceKeys),
			"applied_preference_keys": appliedPreferenceKeys,
			"change_request_id":       changeRequestID,
			"requested_count":         record.RequestedCount,
			"request_reason":          record.Reason,
			"rollout":                 record.Rollout,
			"rollout_max_targets":     record.Rollout.MaxTargets,
			"rollout_mode":            record.Rollout.Mode,
			"rollout_percentage":      record.Rollout.Percentage,
			"rollout_target_ids":      record.Rollout.TargetIDs,
			"skipped_count":           len(skippedPreferenceKeys),
			"skipped_preference_keys": skippedPreferenceKeys,
			"status":                  adminNotificationPreferenceChangeStatusApplied,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	appliedAt := audit.CreatedAt
	record.Status = adminNotificationPreferenceChangeStatusApplied
	record.Applied = true
	record.ApplyReason = reason
	record.AppliedByRole = principal.Role
	record.AppliedByAdmin = principal.ID
	record.AppliedAt = &appliedAt
	record.ApplyAuditID = audit.ID
	record.BatchID = batch.BatchID
	record.BatchSaved = batch.Saved
	record.AppliedKeys = appliedPreferenceKeys
	record.SkippedKeys = skippedPreferenceKeys
	record.SkippedCount = len(skippedPreferenceKeys)
	writeSuccess(w, map[string]any{
		"change_request": record,
		"batch":          batch,
		"audit_log":      audit,
		"auto_applied":   false,
		"applied":        true,
	})
}

func normalizeAdminNotificationPreferenceChangeBatch(preferences []platform.SaveNotificationPreferenceRequest, reason string, updatedAt time.Time) (platform.SaveNotificationPreferenceBatchRequest, []string, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(preferences) == 0 || len(preferences) > 50 {
		return platform.SaveNotificationPreferenceBatchRequest{}, nil, platform.ErrInvalidArgument
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	} else {
		updatedAt = updatedAt.UTC()
	}
	normalized := make([]platform.SaveNotificationPreferenceRequest, 0, len(preferences))
	preferenceKeys := make([]string, 0, len(preferences))
	seen := map[string]bool{}
	for _, item := range preferences {
		item.TargetRole = strings.TrimSpace(item.TargetRole)
		item.TargetID = strings.TrimSpace(item.TargetID)
		item.NotificationType = strings.TrimSpace(item.NotificationType)
		if item.TargetID != "" && item.TargetRole == "" {
			return platform.SaveNotificationPreferenceBatchRequest{}, nil, platform.ErrInvalidArgument
		}
		if item.NotificationType != "" && item.TargetRole != "" && item.TargetID == "" {
			return platform.SaveNotificationPreferenceBatchRequest{}, nil, platform.ErrInvalidArgument
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = updatedAt
		} else {
			item.UpdatedAt = item.UpdatedAt.UTC()
		}
		preferenceKey := adminNotificationPreferenceKey(item.TargetRole, item.TargetID, item.NotificationType)
		if seen[preferenceKey] {
			return platform.SaveNotificationPreferenceBatchRequest{}, nil, platform.ErrInvalidArgument
		}
		seen[preferenceKey] = true
		preferenceKeys = append(preferenceKeys, preferenceKey)
		normalized = append(normalized, item)
	}
	sort.Strings(preferenceKeys)
	return platform.SaveNotificationPreferenceBatchRequest{
		Preferences: normalized,
		Reason:      reason,
		UpdatedAt:   updatedAt,
	}, preferenceKeys, nil
}

func normalizeAdminNotificationPreferenceRolloutPolicy(policy adminNotificationPreferenceRolloutPolicy, preferences []platform.SaveNotificationPreferenceRequest) (adminNotificationPreferenceRolloutPolicy, error) {
	targetIDs := normalizeUniqueStrings(policy.TargetIDs)
	mode := strings.TrimSpace(policy.Mode)
	if mode == "" {
		switch {
		case policy.Percentage > 0:
			mode = adminNotificationPreferenceRolloutPercentage
		case len(targetIDs) > 0:
			mode = adminNotificationPreferenceRolloutTargetIDs
		default:
			mode = adminNotificationPreferenceRolloutAll
		}
	}
	if len(preferences) == 0 || policy.MaxTargets < 0 || policy.MaxTargets > 50 {
		return adminNotificationPreferenceRolloutPolicy{}, platform.ErrInvalidArgument
	}
	normalized := adminNotificationPreferenceRolloutPolicy{
		Mode:       mode,
		MaxTargets: policy.MaxTargets,
	}
	switch mode {
	case adminNotificationPreferenceRolloutAll:
	case adminNotificationPreferenceRolloutTargetIDs:
		if len(targetIDs) == 0 {
			return adminNotificationPreferenceRolloutPolicy{}, platform.ErrInvalidArgument
		}
		normalized.TargetIDs = targetIDs
	case adminNotificationPreferenceRolloutPercentage:
		if policy.Percentage <= 0 || policy.Percentage > 100 {
			return adminNotificationPreferenceRolloutPolicy{}, platform.ErrInvalidArgument
		}
		normalized.Percentage = policy.Percentage
	default:
		return adminNotificationPreferenceRolloutPolicy{}, platform.ErrInvalidArgument
	}
	return normalized, nil
}

func applyAdminNotificationPreferenceRollout(preferences []platform.SaveNotificationPreferenceRequest, policy adminNotificationPreferenceRolloutPolicy) ([]platform.SaveNotificationPreferenceRequest, []string, []string, error) {
	policy, err := normalizeAdminNotificationPreferenceRolloutPolicy(policy, preferences)
	if err != nil {
		return nil, nil, nil, err
	}
	targetIDs := map[string]bool{}
	for _, targetID := range policy.TargetIDs {
		targetIDs[targetID] = true
	}
	type rolloutEntry struct {
		preference platform.SaveNotificationPreferenceRequest
		key        string
		targetKey  string
		applied    bool
	}
	entries := make([]rolloutEntry, 0, len(preferences))
	for _, preference := range preferences {
		key := adminNotificationPreferenceKey(preference.TargetRole, preference.TargetID, preference.NotificationType)
		applied := false
		switch policy.Mode {
		case adminNotificationPreferenceRolloutAll:
			applied = true
		case adminNotificationPreferenceRolloutTargetIDs:
			applied = targetIDs[strings.TrimSpace(preference.TargetID)]
		case adminNotificationPreferenceRolloutPercentage:
			applied = adminNotificationPreferenceRolloutBucket(key) < policy.Percentage
		}
		entries = append(entries, rolloutEntry{
			preference: preference,
			key:        key,
			targetKey:  adminNotificationPreferenceRolloutTargetKey(preference),
			applied:    applied,
		})
	}
	if policy.MaxTargets > 0 {
		targetKeys := make([]string, 0, len(entries))
		seen := map[string]bool{}
		for _, entry := range entries {
			if !entry.applied || seen[entry.targetKey] {
				continue
			}
			seen[entry.targetKey] = true
			targetKeys = append(targetKeys, entry.targetKey)
		}
		sort.Strings(targetKeys)
		allowedTargets := map[string]bool{}
		for index, targetKey := range targetKeys {
			if index >= policy.MaxTargets {
				break
			}
			allowedTargets[targetKey] = true
		}
		for index := range entries {
			if entries[index].applied && !allowedTargets[entries[index].targetKey] {
				entries[index].applied = false
			}
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})
	appliedPreferences := make([]platform.SaveNotificationPreferenceRequest, 0, len(entries))
	appliedKeys := make([]string, 0, len(entries))
	skippedKeys := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.applied {
			appliedPreferences = append(appliedPreferences, entry.preference)
			appliedKeys = append(appliedKeys, entry.key)
			continue
		}
		skippedKeys = append(skippedKeys, entry.key)
	}
	if len(appliedPreferences) == 0 {
		return nil, nil, nil, platform.ErrInvalidArgument
	}
	return appliedPreferences, appliedKeys, skippedKeys, nil
}

func adminNotificationPreferenceRolloutBucket(key string) int {
	sum := sha256.Sum256([]byte(strings.TrimSpace(key)))
	return ((int(sum[0]) << 8) | int(sum[1])) % 100
}

func adminNotificationPreferenceRolloutTargetKey(preference platform.SaveNotificationPreferenceRequest) string {
	targetRole := strings.TrimSpace(preference.TargetRole)
	targetID := strings.TrimSpace(preference.TargetID)
	if targetID == "" {
		return adminNotificationPreferenceKey(targetRole, targetID, preference.NotificationType)
	}
	return targetRole + ":" + targetID
}

func normalizeUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized
}

func adminNotificationPreferenceKey(targetRole string, targetID string, notificationType string) string {
	targetRole = strings.TrimSpace(targetRole)
	targetID = strings.TrimSpace(targetID)
	notificationType = strings.TrimSpace(notificationType)
	if targetRole == "" && targetID == "" && notificationType == "" {
		return "default"
	}
	if targetRole == "" && notificationType != "" {
		return "type:" + notificationType
	}
	if targetID == "" {
		return targetRole
	}
	if notificationType == "" {
		return targetRole + ":" + targetID
	}
	return targetRole + ":" + targetID + ":" + notificationType
}

func isAdminNotificationPreferenceChangeStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case adminNotificationPreferenceChangeStatusPending, adminNotificationPreferenceChangeStatusApproved, adminNotificationPreferenceChangeStatusRejected, adminNotificationPreferenceChangeStatusApplied:
		return true
	default:
		return false
	}
}

func normalizeAdminNotificationPreferenceReviewDecision(decision string) (string, string, bool) {
	switch strings.TrimSpace(decision) {
	case adminNotificationPreferenceReviewApprove, adminNotificationPreferenceChangeStatusApproved:
		return adminNotificationPreferenceReviewApprove, adminNotificationPreferenceChangeStatusApproved, true
	case adminNotificationPreferenceReviewReject, adminNotificationPreferenceChangeStatusRejected:
		return adminNotificationPreferenceReviewReject, adminNotificationPreferenceChangeStatusRejected, true
	default:
		return "", "", false
	}
}

func (r *Router) adminNotificationPreferenceChangeRequestByID(changeRequestID string) (adminNotificationPreferenceChangeRequestRecord, error) {
	requests, err := r.adminNotificationPreferenceChangeRequestLedger()
	if err != nil {
		return adminNotificationPreferenceChangeRequestRecord{}, err
	}
	for _, item := range requests {
		if item.ID == changeRequestID {
			return item, nil
		}
	}
	return adminNotificationPreferenceChangeRequestRecord{}, platform.ErrNotFound
}

func (r *Router) adminNotificationPreferenceChangeRequestLedger() ([]adminNotificationPreferenceChangeRequestRecord, error) {
	requestLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.notification_preferences.change_requested",
		TargetType: "notification_preference_change_request",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	reviewLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.notification_preferences.change_reviewed",
		TargetType: "notification_preference_change_request",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	applyLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.notification_preferences.change_applied",
		TargetType: "notification_preference_batch",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	byID := map[string]*adminNotificationPreferenceChangeRequestRecord{}
	for _, log := range requestLogs {
		payload := log.Payload
		id := auditPayloadString(payload, "change_request_id")
		if id == "" {
			id = strings.TrimSpace(log.TargetID)
		}
		if id == "" {
			continue
		}
		preferences := auditPayloadNotificationPreferenceRequests(payload, "preference_requests")
		if len(preferences) == 0 {
			preferences = auditPayloadNotificationPreferenceRequests(payload, "preferences")
		}
		preferenceKeys := auditPayloadStringSlice(payload, "preference_keys")
		if len(preferenceKeys) == 0 {
			preferenceKeys = adminNotificationPreferenceKeysFromRequests(preferences)
		}
		rollout := auditPayloadNotificationPreferenceRolloutPolicy(payload, preferences)
		status := auditPayloadString(payload, "status")
		if status == "" {
			status = adminNotificationPreferenceChangeStatusPending
		}
		updatedAt := auditPayloadTime(payload, "updated_at")
		if updatedAt.IsZero() {
			updatedAt = log.CreatedAt
		}
		byID[id] = &adminNotificationPreferenceChangeRequestRecord{
			ID:               id,
			Status:           status,
			Reason:           auditPayloadString(payload, "reason"),
			Preferences:      preferences,
			PreferenceKeys:   preferenceKeys,
			Rollout:          rollout,
			RequestedCount:   auditPayloadInt(payload, "requested_count"),
			ApprovalRequired: true,
			AutoApplied:      false,
			RequestedByRole:  log.ActorType,
			RequestedByAdmin: log.ActorID,
			RequestedAt:      log.CreatedAt,
			RequestAuditID:   log.ID,
			UpdatedAt:        updatedAt,
		}
		if byID[id].RequestedCount == 0 {
			byID[id].RequestedCount = len(preferences)
		}
	}
	for _, log := range reviewLogs {
		id := auditPayloadString(log.Payload, "change_request_id")
		if id == "" {
			id = strings.TrimSpace(log.TargetID)
		}
		record := byID[id]
		if record == nil {
			continue
		}
		if record.ReviewedAt != nil && !log.CreatedAt.After(*record.ReviewedAt) {
			continue
		}
		status := auditPayloadString(log.Payload, "status")
		if !isAdminNotificationPreferenceChangeStatus(status) || status == adminNotificationPreferenceChangeStatusPending {
			decision := auditPayloadString(log.Payload, "decision")
			_, normalizedStatus, ok := normalizeAdminNotificationPreferenceReviewDecision(decision)
			if ok {
				status = normalizedStatus
			}
		}
		if status != adminNotificationPreferenceChangeStatusApproved && status != adminNotificationPreferenceChangeStatusRejected {
			continue
		}
		reviewedAt := log.CreatedAt
		record.Status = status
		record.ReviewDecision = auditPayloadString(log.Payload, "decision")
		record.ReviewReason = auditPayloadString(log.Payload, "reason")
		record.ReviewedByRole = log.ActorType
		record.ReviewedByAdmin = log.ActorID
		record.ReviewedAt = &reviewedAt
		record.ReviewAuditID = log.ID
	}
	for _, log := range applyLogs {
		id := auditPayloadString(log.Payload, "change_request_id")
		if id == "" {
			continue
		}
		record := byID[id]
		if record == nil {
			continue
		}
		if record.AppliedAt != nil && !log.CreatedAt.After(*record.AppliedAt) {
			continue
		}
		appliedAt := log.CreatedAt
		record.Status = adminNotificationPreferenceChangeStatusApplied
		record.Applied = true
		record.ApplyReason = auditPayloadString(log.Payload, "apply_reason")
		record.AppliedByRole = log.ActorType
		record.AppliedByAdmin = log.ActorID
		record.AppliedAt = &appliedAt
		record.ApplyAuditID = log.ID
		record.BatchID = auditPayloadString(log.Payload, "batch_id")
		if record.BatchID == "" {
			record.BatchID = strings.TrimSpace(log.TargetID)
		}
		record.BatchSaved = auditPayloadInt(log.Payload, "saved")
		record.AppliedKeys = auditPayloadStringSlice(log.Payload, "applied_preference_keys")
		if len(record.AppliedKeys) == 0 {
			record.AppliedKeys = auditPayloadStringSlice(log.Payload, "preference_keys")
		}
		record.SkippedKeys = auditPayloadStringSlice(log.Payload, "skipped_preference_keys")
		record.SkippedCount = auditPayloadInt(log.Payload, "skipped_count")
		if record.SkippedCount == 0 {
			record.SkippedCount = len(record.SkippedKeys)
		}
	}
	output := make([]adminNotificationPreferenceChangeRequestRecord, 0, len(byID))
	for _, item := range byID {
		output = append(output, *item)
	}
	sort.SliceStable(output, func(i, j int) bool {
		if !output[i].RequestedAt.Equal(output[j].RequestedAt) {
			return output[i].RequestedAt.After(output[j].RequestedAt)
		}
		return output[i].ID > output[j].ID
	})
	return output, nil
}

func adminNotificationPreferenceKeysFromRequests(preferences []platform.SaveNotificationPreferenceRequest) []string {
	keys := make([]string, 0, len(preferences))
	for _, preference := range preferences {
		keys = append(keys, adminNotificationPreferenceKey(preference.TargetRole, preference.TargetID, preference.NotificationType))
	}
	sort.Strings(keys)
	return keys
}

func auditPayloadNotificationPreferenceRequests(payload map[string]any, key string) []platform.SaveNotificationPreferenceRequest {
	value, ok := payload[key]
	if !ok || value == nil {
		return []platform.SaveNotificationPreferenceRequest{}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return []platform.SaveNotificationPreferenceRequest{}
	}
	var output []platform.SaveNotificationPreferenceRequest
	if err := json.Unmarshal(encoded, &output); err != nil {
		return []platform.SaveNotificationPreferenceRequest{}
	}
	return output
}

func auditPayloadNotificationPreferenceRolloutPolicy(payload map[string]any, preferences []platform.SaveNotificationPreferenceRequest) adminNotificationPreferenceRolloutPolicy {
	var policy adminNotificationPreferenceRolloutPolicy
	if value, ok := payload["rollout"]; ok && value != nil {
		encoded, err := json.Marshal(value)
		if err == nil {
			_ = json.Unmarshal(encoded, &policy)
		}
	}
	if strings.TrimSpace(policy.Mode) == "" {
		policy.Mode = auditPayloadString(payload, "rollout_mode")
	}
	if policy.Percentage == 0 {
		policy.Percentage = auditPayloadInt(payload, "rollout_percentage")
	}
	if len(policy.TargetIDs) == 0 {
		policy.TargetIDs = auditPayloadStringSlice(payload, "rollout_target_ids")
	}
	if policy.MaxTargets == 0 {
		policy.MaxTargets = auditPayloadInt(payload, "rollout_max_targets")
	}
	normalized, err := normalizeAdminNotificationPreferenceRolloutPolicy(policy, preferences)
	if err != nil {
		return adminNotificationPreferenceRolloutPolicy{Mode: adminNotificationPreferenceRolloutAll}
	}
	return normalized
}

func auditPayloadTime(payload map[string]any, key string) time.Time {
	text := auditPayloadString(payload, key)
	if text == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func auditPayloadInt(payload map[string]any, key string) int {
	text := auditPayloadString(payload, key)
	if text == "" {
		return 0
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}
	return value
}

func (r *Router) handleMerchantNotifications(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleMerchant && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	merchantID := principal.ID
	if principal.IsAdmin() {
		merchantID = req.URL.Query().Get("merchant_id")
	}
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	if strings.TrimSpace(merchantID) == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	notifications, err := r.store.Notifications(platform.NotificationListRequest{
		TargetRole: RoleMerchant,
		TargetID:   merchantID,
		Status:     req.URL.Query().Get("status"),
		Limit:      limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, notifications)
}

func (r *Router) handleMarkMerchantNotificationRead(w http.ResponseWriter, req *http.Request) {
	var payload platform.MarkNotificationReadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleMerchant && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	merchantID := principal.ID
	if principal.IsAdmin() {
		merchantID = strings.TrimSpace(payload.TargetID)
		if merchantID == "" {
			merchantID = req.URL.Query().Get("merchant_id")
		}
	}
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.NotificationID = req.PathValue("notificationID")
	payload.TargetRole = RoleMerchant
	payload.TargetID = merchantID
	notification, err := r.store.MarkNotificationRead(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, notification)
}

func (r *Router) handleMerchantNotificationPreferences(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleMerchant && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	merchantID := principal.ID
	if principal.IsAdmin() {
		merchantID = req.URL.Query().Get("merchant_id")
	}
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	if strings.TrimSpace(merchantID) == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	preferences, err := r.store.NotificationPreferences(platform.NotificationPreferenceListRequest{
		TargetRole:       RoleMerchant,
		TargetID:         merchantID,
		NotificationType: req.URL.Query().Get("notification_type"),
		Limit:            limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preferences)
}

func (r *Router) handleSaveMerchantNotificationPreference(w http.ResponseWriter, req *http.Request) {
	var payload platform.SaveNotificationPreferenceRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleMerchant && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	merchantID := principal.ID
	if principal.IsAdmin() {
		merchantID = strings.TrimSpace(payload.TargetID)
		if merchantID == "" {
			merchantID = req.URL.Query().Get("merchant_id")
		}
	}
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TargetRole = RoleMerchant
	payload.TargetID = merchantID
	preference, err := r.store.SaveNotificationPreference(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preference)
}

func (r *Router) handleMerchantQualification(w http.ResponseWriter, req *http.Request) {
	var payload platform.UploadMerchantQualificationRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.MerchantID) == "" && principal.Role == RoleMerchant {
		payload.MerchantID = principal.ID
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	profile, err := r.store.SaveMerchantQualification(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, profile)
}

func (r *Router) handleAdminMerchantQualifications(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewMerchantQualifications() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	now, ok := parseOptionalTimeQuery(w, query.Get("now"))
	if !ok {
		return
	}
	qualifications, err := r.store.AdminMerchantQualifications(platform.AdminMerchantQualificationListRequest{
		Status:     query.Get("status"),
		MerchantID: query.Get("merchant_id"),
		Type:       query.Get("type"),
		Limit:      limit,
		Now:        now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, qualifications)
}

func (r *Router) handleAdminMerchantQualificationDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewMerchantQualifications() {
		writeAuthError(w, errForbidden)
		return
	}
	now, ok := parseOptionalTimeQuery(w, req.URL.Query().Get("now"))
	if !ok {
		return
	}
	auditLimit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("audit_limit"))
	if !ok {
		return
	}
	detail, err := r.store.AdminMerchantQualificationDetail(platform.AdminMerchantQualificationDetailRequest{
		QualificationID: req.PathValue("qualificationID"),
		Now:             now,
		AuditLimit:      auditLimit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAdminReviewMerchantQualification(w http.ResponseWriter, req *http.Request) {
	var payload platform.ReviewMerchantQualificationRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewMerchantQualifications() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.QualificationID = req.PathValue("qualificationID")
	profile, qualification, audit, outboxEvent, err := r.store.ReviewMerchantQualificationWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.merchant_qualification.reviewed",
		TargetType: "merchant_qualification",
		TargetID:   payload.QualificationID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.ReviewedAt,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"profile":       profile,
		"qualification": qualification,
		"audit_log":     audit,
		"outbox_event":  outboxEvent,
	})
}

func (r *Router) handleMerchantStaff(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	staff, err := r.store.MerchantStaff(merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, staff)
}

func (r *Router) handleSaveMerchantStaff(w http.ResponseWriter, req *http.Request) {
	var payload platform.UpsertMerchantStaffRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.MerchantID) == "" && principal.Role == RoleMerchant {
		payload.MerchantID = principal.ID
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	staff, err := r.store.SaveMerchantStaff(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, staff)
}

func (r *Router) handleMerchantMaterials(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	materials, err := r.store.MerchantSupplementalMaterials(merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, materials)
}

func (r *Router) handleSaveMerchantMaterial(w http.ResponseWriter, req *http.Request) {
	var payload platform.UploadMerchantSupplementalMaterialRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.MerchantID) == "" && principal.Role == RoleMerchant {
		payload.MerchantID = principal.ID
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	material, err := r.store.SaveMerchantSupplementalMaterial(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, material)
}

func (r *Router) handleMerchantOrders(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	orders, err := r.store.MerchantOrders(merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, orders)
}

func (r *Router) handleMerchantAfterSales(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	requests, err := r.store.MerchantAfterSalesRequests(merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, requests)
}

func (r *Router) handleMerchantDeposit(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	deposit, err := r.store.DepositAccount("merchant", merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deposit)
}

func (r *Router) handlePayMerchantDeposit(w http.ResponseWriter, req *http.Request) {
	var payload platform.PayDepositRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.SubjectType = "merchant"
	payload.SubjectID = merchantID
	deposit, err := r.store.PayDeposit(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deposit)
}

func (r *Router) handleMerchantAcceptOrder(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, err := r.store.MerchantAcceptOrder(req.PathValue("orderID"), merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, order)
}

func (r *Router) handleMerchantOrderReady(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, err := r.store.MerchantMarkOrderReady(req.PathValue("orderID"), merchantID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, order)
}

func (r *Router) handleMerchantProducts(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if !principal.CanActAsMerchant(merchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	products, err := r.store.MerchantProducts(merchantID, req.URL.Query().Get("shop_id"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, products)
}

func (r *Router) handleUpsertMerchantProduct(w http.ResponseWriter, req *http.Request) {
	var payload platform.UpsertMerchantProductRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if strings.TrimSpace(payload.MerchantID) == "" {
		payload.MerchantID = merchantID
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	product, err := r.store.UpsertMerchantProduct(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, product)
}

func (r *Router) handleMerchantProductStatus(w http.ResponseWriter, req *http.Request) {
	var payload platform.SetMerchantProductStatusRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if strings.TrimSpace(payload.MerchantID) == "" {
		payload.MerchantID = merchantID
	}
	if strings.TrimSpace(payload.ProductID) == "" {
		payload.ProductID = req.PathValue("productID")
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	product, err := r.store.SetMerchantProductStatus(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, product)
}

func (r *Router) handleHomeModules(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, r.store.HomeModules())
}

func (r *Router) handleHomeCards(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, r.store.HomeCards())
}

func (r *Router) handleShops(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, r.store.Shops())
}

func (r *Router) handleShopDetail(w http.ResponseWriter, req *http.Request) {
	detail, err := r.store.ShopDetail(req.PathValue("shopID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleShopProducts(w http.ResponseWriter, req *http.Request) {
	products, err := r.store.ShopProducts(req.PathValue("shopID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, products)
}

func (r *Router) handleShopGroupbuyDeals(w http.ResponseWriter, req *http.Request) {
	deals, err := r.store.ShopGroupbuyDeals(req.PathValue("shopID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deals)
}

func (r *Router) handleUserAddresses(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if strings.TrimSpace(userID) == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	writeSuccess(w, r.store.UserAddresses(userID))
}

func (r *Router) handleSaveAddress(w http.ResponseWriter, req *http.Request) {
	var payload platform.UserAddress
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	address, err := r.store.SaveAddress(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, address)
}

func (r *Router) handleUserNotificationPreferences(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleUser && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if strings.TrimSpace(userID) == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	preferences, err := r.store.NotificationPreferences(platform.NotificationPreferenceListRequest{
		TargetRole:       RoleUser,
		TargetID:         userID,
		NotificationType: req.URL.Query().Get("notification_type"),
		Limit:            limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preferences)
}

func (r *Router) handleSaveUserNotificationPreference(w http.ResponseWriter, req *http.Request) {
	var payload platform.SaveNotificationPreferenceRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.Role != RoleUser && !principal.IsAdmin() {
		writeAuthError(w, errForbidden)
		return
	}
	userID := principal.ID
	if principal.IsAdmin() {
		userID = strings.TrimSpace(payload.TargetID)
		if userID == "" {
			userID = req.URL.Query().Get("user_id")
		}
	}
	if strings.TrimSpace(userID) == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TargetRole = RoleUser
	payload.TargetID = userID
	preference, err := r.store.SaveNotificationPreference(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preference)
}

func (r *Router) handleCartSummary(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.CartSummary(userID, req.URL.Query().Get("shop_id"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleUpsertCartItem(w http.ResponseWriter, req *http.Request) {
	var payload platform.UpsertCartItemRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.UpsertCartItem(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleCreateGroupbuyOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateGroupbuyOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, err := r.store.CreateGroupbuyOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, order)
}

func (r *Router) handleUserGroupbuyVouchers(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	vouchers, err := r.store.UserGroupbuyVouchers(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, vouchers)
}

func (r *Router) handleMerchantScanGroupbuyVoucher(w http.ResponseWriter, req *http.Request) {
	var payload platform.RedeemGroupbuyVoucherRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	merchantID := merchantIDFromPrincipal(req, principal)
	if strings.TrimSpace(payload.MerchantID) == "" {
		payload.MerchantID = merchantID
	}
	if !principal.CanActAsMerchant(payload.MerchantID) {
		writeAuthError(w, errForbidden)
		return
	}
	voucher, order, err := r.store.RedeemGroupbuyVoucher(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"voucher": voucher, "order": order})
}

func (r *Router) handleCreateOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, err := r.store.CreateOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, order)
}

func (r *Router) handleUserOrders(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	orders, err := r.store.UserOrders(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, orders)
}

func (r *Router) handleOrderDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	order, err := r.store.OrderByID(req.PathValue("orderID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if !principal.CanActAsUser(order.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	writeSuccess(w, order)
}

func (r *Router) handleUserReviews(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	reviews, err := r.store.UserReviews(platform.ReviewListRequest{
		UserID:  userID,
		OrderID: req.URL.Query().Get("order_id"),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, reviews)
}

func (r *Router) handleCreateReview(w http.ResponseWriter, req *http.Request) {
	var payload platform.Review
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	if strings.TrimSpace(payload.OrderID) != "" {
		order, err := r.store.OrderByID(payload.OrderID)
		if err != nil {
			writePlatformError(w, err)
			return
		}
		if !principal.CanActAsUser(order.UserID) {
			writeAuthError(w, errForbidden)
			return
		}
		if strings.TrimSpace(payload.TargetID) == "" {
			payload.TargetID = order.ID
		}
		if strings.TrimSpace(payload.TargetType) == "" {
			payload.TargetType = platform.ReviewTargetOrder
		}
	}
	review, err := r.store.CreateReview(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, review)
}

func (r *Router) handleCreateReviewImageUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateReviewImageUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, err := r.store.CreateReviewImageUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handleConfirmReviewImageUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.ConfirmReviewImageUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, err := r.store.ConfirmReviewImageUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handleUserAfterSales(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	requests, err := r.store.UserAfterSalesRequests(platform.AfterSalesListRequest{
		UserID:  userID,
		OrderID: req.URL.Query().Get("order_id"),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, requests)
}

func (r *Router) handleCreateAfterSales(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateAfterSalesRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	request, err := r.store.CreateAfterSales(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, request)
}

func (r *Router) handleUserFeedbackTickets(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	tickets, err := r.store.UserFeedbackTickets(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, tickets)
}

func (r *Router) handleCreateFeedback(w http.ResponseWriter, req *http.Request) {
	var payload platform.FeedbackTicket
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, err := r.store.CreateFeedback(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handleUserServiceTickets(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	tickets, err := r.store.UserServiceTickets(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, tickets)
}

func (r *Router) handleCreateServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.CreateServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, detail)
}

func (r *Router) handleServiceTicketDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.ServiceTicketDetail(userID, req.PathValue("ticketID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAddServiceTicketEvent(w http.ResponseWriter, req *http.Request) {
	var payload platform.AddServiceTicketEventRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.ActorID) == "" && principal.Role == RoleUser {
		payload.ActorID = principal.ID
	}
	if strings.TrimSpace(payload.ActorRole) == "" {
		payload.ActorRole = principal.Role
	}
	if !principal.CanActAsUser(payload.ActorID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.AddServiceTicketEvent(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, detail)
}

func (r *Router) handleCloseServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.CloseServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if strings.TrimSpace(payload.ActorID) == "" {
		payload.ActorID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.CloseServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleFollowUpServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.FollowUpServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.FollowUpServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAdminServiceTickets(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanReadServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	now, ok := parseOptionalTimeQuery(w, req.URL.Query().Get("now"))
	if !ok {
		return
	}
	tickets, err := r.store.AdminServiceTickets(platform.ServiceTicketListRequest{
		UserID:            req.URL.Query().Get("user_id"),
		RelatedOrderID:    req.URL.Query().Get("related_order_id"),
		Status:            req.URL.Query().Get("status"),
		SLAStatus:         req.URL.Query().Get("sla_status"),
		AssignedSupportID: req.URL.Query().Get("assigned_support_id"),
		Limit:             limit,
		Now:               now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, tickets)
}

func (r *Router) handleAdminServiceTicketDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanReadServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.AdminServiceTicketDetail(req.PathValue("ticketID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleServiceTicketQualityReviews(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanReadServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	reviews, err := r.store.ServiceTicketQualityReviews(platform.ServiceTicketQualityReviewListRequest{
		TicketID:         req.URL.Query().Get("ticket_id"),
		SupportID:        req.URL.Query().Get("support_id"),
		ReviewerID:       req.URL.Query().Get("reviewer_id"),
		Result:           req.URL.Query().Get("result"),
		CoachingRequired: req.URL.Query().Get("coaching_required"),
		Limit:            limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, reviews)
}

func (r *Router) handleServiceTicketPerformance(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanReadServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	now, ok := parseOptionalTimeQuery(w, req.URL.Query().Get("now"))
	if !ok {
		return
	}
	summaries, err := r.store.ServiceTicketPerformance(platform.ServiceTicketPerformanceRequest{
		SupportID: req.URL.Query().Get("support_id"),
		Limit:     limit,
		Now:       now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summaries)
}

func (r *Router) handleAssignServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.AssignServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanManageServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.ActorID) == "" {
		payload.ActorID = principal.ID
	}
	if strings.TrimSpace(payload.SupportID) == "" {
		payload.SupportID = principal.ID
	}
	detail, err := r.store.AssignServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleResolveServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.ResolveServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanManageServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.ActorID) == "" {
		payload.ActorID = principal.ID
	}
	detail, err := r.store.ResolveServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleEscalateServiceTicket(w http.ResponseWriter, req *http.Request) {
	var payload platform.EscalateServiceTicketRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanManageServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.ActorID) == "" {
		payload.ActorID = principal.ID
	}
	detail, err := r.store.EscalateServiceTicket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleReviewServiceTicketQuality(w http.ResponseWriter, req *http.Request) {
	var payload platform.ServiceTicketQualityReviewRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanManageServiceTickets(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.TicketID = req.PathValue("ticketID")
	if strings.TrimSpace(payload.ReviewerID) == "" {
		payload.ReviewerID = principal.ID
	}
	review, err := r.store.ReviewServiceTicketQuality(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, review)
}

func principalCanReadServiceTickets(principal Principal) bool {
	return principal.IsAdmin() || principal.Role == RoleOpsAdmin || principal.Role == RoleSupportAdmin
}

func principalCanReadAdminOrderDetail(principal Principal) bool {
	return principal.IsAdmin() || principal.Role == RoleOpsAdmin || principal.Role == RoleFinanceAdmin || principal.Role == RoleDispatchAdmin || principal.Role == RoleSupportAdmin
}

func principalCanManageServiceTickets(principal Principal) bool {
	return principal.IsAdmin() || principal.Role == RoleOpsAdmin || principal.Role == RoleSupportAdmin
}

func (r *Router) handleUserProfileOverview(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	overview, err := r.store.UserProfileOverview(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, overview)
}

func (r *Router) handleUserCoupons(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.UserCoupons(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleClaimUserCoupon(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	coupon, err := r.store.ClaimUserCoupon(payload.UserID, payload.Code)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, coupon)
}

func (r *Router) handleUserPointsSummary(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.UserPointsSummary(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleCheckInPoints(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.CheckInPoints(payload.UserID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleInviteSummary(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	summary, err := r.store.InviteSummary(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, summary)
}

func (r *Router) handleUserSearch(w http.ResponseWriter, req *http.Request) {
	catalog, err := r.store.SearchCatalog("", req.URL.Query().Get("keyword"), req.URL.Query().Get("category"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, catalog)
}

func (r *Router) handleMedicineHome(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	home, err := r.store.MedicineHome(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, home)
}

func (r *Router) handleCreatePrescriptionReview(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreatePrescriptionReviewRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	review, err := r.store.CreatePrescriptionReview(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, review)
}

func (r *Router) handleCreatePrescriptionImageUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreatePrescriptionImageUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, err := r.store.CreatePrescriptionImageUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handleConfirmPrescriptionImageUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.ConfirmPrescriptionImageUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, err := r.store.ConfirmPrescriptionImageUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handlePrescriptionReview(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	review, err := r.store.PrescriptionReview(userID, req.PathValue("reviewID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, review)
}

func (r *Router) handleAdminPrescriptionReviews(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadPrescriptions() {
		writeAuthError(w, errForbidden)
		return
	}
	limit := 20
	if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	reviews, err := r.store.AdminPrescriptionReviews(platform.PrescriptionReviewListRequest{
		Status:    req.URL.Query().Get("status"),
		ProductID: req.URL.Query().Get("product_id"),
		Limit:     limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, reviews)
}

func (r *Router) handleAdminReviewPrescription(w http.ResponseWriter, req *http.Request) {
	var payload platform.ReviewPrescriptionRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewPrescriptions() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.ReviewID = req.PathValue("reviewID")
	if strings.TrimSpace(payload.ReviewerID) == "" {
		payload.ReviewerID = principal.ID
	}
	review, err := r.store.ReviewPrescription(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, review)
}

func (r *Router) handleCreateMedicineOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.MedicineOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.CreateMedicineOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, detail)
}

func (r *Router) handleMedicineOrderDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.MedicineOrderDetail(userID, req.PathValue("orderID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleCreateErrandOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.ErrandOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.CreateErrandOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, detail)
}

func (r *Router) handleErrandOrderDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.ErrandOrderDetail(userID, req.PathValue("orderID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleCirclePosts(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	posts, err := r.store.CirclePosts(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, posts)
}

func (r *Router) handleCreateCirclePost(w http.ResponseWriter, req *http.Request) {
	var payload platform.CirclePost
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.AuthorUserID) == "" && principal.Role == RoleUser {
		payload.AuthorUserID = principal.ID
	}
	if !principal.CanActAsUser(payload.AuthorUserID) {
		writeAuthError(w, errForbidden)
		return
	}
	post, err := r.store.CreateCirclePost(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, post)
}

func (r *Router) handleMealMatchProfile(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	profile, err := r.store.UserMealMatchProfile(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	okToUse, missing := platform.CanUseMealMatch(*profile)
	writeSuccess(w, map[string]any{"profile": profile, "can_use": okToUse, "missing": missing})
}

func (r *Router) handleSaveMealMatchProfile(w http.ResponseWriter, req *http.Request) {
	var payload platform.MealMatchProfile
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	profile, err := r.store.SaveMealMatchProfile(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	okToUse, missing := platform.CanUseMealMatch(*profile)
	writeSuccess(w, map[string]any{"profile": profile, "can_use": okToUse, "missing": missing})
}

func (r *Router) handleMealMatchCandidates(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	result, err := r.store.MealMatchCandidates(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleReportMealMatchCandidate(w http.ResponseWriter, req *http.Request) {
	var payload platform.MealMatchReportRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.ReporterUserID) == "" && principal.Role == RoleUser {
		payload.ReporterUserID = principal.ID
	}
	if !principal.CanActAsUser(payload.ReporterUserID) {
		writeAuthError(w, errForbidden)
		return
	}
	record, err := r.store.ReportMealMatchCandidate(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, record)
}

func (r *Router) handleBlockMealMatchCandidate(w http.ResponseWriter, req *http.Request) {
	var payload platform.MealMatchBlockRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	record, err := r.store.BlockMealMatchCandidate(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, record)
}

func (r *Router) handleAdminMealMatchModerationRecords(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadMealMatchModeration() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	records, err := r.store.AdminMealMatchModerationRecords(platform.MealMatchModerationListRequest{
		Status:       query.Get("status"),
		Action:       query.Get("action"),
		UserID:       query.Get("user_id"),
		TargetUserID: query.Get("target_user_id"),
		Limit:        limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"records": records, "count": len(records)})
}

func (r *Router) handleAdminReviewMealMatchModeration(w http.ResponseWriter, req *http.Request) {
	var payload platform.MealMatchModerationReviewRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewMealMatchModeration() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.RecordID = req.PathValue("recordID")
	if strings.TrimSpace(payload.ReviewerID) == "" {
		payload.ReviewerID = principal.ID
	}
	record, err := r.store.ReviewMealMatchModeration(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, record)
}

func (r *Router) handleCreateRedPacket(w http.ResponseWriter, req *http.Request) {
	var payload platform.RedPacket
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.SenderID) == "" && principal.Role == RoleUser {
		payload.SenderID = principal.ID
	}
	if strings.TrimSpace(payload.SenderRole) == "" {
		payload.SenderRole = principal.Role
	}
	if !principal.CanActAsUser(payload.SenderID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.CreateRedPacket(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, detail)
}

func (r *Router) handleRedPacketDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.RedPacketDetail(req.PathValue("packetID"), userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleClaimRedPacket(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	result, err := r.store.ClaimRedPacket(req.PathValue("packetID"), userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleRefundRedPacket(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.RefundRedPacket(req.PathValue("packetID"), userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAutoRefundExpiredRedPackets(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRefunds() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload struct {
		Now string `json:"now"`
	}
	if req.Body != nil && req.ContentLength != 0 {
		if !decodeJSON(w, req, &payload) {
			return
		}
	}
	now := time.Now().UTC()
	if strings.TrimSpace(payload.Now) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.Now))
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		now = parsed.UTC()
	}
	details, err := r.store.AutoRefundExpiredRedPackets(now)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"refunded": details,
		"count":    len(details),
	})
}

func (r *Router) handleMessageThreads(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	threads, err := r.store.MessageThreads(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, threads)
}

func (r *Router) handleChatThreadOverview(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	overview, err := r.store.ChatThreadOverview(userID, req.PathValue("threadID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, overview)
}

func (r *Router) handleChatThreadMembers(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	members, err := r.store.ChatThreadMembers(userID, req.PathValue("threadID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"members": members,
		"count":   len(members),
	})
}

func (r *Router) handleChatThreadMembership(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	membership, err := r.store.ChatThreadMembership(userID, req.PathValue("threadID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, membership)
}

func (r *Router) handleJoinChatThread(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	membership, err := r.store.JoinChatThread(platform.ChatThreadJoinRequest{
		UserID:   userID,
		ThreadID: req.PathValue("threadID"),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, membership)
}

func (r *Router) handleLeaveChatThread(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	membership, err := r.store.LeaveChatThread(platform.ChatThreadLeaveRequest{
		UserID:   userID,
		ThreadID: req.PathValue("threadID"),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, membership)
}

func (r *Router) handleInternalRealtimeAuthorize(w http.ResponseWriter, req *http.Request) {
	if !r.authorizeRealtimeInternal(req) {
		writeAuthError(w, errUnauthorized)
		return
	}
	var payload platform.ChatThreadAccessRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	access, err := r.store.AuthorizeChatThreadAccess(payload)
	if errors.Is(err, platform.ErrNotFound) {
		writeSuccess(w, platform.ChatThreadAccessResult{
			ThreadID:    strings.TrimSpace(payload.ThreadID),
			SubjectType: strings.TrimSpace(payload.SubjectType),
			SubjectID:   strings.TrimSpace(payload.SubjectID),
			Allowed:     false,
			Reason:      "not_member",
		})
		return
	}
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, access)
}

func (r *Router) authorizeRealtimeInternal(req *http.Request) bool {
	expected := strings.TrimSpace(r.realtimeInternalToken)
	if expected == "" {
		return true
	}
	token, err := bearerToken(req)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(token), []byte(expected))
}

func (r *Router) handleChatMessages(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	messages, err := r.store.ChatMessages(userID, req.PathValue("threadID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, messages)
}

func (r *Router) handleChatMessageSync(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	markRead := req.URL.Query().Get("mark_read") != "false"
	result, err := r.store.ChatMessageSync(platform.ChatMessageSyncRequest{
		UserID:   userID,
		ThreadID: req.PathValue("threadID"),
		SinceID:  req.URL.Query().Get("since_id"),
		Limit:    limit,
		MarkRead: markRead,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleChatThreadPreference(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	preference, err := r.store.ChatThreadPreference(userID, req.PathValue("threadID"))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preference)
}

func (r *Router) handleUpdateChatThreadPreference(w http.ResponseWriter, req *http.Request) {
	var payload platform.UpdateChatThreadPreferenceRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	payload.ThreadID = req.PathValue("threadID")
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	preference, err := r.store.UpdateChatThreadPreference(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, preference)
}

func (r *Router) handleMarkChatThreadRead(w http.ResponseWriter, req *http.Request) {
	var payload platform.MarkChatThreadReadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	payload.ThreadID = req.PathValue("threadID")
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	read, err := r.store.MarkChatThreadRead(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, read)
}

func (r *Router) handleSendChatMessage(w http.ResponseWriter, req *http.Request) {
	var payload platform.ChatMessage
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	payload.ThreadID = req.PathValue("threadID")
	if strings.TrimSpace(payload.SenderID) == "" && principal.Role == RoleUser {
		payload.SenderID = principal.ID
	}
	if !principal.CanActAsUser(payload.SenderID) {
		writeAuthError(w, errForbidden)
		return
	}
	message, err := r.store.SendChatMessage(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, message)
}

func (r *Router) handleAdminCompensateOrderState(w http.ResponseWriter, req *http.Request) {
	var payload platform.CompensateOrderStateRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanCompensateOrders() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.OrderID = req.PathValue("orderID")
	payload.ActorID = principal.ID
	result, _, err := r.store.CompensateOrderStateWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.order_state.compensated",
		TargetType: "order",
		TargetID:   payload.OrderID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleAdminRefundSettings(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadRefundSettings() {
		writeAuthError(w, errForbidden)
		return
	}
	settings, err := r.store.RefundSettings()
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, settings)
}

func (r *Router) handleAdminSaveRefundSettings(w http.ResponseWriter, req *http.Request) {
	var payload platform.SaveRefundSettingsRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRefunds() {
		writeAuthError(w, errForbidden)
		return
	}
	settings, _, err := r.store.SaveRefundSettingsWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.refund_settings.updated",
		TargetType: "refund_settings",
		TargetID:   "default",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, settings)
}

func (r *Router) handleAdminRefundTransactions(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadRefundSettings() {
		writeAuthError(w, errForbidden)
		return
	}
	limit, ok := parseOptionalIntQuery(w, req.URL.Query().Get("limit"))
	if !ok {
		return
	}
	refunds, err := r.store.AdminRefundTransactions(platform.RefundTransactionListRequest{
		OrderID:     strings.TrimSpace(req.URL.Query().Get("order_id")),
		UserID:      strings.TrimSpace(req.URL.Query().Get("user_id")),
		Destination: strings.TrimSpace(req.URL.Query().Get("destination")),
		Status:      strings.TrimSpace(req.URL.Query().Get("status")),
		Limit:       limit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, refunds)
}

func (r *Router) handleAdminAfterSales(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAdminAfterSales() {
		writeAuthError(w, errForbidden)
		return
	}
	requests, err := r.store.AdminAfterSalesRequests(platform.AfterSalesListRequest{
		OrderID:   strings.TrimSpace(req.URL.Query().Get("order_id")),
		RequestID: strings.TrimSpace(req.URL.Query().Get("request_id")),
		Status:    strings.TrimSpace(req.URL.Query().Get("status")),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, requests)
}

func (r *Router) handleAdminOrderDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principalCanReadAdminOrderDetail(principal) {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.AdminOrderDetail(strings.TrimSpace(req.PathValue("orderID")))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAdminAfterSalesDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAdminAfterSales() {
		writeAuthError(w, errForbidden)
		return
	}
	detail, err := r.store.AdminAfterSalesDetail(strings.TrimSpace(req.PathValue("requestID")))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAdminOperationsSnapshot(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadOperationsSnapshot() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	now := time.Time{}
	if value := strings.TrimSpace(query.Get("now")); value != "" {
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		now = parsed
	}
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	leaseExpiringWithinSeconds, ok := parseOptionalIntQuery(w, query.Get("lease_expiring_within_seconds"))
	if !ok {
		return
	}
	objectCleanupGraceSeconds, ok := parseOptionalIntQuery(w, query.Get("object_cleanup_grace_seconds"))
	if !ok {
		return
	}
	snapshot, err := r.store.AdminOperationsSnapshot(platform.AdminOperationsSnapshotRequest{
		Now:                        now,
		Limit:                      limit,
		StationManagerID:           query.Get("station_manager_id"),
		LeaseExpiringWithinSeconds: leaseExpiringWithinSeconds,
		ObjectCleanupGraceSeconds:  objectCleanupGraceSeconds,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, snapshot)
}

func parseOptionalIntQuery(w http.ResponseWriter, value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		writePlatformError(w, platform.ErrInvalidArgument)
		return 0, false
	}
	return parsed, true
}

func parseOptionalTimeQuery(w http.ResponseWriter, value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		writePlatformError(w, platform.ErrInvalidArgument)
		return time.Time{}, false
	}
	return parsed, true
}

func (r *Router) handleAdminAuditLogs(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	after, ok := parseOptionalTimeQuery(w, query.Get("after"))
	if !ok {
		return
	}
	before, ok := parseOptionalTimeQuery(w, query.Get("before"))
	if !ok {
		return
	}
	logs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		ActorType:  query.Get("actor_type"),
		ActorID:    query.Get("actor_id"),
		Action:     query.Get("action"),
		TargetType: query.Get("target_type"),
		TargetID:   query.Get("target_id"),
		Limit:      limit,
		After:      after,
		Before:     before,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, logs)
}

func (r *Router) handleAdminAuditLogsExport(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	after, ok := parseOptionalTimeQuery(w, query.Get("after"))
	if !ok {
		return
	}
	before, ok := parseOptionalTimeQuery(w, query.Get("before"))
	if !ok {
		return
	}
	filter := platform.AuditLogsRequest{
		ActorType:  query.Get("actor_type"),
		ActorID:    query.Get("actor_id"),
		Action:     query.Get("action"),
		TargetType: query.Get("target_type"),
		TargetID:   query.Get("target_id"),
		Limit:      limit,
		After:      after,
		Before:     before,
	}
	logs, err := r.store.AuditLogs(filter)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	csvContent, err := buildAdminAuditLogCSV(logs)
	if err != nil {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	generatedAt := time.Now().UTC()
	filename := "audit-logs-" + generatedAt.Format("20060102T150405Z") + ".csv"
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.audit_logs.exported",
		TargetType: "audit_export",
		TargetID:   filename,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload:    adminAuditLogExportPayload(query, limit, after, before, len(logs), generatedAt),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"format":       "csv",
		"content_type": "text/csv; charset=utf-8",
		"filename":     filename,
		"generated_at": generatedAt.Format(time.RFC3339Nano),
		"row_count":    len(logs),
		"csv":          csvContent,
		"audit_log":    audit,
	})
}

func buildAdminAuditLogCSV(logs []platform.AuditLog) (string, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	header := []string{"id", "created_at", "actor_type", "actor_id", "action", "target_type", "target_id", "request_id", "ip_hash", "integrity_algorithm", "integrity_verified", "payload_json"}
	if err := writer.Write(header); err != nil {
		return "", err
	}
	for _, log := range logs {
		payload, err := json.Marshal(log.Payload)
		if err != nil {
			return "", err
		}
		row := []string{
			log.ID,
			log.CreatedAt.Format(time.RFC3339Nano),
			log.ActorType,
			log.ActorID,
			log.Action,
			log.TargetType,
			log.TargetID,
			log.RequestID,
			log.IPHash,
			log.IntegrityAlgorithm,
			strconv.FormatBool(log.IntegrityVerified),
			string(payload),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func adminAuditLogExportPayload(query map[string][]string, limit int, after time.Time, before time.Time, rowCount int, generatedAt time.Time) map[string]any {
	payload := map[string]any{
		"actor_id":      firstQueryValue(query, "actor_id"),
		"actor_type":    firstQueryValue(query, "actor_type"),
		"action_filter": firstQueryValue(query, "action"),
		"export_format": "csv",
		"generated_at":  generatedAt.Format(time.RFC3339Nano),
		"limit":         limit,
		"row_count":     rowCount,
		"target_id":     firstQueryValue(query, "target_id"),
		"target_type":   firstQueryValue(query, "target_type"),
	}
	if !after.IsZero() {
		payload["after"] = after.Format(time.RFC3339Nano)
	}
	if !before.IsZero() {
		payload["before"] = before.Format(time.RFC3339Nano)
	}
	return payload
}

func firstQueryValue(query map[string][]string, key string) string {
	values := query[key]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func (r *Router) handleAdminAuditRetentionReport(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	retentionDays, ok := parseOptionalIntQuery(w, query.Get("retention_days"))
	if !ok {
		return
	}
	hotDays, ok := parseOptionalIntQuery(w, query.Get("hot_days"))
	if !ok {
		return
	}
	integritySampleLimit, ok := parseOptionalIntQuery(w, query.Get("integrity_sample_limit"))
	if !ok {
		return
	}
	now, ok := parseOptionalTimeQuery(w, query.Get("now"))
	if !ok {
		return
	}
	if retentionDays < 0 || hotDays < 0 || integritySampleLimit < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	report, err := r.store.AuditRetentionReport(platform.AuditRetentionReportRequest{
		RetentionDays:        retentionDays,
		HotDays:              hotDays,
		IntegritySampleLimit: integritySampleLimit,
		Now:                  now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, report)
}

type adminAuditRetentionAlertPayload struct {
	RetentionDays        int       `json:"retention_days"`
	HotDays              int       `json:"hot_days"`
	IntegritySampleLimit int       `json:"integrity_sample_limit"`
	Now                  time.Time `json:"now"`
}

func (r *Router) handleAdminEmitAuditRetentionAlerts(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload adminAuditRetentionAlertPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.RetentionDays < 0 || payload.HotDays < 0 || payload.IntegritySampleLimit < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	emission, event, audit, err := r.store.EmitAuditRetentionAlerts(platform.AuditRetentionAlertEmissionRequest{
		RetentionDays:        payload.RetentionDays,
		HotDays:              payload.HotDays,
		IntegritySampleLimit: payload.IntegritySampleLimit,
		Now:                  payload.Now,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.audit_retention_alerts.emitted",
		TargetType: "audit_retention_alerts",
		TargetID:   "default",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"emission":     emission,
		"outbox_event": event,
		"audit_log":    audit,
	})
}

type adminAuditArchiveRequestPayload struct {
	HotDays       int       `json:"hot_days"`
	Limit         int       `json:"limit"`
	StoragePrefix string    `json:"storage_prefix"`
	Now           time.Time `json:"now"`
}

func (r *Router) handleAdminRequestAuditArchive(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload adminAuditArchiveRequestPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.HotDays < 0 || payload.Limit < 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	archive, event, audit, err := r.store.RequestAuditArchive(platform.AuditArchiveRequest{
		HotDays:       payload.HotDays,
		Limit:         payload.Limit,
		StoragePrefix: payload.StoragePrefix,
		Now:           payload.Now,
	}, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.audit_archive.requested",
		TargetType: "audit_archive",
		TargetID:   "pending",
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"archive":      archive,
		"outbox_event": event,
		"audit_log":    audit,
	})
}

func (r *Router) handleAdminAuditArchiveRecords(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	after, ok := parseOptionalTimeQuery(w, query.Get("after"))
	if !ok {
		return
	}
	before, ok := parseOptionalTimeQuery(w, query.Get("before"))
	if !ok {
		return
	}
	archives, err := r.store.AuditArchives(platform.AuditArchiveListRequest{
		ArchiveID: query.Get("archive_id"),
		Limit:     limit,
		After:     after,
		Before:    before,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, archives)
}

func (r *Router) handleAdminCompleteAuditArchive(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload platform.AuditArchiveCompletionRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	archive, audit, err := r.store.CompleteAuditArchive(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.audit_archive.completed",
		TargetType: "audit_archive",
		TargetID:   payload.ArchiveID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.UploadedAt,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"archive":   archive,
		"audit_log": audit,
	})
}

func (r *Router) handleAdminAuditArchiveVerifications(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	after, ok := parseOptionalTimeQuery(w, query.Get("after"))
	if !ok {
		return
	}
	before, ok := parseOptionalTimeQuery(w, query.Get("before"))
	if !ok {
		return
	}
	verifications, err := r.store.AuditArchiveVerifications(platform.AuditArchiveVerificationListRequest{
		ArchiveID: query.Get("archive_id"),
		Status:    query.Get("status"),
		Limit:     limit,
		After:     after,
		Before:    before,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, verifications)
}

func (r *Router) handleAdminVerifyAuditArchive(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAuditLogs() {
		writeAuthError(w, errForbidden)
		return
	}
	var payload platform.AuditArchiveVerifyRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	verification, audit, err := r.store.VerifyAuditArchive(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.audit_archive.verified",
		TargetType: "audit_archive",
		TargetID:   payload.ArchiveID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		CreatedAt:  payload.Now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{
		"verification": verification,
		"audit_log":    audit,
	})
}

func (r *Router) handleAdminRBACPolicy(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	writeSuccess(w, AdminRBACPolicyForPrincipal(principal))
}

type adminRBACChangeRequestPayload struct {
	Role            string   `json:"role"`
	RequestedScopes []string `json:"requested_scopes"`
	Reason          string   `json:"reason"`
}

type adminRBACChangeReviewPayload struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type adminRBACChangeApplyPayload struct {
	Reason string `json:"reason"`
}

type adminRBACChangeRollbackPayload struct {
	Reason string `json:"reason"`
}

type adminRBACChangeRequestRecord struct {
	ID                string     `json:"id"`
	Role              string     `json:"role"`
	CurrentScopes     []string   `json:"current_scopes"`
	RequestedScopes   []string   `json:"requested_scopes"`
	RequestReason     string     `json:"request_reason"`
	Status            string     `json:"status"`
	PolicyVersion     string     `json:"policy_version"`
	ApprovalRequired  bool       `json:"approval_required"`
	AutoApplied       bool       `json:"auto_applied"`
	RequestedByRole   string     `json:"requested_by_role"`
	RequestedByAdmin  string     `json:"requested_by_admin"`
	RequestedAt       time.Time  `json:"requested_at"`
	RequestAuditID    string     `json:"request_audit_id"`
	ReviewDecision    string     `json:"review_decision,omitempty"`
	ReviewReason      string     `json:"review_reason,omitempty"`
	ReviewedByRole    string     `json:"reviewed_by_role,omitempty"`
	ReviewedByAdmin   string     `json:"reviewed_by_admin,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewAuditID     string     `json:"review_audit_id,omitempty"`
	Applied           bool       `json:"applied"`
	AppliedScopes     []string   `json:"applied_scopes,omitempty"`
	PreviousScopes    []string   `json:"previous_scopes,omitempty"`
	AppliedByRole     string     `json:"applied_by_role,omitempty"`
	AppliedByAdmin    string     `json:"applied_by_admin,omitempty"`
	AppliedAt         *time.Time `json:"applied_at,omitempty"`
	ApplyAuditID      string     `json:"apply_audit_id,omitempty"`
	RolledBack        bool       `json:"rolled_back"`
	RollbackFrom      []string   `json:"rollback_from_scopes,omitempty"`
	RollbackTo        []string   `json:"rollback_to_scopes,omitempty"`
	RollbackReason    string     `json:"rollback_reason,omitempty"`
	RolledBackByRole  string     `json:"rolled_back_by_role,omitempty"`
	RolledBackByAdmin string     `json:"rolled_back_by_admin,omitempty"`
	RolledBackAt      *time.Time `json:"rolled_back_at,omitempty"`
	RollbackAuditID   string     `json:"rollback_audit_id,omitempty"`
}

const (
	adminRBACChangeStatusPending    = "pending_approval"
	adminRBACChangeStatusApproved   = "approved"
	adminRBACChangeStatusRejected   = "rejected"
	adminRBACChangeStatusApplied    = "applied"
	adminRBACChangeStatusRolledBack = "rolled_back"
	adminRBACReviewApprove          = "approve"
	adminRBACReviewReject           = "reject"
)

func (r *Router) handleAdminRBACChangeRequests(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	query := req.URL.Query()
	limit, ok := parseOptionalIntQuery(w, query.Get("limit"))
	if !ok {
		return
	}
	status := strings.TrimSpace(query.Get("status"))
	if status != "" && !isAdminRBACChangeStatus(status) {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	requests, err := r.adminRBACChangeRequestLedger()
	if err != nil {
		writePlatformError(w, err)
		return
	}
	filtered := make([]adminRBACChangeRequestRecord, 0, len(requests))
	counts := map[string]int{
		adminRBACChangeStatusPending:    0,
		adminRBACChangeStatusApproved:   0,
		adminRBACChangeStatusRejected:   0,
		adminRBACChangeStatusApplied:    0,
		adminRBACChangeStatusRolledBack: 0,
	}
	for _, item := range requests {
		if _, ok := counts[item.Status]; ok {
			counts[item.Status]++
		}
		if status == "" || item.Status == status {
			filtered = append(filtered, item)
		}
	}
	filteredTotal := len(filtered)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	writeSuccess(w, map[string]any{
		"items":             filtered,
		"total":             len(requests),
		"filtered_total":    filteredTotal,
		"pending_count":     counts[adminRBACChangeStatusPending],
		"approved_count":    counts[adminRBACChangeStatusApproved],
		"rejected_count":    counts[adminRBACChangeStatusRejected],
		"applied_count":     counts[adminRBACChangeStatusApplied],
		"rolled_back_count": counts[adminRBACChangeStatusRolledBack],
		"policy_version":    adminRBACPolicyVersion,
		"auto_apply":        false,
		"manual_apply":      true,
	})
}

func (r *Router) handleAdminRBACChangeRequest(w http.ResponseWriter, req *http.Request) {
	var payload adminRBACChangeRequestPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	role := strings.TrimSpace(payload.Role)
	if !IsBackofficeRoleName(role) {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	requestedScopes, valid := ValidateAdminRBACRoleScopes(role, payload.RequestedScopes)
	if !valid || len(requestedScopes) == 0 {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	changeRequestID := "rbac_change_" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	currentScopes := AdminScopesForRole(role)
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.rbac.change_requested",
		TargetType: "admin_rbac_role",
		TargetID:   role,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"change_request_id": changeRequestID,
			"current_scopes":    currentScopes,
			"policy_version":    adminRBACPolicyVersion,
			"reason":            reason,
			"requested_scopes":  requestedScopes,
			"role":              role,
			"status":            adminRBACChangeStatusPending,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{
		"id":                 changeRequestID,
		"role":               role,
		"current_scopes":     currentScopes,
		"requested_scopes":   requestedScopes,
		"reason":             reason,
		"status":             adminRBACChangeStatusPending,
		"policy_version":     adminRBACPolicyVersion,
		"approval_required":  true,
		"auto_applied":       false,
		"audit_log":          audit,
		"requested_by_role":  principal.Role,
		"requested_by_admin": principal.ID,
	})
}

func (r *Router) handleAdminRBACChangeRequestReview(w http.ResponseWriter, req *http.Request) {
	var payload adminRBACChangeReviewPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	changeRequestID := strings.TrimSpace(req.PathValue("changeRequestID"))
	if changeRequestID == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	decision, status, ok := normalizeAdminRBACReviewDecision(payload.Decision)
	if !ok {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	record, err := r.adminRBACChangeRequestByID(changeRequestID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if record.Status != adminRBACChangeStatusPending {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	if record.RequestedByAdmin == principal.ID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.rbac.change_reviewed",
		TargetType: "admin_rbac_change_request",
		TargetID:   changeRequestID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"change_request_id": changeRequestID,
			"current_scopes":    record.CurrentScopes,
			"decision":          decision,
			"policy_version":    adminRBACPolicyVersion,
			"reason":            reason,
			"requested_scopes":  record.RequestedScopes,
			"role":              record.Role,
			"status":            status,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	reviewedAt := audit.CreatedAt
	record.Status = status
	record.ReviewDecision = decision
	record.ReviewReason = reason
	record.ReviewedByRole = principal.Role
	record.ReviewedByAdmin = principal.ID
	record.ReviewedAt = &reviewedAt
	record.ReviewAuditID = audit.ID
	record.AutoApplied = false
	writeSuccess(w, map[string]any{
		"change_request": record,
		"audit_log":      audit,
		"auto_applied":   false,
	})
}

func (r *Router) handleAdminRBACChangeRequestApply(w http.ResponseWriter, req *http.Request) {
	var payload adminRBACChangeApplyPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	changeRequestID := strings.TrimSpace(req.PathValue("changeRequestID"))
	if changeRequestID == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	record, err := r.adminRBACChangeRequestByID(changeRequestID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if record.Status != adminRBACChangeStatusApproved {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	if record.RequestedByAdmin == principal.ID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	appliedScopes, valid := ValidateAdminRBACRoleScopes(record.Role, record.RequestedScopes)
	if !valid {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	previousScopes := AdminScopesForRole(record.Role)
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.rbac.change_applied",
		TargetType: "admin_rbac_role",
		TargetID:   record.Role,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"applied_scopes":    appliedScopes,
			"change_request_id": changeRequestID,
			"policy_version":    adminRBACPolicyVersion,
			"previous_scopes":   previousScopes,
			"reason":            reason,
			"role":              record.Role,
			"status":            adminRBACChangeStatusApplied,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if _, valid := ApplyAdminRBACRoleScopes(record.Role, appliedScopes); !valid {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	appliedAt := audit.CreatedAt
	record.Status = adminRBACChangeStatusApplied
	record.Applied = true
	record.AppliedScopes = appliedScopes
	record.PreviousScopes = previousScopes
	record.AppliedByRole = principal.Role
	record.AppliedByAdmin = principal.ID
	record.AppliedAt = &appliedAt
	record.ApplyAuditID = audit.ID
	record.AutoApplied = false
	writeSuccess(w, map[string]any{
		"change_request":  record,
		"audit_log":       audit,
		"previous_scopes": previousScopes,
		"applied_scopes":  appliedScopes,
		"auto_applied":    false,
		"runtime_applied": true,
	})
}

func (r *Router) handleAdminRBACChangeRequestRollback(w http.ResponseWriter, req *http.Request) {
	var payload adminRBACChangeRollbackPayload
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRBACPolicy() {
		writeAuthError(w, errForbidden)
		return
	}
	changeRequestID := strings.TrimSpace(req.PathValue("changeRequestID"))
	if changeRequestID == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	record, err := r.adminRBACChangeRequestByID(changeRequestID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if record.Status != adminRBACChangeStatusApplied || !record.Applied {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	if record.RequestedByAdmin == principal.ID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	rollbackTo := append([]string(nil), record.PreviousScopes...)
	if len(rollbackTo) == 0 {
		rollbackTo = append([]string(nil), record.CurrentScopes...)
	}
	rollbackTo, valid := ValidateAdminRBACRoleScopes(record.Role, rollbackTo)
	if !valid {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	latestAction, latestChangeRequestID, err := r.latestAdminRBACPolicyAuditForRole(record.Role)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if latestAction != "admin.rbac.change_applied" || latestChangeRequestID != changeRequestID {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	currentScopes := AdminScopesForRole(record.Role)
	if !sameAdminScopeList(currentScopes, record.AppliedScopes) {
		writePlatformError(w, platform.ErrInvalidOrderState)
		return
	}
	audit, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.rbac.change_rolled_back",
		TargetType: "admin_rbac_role",
		TargetID:   record.Role,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload: map[string]any{
			"applied_scopes":       record.AppliedScopes,
			"change_request_id":    changeRequestID,
			"policy_version":       adminRBACPolicyVersion,
			"previous_scopes":      rollbackTo,
			"reason":               reason,
			"role":                 record.Role,
			"rollback_from_scopes": currentScopes,
			"rollback_to_scopes":   rollbackTo,
			"status":               adminRBACChangeStatusRolledBack,
		},
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	if _, valid := ApplyAdminRBACRoleScopes(record.Role, rollbackTo); !valid {
		writePlatformError(w, platform.ErrInvalidArgument)
		return
	}
	rolledBackAt := audit.CreatedAt
	record.Status = adminRBACChangeStatusRolledBack
	record.RolledBack = true
	record.RollbackFrom = currentScopes
	record.RollbackTo = rollbackTo
	record.RollbackReason = reason
	record.RolledBackByRole = principal.Role
	record.RolledBackByAdmin = principal.ID
	record.RolledBackAt = &rolledBackAt
	record.RollbackAuditID = audit.ID
	record.AutoApplied = false
	writeSuccess(w, map[string]any{
		"change_request":       record,
		"audit_log":            audit,
		"rollback_from_scopes": currentScopes,
		"rollback_to_scopes":   rollbackTo,
		"auto_applied":         false,
		"runtime_applied":      true,
		"rolled_back":          true,
	})
}

func isAdminRBACChangeStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case adminRBACChangeStatusPending, adminRBACChangeStatusApproved, adminRBACChangeStatusRejected, adminRBACChangeStatusApplied, adminRBACChangeStatusRolledBack:
		return true
	default:
		return false
	}
}

func normalizeAdminRBACReviewDecision(decision string) (string, string, bool) {
	switch strings.TrimSpace(decision) {
	case adminRBACReviewApprove, adminRBACChangeStatusApproved:
		return adminRBACReviewApprove, adminRBACChangeStatusApproved, true
	case adminRBACReviewReject, adminRBACChangeStatusRejected:
		return adminRBACReviewReject, adminRBACChangeStatusRejected, true
	default:
		return "", "", false
	}
}

func (r *Router) adminRBACChangeRequestByID(changeRequestID string) (adminRBACChangeRequestRecord, error) {
	requests, err := r.adminRBACChangeRequestLedger()
	if err != nil {
		return adminRBACChangeRequestRecord{}, err
	}
	for _, item := range requests {
		if item.ID == changeRequestID {
			return item, nil
		}
	}
	return adminRBACChangeRequestRecord{}, platform.ErrNotFound
}

func (r *Router) adminRBACChangeRequestLedger() ([]adminRBACChangeRequestRecord, error) {
	requestLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_requested",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	reviewLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_reviewed",
		TargetType: "admin_rbac_change_request",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	applyLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_applied",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	rollbackLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_rolled_back",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return nil, err
	}
	byID := map[string]*adminRBACChangeRequestRecord{}
	for _, log := range requestLogs {
		payload := log.Payload
		id := auditPayloadString(payload, "change_request_id")
		if id == "" {
			continue
		}
		status := auditPayloadString(payload, "status")
		if status == "" {
			status = adminRBACChangeStatusPending
		}
		byID[id] = &adminRBACChangeRequestRecord{
			ID:               id,
			Role:             auditPayloadString(payload, "role"),
			CurrentScopes:    auditPayloadStringSlice(payload, "current_scopes"),
			RequestedScopes:  auditPayloadStringSlice(payload, "requested_scopes"),
			RequestReason:    auditPayloadString(payload, "reason"),
			Status:           status,
			PolicyVersion:    auditPayloadString(payload, "policy_version"),
			ApprovalRequired: true,
			AutoApplied:      false,
			RequestedByRole:  log.ActorType,
			RequestedByAdmin: log.ActorID,
			RequestedAt:      log.CreatedAt,
			RequestAuditID:   log.ID,
		}
	}
	for _, log := range reviewLogs {
		id := auditPayloadString(log.Payload, "change_request_id")
		if id == "" {
			id = strings.TrimSpace(log.TargetID)
		}
		record := byID[id]
		if record == nil {
			continue
		}
		if record.ReviewedAt != nil && !log.CreatedAt.After(*record.ReviewedAt) {
			continue
		}
		status := auditPayloadString(log.Payload, "status")
		if !isAdminRBACChangeStatus(status) || status == adminRBACChangeStatusPending {
			decision := auditPayloadString(log.Payload, "decision")
			_, normalizedStatus, ok := normalizeAdminRBACReviewDecision(decision)
			if ok {
				status = normalizedStatus
			}
		}
		if status != adminRBACChangeStatusApproved && status != adminRBACChangeStatusRejected {
			continue
		}
		reviewedAt := log.CreatedAt
		record.Status = status
		record.ReviewDecision = auditPayloadString(log.Payload, "decision")
		record.ReviewReason = auditPayloadString(log.Payload, "reason")
		record.ReviewedByRole = log.ActorType
		record.ReviewedByAdmin = log.ActorID
		record.ReviewedAt = &reviewedAt
		record.ReviewAuditID = log.ID
		record.AutoApplied = false
	}
	for _, log := range applyLogs {
		id := auditPayloadString(log.Payload, "change_request_id")
		if id == "" {
			continue
		}
		record := byID[id]
		if record == nil {
			continue
		}
		if record.AppliedAt != nil && !log.CreatedAt.After(*record.AppliedAt) {
			continue
		}
		appliedScopes := auditPayloadStringSlice(log.Payload, "applied_scopes")
		if len(appliedScopes) == 0 {
			appliedScopes = auditPayloadStringSlice(log.Payload, "requested_scopes")
		}
		if _, valid := ValidateAdminRBACRoleScopes(record.Role, appliedScopes); !valid {
			continue
		}
		appliedAt := log.CreatedAt
		previousScopes := auditPayloadStringSlice(log.Payload, "previous_scopes")
		if len(previousScopes) == 0 {
			previousScopes = append([]string(nil), record.CurrentScopes...)
		}
		record.Status = adminRBACChangeStatusApplied
		record.Applied = true
		record.AppliedScopes = appliedScopes
		record.PreviousScopes = previousScopes
		record.AppliedByRole = log.ActorType
		record.AppliedByAdmin = log.ActorID
		record.AppliedAt = &appliedAt
		record.ApplyAuditID = log.ID
		record.AutoApplied = false
	}
	for _, log := range rollbackLogs {
		id := auditPayloadString(log.Payload, "change_request_id")
		if id == "" {
			continue
		}
		record := byID[id]
		if record == nil {
			continue
		}
		if record.RolledBackAt != nil && !log.CreatedAt.After(*record.RolledBackAt) {
			continue
		}
		rollbackTo := auditPayloadStringSlice(log.Payload, "rollback_to_scopes")
		if len(rollbackTo) == 0 {
			rollbackTo = auditPayloadStringSlice(log.Payload, "previous_scopes")
		}
		if _, valid := ValidateAdminRBACRoleScopes(record.Role, rollbackTo); !valid {
			continue
		}
		rollbackFrom := auditPayloadStringSlice(log.Payload, "rollback_from_scopes")
		if len(rollbackFrom) == 0 {
			rollbackFrom = auditPayloadStringSlice(log.Payload, "applied_scopes")
		}
		rolledBackAt := log.CreatedAt
		record.Status = adminRBACChangeStatusRolledBack
		record.RolledBack = true
		record.RollbackFrom = rollbackFrom
		record.RollbackTo = rollbackTo
		record.RollbackReason = auditPayloadString(log.Payload, "reason")
		record.RolledBackByRole = log.ActorType
		record.RolledBackByAdmin = log.ActorID
		record.RolledBackAt = &rolledBackAt
		record.RollbackAuditID = log.ID
		record.AutoApplied = false
	}
	output := make([]adminRBACChangeRequestRecord, 0, len(byID))
	for _, item := range byID {
		output = append(output, *item)
	}
	sort.SliceStable(output, func(i, j int) bool {
		if !output[i].RequestedAt.Equal(output[j].RequestedAt) {
			return output[i].RequestedAt.After(output[j].RequestedAt)
		}
		return output[i].ID > output[j].ID
	})
	return output, nil
}

func (r *Router) restoreAdminRBACAppliedPolicyFromAudit() {
	resetAdminRBACRoleScopeOverrides()
	applyLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_applied",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return
	}
	rollbackLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_rolled_back",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return
	}
	type policyReplayEvent struct {
		id        string
		role      string
		scopes    []string
		createdAt time.Time
	}
	events := make([]policyReplayEvent, 0, len(applyLogs)+len(rollbackLogs))
	for _, log := range applyLogs {
		role := auditPayloadString(log.Payload, "role")
		if role == "" {
			role = strings.TrimSpace(log.TargetID)
		}
		appliedScopes := auditPayloadStringSlice(log.Payload, "applied_scopes")
		if len(appliedScopes) == 0 {
			appliedScopes = auditPayloadStringSlice(log.Payload, "requested_scopes")
		}
		events = append(events, policyReplayEvent{id: log.ID, role: role, scopes: appliedScopes, createdAt: log.CreatedAt})
	}
	for _, log := range rollbackLogs {
		role := auditPayloadString(log.Payload, "role")
		if role == "" {
			role = strings.TrimSpace(log.TargetID)
		}
		rollbackTo := auditPayloadStringSlice(log.Payload, "rollback_to_scopes")
		if len(rollbackTo) == 0 {
			rollbackTo = auditPayloadStringSlice(log.Payload, "previous_scopes")
		}
		events = append(events, policyReplayEvent{id: log.ID, role: role, scopes: rollbackTo, createdAt: log.CreatedAt})
	}
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].createdAt.Equal(events[j].createdAt) {
			return events[i].createdAt.Before(events[j].createdAt)
		}
		return events[i].id < events[j].id
	})
	for _, event := range events {
		ApplyAdminRBACRoleScopes(event.role, event.scopes)
	}
}

func (r *Router) latestAdminRBACPolicyAuditForRole(role string) (string, string, error) {
	role = strings.TrimSpace(role)
	applyLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_applied",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return "", "", err
	}
	rollbackLogs, err := r.store.AuditLogs(platform.AuditLogsRequest{
		Action:     "admin.rbac.change_rolled_back",
		TargetType: "admin_rbac_role",
		Limit:      500,
	})
	if err != nil {
		return "", "", err
	}
	var latest *platform.AuditLog
	for index := range applyLogs {
		log := &applyLogs[index]
		if !adminRBACPolicyAuditMatchesRole(*log, role) {
			continue
		}
		if latest == nil || adminRBACPolicyAuditAfter(*log, *latest) {
			latest = log
		}
	}
	for index := range rollbackLogs {
		log := &rollbackLogs[index]
		if !adminRBACPolicyAuditMatchesRole(*log, role) {
			continue
		}
		if latest == nil || adminRBACPolicyAuditAfter(*log, *latest) {
			latest = log
		}
	}
	if latest == nil {
		return "", "", platform.ErrNotFound
	}
	return latest.Action, auditPayloadString(latest.Payload, "change_request_id"), nil
}

func adminRBACPolicyAuditMatchesRole(log platform.AuditLog, role string) bool {
	logRole := auditPayloadString(log.Payload, "role")
	if logRole == "" {
		logRole = strings.TrimSpace(log.TargetID)
	}
	return logRole == role
}

func adminRBACPolicyAuditAfter(candidate platform.AuditLog, current platform.AuditLog) bool {
	if !candidate.CreatedAt.Equal(current.CreatedAt) {
		return candidate.CreatedAt.After(current.CreatedAt)
	}
	return candidate.ID > current.ID
}

func sameAdminScopeList(a []string, b []string) bool {
	normalizedA, validA := NormalizeAdminScopeList(a)
	normalizedB, validB := NormalizeAdminScopeList(b)
	if !validA || !validB || len(normalizedA) != len(normalizedB) {
		return false
	}
	for index := range normalizedA {
		if normalizedA[index] != normalizedB[index] {
			return false
		}
	}
	return true
}

func auditPayloadString(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(auditPayloadValueString(value))
}

func auditPayloadStringSlice(payload map[string]any, key string) []string {
	value, ok := payload[key]
	if !ok || value == nil {
		return []string{}
	}
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		output := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(auditPayloadValueString(item)); text != "" {
				output = append(output, text)
			}
		}
		return output
	case string:
		parts := strings.Split(typed, ",")
		output := make([]string, 0, len(parts))
		for _, part := range parts {
			if text := strings.TrimSpace(part); text != "" {
				output = append(output, text)
			}
		}
		return output
	default:
		return []string{}
	}
}

func auditPayloadValueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}

func (r *Router) recordAuditLog(req *http.Request, principal Principal, action string, targetType string, targetID string, payload map[string]any) error {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		targetID = "*"
	}
	_, err := r.store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
		Payload:    payload,
	})
	return err
}

func outboxTopicAuditTargetID(topic string) string {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "all"
	}
	return topic
}

func requestID(req *http.Request) string {
	for _, header := range []string{"X-Request-Id", "X-Request-ID", "X-Correlation-Id"} {
		if value := strings.TrimSpace(req.Header.Get(header)); value != "" {
			return value
		}
	}
	return ""
}

func (r *Router) handleAdminRefundOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.RefundOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageRefunds() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.OrderID = req.PathValue("orderID")
	payload.ActorID = principal.ID
	payload.ActorRole = principal.Role
	refund, order, account, _, err := r.store.RefundOrderWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.order.refunded",
		TargetType: "order",
		TargetID:   payload.OrderID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"refund": refund, "order": order, "wallet_account": account})
}

func (r *Router) handleReviewAfterSales(w http.ResponseWriter, req *http.Request) {
	var payload platform.ReviewAfterSalesRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReviewAfterSales() && principal.Role != RoleMerchant {
		writeAuthError(w, errForbidden)
		return
	}
	payload.RequestID = req.PathValue("requestID")
	payload.ActorID = principal.ID
	payload.ActorRole = principal.PlatformActorRole()
	request, refund, order, account, _, err := r.store.ReviewAfterSalesWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "after_sales.reviewed",
		TargetType: "after_sales",
		TargetID:   payload.RequestID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"after_sales": request, "refund": refund, "order": order, "wallet_account": account})
}

func (r *Router) handleAfterSalesEvents(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.IsBackofficeRole() && !principal.CanReadAdminAfterSales() {
		writeAuthError(w, errForbidden)
		return
	}
	events, err := r.store.AfterSalesEvents(req.PathValue("requestID"), principal.ID, principal.PlatformActorRole())
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, events)
}

func (r *Router) handleAddAfterSalesEvent(w http.ResponseWriter, req *http.Request) {
	var payload platform.AddAfterSalesEventRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.IsBackofficeRole() && !principal.CanAddAdminAfterSalesEvent() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.RequestID = req.PathValue("requestID")
	payload.ActorID = principal.ID
	payload.ActorRole = principal.PlatformActorRole()
	event, request, err := r.store.AddAfterSalesEvent(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{"event": event, "after_sales": request})
}

func (r *Router) handleCreateAfterSalesEvidenceUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreateAfterSalesEvidenceUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.IsBackofficeRole() && !principal.CanAddAdminAfterSalesEvent() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.RequestID = req.PathValue("requestID")
	payload.ActorID = principal.ID
	payload.ActorRole = principal.PlatformActorRole()
	ticket, err := r.store.CreateAfterSalesEvidenceUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, ticket)
}

func (r *Router) handleConfirmAfterSalesEvidenceUpload(w http.ResponseWriter, req *http.Request) {
	var payload platform.ConfirmAfterSalesEvidenceUploadRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.IsBackofficeRole() && !principal.CanAddAdminAfterSalesEvent() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.RequestID = req.PathValue("requestID")
	payload.ActorID = principal.ID
	payload.ActorRole = principal.PlatformActorRole()
	evidence, event, request, err := r.store.ConfirmAfterSalesEvidenceUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{"evidence": evidence, "event": event, "after_sales": request})
}

func (r *Router) handleObjectStorageUploadCallback(w http.ResponseWriter, req *http.Request) {
	var payload platform.ObjectStorageUploadCallbackRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.Signature == "" {
		payload.Signature = req.Header.Get("X-Object-Callback-Signature")
	}
	ticket, err := r.store.ConfirmObjectStorageUpload(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, ticket)
}

func (r *Router) handleObjectStorageScanResult(w http.ResponseWriter, req *http.Request) {
	var payload platform.ObjectStorageScanResultRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	if payload.Signature == "" {
		payload.Signature = req.Header.Get("X-Object-Callback-Signature")
	}
	ticket, err := r.store.RecordObjectStorageScanResult(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, ticket)
}

func (r *Router) handleAfterSalesEvidence(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if principal.IsBackofficeRole() && !principal.CanReadAdminAfterSales() {
		writeAuthError(w, errForbidden)
		return
	}
	evidence, err := r.store.AfterSalesEvidence(req.PathValue("requestID"), principal.ID, principal.PlatformActorRole())
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, evidence)
}

func (r *Router) handleAdminObjectStorageCleanupCandidates(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadObjectCleanup() {
		writeAuthError(w, errForbidden)
		return
	}
	cleanupReq, ok := parseObjectStorageCleanupQuery(w, req)
	if !ok {
		return
	}
	candidates, err := r.store.ObjectStorageCleanupCandidates(cleanupReq)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, candidates)
}

func (r *Router) handleAdminObjectStorageCleanupStats(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadObjectCleanup() {
		writeAuthError(w, errForbidden)
		return
	}
	cleanupReq, ok := parseObjectStorageCleanupQuery(w, req)
	if !ok {
		return
	}
	stats, err := r.store.ObjectStorageCleanupStats(cleanupReq)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, stats)
}

func parseObjectStorageCleanupQuery(w http.ResponseWriter, req *http.Request) (platform.ObjectStorageCleanupCandidatesRequest, bool) {
	limit := 0
	if value := strings.TrimSpace(req.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return platform.ObjectStorageCleanupCandidatesRequest{}, false
		}
		limit = parsed
	}
	graceSeconds := int64(0)
	if value := strings.TrimSpace(req.URL.Query().Get("grace_seconds")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return platform.ObjectStorageCleanupCandidatesRequest{}, false
		}
		graceSeconds = parsed
	}
	now := time.Time{}
	if value := strings.TrimSpace(req.URL.Query().Get("now")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return platform.ObjectStorageCleanupCandidatesRequest{}, false
		}
		now = parsed
	}
	return platform.ObjectStorageCleanupCandidatesRequest{
		Limit:        limit,
		GraceSeconds: graceSeconds,
		Now:          now,
	}, true
}

func (r *Router) handleAdminObjectStorageCleanupComplete(w http.ResponseWriter, req *http.Request) {
	var payload platform.ObjectStorageCleanupCompleteRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageObjectCleanup() {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, _, err := r.store.CompleteObjectStorageCleanupWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.object_cleanup.completed",
		TargetType: "object_storage_ticket",
		TargetID:   payload.TicketID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, ticket)
}

func (r *Router) handleAdminObjectStorageCleanupFailed(w http.ResponseWriter, req *http.Request) {
	var payload platform.ObjectStorageCleanupFailureRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageObjectCleanup() {
		writeAuthError(w, errForbidden)
		return
	}
	ticket, _, err := r.store.RecordObjectStorageCleanupFailureWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.object_cleanup.failed",
		TargetType: "object_storage_ticket",
		TargetID:   payload.TicketID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, ticket)
}

func (r *Router) handleAdminOutboxEvents(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	limit := 0
	if value := strings.TrimSpace(req.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		limit = parsed
	}
	now := time.Time{}
	if value := strings.TrimSpace(req.URL.Query().Get("now")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		now = parsed
	}
	events, err := r.store.OutboxEvents(platform.OutboxEventsRequest{
		Status: req.URL.Query().Get("status"),
		Topic:  req.URL.Query().Get("topic"),
		Limit:  limit,
		Now:    now,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, events)
}

func (r *Router) handleAdminOutboxEventDetail(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	now := time.Time{}
	if value := strings.TrimSpace(req.URL.Query().Get("now")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		now = parsed
	}
	auditLimit := 0
	if value := strings.TrimSpace(req.URL.Query().Get("audit_limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		auditLimit = parsed
	}
	detail, err := r.store.OutboxEventDetail(platform.OutboxEventDetailRequest{
		EventID:    req.PathValue("eventID"),
		Now:        now,
		AuditLimit: auditLimit,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, detail)
}

func (r *Router) handleAdminOutboxStats(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	now := time.Time{}
	if value := strings.TrimSpace(req.URL.Query().Get("now")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		now = parsed
	}
	leaseExpiringWithinSeconds := 0
	if value := strings.TrimSpace(req.URL.Query().Get("lease_expiring_within_seconds")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writePlatformError(w, platform.ErrInvalidArgument)
			return
		}
		leaseExpiringWithinSeconds = parsed
	}
	stats, err := r.store.OutboxStats(platform.OutboxStatsRequest{
		Topic:                      req.URL.Query().Get("topic"),
		Now:                        now,
		LeaseExpiringWithinSeconds: leaseExpiringWithinSeconds,
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, stats)
}

func (r *Router) handleAdminClaimOutboxEvents(w http.ResponseWriter, req *http.Request) {
	var payload platform.ClaimOutboxEventsRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	result, _, err := r.store.ClaimOutboxEventsWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.claimed",
		TargetType: "outbox_topic",
		TargetID:   outboxTopicAuditTargetID(payload.Topic),
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleAdminRenewOutboxEventLease(w http.ResponseWriter, req *http.Request) {
	var payload platform.RenewOutboxEventLeaseRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.EventID = req.PathValue("eventID")
	event, _, err := r.store.RenewOutboxEventLeaseWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.lease_renewed",
		TargetType: "outbox_event",
		TargetID:   payload.EventID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, event)
}

func (r *Router) handleAdminMarkOutboxEventPublished(w http.ResponseWriter, req *http.Request) {
	var payload platform.MarkOutboxEventPublishedRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.EventID = req.PathValue("eventID")
	event, _, err := r.store.MarkOutboxEventPublishedWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.published",
		TargetType: "outbox_event",
		TargetID:   payload.EventID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, event)
}

func (r *Router) handleAdminMarkOutboxEventFailed(w http.ResponseWriter, req *http.Request) {
	var payload platform.MarkOutboxEventFailedRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.EventID = req.PathValue("eventID")
	event, _, err := r.store.MarkOutboxEventFailedWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.failed",
		TargetType: "outbox_event",
		TargetID:   payload.EventID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, event)
}

func (r *Router) handleAdminReplayOutboxEvent(w http.ResponseWriter, req *http.Request) {
	var payload platform.ReplayOutboxEventRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	payload.EventID = req.PathValue("eventID")
	event, _, err := r.store.ReplayOutboxEventWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.replayed",
		TargetType: "outbox_event",
		TargetID:   payload.EventID,
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, event)
}

func (r *Router) handleAdminReplayOutboxEvents(w http.ResponseWriter, req *http.Request) {
	var payload platform.ReplayOutboxEventsRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageOutbox() {
		writeAuthError(w, errForbidden)
		return
	}
	result, _, err := r.store.ReplayOutboxEventsWithAudit(payload, platform.RecordAuditLogRequest{
		ActorType:  principal.Role,
		ActorID:    principal.ID,
		Action:     "admin.outbox.batch_replayed",
		TargetType: "outbox_topic",
		TargetID:   outboxTopicAuditTargetID(payload.Topic),
		RequestID:  requestID(req),
		IPHash:     requestIPHash(req),
	})
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, result)
}

func (r *Router) handleCheckoutCart(w http.ResponseWriter, req *http.Request) {
	var payload platform.CheckoutCartRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, summary, err := r.store.CheckoutCart(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{"order": order, "cart_summary": summary})
}

func (r *Router) handleCreditWallet(w http.ResponseWriter, req *http.Request) {
	var payload platform.CreditWalletRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	transaction, account, err := r.store.CreditWallet(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"transaction": transaction, "account": account})
}

func (r *Router) handleWalletTransactions(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	transactions, err := r.store.WalletTransactions(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, transactions)
}

func (r *Router) handleWalletOverview(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" && principal.Role == RoleUser {
		userID = principal.ID
	}
	if !principal.CanActAsUser(userID) {
		writeAuthError(w, errForbidden)
		return
	}
	overview, err := r.store.WalletOverview(userID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, overview)
}

func (r *Router) handleWalletWithdraw(w http.ResponseWriter, req *http.Request) {
	var payload platform.WalletWithdrawRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	withdraw, account, err := r.store.RequestWalletWithdraw(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccessStatus(w, http.StatusCreated, map[string]any{"withdraw": withdraw, "account": account})
}

func (r *Router) handleBalancePay(w http.ResponseWriter, req *http.Request) {
	var payload platform.BalancePayRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	transaction, order, account, err := r.store.PayOrderWithBalance(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"transaction": transaction, "order": order, "account": account})
}

func (r *Router) handleSetPaymentPassword(w http.ResponseWriter, req *http.Request) {
	var payload platform.SetWalletPaymentPasswordRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	state, err := r.store.SetWalletPaymentPassword(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, state)
}

func (r *Router) handleWechatPrepay(w http.ResponseWriter, req *http.Request) {
	var payload platform.WechatPrepayRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.UserID) == "" && principal.Role == RoleUser {
		payload.UserID = principal.ID
	}
	if !principal.CanActAsUser(payload.UserID) {
		writeAuthError(w, errForbidden)
		return
	}
	prepay, transaction, err := r.store.CreateWechatPrepay(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"prepay": prepay, "transaction": transaction})
}

func (r *Router) handleWechatCallback(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_BODY", "invalid request body")
		return
	}
	if err := r.wechatPay.Verify(req, body); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_WECHAT_PAY_SIGNATURE", "invalid wechat pay signature")
		return
	}
	var payload platform.WechatPaymentCallbackRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json body")
		return
	}
	transaction, order, err := r.store.ConfirmWechatPayment(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"transaction": transaction, "order": order})
}

func (r *Router) handleRiderDeposit(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	riderID := riderIDFromPrincipal(req, principal)
	if !principal.CanActAsRider(riderID) {
		writeAuthError(w, errForbidden)
		return
	}
	deposit, err := r.store.DepositAccount("rider", riderID)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deposit)
}

func (r *Router) handlePayRiderDeposit(w http.ResponseWriter, req *http.Request) {
	var payload platform.PayDepositRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	riderID := riderIDFromPrincipal(req, principal)
	if !principal.CanActAsRider(riderID) {
		writeAuthError(w, errForbidden)
		return
	}
	payload.SubjectType = "rider"
	payload.SubjectID = riderID
	deposit, err := r.store.PayDeposit(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, deposit)
}

func (r *Router) handleRiderWechatExemption(w http.ResponseWriter, req *http.Request) {
	var payload platform.RiderWechatExemptionRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" {
		payload.RiderID = riderIDFromPrincipal(req, principal)
	}
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	deposit, rider, err := r.store.ApproveRiderWechatExemption(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"deposit": deposit, "rider": rider})
}

func (r *Router) handleRiderDepositRefund(w http.ResponseWriter, req *http.Request) {
	var payload platform.RiderDepositRefundRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" {
		payload.RiderID = riderIDFromPrincipal(req, principal)
	}
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	deposit, rider, err := r.store.RequestRiderDepositRefund(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"deposit": deposit, "rider": rider})
}

func (r *Router) handleRiderOnline(w http.ResponseWriter, req *http.Request) {
	var payload platform.SetRiderOnlineStatusRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" && principal.Role == RoleRider {
		payload.RiderID = principal.ID
	}
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	rider, err := r.store.SetRiderOnlineStatus(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, rider)
}

func (r *Router) handleAutoAssignOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.AutoAssignOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	payload.OrderID = req.PathValue("orderID")
	order, decision, err := r.store.AutoAssignOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"order": order, "decision": decision})
}

func (r *Router) handleRejectRiderAssignment(w http.ResponseWriter, req *http.Request) {
	var payload platform.RejectRiderAssignmentRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" && principal.Role == RoleRider {
		payload.RiderID = principal.ID
	}
	payload.OrderID = req.PathValue("orderID")
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	order, decision, err := r.store.RejectRiderAssignment(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"order": order, "decision": decision})
}

func (r *Router) handleTimeoutReassignOrder(w http.ResponseWriter, req *http.Request) {
	var payload platform.TimeoutReassignOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	payload.OrderID = req.PathValue("orderID")
	if principal.Role == RoleStationManager {
		payload.StationManagerID = principal.ID
	} else if strings.TrimSpace(payload.StationManagerID) == "" {
		payload.StationManagerID = req.URL.Query().Get("station_manager_id")
	}
	order, decision, err := r.store.TimeoutReassignOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"order": order, "decision": decision})
}

func (r *Router) handleDispatchEvents(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	events, err := r.store.DispatchEvents(req.PathValue("orderID"), stationManagerIDFromPrincipal(req, principal))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, events)
}

func (r *Router) handleStationManagerRiders(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	riders, err := r.store.StationRiders(stationManagerIDFromPrincipal(req, principal))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, riders)
}

func (r *Router) handleStationManagerOrders(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	orders, err := r.store.StationOrders(stationManagerIDFromPrincipal(req, principal))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, orders)
}

func (r *Router) handleStationManagerManualAssign(w http.ResponseWriter, req *http.Request) {
	var payload platform.ManualAssignOrderRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	payload.OrderID = req.PathValue("orderID")
	if principal.Role == RoleStationManager {
		payload.StationManagerID = principal.ID
	} else if strings.TrimSpace(payload.StationManagerID) == "" {
		payload.StationManagerID = req.URL.Query().Get("station_manager_id")
	}
	order, decision, err := r.store.ManualAssignOrder(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"order": order, "decision": decision})
}

func (r *Router) handleStationManagerTaskConfig(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	config, err := r.store.StationTaskConfig(stationManagerIDFromPrincipal(req, principal))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, config)
}

func (r *Router) handleSaveStationManagerTaskConfig(w http.ResponseWriter, req *http.Request) {
	var payload platform.SaveStationTaskConfigRequest
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanManageDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	if principal.Role == RoleStationManager {
		payload.StationManagerID = principal.ID
	} else if strings.TrimSpace(payload.StationManagerID) == "" {
		payload.StationManagerID = req.URL.Query().Get("station_manager_id")
	}
	config, err := r.store.SaveStationTaskConfig(payload)
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, config)
}

func (r *Router) handleStationManagerRiderPerformance(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadDispatch() && principal.Role != RoleStationManager {
		writeAuthError(w, errForbidden)
		return
	}
	performance, err := r.store.StationRiderPerformance(stationManagerIDFromPrincipal(req, principal))
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, performance)
}

func (r *Router) handleFreeDispatchCancel(w http.ResponseWriter, req *http.Request) {
	var payload struct {
		RiderID string `json:"rider_id"`
	}
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" && principal.Role == RoleRider {
		payload.RiderID = principal.ID
	}
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	allowed, usedOn, err := r.store.ConsumeFreeDispatchCancel(payload.RiderID, time.Now())
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"allowed": allowed, "used_on": usedOn})
}

func (r *Router) handleRiderOrderAction(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/rider/orders/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
		return
	}

	var payload struct {
		RiderID string `json:"rider_id"`
	}
	if !decodeJSON(w, req, &payload) {
		return
	}
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if strings.TrimSpace(payload.RiderID) == "" && principal.Role == RoleRider {
		payload.RiderID = principal.ID
	}
	if !principal.CanActAsRider(payload.RiderID) {
		writeAuthError(w, errForbidden)
		return
	}
	var order *platform.Order
	var err error
	switch parts[1] {
	case "grab":
		order, err = r.store.GrabOrder(parts[0], payload.RiderID)
	case "pickup":
		order, err = r.store.RiderMarkOrderPickedUp(parts[0], payload.RiderID)
	case "delivered":
		order, err = r.store.RiderMarkOrderDelivered(parts[0], payload.RiderID)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
		return
	}
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, order)
}

func (r *Router) requirePrincipal(w http.ResponseWriter, req *http.Request) (Principal, bool) {
	principal, err := r.authVerifier.Verify(req)
	if err != nil {
		writeAuthError(w, err)
		return Principal{}, false
	}
	return principal, true
}

func (r *Router) issueAccessToken(req *http.Request, principal Principal, ttl time.Duration) (string, time.Time, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return "", time.Time{}, errUnauthorized
	}
	token, expiresAt, err := r.tokenSigner.IssueWithSession(principal, sessionID, ttl)
	if err != nil {
		return "", time.Time{}, err
	}
	if r.authSessions != nil {
		session := newAuthSession(req, sessionID, principal, token, expiresAt)
		if err := r.authSessions.Save(req.Context(), session); err != nil {
			return "", time.Time{}, errUnauthorized
		}
	}
	return token, expiresAt, nil
}

func merchantIDFromPrincipal(req *http.Request, principal Principal) string {
	if principal.IsAdmin() {
		return req.URL.Query().Get("merchant_id")
	}
	if principal.Role == RoleMerchant {
		return principal.ID
	}
	return ""
}

func stationManagerIDFromPrincipal(req *http.Request, principal Principal) string {
	if principal.IsAdmin() || principal.CanReadDispatch() || principal.CanManageDispatch() {
		return req.URL.Query().Get("station_manager_id")
	}
	if principal.Role == RoleStationManager {
		return principal.ID
	}
	return ""
}

func riderIDFromPrincipal(req *http.Request, principal Principal) string {
	if principal.IsAdmin() || principal.Role == RoleStationManager || principal.CanReadDispatch() || principal.CanManageDispatch() {
		return req.URL.Query().Get("rider_id")
	}
	if principal.Role == RoleRider {
		return principal.ID
	}
	return ""
}

func decodeJSON(w http.ResponseWriter, req *http.Request, target any) bool {
	defer req.Body.Close()
	if err := json.NewDecoder(req.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json body")
		return false
	}
	return true
}

func writePlatformError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, platform.ErrInvalidArgument):
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	case errors.Is(err, platform.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, platform.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, platform.ErrInsufficientBalance):
		writeError(w, http.StatusConflict, "INSUFFICIENT_BALANCE", err.Error())
	case errors.Is(err, platform.ErrInsufficientStock):
		writeError(w, http.StatusConflict, "INSUFFICIENT_STOCK", err.Error())
	case errors.Is(err, platform.ErrPaymentPassword):
		writeError(w, http.StatusConflict, "PAYMENT_PASSWORD_REQUIRED_OR_INVALID", err.Error())
	case errors.Is(err, platform.ErrRiskControlRejected):
		writeError(w, http.StatusTooManyRequests, "RISK_CONTROL_REJECTED", err.Error())
	case errors.Is(err, platform.ErrRateLimited):
		writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
	case errors.Is(err, platform.ErrOrderAlreadyAssigned):
		writeError(w, http.StatusConflict, "ORDER_ALREADY_ASSIGNED", err.Error())
	case errors.Is(err, platform.ErrInvalidOrderState):
		writeError(w, http.StatusConflict, "INVALID_ORDER_STATE", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errUnauthorized):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authorization required")
	case errors.Is(err, errForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "permission denied")
	default:
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authorization required")
	}
}

func writeWechatMiniLoginError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errWechatMiniLoginInvalidCode):
		writeError(w, http.StatusBadRequest, "INVALID_WECHAT_LOGIN_CODE", "invalid wechat login code")
	case errors.Is(err, errWechatMiniLoginUnauthorized):
		writeError(w, http.StatusUnauthorized, "WECHAT_LOGIN_REJECTED", "wechat login rejected")
	default:
		writeError(w, http.StatusServiceUnavailable, "WECHAT_LOGIN_UNAVAILABLE", "wechat login unavailable")
	}
}

func writeSuccess(w http.ResponseWriter, data any) {
	writeSuccessStatus(w, http.StatusOK, data)
}

func writeSuccessStatus(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "ok",
		"data":    data,
	})
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"code":    code,
		"message": message,
	})
}
