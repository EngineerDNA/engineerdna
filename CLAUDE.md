# CLAUDE.md - EngineerDNA Navigation Guide

## PRODUCT OVERVIEW
EngineerDNA is an engineering metrics and insights platform with a powerful plugin architecture. **Local desktop application** (like Jupyter or Grafana) that runs on `localhost:3847`. Single cross-platform binary (Go + embedded React) with SQLite database.

### How It Works
1. **Install & Run**: Single binary runs locally on your machine
2. **Connect Data**: Plugins sync from GitHub, Jira, CSV, etc.
3. **Analyze & Export**: Process metrics with AI, export to dashboards/reports
4. **Privacy First**: Anonymization for external processing, encryption for secrets

## NAVIGATION
Each major directory contains its own CLAUDE.md with context-specific instructions.

**Key Documentation:**
- `.claude/CLAUDE.md` - Agent orchestration, skills, and hooks
- `internal/db/CLAUDE.md` - Database patterns and migrations
- `internal/plugin/CLAUDE.md` - Plugin system implementation
- `plugins/CLAUDE.md` - Plugin development guide

## DOMAIN MODEL

### Core Entities

| Term | Definition | Key Rule |
|------|------------|----------|
| Event | Unit of engineering data (PR, issue, metric) | Immutable, timestamped in UTC |
| Engineer | Individual contributor tracked across data sources | Identity resolution via AI matching |
| Team | Group of engineers with hierarchy support | Parent-child relationships allowed |
| Identity | Distinct email/username from data sources | Must be resolved to an Engineer |
| Plugin | External data connector (source/dest/processor) | JSON-RPC over stdin/stdout |
| Source Plugin | Brings data IN (GitHub, CSV, Jira) | Implements `source.sync` |
| Destination Plugin | Sends data OUT (Sheets, Slack, PDF) | Implements `destination.export` |
| Processor Plugin | Transforms data (AI analysis, metrics) | Requires anonymization for external APIs |
| Anonymization | PII protection for external transmission | Bidirectional mapping, 3 strategies |
| Audit Log | Export and processing tracking | All anonymization logged |

### PDR-7 Entities (v1.1.0)

| Term | Definition | Key Rule |
|------|------------|----------|
| Alert Rule | Threshold-based monitoring condition | Enabled/disabled, severity levels |
| Alert Instance | Fired alert with lifecycle | Acknowledge, snooze, dismiss states |
| Alert Channel | Delivery mechanism (Slack, email, browser) | Quiet hours support |
| Goal | OKR-style objective | Individual, team, or org-wide |
| Goal Milestone | Measurable checkpoint within a goal | Target vs current value tracking |
| Goal Progress Log | History of goal progress updates | Manual or automatic detection |
| Goal Dependency | Blocker or relationship between goals | Active or resolved status |
| Skill | Technical, leadership, or communication ability | Defined in taxonomy |
| Engineer Skill | Skill level for an engineer | 0-100 score, trajectory tracking |
| Skill Evidence | Proof of skill usage from events | Strength 0.0-1.0 |
| Skill Progression | Historical skill level changes | Time-series tracking |
| Cost Configuration | Engineering cost by engineer/team/role | Fully-loaded cost (salary + benefits) |
| Feature Value | Business value of a feature | ARR impact, confidence level |
| Feature Work Item | Engineering work linked to feature | Story points, hours, PRs |
| ROI Calculation | Feature return on investment | Investment vs return, payback period |
| Engineering Investment | Time allocation by category | Customer features, tech debt, etc. |
| Manager Note | Qualitative context from managers | Visibility controls, mood tracking |
| Context Annotation | Explanation for metrics/events | Team or org-wide visibility |
| Sentiment Survey | Pulse check for morale | Anonymous responses |
| Sentiment Analysis | Computed sentiment score | -1.0 to 1.0, confidence level |
| Forecast | Predicted future outcome | Confidence interval, expiration |
| Scenario | What-if hypothetical situation | Baseline vs modified parameters |
| Risk Prediction | Predicted risk with probability | Severity and mitigation suggestions |
| Forecast Accuracy | Historical forecast performance | Tracks model effectiveness |
| Recommendation | AI or manual action suggestion | Priority, status, expiration |
| Action | Manager action taken | Evidence-based, outcome tracking |
| Action Outcome | Measured effectiveness of action | Before/after metrics |
| Follow-Up | Scheduled reminder for action | Completion tracking |

