// Package csrf implements a per-session, server-issued CSRF token.
//
// The API authenticates with a Bearer JWT in the Authorization header, which is
// already immune to classic CSRF (a cross-site page cannot set that header).
// This package adds an explicit, defense-in-depth synchronizer token on top:
//
//   - On successful 2FA verification the backend mints a random CSRF token,
//     stores only its SHA-256 hash bound to the provider (session), and returns
//     the raw token to the client.
//   - For every state-changing request the client must echo the token in the
//     X-CSRF-Token header. The middleware re-hashes it and checks it belongs to
//     the authenticated provider and has not expired.
//   - Tokens are revoked on logout and expire with the session.
package csrf

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HeaderName is the HTTP header clients use to send the CSRF token.
const HeaderName = "X-CSRF-Token"

// Manager issues and validates CSRF tokens backed by the database.
type Manager struct {
	db  *pgxpool.Pool
	ttl time.Duration
}

// NewManager creates a CSRF Manager. ttl controls how long a token stays valid;
// it should match (or exceed) the session/refresh-token lifetime.
func NewManager(db *pgxpool.Pool, ttl time.Duration) *Manager {
	return &Manager{db: db, ttl: ttl}
}

// Issue generates a new CSRF token for the given provider, stores its hash, and
// returns the raw token to hand back to the client.
func (m *Manager) Issue(ctx context.Context, providerID string) (string, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return "", err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawToken := hex.EncodeToString(raw)
	tokenHash := hashToken(rawToken)

	const q = `
		INSERT INTO csrf_tokens (provider_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := m.db.Exec(ctx, q, pid, tokenHash, time.Now().Add(m.ttl)); err != nil {
		return "", err
	}

	return rawToken, nil
}

// Validate reports whether rawToken is a live CSRF token belonging to providerID.
func (m *Manager) Validate(ctx context.Context, providerID, rawToken string) (bool, error) {
	if rawToken == "" {
		return false, nil
	}
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return false, nil
	}

	const q = `
		SELECT EXISTS (
			SELECT 1 FROM csrf_tokens
			WHERE token_hash = $1 AND provider_id = $2 AND expires_at > NOW()
		)
	`
	var ok bool
	if err := m.db.QueryRow(ctx, q, hashToken(rawToken), pid).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

// RevokeAll removes all CSRF tokens for a provider (used on logout).
func (m *Manager) RevokeAll(ctx context.Context, providerID string) error {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(ctx, `DELETE FROM csrf_tokens WHERE provider_id = $1`, pid)
	return err
}

// CleanupExpired removes expired CSRF tokens. Call periodically from a goroutine.
func (m *Manager) CleanupExpired(ctx context.Context) error {
	_, err := m.db.Exec(ctx, `DELETE FROM csrf_tokens WHERE expires_at < NOW()`)
	return err
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
