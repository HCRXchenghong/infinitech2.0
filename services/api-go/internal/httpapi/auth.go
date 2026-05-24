package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	RoleAdmin           = "admin"
	RoleDispatchAdmin   = "dispatch_admin"
	RoleFinanceAdmin    = "finance_admin"
	RoleMerchant        = "merchant"
	RoleOpsAdmin        = "ops_admin"
	RoleRider           = "rider"
	RoleSecurityAuditor = "security_auditor"
	RoleStationManager  = "station_manager"
	RoleSuperAdmin      = "super_admin"
	RoleSupportAdmin    = "support_admin"
	RoleUser            = "user"
)

const (
	AdminScopeAll                = "*"
	AdminScopeAfterSalesEvent    = "after_sales:event"
	AdminScopeAfterSalesRead     = "after_sales:read"
	AdminScopeAfterSalesReview   = "after_sales:review"
	AdminScopeAuditRead          = "audit:read"
	AdminScopeAuditWrite         = "audit:write"
	AdminScopeDispatchRead       = "dispatch:read"
	AdminScopeDispatchWrite      = "dispatch:write"
	AdminScopeInviteWrite        = "invite:write"
	AdminScopeObjectCleanupRead  = "object_cleanup:read"
	AdminScopeObjectCleanupWrite = "object_cleanup:write"
	AdminScopeOperationsRead     = "operations:read"
	AdminScopeOrderCompensate    = "order:compensate"
	AdminScopeOutboxRead         = "outbox:read"
	AdminScopeOutboxWrite        = "outbox:write"
	AdminScopeRefundRead         = "refund:read"
	AdminScopeRefundWrite        = "refund:write"
	AdminScopeRBACRead           = "rbac:read"
	AdminScopeRBACWrite          = "rbac:write"
	AdminScopeRiderRead          = "rider:read"
	AdminScopeSettlementRead     = "settlement:read"
	AdminScopeSystemLogsRead     = "system_logs:read"
	AdminScopeWalletRead         = "wallet:read"
)

const adminRBACPolicyVersion = "2026-05-24.rbac.v1"

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
)

type Principal struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func (p Principal) IsZero() bool {
	return strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Role) == ""
}

func (p Principal) IsAdmin() bool {
	return p.Role == RoleAdmin || p.Role == RoleSuperAdmin
}

func (p Principal) IsBackofficeRole() bool {
	switch p.Role {
	case RoleAdmin, RoleSuperAdmin, RoleOpsAdmin, RoleFinanceAdmin, RoleDispatchAdmin, RoleSupportAdmin, RoleSecurityAuditor:
		return true
	default:
		return false
	}
}

func (p Principal) PlatformActorRole() string {
	if p.IsBackofficeRole() {
		return RoleAdmin
	}
	return p.Role
}

func (p Principal) HasAdminScope(scope string) bool {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return false
	}
	adminRoleScopeOverrideMu.RLock()
	defer adminRoleScopeOverrideMu.RUnlock()
	scopes, ok := adminScopesForRoleLocked(p.Role)
	if !ok {
		return false
	}
	if _, ok := scopes[AdminScopeAll]; ok {
		return true
	}
	_, ok = scopes[scope]
	return ok
}

func (p Principal) CanReadAuditLogs() bool {
	return p.HasAdminScope(AdminScopeAuditRead)
}

func (p Principal) CanManageAuditLogs() bool {
	return p.HasAdminScope(AdminScopeAuditWrite)
}

func (p Principal) CanManageInvites() bool {
	return p.HasAdminScope(AdminScopeInviteWrite)
}

func (p Principal) CanReadRefundSettings() bool {
	return p.HasAdminScope(AdminScopeRefundRead) || p.HasAdminScope(AdminScopeRefundWrite)
}

func (p Principal) CanManageRefunds() bool {
	return p.HasAdminScope(AdminScopeRefundWrite)
}

func (p Principal) CanReadOperationsSnapshot() bool {
	return p.HasAdminScope(AdminScopeOperationsRead)
}

