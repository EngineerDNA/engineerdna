# CLAUDE.md - EngineerDNA Navigation Guide

## PRODUCT OVERVIEW
EngineerDNA is an engineering metrics and insights platform with a powerful plugin architecture. **Local desktop application** (like Jupyter or Grafana) that runs on `localhost:3847`. Single cross-platform binary (Go + embedded React) with SQLite database.

**Multi-Modal Data Model**: Supports three data types - Events (discrete actions), Metrics (time-series measurements), Attributes (entity facts).

### How It Works
1. **Install & Run**: Single binary runs locally on your machine
2. **Connect Data**: Plugins sync events (GitHub, Jira), metrics (AWS cost), attributes (HRIS)
3. **Analyze & Export**: Process metrics with AI, compute correlations, export to dashboards/reports
4. **Privacy First**: Anonymization for external processing, encryption for secrets

## NAVIGATION
Each major directory contains its own CLAUDE.md with context-specific instructions.

**Key Documentation:**
- `.claude/CLAUDE.md` - Agent orchestration, skills, and hooks
- `internal/db/CLAUDE.md` - Database patterns and migrations
- `internal/plugin/CLAUDE.md` - Plugin system implementation
- `plugins/CLAUDE.md` - Plugin development guide
- `frontend/CLAUDE.md` - Frontend development

## DOMAIN MODEL

### Core Entities

| Term | Definition | Key Rule |
|------|------------|----------|
| Event | Discrete timestamped action (PR, commit, deploy) | Immutable, UTC timestamp, source + normalized type |
| Metric | Time-series measurement (cost, velocity, hours) | Granular, multi-dimensional, timestamped |
| Attribute | Fact about entity (team size, budget, location) | Temporal validity tracking (valid from/until) |
| Engineer | Individual contributor tracked across data sources | Identity resolution via AI matching |
| Team | Group of engineers with hierarchy support | Parent-child relationships allowed |
| Identity | Distinct email/username from data sources | Must be resolved to an Engineer |
| Plugin | External data connector (5 types) | JSON-RPC over stdin/stdout |
| Source Plugin | Brings event data IN (GitHub, Jira, GitLab) | Implements `source.sync` |
| Metric Source Plugin | Brings metric data IN (AWS, budget, capacity) | Implements `metric_source.sync` |
| Attribute Source Plugin | Brings attributes IN (HRIS, org chart) | Implements `attribute_source.sync` |
| Destination Plugin | Sends data OUT (Sheets, Slack, PDF) | Implements `destination.export` |
| Processor Plugin | Transforms data (AI analysis, correlations) | Requires anonymization for external APIs |
| Anonymization | PII protection for external transmission | Bidirectional mapping, 3 strategies |
| Audit Log | Export and processing tracking | All anonymization logged |
| Alert Rule | Threshold-based monitoring condition | Enabled/disabled, severity levels |
| Alert Instance | Fired alert with lifecycle | Acknowledge, snooze, dismiss states |
| Alert Channel | Delivery mechanism (Slack, email, browser) | Quiet hours support |
| Goal | OKR-style objective | Individual, team, or org-wide |
| Goal Milestone | Measurable checkpoint within a goal | Target vs current value tracking |
| Skill | Technical, leadership, or communication ability | Defined in taxonomy |
| Engineer Skill | Skill level for an engineer | 0-100 score, trajectory tracking |
| Cost Configuration | Engineering cost by engineer/team/role | Fully-loaded cost (salary + benefits) |
| Feature Value | Business value of a feature | ARR impact, confidence level |
| ROI Calculation | Feature return on investment | Investment vs return, payback period |
| Manager Note | Qualitative context from managers | Visibility controls, mood tracking |
| Sentiment Survey | Pulse check for morale | Anonymous responses |
| Forecast | Predicted future outcome | Confidence interval, expiration |
| Recommendation | AI or manual action suggestion | Priority, status, expiration |
| Action | Manager action taken | Evidence-based, outcome tracking |
| Dashboard | Customizable visualization with widget layout | Templates are system dashboards (cannot be deleted) |
| Widget | Visual component on dashboard | 6 types: number, timeseries, bar, table, status, feed |
| Metric Snapshot | Pre-computed metric for performance | Computed hourly, stored by entity and period |
| Metric Value | Time-series measurement (cost, velocity, hours) | Timestamped, granular (hourly/daily/weekly/monthly), multi-dimensional |
| Entity Attribute | Fact about entity with temporal validity | Team size, budget, location - valid from/until tracking |
| Correlation | Cross-data-type relationship | Links events, metrics, attributes (e.g., cost per feature) |
| Plugin Manifest | Plugin capabilities registry | Stored in DB, tracks metrics/events/widgets/correlations provided |
| Event Type Registry | Normalized event types for multi-source | GitHub PR + GitLab MR → code_review |