### PDR-8 Entities (v1.2.0)

| Term | Definition | Key Rule |
|------|------------|----------|
| Dashboard | Customizable visualization with widget layout | Templates are system dashboards (cannot be deleted) |
| Widget | Visual component on dashboard | 6 types: number, timeseries, bar, table, status, feed |
| Metric Snapshot | Pre-computed metric for performance | Computed hourly, stored by entity and period |

## CRITICAL PROJECT-WIDE RULES

1. **Localhost-only (V1)** - Bind to 127.0.0.1:3847, never 0.0.0.0
2. **Single binary distribution** - Frontend embedded via `//go:embed frontend/dist`
3. **SQLite only** - modernc.org/sqlite (pure Go), no CGO
4. **Plugin isolation** - Subprocess with timeouts, no shared memory
5. **Anonymization required** - For processor plugins sending data to external APIs
6. **Encryption at rest** - AES-256-GCM for all secrets (API keys, tokens)
7. **Audit everything** - Log all exports and processor calls with anonymization status
8. **UTC timestamps everywhere** - Always `time.Now().UTC()`, store as RFC3339
9. **No direct database modifications** - Always use migrations (Rule 33)
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
20. **Agent pipeline for features** - engineer → quality → integration-checker → security → ui-tester → docs → advisor
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
# When frontend added: cd frontend && npm run typecheck
```

### After ANY Work:

**Backend (Go):**
```bash
go vet ./...       # Static analysis
staticcheck ./...  # Additional linting
go test ./...      # Run all tests
make build         # Ensure everything builds
```

**Frontend (when added):**
```bash
cd frontend
npm run lint       # ESLint
npm run typecheck  # TypeScript
npm run build      # Production build
npm test           # Unit tests
```

**Always:**
```bash
git diff           # Review all changes
```

### Before Database Work:
**STOP AND CHECK THESE RULES:**
1. **Schema changes**: Modify Go structs → Create migration file → Test migration
2. **Data fixes**: Create migration in `migrations/` → Test migration → Apply
3. **NEVER**: Create scripts/fix-*.go or run SQL outside migrations
4. **Ask yourself**: "Will this apply consistently across all environments?" If no, you're doing it wrong.

**Common Anti-Patterns to Avoid:**
- Creating one-off Go scripts that execute SQL directly
- Running `sqlite3 engineerdna.db` with UPDATE/DELETE/ALTER
- Adding temporary "fix" functions that bypass migrations
- Manual schema modifications in database

**The Right Way:**
```bash
# Create migration file
cat > migrations/004_add_anonymization_flags.sql << 'EOF'
-- Add anonymization tracking to events
ALTER TABLE events ADD COLUMN anonymized BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_events_anonymized ON events(anonymized);
EOF

# Test migration
go run . migrate up

