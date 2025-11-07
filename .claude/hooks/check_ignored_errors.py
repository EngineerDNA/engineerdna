#!/usr/bin/env python3
"""
SubagentStop hook: Check if agent ignored tool errors
Blocks continuation when errors are detected but not handled

Security: No file operations, only JSON analysis
"""

import json
import sys
from typing import Any, Dict, List


def detect_error_in_result(result: Any) -> bool:
    """Check if tool result contains error indicators"""
    if not result:
        return False

    result_str = str(result).lower()

    error_indicators = [
        'error', 'failed', 'failure', 'not found',
        'does not exist', 'cannot', 'unable to',
        'invalid', 'exception', 'errno', 'permission denied'
    ]

    return any(indicator in result_str for indicator in error_indicators)


def was_error_handled(error_call_index: int, all_calls: List[Dict]) -> bool:
    """
    Check if subsequent tool calls addressed the error

    Heuristics:
    - Read/Grep after error (investigating)
    - Same tool retried (fixing and retrying)
    - Different approach attempted
    """
    if error_call_index >= len(all_calls) - 1:
        # Error was last call, no handling possible yet
        return False

    error_tool = all_calls[error_call_index].get('tool', '')

    # Check next few tool calls for error handling signs
    next_calls = all_calls[error_call_index + 1:error_call_index + 4]

    for call in next_calls:
        tool = call.get('tool', '')

        # Investigation tools suggest handling
        if tool in ['Read', 'Grep', 'Glob']:
            return True

        # Retry of same tool suggests fix attempt
        if tool == error_tool:
            return True

    return False


def main():
    try:
        # Read input from stdin
        input_data = json.load(sys.stdin)

        # Extract tool calls from agent execution
        tool_calls = input_data.get('tool_calls', [])

        if not tool_calls:
            # No tool calls, nothing to check
            sys.exit(0)

        # Find all tool calls with errors
        errors_found = []

        for idx, call in enumerate(tool_calls):
            tool_name = call.get('tool', 'Unknown')
            result = call.get('result', '')

            if detect_error_in_result(result):
                # Check if this error was handled
                if not was_error_handled(idx, tool_calls):
                    # Extract first 300 chars of error for display
                    error_snippet = str(result)[:300]
                    if len(str(result)) > 300:
                        error_snippet += "..."

                    errors_found.append({
                        'tool': tool_name,
                        'error': error_snippet,
                        'call_index': idx
                    })

        if errors_found:
            # Build error message
            error_msg = "ERROR HANDLING VIOLATION\n\n"
            error_msg += "Tool errors were detected but not properly handled.\n\n"
            error_msg += "Errors found:\n"

            for err in errors_found:
                error_msg += f"\n{err['tool']} (call #{err['call_index'] + 1}):\n"
                error_msg += f"  {err['error']}\n"

            error_msg += "\nREQUIRED PATTERN:\n"
            error_msg += "1. Tool returns error\n"
            error_msg += "2. STOP immediately\n"
            error_msg += "3. Investigate (Read files, Grep for context)\n"
            error_msg += "4. Fix the issue\n"
            error_msg += "5. Retry the tool\n"
            error_msg += "6. Verify success\n"
            error_msg += "\nPROHIBITED:\n"
            error_msg += "- Continuing after errors without fixing\n"
            error_msg += "- Ignoring tool failures\n"
            error_msg += "- Reporting success when tools failed\n"

            # Write error to stderr and exit with code 2 to block
            print(error_msg, file=sys.stderr)
            sys.exit(2)

        # No errors or all were handled
        sys.exit(0)

    except json.JSONDecodeError:
        # Invalid JSON input, pass through
        sys.exit(0)
    except Exception as e:
        # Hook error, don't block agent
        print(f"Hook error (non-blocking): {e}", file=sys.stderr)
        sys.exit(0)


if __name__ == '__main__':
    main()
