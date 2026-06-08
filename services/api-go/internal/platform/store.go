package platform

import (
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrNotFound             = errors.New("not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrOrderAlreadyAssigned = errors.New("order already assigned")
	ErrInvalidOrderState    = errors.New("invalid order state")
	ErrPaymentPassword      = errors.New("payment password required or invalid")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrRiskControlRejected  = errors.New("risk control rejected")
	ErrRateLimited          = errors.New("rate limited")
)

type Store struct {
	mu                          sync.Mutex
	nextOrderID                 uint64
	nextTransactionID           uint64
	nextAddressID               uint64
	nextMerchantID              uint64
	nextMerchantStaffID         uint64
	nextMerchantMaterialID      uint64
	nextDispatchEventID         uint64
	nextOutboxEventID           uint64
	nextAuditLogID              uint64
	nextNotificationID          uint64
	nextAfterSalesID            uint64
	nextAfterSalesEventID       uint64
	nextRiderID                 uint64
	nextProductID               uint64
	nextVoucherID               uint64
	nextReviewID                uint64
	nextFeedbackID              uint64
	nextCirclePostID            uint64
	nextRedPacketID             uint64
	nextChatMessageID           uint64
	nextWithdrawID              uint64
	nextCouponID                uint64
	nextServiceTicketID         uint64
	nextServiceTicketEventID    uint64
	nextServiceTicketQualityID  uint64
	nextPrescriptionID          uint64
	nextMealMatchModerationID   uint64
	homeModules                 []HomeModule
	homeCards                   []HomeCard
	users                       map[string]*AppUser
	wechatBindings              map[string]string
	phoneBindings               map[string]string
	phoneVerificationCodes      map[string]*PhoneVerificationCodeTicket
	phoneVerificationRequests   map[string][]time.Time
	phoneVerificationConfig     PhoneVerificationConfig
	userPasswordHash            map[string]string
	merchantInvites             map[string]*MerchantOnboardingInvite
	merchants                   map[string]*MerchantAccount
	merchantQualifications      map[string][]*MerchantQualification
	merchantStaff               map[string][]*MerchantStaff
	merchantMaterials           map[string][]*MerchantSupplementalMaterial
	riders                      map[string]*RiderAccount
	deposits                    map[string]*DepositAccount
	stationTaskConfigs          map[string]*StationTaskConfig
	stationServiceAreas         map[string]*StationServiceArea
	shops                       map[string]*Shop
	products                    map[string]*MerchantProduct
	groupbuyDeals               map[string]*MerchantProduct
	addresses                   map[string][]*UserAddress
	cartItems                   map[string][]*CartItem
	orders                      map[string]*Order
	reviews                     map[string]*Review
	wallets                     map[string]*WalletAccount
	paymentPasswordHash         map[string]string
	merchantPasswordHash        map[string]string
	riderPasswordHash           map[string]string
	paymentTransactions         map[string]*PaymentTransaction
	paymentByTradeNo            map[string]*PaymentTransaction
	paymentByProviderID         map[string]*PaymentTransaction
	walletIdempotency           map[string]*WalletTransaction
	refundSettings              RefundSettings
	refundTransactions          map[string]*RefundTransaction
	refundByIdempotency         map[string]string
	afterSalesRequests          map[string]*AfterSalesRequest
	afterSalesEvents            map[string]*AfterSalesEvent
	afterSalesUploadTickets     map[string]*AfterSalesEvidenceUploadTicket
	afterSalesEvidence          map[string]*AfterSalesEvidence
	groupbuyVouchers            map[string]*GroupbuyVoucher
	vouchersByOrderID           map[string][]string
	vouchersByCode              map[string]*GroupbuyVoucher
	dispatchEvents              map[string]*DispatchEvent
	dispatchRejectedRiders      map[string]map[string]bool
	freeCancelUsedByDate        map[string]string
	outboxEvents                map[string]*OutboxEvent
	outboxByIdempotency         map[string]string
	notifications               map[string]*PlatformNotification
	notificationByIdem          map[string]string
	notificationDeliveries      map[string]*PlatformNotificationDelivery
	notificationDeliveryByIdem  map[string]string
	notificationPreferences     map[string]*NotificationPreference
	notificationPreferenceByKey map[string]string
	feedbackTickets             map[string]*FeedbackTicket
	serviceTickets              map[string]*ServiceTicket
	serviceTicketEvents         map[string]*ServiceTicketEvent
	serviceTicketQualityReviews map[string]*ServiceTicketQualityReview
	circlePosts                 map[string]*CirclePost
	mealMatchProfiles           map[string]*MealMatchProfile
	mealMatchModeration         map[string]*MealMatchModerationRecord
	redPackets                  map[string]*RedPacketDetail
	chatThreadMembers           map[string]*ChatThreadMember
	chatMessages                map[string]*ChatMessage
	chatReadStates              map[string]*ChatReadState
	reviewImageTickets          map[string]*ReviewImageUploadTicket
	prescriptionImageTickets    map[string]*PrescriptionImageUploadTicket
	prescriptionReviews         map[string]*PrescriptionReview
	medicineDetails             map[string]*MedicineOrderDetail
	medicineStock               map[string]int
	withdrawRequests            map[string]*WalletWithdrawRequest
	userCoupons                 map[string]*UserCoupon
	pointsTransactions          map[string][]*PointsTransaction
	errandDetails               map[string]*ErrandOrderDetail
	auditLogs                   map[string]*AuditLog
	auditLogSigningSecret       string
	objectStorage               ObjectStorageConfig
}

func NewStore(homeModules []HomeModule) *Store {
	return &Store{
		nextMerchantID:              1,
		nextRiderID:                 2,
		nextProductID:               2,
		homeModules:                 cloneHomeModules(homeModules),
		homeCards:                   DefaultHomeCards(),
		users:                       map[string]*AppUser{},
		wechatBindings:              map[string]string{},
		phoneBindings:               map[string]string{},
		phoneVerificationCodes:      map[string]*PhoneVerificationCodeTicket{},
		phoneVerificationRequests:   map[string][]time.Time{},
		phoneVerificationConfig:     DefaultPhoneVerificationConfig(),
		userPasswordHash:            map[string]string{},
		merchantInvites:             map[string]*MerchantOnboardingInvite{},
		merchants:                   seedMerchants(),
		merchantQualifications:      seedMerchantQualifications(),
		merchantStaff:               seedMerchantStaff(),
		merchantMaterials:           seedMerchantMaterials(),
		riders:                      seedRiders(),
		deposits:                    seedDeposits(),
		stationTaskConfigs:          seedStationTaskConfigs(),
		stationServiceAreas:         seedStationServiceAreas(),
		shops:                       seedShops(),
		products:                    seedProducts(),
		groupbuyDeals:               seedGroupbuyDeals(),
		addresses:                   map[string][]*UserAddress{},
		cartItems:                   map[string][]*CartItem{},
		orders:                      map[string]*Order{},
		reviews:                     map[string]*Review{},
		wallets:                     map[string]*WalletAccount{},
		paymentPasswordHash:         map[string]string{},
		merchantPasswordHash:        map[string]string{},
		riderPasswordHash:           map[string]string{},
		paymentTransactions:         map[string]*PaymentTransaction{},
		paymentByTradeNo:            map[string]*PaymentTransaction{},
		paymentByProviderID:         map[string]*PaymentTransaction{},
		walletIdempotency:           map[string]*WalletTransaction{},
		refundSettings:              RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst},
		refundTransactions:          map[string]*RefundTransaction{},
		refundByIdempotency:         map[string]string{},
		afterSalesRequests:          map[string]*AfterSalesRequest{},
		afterSalesEvents:            map[string]*AfterSalesEvent{},
		afterSalesUploadTickets:     map[string]*AfterSalesEvidenceUploadTicket{},
		afterSalesEvidence:          map[string]*AfterSalesEvidence{},
		groupbuyVouchers:            map[string]*GroupbuyVoucher{},
		vouchersByOrderID:           map[string][]string{},
		vouchersByCode:              map[string]*GroupbuyVoucher{},
		dispatchEvents:              map[string]*DispatchEvent{},
		dispatchRejectedRiders:      map[string]map[string]bool{},
		freeCancelUsedByDate:        map[string]string{},
		outboxEvents:                map[string]*OutboxEvent{},
		outboxByIdempotency:         map[string]string{},
		notifications:               map[string]*PlatformNotification{},
		notificationByIdem:          map[string]string{},
		notificationDeliveries:      map[string]*PlatformNotificationDelivery{},
		notificationDeliveryByIdem:  map[string]string{},
		notificationPreferences:     map[string]*NotificationPreference{},
		notificationPreferenceByKey: map[string]string{},
		feedbackTickets:             map[string]*FeedbackTicket{},
		serviceTickets:              map[string]*ServiceTicket{},
		serviceTicketEvents:         map[string]*ServiceTicketEvent{},
		serviceTicketQualityReviews: map[string]*ServiceTicketQualityReview{},
		circlePosts:                 seedCirclePosts(),
		mealMatchProfiles:           seedMealMatchProfiles(),
		mealMatchModeration:         map[string]*MealMatchModerationRecord{},
		redPackets:                  map[string]*RedPacketDetail{},
		chatThreadMembers:           seedChatThreadMembers(),
		chatMessages:                seedChatMessages(),
		chatReadStates:              seedChatReadStates(),
		reviewImageTickets:          map[string]*ReviewImageUploadTicket{},
		prescriptionImageTickets:    map[string]*PrescriptionImageUploadTicket{},
		prescriptionReviews:         map[string]*PrescriptionReview{},
		medicineDetails:             map[string]*MedicineOrderDetail{},
		medicineStock:               defaultMedicineStock(),
		withdrawRequests:            map[string]*WalletWithdrawRequest{},
		userCoupons:                 seedUserCoupons(),
		pointsTransactions:          seedPointsTransactions(),
		errandDetails:               map[string]*ErrandOrderDetail{},
		auditLogs:                   map[string]*AuditLog{},
		objectStorage:               DefaultObjectStorageConfig(),
	}
}

func DefaultPhoneVerificationConfig() PhoneVerificationConfig {
	return PhoneVerificationConfig{
		Mode:            "dev",
		Provider:        "dev",
		Cooldown:        60 * time.Second,
		ExpiresIn:       10 * time.Minute,
		MaxPerPhoneHour: 5,
		MaxPerPhoneDay:  20,
		ReturnDevCode:   true,
	}
}

func normalizePhoneVerificationConfig(config PhoneVerificationConfig) PhoneVerificationConfig {
	defaults := DefaultPhoneVerificationConfig()
	config.Mode = strings.ToLower(strings.TrimSpace(config.Mode))
	if config.Mode == "" {
		config.Mode = defaults.Mode
	}
	if config.Mode != "provider" {
		config.Mode = "dev"
	}
	config.Provider = strings.TrimSpace(config.Provider)
	if config.Provider == "" {
		if config.Mode == "provider" {
			config.Provider = "sms_provider"
		} else {
			config.Provider = defaults.Provider
		}
	}
	config.TemplateID = strings.TrimSpace(config.TemplateID)
	if config.Cooldown <= 0 {
		config.Cooldown = defaults.Cooldown
	}
	if config.ExpiresIn <= 0 {
		config.ExpiresIn = defaults.ExpiresIn
	}
	if config.MaxPerPhoneHour <= 0 {
		config.MaxPerPhoneHour = defaults.MaxPerPhoneHour
	}
	if config.MaxPerPhoneDay <= 0 {
		config.MaxPerPhoneDay = defaults.MaxPerPhoneDay
	}
	if config.Mode == "provider" && !config.ReturnDevCode {
		config.ReturnDevCode = false
	} else if config.Mode != "provider" {
		config.ReturnDevCode = true
	}
	return config
}

func (s *Store) ConfigurePhoneVerification(config PhoneVerificationConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phoneVerificationConfig = normalizePhoneVerificationConfig(config)
	return nil
}

func (s *Store) ConfigureAuditLogIntegrity(signingSecret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogSigningSecret = strings.TrimSpace(signingSecret)
}

func (s *Store) auditLogSigningSecretSnapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.auditLogSigningSecret
}

func (s *Store) LoginWechatMini(req WechatMiniLoginRequest) (*WechatMiniLoginResult, error) {
	code := strings.TrimSpace(req.Code)
	providerOpenID := strings.TrimSpace(req.ProviderOpenID)
	if providerOpenID == "" && code == "" {
		return nil, ErrInvalidArgument
	}
	if providerOpenID == "" {
		providerOpenID = "wx_" + shortHash(code)
	}
	providerUnionID := strings.TrimSpace(req.ProviderUnionID)

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	userID := s.wechatBindings[providerOpenID]
	isNewUser := false
	if userID == "" {
		isNewUser = true
		userID = "user_" + shortHash(providerOpenID)
		s.wechatBindings[providerOpenID] = userID
		s.users[userID] = &AppUser{
			ID:        userID,
			Nickname:  strings.TrimSpace(req.Nickname),
			AvatarURL: strings.TrimSpace(req.AvatarURL),
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	user := s.users[userID]
	if user == nil {
		user = &AppUser{ID: userID, Status: "active", CreatedAt: now, UpdatedAt: now}
		s.users[userID] = user
	}
	if nickname := strings.TrimSpace(req.Nickname); nickname != "" {
		user.Nickname = nickname
		user.UpdatedAt = now
	}
	if avatarURL := strings.TrimSpace(req.AvatarURL); avatarURL != "" {
		user.AvatarURL = avatarURL
		user.UpdatedAt = now
	}
	return &WechatMiniLoginResult{
		User:            *cloneAppUser(user),
		Provider:        "wechat_mini",
		ProviderOpenID:  providerOpenID,
		ProviderUnionID: providerUnionID,
		IsNewUser:       isNewUser,
	}, nil
}

func (s *Store) SendPhoneVerificationCode(req SendPhoneVerificationCodeRequest) (*PhoneVerificationCodeTicket, error) {
	phone := normalizeMainlandPhone(req.Phone)
	if phone == "" {
		return nil, ErrInvalidArgument
	}
	purpose := normalizePhoneAuthPurpose(req.Purpose)
	now := time.Now().UTC()
	code := phoneVerificationCode(phone, purpose, now)
	requestID := "phone_code_" + shortHash(phone+":"+purpose+":"+now.Format(time.RFC3339Nano))

	s.mu.Lock()
	config := normalizePhoneVerificationConfig(s.phoneVerificationConfig)
	if err := s.recordPhoneVerificationAttemptLocked(phone, now, config); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	ticket := &PhoneVerificationCodeTicket{
		Phone:             phone,
		Purpose:           purpose,
		MaskedPhone:       maskMainlandPhone(phone),
		CreatedAt:         now,
		ExpiresAt:         now.Add(config.ExpiresIn),
		CooldownSeconds:   int64(config.Cooldown / time.Second),
		DeliveryProvider:  config.Provider,
		DeliveryStatus:    "pending",
		DeliveryRequestID: requestID,
		DevCode:           code,
	}
	s.phoneVerificationCodes[phone+"::"+purpose] = clonePhoneVerificationCodeTicket(ticket)
	s.mu.Unlock()

	if config.Mode == "provider" && config.Dispatcher != nil {
		result, err := config.Dispatcher.DispatchPhoneVerificationCode(PhoneVerificationDispatchRequest{
			Phone:       phone,
			MaskedPhone: ticket.MaskedPhone,
			Purpose:     purpose,
			Code:        code,
			TemplateID:  config.TemplateID,
			Provider:    config.Provider,
			RequestID:   requestID,
			ExpiresAt:   ticket.ExpiresAt,
		})
		if err != nil {
			s.updatePhoneVerificationDeliveryLocked(phone, purpose, "failed", config.Provider, requestID)
			return nil, err
		}
		if result != nil {
			provider := defaultString(result.Provider, config.Provider)
			nextRequestID := defaultString(result.RequestID, requestID)
			status := defaultString(result.Status, "delivered")
			s.updatePhoneVerificationDeliveryLocked(phone, purpose, status, provider, nextRequestID)
			ticket.DeliveryProvider = provider
			ticket.DeliveryRequestID = nextRequestID
			ticket.DeliveryStatus = status
		}
	} else if config.Mode == "provider" {
		s.updatePhoneVerificationDeliveryLocked(phone, purpose, "queued", config.Provider, requestID)
		ticket.DeliveryStatus = "queued"
	} else {
		s.updatePhoneVerificationDeliveryLocked(phone, purpose, "dev_returned", config.Provider, requestID)
		ticket.DeliveryStatus = "dev_returned"
	}
	return sanitizePhoneVerificationTicket(ticket, config), nil
}

func (s *Store) recordPhoneVerificationAttemptLocked(phone string, now time.Time, config PhoneVerificationConfig) error {
	history := s.phoneVerificationRequests[phone]
	fresh := make([]time.Time, 0, len(history)+1)
	for _, at := range history {
		if now.Sub(at) <= 24*time.Hour {
			fresh = append(fresh, at)
		}
	}
	if len(fresh) > 0 {
		last := fresh[len(fresh)-1]
		if now.Sub(last) < config.Cooldown {
			s.phoneVerificationRequests[phone] = fresh
			return ErrRateLimited
		}
	}
	hourCount := 0
	for _, at := range fresh {
		if now.Sub(at) <= time.Hour {
			hourCount++
		}
	}
	if hourCount >= config.MaxPerPhoneHour || len(fresh) >= config.MaxPerPhoneDay {
		s.phoneVerificationRequests[phone] = fresh
		return ErrRateLimited
	}
	fresh = append(fresh, now)
	s.phoneVerificationRequests[phone] = fresh
	return nil
}

func (s *Store) updatePhoneVerificationDeliveryLocked(phone string, purpose string, status string, provider string, requestID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.phoneVerificationCodes[phone+"::"+normalizePhoneAuthPurpose(purpose)]
	if ticket == nil {
		return
	}
	ticket.DeliveryStatus = strings.TrimSpace(status)
	ticket.DeliveryProvider = strings.TrimSpace(provider)
	ticket.DeliveryRequestID = strings.TrimSpace(requestID)
}

func (s *Store) LoginWithPhone(req PhoneLoginRequest) (*PhoneAuthResult, error) {
	phone := normalizeMainlandPhone(req.Phone)
	if phone == "" {
		return nil, ErrInvalidArgument
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		if strings.TrimSpace(req.Password) != "" {
			mode = "password"
		} else {
			mode = "code"
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	userID := s.phoneBindings[phone]
	switch mode {
	case "password":
		if userID == "" || !verifyUserPassword(s.userPasswordHash[userID], req.Password) {
			return nil, ErrInvalidCredentials
		}
	case "code":
		if !s.verifyPhoneCodeLocked(phone, "login", req.Code, now) && !s.verifyPhoneCodeLocked(phone, "register", req.Code, now) {
			return nil, ErrInvalidCredentials
		}
	default:
		return nil, ErrInvalidArgument
	}

	isNewUser := false
	if userID == "" {
		isNewUser = true
		userID = "user_phone_" + shortHash(phone)
		s.phoneBindings[phone] = userID
		s.users[userID] = &AppUser{
			ID:        userID,
			Nickname:  "手机用户" + phone[len(phone)-4:],
			Phone:     phone,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	user := s.userForPhoneAuthLocked(userID, phone, now)
	return &PhoneAuthResult{User: *cloneAppUser(user), Provider: "phone", IsNewUser: isNewUser}, nil
}

func (s *Store) RegisterWithPhone(req PhoneRegisterRequest) (*PhoneAuthResult, error) {
	phone := normalizeMainlandPhone(req.Phone)
	if phone == "" || !req.AcceptedAgreement {
		return nil, ErrInvalidArgument
	}
	passwordHash, err := hashUserPassword(req.Password)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if !s.verifyPhoneCodeLocked(phone, "register", req.Code, now) && !s.verifyPhoneCodeLocked(phone, "login", req.Code, now) {
		return nil, ErrInvalidCredentials
	}
	userID := s.phoneBindings[phone]
	isNewUser := false
	if userID == "" {
		isNewUser = true
		userID = "user_phone_" + shortHash(phone)
		s.phoneBindings[phone] = userID
		s.users[userID] = &AppUser{
			ID:        userID,
			Nickname:  defaultString(strings.TrimSpace(req.Nickname), "悦享用户"+phone[len(phone)-4:]),
			Phone:     phone,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	user := s.userForPhoneAuthLocked(userID, phone, now)
	if nickname := strings.TrimSpace(req.Nickname); nickname != "" {
		user.Nickname = nickname
		user.UpdatedAt = now
	}
	s.userPasswordHash[userID] = passwordHash
	return &PhoneAuthResult{User: *cloneAppUser(user), Provider: "phone", IsNewUser: isNewUser}, nil
}

func (s *Store) verifyPhoneCodeLocked(phone string, purpose string, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	ticket := s.phoneVerificationCodes[phone+"::"+normalizePhoneAuthPurpose(purpose)]
	if ticket == nil || now.After(ticket.ExpiresAt) {
		return false
	}
	return code == ticket.DevCode
}

func (s *Store) userForPhoneAuthLocked(userID string, phone string, now time.Time) *AppUser {
	user := s.users[userID]
	if user == nil {
		user = &AppUser{ID: userID, Status: "active", CreatedAt: now}
		s.users[userID] = user
	}
	if user.Status == "" {
		user.Status = "active"
	}
	if user.Phone != phone {
		user.Phone = phone
		user.UpdatedAt = now
	}
	if user.Nickname == "" {
		user.Nickname = "手机用户" + phone[len(phone)-4:]
		user.UpdatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}
	return user
}

func normalizeCreateMerchantInviteRequest(req CreateMerchantInviteRequest) (CreateMerchantInviteRequest, error) {
	adminID := strings.TrimSpace(req.AdminID)
	inviteType := strings.TrimSpace(req.Type)
	if adminID == "" {
		return req, ErrInvalidArgument
	}
	if inviteType == "" {
		inviteType = OnboardingInviteMerchant
	}
	if inviteType != OnboardingInviteMerchant {
		return req, ErrInvalidArgument
	}
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)
	}
	if !expiresAt.After(time.Now().UTC()) {
		return req, ErrInvalidArgument
	}
	req.AdminID = adminID
	req.Type = inviteType
	req.ExpiresAt = expiresAt.UTC()
	return req, nil
}

func (s *Store) createMerchantInviteAlreadyLocked(req CreateMerchantInviteRequest) *MerchantOnboardingInvite {
	token := "mi_" + shortHash(fmt.Sprintf("%s:%s:%d", req.AdminID, req.Type, len(s.merchantInvites)+1))
	invite := &MerchantOnboardingInvite{
		Token:                token,
		Type:                 req.Type,
		Status:               OnboardingInviteActive,
		CreatedByAdminID:     req.AdminID,
		CreatedBySubjectType: "admin",
		CreatedBySubjectID:   req.AdminID,
		ExpiresAt:            req.ExpiresAt.UTC(),
	}
	s.merchantInvites[token] = invite
	return invite
}

func (s *Store) CreateMerchantInvite(req CreateMerchantInviteRequest) (*MerchantOnboardingInvite, error) {
	normalized, err := normalizeCreateMerchantInviteRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.createMerchantInviteAlreadyLocked(normalized)
	return cloneMerchantInvite(invite), nil
}

func (s *Store) CreateMerchantInviteWithAudit(req CreateMerchantInviteRequest, audit RecordAuditLogRequest) (*MerchantOnboardingInvite, *AuditLog, error) {
	normalized, err := normalizeCreateMerchantInviteRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := inviteAuditLogFromRequest(audit, "admin.merchant_invite.created", "merchant_invite")
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.createMerchantInviteAlreadyLocked(normalized)
	log.TargetID = invite.Token
	log.Payload = merchantInviteAuditPayload(invite)
	auditLog := s.appendAuditLogLocked(log)
	return cloneMerchantInvite(invite), auditLog, nil
}

func (s *Store) AcceptMerchantInvite(req AcceptMerchantInviteRequest) (*MerchantProfile, error) {
	token := strings.TrimSpace(req.Token)
	displayName := strings.TrimSpace(req.DisplayName)
	accountType := strings.TrimSpace(req.AccountType)
	if accountType == "" {
		accountType = MerchantAccountStandard
	}
	passwordHash, err := hashLoginPassword(req.Password)
	if token == "" || displayName == "" || !isMerchantAccountType(accountType) || err != nil {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.merchantInvites[token]
	if invite == nil {
		return nil, ErrNotFound
	}
	if invite.Status != OnboardingInviteActive || !invite.ExpiresAt.After(time.Now().UTC()) {
		return nil, ErrInvalidArgument
	}
	s.nextMerchantID++
	merchantID := fmt.Sprintf("merchant_%d", s.nextMerchantID)
	account := &MerchantAccount{
		ID:            merchantID,
		Type:          accountType,
		DisplayName:   displayName,
		Status:        "pending_qualification",
		DepositStatus: DepositStatusUnpaid,
	}
	s.merchants[merchantID] = account
	s.merchantPasswordHash[merchantID] = passwordHash
	invite.Status = OnboardingInviteUsed
	return s.merchantProfileLocked(merchantID), nil
}

func (s *Store) LoginMerchant(req MerchantLoginRequest) (*MerchantProfile, error) {
	accountID := strings.TrimSpace(req.AccountID)
	password := strings.TrimSpace(req.Password)
	if accountID == "" || password == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	account := cloneMerchantAccount(s.merchants[accountID])
	passwordHash := s.merchantPasswordHash[accountID]
	var profile *MerchantProfile
	if account != nil && account.Status != "" && passwordHash != "" {
		profile = s.merchantProfileLocked(accountID)
	}
	s.mu.Unlock()

	if account == nil || account.Status == "" || passwordHash == "" {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return profile, nil
}

func normalizeCreateRiderInviteRequest(req CreateRiderInviteRequest) (CreateRiderInviteRequest, error) {
	createdByID := strings.TrimSpace(req.CreatedByID)
	createdByRole := strings.TrimSpace(req.CreatedByRole)
	inviteType := strings.TrimSpace(req.Type)
	stationID := strings.TrimSpace(req.StationID)
	if createdByID == "" || createdByRole == "" || stationID == "" {
		return req, ErrInvalidArgument
	}
	if inviteType != RiderAccountRider && inviteType != RiderAccountStationManager {
		return req, ErrInvalidArgument
	}
	if createdByRole == RiderAccountStationManager && inviteType != RiderAccountRider {
		return req, ErrInvalidArgument
	}
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)
	}
	if !expiresAt.After(time.Now().UTC()) {
		return req, ErrInvalidArgument
	}
	req.CreatedByID = createdByID
	req.CreatedByRole = createdByRole
	req.Type = inviteType
	req.StationID = stationID
	req.ExpiresAt = expiresAt.UTC()
	return req, nil
}

func (s *Store) createRiderInviteAlreadyLocked(req CreateRiderInviteRequest) (*MerchantOnboardingInvite, error) {
	if req.CreatedByRole == RiderAccountStationManager {
		manager := s.riders[req.CreatedByID]
		if manager == nil || manager.Type != RiderAccountStationManager || manager.Status != "active" || manager.StationID != req.StationID {
			return nil, ErrInvalidArgument
		}
	}
	token := "ri_" + shortHash(fmt.Sprintf("%s:%s:%s:%d", req.CreatedByID, req.Type, req.StationID, len(s.merchantInvites)+1))
	invite := &MerchantOnboardingInvite{
		Token:                token,
		Type:                 req.Type,
		Status:               OnboardingInviteActive,
		CreatedByAdminID:     req.CreatedByID,
		CreatedBySubjectType: req.CreatedByRole,
		CreatedBySubjectID:   req.CreatedByID,
		StationID:            req.StationID,
		ExpiresAt:            req.ExpiresAt.UTC(),
	}
	s.merchantInvites[token] = invite
	return invite, nil
}

func (s *Store) CreateRiderInvite(req CreateRiderInviteRequest) (*MerchantOnboardingInvite, error) {
	normalized, err := normalizeCreateRiderInviteRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, err := s.createRiderInviteAlreadyLocked(normalized)
	if err != nil {
		return nil, err
	}
	return cloneMerchantInvite(invite), nil
}

func (s *Store) CreateRiderInviteWithAudit(req CreateRiderInviteRequest, audit RecordAuditLogRequest) (*MerchantOnboardingInvite, *AuditLog, error) {
	normalized, err := normalizeCreateRiderInviteRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := inviteAuditLogFromRequest(audit, "admin.rider_invite.created", "rider_invite")
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	invite, err := s.createRiderInviteAlreadyLocked(normalized)
	if err != nil {
		return nil, nil, err
	}
	log.TargetID = invite.Token
	log.Payload = riderInviteAuditPayload(invite)
	auditLog := s.appendAuditLogLocked(log)
	return cloneMerchantInvite(invite), auditLog, nil
}

func inviteAuditLogFromRequest(req RecordAuditLogRequest, action string, targetType string) (*AuditLog, error) {
	log, err := auditLogFromRequest(req, "")
	if err != nil {
		return nil, err
	}
	if log.Action != action || log.TargetType != targetType {
		return nil, ErrInvalidArgument
	}
	return log, nil
}

func merchantInviteAuditPayload(invite *MerchantOnboardingInvite) map[string]any {
	if invite == nil {
		return map[string]any{}
	}
	return map[string]any{
		"type":       strings.TrimSpace(invite.Type),
		"expires_at": invite.ExpiresAt,
	}
}

func riderInviteAuditPayload(invite *MerchantOnboardingInvite) map[string]any {
	if invite == nil {
		return map[string]any{}
	}
	return map[string]any{
		"type":       strings.TrimSpace(invite.Type),
		"station_id": strings.TrimSpace(invite.StationID),
		"expires_at": invite.ExpiresAt,
	}
}

func (s *Store) AcceptRiderInvite(req AcceptRiderInviteRequest) (*RiderAccount, error) {
	token := strings.TrimSpace(req.Token)
	passwordHash, err := hashLoginPassword(req.Password)
	if token == "" || err != nil {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	invite := s.merchantInvites[token]
	if invite == nil {
		return nil, ErrNotFound
	}
	if invite.Type != RiderAccountRider && invite.Type != RiderAccountStationManager {
		return nil, ErrInvalidArgument
	}
	if invite.Status != OnboardingInviteActive || !invite.ExpiresAt.After(time.Now().UTC()) || strings.TrimSpace(invite.StationID) == "" {
		return nil, ErrInvalidArgument
	}

	s.nextRiderID++
	prefix := "rider"
	if invite.Type == RiderAccountStationManager {
		prefix = "station_manager"
	}
	accountID := fmt.Sprintf("%s_%d", prefix, s.nextRiderID)
	account := &RiderAccount{
		ID:            accountID,
		StationID:     invite.StationID,
		Type:          invite.Type,
		Status:        "active",
		Online:        false,
		DepositStatus: DepositStatusUnpaid,
		Capacity:      1,
	}
	if invite.Type == RiderAccountStationManager {
		account.Capacity = 0
	}
	s.riders[accountID] = account
	s.riderPasswordHash[accountID] = passwordHash
	if invite.Type == RiderAccountRider {
		s.getOrCreateDepositLocked("rider", accountID)
	}
	invite.Status = OnboardingInviteUsed
	return cloneRiderAccount(account), nil
}

func (s *Store) LoginRider(req RiderLoginRequest) (*RiderAccount, error) {
	accountID := strings.TrimSpace(req.AccountID)
	password := strings.TrimSpace(req.Password)
	if accountID == "" || password == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	account := cloneRiderAccount(s.riders[accountID])
	passwordHash := s.riderPasswordHash[accountID]
	s.mu.Unlock()

	if account == nil || account.Status != "active" || passwordHash == "" {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return account, nil
}

func (s *Store) SaveMerchantQualification(req UploadMerchantQualificationRequest) (*MerchantProfile, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	qualificationType := strings.TrimSpace(req.Type)
	fileURL := strings.TrimSpace(req.FileURL)
	if merchantID == "" || fileURL == "" || !isMerchantQualificationType(qualificationType) || !req.ExpiresAt.After(time.Now().UTC()) {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.merchants[merchantID] == nil {
		return nil, ErrNotFound
	}
	qualification := &MerchantQualification{
		ID:        "mq_" + shortHash(fmt.Sprintf("%s:%s:%s", merchantID, qualificationType, fileURL)),
		Type:      qualificationType,
		FileURL:   fileURL,
		ExpiresAt: req.ExpiresAt.UTC(),
		Status:    QualificationStatusPendingReview,
	}
	existing := s.merchantQualifications[merchantID]
	replaced := false
	for index, item := range existing {
		if item.Type == qualificationType {
			existing[index] = qualification
			replaced = true
			break
		}
	}
	if !replaced {
		existing = append(existing, qualification)
	}
	s.merchantQualifications[merchantID] = existing
	return s.merchantProfileLocked(merchantID), nil
}

func (s *Store) AdminMerchantQualifications(req AdminMerchantQualificationListRequest) (*AdminMerchantQualificationList, error) {
	req, err := normalizeAdminMerchantQualificationListRequest(req)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	counts := AdminMerchantQualificationCounts{}
	qualifications := make([]AdminMerchantQualificationCase, 0)
	for merchantID, entries := range s.merchantQualifications {
		merchantID = strings.TrimSpace(merchantID)
		if req.MerchantID != "" && merchantID != req.MerchantID {
			continue
		}
		for _, item := range entries {
			if item == nil {
				continue
			}
			if req.Type != "" && item.Type != req.Type {
				continue
			}
			status := adminMerchantQualificationStatus(item, req.Now)
			countAdminMerchantQualificationStatus(&counts, status)
			if req.Status != "all" && status != req.Status {
				continue
			}
			qualification, ok := s.adminMerchantQualificationCaseLocked(merchantID, item, req.Now)
			if !ok {
				continue
			}
			qualifications = append(qualifications, qualification)
		}
	}
	sortAdminMerchantQualificationCases(qualifications)
	if len(qualifications) > req.Limit {
		qualifications = qualifications[:req.Limit]
	}
	return &AdminMerchantQualificationList{
		GeneratedAt:    req.Now,
		Status:         req.Status,
		MerchantID:     req.MerchantID,
		Type:           req.Type,
		Counts:         counts,
		Qualifications: qualifications,
	}, nil
}

func (s *Store) AdminMerchantQualificationDetail(req AdminMerchantQualificationDetailRequest) (*AdminMerchantQualificationDetail, error) {
	req, err := normalizeAdminMerchantQualificationDetailRequest(req)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	qualification, ok := s.adminMerchantQualificationCaseByIDLocked(req.QualificationID, req.Now)
	s.mu.Unlock()
	if !ok {
		return nil, ErrNotFound
	}

	audits, err := s.AuditLogs(AuditLogsRequest{
		Action:     "admin.merchant_qualification.reviewed",
		TargetType: "merchant_qualification",
		TargetID:   qualification.Qualification.ID,
		Limit:      req.AuditLimit,
	})
	if err != nil {
		return nil, err
	}

	return &AdminMerchantQualificationDetail{
		GeneratedAt:                req.Now,
		Qualification:              qualification.Qualification,
		Merchant:                   qualification.Merchant,
		Shops:                      qualification.Shops,
		Deposit:                    qualification.Deposit,
		MissingQualifications:      qualification.MissingQualifications,
		CanAcceptOrders:            qualification.CanAcceptOrders,
		QualificationPopupRequired: qualification.QualificationPopupRequired,
		QualificationPopupCode:     qualification.QualificationPopupCode,
		IncidentCode:               qualification.IncidentCode,
		IncidentSeverity:           qualification.IncidentSeverity,
		ExpiresInSeconds:           qualification.ExpiresInSeconds,
		ReviewSLAHours:             qualification.ReviewSLAHours,
		AuditFilters: []AdminAuditFilter{{
			TargetType: "merchant_qualification",
			TargetID:   qualification.Qualification.ID,
			Action:     "admin.merchant_qualification.reviewed",
			Limit:      req.AuditLimit,
		}},
		RecentAudits:         audits,
		RecommendedOperation: qualification.RecommendedOperation,
		Checklist:            adminMerchantQualificationChecklist(qualification),
	}, nil
}

func (s *Store) ReviewMerchantQualification(req ReviewMerchantQualificationRequest) (*MerchantProfile, *MerchantQualification, error) {
	normalized, status, err := normalizeReviewMerchantQualificationRequest(req)
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	profile, qualification, err := s.reviewMerchantQualificationLocked(normalized, status)
	if err != nil {
		return nil, nil, err
	}
	return profile, qualification, nil
}

func (s *Store) ReviewMerchantQualificationWithAudit(req ReviewMerchantQualificationRequest, audit RecordAuditLogRequest) (*MerchantProfile, *MerchantQualification, *AuditLog, *OutboxEvent, error) {
	normalized, status, err := normalizeReviewMerchantQualificationRequest(req)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if log.Action != "admin.merchant_qualification.reviewed" || log.TargetType != "merchant_qualification" || log.TargetID != normalized.QualificationID {
		return nil, nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	profile, qualification, err := s.reviewMerchantQualificationLocked(normalized, status)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	log.Payload = merchantQualificationReviewAuditPayload(normalized, qualification)
	auditLog := s.appendAuditLogLocked(log)
	outboxEvent := s.enqueueMerchantQualificationReviewOutboxLocked(normalized, qualification)
	return profile, qualification, auditLog, cloneOutboxEvent(outboxEvent), nil
}

func normalizeReviewMerchantQualificationRequest(req ReviewMerchantQualificationRequest) (ReviewMerchantQualificationRequest, string, error) {
	req.MerchantID = strings.TrimSpace(req.MerchantID)
	req.QualificationID = strings.TrimSpace(req.QualificationID)
	req.Decision = strings.TrimSpace(req.Decision)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.ReviewedAt.IsZero() {
		req.ReviewedAt = time.Now().UTC()
	} else {
		req.ReviewedAt = req.ReviewedAt.UTC()
	}
	status := ""
	switch req.Decision {
	case "approve", QualificationStatusApproved:
		req.Decision = "approve"
		status = QualificationStatusApproved
	case "reject", QualificationStatusRejected:
		req.Decision = "reject"
		status = QualificationStatusRejected
	default:
		return req, "", ErrInvalidArgument
	}
	if req.MerchantID == "" || req.QualificationID == "" || req.Reason == "" {
		return req, "", ErrInvalidArgument
	}
	return req, status, nil
}

func (s *Store) reviewMerchantQualificationLocked(req ReviewMerchantQualificationRequest, status string) (*MerchantProfile, *MerchantQualification, error) {
	if s.merchants[req.MerchantID] == nil {
		return nil, nil, ErrNotFound
	}
	entries := s.merchantQualifications[req.MerchantID]
	for _, qualification := range entries {
		if qualification == nil || qualification.ID != req.QualificationID {
			continue
		}
		if status == QualificationStatusApproved && !qualification.ExpiresAt.After(req.ReviewedAt) {
			return nil, nil, ErrInvalidArgument
		}
		qualification.Status = status
		profile := s.merchantProfileLocked(req.MerchantID)
		return profile, cloneMerchantQualification(qualification), nil
	}
	return nil, nil, ErrNotFound
}

func merchantQualificationReviewAuditPayload(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) map[string]any {
	if qualification == nil {
		return map[string]any{}
	}
	return map[string]any{
		"merchant_id":      req.MerchantID,
		"qualification_id": qualification.ID,
		"type":             qualification.Type,
		"decision":         req.Decision,
		"status":           qualification.Status,
		"reason":           req.Reason,
		"expires_at":       qualification.ExpiresAt.Format(time.RFC3339Nano),
		"reviewed_at":      req.ReviewedAt.Format(time.RFC3339Nano),
	}
}

func merchantQualificationReviewOutboxEvent(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) *OutboxEvent {
	idempotencyKey := merchantQualificationReviewOutboxIdempotencyKey(req, qualification)
	if qualification == nil || idempotencyKey == "" {
		return nil
	}
	now := req.ReviewedAt
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return &OutboxEvent{
		ID:             "obe_mq_review_" + shortHash(idempotencyKey),
		Topic:          "merchant.qualification_reviewed",
		AggregateType:  "merchant_qualification",
		AggregateID:    qualification.ID,
		EventType:      "merchant.qualification.reviewed",
		IdempotencyKey: idempotencyKey,
		Payload:        merchantQualificationReviewOutboxPayload(req, qualification),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func merchantQualificationReviewOutboxIdempotencyKey(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) string {
	if qualification == nil {
		return ""
	}
	qualificationID := strings.TrimSpace(qualification.ID)
	decision := strings.TrimSpace(req.Decision)
	if qualificationID == "" || decision == "" || req.ReviewedAt.IsZero() {
		return ""
	}
	return fmt.Sprintf("merchant_qualification_review:%s:%s:%s", qualificationID, decision, req.ReviewedAt.UTC().Format(time.RFC3339Nano))
}

func (s *Store) enqueueMerchantQualificationReviewOutboxLocked(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) *OutboxEvent {
	if qualification == nil {
		return nil
	}
	return s.enqueueOutboxEventLocked(
		"merchant.qualification_reviewed",
		"merchant_qualification",
		qualification.ID,
		"merchant.qualification.reviewed",
		merchantQualificationReviewOutboxIdempotencyKey(req, qualification),
		merchantQualificationReviewOutboxPayload(req, qualification),
		req.ReviewedAt,
	)
}

func (s *Store) enqueueMerchantQualificationReviewOutbox(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) *OutboxEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneOutboxEvent(s.enqueueMerchantQualificationReviewOutboxLocked(req, qualification))
}

func merchantQualificationReviewOutboxPayload(req ReviewMerchantQualificationRequest, qualification *MerchantQualification) map[string]any {
	if qualification == nil {
		return map[string]any{}
	}
	title := "商户资质审核结果"
	body := "资质审核未通过，请补充有效文件后重新提交。"
	if qualification.Status == QualificationStatusApproved {
		body = "资质审核已通过，系统已更新商户接单资格。"
	}
	return map[string]any{
		"type":               "merchant.qualification_reviewed",
		"merchant_id":        req.MerchantID,
		"qualification_id":   qualification.ID,
		"qualification_type": qualification.Type,
		"decision":           req.Decision,
		"status":             qualification.Status,
		"reason":             req.Reason,
		"expires_at":         qualification.ExpiresAt.Format(time.RFC3339Nano),
		"reviewed_at":        req.ReviewedAt.Format(time.RFC3339Nano),
		"target": map[string]any{
			"role": "merchant",
			"id":   req.MerchantID,
		},
		"title": title,
		"body":  body,
	}
}

func normalizeAdminMerchantQualificationListRequest(req AdminMerchantQualificationListRequest) (AdminMerchantQualificationListRequest, error) {
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = QualificationStatusPendingReview
	}
	switch req.Status {
	case "all", QualificationStatusPendingReview, QualificationStatusApproved, QualificationStatusRejected, "expired":
	default:
		return req, ErrInvalidArgument
	}
	req.MerchantID = strings.TrimSpace(req.MerchantID)
	req.Type = strings.TrimSpace(req.Type)
	if req.Type != "" && !isMerchantQualificationType(req.Type) {
		return req, ErrInvalidArgument
	}
	req.Now = req.Now.UTC()
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}
	req.Limit = normalizeAdminOperationsSnapshotLimit(req.Limit)
	return req, nil
}

func normalizeAdminMerchantQualificationDetailRequest(req AdminMerchantQualificationDetailRequest) (AdminMerchantQualificationDetailRequest, error) {
	req.QualificationID = strings.TrimSpace(req.QualificationID)
	if req.QualificationID == "" {
		return req, ErrInvalidArgument
	}
	req.Now = req.Now.UTC()
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}
	if req.AuditLimit <= 0 {
		req.AuditLimit = 20
	}
	if req.AuditLimit > 50 {
		req.AuditLimit = 50
	}
	return req, nil
}

func adminMerchantQualificationStatus(qualification *MerchantQualification, now time.Time) string {
	if qualification == nil {
		return ""
	}
	if !qualification.ExpiresAt.IsZero() && !qualification.ExpiresAt.After(now.UTC()) {
		return "expired"
	}
	return strings.TrimSpace(qualification.Status)
}

func countAdminMerchantQualificationStatus(counts *AdminMerchantQualificationCounts, status string) {
	if counts == nil {
		return
	}
	counts.Total++
	switch status {
	case QualificationStatusPendingReview:
		counts.PendingReview++
	case QualificationStatusApproved:
		counts.Approved++
	case QualificationStatusRejected:
		counts.Rejected++
	case "expired":
		counts.Expired++
	}
}

func (s *Store) adminMerchantQualificationCaseByIDLocked(qualificationID string, now time.Time) (AdminMerchantQualificationCase, bool) {
	for merchantID, entries := range s.merchantQualifications {
		for _, item := range entries {
			if item == nil || item.ID != qualificationID {
				continue
			}
			return s.adminMerchantQualificationCaseLocked(merchantID, item, now)
		}
	}
	return AdminMerchantQualificationCase{}, false
}

func (s *Store) adminMerchantQualificationCaseLocked(merchantID string, qualification *MerchantQualification, now time.Time) (AdminMerchantQualificationCase, bool) {
	profile := s.merchantProfileLocked(merchantID)
	if profile == nil || qualification == nil {
		return AdminMerchantQualificationCase{}, false
	}
	cloned := cloneMerchantQualification(qualification)
	status := adminMerchantQualificationStatus(cloned, now)
	incidentCode, incidentSeverity := adminMerchantQualificationIncident(status)
	return AdminMerchantQualificationCase{
		Qualification:              *cloned,
		Merchant:                   profile.Account,
		Shops:                      s.adminMerchantShopsLocked(merchantID),
		Deposit:                    cloneDepositAccount(s.deposits[depositKey("merchant", merchantID)]),
		MissingQualifications:      append([]string{}, profile.MissingQualifications...),
		CanAcceptOrders:            profile.CanAcceptOrders,
		QualificationPopupRequired: profile.QualificationPopupRequired,
		QualificationPopupCode:     profile.QualificationPopupCode,
		IncidentCode:               incidentCode,
		IncidentSeverity:           incidentSeverity,
		ExpiresInSeconds:           adminMerchantQualificationExpiresInSeconds(cloned, now),
		ReviewSLAHours:             24,
		RecommendedOperation:       adminMerchantQualificationRecommendedOperation(merchantID, cloned, status),
	}, true
}

func (s *Store) adminMerchantShopsLocked(merchantID string) []Shop {
	shops := make([]Shop, 0)
	for _, shop := range s.shops {
		if shop == nil || shop.MerchantID != merchantID {
			continue
		}
		cloned := cloneShop(shop)
		if !s.shopCanAcceptOrdersLocked(cloned.ID) {
			cloned.Status = ShopStatusQualificationExpired
		}
		shops = append(shops, *cloned)
	}
	sort.SliceStable(shops, func(i, j int) bool {
		return shops[i].ID < shops[j].ID
	})
	return shops
}

func adminMerchantQualificationIncident(status string) (string, string) {
	switch status {
	case "expired":
		return "merchant_qualification.expired", "critical"
	case QualificationStatusPendingReview:
		return "merchant_qualification.pending_review", "warning"
	case QualificationStatusRejected:
		return "merchant_qualification.rejected", "critical"
	case QualificationStatusApproved:
		return "merchant_qualification.approved", "info"
	default:
		return "merchant_qualification.unknown", "warning"
	}
}

func adminMerchantQualificationExpiresInSeconds(qualification *MerchantQualification, now time.Time) int64 {
	if qualification == nil || qualification.ExpiresAt.IsZero() {
		return 0
	}
	return int64(qualification.ExpiresAt.Sub(now.UTC()).Seconds())
}

func adminMerchantQualificationRecommendedOperation(merchantID string, qualification *MerchantQualification, status string) AdminRecommendedOperation {
	if qualification == nil {
		return AdminRecommendedOperation{}
	}
	switch status {
	case QualificationStatusPendingReview:
		return AdminRecommendedOperation{
			Key:    "merchant-qualification-review",
			Title:  "审核商户资质",
			Reason: "资质待复核，审核结果会直接影响商户接单资格。",
			Values: map[string]any{
				"merchant_id":      merchantID,
				"qualification_id": qualification.ID,
				"decision":         "approve",
				"reason":           "资质原件核验通过",
			},
		}
	case "expired":
		return AdminRecommendedOperation{
			Key:    "merchant-qualification-review",
			Title:  "驳回过期资质",
			Reason: "资质文件已过有效期，应驳回并要求商户补传有效文件。",
			Values: map[string]any{
				"merchant_id":      merchantID,
				"qualification_id": qualification.ID,
				"decision":         "reject",
				"reason":           "资质已过期，需补传有效文件",
			},
		}
	default:
		return AdminRecommendedOperation{
			Key:    "audit-logs",
			Title:  "查看审核审计",
			Reason: "当前资质不处于待审状态，下一步应核对历史审核记录。",
			Values: map[string]any{
				"target_type": "merchant_qualification",
				"target_id":   qualification.ID,
				"action":      "admin.merchant_qualification.reviewed",
				"limit":       50,
			},
		}
	}
}

func adminMerchantQualificationChecklist(qualification AdminMerchantQualificationCase) []string {
	checklist := []string{
		"核验文件主体、证照编号和商户账号主体一致",
		"确认有效期覆盖当前接单周期且原件清晰可追溯",
		"处理后回查商户接单资格、店铺状态和审核审计",
	}
	switch qualification.IncidentCode {
	case "merchant_qualification.pending_review":
		return append([]string{"待审资质应在 24 小时内完成复核"}, checklist...)
	case "merchant_qualification.expired":
		return append([]string{"过期资质不得通过，应要求商户重新上传"}, checklist...)
	case "merchant_qualification.rejected":
		return append([]string{"已驳回资质需等待商户补件后再复核"}, checklist...)
	default:
		return checklist
	}
}

func sortAdminMerchantQualificationCases(qualifications []AdminMerchantQualificationCase) {
	sort.SliceStable(qualifications, func(i, j int) bool {
		leftRank := adminMerchantQualificationSortRank(qualifications[i].IncidentCode)
		rightRank := adminMerchantQualificationSortRank(qualifications[j].IncidentCode)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if !qualifications[i].Qualification.ExpiresAt.Equal(qualifications[j].Qualification.ExpiresAt) {
			return qualifications[i].Qualification.ExpiresAt.Before(qualifications[j].Qualification.ExpiresAt)
		}
		if qualifications[i].Merchant.ID != qualifications[j].Merchant.ID {
			return qualifications[i].Merchant.ID < qualifications[j].Merchant.ID
		}
		if qualifications[i].Qualification.Type != qualifications[j].Qualification.Type {
			return qualifications[i].Qualification.Type < qualifications[j].Qualification.Type
		}
		return qualifications[i].Qualification.ID < qualifications[j].Qualification.ID
	})
}

func adminMerchantQualificationSortRank(incidentCode string) int {
	switch incidentCode {
	case "merchant_qualification.expired":
		return 40
	case "merchant_qualification.pending_review":
		return 30
	case "merchant_qualification.rejected":
		return 20
	case "merchant_qualification.approved":
		return 10
	default:
		return 0
	}
}

func (s *Store) MerchantProfile(merchantID string) (*MerchantProfile, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	profile := s.merchantProfileLocked(merchantID)
	if profile == nil {
		return nil, ErrNotFound
	}
	return profile, nil
}

func (s *Store) MerchantStaff(merchantID string) ([]MerchantStaff, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.merchants[merchantID] == nil {
		return nil, ErrNotFound
	}
	return s.merchantStaffLocked(merchantID), nil
}

func (s *Store) SaveMerchantStaff(req UpsertMerchantStaffRequest) (*MerchantStaff, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	staffID := strings.TrimSpace(req.StaffID)
	shopID := strings.TrimSpace(req.ShopID)
	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)
	role := strings.TrimSpace(req.Role)
	status := strings.TrimSpace(req.Status)
	healthCertificateURL := strings.TrimSpace(req.HealthCertificateURL)
	if role == "" {
		role = "staff"
	}
	if status == "" {
		status = MerchantStaffActive
	}
	if merchantID == "" || shopID == "" || name == "" || phone == "" || healthCertificateURL == "" || !req.HealthCertificateExpiresAt.After(time.Now().UTC()) {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.merchantOwnsShopLocked(merchantID, shopID) {
		return nil, ErrNotFound
	}
	existing := s.merchantStaff[merchantID]
	var staff *MerchantStaff
	if staffID != "" {
		for _, item := range existing {
			if item != nil && item.ID == staffID {
				staff = item
				break
			}
		}
		if staff == nil {
			return nil, ErrNotFound
		}
	} else {
		s.nextMerchantStaffID++
		staff = &MerchantStaff{ID: fmt.Sprintf("staff_%d", s.nextMerchantStaffID), MerchantID: merchantID}
		existing = append(existing, staff)
	}
	staff.ShopID = shopID
	staff.Name = name
	staff.Phone = phone
	staff.Role = role
	staff.Status = status
	staff.HealthCertificateURL = healthCertificateURL
	staff.HealthCertificateExpiresAt = req.HealthCertificateExpiresAt.UTC()
	s.merchantStaff[merchantID] = existing
	return cloneMerchantStaff(staff), nil
}

func (s *Store) MerchantSupplementalMaterials(merchantID string) ([]MerchantSupplementalMaterial, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.merchants[merchantID] == nil {
		return nil, ErrNotFound
	}
	return s.merchantMaterialsLocked(merchantID), nil
}

func (s *Store) SaveMerchantSupplementalMaterial(req UploadMerchantSupplementalMaterialRequest) (*MerchantSupplementalMaterial, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	materialID := strings.TrimSpace(req.MaterialID)
	shopID := strings.TrimSpace(req.ShopID)
	materialType := strings.TrimSpace(req.Type)
	fileURL := strings.TrimSpace(req.FileURL)
	description := strings.TrimSpace(req.Description)
	if merchantID == "" || shopID == "" || materialType == "" || fileURL == "" {
		return nil, ErrInvalidArgument
	}
	if !req.ExpiresAt.IsZero() && !req.ExpiresAt.After(time.Now().UTC()) {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.merchantOwnsShopLocked(merchantID, shopID) {
		return nil, ErrNotFound
	}
	existing := s.merchantMaterials[merchantID]
	var material *MerchantSupplementalMaterial
	if materialID != "" {
		for _, item := range existing {
			if item != nil && item.ID == materialID {
				material = item
				break
			}
		}
		if material == nil {
			return nil, ErrNotFound
		}
	} else {
		s.nextMerchantMaterialID++
		material = &MerchantSupplementalMaterial{
			ID:         fmt.Sprintf("material_%d", s.nextMerchantMaterialID),
			MerchantID: merchantID,
			UploadedAt: time.Now().UTC(),
		}
		existing = append(existing, material)
	}
	material.ShopID = shopID
	material.Type = materialType
	material.FileURL = fileURL
	material.Description = description
	material.Status = "submitted"
	material.ExpiresAt = req.ExpiresAt.UTC()
	s.merchantMaterials[merchantID] = existing
	return cloneMerchantSupplementalMaterial(material), nil
}

func (s *Store) HomeModules() []HomeModule {
	s.mu.Lock()
	defer s.mu.Unlock()
	modules := cloneHomeModules(s.homeModules)
	sort.SliceStable(modules, func(i, j int) bool {
		if modules[i].SortOrder == modules[j].SortOrder {
			return modules[i].Key < modules[j].Key
		}
		return modules[i].SortOrder < modules[j].SortOrder
	})
	return modules
}

func (s *Store) HomeCards() []HomeCard {
	s.mu.Lock()
	defer s.mu.Unlock()
	cards := cloneHomeCards(s.homeCards)
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].SortOrder == cards[j].SortOrder {
			return cards[i].ID < cards[j].ID
		}
		return cards[i].SortOrder < cards[j].SortOrder
	})
	return cards
}

func (s *Store) Shops() []Shop {
	s.mu.Lock()
	defer s.mu.Unlock()
	shops := make([]Shop, 0, len(s.shops))
	for _, shop := range s.shops {
		cloned := cloneShop(shop)
		if !s.shopCanAcceptOrdersLocked(cloned.ID) {
			cloned.Status = ShopStatusQualificationExpired
		}
		shops = append(shops, *cloned)
	}
	sort.SliceStable(shops, func(i, j int) bool {
		return shops[i].ID < shops[j].ID
	})
	return shops
}

func (s *Store) ShopDetail(shopID string) (*ShopDetail, error) {
	shopID = strings.TrimSpace(shopID)
	if shopID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	shop := s.shops[shopID]
	if shop == nil {
		return nil, ErrNotFound
	}
	seed := shopDetailSeed(shopID)
	profile := s.merchantProfileLocked(shop.MerchantID)
	reviews := s.shopReviewEntriesLocked(shopID)
	summary := buildShopReviewSummary(reviews)
	qualificationText := seed.QualificationText
	merchantName := shop.Name
	contactPhone := ""
	qualificationItems := append([]string{}, seed.QualificationItems...)
	if profile != nil {
		merchantName = defaultString(profile.Account.DisplayName, merchantName)
		if qualificationText == "" {
			if profile.CanAcceptOrders {
				qualificationText = "资质齐全"
			} else {
				qualificationText = "资质审核中"
			}
		}
		if len(qualificationItems) == 0 {
			qualificationItems = merchantQualificationItemsForShop(profile)
		}
		if staff := s.merchantStaffLocked(shop.MerchantID); len(staff) > 0 {
			contactPhone = staff[0].Phone
		}
	}
	if qualificationText == "" {
		qualificationText = "资质已审核"
	}
	if contactPhone == "" {
		contactPhone = seed.ContactPhone
	}
	cloned := cloneShop(shop)
	if !s.shopCanAcceptOrdersLocked(cloned.ID) {
		cloned.Status = ShopStatusQualificationExpired
	}
	return &ShopDetail{
		ShopID:            cloned.ID,
		MerchantID:        cloned.MerchantID,
		Name:              cloned.Name,
		Category:          cloned.Category,
		CoverURL:          cloned.CoverURL,
		LogoURL:           cloned.LogoURL,
		Announcement:      defaultString(cloned.Announcement, seed.Announcement),
		RatingText:        defaultString(seed.RatingText, summary.AverageRating),
		SalesText:         seed.SalesText,
		DeliveryText:      seed.DeliveryText,
		QualificationText: qualificationText,
		ActivityTags:      append([]string{}, seed.ActivityTags...),
		ReviewSummary:     summary,
		Reviews:           reviews,
		MerchantInfo: ShopMerchantInfo{
			MerchantName:       merchantName,
			QualificationText:  qualificationText,
			BusinessHours:      seed.BusinessHours,
			ContactPhone:       contactPhone,
			Address:            seed.Address,
			ServiceCommitments: append([]string{}, seed.ServiceCommitments...),
			QualificationItems: qualificationItems,
			SupportBulletins:   append([]string{}, seed.SupportBulletins...),
		},
	}, nil
}

func (s *Store) ShopProducts(shopID string) ([]MerchantProduct, error) {
	shopID = strings.TrimSpace(shopID)
	if shopID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shops[shopID] == nil {
		return nil, ErrNotFound
	}
	products := make([]MerchantProduct, 0)
	for _, product := range s.products {
		if product.ShopID == shopID && product.Status == ProductStatusActive {
			products = append(products, *cloneMerchantProduct(product))
		}
	}
	sort.SliceStable(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func (s *Store) MerchantProducts(merchantID string, shopID string) ([]MerchantProduct, error) {
	merchantID = strings.TrimSpace(merchantID)
	shopID = strings.TrimSpace(shopID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if shopID != "" && !s.merchantOwnsShopLocked(merchantID, shopID) {
		return nil, ErrNotFound
	}
	products := make([]MerchantProduct, 0)
	for _, product := range s.products {
		if product == nil {
			continue
		}
		if shopID != "" && product.ShopID != shopID {
			continue
		}
		if s.merchantOwnsShopLocked(merchantID, product.ShopID) {
			products = append(products, *cloneMerchantProduct(product))
		}
	}
	sort.SliceStable(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func (s *Store) UpsertMerchantProduct(req UpsertMerchantProductRequest) (*MerchantProduct, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	productID := strings.TrimSpace(req.ProductID)
	shopID := strings.TrimSpace(req.ShopID)
	name := strings.TrimSpace(req.Name)
	status := normalizeProductStatus(req.Status)
	if merchantID == "" || name == "" || req.PriceFen <= 0 || req.StockCount < 0 || status == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	existing := s.products[productID]
	if existing != nil {
		if shopID == "" {
			shopID = existing.ShopID
		}
		if existing.ShopID != shopID {
			return nil, ErrInvalidArgument
		}
	}
	if !s.merchantOwnsShopLocked(merchantID, shopID) {
		return nil, ErrNotFound
	}
	if existing == nil {
		s.nextProductID++
		productID = fmt.Sprintf("prod_custom_%d", s.nextProductID)
		existing = &MerchantProduct{ID: productID, ShopID: shopID}
		s.products[productID] = existing
	}
	existing.Name = name
	existing.ImageURL = strings.TrimSpace(req.ImageURL)
	existing.Description = strings.TrimSpace(req.Description)
	existing.IngredientList = normalizeIngredientList(req.IngredientList)
	existing.PriceFen = req.PriceFen
	existing.StockCount = req.StockCount
	existing.Status = status
	return cloneMerchantProduct(existing), nil
}

func (s *Store) SetMerchantProductStatus(req SetMerchantProductStatusRequest) (*MerchantProduct, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	productID := strings.TrimSpace(req.ProductID)
	status := normalizeProductStatus(req.Status)
	if merchantID == "" || productID == "" || status == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	product := s.products[productID]
	if product == nil {
		return nil, ErrNotFound
	}
	if !s.merchantOwnsShopLocked(merchantID, product.ShopID) {
		return nil, ErrNotFound
	}
	product.Status = status
	return cloneMerchantProduct(product), nil
}

func (s *Store) ShopGroupbuyDeals(shopID string) ([]MerchantProduct, error) {
	shopID = strings.TrimSpace(shopID)
	if shopID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shops[shopID] == nil {
		return nil, ErrNotFound
	}
	deals := make([]MerchantProduct, 0)
	for _, deal := range s.groupbuyDeals {
		if deal.ShopID == shopID && deal.Status == ProductStatusActive {
			deals = append(deals, *cloneMerchantProduct(deal))
		}
	}
	sort.SliceStable(deals, func(i, j int) bool {
		return deals[i].ID < deals[j].ID
	})
	return deals, nil
}

func (s *Store) CreateGroupbuyOrder(req CreateGroupbuyOrderRequest) (*Order, error) {
	userID := strings.TrimSpace(req.UserID)
	shopID := strings.TrimSpace(req.ShopID)
	dealID := strings.TrimSpace(req.DealID)
	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}
	if userID == "" || shopID == "" || dealID == "" || quantity <= 0 || quantity > 20 {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	deal := s.groupbuyDeals[dealID]
	if deal == nil || deal.ShopID != shopID || deal.Status != ProductStatusActive {
		return nil, ErrNotFound
	}
	if !s.shopCanAcceptOrdersLocked(shopID) {
		return nil, ErrInvalidOrderState
	}
	if deal.StockCount < quantity {
		return nil, ErrInvalidArgument
	}
	deal.StockCount -= quantity

	s.nextOrderID++
	now := time.Now().UTC()
	amountFen := deal.PriceFen * int64(quantity)
	order := &Order{
		ID:            fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:        userID,
		ShopID:        shopID,
		Type:          OrderTypeGroupbuy,
		Status:        StatusPendingPayment,
		AmountFen:     amountFen,
		ItemsTotalFen: amountFen,
		Items: []OrderItem{{
			ProductID:    deal.ID,
			ProductName:  deal.Name,
			ImageURL:     deal.ImageURL,
			UnitPriceFen: deal.PriceFen,
			Quantity:     quantity,
		}},
		CreatedAt: now,
		UpdatedAt: now,
		Events: []OrderEvent{{
			Type:      "groupbuy.order_created",
			ActorID:   userID,
			Message:   "团购订单已创建，待支付后发券",
			CreatedAt: now,
		}},
	}
	s.orders[order.ID] = order
	return cloneOrder(order), nil
}

func (s *Store) UserGroupbuyVouchers(userID string) ([]GroupbuyVoucher, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	vouchers := make([]GroupbuyVoucher, 0)
	for _, voucher := range s.groupbuyVouchers {
		if voucher.UserID == userID {
			vouchers = append(vouchers, *cloneGroupbuyVoucher(voucher))
		}
	}
	sort.SliceStable(vouchers, func(i, j int) bool {
		return vouchers[i].CreatedAt.After(vouchers[j].CreatedAt)
	})
	return vouchers, nil
}

func (s *Store) RedeemGroupbuyVoucher(req RedeemGroupbuyVoucherRequest) (*GroupbuyVoucher, *Order, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	code := groupbuyVoucherCodeFromScan(req.VoucherCode, req.QRPayload)
	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = GroupbuyRedemptionMethodQR
	}
	if merchantID == "" || code == "" || method != GroupbuyRedemptionMethodQR {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	voucher := s.vouchersByCode[code]
	if voucher == nil {
		return nil, nil, ErrNotFound
	}
	if !s.merchantOwnsShopLocked(merchantID, voucher.ShopID) {
		return nil, nil, ErrNotFound
	}
	if voucher.Status != GroupbuyVoucherStatusIssued {
		return nil, nil, ErrInvalidOrderState
	}
	order := s.orders[voucher.OrderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	now := time.Now().UTC()
	voucher.Status = GroupbuyVoucherRedeemed
	voucher.RedemptionMethod = GroupbuyRedemptionMethodQR
	voucher.RedeemedByMerchantID = merchantID
	voucher.RedeemedAt = now
	order.Status = StatusCompleted
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "groupbuy.voucher_redeemed",
		ActorID:   merchantID,
		Message:   "团购券已到店扫码核销",
		CreatedAt: now,
	})
	return cloneGroupbuyVoucher(voucher), cloneOrder(order), nil
}

func (s *Store) SaveAddress(address UserAddress) (*UserAddress, error) {
	address.UserID = strings.TrimSpace(address.UserID)
	address.ContactName = strings.TrimSpace(address.ContactName)
	address.ContactPhone = strings.TrimSpace(address.ContactPhone)
	address.City = strings.TrimSpace(address.City)
	address.Detail = strings.TrimSpace(address.Detail)
	address.Tag = strings.TrimSpace(address.Tag)
	if address.Tag == "" {
		address.Tag = "other"
	}
	if !UserAddressReady(address) {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(address.ID) == "" {
		s.nextAddressID++
		address.ID = fmt.Sprintf("addr_%d", s.nextAddressID)
	}
	if address.IsDefault {
		for _, existing := range s.addresses[address.UserID] {
			existing.IsDefault = false
		}
	}
	cloned := address
	addresses := s.addresses[address.UserID]
	replaced := false
	for index, existing := range addresses {
		if existing.ID == cloned.ID {
			addresses[index] = &cloned
			replaced = true
			break
		}
	}
	if !replaced {
		addresses = append(addresses, &cloned)
	}
	s.addresses[address.UserID] = addresses
	return cloneUserAddress(&cloned), nil
}

func (s *Store) UserAddresses(userID string) []UserAddress {
	userID = strings.TrimSpace(userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	output := make([]UserAddress, 0, len(s.addresses[userID]))
	for _, address := range s.addresses[userID] {
		output = append(output, *cloneUserAddress(address))
	}
	sort.SliceStable(output, func(i, j int) bool {
		return output[i].ID < output[j].ID
	})
	return output
}

func (s *Store) UpsertCartItem(req UpsertCartItemRequest) (*CartSummary, error) {
	userID := strings.TrimSpace(req.UserID)
	shopID := strings.TrimSpace(req.ShopID)
	productID := strings.TrimSpace(req.ProductID)
	if userID == "" || shopID == "" || productID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	product := s.products[productID]
	if product == nil || product.ShopID != shopID || product.Status != ProductStatusActive {
		return nil, ErrNotFound
	}
	if !s.shopCanAcceptOrdersLocked(shopID) {
		return nil, ErrInvalidOrderState
	}
	if req.Quantity > product.StockCount {
		return nil, ErrInvalidArgument
	}
	selected := true
	if req.Selected != nil {
		selected = *req.Selected
	}
	key := cartKey(userID, shopID)
	items := s.cartItems[key]
	next := make([]*CartItem, 0, len(items)+1)
	for _, item := range items {
		if item.ProductID != productID {
			next = append(next, cloneCartItem(item))
		}
	}
	if req.Quantity > 0 {
		next = append(next, &CartItem{
			UserID:       userID,
			ShopID:       shopID,
			ProductID:    product.ID,
			ProductName:  product.Name,
			ImageURL:     product.ImageURL,
			UnitPriceFen: product.PriceFen,
			Quantity:     req.Quantity,
			Selected:     selected,
		})
	}
	s.cartItems[key] = next
	return s.cartSummaryLocked(userID, shopID), nil
}

func (s *Store) CartSummary(userID string, shopID string) (*CartSummary, error) {
	userID = strings.TrimSpace(userID)
	shopID = strings.TrimSpace(shopID)
	if userID == "" || shopID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cartSummaryLocked(userID, shopID), nil
}

func (s *Store) CreateOrder(req CreateOrderRequest) (*Order, error) {
	userID := strings.TrimSpace(req.UserID)
	orderType := strings.TrimSpace(req.Type)
	if userID == "" || !IsOrderType(orderType) || req.AmountFen <= 0 {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextOrderID++
	now := time.Now().UTC()
	order := &Order{
		ID:        fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:    userID,
		Type:      orderType,
		Status:    StatusPendingPayment,
		AmountFen: req.AmountFen,
		CreatedAt: now,
		UpdatedAt: now,
		Events: []OrderEvent{{
			Type:      "order.created",
			ActorID:   userID,
			Message:   "订单已创建",
			CreatedAt: now,
		}},
	}
	s.orders[order.ID] = order
	return cloneOrder(order), nil
}

func (s *Store) UserOrders(userID string) ([]Order, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	orders := make([]Order, 0)
	for _, order := range s.orders {
		if order.UserID == userID {
			orderCopy := cloneOrder(order)
			orderCopy.Reviewed = s.userReviewedOrderLocked(userID, order.ID)
			orders = append(orders, *orderCopy)
		}
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

func (s *Store) OrderByID(orderID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	orderCopy := cloneOrder(order)
	orderCopy.Reviewed = s.userReviewedOrderLocked(order.UserID, order.ID)
	return orderCopy, nil
}

func (s *Store) CreateReview(review Review) (*Review, error) {
	review.UserID = strings.TrimSpace(review.UserID)
	review.OrderID = strings.TrimSpace(review.OrderID)
	review.TargetType = strings.TrimSpace(review.TargetType)
	review.TargetID = strings.TrimSpace(review.TargetID)
	review.Content = strings.TrimSpace(review.Content)
	if review.TargetType == "" {
		review.TargetType = ReviewTargetOrder
	}
	if review.TargetType == ReviewTargetOrder && review.TargetID == "" {
		review.TargetID = review.OrderID
	}
	if review.Status == "" {
		review.Status = ReviewPublished
	}
	if review.Rating < 1 {
		review.Rating = 1
	}
	if review.Rating > 5 {
		review.Rating = 5
	}
	if review.RiderRating < 1 || review.RiderRating > 5 {
		review.RiderRating = 0
	}
	if review.UserID == "" || review.TargetID == "" || review.Content == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	review.ImageURLs = sanitizedStringSlice(review.ImageURLs)
	review.Tags = sanitizedStringSlice(review.Tags)
	review.ItemRatings = sanitizeReviewItemRatings(review.ItemRatings, s.orders[review.OrderID])
	if review.OrderID != "" {
		for _, existing := range s.reviews {
			if existing == nil || existing.UserID != review.UserID || existing.OrderID != review.OrderID {
				continue
			}
			existing.TargetType = review.TargetType
			existing.TargetID = review.TargetID
			existing.Rating = review.Rating
			existing.RiderRating = review.RiderRating
			existing.Content = review.Content
			existing.ImageURLs = append([]string{}, review.ImageURLs...)
			existing.ItemRatings = cloneReviewItemRatings(review.ItemRatings)
			existing.Anonymous = review.Anonymous
			existing.Status = review.Status
			existing.Tags = append([]string{}, review.Tags...)
			existing.CreatedAt = now
			return cloneReview(existing), nil
		}
	}
	s.nextReviewID++
	review.ID = fmt.Sprintf("rev_%d", s.nextReviewID)
	review.CreatedAt = now
	s.reviews[review.ID] = cloneReview(&review)
	return cloneReview(&review), nil
}

func (s *Store) CreateReviewImageUpload(req CreateReviewImageUploadRequest) (*ObjectUploadTicket, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	if req.UserID == "" || req.OrderID == "" || req.FileName == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[req.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if order.UserID != req.UserID {
		return nil, ErrInvalidArgument
	}
	now := time.Now().UTC()
	storage := s.objectStorageConfigLocked()
	scanStatus := AfterSalesUploadScanNotRequired
	if storage.RequireScanApprovalForConfirm {
		scanStatus = AfterSalesUploadScanPending
	}
	expiresAt := now.Add(storage.TicketTTL)
	objectKey := fmt.Sprintf("reviews/%s/%s/%s/%s", shortHash(req.UserID), shortHash(req.OrderID), shortHash(fmt.Sprintf("%s:%s:%d", req.UserID, req.FileName, now.UnixNano())), req.FileName)
	ticket, err := storage.createObjectUploadTicket(objectUploadTicketInput{
		ObjectKey:    objectKey,
		ContentType:  req.ContentType,
		SizeBytes:    req.SizeBytes,
		MaxSizeBytes: AfterSalesEvidenceMaxBytes,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}
	ticketID := "rvu_" + shortHash(ticket.ObjectKey)
	ticket.TicketID = ticketID
	if s.reviewImageTickets == nil {
		s.reviewImageTickets = map[string]*ReviewImageUploadTicket{}
	}
	s.reviewImageTickets[ticketID] = &ReviewImageUploadTicket{
		ID:           ticketID,
		UserID:       req.UserID,
		OrderID:      req.OrderID,
		Provider:     ticket.Provider,
		Bucket:       ticket.Bucket,
		ObjectKey:    ticket.ObjectKey,
		PublicURL:    ticket.PublicURL,
		FileName:     req.FileName,
		ContentType:  req.ContentType,
		SizeBytes:    req.SizeBytes,
		MaxSizeBytes: ticket.MaxSizeBytes,
		Status:       AfterSalesUploadTicketIssued,
		ScanStatus:   scanStatus,
		CreatedAt:    now,
		ExpiresAt:    ticket.ExpiresAt,
	}
	return ticket, nil
}

func (s *Store) ConfirmReviewImageUpload(req ConfirmReviewImageUploadRequest) (*ReviewImageUploadTicket, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	req.ContentSHA = strings.TrimSpace(req.ContentSHA)
	if req.UserID == "" || req.TicketID == "" || req.ObjectKey == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}
	if !validReviewImageObjectKey(req.ObjectKey) {
		return nil, ErrInvalidArgument
	}
	if req.FileName == "" {
		req.FileName = sanitizeObjectFileName(objectKeyFileName(req.ObjectKey))
	}
	if req.FileName == "" {
		return nil, ErrInvalidArgument
	}

	ticket, storage, err := s.prepareReviewImageConfirmation(req)
	if err != nil {
		return nil, err
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return ticket, nil
	}
	if err := storage.verifyUploadedObject(objectHeadCheckInput{
		ObjectKey:   ticket.ObjectKey,
		ContentType: ticket.ContentType,
		SizeBytes:   ticket.SizeBytes,
	}); err != nil {
		return nil, err
	}
	return s.commitReviewImageConfirmation(req)
}

func (s *Store) UserReviews(req ReviewListRequest) ([]Review, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.OrderID = strings.TrimSpace(req.OrderID)
	if req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	reviews := make([]Review, 0)
	for _, review := range s.reviews {
		if review.UserID != req.UserID {
			continue
		}
		if req.OrderID != "" && review.OrderID != req.OrderID {
			continue
		}
		reviews = append(reviews, *cloneReview(review))
	}
	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
	})
	return reviews, nil
}

func sanitizeReviewItemRatings(input []ReviewItemRating, order *Order) []ReviewItemRating {
	if len(input) == 0 {
		return nil
	}
	orderItems := map[string]OrderItem{}
	if order != nil {
		for _, item := range order.Items {
			if strings.TrimSpace(item.ProductID) == "" {
				continue
			}
			orderItems[strings.TrimSpace(item.ProductID)] = item
		}
	}
	output := make([]ReviewItemRating, 0, len(input))
	seen := map[string]bool{}
	for _, item := range input {
		productID := strings.TrimSpace(item.ProductID)
		productName := strings.TrimSpace(item.ProductName)
		if orderItem, ok := orderItems[productID]; ok {
			if productName == "" {
				productName = strings.TrimSpace(orderItem.ProductName)
			}
		}
		if productID == "" && productName == "" {
			continue
		}
		key := productID
		if key == "" {
			key = productName
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		rating := item.Rating
		if rating < 1 {
			rating = 1
		}
		if rating > 5 {
			rating = 5
		}
		output = append(output, ReviewItemRating{
			ProductID:   productID,
			ProductName: productName,
			Rating:      rating,
			Tags:        sanitizedStringSlice(item.Tags),
		})
		if len(output) >= 20 {
			break
		}
	}
	return output
}

func (s *Store) CompensateOrderState(req CompensateOrderStateRequest) (*CompensateOrderStateResult, error) {
	normalized, err := normalizeCompensateOrderStateRequest(req)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.compensateOrderStateLocked(normalized)
}

func (s *Store) CompensateOrderStateWithAudit(req CompensateOrderStateRequest, audit RecordAuditLogRequest) (*CompensateOrderStateResult, *AuditLog, error) {
	normalized, err := normalizeCompensateOrderStateRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, err
	}
	if log.Action != "admin.order_state.compensated" || log.TargetType != "order" || log.TargetID != normalized.OrderID {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.compensateOrderStateLocked(normalized)
	if err != nil {
		return nil, nil, err
	}
	log.Payload = orderStateCompensationAuditPayload(result)
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return result, cloneAuditLog(log), nil
}

func normalizeCompensateOrderStateRequest(req CompensateOrderStateRequest) (CompensateOrderStateRequest, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	if req.OrderID == "" {
		return req, ErrInvalidArgument
	}
	if req.ActorID == "" {
		req.ActorID = "system"
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	return req, nil
}

type orderStateCompensationPlan struct {
	Result        *CompensateOrderStateResult
	PaymentMethod string
	Event         OrderEvent
}

func (s *Store) compensateOrderStateLocked(req CompensateOrderStateRequest) (*CompensateOrderStateResult, error) {
	order := s.orders[req.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	plan := s.orderStateCompensationPlanLocked(order, req)
	if !plan.Result.Changed {
		return plan.Result, nil
	}
	if plan.Result.PreviousStatus != plan.Result.ExpectedStatus {
		order.Status = plan.Result.ExpectedStatus
	}
	if plan.Result.PreviousRiderID != plan.Result.ExpectedRiderID {
		order.RiderID = plan.Result.ExpectedRiderID
	}
	if plan.PaymentMethod != "" && strings.TrimSpace(order.PaymentMethod) == "" {
		order.PaymentMethod = plan.PaymentMethod
	}
	order.UpdatedAt = req.Now
	s.appendOrderEventLocked(order, plan.Event)
	plan.Result.Order = cloneOrder(order)
	return plan.Result, nil
}

func (s *Store) orderStateCompensationPlanLocked(order *Order, req CompensateOrderStateRequest) orderStateCompensationPlan {
	expectation := s.expectedOrderStateLocked(order)
	if expectation.Status == "" {
		expectation.Status = order.Status
	}
	result := &CompensateOrderStateResult{
		PreviousStatus:  order.Status,
		ExpectedStatus:  expectation.Status,
		PreviousRiderID: order.RiderID,
		ExpectedRiderID: expectation.RiderID,
		Evidence:        append([]string{}, expectation.Evidence...),
	}
	if order.Status == StatusCancelled || order.Status == StatusRefundPending || order.Status == StatusRefunded {
		result.ExpectedStatus = order.Status
		result.ExpectedRiderID = order.RiderID
		result.Evidence = append(result.Evidence, "terminal_status_protected:"+order.Status)
		result.Order = cloneOrder(order)
		return orderStateCompensationPlan{Result: result}
	}

	statusChanged := expectation.Status != "" && expectation.Status != order.Status
	riderChanged := (expectation.RiderID != "" && expectation.RiderID != order.RiderID && stateKeepsRider(expectation.Status)) ||
		(!stateKeepsRider(expectation.Status) && order.RiderID != "")
	paymentMethodChanged := expectation.PaymentMethod != "" && strings.TrimSpace(order.PaymentMethod) == ""
	if !statusChanged && !riderChanged && !paymentMethodChanged {
		result.Order = cloneOrder(order)
		return orderStateCompensationPlan{Result: result}
	}

	result.Changed = true
	result.CompensationType = "order_state_replay"
	return orderStateCompensationPlan{
		Result:        result,
		PaymentMethod: expectation.PaymentMethod,
		Event: OrderEvent{
			Type:      "order.state.compensated",
			ActorID:   req.ActorID,
			Message:   fmt.Sprintf("订单状态机补偿：从 %s 修复为 %s", result.PreviousStatus, expectation.Status),
			CreatedAt: req.Now,
		},
	}
}

func orderStateCompensationAuditPayload(result *CompensateOrderStateResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"changed":           result.Changed,
		"previous_status":   strings.TrimSpace(result.PreviousStatus),
		"expected_status":   strings.TrimSpace(result.ExpectedStatus),
		"compensation_type": strings.TrimSpace(result.CompensationType),
		"evidence_count":    len(result.Evidence),
	}
	if strings.TrimSpace(result.PreviousRiderID) != "" {
		payload["previous_rider_id"] = strings.TrimSpace(result.PreviousRiderID)
	}
	if strings.TrimSpace(result.ExpectedRiderID) != "" {
		payload["expected_rider_id"] = strings.TrimSpace(result.ExpectedRiderID)
	}
	return payload
}

func (s *Store) MerchantOrders(merchantID string) ([]Order, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	orders := make([]Order, 0)
	for _, order := range s.orders {
		if order.ShopID == "" {
			continue
		}
		shop := s.shops[order.ShopID]
		if shop != nil && shop.MerchantID == merchantID {
			orders = append(orders, *cloneOrder(order))
		}
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

func (s *Store) MerchantAcceptOrder(orderID string, merchantID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	merchantID = strings.TrimSpace(merchantID)
	if orderID == "" || merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if !s.orderBelongsToMerchantLocked(order, merchantID) || order.Status != StatusMerchantPending || !s.shopCanAcceptOrdersLocked(order.ShopID) {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	order.Status = StatusPreparing
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "merchant.accepted",
		ActorID:   merchantID,
		Message:   "商户已接单，开始备餐",
		CreatedAt: now,
	})
	return cloneOrder(order), nil
}

func (s *Store) MerchantMarkOrderReady(orderID string, merchantID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	merchantID = strings.TrimSpace(merchantID)
	if orderID == "" || merchantID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if !s.orderBelongsToMerchantLocked(order, merchantID) || order.Status != StatusPreparing {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	order.Status = StatusDispatching
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "merchant.ready_for_pickup",
		ActorID:   merchantID,
		Message:   "商户已出餐，订单进入骑手调度",
		CreatedAt: now,
	})
	return cloneOrder(order), nil
}

func (s *Store) CheckoutCart(req CheckoutCartRequest) (*Order, *CartSummary, error) {
	userID := strings.TrimSpace(req.UserID)
	shopID := strings.TrimSpace(req.ShopID)
	addressID := strings.TrimSpace(req.AddressID)
	if userID == "" || shopID == "" || addressID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	shop := s.shops[shopID]
	if shop == nil {
		return nil, nil, ErrNotFound
	}
	if !s.shopCanAcceptOrdersLocked(shopID) {
		return nil, nil, ErrInvalidOrderState
	}
	address := s.findAddressLocked(userID, addressID)
	if address == nil || !UserAddressReady(*address) {
		return nil, nil, ErrInvalidArgument
	}
	summary := s.cartSummaryLocked(userID, shopID)
	if len(summary.Items) == 0 || summary.PayableFen <= 0 {
		return nil, nil, ErrInvalidArgument
	}

	s.nextOrderID++
	now := time.Now().UTC()
	orderItems := make([]OrderItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		orderItems = append(orderItems, OrderItem{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ImageURL:     item.ImageURL,
			UnitPriceFen: item.UnitPriceFen,
			Quantity:     item.Quantity,
		})
	}
	order := &Order{
		ID:              fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:          userID,
		ShopID:          shopID,
		ShopName:        shop.Name,
		AddressID:       addressID,
		AddressSnapshot: orderAddressSnapshot(*address),
		Type:            OrderTypeTakeout,
		Status:          StatusPendingPayment,
		AmountFen:       summary.PayableFen,
		ItemsTotalFen:   summary.ItemsTotalFen,
		DeliveryFeeFen:  summary.DeliveryFeeFen,
		PackagingFeeFen: summary.PackagingFeeFen,
		DiscountFen:     summary.DiscountFen,
		Items:           orderItems,
		Options:         normalizeOrderOptions(req.Options),
		CreatedAt:       now,
		UpdatedAt:       now,
		Events: []OrderEvent{{
			Type:      "order.checkout_created",
			ActorID:   userID,
			Message:   "购物车结算创建订单",
			CreatedAt: now,
		}},
	}
	s.orders[order.ID] = order
	delete(s.cartItems, cartKey(userID, shopID))
	return cloneOrder(order), cloneCartSummary(summary), nil
}

func (s *Store) CreditWallet(req CreditWalletRequest) (*WalletTransaction, *WalletAccount, error) {
	userID := strings.TrimSpace(req.UserID)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if userID == "" || req.AmountFen <= 0 || idempotencyKey == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.walletIdempotency[idempotencyKey]; existing != nil {
		return cloneWalletTransaction(existing), cloneWalletAccount(s.wallets[userID]), nil
	}

	account := s.getOrCreateWalletLocked(userID)
	account.Balance += req.AmountFen
	account.Version++

	transaction := s.createWalletTransactionLocked(userID, "", "credit", req.AmountFen, "external_recharge", idempotencyKey)
	s.walletIdempotency[idempotencyKey] = transaction
	return cloneWalletTransaction(transaction), cloneWalletAccount(account), nil
}

func (s *Store) WalletTransactions(userID string) ([]WalletTransaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	transactions := make([]WalletTransaction, 0)
	for _, transaction := range s.walletIdempotency {
		if transaction.UserID == userID {
			transactions = append(transactions, *cloneWalletTransaction(transaction))
		}
	}
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})
	return transactions, nil
}

func (s *Store) CreateFeedback(ticket FeedbackTicket) (*FeedbackTicket, error) {
	ticket.UserID = strings.TrimSpace(ticket.UserID)
	ticket.Type = strings.TrimSpace(ticket.Type)
	ticket.Content = strings.TrimSpace(ticket.Content)
	ticket.Contact = strings.TrimSpace(ticket.Contact)
	if ticket.Type == "" {
		ticket.Type = "feedback"
	}
	if ticket.UserID == "" || ticket.Content == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.nextFeedbackID++
	ticket.ID = fmt.Sprintf("fb_%d", s.nextFeedbackID)
	ticket.Status = "pending"
	ticket.CreatedAt = now
	ticket.UpdatedAt = now
	s.feedbackTickets[ticket.ID] = cloneFeedbackTicket(&ticket)
	return cloneFeedbackTicket(&ticket), nil
}

func (s *Store) UserFeedbackTickets(userID string) ([]FeedbackTicket, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tickets := make([]FeedbackTicket, 0)
	for _, ticket := range s.feedbackTickets {
		if ticket.UserID == userID {
			tickets = append(tickets, *cloneFeedbackTicket(ticket))
		}
	}
	sort.SliceStable(tickets, func(i, j int) bool {
		return tickets[i].CreatedAt.After(tickets[j].CreatedAt)
	})
	return tickets, nil
}

func (s *Store) CreateServiceTicket(req CreateServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Type = strings.TrimSpace(req.Type)
	req.Category = strings.TrimSpace(req.Category)
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Contact = strings.TrimSpace(req.Contact)
	req.RelatedOrderID = strings.TrimSpace(req.RelatedOrderID)
	req.RelatedOrderTitle = strings.TrimSpace(req.RelatedOrderTitle)
	req.RelatedOrderStatus = strings.TrimSpace(req.RelatedOrderStatus)
	req.Severity = strings.TrimSpace(req.Severity)
	req.Attachments = sanitizedStringSlice(req.Attachments)
	if req.Type == "" {
		req.Type = "delivery"
	}
	if req.Category == "" {
		req.Category = "配送问题"
	}
	if req.Title == "" {
		req.Title = req.Category + " · 预计送达未更新"
	}
	if req.Content == "" {
		req.Content = "骑手到店很久了，预计送达时间一直没变化。"
	}
	if req.RelatedOrderID == "" {
		req.RelatedOrderID = "DD240518001"
	}
	if req.RelatedOrderTitle == "" {
		req.RelatedOrderTitle = "蓝海餐厅 · 招牌牛肉饭等 3 件"
	}
	if req.RelatedOrderStatus == "" {
		req.RelatedOrderStatus = "配送中"
	}
	if req.Severity == "" {
		req.Severity = "较严重"
	}
	if req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	now := time.Now().UTC()
	risk := messageRiskCheck(req.Content, now)
	if risk.State == MessageRiskBlocked {
		return nil, fmt.Errorf("%w: %s", ErrRiskControlRejected, risk.Reason)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextServiceTicketID++
	ticket := &ServiceTicket{
		ID:                 fmt.Sprintf("st_%d", s.nextServiceTicketID),
		UserID:             req.UserID,
		Type:               req.Type,
		Category:           req.Category,
		Title:              req.Title,
		Content:            req.Content,
		Contact:            req.Contact,
		RelatedOrderID:     req.RelatedOrderID,
		RelatedOrderTitle:  req.RelatedOrderTitle,
		RelatedOrderStatus: req.RelatedOrderStatus,
		Severity:           req.Severity,
		Status:             ServiceTicketStatusProcessing,
		SLAStatus:          ServiceTicketSLAStatusNormal,
		Solution:           "继续等待：预计 14:35 前送达；延误补偿：订单完成后发放 ¥5 延误券",
		Attachments:        req.Attachments,
		RiskState:          risk.State,
		RiskReasonCode:     risk.ReasonCode,
		RiskReason:         risk.Reason,
		RiskCheckedAt:      risk.CheckedAt,
		ReplyDueAt:         now.Add(serviceTicketReplySLA(req.Severity, req.Category)),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	s.serviceTickets[ticket.ID] = cloneServiceTicket(ticket)
	seedEvents := []ServiceTicketEvent{
		{TicketID: ticket.ID, ActorID: "system", ActorRole: "system", Title: "已提交", Message: "问题已同步到客服工单", Status: ServiceTicketEventDone, CreatedAt: now},
		{TicketID: ticket.ID, ActorID: "customer_service", ActorRole: "support", Title: "客服已受理", Message: "正在核实商家出餐情况", Status: ServiceTicketEventDone, CreatedAt: now.Add(time.Minute)},
		{TicketID: ticket.ID, ActorID: "merchant_1", ActorRole: "merchant", Title: "商家反馈", Message: "补做菜品，预计 8 分钟后出餐", Status: ServiceTicketEventActive, CreatedAt: now.Add(5 * time.Minute)},
		{TicketID: ticket.ID, ActorID: "system", ActorRole: "system", Title: "结果确认", Message: "送达后可确认处理结果", Status: ServiceTicketEventPending, CreatedAt: now.Add(6 * time.Minute)},
	}
	for index := range seedEvents {
		s.nextServiceTicketEventID++
		seedEvents[index].ID = fmt.Sprintf("ste_%d", s.nextServiceTicketEventID)
		s.serviceTicketEvents[seedEvents[index].ID] = cloneServiceTicketEvent(&seedEvents[index])
	}
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) UserServiceTickets(userID string) ([]ServiceTicket, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	tickets := make([]ServiceTicket, 0)
	for _, ticket := range s.serviceTickets {
		if ticket != nil && ticket.UserID == userID {
			s.syncServiceTicketSLAStatusLocked(ticket, now)
			tickets = append(tickets, *cloneServiceTicket(ticket))
		}
	}
	if len(tickets) == 0 && userID == "user_1" {
		return defaultServiceTickets(userID), nil
	}
	sort.SliceStable(tickets, func(i, j int) bool {
		return tickets[i].UpdatedAt.After(tickets[j].UpdatedAt)
	})
	return tickets, nil
}

func (s *Store) AdminServiceTicketDetail(ticketID string) (*ServiceTicketDetail, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.serviceTickets[ticketID]; ok {
		return s.serviceTicketDetailLocked(ticketID)
	}
	for _, seeded := range defaultServiceTickets("user_1") {
		if seeded.ID == ticketID {
			return defaultServiceTicketDetail(seeded), nil
		}
	}
	return nil, ErrNotFound
}

func (s *Store) ServiceTicketDetail(userID string, ticketID string) (*ServiceTicketDetail, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	if userID == "" || ticketID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[ticketID]
	if ticket == nil {
		for _, seeded := range defaultServiceTickets(userID) {
			if seeded.ID == ticketID {
				return defaultServiceTicketDetail(seeded), nil
			}
		}
		return nil, ErrNotFound
	}
	if ticket.UserID != userID {
		return nil, ErrNotFound
	}
	return s.serviceTicketDetailLocked(ticketID)
}

func (s *Store) AddServiceTicketEvent(req AddServiceTicketEventRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.Title = strings.TrimSpace(req.Title)
	req.Message = strings.TrimSpace(req.Message)
	req.Status = strings.TrimSpace(req.Status)
	req.Attachments = sanitizedStringSlice(req.Attachments)
	if req.ActorRole == "" {
		req.ActorRole = "user"
	}
	if req.Title == "" {
		req.Title = "补充说明"
	}
	if req.Status == "" {
		req.Status = ServiceTicketEventActive
	}
	if req.TicketID == "" || req.ActorID == "" || req.Message == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	risk := messageRiskCheck(req.Message, now)
	if risk.State == MessageRiskBlocked {
		return nil, fmt.Errorf("%w: %s", ErrRiskControlRejected, risk.Reason)
	}
	event := s.appendServiceTicketEventLocked(req.TicketID, req.ActorID, req.ActorRole, req.Title, req.Message, req.Status, req.Attachments, now)
	applyMessageRiskToServiceTicketEvent(event, risk)
	if event != nil {
		s.serviceTicketEvents[event.ID] = cloneServiceTicketEvent(event)
	}
	ticket.UpdatedAt = now
	if ticket.Status == ServiceTicketStatusClosed {
		ticket.Status = ServiceTicketStatusProcessing
	}
	if ticket.Status == ServiceTicketStatusProcessing && ticket.SLAStatus != ServiceTicketSLAStatusEscalated {
		ticket.SLAStatus = ServiceTicketSLAStatusNormal
		ticket.ReplyDueAt = now.Add(serviceTicketReplySLA(ticket.Severity, ticket.Category))
	}
	return s.serviceTicketDetailLocked(req.TicketID)
}

func (s *Store) AdminServiceTickets(req ServiceTicketListRequest) ([]ServiceTicket, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.RelatedOrderID = strings.TrimSpace(req.RelatedOrderID)
	req.Status = strings.TrimSpace(req.Status)
	req.SLAStatus = strings.TrimSpace(req.SLAStatus)
	req.AssignedSupportID = strings.TrimSpace(req.AssignedSupportID)
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tickets := make([]ServiceTicket, 0)
	for _, ticket := range s.serviceTickets {
		if ticket == nil {
			continue
		}
		s.syncServiceTicketSLAStatusLocked(ticket, now)
		if req.UserID != "" && ticket.UserID != req.UserID {
			continue
		}
		if req.RelatedOrderID != "" && ticket.RelatedOrderID != req.RelatedOrderID {
			continue
		}
		if req.Status != "" && ticket.Status != req.Status {
			continue
		}
		if req.SLAStatus != "" && ticket.SLAStatus != req.SLAStatus {
			continue
		}
		if req.AssignedSupportID != "" && ticket.AssignedSupportID != req.AssignedSupportID {
			continue
		}
		tickets = append(tickets, *cloneServiceTicket(ticket))
	}
	if len(tickets) == 0 && req.UserID == "" && req.RelatedOrderID == "" && req.Status == "" && req.SLAStatus == "" && req.AssignedSupportID == "" {
		tickets = defaultServiceTickets("user_1")
	}
	sort.SliceStable(tickets, func(i, j int) bool {
		return tickets[i].UpdatedAt.After(tickets[j].UpdatedAt)
	})
	if len(tickets) > req.Limit {
		tickets = tickets[:req.Limit]
	}
	return tickets, nil
}

func (s *Store) AssignServiceTicket(req AssignServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.SupportID = strings.TrimSpace(req.SupportID)
	req.SupportName = strings.TrimSpace(req.SupportName)
	req.ActorID = strings.TrimSpace(req.ActorID)
	if req.TicketID == "" || req.SupportID == "" {
		return nil, ErrInvalidArgument
	}
	if req.SupportName == "" {
		req.SupportName = "客服专员"
	}
	if req.ActorID == "" {
		req.ActorID = req.SupportID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	if ticket.Status == ServiceTicketStatusClosed || ticket.Status == ServiceTicketStatusResolved {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	ticket.AssignedSupportID = req.SupportID
	ticket.AssignedSupportName = req.SupportName
	ticket.AssignedAt = now
	ticket.Status = ServiceTicketStatusProcessing
	ticket.SLAStatus = ServiceTicketSLAStatusNormal
	ticket.ReplyDueAt = now.Add(serviceTicketReplySLA(ticket.Severity, ticket.Category))
	ticket.UpdatedAt = now
	s.appendServiceTicketEventLocked(ticket.ID, req.ActorID, "support", "已分派客服", "工单已分派给"+req.SupportName, ServiceTicketEventDone, nil, now)
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) ResolveServiceTicket(req ResolveServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.Solution = strings.TrimSpace(req.Solution)
	if req.TicketID == "" || req.Solution == "" {
		return nil, ErrInvalidArgument
	}
	if req.ActorID == "" {
		req.ActorID = "customer_service"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	if ticket.Status == ServiceTicketStatusClosed {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	ticket.Solution = req.Solution
	ticket.Status = ServiceTicketStatusWaitingConfirm
	ticket.SLAStatus = ServiceTicketSLAStatusCompleted
	ticket.ResolvedAt = now
	ticket.UpdatedAt = now
	s.appendServiceTicketEventLocked(ticket.ID, req.ActorID, "support", "处理方案", req.Solution, ServiceTicketEventActive, nil, now)
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) EscalateServiceTicket(req EscalateServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.EscalationLevel = strings.TrimSpace(req.EscalationLevel)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.TicketID == "" {
		return nil, ErrInvalidArgument
	}
	if req.ActorID == "" {
		req.ActorID = "customer_service"
	}
	if req.EscalationLevel == "" {
		req.EscalationLevel = "support_lead"
	}
	if req.Reason == "" {
		req.Reason = "超过 SLA 未更新，升级给客服主管处理"
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	if ticket.Status == ServiceTicketStatusClosed || ticket.Status == ServiceTicketStatusResolved || ticket.Status == ServiceTicketStatusWaitingConfirm {
		return nil, ErrInvalidOrderState
	}
	ticket.SLAStatus = ServiceTicketSLAStatusEscalated
	ticket.EscalationLevel = req.EscalationLevel
	ticket.EscalationReason = req.Reason
	ticket.EscalatedAt = now
	ticket.UpdatedAt = now
	s.appendServiceTicketEventLocked(ticket.ID, req.ActorID, "support", "已升级处理", req.Reason, ServiceTicketEventActive, nil, now)
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) ReviewServiceTicketQuality(req ServiceTicketQualityReviewRequest) (*ServiceTicketQualityReview, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ReviewerID = strings.TrimSpace(req.ReviewerID)
	req.ReviewerName = strings.TrimSpace(req.ReviewerName)
	req.Result = normalizeServiceTicketQualityResult(req.Result, req.Score)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.TicketID == "" || req.Score < 0 || req.Score > 100 || req.Result == "" {
		return nil, ErrInvalidArgument
	}
	if req.ReviewerID == "" {
		req.ReviewerID = "support_quality"
	}
	if req.ReviewerName == "" {
		req.ReviewerName = "客服质检"
	}
	if req.Result != ServiceTicketQualityPassed {
		req.CoachingRequired = true
	}
	if req.Notes == "" {
		req.Notes = "客服工单抽检已完成"
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	s.syncServiceTicketSLAStatusLocked(ticket, now)
	s.nextServiceTicketQualityID++
	review := &ServiceTicketQualityReview{
		ID:                fmt.Sprintf("stq_%d", s.nextServiceTicketQualityID),
		TicketID:          ticket.ID,
		SupportID:         ticket.AssignedSupportID,
		SupportName:       ticket.AssignedSupportName,
		ReviewerID:        req.ReviewerID,
		ReviewerName:      req.ReviewerName,
		Score:             req.Score,
		Result:            req.Result,
		Notes:             req.Notes,
		CoachingRequired:  req.CoachingRequired,
		TicketTitle:       ticket.Title,
		TicketCategory:    ticket.Category,
		TicketSLAStatus:   ticket.SLAStatus,
		TicketFollowUp:    ticket.FollowUpRating,
		TicketResolvedAt:  ticket.ResolvedAt,
		TicketEscalatedAt: ticket.EscalatedAt,
		CreatedAt:         now,
	}
	s.serviceTicketQualityReviews[review.ID] = cloneServiceTicketQualityReview(review)
	ticket.UpdatedAt = now
	s.appendServiceTicketEventLocked(ticket.ID, req.ReviewerID, "quality", "质检抽检", serviceTicketQualityMessage(review), ServiceTicketEventDone, nil, now)
	return cloneServiceTicketQualityReview(review), nil
}

func (s *Store) ServiceTicketQualityReviews(req ServiceTicketQualityReviewListRequest) ([]ServiceTicketQualityReview, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.SupportID = strings.TrimSpace(req.SupportID)
	req.ReviewerID = strings.TrimSpace(req.ReviewerID)
	req.Result = strings.TrimSpace(req.Result)
	req.CoachingRequired = strings.TrimSpace(req.CoachingRequired)
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	reviews := make([]ServiceTicketQualityReview, 0)
	for _, review := range s.serviceTicketQualityReviews {
		if review == nil {
			continue
		}
		if req.TicketID != "" && review.TicketID != req.TicketID {
			continue
		}
		if req.SupportID != "" && review.SupportID != req.SupportID {
			continue
		}
		if req.ReviewerID != "" && review.ReviewerID != req.ReviewerID {
			continue
		}
		if req.Result != "" && review.Result != req.Result {
			continue
		}
		if req.CoachingRequired != "" && boolQueryValue(req.CoachingRequired) != review.CoachingRequired {
			continue
		}
		reviews = append(reviews, *cloneServiceTicketQualityReview(review))
	}
	if len(reviews) == 0 && req.TicketID == "" && req.SupportID == "" && req.ReviewerID == "" && req.Result == "" && req.CoachingRequired == "" {
		reviews = defaultServiceTicketQualityReviews()
	}
	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
	})
	if len(reviews) > req.Limit {
		reviews = reviews[:req.Limit]
	}
	return reviews, nil
}

func (s *Store) ServiceTicketPerformance(req ServiceTicketPerformanceRequest) ([]ServiceTicketPerformanceSummary, error) {
	req.SupportID = strings.TrimSpace(req.SupportID)
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tickets := make([]ServiceTicket, 0, len(s.serviceTickets))
	for _, ticket := range s.serviceTickets {
		if ticket == nil {
			continue
		}
		s.syncServiceTicketSLAStatusLocked(ticket, now)
		tickets = append(tickets, *cloneServiceTicket(ticket))
	}
	if len(tickets) == 0 {
		return filterServiceTicketPerformance(defaultServiceTicketPerformanceSummaries(), req.SupportID, req.Limit), nil
	}
	summaries := buildServiceTicketPerformanceSummaries(tickets, s.serviceTicketQualityReviews, req.SupportID)
	if len(summaries) == 0 && req.SupportID == "" {
		summaries = defaultServiceTicketPerformanceSummaries()
	}
	return filterServiceTicketPerformance(summaries, req.SupportID, req.Limit), nil
}

func (s *Store) CloseServiceTicket(req CloseServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.TicketID == "" || req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	if req.ActorID == "" {
		req.ActorID = req.UserID
	}
	if req.Reason == "" {
		req.Reason = "用户确认处理结果"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	if ticket.UserID != req.UserID {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	ticket.Status = ServiceTicketStatusClosed
	ticket.SLAStatus = ServiceTicketSLAStatusCompleted
	ticket.ClosedAt = now
	ticket.UpdatedAt = now
	s.appendServiceTicketEventLocked(ticket.ID, req.ActorID, "user", "用户确认", req.Reason, ServiceTicketEventDone, nil, now)
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) FollowUpServiceTicket(req FollowUpServiceTicketRequest) (*ServiceTicketDetail, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Comment = strings.TrimSpace(req.Comment)
	if req.TicketID == "" || req.UserID == "" || req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.serviceTickets[req.TicketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	if ticket.UserID != req.UserID {
		return nil, ErrNotFound
	}
	if ticket.Status != ServiceTicketStatusClosed && ticket.Status != ServiceTicketStatusResolved {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	ticket.FollowUpRating = req.Rating
	ticket.FollowUpComment = req.Comment
	ticket.FollowUpAt = now
	ticket.UpdatedAt = now
	message := fmt.Sprintf("用户给本次服务评分 %d 分", req.Rating)
	if req.Comment != "" {
		message += "：" + req.Comment
	}
	s.appendServiceTicketEventLocked(ticket.ID, req.UserID, "user", "回访评价", message, ServiceTicketEventDone, nil, now)
	return s.serviceTicketDetailLocked(ticket.ID)
}

func (s *Store) UserProfileOverview(userID string) (*UserProfileOverview, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	coupons := s.userCouponsForUserLocked(userID)
	stats := map[string]int{"pending_payment": 0, "in_progress": 0, "pending_review": 0, "after_sales": 0}
	for _, order := range s.orders {
		if order == nil || order.UserID != userID {
			continue
		}
		switch order.Status {
		case StatusPendingPayment:
			stats["pending_payment"]++
		case StatusCompleted:
			stats["pending_review"]++
		case StatusRefundPending, StatusRefunded:
			stats["after_sales"]++
		default:
			stats["in_progress"]++
		}
	}
	for _, request := range s.afterSalesRequests {
		if request != nil && request.UserID == userID {
			stats["after_sales"]++
		}
	}
	return &UserProfileOverview{
		UserID:                userID,
		Nickname:              s.nicknameForUserLocked(userID),
		Phone:                 s.phoneForUserLocked(userID),
		AvatarInitial:         avatarInitial(s.nicknameForUserLocked(userID)),
		MembershipLevel:       MembershipSilver,
		MembershipTitle:       "美食达人",
		CreditText:            "信用良好",
		Verified:              true,
		GrowthValue:           2680,
		NextLevelGrowth:       1320,
		SavingsFen:            3600,
		WalletBalanceFen:      s.walletBalanceForOverviewLocked(userID),
		PendingReceivableFen:  s.pendingReceivableFenForUserLocked(userID),
		CouponCount:           len(coupons),
		RedPacketCount:        s.redPacketCountForUserLocked(userID),
		Points:                s.pointsBalanceForUserLocked(userID),
		FavoriteShopCount:     4,
		PaymentPasswordStatus: s.paymentPasswordStatusForUserLocked(userID),
		OrderStats:            stats,
	}, nil
}

func (s *Store) CirclePosts(userID string) ([]CirclePost, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	posts := make([]CirclePost, 0)
	for _, post := range s.circlePosts {
		if post == nil || post.Status != CirclePostPublished {
			continue
		}
		posts = append(posts, *cloneCirclePost(post))
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})
	return posts, nil
}

func (s *Store) CreateCirclePost(post CirclePost) (*CirclePost, error) {
	post.AuthorUserID = strings.TrimSpace(post.AuthorUserID)
	post.AuthorName = strings.TrimSpace(post.AuthorName)
	post.CircleID = strings.TrimSpace(post.CircleID)
	post.Type = strings.TrimSpace(post.Type)
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)
	post.ImageURLs = sanitizedStringSlice(post.ImageURLs)
	post.Tags = sanitizedStringSlice(post.Tags)
	if post.CircleID == "" {
		post.CircleID = "nearby"
	}
	if post.Type == "" {
		post.Type = CirclePostText
	}
	if post.AuthorName == "" {
		post.AuthorName = "我"
	}
	if post.AuthorUserID == "" || post.Content == "" {
		return nil, ErrInvalidArgument
	}
	switch post.Type {
	case CirclePostText, CirclePostImage, CirclePostFoodInvite:
	default:
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextCirclePostID++
	post.ID = fmt.Sprintf("cpost_%d", s.nextCirclePostID)
	post.Status = CirclePostPublished
	post.CreatedAt = time.Now().UTC()
	s.circlePosts[post.ID] = cloneCirclePost(&post)
	return cloneCirclePost(&post), nil
}

func (s *Store) SaveMealMatchProfile(profile MealMatchProfile) (*MealMatchProfile, error) {
	profile.UserID = strings.TrimSpace(profile.UserID)
	profile.Gender = strings.TrimSpace(profile.Gender)
	profile.SchoolID = strings.TrimSpace(profile.SchoolID)
	profile.SchoolName = strings.TrimSpace(profile.SchoolName)
	profile.CampusName = strings.TrimSpace(profile.CampusName)
	profile.BuildingID = strings.TrimSpace(profile.BuildingID)
	profile.BuildingName = strings.TrimSpace(profile.BuildingName)
	profile.PrivacyScope = normalizeMealMatchPrivacyScope(profile.PrivacyScope)
	profile.LocationPrecision = normalizeMealMatchLocationPrecision(profile.LocationPrecision)
	profile.DeviceID = strings.TrimSpace(profile.DeviceID)
	profile.PersonalityTraits = sanitizedStringSlice(profile.PersonalityTraits)
	profile.DietaryHabits = sanitizedStringSlice(profile.DietaryHabits)
	if profile.UserID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	profile.SchoolName = mealMatchSchoolName(profile.SchoolID, profile.SchoolName)
	profile.CampusName = mealMatchCampusName(profile.SchoolID, profile.CampusName)
	deviceRisk := s.mealMatchDeviceRiskLocked(profile.UserID, profile.DeviceID, now)
	if deviceRisk.State == MealMatchDeviceRiskBlocked {
		return nil, fmt.Errorf("%w: %s", ErrRiskControlRejected, deviceRisk.Reason)
	}
	applyMealMatchDeviceRiskToProfile(&profile, deviceRisk)
	if ok, _ := mealMatchProfilePrerequisites(profile); ok {
		profile.ModerationStatus = MealMatchModerationPending
		profile.ModerationReason = "资料已提交，等待平台人工审核"
		if deviceRisk.State == MealMatchDeviceRiskReview {
			profile.ModerationReason = deviceRisk.Reason
		}
		profile.ModerationReviewerID = ""
		profile.ModerationReviewedAt = ""
		record := s.createMealMatchProfileReviewLocked(profile.UserID, now)
		profile.ModerationRecordID = record.ID
	} else {
		profile.ModerationStatus = ""
		profile.ModerationReason = ""
		profile.ModerationRecordID = ""
		profile.ModerationReviewerID = ""
		profile.ModerationReviewedAt = ""
	}
	s.mealMatchProfiles[profile.UserID] = cloneMealMatchProfile(&profile)
	return cloneMealMatchProfile(&profile), nil
}

func (s *Store) UserMealMatchProfile(userID string) (*MealMatchProfile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile := s.mealMatchProfiles[userID]
	if profile == nil {
		return &MealMatchProfile{UserID: userID}, nil
	}
	return cloneMealMatchProfile(profile), nil
}

func (s *Store) MealMatchCandidates(userID string) (*MealMatchCandidateList, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile := s.mealMatchProfiles[userID]
	if profile == nil {
		profile = &MealMatchProfile{UserID: userID}
	}
	canUse, missing := CanUseMealMatch(*profile)
	result := &MealMatchCandidateList{
		Profile:              *cloneMealMatchProfile(profile),
		CanUse:               canUse,
		Missing:              append([]string{}, missing...),
		PrivacyScope:         normalizeMealMatchPrivacyScope(profile.PrivacyScope),
		LocationPrecision:    normalizeMealMatchLocationPrecision(profile.LocationPrecision),
		PrivacyNotice:        mealMatchProfilePrivacyNotice(*profile),
		DeviceRiskState:      profile.DeviceRiskState,
		DeviceRiskReasonCode: profile.DeviceRiskReasonCode,
		DeviceRiskReason:     profile.DeviceRiskReason,
		DeviceRiskCheckedAt:  profile.DeviceRiskCheckedAt,
		ModerationStatus:     normalizeMealMatchModerationStatus(profile.ModerationStatus),
		ReviewRequired:       mealMatchMissingIncludes(missing, "moderation_pending") || mealMatchMissingIncludes(missing, "moderation_rejected") || mealMatchMissingIncludes(missing, "device_risk_review"),
	}
	if !canUse {
		result.Candidates = []MealMatchCandidate{}
		return result, nil
	}
	candidates := []MealMatchCandidate{}
	for candidateUserID, candidateProfile := range s.mealMatchProfiles {
		if candidateProfile == nil || candidateUserID == userID {
			continue
		}
		if s.mealMatchBlockedLocked(userID, candidateUserID) || s.mealMatchBlockedLocked(candidateUserID, userID) {
			continue
		}
		if ok, _ := CanUseMealMatch(*candidateProfile); !ok {
			continue
		}
		if !mealMatchCanShowCandidate(*profile, *candidateProfile) {
			continue
		}
		candidate := mealMatchCandidateFromProfiles(*profile, *candidateProfile, s.nicknameForUserLocked(candidateUserID))
		if candidate.MatchScore <= 0 {
			candidate.MatchScore = 20
		}
		candidates = append(candidates, candidate)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].MatchScore == candidates[j].MatchScore {
			return candidates[i].UserID < candidates[j].UserID
		}
		return candidates[i].MatchScore > candidates[j].MatchScore
	})
	if len(candidates) > 6 {
		candidates = candidates[:6]
	}
	result.Candidates = candidates
	return result, nil
}

func (s *Store) ReportMealMatchCandidate(req MealMatchReportRequest) (*MealMatchModerationRecord, error) {
	req.ReporterUserID = strings.TrimSpace(req.ReporterUserID)
	req.TargetUserID = strings.TrimSpace(req.TargetUserID)
	req.Reason = strings.TrimSpace(req.Reason)
	req.Description = strings.TrimSpace(req.Description)
	if req.ReporterUserID == "" || req.TargetUserID == "" || req.ReporterUserID == req.TargetUserID || req.Reason == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mealMatchProfiles[req.TargetUserID]; !ok {
		return nil, ErrNotFound
	}
	s.nextMealMatchModerationID++
	record := &MealMatchModerationRecord{
		ID:           fmt.Sprintf("mmod_%d", s.nextMealMatchModerationID),
		UserID:       req.ReporterUserID,
		TargetUserID: req.TargetUserID,
		Action:       MealMatchModerationReported,
		Reason:       req.Reason,
		Description:  req.Description,
		Status:       MealMatchModerationPending,
		CreatedAt:    time.Now().UTC(),
	}
	s.mealMatchModeration[record.ID] = cloneMealMatchModerationRecord(record)
	return cloneMealMatchModerationRecord(record), nil
}

func (s *Store) BlockMealMatchCandidate(req MealMatchBlockRequest) (*MealMatchModerationRecord, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.TargetUserID = strings.TrimSpace(req.TargetUserID)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.UserID == "" || req.TargetUserID == "" || req.UserID == req.TargetUserID {
		return nil, ErrInvalidArgument
	}
	if req.Reason == "" {
		req.Reason = "not_interested"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mealMatchProfiles[req.TargetUserID]; !ok {
		return nil, ErrNotFound
	}
	key := mealMatchBlockKey(req.UserID, req.TargetUserID)
	if existing := s.mealMatchModeration[key]; existing != nil {
		return cloneMealMatchModerationRecord(existing), nil
	}
	record := &MealMatchModerationRecord{
		ID:           key,
		UserID:       req.UserID,
		TargetUserID: req.TargetUserID,
		Action:       MealMatchModerationBlocked,
		Reason:       req.Reason,
		Status:       MealMatchModerationActive,
		CreatedAt:    time.Now().UTC(),
	}
	s.mealMatchModeration[record.ID] = cloneMealMatchModerationRecord(record)
	return cloneMealMatchModerationRecord(record), nil
}

func (s *Store) AdminMealMatchModerationRecords(req MealMatchModerationListRequest) ([]MealMatchModerationRecord, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.Action = strings.TrimSpace(req.Action)
	req.UserID = strings.TrimSpace(req.UserID)
	req.TargetUserID = strings.TrimSpace(req.TargetUserID)
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	records := make([]MealMatchModerationRecord, 0, len(s.mealMatchModeration))
	for _, record := range s.mealMatchModeration {
		if record == nil {
			continue
		}
		if req.Status != "" && record.Status != req.Status {
			continue
		}
		if req.Action != "" && record.Action != req.Action {
			continue
		}
		if req.UserID != "" && record.UserID != req.UserID {
			continue
		}
		if req.TargetUserID != "" && record.TargetUserID != req.TargetUserID {
			continue
		}
		records = append(records, *cloneMealMatchModerationRecord(record))
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	if len(records) > req.Limit {
		records = records[:req.Limit]
	}
	return records, nil
}

func (s *Store) ReviewMealMatchModeration(req MealMatchModerationReviewRequest) (*MealMatchModerationRecord, error) {
	req.RecordID = strings.TrimSpace(req.RecordID)
	req.Decision = normalizeMealMatchModerationDecision(req.Decision)
	req.ReviewerID = strings.TrimSpace(req.ReviewerID)
	req.ReviewNote = strings.TrimSpace(req.ReviewNote)
	if req.RecordID == "" || req.Decision == "" || req.ReviewerID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.mealMatchModeration[req.RecordID]
	if record == nil {
		return nil, ErrNotFound
	}
	if record.Status != MealMatchModerationPending {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	record.Status = req.Decision
	record.ReviewerID = req.ReviewerID
	record.ReviewNote = req.ReviewNote
	record.ReviewedAt = now
	switch record.Action {
	case MealMatchModerationProfileReview:
		profile := s.mealMatchProfiles[record.UserID]
		if profile == nil {
			return nil, ErrNotFound
		}
		profile.ModerationStatus = req.Decision
		profile.ModerationReviewerID = req.ReviewerID
		profile.ModerationReviewedAt = now.Format(time.RFC3339)
		if req.Decision == MealMatchModerationApproved {
			profile.ModerationReason = defaultString(req.ReviewNote, "资料审核通过")
			if normalizeMealMatchDeviceRiskState(profile.DeviceRiskState) == MealMatchDeviceRiskReview {
				profile.DeviceRiskState = MealMatchDeviceRiskPassed
				profile.DeviceRiskReasonCode = ""
				profile.DeviceRiskReason = "设备环境已由人工复核通过"
				profile.DeviceRiskCheckedAt = now
			}
		} else {
			profile.ModerationReason = defaultString(req.ReviewNote, "资料审核未通过，请修改后重新提交")
		}
	case MealMatchModerationReported:
		if req.Decision == MealMatchModerationApproved {
			if target := s.mealMatchProfiles[record.TargetUserID]; target != nil {
				target.ModerationStatus = MealMatchModerationRejected
				target.ModerationReason = defaultString(req.ReviewNote, "举报成立，资料已暂停展示")
				target.ModerationReviewerID = req.ReviewerID
				target.ModerationReviewedAt = now.Format(time.RFC3339)
			}
		}
	}
	s.mealMatchModeration[record.ID] = cloneMealMatchModerationRecord(record)
	return cloneMealMatchModerationRecord(record), nil
}

func (s *Store) CreateRedPacket(packet RedPacket) (*RedPacketDetail, error) {
	packet.SenderID = strings.TrimSpace(packet.SenderID)
	packet.SenderRole = strings.TrimSpace(packet.SenderRole)
	packet.Scene = strings.TrimSpace(packet.Scene)
	packet.TargetID = strings.TrimSpace(packet.TargetID)
	packet.Type = strings.TrimSpace(packet.Type)
	packet.PaymentMethod = strings.TrimSpace(packet.PaymentMethod)
	packet.Message = strings.TrimSpace(packet.Message)
	if packet.SenderRole == "" {
		packet.SenderRole = "user"
	}
	if packet.Scene == "" {
		packet.Scene = RedPacketSceneGroupChat
	}
	if packet.TargetID == "" {
		packet.TargetID = "group_chat_1"
	}
	if packet.Type == "" {
		packet.Type = RedPacketTypeRandom
	}
	if packet.PaymentMethod == "" {
		packet.PaymentMethod = PaymentBalance
	}
	if packet.SenderID == "" || packet.TotalAmountFen <= 0 || packet.Quantity <= 0 || packet.Quantity > 100 {
		return nil, ErrInvalidArgument
	}
	if packet.TotalAmountFen < int64(packet.Quantity) {
		return nil, ErrInvalidArgument
	}
	switch packet.Scene {
	case RedPacketSceneGroupChat, RedPacketSceneDirectMessage:
	default:
		return nil, ErrInvalidArgument
	}
	switch packet.Type {
	case RedPacketTypeFixed, RedPacketTypeRandom:
	default:
		return nil, ErrInvalidArgument
	}
	if packet.PaymentMethod != PaymentBalance {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.nextRedPacketID++
	packet.ID = fmt.Sprintf("rp_%d", s.nextRedPacketID)
	packet.Status = RedPacketStatusCreated
	packet.CreatedAt = now
	if packet.ExpiresAt.IsZero() {
		packet.ExpiresAt = now.Add(24 * time.Hour)
	}
	account := s.getOrCreateWalletLocked(packet.SenderID)
	if account.Balance == 0 && packet.SenderID == "user_1" {
		account.Balance = 12850
		account.Version++
	}
	if account.Balance < packet.TotalAmountFen {
		return nil, ErrInsufficientBalance
	}
	account.Balance -= packet.TotalAmountFen
	account.Frozen += packet.TotalAmountFen
	account.Version++
	freezeTransaction := s.createWalletTransactionLocked(packet.SenderID, packet.ID, "red_packet_freeze", -packet.TotalAmountFen, packet.PaymentMethod, redPacketFreezeKey(packet.ID))
	freezeTransaction.Status = "frozen"
	s.walletIdempotency[freezeTransaction.IdempotencyKey] = freezeTransaction
	detail := &RedPacketDetail{
		Packet: packet,
		Shares: redPacketShares(packet, now),
	}
	s.redPackets[packet.ID] = cloneRedPacketDetail(detail)
	return cloneRedPacketDetail(detail), nil
}

func (s *Store) RedPacketDetail(packetID string, userID string) (*RedPacketDetail, error) {
	packetID = strings.TrimSpace(packetID)
	if packetID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	detail := s.redPackets[packetID]
	if detail == nil {
		return nil, ErrNotFound
	}
	if err := s.autoRefundExpiredRedPacketLocked(detail, time.Now().UTC()); err != nil {
		return nil, err
	}
	output := cloneRedPacketDetail(detail)
	if strings.TrimSpace(userID) != "" {
		output.Risk = s.redPacketClaimRiskLocked(detail, strings.TrimSpace(userID), redPacketNextClaimAmount(detail), time.Now().UTC())
	}
	return output, nil
}

func (s *Store) ClaimRedPacket(packetID string, userID string) (*RedPacketClaimResult, error) {
	packetID = strings.TrimSpace(packetID)
	userID = strings.TrimSpace(userID)
	if packetID == "" || userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	detail := s.redPackets[packetID]
	if detail == nil {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	if err := s.autoRefundExpiredRedPacketLocked(detail, now); err != nil {
		return nil, err
	}
	if detail.Packet.Status == RedPacketStatusRefunded || detail.Packet.Status == RedPacketStatusExpired {
		return nil, ErrInvalidOrderState
	}
	claimIndex := -1
	for index := range detail.Shares {
		if detail.Shares[index].UserID == userID {
			claimIndex = index
			break
		}
	}
	if claimIndex >= 0 && !detail.Shares[claimIndex].ClaimedAt.IsZero() {
		share := detail.Shares[claimIndex]
		risk := &RedPacketRiskCheck{
			State:     RedPacketRiskPassed,
			Reason:    "已领取过该红包，本次返回原领取记录",
			CheckedAt: now,
		}
		output := *cloneRedPacketDetail(detail)
		output.Risk = risk
		return &RedPacketClaimResult{
			Detail: output,
			Share:  share,
			Risk:   risk,
		}, nil
	}
	if claimIndex == -1 {
		for index := range detail.Shares {
			if detail.Shares[index].UserID == "" && detail.Shares[index].ClaimedAt.IsZero() {
				claimIndex = index
				break
			}
		}
	}
	if claimIndex == -1 {
		return nil, ErrInvalidOrderState
	}
	risk := s.redPacketClaimRiskLocked(detail, userID, detail.Shares[claimIndex].AmountFen, now)
	if risk != nil && risk.State == RedPacketRiskBlocked {
		return nil, fmt.Errorf("%w: %s", ErrRiskControlRejected, risk.Reason)
	}
	detail.Shares[claimIndex].UserID = userID
	if detail.Shares[claimIndex].ClaimedAt.IsZero() {
		detail.Shares[claimIndex].ClaimedAt = now
	}
	share := detail.Shares[claimIndex]
	claimKey := redPacketClaimKey(packetID, userID)
	if s.walletIdempotency[claimKey] == nil {
		senderAccount := s.getOrCreateWalletLocked(detail.Packet.SenderID)
		if senderAccount.Frozen < share.AmountFen {
			return nil, ErrInvalidOrderState
		}
		senderAccount.Frozen -= share.AmountFen
		senderAccount.Version++
		receiverAccount := s.getOrCreateWalletLocked(userID)
		receiverAccount.Balance += share.AmountFen
		receiverAccount.Version++
		transaction := s.createWalletTransactionLocked(userID, packetID, "red_packet_claim", share.AmountFen, PaymentBalance, claimKey)
		s.walletIdempotency[transaction.IdempotencyKey] = transaction
	}
	detail.Packet.ClaimedAmountFen = redPacketClaimedAmount(detail)
	if redPacketClaimedCount(detail) >= detail.Packet.Quantity {
		detail.Packet.Status = RedPacketStatusFinished
	}
	output := *cloneRedPacketDetail(detail)
	output.Risk = risk
	return &RedPacketClaimResult{
		Detail: output,
		Share:  share,
		Risk:   risk,
	}, nil
}

func (s *Store) RefundRedPacket(packetID string, userID string) (*RedPacketDetail, error) {
	packetID = strings.TrimSpace(packetID)
	userID = strings.TrimSpace(userID)
	if packetID == "" || userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	detail := s.redPackets[packetID]
	if detail == nil {
		return nil, ErrNotFound
	}
	if detail.Packet.SenderID != userID {
		return nil, ErrInvalidOrderState
	}
	if err := s.refundRedPacketRemainderLocked(detail, time.Now().UTC(), RedPacketStatusRefunded); err != nil {
		return nil, err
	}
	return cloneRedPacketDetail(detail), nil
}

func (s *Store) AutoRefundExpiredRedPackets(now time.Time) ([]RedPacketDetail, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	refunded := make([]RedPacketDetail, 0)
	for _, detail := range s.redPackets {
		if detail == nil || detail.Packet.Status != RedPacketStatusCreated {
			continue
		}
		if detail.Packet.ExpiresAt.IsZero() || now.Before(detail.Packet.ExpiresAt) {
			continue
		}
		if err := s.refundRedPacketRemainderLocked(detail, now, RedPacketStatusExpired); err != nil {
			return nil, err
		}
		refunded = append(refunded, *cloneRedPacketDetail(detail))
	}
	return refunded, nil
}

func (s *Store) MessageThreads(userID string) ([]ChatThread, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads := defaultChatThreads()
	visible := make([]ChatThread, 0, len(threads))
	for index := range threads {
		member := s.chatThreadMemberLocked(threads[index].ID, "user", userID)
		if member == nil {
			continue
		}
		if latest := s.latestMessageForThreadLocked(threads[index].ID); latest != nil {
			threads[index].Subtitle = latest.Content
			threads[index].LastMessageID = latest.ID
			threads[index].UpdatedAt = latest.CreatedAt
		}
		if read := s.chatReadStateLocked(userID, threads[index].ID); read != nil {
			threads[index].LastReadMessageID = read.LastReadMessageID
			threads[index].LastReadAt = read.ReadAt
		}
		threads[index].Muted = member.Muted
		threads[index].UnreadCount = s.unreadChatMessageCountLocked(userID, &threads[index])
		visible = append(visible, threads[index])
	}
	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].UpdatedAt.After(visible[j].UpdatedAt)
	})
	return visible, nil
}

func (s *Store) ChatThreadOverview(userID string, threadID string) (*ChatThreadOverview, error) {
	userID = strings.TrimSpace(userID)
	threadID = strings.TrimSpace(threadID)
	if userID == "" || threadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(threadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	member := s.chatThreadMemberLocked(threadID, "user", userID)
	if member == nil {
		return nil, ErrNotFound
	}
	thread := chatThreadByID(threadID)
	if thread == nil {
		return nil, ErrNotFound
	}
	if latest := s.latestMessageForThreadLocked(thread.ID); latest != nil {
		thread.LastMessageID = latest.ID
		thread.UpdatedAt = latest.CreatedAt
	}
	memberCount := s.chatThreadOverviewMemberCountLocked(threadID)
	thread.Muted = member.Muted
	thread.UnreadCount = s.unreadChatMessageCountLocked(userID, thread)
	seed := chatThreadOverviewSeed(threadID)
	profiles := s.chatThreadMemberProfilesLocked(threadID, userID)
	preview := profiles
	if len(preview) > 5 {
		preview = append([]ChatThreadMemberProfile{}, preview[:5]...)
	}
	return &ChatThreadOverview{
		ThreadID:      thread.ID,
		Type:          thread.Type,
		Title:         thread.Title,
		Icon:          thread.Icon,
		Summary:       chatThreadOverviewSummary(threadID, memberCount),
		Announcement:  seed.Announcement,
		SettingsText:  seed.SettingsText,
		MemberCount:   memberCount,
		Muted:         member.Muted,
		MemberPreview: preview,
	}, nil
}

func (s *Store) ChatThreadMembers(userID string, threadID string) ([]ChatThreadMemberProfile, error) {
	userID = strings.TrimSpace(userID)
	threadID = strings.TrimSpace(threadID)
	if userID == "" || threadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(threadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberLocked(threadID, "user", userID) == nil {
		return nil, ErrNotFound
	}
	return s.chatThreadMemberProfilesLocked(threadID, userID), nil
}

func (s *Store) ChatThreadMembership(userID string, threadID string) (*ChatThreadMembership, error) {
	userID = strings.TrimSpace(userID)
	threadID = strings.TrimSpace(threadID)
	if userID == "" || threadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(threadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.chatThreadMembershipLocked(userID, threadID)
}

func (s *Store) JoinChatThread(req ChatThreadJoinRequest) (*ChatThreadMembership, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	if req.UserID == "" || req.ThreadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	policy := chatThreadSelfServePolicy(req.ThreadID)
	if !policy.Joinable {
		return nil, fmt.Errorf("%w: thread does not support self join", ErrInvalidArgument)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if membership, err := s.chatThreadMembershipLocked(req.UserID, req.ThreadID); err != nil {
		return nil, err
	} else if membership.Joined {
		return membership, nil
	}
	if s.chatThreadMembers == nil {
		s.chatThreadMembers = seedChatThreadMembers()
	}
	member := ChatThreadMember{
		ThreadID:    req.ThreadID,
		SubjectType: "user",
		SubjectID:   req.UserID,
		Muted:       policy.DefaultMuted,
		JoinedAt:    time.Now().UTC(),
	}
	s.chatThreadMembers[chatThreadMemberKey(req.ThreadID, member.SubjectType, member.SubjectID)] = cloneChatThreadMember(&member)
	return s.chatThreadMembershipLocked(req.UserID, req.ThreadID)
}

func (s *Store) LeaveChatThread(req ChatThreadLeaveRequest) (*ChatThreadMembership, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	if req.UserID == "" || req.ThreadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	policy := chatThreadSelfServePolicy(req.ThreadID)
	if !policy.Leaveable {
		return nil, fmt.Errorf("%w: thread does not support self leave", ErrInvalidArgument)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberExactLocked(req.ThreadID, "user", req.UserID) == nil {
		return nil, ErrNotFound
	}
	delete(s.chatThreadMembers, chatThreadMemberKey(req.ThreadID, "user", req.UserID))
	return s.chatThreadMembershipLocked(req.UserID, req.ThreadID)
}

func (s *Store) ChatMessages(userID string, threadID string) ([]ChatMessage, error) {
	userID = strings.TrimSpace(userID)
	threadID = strings.TrimSpace(threadID)
	if userID == "" || threadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(threadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberLocked(threadID, "user", userID) == nil {
		return nil, ErrNotFound
	}
	messages := make([]ChatMessage, 0)
	for _, message := range s.chatMessages {
		if message == nil || message.ThreadID != threadID {
			continue
		}
		messages = append(messages, *cloneChatMessage(message))
	}
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
	return messages, nil
}

func (s *Store) ChatMessageSync(req ChatMessageSyncRequest) (*ChatMessageSyncResult, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	req.SinceID = strings.TrimSpace(req.SinceID)
	if req.UserID == "" || req.ThreadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberLocked(req.ThreadID, "user", req.UserID) == nil {
		return nil, ErrNotFound
	}
	allMessages := s.sortedChatMessagesLocked(req.ThreadID)
	messages := chatMessagesAfterID(allMessages, req.SinceID)
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	lastMessageID := ""
	if len(allMessages) > 0 {
		lastMessageID = allMessages[len(allMessages)-1].ID
	}
	nextCursor := req.SinceID
	if len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID
	}
	var read *ChatReadState
	if req.MarkRead {
		read = s.markChatThreadReadLocked(MarkChatThreadReadRequest{
			UserID:        req.UserID,
			ThreadID:      req.ThreadID,
			LastMessageID: lastMessageID,
		})
	} else {
		read = s.chatReadStateLocked(req.UserID, req.ThreadID)
	}
	thread := chatThreadByID(req.ThreadID)
	unread := 0
	if thread != nil {
		unread = s.unreadChatMessageCountLocked(req.UserID, thread)
	}
	result := &ChatMessageSyncResult{
		ThreadID:      req.ThreadID,
		Messages:      messages,
		LastMessageID: lastMessageID,
		NextCursor:    nextCursor,
		UnreadCount:   unread,
		ReadState:     cloneChatReadState(read),
	}
	return result, nil
}

func (s *Store) MarkChatThreadRead(req MarkChatThreadReadRequest) (*ChatReadState, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	req.LastMessageID = strings.TrimSpace(req.LastMessageID)
	if req.UserID == "" || req.ThreadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberLocked(req.ThreadID, "user", req.UserID) == nil {
		return nil, ErrNotFound
	}
	read := s.markChatThreadReadLocked(req)
	return cloneChatReadState(read), nil
}

func (s *Store) ChatThreadPreference(userID string, threadID string) (*ChatThreadPreference, error) {
	userID = strings.TrimSpace(userID)
	threadID = strings.TrimSpace(threadID)
	if userID == "" || threadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(threadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	member := s.chatThreadMemberLocked(threadID, "user", userID)
	if member == nil {
		return nil, ErrNotFound
	}
	return &ChatThreadPreference{
		ThreadID: threadID,
		Muted:    member.Muted,
	}, nil
}

func (s *Store) UpdateChatThreadPreference(req UpdateChatThreadPreferenceRequest) (*ChatThreadPreference, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	if req.UserID == "" || req.ThreadID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	member := s.chatThreadMemberLocked(req.ThreadID, "user", req.UserID)
	if member == nil {
		return nil, ErrNotFound
	}
	member.SubjectID = req.UserID
	member.SubjectType = "user"
	member.Muted = req.Muted
	if s.chatThreadMembers == nil {
		s.chatThreadMembers = seedChatThreadMembers()
	}
	s.chatThreadMembers[chatThreadMemberKey(req.ThreadID, member.SubjectType, member.SubjectID)] = cloneChatThreadMember(member)
	return &ChatThreadPreference{
		ThreadID: req.ThreadID,
		Muted:    member.Muted,
	}, nil
}

func (s *Store) SendChatMessage(message ChatMessage) (*ChatMessage, error) {
	message.ThreadID = strings.TrimSpace(message.ThreadID)
	message.SenderID = strings.TrimSpace(message.SenderID)
	message.Sender = strings.TrimSpace(message.Sender)
	message.Content = strings.TrimSpace(message.Content)
	message.MessageType = strings.TrimSpace(message.MessageType)
	if message.MessageType == "" {
		message.MessageType = "text"
	}
	if message.Sender == "" {
		message.Sender = "我"
	}
	if message.ThreadID == "" || message.SenderID == "" || message.Content == "" || !knownChatThread(message.ThreadID) {
		return nil, ErrInvalidArgument
	}
	now := time.Now().UTC()
	risk := messageRiskCheck(message.Content, now)
	if risk.State == MessageRiskBlocked {
		return nil, fmt.Errorf("%w: %s", ErrRiskControlRejected, risk.Reason)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThreadMemberLocked(message.ThreadID, chatSubjectTypeForID(message.SenderID), message.SenderID) == nil {
		return nil, ErrNotFound
	}
	s.nextChatMessageID++
	message.ID = fmt.Sprintf("msg_%d", s.nextChatMessageID)
	message.CreatedAt = now
	applyMessageRiskToChatMessage(&message, risk)
	s.chatMessages[message.ID] = cloneChatMessage(&message)
	s.enqueueOutboxEventLocked(
		"message.sent",
		"chat_thread",
		message.ThreadID,
		"message.sent",
		"message.sent:"+message.ID,
		chatMessageOutboxPayload(message),
		message.CreatedAt,
	)
	return cloneChatMessage(&message), nil
}

func (s *Store) AuthorizeChatThreadAccess(req ChatThreadAccessRequest) (*ChatThreadAccessResult, error) {
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	req.SubjectID = strings.TrimSpace(req.SubjectID)
	req.SubjectType = normalizeChatSubjectType(req.SubjectType, req.Role, req.SubjectID)
	if req.ThreadID == "" || req.SubjectType == "" || req.SubjectID == "" {
		return nil, ErrInvalidArgument
	}
	if !knownChatThread(req.ThreadID) {
		return nil, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	member := s.chatThreadMemberLocked(req.ThreadID, req.SubjectType, req.SubjectID)
	if member == nil {
		return nil, ErrNotFound
	}
	return &ChatThreadAccessResult{
		ThreadID:    req.ThreadID,
		SubjectType: member.SubjectType,
		SubjectID:   req.SubjectID,
		Allowed:     true,
		Muted:       member.Muted,
	}, nil
}

func (s *Store) WalletOverview(userID string) (*WalletOverview, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	transactions := s.walletTransactionsForUserLocked(userID)
	if len(transactions) == 0 {
		transactions = defaultWalletTransactions(userID)
	}
	withdrawals := make([]WalletWithdrawRequest, 0)
	for _, request := range s.withdrawRequests {
		if request != nil && request.UserID == userID {
			withdrawals = append(withdrawals, *cloneWalletWithdrawRequest(request))
		}
	}
	sort.SliceStable(withdrawals, func(i, j int) bool {
		return withdrawals[i].CreatedAt.After(withdrawals[j].CreatedAt)
	})

	account := s.wallets[userID]
	accountCopy := WalletAccount{UserID: userID, Balance: s.walletBalanceForOverviewLocked(userID), RiskState: "normal"}
	if account != nil {
		accountCopy = *cloneWalletAccount(account)
	}
	coupons := s.userCouponsForUserLocked(userID)
	return &WalletOverview{
		Account:               accountCopy,
		BalanceFen:            accountCopy.Balance,
		PendingReceivableFen:  s.pendingReceivableFenForUserLocked(userID),
		CouponCount:           len(coupons),
		RedPacketCount:        s.redPacketCountForUserLocked(userID),
		Points:                s.pointsBalanceForUserLocked(userID),
		PaymentPasswordStatus: s.paymentPasswordStatusForUserLocked(userID),
		Transactions:          transactions,
		Withdrawals:           withdrawals,
	}, nil
}

func (s *Store) RequestWalletWithdraw(req WalletWithdrawRequest) (*WalletWithdrawRequest, *WalletAccount, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.AccountNo = strings.TrimSpace(req.AccountNo)
	if req.Channel == "" {
		req.Channel = "wechat_change"
	}
	if req.UserID == "" || req.AmountFen <= 0 {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	account := s.getOrCreateWalletLocked(req.UserID)
	if account.Balance == 0 && req.UserID == "user_1" {
		account.Balance = 12850
		account.Version++
	}
	if account.Balance < req.AmountFen {
		return nil, nil, ErrInsufficientBalance
	}
	account.Balance -= req.AmountFen
	account.Frozen += req.AmountFen
	account.Version++
	now := time.Now().UTC()
	s.nextWithdrawID++
	req.ID = fmt.Sprintf("wd_%d", s.nextWithdrawID)
	req.Status = "processing"
	req.CreatedAt = now
	s.withdrawRequests[req.ID] = cloneWalletWithdrawRequest(&req)
	transaction := s.createWalletTransactionLocked(req.UserID, "", "withdraw", -req.AmountFen, req.Channel, "withdraw:"+req.ID)
	transaction.Status = "processing"
	s.walletIdempotency[transaction.IdempotencyKey] = transaction
	return cloneWalletWithdrawRequest(&req), cloneWalletAccount(account), nil
}

func (s *Store) UserCoupons(userID string) (*UserCouponsSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	coupons := s.userCouponsForUserLocked(userID)
	now := time.Now().UTC()
	expiring := 0
	redPackets := s.redPacketCountForUserLocked(userID)
	for _, coupon := range coupons {
		if coupon.Status == "available" && !coupon.ExpiresAt.IsZero() && coupon.ExpiresAt.Before(now.Add(48*time.Hour)) {
			expiring++
		}
	}
	return &UserCouponsSummary{
		AvailableCount: len(coupons),
		RedPacketCount: redPackets,
		ExpiringCount:  expiring,
		Coupons:        coupons,
	}, nil
}

func (s *Store) ClaimUserCoupon(userID string, code string) (*UserCoupon, error) {
	userID = strings.TrimSpace(userID)
	code = strings.ToUpper(strings.TrimSpace(code))
	if userID == "" || code == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if coupon, handled, err := s.claimMerchantGroupCouponLocked(userID, code); handled {
		return coupon, err
	}
	s.nextCouponID++
	now := time.Now().UTC()
	coupon := &UserCoupon{
		ID:           fmt.Sprintf("uc_%d", s.nextCouponID),
		UserID:       userID,
		Kind:         "platform",
		Title:        "兑换码优惠券",
		Subtitle:     "兑换码 " + strings.ToUpper(code),
		Scope:        "外卖",
		Source:       "平台券",
		Status:       "available",
		ButtonText:   "去使用",
		AccentColor:  "#007aff",
		AmountFen:    500,
		ThresholdFen: 3000,
		ExpiresAt:    now.Add(7 * 24 * time.Hour),
		CreatedAt:    now,
	}
	s.userCoupons[coupon.ID] = cloneUserCoupon(coupon)
	return cloneUserCoupon(coupon), nil
}

func (s *Store) UserPointsSummary(userID string) (*PointsSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pointsSummaryLocked(userID), nil
}

func (s *Store) CheckInPoints(userID string) (*PointsSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().UTC().Format("2006-01-02")
	for _, transaction := range s.pointsTransactions[userID] {
		if transaction != nil && transaction.SourceID == "checkin:"+today {
			return s.pointsSummaryLocked(userID), nil
		}
	}
	now := time.Now().UTC()
	s.pointsTransactions[userID] = append(s.pointsTransactions[userID], &PointsTransaction{
		ID:        "pt_checkin_" + today,
		UserID:    userID,
		Type:      PointsTransactionEarn,
		Title:     "每日签到奖励",
		Points:    5,
		SourceID:  "checkin:" + today,
		CreatedAt: now,
	})
	return s.pointsSummaryLocked(userID), nil
}

func (s *Store) InviteSummary(userID string) (*InviteSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	code := inviteCodeForUser(userID)
	return &InviteSummary{
		UserID:       userID,
		InviteCode:   code,
		RewardText:   "好友首单完成后，双方各得一张优惠券",
		ShareTitle:   "来悦享e食一起点餐",
		SharePath:    "/pages/auth/register/index?invite=" + code,
		AbuseRiskTip: "同一手机号仅计一次，异常订单不参与奖励",
		Records: []InviteRecord{
			{ID: "inv_1", FriendName: "李四", Status: "首单待完成", RewardText: "待发放", CreatedAt: time.Now().UTC().Add(-48 * time.Hour)},
			{ID: "inv_2", FriendName: "王五", Status: "奖励已到账", RewardText: "+1 张优惠券", CreatedAt: time.Now().UTC().Add(-96 * time.Hour)},
		},
	}, nil
}

func (s *Store) SearchCatalog(userID string, keyword string, category string) (*SearchCatalog, error) {
	userID = strings.TrimSpace(userID)
	keyword = strings.TrimSpace(keyword)
	category = strings.TrimSpace(category)
	if category == "" {
		category = "all"
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	results := []SearchResult{}
	for _, shop := range s.shops {
		if shop == nil {
			continue
		}
		resultType := "shop"
		if containsString(shop.Capabilities, ShopCapabilityMedicine) {
			resultType = "medicine"
		}
		results = append(results, SearchResult{
			ID:           shop.ID,
			Type:         resultType,
			Title:        shop.Name,
			Subtitle:     shop.Category + " · 月售 1.2万+",
			DistanceText: "800m",
			ImageURL:     defaultString(shop.CoverURL, shop.LogoURL),
			Badge:        "品牌",
			ButtonText:   buttonTextForSearchType(resultType),
			Route:        "/pages/shop/detail/index?id=" + shop.ID,
		})
	}
	for _, product := range s.products {
		if product == nil || product.Status != ProductStatusActive {
			continue
		}
		results = append(results, SearchResult{
			ID:         product.ID,
			Type:       "product",
			Title:      product.Name,
			Subtitle:   "蓝海餐厅 · 好评率 98%",
			PriceFen:   product.PriceFen,
			ImageURL:   product.ImageURL,
			ButtonText: "加入购物车",
			Route:      "/pages/shop/detail/index?id=" + product.ShopID,
		})
	}
	for _, deal := range s.groupbuyDeals {
		if deal == nil || deal.Status != ProductStatusActive {
			continue
		}
		results = append(results, SearchResult{
			ID:         deal.ID,
			Type:       "groupbuy",
			Title:      deal.Name,
			Subtitle:   "周末通用 · 免预约",
			PriceFen:   deal.PriceFen,
			ImageURL:   deal.ImageURL,
			Badge:      "团购券",
			ButtonText: "去购买",
			Route:      "/pages/shop/detail/index?id=" + deal.ShopID,
		})
	}
	for _, product := range defaultMedicineProducts() {
		results = append(results, SearchResult{
			ID:         product.ID,
			Type:       "medicine",
			Title:      product.Name,
			Subtitle:   product.Category + " · 校医务室",
			PriceFen:   product.PriceFen,
			ImageURL:   product.ImageURL,
			Badge:      "买药",
			ButtonText: "去买药",
			Route:      "/pages/medicine/home/index",
		})
	}
	results = append(results, SearchResult{
		ID:           "errand_nearby",
		Type:         "errand",
		Title:        "附近跑腿",
		Subtitle:     "帮买、帮送、帮取、帮办",
		DistanceText: "约 35 分钟",
		ImageURL:     errandImageURL(),
		Badge:        "跑腿",
		ButtonText:   "去下单",
		Route:        "/pages/errand/home/index",
	})
	results = filterSearchResults(results, keyword, category)
	suggestions := searchSuggestionsFromResults(results)
	return &SearchCatalog{
		Keyword:     keyword,
		Category:    category,
		Total:       len(results),
		Suggestions: suggestions,
		Results:     results,
	}, nil
}

func (s *Store) MedicineHome(userID string) (*MedicineHome, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureMedicineStockLocked()
	products := defaultMedicineProducts()
	cartCount := 0
	var cartAmountFen int64
	for index := range products {
		stock := s.medicineStock[products[index].ID]
		if stock < 0 {
			stock = 0
		}
		products[index].StockCount = stock
		if products[index].SelectedQuantity > stock {
			products[index].SelectedQuantity = stock
		}
		if products[index].SelectedQuantity > 0 {
			cartCount += products[index].SelectedQuantity
			cartAmountFen += products[index].PriceFen * int64(products[index].SelectedQuantity)
		}
	}
	return &MedicineHome{
		Clinic: MedicineClinic{
			Name:         "校医务室",
			Location:     "综合楼一层",
			CoverURL:     "/assets/generated/medicine-clinic-cover.jpg",
			BusinessTime: "今日 08:30-20:30",
			Tags:         []string{"校内服务", "骑手配送"},
			DeliveryText: "校内配送约 20-30 分钟",
		},
		Categories:    []string{"全部", "感冒发热", "处方药", "外伤消毒", "医用耗材"},
		Products:      products,
		CartCount:     cartCount,
		CartAmountFen: cartAmountFen,
	}, nil
}

func (s *Store) CreatePrescriptionImageUpload(req CreatePrescriptionImageUploadRequest) (*ObjectUploadTicket, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	if req.ProductID == "" {
		req.ProductID = "med_amoxicillin"
	}
	if req.UserID == "" || req.FileName == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	storage := s.objectStorageConfigLocked()
	scanStatus := AfterSalesUploadScanNotRequired
	if storage.RequireScanApprovalForConfirm {
		scanStatus = AfterSalesUploadScanPending
	}
	expiresAt := now.Add(storage.TicketTTL)
	objectKey := fmt.Sprintf("prescriptions/%s/%s/%s", shortHash(req.UserID), shortHash(fmt.Sprintf("%s:%s:%d", req.UserID, req.FileName, now.UnixNano())), req.FileName)
	ticket, err := storage.createObjectUploadTicket(objectUploadTicketInput{
		ObjectKey:    objectKey,
		ContentType:  req.ContentType,
		SizeBytes:    req.SizeBytes,
		MaxSizeBytes: AfterSalesEvidenceMaxBytes,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}
	ticketID := "rxu_" + shortHash(ticket.ObjectKey)
	ticket.TicketID = ticketID
	if s.prescriptionImageTickets == nil {
		s.prescriptionImageTickets = map[string]*PrescriptionImageUploadTicket{}
	}
	s.prescriptionImageTickets[ticketID] = &PrescriptionImageUploadTicket{
		ID:           ticketID,
		UserID:       req.UserID,
		ProductID:    req.ProductID,
		Provider:     ticket.Provider,
		Bucket:       ticket.Bucket,
		ObjectKey:    ticket.ObjectKey,
		PublicURL:    ticket.PublicURL,
		FileName:     req.FileName,
		ContentType:  req.ContentType,
		SizeBytes:    req.SizeBytes,
		MaxSizeBytes: ticket.MaxSizeBytes,
		Status:       AfterSalesUploadTicketIssued,
		ScanStatus:   scanStatus,
		CreatedAt:    now,
		ExpiresAt:    ticket.ExpiresAt,
	}
	return ticket, nil
}

func (s *Store) ConfirmPrescriptionImageUpload(req ConfirmPrescriptionImageUploadRequest) (*PrescriptionImageUploadTicket, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	req.ContentSHA = strings.TrimSpace(req.ContentSHA)
	if req.UserID == "" || req.TicketID == "" || req.ObjectKey == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}
	if !validPrescriptionImageObjectKey(req.ObjectKey) {
		return nil, ErrInvalidArgument
	}
	if req.FileName == "" {
		req.FileName = sanitizeObjectFileName(objectKeyFileName(req.ObjectKey))
	}
	if req.FileName == "" {
		return nil, ErrInvalidArgument
	}

	ticket, storage, err := s.preparePrescriptionImageConfirmation(req)
	if err != nil {
		return nil, err
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return ticket, nil
	}
	if err := storage.verifyUploadedObject(objectHeadCheckInput{
		ObjectKey:   ticket.ObjectKey,
		ContentType: ticket.ContentType,
		SizeBytes:   ticket.SizeBytes,
	}); err != nil {
		return nil, err
	}
	return s.commitPrescriptionImageConfirmation(req)
}

func (s *Store) CreatePrescriptionReview(req CreatePrescriptionReviewRequest) (*PrescriptionReview, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.PatientName = strings.TrimSpace(req.PatientName)
	req.PatientPhone = strings.TrimSpace(req.PatientPhone)
	req.Address = strings.TrimSpace(req.Address)
	req.Hospital = strings.TrimSpace(req.Hospital)
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.ProductName = strings.TrimSpace(req.ProductName)
	req.ProductImage = strings.TrimSpace(req.ProductImage)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	req.PrescriptionImageTicketID = strings.TrimSpace(req.PrescriptionImageTicketID)
	req.PrescriptionObjectKey = strings.TrimSpace(req.PrescriptionObjectKey)
	req.PrescriptionContentSHA = strings.TrimSpace(req.PrescriptionContentSHA)
	req.Note = strings.TrimSpace(req.Note)
	if req.PatientName == "" {
		req.PatientName = "张三"
	}
	if req.PatientPhone == "" {
		req.PatientPhone = "13800000000"
	}
	if req.Address == "" {
		req.Address = "望京校区 3 号宿舍楼"
	}
	if req.Hospital == "" {
		req.Hospital = "校医务室"
	}
	if req.ProductID == "" {
		req.ProductID = "med_amoxicillin"
	}
	if req.ProductName == "" {
		req.ProductName = "阿莫西林胶囊"
	}
	if req.ProductImage == "" {
		req.ProductImage = medicineProductImageURL(req.ProductID)
	}
	if req.PriceFen <= 0 {
		req.PriceFen = 1880
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	if req.UserID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	var imageTicket *PrescriptionImageUploadTicket
	if req.PrescriptionImageTicketID != "" || req.PrescriptionObjectKey != "" {
		var err error
		imageTicket, err = s.prescriptionImageUploadTicketForReviewLocked(req)
		if err != nil {
			return nil, err
		}
		if req.ImageURL == "" {
			req.ImageURL = imageTicket.PublicURL
		}
		if req.PrescriptionObjectKey == "" {
			req.PrescriptionObjectKey = imageTicket.ObjectKey
		}
		if req.PrescriptionContentSHA == "" {
			req.PrescriptionContentSHA = imageTicket.ContentSHA
		}
	}
	now := time.Now().UTC()
	reviewedAt := now.Add(3 * time.Minute)
	s.nextPrescriptionID++
	review := &PrescriptionReview{
		ID:              fmt.Sprintf("rx_%d", s.nextPrescriptionID),
		UserID:          req.UserID,
		PatientName:     req.PatientName,
		PatientPhone:    req.PatientPhone,
		Address:         req.Address,
		Hospital:        req.Hospital,
		ProductID:       req.ProductID,
		ProductName:     req.ProductName,
		ProductImage:    req.ProductImage,
		PriceFen:        req.PriceFen,
		Quantity:        req.Quantity,
		ImageURL:        req.ImageURL,
		ImageObjectKey:  req.PrescriptionObjectKey,
		ImageContentSHA: req.PrescriptionContentSHA,
		Note:            req.Note,
		Status:          PrescriptionReviewApproved,
		DoctorName:      "王医生",
		ReviewText:      "处方信息已确认，可加入购物车购买。",
		OCRResult:       prescriptionOCRResult(req, imageTicket),
		Archive:         prescriptionArchiveRecord(fmt.Sprintf("rx_%d", s.nextPrescriptionID), req.PrescriptionObjectKey, req.PrescriptionContentSHA, now),
		CreatedAt:       now,
		ReviewedAt:      reviewedAt,
		UpdatedAt:       reviewedAt,
	}
	if imageTicket != nil {
		review.ImageUploadTicketID = imageTicket.ID
	}
	review.Steps = prescriptionSteps(review)
	s.prescriptionReviews[review.ID] = clonePrescriptionReview(review)
	return clonePrescriptionReview(review), nil
}

func (s *Store) PrescriptionReview(userID string, reviewID string) (*PrescriptionReview, error) {
	userID = strings.TrimSpace(userID)
	reviewID = strings.TrimSpace(reviewID)
	if userID == "" || reviewID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	review := s.prescriptionReviews[reviewID]
	if review == nil {
		if reviewID == "rx_preview" && userID == "user_1" {
			preview := defaultPrescriptionReview(userID)
			return clonePrescriptionReview(&preview), nil
		}
		return nil, ErrNotFound
	}
	if review.UserID != userID {
		return nil, ErrNotFound
	}
	return clonePrescriptionReview(review), nil
}

func (s *Store) AdminPrescriptionReviews(req PrescriptionReviewListRequest) ([]PrescriptionReview, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.ProductID = strings.TrimSpace(req.ProductID)
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	reviews := make([]PrescriptionReview, 0, len(s.prescriptionReviews))
	for _, review := range s.prescriptionReviews {
		if review == nil {
			continue
		}
		if req.Status != "" && review.Status != req.Status {
			continue
		}
		if req.ProductID != "" && review.ProductID != req.ProductID {
			continue
		}
		reviews = append(reviews, *clonePrescriptionReview(review))
	}
	sort.Slice(reviews, func(i, j int) bool {
		return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
	})
	if len(reviews) > req.Limit {
		reviews = reviews[:req.Limit]
	}
	return reviews, nil
}

func (s *Store) ReviewPrescription(req ReviewPrescriptionRequest) (*PrescriptionReview, error) {
	req.ReviewID = strings.TrimSpace(req.ReviewID)
	req.ReviewerID = strings.TrimSpace(req.ReviewerID)
	req.ReviewerName = strings.TrimSpace(req.ReviewerName)
	req.Decision = strings.TrimSpace(req.Decision)
	req.ReviewText = strings.TrimSpace(req.ReviewText)
	if req.ReviewID == "" || req.ReviewerID == "" {
		return nil, ErrInvalidArgument
	}
	if req.Decision != PrescriptionReviewApproved && req.Decision != PrescriptionReviewRejected {
		return nil, ErrInvalidArgument
	}
	if req.ReviewerName == "" {
		req.ReviewerName = "药师"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	review := s.prescriptionReviews[req.ReviewID]
	if review == nil {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	review.Status = req.Decision
	review.DoctorName = req.ReviewerName
	if req.ReviewText != "" {
		review.ReviewText = req.ReviewText
	} else if req.Decision == PrescriptionReviewApproved {
		review.ReviewText = "药师复核通过，可加入购物车购买。"
	} else {
		review.ReviewText = "处方信息未通过复核，请重新上传清晰处方或联系校医。"
	}
	review.ReviewedAt = now
	review.UpdatedAt = now
	if review.OCRResult == nil {
		review.OCRResult = prescriptionOCRResult(CreatePrescriptionReviewRequest{
			ProductID:              review.ProductID,
			ProductName:            review.ProductName,
			PatientName:            review.PatientName,
			ImageURL:               review.ImageURL,
			PriceFen:               review.PriceFen,
			Quantity:               review.Quantity,
			PrescriptionObjectKey:  review.ImageObjectKey,
			PrescriptionContentSHA: review.ImageContentSHA,
		}, nil)
	}
	if review.Archive == nil {
		review.Archive = prescriptionArchiveRecord(review.ID, review.ImageObjectKey, review.ImageContentSHA, review.CreatedAt)
	}
	review.Steps = prescriptionSteps(review)
	return clonePrescriptionReview(review), nil
}

func (s *Store) CreateMedicineOrder(req MedicineOrderRequest) (*MedicineOrderDetail, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Address = strings.TrimSpace(req.Address)
	req.ContactName = strings.TrimSpace(req.ContactName)
	req.ContactPhone = strings.TrimSpace(req.ContactPhone)
	req.ClinicName = strings.TrimSpace(req.ClinicName)
	req.PrescriptionID = strings.TrimSpace(req.PrescriptionID)
	req.PaymentMethod = strings.TrimSpace(req.PaymentMethod)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	if req.Address == "" {
		req.Address = "望京校区 3 号宿舍楼 508"
	}
	if req.ContactName == "" {
		req.ContactName = "张三"
	}
	if req.ContactPhone == "" {
		req.ContactPhone = "13800000000"
	}
	if req.ClinicName == "" {
		req.ClinicName = "校医务室"
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = PaymentBalance
	}
	if req.DeliveryFeeFen <= 0 {
		req.DeliveryFeeFen = 200
	}
	items := normalizedMedicineOrderItems(req.Items)
	if len(items) == 0 {
		items = defaultMedicineOrderItems()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	prescriptionStatus := ""
	doctorName := ""
	if req.PrescriptionID != "" {
		review := s.prescriptionReviews[req.PrescriptionID]
		if review == nil || review.UserID != req.UserID {
			return nil, ErrNotFound
		}
		if review.Status != PrescriptionReviewApproved {
			return nil, ErrInvalidOrderState
		}
		prescriptionStatus = review.Status
		doctorName = review.DoctorName
	}
	if err := s.lockMedicineStockLocked(items); err != nil {
		return nil, err
	}
	for index := range items {
		if items[index].RequiresPrescription && prescriptionStatus == PrescriptionReviewApproved {
			items[index].PrescriptionApproved = true
		}
	}
	now := time.Now().UTC()
	subtotal := medicineItemsTotalFen(items)
	payable := subtotal + req.DeliveryFeeFen - req.CouponAmountFen
	if payable <= 0 {
		payable = subtotal + req.DeliveryFeeFen
	}
	s.nextOrderID++
	orderItems := make([]OrderItem, 0, len(items))
	for _, item := range items {
		orderItems = append(orderItems, OrderItem{
			ProductID:    item.ProductID,
			ProductName:  item.Name,
			ImageURL:     item.ImageURL,
			UnitPriceFen: item.PriceFen,
			Quantity:     item.Quantity,
		})
	}
	order := &Order{
		ID:             fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:         req.UserID,
		Type:           OrderTypeMedicine,
		Status:         StatusPickedUp,
		AmountFen:      payable,
		ItemsTotalFen:  subtotal,
		DeliveryFeeFen: req.DeliveryFeeFen,
		DiscountFen:    req.CouponAmountFen,
		PaymentMethod:  req.PaymentMethod,
		RiderID:        "rider_zhang",
		Items:          orderItems,
		CreatedAt:      now,
		UpdatedAt:      now.Add(12 * time.Minute),
		Events: []OrderEvent{
			{Type: "medicine.created", ActorID: req.UserID, Message: "订单已提交", CreatedAt: now},
			{Type: "medicine.dispensed", ActorID: "clinic", Message: "校医出药完成", CreatedAt: now.Add(4 * time.Minute)},
			{Type: "medicine.picked_up", ActorID: "rider_zhang", Message: "骑手已取药", CreatedAt: now.Add(12 * time.Minute)},
		},
	}
	s.orders[order.ID] = order
	detail := s.medicineDetailForOrderLocked(order, req, items, prescriptionStatus, doctorName)
	s.medicineDetails[order.ID] = cloneMedicineOrderDetail(detail)
	return cloneMedicineOrderDetail(detail), nil
}

func (s *Store) MedicineOrderDetail(userID string, orderID string) (*MedicineOrderDetail, error) {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if detail := s.medicineDetails[orderID]; detail != nil {
		if detail.Order.UserID != userID {
			return nil, ErrNotFound
		}
		return cloneMedicineOrderDetail(detail), nil
	}
	order := s.orders[orderID]
	if order == nil || order.UserID != userID || order.Type != OrderTypeMedicine {
		return nil, ErrNotFound
	}
	items := defaultMedicineOrderItems()
	if len(order.Items) > 0 {
		items = make([]MedicineOrderItem, 0, len(order.Items))
		for _, item := range order.Items {
			items = append(items, MedicineOrderItem{
				ProductID: item.ProductID,
				Name:      item.ProductName,
				ImageURL:  item.ImageURL,
				PriceFen:  item.UnitPriceFen,
				Quantity:  item.Quantity,
			})
		}
	}
	detail := s.medicineDetailForOrderLocked(order, MedicineOrderRequest{UserID: userID}, items, PrescriptionReviewApproved, "王医生")
	s.medicineDetails[order.ID] = cloneMedicineOrderDetail(detail)
	return cloneMedicineOrderDetail(detail), nil
}

func (s *Store) CreateErrandOrder(req ErrandOrderRequest) (*ErrandOrderDetail, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Type = strings.TrimSpace(req.Type)
	req.PickupAddress = strings.TrimSpace(req.PickupAddress)
	req.DeliveryAddress = strings.TrimSpace(req.DeliveryAddress)
	req.ContactName = strings.TrimSpace(req.ContactName)
	req.ContactPhone = strings.TrimSpace(req.ContactPhone)
	req.ItemType = strings.TrimSpace(req.ItemType)
	req.Description = strings.TrimSpace(req.Description)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	req.WeightText = strings.TrimSpace(req.WeightText)
	req.PickupTime = strings.TrimSpace(req.PickupTime)
	if req.Type == "" {
		req.Type = OrderTypeErrandPickup
	}
	if req.AmountFen <= 0 {
		req.AmountFen = 1600
	}
	if req.CouponAmountFen < 0 {
		req.CouponAmountFen = 0
	}
	if req.UserID == "" || !isErrandOrderType(req.Type) {
		return nil, ErrInvalidArgument
	}
	if req.PickupAddress == "" {
		req.PickupAddress = "望京SOHO B座 快递柜"
	}
	if req.DeliveryAddress == "" {
		req.DeliveryAddress = "望京SOHO A座 1208"
	}
	if req.ContactName == "" {
		req.ContactName = "张三"
	}
	if req.ContactPhone == "" {
		req.ContactPhone = "13800000000"
	}
	if req.ItemType == "" {
		req.ItemType = "小件包裹"
	}
	if req.Description == "" {
		req.Description = "取 3 号柜快递，验证码已备注"
	}
	if req.ImageURL == "" {
		req.ImageURL = errandImageURL()
	}
	if req.WeightText == "" {
		req.WeightText = "2kg 内"
	}
	if req.PickupTime == "" {
		req.PickupTime = "立即取送"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextOrderID++
	now := time.Now().UTC()
	payable := req.AmountFen - req.CouponAmountFen
	if payable <= 0 {
		payable = req.AmountFen
	}
	order := &Order{
		ID:            fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:        req.UserID,
		Type:          req.Type,
		Status:        StatusRiderAssigned,
		AmountFen:     payable,
		RiderID:       "rider_zhang",
		PaymentMethod: PaymentBalance,
		CreatedAt:     now,
		UpdatedAt:     now,
		Events: []OrderEvent{
			{Type: "errand.created", ActorID: req.UserID, Message: "订单已创建", CreatedAt: now},
			{Type: "errand.rider_assigned", ActorID: "rider_zhang", Message: "骑手已接单", CreatedAt: now.Add(2 * time.Minute)},
			{Type: "errand.to_pickup", ActorID: "rider_zhang", Message: "骑手正在前往取件地址", CreatedAt: now.Add(4 * time.Minute)},
		},
	}
	s.orders[order.ID] = order
	detail := s.errandDetailForOrderLocked(order, req)
	s.errandDetails[order.ID] = cloneErrandOrderDetail(detail)
	return cloneErrandOrderDetail(detail), nil
}

func (s *Store) ErrandOrderDetail(userID string, orderID string) (*ErrandOrderDetail, error) {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if detail := s.errandDetails[orderID]; detail != nil {
		if detail.Order.UserID != userID {
			return nil, ErrNotFound
		}
		return cloneErrandOrderDetail(detail), nil
	}
	order := s.orders[orderID]
	if order == nil || order.UserID != userID || !isErrandOrderType(order.Type) {
		return nil, ErrNotFound
	}
	detail := s.errandDetailForOrderLocked(order, ErrandOrderRequest{UserID: userID, Type: order.Type, AmountFen: order.AmountFen})
	s.errandDetails[order.ID] = cloneErrandOrderDetail(detail)
	return cloneErrandOrderDetail(detail), nil
}

func (s *Store) SetWalletPaymentPassword(req SetWalletPaymentPasswordRequest) (*WalletPaymentPasswordState, error) {
	userID := strings.TrimSpace(req.UserID)
	password := strings.TrimSpace(req.Password)
	if userID == "" || !isSixDigitPaymentPassword(password) {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.paymentPasswordHash[userID] = hashPaymentPassword(password)
	return &WalletPaymentPasswordState{UserID: userID, Status: WalletPaymentPasswordSet}, nil
}

func (s *Store) CreateWechatPrepay(req WechatPrepayRequest) (*WechatPrepayResponse, *PaymentTransaction, error) {
	userID := strings.TrimSpace(req.UserID)
	orderID := strings.TrimSpace(req.OrderID)
	if userID == "" || orderID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	if order.UserID != userID || order.Status != StatusPendingPayment {
		return nil, nil, ErrInvalidOrderState
	}
	idempotencyKey := "wechat_prepay:" + orderID
	for _, existing := range s.paymentTransactions {
		if existing.IdempotencyKey == idempotencyKey {
			return wechatPrepayResponseForTransaction(existing), clonePaymentTransaction(existing), nil
		}
	}
	s.nextTransactionID++
	now := time.Now().UTC()
	outTradeNo := fmt.Sprintf("wx_%s_%d", orderID, s.nextTransactionID)
	transaction := &PaymentTransaction{
		ID:             fmt.Sprintf("ptx_%d", s.nextTransactionID),
		OrderID:        orderID,
		UserID:         userID,
		Method:         PaymentWechat,
		AmountFen:      order.AmountFen,
		Status:         "prepay_created",
		OutTradeNo:     outTradeNo,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.paymentTransactions[transaction.ID] = transaction
	s.paymentByTradeNo[outTradeNo] = transaction
	return wechatPrepayResponseForTransaction(transaction), clonePaymentTransaction(transaction), nil
}

func (s *Store) ConfirmWechatPayment(req WechatPaymentCallbackRequest) (*PaymentTransaction, *Order, error) {
	outTradeNo := strings.TrimSpace(req.OutTradeNo)
	transactionID := strings.TrimSpace(req.TransactionID)
	if outTradeNo == "" || transactionID == "" || req.AmountFen <= 0 {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.paymentByProviderID[transactionID]; existing != nil {
		return clonePaymentTransaction(existing), cloneOrder(s.orders[existing.OrderID]), nil
	}
	transaction := s.paymentByTradeNo[outTradeNo]
	if transaction == nil {
		return nil, nil, ErrNotFound
	}
	if transaction.AmountFen != req.AmountFen {
		return nil, nil, ErrInvalidArgument
	}
	order := s.orders[transaction.OrderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	now := time.Now().UTC()
	transaction.Status = "success"
	transaction.TransactionID = transactionID
	transaction.UpdatedAt = now
	s.paymentByProviderID[transactionID] = transaction
	if order.Status == StatusPendingPayment {
		order.Status = statusAfterPayment(order)
		order.PaymentMethod = PaymentWechat
		order.UpdatedAt = now
		s.issueGroupbuyVouchersLocked(order, now)
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.payment.success",
			ActorID:   "wechat_pay",
			Message:   paymentSuccessMessage(order),
			CreatedAt: now,
		})
	}
	return clonePaymentTransaction(transaction), cloneOrder(order), nil
}

func (s *Store) PayOrderWithBalance(req BalancePayRequest) (*WalletTransaction, *Order, *WalletAccount, error) {
	userID := strings.TrimSpace(req.UserID)
	orderID := strings.TrimSpace(req.OrderID)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if userID == "" || orderID == "" || idempotencyKey == "" {
		return nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.walletIdempotency[idempotencyKey]; existing != nil {
		return cloneWalletTransaction(existing), cloneOrder(s.orders[orderID]), cloneWalletAccount(s.wallets[userID]), nil
	}
	if !s.verifyPaymentPasswordLocked(userID, req.PaymentPassword) {
		return nil, nil, nil, ErrPaymentPassword
	}

	order := s.orders[orderID]
	if order == nil {
		return nil, nil, nil, ErrNotFound
	}
	if order.UserID != userID || order.Status != StatusPendingPayment {
		return nil, nil, nil, ErrInvalidOrderState
	}

	account := s.getOrCreateWalletLocked(userID)
	if account.Balance < order.AmountFen {
		return nil, nil, nil, ErrInsufficientBalance
	}

	account.Balance -= order.AmountFen
	account.Version++
	now := time.Now().UTC()
	order.Status = statusAfterPayment(order)
	order.PaymentMethod = PaymentBalance
	order.UpdatedAt = now
	s.issueGroupbuyVouchersLocked(order, now)
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "order.payment.success",
		ActorID:   userID,
		Message:   paymentSuccessMessage(order),
		CreatedAt: now,
	})

	transaction := s.createWalletTransactionLocked(userID, orderID, "payment", -order.AmountFen, PaymentBalance, idempotencyKey)
	s.walletIdempotency[idempotencyKey] = transaction
	return cloneWalletTransaction(transaction), cloneOrder(order), cloneWalletAccount(account), nil
}

func (s *Store) RefundSettings() (*RefundSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings := s.refundSettings
	settings.DefaultStrategy = NormalizeRefundStrategy(settings.DefaultStrategy)
	return &settings, nil
}

func (s *Store) SaveRefundSettings(req SaveRefundSettingsRequest) (*RefundSettings, error) {
	settings := RefundSettings{DefaultStrategy: NormalizeRefundStrategy(req.DefaultStrategy)}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundSettings = settings
	return &settings, nil
}

func (s *Store) SaveRefundSettingsWithAudit(req SaveRefundSettingsRequest, audit RecordAuditLogRequest) (*RefundSettings, *AuditLog, error) {
	settings := RefundSettings{DefaultStrategy: NormalizeRefundStrategy(req.DefaultStrategy)}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, err
	}
	if log.Action != "admin.refund_settings.updated" || log.TargetType != "refund_settings" || log.TargetID != "default" {
		return nil, nil, ErrInvalidArgument
	}
	log.Payload = map[string]any{"default_refund_strategy": settings.DefaultStrategy}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundSettings = settings
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return &settings, cloneAuditLog(log), nil
}

func (s *Store) AdminRefundTransactions(req RefundTransactionListRequest) ([]RefundTransaction, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Destination = strings.TrimSpace(req.Destination)
	req.Status = strings.TrimSpace(req.Status)
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	refunds := make([]RefundTransaction, 0, len(s.refundTransactions))
	for _, refund := range s.refundTransactions {
		if refund == nil {
			continue
		}
		if req.OrderID != "" && refund.OrderID != req.OrderID {
			continue
		}
		if req.UserID != "" && refund.UserID != req.UserID {
			continue
		}
		if req.Destination != "" && refund.Destination != req.Destination {
			continue
		}
		if req.Status != "" && refund.Status != req.Status {
			continue
		}
		refunds = append(refunds, *cloneRefundTransaction(refund))
	}
	sort.SliceStable(refunds, func(i, j int) bool {
		if refunds[i].CreatedAt.Equal(refunds[j].CreatedAt) {
			return refunds[i].ID > refunds[j].ID
		}
		return refunds[i].CreatedAt.After(refunds[j].CreatedAt)
	})
	if len(refunds) > req.Limit {
		refunds = refunds[:req.Limit]
	}
	return refunds, nil
}

func (s *Store) RefundOrder(req RefundOrderRequest) (*RefundTransaction, *Order, *WalletAccount, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Reason = strings.TrimSpace(req.Reason)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.OrderID == "" || req.Reason == "" || req.IdempotencyKey == "" {
		return nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.refundOrderLocked(req)
}

func (s *Store) RefundOrderWithAudit(req RefundOrderRequest, audit RecordAuditLogRequest) (*RefundTransaction, *Order, *WalletAccount, *AuditLog, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Reason = strings.TrimSpace(req.Reason)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if log.Action != "admin.order.refunded" || log.TargetType != "order" || log.TargetID != req.OrderID {
		return nil, nil, nil, nil, ErrInvalidArgument
	}
	if req.OrderID == "" || req.Reason == "" || req.IdempotencyKey == "" {
		return nil, nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	refund, order, account, err := s.refundOrderLocked(req)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	log.Payload = refundOrderAuditPayload(refund)
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return refund, order, account, cloneAuditLog(log), nil
}

func refundOrderAuditPayload(refund *RefundTransaction) map[string]any {
	if refund == nil {
		return map[string]any{}
	}
	return map[string]any{
		"refund_id":       strings.TrimSpace(refund.ID),
		"destination":     strings.TrimSpace(refund.Destination),
		"status":          strings.TrimSpace(refund.Status),
		"amount_fen":      refund.AmountFen,
		"idempotency_key": strings.TrimSpace(refund.IdempotencyKey),
	}
}

func (s *Store) refundOrderLocked(req RefundOrderRequest) (*RefundTransaction, *Order, *WalletAccount, error) {
	orderID := strings.TrimSpace(req.OrderID)
	userID := strings.TrimSpace(req.UserID)
	reason := strings.TrimSpace(req.Reason)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if orderID == "" || reason == "" || idempotencyKey == "" {
		return nil, nil, nil, ErrInvalidArgument
	}
	if existingID := s.refundByIdempotency[idempotencyKey]; existingID != "" {
		if existing := s.refundTransactions[existingID]; existing != nil {
			var account *WalletAccount
			if existing.Destination == RefundDestinationBalance {
				account = s.wallets[existing.UserID]
			}
			return cloneRefundTransaction(existing), cloneOrder(s.orders[existing.OrderID]), cloneWalletAccount(account), nil
		}
		delete(s.refundByIdempotency, idempotencyKey)
	}
	if existingWalletTransaction := s.walletIdempotency[idempotencyKey]; existingWalletTransaction != nil {
		return nil, nil, nil, ErrInvalidOrderState
	}

	order := s.orders[orderID]
	if order == nil {
		return nil, nil, nil, ErrNotFound
	}
	if userID != "" && order.UserID != userID {
		return nil, nil, nil, ErrInvalidOrderState
	}
	switch order.Status {
	case StatusPendingPayment, StatusCancelled, StatusRefundPending, StatusRefunded:
		return nil, nil, nil, ErrInvalidOrderState
	}

	refundedBefore := s.refundedAmountForOrderLocked(order.ID)
	remainingFen := order.AmountFen - refundedBefore
	amountFen := req.AmountFen
	if amountFen <= 0 {
		amountFen = remainingFen
	}
	if amountFen <= 0 {
		return nil, nil, nil, ErrInvalidArgument
	}
	if remainingFen <= 0 || amountFen > remainingFen {
		return nil, nil, nil, ErrInvalidArgument
	}

	now := time.Now().UTC()
	actorID := strings.TrimSpace(req.ActorID)
	if actorID == "" {
		actorID = "admin"
	}
	destination := RefundDestinationForStrategy(s.refundSettings.DefaultStrategy, strings.TrimSpace(req.Destination))
	refund := &RefundTransaction{
		ID:             "rfd_" + shortHash(idempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      amountFen,
		Destination:    destination,
		Reason:         reason,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
	}

	var account *WalletAccount
	switch destination {
	case RefundDestinationOriginalRoute:
		refund.Status = RefundStatusPendingOriginal
		order.Status = refundOrderStatusAfter(order.Status, refundedBefore+amountFen, order.AmountFen, destination)
		order.UpdatedAt = now
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.refund.requested",
			ActorID:   actorID,
			Message:   "订单退款已提交原路返回处理",
			AmountFen: amountFen,
			CreatedAt: now,
		})
	default:
		refund.Destination = RefundDestinationBalance
		refund.Status = RefundStatusSuccess
		account = s.getOrCreateWalletLocked(order.UserID)
		account.Balance += amountFen
		account.Version++
		walletTransaction := s.createWalletTransactionLocked(order.UserID, order.ID, "refund", amountFen, RefundDestinationBalance, idempotencyKey)
		s.walletIdempotency[idempotencyKey] = walletTransaction
		order.Status = refundOrderStatusAfter(order.Status, refundedBefore+amountFen, order.AmountFen, destination)
		order.UpdatedAt = now
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.refund.success",
			ActorID:   actorID,
			Message:   "订单退款已退回平台余额",
			AmountFen: amountFen,
			CreatedAt: now,
		})
	}

	s.refundTransactions[refund.ID] = refund
	s.refundByIdempotency[idempotencyKey] = refund.ID
	return cloneRefundTransaction(refund), cloneOrder(order), cloneWalletAccount(account), nil
}

func (s *Store) AdminOperationsSnapshot(req AdminOperationsSnapshotRequest) (*AdminOperationsSnapshot, error) {
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	limit := normalizeAdminOperationsSnapshotLimit(req.Limit)
	_, cleanupGrace, cleanupNow, err := normalizeObjectStorageCleanupWindow(ObjectStorageCleanupCandidatesRequest{
		Limit:        limit,
		GraceSeconds: int64(req.ObjectCleanupGraceSeconds),
		Now:          now,
	})
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	counts := AdminOperationsCounts{}
	orders := make([]Order, 0, len(s.orders))
	for _, order := range s.orders {
		if order == nil {
			continue
		}
		counts.TotalOrders++
		switch order.Status {
		case StatusMerchantPending:
			counts.PendingMerchantOrders++
		case StatusDispatching:
			counts.DispatchingOrders++
		case StatusRiderAssigned, StatusPickedUp, StatusDelivering:
			counts.RiderAssignedOrders++
		case StatusCompleted:
			counts.CompletedOrders++
		case StatusRefunded:
			counts.RefundedOrders++
		}
		if isAdminExceptionOrder(order, now) {
			counts.ExceptionOrders++
		}
		orders = append(orders, *cloneOrder(order))
	}
	sort.SliceStable(orders, func(i, j int) bool {
		leftRisk := isAdminExceptionOrder(&orders[i], now)
		rightRisk := isAdminExceptionOrder(&orders[j], now)
		if leftRisk != rightRisk {
			return leftRisk
		}
		if !orders[i].CreatedAt.Equal(orders[j].CreatedAt) {
			return orders[i].CreatedAt.After(orders[j].CreatedAt)
		}
		return orders[i].ID < orders[j].ID
	})
	if len(orders) > limit {
		orders = orders[:limit]
	}

	merchants := make([]AdminMerchantSnapshot, 0, len(s.merchants))
	for merchantID := range s.merchants {
		profile := s.merchantProfileLocked(merchantID)
		if profile == nil {
			continue
		}
		counts.TotalMerchants++
		deposit := cloneDepositAccount(s.deposits[depositKey("merchant", merchantID)])
		if len(profile.MissingQualifications) > 0 || profile.QualificationPopupRequired {
			counts.MerchantQualificationRisks++
		}
		if deposit == nil || deposit.Status != DepositStatusPaid || profile.Account.DepositStatus != DepositStatusPaid {
			counts.MerchantDepositMissing++
		}
		merchantShops := make([]Shop, 0)
		for _, shop := range s.shops {
			if shop == nil || shop.MerchantID != merchantID {
				continue
			}
			cloned := cloneShop(shop)
			if !s.shopCanAcceptOrdersLocked(cloned.ID) {
				cloned.Status = ShopStatusQualificationExpired
			}
			merchantShops = append(merchantShops, *cloned)
		}
		sort.SliceStable(merchantShops, func(i, j int) bool {
			return merchantShops[i].ID < merchantShops[j].ID
		})
		merchants = append(merchants, AdminMerchantSnapshot{
			Account:                    profile.Account,
			Shops:                      merchantShops,
			Qualifications:             append([]MerchantQualification{}, profile.Qualifications...),
			MissingQualifications:      append([]string{}, profile.MissingQualifications...),
			StaffCount:                 len(profile.Staff),
			SupplementalMaterialCount:  len(profile.SupplementalMaterials),
			Deposit:                    deposit,
			CanAcceptOrders:            profile.CanAcceptOrders,
			QualificationPopupRequired: profile.QualificationPopupRequired,
			QualificationPopupCode:     profile.QualificationPopupCode,
			SupplementalMaterials:      append([]MerchantSupplementalMaterial{}, profile.SupplementalMaterials...),
		})
	}
	sort.SliceStable(merchants, func(i, j int) bool {
		leftRisk := adminMerchantRiskRank(merchants[i])
		rightRisk := adminMerchantRiskRank(merchants[j])
		if leftRisk != rightRisk {
			return leftRisk > rightRisk
		}
		return merchants[i].Account.ID < merchants[j].Account.ID
	})
	if len(merchants) > limit {
		merchants = merchants[:limit]
	}

	riderPerformance, err := s.riderPerformanceSnapshotLocked(req.StationManagerID)
	if err != nil {
		return nil, err
	}
	if len(riderPerformance) > limit {
		riderPerformance = riderPerformance[:limit]
	}

	riders := make([]RiderAccount, 0, len(s.riders))
	for _, rider := range s.riders {
		if rider == nil {
			continue
		}
		if rider.Type == RiderAccountStationManager {
			counts.StationManagers++
		}
		if rider.Type == RiderAccountRider {
			counts.TotalRiders++
			if rider.Online {
				counts.OnlineRiders++
			}
			if rider.DepositStatus != DepositStatusPaid && rider.DepositStatus != DepositStatusWechatExemptApproved {
				counts.RiderDepositMissing++
			}
		}
		riders = append(riders, *cloneRiderAccount(rider))
	}
	sort.SliceStable(riders, func(i, j int) bool {
		if riders[i].Type != riders[j].Type {
			return riders[i].Type == RiderAccountRider
		}
		if riders[i].Online != riders[j].Online {
			return riders[i].Online
		}
		if riders[i].DispatchPriority != riders[j].DispatchPriority {
			return riders[i].DispatchPriority > riders[j].DispatchPriority
		}
		return riders[i].ID < riders[j].ID
	})
	if len(riders) > limit {
		riders = riders[:limit]
	}

	afterSales := make([]AfterSalesRequest, 0, len(s.afterSalesRequests))
	for _, request := range s.afterSalesRequests {
		if request == nil {
			continue
		}
		switch request.Status {
		case AfterSalesPendingMerchant:
			counts.AfterSalesPending++
		case AfterSalesAdminReview:
			counts.AfterSalesAdminReview++
		}
		afterSales = append(afterSales, *s.afterSalesRequestViewLocked(request))
	}
	sortAfterSalesRequests(afterSales)
	if len(afterSales) > limit {
		afterSales = afterSales[:limit]
	}

	dispatchEvents := make([]DispatchEvent, 0, len(s.dispatchEvents))
	for _, event := range s.dispatchEvents {
		if event == nil {
			continue
		}
		counts.DispatchEventCount++
		dispatchEvents = append(dispatchEvents, *cloneDispatchEvent(event))
	}
	sort.SliceStable(dispatchEvents, func(i, j int) bool {
		if !dispatchEvents[i].CreatedAt.Equal(dispatchEvents[j].CreatedAt) {
			return dispatchEvents[i].CreatedAt.After(dispatchEvents[j].CreatedAt)
		}
		return dispatchEvents[i].ID < dispatchEvents[j].ID
	})
	if len(dispatchEvents) > limit {
		dispatchEvents = dispatchEvents[:limit]
	}

	outboxEvents := make([]OutboxEvent, 0, len(s.outboxEvents))
	for _, event := range s.outboxEvents {
		if event != nil {
			outboxEvents = append(outboxEvents, *cloneOutboxEvent(event))
		}
	}
	outboxStats := buildOutboxStats(outboxEvents, "", now, req.LeaseExpiringWithinSeconds)
	objectCleanupStats := s.objectStorageCleanupStatsLocked(cleanupNow, cleanupGrace)
	counts.OutboxReady = outboxStats.Ready
	counts.OutboxBlocked = outboxStats.Blocked
	counts.ObjectCleanupFailed = objectCleanupStats.Failed
	counts.ObjectCleanupTotalCandidate = objectCleanupStats.Pending

	refundSettings := s.refundSettings
	refundSettings.DefaultStrategy = NormalizeRefundStrategy(refundSettings.DefaultStrategy)
	return &AdminOperationsSnapshot{
		GeneratedAt:               now,
		Counts:                    counts,
		Orders:                    orders,
		Merchants:                 merchants,
		Riders:                    riders,
		RiderPerformance:          riderPerformance,
		AfterSales:                afterSales,
		DispatchEvents:            dispatchEvents,
		RefundSettings:            refundSettings,
		OutboxStats:               *outboxStats,
		ObjectStorageCleanupStats: objectCleanupStats,
	}, nil
}

func (s *Store) CreateAfterSales(req CreateAfterSalesRequest) (*AfterSalesRequest, error) {
	userID := strings.TrimSpace(req.UserID)
	orderID := strings.TrimSpace(req.OrderID)
	reason := strings.TrimSpace(req.Reason)
	requestType := normalizeAfterSalesType(req.Type)
	if userID == "" || orderID == "" || reason == "" || requestType == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if order.UserID != userID {
		return nil, ErrInvalidOrderState
	}
	amountFen := req.RequestedAmountFen
	if amountFen <= 0 {
		amountFen = s.refundableRemainingFenLocked(order.ID)
	}
	if requestType == AfterSalesRefundOnly && amountFen < order.AmountFen {
		requestType = AfterSalesPartialRefund
	}
	request := AfterSalesRequest{
		OrderID:            order.ID,
		UserID:             userID,
		Type:               requestType,
		Reason:             reason,
		RequestedAmountFen: amountFen,
		EvidenceURLs:       sanitizedStringSlice(req.EvidenceURLs),
	}
	if !CanCreateAfterSales(*order, request) {
		return nil, ErrInvalidOrderState
	}
	remainingFen := s.refundableRemainingFenLocked(order.ID)
	if remainingFen <= 0 || amountFen > remainingFen {
		return nil, ErrInvalidArgument
	}
	for _, existing := range s.afterSalesRequests {
		if existing == nil || existing.OrderID != order.ID {
			continue
		}
		switch existing.Status {
		case AfterSalesRejected, AfterSalesRefunded:
			continue
		default:
			return nil, ErrInvalidOrderState
		}
	}

	s.nextAfterSalesID++
	now := time.Now().UTC()
	request.ID = fmt.Sprintf("asr_%d", s.nextAfterSalesID)
	request.Status = AfterSalesPendingMerchant
	request.CreatedAt = now
	request.UpdatedAt = now
	s.afterSalesRequests[request.ID] = &request
	s.appendAfterSalesEventLocked(&request, AfterSalesActionCreated, userID, "user", "用户已提交售后申请", true, request.EvidenceURLs, now)
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "order.after_sales.created",
		ActorID:   userID,
		Message:   "用户已提交售后申请",
		CreatedAt: now,
	})
	return s.afterSalesRequestViewLocked(&request), nil
}

func (s *Store) AfterSalesEvents(requestID string, actorID string, actorRole string) ([]AfterSalesEvent, error) {
	requestID = strings.TrimSpace(requestID)
	actorID = strings.TrimSpace(actorID)
	actorRole = strings.TrimSpace(actorRole)
	if requestID == "" || actorID == "" || actorRole == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[requestID]
	if request == nil {
		return nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if !s.canAccessAfterSalesLocked(order, request, actorID, actorRole) {
		return nil, ErrInvalidOrderState
	}
	events := make([]AfterSalesEvent, 0)
	for _, event := range s.afterSalesEvents {
		if event == nil || event.RequestID != request.ID {
			continue
		}
		if actorRole != "admin" && !event.VisibleToUser {
			continue
		}
		events = append(events, *cloneAfterSalesEvent(event))
	}
	sortAfterSalesEvents(events)
	return events, nil
}

func (s *Store) AddAfterSalesEvent(req AddAfterSalesEventRequest) (*AfterSalesEvent, *AfterSalesRequest, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.Action = normalizeAfterSalesAction(req.Action)
	req.Message = strings.TrimSpace(req.Message)
	if req.RequestID == "" || req.ActorID == "" || req.ActorRole == "" || req.Action == "" || req.Message == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[req.RequestID]
	if request == nil {
		return nil, nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	if !s.canAddAfterSalesEventLocked(order, request, req.ActorID, req.ActorRole, req.Action) {
		return nil, nil, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	visible := true
	if req.VisibleToUser != nil {
		visible = *req.VisibleToUser
	}
	if req.ActorRole != "admin" {
		visible = true
	}
	if req.Action == AfterSalesActionInternalNote {
		visible = false
	}
	event := s.appendAfterSalesEventLocked(request, req.Action, req.ActorID, req.ActorRole, req.Message, visible, req.Attachments, now)
	orderEventType := "order.after_sales.event_added"
	orderEventMessage := "售后处理记录已更新"
	if afterSalesActionEscalates(req.Action) && request.Status != AfterSalesRefunded && request.Status != AfterSalesRejected {
		request.Status = AfterSalesAdminReview
		request.ReviewReason = req.Message
		request.ReviewerID = req.ActorID
		request.ReviewerRole = req.ActorRole
		request.ReviewedAt = now
		request.UpdatedAt = now
		orderEventType = "order.after_sales.escalated"
		orderEventMessage = "售后申请已转平台审核"
	} else {
		request.UpdatedAt = now
	}
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      orderEventType,
		ActorID:   req.ActorID,
		Message:   orderEventMessage,
		CreatedAt: now,
	})
	return cloneAfterSalesEvent(event), s.afterSalesRequestViewLocked(request), nil
}

func (s *Store) CreateAfterSalesEvidenceUpload(req CreateAfterSalesEvidenceUploadRequest) (*ObjectUploadTicket, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	if req.RequestID == "" || req.ActorID == "" || req.ActorRole == "" || req.FileName == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[req.RequestID]
	if request == nil {
		return nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if !s.canAccessAfterSalesLocked(order, request, req.ActorID, req.ActorRole) {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	storage := s.objectStorageConfigLocked()
	scanStatus := AfterSalesUploadScanNotRequired
	if storage.RequireScanApprovalForConfirm {
		scanStatus = AfterSalesUploadScanPending
	}
	expiresAt := now.Add(storage.TicketTTL)
	objectKey := fmt.Sprintf("after-sales/%s/%s/%s", request.ID, shortHash(fmt.Sprintf("%s:%s:%d", req.ActorID, req.FileName, now.UnixNano())), req.FileName)
	ticket, err := storage.createObjectUploadTicket(objectUploadTicketInput{
		ObjectKey:    objectKey,
		ContentType:  req.ContentType,
		SizeBytes:    req.SizeBytes,
		MaxSizeBytes: AfterSalesEvidenceMaxBytes,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}
	ticketID := "aset_" + shortHash(ticket.ObjectKey)
	ticket.TicketID = ticketID
	if s.afterSalesUploadTickets == nil {
		s.afterSalesUploadTickets = map[string]*AfterSalesEvidenceUploadTicket{}
	}
	s.afterSalesUploadTickets[ticketID] = &AfterSalesEvidenceUploadTicket{
		ID:             ticketID,
		RequestID:      request.ID,
		OrderID:        request.OrderID,
		Provider:       ticket.Provider,
		Bucket:         ticket.Bucket,
		ObjectKey:      ticket.ObjectKey,
		PublicURL:      ticket.PublicURL,
		FileName:       req.FileName,
		ContentType:    req.ContentType,
		SizeBytes:      req.SizeBytes,
		MaxSizeBytes:   ticket.MaxSizeBytes,
		UploadedByID:   req.ActorID,
		UploadedByRole: req.ActorRole,
		Status:         AfterSalesUploadTicketIssued,
		ScanStatus:     scanStatus,
		CreatedAt:      now,
		ExpiresAt:      ticket.ExpiresAt,
	}
	return ticket, nil
}

func (s *Store) ConfirmObjectStorageUpload(req ObjectStorageUploadCallbackRequest) (*AfterSalesEvidenceUploadTicket, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	req.ContentSHA = strings.TrimSpace(req.ContentSHA)
	req.Signature = strings.TrimSpace(req.Signature)
	if req.TicketID == "" || req.ObjectKey == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, ErrInvalidArgument
	}
	if req.UploadedAt.IsZero() {
		req.UploadedAt = time.Now().UTC()
	} else {
		req.UploadedAt = req.UploadedAt.UTC()
	}

	storage := s.objectStorageSnapshot()
	if err := storage.verifyObjectUploadCallback(objectUploadCallbackSignatureInput{
		TicketID:    req.TicketID,
		ObjectKey:   req.ObjectKey,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
		ContentSHA:  req.ContentSHA,
		UploadedAt:  req.UploadedAt,
	}, req.Signature); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if ticket := s.afterSalesUploadTickets[req.TicketID]; ticket != nil {
		if !afterSalesUploadTicketMatchesObjectCallback(ticket, req) {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			return cloneAfterSalesEvidenceUploadTicket(ticket), nil
		}
		if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		if time.Now().UTC().After(ticket.ExpiresAt) {
			return nil, ErrInvalidOrderState
		}
		ticket.Status = AfterSalesUploadTicketUploaded
		ticket.ContentSHA = req.ContentSHA
		ticket.UploadedAt = req.UploadedAt
		if storage.RequireScanApprovalForConfirm {
			ticket.ScanStatus = AfterSalesUploadScanPending
		} else {
			ticket.ScanStatus = AfterSalesUploadScanNotRequired
		}
		return cloneAfterSalesEvidenceUploadTicket(ticket), nil
	}
	if ticket := s.reviewImageTickets[req.TicketID]; ticket != nil {
		if !reviewImageUploadTicketMatchesObjectCallback(ticket, req) {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			return reviewTicketObjectStorageView(ticket), nil
		}
		if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		if time.Now().UTC().After(ticket.ExpiresAt) {
			return nil, ErrInvalidOrderState
		}
		ticket.Status = AfterSalesUploadTicketUploaded
		ticket.ContentSHA = req.ContentSHA
		ticket.UploadedAt = req.UploadedAt
		if storage.RequireScanApprovalForConfirm {
			ticket.ScanStatus = AfterSalesUploadScanPending
		} else {
			ticket.ScanStatus = AfterSalesUploadScanNotRequired
		}
		return reviewTicketObjectStorageView(ticket), nil
	}
	if ticket := s.prescriptionImageTickets[req.TicketID]; ticket != nil {
		if !prescriptionImageUploadTicketMatchesObjectCallback(ticket, req) {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			return prescriptionTicketObjectStorageView(ticket), nil
		}
		if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		if time.Now().UTC().After(ticket.ExpiresAt) {
			return nil, ErrInvalidOrderState
		}
		ticket.Status = AfterSalesUploadTicketUploaded
		ticket.ContentSHA = req.ContentSHA
		ticket.UploadedAt = req.UploadedAt
		if storage.RequireScanApprovalForConfirm {
			ticket.ScanStatus = AfterSalesUploadScanPending
		} else {
			ticket.ScanStatus = AfterSalesUploadScanNotRequired
		}
		return prescriptionTicketObjectStorageView(ticket), nil
	}
	return nil, ErrInvalidArgument
}

func (s *Store) RecordObjectStorageScanResult(req ObjectStorageScanResultRequest) (*AfterSalesEvidenceUploadTicket, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.ScanStatus = normalizeAfterSalesUploadScanStatus(req.ScanStatus)
	req.ScanResult = strings.TrimSpace(req.ScanResult)
	req.Scanner = strings.TrimSpace(req.Scanner)
	req.Signature = strings.TrimSpace(req.Signature)
	if req.TicketID == "" || req.ObjectKey == "" || (req.ScanStatus != AfterSalesUploadScanPassed && req.ScanStatus != AfterSalesUploadScanRejected) {
		return nil, ErrInvalidArgument
	}
	if req.ScanCheckedAt.IsZero() {
		req.ScanCheckedAt = time.Now().UTC()
	} else {
		req.ScanCheckedAt = req.ScanCheckedAt.UTC()
	}

	storage := s.objectStorageSnapshot()
	if err := storage.verifyObjectScanResult(objectScanResultSignatureInput{
		TicketID:      req.TicketID,
		ObjectKey:     req.ObjectKey,
		ScanStatus:    req.ScanStatus,
		ScanResult:    req.ScanResult,
		Scanner:       req.Scanner,
		ScanCheckedAt: req.ScanCheckedAt,
	}, req.Signature); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if ticket := s.afterSalesUploadTickets[req.TicketID]; ticket != nil {
		if ticket.ObjectKey != req.ObjectKey {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			if ticket.ScanStatus == req.ScanStatus && ticket.ScanResult == req.ScanResult {
				return cloneAfterSalesEvidenceUploadTicket(ticket), nil
			}
			return nil, ErrInvalidOrderState
		}
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		ticket.ScanStatus = req.ScanStatus
		ticket.ScanResult = req.ScanResult
		ticket.ScanCheckedAt = req.ScanCheckedAt
		return cloneAfterSalesEvidenceUploadTicket(ticket), nil
	}
	if ticket := s.reviewImageTickets[req.TicketID]; ticket != nil {
		if ticket.ObjectKey != req.ObjectKey {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			if ticket.ScanStatus == req.ScanStatus && ticket.ScanResult == req.ScanResult {
				return reviewTicketObjectStorageView(ticket), nil
			}
			return nil, ErrInvalidOrderState
		}
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		ticket.ScanStatus = req.ScanStatus
		ticket.ScanResult = req.ScanResult
		ticket.ScanCheckedAt = req.ScanCheckedAt
		return reviewTicketObjectStorageView(ticket), nil
	}
	if ticket := s.prescriptionImageTickets[req.TicketID]; ticket != nil {
		if ticket.ObjectKey != req.ObjectKey {
			return nil, ErrInvalidArgument
		}
		if ticket.Status == AfterSalesUploadTicketConfirmed {
			if ticket.ScanStatus == req.ScanStatus && ticket.ScanResult == req.ScanResult {
				return prescriptionTicketObjectStorageView(ticket), nil
			}
			return nil, ErrInvalidOrderState
		}
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return nil, ErrInvalidOrderState
		}
		ticket.ScanStatus = req.ScanStatus
		ticket.ScanResult = req.ScanResult
		ticket.ScanCheckedAt = req.ScanCheckedAt
		return prescriptionTicketObjectStorageView(ticket), nil
	}
	return nil, ErrInvalidArgument
}

func (s *Store) ObjectStorageCleanupCandidates(req ObjectStorageCleanupCandidatesRequest) ([]ObjectStorageCleanupCandidate, error) {
	limit, grace, now, err := normalizeObjectStorageCleanupWindow(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	candidates := make([]ObjectStorageCleanupCandidate, 0)
	for _, ticket := range s.afterSalesUploadTickets {
		candidate, ok := objectStorageCleanupCandidateFromTicket(ticket, now, grace)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	sortObjectStorageCleanupCandidates(candidates)
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

func (s *Store) ObjectStorageCleanupStats(req ObjectStorageCleanupCandidatesRequest) (*ObjectStorageCleanupStats, error) {
	_, grace, now, err := normalizeObjectStorageCleanupWindow(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	stats := s.objectStorageCleanupStatsLocked(now, grace)
	return &stats, nil
}

func (s *Store) objectStorageCleanupStatsLocked(now time.Time, grace time.Duration) ObjectStorageCleanupStats {
	stats := ObjectStorageCleanupStats{}
	for _, ticket := range s.afterSalesUploadTickets {
		if ticket == nil {
			continue
		}
		stats.CleanupAttempts += ticket.CleanupAttempts
		if ticket.Status == AfterSalesUploadTicketDeleted {
			stats.Deleted++
			if !ticket.DeletedAt.IsZero() && ticket.DeletedAt.After(stats.LastDeletedAt) {
				stats.LastDeletedAt = ticket.DeletedAt
			}
			continue
		}
		if ticket.Status != AfterSalesUploadTicketConfirmed && ticket.LastCleanupError != "" {
			stats.Failed++
			if !ticket.LastCleanupFailedAt.IsZero() && ticket.LastCleanupFailedAt.After(stats.LastCleanupFailedAt) {
				stats.LastCleanupFailedAt = ticket.LastCleanupFailedAt
			}
		}
		candidate, ok := objectStorageCleanupCandidateFromTicket(ticket, now, grace)
		if !ok {
			continue
		}
		stats.Pending++
		switch candidate.Reason {
		case AfterSalesObjectCleanupExpired:
			stats.ExpiredUnconfirmed++
		case AfterSalesObjectCleanupRejected:
			stats.ScanRejected++
		}
	}
	return stats
}

func (s *Store) CompleteObjectStorageCleanup(req ObjectStorageCleanupCompleteRequest) (*AfterSalesEvidenceUploadTicket, error) {
	normalized, err := normalizeObjectStorageCleanupCompleteRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.completeObjectStorageCleanupLocked(normalized)
}

func (s *Store) CompleteObjectStorageCleanupWithAudit(req ObjectStorageCleanupCompleteRequest, audit RecordAuditLogRequest) (*AfterSalesEvidenceUploadTicket, *AuditLog, error) {
	normalized, err := normalizeObjectStorageCleanupCompleteRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, err
	}
	if log.Action != "admin.object_cleanup.completed" || log.TargetType != "object_storage_ticket" || log.TargetID != normalized.TicketID {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, err := s.completeObjectStorageCleanupLocked(normalized)
	if err != nil {
		return nil, nil, err
	}
	log.Payload = objectStorageCleanupCompletedAuditPayload(ticket)
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return ticket, cloneAuditLog(log), nil
}

func normalizeObjectStorageCleanupCompleteRequest(req ObjectStorageCleanupCompleteRequest) (ObjectStorageCleanupCompleteRequest, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.Reason = normalizeObjectStorageCleanupReason(req.Reason)
	if req.TicketID == "" || req.ObjectKey == "" || req.Reason == "" {
		return req, ErrInvalidArgument
	}
	if req.DeletedAt.IsZero() {
		req.DeletedAt = time.Now().UTC()
	} else {
		req.DeletedAt = req.DeletedAt.UTC()
	}
	return req, nil
}

func (s *Store) completeObjectStorageCleanupLocked(req ObjectStorageCleanupCompleteRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket := s.afterSalesUploadTickets[req.TicketID]
	if ticket == nil || ticket.ObjectKey != req.ObjectKey {
		return nil, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return nil, ErrInvalidOrderState
	}
	if ticket.Status == AfterSalesUploadTicketDeleted {
		if ticket.CleanupReason == req.Reason || ticket.CleanupReason == "" {
			return cloneAfterSalesEvidenceUploadTicket(ticket), nil
		}
		return nil, ErrInvalidOrderState
	}
	ticket.Status = AfterSalesUploadTicketDeleted
	ticket.CleanupReason = req.Reason
	ticket.DeletedAt = req.DeletedAt
	ticket.LastCleanupError = ""
	ticket.LastCleanupFailedAt = time.Time{}
	return cloneAfterSalesEvidenceUploadTicket(ticket), nil
}

func (s *Store) RecordObjectStorageCleanupFailure(req ObjectStorageCleanupFailureRequest) (*AfterSalesEvidenceUploadTicket, error) {
	normalized, err := normalizeObjectStorageCleanupFailureRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recordObjectStorageCleanupFailureLocked(normalized)
}

func (s *Store) RecordObjectStorageCleanupFailureWithAudit(req ObjectStorageCleanupFailureRequest, audit RecordAuditLogRequest) (*AfterSalesEvidenceUploadTicket, *AuditLog, error) {
	normalized, err := normalizeObjectStorageCleanupFailureRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, err
	}
	if log.Action != "admin.object_cleanup.failed" || log.TargetType != "object_storage_ticket" || log.TargetID != normalized.TicketID {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, err := s.recordObjectStorageCleanupFailureLocked(normalized)
	if err != nil {
		return nil, nil, err
	}
	log.Payload = objectStorageCleanupFailedAuditPayload(ticket)
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return ticket, cloneAuditLog(log), nil
}

func normalizeObjectStorageCleanupFailureRequest(req ObjectStorageCleanupFailureRequest) (ObjectStorageCleanupFailureRequest, error) {
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.Reason = normalizeObjectStorageCleanupReason(req.Reason)
	req.Error = sanitizeObjectStorageCleanupError(req.Error)
	if req.TicketID == "" || req.ObjectKey == "" || req.Reason == "" || req.Error == "" {
		return req, ErrInvalidArgument
	}
	if req.FailedAt.IsZero() {
		req.FailedAt = time.Now().UTC()
	} else {
		req.FailedAt = req.FailedAt.UTC()
	}
	return req, nil
}

func (s *Store) recordObjectStorageCleanupFailureLocked(req ObjectStorageCleanupFailureRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket := s.afterSalesUploadTickets[req.TicketID]
	if ticket == nil || ticket.ObjectKey != req.ObjectKey {
		return nil, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return nil, ErrInvalidOrderState
	}
	if ticket.Status == AfterSalesUploadTicketDeleted {
		return cloneAfterSalesEvidenceUploadTicket(ticket), nil
	}
	ticket.CleanupAttempts++
	ticket.CleanupReason = req.Reason
	ticket.LastCleanupError = req.Error
	ticket.LastCleanupFailedAt = req.FailedAt
	return cloneAfterSalesEvidenceUploadTicket(ticket), nil
}

func objectStorageCleanupCompletedAuditPayload(ticket *AfterSalesEvidenceUploadTicket) map[string]any {
	if ticket == nil {
		return map[string]any{}
	}
	return map[string]any{
		"object_key": strings.TrimSpace(ticket.ObjectKey),
		"reason":     strings.TrimSpace(ticket.CleanupReason),
		"status":     strings.TrimSpace(ticket.Status),
	}
}

func objectStorageCleanupFailedAuditPayload(ticket *AfterSalesEvidenceUploadTicket) map[string]any {
	if ticket == nil {
		return map[string]any{}
	}
	return map[string]any{
		"object_key":       strings.TrimSpace(ticket.ObjectKey),
		"reason":           strings.TrimSpace(ticket.CleanupReason),
		"status":           strings.TrimSpace(ticket.Status),
		"cleanup_attempts": ticket.CleanupAttempts,
	}
}

func (s *Store) ConfirmAfterSalesEvidenceUpload(req ConfirmAfterSalesEvidenceUploadRequest) (*AfterSalesEvidence, *AfterSalesEvent, *AfterSalesRequest, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.TicketID = strings.TrimSpace(req.TicketID)
	req.ObjectKey = strings.TrimSpace(req.ObjectKey)
	req.FileName = sanitizeObjectFileName(req.FileName)
	req.ContentType = normalizeEvidenceContentType(req.ContentType)
	req.ContentSHA = strings.TrimSpace(req.ContentSHA)
	req.Message = strings.TrimSpace(req.Message)
	if req.RequestID == "" || req.ActorID == "" || req.ActorRole == "" || req.ObjectKey == "" || req.ContentType == "" || req.SizeBytes <= 0 || req.SizeBytes > AfterSalesEvidenceMaxBytes {
		return nil, nil, nil, ErrInvalidArgument
	}
	if !validAfterSalesEvidenceObjectKey(req.RequestID, req.ObjectKey) {
		return nil, nil, nil, ErrInvalidArgument
	}
	if req.FileName == "" {
		req.FileName = sanitizeObjectFileName(objectKeyFileName(req.ObjectKey))
	}
	if req.FileName == "" {
		return nil, nil, nil, ErrInvalidArgument
	}

	ticket, existing, requestView, storage, err := s.prepareAfterSalesEvidenceUploadConfirmation(req)
	if err != nil || existing != nil {
		return existing, nil, requestView, err
	}
	if err := storage.verifyUploadedObject(objectHeadCheckInput{
		ObjectKey:   ticket.ObjectKey,
		ContentType: ticket.ContentType,
		SizeBytes:   ticket.SizeBytes,
	}); err != nil {
		return nil, nil, nil, err
	}
	return s.commitAfterSalesEvidenceUploadConfirmation(req)
}

func (s *Store) prepareAfterSalesEvidenceUploadConfirmation(req ConfirmAfterSalesEvidenceUploadRequest) (*AfterSalesEvidenceUploadTicket, *AfterSalesEvidence, *AfterSalesRequest, ObjectStorageConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[req.RequestID]
	if request == nil {
		return nil, nil, nil, ObjectStorageConfig{}, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, nil, nil, ObjectStorageConfig{}, ErrNotFound
	}
	if !s.canAccessAfterSalesLocked(order, request, req.ActorID, req.ActorRole) {
		return nil, nil, nil, ObjectStorageConfig{}, ErrInvalidOrderState
	}
	evidenceID := "ase_" + shortHash(req.ObjectKey)
	if existing := s.afterSalesEvidence[evidenceID]; existing != nil {
		if existing.RequestID != request.ID {
			return nil, nil, nil, ObjectStorageConfig{}, ErrInvalidArgument
		}
		return nil, cloneAfterSalesEvidence(existing), s.afterSalesRequestViewLocked(request), ObjectStorageConfig{}, nil
	}

	now := time.Now().UTC()
	ticket := s.afterSalesUploadTicketForConfirmLocked(req)
	if ticket == nil {
		return nil, nil, nil, ObjectStorageConfig{}, ErrInvalidArgument
	}
	storage := s.objectStorageConfigLocked()
	if err := afterSalesUploadTicketConfirmReady(ticket, storage, now); err != nil {
		return nil, nil, nil, ObjectStorageConfig{}, err
	}
	return cloneAfterSalesEvidenceUploadTicket(ticket), nil, nil, storage, nil
}

func (s *Store) commitAfterSalesEvidenceUploadConfirmation(req ConfirmAfterSalesEvidenceUploadRequest) (*AfterSalesEvidence, *AfterSalesEvent, *AfterSalesRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[req.RequestID]
	if request == nil {
		return nil, nil, nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, nil, nil, ErrNotFound
	}
	if !s.canAccessAfterSalesLocked(order, request, req.ActorID, req.ActorRole) {
		return nil, nil, nil, ErrInvalidOrderState
	}
	evidenceID := "ase_" + shortHash(req.ObjectKey)
	if existing := s.afterSalesEvidence[evidenceID]; existing != nil {
		if existing.RequestID != request.ID {
			return nil, nil, nil, ErrInvalidArgument
		}
		return cloneAfterSalesEvidence(existing), nil, s.afterSalesRequestViewLocked(request), nil
	}

	now := time.Now().UTC()
	ticket := s.afterSalesUploadTicketForConfirmLocked(req)
	if ticket == nil {
		return nil, nil, nil, ErrInvalidArgument
	}
	if err := afterSalesUploadTicketConfirmReady(ticket, s.objectStorageConfigLocked(), now); err != nil {
		return nil, nil, nil, err
	}
	publicURL := ticket.PublicURL
	evidence := &AfterSalesEvidence{
		ID:             evidenceID,
		RequestID:      request.ID,
		OrderID:        request.OrderID,
		ObjectKey:      req.ObjectKey,
		PublicURL:      publicURL,
		FileName:       req.FileName,
		ContentType:    req.ContentType,
		SizeBytes:      req.SizeBytes,
		ContentSHA:     req.ContentSHA,
		UploadedByID:   req.ActorID,
		UploadedByRole: req.ActorRole,
		Status:         AfterSalesEvidenceUploaded,
		CreatedAt:      now,
		ConfirmedAt:    now,
	}
	ticket.Status = AfterSalesUploadTicketConfirmed
	ticket.ConfirmedAt = now
	if s.afterSalesEvidence == nil {
		s.afterSalesEvidence = map[string]*AfterSalesEvidence{}
	}
	s.afterSalesEvidence[evidence.ID] = evidence
	request.EvidenceURLs = sanitizedStringSlice(append(request.EvidenceURLs, publicURL))
	request.UpdatedAt = now
	message := req.Message
	if message == "" {
		message = "售后证据已上传"
	}
	event := s.appendAfterSalesEventLocked(request, AfterSalesActionEvidenceUploaded, req.ActorID, req.ActorRole, message, true, []string{publicURL}, now)
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "order.after_sales.evidence_uploaded",
		ActorID:   req.ActorID,
		Message:   "售后证据已上传",
		CreatedAt: now,
	})
	return cloneAfterSalesEvidence(evidence), cloneAfterSalesEvent(event), s.afterSalesRequestViewLocked(request), nil
}

func (s *Store) AfterSalesEvidence(requestID string, actorID string, actorRole string) ([]AfterSalesEvidence, error) {
	requestID = strings.TrimSpace(requestID)
	actorID = strings.TrimSpace(actorID)
	actorRole = strings.TrimSpace(actorRole)
	if requestID == "" || actorID == "" || actorRole == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[requestID]
	if request == nil {
		return nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if !s.canAccessAfterSalesLocked(order, request, actorID, actorRole) {
		return nil, ErrInvalidOrderState
	}
	evidence := make([]AfterSalesEvidence, 0)
	for _, item := range s.afterSalesEvidence {
		if item == nil || item.RequestID != requestID {
			continue
		}
		evidence = append(evidence, *cloneAfterSalesEvidence(item))
	}
	sortAfterSalesEvidence(evidence)
	return evidence, nil
}

func (s *Store) UserAfterSalesRequests(req AfterSalesListRequest) ([]AfterSalesRequest, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.OrderID = strings.TrimSpace(req.OrderID)
	if req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	requests := make([]AfterSalesRequest, 0)
	for _, request := range s.afterSalesRequests {
		if request == nil || request.UserID != req.UserID {
			continue
		}
		if req.OrderID != "" && request.OrderID != req.OrderID {
			continue
		}
		requests = append(requests, *s.afterSalesRequestViewLocked(request))
	}
	sortAfterSalesRequests(requests)
	return requests, nil
}

func (s *Store) MerchantAfterSalesRequests(merchantID string) ([]AfterSalesRequest, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	requests := make([]AfterSalesRequest, 0)
	for _, request := range s.afterSalesRequests {
		if request == nil {
			continue
		}
		if !s.orderBelongsToMerchantLocked(s.orders[request.OrderID], merchantID) {
			continue
		}
		requests = append(requests, *s.afterSalesRequestViewLocked(request))
	}
	sortAfterSalesRequests(requests)
	return requests, nil
}

func (s *Store) AdminAfterSalesRequests(req AfterSalesListRequest) ([]AfterSalesRequest, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Status = strings.TrimSpace(req.Status)
	s.mu.Lock()
	defer s.mu.Unlock()
	requests := make([]AfterSalesRequest, 0, len(s.afterSalesRequests))
	for _, request := range s.afterSalesRequests {
		if request == nil {
			continue
		}
		if req.RequestID != "" && request.ID != req.RequestID {
			continue
		}
		if req.OrderID != "" && request.OrderID != req.OrderID {
			continue
		}
		if req.Status != "" && request.Status != req.Status {
			continue
		}
		requests = append(requests, *s.afterSalesRequestViewLocked(request))
	}
	sortAfterSalesRequests(requests)
	return requests, nil
}

func (s *Store) AdminOrderDetail(orderID string) (*AdminOrderDetail, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}

	orderCopy := cloneOrder(order)
	if orderCopy == nil {
		return nil, ErrNotFound
	}
	orderCopy.Reviewed = s.userReviewedOrderLocked(order.UserID, order.ID)

	detail := &AdminOrderDetail{
		Order: *orderCopy,
	}
	for _, request := range s.afterSalesRequests {
		if request == nil || request.OrderID != orderID {
			continue
		}
		cloned := s.afterSalesRequestViewLocked(request)
		if cloned == nil {
			continue
		}
		detail.AfterSalesRequests = append(detail.AfterSalesRequests, *cloned)
		detail.AfterSalesSummary.Total++
		switch cloned.Status {
		case AfterSalesPendingMerchant, AfterSalesAdminReview:
			detail.AfterSalesSummary.OpenCount++
		case AfterSalesRefunded:
			detail.AfterSalesSummary.RefundedCount++
		}
		if detail.AfterSalesSummary.LatestUpdatedAt.IsZero() || cloned.UpdatedAt.After(detail.AfterSalesSummary.LatestUpdatedAt) {
			detail.AfterSalesSummary.LatestUpdatedAt = cloned.UpdatedAt
			detail.AfterSalesSummary.LatestStatus = cloned.Status
		}
	}
	sort.SliceStable(detail.AfterSalesRequests, func(i, j int) bool {
		if detail.AfterSalesRequests[i].UpdatedAt.Equal(detail.AfterSalesRequests[j].UpdatedAt) {
			return detail.AfterSalesRequests[i].ID > detail.AfterSalesRequests[j].ID
		}
		return detail.AfterSalesRequests[i].UpdatedAt.After(detail.AfterSalesRequests[j].UpdatedAt)
	})
	for _, refund := range s.refundTransactions {
		if refund == nil || refund.OrderID != orderID {
			continue
		}
		cloned := cloneRefundTransaction(refund)
		if cloned == nil {
			continue
		}
		detail.Refunds = append(detail.Refunds, *cloned)
		detail.RefundSummary.Total++
		detail.RefundSummary.TotalAmountFen += cloned.AmountFen
		if cloned.Status == RefundStatusSuccess {
			detail.RefundSummary.SuccessCount++
		}
		if detail.RefundSummary.LatestCreatedAt.IsZero() || cloned.CreatedAt.After(detail.RefundSummary.LatestCreatedAt) {
			detail.RefundSummary.LatestCreatedAt = cloned.CreatedAt
			detail.RefundSummary.LatestDestination = cloned.Destination
		}
	}
	sort.SliceStable(detail.Refunds, func(i, j int) bool {
		if detail.Refunds[i].CreatedAt.Equal(detail.Refunds[j].CreatedAt) {
			return detail.Refunds[i].ID < detail.Refunds[j].ID
		}
		return detail.Refunds[i].CreatedAt.Before(detail.Refunds[j].CreatedAt)
	})
	now := time.Now().UTC()
	for _, ticket := range s.serviceTickets {
		if ticket == nil || ticket.RelatedOrderID != orderID {
			continue
		}
		s.syncServiceTicketSLAStatusLocked(ticket, now)
		cloned := cloneServiceTicket(ticket)
		if cloned == nil {
			continue
		}
		detail.ServiceTickets = append(detail.ServiceTickets, *cloned)
		detail.ServiceTicketSummary.Total++
		if cloned.Status != ServiceTicketStatusClosed && cloned.Status != ServiceTicketStatusResolved {
			detail.ServiceTicketSummary.OpenCount++
		}
		if cloned.SLAStatus == ServiceTicketSLAStatusEscalated {
			detail.ServiceTicketSummary.EscalatedCount++
		}
		if detail.ServiceTicketSummary.LatestUpdatedAt.IsZero() || cloned.UpdatedAt.After(detail.ServiceTicketSummary.LatestUpdatedAt) {
			detail.ServiceTicketSummary.LatestUpdatedAt = cloned.UpdatedAt
			detail.ServiceTicketSummary.LatestStatus = cloned.Status
		}
	}
	sort.SliceStable(detail.ServiceTickets, func(i, j int) bool {
		if detail.ServiceTickets[i].UpdatedAt.Equal(detail.ServiceTickets[j].UpdatedAt) {
			return detail.ServiceTickets[i].ID < detail.ServiceTickets[j].ID
		}
		return detail.ServiceTickets[i].UpdatedAt.After(detail.ServiceTickets[j].UpdatedAt)
	})
	for _, event := range s.dispatchEvents {
		if event == nil || event.OrderID != orderID {
			continue
		}
		cloned := cloneDispatchEvent(event)
		if cloned == nil {
			continue
		}
		detail.DispatchEvents = append(detail.DispatchEvents, *cloned)
		detail.DispatchSummary.Total++
		switch cloned.Mode {
		case DispatchModeAutoAssign:
			detail.DispatchSummary.AutoAssignCount++
		case DispatchModeManualAssign:
			detail.DispatchSummary.ManualAssignCount++
		}
		switch cloned.Type {
		case "dispatch.rejected":
			detail.DispatchSummary.RejectCount++
		case "dispatch.timeout":
			detail.DispatchSummary.TimeoutCount++
		}
		if detail.DispatchSummary.LatestEventAt.IsZero() || cloned.CreatedAt.After(detail.DispatchSummary.LatestEventAt) {
			detail.DispatchSummary.LatestEventAt = cloned.CreatedAt
			detail.DispatchSummary.LatestType = cloned.Type
		}
	}
	sort.SliceStable(detail.DispatchEvents, func(i, j int) bool {
		if detail.DispatchEvents[i].CreatedAt.Equal(detail.DispatchEvents[j].CreatedAt) {
			return detail.DispatchEvents[i].ID < detail.DispatchEvents[j].ID
		}
		return detail.DispatchEvents[i].CreatedAt.Before(detail.DispatchEvents[j].CreatedAt)
	})
	auditTargets := map[string]struct{}{
		"order:" + orderID: {},
	}
	for _, request := range detail.AfterSalesRequests {
		auditTargets["after_sales:"+request.ID] = struct{}{}
	}
	for _, ticket := range detail.ServiceTickets {
		auditTargets["service_ticket:"+ticket.ID] = struct{}{}
	}
	for _, auditLog := range s.auditLogs {
		if auditLog == nil {
			continue
		}
		if _, ok := auditTargets[auditLog.TargetType+":"+auditLog.TargetID]; !ok {
			continue
		}
		cloned := cloneAuditLog(auditLog)
		if cloned == nil {
			continue
		}
		cloned.IntegrityVerified = verifyAuditLogIntegrity(*cloned, s.auditLogSigningSecret)
		detail.RelatedAudits = append(detail.RelatedAudits, *cloned)
		detail.AuditSummary.Total++
		if cloned.IntegrityVerified {
			detail.AuditSummary.VerifiedCount++
		}
		switch cloned.TargetType {
		case "order":
			detail.AuditSummary.OrderCount++
		case "after_sales":
			detail.AuditSummary.AfterSalesCount++
		case "service_ticket":
			detail.AuditSummary.ServiceTicketCount++
		}
		if detail.AuditSummary.LatestCreatedAt.IsZero() || cloned.CreatedAt.After(detail.AuditSummary.LatestCreatedAt) {
			detail.AuditSummary.LatestCreatedAt = cloned.CreatedAt
			detail.AuditSummary.LatestAction = cloned.Action
		}
	}
	sort.SliceStable(detail.RelatedAudits, func(i, j int) bool {
		if detail.RelatedAudits[i].CreatedAt.Equal(detail.RelatedAudits[j].CreatedAt) {
			return detail.RelatedAudits[i].ID > detail.RelatedAudits[j].ID
		}
		return detail.RelatedAudits[i].CreatedAt.After(detail.RelatedAudits[j].CreatedAt)
	})
	if len(detail.RelatedAudits) > 20 {
		detail.RelatedAudits = detail.RelatedAudits[:20]
	}
	return detail, nil
}

func (s *Store) AdminAfterSalesDetail(requestID string) (*AdminAfterSalesDetail, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[requestID]
	if request == nil {
		return nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, ErrNotFound
	}
	detail := &AdminAfterSalesDetail{
		Request: *s.afterSalesRequestViewLocked(request),
	}
	for _, event := range s.afterSalesEvents {
		if event == nil || event.RequestID != requestID {
			continue
		}
		cloned := cloneAfterSalesEvent(event)
		if cloned == nil {
			continue
		}
		detail.Events = append(detail.Events, *cloned)
		detail.EventSummary.Total++
		detail.EventSummary.AttachmentCount += len(cloned.Attachments)
		if cloned.VisibleToUser {
			detail.EventSummary.UserVisible++
		} else {
			detail.EventSummary.InternalOnly++
		}
		if detail.EventSummary.LatestEventAt.IsZero() || cloned.CreatedAt.After(detail.EventSummary.LatestEventAt) {
			detail.EventSummary.LatestEventAt = cloned.CreatedAt
			detail.EventSummary.LatestAction = cloned.Action
		}
	}
	sortAfterSalesEvents(detail.Events)
	for _, item := range s.afterSalesEvidence {
		if item == nil || item.RequestID != requestID {
			continue
		}
		cloned := cloneAfterSalesEvidence(item)
		if cloned == nil {
			continue
		}
		detail.Evidence = append(detail.Evidence, *cloned)
		detail.EvidenceSummary.Total++
		detail.EvidenceSummary.TotalSizeBytes += cloned.SizeBytes
		if strings.HasPrefix(cloned.ContentType, "image/") {
			detail.EvidenceSummary.ImageCount++
		}
		if !cloned.ConfirmedAt.IsZero() {
			detail.EvidenceSummary.ConfirmedCount++
			if detail.EvidenceSummary.LatestConfirmedAt.IsZero() || cloned.ConfirmedAt.After(detail.EvidenceSummary.LatestConfirmedAt) {
				detail.EvidenceSummary.LatestConfirmedAt = cloned.ConfirmedAt
			}
		}
	}
	sortAfterSalesEvidence(detail.Evidence)
	for _, event := range s.dispatchEvents {
		if event == nil || event.OrderID != order.ID {
			continue
		}
		cloned := cloneDispatchEvent(event)
		if cloned == nil {
			continue
		}
		detail.DispatchEvents = append(detail.DispatchEvents, *cloned)
		detail.DispatchSummary.Total++
		switch cloned.Mode {
		case "auto_assign":
			detail.DispatchSummary.AutoAssignCount++
		case "manual_assign":
			detail.DispatchSummary.ManualAssignCount++
		}
		switch cloned.Type {
		case "dispatch.rejected":
			detail.DispatchSummary.RejectCount++
		case "dispatch.timeout":
			detail.DispatchSummary.TimeoutCount++
		}
		if detail.DispatchSummary.LatestEventAt.IsZero() || cloned.CreatedAt.After(detail.DispatchSummary.LatestEventAt) {
			detail.DispatchSummary.LatestEventAt = cloned.CreatedAt
			detail.DispatchSummary.LatestType = cloned.Type
		}
	}
	sort.SliceStable(detail.DispatchEvents, func(i, j int) bool {
		if detail.DispatchEvents[i].CreatedAt.Equal(detail.DispatchEvents[j].CreatedAt) {
			return detail.DispatchEvents[i].ID < detail.DispatchEvents[j].ID
		}
		return detail.DispatchEvents[i].CreatedAt.Before(detail.DispatchEvents[j].CreatedAt)
	})
	for _, refund := range s.refundTransactions {
		if refund == nil || refund.OrderID != order.ID {
			continue
		}
		cloned := cloneRefundTransaction(refund)
		if cloned == nil {
			continue
		}
		detail.Refunds = append(detail.Refunds, *cloned)
		detail.RefundSummary.Total++
		detail.RefundSummary.TotalAmountFen += cloned.AmountFen
		if cloned.Status == RefundStatusSuccess {
			detail.RefundSummary.SuccessCount++
		}
		if detail.RefundSummary.LatestCreatedAt.IsZero() || cloned.CreatedAt.After(detail.RefundSummary.LatestCreatedAt) {
			detail.RefundSummary.LatestCreatedAt = cloned.CreatedAt
			detail.RefundSummary.LatestDestination = cloned.Destination
		}
	}
	sort.SliceStable(detail.Refunds, func(i, j int) bool {
		if detail.Refunds[i].CreatedAt.Equal(detail.Refunds[j].CreatedAt) {
			return detail.Refunds[i].ID < detail.Refunds[j].ID
		}
		return detail.Refunds[i].CreatedAt.Before(detail.Refunds[j].CreatedAt)
	})
	now := time.Now().UTC()
	for _, ticket := range s.serviceTickets {
		if ticket == nil || ticket.RelatedOrderID != order.ID {
			continue
		}
		s.syncServiceTicketSLAStatusLocked(ticket, now)
		cloned := cloneServiceTicket(ticket)
		if cloned == nil {
			continue
		}
		detail.ServiceTickets = append(detail.ServiceTickets, *cloned)
		detail.ServiceTicketSummary.Total++
		if cloned.Status != ServiceTicketStatusClosed && cloned.Status != ServiceTicketStatusResolved {
			detail.ServiceTicketSummary.OpenCount++
		}
		if cloned.SLAStatus == ServiceTicketSLAStatusEscalated {
			detail.ServiceTicketSummary.EscalatedCount++
		}
		if detail.ServiceTicketSummary.LatestUpdatedAt.IsZero() || cloned.UpdatedAt.After(detail.ServiceTicketSummary.LatestUpdatedAt) {
			detail.ServiceTicketSummary.LatestUpdatedAt = cloned.UpdatedAt
			detail.ServiceTicketSummary.LatestStatus = cloned.Status
		}
	}
	sort.SliceStable(detail.ServiceTickets, func(i, j int) bool {
		if detail.ServiceTickets[i].UpdatedAt.Equal(detail.ServiceTickets[j].UpdatedAt) {
			return detail.ServiceTickets[i].ID < detail.ServiceTickets[j].ID
		}
		return detail.ServiceTickets[i].UpdatedAt.After(detail.ServiceTickets[j].UpdatedAt)
	})
	auditTargets := map[string]struct{}{
		"after_sales:" + requestID: {},
		"order:" + order.ID:        {},
	}
	for _, ticket := range detail.ServiceTickets {
		auditTargets["service_ticket:"+ticket.ID] = struct{}{}
	}
	for _, auditLog := range s.auditLogs {
		if auditLog == nil {
			continue
		}
		if _, ok := auditTargets[auditLog.TargetType+":"+auditLog.TargetID]; !ok {
			continue
		}
		cloned := cloneAuditLog(auditLog)
		if cloned == nil {
			continue
		}
		cloned.IntegrityVerified = verifyAuditLogIntegrity(*cloned, s.auditLogSigningSecret)
		detail.RelatedAudits = append(detail.RelatedAudits, *cloned)
		detail.AuditSummary.Total++
		if cloned.IntegrityVerified {
			detail.AuditSummary.VerifiedCount++
		}
		switch cloned.TargetType {
		case "order":
			detail.AuditSummary.OrderCount++
		case "after_sales":
			detail.AuditSummary.AfterSalesCount++
		case "service_ticket":
			detail.AuditSummary.ServiceTicketCount++
		}
		if detail.AuditSummary.LatestCreatedAt.IsZero() || cloned.CreatedAt.After(detail.AuditSummary.LatestCreatedAt) {
			detail.AuditSummary.LatestCreatedAt = cloned.CreatedAt
			detail.AuditSummary.LatestAction = cloned.Action
		}
	}
	sort.SliceStable(detail.RelatedAudits, func(i, j int) bool {
		if detail.RelatedAudits[i].CreatedAt.Equal(detail.RelatedAudits[j].CreatedAt) {
			return detail.RelatedAudits[i].ID > detail.RelatedAudits[j].ID
		}
		return detail.RelatedAudits[i].CreatedAt.After(detail.RelatedAudits[j].CreatedAt)
	})
	if len(detail.RelatedAudits) > 20 {
		detail.RelatedAudits = detail.RelatedAudits[:20]
	}
	return detail, nil
}

func (s *Store) ReviewAfterSales(req ReviewAfterSalesRequest) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Decision = strings.TrimSpace(req.Decision)
	req.Reason = strings.TrimSpace(req.Reason)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.RefundIdempotencyKey = strings.TrimSpace(req.RefundIdempotencyKey)
	if req.RequestID == "" || req.Decision == "" || req.ActorID == "" || req.ActorRole == "" {
		return nil, nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reviewAfterSalesLocked(req)
}

func (s *Store) ReviewAfterSalesWithAudit(req ReviewAfterSalesRequest, audit RecordAuditLogRequest) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, *AuditLog, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Decision = strings.TrimSpace(req.Decision)
	req.Reason = strings.TrimSpace(req.Reason)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.RefundIdempotencyKey = strings.TrimSpace(req.RefundIdempotencyKey)
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if log.Action != "after_sales.reviewed" || log.TargetType != "after_sales" || log.TargetID != req.RequestID {
		return nil, nil, nil, nil, nil, ErrInvalidArgument
	}
	if req.RequestID == "" || req.Decision == "" || req.ActorID == "" || req.ActorRole == "" {
		return nil, nil, nil, nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	request, refund, order, account, err := s.reviewAfterSalesLocked(req)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	log.Payload = afterSalesReviewAuditPayload(req, request, refund)
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return request, refund, order, account, cloneAuditLog(log), nil
}

func (s *Store) reviewAfterSalesLocked(req ReviewAfterSalesRequest) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, error) {
	request := s.afterSalesRequests[req.RequestID]
	if request == nil {
		return nil, nil, nil, nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, nil, nil, nil, ErrNotFound
	}
	if !s.canReviewAfterSalesLocked(order, request, req.ActorID, req.ActorRole, req.Decision) {
		return nil, nil, nil, nil, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	switch req.Decision {
	case AfterSalesDecisionReject:
		if req.Reason == "" {
			return nil, nil, nil, nil, ErrInvalidArgument
		}
		request.Status = AfterSalesRejected
		request.ReviewReason = req.Reason
		request.ReviewerID = req.ActorID
		request.ReviewerRole = req.ActorRole
		request.ReviewedAt = now
		request.UpdatedAt = now
		s.appendAfterSalesEventLocked(request, AfterSalesActionReviewRejected, req.ActorID, req.ActorRole, req.Reason, true, nil, now)
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.after_sales.rejected",
			ActorID:   req.ActorID,
			Message:   "售后申请已驳回",
			CreatedAt: now,
		})
		return s.afterSalesRequestViewLocked(request), nil, cloneOrder(order), nil, nil
	case AfterSalesDecisionEscalate:
		request.Status = AfterSalesAdminReview
		request.ReviewReason = req.Reason
		request.ReviewerID = req.ActorID
		request.ReviewerRole = req.ActorRole
		request.ReviewedAt = now
		request.UpdatedAt = now
		message := req.Reason
		if message == "" {
			message = "售后申请已转平台审核"
		}
		s.appendAfterSalesEventLocked(request, AfterSalesActionEscalated, req.ActorID, req.ActorRole, message, true, nil, now)
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.after_sales.escalated",
			ActorID:   req.ActorID,
			Message:   "售后申请已转平台审核",
			CreatedAt: now,
		})
		return s.afterSalesRequestViewLocked(request), nil, cloneOrder(order), nil, nil
	case AfterSalesDecisionApprove:
		idempotencyKey := req.RefundIdempotencyKey
		if idempotencyKey == "" {
			idempotencyKey = "after_sales:" + request.ID
		}
		reason := req.Reason
		if reason == "" {
			reason = request.Reason
		}
		refund, refundedOrder, account, err := s.refundOrderLocked(RefundOrderRequest{
			OrderID:        request.OrderID,
			UserID:         request.UserID,
			AmountFen:      request.RequestedAmountFen,
			Destination:    req.RefundDestination,
			Reason:         reason,
			IdempotencyKey: idempotencyKey,
			ActorID:        req.ActorID,
			ActorRole:      req.ActorRole,
		})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		request.Status = AfterSalesApproved
		if refund.Status == RefundStatusSuccess {
			request.Status = AfterSalesRefunded
		}
		request.ReviewReason = reason
		request.ReviewerID = req.ActorID
		request.ReviewerRole = req.ActorRole
		request.RefundID = refund.ID
		request.ReviewedAt = now
		request.UpdatedAt = now
		s.appendAfterSalesEventLocked(request, AfterSalesActionReviewApproved, req.ActorID, req.ActorRole, reason, true, nil, now)
		s.appendOrderEventLocked(order, OrderEvent{
			Type:      "order.after_sales.approved",
			ActorID:   req.ActorID,
			Message:   "售后申请已通过",
			CreatedAt: now,
		})
		refundedOrder = cloneOrder(order)
		return s.afterSalesRequestViewLocked(request), refund, refundedOrder, account, nil
	default:
		return nil, nil, nil, nil, ErrInvalidArgument
	}
}

func afterSalesReviewAuditPayload(req ReviewAfterSalesRequest, request *AfterSalesRequest, refund *RefundTransaction) map[string]any {
	payload := map[string]any{"decision": strings.TrimSpace(req.Decision)}
	if request != nil {
		payload["status"] = strings.TrimSpace(request.Status)
		if strings.TrimSpace(request.RefundID) != "" {
			payload["refund_id"] = strings.TrimSpace(request.RefundID)
		}
	}
	if refund != nil {
		if strings.TrimSpace(refund.ID) != "" {
			payload["refund_id"] = strings.TrimSpace(refund.ID)
		}
		payload["amount_fen"] = refund.AmountFen
		payload["destination"] = strings.TrimSpace(refund.Destination)
		payload["idempotency_key"] = strings.TrimSpace(refund.IdempotencyKey)
	}
	return payload
}

func (s *Store) DepositAccount(subjectType string, subjectID string) (*DepositAccount, error) {
	subjectType = strings.TrimSpace(subjectType)
	subjectID = strings.TrimSpace(subjectID)
	if subjectType == "" || subjectID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.depositSubjectExistsLocked(subjectType, subjectID) {
		return nil, ErrNotFound
	}
	return cloneDepositAccount(s.getOrCreateDepositLocked(subjectType, subjectID)), nil
}

func (s *Store) PayDeposit(req PayDepositRequest) (*DepositAccount, error) {
	subjectType := strings.TrimSpace(req.SubjectType)
	subjectID := strings.TrimSpace(req.SubjectID)
	if subjectType == "" || subjectID == "" {
		return nil, ErrInvalidArgument
	}
	requiredAmount := requiredDepositAmount(subjectType)
	if requiredAmount <= 0 || req.AmountFen < requiredAmount {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.depositSubjectExistsLocked(subjectType, subjectID) {
		return nil, ErrNotFound
	}
	deposit := s.getOrCreateDepositLocked(subjectType, subjectID)
	deposit.AmountFen = req.AmountFen
	deposit.Status = DepositStatusPaid
	deposit.UpdatedAt = time.Now().UTC()
	s.syncDepositStatusLocked(deposit)
	return cloneDepositAccount(deposit), nil
}

func (s *Store) ApproveRiderWechatExemption(req RiderWechatExemptionRequest) (*DepositAccount, *RiderAccount, error) {
	riderID := strings.TrimSpace(req.RiderID)
	applicationID := strings.TrimSpace(req.ApplicationID)
	if riderID == "" || applicationID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	rider := s.riders[riderID]
	if rider == nil || rider.Type != RiderAccountRider {
		return nil, nil, ErrNotFound
	}
	deposit := s.getOrCreateDepositLocked("rider", riderID)
	deposit.AmountFen = 0
	deposit.Status = DepositStatusWechatExemptApproved
	deposit.WechatExemptApplicationID = applicationID
	deposit.UpdatedAt = time.Now().UTC()
	s.syncDepositStatusLocked(deposit)
	return cloneDepositAccount(deposit), cloneRiderAccount(rider), nil
}

func (s *Store) RequestRiderDepositRefund(req RiderDepositRefundRequest) (*DepositAccount, *RiderAccount, error) {
	riderID := strings.TrimSpace(req.RiderID)
	if riderID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	rider := s.riders[riderID]
	if rider == nil || rider.Type != RiderAccountRider {
		return nil, nil, ErrNotFound
	}
	deposit := s.getOrCreateDepositLocked("rider", riderID)
	if deposit.Status != DepositStatusPaid && deposit.Status != DepositStatusDisputeHold {
		return nil, nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	deposit.Status = DepositStatusRefundPending
	deposit.ResignationSubmittedAt = req.ResignationSubmittedAt
	if deposit.ResignationSubmittedAt.IsZero() {
		deposit.ResignationSubmittedAt = now
	}
	deposit.DisputeClosedAt = req.DisputeClosedAt
	deposit.LastOrderCompletedAt = s.latestRiderCompletedOrderTimeLocked(riderID)
	deposit.UpdatedAt = now
	s.syncDepositStatusLocked(deposit)
	return cloneDepositAccount(deposit), cloneRiderAccount(rider), nil
}

func (s *Store) SetRiderOnlineStatus(req SetRiderOnlineStatusRequest) (*RiderAccount, error) {
	riderID := strings.TrimSpace(req.RiderID)
	if riderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	rider := s.riders[riderID]
	if rider == nil {
		return nil, ErrNotFound
	}
	rider.Online = req.Online
	if req.Capacity > 0 {
		rider.Capacity = req.Capacity
	}
	if req.DistanceMeters >= 0 {
		rider.DistanceMeters = req.DistanceMeters
	}
	return cloneRiderAccount(rider), nil
}

func (s *Store) AutoAssignOrder(req AutoAssignOrderRequest) (*Order, *DispatchDecision, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return nil, nil, ErrInvalidArgument
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	if order.Status != StatusDispatching || order.RiderID != "" {
		return nil, nil, ErrInvalidOrderState
	}
	ageSeconds := int(now.UTC().Sub(order.CreatedAt).Seconds())
	if DispatchModeForOrderAgeSeconds(ageSeconds) != DispatchModeAutoAssign {
		return nil, nil, ErrInvalidOrderState
	}
	decision := s.dispatchDecisionLocked(order, DispatchModeAutoAssign, now.UTC())
	if decision.CandidateRiderID == "" {
		decision.Reason = "no_online_rider"
		s.recordDispatchEventLocked(order, decision, "dispatch.no_candidate", "", "system", decision.Reason, now.UTC())
		return cloneOrder(order), decision, nil
	}
	s.assignOrderToRiderLocked(order, decision.CandidateRiderID, DispatchModeAutoAssign, "system", now.UTC(), decision)
	return cloneOrder(order), decision, nil
}

func (s *Store) RejectRiderAssignment(req RejectRiderAssignmentRequest) (*Order, *DispatchDecision, error) {
	orderID := strings.TrimSpace(req.OrderID)
	riderID := strings.TrimSpace(req.RiderID)
	if orderID == "" || riderID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	if order.Status != StatusRiderAssigned || order.RiderID != riderID {
		return nil, nil, ErrInvalidOrderState
	}
	if s.dispatchRejectedRiders[orderID] == nil {
		s.dispatchRejectedRiders[orderID] = map[string]bool{}
	}
	s.dispatchRejectedRiders[orderID][riderID] = true
	order.Status = StatusDispatching
	order.RiderID = ""
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if !order.UpdatedAt.IsZero() && !now.After(order.UpdatedAt.UTC()) {
		now = order.UpdatedAt.UTC().Add(time.Nanosecond)
	}
	canDeclineWithoutPenalty, completedOrderCount, fixedOrderCount := s.riderCanDeclineWithoutPenaltyLocked(riderID, now)
	message := "骑手拒绝派单，系统顺延下一位在线骑手"
	if canDeclineWithoutPenalty {
		message = "骑手完成固定单量后免责拒绝派单，系统顺延下一位在线骑手"
	}
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "dispatch.rejected",
		ActorID:   riderID,
		Message:   message,
		CreatedAt: now,
	})
	decision := s.dispatchDecisionLocked(order, DispatchModeAutoAssign, now)
	decision.CanDeclineWithoutPenalty = canDeclineWithoutPenalty
	decision.DailyCompletedOrderCount = completedOrderCount
	decision.DailyFixedOrderCount = fixedOrderCount
	if canDeclineWithoutPenalty {
		decision.Reason = "after_fixed_order_count"
	}
	s.recordDispatchEventLocked(order, decision, "dispatch.rejected", riderID, riderID, decision.Reason, now)
	if decision.CandidateRiderID == "" {
		decision.Reason = "no_online_rider"
		s.recordDispatchEventLocked(order, decision, "dispatch.no_candidate", "", "system", decision.Reason, now)
		return cloneOrder(order), decision, nil
	}
	s.assignOrderToRiderLocked(order, decision.CandidateRiderID, DispatchModeAutoAssign, "system", now, decision)
	return cloneOrder(order), decision, nil
}

func (s *Store) TimeoutReassignOrder(req TimeoutReassignOrderRequest) (*Order, *DispatchDecision, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return nil, nil, ErrInvalidArgument
	}
	if req.TimeoutSeconds < 0 {
		return nil, nil, ErrInvalidArgument
	}
	timeoutSeconds := req.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = DispatchAssignmentTimeoutSeconds
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, allStations, err := s.stationScopeLocked(req.StationManagerID)
	if err != nil {
		return nil, nil, err
	}
	order := s.orders[orderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	if !allStations && s.orderStationIDLocked(order) != stationID {
		return nil, nil, ErrNotFound
	}
	riderID := strings.TrimSpace(req.RiderID)
	if riderID == "" {
		riderID = strings.TrimSpace(order.RiderID)
	}
	if order.Status != StatusRiderAssigned || strings.TrimSpace(order.RiderID) == "" || order.RiderID != riderID {
		return nil, nil, ErrInvalidOrderState
	}
	if !allStations {
		if rider := s.riders[riderID]; rider != nil && rider.StationID != stationID {
			return nil, nil, ErrNotFound
		}
	}
	if order.UpdatedAt.IsZero() || now.Sub(order.UpdatedAt.UTC()) < time.Duration(timeoutSeconds)*time.Second {
		return nil, nil, ErrInvalidOrderState
	}

	if s.dispatchRejectedRiders[orderID] == nil {
		s.dispatchRejectedRiders[orderID] = map[string]bool{}
	}
	s.dispatchRejectedRiders[orderID][riderID] = true
	order.Status = StatusDispatching
	order.RiderID = ""
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "dispatch.timeout",
		ActorID:   "system",
		Message:   "骑手确认超时，系统自动转派下一位在线骑手",
		CreatedAt: now,
	})

	decision := s.dispatchDecisionLocked(order, DispatchModeAutoAssign, now)
	decision.Reason = "assignment_timeout"
	s.recordDispatchEventLocked(order, decision, "dispatch.timeout", riderID, "system", decision.Reason, now)
	if decision.CandidateRiderID == "" {
		decision.Reason = "no_online_rider"
		s.recordDispatchEventLocked(order, decision, "dispatch.no_candidate", "", "system", decision.Reason, now)
		return cloneOrder(order), decision, nil
	}
	s.assignOrderToRiderLocked(order, decision.CandidateRiderID, DispatchModeAutoAssign, "system", now, decision)
	return cloneOrder(order), decision, nil
}

func (s *Store) StationRiders(stationManagerID string) ([]RiderAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, allStations, err := s.stationScopeLocked(stationManagerID)
	if err != nil {
		return nil, err
	}

	riders := make([]RiderAccount, 0)
	for _, rider := range s.riders {
		if rider == nil || rider.Type != RiderAccountRider {
			continue
		}
		if !allStations && rider.StationID != stationID {
			continue
		}
		riders = append(riders, *cloneRiderAccount(rider))
	}
	sort.SliceStable(riders, func(i, j int) bool {
		if riders[i].DispatchPriority != riders[j].DispatchPriority {
			return riders[i].DispatchPriority > riders[j].DispatchPriority
		}
		if riders[i].StationID != riders[j].StationID {
			return riders[i].StationID < riders[j].StationID
		}
		return riders[i].ID < riders[j].ID
	})
	return riders, nil
}

func (s *Store) StationOrders(stationManagerID string) ([]Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, allStations, err := s.stationScopeLocked(stationManagerID)
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0)
	for _, order := range s.orders {
		if order == nil || !isStationVisibleDispatchStatus(order.Status) {
			continue
		}
		if !allStations && s.orderStationIDLocked(order) != stationID {
			continue
		}
		if !allStations && order.RiderID != "" {
			if rider := s.riders[order.RiderID]; rider != nil && rider.StationID != stationID {
				continue
			}
		}
		orders = append(orders, *cloneOrder(order))
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders, nil
}

func (s *Store) ManualAssignOrder(req ManualAssignOrderRequest) (*Order, *DispatchDecision, error) {
	orderID := strings.TrimSpace(req.OrderID)
	riderID := strings.TrimSpace(req.RiderID)
	if orderID == "" || riderID == "" {
		return nil, nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, allStations, err := s.stationScopeLocked(req.StationManagerID)
	if err != nil {
		return nil, nil, err
	}
	rider := s.riders[riderID]
	if rider == nil || rider.Type != RiderAccountRider {
		return nil, nil, ErrNotFound
	}
	if !allStations && rider.StationID != stationID {
		return nil, nil, ErrNotFound
	}
	if !riderCanAcceptDispatchLocked(rider) {
		return nil, nil, ErrInvalidOrderState
	}
	order := s.orders[orderID]
	if order == nil {
		return nil, nil, ErrNotFound
	}
	orderStationID := s.orderStationIDLocked(order)
	if !allStations && orderStationID != stationID {
		return nil, nil, ErrNotFound
	}
	if order.Status != StatusDispatching && order.Status != StatusRiderAssigned {
		return nil, nil, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	canDeclineWithoutPenalty, completedOrderCount, fixedOrderCount := s.riderCanDeclineWithoutPenaltyLocked(riderID, now)
	decision := &DispatchDecision{
		OrderID:                      order.ID,
		Mode:                         DispatchModeManualAssign,
		StationID:                    orderStationID,
		CandidateRiderID:             riderID,
		RejectedRiderIDs:             rejectedRiderIDs(s.dispatchRejectedRiders[order.ID]),
		CanDeclineWithoutPenalty:     canDeclineWithoutPenalty,
		DailyCompletedOrderCount:     completedOrderCount,
		DailyFixedOrderCount:         fixedOrderCount,
		IdempotencyKey:               s.nextDispatchIdempotencyKeyLocked(order.ID),
		RemainingOnlineCandidateSize: s.onlineCandidateCountLocked(order),
	}
	s.assignOrderToRiderLocked(order, riderID, DispatchModeManualAssign, strings.TrimSpace(req.StationManagerID), now, decision)
	return cloneOrder(order), decision, nil
}

func (s *Store) StationTaskConfig(stationManagerID string) (*StationTaskConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, err := s.stationIDForTaskConfigLocked(stationManagerID)
	if err != nil {
		return nil, err
	}
	return cloneStationTaskConfig(s.stationTaskConfigLocked(stationID, "")), nil
}

func (s *Store) SaveStationTaskConfig(req SaveStationTaskConfigRequest) (*StationTaskConfig, error) {
	if req.DailyTaskDurationMinutes < 0 || req.DailyTaskDurationMinutes > 24*60 || req.DailyFixedOrderCount < 0 || req.DailyFixedOrderCount > 500 {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, err := s.stationIDForTaskConfigLocked(req.StationManagerID)
	if err != nil {
		return nil, err
	}
	config := s.stationTaskConfigLocked(stationID, req.StationManagerID)
	config.DailyTaskDurationMinutes = req.DailyTaskDurationMinutes
	config.DailyFixedOrderCount = req.DailyFixedOrderCount
	config.ConfiguredByStationManagerID = strings.TrimSpace(req.StationManagerID)
	return cloneStationTaskConfig(config), nil
}

func (s *Store) StationRiderPerformance(stationManagerID string) ([]RiderPerformance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.riderPerformanceSnapshotLocked(stationManagerID)
}

func (s *Store) riderPerformanceSnapshotLocked(stationManagerID string) ([]RiderPerformance, error) {
	stationID, allStations, err := s.stationScopeLocked(stationManagerID)
	if err != nil {
		return nil, err
	}

	return s.riderPerformanceSnapshotForScopeLocked(stationID, allStations), nil
}

func (s *Store) riderPerformanceSnapshotForScopeLocked(stationID string, allStations bool) []RiderPerformance {
	reviewSummaries := s.riderReviewSummariesLocked()
	performances := make([]RiderPerformance, 0)
	now := time.Now().UTC()
	for _, rider := range s.riders {
		if rider == nil || rider.Type != RiderAccountRider {
			continue
		}
		if !allStations && rider.StationID != stationID {
			continue
		}
		acceptedCount, completedCount := s.riderOrderCountsLocked(rider.ID)
		averageDailyOrders := rider.AverageDailyOrders
		if completedCount > 0 {
			averageDailyOrders += float64(completedCount)
		}
		completionRate := rider.CompletionRate
		if acceptedCount > 0 && completedCount > 0 {
			completionRate = float64(completedCount) / float64(acceptedCount)
		}
		reviewSummary := reviewSummaries[rider.ID]
		riderAverageRating := 0.0
		if reviewSummary.Count > 0 {
			riderAverageRating = roundFloat(float64(reviewSummary.TotalRating)/float64(reviewSummary.Count), 2)
		}
		performances = append(performances, RiderPerformance{
			RiderID:              rider.ID,
			StationID:            rider.StationID,
			AverageAcceptSeconds: rider.AverageAcceptSeconds,
			AverageDailyOrders:   averageDailyOrders,
			CompletionRate:       completionRate,
			RiderAverageRating:   riderAverageRating,
			RiderReviewCount:     reviewSummary.Count,
		})
	}

	teamAverageAcceptSeconds := averagePositiveAcceptSeconds(performances)
	teamAverageDailyOrders := averagePositiveDailyOrders(performances)
	for index := range performances {
		performance := &performances[index]
		performance.Score, performance.Level, performance.DispatchPriority, performance.ScoreBreakdown = evaluateRiderPerformanceLevel(*performance, teamAverageAcceptSeconds, teamAverageDailyOrders)
		if rider := s.riders[performance.RiderID]; rider != nil {
			performance.RecentTrend = s.riderPerformanceTrendLocked(rider, *performance, teamAverageAcceptSeconds, teamAverageDailyOrders, now)
		}
		performance.RecentReviews = s.riderRecentReviewExcerptsLocked(performance.RiderID, 2)
		performance.ExceptionDetails, performance.ExceptionSummary = s.riderExceptionDetailsLocked(performance.RiderID, now.AddDate(0, 0, -7), 3)
		if rider := s.riders[performance.RiderID]; rider != nil {
			rider.DispatchPriority = performance.DispatchPriority
		}
		performance.AverageDailyOrders = roundFloat(performance.AverageDailyOrders, 2)
		performance.CompletionRate = roundFloat(performance.CompletionRate, 4)
	}
	sort.SliceStable(performances, func(i, j int) bool {
		if performances[i].DispatchPriority != performances[j].DispatchPriority {
			return performances[i].DispatchPriority > performances[j].DispatchPriority
		}
		if performances[i].Score != performances[j].Score {
			return performances[i].Score > performances[j].Score
		}
		if performances[i].AverageAcceptSeconds != performances[j].AverageAcceptSeconds {
			return performances[i].AverageAcceptSeconds < performances[j].AverageAcceptSeconds
		}
		return performances[i].RiderID < performances[j].RiderID
	})
	return performances
}

func (s *Store) refreshDispatchPrioritiesForStationLocked(stationID string) {
	stationID = strings.TrimSpace(stationID)
	if stationID == "" {
		return
	}
	_ = s.riderPerformanceSnapshotForScopeLocked(stationID, false)
}

type riderReviewSummary struct {
	TotalRating int
	Count       int
}

func (s *Store) riderReviewSummariesLocked() map[string]riderReviewSummary {
	summaries := map[string]riderReviewSummary{}
	for _, review := range s.reviews {
		riderID, rating := s.riderReviewAttributionLocked(review)
		if riderID == "" || rating < 1 || rating > 5 {
			continue
		}
		summary := summaries[riderID]
		summary.TotalRating += rating
		summary.Count++
		summaries[riderID] = summary
	}
	return summaries
}

func (s *Store) riderReviewAttributionLocked(review *Review) (string, int) {
	if review == nil {
		return "", 0
	}
	if strings.TrimSpace(review.TargetType) == ReviewTargetRider && strings.TrimSpace(review.TargetID) != "" {
		rating := review.Rating
		if review.RiderRating > 0 {
			rating = review.RiderRating
		}
		return strings.TrimSpace(review.TargetID), rating
	}
	if review.RiderRating <= 0 || strings.TrimSpace(review.OrderID) == "" {
		return "", 0
	}
	order := s.orders[strings.TrimSpace(review.OrderID)]
	if order == nil || strings.TrimSpace(order.RiderID) == "" {
		return "", 0
	}
	return strings.TrimSpace(order.RiderID), review.RiderRating
}

type riderTrendAggregate struct {
	Date            string
	CompletedOrders int
	TotalRating     int
	ReviewCount     int
	TimeoutCount    int
	RejectCount     int
}

func (s *Store) riderPerformanceTrendLocked(rider *RiderAccount, performance RiderPerformance, teamAverageAcceptSeconds float64, teamAverageDailyOrders float64, now time.Time) []RiderPerformanceTrendPoint {
	if rider == nil {
		return nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	aggregates := map[string]*riderTrendAggregate{}
	dates := make([]string, 0, 3)
	for dayOffset := 2; dayOffset >= 0; dayOffset-- {
		date := now.AddDate(0, 0, -dayOffset).Format("2006-01-02")
		aggregates[date] = &riderTrendAggregate{Date: date}
		dates = append(dates, date)
	}
	for _, order := range s.orders {
		if order == nil || strings.TrimSpace(order.RiderID) != rider.ID || order.Status != StatusCompleted {
			continue
		}
		completedAt := order.UpdatedAt
		if completedAt.IsZero() {
			completedAt = order.CreatedAt
		}
		date := completedAt.UTC().Format("2006-01-02")
		aggregate := aggregates[date]
		if aggregate == nil {
			continue
		}
		aggregate.CompletedOrders++
	}
	for _, review := range s.reviews {
		riderID, rating := s.riderReviewAttributionLocked(review)
		if riderID != rider.ID || rating < 1 || rating > 5 {
			continue
		}
		date := review.CreatedAt.UTC().Format("2006-01-02")
		aggregate := aggregates[date]
		if aggregate == nil {
			continue
		}
		aggregate.TotalRating += rating
		aggregate.ReviewCount++
	}
	for _, event := range s.dispatchEvents {
		if event == nil || strings.TrimSpace(event.RiderID) != rider.ID {
			continue
		}
		date := event.CreatedAt.UTC().Format("2006-01-02")
		aggregate := aggregates[date]
		if aggregate == nil {
			continue
		}
		switch event.Type {
		case "dispatch.timeout":
			aggregate.TimeoutCount++
		case "dispatch.rejected":
			aggregate.RejectCount++
		}
	}
	points := make([]RiderPerformanceTrendPoint, 0, len(dates))
	for _, date := range dates {
		aggregate := aggregates[date]
		if aggregate == nil {
			continue
		}
		averageRating := 0.0
		if aggregate.ReviewCount > 0 {
			averageRating = roundFloat(float64(aggregate.TotalRating)/float64(aggregate.ReviewCount), 2)
		}
		syntheticPerformance := RiderPerformance{
			RiderID:              rider.ID,
			StationID:            rider.StationID,
			AverageAcceptSeconds: performance.AverageAcceptSeconds,
			AverageDailyOrders:   float64(aggregate.CompletedOrders),
			CompletionRate:       performance.CompletionRate,
			RiderAverageRating:   averageRating,
			RiderReviewCount:     aggregate.ReviewCount,
		}
		score, _, _, _ := evaluateRiderPerformanceLevel(syntheticPerformance, teamAverageAcceptSeconds, teamAverageDailyOrders)
		points = append(points, RiderPerformanceTrendPoint{
			Date:            aggregate.Date,
			Score:           score,
			CompletedOrders: aggregate.CompletedOrders,
			AverageRating:   averageRating,
			TimeoutCount:    aggregate.TimeoutCount,
			RejectCount:     aggregate.RejectCount,
		})
	}
	return points
}

func (s *Store) riderRecentReviewExcerptsLocked(riderID string, limit int) []RiderPerformanceReviewExcerpt {
	riderID = strings.TrimSpace(riderID)
	if riderID == "" || limit <= 0 {
		return nil
	}
	excerpts := make([]RiderPerformanceReviewExcerpt, 0, limit)
	for _, review := range s.reviews {
		attributedRiderID, _ := s.riderReviewAttributionLocked(review)
		if attributedRiderID != riderID {
			continue
		}
		excerpts = append(excerpts, RiderPerformanceReviewExcerpt{
			ReviewID:    review.ID,
			OrderID:     review.OrderID,
			Rating:      review.Rating,
			RiderRating: review.RiderRating,
			Content:     review.Content,
			Tags:        append([]string{}, review.Tags...),
			CreatedAt:   review.CreatedAt,
		})
	}
	sort.SliceStable(excerpts, func(i, j int) bool {
		if !excerpts[i].CreatedAt.Equal(excerpts[j].CreatedAt) {
			return excerpts[i].CreatedAt.After(excerpts[j].CreatedAt)
		}
		return excerpts[i].ReviewID < excerpts[j].ReviewID
	})
	if len(excerpts) > limit {
		excerpts = excerpts[:limit]
	}
	return excerpts
}

func (s *Store) riderExceptionDetailsLocked(riderID string, since time.Time, limit int) ([]RiderPerformanceExceptionDetail, RiderPerformanceExceptionSummary) {
	riderID = strings.TrimSpace(riderID)
	if riderID == "" {
		return nil, RiderPerformanceExceptionSummary{}
	}
	summary := RiderPerformanceExceptionSummary{}
	details := make([]RiderPerformanceExceptionDetail, 0)
	for _, event := range s.dispatchEvents {
		if event == nil || strings.TrimSpace(event.RiderID) != riderID {
			continue
		}
		if !since.IsZero() && event.CreatedAt.Before(since) {
			continue
		}
		switch event.Type {
		case "dispatch.timeout":
			summary.DispatchTimeoutCount++
			details = append(details, RiderPerformanceExceptionDetail{
				Kind:            "dispatch_timeout",
				Label:           "派单超时",
				OrderID:         event.OrderID,
				DispatchEventID: event.ID,
				Status:          event.Type,
				Message:         riderExceptionDispatchMessage(event, "派单确认超时，系统已自动转派"),
				CreatedAt:       event.CreatedAt,
			})
		case "dispatch.rejected":
			summary.DispatchRejectCount++
			details = append(details, RiderPerformanceExceptionDetail{
				Kind:            "dispatch_reject",
				Label:           "骑手拒单",
				OrderID:         event.OrderID,
				DispatchEventID: event.ID,
				Status:          event.Type,
				Message:         riderExceptionDispatchMessage(event, "骑手已拒绝当前派单，系统顺延下一位"),
				CreatedAt:       event.CreatedAt,
			})
		default:
			continue
		}
		if event.CreatedAt.After(summary.LastEventAt) {
			summary.LastEventAt = event.CreatedAt
		}
	}
	for _, request := range s.afterSalesRequests {
		if request == nil || !request.CreatedAt.After(since) && !since.IsZero() {
			continue
		}
		order := s.orders[request.OrderID]
		if order == nil || strings.TrimSpace(order.RiderID) != riderID {
			continue
		}
		summary.AfterSalesCount++
		details = append(details, RiderPerformanceExceptionDetail{
			Kind:                "after_sales",
			Label:               "售后介入",
			OrderID:             request.OrderID,
			AfterSalesRequestID: request.ID,
			Status:              request.Status,
			Message:             riderExceptionAfterSalesMessage(s.afterSalesRequestViewLocked(request)),
			CreatedAt:           riderExceptionAfterSalesAt(request),
		})
		detailAt := riderExceptionAfterSalesAt(request)
		if detailAt.After(summary.LastEventAt) {
			summary.LastEventAt = detailAt
		}
	}
	for _, review := range s.reviews {
		attributedRiderID, rating := s.riderReviewAttributionLocked(review)
		if attributedRiderID != riderID || rating > 3 || rating <= 0 {
			continue
		}
		if !since.IsZero() && review.CreatedAt.Before(since) {
			continue
		}
		summary.LowRatingCount++
		details = append(details, RiderPerformanceExceptionDetail{
			Kind:      "low_rating",
			Label:     "低分评价",
			OrderID:   review.OrderID,
			ReviewID:  review.ID,
			Status:    fmt.Sprintf("%d 星", rating),
			Message:   riderExceptionLowRatingMessage(review, rating),
			CreatedAt: review.CreatedAt,
		})
		if review.CreatedAt.After(summary.LastEventAt) {
			summary.LastEventAt = review.CreatedAt
		}
	}
	sort.SliceStable(details, func(i, j int) bool {
		if !details[i].CreatedAt.Equal(details[j].CreatedAt) {
			return details[i].CreatedAt.After(details[j].CreatedAt)
		}
		if details[i].OrderID != details[j].OrderID {
			return details[i].OrderID < details[j].OrderID
		}
		return details[i].Label < details[j].Label
	})
	if limit > 0 && len(details) > limit {
		details = details[:limit]
	}
	return details, summary
}

func riderExceptionDispatchMessage(event *DispatchEvent, fallback string) string {
	if event == nil {
		return fallback
	}
	reason := strings.TrimSpace(event.Reason)
	if reason == "" {
		return fallback
	}
	return reason
}

func riderExceptionAfterSalesAt(request *AfterSalesRequest) time.Time {
	if request == nil {
		return time.Time{}
	}
	if !request.UpdatedAt.IsZero() {
		return request.UpdatedAt
	}
	return request.CreatedAt
}

func riderExceptionAfterSalesMessage(request *AfterSalesRequest) string {
	if request == nil {
		return "售后单已进入处理流程"
	}
	message := strings.TrimSpace(request.LatestEventMessage)
	if message == "" {
		message = strings.TrimSpace(request.Reason)
	}
	if message == "" {
		message = "售后单已进入处理流程"
	}
	return message
}

func riderExceptionLowRatingMessage(review *Review, rating int) string {
	if review == nil {
		return fmt.Sprintf("配送评分 %d 星", rating)
	}
	content := strings.TrimSpace(review.Content)
	if content == "" {
		return fmt.Sprintf("配送评分 %d 星", rating)
	}
	return fmt.Sprintf("配送评分 %d 星：%s", rating, content)
}

func (s *Store) GrabOrder(orderID string, riderID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	riderID = strings.TrimSpace(riderID)
	if orderID == "" || riderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if order.Status != StatusDispatching || order.RiderID != "" {
		return nil, ErrOrderAlreadyAssigned
	}
	rider := s.riders[riderID]
	if !riderCanAcceptDispatchLocked(rider) || !riderMatchesOrderStationLocked(rider, s.orderStationIDLocked(order)) {
		return nil, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	order.Status = StatusRiderAssigned
	order.RiderID = riderID
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "dispatch.grabbed",
		ActorID:   riderID,
		Message:   "骑手抢单成功",
		CreatedAt: now,
	})
	decision := &DispatchDecision{
		OrderID:                      order.ID,
		Mode:                         DispatchModeGrabHall,
		StationID:                    s.orderStationIDLocked(order),
		CandidateRiderID:             riderID,
		RejectedRiderIDs:             rejectedRiderIDs(s.dispatchRejectedRiders[order.ID]),
		IdempotencyKey:               s.nextDispatchIdempotencyKeyLocked(order.ID),
		RemainingOnlineCandidateSize: s.onlineCandidateCountLocked(order),
	}
	s.recordDispatchEventLocked(order, decision, "dispatch.grabbed", riderID, riderID, "", now)
	return cloneOrder(order), nil
}

func (s *Store) DispatchEvents(orderID string, stationManagerID string) ([]DispatchEvent, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	stationID, allStations, err := s.stationScopeLocked(stationManagerID)
	if err != nil {
		return nil, err
	}
	if !allStations && s.orderStationIDLocked(order) != stationID {
		return nil, ErrNotFound
	}
	events := make([]DispatchEvent, 0)
	for _, event := range s.dispatchEvents {
		if event == nil || event.OrderID != orderID {
			continue
		}
		events = append(events, *cloneDispatchEvent(event))
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events, nil
}

func (s *Store) RecordAuditLog(req RecordAuditLogRequest) (*AuditLog, error) {
	actorType := strings.TrimSpace(req.ActorType)
	actorID := strings.TrimSpace(req.ActorID)
	action := strings.TrimSpace(req.Action)
	targetType := strings.TrimSpace(req.TargetType)
	targetID := strings.TrimSpace(req.TargetID)
	if actorType == "" || actorID == "" || action == "" || targetType == "" || targetID == "" {
		return nil, ErrInvalidArgument
	}
	now := req.CreatedAt
	if now.IsZero() {
		now = normalizeAuditLogTime(time.Now())
	}
	now = normalizeAuditLogTime(now)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextAuditLogID++
	log := &AuditLog{
		ID:         fmt.Sprintf("aud_%d", s.nextAuditLogID),
		ActorType:  actorType,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		RequestID:  strings.TrimSpace(req.RequestID),
		IPHash:     strings.TrimSpace(req.IPHash),
		Payload:    sanitizeAuditPayload(req.Payload),
		CreatedAt:  now,
	}
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return cloneAuditLog(log), nil
}

func (s *Store) appendAuditLogLocked(log *AuditLog) *AuditLog {
	s.nextAuditLogID++
	log.ID = fmt.Sprintf("aud_%d", s.nextAuditLogID)
	sealAuditLogIntegrity(log, s.auditLogSigningSecret)
	s.auditLogs[log.ID] = log
	return cloneAuditLog(log)
}

func (s *Store) AuditLogs(req AuditLogsRequest) ([]AuditLog, error) {
	req = normalizeAuditLogsRequest(req)

	s.mu.Lock()
	defer s.mu.Unlock()

	logs := make([]AuditLog, 0, len(s.auditLogs))
	for _, log := range s.auditLogs {
		if !auditLogMatchesRequest(log, req) {
			continue
		}
		output := cloneAuditLog(log)
		output.IntegrityVerified = verifyAuditLogIntegrity(*output, s.auditLogSigningSecret)
		logs = append(logs, *output)
	}
	sort.SliceStable(logs, func(i, j int) bool {
		if !logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].CreatedAt.After(logs[j].CreatedAt)
		}
		return logs[i].ID > logs[j].ID
	})
	if len(logs) > req.Limit {
		logs = logs[:req.Limit]
	}
	return logs, nil
}

func (s *Store) AuditRetentionReport(req AuditRetentionReportRequest) (*AuditRetentionReport, error) {
	req = normalizeAuditRetentionReportRequest(req)

	s.mu.Lock()
	logs := make([]AuditLog, 0, len(s.auditLogs))
	for _, log := range s.auditLogs {
		output := cloneAuditLog(log)
		output.IntegrityVerified = verifyAuditLogIntegrity(*output, s.auditLogSigningSecret)
		logs = append(logs, *output)
	}
	s.mu.Unlock()

	return auditRetentionReportFromLogs(req, logs), nil
}

func (s *Store) EmitAuditRetentionAlerts(req AuditRetentionAlertEmissionRequest, audit RecordAuditLogRequest) (*AuditRetentionAlertEmission, *OutboxEvent, *AuditLog, error) {
	reportReq := normalizeAuditRetentionReportRequest(AuditRetentionReportRequest{
		RetentionDays:        req.RetentionDays,
		HotDays:              req.HotDays,
		IntegritySampleLimit: req.IntegritySampleLimit,
		Now:                  req.Now,
	})
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = reportReq.Now
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, log, err
	}
	if log.Action != "admin.audit_retention_alerts.emitted" || log.TargetType != "audit_retention_alerts" || log.TargetID != "default" {
		return nil, nil, log, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	logs := make([]AuditLog, 0, len(s.auditLogs))
	for _, auditLog := range s.auditLogs {
		output := cloneAuditLog(auditLog)
		output.IntegrityVerified = verifyAuditLogIntegrity(*output, s.auditLogSigningSecret)
		logs = append(logs, *output)
	}
	report := auditRetentionReportFromLogs(reportReq, logs)
	emission := auditRetentionAlertEmissionFromReport(report)
	var event *OutboxEvent
	if len(report.Alerts) > 0 {
		event = s.enqueueOutboxEventLocked(
			auditRetentionAlertTopic,
			"audit_retention",
			"default",
			"audit.retention_alerts.emitted",
			emission.IdempotencyKey,
			auditRetentionAlertOutboxPayload(emission),
			emission.EmittedAt,
		)
		if event != nil {
			emission.OutboxEventID = event.ID
		}
	}
	log.Payload = auditRetentionAlertEmissionAuditPayload(emission)
	auditLog := s.appendAuditLogLocked(log)
	eventCopy := cloneOutboxEvent(event)

	return emission, eventCopy, auditLog, nil
}

func (s *Store) RequestAuditArchive(req AuditArchiveRequest, audit RecordAuditLogRequest) (*AuditArchiveRequestResult, *OutboxEvent, *AuditLog, error) {
	req = normalizeAuditArchiveRequest(req)
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = req.Now
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, log, err
	}
	if log.Action != "admin.audit_archive.requested" || log.TargetType != "audit_archive" {
		return nil, nil, log, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	logs := make([]AuditLog, 0, len(s.auditLogs))
	for _, auditLog := range s.auditLogs {
		output := cloneAuditLog(auditLog)
		output.IntegrityVerified = verifyAuditLogIntegrity(*output, s.auditLogSigningSecret)
		logs = append(logs, *output)
	}
	result := auditArchiveRequestFromLogs(req, logs)
	var event *OutboxEvent
	if result.LogCount > 0 {
		event = s.enqueueOutboxEventLocked(
			auditArchiveRequestedTopic,
			"audit_archive",
			result.ArchiveID,
			"audit.archive_requested",
			result.IdempotencyKey,
			auditArchiveOutboxPayload(result),
			result.RequestedAt,
		)
		if event != nil {
			result.OutboxEventID = event.ID
		}
	}
	log.TargetID = result.ArchiveID
	log.Payload = auditArchiveRequestAuditPayload(result)
	auditLog := s.appendAuditLogLocked(log)
	eventCopy := cloneOutboxEvent(event)

	return result, eventCopy, auditLog, nil
}

func (s *Store) CompleteAuditArchive(req AuditArchiveCompletionRequest, audit RecordAuditLogRequest) (*AuditArchiveCompletion, *AuditLog, error) {
	normalized, err := normalizeAuditArchiveCompletionRequest(req)
	if err != nil {
		return nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = normalized.UploadedAt
	}
	if strings.TrimSpace(audit.TargetID) == "" || strings.TrimSpace(audit.TargetID) == "pending" {
		audit.TargetID = normalized.ArchiveID
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != auditArchiveCompletedAction || log.TargetType != "audit_archive" || log.TargetID != normalized.ArchiveID {
		return nil, log, ErrInvalidArgument
	}
	completion := auditArchiveCompletionFromRequest(normalized, log.CreatedAt)
	log.Payload = auditArchiveCompletionAuditPayload(completion)

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, existingLog, ok := s.auditArchiveCompletionLocked(normalized); ok {
		return &existing, existingLog, nil
	}
	auditLog := s.appendAuditLogLocked(log)
	return completion, auditLog, nil
}

func (s *Store) AuditArchives(req AuditArchiveListRequest) ([]AuditArchiveCompletion, error) {
	req = normalizeAuditArchiveListRequest(req)
	logs, err := s.AuditLogs(AuditLogsRequest{
		Action:     auditArchiveCompletedAction,
		TargetType: "audit_archive",
		TargetID:   req.ArchiveID,
		Limit:      req.Limit,
		After:      req.After,
		Before:     req.Before,
	})
	if err != nil {
		return nil, err
	}
	return auditArchiveCompletionsFromLogs(logs), nil
}

func (s *Store) VerifyAuditArchive(req AuditArchiveVerifyRequest, audit RecordAuditLogRequest) (*AuditArchiveVerification, *AuditLog, error) {
	req = normalizeAuditArchiveVerifyRequest(req)
	if req.ArchiveID == "" {
		return nil, nil, ErrInvalidArgument
	}
	archives, err := s.AuditArchives(AuditArchiveListRequest{ArchiveID: req.ArchiveID, Limit: 1})
	if err != nil {
		return nil, nil, err
	}
	if len(archives) == 0 {
		return nil, nil, ErrNotFound
	}
	verification := verifyAuditArchiveCompletion(archives[0], s.objectStorageSnapshot(), req.Now)
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = verification.VerifiedAt
	}
	if strings.TrimSpace(audit.TargetID) == "" || strings.TrimSpace(audit.TargetID) == "pending" {
		audit.TargetID = verification.ArchiveID
	}
	if strings.TrimSpace(audit.Action) != auditArchiveVerifiedAction || strings.TrimSpace(audit.TargetType) != "audit_archive" || strings.TrimSpace(audit.TargetID) != verification.ArchiveID {
		return verification, nil, ErrInvalidArgument
	}
	audit.Payload = auditArchiveVerificationAuditPayload(verification)
	log, err := s.RecordAuditLog(audit)
	if err != nil {
		return verification, log, err
	}
	return verification, log, nil
}

func (s *Store) AuditArchiveVerifications(req AuditArchiveVerificationListRequest) ([]AuditArchiveVerification, error) {
	req = normalizeAuditArchiveVerificationListRequest(req)
	logs, err := s.AuditLogs(AuditLogsRequest{
		Action:     auditArchiveVerifiedAction,
		TargetType: "audit_archive",
		TargetID:   req.ArchiveID,
		Limit:      req.Limit,
		After:      req.After,
		Before:     req.Before,
	})
	if err != nil {
		return nil, err
	}
	return auditArchiveVerificationsFromLogs(req, logs), nil
}

func (s *Store) CreateNotification(req CreateNotificationRequest) (*PlatformNotification, error) {
	normalized, err := normalizeCreateNotificationRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if normalized.TargetRole == "merchant" && s.merchants[normalized.TargetID] == nil {
		return nil, ErrNotFound
	}
	if existingID := s.notificationByIdem[normalized.IdempotencyKey]; existingID != "" {
		return clonePlatformNotification(s.notifications[existingID]), nil
	}
	now := normalized.CreatedAt
	notification := &PlatformNotification{
		ID:             "ntf_" + shortHash(normalized.IdempotencyKey),
		TargetRole:     normalized.TargetRole,
		TargetID:       normalized.TargetID,
		Type:           normalized.Type,
		Channel:        normalized.Channel,
		Title:          normalized.Title,
		Body:           normalized.Body,
		Status:         NotificationStatusUnread,
		SourceTopic:    normalized.SourceTopic,
		SourceEventID:  normalized.SourceEventID,
		IdempotencyKey: normalized.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.notifications[notification.ID] = notification
	s.notificationByIdem[notification.IdempotencyKey] = notification.ID
	return clonePlatformNotification(notification), nil
}

func (s *Store) Notifications(req NotificationListRequest) ([]PlatformNotification, error) {
	req = normalizeNotificationListRequest(req)
	if (req.TargetRole == "") != (req.TargetID == "") {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	notifications := make([]PlatformNotification, 0)
	for _, notification := range s.notifications {
		if notification == nil {
			continue
		}
		if req.TargetRole != "" && notification.TargetRole != req.TargetRole {
			continue
		}
		if req.TargetID != "" && notification.TargetID != req.TargetID {
			continue
		}
		if req.Status != "all" && notification.Status != req.Status {
			continue
		}
		if req.SourceTopic != "" && notification.SourceTopic != req.SourceTopic {
			continue
		}
		if req.SourceEventID != "" && notification.SourceEventID != req.SourceEventID {
			continue
		}
		notifications = append(notifications, *clonePlatformNotification(notification))
	}
	sort.SliceStable(notifications, func(i, j int) bool {
		if !notifications[i].CreatedAt.Equal(notifications[j].CreatedAt) {
			return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
		}
		return notifications[i].ID > notifications[j].ID
	})
	if len(notifications) > req.Limit {
		notifications = notifications[:req.Limit]
	}
	return notifications, nil
}

func (s *Store) MarkNotificationRead(req MarkNotificationReadRequest) (*PlatformNotification, error) {
	req.NotificationID = strings.TrimSpace(req.NotificationID)
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	if req.NotificationID == "" || req.TargetRole == "" || req.TargetID == "" {
		return nil, ErrInvalidArgument
	}
	readAt := req.ReadAt
	if readAt.IsZero() {
		readAt = time.Now().UTC()
	} else {
		readAt = readAt.UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	notification := s.notifications[req.NotificationID]
	if notification == nil {
		return nil, ErrNotFound
	}
	if notification.TargetRole != req.TargetRole || notification.TargetID != req.TargetID {
		return nil, ErrNotFound
	}
	if notification.Status != NotificationStatusRead {
		notification.Status = NotificationStatusRead
		notification.ReadAt = readAt
	}
	notification.UpdatedAt = readAt
	return clonePlatformNotification(notification), nil
}

func (s *Store) RecordNotificationDelivery(req RecordNotificationDeliveryRequest) (*PlatformNotificationDelivery, error) {
	normalized, err := normalizeRecordNotificationDeliveryRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existingID := s.notificationDeliveryByIdem[normalized.IdempotencyKey]; existingID != "" {
		return clonePlatformNotificationDelivery(s.notificationDeliveries[existingID]), nil
	}
	notification := s.notifications[normalized.NotificationID]
	if notification == nil {
		return nil, ErrNotFound
	}
	if normalized.Channel == "" {
		normalized.Channel = notification.Channel
	}
	if normalized.Provider == "" {
		normalized.Provider = normalized.Channel
	}
	if normalized.Status == NotificationDeliveryDelivered && normalized.DeliveredAt.IsZero() {
		normalized.DeliveredAt = normalized.AttemptedAt
	}
	now := normalized.AttemptedAt
	delivery := &PlatformNotificationDelivery{
		ID:                "ntfd_" + shortHash(normalized.IdempotencyKey),
		NotificationID:    notification.ID,
		TargetRole:        notification.TargetRole,
		TargetID:          notification.TargetID,
		Channel:           normalized.Channel,
		Provider:          normalized.Provider,
		Status:            normalized.Status,
		ProviderMessageID: normalized.ProviderMessageID,
		ErrorCode:         normalized.ErrorCode,
		ErrorMessage:      normalized.ErrorMessage,
		IdempotencyKey:    normalized.IdempotencyKey,
		AttemptedAt:       normalized.AttemptedAt,
		DeliveredAt:       normalized.DeliveredAt,
		RetryAt:           normalized.RetryAt,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	s.notificationDeliveries[delivery.ID] = delivery
	s.notificationDeliveryByIdem[delivery.IdempotencyKey] = delivery.ID
	return clonePlatformNotificationDelivery(delivery), nil
}

func (s *Store) NotificationDeliveries(req NotificationDeliveryListRequest) ([]PlatformNotificationDelivery, error) {
	req = normalizeNotificationDeliveryListRequest(req)
	if (req.TargetRole == "") != (req.TargetID == "") {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	deliveries := make([]PlatformNotificationDelivery, 0)
	for _, delivery := range s.notificationDeliveries {
		if delivery == nil {
			continue
		}
		if req.NotificationID != "" && delivery.NotificationID != req.NotificationID {
			continue
		}
		if req.TargetRole != "" && delivery.TargetRole != req.TargetRole {
			continue
		}
		if req.TargetID != "" && delivery.TargetID != req.TargetID {
			continue
		}
		if req.Channel != "" && delivery.Channel != req.Channel {
			continue
		}
		if req.Provider != "" && delivery.Provider != req.Provider {
			continue
		}
		if req.Status != "all" && delivery.Status != req.Status {
			continue
		}
		if req.ErrorCode != "" && delivery.ErrorCode != req.ErrorCode {
			continue
		}
		if !req.RetryAtBefore.IsZero() && (delivery.RetryAt.IsZero() || delivery.RetryAt.After(req.RetryAtBefore)) {
			continue
		}
		deliveries = append(deliveries, *clonePlatformNotificationDelivery(delivery))
	}
	sort.SliceStable(deliveries, func(i, j int) bool {
		if !deliveries[i].AttemptedAt.Equal(deliveries[j].AttemptedAt) {
			return deliveries[i].AttemptedAt.After(deliveries[j].AttemptedAt)
		}
		return deliveries[i].ID > deliveries[j].ID
	})
	if len(deliveries) > req.Limit {
		deliveries = deliveries[:req.Limit]
	}
	return deliveries, nil
}

func (s *Store) EmitNotificationFailureAlerts(req NotificationFailureAlertEmissionRequest, audit RecordAuditLogRequest) (*NotificationFailureAlertEmission, *OutboxEvent, *AuditLog, error) {
	req, err := normalizeNotificationFailureAlertEmissionRequest(req)
	if err != nil {
		return nil, nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = req.Now
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, log, err
	}
	if log.Action != "admin.notification_delivery_failure_alerts.emitted" || log.TargetType != "notification_delivery_alerts" || log.TargetID != "failed" {
		return nil, nil, log, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	deliveries := notificationFailureAlertDeliveriesFromMap(req, s.notificationDeliveries)
	emission := notificationFailureAlertEmissionFromDeliveries(req, deliveries)
	var event *OutboxEvent
	if emission.FailedCount > 0 {
		event = s.enqueueOutboxEventLocked(
			notificationFailureAlertTopic,
			"notification_delivery",
			"failed",
			"notification.delivery_failed_alerts.emitted",
			emission.IdempotencyKey,
			notificationFailureAlertOutboxPayload(emission),
			emission.EmittedAt,
		)
		if event != nil {
			emission.OutboxEventID = event.ID
		}
	}
	log.Payload = notificationFailureAlertEmissionAuditPayload(emission)
	auditLog := s.appendAuditLogLocked(log)
	return emission, cloneOutboxEvent(event), auditLog, nil
}

func (s *Store) ScheduleNotificationDeliveryRetries(req NotificationDeliveryRetryScheduleRequest, audit RecordAuditLogRequest) (*NotificationDeliveryRetrySchedule, *OutboxEvent, *AuditLog, error) {
	req, err := normalizeNotificationDeliveryRetryScheduleRequest(req)
	if err != nil {
		return nil, nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = req.Now
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, log, err
	}
	if log.Action != "admin.notification_delivery_retries.scheduled" || log.TargetType != "notification_delivery_retries" || log.TargetID != req.Status {
		return nil, nil, log, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	deliveries := notificationDeliveryRetryDeliveriesFromMap(req, s.notificationDeliveries)
	notifications := notificationDeliveryRetryNotificationsFromMap(deliveries, s.notifications)
	schedule := notificationDeliveryRetryScheduleFromDeliveries(req, deliveries, notifications)
	var event *OutboxEvent
	if schedule.ScheduledCount > 0 {
		event = s.enqueueOutboxEventLocked(
			notificationDeliveryRetryTopic,
			"notification_delivery",
			schedule.DeliveryStatus,
			"notification.delivery_retries.scheduled",
			schedule.IdempotencyKey,
			notificationDeliveryRetryOutboxPayload(schedule),
			schedule.RetryAt,
		)
		if event != nil {
			schedule.OutboxEventID = event.ID
		}
	}
	log.Payload = notificationDeliveryRetryScheduleAuditPayload(schedule)
	auditLog := s.appendAuditLogLocked(log)
	return schedule, cloneOutboxEvent(event), auditLog, nil
}

func (s *Store) ScheduleNotificationQuietWindowRetries(req NotificationQuietWindowRetryScheduleRequest, audit RecordAuditLogRequest) (*NotificationDeliveryRetrySchedule, *OutboxEvent, *AuditLog, error) {
	req, err := normalizeNotificationQuietWindowRetryScheduleRequest(req)
	if err != nil {
		return nil, nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = req.Now
	}
	return s.ScheduleNotificationDeliveryRetries(NotificationDeliveryRetryScheduleRequest{
		TargetRole:          req.TargetRole,
		TargetID:            req.TargetID,
		Channel:             req.Channel,
		Provider:            req.Provider,
		Status:              NotificationDeliveryQueued,
		ErrorCode:           "notification_quiet_window",
		Limit:               req.Limit,
		RetryAfterSeconds:   req.RetryAfterSeconds,
		RetryAt:             req.Now.Add(time.Duration(req.RetryAfterSeconds) * time.Second),
		SourceRetryAtBefore: req.Now,
		Now:                 req.Now,
	}, audit)
}

func (s *Store) SaveNotificationPreference(req SaveNotificationPreferenceRequest) (*NotificationPreference, error) {
	normalized, err := normalizeSaveNotificationPreferenceRequest(req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	preference, err := s.saveNotificationPreferenceLocked(normalized)
	if err != nil {
		return nil, err
	}
	s.enqueueNotificationPreferenceChangedOutboxLocked(preference)
	return preference, nil
}

func (s *Store) SaveNotificationPreferenceWithAudit(req SaveNotificationPreferenceRequest, audit RecordAuditLogRequest) (*NotificationPreference, *AuditLog, error) {
	normalized, err := normalizeSaveNotificationPreferenceRequest(req)
	if err != nil {
		return nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = normalized.UpdatedAt
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != "admin.notification_preferences.saved" || log.TargetType != "notification_preference" {
		return nil, log, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	preference, err := s.saveNotificationPreferenceLocked(normalized)
	if err != nil {
		return nil, log, err
	}
	log.TargetID = preference.ID
	log.Payload = notificationPreferenceAuditPayload(preference)
	auditLog := s.appendAuditLogLocked(log)
	s.enqueueNotificationPreferenceChangedOutboxLocked(preference)
	return preference, auditLog, nil
}

func (s *Store) SaveNotificationPreferenceBatchWithAudit(req SaveNotificationPreferenceBatchRequest, audit RecordAuditLogRequest) (*NotificationPreferenceBatchSaveResult, *AuditLog, error) {
	normalized, err := normalizeSaveNotificationPreferenceBatchRequest(req)
	if err != nil {
		return nil, nil, err
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = normalized.UpdatedAt
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if !isNotificationPreferenceBatchAuditAction(log.Action) || log.TargetType != "notification_preference_batch" {
		return nil, log, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, preferenceRequest := range normalized.Preferences {
		if preferenceRequest.TargetRole == "merchant" && preferenceRequest.TargetID != "" && s.merchants[preferenceRequest.TargetID] == nil {
			return nil, log, ErrNotFound
		}
	}
	preferences := make([]NotificationPreference, 0, len(normalized.Preferences))
	for _, preferenceRequest := range normalized.Preferences {
		preference, err := s.saveNotificationPreferenceLocked(preferenceRequest)
		if err != nil {
			return nil, log, err
		}
		preferences = append(preferences, *preference)
		s.enqueueNotificationPreferenceChangedOutboxLocked(preference)
	}
	result := notificationPreferenceBatchSaveResult(preferences, normalized.Reason, normalized.UpdatedAt)
	log.TargetID = result.BatchID
	log.Payload = mergeAuditPayload(log.Payload, notificationPreferenceBatchAuditPayload(result))
	auditLog := s.appendAuditLogLocked(log)
	return result, auditLog, nil
}

func (s *Store) NotificationPreferences(req NotificationPreferenceListRequest) ([]NotificationPreference, error) {
	req = normalizeNotificationPreferenceListRequest(req)
	if req.PreferenceKey == "" && ((req.TargetID != "" && req.TargetRole == "") || (req.NotificationType != "" && req.TargetRole != "" && req.TargetID == "")) {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	preferences := make([]NotificationPreference, 0)
	for _, preference := range s.notificationPreferences {
		if preference == nil {
			continue
		}
		if req.PreferenceKey != "" && preference.PreferenceKey != req.PreferenceKey {
			continue
		}
		if req.TargetRole != "" && preference.TargetRole != req.TargetRole {
			continue
		}
		if req.TargetID != "" && preference.TargetID != req.TargetID {
			continue
		}
		if req.NotificationType != "" && preference.NotificationType != req.NotificationType {
			continue
		}
		preferences = append(preferences, *cloneNotificationPreference(preference))
	}
	sort.SliceStable(preferences, func(i, j int) bool {
		if !preferences[i].UpdatedAt.Equal(preferences[j].UpdatedAt) {
			return preferences[i].UpdatedAt.After(preferences[j].UpdatedAt)
		}
		return preferences[i].PreferenceKey < preferences[j].PreferenceKey
	})
	if len(preferences) > req.Limit {
		preferences = preferences[:req.Limit]
	}
	return preferences, nil
}

func (s *Store) saveNotificationPreferenceLocked(req SaveNotificationPreferenceRequest) (*NotificationPreference, error) {
	if req.TargetRole == "merchant" && req.TargetID != "" && s.merchants[req.TargetID] == nil {
		return nil, ErrNotFound
	}
	preferenceKey := notificationPreferenceKey(req.TargetRole, req.TargetID, req.NotificationType)
	now := req.UpdatedAt
	id := s.notificationPreferenceByKey[preferenceKey]
	preference := s.notificationPreferences[id]
	if preference == nil {
		id = "ntfp_" + shortHash(preferenceKey)
		preference = &NotificationPreference{
			ID:            id,
			PreferenceKey: preferenceKey,
			CreatedAt:     now,
		}
		s.notificationPreferences[id] = preference
		s.notificationPreferenceByKey[preferenceKey] = id
	}
	preference.TargetRole = req.TargetRole
	preference.TargetID = req.TargetID
	preference.NotificationType = req.NotificationType
	preference.EnabledChannels = append([]string{}, req.EnabledChannels...)
	preference.DisabledChannels = append([]string{}, req.DisabledChannels...)
	preference.QuietHours = cloneNotificationQuietHours(req.QuietHours)
	preference.UpdatedAt = now
	return cloneNotificationPreference(preference), nil
}

func normalizeSaveNotificationPreferenceRequest(req SaveNotificationPreferenceRequest) (SaveNotificationPreferenceRequest, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.NotificationType = strings.TrimSpace(req.NotificationType)
	if req.TargetID != "" && req.TargetRole == "" {
		return req, ErrInvalidArgument
	}
	if req.NotificationType != "" && req.TargetRole != "" && req.TargetID == "" {
		return req, ErrInvalidArgument
	}
	if req.UpdatedAt.IsZero() {
		req.UpdatedAt = time.Now().UTC()
	} else {
		req.UpdatedAt = req.UpdatedAt.UTC()
	}
	var err error
	req.EnabledChannels, err = normalizeNotificationPreferenceChannels(req.EnabledChannels)
	if err != nil {
		return req, err
	}
	req.DisabledChannels, err = normalizeNotificationPreferenceChannels(req.DisabledChannels)
	if err != nil {
		return req, err
	}
	if notificationPreferenceChannelsOverlap(req.EnabledChannels, req.DisabledChannels) {
		return req, ErrInvalidArgument
	}
	req.QuietHours, err = normalizeNotificationQuietHours(req.QuietHours)
	if err != nil {
		return req, err
	}
	return req, nil
}

func normalizeSaveNotificationPreferenceBatchRequest(req SaveNotificationPreferenceBatchRequest) (SaveNotificationPreferenceBatchRequest, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.Preferences) == 0 || len(req.Preferences) > maxNotificationPreferenceBatchSize {
		return req, ErrInvalidArgument
	}
	if req.UpdatedAt.IsZero() {
		req.UpdatedAt = time.Now().UTC()
	} else {
		req.UpdatedAt = req.UpdatedAt.UTC()
	}
	normalizedPreferences := make([]SaveNotificationPreferenceRequest, 0, len(req.Preferences))
	seen := map[string]bool{}
	for _, item := range req.Preferences {
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = req.UpdatedAt
		}
		normalized, err := normalizeSaveNotificationPreferenceRequest(item)
		if err != nil {
			return req, err
		}
		preferenceKey := notificationPreferenceKey(normalized.TargetRole, normalized.TargetID, normalized.NotificationType)
		if seen[preferenceKey] {
			return req, ErrInvalidArgument
		}
		seen[preferenceKey] = true
		normalizedPreferences = append(normalizedPreferences, normalized)
	}
	req.Preferences = normalizedPreferences
	return req, nil
}

func normalizeNotificationPreferenceListRequest(req NotificationPreferenceListRequest) NotificationPreferenceListRequest {
	req.PreferenceKey = strings.TrimSpace(req.PreferenceKey)
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.NotificationType = strings.TrimSpace(req.NotificationType)
	if req.PreferenceKey != "" {
		req.TargetRole = ""
		req.TargetID = ""
		req.NotificationType = ""
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	return req
}

func normalizeNotificationPreferenceChannels(channels []string) ([]string, error) {
	normalized := sanitizedStringSlice(channels)
	for _, channel := range normalized {
		if !isNotificationProviderChannel(channel) {
			return normalized, ErrInvalidArgument
		}
	}
	return normalized, nil
}

func normalizeNotificationQuietHours(quiet NotificationQuietHours) (NotificationQuietHours, error) {
	quiet.Start = strings.TrimSpace(quiet.Start)
	quiet.End = strings.TrimSpace(quiet.End)
	quiet.TimezoneOffset = strings.TrimSpace(quiet.TimezoneOffset)
	quiet.Status = strings.TrimSpace(quiet.Status)
	quiet.Channels = sanitizedStringSlice(quiet.Channels)
	quiet.ExemptTypes = sanitizedStringSlice(quiet.ExemptTypes)
	if quiet.TimezoneOffset == "" {
		quiet.TimezoneOffset = "+08:00"
	}
	if quiet.Status == "" {
		quiet.Status = NotificationDeliveryQueued
	}
	for _, channel := range quiet.Channels {
		if !isNotificationProviderChannel(channel) {
			return quiet, ErrInvalidArgument
		}
	}
	if quiet.Start == "" && quiet.End == "" {
		quiet.Enabled = false
		return quiet, nil
	}
	if !validNotificationClock(quiet.Start) || !validNotificationClock(quiet.End) || quiet.Start == quiet.End || !validNotificationTimezoneOffset(quiet.TimezoneOffset) {
		return quiet, ErrInvalidArgument
	}
	switch quiet.Status {
	case NotificationDeliveryQueued, NotificationDeliveryFailed:
	default:
		return quiet, ErrInvalidArgument
	}
	return quiet, nil
}

func notificationPreferenceChannelsOverlap(enabled []string, disabled []string) bool {
	seen := map[string]bool{}
	for _, channel := range enabled {
		seen[channel] = true
	}
	for _, channel := range disabled {
		if seen[channel] {
			return true
		}
	}
	return false
}

func isNotificationProviderChannel(channel string) bool {
	switch strings.TrimSpace(channel) {
	case NotificationWechatSubscribe, NotificationSMS, NotificationEnterpriseWechat, NotificationPush:
		return true
	default:
		return false
	}
}

func validNotificationClock(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 || len(parts[0]) < 1 || len(parts[0]) > 2 || len(parts[1]) != 2 {
		return false
	}
	hour := parseTwoDigitNumber(parts[0])
	minute := parseTwoDigitNumber(parts[1])
	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}

func validNotificationTimezoneOffset(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 6 || (value[0] != '+' && value[0] != '-') || value[3] != ':' {
		return false
	}
	hour := parseTwoDigitNumber(value[1:3])
	minute := parseTwoDigitNumber(value[4:6])
	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}

func parseTwoDigitNumber(value string) int {
	if value == "" {
		return -1
	}
	total := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return -1
		}
		total = total*10 + int(char-'0')
	}
	return total
}

func notificationPreferenceKey(targetRole string, targetID string, notificationType string) string {
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

func notificationPreferenceBatchSaveResult(preferences []NotificationPreference, reason string, updatedAt time.Time) *NotificationPreferenceBatchSaveResult {
	copied := make([]NotificationPreference, 0, len(preferences))
	keys := make([]string, 0, len(preferences))
	for _, preference := range preferences {
		preferenceCopy := *cloneNotificationPreference(&preference)
		copied = append(copied, preferenceCopy)
		keys = append(keys, preference.PreferenceKey)
	}
	sort.Strings(keys)
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	} else {
		updatedAt = updatedAt.UTC()
	}
	return &NotificationPreferenceBatchSaveResult{
		BatchID:        "ntfp_batch_" + shortHash(strings.Join(keys, "|")+"|"+updatedAt.Format(time.RFC3339Nano)),
		Preferences:    copied,
		Saved:          len(copied),
		PreferenceKeys: keys,
		Reason:         strings.TrimSpace(reason),
		UpdatedAt:      updatedAt,
	}
}

func notificationPreferenceChangedIdempotencyKey(preference *NotificationPreference) string {
	if preference == nil || strings.TrimSpace(preference.PreferenceKey) == "" || preference.UpdatedAt.IsZero() {
		return ""
	}
	return fmt.Sprintf("notification_preference_changed:%s:%s", preference.PreferenceKey, preference.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func notificationPreferenceChangedOutboxPayload(preference *NotificationPreference) map[string]any {
	if preference == nil {
		return map[string]any{}
	}
	return map[string]any{
		"preference_id":     preference.ID,
		"preference_key":    preference.PreferenceKey,
		"preference_keys":   []string{preference.PreferenceKey},
		"target_role":       preference.TargetRole,
		"target_id":         preference.TargetID,
		"notification_type": preference.NotificationType,
		"enabled_channels":  append([]string{}, preference.EnabledChannels...),
		"disabled_channels": append([]string{}, preference.DisabledChannels...),
		"quiet_hours":       cloneNotificationQuietHours(preference.QuietHours),
		"updated_at":        preference.UpdatedAt.Format(time.RFC3339Nano),
		"idempotency_key":   notificationPreferenceChangedIdempotencyKey(preference),
	}
}

func notificationPreferenceChangedOutboxEvent(preference *NotificationPreference) *OutboxEvent {
	if preference == nil {
		return nil
	}
	idempotencyKey := notificationPreferenceChangedIdempotencyKey(preference)
	if idempotencyKey == "" {
		return nil
	}
	now := preference.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return &OutboxEvent{
		ID:             "obe_notification_preference_" + shortHash(idempotencyKey),
		Topic:          notificationPreferenceChangedTopic,
		AggregateType:  "notification_preference",
		AggregateID:    preference.ID,
		EventType:      "notification.preferences.changed",
		IdempotencyKey: idempotencyKey,
		Payload:        notificationPreferenceChangedOutboxPayload(preference),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *Store) enqueueNotificationPreferenceChangedOutboxLocked(preference *NotificationPreference) *OutboxEvent {
	event := notificationPreferenceChangedOutboxEvent(preference)
	if event == nil {
		return nil
	}
	return s.applyOutboxEventLocked(event)
}

func (s *Store) applyOutboxEventLocked(event *OutboxEvent) *OutboxEvent {
	if event == nil || strings.TrimSpace(event.ID) == "" {
		return nil
	}
	if event.IdempotencyKey != "" {
		if existingID := s.outboxByIdempotency[event.IdempotencyKey]; existingID != "" {
			return cloneOutboxEvent(s.outboxEvents[existingID])
		}
	}
	eventCopy := *cloneOutboxEvent(event)
	s.outboxEvents[eventCopy.ID] = &eventCopy
	if eventCopy.IdempotencyKey != "" {
		s.outboxByIdempotency[eventCopy.IdempotencyKey] = eventCopy.ID
	}
	return cloneOutboxEvent(&eventCopy)
}

func notificationPreferenceAuditPayload(preference *NotificationPreference) map[string]any {
	if preference == nil {
		return map[string]any{}
	}
	return map[string]any{
		"preference_key":    preference.PreferenceKey,
		"target_role":       preference.TargetRole,
		"target_id":         preference.TargetID,
		"notification_type": preference.NotificationType,
		"enabled_channels":  append([]string{}, preference.EnabledChannels...),
		"disabled_channels": append([]string{}, preference.DisabledChannels...),
		"quiet_hours":       cloneNotificationQuietHours(preference.QuietHours),
		"updated_at":        preference.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func notificationPreferenceBatchAuditPayload(result *NotificationPreferenceBatchSaveResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	return map[string]any{
		"batch_id":        result.BatchID,
		"saved":           result.Saved,
		"preference_keys": append([]string{}, result.PreferenceKeys...),
		"reason":          result.Reason,
		"updated_at":      result.UpdatedAt.Format(time.RFC3339Nano),
		"preferences":     notificationPreferenceAuditSummaries(result.Preferences),
	}
}

func isNotificationPreferenceBatchAuditAction(action string) bool {
	switch strings.TrimSpace(action) {
	case "admin.notification_preferences.batch_saved", "admin.notification_preferences.change_applied":
		return true
	default:
		return false
	}
}

func mergeAuditPayload(base map[string]any, overlay map[string]any) map[string]any {
	output := map[string]any{}
	for key, value := range base {
		output[key] = cloneAny(value)
	}
	for key, value := range overlay {
		output[key] = cloneAny(value)
	}
	return output
}

func notificationPreferenceAuditSummaries(preferences []NotificationPreference) []map[string]any {
	items := make([]map[string]any, 0, len(preferences))
	for _, preference := range preferences {
		items = append(items, notificationPreferenceAuditPayload(&preference))
	}
	return items
}

func normalizeCreateNotificationRequest(req CreateNotificationRequest) (CreateNotificationRequest, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Type = strings.TrimSpace(req.Type)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.SourceTopic = strings.TrimSpace(req.SourceTopic)
	req.SourceEventID = strings.TrimSpace(req.SourceEventID)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.Channel == "" {
		req.Channel = NotificationChannelInApp
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now().UTC()
	} else {
		req.CreatedAt = req.CreatedAt.UTC()
	}
	if req.TargetRole == "" || req.TargetID == "" || req.Type == "" || req.Title == "" || req.Body == "" || req.IdempotencyKey == "" {
		return req, ErrInvalidArgument
	}
	switch req.Channel {
	case NotificationChannelInApp, NotificationWechatSubscribe:
	default:
		return req, ErrInvalidArgument
	}
	return req, nil
}

func normalizeNotificationListRequest(req NotificationListRequest) NotificationListRequest {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Status = strings.TrimSpace(req.Status)
	req.SourceTopic = strings.TrimSpace(req.SourceTopic)
	req.SourceEventID = strings.TrimSpace(req.SourceEventID)
	if req.Status == "" {
		req.Status = "all"
	}
	switch req.Status {
	case "all", NotificationStatusUnread, NotificationStatusRead:
	default:
		req.Status = "all"
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	return req
}

func normalizeRecordNotificationDeliveryRequest(req RecordNotificationDeliveryRequest) (RecordNotificationDeliveryRequest, error) {
	req.NotificationID = strings.TrimSpace(req.NotificationID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Provider = strings.TrimSpace(req.Provider)
	req.Status = strings.TrimSpace(req.Status)
	req.ProviderMessageID = strings.TrimSpace(req.ProviderMessageID)
	req.ErrorCode = strings.TrimSpace(req.ErrorCode)
	req.ErrorMessage = strings.TrimSpace(req.ErrorMessage)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.Status == "" {
		req.Status = NotificationDeliveryDelivered
	}
	if req.AttemptedAt.IsZero() {
		req.AttemptedAt = time.Now().UTC()
	} else {
		req.AttemptedAt = req.AttemptedAt.UTC()
	}
	if !req.DeliveredAt.IsZero() {
		req.DeliveredAt = req.DeliveredAt.UTC()
	}
	if !req.RetryAt.IsZero() {
		req.RetryAt = req.RetryAt.UTC()
	}
	if req.NotificationID == "" || req.IdempotencyKey == "" {
		return req, ErrInvalidArgument
	}
	switch req.Status {
	case NotificationDeliveryQueued, NotificationDeliveryDelivered, NotificationDeliveryFailed:
	default:
		return req, ErrInvalidArgument
	}
	if req.Status == NotificationDeliveryFailed && req.ErrorMessage == "" && req.ErrorCode == "" {
		return req, ErrInvalidArgument
	}
	return req, nil
}

func normalizeNotificationDeliveryListRequest(req NotificationDeliveryListRequest) NotificationDeliveryListRequest {
	req.NotificationID = strings.TrimSpace(req.NotificationID)
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Provider = strings.TrimSpace(req.Provider)
	req.Status = strings.TrimSpace(req.Status)
	req.ErrorCode = strings.TrimSpace(req.ErrorCode)
	if !req.RetryAtBefore.IsZero() {
		req.RetryAtBefore = req.RetryAtBefore.UTC()
	}
	if req.Status == "" {
		req.Status = "all"
	}
	switch req.Status {
	case "all", NotificationDeliveryQueued, NotificationDeliveryDelivered, NotificationDeliveryFailed:
	default:
		req.Status = "all"
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	return req
}

func normalizeNotificationFailureAlertEmissionRequest(req NotificationFailureAlertEmissionRequest) (NotificationFailureAlertEmissionRequest, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Provider = strings.TrimSpace(req.Provider)
	if (req.TargetRole == "") != (req.TargetID == "") {
		return req, ErrInvalidArgument
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	return req, nil
}

func normalizeNotificationDeliveryRetryScheduleRequest(req NotificationDeliveryRetryScheduleRequest) (NotificationDeliveryRetryScheduleRequest, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Provider = strings.TrimSpace(req.Provider)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.ErrorCode = strings.TrimSpace(req.ErrorCode)
	if !req.SourceRetryAtBefore.IsZero() {
		req.SourceRetryAtBefore = req.SourceRetryAtBefore.UTC()
	}
	if (req.TargetRole == "") != (req.TargetID == "") {
		return req, ErrInvalidArgument
	}
	if req.Status == "" {
		req.Status = NotificationDeliveryFailed
	}
	switch req.Status {
	case NotificationDeliveryQueued, NotificationDeliveryFailed:
	default:
		return req, ErrInvalidArgument
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	if req.RetryAfterSeconds < 0 {
		return req, ErrInvalidArgument
	}
	if !req.RetryAt.IsZero() {
		req.RetryAt = req.RetryAt.UTC()
		if req.RetryAt.Before(req.Now) {
			req.RetryAt = req.Now
		}
		req.RetryAfterSeconds = int(math.Ceil(req.RetryAt.Sub(req.Now).Seconds()))
	} else {
		if req.RetryAfterSeconds == 0 {
			req.RetryAfterSeconds = notificationDeliveryRetryDefaultBackoffSeconds(req.Channel, req.Provider)
		}
		if req.RetryAfterSeconds > 86400 {
			req.RetryAfterSeconds = 86400
		}
		req.RetryAt = req.Now.Add(time.Duration(req.RetryAfterSeconds) * time.Second)
	}
	return req, nil
}

func normalizeNotificationQuietWindowRetryScheduleRequest(req NotificationQuietWindowRetryScheduleRequest) (NotificationQuietWindowRetryScheduleRequest, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Provider = strings.TrimSpace(req.Provider)
	if (req.TargetRole == "") != (req.TargetID == "") {
		return req, ErrInvalidArgument
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.RetryAfterSeconds < 0 {
		return req, ErrInvalidArgument
	}
	if req.RetryAfterSeconds > 86400 {
		req.RetryAfterSeconds = 86400
	}
	return req, nil
}

func notificationFailureAlertDeliveriesFromMap(req NotificationFailureAlertEmissionRequest, source map[string]*PlatformNotificationDelivery) []PlatformNotificationDelivery {
	deliveries := make([]PlatformNotificationDelivery, 0)
	for _, delivery := range source {
		if delivery == nil || delivery.Status != NotificationDeliveryFailed {
			continue
		}
		if req.TargetRole != "" && delivery.TargetRole != req.TargetRole {
			continue
		}
		if req.TargetID != "" && delivery.TargetID != req.TargetID {
			continue
		}
		if req.Channel != "" && delivery.Channel != req.Channel {
			continue
		}
		if req.Provider != "" && delivery.Provider != req.Provider {
			continue
		}
		deliveries = append(deliveries, *clonePlatformNotificationDelivery(delivery))
	}
	sort.SliceStable(deliveries, func(i, j int) bool {
		if !deliveries[i].AttemptedAt.Equal(deliveries[j].AttemptedAt) {
			return deliveries[i].AttemptedAt.After(deliveries[j].AttemptedAt)
		}
		return deliveries[i].ID > deliveries[j].ID
	})
	if len(deliveries) > req.Limit {
		deliveries = deliveries[:req.Limit]
	}
	return deliveries
}

func notificationDeliveryRetryDeliveriesFromMap(req NotificationDeliveryRetryScheduleRequest, source map[string]*PlatformNotificationDelivery) []PlatformNotificationDelivery {
	deliveries := make([]PlatformNotificationDelivery, 0)
	for _, delivery := range source {
		if delivery == nil || delivery.Status != req.Status {
			continue
		}
		if req.ErrorCode != "" && delivery.ErrorCode != req.ErrorCode {
			continue
		}
		if !req.SourceRetryAtBefore.IsZero() && (delivery.RetryAt.IsZero() || delivery.RetryAt.After(req.SourceRetryAtBefore)) {
			continue
		}
		if req.TargetRole != "" && delivery.TargetRole != req.TargetRole {
			continue
		}
		if req.TargetID != "" && delivery.TargetID != req.TargetID {
			continue
		}
		if req.Channel != "" && delivery.Channel != req.Channel {
			continue
		}
		if req.Provider != "" && delivery.Provider != req.Provider {
			continue
		}
		deliveries = append(deliveries, *clonePlatformNotificationDelivery(delivery))
	}
	sort.SliceStable(deliveries, func(i, j int) bool {
		if !deliveries[i].AttemptedAt.Equal(deliveries[j].AttemptedAt) {
			return deliveries[i].AttemptedAt.After(deliveries[j].AttemptedAt)
		}
		return deliveries[i].ID > deliveries[j].ID
	})
	if len(deliveries) > req.Limit {
		deliveries = deliveries[:req.Limit]
	}
	return deliveries
}

func notificationFailureAlertEmissionFromDeliveries(req NotificationFailureAlertEmissionRequest, deliveries []PlatformNotificationDelivery) *NotificationFailureAlertEmission {
	emission := &NotificationFailureAlertEmission{
		Status:     "skipped",
		Topic:      notificationFailureAlertTopic,
		TargetRole: req.TargetRole,
		TargetID:   req.TargetID,
		Channel:    req.Channel,
		Provider:   req.Provider,
		EmittedAt:  req.Now,
		Deliveries: append([]PlatformNotificationDelivery{}, deliveries...),
	}
	emission.FailedCount = len(emission.Deliveries)
	if emission.FailedCount > 0 {
		emission.Status = "emitted"
		emission.IdempotencyKey = notificationFailureAlertIdempotencyKey(req, emission.Deliveries)
	}
	return emission
}

func notificationDeliveryRetryScheduleFromDeliveries(req NotificationDeliveryRetryScheduleRequest, deliveries []PlatformNotificationDelivery, notifications []PlatformNotification) *NotificationDeliveryRetrySchedule {
	schedule := &NotificationDeliveryRetrySchedule{
		Status:            "skipped",
		Topic:             notificationDeliveryRetryTopic,
		TargetRole:        req.TargetRole,
		TargetID:          req.TargetID,
		Channel:           req.Channel,
		Provider:          req.Provider,
		DeliveryStatus:    req.Status,
		ErrorCode:         req.ErrorCode,
		RetryPolicy:       notificationDeliveryRetryPolicy(req.Channel, req.Provider, req.RetryAfterSeconds),
		RetryAfterSeconds: req.RetryAfterSeconds,
		RetryAt:           req.RetryAt,
		ScheduledAt:       req.Now,
		Deliveries:        append([]PlatformNotificationDelivery{}, deliveries...),
		Notifications:     append([]PlatformNotification{}, notifications...),
	}
	schedule.ScheduledCount = len(schedule.Deliveries)
	if schedule.ScheduledCount > 0 {
		schedule.Status = "scheduled"
		schedule.IdempotencyKey = notificationDeliveryRetryIdempotencyKey(req, schedule.Deliveries)
	}
	return schedule
}

func notificationDeliveryRetryNotificationsFromMap(deliveries []PlatformNotificationDelivery, source map[string]*PlatformNotification) []PlatformNotification {
	seen := map[string]bool{}
	notifications := make([]PlatformNotification, 0)
	for _, delivery := range deliveries {
		notificationID := strings.TrimSpace(delivery.NotificationID)
		if notificationID == "" || seen[notificationID] {
			continue
		}
		notification := source[notificationID]
		if notification == nil {
			continue
		}
		seen[notificationID] = true
		notifications = append(notifications, *clonePlatformNotification(notification))
	}
	sort.SliceStable(notifications, func(i, j int) bool {
		if !notifications[i].CreatedAt.Equal(notifications[j].CreatedAt) {
			return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
		}
		return notifications[i].ID > notifications[j].ID
	})
	return notifications
}

func (s *Store) notificationDeliveryRetryNotificationsSnapshot(deliveries []PlatformNotificationDelivery) []PlatformNotification {
	s.mu.Lock()
	defer s.mu.Unlock()
	return notificationDeliveryRetryNotificationsFromMap(deliveries, s.notifications)
}

func notificationFailureAlertIdempotencyKey(req NotificationFailureAlertEmissionRequest, deliveries []PlatformNotificationDelivery) string {
	if len(deliveries) == 0 {
		return ""
	}
	latest := deliveries[0]
	parts := []string{
		"notification_delivery_failed_alerts",
		req.Now.UTC().Format("2006-01-02"),
		"target_role:" + req.TargetRole,
		"target_id:" + req.TargetID,
		"channel:" + req.Channel,
		"provider:" + req.Provider,
		fmt.Sprintf("failed:%d", len(deliveries)),
		"latest:" + latest.ID,
		latest.AttemptedAt.UTC().Format(time.RFC3339Nano),
	}
	return strings.Join(parts, ":")
}

func notificationDeliveryRetryIdempotencyKey(req NotificationDeliveryRetryScheduleRequest, deliveries []PlatformNotificationDelivery) string {
	if len(deliveries) == 0 {
		return ""
	}
	latest := deliveries[0]
	parts := []string{
		"notification_delivery_retries",
		req.Now.UTC().Format("2006-01-02"),
		"target_role:" + req.TargetRole,
		"target_id:" + req.TargetID,
		"channel:" + req.Channel,
		"provider:" + req.Provider,
		"status:" + req.Status,
		"error_code:" + req.ErrorCode,
		fmt.Sprintf("deliveries:%d", len(deliveries)),
		fmt.Sprintf("retry_after:%d", req.RetryAfterSeconds),
		"latest:" + latest.ID,
		latest.AttemptedAt.UTC().Format(time.RFC3339Nano),
	}
	return strings.Join(parts, ":")
}

func notificationFailureAlertOutboxPayload(emission *NotificationFailureAlertEmission) map[string]any {
	if emission == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"failed_count":    emission.FailedCount,
		"target_role":     emission.TargetRole,
		"target_id":       emission.TargetID,
		"channel":         emission.Channel,
		"provider":        emission.Provider,
		"emitted_at":      emission.EmittedAt.Format(time.RFC3339Nano),
		"idempotency_key": emission.IdempotencyKey,
		"deliveries":      emission.Deliveries,
	}
	if len(emission.Deliveries) > 0 {
		first := emission.Deliveries[0]
		payload["notification_id"] = first.NotificationID
		payload["delivery_id"] = first.ID
		payload["error_code"] = first.ErrorCode
		payload["error_message"] = first.ErrorMessage
	}
	return payload
}

func notificationDeliveryRetryOutboxPayload(schedule *NotificationDeliveryRetrySchedule) map[string]any {
	if schedule == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"scheduled_count":     schedule.ScheduledCount,
		"target_role":         schedule.TargetRole,
		"target_id":           schedule.TargetID,
		"channel":             schedule.Channel,
		"provider":            schedule.Provider,
		"delivery_status":     schedule.DeliveryStatus,
		"error_code":          schedule.ErrorCode,
		"retry_policy":        schedule.RetryPolicy,
		"retry_after_seconds": schedule.RetryAfterSeconds,
		"retry_at":            schedule.RetryAt.Format(time.RFC3339Nano),
		"scheduled_at":        schedule.ScheduledAt.Format(time.RFC3339Nano),
		"idempotency_key":     schedule.IdempotencyKey,
		"deliveries":          schedule.Deliveries,
		"notifications":       schedule.Notifications,
	}
	if len(schedule.Deliveries) > 0 {
		first := schedule.Deliveries[0]
		payload["notification_id"] = first.NotificationID
		payload["delivery_id"] = first.ID
		payload["error_code"] = first.ErrorCode
		payload["error_message"] = first.ErrorMessage
	}
	return payload
}

func notificationFailureAlertOutboxEvent(emission *NotificationFailureAlertEmission) *OutboxEvent {
	if emission == nil || strings.TrimSpace(emission.IdempotencyKey) == "" {
		return nil
	}
	now := emission.EmittedAt
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	eventID := "obe_notification_failure_alert_" + shortHash(emission.IdempotencyKey)
	emission.OutboxEventID = eventID
	return &OutboxEvent{
		ID:             eventID,
		Topic:          notificationFailureAlertTopic,
		AggregateType:  "notification_delivery",
		AggregateID:    "failed",
		EventType:      "notification.delivery_failed_alerts.emitted",
		IdempotencyKey: emission.IdempotencyKey,
		Payload:        notificationFailureAlertOutboxPayload(emission),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func notificationDeliveryRetryOutboxEvent(schedule *NotificationDeliveryRetrySchedule) *OutboxEvent {
	if schedule == nil || strings.TrimSpace(schedule.IdempotencyKey) == "" {
		return nil
	}
	scheduledAt := schedule.ScheduledAt
	if scheduledAt.IsZero() {
		scheduledAt = time.Now().UTC()
	} else {
		scheduledAt = scheduledAt.UTC()
	}
	retryAt := schedule.RetryAt
	if retryAt.IsZero() {
		retryAt = scheduledAt
	} else {
		retryAt = retryAt.UTC()
	}
	eventID := "obe_notification_retry_" + shortHash(schedule.IdempotencyKey)
	schedule.OutboxEventID = eventID
	return &OutboxEvent{
		ID:             eventID,
		Topic:          notificationDeliveryRetryTopic,
		AggregateType:  "notification_delivery",
		AggregateID:    schedule.DeliveryStatus,
		EventType:      "notification.delivery_retries.scheduled",
		IdempotencyKey: schedule.IdempotencyKey,
		Payload:        notificationDeliveryRetryOutboxPayload(schedule),
		Status:         OutboxStatusPending,
		AvailableAt:    retryAt,
		CreatedAt:      scheduledAt,
		UpdatedAt:      scheduledAt,
	}
}

func notificationFailureAlertEmissionAuditPayload(emission *NotificationFailureAlertEmission) map[string]any {
	if emission == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"failed_count":    emission.FailedCount,
		"status":          emission.Status,
		"topic":           emission.Topic,
		"target_role":     emission.TargetRole,
		"target_id":       emission.TargetID,
		"channel":         emission.Channel,
		"provider":        emission.Provider,
		"idempotency_key": emission.IdempotencyKey,
		"outbox_event_id": emission.OutboxEventID,
	}
	if len(emission.Deliveries) > 0 {
		first := emission.Deliveries[0]
		payload["delivery_id"] = first.ID
		payload["notification_id"] = first.NotificationID
		payload["error_code"] = first.ErrorCode
		payload["error_message"] = first.ErrorMessage
	}
	return payload
}

func notificationDeliveryRetryScheduleAuditPayload(schedule *NotificationDeliveryRetrySchedule) map[string]any {
	if schedule == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"scheduled_count":     schedule.ScheduledCount,
		"status":              schedule.Status,
		"topic":               schedule.Topic,
		"target_role":         schedule.TargetRole,
		"target_id":           schedule.TargetID,
		"channel":             schedule.Channel,
		"provider":            schedule.Provider,
		"delivery_status":     schedule.DeliveryStatus,
		"error_code":          schedule.ErrorCode,
		"retry_policy":        schedule.RetryPolicy,
		"retry_after_seconds": schedule.RetryAfterSeconds,
		"retry_at":            schedule.RetryAt.Format(time.RFC3339Nano),
		"scheduled_at":        schedule.ScheduledAt.Format(time.RFC3339Nano),
		"idempotency_key":     schedule.IdempotencyKey,
		"outbox_event_id":     schedule.OutboxEventID,
	}
	if len(schedule.Deliveries) > 0 {
		first := schedule.Deliveries[0]
		payload["delivery_id"] = first.ID
		payload["notification_id"] = first.NotificationID
		payload["error_code"] = first.ErrorCode
		payload["error_message"] = first.ErrorMessage
	}
	return payload
}

func notificationDeliveryRetryDefaultBackoffSeconds(channel string, provider string) int {
	key := strings.TrimSpace(provider)
	if key == "" {
		key = strings.TrimSpace(channel)
	}
	switch key {
	case NotificationChannelInApp:
		return 30
	case NotificationWechatSubscribe:
		return 300
	case "sms":
		return 600
	case "enterprise_wechat":
		return 300
	case "push":
		return 120
	default:
		return 300
	}
}

func notificationDeliveryRetryPolicy(channel string, provider string, retryAfterSeconds int) string {
	key := strings.TrimSpace(provider)
	if key == "" {
		key = strings.TrimSpace(channel)
	}
	if key == "" {
		key = "default"
	}
	return fmt.Sprintf("%s_backoff_%ds", key, retryAfterSeconds)
}

func normalizeAuditLogsRequest(req AuditLogsRequest) AuditLogsRequest {
	req.ActorType = strings.TrimSpace(req.ActorType)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.Action = strings.TrimSpace(req.Action)
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.TargetID = strings.TrimSpace(req.TargetID)
	if !req.After.IsZero() {
		req.After = req.After.UTC()
	}
	if !req.Before.IsZero() {
		req.Before = req.Before.UTC()
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	return req
}

const (
	defaultAuditRetentionDays                = 2555
	defaultAuditHotDays                      = 180
	defaultAuditIntegritySampleLimit         = 500
	maxAuditIntegritySampleLimit             = 5000
	auditRetentionAlertTopic                 = "audit.retention_alerts"
	notificationFailureAlertTopic            = "notification.delivery_failed_alerts"
	notificationDeliveryRetryTopic           = "notification.delivery_retries"
	notificationPreferenceChangedTopic       = "notification.preferences_changed"
	maxNotificationPreferenceBatchSize       = 50
	defaultAuditArchiveLimit                 = 500
	maxAuditArchiveLimit                     = 5000
	defaultAuditArchiveListLimit             = 100
	maxAuditArchiveListLimit                 = 1000
	defaultAuditArchiveStoragePrefix         = "worm://audit-logs"
	defaultAuditArchiveVerificationListLimit = 100
	maxAuditArchiveVerificationListLimit     = 1000
	auditArchiveManifestAlgorithm            = "sha256:v1"
	auditArchiveRequestedTopic               = "audit.archive_requested"
	auditArchiveCompletedAction              = "admin.audit_archive.completed"
	auditArchiveVerifiedAction               = "admin.audit_archive.verified"
)

var defaultCriticalAuditActions = []string{
	"admin.refund_settings.updated",
	"admin.order.refunded",
	"after_sales.reviewed",
	"admin.order_state.compensated",
	"admin.object_cleanup.completed",
	"admin.object_cleanup.failed",
	"admin.outbox.claimed",
	"admin.outbox.replayed",
	"admin.merchant_invite.created",
	"admin.rider_invite.created",
	"admin.rbac.change_requested",
	"admin.rbac.change_reviewed",
	"admin.rbac.change_applied",
	"admin.rbac.change_rolled_back",
	"admin.notification_preferences.change_requested",
	"admin.notification_preferences.change_reviewed",
	"admin.notification_preferences.change_applied",
	"admin.audit_logs.exported",
}

func normalizeAuditRetentionReportRequest(req AuditRetentionReportRequest) AuditRetentionReportRequest {
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	if req.RetentionDays <= 0 {
		req.RetentionDays = defaultAuditRetentionDays
	}
	if req.HotDays <= 0 {
		req.HotDays = defaultAuditHotDays
	}
	if req.HotDays > req.RetentionDays {
		req.HotDays = req.RetentionDays
	}
	if req.IntegritySampleLimit <= 0 {
		req.IntegritySampleLimit = defaultAuditIntegritySampleLimit
	}
	if req.IntegritySampleLimit > maxAuditIntegritySampleLimit {
		req.IntegritySampleLimit = maxAuditIntegritySampleLimit
	}
	actions := make([]string, 0, len(req.CriticalActions))
	seen := map[string]struct{}{}
	for _, action := range req.CriticalActions {
		action = strings.TrimSpace(action)
		if action == "" {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		actions = append(actions, action)
	}
	if len(actions) == 0 {
		actions = append(actions, defaultCriticalAuditActions...)
	}
	req.CriticalActions = actions
	return req
}

func auditRetentionReportFromLogs(req AuditRetentionReportRequest, logs []AuditLog) *AuditRetentionReport {
	req = normalizeAuditRetentionReportRequest(req)
	sort.SliceStable(logs, func(i, j int) bool {
		if !logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].CreatedAt.After(logs[j].CreatedAt)
		}
		return logs[i].ID > logs[j].ID
	})

	actionCoverage := make([]AuditActionCoverage, 0, len(req.CriticalActions))
	actionCoverageByName := map[string]*AuditActionCoverage{}
	for _, action := range req.CriticalActions {
		coverage := AuditActionCoverage{Action: action}
		actionCoverage = append(actionCoverage, coverage)
		actionCoverageByName[action] = &actionCoverage[len(actionCoverage)-1]
	}

	report := &AuditRetentionReport{
		Status:                 "ok",
		GeneratedAt:            req.Now,
		RetentionDays:          req.RetentionDays,
		HotDays:                req.HotDays,
		RetentionCutoff:        req.Now.AddDate(0, 0, -req.RetentionDays),
		ColdArchiveCutoff:      req.Now.AddDate(0, 0, -req.HotDays),
		TotalLogs:              len(logs),
		CriticalActionCoverage: actionCoverage,
		Alerts:                 []AuditRetentionAlert{},
	}

	for index, log := range logs {
		createdAt := normalizeAuditLogTime(log.CreatedAt)
		if report.OldestCreatedAt.IsZero() || createdAt.Before(report.OldestCreatedAt) {
			report.OldestCreatedAt = createdAt
		}
		if report.NewestCreatedAt.IsZero() || createdAt.After(report.NewestCreatedAt) {
			report.NewestCreatedAt = createdAt
		}
		if createdAt.Before(report.RetentionCutoff) {
			report.ExpiredLogs++
		}
		if createdAt.Before(report.ColdArchiveCutoff) {
			report.ColdArchiveDueLogs++
		}
		if log.Action == "admin.audit_logs.exported" {
			report.ExportEvents++
		}
		if coverage, ok := actionCoverageByName[log.Action]; ok {
			coverage.Count++
			if coverage.LastCreatedAt.IsZero() || createdAt.After(coverage.LastCreatedAt) {
				coverage.LastCreatedAt = createdAt
			}
		}
		if index < req.IntegritySampleLimit {
			report.IntegritySampleSize++
			if !log.IntegrityVerified {
				report.IntegrityFailures++
			}
		}
	}
	report.CriticalActionCoverage = actionCoverage
	for _, coverage := range actionCoverage {
		if coverage.Count == 0 {
			report.MissingCriticalActions = append(report.MissingCriticalActions, coverage.Action)
		}
	}
	report.Alerts = auditRetentionAlertsForReport(report)
	report.Status = auditRetentionStatus(report.Alerts)
	return report
}

func auditRetentionAlertsForReport(report *AuditRetentionReport) []AuditRetentionAlert {
	if report == nil {
		return nil
	}
	alerts := []AuditRetentionAlert{}
	if report.TotalLogs == 0 {
		alerts = append(alerts, AuditRetentionAlert{
			Code:     "audit.no_logs",
			Severity: "critical",
			Message:  "audit ledger is empty",
		})
		return alerts
	}
	if report.IntegrityFailures > 0 {
		alerts = append(alerts, AuditRetentionAlert{
			Code:     "audit.integrity_failed",
			Severity: "critical",
			Message:  "audit integrity verification failed in sampled logs",
			Count:    report.IntegrityFailures,
		})
	}
	if report.ExpiredLogs > 0 {
		alerts = append(alerts, AuditRetentionAlert{
			Code:     "audit.retention_expired",
			Severity: "critical",
			Message:  "audit logs exceeded configured retention window",
			Count:    report.ExpiredLogs,
		})
	}
	if report.ColdArchiveDueLogs > 0 {
		alerts = append(alerts, AuditRetentionAlert{
			Code:     "audit.archive_due",
			Severity: "warning",
			Message:  "audit logs should be moved to cold or WORM archive",
			Count:    report.ColdArchiveDueLogs,
		})
	}
	if len(report.MissingCriticalActions) > 0 {
		alerts = append(alerts, AuditRetentionAlert{
			Code:     "audit.missing_critical_action",
			Severity: "warning",
			Message:  "critical audit action coverage is missing",
			Count:    len(report.MissingCriticalActions),
		})
	}
	return alerts
}

func auditRetentionStatus(alerts []AuditRetentionAlert) string {
	status := "ok"
	for _, alert := range alerts {
		switch alert.Severity {
		case "critical":
			return "critical"
		case "warning":
			status = "warning"
		}
	}
	return status
}

func auditRetentionAlertEmissionFromReport(report *AuditRetentionReport) *AuditRetentionAlertEmission {
	if report == nil {
		return &AuditRetentionAlertEmission{Status: "skipped", ReportStatus: "unknown", Alerts: []AuditRetentionAlert{}}
	}
	alerts := append([]AuditRetentionAlert{}, report.Alerts...)
	emission := &AuditRetentionAlertEmission{
		Status:       "skipped",
		ReportStatus: strings.TrimSpace(report.Status),
		AlertCount:   len(alerts),
		Topic:        auditRetentionAlertTopic,
		EmittedAt:    report.GeneratedAt,
		Alerts:       alerts,
		Report:       report,
	}
	for _, alert := range alerts {
		switch alert.Severity {
		case "critical":
			emission.CriticalCount++
		case "warning":
			emission.WarningCount++
		}
	}
	if len(alerts) > 0 {
		emission.Status = "emitted"
		emission.IdempotencyKey = auditRetentionAlertIdempotencyKey(report)
	}
	return emission
}

func auditRetentionAlertIdempotencyKey(report *AuditRetentionReport) string {
	if report == nil {
		return ""
	}
	parts := []string{
		"audit_retention_alerts",
		report.GeneratedAt.UTC().Format("2006-01-02"),
		fmt.Sprintf("retention:%d", report.RetentionDays),
		fmt.Sprintf("hot:%d", report.HotDays),
		fmt.Sprintf("status:%s", report.Status),
		fmt.Sprintf("expired:%d", report.ExpiredLogs),
		fmt.Sprintf("archive:%d", report.ColdArchiveDueLogs),
		fmt.Sprintf("integrity:%d", report.IntegrityFailures),
		fmt.Sprintf("missing:%d", len(report.MissingCriticalActions)),
	}
	return strings.Join(parts, ":")
}

func auditRetentionAlertOutboxPayload(emission *AuditRetentionAlertEmission) map[string]any {
	if emission == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"alert_count":     emission.AlertCount,
		"critical_count":  emission.CriticalCount,
		"warning_count":   emission.WarningCount,
		"report_status":   emission.ReportStatus,
		"emitted_at":      emission.EmittedAt.Format(time.RFC3339Nano),
		"idempotency_key": emission.IdempotencyKey,
		"alerts":          emission.Alerts,
	}
	if emission.Report != nil {
		payload["retention_days"] = emission.Report.RetentionDays
		payload["hot_days"] = emission.Report.HotDays
		payload["expired_logs"] = emission.Report.ExpiredLogs
		payload["cold_archive_due_logs"] = emission.Report.ColdArchiveDueLogs
		payload["integrity_failures"] = emission.Report.IntegrityFailures
		payload["missing_critical_actions"] = append([]string{}, emission.Report.MissingCriticalActions...)
	}
	return payload
}

func auditRetentionAlertOutboxEvent(emission *AuditRetentionAlertEmission) *OutboxEvent {
	if emission == nil || strings.TrimSpace(emission.IdempotencyKey) == "" {
		return nil
	}
	now := emission.EmittedAt
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	eventID := "obe_audit_alert_" + shortHash(emission.IdempotencyKey)
	emission.OutboxEventID = eventID
	return &OutboxEvent{
		ID:             eventID,
		Topic:          auditRetentionAlertTopic,
		AggregateType:  "audit_retention",
		AggregateID:    "default",
		EventType:      "audit.retention_alerts.emitted",
		IdempotencyKey: emission.IdempotencyKey,
		Payload:        auditRetentionAlertOutboxPayload(emission),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func auditRetentionAlertEmissionAuditPayload(emission *AuditRetentionAlertEmission) map[string]any {
	if emission == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"alert_count":     emission.AlertCount,
		"critical_count":  emission.CriticalCount,
		"warning_count":   emission.WarningCount,
		"status":          emission.Status,
		"topic":           emission.Topic,
		"idempotency_key": emission.IdempotencyKey,
		"outbox_event_id": emission.OutboxEventID,
	}
	if emission.Report != nil {
		payload["cold_archive_due_logs"] = emission.Report.ColdArchiveDueLogs
		payload["expired_logs"] = emission.Report.ExpiredLogs
		payload["hot_days"] = emission.Report.HotDays
		payload["integrity_failures"] = emission.Report.IntegrityFailures
		payload["retention_days"] = emission.Report.RetentionDays
	}
	return payload
}

func normalizeAuditArchiveRequest(req AuditArchiveRequest) AuditArchiveRequest {
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	if req.HotDays <= 0 {
		req.HotDays = defaultAuditHotDays
	}
	if req.Limit <= 0 {
		req.Limit = defaultAuditArchiveLimit
	}
	if req.Limit > maxAuditArchiveLimit {
		req.Limit = maxAuditArchiveLimit
	}
	req.StoragePrefix = strings.TrimRight(strings.TrimSpace(req.StoragePrefix), "/")
	if req.StoragePrefix == "" {
		req.StoragePrefix = defaultAuditArchiveStoragePrefix
	}
	return req
}

func auditArchiveRequestFromLogs(req AuditArchiveRequest, logs []AuditLog) *AuditArchiveRequestResult {
	req = normalizeAuditArchiveRequest(req)
	cutoff := req.Now.AddDate(0, 0, -req.HotDays)
	sort.SliceStable(logs, func(i, j int) bool {
		if !logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].CreatedAt.Before(logs[j].CreatedAt)
		}
		return logs[i].ID < logs[j].ID
	})

	result := &AuditArchiveRequestResult{
		Status:            "skipped",
		Topic:             auditArchiveRequestedTopic,
		StoragePrefix:     req.StoragePrefix,
		HotDays:           req.HotDays,
		ColdArchiveCutoff: cutoff,
		ManifestAlgorithm: auditArchiveManifestAlgorithm,
		ManifestEntries:   []AuditArchiveManifestEntry{},
		RequestedAt:       req.Now,
	}
	for _, log := range logs {
		createdAt := normalizeAuditLogTime(log.CreatedAt)
		if !createdAt.Before(cutoff) {
			continue
		}
		entry := AuditArchiveManifestEntry{
			ID:                 strings.TrimSpace(log.ID),
			CreatedAt:          createdAt,
			Action:             strings.TrimSpace(log.Action),
			TargetType:         strings.TrimSpace(log.TargetType),
			TargetID:           strings.TrimSpace(log.TargetID),
			IntegrityAlgorithm: strings.TrimSpace(log.IntegrityAlgorithm),
			IntegrityHash:      strings.TrimSpace(log.IntegrityHash),
			IntegrityVerified:  log.IntegrityVerified,
		}
		if !entry.IntegrityVerified {
			result.IntegrityFailures++
		}
		if result.OldestCreatedAt.IsZero() || createdAt.Before(result.OldestCreatedAt) {
			result.OldestCreatedAt = createdAt
		}
		if result.NewestCreatedAt.IsZero() || createdAt.After(result.NewestCreatedAt) {
			result.NewestCreatedAt = createdAt
		}
		result.ManifestEntries = append(result.ManifestEntries, entry)
		if len(result.ManifestEntries) >= req.Limit {
			break
		}
	}
	result.LogCount = len(result.ManifestEntries)
	if result.LogCount == 0 {
		result.ArchiveID = "audit_archive_empty_" + req.Now.Format("20060102")
		return result
	}
	manifestHash := auditArchiveManifestHash(result)
	result.ManifestHash = manifestHash
	result.ArchiveID = "audit_archive_" + shortHash(manifestHash)
	result.StorageKey = result.StoragePrefix + "/" + req.Now.Format("2006/01/02") + "/" + result.ArchiveID + ".jsonl"
	result.IdempotencyKey = strings.Join([]string{
		"audit_archive",
		req.Now.Format("2006-01-02"),
		fmt.Sprintf("hot:%d", req.HotDays),
		fmt.Sprintf("limit:%d", req.Limit),
		"manifest:" + shortHash(manifestHash),
	}, ":")
	result.Status = "requested"
	if result.IntegrityFailures > 0 {
		result.Status = "requested_with_integrity_warnings"
	}
	return result
}

func auditArchiveManifestHash(result *AuditArchiveRequestResult) string {
	if result == nil {
		return ""
	}
	payload := map[string]any{
		"manifest_version":    "audit_archive_manifest:v1",
		"hot_days":            result.HotDays,
		"cold_archive_cutoff": result.ColdArchiveCutoff.Format(time.RFC3339Nano),
		"requested_at":        result.RequestedAt.Format(time.RFC3339Nano),
		"log_count":           result.LogCount,
		"integrity_failures":  result.IntegrityFailures,
		"entries":             result.ManifestEntries,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func auditArchiveOutboxEvent(result *AuditArchiveRequestResult) *OutboxEvent {
	if result == nil || strings.TrimSpace(result.IdempotencyKey) == "" {
		return nil
	}
	now := result.RequestedAt
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	eventID := "obe_audit_archive_" + shortHash(result.IdempotencyKey)
	result.OutboxEventID = eventID
	return &OutboxEvent{
		ID:             eventID,
		Topic:          auditArchiveRequestedTopic,
		AggregateType:  "audit_archive",
		AggregateID:    result.ArchiveID,
		EventType:      "audit.archive_requested",
		IdempotencyKey: result.IdempotencyKey,
		Payload:        auditArchiveOutboxPayload(result),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func auditArchiveOutboxPayload(result *AuditArchiveRequestResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	return map[string]any{
		"archive_id":          result.ArchiveID,
		"status":              result.Status,
		"storage_prefix":      result.StoragePrefix,
		"storage_key":         result.StorageKey,
		"hot_days":            result.HotDays,
		"cold_archive_cutoff": result.ColdArchiveCutoff.Format(time.RFC3339Nano),
		"log_count":           result.LogCount,
		"integrity_failures":  result.IntegrityFailures,
		"manifest_algorithm":  result.ManifestAlgorithm,
		"manifest_hash":       result.ManifestHash,
		"manifest_entries":    result.ManifestEntries,
		"requested_at":        result.RequestedAt.Format(time.RFC3339Nano),
		"idempotency_key":     result.IdempotencyKey,
	}
}

func auditArchiveRequestAuditPayload(result *AuditArchiveRequestResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	return map[string]any{
		"archive_id":          result.ArchiveID,
		"status":              result.Status,
		"topic":               result.Topic,
		"storage_prefix":      result.StoragePrefix,
		"storage_key":         result.StorageKey,
		"hot_days":            result.HotDays,
		"cold_archive_cutoff": result.ColdArchiveCutoff.Format(time.RFC3339Nano),
		"log_count":           result.LogCount,
		"integrity_failures":  result.IntegrityFailures,
		"manifest_algorithm":  result.ManifestAlgorithm,
		"manifest_hash":       result.ManifestHash,
		"outbox_event_id":     result.OutboxEventID,
	}
}

func normalizeAuditArchiveCompletionRequest(req AuditArchiveCompletionRequest) (AuditArchiveCompletionRequest, error) {
	req.ArchiveID = strings.TrimSpace(req.ArchiveID)
	req.StorageKey = strings.TrimSpace(req.StorageKey)
	req.ManifestAlgorithm = strings.TrimSpace(req.ManifestAlgorithm)
	if req.ManifestAlgorithm == "" {
		req.ManifestAlgorithm = auditArchiveManifestAlgorithm
	}
	req.ManifestHash = strings.TrimSpace(req.ManifestHash)
	req.ContentHash = strings.TrimSpace(req.ContentHash)
	req.ObjectLockMode = strings.ToUpper(strings.TrimSpace(req.ObjectLockMode))
	req.OutboxEventID = strings.TrimSpace(req.OutboxEventID)
	if req.UploadedAt.IsZero() {
		req.UploadedAt = time.Now().UTC()
	} else {
		req.UploadedAt = req.UploadedAt.UTC()
	}
	if !req.RetainUntil.IsZero() {
		req.RetainUntil = req.RetainUntil.UTC()
	}
	if req.ArchiveID == "" || req.StorageKey == "" || req.ManifestHash == "" || req.ContentHash == "" || req.Bytes <= 0 {
		return req, ErrInvalidArgument
	}
	return req, nil
}

func auditArchiveCompletionFromRequest(req AuditArchiveCompletionRequest, completedAt time.Time) *AuditArchiveCompletion {
	if completedAt.IsZero() {
		completedAt = req.UploadedAt
	}
	return &AuditArchiveCompletion{
		ArchiveID:         req.ArchiveID,
		Status:            "archived",
		StorageKey:        req.StorageKey,
		ManifestAlgorithm: req.ManifestAlgorithm,
		ManifestHash:      req.ManifestHash,
		ContentHash:       req.ContentHash,
		Bytes:             req.Bytes,
		ObjectLockMode:    req.ObjectLockMode,
		RetainUntil:       req.RetainUntil,
		OutboxEventID:     req.OutboxEventID,
		UploadedAt:        req.UploadedAt,
		CompletedAt:       normalizeAuditLogTime(completedAt),
	}
}

func auditArchiveCompletionAuditPayload(completion *AuditArchiveCompletion) map[string]any {
	if completion == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"archive_id":         completion.ArchiveID,
		"status":             completion.Status,
		"storage_key":        completion.StorageKey,
		"manifest_algorithm": completion.ManifestAlgorithm,
		"manifest_hash":      completion.ManifestHash,
		"content_hash":       completion.ContentHash,
		"bytes":              completion.Bytes,
		"object_lock_mode":   completion.ObjectLockMode,
		"outbox_event_id":    completion.OutboxEventID,
		"uploaded_at":        completion.UploadedAt.Format(time.RFC3339Nano),
	}
	if !completion.RetainUntil.IsZero() {
		payload["retain_until"] = completion.RetainUntil.Format(time.RFC3339Nano)
	}
	return payload
}

func normalizeAuditArchiveListRequest(req AuditArchiveListRequest) AuditArchiveListRequest {
	req.ArchiveID = strings.TrimSpace(req.ArchiveID)
	if req.Limit <= 0 {
		req.Limit = defaultAuditArchiveListLimit
	}
	if req.Limit > maxAuditArchiveListLimit {
		req.Limit = maxAuditArchiveListLimit
	}
	if !req.After.IsZero() {
		req.After = req.After.UTC()
	}
	if !req.Before.IsZero() {
		req.Before = req.Before.UTC()
	}
	return req
}

func auditArchiveCompletionsFromLogs(logs []AuditLog) []AuditArchiveCompletion {
	completions := make([]AuditArchiveCompletion, 0, len(logs))
	for _, log := range logs {
		completion, ok := auditArchiveCompletionFromAuditLog(log)
		if ok {
			completions = append(completions, completion)
		}
	}
	return completions
}

func normalizeAuditArchiveVerificationListRequest(req AuditArchiveVerificationListRequest) AuditArchiveVerificationListRequest {
	req.ArchiveID = strings.TrimSpace(req.ArchiveID)
	req.Status = strings.TrimSpace(req.Status)
	if req.Limit <= 0 {
		req.Limit = defaultAuditArchiveVerificationListLimit
	}
	if req.Limit > maxAuditArchiveVerificationListLimit {
		req.Limit = maxAuditArchiveVerificationListLimit
	}
	if !req.After.IsZero() {
		req.After = req.After.UTC()
	}
	if !req.Before.IsZero() {
		req.Before = req.Before.UTC()
	}
	return req
}

func auditArchiveVerificationsFromLogs(req AuditArchiveVerificationListRequest, logs []AuditLog) []AuditArchiveVerification {
	req = normalizeAuditArchiveVerificationListRequest(req)
	verifications := make([]AuditArchiveVerification, 0, len(logs))
	for _, log := range logs {
		verification, ok := auditArchiveVerificationFromAuditLog(log)
		if !ok {
			continue
		}
		if req.Status != "" && verification.Status != req.Status {
			continue
		}
		verifications = append(verifications, verification)
	}
	return verifications
}

func auditArchiveVerificationFromAuditLog(log AuditLog) (AuditArchiveVerification, bool) {
	if log.Action != auditArchiveVerifiedAction || log.TargetType != "audit_archive" {
		return AuditArchiveVerification{}, false
	}
	archiveID := strings.TrimSpace(log.TargetID)
	if archiveID == "" {
		archiveID = auditPayloadString(log.Payload, "archive_id")
	}
	if archiveID == "" {
		return AuditArchiveVerification{}, false
	}
	verifiedAt := auditPayloadTime(log.Payload, "verified_at")
	if verifiedAt.IsZero() {
		verifiedAt = log.CreatedAt
	}
	verification := AuditArchiveVerification{
		ArchiveID:           archiveID,
		Status:              auditPayloadStringWithDefault(log.Payload, "status", "unknown"),
		StorageKey:          auditPayloadString(log.Payload, "storage_key"),
		ManifestAlgorithm:   auditPayloadString(log.Payload, "manifest_algorithm"),
		ManifestHash:        auditPayloadString(log.Payload, "manifest_hash"),
		ExpectedContentHash: auditPayloadString(log.Payload, "expected_content_hash"),
		ActualContentHash:   auditPayloadString(log.Payload, "actual_content_hash"),
		ExpectedBytes:       auditPayloadInt64(log.Payload, "expected_bytes"),
		ActualBytes:         auditPayloadInt64(log.Payload, "actual_bytes"),
		ArchiveIDMatched:    auditPayloadBool(log.Payload, "archive_id_matched"),
		ManifestHashMatched: auditPayloadBool(log.Payload, "manifest_hash_matched"),
		ContentHashMatched:  auditPayloadBool(log.Payload, "content_hash_matched"),
		BytesMatched:        auditPayloadBool(log.Payload, "bytes_matched"),
		LogCountMatched:     auditPayloadBool(log.Payload, "log_count_matched"),
		HeaderLogCount:      auditPayloadInt(log.Payload, "header_log_count"),
		ManifestEntryCount:  auditPayloadInt(log.Payload, "manifest_entry_count"),
		ErrorCode:           auditPayloadString(log.Payload, "error_code"),
		ErrorMessage:        auditPayloadString(log.Payload, "error_message"),
		VerifiedAt:          normalizeAuditLogTime(verifiedAt),
	}
	return verification, true
}

func auditArchiveCompletionFromAuditLog(log AuditLog) (AuditArchiveCompletion, bool) {
	if log.Action != auditArchiveCompletedAction || log.TargetType != "audit_archive" {
		return AuditArchiveCompletion{}, false
	}
	archiveID := strings.TrimSpace(log.TargetID)
	if archiveID == "" {
		archiveID = auditPayloadString(log.Payload, "archive_id")
	}
	if archiveID == "" {
		return AuditArchiveCompletion{}, false
	}
	completion := AuditArchiveCompletion{
		ArchiveID:         archiveID,
		Status:            auditPayloadStringWithDefault(log.Payload, "status", "archived"),
		StorageKey:        auditPayloadString(log.Payload, "storage_key"),
		ManifestAlgorithm: auditPayloadString(log.Payload, "manifest_algorithm"),
		ManifestHash:      auditPayloadString(log.Payload, "manifest_hash"),
		ContentHash:       auditPayloadString(log.Payload, "content_hash"),
		Bytes:             auditPayloadInt64(log.Payload, "bytes"),
		ObjectLockMode:    auditPayloadString(log.Payload, "object_lock_mode"),
		OutboxEventID:     auditPayloadString(log.Payload, "outbox_event_id"),
		UploadedAt:        auditPayloadTime(log.Payload, "uploaded_at"),
		RetainUntil:       auditPayloadTime(log.Payload, "retain_until"),
		CompletedAt:       log.CreatedAt,
	}
	return completion, true
}

func (s *Store) auditArchiveCompletionLocked(req AuditArchiveCompletionRequest) (AuditArchiveCompletion, *AuditLog, bool) {
	for _, log := range s.auditLogs {
		if log == nil || log.Action != auditArchiveCompletedAction || log.TargetType != "audit_archive" || log.TargetID != req.ArchiveID {
			continue
		}
		completion, ok := auditArchiveCompletionFromAuditLog(*log)
		if !ok {
			continue
		}
		if completion.ManifestHash == req.ManifestHash && completion.ContentHash == req.ContentHash {
			return completion, cloneAuditLog(log), true
		}
	}
	return AuditArchiveCompletion{}, nil, false
}

func normalizeAuditArchiveVerifyRequest(req AuditArchiveVerifyRequest) AuditArchiveVerifyRequest {
	req.ArchiveID = strings.TrimSpace(req.ArchiveID)
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	} else {
		req.Now = req.Now.UTC()
	}
	return req
}

type auditArchiveFileHeader struct {
	Type         string `json:"type"`
	ArchiveID    string `json:"archive_id"`
	ManifestHash string `json:"manifest_hash"`
	LogCount     int    `json:"log_count"`
}

func verifyAuditArchiveCompletion(completion AuditArchiveCompletion, storage ObjectStorageConfig, now time.Time) *AuditArchiveVerification {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	verification := &AuditArchiveVerification{
		ArchiveID:           strings.TrimSpace(completion.ArchiveID),
		Status:              "failed",
		StorageKey:          strings.TrimSpace(completion.StorageKey),
		ManifestAlgorithm:   strings.TrimSpace(completion.ManifestAlgorithm),
		ManifestHash:        strings.TrimSpace(completion.ManifestHash),
		ExpectedContentHash: strings.TrimSpace(completion.ContentHash),
		ExpectedBytes:       completion.Bytes,
		VerifiedAt:          now,
	}
	if verification.ArchiveID == "" || verification.StorageKey == "" || verification.ManifestHash == "" || verification.ExpectedContentHash == "" {
		verification.ErrorCode = "invalid_completion"
		verification.ErrorMessage = "archive completion evidence is incomplete"
		return verification
	}
	body, err := storage.downloadAuditArchiveObject(verification.StorageKey, verification.ExpectedBytes)
	if err != nil {
		verification.ErrorCode = "download_failed"
		verification.ErrorMessage = err.Error()
		return verification
	}
	verification.ActualBytes = int64(len(body))
	sum := sha256.Sum256(body)
	verification.ActualContentHash = hex.EncodeToString(sum[:])
	verification.ContentHashMatched = strings.EqualFold(verification.ActualContentHash, verification.ExpectedContentHash)
	verification.BytesMatched = verification.ExpectedBytes <= 0 || verification.ExpectedBytes == verification.ActualBytes

	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		verification.ErrorCode = "empty_archive"
		verification.ErrorMessage = "archive object is empty"
		return verification
	}
	var header auditArchiveFileHeader
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		verification.ErrorCode = "manifest_header_invalid"
		verification.ErrorMessage = err.Error()
		return verification
	}
	verification.ArchiveIDMatched = strings.TrimSpace(header.ArchiveID) == verification.ArchiveID
	verification.ManifestHashMatched = strings.TrimSpace(header.ManifestHash) == verification.ManifestHash
	if len(lines) > 1 {
		verification.ManifestEntryCount = len(lines) - 1
	}
	verification.HeaderLogCount = header.LogCount
	verification.LogCountMatched = header.LogCount == verification.ManifestEntryCount
	if strings.TrimSpace(header.Type) != "audit_archive_manifest" {
		verification.ErrorCode = "manifest_header_type_invalid"
		verification.ErrorMessage = "archive header type is not audit_archive_manifest"
		return verification
	}
	if verification.ArchiveIDMatched && verification.ManifestHashMatched && verification.ContentHashMatched && verification.BytesMatched && verification.LogCountMatched {
		verification.Status = "verified"
		return verification
	}
	verification.ErrorCode = "integrity_mismatch"
	verification.ErrorMessage = "archive object does not match completion evidence"
	return verification
}

func auditArchiveVerificationAuditPayload(verification *AuditArchiveVerification) map[string]any {
	if verification == nil {
		return map[string]any{}
	}
	return map[string]any{
		"archive_id":            verification.ArchiveID,
		"status":                verification.Status,
		"storage_key":           verification.StorageKey,
		"manifest_algorithm":    verification.ManifestAlgorithm,
		"manifest_hash":         verification.ManifestHash,
		"expected_content_hash": verification.ExpectedContentHash,
		"actual_content_hash":   verification.ActualContentHash,
		"expected_bytes":        verification.ExpectedBytes,
		"actual_bytes":          verification.ActualBytes,
		"archive_id_matched":    verification.ArchiveIDMatched,
		"manifest_hash_matched": verification.ManifestHashMatched,
		"content_hash_matched":  verification.ContentHashMatched,
		"bytes_matched":         verification.BytesMatched,
		"log_count_matched":     verification.LogCountMatched,
		"header_log_count":      verification.HeaderLogCount,
		"manifest_entry_count":  verification.ManifestEntryCount,
		"error_code":            verification.ErrorCode,
		"error_message":         verification.ErrorMessage,
		"verified_at":           verification.VerifiedAt.Format(time.RFC3339Nano),
	}
}

func auditPayloadString(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return strings.TrimSpace(typed.String())
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(encoded))
	}
}

func auditPayloadStringWithDefault(payload map[string]any, key string, fallback string) string {
	value := auditPayloadString(payload, key)
	if value == "" {
		return fallback
	}
	return value
}

func auditPayloadInt64(payload map[string]any, key string) int64 {
	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return parsed
		}
		floatParsed, err := typed.Float64()
		if err != nil {
			return 0
		}
		return int64(floatParsed)
	default:
		return 0
	}
}

func auditPayloadInt(payload map[string]any, key string) int {
	return int(auditPayloadInt64(payload, key))
}

func auditPayloadBool(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "y":
			return true
		default:
			return false
		}
	case int:
		return typed != 0
	case int64:
		return typed != 0
	case int32:
		return typed != 0
	case float64:
		return typed != 0
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return parsed != 0
		}
		floatParsed, err := typed.Float64()
		if err != nil {
			return false
		}
		return floatParsed != 0
	default:
		return false
	}
}

func auditPayloadTime(payload map[string]any, key string) time.Time {
	value := auditPayloadString(payload, key)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

var auditPayloadAllowlist = map[string]struct{}{
	"action_filter":           {},
	"alert_count":             {},
	"archive_id":              {},
	"actor_id":                {},
	"actor_type":              {},
	"after":                   {},
	"amount_fen":              {},
	"apply_reason":            {},
	"applied_count":           {},
	"applied_preference_keys": {},
	"applied_scopes":          {},
	"actual_bytes":            {},
	"actual_content_hash":     {},
	"archive_id_matched":      {},
	"attempts":                {},
	"before":                  {},
	"batch_id":                {},
	"bytes_matched":           {},
	"change_request_id":       {},
	"channel":                 {},
	"changed":                 {},
	"claimed":                 {},
	"cleanup_attempts":        {},
	"cold_archive_cutoff":     {},
	"cold_archive_due_logs":   {},
	"compensation_type":       {},
	"bytes":                   {},
	"content_hash":            {},
	"content_hash_matched":    {},
	"critical_count":          {},
	"current_scopes":          {},
	"decision":                {},
	"default_refund_strategy": {},
	"delivery_id":             {},
	"delivery_status":         {},
	"destination":             {},
	"disabled_channels":       {},
	"enabled_channels":        {},
	"error_code":              {},
	"error_message":           {},
	"evidence_count":          {},
	"expected_rider_id":       {},
	"expected_bytes":          {},
	"expected_content_hash":   {},
	"expected_status":         {},
	"export_format":           {},
	"expires_at":              {},
	"expired_logs":            {},
	"failed_count":            {},
	"generated_at":            {},
	"header_log_count":        {},
	"hot_days":                {},
	"idempotency_key":         {},
	"integrity_failures":      {},
	"lease_owner":             {},
	"lease_seconds":           {},
	"limit":                   {},
	"log_count":               {},
	"log_count_matched":       {},
	"manifest_algorithm":      {},
	"manifest_entry_count":    {},
	"manifest_hash":           {},
	"manifest_hash_matched":   {},
	"merchant_id":             {},
	"notification_id":         {},
	"notification_type":       {},
	"object_key":              {},
	"object_lock_mode":        {},
	"outbox_event_id":         {},
	"previous_scopes":         {},
	"previous_rider_id":       {},
	"previous_status":         {},
	"preference_key":          {},
	"preference_keys":         {},
	"preference_requests":     {},
	"preferences":             {},
	"policy_version":          {},
	"provider":                {},
	"quiet_hours":             {},
	"qualification_id":        {},
	"reason":                  {},
	"refund_id":               {},
	"retention_days":          {},
	"replayed":                {},
	"requested_scopes":        {},
	"requested_count":         {},
	"request_reason":          {},
	"retry_after_seconds":     {},
	"retry_policy":            {},
	"retain_until":            {},
	"reviewed_at":             {},
	"role":                    {},
	"rollback_from_scopes":    {},
	"rollback_to_scopes":      {},
	"rollout":                 {},
	"rollout_max_targets":     {},
	"rollout_mode":            {},
	"rollout_percentage":      {},
	"rollout_target_ids":      {},
	"row_count":               {},
	"retry_at":                {},
	"station_id":              {},
	"scheduled_at":            {},
	"scheduled_count":         {},
	"saved":                   {},
	"skipped_count":           {},
	"skipped_preference_keys": {},
	"status":                  {},
	"storage_key":             {},
	"storage_prefix":          {},
	"target_id":               {},
	"target_role":             {},
	"topic":                   {},
	"type":                    {},
	"uploaded_at":             {},
	"verified_at":             {},
	"warning_count":           {},
}

func sanitizeAuditPayload(input map[string]any) map[string]any {
	output := map[string]any{}
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := auditPayloadAllowlist[key]; !ok {
			continue
		}
		output[key] = sanitizeAuditValue(key, value)
	}
	return output
}

func sanitizeAuditValue(key string, value any) any {
	if auditPayloadKeyLooksSensitive(key) {
		return maskAuditScalar(value)
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case time.Time:
		return normalizeAuditLogTime(typed).Format(time.RFC3339Nano)
	default:
		return cloneAny(value)
	}
}

func auditPayloadKeyLooksSensitive(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, part := range []string{"password", "secret", "token", "authorization", "openid", "session", "credential", "certificate", "phone", "mobile", "email", "id_card", "identity", "file_url", "object_key", "signature", "pay_sign", "nonce"} {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

func maskAuditScalar(value any) string {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return ""
	}
	if len(text) <= 6 {
		return "***"
	}
	return text[:3] + "***" + text[len(text)-2:]
}

const (
	auditIntegrityAlgorithmSHA256     = "sha256:v1"
	auditIntegrityAlgorithmHMACSHA256 = "hmac-sha256:v1"
)

type auditLogIntegrityCanonicalPayload struct {
	ID         string         `json:"id"`
	ActorType  string         `json:"actor_type"`
	ActorID    string         `json:"actor_id"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	RequestID  string         `json:"request_id"`
	IPHash     string         `json:"ip_hash"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  string         `json:"created_at"`
}

func sealAuditLogIntegrity(log *AuditLog, signingSecret string) {
	if log == nil {
		return
	}
	log.Payload = sanitizeAuditPayload(log.Payload)
	if log.CreatedAt.IsZero() {
		log.CreatedAt = normalizeAuditLogTime(time.Now())
	} else {
		log.CreatedAt = normalizeAuditLogTime(log.CreatedAt)
	}
	algorithm := auditIntegrityAlgorithmSHA256
	if strings.TrimSpace(signingSecret) != "" {
		algorithm = auditIntegrityAlgorithmHMACSHA256
	}
	hash, ok := computeAuditLogIntegrityHash(*log, algorithm, signingSecret)
	if !ok {
		log.IntegrityAlgorithm = ""
		log.IntegrityHash = ""
		log.IntegrityVerified = false
		return
	}
	log.IntegrityAlgorithm = algorithm
	log.IntegrityHash = hash
	log.IntegrityVerified = true
}

func ensureAuditLogIntegrity(log *AuditLog, signingSecret string) {
	if log == nil {
		return
	}
	log.Payload = sanitizeAuditPayload(log.Payload)
	if !log.CreatedAt.IsZero() {
		log.CreatedAt = normalizeAuditLogTime(log.CreatedAt)
	}
	if strings.TrimSpace(log.IntegrityAlgorithm) == "" || strings.TrimSpace(log.IntegrityHash) == "" {
		sealAuditLogIntegrity(log, signingSecret)
		return
	}
	log.IntegrityVerified = verifyAuditLogIntegrity(*log, signingSecret)
}

func verifyAuditLogIntegrity(log AuditLog, signingSecret string) bool {
	algorithm := strings.TrimSpace(log.IntegrityAlgorithm)
	recordedHash := strings.TrimSpace(log.IntegrityHash)
	if algorithm == "" || recordedHash == "" {
		return false
	}
	expectedHash, ok := computeAuditLogIntegrityHash(log, algorithm, signingSecret)
	if !ok {
		return false
	}
	return hmac.Equal([]byte(recordedHash), []byte(expectedHash))
}

func computeAuditLogIntegrityHash(log AuditLog, algorithm string, signingSecret string) (string, bool) {
	canonical := auditLogIntegrityCanonicalPayload{
		ID:         strings.TrimSpace(log.ID),
		ActorType:  strings.TrimSpace(log.ActorType),
		ActorID:    strings.TrimSpace(log.ActorID),
		Action:     strings.TrimSpace(log.Action),
		TargetType: strings.TrimSpace(log.TargetType),
		TargetID:   strings.TrimSpace(log.TargetID),
		RequestID:  strings.TrimSpace(log.RequestID),
		IPHash:     strings.TrimSpace(log.IPHash),
		Payload:    sanitizeAuditPayload(log.Payload),
		CreatedAt:  normalizeAuditLogTime(log.CreatedAt).Format(time.RFC3339Nano),
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", false
	}
	switch algorithm {
	case auditIntegrityAlgorithmSHA256:
		sum := sha256.Sum256(payload)
		return hex.EncodeToString(sum[:]), true
	case auditIntegrityAlgorithmHMACSHA256:
		secret := strings.TrimSpace(signingSecret)
		if secret == "" {
			return "", false
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		return hex.EncodeToString(mac.Sum(nil)), true
	default:
		return "", false
	}
}

func normalizeAuditLogTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func auditLogMatchesRequest(log *AuditLog, req AuditLogsRequest) bool {
	if log == nil {
		return false
	}
	if req.ActorType != "" && log.ActorType != req.ActorType {
		return false
	}
	if req.ActorID != "" && log.ActorID != req.ActorID {
		return false
	}
	if req.Action != "" && log.Action != req.Action {
		return false
	}
	if req.TargetType != "" && log.TargetType != req.TargetType {
		return false
	}
	if req.TargetID != "" && log.TargetID != req.TargetID {
		return false
	}
	if !req.After.IsZero() && log.CreatedAt.Before(req.After) {
		return false
	}
	if !req.Before.IsZero() && !log.CreatedAt.Before(req.Before) {
		return false
	}
	return true
}

func (s *Store) OutboxEvents(req OutboxEventsRequest) ([]OutboxEvent, error) {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = OutboxStatusPending
	}
	if !IsOutboxStatus(status) {
		return nil, ErrInvalidArgument
	}
	topic := strings.TrimSpace(req.Topic)
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	events := make([]OutboxEvent, 0)
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if status == OutboxStatusPending {
			if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
				continue
			}
		} else if event.Status != status {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		if (event.Status == OutboxStatusPending || event.Status == OutboxStatusFailed) && outboxLeaseActive(event, now) {
			continue
		}
		if (event.Status == OutboxStatusPending || event.Status == OutboxStatusFailed) && event.AvailableAt.After(now) {
			continue
		}
		events = append(events, *cloneOutboxEvent(event))
	}
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].AvailableAt.Equal(events[j].AvailableAt) {
			return events[i].AvailableAt.Before(events[j].AvailableAt)
		}
		if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].ID < events[j].ID
	})
	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

func (s *Store) OutboxEventDetail(req OutboxEventDetailRequest) (*OutboxEventDetail, error) {
	req, err := normalizeOutboxEventDetailRequest(req)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	event := cloneOutboxEvent(s.outboxEvents[req.EventID])
	s.mu.Unlock()
	if event == nil {
		return nil, ErrNotFound
	}

	audits, err := s.AuditLogs(AuditLogsRequest{
		TargetType: "outbox_event",
		TargetID:   req.EventID,
		Limit:      req.AuditLimit,
	})
	if err != nil {
		return nil, err
	}
	return buildOutboxEventDetail(event, audits, req.Now, req.AuditLimit), nil
}

func normalizeOutboxEventDetailRequest(req OutboxEventDetailRequest) (OutboxEventDetailRequest, error) {
	req.EventID = strings.TrimSpace(req.EventID)
	if req.EventID == "" {
		return req, ErrInvalidArgument
	}
	req.Now = normalizeOutboxNow(req.Now)
	if req.AuditLimit <= 0 {
		req.AuditLimit = 20
	}
	if req.AuditLimit > 50 {
		req.AuditLimit = 50
	}
	return req, nil
}

func outboxLeaseActive(event *OutboxEvent, now time.Time) bool {
	if event == nil || strings.TrimSpace(event.LeaseOwner) == "" || event.LeaseExpiresAt.IsZero() {
		return false
	}
	return event.LeaseExpiresAt.UTC().After(now.UTC())
}

func buildOutboxEventDetail(event *OutboxEvent, audits []AuditLog, now time.Time, auditLimit int) *OutboxEventDetail {
	eventCopy := cloneOutboxEvent(event)
	if eventCopy == nil {
		return nil
	}
	if eventCopy.Payload == nil {
		eventCopy.Payload = map[string]any{}
	}
	now = normalizeOutboxNow(now)
	availableAt := eventCopy.AvailableAt
	if availableAt.IsZero() {
		availableAt = eventCopy.CreatedAt
	}
	leaseActive := outboxLeaseActive(eventCopy, now)
	retryAvailableInSeconds := int64(0)
	if (eventCopy.Status == OutboxStatusPending || eventCopy.Status == OutboxStatusFailed) && availableAt.After(now) {
		retryAvailableInSeconds = int64(availableAt.Sub(now).Seconds())
	}
	leaseExpiresInSeconds := int64(0)
	if leaseActive {
		leaseExpiresInSeconds = int64(eventCopy.LeaseExpiresAt.Sub(now).Seconds())
		if leaseExpiresInSeconds < 0 {
			leaseExpiresInSeconds = 0
		}
	}
	ready := (eventCopy.Status == OutboxStatusPending || eventCopy.Status == OutboxStatusFailed) && !leaseActive && !availableAt.After(now)
	blocked := (eventCopy.Status == OutboxStatusPending || eventCopy.Status == OutboxStatusFailed) && !ready

	incidentCode, severity := outboxIncidentState(eventCopy, ready, blocked, leaseActive)
	detail := &OutboxEventDetail{
		GeneratedAt:             now,
		Event:                   *eventCopy,
		IncidentCode:            incidentCode,
		IncidentSeverity:        severity,
		Ready:                   ready,
		Blocked:                 blocked,
		LeaseActive:             leaseActive,
		RetryAvailableInSeconds: retryAvailableInSeconds,
		LeaseExpiresInSeconds:   leaseExpiresInSeconds,
		PayloadSummary:          outboxPayloadSummary(eventCopy.Payload),
		RelatedTargets:          outboxRelatedTargets(eventCopy),
		AuditFilters: []OutboxEventAuditFilter{{
			TargetType: "outbox_event",
			TargetID:   eventCopy.ID,
			Limit:      auditLimit,
		}},
		RecentAudits:         audits,
		RecommendedOperation: outboxRecommendedOperation(eventCopy, ready, blocked, leaseActive),
		Checklist:            outboxIncidentChecklist(eventCopy, ready, blocked, leaseActive),
	}
	if eventCopy.AggregateType != "" && eventCopy.AggregateID != "" {
		detail.AuditFilters = append(detail.AuditFilters, OutboxEventAuditFilter{
			TargetType: outboxRelatedTargetType(eventCopy.AggregateType),
			TargetID:   eventCopy.AggregateID,
			Limit:      auditLimit,
		})
	}
	return detail
}

func outboxIncidentState(event *OutboxEvent, ready bool, blocked bool, leaseActive bool) (string, string) {
	if event == nil {
		return "outbox.unknown", "info"
	}
	switch event.Status {
	case OutboxStatusDeadLetter:
		return "outbox.dead_letter", "critical"
	case OutboxStatusPublished:
		return "outbox.published", "info"
	case OutboxStatusFailed:
		if ready {
			return "outbox.failed_ready", "warning"
		}
		return "outbox.retry_backoff", "warning"
	case OutboxStatusPending:
		if leaseActive {
			return "outbox.lease_active", "warning"
		}
		if blocked {
			return "outbox.pending_blocked", "warning"
		}
		return "outbox.ready", "info"
	default:
		return "outbox.unknown", "info"
	}
}

func outboxRecommendedOperation(event *OutboxEvent, ready bool, blocked bool, leaseActive bool) OutboxRecommendedOperation {
	if event == nil {
		return OutboxRecommendedOperation{}
	}
	values := map[string]any{"event_id": event.ID}
	switch {
	case event.Status == OutboxStatusDeadLetter:
		return OutboxRecommendedOperation{
			Key:    "outbox-release-dead-letter",
			Title:  "解封 Outbox 死信",
			Reason: "事件已进入 dead-letter，需核对下游幂等和业务状态后人工解封。",
			Values: values,
		}
	case event.Status == OutboxStatusPublished:
		return OutboxRecommendedOperation{
			Key:    "audit-logs",
			Title:  "查看处置审计",
			Reason: "事件已发布，下一步应核对发布或人工处置审计。",
			Values: map[string]any{"target_type": "outbox_event", "target_id": event.ID, "limit": 50},
		}
	case leaseActive:
		return OutboxRecommendedOperation{
			Key:    "outbox-renew-lease",
			Title:  "续租 Outbox 租约",
			Reason: "事件正在被 worker 或人工会话持有，继续处理前应确认租约持有人仍健康。",
			Values: map[string]any{"event_id": event.ID, "lease_owner": event.LeaseOwner, "lease_seconds": 60},
		}
	case blocked:
		return OutboxRecommendedOperation{
			Key:    "outbox-replay-event",
			Title:  "恢复单个 Outbox",
			Reason: "事件处于 backoff 或暂不可投递，确认故障解除后可提前恢复。",
			Values: values,
		}
	case ready:
		return OutboxRecommendedOperation{
			Key:    "outbox-claim-events",
			Title:  "领取 Outbox 租约",
			Reason: "事件已可投递，可由人工或 worker 领取短租约后处理。",
			Values: map[string]any{"topic": event.Topic, "limit": 1, "lease_owner": "relay-admin", "lease_seconds": 60},
		}
	default:
		return OutboxRecommendedOperation{
			Key:    "outbox-events",
			Title:  "刷新 Outbox 队列",
			Reason: "当前状态不适合直接写操作，先刷新同主题队列确认最新状态。",
			Values: map[string]any{"topic": event.Topic, "status": event.Status, "limit": 20},
		}
	}
}

func outboxIncidentChecklist(event *OutboxEvent, ready bool, blocked bool, leaseActive bool) []string {
	checklist := []string{
		"核对 payload、aggregate 和当前业务状态是否一致",
		"确认下游服务恢复且幂等键不会造成重复副作用",
		"处理后回查 outbox_event 审计和队列状态",
	}
	if event == nil {
		return checklist
	}
	if event.Status == OutboxStatusDeadLetter {
		checklist = append([]string{"先确认 dead-letter 原因和最大尝试次数"}, checklist...)
	}
	if event.Status == OutboxStatusPublished {
		checklist = append([]string{"仅做审计核验，不再重放已发布事件"}, checklist...)
	}
	if leaseActive {
		checklist = append([]string{"联系租约持有人或等待租约过期后再人工接管"}, checklist...)
	}
	if blocked {
		checklist = append([]string{"确认 backoff 到期时间或故障解除证据"}, checklist...)
	}
	if ready {
		checklist = append([]string{"优先领取短租约，避免多人同时处理"}, checklist...)
	}
	return checklist
}

func outboxPayloadSummary(payload map[string]any) []OutboxPayloadField {
	if len(payload) == 0 {
		return []OutboxPayloadField{}
	}
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	limit := 16
	fields := make([]OutboxPayloadField, 0, minInt(len(keys), limit))
	for index, key := range keys {
		if index >= limit {
			fields = append(fields, OutboxPayloadField{Key: "__truncated__", Value: fmt.Sprintf("%d more fields", len(keys)-limit)})
			break
		}
		fields = append(fields, OutboxPayloadField{Key: key, Value: compactOutboxPayloadValue(payload[key])})
	}
	return fields
}

func compactOutboxPayloadValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return truncateRunes(typed, 240)
	case fmt.Stringer:
		return truncateRunes(typed.String(), 240)
	default:
		encoded, err := json.Marshal(typed)
		if err == nil {
			return truncateRunes(string(encoded), 240)
		}
		return truncateRunes(fmt.Sprint(typed), 240)
	}
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func outboxRelatedTargets(event *OutboxEvent) []OutboxRelatedTarget {
	if event == nil {
		return []OutboxRelatedTarget{}
	}
	targets := []OutboxRelatedTarget{}
	seen := map[string]bool{}
	add := func(targetType string, targetID any, source string) {
		targetType = outboxRelatedTargetType(targetType)
		targetIDString := strings.TrimSpace(fmt.Sprint(targetID))
		if targetType == "" || targetIDString == "" || targetIDString == "<nil>" {
			return
		}
		key := targetType + ":" + targetIDString
		if seen[key] {
			return
		}
		seen[key] = true
		targets = append(targets, OutboxRelatedTarget{
			TargetType:   targetType,
			TargetID:     targetIDString,
			Source:       source,
			OperationKey: "audit-logs",
		})
	}
	add(event.AggregateType, event.AggregateID, "aggregate")
	payloadTargets := map[string]string{
		"archive_id":      "audit_archive",
		"merchant_id":     "merchant_account",
		"object_key":      "object_storage",
		"order_id":        "order",
		"outbox_event_id": "outbox_event",
		"refund_id":       "refund",
		"request_id":      "after_sales",
		"rider_id":        "rider",
		"shop_id":         "shop",
		"station_id":      "station",
		"user_id":         "user",
	}
	for key, targetType := range payloadTargets {
		if value, ok := event.Payload[key]; ok {
			add(targetType, value, "payload."+key)
		}
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].TargetType != targets[j].TargetType {
			return targets[i].TargetType < targets[j].TargetType
		}
		return targets[i].TargetID < targets[j].TargetID
	})
	return targets
}

func outboxRelatedTargetType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "dispatch":
		return "dispatch"
	case "audit_archive":
		return "audit_archive"
	default:
		return value
	}
}

func (s *Store) OutboxStats(req OutboxStatsRequest) (*OutboxStats, error) {
	topic := strings.TrimSpace(req.Topic)
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	events := make([]OutboxEvent, 0, len(s.outboxEvents))
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		events = append(events, *cloneOutboxEvent(event))
	}
	return buildOutboxStats(events, topic, now, req.LeaseExpiringWithinSeconds), nil
}

func outboxTopicAuditTarget(topic string) string {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "all"
	}
	return topic
}

func outboxAuditLogFromRequest(req RecordAuditLogRequest, action string, targetType string, targetID string) (*AuditLog, error) {
	log, err := auditLogFromRequest(req, "")
	if err != nil {
		return nil, err
	}
	if log.Action != action || log.TargetType != targetType || log.TargetID != targetID {
		return nil, ErrInvalidArgument
	}
	return log, nil
}

func outboxEventAuditPayload(event *OutboxEvent) map[string]any {
	if event == nil {
		return map[string]any{}
	}
	return map[string]any{
		"topic":       strings.TrimSpace(event.Topic),
		"status":      strings.TrimSpace(event.Status),
		"attempts":    event.Attempts,
		"lease_owner": strings.TrimSpace(event.LeaseOwner),
	}
}

func outboxClaimAuditPayload(result *ClaimOutboxEventsResult, leaseSeconds int) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	return map[string]any{
		"topic":         outboxTopicAuditTarget(result.Topic),
		"claimed":       result.Claimed,
		"lease_owner":   strings.TrimSpace(result.LeaseOwner),
		"lease_seconds": leaseSeconds,
	}
}

func outboxReplayBatchAuditPayload(result *ReplayOutboxEventsResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	return map[string]any{
		"topic":    outboxTopicAuditTarget(result.Topic),
		"replayed": result.Replayed,
		"limit":    result.Limit,
	}
}

func (s *Store) MarkOutboxEventPublished(req MarkOutboxEventPublishedRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	publishedAt := req.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}
	publishedAt = publishedAt.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, ErrNotFound
	}
	event.Status = OutboxStatusPublished
	event.LastError = ""
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	event.PublishedAt = publishedAt
	event.UpdatedAt = publishedAt
	return cloneOutboxEvent(event), nil
}

func (s *Store) MarkOutboxEventPublishedWithAudit(req MarkOutboxEventPublishedRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.published", "outbox_event", eventID)
	if err != nil {
		return nil, nil, err
	}
	publishedAt := req.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}
	publishedAt = publishedAt.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, nil, ErrNotFound
	}
	event.Status = OutboxStatusPublished
	event.LastError = ""
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	event.PublishedAt = publishedAt
	event.UpdatedAt = publishedAt
	eventCopy := cloneOutboxEvent(event)
	log.Payload = outboxEventAuditPayload(eventCopy)
	auditLog := s.appendAuditLogLocked(log)
	return eventCopy, auditLog, nil
}

func (s *Store) MarkOutboxEventFailed(req MarkOutboxEventFailedRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	retryAfterSeconds := req.RetryAfterSeconds
	if retryAfterSeconds <= 0 {
		retryAfterSeconds = 60
	}
	maxAttempts := req.MaxAttempts

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, ErrNotFound
	}
	event.Attempts++
	event.Status = OutboxStatusFailed
	event.LastError = strings.TrimSpace(req.Error)
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	if maxAttempts > 0 && event.Attempts >= maxAttempts {
		event.Status = OutboxStatusDeadLetter
		event.AvailableAt = now
	} else {
		event.AvailableAt = now.Add(time.Duration(retryAfterSeconds) * time.Second)
	}
	event.UpdatedAt = now
	return cloneOutboxEvent(event), nil
}

func (s *Store) MarkOutboxEventFailedWithAudit(req MarkOutboxEventFailedRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.failed", "outbox_event", eventID)
	if err != nil {
		return nil, nil, err
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	retryAfterSeconds := req.RetryAfterSeconds
	if retryAfterSeconds <= 0 {
		retryAfterSeconds = 60
	}
	maxAttempts := req.MaxAttempts

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, nil, ErrNotFound
	}
	event.Attempts++
	event.Status = OutboxStatusFailed
	event.LastError = strings.TrimSpace(req.Error)
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	if maxAttempts > 0 && event.Attempts >= maxAttempts {
		event.Status = OutboxStatusDeadLetter
		event.AvailableAt = now
	} else {
		event.AvailableAt = now.Add(time.Duration(retryAfterSeconds) * time.Second)
	}
	event.UpdatedAt = now
	eventCopy := cloneOutboxEvent(event)
	log.Payload = outboxEventAuditPayload(eventCopy)
	log.Payload["retry_after_seconds"] = retryAfterSeconds
	auditLog := s.appendAuditLogLocked(log)
	return eventCopy, auditLog, nil
}

func (s *Store) ClaimOutboxEvents(req ClaimOutboxEventsRequest) (*ClaimOutboxEventsResult, error) {
	topic := strings.TrimSpace(req.Topic)
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if leaseOwner == "" {
		leaseOwner = "outbox-relay"
	}
	leaseSeconds := req.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 60
	}
	if leaseSeconds > 3600 {
		leaseSeconds = 3600
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]*OutboxEvent, 0)
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
			continue
		}
		if outboxLeaseActive(event, now) {
			continue
		}
		availableAt := event.AvailableAt
		if availableAt.IsZero() {
			availableAt = event.CreatedAt
		}
		if availableAt.UTC().After(now) {
			continue
		}
		candidates = append(candidates, event)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftAvailableAt := candidates[i].AvailableAt
		if leftAvailableAt.IsZero() {
			leftAvailableAt = candidates[i].CreatedAt
		}
		rightAvailableAt := candidates[j].AvailableAt
		if rightAvailableAt.IsZero() {
			rightAvailableAt = candidates[j].CreatedAt
		}
		if !leftAvailableAt.Equal(rightAvailableAt) {
			return leftAvailableAt.Before(rightAvailableAt)
		}
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := &ClaimOutboxEventsResult{
		Topic:          topic,
		Limit:          limit,
		LeaseOwner:     leaseOwner,
		LeaseExpiresAt: leaseExpiresAt,
		Events:         []OutboxEvent{},
	}
	for _, event := range candidates {
		event.LeaseOwner = leaseOwner
		event.LeaseExpiresAt = leaseExpiresAt
		event.UpdatedAt = now
		result.Events = append(result.Events, *cloneOutboxEvent(event))
	}
	result.Claimed = len(result.Events)
	return result, nil
}

func (s *Store) ClaimOutboxEventsWithAudit(req ClaimOutboxEventsRequest, audit RecordAuditLogRequest) (*ClaimOutboxEventsResult, *AuditLog, error) {
	topic := strings.TrimSpace(req.Topic)
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.claimed", "outbox_topic", outboxTopicAuditTarget(topic))
	if err != nil {
		return nil, nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if leaseOwner == "" {
		leaseOwner = "outbox-relay"
	}
	leaseSeconds := req.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 60
	}
	if leaseSeconds > 3600 {
		leaseSeconds = 3600
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]*OutboxEvent, 0)
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
			continue
		}
		if outboxLeaseActive(event, now) {
			continue
		}
		availableAt := event.AvailableAt
		if availableAt.IsZero() {
			availableAt = event.CreatedAt
		}
		if availableAt.UTC().After(now) {
			continue
		}
		candidates = append(candidates, event)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftAvailableAt := candidates[i].AvailableAt
		if leftAvailableAt.IsZero() {
			leftAvailableAt = candidates[i].CreatedAt
		}
		rightAvailableAt := candidates[j].AvailableAt
		if rightAvailableAt.IsZero() {
			rightAvailableAt = candidates[j].CreatedAt
		}
		if !leftAvailableAt.Equal(rightAvailableAt) {
			return leftAvailableAt.Before(rightAvailableAt)
		}
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := &ClaimOutboxEventsResult{
		Topic:          topic,
		Limit:          limit,
		LeaseOwner:     leaseOwner,
		LeaseExpiresAt: leaseExpiresAt,
		Events:         []OutboxEvent{},
	}
	for _, event := range candidates {
		event.LeaseOwner = leaseOwner
		event.LeaseExpiresAt = leaseExpiresAt
		event.UpdatedAt = now
		result.Events = append(result.Events, *cloneOutboxEvent(event))
	}
	result.Claimed = len(result.Events)
	log.Payload = outboxClaimAuditPayload(result, leaseSeconds)
	auditLog := s.appendAuditLogLocked(log)
	return result, auditLog, nil
}

func (s *Store) RenewOutboxEventLease(req RenewOutboxEventLeaseRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if eventID == "" || leaseOwner == "" {
		return nil, ErrInvalidArgument
	}
	leaseSeconds := req.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 60
	}
	if leaseSeconds > 3600 {
		leaseSeconds = 3600
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, ErrNotFound
	}
	if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
		return nil, ErrInvalidOrderState
	}
	if !outboxLeaseActive(event, now) || event.LeaseOwner != leaseOwner {
		return nil, ErrInvalidOrderState
	}
	event.LeaseExpiresAt = now.Add(time.Duration(leaseSeconds) * time.Second)
	event.UpdatedAt = now
	return cloneOutboxEvent(event), nil
}

func (s *Store) RenewOutboxEventLeaseWithAudit(req RenewOutboxEventLeaseRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if eventID == "" || leaseOwner == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.lease_renewed", "outbox_event", eventID)
	if err != nil {
		return nil, nil, err
	}
	leaseSeconds := req.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 60
	}
	if leaseSeconds > 3600 {
		leaseSeconds = 3600
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, nil, ErrNotFound
	}
	if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
		return nil, nil, ErrInvalidOrderState
	}
	if !outboxLeaseActive(event, now) || event.LeaseOwner != leaseOwner {
		return nil, nil, ErrInvalidOrderState
	}
	event.LeaseExpiresAt = now.Add(time.Duration(leaseSeconds) * time.Second)
	event.UpdatedAt = now
	eventCopy := cloneOutboxEvent(event)
	log.Payload = outboxEventAuditPayload(eventCopy)
	log.Payload["lease_seconds"] = leaseSeconds
	auditLog := s.appendAuditLogLocked(log)
	return eventCopy, auditLog, nil
}

func (s *Store) ReplayOutboxEvent(req ReplayOutboxEventRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, ErrNotFound
	}
	if event.Status == OutboxStatusPublished {
		return nil, ErrInvalidOrderState
	}
	if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed && event.Status != OutboxStatusDeadLetter {
		return nil, ErrInvalidArgument
	}
	event.Status = OutboxStatusPending
	event.LastError = ""
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	event.AvailableAt = now
	event.UpdatedAt = now
	return cloneOutboxEvent(event), nil
}

func (s *Store) ReplayOutboxEventWithAudit(req ReplayOutboxEventRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.replayed", "outbox_event", eventID)
	if err != nil {
		return nil, nil, err
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	event := s.outboxEvents[eventID]
	if event == nil {
		return nil, nil, ErrNotFound
	}
	if event.Status == OutboxStatusPublished {
		return nil, nil, ErrInvalidOrderState
	}
	if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed && event.Status != OutboxStatusDeadLetter {
		return nil, nil, ErrInvalidArgument
	}
	event.Status = OutboxStatusPending
	event.LastError = ""
	event.LeaseOwner = ""
	event.LeaseExpiresAt = time.Time{}
	event.AvailableAt = now
	event.UpdatedAt = now
	eventCopy := cloneOutboxEvent(event)
	log.Payload = outboxEventAuditPayload(eventCopy)
	auditLog := s.appendAuditLogLocked(log)
	return eventCopy, auditLog, nil
}

func (s *Store) ReplayOutboxEvents(req ReplayOutboxEventsRequest) (*ReplayOutboxEventsResult, error) {
	topic := strings.TrimSpace(req.Topic)
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]*OutboxEvent, 0)
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
			continue
		}
		if outboxLeaseActive(event, now) {
			continue
		}
		availableAt := event.AvailableAt
		if availableAt.IsZero() {
			availableAt = event.CreatedAt
		}
		if !availableAt.UTC().After(now) {
			continue
		}
		candidates = append(candidates, event)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftAvailableAt := candidates[i].AvailableAt
		if leftAvailableAt.IsZero() {
			leftAvailableAt = candidates[i].CreatedAt
		}
		rightAvailableAt := candidates[j].AvailableAt
		if rightAvailableAt.IsZero() {
			rightAvailableAt = candidates[j].CreatedAt
		}
		if !leftAvailableAt.Equal(rightAvailableAt) {
			return leftAvailableAt.Before(rightAvailableAt)
		}
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := &ReplayOutboxEventsResult{
		Topic:  topic,
		Limit:  limit,
		Events: []OutboxEvent{},
	}
	for _, event := range candidates {
		event.Status = OutboxStatusPending
		event.LastError = ""
		event.LeaseOwner = ""
		event.LeaseExpiresAt = time.Time{}
		event.AvailableAt = now
		event.UpdatedAt = now
		result.Events = append(result.Events, *cloneOutboxEvent(event))
	}
	result.Replayed = len(result.Events)
	return result, nil
}

func (s *Store) ReplayOutboxEventsWithAudit(req ReplayOutboxEventsRequest, audit RecordAuditLogRequest) (*ReplayOutboxEventsResult, *AuditLog, error) {
	topic := strings.TrimSpace(req.Topic)
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.batch_replayed", "outbox_topic", outboxTopicAuditTarget(topic))
	if err != nil {
		return nil, nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]*OutboxEvent, 0)
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if topic != "" && event.Topic != topic {
			continue
		}
		if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
			continue
		}
		if outboxLeaseActive(event, now) {
			continue
		}
		availableAt := event.AvailableAt
		if availableAt.IsZero() {
			availableAt = event.CreatedAt
		}
		if !availableAt.UTC().After(now) {
			continue
		}
		candidates = append(candidates, event)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		leftAvailableAt := candidates[i].AvailableAt
		if leftAvailableAt.IsZero() {
			leftAvailableAt = candidates[i].CreatedAt
		}
		rightAvailableAt := candidates[j].AvailableAt
		if rightAvailableAt.IsZero() {
			rightAvailableAt = candidates[j].CreatedAt
		}
		if !leftAvailableAt.Equal(rightAvailableAt) {
			return leftAvailableAt.Before(rightAvailableAt)
		}
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := &ReplayOutboxEventsResult{
		Topic:  topic,
		Limit:  limit,
		Events: []OutboxEvent{},
	}
	for _, event := range candidates {
		event.Status = OutboxStatusPending
		event.LastError = ""
		event.LeaseOwner = ""
		event.LeaseExpiresAt = time.Time{}
		event.AvailableAt = now
		event.UpdatedAt = now
		result.Events = append(result.Events, *cloneOutboxEvent(event))
	}
	result.Replayed = len(result.Events)
	log.Payload = outboxReplayBatchAuditPayload(result)
	auditLog := s.appendAuditLogLocked(log)
	return result, auditLog, nil
}

func (s *Store) appendOrderEventLocked(order *Order, event OrderEvent) {
	if order == nil {
		return
	}
	event.Type = strings.TrimSpace(event.Type)
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.Message = strings.TrimSpace(event.Message)
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}
	order.Events = append(order.Events, event)

	topic := orderEventOutboxTopic(event.Type)
	if topic == "" {
		return
	}
	s.enqueueOutboxEventLocked(
		topic,
		"order",
		order.ID,
		event.Type,
		fmt.Sprintf("order_event:%s:%s:%s", order.ID, event.Type, event.CreatedAt.Format(time.RFC3339Nano)),
		orderEventOutboxPayload(order, event),
		event.CreatedAt,
	)
}

func (s *Store) enqueueOutboxEventLocked(topic, aggregateType, aggregateID, eventType, idempotencyKey string, payload map[string]any, now time.Time) *OutboxEvent {
	topic = strings.TrimSpace(topic)
	aggregateType = strings.TrimSpace(aggregateType)
	aggregateID = strings.TrimSpace(aggregateID)
	eventType = strings.TrimSpace(eventType)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if topic == "" || aggregateType == "" || aggregateID == "" || eventType == "" || idempotencyKey == "" {
		return nil
	}
	if existingID := s.outboxByIdempotency[idempotencyKey]; existingID != "" {
		return s.outboxEvents[existingID]
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	s.nextOutboxEventID++
	event := &OutboxEvent{
		ID:             fmt.Sprintf("obe_%d", s.nextOutboxEventID),
		Topic:          topic,
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		EventType:      eventType,
		IdempotencyKey: idempotencyKey,
		Payload:        cloneMapAny(payload),
		Status:         OutboxStatusPending,
		AvailableAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.outboxEvents[event.ID] = event
	s.outboxByIdempotency[idempotencyKey] = event.ID
	return event
}

func orderEventOutboxTopic(eventType string) string {
	switch strings.TrimSpace(eventType) {
	case "order.payment.success":
		return "order.paid"
	case "order.refund.success":
		return "order.refunded"
	case "order.refund.requested":
		return "payment.refund.requested"
	case "order.after_sales.created", "order.after_sales.approved", "order.after_sales.rejected", "order.after_sales.escalated", "order.after_sales.event_added", "order.after_sales.evidence_uploaded":
		return "order.after_sales"
	case "merchant.accepted", "merchant.ready_for_pickup", "delivery.picked_up", "groupbuy.voucher_redeemed", "order.state.compensated":
		return "order.status_changed"
	case "delivery.completed":
		return "order.completed"
	default:
		return ""
	}
}

func orderEventOutboxPayload(order *Order, event OrderEvent) map[string]any {
	payload := map[string]any{
		"order_id":         order.ID,
		"user_id":          order.UserID,
		"shop_id":          order.ShopID,
		"order_type":       order.Type,
		"status":           order.Status,
		"payment_method":   order.PaymentMethod,
		"rider_id":         order.RiderID,
		"amount_fen":       order.AmountFen,
		"event_type":       event.Type,
		"actor_id":         event.ActorID,
		"message":          event.Message,
		"created_at":       event.CreatedAt.UTC(),
		"order_updated_at": order.UpdatedAt.UTC(),
	}
	if event.AmountFen > 0 && (event.Type == "order.refund.success" || event.Type == "order.refund.requested") {
		payload["amount_fen"] = event.AmountFen
		payload["order_amount_fen"] = order.AmountFen
	}
	if order.AddressID != "" {
		payload["address_id"] = order.AddressID
	}
	return payload
}

type orderStateExpectation struct {
	Status          string
	RiderID         string
	PaymentMethod   string
	Evidence        []string
	OccurredAt      time.Time
	AllowRegression bool
}

func (s *Store) expectedOrderStateLocked(order *Order) orderStateExpectation {
	if order == nil {
		return orderStateExpectation{}
	}
	expectation := orderStateExpectation{Status: order.Status, RiderID: order.RiderID}
	if paid, method, evidence, occurredAt := s.orderPaymentEvidenceLocked(order.ID); paid {
		expectation.Status = statusAfterPayment(order)
		expectation.PaymentMethod = method
		expectation.Evidence = append(expectation.Evidence, evidence...)
		expectation.OccurredAt = latestTime(expectation.OccurredAt, occurredAt)
	}
	for _, voucher := range s.groupbuyVouchers {
		if voucher == nil || voucher.OrderID != order.ID {
			continue
		}
		switch voucher.Status {
		case GroupbuyVoucherRedeemed:
			expectation.Status = StatusCompleted
			expectation.OccurredAt = latestTime(expectation.OccurredAt, voucher.RedeemedAt)
			expectation.Evidence = append(expectation.Evidence, "voucher.redeemed:"+voucher.ID)
		case GroupbuyVoucherStatusIssued:
			if orderStatusRank(expectation.Status) < orderStatusRank(StatusVoucherIssued) {
				expectation.Status = StatusVoucherIssued
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, voucher.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "voucher.issued:"+voucher.ID)
		}
	}
	for _, event := range order.Events {
		switch event.Type {
		case "order.payment.success":
			if orderStatusRank(expectation.Status) < orderStatusRank(statusAfterPayment(order)) {
				expectation.Status = statusAfterPayment(order)
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "merchant.accepted":
			if orderStatusRank(expectation.Status) < orderStatusRank(StatusPreparing) {
				expectation.Status = StatusPreparing
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "merchant.ready_for_pickup":
			if orderStatusRank(expectation.Status) < orderStatusRank(StatusDispatching) {
				expectation.Status = StatusDispatching
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "dispatch.rejected", "dispatch.timeout":
			expectation.Status = StatusDispatching
			expectation.RiderID = ""
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.AllowRegression = true
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "dispatch.grabbed":
			expectation.Status = StatusRiderAssigned
			expectation.RiderID = strings.TrimSpace(event.ActorID)
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "delivery.picked_up":
			expectation.Status = StatusPickedUp
			if riderID := strings.TrimSpace(event.ActorID); riderID != "" {
				expectation.RiderID = riderID
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		case "delivery.completed", "groupbuy.voucher_redeemed":
			expectation.Status = StatusCompleted
			if riderID := strings.TrimSpace(event.ActorID); riderID != "" && event.Type == "delivery.completed" {
				expectation.RiderID = riderID
			}
			expectation.OccurredAt = latestTime(expectation.OccurredAt, event.CreatedAt)
			expectation.Evidence = append(expectation.Evidence, "order_event:"+event.Type)
		}
	}
	if dispatchExpectation := s.latestDispatchStateExpectationLocked(order.ID); dispatchExpectation.Status != "" {
		dispatchIsCurrent := expectation.OccurredAt.IsZero() || !dispatchExpectation.OccurredAt.Before(expectation.OccurredAt)
		dispatchCanRegress := dispatchExpectation.AllowRegression && expectation.Status != StatusCompleted
		if dispatchIsCurrent && (orderStatusRank(dispatchExpectation.Status) >= orderStatusRank(expectation.Status) || dispatchCanRegress) {
			expectation.Status = dispatchExpectation.Status
			expectation.RiderID = dispatchExpectation.RiderID
			expectation.OccurredAt = latestTime(expectation.OccurredAt, dispatchExpectation.OccurredAt)
			expectation.AllowRegression = dispatchExpectation.AllowRegression
		}
		expectation.Evidence = append(expectation.Evidence, dispatchExpectation.Evidence...)
	}
	if order.Status == StatusCompleted && orderStatusRank(order.Status) > orderStatusRank(expectation.Status) {
		expectation.Status = order.Status
		if order.RiderID != "" {
			expectation.RiderID = order.RiderID
		}
		expectation.Evidence = append(expectation.Evidence, "terminal_status_protected:"+order.Status)
	} else if orderStatusRank(order.Status) > orderStatusRank(expectation.Status) && !expectation.AllowRegression {
		expectation.Status = order.Status
		if stateKeepsRider(order.Status) && order.RiderID != "" {
			expectation.RiderID = order.RiderID
		}
		expectation.Evidence = append(expectation.Evidence, "current_state_ahead:"+order.Status)
	}
	if !stateKeepsRider(expectation.Status) {
		expectation.RiderID = ""
	}
	if expectation.Status == StatusRiderAssigned && expectation.RiderID == "" && order.RiderID != "" {
		expectation.RiderID = order.RiderID
	}
	return expectation
}

func (s *Store) orderPaymentEvidenceLocked(orderID string) (bool, string, []string, time.Time) {
	evidence := []string{}
	paymentMethod := ""
	occurredAt := time.Time{}
	for _, transaction := range s.paymentTransactions {
		if transaction == nil || transaction.OrderID != orderID || transaction.Status != "success" {
			continue
		}
		if paymentMethod == "" {
			paymentMethod = transaction.Method
		}
		occurredAt = latestTime(occurredAt, transaction.UpdatedAt)
		occurredAt = latestTime(occurredAt, transaction.CreatedAt)
		evidence = append(evidence, "payment_transaction:"+transaction.ID)
	}
	for _, transaction := range s.walletIdempotency {
		if transaction == nil || transaction.OrderID != orderID || transaction.Type != "payment" || transaction.Status != "success" {
			continue
		}
		if paymentMethod == "" {
			paymentMethod = transaction.PaymentMethod
		}
		occurredAt = latestTime(occurredAt, transaction.CreatedAt)
		evidence = append(evidence, "wallet_transaction:"+transaction.ID)
	}
	return len(evidence) > 0, paymentMethod, evidence, occurredAt
}

func (s *Store) latestDispatchStateExpectationLocked(orderID string) orderStateExpectation {
	var latest *DispatchEvent
	for _, event := range s.dispatchEvents {
		if event == nil || event.OrderID != orderID {
			continue
		}
		if !isOrderStateDispatchEvent(event.Type) {
			continue
		}
		if latest == nil || event.CreatedAt.After(latest.CreatedAt) || (event.CreatedAt.Equal(latest.CreatedAt) && event.ID > latest.ID) {
			latest = event
		}
	}
	if latest == nil {
		return orderStateExpectation{}
	}
	expectation := orderStateExpectation{
		Evidence:   []string{"dispatch_event:" + latest.Type},
		OccurredAt: latest.CreatedAt,
	}
	switch latest.Type {
	case "dispatch.grabbed", "dispatch.auto_assign", "dispatch.manual_assign":
		if riderID := strings.TrimSpace(latest.RiderID); riderID != "" {
			expectation.Status = StatusRiderAssigned
			expectation.RiderID = riderID
		}
	case "dispatch.rejected", "dispatch.timeout", "dispatch.no_candidate":
		expectation.Status = StatusDispatching
		expectation.AllowRegression = true
	}
	return expectation
}

func isOrderStateDispatchEvent(eventType string) bool {
	switch eventType {
	case "dispatch.grabbed", "dispatch.auto_assign", "dispatch.manual_assign", "dispatch.rejected", "dispatch.timeout", "dispatch.no_candidate":
		return true
	default:
		return false
	}
}

func stateKeepsRider(status string) bool {
	switch status {
	case StatusRiderAssigned, StatusPickedUp, StatusDelivering, StatusCompleted:
		return true
	default:
		return false
	}
}

func orderStatusRank(status string) int {
	switch status {
	case StatusPendingPayment:
		return 10
	case StatusMerchantPending:
		return 20
	case StatusPreparing, StatusVoucherIssued:
		return 30
	case StatusDispatching:
		return 40
	case StatusRiderAssigned:
		return 50
	case StatusPickedUp, StatusDelivering:
		return 60
	case StatusCompleted:
		return 70
	case StatusCancelled, StatusRefundPending, StatusRefunded:
		return 80
	default:
		return 0
	}
}

func latestTime(left time.Time, right time.Time) time.Time {
	if left.IsZero() {
		return right.UTC()
	}
	if right.IsZero() {
		return left.UTC()
	}
	if right.After(left) {
		return right.UTC()
	}
	return left.UTC()
}

func (s *Store) RiderMarkOrderPickedUp(orderID string, riderID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	riderID = strings.TrimSpace(riderID)
	if orderID == "" || riderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if order.Status != StatusRiderAssigned || order.RiderID != riderID {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	order.Status = StatusPickedUp
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "delivery.picked_up",
		ActorID:   riderID,
		Message:   "骑手已取货",
		CreatedAt: now,
	})
	return cloneOrder(order), nil
}

func (s *Store) RiderMarkOrderDelivered(orderID string, riderID string) (*Order, error) {
	orderID = strings.TrimSpace(orderID)
	riderID = strings.TrimSpace(riderID)
	if orderID == "" || riderID == "" {
		return nil, ErrInvalidArgument
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := s.orders[orderID]
	if order == nil {
		return nil, ErrNotFound
	}
	if (order.Status != StatusPickedUp && order.Status != StatusDelivering) || order.RiderID != riderID {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	order.Status = StatusCompleted
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "delivery.completed",
		ActorID:   riderID,
		Message:   "骑手已送达，订单完成",
		CreatedAt: now,
	})
	return cloneOrder(order), nil
}

func (s *Store) ConsumeFreeDispatchCancel(riderID string, at time.Time) (bool, string, error) {
	riderID = strings.TrimSpace(riderID)
	if riderID == "" {
		return false, "", ErrInvalidArgument
	}
	day := at.UTC().Format("2006-01-02")

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.freeCancelUsedByDate[riderID] == day {
		return false, day, nil
	}
	s.freeCancelUsedByDate[riderID] = day
	return true, day, nil
}

func (s *Store) getOrCreateWalletLocked(userID string) *WalletAccount {
	account := s.wallets[userID]
	if account == nil {
		account = &WalletAccount{UserID: userID, RiskState: "normal"}
		s.wallets[userID] = account
	}
	return account
}

func (s *Store) getOrCreateDepositLocked(subjectType string, subjectID string) *DepositAccount {
	key := depositKey(subjectType, subjectID)
	deposit := s.deposits[key]
	if deposit == nil {
		deposit = &DepositAccount{
			SubjectType: subjectType,
			SubjectID:   subjectID,
			AmountFen:   requiredDepositAmount(subjectType),
			Status:      DepositStatusUnpaid,
			UpdatedAt:   time.Now().UTC(),
		}
		s.deposits[key] = deposit
	}
	return deposit
}

func (s *Store) depositSubjectExistsLocked(subjectType string, subjectID string) bool {
	switch subjectType {
	case "rider":
		rider := s.riders[subjectID]
		return rider != nil && rider.Type == RiderAccountRider
	case "merchant":
		return s.merchants[subjectID] != nil
	default:
		return false
	}
}

func (s *Store) syncDepositStatusLocked(deposit *DepositAccount) {
	if deposit == nil {
		return
	}
	switch deposit.SubjectType {
	case "rider":
		if rider := s.riders[deposit.SubjectID]; rider != nil {
			rider.DepositStatus = deposit.Status
		}
	case "merchant":
		if merchant := s.merchants[deposit.SubjectID]; merchant != nil {
			merchant.DepositStatus = deposit.Status
		}
	}
}

func (s *Store) latestRiderCompletedOrderTimeLocked(riderID string) time.Time {
	var latest time.Time
	for _, order := range s.orders {
		if order == nil || order.RiderID != riderID || order.Status != StatusCompleted {
			continue
		}
		completedAt := order.UpdatedAt
		if completedAt.IsZero() {
			completedAt = order.CreatedAt
		}
		if completedAt.After(latest) {
			latest = completedAt
		}
	}
	return latest
}

func depositKey(subjectType string, subjectID string) string {
	return strings.TrimSpace(subjectType) + ":" + strings.TrimSpace(subjectID)
}

func requiredDepositAmount(subjectType string) int64 {
	switch strings.TrimSpace(subjectType) {
	case "rider":
		return RiderDepositAmountFen
	case "merchant":
		return MerchantDepositAmountFen
	default:
		return 0
	}
}

func (s *Store) verifyPaymentPasswordLocked(userID string, password string) bool {
	expected := s.paymentPasswordHash[userID]
	return expected != "" && expected == hashPaymentPassword(strings.TrimSpace(password))
}

func (s *Store) orderBelongsToMerchantLocked(order *Order, merchantID string) bool {
	if order == nil || order.ShopID == "" || merchantID == "" {
		return false
	}
	shop := s.shops[order.ShopID]
	return shop != nil && shop.MerchantID == merchantID
}

func (s *Store) canReviewAfterSalesLocked(order *Order, request *AfterSalesRequest, actorID string, actorRole string, decision string) bool {
	if order == nil || request == nil || strings.TrimSpace(actorID) == "" {
		return false
	}
	switch request.Status {
	case AfterSalesPendingMerchant:
	case AfterSalesAdminReview:
		if actorRole != "admin" {
			return false
		}
	default:
		return false
	}
	switch actorRole {
	case "admin":
		return true
	case "merchant":
		if request.Status != AfterSalesPendingMerchant || decision == AfterSalesDecisionEscalate {
			return s.orderBelongsToMerchantLocked(order, actorID)
		}
		return s.orderBelongsToMerchantLocked(order, actorID)
	default:
		return false
	}
}

func (s *Store) canAccessAfterSalesLocked(order *Order, request *AfterSalesRequest, actorID string, actorRole string) bool {
	actorID = strings.TrimSpace(actorID)
	actorRole = strings.TrimSpace(actorRole)
	if order == nil || request == nil || actorID == "" {
		return false
	}
	switch actorRole {
	case "admin":
		return true
	case "user":
		return request.UserID == actorID && order.UserID == actorID
	case "merchant":
		return s.orderBelongsToMerchantLocked(order, actorID)
	default:
		return false
	}
}

func (s *Store) canAddAfterSalesEventLocked(order *Order, request *AfterSalesRequest, actorID string, actorRole string, action string) bool {
	if !s.canAccessAfterSalesLocked(order, request, actorID, actorRole) {
		return false
	}
	switch strings.TrimSpace(actorRole) {
	case "admin":
		return true
	case "user":
		return action == AfterSalesActionUserSupplement || action == AfterSalesActionCustomerCare
	case "merchant":
		return action == AfterSalesActionMerchantReply
	default:
		return false
	}
}

func (s *Store) merchantOwnsShopLocked(merchantID string, shopID string) bool {
	shop := s.shops[strings.TrimSpace(shopID)]
	return shop != nil && shop.MerchantID == strings.TrimSpace(merchantID)
}

func (s *Store) shopCanAcceptOrdersLocked(shopID string) bool {
	shop := s.shops[strings.TrimSpace(shopID)]
	if shop == nil || shop.Status != ShopStatusActive {
		return false
	}
	profile := s.merchantProfileLocked(shop.MerchantID)
	return profile != nil && profile.CanAcceptOrders
}

func (s *Store) stationScopeLocked(stationManagerID string) (string, bool, error) {
	stationManagerID = strings.TrimSpace(stationManagerID)
	if stationManagerID == "" {
		return "", true, nil
	}
	manager := s.riders[stationManagerID]
	if manager == nil || manager.Type != RiderAccountStationManager || manager.Status != "active" {
		return "", false, ErrNotFound
	}
	return manager.StationID, false, nil
}

func isStationVisibleDispatchStatus(status string) bool {
	switch status {
	case StatusDispatching, StatusRiderAssigned, StatusPickedUp, StatusDelivering:
		return true
	default:
		return false
	}
}

func (s *Store) stationIDForTaskConfigLocked(stationManagerID string) (string, error) {
	stationID, allStations, err := s.stationScopeLocked(stationManagerID)
	if err != nil {
		return "", err
	}
	if allStations || stationID == "" {
		return "", ErrInvalidArgument
	}
	return stationID, nil
}

func (s *Store) stationTaskConfigLocked(stationID string, stationManagerID string) *StationTaskConfig {
	config := s.stationTaskConfigs[stationID]
	if config == nil {
		config = &StationTaskConfig{
			StationID:                    stationID,
			ConfiguredByStationManagerID: strings.TrimSpace(stationManagerID),
			DailyTaskDurationMinutes:     8 * 60,
			DailyFixedOrderCount:         30,
		}
		s.stationTaskConfigs[stationID] = config
	}
	return config
}

func (s *Store) riderOrderCountsLocked(riderID string) (int, int) {
	accepted := 0
	completed := 0
	for _, order := range s.orders {
		if order == nil || order.RiderID != riderID {
			continue
		}
		if order.Status == StatusCancelled || order.Status == StatusRefundPending || order.Status == StatusRefunded {
			continue
		}
		accepted++
		if order.Status == StatusCompleted {
			completed++
		}
	}
	return accepted, completed
}

func averagePositiveAcceptSeconds(performances []RiderPerformance) float64 {
	total := 0.0
	count := 0
	for _, performance := range performances {
		if performance.AverageAcceptSeconds <= 0 {
			continue
		}
		total += performance.AverageAcceptSeconds
		count++
	}
	if count == 0 {
		return 1
	}
	return total / float64(count)
}

func averagePositiveDailyOrders(performances []RiderPerformance) float64 {
	total := 0.0
	count := 0
	for _, performance := range performances {
		if performance.AverageDailyOrders <= 0 {
			continue
		}
		total += performance.AverageDailyOrders
		count++
	}
	if count == 0 {
		return 1
	}
	return total / float64(count)
}

func evaluateRiderPerformanceLevel(performance RiderPerformance, teamAverageAcceptSeconds float64, teamAverageDailyOrders float64) (int, string, int, RiderPerformanceScoreBreakdown) {
	acceptScore := 0.0
	if performance.AverageAcceptSeconds > 0 {
		acceptScore = (teamAverageAcceptSeconds / performance.AverageAcceptSeconds) * 50
	}
	orderScore := 0.0
	if teamAverageDailyOrders > 0 {
		orderScore = (performance.AverageDailyOrders / teamAverageDailyOrders) * 35
	}
	completionScore := performance.CompletionRate * 15
	ratingScore := 0.0
	ratingConfidence := 0.0
	if performance.RiderAverageRating > 0 && performance.RiderReviewCount > 0 {
		ratingConfidence = math.Min(1, float64(performance.RiderReviewCount)/5)
		ratingScore = (performance.RiderAverageRating / 5) * 12 * ratingConfidence
	}
	score := int(math.Round(math.Max(0, acceptScore+orderScore+completionScore+ratingScore)))
	level := RiderLevelC
	if score >= 120 {
		level = RiderLevelS
	} else if score >= 100 {
		level = RiderLevelA
	} else if score >= 80 {
		level = RiderLevelB
	}
	breakdown := RiderPerformanceScoreBreakdown{
		AcceptScore:              roundFloat(acceptScore, 2),
		OrderVolumeScore:         roundFloat(orderScore, 2),
		CompletionScore:          roundFloat(completionScore, 2),
		RatingScore:              roundFloat(ratingScore, 2),
		RatingConfidence:         roundFloat(ratingConfidence, 2),
		TeamAverageAcceptSeconds: roundFloat(teamAverageAcceptSeconds, 2),
		TeamAverageDailyOrders:   roundFloat(teamAverageDailyOrders, 2),
	}
	return score, level, RiderDispatchPriority(level), breakdown
}

func roundFloat(value float64, precision int) float64 {
	if precision <= 0 {
		return math.Round(value)
	}
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}

func riderCanAcceptDispatchLocked(rider *RiderAccount) bool {
	if rider == nil || rider.Type != RiderAccountRider || rider.Status != "active" || !rider.Online || rider.Capacity <= 0 {
		return false
	}
	return rider.DepositStatus == DepositStatusPaid || rider.DepositStatus == DepositStatusWechatExemptApproved
}

func (s *Store) orderStationIDLocked(order *Order) string {
	if order == nil {
		return "station_1"
	}
	if shop := s.shops[strings.TrimSpace(order.ShopID)]; shop != nil && strings.TrimSpace(shop.StationID) != "" {
		return strings.TrimSpace(shop.StationID)
	}
	if order.ShopID != "" {
		for _, area := range s.stationServiceAreas {
			if area == nil || strings.TrimSpace(area.StationID) == "" {
				continue
			}
			for _, shopID := range area.ShopIDs {
				if strings.TrimSpace(shopID) == order.ShopID {
					return strings.TrimSpace(area.StationID)
				}
			}
		}
	}
	if rider := s.riders[strings.TrimSpace(order.RiderID)]; rider != nil && strings.TrimSpace(rider.StationID) != "" {
		return strings.TrimSpace(rider.StationID)
	}
	return "station_1"
}

func riderMatchesOrderStationLocked(rider *RiderAccount, stationID string) bool {
	if rider == nil {
		return false
	}
	riderStationID := strings.TrimSpace(rider.StationID)
	if riderStationID == "" {
		riderStationID = "station_1"
	}
	return riderStationID == strings.TrimSpace(stationID)
}

func (s *Store) nextDispatchIdempotencyKeyLocked(orderID string) string {
	return fmt.Sprintf("dispatch:%s:%d", strings.TrimSpace(orderID), s.nextDispatchEventID+1)
}

func (s *Store) recordDispatchEventLocked(order *Order, decision *DispatchDecision, eventType string, riderID string, actorID string, reason string, now time.Time) {
	if order == nil {
		return
	}
	if decision == nil {
		decision = &DispatchDecision{
			OrderID:        order.ID,
			StationID:      s.orderStationIDLocked(order),
			IdempotencyKey: s.nextDispatchIdempotencyKeyLocked(order.ID),
		}
	}
	if strings.TrimSpace(decision.OrderID) == "" {
		decision.OrderID = order.ID
	}
	if strings.TrimSpace(decision.StationID) == "" {
		decision.StationID = s.orderStationIDLocked(order)
	}
	if strings.TrimSpace(decision.IdempotencyKey) == "" {
		decision.IdempotencyKey = s.nextDispatchIdempotencyKeyLocked(order.ID)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.nextDispatchEventID++
	event := &DispatchEvent{
		ID:                       fmt.Sprintf("dpe_%d", s.nextDispatchEventID),
		OrderID:                  order.ID,
		StationID:                decision.StationID,
		Mode:                     decision.Mode,
		Type:                     strings.TrimSpace(eventType),
		RiderID:                  strings.TrimSpace(riderID),
		ActorID:                  strings.TrimSpace(actorID),
		Reason:                   strings.TrimSpace(reason),
		IdempotencyKey:           decision.IdempotencyKey,
		OnlineCandidateSize:      decision.RemainingOnlineCandidateSize,
		RejectedRiderIDs:         append([]string{}, decision.RejectedRiderIDs...),
		CanDeclineWithoutPenalty: decision.CanDeclineWithoutPenalty,
		CreatedAt:                now.UTC(),
	}
	if event.Type == "" {
		event.Type = "dispatch.event"
	}
	if event.Mode == "" {
		event.Mode = DispatchModeAutoAssign
	}
	s.dispatchEvents[event.ID] = event
	if topic := dispatchEventOutboxTopic(event.Type); topic != "" {
		s.enqueueOutboxEventLocked(
			topic,
			"dispatch",
			event.OrderID,
			event.Type,
			"dispatch_outbox:"+event.IdempotencyKey+":"+event.Type,
			dispatchEventOutboxPayload(event),
			event.CreatedAt,
		)
	}
}

func (s *Store) appendAfterSalesEventLocked(request *AfterSalesRequest, action string, actorID string, actorRole string, message string, visibleToUser bool, attachments []string, at time.Time) *AfterSalesEvent {
	if request == nil {
		return nil
	}
	action = normalizeAfterSalesAction(action)
	actorID = strings.TrimSpace(actorID)
	actorRole = strings.TrimSpace(actorRole)
	message = strings.TrimSpace(message)
	if action == "" || actorID == "" || actorRole == "" || message == "" {
		return nil
	}
	if at.IsZero() {
		at = time.Now().UTC()
	} else {
		at = at.UTC()
	}
	s.nextAfterSalesEventID++
	event := &AfterSalesEvent{
		ID:            fmt.Sprintf("asev_%d", s.nextAfterSalesEventID),
		RequestID:     request.ID,
		OrderID:       request.OrderID,
		ActorID:       actorID,
		ActorRole:     actorRole,
		Action:        action,
		Message:       message,
		Attachments:   sanitizedStringSlice(attachments),
		VisibleToUser: visibleToUser,
		CreatedAt:     at,
	}
	if s.afterSalesEvents == nil {
		s.afterSalesEvents = map[string]*AfterSalesEvent{}
	}
	s.afterSalesEvents[event.ID] = event
	return event
}

func (s *Store) afterSalesRequestViewLocked(request *AfterSalesRequest) *AfterSalesRequest {
	output := cloneAfterSalesRequest(request)
	if output == nil {
		return nil
	}
	order := s.orders[output.OrderID]
	if order == nil {
		return output
	}
	refundedFen := s.refundedAmountForOrderLocked(order.ID)
	remainingFen := order.AmountFen - refundedFen
	if remainingFen < 0 {
		remainingFen = 0
	}
	output.ShopName = strings.TrimSpace(order.ShopName)
	if output.ShopName == "" {
		output.ShopName = strings.TrimSpace(order.ShopID)
	}
	output.OrderStatus = order.Status
	output.OrderItemSummary = afterSalesOrderItemSummary(order.Items)
	output.LatestEventMessage, output.LatestEventAt = s.afterSalesLatestUserEventLocked(output.ID)
	output.OrderAmountFen = order.AmountFen
	output.RefundedAmountFen = refundedFen
	output.RefundableFen = remainingFen
	return output
}

func (s *Store) afterSalesLatestUserEventLocked(requestID string) (string, time.Time) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return "", time.Time{}
	}
	var latest *AfterSalesEvent
	for _, event := range s.afterSalesEvents {
		if event == nil || event.RequestID != requestID || !event.VisibleToUser {
			continue
		}
		if latest == nil || event.CreatedAt.After(latest.CreatedAt) {
			latest = event
		}
	}
	if latest == nil {
		return "", time.Time{}
	}
	return latest.Message, latest.CreatedAt
}

func afterSalesOrderItemSummary(items []OrderItem) string {
	if len(items) == 0 {
		return ""
	}
	first := strings.TrimSpace(items[0].ProductName)
	if first == "" {
		first = "商品"
	}
	if len(items) == 1 {
		quantity := items[0].Quantity
		if quantity <= 0 {
			quantity = 1
		}
		return fmt.Sprintf("%s x %d", first, quantity)
	}
	totalCount := 0
	for _, item := range items {
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		totalCount += quantity
	}
	return fmt.Sprintf("%s等 %d 件", first, totalCount)
}

func dispatchEventOutboxTopic(eventType string) string {
	switch strings.TrimSpace(eventType) {
	case "dispatch.auto_assign", "dispatch.manual_assign", "dispatch.grabbed":
		return "dispatch.assigned"
	case "dispatch.timeout":
		return "dispatch.timeout"
	case "dispatch.rejected", "dispatch.no_candidate":
		return "dispatch.status_changed"
	default:
		return ""
	}
}

func dispatchEventOutboxPayload(event *DispatchEvent) map[string]any {
	if event == nil {
		return map[string]any{}
	}
	return map[string]any{
		"dispatch_event_id":           event.ID,
		"order_id":                    event.OrderID,
		"station_id":                  event.StationID,
		"mode":                        event.Mode,
		"event_type":                  event.Type,
		"rider_id":                    event.RiderID,
		"actor_id":                    event.ActorID,
		"reason":                      event.Reason,
		"idempotency_key":             event.IdempotencyKey,
		"online_candidate_size":       event.OnlineCandidateSize,
		"rejected_rider_ids":          append([]string{}, event.RejectedRiderIDs...),
		"can_decline_without_penalty": event.CanDeclineWithoutPenalty,
		"created_at":                  event.CreatedAt.UTC(),
	}
}

func (s *Store) onlineCandidateCountLocked(order *Order) int {
	if order == nil {
		return 0
	}
	stationID := s.orderStationIDLocked(order)
	rejected := s.dispatchRejectedRiders[order.ID]
	count := 0
	for _, rider := range s.riders {
		if !riderCanAcceptDispatchLocked(rider) {
			continue
		}
		if !riderMatchesOrderStationLocked(rider, stationID) {
			continue
		}
		if rejected != nil && rejected[rider.ID] {
			continue
		}
		count++
	}
	return count
}

func (s *Store) riderCanDeclineWithoutPenaltyLocked(riderID string, now time.Time) (bool, int, int) {
	rider := s.riders[strings.TrimSpace(riderID)]
	if rider == nil || rider.StationID == "" {
		return false, 0, 0
	}
	config := s.stationTaskConfigLocked(rider.StationID, "")
	completedCount := s.riderCompletedOrderCountOnDateLocked(rider.ID, now.UTC())
	fixedCount := config.DailyFixedOrderCount
	return RiderCanDeclineDispatchWithoutPenalty(completedCount, fixedCount), completedCount, fixedCount
}

func (s *Store) riderCompletedOrderCountOnDateLocked(riderID string, now time.Time) int {
	day := now.UTC().Format("2006-01-02")
	count := 0
	for _, order := range s.orders {
		if order == nil || order.RiderID != riderID || order.Status != StatusCompleted {
			continue
		}
		completedAt := order.UpdatedAt
		if completedAt.IsZero() {
			completedAt = order.CreatedAt
		}
		if completedAt.UTC().Format("2006-01-02") == day {
			count++
		}
	}
	return count
}

func rejectedRiderIDs(rejected map[string]bool) []string {
	rejectedIDs := make([]string, 0, len(rejected))
	for riderID := range rejected {
		rejectedIDs = append(rejectedIDs, riderID)
	}
	sort.Strings(rejectedIDs)
	return rejectedIDs
}

func (s *Store) dispatchDecisionLocked(order *Order, mode string, now time.Time) *DispatchDecision {
	rejected := s.dispatchRejectedRiders[order.ID]
	stationID := s.orderStationIDLocked(order)
	s.refreshDispatchPrioritiesForStationLocked(stationID)
	candidates := make([]RiderAccount, 0)
	for _, rider := range s.riders {
		if !riderCanAcceptDispatchLocked(rider) {
			continue
		}
		if !riderMatchesOrderStationLocked(rider, stationID) {
			continue
		}
		if rejected != nil && rejected[rider.ID] {
			continue
		}
		candidates = append(candidates, *cloneRiderAccount(rider))
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].DispatchPriority != candidates[j].DispatchPriority {
			return candidates[i].DispatchPriority > candidates[j].DispatchPriority
		}
		if candidates[i].AverageAcceptSeconds != candidates[j].AverageAcceptSeconds {
			return candidates[i].AverageAcceptSeconds < candidates[j].AverageAcceptSeconds
		}
		return candidates[i].DistanceMeters < candidates[j].DistanceMeters
	})
	candidateID := ""
	canDeclineWithoutPenalty := false
	completedOrderCount := 0
	fixedOrderCount := 0
	if len(candidates) > 0 {
		candidateID = candidates[0].ID
		canDeclineWithoutPenalty, completedOrderCount, fixedOrderCount = s.riderCanDeclineWithoutPenaltyLocked(candidateID, now)
	}
	return &DispatchDecision{
		OrderID:                      order.ID,
		Mode:                         mode,
		StationID:                    stationID,
		CandidateRiderID:             candidateID,
		RejectedRiderIDs:             rejectedRiderIDs(rejected),
		CanDeclineWithoutPenalty:     canDeclineWithoutPenalty,
		DailyCompletedOrderCount:     completedOrderCount,
		DailyFixedOrderCount:         fixedOrderCount,
		IdempotencyKey:               s.nextDispatchIdempotencyKeyLocked(order.ID),
		RemainingOnlineCandidateSize: len(candidates),
	}
}

func (s *Store) assignOrderToRiderLocked(order *Order, riderID string, mode string, actorID string, now time.Time, decision *DispatchDecision) {
	message := "系统已自动派单给在线骑手"
	if mode == DispatchModeManualAssign {
		message = "站长已手动派单给骑手"
	}
	if strings.TrimSpace(actorID) == "" {
		actorID = riderID
	}
	order.Status = StatusRiderAssigned
	order.RiderID = riderID
	order.UpdatedAt = now
	s.appendOrderEventLocked(order, OrderEvent{
		Type:      "dispatch." + mode,
		ActorID:   actorID,
		Message:   message,
		CreatedAt: now,
	})
	s.recordDispatchEventLocked(order, decision, "dispatch."+mode, riderID, actorID, "", now)
}

func statusAfterPayment(order *Order) string {
	if order != nil && order.Type == OrderTypeGroupbuy {
		return StatusVoucherIssued
	}
	if order != nil && order.ShopID != "" && (order.Type == OrderTypeTakeout || order.Type == OrderTypeMedicine) {
		return StatusMerchantPending
	}
	return StatusDispatching
}

func paymentSuccessMessage(order *Order) string {
	if statusAfterPayment(order) == StatusMerchantPending {
		return "支付成功，订单进入商户待接单"
	}
	if statusAfterPayment(order) == StatusVoucherIssued {
		return "支付成功，团购券已发放"
	}
	return "支付成功，订单进入待调度"
}

func (s *Store) issueGroupbuyVouchersLocked(order *Order, now time.Time) {
	if order == nil || order.Type != OrderTypeGroupbuy || len(order.Items) == 0 || len(s.vouchersByOrderID[order.ID]) > 0 {
		return
	}
	for _, item := range order.Items {
		deal := s.groupbuyDeals[item.ProductID]
		if deal == nil || item.Quantity <= 0 {
			continue
		}
		for index := 0; index < item.Quantity; index++ {
			s.nextVoucherID++
			code := "GBV" + shortHash(fmt.Sprintf("%s:%s:%d", order.ID, item.ProductID, s.nextVoucherID))
			voucher := &GroupbuyVoucher{
				ID:          fmt.Sprintf("gbv_%d", s.nextVoucherID),
				VoucherCode: code,
				QRPayload:   "infinitech://groupbuy/voucher/" + code,
				OrderID:     order.ID,
				UserID:      order.UserID,
				ShopID:      order.ShopID,
				DealID:      item.ProductID,
				DealName:    deal.Name,
				Status:      GroupbuyVoucherStatusIssued,
				CreatedAt:   now,
				ExpiresAt:   now.Add(365 * 24 * time.Hour),
			}
			s.groupbuyVouchers[voucher.ID] = voucher
			s.vouchersByCode[code] = voucher
			s.vouchersByOrderID[order.ID] = append(s.vouchersByOrderID[order.ID], voucher.ID)
		}
	}
}

func groupbuyVoucherCodeFromScan(voucherCode string, qrPayload string) string {
	code := strings.TrimSpace(voucherCode)
	if code != "" {
		return code
	}
	payload := strings.TrimSpace(qrPayload)
	if payload == "" {
		return ""
	}
	if strings.Contains(payload, "/") {
		parts := strings.Split(payload, "/")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return payload
}

func normalizeAfterSalesType(value string) string {
	switch strings.TrimSpace(value) {
	case "", AfterSalesRefundOnly:
		return AfterSalesRefundOnly
	case AfterSalesPartialRefund:
		return AfterSalesPartialRefund
	case AfterSalesFoodSafety:
		return AfterSalesFoodSafety
	default:
		return ""
	}
}

func normalizeAfterSalesAction(value string) string {
	switch strings.TrimSpace(value) {
	case AfterSalesActionCreated:
		return AfterSalesActionCreated
	case AfterSalesActionUserSupplement:
		return AfterSalesActionUserSupplement
	case AfterSalesActionMerchantReply:
		return AfterSalesActionMerchantReply
	case AfterSalesActionCustomerCare:
		return AfterSalesActionCustomerCare
	case AfterSalesActionArbitration:
		return AfterSalesActionArbitration
	case AfterSalesActionInternalNote:
		return AfterSalesActionInternalNote
	case AfterSalesActionEvidenceUploaded:
		return AfterSalesActionEvidenceUploaded
	case AfterSalesActionReviewApproved:
		return AfterSalesActionReviewApproved
	case AfterSalesActionReviewRejected:
		return AfterSalesActionReviewRejected
	case AfterSalesActionEscalated:
		return AfterSalesActionEscalated
	default:
		return ""
	}
}

func afterSalesActionEscalates(action string) bool {
	switch strings.TrimSpace(action) {
	case AfterSalesActionCustomerCare, AfterSalesActionArbitration:
		return true
	default:
		return false
	}
}

func (s *Store) refundableRemainingFenLocked(orderID string) int64 {
	order := s.orders[strings.TrimSpace(orderID)]
	if order == nil || order.AmountFen <= 0 {
		return 0
	}
	remaining := order.AmountFen - s.refundedAmountForOrderLocked(order.ID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *Store) refundedAmountForOrderLocked(orderID string) int64 {
	orderID = strings.TrimSpace(orderID)
	var total int64
	for _, refund := range s.refundTransactions {
		if refund == nil || refund.OrderID != orderID || !refundStatusCountsAgainstOrderTotal(refund.Status) {
			continue
		}
		total += refund.AmountFen
	}
	return total
}

func refundStatusCountsAgainstOrderTotal(status string) bool {
	switch strings.TrimSpace(status) {
	case RefundStatusSuccess, RefundStatusPendingOriginal:
		return true
	default:
		return false
	}
}

func refundOrderStatusAfter(currentStatus string, refundedTotalFen int64, orderAmountFen int64, destination string) string {
	if orderAmountFen <= 0 || refundedTotalFen < orderAmountFen {
		return currentStatus
	}
	if strings.TrimSpace(destination) == RefundDestinationOriginalRoute {
		return StatusRefundPending
	}
	return StatusRefunded
}

func sanitizedStringSlice(values []string) []string {
	output := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		output = append(output, value)
	}
	return output
}

func redPacketShares(packet RedPacket, _ time.Time) []RedPacketShare {
	if packet.Quantity <= 0 {
		return nil
	}
	shares := make([]RedPacketShare, 0, packet.Quantity)
	base := packet.TotalAmountFen / int64(packet.Quantity)
	remain := packet.TotalAmountFen % int64(packet.Quantity)
	for index := 0; index < packet.Quantity; index++ {
		amountFen := base
		if remain > 0 {
			amountFen++
			remain--
		}
		shares = append(shares, RedPacketShare{
			UserID:    "",
			AmountFen: amountFen,
			ClaimedAt: time.Time{},
		})
	}
	return shares
}

func redPacketClaimedCount(detail *RedPacketDetail) int {
	if detail == nil {
		return 0
	}
	count := 0
	for _, share := range detail.Shares {
		if share.UserID != "" && !share.ClaimedAt.IsZero() {
			count++
		}
	}
	return count
}

func redPacketClaimedAmount(detail *RedPacketDetail) int64 {
	if detail == nil {
		return 0
	}
	var amount int64
	for _, share := range detail.Shares {
		if share.UserID != "" && !share.ClaimedAt.IsZero() {
			amount += share.AmountFen
		}
	}
	return amount
}

func redPacketRemainingAmount(detail *RedPacketDetail) int64 {
	if detail == nil {
		return 0
	}
	var amount int64
	for _, share := range detail.Shares {
		if share.UserID == "" || share.ClaimedAt.IsZero() {
			amount += share.AmountFen
		}
	}
	return amount
}

const (
	redPacketRiskClaimWindowSeconds = 10 * 60
	redPacketRiskClaimLimit         = 3
	redPacketRiskAmountWindow       = 24 * time.Hour
	redPacketRiskAmountLimitFen     = 20000
)

func redPacketNextClaimAmount(detail *RedPacketDetail) int64 {
	if detail == nil {
		return 0
	}
	for index := range detail.Shares {
		if detail.Shares[index].UserID == "" && detail.Shares[index].ClaimedAt.IsZero() {
			return detail.Shares[index].AmountFen
		}
	}
	return 0
}

func (s *Store) redPacketClaimRiskLocked(detail *RedPacketDetail, userID string, nextAmountFen int64, now time.Time) *RedPacketRiskCheck {
	if detail == nil {
		return nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	risk := &RedPacketRiskCheck{
		State:          RedPacketRiskPassed,
		Reason:         "当前领取环境正常",
		WindowSeconds:  redPacketRiskClaimWindowSeconds,
		ClaimLimit:     redPacketRiskClaimLimit,
		AmountLimitFen: redPacketRiskAmountLimitFen,
		CheckedAt:      now,
	}
	if detail.Packet.SenderID == userID {
		risk.State = RedPacketRiskBlocked
		risk.ReasonCode = RedPacketRiskSender
		risk.Reason = "发红包人不可领取自己的红包"
		return risk
	}
	claimCount, amountFen := s.redPacketClaimRiskStatsLocked(userID, detail.Packet.TargetID, now)
	risk.ClaimCount = claimCount
	risk.AmountFen = amountFen
	if claimCount >= redPacketRiskClaimLimit {
		risk.State = RedPacketRiskBlocked
		risk.ReasonCode = RedPacketRiskFrequencyLimit
		risk.Reason = "10分钟内领取红包次数过多，请稍后再试"
		return risk
	}
	if nextAmountFen > 0 && amountFen+nextAmountFen > redPacketRiskAmountLimitFen {
		risk.State = RedPacketRiskBlocked
		risk.ReasonCode = RedPacketRiskAmountLimit
		risk.Reason = "今日红包领取金额触达风控上限，请明日再试"
		return risk
	}
	return risk
}

func (s *Store) redPacketClaimRiskStatsLocked(userID string, targetID string, now time.Time) (int, int64) {
	claimWindowStart := now.Add(-time.Duration(redPacketRiskClaimWindowSeconds) * time.Second)
	amountWindowStart := now.Add(-redPacketRiskAmountWindow)
	claimCount := 0
	var amountFen int64
	for _, detail := range s.redPackets {
		if detail == nil {
			continue
		}
		for _, share := range detail.Shares {
			if share.UserID != userID || share.ClaimedAt.IsZero() {
				continue
			}
			if !share.ClaimedAt.Before(amountWindowStart) {
				amountFen += share.AmountFen
			}
			if detail.Packet.TargetID == targetID && !share.ClaimedAt.Before(claimWindowStart) {
				claimCount++
			}
		}
	}
	return claimCount, amountFen
}

func (s *Store) autoRefundExpiredRedPacketLocked(detail *RedPacketDetail, now time.Time) error {
	if detail == nil || detail.Packet.Status != RedPacketStatusCreated {
		return nil
	}
	if detail.Packet.ExpiresAt.IsZero() || now.Before(detail.Packet.ExpiresAt) {
		return nil
	}
	return s.refundRedPacketRemainderLocked(detail, now, RedPacketStatusExpired)
}

func (s *Store) refundRedPacketRemainderLocked(detail *RedPacketDetail, now time.Time, status string) error {
	if detail == nil {
		return ErrNotFound
	}
	if detail.Packet.Status == RedPacketStatusRefunded || detail.Packet.Status == RedPacketStatusExpired {
		return nil
	}
	if detail.Packet.Status == RedPacketStatusFinished {
		return ErrInvalidOrderState
	}
	remaining := redPacketRemainingAmount(detail)
	if remaining <= 0 {
		detail.Packet.Status = RedPacketStatusFinished
		detail.Packet.ClaimedAmountFen = redPacketClaimedAmount(detail)
		return nil
	}
	refundKey := redPacketRefundKey(detail.Packet.ID)
	if s.walletIdempotency[refundKey] == nil {
		account := s.getOrCreateWalletLocked(detail.Packet.SenderID)
		if account.Frozen < remaining {
			return ErrInvalidOrderState
		}
		account.Frozen -= remaining
		account.Balance += remaining
		account.Version++
		transaction := s.createWalletTransactionLocked(detail.Packet.SenderID, detail.Packet.ID, "red_packet_refund", remaining, PaymentBalance, refundKey)
		s.walletIdempotency[transaction.IdempotencyKey] = transaction
	}
	detail.Packet.Status = status
	detail.Packet.ClaimedAmountFen = redPacketClaimedAmount(detail)
	detail.Packet.RefundedAmountFen = remaining
	detail.Packet.RefundedAt = now
	return nil
}

func redPacketFreezeKey(packetID string) string {
	return "red_packet:freeze:" + strings.TrimSpace(packetID)
}

func redPacketClaimKey(packetID string, userID string) string {
	return "red_packet:claim:" + strings.TrimSpace(packetID) + ":" + strings.TrimSpace(userID)
}

func redPacketRefundKey(packetID string) string {
	return "red_packet:refund:" + strings.TrimSpace(packetID)
}

func defaultChatThreads() []ChatThread {
	now := time.Now().UTC()
	return []ChatThread{
		{
			ID:          "official",
			Type:        GroupChatOfficial,
			Title:       "悦享e食官方群",
			Subtitle:    "新用户入群默认静默，重要通知站内信保留",
			Icon:        "官",
			Route:       "/pages/messages/merchant-group/index?thread_id=official&type=official",
			UnreadCount: 3,
			Muted:       true,
			UpdatedAt:   now.Add(-20 * time.Minute),
		},
		{
			ID:          "merchant_blue_sea",
			Type:        GroupChatMerchant,
			Title:       "蓝海餐厅商户群",
			Subtitle:    "群内团购券限时领取，今晚 20:00 前可用",
			Icon:        "店",
			Route:       "/pages/messages/merchant-group/index?thread_id=merchant_blue_sea",
			UnreadCount: 1,
			Muted:       false,
			UpdatedAt:   now.Add(-52 * time.Minute),
		},
		{
			ID:        "customer_service",
			Type:      "customer_service",
			Title:     "客服小悦",
			Subtitle:  "售后申请已提交，商家将在 24 小时内处理",
			Icon:      "客",
			Route:     "/pages/customer-service/chat/index",
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		{
			ID:        "rider_zhang",
			Type:      "private_chat",
			Title:     "骑手 张师傅",
			Subtitle:  "餐品已送达，记得给本次配送评价",
			Icon:      "骑",
			Route:     "/pages/messages/merchant-group/index?thread_id=rider_zhang",
			UpdatedAt: now.Add(-26 * time.Hour),
		},
		{
			ID:        "red_packet_helper",
			Type:      "system",
			Title:     "红包助手",
			Subtitle:  "你收到一个商户群红包，可在余额中查看",
			Icon:      "红",
			Route:     "/pages/red-packet/detail/index",
			UpdatedAt: now.Add(-72 * time.Hour),
		},
	}
}

func knownChatThread(threadID string) bool {
	return chatThreadByID(threadID) != nil
}

func chatThreadByID(threadID string) *ChatThread {
	threadID = strings.TrimSpace(threadID)
	for _, thread := range defaultChatThreads() {
		if thread.ID == threadID {
			threadCopy := thread
			return &threadCopy
		}
	}
	return nil
}

func chatThreadMemberKey(threadID string, subjectType string, subjectID string) string {
	return strings.TrimSpace(threadID) + ":" + strings.TrimSpace(subjectType) + ":" + strings.TrimSpace(subjectID)
}

func defaultChatThreadMembers(thread ChatThread) []ChatThreadMember {
	joinedAt := thread.UpdatedAt
	if joinedAt.IsZero() {
		joinedAt = time.Now().UTC()
	}
	member := func(subjectType string, subjectID string, muted bool) ChatThreadMember {
		return ChatThreadMember{
			ThreadID:    thread.ID,
			SubjectType: strings.TrimSpace(subjectType),
			SubjectID:   strings.TrimSpace(subjectID),
			Muted:       muted,
			JoinedAt:    joinedAt,
		}
	}
	switch thread.ID {
	case "official":
		return []ChatThreadMember{
			member("user", "*", true),
			member("system", "official", true),
			member("support_admin", "*", true),
		}
	case "merchant_blue_sea":
		return []ChatThreadMember{
			member("user", "user_1", false),
			member("user", "user_group_xiaolin", false),
			member("user", "user_group_ajie", false),
			member("merchant", "merchant_1", false),
			member("support_admin", "*", false),
			member("system", "*", false),
		}
	case "customer_service":
		return []ChatThreadMember{
			member("user", "user_1", false),
			member("support_admin", "*", false),
			member("system", "*", false),
		}
	case "rider_zhang":
		return []ChatThreadMember{
			member("user", "user_1", false),
			member("rider", "rider_1", false),
			member("support_admin", "*", false),
			member("system", "*", false),
		}
	case "red_packet_helper":
		return []ChatThreadMember{
			member("user", "user_1", false),
			member("system", "*", false),
		}
	default:
		return nil
	}
}

func seedChatThreadMembers() map[string]*ChatThreadMember {
	members := map[string]*ChatThreadMember{}
	for _, thread := range defaultChatThreads() {
		for _, member := range defaultChatThreadMembers(thread) {
			memberCopy := member
			members[chatThreadMemberKey(member.ThreadID, member.SubjectType, member.SubjectID)] = cloneChatThreadMember(&memberCopy)
		}
	}
	return members
}

func chatThreadMembers(threadID string) []ChatThreadMember {
	thread := chatThreadByID(threadID)
	if thread == nil {
		return nil
	}
	return defaultChatThreadMembers(*thread)
}

func chatThreadMemberForSubject(threadID string, subjectType string, subjectID string) *ChatThreadMember {
	subjectType = normalizeChatSubjectType(subjectType, "", subjectID)
	subjectID = strings.TrimSpace(subjectID)
	if subjectType == "" || subjectID == "" {
		return nil
	}
	for _, member := range chatThreadMembers(threadID) {
		if member.SubjectType != subjectType {
			continue
		}
		if member.SubjectID == "*" || member.SubjectID == subjectID {
			memberCopy := member
			return &memberCopy
		}
	}
	return nil
}

func chatThreadAllowsSubject(threadID string, subjectType string, subjectID string) bool {
	return chatThreadMemberForSubject(threadID, subjectType, subjectID) != nil
}

type chatThreadSelfServePolicyData struct {
	Joinable          bool
	Leaveable         bool
	DefaultMuted      bool
	CouponRequirement string
	CouponCode        string
}

func chatThreadSelfServePolicy(threadID string) chatThreadSelfServePolicyData {
	switch strings.TrimSpace(threadID) {
	case "merchant_blue_sea":
		return chatThreadSelfServePolicyData{
			Joinable:          true,
			Leaveable:         true,
			DefaultMuted:      true,
			CouponRequirement: CouponRequirementGroupMembership,
			CouponCode:        "GROUP8",
		}
	default:
		return chatThreadSelfServePolicyData{}
	}
}

func (s *Store) chatThreadMemberExactLocked(threadID string, subjectType string, subjectID string) *ChatThreadMember {
	subjectType = normalizeChatSubjectType(subjectType, "", subjectID)
	subjectID = strings.TrimSpace(subjectID)
	if threadID == "" || subjectType == "" || subjectID == "" {
		return nil
	}
	if member := s.chatThreadMembers[chatThreadMemberKey(threadID, subjectType, subjectID)]; member != nil {
		return cloneChatThreadMember(member)
	}
	return nil
}

func (s *Store) chatThreadMemberLocked(threadID string, subjectType string, subjectID string) *ChatThreadMember {
	subjectType = normalizeChatSubjectType(subjectType, "", subjectID)
	subjectID = strings.TrimSpace(subjectID)
	if threadID == "" || subjectType == "" || subjectID == "" {
		return nil
	}
	if member := s.chatThreadMembers[chatThreadMemberKey(threadID, subjectType, subjectID)]; member != nil {
		return cloneChatThreadMember(member)
	}
	if member := s.chatThreadMembers[chatThreadMemberKey(threadID, subjectType, "*")]; member != nil {
		return cloneChatThreadMember(member)
	}
	return nil
}

func (s *Store) chatThreadMembershipLocked(userID string, threadID string) (*ChatThreadMembership, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(threadID) == "" {
		return nil, ErrInvalidArgument
	}
	if chatThreadByID(threadID) == nil {
		return nil, ErrNotFound
	}
	policy := chatThreadSelfServePolicy(threadID)
	member := s.chatThreadMemberLocked(threadID, "user", userID)
	exact := s.chatThreadMemberExactLocked(threadID, "user", userID)
	memberCount := s.chatThreadOverviewMemberCountLocked(threadID)
	membership := &ChatThreadMembership{
		ThreadID:          threadID,
		Joined:            member != nil,
		CanJoin:           member == nil && policy.Joinable,
		CanLeave:          exact != nil && policy.Leaveable,
		MemberCount:       memberCount,
		Summary:           chatThreadOverviewSummary(threadID, memberCount),
		CouponRequirement: policy.CouponRequirement,
		CouponCode:        policy.CouponCode,
	}
	if member != nil {
		membership.Muted = member.Muted
		membership.JoinedAt = member.JoinedAt
	}
	return membership, nil
}

func (s *Store) chatThreadOverviewMemberCountLocked(threadID string) int {
	thread := chatThreadByID(threadID)
	if thread == nil {
		return 0
	}
	seed := chatThreadOverviewSeed(threadID)
	currentExplicit := 0
	for _, member := range s.chatThreadMembers {
		if member == nil || strings.TrimSpace(member.ThreadID) != strings.TrimSpace(threadID) || strings.TrimSpace(member.SubjectID) == "*" {
			continue
		}
		currentExplicit++
	}
	defaultExplicit := 0
	for _, member := range defaultChatThreadMembers(*thread) {
		if strings.TrimSpace(member.SubjectID) == "*" {
			continue
		}
		defaultExplicit++
	}
	count := seed.MemberCount + currentExplicit - defaultExplicit
	if count < 1 {
		return 1
	}
	return count
}

type chatThreadOverviewSeedData struct {
	MemberCount  int
	Summary      string
	Announcement string
	SettingsText string
}

func chatThreadOverviewSeed(threadID string) chatThreadOverviewSeedData {
	switch strings.TrimSpace(threadID) {
	case "official":
		return chatThreadOverviewSeedData{
			MemberCount:  1286,
			Summary:      "1286 人已加入 · 新用户默认静音",
			Announcement: "重要通知会同步保留在消息中心，常规讨论默认静默。",
			SettingsText: "群设置",
		}
	case "merchant_blue_sea":
		return chatThreadOverviewSeedData{
			MemberCount:  326,
			Summary:      "326 人已加入 · 新用户默认静音",
			Announcement: "群内优惠每日 10:00 更新，重要通知会保留在消息中心。",
			SettingsText: "群设置",
		}
	case "customer_service":
		return chatThreadOverviewSeedData{
			MemberCount:  2,
			Summary:      "2 人会话 · 消息默认提醒",
			Announcement: "如需人工介入，可先补充订单号或售后单号。",
			SettingsText: "会话设置",
		}
	case "rider_zhang":
		return chatThreadOverviewSeedData{
			MemberCount:  2,
			Summary:      "2 人会话 · 配送消息默认提醒",
			Announcement: "涉及配送争议请优先通过订单页发起售后。",
			SettingsText: "会话设置",
		}
	default:
		return chatThreadOverviewSeedData{
			MemberCount:  1,
			Summary:      "消息助手",
			Announcement: "相关权益和通知会同步记录到消息中心。",
			SettingsText: "会话设置",
		}
	}
}

func chatThreadOverviewSummary(threadID string, memberCount int) string {
	switch strings.TrimSpace(threadID) {
	case "official", "merchant_blue_sea":
		return fmt.Sprintf("%d 人已加入 · 新用户默认静音", memberCount)
	case "customer_service":
		return fmt.Sprintf("%d 人会话 · 消息默认提醒", memberCount)
	case "rider_zhang":
		return fmt.Sprintf("%d 人会话 · 配送消息默认提醒", memberCount)
	default:
		return chatThreadOverviewSeed(threadID).Summary
	}
}

func chatThreadMemberRoleLabel(subjectType string, subjectID string, isSelf bool) string {
	if isSelf {
		return "我"
	}
	switch normalizeChatSubjectType(subjectType, "", subjectID) {
	case "merchant":
		return "商户"
	case "system":
		return "系统"
	case "support_admin":
		return "客服"
	case "rider":
		return "骑手"
	default:
		return "群成员"
	}
}

func chatThreadMemberDisplayName(subjectID string) string {
	switch strings.TrimSpace(subjectID) {
	case "user_group_xiaolin":
		return "小林"
	case "user_group_ajie":
		return "阿杰"
	case "official":
		return "悦享e食官方群"
	default:
		return ""
	}
}

func chatThreadMemberAvatarText(subjectType string, subjectID string, displayName string) string {
	switch normalizeChatSubjectType(subjectType, "", subjectID) {
	case "merchant":
		return "店"
	case "system":
		return "官"
	case "support_admin":
		return "服"
	case "rider":
		return "骑"
	}
	displayName = strings.TrimSpace(displayName)
	if displayName != "" {
		return string([]rune(displayName)[:1])
	}
	return "群"
}

func (s *Store) chatThreadMemberDisplayNameLocked(subjectType string, subjectID string) string {
	if name := chatThreadMemberDisplayName(subjectID); name != "" {
		return name
	}
	switch normalizeChatSubjectType(subjectType, "", subjectID) {
	case "merchant":
		if merchant := s.merchants[subjectID]; merchant != nil && strings.TrimSpace(merchant.DisplayName) != "" {
			return strings.TrimSpace(merchant.DisplayName)
		}
		if shop := s.shopForMerchantLocked(subjectID); shop != nil && strings.TrimSpace(shop.Name) != "" {
			return strings.TrimSpace(shop.Name)
		}
	case "user":
		if nickname := s.nicknameForUserLocked(subjectID); nickname != "" && nickname != "悦享用户" {
			return nickname
		}
	case "rider":
		if strings.TrimSpace(subjectID) == "rider_1" {
			return "骑手 张师傅"
		}
	case "system":
		return "系统助手"
	}
	return strings.TrimSpace(subjectID)
}

func (s *Store) shopForMerchantLocked(merchantID string) *Shop {
	for _, shop := range s.shops {
		if shop != nil && strings.TrimSpace(shop.MerchantID) == strings.TrimSpace(merchantID) {
			return cloneShop(shop)
		}
	}
	return nil
}

type shopDetailSeedData struct {
	RatingText         string
	SalesText          string
	DeliveryText       string
	QualificationText  string
	Announcement       string
	BusinessHours      string
	ContactPhone       string
	Address            string
	ActivityTags       []string
	ServiceCommitments []string
	QualificationItems []string
	SupportBulletins   []string
}

func shopDetailSeed(shopID string) shopDetailSeedData {
	switch strings.TrimSpace(shopID) {
	case "shop_1":
		return shopDetailSeedData{
			RatingText:        "4.8",
			SalesText:         "月售 2381",
			DeliveryText:      "约 32 分钟送达",
			QualificationText: "资质已审核",
			Announcement:      "招牌牛肉饭热卖，到店套餐扫码验券。",
			BusinessHours:     "09:30-22:30",
			ContactPhone:      "13800000001",
			Address:           "大学城生活区 2 号门西侧 18 米",
			ActivityTags:      []string{"满 45 减 8", "新人立减 5", "团购可用"},
			ServiceCommitments: []string{
				"后厨明档公示",
				"准时出餐提醒",
				"支持到店自取",
			},
			QualificationItems: []string{
				"营业执照已公示",
				"健康证在有效期内",
				"平台保证金已缴纳",
			},
			SupportBulletins: []string{
				"遇到缺货会先电话确认再处理订单。",
				"团购券支持到店扫码验券，不支持与到店红包叠加。",
			},
		}
	default:
		return shopDetailSeedData{
			RatingText:        "4.7",
			SalesText:         "月售 999",
			DeliveryText:      "约 30 分钟送达",
			QualificationText: "资质已审核",
			BusinessHours:     "10:00-21:30",
			ContactPhone:      "13800000000",
			Address:           "校园生活区",
			ActivityTags:      []string{"平台补贴"},
			ServiceCommitments: []string{
				"在线客服",
			},
			QualificationItems: []string{
				"平台商户认证",
			},
		}
	}
}

func merchantQualificationItemsForShop(profile *MerchantProfile) []string {
	if profile == nil {
		return nil
	}
	items := make([]string, 0, len(profile.Qualifications)+1)
	for _, qualification := range profile.Qualifications {
		if qualification.Status != QualificationStatusApproved {
			continue
		}
		label := merchantQualificationTypeLabel(qualification.Type)
		if label != "" {
			items = append(items, label+"已审核")
		}
	}
	if profile.Account.DepositStatus == DepositStatusPaid {
		items = append(items, "平台保证金已缴纳")
	}
	if len(items) == 0 && profile.CanAcceptOrders {
		items = append(items, "资质已通过平台审核")
	}
	return items
}

func merchantQualificationTypeLabel(value string) string {
	switch strings.TrimSpace(value) {
	case QualificationBusinessLicense:
		return "营业执照"
	case QualificationHealthCertificate:
		return "健康证"
	default:
		return ""
	}
}

func (s *Store) reviewBelongsToShopLocked(review *Review, shopID string) bool {
	if review == nil || strings.TrimSpace(shopID) == "" {
		return false
	}
	if strings.TrimSpace(review.TargetType) == ReviewTargetShop && strings.TrimSpace(review.TargetID) == strings.TrimSpace(shopID) {
		return true
	}
	orderID := strings.TrimSpace(review.OrderID)
	if orderID == "" {
		orderID = strings.TrimSpace(review.TargetID)
	}
	if order := s.orders[orderID]; order != nil && strings.TrimSpace(order.ShopID) == strings.TrimSpace(shopID) {
		return true
	}
	return false
}

func (s *Store) shopReviewEntriesLocked(shopID string) []ShopReviewEntry {
	reviews := append([]ShopReviewEntry{}, defaultShopReviewEntries(shopID)...)
	for _, review := range s.reviews {
		if !s.reviewBelongsToShopLocked(review, shopID) {
			continue
		}
		userName := s.nicknameForUserLocked(review.UserID)
		avatarText := avatarInitial(userName)
		if review.Anonymous {
			userName = "匿名用户"
			avatarText = "匿"
		}
		reviews = append(reviews, ShopReviewEntry{
			ReviewID:       review.ID,
			UserName:       userName,
			AvatarText:     avatarText,
			Rating:         review.Rating,
			StarsText:      reviewStarsText(review.Rating),
			Content:        review.Content,
			ImageURLs:      append([]string{}, review.ImageURLs...),
			ItemHighlights: shopReviewItemHighlights(review.ItemRatings),
			RiderRating:    review.RiderRating,
			RiderStarsText: reviewStarsTextOrEmpty(review.RiderRating),
			Tags:           append([]string{}, review.Tags...),
			ReplyText:      "商家已查收你的评价，欢迎下次再来。",
			CreatedAt:      review.CreatedAt,
			CreatedText:    reviewTimeText(review.CreatedAt),
		})
	}
	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].CreatedAt.After(reviews[j].CreatedAt)
	})
	return reviews
}

func (s *Store) userReviewedOrderLocked(userID string, orderID string) bool {
	userID = strings.TrimSpace(userID)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" {
		return false
	}
	for _, review := range s.reviews {
		if review == nil || review.UserID != userID || review.OrderID != orderID {
			continue
		}
		return true
	}
	return false
}

func buildShopReviewSummary(reviews []ShopReviewEntry) ShopReviewSummary {
	if len(reviews) == 0 {
		return ShopReviewSummary{
			AverageRating: "4.8",
			ReviewCount:   0,
			PositiveRate:  "100% 好评",
		}
	}
	total := 0
	positive := 0
	tagCounts := map[string]int{}
	for _, review := range reviews {
		total += review.Rating
		if review.Rating >= 4 {
			positive++
		}
		for _, tag := range sanitizedStringSlice(review.Tags) {
			tagCounts[tag]++
		}
	}
	highlightTags := make([]string, 0, len(tagCounts))
	type tagEntry struct {
		tag   string
		count int
	}
	entries := make([]tagEntry, 0, len(tagCounts))
	for tag, count := range tagCounts {
		entries = append(entries, tagEntry{tag: tag, count: count})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].tag < entries[j].tag
		}
		return entries[i].count > entries[j].count
	})
	for _, entry := range entries {
		highlightTags = append(highlightTags, entry.tag)
		if len(highlightTags) >= 4 {
			break
		}
	}
	average := float64(total) / float64(len(reviews))
	positiveRate := int(math.Round((float64(positive) / float64(len(reviews))) * 100))
	return ShopReviewSummary{
		AverageRating: fmt.Sprintf("%.1f", average),
		ReviewCount:   len(reviews),
		PositiveRate:  fmt.Sprintf("%d%% 好评", positiveRate),
		HighlightTags: highlightTags,
	}
}

func reviewStarsText(rating int) string {
	if rating < 1 {
		rating = 1
	}
	if rating > 5 {
		rating = 5
	}
	return strings.Repeat("★", rating) + strings.Repeat("☆", 5-rating)
}

func reviewTimeText(createdAt time.Time) string {
	if createdAt.IsZero() {
		return "刚刚"
	}
	return createdAt.Format("01-02 15:04")
}

func reviewStarsTextOrEmpty(rating int) string {
	if rating < 1 || rating > 5 {
		return ""
	}
	return reviewStarsText(rating)
}

func shopReviewItemHighlights(itemRatings []ReviewItemRating) []string {
	if len(itemRatings) == 0 {
		return nil
	}
	highlights := make([]string, 0, len(itemRatings))
	seen := map[string]bool{}
	for _, item := range itemRatings {
		name := strings.TrimSpace(item.ProductName)
		if name == "" {
			name = strings.TrimSpace(item.ProductID)
		}
		rating := item.Rating
		if name == "" || rating < 1 || rating > 5 {
			continue
		}
		label := fmt.Sprintf("%s %d分", name, rating)
		if seen[label] {
			continue
		}
		seen[label] = true
		highlights = append(highlights, label)
		if len(highlights) >= 3 {
			break
		}
	}
	return highlights
}

func defaultShopReviewEntries(shopID string) []ShopReviewEntry {
	now := time.Now().UTC()
	switch strings.TrimSpace(shopID) {
	case "shop_1":
		return []ShopReviewEntry{
			{
				ReviewID:       "shop_1_seed_review_1",
				UserName:       "林同学",
				AvatarText:     "林",
				Rating:         5,
				StarsText:      reviewStarsText(5),
				ItemHighlights: []string{"招牌牛肉饭 5分", "柠檬茶 4分"},
				RiderRating:    5,
				RiderStarsText: reviewStarsText(5),
				Content:        "牛肉饭分量很稳，晚高峰也没有撒汤，出餐比想象中快。",
				Tags:           []string{"出餐快", "包装完整", "分量足"},
				ReplyText:      "谢谢支持，晚饭时段也会尽量保证出餐速度。",
				CreatedAt:      now.Add(-4 * time.Hour),
				CreatedText:    reviewTimeText(now.Add(-4 * time.Hour)),
			},
			{
				ReviewID:       "shop_1_seed_review_2",
				UserName:       "阿杰",
				AvatarText:     "阿",
				Rating:         4,
				StarsText:      reviewStarsText(4),
				ItemHighlights: []string{"双人套餐 4分"},
				RiderRating:    4,
				RiderStarsText: reviewStarsText(4),
				Content:        "柠檬茶清爽，套餐券到店核销也挺顺，适合中午拼单。",
				Tags:           []string{"适合拼单", "饮品不错", "团购方便"},
				ReplyText:      "团购券工作日和周末都能用，欢迎常来。",
				CreatedAt:      now.Add(-28 * time.Hour),
				CreatedText:    reviewTimeText(now.Add(-28 * time.Hour)),
			},
			{
				ReviewID:       "shop_1_seed_review_3",
				UserName:       "周周",
				AvatarText:     "周",
				Rating:         5,
				StarsText:      reviewStarsText(5),
				ItemHighlights: []string{"番茄鸡蛋面 5分"},
				RiderRating:    5,
				RiderStarsText: reviewStarsText(5),
				Content:        "备注少辣有看到，骑手到店后商家也会同步进度，整体很省心。",
				Tags:           []string{"备注有看", "服务细致", "商家沟通顺畅"},
				ReplyText:      "收到，我们会继续保持备注确认流程。",
				CreatedAt:      now.Add(-52 * time.Hour),
				CreatedText:    reviewTimeText(now.Add(-52 * time.Hour)),
			},
		}
	default:
		return nil
	}
}

func (s *Store) chatThreadMemberProfilesLocked(threadID string, viewerUserID string) []ChatThreadMemberProfile {
	profiles := make([]ChatThreadMemberProfile, 0)
	seen := map[string]bool{}
	for _, member := range s.chatThreadMembers {
		if member == nil || strings.TrimSpace(member.ThreadID) != strings.TrimSpace(threadID) || strings.TrimSpace(member.SubjectID) == "*" {
			continue
		}
		key := chatThreadMemberKey(member.ThreadID, member.SubjectType, member.SubjectID)
		if seen[key] {
			continue
		}
		seen[key] = true
		displayName := s.chatThreadMemberDisplayNameLocked(member.SubjectType, member.SubjectID)
		isSelf := strings.TrimSpace(member.SubjectType) == "user" && strings.TrimSpace(member.SubjectID) == strings.TrimSpace(viewerUserID)
		profiles = append(profiles, ChatThreadMemberProfile{
			ThreadID:    member.ThreadID,
			SubjectType: member.SubjectType,
			SubjectID:   member.SubjectID,
			DisplayName: displayName,
			AvatarText:  chatThreadMemberAvatarText(member.SubjectType, member.SubjectID, displayName),
			RoleLabel:   chatThreadMemberRoleLabel(member.SubjectType, member.SubjectID, isSelf),
			IsSelf:      isSelf,
			JoinedAt:    member.JoinedAt,
		})
	}
	sort.SliceStable(profiles, func(i, j int) bool {
		leftPriority := chatThreadMemberSortPriority(profiles[i])
		rightPriority := chatThreadMemberSortPriority(profiles[j])
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if profiles[i].DisplayName != profiles[j].DisplayName {
			return profiles[i].DisplayName < profiles[j].DisplayName
		}
		return profiles[i].SubjectID < profiles[j].SubjectID
	})
	return profiles
}

func chatThreadMemberSortPriority(profile ChatThreadMemberProfile) int {
	if profile.SubjectType == "merchant" {
		return 0
	}
	if profile.IsSelf {
		return 1
	}
	if profile.SubjectType == "system" {
		return 2
	}
	return 3
}

func normalizeChatSubjectType(subjectType string, role string, subjectID string) string {
	subjectType = strings.TrimSpace(subjectType)
	if subjectType != "" {
		return subjectType
	}
	role = strings.TrimSpace(role)
	switch role {
	case "merchant":
		return "merchant"
	case "rider":
		return "rider"
	case "support_admin", "admin", "super_admin":
		return "support_admin"
	case "station_manager":
		return "station_manager"
	case "user":
		return "user"
	}
	return chatSubjectTypeForID(subjectID)
}

func chatSubjectTypeForID(subjectID string) string {
	subjectID = strings.TrimSpace(subjectID)
	switch {
	case subjectID == "system" || subjectID == "official":
		return "system"
	case strings.HasPrefix(subjectID, "merchant_"):
		return "merchant"
	case strings.HasPrefix(subjectID, "rider_"):
		return "rider"
	case strings.HasPrefix(subjectID, "station_manager_"):
		return "station_manager"
	case strings.HasPrefix(subjectID, "support_"):
		return "support_admin"
	default:
		return "user"
	}
}

func (s *Store) latestMessageForThreadLocked(threadID string) *ChatMessage {
	var latest *ChatMessage
	for _, message := range s.chatMessages {
		if message == nil || message.ThreadID != threadID {
			continue
		}
		if latest == nil || message.CreatedAt.After(latest.CreatedAt) {
			latest = message
		}
	}
	return latest
}

func (s *Store) sortedChatMessagesLocked(threadID string) []ChatMessage {
	messages := make([]ChatMessage, 0)
	for _, message := range s.chatMessages {
		if message == nil || message.ThreadID != threadID {
			continue
		}
		messages = append(messages, *cloneChatMessage(message))
	}
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
	return messages
}

func chatMessagesAfterID(messages []ChatMessage, sinceID string) []ChatMessage {
	if sinceID == "" {
		return append([]ChatMessage{}, messages...)
	}
	for index, message := range messages {
		if message.ID == sinceID {
			return append([]ChatMessage{}, messages[index+1:]...)
		}
	}
	return append([]ChatMessage{}, messages...)
}

func chatReadStateKey(userID string, threadID string) string {
	return strings.TrimSpace(userID) + ":" + strings.TrimSpace(threadID)
}

func seedChatReadStates() map[string]*ChatReadState {
	now := time.Now().UTC()
	return map[string]*ChatReadState{
		chatReadStateKey("user_1", "merchant_blue_sea"): {
			UserID:            "user_1",
			ThreadID:          "merchant_blue_sea",
			LastReadMessageID: "msg_seed_merchant_2",
			ReadAt:            now.Add(-45 * time.Minute),
			UnreadCount:       1,
		},
	}
}

func (s *Store) chatReadStateLocked(userID string, threadID string) *ChatReadState {
	if s.chatReadStates == nil {
		return nil
	}
	return s.chatReadStates[chatReadStateKey(userID, threadID)]
}

func (s *Store) markChatThreadReadLocked(req MarkChatThreadReadRequest) *ChatReadState {
	if s.chatReadStates == nil {
		s.chatReadStates = map[string]*ChatReadState{}
	}
	lastMessageID := strings.TrimSpace(req.LastMessageID)
	readAt := time.Now().UTC()
	if lastMessageID == "" {
		if latest := s.latestMessageForThreadLocked(req.ThreadID); latest != nil {
			lastMessageID = latest.ID
			readAt = latest.CreatedAt
		}
	} else {
		found := false
		for _, message := range s.chatMessages {
			if message != nil && message.ThreadID == req.ThreadID && message.ID == lastMessageID {
				readAt = message.CreatedAt
				found = true
				break
			}
		}
		if !found {
			lastMessageID = ""
		}
	}
	state := &ChatReadState{
		UserID:            strings.TrimSpace(req.UserID),
		ThreadID:          strings.TrimSpace(req.ThreadID),
		LastReadMessageID: lastMessageID,
		ReadAt:            readAt,
	}
	s.chatReadStates[chatReadStateKey(state.UserID, state.ThreadID)] = state
	thread := chatThreadByID(state.ThreadID)
	if thread != nil {
		state.UnreadCount = s.unreadChatMessageCountLocked(state.UserID, thread)
	}
	return state
}

func (s *Store) unreadChatMessageCountLocked(userID string, thread *ChatThread) int {
	if thread == nil {
		return 0
	}
	read := s.chatReadStateLocked(userID, thread.ID)
	if read == nil {
		return thread.UnreadCount
	}
	count := 0
	for _, message := range s.chatMessages {
		if message == nil || message.ThreadID != thread.ID || message.SenderID == userID {
			continue
		}
		if read.LastReadMessageID != "" && message.ID == read.LastReadMessageID {
			continue
		}
		if message.CreatedAt.After(read.ReadAt) {
			count++
		}
	}
	return count
}

func chatMessageOutboxPayload(message ChatMessage) map[string]any {
	payload := map[string]any{
		"id":           message.ID,
		"thread_id":    message.ThreadID,
		"sender_id":    message.SenderID,
		"sender":       message.Sender,
		"content":      message.Content,
		"message_type": message.MessageType,
		"created_at":   message.CreatedAt.Format(time.RFC3339Nano),
	}
	if message.RiskState != "" {
		payload["risk_state"] = message.RiskState
	}
	if message.RiskReasonCode != "" {
		payload["risk_reason_code"] = message.RiskReasonCode
	}
	if message.RiskReason != "" {
		payload["risk_reason"] = message.RiskReason
	}
	if !message.RiskCheckedAt.IsZero() {
		payload["risk_checked_at"] = message.RiskCheckedAt.Format(time.RFC3339Nano)
	}
	return payload
}

func messageRiskCheck(content string, checkedAt time.Time) MessageRiskCheck {
	if checkedAt.IsZero() {
		checkedAt = time.Now().UTC()
	} else {
		checkedAt = checkedAt.UTC()
	}
	normalized := strings.ToLower(strings.TrimSpace(content))
	risk := MessageRiskCheck{
		State:     MessageRiskPassed,
		Reason:    "客服消息敏感信息风控通过",
		CheckedAt: checkedAt,
	}
	if normalized == "" {
		return risk
	}
	reasonCode := ""
	reason := ""
	switch {
	case containsAny(normalized, "支付密码", "付款密码", "payment password", "pay password"):
		reasonCode = MessageRiskPaymentPasswordDisclosed
		reason = "消息疑似包含支付密码信息，已为你拦截。客服不会索要支付密码。"
	case containsAny(normalized, "验证码", "校验码", "短信码", "verification code", "sms code"):
		reasonCode = MessageRiskVerificationCodeShared
		reason = "消息疑似包含验证码信息，已为你拦截。客服不会索要验证码。"
	case containsAny(normalized, "银行卡", "银行卡号", "卡号", "bank card", "card number"):
		reasonCode = MessageRiskBankCardShared
		reason = "消息疑似包含银行卡信息，已为你拦截。客服不会索要银行卡信息。"
	}
	if reasonCode == "" {
		return risk
	}
	risk.State = MessageRiskFlagged
	risk.ReasonCode = MessageRiskSensitiveMention
	risk.Reason = "消息提到了支付密码、验证码或银行卡等敏感信息，已记录风控标签。"
	if messageContainsSecretDisclosure(normalized) {
		risk.State = MessageRiskBlocked
		risk.ReasonCode = reasonCode
		risk.Reason = reason
	}
	return risk
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func messageContainsSecretDisclosure(value string) bool {
	if hasDigitRun(value, 4) {
		return true
	}
	return containsAny(
		value,
		"发给你",
		"发给客服",
		"告诉你",
		"告诉客服",
		"提供给你",
		"提供给客服",
		"报给你",
		"报给客服",
		"给你看",
		"给客服看",
		"my code is",
		"my password is",
		"my card number is",
	)
}

func hasDigitRun(value string, minLength int) bool {
	run := 0
	for _, r := range value {
		if r >= '0' && r <= '9' {
			run++
			if run >= minLength {
				return true
			}
			continue
		}
		run = 0
	}
	return false
}

func applyMessageRiskToChatMessage(message *ChatMessage, risk MessageRiskCheck) {
	if message == nil {
		return
	}
	message.RiskState = risk.State
	message.RiskReasonCode = risk.ReasonCode
	message.RiskReason = risk.Reason
	message.RiskCheckedAt = risk.CheckedAt
}

func applyMessageRiskToServiceTicketEvent(event *ServiceTicketEvent, risk MessageRiskCheck) {
	if event == nil {
		return
	}
	event.RiskState = risk.State
	event.RiskReasonCode = risk.ReasonCode
	event.RiskReason = risk.Reason
	event.RiskCheckedAt = risk.CheckedAt
}

func (s *Store) serviceTicketDetailLocked(ticketID string) (*ServiceTicketDetail, error) {
	ticket := s.serviceTickets[ticketID]
	if ticket == nil {
		return nil, ErrNotFound
	}
	s.syncServiceTicketSLAStatusLocked(ticket, time.Now().UTC())
	events := make([]ServiceTicketEvent, 0)
	for _, event := range s.serviceTicketEvents {
		if event != nil && event.TicketID == ticketID {
			events = append(events, *cloneServiceTicketEvent(event))
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return &ServiceTicketDetail{
		Ticket: *cloneServiceTicket(ticket),
		Events: events,
	}, nil
}

func (s *Store) syncServiceTicketSLAStatusLocked(ticket *ServiceTicket, now time.Time) {
	if ticket == nil {
		return
	}
	ticket.SLAStatus = serviceTicketSLAStatus(ticket, now)
}

func serviceTicketSLAStatus(ticket *ServiceTicket, now time.Time) string {
	if ticket == nil {
		return ServiceTicketSLAStatusNormal
	}
	if ticket.Status == ServiceTicketStatusClosed || ticket.Status == ServiceTicketStatusResolved || ticket.Status == ServiceTicketStatusWaitingConfirm || !ticket.ResolvedAt.IsZero() {
		return ServiceTicketSLAStatusCompleted
	}
	if !ticket.EscalatedAt.IsZero() || strings.TrimSpace(ticket.EscalationLevel) != "" || strings.TrimSpace(ticket.SLAStatus) == ServiceTicketSLAStatusEscalated {
		return ServiceTicketSLAStatusEscalated
	}
	dueAt := ticket.ReplyDueAt
	if dueAt.IsZero() {
		return ServiceTicketSLAStatusNormal
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	dueAt = dueAt.UTC()
	if !now.Before(dueAt) {
		return ServiceTicketSLAStatusOverdue
	}
	if !now.Before(dueAt.Add(-10 * time.Minute)) {
		return ServiceTicketSLAStatusDueSoon
	}
	return ServiceTicketSLAStatusNormal
}

func serviceTicketReplySLA(severity string, category string) time.Duration {
	severity = strings.TrimSpace(severity)
	category = strings.TrimSpace(category)
	switch {
	case strings.Contains(severity, "严重") || strings.Contains(category, "支付") || strings.Contains(category, "红包") || strings.Contains(category, "配送"):
		return 10 * time.Minute
	case strings.Contains(severity, "一般"):
		return 30 * time.Minute
	default:
		return 20 * time.Minute
	}
}

func normalizeServiceTicketQualityResult(result string, score int) string {
	result = strings.TrimSpace(result)
	switch result {
	case ServiceTicketQualityPassed, ServiceTicketQualityNeedsCoaching, ServiceTicketQualityCritical:
		return result
	case "":
		switch {
		case score >= 85:
			return ServiceTicketQualityPassed
		case score >= 70:
			return ServiceTicketQualityNeedsCoaching
		default:
			return ServiceTicketQualityCritical
		}
	default:
		return ""
	}
}

func serviceTicketQualityMessage(review *ServiceTicketQualityReview) string {
	if review == nil {
		return "客服工单抽检已完成"
	}
	result := "通过"
	switch review.Result {
	case ServiceTicketQualityNeedsCoaching:
		result = "需辅导"
	case ServiceTicketQualityCritical:
		result = "严重问题"
	}
	if strings.TrimSpace(review.Notes) == "" {
		return fmt.Sprintf("质检结果：%s，评分 %d", result, review.Score)
	}
	return fmt.Sprintf("质检结果：%s，评分 %d；%s", result, review.Score, review.Notes)
}

func boolQueryValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes" || value == "y"
}

func defaultServiceTicketQualityReviews() []ServiceTicketQualityReview {
	now := time.Now().UTC()
	return []ServiceTicketQualityReview{
		{
			ID:               "stq_preview_1",
			TicketID:         "st_preview_quality",
			SupportID:        "support_1",
			SupportName:      "客服小悦",
			ReviewerID:       "quality_1",
			ReviewerName:     "质检主管",
			Score:            92,
			Result:           ServiceTicketQualityPassed,
			Notes:            "补偿方案清晰，用户确认路径完整。",
			TicketTitle:      "商品质量 · 少送小菜",
			TicketCategory:   "商品质量",
			TicketSLAStatus:  ServiceTicketSLAStatusCompleted,
			TicketFollowUp:   5,
			TicketResolvedAt: now.Add(-36 * time.Hour),
			CreatedAt:        now.Add(-12 * time.Hour),
		},
		{
			ID:               "stq_preview_2",
			TicketID:         "st_preview_delivery",
			SupportID:        "support_2",
			SupportName:      "客服阿宁",
			ReviewerID:       "quality_1",
			ReviewerName:     "质检主管",
			Score:            76,
			Result:           ServiceTicketQualityNeedsCoaching,
			Notes:            "首响接近 SLA，需补充主动同步话术。",
			CoachingRequired: true,
			TicketTitle:      "配送问题 · 预计送达未更新",
			TicketCategory:   "配送问题",
			TicketSLAStatus:  ServiceTicketSLAStatusDueSoon,
			CreatedAt:        now.Add(-2 * time.Hour),
		},
	}
}

func defaultServiceTicketPerformanceSummaries() []ServiceTicketPerformanceSummary {
	return []ServiceTicketPerformanceSummary{
		{
			SupportID:             "support_1",
			SupportName:           "客服小悦",
			AssignedTickets:       24,
			ResolvedTickets:       21,
			ClosedTickets:         18,
			EscalatedTickets:      1,
			OverdueTickets:        1,
			AverageFollowUpRating: 4.8,
			QualityReviewCount:    8,
			QualityAverageScore:   91.5,
			CoachingRequiredCount: 0,
			SLAComplianceRate:     0.96,
			RiskLevel:             "stable",
		},
		{
			SupportID:             "support_2",
			SupportName:           "客服阿宁",
			AssignedTickets:       19,
			ResolvedTickets:       15,
			ClosedTickets:         14,
			EscalatedTickets:      3,
			OverdueTickets:        2,
			AverageFollowUpRating: 4.3,
			QualityReviewCount:    6,
			QualityAverageScore:   82,
			CoachingRequiredCount: 2,
			SLAComplianceRate:     0.84,
			RiskLevel:             "watch",
		},
	}
}

func buildServiceTicketPerformanceSummaries(tickets []ServiceTicket, reviews map[string]*ServiceTicketQualityReview, supportID string) []ServiceTicketPerformanceSummary {
	type accumulator struct {
		summary       ServiceTicketPerformanceSummary
		ratingSum     int
		ratingCount   int
		qualitySum    int
		qualityCount  int
		compliantBase int
	}
	bySupport := map[string]*accumulator{}
	ensure := func(id string, name string) *accumulator {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil
		}
		if supportID != "" && id != supportID {
			return nil
		}
		item := bySupport[id]
		if item == nil {
			item = &accumulator{summary: ServiceTicketPerformanceSummary{SupportID: id, SupportName: strings.TrimSpace(name)}}
			if item.summary.SupportName == "" {
				item.summary.SupportName = id
			}
			bySupport[id] = item
		}
		if item.summary.SupportName == item.summary.SupportID && strings.TrimSpace(name) != "" {
			item.summary.SupportName = strings.TrimSpace(name)
		}
		return item
	}
	for _, ticket := range tickets {
		item := ensure(ticket.AssignedSupportID, ticket.AssignedSupportName)
		if item == nil {
			continue
		}
		item.summary.AssignedTickets++
		item.compliantBase++
		if ticket.Status == ServiceTicketStatusWaitingConfirm || ticket.Status == ServiceTicketStatusResolved || ticket.Status == ServiceTicketStatusClosed || !ticket.ResolvedAt.IsZero() {
			item.summary.ResolvedTickets++
		}
		if ticket.Status == ServiceTicketStatusClosed || !ticket.ClosedAt.IsZero() {
			item.summary.ClosedTickets++
		}
		if serviceTicketSLAStatus(&ticket, time.Now().UTC()) == ServiceTicketSLAStatusOverdue {
			item.summary.OverdueTickets++
		}
		if serviceTicketSLAStatus(&ticket, time.Now().UTC()) == ServiceTicketSLAStatusEscalated {
			item.summary.EscalatedTickets++
		}
		if ticket.FollowUpRating > 0 {
			item.ratingSum += ticket.FollowUpRating
			item.ratingCount++
		}
	}
	for _, review := range reviews {
		if review == nil {
			continue
		}
		item := ensure(review.SupportID, review.SupportName)
		if item == nil {
			continue
		}
		item.summary.QualityReviewCount++
		item.qualitySum += review.Score
		item.qualityCount++
		if review.CoachingRequired || review.Result == ServiceTicketQualityNeedsCoaching || review.Result == ServiceTicketQualityCritical {
			item.summary.CoachingRequiredCount++
		}
	}
	summaries := make([]ServiceTicketPerformanceSummary, 0, len(bySupport))
	for _, item := range bySupport {
		if item.ratingCount > 0 {
			item.summary.AverageFollowUpRating = roundFloat(float64(item.ratingSum)/float64(item.ratingCount), 2)
		}
		if item.qualityCount > 0 {
			item.summary.QualityAverageScore = roundFloat(float64(item.qualitySum)/float64(item.qualityCount), 2)
		}
		if item.compliantBase > 0 {
			nonCompliant := item.summary.OverdueTickets + item.summary.EscalatedTickets
			item.summary.SLAComplianceRate = roundFloat(float64(item.compliantBase-nonCompliant)/float64(item.compliantBase), 4)
			if item.summary.SLAComplianceRate < 0 {
				item.summary.SLAComplianceRate = 0
			}
		}
		item.summary.RiskLevel = serviceTicketPerformanceRisk(item.summary)
		summaries = append(summaries, item.summary)
	}
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].RiskLevel != summaries[j].RiskLevel {
			return serviceTicketRiskRank(summaries[i].RiskLevel) > serviceTicketRiskRank(summaries[j].RiskLevel)
		}
		return summaries[i].AssignedTickets > summaries[j].AssignedTickets
	})
	return summaries
}

func filterServiceTicketPerformance(input []ServiceTicketPerformanceSummary, supportID string, limit int) []ServiceTicketPerformanceSummary {
	supportID = strings.TrimSpace(supportID)
	output := make([]ServiceTicketPerformanceSummary, 0)
	for _, item := range input {
		if supportID != "" && item.SupportID != supportID {
			continue
		}
		output = append(output, item)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if len(output) > limit {
		output = output[:limit]
	}
	return output
}

func serviceTicketPerformanceRisk(summary ServiceTicketPerformanceSummary) string {
	switch {
	case summary.CoachingRequiredCount > 0 || summary.QualityAverageScore > 0 && summary.QualityAverageScore < 75 || summary.SLAComplianceRate > 0 && summary.SLAComplianceRate < 0.8:
		return "risk"
	case summary.EscalatedTickets > 0 || summary.OverdueTickets > 0 || summary.QualityAverageScore > 0 && summary.QualityAverageScore < 85:
		return "watch"
	default:
		return "stable"
	}
}

func serviceTicketRiskRank(level string) int {
	switch level {
	case "risk":
		return 3
	case "watch":
		return 2
	default:
		return 1
	}
}

func (s *Store) appendServiceTicketEventLocked(ticketID string, actorID string, actorRole string, title string, message string, status string, attachments []string, createdAt time.Time) *ServiceTicketEvent {
	ticketID = strings.TrimSpace(ticketID)
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)
	status = strings.TrimSpace(status)
	if ticketID == "" || message == "" {
		return nil
	}
	if title == "" {
		title = "处理进度"
	}
	if status == "" {
		status = ServiceTicketEventActive
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	s.nextServiceTicketEventID++
	event := &ServiceTicketEvent{
		ID:          fmt.Sprintf("ste_%d", s.nextServiceTicketEventID),
		TicketID:    ticketID,
		ActorID:     strings.TrimSpace(actorID),
		ActorRole:   strings.TrimSpace(actorRole),
		Title:       title,
		Message:     message,
		Status:      status,
		Attachments: sanitizedStringSlice(attachments),
		CreatedAt:   createdAt.UTC(),
	}
	s.serviceTicketEvents[event.ID] = cloneServiceTicketEvent(event)
	return event
}

func defaultServiceTickets(userID string) []ServiceTicket {
	now := time.Now().UTC()
	return []ServiceTicket{
		{
			ID:                 "st_preview_delivery",
			UserID:             userID,
			Type:               "delivery",
			Category:           "配送问题",
			Title:              "配送问题 · 预计送达未更新",
			Content:            "骑手到店很久了，预计送达时间一直没变化。",
			Contact:            "138****0000",
			RelatedOrderID:     "DD240518001",
			RelatedOrderTitle:  "蓝海餐厅 · 招牌牛肉饭等 3 件",
			RelatedOrderStatus: "配送中",
			Severity:           "较严重",
			Status:             ServiceTicketStatusProcessing,
			SLAStatus:          ServiceTicketSLAStatusDueSoon,
			Solution:           "继续等待：预计 14:35 前送达；延误补偿：订单完成后发放 ¥5 延误券",
			ReplyDueAt:         now.Add(8 * time.Minute),
			CreatedAt:          now.Add(-24 * time.Minute),
			UpdatedAt:          now.Add(-5 * time.Minute),
		},
		{
			ID:                 "st_preview_quality",
			UserID:             userID,
			Type:               "quality",
			Category:           "商品质量",
			Title:              "商品质量 · 少送小菜",
			Content:            "客服已提出补偿方案，待你确认。",
			RelatedOrderID:     "DD240516006",
			RelatedOrderTitle:  "川味小馆 · 双人套餐",
			RelatedOrderStatus: "已完成",
			Severity:           "一般",
			Status:             ServiceTicketStatusWaitingConfirm,
			SLAStatus:          ServiceTicketSLAStatusCompleted,
			Solution:           "客服已提出补偿方案，待你确认",
			CreatedAt:          now.Add(-48 * time.Hour),
			UpdatedAt:          now.Add(-36 * time.Hour),
		},
		{
			ID:        "st_preview_red_packet",
			UserID:    userID,
			Type:      "red_packet",
			Category:  "红包钱包",
			Title:     "红包钱包 · 红包未到账",
			Content:   "已退回余额 ¥9.14",
			Status:    ServiceTicketStatusResolved,
			SLAStatus: ServiceTicketSLAStatusCompleted,
			CreatedAt: now.Add(-96 * time.Hour),
			UpdatedAt: now.Add(-90 * time.Hour),
		},
	}
}

func defaultServiceTicketDetail(ticket ServiceTicket) *ServiceTicketDetail {
	base := ticket.CreatedAt
	if base.IsZero() {
		base = time.Now().UTC().Add(-24 * time.Minute)
	}
	return &ServiceTicketDetail{
		Ticket: ticket,
		Events: []ServiceTicketEvent{
			{ID: ticket.ID + "_e1", TicketID: ticket.ID, Title: "已提交", Message: "问题已同步到客服工单", Status: ServiceTicketEventDone, CreatedAt: base},
			{ID: ticket.ID + "_e2", TicketID: ticket.ID, Title: "客服已受理", Message: "正在核实商家出餐情况", Status: ServiceTicketEventDone, CreatedAt: base.Add(time.Minute)},
			{ID: ticket.ID + "_e3", TicketID: ticket.ID, Title: "商家反馈", Message: "补做菜品，预计 8 分钟后出餐", Status: ServiceTicketEventActive, CreatedAt: base.Add(5 * time.Minute)},
			{ID: ticket.ID + "_e4", TicketID: ticket.ID, Title: "结果确认", Message: "送达后可确认处理结果", Status: ServiceTicketEventPending, CreatedAt: base.Add(6 * time.Minute)},
		},
	}
}

func sanitizeObjectFileName(value string) string {
	value = strings.TrimSpace(filepath.Base(value))
	if value == "." || value == "/" || value == "\\" {
		return ""
	}
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.Map(func(char rune) rune {
		switch {
		case char >= 'a' && char <= 'z':
			return char
		case char >= 'A' && char <= 'Z':
			return char
		case char >= '0' && char <= '9':
			return char
		case char == '.', char == '-', char == '_':
			return char
		default:
			return -1
		}
	}, value)
	if value == "" || strings.HasPrefix(value, ".") {
		return ""
	}
	switch strings.ToLower(filepath.Ext(value)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".heic", ".pdf":
		return value
	default:
		return ""
	}
}

func validAfterSalesEvidenceObjectKey(requestID string, objectKey string) bool {
	requestID = strings.TrimSpace(requestID)
	objectKey = strings.TrimSpace(objectKey)
	if requestID == "" || objectKey == "" || strings.HasPrefix(objectKey, "/") || strings.Contains(objectKey, "..") {
		return false
	}
	if !strings.HasPrefix(objectKey, "after-sales/"+requestID+"/") {
		return false
	}
	return sanitizeObjectFileName(objectKeyFileName(objectKey)) != ""
}

func validPrescriptionImageObjectKey(objectKey string) bool {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.HasPrefix(objectKey, "/") || strings.Contains(objectKey, "..") {
		return false
	}
	if !strings.HasPrefix(objectKey, "prescriptions/") {
		return false
	}
	return sanitizeObjectFileName(objectKeyFileName(objectKey)) != ""
}

func validReviewImageObjectKey(objectKey string) bool {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.HasPrefix(objectKey, "/") || strings.Contains(objectKey, "..") {
		return false
	}
	if !strings.HasPrefix(objectKey, "reviews/") {
		return false
	}
	return sanitizeObjectFileName(objectKeyFileName(objectKey)) != ""
}

func (s *Store) prepareReviewImageConfirmation(req ConfirmReviewImageUploadRequest) (*ReviewImageUploadTicket, ObjectStorageConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.reviewImageTickets[req.TicketID]
	if !reviewImageUploadTicketMatchesConfirm(ticket, req) {
		return nil, ObjectStorageConfig{}, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return cloneReviewImageUploadTicket(ticket), ObjectStorageConfig{}, nil
	}
	if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return nil, ObjectStorageConfig{}, ErrInvalidOrderState
	}
	storage := s.objectStorageConfigLocked()
	if err := reviewImageUploadTicketConfirmReady(ticket, storage, time.Now().UTC()); err != nil {
		return nil, ObjectStorageConfig{}, err
	}
	return cloneReviewImageUploadTicket(ticket), storage, nil
}

func (s *Store) commitReviewImageConfirmation(req ConfirmReviewImageUploadRequest) (*ReviewImageUploadTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.reviewImageTickets[req.TicketID]
	if !reviewImageUploadTicketMatchesConfirm(ticket, req) {
		return nil, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return cloneReviewImageUploadTicket(ticket), nil
	}
	if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	if err := reviewImageUploadTicketConfirmReady(ticket, s.objectStorageConfigLocked(), now); err != nil {
		return nil, err
	}
	ticket.Status = AfterSalesUploadTicketConfirmed
	if req.ContentSHA != "" {
		ticket.ContentSHA = req.ContentSHA
	}
	if ticket.UploadedAt.IsZero() {
		ticket.UploadedAt = now
	}
	ticket.ConfirmedAt = now
	return cloneReviewImageUploadTicket(ticket), nil
}

func reviewImageUploadTicketMatchesConfirm(ticket *ReviewImageUploadTicket, req ConfirmReviewImageUploadRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.ID == req.TicketID &&
		ticket.UserID == req.UserID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func reviewImageUploadTicketMatchesObjectCallback(ticket *ReviewImageUploadTicket, req ObjectStorageUploadCallbackRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.ID == req.TicketID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func reviewImageUploadTicketConfirmReady(ticket *ReviewImageUploadTicket, storage ObjectStorageConfig, now time.Time) error {
	if ticket == nil {
		return ErrInvalidArgument
	}
	if now.After(ticket.ExpiresAt) {
		return ErrInvalidOrderState
	}
	if storage.RequireUploadCallbackForConfirm || storage.RequireScanApprovalForConfirm {
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return ErrInvalidOrderState
		}
	} else if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return ErrInvalidOrderState
	}
	if ticket.ScanStatus == AfterSalesUploadScanRejected {
		return ErrInvalidOrderState
	}
	if storage.RequireScanApprovalForConfirm && ticket.ScanStatus != AfterSalesUploadScanPassed {
		return ErrInvalidOrderState
	}
	return nil
}

func reviewTicketObjectStorageView(ticket *ReviewImageUploadTicket) *AfterSalesEvidenceUploadTicket {
	if ticket == nil {
		return nil
	}
	return &AfterSalesEvidenceUploadTicket{
		ID:             ticket.ID,
		Provider:       ticket.Provider,
		Bucket:         ticket.Bucket,
		ObjectKey:      ticket.ObjectKey,
		PublicURL:      ticket.PublicURL,
		FileName:       ticket.FileName,
		ContentType:    ticket.ContentType,
		SizeBytes:      ticket.SizeBytes,
		MaxSizeBytes:   ticket.MaxSizeBytes,
		ContentSHA:     ticket.ContentSHA,
		UploadedByID:   ticket.UserID,
		UploadedByRole: "user",
		Status:         ticket.Status,
		ScanStatus:     ticket.ScanStatus,
		ScanResult:     ticket.ScanResult,
		CreatedAt:      ticket.CreatedAt,
		ExpiresAt:      ticket.ExpiresAt,
		UploadedAt:     ticket.UploadedAt,
		ConfirmedAt:    ticket.ConfirmedAt,
		ScanCheckedAt:  ticket.ScanCheckedAt,
	}
}

func (s *Store) preparePrescriptionImageConfirmation(req ConfirmPrescriptionImageUploadRequest) (*PrescriptionImageUploadTicket, ObjectStorageConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.prescriptionImageTickets[req.TicketID]
	if !prescriptionImageUploadTicketMatchesConfirm(ticket, req) {
		return nil, ObjectStorageConfig{}, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return clonePrescriptionImageUploadTicket(ticket), ObjectStorageConfig{}, nil
	}
	if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return nil, ObjectStorageConfig{}, ErrInvalidOrderState
	}
	storage := s.objectStorageConfigLocked()
	if err := prescriptionImageUploadTicketConfirmReady(ticket, storage, time.Now().UTC()); err != nil {
		return nil, ObjectStorageConfig{}, err
	}
	return clonePrescriptionImageUploadTicket(ticket), storage, nil
}

func (s *Store) commitPrescriptionImageConfirmation(req ConfirmPrescriptionImageUploadRequest) (*PrescriptionImageUploadTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket := s.prescriptionImageTickets[req.TicketID]
	if !prescriptionImageUploadTicketMatchesConfirm(ticket, req) {
		return nil, ErrInvalidArgument
	}
	if ticket.Status == AfterSalesUploadTicketConfirmed {
		return clonePrescriptionImageUploadTicket(ticket), nil
	}
	if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return nil, ErrInvalidOrderState
	}
	now := time.Now().UTC()
	if err := prescriptionImageUploadTicketConfirmReady(ticket, s.objectStorageConfigLocked(), now); err != nil {
		return nil, err
	}
	ticket.Status = AfterSalesUploadTicketConfirmed
	if req.ContentSHA != "" {
		ticket.ContentSHA = req.ContentSHA
	}
	if ticket.UploadedAt.IsZero() {
		ticket.UploadedAt = now
	}
	ticket.ConfirmedAt = now
	return clonePrescriptionImageUploadTicket(ticket), nil
}

func prescriptionImageUploadTicketMatchesConfirm(ticket *PrescriptionImageUploadTicket, req ConfirmPrescriptionImageUploadRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.ID == req.TicketID &&
		ticket.UserID == req.UserID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func prescriptionImageUploadTicketMatchesObjectCallback(ticket *PrescriptionImageUploadTicket, req ObjectStorageUploadCallbackRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.ID == req.TicketID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func prescriptionImageUploadTicketConfirmReady(ticket *PrescriptionImageUploadTicket, storage ObjectStorageConfig, now time.Time) error {
	if ticket == nil {
		return ErrInvalidArgument
	}
	if now.After(ticket.ExpiresAt) {
		return ErrInvalidOrderState
	}
	if storage.RequireUploadCallbackForConfirm || storage.RequireScanApprovalForConfirm {
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return ErrInvalidOrderState
		}
	} else if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return ErrInvalidOrderState
	}
	if ticket.ScanStatus == AfterSalesUploadScanRejected {
		return ErrInvalidOrderState
	}
	if storage.RequireScanApprovalForConfirm && ticket.ScanStatus != AfterSalesUploadScanPassed {
		return ErrInvalidOrderState
	}
	return nil
}

func prescriptionTicketObjectStorageView(ticket *PrescriptionImageUploadTicket) *AfterSalesEvidenceUploadTicket {
	if ticket == nil {
		return nil
	}
	return &AfterSalesEvidenceUploadTicket{
		ID:             ticket.ID,
		Provider:       ticket.Provider,
		Bucket:         ticket.Bucket,
		ObjectKey:      ticket.ObjectKey,
		PublicURL:      ticket.PublicURL,
		FileName:       ticket.FileName,
		ContentType:    ticket.ContentType,
		SizeBytes:      ticket.SizeBytes,
		MaxSizeBytes:   ticket.MaxSizeBytes,
		ContentSHA:     ticket.ContentSHA,
		UploadedByID:   ticket.UserID,
		UploadedByRole: "user",
		Status:         ticket.Status,
		ScanStatus:     ticket.ScanStatus,
		ScanResult:     ticket.ScanResult,
		CreatedAt:      ticket.CreatedAt,
		ExpiresAt:      ticket.ExpiresAt,
		UploadedAt:     ticket.UploadedAt,
		ConfirmedAt:    ticket.ConfirmedAt,
		ScanCheckedAt:  ticket.ScanCheckedAt,
	}
}

func (s *Store) prescriptionImageUploadTicketForReviewLocked(req CreatePrescriptionReviewRequest) (*PrescriptionImageUploadTicket, error) {
	if s.prescriptionImageTickets == nil {
		return nil, ErrInvalidArgument
	}
	if req.PrescriptionImageTicketID != "" {
		ticket := s.prescriptionImageTickets[req.PrescriptionImageTicketID]
		return prescriptionImageUploadTicketReadyForReview(ticket, req)
	}
	for _, ticket := range s.prescriptionImageTickets {
		if ticket != nil && ticket.ObjectKey == req.PrescriptionObjectKey && ticket.UserID == req.UserID {
			return prescriptionImageUploadTicketReadyForReview(ticket, req)
		}
	}
	return nil, ErrInvalidArgument
}

func prescriptionImageUploadTicketReadyForReview(ticket *PrescriptionImageUploadTicket, req CreatePrescriptionReviewRequest) (*PrescriptionImageUploadTicket, error) {
	if ticket == nil {
		return nil, ErrInvalidArgument
	}
	if ticket.UserID != req.UserID {
		return nil, ErrInvalidArgument
	}
	if req.PrescriptionObjectKey != "" && ticket.ObjectKey != req.PrescriptionObjectKey {
		return nil, ErrInvalidArgument
	}
	if ticket.ProductID != "" && req.ProductID != "" && ticket.ProductID != req.ProductID {
		return nil, ErrInvalidArgument
	}
	if ticket.Status != AfterSalesUploadTicketConfirmed {
		return nil, ErrInvalidOrderState
	}
	return ticket, nil
}

func (s *Store) afterSalesUploadTicketForConfirmLocked(req ConfirmAfterSalesEvidenceUploadRequest) *AfterSalesEvidenceUploadTicket {
	if s.afterSalesUploadTickets == nil {
		return nil
	}
	if req.TicketID != "" {
		ticket := s.afterSalesUploadTickets[req.TicketID]
		if afterSalesUploadTicketMatchesConfirm(ticket, req) {
			return ticket
		}
		return nil
	}
	for _, ticket := range s.afterSalesUploadTickets {
		if afterSalesUploadTicketMatchesConfirm(ticket, req) {
			return ticket
		}
	}
	return nil
}

func afterSalesUploadTicketMatchesConfirm(ticket *AfterSalesEvidenceUploadTicket, req ConfirmAfterSalesEvidenceUploadRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.RequestID == req.RequestID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.UploadedByID == req.ActorID &&
		ticket.UploadedByRole == req.ActorRole &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func afterSalesUploadTicketMatchesObjectCallback(ticket *AfterSalesEvidenceUploadTicket, req ObjectStorageUploadCallbackRequest) bool {
	if ticket == nil {
		return false
	}
	return ticket.ID == req.TicketID &&
		ticket.ObjectKey == req.ObjectKey &&
		ticket.ContentType == req.ContentType &&
		ticket.SizeBytes == req.SizeBytes
}

func afterSalesUploadTicketConfirmReady(ticket *AfterSalesEvidenceUploadTicket, storage ObjectStorageConfig, now time.Time) error {
	if ticket == nil {
		return ErrInvalidArgument
	}
	if now.After(ticket.ExpiresAt) {
		return ErrInvalidOrderState
	}
	if storage.RequireUploadCallbackForConfirm || storage.RequireScanApprovalForConfirm {
		if ticket.Status != AfterSalesUploadTicketUploaded {
			return ErrInvalidOrderState
		}
	} else if ticket.Status != AfterSalesUploadTicketIssued && ticket.Status != AfterSalesUploadTicketUploaded {
		return ErrInvalidOrderState
	}
	if ticket.ScanStatus == AfterSalesUploadScanRejected {
		return ErrInvalidOrderState
	}
	if storage.RequireScanApprovalForConfirm && ticket.ScanStatus != AfterSalesUploadScanPassed {
		return ErrInvalidOrderState
	}
	return nil
}

func objectStorageCleanupCandidateFromTicket(ticket *AfterSalesEvidenceUploadTicket, now time.Time, grace time.Duration) (ObjectStorageCleanupCandidate, bool) {
	if ticket == nil || ticket.Status == AfterSalesUploadTicketConfirmed || ticket.Status == AfterSalesUploadTicketDeleted || ticket.ObjectKey == "" {
		return ObjectStorageCleanupCandidate{}, false
	}
	reason := ""
	retainUntil := time.Time{}
	if ticket.ScanStatus == AfterSalesUploadScanRejected {
		reason = AfterSalesObjectCleanupRejected
		retainUntil = ticket.ScanCheckedAt
		if retainUntil.IsZero() {
			retainUntil = ticket.ExpiresAt
		}
		retainUntil = retainUntil.Add(grace)
	}
	expiredAt := ticket.ExpiresAt.Add(grace)
	if reason == "" && !ticket.ExpiresAt.IsZero() && !now.Before(expiredAt) {
		reason = AfterSalesObjectCleanupExpired
		retainUntil = expiredAt
	}
	if reason == "" || retainUntil.IsZero() || now.Before(retainUntil) {
		return ObjectStorageCleanupCandidate{}, false
	}
	return ObjectStorageCleanupCandidate{
		TicketID:            ticket.ID,
		RequestID:           ticket.RequestID,
		OrderID:             ticket.OrderID,
		Provider:            ticket.Provider,
		Bucket:              ticket.Bucket,
		ObjectKey:           ticket.ObjectKey,
		PublicURL:           ticket.PublicURL,
		Status:              ticket.Status,
		ScanStatus:          ticket.ScanStatus,
		Reason:              reason,
		RetainUntil:         retainUntil.UTC(),
		CreatedAt:           ticket.CreatedAt,
		ExpiresAt:           ticket.ExpiresAt,
		UploadedAt:          ticket.UploadedAt,
		ScanCheckedAt:       ticket.ScanCheckedAt,
		CleanupAttempts:     ticket.CleanupAttempts,
		LastCleanupError:    ticket.LastCleanupError,
		LastCleanupFailedAt: ticket.LastCleanupFailedAt,
	}, true
}

func normalizeObjectStorageCleanupWindow(req ObjectStorageCleanupCandidatesRequest) (int, time.Duration, time.Time, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	grace := time.Duration(req.GraceSeconds) * time.Second
	if grace < 0 {
		return 0, 0, time.Time{}, ErrInvalidArgument
	}
	if grace == 0 {
		grace = time.Hour
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return limit, grace, now, nil
}

func normalizeObjectStorageCleanupReason(value string) string {
	switch strings.TrimSpace(value) {
	case AfterSalesObjectCleanupExpired:
		return AfterSalesObjectCleanupExpired
	case AfterSalesObjectCleanupRejected:
		return AfterSalesObjectCleanupRejected
	default:
		return ""
	}
}

func sanitizeObjectStorageCleanupError(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > 1000 {
		return string(runes[:1000])
	}
	return value
}

func objectKeyFileName(objectKey string) string {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return ""
	}
	index := strings.LastIndex(objectKey, "/")
	if index >= 0 {
		return objectKey[index+1:]
	}
	return objectKey
}

func normalizeEvidenceContentType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/jpg":
		return "image/jpeg"
	case "image/png":
		return "image/png"
	case "image/webp":
		return "image/webp"
	case "image/heic":
		return "image/heic"
	case "application/pdf":
		return "application/pdf"
	default:
		return ""
	}
}

func normalizeAfterSalesUploadTicketStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "", AfterSalesUploadTicketIssued:
		return AfterSalesUploadTicketIssued
	case AfterSalesUploadTicketUploaded:
		return AfterSalesUploadTicketUploaded
	case AfterSalesUploadTicketConfirmed:
		return AfterSalesUploadTicketConfirmed
	case AfterSalesUploadTicketDeleted:
		return AfterSalesUploadTicketDeleted
	default:
		return ""
	}
}

func normalizeAfterSalesUploadScanStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "", AfterSalesUploadScanNotRequired:
		return AfterSalesUploadScanNotRequired
	case AfterSalesUploadScanPending:
		return AfterSalesUploadScanPending
	case AfterSalesUploadScanPassed:
		return AfterSalesUploadScanPassed
	case AfterSalesUploadScanRejected:
		return AfterSalesUploadScanRejected
	default:
		return ""
	}
}

func sortAfterSalesRequests(requests []AfterSalesRequest) {
	sort.SliceStable(requests, func(i, j int) bool {
		if !requests[i].CreatedAt.Equal(requests[j].CreatedAt) {
			return requests[i].CreatedAt.After(requests[j].CreatedAt)
		}
		return requests[i].ID < requests[j].ID
	})
}

func normalizeAdminOperationsSnapshotLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func isAdminExceptionOrder(order *Order, now time.Time) bool {
	if order == nil {
		return false
	}
	switch order.Status {
	case StatusRefundPending, StatusCancelled:
		return true
	}
	referenceAt := order.UpdatedAt
	if referenceAt.IsZero() {
		referenceAt = order.CreatedAt
	}
	if referenceAt.IsZero() {
		return false
	}
	switch order.Status {
	case StatusDispatching:
		return referenceAt.Before(now.Add(-10 * time.Minute))
	case StatusRiderAssigned:
		return referenceAt.Before(now.Add(-30 * time.Minute))
	default:
		return false
	}
}

func adminMerchantRiskRank(snapshot AdminMerchantSnapshot) int {
	score := 0
	if !snapshot.CanAcceptOrders || snapshot.QualificationPopupRequired || len(snapshot.MissingQualifications) > 0 {
		score += 10
	}
	if snapshot.Deposit == nil || snapshot.Deposit.Status != DepositStatusPaid || snapshot.Account.DepositStatus != DepositStatusPaid {
		score += 5
	}
	return score
}

func sortAfterSalesEvents(events []AfterSalesEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].ID < events[j].ID
	})
}

func sortAfterSalesEvidence(evidence []AfterSalesEvidence) {
	sort.SliceStable(evidence, func(i, j int) bool {
		if !evidence[i].CreatedAt.Equal(evidence[j].CreatedAt) {
			return evidence[i].CreatedAt.Before(evidence[j].CreatedAt)
		}
		return evidence[i].ID < evidence[j].ID
	})
}

func sortAfterSalesUploadTickets(tickets []AfterSalesEvidenceUploadTicket) {
	sort.SliceStable(tickets, func(i, j int) bool {
		if !tickets[i].CreatedAt.Equal(tickets[j].CreatedAt) {
			return tickets[i].CreatedAt.Before(tickets[j].CreatedAt)
		}
		return tickets[i].ID < tickets[j].ID
	})
}

func sortObjectStorageCleanupCandidates(candidates []ObjectStorageCleanupCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if !candidates[i].RetainUntil.Equal(candidates[j].RetainUntil) {
			return candidates[i].RetainUntil.Before(candidates[j].RetainUntil)
		}
		return candidates[i].TicketID < candidates[j].TicketID
	})
}

func isSixDigitPaymentPassword(password string) bool {
	if len(password) != 6 {
		return false
	}
	for _, char := range password {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func normalizeMainlandPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) != 11 || phone[0] != '1' {
		return ""
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return ""
		}
	}
	return phone
}

func normalizePhoneAuthPurpose(purpose string) string {
	switch strings.TrimSpace(purpose) {
	case "register":
		return "register"
	default:
		return "login"
	}
}

func maskMainlandPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

func phoneVerificationCode(phone string, purpose string, now time.Time) string {
	if value, err := crand.Int(crand.Reader, big.NewInt(1000000)); err == nil {
		return fmt.Sprintf("%06d", value.Int64())
	}
	bucket := now.UTC().Format("200601021504")
	sum := sha256.Sum256([]byte("infinitech-phone-code:" + phone + ":" + purpose + ":" + bucket))
	value := int(sum[0])<<16 | int(sum[1])<<8 | int(sum[2])
	return fmt.Sprintf("%06d", value%1000000)
}

func sanitizePhoneVerificationTicket(ticket *PhoneVerificationCodeTicket, config PhoneVerificationConfig) *PhoneVerificationCodeTicket {
	cloned := clonePhoneVerificationCodeTicket(ticket)
	if cloned == nil {
		return nil
	}
	if !config.ReturnDevCode {
		cloned.DevCode = ""
	}
	return cloned
}

func hashUserPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 6 || len(password) > 72 {
		return "", ErrInvalidArgument
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyUserPassword(passwordHash string, password string) bool {
	password = strings.TrimSpace(password)
	if passwordHash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

func hashPaymentPassword(password string) string {
	sum := sha256.Sum256([]byte("infinitech-wallet:" + password))
	return hex.EncodeToString(sum[:])
}

func hashRiderPassword(password string) (string, error) {
	return hashLoginPassword(password)
}

func hashLoginPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 || len(password) > 72 {
		return "", ErrInvalidArgument
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func normalizeProductStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "":
		return ProductStatusActive
	case ProductStatusActive, ProductStatusSoldOut, ProductStatusRemoved:
		return strings.TrimSpace(status)
	default:
		return ""
	}
}

func normalizeIngredientList(items []string) []string {
	normalized := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		normalized = append(normalized, item)
		if len(normalized) == 20 {
			break
		}
	}
	return normalized
}

func (s *Store) merchantProfileLocked(merchantID string) *MerchantProfile {
	account := s.merchants[merchantID]
	if account == nil {
		return nil
	}
	qualifications := make([]MerchantQualification, 0, len(s.merchantQualifications[merchantID]))
	approvedByType := map[string]bool{}
	now := time.Now().UTC()
	for _, item := range s.merchantQualifications[merchantID] {
		if item == nil {
			continue
		}
		qualifications = append(qualifications, *cloneMerchantQualification(item))
		if item.Status == QualificationStatusApproved && item.ExpiresAt.After(now) {
			approvedByType[item.Type] = true
		}
	}
	missing := []string{}
	for _, required := range []string{QualificationBusinessLicense, QualificationHealthCertificate} {
		if !approvedByType[required] {
			missing = append(missing, required)
		}
	}
	canAccept := len(missing) == 0 && account.DepositStatus == DepositStatusPaid
	profile := &MerchantProfile{
		Account:                    *cloneMerchantAccount(account),
		Qualifications:             qualifications,
		MissingQualifications:      missing,
		CanAcceptOrders:            canAccept,
		QualificationPopupRequired: len(missing) > 0,
		Staff:                      s.merchantStaffLocked(merchantID),
		SupplementalMaterials:      s.merchantMaterialsLocked(merchantID),
	}
	if len(missing) > 0 {
		profile.Account.Status = ShopStatusQualificationExpired
		profile.QualificationPopupCode = "MERCHANT_QUALIFICATION_EXPIRED"
	} else if profile.Account.Status == "pending_qualification" {
		profile.Account.Status = ShopStatusActive
	}
	return profile
}

func (s *Store) merchantStaffLocked(merchantID string) []MerchantStaff {
	staff := make([]MerchantStaff, 0, len(s.merchantStaff[merchantID]))
	for _, item := range s.merchantStaff[merchantID] {
		if item == nil {
			continue
		}
		staff = append(staff, *cloneMerchantStaff(item))
	}
	sort.SliceStable(staff, func(i, j int) bool {
		return staff[i].ID < staff[j].ID
	})
	return staff
}

func (s *Store) merchantMaterialsLocked(merchantID string) []MerchantSupplementalMaterial {
	materials := make([]MerchantSupplementalMaterial, 0, len(s.merchantMaterials[merchantID]))
	for _, item := range s.merchantMaterials[merchantID] {
		if item == nil {
			continue
		}
		materials = append(materials, *cloneMerchantSupplementalMaterial(item))
	}
	sort.SliceStable(materials, func(i, j int) bool {
		return materials[i].ID < materials[j].ID
	})
	return materials
}

func isMerchantAccountType(value string) bool {
	switch value {
	case MerchantAccountStandard, MerchantAccountPharmacy, MerchantAccountClinic, MerchantAccountPlatformService:
		return true
	default:
		return false
	}
}

func isMerchantQualificationType(value string) bool {
	switch value {
	case QualificationBusinessLicense, QualificationHealthCertificate, QualificationSupplementalDocument:
		return true
	default:
		return false
	}
}

func wechatPrepayResponseForTransaction(transaction *PaymentTransaction) *WechatPrepayResponse {
	nonce := shortHash(transaction.OutTradeNo + ":nonce")
	timeStamp := fmt.Sprintf("%d", transaction.CreatedAt.Unix())
	prepayID := "prepay_" + transaction.OutTradeNo
	paySign := shortHash(transaction.OutTradeNo + ":" + prepayID + ":" + timeStamp + ":" + nonce)
	return &WechatPrepayResponse{
		AppID:      "wx_app_configured_later",
		PrepayID:   prepayID,
		OutTradeNo: transaction.OutTradeNo,
		TimeStamp:  timeStamp,
		NonceStr:   nonce,
		Package:    "prepay_id=" + prepayID,
		SignType:   "RSA",
		PaySign:    paySign,
		AmountFen:  transaction.AmountFen,
	}
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:16]
}

func (s *Store) createWalletTransactionLocked(userID string, orderID string, transactionType string, amountFen int64, method string, idempotencyKey string) *WalletTransaction {
	s.nextTransactionID++
	return &WalletTransaction{
		ID:             fmt.Sprintf("wtx_%d", s.nextTransactionID),
		UserID:         userID,
		OrderID:        orderID,
		Type:           transactionType,
		AmountFen:      amountFen,
		PaymentMethod:  method,
		IdempotencyKey: idempotencyKey,
		Status:         "success",
		CreatedAt:      time.Now().UTC(),
	}
}

func (s *Store) cartSummaryLocked(userID string, shopID string) *CartSummary {
	key := cartKey(userID, shopID)
	source := s.cartItems[key]
	items := make([]CartItem, 0, len(source))
	var itemsTotal int64
	for _, item := range source {
		if item == nil || !item.Selected || item.Quantity <= 0 || item.UnitPriceFen <= 0 {
			continue
		}
		nextItem := *cloneCartItem(item)
		if strings.TrimSpace(nextItem.ImageURL) == "" {
			if product := s.products[nextItem.ProductID]; product != nil {
				nextItem.ImageURL = product.ImageURL
			}
		}
		items = append(items, nextItem)
		itemsTotal += item.UnitPriceFen * int64(item.Quantity)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})
	deliveryFee := int64(300)
	packagingFee := int64(100 * len(items))
	if len(items) == 0 {
		deliveryFee = 0
		packagingFee = 0
	}
	shopName := ""
	if shop := s.shops[shopID]; shop != nil {
		shopName = shop.Name
	}
	return &CartSummary{
		UserID:          userID,
		ShopID:          shopID,
		ShopName:        shopName,
		Items:           items,
		ItemsTotalFen:   itemsTotal,
		DeliveryFeeFen:  deliveryFee,
		PackagingFeeFen: packagingFee,
		DiscountFen:     0,
		PayableFen:      OrderPayableFen(itemsTotal, deliveryFee, packagingFee, 0),
	}
}

func (s *Store) findAddressLocked(userID string, addressID string) *UserAddress {
	for _, address := range s.addresses[userID] {
		if address.ID == addressID {
			return cloneUserAddress(address)
		}
	}
	return nil
}

func normalizeOrderOptions(options OrderOptions) OrderOptions {
	options.Remark = strings.TrimSpace(options.Remark)
	if len(options.Remark) > 120 {
		options.Remark = options.Remark[:120]
	}
	if options.TablewareCount < 0 {
		options.TablewareCount = 0
	}
	if options.TablewareCount > 99 {
		options.TablewareCount = 99
	}
	return options
}

func cartKey(userID string, shopID string) string {
	return userID + "::" + shopID
}

func orderAddressSnapshot(address UserAddress) OrderAddress {
	return OrderAddress{
		ContactName:  address.ContactName,
		ContactPhone: address.ContactPhone,
		City:         address.City,
		Detail:       address.Detail,
		Tag:          address.Tag,
	}
}

func cloneHomeModules(input []HomeModule) []HomeModule {
	output := make([]HomeModule, len(input))
	copy(output, input)
	return output
}

func cloneHomeCards(input []HomeCard) []HomeCard {
	output := make([]HomeCard, len(input))
	copy(output, input)
	return output
}

func cloneAppUser(input *AppUser) *AppUser {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePhoneVerificationCodeTicket(input *PhoneVerificationCodeTicket) *PhoneVerificationCodeTicket {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMerchantInvite(input *MerchantOnboardingInvite) *MerchantOnboardingInvite {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMerchantAccount(input *MerchantAccount) *MerchantAccount {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMerchantQualification(input *MerchantQualification) *MerchantQualification {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMerchantStaff(input *MerchantStaff) *MerchantStaff {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMerchantSupplementalMaterial(input *MerchantSupplementalMaterial) *MerchantSupplementalMaterial {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneRiderAccount(input *RiderAccount) *RiderAccount {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneDepositAccount(input *DepositAccount) *DepositAccount {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneStationTaskConfig(input *StationTaskConfig) *StationTaskConfig {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneStationServiceArea(input *StationServiceArea) *StationServiceArea {
	if input == nil {
		return nil
	}
	output := *input
	output.ShopIDs = append([]string{}, input.ShopIDs...)
	return &output
}

func cloneShop(input *Shop) *Shop {
	if input == nil {
		return nil
	}
	output := *input
	output.Capabilities = append([]string{}, input.Capabilities...)
	output.Qualifications = append([]string{}, input.Qualifications...)
	return &output
}

func cloneMerchantProduct(input *MerchantProduct) *MerchantProduct {
	if input == nil {
		return nil
	}
	output := *input
	output.IngredientList = append([]string{}, input.IngredientList...)
	return &output
}

func cloneGroupbuyVoucher(input *GroupbuyVoucher) *GroupbuyVoucher {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneUserAddress(input *UserAddress) *UserAddress {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneCartItem(input *CartItem) *CartItem {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneCartSummary(input *CartSummary) *CartSummary {
	if input == nil {
		return nil
	}
	output := *input
	output.Items = append([]CartItem{}, input.Items...)
	return &output
}

func cloneOrder(input *Order) *Order {
	if input == nil {
		return nil
	}
	output := *input
	output.Items = append([]OrderItem{}, input.Items...)
	output.Events = append([]OrderEvent{}, input.Events...)
	return &output
}

func cloneWalletAccount(input *WalletAccount) *WalletAccount {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneWalletTransaction(input *WalletTransaction) *WalletTransaction {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneReview(input *Review) *Review {
	if input == nil {
		return nil
	}
	output := *input
	output.ImageURLs = append([]string{}, input.ImageURLs...)
	output.ItemRatings = cloneReviewItemRatings(input.ItemRatings)
	output.Tags = append([]string{}, input.Tags...)
	return &output
}

func cloneReviewItemRatings(input []ReviewItemRating) []ReviewItemRating {
	if len(input) == 0 {
		return nil
	}
	output := make([]ReviewItemRating, 0, len(input))
	for _, item := range input {
		cloned := item
		cloned.Tags = append([]string{}, item.Tags...)
		output = append(output, cloned)
	}
	return output
}

func cloneFeedbackTicket(input *FeedbackTicket) *FeedbackTicket {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneServiceTicket(input *ServiceTicket) *ServiceTicket {
	if input == nil {
		return nil
	}
	output := *input
	output.Attachments = append([]string{}, input.Attachments...)
	return &output
}

func cloneServiceTicketEvent(input *ServiceTicketEvent) *ServiceTicketEvent {
	if input == nil {
		return nil
	}
	output := *input
	output.Attachments = append([]string{}, input.Attachments...)
	return &output
}

func cloneServiceTicketQualityReview(input *ServiceTicketQualityReview) *ServiceTicketQualityReview {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneCirclePost(input *CirclePost) *CirclePost {
	if input == nil {
		return nil
	}
	output := *input
	output.ImageURLs = append([]string{}, input.ImageURLs...)
	output.Tags = append([]string{}, input.Tags...)
	return &output
}

func cloneMealMatchProfile(input *MealMatchProfile) *MealMatchProfile {
	if input == nil {
		return nil
	}
	output := *input
	output.PersonalityTraits = append([]string{}, input.PersonalityTraits...)
	output.DietaryHabits = append([]string{}, input.DietaryHabits...)
	return &output
}

func cloneMealMatchModerationRecord(input *MealMatchModerationRecord) *MealMatchModerationRecord {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func mealMatchCandidateFromProfiles(user MealMatchProfile, candidate MealMatchProfile, displayName string) MealMatchCandidate {
	matchedPersonality := intersectStrings(user.PersonalityTraits, candidate.PersonalityTraits)
	matchedDietary := intersectStrings(user.DietaryHabits, candidate.DietaryHabits)
	score := len(matchedDietary)*60 + len(matchedPersonality)*40
	if mealMatchSameBuilding(user, candidate) {
		score += 20
	} else if mealMatchSameSchool(user, candidate) {
		score += 10
	}
	name := mealMatchDisplayName(candidate.UserID, displayName)
	return MealMatchCandidate{
		UserID:                   candidate.UserID,
		DisplayName:              name,
		AvatarInitial:            avatarInitial(name),
		Gender:                   candidate.Gender,
		DistanceText:             mealMatchDistanceText(user, candidate),
		SchoolName:               candidate.SchoolName,
		CampusName:               candidate.CampusName,
		BuildingName:             mealMatchVisibleBuildingName(user, candidate),
		SameSchool:               mealMatchSameSchool(user, candidate),
		SameBuilding:             mealMatchSameBuilding(user, candidate),
		PrivacyScope:             normalizeMealMatchPrivacyScope(candidate.PrivacyScope),
		LocationPrecision:        normalizeMealMatchLocationPrecision(candidate.LocationPrecision),
		MatchScore:               score,
		MatchedPersonalityTraits: matchedPersonality,
		MatchedDietaryHabits:     matchedDietary,
		PersonalityTraits:        append([]string{}, candidate.PersonalityTraits...),
		DietaryHabits:            append([]string{}, candidate.DietaryHabits...),
		SafetyBadges:             mealMatchSafetyBadges(user, candidate),
		PrivacyNotice:            mealMatchCandidatePrivacyNotice(user, candidate),
	}
}

func intersectStrings(left []string, right []string) []string {
	seen := map[string]bool{}
	for _, item := range left {
		item = strings.TrimSpace(item)
		if item != "" {
			seen[item] = true
		}
	}
	output := []string{}
	for _, item := range right {
		item = strings.TrimSpace(item)
		if item != "" && seen[item] {
			output = append(output, item)
		}
	}
	return output
}

func mealMatchDisplayName(userID string, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	switch strings.TrimSpace(userID) {
	case "user_buddy_lunch":
		return "同楼午餐搭子"
	case "user_buddy_weekend":
		return "周末探店搭子"
	case "user_buddy_library":
		return "图书馆晚餐搭子"
	default:
		if fallback != "" && fallback != "悦享用户" {
			return fallback
		}
		return "饭搭用户"
	}
}

func mealMatchCanShowCandidate(user MealMatchProfile, candidate MealMatchProfile) bool {
	if !mealMatchSameSchool(user, candidate) {
		return false
	}
	if normalizeMealMatchPrivacyScope(user.PrivacyScope) == MealMatchPrivacySameBuilding && !mealMatchSameBuilding(user, candidate) {
		return false
	}
	if normalizeMealMatchPrivacyScope(candidate.PrivacyScope) == MealMatchPrivacySameBuilding && !mealMatchSameBuilding(user, candidate) {
		return false
	}
	return true
}

func mealMatchSameSchool(user MealMatchProfile, candidate MealMatchProfile) bool {
	return strings.TrimSpace(user.SchoolID) != "" && strings.TrimSpace(user.SchoolID) == strings.TrimSpace(candidate.SchoolID)
}

func mealMatchSameBuilding(user MealMatchProfile, candidate MealMatchProfile) bool {
	return mealMatchSameSchool(user, candidate) && strings.TrimSpace(user.BuildingID) != "" && strings.TrimSpace(user.BuildingID) == strings.TrimSpace(candidate.BuildingID)
}

func mealMatchDistanceText(user MealMatchProfile, candidate MealMatchProfile) string {
	if mealMatchSameBuilding(user, candidate) {
		return "同楼可约"
	}
	if mealMatchSameSchool(user, candidate) {
		return "同校范围"
	}
	return "已隐藏位置"
}

func mealMatchVisibleBuildingName(user MealMatchProfile, candidate MealMatchProfile) string {
	if !mealMatchSameBuilding(user, candidate) {
		return ""
	}
	if normalizeMealMatchPrivacyScope(candidate.PrivacyScope) == MealMatchPrivacySameBuilding || normalizeMealMatchLocationPrecision(candidate.LocationPrecision) == MealMatchLocationBuildingOnly {
		return candidate.BuildingName
	}
	return ""
}

func mealMatchSafetyBadges(user MealMatchProfile, candidate MealMatchProfile) []string {
	badges := []string{"已签真实性承诺", "已签免责承诺", "同校校验"}
	if mealMatchSameBuilding(user, candidate) {
		badges = append(badges, "同楼可见")
	}
	return badges
}

func mealMatchProfilePrivacyNotice(profile MealMatchProfile) string {
	if normalizeMealMatchPrivacyScope(profile.PrivacyScope) == MealMatchPrivacySameBuilding {
		return "仅向同校且同楼用户展示，默认隐藏手机号和精确定位。"
	}
	return "仅向同校用户展示，默认隐藏楼栋、手机号和精确定位。"
}

func mealMatchCandidatePrivacyNotice(user MealMatchProfile, candidate MealMatchProfile) string {
	if mealMatchSameBuilding(user, candidate) && normalizeMealMatchPrivacyScope(candidate.PrivacyScope) == MealMatchPrivacySameBuilding {
		return "同楼匹配，仅展示楼栋标签，不公开手机号与精确定位。"
	}
	return "同校匹配，仅展示校区和偏好标签，不公开手机号、楼栋与精确位置。"
}

func mealMatchSchoolName(schoolID string, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}
	switch strings.TrimSpace(schoolID) {
	case "infinitech_university":
		return "无限科技大学"
	case "city_college":
		return "城市学院"
	default:
		return ""
	}
}

func mealMatchCampusName(schoolID string, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}
	switch strings.TrimSpace(schoolID) {
	case "infinitech_university":
		return "东区"
	case "city_college":
		return "主校区"
	default:
		return ""
	}
}

func (s *Store) mealMatchDeviceRiskLocked(userID string, deviceID string, checkedAt time.Time) MealMatchDeviceRiskCheck {
	if checkedAt.IsZero() {
		checkedAt = time.Now().UTC()
	} else {
		checkedAt = checkedAt.UTC()
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return MealMatchDeviceRiskCheck{
			State:      MealMatchDeviceRiskReview,
			ReasonCode: MealMatchDeviceRiskMissing,
			Reason:     "设备识别缺失，需人工复核后才能开启饭搭推荐。",
			CheckedAt:  checkedAt,
		}
	}
	normalized := strings.ToLower(deviceID)
	if containsAny(normalized, "blocked", "simulator_farm", "risk_device") {
		return MealMatchDeviceRiskCheck{
			State:      MealMatchDeviceRiskBlocked,
			ReasonCode: MealMatchDeviceRiskKnownBlocked,
			Reason:     "设备环境触发高风险规则，暂不可开启找饭搭。",
			CheckedAt:  checkedAt,
		}
	}
	sharedCount := 0
	for otherUserID, profile := range s.mealMatchProfiles {
		if profile == nil || otherUserID == userID {
			continue
		}
		if strings.TrimSpace(profile.DeviceID) == deviceID && normalizeMealMatchModerationStatus(profile.ModerationStatus) != MealMatchModerationRejected {
			sharedCount++
		}
	}
	if sharedCount > 0 {
		return MealMatchDeviceRiskCheck{
			State:      MealMatchDeviceRiskReview,
			ReasonCode: MealMatchDeviceRiskSharedDevice,
			Reason:     "同一设备存在多个饭搭资料，需人工复核设备环境。",
			CheckedAt:  checkedAt,
		}
	}
	return MealMatchDeviceRiskCheck{
		State:     MealMatchDeviceRiskPassed,
		Reason:    "设备环境校验通过",
		CheckedAt: checkedAt,
	}
}

func applyMealMatchDeviceRiskToProfile(profile *MealMatchProfile, risk MealMatchDeviceRiskCheck) {
	if profile == nil {
		return
	}
	profile.DeviceRiskState = risk.State
	profile.DeviceRiskReasonCode = risk.ReasonCode
	profile.DeviceRiskReason = risk.Reason
	profile.DeviceRiskCheckedAt = risk.CheckedAt
}

func mealMatchBlockKey(userID string, targetUserID string) string {
	return "mmod_block_" + shortHash(strings.TrimSpace(userID)+"::"+strings.TrimSpace(targetUserID))
}

func (s *Store) createMealMatchProfileReviewLocked(userID string, now time.Time) *MealMatchModerationRecord {
	s.nextMealMatchModerationID++
	record := &MealMatchModerationRecord{
		ID:           fmt.Sprintf("mmod_%d", s.nextMealMatchModerationID),
		UserID:       userID,
		TargetUserID: userID,
		Action:       MealMatchModerationProfileReview,
		Reason:       "profile_submitted",
		Description:  "找饭搭资料提交后进入人工审核队列",
		Status:       MealMatchModerationPending,
		CreatedAt:    now,
	}
	s.mealMatchModeration[record.ID] = cloneMealMatchModerationRecord(record)
	return record
}

func (s *Store) mealMatchBlockedLocked(userID string, targetUserID string) bool {
	if record := s.mealMatchModeration[mealMatchBlockKey(userID, targetUserID)]; record != nil && record.Action == MealMatchModerationBlocked {
		return true
	}
	return false
}

func mealMatchMissingIncludes(missing []string, value string) bool {
	for _, item := range missing {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	return false
}

func normalizeMealMatchModerationDecision(decision string) string {
	switch strings.TrimSpace(decision) {
	case "approve", MealMatchModerationApproved:
		return MealMatchModerationApproved
	case "reject", MealMatchModerationRejected:
		return MealMatchModerationRejected
	default:
		return ""
	}
}

func cloneRedPacketDetail(input *RedPacketDetail) *RedPacketDetail {
	if input == nil {
		return nil
	}
	output := *input
	output.Shares = append([]RedPacketShare{}, input.Shares...)
	if input.Risk != nil {
		risk := *input.Risk
		output.Risk = &risk
	}
	return &output
}

func cloneChatThread(input *ChatThread) *ChatThread {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneChatThreadMember(input *ChatThreadMember) *ChatThreadMember {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneChatMessage(input *ChatMessage) *ChatMessage {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneChatReadState(input *ChatReadState) *ChatReadState {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePrescriptionReview(input *PrescriptionReview) *PrescriptionReview {
	if input == nil {
		return nil
	}
	output := *input
	output.Steps = append([]PrescriptionReviewStep{}, input.Steps...)
	if input.OCRResult != nil {
		output.OCRResult = clonePrescriptionOCRResult(input.OCRResult)
	}
	if input.Archive != nil {
		output.Archive = clonePrescriptionArchiveRecord(input.Archive)
	}
	return &output
}

func clonePrescriptionOCRResult(input *PrescriptionOCRResult) *PrescriptionOCRResult {
	if input == nil {
		return nil
	}
	output := *input
	output.Warnings = append([]string{}, input.Warnings...)
	return &output
}

func clonePrescriptionArchiveRecord(input *PrescriptionArchiveRecord) *PrescriptionArchiveRecord {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneReviewImageUploadTicket(input *ReviewImageUploadTicket) *ReviewImageUploadTicket {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePrescriptionImageUploadTicket(input *PrescriptionImageUploadTicket) *PrescriptionImageUploadTicket {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneMedicineOrderDetail(input *MedicineOrderDetail) *MedicineOrderDetail {
	if input == nil {
		return nil
	}
	output := *input
	output.Order = *cloneOrder(&input.Order)
	output.Items = append([]MedicineOrderItem{}, input.Items...)
	output.FeeRows = append([]MedicineFeeRow{}, input.FeeRows...)
	output.Timeline = append([]MedicineTimelineItem{}, input.Timeline...)
	return &output
}

func defaultPrescriptionReview(userID string) PrescriptionReview {
	now := time.Now().UTC().Add(-8 * time.Minute)
	review := PrescriptionReview{
		ID:           "rx_preview",
		UserID:       userID,
		PatientName:  "张三",
		PatientPhone: "13800000000",
		Address:      "望京校区 3 号宿舍楼",
		Hospital:     "校医务室",
		ProductID:    "med_amoxicillin",
		ProductName:  "阿莫西林胶囊",
		PriceFen:     1880,
		Quantity:     1,
		ImageURL:     "prescription-preview.jpg",
		Status:       PrescriptionReviewApproved,
		DoctorName:   "王医生",
		ReviewText:   "处方信息已确认，可加入购物车购买。",
		OCRResult: &PrescriptionOCRResult{
			Status:             PrescriptionOCRMatched,
			Provider:           "infinitech_preview_ocr",
			Confidence:         96,
			MatchedProductID:   "med_amoxicillin",
			MatchedProductName: "阿莫西林胶囊",
			DosageText:         "0.5g 每日 3 次，遵医嘱",
			RawText:            "张三 阿莫西林胶囊 0.5g 每日3次",
		},
		Archive: &PrescriptionArchiveRecord{
			ArchiveID:     "rxa_preview",
			ObjectKey:     "prescription-preview.jpg",
			RetainUntil:   now.AddDate(6, 0, 0),
			ArchivedAt:    now.Add(2 * time.Minute),
			RetentionText: "校内处方留档 6 年",
		},
		CreatedAt:  now,
		ReviewedAt: now.Add(3 * time.Minute),
		UpdatedAt:  now.Add(3 * time.Minute),
	}
	review.Steps = prescriptionSteps(&review)
	return review
}

func prescriptionOCRResult(req CreatePrescriptionReviewRequest, ticket *PrescriptionImageUploadTicket) *PrescriptionOCRResult {
	productID := strings.TrimSpace(req.ProductID)
	productName := strings.TrimSpace(req.ProductName)
	if ticket != nil {
		if productID == "" {
			productID = ticket.ProductID
		}
	}
	if productID == "" {
		productID = "med_amoxicillin"
	}
	if productName == "" {
		productName = "阿莫西林胶囊"
	}
	patientName := strings.TrimSpace(req.PatientName)
	if patientName == "" {
		patientName = "张三"
	}
	objectKey := strings.TrimSpace(req.PrescriptionObjectKey)
	confidence := 92
	status := PrescriptionOCRMatched
	warnings := []string{}
	if objectKey == "" && strings.TrimSpace(req.ImageURL) == "" {
		confidence = 78
		status = PrescriptionOCRNeedReview
		warnings = append(warnings, "处方影像对象未绑定，需人工复核原图")
	}
	if strings.TrimSpace(req.PrescriptionContentSHA) == "" {
		warnings = append(warnings, "缺少处方影像内容 hash，生产环境需由对象扫描回调补齐")
	}
	return &PrescriptionOCRResult{
		Status:             status,
		Provider:           "infinitech_rx_ocr_v1",
		Confidence:         confidence,
		MatchedProductID:   productID,
		MatchedProductName: productName,
		DosageText:         defaultPrescriptionDosageText(productID),
		RawText:            fmt.Sprintf("%s %s %s", patientName, productName, defaultPrescriptionDosageText(productID)),
		Warnings:           warnings,
	}
}

func prescriptionArchiveRecord(reviewID string, objectKey string, contentSHA string, now time.Time) *PrescriptionArchiveRecord {
	reviewID = strings.TrimSpace(reviewID)
	if reviewID == "" {
		reviewID = "rx_pending"
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &PrescriptionArchiveRecord{
		ArchiveID:     "rxa_" + shortHash(reviewID),
		ObjectKey:     strings.TrimSpace(objectKey),
		ContentSHA:    strings.TrimSpace(contentSHA),
		RetainUntil:   now.AddDate(6, 0, 0),
		ArchivedAt:    now.Add(2 * time.Minute),
		RetentionText: "校内处方留档 6 年",
	}
}

func defaultPrescriptionDosageText(productID string) string {
	switch strings.TrimSpace(productID) {
	case "med_amoxicillin":
		return "0.5g 每日 3 次，遵医嘱"
	case "med_cooling_patch":
		return "外用，按需贴敷"
	default:
		return "按处方与校医指导使用"
	}
}

func prescriptionSteps(review *PrescriptionReview) []PrescriptionReviewStep {
	now := time.Now().UTC()
	if review != nil && !review.CreatedAt.IsZero() {
		now = review.CreatedAt
	}
	approved := review != nil && review.Status == PrescriptionReviewApproved
	reviewedAt := now.Add(3 * time.Minute)
	if review != nil && !review.ReviewedAt.IsZero() {
		reviewedAt = review.ReviewedAt
	}
	pharmacistStatus := ServiceTicketEventActive
	orderStatus := ServiceTicketEventPending
	ocrStatus := ServiceTicketEventDone
	if review != nil && review.OCRResult != nil && review.OCRResult.Status == PrescriptionOCRNeedReview {
		ocrStatus = ServiceTicketEventActive
	}
	if approved {
		pharmacistStatus = ServiceTicketEventDone
		orderStatus = ServiceTicketEventActive
	}
	if review != nil && review.Status == PrescriptionReviewRejected {
		pharmacistStatus = ServiceTicketEventDone
		orderStatus = ServiceTicketEventPending
	}
	return []PrescriptionReviewStep{
		{Title: "处方上传", Subtitle: "照片清晰度与格式校验", Status: ServiceTicketEventDone, At: now},
		{Title: "OCR 识别", Subtitle: "识别用药人、药品、剂量与影像 hash", Status: ocrStatus, At: now.Add(time.Minute)},
		{Title: "药师复核", Subtitle: "药品、剂量、有效期检查", Status: pharmacistStatus, At: reviewedAt},
		{Title: "订单履约", Subtitle: "审核通过后进入药房备货", Status: orderStatus, At: reviewedAt.Add(time.Minute)},
	}
}

func normalizedMedicineOrderItems(input []MedicineOrderItemRequest) []MedicineOrderItem {
	items := make([]MedicineOrderItem, 0, len(input))
	for _, item := range input {
		item.ProductID = strings.TrimSpace(item.ProductID)
		item.Name = strings.TrimSpace(item.Name)
		item.Category = strings.TrimSpace(item.Category)
		item.ImageURL = strings.TrimSpace(item.ImageURL)
		if item.Name == "" || item.PriceFen <= 0 {
			continue
		}
		if item.ProductID == "" {
			item.ProductID = "med_custom"
		}
		if item.Quantity <= 0 {
			item.Quantity = 1
		}
		if item.ImageURL == "" {
			item.ImageURL = medicineProductImageURL(item.ProductID)
		}
		items = append(items, MedicineOrderItem{
			ProductID:            item.ProductID,
			Name:                 item.Name,
			Category:             item.Category,
			ImageURL:             item.ImageURL,
			PriceFen:             item.PriceFen,
			Quantity:             item.Quantity,
			RequiresPrescription: item.RequiresPrescription,
		})
	}
	return items
}

func defaultMedicineProducts() []MedicineProduct {
	return []MedicineProduct{
		{ID: "med_cooling_patch", Name: "退热贴", Subtitle: "医用退热贴 · 适用于发热物理降温", Category: "感冒发热", ImageURL: "/assets/generated/medicine-cooling-patch.jpg", PriceFen: 1290, StockCount: 26, SelectedQuantity: 1},
		{ID: "med_amoxicillin", Name: "阿莫西林胶囊", Subtitle: "处方药 · 凭处方与校医审核购买", Category: "处方药", ImageURL: "/assets/generated/medicine-capsules.jpg", PriceFen: 1880, StockCount: 12, RequiresPrescription: true},
		{ID: "med_swab", Name: "碘伏棉签", Subtitle: "外伤消毒 · 独立包装", Category: "外伤消毒", ImageURL: "/assets/generated/medicine-first-aid.jpg", PriceFen: 690, StockCount: 38},
		{ID: "med_bandage", Name: "创可贴", Subtitle: "防水透气 · 10 片装", Category: "医用耗材", ImageURL: "/assets/generated/medicine-first-aid.jpg", PriceFen: 550, StockCount: 52},
	}
}

func medicineProductImageURL(productID string) string {
	switch strings.TrimSpace(productID) {
	case "med_cooling_patch":
		return "/assets/generated/medicine-cooling-patch.jpg"
	case "med_amoxicillin":
		return "/assets/generated/medicine-capsules.jpg"
	case "med_swab", "med_bandage":
		return "/assets/generated/medicine-first-aid.jpg"
	default:
		return "/assets/generated/medicine-first-aid.jpg"
	}
}

func defaultMedicineStock() map[string]int {
	stock := map[string]int{}
	for _, product := range defaultMedicineProducts() {
		stock[product.ID] = product.StockCount
	}
	return stock
}

func (s *Store) ensureMedicineStockLocked() {
	if s.medicineStock == nil {
		s.medicineStock = defaultMedicineStock()
		return
	}
	for productID, stock := range defaultMedicineStock() {
		if _, ok := s.medicineStock[productID]; !ok {
			s.medicineStock[productID] = stock
		}
	}
}

func (s *Store) lockMedicineStockLocked(items []MedicineOrderItem) error {
	s.ensureMedicineStockLocked()
	trackedStock := defaultMedicineStock()
	for _, item := range items {
		if _, ok := trackedStock[item.ProductID]; !ok {
			continue
		}
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		if s.medicineStock[item.ProductID] < quantity {
			return ErrInsufficientStock
		}
	}
	for index := range items {
		if _, ok := trackedStock[items[index].ProductID]; !ok {
			continue
		}
		quantity := items[index].Quantity
		if quantity <= 0 {
			quantity = 1
		}
		s.medicineStock[items[index].ProductID] -= quantity
		remaining := s.medicineStock[items[index].ProductID]
		if remaining < 0 {
			remaining = 0
		}
		items[index].StockLocked = true
		items[index].StockRemaining = remaining
	}
	return nil
}

func defaultMedicineOrderItems() []MedicineOrderItem {
	return []MedicineOrderItem{
		{ProductID: "med_cooling_patch", Name: "退热贴", Category: "校医务室", ImageURL: medicineProductImageURL("med_cooling_patch"), PriceFen: 1290, Quantity: 1},
		{ProductID: "med_amoxicillin", Name: "阿莫西林胶囊", Category: "处方药", ImageURL: medicineProductImageURL("med_amoxicillin"), PriceFen: 1880, Quantity: 1, RequiresPrescription: true, PrescriptionApproved: true},
		{ProductID: "med_swab", Name: "碘伏棉签", Category: "外伤消毒", ImageURL: medicineProductImageURL("med_swab"), PriceFen: 690, Quantity: 1},
	}
}

func medicineItemsTotalFen(items []MedicineOrderItem) int64 {
	var total int64
	for _, item := range items {
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		total += item.PriceFen * int64(quantity)
	}
	return total
}

func (s *Store) medicineDetailForOrderLocked(order *Order, req MedicineOrderRequest, items []MedicineOrderItem, prescriptionStatus string, doctorName string) *MedicineOrderDetail {
	if order == nil {
		return nil
	}
	address := strings.TrimSpace(req.Address)
	if address == "" {
		address = "望京校区 3 号宿舍楼 508"
	}
	contactName := strings.TrimSpace(req.ContactName)
	if contactName == "" {
		contactName = "张三"
	}
	contactPhone := strings.TrimSpace(req.ContactPhone)
	if contactPhone == "" {
		contactPhone = "13800000000"
	}
	clinicName := strings.TrimSpace(req.ClinicName)
	if clinicName == "" {
		clinicName = "校医务室"
	}
	if prescriptionStatus == "" && req.PrescriptionID != "" {
		prescriptionStatus = PrescriptionReviewApproved
	}
	if doctorName == "" && prescriptionStatus == PrescriptionReviewApproved {
		doctorName = "王医生"
	}
	return &MedicineOrderDetail{
		Order:              *cloneOrder(order),
		Address:            address,
		ContactName:        contactName,
		ContactPhone:       contactPhone,
		ClinicName:         clinicName,
		ClinicLocation:     "综合楼一层",
		DeliveryText:       "校内骑手配送",
		PrescriptionID:     req.PrescriptionID,
		PrescriptionStatus: prescriptionStatus,
		DoctorName:         doctorName,
		Advice:             "请按校医指导用药；如症状加重请及时就医。",
		Items:              append([]MedicineOrderItem{}, items...),
		FeeRows: []MedicineFeeRow{
			{Title: "商品金额", AmountFen: order.ItemsTotalFen},
			{Title: "配送费", AmountFen: order.DeliveryFeeFen},
			{Title: "实付", AmountFen: order.AmountFen},
		},
		Timeline: medicineTimelineForOrder(order),
	}
}

func medicineTimelineForOrder(order *Order) []MedicineTimelineItem {
	if order == nil {
		return nil
	}
	created := order.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	return []MedicineTimelineItem{
		{Title: "订单已提交", Status: ServiceTicketEventDone, Time: created.Format("15:04"), At: created},
		{Title: "校医出药", Status: ServiceTicketEventDone, Time: created.Add(4 * time.Minute).Format("15:04"), At: created.Add(4 * time.Minute)},
		{Title: "骑手已取药", Status: ServiceTicketEventActive, Time: created.Add(12 * time.Minute).Format("15:04"), At: created.Add(12 * time.Minute)},
		{Title: "送达完成", Subtitle: "待完成", Status: ServiceTicketEventPending, At: created.Add(32 * time.Minute)},
	}
}

func cloneWalletWithdrawRequest(input *WalletWithdrawRequest) *WalletWithdrawRequest {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneUserCoupon(input *UserCoupon) *UserCoupon {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePointsTransaction(input *PointsTransaction) *PointsTransaction {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneErrandOrderDetail(input *ErrandOrderDetail) *ErrandOrderDetail {
	if input == nil {
		return nil
	}
	output := *input
	output.Order = *cloneOrder(&input.Order)
	output.FeeRows = append([]ErrandFeeRow{}, input.FeeRows...)
	output.Timeline = append([]ErrandTimelineItem{}, input.Timeline...)
	return &output
}

func (s *Store) nicknameForUserLocked(userID string) string {
	if user := s.users[userID]; user != nil && strings.TrimSpace(user.Nickname) != "" {
		return strings.TrimSpace(user.Nickname)
	}
	if userID == "user_1" {
		return "张三"
	}
	return "悦享用户"
}

func (s *Store) phoneForUserLocked(userID string) string {
	if user := s.users[userID]; user != nil && strings.TrimSpace(user.Phone) != "" {
		return strings.TrimSpace(user.Phone)
	}
	for phone, boundUserID := range s.phoneBindings {
		if boundUserID == userID {
			return phone
		}
	}
	return "13800000000"
}

func avatarInitial(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "悦"
	}
	runes := []rune(name)
	if len(runes) > 2 {
		return string(runes[:2])
	}
	return name
}

func (s *Store) walletBalanceForOverviewLocked(userID string) int64 {
	if account := s.wallets[userID]; account != nil {
		return account.Balance
	}
	if userID == "user_1" {
		return 12850
	}
	return 0
}

func (s *Store) pendingReceivableFenForUserLocked(userID string) int64 {
	if userID == "" {
		return 0
	}
	return 1200
}

func (s *Store) paymentPasswordStatusForUserLocked(userID string) string {
	if s.paymentPasswordHash[userID] != "" || userID == "user_1" {
		return WalletPaymentPasswordSet
	}
	return WalletPaymentPasswordUnset
}

func (s *Store) redPacketCountForUserLocked(userID string) int {
	count := 0
	for _, detail := range s.redPackets {
		if detail == nil {
			continue
		}
		if detail.Packet.SenderID == userID {
			count++
			continue
		}
		for _, share := range detail.Shares {
			if share.UserID == userID {
				count++
				break
			}
		}
	}
	if count == 0 && userID == "user_1" {
		return 3
	}
	return count
}

func (s *Store) userCouponsForUserLocked(userID string) []UserCoupon {
	couponsByID := map[string]UserCoupon{}
	for _, coupon := range defaultUserCoupons(userID) {
		couponCopy := coupon
		couponsByID[couponCopy.ID] = couponCopy
	}
	for _, coupon := range s.userCoupons {
		if coupon != nil && coupon.UserID == userID {
			couponCopy := *cloneUserCoupon(coupon)
			couponsByID[couponCopy.ID] = couponCopy
		}
	}
	coupons := make([]UserCoupon, 0, len(couponsByID))
	for _, coupon := range couponsByID {
		coupons = append(coupons, coupon)
	}
	sort.SliceStable(coupons, func(i, j int) bool {
		return coupons[i].CreatedAt.After(coupons[j].CreatedAt)
	})
	return coupons
}

func (s *Store) walletTransactionsForUserLocked(userID string) []WalletTransaction {
	transactions := make([]WalletTransaction, 0)
	for _, transaction := range s.walletIdempotency {
		if transaction != nil && transaction.UserID == userID {
			transactions = append(transactions, *cloneWalletTransaction(transaction))
		}
	}
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})
	return transactions
}

func defaultWalletTransactions(userID string) []WalletTransaction {
	now := time.Now().UTC()
	return []WalletTransaction{
		{ID: "wtx_preview_refund", UserID: userID, Type: "refund", AmountFen: 1800, PaymentMethod: RefundDestinationBalance, Status: "success", CreatedAt: now.Add(-35 * time.Minute)},
		{ID: "wtx_preview_pay", UserID: userID, Type: "payment", AmountFen: -5598, PaymentMethod: PaymentBalance, Status: "success", CreatedAt: now.Add(-82 * time.Minute)},
		{ID: "wtx_preview_red_packet", UserID: userID, Type: "red_packet", AmountFen: 666, PaymentMethod: PaymentBalance, Status: "success", CreatedAt: now.Add(-16 * time.Hour)},
		{ID: "wtx_preview_withdraw", UserID: userID, Type: "withdraw", AmountFen: -5000, PaymentMethod: "wechat_change", Status: "processing", CreatedAt: now.Add(-25 * time.Hour)},
		{ID: "wtx_preview_credit", UserID: userID, Type: "credit", AmountFen: 10000, PaymentMethod: PaymentWechat, Status: "success", CreatedAt: now.Add(-72 * time.Hour)},
	}
}

func defaultUserCoupons(userID string) []UserCoupon {
	now := time.Now().UTC()
	return []UserCoupon{
		{ID: "coupon_platform_15", UserID: userID, Kind: "platform", Title: "平台外卖通用券", Subtitle: "蓝海餐厅、晴川咖啡等可用", Scope: "外卖", Source: "平台券", Status: "available", ButtonText: "去使用", AccentColor: "#007aff", AmountFen: 1500, ThresholdFen: 5000, ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "coupon_groupbuy_20", UserID: userID, Kind: "groupbuy", Title: "到店团购立减券", Subtitle: "周末探店可用", Scope: "团购", Source: "团购券", Status: "available", ButtonText: "去使用", AccentColor: "#ff3b30", AmountFen: 2000, ThresholdFen: 0, ExpiresAt: now.Add(7 * 24 * time.Hour), CreatedAt: now.Add(-24 * time.Hour)},
		{ID: "coupon_medicine_5", UserID: userID, Kind: "medicine", Title: "安心药房夜间券", Subtitle: "买药频道可用", Scope: "买药", Source: "买药券", Status: "available", ButtonText: "去使用", AccentColor: "#16a34a", AmountFen: 500, ThresholdFen: 2500, ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now.Add(-12 * time.Hour)},
	}
}

func merchantGroupCouponID(userID string) string {
	return "coupon_group_8_" + strings.TrimSpace(userID)
}

func merchantGroupCoupon(userID string, now time.Time) *UserCoupon {
	return &UserCoupon{
		ID:           merchantGroupCouponID(userID),
		UserID:       userID,
		Kind:         "merchant",
		Title:        "蓝海餐厅商户群券",
		Subtitle:     "进群领取 · 外卖专享",
		Scope:        "外卖",
		Source:       "商户群券",
		Status:       "available",
		ButtonText:   "去使用",
		AccentColor:  "#ff5a1f",
		AmountFen:    800,
		ThresholdFen: 3000,
		ExpiresAt:    now.Add(8 * time.Hour),
		CreatedAt:    now,
	}
}

func (s *Store) claimMerchantGroupCouponLocked(userID string, code string) (*UserCoupon, bool, error) {
	if strings.TrimSpace(code) != chatThreadSelfServePolicy("merchant_blue_sea").CouponCode {
		return nil, false, nil
	}
	if s.chatThreadMemberLocked("merchant_blue_sea", "user", userID) == nil {
		return nil, true, fmt.Errorf("%w: group membership required", ErrInvalidArgument)
	}
	if existing := s.userCoupons[merchantGroupCouponID(userID)]; existing != nil {
		return cloneUserCoupon(existing), true, nil
	}
	coupon := merchantGroupCoupon(userID, time.Now().UTC())
	s.userCoupons[coupon.ID] = cloneUserCoupon(coupon)
	return cloneUserCoupon(coupon), true, nil
}

func (s *Store) pointsBalanceForUserLocked(userID string) int {
	total := 0
	for _, transaction := range s.pointsTransactions[userID] {
		if transaction != nil {
			total += transaction.Points
		}
	}
	if total == 0 && userID == "user_1" {
		return 2680
	}
	return total
}

func (s *Store) pointsSummaryLocked(userID string) *PointsSummary {
	transactions := make([]PointsTransaction, 0)
	for _, transaction := range s.pointsTransactions[userID] {
		if transaction != nil {
			transactions = append(transactions, *clonePointsTransaction(transaction))
		}
	}
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})
	if len(transactions) == 0 && userID == "user_1" {
		transactions = append(transactions, seedPointTransactionsForUser(userID)...)
	}
	return &PointsSummary{
		UserID:          userID,
		Nickname:        s.nicknameForUserLocked(userID),
		MembershipLevel: MembershipSilver,
		LevelName:       "美食达人",
		Verified:        true,
		Points:          s.pointsBalanceForUserLocked(userID),
		GrowthValue:     2680,
		NextLevelGrowth: 1320,
		Benefits: []PointsBenefit{
			{Icon: "券", Title: "专属优惠券", Status: "已解锁", Unlocked: true},
			{Icon: "抵", Title: "积分抵扣", Status: "已解锁", Unlocked: true},
			{Icon: "礼", Title: "生日礼", Status: "V3 解锁"},
			{Icon: "客", Title: "优先客服", Status: "V3 解锁"},
		},
		Tasks: []PointsTask{
			{ID: "order", Title: "完成订单", Subtitle: "每完成一笔订单", Reward: 30, ActionText: "去下单", Route: "/pages/index/index"},
			{ID: "review", Title: "评价订单", Subtitle: "每完成一笔评价", Reward: 10, ActionText: "去评价", Route: "/pages/order/list/index"},
			{ID: "invite", Title: "邀请好友", Subtitle: "好友下单后可得奖励", Reward: 100, ActionText: "去邀请", Route: "/pages/invite-friends/index"},
			{ID: "checkin", Title: "每日签到", Subtitle: "连续签到积分更多", Reward: 5, ActionText: "签到"},
		},
		Rewards: []PointsReward{
			{ID: "coupon_5", Title: "兑 ¥5 优惠券", Subtitle: "满 30 元可用", Points: 500, AmountFen: 500, AccentColor: "#007aff", RedeemText: "兑换", ThresholdFen: 3000},
			{ID: "delivery_12", Title: "兑换配送券", Subtitle: "满 20 元可用", Points: 1200, AmountFen: 300, AccentColor: "#ff7a00", RedeemText: "兑换", ThresholdFen: 2000},
		},
		Transactions: transactions,
	}
}

func seedPointTransactionsForUser(userID string) []PointsTransaction {
	now := time.Now().UTC()
	return []PointsTransaction{
		{ID: "pt_seed_order", UserID: userID, Type: PointsTransactionEarn, Title: "订单完成奖励", Points: 30, SourceID: "order_seed", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "pt_seed_review", UserID: userID, Type: PointsTransactionEarn, Title: "评价订单奖励", Points: 10, SourceID: "review_seed", CreatedAt: now.Add(-24 * time.Hour)},
		{ID: "pt_seed_redeem", UserID: userID, Type: PointsTransactionRedeem, Title: "兑换优惠券", Points: -500, SourceID: "reward_seed", CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "pt_seed_invite", UserID: userID, Type: PointsTransactionEarn, Title: "邀请好友奖励", Points: 100, SourceID: "invite_seed", CreatedAt: now.Add(-120 * time.Hour)},
	}
}

func inviteCodeForUser(userID string) string {
	if userID == "user_1" {
		return "YXES-8K29"
	}
	hash := strings.ToUpper(shortHash(userID))
	return "YXES-" + hash[:4]
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func buttonTextForSearchType(resultType string) string {
	switch resultType {
	case "medicine":
		return "去买药"
	case "errand":
		return "去下单"
	case "groupbuy":
		return "去购买"
	case "product":
		return "加入购物车"
	default:
		return "去使用"
	}
}

func filterSearchResults(results []SearchResult, keyword string, category string) []SearchResult {
	filtered := make([]SearchResult, 0, len(results))
	for _, result := range results {
		if category != "" && category != "all" && result.Type != category {
			continue
		}
		if keyword != "" && !strings.Contains(result.Title+result.Subtitle+result.Badge, keyword) {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func searchSuggestionsFromResults(results []SearchResult) []string {
	suggestions := make([]string, 0, 6)
	seen := map[string]struct{}{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || len(suggestions) >= 6 {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		suggestions = append(suggestions, value)
	}
	for _, result := range results {
		add(result.Title)
		add(result.Badge)
		if len(suggestions) >= 6 {
			break
		}
	}
	return suggestions
}

func isErrandOrderType(orderType string) bool {
	switch orderType {
	case OrderTypeErrandBuy, OrderTypeErrandDeliver, OrderTypeErrandPickup, OrderTypeErrandDo:
		return true
	default:
		return false
	}
}

func errandServiceTitle(orderType string) string {
	switch orderType {
	case OrderTypeErrandBuy:
		return "帮买"
	case OrderTypeErrandDeliver:
		return "帮送"
	case OrderTypeErrandDo:
		return "帮办"
	default:
		return "帮取"
	}
}

func (s *Store) errandDetailForOrderLocked(order *Order, req ErrandOrderRequest) *ErrandOrderDetail {
	now := order.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	baseFee := int64(1000)
	distanceFee := int64(400)
	serviceFee := int64(200)
	coupon := req.CouponAmountFen
	if coupon == 0 {
		coupon = 300
	}
	return &ErrandOrderDetail{
		Order:           *cloneOrder(order),
		ServiceType:     order.Type,
		ServiceTitle:    errandServiceTitle(order.Type),
		PickupAddress:   defaultString(req.PickupAddress, "望京SOHO B座 快递柜"),
		DeliveryAddress: defaultString(req.DeliveryAddress, "望京SOHO A座 1208"),
		ContactName:     defaultString(req.ContactName, "张三"),
		ContactPhone:    defaultString(req.ContactPhone, "13800000000"),
		ItemType:        defaultString(req.ItemType, "小件包裹"),
		Description:     defaultString(req.Description, "取 3 号柜快递，验证码已备注"),
		ImageURL:        defaultString(req.ImageURL, errandImageURL()),
		WeightText:      defaultString(req.WeightText, "2kg 内"),
		PickupTime:      defaultString(req.PickupTime, "立即取送"),
		EstimateText:    "预计 14:25 送达",
		MapStatus:       "骑手正在前往取件地址",
		Rider:           ErrandRider{ID: "rider_zhang", Name: "张师傅", RatingText: "4.9", Vehicle: "电动车", DistanceText: "距取件地 600m"},
		FeeRows: []ErrandFeeRow{
			{Title: "起步价", AmountFen: baseFee},
			{Title: "距离费", AmountFen: distanceFee},
			{Title: "服务费", AmountFen: serviceFee},
			{Title: "优惠券", AmountFen: -coupon},
		},
		Timeline: []ErrandTimelineItem{
			{Title: "订单已创建", Status: "done", Time: "14:02", At: now},
			{Title: "骑手已接单", Status: "done", Time: "14:04", At: now.Add(2 * time.Minute)},
			{Title: "前往取件", Subtitle: "骑手正在前往取件地址", Status: "active", Time: "进行中", At: now.Add(4 * time.Minute)},
			{Title: "已取件", Subtitle: "待骑手取件", Status: "pending"},
			{Title: "送达完成", Subtitle: "请确认收货并评价", Status: "pending"},
		},
	}
}

func errandImageURL() string {
	return "/assets/generated/errand-parcel.jpg"
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func cloneRefundTransaction(input *RefundTransaction) *RefundTransaction {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneAfterSalesRequest(input *AfterSalesRequest) *AfterSalesRequest {
	if input == nil {
		return nil
	}
	output := *input
	output.EvidenceURLs = append([]string{}, input.EvidenceURLs...)
	return &output
}

func cloneAfterSalesEvent(input *AfterSalesEvent) *AfterSalesEvent {
	if input == nil {
		return nil
	}
	output := *input
	output.Attachments = append([]string{}, input.Attachments...)
	return &output
}

func cloneAfterSalesEvidence(input *AfterSalesEvidence) *AfterSalesEvidence {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneAfterSalesEvidenceUploadTicket(input *AfterSalesEvidenceUploadTicket) *AfterSalesEvidenceUploadTicket {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePaymentTransaction(input *PaymentTransaction) *PaymentTransaction {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneDispatchEvent(input *DispatchEvent) *DispatchEvent {
	if input == nil {
		return nil
	}
	output := *input
	output.RejectedRiderIDs = append([]string{}, input.RejectedRiderIDs...)
	return &output
}

func cloneOutboxEvent(input *OutboxEvent) *OutboxEvent {
	if input == nil {
		return nil
	}
	output := *input
	output.Payload = cloneMapAny(input.Payload)
	return &output
}

func clonePlatformNotification(input *PlatformNotification) *PlatformNotification {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func clonePlatformNotificationDelivery(input *PlatformNotificationDelivery) *PlatformNotificationDelivery {
	if input == nil {
		return nil
	}
	output := *input
	return &output
}

func cloneNotificationPreference(input *NotificationPreference) *NotificationPreference {
	if input == nil {
		return nil
	}
	output := *input
	output.EnabledChannels = append([]string{}, input.EnabledChannels...)
	output.DisabledChannels = append([]string{}, input.DisabledChannels...)
	output.QuietHours = cloneNotificationQuietHours(input.QuietHours)
	return &output
}

func cloneNotificationQuietHours(input NotificationQuietHours) NotificationQuietHours {
	output := input
	output.Channels = append([]string{}, input.Channels...)
	output.ExemptTypes = append([]string{}, input.ExemptTypes...)
	return output
}

func cloneAuditLog(input *AuditLog) *AuditLog {
	if input == nil {
		return nil
	}
	output := *input
	output.Payload = cloneMapAny(input.Payload)
	return &output
}

func cloneMapAny(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = cloneAny(value)
	}
	return output
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMapAny(typed)
	case []any:
		output := make([]any, len(typed))
		for index, item := range typed {
			output[index] = cloneAny(item)
		}
		return output
	case []string:
		return append([]string{}, typed...)
	default:
		return value
	}
}

func seedUserCoupons() map[string]*UserCoupon {
	coupons := defaultUserCoupons("user_1")
	output := make(map[string]*UserCoupon, len(coupons))
	for index := range coupons {
		coupon := coupons[index]
		output[coupon.ID] = cloneUserCoupon(&coupon)
	}
	return output
}

func seedPointsTransactions() map[string][]*PointsTransaction {
	transactions := seedPointTransactionsForUser("user_1")
	output := map[string][]*PointsTransaction{"user_1": {}}
	for index := range transactions {
		transaction := transactions[index]
		output["user_1"] = append(output["user_1"], clonePointsTransaction(&transaction))
	}
	output["user_1"] = append(output["user_1"], &PointsTransaction{
		ID:        "pt_seed_base",
		UserID:    "user_1",
		Type:      PointsTransactionEarn,
		Title:     "会员成长值结转",
		Points:    3040,
		SourceID:  "membership_seed",
		CreatedAt: time.Now().UTC().Add(-30 * 24 * time.Hour),
	})
	return output
}

func seedCirclePosts() map[string]*CirclePost {
	now := time.Now().UTC()
	return map[string]*CirclePost{
		"cpost_seed_1": {
			ID:           "cpost_seed_1",
			AuthorUserID: "user_1",
			AuthorName:   "小林",
			CircleID:     "nearby",
			Type:         CirclePostFoodInvite,
			Title:        "午饭拼单有人吗",
			Content:      "蓝海餐厅满减差一份，12:20 前下单。",
			Status:       CirclePostPublished,
			Tags:         []string{"附近 500m", "拼单"},
			DistanceText: "附近 500m",
			LikeCount:    8,
			CommentCount: 3,
			CreatedAt:    now.Add(-18 * time.Minute),
		},
		"cpost_seed_2": {
			ID:           "cpost_seed_2",
			AuthorUserID: "official",
			AuthorName:   "悦享e食官方群",
			CircleID:     "official",
			Type:         CirclePostText,
			Title:        "新用户入群默认静音",
			Content:      "重要通知会通过站内信保留，不会打扰夜间休息。",
			Status:       CirclePostPublished,
			Tags:         []string{"官方"},
			LikeCount:    23,
			CommentCount: 5,
			CreatedAt:    now.Add(-42 * time.Minute),
		},
		"cpost_seed_3": {
			ID:           "cpost_seed_3",
			AuthorUserID: "merchant_1",
			AuthorName:   "蓝海餐厅商户群",
			CircleID:     "merchant_blue_sea",
			Type:         CirclePostText,
			Title:        "群内团购券限时领取",
			Content:      "加入商户群后可领取指定优惠券，今晚 20:00 前可用。",
			Status:       CirclePostPublished,
			Tags:         []string{"商户群", "团购券"},
			LikeCount:    16,
			CommentCount: 2,
			CreatedAt:    now.Add(-76 * time.Minute),
		},
	}
}

func seedMealMatchProfiles() map[string]*MealMatchProfile {
	checkedAt := time.Now().UTC().Add(-24 * time.Hour)
	return map[string]*MealMatchProfile{
		"user_buddy_lunch": {
			UserID:                         "user_buddy_lunch",
			Gender:                         "female",
			SchoolID:                       "infinitech_university",
			SchoolName:                     "无限科技大学",
			CampusName:                     "东区",
			BuildingID:                     "east_canteen",
			BuildingName:                   "东区食堂",
			PrivacyScope:                   MealMatchPrivacySameBuilding,
			LocationPrecision:              MealMatchLocationBuildingOnly,
			IdentityTruthSigned:            true,
			PlatformLiabilityReleaseSigned: true,
			QuestionnaireCompleted:         true,
			PersonalityTraits:              []string{"细心", "守时", "安静"},
			DietaryHabits:                  []string{"清淡", "不浪费", "咖啡"},
			DeviceID:                       "seed_device_lunch",
			DeviceRiskState:                MealMatchDeviceRiskPassed,
			DeviceRiskReason:               "种子设备已通过校验",
			DeviceRiskCheckedAt:            checkedAt,
			ModerationStatus:               MealMatchModerationApproved,
			ModerationReason:               "种子资料已通过审核",
		},
		"user_buddy_weekend": {
			UserID:                         "user_buddy_weekend",
			Gender:                         "male",
			SchoolID:                       "city_college",
			SchoolName:                     "城市学院",
			CampusName:                     "主校区",
			BuildingID:                     "west_gate",
			BuildingName:                   "西门生活区",
			PrivacyScope:                   MealMatchPrivacySameSchool,
			LocationPrecision:              MealMatchLocationCampusOnly,
			IdentityTruthSigned:            true,
			PlatformLiabilityReleaseSigned: true,
			QuestionnaireCompleted:         true,
			PersonalityTraits:              []string{"外向", "守时"},
			DietaryHabits:                  []string{"火锅", "烤肉", "不浪费"},
			DeviceID:                       "seed_device_weekend",
			DeviceRiskState:                MealMatchDeviceRiskPassed,
			DeviceRiskReason:               "种子设备已通过校验",
			DeviceRiskCheckedAt:            checkedAt,
			ModerationStatus:               MealMatchModerationApproved,
			ModerationReason:               "种子资料已通过审核",
		},
		"user_buddy_library": {
			UserID:                         "user_buddy_library",
			Gender:                         "female",
			SchoolID:                       "infinitech_university",
			SchoolName:                     "无限科技大学",
			CampusName:                     "东区",
			BuildingID:                     "east_canteen",
			BuildingName:                   "东区食堂",
			PrivacyScope:                   MealMatchPrivacySameBuilding,
			LocationPrecision:              MealMatchLocationBuildingOnly,
			IdentityTruthSigned:            true,
			PlatformLiabilityReleaseSigned: true,
			QuestionnaireCompleted:         true,
			PersonalityTraits:              []string{"安静", "计划感"},
			DietaryHabits:                  []string{"清淡", "面食"},
			DeviceID:                       "seed_device_library",
			DeviceRiskState:                MealMatchDeviceRiskPassed,
			DeviceRiskReason:               "种子设备已通过校验",
			DeviceRiskCheckedAt:            checkedAt,
			ModerationStatus:               MealMatchModerationApproved,
			ModerationReason:               "种子资料已通过审核",
		},
	}
}

func seedChatMessages() map[string]*ChatMessage {
	now := time.Now().UTC()
	return map[string]*ChatMessage{
		"msg_seed_official_1": {
			ID:          "msg_seed_official_1",
			ThreadID:    "official",
			SenderID:    "official",
			Sender:      "系统",
			Content:     "新用户入群默认静默，重要通知站内信保留。",
			MessageType: "text",
			CreatedAt:   now.Add(-20 * time.Minute),
		},
		"msg_seed_merchant_1": {
			ID:          "msg_seed_merchant_1",
			ThreadID:    "merchant_blue_sea",
			SenderID:    "merchant_1",
			Sender:      "蓝海餐厅",
			Content:     "今晚 20:00 前可领商户群券，下单可叠加平台满减。",
			MessageType: "text",
			CreatedAt:   now.Add(-52 * time.Minute),
		},
		"msg_seed_merchant_2": {
			ID:          "msg_seed_merchant_2",
			ThreadID:    "merchant_blue_sea",
			SenderID:    "user_group_xiaolin",
			Sender:      "小林",
			Content:     "午饭拼单还差一份，有人一起吗？",
			MessageType: "text",
			CreatedAt:   now.Add(-46 * time.Minute),
		},
		"msg_seed_merchant_3": {
			ID:          "msg_seed_merchant_3",
			ThreadID:    "merchant_blue_sea",
			SenderID:    "user_group_ajie",
			Sender:      "阿杰",
			Content:     "我也拼一份，12:20 前下单可以吗？",
			MessageType: "text",
			CreatedAt:   now.Add(-44 * time.Minute),
		},
		"msg_seed_rider_1": {
			ID:          "msg_seed_rider_1",
			ThreadID:    "rider_zhang",
			SenderID:    "rider_1",
			Sender:      "骑手 张师傅",
			Content:     "餐品已送达，记得给本次配送评价。",
			MessageType: "text",
			CreatedAt:   now.Add(-26 * time.Hour),
		},
	}
}

func seedShops() map[string]*Shop {
	return map[string]*Shop{
		"shop_1": {
			ID:             "shop_1",
			MerchantID:     "merchant_1",
			StationID:      "station_1",
			Name:           "蓝海餐厅",
			Category:       "restaurant",
			AccountType:    MerchantAccountStandard,
			Status:         ShopStatusActive,
			Capabilities:   []string{ShopCapabilityTakeout, ShopCapabilityGroupbuy},
			Qualifications: []string{QualificationBusinessLicense, QualificationHealthCertificate},
			CoverURL:       "/assets/generated/shop-detail-cover.jpg",
			LogoURL:        "/assets/brand/logo.svg",
			Announcement:   "主打简餐和团购套餐，当前为 2.0 闭环样例店铺。",
		},
	}
}

func seedMerchants() map[string]*MerchantAccount {
	return map[string]*MerchantAccount{
		"merchant_1": {
			ID:            "merchant_1",
			Type:          MerchantAccountStandard,
			DisplayName:   "蓝海餐厅",
			Status:        ShopStatusActive,
			DepositStatus: DepositStatusPaid,
		},
	}
}

func seedMerchantQualifications() map[string][]*MerchantQualification {
	expiresAt := time.Now().UTC().Add(365 * 24 * time.Hour)
	return map[string][]*MerchantQualification{
		"merchant_1": {
			{
				ID:        "mq_merchant_1_license",
				Type:      QualificationBusinessLicense,
				FileURL:   "/assets/mock/business-license.jpg",
				ExpiresAt: expiresAt,
				Status:    QualificationStatusApproved,
			},
			{
				ID:        "mq_merchant_1_health",
				Type:      QualificationHealthCertificate,
				FileURL:   "/assets/mock/health-certificate.jpg",
				ExpiresAt: expiresAt,
				Status:    QualificationStatusApproved,
			},
		},
	}
}

func seedMerchantStaff() map[string][]*MerchantStaff {
	expiresAt := time.Now().UTC().Add(365 * 24 * time.Hour)
	return map[string][]*MerchantStaff{
		"merchant_1": {
			{
				ID:                         "staff_merchant_1_owner",
				MerchantID:                 "merchant_1",
				ShopID:                     "shop_1",
				Name:                       "蓝海店长",
				Phone:                      "13800000001",
				Role:                       "manager",
				Status:                     MerchantStaffActive,
				HealthCertificateURL:       "/assets/mock/staff-health.jpg",
				HealthCertificateExpiresAt: expiresAt,
			},
		},
	}
}

func seedMerchantMaterials() map[string][]*MerchantSupplementalMaterial {
	now := time.Now().UTC()
	return map[string][]*MerchantSupplementalMaterial{
		"merchant_1": {
			{
				ID:          "material_merchant_1_storefront",
				MerchantID:  "merchant_1",
				ShopID:      "shop_1",
				Type:        "storefront_photo",
				FileURL:     "/assets/mock/storefront.jpg",
				Description: "门头照",
				Status:      "approved",
				UploadedAt:  now,
			},
		},
	}
}

func seedRiders() map[string]*RiderAccount {
	return map[string]*RiderAccount{
		"station_manager_1": {
			ID:        "station_manager_1",
			StationID: "station_1",
			Type:      RiderAccountStationManager,
			Status:    "active",
			Online:    true,
		},
		"rider_1": {
			ID:                   "rider_1",
			StationID:            "station_1",
			Type:                 RiderAccountRider,
			Status:               "active",
			Online:               true,
			DepositStatus:        DepositStatusPaid,
			Capacity:             3,
			DispatchPriority:     RiderDispatchPriority(RiderLevelA),
			AverageAcceptSeconds: 18,
			AverageDailyOrders:   36,
			CompletionRate:       0.98,
			DistanceMeters:       800,
		},
		"rider_2": {
			ID:                   "rider_2",
			StationID:            "station_1",
			Type:                 RiderAccountRider,
			Status:               "active",
			Online:               true,
			DepositStatus:        DepositStatusPaid,
			Capacity:             2,
			DispatchPriority:     RiderDispatchPriority(RiderLevelB),
			AverageAcceptSeconds: 12,
			AverageDailyOrders:   24,
			CompletionRate:       0.95,
			DistanceMeters:       300,
		},
	}
}

func seedDeposits() map[string]*DepositAccount {
	now := time.Now().UTC()
	return map[string]*DepositAccount{
		depositKey("merchant", "merchant_1"): {
			SubjectType: "merchant",
			SubjectID:   "merchant_1",
			AmountFen:   MerchantDepositAmountFen,
			Status:      DepositStatusPaid,
			UpdatedAt:   now,
		},
		depositKey("rider", "rider_1"): {
			SubjectType: "rider",
			SubjectID:   "rider_1",
			AmountFen:   RiderDepositAmountFen,
			Status:      DepositStatusPaid,
			UpdatedAt:   now,
		},
		depositKey("rider", "rider_2"): {
			SubjectType: "rider",
			SubjectID:   "rider_2",
			AmountFen:   RiderDepositAmountFen,
			Status:      DepositStatusPaid,
			UpdatedAt:   now,
		},
	}
}

func seedStationTaskConfigs() map[string]*StationTaskConfig {
	return map[string]*StationTaskConfig{
		"station_1": {
			StationID:                    "station_1",
			ConfiguredByStationManagerID: "station_manager_1",
			DailyTaskDurationMinutes:     8 * 60,
			DailyFixedOrderCount:         30,
		},
	}
}

func seedStationServiceAreas() map[string]*StationServiceArea {
	return map[string]*StationServiceArea{
		"station_1": {
			StationID: "station_1",
			ShopIDs:   []string{"shop_1"},
		},
	}
}

func seedProducts() map[string]*MerchantProduct {
	return map[string]*MerchantProduct{
		"prod_beef_rice": {
			ID:             "prod_beef_rice",
			ShopID:         "shop_1",
			Name:           "招牌牛肉饭",
			ImageURL:       "/assets/generated/product-beef-rice.jpg",
			Description:    "牛肉、米饭、时蔬，适合作为外卖闭环样例。",
			IngredientList: []string{"牛肉", "米饭", "青菜"},
			PriceFen:       2599,
			StockCount:     50,
			Status:         ProductStatusActive,
		},
		"prod_soup": {
			ID:             "prod_soup",
			ShopID:         "shop_1",
			Name:           "每日例汤",
			ImageURL:       "/assets/generated/product-lemon-tea.jpg",
			Description:    "随餐热汤。",
			IngredientList: []string{"汤底", "蔬菜"},
			PriceFen:       599,
			StockCount:     80,
			Status:         ProductStatusActive,
		},
	}
}

func seedGroupbuyDeals() map[string]*MerchantProduct {
	return map[string]*MerchantProduct{
		"deal_two_person_set": {
			ID:             "deal_two_person_set",
			ShopID:         "shop_1",
			Name:           "双人工作餐团购券",
			ImageURL:       "/assets/generated/home-featured-dish.jpg",
			Description:    "到店扫码核销，含两份主食和两份例汤。",
			IngredientList: []string{"主食", "例汤", "到店核销"},
			PriceFen:       3999,
			StockCount:     100,
			Status:         ProductStatusActive,
		},
	}
}
