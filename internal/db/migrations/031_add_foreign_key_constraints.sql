-- Migration 031: Add Missing Foreign Key Constraints
-- Adds foreign key constraints to events.engineer_id and weekly_briefings.team_id
-- SQLite requires table recreation to add foreign keys to existing tables

-- 1. Add foreign key to events.engineer_id
-- Create new table with FK constraint
CREATE TABLE events_new (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,
    normalized_type TEXT,
    normalized_data TEXT,
    anonymized BOOLEAN NOT NULL DEFAULT 0,
    engineer_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_id),
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Copy data from old table
INSERT INTO events_new
SELECT id, type, source, source_id, timestamp, actor, data,
       normalized_type, normalized_data, anonymized, engineer_id,
       created_at, updated_at
FROM events;

-- Drop old table and rename new table
DROP TABLE events;
ALTER TABLE events_new RENAME TO events;

-- Recreate all indexes
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_actor ON events(actor);
CREATE INDEX idx_events_source ON events(source);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_anonymized ON events(anonymized);
CREATE INDEX idx_events_engineer_id ON events(engineer_id);
CREATE INDEX idx_events_normalized_type ON events(normalized_type);

-- 2. Add foreign key to weekly_briefings.team_id
-- Create new table with FK constraint
CREATE TABLE weekly_briefings_new (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    week_start DATETIME NOT NULL,
    tldr TEXT NOT NULL,
    key_metrics TEXT NOT NULL,
    needs_attention TEXT NOT NULL,
    insights TEXT NOT NULL,
    trending_up TEXT NOT NULL,
    trending_down TEXT NOT NULL,
    talking_points TEXT NOT NULL,
    generated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, week_start),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

-- Copy data from old table
INSERT INTO weekly_briefings_new
SELECT id, team_id, week_start, tldr, key_metrics, needs_attention,
       insights, trending_up, trending_down, talking_points,
       generated_at, created_at
FROM weekly_briefings;

-- Drop old table and rename new table
DROP TABLE weekly_briefings;
ALTER TABLE weekly_briefings_new RENAME TO weekly_briefings;

-- Recreate all indexes
CREATE INDEX idx_briefings_team ON weekly_briefings(team_id);
CREATE INDEX idx_briefings_week ON weekly_briefings(week_start);
CREATE INDEX idx_briefings_generated ON weekly_briefings(generated_at);
