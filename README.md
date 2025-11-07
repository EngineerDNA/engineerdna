# EngineerDNA

Engineering metrics and insights platform with privacy-first anonymization and bring-your-own-key (BYOK) encryption.

## Deployment Model

**EngineerDNA V1 is a local desktop application** that runs on `localhost:3847` - similar to Jupyter Notebook, Grafana, or Docker Desktop.

- Runs on your laptop/desktop only
- Single user (you)
- No network exposure
- No authentication needed (you authenticated when you logged in)

**For network/team deployments:** V2 will add authentication, authorization, and TLS.

## Features

### Core Platform

- **Identity Management**: Track and manage engineers across multiple data sources
  - AI-powered identity resolution with confidence scoring
  - Automatic identity linking when ingesting events
  - Team page for managing engineers and resolving identities
  - CSV bulk import with intelligent column mapping
  - Activity metrics per engineer (PRs, reviews, issues, commits)

- **Plugin Architecture**: Extensible system with three types of plugins:
  - **Source**: Bring data IN (GitHub, CSV, manual entry)
  - **Destination**: Send data OUT (Google Sheets, Slack, PDF)
  - **Processor**: Transform and analyze data (AI insights, custom metrics)

- **Privacy First**:
  - Built-in anonymization with multiple strategies (sequential, hash, UUID)
  - Bidirectional mapping for deanonymization
  - Required anonymization for external API calls
  - Full audit trail of data access

- **Security**:
  - AES-256-GCM encryption for API keys and secrets
  - Bring-your-own-key (BYOK) support
  - Data stays local unless explicitly exported
  - No telemetry or phone-home

- **Self-Hosted**: Single binary distribution, runs anywhere
- **SQLite Backend**: Zero external dependencies
- **Web UI**: Built-in dashboard for visualization and configuration

### PDR-7 Features (v1.1.0)

- **Alert System**: Real-time operations monitoring
  - Configurable alert rules with thresholds and severity levels
  - Multiple delivery channels (Slack, email, browser)
  - Alert lifecycle management (acknowledge, snooze, dismiss)
  - Quiet hours support for notification management

- **Goal Tracking**: OKR-style goal management
  - Individual, team, and org-wide goals
  - Milestone tracking with automatic progress detection
  - Goal dependencies and blocker identification
  - Success criteria and quantitative metrics

- **Skill Development**: Evidence-based skill tracking
  - Automatic skill detection from PRs, reviews, and contributions
  - Proficiency scoring (0-100) with trajectory tracking
  - Skill gap identification and peer comparison
  - Targeted development goals

- **Cost & ROI Analysis**: Business intelligence for engineering
  - Cost configuration by engineer, team, and role
  - Feature value estimation with confidence levels
  - ROI calculations with payback period
  - Engineering investment breakdown by category
  - Cost efficiency metrics

- **Manager Context**: Qualitative data to supplement metrics
  - Manager notes with visibility controls
  - Context annotations for metrics and events
  - Sentiment surveys with anonymous responses
  - Team context tracking

- **Predictive Analytics**: Forecast future outcomes
  - Sprint and goal completion predictions
  - What-if scenario modeling
  - Risk predictions (attrition, timeline miss, quality degradation)
  - Forecast accuracy tracking

- **Action Tracking**: Close the feedback loop
  - AI-generated and manual recommendations
  - Action item tracking with assignments
  - Outcome measurement for effectiveness
  - Follow-up scheduling and reminders

- **Dashboard Builder**: Customizable dashboards with 6 widget types (Number Card, Timeseries Chart, Bar Chart, Table, Status Indicator, Activity Feed). Includes 4 default templates: Individual Contributor, Team Lead, Director, and Planning. Pre-computed metric snapshots for performance.

## Dashboard Builder

The Dashboard Builder provides a flexible, drag-and-drop interface for creating custom engineering dashboards tailored to different personas and use cases.

### Overview

Create personalized views of your engineering metrics with:
- **4 pre-built templates** optimized for different roles (IC, Team Lead, Director, Planning)
- **6 widget types** covering KPIs, trends, comparisons, and real-time data
- **Drag-and-drop layout** with responsive grid system
- **Real-time data** auto-refreshing every 30 seconds
- **Clone and customize** templates to fit your specific needs

### Templates

EngineerDNA includes 4 production-ready dashboard templates:

