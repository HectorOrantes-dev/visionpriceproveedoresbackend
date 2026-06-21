package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SupabasePaymentRepository implements the EventStore port for webhook idempotency.
type SupabasePaymentRepository struct {
	db *pgxpool.Pool
}

// NewSupabasePaymentRepository creates a new SupabasePaymentRepository.
func NewSupabasePaymentRepository(db *pgxpool.Pool) *SupabasePaymentRepository {
	return &SupabasePaymentRepository{db: db}
}

// RecordIfNew inserts the event; returns isNew=false if it was already recorded.
// Idempotency is enforced by the UNIQUE(provider, external_event_id) constraint.
func (r *SupabasePaymentRepository) RecordIfNew(ctx context.Context, provider, externalEventID, eventType string, payload []byte) (bool, error) {
	const q = `
		INSERT INTO payment_events (provider, external_event_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, external_event_id) DO NOTHING
	`
	tag, err := r.db.Exec(ctx, q, provider, externalEventID, eventType, payload)
	if err != nil {
		return false, fmt.Errorf("failed to record payment event: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// MarkProcessed flags a recorded event as fully processed.
func (r *SupabasePaymentRepository) MarkProcessed(ctx context.Context, provider, externalEventID string) error {
	const q = `UPDATE payment_events SET processed = TRUE WHERE provider = $1 AND external_event_id = $2`
	_, err := r.db.Exec(ctx, q, provider, externalEventID)
	return err
}

// GetProviderContact returns the provider's business name and email, used to
// create a customer at the payment gateway. Implements the ProviderDirectory port.
func (r *SupabasePaymentRepository) GetProviderContact(ctx context.Context, providerID string) (string, string, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return "", "", fmt.Errorf("invalid provider id: %w", err)
	}
	const q = `SELECT business_name, email FROM providers WHERE id = $1 AND active = TRUE`
	var name, email string
	if err := r.db.QueryRow(ctx, q, pid).Scan(&name, &email); err != nil {
		return "", "", fmt.Errorf("provider contact not found: %w", err)
	}
	return name, email, nil
}
