package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/engineerdna/engineerdna/internal/ai"
	"github.com/engineerdna/engineerdna/internal/alerts"
	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/api"
	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/context"
	"github.com/engineerdna/engineerdna/internal/cost"
	"github.com/engineerdna/engineerdna/internal/dashboards"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/forecasting"
	"github.com/engineerdna/engineerdna/internal/goals"
	"github.com/engineerdna/engineerdna/internal/identity"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/planning"
	"github.com/engineerdna/engineerdna/internal/plugin"
	"github.com/engineerdna/engineerdna/internal/promotion"
	"github.com/engineerdna/engineerdna/internal/roi"
	"github.com/engineerdna/engineerdna/internal/scheduler"
	"github.com/engineerdna/engineerdna/internal/scoring"
	"github.com/engineerdna/engineerdna/internal/sentiment"
	"github.com/engineerdna/engineerdna/internal/skills"
	"github.com/engineerdna/engineerdna/internal/team"
	"github.com/google/uuid"
)

//go:embed frontend/dist
var frontendFiles embed.FS

var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmdInit()
	case "serve":
		cmdServe()
	case "sync":
		cmdSync()
	case "seed":
		cmdSeed()
	case "reset":
		cmdReset()
	case "backfill":
		cmdBackfill()
	case "version":
		cmdVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func cmdInit() {
	fmt.Println("Initializing EngineerDNA...")

	cfg := config.DefaultConfig()
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize encryption key
	keyStore, err := config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}
	_ = keyStore

	fmt.Println("EngineerDNA initialized successfully")
	fmt.Printf("Database: %s\n", cfg.Database.Path)
	fmt.Printf("Plugins directory: %s\n", cfg.Plugins.Directory)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'engineerdna serve' to start the server")
	fmt.Println("2. Open http://localhost:3847 in your browser")
}

