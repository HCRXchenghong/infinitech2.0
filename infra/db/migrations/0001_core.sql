-- Infinitech 2.0 core schema.
-- Amounts are stored as integer fen. High-risk state changes must write
-- append-only ledger/event rows in addition to updating current state columns.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE platform_sequences (
  name TEXT PRIMARY KEY,
  next_value BIGINT NOT NULL CHECK (next_value >= 1),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app_users (
  id TEXT PRIMARY KEY,
  phone_hash TEXT,
  nickname TEXT NOT NULL DEFAULT '',
  avatar_url TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  gender TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_identities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'merchant', 'rider', 'station_manager', 'admin')),
  subject_id TEXT NOT NULL,
  provider TEXT NOT NULL,
  provider_open_id TEXT NOT NULL,
  provider_union_id TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_open_id),
  UNIQUE (subject_type, subject_id, provider)
);

CREATE TABLE merchant_accounts (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK (type IN ('standard', 'pharmacy', 'clinic', 'platform_service')),
  display_name TEXT NOT NULL,
  registration_mode TEXT NOT NULL DEFAULT 'admin_invite_only',
  deposit_status TEXT NOT NULL DEFAULT 'unpaid',
  operation_status TEXT NOT NULL DEFAULT 'pending_review',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE onboarding_invites (
  token TEXT PRIMARY KEY,
  type TEXT NOT NULL CHECK (type IN ('merchant', 'rider', 'station_manager', 'old_user')),
  status TEXT NOT NULL CHECK (status IN ('active', 'used', 'expired', 'revoked')),
  created_by_subject_type TEXT NOT NULL,
  created_by_subject_id TEXT NOT NULL,
  target_subject_id TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shops (
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
);

CREATE TABLE merchant_qualifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  merchant_id TEXT NOT NULL REFERENCES merchant_accounts(id),
  shop_id TEXT REFERENCES shops(id),
  type TEXT NOT NULL CHECK (type IN ('business_license', 'health_certificate', 'supplemental_document')),
  file_url TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending_review',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_merchant_qualifications_expiry
  ON merchant_qualifications (merchant_id, shop_id, type, expires_at);

CREATE TABLE merchant_staff (
  id TEXT PRIMARY KEY,
  merchant_id TEXT NOT NULL REFERENCES merchant_accounts(id),
  shop_id TEXT REFERENCES shops(id),
  name TEXT NOT NULL,
  phone_hash TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  health_certificate_url TEXT NOT NULL DEFAULT '',
  health_certificate_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE merchant_products (
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
);

CREATE INDEX idx_merchant_products_shop_status
  ON merchant_products (shop_id, status);

CREATE TABLE user_addresses (
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
);

CREATE INDEX idx_user_addresses_user_default
  ON user_addresses (user_id, is_default DESC, updated_at DESC);

CREATE TABLE wallet_accounts (
  subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'merchant', 'rider', 'platform')),
  subject_id TEXT NOT NULL,
  balance_fen BIGINT NOT NULL DEFAULT 0 CHECK (balance_fen >= 0),
  frozen_fen BIGINT NOT NULL DEFAULT 0 CHECK (frozen_fen >= 0),
  version BIGINT NOT NULL DEFAULT 0,
  risk_state TEXT NOT NULL DEFAULT 'normal',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (subject_type, subject_id)
);

CREATE TABLE wallet_transactions (
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
);

CREATE INDEX idx_wallet_transactions_subject_time
  ON wallet_transactions (subject_type, subject_id, created_at DESC);

CREATE TABLE rider_accounts (
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
);

CREATE INDEX idx_rider_accounts_dispatch
  ON rider_accounts (station_id, online, status, deposit_status);

CREATE TABLE deposit_accounts (
  subject_type TEXT NOT NULL CHECK (subject_type IN ('rider', 'merchant')),
  subject_id TEXT NOT NULL,
  amount_fen BIGINT NOT NULL DEFAULT 5000 CHECK (amount_fen >= 0),
  status TEXT NOT NULL DEFAULT 'unpaid',
  last_order_completed_at TIMESTAMPTZ,
  resignation_submitted_at TIMESTAMPTZ,
  dispute_closed_at TIMESTAMPTZ,
  wechat_exempt_application_id TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (subject_type, subject_id)
);

CREATE TABLE cart_items (
  user_id TEXT NOT NULL REFERENCES app_users(id),
  shop_id TEXT NOT NULL REFERENCES shops(id),
  product_id TEXT NOT NULL REFERENCES merchant_products(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  selected BOOLEAN NOT NULL DEFAULT true,
  unit_price_fen BIGINT NOT NULL CHECK (unit_price_fen >= 0),
  product_name_snapshot TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, shop_id, product_id)
);

CREATE TABLE orders (
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
);

CREATE INDEX idx_orders_user_time ON orders (user_id, created_at DESC);
CREATE INDEX idx_orders_shop_status ON orders (shop_id, status, created_at DESC);
CREATE INDEX idx_orders_dispatch_status ON orders (status, dispatch_mode, created_at);

CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT NOT NULL,
  product_name_snapshot TEXT NOT NULL,
  unit_price_fen BIGINT NOT NULL CHECK (unit_price_fen >= 0),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE order_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  actor_type TEXT NOT NULL DEFAULT 'system',
  actor_id TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_events_order_time ON order_events (order_id, created_at);

CREATE TABLE dispatch_jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id TEXT NOT NULL REFERENCES orders(id),
  mode TEXT NOT NULL CHECK (mode IN ('grab_hall', 'auto_assign', 'manual_assign')),
  status TEXT NOT NULL DEFAULT 'pending',
  candidate_rider_id TEXT REFERENCES rider_accounts(id),
  expires_at TIMESTAMPTZ,
  idempotency_key TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (idempotency_key)
);

CREATE UNIQUE INDEX uniq_active_dispatch_job_per_order
  ON dispatch_jobs (order_id)
  WHERE status IN ('pending', 'offered');

CREATE TABLE dispatch_events (
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
);

CREATE INDEX idx_dispatch_events_order_time
  ON dispatch_events (order_id, created_at);

CREATE INDEX idx_dispatch_events_station_time
  ON dispatch_events (station_id, created_at DESC);

CREATE UNIQUE INDEX uniq_dispatch_events_idempotency_type
  ON dispatch_events (idempotency_key, type);

CREATE TABLE rider_free_cancel_usage (
  rider_id TEXT NOT NULL REFERENCES rider_accounts(id),
  business_date DATE NOT NULL,
  order_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (rider_id, business_date)
);

CREATE TABLE conversations (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  order_id TEXT REFERENCES orders(id),
  title TEXT NOT NULL DEFAULT '',
  notification_default TEXT NOT NULL DEFAULT 'normal',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE conversation_members (
  conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  muted BOOLEAN NOT NULL DEFAULT false,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (conversation_id, subject_type, subject_id)
);

CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_type TEXT NOT NULL,
  sender_id TEXT NOT NULL,
  type TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  attachment_urls TEXT[] NOT NULL DEFAULT '{}',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_conversation_time
  ON messages (conversation_id, created_at DESC);

CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_type TEXT NOT NULL,
  actor_id TEXT NOT NULL,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  request_id TEXT NOT NULL DEFAULT '',
  ip_hash TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_target_time
  ON audit_logs (target_type, target_id, created_at DESC);
