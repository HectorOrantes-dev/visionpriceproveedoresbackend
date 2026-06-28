package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/payments/infraestructure/controllers"
)

// SetupPaymentRoutes registers checkout (authenticated) and webhook (public) routes.
func SetupPaymentRoutes(router *gin.RouterGroup, controller *controllers.PaymentController, db *pgxpool.Pool, csrfManager *csrf.Manager, jwtSecret string) {
	// Checkout: authenticated + CSRF + step-up reauth (billing is a sensitive action).
	sub := router.Group("/subscription")
	sub.Use(middleware.AuthMiddleware(jwtSecret, middleware.TokenTypeAccess))
	sub.Use(middleware.CSRFMiddleware(csrfManager))
	{
		sub.POST("/checkout", middleware.RequireReauth(db), middleware.IdempotencyMiddleware(db), controller.CreateCheckout)
	}

	// Webhooks: public endpoints authenticated by the gateway's own signature.
	// No JWT/CSRF — they are server-to-server callbacks.
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/conekta", controller.ConektaWebhook)
		webhooks.POST("/paypal", controller.PayPalWebhook)
	}
}
