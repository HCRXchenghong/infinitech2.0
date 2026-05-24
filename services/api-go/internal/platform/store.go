package platform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
	ErrOrderAlreadyAssigned = errors.New("order already assigned")
	ErrInvalidOrderState    = errors.New("invalid order state")
	ErrPaymentPassword      = errors.New("payment password required or invalid")
	ErrInvalidCredentials   = errors.New("invalid credentials")
)

type Store struct {
	mu                      sync.Mutex
	nextOrderID             uint64
	nextTransactionID       uint64
	nextAddressID           uint64
	nextMerchantID          uint64
	nextMerchantStaffID     uint64
	nextMerchantMaterialID  uint64
	nextDispatchEventID     uint64
	nextOutboxEventID       uint64
	nextAuditLogID          uint64
	nextAfterSalesID        uint64
	nextAfterSalesEventID   uint64
	nextRiderID             uint64
	nextProductID           uint64
	nextVoucherID           uint64
	homeModules             []HomeModule
	homeCards               []HomeCard
	users                   map[string]*AppUser
	wechatBindings          map[string]string
	merchantInvites         map[string]*MerchantOnboardingInvite
	merchants               map[string]*MerchantAccount
	merchantQualifications  map[string][]*MerchantQualification
	merchantStaff           map[string][]*MerchantStaff
	merchantMaterials       map[string][]*MerchantSupplementalMaterial
	riders                  map[string]*RiderAccount
	deposits                map[string]*DepositAccount
	stationTaskConfigs      map[string]*StationTaskConfig
	stationServiceAreas     map[string]*StationServiceArea
	shops                   map[string]*Shop
	products                map[string]*MerchantProduct
	groupbuyDeals           map[string]*MerchantProduct
	addresses               map[string][]*UserAddress
	cartItems               map[string][]*CartItem
	orders                  map[string]*Order
	wallets                 map[string]*WalletAccount
	paymentPasswordHash     map[string]string
	merchantPasswordHash    map[string]string
	riderPasswordHash       map[string]string
	paymentTransactions     map[string]*PaymentTransaction
	paymentByTradeNo        map[string]*PaymentTransaction
	paymentByProviderID     map[string]*PaymentTransaction
	walletIdempotency       map[string]*WalletTransaction
	refundSettings          RefundSettings
	refundTransactions      map[string]*RefundTransaction
	refundByIdempotency     map[string]string
	afterSalesRequests      map[string]*AfterSalesRequest
	afterSalesEvents        map[string]*AfterSalesEvent
	afterSalesUploadTickets map[string]*AfterSalesEvidenceUploadTicket
	afterSalesEvidence      map[string]*AfterSalesEvidence
	groupbuyVouchers        map[string]*GroupbuyVoucher
	vouchersByOrderID       map[string][]string
	vouchersByCode          map[string]*GroupbuyVoucher
	dispatchEvents          map[string]*DispatchEvent
	dispatchRejectedRiders  map[string]map[string]bool
	freeCancelUsedByDate    map[string]string
	outboxEvents            map[string]*OutboxEvent
	outboxByIdempotency     map[string]string
	auditLogs               map[string]*AuditLog
	auditLogSigningSecret   string
	objectStorage           ObjectStorageConfig
}

