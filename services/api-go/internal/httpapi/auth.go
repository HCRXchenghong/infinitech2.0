package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
	AdminScopeRiderRead          = "rider:read"
	AdminScopeSettlementRead     = "settlement:read"
	AdminScopeSystemLogsRead     = "system_logs:read"
	AdminScopeWalletRead         = "wallet:read"
)

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
	scopes, ok := adminRoleScopes[p.Role]
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
		AdminScopeRiderRead,
	),
	RoleFinanceAdmin: scopeSet(
		AdminScopeOperationsRead,
		AdminScopeRefundRead,
		AdminScopeRefundWrite,
		AdminScopeSettlementRead,
		AdminScopeWalletRead,
	),
	RoleDispatchAdmin: scopeSet(
		AdminScopeDispatchRead,
		AdminScopeDispatchWrite,
		AdminScopeOperationsRead,
		AdminScopeRiderRead,
	),
	RoleSupportAdmin: scopeSet(
		AdminScopeAfterSalesEvent,
		AdminScopeAfterSalesRead,
		AdminScopeOperationsRead,
	),
	RoleSecurityAuditor: scopeSet(
		AdminScopeAuditRead,
		AdminScopeSystemLogsRead,
	),
}