| Template | Persona | Widgets | Focus |
|----------|---------|---------|-------|
| **Individual Contributor** | Engineer | 7 widgets | Personal performance metrics, PRs merged, cycle time, reviews given, activity feed |
| **Team Lead** | Engineering Manager | 7 widgets | Team performance, sprint progress, top performers, team health, active alerts, velocity trends |
| **Director** | Director/VP of Engineering | 7 widgets | Org-wide metrics, team comparison, executive summary, strategic KPIs, cross-team insights |
| **Planning** | Product/Engineering Leadership | 7 widgets | Sprint burndown, capacity planning, feature timelines, resource allocation, upcoming deadlines |

### Widget Types

| Widget Type | Description | Data Source | Use Cases |
|-------------|-------------|-------------|-----------|
| **Number Card** | Single KPI with trend indicator | Metric aggregates | PR count, cycle time, velocity, score |
| **Timeseries Chart** | Line/area chart over time | Metric timeseries | Throughput trends, quality over time, sprint velocity |
| **Bar Chart** | Compare entities side-by-side | Metric comparisons | Team performance, engineer rankings, sprint capacity |
| **Table** | Sortable data grid | Engineer/team performance | Top performers, detailed breakdowns, skill matrices |
| **Status Indicator** | Health checks and system status | Health checks | System health, plugin status, data freshness |
| **Activity Feed** | Real-time event stream | Recent events | Latest PRs, reviews, commits, issues |

### Quick Start

**View Templates:**
1. Navigate to Dashboards page
2. Select a template from the dropdown (Individual Contributor, Team Lead, Director, or Planning)
3. Explore pre-configured widgets and layouts

**Create Custom Dashboard:**
1. Click "Create Dashboard" button
2. Enter name and description
3. Click "Edit Layout" to enable drag-and-drop
4. Click "Add Widget" to add Number Cards, Charts, Tables, etc.
5. Configure widget data sources and visualization settings
6. Drag widgets to arrange layout
7. Click "Done Editing" to save

**Clone Template:**
1. Select a template dashboard
2. Click "Clone Template" button
3. New dashboard created as "Template Name (Copy)"
4. Customize cloned dashboard to your needs
5. Edit widget configuration, layout, and data sources

### Widget Configuration

Each widget can be configured with:

**Data Source Parameters:**
- `metric`: Which metric to display (prs_merged, cycle_time_days, total_score, etc.)
- `entity_type`: Scope (engineer, team, org)
- `entity_id`: Specific entity to filter by (optional)
- `time_range`: Time window (7d, 30d, 90d)
- `comparison_period`: Previous period for trend analysis
- `sort_by`: Table column to sort by
- `limit`: Number of rows/items to display

**Visualization Options:**
- `primary_color`: Chart color
- `show_legend`: Display legend (true/false)
- `show_grid`: Display grid lines (true/false)
- `y_axis_label`: Y-axis label text
- `number_format`: Number formatting (decimal, percentage, currency)
- `time_format`: Date/time formatting
- `threshold`: Good/warning/critical thresholds for status indicators

### API Usage

**List Dashboards:**
```bash
GET /api/dashboards
GET /api/dashboards?persona=team_lead
GET /api/dashboards?is_template=true
```

**Create Dashboard:**
```bash
POST /api/dashboards
{
  "name": "My Custom Dashboard",
  "description": "Engineering metrics for my team",
  "layout": "{\"widgets\":[...]}"
}
```

**Update Dashboard:**
```bash
PUT /api/dashboards/:id
{
  "name": "Updated Name",
  "layout": "{\"widgets\":[...]}"
}
```

**Clone Dashboard:**
```bash
POST /api/dashboards/:id/clone
{
  "name": "Cloned Dashboard Name"
}
```

**Fetch Widget Data:**
```bash
GET /api/metrics/aggregate?metric_type=prs_merged&entity_type=team&time_range=30d
GET /api/metrics/timeseries?metric_type=cycle_time_days&entity_type=org&start_date=2024-01-01&end_date=2024-01-31
GET /api/metrics/compare?metric_type=total_score&entity_type=engineer&entity_ids[]=1&entity_ids[]=2
GET /api/engineers/performance?team_id=abc&time_range=30d&limit=10
GET /api/metrics/health-status
```

### Best Practices

**Dashboard Design:**
- Start with a template and customize incrementally
- Limit dashboards to 6-8 widgets for clarity
- Group related metrics together
- Use consistent time ranges across widgets
- Place most important metrics in top-left position

