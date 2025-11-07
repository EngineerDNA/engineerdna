#!/usr/bin/env python3
"""
Validate source file size to enforce 500 LOC limit.
Blocks creation/editing of files exceeding reasonable size limits.
"""
import json
import sys
import os
from pathlib import Path


MAX_LINES = 500

# File extensions to check
CHECKED_EXTENSIONS = {
    '.ts', '.tsx', '.js', '.jsx',  # TypeScript/JavaScript
    '.py',                          # Python
    '.go',                          # Go
    '.rs',                          # Rust
    '.java',                        # Java
    '.rb',                          # Ruby
    '.php',                         # PHP
    '.swift',                       # Swift
    '.kt',                          # Kotlin
}

# Paths to exclude from size checks
EXCLUDED_PATHS = [
    'node_modules/',
    '.next/',
    'dist/',
    'build/',
    '.turbo/',
    'coverage/',
    # Generated files
    'generated/',
    'migrations/',  # Database migrations can be large
    '__generated__',
    '/i18n/types.ts',  # Auto-generated translation types
    # Test fixtures and mocks
    'fixtures/',
    'mocks/',
    # Configuration files
    'jest.config',
    'webpack.config',
    'rollup.config',
    'tailwind.config',
    'postcss.config',
    'next.config',
    # Type definitions
    '.d.ts',
]

# Special patterns
EXCLUDED_PATTERNS = [
    lambda p: p.endswith('.js') and os.path.exists(p.replace('.js', '.ts')),  # Compiled JS with TS source
    lambda p: p.endswith('.js') and os.path.exists(p.replace('.js', '.tsx')),  # Compiled JS with TSX source
    lambda p: 'infrastructure/lib/' in p and p.endswith('.js'),  # All CDK compiled output (gitignored)
]


def is_excluded_path(file_path: str) -> bool:
    """Check if file path should be excluded from size validation."""
    # Check static paths
    for excluded in EXCLUDED_PATHS:
        if excluded in file_path:
            return True

    # Check dynamic patterns
    for pattern_check in EXCLUDED_PATTERNS:
        if pattern_check(file_path):
            return True

    return False


def count_lines(file_path: str) -> int:
    """Count non-empty lines in a file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
            # Count non-empty, non-whitespace-only lines
            return sum(1 for line in lines if line.strip())
    except Exception as e:
        # If we can't read the file, assume it's okay
        print(f"Warning: Could not read {file_path}: {e}", file=sys.stderr)
        return 0


def validate_file_size(file_path: str, project_root: str) -> tuple[bool, str]:
    """
    Check if source file exceeds size limit.

    Returns: (is_valid, error_message_if_invalid)
    """
    # Convert to Path for easier manipulation
    path = Path(file_path)

    # Check if it's a file type we care about
    if path.suffix not in CHECKED_EXTENSIONS:
        return (True, "")

    # Convert to relative path from project root if absolute
    if path.is_absolute():
        try:
            rel_path = path.relative_to(project_root)
        except ValueError:
            rel_path = path
    else:
        rel_path = path

    rel_path_str = str(rel_path).replace('\\', '/')

    # Check exclusions
    if is_excluded_path(rel_path_str):
        return (True, "")

    # For new files (Write operation), check if file exists
    # If it doesn't exist yet, we can't count lines, so allow it
    if not path.exists():
        return (True, "")

    # Count lines
    line_count = count_lines(str(path))

    if line_count > MAX_LINES:
        return (False, f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║ FILE SIZE LIMIT EXCEEDED (Rule 16)                                          ║
╠══════════════════════════════════════════════════════════════════════════════╣
║ File: {rel_path_str:<73}║
║ Lines: {line_count} / {MAX_LINES} (exceeded by {line_count - MAX_LINES} lines){' ' * (73 - len(str(line_count)) - len(str(MAX_LINES)) - len(str(line_count - MAX_LINES)) - 33)}║
╠══════════════════════════════════════════════════════════════════════════════╣
║ WHY: Large files are harder to review, test, and maintain.                  ║
║                                                                              ║
║ SOLUTIONS:                                                                   ║
║  1. Extract shared logic into utility functions/modules                     ║
║  2. Split endpoints into separate route files by concern                    ║
║  3. Create reusable middleware for common patterns                          ║
║  4. Move complex logic to service layer                                     ║
║  5. Break into multiple focused files with single responsibility            ║
║                                                                              ║
║ EXAMPLE REFACTORING:                                                         ║
║  routes/admin/analytics.ts (1065 lines)                                     ║
║  ├── routes/admin/analytics/overview.ts (~200 lines)                        ║
║  ├── routes/admin/analytics/providers.ts (~200 lines)                       ║
║  ├── routes/admin/analytics/real-time.ts (~150 lines)                       ║
║  ├── routes/admin/analytics/attribution.ts (~150 lines)                     ║
║  ├── routes/admin/analytics/export.ts (~100 lines)                          ║
║  └── lib/analytics-cache.ts (~100 lines - shared caching logic)             ║
║                                                                              ║
║ If this file MUST be large (generated code, etc.), add to EXCLUDED_PATHS    ║
║ in .claude/hooks/validate_file_size.py                                      ║
╚══════════════════════════════════════════════════════════════════════════════╝
""")

    return (True, "")


def main():
    """Main hook entry point."""
    try:
        # Read hook input from stdin
        input_data = sys.stdin.read()
        if not input_data.strip():
            print(json.dumps({"allowed": True}))
            return

        data = json.loads(input_data)

        # Get file path from tool parameters
        tool_name = data.get("tool", {}).get("name", "")
        params = data.get("tool", {}).get("params", {})

        # Only validate for Edit and Write operations
        if tool_name not in ["Edit", "Write"]:
            print(json.dumps({"allowed": True}))
            return

        file_path = params.get("file_path")
        if not file_path:
            print(json.dumps({"allowed": True}))
            return

        # Get project root from environment
        project_root = os.environ.get("CLAUDE_PROJECT_ROOT", os.getcwd())

        # Validate file size
        is_valid, error_msg = validate_file_size(file_path, project_root)

        if not is_valid:
            print(json.dumps({
                "allowed": False,
                "message": error_msg
            }))
            sys.exit(1)

        print(json.dumps({"allowed": True}))

    except Exception as e:
        # On error, allow the operation but log the issue
        print(f"Error in validate_file_size hook: {e}", file=sys.stderr)
        print(json.dumps({"allowed": True}))


if __name__ == "__main__":
    main()
