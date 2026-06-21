package ml_usecase

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/ml/application/algorithms"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

const (
	// similarityThreshold is the minimum cosine score to consider two items duplicates.
	similarityThreshold = 0.85
	// parallelMinProducts is the catalog size above which the O(N²) comparison is
	// split across CPU cores. Below it, the goroutine overhead outweighs the gain.
	parallelMinProducts = 200
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

	// 2. Build TF-IDF vectors (sequential, O(N); shared read-only afterwards)
	_, tfidfs := algorithms.BuildCorpusTFIDF(corpus)

	// 3. Compare all pairs O(N²/2). The work is pure CPU, so for large catalogs
	//    we fan it out across cores; small catalogs stay sequential.
	if len(products) < parallelMinProducts {
		return detectDuplicatesSequential(products, tfidfs), nil
	}
	return detectDuplicatesParallel(ctx, products, tfidfs), nil
}

// detectDuplicatesSequential is the simple single-goroutine comparison.
func detectDuplicatesSequential(products []*productEntities.Product, tfidfs []map[string]float64) []*entities.DuplicatePair {
	duplicates := []*entities.DuplicatePair{}
	n := len(products)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if pair := comparePair(products, tfidfs, i, j); pair != nil {
				duplicates = append(duplicates, pair)
			}
		}
	}
	return duplicates
}

// detectDuplicatesParallel splits the outer loop across worker goroutines.
//
// Safety: tfidfs and products are read-only here; each worker accumulates into
// its own slice (results[workerID]) so there is no shared mutable state and no
// mutex on the hot path. Rows are handed out via an atomic counter, which both
// distributes work and balances load (early rows do more comparisons than late
// ones, so contiguous splitting would be uneven).
func detectDuplicatesParallel(ctx context.Context, products []*productEntities.Product, tfidfs []map[string]float64) []*entities.DuplicatePair {
	n := len(products)

	numWorkers := runtime.NumCPU()
	if numWorkers > n {
		numWorkers = n
	}

	var nextRow int64 = -1 // atomic.AddInt64 returns the post-increment value
	results := make([][]*entities.DuplicatePair, numWorkers)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var local []*entities.DuplicatePair
			// Publish this worker's results before wg.Done() (deferred LIFO order),
			// regardless of which return path the loop takes.
			defer func() { results[workerID] = local }()
			for {
				i := int(atomic.AddInt64(&nextRow, 1))
				if i >= n {
					return
				}
				// Respect request cancellation between rows.
				if ctx.Err() != nil {
					return
				}
				for j := i + 1; j < n; j++ {
					if pair := comparePair(products, tfidfs, i, j); pair != nil {
						local = append(local, pair)
					}
				}
			}
		}(w)
	}
	wg.Wait()

	duplicates := []*entities.DuplicatePair{}
	for _, r := range results {
		duplicates = append(duplicates, r...)
	}
	return duplicates
}

// comparePair returns a DuplicatePair if products i and j are similar enough, else nil.
func comparePair(products []*productEntities.Product, tfidfs []map[string]float64, i, j int) *entities.DuplicatePair {
	score := algorithms.CosineSimilarity(tfidfs[i], tfidfs[j])
	if score < similarityThreshold {
		return nil
	}
	return &entities.DuplicatePair{
		ProductID1:      products[i].ID,
		ProductName1:    products[i].Name,
		ProductID2:      products[j].ID,
		ProductName2:    products[j].Name,
		SimilarityScore: score,
	}
}