## Metrics Storage Architecture

### Layer 1: Universal Plugin Data
**Purpose:** Extensible data FROM plugins

- `metric_values` - Time-series measurements from any source (AWS costs, velocity, latency)
- `entity_attributes` - Facts about entities from any source (team size, budgets, locations)
- `events` - Discrete actions from any source (PRs, deployments, incidents)
- `correlations` - Cross-data-type calculations defined by plugins

**Use when:** Plugin-provided data, extensible schema, multi-source support needed

### Layer 2: Application Features
**Purpose:** Well-defined features OF the app

- `alert_instances` - Stateful alert lifecycle
- `goals` - OKR tracking with dependencies
- `forecasts` - Predictions with confidence intervals
- `teams` - Organizational hierarchy

**Use when:** Complex lifecycles, foreign keys, CHECK constraints, typed queries needed

### Data Storage Examples

**Performance Scores**: Stored in `metric_values` with metric names:
- `engineer_total_score`, `engineer_throughput_score`, `engineer_quality_score`, etc.
- `team_total_score`, `team_throughput_score`, `team_quality_score`, etc.
- Dimensions contain `engineer_id` or `team_id` for filtering

**Cost Configuration**: Stored in `entity_attributes` with attribute name `monthly_cost`:
- Entity types: `engineer`, `team`, `role`
- Value type: `currency` (stored as string representation of float)
- Temporal validity: `valid_from` and `valid_until` track cost changes over time

## CRITICAL PROJECT-WIDE RULES

1. **Localhost-only** - Bind to 127.0.0.1:3847, never 0.0.0.0
2. **Single binary distribution** - Frontend embedded via `//go:embed frontend/dist`
3. **SQLite only** - modernc.org/sqlite (pure Go), no CGO
4. **Plugin isolation** - Subprocess with timeouts, no shared memory
5. **Anonymization required** - For processor plugins sending data to external APIs
6. **Encryption at rest** - AES-256-GCM for all secrets (API keys, tokens)
7. **Audit everything** - Log all exports and processor calls with anonymization status
8. **UTC timestamps everywhere** - Always `time.Now().UTC()`, store as RFC3339
9. **No direct database modifications** - Always use migrations
10. **Repository pattern** - No raw SQL in handlers, use repository layer
11. **Plugin discovery** - Check both `~/.engineerdna/plugins/` and `./plugins/`
12. **JSON-RPC protocol** - All plugin communication via stdin/stdout
13. **Graceful timeouts** - Context with timeout for all plugin executions
14. **Input validation** - Validate all user inputs, prevent SQL injection
15. **Error boundaries** - Frontend error handling with graceful degradation
16. **All files under 500 lines** - Split large files into focused modules
17. **Professional naming only** - No "refactored", "v2", "new", "old" in names
18. **No phases** - Work is continuous and incremental
19. **Task-focused engineering** - One task = one agent invocation
20. **Agent pipeline for features** - engineer → quality → integration-checker → data-engineer → ui-tester → docs → advisor
21. **Pipeline flows backward on errors** - Fix issues, don't just report them
22. **No partial implementations** - Every line must work, no TODOs/stubs
23. **Test what you build** - Verify functionality, don't assume
24. **Fix what breaks** - Don't move on until current work is solid
25. **Report accurately** - Absolute honesty about results
26. **ZERO emojis ANYWHERE** - Never in code, docs, commits, PRs, or responses
27. **NO ONE-OFF DATABASE SCRIPTS** - Always create migrations in `migrations/` directory
28. **UTC TIMESTAMPS EVERYWHERE** - Use `time.Now().UTC()` and RFC3339 format
29. **NO ASPIRATIONAL CODE** - Delete unused code immediately, don't keep "for later"
30. **Verify plugin.json** - Validate manifest before plugin execution
31. **Master key security** - OS keychain preferred over env var
32. **Event immutability** - Never UPDATE events, create new ones
33. **Bidirectional anonymization** - Store mapping for deanonymization
34. **Plugin SDK patterns** - Follow SDK conventions for consistency
35. **Cross-platform builds** - Test on Linux, macOS, Windows

## QUICK START CHECKLIST

### Before ANY Work:
```bash
pwd                # Verify current directory
git status         # Check branch and uncommitted changes
go vet ./...       # Run Go static analysis
```

### After ANY Work:

**Backend (Go):**
```bash
go vet ./...       # Static analysis
staticcheck ./...  # Additional linting
go test ./...      # Run all tests
make build         # Ensure everything builds
```

