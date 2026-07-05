package domain

import (
	"context"

	"github.com/visionprice/proveedores-backend/src/feature/catalogo/domain/entities"
)

// CatalogoRepository is the port for reading the supplier catalog. Swapping the
// underlying tables/columns only touches the adapter, not the use case.
type CatalogoRepository interface {
	// FindNearby returns active products from active suppliers within radioKm of
	// (lat,lng), optionally filtered by categoria (empty = no filter), ordered by
	// distance. Returns an empty slice (never nil) when there are no matches.
	FindNearby(ctx context.Context, lat, lng, radioKm float64, categoria string) ([]entities.Producto, error)

	// FindByIDs returns active products whose id is in ids. distancia_km is 0
	// (no reference point). Returns an empty slice (never nil) when none match.
	FindByIDs(ctx context.Context, ids []string) ([]entities.Producto, error)
}
