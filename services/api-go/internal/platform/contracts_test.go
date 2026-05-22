package platform

import (
	"testing"
	"time"
)

func TestFulfillmentModeForOrderType(t *testing.T) {
	cases := map[string]string{
		OrderTypeTakeout:   FulfillmentRiderDelivery,
		OrderTypeMedicine:  FulfillmentRiderDelivery,
		OrderTypeCourier:   FulfillmentPlatformErrand,
		OrderTypeErrandBuy: FulfillmentPlatformErrand,
		OrderTypeGroupbuy:  FulfillmentInStoreRedeem,
	}
	for orderType, expected := range cases {
		if actual := FulfillmentModeForOrderType(orderType); actual != expected {
			t.Fatalf("expected %s for %s, got %s", expected, orderType, actual)
		}
	}
}

func TestMerchantOnboardingConstants(t *testing.T) {
	if MerchantRegistrationAdminInviteOnly != "admin_invite_only" {
		t.Fatalf("merchant registration must be admin invite only")
	}
	if QualificationBusinessLicense != "business_license" || QualificationHealthCertificate != "health_certificate" {
		t.Fatalf("merchant qualification constants must include business license and health certificate")
	}
	if ShopStatusQualificationExpired != "qualification_expired" {
		t.Fatalf("shop must support qualification expired closure status")
	}
}

func TestDispatchAndDepositConstants(t *testing.T) {
	if DispatchModeForOrderAgeSeconds(599) != DispatchModeGrabHall {
		t.Fatal("orders younger than ten minutes must stay in grab hall")
	}
	if DispatchModeForOrderAgeSeconds(600) != DispatchModeAutoAssign {
		t.Fatal("orders at ten minutes or older must auto assign")
	}
	if !RiderCanAcceptOrders(
		RiderAccount{Type: RiderAccountRider, Status: "active", Online: true},
		DepositAccount{SubjectType: "rider", AmountFen: RiderDepositAmountFen, Status: DepositStatusPaid},
	) {
		t.Fatal("paid rider deposit must allow order acceptance")
	}
	if !RiderCanAcceptOrders(
		RiderAccount{Type: RiderAccountRider, Status: "active", Online: true},
		DepositAccount{SubjectType: "rider", Status: DepositStatusWechatExemptApproved},
	) {
		t.Fatal("wechat exempt rider must allow order acceptance")
	}
	if MerchantCanAcceptOrders(DepositAccount{SubjectType: "merchant", AmountFen: MerchantDepositAmountFen, Status: DepositStatusWechatExemptApproved}) {
		t.Fatal("merchant must not accept orders with wechat exempt deposit")
	}
	if !MerchantCanAcceptOrders(DepositAccount{SubjectType: "merchant", AmountFen: MerchantDepositAmountFen, Status: DepositStatusPaid}) {
		t.Fatal("paid merchant deposit must allow order acceptance")
	}
}

func TestMerchantRefundGroupAndRedPacketConstants(t *testing.T) {
	if RiderLevelS != "S" || RiderLevelC != "C" {
		t.Fatal("rider levels must include S through C")
	}
	if RiderDispatchPriority(RiderLevelS) <= RiderDispatchPriority(RiderLevelA) {
		t.Fatal("higher rider level must receive higher dispatch priority")
	}
	if !RiderCanDeclineDispatchWithoutPenalty(12, 12) || RiderCanDeclineDispatchWithoutPenalty(11, 12) {
		t.Fatal("fixed daily order count must gate penalty-free dispatch decline")
	}
	if ProductStatusSoldOut != "sold_out" {
		t.Fatal("product sold out status must be stable")
	}
	if StatusVoucherIssued != "voucher_issued" || GroupbuyRedemptionMethodQR != "qr_scan" {
		t.Fatal("groupbuy voucher issuance and qr redemption constants must be stable")
	}
	if RefundDestinationBalance != "balance" || RefundStrategyOriginalFirst != "original_route_first" {
		t.Fatal("refund destination constants must be stable")
	}
	if RefundDestinationForStrategy("", "") != RefundDestinationBalance {
		t.Fatal("default refund strategy must prefer balance")
	}
	refundRequired, destination := GroupbuyUnavailableRefund(
		MerchantProduct{Status: ProductStatusSoldOut, StockCount: 0},
		GroupbuyVoucherStatusIssued,
		RefundStrategyBalanceFirst,
	)
	if !refundRequired || destination != RefundDestinationBalance {
		t.Fatal("unavailable groupbuy voucher must refund to configured destination")
	}
	if WalletPaymentPasswordSet != "set" {
		t.Fatal("wallet payment password must support set status")
	}
	if GroupChatOfficial != "official" || GroupNotifyMuted != "muted" {
		t.Fatal("official group defaults must be stable")
	}
	if CouponRequirementGroupMembership != "group_membership" {
		t.Fatal("coupon group membership requirement must be stable")
	}
	if RedPacketSceneDirectMessage != "direct_message" || RedPacketTypeRandom != "random" {
		t.Fatal("red packet constants must support direct and random packets")
	}
}