# Verify schema
sqlite3 ~/.engineerdna/engineerdna.db ".schema events"
```

## AI AGENT BEHAVIOR PRINCIPLES

### Code Quality Standards
1. **NO SHORTCUTS** - Complete, production-ready implementations only
   - No stubs, TODOs, or "will implement later" comments
   - No placeholder functions or mock data
   - Every feature must be fully functional end-to-end
   - If time/complexity is a concern, propose a properly scoped solution instead

2. **CRITICAL THINKING REQUIRED** - Challenge incorrect assumptions
   - Don't automatically agree - evaluate requests against codebase reality
   - If user's approach conflicts with architecture/best practices, explain why with evidence
   - Present data from codebase searches/tests to support your position
   - Provide full context and recommend the right path
   - Example: "The user wants X, but based on [specific code/pattern], Y would be more appropriate because..."

3. **ABSOLUTE HONESTY** - Report actual results, not desired outcomes
   - If tests fail, show the failure and fix it - don't pretend it passed
   - Never hardcode expected results to make tests pass
   - If a build breaks, acknowledge it and resolve it
   - If you can't find something, say so - don't make up file paths or functions
   - Reality example: "The test failed with error X. Let me fix it..." NOT "The test passed successfully"

### Implementation Integrity
- **Every line of code must work** - No partial implementations
- **Test what you build** - Verify functionality, don't assume
- **Fix what breaks** - Don't move on until current work is solid
- **Report accurately** - Your credibility depends on truthfulness

## AI AGENT ORCHESTRATION

### YOU ARE THE ORCHESTRATOR

### Available Agents (.claude/agents/)
Agents are stateless LLM subagents invoked via Task tool. See `.claude/agents/CLAUDE.md` for details.

### Efficient Routing by Task Type

| Task Type | Agent Pipeline | Skip Agents If |
|-----------|---------------|----------------|
| **Bug/Error Fix** | engineer → quality → Done | Clear fix |
| **Plugin Development** | engineer → quality → integration-checker → Done | Plugin-focused |
| **Database Schema** | engineer → quality → Done | Schema only |
| **New Feature (unclear)** | engineer → quality → integration-checker → security → ui-tester → docs → advisor | Need full validation |
| **New Feature (clear)** | engineer → quality → integration-checker → ui-tester | Requirements clear |
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

**Agent Outputs**: Engineer → Quality → Integration-Checker → Security → UI-Tester → Docs → Advisor

See `.claude/agents/CLAUDE.md` for complete details.

### Iteration & Backward Flow Rules

**When to go BACKWARD in the pipeline:**
- **quality fails tests** → Back to engineer with failing tests
- **integration-checker finds plugin issues** → Back to engineer to fix plugin communication
- **security finds vulnerabilities** → Back to engineer with specific issues
- **ui-tester finds broken UI** → Back to engineer to fix frontend
- **advisor blocks with issues** → Back to engineer with specific fixes
- **engineer realizes requirements unclear** → Ask user for clarification

**When to RETRY same agent:**
- Simple syntax/import errors
- Timeout/transient failures
- Missing file that can be created

**When to STOP and ask user:**
- Multiple back-and-forth attempts fail (>3 cycles)
- Fundamental architecture conflict discovered
- Requirements contradict existing system

### Common Fixes to Apply Automatically
- **Import error**: Check package paths and mod file
- **Test failure**: Read error, fix issue, verify fix
- **Type error**: Run `go vet ./...` for details
- **Build error**: Check for missing dependencies

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

### Alert Status (v1.1.0)
- **fired**: Alert triggered, not yet acknowledged
- **acknowledged**: Manager has seen the alert
- **snoozed**: Temporarily suppressed until date
- **dismissed**: Alert dismissed without action
- **resolved**: Underlying condition fixed

### Goal Status (v1.1.0)
- **active**: In progress
- **completed**: Successfully achieved
- **at_risk**: Behind schedule or blocked
- **off_track**: Significant risk of missing target
- **archived**: Historical record, no longer tracked

### Skill Trajectory (v1.1.0)
- **improving**: Skill level increasing over time
- **stable**: Skill level consistent
- **declining**: Skill level decreasing (e.g., not used recently)

### Recommendation Status (v1.1.0)
- **pending**: Not yet acted upon
- **in_progress**: Action being taken
- **completed**: Action finished
- **dismissed**: Manager chose not to act
- **snoozed**: Deferred to later date

### Risk Severity (v1.1.0)
- **low**: Minor impact if occurs
- **medium**: Moderate impact
- **high**: Significant impact
- **critical**: Severe impact, immediate attention needed

## SECURITY MODEL

### Localhost-Only Security (V1)
- **Primary Defense**: Network isolation (127.0.0.1 binding)
- **No Authentication**: Authenticated via OS login (like Jupyter, Grafana)
- **Plugin Isolation**: Subprocess model with timeouts
- **Encryption**: AES-256-GCM for secrets at rest
- **Anonymization**: Required for external API transmission
- **Audit Trail**: All exports and processing logged

### Critical Security Rules
```go
// CORRECT: Localhost-only binding
http.ListenAndServe("127.0.0.1:3847", handler)

