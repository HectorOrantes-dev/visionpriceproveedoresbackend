package domain

import (
	"context"

	"github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// RegisterRepository defines the port for provider registration persistence.
type RegisterRepository interface {
	// CreateProvider persists a new provider and returns it with the generated ID.
	CreateProvider(ctx context.Context, provider *entities.Provider) (*entities.Provider, error)

	// ExistsByEmail checks if a provider with the given email already exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByRFC checks if a provider with the given RFC already exists.
	ExistsByRFC(ctx context.Context, rfc string) (bool, error)
}
