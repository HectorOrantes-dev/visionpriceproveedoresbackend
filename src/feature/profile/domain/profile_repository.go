package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/profile/domain/entities"
)

// ProfileRepository defines the port for reading provider profile data.
type ProfileRepository interface {
	// GetByID returns the active provider's profile, or an error if not found.
	GetByID(ctx context.Context, providerID uuid.UUID) (*entities.Profile, error)
}
