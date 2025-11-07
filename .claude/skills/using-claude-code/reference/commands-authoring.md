# Command Authoring Best Practices

## When to Use Commands vs Skills

**Use Command** if:
- Simple repeated prompt
- Fits in single file
- Want explicit invocation (`/review`)
- Quick reminder or template

**Use Skill** if:
- Multiple files/scripts needed
- Complex workflow with validation
- Should auto-discover from context
- Team needs detailed guidance

## Basic Structure

```markdown
---
description: Brief description of what this command does
---

Command instructions here.
Use placeholders like $ARGUMENTS for dynamic values.
```

## Argument Handling

### All arguments with $ARGUMENTS
```markdown
---
description: Fix issue following coding standards
---

Fix issue #$ARGUMENTS following our coding standards.
```

Usage: `/fix-issue 123 high-priority`
Result: `$ARGUMENTS` becomes `"123 high-priority"`

### Individual arguments
```markdown
---
argument-hint: [pr-number] [priority] [assignee]
description: Review pull request with specific criteria
---

Review PR #$1 with $2 priority and assign to $3.
Focus on security, performance, and code style.
```

Usage: `/review-pr 456 high alice`
Result: `$1="456"`, `$2="high"`, `$3="alice"`

## Bash Execution Pattern

Execute commands before Claude processes the slash command:

```markdown
---
allowed-tools: Bash(git status:*), Bash(git diff:*)
description: Create git commit based on current changes
---

## Context

Current status: !`git status`
Current diff: !`git diff HEAD`
Recent commits: !`git log --oneline -10`

## Your Task

Based on the above changes, create a single git commit.
```

**Important**: Must include `allowed-tools` with specific bash commands.

## File References

Include file contents with `@` prefix:

```markdown
Review the implementation in @src/utils/helpers.js

Compare @src/old-version.js with @src/new-version.js
```

## Frontmatter Fields

| Field | Purpose | Example |
|-------|---------|---------|
| `description` | Brief description shown in `/help` | `Review code for security issues` |
| `argument-hint` | Arguments shown in autocomplete | `[pr-number] [priority]` |
| `allowed-tools` | Tools this command can use | `Bash(git *:*), Read` |
| `model` | Specific model to use | `claude-3-5-haiku-20241022` |
| `disable-model-invocation` | Prevent SlashCommand tool from calling | `true` |

## Examples

### Simple Prompt Command
```markdown
---
description: Review code for security vulnerabilities
---

Review this code for:
- SQL injection
- XSS vulnerabilities
- Exposed secrets
- Auth bypass

Provide specific fixes for each issue found.
```

### Command with Pre-execution
```markdown
---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*)
description: Create a git commit
---

## Context

- Current git status: !`git status`
- Current git diff (staged and unstaged changes): !`git diff HEAD`
- Current branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -10`

## Your Task

Based on the above changes, create a single git commit.
```

### Command with Arguments
```markdown
---
argument-hint: [service] [environment]
description: Deploy service to environment
---

Deploy $1 to $2 environment.

Checklist:
- Verify tests pass
- Check environment variables
- Monitor deployment logs
- Verify health checks
```

## Location and Naming

**Project commands**: `.claude/commands/command-name.md`
- Shared with team via git
- Shows "(project)" in `/help`

**User commands**: `~/.claude/commands/command-name.md`
- Personal, not committed
- Shows "(user)" in `/help`

**Naming**: Use kebab-case matching the filename without `.md`

## Namespacing with Subdirectories

Organize commands in subdirectories:

`.claude/commands/frontend/component.md` → `/component` (description shows "project:frontend")

Subdirectories are for organization only, don't affect command name.

## SlashCommand Tool

Commands with `description` frontmatter can be invoked by Claude automatically via the SlashCommand tool.

**To prevent automatic invocation**:
```markdown
---
disable-model-invocation: true
---
```

**To encourage invocation**:
Reference the command by name in CLAUDE.md or prompts:
```
Run /write-unit-test when you are about to start writing tests.
```

## Thinking Mode

Commands can trigger extended thinking by including keywords like:
- "think carefully"
- "analyze step by step"
- "consider multiple approaches"

## Project-Specific Patterns

### Database Commands
```markdown
---
description: Show proper database workflow
---

NEVER create one-off scripts. Always use migrations:

1. Create migration file in migrations/
2. Write SQL for schema change
3. Apply migration: `engineerdna migrate up`
4. Verify with `go test ./internal/db/...`
```

### Deployment Commands
```markdown
---
description: Show release workflow
---

NEVER deploy manually. Use semantic versioning:

1. Update CHANGELOG.md with version bump
2. Commit to main branch
3. GitHub Actions creates release with binaries
4. Users download for their platform

For cross-compilation, see Makefile.
```

## Testing Commands

Test commands before committing:

1. Create command file
2. Test with `/command-name`
3. Verify arguments work: `/command-name arg1 arg2`
4. Check bash execution if using `!` prefix
5. Verify file references resolve correctly

## Current Project Commands

See `.claude/commands/` for examples (if any exist in your project).

Common patterns:
- Pre-work verification
- Post-work verification
- Database migration guide
- Dead code detection

Follow these patterns for consistency.
