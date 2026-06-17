package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger returns a Gin middleware that logs structured audit entries
// for every request to the endpoints it protects.
// Fields logged: timestamp, user_id, role, method, path, client_ip, status, latency.
func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request
		c.Next()

		// Extract user info from context (set by AuthMiddleware)
		userID, _ := c.Get("provider_id")
		role, _ := c.Get("role")

		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}
		roleStr := ""
		if role != nil {
			roleStr = role.(string)
		}

		latency := time.Since(start)
		status := c.Writer.Status()

		slog.Info("AUDIT",
			"timestamp", time.Now().UTC().Format(time.RFC3339),
			"user_id", userIDStr,
			"role", roleStr,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
