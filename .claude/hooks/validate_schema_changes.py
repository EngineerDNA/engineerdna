#!/usr/bin/env python3
"""
Pre-commit hook to ensure schema changes have accompanying migrations.
Blocks commits that modify database schema without creating migrations.
Designed for EngineerDNA's SQLite migration system.

Exit codes:
  0: No schema changes or migrations present
  2: Schema changes without migration (blocks commit)
"""
import json
import sys
import subprocess
from pathlib import Path
import os


def is_commit_command(command: str) -> bool:
    """Check if this is a git commit command"""
    return 'git commit' in command


def get_project_root():
    """Get the project root from environment or git."""
    project_root = os.environ.get('CLAUDE_PROJECT_DIR')
    if project_root:
        return Path(project_root)

    result = subprocess.run(
        ['git', 'rev-parse', '--show-toplevel'],
        capture_output=True,
        text=True,
        check=False
    )
    if result.returncode == 0:
        return Path(result.stdout.strip())
    return Path.cwd()


def get_staged_files():
    """Get list of staged files from git."""
    result = subprocess.run(
        ['git', 'diff', '--cached', '--name-only'],
        capture_output=True,
        text=True
    )
    return result.stdout.strip().split('\n') if result.stdout else []


def has_schema_changes(files):
    """
    Check if any files that could affect database schema are being changed.

    EngineerDNA schema could be in:
    - internal/db/ (database models)
    - internal/models/ (data models)
    - Files that define table structures
    """
    schema_indicators = [
        'internal/db/',
        'internal/models/',
    ]

    for f in files:
        if not f.endswith('.go'):
            continue

        for indicator in schema_indicators:
            if indicator in f:
                # Check if file actually contains schema definitions
                try:
                    with open(f, 'r') as file:
                        content = file.read()
                        # Look for common schema patterns in Go
                        if any(pattern in content for pattern in [
                            'CREATE TABLE',
                            'ALTER TABLE',
                            'DROP TABLE',
                            'ADD COLUMN',
                            'sql.Exec(',
                            'db.Exec(',
                            'CREATE INDEX',
                        ]):
                            return True
                except Exception:
                    pass

    return False


def has_migration_changes(files):
    """Check if any migration files are being added."""
    return any(
        'migrations/' in f and f.endswith('.sql')
        for f in files
    )


def main():
    try:
        # Read tool input from stdin (if called as PreToolUse hook)
        try:
            tool_input = json.load(sys.stdin)
            command = tool_input.get('tool_input', {}).get('command', '')

            # Only check on git commit commands
            if not is_commit_command(command):
                sys.exit(0)
        except (json.JSONDecodeError, KeyError):
            # Not called as PreToolUse hook, check anyway
            pass

        root = get_project_root()
        migrations_dir = root / 'migrations'

        # Check if migrations directory exists
        if not migrations_dir.exists():
            # No migrations directory, skip validation
            sys.exit(0)

        staged_files = get_staged_files()

        if has_schema_changes(staged_files):
            if not has_migration_changes(staged_files):
                print("", file=sys.stderr)
                print("=" * 80, file=sys.stderr)
                print("[BLOCKED] Schema changes detected without migration!", file=sys.stderr)
                print("=" * 80, file=sys.stderr)
                print("", file=sys.stderr)
                print("You modified database schema code but didn't create a migration.", file=sys.stderr)
                print("", file=sys.stderr)
                print("To fix this:", file=sys.stderr)
                print("  1. Create a new migration file in migrations/", file=sys.stderr)
                print("     Example: migrations/004_add_new_table.sql", file=sys.stderr)
                print("  2. Write the SQL for your schema changes", file=sys.stderr)
                print("  3. git add migrations/004_add_new_table.sql", file=sys.stderr)
                print("  4. git commit again", file=sys.stderr)
                print("", file=sys.stderr)
                print("This ensures all schema changes are tracked and reproducible.", file=sys.stderr)
                print("(Rule 9 and Rule 27: NO one-off DB scripts, always use migrations)", file=sys.stderr)
                print("=" * 80, file=sys.stderr)
                print("", file=sys.stderr)
                sys.exit(2)

        sys.exit(0)

    except Exception as e:
        print(f"[WARNING] Schema validation hook error: {e}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == '__main__':
    main()
