# Agent Architecture

## Agent Pipeline

For EngineerDNA development:

```
engineer → quality → integration-checker → security → ui-tester → docs → advisor
```

**Simplified pipelines:**
- Backend only: `engineer → quality → integration-checker → security → advisor`
- Frontend only: `engineer → quality → ui-tester → advisor`
- Docs only: `docs → advisor`

## Agent Principles

### 1. Single Responsibility
Each agent has ONE focused purpose:
- **engineer**: Implementation only (Go + React)
- **quality**: Testing and verification only
- **integration-checker**: Plugin system integration testing only
- **security**: Security review only
- **ui-tester**: Frontend/UI testing only
- **docs**: Documentation management only
- **advisor**: Final production readiness review only

### 2. Stateless Design
Agents don't remember previous iterations. Each invocation needs:
- Full context from previous agents
- Complete file list with changes
- All relevant success criteria
- Known issues or blockers

### 3. Evidence-Based Claims
**NO TOOL CALL = NO CLAIM**

Every claim must be backed by tool evidence:
```yaml
claim: "Tests pass"
tool: Bash
command: go test ./...
result: ok    github.com/...
```

### 4. Exhaustiveness Protocol
After finding ANY issue, search for ALL instances:
1. Find first occurrence
2. Extract searchable pattern  
3. Grep entire codebase
4. Document ALL instances
5. Fix/report ALL, not just first

### 5. Error Handling
After EVERY tool call:
1. Check result status
2. If error → STOP immediately
3. Investigate cause
4. Fix issue
5. Retry
6. Verify success before continuing

## Agent Handoff Context

### Engineer → Quality
```yaml
FILES_CHANGED:
  - path: what changed
TESTS_ADDED:
  - test: what it verifies
SECURITY_IMPLEMENTED:
  - organization filtering: where/how
```

### Quality → Integration-Checker
```yaml
TEST_RESULTS:
  - test: pass/fail
VERIFICATION_COMPLETED:
  - go vet: result
  - staticcheck: result
  - go test: result
```

### Integration-Checker → Security
```yaml
INTEGRATION_VERIFIED:
  - Plugin discovery: status
  - JSON-RPC: status
  - Encryption: status
PLUGINS_TESTED:
  - plugin: all methods verified
```

### Security → UI-Tester (if frontend changes)
```yaml
SECURITY_VERIFIED:
  - encryption: verified
  - validation: verified
FRONTEND_CHANGES:
  - files: list of changed files
```

### UI-Tester → Docs
```yaml
UI_VERIFIED:
  - Pages: tested
  - Components: rendering
  - Console: no errors
SCREENSHOTS:
  - page: screenshot file
```

### Docs → Advisor
```yaml
DOCUMENTATION_COMPLETE:
  - Files updated: list
  - CHANGELOG: updated
  - Accuracy verified: yes/no
```

## Tool Access

**engineer**: All tools (Read, Write, Edit, Bash, Grep, Glob)
**quality**: Read-only + Bash (Bash, Read, Grep, Glob)
**integration-checker**: Read-only + Bash (Bash, Read, Grep, Glob)
**security**: Read-only (Read, Grep, Glob)
**ui-tester**: Read-only + Bash + Chrome MCP (Bash, Read, Grep, Glob, chrome_*)
**docs**: Read + Write (Read, Write, Edit, Grep, Glob)
**advisor**: Read-only (Read, Grep, Glob)

## Invocation Patterns

### Automatic
Claude invokes agents based on task description keywords in agent frontmatter.

### Explicit
```
> Use the go-engineer agent to implement the plugin system
> Have the security agent review the encryption implementation
```

### One Task = One Invocation
Do NOT pass todo lists to agents. One task at a time.

**Wrong:**
> Use engineer agent to: 1) add table, 2) create API, 3) add tests

**Right:**
> Use engineer agent to add the events table with anonymization support
[After completion]
> Use engineer agent to create the API endpoint for event ingestion
[After completion]
> Use engineer agent to add tests for the event API

## Agent-Specific Notes

### engineer
- Implements working code (Go + React)
- Follows Go idioms and patterns
- Uses stdlib before external deps
- Enforces error handling
- NO aspirational code (Rule 29: YAGNI)

### quality
- Verifies tests pass
- Checks go vet, staticcheck, go test
- Ensures build succeeds
- CANNOT write code
- Returns to engineer if issues found

### integration-checker
- Tests plugin system integration
- Verifies JSON-RPC communication
- Checks plugin discovery
- Tests encryption/anonymization flows
- CANNOT write code
- Returns to engineer if integration fails

### security
- Reviews threat model compliance
- Validates encryption patterns
- Checks input validation
- Verifies no secrets in code
- CANNOT fix issues, only report

### ui-tester
- Tests frontend/UI functionality
- Uses Chrome MCP for browser testing
- Verifies components render correctly
- Checks console for errors
- Tests user flows
- CANNOT write code
- Returns to engineer if UI broken

### docs
- Creates and updates documentation
- Maintains README, CHANGELOG, guides
- Writes API documentation
- CAN write/edit markdown files only
- Returns to engineer if clarification needed

### advisor
- Final production readiness review
- Checks for regressions
- Verifies SOLID principles
- Fresh-eyes protocol for re-reviews
- APPROVED or BLOCKED decision only

## Common Mistakes

1. **Agent writes code when should only verify** → Use disallowedTools
2. **Agent continues after errors** → Enforce error handling protocol
3. **Agent reports first issue only** → Enforce exhaustiveness protocol
4. **Agent lacks context** → Previous agent must provide full handoff
5. **Multiple tasks in one invocation** → Split into sequential agent calls
