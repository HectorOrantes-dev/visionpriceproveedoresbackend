package dependencies_register

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/register/application/register_usecase"
	registerDomain "github.com/visionprice/proveedores-backend/src/feature/register/domain"
	"github.com/visionprice/proveedores-backend/src/feature/register/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/register/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/register/infraestructure/routers"
)

// Init wires all dependencies for the register feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, subscriptions registerDomain.DefaultSubscriptionCreator, rateLimiter *middleware.RateLimiter) {
	// Repository (adapter)
	repo := adapters.NewSupabaseRegisterRepository(db)

	// Use case
	useCase := register_usecase.NewRegisterUseCase(repo, subscriptions)

	// Controller
	controller := controllers.NewRegisterController(useCase)

	// Rate limiter config: 3 requests per minute for registration
	rlConfig := middleware.RateLimiterConfig{
		EndpointGroup: "auth_register",
		MaxAttempts:   3,
		Window:        1 * time.Minute,
	}

	// Routes
	routers.SetupRegisterRoutes(router, controller, rateLimiter, rlConfig)
}
