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
}
