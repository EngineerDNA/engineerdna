# Skill Troubleshooting and Maintenance

## Skill Not Activating

**Symptom**: Skill exists but doesn't auto-activate

**Diagnosis steps**:
```bash
# 1. Validate all skills are configured
python3 .claude/scripts/validate-skill-config.py

# 2. Test if skill activates for specific prompt
echo '{"prompt":"test the API endpoint"}' | \
  CLAUDE_PROJECT_DIR=$(pwd) \
  python3 .claude/hooks/activate_skills.py

# 3. Check skill file exists
ls .claude/skills/[skill-name]/SKILL.md

# 4. Check skill-rules.json has entry
cat .claude/hooks/skill-rules.json | grep -A 20 '"skill-name"'
```

**Common fixes**:
- Add missing entry to `.claude/hooks/skill-rules.json`
- Expand keywords to include user's phrasing
- Add intent patterns for natural language variations
- Check JSON syntax is valid
- Manual invocation: "Use the [skill-name] skill to..."

---

## Adding New Skill

CRITICAL: Skills must be registered in BOTH locations to work:

**Step 1: Create skill file**
```bash
# Create .claude/skills/my-skill/SKILL.md
---
name: my-skill
description: When to use this skill
tools: Read, Bash, Grep
---

Skill instructions here...
```

**Step 2: Register in skill-rules.json**
```bash
# Add to .claude/hooks/skill-rules.json
"my-skill": {
  "type": "workflow",           # workflow, utility, reference
  "enforcement": "suggest",      # mandatory, suggest
  "priority": "high",           # critical, high, medium, low
  "promptTriggers": {
    "keywords": [
      "keyword1",
      "keyword2"
    ],
    "intentPatterns": [
      "(pattern1|pattern2).*?action",
      "context.*?(trigger)"
    ]
  },
  "fileTriggers": {
    "pathPatterns": [],
    "contentPatterns": []
  }
}
```

**Step 3: Validate configuration**
```bash
python3 .claude/scripts/validate-skill-config.py
```

**Step 4: Test activation**
```bash
echo '{"prompt":"test prompt here"}' | \
  CLAUDE_PROJECT_DIR=$(pwd) \
  python3 .claude/hooks/activate_skills.py
```

---

## Maintaining Skills

Skills can fail to activate if configuration is incomplete or missing.

**Symptoms**:
- Skill exists but never auto-activates
- Claude doesn't suggest relevant skills

**Diagnosis**:
```bash
# Check all skills are configured
python3 .claude/scripts/validate-skill-config.py

# Test specific prompt
echo '{"prompt":"your test prompt"}' | \
  CLAUDE_PROJECT_DIR=$(pwd) \
  python3 .claude/hooks/activate_skills.py
```

**Common fixes**:
1. Add missing skill entry to skill-rules.json
2. Expand keywords to include common phrasings
3. Add intent patterns for natural language variations
4. Test with realistic user prompts

---

## skill-rules.json Format

```json
{
  "skill-name": {
    "type": "workflow",           // workflow, utility, reference
    "enforcement": "suggest",      // mandatory (blocks), suggest (optional)
    "priority": "high",           // critical, high, medium, low

    "promptTriggers": {
      "keywords": [
        // Simple string matching (case-insensitive)
        "database migration",
        "schema change",
        "alter table"
      ],
      "intentPatterns": [
        // Regex patterns for natural language
        "(add|create|modify).*?(table|column|index)",
        "database.*?(change|update|migration)"
      ]
    },

    "fileTriggers": {
      "pathPatterns": [
        // Glob patterns for file paths
        "packages/database/src/schema/*.ts",
        "**/migrations/*.sql"
      ],
      "contentPatterns": [
        // Regex patterns for file content
        "CREATE TABLE",
        "ALTER TABLE"
      ]
    }
  }
}
```

### Priority Levels

| Priority | When to Use | Example |
|----------|-------------|---------|
| critical | MUST run for correctness | Database migrations (Rule 33) |
| high | Strongly recommended | Deployment recovery, testing workflows |
| medium | Helpful but optional | API testing, dev server checking |
| low | Convenience/reference | Documentation lookup |

### Enforcement Levels

| Enforcement | Behavior | Example |
|-------------|----------|---------|
| mandatory | Blocks if not used | migrating-databases (Rule 33) |
| suggest | Shows suggestion only | Most skills |

### Type Categories

| Type | Purpose | Examples |
|------|---------|----------|
| workflow | Multi-step procedures | testing-features, recovering-deployments |
| utility | Single-purpose tools | checking-dev-servers, testing-apis |
| reference | Documentation/lookup | using-mcp-tools, development-patterns |

---

## Validation Script

The validation script (`validate-skill-config.py`) ensures:
1. All SKILL.md files have skill-rules.json entries
2. All skill-rules.json entries have corresponding SKILL.md files
3. JSON syntax is valid
4. Required fields are present (type, enforcement, priority, promptTriggers)

Run after any skill changes:
```bash
python3 .claude/scripts/validate-skill-config.py
```

Exit codes:
- 0: All validations passed
- 1: Validation errors found (see output for details)
