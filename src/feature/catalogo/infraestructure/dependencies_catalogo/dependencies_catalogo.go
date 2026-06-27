package dependencies_catalogo

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/catalogo/application/catalogo_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/infraestructure/routers"
)

// Init wires the catalogo microservice feature and registers its routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, apiKey string) {
	if apiKey == "" {
		log.Println("WARNING: MICROSERVICE_API_KEY no configurada — los endpoints del catálogo rechazarán todo con 401.")
	}

	repo := adapters.NewSupabaseCatalogoRepository(db)
	useCase := catalogo_usecase.NewCatalogoUseCase(repo)
	controller := controllers.NewCatalogoController(useCase)

	routers.SetupCatalogoRoutes(router, controller, apiKey)
}
