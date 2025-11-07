---
name: writing-agents
description: Create specialized AI subagents with focused expertise and separate context. Use when creating or updating subagent definitions in .claude/agents/.
allowed-tools: Read, Bash, Grep, Glob, Edit
---

# Writing Subagents

Create focused AI subagents for task-specific workflows with separate context windows and configured tool access.

## Overview

Subagents are pre-configured agent files (`.claude/agents/*.md`) defining specific expertise, trigger phrases, tool boundaries, and system prompts.

## Quick Start

Save as `.claude/agents/agent-name.md`:

```markdown
---
name: agent-name
description: Purpose and MUST BE USED trigger phrases
allowed-tools: Read, Grep, Bash  # Optional
---

You are a [role].

When invoked:
1. First action
2. Second action

Include checklists as needed.
```

Test via explicit invocation or auto-delegation with trigger keywords.

## References

- [Agent File Format](references/agent-file-format.md) - Directory structure and frontmatter fields
- [Writing Descriptions](references/writing-descriptions.md) - Trigger phrases and effective description patterns
- [System Prompt Guide](references/system-prompt-guide.md) - Best practices for prompts and checklists
- [Tool Configuration](references/tool-configuration.md) - Tool access strategies and security
- [Model Selection](references/model-selection.md) - Choosing the right model for your agent
- [Example Agents](references/example-agents.md) - Code reviewer and debugger templates
- [Best Practices & Don'ts](references/best-practices-and-donts.md) - Patterns to follow and avoid
- [Testing & Maintenance](references/testing-and-maintenance.md) - Testing agents and iterative improvement
