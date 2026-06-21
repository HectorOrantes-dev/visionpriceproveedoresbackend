package payment_usecase

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
	subscriptionEntities "github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain/entities"
)

// Config holds runtime settings for the payments use case.
type Config struct {
	Enabled        bool
	DefaultGateway string
	SuccessURL     string
	CancelURL      string
}

// PaymentUseCase orchestrates checkout creation and webhook processing.
type PaymentUseCase struct {
	cfg      Config
	gateways map[string]domain.PaymentGateway
	events   domain.EventStore
	subs     domain.SubscriptionUpdater
}

// NewPaymentUseCase creates a new PaymentUseCase. gateways is keyed by Name().
func NewPaymentUseCase(cfg Config, gateways map[string]domain.PaymentGateway, events domain.EventStore, subs domain.SubscriptionUpdater) *PaymentUseCase {
	return &PaymentUseCase{cfg: cfg, gateways: gateways, events: events, subs: subs}
}

// Enabled reports whether payment processing is turned on.
func (uc *PaymentUseCase) Enabled() bool { return uc.cfg.Enabled }

// CreateCheckout starts a subscription checkout for the provider on the chosen
// (or default) gateway and returns where to redirect them.
func (uc *PaymentUseCase) CreateCheckout(ctx context.Context, providerID, planCode, gatewayName string) (entities.CheckoutOutput, error) {
	if gatewayName == "" {
		gatewayName = uc.cfg.DefaultGateway
	}
	gw, ok := uc.gateways[gatewayName]
	if !ok {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrValidation, "Pasarela de pago no soportada")
	}

	return gw.CreateSubscriptionCheckout(ctx, entities.CheckoutInput{
		ProviderID: providerID,
		PlanCode:   planCode,
		SuccessURL: uc.cfg.SuccessURL,
		CancelURL:  uc.cfg.CancelURL,
	})
}

// HandleWebhook verifies, deduplicates, and applies a gateway webhook.
func (uc *PaymentUseCase) HandleWebhook(ctx context.Context, gatewayName string, rawBody []byte, headers http.Header) error {
	gw, ok := uc.gateways[gatewayName]
	if !ok {
		return domainErrors.NewDomainError(domainErrors.ErrNotFound, "Pasarela de pago desconocida")
	}

	// 1. Verify signature + normalize. A failure here means the request is
	//    unsigned/forged or the adapter is still a stub.
	ev, err := gw.VerifyWebhook(rawBody, headers)
	if err != nil {
		return err
	}

	// 2. Idempotency: record once. Retries from the gateway become no-ops.
	isNew, err := uc.events.RecordIfNew(ctx, gatewayName, ev.ExternalEventID, ev.Type, rawBody)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al registrar el evento de pago")
	}
	if !isNew {
		return nil // already processed
	}

	// 3. Events that don't change subscription state are recorded but ignored.
	if ev.Ignored {
		return uc.events.MarkProcessed(ctx, gatewayName, ev.ExternalEventID)
	}

	pid, err := uuid.Parse(ev.ProviderID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "provider_id inválido en el evento")
	}

	// 4. Apply to the subscription.
	if err := uc.subs.ApplyWebhook(ctx, subscriptionEntities.WebhookUpdate{
		ProviderID:             pid,
		PlanCode:               ev.PlanCode,
		Status:                 ev.Status,
		PaymentProvider:        ev.Provider,
		ExternalCustomerID:     ev.ExternalCustomerID,
		ExternalSubscriptionID: ev.ExternalSubscriptionID,
		CurrentPeriodEnd:       ev.CurrentPeriodEnd,
	}); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al aplicar el evento a la suscripción")
	}

	return uc.events.MarkProcessed(ctx, gatewayName, ev.ExternalEventID)
}
