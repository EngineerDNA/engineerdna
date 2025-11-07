# Automatic Workflow Systems

Three fully automatic systems enhance the development workflow with zero manual intervention.

## 1. Skills Auto-Activation

**Purpose**: Automatically suggests relevant skills based on your prompts
**Hook**: UserPromptSubmit (runs BEFORE Claude sees your message)
**How it works**:
- Analyzes prompts for keywords and intent patterns
- Matches against 14 skills with priority levels (CRITICAL/HIGH/MEDIUM/LOW)
- Shows suggestions with reasoning before Claude responds
- Non-blocking: Claude decides whether to use suggested skills

**Example Output**:
```
================================================================================
SKILL ACTIVATION CHECK
================================================================================

[CRITICAL] CRITICAL PRIORITY:
  - migrating-databases
    [MANDATORY] Matched: keyword match + intent pattern

To use a skill: "Use the [skill-name] skill to..."
Or let me decide if the skill context is needed.
================================================================================
```

**Configuration**: `.claude/hooks/skill-rules.json` (see skill-troubleshooting.md for format)

---

## 2. Automatic Build Checking

**Purpose**: Catches Go compilation errors immediately after editing files
**Hooks**: PostToolUse (tracks edits) + Stop (runs go vet)
**How it works**:
- Tracks all Go file edits during session
- Automatically runs `go vet ./...` after Claude finishes responding
- Shows errors immediately with smart formatting
- Suggests quality agent for systematic fixes when 5+ errors found

**Example Output**:
```
================================================================================
AUTOMATIC BUILD CHECK
================================================================================

[FAIL] Go vet check failed with 3 error(s)

Errors found:
internal/plugin/loader.go:45:12: undefined: pluginPath
internal/db/events.go:89:5: result declared but not used

================================================================================
```

**Result**: Zero Go compilation errors left behind

---

## 3. Automatic Dev Docs Management

**Purpose**: Preserves context across compaction with structured documentation
**Hook**: Stop (analyzes Claude's responses)
**How it works**:

**Phase 1: Plan Detection & Creation**
- Detects comprehensive plans in Claude's responses
- Automatically creates PLAN.md, CONTEXT.md, TASKS.md at project root
- Extracts feature name for tracking

**Phase 2: Continuous Updates**
- Updates CONTEXT.md with design decisions automatically
- Updates "Next Steps" section for post-compaction resumption
- Tracks progress with timestamps

**Phase 3: Automatic Archiving**
- Detects when starting a different feature
- Archives old dev docs to `.claude/logs/archive/[timestamp]_[feature]/`
- Creates fresh docs for new feature
- Prevents conflicts with fuzzy feature name matching

**The Three Files**:

`PLAN.md` - Implementation strategy
- Executive summary, objectives, phases
- Risks, success metrics, confidence scores
- Timeline estimates

`CONTEXT.md` - Technical context (updated automatically)
- Architecture overview, key files
- Design decisions (auto-populated)
- Next steps (auto-updated)
- Last updated timestamp

`TASKS.md` - Actionable checklist
- Markdown checkboxes: `- [ ]` pending, `- [x]` complete
- Extracted from plan, manually updated

**Archive Retention**: Automatically limits to 50 most recent archives

---

## Disabling Automatic Systems

If automatic systems cause issues, disable them in `.claude/settings.json`:

### Disable Skills Auto-Activation
```json
// Comment out lines 12-22 in settings.json
/*
"UserPromptSubmit": [
  {
    "matcher": "",
    "hooks": [
      {
        "type": "command",
        "comment": "Skills auto-activation",
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/activate_skills.py"
      }
    ]
  }
],
*/
```

### Disable Build Checking
```json
// Comment out:
// - Lines 84-89 (track_edits.py in PostToolUse)
// - Lines 112-116 (check_build.py in Stop)
```

### Disable Dev Docs Management
```json
// Comment out lines 107-111 in Stop hooks
/*
{
  "type": "command",
  "comment": "Automatic dev docs management",
  "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/manage_dev_docs.py"
}
*/
```

### State Cleanup
After disabling, clean up state files:
```bash
rm -f .claude/logs/edit_tracking.json
rm -f .claude/logs/dev_docs_state.json
# Archives preserved: .claude/logs/archive/
```

### Re-enabling
Uncomment the sections in settings.json and reload Claude Code.
