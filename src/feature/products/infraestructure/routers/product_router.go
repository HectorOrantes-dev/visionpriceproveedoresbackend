package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/controllers"
)

// SetupProductRoutes registers all routes for the products feature, including metrics stubs.
func SetupProductRoutes(router *gin.RouterGroup, controller *controllers.ProductController, jwtSecret string) {
	products := router.Group("/products")
	products.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		products.POST("", controller.CreateProduct)
		products.GET("", controller.ListProducts)
		products.GET("/:id", controller.GetProduct)
		products.PUT("/:id", controller.UpdateProduct)
		products.DELETE("/:id", controller.DeleteProduct)
	}

	// Metrics stub routes (HU_PROV_05)
	metrics := router.Group("/metrics")
	metrics.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		metrics.GET("/summary", controller.GetMetricsSummary)
		metrics.GET("/top-products", controller.GetTopProducts)
	}
}
