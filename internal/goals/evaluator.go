package goals

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
)

// Evaluator handles periodic goal evaluation
type Evaluator struct {
	database *sql.DB
	store    *db.GoalsStore
	service  *Service
	stopChan chan bool
}

// NewEvaluator creates a new goal evaluator
func NewEvaluator(database *sql.DB, store *db.GoalsStore, service *Service) *Evaluator {
	return &Evaluator{
		database: database,
		store:    store,
		service:  service,
		stopChan: make(chan bool),
	}
}

// Start begins periodic goal evaluation
func (e *Evaluator) Start() {
	log.Println("Starting goal evaluator...")

	go func() {
		// Run immediately on start
		e.runEvaluation()

		// Schedule daily evaluation at 2am UTC
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().UTC()
				if now.Hour() == 2 {
					e.runEvaluation()
				}
			case <-e.stopChan:
				log.Println("Goal evaluator stopped")
				return
			}
		}
	}()
}

// Stop halts the goal evaluator
func (e *Evaluator) Stop() {
	log.Println("Stopping goal evaluator...")
	e.stopChan <- true
}

// runEvaluation evaluates all active goals
func (e *Evaluator) runEvaluation() {
	log.Println("Running goal evaluation...")

	// Get all active goals
	goals, _, err := e.store.ListGoals("", "", "active", "", 0, 0)
	if err != nil {
		log.Printf("Error listing goals: %v", err)
		return
	}

	log.Printf("Evaluating %d active goals", len(goals))

	for _, goal := range goals {
		if err := e.evaluateGoal(goal.ID); err != nil {
			log.Printf("Error evaluating goal %s: %v", goal.ID, err)
		}
	}

	log.Println("Goal evaluation completed")
}

// evaluateGoal evaluates a single goal
func (e *Evaluator) evaluateGoal(goalID string) error {
	// Update metrics for metric-based goals
	if err := e.service.UpdateGoalMetrics(goalID); err != nil {
		return fmt.Errorf("failed to update metrics: %w", err)
	}

	// Update progress from milestones
	if err := e.service.UpdateGoalProgress(goalID); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	// Check dependencies
	if err := e.checkDependencies(goalID); err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	return nil
}

// checkDependencies checks if any dependencies are blocking the goal
func (e *Evaluator) checkDependencies(goalID string) error {
	dependencies, _, err := e.store.GetDependencies(goalID, 100, 0)
	if err != nil {
		return err
	}

	for _, dep := range dependencies {
		if dep.Status == "active" && dep.DependencyType == "blocks" {
			// Check if the blocking goal is completed
			blockingGoal, err := e.store.GetGoal(dep.DependsOnGoalID)
			if err != nil {
				continue
			}

			if blockingGoal != nil && blockingGoal.Status == "completed" {
				// Resolve the dependency
				dep.Status = "resolved"
				if err := e.updateDependency(dep.ID, "resolved"); err != nil {
					log.Printf("Error updating dependency: %v", err)
				}
			}
		}
	}

	return nil
}

// updateDependency updates a dependency status
func (e *Evaluator) updateDependency(depID, status string) error {
	_, err := e.database.Exec(`
		UPDATE goal_dependencies
		SET status = ?
		WHERE id = ?
	`, status, depID)

	return err
}

// EvaluateAll evaluates all goals (called by scheduler)
func (e *Evaluator) EvaluateAll() error {
	e.runEvaluation()
	return nil
}