**Frontend:**
```bash
cd frontend
npm run lint       # ESLint
npm run typecheck  # TypeScript
npm run build      # Production build
npm test           # Unit tests
```

### Before Database Work:
**STOP AND CHECK THESE RULES:**
1. **Schema changes**: Modify Go structs → Create migration file → Test migration
2. **Data fixes**: Create migration in `migrations/` → Test migration → Apply
3. **NEVER**: Create scripts/fix-*.go or run SQL outside migrations
4. **Ask yourself**: "Will this apply consistently across all environments?" If no, you're doing it wrong.

## AI AGENT BEHAVIOR PRINCIPLES

### Code Quality Standards
1. **NO SHORTCUTS** - Complete, production-ready implementations only
   - No stubs, TODOs, or "will implement later" comments
   - No placeholder functions or mock data
   - Every feature must be fully functional end-to-end

2. **CRITICAL THINKING REQUIRED** - Challenge incorrect assumptions
   - Don't automatically agree - evaluate requests against codebase reality
   - If user's approach conflicts with architecture/best practices, explain why with evidence
   - Present data from codebase searches/tests to support your position

3. **ABSOLUTE HONESTY** - Report actual results, not desired outcomes
   - If tests fail, show the failure and fix it - don't pretend it passed
   - Never hardcode expected results to make tests pass
   - If a build breaks, acknowledge it and resolve it

## AI AGENT ORCHESTRATION

### YOU ARE THE ORCHESTRATOR

### Available Agents (.claude/agents/)
Agents are stateless LLM subagents invoked via Task tool. See `.claude/agents/CLAUDE.md` for details.

### Efficient Routing by Task Type

| Task Type | Agent Pipeline | Skip Agents If |
|-----------|---------------|----------------|
| **Bug/Error Fix** | engineer → quality → Done | Clear fix |
| **Plugin Development** | engineer → quality → integration-checker → Done | Plugin-focused |
| **Database Schema** | engineer → quality → data-engineer → Done | Schema changes |
| **New Feature (unclear)** | engineer → quality → integration-checker → data-engineer → ui-tester → docs → advisor | Need full validation |
| **New Feature (clear)** | engineer → quality → integration-checker → data-engineer → ui-tester | Requirements clear |
| **Refactoring** | engineer → quality → advisor | Code cleanup |
| **Performance Issue** | engineer → quality → Done | Optimization |
| **Plugin Not Working** | integration-checker → engineer → integration-checker | Connection issue |
| **Review Request** | advisor → engineer (if issues) | Code review |

**Documentation Updates (Async)**: Launch `docs` agent in background when domain model changes, new rules, or workflow updates.

### Context Accumulation & Task-Focused Engineering

**One Task = One Engineer invocation. NEVER pass todo lists.**

Agents are stateless - pass complete context each time:
- USER GOAL: One-liner
- TECHNICAL SPEC: Data structures, interfaces, security
- YOUR TASK: ONE specific implementation

See `.claude/agents/CLAUDE.md` for complete details.

### Iteration & Backward Flow Rules

**When to go BACKWARD in the pipeline:**
- **quality fails tests** → Back to engineer with failing tests
- **integration-checker finds plugin issues** → Back to engineer to fix plugin communication
- **data-engineer finds data integrity issues** → Back to engineer with specific issues
- **ui-tester finds broken UI** → Back to engineer to fix frontend
- **advisor blocks with issues** → Back to engineer with specific fixes

**When to RETRY same agent:**
- Simple syntax/import errors
- Timeout/transient failures
- Missing file that can be created

**When to STOP and ask user:**
- Multiple back-and-forth attempts fail (>3 cycles)
- Fundamental architecture conflict discovered
- Requirements contradict existing system

## COMMON MISTAKES (TOP 5)

1. **Not Running Checks** - Always run pre/post work checklists
2. **Wrong Directory** - Always `pwd` first
3. **Direct Database Modifications** - Always use migrations, never run SQL directly
4. **Binding to 0.0.0.0** - Always bind to 127.0.0.1 for localhost-only
5. **No Plugin Timeouts** - Always use `context.WithTimeout` for plugin execution

## STATUS VALUES

### Plugin Status
- **discovering**: Finding plugins in plugin directories
- **loading**: Loading plugin.json manifests
- **configuring**: User providing configuration
- **configured**: Ready to use
- **running**: Actively executing (sync, export, analyze)
- **failed**: Execution error
- **disabled**: User disabled

### Event Status
- Events are immutable - no status field
- Use separate fields: `anonymized` (bool), `exported` (bool)

### Alert Status
- **fired**: Alert triggered, not yet acknowledged
- **acknowledged**: Manager has seen the alert
- **snoozed**: Temporarily suppressed until date
- **dismissed**: Alert dismissed without action
- **resolved**: Underlying condition fixed

