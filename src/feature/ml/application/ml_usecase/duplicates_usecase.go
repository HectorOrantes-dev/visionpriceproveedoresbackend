package ml_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/ml/application/algorithms"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
)

// DetectDuplicates analyzes a provider's catalog to find items that might be the same product
// written slightly differently (e.g. "Tubo PVC 2 pulg" vs "Tubo de PVC 2 in").
func (uc *MLUseCase) DetectDuplicates(ctx context.Context, providerID uuid.UUID) ([]*entities.DuplicatePair, error) {
	products, err := uc.repo.GetProviderProducts(ctx, providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener los productos del proveedor")
	}

	if len(products) < 2 {
		return []*entities.DuplicatePair{}, nil // Need at least 2 products to find duplicates
	}

	// 1. Prepare the corpus
	var corpus [][]string
	for _, p := range products {
		tokens := algorithms.Tokenize(p.Name)
		corpus = append(corpus, tokens)
	}

	// 2. Build TF-IDF vectors
	_, tfidfs := algorithms.BuildCorpusTFIDF(corpus)

	// 3. Compare O(N^2 / 2) to find duplicates
	var duplicates []*entities.DuplicatePair
	const similarityThreshold = 0.85 // High confidence threshold

	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			score := algorithms.CosineSimilarity(tfidfs[i], tfidfs[j])
			if score >= similarityThreshold {
				duplicates = append(duplicates, &entities.DuplicatePair{
					ProductID1:      products[i].ID,
					ProductName1:    products[i].Name,
					ProductID2:      products[j].ID,
					ProductName2:    products[j].Name,
					SimilarityScore: score,
				})
			}
		}
	}

	return duplicates, nil
}
