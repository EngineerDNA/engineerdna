-- Migration: Add index on events.updated_at for performance
-- Created: 2025-11-04
-- Reason: Improve query performance for event updates

CREATE INDEX IF NOT EXISTS idx_events_updated_at ON events(updated_at);
