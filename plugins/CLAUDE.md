# Plugin Development Guide - plugins/

Plugin system for extending EngineerDNA with data sources, destinations, and processors.

## Overview

EngineerDNA uses **subprocess plugin architecture** with JSON-RPC over stdin/stdout:

- **Process Isolation**: Each plugin runs in separate process
- **JSON-RPC Protocol**: Standard request/response over stdin/stdout
- **Language Agnostic**: Any language that can read stdin/write stdout
- **Sandboxed**: No database access, no encryption keys
- **Timeout Protected**: 30-second timeout per operation

## Plugin Types

### Source Plugins

Bring **event data** IN to EngineerDNA (GitHub, Jira, GitLab).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `source.sync`

**Optional**: Declare `provides_event_types` for event normalization.

### Metric Source Plugins (PDR-9)

Bring **metric data** IN to EngineerDNA (AWS Cost, team capacity, budget tracking).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `metric_source.sync`

**Must Declare**: `provides_metrics` in plugin.json with metric specifications.

### Attribute Source Plugins (PDR-9)

Bring **entity attributes** IN to EngineerDNA (HRIS, org charts, team metadata).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `attribute_source.sync`

**Must Handle**: Temporal validity (valid_from/valid_until) for attribute changes.

### Destination Plugins

Send data **OUT** (Google Sheets, Slack, PDF, Webhooks).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `destination.test`, `destination.export`

### Processor Plugins

Transform or analyze data (AI Insights, Custom Metrics, Correlation calculation).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `processor.capabilities`, `processor.analyze`

**CRITICAL**: Must declare `anonymization.required: true` if sending to external APIs.

**Optional**: Declare `provides_correlations` for cross-data-type relationships.

## Directory Structure

```
plugins/
├── CLAUDE.md                    # This file
├── plugin-sdk/                  # Go SDK
│   ├── types.go                # Core types
│   ├── protocol.go             # JSON-RPC helpers
│   └── sdk.go                  # Plugin server
├── github/                      # Source plugin
│   ├── plugin.json
│   ├── main.go
│   └── github                  # Binary
├── google-sheets-export/        # Destination plugin
├── ai-insights/                 # Processor plugin
└── csv-import/                  # Source plugin
```

## Critical Rules

1. **Subprocess isolation** - Always run as separate process
2. **JSON-RPC only** - stdin/stdout, JSON-RPC 2.0
3. **30-second timeout** - All operations complete in 30s
4. **No direct DB access** - Cannot access EngineerDNA database
5. **Secret fields encrypted** - `secret: true` encrypted at rest
6. **Anonymization required** - Processor plugins → external APIs MUST anonymize
7. **Rate limit handling** - Handle API rate limits gracefully
8. **UTC timestamps** - All timestamps in UTC RFC3339
9. **Error codes** - Use SDK error codes (Auth=1001, etc.)
10. **Idempotent sync** - Same `since` returns same events/metrics/attributes
11. **Binary naming** - Binary matches directory name
12. **plugin.json required** - Valid manifest in each plugin
13. **No network during configure** - Configuration must be fast
14. **Health check quick** - Return within 5s
15. **Graceful degradation** - Return partial results + warnings
16. **Declare capabilities** - Use `provides_metrics`, `provides_event_types`, `provides_widgets`, `provides_correlations`
17. **Event normalization** - Map source fields to normalized fields in `provides_event_types`
18. **Metric dimensions** - Support multi-dimensional metrics (service, region, etc.)
19. **Temporal attributes** - Handle valid_from/valid_until for attribute changes
20. **Register plugin manifest** - Host automatically registers capabilities on first info call

## Quick Start with SDK

### Create Plugin

```go
package main

import (
    "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
    "time"
)

type MyPlugin struct {
    apiKey string
}

func (p *MyPlugin) Info() sdk.PluginInfo {
    return sdk.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Type:        "source",
        Description: "My custom source plugin",
        ConfigFields: []sdk.ConfigField{
            {Name: "api_key", Type: "password", Required: true, Secret: true},
        },
    }
}

func (p *MyPlugin) Configure(config map[string]interface{}) error {
    p.apiKey = config["api_key"].(string)
    return nil
}

func (p *MyPlugin) Health() (sdk.HealthResult, error) {
    return sdk.HealthResult{Healthy: true, Message: "OK"}, nil
}

func (p *MyPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
    events := []sdk.Event{
        {
            Type:      "custom_event",
            Source:    "my-plugin",
            SourceID:  "evt_123",
            Timestamp: time.Now().UTC(),
            Actor:     "user@example.com",
            Data:      map[string]interface{}{"key": "value"},
        },
    }
    return sdk.SyncResult{Events: events}, nil
}

func main() {
    plugin := &MyPlugin{}
    server := sdk.NewPluginServer(plugin)
    server.Start()
}
```

