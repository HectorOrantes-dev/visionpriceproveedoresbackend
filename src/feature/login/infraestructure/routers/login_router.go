package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/login/infraestructure/controllers"
)

// SetupLoginRoutes registers all routes for the login feature.
func SetupLoginRoutes(router *gin.RouterGroup, controller *controllers.LoginController, jwtSecret string, rateLimiter *middleware.RateLimiter, loginRL, forgotRL, resetRL, refreshRL middleware.RateLimiterConfig) {
	auth := router.Group("/auth")
	{
		// Public endpoints (rate limited)
		auth.POST("/login", rateLimiter.Middleware(loginRL), controller.Login)
		auth.POST("/forgot-password", rateLimiter.Middleware(forgotRL), controller.ForgotPassword)
		auth.POST("/reset-password", rateLimiter.Middleware(resetRL), controller.ResetPassword)
		auth.POST("/refresh", rateLimiter.Middleware(refreshRL), controller.Refresh)

		// Protected: requires valid access token
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
		{
			protected.POST("/logout", controller.Logout)
		}
	}
}
