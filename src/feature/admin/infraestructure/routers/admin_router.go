package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/admin/infraestructure/controllers"
)

// SetupAdminRoutes registers all routes for the admin feature.
func SetupAdminRoutes(router *gin.RouterGroup, controller *controllers.AdminController, jwtSecret string, rateLimiter *middleware.RateLimiter, adminLoginRL middleware.RateLimiterConfig) {
	admin := router.Group("/admin")
	{
		// Public: admin login (rate limited, no auth required)
		admin.POST("/login", rateLimiter.Middleware(adminLoginRL), controller.AdminLogin)

		// Protected: requires JWT with role USER_SYS_ADMIN + audit logging
		protected := admin.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
		protected.Use(middleware.RequireRole("USER_SYS_ADMIN"))
		protected.Use(middleware.AuditLogger())
		{
			protected.GET("/metrics", controller.GetMetrics)
			protected.GET("/geography/providers", controller.GetProviderMapPins)
		}
	}
}
