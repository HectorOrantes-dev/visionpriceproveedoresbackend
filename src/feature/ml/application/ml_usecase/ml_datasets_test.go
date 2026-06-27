package ml_usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

type ProductToClassify struct {
	Name                 string `json:"name"`
	MatchEsperadoID      int    `json:"match_esperado_id"`
	MatchEsperadoNombre  string `json:"match_esperado_nombre"`
	CategoriaEsperada    string `json:"categoria_esperada"`
	ConfianzaEsperada    string `json:"confianza_esperada"`
}

type Dataset struct {
	Description                string                      `json:"description"`
	Algorithm                  string                      `json:"algorithm"`
	ExpectedBehavior           string                      `json:"expected_behavior"`
	ProviderID                 string                      `json:"provider_id"`
	MarketPricesByCategoryUnit map[string][]float64        `json:"market_prices_by_category_unit"`
	Products                   []DatasetProduct            `json:"products"`
	StandardCatalog            []*entities.StandardProduct `json:"standard_catalog"`
	ProductosAClasificar       []ProductToClassify         `json:"productos_a_clasificar"`
}

type DatasetProduct struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Unit     string  `json:"unit"`
	Price    float64 `json:"price"`
}

func TestMLDatasets(t *testing.T) {
	// Root paths
	datasetsDir := "../../../../../ml_test_datasets"
	catalogPath := filepath.Join(datasetsDir, "classification", "dataset_14_catalogo_maestro_estandar.json")

	// Load the standard catalog first
	catalogBytes, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("Failed to read standard catalog: %v", err)
	}
	var catalogDataset Dataset
	if err := json.Unmarshal(catalogBytes, &catalogDataset); err != nil {
		t.Fatalf("Failed to unmarshal standard catalog: %v", err)
	}
	standardCatalog := catalogDataset.StandardCatalog

	err = filepath.WalkDir(datasetsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip directories, non-JSON files, and the master catalog itself
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") || d.Name() == "dataset_14_catalogo_maestro_estandar.json" {
			return nil
		}

		t.Run(d.Name(), func(t *testing.T) {
			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read dataset: %v", err)
			}
			var ds Dataset
			if err := json.Unmarshal(bytes, &ds); err != nil {
				t.Fatalf("Failed to parse dataset: %v", err)
			}

			// Prepare Mock Repository
			mockRepo := &MockMLRepository{
				StandardCatalog: standardCatalog,
				MarketPrices:    ds.MarketPricesByCategoryUnit,
			}

			// Map mock products
			var mockProducts []*productEntities.Product
			var providerUUID uuid.UUID
			if ds.ProviderID != "" {
				providerUUID = uuid.MustParse(ds.ProviderID)
			}

			for _, p := range ds.Products {
				// Generate UUID if the dataset id is not a real UUID or empty
				prodID, parseErr := uuid.Parse(p.ID)
				if parseErr != nil {
					prodID = uuid.New()
				}
				mockProducts = append(mockProducts, &productEntities.Product{
					ID:         prodID,
					ProviderID: providerUUID,
					Name:       p.Name,
					Category:   p.Category,
					Unit:       p.Unit,
					Price:      p.Price,
				})
			}
			mockRepo.Products = mockProducts

			// Initialize use case with mock repository
			useCase := NewMLUseCase(mockRepo)
			ctx := context.Background()

			fmt.Printf("\n======================================================\n")
			fmt.Printf("🧪 Testing: %s\n", d.Name())
			fmt.Printf("📝 Expected Behavior: %s\n", ds.ExpectedBehavior)
			fmt.Printf("------------------------------------------------------\n")

			switch ds.Algorithm {
			case "isolation_forest":
				anomalies, err := useCase.DetectAnomalies(ctx, providerUUID)
				if err != nil {
					t.Fatalf("DetectAnomalies failed: %v", err)
				}
				fmt.Printf("✅ Result: Found %d anomalies\n", len(anomalies))
				for _, a := range anomalies {
					fmt.Printf("   - ⚠️ %s | Precio: %.2f | Anomaly Score: %.3f\n", a.ProductName, a.Price, a.AnomalyScore)
				}

			case "tfidf_cosine_duplicates":
				duplicates, err := useCase.DetectDuplicates(ctx, providerUUID)
				if err != nil {
					t.Fatalf("DetectDuplicates failed: %v", err)
				}
				fmt.Printf("✅ Result: Found %d duplicate pairs\n", len(duplicates))
				for _, dup := range duplicates {
					fmt.Printf("   - 🔄 '%s' <-> '%s' | Similitud: %.2f\n", dup.ProductName1, dup.ProductName2, dup.SimilarityScore)
				}

			case "tfidf_cosine_classifier":
				for _, prod := range ds.ProductosAClasificar {
					req := &entities.ClassifyProductRequest{Name: prod.Name}
					res, err := useCase.ClassifyProduct(ctx, req)
					if err != nil {
						t.Fatalf("ClassifyProduct failed for '%s': %v", prod.Name, err)
					}
					fmt.Printf("   - 🏷️ '%s'\n", prod.Name)
					fmt.Printf("      Match: '%s' (Confianza: %.2f) | Esperaba confianza: %s\n", res.StandardName, res.Confidence, prod.ConfianzaEsperada)
				}

			case "all":
				fmt.Printf("✅ Edge case (end-to-end integration scenario). Skipping specific algorithm.\n")

			default:
				t.Logf("Unknown algorithm '%s', skipping", ds.Algorithm)
			}
			fmt.Printf("======================================================\n")
		})
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk datasets directory: %v", err)
	}
}
