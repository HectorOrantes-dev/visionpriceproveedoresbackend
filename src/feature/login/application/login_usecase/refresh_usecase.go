package login_usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
)

// RefreshToken validates a refresh token, checks if it's revoked, and returns a new access token.
func (uc *LoginUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Parse the refresh token
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Refresh token inválido o expirado")
	}

	// Validate token type
	if claims.TokenType != middleware.TokenTypeRefresh {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Tipo de token inválido")
	}

	// Check if token is revoked in DB
	tokenHash := hashTokenForRevocation(refreshToken)
	revoked, err := uc.repo.IsTokenRevoked(ctx, tokenHash)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar estado del token")
	}
	if revoked {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Refresh token ha sido revocado")
	}

	// Verify the provider still exists (optional, but good for security)
	providerID, err := uuid.Parse(claims.ProviderID)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "ID de proveedor inválido en el token")
	}

	// Usually we might look up the provider to check if they're banned/inactive,
	// but the DB check could be avoided to save a query if we just trust the unrevoked token.
	// Since revocation is the primary security mechanism here, we can proceed.
	// If you want to check provider status, you'd add: uc.repo.FindByID(ctx, providerID)

	// Generate new access token
	accessToken, err := middleware.GenerateTokenWithRole(
		providerID.String(),
		middleware.TokenTypeAccess,
		claims.Role, // Carry over the role from the refresh token
		uc.jwtSecret,
		time.Duration(uc.jwtExpirationMinutes)*time.Minute,
	)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar nuevo access token")
	}

	return accessToken, nil
}
