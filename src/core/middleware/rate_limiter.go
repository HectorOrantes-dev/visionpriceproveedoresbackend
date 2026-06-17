package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/core/responses"
)

// RateLimiterConfig defines configuration for a rate limiter instance.
type RateLimiterConfig struct {
	// EndpointGroup is a logical name grouping related endpoints (e.g., "auth_login").
	EndpointGroup string
	// MaxAttempts is the maximum number of attempts allowed within the time window.
	MaxAttempts int
	// Window is the time window in which attempts are counted.
	Window time.Duration
}

// RateLimiter provides DB-backed rate limiting.
type RateLimiter struct {
	db *pgxpool.Pool
}

// NewRateLimiter creates a new RateLimiter backed by the given database pool.
func NewRateLimiter(db *pgxpool.Pool) *RateLimiter {
	return &RateLimiter{db: db}
}

// Middleware returns a Gin middleware that enforces rate limits based on client IP.
func (rl *RateLimiter) Middleware(cfg RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		count, err := rl.countRecentAttempts(c.Request.Context(), clientIP, cfg.EndpointGroup, cfg.Window)
		if err != nil {
			slog.Error("rate limiter: failed to count attempts",
				"error", err,
				"ip", clientIP,
				"endpoint_group", cfg.EndpointGroup,
			)
			// Fail open — don't block the request if we can't check the DB
			c.Next()
			return
		}

		if count >= cfg.MaxAttempts {
			slog.Warn("rate limiter: request blocked",
				"ip", clientIP,
				"endpoint_group", cfg.EndpointGroup,
				"attempts", count,
				"max", cfg.MaxAttempts,
				"window", cfg.Window.String(),
			)
			responses.ErrorResponse(c, http.StatusTooManyRequests,
				"Demasiados intentos. Intente nuevamente más tarde.", nil)
			c.Abort()
			return
		}

		// Record the attempt
		if err := rl.recordAttempt(c.Request.Context(), clientIP, cfg.EndpointGroup); err != nil {
			slog.Error("rate limiter: failed to record attempt",
				"error", err,
				"ip", clientIP,
				"endpoint_group", cfg.EndpointGroup,
			)
		}

		c.Next()
	}
}

// countRecentAttempts counts attempts within the time window for a given identifier and endpoint group.
func (rl *RateLimiter) countRecentAttempts(ctx context.Context, identifier, endpointGroup string, window time.Duration) (int, error) {
	query := `
		SELECT COUNT(*) FROM rate_limit_attempts
		WHERE identifier = $1 AND endpoint_group = $2 AND attempted_at > $3
	`
	windowStart := time.Now().Add(-window)

	var count int
	err := rl.db.QueryRow(ctx, query, identifier, endpointGroup, windowStart).Scan(&count)
	return count, err
}

// recordAttempt inserts a new attempt record.
func (rl *RateLimiter) recordAttempt(ctx context.Context, identifier, endpointGroup string) error {
	query := `INSERT INTO rate_limit_attempts (identifier, endpoint_group) VALUES ($1, $2)`
	_, err := rl.db.Exec(ctx, query, identifier, endpointGroup)
	return err
}

// CleanupOldAttempts removes rate limit entries older than 1 hour.
// Should be called periodically (e.g., from a background goroutine).
func (rl *RateLimiter) CleanupOldAttempts(ctx context.Context) error {
	query := `DELETE FROM rate_limit_attempts WHERE attempted_at < NOW() - INTERVAL '1 hour'`
	_, err := rl.db.Exec(ctx, query)
	return err
}
