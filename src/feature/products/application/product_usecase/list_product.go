package product_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// GetProduct retrieves a product by ID, ensuring it belongs to the provider.
func (uc *ProductUseCase) GetProduct(ctx context.Context, providerID string, productID string) (*entities.Product, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	prodID, err := uuid.Parse(productID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de producto inválido")
	}

	product, err := uc.repo.FindByID(ctx, prodID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
	}

	if product.ProviderID != pid {
		return nil, domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "No tiene permisos para ver este producto")
	}

	return product, nil
}

// ListProducts retrieves all active products for a provider.
func (uc *ProductUseCase) ListProducts(ctx context.Context, providerID string) ([]*entities.Product, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	products, err := uc.repo.FindAllActive(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener productos")
	}

	return products, nil
}
