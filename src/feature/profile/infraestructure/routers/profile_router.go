package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/profile/infraestructure/controllers"
)

// SetupProfileRoutes registers the authenticated provider profile routes.
func SetupProfileRoutes(router *gin.RouterGroup, controller *controllers.ProfileController, jwtSecret string) {
	auth := router.Group("/auth")
	auth.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		auth.GET("/me", controller.GetMe)
	}
}
