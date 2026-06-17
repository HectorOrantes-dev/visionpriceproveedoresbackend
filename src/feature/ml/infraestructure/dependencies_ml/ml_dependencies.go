package dependencies_ml

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/ml/application/ml_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/ml/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/ml/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/ml/infraestructure/routers"
)

// Init initializes the entire ML feature domain, connecting adapters, use cases, and controllers.
func Init(r *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string) {
	// 1. Adapters (Repository)
	repo := adapters.NewSupabaseMLRepository(db)

	// 2. Application (Use Cases)
	useCase := ml_usecase.NewMLUseCase(repo)

	// 3. Infrastructure (Controllers)
	controller := controllers.NewMLController(useCase)

	// 4. Register Routes
	routers.RegisterMLRoutes(r, controller, jwtSecret)
}
