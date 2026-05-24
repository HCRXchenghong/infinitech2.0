package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"infinitech2/services/api-go/internal/platform"
)

func TestHealthAndHomeModules(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/home/modules")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var envelope struct {
		Success bool                  `json:"success"`
		Data    []platform.HomeModule `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.Success || len(envelope.Data) < 4 || envelope.Data[0].Key != "takeout" {
		t.Fatalf("unexpected home modules response: %+v", envelope)
	}
	cardsResp, err := http.Get(server.URL + "/api/home/cards")
	if err != nil {
		t.Fatal(err)
	}
	defer cardsResp.Body.Close()
	if cardsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", cardsResp.StatusCode)
	}
	var cardsEnvelope struct {
		Success bool                `json:"success"`
		Data    []platform.HomeCard `json:"data"`
	}
	if err := json.NewDecoder(cardsResp.Body).Decode(&cardsEnvelope); err != nil {
		t.Fatal(err)
	}
	if !cardsEnvelope.Success || len(cardsEnvelope.Data) < 2 || cardsEnvelope.Data[1].Type != platform.HomeCardCircle {
		t.Fatalf("unexpected home cards response: %+v", cardsEnvelope)
	}
}

func TestOrderCreditPayAndGrabHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":900}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)

	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":900,"idempotency_key":"credit_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	payBody := authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_http"}`, http.StatusOK)
	if payBody["data"].(map[string]any)["order"].(map[string]any)["status"] != platform.StatusDispatching {
		t.Fatalf("expected dispatching after pay: %+v", payBody)
	}

	grabBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/grab", riderToken("rider_1"), `{}`, http.StatusOK)
	if grabBody["data"].(map[string]any)["status"] != platform.StatusRiderAssigned {
		t.Fatalf("expected rider assigned after grab: %+v", grabBody)
	}
	pickupBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/pickup", riderToken("rider_1"), `{}`, http.StatusOK)
	if pickupBody["data"].(map[string]any)["status"] != platform.StatusPickedUp {
		t.Fatalf("expected picked up after pickup: %+v", pickupBody)
	}
	deliveredBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/delivered", riderToken("rider_1"), `{}`, http.StatusOK)
	if deliveredBody["data"].(map[string]any)["status"] != platform.StatusCompleted {
		t.Fatalf("expected completed after delivered: %+v", deliveredBody)
	}
}

