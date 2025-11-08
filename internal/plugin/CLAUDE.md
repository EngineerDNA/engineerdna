# Plugin System - internal/plugin

Host-side plugin system for discovering, loading, and executing plugins.

## Overview

**Architecture**:
- **Subprocess Isolation**: Each plugin in separate process
- **JSON-RPC Protocol**: Communication over stdin/stdout
- **Timeout Enforcement**: 30-second timeout per operation
- **Resource Limits**: Memory and stderr limits
- **Security Isolation**: No DB access, no encryption keys, path validation
- **Configuration Encryption**: AES-256-GCM for secret fields
- **Anonymization Flow**: Automatic for processor plugins

**Key Components**:
- **Loader**: Plugin discovery and validation
- **Executor**: Plugin lifecycle with anonymization
- **Client**: JSON-RPC communication with subprocess
- **Protocol**: Request/response with timeout

## Directory Structure

```
internal/plugin/
├── loader.go                  # Discovery and validation
├── executor.go                # Execution with anonymization (events)
├── executor_metric.go         # Metric source execution (PDR-9)
├── executor_attribute.go      # Attribute source execution (PDR-9)
├── event_registry.go          # Event type normalization registry (PDR-9)
├── widget_registry.go         # Widget registry (PDR-9)
├── protocol.go                # JSON-RPC client (cross-platform)
├── protocol_unix.go           # Unix resource limits
└── protocol_windows.go        # Windows resource limits

internal/models/
├── plugin.go                  # Plugin data models
├── metric.go                  # Metric models (PDR-9)
├── attribute.go               # Attribute models (PDR-9)
├── correlation.go             # Correlation models (PDR-9)

internal/services/
├── metric_engine.go           # Metric calculation engine (PDR-9)
├── correlation_engine.go      # Correlation computation (PDR-9)

internal/anonymization/
├── service.go                 # Anonymization service

internal/config/
├── encryption.go              # AES-256-GCM encryption
```

## Critical Rules

1. **Subprocess isolation** - Plugins NEVER in-process
2. **30-second timeout** - All operations timeout after 30s
3. **Path validation** - Prevent path traversal
4. **No symlink exploits** - Resolve, ensure within plugin dir
5. **Metadata size limit** - plugin.json max 1MB
6. **Response size limit** - Plugin responses max 100MB
7. **Stderr limit** - Max 10MB (prevents log flooding)
8. **Config encryption** - Encrypt secret fields before DB
9. **Config decryption** - Decrypt before sending to plugin
10. **Anonymization enforcement** - Processor plugins MUST anonymize for external APIs
11. **Audit all operations** - Log all sync/export/analyze
12. **Kill on timeout** - Process killed if exceeds 30s
13. **Resource limits** - OS-level limits (memory, FDs)
14. **No shell injection** - Direct exec only
15. **Validate plugin type** - Must be source, metric_source, attribute_source, destination, or processor
16. **Register plugin manifests** - Store provides_metrics, provides_event_types, provides_widgets in DB
17. **Event normalization** - Apply event type registry for multi-source support
18. **Metric dimensions** - Support multi-dimensional metrics (service, region, environment)
19. **Temporal attributes** - Track valid_from/valid_until for entity attributes

## Key Components

### Loader (loader.go)

Discovers and loads plugins from filesystem.

```go
type Loader struct {
    pluginDirs []string  // ~/.engineerdna/plugins/, ./plugins/
}

plugins, _ := loader.DiscoverPlugins()
pluginPath, _ := loader.GetPluginPath("github")
```

**Security Features**:
- Name validation: `^[a-zA-Z0-9_-]+$`
- Symlink resolution: Within plugin directory only
- File size check: plugin.json max 1MB
- JSON validation: Rejects malformed

**Discovery Algorithm**:
1. Scan each directory in pluginDirs
2. For each subdirectory:
   - Validate plugin name
   - Read plugin.json (with size check)
   - Parse and validate JSON
   - Find executable (name, main, name.exe)
   - Resolve symlinks, validate path
   - Create PluginEntry

### Executor (executor.go)

Manages plugin execution with anonymization and audit.

```go
type Executor struct {
    loader        *Loader
    anonService   *anonymization.Service
    anonStore     *db.AnonymizationStore
    pluginStore   *db.PluginStore
    auditStore    *db.AuditStore
    keyStore      *config.KeyStore
}

// Configure plugin
executor.ConfigurePlugin("github", config)

// Sync source
events, _ := executor.SyncSourcePlugin("github", since)

// Export to destination
executor.ExportToDestination("google-sheets", events, options)

// Analyze with processor
insights, _ := executor.AnalyzeWithProcessor("ai-insights", events, analysisType)
```

