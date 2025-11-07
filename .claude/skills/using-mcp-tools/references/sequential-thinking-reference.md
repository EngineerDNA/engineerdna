# sequential-thinking - Complex Reasoning Reference

## Overview

sequential-thinking enables structured thinking through multi-step problems with revision and branching support.

## When to Use

- Multi-step problem solving
- Architecture decisions requiring trade-off analysis
- Debugging complex issues
- Planning that needs course correction
- Problems where full scope isn't clear initially

## Tool Signature

```typescript
mcp__sequential-thinking__sequentialthinking({
  thought: "Current thinking step",
  nextThoughtNeeded: true,
  thoughtNumber: 1,
  totalThoughts: 5,
  isRevision: false,  // Set true if revising previous thought
  revisesThought: null,  // Which thought number being revised
  branchFromThought: null,  // If branching to explore alternative
  branchId: null,  // Branch identifier
  needsMoreThoughts: false  // If realizing more thoughts needed
})
```

## Key Features

- **Adjustable scope**: Can increase/decrease totalThoughts as understanding evolves
- **Revision support**: Can question or revise previous thoughts
- **Branching**: Can explore alternative approaches
- **Hypothesis testing**: Generate and verify solution hypotheses
- **Iterative**: Repeat until satisfied with answer

## Usage Pattern

```typescript
// Initial thought
sequentialthinking({
  thought: "User wants to add pause/resume for change requests. Need to understand current state machine.",
  nextThoughtNeeded: true,
  thoughtNumber: 1,
  totalThoughts: 5
})

// Follow-up thought
sequentialthinking({
  thought: "State machine has 'planning' and 'coding' states. Pause should work from both.",
  nextThoughtNeeded: true,
  thoughtNumber: 2,
  totalThoughts: 5
})

// Revision if needed
sequentialthinking({
  thought: "Wait, 'coding' state might not be pausable - workflow might be running. Need to check Temporal workflow pause capabilities.",
  nextThoughtNeeded: true,
  thoughtNumber: 3,
  totalThoughts: 6,  // Increased estimate
  isRevision: true,
  revisesThought: 2
})

// Continue until done
sequentialthinking({
  thought: "Solution: Pause by cancelling Temporal workflow and storing checkpoint. Resume by restarting from checkpoint.",
  nextThoughtNeeded: false,  // Done
  thoughtNumber: 6,
  totalThoughts: 6
})
```

## Common Use Cases

### Architecture Decisions

```
User: "Should we use Temporal Cloud or local workers?"

Use sequential-thinking to:
1. List requirements (reliability, cost, maintenance)
2. Evaluate Temporal Cloud (pros/cons)
3. Evaluate local workers (pros/cons)
4. Compare against requirements
5. Make recommendation with rationale
```

### Complex Debugging

```
User: "Deployments getting stuck at 45 minutes"

Use sequential-thinking to:
1. Understand normal deployment flow
2. Identify where timeout occurs
3. Hypothesis: Workflow not sending heartbeat
4. Check workflow heartbeat code
5. Verify hypothesis against logs
6. Propose fix
```

### Planning with Unknowns

```
User: "Add real-time collaboration"

Use sequential-thinking to:
1. Define collaboration requirements
2. Explore WebSocket vs polling
3. Realize need for conflict resolution
4. Branch: Explore operational transform
5. Branch: Explore CRDTs
6. Compare approaches
7. Select based on complexity/fit
```

## Best Practices

1. **Start with estimate** - Adjust totalThoughts as needed
2. **Revise when needed** - Don't hesitate to revisit thoughts
3. **Be explicit** - Clear thought progression helps
4. **Generate hypotheses** - Then verify them
5. **Express uncertainty** - Branch to explore alternatives

## Troubleshooting

### Problem: Too many thoughts
**Solution**: Reduce scope, focus on key decisions

### Problem: Circular reasoning
**Solution**: Use revision feature to break cycle

## Performance

- **Speed**: Medium - Each thought adds tokens/time
- **Token cost**: Full thought chain included in context
- **Recommendation**: Use for complex decisions, not simple tasks
