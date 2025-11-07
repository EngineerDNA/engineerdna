package models

import "time"

// AlertRule represents a rule that defines when to trigger alerts
type AlertRule struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	AlertType         string    `json:"alert_type"`
	Enabled           bool      `json:"enabled"`
	ThresholdValue    *float64  `json:"threshold_value,omitempty"`
	ThresholdOperator string    `json:"threshold_operator,omitempty"`
	Severity          string    `json:"severity"`
	TargetEntity      string    `json:"target_entity,omitempty"`
	TargetID          string    `json:"target_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AlertChannel represents a delivery channel for alert notifications
type AlertChannel struct {
	ID              string  `json:"id"`
	RuleID          string  `json:"rule_id"`
	ChannelType     string  `json:"channel_type"`
	ChannelConfig   string  `json:"channel_config,omitempty"`
	Enabled         bool    `json:"enabled"`
	QuietHoursStart *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty"`
}

// AlertInstance represents a fired alert
type AlertInstance struct {
	ID             string     `json:"id"`
	RuleID         string     `json:"rule_id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Severity       string     `json:"severity"`
	EntityType     string     `json:"entity_type,omitempty"`
	EntityID       string     `json:"entity_id,omitempty"`
	Context        string     `json:"context,omitempty"`
	FiredAt        time.Time  `json:"fired_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *string    `json:"acknowledged_by,omitempty"`
	SnoozedUntil   *time.Time `json:"snoozed_until,omitempty"`
	DismissedAt    *time.Time `json:"dismissed_at,omitempty"`
	DismissedBy    *string    `json:"dismissed_by,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// AlertDelivery represents a delivery attempt for an alert
type AlertDelivery struct {
	ID           string    `json:"id"`
	InstanceID   string    `json:"instance_id"`
	ChannelType  string    `json:"channel_type"`
	DeliveredAt  time.Time `json:"delivered_at"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}

// AlertContext represents additional context data for alerts (stored as JSON)
type AlertContext struct {
	PRID           string                 `json:"pr_id,omitempty"`
	PRTitle        string                 `json:"pr_title,omitempty"`
	PRAge          float64                `json:"pr_age,omitempty"`
	EngineerName   string                 `json:"engineer_name,omitempty"`
	TeamName       string                 `json:"team_name,omitempty"`
	ScoreCurrent   float64                `json:"score_current,omitempty"`
	ScorePrevious  float64                `json:"score_previous,omitempty"`
	ScoreDrop      float64                `json:"score_drop,omitempty"`
	SprintName     string                 `json:"sprint_name,omitempty"`
	SprintProgress float64                `json:"sprint_progress,omitempty"`
	Additional     map[string]interface{} `json:"additional,omitempty"`
}

// AlertChannelConfig represents configuration for delivery channels (stored as JSON)
type AlertChannelConfig struct {
	WebhookURL string `json:"webhook_url,omitempty"`
	Channel    string `json:"channel,omitempty"`
	Email      string `json:"email,omitempty"`
}
