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

	// UpdateImageURL sets only the image_url of a product, scoped to its owner.
	// Used by the background image upload so it does not overwrite other fields.
	UpdateImageURL(ctx context.Context, providerID, productID uuid.UUID, imageURL string) error

	// SoftDelete marks a product as inactive (logical deletion).
	SoftDelete(ctx context.Context, productID uuid.UUID) error

	// CountActiveByProvider returns how many active products a provider has.
	CountActiveByProvider(ctx context.Context, providerID uuid.UUID) (int, error)

	// MetricsAggregate returns catalog aggregates for a provider: total inventory
	// value (Σ price×stock), units in stock (Σ stock), average price and material count.
	MetricsAggregate(ctx context.Context, providerID uuid.UUID) (inventoryValue float64, unitsInStock int, averagePrice float64, totalMaterials int, err error)

	// CategoryDistribution returns the inventory value grouped by category
	// (descending). The Share field is left at 0 for the caller to compute.
	CategoryDistribution(ctx context.Context, providerID uuid.UUID) ([]entities.CategorySlice, error)

	// MonthlyNewMaterials returns how many materials were created per month for the
	// last 6 months (zero-filled), with a Spanish month label.
	MonthlyNewMaterials(ctx context.Context, providerID uuid.UUID) ([]entities.MonthlyPoint, error)

	// TopByInventoryValue returns the top materials by inventory value (price×stock).
	// The Share field is left at 0 for the caller to compute.
	TopByInventoryValue(ctx context.Context, providerID uuid.UUID, limit int) ([]entities.TopProduct, error)
}

// PlanLimitService is the narrow port (consumer-defined) the products feature
// uses to learn the provider's product cap. Implemented by the subscriptions
// use case. unlimited=true means there is no cap.
type PlanLimitService interface {
	ProductLimit(ctx context.Context, providerID string) (limit int, unlimited bool, err error)
}
