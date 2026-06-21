package entities

import "time"

// CheckoutInput is what a gateway needs to start a subscription checkout.
type CheckoutInput struct {
	ProviderID string
	PlanCode   string
	SuccessURL string
	CancelURL  string
}

// CheckoutOutput is the result of starting a checkout: where to send the user.
type CheckoutOutput struct {
	RedirectURL string `json:"redirect_url"`
	Reference   string `json:"reference,omitempty"`
}

// WebhookEvent is the normalized, gateway-agnostic representation of a payment
// webhook. Each gateway adapter is responsible for verifying the signature and
// mapping its native payload into this shape.
type WebhookEvent struct {
	Provider               string     // "conekta" | "paypal"
	ExternalEventID        string     // gateway event id (idempotency key)
	Type                   string     // gateway event type
	ProviderID             string     // VisionPrice provider UUID (from metadata)
	PlanCode               string     // target plan: free | pro | max
	Status                 string     // active | past_due | canceled
	ExternalCustomerID     string     // gateway customer id
	ExternalSubscriptionID string     // gateway subscription id
	CurrentPeriodEnd       *time.Time // when the current paid period ends
	Ignored                bool       // true if this event type doesn't affect subscription state
}

// CheckoutRequest is the HTTP DTO for POST /subscription/checkout.
type CheckoutRequest struct {
	PlanCode string `json:"plan_code" binding:"required,oneof=pro max"`
	Gateway  string `json:"gateway" binding:"omitempty,oneof=conekta paypal"`
}
