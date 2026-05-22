import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import test from "node:test";

const root = new URL("..", import.meta.url).pathname;

const requiredPaths = [
  "apps/user-wechat-miniprogram",
  "apps/merchant-uni",
  "apps/rider-uni",
  "apps/admin-web",
  "apps/admin-uni",
  "services/api-go",
  "services/bff",
  "services/realtime-gateway",
  "services/dispatch-worker",
  "services/payment-worker",
  "services/notification-worker",
  "services/integration-worker",
  "services/object-lifecycle-worker",
  "services/settlement-worker",
  "services/outbox-relay-worker",
  "packages/contracts",
  "packages/design-tokens",
  "packages/domain-core",
  "packages/client-sdk",
  "packages/admin-core",
  "infra/docker",
  "infra/db/migrations/0001_core.sql",
  "infra/db/migrations/0002_auth_payment.sql",
  "infra/db/migrations/0003_platform_store_snapshots.sql",
  "infra/db/migrations/0004_platform_outbox.sql",
  "infra/k8s/base",
  "infra/observability",
  "infra/loadtest",
  "assets/brand/logo.svg",
  "PLATFORM_MASTER_PLAN.md",
  "EXECUTION_LEDGER.md",
  "PROJECT_STATUS.md",
  ".github/workflows/verify.yml",
  ".github/pull_request_template.md",
  ".github/ISSUE_TEMPLATE/bug_report.yml",
  ".github/ISSUE_TEMPLATE/feature_request.yml",
  ".github/ISSUE_TEMPLATE/commercial_gap.yml",
  ".github/CODEOWNERS",
  ".github/dependabot.yml",
  "apps/admin-web/package.json",
  "apps/admin-web/index.html",
  "apps/admin-web/src/adminApi.mjs",
  "apps/admin-web/src/adminViews.mjs",
  "apps/admin-web/src/config.mjs",
  "apps/admin-web/src/styles.css"
];

test("architecture directories and governance files exist", () => {
  for (const relativePath of requiredPaths) {
    assert.equal(existsSync(join(root, relativePath)), true, `${relativePath} must exist`);
  }
});

test("master plan records the selected architecture", () => {
  const plan = readFileSync(join(root, "PLATFORM_MASTER_PLAN.md"), "utf8");
  assert.match(plan, /模块化核心 API \+ 事件驱动 Worker \+ 多端 BFF \+ 实时网关架构/);
  assert.match(plan, /自建\/混合云/);
  assert.match(plan, /10 万/);
});

test("github collaboration gates protect commercial readiness", () => {
  const workflow = readFileSync(join(root, ".github/workflows/verify.yml"), "utf8");
  const prTemplate = readFileSync(join(root, ".github/pull_request_template.md"), "utf8");
  const commercialGap = readFileSync(join(root, ".github/ISSUE_TEMPLATE/commercial_gap.yml"), "utf8");
  const codeowners = readFileSync(join(root, ".github/CODEOWNERS"), "utf8");
  const dependabot = readFileSync(join(root, ".github/dependabot.yml"), "utf8");
  assert.match(workflow, /npm run verify/);
  assert.match(workflow, /go test -count=1 \.\/\.\.\./);
  assert.match(workflow, /actions\/setup-node@v4/);
  assert.match(workflow, /actions\/setup-go@v5/);
  assert.match(prTemplate, /Commercial Impact/);
  assert.match(prTemplate, /Rollback Notes/);
  assert.match(commercialGap, /Commercial readiness gap/);
  assert.match(commercialGap, /Required proof/);
  assert.match(codeowners, /@HCRXchenghong/);
  assert.match(dependabot, /github-actions/);
  assert.match(dependabot, /gomod/);
});

