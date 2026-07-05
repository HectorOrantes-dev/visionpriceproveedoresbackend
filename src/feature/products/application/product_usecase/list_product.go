package product_usecase

import (
	"context"
	"log/slog"

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

	// IDOR defense: the owner is taken from the session (pid), never from the URL.
	// If the resource exists but does not belong to the caller, respond as if it
	// did not exist (404) so an attacker cannot probe for valid IDs of other users.
	if product.ProviderID != pid {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Producto no encontrado")
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
		// Log the underlying DB error so the real cause is visible in the
		// server logs; the client still gets a generic 500 message.
		slog.Error("products: failed to list", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener productos")
	}

	return products, nil
}
