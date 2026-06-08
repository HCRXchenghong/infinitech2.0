package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

func TestPrescriptionImageUploadReviewHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	uploadBody := authPostJSON(t, server.URL+"/api/prescriptions/upload-ticket", userToken("user_1"), `{"product_id":"med_amoxicillin","file_name":"prescription.jpg","content_type":"image/jpeg","size_bytes":2048}`, http.StatusCreated)
	uploadTicket := uploadBody["data"].(map[string]any)
	if uploadTicket["ticket_id"] == "" || uploadTicket["method"] != "PUT" || !strings.HasPrefix(uploadTicket["object_key"].(string), "prescriptions/") {
		t.Fatalf("expected prescription upload ticket, got %+v", uploadBody)
	}

	confirmBody := authPostJSON(t, server.URL+"/api/prescriptions/upload-confirm", userToken("user_1"), `{"ticket_id":"`+uploadTicket["ticket_id"].(string)+`","object_key":"`+uploadTicket["object_key"].(string)+`","content_type":"image/jpeg","size_bytes":2048,"content_sha":"sha256:rx-http"}`, http.StatusCreated)
	confirmed := confirmBody["data"].(map[string]any)
	if confirmed["status"] != platform.AfterSalesUploadTicketConfirmed || confirmed["public_url"] == "" {
		t.Fatalf("expected confirmed prescription upload ticket, got %+v", confirmBody)
	}

	reviewBody := authPostJSON(t, server.URL+"/api/prescriptions", userToken("user_1"), `{"patient_name":"张三","product_id":"med_amoxicillin","prescription_image_ticket_id":"`+uploadTicket["ticket_id"].(string)+`","prescription_object_key":"`+uploadTicket["object_key"].(string)+`"}`, http.StatusCreated)
	review := reviewBody["data"].(map[string]any)
	if review["status"] != platform.PrescriptionReviewApproved || review["image_object_key"] != uploadTicket["object_key"] || review["image_upload_ticket_id"] != uploadTicket["ticket_id"] {
		t.Fatalf("expected prescription review to bind image ticket, got %+v", reviewBody)
	}
	if review["ocr_result"].(map[string]any)["status"] != platform.PrescriptionOCRMatched || review["archive"].(map[string]any)["archive_id"] == "" {
		t.Fatalf("expected prescription review to expose OCR and archive metadata, got %+v", reviewBody)
	}

	loadedBody := authGetJSON(t, server.URL+"/api/prescriptions/"+review["id"].(string), userToken("user_1"), http.StatusOK)
	if loadedBody["data"].(map[string]any)["image_object_key"] != uploadTicket["object_key"] {
		t.Fatalf("expected loaded review to retain image metadata, got %+v", loadedBody)
	}
	authGetJSON(t, server.URL+"/api/prescriptions/"+review["id"].(string), userToken("user_2"), http.StatusNotFound)

	queueBody := authGetJSON(t, server.URL+"/api/admin/prescriptions?status=approved", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	queue := queueBody["data"].([]any)
	if len(queue) != 1 || queue[0].(map[string]any)["id"] != review["id"] {
		t.Fatalf("expected prescription workbench queue, got %+v", queueBody)
	}
	rejectedBody := authPostJSON(t, server.URL+"/api/admin/prescriptions/"+review["id"].(string)+"/review", roleToken(RoleSupportAdmin, "support_1"), `{"decision":"rejected","reviewer_name":"陈药师","review_text":"处方影像与药品不匹配"}`, http.StatusOK)
	rejected := rejectedBody["data"].(map[string]any)
	if rejected["status"] != platform.PrescriptionReviewRejected || rejected["doctor_name"] != "陈药师" {
		t.Fatalf("expected pharmacist rejection, got %+v", rejectedBody)
	}
	rejectedUserBody := authGetJSON(t, server.URL+"/api/prescriptions/"+review["id"].(string), userToken("user_1"), http.StatusOK)
	if rejectedUserBody["data"].(map[string]any)["status"] != platform.PrescriptionReviewRejected {
		t.Fatalf("expected user prescription result to reflect rejection, got %+v", rejectedUserBody)
	}
}

func TestMedicineOrderInventoryHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	created := authPostJSON(t, server.URL+"/api/medicine/orders", userToken("user_1"), `{"items":[{"product_id":"med_cooling_patch","name":"退热贴","category":"感冒发热","price_fen":1290,"quantity":26}]}`, http.StatusCreated)
	items := created["data"].(map[string]any)["items"].([]any)
	firstItem := items[0].(map[string]any)
	if firstItem["stock_locked"] != true || int(firstItem["stock_remaining"].(float64)) != 0 {
		t.Fatalf("expected medicine order to expose stock lock, got %+v", created)
	}

	home := authGetJSON(t, server.URL+"/api/medicine/home", userToken("user_1"), http.StatusOK)
	products := home["data"].(map[string]any)["products"].([]any)
	if stock := medicineHTTPProductStock(products, "med_cooling_patch"); stock != 0 {
		t.Fatalf("expected locked stock to be visible on home, got %d", stock)
	}

	rejected := authPostJSON(t, server.URL+"/api/medicine/orders", userToken("user_1"), `{"items":[{"product_id":"med_cooling_patch","name":"退热贴","category":"感冒发热","price_fen":1290,"quantity":1}]}`, http.StatusConflict)
	if rejected["code"] != "INSUFFICIENT_STOCK" {
		t.Fatalf("expected insufficient stock conflict, got %+v", rejected)
	}
}

func medicineHTTPProductStock(products []any, productID string) int {
	for _, raw := range products {
		product, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if product["id"] == productID {
			return int(product["stock_count"].(float64))
		}
	}
	return -1
}

func chatHTTPThreadUnread(threads []any, threadID string) int {
	for _, raw := range threads {
		thread, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if thread["id"] == threadID {
			return int(thread["unread_count"].(float64))
		}
	}
	return -1
}

func chatHTTPThreadMuted(threads []any, threadID string) bool {
	for _, raw := range threads {
		thread, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if thread["id"] == threadID {
			return thread["muted"] == true
		}
	}
	return false
}

