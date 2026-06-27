package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders sets HTTP response headers that harden the API against common
// browser-based attacks:
//
//   - X-Content-Type-Options: nosniff      → stops MIME-type sniffing
//   - X-Frame-Options: DENY                → anti-clickjacking (legacy browsers)
//   - Referrer-Policy: no-referrer         → don't leak URLs to third parties
//   - Strict-Transport-Security            → force HTTPS (only effective over TLS)
//   - Content-Security-Policy              → lock down what a response may load
//
// A JSON API serves no markup, so the CSP is locked down hard. It is skipped for
// the Swagger UI (dev only) so its own scripts/styles can still load.
//
// Reusable: drop this into any API and call router.Use(middleware.SecurityHeaders()).
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		if !strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		}

		c.Next()
	}
}
