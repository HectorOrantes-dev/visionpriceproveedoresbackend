package dependencies_login

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/login/application/login_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/login/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/login/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/login/infraestructure/routers"
)

// Init wires all dependencies for the login feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, jwtExpirationMinutes int, otpExpirationMinutes int, passwordResetExpirationMinutes int, rateLimiter *middleware.RateLimiter) {
	// Repository
	repo := adapters.NewSupabaseLoginRepository(db)

	// Use case
	useCase := login_usecase.NewLoginUseCase(repo, jwtSecret, jwtExpirationMinutes, otpExpirationMinutes, passwordResetExpirationMinutes)

	// Controller
	controller := controllers.NewLoginController(useCase)

	// Rate limiter configs
	loginRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_login",
		MaxAttempts:   5,
		Window:        1 * time.Minute,
	}
	forgotRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_forgot_password",
		MaxAttempts:   3,
		Window:        1 * time.Minute,
	}
	resetRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_reset_password",
		MaxAttempts:   5,
		Window:        1 * time.Minute,
	}

	refreshRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_refresh",
		MaxAttempts:   5,
		Window:        1 * time.Minute,
	}

	// Routes
	routers.SetupLoginRoutes(router, controller, jwtSecret, rateLimiter, loginRL, forgotRL, resetRL, refreshRL)
}
