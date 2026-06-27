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

// ProductosCercanos returns products near a point within a radius.
func (uc *CatalogoUseCase) ProductosCercanos(ctx context.Context, lat, lng, radioKm float64, categoria string) ([]entities.Producto, error) {
	return uc.repo.FindNearby(ctx, lat, lng, radioKm, categoria)
}

// ProductosPorIDs returns products by their ids (for price recalculation).
func (uc *CatalogoUseCase) ProductosPorIDs(ctx context.Context, ids []int) ([]entities.Producto, error) {
	if len(ids) == 0 {
		return []entities.Producto{}, nil
	}
	return uc.repo.FindByIDs(ctx, ids)
}
