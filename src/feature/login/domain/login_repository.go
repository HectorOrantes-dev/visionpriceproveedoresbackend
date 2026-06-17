package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	registerEntities "github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// LoginRepository defines the port for login-related persistence operations.
type LoginRepository interface {
	// FindByEmail retrieves a provider by email address.
	FindByEmail(ctx context.Context, email string) (*registerEntities.Provider, error)

	// UpdatePassword updates the password hash for a provider.
	UpdatePassword(ctx context.Context, providerID uuid.UUID, newPasswordHash string) error

	// CreateResetToken stores a password reset token.
	CreateResetToken(ctx context.Context, providerID uuid.UUID, tokenHash string, expiresAtMinutes int) error

	// FindValidResetToken finds a valid (unused, not expired) reset token by its hash.
	FindValidResetToken(ctx context.Context, tokenHash string) (providerID uuid.UUID, err error)

	// MarkResetTokenUsed marks a reset token as used.
	MarkResetTokenUsed(ctx context.Context, tokenHash string) error

	// RevokeToken adds a token hash to the revocation blacklist.
	RevokeToken(ctx context.Context, tokenHash string, providerID uuid.UUID, expiresAt time.Time) error

	// IsTokenRevoked checks if a token hash exists in the revocation blacklist.
	IsTokenRevoked(ctx context.Context, tokenHash string) (bool, error)
}

