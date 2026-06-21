package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/products/application/product_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// ProductController handles HTTP requests for product operations.
type ProductController struct {
	useCase *product_usecase.ProductUseCase
}

// NewProductController creates a new ProductController.
func NewProductController(useCase *product_usecase.ProductUseCase) *ProductController {
	return &ProductController{useCase: useCase}
}

// CreateProduct godoc
// @Summary      Crear un nuevo producto
// @Description  Registra un nuevo material/producto para el proveedor autenticado
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.CreateProductRequest  true  "Datos del producto"
// @Success      201   {object}  responses.APIResponse{data=entities.Product}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      500   {object}  responses.APIResponse
// @Router       /api/v1/products [post]
func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	var req entities.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos del producto inválidos", nil)
		return
	}

	result, err := ctrl.useCase.CreateProduct(c.Request.Context(), providerID.(string), &req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusCreated, "Producto creado exitosamente", result)
}

// ListProducts godoc
// @Summary      Listar productos activos
// @Description  Retorna todos los productos activos del proveedor autenticado
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  responses.APIResponse{data=[]entities.Product}
// @Failure      401  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/products [get]
func (ctrl *ProductController) ListProducts(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	result, err := ctrl.useCase.ListProducts(c.Request.Context(), providerID.(string))
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Productos obtenidos exitosamente", result)
}

// GetProduct godoc
// @Summary      Obtener producto por ID
// @Description  Retorna un producto específico del proveedor autenticado
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID del producto"
// @Success      200  {object}  responses.APIResponse{data=entities.Product}
// @Failure      401  {object}  responses.APIResponse
// @Failure      404  {object}  responses.APIResponse
// @Router       /api/v1/products/{id} [get]
func (ctrl *ProductController) GetProduct(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	productID := c.Param("id")

	result, err := ctrl.useCase.GetProduct(c.Request.Context(), providerID.(string), productID)
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Producto obtenido exitosamente", result)
}

// UpdateProduct godoc
// @Summary      Actualizar producto
// @Description  Actualiza los datos de un producto existente del proveedor autenticado
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                        true  "ID del producto"
// @Param        body  body      entities.UpdateProductRequest  true  "Datos a actualizar"
// @Success      200   {object}  responses.APIResponse{data=entities.Product}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      404   {object}  responses.APIResponse
// @Router       /api/v1/products/{id} [put]
func (ctrl *ProductController) UpdateProduct(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	productID := c.Param("id")

	var req entities.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	result, err := ctrl.useCase.UpdateProduct(c.Request.Context(), providerID.(string), productID, &req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Producto actualizado exitosamente", result)
}

// DeleteProduct godoc
// @Summary      Eliminar producto (baja lógica)
// @Description  Desactiva un producto del proveedor autenticado (soft delete)
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID del producto"
// @Success      200  {object}  responses.APIResponse
// @Failure      401  {object}  responses.APIResponse
// @Failure      404  {object}  responses.APIResponse
// @Router       /api/v1/products/{id} [delete]
func (ctrl *ProductController) DeleteProduct(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	productID := c.Param("id")

	if err := ctrl.useCase.DeleteProduct(c.Request.Context(), providerID.(string), productID); err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Producto eliminado exitosamente", nil)
}

// GetMetricsSummary godoc
// @Summary      Obtener resumen de métricas (stub)
// @Description  Retorna métricas del proveedor — endpoint stub para desarrollo futuro
// @Tags         Metrics
// @Produce      json
// @Security     BearerAuth
// @Param        period  query     string  false  "Periodo: week o month"  default(month)
// @Success      200     {object}  responses.APIResponse{data=entities.MetricsSummary}
// @Failure      401     {object}  responses.APIResponse
// @Router       /api/v1/metrics/summary [get]
func (ctrl *ProductController) GetMetricsSummary(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	// Stub: return zeroed metrics
	stub := entities.MetricsSummary{
		TotalBudgetAppearances: 0,
		ContactClicks:          0,
		Period:                 period,
	}

	responses.SuccessResponse(c, http.StatusOK, "Métricas obtenidas (stub)", stub)
}

// GetTopProducts godoc
// @Summary      Obtener top productos cotizados (stub)
// @Description  Retorna los top 3 productos más cotizados — endpoint stub para desarrollo futuro
// @Tags         Metrics
// @Produce      json
// @Security     BearerAuth
// @Param        period  query     string  false  "Periodo: week o month"  default(month)
// @Success      200     {object}  responses.APIResponse{data=[]entities.TopProduct}
// @Failure      401     {object}  responses.APIResponse
// @Router       /api/v1/metrics/top-products [get]
func (ctrl *ProductController) GetTopProducts(c *gin.Context) {
	// Stub: return empty list
	responses.SuccessResponse(c, http.StatusOK, "Top productos obtenidos (stub)", []entities.TopProduct{})
}

func handleProductError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Base, domainErrors.ErrNotFound):
			responses.ErrorResponse(c, http.StatusNotFound, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrValidation):
			responses.ErrorResponse(c, http.StatusBadRequest, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrUnauthorized):
			responses.ErrorResponse(c, http.StatusForbidden, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrPaymentRequired):
			responses.ErrorResponse(c, http.StatusPaymentRequired, domainErr.Message, nil)
			return
		}
	}
	responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
}
