package alerts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Evaluator checks alert rules and fires alerts when thresholds are breached
type Evaluator struct {
	database     *sql.DB
	alertsStore  *db.AlertsStore
	eventStore   *db.EventStore
	scoringStore *db.ScoringStore
	teamStore    *db.TeamStore
}

// NewEvaluator creates a new alert evaluator
func NewEvaluator(database *sql.DB, alertsStore *db.AlertsStore, eventStore *db.EventStore, scoringStore *db.ScoringStore, teamStore *db.TeamStore) *Evaluator {
	return &Evaluator{
		database:     database,
		alertsStore:  alertsStore,
		eventStore:   eventStore,
		scoringStore: scoringStore,
		teamStore:    teamStore,
	}
}

// EvaluateAll evaluates all enabled alert rules
func (e *Evaluator) EvaluateAll() error {
	rules, _, err := e.alertsStore.ListAlertRules(true, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list alert rules: %w", err)
	}

	for _, rule := range rules {
		if err := e.EvaluateRule(rule); err != nil {
			log.Printf("Error evaluating rule %s (%s): %v", rule.Name, rule.ID, err)
		}
	}

	return nil
}

// EvaluateRule evaluates a single alert rule
func (e *Evaluator) EvaluateRule(rule *models.AlertRule) error {
	switch rule.AlertType {
	case "pr_review_waiting":
		return e.EvaluatePRReviewWaiting(rule)
	case "score_drop":
		return e.EvaluateScoreDrop(rule)
	case "burnout_signal":
		return e.EvaluateBurnoutSignals(rule)
	case "sprint_behind":
		return e.EvaluateSprintProgress(rule)
	default:
		return fmt.Errorf("unknown alert type: %s", rule.AlertType)
	}
}

// EvaluatePRReviewWaiting checks for PRs waiting for review beyond threshold
func (e *Evaluator) EvaluatePRReviewWaiting(rule *models.AlertRule) error {
	if rule.ThresholdValue == nil {
		return fmt.Errorf("threshold_value required for pr_review_waiting alert")
	}

	thresholdHours := *rule.ThresholdValue
	cutoffTime := time.Now().UTC().Add(-time.Duration(thresholdHours) * time.Hour)

	// Query for PR events that are still waiting for review
	query := `
		SELECT id, source_id, data, timestamp, engineer_id
		FROM events
		WHERE type = 'pull_request'
		  AND timestamp < ?
		  AND json_extract(data, '$.state') IN ('open', 'review_requested')
		ORDER BY timestamp ASC
	`

	rows, err := e.database.Query(query, cutoffTime.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to query PR events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventID, sourceID, dataJSON, engineerID string
		var timestamp time.Time

		if err := rows.Scan(&eventID, &sourceID, &dataJSON, &timestamp, &engineerID); err != nil {
			log.Printf("Error scanning PR event: %v", err)
			continue
		}

		// Check if we already have an active alert for this PR
		existingAlerts, _, err := e.alertsStore.ListAlertInstances(map[string]string{
			"rule_id":     rule.ID,
			"entity_type": "pr",
			"entity_id":   sourceID,
			"status":      "active",
		}, 1, 0)
		if err != nil {
			log.Printf("Error checking existing alerts: %v", err)
			continue
		}

		if len(existingAlerts) > 0 {
			continue
		}

		// Parse PR data
		var prData map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &prData); err != nil {
			log.Printf("Error parsing PR data: %v", err)
			continue
		}

		prTitle, _ := prData["title"].(string)
		prAge := time.Since(timestamp).Hours()

		// Create alert context
		context := models.AlertContext{
			PRID:    sourceID,
			PRTitle: prTitle,
			PRAge:   prAge,
		}
		contextJSON, _ := json.Marshal(context)

		// Fire alert
		instance := &models.AlertInstance{
			RuleID:     rule.ID,
			Title:      fmt.Sprintf("PR waiting for review: %s", prTitle),
			Message:    fmt.Sprintf("Pull request has been waiting for review for %.1f hours (threshold: %.0f hours)", prAge, thresholdHours),
			Severity:   rule.Severity,
			EntityType: "pr",
			EntityID:   sourceID,
			Context:    string(contextJSON),
		}

		if err := e.alertsStore.CreateAlertInstance(instance); err != nil {
			log.Printf("Error creating alert instance: %v", err)
			continue
		}

		log.Printf("Fired alert: %s (PR: %s, age: %.1fh)", instance.Title, sourceID, prAge)
	}

	return rows.Err()
}

