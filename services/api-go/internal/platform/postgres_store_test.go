package platform

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStoreSnapshotRoundTripRestoresStateAndIndexes(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	invite, err := store.CreateMerchantInvite(CreateMerchantInviteRequest{AdminID: "admin_1"})
	if err != nil {
		t.Fatal(err)
	}
	merchantProfile, err := store.AcceptMerchantInvite(AcceptMerchantInviteRequest{
		Token:       invite.Token,
		DisplayName: "快蓝商户",
		AccountType: MerchantAccountStandard,
		Password:    "MerchantPass123",
	})
	if err != nil {
		t.Fatal(err)
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
	prepay, _, err := store.CreateWechatPrepay(WechatPrepayRequest{UserID: "user_1", OrderID: order.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_snapshot_tx",
		AmountFen:     3999,
	}); err != nil {
		t.Fatal(err)
	}
	dispatchOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_snapshot_dispatch"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidDispatchOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_1",
		OrderID:         dispatchOrder.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_snapshot_dispatch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, decision, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidDispatchOrder.ID, Now: paidDispatchOrder.CreatedAt.Add(10 * time.Minute)}); err != nil {
		t.Fatal(err)
	} else if decision.StationID != "station_1" {
		t.Fatalf("expected station_1 dispatch decision, got %+v", decision)
	}

	payload, err := store.snapshotPayload()
	if err != nil {
		t.Fatal(err)
	}
	var snapshot storeSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.StationServiceAreas["station_1"] == nil || len(snapshot.StationServiceAreas["station_1"].ShopIDs) == 0 {
		t.Fatalf("expected station service area in snapshot, got %+v", snapshot.StationServiceAreas)
	}
	if len(snapshot.DispatchEvents) == 0 {
		t.Fatalf("expected dispatch events in snapshot, got %+v", snapshot.DispatchEvents)
	}
	if len(snapshot.OutboxEvents) == 0 || snapshot.NextOutboxEventID == 0 {
		t.Fatalf("expected outbox events in snapshot, got %+v next=%d", snapshot.OutboxEvents, snapshot.NextOutboxEventID)
	}

	restored := NewStore(DefaultHomeModules())
	restored.applySnapshot(snapshot)

	restoredMerchant, err := restored.LoginMerchant(MerchantLoginRequest{AccountID: merchantProfile.Account.ID, Password: "MerchantPass123"})
	if err != nil {
		t.Fatal(err)
	}
	if restoredMerchant.Account.ID != merchantProfile.Account.ID {
		t.Fatalf("expected restored merchant password hash to authorize login, got %+v", restoredMerchant)
	}
	events, err := restored.DispatchEvents(paidDispatchOrder.ID, "station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != "dispatch.auto_assign" || events[0].StationID != "station_1" {
		t.Fatalf("expected restored dispatch audit event, got %+v", events)
	}
	outboxEvents, err := restored.OutboxEvents(OutboxEventsRequest{Topic: "dispatch.assigned", Limit: 10, Now: paidDispatchOrder.CreatedAt.Add(11 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxEvents) != 1 || outboxEvents[0].AggregateID != paidDispatchOrder.ID || outboxEvents[0].Status != OutboxStatusPending {
		t.Fatalf("expected restored dispatch outbox event, got %+v", outboxEvents)
	}

	duplicateTx, restoredOrder, err := restored.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_snapshot_tx",
		AmountFen:     3999,
	})
	if err != nil {
		t.Fatal(err)
	}
	if duplicateTx.TransactionID != "wx_snapshot_tx" || restoredOrder.ID != order.ID {
		t.Fatalf("expected restored payment indexes, tx=%+v order=%+v", duplicateTx, restoredOrder)
	}

	vouchers, err := restored.UserGroupbuyVouchers("user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(vouchers) != 1 {
		t.Fatalf("expected restored voucher, got %+v", vouchers)
	}
	redeemed, completedOrder, err := restored.RedeemGroupbuyVoucher(RedeemGroupbuyVoucherRequest{
		MerchantID:  "merchant_1",
		VoucherCode: vouchers[0].VoucherCode,
		Method:      GroupbuyRedemptionMethodQR,
	})
	if err != nil {
		t.Fatal(err)
	}
	if redeemed.Status != GroupbuyVoucherRedeemed || completedOrder.Status != StatusCompleted {
		t.Fatalf("expected restored voucher index to redeem, voucher=%+v order=%+v", redeemed, completedOrder)
	}
}

