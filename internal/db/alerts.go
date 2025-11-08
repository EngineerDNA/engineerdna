package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type AlertsStore struct {
	db *sql.DB
}

func NewAlertsStore(db *sql.DB) *AlertsStore {
	return &AlertsStore{db: db}
}

// Alert Rules CRUD

func (s *AlertsStore) CreateAlertRule(rule *models.AlertRule) error {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now().UTC()
	rule.UpdatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO alert_rules (
			id, name, description, alert_type, enabled,
			threshold_value, threshold_operator, severity,
			target_entity, target_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rule.ID, rule.Name, rule.Description, rule.AlertType, rule.Enabled,
		rule.ThresholdValue, rule.ThresholdOperator, rule.Severity,
		rule.TargetEntity, rule.TargetID, rule.CreatedAt, rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

func (s *AlertsStore) GetAlertRule(id string) (*models.AlertRule, error) {
	var rule models.AlertRule
	var thresholdValue sql.NullFloat64
	var thresholdOperator, targetEntity, targetID, description sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, description, alert_type, enabled,
		       threshold_value, threshold_operator, severity,
		       target_entity, target_id, created_at, updated_at
		FROM alert_rules
		WHERE id = ?
	`, id).Scan(&rule.ID, &rule.Name, &description, &rule.AlertType, &rule.Enabled,
		&thresholdValue, &thresholdOperator, &rule.Severity,
		&targetEntity, &targetID, &rule.CreatedAt, &rule.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if description.Valid {
		rule.Description = description.String
	}
	if thresholdValue.Valid {
		val := thresholdValue.Float64
		rule.ThresholdValue = &val
	}
	if thresholdOperator.Valid {
		rule.ThresholdOperator = thresholdOperator.String
	}
	if targetEntity.Valid {
		rule.TargetEntity = targetEntity.String
	}
	if targetID.Valid {
		rule.TargetID = targetID.String
	}

	return &rule, nil
}

func (s *AlertsStore) ListAlertRules(enabledOnly bool, limit, offset int) ([]*models.AlertRule, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM alert_rules"
	countArgs := []interface{}{}
	if enabledOnly {
		countQuery += " WHERE enabled = true"
	}

	var total int
	err := s.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alert rules: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, name, description, alert_type, enabled,
		       threshold_value, threshold_operator, severity,
		       target_entity, target_id, created_at, updated_at
		FROM alert_rules
	`

	if enabledOnly {
		query += " WHERE enabled = true"
	}

	query += " ORDER BY name LIMIT ? OFFSET ?"

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var thresholdValue sql.NullFloat64
		var thresholdOperator, targetEntity, targetID, description sql.NullString

		err := rows.Scan(&rule.ID, &rule.Name, &description, &rule.AlertType, &rule.Enabled,
			&thresholdValue, &thresholdOperator, &rule.Severity,
			&targetEntity, &targetID, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		if description.Valid {
			rule.Description = description.String
		}
		if thresholdValue.Valid {
			val := thresholdValue.Float64
			rule.ThresholdValue = &val
		}
		if thresholdOperator.Valid {
			rule.ThresholdOperator = thresholdOperator.String
		}
		if targetEntity.Valid {
			rule.TargetEntity = targetEntity.String
		}
		if targetID.Valid {
			rule.TargetID = targetID.String
		}

		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating alert rules: %w", err)
	}

	return rules, total, nil
}

func (s *AlertsStore) UpdateAlertRule(rule *models.AlertRule) error {
	rule.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE alert_rules
		SET name = ?, description = ?, alert_type = ?, enabled = ?,
		    threshold_value = ?, threshold_operator = ?, severity = ?,
		    target_entity = ?, target_id = ?, updated_at = ?
		WHERE id = ?
	`, rule.Name, rule.Description, rule.AlertType, rule.Enabled,
		rule.ThresholdValue, rule.ThresholdOperator, rule.Severity,
		rule.TargetEntity, rule.TargetID, rule.UpdatedAt, rule.ID)

	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found: %s", rule.ID)
	}

	return nil
}

func (s *AlertsStore) DeleteAlertRule(id string) error {
	result, err := s.db.Exec(`DELETE FROM alert_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found: %s", id)
	}

	return nil
}

// Alert Channels CRUD

func (s *AlertsStore) CreateAlertChannel(channel *models.AlertChannel) error {
	channel.ID = uuid.New().String()

	_, err := s.db.Exec(`
		INSERT INTO alert_channels (
			id, rule_id, channel_type, channel_config,
			enabled, quiet_hours_start, quiet_hours_end
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, channel.ID, channel.RuleID, channel.ChannelType, channel.ChannelConfig,
		channel.Enabled, channel.QuietHoursStart, channel.QuietHoursEnd)

	if err != nil {
		return fmt.Errorf("failed to create alert channel: %w", err)
	}

	return nil
}

func (s *AlertsStore) GetAlertChannels(ruleID string) ([]*models.AlertChannel, error) {
	rows, err := s.db.Query(`
		SELECT id, rule_id, channel_type, channel_config,
		       enabled, quiet_hours_start, quiet_hours_end
		FROM alert_channels
		WHERE rule_id = ?
		ORDER BY channel_type
	`, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert channels: %w", err)
	}
	defer rows.Close()

	var channels []*models.AlertChannel
	for rows.Next() {
		var channel models.AlertChannel
		var channelConfig, quietStart, quietEnd sql.NullString

		err := rows.Scan(&channel.ID, &channel.RuleID, &channel.ChannelType, &channelConfig,
			&channel.Enabled, &quietStart, &quietEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert channel: %w", err)
		}

		if channelConfig.Valid {
			channel.ChannelConfig = channelConfig.String
		}
		if quietStart.Valid {
			qs := quietStart.String
			channel.QuietHoursStart = &qs
		}
		if quietEnd.Valid {
			qe := quietEnd.String
			channel.QuietHoursEnd = &qe
		}

		channels = append(channels, &channel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert channels: %w", err)
	}

	return channels, nil
}

func (s *AlertsStore) UpdateAlertChannel(channel *models.AlertChannel) error {
	result, err := s.db.Exec(`
		UPDATE alert_channels
		SET channel_type = ?, channel_config = ?, enabled = ?,
		    quiet_hours_start = ?, quiet_hours_end = ?
		WHERE id = ?
	`, channel.ChannelType, channel.ChannelConfig, channel.Enabled,
		channel.QuietHoursStart, channel.QuietHoursEnd, channel.ID)

	if err != nil {
		return fmt.Errorf("failed to update alert channel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert channel not found: %s", channel.ID)
	}

	return nil
}

func (s *AlertsStore) DeleteAlertChannel(id string) error {
	result, err := s.db.Exec(`DELETE FROM alert_channels WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert channel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert channel not found: %s", id)
	}

	return nil
}

// Alert Instances CRUD

func (s *AlertsStore) CreateAlertInstance(instance *models.AlertInstance) error {
	instance.ID = uuid.New().String()
	instance.FiredAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO alert_instances (
			id, rule_id, title, message, severity,
			entity_type, entity_id, context, fired_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, instance.ID, instance.RuleID, instance.Title, instance.Message, instance.Severity,
		instance.EntityType, instance.EntityID, instance.Context, instance.FiredAt)

	if err != nil {
		return fmt.Errorf("failed to create alert instance: %w", err)
	}

	return nil
}

func (s *AlertsStore) GetAlertInstance(id string) (*models.AlertInstance, error) {
	var instance models.AlertInstance
	var entityType, entityID, context sql.NullString
	var acknowledgedAt, snoozedUntil, dismissedAt, resolvedAt sql.NullString
	var acknowledgedBy, dismissedBy sql.NullString

	err := s.db.QueryRow(`
		SELECT id, rule_id, title, message, severity,
		       entity_type, entity_id, context, fired_at,
		       acknowledged_at, acknowledged_by, snoozed_until,
		       dismissed_at, dismissed_by, resolved_at
		FROM alert_instances
		WHERE id = ?
	`, id).Scan(&instance.ID, &instance.RuleID, &instance.Title, &instance.Message, &instance.Severity,
		&entityType, &entityID, &context, &instance.FiredAt,
		&acknowledgedAt, &acknowledgedBy, &snoozedUntil,
		&dismissedAt, &dismissedBy, &resolvedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert instance: %w", err)
	}

	if entityType.Valid {
		instance.EntityType = entityType.String
	}
	if entityID.Valid {
		instance.EntityID = entityID.String
	}
	if context.Valid {
		instance.Context = context.String
	}
	if acknowledgedAt.Valid {
		t, _ := time.Parse(time.RFC3339, acknowledgedAt.String)
		instance.AcknowledgedAt = &t
	}
	if acknowledgedBy.Valid {
		ab := acknowledgedBy.String
		instance.AcknowledgedBy = &ab
	}
	if snoozedUntil.Valid {
		t, _ := time.Parse(time.RFC3339, snoozedUntil.String)
		instance.SnoozedUntil = &t
	}
	if dismissedAt.Valid {
		t, _ := time.Parse(time.RFC3339, dismissedAt.String)
		instance.DismissedAt = &t
	}
	if dismissedBy.Valid {
		db := dismissedBy.String
		instance.DismissedBy = &db
	}
	if resolvedAt.Valid {
		t, _ := time.Parse(time.RFC3339, resolvedAt.String)
		instance.ResolvedAt = &t
	}

	return &instance, nil
}

func (s *AlertsStore) ListAlertInstances(filters map[string]string, limit, offset int) ([]*models.AlertInstance, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = MaxQueryLimit
	}

	// Build WHERE clause for both count and data queries
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if ruleID, ok := filters["rule_id"]; ok && ruleID != "" {
		whereClause += " AND rule_id = ?"
		args = append(args, ruleID)
	}

	if severity, ok := filters["severity"]; ok && severity != "" {
		whereClause += " AND severity = ?"
		args = append(args, severity)
	}

	if entityType, ok := filters["entity_type"]; ok && entityType != "" {
		whereClause += " AND entity_type = ?"
		args = append(args, entityType)
	}

	if entityID, ok := filters["entity_id"]; ok && entityID != "" {
		whereClause += " AND entity_id = ?"
		args = append(args, entityID)
	}

	if status, ok := filters["status"]; ok {
		switch status {
		case "active":
			whereClause += " AND dismissed_at IS NULL AND resolved_at IS NULL"
		case "acknowledged":
			whereClause += " AND acknowledged_at IS NOT NULL AND dismissed_at IS NULL AND resolved_at IS NULL"
		case "snoozed":
			whereClause += " AND snoozed_until IS NOT NULL AND snoozed_until > datetime('now') AND dismissed_at IS NULL AND resolved_at IS NULL"
		case "dismissed":
			whereClause += " AND dismissed_at IS NOT NULL"
		case "resolved":
			whereClause += " AND resolved_at IS NOT NULL"
		}
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM alert_instances" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alert instances: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, rule_id, title, message, severity,
		       entity_type, entity_id, context, fired_at,
		       acknowledged_at, acknowledged_by, snoozed_until,
		       dismissed_at, dismissed_by, resolved_at
		FROM alert_instances` + whereClause + `
		ORDER BY fired_at DESC
		LIMIT ? OFFSET ?
	`

	paginatedArgs := append(args, limit, offset)
	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list alert instances: %w", err)
	}
	defer rows.Close()

	var instances []*models.AlertInstance
	for rows.Next() {
		var instance models.AlertInstance
		var entityType, entityID, context sql.NullString
		var acknowledgedAt, snoozedUntil, dismissedAt, resolvedAt sql.NullString
		var acknowledgedBy, dismissedBy sql.NullString

		err := rows.Scan(&instance.ID, &instance.RuleID, &instance.Title, &instance.Message, &instance.Severity,
			&entityType, &entityID, &context, &instance.FiredAt,
			&acknowledgedAt, &acknowledgedBy, &snoozedUntil,
			&dismissedAt, &dismissedBy, &resolvedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alert instance: %w", err)
		}

		if entityType.Valid {
			instance.EntityType = entityType.String
		}
		if entityID.Valid {
			instance.EntityID = entityID.String
		}
		if context.Valid {
			instance.Context = context.String
		}
		if acknowledgedAt.Valid {
			t, _ := time.Parse(time.RFC3339, acknowledgedAt.String)
			instance.AcknowledgedAt = &t
		}
		if acknowledgedBy.Valid {
			ab := acknowledgedBy.String
			instance.AcknowledgedBy = &ab
		}
		if snoozedUntil.Valid {
			t, _ := time.Parse(time.RFC3339, snoozedUntil.String)
			instance.SnoozedUntil = &t
		}
		if dismissedAt.Valid {
			t, _ := time.Parse(time.RFC3339, dismissedAt.String)
			instance.DismissedAt = &t
		}
		if dismissedBy.Valid {
			db := dismissedBy.String
			instance.DismissedBy = &db
		}
		if resolvedAt.Valid {
			t, _ := time.Parse(time.RFC3339, resolvedAt.String)
			instance.ResolvedAt = &t
		}

		instances = append(instances, &instance)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating alert instances: %w", err)
	}

	return instances, total, nil
}

func (s *AlertsStore) AcknowledgeAlert(instanceID, acknowledgedBy string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE alert_instances
		SET acknowledged_at = ?, acknowledged_by = ?
		WHERE id = ? AND dismissed_at IS NULL AND resolved_at IS NULL
	`, now, acknowledgedBy, instanceID)

	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert instance not found or already dismissed/resolved: %s", instanceID)
	}

	return nil
}

