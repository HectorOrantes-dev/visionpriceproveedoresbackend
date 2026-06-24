package entities

import (
	"github.com/google/uuid"
	"time"
)

// ProviderLocation represents the geographic location and delivery radius of a provider.
type ProviderLocation struct {
	ID               uuid.UUID `json:"id"`
	ProviderID       uuid.UUID `json:"provider_id"`
	Lat              float64   `json:"lat"`
	Lng              float64   `json:"lng"`
	DeliveryRadiusKm float64   `json:"delivery_radius_km"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SetLocationRequest is the DTO for setting/updating provider location.
//
// Lat/Lng are optional: when omitted, the use case applies a default location
// (Ciudad de México) so the client doesn't have to send coordinates.
type SetLocationRequest struct {
	Lat              *float64 `json:"lat" binding:"omitempty,min=-90,max=90"`
	Lng              *float64 `json:"lng" binding:"omitempty,min=-180,max=180"`
	DeliveryRadiusKm float64  `json:"delivery_radius_km" binding:"required,gt=0,max=500"`
}

// CheckPointRequest is the DTO for checking if a point is within delivery radius.
type CheckPointRequest struct {
	Lat float64 `json:"lat" binding:"required,min=-90,max=90"`
	Lng float64 `json:"lng" binding:"required,min=-180,max=180"`
}

// CheckPointResponse is returned after checking a point against the delivery radius.
type CheckPointResponse struct {
	InRadius   bool    `json:"in_radius"`
	DistanceKm float64 `json:"distance_km"`
	RadiusKm   float64 `json:"radius_km"`
}
