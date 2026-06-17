package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	registerEntities "github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// SupabaseLoginRepository implements LoginRepository using PostgreSQL (Supabase).
type SupabaseLoginRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseLoginRepository creates a new SupabaseLoginRepository.
func NewSupabaseLoginRepository(db *pgxpool.Pool) *SupabaseLoginRepository {
	return &SupabaseLoginRepository{db: db}
}

// FindByEmail retrieves a provider by email.
func (r *SupabaseLoginRepository) FindByEmail(ctx context.Context, email string) (*registerEntities.Provider, error) {
	query := `
		SELECT id, business_name, rfc, email, phone, password_hash, created_at
		FROM providers
		WHERE email = $1
	`
	provider := &registerEntities.Provider{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&provider.ID,
		&provider.BusinessName,
		&provider.RFC,
		&provider.Email,
		&provider.Phone,
		&provider.PasswordHash,
		&provider.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return provider, nil
}

// UpdatePassword updates the password hash for a provider.
func (r *SupabaseLoginRepository) UpdatePassword(ctx context.Context, providerID uuid.UUID, newPasswordHash string) error {
	query := `UPDATE providers SET password_hash = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, newPasswordHash, providerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("provider not found")
	}
	return nil
}

// CreateResetToken stores a password reset token.
func (r *SupabaseLoginRepository) CreateResetToken(ctx context.Context, providerID uuid.UUID, tokenHash string, expiresAtMinutes int) error {
	query := `
		INSERT INTO password_reset_tokens (provider_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	expiresAt := time.Now().Add(time.Duration(expiresAtMinutes) * time.Minute)
	_, err := r.db.Exec(ctx, query, providerID, tokenHash, expiresAt)
	return err
}

// FindValidResetToken finds a valid (unused, not expired) reset token by its hash.
func (r *SupabaseLoginRepository) FindValidResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	query := `
		SELECT provider_id FROM password_reset_tokens
		WHERE token_hash = $1 AND used = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	var providerID uuid.UUID
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&providerID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("reset token not found or expired: %w", err)
	}
	return providerID, nil
}

// MarkResetTokenUsed marks a reset token as used.
func (r *SupabaseLoginRepository) MarkResetTokenUsed(ctx context.Context, tokenHash string) error {
	query := `UPDATE password_reset_tokens SET used = TRUE WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

// RevokeToken adds a token hash to the revocation blacklist.
func (r *SupabaseLoginRepository) RevokeToken(ctx context.Context, tokenHash string, providerID uuid.UUID, expiresAt time.Time) error {
	query := `
		INSERT INTO revoked_tokens (token_hash, provider_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_hash) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, tokenHash, providerID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// IsTokenRevoked checks if a token hash exists in the revocation blacklist.
func (r *SupabaseLoginRepository) IsTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE token_hash = $1)`
	var revoked bool
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&revoked)
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return revoked, nil
}

