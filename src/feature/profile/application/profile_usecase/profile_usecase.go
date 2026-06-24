package profile_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/profile/domain"
	"github.com/visionprice/proveedores-backend/src/feature/profile/domain/entities"
)

// ProfileUseCase contains the business logic for reading a provider's profile.
type ProfileUseCase struct {
	repo domain.ProfileRepository
}

// NewProfileUseCase creates a new ProfileUseCase.
func NewProfileUseCase(repo domain.ProfileRepository) *ProfileUseCase {
	return &ProfileUseCase{repo: repo}
}

// GetProfile returns the authenticated provider's profile.
func (uc *ProfileUseCase) GetProfile(ctx context.Context, providerID string) (*entities.Profile, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	profile, err := uc.repo.GetByID(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Proveedor no encontrado")
	}
	return profile, nil
}
