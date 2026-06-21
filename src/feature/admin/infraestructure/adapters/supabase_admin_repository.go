package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/visionprice/proveedores-backend/src/feature/admin/domain/entities"
)

// SupabaseAdminRepository implements AdminRepository using PostgreSQL (Supabase).
type SupabaseAdminRepository struct {
	db *pgxpool.Pool
}

// NewSupabaseAdminRepository creates a new SupabaseAdminRepository.
func NewSupabaseAdminRepository(db *pgxpool.Pool) *SupabaseAdminRepository {
	return &SupabaseAdminRepository{db: db}
}

// GetGlobalMetrics retrieves system-wide metrics from the database (HU_SYS_01).
func (r *SupabaseAdminRepository) GetGlobalMetrics(ctx context.Context) (*entities.GlobalMetrics, error) {
	metrics := &entities.GlobalMetrics{}

	// Count total active constructors
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM constructors WHERE active = TRUE`).Scan(&metrics.TotalConstructors)
	if err != nil {
		return nil, fmt.Errorf("failed to count constructors: %w", err)
	}

	// Count total active providers
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM providers WHERE active = TRUE`).Scan(&metrics.TotalProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to count providers: %w", err)
	}

	// Count active users in the last 24 hours (providers + constructors with recent last_login_at)
	activeUsersQuery := `
		SELECT (
			(SELECT COUNT(*) FROM providers WHERE last_login_at > NOW() - INTERVAL '24 hours' AND active = TRUE) +
			(SELECT COUNT(*) FROM constructors WHERE last_login_at > NOW() - INTERVAL '24 hours' AND active = TRUE)
		)
	`
	err = r.db.QueryRow(ctx, activeUsersQuery).Scan(&metrics.ActiveUsers24h)
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	return metrics, nil
}

// GetProviderMapPins retrieves all active provider locations from the secure view (HU_SYS_02).
// Uses v_provider_map_pins which only exposes safe data (no RFC, email, phone, password, prices).
func (r *SupabaseAdminRepository) GetProviderMapPins(ctx context.Context) ([]*entities.ProviderMapPin, error) {
	query := `
		SELECT id, business_name, city, state, created_at, latitude, longitude
		FROM v_provider_map_pins
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider map pins: %w", err)
	}
	defer rows.Close()

	var pins []*entities.ProviderMapPin
	for rows.Next() {
		pin := &entities.ProviderMapPin{}
		if err := rows.Scan(
			&pin.ID,
			&pin.BusinessName,
			&pin.City,
			&pin.State,
			&pin.CreatedAt,
			&pin.Latitude,
			&pin.Longitude,
		); err != nil {
			return nil, fmt.Errorf("failed to scan provider map pin: %w", err)
		}
		pins = append(pins, pin)
	}

	if pins == nil {
		pins = []*entities.ProviderMapPin{}
	}

	return pins, nil
}

// GetExpiringSubscriptions retrieves paid subscriptions expiring within withinDays
// (including overdue ones), joined with provider and plan data. Free plans have a
// NULL current_period_end and are naturally excluded.
func (r *SupabaseAdminRepository) GetExpiringSubscriptions(ctx context.Context, withinDays int) ([]*entities.ExpiringSubscription, error) {
	const query = `
		SELECT ps.provider_id, p.business_name, p.email,
		       ps.plan_code, sp.name, ps.status, ps.current_period_end
		FROM provider_subscriptions ps
		JOIN providers p ON p.id = ps.provider_id
		JOIN subscription_plans sp ON sp.code = ps.plan_code
		WHERE ps.current_period_end IS NOT NULL
		  AND ps.status <> 'canceled'
		  AND ps.current_period_end <= NOW() + make_interval(days => $1)
		ORDER BY ps.current_period_end ASC
	`

	rows, err := r.db.Query(ctx, query, withinDays)
	if err != nil {
		return nil, fmt.Errorf("failed to query expiring subscriptions: %w", err)
	}
	defer rows.Close()

	items := []*entities.ExpiringSubscription{}
	for rows.Next() {
		item := &entities.ExpiringSubscription{}
		if err := rows.Scan(
			&item.ProviderID,
			&item.BusinessName,
			&item.Email,
			&item.PlanCode,
			&item.PlanName,
			&item.Status,
			&item.CurrentPeriodEnd,
		); err != nil {
			return nil, fmt.Errorf("failed to scan expiring subscription: %w", err)
		}
		item.DaysUntilExpiry = int(time.Until(item.CurrentPeriodEnd).Hours() / 24)
		items = append(items, item)
	}

	return items, rows.Err()
}

// FindSystemUserByEmail retrieves a system admin user by email for authentication.
func (r *SupabaseAdminRepository) FindSystemUserByEmail(ctx context.Context, email string) (*entities.SystemUser, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM system_users
		WHERE email = $1
	`

	user := &entities.SystemUser{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("system user not found: %w", err)
	}

	return user, nil
}
