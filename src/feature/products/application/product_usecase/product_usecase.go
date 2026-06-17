package product_usecase

import (
	"github.com/visionprice/proveedores-backend/src/feature/products/domain"
)

// ProductUseCase contains business logic for product operations.
type ProductUseCase struct {
	repo domain.ProductRepository
}

// NewProductUseCase creates a new ProductUseCase.
func NewProductUseCase(repo domain.ProductRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}
