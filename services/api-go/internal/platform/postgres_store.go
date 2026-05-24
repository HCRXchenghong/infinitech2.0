package platform

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultPostgresStoreKey = "default"
const postgresOutboxEventColumns = `
	id, topic, aggregate_type, aggregate_id, event_type, idempotency_key,
	payload, status, attempts, last_error, available_at, lease_owner,
	lease_expires_at, published_at, created_at, updated_at`
const postgresOutboxEventReturningColumns = `
	event.id, event.topic, event.aggregate_type, event.aggregate_id, event.event_type, event.idempotency_key,
	event.payload, event.status, event.attempts, event.last_error, event.available_at, event.lease_owner,
	event.lease_expires_at, event.published_at, event.created_at, event.updated_at`
const postgresAuditLogColumns = `
	id, actor_type, actor_id, action, target_type, target_id,
	request_id, ip_hash, payload, integrity_algorithm, integrity_hash, created_at`

var ErrPersistence = errors.New("persistence failed")

type PostgresStore struct {
	*Store
	db       *sql.DB
	storeKey string
}

type storeSnapshot struct {
	Version                 int                                        `json:"version"`
	SavedAt                 time.Time                                  `json:"saved_at"`
	NextOrderID             uint64                                     `json:"next_order_id"`
	NextTransactionID       uint64                                     `json:"next_transaction_id"`
	NextAddressID           uint64                                     `json:"next_address_id"`
	NextMerchantID          uint64                                     `json:"next_merchant_id"`
	NextMerchantStaffID     uint64                                     `json:"next_merchant_staff_id"`
	NextMerchantMaterialID  uint64                                     `json:"next_merchant_material_id"`
	NextDispatchEventID     uint64                                     `json:"next_dispatch_event_id"`
	NextOutboxEventID       uint64                                     `json:"next_outbox_event_id"`
	NextAuditLogID          uint64                                     `json:"next_audit_log_id"`
	NextAfterSalesID        uint64                                     `json:"next_after_sales_id"`
	NextAfterSalesEventID   uint64                                     `json:"next_after_sales_event_id"`
	NextRiderID             uint64                                     `json:"next_rider_id"`
	NextProductID           uint64                                     `json:"next_product_id"`
	NextVoucherID           uint64                                     `json:"next_voucher_id"`
	HomeModules             []HomeModule                               `json:"home_modules"`
	HomeCards               []HomeCard                                 `json:"home_cards"`
	Users                   map[string]*AppUser                        `json:"users"`
	WechatBindings          map[string]string                          `json:"wechat_bindings"`
	MerchantInvites         map[string]*MerchantOnboardingInvite       `json:"merchant_invites"`
	Merchants               map[string]*MerchantAccount                `json:"merchants"`
	MerchantQualifications  map[string][]*MerchantQualification        `json:"merchant_qualifications"`
	MerchantStaff           map[string][]*MerchantStaff                `json:"merchant_staff"`
	MerchantMaterials       map[string][]*MerchantSupplementalMaterial `json:"merchant_materials"`
	Riders                  map[string]*RiderAccount                   `json:"riders"`
	Deposits                map[string]*DepositAccount                 `json:"deposits"`
	StationTaskConfigs      map[string]*StationTaskConfig              `json:"station_task_configs"`
	StationServiceAreas     map[string]*StationServiceArea             `json:"station_service_areas"`
	Shops                   map[string]*Shop                           `json:"shops"`
	Products                map[string]*MerchantProduct                `json:"products"`
	GroupbuyDeals           map[string]*MerchantProduct                `json:"groupbuy_deals"`
	Addresses               map[string][]*UserAddress                  `json:"addresses"`
	CartItems               map[string][]*CartItem                     `json:"cart_items"`
	Orders                  map[string]*Order                          `json:"orders"`
	Wallets                 map[string]*WalletAccount                  `json:"wallets"`
	PaymentPasswordHash     map[string]string                          `json:"payment_password_hash"`
	MerchantPasswordHash    map[string]string                          `json:"merchant_password_hash"`
	RiderPasswordHash       map[string]string                          `json:"rider_password_hash"`
	PaymentTransactions     map[string]*PaymentTransaction             `json:"payment_transactions"`
	PaymentByTradeNo        map[string]*PaymentTransaction             `json:"payment_by_trade_no"`
	PaymentByProviderID     map[string]*PaymentTransaction             `json:"payment_by_provider_id"`
	WalletIdempotency       map[string]*WalletTransaction              `json:"wallet_idempotency"`
	RefundSettings          RefundSettings                             `json:"refund_settings"`
	RefundTransactions      map[string]*RefundTransaction              `json:"refund_transactions"`
	RefundByIdempotency     map[string]string                          `json:"refund_by_idempotency"`
	AfterSalesRequests      map[string]*AfterSalesRequest              `json:"after_sales_requests"`
	AfterSalesEvents        map[string]*AfterSalesEvent                `json:"after_sales_events"`
	AfterSalesUploadTickets map[string]*AfterSalesEvidenceUploadTicket `json:"after_sales_upload_tickets"`
	AfterSalesEvidence      map[string]*AfterSalesEvidence             `json:"after_sales_evidence"`
	GroupbuyVouchers        map[string]*GroupbuyVoucher                `json:"groupbuy_vouchers"`
	VouchersByOrderID       map[string][]string                        `json:"vouchers_by_order_id"`
	VouchersByCode          map[string]*GroupbuyVoucher                `json:"vouchers_by_code"`
	DispatchEvents          map[string]*DispatchEvent                  `json:"dispatch_events"`
	DispatchRejectedRiders  map[string]map[string]bool                 `json:"dispatch_rejected_riders"`
	FreeCancelUsedByDate    map[string]string                          `json:"free_cancel_used_by_date"`
	OutboxEvents            map[string]*OutboxEvent                    `json:"outbox_events"`
	OutboxByIdempotency     map[string]string                          `json:"outbox_by_idempotency"`
	AuditLogs               map[string]*AuditLog                       `json:"audit_logs"`
}

func NewPostgresStore(ctx context.Context, databaseURL string, homeModules []HomeModule) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &PostgresStore{
		Store:    NewStore(homeModules),
		db:       db,
		storeKey: defaultPostgresStoreKey,
	}
	if err := store.ensureSnapshotTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensurePaymentDomainTables(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensureDispatchEventTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensureOutboxTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensureConsumedEventsTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.ensureAuditLogTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.loadSnapshot(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.syncPaymentDomainToTables(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.syncDispatchEventsToTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.syncSnapshotAuditLogsToTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.loadPaymentDomainFromTables(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.loadDispatchEventsFromTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.syncSnapshotOutboxToTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.loadOutboxFromTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.restoreNextAuditLogSequenceFromTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.saveSnapshot(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) ensureSnapshotTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS platform_store_snapshots (
  store_key TEXT PRIMARY KEY,
  payload JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	return err
}

func (s *PostgresStore) ensureOutboxTable(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS platform_outbox_events (
  id TEXT PRIMARY KEY,
  topic TEXT NOT NULL,
  aggregate_type TEXT NOT NULL,
  aggregate_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  idempotency_key TEXT NOT NULL UNIQUE,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'published', 'failed', 'dead_letter')),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  last_error TEXT NOT NULL DEFAULT '',
  available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  lease_owner TEXT NOT NULL DEFAULT '',
  lease_expires_at TIMESTAMPTZ,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
CREATE INDEX IF NOT EXISTS idx_platform_outbox_ready
  ON platform_outbox_events (status, topic, available_at, created_at)`); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
CREATE INDEX IF NOT EXISTS idx_platform_outbox_lease
  ON platform_outbox_events (lease_expires_at, lease_owner)`)
	return err
}

func (s *PostgresStore) ensureConsumedEventsTable(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS platform_consumed_events (
  consumer_event_key TEXT PRIMARY KEY,
  consumer_name TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  outbox_event_id TEXT NOT NULL DEFAULT '',
  topic TEXT NOT NULL DEFAULT '',
  aggregate_type TEXT NOT NULL DEFAULT '',
  aggregate_id TEXT NOT NULL DEFAULT '',
  event_type TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'processed', 'failed')),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  last_error TEXT NOT NULL DEFAULT '',
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (consumer_name, idempotency_key)
)`); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
CREATE INDEX IF NOT EXISTS idx_platform_consumed_events_topic
  ON platform_consumed_events (consumer_name, topic, updated_at)`)
	return err
}

func (s *PostgresStore) ensureAuditLogTable(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS audit_logs (
  id TEXT PRIMARY KEY,
  actor_type TEXT NOT NULL,
  actor_id TEXT NOT NULL,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  request_id TEXT NOT NULL DEFAULT '',
  ip_hash TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  integrity_algorithm TEXT NOT NULL DEFAULT '',
  integrity_hash TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`ALTER TABLE audit_logs ALTER COLUMN id DROP DEFAULT`,
		`ALTER TABLE audit_logs ALTER COLUMN id TYPE TEXT USING id::text`,
		`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS integrity_algorithm TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS integrity_hash TEXT NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_target_time
  ON audit_logs (target_type, target_id, created_at DESC, id DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_time
  ON audit_logs (actor_type, actor_id, created_at DESC, id DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_action_time
  ON audit_logs (action, created_at DESC, id DESC)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) loadSnapshot(ctx context.Context) error {
	var payload []byte
	err := s.db.QueryRowContext(ctx, "SELECT payload FROM platform_store_snapshots WHERE store_key = $1", s.storeKey).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var snapshot storeSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return err
	}
	s.Store.applySnapshot(snapshot)
	return nil
}

func (s *PostgresStore) saveSnapshot(ctx context.Context) error {
	payload, err := s.Store.snapshotPayload()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO platform_store_snapshots (store_key, payload, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (store_key) DO UPDATE
SET payload = EXCLUDED.payload, updated_at = now()`, s.storeKey, payload)
	return err
}

