package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
)

const conektaBaseURL = "https://api.conekta.io"
const conektaAccept = "application/vnd.conekta-v2.1.0+json"

// ConektaConfig holds the credentials needed to talk to Conekta.
type ConektaConfig struct {
	PrivateKey    string // CONEKTA_PRIVATE_KEY (also authenticates webhook re-fetch)
	WebhookSecret string // CONEKTA_WEBHOOK_SECRET (reserved; we verify by event re-fetch)
}

// ConektaGateway integrates with the Conekta API for recurring subscriptions.
type ConektaGateway struct {
	cfg      ConektaConfig
	plans    map[string]string // planCode -> Conekta plan id
	planByID map[string]string // Conekta plan id -> planCode (reverse)
}

// NewConektaGateway creates a Conekta gateway adapter. plans maps our plan codes
// (pro/max) to Conekta plan ids (the "Referencia / ID interno" in the dashboard).
func NewConektaGateway(cfg ConektaConfig, plans map[string]string) *ConektaGateway {
	reverse := make(map[string]string, len(plans))
	for code, id := range plans {
		if id != "" {
			reverse[id] = code
		}
	}
	return &ConektaGateway{cfg: cfg, plans: plans, planByID: reverse}
}

// Name returns the gateway identifier.
func (g *ConektaGateway) Name() string { return "conekta" }

func (g *ConektaGateway) authHeaders() map[string]string {
	basic := base64.StdEncoding.EncodeToString([]byte(g.cfg.PrivateKey + ":"))
	return map[string]string{
		"Authorization": "Basic " + basic,
		"Accept":        conektaAccept,
	}
}

// CreateSubscriptionCheckout creates a Conekta customer with the card token and
// subscribes it to the plan. The card must be tokenized on the frontend with
// Conekta.js; in.PaymentToken carries that token.
func (g *ConektaGateway) CreateSubscriptionCheckout(ctx context.Context, in entities.CheckoutInput) (entities.CheckoutOutput, error) {
	planID := g.plans[in.PlanCode]
	if planID == "" {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrValidation,
			fmt.Sprintf("Plan '%s' no está configurado para Conekta", in.PlanCode))
	}
	if in.PaymentToken == "" {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrValidation,
			"Se requiere un token de tarjeta de Conekta (tokeniza la tarjeta en el frontend con Conekta.js)")
	}

	// 1. Create the customer with the card token as its default payment source.
	customerReq := map[string]any{
		"name":  in.CustomerName,
		"email": in.CustomerEmail,
		"payment_sources": []map[string]any{
			{"type": "card", "token_id": in.PaymentToken},
		},
		"metadata": map[string]any{"provider_id": in.ProviderID},
	}
	var customer struct {
		ID string `json:"id"`
	}
	if _, err := doJSON(ctx, http.MethodPost, conektaBaseURL+"/customers",
		g.authHeaders(), customerReq, &customer); err != nil {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrInternal,
			"Error al crear el cliente en Conekta: "+err.Error())
	}

	// 2. Subscribe the customer to the plan.
	var sub struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if _, err := doJSON(ctx, http.MethodPost, conektaBaseURL+"/customers/"+customer.ID+"/subscription",
		g.authHeaders(), map[string]any{"plan_id": planID}, &sub); err != nil {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrInternal,
			"Error al crear la suscripción en Conekta: "+err.Error())
	}

	return entities.CheckoutOutput{
		Reference:          sub.ID,
		Status:             mapConektaStatus(sub.Status),
		ExternalCustomerID: customer.ID,
	}, nil
}

// VerifyWebhook authenticates a Conekta webhook. The ping event is acknowledged
// directly; real events are verified by re-fetching them from the Conekta API
// (only our private key can read genuine event ids), then normalized.
func (g *ConektaGateway) VerifyWebhook(rawBody []byte, _ http.Header) (entities.WebhookEvent, error) {
	var head struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(rawBody, &head); err != nil {
		return entities.WebhookEvent{}, domainErrors.NewDomainError(domainErrors.ErrValidation,
			"Cuerpo del webhook de Conekta inválido")
	}

	// Acknowledge (200) anything that isn't a subscription lifecycle event WITHOUT
	// calling the API: the connectivity ping ("webhook_ping"), the validation POST
	// Conekta sends when a webhook is created (object "webhook", status
	// "being_pinged", with no "type"), and webhook.*/payout.* meta events. None of
	// these change subscription state, so no verification or private key is needed.
	if head.Action == "webhook_ping" || !strings.HasPrefix(head.Type, "subscription.") {
		return entities.WebhookEvent{Provider: "conekta", ExternalEventID: head.ID, Type: head.Type, Ignored: true}, nil
	}

	// subscription.* events change state: verify authenticity by re-fetching the
	// event by id (a forged id won't exist for our private key), then map it.
	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID              string `json:"id"`
				Status          string `json:"status"`
				PlanID          string `json:"plan_id"`
				CustomerID      string `json:"customer_id"`
				BillingCycleEnd int64  `json:"billing_cycle_end"`
			} `json:"object"`
		} `json:"data"`
	}
	if _, err := doJSON(context.Background(), http.MethodGet, conektaBaseURL+"/events/"+head.ID,
		g.authHeaders(), nil, &event); err != nil {
		return entities.WebhookEvent{}, domainErrors.NewDomainError(domainErrors.ErrUnauthorized,
			"No se pudo verificar el evento de Conekta: "+err.Error())
	}

	// We only act on subscription lifecycle events.
	if !strings.HasPrefix(event.Type, "subscription.") {
		return entities.WebhookEvent{Provider: "conekta", ExternalEventID: event.ID, Type: event.Type, Ignored: true}, nil
	}

	obj := event.Data.Object
	out := entities.WebhookEvent{
		Provider:               "conekta",
		ExternalEventID:        event.ID,
		Type:                   event.Type,
		PlanCode:               g.planByID[obj.PlanID],
		Status:                 mapConektaStatus(obj.Status),
		ExternalCustomerID:     obj.CustomerID,
		ExternalSubscriptionID: obj.ID,
	}
	if obj.BillingCycleEnd > 0 {
		t := time.Unix(obj.BillingCycleEnd, 0).UTC()
		out.CurrentPeriodEnd = &t
	}
	return out, nil
}

// mapConektaStatus maps a Conekta subscription status to our internal status.
func mapConektaStatus(status string) string {
	switch status {
	case "active", "in_trial":
		return "active"
	case "past_due", "paused":
		return "past_due"
	case "canceled", "expired":
		return "canceled"
	default:
		return status
	}
}
