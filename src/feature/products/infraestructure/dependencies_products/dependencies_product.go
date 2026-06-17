package dependencies_products

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/products/application/product_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/routers"
)

// Init wires all dependencies for the products feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string) {
	// Repository
	repo := adapters.NewSupabaseProductRepository(db)

	// Use case
	useCase := product_usecase.NewProductUseCase(repo)

	// Controller
	controller := controllers.NewProductController(useCase)

	// Routes
	routers.SetupProductRoutes(router, controller, jwtSecret)
}
