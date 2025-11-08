# Changelog

All notable changes to EngineerDNA will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-11-08

### What is EngineerDNA?

**EngineerDNA is an AI Chief of Staff for engineering leaders** - a local desktop application that transforms engineering data into actionable insights.

**Core Value:**
- **For Engineering Managers**: "Who needs my attention this week and what should I talk to them about?"
- **For Directors/VPs**: "Is engineering slowing down, and if so, why?"
- **For Product Leaders**: "Can we ship feature X by date Y?" (with data-backed answers)

**Key Principles:**
- Privacy-first: Runs locally on your machine (localhost:3847), data never leaves your laptop
- Single binary for macOS, Linux, Windows - no installation complexity
- Extensible plugin architecture for any data source
- Bring your own API keys (BYOK) for AI analysis
- 100% open source, free forever

---

### Added

**Core Platform**

*Technology*:
- Go 1.21+ backend with SQLite database
- React 18 + TypeScript frontend with Vite
- TanStack Query for data fetching
- Recharts for visualizations
- Tailwind CSS for styling
- Single binary distribution with embedded frontend

*Data Model*:
- Universal multi-modal schema supporting three data types: Events, Metrics, Attributes
- metric_values table for time-series measurements with granularity and dimensions
- entity_attributes table for entity facts with temporal validity tracking
- correlations and correlation_values tables for pre-computed cross-data-type relationships
- Dashboard widgets display fast historical trends (cost per feature over 12 months)

*Security & Privacy*:
- Localhost-only binding (127.0.0.1:3847)
- AES-256-GCM encryption for API keys and secrets
- Master key storage in OS keychain with environment variable fallback
- Three anonymization strategies: sequential (User_A, User_B), UUID, hash
- Bidirectional anonymization mapping for deanonymization
- Comprehensive audit logging for all data exports and AI processing

**Plugin System**

*Architecture*:
- Three plugin types: Source (events/metrics/attributes IN), Processor (transforms/analysis), Destination (data OUT)
- Unified source.sync response: returns events, metrics, and/or attributes in single call
- Event normalization: plugins declare event types with normalized mappings (e.g., GitHub PR + GitLab MR → code_review)
- JSON-RPC communication over stdin/stdout
- Subprocess isolation with configurable timeouts (30s default)
- Plugin discovery from ~/.engineerdna/plugins/ and ./plugins/
- Plugin SDK for Go developers with examples in all plugin directories
- Automatic anonymization for processor plugins sending to external APIs
- Simple plugin.json manifest with provides_event_types for multi-source support

*Built-in Plugins*:
- **GitHub Source**: PRs, issues, commits, reviews via GraphQL API (supports story points from labels)
- **AWS Costs Source**: Daily infrastructure costs and resource usage via Cost Explorer API
- **CSV Import Source**: Flexible column mapping for events, metrics, and attributes
- **AI Insights Processor**: Multi-provider support (Anthropic Claude, OpenAI GPT, Ollama) with BYOK and mandatory anonymization
- **Google Sheets Export Destination**: OAuth 2.0 integration for spreadsheet exports
- **PDF Export Destination**: Formatted reports with charts and tables
- **Markdown Export Destination**: Documentation and changelog generation

**Identity Management**

- Canonical identity system for tracking engineers across data sources
- AI-powered identity matching with confidence scoring
- Manual identity assignment and merging
- CSV bulk import with intelligent column mapping
- Team management interface at /team route
- Activity metrics per engineer (PRs, reviews, issues, commits)
- Backfill command to retroactively link events to engineers
- 10 API endpoints for engineer and identity operations

**Alert System**

- Real-time operations monitoring with configurable rules
- Alert types: threshold breach, change detection, anomaly detection
- Severity levels: info, warning, critical
- Multiple delivery channels: Slack, email, browser notifications
- Lifecycle management: fired, acknowledged, snoozed, dismissed, resolved
- Quiet hours support for off-hours
- Alert delivery tracking
- 7 API endpoints for alert management

**Goal Tracking**

- OKR-style goals with individual, team, and org-wide support
- Goal milestones with target vs current tracking
- Progress logging with manual and automatic detection
- Goal dependencies and blocker tracking
- Goal metrics for quantitative measurement
- Trajectory tracking: on-track, at-risk, off-track
- Performance review integration
- 9 API endpoints for goal management

**Skill Development**

- Skill taxonomy: technical, leadership, communication categories
- Proficiency scoring (0-100) for each engineer
- Evidence-based skill detection from PRs, reviews, contributions
- Skill progression tracking over time
- Skill goals for targeted development
- Peer comparison and skill gap identification
- Trajectory detection: improving, stable, declining
- 3 API endpoints for skill management

**Cost & ROI Analysis**

- Cost configuration by engineer, team, or org-wide
- Fully-loaded cost tracking (salary + benefits + overhead)
- Feature value estimation with ARR impact
- Feature work item mapping to engineering effort
- ROI calculations with payback period analysis
- Engineering investment breakdown (features, tech debt, support, operations)
- Cost efficiency metrics: per PR, per story point, per feature
- 8 API endpoints for cost and ROI tracking

**Manager Context**

- Manager notes with priority levels and visibility controls
- Context annotations for metrics and events
- Team context tracking for velocity and capacity changes
- Sentiment surveys with anonymous responses
- Multi-source sentiment analysis with confidence scoring
- Mood tracking over time
- Qualitative data capture for performance reviews
- 8 API endpoints for manager context

**Predictive Analytics**

