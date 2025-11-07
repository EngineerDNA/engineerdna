# Claude Code Component Decision Guide

## Quick Decision Guide

| Need | Use | See |
|------|-----|-----|
| Repeated simple prompt | Command | [commands-authoring.md](commands-authoring.md) |
| Complex multi-step workflow | Skill | [skills-authoring.md](skills-authoring.md) |
| Enforce critical rule | Hook | [hooks-authoring.md](hooks-authoring.md) |
| Specialized AI task | Agent | [agents-authoring.md](agents-authoring.md) |

## When to Use Each Component

### Commands
**Use for:**
- Repeated prompts you type often (pre-check, post-check, deploy)
- Simple operations without complex logic
- Quick reference reminders

**Don't use for:**
- Multi-file workflows requiring progressive disclosure
- Complex decision trees
- Tasks requiring extensive documentation

### Skills
**Use for:**
- Complex multi-step workflows
- Operations requiring reference documentation
- Procedures with multiple alternatives
- Tasks needing progressive disclosure (keep main SKILL.md < 50 lines)

**Don't use for:**
- Simple one-line commands
- Operations better suited for hooks

### Hooks
**Use for:**
- Enforcing critical rules (security, validation)
- Automatic validation before/after operations
- Event-driven automation
- Blocking dangerous operations

**Don't use for:**
- Tasks requiring Claude's judgment
- Complex decision-making better left to agents

### Agents
**Use for:**
- Specialized tasks requiring context accumulation
- Multi-step processes with handoffs
- Domain-specific expertise (PM, designer, architect)
- Tasks requiring tool orchestration

**Don't use for:**
- Simple operations better suited for commands
- Validation better handled by hooks

## Core Principles (Apply to All Components)

1. **Conciseness** - Assume Claude is smart, document only what's unique to your project
2. **Progressive Disclosure** - Main file points to details, load only what's needed
3. **Verifiable** - Plan → validate → execute patterns for critical operations
4. **Secure** - Handle errors explicitly, validate inputs, quote variables

## Project-Specific Rules

When authoring any component, remember these project rules:

- **Rule 33**: NO one-off DB scripts, always use migrations
- **Rule 34**: UTC timestamps everywhere (use time.Now().UTC() in Go)
- **Rule 35**: NO aspirational code, delete unused immediately
- **Rule 21**: One task = one agent invocation, no todo lists to agents
