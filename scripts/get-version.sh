#!/bin/bash
# Get version from git tags, or fallback to CHANGELOG.md, or use "dev"

# Try git describe first (works if there are git tags)
VERSION=$(git describe --tags --always --dirty 2>/dev/null)

# If no git tags, try to extract from CHANGELOG.md (skip "Unreleased")
if [ -z "$VERSION" ] || [[ "$VERSION" == *"-g"* ]]; then
    VERSION=$(grep '^## \[' CHANGELOG.md 2>/dev/null | grep -v 'Unreleased' | head -n 1 | sed -E 's/## \[([0-9.]+)\].*/\1/')
fi

# If still no version, use "dev"
if [ -z "$VERSION" ]; then
    VERSION="dev"
fi

echo "$VERSION"
