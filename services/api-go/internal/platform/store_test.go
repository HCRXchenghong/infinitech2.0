package platform

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBalancePaymentIsIdempotentAndLedgerBacked(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 2000, IdempotencyKey: "credit_1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}

	tx1, paidOrder1, account1, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_1"})
	if err != nil {
		t.Fatal(err)
	}
	tx2, paidOrder2, account2, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_1"})
	if err != nil {
		t.Fatal(err)
	}

	if tx1.ID != tx2.ID {
		t.Fatalf("expected idempotent transaction, got %s and %s", tx1.ID, tx2.ID)
	}
	if paidOrder1.Status != StatusDispatching || paidOrder2.Status != StatusDispatching {
		t.Fatalf("expected paid order to enter dispatching, got %s/%s", paidOrder1.Status, paidOrder2.Status)
	}
	if account1.Balance != 800 || account2.Balance != 800 {
		t.Fatalf("expected balance 800, got %d/%d", account1.Balance, account2.Balance)
	}
}

func TestRefundSettingsAndBalanceRefundFlow(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 2000, IdempotencyKey: "credit_refund"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_refund"}); err != nil || paidOrder.Status != StatusDispatching || account.Balance != 800 {
		t.Fatalf("expected paid order and debited wallet, order=%+v account=%+v err=%v", paidOrder, account, err)
	}

	settings, err := store.RefundSettings()
	if err != nil {
		t.Fatal(err)
	}
	if settings.DefaultStrategy != RefundStrategyBalanceFirst {
		t.Fatalf("expected balance-first default, got %+v", settings)
	}

	refund1, refundedOrder1, refundedAccount1, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        order.ID,
		Reason:         "商品售罄",
		IdempotencyKey: "refund_balance_1",
		ActorID:        "admin_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	refund2, refundedOrder2, refundedAccount2, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        order.ID,
		Reason:         "商品售罄重复回放",
		IdempotencyKey: "refund_balance_1",
		ActorID:        "admin_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if refund1.ID != refund2.ID || refundedOrder1.Status != StatusRefunded || refundedOrder2.Status != StatusRefunded {
		t.Fatalf("expected idempotent balance refund, refund=%+v/%+v order=%+v/%+v", refund1, refund2, refundedOrder1, refundedOrder2)
	}
	if refund1.Status != RefundStatusSuccess || refund1.Destination != RefundDestinationBalance {
		t.Fatalf("expected successful balance refund, got %+v", refund1)
	}
	if refundedAccount1.Balance != 2000 || refundedAccount2.Balance != 2000 {
		t.Fatalf("expected no double credit on refund replay, got %d/%d", refundedAccount1.Balance, refundedAccount2.Balance)
	}
	refundEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.refunded", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(refundEvents) != 1 || refundEvents[0].AggregateID != order.ID {
		t.Fatalf("expected one order.refunded outbox event, got %+v", refundEvents)
	}

	if _, err := store.SaveRefundSettings(SaveRefundSettingsRequest{DefaultStrategy: RefundStrategyOriginalFirst}); err != nil {
		t.Fatal(err)
	}
	originalOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_2", Type: OrderTypeTakeout, AmountFen: 900})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_2", AmountFen: 900, IdempotencyKey: "credit_refund_original"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_2", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_2", OrderID: originalOrder.ID, PaymentPassword: "123456", IdempotencyKey: "pay_refund_original"}); err != nil {
		t.Fatal(err)
	}
	originalRefund, pendingOrder, pendingAccount, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        originalOrder.ID,
		Reason:         "用户申请原路返回",
		IdempotencyKey: "refund_original_1",
		ActorID:        "admin_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if originalRefund.Status != RefundStatusPendingOriginal || originalRefund.Destination != RefundDestinationOriginalRoute || pendingOrder.Status != StatusRefundPending || pendingAccount != nil {
		t.Fatalf("expected original-route refund to stay pending without wallet credit, refund=%+v order=%+v account=%+v", originalRefund, pendingOrder, pendingAccount)
	}
	pendingEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "payment.refund.requested", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(pendingEvents) != 1 || pendingEvents[0].AggregateID != originalOrder.ID {
		t.Fatalf("expected one payment.refund.requested event, got %+v", pendingEvents)
	}
}

func TestRefundOrderWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("order-refund-audit-secret")
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_atomic_refund", Type: OrderTypeTakeout, AmountFen: 1500})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: order.UserID, AmountFen: 1500, IdempotencyKey: "credit_refund_atomic_audit"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: order.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: order.UserID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_refund_atomic_audit"}); err != nil || paidOrder.Status != StatusDispatching {
		t.Fatalf("expected paid order setup, order=%+v err=%v", paidOrder, err)
	}
	createdAt := time.Date(2026, 5, 23, 11, 0, 0, 123456789, time.UTC)

	refund, refundedOrder, account, audit, err := store.RefundOrderWithAudit(
		RefundOrderRequest{
			OrderID:        order.ID,
			Reason:         "商品售罄",
			IdempotencyKey: "refund_order_atomic_audit",
			ActorID:        "admin_1",
			ActorRole:      "admin",
		},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.order.refunded",
			TargetType: "order",
			TargetID:   order.ID,
			RequestID:  "req_order_refund_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"amount_fen": int64(1),
				"token":      "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if refund.Status != RefundStatusSuccess || refundedOrder.Status != StatusRefunded || account.Balance != 1500 {
		t.Fatalf("expected successful audited refund, refund=%+v order=%+v account=%+v", refund, refundedOrder, account)
	}
	if audit.Action != "admin.order.refunded" || audit.TargetType != "order" || audit.TargetID != order.ID {
		t.Fatalf("expected order refund audit target, got %+v", audit)
	}
	if audit.Payload["refund_id"] != refund.ID || audit.Payload["amount_fen"] != refund.AmountFen || audit.Payload["idempotency_key"] != refund.IdempotencyKey || audit.Payload["token"] != nil {
		t.Fatalf("expected atomic refund audit payload to be server-generated and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 123456000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "admin.order.refunded", TargetType: "order", TargetID: order.ID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["refund_id"] != refund.ID || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified refund audit log from atomic write, got %+v", logs)
	}
}

func TestSaveRefundSettingsWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("refund-settings-audit-secret")
	createdAt := time.Date(2026, 5, 23, 10, 30, 0, 987654321, time.UTC)

	settings, audit, err := store.SaveRefundSettingsWithAudit(
		SaveRefundSettingsRequest{DefaultStrategy: RefundStrategyOriginalFirst},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.refund_settings.updated",
			TargetType: "refund_settings",
			TargetID:   "default",
			RequestID:  "req_refund_settings_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"default_refund_strategy": RefundStrategyBalanceFirst,
				"amount_fen":              int64(999),
				"token":                   "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if settings.DefaultStrategy != RefundStrategyOriginalFirst {
		t.Fatalf("expected normalized refund settings, got %+v", settings)
	}
	if audit.Action != "admin.refund_settings.updated" || audit.TargetType != "refund_settings" || audit.TargetID != "default" {
		t.Fatalf("expected refund-settings audit target, got %+v", audit)
	}
	if audit.Payload["default_refund_strategy"] != RefundStrategyOriginalFirst || audit.Payload["amount_fen"] != nil || audit.Payload["token"] != nil {
		t.Fatalf("expected atomic audit payload to be server-normalized and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 987654000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "admin.refund_settings.updated", TargetType: "refund_settings", TargetID: "default", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["default_refund_strategy"] != RefundStrategyOriginalFirst || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified audit log from atomic write, got %+v", logs)
	}
}

