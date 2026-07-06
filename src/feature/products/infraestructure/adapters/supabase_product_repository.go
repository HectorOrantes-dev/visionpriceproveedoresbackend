package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// spanishMonths maps a time.Month to its 3-letter Spanish abbreviation.
var spanishMonths = [...]string{"", "Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}

// SupabaseProductRepository implements ProductRepository using PostgreSQL (Supabase).
type SupabaseProductRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseProductRepository creates a new SupabaseProductRepository.
func NewSupabaseProductRepository(db *pgxpool.Pool) *SupabaseProductRepository {
	return &SupabaseProductRepository{db: db}
}

// Create inserts a new product into the database.
func (r *SupabaseProductRepository) Create(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	query := `
		INSERT INTO products (provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, rendimiento_m2)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at, rendimiento_m2
	`

	created := &entities.Product{}
	err := r.db.QueryRow(ctx, query,
		product.ProviderID,
		product.Name,
		product.SKU,
		product.Brand,
		product.Price,
		product.Unit,
		product.Category,
		product.Stock,
		product.Status,
		product.Description,
		product.ImageURL,
		product.RendimientoM2,
	).Scan(
		&created.ID,
		&created.ProviderID,
		&created.Name,
		&created.SKU,
		&created.Brand,
		&created.Price,
		&created.Unit,
		&created.Category,
		&created.Stock,
		&created.Status,
		&created.Description,
		&created.ImageURL,
		&created.Active,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.RendimientoM2,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return created, nil
}

// FindByID retrieves an active product by ID.
func (r *SupabaseProductRepository) FindByID(ctx context.Context, productID uuid.UUID) (*entities.Product, error) {
	query := `
		SELECT id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at, rendimiento_m2
		FROM products
		WHERE id = $1 AND active = TRUE
	`

	product := &entities.Product{}
	err := r.db.QueryRow(ctx, query, productID).Scan(
		&product.ID,
		&product.ProviderID,
		&product.Name,
		&product.SKU,
		&product.Brand,
		&product.Price,
		&product.Unit,
		&product.Category,
		&product.Stock,
		&product.Status,
		&product.Description,
		&product.ImageURL,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.RendimientoM2,
	)

	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return product, nil
}

// FindAllActive retrieves all active products for a provider.
func (r *SupabaseProductRepository) FindAllActive(ctx context.Context, providerID uuid.UUID) ([]*entities.Product, error) {
	query := `
		SELECT id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at, rendimiento_m2
		FROM products
		WHERE provider_id = $1 AND active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*entities.Product
	for rows.Next() {
		product := &entities.Product{}
		if err := rows.Scan(
			&product.ID,
			&product.ProviderID,
			&product.Name,
			&product.SKU,
			&product.Brand,
			&product.Price,
			&product.Unit,
			&product.Category,
			&product.Stock,
			&product.Status,
			&product.Description,
			&product.ImageURL,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.RendimientoM2,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if products == nil {
		products = []*entities.Product{}
	}

	return products, nil
}

// Update modifies an existing product.
func (r *SupabaseProductRepository) Update(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	query := `
		UPDATE products
		SET name = $1, sku = $2, brand = $3, price = $4, unit = $5, category = $6,
		    stock = $7, status = $8, description = $9, image_url = $10, rendimiento_m2 = $11, updated_at = NOW()
		WHERE id = $12 AND active = TRUE
		RETURNING id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at, rendimiento_m2
	`

	updated := &entities.Product{}
	err := r.db.QueryRow(ctx, query,
		product.Name,
		product.SKU,
		product.Brand,
		product.Price,
		product.Unit,
		product.Category,
		product.Stock,
		product.Status,
		product.Description,
		product.ImageURL,
		product.RendimientoM2,
		product.ID,
	).Scan(
		&updated.ID,
		&updated.ProviderID,
		&updated.Name,
		&updated.SKU,
		&updated.Brand,
		&updated.Price,
		&updated.Unit,
		&updated.Category,
		&updated.Stock,
		&updated.Status,
		&updated.Description,
		&updated.ImageURL,
		&updated.Active,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.RendimientoM2,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return updated, nil
}

// UpdateImageURL sets only the image_url column for a product owned by the
// given provider. The provider_id predicate is an IDOR defense so the async
// upload can never attach an image to another provider's product.
func (r *SupabaseProductRepository) UpdateImageURL(ctx context.Context, providerID, productID uuid.UUID, imageURL string) error {
	const query = `
		UPDATE products
		SET image_url = $1, updated_at = NOW()
		WHERE id = $2 AND provider_id = $3 AND active = TRUE
	`
	tag, err := r.db.Exec(ctx, query, imageURL, productID, providerID)
	if err != nil {
		return fmt.Errorf("failed to update image_url: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found for image update")
	}
	return nil
}

// CountActiveByProvider returns the number of active products for a provider.
func (r *SupabaseProductRepository) CountActiveByProvider(ctx context.Context, providerID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM products WHERE provider_id = $1 AND active = TRUE`
	var count int
	err := r.db.QueryRow(ctx, query, providerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

// MetricsAggregate returns catalog aggregates for a provider.
func (r *SupabaseProductRepository) MetricsAggregate(ctx context.Context, providerID uuid.UUID) (float64, int, float64, int, error) {
	const query = `
		SELECT
			COALESCE(SUM(price * stock), 0),
			COALESCE(SUM(stock), 0),
			COALESCE(AVG(price), 0),
			COUNT(*)
		FROM products
		WHERE provider_id = $1 AND active = TRUE
	`
	var inventoryValue, averagePrice float64
	var unitsInStock, totalMaterials int
	err := r.db.QueryRow(ctx, query, providerID).Scan(&inventoryValue, &unitsInStock, &averagePrice, &totalMaterials)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to aggregate metrics: %w", err)
	}
	return inventoryValue, unitsInStock, averagePrice, totalMaterials, nil
}

// CategoryDistribution returns inventory value grouped by category (descending).
func (r *SupabaseProductRepository) CategoryDistribution(ctx context.Context, providerID uuid.UUID) ([]entities.CategorySlice, error) {
	const query = `
		SELECT category, COALESCE(SUM(price * stock), 0) AS value
		FROM products
		WHERE provider_id = $1 AND active = TRUE
		GROUP BY category
		ORDER BY value DESC
	`
	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category distribution: %w", err)
	}
	defer rows.Close()

	slices := []entities.CategorySlice{}
	for rows.Next() {
		var slice entities.CategorySlice
		if err := rows.Scan(&slice.Category, &slice.Value); err != nil {
			return nil, fmt.Errorf("failed to scan category slice: %w", err)
		}
		slices = append(slices, slice)
	}
	return slices, rows.Err()
}

// MonthlyNewMaterials returns materials created per month for the last 6 months (zero-filled).
func (r *SupabaseProductRepository) MonthlyNewMaterials(ctx context.Context, providerID uuid.UUID) ([]entities.MonthlyPoint, error) {
	const query = `
		SELECT m.month, COALESCE(COUNT(p.id), 0)::float8 AS value
		FROM generate_series(
			date_trunc('month', NOW()) - INTERVAL '5 months',
			date_trunc('month', NOW()),
			INTERVAL '1 month'
		) AS m(month)
		LEFT JOIN products p
			ON date_trunc('month', p.created_at) = m.month
			AND p.provider_id = $1
			AND p.active = TRUE
		GROUP BY m.month
		ORDER BY m.month
	`
	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly materials: %w", err)
	}
	defer rows.Close()

	points := []entities.MonthlyPoint{}
	for rows.Next() {
		var month time.Time
		var value float64
		if err := rows.Scan(&month, &value); err != nil {
			return nil, fmt.Errorf("failed to scan monthly point: %w", err)
		}
		points = append(points, entities.MonthlyPoint{Label: spanishMonths[month.Month()], Value: value})
	}
	return points, rows.Err()
}

// TopByInventoryValue returns the top materials by inventory value (price×stock).
func (r *SupabaseProductRepository) TopByInventoryValue(ctx context.Context, providerID uuid.UUID, limit int) ([]entities.TopProduct, error) {
	const query = `
		SELECT name, COALESCE(price * stock, 0) AS amount
		FROM products
		WHERE provider_id = $1 AND active = TRUE
		ORDER BY amount DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, providerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top products: %w", err)
	}
	defer rows.Close()

	products := []entities.TopProduct{}
	for rows.Next() {
		var product entities.TopProduct
		if err := rows.Scan(&product.Name, &product.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan top product: %w", err)
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

// SoftDelete marks a product as inactive.
func (r *SupabaseProductRepository) SoftDelete(ctx context.Context, productID uuid.UUID) error {
	query := `UPDATE products SET active = FALSE, updated_at = NOW() WHERE id = $1 AND active = TRUE`
	tag, err := r.db.Exec(ctx, query, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found or already deleted")
	}
	return nil
}