func (s *PostgresStore) saveSnapshotInTx(ctx context.Context, tx *sql.Tx) error {
	payload, err := s.Store.snapshotPayload()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO platform_store_snapshots (store_key, payload, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (store_key) DO UPDATE
SET payload = EXCLUDED.payload, updated_at = now()`, s.storeKey, payload)
	return err
}

func (s *PostgresStore) persistAfter(err error) error {
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := s.saveSnapshot(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	if err := s.syncPaymentDomainToTables(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	if err := s.syncDispatchEventsToTable(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	return nil
}

type merchantQualificationSnapshot struct {
	MerchantID    string
	Qualification MerchantQualification
}

type paymentDomainSnapshot struct {
	Users                   []AppUser
	Merchants               []MerchantAccount
	Qualifications          []merchantQualificationSnapshot
	Shops                   []Shop
	Products                []MerchantProduct
	GroupbuyDeals           []MerchantProduct
	Riders                  []RiderAccount
	Addresses               []UserAddress
	CartItems               []CartItem
	Orders                  []Order
	Wallets                 []WalletAccount
	WalletTransactions      []WalletTransaction
	PaymentPasswordHash     map[string]string
	PaymentTransactions     []PaymentTransaction
	RefundSettings          RefundSettings
	RefundTransactions      []RefundTransaction
	AfterSalesRequests      []AfterSalesRequest
	AfterSalesEvents        []AfterSalesEvent
	AfterSalesUploadTickets []AfterSalesEvidenceUploadTicket
	AfterSalesEvidence      []AfterSalesEvidence
}

func (s *PostgresStore) ensurePaymentDomainTables(ctx context.Context) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE IF NOT EXISTS platform_sequences (
  name TEXT PRIMARY KEY,
  next_value BIGINT NOT NULL CHECK (next_value >= 1),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS app_users (
  id TEXT PRIMARY KEY,
  phone_hash TEXT,
  nickname TEXT NOT NULL DEFAULT '',
  avatar_url TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  gender TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS merchant_accounts (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK (type IN ('standard', 'pharmacy', 'clinic', 'platform_service')),
  display_name TEXT NOT NULL,
  registration_mode TEXT NOT NULL DEFAULT 'admin_invite_only',
  deposit_status TEXT NOT NULL DEFAULT 'unpaid',
  operation_status TEXT NOT NULL DEFAULT 'pending_review',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS shops (
  id TEXT PRIMARY KEY,
  merchant_id TEXT NOT NULL REFERENCES merchant_accounts(id),
  name TEXT NOT NULL,
  category TEXT NOT NULL,
  account_type TEXT NOT NULL,
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  operation_status TEXT NOT NULL DEFAULT 'pending_review',
  service_state TEXT NOT NULL DEFAULT 'closed',
  cover_url TEXT NOT NULL DEFAULT '',
  logo_url TEXT NOT NULL DEFAULT '',
  announcement TEXT NOT NULL DEFAULT '',
  rating NUMERIC(3,2) NOT NULL DEFAULT 5.00,
  monthly_sales INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS merchant_qualifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  merchant_id TEXT NOT NULL REFERENCES merchant_accounts(id),
  shop_id TEXT REFERENCES shops(id),
  type TEXT NOT NULL CHECK (type IN ('business_license', 'health_certificate', 'supplemental_document')),
  file_url TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending_review',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS idx_merchant_qualifications_expiry ON merchant_qualifications (merchant_id, shop_id, type, expires_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uniq_merchant_qualifications_merchant_type ON merchant_qualifications (merchant_id, type)`,
		`CREATE TABLE IF NOT EXISTS merchant_products (
  id TEXT PRIMARY KEY,
  shop_id TEXT NOT NULL REFERENCES shops(id),
  type TEXT NOT NULL DEFAULT 'takeout',
  name TEXT NOT NULL,
  image_url TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  ingredient_list TEXT[] NOT NULL DEFAULT '{}',
  price_fen BIGINT NOT NULL CHECK (price_fen >= 0),
  stock_count INTEGER NOT NULL DEFAULT 0 CHECK (stock_count >= 0),
  status TEXT NOT NULL CHECK (status IN ('active', 'sold_out', 'removed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS idx_merchant_products_shop_status ON merchant_products (shop_id, status)`,
		`CREATE TABLE IF NOT EXISTS user_addresses (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES app_users(id),
  contact_name TEXT NOT NULL,
  contact_phone_hash TEXT NOT NULL,
  city TEXT NOT NULL,
  detail TEXT NOT NULL,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  tag TEXT NOT NULL DEFAULT 'other',
  is_default BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS rider_accounts (
  id TEXT PRIMARY KEY,
  station_id TEXT NOT NULL DEFAULT '',
  type TEXT NOT NULL CHECK (type IN ('station_manager', 'rider')),
  status TEXT NOT NULL DEFAULT 'pending_review',
  online BOOLEAN NOT NULL DEFAULT false,
  deposit_status TEXT NOT NULL DEFAULT 'unpaid',
  current_latitude DOUBLE PRECISION,
  current_longitude DOUBLE PRECISION,
  last_location_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES app_users(id),
  merchant_id TEXT REFERENCES merchant_accounts(id),
  shop_id TEXT REFERENCES shops(id),
  rider_id TEXT REFERENCES rider_accounts(id),
  address_id TEXT REFERENCES user_addresses(id),
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  payment_method TEXT NOT NULL DEFAULT '',
  items_total_fen BIGINT NOT NULL DEFAULT 0 CHECK (items_total_fen >= 0),
  delivery_fee_fen BIGINT NOT NULL DEFAULT 0 CHECK (delivery_fee_fen >= 0),
  packaging_fee_fen BIGINT NOT NULL DEFAULT 0 CHECK (packaging_fee_fen >= 0),
  discount_fen BIGINT NOT NULL DEFAULT 0 CHECK (discount_fen >= 0),
  amount_fen BIGINT NOT NULL CHECK (amount_fen >= 0),
  address_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  options JSONB NOT NULL DEFAULT '{}'::jsonb,
  pricing_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  dispatch_mode TEXT NOT NULL DEFAULT 'grab_hall',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  paid_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT NOT NULL,
  product_name_snapshot TEXT NOT NULL,
  unit_price_fen BIGINT NOT NULL CHECK (unit_price_fen >= 0),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
)`,
		`CREATE TABLE IF NOT EXISTS order_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  actor_type TEXT NOT NULL DEFAULT 'system',
  actor_id TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS wallet_accounts (
  subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'merchant', 'rider', 'platform')),
  subject_id TEXT NOT NULL,
  balance_fen BIGINT NOT NULL DEFAULT 0 CHECK (balance_fen >= 0),
  frozen_fen BIGINT NOT NULL DEFAULT 0 CHECK (frozen_fen >= 0),
  version BIGINT NOT NULL DEFAULT 0,
  risk_state TEXT NOT NULL DEFAULT 'normal',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (subject_type, subject_id)
)`,
		`CREATE TABLE IF NOT EXISTS cart_items (
  user_id TEXT NOT NULL REFERENCES app_users(id),
  shop_id TEXT NOT NULL REFERENCES shops(id),
  product_id TEXT NOT NULL REFERENCES merchant_products(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  selected BOOLEAN NOT NULL DEFAULT true,
  unit_price_fen BIGINT NOT NULL CHECK (unit_price_fen >= 0),
  product_name_snapshot TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, shop_id, product_id)
)`,
		`CREATE TABLE IF NOT EXISTS wallet_transactions (
  id TEXT PRIMARY KEY,
  subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  order_id TEXT,
  type TEXT NOT NULL,
  amount_fen BIGINT NOT NULL,
  payment_method TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  status TEXT NOT NULL,
  balance_after_fen BIGINT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (idempotency_key)
)`,
		`CREATE TABLE IF NOT EXISTS refund_settings (
  id TEXT PRIMARY KEY DEFAULT 'default',
  default_strategy TEXT NOT NULL CHECK (default_strategy IN ('balance_first', 'original_route_first')),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS refund_transactions (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id),
  user_id TEXT NOT NULL REFERENCES app_users(id),
  amount_fen BIGINT NOT NULL CHECK (amount_fen >= 0),
  destination TEXT NOT NULL CHECK (destination IN ('balance', 'original_route')),
  status TEXT NOT NULL,
  reason TEXT NOT NULL,
  idempotency_key TEXT NOT NULL UNIQUE,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS order_after_sales (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id),
  user_id TEXT NOT NULL REFERENCES app_users(id),
  type TEXT NOT NULL CHECK (type IN ('refund_only', 'partial_refund', 'food_safety')),
  reason TEXT NOT NULL,
  requested_amount_fen BIGINT NOT NULL CHECK (requested_amount_fen >= 0),
  evidence_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL CHECK (status IN ('pending_merchant', 'admin_review', 'approved', 'rejected', 'refunded')),
  review_reason TEXT NOT NULL DEFAULT '',
  reviewer_id TEXT NOT NULL DEFAULT '',
  reviewer_role TEXT NOT NULL DEFAULT '',
  refund_id TEXT REFERENCES refund_transactions(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  reviewed_at TIMESTAMPTZ
)`,
		`CREATE TABLE IF NOT EXISTS order_after_sales_events (
	  id TEXT PRIMARY KEY,
	  request_id TEXT NOT NULL REFERENCES order_after_sales(id) ON DELETE CASCADE,
	  order_id TEXT NOT NULL REFERENCES orders(id),
  actor_id TEXT NOT NULL,
  actor_role TEXT NOT NULL,
	  action TEXT NOT NULL CHECK (action IN ('created', 'user_supplement', 'merchant_reply', 'customer_service_intervention', 'arbitration_opened', 'internal_note', 'evidence_uploaded', 'review_approved', 'review_rejected', 'escalated')),
  message TEXT NOT NULL,
  attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
	  visible_to_user BOOLEAN NOT NULL DEFAULT TRUE,
	  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
		`ALTER TABLE order_after_sales_events DROP CONSTRAINT IF EXISTS order_after_sales_events_action_check`,
		`ALTER TABLE order_after_sales_events ADD CONSTRAINT order_after_sales_events_action_check
	  CHECK (action IN ('created', 'user_supplement', 'merchant_reply', 'customer_service_intervention', 'arbitration_opened', 'internal_note', 'evidence_uploaded', 'review_approved', 'review_rejected', 'escalated'))`,
		`CREATE TABLE IF NOT EXISTS order_after_sales_evidence_upload_tickets (
	  id TEXT PRIMARY KEY,
	  request_id TEXT NOT NULL REFERENCES order_after_sales(id) ON DELETE CASCADE,
	  order_id TEXT NOT NULL REFERENCES orders(id),
	  provider TEXT NOT NULL,
	  bucket TEXT NOT NULL,
	  object_key TEXT NOT NULL UNIQUE,
	  public_url TEXT NOT NULL,
	  file_name TEXT NOT NULL,
	  content_type TEXT NOT NULL,
	  size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
	  max_size_bytes BIGINT NOT NULL CHECK (max_size_bytes > 0),
	  content_sha TEXT NOT NULL DEFAULT '',
	  uploaded_by_id TEXT NOT NULL,
	  uploaded_by_role TEXT NOT NULL,
	  status TEXT NOT NULL CHECK (status IN ('issued', 'uploaded', 'confirmed', 'deleted')),
	  scan_status TEXT NOT NULL DEFAULT 'not_required' CHECK (scan_status IN ('not_required', 'pending', 'passed', 'rejected')),
	  scan_result TEXT NOT NULL DEFAULT '',
	  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	  expires_at TIMESTAMPTZ NOT NULL,
	  uploaded_at TIMESTAMPTZ,
	  confirmed_at TIMESTAMPTZ,
	  scan_checked_at TIMESTAMPTZ,
	  cleanup_reason TEXT NOT NULL DEFAULT '',
	  deleted_at TIMESTAMPTZ,
	  cleanup_attempts INTEGER NOT NULL DEFAULT 0 CHECK (cleanup_attempts >= 0),
	  last_cleanup_error TEXT NOT NULL DEFAULT '',
	  last_cleanup_failed_at TIMESTAMPTZ
	)`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets DROP CONSTRAINT IF EXISTS order_after_sales_evidence_upload_tickets_status_check`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD CONSTRAINT order_after_sales_evidence_upload_tickets_status_check
	  CHECK (status IN ('issued', 'uploaded', 'confirmed', 'deleted'))`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS content_sha TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS scan_status TEXT NOT NULL DEFAULT 'not_required'`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS scan_result TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS uploaded_at TIMESTAMPTZ`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS scan_checked_at TIMESTAMPTZ`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS cleanup_reason TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS cleanup_attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS last_cleanup_error TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD COLUMN IF NOT EXISTS last_cleanup_failed_at TIMESTAMPTZ`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets DROP CONSTRAINT IF EXISTS order_after_sales_evidence_upload_tickets_scan_status_check`,
		`ALTER TABLE order_after_sales_evidence_upload_tickets ADD CONSTRAINT order_after_sales_evidence_upload_tickets_scan_status_check
	  CHECK (scan_status IN ('not_required', 'pending', 'passed', 'rejected'))`,
		`CREATE TABLE IF NOT EXISTS order_after_sales_evidence (
	  id TEXT PRIMARY KEY,
	  request_id TEXT NOT NULL REFERENCES order_after_sales(id) ON DELETE CASCADE,
	  order_id TEXT NOT NULL REFERENCES orders(id),
	  object_key TEXT NOT NULL UNIQUE,
	  public_url TEXT NOT NULL,
	  file_name TEXT NOT NULL,
	  content_type TEXT NOT NULL,
	  size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
	  content_sha TEXT NOT NULL DEFAULT '',
	  uploaded_by_id TEXT NOT NULL,
	  uploaded_by_role TEXT NOT NULL,
	  status TEXT NOT NULL CHECK (status IN ('uploaded')),
	  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	  confirmed_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
		`CREATE TABLE IF NOT EXISTS wallet_payment_passwords (
  user_id TEXT PRIMARY KEY REFERENCES app_users(id),
  password_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('unset', 'set', 'locked')),
  failed_count INTEGER NOT NULL DEFAULT 0,
  locked_until TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS payment_transactions (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id),
  user_id TEXT NOT NULL REFERENCES app_users(id),
  method TEXT NOT NULL CHECK (method IN ('wechat_pay', 'balance')),
  amount_fen BIGINT NOT NULL CHECK (amount_fen >= 0),
  status TEXT NOT NULL,
  out_trade_no TEXT NOT NULL UNIQUE,
  transaction_id TEXT UNIQUE,
  idempotency_key TEXT NOT NULL UNIQUE,
  raw_request JSONB NOT NULL DEFAULT '{}'::jsonb,
  raw_response JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_time ON orders (user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_shop_status ON orders (shop_id, status, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_transactions_subject_time ON wallet_transactions (subject_type, subject_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_refund_transactions_order ON refund_transactions (order_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_order_after_sales_order ON order_after_sales (order_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_order_after_sales_user ON order_after_sales (user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_order_after_sales_events_request ON order_after_sales_events (request_id, created_at, id)`,
		`CREATE INDEX IF NOT EXISTS idx_order_after_sales_upload_tickets_request ON order_after_sales_evidence_upload_tickets (request_id, created_at, id)`,
		`CREATE INDEX IF NOT EXISTS idx_order_after_sales_evidence_request ON order_after_sales_evidence (request_id, created_at, id)`,
		`CREATE INDEX IF NOT EXISTS idx_payment_transactions_order ON payment_transactions (order_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_order_events_order_time ON order_events (order_id, created_at)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) ensureDispatchEventTable(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS dispatch_events (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  station_id TEXT NOT NULL DEFAULT '',
  mode TEXT NOT NULL DEFAULT 'auto_assign',
  type TEXT NOT NULL,
  rider_id TEXT NOT NULL DEFAULT '',
  actor_id TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  idempotency_key TEXT NOT NULL,
  online_candidate_size INTEGER NOT NULL DEFAULT 0 CHECK (online_candidate_size >= 0),
  rejected_rider_ids TEXT[] NOT NULL DEFAULT '{}',
  can_decline_without_penalty BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`ALTER TABLE dispatch_events DROP CONSTRAINT IF EXISTS dispatch_events_idempotency_key_key`,
		`CREATE INDEX IF NOT EXISTS idx_dispatch_events_order_time ON dispatch_events (order_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_dispatch_events_station_time ON dispatch_events (station_id, created_at DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uniq_dispatch_events_idempotency_type ON dispatch_events (idempotency_key, type)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) syncPaymentDomainToTables(ctx context.Context) error {
	snapshot := s.Store.paymentDomainSnapshot()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, user := range snapshot.Users {
		if err := upsertSQLUser(ctx, tx, user); err != nil {
			return err
		}
	}
	for _, merchant := range snapshot.Merchants {
		if err := upsertSQLMerchant(ctx, tx, merchant); err != nil {
			return err
		}
	}
	for _, shop := range snapshot.Shops {
		if err := upsertSQLShop(ctx, tx, shop); err != nil {
			return err
		}
	}
	for _, qualification := range snapshot.Qualifications {
		if err := upsertSQLMerchantQualification(ctx, tx, qualification); err != nil {
			return err
		}
	}
	for _, product := range snapshot.Products {
		if err := upsertSQLMerchantProduct(ctx, tx, product, "takeout"); err != nil {
			return err
		}
	}
	for _, deal := range snapshot.GroupbuyDeals {
		if err := upsertSQLMerchantProduct(ctx, tx, deal, "groupbuy"); err != nil {
			return err
		}
	}
	for _, rider := range snapshot.Riders {
		if err := upsertSQLRider(ctx, tx, rider); err != nil {
			return err
		}
	}
	for _, address := range snapshot.Addresses {
		if err := upsertSQLAddress(ctx, tx, address); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM cart_items"); err != nil {
		return err
	}
	for _, item := range snapshot.CartItems {
		if err := upsertSQLCartItem(ctx, tx, item); err != nil {
			return err
		}
	}
	for _, order := range snapshot.Orders {
		if err := upsertSQLOrder(ctx, tx, order, snapshot); err != nil {
			return err
		}
	}
	for _, wallet := range snapshot.Wallets {
		if err := upsertSQLWalletAccount(ctx, tx, wallet); err != nil {
			return err
		}
	}
	for _, transaction := range snapshot.WalletTransactions {
		if err := upsertSQLWalletTransaction(ctx, tx, transaction); err != nil {
			return err
		}
	}
	for userID, passwordHash := range snapshot.PaymentPasswordHash {
		if err := upsertSQLWalletPaymentPassword(ctx, tx, userID, passwordHash); err != nil {
			return err
		}
	}
	for _, transaction := range snapshot.PaymentTransactions {
		if err := upsertSQLPaymentTransaction(ctx, tx, transaction); err != nil {
			return err
		}
	}
	if err := upsertSQLRefundSettings(ctx, tx, snapshot.RefundSettings); err != nil {
		return err
	}
	for _, refund := range snapshot.RefundTransactions {
		if err := upsertSQLRefundTransaction(ctx, tx, refund); err != nil {
			return err
		}
	}
	for _, request := range snapshot.AfterSalesRequests {
		if err := upsertSQLAfterSalesRequest(ctx, tx, request); err != nil {
			return err
		}
	}
	for _, ticket := range snapshot.AfterSalesUploadTickets {
		if err := upsertSQLAfterSalesUploadTicket(ctx, tx, ticket); err != nil {
			return err
		}
	}
	for _, evidence := range snapshot.AfterSalesEvidence {
		if err := upsertSQLAfterSalesEvidence(ctx, tx, evidence); err != nil {
			return err
		}
	}
	for _, event := range snapshot.AfterSalesEvents {
		if err := upsertSQLAfterSalesEvent(ctx, tx, event); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStore) syncDispatchEventsToTable(ctx context.Context) error {
	events := s.Store.dispatchEventSnapshot()
	if len(events) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, event := range events {
		if err := upsertSQLDispatchEvent(ctx, tx, event); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStore) loadPaymentDomainFromTables(ctx context.Context) error {
	orders, err := s.loadSQLOrders(ctx)
	if err != nil {
		return err
	}
	wallets, err := s.loadSQLWalletAccounts(ctx)
	if err != nil {
		return err
	}
	walletTransactions, err := s.loadSQLWalletTransactions(ctx)
	if err != nil {
		return err
	}
	paymentPasswordHash, err := s.loadSQLWalletPaymentPasswords(ctx)
	if err != nil {
		return err
	}
	paymentTransactions, err := s.loadSQLPaymentTransactions(ctx)
	if err != nil {
		return err
	}
	refundSettings, err := s.loadSQLRefundSettings(ctx)
	if err != nil {
		return err
	}
	refundTransactions, err := s.loadSQLRefundTransactions(ctx)
	if err != nil {
		return err
	}
	afterSalesRequests, err := s.loadSQLAfterSalesRequests(ctx)
	if err != nil {
		return err
	}
	afterSalesEvents, err := s.loadSQLAfterSalesEvents(ctx)
	if err != nil {
		return err
	}
	afterSalesUploadTickets, err := s.loadSQLAfterSalesUploadTickets(ctx)
	if err != nil {
		return err
	}
	afterSalesEvidence, err := s.loadSQLAfterSalesEvidence(ctx)
	if err != nil {
		return err
	}
	s.Store.replacePaymentDomainFromTables(orders, wallets, walletTransactions, paymentPasswordHash, paymentTransactions)
	s.Store.replaceRefundDomainFromTables(refundSettings, refundTransactions)
	s.Store.replaceAfterSalesDomainFromTables(afterSalesRequests, afterSalesEvents)
	s.Store.replaceAfterSalesUploadTicketsFromTables(afterSalesUploadTickets)
	s.Store.replaceAfterSalesEvidenceFromTables(afterSalesEvidence)
	return nil
}

func (s *PostgresStore) loadDispatchEventsFromTable(ctx context.Context) error {
	events, err := s.loadSQLDispatchEvents(ctx)
	if err != nil {
		return err
	}
	s.Store.replaceDispatchEventsFromTable(events)
	return nil
}

func (s *Store) paymentDomainSnapshot() paymentDomainSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	usersByID := map[string]AppUser{}
	ensureUser := func(userID string) {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			return
		}
		if _, exists := usersByID[userID]; exists {
			return
		}
		now := time.Now().UTC()
		usersByID[userID] = AppUser{ID: userID, Status: "active", CreatedAt: now, UpdatedAt: now}
	}
	for _, user := range s.users {
		if user == nil || strings.TrimSpace(user.ID) == "" {
			continue
		}
		usersByID[user.ID] = *cloneAppUser(user)
	}

	merchants := make([]MerchantAccount, 0, len(s.merchants))
	for _, merchant := range s.merchants {
		if merchant == nil {
			continue
		}
		merchants = append(merchants, *cloneMerchantAccount(merchant))
	}
	sort.SliceStable(merchants, func(i, j int) bool { return merchants[i].ID < merchants[j].ID })

	qualifications := []merchantQualificationSnapshot{}
	for merchantID, entries := range s.merchantQualifications {
		merchantID = strings.TrimSpace(merchantID)
		if merchantID == "" {
			continue
		}
		for _, item := range entries {
			if item == nil || strings.TrimSpace(item.Type) == "" || strings.TrimSpace(item.FileURL) == "" {
				continue
			}
			qualifications = append(qualifications, merchantQualificationSnapshot{
				MerchantID:    merchantID,
				Qualification: *cloneMerchantQualification(item),
			})
		}
	}
	sort.SliceStable(qualifications, func(i, j int) bool {
		if qualifications[i].MerchantID != qualifications[j].MerchantID {
			return qualifications[i].MerchantID < qualifications[j].MerchantID
		}
		if qualifications[i].Qualification.Type != qualifications[j].Qualification.Type {
			return qualifications[i].Qualification.Type < qualifications[j].Qualification.Type
		}
		return qualifications[i].Qualification.ID < qualifications[j].Qualification.ID
	})

	shops := make([]Shop, 0, len(s.shops))
	for _, shop := range s.shops {
		if shop == nil {
			continue
		}
		shops = append(shops, *cloneShop(shop))
	}
	sort.SliceStable(shops, func(i, j int) bool { return shops[i].ID < shops[j].ID })

	products := make([]MerchantProduct, 0, len(s.products))
	for _, product := range s.products {
		if product == nil {
			continue
		}
		products = append(products, *cloneMerchantProduct(product))
	}
	sort.SliceStable(products, func(i, j int) bool { return products[i].ID < products[j].ID })

	groupbuyDeals := make([]MerchantProduct, 0, len(s.groupbuyDeals))
	for _, deal := range s.groupbuyDeals {
		if deal == nil {
			continue
		}
		groupbuyDeals = append(groupbuyDeals, *cloneMerchantProduct(deal))
	}
	sort.SliceStable(groupbuyDeals, func(i, j int) bool { return groupbuyDeals[i].ID < groupbuyDeals[j].ID })

	riders := make([]RiderAccount, 0, len(s.riders))
	for _, rider := range s.riders {
		if rider == nil {
			continue
		}
		riders = append(riders, *cloneRiderAccount(rider))
	}
	sort.SliceStable(riders, func(i, j int) bool { return riders[i].ID < riders[j].ID })

	addresses := []UserAddress{}
	for userID, entries := range s.addresses {
		ensureUser(userID)
		for _, address := range entries {
			if address == nil {
				continue
			}
			addresses = append(addresses, *cloneUserAddress(address))
			ensureUser(address.UserID)
		}
	}
	sort.SliceStable(addresses, func(i, j int) bool { return addresses[i].ID < addresses[j].ID })

	cartItems := []CartItem{}
	for keyUserShop, entries := range s.cartItems {
		for _, item := range entries {
			if item == nil {
				continue
			}
			cloned := *cloneCartItem(item)
			if strings.TrimSpace(cloned.UserID) == "" || strings.TrimSpace(cloned.ShopID) == "" {
				parts := strings.SplitN(keyUserShop, "::", 2)
				if len(parts) == 2 {
					if strings.TrimSpace(cloned.UserID) == "" {
						cloned.UserID = parts[0]
					}
					if strings.TrimSpace(cloned.ShopID) == "" {
						cloned.ShopID = parts[1]
					}
				}
			}
			if strings.TrimSpace(cloned.UserID) == "" || strings.TrimSpace(cloned.ShopID) == "" || strings.TrimSpace(cloned.ProductID) == "" || cloned.Quantity <= 0 {
				continue
			}
			cartItems = append(cartItems, cloned)
			ensureUser(cloned.UserID)
		}
	}
	sort.SliceStable(cartItems, func(i, j int) bool {
		if cartItems[i].UserID != cartItems[j].UserID {
			return cartItems[i].UserID < cartItems[j].UserID
		}
		if cartItems[i].ShopID != cartItems[j].ShopID {
			return cartItems[i].ShopID < cartItems[j].ShopID
		}
		return cartItems[i].ProductID < cartItems[j].ProductID
	})

	orders := make([]Order, 0, len(s.orders))
	for _, order := range s.orders {
		if order == nil {
			continue
		}
		orders = append(orders, *cloneOrder(order))
		ensureUser(order.UserID)
	}
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].CreatedAt.Equal(orders[j].CreatedAt) {
			return orders[i].ID < orders[j].ID
		}
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	wallets := make([]WalletAccount, 0, len(s.wallets))
	for userID, wallet := range s.wallets {
		ensureUser(userID)
		if wallet == nil {
			continue
		}
		wallets = append(wallets, *cloneWalletAccount(wallet))
		ensureUser(wallet.UserID)
	}
	sort.SliceStable(wallets, func(i, j int) bool { return wallets[i].UserID < wallets[j].UserID })

	walletTransactionsByKey := map[string]WalletTransaction{}
	for idempotencyKey, transaction := range s.walletIdempotency {
		if transaction == nil {
			continue
		}
		key := strings.TrimSpace(transaction.ID)
		if key == "" {
			key = strings.TrimSpace(idempotencyKey)
		}
		if key == "" {
			continue
		}
		walletTransactionsByKey[key] = *cloneWalletTransaction(transaction)
		ensureUser(transaction.UserID)
	}
	walletTransactions := make([]WalletTransaction, 0, len(walletTransactionsByKey))
	for _, transaction := range walletTransactionsByKey {
		walletTransactions = append(walletTransactions, transaction)
	}
	sort.SliceStable(walletTransactions, func(i, j int) bool {
		if walletTransactions[i].CreatedAt.Equal(walletTransactions[j].CreatedAt) {
			return walletTransactions[i].ID < walletTransactions[j].ID
		}
		return walletTransactions[i].CreatedAt.Before(walletTransactions[j].CreatedAt)
	})

	paymentPasswordHash := map[string]string{}
	for userID, passwordHash := range s.paymentPasswordHash {
		userID = strings.TrimSpace(userID)
		passwordHash = strings.TrimSpace(passwordHash)
		if userID == "" || passwordHash == "" {
			continue
		}
		paymentPasswordHash[userID] = passwordHash
		ensureUser(userID)
	}

	paymentTransactions := make([]PaymentTransaction, 0, len(s.paymentTransactions))
	for _, transaction := range s.paymentTransactions {
		if transaction == nil {
			continue
		}
		paymentTransactions = append(paymentTransactions, *clonePaymentTransaction(transaction))
		ensureUser(transaction.UserID)
	}
	sort.SliceStable(paymentTransactions, func(i, j int) bool {
		if paymentTransactions[i].CreatedAt.Equal(paymentTransactions[j].CreatedAt) {
			return paymentTransactions[i].ID < paymentTransactions[j].ID
		}
		return paymentTransactions[i].CreatedAt.Before(paymentTransactions[j].CreatedAt)
	})

	refundTransactions := make([]RefundTransaction, 0, len(s.refundTransactions))
	for _, refund := range s.refundTransactions {
		if refund == nil {
			continue
		}
		refundTransactions = append(refundTransactions, *cloneRefundTransaction(refund))
		ensureUser(refund.UserID)
	}
	sort.SliceStable(refundTransactions, func(i, j int) bool {
		if refundTransactions[i].CreatedAt.Equal(refundTransactions[j].CreatedAt) {
			return refundTransactions[i].ID < refundTransactions[j].ID
		}
		return refundTransactions[i].CreatedAt.Before(refundTransactions[j].CreatedAt)
	})

	afterSalesRequests := make([]AfterSalesRequest, 0, len(s.afterSalesRequests))
	for _, request := range s.afterSalesRequests {
		if request == nil || strings.TrimSpace(request.ID) == "" {
			continue
		}
		clonedRequest := *cloneAfterSalesRequest(request)
		clonedRequest.EvidenceURLs = sanitizedStringSlice(clonedRequest.EvidenceURLs)
		afterSalesRequests = append(afterSalesRequests, clonedRequest)
		ensureUser(request.UserID)
	}
	sortAfterSalesRequests(afterSalesRequests)

	afterSalesEvents := make([]AfterSalesEvent, 0, len(s.afterSalesEvents))
	for _, event := range s.afterSalesEvents {
		if event == nil || strings.TrimSpace(event.ID) == "" {
			continue
		}
		clonedEvent := *cloneAfterSalesEvent(event)
		clonedEvent.Action = normalizeAfterSalesAction(clonedEvent.Action)
		clonedEvent.Attachments = sanitizedStringSlice(clonedEvent.Attachments)
		if clonedEvent.Action == "" {
			continue
		}
		afterSalesEvents = append(afterSalesEvents, clonedEvent)
	}
	sortAfterSalesEvents(afterSalesEvents)

	afterSalesUploadTickets := make([]AfterSalesEvidenceUploadTicket, 0, len(s.afterSalesUploadTickets))
	for _, ticket := range s.afterSalesUploadTickets {
		if ticket == nil || strings.TrimSpace(ticket.ID) == "" {
			continue
		}
		clonedTicket := *cloneAfterSalesEvidenceUploadTicket(ticket)
		if clonedTicket.Status == "" {
			clonedTicket.Status = AfterSalesUploadTicketIssued
		}
		clonedTicket.ScanStatus = normalizeAfterSalesUploadScanStatus(clonedTicket.ScanStatus)
		if clonedTicket.ScanStatus == "" {
			clonedTicket.ScanStatus = AfterSalesUploadScanNotRequired
		}
		afterSalesUploadTickets = append(afterSalesUploadTickets, clonedTicket)
	}
	sortAfterSalesUploadTickets(afterSalesUploadTickets)

	afterSalesEvidence := make([]AfterSalesEvidence, 0, len(s.afterSalesEvidence))
	for _, evidence := range s.afterSalesEvidence {
		if evidence == nil || strings.TrimSpace(evidence.ID) == "" {
			continue
		}
		clonedEvidence := *cloneAfterSalesEvidence(evidence)
		if clonedEvidence.Status == "" {
			clonedEvidence.Status = AfterSalesEvidenceUploaded
		}
		afterSalesEvidence = append(afterSalesEvidence, clonedEvidence)
	}
	sortAfterSalesEvidence(afterSalesEvidence)

	users := make([]AppUser, 0, len(usersByID))
	for _, user := range usersByID {
		users = append(users, user)
	}
	sort.SliceStable(users, func(i, j int) bool { return users[i].ID < users[j].ID })

	return paymentDomainSnapshot{
		Users:                   users,
		Merchants:               merchants,
		Qualifications:          qualifications,
		Shops:                   shops,
		Products:                products,
		GroupbuyDeals:           groupbuyDeals,
		Riders:                  riders,
		Addresses:               addresses,
		CartItems:               cartItems,
		Orders:                  orders,
		Wallets:                 wallets,
		WalletTransactions:      walletTransactions,
		PaymentPasswordHash:     paymentPasswordHash,
		PaymentTransactions:     paymentTransactions,
		RefundSettings:          normalizeStoredRefundSettings(s.refundSettings),
		RefundTransactions:      refundTransactions,
		AfterSalesRequests:      afterSalesRequests,
		AfterSalesEvents:        afterSalesEvents,
		AfterSalesUploadTickets: afterSalesUploadTickets,
		AfterSalesEvidence:      afterSalesEvidence,
	}
}

func (s *Store) dispatchEventSnapshot() []DispatchEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	events := make([]DispatchEvent, 0, len(s.dispatchEvents))
	for _, event := range s.dispatchEvents {
		if event == nil {
			continue
		}
		events = append(events, *cloneDispatchEvent(event))
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events
}

func (s *Store) auditLogSnapshot() []AuditLog {
	s.mu.Lock()
	defer s.mu.Unlock()

	logs := make([]AuditLog, 0, len(s.auditLogs))
	for _, log := range s.auditLogs {
		if log == nil {
			continue
		}
		logs = append(logs, *cloneAuditLog(log))
	}
	sort.SliceStable(logs, func(i, j int) bool {
		if logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].ID < logs[j].ID
		}
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})
	return logs
}

func (s *Store) advanceAuditLogSequence(maxID uint64) {
	if maxID == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if maxID > s.nextAuditLogID {
		s.nextAuditLogID = maxID
	}
}

func (s *Store) applyAuditLogFromSQL(log AuditLog) {
	if strings.TrimSpace(log.ID) == "" {
		return
	}
	log.Payload = sanitizeAuditPayload(log.Payload)
	ensureAuditLogIntegrity(&log, s.auditLogSigningSecretSnapshot())
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogs[log.ID] = cloneAuditLog(&log)
	if value, err := strconv.ParseUint(strings.TrimPrefix(log.ID, "aud_"), 10, 64); err == nil && value > s.nextAuditLogID {
		s.nextAuditLogID = value
	}
}

func (s *PostgresStore) syncSnapshotAuditLogsToTable(ctx context.Context) error {
	logs := s.Store.auditLogSnapshot()
	if len(logs) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, log := range logs {
		if err := upsertSQLAuditLog(ctx, tx, log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStore) restoreNextAuditLogSequenceFromTable(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, "SELECT id FROM audit_logs WHERE id LIKE 'aud_%'")
	if err != nil {
		return err
	}
	defer rows.Close()
	var maxID uint64
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		value, err := strconv.ParseUint(strings.TrimPrefix(id, "aud_"), 10, 64)
		if err == nil && value > maxID {
			maxID = value
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	s.Store.advanceAuditLogSequence(maxID)
	return nil
}

func auditLogFromRequest(req RecordAuditLogRequest, id string) (*AuditLog, error) {
	actorType := strings.TrimSpace(req.ActorType)
	actorID := strings.TrimSpace(req.ActorID)
	action := strings.TrimSpace(req.Action)
	targetType := strings.TrimSpace(req.TargetType)
	targetID := strings.TrimSpace(req.TargetID)
	if actorType == "" || actorID == "" || action == "" || targetType == "" || targetID == "" {
		return nil, ErrInvalidArgument
	}
	now := req.CreatedAt
	if now.IsZero() {
		now = normalizeAuditLogTime(time.Now())
	}
	now = normalizeAuditLogTime(now)
	return &AuditLog{
		ID:         strings.TrimSpace(id),
		ActorType:  actorType,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		RequestID:  strings.TrimSpace(req.RequestID),
		IPHash:     strings.TrimSpace(req.IPHash),
		Payload:    sanitizeAuditPayload(req.Payload),
		CreatedAt:  now,
	}, nil
}

func upsertSQLAuditLog(ctx context.Context, tx *sql.Tx, log AuditLog, signingSecret string) error {
	return writeSQLAuditLog(ctx, tx, log, false, signingSecret)
}

func insertSQLAuditLog(ctx context.Context, tx *sql.Tx, log AuditLog, signingSecret string) error {
	return writeSQLAuditLog(ctx, tx, log, true, signingSecret)
}

func writeSQLAuditLog(ctx context.Context, tx *sql.Tx, log AuditLog, requireInsert bool, signingSecret string) error {
	log.ID = strings.TrimSpace(log.ID)
	log.ActorType = strings.TrimSpace(log.ActorType)
	log.ActorID = strings.TrimSpace(log.ActorID)
	log.Action = strings.TrimSpace(log.Action)
	log.TargetType = strings.TrimSpace(log.TargetType)
	log.TargetID = strings.TrimSpace(log.TargetID)
	log.RequestID = strings.TrimSpace(log.RequestID)
	log.IPHash = strings.TrimSpace(log.IPHash)
	if log.ID == "" || log.ActorType == "" || log.ActorID == "" || log.Action == "" || log.TargetType == "" || log.TargetID == "" {
		return ErrInvalidArgument
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = normalizeAuditLogTime(time.Now())
	} else {
		log.CreatedAt = normalizeAuditLogTime(log.CreatedAt)
	}
	ensureAuditLogIntegrity(&log, signingSecret)
	payload, err := json.Marshal(sanitizeAuditPayload(log.Payload))
	if err != nil {
		return err
	}
	conflictClause := `ON CONFLICT (id) DO UPDATE SET
  integrity_algorithm = EXCLUDED.integrity_algorithm,
  integrity_hash = EXCLUDED.integrity_hash
WHERE (audit_logs.integrity_algorithm = '' OR audit_logs.integrity_hash = '')
  AND audit_logs.actor_type = EXCLUDED.actor_type
  AND audit_logs.actor_id = EXCLUDED.actor_id
  AND audit_logs.action = EXCLUDED.action
  AND audit_logs.target_type = EXCLUDED.target_type
  AND audit_logs.target_id = EXCLUDED.target_id
  AND audit_logs.request_id = EXCLUDED.request_id
  AND audit_logs.ip_hash = EXCLUDED.ip_hash
  AND audit_logs.payload = EXCLUDED.payload
  AND audit_logs.created_at = EXCLUDED.created_at`
	if requireInsert {
		conflictClause = `ON CONFLICT (id) DO NOTHING`
	}
	query := fmt.Sprintf(`
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, target_type, target_id,
  request_id, ip_hash, payload, integrity_algorithm, integrity_hash, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12)
%s`, conflictClause)
	result, err := tx.ExecContext(ctx, query,
		log.ID,
		log.ActorType,
		log.ActorID,
		log.Action,
		log.TargetType,
		log.TargetID,
		log.RequestID,
		log.IPHash,
		string(payload),
		log.IntegrityAlgorithm,
		log.IntegrityHash,
		log.CreatedAt,
	)
	if err != nil {
		return err
	}
	if requireInsert {
		rowsAffected, err := result.RowsAffected()
		if err == nil && rowsAffected == 0 {
			return fmt.Errorf("audit log %s already exists", log.ID)
		}
	}
	return nil
}

func buildSQLAuditLogsQuery(req AuditLogsRequest) (string, []any) {
	req = normalizeAuditLogsRequest(req)
	filters := []string{"1 = 1"}
	args := []any{}
	addFilter := func(column string, value any) {
		args = append(args, value)
		filters = append(filters, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if req.ActorType != "" {
		addFilter("actor_type", req.ActorType)
	}
	if req.ActorID != "" {
		addFilter("actor_id", req.ActorID)
	}
	if req.Action != "" {
		addFilter("action", req.Action)
	}
	if req.TargetType != "" {
		addFilter("target_type", req.TargetType)
	}
	if req.TargetID != "" {
		addFilter("target_id", req.TargetID)
	}
	if !req.After.IsZero() {
		args = append(args, req.After.UTC())
		filters = append(filters, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if !req.Before.IsZero() {
		args = append(args, req.Before.UTC())
		filters = append(filters, fmt.Sprintf("created_at < $%d", len(args)))
	}
	args = append(args, req.Limit)
	query := fmt.Sprintf(`
SELECT %s
FROM audit_logs
WHERE %s
ORDER BY created_at DESC, id DESC
LIMIT $%d`, postgresAuditLogColumns, strings.Join(filters, " AND "), len(args))
	return query, args
}

func (s *PostgresStore) loadSQLAuditLogs(ctx context.Context, req AuditLogsRequest) ([]AuditLog, error) {
	query, args := buildSQLAuditLogsQuery(req)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []AuditLog{}
	for rows.Next() {
		var log AuditLog
		var payload []byte
		if err := rows.Scan(
			&log.ID,
			&log.ActorType,
			&log.ActorID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&log.RequestID,
			&log.IPHash,
			&payload,
			&log.IntegrityAlgorithm,
			&log.IntegrityHash,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &log.Payload); err != nil {
				return nil, err
			}
		}
		if log.Payload == nil {
			log.Payload = map[string]any{}
		}
		log.Payload = sanitizeAuditPayload(log.Payload)
		log.CreatedAt = normalizeAuditLogTime(log.CreatedAt)
		log.IntegrityVerified = verifyAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *PostgresStore) loadSQLAuditRetentionReport(ctx context.Context, req AuditRetentionReportRequest) (*AuditRetentionReport, error) {
	req = normalizeAuditRetentionReportRequest(req)
	retentionCutoff := req.Now.AddDate(0, 0, -req.RetentionDays)
	coldArchiveCutoff := req.Now.AddDate(0, 0, -req.HotDays)

	var totalLogs int64
	var expiredLogs int64
	var coldArchiveDueLogs int64
	var exportEvents int64
	var oldestCreatedAt sql.NullTime
	var newestCreatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT
  COUNT(*),
  MIN(created_at),
  MAX(created_at),
  COUNT(*) FILTER (WHERE created_at < $1),
  COUNT(*) FILTER (WHERE created_at < $2),
  COUNT(*) FILTER (WHERE action = 'admin.audit_logs.exported')
FROM audit_logs`, retentionCutoff, coldArchiveCutoff).Scan(
		&totalLogs,
		&oldestCreatedAt,
		&newestCreatedAt,
		&expiredLogs,
		&coldArchiveDueLogs,
		&exportEvents,
	)
	if err != nil {
		return nil, err
	}

	report := &AuditRetentionReport{
		Status:                 "ok",
		GeneratedAt:            req.Now,
		RetentionDays:          req.RetentionDays,
		HotDays:                req.HotDays,
		RetentionCutoff:        retentionCutoff,
		ColdArchiveCutoff:      coldArchiveCutoff,
		TotalLogs:              int(totalLogs),
		ExpiredLogs:            int(expiredLogs),
		ColdArchiveDueLogs:     int(coldArchiveDueLogs),
		ExportEvents:           int(exportEvents),
		CriticalActionCoverage: []AuditActionCoverage{},
		Alerts:                 []AuditRetentionAlert{},
	}
	if oldestCreatedAt.Valid {
		report.OldestCreatedAt = normalizeAuditLogTime(oldestCreatedAt.Time)
	}
	if newestCreatedAt.Valid {
		report.NewestCreatedAt = normalizeAuditLogTime(newestCreatedAt.Time)
	}

	for _, action := range req.CriticalActions {
		var count int64
		var lastCreatedAt sql.NullTime
		if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*), MAX(created_at)
FROM audit_logs
WHERE action = $1`, action).Scan(&count, &lastCreatedAt); err != nil {
			return nil, err
		}
		coverage := AuditActionCoverage{Action: action, Count: int(count)}
		if lastCreatedAt.Valid {
			coverage.LastCreatedAt = normalizeAuditLogTime(lastCreatedAt.Time)
		}
		report.CriticalActionCoverage = append(report.CriticalActionCoverage, coverage)
		if coverage.Count == 0 {
			report.MissingCriticalActions = append(report.MissingCriticalActions, action)
		}
	}

	sampleLogs, err := s.loadSQLAuditLogs(ctx, AuditLogsRequest{Limit: req.IntegritySampleLimit})
	if err != nil {
		return nil, err
	}
	report.IntegritySampleSize = len(sampleLogs)
	for _, log := range sampleLogs {
		if !log.IntegrityVerified {
			report.IntegrityFailures++
		}
	}

	report.Alerts = auditRetentionAlertsForReport(report)
	report.Status = auditRetentionStatus(report.Alerts)
	return report, nil
}

func upsertSQLUser(ctx context.Context, tx *sql.Tx, user AppUser) error {
	user.ID = strings.TrimSpace(user.ID)
	if user.ID == "" {
		return nil
	}
	user.Nickname = strings.TrimSpace(user.Nickname)
	user.AvatarURL = strings.TrimSpace(user.AvatarURL)
	user.Status = strings.TrimSpace(user.Status)
	if user.Status == "" {
		user.Status = "active"
	}
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	} else {
		user.CreatedAt = user.CreatedAt.UTC()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	} else {
		user.UpdatedAt = user.UpdatedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO app_users (id, nickname, avatar_url, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET nickname = EXCLUDED.nickname,
    avatar_url = EXCLUDED.avatar_url,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at`,
		user.ID, user.Nickname, user.AvatarURL, user.Status, user.CreatedAt, user.UpdatedAt)
	return err
}

func upsertSQLMerchant(ctx context.Context, tx *sql.Tx, merchant MerchantAccount) error {
	merchant.ID = strings.TrimSpace(merchant.ID)
	if merchant.ID == "" {
		return nil
	}
	merchant.Type = strings.TrimSpace(merchant.Type)
	if !isMerchantAccountType(merchant.Type) {
		merchant.Type = MerchantAccountStandard
	}
	merchant.DisplayName = strings.TrimSpace(merchant.DisplayName)
	if merchant.DisplayName == "" {
		merchant.DisplayName = merchant.ID
	}
	merchant.DepositStatus = strings.TrimSpace(merchant.DepositStatus)
	if merchant.DepositStatus == "" {
		merchant.DepositStatus = DepositStatusUnpaid
	}
	operationStatus := strings.TrimSpace(merchant.Status)
	if operationStatus == "" {
		operationStatus = "pending_review"
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO merchant_accounts (id, type, display_name, deposit_status, operation_status, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (id) DO UPDATE
SET type = EXCLUDED.type,
    display_name = EXCLUDED.display_name,
    deposit_status = EXCLUDED.deposit_status,
    operation_status = EXCLUDED.operation_status,
    updated_at = now()`,
		merchant.ID, merchant.Type, merchant.DisplayName, merchant.DepositStatus, operationStatus)
	return err
}

func upsertSQLMerchantQualification(ctx context.Context, tx *sql.Tx, snapshot merchantQualificationSnapshot) error {
	merchantID := strings.TrimSpace(snapshot.MerchantID)
	qualification := snapshot.Qualification
	qualification.Type = strings.TrimSpace(qualification.Type)
	qualification.FileURL = strings.TrimSpace(qualification.FileURL)
	qualification.Status = strings.TrimSpace(qualification.Status)
	if merchantID == "" || qualification.FileURL == "" || !isMerchantQualificationType(qualification.Type) || qualification.ExpiresAt.IsZero() {
		return nil
	}
	if qualification.Status == "" {
		qualification.Status = "pending_review"
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO merchant_qualifications (
  merchant_id, type, file_url, expires_at, status, updated_at
)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (merchant_id, type) DO UPDATE
SET file_url = EXCLUDED.file_url,
    expires_at = EXCLUDED.expires_at,
    status = EXCLUDED.status,
    updated_at = now()`,
		merchantID,
		qualification.Type,
		qualification.FileURL,
		qualification.ExpiresAt.UTC(),
		qualification.Status,
	)
	return err
}

func upsertSQLShop(ctx context.Context, tx *sql.Tx, shop Shop) error {
	shop.ID = strings.TrimSpace(shop.ID)
	shop.MerchantID = strings.TrimSpace(shop.MerchantID)
	if shop.ID == "" || shop.MerchantID == "" {
		return nil
	}
	shop.Name = strings.TrimSpace(shop.Name)
	if shop.Name == "" {
		shop.Name = shop.ID
	}
	shop.Category = strings.TrimSpace(shop.Category)
	if shop.Category == "" {
		shop.Category = "restaurant"
	}
	shop.AccountType = strings.TrimSpace(shop.AccountType)
	if !isMerchantAccountType(shop.AccountType) {
		shop.AccountType = MerchantAccountStandard
	}
	status := strings.TrimSpace(shop.Status)
	if status == "" {
		status = ShopStatusActive
	}
	serviceState := ShopServiceClosed
	if status == ShopStatusActive {
		serviceState = ShopServiceOpen
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO shops (
  id, merchant_id, name, category, account_type, capabilities, operation_status,
  service_state, cover_url, logo_url, announcement, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6::text[], $7, $8, $9, $10, $11, now())
ON CONFLICT (id) DO UPDATE
SET merchant_id = EXCLUDED.merchant_id,
    name = EXCLUDED.name,
    category = EXCLUDED.category,
    account_type = EXCLUDED.account_type,
    capabilities = EXCLUDED.capabilities,
    operation_status = EXCLUDED.operation_status,
    service_state = EXCLUDED.service_state,
    cover_url = EXCLUDED.cover_url,
    logo_url = EXCLUDED.logo_url,
    announcement = EXCLUDED.announcement,
    updated_at = now()`,
		shop.ID, shop.MerchantID, shop.Name, shop.Category, shop.AccountType, postgresTextArrayLiteral(shop.Capabilities), status, serviceState, shop.CoverURL, shop.LogoURL, shop.Announcement)
	return err
}

func upsertSQLMerchantProduct(ctx context.Context, tx *sql.Tx, product MerchantProduct, productType string) error {
	product.ID = strings.TrimSpace(product.ID)
	product.ShopID = strings.TrimSpace(product.ShopID)
	if product.ID == "" || product.ShopID == "" {
		return nil
	}
	productType = strings.TrimSpace(productType)
	if productType == "" {
		productType = "takeout"
	}
	product.Name = strings.TrimSpace(product.Name)
	if product.Name == "" {
		product.Name = product.ID
	}
	product.Status = strings.TrimSpace(product.Status)
	if product.Status == "" {
		product.Status = ProductStatusActive
	}
	if product.Status != ProductStatusActive && product.Status != ProductStatusSoldOut && product.Status != ProductStatusRemoved {
		product.Status = ProductStatusRemoved
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO merchant_products (
  id, shop_id, type, name, image_url, description, ingredient_list,
  price_fen, stock_count, status, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::text[], $8, $9, $10, now())
ON CONFLICT (id) DO UPDATE
SET shop_id = EXCLUDED.shop_id,
    type = EXCLUDED.type,
    name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    description = EXCLUDED.description,
    ingredient_list = EXCLUDED.ingredient_list,
    price_fen = EXCLUDED.price_fen,
    stock_count = EXCLUDED.stock_count,
    status = EXCLUDED.status,
    updated_at = now()`,
		product.ID,
		product.ShopID,
		productType,
		product.Name,
		strings.TrimSpace(product.ImageURL),
		strings.TrimSpace(product.Description),
		postgresTextArrayLiteral(product.IngredientList),
		nonNegativeInt64(product.PriceFen),
		nonNegativeInt(product.StockCount),
		product.Status,
	)
	return err
}

func upsertSQLCartItem(ctx context.Context, tx *sql.Tx, item CartItem) error {
	item.UserID = strings.TrimSpace(item.UserID)
	item.ShopID = strings.TrimSpace(item.ShopID)
	item.ProductID = strings.TrimSpace(item.ProductID)
	if item.UserID == "" || item.ShopID == "" || item.ProductID == "" || item.Quantity <= 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO cart_items (
  user_id, shop_id, product_id, quantity, selected,
  unit_price_fen, product_name_snapshot, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
ON CONFLICT (user_id, shop_id, product_id) DO UPDATE
SET quantity = EXCLUDED.quantity,
    selected = EXCLUDED.selected,
    unit_price_fen = EXCLUDED.unit_price_fen,
    product_name_snapshot = EXCLUDED.product_name_snapshot,
    updated_at = now()`,
		item.UserID,
		item.ShopID,
		item.ProductID,
		item.Quantity,
		item.Selected,
		nonNegativeInt64(item.UnitPriceFen),
		strings.TrimSpace(item.ProductName),
	)
	return err
}

func upsertSQLRider(ctx context.Context, tx *sql.Tx, rider RiderAccount) error {
	rider.ID = strings.TrimSpace(rider.ID)
	if rider.ID == "" {
		return nil
	}
	rider.Type = strings.TrimSpace(rider.Type)
	if rider.Type != RiderAccountStationManager && rider.Type != RiderAccountRider {
		rider.Type = RiderAccountRider
	}
	rider.Status = strings.TrimSpace(rider.Status)
	if rider.Status == "" {
		rider.Status = "active"
	}
	rider.DepositStatus = strings.TrimSpace(rider.DepositStatus)
	if rider.DepositStatus == "" {
		rider.DepositStatus = DepositStatusUnpaid
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO rider_accounts (id, station_id, type, status, online, deposit_status, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (id) DO UPDATE
SET station_id = EXCLUDED.station_id,
    type = EXCLUDED.type,
    status = EXCLUDED.status,
    online = EXCLUDED.online,
    deposit_status = EXCLUDED.deposit_status,
    updated_at = now()`,
		rider.ID, strings.TrimSpace(rider.StationID), rider.Type, rider.Status, rider.Online, rider.DepositStatus)
	return err
}

func upsertSQLAddress(ctx context.Context, tx *sql.Tx, address UserAddress) error {
	address.ID = strings.TrimSpace(address.ID)
	address.UserID = strings.TrimSpace(address.UserID)
	if address.ID == "" || address.UserID == "" {
		return nil
	}
	latitude := 0.0
	longitude := 0.0
	if address.Latitude != nil {
		latitude = *address.Latitude
	}
	if address.Longitude != nil {
		longitude = *address.Longitude
	}
	tag := strings.TrimSpace(address.Tag)
	if tag == "" {
		tag = "other"
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO user_addresses (
  id, user_id, contact_name, contact_phone_hash, city, detail,
  latitude, longitude, tag, is_default, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
ON CONFLICT (id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    contact_name = EXCLUDED.contact_name,
    contact_phone_hash = EXCLUDED.contact_phone_hash,
    city = EXCLUDED.city,
    detail = EXCLUDED.detail,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    tag = EXCLUDED.tag,
    is_default = EXCLUDED.is_default,
    updated_at = now()`,
		address.ID,
		address.UserID,
		strings.TrimSpace(address.ContactName),
		hashOptional(address.ContactPhone),
		strings.TrimSpace(address.City),
		strings.TrimSpace(address.Detail),
		latitude,
		longitude,
		tag,
		address.IsDefault,
	)
	return err
}

func upsertSQLOrder(ctx context.Context, tx *sql.Tx, order Order, snapshot paymentDomainSnapshot) error {
	order.ID = strings.TrimSpace(order.ID)
	order.UserID = strings.TrimSpace(order.UserID)
	if order.ID == "" || order.UserID == "" || !IsOrderType(order.Type) {
		return nil
	}
	if order.Status == "" {
		order.Status = StatusPendingPayment
	}
	now := time.Now().UTC()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	} else {
		order.CreatedAt = order.CreatedAt.UTC()
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = order.CreatedAt
	} else {
		order.UpdatedAt = order.UpdatedAt.UTC()
	}
	shopByID := map[string]Shop{}
	for _, shop := range snapshot.Shops {
		shopByID[shop.ID] = shop
	}
	riderByID := map[string]bool{}
	for _, rider := range snapshot.Riders {
		riderByID[rider.ID] = true
	}
	addressByID := map[string]bool{}
	for _, address := range snapshot.Addresses {
		addressByID[address.ID] = true
	}
	shopID := strings.TrimSpace(order.ShopID)
	merchantID := ""
	if shopID != "" {
		if shop, ok := shopByID[shopID]; ok {
			merchantID = strings.TrimSpace(shop.MerchantID)
		} else {
			shopID = ""
		}
	}
	riderID := strings.TrimSpace(order.RiderID)
	if riderID != "" && !riderByID[riderID] {
		riderID = ""
	}
	addressID := strings.TrimSpace(order.AddressID)
	if addressID != "" && !addressByID[addressID] {
		addressID = ""
	}
	optionsPayload, err := json.Marshal(order.Options)
	if err != nil {
		return err
	}
	pricingPayload, err := json.Marshal(map[string]any{
		"items_total_fen":   order.ItemsTotalFen,
		"delivery_fee_fen":  order.DeliveryFeeFen,
		"packaging_fee_fen": order.PackagingFeeFen,
		"discount_fen":      order.DiscountFen,
		"amount_fen":        order.AmountFen,
	})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO orders (
  id, user_id, merchant_id, shop_id, rider_id, address_id, type, status,
  payment_method, items_total_fen, delivery_fee_fen, packaging_fee_fen,
  discount_fen, amount_fen, options, pricing_snapshot, dispatch_mode,
  paid_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15::jsonb, $16::jsonb, $17, $18, $19, $20)
ON CONFLICT (id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    merchant_id = EXCLUDED.merchant_id,
    shop_id = EXCLUDED.shop_id,
    rider_id = EXCLUDED.rider_id,
    address_id = EXCLUDED.address_id,
    type = EXCLUDED.type,
    status = EXCLUDED.status,
    payment_method = EXCLUDED.payment_method,
    items_total_fen = EXCLUDED.items_total_fen,
    delivery_fee_fen = EXCLUDED.delivery_fee_fen,
    packaging_fee_fen = EXCLUDED.packaging_fee_fen,
    discount_fen = EXCLUDED.discount_fen,
    amount_fen = EXCLUDED.amount_fen,
    options = EXCLUDED.options,
    pricing_snapshot = EXCLUDED.pricing_snapshot,
    dispatch_mode = EXCLUDED.dispatch_mode,
    paid_at = EXCLUDED.paid_at,
    updated_at = EXCLUDED.updated_at`,
		order.ID,
		order.UserID,
		nullableString(merchantID),
		nullableString(shopID),
		nullableString(riderID),
		nullableString(addressID),
		order.Type,
		order.Status,
		strings.TrimSpace(order.PaymentMethod),
		nonNegativeInt64(order.ItemsTotalFen),
		nonNegativeInt64(order.DeliveryFeeFen),
		nonNegativeInt64(order.PackagingFeeFen),
		nonNegativeInt64(order.DiscountFen),
		nonNegativeInt64(order.AmountFen),
		string(optionsPayload),
		string(pricingPayload),
		DispatchModeGrabHall,
		nullableTime(orderPaidAt(order)),
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", order.ID); err != nil {
		return err
	}
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			continue
		}
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" {
			productID = "unknown"
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO order_items (order_id, product_id, product_name_snapshot, unit_price_fen, quantity)
VALUES ($1, $2, $3, $4, $5)`,
			order.ID,
			productID,
			strings.TrimSpace(item.ProductName),
			nonNegativeInt64(item.UnitPriceFen),
			item.Quantity,
		)
		if err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM order_events WHERE order_id = $1", order.ID); err != nil {
		return err
	}
	for _, event := range order.Events {
		event.Type = strings.TrimSpace(event.Type)
		if event.Type == "" {
			continue
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = order.UpdatedAt
		} else {
			event.CreatedAt = event.CreatedAt.UTC()
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO order_events (order_id, type, actor_id, message, created_at)
VALUES ($1, $2, $3, $4, $5)`,
			order.ID,
			event.Type,
			strings.TrimSpace(event.ActorID),
			strings.TrimSpace(event.Message),
			event.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertSQLWalletAccount(ctx context.Context, tx *sql.Tx, wallet WalletAccount) error {
	wallet.UserID = strings.TrimSpace(wallet.UserID)
	if wallet.UserID == "" {
		return nil
	}
	wallet.RiskState = strings.TrimSpace(wallet.RiskState)
	if wallet.RiskState == "" {
		wallet.RiskState = "normal"
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO wallet_accounts (subject_type, subject_id, balance_fen, frozen_fen, version, risk_state, updated_at)
VALUES ('user', $1, $2, $3, $4, $5, now())
ON CONFLICT (subject_type, subject_id) DO UPDATE
SET balance_fen = EXCLUDED.balance_fen,
    frozen_fen = EXCLUDED.frozen_fen,
    version = EXCLUDED.version,
    risk_state = EXCLUDED.risk_state,
    updated_at = now()`,
		wallet.UserID,
		nonNegativeInt64(wallet.Balance),
		nonNegativeInt64(wallet.Frozen),
		wallet.Version,
		wallet.RiskState,
	)
	return err
}

func upsertSQLWalletTransaction(ctx context.Context, tx *sql.Tx, transaction WalletTransaction) error {
	transaction.ID = strings.TrimSpace(transaction.ID)
	transaction.UserID = strings.TrimSpace(transaction.UserID)
	transaction.IdempotencyKey = strings.TrimSpace(transaction.IdempotencyKey)
	if transaction.ID == "" || transaction.UserID == "" || transaction.IdempotencyKey == "" {
		return nil
	}
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	} else {
		transaction.CreatedAt = transaction.CreatedAt.UTC()
	}
	if transaction.Status == "" {
		transaction.Status = "success"
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO wallet_transactions (
  id, subject_type, subject_id, order_id, type, amount_fen,
  payment_method, idempotency_key, status, created_at
)
VALUES ($1, 'user', $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET subject_id = EXCLUDED.subject_id,
    order_id = EXCLUDED.order_id,
    type = EXCLUDED.type,
    amount_fen = EXCLUDED.amount_fen,
    payment_method = EXCLUDED.payment_method,
    idempotency_key = EXCLUDED.idempotency_key,
    status = EXCLUDED.status,
    created_at = EXCLUDED.created_at`,
		transaction.ID,
		transaction.UserID,
		nullableString(transaction.OrderID),
		strings.TrimSpace(transaction.Type),
		transaction.AmountFen,
		strings.TrimSpace(transaction.PaymentMethod),
		transaction.IdempotencyKey,
		strings.TrimSpace(transaction.Status),
		transaction.CreatedAt,
	)
	return err
}

func upsertSQLWalletPaymentPassword(ctx context.Context, tx *sql.Tx, userID string, passwordHash string) error {
	userID = strings.TrimSpace(userID)
	passwordHash = strings.TrimSpace(passwordHash)
	if userID == "" || passwordHash == "" {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO wallet_payment_passwords (user_id, password_hash, status, failed_count, updated_at)
VALUES ($1, $2, $3, 0, now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    updated_at = now()`,
		userID,
		passwordHash,
		WalletPaymentPasswordSet,
	)
	return err
}

func upsertSQLPaymentTransaction(ctx context.Context, tx *sql.Tx, transaction PaymentTransaction) error {
	transaction.ID = strings.TrimSpace(transaction.ID)
	transaction.UserID = strings.TrimSpace(transaction.UserID)
	transaction.OrderID = strings.TrimSpace(transaction.OrderID)
	transaction.OutTradeNo = strings.TrimSpace(transaction.OutTradeNo)
	transaction.IdempotencyKey = strings.TrimSpace(transaction.IdempotencyKey)
	if transaction.ID == "" || transaction.UserID == "" || transaction.OrderID == "" || transaction.OutTradeNo == "" || transaction.IdempotencyKey == "" {
		return nil
	}
	if transaction.Method != PaymentWechat && transaction.Method != PaymentBalance {
		return nil
	}
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	} else {
		transaction.CreatedAt = transaction.CreatedAt.UTC()
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = transaction.CreatedAt
	} else {
		transaction.UpdatedAt = transaction.UpdatedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO payment_transactions (
  id, order_id, user_id, method, amount_fen, status, out_trade_no,
  transaction_id, idempotency_key, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE
SET order_id = EXCLUDED.order_id,
    user_id = EXCLUDED.user_id,
    method = EXCLUDED.method,
    amount_fen = EXCLUDED.amount_fen,
    status = EXCLUDED.status,
    out_trade_no = EXCLUDED.out_trade_no,
    transaction_id = EXCLUDED.transaction_id,
    idempotency_key = EXCLUDED.idempotency_key,
    updated_at = EXCLUDED.updated_at`,
		transaction.ID,
		transaction.OrderID,
		transaction.UserID,
		transaction.Method,
		nonNegativeInt64(transaction.AmountFen),
		strings.TrimSpace(transaction.Status),
		transaction.OutTradeNo,
		nullableString(transaction.TransactionID),
		transaction.IdempotencyKey,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)
	return err
}

func upsertSQLRefundSettings(ctx context.Context, tx *sql.Tx, settings RefundSettings) error {
	settings = normalizeStoredRefundSettings(settings)
	_, err := tx.ExecContext(ctx, `
INSERT INTO refund_settings (id, default_strategy, updated_at)
VALUES ('default', $1, now())
ON CONFLICT (id) DO UPDATE
SET default_strategy = EXCLUDED.default_strategy,
    updated_at = now()`,
		settings.DefaultStrategy,
	)
	return err
}

func upsertSQLRefundTransaction(ctx context.Context, tx *sql.Tx, refund RefundTransaction) error {
	refund.ID = strings.TrimSpace(refund.ID)
	refund.OrderID = strings.TrimSpace(refund.OrderID)
	refund.UserID = strings.TrimSpace(refund.UserID)
	refund.IdempotencyKey = strings.TrimSpace(refund.IdempotencyKey)
	refund.Destination = RefundDestinationForStrategy(RefundStrategyBalanceFirst, refund.Destination)
	refund.Status = strings.TrimSpace(refund.Status)
	refund.Reason = strings.TrimSpace(refund.Reason)
	if refund.ID == "" || refund.OrderID == "" || refund.UserID == "" || refund.IdempotencyKey == "" || refund.AmountFen <= 0 || refund.Status == "" || refund.Reason == "" {
		return nil
	}
	if refund.CreatedAt.IsZero() {
		refund.CreatedAt = time.Now().UTC()
	} else {
		refund.CreatedAt = refund.CreatedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO refund_transactions (
  id, order_id, user_id, amount_fen, destination, status,
  reason, idempotency_key, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
ON CONFLICT (id) DO UPDATE
SET order_id = EXCLUDED.order_id,
    user_id = EXCLUDED.user_id,
    amount_fen = EXCLUDED.amount_fen,
    destination = EXCLUDED.destination,
    status = EXCLUDED.status,
    reason = EXCLUDED.reason,
    idempotency_key = EXCLUDED.idempotency_key,
    updated_at = now()`,
		refund.ID,
		refund.OrderID,
		refund.UserID,
		nonNegativeInt64(refund.AmountFen),
		refund.Destination,
		refund.Status,
		refund.Reason,
		refund.IdempotencyKey,
		refund.CreatedAt,
	)
	return err
}

func upsertSQLAfterSalesRequest(ctx context.Context, tx *sql.Tx, request AfterSalesRequest) error {
	request.ID = strings.TrimSpace(request.ID)
	request.OrderID = strings.TrimSpace(request.OrderID)
	request.UserID = strings.TrimSpace(request.UserID)
	request.Type = normalizeAfterSalesType(request.Type)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Status = strings.TrimSpace(request.Status)
	if request.ID == "" || request.OrderID == "" || request.UserID == "" || request.Type == "" || request.Reason == "" || request.RequestedAmountFen <= 0 || request.Status == "" {
		return nil
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now().UTC()
	} else {
		request.CreatedAt = request.CreatedAt.UTC()
	}
	if request.UpdatedAt.IsZero() {
		request.UpdatedAt = request.CreatedAt
	} else {
		request.UpdatedAt = request.UpdatedAt.UTC()
	}
	if !request.ReviewedAt.IsZero() {
		request.ReviewedAt = request.ReviewedAt.UTC()
	}
	evidencePayload, err := json.Marshal(sanitizedStringSlice(request.EvidenceURLs))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO order_after_sales (
  id, order_id, user_id, type, reason, requested_amount_fen,
  evidence_urls, status, review_reason, reviewer_id, reviewer_role,
  refund_id, created_at, updated_at, reviewed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE
SET order_id = EXCLUDED.order_id,
    user_id = EXCLUDED.user_id,
    type = EXCLUDED.type,
    reason = EXCLUDED.reason,
    requested_amount_fen = EXCLUDED.requested_amount_fen,
    evidence_urls = EXCLUDED.evidence_urls,
    status = EXCLUDED.status,
    review_reason = EXCLUDED.review_reason,
    reviewer_id = EXCLUDED.reviewer_id,
    reviewer_role = EXCLUDED.reviewer_role,
    refund_id = EXCLUDED.refund_id,
    updated_at = EXCLUDED.updated_at,
    reviewed_at = EXCLUDED.reviewed_at`,
		request.ID,
		request.OrderID,
		request.UserID,
		request.Type,
		request.Reason,
		nonNegativeInt64(request.RequestedAmountFen),
		string(evidencePayload),
		request.Status,
		strings.TrimSpace(request.ReviewReason),
		strings.TrimSpace(request.ReviewerID),
		strings.TrimSpace(request.ReviewerRole),
		nullableString(request.RefundID),
		request.CreatedAt,
		request.UpdatedAt,
		nullableTime(request.ReviewedAt),
	)
	return err
}

func upsertSQLAfterSalesEvent(ctx context.Context, tx *sql.Tx, event AfterSalesEvent) error {
	event.ID = strings.TrimSpace(event.ID)
	event.RequestID = strings.TrimSpace(event.RequestID)
	event.OrderID = strings.TrimSpace(event.OrderID)
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.ActorRole = strings.TrimSpace(event.ActorRole)
	event.Action = normalizeAfterSalesAction(event.Action)
	event.Message = strings.TrimSpace(event.Message)
	if event.ID == "" || event.RequestID == "" || event.OrderID == "" || event.ActorID == "" || event.ActorRole == "" || event.Action == "" || event.Message == "" {
		return nil
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}
	attachmentsPayload, err := json.Marshal(sanitizedStringSlice(event.Attachments))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO order_after_sales_events (
  id, request_id, order_id, actor_id, actor_role, action,
  message, attachments, visible_to_user, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10)
ON CONFLICT (id) DO UPDATE
SET request_id = EXCLUDED.request_id,
    order_id = EXCLUDED.order_id,
    actor_id = EXCLUDED.actor_id,
    actor_role = EXCLUDED.actor_role,
    action = EXCLUDED.action,
    message = EXCLUDED.message,
    attachments = EXCLUDED.attachments,
    visible_to_user = EXCLUDED.visible_to_user,
    created_at = EXCLUDED.created_at`,
		event.ID,
		event.RequestID,
		event.OrderID,
		event.ActorID,
		event.ActorRole,
		event.Action,
		event.Message,
		string(attachmentsPayload),
		event.VisibleToUser,
		event.CreatedAt,
	)
	return err
}

func upsertSQLAfterSalesUploadTicket(ctx context.Context, tx *sql.Tx, ticket AfterSalesEvidenceUploadTicket) error {
	ticket.ID = strings.TrimSpace(ticket.ID)
	ticket.RequestID = strings.TrimSpace(ticket.RequestID)
	ticket.OrderID = strings.TrimSpace(ticket.OrderID)
	ticket.Provider = strings.TrimSpace(ticket.Provider)
	if ticket.Provider == "" {
		ticket.Provider = ObjectStorageProviderMinIO
	}
	ticket.Bucket = strings.Trim(strings.TrimSpace(ticket.Bucket), "/")
	ticket.ObjectKey = strings.TrimSpace(ticket.ObjectKey)
	ticket.PublicURL = strings.TrimSpace(ticket.PublicURL)
	ticket.FileName = sanitizeObjectFileName(ticket.FileName)
	ticket.ContentType = normalizeEvidenceContentType(ticket.ContentType)
	ticket.ContentSHA = strings.TrimSpace(ticket.ContentSHA)
	ticket.UploadedByID = strings.TrimSpace(ticket.UploadedByID)
	ticket.UploadedByRole = strings.TrimSpace(ticket.UploadedByRole)
	ticket.Status = normalizeAfterSalesUploadTicketStatus(ticket.Status)
	ticket.ScanStatus = normalizeAfterSalesUploadScanStatus(ticket.ScanStatus)
	ticket.ScanResult = strings.TrimSpace(ticket.ScanResult)
	ticket.CleanupReason = normalizeObjectStorageCleanupReason(ticket.CleanupReason)
	ticket.LastCleanupError = sanitizeObjectStorageCleanupError(ticket.LastCleanupError)
	if ticket.CleanupAttempts < 0 {
		ticket.CleanupAttempts = 0
	}
	if ticket.ID == "" || ticket.RequestID == "" || ticket.OrderID == "" || ticket.Bucket == "" || ticket.ObjectKey == "" || ticket.PublicURL == "" || ticket.FileName == "" || ticket.ContentType == "" || ticket.SizeBytes <= 0 || ticket.MaxSizeBytes <= 0 || ticket.UploadedByID == "" || ticket.UploadedByRole == "" || ticket.Status == "" || ticket.ExpiresAt.IsZero() {
		return nil
	}
	if ticket.ScanStatus == "" {
		return nil
	}
	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = time.Now().UTC()
	} else {
		ticket.CreatedAt = ticket.CreatedAt.UTC()
	}
	ticket.ExpiresAt = ticket.ExpiresAt.UTC()
	_, err := tx.ExecContext(ctx, `
	INSERT INTO order_after_sales_evidence_upload_tickets (
	  id, request_id, order_id, provider, bucket, object_key, public_url,
	  file_name, content_type, size_bytes, max_size_bytes, content_sha, uploaded_by_id,
	  uploaded_by_role, status, scan_status, scan_result, created_at, expires_at,
	  uploaded_at, confirmed_at, scan_checked_at, cleanup_reason, deleted_at,
	  cleanup_attempts, last_cleanup_error, last_cleanup_failed_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
	ON CONFLICT (id) DO UPDATE
	SET request_id = EXCLUDED.request_id,
	    order_id = EXCLUDED.order_id,
	    provider = EXCLUDED.provider,
	    bucket = EXCLUDED.bucket,
	    object_key = EXCLUDED.object_key,
	    public_url = EXCLUDED.public_url,
	    file_name = EXCLUDED.file_name,
	    content_type = EXCLUDED.content_type,
	    size_bytes = EXCLUDED.size_bytes,
	    max_size_bytes = EXCLUDED.max_size_bytes,
	    content_sha = EXCLUDED.content_sha,
	    uploaded_by_id = EXCLUDED.uploaded_by_id,
	    uploaded_by_role = EXCLUDED.uploaded_by_role,
	    status = EXCLUDED.status,
	    scan_status = EXCLUDED.scan_status,
	    scan_result = EXCLUDED.scan_result,
	    expires_at = EXCLUDED.expires_at,
	    uploaded_at = EXCLUDED.uploaded_at,
	    confirmed_at = EXCLUDED.confirmed_at,
	    scan_checked_at = EXCLUDED.scan_checked_at,
	    cleanup_reason = EXCLUDED.cleanup_reason,
	    deleted_at = EXCLUDED.deleted_at,
	    cleanup_attempts = EXCLUDED.cleanup_attempts,
	    last_cleanup_error = EXCLUDED.last_cleanup_error,
	    last_cleanup_failed_at = EXCLUDED.last_cleanup_failed_at`,
		ticket.ID,
		ticket.RequestID,
		ticket.OrderID,
		ticket.Provider,
		ticket.Bucket,
		ticket.ObjectKey,
		ticket.PublicURL,
		ticket.FileName,
		ticket.ContentType,
		ticket.SizeBytes,
		ticket.MaxSizeBytes,
		ticket.ContentSHA,
		ticket.UploadedByID,
		ticket.UploadedByRole,
		ticket.Status,
		ticket.ScanStatus,
		ticket.ScanResult,
		ticket.CreatedAt,
		ticket.ExpiresAt,
		nullableTime(ticket.UploadedAt),
		nullableTime(ticket.ConfirmedAt),
		nullableTime(ticket.ScanCheckedAt),
		ticket.CleanupReason,
		nullableTime(ticket.DeletedAt),
		ticket.CleanupAttempts,
		ticket.LastCleanupError,
		nullableTime(ticket.LastCleanupFailedAt),
	)
	return err
}

func upsertSQLAfterSalesEvidence(ctx context.Context, tx *sql.Tx, evidence AfterSalesEvidence) error {
	evidence.ID = strings.TrimSpace(evidence.ID)
	evidence.RequestID = strings.TrimSpace(evidence.RequestID)
	evidence.OrderID = strings.TrimSpace(evidence.OrderID)
	evidence.ObjectKey = strings.TrimSpace(evidence.ObjectKey)
	evidence.PublicURL = strings.TrimSpace(evidence.PublicURL)
	evidence.FileName = sanitizeObjectFileName(evidence.FileName)
	evidence.ContentType = normalizeEvidenceContentType(evidence.ContentType)
	evidence.ContentSHA = strings.TrimSpace(evidence.ContentSHA)
	evidence.UploadedByID = strings.TrimSpace(evidence.UploadedByID)
	evidence.UploadedByRole = strings.TrimSpace(evidence.UploadedByRole)
	evidence.Status = strings.TrimSpace(evidence.Status)
	if evidence.Status == "" {
		evidence.Status = AfterSalesEvidenceUploaded
	}
	if evidence.ID == "" || evidence.RequestID == "" || evidence.OrderID == "" || evidence.ObjectKey == "" || evidence.PublicURL == "" || evidence.FileName == "" || evidence.ContentType == "" || evidence.SizeBytes <= 0 || evidence.UploadedByID == "" || evidence.UploadedByRole == "" || evidence.Status != AfterSalesEvidenceUploaded {
		return nil
	}
	if evidence.CreatedAt.IsZero() {
		evidence.CreatedAt = time.Now().UTC()
	} else {
		evidence.CreatedAt = evidence.CreatedAt.UTC()
	}
	if evidence.ConfirmedAt.IsZero() {
		evidence.ConfirmedAt = evidence.CreatedAt
	} else {
		evidence.ConfirmedAt = evidence.ConfirmedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
	INSERT INTO order_after_sales_evidence (
	  id, request_id, order_id, object_key, public_url, file_name,
	  content_type, size_bytes, content_sha, uploaded_by_id, uploaded_by_role,
	  status, created_at, confirmed_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (id) DO UPDATE
	SET request_id = EXCLUDED.request_id,
	    order_id = EXCLUDED.order_id,
	    object_key = EXCLUDED.object_key,
	    public_url = EXCLUDED.public_url,
	    file_name = EXCLUDED.file_name,
	    content_type = EXCLUDED.content_type,
	    size_bytes = EXCLUDED.size_bytes,
	    content_sha = EXCLUDED.content_sha,
	    uploaded_by_id = EXCLUDED.uploaded_by_id,
	    uploaded_by_role = EXCLUDED.uploaded_by_role,
	    status = EXCLUDED.status,
	    confirmed_at = EXCLUDED.confirmed_at`,
		evidence.ID,
		evidence.RequestID,
		evidence.OrderID,
		evidence.ObjectKey,
		evidence.PublicURL,
		evidence.FileName,
		evidence.ContentType,
		evidence.SizeBytes,
		evidence.ContentSHA,
		evidence.UploadedByID,
		evidence.UploadedByRole,
		evidence.Status,
		evidence.CreatedAt,
		evidence.ConfirmedAt,
	)
	return err
}

func upsertSQLDispatchEvent(ctx context.Context, tx *sql.Tx, event DispatchEvent) error {
	event.ID = strings.TrimSpace(event.ID)
	event.OrderID = strings.TrimSpace(event.OrderID)
	event.Type = strings.TrimSpace(event.Type)
	event.IdempotencyKey = strings.TrimSpace(event.IdempotencyKey)
	if event.ID == "" || event.OrderID == "" || event.Type == "" || event.IdempotencyKey == "" {
		return nil
	}
	event.StationID = strings.TrimSpace(event.StationID)
	event.Mode = strings.TrimSpace(event.Mode)
	if event.Mode == "" {
		event.Mode = DispatchModeAutoAssign
	}
	event.RiderID = strings.TrimSpace(event.RiderID)
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.Reason = strings.TrimSpace(event.Reason)
	if event.OnlineCandidateSize < 0 {
		event.OnlineCandidateSize = 0
	}
	event.RejectedRiderIDs = normalizedStringSlice(event.RejectedRiderIDs)
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
INSERT INTO dispatch_events (
  id, order_id, station_id, mode, type, rider_id, actor_id, reason,
  idempotency_key, online_candidate_size, rejected_rider_ids,
  can_decline_without_penalty, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::text[], $12, $13)
ON CONFLICT (id) DO UPDATE
SET order_id = EXCLUDED.order_id,
    station_id = EXCLUDED.station_id,
    mode = EXCLUDED.mode,
    type = EXCLUDED.type,
    rider_id = EXCLUDED.rider_id,
    actor_id = EXCLUDED.actor_id,
    reason = EXCLUDED.reason,
    idempotency_key = EXCLUDED.idempotency_key,
    online_candidate_size = EXCLUDED.online_candidate_size,
    rejected_rider_ids = EXCLUDED.rejected_rider_ids,
    can_decline_without_penalty = EXCLUDED.can_decline_without_penalty,
    created_at = EXCLUDED.created_at`,
		event.ID,
		event.OrderID,
		event.StationID,
		event.Mode,
		event.Type,
		event.RiderID,
		event.ActorID,
		event.Reason,
		event.IdempotencyKey,
		event.OnlineCandidateSize,
		postgresTextArrayLiteral(event.RejectedRiderIDs),
		event.CanDeclineWithoutPenalty,
		event.CreatedAt,
	)
	return err
}

func (s *PostgresStore) payOrderWithBalanceInSQL(ctx context.Context, req BalancePayRequest) (WalletTransaction, WalletAccount, string, time.Time, error) {
	userID := strings.TrimSpace(req.UserID)
	orderID := strings.TrimSpace(req.OrderID)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if userID == "" || orderID == "" || idempotencyKey == "" {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrInvalidArgument
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := lockSQLWalletIdempotencyKey(ctx, tx, idempotencyKey); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	if existing, found, err := loadSQLWalletTransactionByIdempotency(ctx, tx, idempotencyKey); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	} else if found {
		if !walletTransactionMatchesBalancePayment(existing, userID, orderID) {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrInvalidOrderState
		}
		account, err := loadSQLWalletAccountForUpdate(ctx, tx, existing.UserID)
		if err != nil {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
		}
		if err := tx.Commit(); err != nil {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
		}
		return existing, account, existing.OrderID, existing.CreatedAt, nil
	}

	order, err := loadSQLOrderForBalancePayment(ctx, tx, orderID)
	if err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	if existing, found, err := loadSQLWalletTransactionByIdempotency(ctx, tx, idempotencyKey); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	} else if found {
		if !walletTransactionMatchesBalancePayment(existing, userID, orderID) {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrInvalidOrderState
		}
		account, err := loadSQLWalletAccountForUpdate(ctx, tx, existing.UserID)
		if err != nil {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
		}
		if err := tx.Commit(); err != nil {
			return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
		}
		return existing, account, existing.OrderID, existing.CreatedAt, nil
	}
	if order.UserID != userID || order.Status != StatusPendingPayment {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrInvalidOrderState
	}
	if ok, err := verifySQLWalletPaymentPassword(ctx, tx, userID, req.PaymentPassword); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	} else if !ok {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrPaymentPassword
	}

	account, err := ensureAndLockSQLWalletAccount(ctx, tx, userID)
	if err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	if account.Balance < order.AmountFen {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, ErrInsufficientBalance
	}

	now := time.Now().UTC()
	order.Status = statusAfterPayment(&order)
	order.PaymentMethod = PaymentBalance
	order.UpdatedAt = now
	if err := updateSQLOrderAfterBalancePayment(ctx, tx, order, userID, now); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	account.Balance -= order.AmountFen
	account.Version++
	if err := updateSQLWalletAccountBalance(ctx, tx, account); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	transaction := WalletTransaction{
		ID:             postgresWalletTransactionID(idempotencyKey),
		UserID:         userID,
		OrderID:        order.ID,
		Type:           "payment",
		AmountFen:      -order.AmountFen,
		PaymentMethod:  PaymentBalance,
		IdempotencyKey: idempotencyKey,
		Status:         "success",
		CreatedAt:      now,
	}
	if err := insertSQLWalletPaymentTransaction(ctx, tx, transaction, account.Balance); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	if err := tx.Commit(); err != nil {
		return WalletTransaction{}, WalletAccount{}, "", time.Time{}, err
	}
	return transaction, account, order.ID, now, nil
}

func lockSQLWalletIdempotencyKey(ctx context.Context, tx *sql.Tx, idempotencyKey string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1)::bigint)`, strings.TrimSpace(idempotencyKey))
	return err
}

func loadSQLWalletTransactionByIdempotency(ctx context.Context, tx *sql.Tx, idempotencyKey string) (WalletTransaction, bool, error) {
	var transaction WalletTransaction
	var orderID sql.NullString
	err := tx.QueryRowContext(ctx, `
SELECT id, subject_id, order_id, type, amount_fen, payment_method, idempotency_key, status, created_at
FROM wallet_transactions
WHERE idempotency_key = $1
FOR UPDATE`, strings.TrimSpace(idempotencyKey)).Scan(
		&transaction.ID,
		&transaction.UserID,
		&orderID,
		&transaction.Type,
		&transaction.AmountFen,
		&transaction.PaymentMethod,
		&transaction.IdempotencyKey,
		&transaction.Status,
		&transaction.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WalletTransaction{}, false, nil
	}
	if err != nil {
		return WalletTransaction{}, false, err
	}
	if orderID.Valid {
		transaction.OrderID = orderID.String
	}
	transaction.CreatedAt = transaction.CreatedAt.UTC()
	return transaction, true, nil
}

func walletTransactionMatchesBalancePayment(transaction WalletTransaction, userID string, orderID string) bool {
	return strings.TrimSpace(transaction.UserID) == strings.TrimSpace(userID) &&
		strings.TrimSpace(transaction.OrderID) == strings.TrimSpace(orderID) &&
		transaction.Type == "payment" &&
		transaction.PaymentMethod == PaymentBalance &&
		transaction.Status == "success"
}

func loadSQLOrderForBalancePayment(ctx context.Context, tx *sql.Tx, orderID string) (Order, error) {
	var order Order
	var shopID sql.NullString
	var riderID sql.NullString
	var addressID sql.NullString
	err := tx.QueryRowContext(ctx, `
SELECT id, user_id, shop_id, rider_id, address_id, type, status, payment_method,
       amount_fen, created_at, updated_at
FROM orders
WHERE id = $1
FOR UPDATE`, strings.TrimSpace(orderID)).Scan(
		&order.ID,
		&order.UserID,
		&shopID,
		&riderID,
		&addressID,
		&order.Type,
		&order.Status,
		&order.PaymentMethod,
		&order.AmountFen,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, err
	}
	if shopID.Valid {
		order.ShopID = shopID.String
	}
	if riderID.Valid {
		order.RiderID = riderID.String
	}
	if addressID.Valid {
		order.AddressID = addressID.String
	}
	order.CreatedAt = order.CreatedAt.UTC()
	order.UpdatedAt = order.UpdatedAt.UTC()
	return order, nil
}

func verifySQLWalletPaymentPassword(ctx context.Context, tx *sql.Tx, userID string, password string) (bool, error) {
	var passwordHash string
	var status string
	err := tx.QueryRowContext(ctx, `
SELECT password_hash, status
FROM wallet_payment_passwords
WHERE user_id = $1`, strings.TrimSpace(userID)).Scan(&passwordHash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == WalletPaymentPasswordSet && passwordHash == hashPaymentPassword(strings.TrimSpace(password)), nil
}

func ensureAndLockSQLWalletAccount(ctx context.Context, tx *sql.Tx, userID string) (WalletAccount, error) {
	userID = strings.TrimSpace(userID)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO wallet_accounts (subject_type, subject_id, balance_fen, frozen_fen, version, risk_state)
VALUES ('user', $1, 0, 0, 0, 'normal')
ON CONFLICT (subject_type, subject_id) DO NOTHING`, userID); err != nil {
		return WalletAccount{}, err
	}
	return loadSQLWalletAccountForUpdate(ctx, tx, userID)
}

func loadSQLWalletAccountForUpdate(ctx context.Context, tx *sql.Tx, userID string) (WalletAccount, error) {
	var account WalletAccount
	err := tx.QueryRowContext(ctx, `
SELECT subject_id, balance_fen, frozen_fen, version, risk_state
FROM wallet_accounts
WHERE subject_type = 'user' AND subject_id = $1
FOR UPDATE`, strings.TrimSpace(userID)).Scan(
		&account.UserID,
		&account.Balance,
		&account.Frozen,
		&account.Version,
		&account.RiskState,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WalletAccount{}, ErrNotFound
	}
	if err != nil {
		return WalletAccount{}, err
	}
	return account, nil
}

func updateSQLOrderAfterBalancePayment(ctx context.Context, tx *sql.Tx, order Order, actorID string, paidAt time.Time) error {
	_, err := tx.ExecContext(ctx, `
	UPDATE orders
	SET status = $2,
    payment_method = $3,
    paid_at = $4,
    updated_at = $4
WHERE id = $1`,
		order.ID,
		order.Status,
		PaymentBalance,
		paidAt.UTC(),
	)
	if err != nil {
		return err
	}
	return insertSQLOrderEvent(ctx, tx, order.ID, OrderEvent{
		Type:      "order.payment.success",
		ActorID:   actorID,
		Message:   paymentSuccessMessage(&order),
		CreatedAt: paidAt.UTC(),
	})
}

func insertSQLOrderEvent(ctx context.Context, tx *sql.Tx, orderID string, event OrderEvent) error {
	orderID = strings.TrimSpace(orderID)
	event.Type = strings.TrimSpace(event.Type)
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.Message = strings.TrimSpace(event.Message)
	if orderID == "" || event.Type == "" {
		return ErrInvalidArgument
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
	INSERT INTO order_events (order_id, type, actor_id, message, created_at)
	VALUES ($1, $2, $3, $4, $5)`,
		orderID,
		event.Type,
		event.ActorID,
		event.Message,
		event.CreatedAt,
	)
	return err
}

func loadSQLOrderForStateCompensation(ctx context.Context, tx *sql.Tx, orderID string) (Order, error) {
	order, err := loadSQLOrderForBalancePayment(ctx, tx, orderID)
	if err != nil {
		return Order{}, err
	}
	rows, err := tx.QueryContext(ctx, `
	SELECT type, actor_id, message, created_at
	FROM order_events
	WHERE order_id = $1
	ORDER BY created_at, id`, strings.TrimSpace(orderID))
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var event OrderEvent
		if err := rows.Scan(&event.Type, &event.ActorID, &event.Message, &event.CreatedAt); err != nil {
			return Order{}, err
		}
		event.CreatedAt = event.CreatedAt.UTC()
		order.Events = append(order.Events, event)
	}
	if err := rows.Err(); err != nil {
		return Order{}, err
	}
	return order, nil
}

func loadSQLWalletTransactionsForOrder(ctx context.Context, tx *sql.Tx, orderID string) ([]WalletTransaction, error) {
	rows, err := tx.QueryContext(ctx, `
	SELECT id, subject_id, order_id, type, amount_fen, payment_method, idempotency_key, status, created_at
	FROM wallet_transactions
	WHERE subject_type = 'user' AND order_id = $1
	ORDER BY created_at, id`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := []WalletTransaction{}
	for rows.Next() {
		var transaction WalletTransaction
		var nullableOrderID sql.NullString
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&nullableOrderID,
			&transaction.Type,
			&transaction.AmountFen,
			&transaction.PaymentMethod,
			&transaction.IdempotencyKey,
			&transaction.Status,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}
		if nullableOrderID.Valid {
			transaction.OrderID = nullableOrderID.String
		}
		transaction.CreatedAt = transaction.CreatedAt.UTC()
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func loadSQLPaymentTransactionsForOrder(ctx context.Context, tx *sql.Tx, orderID string) ([]PaymentTransaction, error) {
	rows, err := tx.QueryContext(ctx, `
	SELECT id, order_id, user_id, method, amount_fen, status, out_trade_no,
	       transaction_id, idempotency_key, created_at, updated_at
	FROM payment_transactions
	WHERE order_id = $1
	ORDER BY created_at, id`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := []PaymentTransaction{}
	for rows.Next() {
		var transaction PaymentTransaction
		var transactionID sql.NullString
		if err := rows.Scan(
			&transaction.ID,
			&transaction.OrderID,
			&transaction.UserID,
			&transaction.Method,
			&transaction.AmountFen,
			&transaction.Status,
			&transaction.OutTradeNo,
			&transactionID,
			&transaction.IdempotencyKey,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if transactionID.Valid {
			transaction.TransactionID = transactionID.String
		}
		transaction.CreatedAt = transaction.CreatedAt.UTC()
		transaction.UpdatedAt = transaction.UpdatedAt.UTC()
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func loadSQLDispatchEventsForOrder(ctx context.Context, tx *sql.Tx, orderID string) ([]DispatchEvent, error) {
	rows, err := tx.QueryContext(ctx, `
	SELECT id, order_id, station_id, mode, type, rider_id, actor_id, reason,
	       idempotency_key, online_candidate_size,
	       COALESCE(to_json(rejected_rider_ids)::text, '[]'),
	       can_decline_without_penalty, created_at
	FROM dispatch_events
	WHERE order_id = $1
	ORDER BY created_at, id`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []DispatchEvent{}
	for rows.Next() {
		var event DispatchEvent
		var rejectedPayload string
		if err := rows.Scan(
			&event.ID,
			&event.OrderID,
			&event.StationID,
			&event.Mode,
			&event.Type,
			&event.RiderID,
			&event.ActorID,
			&event.Reason,
			&event.IdempotencyKey,
			&event.OnlineCandidateSize,
			&rejectedPayload,
			&event.CanDeclineWithoutPenalty,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(rejectedPayload), &event.RejectedRiderIDs); err != nil {
			return nil, err
		}
		event.RejectedRiderIDs = normalizedStringSlice(event.RejectedRiderIDs)
		event.CreatedAt = event.CreatedAt.UTC()
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStore) compensateOrderStateInSQLTx(ctx context.Context, tx *sql.Tx, req CompensateOrderStateRequest) (*CompensateOrderStateResult, OrderEvent, error) {
	order, err := loadSQLOrderForStateCompensation(ctx, tx, req.OrderID)
	if err != nil {
		return nil, OrderEvent{}, err
	}
	walletTransactions, err := loadSQLWalletTransactionsForOrder(ctx, tx, req.OrderID)
	if err != nil {
		return nil, OrderEvent{}, err
	}
	paymentTransactions, err := loadSQLPaymentTransactionsForOrder(ctx, tx, req.OrderID)
	if err != nil {
		return nil, OrderEvent{}, err
	}
	dispatchEvents, err := loadSQLDispatchEventsForOrder(ctx, tx, req.OrderID)
	if err != nil {
		return nil, OrderEvent{}, err
	}

	temp := NewStore(nil)
	temp.orders = map[string]*Order{order.ID: cloneOrder(&order)}
	temp.walletIdempotency = map[string]*WalletTransaction{}
	for _, transaction := range walletTransactions {
		transactionCopy := transaction
		if strings.TrimSpace(transactionCopy.IdempotencyKey) != "" {
			temp.walletIdempotency[transactionCopy.IdempotencyKey] = cloneWalletTransaction(&transactionCopy)
		}
	}
	temp.paymentTransactions = map[string]*PaymentTransaction{}
	for _, transaction := range paymentTransactions {
		transactionCopy := transaction
		temp.paymentTransactions[transactionCopy.ID] = clonePaymentTransaction(&transactionCopy)
	}
	temp.dispatchEvents = map[string]*DispatchEvent{}
	for _, event := range dispatchEvents {
		eventCopy := event
		temp.dispatchEvents[eventCopy.ID] = cloneDispatchEvent(&eventCopy)
	}
	s.Store.mu.Lock()
	for id, voucher := range s.Store.groupbuyVouchers {
		if voucher != nil && voucher.OrderID == order.ID {
			temp.groupbuyVouchers[id] = cloneGroupbuyVoucher(voucher)
		}
	}
	s.Store.mu.Unlock()

	plan := temp.orderStateCompensationPlanLocked(temp.orders[order.ID], req)
	if !plan.Result.Changed {
		return plan.Result, OrderEvent{}, nil
	}
	nextStatus := strings.TrimSpace(plan.Result.ExpectedStatus)
	nextRiderID := strings.TrimSpace(plan.Result.ExpectedRiderID)
	nextPaymentMethod := strings.TrimSpace(order.PaymentMethod)
	if nextPaymentMethod == "" {
		nextPaymentMethod = strings.TrimSpace(plan.PaymentMethod)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE orders
		SET status = $2,
		    rider_id = $3,
		    payment_method = $4,
		    updated_at = $5
		WHERE id = $1`,
		order.ID,
		nextStatus,
		nullableString(nextRiderID),
		nextPaymentMethod,
		req.Now.UTC(),
	)
	if err != nil {
		return nil, OrderEvent{}, err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return nil, OrderEvent{}, ErrNotFound
	}
	if err := insertSQLOrderEvent(ctx, tx, order.ID, plan.Event); err != nil {
		return nil, OrderEvent{}, err
	}
	updatedOrder := order
	updatedOrder.Status = nextStatus
	updatedOrder.RiderID = nextRiderID
	updatedOrder.PaymentMethod = nextPaymentMethod
	updatedOrder.UpdatedAt = req.Now.UTC()
	updatedOrder.Events = append(updatedOrder.Events, plan.Event)
	plan.Result.Order = cloneOrder(&updatedOrder)
	return plan.Result, plan.Event, nil
}

func (s *PostgresStore) createOrderInSQL(ctx context.Context, req CreateOrderRequest) (Order, OrderEvent, error) {
	userID := strings.TrimSpace(req.UserID)
	orderType := strings.TrimSpace(req.Type)
	if userID == "" || !IsOrderType(orderType) || req.AmountFen <= 0 {
		return Order{}, OrderEvent{}, ErrInvalidArgument
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Order{}, OrderEvent{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := ensureSQLUserPlaceholder(ctx, tx, userID); err != nil {
		return Order{}, OrderEvent{}, err
	}
	orderNumber, err := nextSQLOrderNumber(ctx, tx)
	if err != nil {
		return Order{}, OrderEvent{}, err
	}
	now := time.Now().UTC()
	order := Order{
		ID:        fmt.Sprintf("ord_%d", orderNumber),
		UserID:    userID,
		Type:      orderType,
		Status:    StatusPendingPayment,
		AmountFen: req.AmountFen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	event := OrderEvent{
		Type:      "order.created",
		ActorID:   userID,
		Message:   "订单已创建",
		CreatedAt: now,
	}
	if err := insertSQLCreatedOrder(ctx, tx, order); err != nil {
		return Order{}, OrderEvent{}, err
	}
	if err := insertSQLOrderEvent(ctx, tx, order.ID, event); err != nil {
		return Order{}, OrderEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, OrderEvent{}, err
	}
	order.Events = []OrderEvent{event}
	return order, event, nil
}

type sqlCheckoutShop struct {
	ID              string
	MerchantID      string
	OperationStatus string
	ServiceState    string
	MerchantStatus  string
	DepositStatus   string
}

func (s *PostgresStore) checkoutCartInSQL(ctx context.Context, req CheckoutCartRequest) (Order, CartSummary, OrderEvent, error) {
	userID := strings.TrimSpace(req.UserID)
	shopID := strings.TrimSpace(req.ShopID)
	addressID := strings.TrimSpace(req.AddressID)
	if userID == "" || shopID == "" || addressID == "" {
		return Order{}, CartSummary{}, OrderEvent{}, ErrInvalidArgument
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1)::bigint)`, "checkout:"+userID+":"+shopID); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if err := ensureSQLUserPlaceholder(ctx, tx, userID); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	shop, err := loadSQLCheckoutShopForUpdate(ctx, tx, shopID)
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if !sqlShopCanAcceptOrders(shop) {
		return Order{}, CartSummary{}, OrderEvent{}, ErrInvalidOrderState
	}
	qualificationsReady, err := sqlMerchantQualificationsReady(ctx, tx, shop.MerchantID)
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if !qualificationsReady {
		return Order{}, CartSummary{}, OrderEvent{}, ErrInvalidOrderState
	}
	address, err := loadSQLCheckoutAddressForUpdate(ctx, tx, userID, addressID)
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	summary, err := loadSQLCheckoutCartSummaryForUpdate(ctx, tx, userID, shopID)
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if len(summary.Items) == 0 || summary.PayableFen <= 0 {
		return Order{}, CartSummary{}, OrderEvent{}, ErrInvalidArgument
	}
	orderNumber, err := nextSQLOrderNumber(ctx, tx)
	if err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}

	now := time.Now().UTC()
	orderItems := make([]OrderItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		orderItems = append(orderItems, OrderItem{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			UnitPriceFen: item.UnitPriceFen,
			Quantity:     item.Quantity,
		})
	}
	event := OrderEvent{
		Type:      "order.checkout_created",
		ActorID:   userID,
		Message:   "购物车结算创建订单",
		CreatedAt: now,
	}
	order := Order{
		ID:              fmt.Sprintf("ord_%d", orderNumber),
		UserID:          userID,
		ShopID:          shopID,
		AddressID:       addressID,
		Type:            OrderTypeTakeout,
		Status:          StatusPendingPayment,
		AmountFen:       summary.PayableFen,
		ItemsTotalFen:   summary.ItemsTotalFen,
		DeliveryFeeFen:  summary.DeliveryFeeFen,
		PackagingFeeFen: summary.PackagingFeeFen,
		DiscountFen:     summary.DiscountFen,
		Items:           orderItems,
		Options:         normalizeOrderOptions(req.Options),
		CreatedAt:       now,
		UpdatedAt:       now,
		Events:          []OrderEvent{event},
	}
	if err := insertSQLCheckoutOrder(ctx, tx, order, shop.MerchantID, address); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if err := insertSQLOrderEvent(ctx, tx, order.ID, event); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if _, err := tx.ExecContext(ctx, `
DELETE FROM cart_items
WHERE user_id = $1 AND shop_id = $2`, userID, shopID); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, CartSummary{}, OrderEvent{}, err
	}
	return order, summary, event, nil
}

func loadSQLCheckoutShopForUpdate(ctx context.Context, tx *sql.Tx, shopID string) (sqlCheckoutShop, error) {
	var shop sqlCheckoutShop
	err := tx.QueryRowContext(ctx, `
SELECT s.id, s.merchant_id, s.operation_status, s.service_state,
       m.operation_status, m.deposit_status
FROM shops s
JOIN merchant_accounts m ON m.id = s.merchant_id
WHERE s.id = $1
FOR UPDATE OF s, m`, shopID).Scan(
		&shop.ID,
		&shop.MerchantID,
		&shop.OperationStatus,
		&shop.ServiceState,
		&shop.MerchantStatus,
		&shop.DepositStatus,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return sqlCheckoutShop{}, ErrNotFound
	}
	return shop, err
}

func sqlShopCanAcceptOrders(shop sqlCheckoutShop) bool {
	if strings.TrimSpace(shop.ID) == "" || strings.TrimSpace(shop.MerchantID) == "" {
		return false
	}
	if strings.TrimSpace(shop.OperationStatus) != ShopStatusActive {
		return false
	}
	if strings.TrimSpace(shop.MerchantStatus) != ShopStatusActive {
		return false
	}
	if strings.TrimSpace(shop.DepositStatus) != DepositStatusPaid {
		return false
	}
	switch strings.TrimSpace(shop.ServiceState) {
	case "", ShopServiceOpen, ShopServiceBusy:
		return true
	default:
		return false
	}
}

func sqlMerchantQualificationsReady(ctx context.Context, tx *sql.Tx, merchantID string) (bool, error) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" {
		return false, ErrInvalidArgument
	}
	rows, err := tx.QueryContext(ctx, `
SELECT type
FROM merchant_qualifications
WHERE merchant_id = $1
  AND type IN ($2, $3)
  AND status = 'approved'
  AND expires_at > now()
FOR SHARE`, merchantID, QualificationBusinessLicense, QualificationHealthCertificate)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	approved := map[string]bool{}
	for rows.Next() {
		var qualificationType string
		if err := rows.Scan(&qualificationType); err != nil {
			return false, err
		}
		approved[strings.TrimSpace(qualificationType)] = true
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return approved[QualificationBusinessLicense] && approved[QualificationHealthCertificate], nil
}

func loadSQLCheckoutAddressForUpdate(ctx context.Context, tx *sql.Tx, userID string, addressID string) (UserAddress, error) {
	var address UserAddress
	var latitude float64
	var longitude float64
	err := tx.QueryRowContext(ctx, `
SELECT id, user_id, contact_name, contact_phone_hash, city, detail,
       latitude, longitude, tag, is_default
FROM user_addresses
WHERE user_id = $1 AND id = $2
FOR UPDATE`, userID, addressID).Scan(
		&address.ID,
		&address.UserID,
		&address.ContactName,
		&address.ContactPhone,
		&address.City,
		&address.Detail,
		&latitude,
		&longitude,
		&address.Tag,
		&address.IsDefault,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return UserAddress{}, ErrInvalidArgument
	}
	if err != nil {
		return UserAddress{}, err
	}
	address.Latitude = &latitude
	address.Longitude = &longitude
	if !UserAddressReady(address) {
		return UserAddress{}, ErrInvalidArgument
	}
	return address, nil
}

func loadSQLCheckoutCartSummaryForUpdate(ctx context.Context, tx *sql.Tx, userID string, shopID string) (CartSummary, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT ci.product_id, ci.product_name_snapshot, ci.unit_price_fen,
       ci.quantity, ci.selected, mp.shop_id, mp.name, mp.stock_count, mp.status
FROM cart_items ci
JOIN merchant_products mp ON mp.id = ci.product_id
WHERE ci.user_id = $1 AND ci.shop_id = $2
ORDER BY ci.product_id
FOR UPDATE OF ci, mp`, userID, shopID)
	if err != nil {
		return CartSummary{}, err
	}
	defer rows.Close()

	items := []CartItem{}
	var itemsTotal int64
	for rows.Next() {
		var item CartItem
		var productShopID string
		var currentProductName string
		var stockCount int
		var productStatus string
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.UnitPriceFen,
			&item.Quantity,
			&item.Selected,
			&productShopID,
			&currentProductName,
			&stockCount,
			&productStatus,
		); err != nil {
			return CartSummary{}, err
		}
		if !item.Selected || item.Quantity <= 0 {
			continue
		}
		if productShopID != shopID || productStatus != ProductStatusActive {
			return CartSummary{}, ErrNotFound
		}
		if item.Quantity > stockCount || item.UnitPriceFen <= 0 {
			return CartSummary{}, ErrInvalidArgument
		}
		item.UserID = userID
		item.ShopID = shopID
		item.ProductName = strings.TrimSpace(item.ProductName)
		if item.ProductName == "" {
			item.ProductName = strings.TrimSpace(currentProductName)
		}
		items = append(items, item)
		itemsTotal += item.UnitPriceFen * int64(item.Quantity)
	}
	if err := rows.Err(); err != nil {
		return CartSummary{}, err
	}
	deliveryFee := int64(300)
	packagingFee := int64(100 * len(items))
	if len(items) == 0 {
		deliveryFee = 0
		packagingFee = 0
	}
	return CartSummary{
		UserID:          userID,
		ShopID:          shopID,
		Items:           items,
		ItemsTotalFen:   itemsTotal,
		DeliveryFeeFen:  deliveryFee,
		PackagingFeeFen: packagingFee,
		DiscountFen:     0,
		PayableFen:      OrderPayableFen(itemsTotal, deliveryFee, packagingFee, 0),
	}, nil
}

func insertSQLCheckoutOrder(ctx context.Context, tx *sql.Tx, order Order, merchantID string, address UserAddress) error {
	optionsPayload, err := json.Marshal(order.Options)
	if err != nil {
		return err
	}
	pricingPayload, err := json.Marshal(map[string]any{
		"items_total_fen":   order.ItemsTotalFen,
		"delivery_fee_fen":  order.DeliveryFeeFen,
		"packaging_fee_fen": order.PackagingFeeFen,
		"discount_fen":      order.DiscountFen,
		"amount_fen":        order.AmountFen,
	})
	if err != nil {
		return err
	}
	addressSnapshot, err := json.Marshal(map[string]any{
		"id":                 address.ID,
		"user_id":            address.UserID,
		"contact_name":       address.ContactName,
		"contact_phone_hash": address.ContactPhone,
		"city":               address.City,
		"detail":             address.Detail,
		"latitude":           address.Latitude,
		"longitude":          address.Longitude,
		"tag":                address.Tag,
	})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO orders (
  id, user_id, merchant_id, shop_id, address_id, type, status,
  payment_method, items_total_fen, delivery_fee_fen, packaging_fee_fen,
  discount_fen, amount_fen, address_snapshot, options, pricing_snapshot,
  dispatch_mode, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb, $15::jsonb, $16::jsonb, $17, $18, $19)`,
		order.ID,
		order.UserID,
		nullableString(merchantID),
		nullableString(order.ShopID),
		nullableString(order.AddressID),
		order.Type,
		order.Status,
		strings.TrimSpace(order.PaymentMethod),
		nonNegativeInt64(order.ItemsTotalFen),
		nonNegativeInt64(order.DeliveryFeeFen),
		nonNegativeInt64(order.PackagingFeeFen),
		nonNegativeInt64(order.DiscountFen),
		nonNegativeInt64(order.AmountFen),
		string(addressSnapshot),
		string(optionsPayload),
		string(pricingPayload),
		DispatchModeGrabHall,
		order.CreatedAt.UTC(),
		order.UpdatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	for _, item := range order.Items {
		if strings.TrimSpace(item.ProductID) == "" || item.Quantity <= 0 {
			continue
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO order_items (order_id, product_id, product_name_snapshot, unit_price_fen, quantity)
VALUES ($1, $2, $3, $4, $5)`,
			order.ID,
			strings.TrimSpace(item.ProductID),
			strings.TrimSpace(item.ProductName),
			nonNegativeInt64(item.UnitPriceFen),
			item.Quantity,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureSQLUserPlaceholder(ctx context.Context, tx *sql.Tx, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrInvalidArgument
	}
	_, err := tx.ExecContext(ctx, `
	INSERT INTO app_users (id, status, created_at, updated_at)
	VALUES ($1, 'active', now(), now())
	ON CONFLICT (id) DO NOTHING`, userID)
	return err
}

func nextSQLOrderNumber(ctx context.Context, tx *sql.Tx) (uint64, error) {
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO platform_sequences (name, next_value)
	SELECT 'orders', COALESCE(MAX((substring(id FROM '^ord_([0-9]+)$'))::bigint), 0) + 1
	FROM orders
	ON CONFLICT (name) DO NOTHING`); err != nil {
		return 0, err
	}

	var nextValue int64
	if err := tx.QueryRowContext(ctx, `
	SELECT next_value
	FROM platform_sequences
	WHERE name = 'orders'
	FOR UPDATE`).Scan(&nextValue); err != nil {
		return 0, err
	}
	var maxOrderNumber int64
	if err := tx.QueryRowContext(ctx, `
	SELECT COALESCE(MAX((substring(id FROM '^ord_([0-9]+)$'))::bigint), 0)
	FROM orders`).Scan(&maxOrderNumber); err != nil {
		return 0, err
	}
	if nextValue <= maxOrderNumber {
		nextValue = maxOrderNumber + 1
	}
	if nextValue < 1 {
		nextValue = 1
	}
	if _, err := tx.ExecContext(ctx, `
	UPDATE platform_sequences
	SET next_value = $2,
	    updated_at = now()
	WHERE name = $1`, "orders", nextValue+1); err != nil {
		return 0, err
	}
	return uint64(nextValue), nil
}

func nextSQLAuditLogNumber(ctx context.Context, tx *sql.Tx) (uint64, error) {
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO platform_sequences (name, next_value)
	SELECT 'audit_logs', COALESCE(MAX((substring(id FROM '^aud_([0-9]+)$'))::bigint), 0) + 1
	FROM audit_logs
	ON CONFLICT (name) DO NOTHING`); err != nil {
		return 0, err
	}

	var nextValue int64
	if err := tx.QueryRowContext(ctx, `
	SELECT next_value
	FROM platform_sequences
	WHERE name = 'audit_logs'
	FOR UPDATE`).Scan(&nextValue); err != nil {
		return 0, err
	}
	var maxAuditLogNumber int64
	if err := tx.QueryRowContext(ctx, `
	SELECT COALESCE(MAX((substring(id FROM '^aud_([0-9]+)$'))::bigint), 0)
	FROM audit_logs`).Scan(&maxAuditLogNumber); err != nil {
		return 0, err
	}
	if nextValue <= maxAuditLogNumber {
		nextValue = maxAuditLogNumber + 1
	}
	if nextValue < 1 {
		nextValue = 1
	}
	if _, err := tx.ExecContext(ctx, `
	UPDATE platform_sequences
	SET next_value = $2,
	    updated_at = now()
	WHERE name = $1`, "audit_logs", nextValue+1); err != nil {
		return 0, err
	}
	return uint64(nextValue), nil
}

func insertSQLCreatedOrder(ctx context.Context, tx *sql.Tx, order Order) error {
	optionsPayload, err := json.Marshal(order.Options)
	if err != nil {
		return err
	}
	pricingPayload, err := json.Marshal(map[string]any{
		"items_total_fen":   order.ItemsTotalFen,
		"delivery_fee_fen":  order.DeliveryFeeFen,
		"packaging_fee_fen": order.PackagingFeeFen,
		"discount_fen":      order.DiscountFen,
		"amount_fen":        order.AmountFen,
	})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
	INSERT INTO orders (
	  id, user_id, type, status, payment_method, items_total_fen,
	  delivery_fee_fen, packaging_fee_fen, discount_fen, amount_fen,
	  options, pricing_snapshot, dispatch_mode, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12::jsonb, $13, $14, $15)`,
		strings.TrimSpace(order.ID),
		strings.TrimSpace(order.UserID),
		strings.TrimSpace(order.Type),
		strings.TrimSpace(order.Status),
		strings.TrimSpace(order.PaymentMethod),
		nonNegativeInt64(order.ItemsTotalFen),
		nonNegativeInt64(order.DeliveryFeeFen),
		nonNegativeInt64(order.PackagingFeeFen),
		nonNegativeInt64(order.DiscountFen),
		nonNegativeInt64(order.AmountFen),
		string(optionsPayload),
		string(pricingPayload),
		DispatchModeGrabHall,
		order.CreatedAt.UTC(),
		order.UpdatedAt.UTC(),
	)
	return err
}

func (s *PostgresStore) transitionMerchantOrderInSQL(ctx context.Context, orderID string, merchantID string, expectedStatus string, nextStatus string, eventType string, message string, requireAcceptingShop bool) (Order, OrderEvent, error) {
	orderID = strings.TrimSpace(orderID)
	merchantID = strings.TrimSpace(merchantID)
	expectedStatus = strings.TrimSpace(expectedStatus)
	nextStatus = strings.TrimSpace(nextStatus)
	eventType = strings.TrimSpace(eventType)
	message = strings.TrimSpace(message)
	if orderID == "" || merchantID == "" || expectedStatus == "" || nextStatus == "" || eventType == "" {
		return Order{}, OrderEvent{}, ErrInvalidArgument
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Order{}, OrderEvent{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	order, ownerMerchantID, shopStatus, merchantDepositStatus, err := loadSQLMerchantOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return Order{}, OrderEvent{}, err
	}
	if ownerMerchantID != merchantID || order.Status != expectedStatus {
		return Order{}, OrderEvent{}, ErrInvalidOrderState
	}
	if requireAcceptingShop && (shopStatus != ShopStatusActive || merchantDepositStatus != DepositStatusPaid) {
		return Order{}, OrderEvent{}, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	event := OrderEvent{
		Type:      eventType,
		ActorID:   merchantID,
		Message:   message,
		CreatedAt: now,
	}
	if err := updateSQLMerchantOrderStatus(ctx, tx, order.ID, nextStatus, now); err != nil {
		return Order{}, OrderEvent{}, err
	}
	if err := insertSQLOrderEvent(ctx, tx, order.ID, event); err != nil {
		return Order{}, OrderEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, OrderEvent{}, err
	}
	order.Status = nextStatus
	order.UpdatedAt = now
	order.Events = append(order.Events, event)
	return order, event, nil
}

func loadSQLMerchantOrderForUpdate(ctx context.Context, tx *sql.Tx, orderID string) (Order, string, string, string, error) {
	var order Order
	var shopID sql.NullString
	var riderID sql.NullString
	var addressID sql.NullString
	var ownerMerchantID string
	var shopStatus string
	var merchantDepositStatus string
	err := tx.QueryRowContext(ctx, `
	SELECT orders.id, orders.user_id, orders.shop_id, orders.rider_id, orders.address_id,
	       orders.type, orders.status, orders.payment_method, orders.amount_fen,
	       orders.created_at, orders.updated_at,
	       COALESCE(shops.merchant_id, ''),
	       COALESCE(shops.operation_status, ''),
	       COALESCE(merchant_accounts.deposit_status, '')
	FROM orders
	LEFT JOIN shops ON shops.id = orders.shop_id
	LEFT JOIN merchant_accounts ON merchant_accounts.id = shops.merchant_id
	WHERE orders.id = $1
	FOR UPDATE OF orders`, strings.TrimSpace(orderID)).Scan(
		&order.ID,
		&order.UserID,
		&shopID,
		&riderID,
		&addressID,
		&order.Type,
		&order.Status,
		&order.PaymentMethod,
		&order.AmountFen,
		&order.CreatedAt,
		&order.UpdatedAt,
		&ownerMerchantID,
		&shopStatus,
		&merchantDepositStatus,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, "", "", "", ErrNotFound
	}
	if err != nil {
		return Order{}, "", "", "", err
	}
	if shopID.Valid {
		order.ShopID = shopID.String
	}
	if riderID.Valid {
		order.RiderID = riderID.String
	}
	if addressID.Valid {
		order.AddressID = addressID.String
	}
	order.CreatedAt = order.CreatedAt.UTC()
	order.UpdatedAt = order.UpdatedAt.UTC()
	return order, strings.TrimSpace(ownerMerchantID), strings.TrimSpace(shopStatus), strings.TrimSpace(merchantDepositStatus), nil
}

func updateSQLMerchantOrderStatus(ctx context.Context, tx *sql.Tx, orderID string, status string, updatedAt time.Time) error {
	result, err := tx.ExecContext(ctx, `
	UPDATE orders
	SET status = $2,
	    updated_at = $3
	WHERE id = $1`,
		strings.TrimSpace(orderID),
		strings.TrimSpace(status),
		updatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrNotFound
	}
	return nil
}

func updateSQLWalletAccountBalance(ctx context.Context, tx *sql.Tx, account WalletAccount) error {
	_, err := tx.ExecContext(ctx, `
UPDATE wallet_accounts
SET balance_fen = $3,
    frozen_fen = $4,
    version = $5,
    risk_state = $6,
    updated_at = now()
WHERE subject_type = $1 AND subject_id = $2`,
		"user",
		account.UserID,
		nonNegativeInt64(account.Balance),
		nonNegativeInt64(account.Frozen),
		account.Version,
		strings.TrimSpace(account.RiskState),
	)
	return err
}

func insertSQLWalletPaymentTransaction(ctx context.Context, tx *sql.Tx, transaction WalletTransaction, balanceAfterFen int64) error {
	result, err := tx.ExecContext(ctx, `
INSERT INTO wallet_transactions (
  id, subject_type, subject_id, order_id, type, amount_fen,
  payment_method, idempotency_key, status, balance_after_fen, created_at
)
VALUES ($1, 'user', $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (idempotency_key) DO NOTHING`,
		transaction.ID,
		transaction.UserID,
		nullableString(transaction.OrderID),
		transaction.Type,
		transaction.AmountFen,
		transaction.PaymentMethod,
		transaction.IdempotencyKey,
		transaction.Status,
		balanceAfterFen,
		transaction.CreatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrInvalidOrderState
	}
	return err
}

func (s *PostgresStore) refundOrderInSQL(ctx context.Context, req RefundOrderRequest) (RefundTransaction, WalletAccount, string, OrderEvent, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	refund, account, orderID, event, err := refundOrderInSQLTx(ctx, tx, req)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	return refund, account, orderID, event, nil
}

func refundOrderInSQLTx(ctx context.Context, tx *sql.Tx, req RefundOrderRequest) (RefundTransaction, WalletAccount, string, OrderEvent, error) {
	orderID := strings.TrimSpace(req.OrderID)
	userID := strings.TrimSpace(req.UserID)
	reason := strings.TrimSpace(req.Reason)
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if orderID == "" || reason == "" || idempotencyKey == "" {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidArgument
	}

	if err := lockSQLWalletIdempotencyKey(ctx, tx, idempotencyKey); err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	if existing, found, err := loadSQLRefundTransactionByIdempotency(ctx, tx, idempotencyKey); err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	} else if found {
		if !refundTransactionMatchesRequest(existing, req) {
			return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidOrderState
		}
		var account WalletAccount
		if existing.Destination == RefundDestinationBalance {
			account, err = loadSQLWalletAccountForUpdate(ctx, tx, existing.UserID)
			if err != nil {
				return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
			}
		}
		return existing, account, existing.OrderID, refundOrderEvent(existing, strings.TrimSpace(req.ActorID)), nil
	}

	settings, err := loadSQLRefundSettingsForUpdate(ctx, tx)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	order, err := loadSQLOrderForBalancePayment(ctx, tx, orderID)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	if userID != "" && order.UserID != userID {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidOrderState
	}
	switch order.Status {
	case StatusPendingPayment, StatusCancelled, StatusRefundPending, StatusRefunded:
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidOrderState
	}
	refundedBefore, err := refundedAmountForSQLOrder(ctx, tx, order.ID)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}
	remainingFen := order.AmountFen - refundedBefore
	amountFen := req.AmountFen
	if amountFen <= 0 {
		amountFen = remainingFen
	}
	if amountFen <= 0 {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidArgument
	}
	if remainingFen <= 0 || amountFen > remainingFen {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, ErrInvalidArgument
	}

	destination := RefundDestinationForStrategy(settings.DefaultStrategy, strings.TrimSpace(req.Destination))
	now := time.Now().UTC()
	refund := RefundTransaction{
		ID:             "rfd_" + shortHash(idempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      amountFen,
		Destination:    destination,
		Status:         RefundStatusPendingOriginal,
		Reason:         reason,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
	}
	event := refundOrderEvent(refund, strings.TrimSpace(req.ActorID))
	if destination == RefundDestinationBalance {
		refund.Status = RefundStatusSuccess
		refund.Destination = RefundDestinationBalance
		event = refundOrderEvent(refund, strings.TrimSpace(req.ActorID))
	}
	if err := insertSQLRefundTransaction(ctx, tx, refund); err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}

	var account WalletAccount
	if refund.Destination == RefundDestinationBalance {
		account, err = ensureAndLockSQLWalletAccount(ctx, tx, refund.UserID)
		if err != nil {
			return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
		}
		account.Balance += refund.AmountFen
		account.Version++
		if err := updateSQLWalletAccountBalance(ctx, tx, account); err != nil {
			return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
		}
		walletTransaction := WalletTransaction{
			ID:             postgresRefundWalletTransactionID(idempotencyKey),
			UserID:         refund.UserID,
			OrderID:        refund.OrderID,
			Type:           "refund",
			AmountFen:      refund.AmountFen,
			PaymentMethod:  RefundDestinationBalance,
			IdempotencyKey: idempotencyKey,
			Status:         "success",
			CreatedAt:      now,
		}
		if err := insertSQLWalletRefundTransaction(ctx, tx, walletTransaction, account.Balance); err != nil {
			return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
		}
		nextStatus := refundOrderStatusAfter(order.Status, refundedBefore+refund.AmountFen, order.AmountFen, refund.Destination)
		if err := updateSQLOrderAfterRefund(ctx, tx, order.ID, nextStatus, event); err != nil {
			return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
		}
	} else if err := updateSQLOrderAfterRefund(ctx, tx, order.ID, refundOrderStatusAfter(order.Status, refundedBefore+refund.AmountFen, order.AmountFen, refund.Destination), event); err != nil {
		return RefundTransaction{}, WalletAccount{}, "", OrderEvent{}, err
	}

	return refund, account, order.ID, event, nil
}

type sqlAfterSalesReviewResult struct {
	RequestID string
	RefundID  string
	OrderID   string
	Status    string
	Events    []OrderEvent
}

func (s *PostgresStore) reviewAfterSalesInSQL(ctx context.Context, req ReviewAfterSalesRequest) (sqlAfterSalesReviewResult, RefundTransaction, WalletAccount, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result, refund, account, err := reviewAfterSalesInSQLTx(ctx, tx, req)
	if err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	return result, refund, account, nil
}

func reviewAfterSalesInSQLTx(ctx context.Context, tx *sql.Tx, req ReviewAfterSalesRequest) (sqlAfterSalesReviewResult, RefundTransaction, WalletAccount, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Decision = strings.TrimSpace(req.Decision)
	req.Reason = strings.TrimSpace(req.Reason)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.RefundIdempotencyKey = strings.TrimSpace(req.RefundIdempotencyKey)
	if req.RequestID == "" || req.Decision == "" || req.ActorID == "" || req.ActorRole == "" {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, ErrInvalidArgument
	}

	request, order, ownerMerchantID, err := loadSQLAfterSalesForUpdate(ctx, tx, req.RequestID)
	if err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	if !canReviewSQLAfterSales(request, ownerMerchantID, req.ActorID, req.ActorRole, req.Decision) {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, ErrInvalidOrderState
	}

	now := time.Now().UTC()
	result := sqlAfterSalesReviewResult{RequestID: request.ID, OrderID: order.ID}
	var refund RefundTransaction
	var account WalletAccount
	var nextStatus string
	var reviewReason string
	var afterSalesEvent OrderEvent
	var afterSalesAuditEvent AfterSalesEvent

	switch req.Decision {
	case AfterSalesDecisionReject:
		if req.Reason == "" {
			return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, ErrInvalidArgument
		}
		nextStatus = AfterSalesRejected
		reviewReason = req.Reason
		afterSalesEvent = afterSalesOrderEvent("order.after_sales.rejected", req.ActorID, now)
		afterSalesAuditEvent = sqlAfterSalesEvent(request, AfterSalesActionReviewRejected, req.ActorID, req.ActorRole, reviewReason, true, nil, now)
	case AfterSalesDecisionEscalate:
		nextStatus = AfterSalesAdminReview
		reviewReason = req.Reason
		afterSalesEvent = afterSalesOrderEvent("order.after_sales.escalated", req.ActorID, now)
		auditMessage := reviewReason
		if auditMessage == "" {
			auditMessage = "售后申请已转平台审核"
		}
		afterSalesAuditEvent = sqlAfterSalesEvent(request, AfterSalesActionEscalated, req.ActorID, req.ActorRole, auditMessage, true, nil, now)
	case AfterSalesDecisionApprove:
		idempotencyKey := req.RefundIdempotencyKey
		if idempotencyKey == "" {
			idempotencyKey = "after_sales:" + request.ID
		}
		if err := lockSQLWalletIdempotencyKey(ctx, tx, idempotencyKey); err != nil {
			return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
		}
		reviewReason = req.Reason
		if reviewReason == "" {
			reviewReason = request.Reason
		}
		refund, account, err = createSQLRefundForAfterSalesReview(ctx, tx, order, request, RefundOrderRequest{
			OrderID:        request.OrderID,
			UserID:         request.UserID,
			AmountFen:      request.RequestedAmountFen,
			Destination:    req.RefundDestination,
			Reason:         reviewReason,
			IdempotencyKey: idempotencyKey,
			ActorID:        req.ActorID,
			ActorRole:      req.ActorRole,
		}, now)
		if err != nil {
			return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
		}
		result.RefundID = refund.ID
		nextStatus = AfterSalesApproved
		if refund.Status == RefundStatusSuccess {
			nextStatus = AfterSalesRefunded
		}
		refundEvent := refundOrderEvent(refund, req.ActorID)
		result.Events = append(result.Events, refundEvent)
		afterSalesEvent = afterSalesOrderEvent("order.after_sales.approved", req.ActorID, now)
		afterSalesAuditEvent = sqlAfterSalesEvent(request, AfterSalesActionReviewApproved, req.ActorID, req.ActorRole, reviewReason, true, nil, now)
	default:
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, ErrInvalidArgument
	}

	if err := updateSQLAfterSalesReview(ctx, tx, request.ID, nextStatus, reviewReason, req.ActorID, req.ActorRole, result.RefundID, now); err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	if err := upsertSQLAfterSalesEvent(ctx, tx, afterSalesAuditEvent); err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	if err := insertSQLOrderEvent(ctx, tx, order.ID, afterSalesEvent); err != nil {
		return sqlAfterSalesReviewResult{}, RefundTransaction{}, WalletAccount{}, err
	}
	result.Status = nextStatus
	result.Events = append(result.Events, afterSalesEvent)
	return result, refund, account, nil
}

func loadSQLAfterSalesForUpdate(ctx context.Context, tx *sql.Tx, requestID string) (AfterSalesRequest, Order, string, error) {
	var request AfterSalesRequest
	var order Order
	var evidencePayload string
	var refundID sql.NullString
	var reviewedAt sql.NullTime
	var shopID sql.NullString
	var riderID sql.NullString
	var addressID sql.NullString
	var ownerMerchantID string
	err := tx.QueryRowContext(ctx, `
SELECT after_sales.id, after_sales.order_id, after_sales.user_id, after_sales.type,
       after_sales.reason, after_sales.requested_amount_fen,
       COALESCE(after_sales.evidence_urls::text, '[]'), after_sales.status,
       after_sales.review_reason, after_sales.reviewer_id, after_sales.reviewer_role,
       after_sales.refund_id, after_sales.created_at, after_sales.updated_at,
       after_sales.reviewed_at,
       orders.id, orders.user_id, orders.shop_id, orders.rider_id, orders.address_id,
       orders.type, orders.status, orders.payment_method, orders.amount_fen,
       orders.created_at, orders.updated_at,
       COALESCE(shops.merchant_id, '')
FROM order_after_sales AS after_sales
JOIN orders ON orders.id = after_sales.order_id
LEFT JOIN shops ON shops.id = orders.shop_id
WHERE after_sales.id = $1
FOR UPDATE OF after_sales, orders`, strings.TrimSpace(requestID)).Scan(
		&request.ID,
		&request.OrderID,
		&request.UserID,
		&request.Type,
		&request.Reason,
		&request.RequestedAmountFen,
		&evidencePayload,
		&request.Status,
		&request.ReviewReason,
		&request.ReviewerID,
		&request.ReviewerRole,
		&refundID,
		&request.CreatedAt,
		&request.UpdatedAt,
		&reviewedAt,
		&order.ID,
		&order.UserID,
		&shopID,
		&riderID,
		&addressID,
		&order.Type,
		&order.Status,
		&order.PaymentMethod,
		&order.AmountFen,
		&order.CreatedAt,
		&order.UpdatedAt,
		&ownerMerchantID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AfterSalesRequest{}, Order{}, "", ErrNotFound
	}
	if err != nil {
		return AfterSalesRequest{}, Order{}, "", err
	}
	if err := json.Unmarshal([]byte(evidencePayload), &request.EvidenceURLs); err != nil {
		return AfterSalesRequest{}, Order{}, "", err
	}
	request.EvidenceURLs = sanitizedStringSlice(request.EvidenceURLs)
	if refundID.Valid {
		request.RefundID = refundID.String
	}
	if reviewedAt.Valid {
		request.ReviewedAt = reviewedAt.Time.UTC()
	}
	request.CreatedAt = request.CreatedAt.UTC()
	request.UpdatedAt = request.UpdatedAt.UTC()
	if shopID.Valid {
		order.ShopID = shopID.String
	}
	if riderID.Valid {
		order.RiderID = riderID.String
	}
	if addressID.Valid {
		order.AddressID = addressID.String
	}
	order.CreatedAt = order.CreatedAt.UTC()
	order.UpdatedAt = order.UpdatedAt.UTC()
	return request, order, strings.TrimSpace(ownerMerchantID), nil
}

func canReviewSQLAfterSales(request AfterSalesRequest, ownerMerchantID string, actorID string, actorRole string, decision string) bool {
	actorID = strings.TrimSpace(actorID)
	actorRole = strings.TrimSpace(actorRole)
	decision = strings.TrimSpace(decision)
	if request.ID == "" || actorID == "" {
		return false
	}
	switch request.Status {
	case AfterSalesPendingMerchant:
	case AfterSalesAdminReview:
		if actorRole != "admin" {
			return false
		}
	default:
		return false
	}
	switch actorRole {
	case "admin":
		return true
	case "merchant":
		return request.Status == AfterSalesPendingMerchant && strings.TrimSpace(ownerMerchantID) == actorID
	default:
		return false
	}
}

func createSQLRefundForAfterSalesReview(ctx context.Context, tx *sql.Tx, order Order, request AfterSalesRequest, req RefundOrderRequest, now time.Time) (RefundTransaction, WalletAccount, error) {
	if existing, found, err := loadSQLRefundTransactionByIdempotency(ctx, tx, req.IdempotencyKey); err != nil {
		return RefundTransaction{}, WalletAccount{}, err
	} else if found {
		if !refundTransactionMatchesRequest(existing, req) {
			return RefundTransaction{}, WalletAccount{}, ErrInvalidOrderState
		}
		var account WalletAccount
		if existing.Destination == RefundDestinationBalance {
			account, err = loadSQLWalletAccountForUpdate(ctx, tx, existing.UserID)
			if err != nil {
				return RefundTransaction{}, WalletAccount{}, err
			}
		}
		return existing, account, nil
	}

	switch order.Status {
	case StatusPendingPayment, StatusCancelled, StatusRefundPending, StatusRefunded:
		return RefundTransaction{}, WalletAccount{}, ErrInvalidOrderState
	}
	if request.UserID != order.UserID || request.OrderID != order.ID || request.RequestedAmountFen <= 0 {
		return RefundTransaction{}, WalletAccount{}, ErrInvalidArgument
	}
	refundedBefore, err := refundedAmountForSQLOrder(ctx, tx, order.ID)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, err
	}
	remainingFen := order.AmountFen - refundedBefore
	if remainingFen <= 0 || request.RequestedAmountFen > remainingFen {
		return RefundTransaction{}, WalletAccount{}, ErrInvalidArgument
	}
	settings, err := loadSQLRefundSettingsForUpdate(ctx, tx)
	if err != nil {
		return RefundTransaction{}, WalletAccount{}, err
	}
	destination := RefundDestinationForStrategy(settings.DefaultStrategy, strings.TrimSpace(req.Destination))
	refund := RefundTransaction{
		ID:             "rfd_" + shortHash(req.IdempotencyKey),
		OrderID:        order.ID,
		UserID:         order.UserID,
		AmountFen:      request.RequestedAmountFen,
		Destination:    destination,
		Status:         RefundStatusPendingOriginal,
		Reason:         strings.TrimSpace(req.Reason),
		IdempotencyKey: strings.TrimSpace(req.IdempotencyKey),
		CreatedAt:      now.UTC(),
	}
	if refund.Reason == "" {
		refund.Reason = request.Reason
	}
	if refund.Destination == RefundDestinationBalance {
		refund.Status = RefundStatusSuccess
	}
	if err := insertSQLRefundTransaction(ctx, tx, refund); err != nil {
		return RefundTransaction{}, WalletAccount{}, err
	}

	event := refundOrderEvent(refund, strings.TrimSpace(req.ActorID))
	var account WalletAccount
	if refund.Destination == RefundDestinationBalance {
		account, err = ensureAndLockSQLWalletAccount(ctx, tx, refund.UserID)
		if err != nil {
			return RefundTransaction{}, WalletAccount{}, err
		}
		account.Balance += refund.AmountFen
		account.Version++
		if err := updateSQLWalletAccountBalance(ctx, tx, account); err != nil {
			return RefundTransaction{}, WalletAccount{}, err
		}
		walletTransaction := WalletTransaction{
			ID:             postgresRefundWalletTransactionID(refund.IdempotencyKey),
			UserID:         refund.UserID,
			OrderID:        refund.OrderID,
			Type:           "refund",
			AmountFen:      refund.AmountFen,
			PaymentMethod:  RefundDestinationBalance,
			IdempotencyKey: refund.IdempotencyKey,
			Status:         "success",
			CreatedAt:      now.UTC(),
		}
		if err := insertSQLWalletRefundTransaction(ctx, tx, walletTransaction, account.Balance); err != nil {
			return RefundTransaction{}, WalletAccount{}, err
		}
		nextStatus := refundOrderStatusAfter(order.Status, refundedBefore+refund.AmountFen, order.AmountFen, refund.Destination)
		if err := updateSQLOrderAfterRefund(ctx, tx, order.ID, nextStatus, event); err != nil {
			return RefundTransaction{}, WalletAccount{}, err
		}
	} else if err := updateSQLOrderAfterRefund(ctx, tx, order.ID, refundOrderStatusAfter(order.Status, refundedBefore+refund.AmountFen, order.AmountFen, refund.Destination), event); err != nil {
		return RefundTransaction{}, WalletAccount{}, err
	}
	return refund, account, nil
}

func updateSQLAfterSalesReview(ctx context.Context, tx *sql.Tx, requestID string, status string, reason string, reviewerID string, reviewerRole string, refundID string, reviewedAt time.Time) error {
	result, err := tx.ExecContext(ctx, `
UPDATE order_after_sales
SET status = $2,
    review_reason = $3,
    reviewer_id = $4,
    reviewer_role = $5,
    refund_id = $6,
    updated_at = $7,
    reviewed_at = $7
WHERE id = $1`,
		strings.TrimSpace(requestID),
		strings.TrimSpace(status),
		strings.TrimSpace(reason),
		strings.TrimSpace(reviewerID),
		strings.TrimSpace(reviewerRole),
		nullableString(refundID),
		reviewedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrNotFound
	}
	return nil
}

func afterSalesOrderEvent(eventType string, actorID string, createdAt time.Time) OrderEvent {
	message := "售后申请已通过"
	switch strings.TrimSpace(eventType) {
	case "order.after_sales.rejected":
		message = "售后申请已驳回"
	case "order.after_sales.escalated":
		message = "售后申请已转平台审核"
	}
	return OrderEvent{
		Type:      strings.TrimSpace(eventType),
		ActorID:   strings.TrimSpace(actorID),
		Message:   message,
		CreatedAt: createdAt.UTC(),
	}
}

func sqlAfterSalesEvent(request AfterSalesRequest, action string, actorID string, actorRole string, message string, visibleToUser bool, attachments []string, createdAt time.Time) AfterSalesEvent {
	action = normalizeAfterSalesAction(action)
	message = strings.TrimSpace(message)
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	} else {
		createdAt = createdAt.UTC()
	}
	return AfterSalesEvent{
		ID:            "asev_" + shortHash(fmt.Sprintf("%s:%s:%s:%d", request.ID, action, strings.TrimSpace(actorID), createdAt.UnixNano())),
		RequestID:     request.ID,
		OrderID:       request.OrderID,
		ActorID:       strings.TrimSpace(actorID),
		ActorRole:     strings.TrimSpace(actorRole),
		Action:        action,
		Message:       message,
		Attachments:   sanitizedStringSlice(attachments),
		VisibleToUser: visibleToUser,
		CreatedAt:     createdAt,
	}
}

func loadSQLRefundSettingsForUpdate(ctx context.Context, tx *sql.Tx) (RefundSettings, error) {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO refund_settings (id, default_strategy, updated_at)
VALUES ('default', $1, now())
ON CONFLICT (id) DO NOTHING`, RefundStrategyBalanceFirst); err != nil {
		return RefundSettings{}, err
	}
	var settings RefundSettings
	err := tx.QueryRowContext(ctx, `
SELECT default_strategy
FROM refund_settings
WHERE id = 'default'
FOR UPDATE`).Scan(&settings.DefaultStrategy)
	if err != nil {
		return RefundSettings{}, err
	}
	return normalizeStoredRefundSettings(settings), nil
}

func refundedAmountForSQLOrder(ctx context.Context, tx *sql.Tx, orderID string) (int64, error) {
	var refundedFen int64
	err := tx.QueryRowContext(ctx, `
SELECT COALESCE(SUM(amount_fen), 0)
FROM refund_transactions
WHERE order_id = $1
  AND status IN ($2, $3)`,
		strings.TrimSpace(orderID),
		RefundStatusSuccess,
		RefundStatusPendingOriginal,
	).Scan(&refundedFen)
	if err != nil {
		return 0, err
	}
	if refundedFen < 0 {
		return 0, nil
	}
	return refundedFen, nil
}

func loadSQLRefundTransactionByIdempotency(ctx context.Context, tx *sql.Tx, idempotencyKey string) (RefundTransaction, bool, error) {
	var refund RefundTransaction
	err := tx.QueryRowContext(ctx, `
SELECT id, order_id, user_id, amount_fen, destination, status, reason, idempotency_key, created_at
FROM refund_transactions
WHERE idempotency_key = $1
FOR UPDATE`, strings.TrimSpace(idempotencyKey)).Scan(
		&refund.ID,
		&refund.OrderID,
		&refund.UserID,
		&refund.AmountFen,
		&refund.Destination,
		&refund.Status,
		&refund.Reason,
		&refund.IdempotencyKey,
		&refund.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return RefundTransaction{}, false, nil
	}
	if err != nil {
		return RefundTransaction{}, false, err
	}
	refund.CreatedAt = refund.CreatedAt.UTC()
	return refund, true, nil
}

func refundTransactionMatchesRequest(refund RefundTransaction, req RefundOrderRequest) bool {
	if strings.TrimSpace(refund.OrderID) != strings.TrimSpace(req.OrderID) {
		return false
	}
	if userID := strings.TrimSpace(req.UserID); userID != "" && strings.TrimSpace(refund.UserID) != userID {
		return false
	}
	if req.AmountFen > 0 && refund.AmountFen != req.AmountFen {
		return false
	}
	if destination := strings.TrimSpace(req.Destination); destination != "" && refund.Destination != RefundDestinationForStrategy(RefundStrategyBalanceFirst, destination) {
		return false
	}
	return true
}

func refundOrderEvent(refund RefundTransaction, actorID string) OrderEvent {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		actorID = "admin"
	}
	event := OrderEvent{
		Type:      "order.refund.requested",
		ActorID:   actorID,
		Message:   "订单退款已提交原路返回处理",
		AmountFen: refund.AmountFen,
		CreatedAt: refund.CreatedAt.UTC(),
	}
	if refund.Destination == RefundDestinationBalance && refund.Status == RefundStatusSuccess {
		event.Type = "order.refund.success"
		event.Message = "订单退款已退回平台余额"
	}
	return event
}

func insertSQLRefundTransaction(ctx context.Context, tx *sql.Tx, refund RefundTransaction) error {
	result, err := tx.ExecContext(ctx, `
INSERT INTO refund_transactions (
  id, order_id, user_id, amount_fen, destination, status,
  reason, idempotency_key, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
ON CONFLICT (idempotency_key) DO NOTHING`,
		refund.ID,
		refund.OrderID,
		refund.UserID,
		nonNegativeInt64(refund.AmountFen),
		refund.Destination,
		refund.Status,
		refund.Reason,
		refund.IdempotencyKey,
		refund.CreatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrInvalidOrderState
	}
	return nil
}

func insertSQLWalletRefundTransaction(ctx context.Context, tx *sql.Tx, transaction WalletTransaction, balanceAfterFen int64) error {
	result, err := tx.ExecContext(ctx, `
INSERT INTO wallet_transactions (
  id, subject_type, subject_id, order_id, type, amount_fen,
  payment_method, idempotency_key, status, balance_after_fen, created_at
)
VALUES ($1, 'user', $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (idempotency_key) DO NOTHING`,
		transaction.ID,
		transaction.UserID,
		nullableString(transaction.OrderID),
		transaction.Type,
		transaction.AmountFen,
		transaction.PaymentMethod,
		transaction.IdempotencyKey,
		transaction.Status,
		balanceAfterFen,
		transaction.CreatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrInvalidOrderState
	}
	return nil
}

func updateSQLOrderAfterRefund(ctx context.Context, tx *sql.Tx, orderID string, status string, event OrderEvent) error {
	result, err := tx.ExecContext(ctx, `
UPDATE orders
SET status = $2,
    updated_at = $3
WHERE id = $1`,
		strings.TrimSpace(orderID),
		strings.TrimSpace(status),
		event.CreatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrNotFound
	}
	return insertSQLOrderEvent(ctx, tx, orderID, event)
}

func postgresRefundWalletTransactionID(idempotencyKey string) string {
	return "wtx_refund_pg_" + shortHash(strings.TrimSpace(idempotencyKey))
}

func postgresWalletTransactionID(idempotencyKey string) string {
	return "wtx_pg_" + shortHash(strings.TrimSpace(idempotencyKey))
}

func (s *PostgresStore) loadSQLOrders(ctx context.Context) ([]Order, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_id, shop_id, rider_id, address_id, type, status, payment_method,
       items_total_fen, delivery_fee_fen, packaging_fee_fen, discount_fen,
       amount_fen, options, created_at, updated_at
FROM orders
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var order Order
		var shopID sql.NullString
		var riderID sql.NullString
		var addressID sql.NullString
		var optionsPayload []byte
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&shopID,
			&riderID,
			&addressID,
			&order.Type,
			&order.Status,
			&order.PaymentMethod,
			&order.ItemsTotalFen,
			&order.DeliveryFeeFen,
			&order.PackagingFeeFen,
			&order.DiscountFen,
			&order.AmountFen,
			&optionsPayload,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if shopID.Valid {
			order.ShopID = shopID.String
		}
		if riderID.Valid {
			order.RiderID = riderID.String
		}
		if addressID.Valid {
			order.AddressID = addressID.String
		}
		if len(optionsPayload) > 0 {
			if err := json.Unmarshal(optionsPayload, &order.Options); err != nil {
				return nil, err
			}
		}
		order.CreatedAt = order.CreatedAt.UTC()
		order.UpdatedAt = order.UpdatedAt.UTC()
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	itemsByOrderID, err := s.loadSQLOrderItems(ctx)
	if err != nil {
		return nil, err
	}
	eventsByOrderID, err := s.loadSQLOrderEvents(ctx)
	if err != nil {
		return nil, err
	}
	for index := range orders {
		orders[index].Items = itemsByOrderID[orders[index].ID]
		orders[index].Events = eventsByOrderID[orders[index].ID]
	}
	return orders, nil
}

func (s *PostgresStore) loadSQLOrderItems(ctx context.Context) (map[string][]OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT order_id, product_id, product_name_snapshot, unit_price_fen, quantity
FROM order_items
ORDER BY order_id, product_id, product_name_snapshot`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	itemsByOrderID := map[string][]OrderItem{}
	for rows.Next() {
		var orderID string
		var item OrderItem
		if err := rows.Scan(&orderID, &item.ProductID, &item.ProductName, &item.UnitPriceFen, &item.Quantity); err != nil {
			return nil, err
		}
		itemsByOrderID[orderID] = append(itemsByOrderID[orderID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return itemsByOrderID, nil
}

func (s *PostgresStore) loadSQLOrderEvents(ctx context.Context) (map[string][]OrderEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT order_id, type, actor_id, message, created_at
FROM order_events
ORDER BY order_id, created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	eventsByOrderID := map[string][]OrderEvent{}
	for rows.Next() {
		var orderID string
		var event OrderEvent
		if err := rows.Scan(&orderID, &event.Type, &event.ActorID, &event.Message, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.CreatedAt = event.CreatedAt.UTC()
		eventsByOrderID[orderID] = append(eventsByOrderID[orderID], event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return eventsByOrderID, nil
}

func (s *PostgresStore) loadSQLWalletAccounts(ctx context.Context) ([]WalletAccount, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT subject_id, balance_fen, frozen_fen, version, risk_state
FROM wallet_accounts
WHERE subject_type = 'user'
ORDER BY subject_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	wallets := []WalletAccount{}
	for rows.Next() {
		var wallet WalletAccount
		if err := rows.Scan(&wallet.UserID, &wallet.Balance, &wallet.Frozen, &wallet.Version, &wallet.RiskState); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (s *PostgresStore) loadSQLWalletTransactions(ctx context.Context) ([]WalletTransaction, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, subject_id, order_id, type, amount_fen, payment_method, idempotency_key, status, created_at
FROM wallet_transactions
WHERE subject_type = 'user'
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := []WalletTransaction{}
	for rows.Next() {
		var transaction WalletTransaction
		var orderID sql.NullString
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&orderID,
			&transaction.Type,
			&transaction.AmountFen,
			&transaction.PaymentMethod,
			&transaction.IdempotencyKey,
			&transaction.Status,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}
		if orderID.Valid {
			transaction.OrderID = orderID.String
		}
		transaction.CreatedAt = transaction.CreatedAt.UTC()
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *PostgresStore) loadSQLWalletPaymentPasswords(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT user_id, password_hash
FROM wallet_payment_passwords
WHERE status = $1
ORDER BY user_id`, WalletPaymentPasswordSet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hashes := map[string]string{}
	for rows.Next() {
		var userID string
		var passwordHash string
		if err := rows.Scan(&userID, &passwordHash); err != nil {
			return nil, err
		}
		hashes[userID] = passwordHash
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hashes, nil
}

func (s *PostgresStore) loadSQLPaymentTransactions(ctx context.Context) ([]PaymentTransaction, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, order_id, user_id, method, amount_fen, status, out_trade_no,
       transaction_id, idempotency_key, created_at, updated_at
FROM payment_transactions
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := []PaymentTransaction{}
	for rows.Next() {
		var transaction PaymentTransaction
		var transactionID sql.NullString
		if err := rows.Scan(
			&transaction.ID,
			&transaction.OrderID,
			&transaction.UserID,
			&transaction.Method,
			&transaction.AmountFen,
			&transaction.Status,
			&transaction.OutTradeNo,
			&transactionID,
			&transaction.IdempotencyKey,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if transactionID.Valid {
			transaction.TransactionID = transactionID.String
		}
		transaction.CreatedAt = transaction.CreatedAt.UTC()
		transaction.UpdatedAt = transaction.UpdatedAt.UTC()
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *PostgresStore) loadSQLRefundSettings(ctx context.Context) (RefundSettings, error) {
	var settings RefundSettings
	err := s.db.QueryRowContext(ctx, `
SELECT default_strategy
FROM refund_settings
WHERE id = 'default'`).Scan(&settings.DefaultStrategy)
	if errors.Is(err, sql.ErrNoRows) {
		return RefundSettings{DefaultStrategy: RefundStrategyBalanceFirst}, nil
	}
	if err != nil {
		return RefundSettings{}, err
	}
	return normalizeStoredRefundSettings(settings), nil
}

func (s *PostgresStore) loadSQLRefundTransactions(ctx context.Context) ([]RefundTransaction, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, order_id, user_id, amount_fen, destination, status, reason, idempotency_key, created_at
FROM refund_transactions
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refunds := []RefundTransaction{}
	for rows.Next() {
		var refund RefundTransaction
		if err := rows.Scan(
			&refund.ID,
			&refund.OrderID,
			&refund.UserID,
			&refund.AmountFen,
			&refund.Destination,
			&refund.Status,
			&refund.Reason,
			&refund.IdempotencyKey,
			&refund.CreatedAt,
		); err != nil {
			return nil, err
		}
		refund.CreatedAt = refund.CreatedAt.UTC()
		refunds = append(refunds, refund)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return refunds, nil
}

func (s *PostgresStore) loadSQLAfterSalesRequests(ctx context.Context) ([]AfterSalesRequest, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, order_id, user_id, type, reason, requested_amount_fen,
       COALESCE(evidence_urls::text, '[]'), status, review_reason,
       reviewer_id, reviewer_role, refund_id, created_at, updated_at, reviewed_at
FROM order_after_sales
ORDER BY created_at DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := []AfterSalesRequest{}
	for rows.Next() {
		var request AfterSalesRequest
		var evidencePayload string
		var refundID sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&request.ID,
			&request.OrderID,
			&request.UserID,
			&request.Type,
			&request.Reason,
			&request.RequestedAmountFen,
			&evidencePayload,
			&request.Status,
			&request.ReviewReason,
			&request.ReviewerID,
			&request.ReviewerRole,
			&refundID,
			&request.CreatedAt,
			&request.UpdatedAt,
			&reviewedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(evidencePayload), &request.EvidenceURLs); err != nil {
			return nil, err
		}
		request.EvidenceURLs = sanitizedStringSlice(request.EvidenceURLs)
		if refundID.Valid {
			request.RefundID = refundID.String
		}
		if reviewedAt.Valid {
			request.ReviewedAt = reviewedAt.Time.UTC()
		}
		request.CreatedAt = request.CreatedAt.UTC()
		request.UpdatedAt = request.UpdatedAt.UTC()
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *PostgresStore) loadSQLAfterSalesEvents(ctx context.Context) ([]AfterSalesEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, request_id, order_id, actor_id, actor_role, action,
       message, COALESCE(attachments::text, '[]'), visible_to_user, created_at
FROM order_after_sales_events
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []AfterSalesEvent{}
	for rows.Next() {
		var event AfterSalesEvent
		var attachmentsPayload string
		if err := rows.Scan(
			&event.ID,
			&event.RequestID,
			&event.OrderID,
			&event.ActorID,
			&event.ActorRole,
			&event.Action,
			&event.Message,
			&attachmentsPayload,
			&event.VisibleToUser,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(attachmentsPayload), &event.Attachments); err != nil {
			return nil, err
		}
		event.Action = normalizeAfterSalesAction(event.Action)
		event.Attachments = sanitizedStringSlice(event.Attachments)
		event.CreatedAt = event.CreatedAt.UTC()
		if event.Action == "" {
			continue
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStore) loadSQLAfterSalesUploadTickets(ctx context.Context) ([]AfterSalesEvidenceUploadTicket, error) {
	rows, err := s.db.QueryContext(ctx, `
	SELECT id, request_id, order_id, provider, bucket, object_key, public_url,
	       file_name, content_type, size_bytes, max_size_bytes, content_sha,
	       uploaded_by_id, uploaded_by_role, status, scan_status, scan_result,
	       created_at, expires_at, uploaded_at, confirmed_at, scan_checked_at,
	       cleanup_reason, deleted_at, cleanup_attempts, last_cleanup_error, last_cleanup_failed_at
	FROM order_after_sales_evidence_upload_tickets
	ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tickets := []AfterSalesEvidenceUploadTicket{}
	for rows.Next() {
		var ticket AfterSalesEvidenceUploadTicket
		var uploadedAt sql.NullTime
		var confirmedAt sql.NullTime
		var scanCheckedAt sql.NullTime
		var deletedAt sql.NullTime
		var lastCleanupFailedAt sql.NullTime
		if err := rows.Scan(
			&ticket.ID,
			&ticket.RequestID,
			&ticket.OrderID,
			&ticket.Provider,
			&ticket.Bucket,
			&ticket.ObjectKey,
			&ticket.PublicURL,
			&ticket.FileName,
			&ticket.ContentType,
			&ticket.SizeBytes,
			&ticket.MaxSizeBytes,
			&ticket.ContentSHA,
			&ticket.UploadedByID,
			&ticket.UploadedByRole,
			&ticket.Status,
			&ticket.ScanStatus,
			&ticket.ScanResult,
			&ticket.CreatedAt,
			&ticket.ExpiresAt,
			&uploadedAt,
			&confirmedAt,
			&scanCheckedAt,
			&ticket.CleanupReason,
			&deletedAt,
			&ticket.CleanupAttempts,
			&ticket.LastCleanupError,
			&lastCleanupFailedAt,
		); err != nil {
			return nil, err
		}
		ticket.FileName = sanitizeObjectFileName(ticket.FileName)
		ticket.ContentType = normalizeEvidenceContentType(ticket.ContentType)
		ticket.ContentSHA = strings.TrimSpace(ticket.ContentSHA)
		ticket.Status = normalizeAfterSalesUploadTicketStatus(ticket.Status)
		ticket.ScanStatus = normalizeAfterSalesUploadScanStatus(ticket.ScanStatus)
		ticket.ScanResult = strings.TrimSpace(ticket.ScanResult)
		ticket.CleanupReason = normalizeObjectStorageCleanupReason(ticket.CleanupReason)
		ticket.LastCleanupError = sanitizeObjectStorageCleanupError(ticket.LastCleanupError)
		if ticket.CleanupAttempts < 0 {
			ticket.CleanupAttempts = 0
		}
		ticket.CreatedAt = ticket.CreatedAt.UTC()
		ticket.ExpiresAt = ticket.ExpiresAt.UTC()
		if uploadedAt.Valid {
			ticket.UploadedAt = uploadedAt.Time.UTC()
		}
		if confirmedAt.Valid {
			ticket.ConfirmedAt = confirmedAt.Time.UTC()
		}
		if scanCheckedAt.Valid {
			ticket.ScanCheckedAt = scanCheckedAt.Time.UTC()
		}
		if deletedAt.Valid {
			ticket.DeletedAt = deletedAt.Time.UTC()
		}
		if lastCleanupFailedAt.Valid {
			ticket.LastCleanupFailedAt = lastCleanupFailedAt.Time.UTC()
		}
		if ticket.ID == "" || ticket.RequestID == "" || ticket.ObjectKey == "" || ticket.PublicURL == "" || ticket.FileName == "" || ticket.ContentType == "" || ticket.Status == "" || ticket.ScanStatus == "" {
			continue
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func loadSQLAfterSalesUploadTicketForUpdate(ctx context.Context, tx *sql.Tx, ticketID string) (AfterSalesEvidenceUploadTicket, error) {
	var ticket AfterSalesEvidenceUploadTicket
	var uploadedAt sql.NullTime
	var confirmedAt sql.NullTime
	var scanCheckedAt sql.NullTime
	var deletedAt sql.NullTime
	var lastCleanupFailedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
	SELECT id, request_id, order_id, provider, bucket, object_key, public_url,
	       file_name, content_type, size_bytes, max_size_bytes, content_sha,
	       uploaded_by_id, uploaded_by_role, status, scan_status, scan_result,
	       created_at, expires_at, uploaded_at, confirmed_at, scan_checked_at,
	       cleanup_reason, deleted_at, cleanup_attempts, last_cleanup_error, last_cleanup_failed_at
	FROM order_after_sales_evidence_upload_tickets
	WHERE id = $1
	FOR UPDATE`, strings.TrimSpace(ticketID)).Scan(
		&ticket.ID,
		&ticket.RequestID,
		&ticket.OrderID,
		&ticket.Provider,
		&ticket.Bucket,
		&ticket.ObjectKey,
		&ticket.PublicURL,
		&ticket.FileName,
		&ticket.ContentType,
		&ticket.SizeBytes,
		&ticket.MaxSizeBytes,
		&ticket.ContentSHA,
		&ticket.UploadedByID,
		&ticket.UploadedByRole,
		&ticket.Status,
		&ticket.ScanStatus,
		&ticket.ScanResult,
		&ticket.CreatedAt,
		&ticket.ExpiresAt,
		&uploadedAt,
		&confirmedAt,
		&scanCheckedAt,
		&ticket.CleanupReason,
		&deletedAt,
		&ticket.CleanupAttempts,
		&ticket.LastCleanupError,
		&lastCleanupFailedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AfterSalesEvidenceUploadTicket{}, ErrInvalidArgument
	}
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	ticket.FileName = sanitizeObjectFileName(ticket.FileName)
	ticket.ContentType = normalizeEvidenceContentType(ticket.ContentType)
	ticket.ContentSHA = strings.TrimSpace(ticket.ContentSHA)
	ticket.Status = normalizeAfterSalesUploadTicketStatus(ticket.Status)
	ticket.ScanStatus = normalizeAfterSalesUploadScanStatus(ticket.ScanStatus)
	ticket.ScanResult = strings.TrimSpace(ticket.ScanResult)
	ticket.CleanupReason = normalizeObjectStorageCleanupReason(ticket.CleanupReason)
	ticket.LastCleanupError = sanitizeObjectStorageCleanupError(ticket.LastCleanupError)
	if ticket.CleanupAttempts < 0 {
		ticket.CleanupAttempts = 0
	}
	ticket.CreatedAt = ticket.CreatedAt.UTC()
	ticket.ExpiresAt = ticket.ExpiresAt.UTC()
	if uploadedAt.Valid {
		ticket.UploadedAt = uploadedAt.Time.UTC()
	}
	if confirmedAt.Valid {
		ticket.ConfirmedAt = confirmedAt.Time.UTC()
	}
	if scanCheckedAt.Valid {
		ticket.ScanCheckedAt = scanCheckedAt.Time.UTC()
	}
	if deletedAt.Valid {
		ticket.DeletedAt = deletedAt.Time.UTC()
	}
	if lastCleanupFailedAt.Valid {
		ticket.LastCleanupFailedAt = lastCleanupFailedAt.Time.UTC()
	}
	if ticket.ID == "" || ticket.RequestID == "" || ticket.ObjectKey == "" || ticket.PublicURL == "" || ticket.FileName == "" || ticket.ContentType == "" || ticket.Status == "" || ticket.ScanStatus == "" {
		return AfterSalesEvidenceUploadTicket{}, ErrInvalidArgument
	}
	return ticket, nil
}

func completeObjectStorageCleanupInSQLTx(ctx context.Context, tx *sql.Tx, req ObjectStorageCleanupCompleteRequest) (AfterSalesEvidenceUploadTicket, error) {
	normalized, err := normalizeObjectStorageCleanupCompleteRequest(req)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	ticket, err := loadSQLAfterSalesUploadTicketForUpdate(ctx, tx, normalized.TicketID)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	temp := &Store{afterSalesUploadTickets: map[string]*AfterSalesEvidenceUploadTicket{
		ticket.ID: cloneAfterSalesEvidenceUploadTicket(&ticket),
	}}
	updated, err := temp.completeObjectStorageCleanupLocked(normalized)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	if updated == nil {
		return AfterSalesEvidenceUploadTicket{}, ErrInvalidArgument
	}
	if err := upsertSQLAfterSalesUploadTicket(ctx, tx, *updated); err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	return *updated, nil
}

func recordObjectStorageCleanupFailureInSQLTx(ctx context.Context, tx *sql.Tx, req ObjectStorageCleanupFailureRequest) (AfterSalesEvidenceUploadTicket, error) {
	normalized, err := normalizeObjectStorageCleanupFailureRequest(req)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	ticket, err := loadSQLAfterSalesUploadTicketForUpdate(ctx, tx, normalized.TicketID)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	temp := &Store{afterSalesUploadTickets: map[string]*AfterSalesEvidenceUploadTicket{
		ticket.ID: cloneAfterSalesEvidenceUploadTicket(&ticket),
	}}
	updated, err := temp.recordObjectStorageCleanupFailureLocked(normalized)
	if err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	if updated == nil {
		return AfterSalesEvidenceUploadTicket{}, ErrInvalidArgument
	}
	if err := upsertSQLAfterSalesUploadTicket(ctx, tx, *updated); err != nil {
		return AfterSalesEvidenceUploadTicket{}, err
	}
	return *updated, nil
}

func (s *PostgresStore) loadSQLAfterSalesEvidence(ctx context.Context) ([]AfterSalesEvidence, error) {
	rows, err := s.db.QueryContext(ctx, `
	SELECT id, request_id, order_id, object_key, public_url, file_name,
	       content_type, size_bytes, content_sha, uploaded_by_id, uploaded_by_role,
	       status, created_at, confirmed_at
	FROM order_after_sales_evidence
	ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	evidence := []AfterSalesEvidence{}
	for rows.Next() {
		var item AfterSalesEvidence
		if err := rows.Scan(
			&item.ID,
			&item.RequestID,
			&item.OrderID,
			&item.ObjectKey,
			&item.PublicURL,
			&item.FileName,
			&item.ContentType,
			&item.SizeBytes,
			&item.ContentSHA,
			&item.UploadedByID,
			&item.UploadedByRole,
			&item.Status,
			&item.CreatedAt,
			&item.ConfirmedAt,
		); err != nil {
			return nil, err
		}
		item.FileName = sanitizeObjectFileName(item.FileName)
		item.ContentType = normalizeEvidenceContentType(item.ContentType)
		item.CreatedAt = item.CreatedAt.UTC()
		item.ConfirmedAt = item.ConfirmedAt.UTC()
		if item.ID == "" || item.RequestID == "" || item.ObjectKey == "" || item.PublicURL == "" || item.FileName == "" || item.ContentType == "" || item.Status != AfterSalesEvidenceUploaded {
			continue
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return evidence, nil
}

func (s *PostgresStore) loadSQLDispatchEvents(ctx context.Context) ([]DispatchEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, order_id, station_id, mode, type, rider_id, actor_id, reason,
       idempotency_key, online_candidate_size,
       COALESCE(to_json(rejected_rider_ids)::text, '[]'),
       can_decline_without_penalty, created_at
FROM dispatch_events
ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []DispatchEvent{}
	for rows.Next() {
		var event DispatchEvent
		var rejectedPayload string
		if err := rows.Scan(
			&event.ID,
			&event.OrderID,
			&event.StationID,
			&event.Mode,
			&event.Type,
			&event.RiderID,
			&event.ActorID,
			&event.Reason,
			&event.IdempotencyKey,
			&event.OnlineCandidateSize,
			&rejectedPayload,
			&event.CanDeclineWithoutPenalty,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(rejectedPayload), &event.RejectedRiderIDs); err != nil {
			return nil, err
		}
		event.RejectedRiderIDs = normalizedStringSlice(event.RejectedRiderIDs)
		event.CreatedAt = event.CreatedAt.UTC()
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) replacePaymentDomainFromTables(orders []Order, wallets []WalletAccount, walletTransactions []WalletTransaction, paymentPasswordHash map[string]string, paymentTransactions []PaymentTransaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders = map[string]*Order{}
	for _, order := range orders {
		orderCopy := order
		s.orders[order.ID] = cloneOrder(&orderCopy)
		if value, err := strconv.ParseUint(strings.TrimPrefix(order.ID, "ord_"), 10, 64); err == nil && value > s.nextOrderID {
			s.nextOrderID = value
		}
	}
	s.wallets = map[string]*WalletAccount{}
	for _, wallet := range wallets {
		walletCopy := wallet
		s.wallets[wallet.UserID] = cloneWalletAccount(&walletCopy)
	}
	s.walletIdempotency = map[string]*WalletTransaction{}
	for _, transaction := range walletTransactions {
		transactionCopy := transaction
		if strings.TrimSpace(transactionCopy.IdempotencyKey) != "" {
			s.walletIdempotency[transactionCopy.IdempotencyKey] = cloneWalletTransaction(&transactionCopy)
		}
		if value, err := strconv.ParseUint(strings.TrimPrefix(transactionCopy.ID, "wtx_"), 10, 64); err == nil && value > s.nextTransactionID {
			s.nextTransactionID = value
		}
	}
	s.paymentPasswordHash = nonNilMap(paymentPasswordHash)
	s.paymentTransactions = map[string]*PaymentTransaction{}
	for _, transaction := range paymentTransactions {
		transactionCopy := transaction
		s.paymentTransactions[transaction.ID] = clonePaymentTransaction(&transactionCopy)
		if value, err := strconv.ParseUint(strings.TrimPrefix(transaction.ID, "ptx_"), 10, 64); err == nil && value > s.nextTransactionID {
			s.nextTransactionID = value
		}
	}
	s.relinkIndexesLocked()
}

func (s *Store) replaceRefundDomainFromTables(settings RefundSettings, refunds []RefundTransaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundSettings = normalizeStoredRefundSettings(settings)
	s.refundTransactions = map[string]*RefundTransaction{}
	s.refundByIdempotency = map[string]string{}
	for _, refund := range refunds {
		refundCopy := refund
		refundCopy.Destination = RefundDestinationForStrategy(RefundStrategyBalanceFirst, refundCopy.Destination)
		if refundCopy.ID == "" {
			refundCopy.ID = "rfd_" + shortHash(refundCopy.IdempotencyKey)
		}
		s.refundTransactions[refundCopy.ID] = cloneRefundTransaction(&refundCopy)
		if strings.TrimSpace(refundCopy.IdempotencyKey) != "" {
			s.refundByIdempotency[refundCopy.IdempotencyKey] = refundCopy.ID
		}
	}
}

func (s *Store) replaceAfterSalesDomainFromTables(requests []AfterSalesRequest, events ...[]AfterSalesEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.afterSalesRequests = map[string]*AfterSalesRequest{}
	for _, request := range requests {
		requestCopy := request
		requestCopy.Type = normalizeAfterSalesType(requestCopy.Type)
		requestCopy.EvidenceURLs = sanitizedStringSlice(requestCopy.EvidenceURLs)
		if requestCopy.ID == "" || requestCopy.Type == "" {
			continue
		}
		s.afterSalesRequests[requestCopy.ID] = cloneAfterSalesRequest(&requestCopy)
		if value, err := strconv.ParseUint(strings.TrimPrefix(requestCopy.ID, "asr_"), 10, 64); err == nil && value > s.nextAfterSalesID {
			s.nextAfterSalesID = value
		}
	}
	s.afterSalesEvents = map[string]*AfterSalesEvent{}
	if len(events) == 0 {
		return
	}
	for _, event := range events[0] {
		eventCopy := event
		eventCopy.Action = normalizeAfterSalesAction(eventCopy.Action)
		eventCopy.Attachments = sanitizedStringSlice(eventCopy.Attachments)
		if eventCopy.ID == "" || eventCopy.RequestID == "" || eventCopy.Action == "" {
			continue
		}
		s.afterSalesEvents[eventCopy.ID] = cloneAfterSalesEvent(&eventCopy)
		if value, err := strconv.ParseUint(strings.TrimPrefix(eventCopy.ID, "asev_"), 10, 64); err == nil && value > s.nextAfterSalesEventID {
			s.nextAfterSalesEventID = value
		}
	}
}

func (s *Store) replaceAfterSalesEvidenceFromTables(evidence []AfterSalesEvidence) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.afterSalesEvidence = map[string]*AfterSalesEvidence{}
	for _, item := range evidence {
		itemCopy := item
		itemCopy.FileName = sanitizeObjectFileName(itemCopy.FileName)
		itemCopy.ContentType = normalizeEvidenceContentType(itemCopy.ContentType)
		itemCopy.Status = strings.TrimSpace(itemCopy.Status)
		if itemCopy.Status == "" {
			itemCopy.Status = AfterSalesEvidenceUploaded
		}
		if itemCopy.ID == "" || itemCopy.RequestID == "" || itemCopy.ObjectKey == "" || itemCopy.FileName == "" || itemCopy.ContentType == "" || itemCopy.Status != AfterSalesEvidenceUploaded {
			continue
		}
		s.afterSalesEvidence[itemCopy.ID] = cloneAfterSalesEvidence(&itemCopy)
		if request := s.afterSalesRequests[itemCopy.RequestID]; request != nil {
			request.EvidenceURLs = sanitizedStringSlice(append(request.EvidenceURLs, itemCopy.PublicURL))
		}
	}
}

func (s *Store) replaceAfterSalesUploadTicketsFromTables(tickets []AfterSalesEvidenceUploadTicket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.afterSalesUploadTickets = map[string]*AfterSalesEvidenceUploadTicket{}
	for _, ticket := range tickets {
		ticketCopy := ticket
		ticketCopy.FileName = sanitizeObjectFileName(ticketCopy.FileName)
		ticketCopy.ContentType = normalizeEvidenceContentType(ticketCopy.ContentType)
		ticketCopy.Status = normalizeAfterSalesUploadTicketStatus(ticketCopy.Status)
		ticketCopy.ScanStatus = normalizeAfterSalesUploadScanStatus(ticketCopy.ScanStatus)
		ticketCopy.CleanupReason = normalizeObjectStorageCleanupReason(ticketCopy.CleanupReason)
		ticketCopy.LastCleanupError = sanitizeObjectStorageCleanupError(ticketCopy.LastCleanupError)
		if ticketCopy.CleanupAttempts < 0 {
			ticketCopy.CleanupAttempts = 0
		}
		if ticketCopy.ID == "" || ticketCopy.RequestID == "" || ticketCopy.ObjectKey == "" || ticketCopy.FileName == "" || ticketCopy.ContentType == "" || ticketCopy.Status == "" || ticketCopy.ScanStatus == "" {
			continue
		}
		s.afterSalesUploadTickets[ticketCopy.ID] = cloneAfterSalesEvidenceUploadTicket(&ticketCopy)
	}
}

func (s *Store) afterSalesReviewResult(requestID string, refundID string) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, error) {
	requestID = strings.TrimSpace(requestID)
	refundID = strings.TrimSpace(refundID)
	if requestID == "" {
		return nil, nil, nil, nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	request := s.afterSalesRequests[requestID]
	if request == nil {
		return nil, nil, nil, nil, ErrNotFound
	}
	order := s.orders[request.OrderID]
	if order == nil {
		return nil, nil, nil, nil, ErrNotFound
	}
	var refund *RefundTransaction
	var account *WalletAccount
	if refundID != "" {
		refund = s.refundTransactions[refundID]
		if refund == nil {
			return nil, nil, nil, nil, ErrNotFound
		}
		if refund.Destination == RefundDestinationBalance {
			account = s.wallets[refund.UserID]
		}
	}
	return s.afterSalesRequestViewLocked(request), cloneRefundTransaction(refund), cloneOrder(order), cloneWalletAccount(account), nil
}

func (s *Store) refundResult(idempotencyKey string) (*RefundTransaction, *Order, *WalletAccount, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return nil, nil, nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	refundID := s.refundByIdempotency[idempotencyKey]
	refund := s.refundTransactions[refundID]
	if refund == nil {
		return nil, nil, nil, ErrNotFound
	}
	order := s.orders[refund.OrderID]
	if order == nil {
		return nil, nil, nil, ErrNotFound
	}
	var account *WalletAccount
	if refund.Destination == RefundDestinationBalance {
		account = s.wallets[refund.UserID]
	}
	return cloneRefundTransaction(refund), cloneOrder(order), cloneWalletAccount(account), nil
}

func (s *PostgresStore) reloadPaymentDomainAndOutboxAfterSQLPayment(ctx context.Context, orderID string, paidAt time.Time) error {
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return err
	}
	s.Store.applyBalancePaymentSideEffectsAfterSQL(orderID, paidAt)
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return err
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) reloadPaymentDomainAndOutboxAfterSQLRefund(ctx context.Context, orderID string, event OrderEvent) error {
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return err
	}
	s.Store.applyOrderEventOutboxAfterSQL(orderID, event)
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return err
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) reloadPaymentDomainAndOutboxAfterSQLAfterSales(ctx context.Context, orderID string, events []OrderEvent) error {
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return err
	}
	for _, event := range events {
		s.Store.applyOrderEventOutboxAfterSQL(orderID, event)
	}
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return err
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx context.Context, orderID string, event OrderEvent) error {
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return err
	}
	s.Store.applyOrderEventOutboxAfterSQL(orderID, event)
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return err
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) reloadPaymentDomainAndOutboxAfterSQLCheckout(ctx context.Context, orderID string, userID string, shopID string, event OrderEvent) error {
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return err
	}
	s.Store.applyOrderEventOutboxAfterSQL(orderID, event)
	s.Store.clearCartAfterSQLCheckout(userID, shopID)
	if err := s.syncSnapshotOutboxToTable(ctx); err != nil {
		return err
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) clearCartAfterSQLCheckout(userID string, shopID string) {
	userID = strings.TrimSpace(userID)
	shopID = strings.TrimSpace(shopID)
	if userID == "" || shopID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cartItems, cartKey(userID, shopID))
}

func (s *Store) applyBalancePaymentSideEffectsAfterSQL(orderID string, paidAt time.Time) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return
	}
	if paidAt.IsZero() {
		paidAt = time.Now().UTC()
	} else {
		paidAt = paidAt.UTC()
	}
	s.issueGroupbuyVouchersLocked(order, paidAt)
	var paymentEvent OrderEvent
	for _, event := range order.Events {
		if event.Type != "order.payment.success" {
			continue
		}
		if !paidAt.IsZero() && event.CreatedAt.Equal(paidAt.UTC()) {
			paymentEvent = event
			break
		}
		if paymentEvent.Type == "" || event.CreatedAt.After(paymentEvent.CreatedAt) {
			paymentEvent = event
		}
	}
	if paymentEvent.Type == "" {
		paymentEvent = OrderEvent{
			Type:      "order.payment.success",
			ActorID:   order.UserID,
			Message:   paymentSuccessMessage(order),
			CreatedAt: paidAt.UTC(),
		}
		if paymentEvent.CreatedAt.IsZero() {
			paymentEvent.CreatedAt = time.Now().UTC()
		}
	}
	if paymentEvent.Message == "" {
		paymentEvent.Message = paymentSuccessMessage(order)
	}
	paymentEvent.CreatedAt = paymentEvent.CreatedAt.UTC()
	s.enqueueOutboxEventLocked(
		orderEventOutboxTopic(paymentEvent.Type),
		"order",
		order.ID,
		paymentEvent.Type,
		fmt.Sprintf("order_event:%s:%s:%s", order.ID, paymentEvent.Type, paymentEvent.CreatedAt.Format(time.RFC3339Nano)),
		orderEventOutboxPayload(order, paymentEvent),
		paymentEvent.CreatedAt,
	)
}

func (s *Store) applyOrderEventOutboxAfterSQL(orderID string, event OrderEvent) {
	orderID = strings.TrimSpace(orderID)
	event.Type = strings.TrimSpace(event.Type)
	event.ActorID = strings.TrimSpace(event.ActorID)
	event.Message = strings.TrimSpace(event.Message)
	if orderID == "" || event.Type == "" {
		return
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.orders[orderID]
	if order == nil {
		return
	}
	topic := orderEventOutboxTopic(event.Type)
	if topic == "" {
		return
	}
	s.enqueueOutboxEventLocked(
		topic,
		"order",
		order.ID,
		event.Type,
		fmt.Sprintf("order_event:%s:%s:%s", order.ID, event.Type, event.CreatedAt.Format(time.RFC3339Nano)),
		orderEventOutboxPayload(order, event),
		event.CreatedAt,
	)
}

func (s *Store) balancePaymentResult(idempotencyKey string, orderID string, userID string) (*WalletTransaction, *Order, *WalletAccount, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	orderID = strings.TrimSpace(orderID)
	userID = strings.TrimSpace(userID)
	if idempotencyKey == "" {
		return nil, nil, nil, ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	transaction := s.walletIdempotency[idempotencyKey]
	if transaction == nil {
		return nil, nil, nil, ErrNotFound
	}
	if orderID == "" {
		orderID = transaction.OrderID
	}
	if userID == "" {
		userID = transaction.UserID
	}
	order := s.orders[orderID]
	account := s.wallets[userID]
	if order == nil {
		return nil, nil, nil, ErrNotFound
	}
	return cloneWalletTransaction(transaction), cloneOrder(order), cloneWalletAccount(account), nil
}

func (s *Store) replaceDispatchEventsFromTable(events []DispatchEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dispatchEvents = map[string]*DispatchEvent{}
	s.dispatchRejectedRiders = map[string]map[string]bool{}
	nextDispatchEventID := s.nextDispatchEventID
	for _, event := range events {
		event.ID = strings.TrimSpace(event.ID)
		event.OrderID = strings.TrimSpace(event.OrderID)
		event.Type = strings.TrimSpace(event.Type)
		if event.ID == "" || event.OrderID == "" || event.Type == "" {
			continue
		}
		event.StationID = strings.TrimSpace(event.StationID)
		event.Mode = strings.TrimSpace(event.Mode)
		if event.Mode == "" {
			event.Mode = DispatchModeAutoAssign
		}
		event.RiderID = strings.TrimSpace(event.RiderID)
		event.ActorID = strings.TrimSpace(event.ActorID)
		event.Reason = strings.TrimSpace(event.Reason)
		event.IdempotencyKey = strings.TrimSpace(event.IdempotencyKey)
		event.RejectedRiderIDs = normalizedStringSlice(event.RejectedRiderIDs)
		if event.OnlineCandidateSize < 0 {
			event.OnlineCandidateSize = 0
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = time.Now().UTC()
		} else {
			event.CreatedAt = event.CreatedAt.UTC()
		}
		eventCopy := event
		s.dispatchEvents[event.ID] = cloneDispatchEvent(&eventCopy)
		if value, err := strconv.ParseUint(strings.TrimPrefix(event.ID, "dpe_"), 10, 64); err == nil && value > nextDispatchEventID {
			nextDispatchEventID = value
		}
		if s.dispatchRejectedRiders[event.OrderID] == nil {
			s.dispatchRejectedRiders[event.OrderID] = map[string]bool{}
		}
		for _, riderID := range event.RejectedRiderIDs {
			if riderID = strings.TrimSpace(riderID); riderID != "" {
				s.dispatchRejectedRiders[event.OrderID][riderID] = true
			}
		}
		switch event.Type {
		case "dispatch.rejected", "dispatch.timeout":
			if event.RiderID != "" {
				s.dispatchRejectedRiders[event.OrderID][event.RiderID] = true
			}
		}
		if len(s.dispatchRejectedRiders[event.OrderID]) == 0 {
			delete(s.dispatchRejectedRiders, event.OrderID)
		}
	}
	s.nextDispatchEventID = nextDispatchEventID
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nonNegativeInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func nonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func hashOptional(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return shortHash(value)
}

func orderPaidAt(order Order) time.Time {
	if strings.TrimSpace(order.PaymentMethod) == "" {
		return time.Time{}
	}
	for _, event := range order.Events {
		if event.Type == "order.payment.success" && !event.CreatedAt.IsZero() {
			return event.CreatedAt.UTC()
		}
	}
	if !order.UpdatedAt.IsZero() {
		return order.UpdatedAt.UTC()
	}
	return time.Time{}
}

func postgresTextArrayLiteral(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		parts = append(parts, strconv.Quote(value))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func normalizedStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func (s *PostgresStore) syncSnapshotOutboxToTable(ctx context.Context) error {
	events := s.Store.outboxEventSnapshot()
	if len(events) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, event := range events {
		if err := insertMissingOutboxEvent(ctx, tx, event); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) outboxEventSnapshot() []OutboxEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	events := make([]OutboxEvent, 0, len(s.outboxEvents))
	for _, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		events = append(events, *cloneOutboxEvent(event))
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events
}

func insertMissingOutboxEvent(ctx context.Context, tx *sql.Tx, event OutboxEvent) error {
	normalized := normalizeOutboxEventForSQL(event)
	if normalized.ID == "" || normalized.Topic == "" || normalized.AggregateType == "" || normalized.AggregateID == "" || normalized.EventType == "" || normalized.IdempotencyKey == "" {
		return nil
	}
	payload, err := json.Marshal(normalized.Payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO platform_outbox_events (
  id, topic, aggregate_type, aggregate_id, event_type, idempotency_key,
  payload, status, attempts, last_error, available_at, lease_owner,
  lease_expires_at, published_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12, $13, $14, $15, $16)
ON CONFLICT (idempotency_key) DO NOTHING`,
		normalized.ID,
		normalized.Topic,
		normalized.AggregateType,
		normalized.AggregateID,
		normalized.EventType,
		normalized.IdempotencyKey,
		string(payload),
		normalized.Status,
		normalized.Attempts,
		normalized.LastError,
		normalized.AvailableAt,
		normalized.LeaseOwner,
		nullableTime(normalized.LeaseExpiresAt),
		nullableTime(normalized.PublishedAt),
		normalized.CreatedAt,
		normalized.UpdatedAt,
	)
	return err
}

func insertOrGetSQLOutboxEvent(ctx context.Context, tx *sql.Tx, event OutboxEvent) (*OutboxEvent, error) {
	normalized := normalizeOutboxEventForSQL(event)
	if normalized.ID == "" || normalized.Topic == "" || normalized.AggregateType == "" || normalized.AggregateID == "" || normalized.EventType == "" || normalized.IdempotencyKey == "" {
		return nil, ErrInvalidArgument
	}
	payload, err := json.Marshal(normalized.Payload)
	if err != nil {
		return nil, err
	}
	return scanOutboxEvent(tx.QueryRowContext(ctx, `
INSERT INTO platform_outbox_events AS event (
  id, topic, aggregate_type, aggregate_id, event_type, idempotency_key,
  payload, status, attempts, last_error, available_at, lease_owner,
  lease_expires_at, published_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12, $13, $14, $15, $16)
ON CONFLICT (idempotency_key) DO UPDATE
SET updated_at = event.updated_at
RETURNING `+postgresOutboxEventReturningColumns,
		normalized.ID,
		normalized.Topic,
		normalized.AggregateType,
		normalized.AggregateID,
		normalized.EventType,
		normalized.IdempotencyKey,
		string(payload),
		normalized.Status,
		normalized.Attempts,
		normalized.LastError,
		normalized.AvailableAt,
		normalized.LeaseOwner,
		nullableTime(normalized.LeaseExpiresAt),
		nullableTime(normalized.PublishedAt),
		normalized.CreatedAt,
		normalized.UpdatedAt,
	))
}

func normalizeOutboxEventForSQL(event OutboxEvent) OutboxEvent {
	event.ID = strings.TrimSpace(event.ID)
	event.Topic = strings.TrimSpace(event.Topic)
	event.AggregateType = strings.TrimSpace(event.AggregateType)
	event.AggregateID = strings.TrimSpace(event.AggregateID)
	event.EventType = strings.TrimSpace(event.EventType)
	event.IdempotencyKey = strings.TrimSpace(event.IdempotencyKey)
	event.Status = strings.TrimSpace(event.Status)
	if !IsOutboxStatus(event.Status) {
		event.Status = OutboxStatusPending
	}
	event.LastError = strings.TrimSpace(event.LastError)
	event.LeaseOwner = strings.TrimSpace(event.LeaseOwner)
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	now := time.Now().UTC()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	} else {
		event.CreatedAt = event.CreatedAt.UTC()
	}
	if event.AvailableAt.IsZero() {
		event.AvailableAt = event.CreatedAt
	} else {
		event.AvailableAt = event.AvailableAt.UTC()
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = event.CreatedAt
	} else {
		event.UpdatedAt = event.UpdatedAt.UTC()
	}
	if !event.LeaseExpiresAt.IsZero() {
		event.LeaseExpiresAt = event.LeaseExpiresAt.UTC()
	}
	if !event.PublishedAt.IsZero() {
		event.PublishedAt = event.PublishedAt.UTC()
	}
	return event
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}

type outboxScanner interface {
	Scan(dest ...any) error
}

func scanOutboxEvent(scanner outboxScanner) (*OutboxEvent, error) {
	var event OutboxEvent
	var payload []byte
	var leaseExpiresAt sql.NullTime
	var publishedAt sql.NullTime
	if err := scanner.Scan(
		&event.ID,
		&event.Topic,
		&event.AggregateType,
		&event.AggregateID,
		&event.EventType,
		&event.IdempotencyKey,
		&payload,
		&event.Status,
		&event.Attempts,
		&event.LastError,
		&event.AvailableAt,
		&event.LeaseOwner,
		&leaseExpiresAt,
		&publishedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, err
		}
	}
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	if leaseExpiresAt.Valid {
		event.LeaseExpiresAt = leaseExpiresAt.Time.UTC()
	}
	if publishedAt.Valid {
		event.PublishedAt = publishedAt.Time.UTC()
	}
	event.AvailableAt = event.AvailableAt.UTC()
	event.CreatedAt = event.CreatedAt.UTC()
	event.UpdatedAt = event.UpdatedAt.UTC()
	return &event, nil
}

func scanOutboxEventRows(rows *sql.Rows) ([]OutboxEvent, error) {
	defer rows.Close()
	events := []OutboxEvent{}
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStore) loadOutboxFromTable(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT `+postgresOutboxEventColumns+` FROM platform_outbox_events ORDER BY created_at, id`)
	if err != nil {
		return err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return err
	}
	s.Store.replaceOutboxEvents(events)
	return nil
}

func (s *Store) replaceOutboxEvents(events []OutboxEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outboxEvents = map[string]*OutboxEvent{}
	s.outboxByIdempotency = map[string]string{}
	for _, event := range events {
		eventCopy := event
		s.outboxEvents[event.ID] = cloneOutboxEvent(&eventCopy)
		if event.IdempotencyKey != "" {
			s.outboxByIdempotency[event.IdempotencyKey] = event.ID
		}
		if value, err := strconv.ParseUint(strings.TrimPrefix(event.ID, "obe_"), 10, 64); err == nil && value > s.nextOutboxEventID {
			s.nextOutboxEventID = value
		}
	}
}

func (s *PostgresStore) applyOutboxEventsAndSaveSnapshot(ctx context.Context, events []OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}
	s.Store.applyOutboxEvents(events)
	if err := s.saveSnapshot(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	return nil
}

func (s *PostgresStore) applyOutboxEventsAndAuditAfterCommit(ctx context.Context, events []OutboxEvent, log *AuditLog) error {
	if len(events) > 0 {
		s.Store.applyOutboxEvents(events)
	}
	if log != nil {
		s.Store.applyAuditLogFromSQL(*log)
	}
	if err := s.saveSnapshot(ctx); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	return nil
}

func (s *Store) applyOutboxEvents(events []OutboxEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, event := range events {
		eventCopy := event
		s.outboxEvents[event.ID] = cloneOutboxEvent(&eventCopy)
		if event.IdempotencyKey != "" {
			s.outboxByIdempotency[event.IdempotencyKey] = event.ID
		}
	}
}

func normalizeOutboxLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 500 {
		return 500
	}
	return limit
}

func normalizeOutboxNow(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}
	return now.UTC()
}

func normalizeOutboxLeaseSeconds(leaseSeconds int) int {
	if leaseSeconds <= 0 {
		return 60
	}
	if leaseSeconds > 3600 {
		return 3600
	}
	return leaseSeconds
}

func normalizeOutboxLeaseExpiringWithinSeconds(seconds int) int {
	if seconds <= 0 {
		return 15
	}
	if seconds > 3600 {
		return 3600
	}
	return seconds
}

func outboxSecondsUntil(later time.Time, now time.Time) int64 {
	if later.IsZero() {
		return 0
	}
	later = later.UTC()
	now = now.UTC()
	if !later.After(now) {
		return 0
	}
	return int64(later.Sub(now).Seconds())
}

func earlierOutboxTime(current time.Time, candidate time.Time) time.Time {
	if candidate.IsZero() {
		return current
	}
	candidate = candidate.UTC()
	if current.IsZero() || candidate.Before(current) {
		return candidate
	}
	return current
}

func sortOutboxEvents(events []OutboxEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		leftAvailableAt := events[i].AvailableAt
		if leftAvailableAt.IsZero() {
			leftAvailableAt = events[i].CreatedAt
		}
		rightAvailableAt := events[j].AvailableAt
		if rightAvailableAt.IsZero() {
			rightAvailableAt = events[j].CreatedAt
		}
		if !leftAvailableAt.Equal(rightAvailableAt) {
			return leftAvailableAt.Before(rightAvailableAt)
		}
		if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].ID < events[j].ID
	})
}

func buildOutboxStats(events []OutboxEvent, topic string, now time.Time, leaseExpiringWithinSeconds int) *OutboxStats {
	leaseExpiringWithinSeconds = normalizeOutboxLeaseExpiringWithinSeconds(leaseExpiringWithinSeconds)
	leaseExpiringDeadline := now.Add(time.Duration(leaseExpiringWithinSeconds) * time.Second)
	stats := &OutboxStats{
		GeneratedAt:                now,
		Topic:                      topic,
		LeaseExpiringWithinSeconds: leaseExpiringWithinSeconds,
		LeaseOwners:                []OutboxLeaseOwnerStats{},
		Topics:                     []OutboxTopicStats{},
	}
	topicStatsByTopic := map[string]*OutboxTopicStats{}
	ownerStatsByOwner := map[string]*OutboxLeaseOwnerStats{}
	oldestReadyByTopic := map[string]time.Time{}
	for i := range events {
		event := &events[i]
		eventTopic := strings.TrimSpace(event.Topic)
		if eventTopic == "" {
			eventTopic = "unknown"
		}
		topicStats := topicStatsByTopic[eventTopic]
		if topicStats == nil {
			topicStats = &OutboxTopicStats{Topic: eventTopic}
			topicStatsByTopic[eventTopic] = topicStats
		}
		stats.Total++
		topicStats.Total++
		switch event.Status {
		case OutboxStatusPending:
			stats.Pending++
			topicStats.Pending++
		case OutboxStatusFailed:
			stats.Failed++
			topicStats.Failed++
		case OutboxStatusDeadLetter:
			stats.DeadLetter++
			topicStats.DeadLetter++
		case OutboxStatusPublished:
			stats.Published++
			topicStats.Published++
		}
		if event.Status != OutboxStatusPending && event.Status != OutboxStatusFailed {
			continue
		}
		if outboxLeaseActive(event, now) {
			stats.Leased++
			topicStats.Leased++
			leaseExpiresAt := event.LeaseExpiresAt.UTC()
			stats.NextLeaseExpiresAt = earlierOutboxTime(stats.NextLeaseExpiresAt, leaseExpiresAt)
			topicStats.NextLeaseExpiresAt = earlierOutboxTime(topicStats.NextLeaseExpiresAt, leaseExpiresAt)
			leaseOwner := strings.TrimSpace(event.LeaseOwner)
			if leaseOwner != "" {
				ownerStats := ownerStatsByOwner[leaseOwner]
				if ownerStats == nil {
					ownerStats = &OutboxLeaseOwnerStats{Owner: leaseOwner}
					ownerStatsByOwner[leaseOwner] = ownerStats
				}
				ownerStats.Leased++
				ownerStats.NextLeaseExpiresAt = earlierOutboxTime(ownerStats.NextLeaseExpiresAt, leaseExpiresAt)
				if !leaseExpiresAt.After(leaseExpiringDeadline) {
					ownerStats.LeaseExpiringSoon++
				}
			}
			if !leaseExpiresAt.After(leaseExpiringDeadline) {
				stats.LeaseExpiringSoon++
				topicStats.LeaseExpiringSoon++
			}
			continue
		}
		availableAt := event.AvailableAt
		if availableAt.IsZero() {
			availableAt = event.CreatedAt
		}
		availableAt = availableAt.UTC()
		if availableAt.After(now) {
			stats.Blocked++
			topicStats.Blocked++
			continue
		}
		stats.Ready++
		topicStats.Ready++
		if stats.OldestReadyAt.IsZero() || availableAt.Before(stats.OldestReadyAt) {
			stats.OldestReadyAt = availableAt
		}
		if oldestReadyByTopic[eventTopic].IsZero() || availableAt.Before(oldestReadyByTopic[eventTopic]) {
			oldestReadyByTopic[eventTopic] = availableAt
		}
	}
	if !stats.OldestReadyAt.IsZero() {
		stats.OldestReadyLagSeconds = int64(now.Sub(stats.OldestReadyAt).Seconds())
	}
	stats.NextLeaseExpiresInSeconds = outboxSecondsUntil(stats.NextLeaseExpiresAt, now)
	owners := make([]string, 0, len(ownerStatsByOwner))
	for owner := range ownerStatsByOwner {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	for _, owner := range owners {
		ownerStats := *ownerStatsByOwner[owner]
		ownerStats.NextLeaseExpiresInSeconds = outboxSecondsUntil(ownerStats.NextLeaseExpiresAt, now)
		stats.LeaseOwners = append(stats.LeaseOwners, ownerStats)
	}
	topics := make([]string, 0, len(topicStatsByTopic))
	for topic := range topicStatsByTopic {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	for _, topic := range topics {
		topicStats := *topicStatsByTopic[topic]
		if oldest := oldestReadyByTopic[topic]; !oldest.IsZero() {
			topicStats.OldestReadyLagSeconds = int64(now.Sub(oldest).Seconds())
		}
		topicStats.NextLeaseExpiresInSeconds = outboxSecondsUntil(topicStats.NextLeaseExpiresAt, now)
		stats.Topics = append(stats.Topics, topicStats)
	}
	return stats
}

func (s *PostgresStore) outboxEventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM platform_outbox_events WHERE id = $1)", eventID).Scan(&exists)
	return exists, err
}

func (s *Store) snapshot() storeSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshotLocked()
}

func (s *Store) snapshotPayload() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Marshal(s.snapshotLocked())
}

func (s *Store) snapshotLocked() storeSnapshot {
	return storeSnapshot{
		Version:                 1,
		SavedAt:                 time.Now().UTC(),
		NextOrderID:             s.nextOrderID,
		NextTransactionID:       s.nextTransactionID,
		NextAddressID:           s.nextAddressID,
		NextMerchantID:          s.nextMerchantID,
		NextMerchantStaffID:     s.nextMerchantStaffID,
		NextMerchantMaterialID:  s.nextMerchantMaterialID,
		NextDispatchEventID:     s.nextDispatchEventID,
		NextOutboxEventID:       s.nextOutboxEventID,
		NextAuditLogID:          s.nextAuditLogID,
		NextAfterSalesID:        s.nextAfterSalesID,
		NextAfterSalesEventID:   s.nextAfterSalesEventID,
		NextRiderID:             s.nextRiderID,
		NextProductID:           s.nextProductID,
		NextVoucherID:           s.nextVoucherID,
		HomeModules:             s.homeModules,
		HomeCards:               s.homeCards,
		Users:                   s.users,
		WechatBindings:          s.wechatBindings,
		MerchantInvites:         s.merchantInvites,
		Merchants:               s.merchants,
		MerchantQualifications:  s.merchantQualifications,
		MerchantStaff:           s.merchantStaff,
		MerchantMaterials:       s.merchantMaterials,
		Riders:                  s.riders,
		Deposits:                s.deposits,
		StationTaskConfigs:      s.stationTaskConfigs,
		StationServiceAreas:     s.stationServiceAreas,
		Shops:                   s.shops,
		Products:                s.products,
		GroupbuyDeals:           s.groupbuyDeals,
		Addresses:               s.addresses,
		CartItems:               s.cartItems,
		Orders:                  s.orders,
		Wallets:                 s.wallets,
		PaymentPasswordHash:     s.paymentPasswordHash,
		MerchantPasswordHash:    s.merchantPasswordHash,
		RiderPasswordHash:       s.riderPasswordHash,
		PaymentTransactions:     s.paymentTransactions,
		PaymentByTradeNo:        s.paymentByTradeNo,
		PaymentByProviderID:     s.paymentByProviderID,
		WalletIdempotency:       s.walletIdempotency,
		RefundSettings:          normalizeStoredRefundSettings(s.refundSettings),
		RefundTransactions:      s.refundTransactions,
		RefundByIdempotency:     s.refundByIdempotency,
		AfterSalesRequests:      s.afterSalesRequests,
		AfterSalesEvents:        s.afterSalesEvents,
		AfterSalesUploadTickets: s.afterSalesUploadTickets,
		AfterSalesEvidence:      s.afterSalesEvidence,
		GroupbuyVouchers:        s.groupbuyVouchers,
		VouchersByOrderID:       s.vouchersByOrderID,
		VouchersByCode:          s.vouchersByCode,
		DispatchEvents:          s.dispatchEvents,
		DispatchRejectedRiders:  s.dispatchRejectedRiders,
		FreeCancelUsedByDate:    s.freeCancelUsedByDate,
		OutboxEvents:            s.outboxEvents,
		OutboxByIdempotency:     s.outboxByIdempotency,
		AuditLogs:               s.auditLogs,
	}
}

func (s *Store) applySnapshot(snapshot storeSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextOrderID = snapshot.NextOrderID
	s.nextTransactionID = snapshot.NextTransactionID
	s.nextAddressID = snapshot.NextAddressID
	s.nextMerchantID = snapshot.NextMerchantID
	s.nextMerchantStaffID = snapshot.NextMerchantStaffID
	s.nextMerchantMaterialID = snapshot.NextMerchantMaterialID
	s.nextDispatchEventID = snapshot.NextDispatchEventID
	s.nextOutboxEventID = snapshot.NextOutboxEventID
	s.nextAuditLogID = snapshot.NextAuditLogID
	s.nextAfterSalesID = snapshot.NextAfterSalesID
	s.nextAfterSalesEventID = snapshot.NextAfterSalesEventID
	s.nextRiderID = snapshot.NextRiderID
	s.nextProductID = snapshot.NextProductID
	s.nextVoucherID = snapshot.NextVoucherID
	s.homeModules = snapshot.HomeModules
	s.homeCards = snapshot.HomeCards
	s.users = nonNilMap(snapshot.Users)
	s.wechatBindings = nonNilMap(snapshot.WechatBindings)
	s.merchantInvites = nonNilMap(snapshot.MerchantInvites)
	s.merchants = nonNilMap(snapshot.Merchants)
	s.merchantQualifications = nonNilMap(snapshot.MerchantQualifications)
	s.merchantStaff = nonNilMap(snapshot.MerchantStaff)
	s.merchantMaterials = nonNilMap(snapshot.MerchantMaterials)
	s.riders = nonNilMap(snapshot.Riders)
	s.deposits = nonNilMap(snapshot.Deposits)
	s.stationTaskConfigs = nonNilMap(snapshot.StationTaskConfigs)
	s.stationServiceAreas = nonNilMap(snapshot.StationServiceAreas)
	s.shops = nonNilMap(snapshot.Shops)
	s.products = nonNilMap(snapshot.Products)
	s.groupbuyDeals = nonNilMap(snapshot.GroupbuyDeals)
	s.addresses = nonNilMap(snapshot.Addresses)
	s.cartItems = nonNilMap(snapshot.CartItems)
	s.orders = nonNilMap(snapshot.Orders)
	s.wallets = nonNilMap(snapshot.Wallets)
	s.paymentPasswordHash = nonNilMap(snapshot.PaymentPasswordHash)
	s.merchantPasswordHash = nonNilMap(snapshot.MerchantPasswordHash)
	s.riderPasswordHash = nonNilMap(snapshot.RiderPasswordHash)
	s.paymentTransactions = nonNilMap(snapshot.PaymentTransactions)
	s.paymentByTradeNo = nonNilMap(snapshot.PaymentByTradeNo)
	s.paymentByProviderID = nonNilMap(snapshot.PaymentByProviderID)
	s.walletIdempotency = nonNilMap(snapshot.WalletIdempotency)
	s.refundSettings = normalizeStoredRefundSettings(snapshot.RefundSettings)
	s.refundTransactions = nonNilMap(snapshot.RefundTransactions)
	s.refundByIdempotency = nonNilMap(snapshot.RefundByIdempotency)
	s.afterSalesRequests = nonNilMap(snapshot.AfterSalesRequests)
	s.afterSalesEvents = nonNilMap(snapshot.AfterSalesEvents)
	s.afterSalesUploadTickets = nonNilMap(snapshot.AfterSalesUploadTickets)
	s.afterSalesEvidence = nonNilMap(snapshot.AfterSalesEvidence)
	s.groupbuyVouchers = nonNilMap(snapshot.GroupbuyVouchers)
	s.vouchersByOrderID = nonNilMap(snapshot.VouchersByOrderID)
	s.vouchersByCode = nonNilMap(snapshot.VouchersByCode)
	s.dispatchEvents = nonNilMap(snapshot.DispatchEvents)
	s.dispatchRejectedRiders = nonNilMap(snapshot.DispatchRejectedRiders)
	s.freeCancelUsedByDate = nonNilMap(snapshot.FreeCancelUsedByDate)
	s.outboxEvents = nonNilMap(snapshot.OutboxEvents)
	s.outboxByIdempotency = nonNilMap(snapshot.OutboxByIdempotency)
	s.auditLogs = nonNilMap(snapshot.AuditLogs)
	s.relinkIndexesLocked()
}

func normalizeStoredRefundSettings(settings RefundSettings) RefundSettings {
	return RefundSettings{DefaultStrategy: NormalizeRefundStrategy(settings.DefaultStrategy)}
}

func nonNilMap[M ~map[K]V, K comparable, V any](input M) M {
	if input != nil {
		return input
	}
	return M{}
}

func (s *Store) relinkIndexesLocked() {
	s.paymentByTradeNo = map[string]*PaymentTransaction{}
	s.paymentByProviderID = map[string]*PaymentTransaction{}
	for _, transaction := range s.paymentTransactions {
		if transaction == nil {
			continue
		}
		if transaction.OutTradeNo != "" {
			s.paymentByTradeNo[transaction.OutTradeNo] = transaction
		}
		if transaction.TransactionID != "" {
			s.paymentByProviderID[transaction.TransactionID] = transaction
		}
	}
	s.refundByIdempotency = map[string]string{}
	for id, refund := range s.refundTransactions {
		if refund == nil {
			continue
		}
		if refund.ID == "" {
			refund.ID = id
		}
		if refund.IdempotencyKey != "" {
			s.refundByIdempotency[refund.IdempotencyKey] = refund.ID
		}
	}
	s.vouchersByCode = map[string]*GroupbuyVoucher{}
	s.vouchersByOrderID = map[string][]string{}
	for id, voucher := range s.groupbuyVouchers {
		if voucher == nil {
			continue
		}
		if voucher.VoucherCode != "" {
			s.vouchersByCode[voucher.VoucherCode] = voucher
		}
		if voucher.OrderID != "" {
			s.vouchersByOrderID[voucher.OrderID] = append(s.vouchersByOrderID[voucher.OrderID], id)
		}
	}
	s.outboxByIdempotency = map[string]string{}
	for id, event := range s.outboxEvents {
		if event == nil {
			continue
		}
		if event.IdempotencyKey != "" {
			s.outboxByIdempotency[event.IdempotencyKey] = id
		}
		if event.ID == "" {
			event.ID = id
		}
	}
	for id, log := range s.auditLogs {
		if log == nil {
			continue
		}
		if log.ID == "" {
			log.ID = id
		}
		if value, err := strconv.ParseUint(strings.TrimPrefix(log.ID, "aud_"), 10, 64); err == nil && value > s.nextAuditLogID {
			s.nextAuditLogID = value
		}
	}
}

func (s *PostgresStore) LoginWechatMini(req WechatMiniLoginRequest) (*WechatMiniLoginResult, error) {
	result, err := s.Store.LoginWechatMini(req)
	return result, s.persistAfter(err)
}

func (s *PostgresStore) CreateMerchantInvite(req CreateMerchantInviteRequest) (*MerchantOnboardingInvite, error) {
	invite, err := s.Store.CreateMerchantInvite(req)
	return invite, s.persistAfter(err)
}

func (s *PostgresStore) CreateMerchantInviteWithAudit(req CreateMerchantInviteRequest, audit RecordAuditLogRequest) (*MerchantOnboardingInvite, *AuditLog, error) {
	log, err := inviteAuditLogFromRequest(audit, "admin.merchant_invite.created", "merchant_invite")
	if err != nil {
		return nil, log, err
	}
	invite, err := s.Store.CreateMerchantInvite(req)
	if err != nil {
		return nil, log, err
	}
	log.TargetID = invite.Token
	log.Payload = merchantInviteAuditPayload(invite)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := s.saveSnapshotInTx(ctx, tx); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	return invite, cloneAuditLog(log), nil
}

func (s *PostgresStore) RecordAuditLog(req RecordAuditLogRequest) (*AuditLog, error) {
	log, err := auditLogFromRequest(req, "")
	if err != nil {
		return log, err
	}
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	ensureAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
	if err := insertSQLAuditLog(ctx, tx, *log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
		return log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	return cloneAuditLog(log), nil
}

func (s *PostgresStore) insertAuditLogInTx(ctx context.Context, tx *sql.Tx, log *AuditLog) error {
	if log == nil {
		return ErrInvalidArgument
	}
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	signingSecret := s.Store.auditLogSigningSecretSnapshot()
	ensureAuditLogIntegrity(log, signingSecret)
	if err := insertSQLAuditLog(ctx, tx, *log, signingSecret); err != nil {
		return errors.Join(ErrPersistence, err)
	}
	return nil
}

func (s *PostgresStore) AuditLogs(req AuditLogsRequest) ([]AuditLog, error) {
	logs, err := s.loadSQLAuditLogs(context.Background(), req)
	if err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	return logs, nil
}

func (s *PostgresStore) AuditRetentionReport(req AuditRetentionReportRequest) (*AuditRetentionReport, error) {
	report, err := s.loadSQLAuditRetentionReport(context.Background(), req)
	if err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	return report, nil
}

func (s *PostgresStore) EmitAuditRetentionAlerts(req AuditRetentionAlertEmissionRequest, audit RecordAuditLogRequest) (*AuditRetentionAlertEmission, *OutboxEvent, *AuditLog, error) {
	reportReq := normalizeAuditRetentionReportRequest(AuditRetentionReportRequest{
		RetentionDays:        req.RetentionDays,
		HotDays:              req.HotDays,
		IntegritySampleLimit: req.IntegritySampleLimit,
		Now:                  req.Now,
	})
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = reportReq.Now
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, log, err
	}
	if log.Action != "admin.audit_retention_alerts.emitted" || log.TargetType != "audit_retention_alerts" || log.TargetID != "default" {
		return nil, nil, log, ErrInvalidArgument
	}
	ctx := context.Background()
	report, err := s.loadSQLAuditRetentionReport(ctx, reportReq)
	if err != nil {
		return nil, nil, log, errors.Join(ErrPersistence, err)
	}
	emission := auditRetentionAlertEmissionFromReport(report)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	var event *OutboxEvent
	if len(report.Alerts) > 0 {
		event, err = insertOrGetSQLOutboxEvent(ctx, tx, *auditRetentionAlertOutboxEvent(emission))
		if err != nil {
			return nil, nil, log, errors.Join(ErrPersistence, err)
		}
		emission.OutboxEventID = event.ID
	}
	log.Payload = auditRetentionAlertEmissionAuditPayload(emission)
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, event, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, event, log, errors.Join(ErrPersistence, err)
	}
	events := []OutboxEvent{}
	if event != nil {
		events = append(events, *event)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, events, log); err != nil {
		return emission, event, log, err
	}
	return emission, cloneOutboxEvent(event), cloneAuditLog(log), nil
}

func (s *PostgresStore) AcceptMerchantInvite(req AcceptMerchantInviteRequest) (*MerchantProfile, error) {
	profile, err := s.Store.AcceptMerchantInvite(req)
	return profile, s.persistAfter(err)
}

func (s *PostgresStore) LoginMerchant(req MerchantLoginRequest) (*MerchantProfile, error) {
	return s.Store.LoginMerchant(req)
}

func (s *PostgresStore) CreateRiderInvite(req CreateRiderInviteRequest) (*MerchantOnboardingInvite, error) {
	invite, err := s.Store.CreateRiderInvite(req)
	return invite, s.persistAfter(err)
}

func (s *PostgresStore) CreateRiderInviteWithAudit(req CreateRiderInviteRequest, audit RecordAuditLogRequest) (*MerchantOnboardingInvite, *AuditLog, error) {
	log, err := inviteAuditLogFromRequest(audit, "admin.rider_invite.created", "rider_invite")
	if err != nil {
		return nil, log, err
	}
	invite, err := s.Store.CreateRiderInvite(req)
	if err != nil {
		return nil, log, err
	}
	log.TargetID = invite.Token
	log.Payload = riderInviteAuditPayload(invite)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := s.saveSnapshotInTx(ctx, tx); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	return invite, cloneAuditLog(log), nil
}

func (s *PostgresStore) AcceptRiderInvite(req AcceptRiderInviteRequest) (*RiderAccount, error) {
	rider, err := s.Store.AcceptRiderInvite(req)
	return rider, s.persistAfter(err)
}

func (s *PostgresStore) LoginRider(req RiderLoginRequest) (*RiderAccount, error) {
	return s.Store.LoginRider(req)
}

func (s *PostgresStore) SaveMerchantQualification(req UploadMerchantQualificationRequest) (*MerchantProfile, error) {
	profile, err := s.Store.SaveMerchantQualification(req)
	return profile, s.persistAfter(err)
}

func (s *PostgresStore) SaveMerchantStaff(req UpsertMerchantStaffRequest) (*MerchantStaff, error) {
	staff, err := s.Store.SaveMerchantStaff(req)
	return staff, s.persistAfter(err)
}

func (s *PostgresStore) SaveMerchantSupplementalMaterial(req UploadMerchantSupplementalMaterialRequest) (*MerchantSupplementalMaterial, error) {
	material, err := s.Store.SaveMerchantSupplementalMaterial(req)
	return material, s.persistAfter(err)
}

func (s *PostgresStore) UpsertMerchantProduct(req UpsertMerchantProductRequest) (*MerchantProduct, error) {
	product, err := s.Store.UpsertMerchantProduct(req)
	return product, s.persistAfter(err)
}

func (s *PostgresStore) SetMerchantProductStatus(req SetMerchantProductStatusRequest) (*MerchantProduct, error) {
	product, err := s.Store.SetMerchantProductStatus(req)
	return product, s.persistAfter(err)
}

func (s *PostgresStore) CreateGroupbuyOrder(req CreateGroupbuyOrderRequest) (*Order, error) {
	order, err := s.Store.CreateGroupbuyOrder(req)
	return order, s.persistAfter(err)
}

func (s *PostgresStore) RedeemGroupbuyVoucher(req RedeemGroupbuyVoucherRequest) (*GroupbuyVoucher, *Order, error) {
	voucher, order, err := s.Store.RedeemGroupbuyVoucher(req)
	return voucher, order, s.persistAfter(err)
}

func (s *PostgresStore) SaveAddress(address UserAddress) (*UserAddress, error) {
	saved, err := s.Store.SaveAddress(address)
	return saved, s.persistAfter(err)
}

func (s *PostgresStore) UpsertCartItem(req UpsertCartItemRequest) (*CartSummary, error) {
	summary, err := s.Store.UpsertCartItem(req)
	return summary, s.persistAfter(err)
}

func (s *PostgresStore) CreateOrder(req CreateOrderRequest) (*Order, error) {
	ctx := context.Background()
	sqlOrder, event, err := s.createOrderInSQL(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx, sqlOrder.ID, event); err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	order, err := s.Store.OrderByID(sqlOrder.ID)
	if err != nil {
		return cloneOrder(&sqlOrder), err
	}
	return order, nil
}

func (s *PostgresStore) CheckoutCart(req CheckoutCartRequest) (*Order, *CartSummary, error) {
	ctx := context.Background()
	sqlOrder, summary, event, err := s.checkoutCartInSQL(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLCheckout(ctx, sqlOrder.ID, summary.UserID, summary.ShopID, event); err != nil {
		return nil, nil, errors.Join(ErrPersistence, err)
	}
	order, err := s.Store.OrderByID(sqlOrder.ID)
	if err != nil {
		return cloneOrder(&sqlOrder), cloneCartSummary(&summary), err
	}
	return order, cloneCartSummary(&summary), nil
}

func (s *PostgresStore) CompensateOrderState(req CompensateOrderStateRequest) (*CompensateOrderStateResult, error) {
	normalized, err := normalizeCompensateOrderStateRequest(req)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	result, event, err := s.compensateOrderStateInSQLTx(ctx, tx, normalized)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	if result.Changed {
		if err := s.reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx, normalized.OrderID, event); err != nil {
			return result, errors.Join(ErrPersistence, err)
		}
	}
	return result, nil
}

func (s *PostgresStore) CompensateOrderStateWithAudit(req CompensateOrderStateRequest, audit RecordAuditLogRequest) (*CompensateOrderStateResult, *AuditLog, error) {
	normalized, err := normalizeCompensateOrderStateRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != "admin.order_state.compensated" || log.TargetType != "order" || log.TargetID != normalized.OrderID {
		return nil, log, ErrInvalidArgument
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	result, event, err := s.compensateOrderStateInSQLTx(ctx, tx, normalized)
	if err != nil {
		return nil, log, err
	}
	log.Payload = orderStateCompensationAuditPayload(result)
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	ensureAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
	if err := insertSQLAuditLog(ctx, tx, *log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if result.Changed {
		if err := s.reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx, normalized.OrderID, event); err != nil {
			return result, log, errors.Join(ErrPersistence, err)
		}
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return result, log, err
	}
	return result, cloneAuditLog(log), nil
}

func (s *PostgresStore) MerchantAcceptOrder(orderID string, merchantID string) (*Order, error) {
	ctx := context.Background()
	sqlOrder, event, err := s.transitionMerchantOrderInSQL(
		ctx,
		orderID,
		merchantID,
		StatusMerchantPending,
		StatusPreparing,
		"merchant.accepted",
		"商户已接单，开始备餐",
		true,
	)
	if err != nil {
		return nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx, sqlOrder.ID, event); err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	order, err := s.Store.OrderByID(sqlOrder.ID)
	if err != nil {
		return cloneOrder(&sqlOrder), err
	}
	return order, nil
}

func (s *PostgresStore) MerchantMarkOrderReady(orderID string, merchantID string) (*Order, error) {
	ctx := context.Background()
	sqlOrder, event, err := s.transitionMerchantOrderInSQL(
		ctx,
		orderID,
		merchantID,
		StatusPreparing,
		StatusDispatching,
		"merchant.ready_for_pickup",
		"商户已出餐，订单进入骑手调度",
		false,
	)
	if err != nil {
		return nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLOrderEvent(ctx, sqlOrder.ID, event); err != nil {
		return nil, errors.Join(ErrPersistence, err)
	}
	order, err := s.Store.OrderByID(sqlOrder.ID)
	if err != nil {
		return cloneOrder(&sqlOrder), err
	}
	return order, nil
}

func (s *PostgresStore) CreditWallet(req CreditWalletRequest) (*WalletTransaction, *WalletAccount, error) {
	transaction, account, err := s.Store.CreditWallet(req)
	return transaction, account, s.persistAfter(err)
}

func (s *PostgresStore) SetWalletPaymentPassword(req SetWalletPaymentPasswordRequest) (*WalletPaymentPasswordState, error) {
	state, err := s.Store.SetWalletPaymentPassword(req)
	return state, s.persistAfter(err)
}

func (s *PostgresStore) PayOrderWithBalance(req BalancePayRequest) (*WalletTransaction, *Order, *WalletAccount, error) {
	ctx := context.Background()
	sqlTransaction, sqlAccount, orderID, paidAt, err := s.payOrderWithBalanceInSQL(ctx, req)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLPayment(ctx, orderID, paidAt); err != nil {
		return nil, nil, nil, errors.Join(ErrPersistence, err)
	}
	transaction, order, account, err := s.Store.balancePaymentResult(strings.TrimSpace(req.IdempotencyKey), orderID, strings.TrimSpace(req.UserID))
	if err != nil {
		return cloneWalletTransaction(&sqlTransaction), nil, cloneWalletAccount(&sqlAccount), err
	}
	if account == nil {
		account = cloneWalletAccount(&sqlAccount)
	}
	if transaction == nil {
		transaction = cloneWalletTransaction(&sqlTransaction)
	}
	return transaction, order, account, nil
}

func (s *PostgresStore) RefundSettings() (*RefundSettings, error) {
	return s.Store.RefundSettings()
}

func (s *PostgresStore) SaveRefundSettings(req SaveRefundSettingsRequest) (*RefundSettings, error) {
	settings, err := s.Store.SaveRefundSettings(req)
	return settings, s.persistAfter(err)
}

func (s *PostgresStore) SaveRefundSettingsWithAudit(req SaveRefundSettingsRequest, audit RecordAuditLogRequest) (*RefundSettings, *AuditLog, error) {
	settings := normalizeStoredRefundSettings(RefundSettings{DefaultStrategy: req.DefaultStrategy})
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != "admin.refund_settings.updated" || log.TargetType != "refund_settings" || log.TargetID != "default" {
		return nil, log, ErrInvalidArgument
	}
	log.Payload = map[string]any{"default_refund_strategy": settings.DefaultStrategy}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := upsertSQLRefundSettings(ctx, tx, settings); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	ensureAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
	if err := insertSQLAuditLog(ctx, tx, *log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if _, err := s.Store.SaveRefundSettings(req); err != nil {
		return nil, log, err
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return nil, log, err
	}
	return &settings, cloneAuditLog(log), nil
}

func (s *PostgresStore) RefundOrder(req RefundOrderRequest) (*RefundTransaction, *Order, *WalletAccount, error) {
	ctx := context.Background()
	sqlRefund, sqlAccount, orderID, event, err := s.refundOrderInSQL(ctx, req)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLRefund(ctx, orderID, event); err != nil {
		return nil, nil, nil, errors.Join(ErrPersistence, err)
	}
	refund, order, account, err := s.Store.refundResult(strings.TrimSpace(req.IdempotencyKey))
	if err != nil {
		return cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), err
	}
	if refund == nil {
		refund = cloneRefundTransaction(&sqlRefund)
	}
	if account == nil && sqlRefund.Destination == RefundDestinationBalance {
		account = cloneWalletAccount(&sqlAccount)
	}
	return refund, order, account, nil
}

func (s *PostgresStore) RefundOrderWithAudit(req RefundOrderRequest, audit RecordAuditLogRequest) (*RefundTransaction, *Order, *WalletAccount, *AuditLog, error) {
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Reason = strings.TrimSpace(req.Reason)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, nil, log, err
	}
	if log.Action != "admin.order.refunded" || log.TargetType != "order" || log.TargetID != req.OrderID {
		return nil, nil, nil, log, ErrInvalidArgument
	}
	if req.OrderID == "" || req.Reason == "" || req.IdempotencyKey == "" {
		return nil, nil, nil, log, ErrInvalidArgument
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	sqlRefund, sqlAccount, orderID, event, err := refundOrderInSQLTx(ctx, tx, req)
	if err != nil {
		return nil, nil, nil, log, err
	}
	log.Payload = refundOrderAuditPayload(&sqlRefund)
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	ensureAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
	if err := insertSQLAuditLog(ctx, tx, *log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
		return nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLRefund(ctx, orderID, event); err != nil {
		return cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), log, err
	}
	refund, order, account, err := s.Store.refundResult(req.IdempotencyKey)
	if err != nil {
		return cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), cloneAuditLog(log), err
	}
	if refund == nil {
		refund = cloneRefundTransaction(&sqlRefund)
	}
	if account == nil && sqlRefund.Destination == RefundDestinationBalance {
		account = cloneWalletAccount(&sqlAccount)
	}
	return refund, order, account, cloneAuditLog(log), nil
}

func (s *PostgresStore) CreateAfterSales(req CreateAfterSalesRequest) (*AfterSalesRequest, error) {
	request, err := s.Store.CreateAfterSales(req)
	return request, s.persistAfter(err)
}

func (s *PostgresStore) UserAfterSalesRequests(userID string) ([]AfterSalesRequest, error) {
	return s.Store.UserAfterSalesRequests(userID)
}

func (s *PostgresStore) MerchantAfterSalesRequests(merchantID string) ([]AfterSalesRequest, error) {
	return s.Store.MerchantAfterSalesRequests(merchantID)
}

func (s *PostgresStore) AdminAfterSalesRequests() ([]AfterSalesRequest, error) {
	return s.Store.AdminAfterSalesRequests()
}

func (s *PostgresStore) AfterSalesEvents(requestID string, actorID string, actorRole string) ([]AfterSalesEvent, error) {
	return s.Store.AfterSalesEvents(requestID, actorID, actorRole)
}

func (s *PostgresStore) AddAfterSalesEvent(req AddAfterSalesEventRequest) (*AfterSalesEvent, *AfterSalesRequest, error) {
	event, request, err := s.Store.AddAfterSalesEvent(req)
	return event, request, s.persistAfter(err)
}

func (s *PostgresStore) CreateAfterSalesEvidenceUpload(req CreateAfterSalesEvidenceUploadRequest) (*ObjectUploadTicket, error) {
	ticket, err := s.Store.CreateAfterSalesEvidenceUpload(req)
	return ticket, s.persistAfter(err)
}

func (s *PostgresStore) ConfirmObjectStorageUpload(req ObjectStorageUploadCallbackRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket, err := s.Store.ConfirmObjectStorageUpload(req)
	return ticket, s.persistAfter(err)
}

func (s *PostgresStore) RecordObjectStorageScanResult(req ObjectStorageScanResultRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket, err := s.Store.RecordObjectStorageScanResult(req)
	return ticket, s.persistAfter(err)
}

func (s *PostgresStore) ObjectStorageCleanupCandidates(req ObjectStorageCleanupCandidatesRequest) ([]ObjectStorageCleanupCandidate, error) {
	return s.Store.ObjectStorageCleanupCandidates(req)
}

func (s *PostgresStore) CompleteObjectStorageCleanup(req ObjectStorageCleanupCompleteRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket, err := s.Store.CompleteObjectStorageCleanup(req)
	return ticket, s.persistAfter(err)
}

func (s *PostgresStore) CompleteObjectStorageCleanupWithAudit(req ObjectStorageCleanupCompleteRequest, audit RecordAuditLogRequest) (*AfterSalesEvidenceUploadTicket, *AuditLog, error) {
	normalized, err := normalizeObjectStorageCleanupCompleteRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != "admin.object_cleanup.completed" || log.TargetType != "object_storage_ticket" || log.TargetID != normalized.TicketID {
		return nil, log, ErrInvalidArgument
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	ticket, err := completeObjectStorageCleanupInSQLTx(ctx, tx, normalized)
	if err != nil {
		return nil, log, err
	}
	log.Payload = objectStorageCleanupCompletedAuditPayload(&ticket)
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	signingSecret := s.Store.auditLogSigningSecretSnapshot()
	ensureAuditLogIntegrity(log, signingSecret)
	if err := insertSQLAuditLog(ctx, tx, *log, signingSecret); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return cloneAfterSalesEvidenceUploadTicket(&ticket), log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return cloneAfterSalesEvidenceUploadTicket(&ticket), log, err
	}
	return cloneAfterSalesEvidenceUploadTicket(&ticket), cloneAuditLog(log), nil
}

func (s *PostgresStore) RecordObjectStorageCleanupFailure(req ObjectStorageCleanupFailureRequest) (*AfterSalesEvidenceUploadTicket, error) {
	ticket, err := s.Store.RecordObjectStorageCleanupFailure(req)
	return ticket, s.persistAfter(err)
}

func (s *PostgresStore) RecordObjectStorageCleanupFailureWithAudit(req ObjectStorageCleanupFailureRequest, audit RecordAuditLogRequest) (*AfterSalesEvidenceUploadTicket, *AuditLog, error) {
	normalized, err := normalizeObjectStorageCleanupFailureRequest(req)
	if err != nil {
		return nil, nil, err
	}
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, log, err
	}
	if log.Action != "admin.object_cleanup.failed" || log.TargetType != "object_storage_ticket" || log.TargetID != normalized.TicketID {
		return nil, log, ErrInvalidArgument
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	ticket, err := recordObjectStorageCleanupFailureInSQLTx(ctx, tx, normalized)
	if err != nil {
		return nil, log, err
	}
	log.Payload = objectStorageCleanupFailedAuditPayload(&ticket)
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	signingSecret := s.Store.auditLogSigningSecretSnapshot()
	ensureAuditLogIntegrity(log, signingSecret)
	if err := insertSQLAuditLog(ctx, tx, *log, signingSecret); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.loadPaymentDomainFromTables(ctx); err != nil {
		return cloneAfterSalesEvidenceUploadTicket(&ticket), log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return cloneAfterSalesEvidenceUploadTicket(&ticket), log, err
	}
	return cloneAfterSalesEvidenceUploadTicket(&ticket), cloneAuditLog(log), nil
}

func (s *PostgresStore) ObjectStorageCleanupStats(req ObjectStorageCleanupCandidatesRequest) (*ObjectStorageCleanupStats, error) {
	return s.Store.ObjectStorageCleanupStats(req)
}

func (s *PostgresStore) ConfirmAfterSalesEvidenceUpload(req ConfirmAfterSalesEvidenceUploadRequest) (*AfterSalesEvidence, *AfterSalesEvent, *AfterSalesRequest, error) {
	evidence, event, request, err := s.Store.ConfirmAfterSalesEvidenceUpload(req)
	return evidence, event, request, s.persistAfter(err)
}

func (s *PostgresStore) AfterSalesEvidence(requestID string, actorID string, actorRole string) ([]AfterSalesEvidence, error) {
	return s.Store.AfterSalesEvidence(requestID, actorID, actorRole)
}

func (s *PostgresStore) ReviewAfterSales(req ReviewAfterSalesRequest) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, error) {
	ctx := context.Background()
	result, sqlRefund, sqlAccount, err := s.reviewAfterSalesInSQL(ctx, req)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLAfterSales(ctx, result.OrderID, result.Events); err != nil {
		return nil, nil, nil, nil, errors.Join(ErrPersistence, err)
	}
	request, refund, order, account, err := s.Store.afterSalesReviewResult(result.RequestID, result.RefundID)
	if err != nil {
		var fallbackRefund *RefundTransaction
		var fallbackAccount *WalletAccount
		if result.RefundID != "" {
			fallbackRefund = cloneRefundTransaction(&sqlRefund)
			if sqlRefund.Destination == RefundDestinationBalance {
				fallbackAccount = cloneWalletAccount(&sqlAccount)
			}
		}
		return nil, fallbackRefund, nil, fallbackAccount, err
	}
	if refund == nil && result.RefundID != "" {
		refund = cloneRefundTransaction(&sqlRefund)
	}
	if account == nil && sqlRefund.Destination == RefundDestinationBalance {
		account = cloneWalletAccount(&sqlAccount)
	}
	return request, refund, order, account, nil
}

func (s *PostgresStore) ReviewAfterSalesWithAudit(req ReviewAfterSalesRequest, audit RecordAuditLogRequest) (*AfterSalesRequest, *RefundTransaction, *Order, *WalletAccount, *AuditLog, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Decision = strings.TrimSpace(req.Decision)
	req.Reason = strings.TrimSpace(req.Reason)
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.ActorRole = strings.TrimSpace(req.ActorRole)
	req.RefundIdempotencyKey = strings.TrimSpace(req.RefundIdempotencyKey)
	log, err := auditLogFromRequest(audit, "")
	if err != nil {
		return nil, nil, nil, nil, log, err
	}
	if log.Action != "after_sales.reviewed" || log.TargetType != "after_sales" || log.TargetID != req.RequestID {
		return nil, nil, nil, nil, log, ErrInvalidArgument
	}
	if req.RequestID == "" || req.Decision == "" || req.ActorID == "" || req.ActorRole == "" {
		return nil, nil, nil, nil, log, ErrInvalidArgument
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	result, sqlRefund, sqlAccount, err := reviewAfterSalesInSQLTx(ctx, tx, req)
	if err != nil {
		return nil, nil, nil, nil, log, err
	}
	auditRequest := &AfterSalesRequest{ID: result.RequestID, Status: result.Status, RefundID: result.RefundID}
	var auditRefund *RefundTransaction
	if result.RefundID != "" {
		auditRefund = &sqlRefund
	}
	log.Payload = afterSalesReviewAuditPayload(req, auditRequest, auditRefund)
	auditLogNumber, err := nextSQLAuditLogNumber(ctx, tx)
	if err != nil {
		return nil, nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	log.ID = fmt.Sprintf("aud_%d", auditLogNumber)
	ensureAuditLogIntegrity(log, s.Store.auditLogSigningSecretSnapshot())
	if err := insertSQLAuditLog(ctx, tx, *log, s.Store.auditLogSigningSecretSnapshot()); err != nil {
		return nil, nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.reloadPaymentDomainAndOutboxAfterSQLAfterSales(ctx, result.OrderID, result.Events); err != nil {
		return nil, cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), log, errors.Join(ErrPersistence, err)
	}
	s.Store.applyAuditLogFromSQL(*log)
	if err := s.persistAfter(nil); err != nil {
		return nil, cloneRefundTransaction(&sqlRefund), nil, cloneWalletAccount(&sqlAccount), log, err
	}
	request, refund, order, account, err := s.Store.afterSalesReviewResult(result.RequestID, result.RefundID)
	if err != nil {
		var fallbackRefund *RefundTransaction
		var fallbackAccount *WalletAccount
		if result.RefundID != "" {
			fallbackRefund = cloneRefundTransaction(&sqlRefund)
			if sqlRefund.Destination == RefundDestinationBalance {
				fallbackAccount = cloneWalletAccount(&sqlAccount)
			}
		}
		return nil, fallbackRefund, nil, fallbackAccount, cloneAuditLog(log), err
	}
	if refund == nil && result.RefundID != "" {
		refund = cloneRefundTransaction(&sqlRefund)
	}
	if account == nil && sqlRefund.Destination == RefundDestinationBalance {
		account = cloneWalletAccount(&sqlAccount)
	}
	return request, refund, order, account, cloneAuditLog(log), nil
}

func (s *PostgresStore) CreateWechatPrepay(req WechatPrepayRequest) (*WechatPrepayResponse, *PaymentTransaction, error) {
	prepay, transaction, err := s.Store.CreateWechatPrepay(req)
	return prepay, transaction, s.persistAfter(err)
}

func (s *PostgresStore) ConfirmWechatPayment(req WechatPaymentCallbackRequest) (*PaymentTransaction, *Order, error) {
	transaction, order, err := s.Store.ConfirmWechatPayment(req)
	return transaction, order, s.persistAfter(err)
}

func (s *PostgresStore) DepositAccount(subjectType string, subjectID string) (*DepositAccount, error) {
	deposit, err := s.Store.DepositAccount(subjectType, subjectID)
	return deposit, s.persistAfter(err)
}

func (s *PostgresStore) PayDeposit(req PayDepositRequest) (*DepositAccount, error) {
	deposit, err := s.Store.PayDeposit(req)
	return deposit, s.persistAfter(err)
}

func (s *PostgresStore) ApproveRiderWechatExemption(req RiderWechatExemptionRequest) (*DepositAccount, *RiderAccount, error) {
	deposit, rider, err := s.Store.ApproveRiderWechatExemption(req)
	return deposit, rider, s.persistAfter(err)
}

func (s *PostgresStore) RequestRiderDepositRefund(req RiderDepositRefundRequest) (*DepositAccount, *RiderAccount, error) {
	deposit, rider, err := s.Store.RequestRiderDepositRefund(req)
	return deposit, rider, s.persistAfter(err)
}

func (s *PostgresStore) SetRiderOnlineStatus(req SetRiderOnlineStatusRequest) (*RiderAccount, error) {
	rider, err := s.Store.SetRiderOnlineStatus(req)
	return rider, s.persistAfter(err)
}

func (s *PostgresStore) AutoAssignOrder(req AutoAssignOrderRequest) (*Order, *DispatchDecision, error) {
	order, decision, err := s.Store.AutoAssignOrder(req)
	return order, decision, s.persistAfter(err)
}

func (s *PostgresStore) RejectRiderAssignment(req RejectRiderAssignmentRequest) (*Order, *DispatchDecision, error) {
	order, decision, err := s.Store.RejectRiderAssignment(req)
	return order, decision, s.persistAfter(err)
}

func (s *PostgresStore) TimeoutReassignOrder(req TimeoutReassignOrderRequest) (*Order, *DispatchDecision, error) {
	order, decision, err := s.Store.TimeoutReassignOrder(req)
	return order, decision, s.persistAfter(err)
}

func (s *PostgresStore) ManualAssignOrder(req ManualAssignOrderRequest) (*Order, *DispatchDecision, error) {
	order, decision, err := s.Store.ManualAssignOrder(req)
	return order, decision, s.persistAfter(err)
}

func (s *PostgresStore) StationTaskConfig(stationManagerID string) (*StationTaskConfig, error) {
	config, err := s.Store.StationTaskConfig(stationManagerID)
	return config, s.persistAfter(err)
}

func (s *PostgresStore) SaveStationTaskConfig(req SaveStationTaskConfigRequest) (*StationTaskConfig, error) {
	config, err := s.Store.SaveStationTaskConfig(req)
	return config, s.persistAfter(err)
}

func (s *PostgresStore) StationRiderPerformance(stationManagerID string) ([]RiderPerformance, error) {
	performance, err := s.Store.StationRiderPerformance(stationManagerID)
	return performance, s.persistAfter(err)
}

func (s *PostgresStore) DispatchEvents(orderID string, stationManagerID string) ([]DispatchEvent, error) {
	return s.Store.DispatchEvents(orderID, stationManagerID)
}

func (s *PostgresStore) OutboxEvents(req OutboxEventsRequest) ([]OutboxEvent, error) {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = OutboxStatusPending
	}
	if !IsOutboxStatus(status) {
		return nil, ErrInvalidArgument
	}
	topic := strings.TrimSpace(req.Topic)
	limit := normalizeOutboxLimit(req.Limit)
	now := normalizeOutboxNow(req.Now)

	var rows *sql.Rows
	var err error
	if status == OutboxStatusPending {
		rows, err = s.db.QueryContext(context.Background(), `SELECT `+postgresOutboxEventColumns+`
FROM platform_outbox_events
WHERE ($1 = '' OR topic = $1)
  AND status IN ('pending', 'failed')
  AND available_at <= $2
  AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
ORDER BY available_at, created_at, id
LIMIT $3`, topic, now, limit)
	} else if status == OutboxStatusFailed {
		rows, err = s.db.QueryContext(context.Background(), `SELECT `+postgresOutboxEventColumns+`
FROM platform_outbox_events
WHERE ($1 = '' OR topic = $1)
  AND status = 'failed'
  AND available_at <= $2
  AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
ORDER BY available_at, created_at, id
LIMIT $3`, topic, now, limit)
	} else {
		rows, err = s.db.QueryContext(context.Background(), `SELECT `+postgresOutboxEventColumns+`
FROM platform_outbox_events
WHERE ($1 = '' OR topic = $1)
  AND status = $2
ORDER BY available_at, created_at, id
LIMIT $3`, topic, status, limit)
	}
	if err != nil {
		return nil, err
	}
	return scanOutboxEventRows(rows)
}

func (s *PostgresStore) OutboxStats(req OutboxStatsRequest) (*OutboxStats, error) {
	topic := strings.TrimSpace(req.Topic)
	now := normalizeOutboxNow(req.Now)
	rows, err := s.db.QueryContext(context.Background(), `SELECT `+postgresOutboxEventColumns+`
FROM platform_outbox_events
WHERE ($1 = '' OR topic = $1)
ORDER BY topic, available_at, created_at, id`, topic)
	if err != nil {
		return nil, err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return nil, err
	}
	return buildOutboxStats(events, topic, now, req.LeaseExpiringWithinSeconds), nil
}

func (s *PostgresStore) MarkOutboxEventPublished(req MarkOutboxEventPublishedRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	publishedAt := normalizeOutboxNow(req.PublishedAt)
	event, err := scanOutboxEvent(s.db.QueryRowContext(context.Background(), `UPDATE platform_outbox_events event
SET status = 'published',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    published_at = $2,
    updated_at = $2
WHERE event.id = $1
RETURNING `+postgresOutboxEventReturningColumns, eventID, publishedAt))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), []OutboxEvent{*event}); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *PostgresStore) MarkOutboxEventPublishedWithAudit(req MarkOutboxEventPublishedRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.published", "outbox_event", eventID)
	if err != nil {
		return nil, log, err
	}
	publishedAt := normalizeOutboxNow(req.PublishedAt)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	event, err := scanOutboxEvent(tx.QueryRowContext(ctx, `UPDATE platform_outbox_events event
SET status = 'published',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    published_at = $2,
    updated_at = $2
WHERE event.id = $1
RETURNING `+postgresOutboxEventReturningColumns, eventID, publishedAt))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, log, ErrNotFound
	}
	if err != nil {
		return nil, log, err
	}
	log.Payload = outboxEventAuditPayload(event)
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, []OutboxEvent{*event}, log); err != nil {
		return event, log, err
	}
	return event, cloneAuditLog(log), nil
}

func (s *PostgresStore) MarkOutboxEventFailed(req MarkOutboxEventFailedRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	now := normalizeOutboxNow(req.Now)
	retryAfterSeconds := req.RetryAfterSeconds
	if retryAfterSeconds <= 0 {
		retryAfterSeconds = 60
	}
	maxAttempts := req.MaxAttempts
	message := strings.TrimSpace(req.Error)
	retryAt := now.Add(time.Duration(retryAfterSeconds) * time.Second)
	event, err := scanOutboxEvent(s.db.QueryRowContext(context.Background(), `UPDATE platform_outbox_events event
SET attempts = attempts + 1,
    status = CASE WHEN $5 > 0 AND attempts + 1 >= $5 THEN 'dead_letter' ELSE 'failed' END,
    last_error = $2,
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = CASE WHEN $5 > 0 AND attempts + 1 >= $5 THEN $3 ELSE $4 END,
    updated_at = $3
WHERE event.id = $1
RETURNING `+postgresOutboxEventReturningColumns, eventID, message, now, retryAt, maxAttempts))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), []OutboxEvent{*event}); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *PostgresStore) MarkOutboxEventFailedWithAudit(req MarkOutboxEventFailedRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.failed", "outbox_event", eventID)
	if err != nil {
		return nil, log, err
	}
	now := normalizeOutboxNow(req.Now)
	retryAfterSeconds := req.RetryAfterSeconds
	if retryAfterSeconds <= 0 {
		retryAfterSeconds = 60
	}
	maxAttempts := req.MaxAttempts
	message := strings.TrimSpace(req.Error)
	retryAt := now.Add(time.Duration(retryAfterSeconds) * time.Second)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	event, err := scanOutboxEvent(tx.QueryRowContext(ctx, `UPDATE platform_outbox_events event
SET attempts = attempts + 1,
    status = CASE WHEN $5 > 0 AND attempts + 1 >= $5 THEN 'dead_letter' ELSE 'failed' END,
    last_error = $2,
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = CASE WHEN $5 > 0 AND attempts + 1 >= $5 THEN $3 ELSE $4 END,
    updated_at = $3
WHERE event.id = $1
RETURNING `+postgresOutboxEventReturningColumns, eventID, message, now, retryAt, maxAttempts))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, log, ErrNotFound
	}
	if err != nil {
		return nil, log, err
	}
	log.Payload = outboxEventAuditPayload(event)
	log.Payload["retry_after_seconds"] = retryAfterSeconds
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, []OutboxEvent{*event}, log); err != nil {
		return event, log, err
	}
	return event, cloneAuditLog(log), nil
}

func (s *PostgresStore) ClaimOutboxEvents(req ClaimOutboxEventsRequest) (*ClaimOutboxEventsResult, error) {
	topic := strings.TrimSpace(req.Topic)
	limit := normalizeOutboxLimit(req.Limit)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if leaseOwner == "" {
		leaseOwner = "outbox-relay"
	}
	leaseSeconds := normalizeOutboxLeaseSeconds(req.LeaseSeconds)
	now := normalizeOutboxNow(req.Now)
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
	rows, err := s.db.QueryContext(context.Background(), `WITH candidates AS (
  SELECT id
  FROM platform_outbox_events
  WHERE ($1 = '' OR topic = $1)
    AND status IN ('pending', 'failed')
    AND available_at <= $2
    AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
  ORDER BY available_at, created_at, id
  LIMIT $3
  FOR UPDATE SKIP LOCKED
)
UPDATE platform_outbox_events event
SET lease_owner = $4,
    lease_expires_at = $5,
    updated_at = $2
FROM candidates
WHERE event.id = candidates.id
RETURNING `+postgresOutboxEventReturningColumns, topic, now, limit, leaseOwner, leaseExpiresAt)
	if err != nil {
		return nil, err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return nil, err
	}
	sortOutboxEvents(events)
	result := &ClaimOutboxEventsResult{
		Topic:          topic,
		Limit:          limit,
		LeaseOwner:     leaseOwner,
		LeaseExpiresAt: leaseExpiresAt,
		Claimed:        len(events),
		Events:         events,
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), events); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) ClaimOutboxEventsWithAudit(req ClaimOutboxEventsRequest, audit RecordAuditLogRequest) (*ClaimOutboxEventsResult, *AuditLog, error) {
	topic := strings.TrimSpace(req.Topic)
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.claimed", "outbox_topic", outboxTopicAuditTarget(topic))
	if err != nil {
		return nil, log, err
	}
	limit := normalizeOutboxLimit(req.Limit)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if leaseOwner == "" {
		leaseOwner = "outbox-relay"
	}
	leaseSeconds := normalizeOutboxLeaseSeconds(req.LeaseSeconds)
	now := normalizeOutboxNow(req.Now)
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	rows, err := tx.QueryContext(ctx, `WITH candidates AS (
  SELECT id
  FROM platform_outbox_events
  WHERE ($1 = '' OR topic = $1)
    AND status IN ('pending', 'failed')
    AND available_at <= $2
    AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
  ORDER BY available_at, created_at, id
  LIMIT $3
  FOR UPDATE SKIP LOCKED
)
UPDATE platform_outbox_events event
SET lease_owner = $4,
    lease_expires_at = $5,
    updated_at = $2
FROM candidates
WHERE event.id = candidates.id
RETURNING `+postgresOutboxEventReturningColumns, topic, now, limit, leaseOwner, leaseExpiresAt)
	if err != nil {
		return nil, log, err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return nil, log, err
	}
	sortOutboxEvents(events)
	result := &ClaimOutboxEventsResult{
		Topic:          topic,
		Limit:          limit,
		LeaseOwner:     leaseOwner,
		LeaseExpiresAt: leaseExpiresAt,
		Claimed:        len(events),
		Events:         events,
	}
	log.Payload = outboxClaimAuditPayload(result, leaseSeconds)
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, events, log); err != nil {
		return result, log, err
	}
	return result, cloneAuditLog(log), nil
}

func (s *PostgresStore) RenewOutboxEventLease(req RenewOutboxEventLeaseRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if eventID == "" || leaseOwner == "" {
		return nil, ErrInvalidArgument
	}
	leaseSeconds := normalizeOutboxLeaseSeconds(req.LeaseSeconds)
	now := normalizeOutboxNow(req.Now)
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
	event, err := scanOutboxEvent(s.db.QueryRowContext(context.Background(), `UPDATE platform_outbox_events event
SET lease_expires_at = $4,
    updated_at = $3
WHERE event.id = $1
  AND event.status IN ('pending', 'failed')
  AND event.lease_owner = $2
  AND event.lease_expires_at > $3
RETURNING `+postgresOutboxEventReturningColumns, eventID, leaseOwner, now, leaseExpiresAt))
	if errors.Is(err, sql.ErrNoRows) {
		exists, existsErr := s.outboxEventExists(context.Background(), eventID)
		if existsErr != nil {
			return nil, existsErr
		}
		if !exists {
			return nil, ErrNotFound
		}
		return nil, ErrInvalidOrderState
	}
	if err != nil {
		return nil, err
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), []OutboxEvent{*event}); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *PostgresStore) RenewOutboxEventLeaseWithAudit(req RenewOutboxEventLeaseRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	leaseOwner := strings.TrimSpace(req.LeaseOwner)
	if eventID == "" || leaseOwner == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.lease_renewed", "outbox_event", eventID)
	if err != nil {
		return nil, log, err
	}
	leaseSeconds := normalizeOutboxLeaseSeconds(req.LeaseSeconds)
	now := normalizeOutboxNow(req.Now)
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	event, err := scanOutboxEvent(tx.QueryRowContext(ctx, `UPDATE platform_outbox_events event
SET lease_expires_at = $4,
    updated_at = $3
WHERE event.id = $1
  AND event.status IN ('pending', 'failed')
  AND event.lease_owner = $2
  AND event.lease_expires_at > $3
RETURNING `+postgresOutboxEventReturningColumns, eventID, leaseOwner, now, leaseExpiresAt))
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if existsErr := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM platform_outbox_events WHERE id = $1)", eventID).Scan(&exists); existsErr != nil {
			return nil, log, existsErr
		}
		if !exists {
			return nil, log, ErrNotFound
		}
		return nil, log, ErrInvalidOrderState
	}
	if err != nil {
		return nil, log, err
	}
	log.Payload = outboxEventAuditPayload(event)
	log.Payload["lease_seconds"] = leaseSeconds
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, []OutboxEvent{*event}, log); err != nil {
		return event, log, err
	}
	return event, cloneAuditLog(log), nil
}

func (s *PostgresStore) ReplayOutboxEvent(req ReplayOutboxEventRequest) (*OutboxEvent, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, ErrInvalidArgument
	}
	now := normalizeOutboxNow(req.Now)
	event, err := scanOutboxEvent(s.db.QueryRowContext(context.Background(), `UPDATE platform_outbox_events event
SET status = 'pending',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = $2,
    updated_at = $2
WHERE event.id = $1
  AND event.status IN ('pending', 'failed', 'dead_letter')
RETURNING `+postgresOutboxEventReturningColumns, eventID, now))
	if errors.Is(err, sql.ErrNoRows) {
		var status string
		statusErr := s.db.QueryRowContext(context.Background(), "SELECT status FROM platform_outbox_events WHERE id = $1", eventID).Scan(&status)
		if errors.Is(statusErr, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if statusErr != nil {
			return nil, statusErr
		}
		if status == OutboxStatusPublished {
			return nil, ErrInvalidOrderState
		}
		return nil, ErrInvalidArgument
	}
	if err != nil {
		return nil, err
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), []OutboxEvent{*event}); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *PostgresStore) ReplayOutboxEventWithAudit(req ReplayOutboxEventRequest, audit RecordAuditLogRequest) (*OutboxEvent, *AuditLog, error) {
	eventID := strings.TrimSpace(req.EventID)
	if eventID == "" {
		return nil, nil, ErrInvalidArgument
	}
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.replayed", "outbox_event", eventID)
	if err != nil {
		return nil, log, err
	}
	now := normalizeOutboxNow(req.Now)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	event, err := scanOutboxEvent(tx.QueryRowContext(ctx, `UPDATE platform_outbox_events event
SET status = 'pending',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = $2,
    updated_at = $2
WHERE event.id = $1
  AND event.status IN ('pending', 'failed', 'dead_letter')
RETURNING `+postgresOutboxEventReturningColumns, eventID, now))
	if errors.Is(err, sql.ErrNoRows) {
		var status string
		statusErr := tx.QueryRowContext(ctx, "SELECT status FROM platform_outbox_events WHERE id = $1", eventID).Scan(&status)
		if errors.Is(statusErr, sql.ErrNoRows) {
			return nil, log, ErrNotFound
		}
		if statusErr != nil {
			return nil, log, statusErr
		}
		if status == OutboxStatusPublished {
			return nil, log, ErrInvalidOrderState
		}
		return nil, log, ErrInvalidArgument
	}
	if err != nil {
		return nil, log, err
	}
	log.Payload = outboxEventAuditPayload(event)
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, []OutboxEvent{*event}, log); err != nil {
		return event, log, err
	}
	return event, cloneAuditLog(log), nil
}

