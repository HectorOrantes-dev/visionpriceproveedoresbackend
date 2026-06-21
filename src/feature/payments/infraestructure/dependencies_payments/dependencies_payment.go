package dependencies_payments

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/feature/payments/application/payment_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain"
	"github.com/visionprice/proveedores-backend/src/feature/payments/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/payments/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/payments/infraestructure/routers"
)

// Config aggregates all settings the payments feature needs.
type Config struct {
	Enabled        bool
	DefaultGateway string
	SuccessURL     string
	CancelURL      string
	Conekta        adapters.ConektaConfig
	PayPal         adapters.PayPalConfig
}

// Init wires the payments feature (gateways, event store, use case) and routes.
// subs is the subscriptions use case, satisfying the SubscriptionUpdater port.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, csrfManager *csrf.Manager, subs domain.SubscriptionUpdater, cfg Config, jwtSecret string) {
	gateways := map[string]domain.PaymentGateway{
		"conekta": adapters.NewConektaGateway(cfg.Conekta),
		"paypal":  adapters.NewPayPalGateway(cfg.PayPal),
	}

	events := adapters.NewSupabasePaymentRepository(db)

	useCase := payment_usecase.NewPaymentUseCase(payment_usecase.Config{
		Enabled:        cfg.Enabled,
		DefaultGateway: cfg.DefaultGateway,
		SuccessURL:     cfg.SuccessURL,
		CancelURL:      cfg.CancelURL,
	}, gateways, events, subs)

	controller := controllers.NewPaymentController(useCase)

	routers.SetupPaymentRoutes(router, controller, db, csrfManager, jwtSecret)
}