func (p Principal) CanReadRBACPolicy() bool {
	return p.HasAdminScope(AdminScopeRBACRead) || p.HasAdminScope(AdminScopeRBACWrite)
}

func (p Principal) CanManageRBACPolicy() bool {
	return p.HasAdminScope(AdminScopeRBACWrite)
}

func (p Principal) CanReadAdminAfterSales() bool {
	return p.HasAdminScope(AdminScopeAfterSalesRead) || p.HasAdminScope(AdminScopeAfterSalesReview)
}

func (p Principal) CanReviewAfterSales() bool {
	return p.HasAdminScope(AdminScopeAfterSalesReview)
}

func (p Principal) CanAddAdminAfterSalesEvent() bool {
	return p.HasAdminScope(AdminScopeAfterSalesEvent) || p.HasAdminScope(AdminScopeAfterSalesReview)
}

func (p Principal) CanReadObjectCleanup() bool {
	return p.HasAdminScope(AdminScopeObjectCleanupRead) || p.HasAdminScope(AdminScopeObjectCleanupWrite)
}

func (p Principal) CanManageObjectCleanup() bool {
	return p.HasAdminScope(AdminScopeObjectCleanupWrite)
}

func (p Principal) CanReadOutbox() bool {
	return p.HasAdminScope(AdminScopeOutboxRead) || p.HasAdminScope(AdminScopeOutboxWrite)
}

func (p Principal) CanManageOutbox() bool {
	return p.HasAdminScope(AdminScopeOutboxWrite)
}

func (p Principal) CanCompensateOrders() bool {
	return p.HasAdminScope(AdminScopeOrderCompensate)
}

func (p Principal) CanReadDispatch() bool {
	return p.HasAdminScope(AdminScopeDispatchRead) || p.HasAdminScope(AdminScopeDispatchWrite)
}

func (p Principal) CanManageDispatch() bool {
	return p.HasAdminScope(AdminScopeDispatchWrite)
}

func (p Principal) CanActAsUser(userID string) bool {
	return p.IsAdmin() || (p.Role == RoleUser && p.ID == userID)
}

func (p Principal) CanActAsRider(riderID string) bool {
	return p.IsAdmin() || p.Role == RoleStationManager || (p.Role == RoleRider && p.ID == riderID)
}

func (p Principal) CanActAsMerchant(merchantID string) bool {
	return p.IsAdmin() || (p.Role == RoleMerchant && p.ID == merchantID)
}

type AuthVerifier interface {
	Verify(req *http.Request) (Principal, error)
}

type TokenSigner struct {
	secret []byte
	now    func() time.Time
}

func NewTokenSigner(secret string) TokenSigner {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		secret = "infinitech-dev-secret-change-me"
	}
	return TokenSigner{secret: []byte(secret), now: time.Now}
}

func (s TokenSigner) Issue(principal Principal, ttl time.Duration) (string, time.Time, error) {
	return s.IssueWithSession(principal, "", ttl)
}

func (s TokenSigner) IssueWithSession(principal Principal, sessionID string, ttl time.Duration) (string, time.Time, error) {
	if principal.IsZero() || ttl <= 0 {
		return "", time.Time{}, errUnauthorized
	}
	expiresAt := s.now().UTC().Add(ttl)
	claims := tokenClaims{
		SubjectID: principal.ID,
		Role:      principal.Role,
		SessionID: strings.TrimSpace(sessionID),
		ExpiresAt: expiresAt.Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, errUnauthorized
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	return encodedPayload + "." + s.sign(encodedPayload), expiresAt, nil
}

func (s TokenSigner) Verify(req *http.Request) (Principal, error) {
	_, claims, err := s.verifyRequestClaims(req)
	if err != nil {
		return Principal{}, errUnauthorized
	}
	return Principal{ID: claims.SubjectID, Role: claims.Role}, nil
}

func (s TokenSigner) verifyRequestClaims(req *http.Request) (string, tokenClaims, error) {
	token, err := bearerToken(req)
	if err != nil {
		return "", tokenClaims{}, err
	}
	claims, err := s.verifyTokenClaims(token)
	if err != nil {
		return "", tokenClaims{}, err
	}
	return token, claims, nil
}

func (s TokenSigner) verifyTokenClaims(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return tokenClaims{}, errUnauthorized
	}
	expectedSignature := s.sign(parts[0])
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return tokenClaims{}, errUnauthorized
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenClaims{}, errUnauthorized
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, errUnauthorized
	}
	if claims.SubjectID == "" || !isKnownRole(claims.Role) || s.now().UTC().Unix() >= claims.ExpiresAt {
		return tokenClaims{}, errUnauthorized
	}
	return claims, nil
}

