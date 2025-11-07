package dashboards

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// TODO(technical-debt): Refactor to use repository pattern instead of direct database queries.
// Currently this service makes 11 direct SQL calls which violates the repository pattern:
// - Lines 85, 105, 125: Org-wide metric queries (should use eventStore, alertsStore, goalsStore)
// - Lines 168, 241: Team/engineer snapshot queries with JOINs (should use composite repository methods)
// - Lines 303, 337, 343, 353, 357, 398: Individual metric calculations (should use store methods)
// All queries are currently parameterized (no SQL injection risk), but this should be refactored
// for better maintainability and consistency with the repository pattern used elsewhere.

const (
	// HoursPerDay is the number of hours in a day for time conversions
	HoursPerDay = 24
)

// Service handles dashboard metric snapshot computation
type Service struct {
	database       *sql.DB
	dashboardStore *db.DashboardStore
	scoringStore   *db.ScoringStore
	eventStore     *db.EventStore
	alertsStore    *db.AlertsStore
	goalsStore     *db.GoalsStore
	teamStore      *db.TeamStore
}

// NewService creates a new dashboard service
func NewService(
	database *sql.DB,
	dashboardStore *db.DashboardStore,
	scoringStore *db.ScoringStore,
	eventStore *db.EventStore,
	alertsStore *db.AlertsStore,
	goalsStore *db.GoalsStore,
	teamStore *db.TeamStore,
) *Service {
	return &Service{
		database:       database,
		dashboardStore: dashboardStore,
		scoringStore:   scoringStore,
		eventStore:     eventStore,
		alertsStore:    alertsStore,
		goalsStore:     goalsStore,
		teamStore:      teamStore,
	}
}

// ComputeAllSnapshots computes snapshots for all metrics
func (s *Service) ComputeAllSnapshots() error {
	// Compute snapshots for today
	today := time.Now().UTC().Format("2006-01-02")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	// Compute org-wide snapshots
	if err := s.computeOrgSnapshots(today, today); err != nil {
		return fmt.Errorf("failed to compute org snapshots: %w", err)
	}

	// Compute team snapshots
	if err := s.computeTeamSnapshots(today, today); err != nil {
		return fmt.Errorf("failed to compute team snapshots: %w", err)
	}

	// Compute engineer snapshots
	if err := s.computeEngineerSnapshots(today, today); err != nil {
		return fmt.Errorf("failed to compute engineer snapshots: %w", err)
	}

	// Compute yesterday's snapshots for comparison
	if err := s.computeOrgSnapshots(yesterday, yesterday); err != nil {
		return fmt.Errorf("failed to compute org snapshots for yesterday: %w", err)
	}

	return nil
}

