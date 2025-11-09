-- Export scheduling

CREATE TABLE IF NOT EXISTS export_schedules (
    id TEXT PRIMARY KEY,
    plugin_name TEXT NOT NULL,
    frequency TEXT NOT NULL,
    day_of_week INTEGER,
    time_of_day TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    last_run DATETIME,
    next_run DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_export_schedules_next_run ON export_schedules(enabled, next_run);
CREATE INDEX idx_export_schedules_plugin ON export_schedules(plugin_name);
