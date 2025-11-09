-- Migration 018: Alert System
-- Real-Time Operations Alert Engine for monitoring metrics and sending alerts

-- Alert Rules: Define what triggers alerts
CREATE TABLE IF NOT EXISTS alert_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    alert_type TEXT NOT NULL, -- 'pr_review_waiting', 'score_drop', 'burnout_signal', 'sprint_behind'
    enabled BOOLEAN DEFAULT true,
    threshold_value REAL, -- e.g., 48 hours for PR waiting
    threshold_operator TEXT, -- '>', '<', '>=', '<=', '=='
    severity TEXT NOT NULL, -- 'info', 'warning', 'critical'
    target_entity TEXT, -- 'engineer', 'team', 'org'
    target_id TEXT, -- engineer_id, team_id, or null for org-wide
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Alert Delivery Channels: How to deliver alerts
CREATE TABLE IF NOT EXISTS alert_channels (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    channel_type TEXT NOT NULL, -- 'slack', 'email', 'browser'
    channel_config TEXT, -- JSON: {"webhook_url": "...", "channel": "#alerts"}
    enabled BOOLEAN DEFAULT true,
    quiet_hours_start TEXT, -- "18:00"
    quiet_hours_end TEXT, -- "09:00"
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

-- Alert Instances: Fired alerts
CREATE TABLE IF NOT EXISTS alert_instances (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL,
    entity_type TEXT, -- 'engineer', 'team', 'pr', etc.
    entity_id TEXT,
    context TEXT, -- JSON: additional context data
    fired_at TEXT NOT NULL,
    acknowledged_at TEXT,
    acknowledged_by TEXT,
    snoozed_until TEXT,
    dismissed_at TEXT,
    dismissed_by TEXT,
    resolved_at TEXT,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

-- Alert Deliveries: Track delivery attempts
CREATE TABLE IF NOT EXISTS alert_deliveries (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    channel_type TEXT NOT NULL,
    delivered_at TEXT NOT NULL,
    status TEXT NOT NULL, -- 'sent', 'failed', 'skipped_quiet_hours'
    error_message TEXT,
    FOREIGN KEY (instance_id) REFERENCES alert_instances(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_target ON alert_rules(target_entity, target_id);
CREATE INDEX IF NOT EXISTS idx_alert_instances_rule ON alert_instances(rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_instances_fired ON alert_instances(fired_at);
CREATE INDEX IF NOT EXISTS idx_alert_instances_entity ON alert_instances(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_alert_deliveries_instance ON alert_deliveries(instance_id);
