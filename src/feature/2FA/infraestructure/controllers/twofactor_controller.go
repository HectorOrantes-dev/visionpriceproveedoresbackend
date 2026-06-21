package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/application/twofactor_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/domain/entities"
)

// TwoFactorController handles HTTP requests for 2FA operations.
type TwoFactorController struct {
	useCase *twofactor_usecase.TwoFactorUseCase
}

// NewTwoFactorController creates a new TwoFactorController.
func NewTwoFactorController(useCase *twofactor_usecase.TwoFactorUseCase) *TwoFactorController {
	return &TwoFactorController{useCase: useCase}
}

// GenerateOTP godoc
// @Summary      Generar código OTP
// @Description  Genera y envía un código OTP de 6 dígitos al proveedor (requiere token temporal de login)
// @Tags         2FA
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse
// @Failure      401  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/auth/2fa/generate [post]
func (ctrl *TwoFactorController) GenerateOTP(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	if err := ctrl.useCase.GenerateOTP(c.Request.Context(), providerID.(string)); err != nil {
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) {
			responses.ErrorResponse(c, http.StatusInternalServerError, domainErr.Message, nil)
			return
		}
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Código OTP generado y enviado", nil)
}

// VerifyOTP godoc
// @Summary      Verificar código OTP
// @Description  Valida el código OTP y emite tokens de acceso y refresh
// @Tags         2FA
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.VerifyOTPRequest  true  "Código OTP"
// @Success      200   {object}  responses.APIResponse{data=entities.VerifyOTPResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/auth/2fa/verify [post]
func (ctrl *TwoFactorController) VerifyOTP(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	var req entities.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Código OTP inválido", nil)
		return
	}

	accessToken, refreshToken, csrfToken, err := ctrl.useCase.VerifyOTP(c.Request.Context(), providerID.(string), req.Code)
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

	responses.SuccessResponse(c, http.StatusOK, "Verificación 2FA exitosa", entities.VerifyOTPResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CSRFToken:    csrfToken,
	})
}
