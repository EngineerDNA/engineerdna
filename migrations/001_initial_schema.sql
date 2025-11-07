-- Initial schema

CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,
    anonymized BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_id)
);

CREATE TABLE IF NOT EXISTS plugin_configs (
    name TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    config TEXT NOT NULL,
    last_sync DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    external_ids TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_actor ON events(actor);
CREATE INDEX idx_events_source ON events(source);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_anonymized ON events(anonymized);
