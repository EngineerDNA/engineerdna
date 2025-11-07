#!/usr/bin/env python3
"""
Automatic Dev Docs Management (Stop Hook)

Automatically manages PLAN.md, CONTEXT.md, TASKS.md based on conversation state:
- Detects when planning is complete and creates files
- Updates TASKS.md when tasks are completed
- Updates CONTEXT.md with decisions and next steps
- Prepares for compaction when context is high

Everything is automatic - no manual commands needed.
"""
import json
import sys
import os
import re
import time
from pathlib import Path
from datetime import datetime


def log_metric(name: str, value: float, success: bool = True, metadata: dict = None):
    """Log metrics for monitoring"""
    try:
        log_dir = Path(os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())) / '.claude' / 'logs'
        log_dir.mkdir(parents=True, exist_ok=True)
        log_file = log_dir / 'hook_metrics.jsonl'

        metric = {
            'timestamp': datetime.now().isoformat(),
            'hook': 'manage_dev_docs',
            'name': name,
            'value': value,
            'success': success,
            'metadata': metadata or {}
        }

        with open(log_file, 'a') as f:
            f.write(json.dumps(metric) + '\n')
    except Exception:
        # Non-fatal - metrics logging failures don't stop execution
        pass


def get_dev_docs_paths(project_root: str) -> dict:
    """Get paths to dev docs files"""
    root = Path(project_root)
    return {
        'plan': root / 'PLAN.md',
        'context': root / 'CONTEXT.md',
        'tasks': root / 'TASKS.md',
        'state': root / '.claude' / 'logs' / 'dev_docs_state.json'
    }


def load_state(state_path: Path) -> dict:
    """Load dev docs state"""
    if not state_path.exists():
        return {
            'active': False,
            'plan_created': False,
            'last_update': None,
            'tasks_completed': [],
            'feature_name': None
        }

    try:
        with open(state_path, 'r') as f:
            return json.load(f)
    except (json.JSONDecodeError, FileNotFoundError):
        return {'active': False, 'plan_created': False, 'feature_name': None}


def save_state(state_path: Path, state: dict):
    """Save dev docs state"""
    state_path.parent.mkdir(parents=True, exist_ok=True)
    with open(state_path, 'w') as f:
        json.dump(state, f, indent=2)


def extract_feature_name(response: str) -> str:
    """Extract feature name from plan response"""
    # Look for feature name in common patterns
    patterns = [
        r'##\s*(.+?)\s*Plan',
        r'#\s*(.+?)\s*Implementation',
        r'Feature:\s*(.+?)(?:\n|$)',
        r'Building:\s*(.+?)(?:\n|$)',
        r'Implementing:\s*(.+?)(?:\n|$)',
    ]

    for pattern in patterns:
        match = re.search(pattern, response, re.IGNORECASE)
        if match:
            return match.group(1).strip()

    # Fallback: use first heading
    match = re.search(r'^#\s*(.+?)$', response, re.MULTILINE)
    if match:
        return match.group(1).strip()

    return "Unknown Feature"


def detect_plan_in_response(response: str) -> tuple[bool, int, list]:
    """
    Detect if response contains a comprehensive plan.

    Returns: (is_plan, match_count, matched_indicators)
    """
    # Check for explicit plan markers first
    if '[PLAN:START]' in response and '[PLAN:END]' in response:
        return (True, 999, ['explicit_plan_markers'])

    # Look for planning indicators
    plan_indicators = {
        'overview': r'##\s*(Overview|Objectives|Goals|Plan)',
        'phases': r'##\s*(Phase|Step)\s+\d+',
        'tasks': r'###\s*Tasks?',
        'timeline': r'\*\*Timeline\*\*',
        'metrics': r'\*\*Success\s+Metrics?\*\*',
        'confidence': r'Confidence\s+Score',
        'risks': r'\*\*Risks?\*\*',
        'dependencies': r'\*\*Dependencies\*\*',
    }

    matched = []
    for name, pattern in plan_indicators.items():
        if re.search(pattern, response, re.IGNORECASE):
            matched.append(name)

    # Require at least 4 indicators for safety (was 3)
    # This reduces false positives while still catching real plans
    threshold = int(os.environ.get('PLAN_DETECTION_THRESHOLD', '4'))
    is_plan = len(matched) >= threshold

    return (is_plan, len(matched), matched)


def extract_plan_content(response: str) -> str:
    """Extract plan content from response"""
    # Simple extraction - take everything after first heading
    lines = response.split('\n')
    start_idx = 0

    for i, line in enumerate(lines):
        if line.strip().startswith('#'):
            start_idx = i
            break

    return '\n'.join(lines[start_idx:])


