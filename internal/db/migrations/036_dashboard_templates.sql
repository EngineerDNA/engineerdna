-- Migration 036: Dashboard Templates System
-- Dashboard-First UI/UX Simplification
-- Created: 2025-11-08

-- Dashboard templates table
CREATE TABLE IF NOT EXISTS dashboard_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    role TEXT CHECK (role IN ('ic', 'manager', 'director', 'admin') OR role IS NULL),
    category TEXT, -- 'personal', 'team', 'org', 'admin'
    layout TEXT NOT NULL, -- JSON
    is_system INTEGER DEFAULT 0, -- SQLite boolean (0/1)
    created_at TEXT NOT NULL, -- ISO 8601 timestamp
    updated_at TEXT NOT NULL  -- ISO 8601 timestamp
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dashboard_templates_role ON dashboard_templates(role);
CREATE INDEX IF NOT EXISTS idx_dashboard_templates_category ON dashboard_templates(category);
CREATE INDEX IF NOT EXISTS idx_dashboard_templates_system ON dashboard_templates(is_system);
