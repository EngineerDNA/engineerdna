package alerts

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// CreateDefaultAlertRules creates sensible default alert rules
func CreateDefaultAlertRules(alertsStore *db.AlertsStore) error {
	// Check if any rules already exist
	existing, _, err := alertsStore.ListAlertRules(false, 0, 0)
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		log.Printf("Alert rules already exist (%d), skipping defaults", len(existing))
		return nil
	}

	log.Println("Creating default alert rules...")

	// Default alert rules
	rules := []struct {
		rule    *models.AlertRule
		channel *models.AlertChannel
	}{
		{
			rule: &models.AlertRule{
				Name:              "PR Review Waiting > 48 Hours",
				Description:       "Alert when pull requests have been waiting for review for more than 48 hours",
				AlertType:         "pr_review_waiting",
				Enabled:           true,
				ThresholdValue:    floatPtr(48.0),
				ThresholdOperator: ">",
				Severity:          "warning",
				TargetEntity:      "org",
			},
			channel: &models.AlertChannel{
				ChannelType: "browser",
				Enabled:     true,
			},
		},
		{
			rule: &models.AlertRule{
				Name:              "Performance Score Drop > 20%",
				Description:       "Alert when an engineer's performance score drops by more than 20% week-over-week",
				AlertType:         "score_drop",
				Enabled:           true,
				ThresholdValue:    floatPtr(20.0),
				ThresholdOperator: ">",
				Severity:          "warning",
				TargetEntity:      "org",
			},
			channel: &models.AlertChannel{
				ChannelType: "browser",
				Enabled:     true,
			},
		},
		{
			rule: &models.AlertRule{
				Name:         "Burnout Signals Detected",
				Description:  "Alert when burnout risk indicators are detected (medium or high risk)",
				AlertType:    "burnout_signal",
				Enabled:      true,
				Severity:     "critical",
				TargetEntity: "org",
			},
			channel: &models.AlertChannel{
				ChannelType: "browser",
				Enabled:     true,
			},
		},
		{
			rule: &models.AlertRule{
				Name:              fmt.Sprintf("Sprint Behind Pace > %.0f%%", DefaultSprintBehindThreshold),
				Description:       fmt.Sprintf("Alert when sprint is more than %.0f%% behind expected progress", DefaultSprintBehindThreshold),
				AlertType:         "sprint_behind",
				Enabled:           true,
				ThresholdValue:    floatPtr(DefaultSprintBehindThreshold),
				ThresholdOperator: ">",
				Severity:          "warning",
				TargetEntity:      "org",
			},
			channel: &models.AlertChannel{
				ChannelType: "browser",
				Enabled:     true,
			},
		},
	}

	// Create rules and channels
	for _, r := range rules {
		if err := alertsStore.CreateAlertRule(r.rule); err != nil {
			log.Printf("Failed to create alert rule '%s': %v", r.rule.Name, err)
			continue
		}

		// Create browser notification channel
		r.channel.RuleID = r.rule.ID
		if err := alertsStore.CreateAlertChannel(r.channel); err != nil {
			log.Printf("Failed to create alert channel for '%s': %v", r.rule.Name, err)
			continue
		}

		log.Printf("Created default alert rule: %s", r.rule.Name)
	}

	return nil
}

// CreateSlackChannel creates a Slack notification channel for an alert rule
func CreateSlackChannel(alertsStore *db.AlertsStore, ruleID, webhookURL, channel string) error {
	config := models.AlertChannelConfig{
		WebhookURL: webhookURL,
		Channel:    channel,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	slackChannel := &models.AlertChannel{
		RuleID:        ruleID,
		ChannelType:   "slack",
		ChannelConfig: string(configJSON),
		Enabled:       true,
	}

	return alertsStore.CreateAlertChannel(slackChannel)
}

func floatPtr(f float64) *float64 {
	return &f
}
