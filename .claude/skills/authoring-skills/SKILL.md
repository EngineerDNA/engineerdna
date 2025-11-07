---
name: authoring-skills
description: Write effective Agent Skills with concise, discoverable content. Include what/when in description, use progressive disclosure for details. Use when creating or refactoring Skills.
allowed-tools: Read, Write, Bash
---

# Authoring Agent Skills

Write Skills that Claude can discover and use effectively.

## Core Principle: Progressive Disclosure

Keep SKILL.md concise and focused. Point users to detailed references for deep dives.

## Quick Start

**1. Define Skill Metadata:**
```yaml
---
name: skill-name  # lowercase, hyphens only, max 64 chars
description: What it does and when to use it. Use when [triggers].
allowed-tools: Read, Write, Bash  # optional
---
```

**2. Structure:** Overview + Quick Start + References

**3. Organize Details:** Move content to `references/` directory


## Key Principle: Description Drives Discovery

- **Include what AND when**: "Extract PDFs. Use when user mentions PDFs or forms."
- **Write in third person**: "Processes Excel files" (not "I can help you")
- **Add trigger keywords**: users must recognize when skill applies

## References

- [Best Practices](references/best-practices.md) - Principles, conventions, maintenance
- [Structure & Format](references/structure.md) - SKILL.md format and frontmatter fields
- [Progressive Disclosure Patterns](references/progressive-disclosure.md) - Organizing detailed content