func TestAdminRefundSettingsAndOrderRefundHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	order, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_1", Type: platform.OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_http_refund"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, account, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_http_refund"}); err != nil || paidOrder.Status != platform.StatusDispatching || account.Balance != 0 {
		t.Fatalf("expected paid order setup, order=%+v account=%+v err=%v", paidOrder, account, err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/refund-settings", userToken("user_1"), http.StatusForbidden)
	settingsBody := authGetJSON(t, server.URL+"/api/admin/refund-settings", adminToken("admin_1"), http.StatusOK)
	if settingsBody["data"].(map[string]any)["default_refund_strategy"] != platform.RefundStrategyBalanceFirst {
		t.Fatalf("expected balance-first default settings, got %+v", settingsBody)
	}
	savedSettingsBody := authPutJSON(t, server.URL+"/api/admin/refund-settings", adminToken("admin_1"), `{"default_refund_strategy":"balance_first"}`, http.StatusOK)
	if savedSettingsBody["data"].(map[string]any)["default_refund_strategy"] != platform.RefundStrategyBalanceFirst {
		t.Fatalf("expected saved refund settings, got %+v", savedSettingsBody)
	}
	authPostJSON(t, server.URL+"/api/orders/"+order.ID+"/refund", userToken("user_1"), `{"reason":"商品售罄","idempotency_key":"refund_http_1"}`, http.StatusForbidden)
	refundBody := authPostJSON(t, server.URL+"/api/orders/"+order.ID+"/refund", adminToken("admin_1"), `{"reason":"商品售罄","idempotency_key":"refund_http_1"}`, http.StatusOK)
	refundData := refundBody["data"].(map[string]any)
	if refundData["refund"].(map[string]any)["status"] != platform.RefundStatusSuccess ||
		refundData["order"].(map[string]any)["status"] != platform.StatusRefunded ||
		refundData["wallet_account"].(map[string]any)["balance_fen"] != float64(1200) {
		t.Fatalf("expected balance refund response, got %+v", refundBody)
	}
	replayedBody := authPostJSON(t, server.URL+"/api/orders/"+order.ID+"/refund", adminToken("admin_1"), `{"reason":"重复回调","idempotency_key":"refund_http_1"}`, http.StatusOK)
	replayedData := replayedBody["data"].(map[string]any)
	if replayedData["refund"].(map[string]any)["id"] != refundData["refund"].(map[string]any)["id"] ||
		replayedData["wallet_account"].(map[string]any)["balance_fen"] != float64(1200) {
		t.Fatalf("expected idempotent HTTP refund replay, got %+v", replayedBody)
	}
	authGetJSON(t, server.URL+"/api/admin/audit-logs", userToken("user_1"), http.StatusForbidden)
	authGetJSON(t, server.URL+"/api/admin/audit-logs?limit=1", securityAuditorToken("auditor_1"), http.StatusOK)
	authPostJSON(t, server.URL+"/api/admin/merchant-invites", securityAuditorToken("auditor_1"), `{}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/orders/"+order.ID+"/state/compensate", securityAuditorToken("auditor_1"), `{}`, http.StatusForbidden)
	auditBody := authGetJSON(t, server.URL+"/api/admin/audit-logs?target_type=order&target_id="+order.ID+"&limit=5", adminToken("admin_1"), http.StatusOK)
	auditLogs := auditBody["data"].([]any)
	if len(auditLogs) != 2 || auditLogs[0].(map[string]any)["action"] != "admin.order.refunded" || auditLogs[0].(map[string]any)["actor_id"] != "admin_1" {
		t.Fatalf("expected refund audit logs, got %+v", auditBody)
	}
	refundAudit := auditLogs[0].(map[string]any)
	if refundAudit["integrity_algorithm"] != "sha256:v1" || refundAudit["integrity_hash"] == "" || refundAudit["integrity_verified"] != true {
		t.Fatalf("expected refund audit HTTP response to expose verified integrity proof, got %+v", refundAudit)
	}
	refundPayload := auditLogs[0].(map[string]any)["payload"].(map[string]any)
	if refundPayload["idempotency_key"] != "refund_http_1" || refundPayload["amount_fen"] != float64(1200) {
		t.Fatalf("expected refund audit payload without sensitive data, got %+v", refundPayload)
	}
	newestRefundAuditAt, ok := auditLogs[0].(map[string]any)["created_at"].(string)
	if !ok || newestRefundAuditAt == "" {
		t.Fatalf("expected refund audit created_at, got %+v", auditLogs[0])
	}
	windowAuditBody := authGetJSON(t, server.URL+"/api/admin/audit-logs?target_type=order&target_id="+order.ID+"&after="+newestRefundAuditAt+"&limit=5", adminToken("admin_1"), http.StatusOK)
	windowAuditLogs := windowAuditBody["data"].([]any)
	if len(windowAuditLogs) != 1 || windowAuditLogs[0].(map[string]any)["created_at"] != newestRefundAuditAt {
		t.Fatalf("expected after window to keep newest refund audit log, got %+v", windowAuditBody)
	}
	settingsAuditBody := authGetJSON(t, server.URL+"/api/admin/audit-logs?action=admin.refund_settings.updated&limit=1", adminToken("admin_1"), http.StatusOK)
	settingsAuditLogs := settingsAuditBody["data"].([]any)
	if len(settingsAuditLogs) != 1 || settingsAuditLogs[0].(map[string]any)["target_type"] != "refund_settings" {
		t.Fatalf("expected refund settings audit log, got %+v", settingsAuditBody)
	}
	exportBody := authGetJSON(t, server.URL+"/api/admin/audit-logs/export?target_type=order&target_id="+order.ID+"&limit=5", securityAuditorToken("auditor_1"), http.StatusOK)
	exportData := exportBody["data"].(map[string]any)
	if exportData["format"] != "csv" || exportData["row_count"] != float64(2) {
		t.Fatalf("expected audit export csv metadata, got %+v", exportBody)
	}
	exportCSV := exportData["csv"].(string)
	if !strings.Contains(exportCSV, "admin.order.refunded") || !strings.Contains(exportCSV, "integrity_verified") {
		t.Fatalf("expected audit export CSV to include refund audit rows and header, got %q", exportCSV)
	}
	exportAudit := exportData["audit_log"].(map[string]any)
	if exportAudit["action"] != "admin.audit_logs.exported" || exportAudit["target_type"] != "audit_export" || exportAudit["actor_type"] != RoleSecurityAuditor {
		t.Fatalf("expected audit export to be audited, got %+v", exportAudit)
	}
	exportAuditLogsBody := authGetJSON(t, server.URL+"/api/admin/audit-logs?action=admin.audit_logs.exported&limit=1", adminToken("admin_1"), http.StatusOK)
	exportAuditLogs := exportAuditLogsBody["data"].([]any)
	if len(exportAuditLogs) != 1 || exportAuditLogs[0].(map[string]any)["target_type"] != "audit_export" {
		t.Fatalf("expected audit export log to be queryable, got %+v", exportAuditLogsBody)
	}
	retentionBody := authGetJSON(t, server.URL+"/api/admin/audit-logs/retention-report?retention_days=2555&hot_days=180&integrity_sample_limit=10", securityAuditorToken("auditor_1"), http.StatusOK)
	retentionData := retentionBody["data"].(map[string]any)
	if retentionData["total_logs"].(float64) < 4 || retentionData["export_events"] != float64(1) || retentionData["integrity_failures"] != float64(0) {
		t.Fatalf("expected audit retention report to include current ledger health, got %+v", retentionBody)
	}
	if retentionData["status"] != "warning" && retentionData["status"] != "ok" {
		t.Fatalf("expected retention report status to be ok or warning, got %+v", retentionData)
	}
	missingCriticalActions := retentionData["missing_critical_actions"].([]any)
	if len(missingCriticalActions) == 0 {
		t.Fatalf("expected retention report to expose missing critical action coverage, got %+v", retentionBody)
	}
	authGetJSON(t, server.URL+"/api/admin/audit-logs/retention-report", userToken("user_1"), http.StatusForbidden)
}

func TestAdminRBACRoleMatrixHTTPFlow(t *testing.T) {
	resetAdminRBACRoleScopeOverrides()
	t.Cleanup(resetAdminRBACRoleScopeOverrides)
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_1"), http.StatusOK)
	authPutJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_1"), `{"default_refund_strategy":"balance_first"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/admin/merchant-invites", roleToken(RoleFinanceAdmin, "finance_1"), `{}`, http.StatusForbidden)
	rbacPolicy := authGetJSON(t, server.URL+"/api/admin/rbac/policy", roleToken(RoleFinanceAdmin, "finance_1"), http.StatusOK)
	if rbacPolicy["data"].(map[string]any)["current_role"] != RoleFinanceAdmin {
		t.Fatalf("expected finance admin RBAC policy view, got %+v", rbacPolicy)
	}
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests", roleToken(RoleFinanceAdmin, "finance_1"), `{"role":"finance_admin","requested_scopes":["refund:read"],"reason":"least privilege review"}`, http.StatusForbidden)

	opsInvite := authPostJSON(t, server.URL+"/api/admin/merchant-invites", roleToken(RoleOpsAdmin, "ops_1"), `{}`, http.StatusCreated)
	if opsInvite["data"].(map[string]any)["created_by_subject_id"] != "ops_1" {
		t.Fatalf("expected ops admin to create merchant invite with own subject, got %+v", opsInvite)
	}
	authGetJSON(t, server.URL+"/api/admin/audit-logs", roleToken(RoleOpsAdmin, "ops_1"), http.StatusForbidden)

	authGetJSON(t, server.URL+"/api/admin/audit-logs?limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	authGetJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusForbidden)

	authGetJSON(t, server.URL+"/api/admin/after-sales", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	authPutJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleSupportAdmin, "support_1"), `{"default_refund_strategy":"original_route_first"}`, http.StatusForbidden)

	authGetJSON(t, server.URL+"/api/station-manager/riders", roleToken(RoleDispatchAdmin, "dispatch_1"), http.StatusOK)
	authPostJSON(t, server.URL+"/api/admin/outbox/events/claim", roleToken(RoleDispatchAdmin, "dispatch_1"), `{"topic":"order.paid","limit":1,"lease_owner":"dispatch","lease_seconds":30}`, http.StatusForbidden)

	rbacChange := authPostJSON(t, server.URL+"/api/admin/rbac/change-requests", roleToken(RoleSuperAdmin, "super_1"), `{"role":"finance_admin","requested_scopes":["refund:read","rbac:read"],"reason":"finance least privilege recertification"}`, http.StatusCreated)
	rbacData := rbacChange["data"].(map[string]any)
	if rbacData["status"] != "pending_approval" || rbacData["auto_applied"] != false || rbacData["policy_version"] == "" {
		t.Fatalf("expected pending audited RBAC change request, got %+v", rbacChange)
	}
	changeRequestID := rbacData["id"].(string)
	audit := rbacData["audit_log"].(map[string]any)
	if audit["action"] != "admin.rbac.change_requested" || audit["target_type"] != "admin_rbac_role" || audit["target_id"] != RoleFinanceAdmin {
		t.Fatalf("expected RBAC change audit log, got %+v", audit)
	}
	auditPayload := audit["payload"].(map[string]any)
	if auditPayload["role"] != RoleFinanceAdmin || auditPayload["status"] != "pending_approval" {
		t.Fatalf("expected sanitized RBAC audit payload, got %+v", auditPayload)
	}
	pendingChanges := authGetJSON(t, server.URL+"/api/admin/rbac/change-requests?status=pending_approval&limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	pendingData := pendingChanges["data"].(map[string]any)
	if pendingData["pending_count"] != float64(1) || pendingData["auto_apply"] != false {
		t.Fatalf("expected pending RBAC change request list, got %+v", pendingChanges)
	}
	pendingItems := pendingData["items"].([]any)
	if len(pendingItems) != 1 || pendingItems[0].(map[string]any)["id"] != changeRequestID {
		t.Fatalf("expected pending list to include change request, got %+v", pendingItems)
	}
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/review", roleToken(RoleSuperAdmin, "super_1"), `{"decision":"approve","reason":"self approval must fail"}`, http.StatusConflict)
	reviewed := authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/review", roleToken(RoleSuperAdmin, "super_2"), `{"decision":"approve","reason":"least privilege approved"}`, http.StatusOK)
	reviewedData := reviewed["data"].(map[string]any)
	reviewedRequest := reviewedData["change_request"].(map[string]any)
	if reviewedRequest["status"] != "approved" || reviewedRequest["reviewed_by_admin"] != "super_2" || reviewedData["auto_applied"] != false {
		t.Fatalf("expected approved but not auto-applied RBAC request, got %+v", reviewed)
	}
	reviewAudit := reviewedData["audit_log"].(map[string]any)
	if reviewAudit["action"] != "admin.rbac.change_reviewed" || reviewAudit["target_type"] != "admin_rbac_change_request" || reviewAudit["target_id"] != changeRequestID {
		t.Fatalf("expected RBAC review audit log, got %+v", reviewAudit)
	}
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/apply", roleToken(RoleSuperAdmin, "super_1"), `{"reason":"requester cannot apply"}`, http.StatusConflict)
	applied := authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/apply", roleToken(RoleSuperAdmin, "super_2"), `{"reason":"approved least privilege policy"}`, http.StatusOK)
	appliedData := applied["data"].(map[string]any)
	appliedRequest := appliedData["change_request"].(map[string]any)
	if appliedRequest["status"] != "applied" || appliedRequest["applied"] != true || appliedData["runtime_applied"] != true {
		t.Fatalf("expected applied RBAC request, got %+v", applied)
	}
	if len(appliedRequest["previous_scopes"].([]any)) == 0 {
		t.Fatalf("expected applied RBAC request to retain previous scopes for rollback, got %+v", appliedRequest)
	}
	applyAudit := appliedData["audit_log"].(map[string]any)
	if applyAudit["action"] != "admin.rbac.change_applied" || applyAudit["target_type"] != "admin_rbac_role" || applyAudit["target_id"] != RoleFinanceAdmin {
		t.Fatalf("expected RBAC apply audit log, got %+v", applyAudit)
	}
	authGetJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_2"), http.StatusOK)
	authPutJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_2"), `{"default_refund_strategy":"balance_first"}`, http.StatusForbidden)
	approvedChanges := authGetJSON(t, server.URL+"/api/admin/rbac/change-requests?status=approved&limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	approvedItems := approvedChanges["data"].(map[string]any)["items"].([]any)
	if len(approvedItems) != 0 {
		t.Fatalf("expected applied request to leave approved filter, got %+v", approvedChanges)
	}
	appliedChanges := authGetJSON(t, server.URL+"/api/admin/rbac/change-requests?status=applied&limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	appliedItems := appliedChanges["data"].(map[string]any)["items"].([]any)
	if len(appliedItems) != 1 || appliedItems[0].(map[string]any)["status"] != "applied" {
		t.Fatalf("expected applied request in RBAC ledger, got %+v", appliedChanges)
	}
	resetAdminRBACRoleScopeOverrides()
	restoredServer := httptest.NewServer(NewRouter(store))
	defer restoredServer.Close()
	authPutJSON(t, restoredServer.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_3"), `{"default_refund_strategy":"balance_first"}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/rollback", roleToken(RoleSuperAdmin, "super_1"), `{"reason":"requester cannot rollback"}`, http.StatusConflict)
	rolledBack := authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/rollback", roleToken(RoleSuperAdmin, "super_3"), `{"reason":"restore previous finance runtime policy"}`, http.StatusOK)
	rolledBackData := rolledBack["data"].(map[string]any)
	rolledBackRequest := rolledBackData["change_request"].(map[string]any)
	if rolledBackRequest["status"] != "rolled_back" || rolledBackRequest["rolled_back"] != true || rolledBackData["runtime_applied"] != true || rolledBackData["rolled_back"] != true {
		t.Fatalf("expected rolled back RBAC request, got %+v", rolledBack)
	}
	rollbackAudit := rolledBackData["audit_log"].(map[string]any)
	if rollbackAudit["action"] != "admin.rbac.change_rolled_back" || rollbackAudit["target_type"] != "admin_rbac_role" || rollbackAudit["target_id"] != RoleFinanceAdmin {
		t.Fatalf("expected RBAC rollback audit log, got %+v", rollbackAudit)
	}
	authPutJSON(t, server.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_4"), `{"default_refund_strategy":"balance_first"}`, http.StatusOK)
	resetAdminRBACRoleScopeOverrides()
	rollbackRestoredServer := httptest.NewServer(NewRouter(store))
	defer rollbackRestoredServer.Close()
	authPutJSON(t, rollbackRestoredServer.URL+"/api/admin/refund-settings", roleToken(RoleFinanceAdmin, "finance_5"), `{"default_refund_strategy":"balance_first"}`, http.StatusOK)
	postRollbackAppliedChanges := authGetJSON(t, server.URL+"/api/admin/rbac/change-requests?status=applied&limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	postRollbackAppliedItems := postRollbackAppliedChanges["data"].(map[string]any)["items"].([]any)
	if len(postRollbackAppliedItems) != 0 {
		t.Fatalf("expected rolled back request to leave applied filter, got %+v", postRollbackAppliedChanges)
	}
	rolledBackChanges := authGetJSON(t, server.URL+"/api/admin/rbac/change-requests?status=rolled_back&limit=5", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusOK)
	rolledBackItems := rolledBackChanges["data"].(map[string]any)["items"].([]any)
	if len(rolledBackItems) != 1 || rolledBackItems[0].(map[string]any)["status"] != "rolled_back" || rolledBackChanges["data"].(map[string]any)["rolled_back_count"] != float64(1) {
		t.Fatalf("expected rolled back request in RBAC ledger, got %+v", rolledBackChanges)
	}
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/review", roleToken(RoleSuperAdmin, "super_3"), `{"decision":"reject","reason":"already reviewed"}`, http.StatusConflict)
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/apply", roleToken(RoleSuperAdmin, "super_3"), `{"reason":"already applied"}`, http.StatusConflict)
	authPostJSON(t, server.URL+"/api/admin/rbac/change-requests/"+changeRequestID+"/rollback", roleToken(RoleSuperAdmin, "super_2"), `{"reason":"already rolled back"}`, http.StatusConflict)
}

func TestAdminRefundSettingsHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &refundSettingsAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPutJSON(t, server.URL+"/api/admin/refund-settings", adminToken("admin_1"), `{"default_refund_strategy":"original_route_first"}`, http.StatusOK)
	if body["data"].(map[string]any)["default_refund_strategy"] != platform.RefundStrategyOriginalFirst {
		t.Fatalf("expected original-route refund setting, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected refund settings HTTP handler to call atomic audit repository path")
	}
	if store.recordAuditCalled {
		t.Fatal("refund settings HTTP handler must not call standalone RecordAuditLog after settings write")
	}
	if store.atomicAudit.Action != "admin.refund_settings.updated" || store.atomicAudit.TargetType != "refund_settings" || store.atomicAudit.TargetID != "default" {
		t.Fatalf("expected refund settings audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleAdmin || store.atomicAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in atomic audit request, got %+v", store.atomicAudit)
	}
}

func TestAdminRefundOrderHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &refundOrderAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/orders/ord_atomic/refund", adminToken("admin_1"), `{"reason":"商品售罄","idempotency_key":"refund_atomic_http"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["refund"].(map[string]any)["id"] != "rfd_atomic_http" {
		t.Fatalf("expected atomic refund response, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected admin refund HTTP handler to call atomic audit repository path")
	}
	if store.refundOrderCalled {
		t.Fatal("admin refund HTTP handler must not call standalone RefundOrder before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("admin refund HTTP handler must not call standalone RecordAuditLog after refund write")
	}
	if store.atomicReq.OrderID != "ord_atomic" || store.atomicReq.ActorID != "admin_1" || store.atomicReq.ActorRole != RoleAdmin {
		t.Fatalf("expected order path and admin principal in atomic refund request, got %+v", store.atomicReq)
	}
	if store.atomicAudit.Action != "admin.order.refunded" || store.atomicAudit.TargetType != "order" || store.atomicAudit.TargetID != "ord_atomic" {
		t.Fatalf("expected refund audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleAdmin || store.atomicAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in atomic refund audit request, got %+v", store.atomicAudit)
	}
}

func TestCreateMerchantInviteHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &merchantInviteAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/admin/merchant-invites", adminToken("admin_1"), `{}`, http.StatusCreated)
	if body["data"].(map[string]any)["token"] != "mi_atomic_http" {
		t.Fatalf("expected atomic merchant invite response, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected merchant invite HTTP handler to call atomic audit repository path")
	}
	if store.createMerchantInviteCalled {
		t.Fatal("merchant invite HTTP handler must not call standalone CreateMerchantInvite before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("merchant invite HTTP handler must not call standalone RecordAuditLog after invite write")
	}
	if store.atomicReq.AdminID != "admin_1" {
		t.Fatalf("expected admin principal to create merchant invite, got %+v", store.atomicReq)
	}
	if store.atomicAudit.Action != "admin.merchant_invite.created" || store.atomicAudit.TargetType != "merchant_invite" || store.atomicAudit.TargetID != "pending" {
		t.Fatalf("expected pending merchant invite audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleAdmin || store.atomicAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in merchant invite audit request, got %+v", store.atomicAudit)
	}
}

func TestCreateRiderInviteHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &riderInviteAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/station-manager/rider-invites", stationManagerToken("station_manager_1"), `{"type":"rider","station_id":"station_1"}`, http.StatusCreated)
	if body["data"].(map[string]any)["token"] != "ri_atomic_http" {
		t.Fatalf("expected atomic rider invite response, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected rider invite HTTP handler to call atomic audit repository path")
	}
	if store.createRiderInviteCalled {
		t.Fatal("rider invite HTTP handler must not call standalone CreateRiderInvite before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("rider invite HTTP handler must not call standalone RecordAuditLog after invite write")
	}
	if store.atomicReq.CreatedByID != "station_manager_1" || store.atomicReq.CreatedByRole != RoleStationManager || store.atomicReq.Type != platform.RiderAccountRider || store.atomicReq.StationID != "station_1" {
		t.Fatalf("expected station manager principal and station scope in atomic rider invite request, got %+v", store.atomicReq)
	}
	if store.atomicAudit.Action != "admin.rider_invite.created" || store.atomicAudit.TargetType != "rider_invite" || store.atomicAudit.TargetID != "pending" {
		t.Fatalf("expected pending rider invite audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleStationManager || store.atomicAudit.ActorID != "station_manager_1" {
		t.Fatalf("expected station manager principal in rider invite audit request, got %+v", store.atomicAudit)
	}
}

func TestReviewAfterSalesHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &reviewAfterSalesAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/after-sales/asr_atomic/review", merchantToken("merchant_1"), `{"decision":"approve","reason":"确认漏送"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["after_sales"].(map[string]any)["id"] != "asr_atomic" || data["refund"].(map[string]any)["id"] != "rfd_after_sales_atomic" {
		t.Fatalf("expected atomic after-sales review response, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected after-sales review HTTP handler to call atomic audit repository path")
	}
	if store.reviewAfterSalesCalled {
		t.Fatal("after-sales review HTTP handler must not call standalone ReviewAfterSales before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("after-sales review HTTP handler must not call standalone RecordAuditLog after review write")
	}
	if store.atomicReq.RequestID != "asr_atomic" || store.atomicReq.ActorID != "merchant_1" || store.atomicReq.ActorRole != RoleMerchant {
		t.Fatalf("expected request path and merchant principal in atomic after-sales review request, got %+v", store.atomicReq)
	}
	if store.atomicAudit.Action != "after_sales.reviewed" || store.atomicAudit.TargetType != "after_sales" || store.atomicAudit.TargetID != "asr_atomic" {
		t.Fatalf("expected after-sales review audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleMerchant || store.atomicAudit.ActorID != "merchant_1" {
		t.Fatalf("expected merchant principal in atomic after-sales review audit request, got %+v", store.atomicAudit)
	}
}

func TestAdminOperationsSnapshotHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	order, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_1", Type: platform.OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_http_admin_snapshot"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_http_admin_snapshot"}); err != nil || paidOrder.Status != platform.StatusDispatching {
		t.Fatalf("expected paid order setup, order=%+v err=%v", paidOrder, err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/operations/snapshot?limit=3&lease_expiring_within_seconds=60&object_cleanup_grace_seconds=60", userToken("user_1"), http.StatusForbidden)
	body := authGetJSON(t, server.URL+"/api/admin/operations/snapshot?limit=3&lease_expiring_within_seconds=60&object_cleanup_grace_seconds=60", adminToken("admin_1"), http.StatusOK)
	data := body["data"].(map[string]any)
	counts := data["counts"].(map[string]any)
	if counts["total_orders"] != float64(1) || counts["dispatching_orders"] != float64(1) {
		t.Fatalf("expected order counts in operations snapshot, got %+v", body)
	}
	if len(data["orders"].([]any)) != 1 || len(data["merchants"].([]any)) == 0 || len(data["riders"].([]any)) == 0 || len(data["rider_performance"].([]any)) == 0 {
		t.Fatalf("expected P0 lists in operations snapshot, got %+v", body)
	}
	if data["refund_settings"].(map[string]any)["default_refund_strategy"] != platform.RefundStrategyBalanceFirst {
		t.Fatalf("expected refund settings in operations snapshot, got %+v", body)
	}
	if data["outbox_stats"].(map[string]any)["total"] == float64(0) {
		t.Fatalf("expected outbox stats in operations snapshot, got %+v", body)
	}
}

func TestAfterSalesHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	lat := 39.99
	lng := 116.48
	address, err := store.SaveAddress(platform.UserAddress{
		UserID:       "user_1",
		ContactName:  "张三",
		ContactPhone: "13800000000",
		City:         "北京",
		Detail:       "望京SOHO",
		Latitude:     &lat,
		Longitude:    &lng,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertCartItem(platform.UpsertCartItemRequest{UserID: "user_1", ShopID: "shop_1", ProductID: "prod_beef_rice", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	order, _, err := store.CheckoutCart(platform.CheckoutCartRequest{UserID: "user_1", ShopID: "shop_1", AddressID: address.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: order.AmountFen, IdempotencyKey: "credit_http_after_sales"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_http_after_sales"}); err != nil || paidOrder.Status != platform.StatusMerchantPending {
		t.Fatalf("expected paid merchant order, order=%+v err=%v", paidOrder, err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	createBody := authPostJSON(t, server.URL+"/api/after-sales", userToken("user_1"), `{"order_id":"`+order.ID+`","reason":"餐品漏送","requested_amount_fen":`+strconv.FormatInt(order.AmountFen, 10)+`,"evidence_urls":["https://cdn.test/after-sales.jpg"]}`, http.StatusCreated)
	request := createBody["data"].(map[string]any)
	requestID := request["id"].(string)
	if request["status"] != platform.AfterSalesPendingMerchant || request["user_id"] != "user_1" {
		t.Fatalf("expected pending after-sales request, got %+v", createBody)
	}
	if request["order_amount_fen"] != float64(order.AmountFen) || request["refundable_fen"] != float64(order.AmountFen) {
		t.Fatalf("expected after-sales response to expose refund window, got %+v", createBody)
	}

	userListBody := authGetJSON(t, server.URL+"/api/after-sales", userToken("user_1"), http.StatusOK)
	merchantListBody := authGetJSON(t, server.URL+"/api/merchant/after-sales", merchantToken("merchant_1"), http.StatusOK)
	adminListBody := authGetJSON(t, server.URL+"/api/admin/after-sales", adminToken("admin_1"), http.StatusOK)
	if len(userListBody["data"].([]any)) != 1 || len(merchantListBody["data"].([]any)) != 1 || len(adminListBody["data"].([]any)) != 1 {
		t.Fatalf("expected after-sales lists to expose request, user=%+v merchant=%+v admin=%+v", userListBody, merchantListBody, adminListBody)
	}
	uploadBody := authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/evidence/upload-ticket", userToken("user_1"), `{"file_name":"evidence.jpg","content_type":"image/jpeg","size_bytes":1024}`, http.StatusCreated)
	uploadTicket := uploadBody["data"].(map[string]any)
	if uploadTicket["ticket_id"] == "" || uploadTicket["method"] != "PUT" || !strings.Contains(uploadTicket["object_key"].(string), "after-sales/"+requestID+"/") || uploadTicket["max_size_bytes"] != float64(platform.AfterSalesEvidenceMaxBytes) {
		t.Fatalf("expected after-sales upload ticket, got %+v", uploadBody)
	}
	uploadCallbackBody := postJSON(t, server.URL+"/api/object-storage/upload-callback", `{"ticket_id":"`+uploadTicket["ticket_id"].(string)+`","object_key":"`+uploadTicket["object_key"].(string)+`","content_type":"image/jpeg","size_bytes":1024,"content_sha":"sha256:test"}`, http.StatusOK)
	if uploadCallbackBody["data"].(map[string]any)["status"] != platform.AfterSalesUploadTicketUploaded {
		t.Fatalf("expected object storage upload callback to mark ticket uploaded, got %+v", uploadCallbackBody)
	}
	authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/evidence/confirm", userToken("user_1"), `{"object_key":"after-sales/`+requestID+`/forged/evidence.jpg","content_type":"image/jpeg","size_bytes":1024}`, http.StatusBadRequest)
	confirmBody := authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/evidence/confirm", userToken("user_1"), `{"ticket_id":"`+uploadTicket["ticket_id"].(string)+`","object_key":"`+uploadTicket["object_key"].(string)+`","content_type":"image/jpeg","size_bytes":1024,"content_sha":"sha256:test"}`, http.StatusCreated)
	confirmData := confirmBody["data"].(map[string]any)
	if confirmData["evidence"].(map[string]any)["status"] != platform.AfterSalesEvidenceUploaded ||
		confirmData["event"].(map[string]any)["action"] != platform.AfterSalesActionEvidenceUploaded {
		t.Fatalf("expected confirmed after-sales evidence, got %+v", confirmBody)
	}
	evidenceListBody := authGetJSON(t, server.URL+"/api/after-sales/"+requestID+"/evidence", merchantToken("merchant_1"), http.StatusOK)
	if len(evidenceListBody["data"].([]any)) != 1 {
		t.Fatalf("expected merchant to see after-sales evidence, got %+v", evidenceListBody)
	}

	merchantEventBody := authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", merchantToken("merchant_1"), `{"action":"merchant_reply","message":"已核实后厨打包记录","attachments":["https://cdn.test/pack.jpg"]}`, http.StatusCreated)
	if merchantEventBody["data"].(map[string]any)["event"].(map[string]any)["action"] != platform.AfterSalesActionMerchantReply {
		t.Fatalf("expected merchant after-sales event, got %+v", merchantEventBody)
	}
	authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", adminToken("admin_1"), `{"action":"internal_note","message":"客服内部备注","visible_to_user":false}`, http.StatusCreated)
	userEventsBody := authGetJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", userToken("user_1"), http.StatusOK)
	adminEventsBody := authGetJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", adminToken("admin_1"), http.StatusOK)
	if len(userEventsBody["data"].([]any)) != 3 || len(adminEventsBody["data"].([]any)) != 4 {
		t.Fatalf("expected user-visible timeline and admin full audit log, user=%+v admin=%+v", userEventsBody, adminEventsBody)
	}

	authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/review", userToken("user_1"), `{"decision":"approve","reason":"越权"}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/review", merchantToken("merchant_2"), `{"decision":"approve","reason":"非本店"}`, http.StatusConflict)
	reviewBody := authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/review", merchantToken("merchant_1"), `{"decision":"approve","reason":"确认漏送"}`, http.StatusOK)
	reviewData := reviewBody["data"].(map[string]any)
	if reviewData["after_sales"].(map[string]any)["status"] != platform.AfterSalesRefunded ||
		reviewData["refund"].(map[string]any)["status"] != platform.RefundStatusSuccess ||
		reviewData["order"].(map[string]any)["status"] != platform.StatusRefunded ||
		reviewData["wallet_account"].(map[string]any)["balance_fen"] != float64(order.AmountFen) {
		t.Fatalf("expected review to approve and refund to balance, got %+v", reviewBody)
	}
}

func TestAdminObjectStorageCleanupHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	if err := store.ConfigureObjectStorage(platform.ObjectStorageConfig{
		Provider:       platform.ObjectStorageProviderMinIO,
		Bucket:         "after-sales-test",
		UploadBaseURL:  "https://minio.test/upload",
		PublicBaseURL:  "https://cdn.test/assets",
		SigningSecret:  "test-storage-secret",
		MaxUploadBytes: platform.AfterSalesEvidenceMaxBytes,
		TicketTTL:      time.Second,
	}); err != nil {
		t.Fatal(err)
	}
	order, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_1", Type: platform.OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_http_cleanup"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_http_cleanup"}); err != nil {
		t.Fatal(err)
	}
	request, err := store.CreateAfterSales(platform.CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := store.CreateAfterSalesEvidenceUpload(platform.CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "expired.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := ticket.ExpiresAt.Add(2 * time.Hour)

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/object-storage/cleanup-candidates?now="+now.Format(time.RFC3339)+"&grace_seconds=60", userToken("user_1"), http.StatusForbidden)
	body := authGetJSON(t, server.URL+"/api/admin/object-storage/cleanup-candidates?now="+now.Format(time.RFC3339)+"&grace_seconds=60", adminToken("admin_1"), http.StatusOK)
	candidates := body["data"].([]any)
	if len(candidates) != 1 || candidates[0].(map[string]any)["ticket_id"] != ticket.TicketID || candidates[0].(map[string]any)["reason"] != platform.AfterSalesObjectCleanupExpired {
		t.Fatalf("expected one cleanup candidate, got %+v", body)
	}
	failedBody := authPostJSON(t, server.URL+"/api/admin/object-storage/cleanup-failed", adminToken("admin_1"), `{"ticket_id":"`+ticket.TicketID+`","object_key":"`+ticket.ObjectKey+`","reason":"expired_unconfirmed","error":"delete denied","failed_at":"`+now.Add(30*time.Second).Format(time.RFC3339)+`"}`, http.StatusOK)
	failedData := failedBody["data"].(map[string]any)
	if failedData["cleanup_attempts"] != float64(1) || failedData["last_cleanup_error"] != "delete denied" {
		t.Fatalf("expected cleanup failure ledger to be recorded, got %+v", failedBody)
	}
	statsBody := authGetJSON(t, server.URL+"/api/admin/object-storage/cleanup-stats?now="+now.Format(time.RFC3339)+"&grace_seconds=60", adminToken("admin_1"), http.StatusOK)
	statsData := statsBody["data"].(map[string]any)
	if statsData["pending"] != float64(1) || statsData["failed"] != float64(1) || statsData["cleanup_attempts"] != float64(1) {
		t.Fatalf("expected cleanup stats to expose pending failure, got %+v", statsBody)
	}
	completeBody := authPostJSON(t, server.URL+"/api/admin/object-storage/cleanup-complete", adminToken("admin_1"), `{"ticket_id":"`+ticket.TicketID+`","object_key":"`+ticket.ObjectKey+`","reason":"expired_unconfirmed","deleted_at":"`+now.Add(time.Minute).Format(time.RFC3339)+`"}`, http.StatusOK)
	if completeBody["data"].(map[string]any)["status"] != platform.AfterSalesUploadTicketDeleted {
		t.Fatalf("expected cleanup complete to mark ticket deleted, got %+v", completeBody)
	}
	if completeBody["data"].(map[string]any)["last_cleanup_error"] != nil {
		t.Fatalf("expected cleanup complete to clear last cleanup error, got %+v", completeBody)
	}
}

func TestAdminObjectStorageCleanupCompleteHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &objectStorageCleanupAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/admin/object-storage/cleanup-complete", adminToken("admin_1"), `{"ticket_id":"aset_atomic_complete","object_key":"after-sales/asr_atomic/aft123/expired.jpg","reason":"expired_unconfirmed","deleted_at":"2026-05-23T10:00:00Z"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["id"] != "aset_atomic_complete" || data["status"] != platform.AfterSalesUploadTicketDeleted {
		t.Fatalf("expected atomic cleanup-complete response, got %+v", body)
	}
	if !store.completeAtomicAuditCalled {
		t.Fatal("expected object cleanup complete HTTP handler to call atomic audit repository path")
	}
	if store.completeCalled {
		t.Fatal("object cleanup complete HTTP handler must not call standalone CompleteObjectStorageCleanup before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("object cleanup complete HTTP handler must not call standalone RecordAuditLog after cleanup")
	}
	if store.completeReq.TicketID != "aset_atomic_complete" || store.completeReq.ObjectKey != "after-sales/asr_atomic/aft123/expired.jpg" || store.completeReq.Reason != platform.AfterSalesObjectCleanupExpired {
		t.Fatalf("expected cleanup-complete payload to reach atomic repository path, got %+v", store.completeReq)
	}
	if store.completeAudit.Action != "admin.object_cleanup.completed" || store.completeAudit.TargetType != "object_storage_ticket" || store.completeAudit.TargetID != "aset_atomic_complete" {
		t.Fatalf("expected cleanup-complete audit target from HTTP handler, got %+v", store.completeAudit)
	}
	if store.completeAudit.ActorType != RoleAdmin || store.completeAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in cleanup-complete audit request, got %+v", store.completeAudit)
	}
}

func TestAdminObjectStorageCleanupFailedHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &objectStorageCleanupAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/admin/object-storage/cleanup-failed", adminToken("admin_1"), `{"ticket_id":"aset_atomic_failed","object_key":"after-sales/asr_atomic/aft456/rejected.jpg","reason":"scan_rejected","error":"delete denied","failed_at":"2026-05-23T10:01:00Z"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["id"] != "aset_atomic_failed" || data["cleanup_attempts"] != float64(1) || data["last_cleanup_error"] != "delete denied" {
		t.Fatalf("expected atomic cleanup-failed response, got %+v", body)
	}
	if !store.failureAtomicAuditCalled {
		t.Fatal("expected object cleanup failed HTTP handler to call atomic audit repository path")
	}
	if store.failureCalled {
		t.Fatal("object cleanup failed HTTP handler must not call standalone RecordObjectStorageCleanupFailure before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("object cleanup failed HTTP handler must not call standalone RecordAuditLog after cleanup")
	}
	if store.failureReq.TicketID != "aset_atomic_failed" || store.failureReq.ObjectKey != "after-sales/asr_atomic/aft456/rejected.jpg" || store.failureReq.Reason != platform.AfterSalesObjectCleanupRejected || store.failureReq.Error != "delete denied" {
		t.Fatalf("expected cleanup-failed payload to reach atomic repository path, got %+v", store.failureReq)
	}
	if store.failureAudit.Action != "admin.object_cleanup.failed" || store.failureAudit.TargetType != "object_storage_ticket" || store.failureAudit.TargetID != "aset_atomic_failed" {
		t.Fatalf("expected cleanup-failed audit target from HTTP handler, got %+v", store.failureAudit)
	}
	if store.failureAudit.ActorType != RoleAdmin || store.failureAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in cleanup-failed audit request, got %+v", store.failureAudit)
	}
}

func TestWechatMiniLoginIssuesSignedToken(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	login := postJSON(t, server.URL+"/api/auth/wechat-mini/login", `{"code":"wx_code_1","nickname":"小蓝"}`, http.StatusOK)
	data := login["data"].(map[string]any)
	token := data["access_token"].(string)
	if !strings.Contains(token, ".") || data["token_type"] != "Bearer" {
		t.Fatalf("expected signed bearer token, got %+v", data)
	}

	addressBody := authPostJSON(t, server.URL+"/api/user/addresses", token, `{"contact_name":"张三","contact_phone":"13800000000","city":"北京","detail":"望京SOHO","latitude":39.99,"longitude":116.48}`, http.StatusCreated)
	if addressBody["data"].(map[string]any)["user_id"] == "" {
		t.Fatalf("expected address to use token user, got %+v", addressBody)
	}
	authGetJSON(t, server.URL+"/api/user/addresses", token+".tampered", http.StatusUnauthorized)
}

func TestSessionAuthCanDisableDevBearerTokens(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules()), WithDevBearerAuth(false)))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/user/addresses", userToken("user_1"), http.StatusUnauthorized)

	login := postJSON(t, server.URL+"/api/auth/wechat-mini/login", `{"code":"wx_code_1","nickname":"小蓝"}`, http.StatusOK)
	token := login["data"].(map[string]any)["access_token"].(string)
	addresses := authGetJSON(t, server.URL+"/api/user/addresses", token, http.StatusOK)
	if !addresses["success"].(bool) {
		t.Fatalf("expected session-backed token to authorize, got %+v", addresses)
	}
}

func TestLogoutRevokesCurrentSessionToken(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules()), WithDevBearerAuth(false)))
	defer server.Close()

	login := postJSON(t, server.URL+"/api/auth/wechat-mini/login", `{"code":"wx_code_1","nickname":"小蓝"}`, http.StatusOK)
	token := login["data"].(map[string]any)["access_token"].(string)
	authGetJSON(t, server.URL+"/api/user/addresses", token, http.StatusOK)

	logout := authPostJSON(t, server.URL+"/api/auth/logout", token, `{}`, http.StatusOK)
	if logout["data"].(map[string]any)["revoked"] != true {
		t.Fatalf("expected logout to revoke current session, got %+v", logout)
	}
	authGetJSON(t, server.URL+"/api/user/addresses", token, http.StatusUnauthorized)
}

func TestWechatMiniLoginUsesResolvedProviderOpenID(t *testing.T) {
	server := httptest.NewServer(NewRouter(
		platform.NewStore(platform.DefaultHomeModules()),
		WithWechatMiniSessionResolver(staticWechatMiniResolver{session: WechatMiniSession{OpenID: "real_openid_1", UnionID: "union_1"}}),
	))
	defer server.Close()

	first := postJSON(t, server.URL+"/api/auth/wechat-mini/login", `{"code":"wx_code_1","nickname":"小蓝"}`, http.StatusOK)
	second := postJSON(t, server.URL+"/api/auth/wechat-mini/login", `{"code":"wx_code_2","nickname":"小蓝新版"}`, http.StatusOK)
	firstUser := first["data"].(map[string]any)["user"].(map[string]any)
	secondUser := second["data"].(map[string]any)["user"].(map[string]any)
	if firstUser["id"] != secondUser["id"] || second["data"].(map[string]any)["is_new_user"] != false {
		t.Fatalf("expected provider openid to keep stable user binding, first=%+v second=%+v", first, second)
	}
}

func TestWechatPrepayAndSignedCallbackHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":900}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)
	prepayBody := authPostJSON(t, server.URL+"/api/payments/wechat/prepay", userToken("user_1"), `{"order_id":"`+orderID+`"}`, http.StatusOK)
	prepay := prepayBody["data"].(map[string]any)["prepay"].(map[string]any)
	outTradeNo := prepay["out_trade_no"].(string)
	if outTradeNo == "" || prepay["amount_fen"] != float64(900) {
		t.Fatalf("unexpected prepay response: %+v", prepayBody)
	}

	callbackBody := `{"out_trade_no":"` + outTradeNo + `","transaction_id":"wx_tx_1","amount_fen":900}`
	postJSON(t, server.URL+"/api/payments/wechat/callback", callbackBody, http.StatusUnauthorized)
	callback := signedWechatCallbackJSON(t, server.URL+"/api/payments/wechat/callback", callbackBody, http.StatusOK)
	order := callback["data"].(map[string]any)["order"].(map[string]any)
	if order["status"] != platform.StatusDispatching || order["payment_method"] != platform.PaymentWechat {
		t.Fatalf("expected wechat callback to mark order paid, got %+v", callback)
	}
	duplicate := signedWechatCallbackJSON(t, server.URL+"/api/payments/wechat/callback", callbackBody, http.StatusOK)
	if duplicate["data"].(map[string]any)["order"].(map[string]any)["id"] != orderID {
		t.Fatalf("expected duplicate callback to be idempotent, got %+v", duplicate)
	}
}

func TestMerchantInviteRegisterAndQualificationHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	inviteBody := authPostJSON(t, server.URL+"/api/admin/merchant-invites", adminToken("admin_1"), `{}`, http.StatusCreated)
	token := inviteBody["data"].(map[string]any)["token"].(string)
	registerBody := postJSON(t, server.URL+"/api/auth/merchant/invite-register", `{"token":"`+token+`","display_name":"蓝海商户","account_type":"standard","password":"MerchantPass123"}`, http.StatusCreated)
	registerData := registerBody["data"].(map[string]any)
	merchantToken := registerData["access_token"].(string)
	profile := registerData["profile"].(map[string]any)
	merchantID := profile["account"].(map[string]any)["id"].(string)
	if merchantID == "" || profile["can_accept_orders"] != false {
		t.Fatalf("expected merchant profile with gated ordering, got %+v", registerBody)
	}

	expiresAt := "2027-05-21T00:00:00Z"
	authPostJSON(t, server.URL+"/api/merchant/qualifications", merchantToken, `{"type":"business_license","file_url":"https://example.test/license.jpg","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	qualificationBody := authPostJSON(t, server.URL+"/api/merchant/qualifications", merchantToken, `{"type":"health_certificate","file_url":"https://example.test/health.jpg","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	qualificationProfile := qualificationBody["data"].(map[string]any)
	if len(qualificationProfile["missing_qualifications"].([]any)) != 0 {
		t.Fatalf("expected no missing qualifications, got %+v", qualificationBody)
	}
	meBody := authGetJSON(t, server.URL+"/api/merchant/me", merchantToken, http.StatusOK)
	if meBody["data"].(map[string]any)["account"].(map[string]any)["id"] != merchantID {
		t.Fatalf("expected merchant me, got %+v", meBody)
	}
	postJSON(t, server.URL+"/api/auth/merchant/login", `{"account_id":"`+merchantID+`","password":"bad-password"}`, http.StatusUnauthorized)
	merchantLoginBody := postJSON(t, server.URL+"/api/auth/merchant/login", `{"account_id":"`+merchantID+`","password":"MerchantPass123"}`, http.StatusOK)
	merchantLoginToken := merchantLoginBody["data"].(map[string]any)["access_token"].(string)
	if merchantLoginToken == "" {
		t.Fatalf("expected merchant login token, got %+v", merchantLoginBody)
	}
	loginMeBody := authGetJSON(t, server.URL+"/api/merchant/me", merchantLoginToken, http.StatusOK)
	if loginMeBody["data"].(map[string]any)["account"].(map[string]any)["id"] != merchantID {
		t.Fatalf("expected merchant login token to authorize merchant me, got %+v", loginMeBody)
	}
}

func TestMerchantStaffAndMaterialsHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	staffBody := authGetJSON(t, server.URL+"/api/merchant/staff", merchantToken("merchant_1"), http.StatusOK)
	seededStaff := staffBody["data"].([]any)
	if len(seededStaff) != 1 || seededStaff[0].(map[string]any)["shop_id"] != "shop_1" {
		t.Fatalf("expected seeded merchant staff, got %+v", staffBody)
	}
	expiresAt := "2027-05-22T00:00:00Z"
	createStaffBody := authPostJSON(t, server.URL+"/api/merchant/staff", merchantToken("merchant_1"), `{"shop_id":"shop_1","name":"李四","phone":"13900000000","role":"kitchen","health_certificate_url":"https://example.test/staff-health.jpg","health_certificate_expires_at":"`+expiresAt+`"}`, http.StatusOK)
	createdStaff := createStaffBody["data"].(map[string]any)
	if createdStaff["id"] == "" || createdStaff["status"] != platform.MerchantStaffActive {
		t.Fatalf("expected created merchant staff, got %+v", createStaffBody)
	}
	authPostJSON(t, server.URL+"/api/merchant/staff", merchantToken("merchant_2"), `{"shop_id":"shop_1","name":"越权员工","phone":"13900000001","health_certificate_url":"https://example.test/staff-health.jpg","health_certificate_expires_at":"`+expiresAt+`"}`, http.StatusNotFound)

	materialsBody := authGetJSON(t, server.URL+"/api/merchant/materials", merchantToken("merchant_1"), http.StatusOK)
	seededMaterials := materialsBody["data"].([]any)
	if len(seededMaterials) != 1 || seededMaterials[0].(map[string]any)["type"] != "storefront_photo" {
		t.Fatalf("expected seeded merchant materials, got %+v", materialsBody)
	}
	createMaterialBody := authPostJSON(t, server.URL+"/api/merchant/materials", merchantToken("merchant_1"), `{"shop_id":"shop_1","type":"kitchen_photo","file_url":"https://example.test/kitchen.jpg","description":"后厨照","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	createdMaterial := createMaterialBody["data"].(map[string]any)
	if createdMaterial["id"] == "" || createdMaterial["status"] != "submitted" {
		t.Fatalf("expected submitted merchant material, got %+v", createMaterialBody)
	}
	authPostJSON(t, server.URL+"/api/merchant/materials", merchantToken("merchant_2"), `{"shop_id":"shop_1","type":"kitchen_photo","file_url":"https://example.test/kitchen.jpg","description":"越权材料","expires_at":"`+expiresAt+`"}`, http.StatusNotFound)

	meBody := authGetJSON(t, server.URL+"/api/merchant/me", merchantToken("merchant_1"), http.StatusOK)
	profile := meBody["data"].(map[string]any)
	if len(profile["staff"].([]any)) < 2 || len(profile["supplemental_materials"].([]any)) < 2 {
		t.Fatalf("expected merchant profile to include staff and materials, got %+v", meBody)
	}
}

func TestAdminBootstrapLoginIssuesSignedToken(t *testing.T) {
	server := httptest.NewServer(NewRouter(
		platform.NewStore(platform.DefaultHomeModules()),
		WithDevBearerAuth(false),
		WithAdminLoginCredential("admin_1", "AdminPass123"),
	))
	defer server.Close()

	authPostJSON(t, server.URL+"/api/admin/merchant-invites", adminToken("admin_1"), `{}`, http.StatusUnauthorized)
	postJSON(t, server.URL+"/api/auth/admin/login", `{"account_id":"admin_1","password":"bad-password"}`, http.StatusUnauthorized)
	login := postJSON(t, server.URL+"/api/auth/admin/login", `{"account_id":"admin_1","password":"AdminPass123"}`, http.StatusOK)
	data := login["data"].(map[string]any)
	token := data["access_token"].(string)
	admin := data["admin"].(map[string]any)
	if token == "" || admin["id"] != "admin_1" || admin["role"] != RoleAdmin {
		t.Fatalf("expected admin bootstrap login token, got %+v", login)
	}
	invite := authPostJSON(t, server.URL+"/api/admin/merchant-invites", token, `{}`, http.StatusCreated)
	if invite["data"].(map[string]any)["created_by_subject_id"] != "admin_1" {
		t.Fatalf("expected admin login token to create merchant invite, got %+v", invite)
	}
}

func TestRiderInviteRegisterHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	stationInviteBody := authPostJSON(t, server.URL+"/api/station-manager/rider-invites", stationManagerToken("station_manager_1"), `{"type":"rider","station_id":"station_1"}`, http.StatusCreated)
	stationInvite := stationInviteBody["data"].(map[string]any)
	if stationInvite["type"] != platform.RiderAccountRider || stationInvite["created_by_subject_type"] != RoleStationManager {
		t.Fatalf("expected station manager rider invite, got %+v", stationInviteBody)
	}
	riderTokenValue := stationInvite["token"].(string)
	riderRegisterBody := postJSON(t, server.URL+"/api/auth/rider/invite-register", `{"token":"`+riderTokenValue+`","password":"RiderPass123"}`, http.StatusCreated)
	riderData := riderRegisterBody["data"].(map[string]any)
	rider := riderData["rider"].(map[string]any)
	if riderData["access_token"].(string) == "" || rider["type"] != platform.RiderAccountRider || rider["station_id"] != "station_1" || rider["deposit_status"] != platform.DepositStatusUnpaid {
		t.Fatalf("expected invited rider registration, got %+v", riderRegisterBody)
	}
	riderID := rider["id"].(string)
	postJSON(t, server.URL+"/api/auth/rider/login", `{"account_id":"`+riderID+`","password":"bad-password"}`, http.StatusUnauthorized)
	riderLoginBody := postJSON(t, server.URL+"/api/auth/rider/login", `{"account_id":"`+riderID+`","password":"RiderPass123"}`, http.StatusOK)
	riderLoginToken := riderLoginBody["data"].(map[string]any)["access_token"].(string)
	if riderLoginToken == "" {
		t.Fatalf("expected rider login token, got %+v", riderLoginBody)
	}
	riderDepositBody := authGetJSON(t, server.URL+"/api/rider/deposit", riderLoginToken, http.StatusOK)
	if riderDepositBody["data"].(map[string]any)["subject_id"] != riderID {
		t.Fatalf("expected rider login token to authorize rider deposit, got %+v", riderDepositBody)
	}
	postJSON(t, server.URL+"/api/auth/rider/invite-register", `{"token":"`+riderTokenValue+`","password":"RiderPass123"}`, http.StatusBadRequest)

	authPostJSON(t, server.URL+"/api/station-manager/rider-invites", stationManagerToken("station_manager_1"), `{"type":"station_manager","station_id":"station_1"}`, http.StatusBadRequest)
	adminInviteBody := authPostJSON(t, server.URL+"/api/admin/rider-invites", adminToken("admin_1"), `{"type":"station_manager","station_id":"station_2"}`, http.StatusCreated)
	adminInvite := adminInviteBody["data"].(map[string]any)
	if adminInvite["type"] != platform.RiderAccountStationManager || adminInvite["created_by_subject_type"] != RoleAdmin {
		t.Fatalf("expected admin station manager invite, got %+v", adminInviteBody)
	}
	managerRegisterBody := postJSON(t, server.URL+"/api/auth/rider/invite-register", `{"token":"`+adminInvite["token"].(string)+`","password":"StationPass123"}`, http.StatusCreated)
	managerData := managerRegisterBody["data"].(map[string]any)
	manager := managerData["rider"].(map[string]any)
	if managerData["access_token"].(string) == "" || manager["type"] != platform.RiderAccountStationManager || manager["station_id"] != "station_2" {
		t.Fatalf("expected invited station manager registration, got %+v", managerRegisterBody)
	}
	managerID := manager["id"].(string)
	managerLoginBody := postJSON(t, server.URL+"/api/auth/rider/login", `{"account_id":"`+managerID+`","password":"StationPass123"}`, http.StatusOK)
	managerLoginToken := managerLoginBody["data"].(map[string]any)["access_token"].(string)
	stationRidersBody := authGetJSON(t, server.URL+"/api/station-manager/riders", managerLoginToken, http.StatusOK)
	if stationRidersBody["success"] != true {
		t.Fatalf("expected station manager login token to authorize station riders, got %+v", stationRidersBody)
	}
}

func TestMerchantProductManagementHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	productsBody := authGetJSON(t, server.URL+"/api/merchant/products?shop_id=shop_1", merchantToken("merchant_1"), http.StatusOK)
	if len(productsBody["data"].([]any)) != 2 {
		t.Fatalf("expected seeded merchant products, got %+v", productsBody)
	}
	createBody := authPostJSON(t, server.URL+"/api/merchant/products", merchantToken("merchant_1"), `{"shop_id":"shop_1","name":"轻食鸡胸饭","image_url":"/assets/mock/chicken-rice.jpg","description":"鸡胸肉、糙米、蔬菜。","ingredient_list":["鸡胸肉","糙米","蔬菜"],"price_fen":2299,"stock_count":20}`, http.StatusOK)
	product := createBody["data"].(map[string]any)
	productID := product["id"].(string)
	if productID == "" || product["status"] != platform.ProductStatusActive {
		t.Fatalf("expected created active product, got %+v", createBody)
	}
	authPostJSON(t, server.URL+"/api/merchant/products/"+productID+"/status", merchantToken("merchant_2"), `{"status":"removed"}`, http.StatusNotFound)
	statusBody := authPostJSON(t, server.URL+"/api/merchant/products/"+productID+"/status", merchantToken("merchant_1"), `{"status":"sold_out"}`, http.StatusOK)
	if statusBody["data"].(map[string]any)["status"] != platform.ProductStatusSoldOut {
		t.Fatalf("expected sold out status, got %+v", statusBody)
	}
	publicProductsBody := getJSON(t, server.URL+"/api/shops/shop_1/products", http.StatusOK)
	for _, item := range publicProductsBody["data"].([]any) {
		if item.(map[string]any)["id"] == productID {
			t.Fatalf("expected sold-out product to be hidden from public list, got %+v", publicProductsBody)
		}
	}
}

func TestGroupbuyVoucherRedeemHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	dealsBody := getJSON(t, server.URL+"/api/shops/shop_1/groupbuy-deals", http.StatusOK)
	deals := dealsBody["data"].([]any)
	if len(deals) != 1 || deals[0].(map[string]any)["id"] != "deal_two_person_set" {
		t.Fatalf("expected groupbuy deal, got %+v", dealsBody)
	}
	orderBody := authPostJSON(t, server.URL+"/api/groupbuy/orders", userToken("user_1"), `{"shop_id":"shop_1","deal_id":"deal_two_person_set","quantity":1}`, http.StatusCreated)
	order := orderBody["data"].(map[string]any)
	orderID := order["id"].(string)
	if order["type"] != platform.OrderTypeGroupbuy || order["status"] != platform.StatusPendingPayment {
		t.Fatalf("expected pending groupbuy order, got %+v", orderBody)
	}
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":3999,"idempotency_key":"credit_groupbuy_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	payBody := authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_groupbuy_http"}`, http.StatusOK)
	if payBody["data"].(map[string]any)["order"].(map[string]any)["status"] != platform.StatusVoucherIssued {
		t.Fatalf("expected voucher issued after pay, got %+v", payBody)
	}
	vouchersBody := authGetJSON(t, server.URL+"/api/groupbuy/vouchers", userToken("user_1"), http.StatusOK)
	vouchers := vouchersBody["data"].([]any)
	if len(vouchers) != 1 || vouchers[0].(map[string]any)["status"] != platform.GroupbuyVoucherStatusIssued {
		t.Fatalf("expected issued groupbuy voucher, got %+v", vouchersBody)
	}
	qrPayload := vouchers[0].(map[string]any)["qr_payload"].(string)
	authPostJSON(t, server.URL+"/api/merchant/groupbuy/vouchers/scan", merchantToken("merchant_2"), `{"qr_payload":"`+qrPayload+`","method":"qr_scan"}`, http.StatusNotFound)
	scanBody := authPostJSON(t, server.URL+"/api/merchant/groupbuy/vouchers/scan", merchantToken("merchant_1"), `{"qr_payload":"`+qrPayload+`","method":"qr_scan"}`, http.StatusOK)
	voucher := scanBody["data"].(map[string]any)["voucher"].(map[string]any)
	redeemedOrder := scanBody["data"].(map[string]any)["order"].(map[string]any)
	if voucher["status"] != platform.GroupbuyVoucherRedeemed || redeemedOrder["status"] != platform.StatusCompleted {
		t.Fatalf("expected redeemed voucher and completed order, got %+v", scanBody)
	}
	authPostJSON(t, server.URL+"/api/merchant/groupbuy/vouchers/scan", merchantToken("merchant_1"), `{"qr_payload":"`+qrPayload+`","method":"qr_scan"}`, http.StatusConflict)
}

func TestAutoAssignAndRiderRejectHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":900}`, http.StatusCreated)
	order := orderBody["data"].(map[string]any)
	orderID := order["id"].(string)
	createdAt := order["created_at"].(string)
	createdAtTime, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		t.Fatal(err)
	}
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":900,"idempotency_key":"credit_assign_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_assign_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/rider/online", riderToken("rider_1"), `{"online":true,"capacity":2,"distance_meters":500}`, http.StatusOK)

	tooEarly := createdAtTime.Add(9 * time.Minute).Format(time.RFC3339Nano)
	authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/auto-assign", adminToken("admin_1"), `{"now":"`+tooEarly+`"}`, http.StatusConflict)
	assignNow := createdAtTime.Add(10 * time.Minute).Format(time.RFC3339Nano)
	assignBody := authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/auto-assign", adminToken("admin_1"), `{"now":"`+assignNow+`"}`, http.StatusOK)
	assignedOrder := assignBody["data"].(map[string]any)["order"].(map[string]any)
	decision := assignBody["data"].(map[string]any)["decision"].(map[string]any)
	if assignedOrder["rider_id"] != "rider_1" || decision["mode"] != platform.DispatchModeAutoAssign {
		t.Fatalf("expected rider_1 auto assignment, got %+v", assignBody)
	}
	rejectBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/reject-assignment", riderToken("rider_1"), `{}`, http.StatusOK)
	reassignedOrder := rejectBody["data"].(map[string]any)["order"].(map[string]any)
	nextDecision := rejectBody["data"].(map[string]any)["decision"].(map[string]any)
	if reassignedOrder["rider_id"] != "rider_2" || nextDecision["candidate_rider_id"] != "rider_2" {
		t.Fatalf("expected reject to assign next rider, got %+v", rejectBody)
	}
	eventsBody := authGetJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/events", adminToken("admin_1"), http.StatusOK)
	events := eventsBody["data"].([]any)
	if len(events) != 3 || events[0].(map[string]any)["type"] != "dispatch.auto_assign" || events[1].(map[string]any)["type"] != "dispatch.rejected" || events[2].(map[string]any)["type"] != "dispatch.auto_assign" {
		t.Fatalf("expected dispatch audit events through HTTP, got %+v", eventsBody)
	}
	stationEventsBody := authGetJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/events", stationManagerToken("station_manager_1"), http.StatusOK)
	if len(stationEventsBody["data"].([]any)) != 3 {
		t.Fatalf("expected station manager to read own station dispatch events, got %+v", stationEventsBody)
	}
}

func TestTimeoutReassignHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":1100}`, http.StatusCreated)
	order := orderBody["data"].(map[string]any)
	orderID := order["id"].(string)
	createdAt, err := time.Parse(time.RFC3339Nano, order["created_at"].(string))
	if err != nil {
		t.Fatal(err)
	}
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":1100,"idempotency_key":"credit_timeout_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_timeout_http"}`, http.StatusOK)

	assignAt := createdAt.Add(10 * time.Minute)
	assignBody := authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/auto-assign", adminToken("admin_1"), `{"now":"`+assignAt.Format(time.RFC3339Nano)+`"}`, http.StatusOK)
	if assignBody["data"].(map[string]any)["order"].(map[string]any)["rider_id"] != "rider_1" {
		t.Fatalf("expected rider_1 initial assignment, got %+v", assignBody)
	}

	tooEarly := assignAt.Add(59 * time.Second).Format(time.RFC3339Nano)
	authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/timeout-reassign", stationManagerToken("station_manager_1"), `{"now":"`+tooEarly+`","timeout_seconds":60}`, http.StatusConflict)

	timeoutAt := assignAt.Add(60 * time.Second).Format(time.RFC3339Nano)
	timeoutBody := authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/timeout-reassign", stationManagerToken("station_manager_1"), `{"now":"`+timeoutAt+`","timeout_seconds":60}`, http.StatusOK)
	reassignedOrder := timeoutBody["data"].(map[string]any)["order"].(map[string]any)
	decision := timeoutBody["data"].(map[string]any)["decision"].(map[string]any)
	if reassignedOrder["rider_id"] != "rider_2" || decision["candidate_rider_id"] != "rider_2" || decision["reason"] != "assignment_timeout" {
		t.Fatalf("expected timeout to reassign rider_2, got %+v", timeoutBody)
	}

	eventsBody := authGetJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/events", stationManagerToken("station_manager_1"), http.StatusOK)
	events := eventsBody["data"].([]any)
	if len(events) != 3 || events[1].(map[string]any)["type"] != "dispatch.timeout" || events[2].(map[string]any)["type"] != "dispatch.auto_assign" {
		t.Fatalf("expected timeout dispatch audit events, got %+v", eventsBody)
	}
}

