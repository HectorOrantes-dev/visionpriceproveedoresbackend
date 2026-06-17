package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProviderMapPin represents a provider's public data for the admin map (HU_SYS_02).
// Contains ONLY safe fields: no RFC, email, phone, password, or product data.
type ProviderMapPin struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	City         *string   `json:"city"`
	State        *string   `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
}

// SystemUser represents an admin user in the system_users table.
type SystemUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdminLoginRequest is the DTO for admin authentication.
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// AdminLoginResponse is the DTO returned after successful admin login.
type AdminLoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}
