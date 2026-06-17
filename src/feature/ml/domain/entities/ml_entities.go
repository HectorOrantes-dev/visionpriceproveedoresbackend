package entities

import "github.com/google/uuid"

// StandardProduct represents a product from the VisionPrice global master catalog.
type StandardProduct struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ClassifyProductRequest is the payload to classify a single product name.
type ClassifyProductRequest struct {
	Name string `json:"name" binding:"required"`
}

// ClassifyProductResponse represents the closest standard product match.
type ClassifyProductResponse struct {
	OriginalName string  `json:"original_name"`
	StandardID   int     `json:"standard_id"`
	StandardName string  `json:"standard_name"`
	Category     string  `json:"category"`
	Confidence   float64 `json:"confidence"`
}

// DuplicatePair represents two products that are potentially duplicates.
type DuplicatePair struct {
	ProductID1      uuid.UUID `json:"product_id_1"`
	ProductName1    string    `json:"product_name_1"`
	ProductID2      uuid.UUID `json:"product_id_2"`
	ProductName2    string    `json:"product_name_2"`
	SimilarityScore float64   `json:"similarity_score"`
}

// AnomalyItem represents a product with an anomalous price.
type AnomalyItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Price        float64   `json:"price"`
	Unit         string    `json:"unit"`
	Category     string    `json:"category"`
	AnomalyScore float64   `json:"anomaly_score"`
}
