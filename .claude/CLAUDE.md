# .claude Directory - Claude Code Configuration Guide

## WHAT IS THIS DIRECTORY?

The `.claude/` directory contains all Claude Code configuration for the EngineerDNA project:
- **Agents**: Specialized AI subagents for task-specific workflows
- **Skills**: Multi-file workflows for complex procedures
- **Hooks**: Event-driven automation and validation
- **Settings**: Project-wide configuration

## DIRECTORY STRUCTURE

```
.claude/
├── CLAUDE.md                    # This file - navigation guide
├── settings.json                # Configuration and hooks
├── settings.local.json          # Local overrides (gitignored)
├── agents/                      # AI subagents (8 total)
│   ├── CLAUDE.md               # Agent architecture guide
│   ├── engineer.md             # Full-stack implementation (Go + React)
│   ├── quality.md              # Testing and verification
│   ├── data-engineer.md        # Data integrity and database patterns
│   ├── integration-checker.md  # Plugin system testing
│   ├── ui-tester.md            # Frontend/UI testing
│   ├── docs.md                 # Documentation management
│   └── advisor.md              # Code review
├── skills/                      # Complex workflows (3 total)
│   ├── using-claude-code/
│   │   ├── SKILL.md
│   │   └── reference/
│   ├── database-migrations/
│   │   └── SKILL.md
│   └── dev-server-diagnostics/
│       └── SKILL.md
├── hooks/                       # Automated hooks (20 total)
│   ├── activate_skills.py       # Auto-suggest skills
│   ├── skill-rules.json         # Skill mapping config
│   ├── prevent_main_commits.sh  # Block commits to main
│   ├── validate_no_emojis.py    # Enforce no emojis
│   ├── validate_no_dev_markers.sh # Block TODO/FIXME/HACK
│   ├── prevent-no-verify.py     # Block --no-verify flag
│   ├── validate_version_bump.py # Enforce CHANGELOG updates
│   ├── validate_dead_code.py    # Check for dead code (golangci-lint)
│   ├── validate_formatting.py   # Enforce code formatting
│   ├── validate_schema_changes.py # Require migrations for schema changes
│   ├── validate_db_operations.py # Block direct DB modifications
│   ├── validate_critical_patterns.py # Validate security patterns
│   ├── validate_md_files.py     # Block unauthorized markdown
│   ├── validate_file_size.py    # Enforce 500 LOC limit
│   ├── validate_bash_usage.py   # Enforce proper tool usage
│   ├── enforce_restart_script.py # Block manual server management
│   ├── track_edits.py           # Track file edits
│   ├── check_build.py           # Auto-run go vet
│   ├── check_ignored_errors.py  # Detect ignored errors
│   ├── log_agent_metrics.sh     # Log agent metrics
│   └── manage_dev_docs.py       # Auto-manage dev docs
├── scripts/                     # Validation utilities
└── logs/                        # Metrics and archives (gitignored)
```

## QUICK START

### Before ANY Work
```bash
pwd
git status
go vet ./...
# When frontend added: cd frontend && npm run lint
```

### After ANY Work

**Backend (Go):**
```bash
go vet ./...
staticcheck ./...
go test ./...
make build
```

**Frontend (when added):**
```bash
cd frontend
npm run lint
npm run typecheck
npm run build
npm test
```

**Always:**
```bash
git diff  # review changes
```

## AGENTS (AI Subagents)

Specialized AI assistants for specific tasks. Automatically invoked by Claude or explicitly requested.

### Agent Pipeline (for new features)
```
engineer → quality → integration-checker → data-engineer → ui-tester → docs → advisor
```

### Key Agents

**engineer**: Full-stack implementation (Go backend + React frontend)
- When: "implement", "fix", "create", "build", "component", "API", "plugin"
- Output: Working code (Go, React, or both), files modified

**quality**: Testing and verification
- When: Build fail, test fail, errors
- Output: Test results, quality checks

**integration-checker**: Plugin system testing
- When: "plugin system", "test plugins", "JSON-RPC", "plugin discovery", "integration test"
- Output: Plugin system verification, integration test results

