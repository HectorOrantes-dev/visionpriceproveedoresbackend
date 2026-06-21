package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// ProductRepository defines the port for product persistence operations.
type ProductRepository interface {
	// Create inserts a new product.
	Create(ctx context.Context, product *entities.Product) (*entities.Product, error)

	// FindByID retrieves a product by ID (only active).
	FindByID(ctx context.Context, productID uuid.UUID) (*entities.Product, error)

	// FindAllActive retrieves all active products for a provider.
	FindAllActive(ctx context.Context, providerID uuid.UUID) ([]*entities.Product, error)

	// Update modifies an existing product.
	Update(ctx context.Context, product *entities.Product) (*entities.Product, error)

	// SoftDelete marks a product as inactive (logical deletion).
	SoftDelete(ctx context.Context, productID uuid.UUID) error

	// CountActiveByProvider returns how many active products a provider has.
	CountActiveByProvider(ctx context.Context, providerID uuid.UUID) (int, error)
}

// PlanLimitService is the narrow port (consumer-defined) the products feature
// uses to learn the provider's product cap. Implemented by the subscriptions
// use case. unlimited=true means there is no cap.
type PlanLimitService interface {
	ProductLimit(ctx context.Context, providerID string) (limit int, unlimited bool, err error)
}
