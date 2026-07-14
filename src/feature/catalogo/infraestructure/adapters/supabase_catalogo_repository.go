package adapters

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/catalogo/domain/entities"
)

// SupabaseCatalogoRepository implements CatalogoRepository over PostgreSQL.
//
// It reads two tables:
//
//	providers(id, business_name, active)
//	provider_locations(provider_id, lat, lng, delivery_radius_km)
//	products(id, provider_id, name, category, unit, price, image_url, active)
//
// If your real columns differ, this adapter is the ONLY place to adjust the SQL.
type SupabaseCatalogoRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseCatalogoRepository creates a new SupabaseCatalogoRepository.
func NewSupabaseCatalogoRepository(db *pgxpool.Pool) *SupabaseCatalogoRepository {
	return &SupabaseCatalogoRepository{db: db}
}

// nearbyQuery filters by haversine distance. $1=lat, $2=lng, $3=categoria
// patterns (text[]; empty/NULL = no filter), $4=radio_km. LEAST/GREATEST clamp
// the acos argument to [-1,1] to avoid NaN from floating-point rounding on
// exact-same coordinates. The category filter uses ILIKE ANY so a provider whose
// category reads "Pintura vinílica" still matches the item's "pintura".
const nearbyQuery = `
	SELECT * FROM (
		SELECT p.id AS producto_id, p.name AS nombre, p.category AS categoria, p.unit AS unidad,
		       p.price AS precio_unitario, COALESCE(p.rendimiento_m2, 0) AS rendimiento_m2, COALESCE(p.image_url, '') AS image_url,
		       pr.id AS proveedor_id, pr.business_name AS proveedor_nombre,
		       (6371 * acos(LEAST(1, GREATEST(-1,
		          cos(radians($1)) * cos(radians(pl.lat)) *
		          cos(radians(pl.lng) - radians($2)) +
		          sin(radians($1)) * sin(radians(pl.lat))
		       )))) AS distancia_km, pl.lat, pl.lng
		FROM products p
		JOIN providers pr ON pr.id = p.provider_id
		JOIN provider_locations pl ON pl.provider_id = pr.id
		WHERE p.active AND pr.active
		  AND (COALESCE(cardinality($3::text[]), 0) = 0 OR p.category ILIKE ANY($3))
	) t
	WHERE t.distancia_km <= $4
	ORDER BY t.distancia_km
`

// byIDsQuery returns the same shape with distancia_km = 0 (no reference point).
// p.id is a uuid column and $1 arrives as text[]; we compare p.id::text so
// Postgres doesn't error with "operator does not exist: uuid = text". lower()
// makes it case-insensitive, and an id that doesn't exist just yields no rows
// (empty list) instead of a 500.
const byIDsQuery = `
	SELECT p.id AS producto_id, p.name AS nombre, p.category AS categoria, p.unit AS unidad,
	       p.price AS precio_unitario, COALESCE(p.rendimiento_m2, 0) AS rendimiento_m2, COALESCE(p.image_url, '') AS image_url,
	       pr.id AS proveedor_id, pr.business_name AS proveedor_nombre,
	       0::float8 AS distancia_km, pl.lat, pl.lng
	FROM products p
	JOIN providers pr ON pr.id = p.provider_id
	LEFT JOIN provider_locations pl ON pl.provider_id = pr.id
	WHERE p.active AND pr.active AND lower(p.id::text) = ANY($1)
`

// FindNearby returns active products within radioKm of (lat,lng), optionally
// restricted to the given ILIKE category patterns (nil/empty = no filter).
func (r *SupabaseCatalogoRepository) FindNearby(ctx context.Context, lat, lng, radioKm float64, categoriaPatterns []string) ([]entities.Producto, error) {
	rows, err := r.db.Query(ctx, nearbyQuery, lat, lng, categoriaPatterns, radioKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProductos(rows)
}

// FindByIDs returns active products whose id is in ids. IDs are lowercased to
// match lower(p.id::text) in the query (uuid::text is already lowercase).
func (r *SupabaseCatalogoRepository) FindByIDs(ctx context.Context, ids []string) ([]entities.Producto, error) {
	norm := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		if id != "" {
			norm = append(norm, id)
		}
	}
	rows, err := r.db.Query(ctx, byIDsQuery, norm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProductos(rows)
}

// scanProductos maps rows (9 columns, distancia_km last) into Productos. It
// returns an empty slice (never nil) so the JSON response is [] and not null.
func scanProductos(rows pgx.Rows) ([]entities.Producto, error) {
	productos := []entities.Producto{}
	for rows.Next() {
		var p entities.Producto
		if err := rows.Scan(
			&p.ProductoID, &p.Nombre, &p.Categoria, &p.Unidad,
			&p.PrecioUnitario, &p.RendimientoM2, &p.ImageURL,
			&p.Proveedor.ProveedorID, &p.Proveedor.Nombre, &p.Proveedor.DistanciaKm,
			&p.Proveedor.Lat, &p.Proveedor.Lng,
		); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}
	return productos, rows.Err()
}
