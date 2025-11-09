#!/usr/bin/env python3
"""
Enforce use of restart script for server management.

Blocks manual server management commands and directs to use restart script instead.
"""

import json
import sys
import re

def main():
    # Read hook input from stdin
    hook_input = json.load(sys.stdin)

    # Only check Bash tool usage
    if hook_input.get("tool") != "Bash":
        # Allow non-Bash tools
        print(json.dumps({"decision": "allow"}))
        return

    # Get the command being run
    command = hook_input.get("parameters", {}).get("command", "")

    # Patterns to block
    blocked_patterns = [
        r'\.?/?(bin/)?engineerdna\s+serve',  # ./engineerdna serve, bin/engineerdna serve, etc.
        r'go\s+run\s+.*\s+serve',             # go run . serve, go run main.go serve
        r'npm\s+run\s+dev\s*$',              # npm run dev (in frontend dir)
    ]

    # Check if command matches any blocked pattern
    for pattern in blocked_patterns:
        if re.search(pattern, command, re.IGNORECASE):
            # Determine the appropriate error message
            if 'engineerdna' in command.lower() and 'serve' in command.lower():
                error_msg = (
                    "BLOCKED: Do not run 'engineerdna serve' manually.\n\n"
                    "Use the restart script instead:\n"
                    "  ./scripts/restart-dev-server.sh              # Basic restart\n"
                    "  ./scripts/restart-dev-server.sh --rebuild    # Rebuild and restart\n"
                    "  ./scripts/restart-dev-server.sh --frontend   # Include frontend dev server\n\n"
                    "Why? The restart script ensures:\n"
                    "  - Clean process cleanup (no zombie processes)\n"
                    "  - Port conflict resolution (automatically kills port 3847)\n"
                    "  - Consistent logging (/tmp/engineerdna-backend.log)\n"
                    "  - Server health verification\n\n"
                    f"Your command: {command}"
                )
            elif 'npm run dev' in command.lower():
                error_msg = (
                    "BLOCKED: Do not run 'npm run dev' manually.\n\n"
                    "Use the restart script with --frontend flag:\n"
                    "  ./scripts/restart-dev-server.sh --rebuild --frontend\n\n"
                    "This ensures both backend and frontend are managed together.\n\n"
                    f"Your command: {command}"
                )
            else:
                error_msg = (
                    "BLOCKED: Manual server management detected.\n\n"
                    "Use the restart script:\n"
                    "  ./scripts/restart-dev-server.sh\n\n"
                    f"Your command: {command}"
                )

            # Block the command
            print(json.dumps({
                "decision": "deny",
                "message": error_msg
            }))
            return

    # Allow all other Bash commands
    print(json.dumps({"decision": "allow"}))

if __name__ == "__main__":
    main()