**data-engineer**: Data integrity and database patterns
- When: Schema changes, migrations, UTC timestamps, multi-modal data, query optimization
- Output: Data integrity verification, migration compliance

**ui-tester**: Frontend/UI testing
- When: "test UI", "frontend broken", "component not rendering", "React errors"
- Output: UI test results, screenshots, console errors

**docs**: Documentation management
- When: "update README", "write docs", "API documentation", "plugin guide"
- Output: Updated documentation, CHANGELOG entries

**advisor**: Code review
- When: "review", "ready for prod", "looks good?"
- Output: APPROVED/BLOCKED decision

### Invoking Agents

**Automatic** (Claude decides):
```
> I need to add a new table for tracking metrics
Claude uses: engineer → quality → data-engineer → advisor

> Create a dashboard component to display metrics
Claude uses: engineer (frontend mode) → ui-tester → advisor

> Implement the plugin loader
Claude uses: engineer (backend mode) → quality → integration-checker → advisor

> Update the README with installation instructions
Claude uses: docs → advisor
```

**Explicit** (you specify):
```
> Use the engineer agent to implement the plugin loader
> Have the integration-checker agent test the plugin system
> Use the ui-tester agent to verify the dashboard renders correctly
> Have the docs agent update the API documentation
```

## SKILLS

Complex multi-file workflows. Automatically invoked when relevant or explicitly requested.

### using-claude-code
- **When**: Creating Skills, Hooks, Agents, or questions about Claude Code patterns
- **Does**: Guides usage of Claude Code best practices and project structure

### database-migrations
- **When**: Schema changes, migrations, database operations
- **Does**: Guides safe database evolution with SQLite migrations

### dev-server-diagnostics
- **When**: "localhost not loading", "server not responding", connection errors
- **Does**: Diagnoses server issues, checks processes, analyzes logs

**To use explicitly**:
- "Use the database-migrations skill to add a new table"
- "Use the dev-server-diagnostics skill to debug localhost:3847"

## AUTOMATIC WORKFLOW SYSTEMS

### Skills Auto-Activation
The `activate_skills.py` hook analyzes prompts BEFORE Claude sees them and suggests relevant skills based on `skill-rules.json` configuration.

**How it works**:
1. You submit a prompt
2. Hook analyzes keywords and intent patterns
3. Matches to skills in skill-rules.json
4. Suggests relevant skills to Claude
5. Claude decides whether to use them

**Disable**: Remove or comment out the UserPromptSubmit hook in settings.json

### Automatic Build Checking
The `check_build.py` hook runs after Stop events to automatically check Go code compilation.

**How it works**:
1. `track_edits.py` tracks when .go files are modified
2. On Stop, `check_build.py` checks if Go files changed
3. If yes, runs `go vet ./...` automatically
4. Reports any compilation errors

**Disable**: Remove or comment out the Stop hooks in settings.json

## HOOKS REFERENCE

Event-driven automation. Runs automatically at specific points.

### UserPromptSubmit Hooks (Run BEFORE Claude sees prompts)

- **activate_skills.py**: Suggests relevant skills based on prompt analysis

### PreToolUse Hooks (Run BEFORE actions)

**Bash commands**:
- **prevent_main_commits.sh**: Block commits to main branch
- **validate_no_emojis.py**: Block emojis in commits and PRs
- **validate_no_dev_markers.sh**: Block TODO/FIXME/HACK in commits
- **prevent-no-verify.py**: Block --no-verify flag in git commands
- **validate_bash_usage.py**: Enforce proper tool usage

**File operations** (Edit, Write):
- **validate_md_files.py**: Enforce authorized .md files only
- **validate_file_size.py**: Enforce 500 LOC limit on source files
- **validate_no_emojis.py**: Block emoji characters in code

### PostToolUse Hooks (Run AFTER actions)

**Edit/Write operations**:
- **track_edits.py**: Track Go file edits for build checking

**Bash commands**:
- Remind to test after build

### Stop Hooks (Run at END of each response)

