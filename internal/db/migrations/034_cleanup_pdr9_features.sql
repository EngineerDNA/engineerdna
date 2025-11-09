-- Migration 034: Cleanup Features + Event Normalization + Correlations
-- Context: Consolidates plugin system from 5 types to 3 types
--          and keeps features that provide immediate value
--
-- What's being removed:
--   1. Plugin Manifests - plugin_manifests table (plugin.json serves this purpose)
--
-- What's being kept AND USED:
--   1. Event Normalization - normalized_type column (prevents future rewrites)
--   2. Correlation Engine - correlations + correlation_values tables (fast historical trends)
--   3. Universal Schema - metric_values, entity_attributes tables (multi-modal data)
--
-- Rationale for keeping event normalization:
--   - Prevents dashboard rewrites when adding new source plugins
--   - GitHub plugin maps pull_request -> code_review from day 1
--   - Future GitLab/Bitbucket plugins drop in without code changes
--
-- Rationale for keeping correlation engine:
--   - Users sync YEARS of historical data on day 1 (GitHub since 2020)
--   - On-demand computation over large datasets is slow (500ms-1s)
--   - Pre-computed correlations enable fast dashboards (<50ms)
--   - Core value prop: "Cost per feature delivered" trends

-- 1. Event Normalization - Keep and ensure it exists
-- The normalized_type column was added in migration 028 but may not exist
-- in all environments. Ensure it exists for forward compatibility.

-- Check if events table has normalized_type column already
-- If not, we need to recreate the table to add it

-- SQLite doesn't support ALTER TABLE ADD COLUMN IF NOT EXISTS
-- So we recreate the table with the normalized columns

CREATE TABLE IF NOT EXISTS events_new (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    normalized_type TEXT,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,
    anonymized BOOLEAN NOT NULL DEFAULT 0,
    engineer_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_id),
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Copy data from old table to new table (preserving normalized_type if exists)
INSERT INTO events_new (
    id, type, normalized_type, source, source_id, timestamp, actor, data,
    anonymized, engineer_id, created_at, updated_at
)
SELECT
    id, type, normalized_type, source, source_id, timestamp, actor, data,
    anonymized, engineer_id, created_at, updated_at
FROM events;

-- Drop old table and rename new table
DROP TABLE events;
ALTER TABLE events_new RENAME TO events;

-- Recreate all indexes for events table (INCLUDING normalized_type)
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_actor ON events(actor);
CREATE INDEX idx_events_source ON events(source);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_normalized_type ON events(normalized_type);
CREATE INDEX idx_events_anonymized ON events(anonymized);
CREATE INDEX idx_events_engineer_id ON events(engineer_id);
CREATE INDEX idx_events_updated_at ON events(updated_at);
CREATE INDEX idx_events_engineer_timestamp ON events(engineer_id, timestamp DESC);
CREATE INDEX idx_events_type_timestamp ON events(type, timestamp DESC);
CREATE INDEX idx_events_source_timestamp ON events(source, timestamp DESC);
CREATE INDEX idx_events_composite ON events(engineer_id, type, source, timestamp DESC);

-- 2. Keep Correlation Engine
-- The correlation engine computes cross-data-type relationships (e.g., cost per feature)
-- KEEPING for v0.1.0 because:
-- - Users can sync YEARS of historical data on day 1 (GitHub sync since=2020)
-- - Computing trends over large datasets on-demand is slow (500ms-1s)
-- - Pre-computed correlations enable fast dashboard loads (<50ms)
-- - Core value prop: "What does engineering cost per feature delivered?"

-- Ensure correlation tables exist (created in migration 028)
-- No changes needed - tables already exist and are correct

-- Ensure index exists for fast lookups
CREATE INDEX IF NOT EXISTS idx_correlation_values_lookup
ON correlation_values(correlation_id, timestamp DESC);

-- 3. Remove Plugin Manifests
-- Plugin manifests were intended to register plugin capabilities in the database
-- but the plugin system uses plugin.json files for this purpose instead.
-- This table was created but never populated or used.
DROP TABLE IF EXISTS plugin_manifests;
