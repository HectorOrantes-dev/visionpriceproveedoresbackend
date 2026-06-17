package entities

// GlobalMetrics represents the system-wide metrics for the admin dashboard (HU_SYS_01).
type GlobalMetrics struct {
	TotalConstructors int `json:"total_constructors"`
	TotalProviders    int `json:"total_providers"`
	ActiveUsers24h    int `json:"active_users_24h"`
}
