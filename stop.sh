#!/usr/bin/env bash
# =============================================================================
# BusinessOS Stop Script
# =============================================================================
# Stops all running BusinessOS services cleanly.
#
# Usage:
#   ./stop.sh
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Terminal colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
RESET='\033[0m'

print_ok()   { echo -e "  ${GREEN}[STOPPED]${RESET} $1"; }
print_warn() { echo -e "  ${YELLOW}[SKIP]${RESET}    $1"; }
print_info() { echo -e "  ${BOLD}-->${RESET}       $1"; }

# ---------------------------------------------------------------------------
# Script root
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$SCRIPT_DIR/.pids"

echo ""
echo -e "${BOLD}  Stopping BusinessOS services...${RESET}"
echo ""

STOPPED_ANY=false

# ---------------------------------------------------------------------------
# Stop a service by PID file
# ---------------------------------------------------------------------------
stop_service() {
    local name=$1
    local pid_file="$PID_DIR/${name}.pid"

    if [ ! -f "$pid_file" ]; then
        print_warn "$name — no PID file found (may not be running)"
        return
    fi

    local pid
    pid=$(cat "$pid_file")

    if [ -z "$pid" ]; then
        print_warn "$name — PID file is empty"
        rm -f "$pid_file"
        return
    fi

    if kill -0 "$pid" 2>/dev/null; then
        print_info "Sending SIGTERM to $name (PID $pid)..."
        kill -TERM "$pid" 2>/dev/null || true

        # Wait up to 10 seconds for graceful shutdown
        local waited=0
        while kill -0 "$pid" 2>/dev/null && [ "$waited" -lt 10 ]; do
            sleep 1
            waited=$((waited + 1))
        done

        # Force kill if still running
        if kill -0 "$pid" 2>/dev/null; then
            print_warn "$name did not stop gracefully — sending SIGKILL..."
            kill -KILL "$pid" 2>/dev/null || true
        fi

        print_ok "$name (PID $pid)"
        STOPPED_ANY=true
    else
        print_warn "$name — PID $pid is not running (already stopped?)"
    fi

    rm -f "$pid_file"
}

# ---------------------------------------------------------------------------
# Stop all services (in reverse start order)
# ---------------------------------------------------------------------------
stop_service "osa"
stop_service "frontend"
stop_service "backend"

# ---------------------------------------------------------------------------
# Fallback: find any lingering processes by port / name
# ---------------------------------------------------------------------------
echo ""
print_info "Checking for any remaining processes on known ports..."

# Backend on 8001
BACKEND_PID=$(lsof -ti tcp:8001 2>/dev/null || true)
if [ -n "$BACKEND_PID" ]; then
    print_info "Found process on port 8001 (PID $BACKEND_PID) — stopping..."
    kill -TERM $BACKEND_PID 2>/dev/null || true
    sleep 1
    kill -0 $BACKEND_PID 2>/dev/null && kill -KILL $BACKEND_PID 2>/dev/null || true
    print_ok "Port 8001 cleared"
    STOPPED_ANY=true
fi

# Frontend dev server on 5173
FRONTEND_PID=$(lsof -ti tcp:5173 2>/dev/null || true)
if [ -n "$FRONTEND_PID" ]; then
    print_info "Found process on port 5173 (PID $FRONTEND_PID) — stopping..."
    kill -TERM $FRONTEND_PID 2>/dev/null || true
    sleep 1
    kill -0 $FRONTEND_PID 2>/dev/null && kill -KILL $FRONTEND_PID 2>/dev/null || true
    print_ok "Port 5173 cleared"
    STOPPED_ANY=true
fi

# OSA Agent on 8089
OSA_PID=$(lsof -ti tcp:8089 2>/dev/null || true)
if [ -n "$OSA_PID" ]; then
    print_info "Found process on port 8089 (PID $OSA_PID) — stopping..."
    kill -TERM $OSA_PID 2>/dev/null || true
    sleep 1
    kill -0 $OSA_PID 2>/dev/null && kill -KILL $OSA_PID 2>/dev/null || true
    print_ok "Port 8089 cleared"
    STOPPED_ANY=true
fi

# Clean up PID directory
if [ -d "$PID_DIR" ]; then
    rm -rf "$PID_DIR"
fi

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------
echo ""
if [ "$STOPPED_ANY" = true ]; then
    echo -e "  ${GREEN}${BOLD}All BusinessOS services stopped.${RESET}"
else
    echo -e "  ${YELLOW}No running BusinessOS services found.${RESET}"
fi
echo ""
echo -e "  Start again with: ${BOLD}./start.sh${RESET}"
echo ""