### Build & Test

```bash
# Build
cd plugins/my-plugin
go build -o my-plugin
chmod +x my-plugin

# Test
echo '{"jsonrpc":"2.0","method":"plugin.info","id":"1"}' | ./my-plugin | jq .
echo '{"jsonrpc":"2.0","method":"plugin.configure","params":{"api_key":"test"},"id":"2"}' | ./my-plugin
echo '{"jsonrpc":"2.0","method":"source.sync","params":{"since":"2025-01-01T00:00:00Z"},"id":"3"}' | ./my-plugin
```

## JSON-RPC Protocol

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "plugin.info",
  "params": {...},
  "id": "request-id-123"
}
```

### Response (Success)

```json
{
  "jsonrpc": "2.0",
  "result": {...},
  "id": "request-id-123"
}
```

### Response (Error)

```json
{
  "jsonrpc": "2.0",
  "error": {"code": 1001, "message": "Authentication failed"},
  "id": "request-id-123"
}
```

### Error Codes

- `1000` - Configuration error
- `1001` - Authentication failure
- `1002` - API rate limit
- `1003` - Network/timeout
- `2000` - Internal plugin error

## Universal Methods

All plugins implement:

**plugin.info**: Return plugin metadata
**plugin.configure**: Accept configuration
**plugin.health**: Quick health check

## Source Methods

**source.sync**: Fetch events since timestamp

```json
Request: {"method": "source.sync", "params": {"since": "2025-11-01T00:00:00Z"}}
Response: {
  "result": {
    "events": [{"type": "pull_request", "source": "github", ...}],
    "warnings": ["Skipped 2 draft PRs"]
  }
}
```

## Destination Methods

**destination.test**: Test connection
**destination.export**: Export data

```json
Request: {
  "method": "destination.export",
  "params": {
    "data_type": "events",
    "data": {"events": [...], "period_start": "...", "period_end": "..."},
    "options": {"format": "summary"}
  }
}
Response: {
  "result": {
    "status": "success",
    "url": "https://docs.google.com/spreadsheets/d/abc123",
    "rows_written": 100
  }
}
```

## Processor Methods

**processor.capabilities**: List analysis types
**processor.analyze**: Analyze events

```json
Request: {
  "method": "processor.analyze",
  "params": {
    "analysis_type": "team-insights",
    "events": [...],
    "context": {"team_size": 8}
  }
}
Response: {
  "result": {
    "insights": [
      {
        "severity": "warning",
        "title": "High PR review time",
        "description": "Average review time: 3.5 days",
        "recommendation": "Consider setting review SLAs"
      }
    ]
  }
}
```

## Metric Source Methods (PDR-9)

**metric_source.sync**: Fetch metric values since timestamp

```json
Request: {
  "method": "metric_source.sync",
  "params": {"since": "2025-11-01T00:00:00Z"}
}
Response: {
  "result": {
    "metrics": [
      {
        "metric_name": "aws_cost",
        "timestamp": "2025-11-08T00:00:00Z",
        "granularity": "daily",
        "value": 1234.56,
        "unit": "dollars",
        "dimensions": {"service": "ec2", "region": "us-east-1"}
      }
    ],
    "warnings": []
  }
}
```

## Attribute Source Methods (PDR-9)

**attribute_source.sync**: Fetch entity attributes

```json
Request: {
  "method": "attribute_source.sync",
  "params": {"since": "2025-11-01T00:00:00Z"}
}
Response: {
  "result": {
    "attributes": [
      {
        "entity_type": "team",
        "entity_id": "team-backend",
        "attribute_name": "team_size",
        "value": "8",
        "value_type": "number",
        "valid_from": "2025-11-01T00:00:00Z",
        "valid_until": null
      }
    ],
    "warnings": []
  }
}
```

## plugin.json Specification

### Minimal Example

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "type": "source",
  "description": "Fetches data from My Service",
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
    "required": false
  }
}
```

### ConfigField Types

