package httpapi

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestBackofficeRBACScopeMatrix(t *testing.T) {
	cases := []struct {
		name    string
		role    string
		allowed []string
		denied  []string
	}{
		{
			name:    "legacy admin keeps full access",
			role:    RoleAdmin,
			allowed: []string{AdminScopeAuditRead, AdminScopeInviteWrite, AdminScopeRefundWrite, AdminScopeDispatchWrite, AdminScopeOutboxWrite},
		},
		{
			name:    "ops admin can operate but not finance",
			role:    RoleOpsAdmin,
			allowed: []string{AdminScopeInviteWrite, AdminScopeAfterSalesReview, AdminScopeOrderCompensate, AdminScopeOutboxWrite, AdminScopeObjectCleanupWrite},
			denied:  []string{AdminScopeRefundWrite, AdminScopeAuditRead, AdminScopeSettlementRead},
		},
		{
			name:    "finance admin owns refund and settlement scopes only",
			role:    RoleFinanceAdmin,
			allowed: []string{AdminScopeRefundRead, AdminScopeRefundWrite, AdminScopeWalletRead, AdminScopeSettlementRead},
			denied:  []string{AdminScopeInviteWrite, AdminScopeOutboxWrite, AdminScopeDispatchWrite, AdminScopeAuditRead},
		},
		{
			name:    "dispatch admin owns dispatch scopes only",
			role:    RoleDispatchAdmin,
			allowed: []string{AdminScopeDispatchRead, AdminScopeDispatchWrite, AdminScopeRiderRead},
			denied:  []string{AdminScopeRefundWrite, AdminScopeAuditRead, AdminScopeInviteWrite},
		},
		{
			name:    "support admin can read and add after sales events but not approve refunds",
			role:    RoleSupportAdmin,
			allowed: []string{AdminScopeAfterSalesRead, AdminScopeAfterSalesEvent},
			denied:  []string{AdminScopeAfterSalesReview, AdminScopeRefundWrite, AdminScopeAuditRead},
		},
		{
			name:    "security auditor is read only for audit",
			role:    RoleSecurityAuditor,
			allowed: []string{AdminScopeAuditRead, AdminScopeSystemLogsRead},
			denied:  []string{AdminScopeInviteWrite, AdminScopeRefundWrite, AdminScopeOutboxWrite},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			principal := Principal{ID: "subject_1", Role: tc.role}
			for _, scope := range tc.allowed {
				if !principal.HasAdminScope(scope) {
					t.Fatalf("expected %s to have %s", tc.role, scope)
				}
			}
			for _, scope := range tc.denied {
				if principal.HasAdminScope(scope) {
					t.Fatalf("expected %s to be denied %s", tc.role, scope)
				}
			}
		})
	}
}

func TestBackofficeRBACRolesCanUseSignedTokens(t *testing.T) {
	signer := NewTokenSigner("rbac-test-secret")
	for _, role := range []string{RoleSuperAdmin, RoleOpsAdmin, RoleFinanceAdmin, RoleDispatchAdmin, RoleSupportAdmin, RoleSecurityAuditor} {
		token, _, err := signer.Issue(Principal{ID: role + "_1", Role: role}, time.Hour)
		if err != nil {
			t.Fatalf("expected signed token for %s: %v", role, err)
		}
		req := httptest.NewRequest("GET", "/api/admin/audit-logs", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		principal, err := signer.Verify(req)
		if err != nil {
			t.Fatalf("expected token verification for %s: %v", role, err)
		}
		if principal.Role != role || principal.ID != role+"_1" {
			t.Fatalf("expected %s principal round trip, got %+v", role, principal)
		}
	}
}
