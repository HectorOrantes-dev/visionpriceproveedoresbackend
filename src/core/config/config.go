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

	// Brevo (HTTP email API — works without a domain via a verified sender)
	BrevoAPIKey    string
	BrevoFromEmail string
	BrevoFromName  string

	// Gmail API (OAuth2 — sends from a real Gmail account, best deliverability, no domain)
	GmailClientID     string
	GmailClientSecret string
	GmailRefreshToken string
	GmailFrom         string
	GmailFromName     string

	// Payments / billing
	PaymentsEnabled        bool
	PaymentDefaultGateway  string
	SubscriptionSuccessURL string
	SubscriptionCancelURL  string
	ConektaPrivateKey      string
	ConektaWebhookSecret   string
	ConektaPlanPro         string
	ConektaPlanMax         string
	PayPalClientID         string
	PayPalClientSecret     string
	PayPalWebhookID        string
	PayPalEnv              string
	PayPalPlanPro          string
	PayPalPlanMax          string
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
		ServerPort:         getServerPort(),
		DatabaseURL:        getEnvRequired("DATABASE_URL"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseKey:        getEnv("SUPABASE_KEY", ""),
		JWTSecret:          jwtSecret,
		CORSAllowedOrigins: corsOrigins,
		EnableSwagger:      getEnvAsBool("ENABLE_SWAGGER", false),

		BrevoAPIKey:    getEnv("BREVO_API_KEY", ""),
		BrevoFromEmail: getEnv("BREVO_FROM_EMAIL", ""),
		BrevoFromName:  getEnv("BREVO_FROM_NAME", "VisionPrice"),

		GmailClientID:     getEnv("GMAIL_CLIENT_ID", ""),
		GmailClientSecret: getEnv("GMAIL_CLIENT_SECRET", ""),
		GmailRefreshToken: getEnv("GMAIL_REFRESH_TOKEN", ""),
		GmailFrom:         getEnv("GMAIL_FROM", ""),
		GmailFromName:     getEnv("GMAIL_FROM_NAME", "VisionPrice"),

		JWTExpirationMinutes:           getEnvAsInt("JWT_EXPIRATION_MINUTES", 15),
		RefreshTokenExpirationHours:    getEnvAsInt("REFRESH_TOKEN_EXPIRATION_HOURS", 168),
		OTPExpirationMinutes:           getEnvAsInt("OTP_EXPIRATION_MINUTES", 5),
		PasswordResetExpirationMinutes: getEnvAsInt("PASSWORD_RESET_EXPIRATION_MINUTES", 15),

		PaymentsEnabled:        getEnvAsBool("PAYMENTS_ENABLED", false),
		PaymentDefaultGateway:  getEnv("PAYMENT_DEFAULT_GATEWAY", "conekta"),
		SubscriptionSuccessURL: getEnv("SUBSCRIPTION_SUCCESS_URL", ""),
		SubscriptionCancelURL:  getEnv("SUBSCRIPTION_CANCEL_URL", ""),
		ConektaPrivateKey:      getEnv("CONEKTA_PRIVATE_KEY", ""),
		ConektaWebhookSecret:   getEnv("CONEKTA_WEBHOOK_SECRET", ""),
		ConektaPlanPro:         getEnv("CONEKTA_PLAN_PRO", ""),
		ConektaPlanMax:         getEnv("CONEKTA_PLAN_MAX", ""),
		PayPalClientID:         getEnv("PAYPAL_CLIENT_ID", ""),
		PayPalClientSecret:     getEnv("PAYPAL_CLIENT_SECRET", ""),
		PayPalWebhookID:        getEnv("PAYPAL_WEBHOOK_ID", ""),
		PayPalEnv:              getEnv("PAYPAL_ENV", "sandbox"),
		PayPalPlanPro:          getEnv("PAYPAL_PLAN_PRO", ""),
		PayPalPlanMax:          getEnv("PAYPAL_PLAN_MAX", ""),
	}

	return cfg
}

// getServerPort resolves the port to listen on. Platforms like Railway/Heroku
// inject PORT at runtime; it takes precedence over the app's own SERVER_PORT.
func getServerPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return getEnv("SERVER_PORT", "8080")
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
