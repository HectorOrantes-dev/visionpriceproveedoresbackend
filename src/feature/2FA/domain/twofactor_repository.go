package domain

import (
	"context"

	"github.com/google/uuid"
)

// TwoFactorRepository defines the port for 2FA persistence operations.
type TwoFactorRepository interface {
	// CreateOTP stores a new OTP code for a provider.
	CreateOTP(ctx context.Context, providerID uuid.UUID, code string, expirationMinutes int) error

	// ValidateOTP checks if the code is valid (unused, not expired) for the provider and marks it as used.
	ValidateOTP(ctx context.Context, providerID uuid.UUID, code string) (bool, error)

	// InvalidateExpired marks all expired OTP codes as used (cleanup).
	InvalidateExpired(ctx context.Context) error

	// GetProviderContact returns the email and display name of a provider, used
	// to deliver the OTP.
	GetProviderContact(ctx context.Context, providerID uuid.UUID) (email, name string, err error)
}