func NewStore(homeModules []HomeModule) *Store {
	return &Store{
		nextMerchantID:          1,
		nextRiderID:             2,
		nextProductID:           2,
		homeModules:             cloneHomeModules(homeModules),
		homeCards:               DefaultHomeCards(),
		users:                   map[string]*AppUser{},
		wechatBindings:          map[string]string{},
		merchantInvites:         map[string]*MerchantOnboardingInvite{},
		merchants:               seedMerchants(),
		merchantQualifications:  seedMerchantQualifications(),
		merchantStaff:           seedMerchantStaff(),
		merchantMaterials:       seedMerchantMaterials(),
		riders:                  seedRiders(),
		deposits:                seedDeposits(),
		stationTaskConfigs:      seedStationTaskConfigs(),
		stationServiceAreas:     seedStationServiceAreas(),
		shops:                   seedShops(),
		products:                seedProducts(),
		groupbuyDeals:           seedGroupbuyDeals(),
		addresses:               map[string][]*UserAddress{},
		cartItems:               map[string][]*CartItem{},
		orders:                  map[string]*Order{},
		wallets:                 map[string]*WalletAccount{},
		paymentPasswordHash:     map[string]string{},
		merchantPasswordHash:    map[string]string{},
		riderPasswordHash:       map[string]string{},
		paymentTransactions:     map[string]*PaymentTransaction{},
		paymentByTradeNo:        map[string]*PaymentTransaction{},
		paymentByProviderID:     map[string]*PaymentTransaction{},
		walletIdempotency:       map[string]*WalletTransaction{},
		refundSettings:          RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst},
		refundTransactions:      map[string]*RefundTransaction{},
		refundByIdempotency:     map[string]string{},
		afterSalesRequests:      map[string]*AfterSalesRequest{},
		afterSalesEvents:        map[string]*AfterSalesEvent{},
		afterSalesUploadTickets: map[string]*AfterSalesEvidenceUploadTicket{},
		afterSalesEvidence:      map[string]*AfterSalesEvidence{},
		groupbuyVouchers:        map[string]*GroupbuyVoucher{},
		vouchersByOrderID:       map[string][]string{},
		vouchersByCode:          map[string]*GroupbuyVoucher{},
		dispatchEvents:          map[string]*DispatchEvent{},
		dispatchRejectedRiders:  map[string]map[string]bool{},
		freeCancelUsedByDate:    map[string]string{},
		outboxEvents:            map[string]*OutboxEvent{},
		outboxByIdempotency:     map[string]string{},
		auditLogs:               map[string]*AuditLog{},
		objectStorage:           DefaultObjectStorageConfig(),
	}
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
		Status:    "approved",
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
			orders = append(orders, *cloneOrder(order))
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
	return cloneOrder(order), nil
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
	if s.shops[shopID] == nil {
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
			UnitPriceFen: item.UnitPriceFen,
			Quantity:     item.Quantity,
		})
	}
	order := &Order{
		ID:              fmt.Sprintf("ord_%d", s.nextOrderID),
		UserID:          userID,
		ShopID:          shopID,
		AddressID:       addressID,
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

	riderPerformance, err := s.riderPerformanceSnapshotLocked(req.StationManagerID)
	if err != nil {
		return nil, err
	}
	if len(riderPerformance) > limit {
		riderPerformance = riderPerformance[:limit]
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
	ticket := s.afterSalesUploadTickets[req.TicketID]
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
	ticket := s.afterSalesUploadTickets[req.TicketID]
	if ticket == nil || ticket.ObjectKey != req.ObjectKey {
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

func (s *Store) UserAfterSalesRequests(userID string) ([]AfterSalesRequest, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	requests := make([]AfterSalesRequest, 0)
	for _, request := range s.afterSalesRequests {
		if request == nil || request.UserID != userID {
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

func (s *Store) AdminAfterSalesRequests() ([]AfterSalesRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	requests := make([]AfterSalesRequest, 0, len(s.afterSalesRequests))
	for _, request := range s.afterSalesRequests {
		if request == nil {
			continue
		}
		requests = append(requests, *s.afterSalesRequestViewLocked(request))
	}
	sortAfterSalesRequests(requests)
	return requests, nil
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

	performances := make([]RiderPerformance, 0)
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
		performances = append(performances, RiderPerformance{
			RiderID:              rider.ID,
			StationID:            rider.StationID,
			AverageAcceptSeconds: rider.AverageAcceptSeconds,
			AverageDailyOrders:   averageDailyOrders,
			CompletionRate:       completionRate,
		})
	}

	teamAverageAcceptSeconds := averagePositiveAcceptSeconds(performances)
	teamAverageDailyOrders := averagePositiveDailyOrders(performances)
	for index := range performances {
		performance := &performances[index]
		performance.Score, performance.Level, performance.DispatchPriority = evaluateRiderPerformanceLevel(*performance, teamAverageAcceptSeconds, teamAverageDailyOrders)
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
		if performances[i].AverageAcceptSeconds != performances[j].AverageAcceptSeconds {
			return performances[i].AverageAcceptSeconds < performances[j].AverageAcceptSeconds
		}
		return performances[i].RiderID < performances[j].RiderID
	})
	return performances, nil
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
	defaultAuditRetentionDays        = 2555
	defaultAuditHotDays              = 180
	defaultAuditIntegritySampleLimit = 500
	maxAuditIntegritySampleLimit     = 5000
	auditRetentionAlertTopic         = "audit.retention_alerts"
	defaultAuditArchiveLimit         = 500
	maxAuditArchiveLimit             = 5000
	defaultAuditArchiveStoragePrefix = "worm://audit-logs"
	auditArchiveManifestAlgorithm    = "sha256:v1"
	auditArchiveRequestedTopic       = "audit.archive_requested"
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

var auditPayloadAllowlist = map[string]struct{}{
	"action_filter":           {},
	"alert_count":             {},
	"archive_id":              {},
	"actor_id":                {},
	"actor_type":              {},
	"after":                   {},
	"amount_fen":              {},
	"applied_scopes":          {},
	"attempts":                {},
	"before":                  {},
	"change_request_id":       {},
	"changed":                 {},
	"claimed":                 {},
	"cleanup_attempts":        {},
	"cold_archive_cutoff":     {},
	"cold_archive_due_logs":   {},
	"compensation_type":       {},
	"critical_count":          {},
	"current_scopes":          {},
	"decision":                {},
	"default_refund_strategy": {},
	"destination":             {},
	"evidence_count":          {},
	"expected_rider_id":       {},
	"expected_status":         {},
	"export_format":           {},
	"expires_at":              {},
	"expired_logs":            {},
	"generated_at":            {},
	"hot_days":                {},
	"idempotency_key":         {},
	"integrity_failures":      {},
	"lease_owner":             {},
	"lease_seconds":           {},
	"limit":                   {},
	"log_count":               {},
	"manifest_algorithm":      {},
	"manifest_hash":           {},
	"object_key":              {},
	"outbox_event_id":         {},
	"previous_scopes":         {},
	"previous_rider_id":       {},
	"previous_status":         {},
	"policy_version":          {},
	"reason":                  {},
	"refund_id":               {},
	"retention_days":          {},
	"replayed":                {},
	"requested_scopes":        {},
	"retry_after_seconds":     {},
	"role":                    {},
	"rollback_from_scopes":    {},
	"rollback_to_scopes":      {},
	"row_count":               {},
	"station_id":              {},
	"status":                  {},
	"storage_key":             {},
	"storage_prefix":          {},
	"topic":                   {},
	"type":                    {},
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

func outboxLeaseActive(event *OutboxEvent, now time.Time) bool {
	if event == nil || strings.TrimSpace(event.LeaseOwner) == "" || event.LeaseExpiresAt.IsZero() {
		return false
	}
	return event.LeaseExpiresAt.UTC().After(now.UTC())
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

func evaluateRiderPerformanceLevel(performance RiderPerformance, teamAverageAcceptSeconds float64, teamAverageDailyOrders float64) (int, string, int) {
	acceptScore := 0.0
	if performance.AverageAcceptSeconds > 0 {
		acceptScore = (teamAverageAcceptSeconds / performance.AverageAcceptSeconds) * 50
	}
	orderScore := 0.0
	if teamAverageDailyOrders > 0 {
		orderScore = (performance.AverageDailyOrders / teamAverageDailyOrders) * 35
	}
	score := int(math.Round(math.Max(0, acceptScore+orderScore+(performance.CompletionRate*15))))
	level := RiderLevelC
	if score >= 120 {
		level = RiderLevelS
	} else if score >= 100 {
		level = RiderLevelA
	} else if score >= 80 {
		level = RiderLevelB
	}
	return score, level, RiderDispatchPriority(level)
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
	output.OrderAmountFen = order.AmountFen
	output.RefundedAmountFen = refundedFen
	output.RefundableFen = remainingFen
	return output
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
		if item.Status == "approved" && item.ExpiresAt.After(now) {
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
		items = append(items, *cloneCartItem(item))
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
	return &CartSummary{
		UserID:          userID,
		ShopID:          shopID,
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
			CoverURL:       "/assets/mock/blue-sea-cover.jpg",
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
				Status:    "approved",
			},
			{
				ID:        "mq_merchant_1_health",
				Type:      QualificationHealthCertificate,
				FileURL:   "/assets/mock/health-certificate.jpg",
				ExpiresAt: expiresAt,
				Status:    "approved",
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
			ImageURL:       "/assets/mock/beef-rice.jpg",
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
			ImageURL:       "/assets/mock/soup.jpg",
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
			ImageURL:       "/assets/mock/groupbuy-set.jpg",
			Description:    "到店扫码核销，含两份主食和两份例汤。",
			IngredientList: []string{"主食", "例汤", "到店核销"},
			PriceFen:       3999,
			StockCount:     100,
			Status:         ProductStatusActive,
		},
	}
}
