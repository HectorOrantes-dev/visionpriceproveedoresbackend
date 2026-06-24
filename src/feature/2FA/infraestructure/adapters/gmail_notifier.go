package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	gmailTokenURL = "https://oauth2.googleapis.com/token"
	gmailSendURL  = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
)

// GmailConfig holds the OAuth2 credentials for sending via the Gmail API.
// The Gmail API works over HTTPS (443) and sends from a real Gmail account, so
// it works where SMTP is blocked (Railway) and avoids spam folder issues.
//
// Obtain RefreshToken once via the OAuth2 consent flow (see docs/example.env).
type GmailConfig struct {
	ClientID     string // GMAIL_CLIENT_ID
	ClientSecret string // GMAIL_CLIENT_SECRET
	RefreshToken string // GMAIL_REFRESH_TOKEN
	From         string // GMAIL_FROM (the Gmail address, e.g. you@gmail.com)
	FromName     string // GMAIL_FROM_NAME (display name)
}

// GmailOTPNotifier delivers OTP codes via the Gmail API.
type GmailOTPNotifier struct {
	cfg    GmailConfig
	client *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

// NewGmailOTPNotifier creates a Gmail-backed OTP notifier.
func NewGmailOTPNotifier(cfg GmailConfig) *GmailOTPNotifier {
	return &GmailOTPNotifier{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendOTP sends the OTP code to the recipient via the Gmail API.
func (n *GmailOTPNotifier) SendOTP(ctx context.Context, email, name, code string, expirationMinutes int) error {
	if email == "" {
		return fmt.Errorf("gmail: empty recipient email")
	}

	token, err := n.getAccessToken(ctx)
	if err != nil {
		return err
	}

	raw := n.buildMessage(email, name, code, expirationMinutes)
	payload, err := json.Marshal(map[string]string{
		"raw": base64.URLEncoding.EncodeToString([]byte(raw)),
	})
	if err != nil {
		return fmt.Errorf("gmail: failed to encode message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gmailSendURL, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("gmail: failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("gmail: send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("gmail: send API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// buildMessage assembles the RFC 5322 message that Gmail will send.
//
// To avoid UTF-8 mojibake (e.g. "código" → "cÃ³digo"), non-ASCII headers are
// RFC 2047 encoded and the body is base64-encoded with an explicit charset and
// Content-Transfer-Encoding, so the recipient decodes the bytes intact.
func (n *GmailOTPNotifier) buildMessage(email, name, code string, expirationMinutes int) string {
	from := n.cfg.From
	if n.cfg.FromName != "" {
		// mime.QEncoding only encodes when the value has non-ASCII chars.
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", n.cfg.FromName), n.cfg.From)
	}
	greeting := "Hola"
	if name != "" {
		greeting = "Hola " + name
	}
	body := fmt.Sprintf(
		"%s,\r\n\r\nTu código de verificación de VisionPrice es:\r\n\r\n    %s\r\n\r\n"+
			"Este código expira en %d minuto(s). Si no solicitaste este código, ignora este correo.\r\n",
		greeting, code, expirationMinutes)

	subject := mime.QEncoding.Encode("UTF-8", "Tu código de verificación VisionPrice")

	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\n", from)
	fmt.Fprintf(&msg, "To: %s\r\n", email)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(&msg, "\r\n%s", wrapBase64([]byte(body)))
	return msg.String()
}

// wrapBase64 base64-encodes data and wraps it at 76 characters per RFC 2045.
func wrapBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var b strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		b.WriteString(encoded[i:end])
		b.WriteString("\r\n")
	}
	return b.String()
}

// getAccessToken returns a valid OAuth2 access token, refreshing it via the
// refresh token when the cached one is missing or about to expire.
func (n *GmailOTPNotifier) getAccessToken(ctx context.Context) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.accessToken != "" && time.Now().Before(n.tokenExpiry) {
		return n.accessToken, nil
	}

	form := url.Values{}
	form.Set("client_id", n.cfg.ClientID)
	form.Set("client_secret", n.cfg.ClientSecret)
	form.Set("refresh_token", n.cfg.RefreshToken)
	form.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gmailTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("gmail: failed to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gmail: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gmail: token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
		return "", fmt.Errorf("gmail: invalid token response")
	}

	n.accessToken = tok.AccessToken
	// Refresh a minute early to avoid using a token that expires mid-request.
	n.tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn-60) * time.Second)
	return n.accessToken, nil
}
