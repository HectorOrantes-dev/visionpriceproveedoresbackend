package adapters

import (
	"context"

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
// (”=no filter), $4=radio_km. LEAST/GREATEST clamp the acos argument to
// [-1,1] to avoid NaN from floating-point rounding on exact-same coordinates.
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
		  AND ($3 = '' OR p.category = $3)
	) t
	WHERE t.distancia_km <= $4
	ORDER BY t.distancia_km
`

// byIDsQuery returns the same shape with distancia_km = 0 (no reference point).
const byIDsQuery = `
	SELECT p.id AS producto_id, p.name AS nombre, p.category AS categoria, p.unit AS unidad,
	       p.price AS precio_unitario, COALESCE(p.rendimiento_m2, 0) AS rendimiento_m2, COALESCE(p.image_url, '') AS image_url,
	       pr.id AS proveedor_id, pr.business_name AS proveedor_nombre,
	       0::float8 AS distancia_km, pl.lat, pl.lng
	FROM products p
	JOIN providers pr ON pr.id = p.provider_id
	LEFT JOIN provider_locations pl ON pl.provider_id = pr.id
	WHERE p.active AND pr.active AND p.id = ANY($1)
`

// FindNearby returns active products within radioKm of (lat,lng).
func (r *SupabaseCatalogoRepository) FindNearby(ctx context.Context, lat, lng, radioKm float64, categoria string) ([]entities.Producto, error) {
	rows, err := r.db.Query(ctx, nearbyQuery, lat, lng, categoria, radioKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProductos(rows)
}

// FindByIDs returns active products whose id is in ids.
func (r *SupabaseCatalogoRepository) FindByIDs(ctx context.Context, ids []int64) ([]entities.Producto, error) {
	rows, err := r.db.Query(ctx, byIDsQuery, ids)
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
