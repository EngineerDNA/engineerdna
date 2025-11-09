# Agent File Format

## Directory Structure

Agent files can exist at two levels with priority ordering:

```
.claude/agents/
  agent-name.md (project-level, highest priority)

~/.claude/agents/
  agent-name.md (user-level, lower priority)
```

Project-level agents override user-level agents with the same name.

## File Format

All agent files are Markdown with YAML frontmatter:

```markdown
---
name: agent-name
description: When this agent should be invoked
allowed-tools: tool1, tool2  # Optional
model: sonnet  # Optional
forkedContext: true  # Optional - Claude Code specific
isAsync: false  # Optional - Claude Code specific
---

Your agent's system prompt goes here.
```

## Frontmatter Fields

### Required Fields

- **`name`** (string): Agent identifier in lowercase with hyphens only
  - Pattern: `[a-z0-9\-]+`
  - Examples: `code-reviewer`, `debugger`, `db-expert`

- **`description`** (string): Natural language purpose describing when to invoke the agent
  - Should include trigger phrases like "MUST BE USED", "proactively", or specific keywords
  - Example: "Expert code reviewer. MUST BE USED for reviews, audits, security checks."

### Optional Fields

- **`allowed-tools`** (string): Comma-separated list of tools the agent can access
  - If omitted, agent inherits all available tools
  - Examples: `Read, Grep, Bash, Edit`
  - Common tools: Read, Edit, Bash, Grep, Glob, Write

- **`model`** (string): Model alias or `'inherit'`
  - Options: `sonnet`, `opus`, `haiku`, or `inherit`
  - Default: `sonnet` (if omitted)
  - Use `inherit` to match the main conversation's model

- **`forkedContext`** (boolean): Claude Code-specific - creates isolated context
  - Default: `true`

- **`isAsync`** (boolean): Claude Code-specific - runs asynchronously
  - Default: `false`

## System Prompt Content

The content after frontmatter is the agent's system prompt. Best structure:

1. **Role definition** - Start with "You are a..."
2. **When invoked** - Step-by-step instructions for the agent's workflow
3. **Guidelines** - Specific rules and constraints
4. **Checklists** - Items to verify or check
5. **Output format** - How the agent should structure its response

Example structure:

```markdown
You are a senior database architect specializing in data modeling.

When invoked:
1. Examine the current schema
2. Identify optimization opportunities
3. Propose changes with rationale

Guidelines:
- Prefer denormalization for read-heavy patterns
- Document all design decisions
- Consider performance implications

Checklist:
- [ ] Indexes on all foreign keys
- [ ] Appropriate column types
- [ ] RLS policies defined
```
