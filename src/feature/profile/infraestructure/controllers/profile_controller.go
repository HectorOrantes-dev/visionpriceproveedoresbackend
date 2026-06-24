package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/profile/application/profile_usecase"
)

// ProfileController handles HTTP requests for the provider profile.
type ProfileController struct {
	useCase *profile_usecase.ProfileUseCase
}

// NewProfileController creates a new ProfileController.
func NewProfileController(useCase *profile_usecase.ProfileUseCase) *ProfileController {
	return &ProfileController{useCase: useCase}
}

// GetMe godoc
// @Summary      Obtener perfil del proveedor autenticado
// @Description  Devuelve los datos del proveedor de la sesión actual (nombre, RFC, correo, teléfono)
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=entities.Profile}
// @Failure      401  {object}  responses.APIResponse
// @Failure      404  {object}  responses.APIResponse
// @Router       /api/v1/auth/me [get]
func (ctrl *ProfileController) GetMe(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	profile, err := ctrl.useCase.GetProfile(c.Request.Context(), providerID.(string))
	if err != nil {
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
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Perfil obtenido", profile)
}
