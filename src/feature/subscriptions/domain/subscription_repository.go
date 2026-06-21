package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain/entities"
)

// SubscriptionRepository is the port for subscription persistence.
type SubscriptionRepository interface {
	// GetByProvider returns the provider's current subscription.
	GetByProvider(ctx context.Context, providerID uuid.UUID) (*entities.ProviderSubscription, error)

	// GetPlan returns a single plan by its code.
	GetPlan(ctx context.Context, code string) (*entities.Plan, error)

	// ListPlans returns all active plans ordered by price.
	ListPlans(ctx context.Context) ([]*entities.Plan, error)

	// EnsureDefault creates a free/active subscription for the provider if none exists.
	EnsureDefault(ctx context.Context, providerID uuid.UUID) error

	// CountActiveProducts returns how many active products the provider currently has.
	CountActiveProducts(ctx context.Context, providerID uuid.UUID) (int, error)

	// UpsertFromWebhook applies a payment webhook update to the provider's subscription.
	UpsertFromWebhook(ctx context.Context, in entities.WebhookUpdate) error
}
