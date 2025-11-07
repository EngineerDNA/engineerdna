#!/usr/bin/env python3
"""
Validates that no dead code exists before committing to main.

Uses golangci-lint with deadcode, unused, and ineffassign linters.

Exit codes:
  0: No dead code found or not applicable
  2: Dead code found (blocks commit)
"""

import json
import subprocess
import sys
import os


def is_commit_to_main(command: str) -> bool:
    """Check if this is a git commit command"""
    return 'git commit' in command


def get_current_branch() -> str:
    """Get the current git branch"""
    try:
        result = subprocess.run(
            ['git', 'branch', '--show-current'],
            capture_output=True,
            text=True,
            timeout=5
        )
        return result.stdout.strip()
    except Exception:
        return ""


def run_golangci_lint(project_root: str) -> tuple[bool, str]:
    """
    Run golangci-lint with dead code detection linters.

    Returns (success, output)
    """
    try:
        # Run golangci-lint with specific linters for dead code
        result = subprocess.run(
            [
                'golangci-lint', 'run',
                '--enable=deadcode',
                '--enable=unused',
                '--enable=ineffassign',
                '--enable=structcheck',
                '--enable=varcheck',
                './...'
            ],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=60
        )

        # golangci-lint returns 0 if no issues found
        success = result.returncode == 0
        output = result.stdout if result.stdout else result.stderr

        return success, output

    except subprocess.TimeoutExpired:
        return False, "Error: golangci-lint timed out after 60 seconds"
    except FileNotFoundError:
        # golangci-lint not installed, skip check
        return True, "golangci-lint not found, skipping dead code check"
    except Exception as e:
        return False, f"Error running golangci-lint: {str(e)}"


def format_dead_code_output(output: str) -> str:
    """Format golangci-lint output for display"""
    if not output or "golangci-lint not found" in output:
        return ""

    lines = [
        "",
        "=" * 80,
        "DEAD CODE DETECTED",
        "=" * 80,
        "",
        "The following dead code was found:",
        "",
        output,
        "",
        "Please remove all dead code before committing to main.",
        "",
        "Dead code includes:",
        "  - Unused functions, types, constants, variables",
        "  - Ineffectual assignments (assigned but never read)",
        "  - Unused struct fields",
        "",
        "To fix:",
        "  1. Review each item above",
        "  2. Remove or use the dead code",
        "  3. Run: golangci-lint run --enable=deadcode,unused,ineffassign ./...",
        "  4. Commit when clean",
        "",
        "=" * 80,
        ""
    ]

    return "\n".join(lines)


def main():
    try:
        # Read tool input from stdin
        tool_input = json.load(sys.stdin)
        command = tool_input.get('tool_input', {}).get('command', '')

        # Only check on commits to main
        if not is_commit_to_main(command):
            sys.exit(0)

        # Check if we're on main branch
        current_branch = get_current_branch()
        if current_branch != 'main':
            sys.exit(0)

        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Run golangci-lint
        success, output = run_golangci_lint(project_root)

        if success:
            # No dead code found
            sys.exit(0)

        # Dead code found, block commit
        formatted_output = format_dead_code_output(output)
        print(formatted_output, file=sys.stderr)

        sys.exit(2)  # Block

    except json.JSONDecodeError:
        # Not valid JSON input, allow
        sys.exit(0)
    except Exception as e:
        print(f"Error in dead code validation: {str(e)}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == '__main__':
    main()