func TestAdminOperationsSnapshotAggregatesP0Data(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	dispatchOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 900})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 900, IdempotencyKey: "credit_admin_snapshot_dispatch"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: dispatchOrder.ID, PaymentPassword: "123456", IdempotencyKey: "pay_admin_snapshot_dispatch"}); err != nil || paidOrder.Status != StatusDispatching {
		t.Fatalf("expected dispatching order setup, order=%+v err=%v", paidOrder, err)
	}
	if _, err := store.SetRiderOnlineStatus(SetRiderOnlineStatusRequest{RiderID: "rider_1", Online: true, Capacity: 2, DistanceMeters: 500}); err != nil {
		t.Fatal(err)
	}
	assignedOrder, decision, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: dispatchOrder.ID, Now: dispatchOrder.CreatedAt.Add((DispatchGrabHallSeconds + 1) * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if assignedOrder.Status != StatusRiderAssigned || decision.CandidateRiderID == "" {
		t.Fatalf("expected auto-assigned order, order=%+v decision=%+v", assignedOrder, decision)
	}

	lat := 39.99
	lng := 116.48
	address, err := store.SaveAddress(UserAddress{
		UserID:       "user_2",
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
	if _, err := store.UpsertCartItem(UpsertCartItemRequest{UserID: "user_2", ShopID: "shop_1", ProductID: "prod_beef_rice", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	merchantOrder, _, err := store.CheckoutCart(CheckoutCartRequest{UserID: "user_2", ShopID: "shop_1", AddressID: address.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_2", AmountFen: merchantOrder.AmountFen, IdempotencyKey: "credit_admin_snapshot_after_sales"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_2", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidMerchantOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_2", OrderID: merchantOrder.ID, PaymentPassword: "123456", IdempotencyKey: "pay_admin_snapshot_after_sales"}); err != nil || paidMerchantOrder.Status != StatusMerchantPending {
		t.Fatalf("expected merchant-pending order setup, order=%+v err=%v", paidMerchantOrder, err)
	}
	afterSales, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_2",
		OrderID:            merchantOrder.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: merchantOrder.AmountFen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if afterSales.Status != AfterSalesPendingMerchant {
		t.Fatalf("expected pending after-sales request, got %+v", afterSales)
	}

	snapshot, err := store.AdminOperationsSnapshot(AdminOperationsSnapshotRequest{
		Now:                        time.Now().UTC().Add(20 * time.Minute),
		Limit:                      5,
		LeaseExpiringWithinSeconds: 60,
		ObjectCleanupGraceSeconds:  60,
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Counts.TotalOrders != 2 || snapshot.Counts.PendingMerchantOrders != 1 || snapshot.Counts.RiderAssignedOrders != 1 {
		t.Fatalf("expected order counts in snapshot, got %+v", snapshot.Counts)
	}
	if snapshot.Counts.AfterSalesPending != 1 || len(snapshot.AfterSales) != 1 || snapshot.AfterSales[0].ID != afterSales.ID {
		t.Fatalf("expected after-sales queue in snapshot, counts=%+v afterSales=%+v", snapshot.Counts, snapshot.AfterSales)
	}
	if len(snapshot.Merchants) == 0 || len(snapshot.Riders) == 0 || len(snapshot.RiderPerformance) == 0 {
		t.Fatalf("expected merchants riders and performance in snapshot, got %+v", snapshot)
	}
	if snapshot.Counts.DispatchEventCount == 0 || len(snapshot.DispatchEvents) == 0 {
		t.Fatalf("expected dispatch events in snapshot, counts=%+v events=%+v", snapshot.Counts, snapshot.DispatchEvents)
	}
	if snapshot.OutboxStats.Total == 0 || snapshot.Counts.OutboxReady == 0 {
		t.Fatalf("expected outbox health in snapshot, stats=%+v counts=%+v", snapshot.OutboxStats, snapshot.Counts)
	}
	if snapshot.RefundSettings.DefaultStrategy != RefundStrategyBalanceFirst {
		t.Fatalf("expected refund settings in snapshot, got %+v", snapshot.RefundSettings)
	}
}

func TestPartialRefundsAccumulateWithoutOverRefund(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 2000, IdempotencyKey: "credit_partial_refund"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_partial_refund"}); err != nil || paidOrder.Status != StatusDispatching || account.Balance != 800 {
		t.Fatalf("expected paid order setup, order=%+v account=%+v err=%v", paidOrder, account, err)
	}

	firstRefund, firstOrder, firstAccount, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        order.ID,
		AmountFen:      500,
		Reason:         "少送一份小食",
		IdempotencyKey: "partial_refund_1",
		ActorID:        "admin_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstRefund.AmountFen != 500 || firstOrder.Status != StatusDispatching || firstAccount.Balance != 1300 {
		t.Fatalf("expected partial refund to keep order active, refund=%+v order=%+v account=%+v", firstRefund, firstOrder, firstAccount)
	}
	refundEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.refunded", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(refundEvents) != 1 || refundEvents[0].Payload["amount_fen"] != int64(500) || refundEvents[0].Payload["order_amount_fen"] != int64(1200) {
		t.Fatalf("expected partial refund outbox to expose refund amount separately, got %+v", refundEvents)
	}
	if _, _, _, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        order.ID,
		AmountFen:      701,
		Reason:         "超过剩余可退金额",
		IdempotencyKey: "partial_refund_over",
		ActorID:        "admin_1",
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected over-refund to be rejected, got %v", err)
	}

	finalRefund, finalOrder, finalAccount, err := store.RefundOrder(RefundOrderRequest{
		OrderID:        order.ID,
		Reason:         "退回剩余金额",
		IdempotencyKey: "partial_refund_remaining",
		ActorID:        "admin_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if finalRefund.AmountFen != 700 || finalOrder.Status != StatusRefunded || finalAccount.Balance != 2000 {
		t.Fatalf("expected default refund to use remaining amount, refund=%+v order=%+v account=%+v", finalRefund, finalOrder, finalAccount)
	}
}

func TestAfterSalesReviewApprovesAndRefundsBalance(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[order.ID].ShopID = "shop_1"
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_after_sales"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_after_sales"}); err != nil || paidOrder.Status != StatusMerchantPending || account.Balance != 0 {
		t.Fatalf("expected paid merchant order setup, order=%+v account=%+v err=%v", paidOrder, account, err)
	}

	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Type:               AfterSalesRefundOnly,
		Reason:             "餐品漏送",
		RequestedAmountFen: 1200,
		EvidenceURLs:       []string{"https://cdn.test/evidence.jpg", "https://cdn.test/evidence.jpg"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Status != AfterSalesPendingMerchant || len(request.EvidenceURLs) != 1 {
		t.Fatalf("expected sanitized pending after-sales request, got %+v", request)
	}
	if _, err := store.CreateAfterSales(CreateAfterSalesRequest{UserID: "user_1", OrderID: order.ID, Reason: "重复售后"}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected duplicate active after-sales to be blocked, got %v", err)
	}
	merchantRequests, err := store.MerchantAfterSalesRequests("merchant_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(merchantRequests) != 1 || merchantRequests[0].ID != request.ID {
		t.Fatalf("expected merchant after-sales list to include request, got %+v", merchantRequests)
	}

	reviewed, refund, refundedOrder, refundedAccount, err := store.ReviewAfterSales(ReviewAfterSalesRequest{
		RequestID: request.ID,
		Decision:  AfterSalesDecisionApprove,
		Reason:    "确认漏送",
		ActorID:   "merchant_1",
		ActorRole: "merchant",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Status != AfterSalesRefunded || reviewed.RefundID != refund.ID || reviewed.ReviewerID != "merchant_1" {
		t.Fatalf("expected approved after-sales to bind refund, request=%+v refund=%+v", reviewed, refund)
	}
	if refund.Status != RefundStatusSuccess || refund.Destination != RefundDestinationBalance || refundedOrder.Status != StatusRefunded || refundedAccount.Balance != 1200 {
		t.Fatalf("expected balance refund through after-sales, refund=%+v order=%+v account=%+v", refund, refundedOrder, refundedAccount)
	}
	if _, _, _, _, err := store.ReviewAfterSales(ReviewAfterSalesRequest{RequestID: request.ID, Decision: AfterSalesDecisionReject, Reason: "重复", ActorID: "merchant_1", ActorRole: "merchant"}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected reviewed after-sales to reject replay decisions, got %v", err)
	}
	afterSalesEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.after_sales", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(afterSalesEvents) != 2 || afterSalesEvents[0].AggregateID != order.ID {
		t.Fatalf("expected created and approved after-sales outbox events, got %+v", afterSalesEvents)
	}
}

func TestReviewAfterSalesWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("after-sales-review-audit-secret")
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_after_sales_atomic", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[order.ID].ShopID = "shop_1"
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: order.UserID, AmountFen: 1200, IdempotencyKey: "credit_after_sales_atomic_audit"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: order.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: order.UserID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_after_sales_atomic_audit"}); err != nil || paidOrder.Status != StatusMerchantPending {
		t.Fatalf("expected paid order setup, order=%+v err=%v", paidOrder, err)
	}
	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             order.UserID,
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2026, 5, 23, 12, 15, 0, 654321987, time.UTC)

	reviewed, refund, reviewedOrder, account, audit, err := store.ReviewAfterSalesWithAudit(
		ReviewAfterSalesRequest{
			RequestID:            request.ID,
			Decision:             AfterSalesDecisionApprove,
			Reason:               "确认漏送",
			ActorID:              "merchant_1",
			ActorRole:            "merchant",
			RefundIdempotencyKey: "after_sales_atomic_audit",
		},
		RecordAuditLogRequest{
			ActorType:  "merchant",
			ActorID:    "merchant_1",
			Action:     "after_sales.reviewed",
			TargetType: "after_sales",
			TargetID:   request.ID,
			RequestID:  "req_after_sales_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"decision":   AfterSalesDecisionReject,
				"amount_fen": int64(1),
				"token":      "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Status != AfterSalesRefunded || refund.Status != RefundStatusSuccess || reviewedOrder.Status != StatusMerchantPending || account.Balance != 600 {
		t.Fatalf("expected audited after-sales review refund, request=%+v refund=%+v order=%+v account=%+v", reviewed, refund, reviewedOrder, account)
	}
	if audit.Action != "after_sales.reviewed" || audit.TargetType != "after_sales" || audit.TargetID != request.ID {
		t.Fatalf("expected after-sales review audit target, got %+v", audit)
	}
	if audit.Payload["decision"] != AfterSalesDecisionApprove ||
		audit.Payload["status"] != AfterSalesRefunded ||
		audit.Payload["refund_id"] != refund.ID ||
		audit.Payload["amount_fen"] != refund.AmountFen ||
		audit.Payload["destination"] != refund.Destination ||
		audit.Payload["idempotency_key"] != refund.IdempotencyKey ||
		audit.Payload["token"] != nil {
		t.Fatalf("expected after-sales review audit payload to be server-generated and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 654321000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "after_sales.reviewed", TargetType: "after_sales", TargetID: request.ID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["refund_id"] != refund.ID || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified after-sales review audit log from atomic write, got %+v", logs)
	}
}

func TestPartialAfterSalesRefundsAccumulateWithoutOverRefund(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[order.ID].ShopID = "shop_1"
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_partial_after_sales"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_partial_after_sales"}); err != nil || paidOrder.Status != StatusMerchantPending || account.Balance != 0 {
		t.Fatalf("expected paid merchant order setup, order=%+v account=%+v err=%v", paidOrder, account, err)
	}

	firstRequest, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Type:               AfterSalesRefundOnly,
		Reason:             "少送一杯饮品",
		RequestedAmountFen: 500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstRequest.Type != AfterSalesPartialRefund {
		t.Fatalf("expected partial amount to normalize after-sales type, got %+v", firstRequest)
	}
	firstReviewed, firstRefund, firstOrder, firstAccount, err := store.ReviewAfterSales(ReviewAfterSalesRequest{
		RequestID: firstRequest.ID,
		Decision:  AfterSalesDecisionApprove,
		Reason:    "确认少送",
		ActorID:   "merchant_1",
		ActorRole: "merchant",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstReviewed.Status != AfterSalesRefunded || firstRefund.AmountFen != 500 || firstOrder.Status != StatusMerchantPending || firstAccount.Balance != 500 {
		t.Fatalf("expected first partial after-sales refund to keep order active, request=%+v refund=%+v order=%+v account=%+v", firstReviewed, firstRefund, firstOrder, firstAccount)
	}
	if _, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "超过剩余可退金额",
		RequestedAmountFen: 701,
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected after-sales over-refund to be rejected, got %v", err)
	}

	secondRequest, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "再退剩余金额",
		RequestedAmountFen: 700,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondReviewed, secondRefund, secondOrder, secondAccount, err := store.ReviewAfterSales(ReviewAfterSalesRequest{
		RequestID: secondRequest.ID,
		Decision:  AfterSalesDecisionApprove,
		Reason:    "确认补退",
		ActorID:   "merchant_1",
		ActorRole: "merchant",
	})
	if err != nil {
		t.Fatal(err)
	}
	if secondReviewed.Status != AfterSalesRefunded || secondRefund.AmountFen != 700 || secondOrder.Status != StatusRefunded || secondAccount.Balance != 1200 {
		t.Fatalf("expected cumulative after-sales refunds to close order exactly once, request=%+v refund=%+v order=%+v account=%+v", secondReviewed, secondRefund, secondOrder, secondAccount)
	}
}

func TestAfterSalesEventsEscalateToAdminReviewAndAuditTimeline(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	headServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodHead {
			t.Errorf("expected object verifier to use HEAD, got %s", req.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.Contains(req.URL.Path, "/after-sales-test/after-sales/") {
			t.Errorf("expected object verifier path to include bucket and after-sales key, got %s", req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", "2097152")
		w.WriteHeader(http.StatusOK)
	}))
	defer headServer.Close()
	if err := store.ConfigureObjectStorage(ObjectStorageConfig{
		Provider:                ObjectStorageProviderMinIO,
		Bucket:                  "after-sales-test",
		UploadBaseURL:           "https://minio.test/upload",
		PublicBaseURL:           "https://cdn.test/assets",
		HeadBaseURL:             headServer.URL + "/objects",
		SigningSecret:           "test-storage-secret",
		MaxUploadBytes:          AfterSalesEvidenceMaxBytes,
		RequireHeadVerification: true,
	}); err != nil {
		t.Fatal(err)
	}
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[order.ID].ShopID = "shop_1"
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_after_sales_events"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_after_sales_events"}); err != nil || paidOrder.Status != StatusMerchantPending {
		t.Fatalf("expected paid merchant order, order=%+v err=%v", paidOrder, err)
	}
	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.OrderAmountFen != 1200 || request.RefundedAmountFen != 0 || request.RefundableFen != 1200 {
		t.Fatalf("expected after-sales request to expose refund window, got %+v", request)
	}
	ticket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "../evidence photo.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   2 * 1024 * 1024,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.TicketID == "" || ticket.Provider != ObjectStorageProviderMinIO || ticket.Bucket != "after-sales-test" || ticket.Method != "PUT" || ticket.MaxSizeBytes != AfterSalesEvidenceMaxBytes || !strings.Contains(ticket.ObjectKey, "after-sales/"+request.ID+"/") || !strings.HasPrefix(ticket.UploadURL, "https://minio.test/upload/after-sales-test/") || !strings.HasPrefix(ticket.PublicURL, "https://cdn.test/assets/") || !strings.HasSuffix(ticket.PublicURL, ".jpg") {
		t.Fatalf("expected signed upload ticket for after-sales evidence, got %+v", ticket)
	}
	if ticket.Headers["X-Object-Bucket"] != "after-sales-test" || ticket.Headers["X-Upload-Signature"] == "" {
		t.Fatalf("expected object storage headers to include bucket and signature, got %+v", ticket.Headers)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		ObjectKey:   "after-sales/" + request.ID + "/forged/evidence.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   1024,
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected unissued same-prefix evidence object to be rejected, got %v", err)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   1025,
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected evidence confirmation with mismatched ticket size to be rejected, got %v", err)
	}
	evidence, evidenceEvent, evidencedRequest, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2 * 1024 * 1024,
		ContentSHA:  "sha256:test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Status != AfterSalesEvidenceUploaded || evidence.PublicURL != ticket.PublicURL || evidence.ContentSHA != "sha256:test" {
		t.Fatalf("expected confirmed after-sales evidence, got %+v", evidence)
	}
	if evidenceEvent.Action != AfterSalesActionEvidenceUploaded || len(evidenceEvent.Attachments) != 1 || len(evidencedRequest.EvidenceURLs) != 1 {
		t.Fatalf("expected evidence upload audit event and request evidence URL, event=%+v request=%+v", evidenceEvent, evidencedRequest)
	}
	evidenceList, err := store.AfterSalesEvidence(request.ID, "merchant_1", "merchant")
	if err != nil {
		t.Fatal(err)
	}
	if len(evidenceList) != 1 || evidenceList[0].ID != evidence.ID {
		t.Fatalf("expected merchant-visible evidence list, got %+v", evidenceList)
	}
	duplicateEvidence, duplicateEvent, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2 * 1024 * 1024,
	})
	if err != nil || duplicateEvidence.ID != evidence.ID || duplicateEvent != nil {
		t.Fatalf("expected duplicate evidence confirmation to be idempotent, evidence=%+v event=%+v err=%v", duplicateEvidence, duplicateEvent, err)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		ObjectKey:   "after-sales/other_request/sig/evidence.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   1024,
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected evidence outside request prefix to be rejected, got %v", err)
	}
	if _, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "too-large.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   AfterSalesEvidenceMaxBytes + 1,
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected oversized evidence upload to be rejected, got %v", err)
	}

	visible := false
	if _, _, err := store.AddAfterSalesEvent(AddAfterSalesEventRequest{
		RequestID:     request.ID,
		ActorID:       "merchant_1",
		ActorRole:     "merchant",
		Action:        AfterSalesActionInternalNote,
		Message:       "商户不能写内部客服备注",
		VisibleToUser: &visible,
	}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected merchant internal note to be blocked, got %v", err)
	}
	escalation, escalatedRequest, err := store.AddAfterSalesEvent(AddAfterSalesEventRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		Action:      AfterSalesActionCustomerCare,
		Message:     "申请客服介入",
		Attachments: []string{"https://cdn.test/customer-care.jpg", "https://cdn.test/customer-care.jpg"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if escalation.Action != AfterSalesActionCustomerCare || len(escalation.Attachments) != 1 || escalatedRequest.Status != AfterSalesAdminReview {
		t.Fatalf("expected customer care escalation with sanitized evidence, event=%+v request=%+v", escalation, escalatedRequest)
	}
	if _, _, _, _, err := store.ReviewAfterSales(ReviewAfterSalesRequest{RequestID: request.ID, Decision: AfterSalesDecisionApprove, Reason: "商户不能审核已介入售后", ActorID: "merchant_1", ActorRole: "merchant"}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected merchant review to be blocked after customer care escalation, got %v", err)
	}
	reviewed, refund, _, account, err := store.ReviewAfterSales(ReviewAfterSalesRequest{
		RequestID: request.ID,
		Decision:  AfterSalesDecisionApprove,
		Reason:    "平台仲裁通过部分退款",
		ActorID:   "admin_1",
		ActorRole: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Status != AfterSalesRefunded || refund.AmountFen != 600 || account.Balance != 600 {
		t.Fatalf("expected admin arbitration to refund requested amount, request=%+v refund=%+v account=%+v", reviewed, refund, account)
	}
	if reviewed.RefundedAmountFen != 600 || reviewed.RefundableFen != 600 {
		t.Fatalf("expected reviewed after-sales to expose updated refund window, got %+v", reviewed)
	}
	events, err := store.AfterSalesEvents(request.ID, "admin_1", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 || events[0].Action != AfterSalesActionCreated || events[1].Action != AfterSalesActionEvidenceUploaded || events[2].Action != AfterSalesActionCustomerCare || events[3].Action != AfterSalesActionReviewApproved {
		t.Fatalf("expected created, escalation and review audit timeline, got %+v", events)
	}
}

func TestAfterSalesEvidenceRequiresUploadCallbackAndScanApproval(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	if err := store.ConfigureObjectStorage(ObjectStorageConfig{
		Provider:                        ObjectStorageProviderMinIO,
		Bucket:                          "after-sales-test",
		UploadBaseURL:                   "https://minio.test/upload",
		PublicBaseURL:                   "https://cdn.test/assets",
		SigningSecret:                   "test-storage-secret",
		CallbackSigningSecret:           "test-callback-secret",
		MaxUploadBytes:                  AfterSalesEvidenceMaxBytes,
		RequireUploadCallbackForConfirm: true,
		RequireScanApprovalForConfirm:   true,
	}); err != nil {
		t.Fatal(err)
	}
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_after_sales_scan"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_after_sales_scan"}); err != nil {
		t.Fatal(err)
	}
	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "evidence.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected confirmation before upload callback to be blocked, got %v", err)
	}

	storage := store.objectStorageSnapshot()
	uploadedAt := time.Date(2026, 5, 22, 17, 0, 0, 0, time.UTC)
	uploadSignature := storage.signObjectUploadCallback(objectUploadCallbackSignatureInput{
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
		ContentSHA:  "sha256:evidence",
		UploadedAt:  uploadedAt,
	})
	uploadedTicket, err := store.ConfirmObjectStorageUpload(ObjectStorageUploadCallbackRequest{
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
		ContentSHA:  "sha256:evidence",
		UploadedAt:  uploadedAt,
		Signature:   uploadSignature,
	})
	if err != nil {
		t.Fatal(err)
	}
	if uploadedTicket.Status != AfterSalesUploadTicketUploaded || uploadedTicket.ScanStatus != AfterSalesUploadScanPending || uploadedTicket.ContentSHA != "sha256:evidence" || !uploadedTicket.UploadedAt.Equal(uploadedAt) {
		t.Fatalf("expected uploaded ticket pending scan, got %+v", uploadedTicket)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected confirmation before scan approval to be blocked, got %v", err)
	}

	scannedAt := uploadedAt.Add(time.Minute)
	scanSignature := storage.signObjectScanResult(objectScanResultSignatureInput{
		TicketID:      ticket.TicketID,
		ObjectKey:     ticket.ObjectKey,
		ScanStatus:    AfterSalesUploadScanPassed,
		ScanResult:    "clean",
		Scanner:       "clamav",
		ScanCheckedAt: scannedAt,
	})
	scannedTicket, err := store.RecordObjectStorageScanResult(ObjectStorageScanResultRequest{
		TicketID:      ticket.TicketID,
		ObjectKey:     ticket.ObjectKey,
		ScanStatus:    AfterSalesUploadScanPassed,
		ScanResult:    "clean",
		Scanner:       "clamav",
		ScanCheckedAt: scannedAt,
		Signature:     scanSignature,
	})
	if err != nil {
		t.Fatal(err)
	}
	if scannedTicket.ScanStatus != AfterSalesUploadScanPassed || scannedTicket.ScanResult != "clean" || !scannedTicket.ScanCheckedAt.Equal(scannedAt) {
		t.Fatalf("expected scan approval to be recorded, got %+v", scannedTicket)
	}
	evidence, event, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    ticket.TicketID,
		ObjectKey:   ticket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
		ContentSHA:  "sha256:evidence",
	})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Status != AfterSalesEvidenceUploaded || event.Action != AfterSalesActionEvidenceUploaded {
		t.Fatalf("expected scanned evidence to confirm, evidence=%+v event=%+v", evidence, event)
	}
	if _, err := store.RecordObjectStorageScanResult(ObjectStorageScanResultRequest{
		TicketID:   ticket.TicketID,
		ObjectKey:  ticket.ObjectKey,
		ScanStatus: AfterSalesUploadScanRejected,
		Signature:  "bad",
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid scan signature to be rejected, got %v", err)
	}
}

