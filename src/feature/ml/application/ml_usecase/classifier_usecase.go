package ml_usecase

import (
	"context"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/ml/application/algorithms"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
)

// ClassifyProduct compares a product name against the VisionPrice global standard catalog
// using TF-IDF and Cosine Similarity to find the best match.
func (uc *MLUseCase) ClassifyProduct(ctx context.Context, req *entities.ClassifyProductRequest) (*entities.ClassifyProductResponse, error) {
	catalog, err := uc.repo.GetStandardCatalog(ctx)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener el catálogo maestro")
	}

	if len(catalog) == 0 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "El catálogo maestro está vacío")
	}

	// 1. Prepare the corpus from standard products
	var corpus [][]string
	for _, stdProd := range catalog {
		tokens := algorithms.Tokenize(stdProd.Name)
		corpus = append(corpus, tokens)
	}

	// Add the incoming product as the last document in the corpus to build shared IDF
	targetTokens := algorithms.Tokenize(req.Name)
	if len(targetTokens) == 0 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El nombre del producto no contiene palabras válidas")
	}
	corpus = append(corpus, targetTokens)

	// 2. Build TF-IDF vectors
	_, tfidfs := algorithms.BuildCorpusTFIDF(corpus)

	// The last vector is our target product
	targetVector := tfidfs[len(tfidfs)-1]

	// 3. Compare target against all standard products
	bestScore := -1.0
	var bestMatch *entities.StandardProduct

	for i, stdProd := range catalog {
		score := algorithms.CosineSimilarity(targetVector, tfidfs[i])
		if score > bestScore {
			bestScore = score
			bestMatch = stdProd
		}
	}

	if bestMatch == nil || bestScore == 0.0 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "No se encontró una categoría adecuada para este producto")
	}

	return &entities.ClassifyProductResponse{
		OriginalName: req.Name,
		StandardID:   bestMatch.ID,
		StandardName: bestMatch.Name,
		Category:     bestMatch.Category,
		Confidence:   bestScore,
	}, nil
}
