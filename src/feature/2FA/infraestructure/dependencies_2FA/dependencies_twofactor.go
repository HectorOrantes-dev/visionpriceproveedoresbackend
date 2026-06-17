package dependencies_2FA

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/application/twofactor_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/routers"
)

// Init wires all dependencies for the 2FA feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, otpExpirationMinutes int, jwtExpirationMinutes int, refreshTokenExpirationHours int, rateLimiter *middleware.RateLimiter) {
	// Repository
	repo := adapters.NewSupabaseTwoFactorRepository(db)

	// Use case
	useCase := twofactor_usecase.NewTwoFactorUseCase(repo, jwtSecret, otpExpirationMinutes, jwtExpirationMinutes, refreshTokenExpirationHours)

	// Controller
	controller := controllers.NewTwoFactorController(useCase)

	// Rate limiter configs
	generateRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_2fa_generate",
		MaxAttempts:   3,
		Window:        1 * time.Minute,
	}
	verifyRL := middleware.RateLimiterConfig{
		EndpointGroup: "auth_2fa_verify",
		MaxAttempts:   5,
		Window:        1 * time.Minute,
	}

	// Routes
	routers.SetupTwoFactorRoutes(router, controller, jwtSecret, rateLimiter, generateRL, verifyRL)
}
