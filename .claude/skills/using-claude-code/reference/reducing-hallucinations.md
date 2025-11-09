# Reducing Hallucinations

Techniques to minimize Claude generating incorrect or inconsistent information.

## Allow "I Don't Know"

Give Claude explicit permission to admit uncertainty:

```markdown
As our M&A advisor, analyze this report on the potential acquisition.

Focus on financial projections, integration risks, and regulatory hurdles.

If you're unsure about any aspect or if the report lacks necessary information, say "I don't have enough information to confidently assess this."
```

This drastically reduces false information.

## Use Direct Quotes for Grounding

For long documents (>20K tokens), extract quotes first before analysis:

```markdown
Review this updated privacy policy for GDPR and CCPA compliance.

1. Extract exact quotes from the policy that are most relevant to GDPR and CCPA compliance. If you can't find relevant quotes, state "No relevant quotes found."

2. Use the quotes to analyze the compliance of these policy sections, referencing the quotes by number. Only base your analysis on the extracted quotes.
```

This grounds responses in actual text, reducing hallucinations.

## Verify with Citations

Make responses auditable by requiring citations:

```markdown
Draft a press release for our new cybersecurity product, AcmeSecurity Pro, using only information from these product briefs and market reports.

After drafting, review each claim in your press release. For each claim, find a direct quote from the documents that supports it. If you can't find a supporting quote for a claim, remove that claim from the press release and mark where it was removed with empty [] brackets.
```

## Chain-of-Thought Verification

Ask Claude to explain reasoning step-by-step before final answer:

```markdown
Analyze this error and explain your reasoning step-by-step before providing a fix.

1. What is the error message telling us?
2. What code is involved?
3. What could cause this error?
4. How can we verify this is the cause?
5. What is the fix?
```

This reveals faulty logic or assumptions.

## Best-of-N Verification

For critical decisions, run the same prompt multiple times and compare:

```bash
# Run same analysis 3 times
/analyze-security file.ts
/analyze-security file.ts
/analyze-security file.ts

# Compare outputs - inconsistencies indicate potential hallucinations
```

Use this for high-stakes code reviews or security audits.

## External Knowledge Restriction

Explicitly instruct Claude to only use provided information:

```markdown
Analyze this API using ONLY the information in the following documentation. Do not use any general knowledge about similar APIs.

<documentation>
{{API_DOCS}}
</documentation>

If the documentation doesn't contain information needed to answer, say "Documentation does not specify this."
```

## Iterative Refinement

Use Claude's outputs as inputs for follow-up prompts:

```markdown
# First prompt
Explain how the authentication system works.

# Follow-up prompt
You said the system uses JWT tokens. Find the code that creates these tokens and verify this is correct. If you can't find the code, revise your explanation.
```

This catches and corrects inconsistencies.

## Confidence Levels

Ask Claude to rate confidence:

```markdown
Analyze this bug and provide:
1. Your hypothesis (with confidence level: high/medium/low)
2. Evidence supporting your hypothesis
3. Alternative explanations if confidence is not high

Only proceed with a fix if confidence is high.
```

## Validation Loops

For critical operations, verify before executing:

```markdown
1. Propose the fix
2. Explain why this fix is correct
3. Identify what could go wrong
4. Verify the fix addresses all edge cases
5. Only then apply the fix
```

## Project-Specific Applications

### Database Operations
```markdown
Show me the migration needed for this schema change.

Extract the current schema definition from the code first. Then propose the migration. Then verify the migration matches the schema change exactly.

If you can't find the current schema, say so and do not generate a migration.
```

### Plugin System
```markdown
Explain how the plugin discovery system works.

First find and quote the plugin discovery code. Then explain how it works. If you can't find the implementation, say "I cannot locate the plugin discovery implementation."
```

### API Endpoints
```markdown
Show me how to call the /api/events endpoint.

Find the route definition and any documentation. Quote the relevant code. Then show the API call. If the route doesn't exist, say "This endpoint does not appear to be defined."
```

## Testing for Hallucinations

### Fact-checking Process
1. Ask Claude to state the facts
2. Ask Claude to find code/docs supporting each fact
3. If Claude can't find support, it retracts the fact

### Example
```markdown
You mentioned that projects use soft deletes. Find the code that implements this.

[If Claude can't find it]

Since you can't find the code, revise your explanation about how project deletion works.
```

## When to Apply These Techniques

**Always use for**:
- Security audits
- Database migrations
- Production deployments
- Regulatory compliance
- Financial calculations

**Consider using for**:
- Code reviews
- Architecture decisions
- Bug analysis
- API design

**Usually not needed for**:
- Simple file edits
- Documentation updates
- Code formatting
- Test writing

## Red Flags for Hallucinations

Watch for:
- Claude states facts without citing code
- Explanations contradict visible code
- Claude describes files that don't exist
- Responses change significantly when asked to verify
- Claude refuses to say "I don't know"

When you see these, apply verification techniques above.

## Combining Techniques

For maximum reliability, combine multiple approaches:

```markdown
Analyze this security vulnerability.

1. Allow uncertainty: Say "I don't know" if unsure
2. Use quotes: Extract relevant code snippets first
3. Chain of thought: Explain your reasoning step by step
4. Verification: Find evidence for each claim
5. Confidence: Rate your confidence in the analysis

Only proceed with a fix if confidence is high and evidence is strong.
```

This multi-layered approach minimizes hallucinations effectively.
