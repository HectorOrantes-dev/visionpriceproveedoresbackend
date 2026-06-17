package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/controllers"
)

// SetupTwoFactorRoutes registers all routes for the 2FA feature.
func SetupTwoFactorRoutes(router *gin.RouterGroup, controller *controllers.TwoFactorController, jwtSecret string, rateLimiter *middleware.RateLimiter, generateRL, verifyRL middleware.RateLimiterConfig) {
	twoFA := router.Group("/auth/2fa")
	twoFA.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeOTPTemp))
	{
		twoFA.POST("/generate", rateLimiter.Middleware(generateRL), controller.GenerateOTP)
		twoFA.POST("/verify", rateLimiter.Middleware(verifyRL), controller.VerifyOTP)
	}
}
