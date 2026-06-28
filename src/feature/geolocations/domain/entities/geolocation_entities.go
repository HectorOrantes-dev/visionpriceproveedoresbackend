package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProviderLocation represents the warehouse address of a provider.
type ProviderLocation struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Address    string    `json:"address"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SetLocationRequest is the DTO for setting/updating provider location.
type SetLocationRequest struct {
	Address string `json:"address" binding:"required,min=3,max=500"`
}