func chatHTTPMemberExists(members []any, subjectID string, displayName string) bool {
	for _, raw := range members {
		member, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if member["subject_id"] == subjectID && member["display_name"] == displayName {
			return true
		}
	}
	return false
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
	archiveObjects := map[string]string{}
	archiveServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		objectPath := strings.TrimPrefix(req.URL.Path, "/")
		body, ok := archiveObjects[objectPath]
		if !ok {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(body))
	}))
	defer archiveServer.Close()
	if err := store.ConfigureObjectStorage(platform.ObjectStorageConfig{
		Provider:                     platform.ObjectStorageProviderMinIO,
		Bucket:                       "audit-http-test",
		UploadBaseURL:                "https://upload.example.test",
		PublicBaseURL:                "https://cdn.example.test",
		HeadBaseURL:                  "https://cdn.example.test",
		AuditArchiveDownloadBaseURL:  archiveServer.URL,
		AuditArchiveMaxDownloadBytes: 1024 * 1024,
	}); err != nil {
		t.Fatal(err)
	}
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
	alertBody := authPostJSON(t, server.URL+"/api/admin/audit-logs/retention-alerts/emit", adminToken("admin_1"), `{"retention_days":2555,"hot_days":180,"integrity_sample_limit":10}`, http.StatusOK)
	alertData := alertBody["data"].(map[string]any)
	emission := alertData["emission"].(map[string]any)
	if emission["status"] != "emitted" || emission["alert_count"].(float64) == 0 || emission["outbox_event_id"] == "" {
		t.Fatalf("expected emitted retention alert event, got %+v", alertBody)
	}
	outboxEvent := alertData["outbox_event"].(map[string]any)
	if outboxEvent["topic"] != "audit.retention_alerts" || outboxEvent["event_type"] != "audit.retention_alerts.emitted" {
		t.Fatalf("expected audit retention alert outbox event, got %+v", outboxEvent)
	}
	alertAudit := alertData["audit_log"].(map[string]any)
	if alertAudit["action"] != "admin.audit_retention_alerts.emitted" || alertAudit["target_type"] != "audit_retention_alerts" {
		t.Fatalf("expected alert emission to be audited, got %+v", alertAudit)
	}
	archiveBody := authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/request", adminToken("admin_1"), `{"hot_days":1,"limit":10,"storage_prefix":"worm://audit-logs","now":"2027-05-24T12:00:00Z"}`, http.StatusOK)
	archiveData := archiveBody["data"].(map[string]any)
	archive := archiveData["archive"].(map[string]any)
	if archive["status"] != "requested" || archive["log_count"].(float64) == 0 || archive["manifest_hash"] == "" || archive["outbox_event_id"] == "" {
		t.Fatalf("expected audit archive request with manifest and outbox, got %+v", archiveBody)
	}
	archiveEvent := archiveData["outbox_event"].(map[string]any)
	if archiveEvent["topic"] != "audit.archive_requested" || archiveEvent["event_type"] != "audit.archive_requested" {
		t.Fatalf("expected audit archive outbox event, got %+v", archiveEvent)
	}
	archiveAudit := archiveData["audit_log"].(map[string]any)
	if archiveAudit["action"] != "admin.audit_archive.requested" || archiveAudit["target_type"] != "audit_archive" {
		t.Fatalf("expected audit archive request to be audited, got %+v", archiveAudit)
	}
	archiveObject := auditArchiveHTTPObjectBody(t, archive)
	archiveObjectHashBytes := sha256.Sum256([]byte(archiveObject))
	archiveObjectHash := hex.EncodeToString(archiveObjectHashBytes[:])
	archiveObjects[auditArchiveHTTPObjectPath(archive["storage_key"].(string))] = archiveObject
	archiveCompleteBody := authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/complete", adminToken("admin_1"), `{"archive_id":"`+archive["archive_id"].(string)+`","storage_key":"`+archive["storage_key"].(string)+`","manifest_algorithm":"sha256:v1","manifest_hash":"`+archive["manifest_hash"].(string)+`","content_hash":"`+archiveObjectHash+`","bytes":`+strconv.Itoa(len(archiveObject))+`,"object_lock_mode":"COMPLIANCE","retain_until":"2034-05-24T12:00:00Z","outbox_event_id":"`+archive["outbox_event_id"].(string)+`","uploaded_at":"2027-05-24T12:01:00Z"}`, http.StatusOK)
	archiveCompleteData := archiveCompleteBody["data"].(map[string]any)
	archiveCompletion := archiveCompleteData["archive"].(map[string]any)
	if archiveCompletion["status"] != "archived" || archiveCompletion["content_hash"] != archiveObjectHash || archiveCompletion["bytes"].(float64) != float64(len(archiveObject)) {
		t.Fatalf("expected completed archive evidence, got %+v", archiveCompleteBody)
	}
	archiveCompletionAudit := archiveCompleteData["audit_log"].(map[string]any)
	if archiveCompletionAudit["action"] != "admin.audit_archive.completed" || archiveCompletionAudit["target_id"] != archive["archive_id"] {
		t.Fatalf("expected archive completion audit, got %+v", archiveCompletionAudit)
	}
	archiveRecordsBody := authGetJSON(t, server.URL+"/api/admin/audit-logs/archive/records?archive_id="+archive["archive_id"].(string)+"&limit=5", adminToken("admin_1"), http.StatusOK)
	archiveRecords := archiveRecordsBody["data"].([]any)
	if len(archiveRecords) != 1 || archiveRecords[0].(map[string]any)["archive_id"] != archive["archive_id"] {
		t.Fatalf("expected archive completion records, got %+v", archiveRecordsBody)
	}
	verifyBody := authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/verify", securityAuditorToken("auditor_1"), `{"archive_id":"`+archive["archive_id"].(string)+`","now":"2027-05-24T12:02:00Z"}`, http.StatusOK)
	verifyData := verifyBody["data"].(map[string]any)
	verification := verifyData["verification"].(map[string]any)
	if verification["status"] != "verified" || verification["actual_content_hash"] != archiveObjectHash || verification["content_hash_matched"] != true || verification["manifest_hash_matched"] != true {
		t.Fatalf("expected archive object verification, got %+v", verifyBody)
	}
	verifyAudit := verifyData["audit_log"].(map[string]any)
	if verifyAudit["action"] != "admin.audit_archive.verified" || verifyAudit["actor_type"] != RoleSecurityAuditor || verifyAudit["target_id"] != archive["archive_id"] {
		t.Fatalf("expected archive verification audit, got %+v", verifyAudit)
	}
	verificationHistoryBody := authGetJSON(t, server.URL+"/api/admin/audit-logs/archive/verifications?archive_id="+archive["archive_id"].(string)+"&status=verified&limit=5", securityAuditorToken("auditor_1"), http.StatusOK)
	verificationHistory := verificationHistoryBody["data"].([]any)
	if len(verificationHistory) != 1 || verificationHistory[0].(map[string]any)["archive_id"] != archive["archive_id"] || verificationHistory[0].(map[string]any)["actual_content_hash"] != archiveObjectHash {
		t.Fatalf("expected archive verification history, got %+v", verificationHistoryBody)
	}
	authGetJSON(t, server.URL+"/api/admin/audit-logs/retention-report", userToken("user_1"), http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/audit-logs/retention-alerts/emit", securityAuditorToken("auditor_1"), `{}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/request", securityAuditorToken("auditor_1"), `{}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/complete", securityAuditorToken("auditor_1"), `{}`, http.StatusForbidden)
	authGetJSON(t, server.URL+"/api/admin/audit-logs/archive/records?limit=5", securityAuditorToken("auditor_1"), http.StatusOK)
	authPostJSON(t, server.URL+"/api/admin/audit-logs/archive/verify", userToken("user_1"), `{}`, http.StatusForbidden)
	authGetJSON(t, server.URL+"/api/admin/audit-logs/archive/verifications?limit=5", userToken("user_1"), http.StatusForbidden)
}

func auditArchiveHTTPObjectBody(t *testing.T, archive map[string]any) string {
	t.Helper()
	header := map[string]any{
		"type":                "audit_archive_manifest",
		"manifest_version":    "audit_archive_manifest:v1",
		"archive_id":          archive["archive_id"],
		"status":              archive["status"],
		"storage_prefix":      archive["storage_prefix"],
		"storage_key":         archive["storage_key"],
		"hot_days":            archive["hot_days"],
		"cold_archive_cutoff": archive["cold_archive_cutoff"],
		"requested_at":        archive["requested_at"],
		"log_count":           archive["log_count"],
		"integrity_failures":  archive["integrity_failures"],
		"manifest_algorithm":  archive["manifest_algorithm"],
		"manifest_hash":       archive["manifest_hash"],
		"idempotency_key":     archive["idempotency_key"],
	}
	lines := []string{mustJSONLineHTTP(t, header)}
	entries, _ := archive["manifest_entries"].([]any)
	for _, rawEntry := range entries {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			continue
		}
		line := map[string]any{
			"type":                "audit_log_manifest_entry",
			"archive_id":          archive["archive_id"],
			"id":                  entry["id"],
			"created_at":          entry["created_at"],
			"action":              entry["action"],
			"target_type":         entry["target_type"],
			"target_id":           entry["target_id"],
			"integrity_algorithm": entry["integrity_algorithm"],
			"integrity_hash":      entry["integrity_hash"],
			"integrity_verified":  entry["integrity_verified"],
		}
		lines = append(lines, mustJSONLineHTTP(t, line))
	}
	return strings.Join(lines, "\n") + "\n"
}

func mustJSONLineHTTP(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func auditArchiveHTTPObjectPath(storageKey string) string {
	raw := strings.TrimSpace(storageKey)
	raw = strings.TrimPrefix(raw, "worm://")
	raw = strings.TrimPrefix(raw, "s3://")
	raw = strings.TrimPrefix(raw, "minio://")
	return strings.Trim(raw, "/")
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

func TestAdminMerchantQualificationReviewHTTPUsesAtomicAuditRepositoryPath(t *testing.T) {
	store := &merchantQualificationReviewAtomicAuditStore{Store: platform.NewStore(platform.DefaultHomeModules())}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	body := authPostJSON(t, server.URL+"/api/admin/merchant-qualifications/mq_atomic/review", roleToken(RoleOpsAdmin, "ops_1"), `{"merchant_id":"merchant_1","decision":"approve","reason":"资质核验通过","reviewed_at":"2026-05-23T12:00:00Z"}`, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["qualification"].(map[string]any)["status"] != platform.QualificationStatusApproved {
		t.Fatalf("expected qualification approval response, got %+v", body)
	}
	if data["outbox_event"].(map[string]any)["topic"] != "merchant.qualification_reviewed" {
		t.Fatalf("expected qualification review outbox response, got %+v", body)
	}
	if !store.atomicAuditCalled {
		t.Fatal("expected merchant qualification review HTTP handler to call atomic audit repository path")
	}
	if store.reviewQualificationCalled {
		t.Fatal("merchant qualification review HTTP handler must not call standalone ReviewMerchantQualification before audit write")
	}
	if store.recordAuditCalled {
		t.Fatal("merchant qualification review HTTP handler must not call standalone RecordAuditLog after review")
	}
	if store.atomicReq.MerchantID != "merchant_1" || store.atomicReq.QualificationID != "mq_atomic" || store.atomicReq.Decision != "approve" {
		t.Fatalf("expected route path and review payload in atomic request, got %+v", store.atomicReq)
	}
	if store.atomicAudit.Action != "admin.merchant_qualification.reviewed" || store.atomicAudit.TargetType != "merchant_qualification" || store.atomicAudit.TargetID != "mq_atomic" {
		t.Fatalf("expected qualification review audit target from HTTP handler, got %+v", store.atomicAudit)
	}
	if store.atomicAudit.ActorType != RoleOpsAdmin || store.atomicAudit.ActorID != "ops_1" {
		t.Fatalf("expected ops principal in qualification review audit request, got %+v", store.atomicAudit)
	}
}

func TestAdminMerchantQualificationQueueHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	expiresAt := "2027-05-23T00:00:00Z"
	uploadBody := authPostJSON(t, server.URL+"/api/merchant/qualifications", merchantToken("merchant_1"), `{"type":"business_license","file_url":"https://example.test/license-admin-queue.jpg","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	uploadData := uploadBody["data"].(map[string]any)
	qualificationID := uploadData["qualifications"].([]any)[0].(map[string]any)["id"].(string)

	queueBody := authGetJSON(t, server.URL+"/api/admin/merchant-qualifications?status=pending_review&merchant_id=merchant_1&type=business_license&limit=5&now=2026-05-23T12:00:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	queueData := queueBody["data"].(map[string]any)
	if queueData["counts"].(map[string]any)["pending_review"].(float64) != 1 {
		t.Fatalf("expected one pending qualification in admin queue, got %+v", queueBody)
	}
	qualification := queueData["qualifications"].([]any)[0].(map[string]any)
	if qualification["qualification"].(map[string]any)["id"] != qualificationID || qualification["recommended_operation"].(map[string]any)["key"] != "merchant-qualification-review" {
		t.Fatalf("expected queue item to include review recommendation, got %+v", qualification)
	}
	if qualification["merchant"].(map[string]any)["id"] != "merchant_1" || qualification["incident_code"] != "merchant_qualification.pending_review" {
		t.Fatalf("expected merchant and incident context, got %+v", qualification)
	}

	detailBody := authGetJSON(t, server.URL+"/api/admin/merchant-qualifications/"+qualificationID+"?audit_limit=5&now=2026-05-23T12:00:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	detailData := detailBody["data"].(map[string]any)
	if detailData["qualification"].(map[string]any)["id"] != qualificationID || len(detailData["checklist"].([]any)) == 0 {
		t.Fatalf("expected qualification detail with checklist, got %+v", detailBody)
	}
	authGetJSON(t, server.URL+"/api/admin/merchant-qualifications?status=pending_review", roleToken(RoleSupportAdmin, "support_1"), http.StatusForbidden)
	authGetJSON(t, server.URL+"/api/admin/merchant-qualifications/"+qualificationID, roleToken(RoleSupportAdmin, "support_1"), http.StatusForbidden)

	authPostJSON(t, server.URL+"/api/admin/merchant-qualifications/"+qualificationID+"/review", roleToken(RoleOpsAdmin, "ops_1"), `{"merchant_id":"merchant_1","decision":"approve","reason":"营业执照核验通过","reviewed_at":"2026-05-23T12:05:00Z"}`, http.StatusOK)
	reviewedDetailBody := authGetJSON(t, server.URL+"/api/admin/merchant-qualifications/"+qualificationID+"?audit_limit=5", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	reviewedDetail := reviewedDetailBody["data"].(map[string]any)
	if reviewedDetail["qualification"].(map[string]any)["status"] != platform.QualificationStatusApproved || reviewedDetail["recommended_operation"].(map[string]any)["key"] != "audit-logs" {
		t.Fatalf("expected approved detail to recommend audit lookup, got %+v", reviewedDetailBody)
	}
	if len(reviewedDetail["recent_audits"].([]any)) != 1 {
		t.Fatalf("expected recent review audit in detail, got %+v", reviewedDetailBody)
	}
	outboxBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=merchant.qualification_reviewed&limit=10", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	outboxEvents := outboxBody["data"].([]any)
	if len(outboxEvents) != 1 || outboxEvents[0].(map[string]any)["aggregate_id"] != qualificationID {
		t.Fatalf("expected merchant qualification review outbox event, got %+v", outboxBody)
	}
}

func TestMerchantNotificationsHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	createBody := authPostJSON(t, server.URL+"/api/notifications", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","type":"merchant.qualification_reviewed","title":"商户资质审核结果","body":"资质审核已通过，系统已更新商户接单资格。","source_topic":"merchant.qualification_reviewed","source_event_id":"obe_mq_1","idempotency_key":"notify:merchant.qualification_reviewed:obe_mq_1","created_at":"2026-05-25T12:00:00Z"}`, http.StatusCreated)
	created := createBody["data"].(map[string]any)
	notificationID := created["id"].(string)
	if notificationID == "" || created["status"] != platform.NotificationStatusUnread || created["channel"] != platform.NotificationChannelInApp {
		t.Fatalf("expected unread notification to be created, got %+v", createBody)
	}
	duplicateBody := authPostJSON(t, server.URL+"/api/notifications", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","type":"merchant.qualification_reviewed","title":"重复投递","body":"重复投递","idempotency_key":"notify:merchant.qualification_reviewed:obe_mq_1"}`, http.StatusCreated)
	if duplicateBody["data"].(map[string]any)["id"] != notificationID {
		t.Fatalf("expected duplicate notification create to return original record, got %+v", duplicateBody)
	}
	authPostJSON(t, server.URL+"/api/notifications", roleToken(RoleSupportAdmin, "support_1"), `{"target_role":"merchant","target_id":"merchant_1","type":"merchant.qualification_reviewed","title":"无权写入","body":"无权写入","idempotency_key":"notify:denied"}`, http.StatusForbidden)

	adminListBody := authGetJSON(t, server.URL+"/api/admin/notifications?target_role=merchant&target_id=merchant_1&status=unread&source_topic=merchant.qualification_reviewed&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	adminNotifications := adminListBody["data"].([]any)
	if len(adminNotifications) != 1 || adminNotifications[0].(map[string]any)["id"] != notificationID {
		t.Fatalf("expected support admin to read merchant notification ledger, got %+v", adminListBody)
	}
	noSourceBody := authGetJSON(t, server.URL+"/api/admin/notifications?target_role=merchant&target_id=merchant_1&source_topic=audit.retention_alerts", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	if len(noSourceBody["data"].([]any)) != 0 {
		t.Fatalf("expected source topic filter to isolate notification ledger, got %+v", noSourceBody)
	}
	authGetJSON(t, server.URL+"/api/admin/notifications?target_role=merchant", roleToken(RoleOpsAdmin, "ops_1"), http.StatusBadRequest)
	authGetJSON(t, server.URL+"/api/admin/notifications?target_role=merchant&target_id=merchant_1", roleToken(RoleSecurityAuditor, "auditor_1"), http.StatusForbidden)

	deliveryBody := authPostJSON(t, server.URL+"/api/notifications/"+notificationID+"/deliveries", roleToken(RoleOpsAdmin, "ops_1"), `{"channel":"in_app","provider":"in_app","status":"delivered","provider_message_id":"ntf_1","idempotency_key":"delivery:notify:merchant.qualification_reviewed:obe_mq_1:in_app","attempted_at":"2026-05-25T12:00:05Z"}`, http.StatusCreated)
	delivery := deliveryBody["data"].(map[string]any)
	if delivery["notification_id"] != notificationID || delivery["status"] != platform.NotificationDeliveryDelivered || delivery["target_id"] != "merchant_1" {
		t.Fatalf("expected delivered notification receipt, got %+v", deliveryBody)
	}
	failedDeliveryBody := authPostJSON(t, server.URL+"/api/notifications/"+notificationID+"/deliveries", roleToken(RoleOpsAdmin, "ops_1"), `{"channel":"wechat_subscribe","provider":"wechat_subscribe","status":"failed","error_code":"invalid_openid","error_message":"openid missing","idempotency_key":"delivery:notify:merchant.qualification_reviewed:obe_mq_1:wechat","attempted_at":"2026-05-25T12:00:10Z"}`, http.StatusCreated)
	if failedDeliveryBody["data"].(map[string]any)["error_code"] != "invalid_openid" {
		t.Fatalf("expected failed notification receipt, got %+v", failedDeliveryBody)
	}
	authPostJSON(t, server.URL+"/api/notifications/"+notificationID+"/deliveries", roleToken(RoleSupportAdmin, "support_1"), `{"status":"delivered","idempotency_key":"delivery:denied"}`, http.StatusForbidden)
	deliveriesBody := authGetJSON(t, server.URL+"/api/admin/notification-deliveries?target_role=merchant&target_id=merchant_1&status=failed&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	deliveries := deliveriesBody["data"].([]any)
	if len(deliveries) != 1 || deliveries[0].(map[string]any)["error_code"] != "invalid_openid" {
		t.Fatalf("expected support admin to read failed notification deliveries, got %+v", deliveriesBody)
	}
	authGetJSON(t, server.URL+"/api/admin/notification-deliveries?target_role=merchant", roleToken(RoleOpsAdmin, "ops_1"), http.StatusBadRequest)
	authPostJSON(t, server.URL+"/api/admin/notification-deliveries/failure-alerts/emit", roleToken(RoleSupportAdmin, "support_1"), `{"target_role":"merchant","target_id":"merchant_1","limit":10}`, http.StatusForbidden)
	alertBody := authPostJSON(t, server.URL+"/api/admin/notification-deliveries/failure-alerts/emit", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","channel":"wechat_subscribe","limit":10,"now":"2026-05-25T12:01:00Z"}`, http.StatusOK)
	alertData := alertBody["data"].(map[string]any)
	emission := alertData["emission"].(map[string]any)
	if emission["status"] != "emitted" || emission["failed_count"].(float64) != 1 || emission["outbox_event_id"] == "" {
		t.Fatalf("expected notification failure alert emission, got %+v", alertBody)
	}
	alertEvent := alertData["outbox_event"].(map[string]any)
	if alertEvent["topic"] != "notification.delivery_failed_alerts" || alertEvent["event_type"] != "notification.delivery_failed_alerts.emitted" {
		t.Fatalf("expected notification failure alert outbox event, got %+v", alertBody)
	}
	alertAudit := alertData["audit_log"].(map[string]any)
	if alertAudit["action"] != "admin.notification_delivery_failure_alerts.emitted" || alertAudit["target_type"] != "notification_delivery_alerts" {
		t.Fatalf("expected notification failure alert audit log, got %+v", alertBody)
	}
	alertEventsBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=notification.delivery_failed_alerts&limit=10&now=2026-05-25T12:01:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	alertEvents := alertEventsBody["data"].([]any)
	if len(alertEvents) != 1 || alertEvents[0].(map[string]any)["id"] != alertEvent["id"] {
		t.Fatalf("expected notification failure alert outbox event to be queryable, got %+v", alertEventsBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-deliveries/retries/schedule", roleToken(RoleSupportAdmin, "support_1"), `{"target_role":"merchant","target_id":"merchant_1","limit":10}`, http.StatusForbidden)
	retryBody := authPostJSON(t, server.URL+"/api/admin/notification-deliveries/retries/schedule", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","channel":"wechat_subscribe","provider":"wechat_subscribe","limit":10,"retry_after_seconds":300,"now":"2026-05-25T12:02:00Z"}`, http.StatusOK)
	retryData := retryBody["data"].(map[string]any)
	schedule := retryData["schedule"].(map[string]any)
	if schedule["status"] != "scheduled" || schedule["scheduled_count"].(float64) != 1 || schedule["retry_after_seconds"].(float64) != 300 {
		t.Fatalf("expected notification delivery retry schedule, got %+v", retryBody)
	}
	retryEvent := retryData["outbox_event"].(map[string]any)
	if retryEvent["topic"] != "notification.delivery_retries" || retryEvent["event_type"] != "notification.delivery_retries.scheduled" {
		t.Fatalf("expected notification delivery retry outbox event, got %+v", retryBody)
	}
	retryAudit := retryData["audit_log"].(map[string]any)
	if retryAudit["action"] != "admin.notification_delivery_retries.scheduled" || retryAudit["target_type"] != "notification_delivery_retries" {
		t.Fatalf("expected notification delivery retry audit log, got %+v", retryBody)
	}
	blockedRetryEventsBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=notification.delivery_retries&limit=10&now=2026-05-25T12:02:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	if len(blockedRetryEventsBody["data"].([]any)) != 0 {
		t.Fatalf("expected notification retry event hidden before retry_at, got %+v", blockedRetryEventsBody)
	}
	retryEventsBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=notification.delivery_retries&limit=10&now=2026-05-25T12:07:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	retryEvents := retryEventsBody["data"].([]any)
	if len(retryEvents) != 1 || retryEvents[0].(map[string]any)["id"] != retryEvent["id"] {
		t.Fatalf("expected notification delivery retry outbox event after backoff, got %+v", retryEventsBody)
	}
	quietDeliveryBody := authPostJSON(t, server.URL+"/api/notifications/"+notificationID+"/deliveries", roleToken(RoleOpsAdmin, "ops_1"), `{"channel":"push","provider":"push","status":"queued","error_code":"notification_quiet_window","error_message":"notification quiet window suppressed push","idempotency_key":"delivery:notify:merchant.qualification_reviewed:obe_mq_1:push:quiet","attempted_at":"2026-05-25T12:00:15Z","retry_at":"2026-05-25T12:12:00Z"}`, http.StatusCreated)
	quietDelivery := quietDeliveryBody["data"].(map[string]any)
	if quietDelivery["status"] != platform.NotificationDeliveryQueued || quietDelivery["error_code"] != "notification_quiet_window" || quietDelivery["retry_at"] != "2026-05-25T12:12:00Z" {
		t.Fatalf("expected queued quiet-window notification receipt, got %+v", quietDeliveryBody)
	}
	quietDeliveriesBody := authGetJSON(t, server.URL+"/api/admin/notification-deliveries?target_role=merchant&target_id=merchant_1&status=queued&error_code=notification_quiet_window&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	quietDeliveries := quietDeliveriesBody["data"].([]any)
	if len(quietDeliveries) != 1 || quietDeliveries[0].(map[string]any)["id"] != quietDelivery["id"] {
		t.Fatalf("expected queued quiet-window delivery filter, got %+v", quietDeliveriesBody)
	}
	quietRetryBody := authPostJSON(t, server.URL+"/api/admin/notification-deliveries/retries/schedule", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","channel":"push","provider":"push","status":"queued","error_code":"notification_quiet_window","limit":10,"retry_at":"2026-05-25T12:12:00Z","now":"2026-05-25T12:03:00Z"}`, http.StatusOK)
	quietRetryData := quietRetryBody["data"].(map[string]any)
	quietSchedule := quietRetryData["schedule"].(map[string]any)
	if quietSchedule["status"] != "scheduled" || quietSchedule["delivery_status"] != platform.NotificationDeliveryQueued || quietSchedule["error_code"] != "notification_quiet_window" || quietSchedule["retry_at"] != "2026-05-25T12:12:00Z" {
		t.Fatalf("expected quiet-window queued retry schedule, got %+v", quietRetryBody)
	}
	quietRetryEvent := quietRetryData["outbox_event"].(map[string]any)
	if quietRetryEvent["topic"] != "notification.delivery_retries" || quietRetryEvent["aggregate_id"] != platform.NotificationDeliveryQueued {
		t.Fatalf("expected quiet-window queued retry outbox event, got %+v", quietRetryBody)
	}
	quietRetryAudit := quietRetryData["audit_log"].(map[string]any)
	if quietRetryAudit["target_id"] != platform.NotificationDeliveryQueued {
		t.Fatalf("expected quiet-window retry audit target, got %+v", quietRetryBody)
	}
	quietRetryEventsBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=notification.delivery_retries&limit=10&now=2026-05-25T12:12:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	quietRetryEvents := quietRetryEventsBody["data"].([]any)
	quietRetryEventFound := false
	for _, item := range quietRetryEvents {
		if item.(map[string]any)["id"] == quietRetryEvent["id"] {
			quietRetryEventFound = true
			break
		}
	}
	if !quietRetryEventFound {
		t.Fatalf("expected quiet-window retry outbox event after retry_at, got %+v", quietRetryEventsBody)
	}
	autoQuietDeliveryBody := authPostJSON(t, server.URL+"/api/notifications/"+notificationID+"/deliveries", roleToken(RoleOpsAdmin, "ops_1"), `{"channel":"sms","provider":"sms","status":"queued","error_code":"notification_quiet_window","error_message":"notification quiet window suppressed sms","idempotency_key":"delivery:notify:merchant.qualification_reviewed:obe_mq_1:sms:quiet","attempted_at":"2026-05-25T12:00:20Z","retry_at":"2026-05-25T12:10:00Z"}`, http.StatusCreated)
	autoQuietDelivery := autoQuietDeliveryBody["data"].(map[string]any)
	if autoQuietDelivery["status"] != platform.NotificationDeliveryQueued || autoQuietDelivery["retry_at"] != "2026-05-25T12:10:00Z" {
		t.Fatalf("expected auto quiet-window queued receipt, got %+v", autoQuietDeliveryBody)
	}
	notDueQuietDeliveriesBody := authGetJSON(t, server.URL+"/api/admin/notification-deliveries?channel=sms&provider=sms&status=queued&error_code=notification_quiet_window&retry_at_before=2026-05-25T12:09:59Z&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(notDueQuietDeliveriesBody["data"].([]any)) != 0 {
		t.Fatalf("expected retry_at_before to hide future quiet-window delivery, got %+v", notDueQuietDeliveriesBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-deliveries/quiet-window-retries/schedule", roleToken(RoleSupportAdmin, "support_1"), `{"channel":"sms","provider":"sms","limit":10,"now":"2026-05-25T12:10:00Z"}`, http.StatusForbidden)
	autoQuietRetryBody := authPostJSON(t, server.URL+"/api/admin/notification-deliveries/quiet-window-retries/schedule", roleToken(RoleOpsAdmin, "ops_1"), `{"channel":"sms","provider":"sms","limit":10,"now":"2026-05-25T12:10:00Z"}`, http.StatusOK)
	autoQuietRetryData := autoQuietRetryBody["data"].(map[string]any)
	autoQuietSchedule := autoQuietRetryData["schedule"].(map[string]any)
	if autoQuietSchedule["status"] != "scheduled" || autoQuietSchedule["delivery_status"] != platform.NotificationDeliveryQueued || autoQuietSchedule["scheduled_count"].(float64) != 1 || autoQuietSchedule["retry_at"] != "2026-05-25T12:10:00Z" {
		t.Fatalf("expected auto quiet-window retry schedule, got %+v", autoQuietRetryBody)
	}
	autoQuietRetryEvent := autoQuietRetryData["outbox_event"].(map[string]any)
	if autoQuietRetryEvent["topic"] != "notification.delivery_retries" || autoQuietRetryEvent["aggregate_id"] != platform.NotificationDeliveryQueued {
		t.Fatalf("expected auto quiet-window retry outbox event, got %+v", autoQuietRetryBody)
	}
	authPutJSON(t, server.URL+"/api/admin/notification-preferences", roleToken(RoleSupportAdmin, "support_1"), `{"target_role":"merchant","target_id":"merchant_1","notification_type":"merchant.qualification_reviewed","disabled_channels":["sms"]}`, http.StatusForbidden)
	adminPreferenceBody := authPutJSON(t, server.URL+"/api/admin/notification-preferences", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","notification_type":"merchant.qualification_reviewed","enabled_channels":["wechat_subscribe","push"],"disabled_channels":["sms"],"quiet_hours":{"enabled":true,"start":"22:00","end":"08:00","timezone_offset":"+08:00","channels":["wechat_subscribe","push"]},"updated_at":"2026-05-25T12:03:00Z"}`, http.StatusOK)
	adminPreferenceData := adminPreferenceBody["data"].(map[string]any)
	adminPreference := adminPreferenceData["preference"].(map[string]any)
	if adminPreference["preference_key"] != "merchant:merchant_1:merchant.qualification_reviewed" || adminPreference["notification_type"] != "merchant.qualification_reviewed" {
		t.Fatalf("expected admin notification preference save, got %+v", adminPreferenceBody)
	}
	adminPreferenceAudit := adminPreferenceData["audit_log"].(map[string]any)
	if adminPreferenceAudit["action"] != "admin.notification_preferences.saved" || adminPreferenceAudit["target_id"] != adminPreference["id"] {
		t.Fatalf("expected notification preference audit log, got %+v", adminPreferenceBody)
	}
	adminPreferencesBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences?target_role=merchant&target_id=merchant_1&notification_type=merchant.qualification_reviewed&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	adminPreferences := adminPreferencesBody["data"].([]any)
	if len(adminPreferences) != 1 || adminPreferences[0].(map[string]any)["preference_key"] != "merchant:merchant_1:merchant.qualification_reviewed" {
		t.Fatalf("expected support admin to read notification preferences, got %+v", adminPreferencesBody)
	}
	adminPreferenceByKeyBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences?preference_key=merchant:merchant_1:merchant.qualification_reviewed&limit=1", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(adminPreferenceByKeyBody["data"].([]any)) != 1 {
		t.Fatalf("expected notification preference key lookup, got %+v", adminPreferenceByKeyBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/batch", roleToken(RoleSupportAdmin, "support_1"), `{"reason":"support cannot batch save","preferences":[{"target_role":"merchant","target_id":"merchant_1","notification_type":"order.status_changed"}]}`, http.StatusForbidden)
	adminPreferenceBatchBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/batch", roleToken(RoleOpsAdmin, "ops_1"), `{"reason":"批量更新关键通知触达策略","updated_at":"2026-05-25T12:03:30Z","preferences":[{"target_role":"merchant","target_id":"merchant_1","notification_type":"order.status_changed","enabled_channels":["wechat_subscribe","push"],"disabled_channels":["sms"]},{"target_role":"user","target_id":"user_1","notification_type":"after_sales.updated","disabled_channels":["sms"]}]}`, http.StatusOK)
	adminPreferenceBatchData := adminPreferenceBatchBody["data"].(map[string]any)
	adminPreferenceBatch := adminPreferenceBatchData["batch"].(map[string]any)
	if adminPreferenceBatch["saved"].(float64) != 2 || len(adminPreferenceBatch["preferences"].([]any)) != 2 || adminPreferenceBatch["batch_id"] == "" {
		t.Fatalf("expected notification preference batch save, got %+v", adminPreferenceBatchBody)
	}
	adminPreferenceBatchAudit := adminPreferenceBatchData["audit_log"].(map[string]any)
	if adminPreferenceBatchAudit["action"] != "admin.notification_preferences.batch_saved" || adminPreferenceBatchAudit["target_id"] != adminPreferenceBatch["batch_id"] {
		t.Fatalf("expected notification preference batch audit, got %+v", adminPreferenceBatchBody)
	}
	adminBatchPreferencesBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences?target_role=user&target_id=user_1&notification_type=after_sales.updated&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	adminBatchPreferences := adminBatchPreferencesBody["data"].([]any)
	if len(adminBatchPreferences) != 1 || adminBatchPreferences[0].(map[string]any)["preference_key"] != "user:user_1:after_sales.updated" {
		t.Fatalf("expected batch-saved user preference to be queryable, got %+v", adminBatchPreferencesBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests", roleToken(RoleSupportAdmin, "support_1"), `{"reason":"support cannot request","preferences":[{"target_role":"merchant","target_id":"merchant_1","notification_type":"order.status_changed"}]}`, http.StatusForbidden)
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests", roleToken(RoleOpsAdmin, "ops_1"), `{"reason":"无效灰度范围","rollout":{"mode":"target_ids"},"preferences":[{"target_role":"merchant","target_id":"merchant_1","notification_type":"order.status_changed"}]}`, http.StatusBadRequest)
	preferenceChangeRequestBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests", roleToken(RoleOpsAdmin, "ops_1"), `{"reason":"申请批量调整关键通知触达策略","updated_at":"2026-05-25T12:03:40Z","rollout":{"mode":"target_ids","target_ids":["merchant_1"],"max_targets":10},"preferences":[{"target_role":"merchant","target_id":"merchant_1","notification_type":"merchant.qualification_reviewed","disabled_channels":["enterprise_wechat"]},{"target_role":"user","target_id":"user_1","notification_type":"after_sales.updated","disabled_channels":["push"]}]}`, http.StatusCreated)
	preferenceChangeRequest := preferenceChangeRequestBody["data"].(map[string]any)
	changeRequestID := preferenceChangeRequest["id"].(string)
	if preferenceChangeRequest["status"] != "pending_approval" || len(preferenceChangeRequest["preference_keys"].([]any)) != 2 {
		t.Fatalf("expected pending notification preference change request, got %+v", preferenceChangeRequestBody)
	}
	preferenceChangeRollout := preferenceChangeRequest["rollout"].(map[string]any)
	if preferenceChangeRollout["mode"] != "target_ids" || len(preferenceChangeRollout["target_ids"].([]any)) != 1 {
		t.Fatalf("expected target rollout on notification preference change request, got %+v", preferenceChangeRequestBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+changeRequestID+"/review", roleToken(RoleOpsAdmin, "ops_1"), `{"decision":"approve","reason":"self review must fail"}`, http.StatusConflict)
	reviewedPreferenceChangeBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+changeRequestID+"/review", roleToken(RoleOpsAdmin, "ops_2"), `{"decision":"approve","reason":"策略范围和静默窗口已复核"}`, http.StatusOK)
	reviewedPreferenceChange := reviewedPreferenceChangeBody["data"].(map[string]any)["change_request"].(map[string]any)
	if reviewedPreferenceChange["status"] != "approved" || reviewedPreferenceChange["review_decision"] != "approve" {
		t.Fatalf("expected approved notification preference change request, got %+v", reviewedPreferenceChangeBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+changeRequestID+"/apply", roleToken(RoleOpsAdmin, "ops_1"), `{"reason":"requester cannot apply","updated_at":"2026-05-25T12:03:45Z"}`, http.StatusConflict)
	appliedPreferenceChangeBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+changeRequestID+"/apply", roleToken(RoleOpsAdmin, "ops_2"), `{"reason":"按已审批策略应用到生产偏好账本","updated_at":"2026-05-25T12:03:45Z"}`, http.StatusOK)
	appliedPreferenceChangeData := appliedPreferenceChangeBody["data"].(map[string]any)
	appliedPreferenceChange := appliedPreferenceChangeData["change_request"].(map[string]any)
	appliedPreferenceBatch := appliedPreferenceChangeData["batch"].(map[string]any)
	appliedPreferenceAudit := appliedPreferenceChangeData["audit_log"].(map[string]any)
	if appliedPreferenceChange["status"] != "applied" || appliedPreferenceBatch["saved"].(float64) != 1 || appliedPreferenceAudit["action"] != "admin.notification_preferences.change_applied" {
		t.Fatalf("expected applied notification preference change request, got %+v", appliedPreferenceChangeBody)
	}
	if appliedPreferenceChange["skipped_count"].(float64) != 1 || len(appliedPreferenceChange["skipped_preference_keys"].([]any)) != 1 {
		t.Fatalf("expected rollout to skip one notification preference, got %+v", appliedPreferenceChangeBody)
	}
	appliedPreferencePayload := appliedPreferenceAudit["payload"].(map[string]any)
	if appliedPreferencePayload["rollout_mode"] != "target_ids" || appliedPreferencePayload["applied_count"].(float64) != 1 || appliedPreferencePayload["skipped_count"].(float64) != 1 {
		t.Fatalf("expected rollout scope in notification preference apply audit, got %+v", appliedPreferenceAudit)
	}
	skippedPreferenceBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences?target_role=user&target_id=user_1&notification_type=after_sales.updated&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	skippedPreferences := skippedPreferenceBody["data"].([]any)
	if len(skippedPreferences) != 1 || len(skippedPreferences[0].(map[string]any)["disabled_channels"].([]any)) != 1 || skippedPreferences[0].(map[string]any)["disabled_channels"].([]any)[0] != "sms" {
		t.Fatalf("expected skipped rollout preference to keep previous channels, got %+v", skippedPreferenceBody)
	}
	appliedPreferenceChangesBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences/change-requests?status=applied&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	appliedPreferenceChanges := appliedPreferenceChangesBody["data"].(map[string]any)["items"].([]any)
	if len(appliedPreferenceChanges) == 0 || appliedPreferenceChanges[0].(map[string]any)["id"] != changeRequestID {
		t.Fatalf("expected applied notification preference change request to be queryable, got %+v", appliedPreferenceChangesBody)
	}
	if appliedPreferenceChanges[0].(map[string]any)["skipped_count"].(float64) != 1 {
		t.Fatalf("expected applied notification preference ledger to keep rollout skip count, got %+v", appliedPreferenceChangesBody)
	}
	rejectedPreferenceChangeRequestBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests", roleToken(RoleOpsAdmin, "ops_1"), `{"reason":"申请关闭非关键优惠短信触达","updated_at":"2026-05-25T12:03:50Z","preferences":[{"target_role":"user","target_id":"user_1","notification_type":"coupon.campaign","disabled_channels":["sms"]}]}`, http.StatusCreated)
	rejectedPreferenceChangeRequest := rejectedPreferenceChangeRequestBody["data"].(map[string]any)
	rejectedChangeRequestID := rejectedPreferenceChangeRequest["id"].(string)
	rejectedPreferenceChangeBody := authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+rejectedChangeRequestID+"/review", roleToken(RoleOpsAdmin, "ops_2"), `{"decision":"reject","reason":"优惠短信关闭范围需要灰度验证后再提交"}`, http.StatusOK)
	rejectedPreferenceChange := rejectedPreferenceChangeBody["data"].(map[string]any)["change_request"].(map[string]any)
	if rejectedPreferenceChange["status"] != "rejected" || rejectedPreferenceChange["review_decision"] != "reject" {
		t.Fatalf("expected rejected notification preference change request, got %+v", rejectedPreferenceChangeBody)
	}
	authPostJSON(t, server.URL+"/api/admin/notification-preferences/change-requests/"+rejectedChangeRequestID+"/apply", roleToken(RoleOpsAdmin, "ops_2"), `{"reason":"rejected request must not apply","updated_at":"2026-05-25T12:03:55Z"}`, http.StatusConflict)
	rejectedPreferenceChangesBody := authGetJSON(t, server.URL+"/api/admin/notification-preferences/change-requests?status=rejected&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	rejectedPreferenceChanges := rejectedPreferenceChangesBody["data"].(map[string]any)["items"].([]any)
	if len(rejectedPreferenceChanges) == 0 || rejectedPreferenceChanges[0].(map[string]any)["id"] != rejectedChangeRequestID {
		t.Fatalf("expected rejected notification preference change request to be queryable, got %+v", rejectedPreferenceChangesBody)
	}
	authGetJSON(t, server.URL+"/api/admin/notification-preferences?target_id=merchant_1", roleToken(RoleOpsAdmin, "ops_1"), http.StatusBadRequest)
	merchantPreferenceBody := authPutJSON(t, server.URL+"/api/merchant/notification-preferences", merchantToken("merchant_1"), `{"notification_type":"order.status_changed","disabled_channels":["push"],"quiet_hours":{"enabled":true,"start":"21:30","end":"07:30","timezone_offset":"+08:00","channels":["push"]},"updated_at":"2026-05-25T12:04:00Z"}`, http.StatusOK)
	merchantPreference := merchantPreferenceBody["data"].(map[string]any)
	if merchantPreference["target_role"] != "merchant" || merchantPreference["target_id"] != "merchant_1" || merchantPreference["preference_key"] != "merchant:merchant_1:order.status_changed" {
		t.Fatalf("expected merchant notification preference save to be scoped, got %+v", merchantPreferenceBody)
	}
	merchantPreferencesBody := authGetJSON(t, server.URL+"/api/merchant/notification-preferences?notification_type=order.status_changed&limit=10", merchantToken("merchant_1"), http.StatusOK)
	merchantPreferences := merchantPreferencesBody["data"].([]any)
	if len(merchantPreferences) != 1 || merchantPreferences[0].(map[string]any)["preference_key"] != "merchant:merchant_1:order.status_changed" {
		t.Fatalf("expected merchant notification preferences to be queryable, got %+v", merchantPreferencesBody)
	}
	otherMerchantPreferencesBody := authGetJSON(t, server.URL+"/api/merchant/notification-preferences?notification_type=order.status_changed&limit=10", merchantToken("merchant_2"), http.StatusOK)
	if len(otherMerchantPreferencesBody["data"].([]any)) != 0 {
		t.Fatalf("expected merchant preference isolation, got %+v", otherMerchantPreferencesBody)
	}
	authPutJSON(t, server.URL+"/api/user/notification-preferences", merchantToken("merchant_1"), `{"notification_type":"after_sales.updated","disabled_channels":["sms"]}`, http.StatusForbidden)
	userPreferenceBody := authPutJSON(t, server.URL+"/api/user/notification-preferences", userToken("user_1"), `{"target_role":"merchant","target_id":"merchant_1","notification_type":"after_sales.updated","enabled_channels":["wechat_subscribe","push"],"disabled_channels":["sms"],"quiet_hours":{"enabled":true,"start":"22:30","end":"07:30","timezone_offset":"+08:00","channels":["wechat_subscribe","push"]},"updated_at":"2026-05-25T12:05:00Z"}`, http.StatusOK)
	userPreference := userPreferenceBody["data"].(map[string]any)
	if userPreference["target_role"] != "user" || userPreference["target_id"] != "user_1" || userPreference["preference_key"] != "user:user_1:after_sales.updated" {
		t.Fatalf("expected user notification preference save to be scoped, got %+v", userPreferenceBody)
	}
	userPreferencesBody := authGetJSON(t, server.URL+"/api/user/notification-preferences?notification_type=after_sales.updated&limit=10", userToken("user_1"), http.StatusOK)
	userPreferences := userPreferencesBody["data"].([]any)
	if len(userPreferences) != 1 || userPreferences[0].(map[string]any)["preference_key"] != "user:user_1:after_sales.updated" {
		t.Fatalf("expected user notification preferences to be queryable, got %+v", userPreferencesBody)
	}
	otherUserPreferencesBody := authGetJSON(t, server.URL+"/api/user/notification-preferences?notification_type=after_sales.updated&limit=10", userToken("user_2"), http.StatusOK)
	if len(otherUserPreferencesBody["data"].([]any)) != 0 {
		t.Fatalf("expected user preference isolation, got %+v", otherUserPreferencesBody)
	}
	preferenceChangeEventsBody := authGetJSON(t, server.URL+"/api/admin/outbox/events?topic=notification.preferences_changed&limit=10&now=2026-05-25T12:06:00Z", roleToken(RoleOpsAdmin, "ops_1"), http.StatusOK)
	preferenceChangeEvents := preferenceChangeEventsBody["data"].([]any)
	preferenceChangeKeys := map[string]bool{}
	for _, item := range preferenceChangeEvents {
		event := item.(map[string]any)
		payload := event["payload"].(map[string]any)
		preferenceChangeKeys[payload["preference_key"].(string)] = true
		if event["topic"] != "notification.preferences_changed" || event["aggregate_type"] != "notification_preference" || event["event_type"] != "notification.preferences.changed" {
			t.Fatalf("expected notification preference change outbox event, got %+v", event)
		}
	}
	for _, key := range []string{
		"merchant:merchant_1:merchant.qualification_reviewed",
		"merchant:merchant_1:order.status_changed",
		"user:user_1:after_sales.updated",
	} {
		if !preferenceChangeKeys[key] {
			t.Fatalf("expected preference change event for %s, got %+v", key, preferenceChangeEventsBody)
		}
	}

	listBody := authGetJSON(t, server.URL+"/api/merchant/notifications?status=unread&limit=10", merchantToken("merchant_1"), http.StatusOK)
	notifications := listBody["data"].([]any)
	if len(notifications) != 1 || notifications[0].(map[string]any)["id"] != notificationID || notifications[0].(map[string]any)["source_event_id"] != "obe_mq_1" {
		t.Fatalf("expected merchant unread notification list, got %+v", listBody)
	}
	otherMerchantBody := authGetJSON(t, server.URL+"/api/merchant/notifications?status=unread", merchantToken("merchant_2"), http.StatusOK)
	if len(otherMerchantBody["data"].([]any)) != 0 {
		t.Fatalf("expected merchant isolation for notifications, got %+v", otherMerchantBody)
	}
	readBody := authPostJSON(t, server.URL+"/api/merchant/notifications/"+notificationID+"/read", merchantToken("merchant_1"), `{"read_at":"2026-05-25T12:01:00Z"}`, http.StatusOK)
	if readBody["data"].(map[string]any)["status"] != platform.NotificationStatusRead {
		t.Fatalf("expected notification to be marked read, got %+v", readBody)
	}
	emptyBody := authGetJSON(t, server.URL+"/api/merchant/notifications?status=unread&limit=10", merchantToken("merchant_1"), http.StatusOK)
	if len(emptyBody["data"].([]any)) != 0 {
		t.Fatalf("expected no unread notifications after read, got %+v", emptyBody)
	}
}

func TestNotificationProviderCallbackHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store, WithNotificationProviderCallbackSecret("callback-secret")))
	defer server.Close()

	createBody := authPostJSON(t, server.URL+"/api/notifications", roleToken(RoleOpsAdmin, "ops_1"), `{"target_role":"merchant","target_id":"merchant_1","type":"merchant.qualification_reviewed","title":"商户资质审核结果","body":"资质审核已通过，系统已更新商户接单资格。","source_topic":"merchant.qualification_reviewed","source_event_id":"obe_mq_callback_1","idempotency_key":"notify:merchant.qualification_reviewed:obe_mq_callback_1","created_at":"2026-05-25T12:08:00Z"}`, http.StatusCreated)
	notificationID := createBody["data"].(map[string]any)["id"].(string)

	authPostJSON(t, server.URL+"/api/notifications/provider-callback", "", `{"notification_id":"`+notificationID+`","channel":"wechat_subscribe","provider":"wechat_subscribe","status":"delivered","provider_message_id":"wx_msg_1","callback_at":"2026-05-25T12:09:00Z"}`, http.StatusUnauthorized)
	deliveredPayload := notificationProviderCallbackPayload{
		NotificationID:    notificationID,
		Channel:           "wechat_subscribe",
		Provider:          "wechat_subscribe",
		Status:            platform.NotificationDeliveryDelivered,
		ProviderMessageID: "wx_msg_1",
		AttemptedAt:       time.Date(2026, 5, 25, 12, 8, 59, 0, time.UTC),
		DeliveredAt:       time.Date(2026, 5, 25, 12, 9, 0, 0, time.UTC),
		CallbackAt:        time.Date(2026, 5, 25, 12, 9, 1, 0, time.UTC),
	}
	deliveredBody := signedNotificationProviderCallbackJSON(t, server.URL+"/api/notifications/provider-callback", deliveredPayload, "callback-secret", http.StatusCreated)
	deliveredData := deliveredBody["data"].(map[string]any)
	delivered := deliveredData["delivery"].(map[string]any)
	if deliveredData["signature_verified"] != true || delivered["notification_id"] != notificationID || delivered["status"] != platform.NotificationDeliveryDelivered || delivered["provider_message_id"] != "wx_msg_1" {
		t.Fatalf("expected signed provider callback to record delivered receipt, got %+v", deliveredBody)
	}
	duplicateBody := signedNotificationProviderCallbackJSON(t, server.URL+"/api/notifications/provider-callback", deliveredPayload, "callback-secret", http.StatusCreated)
	if duplicateBody["data"].(map[string]any)["delivery"].(map[string]any)["id"] != delivered["id"] {
		t.Fatalf("expected duplicate provider callback to return original delivery, got %+v", duplicateBody)
	}

	failedPayload := notificationProviderCallbackPayload{
		NotificationID:    notificationID,
		Channel:           "sms",
		Provider:          "aliyun_sms",
		Status:            platform.NotificationDeliveryFailed,
		ProviderMessageID: "sms_msg_1",
		ErrorCode:         "recipient_unreachable",
		ErrorMessage:      "phone unreachable",
		CallbackAt:        time.Date(2026, 5, 25, 12, 10, 0, 0, time.UTC),
	}
	failedBody := signedNotificationProviderCallbackJSON(t, server.URL+"/api/notifications/provider-callback", failedPayload, "callback-secret", http.StatusCreated)
	failed := failedBody["data"].(map[string]any)["delivery"].(map[string]any)
	if failed["status"] != platform.NotificationDeliveryFailed || failed["error_code"] != "recipient_unreachable" || failed["provider"] != "aliyun_sms" {
		t.Fatalf("expected signed provider callback to record failed receipt, got %+v", failedBody)
	}
	deliveriesBody := authGetJSON(t, server.URL+"/api/admin/notification-deliveries?notification_id="+notificationID+"&status=failed&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	deliveries := deliveriesBody["data"].([]any)
	if len(deliveries) != 1 || deliveries[0].(map[string]any)["provider_message_id"] != "sms_msg_1" {
		t.Fatalf("expected provider callback failure receipt to be queryable, got %+v", deliveriesBody)
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
	if _, err := store.CreateServiceTicket(platform.CreateServiceTicketRequest{
		UserID:             "user_1",
		Category:           "配送问题",
		Title:              "配送问题 · 预计送达未更新",
		Content:            "骑手到店很久了",
		RelatedOrderID:     order.ID,
		RelatedOrderTitle:  "蓝海餐厅 · 招牌牛肉饭等 1 件",
		RelatedOrderStatus: platform.StatusMerchantPending,
	}); err != nil {
		t.Fatal(err)
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
	if request["shop_name"] != "蓝海餐厅" || request["order_item_summary"] != "招牌牛肉饭 x 1" || request["latest_event_message"] != "用户已提交售后申请" {
		t.Fatalf("expected after-sales response to expose order context, got %+v", createBody)
	}

	userListBody := authGetJSON(t, server.URL+"/api/after-sales", userToken("user_1"), http.StatusOK)
	filteredUserListBody := authGetJSON(t, server.URL+"/api/after-sales?order_id="+order.ID, userToken("user_1"), http.StatusOK)
	merchantListBody := authGetJSON(t, server.URL+"/api/merchant/after-sales", merchantToken("merchant_1"), http.StatusOK)
	adminListBody := authGetJSON(t, server.URL+"/api/admin/after-sales", adminToken("admin_1"), http.StatusOK)
	filteredAdminListBody := authGetJSON(t, server.URL+"/api/admin/after-sales?order_id="+order.ID+"&request_id="+requestID+"&status="+platform.AfterSalesPendingMerchant, adminToken("admin_1"), http.StatusOK)
	if len(userListBody["data"].([]any)) != 1 || len(filteredUserListBody["data"].([]any)) != 1 || len(merchantListBody["data"].([]any)) != 1 || len(adminListBody["data"].([]any)) != 1 || len(filteredAdminListBody["data"].([]any)) != 1 {
		t.Fatalf("expected after-sales lists to expose request, user=%+v filtered=%+v merchant=%+v admin=%+v filteredAdmin=%+v", userListBody, filteredUserListBody, merchantListBody, adminListBody, filteredAdminListBody)
	}
	filteredRequest := filteredUserListBody["data"].([]any)[0].(map[string]any)
	if filteredRequest["id"] != requestID || filteredRequest["shop_name"] != "蓝海餐厅" || filteredRequest["order_status"] != platform.StatusMerchantPending {
		t.Fatalf("expected filtered after-sales list item to keep order context, got %+v", filteredUserListBody)
	}
	filteredAdminRequest := filteredAdminListBody["data"].([]any)[0].(map[string]any)
	if filteredAdminRequest["id"] != requestID || filteredAdminRequest["status"] != platform.AfterSalesPendingMerchant {
		t.Fatalf("expected filtered admin after-sales list item to match request filters, got %+v", filteredAdminListBody)
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
	filteredUserListBody = authGetJSON(t, server.URL+"/api/after-sales?order_id="+order.ID, userToken("user_1"), http.StatusOK)
	filteredRequest = filteredUserListBody["data"].([]any)[0].(map[string]any)
	if filteredRequest["latest_event_message"] != "已核实后厨打包记录" {
		t.Fatalf("expected filtered after-sales list item to expose latest user event, got %+v", filteredUserListBody)
	}
	authPostJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", adminToken("admin_1"), `{"action":"internal_note","message":"客服内部备注","visible_to_user":false}`, http.StatusCreated)
	userEventsBody := authGetJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", userToken("user_1"), http.StatusOK)
	adminEventsBody := authGetJSON(t, server.URL+"/api/after-sales/"+requestID+"/events", adminToken("admin_1"), http.StatusOK)
	if len(userEventsBody["data"].([]any)) != 3 || len(adminEventsBody["data"].([]any)) != 4 {
		t.Fatalf("expected user-visible timeline and admin full audit log, user=%+v admin=%+v", userEventsBody, adminEventsBody)
	}
	adminDetailBody := authGetJSON(t, server.URL+"/api/admin/after-sales/"+requestID, adminToken("admin_1"), http.StatusOK)
	adminDetail := adminDetailBody["data"].(map[string]any)
	if adminDetail["request"].(map[string]any)["id"] != requestID ||
		adminDetail["request"].(map[string]any)["shop_name"] != "蓝海餐厅" ||
		adminDetail["event_summary"].(map[string]any)["total"] != float64(4) ||
		adminDetail["event_summary"].(map[string]any)["internal_only"] != float64(1) ||
		adminDetail["evidence_summary"].(map[string]any)["total"] != float64(1) ||
		adminDetail["dispatch_summary"].(map[string]any)["total"] != float64(0) ||
		adminDetail["service_ticket_summary"].(map[string]any)["total"] != float64(1) ||
		adminDetail["refund_summary"].(map[string]any)["total"] != float64(0) ||
		adminDetail["audit_summary"].(map[string]any)["total"] != float64(0) {
		t.Fatalf("expected admin after-sales detail aggregate, got %+v", adminDetailBody)
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
	if _, err := store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  RoleAdmin,
		ActorID:    "admin_1",
		Action:     "admin.order.refunded",
		TargetType: "order",
		TargetID:   order.ID,
		RequestID:  "req_http_after_sales_refund_audit",
		Payload:    map[string]any{"amount_fen": order.AmountFen},
		CreatedAt:  time.Date(2026, 6, 3, 10, 30, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	adminDetailAfterReview := authGetJSON(t, server.URL+"/api/admin/after-sales/"+requestID, adminToken("admin_1"), http.StatusOK)
	adminDetailAfterReviewData := adminDetailAfterReview["data"].(map[string]any)
	refundSummary := adminDetailAfterReviewData["refund_summary"].(map[string]any)
	if refundSummary["total"] != float64(1) || refundSummary["success_count"] != float64(1) || refundSummary["total_amount_fen"] != float64(order.AmountFen) {
		t.Fatalf("expected admin after-sales detail to expose refund summary after review, got %+v", adminDetailAfterReview)
	}
	auditSummary := adminDetailAfterReviewData["audit_summary"].(map[string]any)
	if auditSummary["total"] != float64(2) || auditSummary["order_count"] != float64(1) || auditSummary["after_sales_count"] != float64(1) || auditSummary["verified_count"] != float64(2) {
		t.Fatalf("expected admin after-sales detail to expose aggregated audits, got %+v", adminDetailAfterReview)
	}
	authGetJSON(t, server.URL+"/api/admin/refunds?order_id="+order.ID+"&limit=10", userToken("user_1"), http.StatusForbidden)
	refundsBody := authGetJSON(t, server.URL+"/api/admin/refunds?order_id="+order.ID+"&status=success&limit=10", adminToken("admin_1"), http.StatusOK)
	refunds := refundsBody["data"].([]any)
	if len(refunds) != 1 || refunds[0].(map[string]any)["order_id"] != order.ID || refunds[0].(map[string]any)["destination"] != platform.RefundDestinationBalance {
		t.Fatalf("expected admin refund list for order, got %+v", refundsBody)
	}
}

func TestAdminOrderDetailHTTPFlow(t *testing.T) {
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
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: order.AmountFen, IdempotencyKey: "credit_http_admin_order_detail"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_http_admin_order_detail"}); err != nil || paidOrder.Status != platform.StatusMerchantPending {
		t.Fatalf("expected paid merchant order, order=%+v err=%v", paidOrder, err)
	}
	request, err := store.CreateAfterSales(platform.CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 800,
	})
	if err != nil {
		t.Fatal(err)
	}
	serviceTicketDetail, err := store.CreateServiceTicket(platform.CreateServiceTicketRequest{
		UserID:             "user_1",
		Category:           "配送问题",
		Title:              "配送问题 · 预计送达未更新",
		Content:            "骑手到店很久了",
		RelatedOrderID:     order.ID,
		RelatedOrderTitle:  "蓝海餐厅 · 招牌牛肉饭等 1 件",
		RelatedOrderStatus: platform.StatusMerchantPending,
	})
	if err != nil {
		t.Fatal(err)
	}
	serviceTicketID := serviceTicketDetail.Ticket.ID
	if _, err := store.MerchantAcceptOrder(order.ID, "merchant_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MerchantMarkOrderReady(order.ID, "merchant_1"); err != nil {
		t.Fatal(err)
	}
	assignedOrder, _, err := store.ManualAssignOrder(platform.ManualAssignOrderRequest{
		OrderID:          order.ID,
		RiderID:          "rider_2",
		StationManagerID: "station_manager_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	timeoutAt := assignedOrder.UpdatedAt.Add(2 * time.Minute)
	if _, _, err := store.TimeoutReassignOrder(platform.TimeoutReassignOrderRequest{
		OrderID:          order.ID,
		RiderID:          "rider_2",
		StationManagerID: "station_manager_1",
		TimeoutSeconds:   60,
		Now:              timeoutAt,
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := store.ReviewAfterSales(platform.ReviewAfterSalesRequest{
		RequestID: request.ID,
		Decision:  platform.AfterSalesDecisionApprove,
		Reason:    "平台仲裁通过部分退款",
		ActorID:   "admin_1",
		ActorRole: "admin",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  RoleAdmin,
		ActorID:    "support_1",
		Action:     "admin.service_ticket.assigned",
		TargetType: "service_ticket",
		TargetID:   serviceTicketID,
		RequestID:  "req_http_admin_order_detail_ticket",
		Payload:    map[string]any{"support_id": "support_1"},
		CreatedAt:  timeoutAt.Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  RoleAdmin,
		ActorID:    "admin_1",
		Action:     "after_sales.reviewed",
		TargetType: "after_sales",
		TargetID:   request.ID,
		RequestID:  "req_http_admin_order_detail_review",
		Payload:    map[string]any{"decision": "approve"},
		CreatedAt:  timeoutAt.Add(2 * time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordAuditLog(platform.RecordAuditLogRequest{
		ActorType:  RoleAdmin,
		ActorID:    "admin_1",
		Action:     "admin.order.refunded",
		TargetType: "order",
		TargetID:   order.ID,
		RequestID:  "req_http_admin_order_detail_refund",
		Payload:    map[string]any{"amount_fen": int64(800)},
		CreatedAt:  timeoutAt.Add(3 * time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/orders/"+order.ID, userToken("user_1"), http.StatusForbidden)
	detailBody := authGetJSON(t, server.URL+"/api/admin/orders/"+order.ID, adminToken("admin_1"), http.StatusOK)
	detail := detailBody["data"].(map[string]any)
	if detail["order"].(map[string]any)["id"] != order.ID || detail["order"].(map[string]any)["shop_name"] != "蓝海餐厅" {
		t.Fatalf("expected admin order detail context, got %+v", detailBody)
	}
	if detail["order"].(map[string]any)["address_snapshot"].(map[string]any)["contact_phone"] != "13800000000" {
		t.Fatalf("expected admin order detail address snapshot, got %+v", detailBody)
	}
	if detail["after_sales_summary"].(map[string]any)["total"] != float64(1) || detail["after_sales_summary"].(map[string]any)["refunded_count"] != float64(1) {
		t.Fatalf("expected admin order detail after-sales summary, got %+v", detailBody)
	}
	if detail["refund_summary"].(map[string]any)["total"] != float64(1) || detail["refund_summary"].(map[string]any)["success_count"] != float64(1) {
		t.Fatalf("expected admin order detail refund summary, got %+v", detailBody)
	}
	if detail["service_ticket_summary"].(map[string]any)["total"] != float64(1) || detail["service_ticket_summary"].(map[string]any)["open_count"] != float64(1) {
		t.Fatalf("expected admin order detail support summary, got %+v", detailBody)
	}
	if detail["dispatch_summary"].(map[string]any)["total"] != float64(3) || detail["dispatch_summary"].(map[string]any)["manual_assign_count"] != float64(1) || detail["dispatch_summary"].(map[string]any)["timeout_count"] != float64(1) || detail["dispatch_summary"].(map[string]any)["auto_assign_count"] != float64(2) {
		t.Fatalf("expected admin order detail dispatch summary, got %+v", detailBody)
	}
	if detail["audit_summary"].(map[string]any)["total"] != float64(3) || detail["audit_summary"].(map[string]any)["order_count"] != float64(1) || detail["audit_summary"].(map[string]any)["after_sales_count"] != float64(1) || detail["audit_summary"].(map[string]any)["service_ticket_count"] != float64(1) {
		t.Fatalf("expected admin order detail audit summary, got %+v", detailBody)
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

func TestPhoneRegisterAndLoginIssuesSignedToken(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	codeBody := postJSON(t, server.URL+"/api/auth/phone/code", `{"phone":"13900002222","purpose":"register"}`, http.StatusOK)
	code := codeBody["data"].(map[string]any)["dev_code"].(string)
	registerBody := postJSON(t, server.URL+"/api/auth/phone/register", `{"phone":"13900002222","code":"`+code+`","password":"Pass123","nickname":"小蓝","accepted_agreement":true}`, http.StatusOK)
	registerData := registerBody["data"].(map[string]any)
	token := registerData["access_token"].(string)
	user := registerData["user"].(map[string]any)
	if !strings.Contains(token, ".") || user["phone"] != "13900002222" {
		t.Fatalf("expected signed phone token and user phone, got %+v", registerData)
	}
	profile := authGetJSON(t, server.URL+"/api/user/profile", token, http.StatusOK)
	if profile["data"].(map[string]any)["phone"] != "13900002222" {
		t.Fatalf("expected profile to expose registered phone, got %+v", profile)
	}
	loginBody := postJSON(t, server.URL+"/api/auth/phone/login", `{"phone":"13900002222","mode":"password","password":"Pass123"}`, http.StatusOK)
	if loginBody["data"].(map[string]any)["is_new_user"] != false {
		t.Fatalf("expected password login to reuse account, got %+v", loginBody)
	}
	postJSON(t, server.URL+"/api/auth/phone/login", `{"phone":"13900002222","mode":"password","password":"wrong"}`, http.StatusUnauthorized)
}

type routerPhoneDispatcher struct {
	last platform.PhoneVerificationDispatchRequest
}

func (dispatcher *routerPhoneDispatcher) DispatchPhoneVerificationCode(req platform.PhoneVerificationDispatchRequest) (*platform.PhoneVerificationDispatchResult, error) {
	dispatcher.last = req
	return &platform.PhoneVerificationDispatchResult{Provider: "mock-sms", RequestID: req.RequestID + "_ok", Status: "delivered", SentAt: time.Now().UTC()}, nil
}

func TestPhoneCodeProviderModeHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	dispatcher := &routerPhoneDispatcher{}
	if err := store.ConfigurePhoneVerification(platform.PhoneVerificationConfig{
		Mode:            "provider",
		Provider:        "mock-sms",
		Cooldown:        time.Minute,
		ExpiresIn:       10 * time.Minute,
		MaxPerPhoneHour: 5,
		MaxPerPhoneDay:  20,
		Dispatcher:      dispatcher,
	}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	codeBody := postJSON(t, server.URL+"/api/auth/phone/code", `{"phone":"13900004444","purpose":"register"}`, http.StatusOK)
	ticket := codeBody["data"].(map[string]any)
	if _, ok := ticket["dev_code"]; ok {
		t.Fatalf("expected provider mode to hide dev code, got %+v", ticket)
	}
	if ticket["delivery_provider"] != "mock-sms" || ticket["delivery_status"] != "delivered" || dispatcher.last.Code == "" {
		t.Fatalf("expected provider dispatch metadata and hidden code, ticket=%+v dispatch=%+v", ticket, dispatcher.last)
	}
	registerBody := postJSON(t, server.URL+"/api/auth/phone/register", `{"phone":"13900004444","code":"`+dispatcher.last.Code+`","password":"Pass123","nickname":"短信用户","accepted_agreement":true}`, http.StatusOK)
	if registerBody["data"].(map[string]any)["user"].(map[string]any)["phone"] != "13900004444" {
		t.Fatalf("expected provider code register to succeed, got %+v", registerBody)
	}
	limited := postJSON(t, server.URL+"/api/auth/phone/code", `{"phone":"13900004444","purpose":"register"}`, http.StatusTooManyRequests)
	if limited["code"] != "RATE_LIMITED" {
		t.Fatalf("expected phone code cooldown to return RATE_LIMITED, got %+v", limited)
	}
}

func TestMealMatchCandidatesReportAndBlockHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	saveBody := authPutJSON(t, server.URL+"/api/meal-match/profile", userToken("user_1"), `{"gender":"female","school_id":"infinitech_university","school_name":"无限科技大学","campus_name":"东区","building_id":"east_canteen","building_name":"东区食堂","privacy_scope":"same_building","location_precision":"building_only","device_id":"device_http_user_1","identity_truth_signed":true,"platform_liability_release_signed":true,"questionnaire_completed":true,"personality_traits":["细心","守时"],"dietary_habits":["清淡","不浪费"]}`, http.StatusOK)
	savedProfile := saveBody["data"].(map[string]any)["profile"].(map[string]any)
	if saveBody["data"].(map[string]any)["can_use"] != false || savedProfile["moderation_status"] != platform.MealMatchModerationPending {
		t.Fatalf("expected profile to wait for moderation, got %+v", saveBody)
	}
	if savedProfile["device_risk_state"] != platform.MealMatchDeviceRiskPassed || savedProfile["privacy_scope"] != platform.MealMatchPrivacySameBuilding {
		t.Fatalf("expected saved profile device and privacy fields, got %+v", saveBody)
	}
	pendingCandidatesBody := authGetJSON(t, server.URL+"/api/meal-match/candidates", userToken("user_1"), http.StatusOK)
	pendingCandidatesData := pendingCandidatesBody["data"].(map[string]any)
	if pendingCandidatesData["can_use"] != false || pendingCandidatesData["review_required"] != true || len(pendingCandidatesData["candidates"].([]any)) != 0 {
		t.Fatalf("expected pending moderation to hide candidates, got %+v", pendingCandidatesBody)
	}
	queueBody := authGetJSON(t, server.URL+"/api/admin/meal-match/moderation?status=pending_review&action=profile_review", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	queueData := queueBody["data"].(map[string]any)
	profileReviews := queueData["records"].([]any)
	if len(profileReviews) != 1 {
		t.Fatalf("expected one meal match profile review, got %+v", queueBody)
	}
	reviewID := profileReviews[0].(map[string]any)["id"].(string)
	reviewBody := authPostJSON(t, server.URL+"/api/admin/meal-match/moderation/"+reviewID+"/review", roleToken(RoleSupportAdmin, "support_1"), `{"decision":"approve","review_note":"资料清晰"}`, http.StatusOK)
	if reviewBody["data"].(map[string]any)["status"] != platform.MealMatchModerationApproved {
		t.Fatalf("expected approved profile review, got %+v", reviewBody)
	}
	candidatesBody := authGetJSON(t, server.URL+"/api/meal-match/candidates", userToken("user_1"), http.StatusOK)
	candidatesData := candidatesBody["data"].(map[string]any)
	if candidatesData["can_use"] != true {
		t.Fatalf("expected meal match to be usable after review, got %+v", candidatesBody)
	}
	if candidatesData["privacy_scope"] != platform.MealMatchPrivacySameBuilding || candidatesData["device_risk_state"] != platform.MealMatchDeviceRiskPassed {
		t.Fatalf("expected privacy and device risk summary, got %+v", candidatesBody)
	}
	candidates := candidatesData["candidates"].([]any)
	if len(candidates) == 0 {
		t.Fatalf("expected meal match candidates, got %+v", candidatesBody)
	}
	for _, item := range candidates {
		candidate := item.(map[string]any)
		if candidate["same_school"] != true || candidate["same_building"] != true || candidate["distance_text"] == "距离 800m" {
			t.Fatalf("expected same-school privacy-filtered candidate, got %+v", candidate)
		}
	}
	sharedDeviceBody := authPutJSON(t, server.URL+"/api/meal-match/profile", userToken("user_2"), `{"gender":"male","school_id":"infinitech_university","building_id":"east_canteen","device_id":"device_http_user_1","identity_truth_signed":true,"platform_liability_release_signed":true,"questionnaire_completed":true,"personality_traits":["守时"],"dietary_habits":["清淡"]}`, http.StatusOK)
	sharedProfile := sharedDeviceBody["data"].(map[string]any)["profile"].(map[string]any)
	if sharedProfile["device_risk_state"] != platform.MealMatchDeviceRiskReview || sharedProfile["device_risk_reason_code"] != platform.MealMatchDeviceRiskSharedDevice {
		t.Fatalf("expected shared device to require review, got %+v", sharedDeviceBody)
	}
	blockedDeviceBody := authPutJSON(t, server.URL+"/api/meal-match/profile", userToken("user_3"), `{"gender":"female","school_id":"infinitech_university","device_id":"blocked_device_farm","identity_truth_signed":true,"platform_liability_release_signed":true,"questionnaire_completed":true,"personality_traits":["守时"],"dietary_habits":["清淡"]}`, http.StatusTooManyRequests)
	if blockedDeviceBody["code"] != "RISK_CONTROL_REJECTED" {
		t.Fatalf("expected blocked device risk rejection, got %+v", blockedDeviceBody)
	}
	targetID := candidates[0].(map[string]any)["user_id"].(string)
	reportBody := authPostJSON(t, server.URL+"/api/meal-match/reports", userToken("user_1"), `{"target_user_id":"`+targetID+`","reason":"unsafe_or_fake_profile"}`, http.StatusCreated)
	if reportBody["data"].(map[string]any)["action"] != platform.MealMatchModerationReported {
		t.Fatalf("expected reported action, got %+v", reportBody)
	}
	reportQueueBody := authGetJSON(t, server.URL+"/api/admin/meal-match/moderation?status=pending_review&action=reported&target_user_id="+targetID, roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	reportRecords := reportQueueBody["data"].(map[string]any)["records"].([]any)
	if len(reportRecords) != 1 {
		t.Fatalf("expected reported record in moderation queue, got %+v", reportQueueBody)
	}
	reportID := reportRecords[0].(map[string]any)["id"].(string)
	authPostJSON(t, server.URL+"/api/admin/meal-match/moderation/"+reportID+"/review", roleToken(RoleSupportAdmin, "support_1"), `{"decision":"approve","review_note":"举报成立"}`, http.StatusOK)
	afterReport := authGetJSON(t, server.URL+"/api/meal-match/candidates", userToken("user_1"), http.StatusOK)
	afterReportCandidates := afterReport["data"].(map[string]any)["candidates"].([]any)
	for _, item := range afterReportCandidates {
		if item.(map[string]any)["user_id"] == targetID {
			t.Fatalf("expected approved report target to disappear, got %+v", afterReport)
		}
	}
	if len(afterReportCandidates) == 0 {
		t.Fatalf("expected another candidate to test blocking, got %+v", afterReport)
	}
	blockTargetID := afterReportCandidates[0].(map[string]any)["user_id"].(string)
	blockBody := authPostJSON(t, server.URL+"/api/meal-match/blocks", userToken("user_1"), `{"target_user_id":"`+blockTargetID+`"}`, http.StatusCreated)
	if blockBody["data"].(map[string]any)["action"] != platform.MealMatchModerationBlocked {
		t.Fatalf("expected blocked action, got %+v", blockBody)
	}
	afterBlock := authGetJSON(t, server.URL+"/api/meal-match/candidates", userToken("user_1"), http.StatusOK)
	for _, item := range afterBlock["data"].(map[string]any)["candidates"].([]any) {
		if item.(map[string]any)["user_id"] == blockTargetID {
			t.Fatalf("expected blocked target to disappear, got %+v", afterBlock)
		}
	}
}

func TestRedPacketWalletAndExpiryHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":5000,"idempotency_key":"credit_red_packet_http"}`, http.StatusOK)
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	createBody := authPostJSON(t, server.URL+"/api/red-packets", userToken("user_1"), `{"scene":"group_chat","target_id":"merchant_blue_sea","type":"fixed","total_amount_fen":3000,"quantity":3,"payment_method":"balance","expires_at":"`+expiresAt+`"}`, http.StatusCreated)
	packet := createBody["data"].(map[string]any)["packet"].(map[string]any)
	packetID := packet["id"].(string)
	if packet["status"] != platform.RedPacketStatusCreated || packet["expires_at"] == "" {
		t.Fatalf("expected created packet with expiry, got %+v", createBody)
	}

	claimBody := authPostJSON(t, server.URL+"/api/red-packets/"+packetID+"/claim", userToken("user_2"), `{}`, http.StatusOK)
	if claimBody["data"].(map[string]any)["share"].(map[string]any)["amount_fen"] != float64(1000) {
		t.Fatalf("expected fixed red packet share, got %+v", claimBody)
	}
	expireAt := time.Now().UTC().Add(25 * time.Hour).Format(time.RFC3339)
	authPostJSON(t, server.URL+"/api/admin/red-packets/expire", userToken("user_1"), `{"now":"`+expireAt+`"}`, http.StatusForbidden)
	expireBody := authPostJSON(t, server.URL+"/api/admin/red-packets/expire", adminToken("admin_1"), `{"now":"`+expireAt+`"}`, http.StatusOK)
	data := expireBody["data"].(map[string]any)
	if data["count"] != float64(1) {
		t.Fatalf("expected one expired red packet refund, got %+v", expireBody)
	}
	detailBody := authGetJSON(t, server.URL+"/api/red-packets/"+packetID, userToken("user_1"), http.StatusOK)
	detailPacket := detailBody["data"].(map[string]any)["packet"].(map[string]any)
	if detailPacket["status"] != platform.RedPacketStatusExpired || detailPacket["refunded_amount_fen"] != float64(2000) {
		t.Fatalf("expected expired refunded red packet detail, got %+v", detailBody)
	}
	authPostJSON(t, server.URL+"/api/red-packets/"+packetID+"/claim", userToken("user_3"), `{}`, http.StatusConflict)
}

func TestRedPacketClaimRiskHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	authPostJSON(t, server.URL+"/api/wallet/credit", userToken("user_1"), `{"amount_fen":2000,"idempotency_key":"credit_red_packet_risk_http"}`, http.StatusOK)
	var firstPacketID string
	for index := 0; index < 3; index++ {
		createBody := authPostJSON(t, server.URL+"/api/red-packets", userToken("user_1"), `{"scene":"group_chat","target_id":"merchant_blue_sea","type":"fixed","total_amount_fen":100,"quantity":1,"payment_method":"balance"}`, http.StatusCreated)
		packetID := createBody["data"].(map[string]any)["packet"].(map[string]any)["id"].(string)
		if firstPacketID == "" {
			firstPacketID = packetID
		}
		claimBody := authPostJSON(t, server.URL+"/api/red-packets/"+packetID+"/claim", userToken("user_2"), `{}`, http.StatusOK)
		risk := claimBody["data"].(map[string]any)["risk"].(map[string]any)
		if risk["state"] != platform.RedPacketRiskPassed {
			t.Fatalf("expected passed risk check, got %+v", claimBody)
		}
	}
	authPostJSON(t, server.URL+"/api/red-packets/"+firstPacketID+"/claim", userToken("user_2"), `{}`, http.StatusOK)
	createBody := authPostJSON(t, server.URL+"/api/red-packets", userToken("user_1"), `{"scene":"group_chat","target_id":"merchant_blue_sea","type":"fixed","total_amount_fen":100,"quantity":1,"payment_method":"balance"}`, http.StatusCreated)
	packetID := createBody["data"].(map[string]any)["packet"].(map[string]any)["id"].(string)
	blockedBody := authPostJSON(t, server.URL+"/api/red-packets/"+packetID+"/claim", userToken("user_2"), `{}`, http.StatusTooManyRequests)
	if blockedBody["code"] != "RISK_CONTROL_REJECTED" {
		t.Fatalf("expected risk rejection body, got %+v", blockedBody)
	}
	detailBody := authGetJSON(t, server.URL+"/api/red-packets/"+packetID, userToken("user_2"), http.StatusOK)
	risk := detailBody["data"].(map[string]any)["risk"].(map[string]any)
	if risk["state"] != platform.RedPacketRiskBlocked || risk["reason_code"] != platform.RedPacketRiskFrequencyLimit {
		t.Fatalf("expected blocked risk detail, got %+v", detailBody)
	}
}

func TestChatSyncReadAndRealtimeOutboxHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store, WithRealtimeInternalToken("rt-secret")))
	defer server.Close()

	threadsBody := authGetJSON(t, server.URL+"/api/messages/threads", userToken("user_1"), http.StatusOK)
	if unread := chatHTTPThreadUnread(threadsBody["data"].([]any), "merchant_blue_sea"); unread != 1 {
		t.Fatalf("expected seeded unread merchant group thread, got %+v", threadsBody)
	}
	otherThreadsBody := authGetJSON(t, server.URL+"/api/messages/threads", userToken("user_2"), http.StatusOK)
	if unread := chatHTTPThreadUnread(otherThreadsBody["data"].([]any), "merchant_blue_sea"); unread != -1 {
		t.Fatalf("expected non-member user to be hidden from merchant thread, got %+v", otherThreadsBody)
	}
	overviewBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/overview", userToken("user_1"), http.StatusOK)
	overviewData := overviewBody["data"].(map[string]any)
	if overviewData["member_count"] != float64(326) || overviewData["settings_text"] != "群设置" {
		t.Fatalf("expected merchant group overview, got %+v", overviewBody)
	}
	membersBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/members", userToken("user_1"), http.StatusOK)
	membersData := membersBody["data"].(map[string]any)
	if membersData["count"] == float64(0) || !chatHTTPMemberExists(membersData["members"].([]any), "user_group_xiaolin", "小林") {
		t.Fatalf("expected merchant group members endpoint to expose active members, got %+v", membersBody)
	}
	preferenceBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/preference", userToken("user_1"), http.StatusOK)
	if preferenceBody["data"].(map[string]any)["muted"] != false {
		t.Fatalf("expected merchant group preference to default unmuted, got %+v", preferenceBody)
	}
	updatedPreferenceBody := authPutJSON(t, server.URL+"/api/messages/merchant_blue_sea/preference", userToken("user_1"), `{"muted":true}`, http.StatusOK)
	if updatedPreferenceBody["data"].(map[string]any)["muted"] != true {
		t.Fatalf("expected mute update response, got %+v", updatedPreferenceBody)
	}
	threadsBody = authGetJSON(t, server.URL+"/api/messages/threads", userToken("user_1"), http.StatusOK)
	if !chatHTTPThreadMuted(threadsBody["data"].([]any), "merchant_blue_sea") {
		t.Fatalf("expected thread list to expose mute state, got %+v", threadsBody)
	}
	authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/overview", userToken("user_2"), http.StatusNotFound)
	authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/members", userToken("user_2"), http.StatusNotFound)
	authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/preference", userToken("user_2"), http.StatusNotFound)
	authPostJSON(t, server.URL+"/internal/realtime/authorize", "wrong-secret", `{"thread_id":"merchant_blue_sea","subject_type":"user","subject_id":"user_1"}`, http.StatusUnauthorized)
	allowedBody := authPostJSON(t, server.URL+"/internal/realtime/authorize", "rt-secret", `{"thread_id":"merchant_blue_sea","subject_type":"user","subject_id":"user_1"}`, http.StatusOK)
	if allowed := allowedBody["data"].(map[string]any)["allowed"]; allowed != true {
		t.Fatalf("expected realtime member authorization to allow user_1, got %+v", allowedBody)
	}
	if muted := allowedBody["data"].(map[string]any)["muted"]; muted != true {
		t.Fatalf("expected realtime member authorization to carry mute state, got %+v", allowedBody)
	}
	deniedBody := authPostJSON(t, server.URL+"/internal/realtime/authorize", "rt-secret", `{"thread_id":"merchant_blue_sea","subject_type":"user","subject_id":"user_2"}`, http.StatusOK)
	deniedData := deniedBody["data"].(map[string]any)
	if deniedData["allowed"] != false || deniedData["reason"] != "not_member" {
		t.Fatalf("expected realtime member authorization to deny user_2 without leaking thread, got %+v", deniedBody)
	}
	syncBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/sync?mark_read=true", userToken("user_1"), http.StatusOK)
	syncData := syncBody["data"].(map[string]any)
	if len(syncData["messages"].([]any)) == 0 || syncData["unread_count"] != float64(0) || syncData["next_cursor"] == "" {
		t.Fatalf("expected initial chat sync to clear unread, got %+v", syncBody)
	}
	authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/sync?mark_read=true", userToken("user_2"), http.StatusNotFound)
	threadsBody = authGetJSON(t, server.URL+"/api/messages/threads", userToken("user_1"), http.StatusOK)
	if unread := chatHTTPThreadUnread(threadsBody["data"].([]any), "merchant_blue_sea"); unread != 0 {
		t.Fatalf("expected thread unread to clear, got %+v", threadsBody)
	}

	sent, err := store.SendChatMessage(platform.ChatMessage{ThreadID: "merchant_blue_sea", SenderID: "merchant_1", Sender: "蓝海餐厅", Content: "离线补偿消息", MessageType: "text"})
	if err != nil {
		t.Fatal(err)
	}
	offlineBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/sync?since_id="+syncData["next_cursor"].(string)+"&mark_read=false", userToken("user_1"), http.StatusOK)
	offlineData := offlineBody["data"].(map[string]any)
	if len(offlineData["messages"].([]any)) != 1 || offlineData["messages"].([]any)[0].(map[string]any)["id"] != sent.ID || offlineData["unread_count"] != float64(1) {
		t.Fatalf("expected offline sync to return new unread message, got %+v", offlineBody)
	}
	readBody := authPostJSON(t, server.URL+"/api/messages/merchant_blue_sea/read", userToken("user_1"), `{"last_message_id":"`+sent.ID+`"}`, http.StatusOK)
	if readBody["data"].(map[string]any)["unread_count"] != float64(0) {
		t.Fatalf("expected read receipt to clear unread, got %+v", readBody)
	}
	blockedBody := authPostJSON(t, server.URL+"/api/messages/merchant_blue_sea", userToken("user_1"), `{"content":"我的验证码是 123456"}`, http.StatusTooManyRequests)
	if blockedBody["code"] != "RISK_CONTROL_REJECTED" {
		t.Fatalf("expected sensitive chat message to be rejected, got %+v", blockedBody)
	}
	authPostJSON(t, server.URL+"/api/messages/merchant_blue_sea", userToken("user_2"), `{"content":"越权消息"}`, http.StatusNotFound)
	events, err := store.OutboxEvents(platform.OutboxEventsRequest{Topic: "message.sent", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].AggregateID != "merchant_blue_sea" {
		t.Fatalf("expected message.sent outbox event for realtime gateway, got %+v", events)
	}
}

func TestMerchantGroupMembershipAndCouponHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	membershipBody := authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/membership", userToken("user_2"), http.StatusOK)
	membershipData := membershipBody["data"].(map[string]any)
	if membershipData["joined"] != false || membershipData["can_join"] != true || membershipData["member_count"] != float64(326) {
		t.Fatalf("expected non-member membership payload to expose join affordance, got %+v", membershipBody)
	}
	blockedClaimBody := authPostJSON(t, server.URL+"/api/user/coupons/claim", userToken("user_2"), `{"code":"GROUP8"}`, http.StatusBadRequest)
	if blockedClaimBody["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("expected group coupon to require membership before join, got %+v", blockedClaimBody)
	}
	joinBody := authPostJSON(t, server.URL+"/api/messages/merchant_blue_sea/join", userToken("user_2"), `{}`, http.StatusOK)
	joinData := joinBody["data"].(map[string]any)
	if joinData["joined"] != true || joinData["muted"] != true || joinData["member_count"] != float64(327) {
		t.Fatalf("expected join response to add muted member and increment count, got %+v", joinBody)
	}
	threadsBody := authGetJSON(t, server.URL+"/api/messages/threads", userToken("user_2"), http.StatusOK)
	if unread := chatHTTPThreadUnread(threadsBody["data"].([]any), "merchant_blue_sea"); unread < 0 {
		t.Fatalf("expected joined user thread list to include merchant group, got %+v", threadsBody)
	}
	claimBody := authPostJSON(t, server.URL+"/api/user/coupons/claim", userToken("user_2"), `{"code":"GROUP8"}`, http.StatusCreated)
	claimData := claimBody["data"].(map[string]any)
	if claimData["amount_fen"] != float64(800) || claimData["source"] != "商户群券" {
		t.Fatalf("expected claimed merchant group coupon after join, got %+v", claimBody)
	}
	leaveBody := authPostJSON(t, server.URL+"/api/messages/merchant_blue_sea/leave", userToken("user_2"), `{}`, http.StatusOK)
	leaveData := leaveBody["data"].(map[string]any)
	if leaveData["joined"] != false || leaveData["can_join"] != true || leaveData["member_count"] != float64(326) {
		t.Fatalf("expected leave response to revoke membership, got %+v", leaveBody)
	}
	authGetJSON(t, server.URL+"/api/messages/merchant_blue_sea/sync?mark_read=true", userToken("user_2"), http.StatusNotFound)
}

func TestServiceTicketAdminAndUserClosureHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	createBody := authPostJSON(t, server.URL+"/api/service-tickets", userToken("user_1"), `{"category":"配送问题","title":"配送问题 · 预计送达未更新","content":"骑手到店很久了","related_order_id":"ord_support_http"}`, http.StatusCreated)
	ticketID := createBody["data"].(map[string]any)["ticket"].(map[string]any)["id"].(string)
	blockedEvent := authPostJSON(t, server.URL+"/api/service-tickets/"+ticketID+"/events", userToken("user_1"), `{"message":"我的银行卡号是 6222020202020202"}`, http.StatusTooManyRequests)
	if blockedEvent["code"] != "RISK_CONTROL_REJECTED" {
		t.Fatalf("expected sensitive service ticket event to be rejected, got %+v", blockedEvent)
	}
	authGetJSON(t, server.URL+"/api/admin/service-tickets?limit=10", userToken("user_1"), http.StatusForbidden)
	adminList := authGetJSON(t, server.URL+"/api/admin/service-tickets?limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(adminList["data"].([]any)) == 0 {
		t.Fatalf("expected support workbench tickets, got %+v", adminList)
	}
	authGetJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID, userToken("user_1"), http.StatusForbidden)
	adminDetail := authGetJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID, roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if adminDetail["data"].(map[string]any)["ticket"].(map[string]any)["id"] != ticketID {
		t.Fatalf("expected admin service ticket detail, got %+v", adminDetail)
	}
	filteredAdminList := authGetJSON(t, server.URL+"/api/admin/service-tickets?related_order_id=ord_support_http&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(filteredAdminList["data"].([]any)) != 1 || filteredAdminList["data"].([]any)[0].(map[string]any)["related_order_id"] != "ord_support_http" {
		t.Fatalf("expected support workbench to filter by related order, got %+v", filteredAdminList)
	}
	assignBody := authPostJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID+"/assign", roleToken(RoleSupportAdmin, "support_1"), `{"support_name":"客服小悦"}`, http.StatusOK)
	if assignBody["data"].(map[string]any)["ticket"].(map[string]any)["assigned_support_id"] != "support_1" {
		t.Fatalf("expected assigned support id, got %+v", assignBody)
	}
	escalateBody := authPostJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID+"/escalate", roleToken(RoleSupportAdmin, "support_1"), `{"reason":"超过 10 分钟未更新","escalation_level":"support_lead"}`, http.StatusOK)
	if escalateBody["data"].(map[string]any)["ticket"].(map[string]any)["sla_status"] != platform.ServiceTicketSLAStatusEscalated {
		t.Fatalf("expected escalated SLA status, got %+v", escalateBody)
	}
	escalatedList := authGetJSON(t, server.URL+"/api/admin/service-tickets?sla_status=escalated&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(escalatedList["data"].([]any)) == 0 {
		t.Fatalf("expected escalated support workbench tickets, got %+v", escalatedList)
	}
	resolveBody := authPostJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID+"/resolve", roleToken(RoleSupportAdmin, "support_1"), `{"solution":"已发放 5 元延误券，请确认处理结果"}`, http.StatusOK)
	if resolveBody["data"].(map[string]any)["ticket"].(map[string]any)["status"] != platform.ServiceTicketStatusWaitingConfirm {
		t.Fatalf("expected waiting confirm status, got %+v", resolveBody)
	}
	closeBody := authPostJSON(t, server.URL+"/api/service-tickets/"+ticketID+"/close", userToken("user_1"), `{"reason":"接受方案"}`, http.StatusOK)
	if closeBody["data"].(map[string]any)["ticket"].(map[string]any)["status"] != platform.ServiceTicketStatusClosed {
		t.Fatalf("expected closed ticket, got %+v", closeBody)
	}
	followBody := authPostJSON(t, server.URL+"/api/service-tickets/"+ticketID+"/follow-up", userToken("user_1"), `{"rating":5,"comment":"处理及时"}`, http.StatusOK)
	if followBody["data"].(map[string]any)["ticket"].(map[string]any)["follow_up_rating"] != float64(5) {
		t.Fatalf("expected follow-up rating, got %+v", followBody)
	}
	reviewBody := authPostJSON(t, server.URL+"/api/admin/service-tickets/"+ticketID+"/quality-review", roleToken(RoleSupportAdmin, "support_1"), `{"score":74,"notes":"需补充主动同步话术","coaching_required":true}`, http.StatusOK)
	if reviewBody["data"].(map[string]any)["result"] != platform.ServiceTicketQualityNeedsCoaching {
		t.Fatalf("expected quality review result, got %+v", reviewBody)
	}
	qualityList := authGetJSON(t, server.URL+"/api/admin/service-ticket-quality-reviews?support_id=support_1&coaching_required=true&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(qualityList["data"].([]any)) == 0 {
		t.Fatalf("expected quality review list, got %+v", qualityList)
	}
	performanceBody := authGetJSON(t, server.URL+"/api/admin/service-ticket-performance?support_id=support_1&limit=10", roleToken(RoleSupportAdmin, "support_1"), http.StatusOK)
	if len(performanceBody["data"].([]any)) != 1 {
		t.Fatalf("expected support performance, got %+v", performanceBody)
	}
	authPostJSON(t, server.URL+"/api/service-tickets/"+ticketID+"/close", userToken("user_2"), `{"reason":"越权"}`, http.StatusNotFound)
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
	licenseBody := authPostJSON(t, server.URL+"/api/merchant/qualifications", merchantToken, `{"type":"business_license","file_url":"https://example.test/license.jpg","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	licenseProfile := licenseBody["data"].(map[string]any)
	licenseQualifications := licenseProfile["qualifications"].([]any)
	licenseID := licenseQualifications[0].(map[string]any)["id"].(string)
	if licenseQualifications[0].(map[string]any)["status"] != platform.QualificationStatusPendingReview {
		t.Fatalf("expected merchant upload to enter review, got %+v", licenseBody)
	}
	authPostJSON(t, server.URL+"/api/admin/merchant-qualifications/"+licenseID+"/review", roleToken(RoleOpsAdmin, "ops_1"), `{"merchant_id":"`+merchantID+`","decision":"approve","reason":"营业执照核验通过","reviewed_at":"2026-05-21T12:00:00Z"}`, http.StatusOK)
	qualificationBody := authPostJSON(t, server.URL+"/api/merchant/qualifications", merchantToken, `{"type":"health_certificate","file_url":"https://example.test/health.jpg","expires_at":"`+expiresAt+`"}`, http.StatusOK)
	qualificationProfile := qualificationBody["data"].(map[string]any)
	healthQualifications := qualificationProfile["qualifications"].([]any)
	healthID := healthQualifications[1].(map[string]any)["id"].(string)
	if len(qualificationProfile["missing_qualifications"].([]any)) != 1 {
		t.Fatalf("expected health certificate to remain missing before admin review, got %+v", qualificationBody)
	}
	reviewBody := authPostJSON(t, server.URL+"/api/admin/merchant-qualifications/"+healthID+"/review", roleToken(RoleOpsAdmin, "ops_2"), `{"merchant_id":"`+merchantID+`","decision":"approve","reason":"健康证核验通过","reviewed_at":"2026-05-21T12:05:00Z"}`, http.StatusOK)
	reviewData := reviewBody["data"].(map[string]any)
	if reviewData["qualification"].(map[string]any)["status"] != platform.QualificationStatusApproved {
		t.Fatalf("expected admin qualification review approval, got %+v", reviewBody)
	}
	reviewProfile := reviewData["profile"].(map[string]any)
	if len(reviewProfile["missing_qualifications"].([]any)) != 0 {
		t.Fatalf("expected no missing qualifications after review, got %+v", reviewBody)
	}
	authPostJSON(t, server.URL+"/api/admin/merchant-qualifications/"+healthID+"/review", roleToken(RoleSupportAdmin, "support_1"), `{"merchant_id":"`+merchantID+`","decision":"reject","reason":"客服无权审核"}`, http.StatusForbidden)
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

func TestShopDetailHTTPFlow(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	detailBody := getJSON(t, server.URL+"/api/shops/shop_1/detail", http.StatusOK)
	detail := detailBody["data"].(map[string]any)
	if detail["shop_id"] != "shop_1" || detail["name"] != "蓝海餐厅" {
		t.Fatalf("expected seeded shop detail, got %+v", detailBody)
	}
	reviewSummary := detail["review_summary"].(map[string]any)
	if reviewSummary["review_count"].(float64) < 3 || reviewSummary["average_rating"] == "" {
		t.Fatalf("expected review summary data, got %+v", detailBody)
	}
	merchantInfo := detail["merchant_info"].(map[string]any)
	if merchantInfo["contact_phone"] != "13800000001" || merchantInfo["business_hours"] == "" {
		t.Fatalf("expected merchant info data, got %+v", detailBody)
	}
	reviews := detail["reviews"].([]any)
	if len(reviews) < 3 {
		t.Fatalf("expected seeded reviews, got %+v", detailBody)
	}
	if reviews[0].(map[string]any)["rider_stars_text"] == "" || len(reviews[0].(map[string]any)["item_highlights"].([]any)) == 0 {
		t.Fatalf("expected review cards to expose rider and item highlights, got %+v", detailBody)
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
	if decision["mode"] != platform.DispatchModeAutoAssign {
		t.Fatalf("expected auto assignment, got %+v", assignBody)
	}
	firstRiderID := assignedOrder["rider_id"].(string)
	nextRiderID := "rider_1"
	if firstRiderID == "rider_1" {
		nextRiderID = "rider_2"
	}
	rejectBody := authPostJSON(t, server.URL+"/api/rider/orders/"+orderID+"/reject-assignment", riderToken(firstRiderID), `{}`, http.StatusOK)
	reassignedOrder := rejectBody["data"].(map[string]any)["order"].(map[string]any)
	nextDecision := rejectBody["data"].(map[string]any)["decision"].(map[string]any)
	if reassignedOrder["rider_id"] != nextRiderID || nextDecision["candidate_rider_id"] != nextRiderID {
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
	firstRiderID := assignBody["data"].(map[string]any)["order"].(map[string]any)["rider_id"].(string)
	nextRiderID := "rider_1"
	if firstRiderID == "rider_1" {
		nextRiderID = "rider_2"
	}

	tooEarly := assignAt.Add(59 * time.Second).Format(time.RFC3339Nano)
	authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/timeout-reassign", stationManagerToken("station_manager_1"), `{"now":"`+tooEarly+`","timeout_seconds":60}`, http.StatusConflict)

	timeoutAt := assignAt.Add(60 * time.Second).Format(time.RFC3339Nano)
	timeoutBody := authPostJSON(t, server.URL+"/api/dispatch/orders/"+orderID+"/timeout-reassign", stationManagerToken("station_manager_1"), `{"now":"`+timeoutAt+`","timeout_seconds":60}`, http.StatusOK)
	reassignedOrder := timeoutBody["data"].(map[string]any)["order"].(map[string]any)
	decision := timeoutBody["data"].(map[string]any)["decision"].(map[string]any)
	if reassignedOrder["rider_id"] != nextRiderID || decision["candidate_rider_id"] != nextRiderID || decision["reason"] != "assignment_timeout" {
		t.Fatalf("expected timeout to reassign next rider, got %+v", timeoutBody)
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

func TestAdminOutboxEventDetailHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	order, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_outbox_detail_http", Type: platform.OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: order.UserID, AmountFen: 1200, IdempotencyKey: "credit_outbox_detail_http"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: order.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: order.UserID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_outbox_detail_http"}); err != nil {
		t.Fatal(err)
	}
	events, err := store.OutboxEvents(platform.OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one setup outbox event, got %+v", events)
	}
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	if _, _, err := store.MarkOutboxEventFailedWithAudit(
		platform.MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "relay down", RetryAfterSeconds: 120, Now: now},
		platform.RecordAuditLogRequest{ActorType: RoleAdmin, ActorID: "ops_1", Action: "admin.outbox.failed", TargetType: "outbox_event", TargetID: events[0].ID, RequestID: "req_http_outbox_detail", IPHash: "ip_hash"},
	); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	authGetJSON(t, server.URL+"/api/admin/outbox/events/"+events[0].ID+"?now=2026-05-22T12:00:30Z&audit_limit=5", userToken("user_1"), http.StatusForbidden)
	body := authGetJSON(t, server.URL+"/api/admin/outbox/events/"+events[0].ID+"?now=2026-05-22T12:00:30Z&audit_limit=5", adminToken("admin_1"), http.StatusOK)
	detail := body["data"].(map[string]any)
	event := detail["event"].(map[string]any)
	if event["id"] != events[0].ID || event["last_error"] != "relay down" {
		t.Fatalf("expected outbox event detail, got %+v", body)
	}
	if detail["incident_code"] != "outbox.retry_backoff" || detail["blocked"] != true || detail["retry_available_in_seconds"] != float64(90) {
		t.Fatalf("expected retry-backoff incident metadata, got %+v", detail)
	}
	recommendation := detail["recommended_operation"].(map[string]any)
	if recommendation["key"] != "outbox-replay-event" {
		t.Fatalf("expected replay recommendation, got %+v", recommendation)
	}
	recentAudits := detail["recent_audits"].([]any)
	if len(recentAudits) != 1 || recentAudits[0].(map[string]any)["action"] != "admin.outbox.failed" {
		t.Fatalf("expected recent failure audit, got %+v", recentAudits)
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
	authPostJSON(t, server.URL+"/api/reviews", userToken("user_1"), `{"order_id":"`+orderID+`","rating":5,"rider_rating":4,"content":"配送准时，餐品完好。","tags":["准时送达"]}`, http.StatusCreated)
	performanceBody := authGetJSON(t, server.URL+"/api/station-manager/rider-performance", stationManagerToken("station_manager_1"), http.StatusOK)
	performance := performanceBody["data"].([]any)
	if len(performance) != 2 || performance[0].(map[string]any)["dispatch_priority"] == float64(0) {
		t.Fatalf("expected station rider performance, got %+v", performanceBody)
	}
	if performance[0].(map[string]any)["rider_average_rating"] != float64(4) || performance[0].(map[string]any)["rider_review_count"] != float64(1) {
		t.Fatalf("expected station rider performance to surface rider review aggregation, got %+v", performanceBody)
	}
	breakdown := performance[0].(map[string]any)["score_breakdown"].(map[string]any)
	if breakdown["rating_score"] == float64(0) || breakdown["team_average_accept_seconds"] == float64(0) || breakdown["completion_score"] == float64(0) {
		t.Fatalf("expected station rider performance to surface score breakdown, got %+v", performanceBody)
	}
	recentTrend := performance[0].(map[string]any)["recent_trend"].([]any)
	if len(recentTrend) != 3 {
		t.Fatalf("expected station rider performance to surface 3-day trend, got %+v", performanceBody)
	}
	recentReviews := performance[0].(map[string]any)["recent_reviews"].([]any)
	if len(recentReviews) != 1 || recentReviews[0].(map[string]any)["content"] != "配送准时，餐品完好。" {
		t.Fatalf("expected station rider performance to surface recent review excerpts, got %+v", performanceBody)
	}
	exceptionSummary := performance[0].(map[string]any)["exception_summary"].(map[string]any)
	if exceptionSummary["dispatch_timeout_count"] != float64(0) || exceptionSummary["dispatch_reject_count"] != float64(0) || exceptionSummary["low_rating_count"] != float64(0) {
		t.Fatalf("expected station rider performance to surface zeroed exception summary for happy-path order, got %+v", performanceBody)
	}
}

func TestStationManagerRiderPerformanceExceptionDrilldownHTTPFlow(t *testing.T) {
	store := platform.NewStore(platform.DefaultHomeModules())
	dispatchOrder, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_1", Type: platform.OrderTypeTakeout, AmountFen: 1800})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_1", AmountFen: 1800, IdempotencyKey: "credit_exception_http_dispatch"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidDispatchOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_1", OrderID: dispatchOrder.ID, PaymentPassword: "123456", IdempotencyKey: "pay_exception_http_dispatch"})
	if err != nil {
		t.Fatal(err)
	}
	assignAt := paidDispatchOrder.CreatedAt.Add(10 * time.Minute)
	assignedOrder, _, err := store.AutoAssignOrder(platform.AutoAssignOrderRequest{OrderID: paidDispatchOrder.ID, Now: assignAt})
	if err != nil {
		t.Fatal(err)
	}
	riderID := assignedOrder.RiderID
	if riderID == "" {
		t.Fatalf("expected auto-assigned rider for exception drilldown, got %+v", assignedOrder)
	}
	if _, _, err := store.TimeoutReassignOrder(platform.TimeoutReassignOrderRequest{
		OrderID:        assignedOrder.ID,
		RiderID:        riderID,
		TimeoutSeconds: 60,
		Now:            assignAt.Add(60 * time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	reviewOrder, err := store.CreateOrder(platform.CreateOrderRequest{UserID: "user_2", Type: platform.OrderTypeTakeout, AmountFen: 1999})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(platform.CreditWalletRequest{UserID: "user_2", AmountFen: 1999, IdempotencyKey: "credit_exception_http_review"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(platform.SetWalletPaymentPasswordRequest{UserID: "user_2", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidReviewOrder, _, err := store.PayOrderWithBalance(platform.BalancePayRequest{UserID: "user_2", OrderID: reviewOrder.ID, PaymentPassword: "123456", IdempotencyKey: "pay_exception_http_review"})
	if err != nil {
		t.Fatal(err)
	}
	assignedReviewOrder, _, err := store.ManualAssignOrder(platform.ManualAssignOrderRequest{OrderID: paidReviewOrder.ID, RiderID: riderID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RiderMarkOrderPickedUp(assignedReviewOrder.ID, riderID); err != nil {
		t.Fatal(err)
	}
	completedReviewOrder, err := store.RiderMarkOrderDelivered(assignedReviewOrder.ID, riderID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateReview(platform.Review{
		UserID:      "user_2",
		OrderID:     completedReviewOrder.ID,
		TargetType:  platform.ReviewTargetOrder,
		TargetID:    completedReviewOrder.ID,
		Rating:      3,
		RiderRating: 2,
		Content:     "高峰期晚到了十分钟",
	}); err != nil {
		t.Fatal(err)
	}
	afterSalesRequest, err := store.CreateAfterSales(platform.CreateAfterSalesRequest{
		UserID:             "user_2",
		OrderID:            completedReviewOrder.ID,
		Reason:             "餐品撒漏",
		RequestedAmountFen: completedReviewOrder.AmountFen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.AddAfterSalesEvent(platform.AddAfterSalesEventRequest{
		RequestID: afterSalesRequest.ID,
		ActorID:   "admin_1",
		ActorRole: "admin",
		Action:    platform.AfterSalesActionCustomerCare,
		Message:   "平台已介入核实",
	}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(NewRouter(store))
	defer server.Close()

	performanceBody := authGetJSON(t, server.URL+"/api/station-manager/rider-performance", stationManagerToken("station_manager_1"), http.StatusOK)
	performance := performanceBody["data"].([]any)
	var riderPerformance map[string]any
	for _, item := range performance {
		entry := item.(map[string]any)
		if entry["rider_id"] == riderID {
			riderPerformance = entry
			break
		}
	}
	if riderPerformance == nil {
		t.Fatalf("expected %s performance drilldown, got %+v", riderID, performanceBody)
	}
	exceptionSummary := riderPerformance["exception_summary"].(map[string]any)
	if exceptionSummary["dispatch_timeout_count"] != float64(1) || exceptionSummary["after_sales_count"] != float64(1) || exceptionSummary["low_rating_count"] != float64(1) {
		t.Fatalf("expected rider performance to aggregate exception drilldown summary, got %+v", performanceBody)
	}
	exceptionDetails := riderPerformance["exception_details"].([]any)
	if len(exceptionDetails) != 3 {
		t.Fatalf("expected rider performance to expose latest exception details, got %+v", performanceBody)
	}
	if exceptionDetails[0].(map[string]any)["kind"] != "dispatch_timeout" || exceptionDetails[0].(map[string]any)["dispatch_event_id"] == "" {
		t.Fatalf("expected latest exception detail to be dispatch-timeout drilldown, got %+v", performanceBody)
	}
	if exceptionDetails[1].(map[string]any)["kind"] != "after_sales" || exceptionDetails[1].(map[string]any)["after_sales_request_id"] != afterSalesRequest.ID {
		t.Fatalf("expected second exception detail to be after-sales drilldown, got %+v", performanceBody)
	}
	if exceptionDetails[2].(map[string]any)["kind"] != "low_rating" || exceptionDetails[2].(map[string]any)["order_id"] != completedReviewOrder.ID {
		t.Fatalf("expected third exception detail to be low-rating review drilldown, got %+v", performanceBody)
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

	searchBody := getJSON(t, server.URL+"/api/search?keyword=%E7%89%9B%E8%82%89%E9%A5%AD&category=all", http.StatusOK)
	searchData := searchBody["data"].(map[string]any)
	searchResults := searchData["results"].([]any)
	if searchData["total"] == float64(0) || len(searchResults) == 0 {
		t.Fatalf("expected anonymous search to return real catalog results, got %+v", searchBody)
	}
	guessBody := getJSON(t, server.URL+"/api/search?keyword=&category=all", http.StatusOK)
	guessData := guessBody["data"].(map[string]any)
	guessSuggestions := guessData["suggestions"].([]any)
	if len(guessSuggestions) == 0 || guessSuggestions[0] == "" {
		t.Fatalf("expected anonymous empty search to return backend suggestions, got %+v", guessBody)
	}

	addressBody := authPostJSON(t, server.URL+"/api/user/addresses", userToken("user_1"), `{"contact_name":"张三","contact_phone":"13800000000","city":"北京","detail":"望京SOHO","latitude":39.99,"longitude":116.48,"tag":"home","is_default":true}`, http.StatusCreated)
	addressID := addressBody["data"].(map[string]any)["id"].(string)

	cartBody := authPostJSON(t, server.URL+"/api/cart/items", userToken("user_1"), `{"shop_id":"shop_1","product_id":"prod_beef_rice","quantity":2}`, http.StatusOK)
	if cartBody["data"].(map[string]any)["payable_fen"] != float64(5598) || cartBody["data"].(map[string]any)["shop_name"] != "蓝海餐厅" {
		t.Fatalf("expected cart payable 5598, got %+v", cartBody)
	}
	cartGetBody := authGetJSON(t, server.URL+"/api/cart?shop_id=shop_1", userToken("user_1"), http.StatusOK)
	if cartGetBody["data"].(map[string]any)["items_total_fen"] != float64(5198) {
		t.Fatalf("expected cart items total 5198, got %+v", cartGetBody)
	}

	checkoutBody := authPostJSON(t, server.URL+"/api/orders/checkout", userToken("user_1"), `{"shop_id":"shop_1","address_id":"`+addressID+`","options":{"remark":"少放辣","tableware_count":2}}`, http.StatusCreated)
	order := checkoutBody["data"].(map[string]any)["order"].(map[string]any)
	orderID := order["id"].(string)
	if order["status"] != platform.StatusPendingPayment || order["amount_fen"] != float64(5598) || order["shop_name"] != "蓝海餐厅" {
		t.Fatalf("expected pending payment checkout order, got %+v", checkoutBody)
	}
	addressSnapshot := order["address_snapshot"].(map[string]any)
	if addressSnapshot["contact_name"] != "张三" || addressSnapshot["detail"] != "望京SOHO" {
		t.Fatalf("expected order address snapshot, got %+v", checkoutBody)
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
	detail := detailBody["data"].(map[string]any)
	if detail["id"] != orderID || detail["shop_name"] != "蓝海餐厅" {
		t.Fatalf("expected order detail, got %+v", detailBody)
	}
	detailAddress := detail["address_snapshot"].(map[string]any)
	if detailAddress["contact_phone"] != "13800000000" {
		t.Fatalf("expected order detail address snapshot, got %+v", detailBody)
	}
}

func TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate(t *testing.T) {
	server := httptest.NewServer(NewRouter(platform.NewStore(platform.DefaultHomeModules())))
	defer server.Close()

	orderBody := authPostJSON(t, server.URL+"/api/orders", userToken("user_1"), `{"type":"takeout","amount_fen":1990}`, http.StatusCreated)
	orderID := orderBody["data"].(map[string]any)["id"].(string)
	uploadBody := authPostJSON(t, server.URL+"/api/reviews/upload-ticket", userToken("user_1"), `{"order_id":"`+orderID+`","file_name":"review.jpg","content_type":"image/jpeg","size_bytes":1024}`, http.StatusCreated)
	uploadTicket := uploadBody["data"].(map[string]any)
	if uploadTicket["ticket_id"] == "" || uploadTicket["public_url"] == "" || !strings.HasPrefix(uploadTicket["object_key"].(string), "reviews/") {
		t.Fatalf("expected review image upload ticket, got %+v", uploadBody)
	}
	confirmBody := authPostJSON(t, server.URL+"/api/reviews/upload-confirm", userToken("user_1"), `{"ticket_id":"`+uploadTicket["ticket_id"].(string)+`","object_key":"`+uploadTicket["object_key"].(string)+`","content_type":"image/jpeg","size_bytes":1024,"content_sha":"sha256:review-http"}`, http.StatusCreated)
	confirmed := confirmBody["data"].(map[string]any)
	if confirmed["status"] != platform.AfterSalesUploadTicketConfirmed || confirmed["public_url"] != uploadTicket["public_url"] {
		t.Fatalf("expected confirmed review image ticket, got %+v", confirmBody)
	}

	createdBody := authPostJSON(t, server.URL+"/api/reviews", userToken("user_1"), `{"order_id":"`+orderID+`","rating":4,"rider_rating":5,"content":"整体体验不错","tags":["出餐快","包装完整"],"anonymous":true,"image_urls":["`+confirmed["public_url"].(string)+`"],"item_ratings":[{"product_id":"prod_beef_rice","product_name":"招牌牛肉饭","rating":5,"tags":["分量在线","值得回购"]}]}`, http.StatusCreated)
	created := createdBody["data"].(map[string]any)
	if created["target_type"] != platform.ReviewTargetOrder || created["target_id"] != orderID || created["anonymous"] != true {
		t.Fatalf("expected order review to bind order and keep anonymous flag, got %+v", createdBody)
	}
	if len(created["image_urls"].([]any)) != 1 || len(created["item_ratings"].([]any)) != 1 || created["rider_rating"] != float64(5) {
		t.Fatalf("expected review to keep image urls and item ratings, got %+v", createdBody)
	}

	filteredBody := authGetJSON(t, server.URL+"/api/reviews?order_id="+orderID, userToken("user_1"), http.StatusOK)
	filtered := filteredBody["data"].([]any)
	if len(filtered) != 1 || filtered[0].(map[string]any)["id"] != created["id"] || filtered[0].(map[string]any)["anonymous"] != true {
		t.Fatalf("expected one order-scoped anonymous review, got %+v", filteredBody)
	}
	if len(filtered[0].(map[string]any)["image_urls"].([]any)) != 1 || len(filtered[0].(map[string]any)["item_ratings"].([]any)) != 1 {
		t.Fatalf("expected filtered review to retain images and item ratings, got %+v", filteredBody)
	}

	reviewedOrderBody := authGetJSON(t, server.URL+"/api/orders/"+orderID, userToken("user_1"), http.StatusOK)
	if reviewedOrderBody["data"].(map[string]any)["reviewed"] != true {
		t.Fatalf("expected order detail to surface reviewed state, got %+v", reviewedOrderBody)
	}

	updatedBody := authPostJSON(t, server.URL+"/api/reviews", userToken("user_1"), `{"order_id":"`+orderID+`","rating":5,"rider_rating":3,"content":"更新成五星好评","tags":["味道不错"],"anonymous":false,"image_urls":["`+confirmed["public_url"].(string)+`"],"item_ratings":[{"product_id":"prod_beef_rice","product_name":"招牌牛肉饭","rating":4,"tags":["口味稳定"]}]}`, http.StatusCreated)
	updated := updatedBody["data"].(map[string]any)
	if updated["id"] != created["id"] || updated["content"] != "更新成五星好评" || updated["anonymous"] != false {
		t.Fatalf("expected review update in place, got created=%+v updated=%+v", createdBody, updatedBody)
	}

	filteredBody = authGetJSON(t, server.URL+"/api/reviews?order_id="+orderID, userToken("user_1"), http.StatusOK)
	filtered = filteredBody["data"].([]any)
	if len(filtered) != 1 || filtered[0].(map[string]any)["content"] != "更新成五星好评" || filtered[0].(map[string]any)["anonymous"] != false {
		t.Fatalf("expected updated order-scoped review, got %+v", filteredBody)
	}
	if len(filtered[0].(map[string]any)["image_urls"].([]any)) != 1 || filtered[0].(map[string]any)["item_ratings"].([]any)[0].(map[string]any)["rating"] != float64(4) || filtered[0].(map[string]any)["rider_rating"] != float64(3) {
		t.Fatalf("expected updated review image urls and item ratings, got %+v", filteredBody)
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

type merchantQualificationReviewAtomicAuditStore struct {
	*platform.Store
	atomicAuditCalled         bool
	reviewQualificationCalled bool
	recordAuditCalled         bool
	atomicReq                 platform.ReviewMerchantQualificationRequest
	atomicAudit               platform.RecordAuditLogRequest
}

func (store *merchantQualificationReviewAtomicAuditStore) ReviewMerchantQualificationWithAudit(req platform.ReviewMerchantQualificationRequest, audit platform.RecordAuditLogRequest) (*platform.MerchantProfile, *platform.MerchantQualification, *platform.AuditLog, *platform.OutboxEvent, error) {
	store.atomicAuditCalled = true
	store.atomicReq = req
	store.atomicAudit = audit
	profile := &platform.MerchantProfile{
		Account:               platform.MerchantAccount{ID: req.MerchantID, Type: platform.MerchantAccountStandard, Status: platform.ShopStatusActive, DepositStatus: platform.DepositStatusPaid},
		MissingQualifications: []string{},
		CanAcceptOrders:       true,
	}
	qualification := &platform.MerchantQualification{
		ID:        req.QualificationID,
		Type:      platform.QualificationBusinessLicense,
		FileURL:   "https://example.test/license.jpg",
		ExpiresAt: time.Date(2027, 5, 23, 12, 0, 0, 0, time.UTC),
		Status:    platform.QualificationStatusApproved,
	}
	profile.Qualifications = []platform.MerchantQualification{*qualification}
	auditLog := &platform.AuditLog{ID: "aud_qualification_review_atomic_http", Action: audit.Action, TargetType: audit.TargetType, TargetID: audit.TargetID}
	outboxEvent := &platform.OutboxEvent{ID: "obe_qualification_review_atomic_http", Topic: "merchant.qualification_reviewed", AggregateType: "merchant_qualification", AggregateID: req.QualificationID, EventType: "merchant.qualification.reviewed", Status: platform.OutboxStatusPending}
	return profile, qualification, auditLog, outboxEvent, nil
}

func (store *merchantQualificationReviewAtomicAuditStore) ReviewMerchantQualification(req platform.ReviewMerchantQualificationRequest) (*platform.MerchantProfile, *platform.MerchantQualification, error) {
	store.reviewQualificationCalled = true
	return nil, nil, platform.ErrInvalidArgument
}

func (store *merchantQualificationReviewAtomicAuditStore) RecordAuditLog(req platform.RecordAuditLogRequest) (*platform.AuditLog, error) {
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

func signedNotificationProviderCallbackJSON(t *testing.T, url string, payload notificationProviderCallbackPayload, secret string, expectedStatus int) map[string]any {
	t.Helper()
	payload.Signature = signNotificationProviderCallback(payload, secret)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return authPostJSON(t, url, "", string(body), expectedStatus)
}

type staticWechatMiniResolver struct {
	session WechatMiniSession
	err     error
}

func (resolver staticWechatMiniResolver) Resolve(_ context.Context, _ string) (WechatMiniSession, error) {
	return resolver.session, resolver.err
}
