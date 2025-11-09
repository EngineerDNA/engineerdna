# Progressive Disclosure Patterns

Progressive disclosure means showing essential information first, with links to deeper content as needed. This keeps SKILL.md concise while supporting users who need details.

## Pattern 1: Quick Start + References

Use this when your skill has clear basic usage plus multiple advanced topics.

```markdown
---
name: processing-pdfs
description: Extract text and tables from PDF files. Use when working with PDFs or document extraction.
---

# PDF Processing Skill

Extract text, tables, and metadata from PDF files using pdfplumber.

## Quick Start

```bash
# Extract text from PDF
python -c "
import pdfplumber
with pdfplumber.open('document.pdf') as pdf:
    print(pdf.pages[0].extract_text())
"
```

## Common Uses

- Extracting tables from reports
- Converting PDFs to searchable text
- Bulk processing multiple files

## References

- [Text Extraction Guide](references/text-extraction.md) - Extract text, handle formatting
- [Table Extraction](references/tables.md) - Extract structured data
- [Form Filling](references/forms.md) - Fill PDF forms programmatically
- [Troubleshooting](references/troubleshooting.md) - Common issues and solutions
```

## Pattern 2: Domain-Specific Organization

Use this when your skill covers multiple related domains (e.g., different database systems, frameworks, or data types).

```
bigquery-skill/
  SKILL.md (overview + setup)
  references/
    finance.md
    sales.md
    product.md
    common-patterns.md
    troubleshooting.md
```

SKILL.md would have:
```markdown
---
name: querying-bigquery
description: Query and analyze data in Google BigQuery. Use for big data analysis, reporting, or data exploration.
---

# BigQuery Querying

Query and visualize data in Google BigQuery using SQL and Python.

## Quick Start

```python
from google.cloud import bigquery
client = bigquery.Client()
query = "SELECT COUNT(*) FROM `project.dataset.table`"
results = client.query(query).result()
```

## Domains

- [Finance Queries](references/finance.md) - Financial data analysis
- [Sales Reports](references/sales.md) - Sales metrics and trends
- [Product Analytics](references/product.md) - Usage and feature tracking

## References

- [Common Patterns](references/common-patterns.md)
- [Troubleshooting](references/troubleshooting.md)
```

## Pattern 3: Workflows with Checklists

Use this for complex multi-step processes.

### SKILL.md (Concise Overview)

```markdown
---
name: migrating-databases
description: Safe database schema evolution with Drizzle. Use when making schema changes.
---

# Safe Database Migrations

Master database schema evolution using Drizzle ORM.

## Quick Start: 6-Step Workflow

1. Modify schema (`src/lib/db/schema/`)
2. Generate migration (`pnpm db:generate`)
3. Review SQL in `drizzle/`
4. Test locally (`pnpm db:migrate`)
5. Test application (`pnpm dev`)
6. Commit both files together

See [Migration Workflow](references/workflow.md) for detailed steps.

## References

- [Workflow](references/workflow.md) - Step-by-step guide
- [Schema Patterns](references/patterns.md) - Common patterns
- [Troubleshooting](references/troubleshooting.md) - Recovery strategies
```

### references/workflow.md (Detailed Steps)

```markdown
# Migration Workflow

## Step 1: Modify Schema

Edit your schema file in `src/lib/db/schema/`:

```typescript
import { pgTable, serial, text, timestamp } from "drizzle-orm/pg-core";

export const users = pgTable("users", {
  id: serial("id").primaryKey(),
  name: text("name").notNull(),
  newField: text("new_field"),  // ← New column
  createdAt: timestamp("created_at").defaultNow(),
});
```

## Step 2: Generate Migration

```bash
pnpm db:generate
```

Drizzle creates a new file in `drizzle/` with the SQL.

## Step 3: Review SQL

Always review the generated SQL before applying:

```bash
cat drizzle/0001_migration_name.sql
```

...
```

## Pattern 4: Decision Trees

Use this when your skill needs to handle multiple scenarios with different solutions.

```markdown
---
name: finding-unused-code
description: Detect unused code and dependencies with knip. Use when optimizing bundle or auditing codebase.
---

# Dead Code Detection

Find unused code, dependencies, and exports using knip.

## Quick Start

```bash
pnpm run knip
```

## Is This a False Positive?

See [False Positives Guide](references/false-positives.md) to determine if knip is correct:

- **Next.js routes** - Often reported as unused
- **API endpoints** - Framework-registered routes
- **Exports from activities** - Used by workers
- **Exported types** - May be public API

## References

- [False Positives](references/false-positives.md) - How to identify them
- [Configuration](references/configuration.md) - Improve knip.json
- [Best Practices](references/best-practices.md) - Automation workflows
```

## Guidelines for All Patterns

### Keep SKILL.md Focused
- One main concept
- Quick Start section always
- 2-3 reference links
- <50 lines total

### Reference File Organization
- Name clearly: `workflow.md`, `patterns.md`, `troubleshooting.md`
- Keep flat (no nesting): `references/file.md` not `references/guides/file.md`
- Link directly from SKILL.md: `[text](references/file.md)`
- Add navigation within reference files (see also sections)

### When to Create References

- If SKILL.md exceeds 50 lines → move content to references/
- If you're explaining a concept for >10 lines → separate file
- If multiple related topics exist → group in one reference file
- If you have working code examples → create `examples.md` or task-specific file

### Cross-Linking Between References

In reference files, you can link to other references:

```markdown
See also [Troubleshooting](troubleshooting.md) for recovery strategies.
```

This helps users navigate deeper without overloading SKILL.md.
