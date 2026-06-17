package entities

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest is the DTO for provider login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned after successful login (before 2FA).
type LoginResponse struct {
	TempToken   string `json:"temp_token"`
	Requires2FA bool   `json:"requires_2fa"`
}

// TokenPair represents a full access + refresh token pair (after 2FA verification).
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordRequest is the DTO for requesting a password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest is the DTO for resetting the password with a token.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PasswordResetToken represents a password reset token stored in the database.
type PasswordResetToken struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	TokenHash  string    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	Used       bool      `json:"used"`
	CreatedAt  time.Time `json:"created_at"`
}

// LogoutRequest is the DTO for provider logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshRequest is the DTO for requesting a new access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

