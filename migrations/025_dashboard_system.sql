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
    created_at TEXT NOT NULL,          -- ISO 8601 timestamp (stored as TEXT per migrations.go)
    updated_at TEXT NOT NULL           -- ISO 8601 timestamp (stored as TEXT per migrations.go)
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
    period_start TEXT NOT NULL,       -- ISO 8601 date (stored as TEXT per migrations.go)
    period_end TEXT NOT NULL,         -- ISO 8601 date (stored as TEXT per migrations.go)
    value REAL NOT NULL,
    metadata TEXT,                    -- JSON: Additional context
    computed_at TEXT NOT NULL,        -- ISO 8601 timestamp (stored as TEXT per migrations.go)
    PRIMARY KEY (metric_name, entity_type, entity_id, period_start)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dashboards_persona ON dashboards(persona);
CREATE INDEX IF NOT EXISTS idx_dashboards_is_template ON dashboards(is_template);
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_lookup ON metric_snapshots(metric_name, entity_type, period_start);
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_entity ON metric_snapshots(entity_type, entity_id);
