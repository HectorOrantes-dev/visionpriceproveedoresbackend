package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// SupabaseMLRepository implements MLRepository using PostgreSQL (Supabase).
type SupabaseMLRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseMLRepository creates a new SupabaseMLRepository.
func NewSupabaseMLRepository(db *pgxpool.Pool) *SupabaseMLRepository {
	return &SupabaseMLRepository{db: db}
}

// GetStandardCatalog retrieves the full master catalog for TF-IDF training.
func (r *SupabaseMLRepository) GetStandardCatalog(ctx context.Context) ([]*entities.StandardProduct, error) {
	query := `SELECT id, name, category FROM standard_products`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query standard products: %w", err)
	}
	defer rows.Close()

	var catalog []*entities.StandardProduct
	for rows.Next() {
		sp := &entities.StandardProduct{}
		if err := rows.Scan(&sp.ID, &sp.Name, &sp.Category); err != nil {
			return nil, fmt.Errorf("failed to scan standard product: %w", err)
		}
		catalog = append(catalog, sp)
	}

	if catalog == nil {
		catalog = []*entities.StandardProduct{}
	}

	return catalog, nil
}

// GetProviderProducts retrieves all active products for a specific provider.
func (r *SupabaseMLRepository) GetProviderProducts(ctx context.Context, providerID uuid.UUID) ([]*productEntities.Product, error) {
	query := `
		SELECT id, provider_id, name, price, unit, category, description, active, created_at, updated_at
		FROM products
		WHERE provider_id = $1 AND active = TRUE
	`

	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider products: %w", err)
	}
	defer rows.Close()

	var products []*productEntities.Product
	for rows.Next() {
		p := &productEntities.Product{}
		if err := rows.Scan(
			&p.ID, &p.ProviderID, &p.Name, &p.Price, &p.Unit,
			&p.Category, &p.Description, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if products == nil {
		products = []*productEntities.Product{}
	}

	return products, nil
}

// GetCategoryPrices retrieves all prices for a specific category and unit across the entire system.
func (r *SupabaseMLRepository) GetCategoryPrices(ctx context.Context, category string, unit string) ([]float64, error) {
	query := `
		SELECT price 
		FROM products 
		WHERE category = $1 AND unit = $2 AND active = TRUE
	`

	rows, err := r.db.Query(ctx, query, category, unit)
	if err != nil {
		return nil, fmt.Errorf("failed to query category prices: %w", err)
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var price float64
		if err := rows.Scan(&price); err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, price)
	}

	if prices == nil {
		prices = []float64{}
	}

	return prices, nil
}
