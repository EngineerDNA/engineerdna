# File Structure & Format

## Directory Layout

```
.claude/skills/
  skill-name/
    SKILL.md              # Required: main skill file
    references/           # Optional: detailed guides
      best-practices.md
      structure.md
      progressive-disclosure.md
    scripts/              # Optional: helper scripts
      helper.py
      analyze.sh
    examples/             # Optional: code examples
      example1.py
      example2.ts
```

## SKILL.md Format

### Metadata (Frontmatter)

All Skills start with YAML frontmatter:

```yaml
---
name: skill-name
description: What it does and when to use it.
allowed-tools: Read, Write, Bash  # optional
model: sonnet                      # optional
---
```

### Content Structure

```markdown
# Skill Name

Brief overview (1-2 sentences).

## Quick Start

Essential steps or command.

## Common Use Cases

(optional) List 2-3 typical scenarios.

## References

- [Reference 1](references/file1.md)
- [Reference 2](references/file2.md)
```

## Frontmatter Fields

### Required Fields

- **name** (string)
  - Must be lowercase letters, numbers, hyphens only
  - Max 64 characters
  - Use gerund form: `processing-pdfs`, `analyzing-spreadsheets`
  - Avoid: `helper`, `utils`, `tools`, `documents`

- **description** (string)
  - Max 1024 characters
  - **Must include**: What the skill does + when to use it
  - Format: "[What it does]. Use when [trigger keywords]."
  - Example: "Extract text and tables from PDF files. Use when working with PDFs or when the user mentions document extraction."
  - Always write in third person

### Optional Fields

- **allowed-tools** (string, comma-separated)
  - Tools the skill is permitted to use
  - Common: `Read, Write, Bash, Grep, Glob`
  - Omit if no restrictions (skill inherits all tools)
  - Use to limit scope (e.g., read-only skills omit Write/Edit)

- **model** (string)
  - Options: `sonnet`, `opus`, `haiku`, `inherit`
  - Default: `inherit` (uses project/user setting)
  - Use `opus` for complex reasoning
  - Use `haiku` for simple, fast operations

### Invalid/Unused Fields

These don't work in frontmatter:
- `author` - Not supported
- `version` - Skills versioned with plugin
- `tags` - Not supported (use description keywords)

## Best Practices for Structure

### Progressive Disclosure Template

```yaml
---
name: skill-name
description: Brief what + when to use it.
allowed-tools: Read, Bash
---

# Skill Title

[1-2 sentence overview]

## Quick Start

[Essential commands or steps]

## Common Patterns

[2-3 typical use cases with brief examples]

## References

- [Detailed Guide](references/guide.md)
- [Advanced Examples](references/examples.md)
- [Troubleshooting](references/troubleshooting.md)
```

### File Naming

- `best-practices.md` - Principles, conventions, maintenance
- `structure.md` - Format, templates, file layout
- `progressive-disclosure.md` - How to organize content
- `examples.md` - Detailed walkthroughs and examples
- `troubleshooting.md` - Common issues and solutions
- `quick-reference.md` - Command/syntax reference

### Keep References Flat

- All reference files live directly in `references/`
- No nested subdirectories
- Link directly from SKILL.md: `[text](references/file.md)`
- Avoid: `[text](references/guides/file.md)`

## Validation Checklist

Before publishing your skill:

- [ ] `name` is lowercase, hyphens only, ≤64 characters
- [ ] `description` includes what + when, third person, ≤1024 chars
- [ ] SKILL.md is ≤50 lines total (excluding excessive blank lines)
- [ ] SKILL.md has Quick Start section with runnable example
- [ ] All references are one level deep in `references/`
- [ ] No broken links in SKILL.md
- [ ] `allowed-tools` accurately reflects tool usage
- [ ] No vendor-specific paths (use `${CLAUDE_PLUGIN_ROOT}` if portability needed)
