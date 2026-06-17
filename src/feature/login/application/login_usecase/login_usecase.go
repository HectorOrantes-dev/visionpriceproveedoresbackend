package login_usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/login/domain"
)

// LoginUseCase contains business logic for authentication.
type LoginUseCase struct {
	repo                           domain.LoginRepository
	jwtSecret                      string
	jwtExpirationMinutes           int
	otpExpirationMinutes           int
	passwordResetExpirationMinutes int
}

// NewLoginUseCase creates a new LoginUseCase.
func NewLoginUseCase(
	repo domain.LoginRepository,
	jwtSecret string,
	jwtExpirationMinutes int,
	otpExpirationMinutes int,
	passwordResetExpirationMinutes int,
) *LoginUseCase {
	return &LoginUseCase{
		repo:                           repo,
		jwtSecret:                      jwtSecret,
		jwtExpirationMinutes:           jwtExpirationMinutes,
		otpExpirationMinutes:           otpExpirationMinutes,
		passwordResetExpirationMinutes: passwordResetExpirationMinutes,
	}
}

// Login validates credentials and returns a temporary JWT for the 2FA flow.
func (uc *LoginUseCase) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	provider, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(provider.PasswordHash), []byte(password)); err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Credenciales inválidas")
	}

	// Generate temporary JWT for 2FA
	tempToken, err := middleware.GenerateToken(
		provider.ID.String(),
		middleware.TokenTypeOTPTemp,
		uc.jwtSecret,
		time.Duration(uc.otpExpirationMinutes)*time.Minute,
	)
	if err != nil {
		return "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar token temporal")
	}

	return tempToken, nil
}

// ForgotPassword generates a password reset token and logs it (stub email).
func (uc *LoginUseCase) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	provider, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		// Return nil to not leak whether the email exists
		log.Printf("INFO: Password reset requested for non-existent email: %s", email)
		return nil
	}

	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar token de recuperación")
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash the token for storage
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	if err := uc.repo.CreateResetToken(ctx, provider.ID, tokenHash, uc.passwordResetExpirationMinutes); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al crear token de recuperación")
	}

	// Stub: log the reset URL instead of sending an email
	resetURL := fmt.Sprintf("https://visionprice.app/reset-password?token=%s", rawToken)
	log.Printf("📧 PASSWORD RESET (stub email) for %s: %s", email, resetURL)

	return nil
}

// ResetPassword validates the reset token and updates the password.
func (uc *LoginUseCase) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	// Hash the provided token to compare with stored hash
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	providerID, err := uc.repo.FindValidResetToken(ctx, tokenHash)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Token de recuperación inválido o expirado")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al procesar la nueva contraseña")
	}

	if err := uc.repo.UpdatePassword(ctx, providerID, string(hashedPassword)); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al actualizar la contraseña")
	}

	// Mark token as used
	if err := uc.repo.MarkResetTokenUsed(ctx, tokenHash); err != nil {
		log.Printf("WARNING: could not mark reset token as used: %v", err)
	}

	return nil
}