**Configuration Flow**:
1. Get plugin info (identify secret fields)
2. Encrypt secret fields with AES-256-GCM
3. Save encrypted config to database
4. Start plugin subprocess
5. Send decrypted config to plugin.configure
6. Verify plugin accepts config

**Sync Flow** (Source):
1. Load config from database
2. Decrypt secret fields
3. Start plugin subprocess
4. Call plugin.configure
5. Call source.sync
6. Parse events
7. Update last_sync timestamp
8. Return events

**Export Flow** (Destination):
1. Load config and info
2. Check anonymization requirement
3. If required, anonymize events
4. Start plugin subprocess
5. Call plugin.configure
6. Call destination.export
7. Log export in audit_log
8. Return result

**Analysis Flow** (Processor):
1. Load config and info
2. Check anonymization (MUST be true for external APIs)
3. Anonymize events (required fields)
4. Start plugin subprocess
5. Call plugin.configure
6. Call processor.analyze
7. Deanonymize insights
8. Log analysis in audit_log
9. Return deanonymized insights

**Metric Sync Flow** (Metric Source - PDR-9):
1. Load config from database
2. Decrypt secret fields
3. Start plugin subprocess
4. Call plugin.configure
5. Call metric_source.sync
6. Parse metric values (name, timestamp, granularity, value, dimensions)
7. Validate metric specs match plugin manifest
8. Store in metric_values table
9. Update last_sync timestamp
10. Return metric values

**Attribute Sync Flow** (Attribute Source - PDR-9):
1. Load config from database
2. Decrypt secret fields
3. Start plugin subprocess
4. Call plugin.configure
5. Call attribute_source.sync
6. Parse entity attributes (entity_type, entity_id, attribute_name, value, valid_from)
7. Handle temporal validity (expire old attributes with valid_until)
8. Store in entity_attributes table
9. Update last_sync timestamp
10. Return attributes

### Client/Protocol (protocol.go)

JSON-RPC communication with subprocess.

```go
type Client struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout *bufio.Reader
    stderr io.ReadCloser
}

client, _ := NewClient(pluginPath)
defer client.Close()

result, _ := client.Call("plugin.info", nil)
```

**Communication Pattern**:
1. Start subprocess: `exec.Command(pluginPath)`
2. Create stdin/stdout/stderr pipes
3. Forward stderr (with 10MB limit)
4. Send JSON-RPC request to stdin
5. Read JSON-RPC response from stdout (100MB limit)
6. Enforce 30-second timeout
7. Kill process on timeout

**Timeout Implementation**:
```go
func (c *Client) Call(method string, params interface{}) (interface{}, error) {
    c.stdin.Write(jsonRequest)

    done := make(chan []byte, 1)
    go func() {
        response, _ := c.stdout.ReadBytes('\n')
        done <- response
    }()

    select {
    case data := <-done:
        return parseResponse(data)
    case <-time.After(30 * time.Second):
        c.cmd.Process.Kill()
        return nil, fmt.Errorf("plugin timeout")
    }
}
```

**Resource Limits** (platform-specific):
```go
// Unix
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true,  // New process group (kill whole tree)
}

// Windows
cmd.SysProcAttr = &syscall.SysProcAttr{
    CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
}
```

## Plugin Discovery

### Discovery Locations

```go
pluginDirs := []string{
    filepath.Join(os.UserHomeDir(), ".engineerdna", "plugins"),  // User
    "./plugins",                                                  // Bundled
}
```

**Priority**: User plugins override bundled.

### Validation Rules

```go
// Name validation (prevents path traversal)
validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
if !validName.MatchString(name) {
    return fmt.Errorf("invalid plugin name")
}

// File size validation
const maxMetadataSize = 1 * 1024 * 1024 // 1MB
if fileInfo.Size() > maxMetadataSize {
    return fmt.Errorf("plugin.json too large")
}

// Symlink validation
resolvedPath, _ := filepath.EvalSymlinks(candidate)
relPath, _ := filepath.Rel(pluginDir, resolvedPath)
if strings.HasPrefix(relPath, "..") {
    continue // Skip if outside plugin dir
}
```

## Security Isolation

### Subprocess Isolation

**Why Subprocess?**
- Memory isolation
- Resource limits
- Timeout enforcement
- No direct DB access
- No encryption keys
- Language agnostic

**How**:
```go
cmd := exec.Command(pluginPath)
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
stdin, _ := cmd.StdinPipe()
stdout, _ := cmd.StdoutPipe()
cmd.Start()
```

### Timeout Enforcement

30-second rule prevents hung plugins:

