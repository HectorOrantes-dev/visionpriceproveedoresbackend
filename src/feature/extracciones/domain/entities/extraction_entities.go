package entities

import (
	"time"

	"github.com/google/uuid"
)

// ImportMapping represents the column mapping configuration for Excel imports.
type ImportMapping struct {
	ID         uuid.UUID         `json:"id"`
	ProviderID uuid.UUID         `json:"provider_id"`
	ColumnMap  map[string]string `json:"column_map"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// SaveMappingRequest is the DTO for saving column mappings.
// The keys are target fields (name, price, unit, category, description, sku)
// and the values are the Excel column names detected from headers.
type SaveMappingRequest struct {
	ColumnMap map[string]string `json:"column_map" binding:"required"`
}

// DetectColumnsResponse is returned after detecting Excel headers.
type DetectColumnsResponse struct {
	Columns []string `json:"columns"`
}

// ImportSummary is the result of processing an Excel import.
type ImportSummary struct {
	TotalRows       int           `json:"total_rows"`
	NewProducts     int           `json:"new_products"`
	UpdatedProducts int           `json:"updated_products"`
	SkippedRows     int           `json:"skipped_rows"`
	Errors          []ImportError `json:"errors"`
}

// ImportError represents a validation error in a specific row during import.
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ImportProduct is a product extracted from the Excel file, ready for persistence.
type ImportProduct struct {
	Name     string
	Price    float64
	Unit     string
	Category string
	SKU      string
}
