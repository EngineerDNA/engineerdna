#!/bin/bash
# Prevents commits with development markers like NEW:, FIXME, HACK, etc.

# Check for development markers in staged changes (excluding hook config files)
if git diff --cached --diff-filter=ACM -- ':!.claude/hooks/' ':!.claude/settings.json' | grep -E "^\+.*\b(NEW:|FIXME|HACK|XXX|TEMP:|WIP:|DEBUG:|@deprecated)" > /dev/null 2>&1 ; then
    echo "[BLOCKED] COMMIT BLOCKED: Found development markers in staged changes" >&2
    echo "" >&2
    echo "Found markers that should be removed before committing:" >&2
    echo "  - NEW:         Temporary marker for new code" >&2
    echo "  - FIXME        Code that needs fixing" >&2
    echo "  - HACK         Temporary workaround" >&2
    echo "  - XXX          Warning marker" >&2
    echo "  - TEMP:        Temporary code" >&2
    echo "  - WIP:         Work in progress" >&2
    echo "  - DEBUG:       Debug code" >&2
    echo "  - @deprecated  Use proper deprecation with migration path" >&2
    echo "" >&2
    echo "Please remove these markers and commit again." >&2
    echo "" >&2
    # Exit code 2 blocks the tool call and shows stderr to Claude
    exit 2
fi

exit 0
