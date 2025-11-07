# Changelog

All notable changes to EngineerDNA will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

**Dashboard Builder & Visualization System (PDR-8) - Complete**

*Backend (Weeks 1-2)*:
- Migration 025: Dashboard System schema with dashboards and metric_snapshots tables
- Dashboard data models (Dashboard, DashboardLayout, Widget, MetricSnapshot)
- Dashboard repository layer with full CRUD operations
- Metric snapshot computation service with hourly scheduler
- 13 new API endpoints:
  - GET/POST /api/dashboards - List/create dashboards with filtering
  - GET /api/dashboards/{id} - Get dashboard by ID
  - PUT /api/dashboards/{id} - Update dashboard layout and settings
  - DELETE /api/dashboards/{id} - Delete dashboard (system dashboards protected)
  - POST /api/dashboards/{id}/clone - Clone dashboard with new name
  - GET /api/metrics/snapshots - Query pre-computed metric snapshots
  - GET /api/metrics/aggregate - Single KPI values for Number Card widgets
  - GET /api/metrics/timeseries - Time-series data for Timeseries Chart widgets
  - GET /api/metrics/compare - Entity comparison for Bar Chart widgets
  - GET /api/engineers/performance - Performance table data for Table widgets
  - GET /api/metrics/health-status - System health for Status Indicator widgets
  - GET /api/alerts/timeline - Alert overlays for chart widgets
- Automated metric snapshot computation for org, team, and engineer levels
- Metrics: PR volume, team scores, engineer scores, active alerts, active goals, cycle time

*Frontend (Weeks 3-4)*:
- React dashboard system with drag-and-drop grid layout (react-grid-layout)
- 6 widget components (all presentational, dark mode, responsive):
  - Number Card - Single KPI with comparison and change indicators
  - Timeseries Chart - Line charts with Recharts, alert overlay markers
  - Bar Chart - Horizontal/vertical comparison charts
  - Table - Sortable performance tables with 6 columns
  - Status Indicator - Color-coded system health (ok/warning/critical)
  - Activity Feed - Scrollable recent events feed
- Dashboard management UI:
  - List view with templates and user dashboards
  - Create, edit, delete, clone dashboards
  - Add widgets modal with configuration
  - Drag-and-drop widget positioning
  - Responsive grid (breakpoints: lg/md/sm/xs/xxs)
- 6 TanStack Query hooks for data fetching
- Container/Presentational pattern strictly enforced

*Templates (Week 4)*:
- 4 default dashboard templates automatically created on first run:
  - Individual Contributor - Personal performance, goals, activity (7 widgets)
  - Team Lead - Team performance, member comparison, alerts (7 widgets)
  - Director - Org-wide metrics, team comparison, trends (7 widgets)
  - Planning - Sprint planning, goal tracking, burndown (7 widgets)
- All templates are system dashboards (cannot be deleted, can be cloned)

*Developer Experience*:
- TypeScript interfaces for all dashboard types
- API client functions with type safety
- Loading skeletons (no "Loading..." text)
- Error boundaries for graceful failure
- All components under 300 lines

## [1.1.0] - 2025-11-07

### Added

**Alert System (Migration 018)**
- Real-time operations alert engine for monitoring metrics and thresholds
- Alert rules with configurable conditions and thresholds
- Alert instances with severity levels (info, warning, critical)
- Alert channels for Slack, email, and browser notifications
- Alert delivery tracking with quiet hours support
- 7 new API endpoints for alert management

**Goal Tracking System (Migration 019)**
- OKR-style goals with individual, team, and org-wide support
- Goal milestones with target and current value tracking
- Progress logging with manual and automatic detection
- Goal dependencies and blockers tracking
- Goal metrics for quantitative measurement
- 9 new API endpoints for goal management

**Skill Development Tracking (Migration 020)**
- Skill taxonomy with technical, leadership, and communication categories
- Engineer skill levels with proficiency scoring (0-100)
- Evidence-based skill detection from PRs, reviews, and contributions
- Skill progression tracking over time
- Skill goals for targeted development
- Peer comparison and skill gap identification
- 3 new API endpoints for skill management

**Cost and ROI Analysis (Migration 021)**
- Cost configuration for engineers, teams, and org-wide defaults
- Feature value estimation with ARR impact and confidence levels
- Feature work item mapping to engineering effort
- ROI calculations with payback period analysis
- Engineering investment breakdown by category
- Cost efficiency metrics (cost per PR, per story point, per feature)
- 8 new API endpoints for cost and ROI tracking

