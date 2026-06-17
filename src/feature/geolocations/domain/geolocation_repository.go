package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain/entities"
)

// GeolocationRepository defines the port for geolocation persistence operations.
type GeolocationRepository interface {
	// UpsertLocation creates or updates the provider's location.
	UpsertLocation(ctx context.Context, location *entities.ProviderLocation) (*entities.ProviderLocation, error)

	// GetLocation retrieves the provider's location.
	GetLocation(ctx context.Context, providerID uuid.UUID) (*entities.ProviderLocation, error)

	// DeleteLocation removes the provider's location.
	DeleteLocation(ctx context.Context, providerID uuid.UUID) error
}