func TestPaymentDomainTableRestoreRebuildsPaymentAndWalletIndexes(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 2000, IdempotencyKey: "credit_table_restore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	walletTx, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_1",
		OrderID:         order.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_table_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if account.Balance != 800 || paidOrder.PaymentMethod != PaymentBalance {
		t.Fatalf("expected paid balance order before restore, order=%+v account=%+v", paidOrder, account)
	}

	wechatOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_2", Type: OrderTypeCourier, AmountFen: 900})
	if err != nil {
		t.Fatal(err)
	}
	prepay, _, err := store.CreateWechatPrepay(WechatPrepayRequest{UserID: "user_2", OrderID: wechatOrder.ID})
	if err != nil {
		t.Fatal(err)
	}
	paidWechatTx, _, err := store.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_table_restore_tx",
		AmountFen:     900,
	})
	if err != nil {
		t.Fatal(err)
	}

	domain := store.paymentDomainSnapshot()
	if len(domain.Orders) != 2 || len(domain.WalletTransactions) != 2 || len(domain.PaymentTransactions) != 1 {
		t.Fatalf("expected payment domain snapshot to include orders and ledgers, got %+v", domain)
	}

	restored := NewStore(DefaultHomeModules())
	restored.replacePaymentDomainFromTables(domain.Orders, domain.Wallets, domain.WalletTransactions, domain.PaymentPasswordHash, domain.PaymentTransactions)

	duplicateWalletTx, duplicateOrder, duplicateAccount, err := restored.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_1",
		OrderID:         order.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_table_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if duplicateWalletTx.ID != walletTx.ID || duplicateOrder.ID != paidOrder.ID || duplicateAccount.Balance != 800 {
		t.Fatalf("expected restored wallet idempotency, tx=%+v order=%+v account=%+v", duplicateWalletTx, duplicateOrder, duplicateAccount)
	}

	newOrder, err := restored.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := restored.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_1",
		OrderID:         newOrder.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_after_table_restore",
	}); err != nil {
		t.Fatalf("expected restored payment password hash to authorize new balance payment, got %v", err)
	}

	duplicateWechatTx, duplicateWechatOrder, err := restored.ConfirmWechatPayment(WechatPaymentCallbackRequest{
		OutTradeNo:    prepay.OutTradeNo,
		TransactionID: "wx_table_restore_tx",
		AmountFen:     900,
	})
	if err != nil {
		t.Fatal(err)
	}
	if duplicateWechatTx.ID != paidWechatTx.ID || duplicateWechatOrder.ID != wechatOrder.ID {
		t.Fatalf("expected restored wechat transaction indexes, tx=%+v order=%+v", duplicateWechatTx, duplicateWechatOrder)
	}

	detail, err := restored.OrderByID(order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Events) == 0 || detail.Events[len(detail.Events)-1].Type != "order.payment.success" {
		t.Fatalf("expected restored order events, got %+v", detail.Events)
	}
}

