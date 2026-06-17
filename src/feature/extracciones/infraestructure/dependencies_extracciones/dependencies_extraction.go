package dependencies_extracciones

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/extracciones/application/extraction_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/extracciones/infraestructure/routers"
)

// Init wires all dependencies for the extracciones feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string) {
	// Repository
	repo := adapters.NewSupabaseExtractionRepository(db)

	// Use case
	useCase := extraction_usecase.NewExtractionUseCase(repo)

	// Controller
	controller := controllers.NewExtractionController(useCase)

	// Routes
	routers.SetupExtractionRoutes(router, controller, jwtSecret)
}
