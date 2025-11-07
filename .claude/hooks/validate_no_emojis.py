#!/usr/bin/env python3
"""
Emoji validator for Claude Code.
Blocks file operations that would introduce emoji characters.
Allows technical characters like box-drawing and mathematical symbols.
"""
import json
import sys
import re

# Emoji pattern - actual emojis only, excluding technical/box-drawing characters
#
# Technical characters that ARE ALLOWED (not emojis):
#   U+2500-U+257F: Box Drawing (├──, └──, │, ─, etc.)
#   U+2580-U+259F: Block Elements (█, ▀, ▄, etc.)
#   U+2190-U+21FF: Arrows (←, →, ↑, ↓, etc.)
#   U+2200-U+22FF: Mathematical Operators (∀, ∃, ∈, ∑, etc.)
#   U+2300-U+23FF: Miscellaneous Technical
#   U+25A0-U+25FF: Geometric Shapes (■, □, ▲, etc.)
#
# These ranges are excluded from the pattern below, so they won't be flagged as emojis.
EMOJI_PATTERN = re.compile(
    "["
    "\U0001F600-\U0001F64F"  # emoticons
    "\U0001F300-\U0001F5FF"  # symbols & pictographs
    "\U0001F680-\U0001F6FF"  # transport & map symbols
    "\U0001F700-\U0001F77F"  # alchemical symbols
    "\U0001F780-\U0001F7FF"  # Geometric Shapes Extended
    "\U0001F800-\U0001F8FF"  # Supplemental Arrows-C
    "\U0001F900-\U0001F9FF"  # Supplemental Symbols and Pictographs
    "\U0001FA00-\U0001FA6F"  # Chess Symbols
    "\U0001FA70-\U0001FAFF"  # Symbols and Pictographs Extended-A
    "\U0001F004"             # Mahjong tile
    "\U0001F0CF"             # Playing card
    "\U0001F18E"             # Negative squared AB
    "\U0001F1E0-\U0001F1FF"  # Regional indicator symbols (flags)
    # Selective ranges that avoid technical characters:
    "\U00002600-\U000026FF"  # Misc symbols (weather, zodiac, etc.)
    "\U00002700-\U000027BF"  # Dingbats (but NOT box-drawing at 2500-257F)
    "]+",
    flags=re.UNICODE
)

def find_emojis(text):
    """Find all emoji characters in text."""
    matches = EMOJI_PATTERN.finditer(text)
    return [(m.group(), m.start(), m.end()) for m in matches]

def get_line_number(text, position):
    """Get line number for a character position."""
    return text[:position].count('\n') + 1

try:
    input_data = json.load(sys.stdin)
    tool_input = input_data.get('tool_input', {})

    # Check different tool types
    content_to_check = None
    file_context = ""

    if 'new_string' in tool_input:  # Edit tool
        content_to_check = tool_input['new_string']
        file_context = f"in Edit to {tool_input.get('file_path', 'unknown file')}"
    elif 'content' in tool_input:  # Write tool
        content_to_check = tool_input['content']
        file_context = f"in Write to {tool_input.get('file_path', 'unknown file')}"
    elif 'command' in tool_input:  # Bash tool
        command = tool_input['command']
        # Check git commits and PR operations
        if any(cmd in command for cmd in ['git commit', 'gh pr create', 'gh pr edit']):
            content_to_check = command
            if 'git commit' in command:
                file_context = "in git commit message"
            elif 'gh pr create' in command:
                file_context = "in PR creation (title/body)"
            elif 'gh pr edit' in command:
                file_context = "in PR edit (title/body)"

    if content_to_check:
        emojis = find_emojis(content_to_check)

        if emojis:
            print("\n[BLOCKED] Emoji characters detected!", file=sys.stderr)
            print("\nProject rules prohibit emojis in all code, documentation, commits, and PRs.", file=sys.stderr)
            print(f"\nFound {len(emojis)} emoji(s) {file_context}:", file=sys.stderr)

            for emoji, start, end in emojis[:5]:  # Show first 5
                line = get_line_number(content_to_check, start)
                print(f"  Position {start}: '{emoji}'", file=sys.stderr)

            if len(emojis) > 5:
                print(f"  ... and {len(emojis) - 5} more", file=sys.stderr)

            print("\nPlease use plain text or markdown instead:", file=sys.stderr)
            print("  - Instead of checkmark emoji: use '- [x]' in markdown", file=sys.stderr)
            print("  - Instead of X emoji: 'Failed', '[ERROR]', or 'FAIL'", file=sys.stderr)
            print("  - Instead of warning emoji: 'Warning', '[WARNING]', or 'WARN'", file=sys.stderr)
            print("  - Instead of rocket emoji: 'Deploying', 'Starting', etc.", file=sys.stderr)

            sys.exit(2)  # Exit code 2 blocks the tool use

    sys.exit(0)  # Allow the operation

except Exception as e:
    print(f"Error validating emojis: {e}", file=sys.stderr)
    sys.exit(1)
