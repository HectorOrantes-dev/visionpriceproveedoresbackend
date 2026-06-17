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

// UpsertLocation creates or updates the provider's location.
func (r *SupabaseGeolocationRepository) UpsertLocation(ctx context.Context, location *entities.ProviderLocation) (*entities.ProviderLocation, error) {
	query := `
		INSERT INTO provider_locations (provider_id, lat, lng, delivery_radius_km, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (provider_id)
		DO UPDATE SET lat = $2, lng = $3, delivery_radius_km = $4, updated_at = NOW()
		RETURNING id, provider_id, lat, lng, delivery_radius_km, updated_at
	`

	result := &entities.ProviderLocation{}
	err := r.db.QueryRow(ctx, query,
		location.ProviderID,
		location.Lat,
		location.Lng,
		location.DeliveryRadiusKm,
	).Scan(
		&result.ID,
		&result.ProviderID,
		&result.Lat,
		&result.Lng,
		&result.DeliveryRadiusKm,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert location: %w", err)
	}

	return result, nil
}

// GetLocation retrieves the provider's location.
func (r *SupabaseGeolocationRepository) GetLocation(ctx context.Context, providerID uuid.UUID) (*entities.ProviderLocation, error) {
	query := `
		SELECT id, provider_id, lat, lng, delivery_radius_km, updated_at
		FROM provider_locations
		WHERE provider_id = $1
	`

	location := &entities.ProviderLocation{}
	err := r.db.QueryRow(ctx, query, providerID).Scan(
		&location.ID,
		&location.ProviderID,
		&location.Lat,
		&location.Lng,
		&location.DeliveryRadiusKm,
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