- `string` - Text input
- `password` - Password (masked in UI)
- `boolean` - Checkbox
- `select` - Dropdown (requires `options`)
- `json` - JSON object

### Anonymization Spec

```json
{
  "anonymization": {
    "required": true,
    "reason": "Data sent to OpenAI API",
    "strategy": "sequential",
    "fields": ["actor", "author", "committer"]
  }
}
```

**Strategies**: `sequential` (User_1, User_2), `uuid` (random UUIDs), `hash` (SHA-256)

### PDR-9 Capability Declarations

**Metric Source Plugin**:
```json
{
  "type": "metric_source",
  "provides_metrics": [
    {
      "name": "aws_cost",
      "description": "AWS infrastructure cost",
      "unit": "dollars",
      "granularity": ["hourly", "daily", "monthly"],
      "dimensions": ["service", "region"]
    }
  ]
}
```

**Source Plugin with Event Normalization**:
```json
{
  "type": "source",
  "provides_event_types": [
    {
      "type": "pull_request",
      "description": "GitHub pull request events",
      "normalized_type": "code_review",
      "normalization_map": {
        "title": "title",
        "author": "user.login",
        "state": "state",
        "created_at": "created_at",
        "merged_at": "merged_at"
      }
    }
  ]
}
```

**Plugin with Custom Widgets**:
```json
{
  "provides_widgets": [
    {
      "type": "aws_cost_breakdown",
      "name": "AWS Cost Breakdown",
      "description": "Shows costs by service and region"
    }
  ]
}
```

**Processor Plugin with Correlations**:
```json
{
  "type": "processor",
  "provides_correlations": [
    {
      "name": "cost_per_feature",
      "description": "Engineering cost per feature delivered",
      "inputs": ["aws_cost", "feature_work"]
    }
  ]
}
```

## Common Patterns

### Common Patterns Examples

```go
// Rate limiting
func (p *Plugin) checkRateLimit(resp *github.Response) error {
    if resp.Rate.Remaining < 100 {
        time.Sleep(time.Until(resp.Rate.Reset.Time))
    }
    return nil
}

// Error handling + partial results + timeout
func (p *Plugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
    defer cancel()

    events, warnings := []sdk.Event{}, []string{}
    for _, item := range fetchedItems {
        if event, err := p.convertToEvent(item); err != nil {
            warnings = append(warnings, fmt.Sprintf("Skipped %s: %v", item.ID, err))
        } else {
            events = append(events, event)
        }
    }
    return sdk.SyncResult{Events: events, Warnings: warnings}, nil
}

// Configuration validation
func (p *Plugin) Configure(config map[string]interface{}) error {
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" || !strings.HasPrefix(apiKey, "sk-") {
        return sdk.NewConfigError("invalid api_key")
    }
    p.apiKey = apiKey
    return nil
}
```

## Configuration and Secrets

### Secret Field Encryption

Fields with `secret: true` are encrypted at rest:

```json
{
  "config_fields": [
    {"name": "api_key", "secret": true},      // Encrypted
    {"name": "workspace", "secret": false}     // Plaintext
  ]
}
```

**Flow**:
1. User enters API key
2. Host encrypts with AES-256-GCM
3. Stored encrypted in database
4. Host decrypts before sending to plugin
5. Plugin never sees encryption keys

### Environment Variables

```go
func (p *Plugin) Configure(config map[string]interface{}) error {
    apiKey, ok := config["api_key"].(string)
    if !ok {
        apiKey = os.Getenv("MY_PLUGIN_API_KEY")
    }
    if apiKey == "" {
        return sdk.NewConfigError("api_key required")
    }
    p.apiKey = apiKey
    return nil
}
```

## Anonymization

### When Required

**MANDATORY** for processor plugins sending to external APIs:

```json
{
  "type": "processor",
  "anonymization": {
    "required": true,
    "reason": "Events sent to OpenAI",
    "strategy": "sequential",
    "fields": ["actor", "author", "committer"]
  }
}
```

### Flow

1. Host fetches events
2. Host checks `anonymization.required == true`
3. Host anonymizes fields (alice@example.com → User_1)
4. Host sends anonymized events to plugin
5. Plugin sends anonymized data to external API
6. Plugin returns insights with anonymized IDs
7. Host deanonymizes before display (User_1 → alice@example.com)

### Mapping Storage

Bidirectional mapping in database:

