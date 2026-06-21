package admin_usecase

import (
	"context"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// defaultExpiryWindowDays is used when the admin doesn't specify a window.
const defaultExpiryWindowDays = 7

// maxExpiryWindowDays caps the lookahead window.
const maxExpiryWindowDays = 365

// GetExpiringSubscriptions lists accounts whose paid subscription is about to
// expire (or is overdue) within the given window, along with their plan.
func (uc *AdminUseCase) GetExpiringSubscriptions(ctx context.Context, withinDays int) ([]*entities.ExpiringSubscription, error) {
	if withinDays <= 0 {
		withinDays = defaultExpiryWindowDays
	}
	if withinDays > maxExpiryWindowDays {
		withinDays = maxExpiryWindowDays
	}

	list, err := uc.repo.GetExpiringSubscriptions(ctx, withinDays)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener las suscripciones por vencer")
	}

	return list, nil
}
