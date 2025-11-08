# Database Package - internal/db

Shared database layer using raw SQL with SQLite.

## Architecture

- **Database**: SQLite 3 (pure Go via modernc.org/sqlite)
- **Driver**: modernc.org/sqlite (CGO-free, cross-platform)
- **Pattern**: Repository/Store with raw SQL
- **Migrations**: Inline migrations in migrations.go
- **Location**: `~/.engineerdna/engineerdna.db`
- **Concurrency**: Single connection, mutex-protected
- **Transactions**: Explicit transaction management

## Directory Structure

```
internal/db/
├── anonymization.go       # Anonymization mapping store
├── attributes.go          # Entity attribute store (PDR-9)
├── audit.go               # Audit log store
├── correlations.go        # Correlation store (PDR-9)
├── events.go              # Event store (core)
├── metrics.go             # Metric value store (PDR-9)
├── migrations.go          # Inline migrations
├── plugin_manifests.go    # Plugin manifest store (PDR-9)
└── plugins.go             # Plugin config store

migrations/             # Reference SQL files (for VCS)
├── 001_initial_schema.sql
├── 002_anonymization.sql
├── 003_audit_log.sql
...
└── 028_multi_modal_data.sql  # PDR-9: Metrics, Attributes, Correlations
```

## Critical Rules

1. **NO one-off scripts** - ALWAYS use migrations
2. **UTC timestamps** - `time.Now().UTC()`, RFC3339
3. **Repository pattern** - No direct DB access outside internal/db
4. **Transactions for batches** - Use `BeginTx()`
5. **Inline migrations** - Define in migrations.go
6. **Idempotent** - Use `IF NOT EXISTS`
7. **No CGO** - modernc.org/sqlite only
8. **Parameterized queries** - `?` placeholders only
9. **JSON for flexible data** - TEXT (JSON) with validation
10. **Single connection** - SQLite = one writer
11. **Versioned migrations** - Track in `schema_migrations`
12. **Wrap errors** - `fmt.Errorf("context: %w", err)`
13. **Close resources** - `defer rows.Close()`
14. **Check ErrNoRows** - Handle explicitly
15. **Index critical queries** - timestamp, actor, source, type

## Key Patterns

### Store Pattern

```go
type EventStore struct {
    db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
    return &EventStore{db: db}
}

func (s *EventStore) Create(event *models.Event) error {
    _, err := s.db.Exec(`INSERT INTO events (...) VALUES (?, ?, ?)`,
        event.ID, event.Type, event.Data)
    return err
}

func (s *EventStore) GetByID(id string) (*models.Event, error) {
    var event models.Event
    err := s.db.QueryRow(`SELECT ... FROM events WHERE id = ?`, id).Scan(&event)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("not found: %s", id)
    }
    return &event, err
}
```

### Migration System

```go
func RunMigrations(db *sql.DB) error {
    db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (...)`)

    migrations := []struct {
        version string
        sql     string
    }{
        {"001_initial", `CREATE TABLE IF NOT EXISTS events (...);`},
        {"002_indexes", `CREATE INDEX idx_events_timestamp ON events(timestamp);`},
    }

    for _, m := range migrations {
        var exists bool
        db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)",
            m.version).Scan(&exists)
        if exists {
            continue
        }

        tx, _ := db.Begin()
        tx.Exec(m.sql)
        tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version)
        tx.Commit()
    }
}
```

### Transaction Pattern

```go
func (s *EventStore) CreateBatch(events []*models.Event) error {
    tx, _ := s.db.Begin()
    defer tx.Rollback() // No-op if committed

    stmt, _ := tx.Prepare(`INSERT INTO events (...) VALUES (?, ?, ?)`)
    defer stmt.Close()

    for _, event := range events {
        stmt.Exec(event.ID, event.Type, event.Data)
    }

    return tx.Commit()
}
```

### JSON Storage

```go
// Store
dataJSON, _ := json.Marshal(event.Data)
db.Exec(`INSERT INTO events (..., data) VALUES (?, ?)`, ..., string(dataJSON))

// Retrieve
var dataJSON string
db.QueryRow(`SELECT ..., data FROM events WHERE id = ?`, id).Scan(..., &dataJSON)
json.Unmarshal([]byte(dataJSON), &event.Data)
```

## Common Commands

```bash
# Development
go vet ./...
go test ./internal/db/...
make build

# Inspect
sqlite3 ~/.engineerdna/engineerdna.db ".tables"
sqlite3 ~/.engineerdna/engineerdna.db ".schema events"
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM events LIMIT 5"

# Migrations
sqlite3 ~/.engineerdna/engineerdna.db "SELECT version FROM schema_migrations ORDER BY applied_at"

# Backup
cp ~/.engineerdna/engineerdna.db{,.backup}
```

## Schema Conventions

### Table Structure

```sql
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,              -- JSON
    timestamp DATETIME NOT NULL,     -- UTC
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_id)
);

CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_actor ON events(actor);
CREATE INDEX idx_events_source ON events(source);
```

### Timestamp Rules

```go
// [GOOD]
event.Timestamp = time.Now().UTC()

// [BAD]
event.Timestamp = time.Now() // Wrong! Local timezone
```

**Storage**: TEXT in RFC3339 format (`"2025-11-04T10:30:00Z"`)

### Multi-Modal Data (PDR-9)

EngineerDNA supports three data types:

**Events** - Discrete timestamped actions
```sql
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,              -- Source type (e.g., pull_request)
    normalized_type TEXT,             -- Normalized type (e.g., code_review)
    normalized_data TEXT,             -- Normalized fields as JSON
    source TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    actor TEXT NOT NULL,
    data TEXT NOT NULL,               -- Source-specific data as JSON
    ...
);
CREATE INDEX idx_events_normalized_type ON events(normalized_type);
```

**Metrics** - Time-series measurements
```sql
CREATE TABLE metric_values (
    id TEXT PRIMARY KEY,
    metric_name TEXT NOT NULL,       -- aws_cost, team_velocity, deploy_frequency
    source TEXT NOT NULL,             -- Plugin that created metric
    timestamp TEXT NOT NULL,          -- ISO 8601
    granularity TEXT NOT NULL,        -- hourly, daily, weekly, monthly
    value REAL NOT NULL,
    unit TEXT,                        -- dollars, hours, count, percentage
    dimensions TEXT,                  -- JSON: {service: ec2, region: us-east-1}
    created_at TEXT NOT NULL
);
CREATE INDEX idx_metric_values_name_timestamp ON metric_values(metric_name, timestamp);
```

**Attributes** - Facts about entities
```sql
CREATE TABLE entity_attributes (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,        -- team, engineer, org
    entity_id TEXT NOT NULL,
    attribute_name TEXT NOT NULL,     -- team_size, budget, location
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,         -- string, number, boolean, json
    valid_from TEXT NOT NULL,         -- When this became valid
    valid_until TEXT,                 -- NULL = still valid
    source TEXT NOT NULL,
    created_at TEXT NOT NULL,
    UNIQUE(entity_type, entity_id, attribute_name, valid_from)
);
CREATE INDEX idx_entity_attributes_lookup ON entity_attributes(entity_type, entity_id, attribute_name);
```

**Correlations** - Cross-data-type relationships
```sql
CREATE TABLE correlations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,       -- cost_per_feature, velocity_vs_team_size
    plugin TEXT NOT NULL,
    definition TEXT NOT NULL,         -- JSON: How to compute
    created_at TEXT NOT NULL
);

