package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/application/geolocation_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain/entities"

)

// GeolocationController handles HTTP requests for geolocation operations.
type GeolocationController struct {
	useCase *geolocation_usecase.GeolocationUseCase
}

// NewGeolocationController creates a new GeolocationController.
func NewGeolocationController(useCase *geolocation_usecase.GeolocationUseCase) *GeolocationController {
	return &GeolocationController{useCase: useCase}
}

// SetLocation godoc
// @Summary      Fijar o actualizar dirección del almacén
// @Description  Guarda la dirección del almacén o sucursal del proveedor
// @Tags         Geolocations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.SetLocationRequest  true  "Dirección del almacén"
// @Success      200   {object}  responses.APIResponse{data=entities.ProviderLocation}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/providers/location [put]
func (ctrl *GeolocationController) SetLocation(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	var req entities.SetLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de ubicación inválidos", nil)
		return
	}

	result, err := ctrl.useCase.SetLocation(c.Request.Context(), providerID.(string), &req)
	if err != nil {
		handleGeolocationError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Ubicación actualizada exitosamente", result)
}

// GetLocation godoc
// @Summary      Obtener dirección del almacén
// @Description  Retorna la dirección del almacén o sucursal registrada por el proveedor
// @Tags         Geolocations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=entities.ProviderLocation}
// @Failure      401  {object}  responses.APIResponse
// @Failure      404  {object}  responses.APIResponse
// @Router       /api/v1/providers/location [get]
func (ctrl *GeolocationController) GetLocation(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	result, err := ctrl.useCase.GetLocation(c.Request.Context(), providerID.(string))
	if err != nil {
		handleGeolocationError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Ubicación obtenida exitosamente", result)
}

func handleGeolocationError(c *gin.Context, err error) {
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
