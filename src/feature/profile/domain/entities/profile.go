package entities

import (
	"time"

	"github.com/google/uuid"
)

// Profile is the authenticated provider's own profile data.
type Profile struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	RFC          string    `json:"rfc"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
}

// UpdateProfileRequest is the DTO for updating provider profile fields.
type UpdateProfileRequest struct {
	BusinessName *string `json:"business_name" binding:"omitempty,min=2,max=255,nohtml"`
	Email        *string `json:"email" binding:"omitempty,email"`
	Phone        *string `json:"phone" binding:"omitempty,min=10,max=20,nohtml"`
}
