-- Authentication sessions and payment hardening.

CREATE TABLE auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'merchant', 'rider', 'station_manager', 'admin')),
  subject_id TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  device_id TEXT NOT NULL DEFAULT '',
  ip_hash TEXT NOT NULL DEFAULT '',
  user_agent_hash TEXT NOT NULL DEFAULT '',
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_subject
  ON auth_sessions (subject_type, subject_id, expires_at DESC);

CREATE TABLE wallet_payment_passwords (
  user_id TEXT PRIMARY KEY REFERENCES app_users(id),
  password_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('unset', 'set', 'locked')),
  failed_count INTEGER NOT NULL DEFAULT 0,
  locked_until TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment_transactions (
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
);

CREATE INDEX idx_payment_transactions_order
  ON payment_transactions (order_id, created_at DESC);

CREATE TABLE refund_settings (
  id TEXT PRIMARY KEY DEFAULT 'default',
  default_strategy TEXT NOT NULL CHECK (default_strategy IN ('balance_first', 'original_route_first')),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refund_transactions (
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
);

CREATE INDEX idx_refund_transactions_order
  ON refund_transactions (order_id, created_at DESC);

CREATE TABLE order_after_sales (
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
);

CREATE INDEX idx_order_after_sales_order
  ON order_after_sales (order_id, created_at DESC);

CREATE INDEX idx_order_after_sales_user
  ON order_after_sales (user_id, created_at DESC);

CREATE TABLE order_after_sales_events (
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
);

CREATE INDEX idx_order_after_sales_events_request
  ON order_after_sales_events (request_id, created_at, id);

CREATE TABLE order_after_sales_evidence_upload_tickets (
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
);

CREATE INDEX idx_order_after_sales_upload_tickets_request
  ON order_after_sales_evidence_upload_tickets (request_id, created_at, id);

CREATE TABLE order_after_sales_evidence (
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
);

CREATE INDEX idx_order_after_sales_evidence_request
  ON order_after_sales_evidence (request_id, created_at, id);

CREATE TABLE payment_callbacks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider TEXT NOT NULL,
  out_trade_no TEXT NOT NULL,
  transaction_id TEXT NOT NULL,
  signature_valid BOOLEAN NOT NULL,
  idempotency_key TEXT NOT NULL UNIQUE,
  payload JSONB NOT NULL,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
