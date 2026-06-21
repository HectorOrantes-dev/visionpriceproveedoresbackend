package entities

import "time"

// CheckoutInput is what a gateway needs to start a subscription checkout.
//
// The two gateways differ: PayPal uses a redirect-approval flow (no card data
// reaches us), while Conekta charges a card token that the frontend tokenizes
// with Conekta.js. The optional fields below cover both.
type CheckoutInput struct {
	ProviderID string
	PlanCode   string
	SuccessURL string
	CancelURL  string

	// Conekta-specific: card token created client-side with Conekta.js, plus the
	// customer contact required to create a Conekta customer.
	PaymentToken  string
	CustomerName  string
	CustomerEmail string
}

// CheckoutOutput is the result of starting a checkout.
//   - PayPal: RedirectURL points to PayPal's approval page; Status is "pending".
//   - Conekta: the subscription is created immediately; Status reflects it
//     (e.g. "active") and RedirectURL is empty.
type CheckoutOutput struct {
	RedirectURL string `json:"redirect_url,omitempty"`
	Reference   string `json:"reference,omitempty"` // gateway subscription id
	Status      string `json:"status,omitempty"`
	// ExternalCustomerID is persisted by the use case so later webhooks resolve.
	ExternalCustomerID string `json:"-"`
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
	// PaymentToken is required only for Conekta (card token from Conekta.js).
	PaymentToken string `json:"payment_token" binding:"omitempty"`
}
