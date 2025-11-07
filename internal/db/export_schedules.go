package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type ExportScheduleStore struct {
	db *sql.DB
}

func NewExportScheduleStore(db *sql.DB) *ExportScheduleStore {
	return &ExportScheduleStore{db: db}
}

func (s *ExportScheduleStore) Create(schedule *models.ExportSchedule) error {
	if schedule.ID == "" {
		schedule.ID = uuid.New().String()
	}
	if schedule.CreatedAt.IsZero() {
		schedule.CreatedAt = time.Now().UTC()
	}
	if schedule.UpdatedAt.IsZero() {
		schedule.UpdatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO export_schedules (
			id, plugin_name, frequency, day_of_week, time_of_day,
			enabled, last_run, next_run, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, schedule.ID, schedule.PluginName, schedule.Frequency, schedule.DayOfWeek,
		schedule.TimeOfDay, schedule.Enabled, schedule.LastRun, schedule.NextRun,
		schedule.CreatedAt, schedule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create export schedule: %w", err)
	}

	return nil
}

func (s *ExportScheduleStore) Get(id string) (*models.ExportSchedule, error) {
	var schedule models.ExportSchedule
	var dayOfWeek sql.NullInt64
	var lastRun sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, plugin_name, frequency, day_of_week, time_of_day,
		       enabled, last_run, next_run, created_at, updated_at
		FROM export_schedules
		WHERE id = ?
	`, id).Scan(
		&schedule.ID, &schedule.PluginName, &schedule.Frequency, &dayOfWeek,
		&schedule.TimeOfDay, &schedule.Enabled, &lastRun, &schedule.NextRun,
		&schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export schedule not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export schedule: %w", err)
	}

	if dayOfWeek.Valid {
		day := int(dayOfWeek.Int64)
		schedule.DayOfWeek = &day
	}
	if lastRun.Valid {
		schedule.LastRun = &lastRun.Time
	}

	return &schedule, nil
}

func (s *ExportScheduleStore) List(limit int) ([]*models.ExportSchedule, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > DefaultQueryLimit {
		limit = DefaultQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, plugin_name, frequency, day_of_week, time_of_day,
		       enabled, last_run, next_run, created_at, updated_at
		FROM export_schedules
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list export schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*models.ExportSchedule
	for rows.Next() {
		var schedule models.ExportSchedule
		var dayOfWeek sql.NullInt64
		var lastRun sql.NullTime

		err := rows.Scan(
			&schedule.ID, &schedule.PluginName, &schedule.Frequency, &dayOfWeek,
			&schedule.TimeOfDay, &schedule.Enabled, &lastRun, &schedule.NextRun,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export schedule: %w", err)
		}

		if dayOfWeek.Valid {
			day := int(dayOfWeek.Int64)
			schedule.DayOfWeek = &day
		}
		if lastRun.Valid {
			schedule.LastRun = &lastRun.Time
		}

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (s *ExportScheduleStore) Update(schedule *models.ExportSchedule) error {
	schedule.UpdatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		UPDATE export_schedules
		SET plugin_name = ?, frequency = ?, day_of_week = ?, time_of_day = ?,
		    enabled = ?, last_run = ?, next_run = ?, updated_at = ?
		WHERE id = ?
	`, schedule.PluginName, schedule.Frequency, schedule.DayOfWeek, schedule.TimeOfDay,
		schedule.Enabled, schedule.LastRun, schedule.NextRun, schedule.UpdatedAt, schedule.ID)

	if err != nil {
		return fmt.Errorf("failed to update export schedule: %w", err)
	}

	return nil
}

func (s *ExportScheduleStore) Delete(id string) error {
	result, err := s.db.Exec("DELETE FROM export_schedules WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete export schedule: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("export schedule not found: %s", id)
	}

	return nil
}

func (s *ExportScheduleStore) GetDue(now time.Time) ([]*models.ExportSchedule, error) {
	rows, err := s.db.Query(`
		SELECT id, plugin_name, frequency, day_of_week, time_of_day,
		       enabled, last_run, next_run, created_at, updated_at
		FROM export_schedules
		WHERE enabled = 1 AND next_run <= ?
		ORDER BY next_run ASC
	`, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get due export schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*models.ExportSchedule
	for rows.Next() {
		var schedule models.ExportSchedule
		var dayOfWeek sql.NullInt64
		var lastRun sql.NullTime

		err := rows.Scan(
			&schedule.ID, &schedule.PluginName, &schedule.Frequency, &dayOfWeek,
			&schedule.TimeOfDay, &schedule.Enabled, &lastRun, &schedule.NextRun,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export schedule: %w", err)
		}

		if dayOfWeek.Valid {
			day := int(dayOfWeek.Int64)
			schedule.DayOfWeek = &day
		}
		if lastRun.Valid {
			schedule.LastRun = &lastRun.Time
		}

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}
