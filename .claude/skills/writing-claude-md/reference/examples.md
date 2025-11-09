# Examples: Verbose vs. Concise

## Example 1: API Routes

### TOO VERBOSE (150+ tokens)

```markdown
## API Routes

API routes in Next.js are server-side endpoints. They run on the server and can
access databases without exposing sensitive information. To create an API route,
you need to make a file in the app/api directory with a handler function that
accepts Request and Response objects. Here's a detailed example:

[50 lines of generic Next.js documentation about creating API routes, request/response
handling, error handling, etc.]
```

**Problem**: Describes how Next.js API routes work (Claude already knows this) instead of project-specific patterns.

### CONCISE (50 tokens)

```markdown
## API ROUTE PATTERNS

**Standard endpoints**: Wrap with `withAuth` middleware from `@/lib/api-middleware/with-auth`
- Automatically provides `user` context
- Returns 401 if not authenticated

**Admin-only endpoints**: Use `withAuth` + admin check in handler
```

**Better**: Only documents project-specific patterns.

---

## Example 2: Environment Variables

### TOO VERBOSE (200+ tokens)

```markdown
## Environment Variables

The application requires several environment variables to function properly:

NEXT_PUBLIC_API_URL - The URL where the API is hosted. This should point to your
production API server. Required for all environments. Typically looks like https://api.example.com

NEXT_PUBLIC_STRIPE_KEY - Your Stripe publishable key for handling payments. You can
find this in your Stripe dashboard under Settings. Required for payment features.

DATABASE_URL - The PostgreSQL connection string used to connect to the database.
Should be in the format: postgresql://user:password@host:port/database. Required.

[20 more environment variables with similar verbose descriptions...]
```

**Problem**: 200+ tokens to list what's already documented in `.env.local.example`.

### CONCISE (20 tokens)

```markdown
## Environment Variables

See `.env.local.example` for the complete list.
```

**Better**: References the source of truth.

---

## Example 3: Domain Model

### TOO VERBOSE (150+ tokens)

```markdown
## Domain Model

The system is built around users and organizations. Users are people who log into
the system. Each user has an email address for authentication, a name for display,
and a role that determines what they can do. Organizations are containers for
projects and represent teams or companies. An organization has a name, and every
organization is owned by one user. Organizations can have multiple users as members.

[More paragraph-based descriptions of relationships...]
```

**Problem**: Narrative description is hard to scan and understand relationships.

### CONCISE (50 tokens)

```markdown
## Domain Model

| Entity | Purpose | Key Fields |
|--------|---------|-----------|
| User | Authenticated user | id, email, role (owner/admin/member/viewer) |
| Organization | Team container | id, name, ownerId, createdAt |
| Project | Org sub-resource | id, orgId, name, description |

Relationships: User owns Organization; Organization has many Projects
```

**Better**: Table format is scannable and relationships are clear.

---

## Condensing Techniques

### Technique 1: Point to Authoritative Sources

**Instead of**:
```markdown
## Import Paths

Components: import { Button } from '@/components/ui/button'
Utils: import { formatDate } from '@/lib/utils'
Types: import type { User } from '@/lib/types'
...
```

**Write**:
```markdown
## Import Paths

Use `@/` prefix for all src imports (path alias configured in tsconfig.json)
```

### Technique 2: One-Line Summaries for Patterns

**Instead of**:
```markdown
## Database Operations

To query the database, you use Drizzle ORM. Drizzle provides a type-safe query builder...
[Details on how to use Drizzle]
```

**Write**:
```markdown
## Database Operations

Use Drizzle ORM (configured at `src/lib/db`) for all queries - provides type safety and migrations
```

### Technique 3: Tables Over Paragraphs

**Instead of**:
```markdown
## Roles

Admin can do everything. Member can see projects and create tasks but can't change
settings. Viewer can only read projects and tasks. Owner is the person who created
the organization and can delete it.
```

**Write**:
```markdown
## Roles

| Role | Create | Edit | Delete | Admin |
|------|--------|------|--------|-------|
| Owner | Y | Y | Y | Y |
| Admin | Y | Y | Y | Y |
| Member | Y | Own | Own | N |
| Viewer | N | N | N | N |
```

### Technique 4: Hierarchy for Complex Topics

**Instead of mixing everything**:
```markdown
## Authentication

Users log in with email and password. We use... [mixing theory and implementation]
```

**Use clear sections**:
```markdown
## Authentication

- **Login flow**: Email + password -> NextAuth session
- **Protected routes**: Use `withAuth` middleware
- **User context**: Injected via `useAuth()` hook
```

### Technique 5: Remove Redundancy

**Instead of repeating in multiple files**:
- Root CLAUDE.md: Domain model and rules
- Feature CLAUDE.md: Domain model again (for context)

**Keep only in root**:
- Root CLAUDE.md: Domain model (source of truth)
- Feature CLAUDE.md: Just reference it: "See root CLAUDE.md for domain model"

### Technique 6: Challenge Every Example

**Instead of**:
```markdown
## Component Props

Here's an example component:

export interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  variant?: 'primary' | 'secondary';
  size?: 'sm' | 'md' | 'lg';
}
```

**Ask**: Is this project-specific? If it's a shadcn component, Claude knows this. If it's custom project logic, only document the custom parts.

---

## Checklist for Conciseness

Before finalizing CLAUDE.md, audit each section:

- [ ] Does this explain basic programming? THEN Delete
- [ ] Does this duplicate library documentation? THEN Delete or link to official source
- [ ] Does this explain how (standard way)? THEN Delete
- [ ] Does this explain what's different about THIS project? THEN Keep and condense
- [ ] Can this be replaced with a one-liner? THEN Replace
- [ ] Can this be replaced with a table? THEN Replace
- [ ] Does this point outward to a source of truth? THEN Good
- [ ] Would Claude ask for this if it was unclear? THEN Keep