func (s TokenSigner) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

type tokenClaims struct {
	SubjectID string `json:"sub"`
	Role      string `json:"role"`
	SessionID string `json:"sid,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

type SessionAuthVerifier struct {
	signer   TokenSigner
	sessions AuthSessionStore
	now      func() time.Time
}

func NewSessionAuthVerifier(signer TokenSigner, sessions AuthSessionStore) SessionAuthVerifier {
	return SessionAuthVerifier{signer: signer, sessions: sessions, now: time.Now}
}

func (verifier SessionAuthVerifier) Verify(req *http.Request) (Principal, error) {
	token, claims, err := verifier.signer.verifyRequestClaims(req)
	if err != nil {
		return Principal{}, errUnauthorized
	}
	sessionID := strings.TrimSpace(claims.SessionID)
	if sessionID == "" || verifier.sessions == nil {
		return Principal{}, errUnauthorized
	}
	principal := Principal{ID: claims.SubjectID, Role: claims.Role}
	now := time.Now
	if verifier.now != nil {
		now = verifier.now
	}
	if err := verifier.sessions.Verify(req.Context(), sessionID, tokenHash(token), principal, now().UTC()); err != nil {
		return Principal{}, errUnauthorized
	}
	return principal, nil
}

type ChainedVerifier []AuthVerifier

func (verifiers ChainedVerifier) Verify(req *http.Request) (Principal, error) {
	for _, verifier := range verifiers {
		principal, err := verifier.Verify(req)
		if err == nil {
			return principal, nil
		}
	}
	return Principal{}, errUnauthorized
}

type DevBearerVerifier struct{}

func (DevBearerVerifier) Verify(req *http.Request) (Principal, error) {
	token, err := bearerToken(req)
	if err != nil {
		return Principal{}, errUnauthorized
	}
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return Principal{}, errUnauthorized
	}
	role := strings.TrimSpace(parts[0])
	id := strings.TrimSpace(parts[1])
	if id == "" || !isKnownRole(role) {
		return Principal{}, errUnauthorized
	}
	return Principal{ID: id, Role: role}, nil
}

func bearerToken(req *http.Request) (string, error) {
	header := strings.TrimSpace(req.Header.Get("Authorization"))
	if header == "" {
		return "", errUnauthorized
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == header || token == "" {
		return "", errUnauthorized
	}
	return token, nil
}

func isKnownRole(role string) bool {
	switch role {
	case RoleAdmin, RoleDispatchAdmin, RoleFinanceAdmin, RoleMerchant, RoleOpsAdmin, RoleRider, RoleSecurityAuditor, RoleStationManager, RoleSuperAdmin, RoleSupportAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func scopeSet(scopes ...string) map[string]struct{} {
	output := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			output[scope] = struct{}{}
		}
	}
	return output
}

var (
	adminRoleScopeOverrideMu sync.RWMutex
	adminRoleScopeOverrides  = map[string]map[string]struct{}{}
)

var adminRoleScopes = map[string]map[string]struct{}{
	RoleAdmin:      scopeSet(AdminScopeAll),
	RoleSuperAdmin: scopeSet(AdminScopeAll),
	RoleOpsAdmin: scopeSet(
		AdminScopeAfterSalesRead,
		AdminScopeAfterSalesReview,
		AdminScopeDispatchRead,
		AdminScopeInviteWrite,
		AdminScopeObjectCleanupRead,
		AdminScopeObjectCleanupWrite,
		AdminScopeOperationsRead,
		AdminScopeOrderCompensate,
		AdminScopeOutboxRead,
		AdminScopeOutboxWrite,
		AdminScopeRBACRead,
		AdminScopeRiderRead,
	),
	RoleFinanceAdmin: scopeSet(
		AdminScopeOperationsRead,
		AdminScopeRBACRead,
		AdminScopeRefundRead,
		AdminScopeRefundWrite,
		AdminScopeSettlementRead,
		AdminScopeWalletRead,
	),
	RoleDispatchAdmin: scopeSet(
		AdminScopeDispatchRead,
		AdminScopeDispatchWrite,
		AdminScopeOperationsRead,
		AdminScopeRBACRead,
		AdminScopeRiderRead,
	),
	RoleSupportAdmin: scopeSet(
		AdminScopeAfterSalesEvent,
		AdminScopeAfterSalesRead,
		AdminScopeOperationsRead,
		AdminScopeRBACRead,
	),
	RoleSecurityAuditor: scopeSet(
		AdminScopeAuditRead,
		AdminScopeRBACRead,
		AdminScopeSystemLogsRead,
	),
}

type AdminScopePolicy struct {
	Key         string `json:"key"`
	Category    string `json:"category"`
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
}

type AdminRolePolicy struct {
	Role        string   `json:"role"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
	BuiltIn     bool     `json:"built_in"`
	ReadOnly    bool     `json:"read_only"`
	DataDomain  string   `json:"data_domain"`
}