test("admin web has a minimum operable control center", () => {
  const main = readFileSync(join(root, "apps/admin-web/src/main.js"), "utf8");
  const api = readFileSync(join(root, "apps/admin-web/src/adminApi.mjs"), "utf8");
  const views = readFileSync(join(root, "apps/admin-web/src/adminViews.mjs"), "utf8");
  const config = readFileSync(join(root, "apps/admin-web/src/config.mjs"), "utf8");
  const styles = readFileSync(join(root, "apps/admin-web/src/styles.css"), "utf8");
  assert.match(main, /executeAdminOperation/);
  assert.match(main, /renderModuleView/);
  assert.match(main, /运营后台/);
  assert.match(api, /\/api\/auth\/admin\/login/);
  assert.match(api, /\/api\/admin\/merchant-invites/);
  assert.match(api, /\/api\/admin\/operations\/snapshot/);
  assert.match(api, /\/api\/admin\/outbox\/stats/);
  assert.match(api, /\/api\/admin\/object-storage\/cleanup-stats/);
  assert.match(api, /\/api\/station-manager\/rider-performance/);
  assert.match(views, /订单监控/);
  assert.match(views, /商户资质/);
  assert.match(views, /骑手\/站长/);
  assert.match(views, /售后审核/);
  assert.match(views, /订单状态补偿必须写审计/);
  assert.match(config, /ADMIN_WEB_RBAC/);
  assert.match(config, /refund-settings/);
  assert.match(config, /rtc/);
  assert.match(styles, /#009bf5/);
});

test("core database migration records commercial-grade ledgers and events", () => {
  const schema = [
    readFileSync(join(root, "infra/db/migrations/0001_core.sql"), "utf8"),
    readFileSync(join(root, "infra/db/migrations/0002_auth_payment.sql"), "utf8"),
    readFileSync(join(root, "infra/db/migrations/0003_platform_store_snapshots.sql"), "utf8"),
    readFileSync(join(root, "infra/db/migrations/0004_platform_outbox.sql"), "utf8")
  ].join("\n");
  for (const tableName of [
    "app_users",
    "auth_identities",
    "merchant_accounts",
    "shops",
    "merchant_products",
    "user_addresses",
    "wallet_accounts",
    "wallet_transactions",
    "refund_settings",
    "refund_transactions",
    "order_after_sales",
    "order_after_sales_events",
    "order_after_sales_evidence_upload_tickets",
    "order_after_sales_evidence",
    "cart_items",
    "orders",
    "order_items",
    "order_events",
    "dispatch_jobs",
    "dispatch_events",
    "rider_free_cancel_usage",
    "messages",
    "auth_sessions",
    "platform_sequences",
    "wallet_payment_passwords",
    "payment_transactions",
    "payment_callbacks",
    "audit_logs",
    "platform_store_snapshots",
    "platform_outbox_events",
    "platform_consumed_events"
  ]) {
    assert.match(schema, new RegExp(`CREATE TABLE ${tableName}\\b`), `${tableName} table must exist`);
  }
  assert.match(schema, /UNIQUE \(idempotency_key\)/);
  assert.match(schema, /CREATE UNIQUE INDEX uniq_dispatch_events_idempotency_type/);
  assert.match(schema, /ON dispatch_events \(idempotency_key, type\)/);
  assert.match(schema, /PRIMARY KEY \(rider_id, business_date\)/);
  assert.match(schema, /CHECK \(status IN \('pending', 'published', 'failed', 'dead_letter'\)\)/);
  assert.match(schema, /lease_owner TEXT NOT NULL DEFAULT ''/);
  assert.match(schema, /lease_expires_at TIMESTAMPTZ/);
  assert.match(schema, /CREATE INDEX idx_platform_outbox_lease/);
  assert.match(schema, /consumer_event_key TEXT PRIMARY KEY/);
  assert.match(schema, /UNIQUE \(consumer_name, idempotency_key\)/);
  assert.match(schema, /CHECK \(status IN \('processing', 'processed', 'failed'\)\)/);
  assert.match(schema, /CREATE INDEX idx_platform_consumed_events_topic/);
});

test("postgres store uses normalized outbox table for relay concurrency", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  assert.match(postgresStore, /func \(s \*PostgresStore\) ensureOutboxTable/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS platform_outbox_events/);
  assert.match(postgresStore, /ON CONFLICT \(idempotency_key\) DO NOTHING/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) OutboxEvents/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) ClaimOutboxEvents/);
  assert.match(postgresStore, /FOR UPDATE SKIP LOCKED/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) RenewOutboxEventLease/);
  assert.match(postgresStore, /event\.lease_owner = \$2/);
  assert.match(postgresStore, /event\.lease_expires_at > \$3/);
});

test("postgres store normalizes payment and wallet domain tables", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(postgresStore, /func \(s \*PostgresStore\) ensurePaymentDomainTables/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS wallet_accounts/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS wallet_transactions/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS wallet_payment_passwords/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS payment_transactions/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) syncPaymentDomainToTables/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadPaymentDomainFromTables/);
  assert.match(postgresStore, /func \(s \*Store\) paymentDomainSnapshot/);
  assert.match(postgresStore, /func \(s \*Store\) replacePaymentDomainFromTables/);
  assert.match(postgresStore, /upsertSQLWalletTransaction/);
  assert.match(postgresStore, /upsertSQLPaymentTransaction/);
  assert.match(postgresStoreTest, /TestPaymentDomainTableRestoreRebuildsPaymentAndWalletIndexes/);
});

test("postgres store uses transactional SQL order creation", () => {
  const coreSchema = readFileSync(join(root, "infra/db/migrations/0001_core.sql"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(coreSchema, /CREATE TABLE platform_sequences\b/);
  assert.match(coreSchema, /next_value BIGINT NOT NULL CHECK \(next_value >= 1\)/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS platform_sequences/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) createOrderInSQL/);
  assert.match(postgresStore, /BeginTx\(ctx, &sql\.TxOptions\{Isolation: sql\.LevelReadCommitted\}\)/);
  assert.match(postgresStore, /func ensureSQLUserPlaceholder/);
  assert.match(postgresStore, /ON CONFLICT \(id\) DO NOTHING/);
  assert.match(postgresStore, /func nextSQLOrderNumber/);
  assert.match(postgresStore, /FROM platform_sequences[\s\S]*WHERE name = 'orders'[\s\S]*FOR UPDATE/);
  assert.match(postgresStore, /MAX\(\(substring\(id FROM '\^ord_\(\[0-9\]\+\)\$'\)\)::bigint\)/);
  assert.match(postgresStore, /func insertSQLCreatedOrder/);
  assert.match(postgresStore, /INSERT INTO orders/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) CreateOrder/);
  assert.match(postgresStore, /s\.createOrderInSQL\(ctx, req\)/);
  assert.match(postgresStore, /s\.reloadPaymentDomainAndOutboxAfterSQLOrderEvent\(ctx, sqlOrder\.ID, event\)/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) CreateOrder\(req CreateOrderRequest\)[\s\S]{0,320}s\.Store\.CreateOrder/);
  assert.match(postgresStoreTest, /TestSQLCreateOrderSideEffectsRestoreMirror/);
});