**Widget Selection:**
- Number Cards: Key metrics you check daily
- Timeseries Charts: Trends and patterns over time
- Bar Charts: Comparisons and rankings
- Tables: Detailed drill-down data
- Status Indicators: System health and alerts
- Activity Feed: Real-time awareness

**Performance:**
- Dashboards auto-refresh every 30 seconds
- Pre-computed metric snapshots ensure fast loading
- Limit table widgets to 10-20 rows for responsiveness
- Use time ranges appropriate to your data volume (7d for daily checks, 90d for trends)

## Quick Start

### Installation

#### Quick Install (Recommended)

Install with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh | sh
```

Or download and inspect first:

```bash
curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

The script automatically:
- Detects your OS and architecture (Linux, macOS, Windows on amd64/arm64)
- Downloads the latest release from GitHub
- Installs to `/usr/local/bin` (customizable with `INSTALL_DIR`)
- Verifies installation

#### Custom Installation Directory

```bash
# Install to a custom directory
INSTALL_DIR=$HOME/.local/bin curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh | sh

# Install a specific version
ENGINEERDNA_VERSION=v1.2.3 curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh | sh
```

#### Homebrew (Coming Soon)

```bash
brew tap engineerdna/tap
brew install engineerdna
```

#### NPM (Coming Soon)

```bash
npm install -g engineerdna
```

#### From Source

For development or if you prefer building from source:

```bash
# Clone repository
git clone https://github.com/engineerdna/engineerdna.git
cd engineerdna

# Build and install
make build
sudo make install

# Or use the development install script
./scripts/install-dev.sh
```

#### Installation Environment Variables

The install script supports these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `INSTALL_DIR` | Installation directory | `/usr/local/bin` |
| `ENGINEERDNA_VERSION` | Version to install | `latest` |

Examples:

```bash
# Install to home directory
INSTALL_DIR=$HOME/.local/bin ./install.sh

# Install specific version
ENGINEERDNA_VERSION=v1.2.3 ./install.sh

# Combine both
INSTALL_DIR=/opt/bin ENGINEERDNA_VERSION=v1.2.3 ./install.sh
```

### Initialize

```bash
engineerdna init
```

This creates:
- `~/.engineerdna/` directory
- SQLite database
- Plugin directory
- Encryption key (displayed once)

### Start Server

```bash
engineerdna serve
```

Server runs at `http://localhost:3847`

### Configuration

Set the master encryption key (optional but recommended):

```bash
export ENGINEERDNA_MASTER_KEY=<base64-encoded-32-byte-key>
```

If not set, a new key is generated on each run.

## Docker Deployment

### Localhost-Only (Recommended - Secure)

This configuration binds to `127.0.0.1` only - accessible only from your machine:

```bash
# Start (default configuration - secure)
docker-compose up -d

# Access
http://localhost:3847
```

**Security:** Same as native binary - secured by OS login + localhost-only binding.

### Network Exposure (Not Recommended - Risky)

**WARNING:** This exposes EngineerDNA to your local network with NO authentication.

```bash
# Start with network exposure (use with caution)
docker-compose -f docker-compose.network.yml up -d

# Access from any device on your network
http://<your-machine-ip>:3847
```

**EngineerDNA V1 has NO built-in network security:**
- **NO authentication** - Anyone with network access can use the application
- **NO authorization** - All data is visible to all users
- **NO TLS/HTTPS** - All traffic is transmitted unencrypted
- **NO access control** - Anyone can view, modify, export data

**Only use network exposure if:**
- You are on a trusted, private network
- You trust everyone on your local network
- You have firewall rules restricting access
- You understand the security implications

**For production use, wait for V2 which includes:**
- JWT authentication
- Role-based authorization
- TLS/HTTPS support
- API key management

### Quick Start with Docker

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Access application
# From local machine:
open http://localhost:3847

# From other devices on network:
open http://<your-machine-ip>:3847

# Stop (keeps data)
docker-compose down

# Stop and remove all data
docker-compose down -v
```

### Environment Variables

Configure via environment variables in `docker-compose.yml`:

```yaml
services:
  engineerdna:
    environment:
      # WARNING: 0.0.0.0 exposes to local network
      - ENGINEERDNA_HOST=0.0.0.0
      - ENGINEERDNA_PORT=3847
      # Optional: Set master encryption key
      - ENGINEERDNA_MASTER_KEY=your-base64-encoded-32-byte-key
