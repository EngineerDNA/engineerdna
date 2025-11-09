package models

import "time"

// WeeklyBriefing represents an AI-generated weekly summary and insights
type WeeklyBriefing struct {
	ID             string          `json:"id"`
	TeamID         string          `json:"team_id"`
	WeekStart      time.Time       `json:"week_start"`
	TLDR           string          `json:"tldr"`
	KeyMetrics     []MetricSummary `json:"key_metrics"`
	NeedsAttention []AttentionItem `json:"needs_attention"`
	Insights       []AIInsight     `json:"insights"`
	TrendingUp     []string        `json:"trending_up"`
	TrendingDown   []string        `json:"trending_down"`
	TalkingPoints  string          `json:"talking_points"`
	GeneratedAt    time.Time       `json:"generated_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

// AttentionItem represents an engineer or issue that needs manager attention
type AttentionItem struct {
	EngineerID     string   `json:"engineer_id"`
	EngineerName   string   `json:"engineer_name"`
	Severity       string   `json:"severity"`
	Issue          string   `json:"issue"`
	Evidence       []string `json:"evidence"`
	SuggestedTopic string   `json:"suggested_topic"`
	DataLinks      []string `json:"data_links,omitempty"`
}

// AIInsight represents an observation from the AI with context and recommendation
type AIInsight struct {
	Observation    string `json:"observation"`
	Context        string `json:"context"`
	Recommendation string `json:"recommendation,omitempty"`
}

// MetricSummary represents a key metric with trend information
type MetricSummary struct {
	Name          string  `json:"name"`
	Value         string  `json:"value"`
	ChangePercent float64 `json:"change_percent"`
	Direction     string  `json:"direction"` // "up", "down", "stable"
	Context       string  `json:"context,omitempty"`
}

// BriefingPrompt represents the prompt data sent to the AI
type BriefingPrompt struct {
	TeamName       string                 `json:"team_name"`
	ManagerName    string                 `json:"manager_name"`
	TeamSize       int                    `json:"team_size"`
	WeekStart      time.Time              `json:"week_start"`
	WeekEnd        time.Time              `json:"week_end"`
	Events         []Event                `json:"events"`
	TeamMetrics    map[string]interface{} `json:"team_metrics"`
	UserPriorities map[string]float64     `json:"user_priorities"`
}

// BriefingResponse represents the AI's response structure
type BriefingResponse struct {
	TLDR              string                  `json:"tldr"`
	NeedsAttention    []BriefingAttentionItem `json:"needs_attention"`
	WhyMetricsChanged []MetricChange          `json:"why_metrics_changed"`
	Insights          []BriefingInsight       `json:"insights"`
	TrendingUp        []string                `json:"trending_up"`
	TrendingDown      []string                `json:"trending_down"`
	TalkingPoints     string                  `json:"talking_points"`
}

// BriefingAttentionItem represents attention items from AI response
type BriefingAttentionItem struct {
	Person         string   `json:"person"`
	Severity       string   `json:"severity"`
	Issue          string   `json:"issue"`
	Evidence       []string `json:"evidence"`
	SuggestedTopic string   `json:"suggested_topic"`
}

// MetricChange represents why a metric changed
type MetricChange struct {
	Metric      string `json:"metric"`
	Direction   string `json:"direction"`
	Explanation string `json:"explanation"`
	DataLink    string `json:"data_link,omitempty"`
}

// BriefingInsight represents an insight from AI response
type BriefingInsight struct {
	Observation    string `json:"observation"`
	Context        string `json:"context"`
	Recommendation string `json:"recommendation,omitempty"`
}
