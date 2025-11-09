-- Migration 035: Add Role and Dashboard Preferences to Settings
-- Dashboard-First UI/UX Simplification
-- Created: 2025-11-08

-- Add role and dashboard preferences to settings table
ALTER TABLE settings ADD COLUMN role TEXT DEFAULT 'manager'
    CHECK (role IN ('ic', 'manager', 'director', 'admin'));

ALTER TABLE settings ADD COLUMN primary_dashboard_id TEXT;

ALTER TABLE settings ADD COLUMN favorite_dashboards TEXT; -- JSON array of dashboard IDs

ALTER TABLE settings ADD COLUMN onboarding_completed INTEGER DEFAULT 0; -- SQLite boolean (0/1)

-- Create index for role lookups
CREATE INDEX IF NOT EXISTS idx_settings_role ON settings(role);
