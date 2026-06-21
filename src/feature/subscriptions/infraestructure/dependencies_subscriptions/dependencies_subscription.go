package dependencies_subscriptions

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/application/subscription_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/infraestructure/adapters"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/infraestructure/controllers"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/infraestructure/routers"
)

// Init wires the subscriptions feature, registers its routes, and returns the
// use case so other features (products, extracciones, register, payments) can
// consume it through their own narrow ports.
func Init(router *gin.RouterGroup, db *pgxpool.Pool, jwtSecret string) *subscription_usecase.SubscriptionUseCase {
	repo := adapters.NewSupabaseSubscriptionRepository(db)
	useCase := subscription_usecase.NewSubscriptionUseCase(repo)
	controller := controllers.NewSubscriptionController(useCase)

	routers.SetupSubscriptionRoutes(router, controller, jwtSecret)

	return useCase
}
