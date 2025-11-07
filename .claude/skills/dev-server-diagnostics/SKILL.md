---
name: dev-server-diagnostics
description: Diagnose localhost:3847 issues by checking process status, reading logs, and identifying root causes. Use when debugging connection errors, blank pages, or server not responding.
tools: Bash, Read, Grep, Glob, BashOutput
disallowedTools: Write, Edit
---

# Dev Server Diagnostics

## Purpose

Diagnose EngineerDNA server issues without immediately restarting. Identify root causes through process inspection, log analysis, and port conflict detection.

**Key Rule: Diagnose FIRST, restart if needed SECOND.**

## Quick Start

### Diagnostic Workflow

1. **Check if server is running**
   ```bash
   ps aux | grep engineerdna
   lsof -i :3847
   ```

2. **Test connection**
   ```bash
   curl -I http://127.0.0.1:3847/health
   ```

3. **Check background process** (if started with Bash tool)
   ```bash
   # Use bash_id from system reminders
   BashOutput(bash_id: "xxx")
   ```

4. **Read server logs** (if logging to file)
   ```bash
   tail -50 /tmp/engineerdna.log
   # Or check stderr/stdout from background process
   ```

5. **Diagnose and fix**

## Common Issues

### Issue 1: Connection Refused

**Symptoms:**
- `curl: (7) Failed to connect to 127.0.0.1 port 3847: Connection refused`
- Browser shows "This site can't be reached"

**Diagnosis:**
```bash
# Is server running?
ps aux | grep engineerdna
# Should show: ./bin/engineerdna or go run .

# If not running, check why it stopped
# Check last build status
go build -o bin/engineerdna .
echo $?  # Should be 0
```

**Root Causes:**
- Server not started
- Compilation errors
- Crash on startup
- Wrong port binding

**Fix:**
```bash
# Use restart script (RECOMMENDED)
./scripts/restart-dev-server.sh --rebuild

# Or manually (NOT RECOMMENDED - use restart script instead)
# ./bin/engineerdna > /tmp/engineerdna.log 2>&1 &
```

### Issue 2: Server Running But Not Responding

**Symptoms:**
- Process exists in `ps aux`
- Port 3847 is bound
- But requests hang or timeout

**Diagnosis:**
```bash
# Check if port is listening
lsof -i :3847
# Should show: engineerdna

# Test different endpoints
curl -v http://127.0.0.1:3847/health
curl -v http://127.0.0.1:3847/api/events

# Check for deadlock or panic in logs
tail -100 /tmp/engineerdna.log
```

**Root Causes:**
- Database locked
- Deadlock in code
- Panic in handler
- Slow database query

**Fix:**
```bash
# Check database lock
lsof ~/.engineerdna/engineerdna.db

# If locked by another process, kill it
pkill -f "sqlite3.*engineerdna.db"

# Restart server (RECOMMENDED)
./scripts/restart-dev-server.sh
```

### Issue 3: Port Already in Use

**Symptoms:**
- Server fails to start with "address already in use"
- Error: `bind: address already in use`

**Diagnosis:**
```bash
# What's using port 3847?
lsof -i :3847

# How many engineerdna processes?
pgrep -f engineerdna | wc -l
```

**Root Causes:**
- Previous instance didn't shut down
- Another process using port 3847
- Multiple instances started accidentally

**Fix:**
```bash
# Use restart script (handles cleanup automatically)
./scripts/restart-dev-server.sh

# Manual cleanup (NOT RECOMMENDED - use restart script instead)
# pkill engineerdna
# lsof -i :3847  # Verify port is free
# ./bin/engineerdna > /tmp/engineerdna.log 2>&1 &
```

### Issue 4: Build Errors

**Symptoms:**
- `./bin/engineerdna` doesn't exist
- Binary is outdated
- Compilation fails

**Diagnosis:**
```bash
# Check binary exists and is recent
ls -lah ./bin/engineerdna
stat -f "%Sm" bin/engineerdna

# Check last build time vs code changes
git diff --name-only HEAD~1

# Try building
go build -o bin/engineerdna .
```

**Root Causes:**
- Forgot to rebuild after code changes
- Compilation errors
- Missing dependencies

**Fix:**
```bash
# Use restart script with rebuild (RECOMMENDED)
./scripts/restart-dev-server.sh --rebuild
```

### Issue 5: Database Issues

**Symptoms:**
- Server starts but crashes immediately
- Errors like "unable to open database file"
- "database is locked"

**Diagnosis:**
```bash
# Check database exists
ls -lah ~/.engineerdna/engineerdna.db

# Check permissions
ls -la ~/.engineerdna/

# Check if locked
lsof ~/.engineerdna/engineerdna.db

# Test database directly
sqlite3 ~/.engineerdna/engineerdna.db "SELECT COUNT(*) FROM events"
```

