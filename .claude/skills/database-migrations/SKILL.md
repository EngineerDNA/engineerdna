---
name: database-migrations
description: Safe database schema evolution with SQLite migrations. Use when making schema changes. Enforces migrations-only approach (no one-off scripts, Rule 33).
tools: Read, Bash, Grep, Glob
disallowedTools: Write, Edit
---

# Safe Database Migrations

## Purpose

Master database schema evolution with SQLite migrations. **NEVER use one-off scripts** - always use migrations for reproducibility, version control, and safety.

## Quick Start

### The Migration Workflow

1. **Create Migration File** - Create `migrations/00X_description.sql`
2. **Write SQL** - Add your schema changes
3. **Test Migration** - Apply with migration command
4. **Test Application** - Verify with `./bin/engineerdna`
5. **Commit Migration** - Stage migration file with code changes

```bash
# Example workflow
# 1. Create migration file
cat > migrations/004_add_plugin_config.sql << 'EOF'
CREATE TABLE IF NOT EXISTS plugin_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    config_data TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_plugin_configs_plugin_id ON plugin_configs(plugin_id);
EOF

# 2. Apply migration (when implemented)
./bin/engineerdna migrate up

# 3. Test the application
./bin/engineerdna

# 4. Verify in database
sqlite3 ~/.engineerdna/engineerdna.db ".schema plugin_configs"

# 5. Commit together
git add migrations/004_add_plugin_config.sql internal/db/
git commit -m "Add plugin_configs table"
```

## Critical Rule: No One-Off Scripts

**NEVER** create standalone database scripts. **ALWAYS** use migrations in `migrations/` directory.

**Bad (BLOCKED by hooks):**
```bash
# BLOCKED - Direct DB modification
sqlite3 ~/.engineerdna/engineerdna.db "ALTER TABLE events ADD COLUMN anonymized INTEGER"
```

**Good:**
```bash
# Create migration file
cat > migrations/005_add_anonymized_flag.sql << 'EOF'
ALTER TABLE events ADD COLUMN anonymized INTEGER DEFAULT 0;
EOF

# Apply migration
./bin/engineerdna migrate up
```

## Migration Checklist

Before committing:
- [ ] Created numbered migration file in `migrations/`
- [ ] Wrote clear SQL with schema changes
- [ ] Tested migration locally
- [ ] Verified in database browser (sqlite3 or GUI)
- [ ] Updated Go models in `internal/db/` or `internal/models/`
- [ ] Tested application code works with new schema
- [ ] Committed migration AND code together

## Common Migration Patterns

### Add New Table

```sql
-- migrations/006_add_audit_log.sql
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    user_id TEXT,
    details TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
```

### Add Column

```sql
-- migrations/007_add_event_metadata.sql
ALTER TABLE events ADD COLUMN metadata TEXT;
```

### Add Index

```sql
-- migrations/008_index_events_timestamp.sql
CREATE INDEX idx_events_timestamp ON events(timestamp);
```

### Modify Data

```sql
-- migrations/009_migrate_legacy_data.sql
-- Migrations can also update existing data
UPDATE events
SET anonymized = 1
WHERE source IN ('ai-insights', 'external-processor');
```

## Migration Naming

Format: `XXX_description.sql`

- **XXX** - Sequential number (001, 002, 003, etc.)
- **description** - Clear snake_case description

Examples:
- `001_create_initial_schema.sql`
- `002_add_encryption_keys_table.sql`
- `003_add_anonymization_mapping.sql`
- `004_add_plugin_configs.sql`

## Rollback Strategy

SQLite doesn't have built-in transaction support for DDL in all cases. For critical migrations:

1. **Backup first:**
   ```bash
   cp ~/.engineerdna/engineerdna.db ~/.engineerdna/engineerdna.db.backup
   ```

2. **Test in dev environment**

3. **Create reverse migration if needed:**
   ```sql
   -- migrations/010_add_feature_column.sql
   ALTER TABLE events ADD COLUMN feature_flag INTEGER DEFAULT 0;

   -- migrations/011_rollback_feature_column.sql (if needed)
   -- SQLite doesn't support DROP COLUMN easily
   -- May need to recreate table without column
   ```

## Verifying Migrations

After applying migration:

```bash
# Check table structure
sqlite3 ~/.engineerdna/engineerdna.db ".schema events"

# Check indexes
sqlite3 ~/.engineerdna/engineerdna.db ".indexes events"

# Verify data
sqlite3 ~/.engineerdna/engineerdna.db "SELECT * FROM events LIMIT 5"
```

## Troubleshooting

### Migration Already Applied

Migrations should be idempotent when possible:

```sql
-- Use IF NOT EXISTS
CREATE TABLE IF NOT EXISTS my_table (...);

-- Check before adding column (SQLite limitation workaround)
-- Can't easily add "IF NOT EXISTS" for columns in SQLite
-- Best practice: Track applied migrations
```

### Schema Out of Sync

```bash
# Check current schema
sqlite3 ~/.engineerdna/engineerdna.db ".schema"

# Compare with migrations
ls -1 migrations/
```

If out of sync, apply missing migrations in order.

### Database Locked

```bash
# Check for processes using database
lsof ~/.engineerdna/engineerdna.db

# Kill if needed
pkill -f engineerdna
```

## Best Practices

1. **One Concept Per Migration** - Don't mix unrelated changes
2. **Test Locally First** - Always test before committing
3. **Use Transactions** - Wrap multiple operations when possible
4. **Document Complex Changes** - Add comments in SQL
5. **Backup Production** - Before applying migrations
6. **Sequential Numbers** - Keep migrations ordered
7. **Commit Together** - Migration + code that uses it

## Integration with Code

When adding a migration, update corresponding Go code:

```go
// migrations/012_add_plugin_metadata.sql
ALTER TABLE plugins ADD COLUMN metadata TEXT;

// internal/models/plugin.go
type Plugin struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Metadata string `json:"metadata"` // NEW
}
```

Always commit migration and code changes together.

## References

- Rule: NO one-off DB scripts, always use migrations
- Rule: UTC timestamps everywhere (`time.Now().UTC()`)
- Hook: `validate_schema_changes.py` enforces migrations
- Hook: `validate_db_operations.py` blocks direct DB commands
