package product_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// CreateProduct creates a new product for the given provider.
func (uc *ProductUseCase) CreateProduct(ctx context.Context, providerID string, req *entities.CreateProductRequest) (*entities.Product, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
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