// EvaluateScoreDrop checks for significant performance score drops
func (e *Evaluator) EvaluateScoreDrop(rule *models.AlertRule) error {
	if rule.ThresholdValue == nil {
		return fmt.Errorf("threshold_value required for score_drop alert")
	}

	dropPercentage := *rule.ThresholdValue

	// Get engineers to check (based on target)
	var engineerIDs []string
	if rule.TargetEntity == "engineer" && rule.TargetID != "" {
		engineerIDs = []string{rule.TargetID}
	} else {
		// Check all engineers
		rows, err := e.database.Query("SELECT id FROM engineers WHERE active = true LIMIT ?", db.MaxQueryLimit)
		if err != nil {
			return fmt.Errorf("failed to query engineers: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				continue
			}
			engineerIDs = append(engineerIDs, id)
		}
	}

	now := time.Now().UTC()
	currentWeek := now.Truncate(7 * 24 * time.Hour)
	previousWeek := currentWeek.Add(-7 * 24 * time.Hour)

	for _, engineerID := range engineerIDs {
		// Get current and previous week scores
		currentScores, err := e.scoringStore.GetPerformanceScores(engineerID, currentWeek, now, 10)
		if err != nil || len(currentScores) == 0 {
			continue
		}

		previousScores, err := e.scoringStore.GetPerformanceScores(engineerID, previousWeek, previousWeek, 10)
		if err != nil || len(previousScores) == 0 {
			continue
		}

		currentScore := currentScores[0].TotalScore
		previousScore := previousScores[0].TotalScore

		if previousScore == 0 {
			continue
		}

		drop := ((previousScore - currentScore) / previousScore) * 100

		if drop >= dropPercentage {
			// Check if we already have an active alert for this engineer
			existingAlerts, _, err := e.alertsStore.ListAlertInstances(map[string]string{
				"rule_id":     rule.ID,
				"entity_type": "engineer",
				"entity_id":   engineerID,
				"status":      "active",
			}, 1, 0)
			if err != nil || len(existingAlerts) > 0 {
				continue
			}

			// Get engineer name
			var engineerName string
			e.database.QueryRow("SELECT canonical_name FROM engineers WHERE id = ?", engineerID).Scan(&engineerName)

			// Create alert context
			context := models.AlertContext{
				EngineerName:  engineerName,
				ScoreCurrent:  currentScore,
				ScorePrevious: previousScore,
				ScoreDrop:     drop,
			}
			contextJSON, _ := json.Marshal(context)

			// Fire alert
			instance := &models.AlertInstance{
				RuleID:     rule.ID,
				Title:      fmt.Sprintf("Performance score drop: %s", engineerName),
				Message:    fmt.Sprintf("Score dropped %.1f%% (from %.1f to %.1f)", drop, previousScore, currentScore),
				Severity:   rule.Severity,
				EntityType: "engineer",
				EntityID:   engineerID,
				Context:    string(contextJSON),
			}

			if err := e.alertsStore.CreateAlertInstance(instance); err != nil {
				log.Printf("Error creating alert instance: %v", err)
				continue
			}

			log.Printf("Fired alert: %s (drop: %.1f%%)", instance.Title, drop)
		}
	}

	return nil
}