```

Generate a master key:
```bash
openssl rand -base64 32
```

### Security Hardening

If you must use Docker deployment, implement these measures:

**1. Firewall Rules** (Recommended)
```bash
# Linux - Allow specific IP only
sudo iptables -A INPUT -p tcp --dport 3847 -s 192.168.1.100 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3847 -j DROP
```

**2. Reverse Proxy with Authentication**
Use nginx, Caddy, or Traefik to add authentication and TLS.

**3. VPN-Only Access**
Only expose on VPN network, require VPN connection to access.

**4. SSH Tunnel** (Most Secure)
Keep localhost-only binding and use SSH tunneling:
```bash
# On docker-compose.yml, set:
- ENGINEERDNA_HOST=127.0.0.1
# Bind to localhost only:
ports:
  - "127.0.0.1:3847:3847"

# Access remotely via SSH tunnel:
ssh -L 3847:localhost:3847 user@docker-host
# Then open http://localhost:3847 on local machine
```

For comprehensive Docker documentation including monitoring, troubleshooting, and advanced configuration, see comments in `docker-compose.yml` and `Dockerfile`.

## Development

### Quick Restart Script

For active development, use the restart script to cleanly stop and start the server:

```bash
# Basic restart (just restart the server)
./scripts/restart-dev-server.sh

# Rebuild frontend and backend, then restart
./scripts/restart-dev-server.sh --rebuild

# Include frontend dev server (Vite on port 5173)
./scripts/restart-dev-server.sh --frontend

# Rebuild everything and start both servers
./scripts/restart-dev-server.sh --rebuild --frontend
```

The script automatically:
- Kills any processes using port 3847 (backend) or 5173 (frontend)
- Cleans up zombie processes
- Optionally rebuilds the binary and frontend
- Starts the server(s) in the background
- Verifies they started successfully
- Outputs log file locations

Logs are written to:
- Backend: `/tmp/engineerdna-backend.log`
- Frontend: `/tmp/engineerdna-frontend.log`

### Manual Development Workflow

```bash
# Backend development
make build          # Build Go binary
./bin/engineerdna serve

# Frontend development (with hot reload)
cd frontend
npm run dev         # Starts Vite dev server on port 5173
# Note: Vite proxies /api calls to localhost:3847

# Production build
cd frontend && npm run build
make build          # Embeds frontend/dist/ in binary
```

### Pre-commit Checks

Before committing, ensure all checks pass:

```bash
# Backend
go vet ./...
staticcheck ./...
go test ./...
make build

# Frontend
cd frontend
npm run lint
npm run typecheck
npm run build
```

## Plugin Development

See [plugins/plugin-sdk/](plugins/plugin-sdk/) for the SDK.

### Example Plugin

```go
package main

import "github.com/engineerdna/engineerdna/plugins/plugin-sdk"

type MyPlugin struct {
    config map[string]string
}

func (p *MyPlugin) Info() sdk.PluginInfo {
    return sdk.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Type:        "source",
        Description: "My custom plugin",
    }
}

func (p *MyPlugin) Configure(config map[string]string) error {
    p.config = config
    return nil
}

func (p *MyPlugin) Health() sdk.HealthResult {
    return sdk.HealthResult{Healthy: true, Message: "OK"}
}

func (p *MyPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
    // Implement sync logic
    return sdk.SyncResult{Events: []sdk.Event{}}, nil
}

func main() {
    sdk.Serve(&MyPlugin{})
}
```

### Plugin Metadata

Create `plugin.json`:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "type": "source",
  "description": "My custom plugin",
  "author": "Your Name",
  "config_fields": [
    {
      "name": "api_key",
      "type": "password",
      "required": true,
      "description": "API Key",
      "secret": true
    }
  ],
  "anonymization": {
    "required": false,
    "strategy": "sequential",
    "fields": ["actor", "author"]
  }
}
```

### Build and Install Plugin

```bash
cd plugins/my-plugin
go build -o my-plugin main.go
```

Place in `~/.engineerdna/plugins/my-plugin/` or `./plugins/my-plugin/`

## CLI Commands

```bash
engineerdna init          # Initialize application
engineerdna serve         # Start HTTP server
engineerdna sync          # Sync all source plugins
engineerdna version       # Show version
engineerdna help          # Show help
```

