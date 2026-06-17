package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/config"
	"github.com/visionprice/proveedores-backend/src/core/database"
	"github.com/visionprice/proveedores-backend/src/core/middleware"

	dependencies2FA "github.com/visionprice/proveedores-backend/src/feature/2FA/infraestructure/dependencies_2FA"
	dependenciesAdmin "github.com/visionprice/proveedores-backend/src/feature/admin/infraestructure/dependencies_admin"
	dependenciesExtracciones "github.com/visionprice/proveedores-backend/src/feature/extracciones/infraestructure/dependencies_extracciones"
	dependenciesGeolocations "github.com/visionprice/proveedores-backend/src/feature/geolocations/infraestructure/dependencies_geolocations"
	dependenciesLogin "github.com/visionprice/proveedores-backend/src/feature/login/infraestructure/dependencies_login"
	dependenciesProducts "github.com/visionprice/proveedores-backend/src/feature/products/infraestructure/dependencies_products"
	dependenciesRegister "github.com/visionprice/proveedores-backend/src/feature/register/infraestructure/dependencies_register"
	dependenciesML "github.com/visionprice/proveedores-backend/src/feature/ml/infraestructure/dependencies_ml"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/visionprice/proveedores-backend/docs"
)

// @title           VisionPrice Proveedores API
// @version         1.0
// @description     API backend para el módulo de Proveedores de VisionPrice
// @termsOfService  http://swagger.io/terms/

// @contact.name   VisionPrice Support
// @contact.email  soporte@visionprice.app

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ingrese el token JWT con el prefijo Bearer. Ejemplo: "Bearer {token}"

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	ctx := context.Background()
	dbPool := database.NewPool(ctx, cfg.DatabaseURL)
	defer dbPool.Close()

	// Initialize rate limiter (DB-backed)
	rateLimiter := middleware.NewRateLimiter(dbPool)

	// Start background cleanup of old rate limit entries (every 10 minutes)
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := rateLimiter.CleanupOldAttempts(context.Background()); err != nil {
				log.Printf("WARNING: rate limiter cleanup failed: %v", err)
			}
		}
	}()

	// Create Gin engine
	router := gin.Default()

	// CORS configuration
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	if cfg.CORSAllowedOrigins == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(cfg.CORSAllowedOrigins, ",")
	}

	router.Use(cors.New(corsConfig))

	// API v1 group
	v1 := router.Group("/api/v1")

	// Initialize features
	dependenciesRegister.Init(v1, dbPool, rateLimiter)
	dependenciesLogin.Init(v1, dbPool, cfg.JWTSecret, cfg.JWTExpirationMinutes, cfg.OTPExpirationMinutes, cfg.PasswordResetExpirationMinutes, rateLimiter)
	dependencies2FA.Init(v1, dbPool, cfg.JWTSecret, cfg.OTPExpirationMinutes, cfg.JWTExpirationMinutes, cfg.RefreshTokenExpirationHours, rateLimiter)
	dependenciesGeolocations.Init(v1, dbPool, cfg.JWTSecret)
	dependenciesProducts.Init(v1, dbPool, cfg.JWTSecret)
	dependenciesExtracciones.Init(v1, dbPool, cfg.JWTSecret)
	dependenciesAdmin.Init(v1, dbPool, cfg.JWTSecret, cfg.JWTExpirationMinutes, rateLimiter)
	dependenciesML.Init(v1, dbPool, cfg.JWTSecret)

	// Swagger documentation (only enabled when ENABLE_SWAGGER=true)
	if cfg.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Printf("📚 Swagger docs enabled at http://localhost:%s/swagger/index.html", cfg.ServerPort)
	} else {
		log.Println("📚 Swagger docs disabled. Set ENABLE_SWAGGER=true to enable.")
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "visionprice-proveedores-backend",
		})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 VisionPrice Proveedores Backend starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("FATAL: server failed to start: %v", err)
	}
}
