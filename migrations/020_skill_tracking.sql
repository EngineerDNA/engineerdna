-- Migration 020: Skill Development Tracking System
-- Automatic skill detection and progression tracking

-- Skill Taxonomy: Define available skills
CREATE TABLE IF NOT EXISTS skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL, -- 'technical', 'leadership', 'communication'
    subcategory TEXT, -- 'system_design', 'code_quality', 'mentoring', etc.
    description TEXT,
    measurement_criteria TEXT, -- JSON: how to measure this skill
    created_at TEXT NOT NULL
);

-- Engineer Skills: Track skill levels for each engineer
CREATE TABLE IF NOT EXISTS engineer_skills (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    level_score INTEGER NOT NULL DEFAULT 0, -- 0-100
    previous_level_score INTEGER,
    trajectory TEXT DEFAULT 'stable', -- 'improving', 'stable', 'declining'
    last_evaluated TEXT NOT NULL,
    evidence_count INTEGER DEFAULT 0, -- # of supporting data points
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    UNIQUE(engineer_id, skill_id)
);

-- Skill Evidence: Track proof of skill usage/growth
CREATE TABLE IF NOT EXISTS skill_evidence (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    evidence_type TEXT NOT NULL, -- 'pr_complexity', 'review_depth', 'design_doc', 'mentoring_session'
    evidence_source TEXT, -- event_id, pr_id, etc.
    strength REAL NOT NULL, -- 0.0-1.0 (how strongly this proves the skill)
    context TEXT, -- JSON: additional context
    detected_at TEXT NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

-- Skill Progression History: Track changes over time
CREATE TABLE IF NOT EXISTS skill_progression (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    previous_score INTEGER NOT NULL,
    new_score INTEGER NOT NULL,
    change_reason TEXT, -- 'evidence_accumulated', 'manual_update', 'evaluation'
    evaluated_at TEXT NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE
);

-- Skill Goals: Link skills to goals
CREATE TABLE IF NOT EXISTS skill_goals (
    id TEXT PRIMARY KEY,
    engineer_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    goal_id TEXT, -- optional: link to formal goal
    current_level INTEGER NOT NULL,
    target_level INTEGER NOT NULL,
    target_date TEXT,
    milestones TEXT, -- JSON: array of milestones
    status TEXT DEFAULT 'active', -- 'active', 'completed', 'abandoned'
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (engineer_id) REFERENCES engineers(id) ON DELETE CASCADE,
    FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_engineer_skills_engineer ON engineer_skills(engineer_id);
CREATE INDEX IF NOT EXISTS idx_engineer_skills_skill ON engineer_skills(skill_id);
CREATE INDEX IF NOT EXISTS idx_engineer_skills_trajectory ON engineer_skills(trajectory);
CREATE INDEX IF NOT EXISTS idx_skill_evidence_engineer ON skill_evidence(engineer_id, skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_evidence_detected ON skill_evidence(detected_at);
CREATE INDEX IF NOT EXISTS idx_skill_progression_engineer ON skill_progression(engineer_id);
CREATE INDEX IF NOT EXISTS idx_skill_progression_evaluated ON skill_progression(evaluated_at);
CREATE INDEX IF NOT EXISTS idx_skill_goals_engineer ON skill_goals(engineer_id);
CREATE INDEX IF NOT EXISTS idx_skill_goals_status ON skill_goals(status);
