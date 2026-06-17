package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/register/infraestructure/controllers"
)

// SetupRegisterRoutes registers all routes for the register feature.
func SetupRegisterRoutes(router *gin.RouterGroup, controller *controllers.RegisterController, rateLimiter *middleware.RateLimiter, rlConfig middleware.RateLimiterConfig) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", rateLimiter.Middleware(rlConfig), controller.Register)
	}
}
