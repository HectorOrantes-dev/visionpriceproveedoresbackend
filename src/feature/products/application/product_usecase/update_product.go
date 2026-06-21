package product_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// UpdateProduct updates an existing product ensuring ownership.
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, providerID string, productID string, req *entities.UpdateProductRequest) (*entities.Product, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	prodID, err := uuid.Parse(productID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de producto inválido")
	}

	// Fetch existing product
	existing, err := uc.repo.FindByID(ctx, prodID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
	}

	// IDOR defense: ownership is checked against the session provider (pid).
	// A non-owner gets a 404, identical to a non-existent product, to avoid leaking IDs.
	if existing.ProviderID != pid {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
	}

	// Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.Unit != nil {
		existing.Unit = *req.Unit
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.Description != nil {
		existing.Description = req.Description
	}

	updated, err := uc.repo.Update(ctx, existing)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al actualizar el producto")
	}

	return updated, nil
}