**Manager Context System (Migration 022)**
- Manager notes with priority levels and visibility controls
- Context annotations for metrics and events
- Team context tracking for velocity and capacity changes
- Sentiment surveys with anonymous responses
- Sentiment analysis from multiple sources
- 8 new API endpoints for manager context

**Predictive Analytics (Migration 023)**
- Forecasting engine with multiple model types
- Sprint completion, goal completion, and velocity predictions
- What-if scenario modeling for capacity and timeline changes
- Risk predictions (attrition, timeline miss, quality degradation, burnout)
- Forecast accuracy tracking for model improvement
- 16 new API endpoints for forecasting and scenarios

**Action Tracking (Migration 024)**
- AI-generated and manual recommendations
- Action items with assignments and tracking
- Action outcomes measurement for effectiveness
- Recommendation lifecycle history
- Follow-up scheduling and reminders
- 9 new API endpoints for actions and recommendations

**Additional Features**
- Real-time metrics dashboard endpoint (GET /api/metrics/today)
- Sprint burndown chart endpoint (GET /api/planning/sprint-burndown)
- 35+ new database tables across 7 migrations
- 70+ new API endpoints total

### Changed

**Breaking Changes**
- Pagination format updated for consistency across all list endpoints
  - Old format: `{items: [], total: N, limit: N, offset: N}`
  - New format: `{items: [], pagination: {total: N, limit: N, offset: N, hasMore: boolean}}`
  - Affected endpoints: GET /api/events, GET /api/engineers, GET /api/exports
- Clients must update to handle new pagination structure

### Database Migrations
- Migration 018: Alert System (4 tables, 6 indexes)
- Migration 019: Goal Tracking System (5 tables, 7 indexes)
- Migration 020: Skill Development Tracking (5 tables, 7 indexes)
- Migration 021: Cost and ROI Analysis (7 tables, 7 indexes)
- Migration 022: Manager Context and Sentiment (6 tables, 7 indexes)
- Migration 023: Predictive Analytics (5 tables, 7 indexes)
- Migration 024: Action Tracking (5 tables, 8 indexes)

## [1.0.0] - 2025-11-04

### Added
- Identity Management System for tracking engineers across multiple data sources
- Team page at `/team` for managing engineers and resolving identities
- AI-powered identity matching with confidence scoring for suggested matches
- CSV bulk import for engineers with intelligent column mapping
- Engineer activity metrics display (pull requests, reviews, issues, commits)
- 10 new API endpoints for engineer and identity management operations
- Backfill command to retroactively link existing events to engineers
- Automatic identity resolution when ingesting events from any source
- Engineer create, read, update, and delete operations
- Unresolved identities panel with suggestion workflow
- Identity assignment and merging capabilities
- Advanced Claude Code setup with agents, skills, and hooks
- Full-stack engineer agent (Go + React/Vite)
- Auto build-checking (go vet + npm typecheck)
- Skill auto-activation system
- Version bump validation hook

### Fixed
- Fix orphaned foreign key references when assigning identities to events
- Fix DNS rebinding vulnerability by binding to 127.0.0.1 instead of localhost
- Fix activity metrics display error on Team page
- Fix API query parameter filtering for identity suggestions
- Fix CSV import implementation with proper validation

### Security
- Bind server to 127.0.0.1 instead of localhost to prevent DNS rebinding attacks
- Add CSV file validation and size limits for import operations
- Enforce proper input validation on all identity management endpoints

## [0.1.0] - 2025-11-04

### Added
- Initial project structure
- Plugin architecture (Source, Destination, Processor)
- SQLite database with event model
- Anonymization system (sequential, uuid, hash strategies)
- AES-256-GCM encryption for API keys
- Audit logging for data exports
- Plugin SDK for Go
- Core plugins: github, csv-import, google-sheets-export, ai-insights
- Single binary distribution (Go + embedded React)
- Localhost-only security model (port 3847)

### Architecture
- Go 1.21+ backend with stdlib net/http
- SQLite database (modernc.org/sqlite)
- JSON-RPC plugin communication over stdin/stdout
- Plugin discovery: ~/.engineerdna/plugins/ and ./plugins/
- Master key in OS keychain or env var

[Unreleased]: https://github.com/username/engineerdna/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/username/engineerdna/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/username/engineerdna/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/username/engineerdna/releases/tag/v0.1.0
