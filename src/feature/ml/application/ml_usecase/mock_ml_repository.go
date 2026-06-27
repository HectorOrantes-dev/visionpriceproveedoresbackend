package ml_usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// MockMLRepository is a mock implementation of domain.MLRepository for testing.
type MockMLRepository struct {
	StandardCatalog []*entities.StandardProduct
	Products        []*productEntities.Product
	MarketPrices    map[string][]float64
}

// GetStandardCatalog retrieves the full master catalog for TF-IDF training.
func (m *MockMLRepository) GetStandardCatalog(ctx context.Context) ([]*entities.StandardProduct, error) {
	return m.StandardCatalog, nil
}

// GetProviderProducts retrieves all active products for a specific provider.
func (m *MockMLRepository) GetProviderProducts(ctx context.Context, providerID uuid.UUID) ([]*productEntities.Product, error) {
	var result []*productEntities.Product
	for _, p := range m.Products {
		if p.ProviderID == providerID {
			result = append(result, p)
		}
	}
	return result, nil
}

// GetCategoryPrices retrieves all prices for a specific category and unit.
func (m *MockMLRepository) GetCategoryPrices(ctx context.Context, category string, unit string) ([]float64, error) {
	key := fmt.Sprintf("%s|%s", category, unit)
	if prices, exists := m.MarketPrices[key]; exists {
		return prices, nil
	}
	return []float64{}, nil
}