func TestSQLCreateOrderSideEffectsRestoreMirror(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	createdAt := time.Date(2026, 5, 22, 11, 0, 0, 0, time.UTC)
	createdEvent := OrderEvent{
		Type:      "order.created",
		ActorID:   "user_sql_create",
		Message:   "订单已创建",
		CreatedAt: createdAt,
	}
	sqlOrder := Order{
		ID:        "ord_9",
		UserID:    "user_sql_create",
		Type:      OrderTypeCourier,
		Status:    StatusPendingPayment,
		AmountFen: 1500,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		Events:    []OrderEvent{createdEvent},
	}

	store.replacePaymentDomainFromTables([]Order{sqlOrder}, nil, nil, nil, nil)
	store.applyOrderEventOutboxAfterSQL(sqlOrder.ID, createdEvent)

	detail, err := store.OrderByID(sqlOrder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.UserID != sqlOrder.UserID || len(detail.Events) != 1 || detail.Events[0].Type != "order.created" {
		t.Fatalf("expected SQL-created order mirror to restore detail and event, got %+v", detail)
	}
	userOrders, err := store.UserOrders(sqlOrder.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(userOrders) != 1 || userOrders[0].ID != sqlOrder.ID {
		t.Fatalf("expected SQL-created order mirror to restore user order index, got %+v", userOrders)
	}
	outboxEvents, err := store.OutboxEvents(OutboxEventsRequest{Limit: 10, Now: createdAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxEvents) != 0 {
		t.Fatalf("expected order.created to remain local audit only, got %+v", outboxEvents)
	}

	nextOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_sql_create", Type: OrderTypeCourier, AmountFen: 300})
	if err != nil {
		t.Fatal(err)
	}
	if nextOrder.ID != "ord_10" {
		t.Fatalf("expected restored SQL order number to advance next in-memory order id, got %s", nextOrder.ID)
	}

	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: sqlOrder.UserID, AmountFen: 1500, IdempotencyKey: "credit_sql_create_restore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: sqlOrder.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{
		UserID:          sqlOrder.UserID,
		OrderID:         sqlOrder.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_sql_create_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if paidOrder.Status != StatusDispatching || account.Balance != 0 {
		t.Fatalf("expected restored SQL-created order to continue through balance payment, order=%+v account=%+v", paidOrder, account)
	}
	paidOutboxEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: paidOrder.UpdatedAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(paidOutboxEvents) != 1 || paidOutboxEvents[0].AggregateID != sqlOrder.ID {
		t.Fatalf("expected restored SQL-created order payment to enqueue order.paid outbox, got %+v", paidOutboxEvents)
	}
}

func TestPaymentDomainSnapshotIncludesProductsAndCartForSQLCheckout(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	summary, err := store.UpsertCartItem(UpsertCartItemRequest{
		UserID:    "user_sql_checkout",
		ShopID:    "shop_1",
		ProductID: "prod_beef_rice",
		Quantity:  2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.PayableFen != 5598 {
		t.Fatalf("expected seeded product cart payable, got %+v", summary)
	}

	domain := store.paymentDomainSnapshot()
	productFound := false
	for _, product := range domain.Products {
		if product.ID == "prod_beef_rice" && product.ShopID == "shop_1" && product.Status == ProductStatusActive {
			productFound = true
			break
		}
	}
	if !productFound {
		t.Fatalf("expected takeout products in normalized SQL snapshot, got %+v", domain.Products)
	}
	qualificationTypes := map[string]bool{}
	for _, qualification := range domain.Qualifications {
		if qualification.MerchantID == "merchant_1" {
			qualificationTypes[qualification.Qualification.Type] = true
		}
	}
	if !qualificationTypes[QualificationBusinessLicense] || !qualificationTypes[QualificationHealthCertificate] {
		t.Fatalf("expected merchant qualifications in normalized SQL snapshot, got %+v", domain.Qualifications)
	}
	if len(domain.CartItems) != 1 || domain.CartItems[0].UserID != "user_sql_checkout" || domain.CartItems[0].ProductID != "prod_beef_rice" {
		t.Fatalf("expected cart items in normalized SQL snapshot, got %+v", domain.CartItems)
	}
	userFound := false
	for _, user := range domain.Users {
		if user.ID == "user_sql_checkout" {
			userFound = true
			break
		}
	}
	if !userFound {
		t.Fatalf("expected cart user placeholder in normalized SQL snapshot, got %+v", domain.Users)
	}
}

func TestSQLCheckoutCartSideEffectsRestoreMirror(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	summary, err := store.UpsertCartItem(UpsertCartItemRequest{
		UserID:    "user_sql_checkout",
		ShopID:    "shop_1",
		ProductID: "prod_beef_rice",
		Quantity:  2,
	})
	if err != nil {
		t.Fatal(err)
	}

	createdAt := time.Date(2026, 5, 22, 11, 30, 0, 0, time.UTC)
	event := OrderEvent{
		Type:      "order.checkout_created",
		ActorID:   "user_sql_checkout",
		Message:   "购物车结算创建订单",
		CreatedAt: createdAt,
	}
	orderItems := make([]OrderItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		orderItems = append(orderItems, OrderItem{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			UnitPriceFen: item.UnitPriceFen,
			Quantity:     item.Quantity,
		})
	}
	sqlOrder := Order{
		ID:              "ord_12",
		UserID:          "user_sql_checkout",
		ShopID:          "shop_1",
		AddressID:       "addr_sql_checkout",
		Type:            OrderTypeTakeout,
		Status:          StatusPendingPayment,
		AmountFen:       summary.PayableFen,
		ItemsTotalFen:   summary.ItemsTotalFen,
		DeliveryFeeFen:  summary.DeliveryFeeFen,
		PackagingFeeFen: summary.PackagingFeeFen,
		DiscountFen:     summary.DiscountFen,
		Items:           orderItems,
		Options:         OrderOptions{Remark: "少放辣", TablewareCount: 2},
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
		Events:          []OrderEvent{event},
	}

	store.replacePaymentDomainFromTables([]Order{sqlOrder}, nil, nil, nil, nil)
	store.applyOrderEventOutboxAfterSQL(sqlOrder.ID, event)
	store.clearCartAfterSQLCheckout("user_sql_checkout", "shop_1")

	emptySummary, err := store.CartSummary("user_sql_checkout", "shop_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(emptySummary.Items) != 0 || emptySummary.PayableFen != 0 {
		t.Fatalf("expected SQL checkout mirror to clear cart, got %+v", emptySummary)
	}
	detail, err := store.OrderByID(sqlOrder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.AmountFen != summary.PayableFen || len(detail.Items) != 1 || detail.Events[0].Type != "order.checkout_created" {
		t.Fatalf("expected SQL checkout order mirror to restore order detail, got %+v", detail)
	}
	outboxEvents, err := store.OutboxEvents(OutboxEventsRequest{Limit: 10, Now: createdAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxEvents) != 0 {
		t.Fatalf("expected checkout creation to stay local audit only, got %+v", outboxEvents)
	}
	nextOrder, err := store.CreateOrder(CreateOrderRequest{UserID: "user_sql_checkout", Type: OrderTypeCourier, AmountFen: 300})
	if err != nil {
		t.Fatal(err)
	}
	if nextOrder.ID != "ord_13" {
		t.Fatalf("expected restored SQL checkout order number to advance next in-memory order id, got %s", nextOrder.ID)
	}

	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: sqlOrder.UserID, AmountFen: sqlOrder.AmountFen, IdempotencyKey: "credit_sql_checkout_restore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: sqlOrder.UserID, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, account, err := store.PayOrderWithBalance(BalancePayRequest{
		UserID:          sqlOrder.UserID,
		OrderID:         sqlOrder.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_sql_checkout_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if paidOrder.Status != StatusMerchantPending || account.Balance != 0 {
		t.Fatalf("expected restored SQL checkout order to continue through balance payment, order=%+v account=%+v", paidOrder, account)
	}
}

func TestDispatchEventTableRestoreRebuildsAuditAndCompensationIndexes(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateOrder(CreateOrderRequest{UserID: "user_1", Type: OrderTypeTakeout, AmountFen: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreditWallet(CreditWalletRequest{UserID: "user_1", AmountFen: 1200, IdempotencyKey: "credit_dispatch_event_table_restore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_1", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, paidOrder, _, err := store.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_1",
		OrderID:         order.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_dispatch_event_table_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	assignNow := paidOrder.CreatedAt.Add(10 * time.Minute)
	if _, _, err := store.AutoAssignOrder(AutoAssignOrderRequest{OrderID: paidOrder.ID, Now: assignNow}); err != nil {
		t.Fatal(err)
	}
	reassignedOrder, _, err := store.RejectRiderAssignment(RejectRiderAssignmentRequest{OrderID: paidOrder.ID, RiderID: "rider_1", Now: assignNow.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if reassignedOrder.Status != StatusRiderAssigned || reassignedOrder.RiderID != "rider_2" {
		t.Fatalf("expected rejection to reassign rider_2 before restore, got %+v", reassignedOrder)
	}

	domain := store.paymentDomainSnapshot()
	for i := range domain.Orders {
		if domain.Orders[i].ID != paidOrder.ID {
			continue
		}
		domain.Orders[i].Status = StatusDispatching
		domain.Orders[i].RiderID = ""
		domain.Orders[i].Events = []OrderEvent{{
			Type:      "order.payment.success",
			ActorID:   "payment",
			Message:   "支付成功，订单进入派送中",
			CreatedAt: paidOrder.UpdatedAt,
		}}
	}
	events := store.dispatchEventSnapshot()
	if len(events) != 3 || events[0].Type != "dispatch.auto_assign" || events[1].Type != "dispatch.rejected" || events[2].RiderID != "rider_2" {
		t.Fatalf("expected auto/reject/auto dispatch snapshot, got %+v", events)
	}

	restored := NewStore(DefaultHomeModules())
	restored.replacePaymentDomainFromTables(domain.Orders, domain.Wallets, domain.WalletTransactions, domain.PaymentPasswordHash, domain.PaymentTransactions)
	restored.replaceDispatchEventsFromTable(events)

	auditEvents, err := restored.DispatchEvents(paidOrder.ID, "station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(auditEvents) != 3 || auditEvents[1].RiderID != "rider_1" || auditEvents[2].RiderID != "rider_2" {
		t.Fatalf("expected restored station audit events, got %+v", auditEvents)
	}
	if len(auditEvents[2].RejectedRiderIDs) != 1 || auditEvents[2].RejectedRiderIDs[0] != "rider_1" {
		t.Fatalf("expected restored rejected-rider ledger on latest dispatch event, got %+v", auditEvents[2])
	}

	repaired, err := restored.CompensateOrderState(CompensateOrderStateRequest{OrderID: paidOrder.ID, ActorID: "admin_1", Now: assignNow.Add(2 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.Changed || repaired.ExpectedStatus != StatusRiderAssigned || repaired.ExpectedRiderID != "rider_2" || repaired.Order.RiderID != "rider_2" {
		t.Fatalf("expected dispatch table restore to drive order-state compensation, got %+v", repaired)
	}

	nextOrder, err := restored.CreateOrder(CreateOrderRequest{UserID: "user_2", Type: OrderTypeTakeout, AmountFen: 800})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := restored.CreditWallet(CreditWalletRequest{UserID: "user_2", AmountFen: 800, IdempotencyKey: "credit_dispatch_event_after_restore"}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.SetWalletPaymentPassword(SetWalletPaymentPasswordRequest{UserID: "user_2", Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	_, nextPaidOrder, _, err := restored.PayOrderWithBalance(BalancePayRequest{
		UserID:          "user_2",
		OrderID:         nextOrder.ID,
		PaymentPassword: "123456",
		IdempotencyKey:  "pay_dispatch_event_after_restore",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := restored.AutoAssignOrder(AutoAssignOrderRequest{OrderID: nextPaidOrder.ID, Now: nextPaidOrder.CreatedAt.Add(10 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	nextEvents, err := restored.DispatchEvents(nextPaidOrder.ID, "station_manager_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(nextEvents) != 1 || nextEvents[0].ID != "dpe_4" {
		t.Fatalf("expected next dispatch event id to continue after table restore, got %+v", nextEvents)
	}
}

func TestSQLBalancePaymentSideEffectsIssueGroupbuyVoucherAndOutbox(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	order, err := store.CreateGroupbuyOrder(CreateGroupbuyOrderRequest{
		UserID:   "user_1",
		ShopID:   "shop_1",
		DealID:   "deal_two_person_set",
		Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	paidAt := order.CreatedAt.Add(time.Minute)
	idempotencyKey := "pay_sql_groupbuy_side_effects"
	paidOrder := *order
	paidOrder.Status = StatusVoucherIssued
	paidOrder.PaymentMethod = PaymentBalance
	paidOrder.UpdatedAt = paidAt
	paidOrder.Events = append(paidOrder.Events, OrderEvent{
		Type:      "order.payment.success",
		ActorID:   "user_1",
		Message:   paymentSuccessMessage(&paidOrder),
		CreatedAt: paidAt,
	})
	walletTx := WalletTransaction{
		ID:             postgresWalletTransactionID(idempotencyKey),
		UserID:         "user_1",
		OrderID:        order.ID,
		Type:           "payment",
		AmountFen:      -order.AmountFen,
		PaymentMethod:  PaymentBalance,
		IdempotencyKey: idempotencyKey,
		Status:         "success",
		CreatedAt:      paidAt,
	}
	store.replacePaymentDomainFromTables(
		[]Order{paidOrder},
		[]WalletAccount{{UserID: "user_1", Balance: 1000, Version: 2, RiskState: "normal"}},
		[]WalletTransaction{walletTx},
		nil,
		nil,
	)

	store.applyBalancePaymentSideEffectsAfterSQL(order.ID, paidAt)

	vouchers, err := store.UserGroupbuyVouchers("user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(vouchers) != 1 || vouchers[0].OrderID != order.ID || vouchers[0].Status != GroupbuyVoucherStatusIssued {
		t.Fatalf("expected SQL balance payment side effects to issue voucher, got %+v", vouchers)
	}
	outboxEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.paid", Limit: 10, Now: paidAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxEvents) != 1 || outboxEvents[0].AggregateID != order.ID || outboxEvents[0].Payload["payment_method"] != PaymentBalance {
		t.Fatalf("expected SQL balance payment side effects to enqueue order.paid outbox, got %+v", outboxEvents)
	}
	resultTx, resultOrder, resultAccount, err := store.balancePaymentResult(idempotencyKey, order.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if resultTx.ID != walletTx.ID || resultOrder.Status != StatusVoucherIssued || resultAccount.Balance != 1000 {
		t.Fatalf("expected SQL balance payment result from restored ledgers, tx=%+v order=%+v account=%+v", resultTx, resultOrder, resultAccount)
	}
}

func TestSQLBalancePaymentIdempotencyMatcherRejectsCrossOrderReplay(t *testing.T) {
	transaction := WalletTransaction{
		UserID:        "user_1",
		OrderID:       "ord_1",
		Type:          "payment",
		PaymentMethod: PaymentBalance,
		Status:        "success",
	}
	if !walletTransactionMatchesBalancePayment(transaction, "user_1", "ord_1") {
		t.Fatalf("expected matching transaction to be accepted")
	}
	for name, candidate := range map[string]WalletTransaction{
		"wrong_user":   {UserID: "user_2", OrderID: "ord_1", Type: "payment", PaymentMethod: PaymentBalance, Status: "success"},
		"wrong_order":  {UserID: "user_1", OrderID: "ord_2", Type: "payment", PaymentMethod: PaymentBalance, Status: "success"},
		"wrong_type":   {UserID: "user_1", OrderID: "ord_1", Type: "credit", PaymentMethod: PaymentBalance, Status: "success"},
		"wrong_method": {UserID: "user_1", OrderID: "ord_1", Type: "payment", PaymentMethod: PaymentWechat, Status: "success"},
		"wrong_status": {UserID: "user_1", OrderID: "ord_1", Type: "payment", PaymentMethod: PaymentBalance, Status: "pending"},
	} {
		if walletTransactionMatchesBalancePayment(candidate, "user_1", "ord_1") {
			t.Fatalf("expected %s transaction to be rejected: %+v", name, candidate)
		}
	}
}

func TestSQLRefundSideEffectsRestoreRefundAndOutbox(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	refundedAt := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)
	idempotencyKey := "refund_sql_restore"
	order := Order{
		ID:            "ord_sql_refund",
		UserID:        "user_1",
		ShopID:        "shop_1",
		Type:          OrderTypeTakeout,
		Status:        StatusRefunded,
		PaymentMethod: PaymentBalance,
		AmountFen:     1200,
		CreatedAt:     refundedAt.Add(-time.Hour),
		UpdatedAt:     refundedAt,
		Events: []OrderEvent{{
			Type:      "order.payment.success",
			ActorID:   "user_1",
			Message:   "余额支付成功，订单进入骑手调度",
			CreatedAt: refundedAt.Add(-30 * time.Minute),
		}, {
			Type:      "order.refund.success",
			ActorID:   "admin_1",
			Message:   "订单退款已退回平台余额",
			CreatedAt: refundedAt,
		}},
	}
	refund := RefundTransaction{
		ID:             "rfd_" + shortHash(idempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      order.AmountFen,
		Destination:    RefundDestinationBalance,
		Status:         RefundStatusSuccess,
		Reason:         "商品售罄",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      refundedAt,
	}
	walletTx := WalletTransaction{
		ID:             postgresRefundWalletTransactionID(idempotencyKey),
		UserID:         order.UserID,
		OrderID:        order.ID,
		Type:           "refund",
		AmountFen:      order.AmountFen,
		PaymentMethod:  RefundDestinationBalance,
		IdempotencyKey: idempotencyKey,
		Status:         "success",
		CreatedAt:      refundedAt,
	}
	store.replacePaymentDomainFromTables(
		[]Order{order},
		[]WalletAccount{{UserID: "user_1", Balance: 2000, Version: 3, RiskState: "normal"}},
		[]WalletTransaction{walletTx},
		nil,
		nil,
	)
	store.replaceRefundDomainFromTables(RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst}, []RefundTransaction{refund})
	store.applyOrderEventOutboxAfterSQL(order.ID, order.Events[1])

	replayedRefund, replayedOrder, replayedAccount, err := store.refundResult(idempotencyKey)
	if err != nil {
		t.Fatal(err)
	}
	if replayedRefund.ID != refund.ID || replayedOrder.Status != StatusRefunded || replayedAccount.Balance != 2000 {
		t.Fatalf("expected SQL refund result from restored ledgers, refund=%+v order=%+v account=%+v", replayedRefund, replayedOrder, replayedAccount)
	}
	outboxEvents, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.refunded", Limit: 10, Now: refundedAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxEvents) != 1 || outboxEvents[0].AggregateID != order.ID || outboxEvents[0].Payload["amount_fen"] != int64(1200) {
		t.Fatalf("expected restored SQL refund to enqueue order.refunded outbox, got %+v", outboxEvents)
	}
}

func TestSQLRefundIdempotencyMatcherRejectsCrossOrderReplay(t *testing.T) {
	refund := RefundTransaction{
		OrderID:     "ord_1",
		UserID:      "user_1",
		AmountFen:   1200,
		Destination: RefundDestinationBalance,
	}
	if !refundTransactionMatchesRequest(refund, RefundOrderRequest{OrderID: "ord_1", UserID: "user_1", AmountFen: 1200, Destination: RefundDestinationBalance}) {
		t.Fatalf("expected matching refund replay to be accepted")
	}
	for name, req := range map[string]RefundOrderRequest{
		"wrong_order":       {OrderID: "ord_2", UserID: "user_1", AmountFen: 1200},
		"wrong_user":        {OrderID: "ord_1", UserID: "user_2", AmountFen: 1200},
		"wrong_amount":      {OrderID: "ord_1", UserID: "user_1", AmountFen: 1100},
		"wrong_destination": {OrderID: "ord_1", UserID: "user_1", AmountFen: 1200, Destination: RefundDestinationOriginalRoute},
	} {
		if refundTransactionMatchesRequest(refund, req) {
			t.Fatalf("expected %s refund replay to be rejected: %+v", name, req)
		}
	}
}

func TestAfterSalesDomainSnapshotAndRestore(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	createdAt := time.Date(2026, 5, 22, 15, 0, 0, 0, time.UTC)
	request := AfterSalesRequest{
		ID:                 "asr_9",
		OrderID:            "ord_after_sales",
		UserID:             "user_1",
		Type:               AfterSalesRefundOnly,
		Reason:             "餐品漏送",
		RequestedAmountFen: 1200,
		EvidenceURLs:       []string{"https://cdn.test/a.jpg", "https://cdn.test/a.jpg"},
		Status:             AfterSalesPendingMerchant,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	}
	store.afterSalesRequests[request.ID] = cloneAfterSalesRequest(&request)
	event := AfterSalesEvent{
		ID:            "asev_7",
		RequestID:     request.ID,
		OrderID:       request.OrderID,
		ActorID:       "user_1",
		ActorRole:     "user",
		Action:        AfterSalesActionCreated,
		Message:       "用户已提交售后申请",
		Attachments:   []string{"https://cdn.test/a.jpg", "https://cdn.test/a.jpg"},
		VisibleToUser: true,
		CreatedAt:     createdAt,
	}
	store.afterSalesEvents[event.ID] = cloneAfterSalesEvent(&event)
	evidence := AfterSalesEvidence{
		ID:             "ase_snapshot",
		RequestID:      request.ID,
		OrderID:        request.OrderID,
		ObjectKey:      "after-sales/asr_9/sig/a.jpg",
		PublicURL:      "https://cdn.infinitech.local/after-sales/asr_9/sig/a.jpg",
		FileName:       "a.jpg",
		ContentType:    "image/jpeg",
		SizeBytes:      1024,
		ContentSHA:     "sha256:test",
		UploadedByID:   "user_1",
		UploadedByRole: "user",
		Status:         AfterSalesEvidenceUploaded,
		CreatedAt:      createdAt,
		ConfirmedAt:    createdAt,
	}
	store.afterSalesEvidence[evidence.ID] = cloneAfterSalesEvidence(&evidence)
	uploadTicket := AfterSalesEvidenceUploadTicket{
		ID:             "aset_snapshot",
		RequestID:      request.ID,
		OrderID:        request.OrderID,
		Provider:       ObjectStorageProviderMinIO,
		Bucket:         "after-sales-test",
		ObjectKey:      evidence.ObjectKey,
		PublicURL:      evidence.PublicURL,
		FileName:       evidence.FileName,
		ContentType:    evidence.ContentType,
		SizeBytes:      evidence.SizeBytes,
		MaxSizeBytes:   AfterSalesEvidenceMaxBytes,
		ContentSHA:     evidence.ContentSHA,
		UploadedByID:   evidence.UploadedByID,
		UploadedByRole: evidence.UploadedByRole,
		Status:         AfterSalesUploadTicketConfirmed,
		ScanStatus:     AfterSalesUploadScanPassed,
		ScanResult:     "clean",
		CreatedAt:      createdAt.Add(-time.Minute),
		ExpiresAt:      createdAt.Add(15 * time.Minute),
		UploadedAt:     createdAt.Add(-30 * time.Second),
		ConfirmedAt:    createdAt,
		ScanCheckedAt:  createdAt.Add(-15 * time.Second),
	}
	store.afterSalesUploadTickets[uploadTicket.ID] = cloneAfterSalesEvidenceUploadTicket(&uploadTicket)

	snapshot := store.paymentDomainSnapshot()
	if len(snapshot.AfterSalesRequests) != 1 || len(snapshot.AfterSalesRequests[0].EvidenceURLs) != 1 || len(snapshot.AfterSalesEvents) != 1 || len(snapshot.AfterSalesEvents[0].Attachments) != 1 || len(snapshot.AfterSalesUploadTickets) != 1 || len(snapshot.AfterSalesEvidence) != 1 {
		t.Fatalf("expected after-sales request, event, ticket and evidence in payment domain snapshot, requests=%+v events=%+v tickets=%+v evidence=%+v", snapshot.AfterSalesRequests, snapshot.AfterSalesEvents, snapshot.AfterSalesUploadTickets, snapshot.AfterSalesEvidence)
	}

	restored := NewStore(DefaultHomeModules())
	restored.replaceAfterSalesDomainFromTables(snapshot.AfterSalesRequests, snapshot.AfterSalesEvents)
	restored.replaceAfterSalesUploadTicketsFromTables(snapshot.AfterSalesUploadTickets)
	restored.replaceAfterSalesEvidenceFromTables(snapshot.AfterSalesEvidence)
	requests, err := restored.AdminAfterSalesRequests()
	if err != nil {
		t.Fatal(err)
	}
	restoredEvent := restored.afterSalesEvents[event.ID]
	restoredTicket := restored.afterSalesUploadTickets[uploadTicket.ID]
	restoredEvidence := restored.afterSalesEvidence[evidence.ID]
	if len(requests) != 1 || requests[0].ID != request.ID || restored.nextAfterSalesID != 9 || restoredEvent == nil || len(restoredEvent.Attachments) != 1 || restored.nextAfterSalesEventID != 7 || restoredTicket == nil || restoredTicket.Status != AfterSalesUploadTicketConfirmed || restoredTicket.ScanStatus != AfterSalesUploadScanPassed || restoredTicket.ContentSHA != evidence.ContentSHA || restoredEvidence == nil || restoredEvidence.ObjectKey != evidence.ObjectKey {
		t.Fatalf("expected restored after-sales request, event, ticket, evidence and sequences, requests=%+v event=%+v ticket=%+v evidence=%+v next=%d eventNext=%d", requests, restoredEvent, restoredTicket, restoredEvidence, restored.nextAfterSalesID, restored.nextAfterSalesEventID)
	}
}

func TestSQLAfterSalesReviewResultRestoresRefundAndRequest(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	reviewedAt := time.Date(2026, 5, 22, 16, 0, 0, 0, time.UTC)
	idempotencyKey := "after_sales:asr_sql"
	order := Order{
		ID:            "ord_after_sales_sql",
		UserID:        "user_1",
		ShopID:        "shop_1",
		Type:          OrderTypeTakeout,
		Status:        StatusRefunded,
		PaymentMethod: PaymentBalance,
		AmountFen:     1200,
		CreatedAt:     reviewedAt.Add(-time.Hour),
		UpdatedAt:     reviewedAt,
	}
	refund := RefundTransaction{
		ID:             "rfd_" + shortHash(idempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      order.AmountFen,
		Destination:    RefundDestinationBalance,
		Status:         RefundStatusSuccess,
		Reason:         "确认漏送",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      reviewedAt,
	}
	request := AfterSalesRequest{
		ID:                 "asr_sql",
		OrderID:            order.ID,
		UserID:             order.UserID,
		Type:               AfterSalesRefundOnly,
		Reason:             "餐品漏送",
		RequestedAmountFen: order.AmountFen,
		Status:             AfterSalesRefunded,
		ReviewReason:       "确认漏送",
		ReviewerID:         "merchant_1",
		ReviewerRole:       "merchant",
		RefundID:           refund.ID,
		CreatedAt:          reviewedAt.Add(-30 * time.Minute),
		UpdatedAt:          reviewedAt,
		ReviewedAt:         reviewedAt,
	}
	store.replacePaymentDomainFromTables(
		[]Order{order},
		[]WalletAccount{{UserID: "user_1", Balance: 1200, Version: 2, RiskState: "normal"}},
		nil,
		nil,
		nil,
	)
	store.replaceRefundDomainFromTables(RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst}, []RefundTransaction{refund})
	store.replaceAfterSalesDomainFromTables([]AfterSalesRequest{request})

	reviewed, replayedRefund, replayedOrder, account, err := store.afterSalesReviewResult(request.ID, refund.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Status != AfterSalesRefunded || reviewed.RefundID != refund.ID || replayedRefund.ID != refund.ID || replayedOrder.Status != StatusRefunded || account.Balance != 1200 {
		t.Fatalf("expected restored after-sales review result, request=%+v refund=%+v order=%+v account=%+v", reviewed, replayedRefund, replayedOrder, account)
	}
}

func TestSQLAfterSalesReviewResultRestoresPartialRefundWithoutClosingOrder(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	reviewedAt := time.Date(2026, 5, 22, 16, 30, 0, 0, time.UTC)
	idempotencyKey := "after_sales:asr_partial_sql"
	order := Order{
		ID:            "ord_after_sales_partial_sql",
		UserID:        "user_1",
		ShopID:        "shop_1",
		Type:          OrderTypeTakeout,
		Status:        StatusMerchantPending,
		PaymentMethod: PaymentBalance,
		AmountFen:     1200,
		CreatedAt:     reviewedAt.Add(-time.Hour),
		UpdatedAt:     reviewedAt,
	}
	refund := RefundTransaction{
		ID:             "rfd_" + shortHash(idempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      500,
		Destination:    RefundDestinationBalance,
		Status:         RefundStatusSuccess,
		Reason:         "确认少送",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      reviewedAt,
	}
	request := AfterSalesRequest{
		ID:                 "asr_partial_sql",
		OrderID:            order.ID,
		UserID:             order.UserID,
		Type:               AfterSalesPartialRefund,
		Reason:             "少送一杯饮品",
		RequestedAmountFen: 500,
		Status:             AfterSalesRefunded,
		ReviewReason:       "确认少送",
		ReviewerID:         "merchant_1",
		ReviewerRole:       "merchant",
		RefundID:           refund.ID,
		CreatedAt:          reviewedAt.Add(-30 * time.Minute),
		UpdatedAt:          reviewedAt,
		ReviewedAt:         reviewedAt,
	}
	store.replacePaymentDomainFromTables(
		[]Order{order},
		[]WalletAccount{{UserID: "user_1", Balance: 500, Version: 2, RiskState: "normal"}},
		nil,
		nil,
		nil,
	)
	store.replaceRefundDomainFromTables(RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst}, []RefundTransaction{refund})
	store.replaceAfterSalesDomainFromTables([]AfterSalesRequest{request})

	reviewed, replayedRefund, replayedOrder, account, err := store.afterSalesReviewResult(request.ID, refund.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Type != AfterSalesPartialRefund || replayedRefund.AmountFen != 500 || replayedOrder.Status != StatusMerchantPending || account.Balance != 500 {
		t.Fatalf("expected restored partial after-sales refund to keep order active, request=%+v refund=%+v order=%+v account=%+v", reviewed, replayedRefund, replayedOrder, account)
	}
}

func TestSQLMerchantOrderEventSideEffectsEnqueueStatusOutbox(t *testing.T) {
	store := NewStore(DefaultHomeModules())
	createdAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	acceptedAt := createdAt.Add(time.Minute)
	readyAt := acceptedAt.Add(2 * time.Minute)
	order := Order{
		ID:            "ord_sql_merchant_event",
		UserID:        "user_1",
		ShopID:        "shop_1",
		Type:          OrderTypeTakeout,
		Status:        StatusPreparing,
		PaymentMethod: PaymentBalance,
		AmountFen:     3200,
		CreatedAt:     createdAt,
		UpdatedAt:     acceptedAt,
		Events: []OrderEvent{{
			Type:      "merchant.accepted",
			ActorID:   "merchant_1",
			Message:   "商户已接单，开始备餐",
			CreatedAt: acceptedAt,
		}},
	}
	store.replacePaymentDomainFromTables([]Order{order}, nil, nil, nil, nil)
	store.applyOrderEventOutboxAfterSQL(order.ID, order.Events[0])

	readyOrder := order
	readyOrder.Status = StatusDispatching
	readyOrder.UpdatedAt = readyAt
	readyEvent := OrderEvent{
		Type:      "merchant.ready_for_pickup",
		ActorID:   "merchant_1",
		Message:   "商户已出餐，订单进入骑手调度",
		CreatedAt: readyAt,
	}
	readyOrder.Events = append(readyOrder.Events, readyEvent)
	store.replacePaymentDomainFromTables([]Order{readyOrder}, nil, nil, nil, nil)
	store.applyOrderEventOutboxAfterSQL(order.ID, readyEvent)

	events, err := store.OutboxEvents(OutboxEventsRequest{Topic: "order.status_changed", Limit: 10, Now: readyAt.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("expected accepted and ready outbox events, got %+v", events)
	}
	if events[0].EventType != "merchant.accepted" || events[0].AggregateID != order.ID || events[0].Payload["status"] != StatusPreparing {
		t.Fatalf("expected merchant.accepted status outbox, got %+v", events[0])
	}
	if events[1].EventType != "merchant.ready_for_pickup" || events[1].AggregateID != order.ID || events[1].Payload["status"] != StatusDispatching {
		t.Fatalf("expected merchant.ready_for_pickup status outbox, got %+v", events[1])
	}
	if _, err := store.GrabOrder(order.ID, "rider_1"); err != nil {
		t.Fatalf("expected restored SQL ready order to be grabbable, got %v", err)
	}
}