// WRONG: Network-accessible
http.ListenAndServe(":3847", handler) // Binds to 0.0.0.0!

// CORRECT: Plugin with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, pluginPath, args...)

// WRONG: No timeout
cmd := exec.Command(pluginPath, args...) // Can hang forever!

// CORRECT: Encrypt secrets
if field.Secret {
    encrypted, err := encryptionSvc.Encrypt(value)
    config[field.Name] = encrypted
}

// WRONG: Plaintext secrets
config[field.Name] = apiKey // Stored unencrypted!
```

### V2 Security (Future - Network Access)
When adding network access:
- Add JWT-based authentication
- Add API key management
- Add TLS/HTTPS
- Add rate limiting
- Restrict anonymization mappings endpoint
- Add RBAC for multi-user

## QUICK REFERENCE

### Development
```bash
# Backend Development
go run .                            # Run server
go test ./...                       # Run tests
go vet ./...                        # Static analysis
staticcheck ./...                   # Advanced linting

# Database Operations
sqlite3 ~/.engineerdna/engineerdna.db ".schema"  # View schema
# NEVER: sqlite3 ... "UPDATE/DELETE/ALTER"
# ALWAYS: Create migration in migrations/

# Frontend Development (when added)
cd frontend
npm run dev                         # Dev server
npm run build                       # Production build
npm run typecheck                   # TypeScript check
```

### Build & Deploy
```bash
# Local build
make build                          # Build for current platform

# Cross-platform builds
make cross-compile                  # All platforms
# Outputs: engineerdna-{platform}-{arch}

# Version management
# 1. Update CHANGELOG.md with new version
# 2. git commit -m "Release vX.Y.Z"
# 3. git push origin main
# GitHub Actions creates release automatically
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
go run . migrate down  # If reversible
```

### Environment Variables
```bash
# Optional - Encryption master key (OS keychain preferred)
export ENGINEERDNA_MASTER_KEY="your-32-byte-key-here"

# Optional - Custom data directory
export ENGINEERDNA_DATA_DIR="$HOME/.engineerdna"

# Optional - Custom port
export ENGINEERDNA_PORT="3847"
```

## TROUBLESHOOTING GUIDE

### Plugin Issues

**Problem**: Plugin not discovered
```bash
# Check plugin directories
ls -la ~/.engineerdna/plugins/
ls -la ./plugins/

# Verify plugin.json exists
find . -name "plugin.json"

# Check plugin.json is valid JSON
cat plugins/github/plugin.json | jq .
```

**Problem**: Plugin execution timeout
```go
// Increase timeout if legitimate long-running operation
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
// Or add streaming output to show progress
```

**Problem**: JSON-RPC communication error
```bash
# Test plugin directly
echo '{"jsonrpc":"2.0","method":"plugin.info","id":1}' | ./plugins/github/github

# Common issues:
# - Plugin not executable: chmod +x plugins/github/github
# - Missing dependencies: Check plugin's go.mod
# - Wrong entrypoint: Check plugin.json "entrypoint" field
```

### Database Issues

**Problem**: Migration failed
```bash
# Check migration syntax
sqlite3 ~/.engineerdna/engineerdna.db < migrations/004_test.sql

# Rollback if reversible
# Create down migration: migrations/004_test_down.sql

# If migration is stuck, check schema version
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM schema_migrations"
```

**Problem**: Database locked
```bash
# Check for other processes
lsof ~/.engineerdna/engineerdna.db

# Close all connections and retry
pkill engineerdna
```

### Build Issues

**Problem**: Frontend not embedded
```bash
# Verify frontend build exists
ls -la frontend/dist/

# Rebuild frontend
cd frontend && npm run build

