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
	SKU         *string   `json:"sku,omitempty"`
	Brand       *string   `json:"brand,omitempty"`
	Price       float64   `json:"price"`
	Unit        string    `json:"unit"`
	Category    string    `json:"category"`
	Stock       int       `json:"stock"`
	Status      string    `json:"status"`
	Description *string   `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the DTO for creating a new product.
// String fields carry the "nohtml" validator as a backend anti-XSS defense.
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255,nohtml"`
	SKU         *string `json:"sku,omitempty" binding:"omitempty,max=100,nohtml"`
	Brand       *string `json:"brand,omitempty" binding:"omitempty,max=255,nohtml"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"required,min=1,max=50,nohtml"`
	Category    string  `json:"category" binding:"required,min=1,max=100,nohtml"`
	Stock       *int    `json:"stock,omitempty" binding:"omitempty,gte=0"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=active draft inactive out_of_stock"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=2000,nohtml"`
}

// UpdateProductRequest is the DTO for updating a product.
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=1,max=255,nohtml"`
	SKU         *string  `json:"sku,omitempty" binding:"omitempty,max=100,nohtml"`
	Brand       *string  `json:"brand,omitempty" binding:"omitempty,max=255,nohtml"`
	Price       *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
	Unit        *string  `json:"unit,omitempty" binding:"omitempty,min=1,max=50,nohtml"`
	Category    *string  `json:"category,omitempty" binding:"omitempty,min=1,max=100,nohtml"`
	Stock       *int     `json:"stock,omitempty" binding:"omitempty,gte=0"`
	Status      *string  `json:"status,omitempty" binding:"omitempty,oneof=active draft inactive out_of_stock"`
	Description *string  `json:"description,omitempty" binding:"omitempty,max=2000,nohtml"`
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
