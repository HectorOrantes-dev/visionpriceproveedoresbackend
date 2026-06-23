package adapters

import (
	"context"
	"log"
)

// LogOTPNotifier is a development fallback that logs the OTP instead of emailing
// it. It is used when no email provider is configured, so local development keeps
// working without external credentials.
type LogOTPNotifier struct{}

// NewLogOTPNotifier creates a notifier that only logs the code.
func NewLogOTPNotifier() *LogOTPNotifier { return &LogOTPNotifier{} }

// SendOTP logs the OTP code to stdout.
func (n *LogOTPNotifier) SendOTP(_ context.Context, email, _, code string, expirationMinutes int) error {
	log.Printf("🔐 [EMAIL NO CONFIGURADO] OTP para %s: %s (expira en %d min)", email, code, expirationMinutes)
	return nil
}
