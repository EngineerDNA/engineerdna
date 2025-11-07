---
name: writing-claude-md
description: Write concise CLAUDE.md files following "less is more" philosophy. Only include context Claude doesn't already know. Keep under 500 lines.
allowed-tools: Read, Edit, Write
---

# Writing CLAUDE.md Files

Create concise, focused CLAUDE.md files that provide only critical context Claude doesn't already have.

## Core Philosophy: "Less is More"

Only document what Claude doesn't already know. Challenge every line: "Does Claude really need this?" If not, delete it.

**Keep under 500 lines** for optimal performance.

## Quick Start

1. **Identify critical content** - Only include project-specific rules, domain models, gotchas, security patterns, and conventions
2. **Organize hierarchically** - Root CLAUDE.md for project-wide rules; subdirectories for context-specific rules
3. **Condense aggressively** - Use tables, one-liners, and references instead of verbose explanations
4. **Point outward** - Reference authoritative sources instead of duplicating content
5. **Avoid common pitfalls** - Never include basic programming knowledge or generic library documentation

## Critical Rules

- **DO**: Critical project rules | Domain model | Non-obvious gotchas | Security patterns | Project conventions
- **DON'T**: Basic programming | Library docs | Verbose examples | Generic advice | API listings
- **Structure**: Root CLAUDE.md (domain model + critical rules) | Subdirectory CLAUDE.md (directory-specific only)

## References

- [Philosophy & Principles](reference/philosophy.md) - "Less is more" philosophy and core principles
- [Structure & Content](reference/structure.md) - CLAUDE.md templates and what to include/exclude
- [Examples](reference/examples.md) - Verbose vs. concise examples and condensing patterns
- [Best Practices](reference/best-practices.md) - Root vs. subdirectory content, maintenance workflow
