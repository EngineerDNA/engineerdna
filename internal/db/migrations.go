package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

// Embed all migration SQL files
//
//go:embed migrations/001_initial_schema.sql
var migration001 string

//go:embed migrations/002_anonymization.sql
var migration002 string

//go:embed migrations/003_audit_log.sql
var migration003 string

//go:embed migrations/004_identity_system.sql
var migration004 string

//go:embed migrations/005_add_ignored_flag.sql
var migration005 string

//go:embed migrations/006_add_events_updated_at_index.sql
var migration006 string

//go:embed migrations/007_add_app_config.sql
var migration007 string

//go:embed migrations/008_add_export_schedules.sql
var migration008 string

//go:embed migrations/009_performance_indexes.sql
var migration009 string

//go:embed migrations/010_scoring_system.sql
var migration010 string

//go:embed migrations/011_add_engineer_role_id.sql
var migration011 string

//go:embed migrations/012_team_hierarchy.sql
var migration012 string

//go:embed migrations/013_weekly_briefings.sql
var migration013 string

//go:embed migrations/014_add_burnout_risk.sql
var migration014 string

//go:embed migrations/015_planning_features.sql
var migration015 string

//go:embed migrations/016_add_settings.sql
var migration016 string

//go:embed migrations/017_fix_settings_timestamps.sql
var migration017 string

//go:embed migrations/018_alert_system.sql
var migration018 string

//go:embed migrations/019_goal_system.sql
var migration019 string

//go:embed migrations/020_skill_tracking.sql
var migration020 string

//go:embed migrations/021_cost_roi_tracking.sql
var migration021 string

//go:embed migrations/022_manager_context.sql
var migration022 string

//go:embed migrations/023_predictive_analytics.sql
var migration023 string

//go:embed migrations/024_action_tracking.sql
var migration024 string

//go:embed migrations/025_dashboard_system.sql
var migration025 string

//go:embed migrations/026_fix_dashboard_timestamp_types.sql
var migration026 string

//go:embed migrations/027_convert_dashboard_timestamp_format.sql
var migration027 string

//go:embed migrations/028_multi_modal_data.sql
var migration028 string

//go:embed migrations/029_remove_duplicate_users_table.sql
var migration029 string

//go:embed migrations/030_remove_unused_tables.sql
var migration030 string

//go:embed migrations/031_add_foreign_key_constraints.sql
var migration031 string

//go:embed migrations/032_migrate_performance_scores.sql
var migration032 string

//go:embed migrations/033_migrate_cost_configuration.sql
var migration033 string

//go:embed migrations/034_cleanup_pdr9_features.sql
var migration034 string

//go:embed migrations/035_add_role_and_dashboard_prefs.sql
var migration035 string

//go:embed migrations/036_dashboard_templates.sql
var migration036 string

//go:embed migrations/037_seed_dashboard_templates.sql
var migration037 string

//go:embed migrations/038_fix_events_normalized_data.sql
var migration038 string

// RunMigrations applies all pending database migrations
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

	// Define migrations mapping version to embedded SQL
	migrations := []struct {
		version string
		sql     string
	}{
		{"001_initial_schema", migration001},
		{"002_anonymization", migration002},
		{"003_audit_log", migration003},
		{"004_identity_system", migration004},
		{"005_add_ignored_flag", migration005},
		{"006_add_events_updated_at_index", migration006},
		{"007_add_app_config", migration007},
		{"008_add_export_schedules", migration008},
		{"009_performance_indexes", migration009},
		{"010_scoring_system", migration010},
		{"011_add_engineer_role_id", migration011},
		{"012_team_hierarchy", migration012},
		{"013_weekly_briefings", migration013},
		{"014_add_burnout_risk", migration014},
		{"015_planning_features", migration015},
		{"016_add_settings", migration016},
		{"017_fix_settings_timestamps", migration017},
		{"018_alert_system", migration018},
		{"019_goal_system", migration019},
		{"020_skill_tracking", migration020},
		{"021_cost_roi_tracking", migration021},
		{"022_manager_context", migration022},
		{"023_predictive_analytics", migration023},
		{"024_action_tracking", migration024},
		{"025_dashboard_system", migration025},
		{"026_fix_dashboard_timestamp_types", migration026},
		{"027_convert_dashboard_timestamp_format", migration027},
		{"028_multi_modal_data", migration028},
		{"029_remove_duplicate_users_table", migration029},
		{"030_remove_unused_tables", migration030},
		{"031_add_foreign_key_constraints", migration031},
		{"032_migrate_performance_scores", migration032},
		{"033_migrate_cost_configuration", migration033},
		{"034_cleanup_pdr9_features", migration034},
		{"035_add_role_and_dashboard_prefs", migration035},
		{"036_dashboard_templates", migration036},
		{"037_seed_dashboard_templates", migration037},
		{"038_fix_events_normalized_data", migration038},
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

		// Clean up SQL (remove comments and normalize whitespace)
		cleanSQL := cleanMigrationSQL(migration.sql)

		// Execute migration
		if _, err := tx.Exec(cleanSQL); err != nil {
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

// cleanMigrationSQL removes comments and normalizes whitespace
func cleanMigrationSQL(sql string) string {
	lines := strings.Split(sql, "\n")
	var cleaned []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment-only lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		// Remove inline comments
		if idx := strings.Index(trimmed, "--"); idx > 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}
