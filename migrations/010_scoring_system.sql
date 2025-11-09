-- Scoring system tables

-- Roles and expectations
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    target_score INTEGER DEFAULT 100,
    expectations TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Scoring weights (global configuration)
CREATE TABLE IF NOT EXISTS scoring_weights (
    id TEXT PRIMARY KEY,
    throughput_weight REAL DEFAULT 30,
    quality_weight REAL DEFAULT 25,
    speed_weight REAL DEFAULT 20,
    collaboration_weight REAL DEFAULT 15,
    impact_weight REAL DEFAULT 10,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Individual performance scores (weekly)
CREATE TABLE IF NOT EXISTS performance_scores (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    week_start DATE NOT NULL,
    total_score REAL,
    throughput_score REAL,
    quality_score REAL,
    speed_score REAL,
    collaboration_score REAL,
    impact_score REAL,
    raw_metrics TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (engineer_id) REFERENCES engineers(id),
    UNIQUE(engineer_id, week_start)
);

-- Team performance scores (weekly aggregated)
CREATE TABLE IF NOT EXISTS team_performance_scores (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    week_start DATE NOT NULL,
    total_score REAL,
    member_count INTEGER,
    throughput_score REAL,
    quality_score REAL,
    speed_score REAL,
    collaboration_score REAL,
    impact_score REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(team_id, week_start)
);

-- Promotion signals
CREATE TABLE IF NOT EXISTS promotion_signals (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    signal_type TEXT,
    detected_at DATETIME,
    weeks_duration INTEGER,
    avg_score REAL,
    dismissed BOOLEAN DEFAULT FALSE,
    dismissed_at DATETIME,
    notes TEXT,

    FOREIGN KEY (engineer_id) REFERENCES engineers(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_performance_scores_engineer ON performance_scores(engineer_id);
CREATE INDEX IF NOT EXISTS idx_performance_scores_week ON performance_scores(week_start);
CREATE INDEX IF NOT EXISTS idx_performance_scores_engineer_week ON performance_scores(engineer_id, week_start DESC);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_team ON team_performance_scores(team_id);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_week ON team_performance_scores(week_start);
CREATE INDEX IF NOT EXISTS idx_promotion_signals_engineer ON promotion_signals(engineer_id);
CREATE INDEX IF NOT EXISTS idx_promotion_signals_dismissed ON promotion_signals(dismissed);
