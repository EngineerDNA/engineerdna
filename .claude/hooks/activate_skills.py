#!/usr/bin/env python3
"""
Skills Auto-Activation Hook (UserPromptSubmit)

Analyzes user prompts and suggests relevant skills BEFORE Claude sees the message.
Uses skill-rules.json to map keywords, intent patterns, and context to skills.

This solves the problem of Claude not consistently using available skills.
"""
import json
import sys
import os
import re
from pathlib import Path
from typing import List, Dict, Tuple


def load_skill_rules(project_root: str) -> Dict:
    """Load skill rules from skill-rules.json"""
    rules_path = Path(project_root) / '.claude' / 'hooks' / 'skill-rules.json'

    try:
        with open(rules_path, 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"Warning: skill-rules.json not found at {rules_path}", file=sys.stderr)
        return {}
    except json.JSONDecodeError as e:
        print(f"Warning: Invalid JSON in skill-rules.json: {e}", file=sys.stderr)
        return {}


def check_keyword_match(prompt: str, keywords: List[str]) -> bool:
    """Check if any keywords match in the prompt (case-insensitive)"""
    prompt_lower = prompt.lower()
    return any(keyword.lower() in prompt_lower for keyword in keywords)


def check_intent_patterns(prompt: str, patterns: List[str]) -> bool:
    """Check if any intent patterns match using regex"""
    for pattern in patterns:
        try:
            if re.search(pattern, prompt, re.IGNORECASE):
                return True
        except re.error:
            # Invalid regex, skip
            continue
    return False


def analyze_prompt(prompt: str, skill_rules: Dict) -> List[Tuple[str, str, str]]:
    """
    Analyze prompt and return list of relevant skills.

    Returns: List of (skill_name, priority, reason) tuples
    """
    matches = []

    for skill_name, rules in skill_rules.items():
        priority = rules.get('priority', 'medium')
        enforcement = rules.get('enforcement', 'suggest')
        triggers = rules.get('promptTriggers', {})

        # Check keyword matches
        keywords = triggers.get('keywords', [])
        keyword_match = check_keyword_match(prompt, keywords)

        # Check intent patterns
        intent_patterns = triggers.get('intentPatterns', [])
        intent_match = check_intent_patterns(prompt, intent_patterns)

        # Determine if skill should be suggested
        if keyword_match or intent_match:
            reason = []
            if keyword_match:
                reason.append("keyword match")
            if intent_match:
                reason.append("intent pattern")

            matches.append((
                skill_name,
                priority,
                enforcement,
                " + ".join(reason)
            ))

    return matches


def format_skill_suggestions(matches: List[Tuple[str, str, str, str]]) -> str:
    """Format skill suggestions for display"""
    if not matches:
        return ""

    # Sort by priority (critical > high > medium > low)
    priority_order = {'critical': 0, 'high': 1, 'medium': 2, 'low': 3}
    sorted_matches = sorted(
        matches,
        key=lambda x: priority_order.get(x[1], 4)
    )

    output = [
        "",
        "=" * 80,
        "SKILL ACTIVATION CHECK",
        "=" * 80,
        ""
    ]

    # Group by priority
    by_priority = {}
    for skill_name, priority, enforcement, reason in sorted_matches:
        if priority not in by_priority:
            by_priority[priority] = []
        by_priority[priority].append((skill_name, enforcement, reason))

    # Display each priority group
    for priority in ['critical', 'high', 'medium', 'low']:
        if priority not in by_priority:
            continue

        priority_prefix = {
            'critical': '[CRITICAL]',
            'high': '[HIGH]',
            'medium': '[MEDIUM]',
            'low': '[LOW]'
        }.get(priority, '[INFO]')

        output.append(f"{priority_prefix} {priority.upper()} PRIORITY:")

        for skill_name, enforcement, reason in by_priority[priority]:
            enforcement_text = "MANDATORY" if enforcement == "mandatory" else "Suggested"
            output.append(f"  - {skill_name}")
            output.append(f"    [{enforcement_text}] Matched: {reason}")

        output.append("")

    output.extend([
        "To use a skill: \"Use the [skill-name] skill to...\"",
        "Or let me decide if the skill context is needed.",
        "=" * 80,
        ""
    ])

    return "\n".join(output)


def main():
    try:
        # Read hook input from stdin
        input_data = json.load(sys.stdin)

        # Get user prompt
        user_prompt = input_data.get('prompt', '')

        if not user_prompt:
            # No prompt to analyze
            sys.exit(0)

        # Get project root
        project_root = os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd())

        # Load skill rules
        skill_rules = load_skill_rules(project_root)

        if not skill_rules:
            # No rules to check
            sys.exit(0)

        # Analyze prompt for relevant skills
        matches = analyze_prompt(user_prompt, skill_rules)

        if matches:
            # Print suggestions to stdout (will be shown to Claude)
            suggestions = format_skill_suggestions(matches)
            print(suggestions)

        # Always allow (non-blocking)
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
        sys.exit(0)  # Non-blocking even on errors
    except Exception as e:
        print(f"Error in skill activation: {e}", file=sys.stderr)
        sys.exit(0)  # Non-blocking


if __name__ == "__main__":
    main()
