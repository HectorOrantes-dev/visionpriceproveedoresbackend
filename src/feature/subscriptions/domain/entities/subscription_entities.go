package entities

import (
	"time"

	"github.com/google/uuid"
)

// Plan is a subscription tier from the catalog. A nil ProductLimit means unlimited.
type Plan struct {
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	ProductLimit *int           `json:"product_limit"`
	PriceCents   int            `json:"price_cents"`
	Currency     string         `json:"currency"`
	Features     map[string]any `json:"features"`
}

// ProviderSubscription is a provider's current subscription state.
type ProviderSubscription struct {
	ProviderID             uuid.UUID  `json:"provider_id"`
	PlanCode               string     `json:"plan_code"`
	Status                 string     `json:"status"`
	PaymentProvider        *string    `json:"payment_provider,omitempty"`
	ExternalCustomerID     *string    `json:"-"`
	ExternalSubscriptionID *string    `json:"-"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
}

// SubscriptionView is the read model returned by GET /subscription: the plan plus
// the provider's current usage against the limit.
type SubscriptionView struct {
	Plan             Plan       `json:"plan"`
	Status           string     `json:"status"`
	ProductCount     int        `json:"product_count"`
	ProductLimit     *int       `json:"product_limit"`
	Unlimited        bool       `json:"unlimited"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
}

// WebhookUpdate carries the fields a payment webhook applies to a subscription.
type WebhookUpdate struct {
	ProviderID             uuid.UUID
	PlanCode               string
	Status                 string
	PaymentProvider        string
	ExternalCustomerID     string
	ExternalSubscriptionID string
	CurrentPeriodEnd       *time.Time
}