func (s *PostgresStore) ReplayOutboxEvents(req ReplayOutboxEventsRequest) (*ReplayOutboxEventsResult, error) {
	topic := strings.TrimSpace(req.Topic)
	limit := normalizeOutboxLimit(req.Limit)
	now := normalizeOutboxNow(req.Now)
	rows, err := s.db.QueryContext(context.Background(), `WITH candidates AS (
  SELECT id
  FROM platform_outbox_events
  WHERE ($1 = '' OR topic = $1)
    AND status IN ('pending', 'failed')
    AND available_at > $2
    AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
  ORDER BY available_at, created_at, id
  LIMIT $3
  FOR UPDATE SKIP LOCKED
)
UPDATE platform_outbox_events event
SET status = 'pending',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = $2,
    updated_at = $2
FROM candidates
WHERE event.id = candidates.id
RETURNING `+postgresOutboxEventReturningColumns, topic, now, limit)
	if err != nil {
		return nil, err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return nil, err
	}
	sortOutboxEvents(events)
	result := &ReplayOutboxEventsResult{
		Topic:    topic,
		Limit:    limit,
		Replayed: len(events),
		Events:   events,
	}
	if err := s.applyOutboxEventsAndSaveSnapshot(context.Background(), events); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) ReplayOutboxEventsWithAudit(req ReplayOutboxEventsRequest, audit RecordAuditLogRequest) (*ReplayOutboxEventsResult, *AuditLog, error) {
	topic := strings.TrimSpace(req.Topic)
	log, err := outboxAuditLogFromRequest(audit, "admin.outbox.batch_replayed", "outbox_topic", outboxTopicAuditTarget(topic))
	if err != nil {
		return nil, log, err
	}
	limit := normalizeOutboxLimit(req.Limit)
	now := normalizeOutboxNow(req.Now)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	rows, err := tx.QueryContext(ctx, `WITH candidates AS (
  SELECT id
  FROM platform_outbox_events
  WHERE ($1 = '' OR topic = $1)
    AND status IN ('pending', 'failed')
    AND available_at > $2
    AND (lease_owner = '' OR lease_expires_at IS NULL OR lease_expires_at <= $2)
  ORDER BY available_at, created_at, id
  LIMIT $3
  FOR UPDATE SKIP LOCKED
)
UPDATE platform_outbox_events event
SET status = 'pending',
    last_error = '',
    lease_owner = '',
    lease_expires_at = NULL,
    available_at = $2,
    updated_at = $2
FROM candidates
WHERE event.id = candidates.id
RETURNING `+postgresOutboxEventReturningColumns, topic, now, limit)
	if err != nil {
		return nil, log, err
	}
	events, err := scanOutboxEventRows(rows)
	if err != nil {
		return nil, log, err
	}
	sortOutboxEvents(events)
	result := &ReplayOutboxEventsResult{
		Topic:    topic,
		Limit:    limit,
		Replayed: len(events),
		Events:   events,
	}
	log.Payload = outboxReplayBatchAuditPayload(result)
	if err := s.insertAuditLogInTx(ctx, tx, log); err != nil {
		return nil, log, err
	}
	if err := tx.Commit(); err != nil {
		return nil, log, errors.Join(ErrPersistence, err)
	}
	if err := s.applyOutboxEventsAndAuditAfterCommit(ctx, events, log); err != nil {
		return result, log, err
	}
	return result, cloneAuditLog(log), nil
}

func (s *PostgresStore) GrabOrder(orderID string, riderID string) (*Order, error) {
	order, err := s.Store.GrabOrder(orderID, riderID)
	return order, s.persistAfter(err)
}

func (s *PostgresStore) RiderMarkOrderPickedUp(orderID string, riderID string) (*Order, error) {
	order, err := s.Store.RiderMarkOrderPickedUp(orderID, riderID)
	return order, s.persistAfter(err)
}

func (s *PostgresStore) RiderMarkOrderDelivered(orderID string, riderID string) (*Order, error) {
	order, err := s.Store.RiderMarkOrderDelivered(orderID, riderID)
	return order, s.persistAfter(err)
}

func (s *PostgresStore) ConsumeFreeDispatchCancel(riderID string, at time.Time) (bool, string, error) {
	allowed, usedOn, err := s.Store.ConsumeFreeDispatchCancel(riderID, at)
	return allowed, usedOn, s.persistAfter(err)
}

var _ Repository = (*PostgresStore)(nil)