func TestAdminCompensateOrderStateHTTPFlow(t *testing.T) {
	store := &compensateOrderStateStore{
		Store: platform.NewStore(platform.DefaultHomeModules()),
		result: &platform.CompensateOrderStateResult{
			Order: &platform.Order{
				ID:        "ord_1",
				UserID:    "user_1",
				Type:      platform.OrderTypeTakeout,
				Status:    platform.StatusDispatching,
				AmountFen: 1200,
			},
			Changed:          true,
			PreviousStatus:   platform.StatusPendingPayment,
			ExpectedStatus:   platform.StatusDispatching,
			Evidence:         []string{"wallet_transaction:wtx_1"},
			CompensationType: "order_state_replay",
		},
	}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authPostJSON(t, server.URL+"/api/admin/orders/ord_1/state/compensate", userToken("user_1"), `{}`, http.StatusForbidden)
	body := authPostJSON(t, server.URL+"/api/admin/orders/ord_1/state/compensate", adminToken("admin_1"), `{"now":"2026-05-22T12:00:00Z"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	order := data["order"].(map[string]any)
	if data["changed"] != true || data["previous_status"] != platform.StatusPendingPayment || data["expected_status"] != platform.StatusDispatching || order["status"] != platform.StatusDispatching {
		t.Fatalf("expected compensated order response, got %+v", body)
	}
	if store.req.OrderID != "ord_1" || store.req.ActorID != "admin_1" || store.req.Now.IsZero() {
		t.Fatalf("expected admin compensation request metadata, got %+v", store.req)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected order-state compensation HTTP handler to call atomic audit repository path")
	}
	if store.compensateOrderStateCalled {
		t.Fatal("order-state compensation HTTP handler must not call standalone CompensateOrderState before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("order-state compensation HTTP handler must not call standalone RecordAuditLog after compensation")
	}
	if store.atomicAudit.Action != "admin.order_state.compensated" || store.atomicAudit.TargetType != "order" || store.atomicAudit.TargetID != "ord_1" {
		t.Fatalf("expected order-state compensation audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleAdmin || store.atomicAudit.ActorID != "admin_1" {
		t.Fatalf("expected admin principal in order-state compensation audit request, got %+v", store.atomicAudit)
	}
}

func TestAdminOutboxHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":1200}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":1200,"idempotency_key":"credit_outbox_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_outbox_http"}`, http.StatusOK)

	authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid", userToken("user_1"), http.StatusForbidden)
	authGetJSON(t, server.URL+"/api/admin/outbox/stats?topic=order.paid", userToken("user_1"), http.StatusForbidden)
	body := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=1", adminToken("admin_1"), http.StatusOK)
	events := body["data"].([]any)
	if len(events) != 1 {
		t.Fatalf("expected one outbox event, got %+v", body)
	}
	event := events[0].(map[string]any)
	eventID := event["id"].(string)
	if event["topic"] != "order.paid" || event["aggregate_id"] != orderID || event["status"] != platform.OutboxStatusPending {
		t.Fatalf("expected pending order.paid event, got %+v", event)
	}
	eventCreatedAt, err := time.Parse(time.RFC3339Nano, event["created_at"].(string))
	if err != nil {
		t.Fatal(err)
	}
	claimNow := eventCreatedAt.Add(time.Second).UTC()
	renewNow := claimNow.Add(10 * time.Second)
	afterRenewNow := renewNow.Add(time.Second)

	authPostJSON(t, server.URL+"/api/admin/outbox/events/claim", userToken("user_1"), `{"topic":"order.paid","limit":1,"lease_owner":"relay-http","lease_seconds":30,"now":"`+claimNow.Format(time.RFC3339)+`"}`, http.StatusForbidden)
	claimedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/claim", adminToken("admin_1"), `{"topic":"order.paid","limit":1,"lease_owner":"relay-http","lease_seconds":30,"now":"`+claimNow.Format(time.RFC3339)+`"}`, http.StatusOK)
	claimed := claimedBody["data"].(map[string]any)
	claimedEvents := claimed["events"].([]any)
	if claimed["topic"] != "order.paid" || claimed["claimed"] != float64(1) || len(claimedEvents) != 1 || claimedEvents[0].(map[string]any)["id"] != eventID {
		t.Fatalf("expected admin claim to lease one outbox event, got %+v", claimedBody)
	}
	claimedEvent := claimedEvents[0].(map[string]any)
	if claimed["lease_owner"] != "relay-http" || claimedEvent["lease_owner"] != "relay-http" || claimedEvent["lease_expires_at"] != claimNow.Add(30*time.Second).Format(time.RFC3339) {
		t.Fatalf("expected claimed event lease metadata, got %+v", claimedBody)
	}
	authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/lease/renew", userToken("user_1"), `{"lease_owner":"relay-http","lease_seconds":60,"now":"`+renewNow.Format(time.RFC3339)+`"}`, http.StatusForbidden)
	renewedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/lease/renew", adminToken("admin_1"), `{"lease_owner":"relay-http","lease_seconds":60,"now":"`+renewNow.Format(time.RFC3339)+`"}`, http.StatusOK)
	renewed := renewedBody["data"].(map[string]any)
	if renewed["lease_owner"] != "relay-http" || renewed["lease_expires_at"] != renewNow.Add(time.Minute).Format(time.RFC3339) {
		t.Fatalf("expected renewed lease metadata, got %+v", renewedBody)
	}
	authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/lease/renew", adminToken("admin_1"), `{"lease_owner":"relay-other","lease_seconds":60,"now":"`+afterRenewNow.Format(time.RFC3339)+`"}`, http.StatusConflict)
	repeatedClaimBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/claim", adminToken("admin_1"), `{"topic":"order.paid","limit":1,"lease_owner":"relay-other","lease_seconds":30,"now":"`+afterRenewNow.Format(time.RFC3339)+`"}`, http.StatusOK)
	repeatedClaim := repeatedClaimBody["data"].(map[string]any)
	if repeatedClaim["claimed"] != float64(0) || len(repeatedClaim["events"].([]any)) != 0 {
		t.Fatalf("expected active lease to prevent duplicate claim, got %+v", repeatedClaimBody)
	}
	leasedStatsBody := authGetJSON(t, server.URL+"/api/admin/outbox/stats?topic=order.paid&now="+afterRenewNow.Format(time.RFC3339)+"&lease_expiring_within_seconds=60", adminToken("admin_1"), http.StatusOK)
	leasedStats := leasedStatsBody["data"].(map[string]any)
	if leasedStats["leased"] != float64(1) || leasedStats["ready"] != float64(0) || leasedStats["lease_expiring_within_seconds"] != float64(60) || leasedStats["lease_expiring_soon"] != float64(1) || leasedStats["next_lease_expires_in_seconds"] != float64(59) {
		t.Fatalf("expected leased outbox stats before ack, got %+v", leasedStatsBody)
	}
	leaseOwners := leasedStats["lease_owners"].([]any)
	if len(leaseOwners) != 1 || leaseOwners[0].(map[string]any)["owner"] != "relay-http" || leaseOwners[0].(map[string]any)["lease_expiring_soon"] != float64(1) {
		t.Fatalf("expected lease owner health stats, got %+v", leasedStatsBody)
	}

	failedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/failed", adminToken("admin_1"), `{"error":"relay down","retry_after_seconds":120,"now":"2026-05-22T12:00:00Z"}`, http.StatusOK)
	failed := failedBody["data"].(map[string]any)
	if failed["status"] != platform.OutboxStatusFailed || failed["attempts"] != float64(1) || failed["last_error"] != "relay down" {
		t.Fatalf("expected failed outbox event, got %+v", failedBody)
	}
	hiddenBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=10&now=2026-05-22T12:01:00Z", adminToken("admin_1"), http.StatusOK)
	if len(hiddenBody["data"].([]any)) != 0 {
		t.Fatalf("expected failed event to be hidden during backoff, got %+v", hiddenBody)
	}
	blockedStatsBody := authGetJSON(t, server.URL+"/api/admin/outbox/stats?topic=order.paid&now=2026-05-22T12:01:00Z", adminToken("admin_1"), http.StatusOK)
	blockedStats := blockedStatsBody["data"].(map[string]any)
	if blockedStats["total"] != float64(1) || blockedStats["failed"] != float64(1) || blockedStats["ready"] != float64(0) || blockedStats["blocked"] != float64(1) {
		t.Fatalf("expected blocked outbox stats during backoff, got %+v", blockedStatsBody)
	}
	blockedTopics := blockedStats["topics"].([]any)
	if len(blockedTopics) != 1 || blockedTopics[0].(map[string]any)["topic"] != "order.paid" || blockedTopics[0].(map[string]any)["blocked"] != float64(1) {
		t.Fatalf("expected per-topic blocked outbox stats, got %+v", blockedStatsBody)
	}
	authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/replay", userToken("user_1"), `{}`, http.StatusForbidden)
	replayedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/replay", adminToken("admin_1"), `{"now":"2026-05-22T12:01:30Z"}`, http.StatusOK)
	replayed := replayedBody["data"].(map[string]any)
	if replayed["status"] != platform.OutboxStatusPending || replayed["last_error"] != nil || replayed["available_at"] != "2026-05-22T12:01:30Z" {
		t.Fatalf("expected replayed outbox event to become pending-ready, got %+v", replayedBody)
	}
	readyAfterReplayBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=10&now=2026-05-22T12:01:31Z", adminToken("admin_1"), http.StatusOK)
	readyAfterReplay := readyAfterReplayBody["data"].([]any)
	if len(readyAfterReplay) != 1 || readyAfterReplay[0].(map[string]any)["id"] != eventID {
		t.Fatalf("expected replayed event to return to pending relay query, got %+v", readyAfterReplayBody)
	}
	authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/published", userToken("user_1"), `{}`, http.StatusForbidden)
	publishedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/published", adminToken("admin_1"), `{"published_at":"2026-05-22T12:02:00Z"}`, http.StatusOK)
	published := publishedBody["data"].(map[string]any)
	if published["status"] != platform.OutboxStatusPublished || published["published_at"] == "" {
		t.Fatalf("expected published outbox event, got %+v", publishedBody)
	}
	authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/replay", adminToken("admin_1"), `{"now":"2026-05-22T12:02:30Z"}`, http.StatusConflict)
	publishedStatsBody := authGetJSON(t, server.URL+"/api/admin/outbox/stats?topic=order.paid&now=2026-05-22T12:03:00Z", adminToken("admin_1"), http.StatusOK)
	publishedStats := publishedStatsBody["data"].(map[string]any)
	if publishedStats["total"] != float64(1) || publishedStats["published"] != float64(1) || publishedStats["ready"] != float64(0) || publishedStats["blocked"] != float64(0) {
		t.Fatalf("expected published outbox stats after ack, got %+v", publishedStatsBody)
	}
	auditBody := authGetJSON(t, server.URL+"/api/admin/audit-logs?target_type=outbox_event&target_id="+eventID+"&limit=10", adminToken("admin_1"), http.StatusOK)
	auditLogs := auditBody["data"].([]any)
	seenActions := map[string]bool{}
	for _, rawLog := range auditLogs {
		log := rawLog.(map[string]any)
		seenActions[log["action"].(string)] = true
		if log["integrity_verified"] != true {
			t.Fatalf("expected verified outbox audit log, got %+v", log)
		}
	}
	for _, action := range []string{"admin.outbox.lease_renewed", "admin.outbox.failed", "admin.outbox.replayed", "admin.outbox.published"} {
		if !seenActions[action] {
			t.Fatalf("expected outbox audit action %s, got %+v", action, auditBody)
		}
	}
}

func TestAdminOutboxDeadLetterHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":1200}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":1200,"idempotency_key":"credit_outbox_dead_letter_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_outbox_dead_letter_http"}`, http.StatusOK)

	body := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=1", adminToken("admin_1"), http.StatusOK)
	events := body["data"].([]any)
	if len(events) != 1 {
		t.Fatalf("expected one outbox event, got %+v", body)
	}
	eventID := events[0].(map[string]any)["id"].(string)

	firstFailureBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/failed", adminToken("admin_1"), `{"error":"relay down","retry_after_seconds":60,"max_attempts":2,"now":"2026-05-22T12:00:00Z"}`, http.StatusOK)
	firstFailure := firstFailureBody["data"].(map[string]any)
	if firstFailure["status"] != platform.OutboxStatusFailed || firstFailure["attempts"] != float64(1) {
		t.Fatalf("expected first failure to remain retryable, got %+v", firstFailureBody)
	}
	deadLetterBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/failed", adminToken("admin_1"), `{"error":"poison message","retry_after_seconds":60,"max_attempts":2,"now":"2026-05-22T12:01:00Z"}`, http.StatusOK)
	deadLetter := deadLetterBody["data"].(map[string]any)
	if deadLetter["status"] != platform.OutboxStatusDeadLetter || deadLetter["attempts"] != float64(2) || deadLetter["last_error"] != "poison message" {
		t.Fatalf("expected max attempts to isolate dead-letter event, got %+v", deadLetterBody)
	}
	pendingBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=10&now=2026-05-22T12:05:00Z", adminToken("admin_1"), http.StatusOK)
	if len(pendingBody["data"].([]any)) != 0 {
		t.Fatalf("expected dead-letter event to leave pending relay query, got %+v", pendingBody)
	}
	deadLettersBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?status=dead_letter&topic=order.paid&limit=10&now=2026-05-22T12:05:00Z", adminToken("admin_1"), http.StatusOK)
	deadLetters := deadLettersBody["data"].([]any)
	if len(deadLetters) != 1 || deadLetters[0].(map[string]any)["id"] != eventID {
		t.Fatalf("expected explicit dead-letter query to return event, got %+v", deadLettersBody)
	}
	statsBody := authGetJSON(t, server.URL+"/api/admin/outbox/stats?topic=order.paid&now=2026-05-22T12:05:00Z", adminToken("admin_1"), http.StatusOK)
	stats := statsBody["data"].(map[string]any)
	if stats["total"] != float64(1) || stats["dead_letter"] != float64(1) || stats["failed"] != float64(0) || stats["ready"] != float64(0) {
		t.Fatalf("expected dead-letter stats, got %+v", statsBody)
	}
	replayedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/"+eventID+"/replay", adminToken("admin_1"), `{"now":"2026-05-22T12:06:00Z"}`, http.StatusOK)
	replayed := replayedBody["data"].(map[string]any)
	if replayed["status"] != platform.OutboxStatusPending || replayed["attempts"] != float64(2) || replayed["last_error"] != nil {
		t.Fatalf("expected manual replay to release dead-letter event, got %+v", replayedBody)
	}
}

func TestAdminReplayOutboxEventsHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	createPaidOrder := func(suffix string) *platform.Order {
		t.Helper()
		userID := "user_" + suffix
		order, err := store.CreateOrder(platform.CreateOrderRequest{UserID: userID, Type: platform.OrderTypeTakeout, AmountFen: 1200})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: userID, AmountFen: 1200, IdempotencyKey: "credit_" + suffix}); err != nil {
			t.Fatal(err)
		}
		if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: userID, Password: "123456"}); err != nil {
			t.Fatal(err)
		}
		_, paidOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: userID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_" + suffix})
		if err != nil {
			t.Fatal(err)
		}
		return paidOrder
	}

	blockedOrder := createPaidOrder("outbox_http_batch_blocked")
	readyOrder := createPaidOrder("outbox_http_batch_ready")
	publishedOrder := createPaidOrder("outbox_http_batch_published")
	events, err := store.OutboxEvents(platform.OutboxEventsRequest{Topic: "order.paid", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	eventsByAggregateID := map[string]platform.OutboxEvent{}
	for _, event := range events {
		eventsByAggregateID[event.AggregateID] = event
	}
	blockedEvent := eventsByAggregateID[blockedOrder.ID]
	readyEvent := eventsByAggregateID[readyOrder.ID]
	publishedEvent := eventsByAggregateID[publishedOrder.ID]
	if blockedEvent.ID == "" || readyEvent.ID == "" || publishedEvent.ID == "" {
		t.Fatalf("expected setup outbox events, got %+v", events)
	}
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	if _, err := store.MarkOutboxEventFailed(platform.MarkOutboxEventFailedRequest{EventID: blockedEvent.ID, Error: "relay down", RetryAfterSeconds: 300, Now: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkOutboxEventFailed(platform.MarkOutboxEventFailedRequest{EventID: readyEvent.ID, Error: "already due", RetryAfterSeconds: 30, Now: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkOutboxEventPublished(platform.MarkOutboxEventPublishedRequest{EventID: publishedEvent.ID, PublishedAt: now.Add(10 * time.Second)}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authPostJSON(t, server.URL+"/api/admin/outbox/events/replay", userToken("user_1"), `{"topic":"order.paid","limit":10,"now":"2026-05-22T12:01:00Z"}`, http.StatusForbidden)
	replayedBody := authPostJSON(t, server.URL+"/api/admin/outbox/events/replay", adminToken("admin_1"), `{"topic":"order.paid","limit":10,"now":"2026-05-22T12:01:00Z"}`, http.StatusOK)
	replayed := replayedBody["data"].(map[string]any)
	replayedEvents := replayed["events"].([]any)
	if replayed["topic"] != "order.paid" || replayed["limit"] != float64(10) || replayed["replayed"] != float64(1) || len(replayedEvents) != 1 {
		t.Fatalf("expected one batch replayed event, got %+v", replayedBody)
	}
	replayedEvent := replayedEvents[0].(map[string]any)
	if replayedEvent["id"] != blockedEvent.ID || replayedEvent["status"] != platform.OutboxStatusPending || replayedEvent["attempts"] != float64(1) || replayedEvent["available_at"] != "2026-05-22T12:01:00Z" {
		t.Fatalf("expected blocked event to become pending-ready, got %+v", replayedEvent)
	}

	readyAfterReplayBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=order.paid&limit=10&now=2026-05-22T12:01:01Z", adminToken("admin_1"), http.StatusOK)
	readyAfterReplay := readyAfterReplayBody["data"].([]any)
	readyByID := map[string]map[string]any{}
	for _, rawEvent := range readyAfterReplay {
		event := rawEvent.(map[string]any)
		readyByID[event["id"].(string)] = event
	}
	if len(readyByID) != 2 || readyByID[blockedEvent.ID] == nil || readyByID[readyEvent.ID] == nil {
		t.Fatalf("expected replayed and already-ready events in pending query, got %+v", readyAfterReplayBody)
	}
	if readyByID[readyEvent.ID]["available_at"] == "2026-05-22T12:01:00Z" || readyByID[readyEvent.ID]["last_error"] != "already due" {
		t.Fatalf("expected already-ready event to be skipped by batch replay, got %+v", readyByID[readyEvent.ID])
	}
	if readyByID[publishedEvent.ID] != nil {
		t.Fatalf("expected published event to stay out of pending query, got %+v", readyAfterReplayBody)
	}
}

func TestStationManagerManualDispatchHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":1200}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)
	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":1200,"idempotency_key":"credit_station_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_station_http"}`, http.StatusOK)

	ridersBody := authGetJSON(t, server.URL+"/api/station-manager/riders", stationManagerToken("station_manager_1"), http.StatusOK)
	riders := ridersBody["data"].([]any)
	if len(riders) != 2 || riders[0].(map[string]any)["type"] != platform.RiderAccountRider {
		t.Fatalf("expected station rider list, got %+v", ridersBody)
	}
	ordersBody := authGetJSON(t, server.URL+"/api/station-manager/orders", stationManagerToken("station_manager_1"), http.StatusOK)
	orders := ordersBody["data"].([]any)
	if len(orders) != 1 || orders[0].(map[string]any)["id"] != orderID {
		t.Fatalf("expected station order list, got %+v", ordersBody)
	}
	manualAssignBody := authPostJSON(t, server.URL+"/api/station-manager/dispatch/"+orderID+"/manual-assign", stationManagerToken("station_manager_1"), `{"rider_id":"rider_2"}`, http.StatusOK)
	assignedOrder := manualAssignBody["data"].(map[string]any)["order"].(map[string]any)
	decision := manualAssignBody["data"].(map[string]any)["decision"].(map[string]any)
	if assignedOrder["rider_id"] != "rider_2" || decision["mode"] != platform.DispatchModeManualAssign {
		t.Fatalf("expected manual station assignment to rider_2, got %+v", manualAssignBody)
	}
	taskConfigBody := authGetJSON(t, server.URL+"/api/station-manager/task-duration", stationManagerToken("station_manager_1"), http.StatusOK)
	if taskConfigBody["data"].(map[string]any)["daily_task_duration_minutes"] != float64(8*60) {
		t.Fatalf("expected default task duration config, got %+v", taskConfigBody)
	}
	savedTaskConfigBody := authPutJSON(t, server.URL+"/api/station-manager/task-duration", stationManagerToken("station_manager_1"), `{"daily_task_duration_minutes":420,"daily_fixed_order_count":28}`, http.StatusOK)
	savedConfig := savedTaskConfigBody["data"].(map[string]any)
	if savedConfig["daily_task_duration_minutes"] != float64(420) || savedConfig["daily_fixed_order_count"] != float64(28) {
		t.Fatalf("expected saved task config, got %+v", savedTaskConfigBody)
	}
	performanceBody := authGetJSON(t, server.URL+"/api/station-manager/rider-performance", stationManagerToken("station_manager_1"), http.StatusOK)
	performance := performanceBody["data"].([]any)
	if len(performance) != 2 || performance[0].(map[string]any)["dispatch_priority"] == float64(0) {
		t.Fatalf("expected station rider performance, got %+v", performanceBody)
	}
}

func TestDepositHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	riderDepositBody := authGetJSON(t, server.URL+"/api/rider/deposit", riderToken("rider_1"), http.StatusOK)
	if riderDepositBody["data"].(map[string]any)["status"] != platform.DepositStatusPaid {
		t.Fatalf("expected seeded rider paid deposit, got %+v", riderDepositBody)
	}
	wechatExemptBody := authPostJSON(t, server.URL+"/api/rider/deposit/wechat-exempt", riderToken("rider_2"), `{"application_id":"wx_exempt_http"}`, http.StatusOK)
	if wechatExemptBody["data"].(map[string]any)["deposit"].(map[string]any)["status"] != platform.DepositStatusWechatExemptApproved {
		t.Fatalf("expected rider wechat exemption, got %+v", wechatExemptBody)
	}
	refundBody := authPostJSON(t, server.URL+"/api/rider/deposit/refund-request", riderToken("rider_1"), `{}`, http.StatusOK)
	if refundBody["data"].(map[string]any)["deposit"].(map[string]any)["status"] != platform.DepositStatusRefundPending {
		t.Fatalf("expected rider refund pending, got %+v", refundBody)
	}
	merchantDepositBody := authGetJSON(t, server.URL+"/api/merchant/deposit", merchantToken("merchant_1"), http.StatusOK)
	if merchantDepositBody["data"].(map[string]any)["amount_fen"] != float64(platform.MerchantDepositAmountFen) {
		t.Fatalf("expected merchant deposit, got %+v", merchantDepositBody)
	}
	payMerchantDepositBody := authPostJSON(t, server.URL+"/api/merchant/deposit/pay", merchantToken("merchant_1"), `{"amount_fen":5000}`, http.StatusOK)
	if payMerchantDepositBody["data"].(map[string]any)["status"] != platform.DepositStatusPaid {
		t.Fatalf("expected paid merchant deposit, got %+v", payMerchantDepositBody)
	}
}

func TestShopAddressCartCheckoutHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	shopsBody := getJSON(t, server.URL+"/api/shops", http.StatusOK)
	shops := shopsBody["data"].([]any)
	if len(shops) == 0 || shops[0].(map[string]any)["id"] != "shop_1" {
		t.Fatalf("expected seeded shop, got %+v", shopsBody)
	}
	if shops[0].(map[string]any)["status"] != platform.ShopStatusActive {
		t.Fatalf("expected seeded shop to be active, got %+v", shopsBody)
	}

	productsBody := getJSON(t, server.URL+"/api/shops/shop_1/products", http.StatusOK)
	products := productsBody["data"].([]any)
	if len(products) == 0 || products[0].(map[string]any)["id"] != "prod_beef_rice" {
		t.Fatalf("expected seeded products, got %+v", productsBody)
	}

	addressBody := authPostJSON(t, server.URL+"/api/user/addresses", userToken("user_1"), `{"contact_name":"张三","contact_phone":"13800000000","city":"北京","detail":"望京SOHO","latitude":39.99,"longitude":116.48,"tag":"home","is_default":true}`, http.StatusCreated)
	addressID := addressBody["data"].(map[string]any)["id"].(string)

	cartBody := authPostJSON(t, server.URL+"/api/cart/items", userToken("user_1"), `{"shop_id":"shop_1","product_id":"prod_beef_rice","quantity":2}`, http.StatusOK)
	if cartBody["data"].(map[string]any)["payable_fen"] != float64(5598) {
		t.Fatalf("expected cart payable 5598, got %+v", cartBody)
	}
	cartGetBody := authGetJSON(t, server.URL+"/api/cart?shop_id=shop_1", userToken("user_1"), http.StatusOK)
	if cartGetBody["data"].(map[string]any)["items_total_fen"] != float64(5198) {
		t.Fatalf("expected cart items total 5198, got %+v", cartGetBody)
	}

	checkoutBody := authPostJSON(t, server.URL+"/api/orders/checkout", userToken("user_1"), `{"shop_id":"shop_1","address_id":"`+addressID+`","options":{"remark":"少放辣","tableware_count":2}}`, http.StatusCreated)
	order := checkoutBody["data"].(map[string]any)["order"].(map[string]any)
	orderID := order["id"].(string)
	if order["status"] != platform.StatusPendingPayment || order["amount_fen"] != float64(5598) {
		t.Fatalf("expected pending payment checkout order, got %+v", checkoutBody)
	}

	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":5598,"idempotency_key":"credit_checkout_http"}`, http.StatusOK)
	authPostJSON(t, server.URL+"/api/wallet/payment-password", userToken("user_1"), `{"password":"123456"}`, http.StatusOK)
	payBody := authPostJSON(t, server.URL+"/api/wallet/pay", userToken("user_1"), `{"order_id":"`+orderID+`","payment_password":"123456","idempotency_key":"pay_checkout_http"}`, http.StatusOK)
	if payBody["data"].(map[string]any)["order"].(map[string]any)["status"] != platform.StatusMerchantPending {
		t.Fatalf("expected checkout order to enter merchant pending, got %+v", payBody)
	}
	merchantMeBody := authGetJSON(t, server.URL+"/api/merchant/me", merchantToken("merchant_1"), http.StatusOK)
	if merchantMeBody["data"].(map[string]any)["account"].(map[string]any)["deposit_status"] != platform.DepositStatusPaid {
		t.Fatalf("expected seeded merchant profile to be active, got %+v", merchantMeBody)
	}
	authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/grab", riderToken("rider_1"), `{}`, http.StatusConflict)
	merchantOrdersBody := authGetJSON(t, server.URL+"/api/merchant/orders", merchantToken("merchant_1"), http.StatusOK)
	merchantOrders := merchantOrdersBody["data"].([]any)
	if len(merchantOrders) != 1 || merchantOrders[0].(map[string]any)["id"] != orderID {
		t.Fatalf("expected merchant order list to include checkout order, got %+v", merchantOrdersBody)
	}
	authPostJSON(t, server.URL+"/api/merchant/orders/"+orderID+"/accept", merchantToken("merchant_2"), `{}`, http.StatusConflict)
	acceptedBody := authPostJSON(t, server.URL+"/api/merchant/orders/"+orderID+"/accept", merchantToken("merchant_1"), `{}`, http.StatusOK)
	if acceptedBody["data"].(map[string]any)["status"] != platform.StatusPreparing {
		t.Fatalf("expected merchant accept to mark preparing, got %+v", acceptedBody)
	}
	readyBody := authPostJSON(t, server.URL+"/api/merchant/orders/"+orderID+"/ready", merchantToken("merchant_1"), `{}`, http.StatusOK)
	if readyBody["data"].(map[string]any)["status"] != platform.StatusDispatching {
		t.Fatalf("expected merchant ready to enter dispatching, got %+v", readyBody)
	}
	grabBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/grab", riderToken("rider_1"), `{}`, http.StatusOK)
	if grabBody["data"].(map[string]any)["status"] != platform.StatusRiderAssigned {
		t.Fatalf("expected rider to grab after merchant ready, got %+v", grabBody)
	}
	ordersBody := authGetJSON(t, server.URL+"/api/orders", userToken("user_1"), http.StatusOK)
	orders := ordersBody["data"].([]any)
	if len(orders) != 1 || orders[0].(map[string]any)["id"] != orderID {
		t.Fatalf("expected order list to include checkout order, got %+v", ordersBody)
	}
	detailBody := authGetJSON(t, server.URL+"/api/orders/"+orderID, userToken("user_1"), http.StatusOK)
	if detailBody["data"].(map[string]any)["id"] != orderID {
		t.Fatalf("expected order detail, got %+v", detailBody)
	}
}

