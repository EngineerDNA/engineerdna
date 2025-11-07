#!/usr/bin/env python3
"""
Validate markdown file creation against .gitignore rules.
Blocks creation of unauthorized .md files.
"""
import json
import sys
import os
from pathlib import Path


def is_allowed_markdown_file(file_path: str, project_root: str) -> tuple[bool, str]:
    """
    Check if markdown file is allowed per .gitignore rules.

    Returns: (is_allowed, reason_if_blocked)
    """
    if not file_path.endswith('.md'):
        return (True, "")

    # Convert to Path for easier manipulation
    path = Path(file_path)

    # Convert to relative path from project root if absolute
    if path.is_absolute():
        try:
            rel_path = path.relative_to(project_root)
        except ValueError:
            # Path is outside project root
            rel_path = path
    else:
        rel_path = path

    # Convert to string with forward slashes
    rel_path_str = str(rel_path).replace('\\', '/')

    # Allowed patterns from .gitignore (lines 98-111)
    allowed_patterns = [
        # .claude directory patterns
        ('.claude/agents/', lambda p: p.startswith('.claude/agents/') and p.count('/') == 2),
        ('.claude/commands/', lambda p: p.startswith('.claude/commands/') and p.count('/') == 2),
        ('.claude/hooks/', lambda p: p.startswith('.claude/hooks/') and p.count('/') == 2),
        ('.claude/skills/', lambda p: p.startswith('.claude/skills/') and '/' in p[len('.claude/skills/'):]),

        # CLAUDE.md files anywhere
        ('CLAUDE.md', lambda p: p.endswith('/CLAUDE.md') or p == 'CLAUDE.md'),

        # Root level files
        ('README.md', lambda p: p == 'README.md'),
        ('CHANGELOG.md', lambda p: p == 'CHANGELOG.md'),
        ('CONTRIBUTING.md', lambda p: p == 'CONTRIBUTING.md'),
        ('LICENSE.md', lambda p: p == 'LICENSE.md'),

        # PDR files (Product Development Records)
        ('PDR-*.md', lambda p: p.startswith('PDR-') and p.endswith('.md') and '/' not in p),

        # Dev docs (temporary workflow files)
        ('PLAN.md', lambda p: p == 'PLAN.md'),
        ('CONTEXT.md', lambda p: p == 'CONTEXT.md'),
        ('TASKS.md', lambda p: p == 'TASKS.md'),
    ]

    # Check against allowed patterns
    for pattern_name, check_func in allowed_patterns:
        if check_func(rel_path_str):
            return (True, "")

    # Not in allowed list
    reason = f"""[BLOCKED] Cannot create unauthorized markdown file: {rel_path_str}

Only specific .md files are allowed per .gitignore.

Allowed patterns:
  - **/CLAUDE.md (anywhere)
  - .claude/agents/*.md
  - .claude/commands/*.md
  - .claude/hooks/*.md
  - .claude/skills/**/*.md
  - Root: README.md, CHANGELOG.md, CONTRIBUTING.md, LICENSE.md, PDR-*.md, PLAN.md, CONTEXT.md, TASKS.md

Instead:
  - For agent docs → agents/CLAUDE.md
  - For skill docs → .claude/skills/your-skill/SKILL.md or reference/*.md
  - For project rules → CLAUDE.md files
  - For strategic docs → .claude/skills/ (e.g., accessing-business-context skill)
  - For workflows/patterns → .claude/skills/ (comprehensive guides)
  - For dev workflow → PLAN.md, CONTEXT.md, TASKS.md (use /create-dev-docs)
  - For product development records → PDR-*.md (gitignored)
"""
    return (False, reason)


def main():
    try:
        # Read hook input from stdin
        input_data = json.load(sys.stdin)

        # Get file path from tool input
        file_path = input_data.get('tool_input', {}).get('file_path', '')

        if not file_path:
            # No file path, nothing to validate
            sys.exit(0)

        # Get project root from environment
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Check if file is allowed
        is_allowed, reason = is_allowed_markdown_file(file_path, project_root)

        if not is_allowed:
            print(reason, file=sys.stderr)
            sys.exit(2)  # Block the operation

        # File is allowed
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error validating markdown file: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
