package entities

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorCode represents an OTP code for two-factor authentication.
type TwoFactorCode struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Code       string    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	Used       bool      `json:"used"`
	CreatedAt  time.Time `json:"created_at"`
}

// VerifyOTPRequest is the DTO for OTP verification.
type VerifyOTPRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// VerifyOTPResponse is returned after successful OTP verification.
type VerifyOTPResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