type AdminRBACPolicy struct {
	Version             string             `json:"version"`
	Roles               []AdminRolePolicy  `json:"roles"`
	Scopes              []AdminScopePolicy `json:"scopes"`
	CurrentRole         string             `json:"current_role"`
	CurrentRoleScopes   []string           `json:"current_role_scopes"`
	CanRequestChanges   bool               `json:"can_request_changes"`
	ChangeApprovalModel string             `json:"change_approval_model"`
	Notes               []string           `json:"notes"`
}

var adminRolePolicyCatalog = []AdminRolePolicy{
	{
		Role:        RoleAdmin,
		Name:        "Legacy Admin",
		Description: "Compatibility administrator. Keeps full access while older bootstrap flows migrate to super_admin.",
		BuiltIn:     true,
		ReadOnly:    true,
		DataDomain:  "platform",
	},
	{
		Role:        RoleSuperAdmin,
		Name:        "Super Admin",
		Description: "Full platform administrator. Owns RBAC policy change requests and emergency operations.",
		BuiltIn:     true,
		ReadOnly:    false,
		DataDomain:  "platform",
	},
	{
		Role:        RoleOpsAdmin,
		Name:        "Operations Admin",
		Description: "Runs onboarding, after-sales review, object cleanup, outbox recovery and order compensation.",
		BuiltIn:     true,
		ReadOnly:    false,
		DataDomain:  "platform",
	},
	{
		Role:        RoleFinanceAdmin,
		Name:        "Finance Admin",
		Description: "Handles refund policy, refund operations, wallet visibility and settlement read models.",
		BuiltIn:     true,
		ReadOnly:    false,
		DataDomain:  "finance",
	},
	{
		Role:        RoleDispatchAdmin,
		Name:        "Dispatch Admin",
		Description: "Reads and manages dispatch work, station rider views and dispatch task configuration.",
		BuiltIn:     true,
		ReadOnly:    false,
		DataDomain:  "dispatch",
	},
	{
		Role:        RoleSupportAdmin,
		Name:        "Support Admin",
		Description: "Reads after-sales queues and adds customer-service events without approving refunds.",
		BuiltIn:     true,
		ReadOnly:    false,
		DataDomain:  "support",
	},
	{
		Role:        RoleSecurityAuditor,
		Name:        "Security Auditor",
		Description: "Reads audit logs and security policy metadata without operational write access.",
		BuiltIn:     true,
		ReadOnly:    true,
		DataDomain:  "security",
	},
}

