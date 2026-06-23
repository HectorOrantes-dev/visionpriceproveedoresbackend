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

const resendEndpoint = "https://api.resend.com/emails"

// ResendConfig holds the settings for the Resend HTTP email API.
// Resend sends over HTTPS (443), so it works on platforms that block SMTP ports.
type ResendConfig struct {
	APIKey string // RESEND_API_KEY
	From   string // RESEND_FROM, e.g. "VisionPrice <onboarding@resend.dev>"
}

// ResendOTPNotifier delivers OTP codes via the Resend API.
type ResendOTPNotifier struct {
	cfg    ResendConfig
	client *http.Client
}

// NewResendOTPNotifier creates a Resend-backed OTP notifier.
func NewResendOTPNotifier(cfg ResendConfig) *ResendOTPNotifier {
	return &ResendOTPNotifier{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendOTP sends the OTP code to the recipient via Resend.
func (n *ResendOTPNotifier) SendOTP(ctx context.Context, email, name, code string, expirationMinutes int) error {
	if email == "" {
		return fmt.Errorf("resend: empty recipient email")
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
		"from":    n.cfg.From,
		"to":      []string{email},
		"subject": "Tu código de verificación VisionPrice",
		"text":    text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("resend: failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+n.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("resend: API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
