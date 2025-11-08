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
	"github.com/engineerdna/engineerdna/internal/constants"
	"github.com/engineerdna/engineerdna/internal/context"
	"github.com/engineerdna/engineerdna/internal/cost"
	"github.com/engineerdna/engineerdna/internal/dashboards"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/forecasting"
	"github.com/engineerdna/engineerdna/internal/goals"
	"github.com/engineerdna/engineerdna/internal/identity"
	"github.com/engineerdna/engineerdna/internal/planning"
	"github.com/engineerdna/engineerdna/internal/plugin"
	"github.com/engineerdna/engineerdna/internal/promotion"
	"github.com/engineerdna/engineerdna/internal/roi"
	"github.com/engineerdna/engineerdna/internal/scheduler"
	"github.com/engineerdna/engineerdna/internal/scoring"
	"github.com/engineerdna/engineerdna/internal/sentiment"
	"github.com/engineerdna/engineerdna/internal/services"
	"github.com/engineerdna/engineerdna/internal/skills"
	"github.com/engineerdna/engineerdna/internal/team"
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
	metricStore := db.NewMetricStore(database.DB)
	attributeStore := db.NewAttributeStore(database.DB)
	correlationStore := db.NewCorrelationStore(database.DB)
	manifestStore := db.NewPluginManifestStore(database.DB)

	// Initialize services
	keyStore, err := config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}

	anonService := anonymization.NewService(anonStore)
	identityService := identity.NewService(identityStore)
	scoringService := scoring.NewScoringService(database.DB, metricStore)
	teamService := team.NewService(database.DB, metricStore)
	promotionService := promotion.NewService(database.DB)
	planningService := planning.NewService(planningStore)
	goalsService := goals.NewService(goalsStore, scoringStore, metricStore)
	skillsService := skills.NewService(skillsStore, eventStore)
	costService := cost.NewService(costROIStore, teamStore, attributeStore)
	roiService := roi.NewService(costROIStore)
	costAnalyzer := cost.NewAnalyzer(costROIStore, eventStore, costService)
	contextService := context.NewService(contextStore, identityStore)
	sentimentService := sentiment.NewService(contextStore, teamStore)
	sentimentAnalyzer := sentiment.NewAnalyzer(contextStore, eventStore)
	forecastingService := forecasting.NewForecastingService(database.DB, forecastingStore, planningStore, scoringStore, metricStore, goalsStore)
	dashboardService := dashboards.NewService(dashboardStore, scoringStore, eventStore, alertsStore, goalsStore, teamStore)
	metricEngine := services.NewMetricEngine(eventStore, metricStore, attributeStore)
	correlationEngine := services.NewCorrelationEngine(metricEngine, eventStore, metricStore, attributeStore, correlationStore)
	ruleInsights := services.NewRuleBasedInsights(metricStore, eventStore, alertsStore, goalsStore)

	// Initialize plugin system
	pluginDirs := cfg.GetPluginDirs()
	loader := plugin.NewLoader(pluginDirs)
	executor := plugin.NewExecutor(loader, anonService, anonStore, pluginStore, auditStore, keyStore)

	// Initialize event normalizer and wire to event store
	eventNormalizer := services.NewEventNormalizer(executor.GetEventRegistry())
	eventStore.SetNormalizer(eventNormalizer)

	// Initialize widget registry
	widgetRegistry := plugin.NewWidgetRegistry(manifestStore)
	if err := widgetRegistry.LoadFromDatabase(); err != nil {
		log.Printf("Warning: failed to load widgets from database: %v", err)
	}

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
	alertScheduler := alerts.NewScheduler(database.DB, alertsStore, eventStore, scoringStore, metricStore, teamStore)
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
	if err := cost.CreateDefaultCostConfiguration(attributeStore); err != nil {
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
	server := api.NewServer(database.DB, eventStore, pluginStore, anonStore, auditStore, identityStore, configStore, scheduleStore, exportsStore, scoringStore, teamStore, briefingStore, planningStore, settingsStore, alertsStore, goalsStore, skillsStore, costROIStore, contextStore, forecastingStore, actionsStore, dashboardStore, engineerStore, healthStore, metricStore, attributeStore, correlationStore, manifestStore, executor, metricEngine, correlationEngine, ruleInsights, widgetRegistry, anonService, identityService, scoringService, teamService, promotionService, briefingService, planningService, goalsService, skillsService, costService, roiService, costAnalyzer, contextService, sentimentService, sentimentAnalyzer, forecastingService, sched, cfg, frontendFS, Version, log.Default())
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

	// Initialize services
	keyStore, err := config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}

	anonService := anonymization.NewService(anonStore)

	// Initialize plugin system
	pluginDirs := cfg.GetPluginDirs()
	loader := plugin.NewLoader(pluginDirs)
	executor := plugin.NewExecutor(loader, anonService, anonStore, pluginStore, auditStore, keyStore)

	// Get all configured source plugins
	configs, err := pluginStore.List(db.MaxQueryLimit)
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
		since := time.Now().UTC().Add(-constants.DefaultSyncLookback)
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
