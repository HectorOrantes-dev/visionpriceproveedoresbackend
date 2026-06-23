package domain

import "context"

// OTPNotifier is the port for delivering a one-time password to a provider
// through some channel (email, SMS, ...). Adapters implement the actual transport.
type OTPNotifier interface {
	// SendOTP delivers the code to the given recipient. name is the provider's
	// display name (may be empty); expirationMinutes is informational for the body.
	SendOTP(ctx context.Context, email, name, code string, expirationMinutes int) error
}
