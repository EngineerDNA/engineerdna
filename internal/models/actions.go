package models

import "time"

// Recommendation represents an AI-generated or manual recommendation for a manager
type Recommendation struct {
	ID                 string     `json:"id"`
	SourceType         string     `json:"source_type"`         // ai_insight, alert, briefing, manual
	SourceID           *string    `json:"source_id,omitempty"` // alert_id, insight_id, etc.
	RecommendationType string     `json:"recommendation_type"` // check_in, adjust_workload, etc.
	Priority           string     `json:"priority"`            // low, medium, high, urgent
	SubjectType        string     `json:"subject_type"`        // engineer, team, process
	SubjectID          *string    `json:"subject_id,omitempty"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	SuggestedActions   []string   `json:"suggested_actions,omitempty"` // stored as JSON
	Context            string     `json:"context,omitempty"`           // JSON: why this recommendation was made
	AssignedTo         *string    `json:"assigned_to,omitempty"`       // manager_id
	CreatedAt          time.Time  `json:"created_at"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	Status             string     `json:"status"` // pending, in_progress, completed, dismissed, snoozed
}

// Action represents an actual action taken by a manager
type Action struct {
	ID               string     `json:"id"`
	RecommendationID *string    `json:"recommendation_id,omitempty"` // can be null for ad-hoc actions
	ActionType       string     `json:"action_type"`                 // conversation, process_change, etc.
	SubjectType      string     `json:"subject_type"`
	SubjectID        *string    `json:"subject_id,omitempty"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	TakenBy          string     `json:"taken_by"` // manager_id
	TakenAt          time.Time  `json:"taken_at"`
	Evidence         string     `json:"evidence,omitempty"` // JSON: notes, meeting ID, etc.
	ExpectedOutcome  string     `json:"expected_outcome,omitempty"`
	FollowUpDate     *time.Time `json:"follow_up_date,omitempty"`
}

// ActionOutcome represents the measured effectiveness of an action
type ActionOutcome struct {
	ID               string    `json:"id"`
	ActionID         string    `json:"action_id"`
	OutcomeType      string    `json:"outcome_type"`    // metric_improvement, issue_resolved, etc.
	MeasuredMetric   string    `json:"measured_metric"` // score, velocity, sentiment
	BeforeValue      *float64  `json:"before_value,omitempty"`
	AfterValue       *float64  `json:"after_value,omitempty"`
	ChangePercentage *float64  `json:"change_percentage,omitempty"`
	TimeToImpactDays *int      `json:"time_to_impact_days,omitempty"`
	Effectiveness    string    `json:"effectiveness"` // highly_effective, somewhat_effective, not_effective
	Notes            string    `json:"notes,omitempty"`
	MeasuredAt       time.Time `json:"measured_at"`
}

// RecommendationHistory tracks the lifecycle of a recommendation
type RecommendationHistory struct {
	ID               string    `json:"id"`
	RecommendationID string    `json:"recommendation_id"`
	StatusChange     string    `json:"status_change"` // created, assigned, in_progress, etc.
	ChangedBy        *string   `json:"changed_by,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	ChangedAt        time.Time `json:"changed_at"`
}

// FollowUp represents a scheduled follow-up reminder for an action
type FollowUp struct {
	ID           string     `json:"id"`
	ActionID     string     `json:"action_id"`
	FollowUpDate time.Time  `json:"follow_up_date"`
	FollowUpType string     `json:"follow_up_type"` // check_metric, check_in, verify_resolution
	Description  string     `json:"description,omitempty"`
	Completed    bool       `json:"completed"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// ActionEffectiveness is a computed summary of action effectiveness
type ActionEffectiveness struct {
	ActionID               string    `json:"action_id"`
	ActionType             string    `json:"action_type"`
	RecommendationType     string    `json:"recommendation_type,omitempty"`
	SubjectType            string    `json:"subject_type"`
	TakenAt                time.Time `json:"taken_at"`
	OutcomeCount           int       `json:"outcome_count"`
	HighlyEffectiveCount   int       `json:"highly_effective_count"`
	SomewhatEffectiveCount int       `json:"somewhat_effective_count"`
	NotEffectiveCount      int       `json:"not_effective_count"`
	AverageTimeToImpact    *float64  `json:"average_time_to_impact,omitempty"`
	AverageChangePercent   *float64  `json:"average_change_percent,omitempty"`
}

// RecommendationWithDetails includes recommendation with related entities
type RecommendationWithDetails struct {
	Recommendation *Recommendation          `json:"recommendation"`
	Actions        []*Action                `json:"actions,omitempty"`
	History        []*RecommendationHistory `json:"history,omitempty"`
}

// ActionWithDetails includes action with outcomes and follow-ups
type ActionWithDetails struct {
	Action    *Action          `json:"action"`
	Outcomes  []*ActionOutcome `json:"outcomes,omitempty"`
	FollowUps []*FollowUp      `json:"follow_ups,omitempty"`
}

// EffectivenessPattern represents a learned pattern of what works
type EffectivenessPattern struct {
	ActionType         string  `json:"action_type"`
	RecommendationType string  `json:"recommendation_type"`
	SuccessRate        float64 `json:"success_rate"` // % of actions that were highly effective
	SampleSize         int     `json:"sample_size"`
	AvgTimeToImpact    float64 `json:"avg_time_to_impact_days"`
	Recommendation     string  `json:"recommendation"` // what to do based on this pattern
}

// EffectivenessReport is a summary report for managers
type EffectivenessReport struct {
	TimePeriod             string                  `json:"time_period"`
	TotalActions           int                     `json:"total_actions"`
	ActionsWithOutcomes    int                     `json:"actions_with_outcomes"`
	HighlyEffectiveCount   int                     `json:"highly_effective_count"`
	SomewhatEffectiveCount int                     `json:"somewhat_effective_count"`
	NotEffectiveCount      int                     `json:"not_effective_count"`
	SuccessRate            float64                 `json:"success_rate"` // % highly effective
	SuccessfulPatterns     []*EffectivenessPattern `json:"successful_patterns"`
	UnsuccessfulPatterns   []*EffectivenessPattern `json:"unsuccessful_patterns"`
	TopActions             []*Action               `json:"top_actions"` // most effective recent actions
}