func cmdServe() {
	fmt.Println("Starting EngineerDNA server...")

	cfg := config.DefaultConfig()
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize stores
	eventStore := db.NewEventStore(database.DB)
	pluginStore := db.NewPluginStore(database.DB)
	anonStore := db.NewAnonymizationStore(database.DB)
	auditStore := db.NewAuditStore(database.DB)
	identityStore := db.NewIdentityStore(database.DB)
	configStore := db.NewConfigStore(database.DB)
	scheduleStore := db.NewExportScheduleStore(database.DB)
	exportsStore := db.NewExportsStore(database.DB)
	scoringStore := db.NewScoringStore(database.DB)
	teamStore := db.NewTeamStore(database.DB)
	briefingStore := db.NewBriefingsStore(database.DB)
	planningStore := db.NewPlanningStore(database.DB)
	settingsStore := db.NewSettingsStore(database.DB)
	alertsStore := db.NewAlertsStore(database.DB)
	goalsStore := db.NewGoalsStore(database.DB)
	skillsStore := db.NewSkillsStore(database.DB)
	costROIStore := db.NewCostROIStore(database.DB)
	contextStore := db.NewContextStore(database.DB)
	forecastingStore := db.NewForecastingStore(database.DB)
	actionsStore := db.NewActionsStore(database.DB)
	dashboardStore := db.NewDashboardStore(database.DB)
	engineerStore := db.NewEngineerStore(database.DB)
	healthStore := db.NewHealthStore(database.DB)

	// Initialize services
	keyStore, err := config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}

	anonService := anonymization.NewService(anonStore)
	identityService := identity.NewService(identityStore)
	scoringService := scoring.NewScoringService(database.DB)
	teamService := team.NewService(database.DB)
	promotionService := promotion.NewService(database.DB)
	planningService := planning.NewService(planningStore)
	goalsService := goals.NewService(goalsStore, scoringStore)
	skillsService := skills.NewService(skillsStore, eventStore)
	costService := cost.NewService(costROIStore, teamStore)
	roiService := roi.NewService(costROIStore)
	costAnalyzer := cost.NewAnalyzer(costROIStore, eventStore, costService)
	contextService := context.NewService(contextStore, identityStore)
	sentimentService := sentiment.NewService(contextStore, teamStore)
	sentimentAnalyzer := sentiment.NewAnalyzer(contextStore, eventStore)
	forecastingService := forecasting.NewForecastingService(database.DB, forecastingStore, planningStore, scoringStore, goalsStore)
	dashboardService := dashboards.NewService(database.DB, dashboardStore, scoringStore, eventStore, alertsStore, goalsStore, teamStore)

	// Initialize plugin system
	pluginDirs := cfg.GetPluginDirs()
	loader := plugin.NewLoader(pluginDirs)
	executor := plugin.NewExecutor(loader, anonService, anonStore, pluginStore, auditStore, keyStore)

	// Initialize AI briefing service
	briefingService := ai.NewBriefingService(database.DB, briefingStore, eventStore, teamStore, anonService, loader)

	// Discover plugins
	plugins, err := loader.DiscoverPlugins()
	if err != nil {
		log.Printf("Warning: Failed to discover plugins: %v", err)
	} else {
		fmt.Printf("Discovered %d plugins:\n", len(plugins))
		for name := range plugins {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Initialize scheduler
	sched := scheduler.New(database.DB, scheduleStore, executor)
	sched.Start()
	defer sched.Stop()

	// Initialize alert system
	alertScheduler := alerts.NewScheduler(database.DB, alertsStore, eventStore, scoringStore, teamStore)
	alertScheduler.Start()
	defer alertScheduler.Stop()

	// Create default alert rules (only on first run)
	if err := alerts.CreateDefaultAlertRules(alertsStore); err != nil {
		log.Printf("Warning: Failed to create default alert rules: %v", err)
	}

	// Create default skills taxonomy (only on first run)
	if err := skills.CreateDefaultSkills(skillsStore); err != nil {
		log.Printf("Warning: Failed to create default skills: %v", err)
	}

	// Create default cost configuration (only on first run)
	if err := cost.CreateDefaultCostConfiguration(costROIStore); err != nil {
		log.Printf("Warning: Failed to create default cost configuration: %v", err)
	}

	// Create default dashboard templates (only on first run)
	if err := dashboards.CreateDefaultTemplates(dashboardStore); err != nil {
		log.Printf("Warning: Failed to create default dashboard templates: %v", err)
	}

	// Prepare frontend filesystem
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		// If fs.Sub fails, use the full embedded FS
		log.Printf("Warning: Failed to create frontend subFS: %v", err)
		frontendFS = frontendFiles
	}

	// Initialize goal evaluator
	goalEvaluator := goals.NewEvaluator(database.DB, goalsStore, goalsService)
	goalEvaluator.Start()
	defer goalEvaluator.Stop()

	// Initialize dashboard metrics scheduler
	dashboardScheduler := dashboards.NewScheduler(dashboardService)
	dashboardScheduler.Start()
	defer dashboardScheduler.Stop()

	// Initialize API server with frontend
	server := api.NewServer(database.DB, eventStore, pluginStore, anonStore, auditStore, identityStore, configStore, scheduleStore, exportsStore, scoringStore, teamStore, briefingStore, planningStore, settingsStore, alertsStore, goalsStore, skillsStore, costROIStore, contextStore, forecastingStore, actionsStore, dashboardStore, engineerStore, healthStore, executor, anonService, identityService, scoringService, teamService, promotionService, briefingService, planningService, goalsService, skillsService, costService, roiService, costAnalyzer, contextService, sentimentService, sentimentAnalyzer, forecastingService, sched, cfg, frontendFS, Version, log.Default())
	server.SetupRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("\nServer running at http://%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func cmdSync() {
	fmt.Println("Syncing plugins...")

	cfg := config.DefaultConfig()

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize stores
	eventStore := db.NewEventStore(database.DB)
	pluginStore := db.NewPluginStore(database.DB)
	anonStore := db.NewAnonymizationStore(database.DB)
	auditStore := db.NewAuditStore(database.DB)
	identityStore := db.NewIdentityStore(database.DB)

	// Initialize services
	keyStore, err := config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}

	anonService := anonymization.NewService(anonStore)
	identityService := identity.NewService(identityStore)

	// Initialize plugin system
	pluginDirs := cfg.GetPluginDirs()
	loader := plugin.NewLoader(pluginDirs)
	executor := plugin.NewExecutor(loader, anonService, anonStore, pluginStore, auditStore, keyStore)
	_ = identityService

	// Get all configured source plugins
	configs, err := pluginStore.List(1000)
	if err != nil {
		log.Fatalf("Failed to list plugins: %v", err)
	}

	for _, pluginCfg := range configs {
		if !pluginCfg.Enabled {
			continue
		}

		// Check if it's a source plugin
		info, err := executor.GetPluginInfo(pluginCfg.Name)
		if err != nil {
			log.Printf("Failed to get plugin info for %s: %v", pluginCfg.Name, err)
			continue
		}

		if info.Type != "source" {
			continue
		}

		fmt.Printf("Syncing %s...\n", pluginCfg.Name)

		// Get last sync time
		since := time.Now().UTC().Add(-24 * time.Hour) // Default to 24 hours ago
		if pluginCfg.LastSync != nil {
			since = *pluginCfg.LastSync
		}

		// Sync events
		events, err := executor.SyncSourcePlugin(pluginCfg.Name, since)
		if err != nil {
			log.Printf("Failed to sync %s: %v", pluginCfg.Name, err)
			continue
		}

		// Save events atomically in a transaction
		if err := eventStore.CreateBatch(events); err != nil {
			log.Printf("Failed to save events for %s: %v", pluginCfg.Name, err)
			continue
		}

		fmt.Printf("%s: synced %d events\n", pluginCfg.Name, len(events))
	}

	fmt.Println("\nSync completed")
}

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

	// PDR-7 stores
	alertsStore := db.NewAlertsStore(database.DB)
	goalsStore := db.NewGoalsStore(database.DB)
	skillsStore := db.NewSkillsStore(database.DB)
	costROIStore := db.NewCostROIStore(database.DB)
	contextStore := db.NewContextStore(database.DB)
	forecastingStore := db.NewForecastingStore(database.DB)
	actionsStore := db.NewActionsStore(database.DB)

	now := time.Now().UTC()

	// 1. Create Roles
	fmt.Println("\n1. Creating roles...")
	roles := []struct {
		name         string
		targetScore  int
		expectations map[string]float64
	}{
		{
			name:        "Junior Engineer",
			targetScore: 70,
			expectations: map[string]float64{
				"throughput":    5.0,
				"quality":       85.0,
				"speed":         3.0,
				"collaboration": 2.0,
			},
		},
		{
			name:        "Mid-Level Engineer",
			targetScore: 90,
			expectations: map[string]float64{
				"throughput":    8.0,
				"quality":       90.0,
				"speed":         2.5,
				"collaboration": 4.0,
			},
		},
		{
			name:        "Senior Engineer",
			targetScore: 100,
			expectations: map[string]float64{
				"throughput":    10.0,
				"quality":       92.0,
				"speed":         2.0,
				"collaboration": 6.0,
			},
		},
		{
			name:        "Staff Engineer",
			targetScore: 110,
			expectations: map[string]float64{
				"throughput":    12.0,
				"quality":       95.0,
				"speed":         1.5,
				"collaboration": 8.0,
			},
		},
	}

	roleMap := make(map[string]string)
	for _, r := range roles {
		role, err := scoringStore.CreateRole(r.name, r.targetScore, r.expectations)
		if err != nil {
			log.Fatalf("Failed to create role %s: %v", r.name, err)
		}
		roleMap[r.name] = role.ID
		fmt.Printf("  Created role: %s (target: %d)\n", r.name, r.targetScore)
	}

	// 2. Create Scoring Weights
	fmt.Println("\n2. Creating scoring weights...")
	_, err = scoringStore.UpdateWeights(30.0, 25.0, 20.0, 15.0, 10.0)
	if err != nil {
		log.Fatalf("Failed to create scoring weights: %v", err)
	}
	fmt.Println("  Created balanced scoring weights (30/25/20/15/10)")

	// 3. Create Teams with hierarchy
	fmt.Println("\n3. Creating team hierarchy...")

	// Engineering parent team
	engineeringTeam, err := teamStore.CreateTeam("Engineering", nil, nil)
	if err != nil {
		log.Fatalf("Failed to create Engineering team: %v", err)
	}
	fmt.Printf("  Created team: %s\n", engineeringTeam.Name)

	// Child teams
	backendTeam, err := teamStore.CreateTeam("Backend", &engineeringTeam.ID, nil)
	if err != nil {
		log.Fatalf("Failed to create Backend team: %v", err)
	}
	fmt.Printf("  Created team: %s (parent: Engineering)\n", backendTeam.Name)

	frontendTeam, err := teamStore.CreateTeam("Frontend", &engineeringTeam.ID, nil)
	if err != nil {
		log.Fatalf("Failed to create Frontend team: %v", err)
	}
	fmt.Printf("  Created team: %s (parent: Engineering)\n", frontendTeam.Name)

	platformTeam, err := teamStore.CreateTeam("Platform", &engineeringTeam.ID, nil)
	if err != nil {
		log.Fatalf("Failed to create Platform team: %v", err)
	}
	fmt.Printf("  Created team: %s (parent: Engineering)\n", platformTeam.Name)

	// 4. Create Engineers
	fmt.Println("\n4. Creating engineers...")
	type engineerDef struct {
		name        string
		email       string
		manager     string
		role        string
		teamID      string
		identifiers map[string]string
	}

	engineers := []engineerDef{
		{
			name:    "Sarah Chen",
			email:   "sarah.chen@example.com",
			manager: "",
			role:    "Staff Engineer",
			teamID:  engineeringTeam.ID,
			identifiers: map[string]string{
				"github": "schen",
				"jira":   "sarah.chen@example.com",
			},
		},
		{
			name:    "Marcus Johnson",
			email:   "marcus.johnson@example.com",
			manager: "Sarah Chen",
			role:    "Senior Engineer",
			teamID:  backendTeam.ID,
			identifiers: map[string]string{
				"github": "mjohnson",
				"jira":   "marcus.johnson@example.com",
			},
		},
		{
			name:    "Elena Rodriguez",
			email:   "elena.rodriguez@example.com",
			manager: "Sarah Chen",
			role:    "Senior Engineer",
			teamID:  backendTeam.ID,
			identifiers: map[string]string{
				"github": "erodriguez",
				"jira":   "elena.rodriguez@example.com",
			},
		},
		{
			name:    "David Kim",
			email:   "david.kim@example.com",
			manager: "Sarah Chen",
			role:    "Mid-Level Engineer",
			teamID:  backendTeam.ID,
			identifiers: map[string]string{
				"github": "dkim",
				"jira":   "david.kim@example.com",
			},
		},
		{
			name:    "Priya Sharma",
			email:   "priya.sharma@example.com",
			manager: "Sarah Chen",
			role:    "Senior Engineer",
			teamID:  frontendTeam.ID,
			identifiers: map[string]string{
				"github": "psharma",
				"jira":   "priya.sharma@example.com",
			},
		},
		{
			name:    "Alex Thompson",
			email:   "alex.thompson@example.com",
			manager: "Sarah Chen",
			role:    "Mid-Level Engineer",
			teamID:  frontendTeam.ID,
			identifiers: map[string]string{
				"github": "athompson",
				"jira":   "alex.thompson@example.com",
			},
		},
		{
			name:    "Maya Patel",
			email:   "maya.patel@example.com",
			manager: "Sarah Chen",
			role:    "Junior Engineer",
			teamID:  frontendTeam.ID,
			identifiers: map[string]string{
				"github": "mpatel",
				"jira":   "maya.patel@example.com",
			},
		},
		{
			name:    "James Wilson",
			email:   "james.wilson@example.com",
			manager: "Sarah Chen",
			role:    "Senior Engineer",
			teamID:  platformTeam.ID,
			identifiers: map[string]string{
				"github": "jwilson",
				"jira":   "james.wilson@example.com",
			},
		},
		{
			name:    "Li Wei",
			email:   "li.wei@example.com",
			manager: "Sarah Chen",
			role:    "Mid-Level Engineer",
			teamID:  platformTeam.ID,
			identifiers: map[string]string{
				"github": "lwei",
				"jira":   "li.wei@example.com",
			},
		},
		{
			name:    "Nina Okafor",
			email:   "nina.okafor@example.com",
			manager: "Sarah Chen",
			role:    "Junior Engineer",
			teamID:  platformTeam.ID,
			identifiers: map[string]string{
				"github": "nokafor",
				"jira":   "nina.okafor@example.com",
			},
		},
	}

	engineerIDs := make([]string, len(engineers))
	engineerMap := make(map[string]engineerDef)
	for i, eng := range engineers {
		engineerID, err := identityService.CreateEngineer(eng.name, eng.email, eng.manager, eng.identifiers)
		if err != nil {
			log.Fatalf("Failed to create engineer %s: %v", eng.name, err)
		}
		engineerIDs[i] = engineerID
		engineerMap[engineerID] = eng

		// Assign role
		roleID := roleMap[eng.role]
		_, err = database.DB.Exec("UPDATE engineers SET role_id = ? WHERE id = ?", roleID, engineerID)
		if err != nil {
			log.Fatalf("Failed to assign role to %s: %v", eng.name, err)
		}

		// Add to team
		_, err = teamStore.AddTeamMember(eng.teamID, engineerID, eng.role)
		if err != nil {
			log.Fatalf("Failed to add %s to team: %v", eng.name, err)
		}

		fmt.Printf("  Created: %s (%s) -> %s team\n", eng.name, eng.role, getTeamName(eng.teamID, engineeringTeam, backendTeam, frontendTeam, platformTeam))
	}

	// Update team manager IDs
	sarahID := engineerIDs[0] // Sarah Chen is the manager
	err = teamStore.UpdateTeam(engineeringTeam.ID, engineeringTeam.Name, nil, &sarahID)
	if err != nil {
		log.Printf("Warning: Failed to set team manager: %v", err)
	}

	// 5. Generate Events (250 events over 8 weeks)
	fmt.Println("\n5. Generating events...")
	events := []*models.Event{}

	eventCounter := 1
	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := now.AddDate(0, 0, -7*weekOffset)

		// More recent weeks have more events
		eventsThisWeek := 30 + (8-weekOffset)*2

		for i := 0; i < eventsThisWeek; i++ {
			// Pick random engineer
			engIndex := i % len(engineerIDs)
			engineerID := engineerIDs[engIndex]
			eng := engineerMap[engineerID]

			// Pick event type based on weights
			typeIndex := i % 10
			var eventType, source string
			if typeIndex < 4 {
				eventType = "pull_request"
				source = "github"
			} else if typeIndex < 7 {
				eventType = "commit"
				source = "github"
			} else if typeIndex < 9 {
				eventType = "issue"
				source = "jira"
			} else {
				eventType = "code_review"
				source = "github"
			}

			// Random day within the week
			dayOffset := i % 7
			timestamp := weekStart.AddDate(0, 0, -dayOffset)

			// Get identifier for this source
			identifier := eng.identifiers[source]
			if identifier == "" {
				identifier = eng.email
			}

			event := &models.Event{
				Type:       eventType,
				Source:     source,
				SourceID:   fmt.Sprintf("demo-%s-%d", eventType, eventCounter),
				Timestamp:  timestamp,
				Actor:      identifier,
				EngineerID: engineerID,
				Data: map[string]interface{}{
					"title":  fmt.Sprintf("Demo %s #%d", eventType, eventCounter),
					"status": "completed",
				},
				Anonymized: false,
			}

			events = append(events, event)
			eventCounter++
		}
	}

	// Save events in batch
	if err := eventStore.CreateBatch(events); err != nil {
		log.Fatalf("Failed to save demo events: %v", err)
	}
	fmt.Printf("  Created %d events over 8 weeks\n", len(events))

	// 6. Generate Performance Scores
	fmt.Println("\n6. Generating performance scores...")
	scoreCounter := 0
	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := now.AddDate(0, 0, -7*weekOffset).Truncate(24 * time.Hour)

		for i, engineerID := range engineerIDs {
			eng := engineerMap[engineerID]

			// Base score from role
			var baseScore float64
			if eng.role == "Junior Engineer" {
				baseScore = 70
			} else if eng.role == "Mid-Level Engineer" {
				baseScore = 90
			} else if eng.role == "Senior Engineer" {
				baseScore = 100
			} else if eng.role == "Staff Engineer" {
				baseScore = 110
			}

			// Add realistic variation
			// Some engineers improving, some declining, some stable
			var trendFactor float64
			if i%3 == 0 {
				// Improving trend
				trendFactor = float64(weekOffset) * 1.5
			} else if i%3 == 1 {
				// Declining trend
				trendFactor = -float64(weekOffset) * 1.0
			} else {
				// Stable with noise
				trendFactor = float64(weekOffset%2)*2 - 1
			}

			totalScore := baseScore + trendFactor + float64((i*7)%10-5)

			// Component scores (realistic distribution)
			throughputScore := totalScore * 0.3
			qualityScore := totalScore * 0.25
			speedScore := totalScore * 0.2
			collaborationScore := totalScore * 0.15
			impactScore := totalScore * 0.1

			scoreID := uuid.New().String()
			_, err := database.DB.Exec(`
				INSERT INTO performance_scores (
					id, engineer_id, week_start, total_score,
					throughput_score, quality_score, speed_score,
					collaboration_score, impact_score, raw_metrics, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, scoreID, engineerID, weekStart, totalScore,
				throughputScore, qualityScore, speedScore,
				collaborationScore, impactScore, "{}", now)

			if err != nil {
				log.Printf("Warning: Failed to create performance score: %v", err)
			}
			scoreCounter++
		}
	}
	fmt.Printf("  Created %d performance scores (8 weeks x %d engineers)\n", scoreCounter, len(engineerIDs))

	// 6b. Aggregate Team Performance Scores
	fmt.Println("\n6b. Aggregating team performance scores...")
	teamScoreCounter := 0

	// Get all child teams (not parent Engineering team)
	childTeams := []struct {
		id   string
		name string
	}{
		{backendTeam.ID, "Backend"},
		{frontendTeam.ID, "Frontend"},
		{platformTeam.ID, "Platform"},
	}

	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := now.AddDate(0, 0, -7*weekOffset).Truncate(24 * time.Hour)

		for _, team := range childTeams {
			// Get team member IDs
			var memberIDs []string
			for engID, eng := range engineerMap {
				if eng.teamID == team.id {
					memberIDs = append(memberIDs, engID)
				}
			}

			if len(memberIDs) == 0 {
				continue
			}

			// Aggregate scores from team members
			var totalScore, throughputScore, qualityScore, speedScore, collaborationScore, impactScore float64
			var count int

			for _, memberID := range memberIDs {
				var score, tp, qual, spd, collab, imp float64
				err := database.DB.QueryRow(`
					SELECT total_score, throughput_score, quality_score, speed_score,
					       collaboration_score, impact_score
					FROM performance_scores
					WHERE engineer_id = ? AND week_start = ?
				`, memberID, weekStart).Scan(&score, &tp, &qual, &spd, &collab, &imp)

				if err == nil {
					totalScore += score
					throughputScore += tp
					qualityScore += qual
					speedScore += spd
					collaborationScore += collab
					impactScore += imp
					count++
				}
			}

			if count > 0 {
				// Calculate team averages
				teamScoreID := uuid.New().String()
				_, err := database.DB.Exec(`
					INSERT INTO team_performance_scores (
						id, team_id, week_start, total_score,
						throughput_score, quality_score, speed_score,
						collaboration_score, impact_score, member_count, created_at
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, teamScoreID, team.id, weekStart, totalScore/float64(count),
					throughputScore/float64(count), qualityScore/float64(count), speedScore/float64(count),
					collaborationScore/float64(count), impactScore/float64(count), count, now)

				if err != nil {
					log.Printf("Warning: Failed to create team performance score: %v", err)
				} else {
					teamScoreCounter++
				}
			}
		}
	}
	fmt.Printf("  Created %d team performance scores (8 weeks x %d teams)\n", teamScoreCounter, len(childTeams))

	// 7. Generate Weekly Briefings
	fmt.Println("\n7. Generating weekly briefings...")
	for weekOffset := 1; weekOffset <= 3; weekOffset++ {
		weekStart := now.AddDate(0, 0, -7*weekOffset).Truncate(24 * time.Hour)

		briefing := &models.WeeklyBriefing{
			TeamID:    engineeringTeam.ID,
			WeekStart: weekStart,
			TLDR:      fmt.Sprintf("Week of %s: Strong delivery with %d PRs merged, %d issues closed. Team velocity tracking above target.", weekStart.Format("Jan 2"), 30+(weekOffset*2), 15+weekOffset),
			KeyMetrics: []models.MetricSummary{
				{Name: "Team Score", Value: fmt.Sprintf("%.1f", 95.0-float64(weekOffset)*2), ChangePercent: -2.0, Direction: "down"},
				{Name: "PRs Merged", Value: fmt.Sprintf("%d", 30+(weekOffset*2)), ChangePercent: 5.0, Direction: "up"},
				{Name: "Issues Closed", Value: fmt.Sprintf("%d", 15+weekOffset), ChangePercent: 3.0, Direction: "up"},
			},
			NeedsAttention: []models.AttentionItem{
				{
					EngineerID:     engineerIDs[6], // Maya Patel (Junior)
					EngineerName:   "Maya Patel",
					Severity:       "warning",
					Issue:          "Slower PR cycle time than peers",
					Evidence:       []string{"Average 4.5 days vs team average of 2.1 days", "3 PRs currently in review > 3 days"},
					SuggestedTopic: "Discuss code review process and identify blockers",
				},
			},
			Insights: []models.AIInsight{
				{
					Observation:    fmt.Sprintf("Backend team velocity increased %d%% this week", 5+(weekOffset*2)),
					Context:        "Marcus and Elena shipped major refactoring, unlocking faster feature development",
					Recommendation: "Consider documenting the new patterns for other teams",
				},
				{
					Observation:    "Frontend team showing strong collaboration metrics",
					Context:        "Code review participation up 25%, average review depth improved",
					Recommendation: "Recognize team collaboration in next all-hands",
				},
			},
			TrendingUp:    []string{"Marcus Johnson", "Elena Rodriguez", "Priya Sharma"},
			TrendingDown:  []string{"Maya Patel", "Nina Okafor"},
			TalkingPoints: fmt.Sprintf("1. Celebrate backend refactoring completion\n2. Check in with Maya on PR blockers\n3. Platform team capacity planning for Q%d", (weekOffset%4)+1),
			GeneratedAt:   weekStart.Add(24 * 6 * time.Hour), // Saturday morning
		}

		if err := briefingStore.CreateBriefing(briefing); err != nil {
			log.Printf("Warning: Failed to create briefing: %v", err)
		}
		fmt.Printf("  Created briefing for week of %s\n", weekStart.Format("Jan 2"))
	}

	// 8. Create Sprints (one per child team)
	fmt.Println("\n8. Creating sprints...")

	sprintConfigs := []struct {
		name            string
		weeksAgo        int
		status          string
		committedPoints int
		completedPoints int
		capacity        int
	}{
		{"Sprint 45", 4, "completed", 85, 82, 30},
		{"Sprint 46", 2, "active", 90, 45, 30},
		{"Sprint 47", 0, "planning", 0, 0, 30},
		{"Sprint 48", -2, "planning", 0, 0, 30},
	}

	sprintTeams := []struct {
		id   string
		name string
	}{
		{backendTeam.ID, "Backend"},
		{frontendTeam.ID, "Frontend"},
		{platformTeam.ID, "Platform"},
	}

	var sprintIDs []string
	for _, team := range sprintTeams {
		for _, s := range sprintConfigs {
			startDate := now.AddDate(0, 0, -7*s.weeksAgo)
			endDate := startDate.AddDate(0, 0, 14) // 2-week sprints

			sprint := &models.Sprint{
				Name:            fmt.Sprintf("%s - %s", team.name, s.name),
				StartDate:       startDate,
				EndDate:         endDate,
				CommittedPoints: s.committedPoints,
				CompletedPoints: s.completedPoints,
				TeamCapacity:    s.capacity,
				Status:          s.status,
				TeamID:          team.id,
			}

			if err := planningStore.CreateSprint(sprint); err != nil {
				log.Printf("Warning: Failed to create sprint: %v", err)
			} else {
				sprintIDs = append(sprintIDs, sprint.ID)
				fmt.Printf("  Created sprint: %s (%s, %d/%d points)\n", sprint.Name, s.status, s.completedPoints, s.committedPoints)
			}
		}
	}

	// 9. Create Stories
	fmt.Println("\n9. Creating stories...")
	stories := []struct {
		title       string
		points      int
		status      string
		sprintIndex int
		assignee    int // engineer index
	}{
		{"Implement user authentication", 8, "completed", 0, 1},
		{"Add OAuth2 support", 5, "completed", 0, 2},
		{"Create API rate limiting", 5, "completed", 0, 1},
		{"Build dashboard UI", 8, "completed", 0, 4},
		{"Add chart components", 5, "completed", 0, 5},
		{"Implement data export", 5, "completed", 0, 7},

		{"Refactor backend services", 13, "in_progress", 1, 1},
		{"Optimize database queries", 8, "in_progress", 1, 2},
		{"Add frontend tests", 5, "in_progress", 1, 5},
		{"Update documentation", 3, "in_progress", 1, 8},
		{"Fix critical bugs", 8, "completed", 1, 3},

		{"Build notification system", 13, "todo", 2, -1},
		{"Add email templates", 5, "todo", 2, -1},
		{"Implement webhooks", 8, "todo", 2, -1},
		{"Create admin panel", 13, "todo", 2, -1},
	}

	for i, s := range stories {
		var assigneeID *string
		if s.assignee >= 0 && s.assignee < len(engineerIDs) {
			assigneeID = &engineerIDs[s.assignee]
		}

		var sprintID *string
		if s.sprintIndex < len(sprintIDs) && sprintIDs[s.sprintIndex] != "" {
			sprintID = &sprintIDs[s.sprintIndex]
		}

		story := &models.Story{
			Title:       s.title,
			StoryPoints: s.points,
			Status:      s.status,
			SprintID:    sprintID,
			AssigneeID:  assigneeID,
		}

		if err := planningStore.CreateStory(story); err != nil {
			log.Printf("Warning: Failed to create story %d: %v", i, err)
		}
	}
	fmt.Printf("  Created %d stories across sprints\n", len(stories))

	// ========================================
	// PDR-7 FEATURE SEED DATA
	// ========================================

	// 10. Create Alert System Data
	fmt.Println("\n10. Creating alert system data...")

	// Alert Rules
	alertRules := []struct {
		name      string
		alertType string
		severity  string
		threshold *float64
		operator  string
		target    string
		targetID  *string
	}{
		{"High Cycle Time Alert", "cycle_time_high", "warning", ptrFloat64(48.0), ">", "org", nil},
		{"Low Velocity Alert", "velocity_drop", "warning", ptrFloat64(0.8), "<", "team", &backendTeam.ID},
		{"PR Queue Buildup", "pr_queue", "critical", ptrFloat64(10.0), ">", "team", &frontendTeam.ID},
		{"Test Coverage Drop", "coverage_drop", "warning", ptrFloat64(70.0), "<", "team", &platformTeam.ID},
		{"Burnout Risk", "burnout_signal", "critical", nil, "", "engineer", &engineerIDs[6]}, // Maya Patel
	}

	alertRuleIDs := make([]string, 0)
	for _, r := range alertRules {
		rule := &models.AlertRule{
			Name:              r.name,
			Description:       fmt.Sprintf("Monitors %s and alerts when threshold is exceeded", r.alertType),
			AlertType:         r.alertType,
			Enabled:           true,
			ThresholdValue:    r.threshold,
			ThresholdOperator: r.operator,
			Severity:          r.severity,
			TargetEntity:      r.target,
			TargetID:          getStringValue(r.targetID),
		}
		if err := alertsStore.CreateAlertRule(rule); err != nil {
			log.Printf("Warning: Failed to create alert rule: %v", err)
		} else {
			alertRuleIDs = append(alertRuleIDs, rule.ID)
		}
	}
	fmt.Printf("  Created %d alert rules\n", len(alertRuleIDs))

	// Alert Instances
	alertInstances := []struct {
		ruleIndex  int
		title      string
		message    string
		entityType string
		entityID   *string
		acked      bool
		dismissed  bool
		resolved   bool
	}{
		{0, "Backend cycle time exceeds 48 hours", "Average cycle time for Backend team is 52 hours, above threshold of 48 hours", "team", &backendTeam.ID, true, false, false},
		{1, "Backend team velocity dropped 25%", "Sprint velocity dropped from 45 to 34 story points", "team", &backendTeam.ID, false, false, false},
		{2, "10+ PRs waiting for review", "Frontend team has 12 PRs waiting for review, blocking progress", "team", &frontendTeam.ID, true, false, true},
		{3, "Test coverage below 70%", "Platform team coverage is 65%, below target of 70%", "team", &platformTeam.ID, false, true, false},
		{4, "High workload detected for Maya Patel", "Working hours exceed 50hrs/week for 3 consecutive weeks", "engineer", &engineerIDs[6], true, false, false},
	}

	alertInstanceIDs := make([]string, 0)
	for i, a := range alertInstances {
		if a.ruleIndex >= len(alertRuleIDs) {
			continue
		}
		instance := &models.AlertInstance{
			RuleID:     alertRuleIDs[a.ruleIndex],
			Title:      a.title,
			Message:    a.message,
			Severity:   alertRules[a.ruleIndex].severity,
			EntityType: a.entityType,
			EntityID:   getStringValue(a.entityID),
			FiredAt:    now.AddDate(0, 0, -i-1),
		}
		if a.acked {
			ackTime := now.AddDate(0, 0, -i).Add(time.Hour * 2)
			instance.AcknowledgedAt = &ackTime
			instance.AcknowledgedBy = &engineerIDs[0] // Sarah Chen
		}
		if a.dismissed {
			dismissTime := now.AddDate(0, 0, -i).Add(time.Hour * 4)
			instance.DismissedAt = &dismissTime
			instance.DismissedBy = &engineerIDs[0]
		}
		if a.resolved {
			resolveTime := now.AddDate(0, 0, -i).Add(time.Hour * 6)
			instance.ResolvedAt = &resolveTime
		}
		if err := alertsStore.CreateAlertInstance(instance); err != nil {
			log.Printf("Warning: Failed to create alert instance: %v", err)
		} else {
			alertInstanceIDs = append(alertInstanceIDs, instance.ID)
		}
	}
	fmt.Printf("  Created %d alert instances\n", len(alertInstanceIDs))

	// Alert Channels
	for _, ruleID := range alertRuleIDs[:3] { // First 3 rules have channels
		channel := &models.AlertChannel{
			RuleID:        ruleID,
			ChannelType:   "slack",
			ChannelConfig: `{"webhook_url": "https://hooks.slack.com/services/EXAMPLE", "channel": "#eng-alerts"}`,
			Enabled:       true,
		}
		if err := alertsStore.CreateAlertChannel(channel); err != nil {
			log.Printf("Warning: Failed to create alert channel: %v", err)
		}
	}
	fmt.Printf("  Created alert channels\n")

	// 11. Create Goal Tracking Data
	fmt.Println("\n11. Creating goal tracking data...")

	// Goals
	goals := []struct {
		title       string
		goalType    string
		ownerType   string
		ownerID     *string
		timePeriod  string
		tracking    string
		status      string
		progress    int
		weeksOffset int
	}{
		{"Improve team velocity by 20%", "team", "team", &backendTeam.ID, "Q1 2025", "automatic", "active", 45, -4},
		{"Reduce average cycle time to under 36 hours", "team", "team", &frontendTeam.ID, "Q1 2025", "automatic", "active", 60, -4},
		{"Achieve 90% test coverage", "team", "team", &platformTeam.ID, "Q1 2025", "automatic", "at_risk", 30, -4},
		{"Lead 2 major feature initiatives", "individual", "engineer", &engineerIDs[1], "Q1 2025", "manual", "active", 50, -4},
		{"Mentor 2 junior engineers", "individual", "engineer", &engineerIDs[2], "Q1 2025", "hybrid", "active", 75, -4},
		{"Learn Kubernetes and deploy 3 services", "individual", "engineer", &engineerIDs[8], "Q1 2025", "manual", "active", 33, -4},
		{"Improve code review response time", "team", "team", &engineeringTeam.ID, "Q1 2025", "automatic", "active", 55, -4},
		{"Complete Go certification", "individual", "engineer", &engineerIDs[3], "Q1 2025", "manual", "completed", 100, -8},
	}

	goalIDs := make([]string, 0)
	for _, g := range goals {
		startDate := now.AddDate(0, 0, g.weeksOffset*7)
		endDate := startDate.AddDate(0, 3, 0) // 3 months

		goal := &models.Goal{
			Title:              g.title,
			Description:        fmt.Sprintf("Goal for %s to %s", g.ownerType, g.title),
			GoalType:           g.goalType,
			OwnerType:          g.ownerType,
			OwnerID:            g.ownerID,
			TimePeriod:         g.timePeriod,
			StartDate:          startDate,
			EndDate:            endDate,
			TrackingMethod:     g.tracking,
			Status:             g.status,
			ProgressPercentage: g.progress,
		}
		if err := goalsStore.CreateGoal(goal); err != nil {
			log.Printf("Warning: Failed to create goal: %v", err)
		} else {
			goalIDs = append(goalIDs, goal.ID)
		}
	}
	fmt.Printf("  Created %d goals\n", len(goalIDs))

	// Goal Milestones
	milestoneCount := 0
	for i, goalID := range goalIDs {
		numMilestones := 2
		if i < 3 {
			numMilestones = 3 // Team goals have more milestones
		}
		for j := 0; j < numMilestones; j++ {
			completed := j == 0 && goals[i].progress > 30 // First milestone completed if progress > 30%
			milestone := &models.GoalMilestone{
				GoalID:       goalID,
				Title:        fmt.Sprintf("Milestone %d", j+1),
				Description:  fmt.Sprintf("Key milestone %d for goal", j+1),
				TargetValue:  float64(100 / numMilestones),
				CurrentValue: float64(goals[i].progress / numMilestones),
				Unit:         "percent",
				Completed:    completed,
			}
			if completed {
				compTime := now.AddDate(0, 0, -7*(numMilestones-j))
				milestone.CompletedAt = &compTime
			}
			if err := goalsStore.CreateMilestone(milestone); err != nil {
				log.Printf("Warning: Failed to create milestone: %v", err)
			} else {
				milestoneCount++
			}
		}
	}
	fmt.Printf("  Created %d goal milestones\n", milestoneCount)

	// Goal Progress Logs
	progressLogCount := 0
	for i, goalID := range goalIDs {
		numLogs := 3
		for j := 0; j < numLogs; j++ {
			prevValue := float64((j * 100) / numLogs)
			newValue := float64(((j + 1) * 100) / numLogs)
			if newValue > float64(goals[i].progress) {
				newValue = float64(goals[i].progress)
			}

			progressLog := &models.GoalProgressLog{
				GoalID:        goalID,
				PreviousValue: &prevValue,
				NewValue:      &newValue,
				ChangeType:    "auto_detected",
				Evidence:      fmt.Sprintf(`{"event_count": %d, "source": "automatic"}`, j+1),
				LoggedAt:      now.AddDate(0, 0, -7*(numLogs-j)),
				LoggedBy:      ptrString("system"),
			}
			if err := goalsStore.LogProgress(progressLog); err != nil {
				log.Printf("Warning: Failed to create progress log: %v", err)
			} else {
				progressLogCount++
			}
		}
	}
	fmt.Printf("  Created %d goal progress logs\n", progressLogCount)

	// Goal Dependencies (first goal depends on second)
	if len(goalIDs) >= 2 {
		dep := &models.GoalDependency{
			GoalID:          goalIDs[1],
			DependsOnGoalID: goalIDs[0],
			DependencyType:  "relates_to",
			Status:          "active",
		}
		if err := goalsStore.CreateDependency(dep); err != nil {
			log.Printf("Warning: Failed to create goal dependency: %v", err)
		}
	}
	fmt.Printf("  Created goal dependencies\n")

	// 12. Create Skill Development Data
	fmt.Println("\n12. Creating skill development data...")

	// Skills
	skills := []struct {
		name        string
		category    string
		subcategory string
	}{
		{"Go Programming", "technical", "backend"},
		{"React Development", "technical", "frontend"},
		{"PostgreSQL", "technical", "database"},
		{"Docker", "technical", "devops"},
		{"Kubernetes", "technical", "devops"},
		{"System Design", "technical", "architecture"},
		{"Code Review", "technical", "quality"},
		{"Mentoring", "leadership", "people"},
		{"Documentation", "communication", "writing"},
		{"API Design", "technical", "backend"},
		{"Testing", "technical", "quality"},
		{"Communication", "communication", "verbal"},
	}

	skillIDs := make([]string, 0)
	for _, s := range skills {
		skill := &models.Skill{
			Name:        s.name,
			Category:    s.category,
			Subcategory: s.subcategory,
			Description: fmt.Sprintf("Proficiency in %s", s.name),
		}
		if err := skillsStore.CreateSkill(skill); err != nil {
			log.Printf("Warning: Failed to create skill: %v", err)
		} else {
			skillIDs = append(skillIDs, skill.ID)
		}
	}
	fmt.Printf("  Created %d skills\n", len(skillIDs))

	// Engineer Skills
	engineerSkillCount := 0
	for _, engID := range engineerIDs {
		// Each engineer has 3-5 skills
		numSkills := 3 + (len(engID) % 3)
		for i := 0; i < numSkills && i < len(skillIDs); i++ {
			skillIndex := (len(engID) + i) % len(skillIDs)
			levelScore := 50 + ((len(engID)+i)*7)%50

			engSkill := &models.EngineerSkill{
				EngineerID:    engID,
				SkillID:       skillIDs[skillIndex],
				LevelScore:    levelScore,
				Trajectory:    []string{"improving", "stable", "improving"}[i%3],
				LastEvaluated: now.AddDate(0, 0, -7),
				EvidenceCount: 5 + i,
			}
			if err := skillsStore.CreateEngineerSkill(engSkill); err != nil {
				log.Printf("Warning: Failed to create engineer skill: %v", err)
			} else {
				engineerSkillCount++
			}
		}
	}
	fmt.Printf("  Created %d engineer skills\n", engineerSkillCount)

	// Skill Evidence
	evidenceCount := 0
	for i, engID := range engineerIDs[:5] { // First 5 engineers
		for j := 0; j < 3; j++ {
			evidence := &models.SkillEvidence{
				EngineerID:     engID,
				SkillID:        skillIDs[j%len(skillIDs)],
				EvidenceType:   []string{"pr_complexity", "code_review", "design_doc"}[j%3],
				EvidenceSource: fmt.Sprintf("event_%d", i*3+j),
				Strength:       0.7 + float64(i)*0.05,
				Context:        `{"details": "Strong evidence of skill usage"}`,
				DetectedAt:     now.AddDate(0, 0, -(i*7 + j)),
			}
			if err := skillsStore.CreateSkillEvidence(evidence); err != nil {
				log.Printf("Warning: Failed to create skill evidence: %v", err)
			} else {
				evidenceCount++
			}
		}
	}
	fmt.Printf("  Created %d skill evidence entries\n", evidenceCount)

	// 13. Create Cost/ROI Analysis Data
	fmt.Println("\n13. Creating cost/ROI analysis data...")

	// Cost Configuration (one org-wide default)
	costConfig := &models.CostConfiguration{
		EntityType:    "org",
		MonthlyCost:   12500.0, // $150k/year average
		Currency:      "USD",
		EffectiveFrom: now.AddDate(-1, 0, 0),
		Notes:         "Average fully-loaded cost per engineer",
	}
	if err := costROIStore.CreateCostConfiguration(costConfig); err != nil {
		log.Printf("Warning: Failed to create cost configuration: %v", err)
	}
	fmt.Printf("  Created cost configuration\n")

	// Feature Values
	features := []struct {
		name       string
		valueType  string
		value      float64
		confidence string
	}{
		{"Authentication System", "arr_impact", 250000, "validated"},
		{"Dashboard Analytics", "efficiency_gain", 100000, "estimated"},
		{"API Gateway", "customer_acquisition", 500000, "validated"},
		{"Mobile App", "arr_impact", 750000, "estimated"},
		{"Notification System", "retention_improvement", 150000, "estimated"},
		{"Admin Panel", "efficiency_gain", 75000, "actual"},
		{"Search Feature", "customer_acquisition", 200000, "estimated"},
		{"Export Functionality", "arr_impact", 100000, "validated"},
	}

	featureIDs := make([]string, 0)
	for _, f := range features {
		feature := &models.FeatureValue{
			FeatureName:        f.name,
			FeatureDescription: fmt.Sprintf("Implementation of %s", f.name),
			ValueType:          f.valueType,
			ValueAmount:        &f.value,
			ValueCurrency:      "USD",
			ConfidenceLevel:    f.confidence,
			Source:             "manual",
			TimePeriod:         "Q4 2024",
		}
		if err := costROIStore.CreateFeatureValue(feature); err != nil {
			log.Printf("Warning: Failed to create feature value: %v", err)
		} else {
			featureIDs = append(featureIDs, feature.ID)
		}
	}
	fmt.Printf("  Created %d feature values\n", len(featureIDs))

	// Feature Work Items (link features to events)
	workItemCount := 0
	for i, featureID := range featureIDs[:5] {
		for j := 0; j < 3; j++ {
			workItem := &models.FeatureWorkItem{
				FeatureID:    featureID,
				WorkItemType: "story",
				WorkItemID:   fmt.Sprintf("STORY-%d", i*3+j+1),
				StoryPoints:  ptrInt(5 + j*2),
				ActualHours:  ptrFloat64(float64(20 + j*10)),
				EngineerID:   &engineerIDs[i],
				TeamID:       &backendTeam.ID,
				CompletedAt:  ptrTime(now.AddDate(0, 0, -(30 - i*7 - j))),
			}
			if err := costROIStore.CreateFeatureWorkItem(workItem); err != nil {
				log.Printf("Warning: Failed to create feature work item: %v", err)
			} else {
				workItemCount++
			}
		}
	}
	fmt.Printf("  Created %d feature work items\n", workItemCount)

	// Feature Costs
	for i, featureID := range featureIDs[:5] {
		cost := &models.FeatureCost{
			FeatureID:         featureID,
			TotalStoryPoints:  15 + i*5,
			TotalHours:        120.0 + float64(i*40),
			TotalCost:         25000.0 + float64(i*5000),
			CostBreakdown:     `{"labor": 20000, "infrastructure": 5000}`,
			ComputationMethod: "story_points",
			ComputedAt:        now,
		}
		if err := costROIStore.CreateFeatureCost(cost); err != nil {
			log.Printf("Warning: Failed to create feature cost: %v", err)
		}
	}
	fmt.Printf("  Created feature costs\n")

	// ROI Calculations
	roiCount := 0
	for i, featureID := range featureIDs[:5] {
		roi := &models.ROICalculation{
			FeatureID:     featureID,
			Investment:    25000.0 + float64(i*5000),
			ReturnValue:   ptrFloat64(features[i].value),
			ROIPercentage: ptrFloat64((features[i].value - 25000.0 - float64(i*5000)) / (25000.0 + float64(i*5000)) * 100),
			PaybackMonths: ptrFloat64(3.0 + float64(i)*0.5),
			Confidence:    features[i].confidence,
			CalculatedAt:  now,
		}
		if err := costROIStore.CreateROICalculation(roi); err != nil {
			log.Printf("Warning: Failed to create ROI calculation: %v", err)
		} else {
			roiCount++
		}
	}
	fmt.Printf("  Created %d ROI calculations\n", roiCount)

	// 14. Create Manager Context Data
	fmt.Println("\n14. Creating manager context data...")

	// Manager Notes (Sarah Chen making notes about team members)
	notes := []struct {
		subjectType string
		subjectID   string
		noteType    string
		title       string
		content     string
		mood        string
	}{
		{"engineer", engineerIDs[1], "1on1", "Marcus - Q1 Check-in", "Strong technical performance. Interested in leading more initiatives. Consider for tech lead role.", "positive"},
		{"engineer", engineerIDs[2], "performance", "Elena - Code Review Excellence", "Consistently provides high-quality code reviews. Mentoring David effectively.", "positive"},
		{"engineer", engineerIDs[3], "context", "David - Learning Progress", "Making good progress on backend skills. Needs more exposure to system design.", "neutral"},
		{"engineer", engineerIDs[6], "1on1", "Maya - Workload Discussion", "Concerned about workload. Discussed prioritization and delegation strategies.", "concerned"},
		{"team", backendTeam.ID, "context", "Backend Team - Sprint 5 Retrospective", "Team velocity improving. Tech debt paydown showing results.", "positive"},
		{"team", frontendTeam.ID, "performance", "Frontend Team - Q1 Performance", "Solid delivery but PR review time needs improvement.", "neutral"},
	}

	noteCount := 0
	for _, n := range notes {
		note := &models.ManagerNote{
			ManagerID:   engineerIDs[0], // Sarah Chen
			SubjectType: n.subjectType,
			SubjectID:   n.subjectID,
			NoteType:    n.noteType,
			Title:       n.title,
			Content:     n.content,
			Visibility:  "private",
			Mood:        n.mood,
		}
		if err := contextStore.CreateManagerNote(note); err != nil {
			log.Printf("Warning: Failed to create manager note: %v", err)
		} else {
			noteCount++
		}
	}
	fmt.Printf("  Created %d manager notes\n", noteCount)

	// Sentiment Surveys
	survey := &models.SentimentSurvey{
		SurveyType:  "quarterly",
		Title:       "Q1 2025 Team Sentiment Survey",
		Description: "Quarterly pulse check on team morale and satisfaction",
		Questions: []models.SurveyQuestion{
			{ID: "q1", Type: "scale", Question: "How satisfied are you with your work?"},
			{ID: "q2", Type: "scale", Question: "How is team collaboration?"},
			{ID: "q3", Type: "scale", Question: "Do you feel supported?"},
		},
		TargetAudience: "all",
		Anonymous:      true,
		Active:         true,
		CreatedBy:      engineerIDs[0],
		ExpiresAt:      ptrTime(now.AddDate(0, 0, 30)),
	}
	if err := contextStore.CreateSentimentSurvey(survey); err != nil {
		log.Printf("Warning: Failed to create sentiment survey: %v", err)
	}
	fmt.Printf("  Created sentiment survey\n")

	// Survey Responses
	responseCount := 0
	for i, engID := range engineerIDs {
		response := &models.SurveyResponse{
			SurveyID:     survey.ID,
			RespondentID: &engID,
			Responses: map[string]interface{}{
				"q1": 4 + i%2,
				"q2": 4 + i%2,
				"q3": 3 + i%3,
			},
			MoodRating:   ptrInt(4 + i%2),
			TextFeedback: "Overall positive experience. Team collaboration is strong.",
			SubmittedAt:  now.AddDate(0, 0, -i),
		}
		if err := contextStore.CreateSurveyResponse(response); err != nil {
			log.Printf("Warning: Failed to create survey response: %v", err)
		} else {
			responseCount++
		}
	}
	fmt.Printf("  Created %d survey responses\n", responseCount)

	// 15. Create Predictive Analytics Data
	fmt.Println("\n15. Creating predictive analytics data...")

	// Forecasts
	forecasts := []struct {
		forecastType string
		entityType   string
		entityID     *string
		timeHorizon  string
		value        float64
		confidence   float64
		modelType    string
	}{
		{"sprint_completion", "sprint", &sprintIDs[0], "next_sprint", 85.0, 75.0, "linear_regression"},
		{"team_velocity", "team", &backendTeam.ID, "next_quarter", 120.0, 70.0, "moving_average"},
		{"goal_completion", "goal", &goalIDs[0], "Q1_end", 92.0, 65.0, "monte_carlo"},
		{"attrition_risk", "engineer", &engineerIDs[6], "next_quarter", 0.35, 60.0, "linear_regression"},
		{"cost_projection", "team", &frontendTeam.ID, "next_quarter", 375000.0, 80.0, "linear_regression"},
	}

	forecastCount := 0
	for _, f := range forecasts {
		forecast := &models.Forecast{
			ForecastType:           f.forecastType,
			EntityType:             f.entityType,
			EntityID:               f.entityID,
			TimeHorizon:            f.timeHorizon,
			PredictedValue:         f.value,
			PredictedDate:          ptrTime(now.AddDate(0, 3, 0)),
			ConfidencePercentage:   f.confidence,
			ConfidenceIntervalLow:  ptrFloat64(f.value * 0.9),
			ConfidenceIntervalHigh: ptrFloat64(f.value * 1.1),
			ModelType:              f.modelType,
			InputData:              `{"historical_data_points": 12}`,
			Assumptions:            `{"assumes_stable_team": true}`,
			ExpiresAt:              ptrTime(now.AddDate(0, 3, 0)),
		}
		if err := forecastingStore.CreateForecast(forecast); err != nil {
			log.Printf("Warning: Failed to create forecast: %v", err)
		} else {
			forecastCount++
		}
	}
	fmt.Printf("  Created %d forecasts\n", forecastCount)

	// Risk Predictions
	risks := []struct {
		riskType    string
		entityType  string
		entityID    *string
		riskScore   float64
		probability float64
		severity    string
	}{
		{"attrition", "engineer", &engineerIDs[6], 0.35, 35.0, "medium"},
		{"timeline_miss", "team", &backendTeam.ID, 0.28, 28.0, "low"},
		{"burnout", "engineer", &engineerIDs[1], 0.42, 42.0, "medium"},
		{"quality_degradation", "team", &frontendTeam.ID, 0.18, 18.0, "low"},
	}

	riskCount := 0
	for _, r := range risks {
		risk := &models.RiskPrediction{
			RiskType:              r.riskType,
			EntityType:            r.entityType,
			EntityID:              r.entityID,
			RiskScore:             r.riskScore,
			ProbabilityPercentage: r.probability,
			ImpactSeverity:        r.severity,
			ContributingFactors:   `["high_workload", "long_hours"]`,
			MitigationSuggestions: `["redistribute_work", "add_resources"]`,
			PredictedAt:           now,
			ValidUntil:            ptrTime(now.AddDate(0, 1, 0)),
		}
		if err := forecastingStore.CreateRiskPrediction(risk); err != nil {
			log.Printf("Warning: Failed to create risk prediction: %v", err)
		} else {
			riskCount++
		}
	}
	fmt.Printf("  Created %d risk predictions\n", riskCount)

	// 16. Create Action Tracking Data
	fmt.Println("\n16. Creating action tracking data...")

	// Recommendations
	recommendations := []struct {
		sourceType  string
		recType     string
		priority    string
		subjectType string
		subjectID   *string
		title       string
		description string
		status      string
	}{
		{"ai_insight", "check_in", "high", "engineer", &engineerIDs[6], "Check in with Maya Patel", "High workload detected. Schedule 1-on-1 to discuss priorities and workload distribution.", "in_progress"},
		{"alert", "adjust_workload", "high", "team", &frontendTeam.ID, "Address PR review backlog", "12 PRs waiting for review. Consider pair review sessions or dedicated review time.", "completed"},
		{"briefing", "recognize_achievement", "medium", "engineer", &engineerIDs[2], "Recognize Elena's mentoring", "Elena has been providing excellent mentorship to David. Consider recognition.", "pending"},
		{"ai_insight", "skill_development", "medium", "engineer", &engineerIDs[3], "System design training for David", "David shows interest in architecture. Recommend system design course.", "pending"},
		{"manual", "address_blocker", "high", "team", &backendTeam.ID, "Address tech debt", "Tech debt is impacting velocity. Schedule tech debt sprint.", "in_progress"},
	}

	recommendationIDs := make([]string, 0)
	for _, r := range recommendations {
		rec := &models.Recommendation{
			SourceType:         r.sourceType,
			RecommendationType: r.recType,
			Priority:           r.priority,
			SubjectType:        r.subjectType,
			SubjectID:          r.subjectID,
			Title:              r.title,
			Description:        r.description,
			SuggestedActions:   []string{"Schedule meeting", "Discuss options", "Document decision"},
			Context:            `{"confidence": "high", "data_points": 15}`,
			AssignedTo:         &engineerIDs[0], // Sarah Chen
			Status:             r.status,
		}
		if err := actionsStore.CreateRecommendation(rec); err != nil {
			log.Printf("Warning: Failed to create recommendation: %v", err)
		} else {
			recommendationIDs = append(recommendationIDs, rec.ID)
		}
	}
	fmt.Printf("  Created %d recommendations\n", len(recommendationIDs))

	// Actions
	actions := []struct {
		recIndex    int
		actionType  string
		title       string
		description string
	}{
		{0, "conversation", "1-on-1 with Maya about workload", "Discussed current workload, identified blockers, reprioritized tasks"},
		{1, "process_change", "Implemented daily review time", "Team now dedicates 30 minutes daily to PR reviews"},
		{4, "workload_adjustment", "Scheduled tech debt sprint", "Allocated Sprint 8 for tech debt paydown"},
	}

	actionCount := 0
	for _, a := range actions {
		if a.recIndex >= len(recommendationIDs) {
			continue
		}
		action := &models.Action{
			RecommendationID: &recommendationIDs[a.recIndex],
			ActionType:       a.actionType,
			SubjectType:      recommendations[a.recIndex].subjectType,
			SubjectID:        recommendations[a.recIndex].subjectID,
			Title:            a.title,
			Description:      a.description,
			TakenBy:          engineerIDs[0], // Sarah Chen
			TakenAt:          now.AddDate(0, 0, -3),
			Evidence:         `{"meeting_notes": "Positive discussion, action items identified"}`,
			ExpectedOutcome:  "Improved workload balance and reduced stress",
			FollowUpDate:     ptrTime(now.AddDate(0, 0, 14)),
		}
		if err := actionsStore.CreateAction(action); err != nil {
			log.Printf("Warning: Failed to create action: %v", err)
		} else {
			actionCount++
		}
	}
	fmt.Printf("  Created %d actions\n", actionCount)

	// Summary
	fmt.Println("\n========================================")
	fmt.Println("Demo data seeding completed successfully!")
	fmt.Println("========================================")
	fmt.Println("\nV1.0 Core Features:")
	fmt.Printf("  - %d roles (Junior, Mid, Senior, Staff)\n", len(roles))
	fmt.Printf("  - %d teams (Engineering, Backend, Frontend, Platform)\n", 4)
	fmt.Printf("  - %d engineers with realistic profiles\n", len(engineers))
	fmt.Printf("  - %d events over 8 weeks\n", len(events))
	fmt.Printf("  - %d performance scores\n", scoreCounter)
	fmt.Printf("  - %d team performance scores\n", teamScoreCounter)
	fmt.Printf("  - %d weekly briefings\n", 3)
	fmt.Printf("  - %d sprints\n", len(sprintIDs))
	fmt.Printf("  - %d stories\n", len(stories))
	fmt.Println("\nPDR-7 Advanced Features:")
	fmt.Printf("  - %d alert rules and %d alert instances\n", len(alertRuleIDs), len(alertInstanceIDs))
	fmt.Printf("  - %d goals with %d milestones and %d progress logs\n", len(goalIDs), milestoneCount, progressLogCount)
	fmt.Printf("  - %d skills tracked across %d engineer skill assignments with %d evidence entries\n", len(skillIDs), engineerSkillCount, evidenceCount)
	fmt.Printf("  - %d features with cost analysis and %d ROI calculations\n", len(featureIDs), roiCount)
	fmt.Printf("  - %d manager notes, 1 sentiment survey with %d responses\n", noteCount, responseCount)
	fmt.Printf("  - %d forecasts and %d risk predictions\n", forecastCount, riskCount)
	fmt.Printf("  - %d recommendations and %d actions taken\n", len(recommendationIDs), actionCount)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'engineerdna serve' to start the server")
	fmt.Println("2. Visit http://127.0.0.1:3847 to explore all features with demo data")
	fmt.Println("3. Run 'engineerdna reset' to remove all demo data")
}

