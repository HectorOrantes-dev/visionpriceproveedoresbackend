package errors

import "errors"

// Domain error types for consistent error handling across features.
var (
	ErrNotFound        = errors.New("resource not found")
	ErrConflict        = errors.New("resource already exists")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrValidation      = errors.New("validation error")
	ErrInternal        = errors.New("internal server error")
	ErrExpired         = errors.New("resource has expired")
	ErrInvalidInput    = errors.New("invalid input provided")
	ErrPaymentRequired = errors.New("payment required") // plan limit reached / upgrade needed (HTTP 402)
	ErrNotImplemented  = errors.New("not implemented")  // feature scaffolded but not wired (HTTP 501)
)

// DomainError wraps a base error with a human-readable message.
type DomainError struct {
	Base    error
	Message string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	return e.Message
}

// Unwrap allows errors.Is/errors.As to work with the base error.
func (e *DomainError) Unwrap() error {
	return e.Base
}

// NewDomainError creates a new DomainError.
func NewDomainError(base error, message string) *DomainError {
	return &DomainError{
		Base:    base,
		Message: message,
	}
}

// Is checks if a DomainError wraps a specific base error.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
