# Testing and Maintenance

How to test agents after creation and maintain them over time.

## Testing Agents

### 1. Explicit Invocation Test

Test by explicitly asking Claude Code to use the agent.

**Procedure:**
```
1. Create agent file in .claude/agents/agent-name.md
2. In Claude Code chat, ask: "Use the agent-name agent to [task]"
3. Verify agent activates and shows correct output
4. Check that the agent follows its instructions
5. Verify tool access is working as expected
```

**Example:**
```
Me: Use the code-reviewer agent to check my recent changes.

Expected: Agent activates and shows:
- "I'm the code-reviewer agent"
- Runs git diff
- Shows code review findings
```

**What to check:**
- Agent initializes with correct name
- System prompt instructions are followed
- Output format matches the prompt
- Tools are used correctly
- Response is specific to the task

### 2. Auto-Delegation Test

Test that the agent activates when trigger keywords are used.

**Procedure:**
```
1. Create agent with trigger keywords in description
2. In Claude Code chat, use a trigger keyword naturally: "Can you review this?"
3. Verify agent auto-activates without explicit invocation
4. Check that behavior matches when explicitly invoked
```

**Example:**
```
Agent description includes: "MUST BE USED for code review, audit, sanity check"

Me: Can you do a code review of my changes?

Expected: Agent auto-activates even though I didn't explicitly invoke it
```

