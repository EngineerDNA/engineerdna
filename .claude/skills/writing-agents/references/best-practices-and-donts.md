# Best Practices and Critical Don'Ts

Patterns to follow and avoid when creating and maintaining agents.

## Best Practices

### 1. Start with Claude-Generated Agents

Use Claude to generate the initial agent:

**Good workflow:**
```
1. Ask Claude: "Generate an agent for [purpose]"
2. Claude creates the initial agent.md
3. Save the agent file to .claude/agents/
4. Test it in practice
5. Refine based on actual behavior
```

**Benefits:**
- Claude understands good agent structure
- Initial version is well-formed
- Saves time writing frontmatter
- Better prompts from the start

### 2. Single Responsibility Per Agent

Each agent should have one clear focus:

**Good:** Specialized agents
```markdown
name: code-reviewer    # Only reviews code
name: debugger         # Only debugs issues
name: db-architect     # Only designs schemas
```

**Bad:** Multi-purpose agents
```markdown
name: ai-helper        # Does everything (unfocused)
name: expert           # Too vague (lacks clear purpose)
name: fixer            # Unclear scope (could be code, infrastructure, etc.)
```

**Why:** Focused agents are better at their task, more predictable, and easier to reason about.

### 3. Write Detailed, Specific Prompts

Agent prompts should be detailed and specific:

**Good:** Detailed prompt
```markdown
You are a senior code reviewer ensuring code meets our standards.

When invoked:
1. Run git diff to see changes
2. Focus on modified files
3. Apply the review checklist

Review checklist:
- No exposed secrets
- Error handling is complete
- Types are non-nullable where possible
- No TODO comments left behind
```

**Bad:** Vague prompt
```markdown
You are a helpful code reviewer.
```

**Why:** Specific prompts guide agent behavior and prevent errors.

### 4. Limit Tool Access

Only grant the tools an agent actually needs:

**Good:** Restricted access
```yaml
name: reviewer
allowed-tools: Read, Grep, Glob, Bash
# Can't edit, so can't accidentally modify code
```

**Bad:** Unrestricted access
```yaml
name: reviewer
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
# Can modify anything - why would a reviewer need Edit?
```

