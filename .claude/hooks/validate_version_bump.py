#!/usr/bin/env python3
"""
Validates that version numbers and CHANGELOG.md are updated when code changes.

Enforces validation in two scenarios:
1. When committing to the main branch
2. When creating a PR (gh pr create)

For EngineerDNA Go binary releases.
"""

import json
import subprocess
import sys
import re
from datetime import datetime
from pathlib import Path


def run_command(cmd: list[str]) -> str:
    """Run a command and return its output."""
    try:
        result = subprocess.run(
            cmd, capture_output=True, text=True, check=True, cwd=Path(__file__).parent.parent.parent
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        return ""


def get_staged_code_files() -> list[str]:
    """Get list of staged files that are source code."""
    output = run_command(["git", "diff", "--cached", "--name-only"])
    if not output:
        return []

    code_paths = [
        "internal/",
        "migrations/",
        "plugins/",
        "frontend/src/",
        "go.mod",
        "go.sum",
        "main.go",
        "Makefile",
        "frontend/package.json",
        "frontend/package-lock.json",
    ]

    return [
        f for f in output.split("\n")
        if any(f.startswith(p) for p in code_paths) or f.endswith((".go", ".ts", ".tsx"))
    ]


def get_branch_code_files(base_branch: str = "main") -> list[str]:
    """Get list of code files changed in current branch compared to base branch."""
    output = run_command(["git", "diff", f"{base_branch}...HEAD", "--name-only"])
    if not output:
        return []

    code_paths = [
        "internal/",
        "migrations/",
        "plugins/",
        "frontend/src/",
        "go.mod",
        "go.sum",
        "main.go",
        "Makefile",
        "frontend/package.json",
        "frontend/package-lock.json",
    ]

    return [
        f for f in output.split("\n")
        if any(f.startswith(p) for p in code_paths) or f.endswith((".go", ".ts", ".tsx"))
    ]


def get_current_version() -> str:
    """Get current version from CHANGELOG.md."""
    try:
        with open("CHANGELOG.md") as f:
            content = f.read()
            # Look for ## [x.y.z] pattern
            match = re.search(r'^## \[(\d+\.\d+\.\d+)\]', content, re.MULTILINE)
            if match:
                return match.group(1)
    except Exception:
        pass
    return ""


def get_previous_version() -> str:
    """Get version from the last commit."""
    output = run_command(["git", "show", "HEAD:CHANGELOG.md"])
    if output:
        match = re.search(r'^## \[(\d+\.\d+\.\d+)\]', output, re.MULTILINE)
        if match:
            return match.group(1)
    return ""


def get_base_branch_version(base_branch: str = "main") -> str:
    """Get version from base branch."""
    output = run_command(["git", "show", f"{base_branch}:CHANGELOG.md"])
    if output:
        match = re.search(r'^## \[(\d+\.\d+\.\d+)\]', output, re.MULTILINE)
        if match:
            return match.group(1)
    return ""


def check_changelog_updated() -> bool:
    """Check if CHANGELOG.md has staged changes."""
    output = run_command(["git", "diff", "--cached", "--name-only"])
    return "CHANGELOG.md" in output.split("\n")


def check_changelog_updated_in_branch(base_branch: str = "main") -> bool:
    """Check if CHANGELOG.md was updated in current branch compared to base branch."""
    output = run_command(["git", "diff", f"{base_branch}...HEAD", "--name-only"])
    return "CHANGELOG.md" in output.split("\n")


def check_hardcoded_versions(expected_version: str) -> list[str]:
    """Check for hardcoded versions that don't match CHANGELOG.md version.

    Returns list of files with mismatched hardcoded versions.
    """
    mismatches = []

    # Check handlers_system.go for hardcoded versions
    handlers_path = Path("internal/api/handlers_system.go")
    if handlers_path.exists():
        content = handlers_path.read_text()
        # Look for hardcoded version strings like "1.0.0" but ignore main.Version references
        hardcoded_pattern = r'"version":\s*"(\d+\.\d+\.\d+)"'
        matches = re.findall(hardcoded_pattern, content)
        for match in matches:
            if match != expected_version:
                mismatches.append(f"internal/api/handlers_system.go (has {match}, expected {expected_version})")

    return mismatches


def categorize_changes(files: list[str]) -> dict[str, list[str]]:
    """Categorize changed files by type."""
    categories = {
        "backend": [],
        "plugins": [],
        "migrations": [],
        "frontend": [],
        "build": [],
    }

    for f in files:
        if f.startswith("internal/") or f == "main.go":
            categories["backend"].append(f)
        elif f.startswith("plugins/"):
            categories["plugins"].append(f)
        elif f.startswith("migrations/"):
            categories["migrations"].append(f)
        elif f.startswith("frontend/"):
            categories["frontend"].append(f)
        elif f in ["go.mod", "go.sum", "Makefile", "frontend/package.json", "frontend/package-lock.json"]:
            categories["build"].append(f)

    return {k: v for k, v in categories.items() if v}


def suggest_version_bump(categories: dict[str, list[str]], current_version: str) -> tuple[str, str]:
    """Suggest version bump type and new version."""
    # Parse current version
    try:
        major, minor, patch = map(int, current_version.split("."))
    except (ValueError, AttributeError):
        return "minor", "0.1.0"

    # Determine bump type based on changes
    has_breaking_changes = "migrations" in categories or any(
        "BREAKING" in f for files in categories.values() for f in files
    )
    has_new_features = len(categories) > 1 or "plugins" in categories

    if has_breaking_changes:
        bump_type = "major"
        new_version = f"{major + 1}.0.0"
    elif has_new_features:
        bump_type = "minor"
        new_version = f"{major}.{minor + 1}.0"
    else:
        bump_type = "patch"
        new_version = f"{major}.{minor}.{patch + 1}"

    return bump_type, new_version


def generate_changelog_template(categories: dict[str, list[str]], bump_type: str) -> str:
    """Generate a CHANGELOG.md template based on changes."""
    lines = []

    if bump_type == "major":
        lines.append("### ⚠️ Breaking Changes")
        lines.append("- TODO: Describe breaking changes")
        lines.append("")

    if "plugins" in categories:
        lines.append("### Added")
        lines.append("- TODO: New plugin features or plugin types")
        lines.append("")

    if "backend" in categories:
        lines.append("### Changed")
        lines.append("- TODO: Backend improvements or API changes")
        lines.append("")

    if "migrations" in categories:
        lines.append("### Database")
        lines.append("- TODO: Database schema changes")
        lines.append("")

    if "frontend" in categories:
        lines.append("### UI")
        lines.append("- TODO: UI improvements or new features")
        lines.append("")

    lines.append("### Fixed")
    lines.append("- TODO: Bug fixes")

    return "\n".join(lines)


def validate_commit_on_main():
    """Validate version bump for commits on main branch."""
    # Get staged code files
    staged_files = get_staged_code_files()

    # If no code files changed, allow commit
    if not staged_files:
        sys.exit(0)

    # Get current and previous versions
    current_version = get_current_version()
    previous_version = get_previous_version()

    # Check if version was bumped
    version_updated = current_version != previous_version

    # Check if CHANGELOG was updated
    changelog_updated = check_changelog_updated()

    # If both updated, check for hardcoded version mismatches
    if version_updated and changelog_updated:
        version_mismatches = check_hardcoded_versions(current_version)
        if version_mismatches:
            print("[ERROR] Hardcoded version mismatches found!", file=sys.stderr)
            print("", file=sys.stderr)
            print(f"CHANGELOG.md version: {current_version}", file=sys.stderr)
            print("", file=sys.stderr)
            print("Files with mismatched hardcoded versions:", file=sys.stderr)
            for mismatch in version_mismatches:
                print(f"  - {mismatch}", file=sys.stderr)
            print("", file=sys.stderr)
            print("Fix: Update hardcoded versions to use main.Version instead:", file=sys.stderr)
            print("", file=sys.stderr)
            print("  // BAD: Hardcoded version", file=sys.stderr)
            print('  respondJSON(w, http.StatusOK, map[string]string{"version": "1.0.0"})', file=sys.stderr)
            print("", file=sys.stderr)
            print("  // GOOD: Use version from server struct", file=sys.stderr)
            print('  respondJSON(w, http.StatusOK, map[string]string{"version": s.version})', file=sys.stderr)
            print("", file=sys.stderr)
            sys.exit(2)
        sys.exit(0)

    # Version bump needed - provide helpful feedback
    categories = categorize_changes(staged_files)
    bump_type, suggested_version = suggest_version_bump(categories, current_version or "0.0.0")
    changelog_template = generate_changelog_template(categories, bump_type)

    print("[NOTICE] Code changed - version bump required!", file=sys.stderr)
    print("", file=sys.stderr)
    print(f"Current version: {current_version or 'none'}", file=sys.stderr)
    print(f"Suggested: {suggested_version} ({bump_type} bump)", file=sys.stderr)
    print("", file=sys.stderr)
    print("Files changed:", file=sys.stderr)
    for category, files in categories.items():
        print(f"  {category}:", file=sys.stderr)
        for f in files[:3]:  # Show first 3 files
            print(f"    - {f}", file=sys.stderr)
        if len(files) > 3:
            print(f"    ... and {len(files) - 3} more", file=sys.stderr)
    print("", file=sys.stderr)

    if not changelog_updated:
        print("[ERROR] CHANGELOG.md not updated", file=sys.stderr)
        print("", file=sys.stderr)
        print("Suggested CHANGELOG entry:", file=sys.stderr)
        print("", file=sys.stderr)
        print(f"## [{suggested_version}] - {datetime.now().strftime('%Y-%m-%d')}", file=sys.stderr)
        print("", file=sys.stderr)
        print(changelog_template, file=sys.stderr)

    print("", file=sys.stderr)
    print("Please update CHANGELOG.md before committing to main.", file=sys.stderr)

    # Exit code 2 blocks the tool call and shows this message to Claude
    sys.exit(2)


def validate_pr_creation(base_branch: str = "main"):
    """Validate version bump when creating a PR."""
    # Get code files changed in this branch compared to base
    branch_files = get_branch_code_files(base_branch)

    # If no code files changed in this branch, allow PR creation
    if not branch_files:
        sys.exit(0)

    # Get current version and base branch version
    current_version = get_current_version()
    base_version = get_base_branch_version(base_branch)

    # Check if version was bumped in this branch
    version_updated = current_version != base_version

    # Check if CHANGELOG was updated in this branch
    changelog_updated = check_changelog_updated_in_branch(base_branch)

    # If both updated, check for hardcoded version mismatches
    if version_updated and changelog_updated:
        version_mismatches = check_hardcoded_versions(current_version)
        if version_mismatches:
            print("[ERROR] Hardcoded version mismatches found in PR!", file=sys.stderr)
            print("", file=sys.stderr)
            print(f"CHANGELOG.md version: {current_version}", file=sys.stderr)
            print("", file=sys.stderr)
            print("Files with mismatched hardcoded versions:", file=sys.stderr)
            for mismatch in version_mismatches:
                print(f"  - {mismatch}", file=sys.stderr)
            print("", file=sys.stderr)
            print("Fix: Update hardcoded versions to use main.Version instead:", file=sys.stderr)
            print("", file=sys.stderr)
            print("  // BAD: Hardcoded version", file=sys.stderr)
            print('  respondJSON(w, http.StatusOK, map[string]string{"version": "1.0.0"})', file=sys.stderr)
            print("", file=sys.stderr)
            print("  // GOOD: Use version from server struct", file=sys.stderr)
            print('  respondJSON(w, http.StatusOK, map[string]string{"version": s.version})', file=sys.stderr)
            print("", file=sys.stderr)
            sys.exit(2)
        sys.exit(0)

    # Version bump needed - provide helpful feedback
    categories = categorize_changes(branch_files)
    bump_type, suggested_version = suggest_version_bump(categories, base_version or "0.0.0")
    changelog_template = generate_changelog_template(categories, bump_type)

    print("[NOTICE] Code changed in this branch - version bump required for PR!", file=sys.stderr)
    print("", file=sys.stderr)
    print(f"Base branch ({base_branch}) version: {base_version or 'none'}", file=sys.stderr)
    print(f"Current branch version: {current_version or 'none'}", file=sys.stderr)
    print(f"Suggested: {suggested_version} ({bump_type} bump)", file=sys.stderr)
    print("", file=sys.stderr)
    print("Files changed in this branch:", file=sys.stderr)
    for category, files in categories.items():
        print(f"  {category}:", file=sys.stderr)
        for f in files[:3]:
            print(f"    - {f}", file=sys.stderr)
        if len(files) > 3:
            print(f"    ... and {len(files) - 3} more", file=sys.stderr)
    print("", file=sys.stderr)

    if not changelog_updated:
        print("[ERROR] CHANGELOG.md not updated in this branch", file=sys.stderr)
        print("", file=sys.stderr)
        print("Suggested CHANGELOG entry:", file=sys.stderr)
        print("", file=sys.stderr)
        print(f"## [{suggested_version}] - {datetime.now().strftime('%Y-%m-%d')}", file=sys.stderr)
        print("", file=sys.stderr)
        print(changelog_template, file=sys.stderr)
        print("", file=sys.stderr)
        print("Add this to CHANGELOG.md and commit to your branch before creating the PR", file=sys.stderr)

    print("", file=sys.stderr)
    print("Please update CHANGELOG.md in your branch before creating the PR.", file=sys.stderr)

    # Exit code 2 blocks the tool call and shows this message to Claude
    sys.exit(2)


def main():
    """Main validation logic."""
    try:
        # Read hook input from stdin
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # Not valid JSON, allow to proceed
        sys.exit(0)

    # Only run for Bash commands
    tool_name = input_data.get("tool_name", "")
    if tool_name != "Bash":
        sys.exit(0)

    # Get the command
    tool_input = input_data.get("tool_input", {})
    command = tool_input.get("command", "")

    # Check if it's a PR creation command
    if "gh pr create" in command:
        # Extract base branch if specified (default to main)
        base_branch = "main"
        if "--base" in command:
            parts = command.split("--base")
            if len(parts) > 1:
                # Extract branch name (handle quotes and spaces)
                base_part = parts[1].strip().split()[0].strip('"\'')
                if base_part:
                    base_branch = base_part
        validate_pr_creation(base_branch)
        return

    # Check if it's a git commit command
    if "git commit" in command:
        # Only enforce on main branch - allow free commits on feature branches
        current_branch = run_command(["git", "branch", "--show-current"])
        if current_branch == "main":
            validate_commit_on_main()
        sys.exit(0)

    # Not a command we care about
    sys.exit(0)


if __name__ == "__main__":
    main()
