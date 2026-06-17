package extraction_usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/domain"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/domain/entities"
	productEntities "github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

const maxRows = 1000

// ExtractionUseCase contains business logic for Excel import operations.
type ExtractionUseCase struct {
	repo domain.ExtractionRepository
}

// NewExtractionUseCase creates a new ExtractionUseCase.
func NewExtractionUseCase(repo domain.ExtractionRepository) *ExtractionUseCase {
	return &ExtractionUseCase{repo: repo}
}

// DetectColumns reads the headers from an uploaded Excel file and returns the column names.
func (uc *ExtractionUseCase) DetectColumns(file io.Reader) (*entities.DetectColumnsResponse, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "No se pudo leer el archivo Excel")
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El archivo Excel no contiene hojas")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) == 0 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El archivo Excel está vacío")
	}

	// First row = headers
	headers := rows[0]
	var columns []string
	for _, h := range headers {
		trimmed := strings.TrimSpace(h)
		if trimmed != "" {
			columns = append(columns, trimmed)
		}
	}

	if len(columns) == 0 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "No se detectaron columnas en el archivo")
	}

	return &entities.DetectColumnsResponse{Columns: columns}, nil
}

// SaveMapping saves the column mapping for a provider.
func (uc *ExtractionUseCase) SaveMapping(ctx context.Context, providerID string, req *entities.SaveMappingRequest) (*entities.ImportMapping, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	// Validate that at least name and price mappings exist
	if _, ok := req.ColumnMap["name"]; !ok {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El mapeo debe incluir al menos el campo 'name'")
	}
	if _, ok := req.ColumnMap["price"]; !ok {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El mapeo debe incluir al menos el campo 'price'")
	}

	mapping := &entities.ImportMapping{
		ProviderID: pid,
		ColumnMap:  req.ColumnMap,
	}

	result, err := uc.repo.SaveMapping(ctx, mapping)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al guardar el mapeo")
	}

	return result, nil
}

// GetMapping retrieves the saved column mapping for a provider.
func (uc *ExtractionUseCase) GetMapping(ctx context.Context, providerID string) (*entities.ImportMapping, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	mapping, err := uc.repo.GetMapping(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "No se encontró mapeo de columnas. Suba un archivo primero.")
	}

	return mapping, nil
}

// ProcessImport reads the Excel file using the saved mapping and bulk upserts products.
func (uc *ExtractionUseCase) ProcessImport(ctx context.Context, providerID string, file io.Reader) (*entities.ImportSummary, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	// Get the saved mapping
	mapping, err := uc.repo.GetMapping(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "No se encontró mapeo de columnas. Configure el mapeo primero.")
	}

	// Read the Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "No se pudo leer el archivo Excel")
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El archivo Excel no contiene hojas")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "El archivo Excel está vacío o no tiene datos")
	}

	// Build a header-to-index map
	headerIndex := make(map[string]int)
	for i, h := range rows[0] {
		headerIndex[strings.TrimSpace(h)] = i
	}

	// Map target fields to column indices
	fieldIndex := make(map[string]int)
	for targetField, excelColumn := range mapping.ColumnMap {
		if idx, ok := headerIndex[excelColumn]; ok {
			fieldIndex[targetField] = idx
		}
	}

	// Process data rows (skip header, limit to maxRows)
	var products []*productEntities.Product
	var importErrors []entities.ImportError
	dataRows := rows[1:]
	if len(dataRows) > maxRows {
		dataRows = dataRows[:maxRows]
	}

	for i, row := range dataRows {
		rowNum := i + 2 // 1-indexed, +1 for header

		product, rowErrors := uc.parseRow(row, fieldIndex, rowNum, pid)
		if len(rowErrors) > 0 {
			importErrors = append(importErrors, rowErrors...)
			continue
		}

		if product != nil {
			products = append(products, product)
		}
	}

	// Bulk upsert valid products
	newCount, updatedCount := 0, 0
	if len(products) > 0 {
		newCount, updatedCount, err = uc.repo.BulkUpsertProducts(ctx, pid, products)
		if err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al importar productos")
		}
	}

	summary := &entities.ImportSummary{
		TotalRows:       len(dataRows),
		NewProducts:     newCount,
		UpdatedProducts: updatedCount,
		SkippedRows:     len(dataRows) - len(products),
		Errors:          importErrors,
	}

	if summary.Errors == nil {
		summary.Errors = []entities.ImportError{}
	}

	// Audit log: bulk import completed
	slog.Info("AUDIT: bulk import completed",
		"provider_id", providerID,
		"total_rows", summary.TotalRows,
		"new_products", summary.NewProducts,
		"updated_products", summary.UpdatedProducts,
		"skipped_rows", summary.SkippedRows,
		"error_count", len(summary.Errors),
	)

	return summary, nil
}

// parseRow extracts a product from a single Excel row using the field-to-column mapping.
func (uc *ExtractionUseCase) parseRow(row []string, fieldIndex map[string]int, rowNum int, providerID uuid.UUID) (*productEntities.Product, []entities.ImportError) {
	var errs []entities.ImportError

	getCell := func(field string) string {
		if idx, ok := fieldIndex[field]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	name := getCell("name")
	if name == "" {
		errs = append(errs, entities.ImportError{Row: rowNum, Field: "name", Message: "Nombre del producto vacío"})
		return nil, errs
	}

	priceStr := getCell("price")
	if priceStr == "" {
		errs = append(errs, entities.ImportError{Row: rowNum, Field: "price", Message: "Precio vacío"})
		return nil, errs
	}

	// Clean price string: remove currency symbols, commas, spaces
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	priceStr = strings.TrimSpace(priceStr)

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		errs = append(errs, entities.ImportError{
			Row:     rowNum,
			Field:   "price",
			Message: fmt.Sprintf("Precio no es un número válido: '%s'", priceStr),
		})
		return nil, errs
	}

	if price <= 0 {
		errs = append(errs, entities.ImportError{Row: rowNum, Field: "price", Message: "El precio debe ser mayor a 0"})
		return nil, errs
	}

	unit := getCell("unit")
	if unit == "" {
		unit = "pieza" // default unit
	}

	category := getCell("category")
	if category == "" {
		category = "General" // default category
	}

	description := getCell("description")
	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	return &productEntities.Product{
		ProviderID:  providerID,
		Name:        name,
		Price:       price,
		Unit:        unit,
		Category:    category,
		Description: descPtr,
		Active:      true,
	}, nil
}
