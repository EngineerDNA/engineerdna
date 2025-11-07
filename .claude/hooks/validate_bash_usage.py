#!/usr/bin/env python3
"""
Validate Bash command usage to enforce proper tool usage.
Blocks file operations that should use specialized tools.
"""
import json
import sys
import re


def analyze_bash_command(command: str) -> tuple[bool, str]:
    """
    Analyze bash command and block if it should use a specialized tool.

    Returns: (is_allowed, reason_if_blocked)
    """
    # Skip empty commands
    if not command or not command.strip():
        return (True, "")

    # Remove leading/trailing whitespace
    command = command.strip()

    # === EXPLICIT BLOCKS ===

    # Block file creation via redirection
    if re.search(r'(?:^|\||;|&&)\s*\w+[^|]*>\s*[^\s]+\.md(?:\s|$|;|\||&&)', command):
        return (False, f"""[BLOCKED] Creating .md files via Bash redirection bypasses validation hooks.

Command: {command}

Use the Write tool instead:
  Write tool → file_path: "path/to/file.md", content: "..."

Why: Write operations are tracked by hooks (emoji validation, size limits)
Bash redirection bypasses all validation.""")

    # Block file creation via heredoc
    if re.search(r'(?:cat|tee)\s+(?:>|>>)\s*[^\s]+\.md\s*<<', command) or \
       re.search(r'<<\s*EOF.*\.md', command):
        return (False, f"""[BLOCKED] Creating files via heredoc bypasses validation hooks.

Command: {command}

Use the Write tool instead:
  Write tool → file_path: "path/to/file.md", content: "..."

Why: Heredoc creation bypasses emoji validation and size limits.""")

    # Block general file writes (any extension)
    write_patterns = [
        r'(?:^|\||;|&&)\s*(?:echo|printf)\s+[^|]*>\s*[^\s]+',  # echo/printf to file
        r'(?:^|\||;|&&)\s*cat\s+>\s*[^\s]+',  # cat > file
        r'(?:^|\||;|&&)\s*tee\s+[^\s]+',  # tee (creates/overwrites)
    ]

    for pattern in write_patterns:
        if re.search(pattern, command):
            # Exception: Temporary files in /tmp are OK
            if '/tmp/' in command or re.search(r'>\s*/dev/', command):
                continue

            return (False, f"""[BLOCKED] File write operations should use the Write tool.

Command: {command}

Use the Write tool instead:
  Write tool → file_path: "path/to/file", content: "..."

Why: Write tool provides proper tracking, validation, and error handling.""")

    # Block file reading (suggest Read tool)
    read_patterns = [
        (r'(?:^|\||;|&&)\s*cat\s+[^\s|>&]+(?:\s|$|;|\||&&)',
         'Use Read tool instead of cat'),
        (r'(?:^|\||;|&&)\s*(?:head|tail)\s+',
         'Use Read tool with offset/limit instead of head/tail'),
        (r'(?:^|\||;|&&)\s*less\s+',
         'Use Read tool instead of less'),
    ]

    for pattern, suggestion in read_patterns:
        if re.search(pattern, command):
            # Exceptions for legitimate uses
            if any(allowed in command for allowed in [
                'cat /proc/',  # System information
                'cat /sys/',   # System information
                'cat /dev/',   # Device files
                'head -',      # Reading from stdin
                'tail -',      # Reading from stdin
                'tail -f',     # Log following
                '| cat',       # Piping
                '| head',      # Piping
                '| tail',      # Piping
            ]):
                continue

            return (False, f"""[BLOCKED] {suggestion}

Command: {command}

Use the Read tool instead:
  Read tool → file_path: "path/to/file"

For specific lines:
  Read tool → file_path: "path/to/file", offset: 10, limit: 20

Why: Read tool handles large files, pagination, and line numbers properly.""")

    # Block content searching (suggest Grep tool)
    if re.search(r'(?:^|\||;|&&)\s*(?:grep|rg|ag|ack)\s+[^|]', command):
        # Exception: Pipes are legitimate (filtering other command output)
        if not command.startswith('grep') and not command.startswith('rg'):
            # grep/rg in middle of pipe is usually OK
            pass
        else:
            return (False, f"""[BLOCKED] Content search should use the Grep tool.

Command: {command}

Use the Grep tool instead:
  Grep tool → pattern: "search_term", path: "directory/"

With glob filtering:
  Grep tool → pattern: "search_term", glob: "*.ts"

Why: Grep tool is optimized for Claude Code and provides better results.""")

    # Block file finding (suggest Glob tool)
    if re.search(r'(?:^|\||;|&&)\s*find\s+', command):
        # Exception: find with -exec for operations is OK
        if '-exec' in command or '-delete' in command:
            pass
        else:
            return (False, f"""[BLOCKED] File finding should use the Glob tool.

Command: {command}

Use the Glob tool instead:
  Glob tool → pattern: "**/*.ts"

Why: Glob tool is faster and provides better pattern matching.""")

    # Block sed/awk for editing (suggest Edit tool)
    edit_patterns = [
        (r'(?:^|\||;|&&)\s*sed\s+-i', 'sed -i'),
        (r'(?:^|\||;|&&)\s*awk.*>\s*[^\s]+', 'awk with output redirection'),
        (r'(?:^|\||;|&&)\s*perl\s+-pi', 'perl -pi'),
    ]

    for pattern, desc in edit_patterns:
        if re.search(pattern, command):
            return (False, f"""[BLOCKED] File editing via {desc} should use the Edit tool.

Command: {command}

Use the Edit tool instead:
  Edit tool → file_path: "path/to/file", old_string: "...", new_string: "..."

Why: Edit tool preserves context and provides proper error handling.""")

    # === ALLOWED CATEGORIES ===

    # Version control (git, gh)
    if re.search(r'(?:^|\s)(?:git|gh)\s+', command):
        return (True, "")

    # Package managers
    if re.search(r'(?:^|\s)(?:npm|pnpm|yarn|pip|cargo|gem|brew)\s+', command):
        return (True, "")

    # Build/test tools
    if re.search(r'(?:^|\s)(?:make|cmake|pytest|jest|cargo test|go test)\s+', command):
        return (True, "")

    # Docker/containers
    if re.search(r'(?:^|\s)(?:docker|docker-compose|kubectl|helm)\s+', command):
        return (True, "")

    # Process management
    if re.search(r'(?:^|\s)(?:ps|kill|pkill|killall|lsof|pgrep)\s+', command):
        return (True, "")

    # Directory operations
    if re.search(r'(?:^|\s)(?:pwd|cd|mkdir|rmdir|ls(?:\s|$))\s*', command):
        return (True, "")

    # File operations (legitimate)
    if re.search(r'(?:^|\s)(?:chmod|chown|mv|cp|rm|touch|ln)\s+', command):
        return (True, "")

    # System operations
    if re.search(r'(?:^|\s)(?:systemctl|service|launchctl)\s+', command):
        return (True, "")

    # Database operations (will be validated by other hooks)
    if re.search(r'(?:^|\s)(?:psql|mysql|sqlite3)\s+', command):
        return (True, "")

    # Stat operations
    if re.search(r'(?:^|\s)stat\s+', command):
        return (True, "")

    # Date/time
    if re.search(r'(?:^|\s)date\s+', command):
        return (True, "")

    # Environment/path
    if re.search(r'(?:^|\s)(?:which|whereis|type|command -v)\s+', command):
        return (True, "")

    # Network operations
    if re.search(r'(?:^|\s)(?:curl|wget|nc|telnet|ssh|scp|rsync)\s+', command):
        return (True, "")

    # jq (JSON processing)
    if re.search(r'(?:^|\s)jq\s+', command):
        return (True, "")

    # Simple echo for debugging (no redirection)
    if re.search(r'^echo\s+"[^"]*"$', command) and '>' not in command:
        return (True, "")

    # Allowed if just piping or chaining allowed commands
    if all(keyword not in command for keyword in ['>', '>>', '<<', 'cat ', 'grep ', 'find ', 'sed ', 'awk ', 'head ', 'tail ', 'less ']):
        return (True, "")

    # Default: allow (only block specific anti-patterns)
    return (True, "")


def main():
    try:
        # Read hook input from stdin
        input_data = json.load(sys.stdin)

        # Get command from tool input
        command = input_data.get('tool_input', {}).get('command', '')

        if not command:
            # No command, nothing to validate
            sys.exit(0)

        # Check if command should use a specialized tool
        is_allowed, reason = analyze_bash_command(command)

        if not is_allowed:
            print(reason, file=sys.stderr)
            sys.exit(2)  # Block the operation

        # Command is allowed
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error validating bash command: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
