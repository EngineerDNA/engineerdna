package db

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations inline
	migrations := []struct {
		version string
		sql     string
	}{
		{
			version: "001_initial_schema",
			sql: `
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,
    anonymized BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_id)
);

CREATE TABLE IF NOT EXISTS plugin_configs (
    name TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    config TEXT NOT NULL,
    last_sync DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    external_ids TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_actor ON events(actor);
CREATE INDEX idx_events_source ON events(source);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_anonymized ON events(anonymized);
			`,
		},
		{
			version: "002_anonymization",
			sql: `
CREATE TABLE IF NOT EXISTS anonymization_map (
    id TEXT PRIMARY KEY,
    anonymized_id TEXT UNIQUE NOT NULL,
    real_name TEXT NOT NULL,
    real_email TEXT,
    external_identifiers TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS anonymization_policies (
    plugin_name TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT 0,
    strategy TEXT NOT NULL,
    fields TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_anon_map_real ON anonymization_map(real_name);
CREATE INDEX idx_anon_map_anon ON anonymization_map(anonymized_id);
			`,
		},
		{
			version: "003_audit_log",
			sql: `
CREATE TABLE IF NOT EXISTS audit_log (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    plugin_name TEXT NOT NULL,
    action TEXT NOT NULL,
    event_count INTEGER NOT NULL,
    anonymized BOOLEAN NOT NULL,
    destination TEXT,
    data_summary TEXT,
    user_initiated BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    generated_at DATETIME NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    severity TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    recommendation TEXT,
    metrics TEXT NOT NULL,
    dismissed BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exports (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    exported_at DATETIME NOT NULL,
    status TEXT NOT NULL,
    destination_url TEXT,
    event_count INTEGER NOT NULL,
    anonymized BOOLEAN NOT NULL,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_plugin ON audit_log(plugin_name);
CREATE INDEX idx_insights_generated ON insights(generated_at);
CREATE INDEX idx_exports_exported ON exports(exported_at);
			`,
		},
		{
			version: "004_identity_system",
			sql: `
CREATE TABLE IF NOT EXISTS engineers (
    id TEXT PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    email TEXT,
    manager TEXT,
    identifiers TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS unresolved_identities (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    identifier TEXT NOT NULL,
    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_count INTEGER NOT NULL DEFAULT 1,
    UNIQUE(source, identifier)
);

ALTER TABLE events ADD COLUMN engineer_id TEXT;

CREATE INDEX idx_engineers_name ON engineers(canonical_name);
CREATE INDEX idx_engineers_email ON engineers(email);
CREATE INDEX idx_engineers_active ON engineers(active);
CREATE INDEX idx_unresolved_source ON unresolved_identities(source);
CREATE INDEX idx_unresolved_identifier ON unresolved_identities(identifier);
CREATE INDEX idx_events_engineer_id ON events(engineer_id);
			`,
		},
		{
			version: "005_add_ignored_flag",
			sql: `
ALTER TABLE unresolved_identities ADD COLUMN ignored BOOLEAN NOT NULL DEFAULT 0;
CREATE INDEX idx_unresolved_ignored ON unresolved_identities(ignored);
			`,
		},
		{
			version: "006_add_events_updated_at_index",
			sql: `
CREATE INDEX IF NOT EXISTS idx_events_updated_at ON events(updated_at);
			`,
		},
		{
			version: "007_add_app_config",
			sql: `
CREATE TABLE IF NOT EXISTS app_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
			`,
		},
		{
			version: "008_add_export_schedules",
			sql: `
CREATE TABLE IF NOT EXISTS export_schedules (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    frequency TEXT NOT NULL,
    day_of_week INTEGER,
    time_of_day TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    last_run DATETIME,
    next_run DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_export_schedules_next_run ON export_schedules(enabled, next_run);
CREATE INDEX idx_export_schedules_plugin ON export_schedules(plugin_name);
			`,
		},
		{
			version: "009_performance_indexes",
			sql: `
CREATE INDEX IF NOT EXISTS idx_events_engineer_timestamp ON events(engineer_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_events_type_timestamp ON events(type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_events_source_timestamp ON events(source, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_events_composite ON events(engineer_id, type, source, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_unresolved_source_identifier ON unresolved_identities(source, identifier);
ANALYZE;
			`,
		},
		{
			version: "010_scoring_system",
			sql: `
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    target_score INTEGER DEFAULT 100,
    expectations TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scoring_weights (
    id TEXT PRIMARY KEY,
    throughput_weight REAL DEFAULT 30,
    quality_weight REAL DEFAULT 25,
    speed_weight REAL DEFAULT 20,
    collaboration_weight REAL DEFAULT 15,
    impact_weight REAL DEFAULT 10,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS performance_scores (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    week_start DATE NOT NULL,
    total_score REAL,
    throughput_score REAL,
    quality_score REAL,
    speed_score REAL,
    collaboration_score REAL,
    impact_score REAL,
    raw_metrics TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (engineer_id) REFERENCES engineers(id),
    UNIQUE(engineer_id, week_start)
);

CREATE TABLE IF NOT EXISTS team_performance_scores (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    week_start DATE NOT NULL,
    total_score REAL,
    member_count INTEGER,
    throughput_score REAL,
    quality_score REAL,
    speed_score REAL,
    collaboration_score REAL,
    impact_score REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(team_id, week_start)
);

CREATE TABLE IF NOT EXISTS promotion_signals (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    signal_type TEXT,
    detected_at DATETIME,
    weeks_duration INTEGER,
    avg_score REAL,
    dismissed BOOLEAN DEFAULT FALSE,
    dismissed_at DATETIME,
    notes TEXT,

    FOREIGN KEY (engineer_id) REFERENCES engineers(id)
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_performance_scores_engineer ON performance_scores(engineer_id);
CREATE INDEX IF NOT EXISTS idx_performance_scores_week ON performance_scores(week_start);
CREATE INDEX IF NOT EXISTS idx_performance_scores_engineer_week ON performance_scores(engineer_id, week_start DESC);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_team ON team_performance_scores(team_id);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_week ON team_performance_scores(week_start);
CREATE INDEX IF NOT EXISTS idx_promotion_signals_engineer ON promotion_signals(engineer_id);
CREATE INDEX IF NOT EXISTS idx_promotion_signals_dismissed ON promotion_signals(dismissed);
			`,
		},
		{
			version: "011_add_engineer_role_id",
			sql: `
ALTER TABLE engineers ADD COLUMN role_id TEXT;
CREATE INDEX IF NOT EXISTS idx_engineers_role_id ON engineers(role_id);
			`,
		},
		{
			version: "012_team_hierarchy",
			sql: `
CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_team_id TEXT,
    manager_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_team_id) REFERENCES teams(id) ON DELETE SET NULL,
    FOREIGN KEY (manager_id) REFERENCES engineers(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS team_membership (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    role TEXT,
    joined_at DATETIME NOT NULL,
    left_at DATETIME,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES engineers(id) ON DELETE CASCADE,
    UNIQUE(team_id, member_id, joined_at)
);

CREATE TABLE IF NOT EXISTS view_configurations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT,
    default_mode TEXT,
    team_filter TEXT,
    alert_filter TEXT,
    sort_order TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team_performance_scores_new (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    week_start DATE NOT NULL,
    total_score REAL,
    member_count INTEGER,
    throughput_score REAL,
    quality_score REAL,
    speed_score REAL,
    collaboration_score REAL,
    impact_score REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    UNIQUE(team_id, week_start)
);

INSERT INTO team_performance_scores_new
SELECT * FROM team_performance_scores;

DROP TABLE team_performance_scores;

ALTER TABLE team_performance_scores_new RENAME TO team_performance_scores;

CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);
CREATE INDEX IF NOT EXISTS idx_teams_manager_id ON teams(manager_id);
CREATE INDEX IF NOT EXISTS idx_teams_parent_team_id ON teams(parent_team_id);

CREATE INDEX IF NOT EXISTS idx_team_membership_team_id ON team_membership(team_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_member_id ON team_membership(member_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_team_member ON team_membership(team_id, member_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_left_at ON team_membership(left_at);

CREATE INDEX IF NOT EXISTS idx_view_configurations_name ON view_configurations(name);
CREATE INDEX IF NOT EXISTS idx_view_configurations_role ON view_configurations(role);
CREATE INDEX IF NOT EXISTS idx_view_configurations_is_default ON view_configurations(is_default);

CREATE INDEX IF NOT EXISTS idx_team_performance_scores_team ON team_performance_scores(team_id);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_week ON team_performance_scores(week_start);
			`,
		},
		{
			version: "013_weekly_briefings",
			sql: `
CREATE TABLE IF NOT EXISTS weekly_briefings (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    week_start DATETIME NOT NULL,
    tldr TEXT NOT NULL,
    key_metrics TEXT NOT NULL,
    needs_attention TEXT NOT NULL,
    insights TEXT NOT NULL,
    trending_up TEXT NOT NULL,
    trending_down TEXT NOT NULL,
    talking_points TEXT NOT NULL,
    generated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, week_start)
);

CREATE INDEX idx_briefings_team ON weekly_briefings(team_id);
CREATE INDEX idx_briefings_week ON weekly_briefings(week_start);
CREATE INDEX idx_briefings_generated ON weekly_briefings(generated_at);
			`,
		},
		{
			version: "014_add_burnout_risk",
			sql: `
ALTER TABLE performance_scores ADD COLUMN burnout_risk TEXT;
			`,
		},
		{
			version: "015_planning_features",
			sql: `
CREATE TABLE IF NOT EXISTS sprints (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    committed_points INTEGER,
    completed_points INTEGER,
    team_capacity INTEGER,
    status TEXT,
    team_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS epics (
    id TEXT PRIMARY KEY,
    jira_id TEXT UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    original_estimate_points INTEGER,
    current_scope_points INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    status TEXT,
    team_id TEXT,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS stories (
    id TEXT PRIMARY KEY,
    jira_id TEXT UNIQUE,
    epic_id TEXT,
    sprint_id TEXT,
    title TEXT NOT NULL,
    story_points INTEGER,
    status TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    actual_days_to_complete REAL,
    assignee_id TEXT,

    FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE SET NULL,
    FOREIGN KEY (sprint_id) REFERENCES sprints(id) ON DELETE SET NULL,
    FOREIGN KEY (assignee_id) REFERENCES engineers(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS capacity_history (
    id TEXT PRIMARY KEY,
    week_start DATE NOT NULL,
    team_id TEXT,
    team_size INTEGER,
    available_engineers REAL,
    completed_points INTEGER,
    meeting_hours REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL,
    UNIQUE(week_start, team_id)
);

CREATE TABLE IF NOT EXISTS planning_alerts (
    id TEXT PRIMARY KEY,
    alert_type TEXT,
    entity_type TEXT,
    entity_id TEXT,
    severity TEXT,
    message TEXT,
    ai_recommendation TEXT,
    dismissed BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sprints_team_id ON sprints(team_id);
CREATE INDEX IF NOT EXISTS idx_sprints_status ON sprints(status);
CREATE INDEX IF NOT EXISTS idx_sprints_start_date ON sprints(start_date);
CREATE INDEX IF NOT EXISTS idx_sprints_end_date ON sprints(end_date);
CREATE INDEX IF NOT EXISTS idx_sprints_team_status ON sprints(team_id, status);

CREATE INDEX IF NOT EXISTS idx_epics_team_id ON epics(team_id);
CREATE INDEX IF NOT EXISTS idx_epics_status ON epics(status);
CREATE INDEX IF NOT EXISTS idx_epics_jira_id ON epics(jira_id);
CREATE INDEX IF NOT EXISTS idx_epics_team_status ON epics(team_id, status);

CREATE INDEX IF NOT EXISTS idx_stories_epic_id ON stories(epic_id);
CREATE INDEX IF NOT EXISTS idx_stories_sprint_id ON stories(sprint_id);
CREATE INDEX IF NOT EXISTS idx_stories_assignee_id ON stories(assignee_id);
CREATE INDEX IF NOT EXISTS idx_stories_status ON stories(status);
CREATE INDEX IF NOT EXISTS idx_stories_jira_id ON stories(jira_id);
CREATE INDEX IF NOT EXISTS idx_stories_sprint_status ON stories(sprint_id, status);
CREATE INDEX IF NOT EXISTS idx_stories_assignee_status ON stories(assignee_id, status);

CREATE INDEX IF NOT EXISTS idx_capacity_history_team_id ON capacity_history(team_id);
CREATE INDEX IF NOT EXISTS idx_capacity_history_week_start ON capacity_history(week_start);
CREATE INDEX IF NOT EXISTS idx_capacity_history_team_week ON capacity_history(team_id, week_start);

CREATE INDEX IF NOT EXISTS idx_planning_alerts_entity ON planning_alerts(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_severity ON planning_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_dismissed ON planning_alerts(dismissed);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_alert_type ON planning_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_type_severity ON planning_alerts(alert_type, severity, dismissed);
			`,
		},
		{
			version: "016_add_settings",
			sql: `
CREATE TABLE IF NOT EXISTS settings (
    id TEXT PRIMARY KEY,
    sync_schedule TEXT NOT NULL DEFAULT 'manual',
    anonymization_strategy TEXT NOT NULL DEFAULT 'sequential',
    port INTEGER NOT NULL DEFAULT 3847,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

INSERT INTO settings (id, sync_schedule, anonymization_strategy, port, created_at, updated_at)
VALUES (
    'singleton',
    'manual',
    'sequential',
    3847,
    datetime('now'),
    datetime('now')
);
			`,
		},
		{
			version: "017_fix_settings_timestamps",
			sql: `
UPDATE settings
SET updated_at = datetime('now')
WHERE id = 'singleton';
			`,
		},
		{
			version: "018_alert_system",
			sql: `
CREATE TABLE IF NOT EXISTS alert_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    alert_type TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    threshold_value REAL,
    threshold_operator TEXT,
    severity TEXT NOT NULL,
    target_entity TEXT,
    target_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_channels (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    channel_type TEXT NOT NULL,
    channel_config TEXT,
    enabled BOOLEAN DEFAULT true,
    quiet_hours_start TEXT,
    quiet_hours_end TEXT,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alert_instances (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL,
    entity_type TEXT,
    entity_id TEXT,
    context TEXT,
    fired_at DATETIME NOT NULL,
    acknowledged_at DATETIME,
    acknowledged_by TEXT,
    snoozed_until DATETIME,
    dismissed_at DATETIME,
    dismissed_by TEXT,
    resolved_at DATETIME,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alert_deliveries (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    channel_type TEXT NOT NULL,
    delivered_at DATETIME NOT NULL,
    status TEXT NOT NULL,
    error_message TEXT,
    FOREIGN KEY (instance_id) REFERENCES alert_instances(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_target ON alert_rules(target_entity, target_id);
CREATE INDEX IF NOT EXISTS idx_alert_instances_rule ON alert_instances(rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_instances_fired ON alert_instances(fired_at);
CREATE INDEX IF NOT EXISTS idx_alert_instances_entity ON alert_instances(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_alert_deliveries_instance ON alert_deliveries(instance_id);
			`,
		},
		{
			version: "019_goal_system",
			sql: `
CREATE TABLE IF NOT EXISTS goals (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    goal_type TEXT NOT NULL,
    owner_type TEXT NOT NULL,
    owner_id TEXT,
    time_period TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    tracking_method TEXT NOT NULL,
    success_criteria TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    progress_percentage INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS goal_milestones (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    target_value REAL,
    current_value REAL DEFAULT 0,
    unit TEXT,
    completed BOOLEAN DEFAULT false,
    completed_at DATETIME,
    due_date DATETIME,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS goal_progress_logs (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    milestone_id TEXT,
    previous_value REAL,
    new_value REAL,
    change_type TEXT NOT NULL,
    evidence TEXT,
    logged_at DATETIME NOT NULL,
    logged_by TEXT,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (milestone_id) REFERENCES goal_milestones(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS goal_dependencies (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    depends_on_goal_id TEXT NOT NULL,
    dependency_type TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at DATETIME NOT NULL,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS goal_metrics (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    target_value REAL NOT NULL,
    current_value REAL,
    operator TEXT NOT NULL,
    last_evaluated DATETIME,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_goals_owner ON goals(owner_type, owner_id);
CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);
CREATE INDEX IF NOT EXISTS idx_goals_time_period ON goals(time_period);
CREATE INDEX IF NOT EXISTS idx_goal_milestones_goal ON goal_milestones(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_goal ON goal_progress_logs(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_metrics_goal ON goal_metrics(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_dependencies_goal ON goal_dependencies(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_dependencies_depends_on ON goal_dependencies(depends_on_goal_id);
			`,
		},
		{
			version: "020_skill_tracking",
			sql: `
CREATE TABLE IF NOT EXISTS skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL,
    subcategory TEXT,
    description TEXT,
    measurement_criteria TEXT,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS engineer_skills (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    level_score INTEGER NOT NULL DEFAULT 0,
    previous_level_score INTEGER,
    trajectory TEXT DEFAULT 'stable',
    last_evaluated DATETIME NOT NULL,
    evidence_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    UNIQUE(engineer_id, skill_id)
);

CREATE TABLE IF NOT EXISTS skill_evidence (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    evidence_type TEXT NOT NULL,
    evidence_source TEXT,
    strength REAL NOT NULL,
    context TEXT,
    detected_at DATETIME NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skill_progression (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    previous_score INTEGER NOT NULL,
    new_score INTEGER NOT NULL,
    change_reason TEXT,
    evaluated_at DATETIME NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skill_goals (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    goal_id TEXT,
    current_level INTEGER NOT NULL,
    target_level INTEGER NOT NULL,
    target_date DATETIME,
    milestones TEXT,
    status TEXT DEFAULT 'active',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_engineer_skills_engineer ON engineer_skills(engineer_id);
CREATE INDEX IF NOT EXISTS idx_engineer_skills_skill ON engineer_skills(skill_id);
CREATE INDEX IF NOT EXISTS idx_engineer_skills_trajectory ON engineer_skills(trajectory);
CREATE INDEX IF NOT EXISTS idx_skill_evidence_engineer ON skill_evidence(engineer_id, skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_evidence_detected ON skill_evidence(detected_at);
CREATE INDEX IF NOT EXISTS idx_skill_progression_engineer ON skill_progression(engineer_id);
CREATE INDEX IF NOT EXISTS idx_skill_progression_evaluated ON skill_progression(evaluated_at);
CREATE INDEX IF NOT EXISTS idx_skill_goals_engineer ON skill_goals(engineer_id);
CREATE INDEX IF NOT EXISTS idx_skill_goals_status ON skill_goals(status);
			`,
		},
		{
			version: "021_cost_roi_tracking",
			sql: `
CREATE TABLE IF NOT EXISTS cost_configuration (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    role TEXT,
    monthly_cost REAL NOT NULL,
    currency TEXT DEFAULT 'USD',
    effective_from DATETIME NOT NULL,
    effective_to DATETIME,
    notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS feature_values (
    id TEXT PRIMARY KEY,
    feature_name TEXT NOT NULL,
    feature_description TEXT,
    value_type TEXT NOT NULL,
    value_amount REAL,
    value_currency TEXT DEFAULT 'USD',
    confidence_level TEXT DEFAULT 'estimated',
    source TEXT,
    source_reference TEXT,
    time_period TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS feature_work_items (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    work_item_type TEXT NOT NULL,
    work_item_id TEXT NOT NULL,
    story_points INTEGER,
    actual_hours REAL,
    engineer_id TEXT,
    team_id TEXT,
    completed_at DATETIME,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS feature_costs (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    total_story_points INTEGER NOT NULL DEFAULT 0,
    total_hours REAL NOT NULL DEFAULT 0,
    total_cost REAL NOT NULL DEFAULT 0,
    cost_breakdown TEXT,
    computation_method TEXT NOT NULL,
    computed_at DATETIME NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS roi_calculations (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    investment REAL NOT NULL,
    return_value REAL,
    roi_percentage REAL,
    payback_months REAL,
    confidence TEXT DEFAULT 'low',
    notes TEXT,
    calculated_at DATETIME NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS engineering_investment (
    id TEXT PRIMARY KEY,
    time_period TEXT NOT NULL,
    team_id TEXT,
    category TEXT NOT NULL,
    story_points INTEGER DEFAULT 0,
    hours REAL DEFAULT 0,
    cost REAL DEFAULT 0,
    feature_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cost_efficiency_metrics (
    id TEXT PRIMARY KEY,
    time_period TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    currency TEXT DEFAULT 'USD',
    computed_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cost_config_entity ON cost_configuration(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_feature_values_period ON feature_values(time_period);
CREATE INDEX IF NOT EXISTS idx_feature_work_items_feature ON feature_work_items(feature_id);
CREATE INDEX IF NOT EXISTS idx_feature_work_items_engineer ON feature_work_items(engineer_id);
CREATE INDEX IF NOT EXISTS idx_feature_costs_feature ON feature_costs(feature_id);
CREATE INDEX IF NOT EXISTS idx_roi_calculations_feature ON roi_calculations(feature_id);
CREATE INDEX IF NOT EXISTS idx_engineering_investment_period ON engineering_investment(time_period, team_id);
CREATE INDEX IF NOT EXISTS idx_cost_efficiency_period ON cost_efficiency_metrics(time_period, entity_type, entity_id);
			`,
		},
		{
			version: "022_manager_context",
			sql: `
CREATE TABLE IF NOT EXISTS manager_notes (
    id TEXT PRIMARY KEY,
    manager_id TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    note_type TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    visibility TEXT DEFAULT 'private',
    tags TEXT,
    mood TEXT,
    action_items TEXT,
    linked_events TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (manager_id) REFERENCES engineers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS context_annotations (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    annotation_type TEXT NOT NULL,
    content TEXT NOT NULL,
    author_id TEXT NOT NULL,
    visibility TEXT DEFAULT 'team',
    created_at DATETIME NOT NULL,
    FOREIGN KEY (author_id) REFERENCES engineers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team_context (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    context_type TEXT NOT NULL,
    time_period TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    impact TEXT,
    author_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES engineers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sentiment_surveys (
    id TEXT PRIMARY KEY,
    survey_type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    questions TEXT NOT NULL,
    target_audience TEXT NOT NULL,
    anonymous BOOLEAN DEFAULT 1,
    active BOOLEAN DEFAULT 1,
    created_by TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    FOREIGN KEY (created_by) REFERENCES engineers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS survey_responses (
    id TEXT PRIMARY KEY,
    survey_id TEXT NOT NULL,
    respondent_id TEXT,
    responses TEXT NOT NULL,
    mood_rating INTEGER,
    text_feedback TEXT,
    submitted_at DATETIME NOT NULL,
    FOREIGN KEY (survey_id) REFERENCES sentiment_surveys(id) ON DELETE CASCADE,
    FOREIGN KEY (respondent_id) REFERENCES engineers(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS sentiment_analysis (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    time_period TEXT NOT NULL,
    source TEXT NOT NULL,
    sentiment_score REAL NOT NULL,
    confidence REAL NOT NULL,
    sample_size INTEGER,
    details TEXT,
    computed_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_manager_notes_manager ON manager_notes(manager_id);
CREATE INDEX IF NOT EXISTS idx_manager_notes_subject ON manager_notes(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_manager_notes_created ON manager_notes(created_at);
CREATE INDEX IF NOT EXISTS idx_context_annotations_entity ON context_annotations(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_team_context_team ON team_context(team_id, time_period);
CREATE INDEX IF NOT EXISTS idx_sentiment_surveys_active ON sentiment_surveys(active);
CREATE INDEX IF NOT EXISTS idx_survey_responses_survey ON survey_responses(survey_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_analysis_entity ON sentiment_analysis(entity_type, entity_id, time_period);
			`,
		},
		{
			version: "023_predictive_analytics",
			sql: `
CREATE TABLE IF NOT EXISTS forecasts (
    id TEXT PRIMARY KEY,
    forecast_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    time_horizon TEXT NOT NULL,
    predicted_value REAL NOT NULL,
    predicted_date DATETIME,
    confidence_percentage REAL NOT NULL,
    confidence_interval_low REAL,
    confidence_interval_high REAL,
    model_type TEXT NOT NULL,
    input_data TEXT,
    assumptions TEXT,
    created_at DATETIME NOT NULL,
    expires_at DATETIME
);

CREATE TABLE IF NOT EXISTS scenarios (
    id TEXT PRIMARY KEY,
    scenario_name TEXT NOT NULL,
    scenario_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    description TEXT,
    parameters TEXT NOT NULL,
    baseline_forecast_id TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (baseline_forecast_id) REFERENCES forecasts(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS scenario_results (
    id TEXT PRIMARY KEY,
    scenario_id TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    predicted_value REAL NOT NULL,
    difference_from_baseline REAL,
    impact TEXT,
    computed_at DATETIME NOT NULL,
    FOREIGN KEY (scenario_id) REFERENCES scenarios(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS risk_predictions (
    id TEXT PRIMARY KEY,
    risk_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    risk_score REAL NOT NULL,
    probability_percentage REAL NOT NULL,
    impact_severity TEXT NOT NULL,
    contributing_factors TEXT,
    mitigation_suggestions TEXT,
    predicted_at DATETIME NOT NULL,
    valid_until DATETIME
);

CREATE TABLE IF NOT EXISTS forecast_accuracy (
    id TEXT PRIMARY KEY,
    forecast_id TEXT NOT NULL,
    actual_value REAL NOT NULL,
    actual_date DATETIME NOT NULL,
    error_amount REAL NOT NULL,
    error_percentage REAL NOT NULL,
    was_within_confidence_interval BOOLEAN NOT NULL,
    measured_at DATETIME NOT NULL,
    FOREIGN KEY (forecast_id) REFERENCES forecasts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_forecasts_entity ON forecasts(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_forecasts_type ON forecasts(forecast_type);
CREATE INDEX IF NOT EXISTS idx_forecasts_expires ON forecasts(expires_at);
CREATE INDEX IF NOT EXISTS idx_scenarios_entity ON scenarios(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_scenario_results_scenario ON scenario_results(scenario_id);
CREATE INDEX IF NOT EXISTS idx_risk_predictions_entity ON risk_predictions(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_risk_predictions_score ON risk_predictions(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_forecast ON forecast_accuracy(forecast_id);
			`,
		},
		{
			version: "024_action_tracking",
			sql: `
CREATE TABLE IF NOT EXISTS recommendations (
    id TEXT PRIMARY KEY,
    source_type TEXT NOT NULL,
    source_id TEXT,
    recommendation_type TEXT NOT NULL,
    priority TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    suggested_actions TEXT,
    context TEXT,
    assigned_to TEXT,
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    status TEXT DEFAULT 'pending',
    FOREIGN KEY (assigned_to) REFERENCES engineers(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS actions (
    id TEXT PRIMARY KEY,
    recommendation_id TEXT,
    action_type TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    taken_by TEXT NOT NULL,
    taken_at DATETIME NOT NULL,
    evidence TEXT,
    expected_outcome TEXT,
    follow_up_date DATETIME,
    FOREIGN KEY (recommendation_id) REFERENCES recommendations(id) ON DELETE SET NULL,
    FOREIGN KEY (taken_by) REFERENCES engineers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS action_outcomes (
    id TEXT PRIMARY KEY,
    action_id TEXT NOT NULL,
    outcome_type TEXT NOT NULL,
    measured_metric TEXT,
    before_value REAL,
    after_value REAL,
    change_percentage REAL,
    time_to_impact_days INTEGER,
    effectiveness TEXT,
    notes TEXT,
    measured_at DATETIME NOT NULL,
    FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS recommendation_history (
    id TEXT PRIMARY KEY,
    recommendation_id TEXT NOT NULL,
    status_change TEXT NOT NULL,
    changed_by TEXT,
    reason TEXT,
    changed_at DATETIME NOT NULL,
    FOREIGN KEY (recommendation_id) REFERENCES recommendations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS follow_ups (
    id TEXT PRIMARY KEY,
    action_id TEXT NOT NULL,
    follow_up_date DATETIME NOT NULL,
    follow_up_type TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT false,
    completed_at DATETIME,
    FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_recommendations_assigned ON recommendations(assigned_to, status);
CREATE INDEX IF NOT EXISTS idx_recommendations_subject ON recommendations(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_status ON recommendations(status);
CREATE INDEX IF NOT EXISTS idx_recommendations_expires ON recommendations(expires_at);
CREATE INDEX IF NOT EXISTS idx_recommendations_source ON recommendations(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_actions_recommendation ON actions(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_actions_taken_by ON actions(taken_by);
CREATE INDEX IF NOT EXISTS idx_actions_subject ON actions(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_action_outcomes_action ON action_outcomes(action_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_history_rec ON recommendation_history(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_action ON follow_ups(action_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_date ON follow_ups(follow_up_date, completed);
			`,
		},
		{
			version: "025_dashboard_system",
			sql: `
-- Migration 025: Dashboard System
-- Dashboard Builder & Visualization System for v1.2.0 (PDR-8)

-- Dashboard definitions
CREATE TABLE IF NOT EXISTS dashboards (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    persona TEXT,              -- 'ic', 'team_lead', 'director' for templates
    is_template BOOLEAN DEFAULT FALSE,
    is_system BOOLEAN DEFAULT FALSE,  -- System dashboards can't be deleted
    layout TEXT NOT NULL,              -- JSON: Grid configuration with widgets
    filters TEXT,                      -- JSON: Global filters (date range, team, engineer)
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Widget definitions (embedded in dashboard layout JSON)
-- Each widget in layout JSON has:
-- {
--   id: string,
--   type: 'number' | 'timeseries' | 'bar' | 'table' | 'status' | 'feed',
--   title: string,
--   data_source: string,  // Which API endpoint to query
--   query_params: object, // Parameters for the query
--   visualization_config: object, // Colors, axes, formatting
--   position: {x: number, y: number, w: number, h: number}
-- }

-- Metric snapshots for fast queries (pre-computed hourly)
CREATE TABLE IF NOT EXISTS metric_snapshots (
    metric_name TEXT NOT NULL,        -- 'team_score', 'pr_volume', 'cycle_time'
    entity_type TEXT NOT NULL,        -- 'team', 'engineer', 'org'
    entity_id TEXT,                   -- team_id, engineer_id, null for org
    period_start TEXT NOT NULL,       -- ISO 8601 date
    period_end TEXT NOT NULL,         -- ISO 8601 date
    value REAL NOT NULL,
    metadata TEXT,                    -- JSON: Additional context
    computed_at TEXT NOT NULL,    -- ISO 8601 timestamp
    PRIMARY KEY (metric_name, entity_type, entity_id, period_start)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dashboards_persona ON dashboards(persona);
CREATE INDEX IF NOT EXISTS idx_dashboards_is_template ON dashboards(is_template);
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_lookup ON metric_snapshots(metric_name, entity_type, period_start);
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_entity ON metric_snapshots(entity_type, entity_id);
			`,
		},
		{
			version: "026_fix_dashboard_timestamp_types",
			sql: `
-- Migration 026: Fix timestamp column types in dashboards and metric_snapshots
-- Issue: Migration 025 used DATETIME instead of TEXT for timestamps (Rule 34 violation)
-- Fix: Recreate tables with TEXT columns for created_at, updated_at, computed_at

-- Fix dashboards table
CREATE TABLE IF NOT EXISTS dashboards_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    persona TEXT,
    is_template BOOLEAN DEFAULT FALSE,
    is_system BOOLEAN DEFAULT FALSE,
    layout TEXT NOT NULL,
    filters TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

INSERT INTO dashboards_new SELECT * FROM dashboards;
DROP TABLE dashboards;
ALTER TABLE dashboards_new RENAME TO dashboards;

CREATE INDEX IF NOT EXISTS idx_dashboards_persona ON dashboards(persona);
CREATE INDEX IF NOT EXISTS idx_dashboards_is_template ON dashboards(is_template);

-- Fix metric_snapshots table
CREATE TABLE IF NOT EXISTS metric_snapshots_new (
    metric_name TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    period_start TEXT NOT NULL,
    period_end TEXT NOT NULL,
    value REAL NOT NULL,
    metadata TEXT,
    computed_at TEXT NOT NULL,
    PRIMARY KEY (metric_name, entity_type, entity_id, period_start)
);

INSERT INTO metric_snapshots_new SELECT * FROM metric_snapshots;
DROP TABLE metric_snapshots;
ALTER TABLE metric_snapshots_new RENAME TO metric_snapshots;

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_lookup ON metric_snapshots(metric_name, entity_type, period_start);
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_entity ON metric_snapshots(entity_type, entity_id);
			`,
		},
		{
			version: "027_convert_dashboard_timestamp_format",
			sql: `
-- Migration 027: Convert dashboard timestamp data to RFC3339 format
-- Issue: Timestamps were stored in Go's default format "2006-01-02 15:04:05.999999 -0700 MST"
-- Fix: Convert to RFC3339 format "2006-01-02T15:04:05Z"
-- Rule 34: UTC timestamps everywhere in RFC3339 format

-- Convert dashboards timestamps
UPDATE dashboards
SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(created_at, 1, 19))
WHERE created_at NOT LIKE '%T%Z';

UPDATE dashboards
SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(updated_at, 1, 19))
WHERE updated_at NOT LIKE '%T%Z';

-- Convert metric_snapshots timestamps
UPDATE metric_snapshots
SET computed_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(computed_at, 1, 19))
WHERE computed_at NOT LIKE '%T%Z';
			`,
		},
	}

	// Apply each migration
	for _, migration := range migrations {
		// Check if already applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", migration.version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if exists {
			continue
		}

		// Execute migration and record it atomically in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.version, err)
		}

		// Execute migration
		if _, err := tx.Exec(migration.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.version, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.version, err)
		}

		fmt.Printf("Applied migration: %s\n", migration.version)
	}

	return nil
}
