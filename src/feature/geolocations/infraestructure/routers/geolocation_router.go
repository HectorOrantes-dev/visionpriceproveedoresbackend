package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/infraestructure/controllers"
)

// SetupGeolocationRoutes registers all routes for the geolocations feature.
func SetupGeolocationRoutes(router *gin.RouterGroup, controller *controllers.GeolocationController, jwtSecret string) {
	providers := router.Group("/providers")
	providers.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		providers.PUT("/location", controller.SetLocation)
		providers.GET("/location", controller.GetLocation)
	}
}
