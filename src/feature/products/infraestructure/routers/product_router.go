package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/controllers"
)

// SetupProductRoutes registers all routes for the products feature, including metrics stubs.
func SetupProductRoutes(router *gin.RouterGroup, controller *controllers.ProductController, db *pgxpool.Pool, csrfManager *csrf.Manager, jwtSecret string) {
	products := router.Group("/products")
	products.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	// CSRF is enforced for all state-changing methods (POST/PUT/PATCH/DELETE);
	// read-only GETs pass through untouched.
	products.Use(middleware.CSRFMiddleware(csrfManager))
	{
		products.POST("", controller.CreateProduct)
		products.GET("", controller.ListProducts)
		products.GET("/:id", controller.GetProduct)
		products.PUT("/:id", controller.UpdateProduct)
		// Deleting a product is a sensitive action: require step-up re-authentication.
		products.DELETE("/:id", middleware.RequireReauth(db), controller.DeleteProduct)
	}

	// Metrics stub routes (HU_PROV_05)
	metrics := router.Group("/metrics")
	metrics.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		metrics.GET("/summary", controller.GetMetricsSummary)
		metrics.GET("/top-products", controller.GetTopProducts)
	}
}
