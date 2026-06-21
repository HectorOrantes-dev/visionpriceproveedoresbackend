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
	cfg       Config
	gateways  map[string]domain.PaymentGateway
	events    domain.EventStore
	subs      domain.SubscriptionUpdater
	directory domain.ProviderDirectory
}

// NewPaymentUseCase creates a new PaymentUseCase. gateways is keyed by Name().
func NewPaymentUseCase(cfg Config, gateways map[string]domain.PaymentGateway, events domain.EventStore, subs domain.SubscriptionUpdater, directory domain.ProviderDirectory) *PaymentUseCase {
	return &PaymentUseCase{cfg: cfg, gateways: gateways, events: events, subs: subs, directory: directory}
}

// Enabled reports whether payment processing is turned on.
func (uc *PaymentUseCase) Enabled() bool { return uc.cfg.Enabled }

// CreateCheckout starts a subscription checkout for the provider on the chosen
// (or default) gateway. For gateways that activate synchronously (Conekta) the
// subscription is applied immediately; redirect-based gateways (PayPal) are
// finalized later by their webhook.
func (uc *PaymentUseCase) CreateCheckout(ctx context.Context, providerID, planCode, gatewayName, paymentToken string) (entities.CheckoutOutput, error) {
	if gatewayName == "" {
		gatewayName = uc.cfg.DefaultGateway
	}
	gw, ok := uc.gateways[gatewayName]
	if !ok {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrValidation, "Pasarela de pago no soportada")
	}

	name, email, err := uc.directory.GetProviderContact(ctx, providerID)
	if err != nil {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener los datos del proveedor")
	}

	out, err := gw.CreateSubscriptionCheckout(ctx, entities.CheckoutInput{
		ProviderID:    providerID,
		PlanCode:      planCode,
		SuccessURL:    uc.cfg.SuccessURL,
		CancelURL:     uc.cfg.CancelURL,
		PaymentToken:  paymentToken,
		CustomerName:  name,
		CustomerEmail: email,
	})
	if err != nil {
		return entities.CheckoutOutput{}, err
	}

	// Synchronous activation (Conekta): persist the upgrade now and store the
	// external ids so future renewal/cancel webhooks resolve to this provider.
	if out.Status == "active" && out.Reference != "" {
		pid, perr := uuid.Parse(providerID)
		if perr == nil {
			_ = uc.subs.ApplyWebhook(ctx, subscriptionEntities.WebhookUpdate{
				ProviderID:             pid,
				PlanCode:               planCode,
				Status:                 "active",
				PaymentProvider:        gatewayName,
				ExternalCustomerID:     out.ExternalCustomerID,
				ExternalSubscriptionID: out.Reference,
			})
		}
	}

	return out, nil
}

// HandleWebhook verifies, deduplicates, and applies a gateway webhook.
func (uc *PaymentUseCase) HandleWebhook(ctx context.Context, gatewayName string, rawBody []byte, headers http.Header) error {
	gw, ok := uc.gateways[gatewayName]
	if !ok {
		return domainErrors.NewDomainError(domainErrors.ErrNotFound, "Pasarela de pago desconocida")
	}

	// 1. Verify signature/authenticity + normalize.
	ev, err := gw.VerifyWebhook(rawBody, headers)
	if err != nil {
		return err
	}

	// 2. Idempotency (when the event has an id). Retries become no-ops.
	if ev.ExternalEventID != "" {
		isNew, err := uc.events.RecordIfNew(ctx, gatewayName, ev.ExternalEventID, ev.Type, rawBody)
		if err != nil {
			return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al registrar el evento de pago")
		}
		if !isNew {
			return nil // already processed
		}
	}

	// 3. Events that don't change subscription state are acknowledged (e.g. ping).
	if ev.Ignored {
		if ev.ExternalEventID != "" {
			return uc.events.MarkProcessed(ctx, gatewayName, ev.ExternalEventID)
		}
		return nil
	}

	// 4. Resolve the provider. PayPal carries it (custom_id); Conekta needs a
	//    lookup by the stored external subscription id.
	providerID := ev.ProviderID
	if providerID == "" && ev.ExternalSubscriptionID != "" {
		resolved, rerr := uc.subs.ResolveProviderByExternalSubscription(ctx, ev.ExternalSubscriptionID)
		if rerr != nil {
			return domainErrors.NewDomainError(domainErrors.ErrNotFound, "No se encontró el proveedor para esta suscripción")
		}
		providerID = resolved
	}
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "No se pudo resolver el proveedor del evento")
	}

	// 5. Apply to the subscription.
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

	if ev.ExternalEventID != "" {
		return uc.events.MarkProcessed(ctx, gatewayName, ev.ExternalEventID)
	}
	return nil
}
