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
