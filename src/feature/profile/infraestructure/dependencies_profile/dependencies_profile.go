package dependencies_profile

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/feature/profile/application/profile_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/profile/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/profile/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/profile/infraestructure/routers"
)

// Init wires all dependencies for the profile feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, csrfManager *csrf.Manager, jwtSecret string) {
	repo := adapters.NewSupabaseProfileRepository(db)
	useCase := profile_usecase.NewProfileUseCase(repo)
	controller := controllers.NewProfileController(useCase)
	routers.SetupProfileRoutes(router, controller, jwtSecret, csrfManager)
}
