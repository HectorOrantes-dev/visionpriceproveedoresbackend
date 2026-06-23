package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const brevoEndpoint = "https://api.brevo.com/v3/smtp/email"

// BrevoConfig holds the settings for the Brevo (Sendinblue) HTTP email API.
// Brevo sends over HTTPS (443), so it works where SMTP ports are blocked (Railway).
// FromEmail must be a verified sender in your Brevo account.
type BrevoConfig struct {
	APIKey    string // BREVO_API_KEY
	FromEmail string // BREVO_FROM_EMAIL (a verified sender)
	FromName  string // BREVO_FROM_NAME (display name)
}

// BrevoOTPNotifier delivers OTP codes via the Brevo API.
type BrevoOTPNotifier struct {
	cfg    BrevoConfig
	client *http.Client
}

// NewBrevoOTPNotifier creates a Brevo-backed OTP notifier.
func NewBrevoOTPNotifier(cfg BrevoConfig) *BrevoOTPNotifier {
	return &BrevoOTPNotifier{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendOTP sends the OTP code to the recipient via Brevo.
func (n *BrevoOTPNotifier) SendOTP(ctx context.Context, email, name, code string, expirationMinutes int) error {
	if email == "" {
		return fmt.Errorf("brevo: empty recipient email")
	}

	greeting := "Hola"
	if name != "" {
		greeting = "Hola " + name
	}
	text := fmt.Sprintf(
		"%s,\n\nTu código de verificación de VisionPrice es:\n\n    %s\n\n"+
			"Este código expira en %d minuto(s). Si no solicitaste este código, ignora este correo.\n",
		greeting, code, expirationMinutes)

	payload := map[string]any{
		"sender":      map[string]string{"name": n.cfg.FromName, "email": n.cfg.FromEmail},
		"to":          []map[string]string{{"email": email}},
		"subject":     "Tu código de verificación VisionPrice",
		"textContent": text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("brevo: failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("brevo: failed to build request: %w", err)
	}
	req.Header.Set("api-key", n.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("brevo: API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