## API Endpoints

### Health & System
- `GET /api/health` - Health check
- `GET /api/version` - Version info
- `GET /api/system/info` - System information

### Events
- `GET /api/events` - List events (supports filtering, pagination)
- `GET /api/events/:id` - Get event by ID

### Engineers
- `GET /api/engineers` - List all engineers (paginated)
- `GET /api/engineers/:id` - Get engineer by ID
- `POST /api/engineers` - Create new engineer
- `PUT /api/engineers/:id` - Update engineer
- `DELETE /api/engineers/:id` - Delete engineer
- `POST /api/engineers/import` - Bulk import engineers from CSV

### Identity Resolution
- `GET /api/identity/unresolved` - List unresolved identities
- `GET /api/identity/suggestions` - Get AI-powered match suggestions
- `POST /api/identity/resolve` - Assign identity to engineer
- `POST /api/identity/merge` - Merge multiple identities
- `POST /api/identity/ignore` - Ignore an identity
- `GET /api/identity/ignored` - List ignored identities

### Plugins
- `GET /api/plugins` - List all plugins
- `GET /api/plugins/:name` - Get plugin details
- `POST /api/plugins/:name/configure` - Configure plugin
- `POST /api/plugins/:name/sync` - Sync plugin data
- `POST /api/plugins/:name/test` - Test plugin connection
- `PUT /api/plugins/:name/enable` - Enable plugin
- `PUT /api/plugins/:name/disable` - Disable plugin

### Anonymization
- `GET /api/anonymization/policies` - List policies
- `GET /api/anonymization/mappings` - List mappings
- `GET /api/anonymization/audit` - Audit log

### Metrics & Insights
- `GET /api/metrics/throughput` - Throughput metrics
- `GET /api/metrics/cycle-time` - Cycle time metrics
- `GET /api/metrics/today` - Real-time metrics for today
- `GET /api/insights` - List insights
- `POST /api/insights` - Generate insights

### Dashboard Management
- `GET /api/dashboards` - List dashboards (filter by persona, templates)
- `POST /api/dashboards` - Create dashboard
- `GET /api/dashboards/:id` - Get dashboard by ID
- `PUT /api/dashboards/:id` - Update dashboard layout and settings
- `DELETE /api/dashboards/:id` - Delete dashboard (system dashboards protected)
- `POST /api/dashboards/:id/clone` - Clone dashboard with new name

### Dashboard Widgets
- `GET /api/metrics/aggregate` - Single KPI values for Number Card widgets
- `GET /api/metrics/timeseries` - Time-series data for Timeseries Chart widgets
- `GET /api/metrics/compare` - Entity comparison for Bar Chart widgets
- `GET /api/metrics/snapshots` - Query pre-computed metric snapshots
- `GET /api/engineers/performance` - Performance table data for Table widgets
- `GET /api/metrics/health-status` - System health for Status Indicator widgets
- `GET /api/alerts/timeline` - Alert overlays for chart widgets

### Exports
- `GET /api/exports` - List exports (paginated)
- `POST /api/exports` - Create export
- `GET /api/exports/schedules` - List export schedules
- `GET /api/exports/schedules/:id` - Get export schedule by ID

### Teams
- `GET /api/teams` - List all teams
- `GET /api/teams/:id` - Get team by ID
- `POST /api/teams` - Create new team
- `PUT /api/teams/:id` - Update team
- `DELETE /api/teams/:id` - Delete team
- `GET /api/teams/hierarchy` - Get team hierarchy
- `GET /api/teams/org/scorecard` - Get org-wide scorecard

### Roles & Scoring
- `GET /api/roles` - List all roles
- `GET /api/roles/:id` - Get role by ID
- `POST /api/roles` - Create new role
- `PUT /api/roles/:id` - Update role
- `GET /api/scoring/weights` - Get scoring weights
- `PUT /api/scoring/weights` - Update scoring weights
- `POST /api/scoring/calculate` - Calculate performance score

### Performance
- `GET /api/performance/individual/:id` - Individual performance metrics
- `GET /api/performance/team/:id` - Team performance metrics

### Promotion Signals
- `GET /api/promotion/signals` - List promotion signals
- `GET /api/promotion/signals/:id` - Get signals for engineer
- `POST /api/promotion/detect` - Detect promotion signals
- `POST /api/promotion/dismiss/:id` - Dismiss promotion signal