```go
select {
case data := <-responseChan:
    return parseResponse(data)
case <-time.After(30 * time.Second):
    c.cmd.Process.Kill()
    return nil, fmt.Errorf("plugin timeout (30s)")
}
```

**Plugin Best Practices**:
- Paginate large syncs
- Use shorter timeout internally (25s)
- Return partial results + warnings
- Don't wait for slow external APIs

### Resource Limits

- Response size: 100MB max
- Stderr output: 10MB max
- plugin.json: 1MB max

```go
const maxResponseSize = 100 * 1024 * 1024
limitedReader := io.LimitReader(stdout, maxResponseSize)

const maxStderrBytes = 10 * 1024 * 1024
limitedStderr := io.LimitReader(stderr, maxStderrBytes)
```

### Path Traversal Prevention

**Attack Vectors**:
1. Malicious name: `../../etc/passwd`
2. Symlink to sensitive file
3. Executable outside plugin dir

**Mitigations**:
```go
// 1. Name validation
validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
if !validName.MatchString(name) {
    return "", fmt.Errorf("invalid plugin name")
}

// 2. Symlink resolution
resolvedPath, _ := filepath.EvalSymlinks(candidate)
relPath, _ := filepath.Rel(pluginDir, resolvedPath)
if strings.HasPrefix(relPath, "..") {
    continue // Reject
}

// 3. Explicit path construction
pluginPath := filepath.Join(pluginDir, name)
```

### No Shell Injection

```go
// [BAD]
cmd := exec.Command("sh", "-c", fmt.Sprintf("./plugins/%s", pluginName))

// [GOOD]
cmd := exec.Command(pluginPath)
```

## Configuration Encryption

**Flow**: User Input → Encryption (AES-256-GCM) → Database → Decryption → Plugin

```go
// Identify secret fields, encrypt, store
secretFields := []string{}
for _, field := range info.ConfigFields {
    if field.Secret { secretFields = append(secretFields, field.Name) }
}
encryptedConfig, _ := keyStore.EncryptPluginConfig(config, secretFields)

// Decrypt before sending to plugin
decryptedConfig, _ := keyStore.DecryptPluginConfig(encryptedConfig, secretFields)
client.Configure(decryptedConfig)
```

AES-256-GCM with master key from OS keychain. Plugin never sees encryption keys.

## Anonymization Flow

### When Required

MANDATORY for processor plugins sending to external APIs:

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

1. executor.AnalyzeWithProcessor(name, events, analysisType)
2. Load plugin.json, check anonymization.required == true
3. Anonymize fields (alice@example.com → User_1)
4. Send anonymized events to plugin
5. Plugin sends anonymized data to OpenAI
6. Plugin returns insights
7. Deanonymize insights (User_1 → alice@example.com)
8. Log in audit_log with anonymized=true

```go
// Check requirement
if info.Anonymization.Required {
    // Anonymize
    anonymizedEvents, _ := anonService.AnonymizeEvents(events, ...)

    // Send to plugin
    result, _ := client.Analyze(analysisType, anonymizedEvents, context)

    // Deanonymize
    deanonymizedInsights, _ := anonService.DeanonymizeInsights(result.Insights)

    // Audit
    auditStore.Log(&models.AuditLog{
        PluginName:  pluginName,
        Action:      "processor_analysis",
        EventCount:  len(events),
        Anonymized:  true,
    })

    return deanonymizedInsights
}
```

### Mapping Storage

```sql
CREATE TABLE anonymization_map (
    anonymized_id TEXT UNIQUE,  -- "User_1"
    real_name TEXT,              -- "alice@example.com"
    enabled BOOLEAN DEFAULT 1
);
```

## Testing & Debugging

### Unit Tests

```go
func TestLoaderDiscoverPlugins(t *testing.T) {
    tmpDir := t.TempDir()
    pluginDir := filepath.Join(tmpDir, "test-plugin")
    os.Mkdir(pluginDir, 0755)
    os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(`{"name":"test","version":"1.0.0","type":"source"}`), 0644)

    loader := NewLoader([]string{tmpDir})
    plugins, _ := loader.DiscoverPlugins()
    if len(plugins) != 1 {
        t.Errorf("Expected 1 plugin, got %d", len(plugins))
    }
}
```

### Plugin Logging

```go
// In plugin
fmt.Fprintf(os.Stderr, "Syncing events since %s\n", params.Since)

// View
./bin/engineerdna 2>&1 | grep "plugin:"
```

### Manual Testing

```bash
echo '{"jsonrpc":"2.0","method":"plugin.info","id":"1"}' | ./github | jq .
```

### Common Issues

