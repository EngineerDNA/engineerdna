package db

import (
	"database/sql"
	"fmt"
	"time"
)

// TodayMetrics contains aggregated metrics for today's engineering activity
type TodayMetrics struct {
	TotalEvents     int
	ActiveEngineers int
	PRsCreated      int
	PRsMerged       int
	IssuesClosed    int
	CommitsToday    int
	AvgCycleTime    float64
	AvgReviewTime   float64
}

// GetTodayMetricsAggregated retrieves aggregated metrics for today using SQL aggregation
func (s *EventStore) GetTodayMetricsAggregated(startOfDay time.Time) (*TodayMetrics, error) {
	metrics := &TodayMetrics{}

	// Count total events
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM events WHERE timestamp >= ?
	`, startOfDay).Scan(&metrics.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Count unique engineers
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT engineer_id) FROM events
		WHERE timestamp >= ? AND engineer_id != ''
	`, startOfDay).Scan(&metrics.ActiveEngineers)
	if err != nil {
		return nil, fmt.Errorf("failed to count unique engineers: %w", err)
	}

	// Count PRs created, merged, issues closed, and commits using JSON extraction
	err = s.db.QueryRow(`
		SELECT
			COUNT(CASE WHEN type = 'pull_request' AND (json_extract(data, '$.status') = 'open' OR json_extract(data, '$.status') = 'created') THEN 1 END) as prs_created,
			COUNT(CASE WHEN type = 'pull_request' AND (json_extract(data, '$.status') = 'merged' OR json_extract(data, '$.status') = 'closed') THEN 1 END) as prs_merged,
			COUNT(CASE WHEN type = 'issue' AND (json_extract(data, '$.status') = 'closed' OR json_extract(data, '$.status') = 'completed') THEN 1 END) as issues_closed,
			COUNT(CASE WHEN type = 'commit' THEN 1 END) as commits
		FROM events WHERE timestamp >= ?
	`, startOfDay).Scan(&metrics.PRsCreated, &metrics.PRsMerged, &metrics.IssuesClosed, &metrics.CommitsToday)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate event counts: %w", err)
	}

	// Calculate average cycle time (PR created to merged)
	// This is more complex and requires loading some data, but with LIMIT
	// For V1, we'll use a simplified approach
	filters := map[string]interface{}{
		"since": startOfDay,
		"type":  "pull_request",
	}
	prEvents, err := s.List(filters, DefaultQueryLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR events for cycle time: %w", err)
	}

	// Group by PR ID and calculate cycle times
	prCreatedTimes := make(map[string]time.Time)
	prMergedTimes := make(map[string]time.Time)
	for _, event := range prEvents {
		var prID string
		if id, ok := event.Data["pr_id"].(string); ok && id != "" {
			prID = id
		} else {
			prID = event.SourceID
		}

		if status, ok := event.Data["status"].(string); ok {
			if status == "open" || status == "created" {
				prCreatedTimes[prID] = event.Timestamp
			} else if status == "merged" || status == "closed" {
				prMergedTimes[prID] = event.Timestamp
			}
		}
	}

	var totalCycleTime float64
	var cycleTimeCount int
	for prID, createdTime := range prCreatedTimes {
		if mergedTime, exists := prMergedTimes[prID]; exists {
			cycleHours := mergedTime.Sub(createdTime).Hours()
			if cycleHours >= 0 {
				totalCycleTime += cycleHours
				cycleTimeCount++
			}
		}
	}

	if cycleTimeCount > 0 {
		metrics.AvgCycleTime = totalCycleTime / float64(cycleTimeCount)
		// V1 Implementation: Uses percentage of cycle time as proxy for review time
		metrics.AvgReviewTime = metrics.AvgCycleTime * ReviewTimePercentOfCycleTime
	}

	return metrics, nil
}

// ThroughputDay contains throughput metrics for a single day
type ThroughputDay struct {
	Date         string
	PullRequests int
	Issues       int
}

