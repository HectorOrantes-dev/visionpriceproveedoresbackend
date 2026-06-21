package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/visionprice/proveedores-backend/src/core/responses"
)

// ReauthHeader is the header through which the client supplies the current
// password to confirm a sensitive ("grave") action.
const ReauthHeader = "X-Confirm-Password"

// RequireReauth returns middleware that forces step-up re-authentication for
// sensitive actions: the caller must resend their current password in the
// X-Confirm-Password header, which is verified against the stored bcrypt hash.
//
// It must run AFTER AuthMiddleware so provider_id is present in the context.
func RequireReauth(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, exists := c.Get("provider_id")
		if !exists {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
			c.Abort()
			return
		}

		password := c.GetHeader(ReauthHeader)
		if password == "" {
			responses.ErrorResponse(c, http.StatusUnauthorized,
				"Reautenticación requerida para esta acción. Reenvíe su contraseña en el header "+ReauthHeader, nil)
			c.Abort()
			return
		}

		pid, err := uuid.Parse(providerID.(string))
		if err != nil {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Sesión inválida", nil)
			c.Abort()
			return
		}

		var passwordHash string
		err = db.QueryRow(c.Request.Context(),
			`SELECT password_hash FROM providers WHERE id = $1 AND active = TRUE`, pid).Scan(&passwordHash)
		if err != nil {
			// Do not distinguish "not found" from other errors to the client.
			slog.Warn("reauth: provider lookup failed", "error", err, "provider_id", pid)
			responses.ErrorResponse(c, http.StatusUnauthorized, "Reautenticación fallida", nil)
			c.Abort()
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			slog.Warn("reauth: password mismatch", "provider_id", pid)
			responses.ErrorResponse(c, http.StatusUnauthorized, "Reautenticación fallida", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
