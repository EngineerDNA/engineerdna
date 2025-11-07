# Hook Authoring Best Practices

## Core Principle: Solve, Don't Punt

Hooks should handle errors, not pass them to Claude.

**Bad Hook** (punts to Claude):
```python
def process(path):
    return open(path).read()  # Fails if missing
```

**Good Hook** (solves the problem):
```python
def process(path):
    try:
        with open(path) as f:
            return f.read()
    except FileNotFoundError:
        create_default(path)  # Handle it
        return ""
    except PermissionError:
        return ""  # Provide alternative
```

## Security Checklist

Every hook MUST:
- Quote shell variables: `"$VAR"` not `$VAR`
- Block path traversal: check for `..` in paths
- Use absolute paths: leverage `$CLAUDE_PROJECT_DIR`
- Skip sensitive files: `.env`, `.git/`, keys, credentials

## Hook Types and When to Use

### PreToolUse (Validation)
**Purpose**: Block actions before they happen

```python
# Block if issues found
if has_issues(command):
    print("Specific issue details", file=sys.stderr)
    sys.exit(2)  # Blocks tool call, shows stderr to Claude
```

**Common uses**:
- Validate bash commands
- Block commits to main
- Check for secrets in code
- Verify schema changes have migrations

### PostToolUse (Formatting/Cleanup)
**Purpose**: Auto-fix after action completes

```python
# Auto-format after edit
if is_typescript(file_path):
    run_prettier(file_path)
sys.exit(0)  # Success, stdout shown to user
```

**Common uses**:
- Auto-format code
- Remind about next steps
- Log operations
- Validate output

### UserPromptSubmit (Context Injection)
**Purpose**: Add context before Claude processes prompt

```python
# Add context
context = get_current_state()
print(context)  # Added to context for Claude
sys.exit(0)
```

**Common uses**:
- Add environment info
- Inject current state
- Block sensitive prompts
- Validate user input

### SessionStart (Setup)
**Purpose**: Initialize environment at session start

```python
# Set up environment
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export NODE_ENV=development' >> "$CLAUDE_ENV_FILE"
fi
```

**Common uses**:
- Set environment variables
- Load project context
- Check dependencies
- Display reminders

## Input/Output Patterns

### Exit Codes
- **0**: Success (stdout shown to user in transcript, or added to context for UserPromptSubmit/SessionStart)
- **2**: Blocking error (stderr fed to Claude automatically)
- **Other**: Non-blocking error (stderr shown to user)

### JSON Output
For advanced control:

```python
output = {
    "hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "permissionDecision": "deny",
        "permissionDecisionReason": "Specific reason here"
    }
}
print(json.dumps(output))
```

## Configuration Pattern

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "python \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/validate.py",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

**Note**: Use `$CLAUDE_PROJECT_DIR` for project-relative paths.

## Common Patterns

### Bash Command Validation
```python
#!/usr/bin/env python3
import json, sys

data = json.load(sys.stdin)
command = data.get('tool_input', {}).get('command', '')

# Check patterns
if 'rm -rf /' in command:
    print("Dangerous command blocked", file=sys.stderr)
    sys.exit(2)

# Or validate against rules
issues = validate_command(command)
if issues:
    for issue in issues:
        print(f"• {issue}", file=sys.stderr)
    sys.exit(2)
```

### File Protection
```python
data = json.load(sys.stdin)
path = data.get('tool_input', {}).get('file_path', '')

# Block sensitive files
sensitive = ['.env', 'package-lock.json', '.git/']
if any(p in path for p in sensitive):
    print(f"Cannot modify {path}", file=sys.stderr)
    sys.exit(2)
```

### Auto-formatting
```python
data = json.load(sys.stdin)
path = data.get('tool_input', {}).get('file_path', '')

if path.endswith('.ts'):
    subprocess.run(['prettier', '--write', path])
    print(f"Formatted {path}")
sys.exit(0)
```

## Avoiding Voodoo Constants

Document why, not just what:

```python
# Good - explains reasoning
REQUEST_TIMEOUT = 30  # HTTP requests typically complete within 30s
MAX_RETRIES = 3       # Most intermittent failures resolve by 2nd retry

# Bad - magic numbers
TIMEOUT = 47  # Why 47?
RETRIES = 5   # Why 5?
```

## Testing Hooks

1. **Test manually first**: Run hook script directly
2. **Check permissions**: Ensure scripts are executable
3. **Verify paths**: Use absolute paths or $CLAUDE_PROJECT_DIR
4. **Test edge cases**: Empty input, missing files, permissions
5. **Check stderr/stdout**: Verify correct output streams

## Project Hooks

**Current hooks** (see .claude/settings.json):
- `prevent_main_commits.sh` - Block commits to main
- `validate_version_bump.py` - Require CHANGELOG updates for code changes
- `validate_no_emojis.py` - Enforce no emoji rule
- `validate_no_dev_markers.sh` - Block TODO/FIXME/HACK in commits
- `validate_bash_usage.py` - Enforce proper tool usage
- `validate_file_size.py` - Enforce 500 LOC limit
- `check_build.py` - Auto-run go vet after edits
- `activate_skills.py` - Auto-suggest relevant skills

Follow these patterns when creating new hooks.
