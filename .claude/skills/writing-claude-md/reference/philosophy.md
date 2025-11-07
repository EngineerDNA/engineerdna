# Philosophy & Principles

## Core Philosophy: "Less is More"

**Default assumption**: Claude is already very smart. Only add context Claude doesn't already have.

Challenge each piece of information:
- "Does Claude really need this explanation?"
- "Can I assume Claude knows this?"
- "Does this paragraph justify its token cost?"

### Why Conciseness Matters

1. **Token efficiency** - Every line consumes tokens; fewer tokens = faster responses
2. **Signal-to-noise ratio** - Verbose content obscures critical information
3. **Maintenance burden** - Short files are easier to keep updated
4. **Focus** - Constraints force clarity on what's truly important

### The "Less is More" Challenge

When writing CLAUDE.md, follow this decision tree:

```
Does Claude need this information?
├─ No → Delete it (saves tokens, reduces noise)
├─ Yes, it's standard knowledge → Delete it (Claude knows this)
├─ Yes, it's project-specific → Keep it and make it concise
└─ Yes, but it's detailed → Move to reference docs, keep one-liner in CLAUDE.md
```

## Key Principles

### Principle 1: Assume Claude Knows Basics

Don't explain:
- How React works
- TypeScript syntax and types
- Standard library functions
- Common design patterns
- How to use popular libraries (React, Next.js, etc.)

Claude already knows these and will ask if clarification is needed.

### Principle 2: Document Project-Specific Rules Only

Include:
- Rules that violate standard practices
- Domain-specific terminology and entities
- Project-specific conventions
- Non-obvious gotchas with non-obvious fixes
- Security patterns unique to this codebase

### Principle 3: Point Outward, Don't Duplicate

Instead of:
```markdown
## Environment Variables

NEXT_PUBLIC_API_URL - URL for API (required)
NEXT_PUBLIC_STRIPE_KEY - Stripe publishable key (required)
DATABASE_URL - PostgreSQL connection string (required)
... (20 more variables)
```

Write:
```markdown
## Environment Variables

See `.env.local.example` for complete list.
```

### Principle 4: Prefer Tables Over Paragraphs

Instead of:
```markdown
The user table has an id field which is the primary key. It also has an email field
which must be unique. There's a role field for authorization. Users have a createdAt
and updatedAt timestamp.
```

Write:
```markdown
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | Primary key |
| email | string | Unique, required |
| role | enum | 'admin' \| 'user' \| 'viewer' |
| createdAt | timestamp | Auto-set |
| updatedAt | timestamp | Auto-updated |
```

### Principle 5: Use One-Liners for Common Patterns

Instead of showing all imports:
```markdown
## Components

Import UI components from `@/components/ui/*` - all follow shadcn/ui patterns.
```

Instead of verbose API documentation:
```markdown
## API Routes

All routes use `withAuth` wrapper from `@/lib/api-middleware` - provides user context automatically.
```
