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
