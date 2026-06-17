package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/infraestructure/controllers"
)

// SetupExtractionRoutes registers all routes for the extracciones feature.
func SetupExtractionRoutes(router *gin.RouterGroup, controller *controllers.ExtractionController, jwtSecret string) {
	extractions := router.Group("/extractions")
	extractions.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		extractions.POST("/detect-columns", controller.DetectColumns)
		extractions.POST("/mapping", controller.SaveMapping)
		extractions.GET("/mapping", controller.GetMapping)
		extractions.POST("/import", controller.ProcessImport)
	}
}
