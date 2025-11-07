# Agent Authoring Best Practices

## Core Principle: One Task = One Engineer

**NEVER pass todo lists to engineer agents.**

```
# Bad - loses focus, pollutes context
"Engineer: Implement these 10 tasks: [long list]"

# Good - one focused task
"Engineer: Create the UserRepository with findById method"
[After completion]
"Engineer: Add findByEmail method to UserRepository"
```

## Agent Pipeline

### Forward Flow
```
engineer → quality → security → advisor
```

### Backward Flow (on errors)
```
quality fails → back to engineer (with failing tests)
security issues → back to engineer (with specific vulnerabilities)
advisor blocks → back to engineer (with specific fixes)
```

**Don't just report errors - go back and fix them.**

For larger projects, you may want additional agents like:
- product-manager (requirements gathering)
- designer (UI/UX design)
- architect (technical design)
- integration-checker (connectivity verification)
- ui-tester (browser UI testing)

## Context Accumulation

Agents are stateless. Pass complete context each time:

```
For engineer agent:
- USER GOAL: One-liner from PM (not full requirements)
- ARCHITECT SPEC: Data model, API contracts, repository methods, patterns
- YOUR TASK: The ONE specific implementation (e.g., "Create UserRepository")
```

## Agent File Structure

```markdown
---
name: agent-name
description: When this agent should be invoked. Be specific with triggers.
tools: tool1, tool2, tool3  # Optional - inherits all if omitted
model: sonnet  # Optional - specify or 'inherit'
---

Your agent's system prompt here.

Define role, capabilities, approach clearly.
Include specific instructions and constraints.
```

### Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Lowercase with hyphens |
| `description` | Yes | Natural language, include triggers |
| `tools` | No | Comma-separated. Inherits all if omitted |
| `model` | No | `sonnet`, `opus`, `haiku`, or `'inherit'` |

## Agent Design Principles

### Focused Purpose

Each agent has ONE clear responsibility:

**Good Agent** (focused):
```markdown
---
name: test-runner
description: Run tests and fix failures. Use proactively after code changes.
tools: Read, Edit, Bash
---

You run tests and fix failures while preserving test intent.

When invoked:
1. Run tests
2. If failures, analyze
3. Fix issues
4. Verify fixes
```

**Bad Agent** (too broad):
```markdown
---
name: helper
description: Helps with various tasks
---

You help with whatever is needed.
```

### Tool Permissions

Only grant necessary tools:

```markdown
# Test runner doesn't need Write
tools: Read, Edit, Bash

# Code reviewer doesn't need Edit
tools: Read, Grep, Glob, Bash

# Debugger needs everything
tools: Read, Edit, Bash, Grep, Glob
```

### Prompt Structure

```markdown
You are [specific role].

When invoked:
1. [Concrete first step]
2. [Next step]
3. [Final step]

Key practices:
- [Specific practice]
- [Another practice]

For each task:
- [Output requirement]
- [Another requirement]
```

## Model Selection

```markdown
# Use specific model
model: haiku

# Use fast model alias
model: sonnet

# Inherit from main conversation
model: 'inherit'

# Omit to use default subagent model
# (no model field)
```

## Automatic Delegation

Claude delegates based on:
- Task description in request
- `description` field in agent config
- Current context
- Available tools

**To encourage proactive use**, include in description:
- "use PROACTIVELY"
- "MUST BE USED for"
- "Use immediately after"

## Explicit Invocation

```
> Use the test-runner agent to fix failing tests
> Have the code-reviewer agent look at my recent changes
> Ask the debugger agent to investigate this error
```

## Agent Locations

**Project agents**: `.claude/agents/agent-name.md`
- Available in current project
- Shared with team

**User agents**: `~/.claude/agents/agent-name.md`
- Available across all projects
- Personal, not committed

**Priority**: Project agents override user agents with same name.

## Example Agents

### Code Reviewer
```markdown
---
name: code-reviewer
description: Expert code review specialist. Proactively reviews code for quality, security, and maintainability. Use immediately after writing or modifying code.
tools: Read, Grep, Glob, Bash
model: inherit
---

You are a senior code reviewer ensuring high standards.

When invoked:
1. Run git diff to see recent changes
2. Focus on modified files
3. Begin review immediately

Review checklist:
- Code is simple and readable
- Functions and variables are well-named
- No duplicated code
- Proper error handling
- No exposed secrets
- Input validation
- Good test coverage
- Performance considerations

Provide feedback by priority:
- Critical (must fix)
- Warnings (should fix)
- Suggestions (consider)

Include specific examples of how to fix issues.
```

### Debugger
```markdown
---
name: debugger
description: Debugging specialist for errors, test failures, and unexpected behavior. Use proactively when encountering any issues.
tools: Read, Edit, Bash, Grep, Glob
---

You are an expert debugger specializing in root cause analysis.

When invoked:
1. Capture error message and stack trace
2. Identify reproduction steps
3. Isolate failure location
4. Implement minimal fix
5. Verify solution works

Debugging process:
- Analyze error messages and logs
- Check recent code changes
- Form and test hypotheses
- Add strategic debug logging
- Inspect variable states

For each issue, provide:
- Root cause explanation
- Evidence supporting diagnosis
- Specific code fix
- Testing approach
- Prevention recommendations

Focus on fixing the underlying issue, not symptoms.
```

## Managing Agents

### Using /agents Command
```
/agents
```

Opens interface to:
- View all available agents
- Create new agents
- Edit existing agents
- Delete custom agents
- See which agent is active for duplicates
- Manage tool permissions easily

### Direct File Management
```bash
# Create project agent
mkdir -p .claude/agents
echo '---
name: my-agent
description: What it does and when to use it
---

System prompt here.' > .claude/agents/my-agent.md
```

## Testing Agents

1. **Create agent file**
2. **Test explicit invocation**: "Use the my-agent agent to..."
3. **Test automatic delegation**: Give it a task matching the description
4. **Verify tools work**: Ensure only granted tools are used
5. **Check model**: Verify correct model is used

## Chaining Agents

For complex workflows:
```
> First use the code-analyzer agent to find performance issues, then use the optimizer agent to fix them
```

Claude orchestrates the chain automatically.

## Project Agents

See `.claude/agents/` for current agents:
- `engineer` - Full-stack implementation (Go + React)
- `quality` - Testing and verification
- `security` - Security review
- `advisor` - Code review

Follow these patterns for consistency with existing agents.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Passing todo lists | One task per invocation |
| Too broad purpose | Single, focused responsibility |
| Too many tools | Only grant what's needed |
| Vague description | Include specific triggers |
| Generic prompt | Concrete steps and checklists |
| No validation | Add verification steps |
