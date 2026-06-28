package geolocation_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain/entities"
)

// GeolocationUseCase contains business logic for geolocation operations.
type GeolocationUseCase struct {
	repo domain.GeolocationRepository
}

// NewGeolocationUseCase creates a new GeolocationUseCase.
func NewGeolocationUseCase(repo domain.GeolocationRepository) *GeolocationUseCase {
	return &GeolocationUseCase{repo: repo}
}

// SetLocation creates or updates the provider's warehouse address.
func (uc *GeolocationUseCase) SetLocation(ctx context.Context, providerID string, req *entities.SetLocationRequest) (*entities.ProviderLocation, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	location := &entities.ProviderLocation{
		ProviderID: pid,
		Address:    req.Address,
	}

	result, err := uc.repo.UpsertLocation(ctx, location)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al guardar la ubicación")
	}

	return result, nil
}

// GetLocation retrieves the provider's warehouse address.
func (uc *GeolocationUseCase) GetLocation(ctx context.Context, providerID string) (*entities.ProviderLocation, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	location, err := uc.repo.GetLocation(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Ubicación no encontrada")
	}

	return location, nil
}
