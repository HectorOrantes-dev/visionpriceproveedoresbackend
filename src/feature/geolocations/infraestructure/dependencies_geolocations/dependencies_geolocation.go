package dependencies_geolocations

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/geolocations/application/geolocation_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/infraestructure/routers"
)

// Init wires all dependencies for the geolocations feature and registers routes.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string) {
	// Repository
	repo := adapters.NewSupabaseGeolocationRepository(db)

	// Use case
	useCase := geolocation_usecase.NewGeolocationUseCase(repo)

	// Controller
	controller := controllers.NewGeolocationController(useCase)

	// Routes
	routers.SetupGeolocationRoutes(router, controller, jwtSecret)
}
