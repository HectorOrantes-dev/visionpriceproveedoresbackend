package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/application/subscription_usecase"
)

// SubscriptionController handles HTTP requests for subscription operations.
type SubscriptionController struct {
	useCase *subscription_usecase.SubscriptionUseCase
}

// NewSubscriptionController creates a new SubscriptionController.
func NewSubscriptionController(useCase *subscription_usecase.SubscriptionUseCase) *SubscriptionController {
	return &SubscriptionController{useCase: useCase}
}

// GetCurrent godoc
// @Summary      Suscripción actual del proveedor
// @Description  Retorna el plan vigente, su estado y el uso de productos (X/límite)
// @Tags         Subscriptions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=entities.SubscriptionView}
// @Failure      401  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/subscription [get]
func (ctrl *SubscriptionController) GetCurrent(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	result, err := ctrl.useCase.GetCurrent(c.Request.Context(), providerID.(string))
	if err != nil {
		handleSubscriptionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Suscripción obtenida exitosamente", result)
}

// ListPlans godoc
// @Summary      Catálogo de planes
// @Description  Retorna los planes de suscripción disponibles (free, pro, max)
// @Tags         Subscriptions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=[]entities.Plan}
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/plans [get]
func (ctrl *SubscriptionController) ListPlans(c *gin.Context) {
	result, err := ctrl.useCase.ListPlans(c.Request.Context())
	if err != nil {
		handleSubscriptionError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Planes obtenidos exitosamente", result)
}

func handleSubscriptionError(c *gin.Context, err error) {
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
	responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
}
