package adapters

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig holds the credentials and settings for an SMTP server.
// Use a STARTTLS port (typically 587); implicit-TLS port 465 is not supported
// by net/smtp's SendMail.
type SMTPConfig struct {
	Host     string // SMTP_HOST
	Port     string // SMTP_PORT (e.g. "587")
	Username string // SMTP_USERNAME
	Password string // SMTP_PASSWORD
	From     string // SMTP_FROM (sender address)
	FromName string // SMTP_FROM_NAME (display name, optional)
}

// SMTPOTPNotifier delivers OTP codes via email using an SMTP server.
type SMTPOTPNotifier struct {
	cfg SMTPConfig
}

// NewSMTPOTPNotifier creates an SMTP-backed OTP notifier.
func NewSMTPOTPNotifier(cfg SMTPConfig) *SMTPOTPNotifier {
	return &SMTPOTPNotifier{cfg: cfg}
}

// SendOTP sends the OTP code to the recipient's email address.
func (n *SMTPOTPNotifier) SendOTP(_ context.Context, email, name, code string, expirationMinutes int) error {
	if email == "" {
		return fmt.Errorf("smtp: empty recipient email")
	}

	addr := net.JoinHostPort(n.cfg.Host, n.cfg.Port)
	auth := smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.Host)

	from := n.cfg.From
	fromHeader := from
	if n.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", n.cfg.FromName, from)
	}

	greeting := "Hola"
	if name != "" {
		greeting = "Hola " + name
	}
	body := fmt.Sprintf(
		"%s,\r\n\r\n"+
			"Tu código de verificación de VisionPrice es:\r\n\r\n"+
			"    %s\r\n\r\n"+
			"Este código expira en %d minuto(s). Si no solicitaste este código, ignora este correo.\r\n",
		greeting, code, expirationMinutes)

	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\n", fromHeader)
	fmt.Fprintf(&msg, "To: %s\r\n", email)
	fmt.Fprintf(&msg, "Subject: Tu código de verificación VisionPrice\r\n")
	fmt.Fprintf(&msg, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	fmt.Fprintf(&msg, "\r\n%s", body)

	if err := smtp.SendMail(addr, auth, from, []string{email}, []byte(msg.String())); err != nil {
		return fmt.Errorf("smtp: failed to send OTP email: %w", err)
	}
	return nil
}

// LogOTPNotifier is a development fallback that logs the OTP instead of emailing
// it. It is used when SMTP is not configured, preserving the previous stub
// behavior so local development keeps working.
type LogOTPNotifier struct{}

// NewLogOTPNotifier creates a notifier that only logs the code.
func NewLogOTPNotifier() *LogOTPNotifier { return &LogOTPNotifier{} }

// SendOTP logs the OTP code to stdout.
func (n *LogOTPNotifier) SendOTP(_ context.Context, email, _, code string, expirationMinutes int) error {
	log.Printf("🔐 [SMTP NO CONFIGURADO] OTP para %s: %s (expira en %d min)", email, code, expirationMinutes)
	return nil
}