func TestUserScopedEndpointsRequireAuthAndRejectCrossUserWrites(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	postJSON(t, server.URL+"/api/cart/items", `{"shop_id":"shop_1","product_id":"prod_beef_rice","quantity":1}`, http.StatusUnauthorized)
	authPostJSON(t, server.URL+"/api/cart/items", userToken("user_1"), `{"user_id":"user_2","shop_id":"shop_1","product_id":"prod_beef_rice","quantity":1}`, http.StatusForbidden)
}

func getJSON(t *testing.T, url string, expectedStatus int) map[string]any {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func authGetJSON(t *testing.T, url string, token string, expectedStatus int) map[string]any {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func postJSON(t *testing.T, url string, body string, expectedStatus int) map[string]any {
	t.Helper()
	return authPostJSON(t, url, "", body, expectedStatus)
}

func authPostJSON(t *testing.T, url string, token string, body string, expectedStatus int) map[string]any {
	t.Helper()
	return authRequestJSON(t, http.MethodPost, url, token, body, expectedStatus)
}

func authPutJSON(t *testing.T, url string, token string, body string, expectedStatus int) map[string]any {
	t.Helper()
	return authRequestJSON(t, http.MethodPut, url, token, body, expectedStatus)
}

func authRequestJSON(t *testing.T, method string, url string, token string, body string, expectedStatus int) map[string]any {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func userToken(userID string) string {
	return RoleUser + ":" + userID
}

func riderToken(riderID string) string {
	return RoleRider + ":" + riderID
}

func stationManagerToken(stationManagerID string) string {
	return RoleStationManager + ":" + stationManagerID
}

func merchantToken(merchantID string) string {
	return RoleMerchant + ":" + merchantID
}

func adminToken(adminID string) string {
	return RoleAdmin + ":" + adminID
}

func roleToken(role string, subjectID string) string {
	return role + ":" + subjectID
}

func securityAuditorToken(auditorID string) string {
	return RoleSecurityAuditor + ":" + auditorID
}

type compensateOrderStateStore struct {
	*platform.Store
	result                     *platform.CompensateOrderStateResult
	req                        platform.CompensateOrderStateRequest
	atomicAuditCalled          bool
	compensateOrderStateCalled bool
	recordAuditCalled          bool
	atomicAudit                platform.RecordAuditLogRequest
}

func (store *compensateOrderStateStore) CompensateOrderState(req platform.CompensateOrderStateRequest) (*platform.CompensateOrderStateResult, error) {
	store.compensateOrderStateCalled = true
	store.req = req
	return store.result, nil
}

func (store *compensateOrderStateStore) CompensateOrderStateWithAudit(req platform.CompensateOrderStateRequest, audit platform.RecordAuditLogRequest) (*platform.CompensateOrderStateResult, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.req = req
	store.atomicAudit = audit
	return store.result, &platform.AuditLog{ID: "aud_order_state_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}, nil
}

func (store *compensateOrderStateStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, nil
}

type refundSettingsAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled bool
	recordAuditCalled bool
	atomicAudit       platform.RecordAuditLogRequest
}

type merchantInviteAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled          bool
	createMerchantInviteCalled bool
	recordAuditCalled          bool
	atomicReq                  platform.CreateMerchantInviteRequest
	atomicAudit                platform.RecordAuditLogRequest
}

func (store *merchantInviteAtomicAuditStore) CreateMerchantInviteWithAudit(req platform.CreateMerchantInviteRequest, audit platform.RecordAuditLogRequest) (*platform.MerchantOnboardingInvite, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.atomicReq = req
	store.atomicAudit = audit
	invite := &platform.MerchantOnboardingInvite{
		Token:                "mi_atomic_http",
		Type:                 platform.OnboardingInviteMerchant,
		Status:               platform.OnboardingInviteActive,
		CreatedByAdminID:     req.AdminID,
		CreatedBySubjectType: RoleAdmin,
		CreatedBySubjectID:   req.AdminID,
		ExpiresAt:            time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
	}
	auditLog := &platform.AuditLog{ID: "aud_merchant_invite_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: "mi_atomic_http"}
	return invite, auditLog, nil
}

func (store *merchantInviteAtomicAuditStore) CreateMerchantInvite(req platform.CreateMerchantInviteRequest) (*platform.MerchantOnboardingInvite, error) {
	store.createMerchantInviteCalled = true
	return nil, platform.ErrInvalidArgument
}

func (store *merchantInviteAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

type riderInviteAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled       bool
	createRiderInviteCalled bool
	recordAuditCalled       bool
	atomicReq               platform.CreateRiderInviteRequest
	atomicAudit             platform.RecordAuditLogRequest
}

func (store *riderInviteAtomicAuditStore) CreateRiderInviteWithAudit(req platform.CreateRiderInviteRequest, audit platform.RecordAuditLogRequest) (*platform.MerchantOnboardingInvite, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.atomicReq = req
	store.atomicAudit = audit
	invite := &platform.MerchantOnboardingInvite{
		Token:                "ri_atomic_http",
		Type:                 req.Type,
		Status:               platform.OnboardingInviteActive,
		CreatedByAdminID:     req.CreatedByID,
		CreatedBySubjectType: req.CreatedByRole,
		CreatedBySubjectID:   req.CreatedByID,
		StationID:            req.StationID,
		ExpiresAt:            time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC),
	}
	auditLog := &platform.AuditLog{ID: "aud_rider_invite_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: "ri_atomic_http"}
	return invite, auditLog, nil
}

func (store *riderInviteAtomicAuditStore) CreateRiderInvite(req platform.CreateRiderInviteRequest) (*platform.MerchantOnboardingInvite, error) {
	store.createRiderInviteCalled = true
	return nil, platform.ErrInvalidArgument
}

func (store *riderInviteAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

func (store *refundSettingsAtomicAuditStore) SaveRefundSettingsWithAudit(req platform.SaveRefundSettingsRequest, audit platform.RecordAuditLogRequest) (*platform.RefundSettings, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.atomicAudit = audit
	return store.Store.SaveRefundSettingsWithAudit(req, audit)
}

func (store *refundSettingsAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

type refundOrderAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled bool
	refundOrderCalled bool
	recordAuditCalled bool
	atomicReq         platform.RefundOrderRequest
	atomicAudit       platform.RecordAuditLogRequest
}

func (store *refundOrderAtomicAuditStore) RefundOrderWithAudit(req platform.RefundOrderRequest, audit platform.RecordAuditLogRequest) (*platform.RefundTransaction, *platform.Order, *platform.WalletAccount, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.atomicReq = req
	store.atomicAudit = audit
	refund := &platform.RefundTransaction{
		ID:             "rfd_atomic_http",
		OrderID:        req.OrderID,
		UserID:         "user_atomic",
		AmountFen:      1200,
		Destination:    platform.RefundDestinationBalance,
		Status:         platform.RefundStatusSuccess,
		Reason:         req.Reason,
		IdempotencyKey: req.IdempotencyKey,
		CreatedAt:      time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
	}
	order := &platform.Order{ID: req.OrderID, UserID: refund.UserID, Status: platform.StatusRefunded, AmountFen: refund.AmountFen}
	account := &platform.WalletAccount{UserID: refund.UserID, Balance: refund.AmountFen, Version: 1}
	auditLog := &platform.AuditLog{ID: "aud_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}
	return refund, order, account, auditLog, nil
}

func (store *refundOrderAtomicAuditStore) RefundOrder(req platform.RefundOrderRequest) (*platform.RefundTransaction, *platform.Order, *platform.WalletAccount, error) {
	store.refundOrderCalled = true
	return nil, nil, nil, platform.ErrInvalidArgument
}

func (store *refundOrderAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

type reviewAfterSalesAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled      bool
	reviewAfterSalesCalled bool
	recordAuditCalled      bool
	atomicReq              platform.ReviewAfterSalesRequest
	atomicAudit            platform.RecordAuditLogRequest
}

func (store *reviewAfterSalesAtomicAuditStore) ReviewAfterSalesWithAudit(req platform.ReviewAfterSalesRequest, audit platform.RecordAuditLogRequest) (*platform.AfterSalesRequest, *platform.RefundTransaction, *platform.Order, *platform.WalletAccount, *platform.AuditLog, error) {
	store.atomicAuditCalled = true
	store.atomicReq = req
	store.atomicAudit = audit
	afterSales := &platform.AfterSalesRequest{
		ID:                 req.RequestID,
		OrderID:            "ord_after_sales_atomic",
		UserID:             "user_atomic",
		Status:             platform.AfterSalesRefunded,
		ReviewReason:       req.Reason,
		ReviewerID:         req.ActorID,
		ReviewerRole:       req.ActorRole,
		RefundID:           "rfd_after_sales_atomic",
		RequestedAmountFen: 600,
		RefundedAmountFen:  600,
		RefundableFen:      600,
	}
	refund := &platform.RefundTransaction{
		ID:             afterSales.RefundID,
		OrderID:        afterSales.OrderID,
		UserID:         afterSales.UserID,
		AmountFen:      600,
		Destination:    platform.RefundDestinationBalance,
		Status:         platform.RefundStatusSuccess,
		Reason:         req.Reason,
		IdempotencyKey: "after_sales_atomic_http",
		CreatedAt:      time.Date(2026, 5, 23, 12, 30, 0, 0, time.UTC),
	}
	order := &platform.Order{ID: afterSales.OrderID, UserID: afterSales.UserID, Status: platform.StatusMerchantPending, AmountFen: 1200}
	account := &platform.WalletAccount{UserID: afterSales.UserID, Balance: refund.AmountFen, Version: 1}
	auditLog := &platform.AuditLog{ID: "aud_after_sales_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}
	return afterSales, refund, order, account, auditLog, nil
}

func (store *reviewAfterSalesAtomicAuditStore) ReviewAfterSales(req platform.ReviewAfterSalesRequest) (*platform.AfterSalesRequest, *platform.RefundTransaction, *platform.Order, *platform.WalletAccount, error) {
	store.reviewAfterSalesCalled = true
	return nil, nil, nil, nil, platform.ErrInvalidArgument
}

func (store *reviewAfterSalesAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

type objectStorageCleanupAtomicAuditStore struct {
	*platform.Store
	completeAtomicAuditCalled bool
	failureAtomicAuditCalled  bool
	completeCalled            bool
	failureCalled             bool
	recordAuditCalled         bool
	completeReq               platform.ObjectStorageCleanupCompleteRequest
	failureReq                platform.ObjectStorageCleanupFailureRequest
	completeAudit             platform.RecordAuditLogRequest
	failureAudit              platform.RecordAuditLogRequest
}

func (store *objectStorageCleanupAtomicAuditStore) CompleteObjectStorageCleanupWithAudit(req platform.ObjectStorageCleanupCompleteRequest, audit platform.RecordAuditLogRequest) (*platform.AfterSalesEvidenceUploadTicket, *platform.AuditLog, error) {
	store.completeAtomicAuditCalled = true
	store.completeReq = req
	store.completeAudit = audit
	ticket := &platform.AfterSalesEvidenceUploadTicket{
		ID:            req.TicketID,
		ObjectKey:     req.ObjectKey,
		Status:        platform.AfterSalesUploadTicketDeleted,
		CleanupReason: req.Reason,
		DeletedAt:     req.DeletedAt,
	}
	auditLog := &platform.AuditLog{ID: "aud_object_cleanup_complete_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}
	return ticket, auditLog, nil
}

func (store *objectStorageCleanupAtomicAuditStore) CompleteObjectStorageCleanup(req platform.ObjectStorageCleanupCompleteRequest) (*platform.AfterSalesEvidenceUploadTicket, error) {
	store.completeCalled = true
	store.completeReq = req
	return nil, platform.ErrInvalidArgument
}

func (store *objectStorageCleanupAtomicAuditStore) RecordObjectStorageCleanupFailureWithAudit(req platform.ObjectStorageCleanupFailureRequest, audit platform.RecordAuditLogRequest) (*platform.AfterSalesEvidenceUploadTicket, *platform.AuditLog, error) {
	store.failureAtomicAuditCalled = true
	store.failureReq = req
	store.failureAudit = audit
	ticket := &platform.AfterSalesEvidenceUploadTicket{
		ID:                  req.TicketID,
		ObjectKey:           req.ObjectKey,
		Status:              platform.AfterSalesUploadTicketUploaded,
		CleanupReason:       req.Reason,
		CleanupAttempts:     1,
		LastCleanupError:    req.Error,
		LastCleanupFailedAt: req.FailedAt,
	}
	auditLog := &platform.AuditLog{ID: "aud_object_cleanup_failed_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}
	return ticket, auditLog, nil
}

func (store *objectStorageCleanupAtomicAuditStore) RecordObjectStorageCleanupFailure(req platform.ObjectStorageCleanupFailureRequest) (*platform.AfterSalesEvidenceUploadTicket, error) {
	store.failureCalled = true
	store.failureReq = req
	return nil, platform.ErrInvalidArgument
}

func (store *objectStorageCleanupAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
	store.recordAuditCalled = true
	return nil, platform.ErrInvalidArgument
}

func signedWechatCallbackJSON(t *testing.T, url string, body string, expectedStatus int) map[string]any {
	t.Helper()
	timestamp := "1716259200"
	nonce := "nonce_1"
	message := timestamp + "\n" + nonce + "\n" + body + "\n"
	mac := hmac.New(sha256.New, []byte("infinitech-wechat-pay-dev-secret"))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Wechatpay-Timestamp", timestamp)
	req.Header.Set("Wechatpay-Nonce", nonce)
	req.Header.Set("Wechatpay-Signature", signature)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

type staticWechatMiniResolver struct {
	session WechatMiniSession
	err     error
}

func (resolver staticWechatMiniResolver) Resolve(_ context.Context, _ string) (WechatMiniSession, error) {
	return resolver.session, resolver.err
}
