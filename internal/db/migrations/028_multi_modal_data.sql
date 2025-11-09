-- Migration 028: Multi-Modal Data Model
-- Extends EngineerDNA to support three data types: Events, Metrics, Attributes
-- Enables AWS cost tracking, team size tracking, cross-data correlations

-- Metric values table for time-series measurements
CREATE TABLE IF NOT EXISTS metric_values (
    id TEXT PRIMARY KEY,
    metric_name TEXT NOT NULL,
    source TEXT NOT NULL,             -- Plugin name that created this metric
    timestamp TEXT NOT NULL,          -- ISO 8601 timestamp
    granularity TEXT NOT NULL,        -- hourly, daily, weekly, monthly
    value REAL NOT NULL,
    unit TEXT,                        -- dollars, hours, count, percentage
    dimensions TEXT,                  -- JSON: Multi-dimensional metrics {service: ec2, region: us-east-1}
    created_at TEXT NOT NULL          -- ISO 8601 timestamp
);

-- Entity attributes table for facts about teams, engineers, org
CREATE TABLE IF NOT EXISTS entity_attributes (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,        -- team, engineer, org
    entity_id TEXT NOT NULL,
    attribute_name TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,         -- string, number, boolean, json
    valid_from TEXT NOT NULL,         -- ISO 8601 timestamp
    valid_until TEXT,                 -- ISO 8601 timestamp, NULL = still valid
    source TEXT NOT NULL,             -- Plugin name
    created_at TEXT NOT NULL,         -- ISO 8601 timestamp
    UNIQUE(entity_type, entity_id, attribute_name, valid_from)
);

-- Correlations table for cross-data-type relationships
CREATE TABLE IF NOT EXISTS correlations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    plugin TEXT NOT NULL,             -- Plugin that defines this correlation
    definition TEXT NOT NULL,         -- JSON: How to compute correlation
    created_at TEXT NOT NULL          -- ISO 8601 timestamp
);

-- Correlation values table for computed correlation results
CREATE TABLE IF NOT EXISTS correlation_values (
    id TEXT PRIMARY KEY,
    correlation_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,          -- ISO 8601 timestamp
    time_window TEXT NOT NULL,        -- day, week, month, quarter
    value REAL NOT NULL,
    breakdown TEXT,                   -- JSON: Breakdown by dimension
    created_at TEXT NOT NULL,         -- ISO 8601 timestamp
    FOREIGN KEY (correlation_id) REFERENCES correlations(id)
);

-- Plugin manifests table for plugin capabilities
CREATE TABLE IF NOT EXISTS plugin_manifests (
    plugin_name TEXT PRIMARY KEY,
    version TEXT NOT NULL,
    type TEXT NOT NULL,               -- source, destination, processor, metric_source, attribute_source
    capabilities TEXT,                -- JSON array
    provides_metrics TEXT,            -- JSON array of MetricSpec
    provides_event_types TEXT,        -- JSON array of EventTypeSpec
    provides_widgets TEXT,            -- JSON array of WidgetSpec
    provides_correlations TEXT,       -- JSON array of CorrelationSpec
    created_at TEXT NOT NULL,         -- ISO 8601 timestamp
    updated_at TEXT NOT NULL          -- ISO 8601 timestamp
);

-- Add normalized fields to events table for multi-source event support
ALTER TABLE events ADD COLUMN normalized_type TEXT;
ALTER TABLE events ADD COLUMN normalized_data TEXT;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_metric_values_name_timestamp ON metric_values(metric_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_metric_values_source ON metric_values(source);
CREATE INDEX IF NOT EXISTS idx_entity_attributes_lookup ON entity_attributes(entity_type, entity_id, attribute_name);
CREATE INDEX IF NOT EXISTS idx_entity_attributes_type ON entity_attributes(entity_type, attribute_name);
CREATE INDEX IF NOT EXISTS idx_entity_attributes_valid ON entity_attributes(entity_type, entity_id, valid_until);
CREATE INDEX IF NOT EXISTS idx_correlation_values_lookup ON correlation_values(correlation_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_normalized_type ON events(normalized_type);
