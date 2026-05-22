-- Reliable platform event outbox for future Kafka/NATS relays.
-- The current Store persists the same shape in snapshots; this table is the
-- normalized target for relay workers and domain-by-domain PostgreSQL storage.

CREATE TABLE platform_outbox_events (
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
);

CREATE INDEX idx_platform_outbox_ready
  ON platform_outbox_events (status, topic, available_at, created_at);

CREATE INDEX idx_platform_outbox_lease
  ON platform_outbox_events (lease_expires_at, lease_owner);

CREATE TABLE platform_consumed_events (
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
);

CREATE INDEX idx_platform_consumed_events_topic
  ON platform_consumed_events (consumer_name, topic, updated_at);