test("postgres store uses transactional SQL cart checkout", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS merchant_products/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS cart_items/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS merchant_qualifications/);
  assert.match(postgresStore, /func upsertSQLMerchantQualification/);
  assert.match(postgresStore, /func upsertSQLMerchantProduct/);
  assert.match(postgresStore, /func upsertSQLCartItem/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) checkoutCartInSQL/);
  assert.match(postgresStore, /pg_advisory_xact_lock\(hashtext\(\$1\)::bigint\)/);
  assert.match(postgresStore, /func loadSQLCheckoutShopForUpdate/);
  assert.match(postgresStore, /func sqlMerchantQualificationsReady/);
  assert.match(postgresStore, /expires_at > now\(\)/);
  assert.match(postgresStore, /FOR UPDATE OF s, m/);
  assert.match(postgresStore, /FROM cart_items ci[\s\S]*JOIN merchant_products mp[\s\S]*FOR UPDATE OF ci, mp/);
  assert.match(postgresStore, /func insertSQLCheckoutOrder/);
  assert.match(postgresStore, /INSERT INTO orders/);
  assert.match(postgresStore, /DELETE FROM cart_items[\s\S]*WHERE user_id = \$1 AND shop_id = \$2/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) reloadPaymentDomainAndOutboxAfterSQLCheckout/);
  assert.match(postgresStore, /func \(s \*Store\) clearCartAfterSQLCheckout/);
  assert.match(postgresStore, /s\.checkoutCartInSQL\(ctx, req\)/);
  assert.match(postgresStore, /s\.reloadPaymentDomainAndOutboxAfterSQLCheckout\(ctx, sqlOrder\.ID, summary\.UserID, summary\.ShopID, event\)/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) CheckoutCart\(req CheckoutCartRequest\)[\s\S]{0,360}s\.Store\.CheckoutCart/);
  assert.match(postgresStoreTest, /TestSQLCheckoutCartSideEffectsRestoreMirror/);
  assert.match(postgresStoreTest, /TestPaymentDomainSnapshotIncludesProductsAndCartForSQLCheckout/);
});

test("postgres store uses transactional wallet debit for balance payments", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(postgresStore, /func \(s \*PostgresStore\) payOrderWithBalanceInSQL/);
  assert.match(postgresStore, /BeginTx\(ctx, &sql\.TxOptions\{Isolation: sql\.LevelReadCommitted\}\)/);
  assert.match(postgresStore, /pg_advisory_xact_lock\(hashtext\(\$1\)::bigint\)/);
  assert.match(postgresStore, /FROM wallet_transactions[\s\S]*WHERE idempotency_key = \$1[\s\S]*FOR UPDATE/);
  assert.match(postgresStore, /FROM orders[\s\S]*WHERE id = \$1[\s\S]*FOR UPDATE/);
  assert.match(postgresStore, /FROM wallet_accounts[\s\S]*WHERE subject_type = 'user' AND subject_id = \$1[\s\S]*FOR UPDATE/);
  assert.match(postgresStore, /ON CONFLICT \(idempotency_key\) DO NOTHING/);
  assert.match(postgresStore, /walletTransactionMatchesBalancePayment/);
  assert.match(postgresStore, /applyBalancePaymentSideEffectsAfterSQL/);
  assert.match(postgresStore, /reloadPaymentDomainAndOutboxAfterSQLPayment/);
  assert.match(postgresStore, /s\.payOrderWithBalanceInSQL\(ctx, req\)/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) PayOrderWithBalance\(req BalancePayRequest\)[\s\S]{0,220}s\.Store\.PayOrderWithBalance/);
  assert.match(postgresStoreTest, /TestSQLBalancePaymentSideEffectsIssueGroupbuyVoucherAndOutbox/);
  assert.match(postgresStoreTest, /TestSQLBalancePaymentIdempotencyMatcherRejectsCrossOrderReplay/);
});

test("refund settings and admin order refund are wired end to end", () => {
  const contracts = readFileSync(join(root, "services/api-go/internal/platform/contracts.go"), "utf8");
  const repository = readFileSync(join(root, "services/api-go/internal/platform/repository.go"), "utf8");
  const store = readFileSync(join(root, "services/api-go/internal/platform/store.go"), "utf8");
  const objectStorage = readFileSync(join(root, "services/api-go/internal/platform/object_storage.go"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const apiMain = readFileSync(join(root, "services/api-go/cmd/api/main.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  const router = readFileSync(join(root, "services/api-go/internal/httpapi/router.go"), "utf8");
  const routerTest = readFileSync(join(root, "services/api-go/internal/httpapi/router_test.go"), "utf8");
  const bff = readFileSync(join(root, "services/bff/src/server.mjs"), "utf8");
  const bffTest = readFileSync(join(root, "services/bff/src/runtime.test.mjs"), "utf8");
  assert.match(contracts, /type RefundSettings struct/);
  assert.match(contracts, /type RefundOrderRequest struct/);
  assert.match(contracts, /type RefundTransaction struct/);
  assert.match(contracts, /AmountFen int64\s+`json:"amount_fen,omitempty"`/);
  assert.match(repository, /RefundSettings\(\) \(\*RefundSettings, error\)/);
  assert.match(repository, /RefundOrder\(req RefundOrderRequest\) \(\*RefundTransaction, \*Order, \*WalletAccount, error\)/);
  assert.match(store, /func \(s \*Store\) RefundOrder\(req RefundOrderRequest\)/);
  assert.match(store, /func \(s \*Store\) refundedAmountForOrderLocked/);
  assert.match(store, /func refundOrderStatusAfter/);
  assert.match(store, /order_amount_fen/);
  assert.match(store, /order\.refund\.success/);
  assert.match(store, /payment\.refund\.requested/);
  assert.match(postgresStore, /RefundTransactions\s+map\[string\]\*RefundTransaction/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) RefundOrder\(req RefundOrderRequest\)/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS refund_settings/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS refund_transactions/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) refundOrderInSQL/);
  assert.match(postgresStore, /func refundedAmountForSQLOrder/);
  assert.match(postgresStore, /FROM refund_transactions[\s\S]*WHERE idempotency_key = \$1[\s\S]*FOR UPDATE/);
  assert.match(postgresStore, /func insertSQLRefundTransaction/);
  assert.match(postgresStore, /func insertSQLWalletRefundTransaction/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) reloadPaymentDomainAndOutboxAfterSQLRefund/);
  assert.match(postgresStore, /s\.refundOrderInSQL\(ctx, req\)/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) RefundOrder\(req RefundOrderRequest\)[\s\S]{0,260}s\.Store\.RefundOrder/);
  assert.match(router, /GET \/api\/admin\/refund-settings/);
  assert.match(router, /PUT \/api\/admin\/refund-settings/);
  assert.match(router, /POST \/api\/orders\/\{orderID\}\/refund/);
  assert.match(router, /handleAdminRefundOrder/);
  assert.match(bff, /\/api\/admin\/refund-settings/);
  assert.match(bff, /\^\\\/api\\\/orders\\\/\[\^\/\]\+\\\/refund\$/);
  assert.match(routerTest, /TestAdminRefundSettingsAndOrderRefundHTTPFlow/);
  assert.match(postgresStoreTest, /TestSQLRefundSideEffectsRestoreRefundAndOutbox/);
  assert.match(bffTest, /refundedOrder/);
});

