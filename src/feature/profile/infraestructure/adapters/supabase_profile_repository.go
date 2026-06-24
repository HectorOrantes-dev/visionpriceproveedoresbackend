package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/profile/domain/entities"
)

// SupabaseProfileRepository implements ProfileRepository using PostgreSQL (Supabase).
type SupabaseProfileRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseProfileRepository creates a new SupabaseProfileRepository.
func NewSupabaseProfileRepository(db *pgxpool.Pool) *SupabaseProfileRepository {
	return &SupabaseProfileRepository{db: db}
}

// GetByID returns the active provider's profile by id.
func (r *SupabaseProfileRepository) GetByID(ctx context.Context, providerID uuid.UUID) (*entities.Profile, error) {
	query := `
		SELECT id, business_name, rfc, email, phone, created_at
		FROM providers
		WHERE id = $1 AND active = TRUE
	`
	var p entities.Profile
	err := r.db.QueryRow(ctx, query, providerID).Scan(
		&p.ID, &p.BusinessName, &p.RFC, &p.Email, &p.Phone, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("profile lookup failed: %w", err)
	}
	return &p, nil
}
