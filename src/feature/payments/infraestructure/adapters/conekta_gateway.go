package adapters

import (
	"context"
	"net/http"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
)

// ConektaConfig holds the credentials needed to talk to Conekta.
type ConektaConfig struct {
	PrivateKey    string // CONEKTA_PRIVATE_KEY
	WebhookSecret string // CONEKTA_WEBHOOK_SECRET (HMAC signature verification)
}

// ConektaGateway is a SCAFFOLD/STUB implementation of PaymentGateway.
// It defines the integration surface; the real SDK calls are left as TODOs.
type ConektaGateway struct {
	cfg ConektaConfig
}

// NewConektaGateway creates a Conekta gateway adapter.
func NewConektaGateway(cfg ConektaConfig) *ConektaGateway {
	return &ConektaGateway{cfg: cfg}
}

// Name returns the gateway identifier.
func (g *ConektaGateway) Name() string { return "conekta" }

// CreateSubscriptionCheckout starts a Conekta subscription checkout.
//
// TODO(payments): implement with the Conekta API/SDK:
//  1. Find or create a Customer for in.ProviderID (store conekta customer id).
//  2. Create a Subscription on the plan mapped from in.PlanCode.
//  3. Attach metadata{provider_id, plan_code} so webhooks can resolve the provider.
//  4. Return the hosted checkout / payment URL in CheckoutOutput.RedirectURL.
func (g *ConektaGateway) CreateSubscriptionCheckout(_ context.Context, _ entities.CheckoutInput) (entities.CheckoutOutput, error) {
	return entities.CheckoutOutput{}, domainErrors.NewDomainError(
		domainErrors.ErrNotImplemented, "Integración con Conekta pendiente (stub)")
}

// VerifyWebhook validates the Conekta signature and normalizes the payload.
//
// TODO(payments): implement signature verification using g.cfg.WebhookSecret
// (Conekta sends an HMAC in the request header), then map the event type
// (e.g. "subscription.paid", "subscription.canceled") into entities.WebhookEvent,
// reading provider_id/plan_code from the object metadata.
func (g *ConektaGateway) VerifyWebhook(_ []byte, _ http.Header) (entities.WebhookEvent, error) {
	return entities.WebhookEvent{}, domainErrors.NewDomainError(
		domainErrors.ErrNotImplemented, "Verificación de webhook de Conekta pendiente (stub)")
}