func TestCircleMealMatchCouponAndHomeCards(t *testing.T) {
	modules := DefaultHomeModules()
	if modules[4].Key != HomeModuleCircle || !modules[4].Enabled {
		t.Fatal("home modules must include enabled circle entry")
	}
	if modules[5].Key != HomeModuleCharity || modules[5].Enabled {
		t.Fatal("legacy charity entry must remain backend configurable but disabled by default")
	}
	cards := DefaultHomeCards()
	if len(cards) < 2 || cards[1].Type != HomeCardCircle {
		t.Fatal("home cards must include admin-controlled circle card")
	}
	ok, missing := CanUseMealMatch(MealMatchProfile{Gender: "female"})
	if ok || len(missing) != 3 {
		t.Fatalf("meal match must require truth release and questionnaire, got ok=%v missing=%v", ok, missing)
	}
	ok, missing = CanUseMealMatch(MealMatchProfile{
		UserID:                         "u1",
		Gender:                         "female",
		IdentityTruthSigned:            true,
		PlatformLiabilityReleaseSigned: true,
		QuestionnaireCompleted:         true,
		PersonalityTraits:              []string{"准时"},
		DietaryHabits:                  []string{"不吃辣"},
	})
	if !ok || len(missing) != 0 {
		t.Fatalf("complete meal match profile must pass, got ok=%v missing=%v", ok, missing)
	}

	platformCoupon := CouponPolicyFromInput(CouponPolicy{IssuerType: CouponIssuerPlatform, CostBearer: CouponCostBearerPlatform})
	if !platformCoupon.SubsidySettlementRequired || platformCoupon.CostBearer != CouponCostBearerPlatform {
		t.Fatal("platform-funded coupon must require platform subsidy settlement")
	}
	merchantCampaign := CouponPolicy{
		IssuerType:               CouponIssuerPlatform,
		CostBearer:               CouponCostBearerMerchant,
		ScopeType:                CouponScopeParticipatingShops,
		MerchantAcceptanceStatus: CouponActivityPending,
		ParticipatingShopIDs:     []string{"shop_1"},
	}
	if CouponCanApplyToShop(merchantCampaign, "shop_1") {
		t.Fatal("merchant-borne platform activity must require merchant acceptance before use")
	}
	merchantCampaign.MerchantAcceptanceStatus = CouponActivityAccepted
	if !CouponCanApplyToShop(merchantCampaign, "shop_1") {
		t.Fatal("accepted merchant-borne platform activity must apply to participating shops")
	}
}

func TestLegacyParityContracts(t *testing.T) {
	lat := 39.9
	lng := 116.4
	if !UserAddressReady(UserAddress{
		ContactName:  "张三",
		ContactPhone: "13800000000",
		City:         "北京",
		Detail:       "望京",
		Latitude:     &lat,
		Longitude:    &lng,
	}) {
		t.Fatal("complete delivery address must be usable")
	}
	if UserAddressReady(UserAddress{ContactName: "张三"}) {
		t.Fatal("incomplete delivery address must be rejected")
	}
	if OrderPayableFen(3000, 500, 200, 800) != 2900 {
		t.Fatal("order payable must include item delivery packaging minus discount")
	}
	if !CanCreateAfterSales(
		Order{ID: "o1", Status: StatusCompleted, AmountFen: 1200},
		AfterSalesRequest{OrderID: "o1", Reason: "餐品异常", RequestedAmountFen: 800},
	) {
		t.Fatal("completed order with valid reason and amount must support after sales")
	}
	if CanCreateAfterSales(
		Order{ID: "o1", Status: StatusRefunded, AmountFen: 1200},
		AfterSalesRequest{OrderID: "o1", Reason: "重复售后", RequestedAmountFen: 800},
	) {
		t.Fatal("refunded order must not create duplicate after sales")
	}
	deadline := time.Date(2026, 5, 21, 8, 30, 0, 0, time.UTC)
	deliveredAt := time.Date(2026, 5, 21, 8, 35, 0, 0, time.UTC)
	if DeliveryPromiseStatus(deadline, deliveredAt, false) != DeliveryPromiseTimeout {
		t.Fatal("late delivery must be timeout unless exempt")
	}
	if ResolveMembershipTier(12000) != MembershipBlackGold {
		t.Fatal("growth value must resolve membership tier")
	}
	if !RiskDecisionBlocked([]RiskEvent{{Type: RiskEventFakeTransaction}}, 7, 3) {
		t.Fatal("fake transaction must block ordering")
	}
}
