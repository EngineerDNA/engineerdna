---
name: using-mcp-tools
description: Use MCP tools—context7 for library docs, sequential-thinking for reasoning, chrome-mcp-server for browser automation.
allowed-tools: mcp__context7__*, mcp__sequential-thinking__*, mcp__chrome-mcp-server__*, Read, WebSearch
---

# Using MCP Tools

## Overview

Model Context Protocol (MCP) tools extend Claude with specialized capabilities:
- **context7**: Look up up-to-date library and framework documentation
- **sequential-thinking**: Structured reasoning for complex multi-step problems
- **chrome-mcp-server**: Browser automation for UI testing and authenticated API calls

## Quick Start

### Library Documentation Lookup
```typescript
// 1. Resolve library ID
mcp__context7__resolve-library-id({ libraryName: "drizzle-orm" })

// 2. Get documentation
mcp__context7__get-library-docs({ context7CompatibleLibraryID: "/org/project" })
```
[context7 Reference](references/context7-reference.md)

### Complex Reasoning
Use when facing architecture decisions, multi-step debugging, or planning with unknowns.
[sequential-thinking Reference](references/sequential-thinking-reference.md)

### Browser Testing
Navigate → interact → verify. Always start with small parameters (maxMessages: 10).
[chrome-mcp-server Reference](references/chrome-mcp-reference.md)

## Key Patterns

- **Resolve library ID first** before getting context7 docs
- **Adjust thinking count** as your understanding evolves
- **Start small, adapt** to token limits when hitting errors
- **Never give up**—reduce parameters and retry with different filters

## References

- [context7: Library Documentation](references/context7-reference.md)
- [sequential-thinking: Complex Reasoning](references/sequential-thinking-reference.md)
- [chrome-mcp-server: Browser Automation](references/chrome-mcp-reference.md)
- [Agent Examples by Role](references/agent-examples.md)
