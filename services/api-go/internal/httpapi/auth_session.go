package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type AuthSession struct {
	ID            string
	Principal     Principal
	TokenHash     string
	DeviceID      string
	IPHash        string
	UserAgentHash string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	CreatedAt     time.Time
}

type AuthSessionStore interface {
	Save(ctx context.Context, session AuthSession) error
	Verify(ctx context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error
	Revoke(ctx context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error
}

type MemoryAuthSessionStore struct {
	mu       sync.Mutex
	sessions map[string]AuthSession
}

func NewMemoryAuthSessionStore() *MemoryAuthSessionStore {
	return &MemoryAuthSessionStore{sessions: map[string]AuthSession{}}
}

func (store *MemoryAuthSessionStore) Save(_ context.Context, session AuthSession) error {
	if store == nil || session.ID == "" || session.Principal.IsZero() || session.TokenHash == "" || session.ExpiresAt.IsZero() {
		return errUnauthorized
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sessions[session.ID] = session
	return nil
}

func (store *MemoryAuthSessionStore) Verify(_ context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error {
	if store == nil || sessionID == "" || tokenHash == "" || principal.IsZero() {
		return errUnauthorized
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	session, ok := store.sessions[sessionID]
	if !ok || session.TokenHash != tokenHash || session.Principal != principal {
		return errUnauthorized
	}
	if session.RevokedAt != nil || !session.ExpiresAt.After(now.UTC()) {
		return errUnauthorized
	}
	return nil
}

func (store *MemoryAuthSessionStore) Revoke(_ context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error {
	if store == nil || sessionID == "" || tokenHash == "" || principal.IsZero() {
		return errUnauthorized
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	session, ok := store.sessions[sessionID]
	if !ok || session.TokenHash != tokenHash || session.Principal != principal {
		return errUnauthorized
	}
	revokedAt := now.UTC()
	session.RevokedAt = &revokedAt
	store.sessions[sessionID] = session
	return nil
}

type PostgresAuthSessionStore struct {
	db *sql.DB
}

func NewPostgresAuthSessionStore(ctx context.Context, databaseURL string) (*PostgresAuthSessionStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &PostgresAuthSessionStore{db: db}
	if err := store.ensureTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (store *PostgresAuthSessionStore) Close() error {
	if store == nil || store.db == nil {
		return nil
	}
	return store.db.Close()
}

func (store *PostgresAuthSessionStore) ensureTable(ctx context.Context) error {
	_, err := store.db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	if err != nil {
		return err
	}
	_, err = store.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_type TEXT NOT NULL CHECK (subject_type IN ('user', 'merchant', 'rider', 'station_manager', 'admin', 'super_admin', 'ops_admin', 'finance_admin', 'dispatch_admin', 'support_admin', 'security_auditor')),
  subject_id TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  device_id TEXT NOT NULL DEFAULT '',
  ip_hash TEXT NOT NULL DEFAULT '',
  user_agent_hash TEXT NOT NULL DEFAULT '',
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	if err != nil {
		return err
	}
	_, err = store.db.ExecContext(ctx, `
ALTER TABLE auth_sessions DROP CONSTRAINT IF EXISTS auth_sessions_subject_type_check;
ALTER TABLE auth_sessions ADD CONSTRAINT auth_sessions_subject_type_check
  CHECK (subject_type IN ('user', 'merchant', 'rider', 'station_manager', 'admin', 'super_admin', 'ops_admin', 'finance_admin', 'dispatch_admin', 'support_admin', 'security_auditor'))`)
	if err != nil {
		return err
	}
	_, err = store.db.ExecContext(ctx, `
CREATE INDEX IF NOT EXISTS idx_auth_sessions_subject
  ON auth_sessions (subject_type, subject_id, expires_at DESC)`)
	return err
}

func (store *PostgresAuthSessionStore) Save(ctx context.Context, session AuthSession) error {
	if store == nil || session.ID == "" || session.Principal.IsZero() || session.TokenHash == "" || session.ExpiresAt.IsZero() {
		return errUnauthorized
	}
	_, err := store.db.ExecContext(ctx, `
INSERT INTO auth_sessions (id, subject_type, subject_id, token_hash, device_id, ip_hash, user_agent_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())`,
		session.ID,
		session.Principal.Role,
		session.Principal.ID,
		session.TokenHash,
		session.DeviceID,
		session.IPHash,
		session.UserAgentHash,
		session.ExpiresAt.UTC(),
	)
	if err != nil {
		return errUnauthorized
	}
	return nil
}

func (store *PostgresAuthSessionStore) Verify(ctx context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error {
	if store == nil || sessionID == "" || tokenHash == "" || principal.IsZero() {
		return errUnauthorized
	}
	var storedTokenHash string
	var expiresAt time.Time
	var revokedAt sql.NullTime
	err := store.db.QueryRowContext(ctx, `
SELECT token_hash, expires_at, revoked_at
FROM auth_sessions
WHERE id = $1 AND subject_type = $2 AND subject_id = $3`,
		sessionID,
		principal.Role,
		principal.ID,
	).Scan(&storedTokenHash, &expiresAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) || err != nil {
		return errUnauthorized
	}
	if storedTokenHash != tokenHash || revokedAt.Valid || !expiresAt.After(now.UTC()) {
		return errUnauthorized
	}
	return nil
}

func (store *PostgresAuthSessionStore) Revoke(ctx context.Context, sessionID string, tokenHash string, principal Principal, now time.Time) error {
	if store == nil || sessionID == "" || tokenHash == "" || principal.IsZero() {
		return errUnauthorized
	}
	result, err := store.db.ExecContext(ctx, `
UPDATE auth_sessions
SET revoked_at = $1
WHERE id = $2 AND token_hash = $3 AND subject_type = $4 AND subject_id = $5 AND revoked_at IS NULL`,
		now.UTC(),
		sessionID,
		tokenHash,
		principal.Role,
		principal.ID,
	)
	if err != nil {
		return errUnauthorized
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return errUnauthorized
	}
	return nil
}

func newAuthSession(req *http.Request, sessionID string, principal Principal, token string, expiresAt time.Time) AuthSession {
	return AuthSession{
		ID:            strings.TrimSpace(sessionID),
		Principal:     principal,
		TokenHash:     tokenHash(token),
		DeviceID:      strings.TrimSpace(req.Header.Get("X-Device-Id")),
		IPHash:        requestIPHash(req),
		UserAgentHash: shortValueHash(req.UserAgent()),
		ExpiresAt:     expiresAt.UTC(),
		CreatedAt:     time.Now().UTC(),
	}
}

func newSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes[0:4]) + "-" +
		hex.EncodeToString(bytes[4:6]) + "-" +
		hex.EncodeToString(bytes[6:8]) + "-" +
		hex.EncodeToString(bytes[8:10]) + "-" +
		hex.EncodeToString(bytes[10:16]), nil
}

func tokenHash(token string) string {
	return shortValueHash(token)
}

func requestIPHash(req *http.Request) string {
	ip := strings.TrimSpace(req.Header.Get("X-Forwarded-For"))
	if comma := strings.Index(ip, ","); comma >= 0 {
		ip = strings.TrimSpace(ip[:comma])
	}
	if ip == "" {
		ip = strings.TrimSpace(req.Header.Get("X-Real-IP"))
	}
	if ip == "" {
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			ip = host
		}
	}
	return shortValueHash(ip)
}

func shortValueHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}
