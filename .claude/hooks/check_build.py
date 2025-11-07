#!/usr/bin/env python3
"""
Build Checking Hook (Stop)

Runs after Claude finishes responding. If TypeScript or Go files were edited,
automatically runs typecheck (TS) or go vet (Go) and shows results.

This prevents compilation errors from being left behind.
"""
import json
import sys
import os
import subprocess
from pathlib import Path


def get_edit_log_path(project_root: str) -> Path:
    """Get path to edit tracking log"""
    return Path(project_root) / '.claude' / 'logs' / 'edit_tracking.json'


def load_edit_log(log_path: Path) -> dict:
    """Load edit log"""
    if not log_path.exists():
        return {'edits': []}

    try:
        with open(log_path, 'r') as f:
            return json.load(f)
    except (json.JSONDecodeError, FileNotFoundError):
        return {'edits': []}


def clear_edit_log(log_path: Path):
    """Clear edit log after checking"""
    if log_path.exists():
        log_path.unlink()


def has_typescript_edits(log_data: dict) -> bool:
    """Check if any TypeScript files were edited"""
    return any(edit.get('is_typescript', False) for edit in log_data.get('edits', []))


def has_go_edits(log_data: dict) -> bool:
    """Check if any Go files were edited"""
    return any(edit.get('is_go', False) for edit in log_data.get('edits', []))


def run_typecheck(project_root: str) -> tuple[bool, str]:
    """
    Run pnpm typecheck and return (success, output)
    """
    try:
        result = subprocess.run(
            ['pnpm', 'typecheck'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=120
        )

        return result.returncode == 0, result.stdout + result.stderr

    except subprocess.TimeoutExpired:
        return False, "Error: typecheck command timed out after 120 seconds"
    except FileNotFoundError:
        return False, "Error: pnpm command not found"
    except Exception as e:
        return False, f"Error running typecheck: {str(e)}"


def run_go_vet(project_root: str) -> tuple[bool, str]:
    """
    Run go vet ./... and return (success, output)
    """
    try:
        result = subprocess.run(
            ['go', 'vet', './...'],
            cwd=project_root,
            capture_output=True,
            text=True,
            timeout=120
        )

        return result.returncode == 0, result.stdout + result.stderr

    except subprocess.TimeoutExpired:
        return False, "Error: go vet command timed out after 120 seconds"
    except FileNotFoundError:
        return False, "Error: go command not found"
    except Exception as e:
        return False, f"Error running go vet: {str(e)}"


def count_errors(output: str) -> int:
    """Count TypeScript errors in output"""
    # Look for "Found X errors" pattern
    import re
    match = re.search(r'Found (\d+) errors?', output)
    if match:
        return int(match.group(1))
    return 0


def format_output(success: bool, output: str, error_count: int) -> str:
    """Format typecheck output for display"""
    lines = [
        "",
        "=" * 80,
        "AUTOMATIC BUILD CHECK",
        "=" * 80,
        ""
    ]

    if success:
        lines.extend([
            "[PASS] TypeScript check passed",
            "No type errors found.",
            ""
        ])
    else:
        lines.extend([
            f"[FAIL] TypeScript check failed with {error_count} error(s)",
            ""
        ])

        if error_count >= 5:
            lines.extend([
                "Recommendation: Use the quality agent to systematically fix errors:",
                "  \"Launch the quality agent to fix these TypeScript errors\"",
                "",
                "Showing first 50 lines of output:",
                ""
            ])
            # Show limited output for many errors
            output_lines = output.split('\n')[:50]
            lines.extend(output_lines)
            if len(output.split('\n')) > 50:
                lines.append("... (output truncated)")
        else:
            lines.extend([
                "Errors found:",
                "",
                output
            ])

    lines.extend([
        "",
        "=" * 80,
        ""
    ])

    return "\n".join(lines)


def main():
    try:
        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Get edit log
        log_path = get_edit_log_path(project_root)
        log_data = load_edit_log(log_path)

        has_ts = has_typescript_edits(log_data)
        has_go = has_go_edits(log_data)

        # Check if any relevant files were edited
        if not has_ts and not has_go:
            # No TypeScript or Go files edited, skip check
            clear_edit_log(log_path)
            sys.exit(0)

        outputs = []

        # Run Go vet if Go files were edited
        if has_go:
            success, output = run_go_vet(project_root)
            if success:
                outputs.append("\n[PASS] Go vet check passed\nNo issues found.\n")
            else:
                outputs.append(f"\n[FAIL] Go vet check failed\n\nErrors:\n{output}\n")

        # Run typecheck if TypeScript files were edited
        if has_ts:
            success, output = run_typecheck(project_root)
            error_count = count_errors(output) if not success else 0
            formatted_output = format_output(success, output, error_count)
            outputs.append(formatted_output)

        # Display results
        if outputs:
            print("\n" + "=" * 80)
            print("AUTOMATIC BUILD CHECK")
            print("=" * 80)
            for output in outputs:
                print(output)
            print("=" * 80 + "\n")

        # Clear edit log
        clear_edit_log(log_path)

        # Always allow (non-blocking)
        sys.exit(0)

    except Exception as e:
        print(f"Error in build check: {e}", file=sys.stderr)
        sys.exit(0)


if __name__ == "__main__":
    main()
