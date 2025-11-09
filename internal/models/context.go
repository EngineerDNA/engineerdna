package models

import "time"

// ManagerNote represents a private note from a manager about an engineer, team, event, or metric
type ManagerNote struct {
	ID           string    `json:"id"`
	ManagerID    string    `json:"manager_id"`
	SubjectType  string    `json:"subject_type"` // 'engineer', 'team', 'event', 'metric'
	SubjectID    string    `json:"subject_id"`   // engineer_id, team_id, event_id, etc.
	NoteType     string    `json:"note_type"`    // '1on1', 'performance', 'incident', 'context', 'feedback'
	Title        string    `json:"title,omitempty"`
	Content      string    `json:"content"`
	Visibility   string    `json:"visibility"` // 'private', 'shared_with_subject', 'team', 'org'
	Tags         []string  `json:"tags,omitempty"`
	Mood         string    `json:"mood,omitempty"` // 'positive', 'neutral', 'negative', 'concerned'
	ActionItems  []string  `json:"action_items,omitempty"`
	LinkedEvents []string  `json:"linked_events,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ContextAnnotation represents context linked to a specific metric, event, or alert
type ContextAnnotation struct {
	ID             string    `json:"id"`
	EntityType     string    `json:"entity_type"` // 'metric_drop', 'score_change', 'event', 'alert'
	EntityID       string    `json:"entity_id"`
	AnnotationType string    `json:"annotation_type"` // 'explanation', 'mitigation', 'expectation'
	Content        string    `json:"content"`
	AuthorID       string    `json:"author_id"`
	Visibility     string    `json:"visibility"` // 'private', 'team', 'org'
	CreatedAt      time.Time `json:"created_at"`
}

// TeamContext represents broader team-level context for a time period
type TeamContext struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	ContextType string    `json:"context_type"` // 'velocity_explanation', 'capacity_change', 'process_change'
	TimePeriod  string    `json:"time_period"`  // 'Q4 2024', '2024-W45', etc.
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact,omitempty"` // 'positive', 'negative', 'neutral'
	AuthorID    string    `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SurveyQuestion represents a single question in a sentiment survey
type SurveyQuestion struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"` // 'scale', 'text', 'choice'
	Question string              `json:"question"`
	Scale    []SurveyScaleOption `json:"scale,omitempty"`
	Choices  []string            `json:"choices,omitempty"`
}

// SurveyScaleOption represents a scale option for a survey question
type SurveyScaleOption struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// SentimentSurvey represents a pulse survey for morale/sentiment tracking
type SentimentSurvey struct {
	ID             string           `json:"id"`
	SurveyType     string           `json:"survey_type"` // 'weekly_pulse', 'quarterly', 'ad_hoc'
	Title          string           `json:"title"`
	Description    string           `json:"description,omitempty"`
	Questions      []SurveyQuestion `json:"questions"`
	TargetAudience string           `json:"target_audience"` // 'all', 'team:team_id', 'engineer:engineer_id'
	Anonymous      bool             `json:"anonymous"`
	Active         bool             `json:"active"`
	CreatedBy      string           `json:"created_by"`
	CreatedAt      time.Time        `json:"created_at"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
}

// SurveyResponse represents a response to a sentiment survey
type SurveyResponse struct {
	ID           string                 `json:"id"`
	SurveyID     string                 `json:"survey_id"`
	RespondentID *string                `json:"respondent_id,omitempty"` // null if anonymous
	Responses    map[string]interface{} `json:"responses"`               // question_id -> response
	MoodRating   *int                   `json:"mood_rating,omitempty"`   // 1-5 scale
	TextFeedback string                 `json:"text_feedback,omitempty"`
	SubmittedAt  time.Time              `json:"submitted_at"`
}

// SentimentAnalysis represents computed sentiment from various sources
type SentimentAnalysis struct {
	ID             string                 `json:"id"`
	EntityType     string                 `json:"entity_type"` // 'engineer', 'team', 'org'
	EntityID       string                 `json:"entity_id,omitempty"`
	TimePeriod     string                 `json:"time_period"`
	Source         string                 `json:"source"`          // 'survey', 'code_review_tone', 'commit_messages'
	SentimentScore float64                `json:"sentiment_score"` // -1.0 to 1.0
	Confidence     float64                `json:"confidence"`      // 0.0 to 1.0
	SampleSize     int                    `json:"sample_size,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"` // breakdown by dimension
	ComputedAt     time.Time              `json:"computed_at"`
}

// SentimentSummary represents aggregated sentiment for an entity
type SentimentSummary struct {
	EntityType      string             `json:"entity_type"`
	EntityID        string             `json:"entity_id,omitempty"`
	TimePeriod      string             `json:"time_period"`
	OverallScore    float64            `json:"overall_score"`
	Confidence      float64            `json:"confidence"`
	SourceBreakdown map[string]float64 `json:"source_breakdown"` // source -> score
	Trend           string             `json:"trend"`            // 'improving', 'declining', 'stable'
	ComputedAt      time.Time          `json:"computed_at"`
}

// EngineerContext represents complete context profile for an engineer
type EngineerContext struct {
	EngineerID   string              `json:"engineer_id"`
	EngineerName string              `json:"engineer_name"`
	Notes        []ManagerNote       `json:"notes"`
	Annotations  []ContextAnnotation `json:"annotations"`
	Sentiment    *SentimentSummary   `json:"sentiment,omitempty"`
	ActionItems  []string            `json:"action_items"`
}
