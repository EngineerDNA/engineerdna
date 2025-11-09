-- Migration 026: Fix timestamp column types in dashboards and metric_snapshots
-- Issue: Migration 025 used DATETIME instead of TEXT for timestamps (Rule 34 violation)
-- Fix: Recreate tables with TEXT columns for created_at, updated_at, computed_at

-- Fix dashboards table
CREATE TABLE IF NOT EXISTS dashboards_new (
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

INSERT INTO dashboards_new SELECT * FROM dashboards;
DROP TABLE dashboards;
ALTER TABLE dashboards_new RENAME TO dashboards;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_dashboards_persona ON dashboards(persona);
CREATE INDEX IF NOT EXISTS idx_dashboards_is_template ON dashboards(is_template);

-- Fix metric_snapshots table
CREATE TABLE IF NOT EXISTS metric_snapshots_new (
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

INSERT INTO metric_snapshots_new SELECT * FROM metric_snapshots;
DROP TABLE metric_snapshots;
ALTER TABLE metric_snapshots_new RENAME TO metric_snapshots;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_lookup ON metric_snapshots(metric_name, entity_type, period_start);