def cleanup_old_archives(archive_root: Path, max_archives: int = 50):
    """Keep only the N most recent archives to prevent unbounded growth"""
    if not archive_root.exists():
        return

    archives = sorted(archive_root.glob('*'), key=lambda p: p.stat().st_mtime, reverse=True)

    if len(archives) <= max_archives:
        return

    # Remove oldest archives
    for old_archive in archives[max_archives:]:
        try:
            if old_archive.is_dir():
                import shutil
                shutil.rmtree(old_archive)
        except Exception:
            # Non-fatal - continue cleanup
            pass


def archive_existing_docs(paths: dict, feature_name: str) -> str:
    """
    Archive existing dev docs to .claude/logs/archive/
    Returns archive directory path
    """
    # Create archive directory with timestamp
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    safe_feature_name = re.sub(r'[^\w\-]', '_', feature_name)[:50]
    archive_root = paths['plan'].parent / '.claude' / 'logs' / 'archive'
    archive_dir = archive_root / f"{timestamp}_{safe_feature_name}"
    archive_dir.mkdir(parents=True, exist_ok=True)

    # Move existing files to archive
    archived_files = []
    for doc_type, path in [('plan', paths['plan']), ('context', paths['context']), ('tasks', paths['tasks'])]:
        if path.exists():
            archive_path = archive_dir / path.name
            path.rename(archive_path)
            archived_files.append(path.name)

    # Cleanup old archives (keep 50 most recent)
    cleanup_old_archives(archive_root, max_archives=50)

    return str(archive_dir) if archived_files else None


def is_different_feature(current_feature: str, new_feature: str) -> bool:
    """
    Check if new feature is different from current feature.
    Uses fuzzy matching to handle variations.
    """
    if not current_feature or not new_feature:
        return True

    # Normalize for comparison
    current = current_feature.lower().strip()
    new = new_feature.lower().strip()

    # Exact match
    if current == new:
        return False

    # Check if one contains the other (handles variations)
    if current in new or new in current:
        return False

    # Different features
    return True


def create_initial_plan(plan_path: Path, content: str):
    """Create PLAN.md from detected plan"""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    plan_content = f"""# Implementation Plan

**Created:** {timestamp}
**Status:** Active

---

{content}

---

**Notes:**
- This file is automatically managed by Claude Code
- It will be gitignored (temporary workflow file)
- Update confidence scores as you learn more
"""

    with open(plan_path, 'w') as f:
        f.write(plan_content)


def create_initial_context(context_path: Path):
    """Create initial CONTEXT.md with section markers"""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    context_content = f"""# Technical Context

**Created:** {timestamp}
**Last Updated:** {timestamp}

## Architecture Overview
[Auto-populated as work begins]

## Key Files
[Auto-populated as files are identified]

## Dependencies
[Auto-populated as dependencies are discovered]

## Data Models
[Auto-populated as database/types are worked on]

## Integration Points
[Auto-populated as integrations are built]

<!-- DESIGN_DECISIONS_START -->
## Design Decisions
[Auto-populated as decisions are made]
<!-- DESIGN_DECISIONS_END -->

## Assumptions
[Auto-populated as assumptions are identified]

## Open Questions
[Auto-populated as questions arise]

<!-- NEXT_STEPS_START -->
## Next Steps
Begin implementation following PLAN.md
<!-- NEXT_STEPS_END -->

---

**Notes:**
- This file is automatically updated by Claude Code
- It preserves context across compaction
- Review "Next Steps" after compaction to resume work
- HTML comments are section markers (do not remove)
"""

    with open(context_path, 'w') as f:
        f.write(context_content)


def create_initial_tasks(tasks_path: Path, plan_content: str):
    """Create initial TASKS.md by extracting tasks from plan"""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    # Try to extract tasks from plan
    tasks = []
    in_task_section = False

    for line in plan_content.split('\n'):
        if re.search(r'###?\s*Tasks?', line, re.IGNORECASE):
            in_task_section = True
            continue

        if in_task_section:
            if line.strip().startswith(('- ', '* ', '1. ', '2. ')):
                # Convert to unchecked task
                task = re.sub(r'^[\s\-\*\d.]+', '', line).strip()
                tasks.append(f"- [ ] {task}")
            elif line.strip().startswith('#'):
                # New section, stop extracting
                in_task_section = False

    tasks_content = f"""# Tasks Checklist

**Created:** {timestamp}
**Last Updated:** {timestamp}

## Implementation Tasks

{chr(10).join(tasks) if tasks else "- [ ] Tasks will be added as implementation begins"}

---

**Notes:**
- Use `- [ ]` for pending tasks
- Use `- [x]` for completed tasks
- This file is automatically updated by Claude Code
- Mark tasks complete immediately after finishing
"""

    with open(tasks_path, 'w') as f:
        f.write(tasks_content)


