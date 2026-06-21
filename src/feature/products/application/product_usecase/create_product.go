package product_usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// CreateProduct creates a new product for the given provider, enforcing the
// product limit of the provider's subscription plan.
func (uc *ProductUseCase) CreateProduct(ctx context.Context, providerID string, req *entities.CreateProductRequest) (*entities.Product, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	// Enforce the plan's product cap (Plan Max is unlimited).
	limit, unlimited, err := uc.planLimit.ProductLimit(ctx, providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar el límite del plan")
	}
	if !unlimited {
		count, err := uc.repo.CountActiveByProvider(ctx, pid)
		if err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar el inventario actual")
		}
		if count >= limit {
			return nil, domainErrors.NewDomainError(domainErrors.ErrPaymentRequired,
				fmt.Sprintf("Has alcanzado el límite de tu plan (%d productos). Sube de plan para agregar más.", limit))
		}
	}

	product := &entities.Product{
		ProviderID:  pid,
		Name:        req.Name,
		Price:       req.Price,
		Unit:        req.Unit,
		Category:    req.Category,
		Description: req.Description,
		Active:      true,
	}

	created, err := uc.repo.Create(ctx, product)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al crear el producto")
	}

	return created, nil
}