**Root Causes:**
- Database file doesn't exist
- Permission issues
- Database locked by another process
- Corrupted database

**Fix:**
```bash
# Create directory if needed
mkdir -p ~/.engineerdna

# Unlock database
pkill -f "sqlite3.*engineerdna.db"
pkill engineerdna

# Check integrity
sqlite3 ~/.engineerdna/engineerdna.db "PRAGMA integrity_check"

# Restart server
./scripts/restart-dev-server.sh
```

### Issue 6: Plugin Errors

**Symptoms:**
- Server starts but errors when loading plugins
- Errors like "plugin not found" or "plugin.json invalid"

**Diagnosis:**
```bash
# Check plugin directories
ls -la ~/.engineerdna/plugins/
ls -la ./plugins/

# Check plugin.json validity
jq . ./plugins/github/plugin.json

# Check plugin executable
file ./plugins/github/github
./plugins/github/github --help
```

**Root Causes:**
- Plugin not found
- Invalid plugin.json
- Plugin not executable
- Plugin crashes on startup

**Fix:**
```bash
# Make plugin executable
chmod +x ./plugins/*/$(basename ./plugins/*)

# Validate plugin.json
jq . ./plugins/*/plugin.json

# Test plugin directly
echo '{"jsonrpc":"2.0","method":"plugin.info","id":1}' | ./plugins/github/github
```

## Diagnostic Checklist

When debugging localhost issues:

- [ ] Checked if server process is running (`ps aux`)
- [ ] Checked if port 3847 is bound (`lsof -i :3847`)
- [ ] Tested connection with curl
- [ ] Checked server logs or background process output
- [ ] Identified specific error (not just "doesn't work")
- [ ] Determined root cause
- [ ] Applied targeted fix
- [ ] Verified fix works

## Tools Reference

### Check Process Status
```bash
ps aux | grep engineerdna          # Is it running?
pgrep -f engineerdna               # Get PID
lsof -i :3847                      # What's on port 3847?
```

### Test Connection
```bash
curl -I http://127.0.0.1:3847/health          # Health check
curl -v http://127.0.0.1:3847/api/events      # API endpoint
nc -zv 127.0.0.1 3847                         # Port connectivity
```

### Check Logs
```bash
tail -50 /tmp/engineerdna.log      # Recent logs
BashOutput(bash_id: "xxx")         # Background process output
```

### Check Database
```bash
ls -lah ~/.engineerdna/engineerdna.db         # Exists?
lsof ~/.engineerdna/engineerdna.db            # Locked?
sqlite3 ~/.engineerdna/engineerdna.db ".schema"  # Schema
```

### Kill and Restart
```bash
# RECOMMENDED: Use restart script
./scripts/restart-dev-server.sh

# Manual (NOT RECOMMENDED)
# pkill engineerdna && ./bin/engineerdna > /tmp/engineerdna.log 2>&1 &
```

## Frontend Diagnostics (when added)

When React frontend is added:

### Check Frontend Build
```bash
ls -la frontend/dist/              # Build artifacts exist?
ls -la frontend/dist/index.html    # Entry point exists?
```

### Check Frontend Serving
```bash
curl http://127.0.0.1:3847/        # Returns HTML?
curl http://127.0.0.1:3847/assets/ # Assets load?
```

### Check Browser Console
- Open http://127.0.0.1:3847
- Open DevTools (F12)
- Check Console tab for JavaScript errors
- Check Network tab for failed requests

## Best Practices

1. **Diagnose Before Restarting** - Understand the problem first
2. **Check Logs** - They usually tell you what's wrong
3. **Test Incrementally** - Process → Port → Connection → Response
4. **Document Findings** - What was wrong and how you fixed it
5. **Prevent Recurrence** - Fix root cause, not symptoms

## When to Restart

Restart is needed when:
- Configuration changes (port, database path)
- Code changes compiled to new binary
- Database issues resolved
- Resource leak (memory/connections)

Restart is NOT needed for:
- Frontend changes (hot reload when added)
- Database content changes (inserts/updates)
- Plugin configuration changes (should hot reload)

## Integration with Development Workflow

**ALWAYS use the restart script for development server management.**

After code changes:
```bash
# RECOMMENDED: Rebuild and restart with one command
./scripts/restart-dev-server.sh --rebuild

# Logs at: /tmp/engineerdna-backend.log
```

After database migrations:
```bash
# Migrations applied automatically on startup
./scripts/restart-dev-server.sh

# With frontend dev server (hot reload)
./scripts/restart-dev-server.sh --frontend
```

**Why use the restart script?**
- Automatically kills zombie processes (no "address already in use")
- Clean port cleanup (handles port 3847 conflicts)
- Consistent logging location
- Verifies server started successfully
- Optional rebuild flag
- Optional frontend dev server