def detect_task_completion(response: str) -> list:
    """Detect completed tasks mentioned in response"""
    completed_indicators = [
        r'completed?\s+(?:the\s+)?(.+?)(?:\.|$)',
        r'finished?\s+(?:the\s+)?(.+?)(?:\.|$)',
        r'implemented?\s+(?:the\s+)?(.+?)(?:\.|$)',
        r'created?\s+(?:the\s+)?(.+?)(?:\.|$)',
        r'added?\s+(?:the\s+)?(.+?)(?:\.|$)',
    ]

    completed = []
    for pattern in completed_indicators:
        matches = re.findall(pattern, response, re.IGNORECASE)
        completed.extend(matches)

    return completed[:5]  # Limit to 5 to avoid noise


def update_context_with_decisions(context_path: Path, response: str):
    """Update CONTEXT.md with design decisions from response using markers"""
    if not context_path.exists():
        return

    # Look for decision-making language
    decision_indicators = [
        r'decided?\s+to\s+(.+?)(?:\.|$)',
        r'chose\s+(.+?)(?:\.|$)',
        r'using\s+(.+?)\s+because',
        r'pattern:\s+(.+?)(?:\.|$)',
    ]

    decisions = []
    for pattern in decision_indicators:
        matches = re.findall(pattern, response, re.IGNORECASE)
        decisions.extend(matches)

    if not decisions:
        return

    # Read existing content
    with open(context_path, 'r') as f:
        content = f.read()

    # Use markers for safer updates
    if '<!-- DESIGN_DECISIONS_START -->' not in content:
        return  # No markers, skip update

    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    # Extract current decisions between markers
    decision_match = re.search(
        r'<!-- DESIGN_DECISIONS_START -->\s*## Design Decisions\s*(.*?)\s*<!-- DESIGN_DECISIONS_END -->',
        content,
        re.DOTALL
    )

    if decision_match:
        current_decisions = decision_match.group(1).strip()

        # Add new decisions (limit to 3 most recent)
        new_decision_text = '\n'.join(f"- {d.strip()}" for d in decisions[:3])

        # Combine with existing (if not placeholder)
        if current_decisions and current_decisions != '[Auto-populated as decisions are made]':
            decision_text = f"{current_decisions}\n{new_decision_text}"
        else:
            decision_text = new_decision_text

        # Replace content between markers
        content = re.sub(
            r'(<!-- DESIGN_DECISIONS_START -->\s*## Design Decisions\s*).*?(\s*<!-- DESIGN_DECISIONS_END -->)',
            f'\\1\n{decision_text}\n\\2',
            content,
            flags=re.DOTALL
        )

        # Update timestamp
        content = re.sub(
            r'\*\*Last Updated:\*\* .+',
            f'**Last Updated:** {timestamp}',
            content
        )

        with open(context_path, 'w') as f:
            f.write(content)


def update_next_steps(context_path: Path, response: str):
    """Update Next Steps in CONTEXT.md using markers"""
    if not context_path.exists():
        return

    # Look for next step indicators
    next_step_patterns = [
        r'[Nn]ext,?\s+(.+?)(?:\.|$)',
        r'[Nn]ow\s+(?:we\s+)?(?:need\s+to\s+)?(.+?)(?:\.|$)',
        r'[Tt]hen\s+(.+?)(?:\.|$)',
    ]

    next_steps = []
    for pattern in next_step_patterns:
        matches = re.findall(pattern, response, re.IGNORECASE)
        next_steps.extend(matches)

    if not next_steps:
        return

    # Read existing content
    with open(context_path, 'r') as f:
        content = f.read()

    # Use markers for safer updates
    if '<!-- NEXT_STEPS_START -->' not in content:
        return  # No markers, skip update

    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    next_steps_text = '\n'.join(f"{i+1}. {step.strip()}" for i, step in enumerate(next_steps[:5]))

    # Replace content between markers
    content = re.sub(
        r'(<!-- NEXT_STEPS_START -->\s*## Next Steps\s*).*?(\s*<!-- NEXT_STEPS_END -->)',
        f'\\1\n{next_steps_text}\n\\2',
        content,
        flags=re.DOTALL
    )

    # Update timestamp
    content = re.sub(
        r'\*\*Last Updated:\*\* .+',
        f'**Last Updated:** {timestamp}',
        content
    )

    with open(context_path, 'w') as f:
        f.write(content)


