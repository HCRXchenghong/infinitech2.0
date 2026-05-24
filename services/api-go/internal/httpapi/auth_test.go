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
			allowed: []string{AdminScopeInviteWrite, AdminScopeAfterSalesReview, AdminScopeOrderCompensate, AdminScopeOutboxWrite, AdminScopeObjectCleanupWrite, AdminScopeRBACRead},
			denied:  []string{AdminScopeRefundWrite, AdminScopeAuditRead, AdminScopeRBACWrite, AdminScopeSettlementRead},
		},
		{
			name:    "finance admin owns refund and settlement scopes only",
			role:    RoleFinanceAdmin,
			allowed: []string{AdminScopeRefundRead, AdminScopeRefundWrite, AdminScopeRBACRead, AdminScopeWalletRead, AdminScopeSettlementRead},
			denied:  []string{AdminScopeInviteWrite, AdminScopeOutboxWrite, AdminScopeDispatchWrite, AdminScopeRBACWrite, AdminScopeAuditRead},
		},
		{
			name:    "dispatch admin owns dispatch scopes only",
			role:    RoleDispatchAdmin,
			allowed: []string{AdminScopeDispatchRead, AdminScopeDispatchWrite, AdminScopeRBACRead, AdminScopeRiderRead},
			denied:  []string{AdminScopeRefundWrite, AdminScopeAuditRead, AdminScopeInviteWrite, AdminScopeRBACWrite},
		},
		{
			name:    "support admin can read and add after sales events but not approve refunds",
			role:    RoleSupportAdmin,
			allowed: []string{AdminScopeAfterSalesRead, AdminScopeAfterSalesEvent, AdminScopeRBACRead},
			denied:  []string{AdminScopeAfterSalesReview, AdminScopeRefundWrite, AdminScopeAuditRead, AdminScopeRBACWrite},
		},
		{
			name:    "security auditor is read only for audit",
			role:    RoleSecurityAuditor,
			allowed: []string{AdminScopeAuditRead, AdminScopeRBACRead, AdminScopeSystemLogsRead},
			denied:  []string{AdminScopeInviteWrite, AdminScopeRefundWrite, AdminScopeOutboxWrite, AdminScopeRBACWrite},
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

func TestBackofficeRBACPolicyCatalog(t *testing.T) {
	policy := AdminRBACPolicyForPrincipal(Principal{ID: "admin_1", Role: RoleSuperAdmin})
	if policy.Version != adminRBACPolicyVersion || !policy.CanRequestChanges {
		t.Fatalf("expected super admin policy with change access, got %+v", policy)
	}
	if len(policy.Roles) < 7 || len(policy.Scopes) < 20 {
		t.Fatalf("expected built-in roles and scopes, got %+v", policy)
	}
	if !IsKnownAdminScope(AdminScopeRBACRead) || !IsKnownAdminScope(AdminScopeRBACWrite) {
		t.Fatalf("expected RBAC scopes to be catalogued")
	}
	normalized, ok := NormalizeAdminScopeList([]string{AdminScopeRefundWrite, AdminScopeRefundWrite, " "})
	if !ok || len(normalized) != 1 || normalized[0] != AdminScopeRefundWrite {
		t.Fatalf("expected scope list normalization, got scopes=%+v ok=%v", normalized, ok)
	}
	if _, ok := NormalizeAdminScopeList([]string{"unknown:scope"}); ok {
		t.Fatalf("expected unknown scope to be rejected")
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