// EvaluateBurnoutSignals checks for burnout indicators
func (e *Evaluator) EvaluateBurnoutSignals(rule *models.AlertRule) error {
	// Get engineers to check
	var engineerIDs []string
	if rule.TargetEntity == "engineer" && rule.TargetID != "" {
		engineerIDs = []string{rule.TargetID}
	} else {
		rows, err := e.database.Query("SELECT id FROM engineers WHERE active = true LIMIT ?", db.MaxQueryLimit)
		if err != nil {
			return fmt.Errorf("failed to query engineers: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				continue
			}
			engineerIDs = append(engineerIDs, id)
		}
	}

	now := time.Now().UTC()
	weekStart := now.Truncate(7 * 24 * time.Hour)

	for _, engineerID := range engineerIDs {
		// Get performance scores with burnout risk
		scores, err := e.scoringStore.GetPerformanceScores(engineerID, weekStart, now, 10)
		if err != nil || len(scores) == 0 {
			continue
		}

		score := scores[0]
		if score.BurnoutRisk == "" {
			continue
		}

		var burnoutRisk models.BurnoutRisk
		if err := json.Unmarshal([]byte(score.BurnoutRisk), &burnoutRisk); err != nil {
			continue
		}

		if burnoutRisk.RiskLevel == "none" || burnoutRisk.RiskLevel == "low" {
			continue
		}

		// Check if we already have an active alert for this engineer
		existingAlerts, _, err := e.alertsStore.ListAlertInstances(map[string]string{
			"rule_id":     rule.ID,
			"entity_type": "engineer",
			"entity_id":   engineerID,
			"status":      "active",
		}, 1, 0)
		if err != nil || len(existingAlerts) > 0 {
			continue
		}

		// Get engineer name
		var engineerName string
		e.database.QueryRow("SELECT canonical_name FROM engineers WHERE id = ?", engineerID).Scan(&engineerName)

		// Create alert context
		context := models.AlertContext{
			EngineerName: engineerName,
			Additional: map[string]interface{}{
				"risk_level":     burnoutRisk.RiskLevel,
				"recommendation": burnoutRisk.Recommendation,
				"red_flags":      burnoutRisk.RedFlags,
			},
		}
		contextJSON, _ := json.Marshal(context)

		// Fire alert
		instance := &models.AlertInstance{
			RuleID:     rule.ID,
			Title:      fmt.Sprintf("Burnout risk detected: %s", engineerName),
			Message:    fmt.Sprintf("Risk level: %s. %s", burnoutRisk.RiskLevel, burnoutRisk.Recommendation),
			Severity:   rule.Severity,
			EntityType: "engineer",
			EntityID:   engineerID,
			Context:    string(contextJSON),
		}

		if err := e.alertsStore.CreateAlertInstance(instance); err != nil {
			log.Printf("Error creating alert instance: %v", err)
			continue
		}

		log.Printf("Fired alert: %s (risk: %s)", instance.Title, burnoutRisk.RiskLevel)
	}

	return nil
}

// EvaluateSprintProgress checks if sprint is behind pace
func (e *Evaluator) EvaluateSprintProgress(rule *models.AlertRule) error {
	if rule.ThresholdValue == nil {
		return fmt.Errorf("threshold_value required for sprint_behind alert")
	}

	behindPercentage := *rule.ThresholdValue

	// Query active sprints
	query := `
		SELECT id, name, start_date, end_date, committed_points, completed_points, team_id
		FROM sprints
		WHERE status = 'active'
		  AND date('now') BETWEEN start_date AND end_date
	`

	rows, err := e.database.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query sprints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sprintID, name, teamID string
		var startDate, endDate string
		var committedPoints, completedPoints int

		if err := rows.Scan(&sprintID, &name, &startDate, &endDate, &committedPoints, &completedPoints, &teamID); err != nil {
			log.Printf("Error scanning sprint: %v", err)
			continue
		}

		if committedPoints == 0 {
			continue
		}

		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		now := time.Now().UTC()

		sprintDuration := end.Sub(start).Hours() / 24
		elapsed := now.Sub(start).Hours() / 24
		expectedProgress := (elapsed / sprintDuration) * 100
		actualProgress := (float64(completedPoints) / float64(committedPoints)) * 100

		behind := expectedProgress - actualProgress

		if behind >= behindPercentage {
			// Check if we already have an active alert for this sprint
			existingAlerts, _, err := e.alertsStore.ListAlertInstances(map[string]string{
				"rule_id":     rule.ID,
				"entity_type": "sprint",
				"entity_id":   sprintID,
				"status":      "active",
			}, 1, 0)
			if err != nil || len(existingAlerts) > 0 {
				continue
			}

			// Create alert context
			context := models.AlertContext{
				SprintName:     name,
				SprintProgress: actualProgress,
				Additional: map[string]interface{}{
					"committed_points":  committedPoints,
					"completed_points":  completedPoints,
					"expected_progress": expectedProgress,
					"behind_percentage": behind,
				},
			}
			contextJSON, _ := json.Marshal(context)

			// Fire alert
			instance := &models.AlertInstance{
				RuleID:     rule.ID,
				Title:      fmt.Sprintf("Sprint behind pace: %s", name),
				Message:    fmt.Sprintf("Sprint is %.1f%% behind expected progress (%d/%d points completed, %.0f%% expected)", behind, completedPoints, committedPoints, expectedProgress),
				Severity:   rule.Severity,
				EntityType: "sprint",
				EntityID:   sprintID,
				Context:    string(contextJSON),
			}

			if err := e.alertsStore.CreateAlertInstance(instance); err != nil {
				log.Printf("Error creating alert instance: %v", err)
				continue
			}

			log.Printf("Fired alert: %s (behind: %.1f%%)", instance.Title, behind)
		}
	}

	return rows.Err()
}
