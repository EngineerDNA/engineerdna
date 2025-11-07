package alerts

import (
	"database/sql"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
)

// Scheduler manages periodic alert evaluation and delivery
type Scheduler struct {
	evaluator *Evaluator
	delivery  *Delivery
	stopChan  chan bool
	interval  time.Duration
}

// NewScheduler creates a new alert scheduler
func NewScheduler(database *sql.DB, alertsStore *db.AlertsStore, eventStore *db.EventStore, scoringStore *db.ScoringStore, teamStore *db.TeamStore) *Scheduler {
	evaluator := NewEvaluator(database, alertsStore, eventStore, scoringStore, teamStore)
	delivery := NewDelivery(alertsStore)

	return &Scheduler{
		evaluator: evaluator,
		delivery:  delivery,
		stopChan:  make(chan bool),
		interval:  1 * time.Minute, // Evaluate every minute
	}
}

// Start begins periodic alert evaluation and delivery
func (s *Scheduler) Start() {
	log.Println("Starting alert scheduler...")

	go func() {
		// Run immediately on start
		s.runEvaluation()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runEvaluation()
			case <-s.stopChan:
				log.Println("Alert scheduler stopped")
				return
			}
		}
	}()
}

// Stop halts the alert scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping alert scheduler...")
	s.stopChan <- true
}

// runEvaluation runs alert evaluation and delivery
func (s *Scheduler) runEvaluation() {
	// Evaluate all alert rules
	if err := s.evaluator.EvaluateAll(); err != nil {
		log.Printf("Error evaluating alerts: %v", err)
	}

	// Deliver pending alerts
	if err := s.delivery.DeliverPendingAlerts(); err != nil {
		log.Printf("Error delivering alerts: %v", err)
	}
}
