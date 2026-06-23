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
func Init(router *gin.RouterGroup, db *pgxpool.Pool, csrfManager *csrf.Manager, jwtSecret string, otpExpirationMinutes int, jwtExpirationMinutes int, refreshTokenExpirationHours int, gmail adapters.GmailConfig, brevo adapters.BrevoConfig, smtp adapters.SMTPConfig, rateLimiter *middleware.RateLimiter) {
	// Repository
	repo := adapters.NewSupabaseTwoFactorRepository(db)

	// OTP notifier selection by priority (the HTTP-based ones use 443, so they
	// work where SMTP ports are blocked, e.g. Railway):
	//   1. Gmail API — OAuth2; sends from a real Gmail, best deliverability, no domain.
	//   2. Brevo     — HTTP API; sends to any recipient with just a verified sender (no domain).
	//   3. SMTP      — only where outbound SMTP is allowed.
	//   4. Log fallback — nothing configured (dev): the code is logged, not emailed.
	var notifier domain.OTPNotifier
	switch {
	case gmail.RefreshToken != "":
		notifier = adapters.NewGmailOTPNotifier(gmail)
	case brevo.APIKey != "":
		notifier = adapters.NewBrevoOTPNotifier(brevo)
	case smtp.Host != "":
		notifier = adapters.NewSMTPOTPNotifier(smtp)
	default:
		log.Println("WARNING: no email provider configured (GMAIL_REFRESH_TOKEN/BREVO_API_KEY/SMTP_HOST) — 2FA OTP codes will be logged, not emailed.")
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
