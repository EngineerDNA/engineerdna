package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/utils"
)

// SeedAlerts creates alert system data
func SeedAlerts(data *SeedData) (alertRuleIDs, alertInstanceIDs []string) {
	fmt.Println("\n10. Creating alert system data...")

	// Alert Rules (only using supported types: pr_review_waiting, score_drop, burnout_signal, sprint_behind)
	alertRules := []struct {
		name      string
		alertType string
		severity  string
		threshold *float64
		operator  string
		target    string
		targetID  *string
	}{
		{"PR Review Waiting Too Long", "pr_review_waiting", "warning", utils.PtrFloat64(24.0), ">", "org", nil},
		{"Performance Score Drop", "score_drop", "warning", utils.PtrFloat64(15.0), ">", "team", &data.BackendTeam.ID},
		{"Sprint Behind Schedule", "sprint_behind", "critical", utils.PtrFloat64(20.0), ">", "team", &data.FrontendTeam.ID},
		{"Burnout Risk", "burnout_signal", "critical", nil, "", "engineer", &data.EngineerIDs[6]}, // Maya Patel
	}

	alertRuleIDs = make([]string, 0)
	for _, r := range alertRules {
		rule := &models.AlertRule{
			Name:              r.name,
			Description:       fmt.Sprintf("Monitors %s and alerts when threshold is exceeded", r.alertType),
			AlertType:         r.alertType,
			Enabled:           true,
			ThresholdValue:    r.threshold,
			ThresholdOperator: r.operator,
			Severity:          r.severity,
			TargetEntity:      r.target,
			TargetID:          GetStringValue(r.targetID),
		}
		if err := data.AlertsStore.CreateAlertRule(rule); err != nil {
			log.Printf("Warning: Failed to create alert rule: %v", err)
		} else {
			alertRuleIDs = append(alertRuleIDs, rule.ID)
		}
	}
	fmt.Printf("  Created %d alert rules\n", len(alertRuleIDs))

	// Alert Instances
	alertInstances := []struct {
		ruleIndex  int
		title      string
		message    string
		entityType string
		entityID   *string
		acked      bool
		dismissed  bool
		resolved   bool
	}{
		{0, "PR waiting for review: Implement caching layer", "Pull request has been waiting for review for 28.5 hours (threshold: 24 hours)", "pr", nil, true, false, false},
		{1, "Performance score drop: Backend Team", "Score dropped 18.5% (from 85.0 to 69.2)", "team", &data.BackendTeam.ID, false, false, false},
		{2, "Sprint behind pace: Frontend Sprint 47", "Sprint is 22.0% behind expected progress (15/40 points completed, 50% expected)", "sprint", nil, true, false, true},
		{3, "High workload detected for Maya Patel", "Working hours exceed 50hrs/week for 3 consecutive weeks", "engineer", &data.EngineerIDs[6], true, false, false},
	}

	alertInstanceIDs = make([]string, 0)
	for i, a := range alertInstances {
		if a.ruleIndex >= len(alertRuleIDs) {
			continue
		}
		instance := &models.AlertInstance{
			RuleID:     alertRuleIDs[a.ruleIndex],
			Title:      a.title,
			Message:    a.message,
			Severity:   alertRules[a.ruleIndex].severity,
			EntityType: a.entityType,
			EntityID:   GetStringValue(a.entityID),
			FiredAt:    data.Now.AddDate(0, 0, -i-1),
		}
		if a.acked {
			ackTime := data.Now.AddDate(0, 0, -i).Add(7200000000000) // 2 hours
			instance.AcknowledgedAt = &ackTime
			instance.AcknowledgedBy = &data.EngineerIDs[0] // Sarah Chen
		}
		if a.dismissed {
			dismissTime := data.Now.AddDate(0, 0, -i).Add(14400000000000) // 4 hours
			instance.DismissedAt = &dismissTime
			instance.DismissedBy = &data.EngineerIDs[0]
		}
		if a.resolved {
			resolveTime := data.Now.AddDate(0, 0, -i).Add(21600000000000) // 6 hours
			instance.ResolvedAt = &resolveTime
		}
		if err := data.AlertsStore.CreateAlertInstance(instance); err != nil {
			log.Printf("Warning: Failed to create alert instance: %v", err)
		} else {
			alertInstanceIDs = append(alertInstanceIDs, instance.ID)
		}
	}
	fmt.Printf("  Created %d alert instances\n", len(alertInstanceIDs))

	// Alert Channels
	for _, ruleID := range alertRuleIDs[:3] { // First 3 rules have channels
		channel := &models.AlertChannel{
			RuleID:        ruleID,
			ChannelType:   "slack",
			ChannelConfig: `{"webhook_url": "https://hooks.slack.com/services/EXAMPLE", "channel": "#eng-alerts"}`,
			Enabled:       true,
		}
		if err := data.AlertsStore.CreateAlertChannel(channel); err != nil {
			log.Printf("Warning: Failed to create alert channel: %v", err)
		}
	}
	fmt.Printf("  Created alert channels\n")

	return alertRuleIDs, alertInstanceIDs
}
