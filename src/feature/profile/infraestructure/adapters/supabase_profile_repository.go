package adapters

import (
	"context"
	"fmt"
	"strings"

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

// Update applies partial updates and returns the refreshed profile.
func (r *SupabaseProfileRepository) Update(ctx context.Context, providerID uuid.UUID, req *entities.UpdateProfileRequest) (*entities.Profile, error) {
	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 4)
	idx := 1

	if req.BusinessName != nil {
		setClauses = append(setClauses, fmt.Sprintf("business_name = $%d", idx))
		args = append(args, *req.BusinessName)
		idx++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", idx))
		args = append(args, strings.ToLower(*req.Email))
		idx++
	}
	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", idx))
		args = append(args, *req.Phone)
		idx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, providerID)
	}

	args = append(args, providerID)
	query := fmt.Sprintf(`
		UPDATE providers SET %s
		WHERE id = $%d AND active = TRUE
		RETURNING id, business_name, rfc, email, phone, created_at
	`, strings.Join(setClauses, ", "), idx)

	var p entities.Profile
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.BusinessName, &p.RFC, &p.Email, &p.Phone, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("profile update failed: %w", err)
	}
	return &p, nil
}

// EmailExists checks whether a different provider already owns the given email.
func (r *SupabaseProfileRepository) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM providers WHERE LOWER(email) = LOWER($1) AND id != $2 AND active = TRUE)`,
		email, excludeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("email existence check failed: %w", err)
	}
	return exists, nil
}
