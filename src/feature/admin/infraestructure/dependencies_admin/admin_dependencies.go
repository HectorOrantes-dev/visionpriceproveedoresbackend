package dependencies_admin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/admin/application/admin_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/admin/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/admin/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/admin/infraestructure/routers"
)

// Init wires all dependencies for the admin feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string, jwtExpirationMinutes int, rateLimiter *middleware.RateLimiter) {
	// Repository
	repo := adapters.NewSupabaseAdminRepository(db)

	// Use case
	useCase := admin_usecase.NewAdminUseCase(repo)

	// Controller
	controller := controllers.NewAdminController(useCase, jwtSecret, jwtExpirationMinutes)

	// Rate limiter config: 5 requests per minute for admin login
	adminLoginRL := middleware.RateLimiterConfig{
		EndpointGroup: "admin_login",
		MaxAttempts:   5,
		Window:        1 * time.Minute,
	}

	// Routes
	routers.SetupAdminRoutes(router, controller, jwtSecret, rateLimiter, adminLoginRL)
}
