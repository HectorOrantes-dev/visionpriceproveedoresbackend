package entities

import (
	"time"

	"github.com/google/uuid"
)

// Provider represents a supplier/business entity in the system.
type Provider struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	RFC          string    `json:"rfc"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest is the DTO for provider registration.
type RegisterRequest struct {
	BusinessName string `json:"business_name" binding:"required,min=2,max=255"`
	RFC          string `json:"rfc" binding:"required,min=12,max=13"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required,min=10,max=20"`
	Password     string `json:"password" binding:"required,min=8"`
}

// RegisterResponse is the DTO returned after successful registration.
type RegisterResponse struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	RFC          string    `json:"rfc"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
}

// MaskRFC returns a masked version of an RFC string.
// Example: "XAXX010101AB3" → "XAXX*******B3"
func MaskRFC(rfc string) string {
	if len(rfc) <= 6 {
		return rfc
	}
	masked := rfc[:4]
	for i := 4; i < len(rfc)-2; i++ {
		masked += "*"
	}
	masked += rfc[len(rfc)-2:]
	return masked
}
