-- Transitional PostgreSQL persistence for the current modular Store.
-- This gives api-go restart-safe state while high-risk domains are moved to
-- normalized ledger/event tables behind the same Repository interface.

CREATE TABLE platform_store_snapshots (
  store_key TEXT PRIMARY KEY,
  payload JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

