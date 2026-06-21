package domain

import (
	"context"

	"github.com/google/uuid"

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

// DefaultSubscriptionCreator is the narrow port (consumer-defined) used to assign
// the default free subscription to a newly registered provider. Implemented by
// the subscriptions use case.
type DefaultSubscriptionCreator interface {
	EnsureDefault(ctx context.Context, providerID uuid.UUID) error
}
