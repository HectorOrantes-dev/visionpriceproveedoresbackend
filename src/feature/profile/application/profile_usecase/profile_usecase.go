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

// UpdateProfile applies partial updates to the authenticated provider's profile.
func (uc *ProfileUseCase) UpdateProfile(ctx context.Context, providerID string, req *entities.UpdateProfileRequest) (*entities.Profile, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	if req.Email != nil {
		taken, err := uc.repo.EmailExists(ctx, *req.Email, pid)
		if err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar email")
		}
		if taken {
			return nil, domainErrors.NewDomainError(domainErrors.ErrConflict, "El correo ya está registrado por otro proveedor")
		}
	}

	profile, err := uc.repo.Update(ctx, pid, req)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al actualizar perfil")
	}
	return profile, nil
}
