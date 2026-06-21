package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/subscriptions/domain/entities"
)

// SupabaseSubscriptionRepository implements SubscriptionRepository using PostgreSQL.
type SupabaseSubscriptionRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseSubscriptionRepository creates a new SupabaseSubscriptionRepository.
func NewSupabaseSubscriptionRepository(db *pgxpool.Pool) *SupabaseSubscriptionRepository {
	return &SupabaseSubscriptionRepository{db: db}
}

// GetByProvider returns the provider's current subscription.
func (r *SupabaseSubscriptionRepository) GetByProvider(ctx context.Context, providerID uuid.UUID) (*entities.ProviderSubscription, error) {
	const q = `
		SELECT provider_id, plan_code, status, payment_provider,
		       external_customer_id, external_subscription_id, current_period_end
		FROM provider_subscriptions
		WHERE provider_id = $1
	`
	sub := &entities.ProviderSubscription{}
	err := r.db.QueryRow(ctx, q, providerID).Scan(
		&sub.ProviderID,
		&sub.PlanCode,
		&sub.Status,
		&sub.PaymentProvider,
		&sub.ExternalCustomerID,
		&sub.ExternalSubscriptionID,
		&sub.CurrentPeriodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	return sub, nil
}

// GetPlan returns a single plan by its code.
func (r *SupabaseSubscriptionRepository) GetPlan(ctx context.Context, code string) (*entities.Plan, error) {
	const q = `
		SELECT code, name, product_limit, price_cents, currency, features
		FROM subscription_plans
		WHERE code = $1
	`
	return scanPlan(r.db.QueryRow(ctx, q, code))
}

// ListPlans returns all active plans ordered by price.
func (r *SupabaseSubscriptionRepository) ListPlans(ctx context.Context) ([]*entities.Plan, error) {
	const q = `
		SELECT code, name, product_limit, price_cents, currency, features
		FROM subscription_plans
		WHERE active = TRUE
		ORDER BY price_cents ASC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	plans := []*entities.Plan{}
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

// EnsureDefault creates a free/active subscription if the provider has none.
func (r *SupabaseSubscriptionRepository) EnsureDefault(ctx context.Context, providerID uuid.UUID) error {
	const q = `
		INSERT INTO provider_subscriptions (provider_id, plan_code, status)
		VALUES ($1, 'free', 'active')
		ON CONFLICT (provider_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, q, providerID)
	return err
}

// CountActiveProducts returns how many active products the provider has.
func (r *SupabaseSubscriptionRepository) CountActiveProducts(ctx context.Context, providerID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM products WHERE provider_id = $1 AND active = TRUE`
	var count int
	err := r.db.QueryRow(ctx, q, providerID).Scan(&count)
	return count, err
}

// UpsertFromWebhook applies a normalized webhook update to the subscription.
func (r *SupabaseSubscriptionRepository) UpsertFromWebhook(ctx context.Context, in entities.WebhookUpdate) error {
	const q = `
		INSERT INTO provider_subscriptions
			(provider_id, plan_code, status, payment_provider,
			 external_customer_id, external_subscription_id, current_period_end, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (provider_id) DO UPDATE SET
			plan_code = EXCLUDED.plan_code,
			status = EXCLUDED.status,
			payment_provider = EXCLUDED.payment_provider,
			external_customer_id = EXCLUDED.external_customer_id,
			external_subscription_id = EXCLUDED.external_subscription_id,
			current_period_end = EXCLUDED.current_period_end,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, q,
		in.ProviderID,
		in.PlanCode,
		in.Status,
		nullIfEmpty(in.PaymentProvider),
		nullIfEmpty(in.ExternalCustomerID),
		nullIfEmpty(in.ExternalSubscriptionID),
		in.CurrentPeriodEnd,
	)
	return err
}

// scanPlan scans one plan row, decoding the features JSONB column.
func scanPlan(row pgx.Row) (*entities.Plan, error) {
	plan := &entities.Plan{}
	var featuresRaw []byte
	if err := row.Scan(
		&plan.Code,
		&plan.Name,
		&plan.ProductLimit,
		&plan.PriceCents,
		&plan.Currency,
		&featuresRaw,
	); err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if len(featuresRaw) > 0 {
		if err := json.Unmarshal(featuresRaw, &plan.Features); err != nil {
			return nil, fmt.Errorf("failed to decode plan features: %w", err)
		}
	}
	return plan, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
