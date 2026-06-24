package geolocation_usecase

import (
	"context"
	"math"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain"
	"github.com/visionprice/proveedores-backend/src/feature/geolocations/domain/entities"
)

const earthRadiusKm = 6371.0

// Default location (Ciudad de México) used when the client omits coordinates.
const (
	defaultLat = 19.4326
	defaultLng = -99.1332
)

// GeolocationUseCase contains business logic for geolocation operations.
type GeolocationUseCase struct {
	repo domain.GeolocationRepository
}

// NewGeolocationUseCase creates a new GeolocationUseCase.
func NewGeolocationUseCase(repo domain.GeolocationRepository) *GeolocationUseCase {
	return &GeolocationUseCase{repo: repo}
}

// SetLocation creates or updates the provider's location and delivery radius.
func (uc *GeolocationUseCase) SetLocation(ctx context.Context, providerID string, req *entities.SetLocationRequest) (*entities.ProviderLocation, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	// Default to Ciudad de México when coordinates are not provided.
	lat := defaultLat
	if req.Lat != nil {
		lat = *req.Lat
	}
	lng := defaultLng
	if req.Lng != nil {
		lng = *req.Lng
	}

	location := &entities.ProviderLocation{
		ProviderID:       pid,
		Lat:              lat,
		Lng:              lng,
		DeliveryRadiusKm: req.DeliveryRadiusKm,
	}

	result, err := uc.repo.UpsertLocation(ctx, location)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al guardar la ubicación")
	}

	return result, nil
}

// GetLocation retrieves the provider's location.
func (uc *GeolocationUseCase) GetLocation(ctx context.Context, providerID string) (*entities.ProviderLocation, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	location, err := uc.repo.GetLocation(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Ubicación no encontrada")
	}

	return location, nil
}

// IsPointInRadius checks if a given point is within the provider's delivery radius.
// Uses the Haversine formula for great-circle distance calculation.
func (uc *GeolocationUseCase) IsPointInRadius(ctx context.Context, providerID string, lat, lng float64) (*entities.CheckPointResponse, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	location, err := uc.repo.GetLocation(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrNotFound, "Ubicación del proveedor no configurada")
	}

	distance := haversineDistance(location.Lat, location.Lng, lat, lng)

	return &entities.CheckPointResponse{
		InRadius:   distance <= location.DeliveryRadiusKm,
		DistanceKm: math.Round(distance*100) / 100,
		RadiusKm:   location.DeliveryRadiusKm,
	}, nil
}

// haversineDistance calculates the great-circle distance in km between two points.
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLng := degreesToRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
