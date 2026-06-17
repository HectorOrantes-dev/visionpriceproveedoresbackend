package admin_usecase

import (
	"context"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// GetProviderMapPins retrieves all active provider locations for the admin map (HU_SYS_02).
// Returns ONLY public-safe data: business_name, city, state, created_at, latitude, longitude.
func (uc *AdminUseCase) GetProviderMapPins(ctx context.Context) ([]*entities.ProviderMapPin, error) {
	pins, err := uc.repo.GetProviderMapPins(ctx)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener los pines del mapa")
	}

	return pins, nil
}
