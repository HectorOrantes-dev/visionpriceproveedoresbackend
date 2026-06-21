package register_usecase

import (
	"context"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/register/domain"
	"github.com/visionprice/proveedores-backend/src/feature/register/domain/entities"
)

// RegisterUseCase contains the business logic for provider registration.
type RegisterUseCase struct {
	repo          domain.RegisterRepository
	subscriptions domain.DefaultSubscriptionCreator
}

// NewRegisterUseCase creates a new RegisterUseCase with the given repository.
func NewRegisterUseCase(repo domain.RegisterRepository, subscriptions domain.DefaultSubscriptionCreator) *RegisterUseCase {
	return &RegisterUseCase{repo: repo, subscriptions: subscriptions}
}

// Execute registers a new provider after validating uniqueness and hashing the password.
func (uc *RegisterUseCase) Execute(ctx context.Context, req *entities.RegisterRequest) (*entities.RegisterResponse, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))
	rfc := strings.ToUpper(strings.TrimSpace(req.RFC))

	// Check email uniqueness
	emailExists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar el correo electrónico")
	}
	if emailExists {
		return nil, domainErrors.NewDomainError(domainErrors.ErrConflict, "El correo electrónico ya está registrado")
	}

	// Check RFC uniqueness
	rfcExists, err := uc.repo.ExistsByRFC(ctx, rfc)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al verificar el RFC")
	}
	if rfcExists {
		return nil, domainErrors.NewDomainError(domainErrors.ErrConflict, "El RFC ya está registrado")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al procesar la contraseña")
	}

	provider := &entities.Provider{
		BusinessName: strings.TrimSpace(req.BusinessName),
		RFC:          rfc,
		Email:        email,
		Phone:        strings.TrimSpace(req.Phone),
		PasswordHash: string(hashedPassword),
	}

	created, err := uc.repo.CreateProvider(ctx, provider)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al registrar el proveedor")
	}

	// Assign the default free subscription (Plan Gratuito, never expires).
	// A failure here is non-fatal: GetCurrent/ProductLimit self-heal the row later.
	if err := uc.subscriptions.EnsureDefault(ctx, created.ID); err != nil {
		log.Printf("WARNING: could not create default subscription for provider %s: %v", created.ID, err)
	}

	return &entities.RegisterResponse{
		ID:           created.ID,
		BusinessName: created.BusinessName,
		RFC:          entities.MaskRFC(created.RFC),
		Email:        created.Email,
		Phone:        created.Phone,
		CreatedAt:    created.CreatedAt,
	}, nil
}
