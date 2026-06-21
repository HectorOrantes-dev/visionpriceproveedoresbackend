package domain

import (
	"context"

	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// AdminRepository defines the port for admin-related persistence operations.
type AdminRepository interface {
	// GetGlobalMetrics retrieves system-wide metrics (constructor count, provider count, active users in 24h).
	GetGlobalMetrics(ctx context.Context) (*entities.GlobalMetrics, error)

	// GetProviderMapPins retrieves all active provider locations for the admin map.
	// Returns ONLY public-safe data (no RFC, email, phone, prices, or catalogs).
	GetProviderMapPins(ctx context.Context) ([]*entities.ProviderMapPin, error)

	// FindSystemUserByEmail retrieves a system admin user by email for authentication.
	FindSystemUserByEmail(ctx context.Context, email string) (*entities.SystemUser, error)

	// GetExpiringSubscriptions retrieves paid subscriptions whose current period
	// ends within the next `withinDays` days (includes already-overdue ones),
	// joined with the provider and plan info. Free plans (no expiry) are excluded.
	GetExpiringSubscriptions(ctx context.Context, withinDays int) ([]*entities.ExpiringSubscription, error)
}