// computeOrgSnapshots computes organization-wide metric snapshots
func (s *Service) computeOrgSnapshots(periodStart, periodEnd string) error {
	snapshots := []*models.MetricSnapshot{}

	// PR Volume (org-wide)
	var prCount int
	err := s.database.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE type = 'pull_request'
		  AND DATE(timestamp) = ?
	`, periodStart).Scan(&prCount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to count PRs: %w", err)
	}

	snapshots = append(snapshots, &models.MetricSnapshot{
		MetricName:  "pr_volume",
		EntityType:  "org",
		EntityID:    "",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       float64(prCount),
	})

	// Active Alerts (org-wide)
	var alertCount int
	err = s.database.QueryRow(`
		SELECT COUNT(*) FROM alert_instances
		WHERE fired_at >= ? AND fired_at < ?
		  AND resolved_at IS NULL
	`, periodStart, periodEnd+" 23:59:59").Scan(&alertCount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to count alerts: %w", err)
	}

	snapshots = append(snapshots, &models.MetricSnapshot{
		MetricName:  "active_alerts",
		EntityType:  "org",
		EntityID:    "",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       float64(alertCount),
	})

	// Active Goals (org-wide)
	var goalCount int
	err = s.database.QueryRow(`
		SELECT COUNT(*) FROM goals
		WHERE status = 'active'
	`).Scan(&goalCount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to count goals: %w", err)
	}

	snapshots = append(snapshots, &models.MetricSnapshot{
		MetricName:  "active_goals",
		EntityType:  "org",
		EntityID:    "",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       float64(goalCount),
	})

	// Save all snapshots
	return s.dashboardStore.CreateMetricSnapshotsBatch(snapshots)
}

// computeTeamSnapshots computes team-level metric snapshots
func (s *Service) computeTeamSnapshots(periodStart, periodEnd string) error {
	// Single query with JOINs to get all team metrics at once
	query := `
		SELECT
			t.id,
			COALESCE(tps.total_score, 0) as team_score,
			COUNT(DISTINCT CASE WHEN e.type = 'pull_request' AND DATE(e.timestamp) = ? THEN e.id END) as pr_count,
			COUNT(DISTINCT CASE WHEN ai.fired_at >= ? AND ai.fired_at < ? AND ai.resolved_at IS NULL THEN ai.id END) as alert_count
		FROM teams t
		LEFT JOIN (
			SELECT team_id, total_score, created_at,
				ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY created_at DESC) as rn
			FROM team_performance_scores
		) tps ON t.id = tps.team_id AND tps.rn = 1
		LEFT JOIN team_membership tm ON t.id = tm.team_id AND tm.left_at IS NULL
		LEFT JOIN engineers eng ON tm.member_id = eng.id
		LEFT JOIN events e ON eng.email = e.actor
		LEFT JOIN alert_instances ai ON ai.entity_type = 'team' AND ai.entity_id = t.id
		GROUP BY t.id, tps.total_score
	`

	rows, err := s.database.Query(query, periodStart, periodStart, periodEnd+" 23:59:59")
	if err != nil {
		return fmt.Errorf("failed to compute team snapshots: %w", err)
	}
	defer rows.Close()

	snapshots := []*models.MetricSnapshot{}
	for rows.Next() {
		var teamID string
		var teamScore, prCount, alertCount float64

		if err := rows.Scan(&teamID, &teamScore, &prCount, &alertCount); err != nil {
			return fmt.Errorf("failed to scan team snapshot row: %w", err)
		}

		// Create snapshots for each metric
		if teamScore > 0 {
			snapshots = append(snapshots, &models.MetricSnapshot{
				MetricName:  "team_score",
				EntityType:  "team",
				EntityID:    teamID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Value:       teamScore,
			})
		}
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "pr_volume",
			EntityType:  "team",
			EntityID:    teamID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       prCount,
		})
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "active_alerts",
			EntityType:  "team",
			EntityID:    teamID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       alertCount,
		})
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating team snapshot rows: %w", err)
	}

	if len(snapshots) > 0 {
		return s.dashboardStore.CreateMetricSnapshotsBatch(snapshots)
	}
	return nil
}

// computeEngineerSnapshots computes engineer-level metric snapshots
func (s *Service) computeEngineerSnapshots(periodStart, periodEnd string) error {
	// Single query with JOINs to get all engineer metrics at once
	query := `
		SELECT
			e.id,
			COALESCE(ps.total_score, 0) as engineer_score,
			COUNT(DISTINCT CASE WHEN ev.type = 'pull_request' AND DATE(ev.timestamp) = ? THEN ev.id END) as pr_count
		FROM engineers e
		LEFT JOIN (
			SELECT engineer_id, total_score, created_at,
				ROW_NUMBER() OVER (PARTITION BY engineer_id ORDER BY created_at DESC) as rn
			FROM performance_scores
		) ps ON e.id = ps.engineer_id AND ps.rn = 1
		LEFT JOIN events ev ON e.email = ev.actor
		GROUP BY e.id, ps.total_score
		LIMIT ?
	`

	rows, err := s.database.Query(query, periodStart, db.MaxQueryLimit)
	if err != nil {
		return fmt.Errorf("failed to compute engineer snapshots: %w", err)
	}
	defer rows.Close()

	snapshots := []*models.MetricSnapshot{}
	for rows.Next() {
		var engineerID string
		var engineerScore, prCount float64

		if err := rows.Scan(&engineerID, &engineerScore, &prCount); err != nil {
			return fmt.Errorf("failed to scan engineer snapshot row: %w", err)
		}

		// Create snapshots for each metric
		if engineerScore > 0 {
			snapshots = append(snapshots, &models.MetricSnapshot{
				MetricName:  "engineer_score",
				EntityType:  "engineer",
				EntityID:    engineerID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Value:       engineerScore,
			})
		}
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "pr_count",
			EntityType:  "engineer",
			EntityID:    engineerID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       prCount,
		})
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating engineer snapshot rows: %w", err)
	}

	if len(snapshots) > 0 {
		return s.dashboardStore.CreateMetricSnapshotsBatch(snapshots)
	}
	return nil
}

// ComputeSnapshotForPeriod computes a specific metric snapshot for a date range
func (s *Service) ComputeSnapshotForPeriod(metricName, entityType, entityID, periodStart, periodEnd string) (*models.MetricSnapshot, error) {
	switch metricName {
	case "team_score":
		return s.computeTeamScore(entityID, periodStart, periodEnd)
	case "pr_volume", "pr_count":
		return s.computePRVolume(entityType, entityID, periodStart, periodEnd)
	case "cycle_time":
		return s.computeCycleTime(entityType, entityID, periodStart, periodEnd)
	default:
		return nil, fmt.Errorf("unknown metric: %s", metricName)
	}
}

func (s *Service) computeTeamScore(teamID, periodStart, periodEnd string) (*models.MetricSnapshot, error) {
	var score sql.NullFloat64
	err := s.database.QueryRow(`
		SELECT total_score FROM team_performance_scores
		WHERE team_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, teamID).Scan(&score)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if !score.Valid {
		return nil, nil
	}

	return &models.MetricSnapshot{
		MetricName:  "team_score",
		EntityType:  "team",
		EntityID:    teamID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       score.Float64,
	}, nil
}

