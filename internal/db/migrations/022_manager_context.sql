-- Migration 022: Manager Context and Sentiment System
-- Qualitative context to supplement quantitative metrics

-- Manager Notes: Private notes from managers about engineers/teams
CREATE TABLE IF NOT EXISTS manager_notes (
    id TEXT PRIMARY KEY,
    manager_id TEXT NOT NULL,
    subject_type TEXT NOT NULL, -- 'engineer', 'team', 'event', 'metric'
    subject_id TEXT NOT NULL, -- engineer_id, team_id, event_id, etc.
    note_type TEXT NOT NULL, -- '1on1', 'performance', 'incident', 'context', 'feedback'
    title TEXT,
    content TEXT NOT NULL,
    visibility TEXT DEFAULT 'private', -- 'private', 'shared_with_subject', 'team', 'org'
    tags TEXT, -- JSON array: ['personal', 'blocker', 'growth', etc.]
    mood TEXT, -- 'positive', 'neutral', 'negative', 'concerned'
    action_items TEXT, -- JSON array of action items
    linked_events TEXT, -- JSON array of event IDs
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (manager_id) REFERENCES engineers(id) ON DELETE CASCADE
);

-- Context Annotations: Link context to specific metrics/events
CREATE TABLE IF NOT EXISTS context_annotations (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL, -- 'metric_drop', 'score_change', 'event', 'alert'
    entity_id TEXT NOT NULL,
    annotation_type TEXT NOT NULL, -- 'explanation', 'mitigation', 'expectation'
    content TEXT NOT NULL,
    author_id TEXT NOT NULL,
    visibility TEXT DEFAULT 'team', -- 'private', 'team', 'org'
    created_at TEXT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES engineers(id) ON DELETE CASCADE
);

-- Team Context: Broader team-level context
CREATE TABLE IF NOT EXISTS team_context (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    context_type TEXT NOT NULL, -- 'velocity_explanation', 'capacity_change', 'process_change'
    time_period TEXT NOT NULL, -- 'Q4 2024', '2024-W45', etc.
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    impact TEXT, -- 'positive', 'negative', 'neutral'
    author_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES engineers(id) ON DELETE CASCADE
);

-- Sentiment Surveys: Pulse surveys for morale/sentiment
CREATE TABLE IF NOT EXISTS sentiment_surveys (
    id TEXT PRIMARY KEY,
    survey_type TEXT NOT NULL, -- 'weekly_pulse', 'quarterly', 'ad_hoc'
    title TEXT NOT NULL,
    description TEXT,
    questions TEXT NOT NULL, -- JSON array of questions
    target_audience TEXT NOT NULL, -- 'all', 'team:team_id', 'engineer:engineer_id'
    anonymous BOOLEAN DEFAULT 1,
    active BOOLEAN DEFAULT 1,
    created_by TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT,
    FOREIGN KEY (created_by) REFERENCES engineers(id) ON DELETE CASCADE
);

-- Survey Responses: Anonymous or identified responses
CREATE TABLE IF NOT EXISTS survey_responses (
    id TEXT PRIMARY KEY,
    survey_id TEXT NOT NULL,
    respondent_id TEXT, -- null if anonymous
    responses TEXT NOT NULL, -- JSON: question_id -> response
    mood_rating INTEGER, -- 1-5 scale
    text_feedback TEXT,
    submitted_at TEXT NOT NULL,
    FOREIGN KEY (survey_id) REFERENCES sentiment_surveys(id) ON DELETE CASCADE,
    FOREIGN KEY (respondent_id) REFERENCES engineers(id) ON DELETE SET NULL
);

-- Sentiment Analysis: Computed sentiment from various sources
CREATE TABLE IF NOT EXISTS sentiment_analysis (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL, -- 'engineer', 'team', 'org'
    entity_id TEXT,
    time_period TEXT NOT NULL,
    source TEXT NOT NULL, -- 'survey', 'code_review_tone', 'commit_messages'
    sentiment_score REAL NOT NULL, -- -1.0 to 1.0
    confidence REAL NOT NULL, -- 0.0 to 1.0
    sample_size INTEGER,
    details TEXT, -- JSON: breakdown by dimension
    computed_at TEXT NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_manager_notes_manager ON manager_notes(manager_id);
CREATE INDEX IF NOT EXISTS idx_manager_notes_subject ON manager_notes(subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_manager_notes_created ON manager_notes(created_at);
CREATE INDEX IF NOT EXISTS idx_context_annotations_entity ON context_annotations(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_team_context_team ON team_context(team_id, time_period);
CREATE INDEX IF NOT EXISTS idx_sentiment_surveys_active ON sentiment_surveys(active);
CREATE INDEX IF NOT EXISTS idx_survey_responses_survey ON survey_responses(survey_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_analysis_entity ON sentiment_analysis(entity_type, entity_id, time_period);
