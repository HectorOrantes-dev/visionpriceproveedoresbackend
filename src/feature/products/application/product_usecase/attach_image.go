package product_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
)

// AttachImage stores an already-uploaded image URL on a product, scoped to its
// owner. It runs from the background upload goroutine, after the file has been
// pushed to R2, so the create/response path never waits for object storage.
func (uc *ProductUseCase) AttachImage(ctx context.Context, providerID string, productID string, imageURL string) error {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	prodID, err := uuid.Parse(productID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de producto inválido")
	}

	if err := uc.repo.UpdateImageURL(ctx, pid, prodID, imageURL); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al adjuntar la imagen")
	}

	return nil
}
