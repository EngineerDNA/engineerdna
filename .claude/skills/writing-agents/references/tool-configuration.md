# Tool Configuration

The `allowed-tools` field controls which tools an agent can access. Proper tool configuration improves security, focus, and performance.

## Strategy Overview

**Two approaches:**

1. **Inherit all tools** (recommended for most agents) - Simpler, more powerful
2. **Restrict to specific tools** (for focused agents) - Better security and focus

## Strategy 1: Inherit All Tools (Recommended)

**When:** The agent needs flexible tool access for its domain

**How:** Omit the `allowed-tools` field entirely

```yaml
---
name: full-access-agent
description: Senior architect with full system access
---
```

**Benefits:**
- Agent has maximum capability
- Can use appropriate tools as situations arise
- Less maintenance (no tool list to update)
- Better for complex, multi-faceted tasks

**Best for:**
- Senior/lead roles (architect, advisor)
- Complex analysis requiring many tools
- Meta-agents that help with agent creation
- General-purpose problem solving

**Example:** An architect reviewing system design needs Read, Bash, Grep, Edit, Write, and potentially more tools.

## Strategy 2: Restrict to Specific Tools

**When:** The agent should focus on specific tasks with limited scope

**How:** List exact tools as comma-separated string

```yaml
---
name: code-reviewer
description: Expert code review specialist
allowed-tools: Read, Grep, Glob, Bash
---
```

**Benefits:**
- Clear scope and focus
- Prevents agent from doing unintended actions
- Better security (can't edit if not allowed)
- Faster execution (fewer tool options to consider)

**Common Tool Combinations:**

### Read-Only Analysis
For agents that only analyze code:
```yaml
allowed-tools: Read, Grep, Glob
```

Used by: Code reviewers, dependency checkers, static analysis agents

### Read + Search + Execute
For agents that analyze and run commands:
```yaml
allowed-tools: Read, Grep, Glob, Bash
```

Used by: Debuggers, test runners, performance analysts

### Read + Modify + Execute
For agents that can make code changes:
```yaml
allowed-tools: Read, Edit, Bash, Grep, Glob
```

Used by: Refactoring agents, bug fixers, code formatters

### Write + Modify
For agents that create and modify files:
```yaml
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
```

Used by: Code generators, scaffold creators, documentation writers

### Search Only
For agents that just search and report:
```yaml
allowed-tools: Grep, Glob
```

Used by: Code detectors, pattern finders, auditors

## Available Tools Reference

Common Claude Code tools:

| Tool | Purpose | Use Case |
|------|---------|----------|
| `Read` | Read file contents | Understanding code, reading configs |
| `Edit` | Modify file contents | Fixing bugs, refactoring |
| `Write` | Create/overwrite files | Generating new files |
| `Bash` | Execute commands | Running scripts, builds, tests |
| `Grep` | Search file contents | Finding patterns, code search |
| `Glob` | Find files by pattern | Locating files by name |
| `Bash` | Run shell commands | Executing scripts and tools |
| `NotebookEdit` | Modify Jupyter notebooks | Working with notebooks |

## Decision Framework

Ask these questions to choose the right tools:

### 1. What does the agent need to READ?
- Just code? → `Read, Grep, Glob`
- Code and configs? → `Read, Grep, Glob`
- Everything including logs? → Inherit all

### 2. Can the agent MODIFY code?
- No, analysis only? → Remove `Edit` and `Write`
- Yes, fix bugs? → Include `Edit`
- Yes, create files? → Include `Write`

### 3. Does the agent need to RUN commands?
- No, static analysis only? → Remove `Bash`
- Yes, run tests? → Include `Bash`
- Yes, build/deploy? → Include `Bash`

### 4. Does the agent need other tools?
- Task management? → Include `TodoWrite`
- Browser interaction? → Include browser tools
- MCP tools? → Include as needed

## Examples by Agent Type

### Code Reviewer (Read-Only)
```yaml
allowed-tools: Read, Grep, Glob, Bash
```
Why: Reads files, searches for issues, runs git commands to see changes

### Bug Fixer (Can Edit)
```yaml
allowed-tools: Read, Edit, Bash, Grep, Glob
```
Why: Reads code, edits to fix, runs tests to verify

### Documentation Generator (Can Write)
```yaml
allowed-tools: Read, Write, Bash, Grep, Glob
```
Why: Reads code for examples, writes docs, runs commands for API info

### Security Auditor (Read-Only)
```yaml
allowed-tools: Read, Grep, Glob, Bash
```
Why: Deep analysis without modification capability

### Database Administrator (Limited Bash)
```yaml
allowed-tools: Read, Bash, Grep
```
Why: Runs DB commands but shouldn't modify application code

## Tool Permission Patterns

### Pattern 1: Safe Exploration
Developers learning the codebase:
```yaml
allowed-tools: Read, Grep, Glob, Bash
```

### Pattern 2: Safe Fixing
Fixing bugs with verification:
```yaml
allowed-tools: Read, Edit, Bash, Grep, Glob
```

### Pattern 3: Safe Generation
Creating new files safely:
```yaml
allowed-tools: Read, Write, Bash, Grep, Glob
```

### Pattern 4: Unsafe (Avoid)
Agents with Write + Bash + Edit can modify anything:
```yaml
# Not recommended - too much power
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
```

## Best Practices

1. **Start restrictive** - Begin with minimal tools, add as needed
2. **Match the task** - Agent should have exactly what it needs
3. **Security first** - Don't grant Write if Edit is sufficient
4. **Document reasoning** - Add comments if tool choice is non-obvious
5. **Test tool access** - Verify agent can perform intended tasks

## Tool Evolution

When a tool list seems wrong:

1. **Too restricted?** - Agent says "I can't access X tool"
   - Add the tool and test
   - Document why it's needed

2. **Too permissive?** - Agent does something unexpected
   - Remove unnecessary tools
   - Update system prompt to prevent misuse

3. **Unbalanced?** - Tool list doesn't match agent's purpose
   - Review agent description and purpose
   - Update tools to match scope
