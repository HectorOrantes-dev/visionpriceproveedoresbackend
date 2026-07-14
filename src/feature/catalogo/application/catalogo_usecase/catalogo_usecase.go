package catalogo_usecase

import (
	"context"

	"github.com/visionprice/proveedores-backend/src/feature/catalogo/domain"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/domain/entities"
)

// CatalogoUseCase orchestrates catalog queries for the gateway.
type CatalogoUseCase struct {
	repo domain.CatalogoRepository
}

// NewCatalogoUseCase creates a new CatalogoUseCase.
func NewCatalogoUseCase(repo domain.CatalogoRepository) *CatalogoUseCase {
	return &CatalogoUseCase{repo: repo}
}

// ProductosCercanos returns products near a point within a radius. `categoria`
// may carry one or several canonical categories (comma-separated) sent by the
// app for the item being quoted; we expand them to the supplier-category
// synonyms so only providers that carry that material are returned.
func (uc *CatalogoUseCase) ProductosCercanos(ctx context.Context, lat, lng, radioKm float64, categoria string) ([]entities.Producto, error) {
	patrones := domain.ExpandCategorias(categoria)
	return uc.repo.FindNearby(ctx, lat, lng, radioKm, patrones)
}

// ProductosPorIDs returns products by their ids (for price recalculation).
func (uc *CatalogoUseCase) ProductosPorIDs(ctx context.Context, ids []string) ([]entities.Producto, error) {
	if len(ids) == 0 {
		return []entities.Producto{}, nil
	}
	return uc.repo.FindByIDs(ctx, ids)
}