### Goal Status
- **active**: In progress
- **completed**: Successfully achieved
- **at_risk**: Behind schedule or blocked
- **off_track**: Significant risk of missing target
- **archived**: Historical record, no longer tracked

## SECURITY MODEL

### Localhost-Only Security
- **Primary Defense**: Network isolation (127.0.0.1 binding)
- **No Authentication**: Authenticated via OS login (like Jupyter, Grafana)
- **Plugin Isolation**: Subprocess model with timeouts
- **Encryption**: AES-256-GCM for secrets at rest
- **Anonymization**: Required for external API transmission
- **Audit Trail**: All exports and processing logged

### Critical Security Patterns

```go
// CORRECT: Localhost-only binding
http.ListenAndServe("127.0.0.1:3847", handler)

// CORRECT: Plugin with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, pluginPath, args...)

// CORRECT: Encrypt secrets
if field.Secret {
    encrypted, err := encryptionSvc.Encrypt(value)
    config[field.Name] = encrypted
}
```

## QUICK REFERENCE

### Development
```bash
go run .                            # Run server
go test ./...                       # Run tests
go vet ./...                        # Static analysis
staticcheck ./...                   # Advanced linting

# Database Operations
sqlite3 ~/.engineerdna/engineerdna.db ".schema"  # View schema
# NEVER: sqlite3 ... "UPDATE/DELETE/ALTER"
# ALWAYS: Create migration in migrations/

# Frontend Development
cd frontend
npm run dev                         # Dev server
npm run build                       # Production build
npm run typecheck                   # TypeScript check
```

### Build & Deploy
```bash
make build                          # Build for current platform
make cross-compile                  # All platforms

# Version management
# 1. Update CHANGELOG.md with new version
# 2. git commit -m "Release vX.Y.Z"
# 3. git push origin main
```

### Common Operations
```bash
# Plugin Development
cd plugins/my-plugin
go build -o my-plugin              # Build plugin binary
cat plugin.json                     # View manifest

# Create Migration
cat > migrations/005_description.sql << 'EOF'
-- Migration SQL here
ALTER TABLE events ADD COLUMN new_field TEXT;
EOF

# Test Migration
go run . migrate up
```

### Environment Variables
```bash
export ENGINEERDNA_MASTER_KEY="your-32-byte-key-here"  # Optional - OS keychain preferred
export ENGINEERDNA_DATA_DIR="$HOME/.engineerdna"       # Optional - Custom data directory
export ENGINEERDNA_PORT="3847"                         # Optional - Custom port
```

## TROUBLESHOOTING

### Common Error Patterns
| Error | Fix |
|-------|-----|
| Plugin not found | Check plugin directories and plugin.json |
| JSON-RPC error | Verify plugin is executable and implements required methods |
| Plugin timeout | Increase context timeout or optimize plugin code |
| Database locked | Close other connections, check for zombie processes |
| Encryption failed | Verify ENGINEERDNA_MASTER_KEY or OS keychain setup |
| Bind address in use | Port 3847 already taken, kill other process or change port |
| Frontend blank page | Rebuild frontend: `cd frontend && npm run build && cd .. && make build` |

### Debug Commands
```bash
# Plugin Issues
ls -la ~/.engineerdna/plugins/
echo '{"jsonrpc":"2.0","method":"plugin.info","id":1}' | ./plugins/github/github

# Database Issues
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM schema_migrations"
lsof ~/.engineerdna/engineerdna.db

# Build Issues
ls -la frontend/dist/
CGO_ENABLED=0 go build
```

## ARCHITECTURE DETAILS

### Plugin System
**Implementation**: See `internal/plugin/CLAUDE.md`
- Subprocess isolation
- JSON-RPC over stdin/stdout
- Timeout enforcement
- Universal + type-specific methods

### Database Layer
**Implementation**: See `internal/db/CLAUDE.md`
- SQLite with modernc.org driver (pure Go)
- Repository pattern
- Migration system
- UTC timestamps (RFC3339)

### Anonymization
**Implementation**: See `internal/anonymization/`
- Three strategies: sequential (User1, User2), UUID, hash
- Bidirectional mapping storage
- Required for processor plugins
- Audit logging

### Encryption
**Implementation**: See `internal/config/`
- AES-256-GCM for secrets
- Master key from OS keychain or env var
- Automatic encrypt/decrypt for `secret: true` fields

### Frontend
**Implementation**: See `frontend/CLAUDE.md`
- Vite + React 18 + TypeScript
- TanStack Query for data fetching
- Recharts for visualizations
- Tailwind CSS for styling

---

For detailed implementation information, navigate to the appropriate subdirectory CLAUDE.md file listed above.
