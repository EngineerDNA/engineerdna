-- Migration 009: Performance Indexes for Query Optimization
-- Add composite indexes for common query patterns to improve dashboard load times

-- Events by engineer and time range (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_events_engineer_timestamp
ON events(engineer_id, timestamp DESC);

-- Events by type and time range (for filtering by event type)
CREATE INDEX IF NOT EXISTS idx_events_type_timestamp
ON events(type, timestamp DESC);

-- Events by source and time range (for filtering by data source)
CREATE INDEX IF NOT EXISTS idx_events_source_timestamp
ON events(source, timestamp DESC);

-- Composite index for complex filtered queries
CREATE INDEX IF NOT EXISTS idx_events_composite
ON events(engineer_id, type, source, timestamp DESC);

-- Unresolved identities lookup for identity management
CREATE INDEX IF NOT EXISTS idx_unresolved_source_identifier
ON unresolved_identities(source, identifier);

-- Analyze tables to update query planner statistics
ANALYZE;