func (s *Service) computePRVolume(entityType, entityID, periodStart, periodEnd string) (*models.MetricSnapshot, error) {
	var count int
	var err error

	switch entityType {
	case "org":
		err = s.database.QueryRow(`
			SELECT COUNT(*) FROM events
			WHERE type = 'pull_request'
			  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
		`, periodStart, periodEnd).Scan(&count)
	case "team":
		err = s.database.QueryRow(`
			SELECT COUNT(DISTINCT e.id) FROM events e
			INNER JOIN engineers eng ON e.actor = eng.email
			INNER JOIN team_membership tm ON eng.id = tm.member_id AND tm.left_at IS NULL
			WHERE tm.team_id = ?
			  AND e.type = 'pull_request'
			  AND DATE(e.timestamp) >= ? AND DATE(e.timestamp) <= ?
		`, entityID, periodStart, periodEnd).Scan(&count)
	case "engineer":
		var email string
		err = s.database.QueryRow(`SELECT email FROM engineers WHERE id = ?`, entityID).Scan(&email)
		if err != nil {
			return nil, err
		}
		err = s.database.QueryRow(`
			SELECT COUNT(*) FROM events
			WHERE actor = ?
			  AND type = 'pull_request'
			  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
		`, email, periodStart, periodEnd).Scan(&count)
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &models.MetricSnapshot{
		MetricName:  "pr_volume",
		EntityType:  entityType,
		EntityID:    entityID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       float64(count),
		Metadata:    fmt.Sprintf(`{"count":%d}`, count),
	}, nil
}

func (s *Service) computeCycleTime(entityType, entityID, periodStart, periodEnd string) (*models.MetricSnapshot, error) {
	// Cycle time calculation: average time from PR open to merge
	query := `
		SELECT
			AVG(CAST((julianday(merged_at) - julianday(created_at)) * 24 AS REAL)) as avg_hours
		FROM (
			SELECT
				json_extract(data, '$.created_at') as created_at,
				json_extract(data, '$.merged_at') as merged_at
			FROM events
			WHERE type = 'pull_request'
			  AND json_extract(data, '$.merged') = 1
			  AND DATE(timestamp) >= ? AND DATE(timestamp) <= ?
		)`

	args := []interface{}{periodStart, periodEnd}

	var avgCycleTime sql.NullFloat64
	err := s.database.QueryRow(query, args...).Scan(&avgCycleTime)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	value := 0.0
	if avgCycleTime.Valid {
		value = avgCycleTime.Float64
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"avg_hours": value,
		"avg_days":  value / HoursPerDay,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cycle time metadata: %w", err)
	}

	return &models.MetricSnapshot{
		MetricName:  "cycle_time",
		EntityType:  entityType,
		EntityID:    entityID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       value,
		Metadata:    string(metadata),
	}, nil
}
