package adapters

import (
	"context"
	"net/http"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
)

// PayPalConfig holds the credentials needed to talk to PayPal.
type PayPalConfig struct {
	ClientID     string // PAYPAL_CLIENT_ID
	ClientSecret string // PAYPAL_CLIENT_SECRET
	WebhookID    string // PAYPAL_WEBHOOK_ID (used to verify webhook signatures)
	Env          string // "sandbox" | "live"
}

// PayPalGateway is a SCAFFOLD/STUB implementation of PaymentGateway.
// It defines the integration surface; the real SDK calls are left as TODOs.
type PayPalGateway struct {
	cfg PayPalConfig
}

// NewPayPalGateway creates a PayPal gateway adapter.
func NewPayPalGateway(cfg PayPalConfig) *PayPalGateway {
	return &PayPalGateway{cfg: cfg}
}

// Name returns the gateway identifier.
func (g *PayPalGateway) Name() string { return "paypal" }

// CreateSubscriptionCheckout starts a PayPal subscription checkout.
//
// TODO(payments): implement with the PayPal Subscriptions API:
//  1. Obtain an OAuth2 access token (ClientID/ClientSecret against Env base URL).
//  2. Create a subscription for the billing plan mapped from in.PlanCode.
//  3. Pass custom_id = in.ProviderID so webhooks can resolve the provider.
//  4. Return the approval link (rel="approve") in CheckoutOutput.RedirectURL.
func (g *PayPalGateway) CreateSubscriptionCheckout(_ context.Context, _ entities.CheckoutInput) (entities.CheckoutOutput, error) {
	return entities.CheckoutOutput{}, domainErrors.NewDomainError(
		domainErrors.ErrNotImplemented, "Integración con PayPal pendiente (stub)")
}

// VerifyWebhook validates the PayPal webhook signature and normalizes the payload.
//
// TODO(payments): verify using the /v1/notifications/verify-webhook-signature
// endpoint with g.cfg.WebhookID and the transmission headers, then map event
// types (e.g. BILLING.SUBSCRIPTION.ACTIVATED / .CANCELLED / PAYMENT.SALE.DENIED)
// into entities.WebhookEvent, reading provider_id from custom_id.
func (g *PayPalGateway) VerifyWebhook(_ []byte, _ http.Header) (entities.WebhookEvent, error) {
	return entities.WebhookEvent{}, domainErrors.NewDomainError(
		domainErrors.ErrNotImplemented, "Verificación de webhook de PayPal pendiente (stub)")
}
