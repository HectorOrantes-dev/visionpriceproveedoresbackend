package subscription_usecase

import (
	"context"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain"
	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain/entities"
)

// defaultPlanCode is the plan assigned to every new provider; it never expires.
const defaultPlanCode = "free"

// SubscriptionUseCase contains business logic for subscriptions and plan limits.
type SubscriptionUseCase struct {
	repo domain.SubscriptionRepository
}

// NewSubscriptionUseCase creates a new SubscriptionUseCase.
func NewSubscriptionUseCase(repo domain.SubscriptionRepository) *SubscriptionUseCase {
	return &SubscriptionUseCase{repo: repo}
}

// GetCurrent returns the provider's plan, status, and product usage.
func (uc *SubscriptionUseCase) GetCurrent(ctx context.Context, providerID string) (*entities.SubscriptionView, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	sub, err := uc.repo.GetByProvider(ctx, pid)
	if err != nil {
		// Self-heal: a provider should always have a subscription. Create the
		// default one and continue with free defaults.
		if ensureErr := uc.repo.EnsureDefault(ctx, pid); ensureErr != nil {
			return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener la suscripción")
		}
		sub, err = uc.repo.GetByProvider(ctx, pid)
		if err != nil {
			return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener la suscripción")
		}
	}

	plan, err := uc.repo.GetPlan(ctx, sub.PlanCode)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener el plan")
	}

	count, err := uc.repo.CountActiveProducts(ctx, pid)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al contar productos")
	}

	return &entities.SubscriptionView{
		Plan:             *plan,
		Status:           sub.Status,
		ProductCount:     count,
		ProductLimit:     plan.ProductLimit,
		Unlimited:        plan.ProductLimit == nil,
		CurrentPeriodEnd: sub.CurrentPeriodEnd,
	}, nil
}

// ListPlans returns the catalog of available plans.
func (uc *SubscriptionUseCase) ListPlans(ctx context.Context) ([]*entities.Plan, error) {
	plans, err := uc.repo.ListPlans(ctx)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener los planes")
	}
	return plans, nil
}

// ProductLimit returns the product cap for the provider's current plan.
// It implements the PlanLimitService port consumed by products and extracciones.
// unlimited=true means there is no cap (Plan Max).
func (uc *SubscriptionUseCase) ProductLimit(ctx context.Context, providerID string) (limit int, unlimited bool, err error) {
	pid, parseErr := uuid.Parse(providerID)
	if parseErr != nil {
		return 0, false, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	sub, err := uc.repo.GetByProvider(ctx, pid)
	if err != nil {
		// No subscription yet: fall back to the free plan limit (fail closed).
		return uc.freePlanLimit(ctx)
	}

	// Only an active subscription grants its plan's limit. A canceled or past_due
	// account reverts to the free tier's cap until it pays again.
	planCode := sub.PlanCode
	if sub.Status != "active" {
		return uc.freePlanLimit(ctx)
	}

	plan, err := uc.repo.GetPlan(ctx, planCode)
	if err != nil {
		return 0, false, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener el plan")
	}
	if plan.ProductLimit == nil {
		return 0, true, nil
	}
	return *plan.ProductLimit, false, nil
}

// freePlanLimit returns the free plan's product cap (fail-closed default).
func (uc *SubscriptionUseCase) freePlanLimit(ctx context.Context) (int, bool, error) {
	freePlan, err := uc.repo.GetPlan(ctx, defaultPlanCode)
	if err != nil || freePlan.ProductLimit == nil {
		return 0, false, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al determinar el límite del plan")
	}
	return *freePlan.ProductLimit, false, nil
}

// EnsureDefault creates the default free subscription for a provider if missing.
// It implements the DefaultSubscriptionCreator port consumed by the register feature.
func (uc *SubscriptionUseCase) EnsureDefault(ctx context.Context, providerID uuid.UUID) error {
	return uc.repo.EnsureDefault(ctx, providerID)
}

// ApplyWebhook applies a normalized payment webhook update to a subscription.
// It implements the SubscriptionUpdater port consumed by the payments feature.
func (uc *SubscriptionUseCase) ApplyWebhook(ctx context.Context, update entities.WebhookUpdate) error {
	return uc.repo.UpsertFromWebhook(ctx, update)
}

// ResolveProviderByExternalSubscription maps a gateway subscription id to a
// provider id. It implements part of the SubscriptionUpdater port (used for
// Conekta webhooks, which don't carry our provider id).
func (uc *SubscriptionUseCase) ResolveProviderByExternalSubscription(ctx context.Context, externalSubscriptionID string) (string, error) {
	pid, err := uc.repo.FindProviderByExternalSubscription(ctx, externalSubscriptionID)
	if err != nil {
		return "", err
	}
	return pid.String(), nil
}
