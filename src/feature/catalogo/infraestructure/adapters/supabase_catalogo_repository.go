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
//	proveedores(id, nombre, latitud, longitud, activo)
//	productos(id, proveedor_id, nombre, categoria, unidad, precio_unitario, rendimiento_m2, activo)
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
// (''=no filter), $4=radio_km. LEAST/GREATEST clamp the acos argument to
// [-1,1] to avoid NaN from floating-point rounding on exact-same coordinates.
const nearbyQuery = `
	SELECT * FROM (
		SELECT p.id AS producto_id, p.nombre, p.categoria, p.unidad,
		       p.precio_unitario, p.rendimiento_m2,
		       pr.id AS proveedor_id, pr.nombre AS proveedor_nombre,
		       (6371 * acos(LEAST(1, GREATEST(-1,
		          cos(radians($1)) * cos(radians(pr.latitud)) *
		          cos(radians(pr.longitud) - radians($2)) +
		          sin(radians($1)) * sin(radians(pr.latitud))
		       )))) AS distancia_km
		FROM productos p
		JOIN proveedores pr ON pr.id = p.proveedor_id
		WHERE p.activo AND pr.activo
		  AND ($3 = '' OR p.categoria = $3)
	) t
	WHERE t.distancia_km <= $4
	ORDER BY t.distancia_km
`

// byIDsQuery returns the same shape with distancia_km = 0 (no reference point).
const byIDsQuery = `
	SELECT p.id AS producto_id, p.nombre, p.categoria, p.unidad,
	       p.precio_unitario, p.rendimiento_m2,
	       pr.id AS proveedor_id, pr.nombre AS proveedor_nombre,
	       0::float8 AS distancia_km
	FROM productos p
	JOIN proveedores pr ON pr.id = p.proveedor_id
	WHERE p.activo AND pr.activo AND p.id = ANY($1)
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
func (r *SupabaseCatalogoRepository) FindByIDs(ctx context.Context, ids []int) ([]entities.Producto, error) {
	// Use int64 so the array encodes as bigint[] and matches the bigint id column.
	arr := make([]int64, len(ids))
	for i, v := range ids {
		arr[i] = int64(v)
	}
	rows, err := r.db.Query(ctx, byIDsQuery, arr)
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
			&p.PrecioUnitario, &p.RendimientoM2,
			&p.Proveedor.ProveedorID, &p.Proveedor.Nombre, &p.Proveedor.DistanciaKm,
		); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}
	return productos, rows.Err()
}
