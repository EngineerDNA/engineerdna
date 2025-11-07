package dashboards

import (
	"log"
	"time"
)

// Scheduler manages periodic metric snapshot computation
type Scheduler struct {
	service  *Service
	stopChan chan bool
	interval time.Duration
}

// NewScheduler creates a new dashboard metrics scheduler
func NewScheduler(service *Service) *Scheduler {
	return &Scheduler{
		service:  service,
		stopChan: make(chan bool),
		interval: 1 * time.Hour, // Compute snapshots every hour
	}
}

// Start begins periodic snapshot computation
func (s *Scheduler) Start() {
	log.Println("Starting dashboard metrics scheduler...")

	go func() {
		// Run immediately on start
		s.runComputation()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runComputation()
			case <-s.stopChan:
				log.Println("Dashboard metrics scheduler stopped")
				return
			}
		}
	}()
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping dashboard metrics scheduler...")
	s.stopChan <- true
}

// runComputation computes all metric snapshots
func (s *Scheduler) runComputation() {
	log.Println("Computing dashboard metric snapshots...")
	startTime := time.Now()

	if err := s.service.ComputeAllSnapshots(); err != nil {
		log.Printf("Error computing snapshots: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("Snapshot computation completed in %v", duration)
}
