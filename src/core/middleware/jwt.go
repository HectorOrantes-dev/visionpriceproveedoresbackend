package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/visionprice/proveedores-backend/src/core/responses"
)

// TokenType represents the different JWT token types in the system.
type TokenType string

const (
	TokenTypeAccess        TokenType = "access"
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeOTPTemp       TokenType = "otp_temp"
	TokenTypePasswordReset TokenType = "password_reset"
)

// Claims represents the custom JWT claims.
type Claims struct {
	ProviderID string    `json:"provider_id"`
	TokenType  TokenType `json:"token_type"`
	Role       string    `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens and injects provider_id into the context.
// It accepts one or more allowed token types.
func AuthMiddleware(jwtSecret string, allowedTypes ...TokenType) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Token de autorización requerido", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Formato de token inválido. Use: Bearer <token>", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			responses.ErrorResponse(c, http.StatusUnauthorized, "Token inválido o expirado", nil)
			c.Abort()
			return
		}

		// Validate token type
		if len(allowedTypes) > 0 {
			typeAllowed := false
			for _, at := range allowedTypes {
				if claims.TokenType == at {
					typeAllowed = true
					break
				}
			}
			if !typeAllowed {
				responses.ErrorResponse(c, http.StatusUnauthorized, "Tipo de token no permitido para este recurso", nil)
				c.Abort()
				return
			}
		}

		c.Set("provider_id", claims.ProviderID)
		c.Set("token_type", string(claims.TokenType))
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole returns a middleware that checks the JWT role claim.
// If the user's role does not match the required role, it responds with 403 Forbidden.
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != requiredRole {
			responses.ErrorResponse(c, http.StatusForbidden, "Acceso denegado. Rol insuficiente.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GenerateToken creates a new signed JWT token.
func GenerateToken(providerID string, tokenType TokenType, jwtSecret string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		ProviderID: providerID,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// GenerateTokenWithRole creates a new signed JWT token that includes a role claim.
func GenerateTokenWithRole(userID string, tokenType TokenType, role string, jwtSecret string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		ProviderID: userID,
		TokenType:  tokenType,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
