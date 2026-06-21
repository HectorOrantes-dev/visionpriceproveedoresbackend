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
	// remaining caps how many NEW products may be inserted (updates never count
	// against the plan); a negative value means unlimited. Returns the count of
	// new, updated, and products skipped because the plan limit was reached.
	BulkUpsertProducts(ctx context.Context, providerID uuid.UUID, products []*productEntities.Product, remaining int) (newCount int, updatedCount int, skippedByLimit int, err error)

	// CountActiveProducts returns how many active products the provider has.
	CountActiveProducts(ctx context.Context, providerID uuid.UUID) (int, error)
}

// PlanLimitService is the narrow port (consumer-defined) used to learn the
// provider's product cap. Implemented by the subscriptions use case.
type PlanLimitService interface {
	ProductLimit(ctx context.Context, providerID string) (limit int, unlimited bool, err error)
}
