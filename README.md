# EngineerDNA

Engineering metrics and insights platform with privacy-first anonymization and bring-your-own-key (BYOK) encryption.

## Deployment Model

**EngineerDNA is a local desktop application** that runs on `localhost:3847` - similar to Jupyter Notebook, Grafana, or Docker Desktop.

- Runs on your laptop/desktop only
- Single user (you)
- No network exposure
- No authentication needed (you authenticated when you logged in)

**For network/team deployments:** V2 will add authentication, authorization, and TLS.

## Features

### Core Platform

- **Identity Management**: Track engineers across multiple data sources with AI-powered identity resolution
- **Plugin Architecture**: Extensible system with five plugin types (Source, Metric Source, Attribute Source, Destination, Processor)
- **Privacy First**: Built-in anonymization, bidirectional mapping, required for external API calls, full audit trail
- **Security**: AES-256-GCM encryption for secrets, BYOK support, data stays local, no telemetry
- **Self-Hosted**: Single binary distribution, runs anywhere
- **SQLite Backend**: Zero external dependencies
- **Web UI**: Built-in dashboard for visualization and configuration
- **Alert System**: Real-time monitoring with configurable rules, multiple delivery channels, lifecycle management
- **Goal Tracking**: OKR-style goals with milestones, dependencies, automatic progress detection
- **Skill Development**: Evidence-based skill tracking with automatic detection from PRs/reviews
- **Cost & ROI Analysis**: Engineering cost configuration, feature value estimation, ROI calculations, investment breakdown
- **Manager Context**: Manager notes, context annotations, sentiment surveys, team context
- **Predictive Analytics**: Sprint/goal completion predictions, what-if scenarios, risk predictions, forecast accuracy
- **Action Tracking**: AI-generated recommendations, action tracking with assignments, outcome measurement
- **Dashboard Builder**: Customizable dashboards with 6 widget types, 4 default templates, pre-computed metric snapshots
- **Multi-Modal Data System**: Metrics (time-series), Attributes (temporal validity), Correlations (cross-data-type)
- **Event Normalization**: Multi-source event support (GitHub PR + GitLab MR → code_review)
- **Metric Calculation Engine**: Compute metrics from events using SQL aggregations
- **Widget Registry**: Plugin-provided UI components
- **New Plugin Types**: Metric Source, Attribute Source

## Quick Start

### Installation

**Quick Install** (Recommended):
```bash
curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh | sh
```

**Custom Directory**:
```bash
INSTALL_DIR=$HOME/.local/bin curl -sSL https://raw.githubusercontent.com/engineerdna/engineerdna/main/install.sh | sh
```

**From Source**:
```bash
git clone https://github.com/engineerdna/engineerdna.git
cd engineerdna
make build
sudo make install
```

### Initialize & Start

```bash
# Initialize
engineerdna init

# Start server
engineerdna serve
```

Server runs at `http://localhost:3847`

### Configuration

Set master encryption key (optional but recommended):
```bash
export ENGINEERDNA_MASTER_KEY=<base64-encoded-32-byte-key>
```

Generate key:
```bash
openssl rand -base64 32
```

## Docker Deployment

### Localhost-Only (Recommended - Secure)

```bash
# Start
docker-compose up -d

# Access
http://localhost:3847
```

**Security:** Same as native binary - secured by OS login + localhost-only binding.

### Network Exposure (Not Recommended)

**WARNING:** EngineerDNA has NO network security (no authentication, authorization, TLS, or access control).

```bash
# Network exposure (use with extreme caution)
docker-compose -f docker-compose.network.yml up -d
```

**Only use if:**
- You are on a trusted, private network
- You trust everyone on your local network
- You have firewall rules restricting access
- You understand the security implications

**For production use, wait for V2** which includes JWT authentication, role-based authorization, TLS/HTTPS, and API key management.

### Security Hardening

**SSH Tunnel** (Most Secure):
```bash
# Keep localhost-only binding in docker-compose.yml
ports:
  - "127.0.0.1:3847:3847"

# Access remotely via SSH tunnel
ssh -L 3847:localhost:3847 user@docker-host
# Then open http://localhost:3847 on local machine
```

**Other Options**: Firewall rules, reverse proxy with authentication, VPN-only access. See `docker-compose.yml` for details.

## Development

### Quick Restart Script

```bash
# Basic restart
./scripts/restart-dev-server.sh

# Rebuild and restart
./scripts/restart-dev-server.sh --rebuild

# Include frontend dev server
./scripts/restart-dev-server.sh --frontend

# Rebuild everything and start both servers
./scripts/restart-dev-server.sh --rebuild --frontend
```

Logs: `/tmp/engineerdna-backend.log`, `/tmp/engineerdna-frontend.log`

### Manual Workflow

```bash
# Backend
make build
./bin/engineerdna serve

# Frontend (with hot reload)
cd frontend
npm run dev  # Starts Vite dev server on port 5173

# Production build
cd frontend && npm run build
make build  # Embeds frontend/dist/ in binary
```

### Pre-commit Checks

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

See [plugins/plugin-sdk/](plugins/plugin-sdk/) for SDK and examples.

