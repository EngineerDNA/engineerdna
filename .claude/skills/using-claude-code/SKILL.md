---
name: using-claude-code
description: Use when creating Skills, Hooks, Commands, or Agents in Claude Code, or questions about proper patterns and structure
tools: Read
---

# Using Claude Code

Best practices for authoring Claude Code components following established patterns.

## When to Use

- Creating or modifying Skills, Hooks, Commands, or Agents
- Questions about proper structure or patterns
- Understanding when to use which component type
- Following progressive disclosure and conciseness principles

## Quick Decision

| Need | Component | Reference |
|------|-----------|-----------|
| Repeated prompt | Command | commands-authoring.md |
| Complex workflow | Skill | skills-authoring.md |
| Enforce rule | Hook | hooks-authoring.md |
| Specialized AI | Agent | agents-authoring.md |

## Available Guides

**Component Authoring**:
- **[Skills Authoring](reference/skills-authoring.md)** - Progressive disclosure, when to use
- **[Hooks Authoring](reference/hooks-authoring.md)** - Solve don't punt, security
- **[Commands Authoring](reference/commands-authoring.md)** - When to use vs skills
- **[Agents Authoring](reference/agents-authoring.md)** - One task principle, context

**Quality & Automation**:
- **[Reducing Hallucinations](reference/reducing-hallucinations.md)** - Grounding, verification
- **[Automatic Workflows](reference/automatic-workflows.md)** - Auto-activation, build checks
- **[Skill Troubleshooting](reference/skill-troubleshooting.md)** - Diagnosis, maintenance

**Decision Reference**:
- **[Decision Guide](reference/decision-guide.md)** - When to use each component type
