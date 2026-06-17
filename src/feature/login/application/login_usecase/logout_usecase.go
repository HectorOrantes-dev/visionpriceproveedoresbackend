package login_usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
)

// refreshTokenMaxTTL is the maximum time a refresh token could be valid.
// Used as expiration for revocation entries to ensure they're cleaned up.
const refreshTokenMaxTTL = 168 * time.Hour // 7 days

// Logout revokes the given refresh token so it cannot be used again.
func (uc *LoginUseCase) Logout(ctx context.Context, providerID string, refreshToken string) error {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	// Hash the refresh token for storage (never store raw tokens)
	tokenHash := hashTokenForRevocation(refreshToken)

	// Set expiration to refresh token TTL from now (worst case, it expires naturally)
	expiresAt := time.Now().Add(refreshTokenMaxTTL)

	if err := uc.repo.RevokeToken(ctx, tokenHash, pid, expiresAt); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al cerrar sesión")
	}

	slog.Info("AUDIT: logout",
		"provider_id", providerID,
		"action", "token_revoked",
	)

	return nil
}

// hashTokenForRevocation creates a SHA-256 hash of a token string for revocation storage.
func hashTokenForRevocation(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
