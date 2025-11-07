#!/bin/bash
# SubagentStop hook: Log agent completion metrics for monitoring
#
# This hook fires when any agent completes (success or failure)
# Logs to .claude/logs/agent-metrics.log for performance tracking

set -euo pipefail

# Create logs directory if it doesn't exist
LOGS_DIR="${CLAUDE_PROJECT_DIR}/.claude/logs"
mkdir -p "$LOGS_DIR"

LOG_FILE="${LOGS_DIR}/agent-metrics.log"

# Read hook input from stdin (JSON)
HOOK_INPUT=$(cat)

# Extract agent info if available
AGENT_NAME=$(echo "$HOOK_INPUT" | jq -r '.agent_name // "unknown"' 2>/dev/null || echo "unknown")
STOP_HOOK_ACTIVE=$(echo "$HOOK_INPUT" | jq -r '.stop_hook_active // "false"' 2>/dev/null || echo "false")

# Log timestamp and agent completion
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Simple log entry
echo "${TIMESTAMP} | Agent: ${AGENT_NAME} | StopHookActive: ${STOP_HOOK_ACTIVE}" >> "$LOG_FILE"

# Silent success - don't output anything (hook succeeded)
exit 0
