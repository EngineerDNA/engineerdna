-- Migration: Add settings table for user-configurable application settings
-- Created: 2025-11-06
-- Purpose: Store user preferences (sync schedule, anonymization strategy, port) with singleton pattern

CREATE TABLE IF NOT EXISTS settings (
    id TEXT PRIMARY KEY,
    sync_schedule TEXT NOT NULL DEFAULT 'manual',
    anonymization_strategy TEXT NOT NULL DEFAULT 'sequential',
    port INTEGER NOT NULL DEFAULT 3847,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Insert default settings row (singleton pattern with id='singleton')
INSERT INTO settings (id, sync_schedule, anonymization_strategy, port, created_at, updated_at)
VALUES (
    'singleton',
    'manual',
    'sequential',
    3847,
    datetime('now'),
    datetime('now')
);
