package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/ml/application/ml_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/ml/domain/entities"
)

// MLController handles HTTP requests for Machine Learning operations.
type MLController struct {
	useCase *ml_usecase.MLUseCase
}

// NewMLController creates a new MLController.
func NewMLController(useCase *ml_usecase.MLUseCase) *MLController {
	return &MLController{useCase: useCase}
}

// ClassifyProduct godoc
// @Summary      Clasificar producto automáticamente
// @Description  Utiliza TF-IDF y Similitud Coseno para mapear un nombre de producto al catálogo maestro
// @Tags         Machine Learning
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.ClassifyProductRequest  true  "Nombre del producto"
// @Success      200   {object}  responses.APIResponse{data=entities.ClassifyProductResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      404   {object}  responses.APIResponse
// @Router       /api/v1/ml/products/classify [post]
func (ctrl *MLController) ClassifyProduct(c *gin.Context) {
	var req entities.ClassifyProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	result, err := ctrl.useCase.ClassifyProduct(c.Request.Context(), &req)
	if err != nil {
		handleMLError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Producto clasificado exitosamente", result)
}

// DetectDuplicates godoc
// @Summary      Detectar productos duplicados
// @Description  Encuentra posibles productos duplicados en el catálogo del proveedor logueado usando TF-IDF
// @Tags         Machine Learning
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  responses.APIResponse{data=[]entities.DuplicatePair}
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/ml/products/duplicates [get]
func (ctrl *MLController) DetectDuplicates(c *gin.Context) {
	providerIDStr, _ := c.Get("provider_id")
	providerID, err := uuid.Parse(providerIDStr.(string))
	if err != nil {
		responses.ErrorResponse(c, http.StatusUnauthorized, "ID de proveedor inválido en el token", nil)
		return
	}

	duplicates, err := ctrl.useCase.DetectDuplicates(c.Request.Context(), providerID)
	if err != nil {
		handleMLError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Análisis de duplicados completado", duplicates)
}

// DetectAnomalies godoc
// @Summary      Detectar anomalías en precios
// @Description  Utiliza Isolation Forest comparando con el mercado global para detectar precios anómalos
// @Tags         Machine Learning
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  responses.APIResponse{data=[]entities.AnomalyItem}
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/ml/products/anomalies [get]
func (ctrl *MLController) DetectAnomalies(c *gin.Context) {
	providerIDStr, _ := c.Get("provider_id")
	providerID, err := uuid.Parse(providerIDStr.(string))
	if err != nil {
		responses.ErrorResponse(c, http.StatusUnauthorized, "ID de proveedor inválido en el token", nil)
		return
	}

	anomalies, err := ctrl.useCase.DetectAnomalies(c.Request.Context(), providerID)
	if err != nil {
		handleMLError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Análisis de anomalías de precio completado", anomalies)
}

func handleMLError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Base, domainErrors.ErrValidation):
			responses.ErrorResponse(c, http.StatusBadRequest, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrNotFound):
			responses.ErrorResponse(c, http.StatusNotFound, domainErr.Message, nil)
			return
		}
	}
	responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor en módulo ML", nil)
}
