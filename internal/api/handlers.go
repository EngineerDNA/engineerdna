package api

import (
	"database/sql"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/ai"
	"github.com/engineerdna/engineerdna/internal/anonymization"
	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/context"
	"github.com/engineerdna/engineerdna/internal/cost"
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
	"github.com/engineerdna/engineerdna/internal/skills"
	"github.com/engineerdna/engineerdna/internal/team"
)

// Server holds the API server dependencies
type Server struct {
	db                 *sql.DB
	eventStore         *db.EventStore
	pluginStore        *db.PluginStore
	anonStore          *db.AnonymizationStore
	auditStore         *db.AuditStore
	identityStore      *db.IdentityStore
	configStore        *db.ConfigStore
	scheduleStore      *db.ExportScheduleStore
	exportsStore       *db.ExportsStore
	scoringStore       *db.ScoringStore
	teamStore          *db.TeamStore
	briefingStore      *db.BriefingsStore
	planningStore      *db.PlanningStore
	settingsStore      *db.SettingsStore
	alertsStore        *db.AlertsStore
	goalsStore         *db.GoalsStore
	skillsStore        *db.SkillsStore
	costROIStore       *db.CostROIStore
	contextStore       *db.ContextStore
	forecastingStore   *db.ForecastingStore
	actionsStore       *db.ActionsStore
	dashboardStore     *db.DashboardStore
	engineerStore      *db.EngineerStore
	healthStore        *db.HealthStore
	executor           *plugin.Executor
	anonService        *anonymization.Service
	identityService    *identity.Service
	scoringService     *scoring.ScoringService
	teamService        *team.Service
	promotionService   *promotion.Service
	briefingService    *ai.BriefingService
	planningService    *planning.Service
	goalsService       *goals.Service
	skillsService      *skills.Service
	costService        *cost.Service
	roiService         *roi.Service
	costAnalyzer       *cost.Analyzer
	contextService     *context.Service
	sentimentService   *sentiment.Service
	sentimentAnalyzer  *sentiment.Analyzer
	forecastingService *forecasting.ForecastingService
	scheduler          *scheduler.Scheduler
	config             *config.Config
	frontendFS         fs.FS
	mux                *http.ServeMux
	version            string
	logger             *log.Logger
}

// NewServer creates a new API server
func NewServer(
	database *sql.DB,
	eventStore *db.EventStore,
	pluginStore *db.PluginStore,
	anonStore *db.AnonymizationStore,
	auditStore *db.AuditStore,
	identityStore *db.IdentityStore,
	configStore *db.ConfigStore,
	scheduleStore *db.ExportScheduleStore,
	exportsStore *db.ExportsStore,
	scoringStore *db.ScoringStore,
	teamStore *db.TeamStore,
	briefingStore *db.BriefingsStore,
	planningStore *db.PlanningStore,
	settingsStore *db.SettingsStore,
	alertsStore *db.AlertsStore,
	goalsStore *db.GoalsStore,
	skillsStore *db.SkillsStore,
	costROIStore *db.CostROIStore,
	contextStore *db.ContextStore,
	forecastingStore *db.ForecastingStore,
	actionsStore *db.ActionsStore,
	dashboardStore *db.DashboardStore,
	engineerStore *db.EngineerStore,
	healthStore *db.HealthStore,
	executor *plugin.Executor,
	anonService *anonymization.Service,
	identityService *identity.Service,
	scoringService *scoring.ScoringService,
	teamService *team.Service,
	promotionService *promotion.Service,
	briefingService *ai.BriefingService,
	planningService *planning.Service,
	goalsService *goals.Service,
	skillsService *skills.Service,
	costService *cost.Service,
	roiService *roi.Service,
	costAnalyzer *cost.Analyzer,
	contextService *context.Service,
	sentimentService *sentiment.Service,
	sentimentAnalyzer *sentiment.Analyzer,
	forecastingService *forecasting.ForecastingService,
	sched *scheduler.Scheduler,
	cfg *config.Config,
	frontendFS fs.FS,
	version string,
	logger *log.Logger,
) *Server {
	return &Server{
		db:                 database,
		eventStore:         eventStore,
		pluginStore:        pluginStore,
		anonStore:          anonStore,
		auditStore:         auditStore,
		identityStore:      identityStore,
		configStore:        configStore,
		scheduleStore:      scheduleStore,
		exportsStore:       exportsStore,
		scoringStore:       scoringStore,
		teamStore:          teamStore,
		briefingStore:      briefingStore,
		planningStore:      planningStore,
		settingsStore:      settingsStore,
		alertsStore:        alertsStore,
		goalsStore:         goalsStore,
		skillsStore:        skillsStore,
		costROIStore:       costROIStore,
		contextStore:       contextStore,
		forecastingStore:   forecastingStore,
		actionsStore:       actionsStore,
		dashboardStore:     dashboardStore,
		engineerStore:      engineerStore,
		healthStore:        healthStore,
		executor:           executor,
		anonService:        anonService,
		identityService:    identityService,
		scoringService:     scoringService,
		teamService:        teamService,
		promotionService:   promotionService,
		briefingService:    briefingService,
		planningService:    planningService,
		goalsService:       goalsService,
		skillsService:      skillsService,
		costService:        costService,
		roiService:         roiService,
		costAnalyzer:       costAnalyzer,
		contextService:     contextService,
		sentimentService:   sentimentService,
		sentimentAnalyzer:  sentimentAnalyzer,
		forecastingService: forecastingService,
		scheduler:          sched,
		config:             cfg,
		frontendFS:         frontendFS,
		mux:                http.NewServeMux(),
		version:            version,
		logger:             logger,
	}
}

// handleFrontend serves React app and handles client-side routing
func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	// API routes should not be handled by frontend
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// Try to open the requested file
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Open file from embedded FS
	file, err := s.frontendFS.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		// File not found - serve index.html for client-side routing
		file, err = s.frontendFS.Open("index.html")
		if err != nil {
			// No index.html found - serve a basic message
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>EngineerDNA</title></head>
<body>
<h1>EngineerDNA</h1>
<p>Frontend is building. Check <a href="/api/health">/api/health</a> to verify the API is running.</p>
</body>
</html>`))
			return
		}
	}
	defer file.Close()

	// Serve the file
	http.ServeContent(w, r, path, time.Time{}, file.(io.ReadSeeker))
}
