# CSV Import Plugin - plugins/csv-import/

Source plugin for importing engineering metrics from CSV files.

## Purpose

Simple file-based import for:
- Testing and demos
- Migrating from other tools
- Manual metric tracking
- Batch data import

## CSV Format

Required columns:

```csv
timestamp,actor,type,description
2024-01-15,alice,pull_request,Merged PR #123: Add authentication
2024-01-16,bob,issue,Closed issue #456: Fix login bug
2024-01-17,alice,pull_request,Merged PR #124: Update README
```

### Column Specifications

- **timestamp**: RFC3339 or ISO 8601 date (YYYY-MM-DD)
- **actor**: Email or username
- **type**: Event type (pull_request, issue, commit, etc.)
- **description**: Human-readable description

## Configuration

Single field: `file_path` (absolute path to CSV file)

```json
{
  "file_path": "/tmp/test-events.csv"
}
```

## Quick Start

### 1. Create Test CSV

```bash
cat > /tmp/test-events.csv <<EOF
timestamp,actor,type,description
2024-01-15,alice,pull_request,Merged PR #123: Add authentication
2024-01-16,bob,issue,Closed issue #456: Fix login bug
2024-01-17,alice,pull_request,Merged PR #124: Update README
2024-01-18,charlie,pull_request,Opened PR #125: Add new feature
EOF
```

### 2. Build Plugin

```bash
cd plugins/csv-import
go build -o csv-import main.go
```

### 3. Configure via API

```bash
curl -X POST http://localhost:3847/api/plugins/csv-import/configure \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/tmp/test-events.csv"}'
```

### 4. Sync Data

```bash
curl -X POST http://localhost:3847/api/plugins/csv-import/sync
```

### 5. Verify Import

```bash
curl http://localhost:3847/api/events | jq '.events[] | select(.source == "csv-import")'
```

## Implementation Details

- Reads entire file on each sync
- Filters by `since` timestamp parameter
- Converts CSV rows to Event objects
- Handles malformed rows gracefully (warning in logs)
- No external API calls (no anonymization required)

## Error Handling

- Missing file: Returns config error
- Invalid CSV format: Skips row, adds warning
- Permission denied: Returns error with file path

## Limitations

- No incremental sync (reads full file each time)
- No file watching (manual sync only)
- Maximum file size: 10MB (performance)
- No CSV header validation

## Use Cases

**Testing**: Generate demo data for development
**Migration**: Import from Jira CSV exports
**Manual Entry**: Track metrics not in GitHub
**Batch Import**: Load historical data

## References

- Plugin SDK: plugins/plugin-sdk/
- Plugin Protocol: plugins/CLAUDE.md
- Main plugin guide: plugins/CLAUDE.md
