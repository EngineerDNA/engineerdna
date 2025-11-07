# context7 - Library Documentation Reference

## Overview

context7 provides up-to-date documentation for npm packages and libraries. Use before implementing features to verify correct API usage.

## Available Tools

### mcp__context7__resolve-library-id

Finds the library ID for a given package name.

```typescript
mcp__context7__resolve-library-id({
  name: "react"
})
// Returns: library ID for use in get-library-docs
```

### mcp__context7__get-library-docs

Retrieves documentation for a specific library.

```typescript
mcp__context7__get-library-docs({
  libraryId: "react",  // From resolve-library-id
  query: "useState hook"  // Optional: specific topic
})
// Returns: Documentation content
```

## Usage Pattern

```typescript
// 1. Resolve library ID
const libraryId = mcp__context7__resolve-library-id({ name: "drizzle-orm" });

// 2. Get documentation
const docs = mcp__context7__get-library-docs({
  libraryId: libraryId,
  query: "PostgreSQL schema definition"
});

// 3. Use documentation in implementation
// Now implement with correct API usage
```

## Common Use Cases

### Before Implementing New Feature

```
User: "Add Stripe payment processing"

1. resolve-library-id for "stripe"
2. get-library-docs for payment intents API
3. Implement with correct Stripe patterns
```

### When Fixing Library Errors

```
User: "Drizzle query not working"

1. resolve-library-id for "drizzle-orm"
2. get-library-docs for query API
3. Compare with current usage
4. Fix incorrect usage
```

### Checking Compatibility

```
User: "Can we use Next.js 15?"

1. resolve-library-id for "next"
2. get-library-docs for version 15 features
3. Check breaking changes
4. Recommend migration path
```

## Best Practices

1. **Always resolve before getting docs** - Need library ID first
2. **Be specific in queries** - "useState hook" not "hooks"
3. **Check version compatibility** - Docs reflect latest version
4. **Use before implementing** - Verify API first

## Troubleshooting

### Problem: Library not found
**Solution**: Check exact package name on npm

### Problem: Docs outdated
**Solution**: Check package version, may need manual verification

## Performance

- **Speed**: Sub-second documentation retrieval
- **Caching**: Results cached for session
- **Recommendation**: Use liberally for verification