- **Plugin not found**: Check binary exists, run `go build -o github`
- **Plugin timeout**: Reduce work per sync, add timeouts to API calls
- **Decryption fails**: Reconfigure plugin (will re-encrypt)
- **Anonymization not applied**: Check `plugin.json` has `anonymization.required: true`

## Performance Optimization

```go
// Plugin pooling (future): Reuse plugin processes
// Batch operations: Sync plugins concurrently
var wg sync.WaitGroup
for _, plugin := range plugins {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        events, _ := executor.SyncSourcePlugin(name, since)
    }(plugin.Name)
}
wg.Wait()
```

## Event Type Registry (PDR-9)

**Purpose**: Normalize events from different sources to common types for cross-source analysis.

**Example**: GitHub PR + GitLab MR → `code_review`

```go
type EventRegistry struct {
    eventTypes map[string]EventTypeSpec  // Loaded from plugin manifests
}

// Plugin declares event types in plugin.json
{
  "provides_event_types": [
    {
      "type": "pull_request",
      "normalized_type": "code_review",
      "normalization_map": {
        "title": "title",
        "author": "user.login",
        "state": "state",
        "created_at": "created_at"
      }
    }
  ]
}

// On event creation, registry normalizes
event := Event{
    Type: "pull_request",              // Source type
    NormalizedType: "code_review",     // Normalized type
    Data: {...},                       // Source data
    NormalizedData: "{...}",           // Normalized fields as JSON
}
```

**Benefits**:
- Query GitHub + GitLab PRs as `code_review`
- Widgets work across multiple sources
- Analytics don't care about source

**Implementation**: `internal/plugin/event_registry.go`

## Widget Registry (PDR-9)

**Purpose**: Allow plugins to register custom dashboard widgets.

**Example**: AWS Cost plugin registers cost breakdown widget

```go
type WidgetRegistry struct {
    widgets map[string]WidgetSpec  // Loaded from plugin manifests
}

// Plugin declares widgets in plugin.json
{
  "provides_widgets": [
    {
      "type": "aws_cost_breakdown",
      "name": "AWS Cost Breakdown",
      "description": "Shows costs by service and region"
    }
  ]
}

// Frontend queries registry
GET /api/widgets/registry
{
  "widgets": [
    {"type": "aws_cost_breakdown", "plugin": "aws-cost", ...}
  ]
}
```

**Widget Lifecycle**:
1. Plugin declares widget in `provides_widgets`
2. Host stores in `plugin_manifests` table
3. Widget registry exposes via API
4. Frontend fetches widget list
5. Frontend renders widget (if component implemented)

**Implementation**: `internal/plugin/widget_registry.go`

## Metric Calculation Engine (PDR-9)

**Purpose**: Compute metrics from events using SQL aggregations.

**Example**: Calculate deployment frequency from deploy events

```go
type MetricEngine struct {
    eventStore *db.EventStore
}

// Define metric calculation
metricDef := MetricDefinition{
    Name: "deploy_frequency",
    Query: `
        SELECT COUNT(*) as value
        FROM events
        WHERE normalized_type = 'deployment'
        AND timestamp >= ?
        AND timestamp < ?
    `,
}

// Calculate metric for time period
value, _ := engine.Calculate("deploy_frequency", startTime, endTime)
```

**Security**: Parameterized queries only, no dynamic SQL construction.

**Implementation**: `internal/services/metric_engine.go`

## Correlation Engine (PDR-9)

**Purpose**: Link events, metrics, and attributes for cross-data-type analysis.

**Example**: Cost per feature (links AWS cost metrics + feature work events)

```go
type CorrelationEngine struct {
    metricStore    *db.MetricStore
    eventStore     *db.EventStore
    attributeStore *db.AttributeStore
}

// Define correlation
correlation := Correlation{
    Name: "cost_per_feature",
    Definition: {
        "metric": "aws_cost",
        "events": "feature_work",
        "calculation": "total_cost / feature_count"
    },
}

// Compute correlation for time window
result, _ := engine.Compute("cost_per_feature", startTime, endTime)
// Returns: CorrelationValue{Value: 1250.50, Breakdown: {"feature_A": 500, "feature_B": 750}}
```

**Implementation**: `internal/services/correlation_engine.go`

## References

- Rule 8: Plugin isolation (subprocess, timeout, no DB)
- Rule 9: Anonymization required for external APIs
- Rule 19: Encrypt secrets at rest
- plugins/CLAUDE.md: Plugin development guide
- plugins/plugin-sdk/: SDK for building plugins
- Skill: `plugin-development` - Plugin workflows
- Agent: `integration-checker` - Plugin system testing
