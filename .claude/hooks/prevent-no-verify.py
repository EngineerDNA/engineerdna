#!/usr/bin/env python3
"""
PreToolUse hook to prevent --no-verify in git commands.

Blocks git commands that use --no-verify flag to bypass quality gates.
This enforces that all commits and pushes go through proper validation.
"""

import json
import sys
import re

try:
    input_data = json.load(sys.stdin)
except json.JSONDecodeError as e:
    print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
    sys.exit(1)

tool_name = input_data.get("tool_name", "")
tool_input = input_data.get("tool_input", {})
command = tool_input.get("command", "")

# Only check Bash commands
if tool_name != "Bash" or not command:
    sys.exit(0)

# Check for git commands with --no-verify or -n flag
# Pattern: git <subcommand> <options> where options include --no-verify or -n
# Stops at: pipe (|), semicolon (;), &&, ||, or end of string
# Uses quote-awareness to avoid matching -n inside quoted strings
if re.search(r'\bgit\s+\w+[^|;&]*?\s(--no-verify|-n)(?=\s|$)', command):
    # Additional check: skip if the flag appears inside a quoted string
    # Find the match position and check if it's between quotes
    match = re.search(r'\bgit\s+\w+[^|;&]*?\s(--no-verify|-n)(?=\s|$)', command)
    if match:
        # Extract everything before the flag to count quotes
        before_flag = command[:match.start(1)]
        # Count unescaped quotes
        double_quotes = before_flag.count('"') - before_flag.count('\\"')
        single_quotes = before_flag.count("'") - before_flag.count("\\'")
        # If odd number of quotes, we're inside a string - skip this match
        if double_quotes % 2 == 1 or single_quotes % 2 == 1:
            sys.exit(0)  # Allow the command
    # Flag is outside quotes - block it
    error_message = """
╔════════════════════════════════════════════════════════════════╗
║  ERROR: --no-verify is FORBIDDEN                               ║
╚════════════════════════════════════════════════════════════════╝

You attempted to bypass quality gates with --no-verify.

Command: {command}

Why this is blocked:
  • Pre-commit hooks run tests and linting
  • Pre-push hooks verify build success
  • These checks catch issues before they reach CI
  • Bypassing them defeats the entire purpose

The right way to proceed:
  1. If tests are failing → Fix the tests
  2. If hooks are broken → Fix the hooks
  3. If you disagree with a check → Discuss it with the team

NEVER use --no-verify to work around quality checks.
This is a hard rule. No exceptions.
""".format(command=command)

    print(error_message.strip(), file=sys.stderr)
    sys.exit(2)  # Exit code 2 blocks the tool call and shows stderr to Claude

# Allow the command
sys.exit(0)
