package entities

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a material/product offered by a provider.
type Product struct {
	ID          uuid.UUID `json:"id"`
	ProviderID  uuid.UUID `json:"provider_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Unit        string    `json:"unit"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the DTO for creating a new product.
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"required,min=1,max=50"`
	Category    string  `json:"category" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// UpdateProductRequest is the DTO for updating a product.
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Price       *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
	Unit        *string  `json:"unit,omitempty" binding:"omitempty,min=1,max=50"`
	Category    *string  `json:"category,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string  `json:"description,omitempty"`
}

// MetricsSummary is the stub DTO for provider metrics (HU_PROV_05).
type MetricsSummary struct {
	TotalBudgetAppearances int    `json:"total_budget_appearances"`
	ContactClicks          int    `json:"contact_clicks"`
	Period                 string `json:"period"`
}

// TopProduct is the stub DTO for top quoted products (HU_PROV_05).
type TopProduct struct {
	ProductID  uuid.UUID `json:"product_id"`
	Name       string    `json:"name"`
	QuoteCount int       `json:"quote_count"`
}
