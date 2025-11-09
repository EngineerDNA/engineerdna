# Skill Authoring Best Practices

## The 500-Line Rule

**IMPORTANT**: This pattern is ONLY for skills (`.claude/skills/skill-name/`).

Keep SKILL.md under 500 lines using progressive disclosure:

```
.claude/skills/skill-name/
├── SKILL.md (overview + navigation, <500 lines)
├── reference/
│   ├── domain-1.md (loaded when needed)
│   └── domain-2.md
└── scripts/
    └── validate.py (executed, not loaded into context)
```

**For CLAUDE.md files elsewhere** (internal/db/CLAUDE.md, plugins/CLAUDE.md, etc.):
- Keep under 500 lines through concise writing
- DO NOT create reference/ subdirectories
- Condense inline, link to existing docs when appropriate

**Token efficiency**: Only pay for what you actually read.

## Structure Pattern

```markdown
---
name: skill-name
description: What it does and when to use it. Be specific with keywords.
---

# Quick Start
[Core instructions, assume Claude is smart]

## Advanced
See [reference/details.md](reference/details.md) for schema.

## Scripts
Run: `python scripts/validate.py input.json`
```

## What to Document

**DO document**:
- Project-specific patterns (plugin system, anonymization, encryption)
- Critical project rules (UTC, no one-off scripts, localhost-only binding)
- Domain knowledge (table schemas, API contracts, event model)
- Gotchas unique to this project

**DON'T document**:
- General programming concepts Claude already knows
- How standard libraries work
- What common file formats are

## Validation Pattern

For critical operations, use plan-validate-execute:

```markdown
## Workflow
1. Generate plan file (changes.json)
2. Validate: `python scripts/validate.py changes.json`
3. If fails → fix plan → validate again
4. Execute only when validation passes
```

This catches errors before they happen.

## File Organization

**One level deep from SKILL.md**:
```
SKILL.md → reference/domain.md → (scripts executed, not read)
```

**NOT multiple levels**:
```
# Bad - Claude may only partially read nested files
SKILL.md → advanced.md → details.md → actual-info.md
```

## Description Field

Must enable discovery. Include both **what** and **when**:

```yaml
# Good - specific with triggers
description: Extract text from PDFs and fill forms. Use when working with PDF files, forms, or document extraction.

# Bad - vague
description: Helps with documents
```

## Degrees of Freedom

Match specificity to task fragility:

**High freedom** (text instructions):
```markdown
1. Analyze code structure
2. Check for potential bugs
3. Suggest improvements
```

**Low freedom** (exact scripts for fragile operations):
```markdown
Run exactly: `python scripts/migrate.py --verify --backup`
Do not modify or add flags.
```

## When to Use Skills vs Commands

**Use Skill** if:
- Multiple files/scripts needed
- Complex workflows with validation
- Should auto-discover based on context
- Team needs standardized detailed guidance

**Use Command** if:
- Simple repeated prompt
- Fits in single file
- Want explicit invocation control

## Progressive Disclosure Example

```markdown
# SKILL.md (navigation hub)
## Quick Start
Basic usage here

## Domain-Specific Operations
**Finance data**: See [reference/finance.md](reference/finance.md)
**Sales data**: See [reference/sales.md](reference/sales.md)

Claude only reads the domain file when user asks about that domain.
```

## Conciseness Test

Before adding content, ask:
1. "Does Claude already know this?"
2. "Is this specific to your project?"
3. "Does this justify its token cost?"

All three must be "yes".

## Table of Contents for Long Files

For reference files >100 lines, add TOC at top:

```markdown
# API Reference

## Contents
- Authentication and setup
- Core methods
- Advanced features
- Error handling

## Authentication and setup
...
```

This helps Claude navigate even with partial reads.
