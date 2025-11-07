#!/bin/bash
# Restart EngineerDNA dev server with clean state
# Usage:
#   ./scripts/restart-dev-server.sh              # Just restart server
#   ./scripts/restart-dev-server.sh --rebuild    # Rebuild and restart
#   ./scripts/restart-dev-server.sh --frontend   # Include frontend dev server

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

REBUILD=false
FRONTEND=false

# Parse arguments
for arg in "$@"; do
  case $arg in
    --rebuild)
      REBUILD=true
      shift
      ;;
    --frontend)
      FRONTEND=true
      shift
      ;;
    *)
      echo "Unknown option: $arg"
      echo "Usage: $0 [--rebuild] [--frontend]"
      exit 1
      ;;
  esac
done

echo "Restarting EngineerDNA dev server..."

# Function to kill process on port
kill_port() {
  local port=$1
  local name=$2

  local pids=$(lsof -ti:$port -sTCP:LISTEN 2>/dev/null || true)

  if [ ! -z "$pids" ]; then
    for pid in $pids; do
      echo "  Stopping $name on port $port (PID: $pid)"
      kill $pid 2>/dev/null || true
      sleep 0.5
      # Force kill if still running
      if kill -0 $pid 2>/dev/null; then
        echo "    Force killing stubborn process"
        kill -9 $pid 2>/dev/null || true
      fi
    done
  fi
}

# Stop backend server
echo "Stopping backend server..."
kill_port 3847 "EngineerDNA backend"

# Stop frontend dev server if requested
if [ "$FRONTEND" = true ]; then
  echo "Stopping frontend dev server..."
  kill_port 5173 "Vite dev server"
fi

# Kill any remaining engineerdna processes
echo "Cleaning up any remaining processes..."
for pid in $(ps aux | grep "[b]in/engineerdna" | awk '{print $2}'); do
  echo "  Cleaning up engineerdna process (PID: $pid)"
  kill $pid 2>/dev/null || true
  sleep 0.2
  kill -9 $pid 2>/dev/null || true
done

# Wait for processes to terminate
sleep 1

# Rebuild if requested
if [ "$REBUILD" = true ]; then
  echo "Rebuilding EngineerDNA..."
  echo "  Building frontend..."
  cd "$PROJECT_ROOT/frontend" && npm run build
  echo "  Building backend..."
  cd "$PROJECT_ROOT" && make build
  echo "  Build complete"
fi

# Verify binary exists
if [ ! -f "$PROJECT_ROOT/bin/engineerdna" ]; then
  echo "ERROR: Binary not found at bin/engineerdna"
  echo "Run with --rebuild flag to build the binary"
  exit 1
fi

# Start backend server
echo "Starting EngineerDNA backend server..."
( cd "$PROJECT_ROOT" && ./bin/engineerdna serve > /tmp/engineerdna-backend.log 2>&1 </dev/null & )
echo "  Backend server starting on port 3847..."

# Start frontend dev server if requested
if [ "$FRONTEND" = true ]; then
  echo "Starting frontend dev server..."
  ( cd "$PROJECT_ROOT/frontend" && npm run dev > /tmp/engineerdna-frontend.log 2>&1 </dev/null & )
  echo "  Frontend dev server starting on port 5173..."
fi

# Wait for server to initialize
echo "Waiting for server to initialize..."
sleep 3

# Verify backend is running
echo "Verification:"
success=true

if lsof -ti:3847 > /dev/null 2>&1; then
  echo "  [OK] Backend server is active on port 3847"
else
  echo "  [FAILED] Backend server failed to start"
  echo "  Check logs: tail -f /tmp/engineerdna-backend.log"
  success=false
fi

if [ "$FRONTEND" = true ]; then
  if lsof -ti:5173 > /dev/null 2>&1; then
    echo "  [OK] Frontend dev server is active on port 5173"
  else
    echo "  [FAILED] Frontend dev server failed to start"
    echo "  Check logs: tail -f /tmp/engineerdna-frontend.log"
    success=false
  fi
fi

echo ""

if [ "$success" = true ]; then
  echo "SUCCESS: EngineerDNA restarted successfully!"
  echo ""
  if [ "$FRONTEND" = true ]; then
    echo "Access the application at:"
    echo "  - Frontend (dev): http://localhost:5173"
    echo "  - Backend API:    http://localhost:3847"
  else
    echo "Access the application at:"
    echo "  - Application: http://localhost:3847"
  fi
  echo ""
  echo "Logs:"
  echo "  - Backend: tail -f /tmp/engineerdna-backend.log"
  if [ "$FRONTEND" = true ]; then
    echo "  - Frontend: tail -f /tmp/engineerdna-frontend.log"
  fi
else
  echo "FAILED: Server failed to start. Check logs above."
  exit 1
fi

exit 0