func TestObjectStorageCleanupCandidatesAndCompletion(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_cleanup"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_cleanup"}); err != nil {
		t.Fatal(err)
	}
	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             "user_1",
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	expiredTicket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
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
	rejectedTicket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "rejected.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	confirmedTicket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		FileName:    "confirmed.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 22, 18, 0, 0, 0, time.UTC)
	store.mu.Lock()
	store.afterSalesUploadTickets[expiredTicket.TicketID].ExpiresAt = now.Add(-3 * time.Hour)
	store.afterSalesUploadTickets[rejectedTicket.TicketID].Status = AfterSalesUploadTicketUploaded
	store.afterSalesUploadTickets[rejectedTicket.TicketID].ScanStatus = AfterSalesUploadScanRejected
	store.afterSalesUploadTickets[rejectedTicket.TicketID].ScanResult = "stream: Eicar-Test-Signature FOUND"
	store.afterSalesUploadTickets[rejectedTicket.TicketID].ScanCheckedAt = now.Add(-2 * time.Hour)
	store.afterSalesUploadTickets[confirmedTicket.TicketID].Status = AfterSalesUploadTicketConfirmed
	store.afterSalesUploadTickets[confirmedTicket.TicketID].ConfirmedAt = now.Add(-time.Hour)
	store.mu.Unlock()

	limited, err := store.ObjectStorageCleanupCandidates(ObjectStorageCleanupCandidatesRequest{
		Limit:        1,
		GraceSeconds: 60,
		Now:          now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) != 1 || limited[0].TicketID != expiredTicket.TicketID || limited[0].Reason != AfterSalesObjectCleanupExpired {
		t.Fatalf("expected oldest expired cleanup candidate first, got %+v", limited)
	}
	candidates, err := store.ObjectStorageCleanupCandidates(ObjectStorageCleanupCandidatesRequest{
		GraceSeconds: 60,
		Now:          now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 || candidates[1].TicketID != rejectedTicket.TicketID || candidates[1].Reason != AfterSalesObjectCleanupRejected {
		t.Fatalf("expected expired and rejected cleanup candidates only, got %+v", candidates)
	}
	failedAt := now.Add(30 * time.Second)
	failed, err := store.RecordObjectStorageCleanupFailure(ObjectStorageCleanupFailureRequest{
		TicketID:  rejectedTicket.TicketID,
		ObjectKey: rejectedTicket.ObjectKey,
		Reason:    AfterSalesObjectCleanupRejected,
		Error:     "delete denied",
		FailedAt:  failedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if failed.CleanupAttempts != 1 || failed.LastCleanupError != "delete denied" || !failed.LastCleanupFailedAt.Equal(failedAt) {
		t.Fatalf("expected cleanup failure to be recorded, got %+v", failed)
	}
	candidatesAfterFailure, err := store.ObjectStorageCleanupCandidates(ObjectStorageCleanupCandidatesRequest{
		GraceSeconds: 60,
		Now:          now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	var failedCandidate ObjectStorageCleanupCandidate
	for _, candidate := range candidatesAfterFailure {
		if candidate.TicketID == rejectedTicket.TicketID {
			failedCandidate = candidate
		}
	}
	if failedCandidate.TicketID == "" || failedCandidate.CleanupAttempts != 1 || failedCandidate.LastCleanupError != "delete denied" || !failedCandidate.LastCleanupFailedAt.Equal(failedAt) {
		t.Fatalf("expected cleanup candidate to expose failure ledger, got %+v", failedCandidate)
	}
	stats, err := store.ObjectStorageCleanupStats(ObjectStorageCleanupCandidatesRequest{
		GraceSeconds: 60,
		Now:          now.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pending != 2 || stats.ExpiredUnconfirmed != 1 || stats.ScanRejected != 1 || stats.Failed != 1 || stats.Deleted != 0 || stats.CleanupAttempts != 1 {
		t.Fatalf("expected cleanup stats to expose pending and failed counts, got %+v", stats)
	}
	deletedAt := now.Add(time.Minute)
	cleaned, err := store.CompleteObjectStorageCleanup(ObjectStorageCleanupCompleteRequest{
		TicketID:  rejectedTicket.TicketID,
		ObjectKey: rejectedTicket.ObjectKey,
		Reason:    AfterSalesObjectCleanupRejected,
		DeletedAt: deletedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cleaned.Status != AfterSalesUploadTicketDeleted || cleaned.CleanupReason != AfterSalesObjectCleanupRejected || !cleaned.DeletedAt.Equal(deletedAt) {
		t.Fatalf("expected rejected object cleanup to be recorded, got %+v", cleaned)
	}
	if cleaned.CleanupAttempts != 1 || cleaned.LastCleanupError != "" || !cleaned.LastCleanupFailedAt.IsZero() {
		t.Fatalf("expected successful cleanup to retain attempt count and clear last failure, got %+v", cleaned)
	}
	statsAfterCleanup, err := store.ObjectStorageCleanupStats(ObjectStorageCleanupCandidatesRequest{
		GraceSeconds: 60,
		Now:          now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if statsAfterCleanup.Pending != 1 || statsAfterCleanup.Failed != 0 || statsAfterCleanup.Deleted != 1 || !statsAfterCleanup.LastDeletedAt.Equal(deletedAt) {
		t.Fatalf("expected cleanup stats to reflect completed deletion, got %+v", statsAfterCleanup)
	}
	replayed, err := store.CompleteObjectStorageCleanup(ObjectStorageCleanupCompleteRequest{
		TicketID:  rejectedTicket.TicketID,
		ObjectKey: rejectedTicket.ObjectKey,
		Reason:    AfterSalesObjectCleanupRejected,
		DeletedAt: deletedAt.Add(time.Minute),
	})
	if err != nil || replayed.Status != AfterSalesUploadTicketDeleted {
		t.Fatalf("expected cleanup completion to be idempotent, got ticket=%+v err=%v", replayed, err)
	}
	if _, _, _, err := store.ConfirmAfterSalesEvidenceUpload(ConfirmAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     "user_1",
		ActorRole:   "user",
		TicketID:    rejectedTicket.TicketID,
		ObjectKey:   rejectedTicket.ObjectKey,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected deleted object ticket to be blocked from confirmation, got %v", err)
	}
	if _, err := store.CompleteObjectStorageCleanup(ObjectStorageCleanupCompleteRequest{
		TicketID:  confirmedTicket.TicketID,
		ObjectKey: confirmedTicket.ObjectKey,
		Reason:    AfterSalesObjectCleanupExpired,
		DeletedAt: deletedAt,
	}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected confirmed object cleanup to be blocked, got %v", err)
	}
}

func TestCompleteObjectStorageCleanupWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("object-cleanup-audit-secret")
	ticket := createObjectCleanupAuditTicket(t, store, "complete", "expired-audit.jpg")
	deletedAt := time.Date(2026, 5, 23, 14, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 23, 14, 5, 0, 111222333, time.UTC)

	cleaned, audit, err := store.CompleteObjectStorageCleanupWithAudit(
		ObjectStorageCleanupCompleteRequest{
			TicketID:  ticket.TicketID,
			ObjectKey: ticket.ObjectKey,
			Reason:    AfterSalesObjectCleanupExpired,
			DeletedAt: deletedAt,
		},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.object_cleanup.completed",
			TargetType: "object_storage_ticket",
			TargetID:   ticket.TicketID,
			RequestID:  "req_object_cleanup_complete_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"object_key":       "caller/object/key/should-not-persist.jpg",
				"reason":           AfterSalesObjectCleanupRejected,
				"status":           AfterSalesUploadTicketConfirmed,
				"cleanup_attempts": 99,
				"token":            "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if cleaned.Status != AfterSalesUploadTicketDeleted || cleaned.CleanupReason != AfterSalesObjectCleanupExpired || !cleaned.DeletedAt.Equal(deletedAt) {
		t.Fatalf("expected audited object cleanup to mark ticket deleted, got %+v", cleaned)
	}
	if audit.Action != "admin.object_cleanup.completed" || audit.TargetType != "object_storage_ticket" || audit.TargetID != ticket.TicketID {
		t.Fatalf("expected object cleanup completed audit target, got %+v", audit)
	}
	if audit.Payload["object_key"] != maskAuditScalar(cleaned.ObjectKey) ||
		audit.Payload["reason"] != AfterSalesObjectCleanupExpired ||
		audit.Payload["status"] != AfterSalesUploadTicketDeleted ||
		audit.Payload["cleanup_attempts"] != nil ||
		audit.Payload["token"] != nil {
		t.Fatalf("expected cleanup-completed audit payload to be server-generated and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 111222000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "admin.object_cleanup.completed", TargetType: "object_storage_ticket", TargetID: ticket.TicketID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["object_key"] != maskAuditScalar(cleaned.ObjectKey) || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified object cleanup completed audit log from atomic write, got %+v", logs)
	}
}

func TestRecordObjectStorageCleanupFailureWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("object-cleanup-audit-secret")
	ticket := createObjectCleanupAuditTicket(t, store, "failure", "rejected-audit.jpg")
	failedAt := time.Date(2026, 5, 23, 14, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 23, 14, 35, 0, 444555666, time.UTC)

	failed, audit, err := store.RecordObjectStorageCleanupFailureWithAudit(
		ObjectStorageCleanupFailureRequest{
			TicketID:  ticket.TicketID,
			ObjectKey: ticket.ObjectKey,
			Reason:    AfterSalesObjectCleanupRejected,
			Error:     "delete denied by provider",
			FailedAt:  failedAt,
		},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.object_cleanup.failed",
			TargetType: "object_storage_ticket",
			TargetID:   ticket.TicketID,
			RequestID:  "req_object_cleanup_failed_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"object_key":       "caller/object/key/should-not-persist.jpg",
				"reason":           AfterSalesObjectCleanupExpired,
				"status":           AfterSalesUploadTicketDeleted,
				"cleanup_attempts": 99,
				"token":            "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != AfterSalesUploadTicketIssued || failed.CleanupReason != AfterSalesObjectCleanupRejected || failed.CleanupAttempts != 1 || failed.LastCleanupError != "delete denied by provider" || !failed.LastCleanupFailedAt.Equal(failedAt) {
		t.Fatalf("expected audited object cleanup failure to update failure ledger, got %+v", failed)
	}
	if audit.Action != "admin.object_cleanup.failed" || audit.TargetType != "object_storage_ticket" || audit.TargetID != ticket.TicketID {
		t.Fatalf("expected object cleanup failed audit target, got %+v", audit)
	}
	if audit.Payload["object_key"] != maskAuditScalar(failed.ObjectKey) ||
		audit.Payload["reason"] != AfterSalesObjectCleanupRejected ||
		audit.Payload["status"] != AfterSalesUploadTicketIssued ||
		audit.Payload["cleanup_attempts"] != 1 ||
		audit.Payload["token"] != nil {
		t.Fatalf("expected cleanup-failed audit payload to be server-generated and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 444555000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "admin.object_cleanup.failed", TargetType: "object_storage_ticket", TargetID: ticket.TicketID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["cleanup_attempts"] != 1 || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified object cleanup failed audit log from atomic write, got %+v", logs)
	}
}

func createObjectCleanupAuditTicket(t *testing.T, store *Store, suffix string, fileName string) *ObjectUploadTicket {
	t.Helper()
	userID := "user_cleanup_audit_" + suffix
	order, err := store.CreateOrder(CreateOrderRequest{UserID: userID, Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: userID, AmountFen: 1200, IdempotencyKey: "credit_cleanup_audit_" + suffix}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: userID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: userID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_cleanup_audit_" + suffix}); err != nil {
		t.Fatal(err)
	}
	request, err := store.CreateAfterSales(CreateAfterSalesRequest{
		UserID:             userID,
		OrderID:            order.ID,
		Reason:             "餐品漏送",
		RequestedAmountFen: 600,
	})
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := store.CreateAfterSalesEvidenceUpload(CreateAfterSalesEvidenceUploadRequest{
		RequestID:   request.ID,
		ActorID:     userID,
		ActorRole:   "user",
		FileName:    fileName,
		ContentType: "image/jpeg",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	return ticket
}

func TestWechatMiniLoginCreatesStableUserBinding(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	first, err := store.LoginWechatMini(WechatMiniLoginRequest{Code: "wx_code_1", Nickname: "小蓝"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.LoginWechatMini(WechatMiniLoginRequest{Code: "wx_code_1", Nickname: "小蓝新版"})
	if err != nil {
		t.Fatal(err)
	}
	if !first.IsNewUser || second.IsNewUser {
		t.Fatalf("expected only first login to be new, got first=%+v second=%+v", first, second)
	}
	if first.User.ID != second.User.ID || second.User.Nickname != "小蓝新版" {
		t.Fatalf("expected stable user binding and profile update, got first=%+v second=%+v", first, second)
	}
}

func TestWechatMiniLoginUsesProviderOpenID(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	first, err := store.LoginWechatMini(WechatMiniLoginRequest{Code: "wx_code_1", ProviderOpenID: "real_openid_1", ProviderUnionID: "union_1", Nickname: "小蓝"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.LoginWechatMini(WechatMiniLoginRequest{Code: "wx_code_2", ProviderOpenID: "real_openid_1", Nickname: "小蓝新版"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ProviderOpenID != "real_openid_1" || first.ProviderUnionID != "union_1" {
		t.Fatalf("expected provider identity in login result, got %+v", first)
	}
	if first.User.ID != second.User.ID || second.IsNewUser {
		t.Fatalf("expected provider openid to drive stable binding, got first=%+v second=%+v", first, second)
	}
}

func TestWechatPrepayAndCallbackAreIdempotent(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 900})
	if err != nil {
		t.Fatal(err)
	}
	prepay, tx, err := store.CreateWechatPrepay(WechatPrepayRequest{UserID: "user_1", OrderID: order.ID})
	if err != nil {
		t.Fatal(err)
	}
	if prepay.OutTradeNo == "" || tx.Status != "prepay_created" {
		t.Fatalf("unexpected prepay result: prepay=%+v tx=%+v", prepay, tx)
	}
	prepayAgain, txAgain, err := store.CreateWechatPrepay(WechatPrepayRequest{UserID: "user_1", OrderID: order.ID})
	if err != nil {
		t.Fatal(err)
	}
	if prepay.OutTradeNo != prepayAgain.OutTradeNo || tx.ID != txAgain.ID {
		t.Fatalf("expected prepay idempotency, got %s/%s", prepay.OutTradeNo, prepayAgain.OutTradeNo)
	}
	paidTx, paidOrder, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_tx_1",
		AmountFen:     900,
	})
	if err != nil {
		t.Fatal(err)
	}
	if paidTx.Status != "success" || paidOrder.Status != StatusDispatching || paidOrder.PaymentMethod != PaymentWechat {
		t.Fatalf("expected wechat payment success, tx=%+v order=%+v", paidTx, paidOrder)
	}
	duplicateTx, duplicateOrder, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_tx_1",
		AmountFen:     900,
	})
	if err != nil {
		t.Fatal(err)
	}
	if duplicateTx.ID != paidTx.ID || duplicateOrder.ID != paidOrder.ID {
		t.Fatalf("expected callback idempotency, tx=%+v order=%+v", duplicateTx, duplicateOrder)
	}
}

func TestMerchantInviteRegistrationAndQualificationGate(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	invite, err := store.CreateMerchantInvite(CreateMerchantInviteRequest{AdminID: "admin_1"})
	if err != nil {
		t.Fatal(err)
	}
	profile, err := store.AcceptMerchantInvite(AcceptMerchantInviteRequest{
		Token:       invite.Token,
		DisplayName: "蓝海商户",
		AccountType: MerchantAccountStandard,
		Password:    "MerchantPass123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if profile.Account.ID == "" || profile.CanAcceptOrders || len(profile.MissingQualifications) != 2 {
		t.Fatalf("expected merchant to require qualifications and deposit, got %+v", profile)
	}
	loggedInMerchant, err := store.LoginMerchant(MerchantLoginRequest{AccountID: profile.Account.ID, Password: "MerchantPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if loggedInMerchant.Account.ID != profile.Account.ID {
		t.Fatalf("expected merchant account login, got %+v", loggedInMerchant)
	}
	if _, err := store.LoginMerchant(MerchantLoginRequest{AccountID: profile.Account.ID, Password: "bad-password"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected bad merchant password to be rejected, got %v", err)
	}
	if _, err := store.AcceptMerchantInvite(AcceptMerchantInviteRequest{Token: invite.Token, DisplayName: "重复注册", AccountType: MerchantAccountStandard, Password: "MerchantPass123"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected used invite to be rejected, got %v", err)
	}
	expiresAt := time.Now().UTC().Add(365 * 24 * time.Hour)
	profile, err = store.SaveMerchantQualification(UploadMerchantQualificationRequest{
		MerchantID: profile.Account.ID,
		Type:       QualificationBusinessLicense,
		FileURL:    "https://example.test/license.jpg",
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.MissingQualifications) != 1 || profile.MissingQualifications[0] != QualificationHealthCertificate {
		t.Fatalf("expected health certificate still missing, got %+v", profile)
	}
	profile, err = store.SaveMerchantQualification(UploadMerchantQualificationRequest{
		MerchantID: profile.Account.ID,
		Type:       QualificationHealthCertificate,
		FileURL:    "https://example.test/health.jpg",
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.MissingQualifications) != 0 || profile.CanAcceptOrders {
		t.Fatalf("expected qualifications complete but deposit still gating orders, got %+v", profile)
	}
}

func TestMerchantStaffAndSupplementalMaterialsAreScoped(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	expiresAt := time.Now().UTC().Add(365 * 24 * time.Hour)

	seededStaff, err := store.MerchantStaff("merchant_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(seededStaff) != 1 || seededStaff[0].ShopID != "shop_1" {
		t.Fatalf("expected seeded merchant staff, got %+v", seededStaff)
	}
	staff, err := store.SaveMerchantStaff(UpsertMerchantStaffRequest{
		MerchantID:                 "merchant_1",
		ShopID:                     "shop_1",
		Name:                       "李四",
		Phone:                      "13900000000",
		Role:                       "kitchen",
		HealthCertificateURL:       "https://example.test/staff-health.jpg",
		HealthCertificateExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if staff.ID == "" || staff.MerchantID != "merchant_1" || staff.ShopID != "shop_1" || staff.Status != MerchantStaffActive {
		t.Fatalf("expected scoped staff to be saved, got %+v", staff)
	}
	if _, err := store.SaveMerchantStaff(UpsertMerchantStaffRequest{
		MerchantID:                 "merchant_2",
		ShopID:                     "shop_1",
		Name:                       "越权员工",
		Phone:                      "13900000001",
		HealthCertificateURL:       "https://example.test/staff-health.jpg",
		HealthCertificateExpiresAt: expiresAt,
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-merchant staff write to be hidden, got %v", err)
	}
	if _, err := store.SaveMerchantStaff(UpsertMerchantStaffRequest{
		MerchantID:                 "merchant_1",
		ShopID:                     "shop_1",
		Name:                       "过期员工",
		Phone:                      "13900000002",
		HealthCertificateURL:       "https://example.test/staff-health.jpg",
		HealthCertificateExpiresAt: time.Now().UTC().Add(-time.Hour),
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected expired staff health certificate to be rejected, got %v", err)
	}

	material, err := store.SaveMerchantSupplementalMaterial(UploadMerchantSupplementalMaterialRequest{
		MerchantID:  "merchant_1",
		ShopID:      "shop_1",
		Type:        "storefront_photo",
		FileURL:     "https://example.test/storefront.jpg",
		Description: "门头照",
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if material.ID == "" || material.Status != "submitted" || material.ShopID != "shop_1" {
		t.Fatalf("expected supplemental material to be submitted, got %+v", material)
	}
	if _, err := store.SaveMerchantSupplementalMaterial(UploadMerchantSupplementalMaterialRequest{
		MerchantID:  "merchant_2",
		ShopID:      "shop_1",
		Type:        "storefront_photo",
		FileURL:     "https://example.test/storefront.jpg",
		Description: "越权材料",
		ExpiresAt:   expiresAt,
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-merchant material write to be hidden, got %v", err)
	}

	profile, err := store.MerchantProfile("merchant_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Staff) < 2 || len(profile.SupplementalMaterials) < 2 {
		t.Fatalf("expected merchant profile to include staff and supplemental materials, got %+v", profile)
	}
}

func TestRiderAndStationManagerInviteRegistration(t *testing.T) {
	store := NewStore(DefaultHomeModules())

	riderInvite, err := store.CreateRiderInvite(CreateRiderInviteRequest{
		CreatedByID:   "station_manager_1",
		CreatedByRole: RiderAccountStationManager,
		Type:          RiderAccountRider,
		StationID:     "station_1",
	})
	if err != nil {
		t.Fatal(err)
	}
	rider, err := store.AcceptRiderInvite(AcceptRiderInviteRequest{Token: riderInvite.Token, Password: "RiderPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if rider.ID != "rider_3" || rider.Type != RiderAccountRider || rider.StationID != "station_1" || rider.DepositStatus != DepositStatusUnpaid {
		t.Fatalf("expected invited rider in station_1 with unpaid deposit, got %+v", rider)
	}
	loggedInRider, err := store.LoginRider(RiderLoginRequest{AccountID: rider.ID, Password: "RiderPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if loggedInRider.ID != rider.ID || loggedInRider.Type != RiderAccountRider {
		t.Fatalf("expected rider account login, got %+v", loggedInRider)
	}
	if _, err := store.LoginRider(RiderLoginRequest{AccountID: rider.ID, Password: "bad-password"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected bad rider password to be rejected, got %v", err)
	}
	if _, err := store.AcceptRiderInvite(AcceptRiderInviteRequest{Token: riderInvite.Token, Password: "RiderPass123"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected used rider invite to be rejected, got %v", err)
	}
	if _, err := store.CreateRiderInvite(CreateRiderInviteRequest{
		CreatedByID:   "station_manager_1",
		CreatedByRole: RiderAccountStationManager,
		Type:          RiderAccountStationManager,
		StationID:     "station_1",
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected station manager to be unable to invite station manager, got %v", err)
	}

	managerInvite, err := store.CreateRiderInvite(CreateRiderInviteRequest{
		CreatedByID:   "admin_1",
		CreatedByRole: "admin",
		Type:          RiderAccountStationManager,
		StationID:     "station_2",
	})
	if err != nil {
		t.Fatal(err)
	}
	manager, err := store.AcceptRiderInvite(AcceptRiderInviteRequest{Token: managerInvite.Token, Password: "StationPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if manager.ID != "station_manager_4" || manager.Type != RiderAccountStationManager || manager.StationID != "station_2" {
		t.Fatalf("expected invited station manager in station_2, got %+v", manager)
	}
	loggedInManager, err := store.LoginRider(RiderLoginRequest{AccountID: manager.ID, Password: "StationPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if loggedInManager.ID != manager.ID || loggedInManager.Type != RiderAccountStationManager {
		t.Fatalf("expected station manager account login, got %+v", loggedInManager)
	}
}

func TestCreateMerchantInviteWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("invite-admin-audit-secret")
	expiresAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	invite, audit, err := store.CreateMerchantInviteWithAudit(
		CreateMerchantInviteRequest{AdminID: "admin_1", ExpiresAt: expiresAt},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.merchant_invite.created",
			TargetType: "merchant_invite",
			TargetID:   "pending",
			RequestID:  "req_merchant_invite",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"expires_at": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
				"token":      "caller-token-must-not-persist",
			},
			CreatedAt: time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if invite.Token == "" || audit.TargetID != invite.Token || audit.TargetType != "merchant_invite" {
		t.Fatalf("expected merchant invite audit to target generated token, invite=%+v audit=%+v", invite, audit)
	}
	if audit.Payload["type"] != OnboardingInviteMerchant || audit.Payload["expires_at"] != invite.ExpiresAt.Format(time.RFC3339Nano) || audit.Payload["token"] != nil {
		t.Fatalf("expected server-generated merchant invite audit payload, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified merchant invite audit proof, got %+v", audit)
	}
	logs, err := store.AuditLogs(AuditLogsRequest{TargetType: "merchant_invite", TargetID: invite.Token, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Action != "admin.merchant_invite.created" || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable merchant invite audit log, got %+v", logs)
	}
}

func TestCreateRiderInviteWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("invite-rider-audit-secret")
	expiresAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)

	invite, audit, err := store.CreateRiderInviteWithAudit(
		CreateRiderInviteRequest{
			CreatedByID:   "station_manager_1",
			CreatedByRole: RiderAccountStationManager,
			Type:          RiderAccountRider,
			StationID:     "station_1",
			ExpiresAt:     expiresAt,
		},
		RecordAuditLogRequest{
			ActorType:  RiderAccountStationManager,
			ActorID:    "station_manager_1",
			Action:     "admin.rider_invite.created",
			TargetType: "rider_invite",
			TargetID:   "pending",
			RequestID:  "req_rider_invite",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"type":       RiderAccountStationManager,
				"station_id": "station_9",
				"token":      "caller-token-must-not-persist",
			},
			CreatedAt: time.Date(2026, 5, 24, 9, 10, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if invite.Token == "" || audit.TargetID != invite.Token || audit.TargetType != "rider_invite" {
		t.Fatalf("expected rider invite audit to target generated token, invite=%+v audit=%+v", invite, audit)
	}
	if audit.Payload["type"] != RiderAccountRider || audit.Payload["station_id"] != "station_1" || audit.Payload["expires_at"] != invite.ExpiresAt.Format(time.RFC3339Nano) || audit.Payload["token"] != nil {
		t.Fatalf("expected server-generated rider invite audit payload, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified rider invite audit proof, got %+v", audit)
	}
	logs, err := store.AuditLogs(AuditLogsRequest{TargetType: "rider_invite", TargetID: invite.Token, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Action != "admin.rider_invite.created" || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable rider invite audit log, got %+v", logs)
	}
}

func TestMerchantProductManagementScopesToOwnedShop(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	products, err := store.MerchantProducts("merchant_1", "shop_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("expected seeded merchant products, got %+v", products)
	}

	created, err := store.UpsertMerchantProduct(UpsertMerchantProductRequest{
		MerchantID:     "merchant_1",
		ShopID:         "shop_1",
		Name:           "轻食鸡胸饭",
		ImageURL:       "/assets/mock/chicken-rice.jpg",
		Description:    "鸡胸肉、糙米、蔬菜。",
		IngredientList: []string{"鸡胸肉", "糙米", "蔬菜", "鸡胸肉"},
		PriceFen:       2299,
		StockCount:     20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Status != ProductStatusActive || len(created.IngredientList) != 3 {
		t.Fatalf("expected normalized created product, got %+v", created)
	}

	publicProducts, err := store.ShopProducts("shop_1")
	if err != nil {
		t.Fatal(err)
	}
	foundPublic := false
	for _, product := range publicProducts {
		if product.ID == created.ID {
			foundPublic = true
			break
		}
	}
	if !foundPublic {
		t.Fatalf("expected active merchant product to be public, got %+v", publicProducts)
	}

	soldOut, err := store.SetMerchantProductStatus(SetMerchantProductStatusRequest{MerchantID: "merchant_1", ProductID: created.ID, Status: ProductStatusSoldOut})
	if err != nil {
		t.Fatal(err)
	}
	if soldOut.Status != ProductStatusSoldOut {
		t.Fatalf("expected product sold out, got %+v", soldOut)
	}
	if _, err := store.SetMerchantProductStatus(SetMerchantProductStatusRequest{MerchantID: "merchant_2", ProductID: created.ID, Status: ProductStatusRemoved}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected wrong merchant to be rejected, got %v", err)
	}
	publicProducts, err = store.ShopProducts("shop_1")
	if err != nil {
		t.Fatal(err)
	}
	for _, product := range publicProducts {
		if product.ID == created.ID {
			t.Fatalf("expected sold-out product to be hidden publicly, got %+v", publicProducts)
		}
	}
}

func TestGroupbuyOrderPaymentIssuesAndRedeemsVoucher(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	deals, err := store.ShopGroupbuyDeals("shop_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(deals) != 1 || deals[0].ID != "deal_two_person_set" {
		t.Fatalf("expected seeded groupbuy deal, got %+v", deals)
	}
	order, err := store.CreateGroupbuyOrder(CreateGroupbuyOrderRequest{
		UserID:   "user_1",
		ShopID:   "shop_1",
		DealID:   "deal_two_person_set",
		Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.Type != OrderTypeGroupbuy || order.Status != StatusPendingPayment || order.AmountFen != 3999 {
		t.Fatalf("expected pending groupbuy order, got %+v", order)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 3999, IdempotencyKey: "credit_groupbuy"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_groupbuy"}); err != nil || paidOrder.Status != StatusVoucherIssued {
		t.Fatalf("expected groupbuy payment to issue voucher, order=%+v err=%v", paidOrder, err)
	}
	vouchers, err := store.UserGroupbuyVouchers("user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(vouchers) != 1 || vouchers[0].Status != GroupbuyVoucherStatusIssued || vouchers[0].QRPayload == "" {
		t.Fatalf("expected issued voucher with qr payload, got %+v", vouchers)
	}
	if _, _, err := store.RedeemGroupbuyVoucher(RedeemGroupbuyVoucherRequest{MerchantID: "merchant_2", QRPayload: vouchers[0].QRPayload, Method: GroupbuyRedemptionMethodQR}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected wrong merchant to be rejected, got %v", err)
	}
	redeemed, completedOrder, err := store.RedeemGroupbuyVoucher(RedeemGroupbuyVoucherRequest{MerchantID: "merchant_1", QRPayload: vouchers[0].QRPayload, Method: GroupbuyRedemptionMethodQR})
	if err != nil {
		t.Fatal(err)
	}
	if redeemed.Status != GroupbuyVoucherRedeemed || redeemed.RedemptionMethod != GroupbuyRedemptionMethodQR || completedOrder.Status != StatusCompleted {
		t.Fatalf("expected voucher redemption to complete order, voucher=%+v order=%+v", redeemed, completedOrder)
	}
	if _, _, err := store.RedeemGroupbuyVoucher(RedeemGroupbuyVoucherRequest{MerchantID: "merchant_1", VoucherCode: vouchers[0].VoucherCode, Method: GroupbuyRedemptionMethodQR}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected duplicate redemption to be rejected, got %v", err)
	}
}

func TestBalancePaymentRequiresPaymentPassword(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 2000, IdempotencyKey: "credit_pwd"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_without_pwd"}); !errors.Is(err, ErrPaymentPassword) {
		t.Fatalf("expected payment password error before password is set, got %v", err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "abc123"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for non-digit password, got %v", err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "000000", IdempotencyKey: "pay_wrong_pwd"}); !errors.Is(err, ErrPaymentPassword) {
		t.Fatalf("expected payment password error for wrong password, got %v", err)
	}
}

func TestShopAddressCartCheckoutPaymentAndGrabFlow(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	shops := store.Shops()
	if len(shops) == 0 || shops[0].ID != "shop_1" {
		t.Fatalf("expected seeded shop, got %+v", shops)
	}
	if shops[0].Status != ShopStatusActive {
		t.Fatalf("expected seeded shop to be active, got %+v", shops[0])
	}
	products, err := store.ShopProducts("shop_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(products) == 0 || products[0].ID != "prod_beef_rice" {
		t.Fatalf("expected seeded products, got %+v", products)
	}

	lat := 39.99
	lng := 116.48
	address, err := store.SaveAddress(UserAddress{
		UserID:       "user_1",
		ContactName:  "张三",
		ContactPhone: "13800000000",
		City:         "北京",
		Detail:       "望京SOHO",
		Latitude:     &lat,
		Longitude:    &lng,
		Tag:          AddressTagHome,
		IsDefault:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if address.ID == "" {
		t.Fatal("expected generated address id")
	}

	summary, err := store.UpsertCartItem(UpsertCartItemRequest{
		UserID:    "user_1",
		ShopID:    "shop_1",
		ProductID: "prod_beef_rice",
		Quantity:  2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.PayableFen != 5598 {
		t.Fatalf("expected payable 5598, got %+v", summary)
	}

	order, checkoutSummary, err := store.CheckoutCart(CheckoutCartRequest{
		UserID:    "user_1",
		ShopID:    "shop_1",
		AddressID: address.ID,
		Options: OrderOptions{
			Remark:         "少放辣",
			TablewareCount: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != StatusPendingPayment || order.AmountFen != checkoutSummary.PayableFen || len(order.Items) != 1 {
		t.Fatalf("unexpected checkout order: order=%+v summary=%+v", order, checkoutSummary)
	}
	emptySummary, err := store.CartSummary("user_1", "shop_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(emptySummary.Items) != 0 {
		t.Fatalf("expected cart to be cleared after checkout, got %+v", emptySummary)
	}

	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: order.AmountFen, IdempotencyKey: "credit_checkout"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_checkout"}); err != nil || paidOrder.Status != StatusMerchantPending {
		t.Fatalf("expected checkout order to enter merchant pending, order=%+v err=%v", paidOrder, err)
	}
	if _, err := store.GrabOrder(order.ID, "rider_1"); !errors.Is(err, ErrOrderAlreadyAssigned) {
		t.Fatalf("expected rider grab to be blocked before merchant ready, got %v", err)
	}
	merchantOrders, err := store.MerchantOrders("merchant_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(merchantOrders) != 1 || merchantOrders[0].ID != order.ID {
		t.Fatalf("expected merchant order list to include checkout order, got %+v", merchantOrders)
	}
	acceptedOrder, err := store.MerchantAcceptOrder(order.ID, "merchant_1")
	if err != nil || acceptedOrder.Status != StatusPreparing {
		t.Fatalf("expected merchant to accept checkout order, order=%+v err=%v", acceptedOrder, err)
	}
	if _, err := store.MerchantAcceptOrder(order.ID, "merchant_2"); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected wrong merchant to be rejected, got %v", err)
	}
	readyOrder, err := store.MerchantMarkOrderReady(order.ID, "merchant_1")
	if err != nil || readyOrder.Status != StatusDispatching {
		t.Fatalf("expected merchant ready order to enter dispatching, order=%+v err=%v", readyOrder, err)
	}
	if grabbed, err := store.GrabOrder(order.ID, "rider_1"); err != nil || grabbed.Status != StatusRiderAssigned {
		t.Fatalf("expected rider to grab checkout order, order=%+v err=%v", grabbed, err)
	}
	orders, err := store.UserOrders("user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 1 || orders[0].ID != order.ID {
		t.Fatalf("expected user order list to include checkout order, got %+v", orders)
	}
	orderDetail, err := store.OrderByID(order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if orderDetail.ID != order.ID || len(orderDetail.Events) < 2 {
		t.Fatalf("expected order detail with events, got %+v", orderDetail)
	}
}

func TestConcurrentGrabAllowsExactlyOneWinner(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.riders["rider"] = &RiderAccount{
		ID:            "rider",
		StationID:     "station_1",
		Type:          RiderAccountRider,
		Status:        "active",
		Online:        true,
		DepositStatus: DepositStatusPaid,
		Capacity:      1,
	}
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1, IdempotencyKey: "credit_grab"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_grab"}); err != nil {
		t.Fatal(err)
	}

	var successes atomic.Int64
	var alreadyAssigned atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.GrabOrder(order.ID, "rider")
			if err == nil {
				successes.Add(1)
				return
			}
			if errors.Is(err, ErrOrderAlreadyAssigned) {
				alreadyAssigned.Add(1)
				return
			}
			t.Errorf("unexpected grab error: %v", err)
		}(i)
	}
	wg.Wait()

	if successes.Load() != 1 {
		t.Fatalf("expected exactly one successful grab, got %d", successes.Load())
	}
	if alreadyAssigned.Load() != 9999 {
		t.Fatalf("expected 9999 already assigned responses, got %d", alreadyAssigned.Load())
	}
}

func TestAutoAssignAndRejectDispatchesNextOnlineRider(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_assign"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_assign"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidOrder.ID, Now: paidOrder.CreatedAt.Add(9 * time.Minute)}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected auto assign before ten minutes to be rejected, got %v", err)
	}
	assignedOrder, decision, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidOrder.ID, Now: paidOrder.CreatedAt.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if assignedOrder.Status != StatusRiderAssigned || assignedOrder.RiderID != "rider_1" || decision.Mode != DispatchModeAutoAssign {
		t.Fatalf("expected rider_1 auto assignment, order=%+v decision=%+v", assignedOrder, decision)
	}
	if decision.StationID != "station_1" {
		t.Fatalf("expected station_1 dispatch decision, got %+v", decision)
	}
	reassignedOrder, nextDecision, err := store.RejectRiderAssignment(RejectRiderAssignmentRequest{OrderID: paidOrder.ID, RiderID: "rider_1"})
	if err != nil {
		t.Fatal(err)
	}
	if reassignedOrder.Status != StatusRiderAssigned || reassignedOrder.RiderID != "rider_2" || nextDecision.CandidateRiderID != "rider_2" {
		t.Fatalf("expected rejection to assign rider_2, order=%+v decision=%+v", reassignedOrder, nextDecision)
	}
	events, err := store.DispatchEvents(paidOrder.ID, "station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].Type != "dispatch.auto_assign" || events[1].Type != "dispatch.rejected" || events[2].Type != "dispatch.auto_assign" {
		t.Fatalf("expected auto/reject/auto dispatch events, got %+v", events)
	}
}

func TestTimeoutReassignSkipsTimedOutRider(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_timeout"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_timeout"})
	if err != nil {
		t.Fatal(err)
	}
	assignNow := paidOrder.CreatedAt.Add(10 * time.Minute)
	assignedOrder, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidOrder.ID, Now: assignNow})
	if err != nil {
		t.Fatal(err)
	}
	if assignedOrder.RiderID != "rider_1" {
		t.Fatalf("expected rider_1 initial assignment, got %+v", assignedOrder)
	}
	if _, _, err := store.TimeoutReassignOrder(TimeoutReassignOrderRequest{OrderID: paidOrder.ID, Now: assignNow.Add(59 * time.Second), TimeoutSeconds: 60}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected early timeout reassign to be rejected, got %v", err)
	}
	reassignedOrder, decision, err := store.TimeoutReassignOrder(TimeoutReassignOrderRequest{OrderID: paidOrder.ID, Now: assignNow.Add(60 * time.Second), TimeoutSeconds: 60})
	if err != nil {
		t.Fatal(err)
	}
	if reassignedOrder.Status != StatusRiderAssigned || reassignedOrder.RiderID != "rider_2" || decision.CandidateRiderID != "rider_2" || decision.Reason != "assignment_timeout" {
		t.Fatalf("expected timeout to reassign rider_2, order=%+v decision=%+v", reassignedOrder, decision)
	}
	events, err := store.DispatchEvents(paidOrder.ID, "station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].Type != "dispatch.auto_assign" || events[1].Type != "dispatch.timeout" || events[2].Type != "dispatch.auto_assign" {
		t.Fatalf("expected auto/timeout/auto dispatch events, got %+v", events)
	}
	if len(events[1].RejectedRiderIDs) != 1 || events[1].RejectedRiderIDs[0] != "rider_1" {
		t.Fatalf("expected timeout event to skip rider_1, got %+v", events[1])
	}
}

func TestRejectDispatchAfterDailyFixedOrderCountIsPenaltyFree(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	store.stationTaskConfigs["station_1"].DailyFixedOrderCount = 1
	store.orders["completed_today"] = &Order{
		ID:        "completed_today",
		UserID:    "user_1",
		Type:      OrderTypeTakeout,
		Status:    StatusCompleted,
		AmountFen: 1200,
		RiderID:   "rider_1",
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-90 * time.Minute),
	}
	store.orders["assigned_after_quota"] = &Order{
		ID:        "assigned_after_quota",
		UserID:    "user_2",
		Type:      OrderTypeTakeout,
		Status:    StatusRiderAssigned,
		AmountFen: 1600,
		RiderID:   "rider_1",
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	reassignedOrder, decision, err := store.RejectRiderAssignment(RejectRiderAssignmentRequest{OrderID: "assigned_after_quota", RiderID: "rider_1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if reassignedOrder.RiderID != "rider_2" {
		t.Fatalf("expected next rider after quota-free decline, got %+v", reassignedOrder)
	}
	if !decision.CanDeclineWithoutPenalty || decision.DailyCompletedOrderCount != 1 || decision.DailyFixedOrderCount != 1 || decision.Reason != "after_fixed_order_count" {
		t.Fatalf("expected fixed-order quota decline exemption, got %+v", decision)
	}
}

func TestRiderPickupAndDeliveredCompletesOrder(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.orders["assigned_delivery"] = &Order{
		ID:        "assigned_delivery",
		UserID:    "user_1",
		Type:      OrderTypeTakeout,
		Status:    StatusRiderAssigned,
		AmountFen: 1800,
		RiderID:   "rider_1",
		CreatedAt: time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC),
	}

	pickedUp, err := store.RiderMarkOrderPickedUp("assigned_delivery", "rider_1")
	if err != nil {
		t.Fatal(err)
	}
	if pickedUp.Status != StatusPickedUp {
		t.Fatalf("expected picked up status, got %+v", pickedUp)
	}
	completed, err := store.RiderMarkOrderDelivered("assigned_delivery", "rider_1")
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != StatusCompleted || completed.Events[len(completed.Events)-1].Type != "delivery.completed" {
		t.Fatalf("expected completed delivery order, got %+v", completed)
	}
	if _, err := store.RiderMarkOrderDelivered("assigned_delivery", "rider_1"); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected duplicate delivered to be rejected, got %v", err)
	}
}

func TestCompensateOrderStateRepairsPaymentAndDispatchDrift(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_compensate"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_compensate"})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[paidOrder.ID].Status = StatusPendingPayment
	store.orders[paidOrder.ID].PaymentMethod = ""

	repaired, err := store.CompensateOrderState(CompensateOrderStateRequest{OrderID: paidOrder.ID, ActorID: "admin_1"})
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.Changed || repaired.PreviousStatus != StatusPendingPayment || repaired.ExpectedStatus != StatusDispatching || repaired.Order.Status != StatusDispatching || repaired.Order.PaymentMethod != PaymentBalance {
		t.Fatalf("expected payment drift compensation to restore dispatching balance order, got %+v", repaired)
	}
	if repaired.Order.Events[len(repaired.Order.Events)-1].Type != "order.state.compensated" {
		t.Fatalf("expected compensation audit event, got %+v", repaired.Order.Events)
	}

	assignNow := paidOrder.CreatedAt.Add(10 * time.Minute)
	assignedOrder, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidOrder.ID, Now: assignNow})
	if err != nil {
		t.Fatal(err)
	}
	if assignedOrder.RiderID != "rider_1" {
		t.Fatalf("expected auto assignment to rider_1, got %+v", assignedOrder)
	}
	store.orders[paidOrder.ID].Status = StatusDispatching
	store.orders[paidOrder.ID].RiderID = ""

	repaired, err = store.CompensateOrderState(CompensateOrderStateRequest{OrderID: paidOrder.ID, ActorID: "admin_1", Now: assignNow.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.Changed || repaired.ExpectedStatus != StatusRiderAssigned || repaired.ExpectedRiderID != "rider_1" || repaired.Order.Status != StatusRiderAssigned || repaired.Order.RiderID != "rider_1" {
		t.Fatalf("expected dispatch event compensation to restore rider assignment, got %+v", repaired)
	}
	rechecked, err := store.CompensateOrderState(CompensateOrderStateRequest{OrderID: paidOrder.ID, ActorID: "admin_1"})
	if err != nil {
		t.Fatal(err)
	}
	if rechecked.Changed {
		t.Fatalf("expected second compensation to be idempotent, got %+v", rechecked)
	}
}

func TestCompensateOrderStateWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("order-state-compensation-audit-secret")
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_atomic_compensate", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: order.UserID, AmountFen: 1200, IdempotencyKey: "credit_compensate_atomic_audit"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: order.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: order.UserID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_compensate_atomic_audit"})
	if err != nil {
		t.Fatal(err)
	}
	store.orders[paidOrder.ID].Status = StatusPendingPayment
	store.orders[paidOrder.ID].PaymentMethod = ""
	createdAt := time.Date(2026, 5, 23, 13, 20, 0, 222333444, time.UTC)

	repaired, audit, err := store.CompensateOrderStateWithAudit(
		CompensateOrderStateRequest{
			OrderID: paidOrder.ID,
			ActorID: "admin_1",
			Now:     createdAt.Add(time.Minute),
		},
		RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     "admin.order_state.compensated",
			TargetType: "order",
			TargetID:   paidOrder.ID,
			RequestID:  "req_order_state_compensate_atomic",
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"changed":         false,
				"previous_status": StatusCompleted,
				"expected_status": StatusRefunded,
				"token":           "must-not-persist",
			},
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.Changed || repaired.PreviousStatus != StatusPendingPayment || repaired.ExpectedStatus != StatusDispatching || repaired.Order.Status != StatusDispatching || repaired.Order.PaymentMethod != PaymentBalance {
		t.Fatalf("expected audited order-state compensation to restore dispatching balance order, got %+v", repaired)
	}
	if audit.Action != "admin.order_state.compensated" || audit.TargetType != "order" || audit.TargetID != paidOrder.ID {
		t.Fatalf("expected order-state compensation audit target, got %+v", audit)
	}
	if audit.Payload["changed"] != true ||
		audit.Payload["previous_status"] != StatusPendingPayment ||
		audit.Payload["expected_status"] != StatusDispatching ||
		audit.Payload["compensation_type"] != "order_state_replay" ||
		audit.Payload["evidence_count"] != len(repaired.Evidence) ||
		audit.Payload["token"] != nil {
		t.Fatalf("expected compensation audit payload to be server-generated and sanitized, got %+v", audit.Payload)
	}
	if audit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || audit.IntegrityHash == "" || !audit.IntegrityVerified {
		t.Fatalf("expected verified HMAC integrity proof, got %+v", audit)
	}
	if audit.CreatedAt.Nanosecond() != 222333000 {
		t.Fatalf("expected audit timestamp normalized to PostgreSQL precision, got %s", audit.CreatedAt.Format(time.RFC3339Nano))
	}

	logs, err := store.AuditLogs(AuditLogsRequest{Action: "admin.order_state.compensated", TargetType: "order", TargetID: paidOrder.ID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != audit.ID || logs[0].Payload["changed"] != true || logs[0].Payload["previous_status"] != StatusPendingPayment || !logs[0].IntegrityVerified {
		t.Fatalf("expected queryable verified compensation audit log from atomic write, got %+v", logs)
	}
}

func TestCompensateOrderStateProtectsCompletedOrders(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	store.orders["completed_order"] = &Order{
		ID:        "completed_order",
		UserID:    "user_1",
		Type:      OrderTypeTakeout,
		Status:    StatusCompleted,
		AmountFen: 1600,
		RiderID:   "rider_1",
		CreatedAt: now.Add(-30 * time.Minute),
		UpdatedAt: now,
		Events: []OrderEvent{{
			Type:      "delivery.completed",
			ActorID:   "rider_1",
			Message:   "骑手已送达，订单完成",
			CreatedAt: now,
		}},
	}
	store.recordDispatchEventLocked(store.orders["completed_order"], &DispatchDecision{
		OrderID:          "completed_order",
		Mode:             DispatchModeAutoAssign,
		StationID:        "station_1",
		CandidateRiderID: "rider_2",
		IdempotencyKey:   "dispatch:completed_order:late",
	}, "dispatch.timeout", "rider_1", "system", "assignment_timeout", now.Add(time.Minute))

	result, err := store.CompensateOrderState(CompensateOrderStateRequest{OrderID: "completed_order", ActorID: "admin_1", Now: now.Add(2 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || result.Order.Status != StatusCompleted || result.Order.RiderID != "rider_1" {
		t.Fatalf("expected completed order to be protected from dispatch regression, got %+v", result)
	}
}

func TestDepositOperationsGateRiderAndMerchant(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.riders["rider_unpaid"] = &RiderAccount{
		ID:            "rider_unpaid",
		StationID:     "station_1",
		Type:          RiderAccountRider,
		Status:        "active",
		Online:        true,
		DepositStatus: DepositStatusUnpaid,
		Capacity:      1,
	}
	store.merchants["merchant_unpaid"] = &MerchantAccount{
		ID:            "merchant_unpaid",
		Type:          MerchantAccountStandard,
		DisplayName:   "未缴保证金商户",
		Status:        ShopStatusActive,
		DepositStatus: DepositStatusUnpaid,
	}

	deposit, err := store.DepositAccount("rider", "rider_unpaid")
	if err != nil {
		t.Fatal(err)
	}
	if deposit.Status != DepositStatusUnpaid || deposit.AmountFen != RiderDepositAmountFen {
		t.Fatalf("expected default unpaid rider deposit, got %+v", deposit)
	}
	if _, err := store.PayDeposit(PayDepositRequest{SubjectType: "rider", SubjectID: "rider_unpaid", AmountFen: RiderDepositAmountFen - 1}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected insufficient rider deposit to be rejected, got %v", err)
	}
	paidRiderDeposit, err := store.PayDeposit(PayDepositRequest{SubjectType: "rider", SubjectID: "rider_unpaid", AmountFen: RiderDepositAmountFen})
	if err != nil {
		t.Fatal(err)
	}
	if paidRiderDeposit.Status != DepositStatusPaid || store.riders["rider_unpaid"].DepositStatus != DepositStatusPaid {
		t.Fatalf("expected paid rider deposit synced to rider, deposit=%+v rider=%+v", paidRiderDeposit, store.riders["rider_unpaid"])
	}
	store.orders["rider_unpaid_completed"] = &Order{ID: "rider_unpaid_completed", RiderID: "rider_unpaid", Status: StatusCompleted, UpdatedAt: time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)}
	refundDeposit, rider, err := store.RequestRiderDepositRefund(RiderDepositRefundRequest{RiderID: "rider_unpaid", ResignationSubmittedAt: time.Date(2026, 5, 21, 13, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if refundDeposit.Status != DepositStatusRefundPending || rider.DepositStatus != DepositStatusRefundPending || refundDeposit.LastOrderCompletedAt.IsZero() {
		t.Fatalf("expected rider refund pending with last completed order, deposit=%+v rider=%+v", refundDeposit, rider)
	}

	store.riders["rider_exempt"] = &RiderAccount{ID: "rider_exempt", StationID: "station_1", Type: RiderAccountRider, Status: "active", Online: true, DepositStatus: DepositStatusUnpaid, Capacity: 1}
	exemptDeposit, exemptRider, err := store.ApproveRiderWechatExemption(RiderWechatExemptionRequest{RiderID: "rider_exempt", ApplicationID: "wx_exempt_1"})
	if err != nil {
		t.Fatal(err)
	}
	if exemptDeposit.Status != DepositStatusWechatExemptApproved || exemptRider.DepositStatus != DepositStatusWechatExemptApproved {
		t.Fatalf("expected wechat exempt rider deposit, deposit=%+v rider=%+v", exemptDeposit, exemptRider)
	}

	merchantDeposit, err := store.PayDeposit(PayDepositRequest{SubjectType: "merchant", SubjectID: "merchant_unpaid", AmountFen: MerchantDepositAmountFen})
	if err != nil {
		t.Fatal(err)
	}
	if merchantDeposit.Status != DepositStatusPaid || store.merchants["merchant_unpaid"].DepositStatus != DepositStatusPaid {
		t.Fatalf("expected paid merchant deposit synced to merchant, deposit=%+v merchant=%+v", merchantDeposit, store.merchants["merchant_unpaid"])
	}
}

func TestStationManagerManualAssignUsesStationScope(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.riders["rider_3"] = &RiderAccount{
		ID:                   "rider_3",
		StationID:            "station_2",
		Type:                 RiderAccountRider,
		Status:               "active",
		Online:               true,
		DepositStatus:        DepositStatusPaid,
		Capacity:             2,
		DispatchPriority:     RiderDispatchPriority(RiderLevelA),
		AverageAcceptSeconds: 16,
		AverageDailyOrders:   30,
		CompletionRate:       0.97,
		DistanceMeters:       450,
	}

	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1600})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1600, IdempotencyKey: "credit_manual_assign"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: "user_1", OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_manual_assign"})
	if err != nil {
		t.Fatal(err)
	}

	riders, err := store.StationRiders("station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(riders) != 2 {
		t.Fatalf("expected station_1 riders only, got %+v", riders)
	}
	orders, err := store.StationOrders("station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 1 || orders[0].ID != paidOrder.ID {
		t.Fatalf("expected station dispatching order, got %+v", orders)
	}
	if _, _, err := store.ManualAssignOrder(ManualAssignOrderRequest{OrderID: paidOrder.ID, RiderID: "rider_3", StationManagerID: "station_manager_1"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-station manual assignment to be hidden, got %v", err)
	}
	assignedOrder, decision, err := store.ManualAssignOrder(ManualAssignOrderRequest{OrderID: paidOrder.ID, RiderID: "rider_2", StationManagerID: "station_manager_1"})
	if err != nil {
		t.Fatal(err)
	}
	if assignedOrder.Status != StatusRiderAssigned || assignedOrder.RiderID != "rider_2" || decision.Mode != DispatchModeManualAssign {
		t.Fatalf("expected manual assignment to rider_2, order=%+v decision=%+v", assignedOrder, decision)
	}
	config, err := store.StationTaskConfig("station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if config.DailyTaskDurationMinutes != 8*60 || config.DailyFixedOrderCount != 30 {
		t.Fatalf("expected seeded station task config, got %+v", config)
	}
	savedConfig, err := store.SaveStationTaskConfig(SaveStationTaskConfigRequest{StationManagerID: "station_manager_1", DailyTaskDurationMinutes: 7 * 60, DailyFixedOrderCount: 28})
	if err != nil {
		t.Fatal(err)
	}
	if savedConfig.DailyTaskDurationMinutes != 7*60 || savedConfig.DailyFixedOrderCount != 28 {
		t.Fatalf("expected saved station task config, got %+v", savedConfig)
	}
	if _, err := store.SaveStationTaskConfig(SaveStationTaskConfigRequest{StationManagerID: "station_manager_1", DailyTaskDurationMinutes: 25 * 60, DailyFixedOrderCount: 28}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid duration to be rejected, got %v", err)
	}
	performance, err := store.StationRiderPerformance("station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(performance) != 2 || performance[0].RiderID != "rider_2" || performance[0].Level == "" || performance[0].DispatchPriority == 0 {
		t.Fatalf("expected station rider performance ranked for station_1, got %+v", performance)
	}
}

func TestStationServiceAreaScopesOrdersAndDispatchCandidates(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.riders["station_manager_2"] = &RiderAccount{
		ID:        "station_manager_2",
		StationID: "station_2",
		Type:      RiderAccountStationManager,
		Status:    "active",
		Online:    true,
	}
	store.riders["rider_3"] = &RiderAccount{
		ID:                   "rider_3",
		StationID:            "station_2",
		Type:                 RiderAccountRider,
		Status:               "active",
		Online:               true,
		DepositStatus:        DepositStatusPaid,
		Capacity:             2,
		DispatchPriority:     RiderDispatchPriority(RiderLevelB),
		AverageAcceptSeconds: 10,
		AverageDailyOrders:   30,
		CompletionRate:       0.98,
		DistanceMeters:       120,
	}
	store.shops["shop_2"] = &Shop{
		ID:             "shop_2",
		MerchantID:     "merchant_1",
		StationID:      "station_2",
		Name:           "东城热炒",
		Category:       "restaurant",
		AccountType:    MerchantAccountStandard,
		Status:         ShopStatusActive,
		Capabilities:   []string{ShopCapabilityTakeout},
		Qualifications: []string{QualificationBusinessLicense, QualificationHealthCertificate},
	}
	store.stationServiceAreas["station_2"] = &StationServiceArea{StationID: "station_2", ShopIDs: []string{"shop_2"}}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	autoOrder := &Order{
		ID:        "order_station_2_auto",
		UserID:    "user_2",
		ShopID:    "shop_2",
		Type:      OrderTypeTakeout,
		Status:    StatusDispatching,
		AmountFen: 2200,
		CreatedAt: now.Add(-11 * time.Minute),
		UpdatedAt: now.Add(-11 * time.Minute),
	}
	manualOrder := &Order{
		ID:        "order_station_2_manual",
		UserID:    "user_3",
		ShopID:    "shop_2",
		Type:      OrderTypeTakeout,
		Status:    StatusDispatching,
		AmountFen: 2300,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.orders[autoOrder.ID] = autoOrder
	store.orders[manualOrder.ID] = manualOrder

	station1Orders, err := store.StationOrders("station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	for _, order := range station1Orders {
		if order.ID == autoOrder.ID || order.ID == manualOrder.ID {
			t.Fatalf("expected station_1 manager not to see station_2 orders, got %+v", station1Orders)
		}
	}
	station2Orders, err := store.StationOrders("station_manager_2")
	if err != nil {
		t.Fatal(err)
	}
	if len(station2Orders) != 2 {
		t.Fatalf("expected station_2 manager to see both station_2 orders, got %+v", station2Orders)
	}

	assigned, decision, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: autoOrder.ID, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if assigned.RiderID != "rider_3" || decision.StationID != "station_2" || decision.CandidateRiderID != "rider_3" {
		t.Fatalf("expected station_2 auto dispatch to rider_3, order=%+v decision=%+v", assigned, decision)
	}
	if _, _, err := store.ManualAssignOrder(ManualAssignOrderRequest{OrderID: manualOrder.ID, RiderID: "rider_3", StationManagerID: "station_manager_1"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected station_1 manager manual assignment to be hidden, got %v", err)
	}
	manualAssigned, manualDecision, err := store.ManualAssignOrder(ManualAssignOrderRequest{OrderID: manualOrder.ID, RiderID: "rider_3", StationManagerID: "station_manager_2"})
	if err != nil {
		t.Fatal(err)
	}
	if manualAssigned.RiderID != "rider_3" || manualDecision.StationID != "station_2" || manualDecision.Mode != DispatchModeManualAssign {
		t.Fatalf("expected station_2 manual assignment, order=%+v decision=%+v", manualAssigned, manualDecision)
	}
	if _, err := store.DispatchEvents(autoOrder.ID, "station_manager_1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected station_1 manager not to read station_2 dispatch events, got %v", err)
	}
}

func TestFreeDispatchCancelCanOnlyBeConsumedOncePerDay(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	now := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)

	allowed, usedOn, err := store.ConsumeFreeDispatchCancel("rider_1", now)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed || usedOn != "2026-05-21" {
		t.Fatalf("expected first cancel allowed on 2026-05-21, got allowed=%v usedOn=%s", allowed, usedOn)
	}

	allowed, _, err = store.ConsumeFreeDispatchCancel("rider_1", now)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected second same-day cancel to be non-free")
	}
}

func TestPaymentSuccessCreatesOrderOutboxEvents(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	balanceOrder := mustPaidDispatchOrder(t, store, "outbox_balance")

	wechatOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_2", Type: OrderTypeTakeout, AmountFen: 900})
	if err != nil {
		t.Fatal(err)
	}
	prepay, _, err := store.CreateWechatPrepay(WechatPrepayRequest{UserID: "user_2", OrderID: wechatOrder.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{OutTradeNo: prepay.OutTradeNo, TransactionID: "wx_outbox_tx", AmountFen: 900}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{OutTradeNo: prepay.OutTradeNo, TransactionID: "wx_outbox_tx", AmountFen: 900}); err != nil {
		t.Fatal(err)
	}

	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byOrder := map[string]OutboxEvent{}
	for _, event := range events {
		byOrder[event.AggregateID] = event
	}
	if len(byOrder) != 2 {
		t.Fatalf("expected one order.paid event per paid order, got %+v", events)
	}
	if event := byOrder[balanceOrder.ID]; event.EventType != "order.payment.success" || event.Payload["payment_method"] != PaymentBalance || event.Status != OutboxStatusPending {
		t.Fatalf("expected balance payment outbox event, got %+v", event)
	}
	if event := byOrder[wechatOrder.ID]; event.EventType != "order.payment.success" || event.Payload["payment_method"] != PaymentWechat || event.Status != OutboxStatusPending {
		t.Fatalf("expected wechat payment outbox event, got %+v", event)
	}
}

func TestDispatchEventsCreateOutboxEvents(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	rejectOrder := mustPaidDispatchOrder(t, store, "outbox_reject")
	assignNow := rejectOrder.CreatedAt.Add(10 * time.Minute)
	if _, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: rejectOrder.ID, Now: assignNow}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.RejectRiderAssignment(RejectRiderAssignmentRequest{OrderID: rejectOrder.ID, RiderID: "rider_1", Now: assignNow.Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}

	timeoutOrder := mustPaidDispatchOrder(t, store, "outbox_timeout")
	timeoutAssignNow := timeoutOrder.CreatedAt.Add(10 * time.Minute)
	if _, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: timeoutOrder.ID, Now: timeoutAssignNow}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.TimeoutReassignOrder(TimeoutReassignOrderRequest{OrderID: timeoutOrder.ID, Now: timeoutAssignNow.Add(DispatchAssignmentTimeoutSeconds * time.Second), TimeoutSeconds: DispatchAssignmentTimeoutSeconds}); err != nil {
		t.Fatal(err)
	}

	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "dispatch.assigned", Limit: 20, Now: timeoutAssignNow.Add(DispatchAssignmentTimeoutSeconds*time.Second + time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	assignedByOrder := map[string]int{}
	for _, event := range events {
		assignedByOrder[event.AggregateID]++
		if event.AggregateType != "dispatch" || event.Payload["station_id"] == "" {
			t.Fatalf("expected dispatch assigned outbox payload, got %+v", event)
		}
	}
	if assignedByOrder[rejectOrder.ID] != 2 || assignedByOrder[timeoutOrder.ID] != 2 {
		t.Fatalf("expected initial and reassigned dispatch outbox events, got %+v", assignedByOrder)
	}
	statusEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "dispatch.status_changed", Limit: 20, Now: timeoutAssignNow.Add(DispatchAssignmentTimeoutSeconds*time.Second + time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	timeoutEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "dispatch.timeout", Limit: 20, Now: timeoutAssignNow.Add(DispatchAssignmentTimeoutSeconds*time.Second + time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(statusEvents) == 0 || statusEvents[0].EventType != "dispatch.rejected" {
		t.Fatalf("expected rejected dispatch status outbox event, got %+v", statusEvents)
	}
	if len(timeoutEvents) != 1 || timeoutEvents[0].EventType != "dispatch.timeout" {
		t.Fatalf("expected timeout dispatch outbox event, got %+v", timeoutEvents)
	}
}

func TestOutboxFailedBackoffAndPublishedAck(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_retry")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	failed, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "kafka unavailable", RetryAfterSeconds: 120, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != OutboxStatusFailed || failed.Attempts != 1 || failed.LastError != "kafka unavailable" {
		t.Fatalf("expected failed outbox event with attempt count, got %+v", failed)
	}
	hidden, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Now: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(hidden) != 0 {
		t.Fatalf("expected failed event hidden before retry time, got %+v", hidden)
	}
	readyAgain, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Now: now.Add(121 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(readyAgain) != 1 || readyAgain[0].ID != failed.ID {
		t.Fatalf("expected failed event to become pending-ready after backoff, got %+v", readyAgain)
	}
	published, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: failed.ID, PublishedAt: now.Add(2 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != OutboxStatusPublished || published.PublishedAt.IsZero() || published.LastError != "" {
		t.Fatalf("expected published outbox ack, got %+v", published)
	}
	pending, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Now: now.Add(3 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected published event to leave pending relay query, got %+v", pending)
	}
}

func TestClaimOutboxEventsLeasesReadyEvents(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_claim")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := events[0].CreatedAt.Add(time.Second)
	firstClaim, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-a", LeaseSeconds: 30, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if firstClaim.Topic != "order.paid" || firstClaim.Limit != 1 || firstClaim.LeaseOwner != "relay-a" || firstClaim.Claimed != 1 || len(firstClaim.Events) != 1 {
		t.Fatalf("expected relay-a to claim one event, got %+v", firstClaim)
	}
	if firstClaim.Events[0].ID != events[0].ID || firstClaim.Events[0].LeaseOwner != "relay-a" || !firstClaim.Events[0].LeaseExpiresAt.Equal(now.Add(30*time.Second)) {
		t.Fatalf("expected claimed event to carry relay-a lease, got %+v", firstClaim.Events[0])
	}

	secondClaim, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-b", LeaseSeconds: 30, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if secondClaim.Claimed != 0 || len(secondClaim.Events) != 0 {
		t.Fatalf("expected active lease to hide event from relay-b, got %+v", secondClaim)
	}
	pending, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected active lease to hide event from pending query, got %+v", pending)
	}
	stats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 1 || stats.Pending != 1 || stats.Leased != 1 || stats.Ready != 0 || stats.Blocked != 0 {
		t.Fatalf("expected leased event to be counted separately from ready backlog, got %+v", stats)
	}
	if len(stats.Topics) != 1 || stats.Topics[0].Leased != 1 || stats.Topics[0].Ready != 0 {
		t.Fatalf("expected per-topic lease stats, got %+v", stats.Topics)
	}

	expiredClaim, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-b", LeaseSeconds: 45, Now: now.Add(31 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if expiredClaim.Claimed != 1 || len(expiredClaim.Events) != 1 || expiredClaim.Events[0].ID != events[0].ID || expiredClaim.Events[0].LeaseOwner != "relay-b" {
		t.Fatalf("expected expired lease to be claimable by relay-b, got %+v", expiredClaim)
	}
	published, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: events[0].ID, PublishedAt: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != OutboxStatusPublished || published.LeaseOwner != "" || !published.LeaseExpiresAt.IsZero() {
		t.Fatalf("expected publish ack to clear lease, got %+v", published)
	}
	afterPublished, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(afterPublished) != 0 {
		t.Fatalf("expected published event to leave pending query, got %+v", afterPublished)
	}
}

func TestOutboxStatsReportsLeaseHealthByTopicAndOwner(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_lease_health")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := events[0].CreatedAt.Add(time.Second)
	if _, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-a", LeaseSeconds: 30, Now: now}); err != nil {
		t.Fatal(err)
	}

	store.mu.Lock()
	dispatchEvent := store.enqueueOutboxEventLocked("dispatch.assigned", "dispatch", "dispatch_lease_health", "dispatch.assigned", "dispatch:lease_health", map[string]any{"order_id": "ord_dispatch"}, now)
	store.mu.Unlock()
	if dispatchEvent == nil {
		t.Fatal("expected synthetic dispatch outbox event")
	}
	dispatchClaim, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "dispatch.assigned", Limit: 1, LeaseOwner: "relay-b", LeaseSeconds: 120, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if dispatchClaim.Claimed != 1 || dispatchClaim.Events[0].ID != dispatchEvent.ID {
		t.Fatalf("expected relay-b to claim dispatch event, got %+v", dispatchClaim)
	}

	healthyStats, err := store.OutboxStats(OutboxStatsRequest{Now: now.Add(5 * time.Second), LeaseExpiringWithinSeconds: 15})
	if err != nil {
		t.Fatal(err)
	}
	if healthyStats.Total != 2 || healthyStats.Pending != 2 || healthyStats.Leased != 2 || healthyStats.Ready != 0 || healthyStats.LeaseExpiringSoon != 0 {
		t.Fatalf("expected two healthy active leases, got %+v", healthyStats)
	}
	if healthyStats.LeaseExpiringWithinSeconds != 15 || !healthyStats.NextLeaseExpiresAt.Equal(now.Add(30*time.Second)) || healthyStats.NextLeaseExpiresInSeconds != 25 {
		t.Fatalf("expected next lease expiry metadata, got %+v", healthyStats)
	}
	if len(healthyStats.LeaseOwners) != 2 || healthyStats.LeaseOwners[0].Owner != "relay-a" || healthyStats.LeaseOwners[0].Leased != 1 || healthyStats.LeaseOwners[0].LeaseExpiringSoon != 0 || healthyStats.LeaseOwners[0].NextLeaseExpiresInSeconds != 25 {
		t.Fatalf("expected relay-a healthy owner stats, got %+v", healthyStats.LeaseOwners)
	}
	if healthyStats.LeaseOwners[1].Owner != "relay-b" || healthyStats.LeaseOwners[1].Leased != 1 || healthyStats.LeaseOwners[1].NextLeaseExpiresInSeconds != 115 {
		t.Fatalf("expected relay-b healthy owner stats, got %+v", healthyStats.LeaseOwners)
	}
	if len(healthyStats.Topics) != 2 || healthyStats.Topics[0].Topic != "dispatch.assigned" || healthyStats.Topics[0].Leased != 1 || healthyStats.Topics[0].NextLeaseExpiresInSeconds != 115 || healthyStats.Topics[1].Topic != "order.paid" || healthyStats.Topics[1].NextLeaseExpiresInSeconds != 25 {
		t.Fatalf("expected per-topic lease expiry stats, got %+v", healthyStats.Topics)
	}

	soonStats, err := store.OutboxStats(OutboxStatsRequest{Now: now.Add(20 * time.Second), LeaseExpiringWithinSeconds: 15})
	if err != nil {
		t.Fatal(err)
	}
	if soonStats.Leased != 2 || soonStats.LeaseExpiringSoon != 1 || soonStats.NextLeaseExpiresInSeconds != 10 {
		t.Fatalf("expected one lease to be expiring soon, got %+v", soonStats)
	}
	if len(soonStats.LeaseOwners) != 2 || soonStats.LeaseOwners[0].Owner != "relay-a" || soonStats.LeaseOwners[0].LeaseExpiringSoon != 1 || soonStats.LeaseOwners[1].LeaseExpiringSoon != 0 {
		t.Fatalf("expected owner lease expiry health, got %+v", soonStats.LeaseOwners)
	}
	if len(soonStats.Topics) != 2 || soonStats.Topics[0].LeaseExpiringSoon != 0 || soonStats.Topics[1].Topic != "order.paid" || soonStats.Topics[1].LeaseExpiringSoon != 1 {
		t.Fatalf("expected topic lease expiry health, got %+v", soonStats.Topics)
	}

	topicStats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(20 * time.Second), LeaseExpiringWithinSeconds: 15})
	if err != nil {
		t.Fatal(err)
	}
	if topicStats.Total != 1 || topicStats.Leased != 1 || topicStats.LeaseExpiringSoon != 1 || len(topicStats.LeaseOwners) != 1 || topicStats.LeaseOwners[0].Owner != "relay-a" {
		t.Fatalf("expected topic-filtered lease owner health, got %+v", topicStats)
	}
}

func TestRenewOutboxEventLeaseRequiresCurrentActiveOwner(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_renew")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := events[0].CreatedAt.Add(time.Second)
	claimed, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-a", LeaseSeconds: 30, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Claimed != 1 || len(claimed.Events) != 1 {
		t.Fatalf("expected relay-a to claim one event before renewal, got %+v", claimed)
	}
	eventID := claimed.Events[0].ID
	renewedAt := now.Add(10 * time.Second)
	renewed, err := store.RenewOutboxEventLease(RenewOutboxEventLeaseRequest{EventID: eventID, LeaseOwner: "relay-a", LeaseSeconds: 60, Now: renewedAt})
	if err != nil {
		t.Fatal(err)
	}
	if renewed.LeaseOwner != "relay-a" || !renewed.LeaseExpiresAt.Equal(renewedAt.Add(60*time.Second)) || !renewed.UpdatedAt.Equal(renewedAt) {
		t.Fatalf("expected relay-a lease renewal to extend expiry, got %+v", renewed)
	}
	if _, err := store.RenewOutboxEventLease(RenewOutboxEventLeaseRequest{EventID: eventID, LeaseOwner: "relay-b", LeaseSeconds: 60, Now: renewedAt.Add(time.Second)}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected wrong owner renewal to conflict, got %v", err)
	}
	expiredAt := renewed.LeaseExpiresAt.Add(time.Second)
	if _, err := store.RenewOutboxEventLease(RenewOutboxEventLeaseRequest{EventID: eventID, LeaseOwner: "relay-a", LeaseSeconds: 60, Now: expiredAt}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected expired renewal to conflict, got %v", err)
	}
	reclaimed, err := store.ClaimOutboxEvents(ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-b", LeaseSeconds: 45, Now: expiredAt})
	if err != nil {
		t.Fatal(err)
	}
	if reclaimed.Claimed != 1 || reclaimed.Events[0].LeaseOwner != "relay-b" {
		t.Fatalf("expected expired lease to be claimable by relay-b, got %+v", reclaimed)
	}
	published, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: eventID, PublishedAt: expiredAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if published.LeaseOwner != "" || !published.LeaseExpiresAt.IsZero() {
		t.Fatalf("expected publish ack to clear renewed lease, got %+v", published)
	}
	if _, err := store.RenewOutboxEventLease(RenewOutboxEventLeaseRequest{EventID: eventID, LeaseOwner: "relay-b", LeaseSeconds: 60, Now: expiredAt.Add(2 * time.Second)}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected published event renewal to conflict, got %v", err)
	}
}

func TestOutboxMaxAttemptsMovesEventToDeadLetter(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_dead_letter")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	firstFailure, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "kafka unavailable", RetryAfterSeconds: 60, MaxAttempts: 2, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if firstFailure.Status != OutboxStatusFailed || firstFailure.Attempts != 1 || !firstFailure.AvailableAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("expected first failure to stay in backoff, got %+v", firstFailure)
	}
	deadLetter, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "poison message", RetryAfterSeconds: 60, MaxAttempts: 2, Now: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if deadLetter.Status != OutboxStatusDeadLetter || deadLetter.Attempts != 2 || deadLetter.LastError != "poison message" {
		t.Fatalf("expected max attempts to isolate dead-letter event, got %+v", deadLetter)
	}
	pending, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: now.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected dead-letter event to leave pending relay query, got %+v", pending)
	}
	deadLetters, err := store.OutboxEvents(OutboxEventsRequest{Status: OutboxStatusDeadLetter, Topic: "order.paid", Limit: 10, Now: now.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(deadLetters) != 1 || deadLetters[0].ID != deadLetter.ID {
		t.Fatalf("expected explicit dead-letter query to return isolated event, got %+v", deadLetters)
	}
	stats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 1 || stats.DeadLetter != 1 || stats.Failed != 0 || stats.Ready != 0 || stats.Blocked != 0 {
		t.Fatalf("expected stats to count dead-letter separately, got %+v", stats)
	}
	if len(stats.Topics) != 1 || stats.Topics[0].DeadLetter != 1 {
		t.Fatalf("expected per-topic dead-letter stats, got %+v", stats.Topics)
	}
	batch, err := store.ReplayOutboxEvents(ReplayOutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: now.Add(10 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Replayed != 0 || len(batch.Events) != 0 {
		t.Fatalf("expected batch replay to skip dead-letter events, got %+v", batch)
	}
	replayed, err := store.ReplayOutboxEvent(ReplayOutboxEventRequest{EventID: deadLetter.ID, Now: now.Add(11 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Status != OutboxStatusPending || replayed.Attempts != 2 || replayed.LastError != "" {
		t.Fatalf("expected manual replay to release dead-letter event with attempts preserved, got %+v", replayed)
	}
}

func TestReplayOutboxEventMakesBackoffEventReady(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_replay")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	failed, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "relay down", RetryAfterSeconds: 300, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != OutboxStatusFailed || failed.Attempts != 1 || !failed.AvailableAt.Equal(now.Add(5*time.Minute)) {
		t.Fatalf("expected failed event with delayed retry, got %+v", failed)
	}

	replayedAt := now.Add(time.Minute)
	replayed, err := store.ReplayOutboxEvent(ReplayOutboxEventRequest{EventID: failed.ID, Now: replayedAt})
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Status != OutboxStatusPending || replayed.Attempts != 1 || replayed.LastError != "" || !replayed.AvailableAt.Equal(replayedAt) {
		t.Fatalf("expected replay to make event immediately pending-ready, got %+v", replayed)
	}
	ready, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Now: replayedAt})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 || ready[0].ID != replayed.ID {
		t.Fatalf("expected replayed event in pending query, got %+v", ready)
	}

	if _, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: replayed.ID, PublishedAt: replayedAt.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReplayOutboxEvent(ReplayOutboxEventRequest{EventID: replayed.ID, Now: replayedAt.Add(2 * time.Second)}); !errors.Is(err, ErrInvalidOrderState) {
		t.Fatalf("expected published outbox replay to be rejected, got %v", err)
	}
}

func TestReplayOutboxEventsOnlyRestoresBlockedEvents(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	blockedOrderSlow := mustPaidDispatchOrder(t, store, "outbox_batch_blocked_slow")
	blockedOrderFast := mustPaidDispatchOrder(t, store, "outbox_batch_blocked_fast")
	readyOrder := mustPaidDispatchOrder(t, store, "outbox_batch_ready")
	publishedOrder := mustPaidDispatchOrder(t, store, "outbox_batch_published")

	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	eventsByAggregateID := map[string]OutboxEvent{}
	for _, event := range events {
		eventsByAggregateID[event.AggregateID] = event
	}
	blockedSlow := eventsByAggregateID[blockedOrderSlow.ID]
	blockedFast := eventsByAggregateID[blockedOrderFast.ID]
	ready := eventsByAggregateID[readyOrder.ID]
	published := eventsByAggregateID[publishedOrder.ID]
	if blockedSlow.ID == "" || blockedFast.ID == "" || ready.ID == "" || published.ID == "" {
		t.Fatalf("expected order.paid outbox events for all orders, got %+v", events)
	}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	if _, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: blockedSlow.ID, Error: "kafka down", RetryAfterSeconds: 300, Now: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: blockedFast.ID, Error: "kafka down", RetryAfterSeconds: 180, Now: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: ready.ID, Error: "already due", RetryAfterSeconds: 30, Now: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: published.ID, PublishedAt: now.Add(10 * time.Second)}); err != nil {
		t.Fatal(err)
	}

	store.mu.Lock()
	otherTopic := store.enqueueOutboxEventLocked("dispatch.assigned", "dispatch", "dispatch_1", "dispatch.assigned", "dispatch:1", map[string]any{"order_id": "ord_other"}, now)
	store.mu.Unlock()
	if otherTopic == nil {
		t.Fatal("expected synthetic dispatch outbox event")
	}
	if _, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: otherTopic.ID, Error: "other topic", RetryAfterSeconds: 300, Now: now}); err != nil {
		t.Fatal(err)
	}

	replayAt := now.Add(time.Minute)
	firstBatch, err := store.ReplayOutboxEvents(ReplayOutboxEventsRequest{Topic: "order.paid", Limit: 1, Now: replayAt})
	if err != nil {
		t.Fatal(err)
	}
	if firstBatch.Topic != "order.paid" || firstBatch.Limit != 1 || firstBatch.Replayed != 1 || len(firstBatch.Events) != 1 {
		t.Fatalf("expected one replayed event in first batch, got %+v", firstBatch)
	}
	if firstBatch.Events[0].ID != blockedFast.ID {
		t.Fatalf("expected earliest blocked event to replay first, got %+v", firstBatch.Events)
	}
	if firstBatch.Events[0].Status != OutboxStatusPending || firstBatch.Events[0].Attempts != 1 || firstBatch.Events[0].LastError != "" || !firstBatch.Events[0].AvailableAt.Equal(replayAt) {
		t.Fatalf("expected replayed event to become pending-ready with audit attempts preserved, got %+v", firstBatch.Events[0])
	}

	secondBatch, err := store.ReplayOutboxEvents(ReplayOutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: replayAt})
	if err != nil {
		t.Fatal(err)
	}
	if secondBatch.Replayed != 1 || len(secondBatch.Events) != 1 || secondBatch.Events[0].ID != blockedSlow.ID {
		t.Fatalf("expected only remaining blocked order.paid event to replay, got %+v", secondBatch)
	}

	readyEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: replayAt})
	if err != nil {
		t.Fatal(err)
	}
	readyByID := map[string]OutboxEvent{}
	for _, event := range readyEvents {
		readyByID[event.ID] = event
	}
	if len(readyByID) != 3 {
		t.Fatalf("expected two replayed events plus already-ready event, got %+v", readyEvents)
	}
	if readyByID[blockedFast.ID].ID == "" || readyByID[blockedSlow.ID].ID == "" || readyByID[ready.ID].ID == "" {
		t.Fatalf("expected replayed and already-ready events in pending relay query, got %+v", readyEvents)
	}
	if readyByID[ready.ID].AvailableAt.Equal(replayAt) || readyByID[ready.ID].LastError == "" {
		t.Fatalf("expected already-ready failed event to be skipped by batch replay, got %+v", readyByID[ready.ID])
	}
	if _, ok := readyByID[published.ID]; ok {
		t.Fatalf("expected published event to stay out of pending relay query, got %+v", readyEvents)
	}
}

func TestOutboxAdminOperationsWithAuditRecordsVerifiedAudit(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("outbox-admin-audit-secret")
	mustPaidDispatchOrder(t, store, "outbox_admin_audit")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid event, got %+v", events)
	}
	now := events[0].AvailableAt.Add(time.Minute)
	if now.IsZero() {
		now = events[0].CreatedAt.Add(time.Minute)
	}
	auditRequest := func(action string, targetType string, targetID string, at time.Time) RecordAuditLogRequest {
		return RecordAuditLogRequest{
			ActorType:  "admin",
			ActorID:    "admin_1",
			Action:     action,
			TargetType: targetType,
			TargetID:   targetID,
			RequestID:  "req_" + action,
			IPHash:     "ip_hash",
			Payload: map[string]any{
				"topic":  "caller-supplied-topic",
				"status": "caller-supplied-status",
				"token":  "must-not-persist",
			},
			CreatedAt: at,
		}
	}

	claim, claimAudit, err := store.ClaimOutboxEventsWithAudit(
		ClaimOutboxEventsRequest{Topic: "order.paid", Limit: 1, LeaseOwner: "relay-a", LeaseSeconds: 30, Now: now},
		auditRequest("admin.outbox.claimed", "outbox_topic", outboxTopicAuditTarget("order.paid"), now.Add(time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if claim.Claimed != 1 || len(claim.Events) != 1 || claim.Events[0].LeaseOwner != "relay-a" {
		t.Fatalf("expected audited claim to lease one event, got %+v", claim)
	}
	if claimAudit.Payload["claimed"] != 1 || claimAudit.Payload["lease_owner"] != "relay-a" || claimAudit.Payload["lease_seconds"] != 30 || claimAudit.Payload["token"] != nil {
		t.Fatalf("expected server-generated claim audit payload, got %+v", claimAudit.Payload)
	}
	if claimAudit.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || claimAudit.IntegrityHash == "" || !claimAudit.IntegrityVerified {
		t.Fatalf("expected verified claim audit integrity proof, got %+v", claimAudit)
	}
	eventID := claim.Events[0].ID

	renewed, renewAudit, err := store.RenewOutboxEventLeaseWithAudit(
		RenewOutboxEventLeaseRequest{EventID: eventID, LeaseOwner: "relay-a", LeaseSeconds: 90, Now: now.Add(10 * time.Second)},
		auditRequest("admin.outbox.lease_renewed", "outbox_event", eventID, now.Add(11*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if renewed.LeaseOwner != "relay-a" || renewAudit.Payload["lease_seconds"] != 90 || renewAudit.Payload["status"] != OutboxStatusPending {
		t.Fatalf("expected audited lease renewal payload, event=%+v audit=%+v", renewed, renewAudit.Payload)
	}

	failed, failedAudit, err := store.MarkOutboxEventFailedWithAudit(
		MarkOutboxEventFailedRequest{EventID: eventID, Error: "relay down", RetryAfterSeconds: 120, Now: now.Add(20 * time.Second)},
		auditRequest("admin.outbox.failed", "outbox_event", eventID, now.Add(21*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != OutboxStatusFailed || failed.Attempts != 1 || failedAudit.Payload["attempts"] != 1 || failedAudit.Payload["retry_after_seconds"] != 120 {
		t.Fatalf("expected audited failure payload, event=%+v audit=%+v", failed, failedAudit.Payload)
	}

	replayed, replayAudit, err := store.ReplayOutboxEventWithAudit(
		ReplayOutboxEventRequest{EventID: eventID, Now: now.Add(30 * time.Second)},
		auditRequest("admin.outbox.replayed", "outbox_event", eventID, now.Add(31*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Status != OutboxStatusPending || replayAudit.Payload["status"] != OutboxStatusPending || replayAudit.Payload["topic"] != "order.paid" {
		t.Fatalf("expected audited replay payload, event=%+v audit=%+v", replayed, replayAudit.Payload)
	}

	published, publishedAudit, err := store.MarkOutboxEventPublishedWithAudit(
		MarkOutboxEventPublishedRequest{EventID: eventID, PublishedAt: now.Add(40 * time.Second)},
		auditRequest("admin.outbox.published", "outbox_event", eventID, now.Add(41*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != OutboxStatusPublished || publishedAudit.Payload["status"] != OutboxStatusPublished || !publishedAudit.IntegrityVerified {
		t.Fatalf("expected audited publish payload, event=%+v audit=%+v", published, publishedAudit)
	}

	blockedOrder := mustPaidDispatchOrder(t, store, "outbox_admin_batch_audit")
	pendingEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	var blockedEvent OutboxEvent
	for _, event := range pendingEvents {
		if event.AggregateID == blockedOrder.ID {
			blockedEvent = event
			break
		}
	}
	if blockedEvent.ID == "" {
		t.Fatalf("expected second order outbox event, got %+v", pendingEvents)
	}
	if _, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: blockedEvent.ID, Error: "kafka down", RetryAfterSeconds: 300, Now: now}); err != nil {
		t.Fatal(err)
	}
	batch, batchAudit, err := store.ReplayOutboxEventsWithAudit(
		ReplayOutboxEventsRequest{Topic: "order.paid", Limit: 5, Now: now.Add(time.Minute)},
		auditRequest("admin.outbox.batch_replayed", "outbox_topic", outboxTopicAuditTarget("order.paid"), now.Add(61*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Replayed != 1 || len(batch.Events) != 1 || batch.Events[0].ID != blockedEvent.ID {
		t.Fatalf("expected audited batch replay for blocked event, got %+v", batch)
	}
	if batchAudit.Payload["replayed"] != 1 || batchAudit.Payload["limit"] != 5 || batchAudit.Payload["topic"] != "order.paid" || !batchAudit.IntegrityVerified {
		t.Fatalf("expected audited batch replay payload, got %+v", batchAudit)
	}

	logs, err := store.AuditLogs(AuditLogsRequest{TargetType: "outbox_event", TargetID: eventID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 4 {
		t.Fatalf("expected four outbox event audit logs for event %s, got %+v", eventID, logs)
	}
	for _, log := range logs {
		if !log.IntegrityVerified || log.Payload["token"] != nil {
			t.Fatalf("expected verified sanitized outbox audit log, got %+v", log)
		}
	}
	topicLogs, err := store.AuditLogs(AuditLogsRequest{TargetType: "outbox_topic", TargetID: outboxTopicAuditTarget("order.paid"), Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(topicLogs) != 2 || topicLogs[0].Action != "admin.outbox.batch_replayed" || topicLogs[1].Action != "admin.outbox.claimed" {
		t.Fatalf("expected claim and batch replay outbox topic audit logs, got %+v", topicLogs)
	}
}

func TestOutboxStatsBacklogReadiness(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	mustPaidDispatchOrder(t, store, "outbox_stats")
	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one order.paid outbox event, got %+v", events)
	}

	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	failed, err := store.MarkOutboxEventFailed(MarkOutboxEventFailedRequest{EventID: events[0].ID, Error: "relay down", RetryAfterSeconds: 120, Now: now})
	if err != nil {
		t.Fatal(err)
	}

	blockedStats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if blockedStats.Total != 1 || blockedStats.Failed != 1 || blockedStats.Ready != 0 || blockedStats.Blocked != 1 || blockedStats.Published != 0 {
		t.Fatalf("expected failed event blocked during backoff, got %+v", blockedStats)
	}
	if len(blockedStats.Topics) != 1 || blockedStats.Topics[0].Topic != "order.paid" || blockedStats.Topics[0].Blocked != 1 || blockedStats.Topics[0].Ready != 0 {
		t.Fatalf("expected per-topic blocked stats, got %+v", blockedStats.Topics)
	}

	readyStats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(3 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if readyStats.Total != 1 || readyStats.Failed != 1 || readyStats.Ready != 1 || readyStats.Blocked != 0 {
		t.Fatalf("expected failed event ready after backoff, got %+v", readyStats)
	}
	if readyStats.OldestReadyLagSeconds != 60 {
		t.Fatalf("expected oldest ready lag of 60 seconds, got %+v", readyStats)
	}
	if len(readyStats.Topics) != 1 || readyStats.Topics[0].Ready != 1 || readyStats.Topics[0].OldestReadyLagSeconds != 60 {
		t.Fatalf("expected per-topic ready lag stats, got %+v", readyStats.Topics)
	}

	if _, err := store.MarkOutboxEventPublished(MarkOutboxEventPublishedRequest{EventID: failed.ID, PublishedAt: now.Add(4 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	publishedStats, err := store.OutboxStats(OutboxStatsRequest{Topic: "order.paid", Now: now.Add(5 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if publishedStats.Total != 1 || publishedStats.Published != 1 || publishedStats.Failed != 0 || publishedStats.Ready != 0 || publishedStats.Blocked != 0 {
		t.Fatalf("expected published event to leave ready backlog, got %+v", publishedStats)
	}
	if len(publishedStats.Topics) != 1 || publishedStats.Topics[0].Published != 1 || publishedStats.Topics[0].Ready != 0 {
		t.Fatalf("expected per-topic published stats, got %+v", publishedStats.Topics)
	}
}

func TestAuditLogsRecordFilterAndProtectPayloadCopies(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	payload := map[string]any{
		"default_refund_strategy": RefundStrategyBalanceFirst,
		"object_key":              "after-sales/asr_1/private/evidence.jpg",
		"password":                "PlainTextPassword",
		"token":                   "secret-token",
		"phone":                   "13900000000",
		"nested":                  map[string]any{"authorization": "Bearer secret"},
		"raw_request":             map[string]any{"body": "full request"},
	}
	log, err := store.RecordAuditLog(RecordAuditLogRequest{
		ActorType:  "admin",
		ActorID:    "admin_1",
		Action:     "admin.refund_settings.updated",
		TargetType: "refund_settings",
		TargetID:   "default",
		RequestID:  "req_1",
		IPHash:     "ip_hash",
		Payload:    payload,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatal(err)
	}
	payload["default_refund_strategy"] = "mutated"
	log.Payload["default_refund_strategy"] = "mutated_result"

	if _, err := store.RecordAuditLog(RecordAuditLogRequest{ActorType: "admin", ActorID: "admin_1", Action: "admin.outbox.replayed", TargetType: "outbox_event", TargetID: "obe_1", CreatedAt: now.Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}

	logs, err := store.AuditLogs(AuditLogsRequest{ActorType: "admin", TargetType: "refund_settings", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != log.ID || logs[0].Payload["default_refund_strategy"] != RefundStrategyBalanceFirst || logs[0].Payload["object_key"] != "aft***pg" || logs[0].RequestID != "req_1" || logs[0].IPHash != "ip_hash" {
		t.Fatalf("expected filtered immutable audit log, got %+v", logs)
	}
	if logs[0].IntegrityAlgorithm != auditIntegrityAlgorithmSHA256 || logs[0].IntegrityHash == "" || !logs[0].IntegrityVerified {
		t.Fatalf("expected default audit integrity proof to verify, got %+v", logs[0])
	}
	for _, key := range []string{"password", "token", "phone", "nested", "raw_request"} {
		if _, ok := logs[0].Payload[key]; ok {
			t.Fatalf("expected audit payload to drop sensitive or non-allowlisted key %q, got %+v", key, logs[0].Payload)
		}
	}

	allLogs, err := store.AuditLogs(AuditLogsRequest{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(allLogs) != 1 || allLogs[0].Action != "admin.outbox.replayed" {
		t.Fatalf("expected newest audit log first with limit, got %+v", allLogs)
	}

	windowLogs, err := store.AuditLogs(AuditLogsRequest{
		After:  now.Add(30 * time.Second),
		Before: now.Add(2 * time.Minute),
		Limit:  10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(windowLogs) != 1 || windowLogs[0].Action != "admin.outbox.replayed" {
		t.Fatalf("expected after/before window to keep only replay audit log, got %+v", windowLogs)
	}
}

func TestAuditLogIntegrityDetectsTamperedHMACPayload(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("audit-secret")
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)

	log, err := store.RecordAuditLog(RecordAuditLogRequest{
		ActorType:  "admin",
		ActorID:    "admin_1",
		Action:     "admin.order.refunded",
		TargetType: "order",
		TargetID:   "ord_integrity",
		RequestID:  "req_integrity",
		IPHash:     "ip_hash",
		Payload:    map[string]any{"amount_fen": int64(1200), "idempotency_key": "refund_integrity"},
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if log.IntegrityAlgorithm != auditIntegrityAlgorithmHMACSHA256 || log.IntegrityHash == "" || !log.IntegrityVerified {
		t.Fatalf("expected HMAC audit integrity proof, got %+v", log)
	}
	if verifyAuditLogIntegrity(*log, "wrong-secret") {
		t.Fatalf("expected wrong audit signing secret to fail verification")
	}

	logs, err := store.AuditLogs(AuditLogsRequest{TargetType: "order", TargetID: "ord_integrity", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || !logs[0].IntegrityVerified {
		t.Fatalf("expected sealed audit log to verify before tamper, got %+v", logs)
	}

	store.mu.Lock()
	store.auditLogs[log.ID].Payload["amount_fen"] = int64(900)
	store.mu.Unlock()

	tamperedLogs, err := store.AuditLogs(AuditLogsRequest{TargetType: "order", TargetID: "ord_integrity", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(tamperedLogs) != 1 || tamperedLogs[0].IntegrityVerified || tamperedLogs[0].IntegrityHash != log.IntegrityHash {
		t.Fatalf("expected tampered audit payload to fail integrity verification without resealing, got %+v", tamperedLogs)
	}
}

func TestAuditRetentionReportFlagsRetentionCoverageAndIntegrity(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	store.ConfigureAuditLogIntegrity("audit-retention-secret")
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	if _, err := store.RecordAuditLog(RecordAuditLogRequest{
		ActorType:  "admin",
		ActorID:    "admin_1",
		Action:     "admin.order.refunded",
		TargetType: "order",
		TargetID:   "ord_retention",
		Payload:    map[string]any{"amount_fen": int64(1200), "idempotency_key": "refund_retention"},
		CreatedAt:  now.AddDate(0, 0, -10),
	}); err != nil {
		t.Fatal(err)
	}
	export, err := store.RecordAuditLog(RecordAuditLogRequest{
		ActorType:  "security_auditor",
		ActorID:    "auditor_1",
		Action:     "admin.audit_logs.exported",
		TargetType: "audit_export",
		TargetID:   "audit-logs-retention.csv",
		Payload:    map[string]any{"export_format": "csv", "row_count": 1},
		CreatedAt:  now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	store.mu.Lock()
	store.auditLogs[export.ID].Payload["row_count"] = 99
	store.mu.Unlock()

	report, err := store.AuditRetentionReport(AuditRetentionReportRequest{
		RetentionDays:        7,
		HotDays:              1,
		IntegritySampleLimit: 10,
		CriticalActions:      []string{"admin.order.refunded", "admin.audit_logs.exported", "after_sales.reviewed"},
		Now:                  now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "critical" || report.TotalLogs != 2 || report.ExpiredLogs != 1 || report.ColdArchiveDueLogs != 1 || report.ExportEvents != 1 {
		t.Fatalf("expected critical retention report with expired/archive/export counts, got %+v", report)
	}
	if report.IntegritySampleSize != 2 || report.IntegrityFailures != 1 {
		t.Fatalf("expected integrity failure in sampled logs, got %+v", report)
	}
	if len(report.MissingCriticalActions) != 1 || report.MissingCriticalActions[0] != "after_sales.reviewed" {
		t.Fatalf("expected missing after-sales critical action, got %+v", report.MissingCriticalActions)
	}
	alerts := map[string]AuditRetentionAlert{}
	for _, alert := range report.Alerts {
		alerts[alert.Code] = alert
	}
	for _, code := range []string{"audit.integrity_failed", "audit.retention_expired", "audit.archive_due", "audit.missing_critical_action"} {
		if _, ok := alerts[code]; !ok {
			t.Fatalf("expected alert %s in report, got %+v", code, report.Alerts)
		}
	}
}

func mustPaidDispatchOrder(t *testing.T, store *Store, suffix string) *Order {
	t.Helper()
	userID := "user_" + suffix
	order, err := store.CreateOrder(CreateOrderRequest{UserID: userID, Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: userID, AmountFen: 1200, IdempotencyKey: "credit_" + suffix}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: userID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{UserID: userID, OrderID: order.ID, PaymentPassword: "123456", IdempotencyKey: "pay_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	return paidOrder
}
