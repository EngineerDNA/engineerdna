# Best Practices & Maintenance

## Root CLAUDE.md: Essential Content

The root CLAUDE.md is the single source of truth for project-wide information. Include only:

### 1. Domain Model (Required)

Present as a table for scannability:

```markdown
| Entity | Purpose | Key Fields | Notes |
|--------|---------|-----------|-------|
| User | Authenticated system user | id, email, role | Roles determine permissions |
| Organization | Team/account container | id, name, ownerId | Isolation boundary |
| Project | Work item container | id, orgId, name, status | Belongs to org |
```

**Why**: Domain model is the foundation - every developer needs to understand it immediately.

### 2. Critical Rules (1-30)

Numbered list of non-negotiable requirements:

```markdown
## Critical Rules

1. All user-facing data must be isolated by organizationId
2. Always use Drizzle migrations - never raw SQL scripts
3. API routes must be wrapped with withAuth middleware
4. Components must include JSDoc describing behavior
5. ...
```

**Why**: Forces prioritization. If you have >30, some aren't actually critical.

### 3. Common Gotchas

Non-obvious project-specific issues with fixes:

```markdown
## Critical Gotchas

1. **RLS Policies**: New tables must have RLS policies or queries fail in prod
   - Fix: Add policy in migration SQL: `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`

2. **Session Middleware**: Verify user in middleware before route handler
   - Fix: Use withAuth wrapper - it does this automatically
```

**Why**: Prevents repeated mistakes and saves debugging time.

### 4. Quick Reference (Optional)

One-liners for common operations:

```markdown
## Quick Reference

- **Build**: `pnpm run build` (builds Next.js)
- **Dev server**: `pnpm run dev` (runs on localhost:3000)
- **Migrations**: `pnpm db:generate` -> `pnpm db:migrate`
```

---

## Subdirectory CLAUDE.md: Minimal Content

Keep these SHORT. Only include directory-specific content NOT in root:

### What SHOULD Be in Subdirectory Files

- Directory-specific gotchas or quirks
- Patterns unique to that directory
- References to detailed documentation

### What SHOULD NOT Be in Subdirectory Files

- Domain model (link to root instead)
- General project rules (link to root instead)
- Basic library usage (Claude knows this)

### Example: `src/lib/auth/CLAUDE.md`

```markdown
# Authentication Library

## Patterns

- Use `getAuth()` in API routes (server-side)
- Use `useAuth()` hook in components (client-side)
- Token refresh happens automatically in middleware

## See Also

- Root CLAUDE.md for domain model and critical rules
- NextAuth.js documentation for configuration
```

---

## Maintenance Workflow

### When CLAUDE.md Gets Stale

Signs your CLAUDE.md is out of sync with reality:
- New conventions aren't documented
- Old patterns are still listed
- Developers ask questions that CLAUDE.md already answers
- More than 50% of content is about how libraries work

### Monthly Maintenance

1. **Remove outdated content** - Delete rules, patterns, gotchas that no longer apply
2. **Add new gotchas** - Document recent mistakes/lessons learned
3. **Trim verbose sections** - Challenge every paragraph
4. **Check for redundancy** - Ensure content isn't duplicated across files
5. **Update examples** - Replace outdated code samples

### When Adding New Rules

1. **Keep the list ordered by importance** - Critical rules at the top
2. **Make rules specific and testable** - "Use migrations" (good), "Write good code" (bad)
3. **Include the "why"** - Brevity: "**No hardcoded configs** - breaks across environments"

---

## Common Mistakes to Avoid

### Mistake 1: Duplicating Content Across Files

**Bad**:
- Root CLAUDE.md: Full domain model
- `src/lib/auth/CLAUDE.md`: Domain model again (for context)
- `src/lib/db/CLAUDE.md`: Domain model yet again

**Good**:
- Root CLAUDE.md: Domain model (single source of truth)
- `src/lib/auth/CLAUDE.md`: "See root CLAUDE.md for domain model"
- `src/lib/db/CLAUDE.md`: "See root CLAUDE.md for domain model"

### Mistake 2: Including Library Documentation

**Bad**:
```markdown
## React Hooks

useEffect runs after render. useState returns a value and setter function...
[100 lines of React documentation]
```

**Good**:
```markdown
## Project Hooks

- `useAuth()` - Returns current user and auth state
- `useOrganization()` - Returns current org context
- See React docs for standard hooks usage
```

### Mistake 3: Verbose Examples Instead of Project-Specific Patterns

**Bad**:
```markdown
## API Routes

Create a file in pages/api/, export a function:

export default function handler(req, res) {
  // Handle GET requests
  if (req.method === 'GET') { ... }
  // Handle POST requests
  if (req.method === 'POST') { ... }
}
```

**Good**:
```markdown
## API Routes

Use `withAuth` middleware wrapper from `@/lib/api-middleware`:
- Automatically injects user context
- Returns 401 if not authenticated
```

### Mistake 4: Exceeding Size Limits

**Signals you're exceeding size limits**:
- Root CLAUDE.md >500 lines
- Subdirectory CLAUDE.md >200 lines
- Reading it takes >5 minutes

**Solution**: Extract detailed content to reference files, keep only links in CLAUDE.md.

### Mistake 5: Letting Documentation Drift from Code

**Bad scenario**:
1. Wrote CLAUDE.md 6 months ago
2. Architecture changed but CLAUDE.md wasn't updated
3. New developers follow outdated patterns

**Prevention**:
- Update CLAUDE.md when architecture changes
- Include in code reviews: "Is CLAUDE.md still accurate?"
- Audit quarterly

---

## Structure Template: Root CLAUDE.md

```markdown
# [Project Name] - Domain & Rules

## Domain Model

[Table with entities]

## Critical Rules

1. Rule with consequence
2. Another rule
... (prioritized)

## Critical Gotchas

1. Gotcha -> How to fix
2. Another gotcha -> Fix

## Quick Reference

- Key commands
- Important imports
- Links to detailed docs

## See Also

- Architecture documentation
- Development setup guide
```

---

## Structure Template: Subdirectory CLAUDE.md

```markdown
# [Directory Name]

## Directory-Specific Patterns

Brief description unique to this directory.

## Common Gotchas

Issues specific to this directory.

## See Also

- Root CLAUDE.md for domain model and critical rules
- Parent directory CLAUDE.md for context
- External documentation links
```

---

## Integration with Development Workflow

### During Code Review

**Checklist for reviewers**:
- Does this change violate any rules in CLAUDE.md? If yes, request fix
- Should CLAUDE.md be updated for this change? If yes, request addition/update
- Is this a gotcha that should be documented? If yes, suggest adding to CLAUDE.md

### During Onboarding

**For new team members**:
1. Read root CLAUDE.md (5 minutes)
2. Read relevant subdirectory CLAUDE.md files
3. Reference CLAUDE.md when confused about patterns
4. Ask team if something in CLAUDE.md seems outdated

### During Refactoring

When refactoring:
1. Update CLAUDE.md to reflect new patterns
2. Remove documentation of old patterns
3. Mark gotchas as resolved if fixed
