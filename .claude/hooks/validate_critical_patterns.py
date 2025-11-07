#!/usr/bin/env python3
"""
Hook: Validate Critical Patterns (Stop)
Runs after Claude finishes responding to validate critical patterns.
Enforces BLOCKING rules defined in skill-rules.json.

Exit codes:
  0: No violations or warnings only
  2: Blocking violations found
"""

import json
import sys
import os
import re
from pathlib import Path
from typing import List, Dict, Any, Tuple
import subprocess


def load_skill_rules() -> Dict[str, Any]:
    """Load skill-rules.json configuration."""
    project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())
    rules_path = Path(project_root) / '.claude' / 'hooks' / 'skill-rules.json'

    if not rules_path.exists():
        return {}

    with open(rules_path, 'r') as f:
        return json.load(f)


def get_recent_edits() -> List[str]:
    """
    Get list of recently modified files from git.
    Returns files that are modified or staged.
    """
    try:
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())
        os.chdir(project_root)

        # Get modified and staged files
        result = subprocess.run(
            ['git', 'diff', '--name-only', 'HEAD'],
            capture_output=True,
            text=True,
            check=False
        )

        if result.returncode != 0:
            return []

        files = [f.strip() for f in result.stdout.split('\n') if f.strip()]

        # Also check unstaged changes
        result2 = subprocess.run(
            ['git', 'diff', '--name-only'],
            capture_output=True,
            text=True,
            check=False
        )

        if result2.returncode == 0:
            unstaged = [f.strip() for f in result2.stdout.split('\n') if f.strip()]
            files.extend(unstaged)

        # Deduplicate and filter for Go/TypeScript files
        files = list(set(files))
        files = [f for f in files if f.endswith(('.go', '.ts', '.tsx', '.js', '.jsx'))]

        return files

    except Exception as e:
        print(f"Error getting recent edits: {e}", file=sys.stderr)
        return []


def check_file_for_violations(
    file_path: str,
    rules: Dict[str, Any]
) -> List[Tuple[str, str, str, str]]:
    """
    Check a file for critical pattern violations.
    Returns: List of (skill_name, pattern, message, severity)
    """
    violations = []
    project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())
    full_path = Path(project_root) / file_path

    if not full_path.exists():
        return violations

    try:
        with open(full_path, 'r') as f:
            content = f.read()
    except Exception:
        return violations

    # Check each skill's critical patterns
    for skill_name, skill_config in rules.items():
        if not isinstance(skill_config, dict):
            continue

        critical_patterns = skill_config.get('criticalPatterns', {})
        pattern_violations = critical_patterns.get('violations', [])

        for violation_config in pattern_violations:
            pattern = violation_config.get('pattern', '')
            message = violation_config.get('message', '')
            severity = violation_config.get('severity', 'WARNING')

            if pattern and re.search(pattern, content, re.MULTILINE):
                violations.append((skill_name, pattern, message, severity))

    return violations


def format_violations(
    violations: List[Tuple[str, str, str, str, str]]
) -> Tuple[bool, str]:
    """
    Format violations for display.
    Returns: (has_blocking, formatted_message)
    """
    if not violations:
        return (False, "")

    blocking = [v for v in violations if v[4] == 'BLOCKING']
    warnings = [v for v in violations if v[4] == 'WARNING']

    has_blocking = len(blocking) > 0

    lines = [
        "",
        "=" * 80,
    ]

    if has_blocking:
        lines.extend([
            "[CRITICAL VALIDATION FAILURE]",
            "=" * 80,
            "",
            "BLOCKING violations detected:",
            ""
        ])

        for file_path, skill, pattern, message, severity in blocking:
            lines.append(f"[BLOCKED] {file_path}")
            lines.append(f"   └─ {message}")
            lines.append(f"   └─ Related skill: {skill}")
            lines.append("")

        lines.extend([
            "These must be fixed before proceeding.",
            "=" * 80,
            ""
        ])

    elif warnings:
        lines.extend([
            "[PATTERN VALIDATION WARNINGS]",
            "=" * 80,
            "",
            "Potential issues detected:",
            ""
        ])

        for file_path, skill, pattern, message, severity in warnings:
            lines.append(f"[WARNING] {file_path}")
            lines.append(f"   └─ {message}")
            lines.append(f"   └─ Consider: {skill}")
            lines.append("")

        lines.extend([
            "Review these warnings before moving forward.",
            "=" * 80,
            ""
        ])

    return (has_blocking, "\n".join(lines))


def main():
    try:
        # Load skill rules
        rules = load_skill_rules()

        if not rules:
            # No rules configured, skip validation
            sys.exit(0)

        # Get recently edited files
        edited_files = get_recent_edits()

        if not edited_files:
            # No files edited, nothing to validate
            sys.exit(0)

        # Check each file for violations
        all_violations = []
        for file_path in edited_files:
            violations = check_file_for_violations(file_path, rules)
            for skill, pattern, message, severity in violations:
                all_violations.append((file_path, skill, pattern, message, severity))

        # Format and display violations
        has_blocking, message = format_violations(all_violations)

        if message:
            print(message, file=sys.stderr)

        # Exit with appropriate code
        if has_blocking:
            sys.exit(2)  # Block execution
        else:
            sys.exit(0)  # Allow execution

    except Exception as e:
        print(f"[WARNING] Error in critical pattern validation: {e}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == "__main__":
    main()
