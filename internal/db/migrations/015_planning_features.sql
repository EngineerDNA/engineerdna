-- Planning features tables

-- Sprints table for sprint planning and tracking
CREATE TABLE IF NOT EXISTS sprints (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    committed_points INTEGER,
    completed_points INTEGER,
    team_capacity INTEGER,
    status TEXT,
    team_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);

-- Epics table for tracking large initiatives
CREATE TABLE IF NOT EXISTS epics (
    id TEXT PRIMARY KEY,
    jira_id TEXT UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    original_estimate_points INTEGER,
    current_scope_points INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    status TEXT,
    team_id TEXT,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);

-- Stories table for user stories and tasks
CREATE TABLE IF NOT EXISTS stories (
    id TEXT PRIMARY KEY,
    jira_id TEXT UNIQUE,
    epic_id TEXT,
    sprint_id TEXT,
    title TEXT NOT NULL,
    story_points INTEGER,
    status TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    actual_days_to_complete REAL,
    assignee_id TEXT,

    FOREIGN KEY (epic_id) REFERENCES epics(id) ON DELETE SET NULL,
    FOREIGN KEY (sprint_id) REFERENCES sprints(id) ON DELETE SET NULL,
    FOREIGN KEY (assignee_id) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Capacity history for tracking team capacity and delivery over time
CREATE TABLE IF NOT EXISTS capacity_history (
    id TEXT PRIMARY KEY,
    week_start DATE NOT NULL,
    team_id TEXT,
    team_size INTEGER,
    available_engineers REAL,
    completed_points INTEGER,
    meeting_hours REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL,
    UNIQUE(week_start, team_id)
);

-- Planning alerts for detecting planning issues
CREATE TABLE IF NOT EXISTS planning_alerts (
    id TEXT PRIMARY KEY,
    alert_type TEXT,
    entity_type TEXT,
    entity_id TEXT,
    severity TEXT,
    message TEXT,
    ai_recommendation TEXT,
    dismissed BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for sprints table
CREATE INDEX IF NOT EXISTS idx_sprints_team_id ON sprints(team_id);
CREATE INDEX IF NOT EXISTS idx_sprints_status ON sprints(status);
CREATE INDEX IF NOT EXISTS idx_sprints_start_date ON sprints(start_date);
CREATE INDEX IF NOT EXISTS idx_sprints_end_date ON sprints(end_date);
CREATE INDEX IF NOT EXISTS idx_sprints_team_status ON sprints(team_id, status);

-- Indexes for epics table
CREATE INDEX IF NOT EXISTS idx_epics_team_id ON epics(team_id);
CREATE INDEX IF NOT EXISTS idx_epics_status ON epics(status);
CREATE INDEX IF NOT EXISTS idx_epics_jira_id ON epics(jira_id);
CREATE INDEX IF NOT EXISTS idx_epics_team_status ON epics(team_id, status);

-- Indexes for stories table
CREATE INDEX IF NOT EXISTS idx_stories_epic_id ON stories(epic_id);
CREATE INDEX IF NOT EXISTS idx_stories_sprint_id ON stories(sprint_id);
CREATE INDEX IF NOT EXISTS idx_stories_assignee_id ON stories(assignee_id);
CREATE INDEX IF NOT EXISTS idx_stories_status ON stories(status);
CREATE INDEX IF NOT EXISTS idx_stories_jira_id ON stories(jira_id);
CREATE INDEX IF NOT EXISTS idx_stories_sprint_status ON stories(sprint_id, status);
CREATE INDEX IF NOT EXISTS idx_stories_assignee_status ON stories(assignee_id, status);

-- Indexes for capacity_history table
CREATE INDEX IF NOT EXISTS idx_capacity_history_team_id ON capacity_history(team_id);
CREATE INDEX IF NOT EXISTS idx_capacity_history_week_start ON capacity_history(week_start);
CREATE INDEX IF NOT EXISTS idx_capacity_history_team_week ON capacity_history(team_id, week_start);

-- Indexes for planning_alerts table
CREATE INDEX IF NOT EXISTS idx_planning_alerts_entity ON planning_alerts(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_severity ON planning_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_dismissed ON planning_alerts(dismissed);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_alert_type ON planning_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_planning_alerts_type_severity ON planning_alerts(alert_type, severity, dismissed);
