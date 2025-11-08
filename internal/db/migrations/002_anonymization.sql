-- Anonymization tables

CREATE TABLE IF NOT EXISTS anonymization_map (
    id TEXT PRIMARY KEY,
    anonymized_id TEXT UNIQUE NOT NULL,
    real_name TEXT NOT NULL,
    real_email TEXT,
    external_identifiers TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS anonymization_policies (
    plugin_name TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT 0,
    strategy TEXT NOT NULL,
    fields TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_anon_map_real ON anonymization_map(real_name);
CREATE INDEX idx_anon_map_anon ON anonymization_map(anonymized_id);
