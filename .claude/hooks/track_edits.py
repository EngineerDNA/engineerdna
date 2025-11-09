#!/usr/bin/env python3
"""
Edit Tracking Hook (PostToolUse)

Tracks which files are edited during a session for build checking.
Runs after Edit, Write, and MultiEdit operations.
"""
import json
import sys
import os
from pathlib import Path
from datetime import datetime


def get_edit_log_path(project_root: str) -> Path:
    """Get path to edit tracking log"""
    return Path(project_root) / '.claude' / 'logs' / 'edit_tracking.json'


def load_edit_log(log_path: Path) -> dict:
    """Load existing edit log"""
    if not log_path.exists():
        return {'edits': [], 'session_start': datetime.now().isoformat()}

    try:
        with open(log_path, 'r') as f:
            return json.load(f)
    except (json.JSONDecodeError, FileNotFoundError):
        return {'edits': [], 'session_start': datetime.now().isoformat()}


def save_edit_log(log_path: Path, log_data: dict):
    """Save edit log to file"""
    log_path.parent.mkdir(parents=True, exist_ok=True)

    with open(log_path, 'w') as f:
        json.dump(log_data, f, indent=2)


def main():
    try:
        # Read hook input from stdin
        input_data = json.load(sys.stdin)

        # Get file path from tool input
        tool_input = input_data.get('tool_input', {})
        file_path = tool_input.get('file_path', '')

        if not file_path:
            # MultiEdit or other operation
            sys.exit(0)

        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Get edit log
        log_path = get_edit_log_path(project_root)
        log_data = load_edit_log(log_path)

        # Record edit
        edit_entry = {
            'file': file_path,
            'timestamp': datetime.now().isoformat(),
            'is_typescript': file_path.endswith(('.ts', '.tsx')),
            'is_go': file_path.endswith('.go')
        }

        log_data['edits'].append(edit_entry)

        # Save updated log
        save_edit_log(log_path, log_data)

        # Always allow (non-blocking)
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
        sys.exit(0)
    except Exception as e:
        print(f"Error tracking edit: {e}", file=sys.stderr)
        sys.exit(0)


if __name__ == "__main__":
    main()