test("payment worker understands original-route refund outbox events", () => {
  const paymentWorker = readFileSync(join(root, "services/payment-worker/src/index.mjs"), "utf8");
  const paymentWorkerTest = readFileSync(join(root, "services/payment-worker/src/index.test.mjs"), "utf8");
  assert.match(paymentWorker, /payment\.refund\.requested/);
  assert.match(paymentWorker, /export function normalizePaymentRefundRequest/);
  assert.match(paymentWorker, /type: "refund_requested"/);
  assert.match(paymentWorker, /pending_original_route/);
  assert.match(paymentWorker, /createIdempotentConsumer/);
  assert.match(paymentWorkerTest, /payment refund request normalization creates original-route task payload/);
});

test("after-sales application and review flow is wired end to end", () => {
  const contracts = readFileSync(join(root, "services/api-go/internal/platform/contracts.go"), "utf8");
  const repository = readFileSync(join(root, "services/api-go/internal/platform/repository.go"), "utf8");
  const store = readFileSync(join(root, "services/api-go/internal/platform/store.go"), "utf8");
  const objectStorage = readFileSync(join(root, "services/api-go/internal/platform/object_storage.go"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const apiMain = readFileSync(join(root, "services/api-go/cmd/api/main.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  const router = readFileSync(join(root, "services/api-go/internal/httpapi/router.go"), "utf8");
  const routerTest = readFileSync(join(root, "services/api-go/internal/httpapi/router_test.go"), "utf8");
  const bff = readFileSync(join(root, "services/bff/src/server.mjs"), "utf8");
  const bffTest = readFileSync(join(root, "services/bff/src/runtime.test.mjs"), "utf8");
  assert.match(contracts, /type CreateAfterSalesRequest struct/);
  assert.match(contracts, /type AfterSalesEvent struct/);
  assert.match(contracts, /type AfterSalesEvidence struct/);
  assert.match(contracts, /type AfterSalesEvidenceUploadTicket struct/);
  assert.match(contracts, /type AddAfterSalesEventRequest struct/);
  assert.match(contracts, /type CreateAfterSalesEvidenceUploadRequest struct/);
  assert.match(contracts, /type ConfirmAfterSalesEvidenceUploadRequest struct/);
  assert.match(contracts, /type ObjectStorageUploadCallbackRequest struct/);
  assert.match(contracts, /type ObjectStorageScanResultRequest struct/);
  assert.match(contracts, /type ObjectUploadTicket struct/);
  assert.match(contracts, /type ReviewAfterSalesRequest struct/);
  assert.match(contracts, /AfterSalesPartialRefund/);
  assert.match(contracts, /AfterSalesActionCustomerCare/);
  assert.match(contracts, /AfterSalesEvidenceMaxBytes/);
  assert.match(contracts, /AfterSalesUploadTicketIssued/);
  assert.match(contracts, /AfterSalesUploadTicketUploaded/);
  assert.match(contracts, /AfterSalesUploadScanPassed/);
  assert.match(contracts, /TicketID\s+string\s+`json:"ticket_id/);
  assert.match(contracts, /AfterSalesDecisionApprove/);
  assert.match(repository, /CreateAfterSales\(req CreateAfterSalesRequest\) \(\*AfterSalesRequest, error\)/);
  assert.match(repository, /AfterSalesEvents\(requestID string, actorID string, actorRole string\) \(\[\]AfterSalesEvent, error\)/);
  assert.match(repository, /AddAfterSalesEvent\(req AddAfterSalesEventRequest\) \(\*AfterSalesEvent, \*AfterSalesRequest, error\)/);
  assert.match(repository, /CreateAfterSalesEvidenceUpload\(req CreateAfterSalesEvidenceUploadRequest\) \(\*ObjectUploadTicket, error\)/);
  assert.match(repository, /ConfirmObjectStorageUpload\(req ObjectStorageUploadCallbackRequest\) \(\*AfterSalesEvidenceUploadTicket, error\)/);
  assert.match(repository, /RecordObjectStorageScanResult\(req ObjectStorageScanResultRequest\) \(\*AfterSalesEvidenceUploadTicket, error\)/);
  assert.match(repository, /ConfirmAfterSalesEvidenceUpload\(req ConfirmAfterSalesEvidenceUploadRequest\) \(\*AfterSalesEvidence, \*AfterSalesEvent, \*AfterSalesRequest, error\)/);
  assert.match(repository, /AfterSalesEvidence\(requestID string, actorID string, actorRole string\) \(\[\]AfterSalesEvidence, error\)/);
  assert.match(repository, /ReviewAfterSales\(req ReviewAfterSalesRequest\) \(\*AfterSalesRequest, \*RefundTransaction, \*Order, \*WalletAccount, error\)/);
  assert.match(store, /func \(s \*Store\) CreateAfterSales/);
  assert.match(store, /func \(s \*Store\) AddAfterSalesEvent/);
  assert.match(store, /func \(s \*Store\) AfterSalesEvents/);
  assert.match(store, /func \(s \*Store\) CreateAfterSalesEvidenceUpload/);
  assert.match(store, /func \(s \*Store\) ConfirmObjectStorageUpload/);
  assert.match(store, /func \(s \*Store\) RecordObjectStorageScanResult/);
  assert.match(store, /func \(s \*Store\) ConfirmAfterSalesEvidenceUpload/);
  assert.match(store, /prepareAfterSalesEvidenceUploadConfirmation/);
  assert.match(store, /verifyUploadedObject/);
  assert.match(store, /afterSalesUploadTickets/);
  assert.match(store, /func \(s \*Store\) afterSalesUploadTicketForConfirmLocked/);
  assert.match(store, /afterSalesUploadTicketMatchesConfirm/);
  assert.match(store, /afterSalesUploadTicketConfirmReady/);
  assert.match(store, /func \(s \*Store\) AfterSalesEvidence/);
  assert.match(store, /AfterSalesActionEvidenceUploaded/);
  assert.match(store, /func \(s \*Store\) afterSalesRequestViewLocked/);
  assert.match(store, /refundedAmountForOrderLocked/);
  assert.match(store, /func \(s \*Store\) appendAfterSalesEventLocked/);
  assert.match(store, /func \(s \*Store\) ReviewAfterSales/);
  assert.match(store, /AfterSalesPartialRefund/);
  assert.match(store, /refundableRemainingFenLocked/);
  assert.match(store, /s\.refundOrderLocked\(RefundOrderRequest/);
  assert.match(store, /order\.after_sales\.created/);
  assert.match(store, /order\.after_sales\.evidence_uploaded/);
  assert.match(store, /order\.after_sales\.approved/);
  assert.match(objectStorage, /type ObjectStorageConfig struct/);
  assert.match(objectStorage, /RequireHeadVerification\s+bool/);
  assert.match(objectStorage, /RequireUploadCallbackForConfirm\s+bool/);
  assert.match(objectStorage, /RequireScanApprovalForConfirm\s+bool/);
  assert.match(objectStorage, /func NormalizeObjectStorageConfig/);
  assert.match(objectStorage, /func \(s \*Store\) ConfigureObjectStorage/);
  assert.match(objectStorage, /func \(config ObjectStorageConfig\) createObjectUploadTicket/);
  assert.match(objectStorage, /func \(config ObjectStorageConfig\) verifyObjectUploadCallback/);
  assert.match(objectStorage, /func \(config ObjectStorageConfig\) verifyObjectScanResult/);
  assert.match(objectStorage, /func \(config ObjectStorageConfig\) verifyUploadedObject/);
  assert.match(objectStorage, /http\.MethodHead/);
  assert.match(objectStorage, /hmac\.New\(sha256\.New/);
  assert.match(objectStorage, /X-Upload-Signature/);
  assert.match(apiMain, /OBJECT_STORAGE_UPLOAD_BASE_URL/);
  assert.match(apiMain, /OBJECT_STORAGE_PUBLIC_BASE_URL/);
  assert.match(apiMain, /OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION/);
  assert.match(apiMain, /OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK/);
  assert.match(apiMain, /OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL/);
  assert.match(apiMain, /OBJECT_STORAGE_CALLBACK_SIGNING_SECRET/);
  assert.match(apiMain, /OBJECT_STORAGE_HEAD_BASE_URL/);
  assert.match(apiMain, /OBJECT_STORAGE_SIGNING_SECRET/);
  assert.match(postgresStore, /AfterSalesRequests\s+map\[string\]\*AfterSalesRequest/);
  assert.match(postgresStore, /AfterSalesEvents\s+map\[string\]\*AfterSalesEvent/);
  assert.match(postgresStore, /AfterSalesUploadTickets\s+map\[string\]\*AfterSalesEvidenceUploadTicket/);
  assert.match(postgresStore, /AfterSalesEvidence\s+map\[string\]\*AfterSalesEvidence/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS order_after_sales/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS order_after_sales_events/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS order_after_sales_evidence_upload_tickets/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS order_after_sales_evidence/);
  assert.match(postgresStore, /func upsertSQLAfterSalesRequest/);
  assert.match(postgresStore, /func upsertSQLAfterSalesEvent/);
  assert.match(postgresStore, /func upsertSQLAfterSalesUploadTicket/);
  assert.match(postgresStore, /func upsertSQLAfterSalesEvidence/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadSQLAfterSalesRequests/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadSQLAfterSalesEvents/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadSQLAfterSalesUploadTickets/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadSQLAfterSalesEvidence/);
  assert.match(postgresStore, /func \(s \*Store\) replaceAfterSalesDomainFromTables/);
  assert.match(postgresStore, /func \(s \*Store\) replaceAfterSalesUploadTicketsFromTables/);
  assert.match(postgresStore, /func \(s \*Store\) replaceAfterSalesEvidenceFromTables/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) CreateAfterSales/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) ConfirmObjectStorageUpload/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) RecordObjectStorageScanResult/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) ReviewAfterSales/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) reviewAfterSalesInSQL/);
  assert.match(postgresStore, /FROM order_after_sales AS after_sales[\s\S]*FOR UPDATE OF after_sales, orders/);
  assert.match(postgresStore, /func createSQLRefundForAfterSalesReview/);
  assert.match(postgresStore, /refundedAmountForSQLOrder\(ctx, tx, order\.ID\)/);
  assert.match(postgresStore, /func updateSQLAfterSalesReview/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) reloadPaymentDomainAndOutboxAfterSQLAfterSales/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) ReviewAfterSales\(req ReviewAfterSalesRequest\)[\s\S]{0,260}s\.Store\.ReviewAfterSales/);
  assert.match(postgresStoreTest, /TestAfterSalesDomainSnapshotAndRestore/);
  assert.match(postgresStoreTest, /AfterSalesUploadTicketConfirmed/);
  assert.match(postgresStoreTest, /TestSQLAfterSalesReviewResultRestoresRefundAndRequest/);
  assert.match(postgresStoreTest, /TestSQLAfterSalesReviewResultRestoresPartialRefundWithoutClosingOrder/);
  assert.match(router, /POST \/api\/after-sales/);
  assert.match(router, /GET \/api\/after-sales\/\{requestID\}\/events/);
  assert.match(router, /POST \/api\/after-sales\/\{requestID\}\/events/);
  assert.match(router, /GET \/api\/after-sales\/\{requestID\}\/evidence/);
  assert.match(router, /POST \/api\/after-sales\/\{requestID\}\/evidence\/upload-ticket/);
  assert.match(router, /POST \/api\/after-sales\/\{requestID\}\/evidence\/confirm/);
  assert.match(router, /POST \/api\/object-storage\/upload-callback/);
  assert.match(router, /POST \/api\/object-storage\/scan-result/);
  assert.match(router, /GET \/api\/merchant\/after-sales/);
  assert.match(router, /GET \/api\/admin\/after-sales/);
  assert.match(router, /POST \/api\/after-sales\/\{requestID\}\/review/);
  assert.match(bff, /\/api\/after-sales/);
  assert.match(bff, /\^\\\/api\\\/after-sales\\\/\[\^\/\]\+\\\/events\$/);
  assert.match(bff, /\^\\\/api\\\/after-sales\\\/\[\^\/\]\+\\\/evidence\$/);
  assert.match(bff, /\^\\\/api\\\/after-sales\\\/\[\^\/\]\+\\\/evidence\\\/upload-ticket\$/);
  assert.match(bff, /\^\\\/api\\\/after-sales\\\/\[\^\/\]\+\\\/evidence\\\/confirm\$/);
  assert.match(bff, /\^\\\/api\\\/after-sales\\\/\[\^\/\]\+\\\/review\$/);
  assert.match(routerTest, /TestAfterSalesHTTPFlow/);
  assert.match(routerTest, /ticket_id/);
  assert.match(bffTest, /reviewedAfterSales/);
});

test("postgres store uses transactional merchant order transitions", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(postgresStore, /func \(s \*PostgresStore\) transitionMerchantOrderInSQL/);
  assert.match(postgresStore, /BeginTx\(ctx, &sql\.TxOptions\{Isolation: sql\.LevelReadCommitted\}\)/);
  assert.match(postgresStore, /func loadSQLMerchantOrderForUpdate/);
  assert.match(postgresStore, /FOR UPDATE OF orders/);
  assert.match(postgresStore, /func updateSQLMerchantOrderStatus/);
  assert.match(postgresStore, /func insertSQLOrderEvent/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) reloadPaymentDomainAndOutboxAfterSQLOrderEvent/);
  assert.match(postgresStore, /func \(s \*Store\) applyOrderEventOutboxAfterSQL/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) MerchantAcceptOrder/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) MerchantMarkOrderReady/);
  assert.match(postgresStore, /StatusMerchantPending[\s\S]*StatusPreparing[\s\S]*"merchant\.accepted"/);
  assert.match(postgresStore, /StatusPreparing[\s\S]*StatusDispatching[\s\S]*"merchant\.ready_for_pickup"/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) MerchantAcceptOrder\(orderID string, merchantID string\)[\s\S]{0,280}s\.Store\.MerchantAcceptOrder/);
  assert.doesNotMatch(postgresStore, /func \(s \*PostgresStore\) MerchantMarkOrderReady\(orderID string, merchantID string\)[\s\S]{0,280}s\.Store\.MerchantMarkOrderReady/);
  assert.match(postgresStoreTest, /TestSQLMerchantOrderEventSideEffectsEnqueueStatusOutbox/);
});

test("postgres store normalizes dispatch audit events", () => {
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const postgresStoreTest = readFileSync(join(root, "services/api-go/internal/platform/postgres_store_test.go"), "utf8");
  assert.match(postgresStore, /func \(s \*PostgresStore\) ensureDispatchEventTable/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS dispatch_events/);
  assert.match(postgresStore, /ALTER TABLE dispatch_events DROP CONSTRAINT IF EXISTS dispatch_events_idempotency_key_key/);
  assert.match(postgresStore, /CREATE UNIQUE INDEX IF NOT EXISTS uniq_dispatch_events_idempotency_type/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) syncDispatchEventsToTable/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadDispatchEventsFromTable/);
  assert.match(postgresStore, /func \(s \*Store\) dispatchEventSnapshot/);
  assert.match(postgresStore, /func \(s \*Store\) replaceDispatchEventsFromTable/);
  assert.match(postgresStore, /func upsertSQLDispatchEvent/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) loadSQLDispatchEvents/);
  assert.match(postgresStoreTest, /TestDispatchEventTableRestoreRebuildsAuditAndCompensationIndexes/);
});

test("workers use a consumed-event ledger for idempotent event handling", () => {
  const domainCore = readFileSync(join(root, "packages/domain-core/src/index.mjs"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  assert.match(domainCore, /export function normalizeOutboxConsumerEvent/);
  assert.match(domainCore, /export function createConsumedEventLedger/);
  assert.match(domainCore, /export function createIdempotentConsumer/);
  assert.match(domainCore, /DUPLICATE_CONSUMED_EVENT/);
  assert.match(postgresStore, /func \(s \*PostgresStore\) ensureConsumedEventsTable/);
  assert.match(postgresStore, /CREATE TABLE IF NOT EXISTS platform_consumed_events/);
  assert.match(postgresStore, /consumer_event_key TEXT PRIMARY KEY/);
  assert.match(postgresStore, /UNIQUE \(consumer_name, idempotency_key\)/);
  for (const [workerPath, consumerFactory] of [
    ["services/dispatch-worker/src/index.mjs", "createDispatchConsumer"],
    ["services/payment-worker/src/index.mjs", "createPaymentConsumer"],
    ["services/notification-worker/src/index.mjs", "createNotificationConsumer"],
    ["services/integration-worker/src/index.mjs", "createIntegrationConsumer"],
    ["services/object-scan-worker/src/index.mjs", "createObjectScanConsumer"],
    ["services/settlement-worker/src/index.mjs", "createSettlementConsumer"]
  ]) {
    const worker = readFileSync(join(root, workerPath), "utf8");
    assert.match(worker, /createIdempotentConsumer/);
    assert.match(worker, new RegExp(`export function ${consumerFactory}`));
  }
});

test("outbox stats expose relay lease health signals", () => {
  const contracts = readFileSync(join(root, "services/api-go/internal/platform/contracts.go"), "utf8");
  const router = readFileSync(join(root, "services/api-go/internal/httpapi/router.go"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  assert.match(contracts, /type OutboxLeaseOwnerStats struct/);
  assert.match(contracts, /LeaseExpiringWithinSeconds\s+int\s+`json:"lease_expiring_within_seconds"`/);
  assert.match(contracts, /LeaseExpiringSoon\s+int\s+`json:"lease_expiring_soon"`/);
  assert.match(contracts, /NextLeaseExpiresAt\s+time\.Time\s+`json:"next_lease_expires_at,omitempty"`/);
  assert.match(contracts, /LeaseOwners\s+\[\]OutboxLeaseOwnerStats\s+`json:"lease_owners"`/);
  assert.match(router, /lease_expiring_within_seconds/);
  assert.match(router, /LeaseExpiringWithinSeconds:\s+leaseExpiringWithinSeconds/);
  assert.match(postgresStore, /func normalizeOutboxLeaseExpiringWithinSeconds/);
  assert.match(postgresStore, /ownerStatsByOwner := map\[string\]\*OutboxLeaseOwnerStats/);
  assert.match(postgresStore, /NextLeaseExpiresInSeconds/);
  assert.match(postgresStore, /return buildOutboxStats\(events, topic, now, req\.LeaseExpiringWithinSeconds\)/);
});

test("outbox relay worker is wired into deployment skeletons", () => {
  const compose = readFileSync(join(root, "infra/docker/compose.yml"), "utf8");
  const k8s = readFileSync(join(root, "infra/k8s/base/app-stack.yaml"), "utf8");
  assert.match(compose, /outbox-relay-worker:/);
  assert.match(compose, /OUTBOX_RELAY_TOPICS:/);
  assert.match(compose, /OUTBOX_RELAY_BATCH_LIMIT:/);
  assert.match(compose, /OUTBOX_RELAY_MAX_ATTEMPTS:/);
  assert.match(compose, /OUTBOX_RELAY_WORKER_ID:/);
  assert.match(compose, /OUTBOX_RELAY_LEASE_SECONDS:/);
  assert.match(compose, /OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS:/);
  assert.match(k8s, /name: outbox-relay-worker/);
  assert.match(k8s, /replicas: 2/);
  assert.match(k8s, /KAFKA_REST_URL/);
  assert.match(k8s, /OUTBOX_RELAY_MAX_ATTEMPTS/);
  assert.match(k8s, /OUTBOX_RELAY_WORKER_ID/);
  assert.match(k8s, /OUTBOX_RELAY_LEASE_SECONDS/);
  assert.match(k8s, /OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS/);
  assert.match(k8s, /outbox-relay-token/);
});

test("object scan worker is wired into scan callbacks and deployment skeletons", () => {
  const rootPackage = readFileSync(join(root, "package.json"), "utf8");
  const worker = readFileSync(join(root, "services/object-scan-worker/src/index.mjs"), "utf8");
  const workerTest = readFileSync(join(root, "services/object-scan-worker/src/index.test.mjs"), "utf8");
  const compose = readFileSync(join(root, "infra/docker/compose.yml"), "utf8");
  const k8s = readFileSync(join(root, "infra/k8s/base/app-stack.yaml"), "utf8");
  assert.match(rootPackage, /services\/object-scan-worker/);
  assert.match(rootPackage, /@infinitech\/object-scan-worker/);
  assert.match(worker, /export const workerName = "object-scan-worker"/);
  assert.match(worker, /object\.uploaded/);
  assert.match(worker, /after_sales\.evidence\.object_uploaded/);
  assert.match(worker, /createIdempotentConsumer/);
  assert.match(worker, /signUploadCallback/);
  assert.match(worker, /signScanResult/);
  assert.match(worker, /buildObjectDownloadURL/);
  assert.match(worker, /downloadObjectForScan/);
  assert.match(worker, /scanBufferWithClamAV/);
  assert.match(worker, /createClamAVScanner/);
  assert.match(worker, /net\.createConnection/);
  assert.match(worker, /OBJECT_STORAGE_DOWNLOAD_BASE_URL/);
  assert.match(worker, /OBJECT_SCAN_MAX_BYTES/);
  assert.match(worker, /\/api\/object-storage\/upload-callback/);
  assert.match(worker, /\/api\/object-storage\/scan-result/);
  assert.match(workerTest, /object scan consumer reports scan once/);
  assert.match(workerTest, /clamav scanner streams INSTREAM frames/);
  assert.match(workerTest, /object download enforces scan size limits/);
  assert.match(compose, /object-scan-worker:/);
  assert.match(compose, /OBJECT_SCAN_TOPICS:/);
  assert.match(compose, /OBJECT_STORAGE_DOWNLOAD_BASE_URL/);
  assert.match(compose, /OBJECT_SCAN_MAX_BYTES/);
  assert.match(compose, /clamav:/);
  assert.match(compose, /OBJECT_STORAGE_CALLBACK_SIGNING_SECRET/);
  assert.match(k8s, /name: object-scan-worker/);
  assert.match(k8s, /replicas: 2/);
  assert.match(k8s, /OBJECT_STORAGE_DOWNLOAD_BASE_URL/);
  assert.match(k8s, /OBJECT_SCAN_CLAMAV_TIMEOUT_MS/);
  assert.match(k8s, /name: clamav/);
  assert.match(k8s, /object-scan-worker-token/);
  assert.match(k8s, /object-storage-callback-signing-secret/);
});

test("object lifecycle cleanup is wired through API, BFF and worker skeletons", () => {
  const rootPackage = readFileSync(join(root, "package.json"), "utf8");
  const contracts = readFileSync(join(root, "services/api-go/internal/platform/contracts.go"), "utf8");
  const store = readFileSync(join(root, "services/api-go/internal/platform/store.go"), "utf8");
  const postgresStore = readFileSync(join(root, "services/api-go/internal/platform/postgres_store.go"), "utf8");
  const migration = readFileSync(join(root, "infra/db/migrations/0002_auth_payment.sql"), "utf8");
  const router = readFileSync(join(root, "services/api-go/internal/httpapi/router.go"), "utf8");
  const routerTest = readFileSync(join(root, "services/api-go/internal/httpapi/router_test.go"), "utf8");
  const bff = readFileSync(join(root, "services/bff/src/server.mjs"), "utf8");
  const bffTest = readFileSync(join(root, "services/bff/src/runtime.test.mjs"), "utf8");
  const worker = readFileSync(join(root, "services/object-lifecycle-worker/src/index.mjs"), "utf8");
  const workerTest = readFileSync(join(root, "services/object-lifecycle-worker/src/index.test.mjs"), "utf8");
  const compose = readFileSync(join(root, "infra/docker/compose.yml"), "utf8");
  const k8s = readFileSync(join(root, "infra/k8s/base/app-stack.yaml"), "utf8");
  assert.match(rootPackage, /services\/object-lifecycle-worker/);
  assert.match(rootPackage, /@infinitech\/object-lifecycle-worker/);
  assert.match(contracts, /AfterSalesUploadTicketDeleted/);
  assert.match(contracts, /type ObjectStorageCleanupCandidate struct/);
  assert.match(store, /func \(s \*Store\) ObjectStorageCleanupCandidates/);
  assert.match(store, /func \(s \*Store\) CompleteObjectStorageCleanup/);
  assert.match(store, /func \(s \*Store\) RecordObjectStorageCleanupFailure/);
  assert.match(store, /func \(s \*Store\) ObjectStorageCleanupStats/);
  assert.match(postgresStore, /cleanup_reason TEXT NOT NULL DEFAULT ''/);
  assert.match(postgresStore, /deleted_at TIMESTAMPTZ/);
  assert.match(postgresStore, /cleanup_attempts INTEGER NOT NULL DEFAULT 0/);
  assert.match(postgresStore, /last_cleanup_error TEXT NOT NULL DEFAULT ''/);
  assert.match(migration, /'deleted'/);
  assert.match(router, /\/api\/admin\/object-storage\/cleanup-candidates/);
  assert.match(router, /\/api\/admin\/object-storage\/cleanup-stats/);
  assert.match(router, /\/api\/admin\/object-storage\/cleanup-complete/);
  assert.match(router, /\/api\/admin\/object-storage\/cleanup-failed/);
  assert.match(routerTest, /TestAdminObjectStorageCleanupHTTPFlow/);
  assert.match(bff, /cleanup-candidates/);
  assert.match(bff, /cleanup-stats/);
  assert.match(bff, /cleanup-failed/);
  assert.match(bffTest, /completedObjectCleanup/);
  assert.match(bffTest, /failedObjectCleanup/);
  assert.match(bffTest, /objectCleanupStats/);
  assert.match(worker, /export const workerName = "object-lifecycle-worker"/);
  assert.match(worker, /cleanupObjectBatch/);
  assert.match(worker, /OBJECT_STORAGE_DELETE_BASE_URL/);
  assert.match(worker, /DELETE/);
  assert.match(worker, /recordCleanupFailure/);
  assert.match(workerTest, /cleanup batch deletes every candidate/);
  assert.match(compose, /object-lifecycle-worker:/);
  assert.match(compose, /OBJECT_LIFECYCLE_WORKER_TOKEN/);
  assert.match(k8s, /name: object-lifecycle-worker/);
  assert.match(k8s, /object-lifecycle-worker-token/);
});
