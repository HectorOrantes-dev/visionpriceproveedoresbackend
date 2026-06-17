package controllers

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/application/extraction_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/domain/entities"
)

// maxUploadSize is the maximum allowed file size for Excel uploads (5 MB).
const maxUploadSize = 5 << 20 // 5 MB

// ExtractionController handles HTTP requests for Excel import operations.
type ExtractionController struct {
	useCase *extraction_usecase.ExtractionUseCase
}

// NewExtractionController creates a new ExtractionController.
func NewExtractionController(useCase *extraction_usecase.ExtractionUseCase) *ExtractionController {
	return &ExtractionController{useCase: useCase}
}

// DetectColumns godoc
// @Summary      Detectar columnas de archivo Excel
// @Description  Sube un archivo Excel y retorna las columnas detectadas en los headers
// @Tags         Extracciones
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Archivo Excel (.xlsx)"
// @Success      200   {object}  responses.APIResponse{data=entities.DetectColumnsResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Router       /api/v1/extractions/detect-columns [post]
func (ctrl *ExtractionController) DetectColumns(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Archivo no proporcionado", nil)
		return
	}
	defer file.Close()

	// Validate file extension
	if !isValidExcelExtension(header.Filename) {
		responses.ErrorResponse(c, http.StatusBadRequest, "Solo se permiten archivos Excel (.xlsx)", nil)
		return
	}

	// Validate file size
	if header.Size > maxUploadSize {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo excede el tamaño máximo permitido (5 MB)", nil)
		return
	}

	result, err := ctrl.useCase.DetectColumns(file)
	if err != nil {
		handleExtractionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Columnas detectadas exitosamente", result)
}

// SaveMapping godoc
// @Summary      Guardar mapeo de columnas
// @Description  Guarda la correspondencia entre columnas del Excel y campos del producto
// @Tags         Extracciones
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.SaveMappingRequest  true  "Mapeo de columnas"
// @Success      200   {object}  responses.APIResponse{data=entities.ImportMapping}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Router       /api/v1/extractions/mapping [post]
func (ctrl *ExtractionController) SaveMapping(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	var req entities.SaveMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de mapeo inválidos", nil)
		return
	}

	result, err := ctrl.useCase.SaveMapping(c.Request.Context(), providerID.(string), &req)
	if err != nil {
		handleExtractionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Mapeo guardado exitosamente", result)
}

// GetMapping godoc
// @Summary      Obtener mapeo de columnas guardado
// @Description  Retorna el mapeo de columnas configurado por el proveedor
// @Tags         Extracciones
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=entities.ImportMapping}
// @Failure      401  {object}  responses.APIResponse
// @Failure      404  {object}  responses.APIResponse
// @Router       /api/v1/extractions/mapping [get]
func (ctrl *ExtractionController) GetMapping(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	result, err := ctrl.useCase.GetMapping(c.Request.Context(), providerID.(string))
	if err != nil {
		handleExtractionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Mapeo obtenido exitosamente", result)
}

// ProcessImport godoc
// @Summary      Importar productos desde Excel
// @Description  Procesa un archivo Excel usando el mapeo guardado y crea/actualiza productos
// @Tags         Extracciones
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Archivo Excel (.xlsx)"
// @Success      200   {object}  responses.APIResponse{data=entities.ImportSummary}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      404   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/extractions/import [post]
func (ctrl *ExtractionController) ProcessImport(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Archivo no proporcionado", nil)
		return
	}
	defer file.Close()

	// Validate file extension
	if !isValidExcelExtension(header.Filename) {
		responses.ErrorResponse(c, http.StatusBadRequest, "Solo se permiten archivos Excel (.xlsx)", nil)
		return
	}

	// Validate file size
	if header.Size > maxUploadSize {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo excede el tamaño máximo permitido (5 MB)", nil)
		return
	}

	result, err := ctrl.useCase.ProcessImport(c.Request.Context(), providerID.(string), file)
	if err != nil {
		handleExtractionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Importación completada", result)
}

// isValidExcelExtension checks if the file has a valid Excel extension.
func isValidExcelExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".xlsx"
}

func handleExtractionError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Base, domainErrors.ErrNotFound):
			responses.ErrorResponse(c, http.StatusNotFound, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrValidation):
			responses.ErrorResponse(c, http.StatusBadRequest, domainErr.Message, nil)
			return
		}
	}
	responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
}
