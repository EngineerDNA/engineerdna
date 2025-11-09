# Writing Effective Descriptions

The `description` field in agent frontmatter serves two critical purposes:

1. **Human understanding** - Explains when and why to use the agent
2. **Auto-delegation** - Contains trigger phrases that enable Claude Code to automatically invoke the agent

## Auto-Delegation Trigger Phrases

Include explicit phrases that signal when the agent should be used:

### Strong Phrases (Recommended)

- **"MUST BE USED"** - Forces manual consideration
  - Example: "MUST BE USED for code/architecture review"

- **"Use proactively"** - Suggests automatic invocation
  - Example: "Use proactively when encountering test failures"

- **"Proactively reviews"** - Agent actively searches for issues
  - Example: "Proactively reviews code for security vulnerabilities"

### Specific Keywords

List concrete trigger situations:

```yaml
description: Code review specialist. MUST BE USED for review, audit, code check, ready for production, LGTM, sanity check, verify changes.
```

### Contextual Keywords

Include the types of tasks the agent handles:

```yaml
description: Debugging expert. Use proactively for errors, test failures, stack traces, unexpected behavior, debugging, troubleshooting, root cause analysis.
```

## Examples: GOOD vs BAD

### Good: Code Reviewer

```yaml
name: code-reviewer
description: Expert code review specialist. MUST BE USED for code/architecture review - review, check, audit, assess, ready for prod, LGTM, sanity check, verify.
allowed-tools: Read, Grep, Glob, Bash
```

**Why it works:**
- Clear purpose and expertise
- Multiple trigger keywords
- Specific tool scope
- Action-oriented ("review", "check", "audit")

### Bad: Code Reviewer

```yaml
name: code-reviewer
description: Reviews code
```

**Why it fails:**
- Too vague
- No trigger phrases
- Unclear when to invoke
- Won't trigger auto-delegation

### Good: Debugger

```yaml
name: debugger
description: Debugging specialist for errors, test failures, unexpected behavior. Use proactively when encountering issues.
allowed-tools: Read, Edit, Bash, Grep, Glob
```

**Why it works:**
- Lists specific problem types (errors, test failures)
- Clear when to invoke (encountering issues)
- Permissions match needs (includes Edit for fixes)

### Bad: Debugger

```yaml
name: debugger
description: Helps with debugging
```

**Why it fails:**
- Passive ("helps with")
- Vague trigger conditions
- No sense of urgency or proactivity

## Writing Strategy

### 1. Start with Purpose

```
[Agent] specialist in [domain]. [Auto-delegation signal] for [trigger keywords].
```

Example: "Security specialist. MUST BE USED for security review, vulnerability assessment, authentication checks."

### 2. Add Specific Situations

List the exact contexts where the agent should be invoked:

```yaml
description: Database migration expert. MUST BE USED for schema changes, migrations, new tables, column additions, rollbacks, data transformations.
```

### 3. Include Behavior Cues

Describe what the agent actively does:

```yaml
description: Proactively checks for unused code, dead imports, abandoned components. Use when optimizing bundle size or auditing codebase.
```

### 4. Combine Multiple Signals

Use both strong phrases and specific keywords:

```yaml
description: Documentation reviewer. MUST BE USED for documentation review, CLAUDE.md updates, missing docs, API docs, README updates.
allowed-tools: Read, Bash, Grep, Glob
```

## Testing Your Description

Ask these questions:

1. **Is the purpose crystal clear?** - Would someone unfamiliar with the agent understand what it does?
2. **Are trigger phrases present?** - Would Claude Code recognize when to invoke it?
3. **Are situations specific?** - Do the keywords match real use cases in your projects?
4. **Would you want auto-invocation?** - Does the purpose warrant automatic delegation?

If you can't answer "yes" to all four, revise the description.
