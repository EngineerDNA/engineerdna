-- Migration 024: Action Tracking & Recommendation Management
-- Closes the loop: AI recommendations → manager actions → measured outcomes

-- Recommendations: AI-generated or manual recommendations
CREATE TABLE IF NOT EXISTS recommendations (
    id TEXT PRIMARY KEY,
    source_type TEXT NOT NULL, -- 'ai_insight', 'alert', 'briefing', 'manual'
    source_id TEXT, -- alert_id, insight_id, etc.
    recommendation_type TEXT NOT NULL, -- 'check_in', 'adjust_workload', 'recognize_achievement', 'address_blocker', 'skill_development'
    priority TEXT NOT NULL, -- 'low', 'medium', 'high', 'urgent'
    subject_type TEXT NOT NULL, -- 'engineer', 'team', 'process'
    subject_id TEXT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    suggested_actions TEXT, -- JSON: array of specific actions
    context TEXT, -- JSON: why this recommendation was made
    assigned_to TEXT, -- manager_id
    created_at TEXT NOT NULL,
    expires_at TEXT, -- recommendations can expire
    status TEXT DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'dismissed', 'snoozed'
    FOREIGN KEY (assigned_to) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Actions: Actual actions taken by managers
CREATE TABLE IF NOT EXISTS actions (
    id TEXT PRIMARY KEY,
    recommendation_id TEXT, -- can be null for ad-hoc actions
    action_type TEXT NOT NULL, -- 'conversation', 'process_change', 'workload_adjustment', 'recognition', 'escalation'
    subject_type TEXT NOT NULL,
    subject_id TEXT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    taken_by TEXT NOT NULL, -- manager_id
    taken_at TEXT NOT NULL,
    evidence TEXT, -- JSON: notes, meeting ID, Slack thread, etc.
    expected_outcome TEXT,
    follow_up_date TEXT,
    FOREIGN KEY (recommendation_id) REFERENCES recommendations(id) ON DELETE SET NULL,
    FOREIGN KEY (taken_by) REFERENCES engineers(id) ON DELETE CASCADE
);

-- Action Outcomes: Measured effectiveness of actions
CREATE TABLE IF NOT EXISTS action_outcomes (
    id TEXT PRIMARY KEY,
    action_id TEXT NOT NULL,
    outcome_type TEXT NOT NULL, -- 'metric_improvement', 'issue_resolved', 'no_change', 'worsened'
    measured_metric TEXT, -- 'score', 'velocity', 'sentiment', etc.
    before_value REAL,
    after_value REAL,
    change_percentage REAL,
    time_to_impact_days INTEGER, -- how long until improvement seen
    effectiveness TEXT, -- 'highly_effective', 'somewhat_effective', 'not_effective'
    notes TEXT,
    measured_at TEXT NOT NULL,
    FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE CASCADE
);

-- Recommendation History: Track recommendation lifecycle
CREATE TABLE IF NOT EXISTS recommendation_history (
    id TEXT PRIMARY KEY,
    recommendation_id TEXT NOT NULL,
    status_change TEXT NOT NULL, -- 'created', 'assigned', 'in_progress', 'completed', 'dismissed', 'snoozed'
    changed_by TEXT,
    reason TEXT, -- why status changed
    changed_at TEXT NOT NULL,
    FOREIGN KEY (recommendation_id) REFERENCES recommendations(id) ON DELETE CASCADE
);

-- Follow-Ups: Scheduled follow-up reminders
CREATE TABLE IF NOT EXISTS follow_ups (
    id TEXT PRIMARY KEY,
    action_id TEXT NOT NULL,
    follow_up_date TEXT NOT NULL,
    follow_up_type TEXT NOT NULL, -- 'check_metric', 'check_in', 'verify_resolution'
    description TEXT,
    completed BOOLEAN DEFAULT false,
    completed_at TEXT,
    FOREIGN KEY (action_id) REFERENCES actions(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_recommendations_assigned ON recommendations(assigned_to, status);
CREATE INDEX IF NOT EXISTS idx_recommendations_subject ON recommendations(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_status ON recommendations(status);
CREATE INDEX IF NOT EXISTS idx_recommendations_expires ON recommendations(expires_at);
CREATE INDEX IF NOT EXISTS idx_recommendations_source ON recommendations(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_actions_recommendation ON actions(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_actions_taken_by ON actions(taken_by);
CREATE INDEX IF NOT EXISTS idx_actions_subject ON actions(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_action_outcomes_action ON action_outcomes(action_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_history_rec ON recommendation_history(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_action ON follow_ups(action_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_date ON follow_ups(follow_up_date, completed);
