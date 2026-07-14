package controllers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/core/storage"
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
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	summary, err := ctrl.useCase.GetMetricsSummary(c.Request.Context(), providerID.(string))
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Métricas obtenidas exitosamente", summary)
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
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	top, err := ctrl.useCase.GetTopProducts(c.Request.Context(), providerID.(string))
	if err != nil {
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Top de materiales obtenido exitosamente", top)
}

// UploadImage godoc
// @Summary      Subir imagen del producto
// @Description  Sube una imagen para un material/producto a Cloudflare R2 y retorna la URL pública.
// @Tags         Products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Imagen a subir (máx 5MB, jpeg/png/webp)"
// @Success      200  {object}  responses.APIResponse
// @Failure      400  {object}  responses.APIResponse
// @Failure      401  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/products/upload-image [post]
func (ctrl *ProductController) UploadImage(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "No se proporcionó ningún archivo", nil)
		return
	}

	// Validate file size (e.g. max 5MB)
	const maxSize = 5 * 1024 * 1024
	if fileHeader.Size > maxSize {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo excede el límite de 5MB", nil)
		return
	}

	// Read file content
	file, err := fileHeader.Open()
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al procesar el archivo", nil)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al leer el archivo", nil)
		return
	}

	// Determine content type
	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo debe ser una imagen válida", nil)
		return
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		// fallback extension based on content type
		if strings.Contains(contentType, "png") {
			ext = ".png"
		} else if strings.Contains(contentType, "webp") {
			ext = ".webp"
		} else {
			ext = ".jpg"
		}
	}

	// Generate a unique filename using UUID and providerID for namespace
	uniqueID := uuid.New().String()
	filename := fmt.Sprintf("products/%s/%d-%s%s", providerID.(string), time.Now().Unix(), uniqueID[:8], ext)

	// Upload to R2
	url, err := storage.UploadImage(c.Request.Context(), fileBytes, filename, contentType)
	if err != nil {
		fmt.Printf("UploadImage error: %v\n", err)
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al subir la imagen al servidor", nil)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Imagen subida exitosamente", gin.H{
		"url": url,
	})
}

// UploadProductImage godoc
// @Summary      Adjuntar imagen a un producto
// @Description  Recibe la imagen de un producto ya creado, la sube a R2 y adjunta la URL pública al producto antes de responder.
// @Tags         Products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID del producto"
// @Param        file formData  file    true  "Imagen a subir (máx 5MB, jpeg/png/webp)"
// @Success      200  {object}  responses.APIResponse
// @Failure      400  {object}  responses.APIResponse
// @Failure      401  {object}  responses.APIResponse
// @Failure      500  {object}  responses.APIResponse
// @Router       /api/v1/products/{id}/image [post]
func (ctrl *ProductController) UploadProductImage(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	productID := c.Param("id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "No se proporcionó ningún archivo", nil)
		return
	}

	const maxSize = 5 * 1024 * 1024
	if fileHeader.Size > maxSize {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo excede el límite de 5MB", nil)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al procesar el archivo", nil)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al leer el archivo", nil)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		responses.ErrorResponse(c, http.StatusBadRequest, "El archivo debe ser una imagen válida", nil)
		return
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		if strings.Contains(contentType, "png") {
			ext = ".png"
		} else if strings.Contains(contentType, "webp") {
			ext = ".webp"
		} else {
			ext = ".jpg"
		}
	}

	uniqueID := uuid.New().String()
	filename := fmt.Sprintf("products/%s/%d-%s%s", providerID.(string), time.Now().Unix(), uniqueID[:8], ext)

	pid := providerID.(string)

	url, err := storage.UploadImage(c.Request.Context(), fileBytes, filename, contentType)
	if err != nil {
		slog.Error("products: image upload failed", "error", err, "provider_id", pid, "product_id", productID)
		responses.ErrorResponse(c, http.StatusInternalServerError, "Error al subir la imagen al servidor", nil)
		return
	}

	if err := ctrl.useCase.AttachImage(c.Request.Context(), pid, productID, url); err != nil {
		slog.Error("products: failed to attach image url", "error", err, "provider_id", pid, "product_id", productID)
		handleProductError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Imagen subida exitosamente", gin.H{
		"url": url,
	})
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