# Rebuild Go binary
make build
```

**Problem**: Cross-compilation fails
```bash
# Ensure CGO is disabled (required for SQLite pure Go)
CGO_ENABLED=0 go build

# Check GOOS and GOARCH
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
```

### Common Error Patterns
| Error | Fix |
|-------|-----|
| Plugin not found | Check plugin directories and plugin.json |
| JSON-RPC error | Verify plugin is executable and implements required methods |
| Plugin timeout | Increase context timeout or optimize plugin code |
| Database locked | Close other connections, check for zombie processes |
| Encryption failed | Verify ENGINEERDNA_MASTER_KEY or OS keychain setup |
| Anonymization mapping not found | Check if event was actually anonymized |
| Bind address in use | Port 3847 already taken, kill other process or change port |
| Frontend blank page | Rebuild frontend: `cd frontend && npm run build && cd .. && make build` |

### PDR-7 Feature Issues (v1.1.0)

**Problem**: Alerts not firing
```bash
# Check alert rules are enabled
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM alert_rules WHERE enabled = 1"

# Check alert evaluation logs
grep "alert evaluation" /tmp/engineerdna-backend.log

# Verify threshold values are correct
# Common issue: threshold_operator wrong ('>' vs '<')
```

**Problem**: Goal progress not updating automatically
```bash
# Verify goal tracking_method is 'automatic' or 'hybrid'
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, title, tracking_method FROM goals"

# Check if events match goal criteria
# Goal metrics must be properly configured
```

**Problem**: Skills not detected
```bash
# Verify skill taxonomy is populated
sqlite3 ~/.engineerdna/engineerdna.db "SELECT COUNT(*) FROM skills"

# Check if skill detection is running
# Skills detected from PR complexity, review depth, etc.

# Verify events have required fields for detection
```

**Problem**: Cost calculations showing zero
```bash
# Verify cost configuration exists
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM cost_configuration"

# Check if engineers have role assignments
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name, role_id FROM engineers"

# Ensure feature work items are linked
```

**Problem**: Forecasts showing unrealistic values
```bash
# Check if enough historical data exists
# Forecasts require minimum 2-3 data points

# Verify model type is appropriate for data
# Linear regression needs linear trends
# Monte Carlo needs variability data

# Check for outliers in input data
```

**Problem**: Recommendations not appearing
```bash
# Verify recommendation source is active
# AI insights plugin must be configured

# Check recommendation expiration dates
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM recommendations WHERE expires_at > datetime('now')"

# Verify assigned_to matches manager_id
```

**Problem**: Sentiment surveys not visible
```bash
# Check survey is active and not expired
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM sentiment_surveys WHERE active = 1"

# Verify target_audience matches current user
# 'all', 'team:team_id', or 'engineer:engineer_id'

# Check if survey already responded to
```

### PDR-8 Feature Issues (v1.2.0)

**Problem**: Dashboards not loading or displaying
```bash
# Check if dashboards exist
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name, is_template FROM dashboards"

# Verify dashboard layout is valid JSON
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name, layout FROM dashboards" | jq .

# Check API endpoint
curl http://127.0.0.1:3847/api/dashboards
# Should return 200 with JSON list

# Common issue: Invalid JSON in layout field
# Fix: Re-create dashboard with valid layout
```

**Problem**: Widgets not rendering or showing no data
```bash
# Check widget configuration in dashboard layout
sqlite3 ~/.engineerdna/engineerdna.db "SELECT layout FROM dashboards WHERE id = 'dashboard_id'"

# Verify metric snapshots exist for widget queries
sqlite3 ~/.engineerdna/engineerdna.db "SELECT COUNT(*) FROM metric_snapshots"

# Check widget data API endpoint
curl "http://127.0.0.1:3847/api/metrics/aggregate?metric=team_score&period=week"
# Should return 200 with metric data

# Common issue: Widget type mismatch or missing required fields
# Required fields by widget type:
#   - number: metric, entity_type, entity_id, period
#   - timeseries: metric, entity_type, entity_id, period, count
#   - bar: metric, group_by, period
#   - table: period, sort_by, limit
#   - status: (no params required)
#   - feed: start, end
```

**Problem**: Metric snapshots not computed
```bash
# Check if snapshot computation is running
# Scheduled hourly by default

