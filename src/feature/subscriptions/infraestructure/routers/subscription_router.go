package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/infraestructure/controllers"
)

// SetupSubscriptionRoutes registers the subscription read endpoints.
func SetupSubscriptionRoutes(router *gin.RouterGroup, controller *controllers.SubscriptionController, jwtSecret string) {
	authed := router.Group("")
	authed.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	{
		authed.GET("/subscription", controller.GetCurrent)
		authed.GET("/plans", controller.ListPlans)
	}
}
