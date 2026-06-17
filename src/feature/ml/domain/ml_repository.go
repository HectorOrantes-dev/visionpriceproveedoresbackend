package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// MLRepository defines the port for ML-related persistence operations.
type MLRepository interface {
	// GetStandardCatalog retrieves the full master catalog for TF-IDF training.
	GetStandardCatalog(ctx context.Context) ([]*entities.StandardProduct, error)

	// GetProviderProducts retrieves all active products for a specific provider.
	GetProviderProducts(ctx context.Context, providerID uuid.UUID) ([]*productEntities.Product, error)

	// GetCategoryPrices retrieves all prices for a specific category and unit across the entire system.
	// Used for training the Isolation Forest model on real market distributions.
	GetCategoryPrices(ctx context.Context, category string, unit string) ([]float64, error)
}
