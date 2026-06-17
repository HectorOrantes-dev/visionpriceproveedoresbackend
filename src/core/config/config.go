package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
type Config struct {
	ServerPort                     string
	DatabaseURL                    string
	SupabaseURL                    string
	SupabaseKey                    string
	JWTSecret                      string
	JWTExpirationMinutes           int
	RefreshTokenExpirationHours    int
	OTPExpirationMinutes           int
	PasswordResetExpirationMinutes int
	CORSAllowedOrigins             string
	EnableSwagger                  bool
}

// Load reads environment variables from .env and returns a Config.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: .env file not found, using system environment variables")
	}

	jwtSecret := getEnvRequired("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Println("WARNING: JWT_SECRET is shorter than 32 characters. Use a strong secret in production (e.g., openssl rand -hex 32)")
	}

	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if corsOrigins == "" {
		log.Fatal("FATAL: CORS_ALLOWED_ORIGINS is not set. Set explicit origins or '*' for development only.")
	}
	if corsOrigins == "*" {
		log.Println("WARNING: CORS_ALLOWED_ORIGINS is set to '*'. This is insecure for production. Set explicit origins.")
	}

	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DatabaseURL:        getEnvRequired("DATABASE_URL"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseKey:        getEnv("SUPABASE_KEY", ""),
		JWTSecret:          jwtSecret,
		CORSAllowedOrigins: corsOrigins,
		EnableSwagger:      getEnvAsBool("ENABLE_SWAGGER", false),

		JWTExpirationMinutes:           getEnvAsInt("JWT_EXPIRATION_MINUTES", 15),
		RefreshTokenExpirationHours:    getEnvAsInt("REFRESH_TOKEN_EXPIRATION_HOURS", 168),
		OTPExpirationMinutes:           getEnvAsInt("OTP_EXPIRATION_MINUTES", 5),
		PasswordResetExpirationMinutes: getEnvAsInt("PASSWORD_RESET_EXPIRATION_MINUTES", 15),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvRequired(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("FATAL: required environment variable %s is not set", key)
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		log.Printf("WARNING: invalid integer for %s, using default %d", key, fallback)
		return fallback
	}
	return value
}

func getEnvAsBool(key string, fallback bool) bool {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	value, err := strconv.ParseBool(strValue)
	if err != nil {
		log.Printf("WARNING: invalid boolean for %s, using default %v", key, fallback)
		return fallback
	}
	return value
}

