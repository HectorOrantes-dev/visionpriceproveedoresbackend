package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyHeader is the header through which service-to-service callers (the
// gateway) authenticate against internal microservice endpoints.
const APIKeyHeader = "X-Api-Key"

// APIKeyMiddleware authenticates service-to-service requests by comparing the
// X-Api-Key header against the expected key in constant time. It is independent
// from the user JWT auth: these endpoints are consumed by the gateway, not by
// end users. A missing/empty expected key denies all requests (closed by default).
func APIKeyMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader(APIKeyHeader)
		if expectedKey == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key inválida o ausente"})
			return
		}
		c.Next()
	}
}