- Multiple forecasting models: linear regression, Monte Carlo, historical
- Sprint completion predictions
- Goal completion predictions
- Velocity forecasting
- What-if scenario modeling for capacity and timeline changes
- Risk predictions: attrition, timeline miss, quality degradation, burnout
- Forecast accuracy tracking for continuous improvement
- Confidence intervals for all predictions
- 16 API endpoints for forecasting and scenarios

**Action Tracking**

- AI-generated and manual recommendations
- Action item assignments with owner tracking
- Action outcome measurement for effectiveness
- Recommendation lifecycle: pending, in-progress, completed, dismissed
- Follow-up scheduling with reminders
- Priority levels and expiration dates
- Close the feedback loop on insights
- 9 API endpoints for actions and recommendations

**Dashboard Builder**

*Backend*:
- Flexible dashboard schema with JSON layout storage
- Dashboard CRUD operations with templates
- Metric snapshot computation service with hourly scheduler
- Pre-computed snapshots for org, team, and engineer levels
- Snapshot metrics: PR volume, team scores, engineer scores, alerts, goals, cycle time

*Frontend*:
- Drag-and-drop dashboard builder with react-grid-layout
- Six widget types:
  - Number Card: Single KPI with trend comparison
  - Timeseries Chart: Line charts with alert overlay markers
  - Bar Chart: Horizontal/vertical entity comparison
  - Table: Sortable performance tables
  - Status Indicator: Color-coded system health
  - Activity Feed: Scrollable recent events
- Four default templates:
  - Individual Contributor: Personal performance and goals
  - Team Lead: Team performance and member comparison
  - Director: Org-wide metrics and team comparison
  - Planning: Sprint planning and goal tracking
- Dashboard management: create, edit, delete, clone
- Widget configuration modal
- Responsive grid with multiple breakpoints
- Dark mode support
- Loading skeletons and error boundaries

*Data Layer*:
- 13 dashboard and metrics API endpoints
- TanStack Query hooks for data fetching
- Container/presentational component pattern
- TypeScript type safety throughout
- Optimized queries for fast dashboard loads

**Frontend Application**

*Pages*:
- Dashboard: Metrics overview with customizable widgets
- Plugins: Plugin configuration and management
- Events: Filterable event list
- Team: Engineer management and identity resolution
- Settings: System configuration

*Components*:
- Plugin card with status display
- Plugin configuration modal
- Anonymization settings
- Event list with filtering
- Insights panel for AI-generated recommendations
- Audit log viewer
- Velocity charts
- Activity breakdown charts

*Developer Experience*:
- Type-safe API client with full coverage
- Comprehensive TypeScript type definitions
- Error handling with user-friendly messages
- Request/response interceptors

**Production Features**

- First-run onboarding flow
- Step-by-step setup wizard
- Plugin configuration guidance
- Graceful error recovery
- User-friendly error messages
- Plugin error diagnostics
- Multiple export formats
- Export scheduling
- Automated exports
- Dashboard loads in <2s with 10k+ events
- Optimized database queries
- Efficient metric computation

**Distribution & Documentation**

- Cross-platform builds: Linux (amd64), macOS (amd64/arm64), Windows (amd64)
- GitHub Actions release pipeline
- Automated version bumping
- Single binary distribution
- Comprehensive README
- Plugin development guide
- API documentation
- Troubleshooting guide

**Developer Tools**

- Advanced Claude Code agent system with 8 specialized agents
- Full-stack engineer agent for Go + React development
- Quality, security, integration-checker, ui-tester, docs, advisor agents
- Auto build-checking: go vet + npm typecheck
- Skill auto-activation system
- 20+ hooks for workflow automation
- Database migration enforcement
- Automated dev server restart script
- Frontend and backend hot reloading
- Git hooks for code quality
- Version bump validation
- Dead code detection

### Database

**34 migrations total**:
- Core schema: events, plugins, engineers, anonymization, audit log
- Alert System: 4 tables, 6 indexes
- Goal Tracking: 5 tables, 7 indexes
- Skill Development: 5 tables, 7 indexes
- Cost & ROI: 7 tables, 7 indexes
- Manager Context: 6 tables, 7 indexes
- Predictive Analytics: 5 tables, 7 indexes
- Action Tracking: 5 tables, 8 indexes
- Dashboard System: 2 tables, 4 indexes
- Universal Schema: metric_values, entity_attributes, correlations, correlation_values

### API Endpoints

**90+ endpoints organized by category**:
- Core: 10 (health, version, events, metrics)
- Plugins: 7 (list, configure, sync, test, status, enable, disable)
- Identity: 10 (engineers, identities, resolution, bulk import)
- Anonymization: 7 (policies, mappings, audit, reset, export)
- Insights: 4 (list, generate, dismiss)
- Exports: 4 (list, create, status)
- Alerts: 7 (rules, instances, channels, acknowledge, snooze, dismiss)
- Goals: 9 (goals, milestones, progress, dependencies, metrics)
- Skills: 3 (taxonomy, engineer skills, evidence)
- Cost & ROI: 8 (configuration, values, work items, calculations)
- Manager Context: 8 (notes, annotations, surveys, sentiment)
- Predictive: 16 (forecasts, scenarios, risks, accuracy)
- Actions: 9 (recommendations, actions, outcomes, follow-ups)
- Dashboards: 13 (dashboards, widgets, metrics, performance)

[Unreleased]: https://github.com/username/engineerdna/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/username/engineerdna/releases/tag/v0.1.0
