package admin_usecase

import (
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain"
)

// AdminUseCase contains business logic for admin operations.
type AdminUseCase struct {
	repo domain.AdminRepository
}

// NewAdminUseCase creates a new AdminUseCase.
func NewAdminUseCase(repo domain.AdminRepository) *AdminUseCase {
	return &AdminUseCase{repo: repo}
}
