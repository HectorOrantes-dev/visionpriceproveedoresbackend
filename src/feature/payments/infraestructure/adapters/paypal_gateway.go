package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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

// PayPalGateway integrates with the PayPal Subscriptions REST API.
type PayPalGateway struct {
	cfg      PayPalConfig
	plans    map[string]string // planCode -> PayPal plan id (P-xxxx)
	planByID map[string]string // PayPal plan id -> planCode (reverse)
	baseURL  string
}

// NewPayPalGateway creates a PayPal gateway adapter. plans maps our plan codes
// (pro/max) to PayPal billing plan ids.
func NewPayPalGateway(cfg PayPalConfig, plans map[string]string) *PayPalGateway {
	base := "https://api-m.paypal.com"
	if !strings.EqualFold(cfg.Env, "live") {
		base = "https://api-m.sandbox.paypal.com"
	}
	reverse := make(map[string]string, len(plans))
	for code, id := range plans {
		if id != "" {
			reverse[id] = code
		}
	}
	return &PayPalGateway{cfg: cfg, plans: plans, planByID: reverse, baseURL: base}
}

// Name returns the gateway identifier.
func (g *PayPalGateway) Name() string { return "paypal" }

// CreateSubscriptionCheckout creates a PayPal subscription and returns the
// approval URL the customer must visit to authorize the recurring payment.
func (g *PayPalGateway) CreateSubscriptionCheckout(ctx context.Context, in entities.CheckoutInput) (entities.CheckoutOutput, error) {
	planID := g.plans[in.PlanCode]
	if planID == "" {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrValidation,
			fmt.Sprintf("Plan '%s' no está configurado para PayPal", in.PlanCode))
	}

	token, err := g.accessToken(ctx)
	if err != nil {
		return entities.CheckoutOutput{}, err
	}

	reqBody := map[string]any{
		"plan_id":   planID,
		"custom_id": in.ProviderID,
		"application_context": map[string]any{
			"brand_name":  "VisionPrice",
			"user_action": "SUBSCRIBE_NOW",
			"return_url":  in.SuccessURL,
			"cancel_url":  in.CancelURL,
		},
	}

	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}

	_, err = doJSON(ctx, http.MethodPost, g.baseURL+"/v1/billing/subscriptions",
		map[string]string{
			"Authorization": "Bearer " + token,
			"Prefer":        "return=representation",
		}, reqBody, &resp)
	if err != nil {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrInternal,
			"Error al crear la suscripción en PayPal: "+err.Error())
	}

	approveURL := ""
	for _, l := range resp.Links {
		if strings.EqualFold(l.Rel, "approve") {
			approveURL = l.Href
			break
		}
	}
	if approveURL == "" {
		return entities.CheckoutOutput{}, domainErrors.NewDomainError(domainErrors.ErrInternal,
			"PayPal no devolvió un enlace de aprobación")
	}

	return entities.CheckoutOutput{
		RedirectURL: approveURL,
		Reference:   resp.ID,
		Status:      "pending",
	}, nil
}

// VerifyWebhook validates the PayPal signature via the verify-webhook-signature
// endpoint and normalizes the payload into a WebhookEvent.
func (g *PayPalGateway) VerifyWebhook(rawBody []byte, headers http.Header) (entities.WebhookEvent, error) {
	ctx := context.Background()

	token, err := g.accessToken(ctx)
	if err != nil {
		return entities.WebhookEvent{}, err
	}

	verifyReq := map[string]any{
		"auth_algo":         headers.Get("Paypal-Auth-Algo"),
		"cert_url":          headers.Get("Paypal-Cert-Url"),
		"transmission_id":   headers.Get("Paypal-Transmission-Id"),
		"transmission_sig":  headers.Get("Paypal-Transmission-Sig"),
		"transmission_time": headers.Get("Paypal-Transmission-Time"),
		"webhook_id":        g.cfg.WebhookID,
		"webhook_event":     json.RawMessage(rawBody),
	}

	var verifyResp struct {
		VerificationStatus string `json:"verification_status"`
	}
	if _, err := doJSON(ctx, http.MethodPost, g.baseURL+"/v1/notifications/verify-webhook-signature",
		map[string]string{"Authorization": "Bearer " + token}, verifyReq, &verifyResp); err != nil {
		return entities.WebhookEvent{}, domainErrors.NewDomainError(domainErrors.ErrInternal,
			"Error al verificar la firma del webhook de PayPal: "+err.Error())
	}
	if !strings.EqualFold(verifyResp.VerificationStatus, "SUCCESS") {
		return entities.WebhookEvent{}, domainErrors.NewDomainError(domainErrors.ErrUnauthorized,
			"Firma del webhook de PayPal inválida")
	}

	var ev struct {
		ID        string `json:"id"`
		EventType string `json:"event_type"`
		Resource  struct {
			ID          string `json:"id"`
			CustomID    string `json:"custom_id"`
			PlanID      string `json:"plan_id"`
			BillingInfo struct {
				NextBillingTime *time.Time `json:"next_billing_time"`
			} `json:"billing_info"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(rawBody, &ev); err != nil {
		return entities.WebhookEvent{}, domainErrors.NewDomainError(domainErrors.ErrValidation,
			"Cuerpo del webhook de PayPal inválido")
	}

	out := entities.WebhookEvent{
		Provider:               "paypal",
		ExternalEventID:        ev.ID,
		Type:                   ev.EventType,
		ProviderID:             ev.Resource.CustomID,
		PlanCode:               g.planByID[ev.Resource.PlanID],
		ExternalSubscriptionID: ev.Resource.ID,
		CurrentPeriodEnd:       ev.Resource.BillingInfo.NextBillingTime,
	}

	switch ev.EventType {
	case "BILLING.SUBSCRIPTION.ACTIVATED", "BILLING.SUBSCRIPTION.RE-ACTIVATED", "PAYMENT.SALE.COMPLETED":
		out.Status = "active"
	case "BILLING.SUBSCRIPTION.CANCELLED", "BILLING.SUBSCRIPTION.EXPIRED":
		out.Status = "canceled"
	case "BILLING.SUBSCRIPTION.SUSPENDED", "BILLING.SUBSCRIPTION.PAYMENT.FAILED", "PAYMENT.SALE.DENIED":
		out.Status = "past_due"
	default:
		out.Ignored = true // event we don't act on (still recorded for audit)
	}

	return out, nil
}

// accessToken fetches an OAuth2 client-credentials token from PayPal.
func (g *PayPalGateway) accessToken(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.baseURL+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al preparar la autenticación con PayPal")
	}
	basic := base64.StdEncoding.EncodeToString([]byte(g.cfg.ClientID + ":" + g.cfg.ClientSecret))
	req.Header.Set("Authorization", "Basic "+basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al autenticar con PayPal: "+err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized,
			fmt.Sprintf("PayPal rechazó las credenciales (status %d)", resp.StatusCode))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Respuesta de token de PayPal inválida")
	}
	return tok.AccessToken, nil
}
