#!/bin/bash
# Hook to prevent direct commits to main branch
# This ensures all changes go through PR review process

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

if [ "$CURRENT_BRANCH" = "main" ]; then
    echo "BLOCKED: Direct commit to main branch detected!"
    echo ""
    echo "You are attempting to commit directly to the main branch."
    echo ""
    echo "Please use a feature branch instead:"
    echo "  1. git checkout -b feature/your-feature-name"
    echo "  2. Make your changes and commit"
    echo "  3. git push -u origin feature/your-feature-name"
    echo "  4. Create a PR: gh pr create"
    echo ""
    echo "This ensures all changes are reviewed before merging to main."
    echo ""

    exit 1
fi

exit 0
