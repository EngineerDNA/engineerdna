package models

import "time"

// Engineer represents a canonical engineer identity that aggregates identifiers
// from multiple sources (GitHub, Jira, Zoom, CSV, etc.)
type Engineer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email,omitempty"`
	Manager     string            `json:"manager,omitempty"`
	Identifiers map[string]string `json:"identifiers"`
	Active      bool              `json:"active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// UnresolvedIdentity represents an identity (source + identifier) that hasn't
// been mapped to a canonical Engineer record yet
type UnresolvedIdentity struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	Identifier string    `json:"identifier"`
	FirstSeen  time.Time `json:"first_seen"`
	EventCount int       `json:"event_count"`
	Ignored    bool      `json:"ignored"`
}

// MatchSuggestion represents a suggested mapping between an unresolved identity
// and an engineer, with confidence scoring
type MatchSuggestion struct {
	UnresolvedID string  `json:"unresolved_id"`
	EngineerID   string  `json:"engineer_id"`
	EngineerName string  `json:"engineer_name"`
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
}

// ActivityMetrics represents aggregate activity metrics for an engineer
type ActivityMetrics struct {
	EngineerID    string    `json:"engineer_id"`
	EngineerName  string    `json:"engineer_name"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	PullRequests  int       `json:"pull_requests"`
	Issues        int       `json:"issues"`
	Commits       int       `json:"commits"`
	CodeReviews   int       `json:"code_reviews"`
	TotalEvents   int       `json:"total_events"`
	UniqueSources []string  `json:"unique_sources"`
}
