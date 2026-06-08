-- Normalize user chat messages, realtime outbox payload recovery, and read receipts.

ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS sender_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS risk_state TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS risk_reason_code TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS risk_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS risk_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_messages_conversation_time
  ON messages (conversation_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_messages_sender_time
  ON messages (sender_type, sender_id, created_at DESC, id DESC);

CREATE TABLE conversation_read_states (
  user_id TEXT NOT NULL,
  conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  last_read_message_id TEXT NOT NULL DEFAULT '',
  read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, conversation_id)
);

CREATE INDEX IF NOT EXISTS idx_conversation_read_states_user
  ON conversation_read_states (user_id, updated_at DESC);