CREATE TABLE correlation_values (
    id TEXT PRIMARY KEY,
    correlation_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    time_window TEXT NOT NULL,        -- day, week, month, quarter
    value REAL NOT NULL,
    breakdown TEXT,                   -- JSON: Breakdown by dimension
    created_at TEXT NOT NULL,
    FOREIGN KEY (correlation_id) REFERENCES correlations(id)
);
```

**Plugin Manifests** - Capability registry
```sql
CREATE TABLE plugin_manifests (
    plugin_name TEXT PRIMARY KEY,
    version TEXT NOT NULL,
    type TEXT NOT NULL,               -- source, metric_source, attribute_source, destination, processor
    capabilities TEXT,                -- JSON array
    provides_metrics TEXT,            -- JSON: MetricSpec[]
    provides_event_types TEXT,        -- JSON: EventTypeSpec[]
    provides_widgets TEXT,            -- JSON: WidgetSpec[]
    provides_correlations TEXT,       -- JSON: CorrelationSpec[]
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

## Common Gotchas

### 1. Database Locked

```bash
# Check processes
lsof ~/.engineerdna/engineerdna.db
pkill engineerdna

# Prevention: Keep transactions short
data := callExternalAPI()  // I/O outside transaction
tx.Begin()
tx.Exec("INSERT ...", data)
tx.Commit()
```

### 2. Time Zone Confusion

```go
// [BAD]
event.Timestamp = time.Now()           // Local
event.CreatedAt = time.Now().UTC()     // UTC

// [GOOD]
event.Timestamp = time.Now().UTC()
event.CreatedAt = time.Now().UTC()
```

### 3. Migration Ordering

Never skip version numbers:
```
001_initial_schema
002_anonymization
003_audit_log    # Not 004!
```

### 4. Deferred Rollback

```go
// [BAD]
tx, _ := db.Begin()
if err := tx.Exec(...); err != nil {
    return err  // Never rolled back!
}

// [GOOD]
tx, _ := db.Begin()
defer tx.Rollback()  // No-op if committed
if err := tx.Exec(...); err != nil {
    return err
}
return tx.Commit()
```

### 5. SELECT * Anti-Pattern

```go
// [BAD]
rows, _ := db.Query("SELECT * FROM events")

// [GOOD]
rows, _ := db.Query("SELECT id, type, timestamp FROM events")
```

## Migrations

### Adding New Migration

1. **Update migrations.go**:
```go
{version: "004_add_metadata", sql: `ALTER TABLE plugins ADD COLUMN metadata TEXT;`},
```

2. **Create reference file**:
```bash
cat > migrations/004_add_metadata.sql << 'EOF'
ALTER TABLE plugins ADD COLUMN metadata TEXT;
EOF
```

3. **Test**:
```bash
cp ~/.engineerdna/engineerdna.db{,.backup}
pkill engineerdna
./engineerdna
sqlite3 ~/.engineerdna/engineerdna.db "SELECT version FROM schema_migrations"
```

4. **Commit together**: Migration code + reference SQL + Go code

### Migration Safety

```sql
-- [GOOD] Idempotent
CREATE TABLE IF NOT EXISTS new_table (...);

-- [GOOD] Backward compatible
ALTER TABLE events ADD COLUMN priority INTEGER DEFAULT 0;

-- [BAD] Breaks existing data
ALTER TABLE events ADD COLUMN priority INTEGER NOT NULL;
```

### SQLite Limitations

**Supported**: ADD COLUMN, RENAME COLUMN (3.25+)
**Not Supported**: DROP COLUMN, MODIFY COLUMN, ADD CONSTRAINT

**Workaround** (recreate table):
```sql
CREATE TABLE events_new (...);
INSERT INTO events_new SELECT ... FROM events;
DROP TABLE events;
ALTER TABLE events_new RENAME TO events;
```

## Testing

### Table-Driven Tests

```go
func TestEventStore_Create(t *testing.T) {
    db, _ := sql.Open("sqlite", ":memory:")
    defer db.Close()

    RunMigrations(db)
    store := NewEventStore(db)

    tests := []struct {
        name    string
        event   *models.Event
        wantErr bool
    }{
        {"valid", &models.Event{...}, false},
        {"duplicate", &models.Event{...}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := store.Create(tt.event)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mocking

```go
type EventStore interface {
    Create(event *models.Event) error
    GetByID(id string) (*models.Event, error)
}

type MockEventStore struct {
    events []*models.Event
}

func (m *MockEventStore) Create(e *models.Event) error {
    m.events = append(m.events, e)
    return nil
}
```

## Performance

### Query Optimization

```go
// [BAD] N+1 queries
for _, event := range events {
    user, _ := userStore.GetByID(event.Actor)  // N queries!
}

// [GOOD] JOIN
query := `SELECT e.*, u.name FROM events e LEFT JOIN users u ON e.actor = u.id`
```

### Index Strategy

```sql
-- Single-column
CREATE INDEX idx_events_timestamp ON events(timestamp);

-- Composite (for specific queries)
CREATE INDEX idx_events_source_timestamp ON events(source, timestamp);

-- Analyze
EXPLAIN QUERY PLAN SELECT * FROM events WHERE source = 'github';
```

### Batch Operations

```go
// [BAD]
for _, event := range events {
    store.Create(event)  // N transactions!
}

// [GOOD]
store.CreateBatch(events)  // Single transaction
```

### Connection Config

```go
db.SetMaxOpenConns(1)  // SQLite = 1
db.SetMaxIdleConns(1)
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA busy_timeout=5000")
```

## Security

### SQL Injection Prevention

```go
// [BAD]
query := fmt.Sprintf("SELECT * FROM events WHERE actor = '%s'", actor)
db.Query(query)  // Vulnerable!

// [GOOD]
db.Query("SELECT * FROM events WHERE actor = ?", actor)
```

### Path Traversal Prevention

```go
// [GOOD]
home, _ := os.UserHomeDir()
basePath := filepath.Join(home, ".engineerdna")
fullPath := filepath.Join(basePath, filepath.Clean(userPath))
if !strings.HasPrefix(fullPath, basePath) {
    return errors.New("invalid path")
}
```

## Debugging

### Common Issues

**Database locked**:
```bash
lsof ~/.engineerdna/engineerdna.db
pkill engineerdna
```

**Migration failed**:
```bash
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM schema_migrations"
```

**Slow queries**:
```sql
EXPLAIN QUERY PLAN SELECT * FROM events WHERE actor = 'user@example.com';
```

### Useful Queries

```sql
-- Database size
SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size();

-- List indexes
SELECT name, tbl_name FROM sqlite_master WHERE type = 'index';

-- Find duplicates
SELECT source, source_id, COUNT(*) FROM events
GROUP BY source, source_id HAVING COUNT(*) > 1;

-- Integrity check
PRAGMA integrity_check;
```

## Environment Variables

```bash
ENGINEERDNA_DB_PATH=/custom/path/engineerdna.db
ENGINEERDNA_MASTER_KEY=your-32-byte-key
ENGINEERDNA_DB_TIMEOUT=30s
```

## References

- Rule 33: NO one-off scripts, always use migrations
- Rule 34: UTC timestamps everywhere
- Hook: `validate_schema_changes.py` enforces migrations
- Hook: `validate_db_operations.py` blocks direct SQL
- Skill: `database-migrations` for detailed workflows
