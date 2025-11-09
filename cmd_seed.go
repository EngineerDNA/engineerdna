package main

import (
	"fmt"
	"log"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/identity"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/seed"
)

func cmdSeed() {
	fmt.Println("Seeding comprehensive demo data...")

	cfg := config.DefaultConfig()

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize stores and services
	eventStore := db.NewEventStore(database.DB)
	identityStore := db.NewIdentityStore(database.DB)
	identityService := identity.NewService(identityStore)
	teamStore := db.NewTeamStore(database.DB)
	scoringStore := db.NewScoringStore(database.DB)
	briefingStore := db.NewBriefingsStore(database.DB)
	planningStore := db.NewPlanningStore(database.DB)

	// Stores
	alertsStore := db.NewAlertsStore(database.DB)
	goalsStore := db.NewGoalsStore(database.DB)
	skillsStore := db.NewSkillsStore(database.DB)
	costROIStore := db.NewCostROIStore(database.DB)
	contextStore := db.NewContextStore(database.DB)
	forecastingStore := db.NewForecastingStore(database.DB)
	actionsStore := db.NewActionsStore(database.DB)
	attributeStore := db.NewAttributeStore(database.DB)
	metricStore := db.NewMetricStore(database.DB)

	now := time.Now().UTC()

	// Initialize seed data structure
	data := &seed.SeedData{
		Database:         database,
		EventStore:       eventStore,
		IdentityService:  identityService,
		TeamStore:        teamStore,
		ScoringStore:     scoringStore,
		BriefingStore:    briefingStore,
		PlanningStore:    planningStore,
		AlertsStore:      alertsStore,
		GoalsStore:       goalsStore,
		SkillsStore:      skillsStore,
		CostROIStore:     costROIStore,
		ContextStore:     contextStore,
		ForecastingStore: forecastingStore,
		ActionsStore:     actionsStore,
		AttributeStore:   attributeStore,
		MetricStore:      metricStore,
		Now:              now,
		TeamsByName:      make(map[string]*models.Team),
		EngineerMap:      make(map[string]seed.EngineerDef),
		RoleMap:          make(map[string]string),
	}

	// Execute seed steps using seed package functions
	if err := seed.SeedRoles(data); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}

	if err := seed.SeedTeams(data); err != nil {
		log.Fatalf("Failed to seed teams: %v", err)
	}

	if err := seed.SeedEngineers(data); err != nil {
		log.Fatalf("Failed to seed engineers: %v", err)
	}

	if err := seed.SeedEvents(data); err != nil {
		log.Fatalf("Failed to seed events: %v", err)
	}

	if err := seed.SeedMetrics(data); err != nil {
		log.Fatalf("Failed to seed metrics: %v", err)
	}

	if err := seed.SeedAttributes(data); err != nil {
		log.Fatalf("Failed to seed attributes: %v", err)
	}

	if err := seed.SeedScores(data); err != nil {
		log.Fatalf("Failed to seed scores: %v", err)
	}

	if err := seed.SeedBriefings(data); err != nil {
		log.Fatalf("Failed to seed briefings: %v", err)
	}

	if err := seed.SeedSprints(data); err != nil {
		log.Fatalf("Failed to seed sprints: %v", err)
	}

	// Feature Seed Data
	alertRuleIDs, alertInstanceIDs := seed.SeedAlerts(data)
	goalIDs, milestoneCount, progressLogCount := seed.SeedGoals(data)
	skillIDs, engineerSkillCount, evidenceCount := seed.SeedSkills(data)
	featureIDs, roiCount := seed.SeedCostROI(data)
	noteCount, responseCount := seed.SeedManagerContext(data)
	forecastCount, riskCount := seed.SeedForecasting(data, goalIDs)
	recommendationCount, actionCount := seed.SeedActions(data)

	// Summary
	fmt.Println("\n========================================")
	fmt.Println("Demo data seeding completed successfully!")
	fmt.Println("========================================")
	fmt.Println("\nV1.0 Core Features:")
	fmt.Printf("  - 4 roles (Junior, Mid, Senior, Staff)\n")
	fmt.Printf("  - 8 teams across 3-level hierarchy\n")
	fmt.Printf("  - 50 engineers with realistic profiles and distribution\n")
	fmt.Printf("  - Events over 12 months (~45,000 events)\n")
	fmt.Printf("  - Computed metrics (engineer, team, org levels)\n")
	fmt.Printf("  - Entity attributes (team sizes, costs, locations)\n")
	fmt.Printf("  - Performance scores\n")
	fmt.Printf("  - Team performance scores\n")
	fmt.Printf("  - 3 weekly briefings\n")
	fmt.Printf("  - Sprints and stories\n")
	fmt.Println("\nAdvanced Features:")
	fmt.Printf("  - %d alert rules and %d alert instances\n", len(alertRuleIDs), len(alertInstanceIDs))
	fmt.Printf("  - %d goals with %d milestones and %d progress logs\n", len(goalIDs), milestoneCount, progressLogCount)
	fmt.Printf("  - %d skills tracked across %d engineer skill assignments with %d evidence entries\n", len(skillIDs), engineerSkillCount, evidenceCount)
	fmt.Printf("  - %d features with cost analysis and %d ROI calculations\n", len(featureIDs), roiCount)
	fmt.Printf("  - %d manager notes, 1 sentiment survey with %d responses\n", noteCount, responseCount)
	fmt.Printf("  - %d forecasts and %d risk predictions\n", forecastCount, riskCount)
	fmt.Printf("  - %d recommendations and %d actions taken\n", recommendationCount, actionCount)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'engineerdna serve' to start the server")
	fmt.Println("2. Visit http://127.0.0.1:3847 to explore all features with demo data")
	fmt.Println("3. Run 'engineerdna reset' to remove all demo data")
}
