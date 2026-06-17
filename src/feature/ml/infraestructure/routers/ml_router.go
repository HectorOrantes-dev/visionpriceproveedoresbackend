package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/ml/infraestructure/controllers"
)

// RegisterMLRoutes registers all routes for the Machine Learning module.
func RegisterMLRoutes(r *gin.RouterGroup, controller *controllers.MLController, jwtSecret string) {
	// Group for ML routes under /ml
	mlGroup := r.Group("/ml")
	
	// Apply JWT authentication middleware to all ML routes
	mlGroup.Use(middleware.AuthMiddleware(jwtSecret))

	{
		mlGroup.POST("/products/classify", controller.ClassifyProduct)
		mlGroup.GET("/products/duplicates", controller.DetectDuplicates)
		mlGroup.GET("/products/anomalies", controller.DetectAnomalies)
	}
}