def main():
    start_time = datetime.now()

    try:
        # Read hook input
        input_data = json.load(sys.stdin)

        # Get my response
        assistant_response = input_data.get('assistant_response', {})
        content = assistant_response.get('content', '')

        if not content:
            sys.exit(0)

        # Extract text from content blocks
        response_text = ''
        if isinstance(content, list):
            for block in content:
                if isinstance(block, dict) and block.get('type') == 'text':
                    response_text += block.get('text', '')
        elif isinstance(content, str):
            response_text = content

        if not response_text:
            sys.exit(0)

        # Get project root and paths
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())
        paths = get_dev_docs_paths(project_root)

        # Load state
        state = load_state(paths['state'])

        # Check if plan was just created
        is_plan, match_count, matched_indicators = detect_plan_in_response(response_text)

        if is_plan:
            # Log plan detection
            log_metric('plan_detected', match_count, True, {
                'matched_indicators': matched_indicators,
                'explicit_markers': '[PLAN:START]' in response_text
            })

            new_feature_name = extract_feature_name(response_text)
            current_feature = state.get('feature_name')

            # Check if this is a different feature than current
            if state['plan_created'] and is_different_feature(current_feature, new_feature_name):
                print("\n" + "=" * 80)
                print("NEW FEATURE DETECTED - ARCHIVING OLD DEV DOCS")
                print("=" * 80)
                print(f"\nCurrent feature: {current_feature}")
                print(f"New feature: {new_feature_name}")
                print("\nArchiving existing dev docs...")

                # Archive old docs
                archive_start = datetime.now()
                archive_path = archive_existing_docs(paths, current_feature)
                archive_time = (datetime.now() - archive_start).total_seconds()

                if archive_path:
                    print(f"Archived to: {archive_path}")
                    log_metric('archive_created', archive_time, True, {
                        'old_feature': current_feature,
                        'new_feature': new_feature_name,
                        'archive_path': str(archive_path)
                    })

                    # Run archive cleanup
                    try:
                        cleanup_old_archives(archive_path.parent)
                        log_metric('archive_cleanup', 1, True)
                    except Exception as e:
                        print(f"\n[WARNING] Archive cleanup failed: {e}", file=sys.stderr)
                        log_metric('archive_cleanup', 0, False, {'error': str(e)})
                else:
                    print("\n[WARNING] Archive creation failed", file=sys.stderr)
                    log_metric('archive_created', archive_time, False, {
                        'old_feature': current_feature,
                        'new_feature': new_feature_name
                    })

                # Reset state for new feature
                state['plan_created'] = False
                state['active'] = False
                state['feature_name'] = None
                print("=" * 80 + "\n")

            # Create new dev docs if not already created for this feature
            if not state['plan_created']:
                print("\n" + "=" * 80)
                print("AUTOMATIC DEV DOCS CREATION")
                print("=" * 80)
                print(f"\nFeature: {new_feature_name}")
                print("Creating PLAN.md, CONTEXT.md, TASKS.md automatically...")

                try:
                    # Create all three files
                    plan_content = extract_plan_content(response_text)
                    create_initial_plan(paths['plan'], plan_content)
                    create_initial_context(paths['context'])
                    create_initial_tasks(paths['tasks'], plan_content)

                    # Update state
                    state['plan_created'] = True
                    state['active'] = True
                    state['feature_name'] = new_feature_name
                    state['last_update'] = datetime.now().isoformat()
                    save_state(paths['state'], state)

                    print("\nDev docs created:")
                    print("  - PLAN.md (implementation strategy)")
                    print("  - CONTEXT.md (technical context)")
                    print("  - TASKS.md (actionable checklist)")
                    print("\nThese files will be automatically updated as work progresses.")
                    print("=" * 80 + "\n")

                    log_metric('dev_docs_created', 1, True, {
                        'feature_name': new_feature_name
                    })
                except Exception as e:
                    print(f"\n[WARNING] Dev docs creation failed: {e}", file=sys.stderr)
                    log_metric('dev_docs_created', 0, False, {
                        'feature_name': new_feature_name,
                        'error': str(e)
                    })

        # If dev docs are active, update them
        elif state['active']:
            try:
                # Update CONTEXT.md with decisions
                update_context_with_decisions(paths['context'], response_text)

                # Update Next Steps
                update_next_steps(paths['context'], response_text)

                # Update state
                state['last_update'] = datetime.now().isoformat()
                save_state(paths['state'], state)

                log_metric('context_updated', 1, True)
            except Exception as e:
                print(f"\n[WARNING] Context update failed: {e}", file=sys.stderr)
                log_metric('context_updated', 0, False, {'error': str(e)})

        # Log total execution time
        execution_time = (datetime.now() - start_time).total_seconds()
        log_metric('total_execution_time', execution_time, True)

        # Always allow (non-blocking)
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"\n[ERROR] Invalid JSON input: {e}", file=sys.stderr)
        log_metric('hook_execution', 0, False, {'error': 'json_decode_error', 'message': str(e)})
        sys.exit(0)
    except Exception as e:
        print(f"\n[ERROR] Dev docs hook failed: {e}", file=sys.stderr)
        log_metric('hook_execution', 0, False, {'error': 'unexpected_exception', 'message': str(e)})
        sys.exit(0)


if __name__ == "__main__":
    main()
