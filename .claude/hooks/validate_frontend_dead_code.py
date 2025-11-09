#!/usr/bin/env python3
"""
Validates that no dead code exists in the frontend before committing.

Uses knip to detect:
- Unused files
- Unused dependencies
- Unused exports
- Unused types
- Unlisted dependencies

Exit codes:
  0: No dead code found or not applicable
  2: Dead code found (blocks commit)
"""

import json
import subprocess
import sys
import os


def is_git_command(command: str) -> bool:
    """Check if this is a git command"""
    return command.startswith('git ')


def is_frontend_change() -> bool:
    """Check if any frontend files were changed"""
    try:
        # Get staged files
        result = subprocess.run(
            ['git', 'diff', '--cached', '--name-only'],
            capture_output=True,
            text=True,
            timeout=5
        )

        staged_files = result.stdout.strip().split('\n')

        # Check if any files are in frontend/
        for file in staged_files:
            if file.startswith('frontend/src/') and file.endswith(('.ts', '.tsx')):
                return True

        return False

    except Exception:
        # If we can't determine, skip the check
        return False


def run_knip(project_root: str) -> tuple[bool, str]:
    """
    Run knip to detect dead code in frontend.

    Returns (success, output)
    """
    frontend_dir = os.path.join(project_root, 'frontend')

    if not os.path.exists(frontend_dir):
        return True, "Frontend directory not found, skipping"

    try:
        # Run knip in frontend directory
        result = subprocess.run(
            ['npm', 'run', 'knip'],
            cwd=frontend_dir,
            capture_output=True,
            text=True,
            timeout=30
        )

        # knip returns non-zero if issues found
        success = result.returncode == 0
        output = result.stdout if result.stdout else result.stderr

        return success, output

    except subprocess.TimeoutExpired:
        return False, "Error: knip timed out after 30 seconds"
    except FileNotFoundError:
        # knip not installed yet, skip check
        return True, "knip not installed, skipping dead code check"
    except Exception as e:
        return False, f"Error running knip: {str(e)}"


def format_dead_code_output(output: str) -> str:
    """Format knip output for display"""
    if not output or "knip not installed" in output:
        return ""

    # knip already has good formatting, just add header/footer
    lines = [
        "",
        "=" * 80,
        "FRONTEND DEAD CODE DETECTED",
        "=" * 80,
        "",
        output,
        "",
        "Please clean up dead code before committing.",
        "",
        "To fix:",
        "  1. cd frontend",
        "  2. npm run knip  # See all issues",
        "  3. Remove unused files, exports, dependencies",
        "  4. npm run knip  # Verify clean",
        "  5. Commit when clean",
        "",
        "For help: https://knip.dev/overview/getting-started",
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

        # Only check on git commands
        if not is_git_command(command):
            sys.exit(0)

        # Only check if frontend files changed
        if not is_frontend_change():
            sys.exit(0)

        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Run knip
        success, output = run_knip(project_root)

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
        print(f"Error in frontend dead code validation: {str(e)}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == '__main__':
    main()
