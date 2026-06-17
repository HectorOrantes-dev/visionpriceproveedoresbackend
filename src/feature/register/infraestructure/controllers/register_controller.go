package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/register/application/register_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// RegisterController handles HTTP requests for provider registration.
type RegisterController struct {
	useCase *register_usecase.RegisterUseCase
}

// NewRegisterController creates a new RegisterController.
func NewRegisterController(useCase *register_usecase.RegisterUseCase) *RegisterController {
	return &RegisterController{useCase: useCase}
}

// Register godoc
// @Summary      Registrar un nuevo proveedor
// @Description  Crea una nueva cuenta de proveedor con nombre comercial, RFC, correo, teléfono y contraseña
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      entities.RegisterRequest  true  "Datos de registro"
// @Success      201   {object}  responses.APIResponse{data=entities.RegisterResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      409   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/register [post]
func (ctrl *RegisterController) Register(c *gin.Context) {
	var req entities.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de registro inválidos", nil)
		return
	}

	result, err := ctrl.useCase.Execute(c.Request.Context(), &req)
	if err != nil {
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) {
			switch {
			case errors.Is(domainErr.Base, domainErrors.ErrConflict):
				responses.ErrorResponse(c, http.StatusConflict, domainErr.Message, nil)
				return
			case errors.Is(domainErr.Base, domainErrors.ErrValidation):
				responses.ErrorResponse(c, http.StatusBadRequest, domainErr.Message, nil)
				return
			}
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusCreated, "Proveedor registrado exitosamente", result)
}
