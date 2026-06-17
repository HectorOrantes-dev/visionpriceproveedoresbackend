package adapters

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// SupabaseRegisterRepository implements RegisterRepository using PostgreSQL (Supabase).
type SupabaseRegisterRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseRegisterRepository creates a new SupabaseRegisterRepository.
func NewSupabaseRegisterRepository(db *pgxpool.Pool) *SupabaseRegisterRepository {
	return &SupabaseRegisterRepository{db: db}
}

// CreateProvider inserts a new provider into the database.
func (r *SupabaseRegisterRepository) CreateProvider(ctx context.Context, provider *entities.Provider) (*entities.Provider, error) {
	query := `
		INSERT INTO providers (business_name, rfc, email, phone, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, business_name, rfc, email, phone, password_hash, created_at
	`

	created := &entities.Provider{}
	err := r.db.QueryRow(ctx, query,
		provider.BusinessName,
		provider.RFC,
		provider.Email,
		provider.Phone,
		provider.PasswordHash,
	).Scan(
		&created.ID,
		&created.BusinessName,
		&created.RFC,
		&created.Email,
		&created.Phone,
		&created.PasswordHash,
		&created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return created, nil
}

// ExistsByEmail checks if a provider with the given email exists.
func (r *SupabaseRegisterRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM providers WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

// ExistsByRFC checks if a provider with the given RFC exists.
func (r *SupabaseRegisterRepository) ExistsByRFC(ctx context.Context, rfc string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM providers WHERE rfc = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, rfc).Scan(&exists)
	return exists, err
}
