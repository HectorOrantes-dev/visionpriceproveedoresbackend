package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/login/application/login_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/login/domain/entities"
)

// LoginController handles HTTP requests for authentication.
type LoginController struct {
	useCase *login_usecase.LoginUseCase
}

// NewLoginController creates a new LoginController.
func NewLoginController(useCase *login_usecase.LoginUseCase) *LoginController {
	return &LoginController{useCase: useCase}
}

// Login godoc
// @Summary      Iniciar sesión
// @Description  Autentica un proveedor con correo y contraseña. Retorna un token temporal para el flujo de 2FA.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      entities.LoginRequest  true  "Credenciales de acceso"
// @Success      200   {object}  responses.APIResponse{data=entities.LoginResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/login [post]
func (ctrl *LoginController) Login(c *gin.Context) {
	var req entities.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de acceso inválidos", nil)
		return
	}

	tempToken, err := ctrl.useCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) {
			if errors.Is(domainErr.Base, domainErrors.ErrUnauthorized) {
				responses.ErrorResponse(c, http.StatusUnauthorized, domainErr.Message, nil)
				return
			}
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Autenticación exitosa. Se requiere verificación 2FA.", entities.LoginResponse{
		TempToken:   tempToken,
		Requires2FA: true,
	})
}

// ForgotPassword godoc
// @Summary      Solicitar recuperación de contraseña
// @Description  Envía un correo con un enlace para restablecer la contraseña (stub: se registra en logs)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      entities.ForgotPasswordRequest  true  "Correo electrónico"
// @Success      200   {object}  responses.APIResponse
// @Failure      400   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/forgot-password [post]
func (ctrl *LoginController) ForgotPassword(c *gin.Context) {
	var req entities.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Correo electrónico inválido", nil)
		return
	}

	if err := ctrl.useCase.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	// Always return success to not leak email existence
	responses.SuccessResponse(c, http.StatusOK, "Si el correo existe, se enviará un enlace de recuperación", nil)
}

// ResetPassword godoc
// @Summary      Restablecer contraseña
// @Description  Restablece la contraseña usando el token de recuperación
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      entities.ResetPasswordRequest  true  "Token y nueva contraseña"
// @Success      200   {object}  responses.APIResponse
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/reset-password [post]
func (ctrl *LoginController) ResetPassword(c *gin.Context) {
	var req entities.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	if err := ctrl.useCase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) {
			if errors.Is(domainErr.Base, domainErrors.ErrUnauthorized) {
				responses.ErrorResponse(c, http.StatusUnauthorized, domainErr.Message, nil)
				return
			}
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Contraseña actualizada exitosamente", nil)
}

// Logout godoc
// @Summary      Cerrar sesión
// @Description  Revoca el refresh token proporcionado para impedir su uso futuro
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.LogoutRequest  true  "Refresh token a revocar"
// @Success      200   {object}  responses.APIResponse
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/logout [post]
func (ctrl *LoginController) Logout(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	var req entities.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	if err := ctrl.useCase.Logout(c.Request.Context(), providerID.(string), req.RefreshToken); err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Sesión cerrada exitosamente", nil)
}

// Refresh godoc
// @Summary      Refrescar access token
// @Description  Emite un nuevo access token usando un refresh token válido (verificando que no esté revocado)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      entities.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  responses.APIResponse{data=map[string]string}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/refresh [post]
func (ctrl *LoginController) Refresh(c *gin.Context) {
	var req entities.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	accessToken, err := ctrl.useCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) {
			if errors.Is(domainErr.Base, domainErrors.ErrUnauthorized) {
				responses.ErrorResponse(c, http.StatusUnauthorized, domainErr.Message, nil)
				return
			}
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Token renovado exitosamente", map[string]string{
		"access_token": accessToken,
	})
}