- **manage_dev_docs.py**: Auto-create and update development docs
- **check_build.py**: Auto-run `go vet` if Go files were edited
- Display post-work checklist reminder

### SessionStart Hooks (Run at START of session)

- Show environment info (directory, branch)
- Display pre-work and post-work checklists

### SubagentStop Hooks (Run AFTER subagent completes)

- **check_ignored_errors.py**: Detect and block ignored tool errors
- **log_agent_metrics.sh**: Track agent usage metrics

---

## COMMON WORKFLOWS

| Workflow | Steps | Notes |
|----------|-------|-------|
| New Feature | Implement → quality → data-engineer → advisor | Agent pipeline auto-invoked |
| Bug Fix | Fix → quality → data-engineer | Engineer agent direct |
| Security Audit | data-engineer agent review | Reviews encryption, validation |
| Code Review | advisor agent review | Final production readiness |

### Quick Reference

**Agent Usage**:
- Let Claude invoke agents automatically (it knows when)
- One task = one agent invocation (don't pass todo lists)
- Agent pipeline flows backward on errors (quality → engineer)

**Skills**:
- Keep SKILL.md under 500 lines (progressive disclosure)
- Use reference/ subdirectory for detailed docs
- Auto-invoked when description matches context

**Hooks**:
- Hooks solve problems, don't punt to Claude
- Enforce critical rules automatically (security, validation)
- Always quote shell variables, block path traversal

## TROUBLESHOOTING

### Agent Not Invoked
```
# Check agent description in .claude/agents/[name].md
# Use explicit invocation: "Use the [agent] agent to..."
```

### Skill Not Working
```
# Check skill description in .claude/skills/[name]/SKILL.md
# Verify skill-rules.json has correct triggers
# Use explicit invocation: "Use the [skill-name] skill to..."
```

### Hook Blocking Action
```
# Hooks enforce critical rules (main commits, emojis, dev markers)
# Fix the underlying issue, don't bypass the hook
# Check hook error message for specific violation
```

## EXTENDING CLAUDE CODE

### Adding New Skill

1. Create directory: `.claude/skills/my-skill/`
2. Create SKILL.md with frontmatter:
```yaml
---
name: my-skill
description: What this skill does and when to use it
---
# Content here
```
3. Add entry to `.claude/hooks/skill-rules.json`
4. Add reference/ subdirectory for detailed docs

### Adding New Agent

1. Create `.claude/agents/my-agent.md`
2. Add frontmatter with name, description, tools
3. Document agent's responsibility and protocols
4. Add to agent pipeline if needed

### Adding New Hook

Edit `.claude/settings.json` to add hook configuration.

See Claude Code documentation for hook event types.

## REFERENCE

### Project Documentation
- **Root CLAUDE.md**: Project-wide context and navigation (../CLAUDE.md)
- **.claude/CLAUDE.md**: This file - Claude Code configuration guide
- **Agent CLAUDE.md**: Agent architecture and patterns (agents/CLAUDE.md)

### Implementation Documentation
- **internal/db/CLAUDE.md**: Database patterns and migrations (../internal/db/CLAUDE.md)
- **internal/plugin/CLAUDE.md**: Host-side plugin system implementation (../internal/plugin/CLAUDE.md)
- **plugins/CLAUDE.md**: Plugin development guide (../plugins/CLAUDE.md)

### Official Documentation
- [Claude Code Documentation](https://docs.claude.com/en/docs/claude-code)
- [Skills Best Practices](https://docs.anthropic.com/en/docs/agents-and-tools/agent-skills/best-practices)
- [Hooks Guide](https://docs.anthropic.com/en/docs/claude-code/hooks-guide)
- [Hooks Reference](https://docs.anthropic.com/en/docs/claude-code/hooks)
- [Subagents](https://docs.anthropic.com/en/docs/claude-code/sub-agents)

## GETTING HELP

- Ask Claude: "What agents are available?"
- Ask Claude: "What skills are available?"
- Check agent files in `.claude/agents/` for capabilities
- Check skill files in `.claude/skills/` for workflows
