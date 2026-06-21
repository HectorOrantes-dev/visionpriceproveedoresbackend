package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	"github.com/visionprice/proveedores-backend/src/core/responses"
)

// safeMethods are read-only and never require a CSRF token.
var safeMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// CSRFMiddleware enforces the per-session synchronizer token on state-changing
// requests. It must run AFTER AuthMiddleware so that provider_id is in context.
func CSRFMiddleware(mgr *csrf.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if safeMethods[c.Request.Method] {
			c.Next()
			return
		}

		providerID, exists := c.Get("provider_id")
		if !exists {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
			c.Abort()
			return
		}

		token := c.GetHeader(csrf.HeaderName)
		valid, err := mgr.Validate(c.Request.Context(), providerID.(string), token)
		if err != nil {
			slog.Error("csrf: validation failed", "error", err, "provider_id", providerID)
			responses.ErrorResponse(c, http.StatusInternalServerError, "Error al validar token CSRF", nil)
			c.Abort()
			return
		}
		if !valid {
			slog.Warn("csrf: token rejected", "provider_id", providerID, "method", c.Request.Method, "path", c.Request.URL.Path)
			responses.ErrorResponse(c, http.StatusForbidden, "Token CSRF inválido o ausente", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
