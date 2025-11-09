-- Migration 021: Cost and ROI Tracking System
-- Business Intelligence: Connect engineering effort to business value

-- Cost Configuration: Define engineering costs
CREATE TABLE IF NOT EXISTS cost_configuration (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL, -- 'engineer', 'team', 'org'
    entity_id TEXT, -- engineer_id, team_id, or null for org defaults
    role TEXT, -- 'junior', 'mid', 'senior', 'staff', 'principal'
    monthly_cost REAL NOT NULL, -- fully-loaded cost (salary + benefits + overhead)
    currency TEXT DEFAULT 'USD',
    effective_from TEXT NOT NULL,
    effective_to TEXT, -- null = current
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Feature Values: Business value of features
CREATE TABLE IF NOT EXISTS feature_values (
    id TEXT PRIMARY KEY,
    feature_name TEXT NOT NULL,
    feature_description TEXT,
    value_type TEXT NOT NULL, -- 'arr_impact', 'customer_acquisition', 'retention_improvement', 'efficiency_gain'
    value_amount REAL, -- monetary value or percentage
    value_currency TEXT DEFAULT 'USD',
    confidence_level TEXT DEFAULT 'estimated', -- 'estimated', 'actual', 'validated'
    source TEXT, -- 'manual', 'productboard', 'jira', etc.
    source_reference TEXT, -- external ID
    time_period TEXT, -- 'Q4 2024', etc.
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Feature Work Items: Link features to engineering work
CREATE TABLE IF NOT EXISTS feature_work_items (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    work_item_type TEXT NOT NULL, -- 'story', 'epic', 'pr', 'issue'
    work_item_id TEXT NOT NULL, -- external ID from Jira, GitHub, etc.
    story_points INTEGER,
    actual_hours REAL,
    engineer_id TEXT,
    team_id TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

-- Feature Cost Analysis: Computed costs per feature
CREATE TABLE IF NOT EXISTS feature_costs (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    total_story_points INTEGER NOT NULL DEFAULT 0,
    total_hours REAL NOT NULL DEFAULT 0,
    total_cost REAL NOT NULL DEFAULT 0,
    cost_breakdown TEXT, -- JSON: cost by team, engineer, etc.
    computation_method TEXT NOT NULL, -- 'story_points', 'hours', 'hybrid'
    computed_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

-- ROI Calculations: Feature ROI analysis
CREATE TABLE IF NOT EXISTS roi_calculations (
    id TEXT PRIMARY KEY,
    feature_id TEXT NOT NULL,
    investment REAL NOT NULL, -- total engineering cost
    return_value REAL, -- business value
    roi_percentage REAL, -- (return - investment) / investment * 100
    payback_months REAL, -- months to break even
    confidence TEXT DEFAULT 'low', -- 'low', 'medium', 'high'
    notes TEXT,
    calculated_at TEXT NOT NULL,
    FOREIGN KEY (feature_id) REFERENCES feature_values(id) ON DELETE CASCADE
);

-- Engineering Investment: Time allocation categories
CREATE TABLE IF NOT EXISTS engineering_investment (
    id TEXT PRIMARY KEY,
    time_period TEXT NOT NULL, -- 'Q4 2024', '2024-W45', etc.
    team_id TEXT,
    category TEXT NOT NULL, -- 'customer_features', 'internal_tools', 'quality', 'tech_debt', 'infrastructure', 'operational'
    story_points INTEGER DEFAULT 0,
    hours REAL DEFAULT 0,
    cost REAL DEFAULT 0,
    feature_count INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Cost Efficiency Metrics: Derived metrics
CREATE TABLE IF NOT EXISTS cost_efficiency_metrics (
    id TEXT PRIMARY KEY,
    time_period TEXT NOT NULL,
    entity_type TEXT NOT NULL, -- 'team', 'org'
    entity_id TEXT,
    metric_name TEXT NOT NULL, -- 'cost_per_pr', 'cost_per_point', 'cost_per_feature'
    metric_value REAL NOT NULL,
    currency TEXT DEFAULT 'USD',
    computed_at TEXT NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_cost_config_entity ON cost_configuration(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_feature_values_period ON feature_values(time_period);
CREATE INDEX IF NOT EXISTS idx_feature_work_items_feature ON feature_work_items(feature_id);
CREATE INDEX IF NOT EXISTS idx_feature_work_items_engineer ON feature_work_items(engineer_id);
CREATE INDEX IF NOT EXISTS idx_feature_costs_feature ON feature_costs(feature_id);
CREATE INDEX IF NOT EXISTS idx_roi_calculations_feature ON roi_calculations(feature_id);
CREATE INDEX IF NOT EXISTS idx_engineering_investment_period ON engineering_investment(time_period, team_id);
CREATE INDEX IF NOT EXISTS idx_cost_efficiency_period ON cost_efficiency_metrics(time_period, entity_type, entity_id);
