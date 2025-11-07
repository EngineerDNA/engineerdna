#!/usr/bin/env python3
"""
Hook to prevent manual database modifications.
Blocks direct sqlite3 commands that modify schema or data.
Enforces Rule 9 and Rule 27: NO one-off DB scripts, always use migrations.

Exit codes:
  0: Command allowed
  2: Command blocked (direct DB modification)
"""

import json
import sys
import re


def is_dangerous_sqlite_command(command: str) -> bool:
    """Check if command is a dangerous direct sqlite3 modification."""
    # Pattern for sqlite3 commands with SQL that modifies schema or data
    sqlite_pattern = r'sqlite3.*?(?:engineerdna\.db|\.db)'

    if not re.search(sqlite_pattern, command, re.IGNORECASE):
        return False

    # Check for dangerous SQL operations
    dangerous_operations = [
        r'CREATE\s+TABLE',
        r'ALTER\s+TABLE',
        r'DROP\s+TABLE',
        r'CREATE\s+INDEX',
        r'DROP\s+INDEX',
        r'INSERT\s+INTO',
        r'UPDATE\s+',
        r'DELETE\s+FROM',
        r'TRUNCATE',
    ]

    for operation in dangerous_operations:
        if re.search(operation, command, re.IGNORECASE):
            return True

    return False


def main():
    try:
        # Read tool input from stdin
        tool_input = json.load(sys.stdin)
        command = tool_input.get('tool_input', {}).get('command', '')

        if not is_dangerous_sqlite_command(command):
            sys.exit(0)

        # Block the command
        print("", file=sys.stderr)
        print("=" * 80, file=sys.stderr)
        print("[BLOCKED] Direct database modification detected!", file=sys.stderr)
        print("=" * 80, file=sys.stderr)
        print("", file=sys.stderr)
        print("You attempted to run a schema-changing sqlite3 command.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Use migrations instead:", file=sys.stderr)
        print("  1. Create migration file: migrations/00X_description.sql", file=sys.stderr)
        print("  2. Write SQL for your schema/data changes", file=sys.stderr)
        print("  3. Test migration: engineerdna migrate up", file=sys.stderr)
        print("  4. Commit migration file to version control", file=sys.stderr)
        print("", file=sys.stderr)
        print("This ensures all database changes are:", file=sys.stderr)
        print("  - Tracked in version control", file=sys.stderr)
        print("  - Reproducible across environments", file=sys.stderr)
        print("  - Reversible (if you write down migrations)", file=sys.stderr)
        print("  - Documented for other developers", file=sys.stderr)
        print("", file=sys.stderr)
        print("Rule 9 and Rule 27: NO one-off DB scripts, always use migrations", file=sys.stderr)
        print("=" * 80, file=sys.stderr)
        print("", file=sys.stderr)

        sys.exit(2)

    except json.JSONDecodeError:
        # Not valid JSON input, allow
        sys.exit(0)
    except Exception as e:
        print(f"[WARNING] DB operations validation error: {e}", file=sys.stderr)
        # Don't block on hook errors
        sys.exit(0)


if __name__ == '__main__':
    main()
