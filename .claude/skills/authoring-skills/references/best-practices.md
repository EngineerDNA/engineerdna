# Best Practices for Authoring Skills

## Conciseness: Your Most Important Principle

**Default assumption**: Claude is already very smart. Only add context Claude doesn't already have.

### Token Budget

- At startup, only metadata (name/description) is pre-loaded
- SKILL.md loads when relevant, but every token competes with conversation history
- **Target**: Keep SKILL.md under 50 lines total

### Concise Code Examples

**GOOD (50 tokens)**:
```markdown
## Extract PDF text

Use pdfplumber:

```python
import pdfplumber
with pdfplumber.open("file.pdf") as pdf:
    text = pdf.pages[0].extract_text()
```
```

**BAD (150 tokens)**:
```markdown
PDF (Portable Document Format) files are a common file format...
There are many libraries available for PDF processing...
First, you'll need to install it using pip...
```

## Effective Descriptions Drive Discovery

**CRITICAL**: Description is your gateway to being useful. Include BOTH what the Skill does AND when to use it.

### Pattern: [What] + [When to use]

**GOOD**:
```yaml
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
```

**BAD**:
```yaml
description: Helps with documents
```

### Voice: Always Write in Third Person

(These are injected into system prompt)

- GOOD: "Processes Excel files and generates reports"
- AVOID: "I can help you process Excel files"

## Naming Conventions

**Recommended (gerund form)**:
- `processing-pdfs`
- `analyzing-spreadsheets`
- `managing-databases`

**Avoid**:
- Vague: `helper`, `utils`, `tools`
- Generic: `documents`, `data`, `files`

## What NOT to Include

1. **Basic programming knowledge** - Claude knows this
2. **Library documentation** - Claude knows common libraries (use Context7 for advanced docs)
3. **Time-sensitive information** - Create "old patterns" reference section instead
4. **Windows-style paths** - Always use forward slashes
5. **Multiple options for the same task** - Provide a default with escape hatch

## Critical DON'Ts

1. **Never assume packages are installed** - List requirements explicitly
2. **Never use deeply nested references** - Keep one level deep from SKILL.md
3. **Never offer too many options** - Provide default + escape hatch
4. **Never exceed 500 lines in SKILL.md** - Split into separate files with references/

## Maintenance: Keep It Fresh

When updating Skills:

1. **Remove verbose explanations** Claude already knows
2. **Condense code examples** to essentials only
3. **Use progressive disclosure** for lengthy content (move to references/)
4. **Keep description updated** with trigger keywords as scope evolves

## Testing Your Skill

1. **Make description specific** - Include keywords users would naturally mention
2. **Test discovery** - Ask Claude relevant questions about your domain, verify Skill activates
3. **Check YAML** - Run `head -10 SKILL.md` to verify frontmatter is valid
4. **Verify tool access** - Ensure `allowed-tools` matches what the skill needs
