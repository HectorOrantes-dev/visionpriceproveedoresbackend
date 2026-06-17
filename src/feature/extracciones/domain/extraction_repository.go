package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/extracciones/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// ExtractionRepository defines the port for extraction-related persistence operations.
type ExtractionRepository interface {
	// GetMapping retrieves the saved column mapping for a provider.
	GetMapping(ctx context.Context, providerID uuid.UUID) (*entities.ImportMapping, error)

	// SaveMapping creates or updates the column mapping for a provider.
	SaveMapping(ctx context.Context, mapping *entities.ImportMapping) (*entities.ImportMapping, error)

	// BulkUpsertProducts inserts new products or updates existing ones by name for a provider.
	// Returns the count of new and updated products.
	BulkUpsertProducts(ctx context.Context, providerID uuid.UUID, products []*productEntities.Product) (newCount int, updatedCount int, err error)
}