### Briefings
- `GET /api/briefing/weekly/:id` - Weekly briefing for engineer
- `GET /api/briefing/history/:id` - Briefing history for engineer
- `GET /api/briefing/org` - Org-wide briefing

### Planning
- `GET /api/planning/sprints` - List sprints
- `GET /api/planning/sprints/:id` - Get sprint by ID
- `POST /api/planning/sprints` - Create sprint
- `PUT /api/planning/sprints/:id` - Update sprint
- `GET /api/planning/sprint-burndown` - Sprint burndown chart
- `GET /api/planning/velocity/:id` - Team velocity
- `POST /api/planning/estimate` - Estimate completion
- `GET /api/planning/capacity` - Team capacity

### Alerts (v1.1.0)
- `GET /api/alerts/rules` - List alert rules
- `GET /api/alerts/rules/:id` - Get alert rule by ID
- `POST /api/alerts/rules` - Create alert rule
- `PUT /api/alerts/rules/:id` - Update alert rule
- `DELETE /api/alerts/rules/:id` - Delete alert rule
- `GET /api/alerts/instances` - List alert instances
- `GET /api/alerts/instances/:id` - Get alert instance by ID
- `POST /api/alerts/instances/:id/acknowledge` - Acknowledge alert
- `POST /api/alerts/instances/:id/snooze` - Snooze alert
- `POST /api/alerts/instances/:id/dismiss` - Dismiss alert
- `GET /api/alerts/channels` - List alert channels
- `GET /api/alerts/channels/:id` - Get alert channel by ID
- `POST /api/alerts/channels` - Create alert channel
- `PUT /api/alerts/channels/:id` - Update alert channel
- `DELETE /api/alerts/channels/:id` - Delete alert channel

### Goals (v1.1.0)
- `GET /api/goals` - List goals
- `GET /api/goals/:id` - Get goal by ID
- `POST /api/goals` - Create goal
- `PUT /api/goals/:id` - Update goal
- `DELETE /api/goals/:id` - Delete goal
- `GET /api/goals/:id/milestones` - List milestones for goal
- `POST /api/goals/:id/milestones` - Create milestone
- `PUT /api/goals/milestones/:id` - Update milestone
- `GET /api/goals/:id/progress` - Get goal progress
- `POST /api/goals/:id/progress` - Log progress update
- `GET /api/goals/:id/dependencies` - List goal dependencies
- `POST /api/goals/:id/dependencies` - Add goal dependency
- `GET /api/goals/summary` - Get goals summary

### Skills (v1.1.0)
- `GET /api/skills` - List skills
- `POST /api/skills` - Create skill
- `GET /api/engineers/:id/skills` - Get engineer skills
- `POST /api/engineers/:id/skills` - Update engineer skill
- `GET /api/skills/evidence` - List skill evidence
- `POST /api/skills/detect` - Detect skills from events
- `GET /api/skills/progression` - Get skill progression history

### Cost & ROI (v1.1.0)
- `GET /api/cost/configuration` - Get cost configuration
- `POST /api/cost/configuration` - Update cost configuration
- `GET /api/cost/team/:id` - Get team cost
- `GET /api/cost/engineer/:id` - Get engineer cost
- `GET /api/features` - List features
- `GET /api/features/:id` - Get feature by ID
- `POST /api/features` - Create feature
- `PUT /api/features/:id` - Update feature
- `GET /api/features/:id/roi` - Calculate feature ROI
- `GET /api/roi/report` - ROI report
- `GET /api/investment/breakdown` - Investment breakdown
- `GET /api/cost/efficiency` - Cost efficiency metrics

### Manager Context (v1.1.0)
- `GET /api/notes` - List manager notes
- `GET /api/notes/:id` - Get note by ID
- `POST /api/notes` - Create note
- `PUT /api/notes/:id` - Update note
- `DELETE /api/notes/:id` - Delete note
- `GET /api/context/annotations` - List context annotations
- `POST /api/context/annotations` - Create annotation

### Sentiment (v1.1.0)
- `GET /api/surveys` - List surveys
- `GET /api/surveys/:id` - Get survey by ID
- `POST /api/surveys` - Create survey
- `PUT /api/surveys/:id` - Update survey
- `POST /api/surveys/:id/responses` - Submit survey response
- `GET /api/sentiment/team/:id` - Team sentiment analysis
- `GET /api/sentiment/engineer/:id` - Engineer sentiment analysis