func (s *AlertsStore) SnoozeAlert(instanceID string, snoozedUntil time.Time) error {
	result, err := s.db.Exec(`
		UPDATE alert_instances
		SET snoozed_until = ?
		WHERE id = ? AND dismissed_at IS NULL AND resolved_at IS NULL
	`, snoozedUntil, instanceID)

	if err != nil {
		return fmt.Errorf("failed to snooze alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert instance not found or already dismissed/resolved: %s", instanceID)
	}

	return nil
}

func (s *AlertsStore) DismissAlert(instanceID, dismissedBy string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE alert_instances
		SET dismissed_at = ?, dismissed_by = ?
		WHERE id = ? AND resolved_at IS NULL
	`, now, dismissedBy, instanceID)

	if err != nil {
		return fmt.Errorf("failed to dismiss alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert instance not found or already resolved: %s", instanceID)
	}

	return nil
}

func (s *AlertsStore) ResolveAlert(instanceID string) error {
	now := time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE alert_instances
		SET resolved_at = ?
		WHERE id = ?
	`, now, instanceID)

	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert instance not found: %s", instanceID)
	}

	return nil
}

// Alert Deliveries

func (s *AlertsStore) CreateAlertDelivery(delivery *models.AlertDelivery) error {
	delivery.ID = uuid.New().String()
	delivery.DeliveredAt = time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO alert_deliveries (
			id, instance_id, channel_type, delivered_at, status, error_message
		) VALUES (?, ?, ?, ?, ?, ?)
	`, delivery.ID, delivery.InstanceID, delivery.ChannelType,
		delivery.DeliveredAt, delivery.Status, delivery.ErrorMessage)

	if err != nil {
		return fmt.Errorf("failed to create alert delivery: %w", err)
	}

	return nil
}

func (s *AlertsStore) GetAlertDeliveries(instanceID string) ([]*models.AlertDelivery, error) {
	rows, err := s.db.Query(`
		SELECT id, instance_id, channel_type, delivered_at, status, error_message
		FROM alert_deliveries
		WHERE instance_id = ?
		ORDER BY delivered_at DESC
	`, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.AlertDelivery
	for rows.Next() {
		var delivery models.AlertDelivery
		var errorMessage sql.NullString

		err := rows.Scan(&delivery.ID, &delivery.InstanceID, &delivery.ChannelType,
			&delivery.DeliveredAt, &delivery.Status, &errorMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert delivery: %w", err)
		}

		if errorMessage.Valid {
			em := errorMessage.String
			delivery.ErrorMessage = &em
		}

		deliveries = append(deliveries, &delivery)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert deliveries: %w", err)
	}

	return deliveries, nil
}
