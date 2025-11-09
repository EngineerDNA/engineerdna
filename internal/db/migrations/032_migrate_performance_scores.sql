-- Migration 032: Migrate Performance Scores to metric_values
-- Consolidates performance_scores and team_performance_scores into universal metric_values table

-- 1. Migrate individual engineer performance scores
-- Total score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_total_' || id,
    'engineer_total_score',
    'scoring_system',
    week_start,
    'weekly',
    total_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE total_score IS NOT NULL
AND 'ps_total_' || id NOT IN (SELECT id FROM metric_values);

-- Throughput score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_throughput_' || id,
    'engineer_throughput_score',
    'scoring_system',
    week_start,
    'weekly',
    throughput_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE throughput_score IS NOT NULL
AND 'ps_throughput_' || id NOT IN (SELECT id FROM metric_values);

-- Quality score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_quality_' || id,
    'engineer_quality_score',
    'scoring_system',
    week_start,
    'weekly',
    quality_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE quality_score IS NOT NULL
AND 'ps_quality_' || id NOT IN (SELECT id FROM metric_values);

-- Speed score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_speed_' || id,
    'engineer_speed_score',
    'scoring_system',
    week_start,
    'weekly',
    speed_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE speed_score IS NOT NULL
AND 'ps_speed_' || id NOT IN (SELECT id FROM metric_values);

-- Collaboration score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_collaboration_' || id,
    'engineer_collaboration_score',
    'scoring_system',
    week_start,
    'weekly',
    collaboration_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE collaboration_score IS NOT NULL
AND 'ps_collaboration_' || id NOT IN (SELECT id FROM metric_values);

-- Impact score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'ps_impact_' || id,
    'engineer_impact_score',
    'scoring_system',
    week_start,
    'weekly',
    impact_score,
    'score',
    json_object('engineer_id', engineer_id),
    created_at
FROM performance_scores
WHERE impact_score IS NOT NULL
AND 'ps_impact_' || id NOT IN (SELECT id FROM metric_values);

-- 2. Migrate team performance scores
-- Total score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_total_' || id,
    'team_total_score',
    'scoring_system',
    week_start,
    'weekly',
    total_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE total_score IS NOT NULL
AND 'tps_total_' || id NOT IN (SELECT id FROM metric_values);

-- Throughput score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_throughput_' || id,
    'team_throughput_score',
    'scoring_system',
    week_start,
    'weekly',
    throughput_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE throughput_score IS NOT NULL
AND 'tps_throughput_' || id NOT IN (SELECT id FROM metric_values);

-- Quality score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_quality_' || id,
    'team_quality_score',
    'scoring_system',
    week_start,
    'weekly',
    quality_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE quality_score IS NOT NULL
AND 'tps_quality_' || id NOT IN (SELECT id FROM metric_values);

-- Speed score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_speed_' || id,
    'team_speed_score',
    'scoring_system',
    week_start,
    'weekly',
    speed_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE speed_score IS NOT NULL
AND 'tps_speed_' || id NOT IN (SELECT id FROM metric_values);

-- Collaboration score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_collaboration_' || id,
    'team_collaboration_score',
    'scoring_system',
    week_start,
    'weekly',
    collaboration_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE collaboration_score IS NOT NULL
AND 'tps_collaboration_' || id NOT IN (SELECT id FROM metric_values);

-- Impact score
INSERT INTO metric_values (id, metric_name, source, timestamp, granularity, value, unit, dimensions, created_at)
SELECT
    'tps_impact_' || id,
    'team_impact_score',
    'scoring_system',
    week_start,
    'weekly',
    impact_score,
    'score',
    json_object('team_id', team_id, 'member_count', member_count),
    created_at
FROM team_performance_scores
WHERE impact_score IS NOT NULL
AND 'tps_impact_' || id NOT IN (SELECT id FROM metric_values);

-- 3. Drop old tables
DROP TABLE IF EXISTS performance_scores;
DROP TABLE IF EXISTS team_performance_scores;
