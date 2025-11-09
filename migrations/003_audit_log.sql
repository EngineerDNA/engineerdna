-- Audit log and insights

CREATE TABLE IF NOT EXISTS audit_log (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    plugin_name TEXT NOT NULL,
    action TEXT NOT NULL,
    event_count INTEGER NOT NULL,
    anonymized BOOLEAN NOT NULL,
    destination TEXT,
    data_summary TEXT,
    user_initiated BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS insights (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    generated_at DATETIME NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    severity TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    recommendation TEXT,
    metrics TEXT NOT NULL,
    dismissed BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exports (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    exported_at DATETIME NOT NULL,
    status TEXT NOT NULL,
    destination_url TEXT,
    event_count INTEGER NOT NULL,
    anonymized BOOLEAN NOT NULL,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_plugin ON audit_log(plugin_name);
CREATE INDEX idx_insights_generated ON insights(generated_at);
CREATE INDEX idx_exports_exported ON exports(exported_at);
