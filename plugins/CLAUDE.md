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

Bring data **IN** to EngineerDNA (GitHub, Jira, CSV).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `source.sync`

### Destination Plugins

Send data **OUT** (Google Sheets, Slack, PDF, Webhooks).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `destination.test`, `destination.export`

### Processor Plugins

Transform or analyze data (AI Insights, Custom Metrics, Validation).

**Required Methods**: `plugin.info`, `plugin.configure`, `plugin.health`, `processor.capabilities`, `processor.analyze`

**CRITICAL**: Must declare `anonymization.required: true` if sending to external APIs.

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
10. **Idempotent sync** - Same `since` returns same events
11. **Binary naming** - Binary matches directory name
12. **plugin.json required** - Valid manifest in each plugin
13. **No network during configure** - Configuration must be fast
14. **Health check quick** - Return within 5s
15. **Graceful degradation** - Return partial results + warnings

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
