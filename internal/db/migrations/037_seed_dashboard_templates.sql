-- Migration 037: Seed Dashboard Templates
-- Dashboard-First UI/UX Simplification
-- Created: 2025-11-08
-- Populates 12 default dashboard templates for role-based onboarding

-- Individual Contributor Templates (2)

-- IC: Personal Performance (primary)
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-ic-personal-performance',
    'Personal Performance',
    'Track your contributions, code quality, and growth',
    'ic',
    'personal',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Performance Score", "data_source": "/api/engineers/me/score", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "timeseries", "title": "Weekly Contributions", "data_source": "/api/engineers/me/activity", "query_params": {"period": "30d", "granularity": "day"}, "position": {"x": 3, "y": 0, "w": 9, "h": 4}},
        {"id": "w3", "type": "bar", "title": "Code Quality Metrics", "data_source": "/api/engineers/me/quality", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 2, "w": 3, "h": 2}},
        {"id": "w4", "type": "feed", "title": "Recent Activity", "data_source": "/api/events", "query_params": {"actor": "me", "limit": "10"}, "position": {"x": 0, "y": 4, "w": 6, "h": 4}},
        {"id": "w5", "type": "table", "title": "Active Goals", "data_source": "/api/goals", "query_params": {"engineer_id": "me", "status": "active"}, "position": {"x": 6, "y": 4, "w": 6, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- IC: My Activity
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-ic-my-activity',
    'My Activity',
    'Detailed view of your commits, PRs, and code reviews',
    'ic',
    'personal',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "PRs This Week", "data_source": "/api/events/count", "query_params": {"type": "pull_request", "actor": "me", "period": "7d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Code Reviews", "data_source": "/api/events/count", "query_params": {"type": "code_review", "actor": "me", "period": "7d"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Commits", "data_source": "/api/events/count", "query_params": {"type": "commit", "actor": "me", "period": "7d"}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "timeseries", "title": "Activity Trend", "data_source": "/api/engineers/me/activity-trend", "query_params": {"period": "90d"}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w5", "type": "feed", "title": "Activity Feed", "data_source": "/api/events", "query_params": {"actor": "me", "limit": "20"}, "position": {"x": 0, "y": 6, "w": 12, "h": 6}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Team Lead / Manager Templates (5)

-- Manager: Command Center (primary)
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-manager-command-center',
    'Command Center',
    'Real-time team health, alerts, and quick actions',
    'manager',
    'team',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Team Score", "data_source": "/api/teams/my-team/score", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Active Alerts", "data_source": "/api/alerts/count", "query_params": {"team_id": "my-team", "status": "fired"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Active Goals", "data_source": "/api/goals/count", "query_params": {"team_id": "my-team", "status": "active"}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "Team Health", "data_source": "/api/teams/my-team/health", "query_params": {}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "feed", "title": "Alerts & Attention Items", "data_source": "/api/alerts", "query_params": {"team_id": "my-team", "limit": "5"}, "position": {"x": 0, "y": 2, "w": 6, "h": 4}},
        {"id": "w6", "type": "table", "title": "Team Members", "data_source": "/api/teams/my-team/members", "query_params": {}, "position": {"x": 6, "y": 2, "w": 6, "h": 4}},
        {"id": "w7", "type": "timeseries", "title": "Team Velocity", "data_source": "/api/teams/my-team/velocity", "query_params": {"period": "30d"}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Manager: Weekly Briefing
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-manager-weekly-briefing',
    'Weekly Briefing',
    'AI-generated summary and talking points for your team',
    'manager',
    'team',
    '{"widgets": [
        {"id": "w1", "type": "status", "title": "Week Summary", "data_source": "/api/briefing/summary", "query_params": {"team_id": "my-team", "period": "7d"}, "position": {"x": 0, "y": 0, "w": 12, "h": 3}},
        {"id": "w2", "type": "table", "title": "Key Metrics", "data_source": "/api/briefing/metrics", "query_params": {"team_id": "my-team", "period": "7d"}, "position": {"x": 0, "y": 3, "w": 6, "h": 4}},
        {"id": "w3", "type": "feed", "title": "Talking Points", "data_source": "/api/briefing/talking-points", "query_params": {"team_id": "my-team", "period": "7d"}, "position": {"x": 6, "y": 3, "w": 6, "h": 4}},
        {"id": "w4", "type": "table", "title": "Top Performers", "data_source": "/api/teams/my-team/top-performers", "query_params": {"period": "7d", "limit": "5"}, "position": {"x": 0, "y": 7, "w": 6, "h": 3}},
        {"id": "w5", "type": "feed", "title": "Attention Items", "data_source": "/api/briefing/attention-items", "query_params": {"team_id": "my-team"}, "position": {"x": 6, "y": 7, "w": 6, "h": 3}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Manager: Team Performance
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-manager-team-performance',
    'Team Performance',
    'Detailed performance metrics and trends for your team',
    'manager',
    'team',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Team Score", "data_source": "/api/teams/my-team/score", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Throughput", "data_source": "/api/teams/my-team/throughput", "query_params": {"period": "7d"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Quality Score", "data_source": "/api/teams/my-team/quality", "query_params": {"period": "7d"}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "number", "title": "Cycle Time (hrs)", "data_source": "/api/teams/my-team/cycle-time", "query_params": {"period": "7d"}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "timeseries", "title": "Performance Trend", "data_source": "/api/teams/my-team/score-trend", "query_params": {"period": "90d"}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w6", "type": "table", "title": "Engineer Performance", "data_source": "/api/teams/my-team/members-performance", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Manager: Sprint Planning
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-manager-sprint-planning',
    'Sprint Planning',
    'Capacity, velocity, and sprint goal tracking',
    'manager',
    'team',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Team Velocity", "data_source": "/api/planning/velocity", "query_params": {"team_id": "my-team", "period": "30d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Available Capacity", "data_source": "/api/planning/capacity", "query_params": {"team_id": "my-team"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Sprint Progress", "data_source": "/api/planning/sprint-progress", "query_params": {"team_id": "my-team"}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "On Track Status", "data_source": "/api/planning/status", "query_params": {"team_id": "my-team"}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "timeseries", "title": "Velocity Trend", "data_source": "/api/planning/velocity-trend", "query_params": {"team_id": "my-team", "period": "90d"}, "position": {"x": 0, "y": 2, "w": 6, "h": 4}},
        {"id": "w6", "type": "bar", "title": "Capacity by Engineer", "data_source": "/api/planning/capacity-breakdown", "query_params": {"team_id": "my-team"}, "position": {"x": 6, "y": 2, "w": 6, "h": 4}},
        {"id": "w7", "type": "table", "title": "Sprint Goals", "data_source": "/api/goals", "query_params": {"team_id": "my-team", "type": "sprint"}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Manager: Team Management
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-manager-team-management',
    'Team Management',
    'Engineer growth, skills, and 1-on-1 tracking',
    'manager',
    'team',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Team Size", "data_source": "/api/teams/my-team/size", "query_params": {}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Avg Tenure (months)", "data_source": "/api/teams/my-team/avg-tenure", "query_params": {}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Skills Coverage", "data_source": "/api/teams/my-team/skill-coverage", "query_params": {}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "Sentiment", "data_source": "/api/teams/my-team/sentiment", "query_params": {}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "table", "title": "Team Members", "data_source": "/api/teams/my-team/members", "query_params": {}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w6", "type": "table", "title": "Skill Gaps", "data_source": "/api/teams/my-team/skill-gaps", "query_params": {}, "position": {"x": 0, "y": 6, "w": 6, "h": 4}},
        {"id": "w7", "type": "feed", "title": "Recent 1-on-1s", "data_source": "/api/context/notes", "query_params": {"team_id": "my-team", "limit": "10"}, "position": {"x": 6, "y": 6, "w": 6, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Director / VP Templates (5)

-- Director: Executive Overview (primary)
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-director-executive-overview',
    'Executive Overview',
    'High-level organization metrics and key initiatives',
    'director',
    'org',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Org Performance", "data_source": "/api/org/score", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Active Teams", "data_source": "/api/teams/count", "query_params": {"status": "active"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Total Engineers", "data_source": "/api/engineers/count", "query_params": {}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "number", "title": "Active Goals", "data_source": "/api/goals/count", "query_params": {"status": "active"}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "timeseries", "title": "Org Performance Trend", "data_source": "/api/org/score-trend", "query_params": {"period": "90d"}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w6", "type": "table", "title": "Team Scorecard", "data_source": "/api/teams", "query_params": {"include_metrics": "true"}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Director: Organization Scorecard
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-director-org-scorecard',
    'Organization Scorecard',
    'Comprehensive metrics across all teams',
    'director',
    'org',
    '{"widgets": [
        {"id": "w1", "type": "table", "title": "Team Performance", "data_source": "/api/teams/scorecard", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 0, "w": 12, "h": 6}},
        {"id": "w2", "type": "bar", "title": "Team Comparison", "data_source": "/api/teams/comparison", "query_params": {"metric": "total_score", "period": "7d"}, "position": {"x": 0, "y": 6, "w": 6, "h": 4}},
        {"id": "w3", "type": "timeseries", "title": "Org Metrics Trend", "data_source": "/api/org/metrics-trend", "query_params": {"period": "90d"}, "position": {"x": 6, "y": 6, "w": 6, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Director: Team Comparison
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-director-team-comparison',
    'Team Comparison',
    'Side-by-side team performance analysis',
    'director',
    'org',
    '{"widgets": [
        {"id": "w1", "type": "bar", "title": "Performance Scores", "data_source": "/api/teams/comparison", "query_params": {"metric": "total_score", "period": "7d"}, "position": {"x": 0, "y": 0, "w": 6, "h": 4}},
        {"id": "w2", "type": "bar", "title": "Throughput", "data_source": "/api/teams/comparison", "query_params": {"metric": "throughput", "period": "7d"}, "position": {"x": 6, "y": 0, "w": 6, "h": 4}},
        {"id": "w3", "type": "bar", "title": "Quality Scores", "data_source": "/api/teams/comparison", "query_params": {"metric": "quality", "period": "7d"}, "position": {"x": 0, "y": 4, "w": 6, "h": 4}},
        {"id": "w4", "type": "bar", "title": "Cycle Time", "data_source": "/api/teams/comparison", "query_params": {"metric": "cycle_time", "period": "7d"}, "position": {"x": 6, "y": 4, "w": 6, "h": 4}},
        {"id": "w5", "type": "table", "title": "Detailed Comparison", "data_source": "/api/teams/detailed-comparison", "query_params": {"period": "7d"}, "position": {"x": 0, "y": 8, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Director: Cost & ROI Analysis
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-director-cost-roi',
    'Cost & ROI Analysis',
    'Engineering costs, ROI, and resource allocation',
    'director',
    'org',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Monthly Eng Cost", "data_source": "/api/cost/total", "query_params": {"period": "30d"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Cost per Feature", "data_source": "/api/roi/cost-per-feature", "query_params": {"period": "30d"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "ROI Score", "data_source": "/api/roi/score", "query_params": {"period": "90d"}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "number", "title": "Avg Payback (days)", "data_source": "/api/roi/avg-payback", "query_params": {"period": "90d"}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "timeseries", "title": "Cost Trend", "data_source": "/api/cost/trend", "query_params": {"period": "180d"}, "position": {"x": 0, "y": 2, "w": 6, "h": 4}},
        {"id": "w6", "type": "bar", "title": "Cost by Team", "data_source": "/api/cost/by-team", "query_params": {"period": "30d"}, "position": {"x": 6, "y": 2, "w": 6, "h": 4}},
        {"id": "w7", "type": "table", "title": "Feature ROI", "data_source": "/api/roi/features", "query_params": {"period": "90d", "limit": "10"}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Director: Capacity Planning
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-director-capacity-planning',
    'Capacity Planning',
    'Org-wide capacity, hiring needs, and resource forecasting',
    'director',
    'org',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Total Capacity", "data_source": "/api/planning/org-capacity", "query_params": {}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Utilization %", "data_source": "/api/planning/utilization", "query_params": {}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Hiring Target", "data_source": "/api/planning/hiring-target", "query_params": {}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "Capacity Status", "data_source": "/api/planning/capacity-status", "query_params": {}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "bar", "title": "Capacity by Team", "data_source": "/api/planning/capacity-by-team", "query_params": {}, "position": {"x": 0, "y": 2, "w": 6, "h": 4}},
        {"id": "w6", "type": "timeseries", "title": "Capacity Forecast", "data_source": "/api/planning/capacity-forecast", "query_params": {"period": "180d"}, "position": {"x": 6, "y": 2, "w": 6, "h": 4}},
        {"id": "w7", "type": "table", "title": "Team Capacity Details", "data_source": "/api/planning/team-capacity-details", "query_params": {}, "position": {"x": 0, "y": 6, "w": 12, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Admin Templates (2)

-- Admin: System Health
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-admin-system-health',
    'System Health',
    'Plugin status, data quality, and system diagnostics',
    'admin',
    'admin',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Active Plugins", "data_source": "/api/plugins/count", "query_params": {"status": "configured"}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Events Today", "data_source": "/api/events/count", "query_params": {"period": "1d"}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Last Sync", "data_source": "/api/plugins/last-sync", "query_params": {}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "System Status", "data_source": "/api/health", "query_params": {}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "table", "title": "Plugin Status", "data_source": "/api/plugins", "query_params": {}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w6", "type": "timeseries", "title": "Event Volume", "data_source": "/api/events/volume", "query_params": {"period": "30d"}, "position": {"x": 0, "y": 6, "w": 6, "h": 4}},
        {"id": "w7", "type": "feed", "title": "System Logs", "data_source": "/api/logs", "query_params": {"limit": "20"}, "position": {"x": 6, "y": 6, "w": 6, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);

-- Admin: Data Management
INSERT INTO dashboard_templates (id, name, description, role, category, layout, is_system, created_at, updated_at)
VALUES (
    'template-admin-data-management',
    'Data Management',
    'Identity resolution, data quality, and anonymization',
    'admin',
    'admin',
    '{"widgets": [
        {"id": "w1", "type": "number", "title": "Unresolved Identities", "data_source": "/api/identities/unresolved-count", "query_params": {}, "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
        {"id": "w2", "type": "number", "title": "Total Engineers", "data_source": "/api/engineers/count", "query_params": {}, "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
        {"id": "w3", "type": "number", "title": "Total Events", "data_source": "/api/events/count", "query_params": {}, "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
        {"id": "w4", "type": "status", "title": "Data Quality", "data_source": "/api/data-quality/status", "query_params": {}, "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
        {"id": "w5", "type": "table", "title": "Unresolved Identities", "data_source": "/api/identities/unresolved", "query_params": {"limit": "10"}, "position": {"x": 0, "y": 2, "w": 12, "h": 4}},
        {"id": "w6", "type": "table", "title": "Recent Anonymizations", "data_source": "/api/anonymization/recent", "query_params": {"limit": "10"}, "position": {"x": 0, "y": 6, "w": 6, "h": 4}},
        {"id": "w7", "type": "table", "title": "Export History", "data_source": "/api/exports", "query_params": {"limit": "10"}, "position": {"x": 6, "y": 6, "w": 6, "h": 4}}
    ]}',
    1,
    datetime('now', 'utc'),
    datetime('now', 'utc')
);
