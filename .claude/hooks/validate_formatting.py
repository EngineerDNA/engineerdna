#!/usr/bin/env python3
"""
Validates that all code is properly formatted before committing.

Go: gofmt, goimports
TypeScript/React: prettier (when frontend added)

Exit codes:
  0: All code is formatted or not applicable
  2: Unformatted code found (blocks commit)
"""

import json
import subprocess
import sys
import os
from pathlib import Path


def is_commit_command(command: str) -> bool:
    """Check if this is a git commit command"""
    return 'git commit' in command


def has_go_files(project_root: str) -> bool:
    """Check if there are any Go files in the staging area"""
    try:
        result = subprocess.run(
            ['git', 'diff', '--cached', '--name-only', '--diff-filter=ACM'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=5
        )
        files = result.stdout.strip().split('\n')
        return any(f.endswith('.go') for f in files if f)
    except Exception:
        return False


def has_typescript_files(project_root: str) -> bool:
    """Check if there are any TypeScript/TSX files in the staging area"""
    try:
        result = subprocess.run(
            ['git', 'diff', '--cached', '--name-only', '--diff-filter=ACM'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=5
        )
        files = result.stdout.strip().split('\n')
        return any(f.endswith(('.ts', '.tsx')) for f in files if f)
    except Exception:
        return False


def check_gofmt(project_root: str) -> tuple[bool, list[str]]:
    """
    Check if Go files are formatted with gofmt.

    Returns (all_formatted, unformatted_files)
    """
    try:
        # gofmt -l lists files that need formatting
        result = subprocess.run(
            ['gofmt', '-l', '.'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=30
        )

        unformatted = [f for f in result.stdout.strip().split('\n') if f]
        return len(unformatted) == 0, unformatted

    except FileNotFoundError:
        # gofmt not found, skip check
        return True, []
    except Exception as e:
        print(f"Error running gofmt: {str(e)}", file=sys.stderr)
        return True, []


def check_goimports(project_root: str) -> tuple[bool, list[str]]:
    """
    Check if Go files have correct imports (goimports).

    Returns (all_correct, files_with_issues)
    """
    try:
        # goimports -l lists files with incorrect imports
        result = subprocess.run(
            ['goimports', '-l', '.'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=30
        )

        files_with_issues = [f for f in result.stdout.strip().split('\n') if f]
        return len(files_with_issues) == 0, files_with_issues

    except FileNotFoundError:
        # goimports not found, skip check
        return True, []
    except Exception as e:
        print(f"Error running goimports: {str(e)}", file=sys.stderr)
        return True, []


def check_prettier(project_root: str) -> tuple[bool, list[str]]:
    """
    Check if TypeScript/React files are formatted with prettier.

    Returns (all_formatted, unformatted_files)
    """
    try:
        frontend_dir = os.path.join(project_root, 'frontend')
        if not os.path.exists(frontend_dir):
            # Frontend not yet added
            return True, []

        # prettier --check lists files that need formatting
        result = subprocess.run(
            ['npx', 'prettier', '--check', 'src/**/*.{ts,tsx}'],
            cwd=frontend_dir,
            capture_output=True,
            text=True,
            timeout=30
        )

        # prettier returns non-zero if files need formatting
        if result.returncode == 0:
            return True, []

        # Parse output for unformatted files
        unformatted = []
        for line in result.stdout.split('\n'):
            if line.strip() and not line.startswith('Checking'):
                unformatted.append(line.strip())

        return False, unformatted

    except FileNotFoundError:
        # prettier not found, skip check
        return True, []
    except Exception as e:
        print(f"Error running prettier: {str(e)}", file=sys.stderr)
        return True, []


def format_go_output(gofmt_files: list[str], goimports_files: list[str]) -> str:
    """Format Go formatting issues for display"""
    lines = [
        "",
        "=" * 80,
        "GO CODE FORMATTING ISSUES",
        "=" * 80,
        ""
    ]

    if gofmt_files:
        lines.extend([
            "Files not formatted with gofmt:",
            ""
        ])
        for file in gofmt_files:
            lines.append(f"  - {file}")
        lines.append("")

    if goimports_files:
        lines.extend([
            "Files with incorrect imports:",
            ""
        ])
        for file in goimports_files:
            lines.append(f"  - {file}")
        lines.append("")

    lines.extend([
        "To fix:",
        "  # Format all Go files",
        "  gofmt -w .",
        "  goimports -w .",
        "",
        "  # Or use your editor's format-on-save",
        "",
        "=" * 80,
        ""
    ])

    return "\n".join(lines)


def format_typescript_output(prettier_files: list[str]) -> str:
    """Format TypeScript formatting issues for display"""
    lines = [
        "",
        "=" * 80,
        "TYPESCRIPT CODE FORMATTING ISSUES",
        "=" * 80,
        "",
        "Files not formatted with prettier:",
        ""
    ]

    for file in prettier_files:
        lines.append(f"  - {file}")

    lines.extend([
        "",
        "To fix:",
        "  cd frontend",
        "  npx prettier --write 'src/**/*.{ts,tsx}'",
        "",
        "  # Or use your editor's format-on-save",
        "",
        "=" * 80,
        ""
    ])

    return "\n".join(lines)


def main():
    try:
        # Read tool input from stdin
        tool_input = json.load(sys.stdin)
        command = tool_input.get('tool_input', {}).get('command', '')

        # Only check on git commit commands
        if not is_commit_command(command):
            sys.exit(0)

        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        issues_found = False

        # Check Go formatting if Go files are staged
        if has_go_files(project_root):
            gofmt_ok, gofmt_files = check_gofmt(project_root)
            goimports_ok, goimports_files = check_goimports(project_root)

            if not gofmt_ok or not goimports_ok:
                output = format_go_output(gofmt_files, goimports_files)
                print(output, file=sys.stderr)
                issues_found = True

        # Check TypeScript formatting if TypeScript files are staged
        if has_typescript_files(project_root):
            prettier_ok, prettier_files = check_prettier(project_root)

            if not prettier_ok:
                output = format_typescript_output(prettier_files)
                print(output, file=sys.stderr)
                issues_found = True

        if issues_found:
            sys.exit(2)  # Block commit

        sys.exit(0)  # Allow commit

    except json.JSONDecodeError:
        # Not valid JSON input, allow
        sys.exit(0)
    except Exception as e:
        print(f"Error in formatting validation: {str(e)}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == '__main__':
    main()