**Why:** Limited tool access improves:
- Security (agent can't do more than intended)
- Focus (fewer options means clearer choices)
- Predictability (you know what it can do)

### 5. Version Control Project Agents

Check project agents into git:

**Good:**
```bash
# Commit agents to git
git add .claude/agents/*.md
git commit -m "feat: add code-reviewer agent"
git push
```

**Bad:**
```bash
# Keeping agents out of version control
# (other team members can't see/use them)
```

**Why:** Agents are team infrastructure - should be versioned and shared.

### 6. Include Trigger Phrases in Descriptions

Descriptions should signal when to use the agent:

**Good:** Clear triggers
```yaml
description: Code review expert. MUST BE USED for review, audit, check, security assessment, sanity check.
```

**Bad:** No triggers
```yaml
description: Reviews code
```

**Why:** Claude Code uses description keywords for auto-delegation.

### 7. Test Auto-Delegation

Verify agents activate when expected:

**Good workflow:**
```
1. Create agent with trigger keywords
2. Use a keyword in conversation: "Can you review this?"
3. Agent should activate automatically
4. Verify behavior matches expectations
```

**Benefits:**
- Agents help without explicit invocation
- Saves typing
- Ensures keywords are effective

## Critical Don'Ts

### DON'T: Create Multi-Purpose Agents

**WRONG:**
```markdown
---
name: ai-helper
description: Helps with various tasks
---

You are a helpful AI assistant. Help with code, debugging, documentation, tests, and anything else.
```

**Why it fails:**
- No clear focus
- Agent doesn't know when to specialize
- Conflicts with other agents
- Hard to test and improve
- Triggers not specific enough

**RIGHT:**
```markdown
---
name: code-reviewer
description: Expert code review. MUST BE USED for review and verification.
allowed-tools: Read, Grep, Glob, Bash
---

You are a senior code reviewer.
```

### DON'T: Omit When-Invoked Instructions

**WRONG:**
```markdown
You are a senior debugger.

Guidelines:
- Check error handling
- Verify types
```

Agent doesn't know what to do when invoked!

**RIGHT:**
```markdown
You are a senior debugger.

When invoked:
1. Capture error message
2. Reproduce the issue
3. Identify root cause
4. Implement fix

Guidelines:
- Check error handling
- Verify types
```

**Why:** Agents need clear workflow steps.

### DON'T: Grant Unnecessary Tools

**WRONG:**
```yaml
name: reviewer
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
```

Why would a reviewer need Edit and Write?

**RIGHT:**
```yaml
name: reviewer
allowed-tools: Read, Grep, Glob, Bash
```

**Why:**
- Security (can't make unintended changes)
- Focus (clear scope)
- Safety (can't break things)

### DON'T: Use Vague Descriptions

**WRONG:**
```yaml
description: A helpful assistant
```

No one knows when to use this!

**RIGHT:**
```yaml
description: Database architect. Use proactively for schema design, migrations, and performance optimization.
```

**Why:** Good descriptions enable auto-delegation.

### DON'T: Create Agents Without Testing

**WRONG:**
```
1. Write agent
2. Commit to git
3. Share with team
# (never tested)
```

**RIGHT:**
```
1. Write agent
2. Test explicit invocation
3. Test auto-delegation with keywords
4. Refine based on behavior
5. Commit to git
```

**Why:** Testing reveals issues before team adoption.

### DON'T: Duplicate Agent Names

**WRONG:**
```
.claude/agents/
  reviewer.md       # Project reviewer
  reviewer.md       # User reviewer (duplicate!)
```

Confuses tool lookup and creates conflicts.

**RIGHT:**
```
.claude/agents/
  code-reviewer.md           # Code reviews
  documentation-reviewer.md  # Docs reviews
```

Or use hyphens for clarity:
```
.claude/agents/
  code-reviewer.md
  schema-reviewer.md
```

### DON'T: Use Model Incorrectly

**WRONG:**
```yaml
name: quick-checker
model: opus  # Wrong - opus is expensive for quick checks
```

**RIGHT:**
```yaml
name: quick-checker
model: haiku  # Fast and cheap for simple tasks
```

Or for complex analysis:
```yaml
name: architect
model: opus  # Worth the cost for complex decisions
```

### DON'T: Forget Context for Remote Use

**WRONG:** Agent assumes specific project structure
```markdown
When invoked:
1. Check the schema in database/schema.ts
2. Look at migrations in scripts/migrate.js
```

Won't work if project structure differs!

**RIGHT:** Use discoverable patterns
```markdown
When invoked:
1. Find schema files (typically src/lib/db/schema/)
2. Look for migration files (typically drizzle/)
3. Adapt if your project structure differs
```

### DON'T: Create Agents Without Documentation

**WRONG:** Agent with no comments or guidance
```markdown
---
name: x
description: something
---

# Agent content with no explanation
```

**RIGHT:** Well-documented agent
```markdown
---
name: code-reviewer
description: Expert code review specialist. MUST BE USED for review and verification.
allowed-tools: Read, Grep, Glob, Bash
---

You are a senior code reviewer...
(clear, detailed prompt with examples)
```

## Agent Lifecycle

### Creation Phase
- [x] Start with Claude-generated version
- [x] Add specific project context
- [x] Include detailed workflow steps
- [x] Define clear tool scope

### Testing Phase
- [x] Test explicit invocation
- [x] Test auto-delegation with keywords
- [x] Verify tool access is appropriate
- [x] Check for unexpected behavior

### Deployment Phase
- [x] Version control agent file
- [x] Share with team
- [x] Document agent in project README
- [x] Add usage examples

### Maintenance Phase
- [x] Monitor agent behavior
- [x] Refine based on real usage
- [x] Update trigger keywords if needed
- [x] Improve based on feedback

## Common Mistakes and Fixes

| Mistake | Problem | Fix |
|---------|---------|-----|
| Vague description | Won't auto-delegate | Add specific trigger keywords |
| No when-invoked section | Agent doesn't know what to do | Add step-by-step instructions |
| Too many tools | Unfocused, security risk | Restrict to necessary tools only |
| Multi-purpose agent | No clear focus | Split into focused agents |
| No testing | Broken behavior in production | Test before committing |
| Unused agents | Clutter | Delete agents not in active use |
| Conflicting agents | Confusing behavior | Clarify purposes and triggers |
| No documentation | Team doesn't know about agent | Add to README and examples |

## Review Checklist for New Agents

Before committing an agent, verify:

- [ ] Agent has a single, clear focus
- [ ] Description includes trigger phrases ("MUST BE USED", "proactively", etc.)
- [ ] When-invoked section provides step-by-step workflow
- [ ] Tool access is minimal and appropriate
- [ ] Model choice matches agent complexity
- [ ] System prompt is detailed and specific
- [ ] Agent tested with explicit invocation
- [ ] Agent tested with auto-delegation keywords
- [ ] No conflicting agents with similar names
- [ ] Agent documented in project README
