package api

import (
	"net/http"
)

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() {
	// Health & System
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/version", s.handleVersion)
	s.mux.HandleFunc("/api/system/info", s.handleSystemInfo)

	// Events
	s.mux.HandleFunc("/api/events", s.handleEvents)
	s.mux.HandleFunc("/api/events/", s.handleEventByID)

	// Plugins
	s.mux.HandleFunc("/api/plugins", s.handlePlugins)
	s.mux.HandleFunc("/api/plugins/", s.handlePluginOperations)

	// Anonymization
	s.mux.HandleFunc("/api/anonymization/policies", s.handleAnonymizationPolicies)
	s.mux.HandleFunc("/api/anonymization/mappings", s.handleAnonymizationMappings)
	s.mux.HandleFunc("/api/anonymization/audit", s.handleAnonymizationAudit)

	// Metrics
	s.mux.HandleFunc("/api/metrics/throughput", s.handleMetricsThroughput)
	s.mux.HandleFunc("/api/metrics/cycle-time", s.handleMetricsCycleTime)
	s.mux.HandleFunc("/api/metrics/today", s.handleMetricsToday)
	s.mux.HandleFunc("/api/metrics/values", s.handleMetricValues)
	s.mux.HandleFunc("/api/metrics/calculate", s.handleMetricCalculate)

	// Activity
	s.mux.HandleFunc("/api/activity/today", s.handleActivityToday)

	// Insights
	s.mux.HandleFunc("/api/insights", s.handleInsights)

	// Exports
	s.mux.HandleFunc("/api/exports", s.handleExports)
	s.mux.HandleFunc("/api/exports/schedules", s.handleExportSchedules)
	s.mux.HandleFunc("/api/exports/schedules/", s.handleExportScheduleOperations)

	// Engineers
	s.mux.HandleFunc("/api/engineers", s.handleEngineers)
	s.mux.HandleFunc("/api/engineers/import", s.handleEngineersImport)
	s.mux.HandleFunc("/api/engineers/", s.handleEngineerOperations)

	// Identity Resolution
	s.mux.HandleFunc("/api/identity/unresolved", s.handleIdentityUnresolved)
	s.mux.HandleFunc("/api/identity/resolve", s.handleIdentityResolve)
	s.mux.HandleFunc("/api/identity/suggestions", s.handleIdentitySuggestions)
	s.mux.HandleFunc("/api/identity/merge", s.handleIdentityMerge)
	s.mux.HandleFunc("/api/identity/ignore", s.handleIdentityIgnore)
	s.mux.HandleFunc("/api/identity/ignored", s.handleIdentityIgnoredList)

	// Onboarding
	s.mux.HandleFunc("/api/onboarding/status", s.handleOnboardingStatus)
	s.mux.HandleFunc("/api/onboarding/complete", s.handleOnboardingComplete)

	// Settings
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/settings/sync-schedule", s.handleSettingsSyncSchedule)
	s.mux.HandleFunc("/api/settings/anonymization-strategy", s.handleSettingsAnonymizationStrategy)
	s.mux.HandleFunc("/api/settings/port", s.handleSettingsPort)

	// Roles
	s.mux.HandleFunc("/api/roles", s.handleRoles)
	s.mux.HandleFunc("/api/roles/", s.handleRoleByID)

	// Scoring Weights
	s.mux.HandleFunc("/api/scoring/weights", s.handleScoringWeights)

	// Score Calculation
	s.mux.HandleFunc("/api/scoring/calculate", s.handleCalculateScore)

	// Performance Scores
	s.mux.HandleFunc("/api/performance/individual/", s.handlePerformanceIndividual)
	s.mux.HandleFunc("/api/performance/team/", s.handlePerformanceTeam)

	// Teams
	s.mux.HandleFunc("/api/teams", s.handleTeams)
	s.mux.HandleFunc("/api/teams/hierarchy", s.handleTeamsHierarchy)
	s.mux.HandleFunc("/api/teams/org/scorecard", s.handleOrgScorecard)
	s.mux.HandleFunc("/api/teams/", s.handleTeamByID)

	// Promotion Signals
	s.mux.HandleFunc("/api/promotion/signals", s.handlePromotionSignals)
	s.mux.HandleFunc("/api/promotion/signals/", s.handlePromotionSignalsByEngineer)
	s.mux.HandleFunc("/api/promotion/dismiss/", s.handlePromotionSignalDismiss)
	s.mux.HandleFunc("/api/promotion/detect", s.handlePromotionDetect)

	// Briefings
	s.mux.HandleFunc("/api/briefing/weekly/", s.handleBriefingWeekly)
	s.mux.HandleFunc("/api/briefing/history/", s.handleBriefingHistory)
	s.mux.HandleFunc("/api/briefing/org", s.handleBriefingOrg)

	// Planning
	s.mux.HandleFunc("/api/planning/sprints", s.handleSprints)
	s.mux.HandleFunc("/api/planning/sprints/", s.handleSprintByID)
	s.mux.HandleFunc("/api/planning/sprint-burndown", s.handleSprintBurndown)
	s.mux.HandleFunc("/api/planning/velocity/", s.handleVelocity)
	s.mux.HandleFunc("/api/planning/estimate", s.handleEstimate)
	s.mux.HandleFunc("/api/planning/capacity", s.handleCapacity)

	// Alerts
	s.mux.HandleFunc("/api/alerts/rules", s.handleAlertRules)
	s.mux.HandleFunc("/api/alerts/rules/", s.handleAlertRuleByID)
	s.mux.HandleFunc("/api/alerts/instances", s.handleAlertInstances)
	s.mux.HandleFunc("/api/alerts/instances/", s.handleAlertInstanceByID)
	s.mux.HandleFunc("/api/alerts/channels", s.handleAlertChannels)
	s.mux.HandleFunc("/api/alerts/channels/", s.handleAlertChannelByID)

	// Widget Queries
	s.mux.HandleFunc("/api/metrics/aggregate", s.handleMetricsAggregate)
	s.mux.HandleFunc("/api/metrics/timeseries", s.handleMetricsTimeseries)
	s.mux.HandleFunc("/api/metrics/compare", s.handleMetricsCompare)
	s.mux.HandleFunc("/api/engineers/performance", s.handleEngineersPerformance)
	s.mux.HandleFunc("/api/metrics/health-status", s.handleMetricsHealthStatus)
	s.mux.HandleFunc("/api/alerts/timeline", s.handleAlertsTimeline)

	// Goals
	s.mux.HandleFunc("/api/goals", s.handleGoals)
	s.mux.HandleFunc("/api/goals/summary", s.handleGoalsSummary)
	s.mux.HandleFunc("/api/goals/", s.handleGoalByID)

	// Skills
	s.mux.HandleFunc("/api/skills", s.handleSkills)

	// Cost & ROI
	s.mux.HandleFunc("/api/cost/configuration", s.handleCostConfiguration)
	s.mux.HandleFunc("/api/cost/team/", s.handleTeamCost)
	s.mux.HandleFunc("/api/cost/engineer/", s.handleEngineerCost)
	s.mux.HandleFunc("/api/features", s.handleFeatures)
	s.mux.HandleFunc("/api/features/", s.handleFeatureByID)
	s.mux.HandleFunc("/api/roi/report", s.handleROIReport)
	s.mux.HandleFunc("/api/investment/breakdown", s.handleInvestmentBreakdown)
	s.mux.HandleFunc("/api/cost/efficiency", s.handleCostEfficiency)

	// Manager Notes & Context
	s.mux.HandleFunc("/api/notes", s.handleNotes)
	s.mux.HandleFunc("/api/notes/", s.handleNoteByID)
	s.mux.HandleFunc("/api/context/annotations", s.handleContextAnnotations)

	// Sentiment Surveys
	s.mux.HandleFunc("/api/surveys", s.handleSurveys)
	s.mux.HandleFunc("/api/surveys/", s.handleSurveyByID)

	// Sentiment Analysis
	s.mux.HandleFunc("/api/sentiment/team/", s.handleSentimentTeam)
	s.mux.HandleFunc("/api/sentiment/engineer/", s.handleSentimentEngineer)

	// Forecasting & Predictive Analytics
	s.mux.HandleFunc("/api/forecasts", s.handleForecasts)
	s.mux.HandleFunc("/api/forecasts/", s.handleForecastByID)
	s.mux.HandleFunc("/api/forecasts/sprint/", s.handleForecastSprintCompletion)
	s.mux.HandleFunc("/api/forecasts/goal/", s.handleForecastGoalCompletion)
	s.mux.HandleFunc("/api/forecasts/team/", s.handleForecastTeamVelocity)
	s.mux.HandleFunc("/api/forecasts/accuracy", s.handleForecastAccuracy)
	s.mux.HandleFunc("/api/forecasts/accuracy/record", s.handleRecordForecastAccuracy)

	// Risk Predictions
	s.mux.HandleFunc("/api/risks", s.handleRisks)
	s.mux.HandleFunc("/api/risks/engineer/", s.handleRiskEngineer)
	s.mux.HandleFunc("/api/risks/team/", s.handleRiskTeam)
	s.mux.HandleFunc("/api/risks/predict/attrition", s.handlePredictAttritionRisk)
	s.mux.HandleFunc("/api/risks/predict/timeline", s.handlePredictTimelineMiss)
	s.mux.HandleFunc("/api/risks/predict/quality", s.handlePredictQualityDegradation)

	// Scenarios
	s.mux.HandleFunc("/api/scenarios", s.handleScenarios)
	s.mux.HandleFunc("/api/scenarios/", s.handleScenarioByID)
	s.mux.HandleFunc("/api/scenarios/simulate/", s.handleSimulateScenario)
	s.mux.HandleFunc("/api/scenarios/compare", s.handleCompareScenarios)

	// Action Tracking & Recommendations
	s.mux.HandleFunc("/api/recommendations", s.handleRecommendations)
	s.mux.HandleFunc("/api/recommendations/pending", s.handleRecommendationsPending)
	s.mux.HandleFunc("/api/recommendations/", s.handleRecommendationByID)
	s.mux.HandleFunc("/api/actions", s.handleActions)
	s.mux.HandleFunc("/api/actions/", s.handleActionByID)
	s.mux.HandleFunc("/api/actions/effectiveness-report", s.handleEffectivenessReport)
	s.mux.HandleFunc("/api/follow-ups", s.handleFollowUps)
	s.mux.HandleFunc("/api/follow-ups/", s.handleFollowUpByID)

	// Attributes (multi-modal data)
	s.mux.HandleFunc("/api/attributes", s.handleAttributes)

	// Correlations (multi-modal data)
	s.mux.HandleFunc("/api/correlations", s.handleCorrelations)
	s.mux.HandleFunc("/api/correlations/", s.handleCorrelationByName)

	// Widget Registry
	s.mux.HandleFunc("/api/widgets/registry", s.handleWidgetRegistry)
	s.mux.HandleFunc("/api/widgets/registry/", s.handleWidgetByID)

	// Dashboards
	s.mux.HandleFunc("/api/dashboards", s.handleDashboards)
	s.mux.HandleFunc("/api/dashboards/", s.handleDashboardByID)
	s.mux.HandleFunc("/api/metrics/snapshots", s.handleMetricSnapshots)

	// Frontend - must be registered last to catch all non-API routes
	s.mux.HandleFunc("/", s.handleFrontend)
}

// Handler returns the HTTP handler with middleware applied
func (s *Server) Handler() http.Handler {
	// Apply middleware: MaxBytes (10MB) -> Compression -> CORS -> Logging
	handler := MaxBytesMiddleware(10 * 1024 * 1024)(s.mux)
	handler = CompressionMiddleware(handler)
	handler = CORSMiddleware(handler)
	handler = LoggingMiddleware(handler)
	return handler
}
