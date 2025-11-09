-- Migration 027: Convert dashboard timestamp data to RFC3339 format
-- Issue: Timestamps were stored in Go's default format "2006-01-02 15:04:05.999999 -0700 MST"
-- Fix: Convert to RFC3339 format "2006-01-02T15:04:05Z"
-- Rule 34: UTC timestamps everywhere in RFC3339 format

-- Convert dashboards timestamps
UPDATE dashboards
SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(created_at, 1, 19))
WHERE created_at NOT LIKE '%T%Z';

UPDATE dashboards
SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(updated_at, 1, 19))
WHERE updated_at NOT LIKE '%T%Z';

-- Convert metric_snapshots timestamps
UPDATE metric_snapshots
SET computed_at = strftime('%Y-%m-%dT%H:%M:%SZ',
    substr(computed_at, 1, 19))
WHERE computed_at NOT LIKE '%T%Z';