var adminScopePolicyCatalog = []AdminScopePolicy{
	{Key: AdminScopeAll, Category: "system", Description: "All backoffice permissions.", RiskLevel: "critical"},
	{Key: AdminScopeAfterSalesEvent, Category: "support", Description: "Add customer-service events and evidence to after-sales requests.", RiskLevel: "medium"},
	{Key: AdminScopeAfterSalesRead, Category: "support", Description: "Read platform after-sales queues and evidence.", RiskLevel: "medium"},
	{Key: AdminScopeAfterSalesReview, Category: "support", Description: "Review after-sales requests and trigger approved refunds.", RiskLevel: "high"},
	{Key: AdminScopeAuditRead, Category: "security", Description: "Read operation audit logs.", RiskLevel: "high"},
	{Key: AdminScopeAuditWrite, Category: "security", Description: "Emit audit retention alert events for downstream notification and incident workflows.", RiskLevel: "critical"},
	{Key: AdminScopeDispatchRead, Category: "dispatch", Description: "Read dispatch events, station riders and station order queues.", RiskLevel: "medium"},
	{Key: AdminScopeDispatchWrite, Category: "dispatch", Description: "Run auto assign, timeout reassignment, manual assignment and station task writes.", RiskLevel: "high"},
	{Key: AdminScopeInviteWrite, Category: "onboarding", Description: "Create merchant, station manager and rider onboarding invites.", RiskLevel: "high"},
	{Key: AdminScopeObjectCleanupRead, Category: "storage", Description: "Read object cleanup candidates and cleanup statistics.", RiskLevel: "medium"},
	{Key: AdminScopeObjectCleanupWrite, Category: "storage", Description: "Mark object cleanup completion and failure results.", RiskLevel: "high"},
	{Key: AdminScopeOperationsRead, Category: "operations", Description: "Read the admin operations snapshot.", RiskLevel: "medium"},
	{Key: AdminScopeOrderCompensate, Category: "orders", Description: "Run order state compensation for drift recovery.", RiskLevel: "critical"},
	{Key: AdminScopeOutboxRead, Category: "events", Description: "Read outbox events, stats and relay health.", RiskLevel: "medium"},
	{Key: AdminScopeOutboxWrite, Category: "events", Description: "Claim, renew, publish, fail and replay outbox events.", RiskLevel: "critical"},
	{Key: AdminScopeRefundRead, Category: "finance", Description: "Read refund settings and refund policy metadata.", RiskLevel: "medium"},
	{Key: AdminScopeRefundWrite, Category: "finance", Description: "Change refund settings and execute admin refunds.", RiskLevel: "critical"},
	{Key: AdminScopeRBACRead, Category: "security", Description: "Read the backoffice RBAC policy matrix.", RiskLevel: "medium"},
	{Key: AdminScopeRBACWrite, Category: "security", Description: "Request RBAC policy changes and security governance actions.", RiskLevel: "critical"},
	{Key: AdminScopeRiderRead, Category: "dispatch", Description: "Read rider and rider performance views.", RiskLevel: "medium"},
	{Key: AdminScopeSettlementRead, Category: "finance", Description: "Read settlement and commission reports.", RiskLevel: "high"},
	{Key: AdminScopeSystemLogsRead, Category: "security", Description: "Read system log and audit-health surfaces.", RiskLevel: "high"},
	{Key: AdminScopeWalletRead, Category: "finance", Description: "Read wallet and balance ledgers.", RiskLevel: "high"},
}

