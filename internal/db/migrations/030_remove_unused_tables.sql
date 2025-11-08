-- Migration 030: Remove Unused Tables
-- Removes 8 tables that are either unused or have better replacements
-- Data is preserved where needed by migrating to appropriate tables

-- 1. Migrate exports data to audit_log before dropping
-- exports table duplicates audit_log functionality
INSERT OR IGNORE INTO audit_log (
    id,
    timestamp,
    plugin_name,
    action,
    event_count,
    anonymized,
    destination,
    user_initiated,
    data_summary
)
SELECT
    id,
    exported_at,
    plugin_name,
    'export',
    event_count,
    anonymized,
    destination_url,
    1,
    json_object(
        'status', status,
        'error_message', error_message
    )
FROM exports
WHERE id NOT IN (SELECT id FROM audit_log);

-- 2. Drop unused tables

-- view_configurations: Dashboard views (replaced by dashboards table in PDR-8)
DROP TABLE IF EXISTS view_configurations;

-- anonymization_policies: Per-plugin settings (should be in config file, not DB)
DROP TABLE IF EXISTS anonymization_policies;

-- capacity_history: Use metric_values instead for capacity tracking
DROP TABLE IF EXISTS capacity_history;

-- cost_efficiency_metrics: Use metric_values instead for cost metrics
DROP TABLE IF EXISTS cost_efficiency_metrics;

-- planning_alerts: Merge into alert_instances from PDR-7
DROP TABLE IF EXISTS planning_alerts;

-- team_context: Merge into manager_notes from PDR-7
DROP TABLE IF EXISTS team_context;

-- recommendation_history: audit_log tracks this
DROP TABLE IF EXISTS recommendation_history;

-- exports: Now migrated to audit_log
DROP TABLE IF EXISTS exports;
