package product_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
)

// DeleteProduct performs a soft delete on a product ensuring ownership.
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, providerID string, productID string) error {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	prodID, err := uuid.Parse(productID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de producto inválido")
	}

	existing, err := uc.repo.FindByID(ctx, prodID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
	}

	// IDOR defense: ownership is checked against the session provider (pid).
	// A non-owner gets a 404, identical to a non-existent product, to avoid leaking IDs.
	if existing.ProviderID != pid {
		return domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
	}

	if err := uc.repo.SoftDelete(ctx, prodID); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al eliminar el producto")
	}

	return nil
}
