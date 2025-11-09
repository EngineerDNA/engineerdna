package dashboards

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

const (
	// HoursPerDay is the number of hours in a day for time conversions
	HoursPerDay = 24
)

// Service handles dashboard metric snapshot computation
type Service struct {
	dashboardStore *db.DashboardStore
	scoringStore   *db.ScoringStore
	eventStore     *db.EventStore
	alertsStore    *db.AlertsStore
	goalsStore     *db.GoalsStore
	teamStore      *db.TeamStore
}

// NewService creates a new dashboard service
func NewService(
	dashboardStore *db.DashboardStore,
	scoringStore *db.ScoringStore,
	eventStore *db.EventStore,
	alertsStore *db.AlertsStore,
	goalsStore *db.GoalsStore,
	teamStore *db.TeamStore,
) *Service {
	return &Service{
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
	now := time.Now().UTC()

	// Compute snapshots for the last 30 days to capture historical data
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")

		// Compute org-wide snapshots
		if err := s.computeOrgSnapshots(date, date); err != nil {
			return fmt.Errorf("failed to compute org snapshots for %s: %w", date, err)
		}

		// Compute team snapshots
		if err := s.computeTeamSnapshots(date, date); err != nil {
			return fmt.Errorf("failed to compute team snapshots for %s: %w", date, err)
		}

		// Compute engineer snapshots
		if err := s.computeEngineerSnapshots(date, date); err != nil {
			return fmt.Errorf("failed to compute engineer snapshots for %s: %w", date, err)
		}
	}

	return nil
}

// computeOrgSnapshots computes organization-wide metric snapshots
func (s *Service) computeOrgSnapshots(periodStart, periodEnd string) error {
	snapshots := []*models.MetricSnapshot{}

	// PR Volume (org-wide)
	prCount, err := s.dashboardStore.CountPRsByDate(periodStart, periodEnd)
	if err != nil {
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
	alertCount, err := s.dashboardStore.CountActiveAlerts(periodStart, periodEnd+" 23:59:59")
	if err != nil {
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
	goalCount, err := s.dashboardStore.CountActiveGoals()
	if err != nil {
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

	// Org-level Team Score (average of all team scores)
	teamMetrics, err := s.dashboardStore.GetTeamMetricsForDate(periodStart, periodEnd+" 23:59:59")
	if err != nil {
		return fmt.Errorf("failed to get team metrics for org score: %w", err)
	}

	if len(teamMetrics) > 0 {
		var totalScore float64
		var teamCount int
		for _, tm := range teamMetrics {
			if tm.TeamScore > 0 {
				totalScore += tm.TeamScore
				teamCount++
			}
		}
		if teamCount > 0 {
			avgScore := totalScore / float64(teamCount)
			snapshots = append(snapshots, &models.MetricSnapshot{
				MetricName:  "team_score",
				EntityType:  "org",
				EntityID:    "",
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Value:       avgScore,
			})
		}
	}

	// Save all snapshots
	return s.dashboardStore.CreateMetricSnapshotsBatch(snapshots)
}

// computeTeamSnapshots computes team-level metric snapshots
func (s *Service) computeTeamSnapshots(periodStart, periodEnd string) error {
	// Get all team metrics using repository method
	teamMetrics, err := s.dashboardStore.GetTeamMetricsForDate(periodStart, periodEnd+" 23:59:59")
	if err != nil {
		return fmt.Errorf("failed to get team metrics: %w", err)
	}

	snapshots := []*models.MetricSnapshot{}
	for _, tm := range teamMetrics {
		// Create snapshots for each metric
		if tm.TeamScore > 0 {
			snapshots = append(snapshots, &models.MetricSnapshot{
				MetricName:  "team_score",
				EntityType:  "team",
				EntityID:    tm.TeamID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Value:       tm.TeamScore,
			})
		}
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "pr_volume",
			EntityType:  "team",
			EntityID:    tm.TeamID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       tm.PRCount,
		})
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "active_alerts",
			EntityType:  "team",
			EntityID:    tm.TeamID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       tm.AlertCount,
		})
	}

	if len(snapshots) > 0 {
		return s.dashboardStore.CreateMetricSnapshotsBatch(snapshots)
	}
	return nil
}

// computeEngineerSnapshots computes engineer-level metric snapshots
func (s *Service) computeEngineerSnapshots(periodStart, periodEnd string) error {
	// Get all engineer metrics using repository method
	engineerMetrics, err := s.dashboardStore.GetEngineerMetricsForDate(periodStart, periodEnd, db.MaxQueryLimit)
	if err != nil {
		return fmt.Errorf("failed to get engineer metrics: %w", err)
	}

	snapshots := []*models.MetricSnapshot{}
	for _, em := range engineerMetrics {
		// Create snapshots for each metric
		if em.EngineerScore > 0 {
			snapshots = append(snapshots, &models.MetricSnapshot{
				MetricName:  "engineer_score",
				EntityType:  "engineer",
				EntityID:    em.EngineerID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Value:       em.EngineerScore,
			})
		}
		snapshots = append(snapshots, &models.MetricSnapshot{
			MetricName:  "pr_count",
			EntityType:  "engineer",
			EntityID:    em.EngineerID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Value:       em.PRCount,
		})
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
	score, err := s.dashboardStore.GetLatestTeamScore(teamID)
	if err != nil {
		return nil, err
	}

	if score == 0 {
		return nil, nil
	}

	return &models.MetricSnapshot{
		MetricName:  "team_score",
		EntityType:  "team",
		EntityID:    teamID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Value:       score,
	}, nil
}

func (s *Service) computePRVolume(entityType, entityID, periodStart, periodEnd string) (*models.MetricSnapshot, error) {
	var count int
	var err error

	switch entityType {
	case "org":
		count, err = s.dashboardStore.CountOrgPRsInRange(periodStart, periodEnd)
	case "team":
		count, err = s.dashboardStore.CountTeamPRsInRange(entityID, periodStart, periodEnd)
	case "engineer":
		var email string
		email, err = s.dashboardStore.GetEngineerEmail(entityID)
		if err != nil {
			return nil, err
		}
		count, err = s.dashboardStore.CountEngineerPRsInRange(email, periodStart, periodEnd)
	}

	if err != nil {
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
	value, err := s.dashboardStore.CalculateAvgCycleTime(periodStart, periodEnd)
	if err != nil {
		return nil, err
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
