package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SupabaseTwoFactorRepository implements TwoFactorRepository using PostgreSQL (Supabase).
type SupabaseTwoFactorRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseTwoFactorRepository creates a new SupabaseTwoFactorRepository.
func NewSupabaseTwoFactorRepository(db *pgxpool.Pool) *SupabaseTwoFactorRepository {
	return &SupabaseTwoFactorRepository{db: db}
}

// CreateOTP stores a new OTP code for a provider.
func (r *SupabaseTwoFactorRepository) CreateOTP(ctx context.Context, providerID uuid.UUID, code string, expirationMinutes int) error {
	// Invalidate any existing unused OTP for this provider first
	invalidateQuery := `
		UPDATE two_factor_codes SET used = TRUE
		WHERE provider_id = $1 AND used = FALSE
	`
	_, _ = r.db.Exec(ctx, invalidateQuery, providerID)

	query := `
		INSERT INTO two_factor_codes (provider_id, code, expires_at)
		VALUES ($1, $2, $3)
	`
	expiresAt := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)
	_, err := r.db.Exec(ctx, query, providerID, code, expiresAt)
	return err
}

// ValidateOTP checks if the code is valid and marks it as used atomically.
func (r *SupabaseTwoFactorRepository) ValidateOTP(ctx context.Context, providerID uuid.UUID, code string) (bool, error) {
	query := `
		UPDATE two_factor_codes
		SET used = TRUE
		WHERE provider_id = $1 AND code = $2 AND used = FALSE AND expires_at > NOW()
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, providerID, code).Scan(&id)
	if err != nil {
		return false, fmt.Errorf("OTP validation failed: %w", err)
	}
	return true, nil
}

// InvalidateExpired marks all expired OTP codes as used.
func (r *SupabaseTwoFactorRepository) InvalidateExpired(ctx context.Context) error {
	query := `UPDATE two_factor_codes SET used = TRUE WHERE expires_at <= NOW() AND used = FALSE`
	_, err := r.db.Exec(ctx, query)
	return err
}

// GetProviderContact returns the provider's email and business name for OTP delivery.
func (r *SupabaseTwoFactorRepository) GetProviderContact(ctx context.Context, providerID uuid.UUID) (string, string, error) {
	query := `SELECT email, business_name FROM providers WHERE id = $1 AND active = TRUE`
	var email, name string
	if err := r.db.QueryRow(ctx, query, providerID).Scan(&email, &name); err != nil {
		return "", "", fmt.Errorf("provider contact lookup failed: %w", err)
	}
	return email, name, nil
}
