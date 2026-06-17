package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/extracciones/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// SupabaseExtractionRepository implements ExtractionRepository using PostgreSQL (Supabase).
type SupabaseExtractionRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseExtractionRepository creates a new SupabaseExtractionRepository.
func NewSupabaseExtractionRepository(db *pgxpool.Pool) *SupabaseExtractionRepository {
	return &SupabaseExtractionRepository{db: db}
}

// GetMapping retrieves the saved column mapping for a provider.
func (r *SupabaseExtractionRepository) GetMapping(ctx context.Context, providerID uuid.UUID) (*entities.ImportMapping, error) {
	query := `
		SELECT id, provider_id, column_map, created_at, updated_at
		FROM import_mappings
		WHERE provider_id = $1
	`

	mapping := &entities.ImportMapping{}
	var columnMapBytes []byte

	err := r.db.QueryRow(ctx, query, providerID).Scan(
		&mapping.ID,
		&mapping.ProviderID,
		&columnMapBytes,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("mapping not found: %w", err)
	}

	if err := json.Unmarshal(columnMapBytes, &mapping.ColumnMap); err != nil {
		return nil, fmt.Errorf("failed to parse column map: %w", err)
	}

	return mapping, nil
}

// SaveMapping creates or updates the column mapping for a provider.
func (r *SupabaseExtractionRepository) SaveMapping(ctx context.Context, mapping *entities.ImportMapping) (*entities.ImportMapping, error) {
	columnMapJSON, err := json.Marshal(mapping.ColumnMap)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize column map: %w", err)
	}

	query := `
		INSERT INTO import_mappings (provider_id, column_map, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (provider_id)
		DO UPDATE SET column_map = $2, updated_at = NOW()
		RETURNING id, provider_id, column_map, created_at, updated_at
	`

	result := &entities.ImportMapping{}
	var resultColumnMapBytes []byte

	err = r.db.QueryRow(ctx, query, mapping.ProviderID, columnMapJSON).Scan(
		&result.ID,
		&result.ProviderID,
		&resultColumnMapBytes,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save mapping: %w", err)
	}

	if err := json.Unmarshal(resultColumnMapBytes, &result.ColumnMap); err != nil {
		return nil, fmt.Errorf("failed to parse saved column map: %w", err)
	}

	return result, nil
}

// BulkUpsertProducts inserts or updates products by name for a provider.
func (r *SupabaseExtractionRepository) BulkUpsertProducts(ctx context.Context, providerID uuid.UUID, products []*productEntities.Product) (int, int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	newCount := 0
	updatedCount := 0

	for _, product := range products {
		// Check if product with same name exists for this provider
		var existingID uuid.UUID
		checkQuery := `SELECT id FROM products WHERE provider_id = $1 AND LOWER(name) = LOWER($2) AND active = TRUE LIMIT 1`
		err := tx.QueryRow(ctx, checkQuery, providerID, product.Name).Scan(&existingID)

		if err != nil {
			// Product doesn't exist — insert
			insertQuery := `
				INSERT INTO products (provider_id, name, price, unit, category, description)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = tx.Exec(ctx, insertQuery,
				providerID,
				product.Name,
				product.Price,
				product.Unit,
				product.Category,
				product.Description,
			)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to insert product '%s': %w", product.Name, err)
			}
			newCount++
		} else {
			// Product exists — update price and other fields
			updateQuery := `
				UPDATE products
				SET price = $1, unit = $2, category = $3, description = $4, updated_at = NOW()
				WHERE id = $5
			`
			_, err = tx.Exec(ctx, updateQuery,
				product.Price,
				product.Unit,
				product.Category,
				product.Description,
				existingID,
			)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to update product '%s': %w", product.Name, err)
			}
			updatedCount++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCount, updatedCount, nil
}
