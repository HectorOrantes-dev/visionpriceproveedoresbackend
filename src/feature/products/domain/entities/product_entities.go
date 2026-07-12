package entities

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a material/product offered by a provider.
type Product struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Name       string    `json:"name"`
	SKU        *string   `json:"sku,omitempty"`
	Brand      *string   `json:"brand,omitempty"`
	Price      float64   `json:"price"`
	Unit       string    `json:"unit"`
	Category   string    `json:"category"`
	// RendimientoM2 is the m² covered by one sales unit — whatever that unit is
	// (caja, saco, cubeta, m³…). It's the universal conversion factor.
	RendimientoM2 float64 `json:"rendimiento_m2"`
	// Piece geometry, used by the kit/crucetas engine for piso/azulejo/zoclo.
	PiezaLargoM float64 `json:"pieza_largo_m"`
	PiezaAnchoM float64 `json:"pieza_ancho_m"`
	// PiezasPorPaquete is pieces per package (crucetas, etc.).
	PiezasPorPaquete int       `json:"piezas_por_paquete"`
	Stock            int       `json:"stock"`
	Status           string    `json:"status"`
	Description      *string   `json:"description,omitempty"`
	ImageURL         *string   `json:"image_url,omitempty"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateProductRequest is the DTO for creating a new product.
// String fields carry the "nohtml" validator as a backend anti-XSS defense.
type CreateProductRequest struct {
	Name             string   `json:"name" binding:"required,min=1,max=255,nohtml"`
	SKU              *string  `json:"sku,omitempty" binding:"omitempty,max=100,nohtml"`
	Brand            *string  `json:"brand,omitempty" binding:"omitempty,max=255,nohtml"`
	Price            float64  `json:"price" binding:"required,gt=0"`
	Unit             string   `json:"unit" binding:"required,min=1,max=50,nohtml"`
	Category         string   `json:"category" binding:"required,min=1,max=100,nohtml"`
	RendimientoM2    *float64 `json:"rendimiento_m2,omitempty" binding:"omitempty,gte=0"`
	PiezaLargoM      *float64 `json:"pieza_largo_m,omitempty" binding:"omitempty,gte=0"`
	PiezaAnchoM      *float64 `json:"pieza_ancho_m,omitempty" binding:"omitempty,gte=0"`
	PiezasPorPaquete *int     `json:"piezas_por_paquete,omitempty" binding:"omitempty,gte=0"`
	Stock            *int     `json:"stock,omitempty" binding:"omitempty,gte=0"`
	Status           *string  `json:"status,omitempty" binding:"omitempty,oneof=active draft inactive out_of_stock"`
	Description      *string  `json:"description,omitempty" binding:"omitempty,max=2000,nohtml"`
	ImageURL         *string  `json:"image_url,omitempty" binding:"omitempty,url"`
}

// UpdateProductRequest is the DTO for updating a product.
type UpdateProductRequest struct {
	Name             *string  `json:"name,omitempty" binding:"omitempty,min=1,max=255,nohtml"`
	SKU              *string  `json:"sku,omitempty" binding:"omitempty,max=100,nohtml"`
	Brand            *string  `json:"brand,omitempty" binding:"omitempty,max=255,nohtml"`
	Price            *float64 `json:"price,omitempty" binding:"omitempty,gt=0"`
	Unit             *string  `json:"unit,omitempty" binding:"omitempty,min=1,max=50,nohtml"`
	Category         *string  `json:"category,omitempty" binding:"omitempty,min=1,max=100,nohtml"`
	RendimientoM2    *float64 `json:"rendimiento_m2,omitempty" binding:"omitempty,gte=0"`
	PiezaLargoM      *float64 `json:"pieza_largo_m,omitempty" binding:"omitempty,gte=0"`
	PiezaAnchoM      *float64 `json:"pieza_ancho_m,omitempty" binding:"omitempty,gte=0"`
	PiezasPorPaquete *int     `json:"piezas_por_paquete,omitempty" binding:"omitempty,gte=0"`
	Stock            *int     `json:"stock,omitempty" binding:"omitempty,gte=0"`
	Status           *string  `json:"status,omitempty" binding:"omitempty,oneof=active draft inactive out_of_stock"`
	Description      *string  `json:"description,omitempty" binding:"omitempty,max=2000,nohtml"`
	ImageURL         *string  `json:"image_url,omitempty" binding:"omitempty,url"`
}

// MetricsSummary aggregates catalog/inventory metrics for a provider, computed
// from the products table. There is no sales/orders data in the system, so these
// reflect the current catalog (inventory value, stock, price, distribution).
type MetricsSummary struct {
	InventoryValue float64         `json:"inventoryValue"` // Σ(price × stock) of active products
	UnitsInStock   int             `json:"unitsInStock"`   // Σ(stock)
	AveragePrice   float64         `json:"averagePrice"`   // AVG(price)
	TotalMaterials int             `json:"totalMaterials"` // COUNT(active)
	MonthlyNew     []MonthlyPoint  `json:"monthlyNew"`     // materials created per month (last 6)
	Distribution   []CategorySlice `json:"distribution"`   // inventory value share by category
}

// MonthlyPoint is one bar of the "materials created per month" chart.
type MonthlyPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// CategorySlice is one slice of the category distribution donut.
type CategorySlice struct {
	Category string  `json:"category"`
	Value    float64 `json:"value"`
	Share    float64 `json:"share"` // percentage 0..100 of total inventory value
}

// TopProduct is a top material by inventory value (price × stock).
type TopProduct struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"` // inventory value of this material
	Share  float64 `json:"share"`  // percentage 0..100 of total inventory value
}
