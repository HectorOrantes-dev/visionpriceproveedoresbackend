package entities

import (
	"time"

	"github.com/google/uuid"
)

// ExpiringSubscription is a provider account whose paid subscription is
// approaching (or past) its renewal date, shown in the admin dashboard.
// Free-plan accounts never expire and are not included.
type ExpiringSubscription struct {
	ProviderID       uuid.UUID `json:"provider_id"`
	BusinessName     string    `json:"business_name"`
	Email            string    `json:"email"`
	PlanCode         string    `json:"plan_code"`
	PlanName         string    `json:"plan_name"`
	Status           string    `json:"status"`
	CurrentPeriodEnd time.Time `json:"current_period_end"`
	// DaysUntilExpiry is positive for upcoming renewals and negative if overdue.
	DaysUntilExpiry int `json:"days_until_expiry"`
}
