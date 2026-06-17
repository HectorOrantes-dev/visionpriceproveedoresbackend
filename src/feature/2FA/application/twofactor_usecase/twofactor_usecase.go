package twofactor_usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/middleware"
	"github.com/visionprice/proveedores-backend/src/feature/2FA/domain"
)

// TwoFactorUseCase contains business logic for 2FA OTP.
type TwoFactorUseCase struct {
	repo                        domain.TwoFactorRepository
	jwtSecret                   string
	otpExpirationMinutes        int
	jwtExpirationMinutes        int
	refreshTokenExpirationHours int
}

// NewTwoFactorUseCase creates a new TwoFactorUseCase.
func NewTwoFactorUseCase(
	repo domain.TwoFactorRepository,
	jwtSecret string,
	otpExpirationMinutes int,
	jwtExpirationMinutes int,
	refreshTokenExpirationHours int,
) *TwoFactorUseCase {
	return &TwoFactorUseCase{
		repo:                        repo,
		jwtSecret:                   jwtSecret,
		otpExpirationMinutes:        otpExpirationMinutes,
		jwtExpirationMinutes:        jwtExpirationMinutes,
		refreshTokenExpirationHours: refreshTokenExpirationHours,
	}
}

// GenerateOTP creates a 6-digit OTP, stores it, and logs it (stub email).
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

	// Stub: log the OTP code instead of sending via email/SMS
	log.Printf("🔐 OTP CODE for provider %s: %s (expires in %d min)", providerID, code, uc.otpExpirationMinutes)

	return nil
}

// VerifyOTP validates the OTP code and returns a full JWT token pair.
func (uc *TwoFactorUseCase) VerifyOTP(ctx context.Context, providerID string, code string) (string, string, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return "", "", domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	valid, err := uc.repo.ValidateOTP(ctx, pid, code)
	if err != nil || !valid {
		return "", "", domainErrors.NewDomainError(domainErrors.ErrUnauthorized, "Código OTP inválido o expirado")
	}

	// Generate full access token
	accessToken, err := middleware.GenerateToken(
		providerID,
		middleware.TokenTypeAccess,
		uc.jwtSecret,
		time.Duration(uc.jwtExpirationMinutes)*time.Minute,
	)
	if err != nil {
		return "", "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar token de acceso")
	}

	// Generate refresh token
	refreshToken, err := middleware.GenerateToken(
		providerID,
		middleware.TokenTypeRefresh,
		uc.jwtSecret,
		time.Duration(uc.refreshTokenExpirationHours)*time.Hour,
	)
	if err != nil {
		return "", "", domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al generar refresh token")
	}

	return accessToken, refreshToken, nil
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
