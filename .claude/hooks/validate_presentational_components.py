#!/usr/bin/env python3
"""
Validate Container/Presentational Component Pattern

This hook enforces that components in frontend/src/components/
do not contain data fetching logic (useQuery, useMutation, API calls).

Components should be pure presentational - they accept props and render UI.
Pages in frontend/src/pages/ handle all data fetching.
"""

import sys
import json
import re
from pathlib import Path

def main():
    # Read the tool use from stdin
    tool_use = json.loads(sys.stdin.read())

    # Only check Edit and Write operations on frontend components
    if tool_use["type"] not in ["Edit", "Write"]:
        return  # Allow

    file_path = tool_use["params"].get("file_path", "")

    # Only check files in frontend/src/components/
    if not file_path or "frontend/src/components/" not in file_path:
        return  # Allow - not a component file

    # Skip certain files that are allowed to have hooks
    skip_patterns = [
        "/hooks/",  # Custom hooks are OK
        "/contexts/",  # Context providers are OK
        "onboarding-modal.tsx",  # Modal components may need queries
    ]

    if any(pattern in file_path for pattern in skip_patterns):
        return  # Allow

    # Get the content being written/edited
    if tool_use["type"] == "Write":
        content = tool_use["params"].get("content", "")
    else:  # Edit
        new_string = tool_use["params"].get("new_string", "")
        content = new_string

    if not content:
        return  # Allow if no content to check

    # Check for data fetching patterns
    violations = []

    # Pattern 1: useQuery
    if re.search(r'\buseQuery\s*\(', content):
        violations.append("useQuery() hook detected")

    # Pattern 2: useMutation
    if re.search(r'\buseMutation\s*\(', content):
        violations.append("useMutation() hook detected")

    # Pattern 3: Direct API client calls (less common but check anyway)
    if re.search(r'\bapi\.\w+\s*\(', content):
        violations.append("Direct API call detected (api.*)")

    # Pattern 4: fetch() calls
    if re.search(r'\bfetch\s*\(', content):
        violations.append("fetch() call detected")

    if violations:
        error_msg = f"""[BLOCKED] Container/Presentational Pattern Violation

File: {file_path}

Violations found:
{chr(10).join(f"  - {v}" for v in violations)}

Rule: Components in frontend/src/components/ must be PRESENTATIONAL ONLY.

They should:
  [OK] Accept data as props
  [OK] Render UI based on props
  [OK] Use local state (useState) for UI state only
  [OK] Use callbacks passed as props

They should NOT:
  [NO] Use useQuery or useMutation
  [NO] Make API calls directly
  [NO] Fetch data

Solution:
  1. Move data fetching to the parent Page component (frontend/src/pages/)
  2. Pass data as props to this component
  3. Pass callbacks for mutations as props

Example:
  // Page (Container)
  export function DashboardPage() {{
    const {{ data }} = useQuery(...);
    return <Dashboard data={{data}} />;
  }}

  // Component (Presentational)
  export function Dashboard({{ data }}: {{ data: Data }}) {{
    return <div>{{data.value}}</div>;
  }}

See: frontend/CLAUDE.md - Container/Presentational Pattern
"""
        print(error_msg, file=sys.stderr)
        sys.exit(1)

    # Allow - no violations found

if __name__ == "__main__":
    main()
