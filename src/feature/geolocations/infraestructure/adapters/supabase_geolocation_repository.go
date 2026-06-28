package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain/entities"
)

// SupabaseGeolocationRepository implements GeolocationRepository using PostgreSQL (Supabase).
type SupabaseGeolocationRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseGeolocationRepository creates a new SupabaseGeolocationRepository.
func NewSupabaseGeolocationRepository(db *pgxpool.Pool) *SupabaseGeolocationRepository {
	return &SupabaseGeolocationRepository{db: db}
}

// UpsertLocation creates or updates the provider's warehouse address.
func (r *SupabaseGeolocationRepository) UpsertLocation(ctx context.Context, location *entities.ProviderLocation) (*entities.ProviderLocation, error) {
	query := `
		INSERT INTO provider_locations (provider_id, address, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (provider_id)
		DO UPDATE SET address = $2, updated_at = NOW()
		RETURNING id, provider_id, address, updated_at
	`

	result := &entities.ProviderLocation{}
	err := r.db.QueryRow(ctx, query,
		location.ProviderID,
		location.Address,
	).Scan(
		&result.ID,
		&result.ProviderID,
		&result.Address,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert location: %w", err)
	}

	return result, nil
}

// GetLocation retrieves the provider's warehouse address.
func (r *SupabaseGeolocationRepository) GetLocation(ctx context.Context, providerID uuid.UUID) (*entities.ProviderLocation, error) {
	query := `
		SELECT id, provider_id, address, updated_at
		FROM provider_locations
		WHERE provider_id = $1
	`

	location := &entities.ProviderLocation{}
	err := r.db.QueryRow(ctx, query, providerID).Scan(
		&location.ID,
		&location.ProviderID,
		&location.Address,
		&location.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	return location, nil
}

// DeleteLocation removes the provider's location.
func (r *SupabaseGeolocationRepository) DeleteLocation(ctx context.Context, providerID uuid.UUID) error {
	query := `DELETE FROM provider_locations WHERE provider_id = $1`
	tag, err := r.db.Exec(ctx, query, providerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("location not found")
	}
	return nil
}
