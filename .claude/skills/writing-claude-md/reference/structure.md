# Structure & Content Guidelines

## CLAUDE.md Hierarchy

CLAUDE.md files exist at different levels, each with a specific purpose:

| Location | Purpose | Priority | Keep Under |
|----------|---------|----------|-----------|
| Root `CLAUDE.md` | Project-wide rules, domain model, critical gotchas | Highest | 500 lines |
| App-level `CLAUDE.md` (e.g., `src/app/CLAUDE.md`) | App routing and API patterns specific to Next.js structure | Medium | 200 lines |
| Feature `CLAUDE.md` (e.g., `src/lib/auth/CLAUDE.md`) | Context-specific rules and patterns for that subsystem | Lower | 100 lines |

**Key rule**: Never duplicate content across files. If it's in root, don't repeat it in subdirectories.

## What to Include

### DO Include

1. **Critical project-specific rules** that violate defaults or common practices
2. **Domain model** - Core entities and their relationships (usually a table)
3. **Non-obvious gotchas** with non-obvious fixes (not "React is reactive")
4. **Security patterns** specific to this codebase
5. **Project-specific conventions** (import paths, naming, structure)
6. **Integration patterns** - How different parts of the system work together

### DON'T Include

1. **Basic programming concepts** - Claude knows React, TypeScript, databases
2. **Library documentation** - Claude knows how to use common libraries (Next.js, React, Drizzle, etc.)
3. **Verbose code examples** - Only show project-specific patterns, not library usage
4. **Generic advice** - "Write good code" or "use best practices" adds no value
5. **Detailed API listings** - Point to authoritative sources (code, official docs) instead
6. **Information already in code** - If it's self-documenting in the code, don't duplicate

## Template: Root CLAUDE.md

```markdown
# Project Name - Domain & Rules

## Domain Model

| Entity | Purpose | Key Fields | Notes |
|--------|---------|-----------|-------|
| User | System users | id, email, role | Roles: admin, member, viewer |
| Organization | Team container | id, name, ownerId | Isolation boundary |
| Project | Org sub-resource | id, orgId, name | Belongs to org |

## Critical Rules

1. **Rule name** - Description with consequence
2. **Another rule** - Why it matters
... (up to 30 rules max)

## Quick Reference

- Import paths: Use `@/` prefix for src aliases
- API routes: Use `withAuth` wrapper from `@/lib/api-middleware`
- Database: Always use Drizzle migrations, never raw SQL
```

## Template: Subdirectory CLAUDE.md

Keep these SHORT - only directory-specific gotchas and patterns.

```markdown
# Directory CLAUDE.md

## CRITICAL GOTCHAS

1. Non-obvious issue → How to fix it
2. Security requirement → Why it matters

## PATTERNS

Brief description of patterns specific to this directory.

## See Also

- Link to related root CLAUDE.md sections
- Reference to external documentation
```

## What Not to Repeat in Subdirectories

If it's already in root CLAUDE.md, don't repeat it in subdirectories:
- Domain model
- General security patterns (only mention overrides/specifics)
- General conventions (only mention directory-specific exceptions)
- Generic project rules

Only document what's unique to that subdirectory.

## Size Targets

- **Root**: Keep under 500 lines (includes domain model + rules)
- **App-level**: Keep under 200 lines
- **Feature-level**: Keep under 100 lines

If exceeding these, extract detailed content into reference files and link to them.
