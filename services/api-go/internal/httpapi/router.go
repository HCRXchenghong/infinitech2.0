package httpapi

import (
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
	store             platform.Repository
	mux               *http.ServeMux
	authVerifier      AuthVerifier
	tokenSigner       TokenSigner
	authSessions      AuthSessionStore
	adminPasswordHash map[string]string
	allowDevAuth      bool
	wechatPay         WechatPayVerifier
	wechatMini        WechatMiniSessionResolver
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

func NewRouter(store platform.Repository, options ...RouterOption) http.Handler {
	tokenSigner := NewTokenSigner(os.Getenv("AUTH_TOKEN_SECRET"))
	authSessions := NewMemoryAuthSessionStore()
	router := &Router{
		store:             store,
		mux:               http.NewServeMux(),
		tokenSigner:       tokenSigner,
		authSessions:      authSessions,
		adminPasswordHash: map[string]string{},
		allowDevAuth:      true,
		wechatPay:         NewWechatPayVerifier(os.Getenv("WECHAT_PAY_CALLBACK_SECRET")),
		wechatMini:        DevWechatMiniSessionResolver{},
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
	r.mux.HandleFunc("POST /api/auth/wechat-mini/login", r.handleWechatMiniLogin)
	r.mux.HandleFunc("POST /api/auth/logout", r.handleLogout)
	r.mux.HandleFunc("POST /api/admin/merchant-invites", r.handleCreateMerchantInvite)
	r.mux.HandleFunc("POST /api/admin/rider-invites", r.handleCreateRiderInvite)
	r.mux.HandleFunc("GET /api/admin/refund-settings", r.handleAdminRefundSettings)
	r.mux.HandleFunc("PUT /api/admin/refund-settings", r.handleAdminSaveRefundSettings)
	r.mux.HandleFunc("GET /api/admin/after-sales", r.handleAdminAfterSales)
	r.mux.HandleFunc("GET /api/admin/operations/snapshot", r.handleAdminOperationsSnapshot)
	r.mux.HandleFunc("GET /api/admin/audit-logs", r.handleAdminAuditLogs)
	r.mux.HandleFunc("GET /api/admin/rbac/policy", r.handleAdminRBACPolicy)
	r.mux.HandleFunc("GET /api/admin/rbac/change-requests", r.handleAdminRBACChangeRequests)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests", r.handleAdminRBACChangeRequest)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/review", r.handleAdminRBACChangeRequestReview)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/apply", r.handleAdminRBACChangeRequestApply)
	r.mux.HandleFunc("POST /api/admin/rbac/change-requests/{changeRequestID}/rollback", r.handleAdminRBACChangeRequestRollback)
	r.mux.HandleFunc("GET /api/admin/object-storage/cleanup-candidates", r.handleAdminObjectStorageCleanupCandidates)
	r.mux.HandleFunc("GET /api/admin/object-storage/cleanup-stats", r.handleAdminObjectStorageCleanupStats)
	r.mux.HandleFunc("POST /api/admin/object-storage/cleanup-complete", r.handleAdminObjectStorageCleanupComplete)
	r.mux.HandleFunc("POST /api/admin/object-storage/cleanup-failed", r.handleAdminObjectStorageCleanupFailed)
	r.mux.HandleFunc("POST /api/admin/orders/{orderID}/state/compensate", r.handleAdminCompensateOrderState)
	r.mux.HandleFunc("GET /api/admin/outbox/events", r.handleAdminOutboxEvents)
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
	r.mux.HandleFunc("GET /api/merchant/me", r.handleMerchantMe)
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
	r.mux.HandleFunc("GET /api/shops/{shopID}/products", r.handleShopProducts)
	r.mux.HandleFunc("GET /api/shops/{shopID}/groupbuy-deals", r.handleShopGroupbuyDeals)
	r.mux.HandleFunc("GET /api/user/addresses", r.handleUserAddresses)
	r.mux.HandleFunc("POST /api/user/addresses", r.handleSaveAddress)
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
	r.mux.HandleFunc("GET /api/after-sales", r.handleUserAfterSales)
	r.mux.HandleFunc("POST /api/after-sales", r.handleCreateAfterSales)
	r.mux.HandleFunc("GET /api/after-sales/{requestID}/events", r.handleAfterSalesEvents)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/events", r.handleAddAfterSalesEvent)
	r.mux.HandleFunc("GET /api/after-sales/{requestID}/evidence", r.handleAfterSalesEvidence)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/evidence/upload-ticket", r.handleCreateAfterSalesEvidenceUpload)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/evidence/confirm", r.handleConfirmAfterSalesEvidenceUpload)
	r.mux.HandleFunc("POST /api/after-sales/{requestID}/review", r.handleReviewAfterSales)
	r.mux.HandleFunc("POST /api/wallet/credit", r.handleCreditWallet)
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
	requests, err := r.store.UserAfterSalesRequests(userID)
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

func (r *Router) handleAdminAfterSales(w http.ResponseWriter, req *http.Request) {
	principal, ok := r.requirePrincipal(w, req)
	if !ok {
		return
	}
	if !principal.CanReadAdminAfterSales() {
		writeAuthError(w, errForbidden)
		return
	}
	requests, err := r.store.AdminAfterSalesRequests()
	if err != nil {
		writePlatformError(w, err)
		return
	}
	writeSuccess(w, requests)
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
	case errors.Is(err, platform.ErrPaymentPassword):
		writeError(w, http.StatusConflict, "PAYMENT_PASSWORD_REQUIRED_OR_INVALID", err.Error())
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
