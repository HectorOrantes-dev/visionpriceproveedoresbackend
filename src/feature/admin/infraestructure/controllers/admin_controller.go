package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/admin/application/admin_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// AdminController handles HTTP requests for admin operations.
type AdminController struct {
	useCase              *admin_usecase.AdminUseCase
	jwtSecret            string
	jwtExpirationMinutes int
}

// NewAdminController creates a new AdminController.
func NewAdminController(useCase *admin_usecase.AdminUseCase, jwtSecret string, jwtExpirationMinutes int) *AdminController {
	return &AdminController{
		useCase:              useCase,
		jwtSecret:            jwtSecret,
		jwtExpirationMinutes: jwtExpirationMinutes,
	}
}

// AdminLogin godoc
// @Summary      Login de administrador del sistema
// @Description  Autentica un usuario administrador y devuelve un JWT con rol USER_SYS_ADMIN
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        body  body      entities.AdminLoginRequest  true  "Credenciales de admin"
// @Success      200   {object}  responses.APIResponse{data=entities.AdminLoginResponse}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Router       /api/v1/admin/login [post]
func (ctrl *AdminController) AdminLogin(c *gin.Context) {
	var req entities.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de login inválidos", nil)
		return
	}

	user, err := ctrl.useCase.AdminLogin(c.Request.Context(), &req)
	if err != nil {
		handleAdminError(c, err)
		return
	}

	// Generate JWT with role claim
	duration := time.Duration(ctrl.jwtExpirationMinutes) * time.Minute
	token, err := middleware.GenerateTokenWithRole(user.ID.String(), middleware.TokenTypeAccess, user.Role, ctrl.jwtSecret, duration)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al generar el token", nil)
		return
	}

	loginResponse := entities.AdminLoginResponse{
		Token: token,
		Role:  user.Role,
	}

	responses.SuccessResponse(c, http.StatusOK, "Login exitoso", loginResponse)
}

// GetMetrics godoc
// @Summary      Obtener métricas globales del sistema (HU_SYS_01)
// @Description  Retorna conteo de constructores, proveedores y usuarios activos en las últimas 24h
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=entities.GlobalMetrics}
// @Failure      401  {object}  responses.APIResponse
// @Failure      403  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/admin/metrics [get]
func (ctrl *AdminController) GetMetrics(c *gin.Context) {
	result, err := ctrl.useCase.GetGlobalMetrics(c.Request.Context())
	if err != nil {
		handleAdminError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Métricas globales obtenidas exitosamente", result)
}

// GetProviderMapPins godoc
// @Summary      Obtener pines de proveedores para el mapa (HU_SYS_02)
// @Description  Retorna ubicaciones de proveedores activos con datos seguros (sin RFC, email, precios)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=[]entities.ProviderMapPin}
// @Failure      401  {object}  responses.APIResponse
// @Failure      403  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/admin/geography/providers [get]
func (ctrl *AdminController) GetProviderMapPins(c *gin.Context) {
	result, err := ctrl.useCase.GetProviderMapPins(c.Request.Context())
	if err != nil {
		handleAdminError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Pines de proveedores obtenidos exitosamente", result)
}

func handleAdminError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Base, domainErrors.ErrUnauthorized):
			responses.ErrorResponse(c, http.StatusUnauthorized, domainErr.Message, nil)
			return
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