### Example Plugin

```go
package main

import "github.com/engineerdna/engineerdna/plugins/plugin-sdk"

type MyPlugin struct {
    config map[string]string
}

func (p *MyPlugin) Info() sdk.PluginInfo {
    return sdk.PluginInfo{
        Name: "my-plugin", Version: "1.0.0", Type: "source",
    }
}

func (p *MyPlugin) Configure(config map[string]string) error {
    p.config = config
    return nil
}

func (p *MyPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
    return sdk.SyncResult{Events: []sdk.Event{}}, nil
}

func main() {
    sdk.Serve(&MyPlugin{})
}
```

### Build and Install

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

### Core

- **Health & System**: `/api/health`, `/api/version`, `/api/system/info`
- **Events**: `/api/events`, `/api/events/:id`
- **Engineers**: `/api/engineers`, `/api/engineers/:id`, `/api/engineers/import`
- **Identity Resolution**: `/api/identity/unresolved`, `/api/identity/suggestions`, `/api/identity/resolve`
- **Plugins**: `/api/plugins`, `/api/plugins/:name/configure`, `/api/plugins/:name/sync`
- **Metrics**: `/api/metrics/throughput`, `/api/metrics/cycle-time`, `/api/metrics/today`
- **Insights**: `/api/insights`

### Dashboards

- **Management**: `/api/dashboards`, `/api/dashboards/:id`, `/api/dashboards/:id/clone`
- **Widget Data**: `/api/metrics/aggregate`, `/api/metrics/timeseries`, `/api/metrics/compare`, `/api/engineers/performance`, `/api/metrics/health-status`

### Teams & Performance

- **Teams**: `/api/teams`, `/api/teams/:id`, `/api/teams/hierarchy`, `/api/teams/org/scorecard`
- **Roles & Scoring**: `/api/roles`, `/api/scoring/weights`, `/api/scoring/calculate`
- **Performance**: `/api/performance/individual/:id`, `/api/performance/team/:id`
- **Promotion**: `/api/promotion/signals`, `/api/promotion/detect`
- **Briefings**: `/api/briefing/weekly/:id`, `/api/briefing/org`

### Planning & Forecasting

- **Planning**: `/api/planning/sprints`, `/api/planning/velocity/:id`, `/api/planning/estimate`, `/api/planning/capacity`
- **Alerts**: `/api/alerts/rules`, `/api/alerts/instances`, `/api/alerts/channels`
- **Goals**: `/api/goals`, `/api/goals/:id/milestones`, `/api/goals/:id/progress`
- **Skills**: `/api/skills`, `/api/engineers/:id/skills`, `/api/skills/detect`
- **Cost & ROI**: `/api/cost/configuration`, `/api/features`, `/api/features/:id/roi`
- **Context**: `/api/notes`, `/api/context/annotations`, `/api/surveys`
- **Forecasts**: `/api/forecasts`, `/api/forecasts/sprint/:id`, `/api/forecasts/accuracy`
- **Risks**: `/api/risks`, `/api/risks/predict/attrition`
- **Actions**: `/api/recommendations`, `/api/actions`, `/api/follow-ups`

### Settings

- **Configuration**: `/api/settings`, `/api/settings/sync-schedule`, `/api/settings/anonymization-strategy`
- **Onboarding**: `/api/onboarding/status`, `/api/onboarding/complete`

For complete API documentation, see the [API Reference](docs/api-reference.md).

## Dashboard Builder

Create personalized engineering dashboards with:
- **4 pre-built templates** (IC, Team Lead, Director, Planning)
- **6 widget types** (Number Card, Timeseries, Bar Chart, Table, Status Indicator, Activity Feed)
- **Drag-and-drop layout** with responsive grid
- **Real-time data** auto-refreshing every 30 seconds

### Quick Start

1. Navigate to Dashboards page
2. Select a template or create custom dashboard
3. Click "Edit Layout" to enable drag-and-drop
4. Add widgets and configure data sources
5. Drag to arrange, click "Done Editing" to save

For detailed dashboard documentation, see [Dashboard Guide](docs/dashboard-guide.md).

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
│   ├── github/             # GitHub plugin
│   ├── csv-import/         # CSV import plugin
│   └── ai-insights/        # AI insights plugin
├── migrations/             # Database migrations
└── frontend/               # React frontend (embedded)
```

## Development Commands

```bash
make fmt          # Format code
make lint         # Run linter
make test         # Run tests
make build        # Build binary
make dev          # Run in dev mode
```

## License

Apache 2.0 - See [LICENSE.md](LICENSE.md)

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## Security & Threat Model

**Designed for local, single-user use only.**

- Runs on localhost (127.0.0.1:3847) - not exposed to network
- User authenticates via OS login (like Jupyter, MySQL local, Docker Desktop)
- Plugin isolation via subprocess model with timeouts
- Encryption for API keys (AES-256-GCM, BYOK)
- Anonymization required before sending data to external APIs

**Not suitable for:** Multi-user deployments, team dashboards, network access (use V2 when available)

For security issues, open a GitHub issue or contact maintainers.
