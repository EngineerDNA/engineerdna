-- Identity management system

CREATE TABLE IF NOT EXISTS engineers (
    id TEXT PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    email TEXT,
    manager TEXT,
    identifiers TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS unresolved_identities (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    identifier TEXT NOT NULL,
    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_count INTEGER NOT NULL DEFAULT 1,
    UNIQUE(source, identifier)
);

ALTER TABLE events ADD COLUMN engineer_id TEXT;

CREATE INDEX idx_engineers_name ON engineers(canonical_name);
CREATE INDEX idx_engineers_email ON engineers(email);
CREATE INDEX idx_engineers_active ON engineers(active);
CREATE INDEX idx_unresolved_source ON unresolved_identities(source);
CREATE INDEX idx_unresolved_identifier ON unresolved_identities(identifier);
CREATE INDEX idx_events_engineer_id ON events(engineer_id);
