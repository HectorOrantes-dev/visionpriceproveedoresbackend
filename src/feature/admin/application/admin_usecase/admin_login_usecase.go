package admin_usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// AdminLogin authenticates a system admin user by email and password.
// Returns the SystemUser if credentials are valid.
func (uc *AdminUseCase) AdminLogin(ctx context.Context, req *entities.AdminLoginRequest) (*entities.SystemUser, error) {
	user, err := uc.repo.FindSystemUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Credenciales inválidas")
	}

	return user, nil
}
