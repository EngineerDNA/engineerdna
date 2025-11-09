-- Migration 019: Goal Tracking System
-- OKR and Goal Management with automatic progress detection

-- Goals: Individual, team, and org-wide goals
CREATE TABLE IF NOT EXISTS goals (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    goal_type TEXT NOT NULL, -- 'individual', 'team', 'org'
    owner_type TEXT NOT NULL, -- 'engineer', 'team', 'org'
    owner_id TEXT, -- engineer_id, team_id, or null for org
    time_period TEXT NOT NULL, -- 'Q1 2025', 'Q4 2024', etc.
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    tracking_method TEXT NOT NULL, -- 'manual', 'automatic', 'hybrid'
    success_criteria TEXT, -- JSON array of criteria
    status TEXT NOT NULL DEFAULT 'active', -- 'active', 'completed', 'at_risk', 'off_track', 'archived'
    progress_percentage INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Goal Milestones: Breakdown of goals into milestones
CREATE TABLE IF NOT EXISTS goal_milestones (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    target_value REAL, -- e.g., 2 for "Lead 2 complex features"
    current_value REAL DEFAULT 0,
    unit TEXT, -- 'features', 'PRs', 'bugs', etc.
    completed BOOLEAN DEFAULT false,
    completed_at TEXT,
    due_date TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

-- Goal Progress Logs: Track progress updates
CREATE TABLE IF NOT EXISTS goal_progress_logs (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    milestone_id TEXT,
    previous_value REAL,
    new_value REAL,
    change_type TEXT NOT NULL, -- 'manual_update', 'auto_detected', 'milestone_complete'
    evidence TEXT, -- JSON: reference to PR, event, or manual note
    logged_at TEXT NOT NULL,
    logged_by TEXT, -- engineer_id or 'system'
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (milestone_id) REFERENCES goal_milestones(id) ON DELETE SET NULL
);

-- Goal Dependencies: Track blockers
CREATE TABLE IF NOT EXISTS goal_dependencies (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    depends_on_goal_id TEXT NOT NULL,
    dependency_type TEXT NOT NULL, -- 'blocks', 'relates_to'
    status TEXT DEFAULT 'active', -- 'active', 'resolved'
    created_at TEXT NOT NULL,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

-- Goal Metrics: Link goals to quantitative metrics
CREATE TABLE IF NOT EXISTS goal_metrics (
    id TEXT PRIMARY KEY,
    goal_id TEXT NOT NULL,
    metric_name TEXT NOT NULL, -- 'cycle_time', 'pr_count', 'bug_rate', etc.
    target_value REAL NOT NULL,
    current_value REAL,
    operator TEXT NOT NULL, -- 'decrease_to', 'increase_to', 'maintain'
    last_evaluated TEXT,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_goals_owner ON goals(owner_type, owner_id);
CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);
CREATE INDEX IF NOT EXISTS idx_goals_time_period ON goals(time_period);
CREATE INDEX IF NOT EXISTS idx_goal_milestones_goal ON goal_milestones(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_goal ON goal_progress_logs(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_metrics_goal ON goal_metrics(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_dependencies_goal ON goal_dependencies(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_dependencies_depends_on ON goal_dependencies(depends_on_goal_id);
