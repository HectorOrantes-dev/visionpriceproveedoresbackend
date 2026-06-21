package domain

import (
	"context"
	"net/http"

	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
	subscriptionEntities "github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain/entities"
)

// PaymentGateway is the port every payment provider (Conekta, PayPal) implements.
type PaymentGateway interface {
	// Name returns the gateway identifier ("conekta" | "paypal").
	Name() string

	// CreateSubscriptionCheckout starts a recurring subscription checkout and
	// returns where to redirect the customer to complete payment.
	CreateSubscriptionCheckout(ctx context.Context, in entities.CheckoutInput) (entities.CheckoutOutput, error)

	// VerifyWebhook validates the request signature and normalizes the native
	// payload into a WebhookEvent. It must reject unsigned/forged requests.
	VerifyWebhook(rawBody []byte, headers http.Header) (entities.WebhookEvent, error)
}

// EventStore provides webhook idempotency: an event is processed at most once.
type EventStore interface {
	// RecordIfNew inserts the event and returns isNew=false if it already existed.
	RecordIfNew(ctx context.Context, provider, externalEventID, eventType string, payload []byte) (isNew bool, err error)

	// MarkProcessed flags a recorded event as fully processed.
	MarkProcessed(ctx context.Context, provider, externalEventID string) error
}

// SubscriptionUpdater is the narrow port (consumer-defined) the payments feature
// uses to apply a verified webhook to a provider's subscription. Implemented by
// the subscriptions use case.
type SubscriptionUpdater interface {
	ApplyWebhook(ctx context.Context, update subscriptionEntities.WebhookUpdate) error

	// ResolveProviderByExternalSubscription maps a gateway subscription id back to
	// the VisionPrice provider id. Needed for Conekta, whose webhook payloads
	// don't carry our provider id (PayPal does, via custom_id).
	ResolveProviderByExternalSubscription(ctx context.Context, externalSubscriptionID string) (string, error)
}

// ProviderDirectory is the narrow port used to fetch the contact data required
// to create a customer at the payment gateway (Conekta needs name + email).
type ProviderDirectory interface {
	GetProviderContact(ctx context.Context, providerID string) (name, email string, err error)
}