// Helper functions for seed data
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getTeamName(teamID string, engineering, backend, frontend, platform *models.Team) string {
	switch teamID {
	case engineering.ID:
		return "Engineering"
	case backend.ID:
		return "Backend"
	case frontend.ID:
		return "Frontend"
	case platform.ID:
		return "Platform"
	default:
		return "Unknown"
	}
}

func cmdReset() {
	fmt.Println("Removing demo data...")

	cfg := config.DefaultConfig()

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Count demo data across all tables
	var eventCount, engineerCount, teamCount, roleCount, scoreCount, briefingCount, sprintCount, storyCount int

	database.DB.QueryRow("SELECT COUNT(*) FROM events WHERE source IN ('demo', 'github', 'jira')").Scan(&eventCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM engineers WHERE email LIKE '%@example.com'").Scan(&engineerCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM teams").Scan(&teamCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM roles").Scan(&roleCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM performance_scores").Scan(&scoreCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM weekly_briefings").Scan(&briefingCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM sprints").Scan(&sprintCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM stories").Scan(&storyCount)

	totalCount := eventCount + engineerCount + teamCount + roleCount + scoreCount + briefingCount + sprintCount + storyCount
	if totalCount == 0 {
		fmt.Println("No demo data found")
		return
	}

	// Show what will be deleted
	fmt.Println("\nFound demo data:")
	if eventCount > 0 {
		fmt.Printf("  %d events\n", eventCount)
	}
	if scoreCount > 0 {
		fmt.Printf("  %d performance scores\n", scoreCount)
	}
	if briefingCount > 0 {
		fmt.Printf("  %d weekly briefings\n", briefingCount)
	}
	if storyCount > 0 {
		fmt.Printf("  %d stories\n", storyCount)
	}
	if sprintCount > 0 {
		fmt.Printf("  %d sprints\n", sprintCount)
	}
	if engineerCount > 0 {
		fmt.Printf("  %d engineers\n", engineerCount)
	}
	if teamCount > 0 {
		fmt.Printf("  %d teams\n", teamCount)
	}
	if roleCount > 0 {
		fmt.Printf("  %d roles\n", roleCount)
	}

	// Confirm deletion
	fmt.Print("\nDelete all demo data? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Cancelled")
		return
	}

	// Delete in correct order (respecting foreign keys)
	deleted := 0

	// 1. Delete stories (references sprints and engineers)
	if storyCount > 0 {
		result, err := database.DB.Exec("DELETE FROM stories")
		if err != nil {
			log.Printf("Warning: Failed to delete stories: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d stories\n", rows)
		}
	}

	// 2. Delete sprints
	if sprintCount > 0 {
		result, err := database.DB.Exec("DELETE FROM sprints")
		if err != nil {
			log.Printf("Warning: Failed to delete sprints: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d sprints\n", rows)
		}
	}

	// 3. Delete weekly briefings
	if briefingCount > 0 {
		result, err := database.DB.Exec("DELETE FROM weekly_briefings")
		if err != nil {
			log.Printf("Warning: Failed to delete briefings: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d weekly briefings\n", rows)
		}
	}

	// 4. Delete performance scores
	if scoreCount > 0 {
		result, err := database.DB.Exec("DELETE FROM performance_scores")
		if err != nil {
			log.Printf("Warning: Failed to delete performance scores: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d performance scores\n", rows)
		}
	}

	// 5. Delete team performance scores
	result, err := database.DB.Exec("DELETE FROM team_performance_scores")
	if err != nil {
		log.Printf("Warning: Failed to delete team performance scores: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d team performance scores\n", rows)
		}
	}

	// 6. Delete events
	if eventCount > 0 {
		result, err := database.DB.Exec("DELETE FROM events WHERE source IN ('demo', 'github', 'jira')")
		if err != nil {
			log.Printf("Warning: Failed to delete events: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d events\n", rows)
		}
	}

	// 7. Delete team memberships
	result, err = database.DB.Exec("DELETE FROM team_membership")
	if err != nil {
		log.Printf("Warning: Failed to delete team memberships: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d team memberships\n", rows)
		}
	}

	// 8. Delete teams
	if teamCount > 0 {
		result, err := database.DB.Exec("DELETE FROM teams")
		if err != nil {
			log.Printf("Warning: Failed to delete teams: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d teams\n", rows)
		}
	}

	// 9. Delete engineers
	if engineerCount > 0 {
		result, err := database.DB.Exec("DELETE FROM engineers WHERE email LIKE '%@example.com'")
		if err != nil {
			log.Printf("Warning: Failed to delete engineers: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d engineers\n", rows)
		}
	}

	// 10. Delete roles
	if roleCount > 0 {
		result, err := database.DB.Exec("DELETE FROM roles")
		if err != nil {
			log.Printf("Warning: Failed to delete roles: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d roles\n", rows)
		}
	}

	// 11. Delete scoring weights
	result, err = database.DB.Exec("DELETE FROM scoring_weights")
	if err != nil {
		log.Printf("Warning: Failed to delete scoring weights: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d scoring weights\n", rows)
		}
	}

	// 12. Delete unresolved identities
	result, err = database.DB.Exec("DELETE FROM unresolved_identities")
	if err != nil {
		log.Printf("Warning: Failed to delete unresolved identities: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d unresolved identities\n", rows)
		}
	}

	fmt.Printf("\n========================================")
	fmt.Printf("\nTotal: Deleted %d demo records\n", deleted)
	fmt.Println("Database reset completed successfully!")
	fmt.Println("========================================")
}