func AdminRBACPolicyForPrincipal(principal Principal) AdminRBACPolicy {
	roles := make([]AdminRolePolicy, 0, len(adminRolePolicyCatalog))
	for _, role := range adminRolePolicyCatalog {
		role.Scopes = AdminScopesForRole(role.Role)
		roles = append(roles, role)
	}
	return AdminRBACPolicy{
		Version:             adminRBACPolicyVersion,
		Roles:               roles,
		Scopes:              append([]AdminScopePolicy(nil), adminScopePolicyCatalog...),
		CurrentRole:         principal.Role,
		CurrentRoleScopes:   AdminScopesForRole(principal.Role),
		CanRequestChanges:   principal.CanManageRBACPolicy(),
		ChangeApprovalModel: "request_review_apply_with_audit_and_runtime_replay",
		Notes: []string{
			"Built-in scopes are enforced by api-go route guards.",
			"Approved RBAC change requests can be manually applied and are replayed from audit logs when api-go starts.",
			"Field-level, station-level and merchant-level policy rules are still pending commercial governance work.",
		},
	}
}

func AdminScopesForRole(role string) []string {
	adminRoleScopeOverrideMu.RLock()
	defer adminRoleScopeOverrideMu.RUnlock()
	scopes, ok := adminScopesForRoleLocked(role)
	if !ok {
		return []string{}
	}
	return adminScopeSetToSortedSlice(scopes)
}

func adminScopesForRoleLocked(role string) (map[string]struct{}, bool) {
	role = strings.TrimSpace(role)
	if scopes, ok := adminRoleScopeOverrides[role]; ok {
		return scopes, true
	}
	scopes, ok := adminRoleScopes[role]
	return scopes, ok
}

func adminScopeSetToSortedSlice(scopes map[string]struct{}) []string {
	if _, ok := scopes[AdminScopeAll]; ok {
		return []string{AdminScopeAll}
	}
	output := make([]string, 0, len(scopes))
	for scope := range scopes {
		output = append(output, scope)
	}
	sort.Strings(output)
	return output
}

func ValidateAdminRBACRoleScopes(role string, scopes []string) ([]string, bool) {
	role = strings.TrimSpace(role)
	if !IsBackofficeRoleName(role) {
		return nil, false
	}
	normalized, valid := NormalizeAdminScopeList(scopes)
	if !valid || len(normalized) == 0 {
		return nil, false
	}
	hasAllScope := adminScopeListContains(normalized, AdminScopeAll)
	if hasAllScope && role != RoleAdmin && role != RoleSuperAdmin {
		return nil, false
	}
	if (role == RoleAdmin || role == RoleSuperAdmin) && !hasAllScope {
		return nil, false
	}
	return normalized, true
}

func ApplyAdminRBACRoleScopes(role string, scopes []string) ([]string, bool) {
	role = strings.TrimSpace(role)
	normalized, valid := ValidateAdminRBACRoleScopes(role, scopes)
	if !valid {
		return nil, false
	}
	adminRoleScopeOverrideMu.Lock()
	defer adminRoleScopeOverrideMu.Unlock()
	adminRoleScopeOverrides[role] = scopeSet(normalized...)
	return normalized, true
}

func resetAdminRBACRoleScopeOverrides() {
	adminRoleScopeOverrideMu.Lock()
	defer adminRoleScopeOverrideMu.Unlock()
	adminRoleScopeOverrides = map[string]map[string]struct{}{}
}

func IsBackofficeRoleName(role string) bool {
	return (Principal{Role: strings.TrimSpace(role), ID: "role_check"}).IsBackofficeRole()
}

func NormalizeAdminScopeList(scopes []string) ([]string, bool) {
	seen := map[string]struct{}{}
	output := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if !IsKnownAdminScope(scope) {
			return nil, false
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		output = append(output, scope)
	}
	sort.Strings(output)
	return output, true
}

func IsKnownAdminScope(scope string) bool {
	scope = strings.TrimSpace(scope)
	for _, item := range adminScopePolicyCatalog {
		if item.Key == scope {
			return true
		}
	}
	return false
}

func adminScopeListContains(scopes []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == target {
			return true
		}
	}
	return false
}
