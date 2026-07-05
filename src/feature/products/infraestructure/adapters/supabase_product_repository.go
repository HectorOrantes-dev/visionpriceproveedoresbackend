package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

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
		INSERT INTO products (provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at
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
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return created, nil
}

// FindByID retrieves an active product by ID.
func (r *SupabaseProductRepository) FindByID(ctx context.Context, productID uuid.UUID) (*entities.Product, error) {
	query := `
		SELECT id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at
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
	)

	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return product, nil
}

// FindAllActive retrieves all active products for a provider.
func (r *SupabaseProductRepository) FindAllActive(ctx context.Context, providerID uuid.UUID) ([]*entities.Product, error) {
	query := `
		SELECT id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at
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
		    stock = $7, status = $8, description = $9, image_url = $10, updated_at = NOW()
		WHERE id = $11 AND active = TRUE
		RETURNING id, provider_id, name, sku, brand, price, unit, category, stock, status, description, image_url, active, created_at, updated_at
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