// GetThroughputByDay retrieves daily throughput metrics using SQL aggregation
func (s *EventStore) GetThroughputByDay(startDate, endDate time.Time) ([]ThroughputDay, error) {
	rows, err := s.db.Query(`
		SELECT
			substr(timestamp, 1, 10) as day,
			COUNT(CASE WHEN type = 'pull_request' THEN 1 END) as pr_count,
			COUNT(CASE WHEN type = 'issue' THEN 1 END) as issue_count
		FROM events
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY substr(timestamp, 1, 10)
		ORDER BY day ASC
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query throughput by day: %w", err)
	}
	defer rows.Close()

	var results []ThroughputDay
	for rows.Next() {
		var day ThroughputDay
		err := rows.Scan(&day.Date, &day.PullRequests, &day.Issues)
		if err != nil {
			return nil, fmt.Errorf("failed to scan throughput day: %w", err)
		}
		results = append(results, day)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating throughput days: %w", err)
	}

	return results, nil
}

// CountOpenPRs counts the number of currently open pull requests
func (s *EventStore) CountOpenPRs() (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT source_id)
		FROM events
		WHERE type = 'pull_request'
		  AND (
			json_extract(data, '$.state') = 'open'
			OR json_extract(data, '$.state') = 'review_requested'
		  )
		  AND source_id NOT IN (
			SELECT source_id FROM events
			WHERE type = 'pull_request'
			  AND json_extract(data, '$.state') IN ('merged', 'closed')
		  )
		LIMIT ?
	`, DefaultQueryLimit).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count open PRs: %w", err)
	}

	return count, nil
}

// CountEvents counts events matching filters using SQL aggregation
func (s *EventStore) CountEvents(filters map[string]interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM events WHERE 1=1"
	args := []interface{}{}

	// Add time range filters
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query += " AND timestamp >= ?"
		args = append(args, startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok {
		query += " AND timestamp < ?"
		args = append(args, endTime)
	}

	// Add type filter (can be a single string or array)
	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		placeholders := make([]string, len(types))
		for i, t := range types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND type IN (%s)", joinStrings(placeholders, ","))
	} else if eventType, ok := filters["type"].(string); ok {
		query += " AND type = ?"
		args = append(args, eventType)
	}

	// Add source filter
	if source, ok := filters["source"].(string); ok {
		query += " AND source = ?"
		args = append(args, source)
	}

	// Add actor filter
	if actor, ok := filters["actor"].(string); ok {
		query += " AND actor = ?"
		args = append(args, actor)
	}

	// Add engineer_id filter
	if engineerID, ok := filters["engineer_id"].(string); ok {
		query += " AND engineer_id = ?"
		args = append(args, engineerID)
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// AvgFieldValue calculates the average of a numeric field in event data using SQL
func (s *EventStore) AvgFieldValue(field string, filters map[string]interface{}) (float64, error) {
	if field == "" {
		return 0, fmt.Errorf("field name is required")
	}

	// Validate field name to prevent SQL injection
	if !isValidFieldName(field) {
		return 0, fmt.Errorf("invalid field name: %s", field)
	}

	query := fmt.Sprintf("SELECT AVG(CAST(json_extract(data, '$.%s') AS REAL)) FROM events WHERE 1=1", field)
	args := []interface{}{}

	// Add time range filters
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query += " AND timestamp >= ?"
		args = append(args, startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok {
		query += " AND timestamp < ?"
		args = append(args, endTime)
	}

	// Add type filter
	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		placeholders := make([]string, len(types))
		for i, t := range types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND type IN (%s)", joinStrings(placeholders, ","))
	} else if eventType, ok := filters["type"].(string); ok {
		query += " AND type = ?"
		args = append(args, eventType)
	}

	// Add source filter
	if source, ok := filters["source"].(string); ok {
		query += " AND source = ?"
		args = append(args, source)
	}

	// Only include rows where the field exists and is not null
	query += fmt.Sprintf(" AND json_extract(data, '$.%s') IS NOT NULL", field)

	var avg sql.NullFloat64
	err := s.db.QueryRow(query, args...).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate average: %w", err)
	}

	if !avg.Valid {
		return 0, nil
	}

	return avg.Float64, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
