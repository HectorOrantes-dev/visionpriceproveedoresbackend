package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/infraestructure/controllers"
)

// SetupCatalogoRoutes registers the service-to-service catalog endpoints,
// authenticated by the X-Api-Key header (not the user JWT).
func SetupCatalogoRoutes(router *gin.RouterGroup, controller *controllers.CatalogoController, apiKey string) {
	catalogo := router.Group("")
	catalogo.Use(middleware.APIKeyMiddleware(apiKey))
	{
		catalogo.GET("/productos/cercanos", controller.ProductosCercanos)
		catalogo.GET("/productos", controller.ProductosPorIDs)
	}
}
