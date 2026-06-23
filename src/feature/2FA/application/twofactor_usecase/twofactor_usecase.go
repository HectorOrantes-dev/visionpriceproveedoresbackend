package twofactor_usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/visionprice/proveedores-backend/src/core/csrf"
	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/domain"
)

// TwoFactorUseCase contains business logic for 2FA OTP.
type TwoFactorUseCase struct {
	repo                        domain.TwoFactorRepository
	notifier                    domain.OTPNotifier
	csrfManager                 *csrf.Manager
	jwtSecret                   string
	otpExpirationMinutes        int
	jwtExpirationMinutes        int
	refreshTokenExpirationHours int
}

// NewTwoFactorUseCase creates a new TwoFactorUseCase.
func NewTwoFactorUseCase(
	repo domain.TwoFactorRepository,
	notifier domain.OTPNotifier,
	csrfManager *csrf.Manager,
	jwtSecret string,
	otpExpirationMinutes int,
	jwtExpirationMinutes int,
	refreshTokenExpirationHours int,
) *TwoFactorUseCase {
	return &TwoFactorUseCase{
		repo:                        repo,
		notifier:                    notifier,
		csrfManager:                 csrfManager,
		jwtSecret:                   jwtSecret,
		otpExpirationMinutes:        otpExpirationMinutes,
		jwtExpirationMinutes:        jwtExpirationMinutes,
		refreshTokenExpirationHours: refreshTokenExpirationHours,
	}
}

// GenerateOTP creates a 6-digit OTP, stores it, and delivers it to the provider
// via the configured notifier (email).
func (uc *TwoFactorUseCase) GenerateOTP(ctx context.Context, providerID string) error {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	code, err := generateSecureOTP(6)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar código OTP")
	}

	if err := uc.repo.CreateOTP(ctx, pid, code, uc.otpExpirationMinutes); err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al almacenar código OTP")
	}

	email, name, err := uc.repo.GetProviderContact(ctx, pid)
	if err != nil {
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener el correo del proveedor")
	}

	if err := uc.notifier.SendOTP(ctx, email, name, code, uc.otpExpirationMinutes); err != nil {
		// Log the underlying transport error (network timeout, auth rejected, API
		// error, etc.) so the real cause is visible in runtime logs; client gets 500.
		log.Printf("ERROR: failed to send OTP email to provider %s: %v", providerID, err)
		return domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al enviar el código OTP")
	}

	return nil
}

// VerifyOTP validates the OTP code and returns a full JWT token pair plus a
// per-session CSRF token bound to the provider.
func (uc *TwoFactorUseCase) VerifyOTP(ctx context.Context, providerID string, code string) (accessToken, refreshToken, csrfToken string, err error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return "", "", "", domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	valid, err := uc.repo.ValidateOTP(ctx, pid, code)
	if err != nil || !valid {
		return "", "", "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Código OTP inválido o expirado")
	}

	// Generate full access token
	accessToken, err = middleware.GenerateToken(
		providerID,
		middleware.TokenTypeAccess,
		uc.jwtSecret,
		time.Duration(uc.jwtExpirationMinutes)*time.Minute,
	)
	if err != nil {
		return "", "", "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar token de acceso")
	}

	// Generate refresh token
	refreshToken, err = middleware.GenerateToken(
		providerID,
		middleware.TokenTypeRefresh,
		uc.jwtSecret,
		time.Duration(uc.refreshTokenExpirationHours)*time.Hour,
	)
	if err != nil {
		return "", "", "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar refresh token")
	}

	// Issue a per-session CSRF token bound to this provider.
	csrfToken, err = uc.csrfManager.Issue(ctx, providerID)
	if err != nil {
		return "", "", "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar token CSRF")
	}

	return accessToken, refreshToken, csrfToken, nil
}

// generateSecureOTP generates a cryptographically secure N-digit numeric code.
func generateSecureOTP(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
