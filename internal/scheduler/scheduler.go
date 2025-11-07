package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/plugin"
	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

type Scheduler struct {
	db       *sql.DB
	store    *db.ExportScheduleStore
	executor *plugin.Executor
	ticker   *time.Ticker
	stopChan chan bool
	running  bool
}

func New(database *sql.DB, store *db.ExportScheduleStore, executor *plugin.Executor) *Scheduler {
	return &Scheduler{
		db:       database,
		store:    store,
		executor: executor,
		stopChan: make(chan bool),
		running:  false,
	}
}

func (s *Scheduler) Start() {
	if s.running {
		log.Println("Scheduler already running")
		return
	}

	s.running = true
	s.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		log.Println("Export scheduler started (checking every 1 minute)")
		for {
			select {
			case <-s.ticker.C:
				s.RunScheduledExports()
			case <-s.stopChan:
				log.Println("Export scheduler stopped")
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}

	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	s.running = false
}

func (s *Scheduler) RunScheduledExports() {
	now := time.Now().UTC()

	schedules, err := s.store.GetDue(now)
	if err != nil {
		log.Printf("Failed to get due schedules: %v", err)
		return
	}

	if len(schedules) == 0 {
		return
	}

	log.Printf("Found %d due schedules to execute", len(schedules))

	for _, schedule := range schedules {
		go func(sched *models.ExportSchedule) {
			if err := s.ExecuteSchedule(sched); err != nil {
				log.Printf("Failed to execute schedule %s (%s): %v", sched.ID, sched.PluginName, err)
			} else {
				log.Printf("Successfully executed schedule %s (%s)", sched.ID, sched.PluginName)
			}
		}(schedule)
	}
}

func (s *Scheduler) ExecuteSchedule(schedule *models.ExportSchedule) error {
	now := time.Now().UTC()

	log.Printf("Executing export schedule %s for plugin %s", schedule.ID, schedule.PluginName)

	exportParams := sdk.ExportParams{
		DataType: "events",
		Data:     map[string]interface{}{},
	}

	_, err := s.executor.ExportToDestination(schedule.PluginName, exportParams, false)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	schedule.LastRun = &now
	schedule.NextRun = CalculateNextRun(schedule.Frequency, schedule.DayOfWeek, schedule.TimeOfDay, now)

	if err := s.store.Update(schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}

func CalculateNextRun(frequency string, dayOfWeek *int, timeOfDay string, from time.Time) time.Time {
	hour, minute := parseTimeOfDay(timeOfDay)
	fromDate := from.UTC()

	switch frequency {
	case "daily":
		next := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), hour, minute, 0, 0, time.UTC)
		if next.Before(fromDate) || next.Equal(fromDate) {
			next = next.AddDate(0, 0, 1)
		}
		return next

	case "weekly":
		if dayOfWeek == nil {
			log.Printf("Warning: weekly frequency requires day_of_week, defaulting to current day")
			dow := int(fromDate.Weekday())
			dayOfWeek = &dow
		}

		next := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), hour, minute, 0, 0, time.UTC)

		currentDow := int(next.Weekday())
		targetDow := *dayOfWeek

		daysUntilTarget := (targetDow - currentDow + 7) % 7
		if daysUntilTarget == 0 {
			if next.Before(fromDate) || next.Equal(fromDate) {
				daysUntilTarget = 7
			}
		}

		next = next.AddDate(0, 0, daysUntilTarget)
		return next

	case "monthly":
		next := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), hour, minute, 0, 0, time.UTC)
		if next.Before(fromDate) || next.Equal(fromDate) {
			next = next.AddDate(0, 1, 0)
		}

		for next.Day() != fromDate.Day() {
			next = next.AddDate(0, 0, -1)
		}

		return next

	default:
		log.Printf("Warning: unknown frequency %s, defaulting to daily", frequency)
		return CalculateNextRun("daily", nil, timeOfDay, from)
	}
}

func parseTimeOfDay(timeOfDay string) (hour, minute int) {
	parsed, err := time.Parse("15:04", timeOfDay)
	if err != nil {
		log.Printf("Warning: failed to parse time_of_day %s, defaulting to 09:00", timeOfDay)
		return 9, 0
	}
	return parsed.Hour(), parsed.Minute()
}