# Manually trigger snapshot computation
# (requires implementation of CLI command)

# Verify events exist to compute metrics from
sqlite3 ~/.engineerdna/engineerdna.db "SELECT COUNT(*) FROM events WHERE DATE(timestamp) = date('now')"

# Check if snapshots are being created
sqlite3 ~/.engineerdna/engineerdna.db "SELECT metric_name, COUNT(*) FROM metric_snapshots GROUP BY metric_name"

# Common issue: No events to compute from
# Fix: Sync data using source plugins (GitHub, Jira, etc.)
```

**Problem**: Dashboard mutations failing (create, update, delete)
```bash
# Check API logs for error details
grep "dashboard" /tmp/engineerdna-backend.log | tail -20

# Verify dashboard exists before update
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name FROM dashboards WHERE id = 'dashboard_id'"

# Common issue: Missing required fields (name, layout)
# All dashboard updates must include name field

# Common issue: Trying to delete template dashboard
# Template dashboards (is_template=1) cannot be deleted
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name, is_template FROM dashboards WHERE id = 'dashboard_id'"
```

**Problem**: Widget data queries timing out or slow
```bash
# Check metric_snapshots table size
sqlite3 ~/.engineerdna/engineerdna.db "SELECT COUNT(*) FROM metric_snapshots"

# Verify indexes exist
sqlite3 ~/.engineerdna/engineerdna.db ".schema metric_snapshots"
# Should have indexes on: metric_name, entity_type, entity_id, period_start

# Check query performance
sqlite3 ~/.engineerdna/engineerdna.db "EXPLAIN QUERY PLAN SELECT * FROM metric_snapshots WHERE metric_name = 'team_score' AND period_start >= '2025-01-01'"

# Common issue: Missing indexes
# Fix: Run migrations to add missing indexes
```

**Problem**: Template dashboards missing or incorrect
```bash
# Check if default templates were created
sqlite3 ~/.engineerdna/engineerdna.db "SELECT id, name, is_template FROM dashboards WHERE is_template = 1"
# Should have 4 templates: IC Dashboard, Team Lead Dashboard, Director Dashboard, Planning Dashboard

# Verify widget counts in templates
sqlite3 ~/.engineerdna/engineerdna.db "SELECT name, json_array_length(json_extract(layout, '$.widgets')) as widget_count FROM dashboards WHERE is_template = 1"
# All templates should have 7 widgets

# Re-create templates if missing
# (requires running CreateDefaultTemplates on startup or via CLI)
```

**Problem**: Dashboard layout changes not persisting
```bash
# Check if UpdateDashboard mutation is being called
# Browser DevTools -> Network -> Filter: /api/dashboards

# Verify layout is being sent correctly
# Should include both layout AND name fields in update request

# Check response status
# 200 = success, 400 = validation error, 500 = server error

# Common issue: Frontend not sending name field
# Fix: Ensure all dashboard mutations include name
```

**Problem**: Widgets displaying old or stale data
```bash
# Check when snapshots were last computed
sqlite3 ~/.engineerdna/engineerdna.db "SELECT metric_name, MAX(computed_at) as last_computed FROM metric_snapshots GROUP BY metric_name"

# Verify snapshot computation is running hourly
# Check cron job or scheduler logs

# Force browser cache clear
# Browser DevTools -> Network -> Disable cache

# Common issue: TanStack Query caching
# staleTime: 30000ms = data cached for 30 seconds
# Refresh page or wait 30 seconds for fresh data
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

### Frontend (Future)
**Implementation**: See `frontend/CLAUDE.md` (when added)
- Vite + React 18 + TypeScript
- TanStack Query for data fetching
- Recharts for visualizations
- Tailwind CSS for styling

---

For detailed implementation information, navigate to the appropriate subdirectory CLAUDE.md file listed above.
