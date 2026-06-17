package ml_usecase

import (
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain"
)

// MLUseCase contains the business logic for the Machine Learning features.
type MLUseCase struct {
	repo domain.MLRepository
}

// NewMLUseCase creates a new MLUseCase.
func NewMLUseCase(repo domain.MLRepository) *MLUseCase {
	return &MLUseCase{repo: repo}
}
