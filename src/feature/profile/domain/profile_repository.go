package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/profile/domain/entities"
)

// ProfileRepository defines the port for reading and updating provider profile data.
type ProfileRepository interface {
	// GetByID returns the active provider's profile, or an error if not found.
	GetByID(ctx context.Context, providerID uuid.UUID) (*entities.Profile, error)
	// Update applies partial updates to the provider's profile fields.
	Update(ctx context.Context, providerID uuid.UUID, req *entities.UpdateProfileRequest) (*entities.Profile, error)
	// EmailExists checks whether a different provider already owns the given email.
	EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error)
}