**What to check:**
- Trigger keywords are recognized
- Agent activates without explicit request
- Agent behavior is consistent
- No false positives (agent shouldn't activate for unrelated keywords)

### 3. Tool Access Verification

Verify the agent can access intended tools and cannot access restricted tools.

**Procedure:**
```
1. Ask agent to use a tool it should have access to
2. Verify tool works (e.g., "Show me file X")
3. Ask agent to use a tool it shouldn't have access to
4. Verify tool is blocked or agent doesn't attempt it
```

**Example:**
```
Agent: allowed-tools: Read, Grep, Glob, Bash
(no Edit or Write)

Test 1 - Can read files:
Me: "Read /path/to/file.ts"
Expected: Agent reads and shows content (success)

Test 2 - Cannot edit files:
Me: "Fix the bug in /path/to/file.ts"
Expected: Agent explains it can't edit but suggests the fix (success)
```

**What to check:**
- Intended tools are accessible
- Restricted tools are not used
- Agent gracefully handles tool limitations
- Tool access matches frontmatter

### 4. Model Behavior Test

Verify the agent uses the correct model and behaves as expected.

**Procedure:**
```
1. Check agent's model setting (sonnet, opus, haiku, inherit)
2. Evaluate agent response quality and speed
3. Compare against your expectations for that model
```

**What to check:**
- Model matches frontmatter setting
- Response quality matches model capability
- Response latency matches model profile
- Complex tasks have appropriate model
- Simple tasks use economical model

### 5. Integration Test

Test that the agent works well with the broader agent ecosystem.

**Procedure:**
```
1. Create a multi-agent workflow
2. Have one agent delegate to another
3. Verify smooth handoff and information passing
4. Check that agents don't conflict
```

**Example:**
```
Me: "Can you review my database schema changes?"

Expected flow:
1. Code-reviewer agent activates
2. Asks db-architect to review schema
3. Db-architect provides database-specific feedback
4. Code-reviewer incorporates feedback
```

**What to check:**
- Agents can reference each other
- Information flows between agents
- No naming conflicts
- Complementary agents work well together

## Testing Checklist

Before marking an agent as complete:

- [ ] Explicit invocation works correctly
- [ ] Auto-delegation with trigger keywords works
- [ ] Tool access is appropriate and functional
- [ ] System prompt instructions are followed
- [ ] Output format matches specifications
- [ ] Model choice is appropriate
- [ ] No conflicts with existing agents
- [ ] Agent gracefully handles edge cases
- [ ] Documentation is clear and accurate

## Maintenance and Improvement

### 1. Monitor Agent Performance

Track how the agent actually behaves in use:

**Metrics to track:**
- How often the agent is invoked
- Whether auto-delegation triggers correctly
- User satisfaction with agent responses
- Issues or unexpected behaviors
- Tool access patterns

**Where to track:**
- Note issues when they occur
- Keep a log of improvements needed
- Collect feedback from team members

### 2. Refine Trigger Keywords

Add new keywords as you discover patterns:

**Initial description:**
```yaml
description: Debugging expert. Use proactively when encountering errors, test failures, unexpected behavior.
```

**After discovering common uses:**
```yaml
description: Debugging expert. Use proactively for errors, test failures, unexpected behavior, stack traces, null pointer exceptions, race conditions, integration issues.
```

**Process:**
1. Notice a new situation where you want the agent
2. Use a keyword you think is clear
3. If agent doesn't activate, add that keyword to description
4. Update agent file and test
5. Commit updated description

### 3. Improve System Prompt

Refine the system prompt based on actual behavior:

**When to improve:**
- Agent is unclear or vague in responses
- Agent misses important aspects
- Agent doesn't follow workflow steps
- Agent uses tools incorrectly
- Agent is too verbose or too brief

**How to improve:**
1. Identify specific issue with agent behavior
2. Add clarification to system prompt
3. Test revised agent
4. Commit improved version

**Example improvement:**

Original prompt:
```markdown
When invoked, review the code.
```

Improved prompt:
```markdown
When invoked:
1. Run git diff to see what changed
2. Focus on modified files only (ignore unchanged code)
3. Apply the review checklist
4. Report critical issues separately from suggestions
5. Provide specific code references for each issue
```

### 4. Update Tool Access

Adjust tool list as agent needs evolve:

**When to add tools:**
- Agent says "I don't have access to X tool"
- Agent needs to use a new tool for its work
- Tool becomes necessary for improved workflow

**When to remove tools:**
- Agent never uses a particular tool
- Tool access creates security concerns
- Tool is no longer necessary for agent's purpose

**Process:**
1. Identify tool needs
2. Update allowed-tools in frontmatter
3. Test agent with new/removed tools
4. Commit updated agent
5. Document reason for change in commit message

### 5. Adjust Model

Change model based on performance and cost:

**When to upgrade to Opus:**
- Agent often needs complex reasoning
- Agent gives incomplete responses
- Quality issues persist despite prompt improvements

**When to downgrade to Haiku:**
- Agent is always fast (could be haiku)
- Agent's tasks are simple or formulaic
- Cost optimization is important

**Process:**
1. Identify performance or cost issue
2. Test with new model
3. Compare quality and cost
4. Update model in frontmatter if beneficial
5. Monitor real-world usage

### 6. Version Control Updates

Keep agent improvements in git history:

**Good commit messages:**
```bash
git commit -m "refactor(agents): improve code-reviewer description with specific trigger keywords"
git commit -m "feat(agents): add Read/Edit tool access to debugger agent"
git commit -m "docs(agents): clarify workflow steps in code-reviewer system prompt"
git commit -m "perf(agents): switch architecture-advisor to opus model for complex analysis"
```

**Process:**
1. Make improvement to agent file
2. Test thoroughly
3. Commit with descriptive message
4. Include reasoning in commit message if significant
5. Share improvements with team

## Deprecating Agents

Sometimes agents become obsolete:

**Signs an agent should be deprecated:**
- No one uses it (track invocation frequency)
- Another agent does the same thing better
- Functionality moved to commands or skills
- Team consensus to retire it

**Deprecation process:**

1. **Announce** - Tell team the agent is being retired
2. **Document** - Add deprecation notice to agent
3. **Wait** - Give people time to stop using it
4. **Remove** - Delete from .claude/agents/
5. **Commit** - Record removal in git with clear message

**Example deprecation:**
```markdown
---
name: old-agent
description: DEPRECATED - Use new-agent instead
---

This agent is deprecated. Use the new-agent agent instead.
```

After deprecation period:
```bash
rm .claude/agents/old-agent.md
git commit -m "refactor: deprecate old-agent in favor of new-agent"
```

## Agent Evolution Over Time

### Phase 1: Initial Creation
- Created with basic structure
- Tested for basic functionality
- Simple trigger keywords
- Minimal tool access

### Phase 2: Early Usage
- Trigger keywords refined based on use cases
- System prompt clarified based on feedback
- Tool access adjusted as needed
- Team learns how to best use agent

### Phase 3: Optimization
- Well-tuned trigger keywords
- Detailed, specific system prompt
- Perfect tool access for the task
- Consistent behavior and quality

### Phase 4: Specialization
- Agent refined for specific use cases
- May have multiple specialized versions
- Clear documentation
- Deep integration with other agents

### Phase 5: Retirement
- Functionality stabilizes
- Less frequent improvements
- Eventually may be superseded
- Documented for historical reference or removed

## Common Maintenance Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| Agent not triggering | Need explicit invocation every time | Add more trigger keywords to description |
| Agent too slow | Laggy responses | Consider switching to haiku model |
| Agent too simple | Incomplete responses | Upgrade model to opus |
| Agent vague | Unclear output | Rewrite system prompt with more detail |
| Tool errors | Can't complete tasks | Check allowed-tools and add missing tools |
| Conflicts with other agents | Both activate for same task | Refine descriptions to differentiate |
| Team unaware | No one uses it | Better documentation or examples |

## Documentation for Maintenance

Keep agents well-documented:

**In the agent file:**
- Clear description with trigger keywords
- Detailed system prompt with workflow
- Tool list with reasoning
- Examples in comments if complex

**In project README:**
- List all available agents
- Describe what each does
- Provide usage examples
- Link to agent files

**In CLAUDE.md:**
- Document agent conventions
- List approved agents
- Describe how to create new agents
- Link to this skill

## Continuous Improvement Workflow

1. **Create agent** with best practices
2. **Test thoroughly** before team use
3. **Monitor usage** and collect feedback
4. **Identify improvements** over time
5. **Implement improvements** incrementally
6. **Document changes** in git history
7. **Share improvements** with team
8. **Repeat** as agent evolves

This iterative approach ensures agents improve over time and stay valuable to the team.
