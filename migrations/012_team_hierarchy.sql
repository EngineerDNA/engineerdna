-- Team hierarchy tables

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_team_id TEXT,
    manager_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_team_id) REFERENCES teams(id) ON DELETE SET NULL,
    FOREIGN KEY (manager_id) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Team membership (many-to-many with history tracking)
CREATE TABLE IF NOT EXISTS team_membership (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    role TEXT,
    joined_at DATETIME NOT NULL,
    left_at DATETIME,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES engineers(id) ON DELETE CASCADE,
    UNIQUE(team_id, member_id, joined_at)
);

-- View configurations (saved dashboard views per user role)
CREATE TABLE IF NOT EXISTS view_configurations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT,
    default_mode TEXT,
    team_filter TEXT,
    alert_filter TEXT,
    sort_order TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key constraint to existing team_performance_scores table
-- SQLite requires recreating table to add FK constraint
CREATE TABLE IF NOT EXISTS team_performance_scores_new (
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

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    UNIQUE(team_id, week_start)
);

-- Copy existing data if any
INSERT INTO team_performance_scores_new
SELECT * FROM team_performance_scores;

-- Drop old table
DROP TABLE team_performance_scores;

-- Rename new table
ALTER TABLE team_performance_scores_new RENAME TO team_performance_scores;

-- Indexes for teams table
CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);
CREATE INDEX IF NOT EXISTS idx_teams_manager_id ON teams(manager_id);
CREATE INDEX IF NOT EXISTS idx_teams_parent_team_id ON teams(parent_team_id);

-- Indexes for team_membership table
CREATE INDEX IF NOT EXISTS idx_team_membership_team_id ON team_membership(team_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_member_id ON team_membership(member_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_team_member ON team_membership(team_id, member_id);
CREATE INDEX IF NOT EXISTS idx_team_membership_left_at ON team_membership(left_at);

-- Indexes for view_configurations table
CREATE INDEX IF NOT EXISTS idx_view_configurations_name ON view_configurations(name);
CREATE INDEX IF NOT EXISTS idx_view_configurations_role ON view_configurations(role);
CREATE INDEX IF NOT EXISTS idx_view_configurations_is_default ON view_configurations(is_default);

-- Recreate indexes for team_performance_scores (lost during table recreation)
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_team ON team_performance_scores(team_id);
CREATE INDEX IF NOT EXISTS idx_team_performance_scores_week ON team_performance_scores(week_start);