func cmdVersion() {
	fmt.Printf("EngineerDNA version %s\n", Version)
}

func cmdBackfill() {
	fmt.Println("Backfilling engineer_id for existing events...")

	cfg := config.DefaultConfig()

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize stores and services
	identityStore := db.NewIdentityStore(database.DB)
	identityService := identity.NewService(identityStore)

	// Get all events without engineer_id
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM events WHERE engineer_id IS NULL OR engineer_id = ''").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count events: %v", err)
	}

	if count == 0 {
		fmt.Println("No events to backfill")
		return
	}

	fmt.Printf("Found %d events without engineer_id\n", count)

	// Get all unresolved events
	rows, err := database.DB.Query("SELECT id, source, actor FROM events WHERE engineer_id IS NULL OR engineer_id = ''")
	if err != nil {
		log.Fatalf("Failed to query events: %v", err)
	}
	defer rows.Close()

	resolved := 0
	unresolved := 0

	for rows.Next() {
		var eventID, source, actor string
		if err := rows.Scan(&eventID, &source, &actor); err != nil {
			log.Printf("Failed to scan event: %v", err)
			continue
		}

		// Attempt to resolve identity
		engineerID, err := identityService.ResolveIdentity(source, actor)
		if err != nil {
			log.Printf("Failed to resolve identity for %s:%s: %v", source, actor, err)
			continue
		}

		if engineerID != "" {
			// Update event with engineer_id
			_, err := database.DB.Exec("UPDATE events SET engineer_id = ?, updated_at = ? WHERE id = ?",
				engineerID, time.Now().UTC(), eventID)
			if err != nil {
				log.Printf("Failed to update event %s: %v", eventID, err)
				continue
			}
			resolved++
		} else {
			unresolved++
		}
	}

	fmt.Printf("\nBackfill completed:\n")
	fmt.Printf("  Resolved: %d events\n", resolved)
	fmt.Printf("  Unresolved: %d events\n", unresolved)
	fmt.Println("\nCheck /api/identity/unresolved for unresolved identities")
}

func printUsage() {
	fmt.Println("EngineerDNA - Engineering metrics and insights platform")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  engineerdna <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init      Initialize EngineerDNA (create directories and database)")
	fmt.Println("  serve     Start the HTTP server")
	fmt.Println("  sync      Sync data from all enabled source plugins")
	fmt.Println("  seed      Generate comprehensive demo data (teams, engineers, events, scores, briefings, sprints)")
	fmt.Println("  reset     Remove all demo data from database")
	fmt.Println("  backfill  Backfill engineer_id for existing events")
	fmt.Println("  version   Print version information")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  engineerdna init")
	fmt.Println("  engineerdna serve")
	fmt.Println("  engineerdna sync")
	fmt.Println("  engineerdna seed")
	fmt.Println("  engineerdna backfill")
	fmt.Println("  engineerdna reset")
}
