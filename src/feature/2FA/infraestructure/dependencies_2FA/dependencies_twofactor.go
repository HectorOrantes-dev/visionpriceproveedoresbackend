package dependencies_2FA

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/application/twofactor_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/domain"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/routers"
)

// Init wires all dependencies for the 2FA feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, csrfManager *csrf.Manager, jwtSecret string, otpExpirationMinutes int, jwtExpirationMinutes int, refreshTokenExpirationHours int, smtp adapters.SMTPConfig, rateLimiter *middleware.RateLimiter) {
	// Repository
	repo := adapters.NewSupabaseTwoFactorRepository(db)

	// OTP notifier: real SMTP when configured, log fallback otherwise (dev).
	var notifier domain.OTPNotifier
	if smtp.Host != "" {
		notifier = adapters.NewSMTPOTPNotifier(smtp)
	} else {
		log.Println("WARNING: SMTP_HOST not set — 2FA OTP codes will be logged, not emailed.")
		notifier = adapters.NewLogOTPNotifier()
	}

	// Use case
	useCase := twofactor_usecase.NewTwoFactorUseCase(repo, notifier, csrfManager, jwtSecret, otpExpirationMinutes, jwtExpirationMinutes, refreshTokenExpirationHours)

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