### Forecasts (v1.1.0)
- `GET /api/forecasts` - List forecasts
- `GET /api/forecasts/:id` - Get forecast by ID
- `POST /api/forecasts` - Create forecast
- `GET /api/forecasts/sprint/:id` - Sprint completion forecast
- `GET /api/forecasts/goal/:id` - Goal completion forecast
- `GET /api/forecasts/team/:id` - Team velocity forecast
- `GET /api/forecasts/accuracy` - Forecast accuracy metrics
- `POST /api/forecasts/accuracy/record` - Record forecast accuracy

### Risk Predictions (v1.1.0)
- `GET /api/risks` - List risks
- `GET /api/risks/engineer/:id` - Engineer-specific risks
- `GET /api/risks/team/:id` - Team-specific risks
- `POST /api/risks/predict/attrition` - Predict attrition risk
- `POST /api/risks/predict/timeline` - Predict timeline miss risk
- `POST /api/risks/predict/quality` - Predict quality degradation risk

### Scenarios (v1.1.0)
- `GET /api/scenarios` - List scenarios
- `GET /api/scenarios/:id` - Get scenario by ID
- `POST /api/scenarios` - Create scenario
- `PUT /api/scenarios/:id` - Update scenario
- `POST /api/scenarios/simulate/:id` - Simulate scenario
- `POST /api/scenarios/compare` - Compare scenarios

### Actions & Recommendations (v1.1.0)
- `GET /api/recommendations` - List recommendations
- `GET /api/recommendations/pending` - List pending recommendations
- `GET /api/recommendations/:id` - Get recommendation by ID
- `POST /api/recommendations` - Create recommendation
- `PUT /api/recommendations/:id` - Update recommendation
- `GET /api/actions` - List actions
- `GET /api/actions/:id` - Get action by ID
- `POST /api/actions` - Create action
- `PUT /api/actions/:id` - Update action
- `POST /api/actions/:id/outcomes` - Record action outcome
- `GET /api/actions/effectiveness-report` - Action effectiveness report
- `GET /api/follow-ups` - List follow-ups
- `GET /api/follow-ups/:id` - Get follow-up by ID
- `POST /api/follow-ups` - Create follow-up
- `PUT /api/follow-ups/:id` - Mark follow-up complete

### Settings
- `GET /api/settings` - Get all settings
- `PUT /api/settings` - Update settings
- `GET /api/settings/sync-schedule` - Get sync schedule
- `PUT /api/settings/sync-schedule` - Update sync schedule
- `GET /api/settings/anonymization-strategy` - Get anonymization strategy
- `PUT /api/settings/anonymization-strategy` - Update anonymization strategy
- `GET /api/settings/port` - Get port setting
- `PUT /api/settings/port` - Update port setting

### Onboarding
- `GET /api/onboarding/status` - Get onboarding status
- `POST /api/onboarding/complete` - Mark onboarding complete

## Architecture

```
engineerdna/
├── main.go                 # CLI entry point
├── internal/
│   ├── api/                # HTTP handlers
│   ├── anonymization/      # Anonymization service
│   ├── config/             # Configuration & encryption
│   ├── db/                 # Database layer
│   ├── models/             # Data models
│   └── plugin/             # Plugin system
├── plugins/
│   ├── plugin-sdk/         # Plugin SDK
│   ├── github/             # GitHub plugin (example)
│   ├── csv-import/         # CSV import plugin (example)
│   └── ai-insights/        # AI insights plugin (example)
├── migrations/             # Database migrations
└── frontend/               # React frontend (embedded)
```

## Development

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build
make build

# Run in dev mode
make dev
```

## License

Apache 2.0 - See [LICENSE.md](LICENSE.md)

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## Security & Threat Model

**V1 is designed for local, single-user use only.**

- Runs on localhost (127.0.0.1:3847) - not exposed to network
- User authenticates via OS login (like Jupyter, MySQL local, Docker Desktop)
- Plugin isolation via subprocess model with timeouts
- Encryption for API keys (AES-256-GCM, BYOK)
- Anonymization required before sending data to external APIs

**Not suitable for:** Multi-user deployments, team dashboards, network access (use V2 when available)

For security issues, open a GitHub issue or contact maintainers.
