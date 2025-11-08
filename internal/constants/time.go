package constants

import "time"

// Time duration constants for common time windows and periods.
// These constants ensure consistency across the application and make
// time-based calculations more readable and maintainable.
const (
	// OneDayDuration is the duration of one day (24 hours)
	OneDayDuration = 24 * time.Hour

	// OneWeekDuration is the duration of one week (7 days)
	OneWeekDuration = 7 * 24 * time.Hour

	// ThirtyDayDuration is the duration of thirty days
	ThirtyDayDuration = 30 * 24 * time.Hour
)

// API timeout constants for external service calls.
const (
	// DefaultAPITimeout is the standard timeout for API calls (30 seconds)
	DefaultAPITimeout = 30 * time.Second

	// LongAPITimeout is the extended timeout for longer-running API calls (60 seconds)
	// Used for operations that may take longer, such as AI analysis
	LongAPITimeout = 60 * time.Second
)

// Lookback period constants for data synchronization and analysis.
const (
	// DefaultSyncLookback is the default lookback period for plugin sync operations
	// when no last sync time is available (24 hours)
	DefaultSyncLookback = 24 * time.Hour

	// DefaultMetricsWindow is the default time window for metrics calculations
	// when no specific period is requested (24 hours)
	DefaultMetricsWindow = 24 * time.Hour

	// RiskValidityPeriod is how long risk predictions remain valid before expiring (30 days)
	RiskValidityPeriod = 30 * 24 * time.Hour
)
