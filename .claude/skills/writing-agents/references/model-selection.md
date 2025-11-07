# Model Selection

The `model` field in agent frontmatter determines which Claude model the agent uses. Default is `sonnet`.

## Available Models

### Sonnet (Recommended Default)
```yaml
model: sonnet
```

**Characteristics:**
- Balanced performance and intelligence
- Good for most tasks
- Reasonable latency
- Cost-effective

**Best for:** Most agents - reviewers, debuggers, analysts, content creators

**Example:**
```yaml
---
name: code-reviewer
model: sonnet
---
```

### Opus (Most Capable)
```yaml
model: opus
```

**Characteristics:**
- Highest intelligence and reasoning
- Longer context window (if available)
- Slower and more expensive
- Best for complex analysis

**Best for:** Architecture decisions, complex debugging, strategic planning

**Example:**
```yaml
---
name: architect
model: opus
---
```

### Haiku (Fast & Economical)
```yaml
model: haiku
```

**Characteristics:**
- Fastest inference
- Cheapest option
- Lower capability
- Good for simple tasks

**Best for:** Quick checks, simple searches, categorization tasks

**Example:**
```yaml
---
name: quick-checker
model: haiku
---
```

### Inherit (Match Main Conversation)
```yaml
model: inherit
```

**Characteristics:**
- Uses the same model as the main conversation
- Ensures consistency
- Useful for collaborative agents
- Agent adapts to user's model choice

**Best for:** General-purpose agents that should work with any model

**Example:**
```yaml
---
name: helper-agent
model: inherit
---
```

## Decision Framework

Choose your agent's model based on:

### 1. Task Complexity

**Simple tasks** (quick checks, finding info):
```yaml
model: haiku
```

**Medium complexity** (code review, testing):
```yaml
model: sonnet
```

**Complex reasoning** (architecture, strategic decisions):
```yaml
model: opus
```

### 2. Latency Sensitivity

**Must be fast** (real-time feedback):
```yaml
model: haiku
```

**Moderate latency OK** (most tasks):
```yaml
model: sonnet
```

**Latency acceptable** (complex analysis):
```yaml
model: opus
```

### 3. Budget Constraints

**Cost is important**:
```yaml
model: haiku  # Cheapest
```

**Balanced approach**:
```yaml
model: sonnet  # Best value
```

**Cost is secondary to quality**:
```yaml
model: opus  # Most capable
```

### 4. Agent Purpose

| Agent Type | Recommended | Rationale |
|-----------|-------------|-----------|
| Code Reviewer | `sonnet` | Good balance of analysis depth and speed |
| Debugger | `sonnet` or `opus` | Opus for complex issues, sonnet for common problems |
| Architect | `opus` | Complex reasoning needed |
| Content Generator | `sonnet` | Good writing quality, reasonable speed |
| Quick Checker | `haiku` | Speed matters more than depth |
| Helper/Assistant | `inherit` | Adapts to user preference |
| Security Auditor | `opus` | High stakes, complex reasoning |
| Document Scanner | `haiku` | Speed matters, simple search task |

## Model Combinations in Practice

### Lightweight Agent Suite
```yaml
# Fast feedback agents
name: quick-linter
model: haiku

name: grep-helper
model: haiku

name: test-runner
model: sonnet
```

### Heavyweight Agent Suite
```yaml
# For complex projects requiring deep analysis
name: architect
model: opus

name: security-auditor
model: opus

name: code-reviewer
model: sonnet
```

### Balanced Agent Suite
```yaml
# Mix matched to actual needs
name: code-reviewer
model: sonnet

name: debugger
model: sonnet

name: architect
model: opus

name: quick-checker
model: haiku
```

## Model Performance Comparison

| Aspect | Haiku | Sonnet | Opus |
|--------|-------|--------|------|
| Speed | Fast | Medium | Slower |
| Intelligence | Good | Excellent | Outstanding |
| Cost | Lowest | Medium | Higher |
| Context | Standard | Standard | Large |
| Reasoning | Good | Strong | Advanced |
| Code Quality | Good | Excellent | Excellent |

## Best Practices

### 1. Start with Sonnet
```yaml
model: sonnet  # Good default, good value
```

Only move to Opus or Haiku if you have specific requirements.

### 2. Use Inherit for Flexibility
```yaml
model: inherit  # Agent adapts to user's choice
```

Good for multi-purpose agents and helpers that should work at any intelligence level.

### 3. Match Task to Model
```yaml
# Simple task = Haiku
name: file-counter
model: haiku

# Complex task = Opus
name: architecture-advisor
model: opus
```

### 4. Monitor Agent Performance
If an agent seems:
- **Too slow?** Consider downgrading to Haiku
- **Not smart enough?** Upgrade to Opus
- **Too expensive?** Downgrade to Haiku or Sonnet

### 5. Cost Optimization
For teams managing costs:

```yaml
# Expensive agents (use Sonnet/Haiku)
name: code-reviewer
model: sonnet

name: quick-search
model: haiku

# High-value agents (can use Opus)
name: architect
model: opus
```

## No Model Specified (Defaults)

If you omit the `model` field:

```yaml
---
name: agent-name
description: Something
---
```

The agent will use `sonnet` (Anthropic's default).

## Changing Models Over Time

As your team's needs evolve:

1. **Start conservative** - Use Sonnet for everything
2. **Identify pain points** - Which agents feel too slow or not smart enough?
3. **Optimize gradually** - Update specific agents where you see ROI
4. **Monitor results** - Track whether changes improved the workflow

Example evolution:
```yaml
# Initial: All Sonnet
model: sonnet

# After feedback: Opus for architecture decisions
name: architect
model: opus

# After feedback: Haiku for quick checks
name: linter
model: haiku
```

## Edge Cases

### Agent that Must Match User's Model
```yaml
model: inherit
```

Use when agent should provide consistent experience with main conversation.

### Agent that Explicitly Needs Upgrading
```yaml
model: opus  # This agent always needs maximum capability
```

Use when the task is inherently complex and Sonnet falls short.

### Agent that Should Be Fast Regardless
```yaml
model: haiku  # Speed is critical even if slightly less capable
```

Use when latency matters more than capability (e.g., real-time feedback).