```sql
CREATE TABLE anonymization_map (
    anonymized_id TEXT UNIQUE,  -- "User_1"
    real_name TEXT,              -- "alice@example.com"
    enabled BOOLEAN DEFAULT 1
);
```

### Audit Trail

All anonymization logged:

```sql
CREATE TABLE audit_log (
    plugin_name TEXT,
    action TEXT,
    event_count INTEGER,
    anonymized BOOLEAN,  -- Always true for processors
    timestamp DATETIME
);
```

## Testing & Debugging

### Manual Testing

```bash
cd plugins/my-plugin && go build -o my-plugin
echo '{"jsonrpc":"2.0","method":"plugin.info","id":"1"}' | ./my-plugin | jq .
echo '{"jsonrpc":"2.0","method":"plugin.configure","params":{"api_key":"test"},"id":"2"}' | ./my-plugin
```

### Plugin Logs

```go
fmt.Fprintf(os.Stderr, "Syncing events since %s\n", params.Since)  // Captured by host
```

### Common Issues

- **Plugin timeout**: Paginate results, return partial data
- **Configuration not persisted**: SDK handles automatically
- **Invalid JSON-RPC**: Use stderr for debug, stdout for JSON-RPC only
- **Secret not decrypted**: Report bug (shouldn't happen)

### PDR-9 Plugin Issues

**Metric source not syncing**:
```bash
# Check if plugin declares provides_metrics
cat plugin.json | jq '.provides_metrics'

# Verify metric_source.sync method is implemented
echo '{"jsonrpc":"2.0","method":"metric_source.sync","params":{"since":"2025-01-01T00:00:00Z"},"id":"1"}' | ./my-plugin

# Common issue: Wrong plugin type
# Fix: Ensure "type": "metric_source" in plugin.json
```

**Attribute source validation errors**:
```bash
# Check if attributes have required fields
# Must have: entity_type, entity_id, attribute_name, value, value_type, valid_from

# Common issue: Missing valid_from
# Fix: Always set valid_from to when attribute became valid (use timestamp if current)

# Common issue: valid_until set for current attributes
# Fix: Set valid_until = null for current attributes, only set when attribute changes
```

**Event normalization not working**:
```bash
# Verify provides_event_types is declared
cat plugin.json | jq '.provides_event_types'

# Check normalization_map is complete
# Must map all fields used by normalized type

# Common issue: Missing normalized_type field
# Fix: Add normalized_type to provides_event_types
```

**Custom widget not appearing**:
```bash
# Check widget declaration
cat plugin.json | jq '.provides_widgets'

# Verify host registered widget
curl http://127.0.0.1:3847/api/widgets/registry | jq '.widgets[] | select(.plugin == "my-plugin")'

# Common issue: Frontend component not implemented
# Custom widgets require React component in frontend
# Fix: Either use standard widget types or implement custom component
```

**Metric dimensions not queryable**:
```bash
# Ensure dimensions is a JSON object, not array
# [GOOD] "dimensions": {"service": "ec2", "region": "us-east-1"}
# [BAD]  "dimensions": ["ec2", "us-east-1"]

# Check dimension keys match provides_metrics
cat plugin.json | jq '.provides_metrics[].dimensions'
```

## Security

```go
// Input validation
func (p *Plugin) Configure(config map[string]interface{}) error {
    apiKey := config["api_key"].(string)
    if len(apiKey) < 10 || len(apiKey) > 200 {
        return sdk.NewConfigError("invalid key length")
    }
    if !strings.HasPrefix(baseURL, "https://") {
        return sdk.NewConfigError("HTTPS required")
    }
    return nil
}

// Credential redaction
fmt.Fprintf(os.Stderr, "API key: %s\n", apiKey[:4]+"****")  // [GOOD]

// No shell injection
cmd := exec.Command("git", "clone", userRepo)  // [GOOD] Direct exec
```

## Example Implementations

Reference implementations:
- **plugins/github/** - Source plugin (GitHub API)
- **plugins/csv-import/** - Source plugin (file parsing)
- **plugins/google-sheets-export/** - Destination plugin (OAuth + Sheets)
- **plugins/ai-insights/** - Processor plugin (OpenAI + anonymization)

## References

- Rule 8: Plugin isolation (subprocess, timeout, no DB)
- Rule 9: Anonymization required for external APIs
- Rule 19: Encrypt secrets at rest
- Rule 34: UTC timestamps everywhere
- Skill: `plugin-development` - Detailed workflows
- Agent: `integration-checker` - Plugin system testing
