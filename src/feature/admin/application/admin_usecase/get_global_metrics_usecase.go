package admin_usecase

import (
	"context"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// GetGlobalMetrics retrieves the system-wide metrics for the admin dashboard (HU_SYS_01).
func (uc *AdminUseCase) GetGlobalMetrics(ctx context.Context) (*entities.GlobalMetrics, error) {
	metrics, err := uc.repo.GetGlobalMetrics(ctx)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener las métricas globales")
	}

	return metrics, nil
}
