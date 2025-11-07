-- Weekly briefings table for AI-generated team summaries

CREATE TABLE IF NOT EXISTS weekly_briefings (
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
    UNIQUE(team_id, week_start)
);

CREATE INDEX idx_briefings_team ON weekly_briefings(team_id);
CREATE INDEX idx_briefings_week ON weekly_briefings(week_start);
CREATE INDEX idx_briefings_generated ON weekly_briefings(generated_at);
