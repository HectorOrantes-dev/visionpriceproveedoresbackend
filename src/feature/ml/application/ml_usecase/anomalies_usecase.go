package ml_usecase

import (
	"context"

	"github.com/e-XpertSolutions/go-iforest/v2/iforest"
	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
)

// DetectAnomalies uses Isolation Forest to find anomalous prices in a provider's catalog.
func (uc *MLUseCase) DetectAnomalies(ctx context.Context, providerID uuid.UUID) ([]*entities.AnomalyItem, error) {
	products, err := uc.repo.GetProviderProducts(ctx, providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener los productos del proveedor")
	}

	var anomalies []*entities.AnomalyItem
	const anomalyThreshold = 0.65 // Isolation forest score threshold (typically > 0.6 is anomalous)

	// Group provider's products by Category and Unit to ensure we are comparing apples to apples
	type groupKey struct {
		Category string
		Unit     string
	}
	groupedProducts := make(map[groupKey][]int) // value is the index in the original products slice
	for i, p := range products {
		key := groupKey{Category: p.Category, Unit: p.Unit}
		groupedProducts[key] = append(groupedProducts[key], i)
	}

	// For each group, fetch global market data and train an Isolation Forest
	for key, indices := range groupedProducts {
		marketPrices, err := uc.repo.GetCategoryPrices(ctx, key.Category, key.Unit)
		if err != nil || len(marketPrices) < 10 {
			// Skip if error or not enough global data to establish a baseline
			continue
		}

		// Prepare 2D slice for iforest (it expects [][]float64 where inner slice is features)
		// We only have 1 feature: price
		var trainingData [][]float64
		for _, price := range marketPrices {
			trainingData = append(trainingData, []float64{price})
		}

		// Initialize and train Isolation Forest
		// Trees: 100, SubsampleSize: min(256, len(data))
		subsampleSize := 256
		if len(trainingData) < 256 {
			subsampleSize = len(trainingData)
		}
		
		// 0.1 is the expected anomaly ratio
		forest := iforest.NewForest(100, subsampleSize, 0.1)
		forest.Train(trainingData)

		// Prepare test data for the provider's products
		var testData [][]float64
		for _, idx := range indices {
			testData = append(testData, []float64{products[idx].Price})
		}
		
		// Get anomaly scores
		_, scores, err := forest.Predict(testData)
		if err == nil {
			for i, idx := range indices {
				score := scores[i]
				if score >= anomalyThreshold {
					p := products[idx]
					anomalies = append(anomalies, &entities.AnomalyItem{
						ProductID:    p.ID,
						ProductName:  p.Name,
						Price:        p.Price,
						Unit:         p.Unit,
						Category:     p.Category,
						AnomalyScore: score,
					})
				}
			}
		}
	}

	if anomalies == nil {
		anomalies = []*entities.AnomalyItem{}
	}

	return anomalies, nil
}
