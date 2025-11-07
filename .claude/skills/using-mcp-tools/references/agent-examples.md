# Agent-Specific MCP Usage Examples

## Overview

Different agents have access to different MCP tools based on their role. This file provides agent-specific usage examples.

## Advisor Agent

**Has access to**: sequential-thinking

**Usage**: Use for complex review decisions requiring trade-off analysis

```typescript
// When reviewing architecture
sequentialthinking({
  thought: "Reviewing proposed RLS implementation. Need to verify: security, performance, maintainability.",
  thoughtNumber: 1,
  totalThoughts: 4,
  nextThoughtNeeded: true
})
```

## Product-Manager Agent

**Has access to**: sequential-thinking, WebSearch, ListMcpResourcesTool

**Usage**: Use for requirements analysis and user research

```typescript
// When gathering requirements
sequentialthinking({
  thought: "User wants 'pause feature'. Need to understand: what states are pausable, what happens to resources, how to resume.",
  thoughtNumber: 1,
  totalThoughts: 5,
  nextThoughtNeeded: true
})
```

## Architect Agent

**Has access to**: sequential-thinking, context7, WebSearch

**Usage**: Use for technical design and library evaluation

```typescript
// Before choosing library
const libraryId = resolve-library-id({ name: "temporal-sdk" });
const docs = get-library-docs({ libraryId, query: "workflow pause and resume" });

// Then design with correct API
sequentialthinking({
  thought: "Temporal supports workflow cancellation with checkpoints. Design: Cancel workflow, save state, resume from checkpoint.",
  thoughtNumber: 1,
  totalThoughts: 3,
  nextThoughtNeeded: true
})
```

## Engineer Agent

**Has access to**: context7, WebSearch

**Usage**: Use for implementation guidance and API verification

```typescript
// Before implementing
const libraryId = resolve-library-id({ name: "drizzle-orm" });
const docs = get-library-docs({ libraryId, query: "timestamp with timezone" });

// Implement with correct API
// Create schema with timestamp('created_at', { mode: 'string', withTimezone: true })
```

## Designer Agent

**Has access to**: sequential-thinking, chrome-mcp-server, WebSearch

**Usage**: Use for UI research and competitor analysis

```typescript
// Research existing patterns
chrome_navigate({ url: "https://github.com/pulls" })
chrome_screenshot({ fullPage: true, storeBase64: true })

// Analyze UI patterns
sequentialthinking({
  thought: "GitHub uses tabs for filtering. Should we use tabs or dropdown for change request filters?",
  thoughtNumber: 1,
  totalThoughts: 3,
  nextThoughtNeeded: true
})
```

## UI-Tester Agent

**Has access to**: chrome-mcp-server, Bash, Read

**Usage**: Primary user of chrome-mcp-server for testing

```typescript
// Complete UI test flow
chrome_navigate({ url: "http://localhost:3000/dashboard/test/pause-button" })
chrome_screenshot({ savePng: false, storeBase64: true })
chrome_get_interactive_elements({ textQuery: "Pause" })
chrome_click_element({ selector: "button[data-testid='pause']" })
chrome_console({ includeExceptions: true })
```

## Integration-Checker Agent

**Has access to**: chrome-mcp-server, WebFetch, Bash, Read, Grep, Glob

**Usage**: Verify end-to-end connectivity

```typescript
// Check API endpoint connectivity
chrome_navigate({ url: "http://localhost:3000/dashboard" })
chrome_inject_script({
  type: "MAIN",
  jsScript: `
    fetch('/api/change-requests/123/pause', {
      method: 'POST',
      credentials: 'include'
    })
    .then(r => console.log('Status:', r.status))
    .catch(err => console.error('Failed:', err));
  `
})
chrome_console()
```

## Examples by Scenario

### Scenario: Adding New Library

```typescript
// Architect uses context7 + sequential-thinking
// 1. Look up documentation
const libId = resolve-library-id({ name: "zustand" });
const docs = get-library-docs({ libId, query: "store creation" });

// 2. Reason through integration
sequentialthinking({
  thought: "Zustand is lightweight state management. Fits our needs for pause button state. Alternative: React Context, but Zustand simpler.",
  nextThoughtNeeded: true,
  thoughtNumber: 1,
  totalThoughts: 3
})

// Engineer then implements with correct API
```

### Scenario: Debugging UI Issue

```typescript
// UI-Tester or Integration-Checker uses chrome-mcp-server
// 1. Navigate to problem page
chrome_navigate({ url: "http://localhost:3000/dashboard" });

// 2. Capture state
chrome_screenshot({ fullPage: true, storeBase64: true });

// 3. Check console
chrome_console({ includeExceptions: true });

// Advisor uses sequential-thinking to analyze
// 4. Reason through issue
sequentialthinking({
  thought: "Console shows 'Cannot read property of undefined'. Screenshot shows blank component. Hypothesis: Data fetch failing.",
  nextThoughtNeeded: true,
  thoughtNumber: 1,
  totalThoughts: 4
})

// Integration-Checker verifies with network debugger
// 5. Verify hypothesis
chrome_network_debugger_start();
chrome_navigate({ url: "http://localhost:3000/dashboard", refresh: true });
const network = chrome_network_debugger_stop();
```

### Scenario: Architecture Decision

```typescript
// Architect uses context7 + sequential-thinking
// 1. Research options
const temporalDocs = get-library-docs({
  libraryId: resolve-library-id({ name: "@temporalio/client" }),
  query: "workflow pause"
});

const bullDocs = get-library-docs({
  libraryId: resolve-library-id({ name: "bull" }),
  query: "job pause"
});

// 2. Analyze trade-offs
sequentialthinking({
  thought: "Need to pause long-running workflows. Two approaches: Temporal (current system) vs Bull (job queue).",
  nextThoughtNeeded: true,
  thoughtNumber: 1,
  totalThoughts: 6
})

sequentialthinking({
  thought: "Temporal: Already integrated, supports checkpoints, harder to pause mid-workflow. Bull: Easy pause, need migration, simpler mental model.",
  nextThoughtNeeded: true,
  thoughtNumber: 2,
  totalThoughts: 6
})

// Continue until decision made
```
