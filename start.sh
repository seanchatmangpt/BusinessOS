#!/usr/bin/env bash
# =============================================================================
# BusinessOS Start Script
# =============================================================================
# Starts all services in one command.
#
# Usage:
#   ./start.sh           Start all services (backend + frontend)
#   ./start.sh --no-browser  Don't open the browser automatically
#
# Stop everything with:  ./stop.sh
# Or press Ctrl+C here and all services will shut down cleanly.
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Terminal colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
print_header() {
    echo ""
    echo -e "${BOLD}${BLUE}  $1${RESET}"
    echo -e "${BLUE}  $(printf '─%.0s' $(seq 1 ${#1}))${RESET}"
}

print_ok()   { echo -e "  ${GREEN}[UP]${RESET}   $1"; }
print_warn() { echo -e "  ${YELLOW}[WARN]${RESET} $1"; }
print_info() { echo -e "  ${CYAN}[INFO]${RESET} $1"; }
print_error(){ echo -e "  ${RED}[DOWN]${RESET} $1"; }

command_exists() { command -v "$1" &>/dev/null; }

# ---------------------------------------------------------------------------
# Script root
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---------------------------------------------------------------------------
# Parse flags
# ---------------------------------------------------------------------------
OPEN_BROWSER=true
for arg in "$@"; do
    case "$arg" in
        --no-browser) OPEN_BROWSER=false ;;
        --help|-h)
            echo "Usage: $0 [--no-browser]"
            echo "  --no-browser   Do not open the browser automatically"
            exit 0
            ;;
    esac
done

# ---------------------------------------------------------------------------
# PID file directory — used by stop.sh to find the processes
# ---------------------------------------------------------------------------
PID_DIR="$SCRIPT_DIR/.pids"
mkdir -p "$PID_DIR"

# ---------------------------------------------------------------------------
# Port check helper
# ---------------------------------------------------------------------------
wait_for_port() {
    local port=$1
    local name=$2
    local timeout=30
    local elapsed=0
    while ! nc -z localhost "$port" 2>/dev/null; do
        if [ "$elapsed" -ge "$timeout" ]; then
            return 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    return 0
}

port_in_use() {
    nc -z localhost "$1" 2>/dev/null
}

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
clear
echo ""
echo -e "${BOLD}${BLUE}"
echo "  ╔══════════════════════════════════════════════════╗"
echo "  ║           BusinessOS  —  Starting Up             ║"
echo "  ╚══════════════════════════════════════════════════╝"
echo -e "${RESET}"

# Make sure setup has been run
if [ ! -f "$SCRIPT_DIR/bin/businessos-server" ]; then
    echo -e "${RED}${BOLD}Error: Backend binary not found at bin/businessos-server${RESET}"
    echo ""
    echo -e "  Run setup first:"
    echo -e "  ${BOLD}./setup.sh${RESET}"
    echo ""
    exit 1
fi

if [ ! -d "$SCRIPT_DIR/frontend/node_modules" ]; then
    echo -e "${RED}${BOLD}Error: Frontend node_modules not found.${RESET}"
    echo ""
    echo -e "  Run setup first:"
    echo -e "  ${BOLD}./setup.sh${RESET}"
    echo ""
    exit 1
fi

if [ ! -f "$SCRIPT_DIR/backend/.env" ]; then
    echo -e "${RED}${BOLD}Error: backend/.env not found.${RESET}"
    echo ""
    echo -e "  Run setup first:"
    echo -e "  ${BOLD}./setup.sh${RESET}"
    echo ""
    exit 1
fi

# Check for port conflicts
if port_in_use 8001; then
    print_warn "Port 8001 is already in use. Is the backend already running?"
    print_warn "Run ./stop.sh first, or use a different port in backend/.env"
fi

if port_in_use 5173; then
    print_warn "Port 5173 is already in use. Is the frontend already running?"
fi

# ---------------------------------------------------------------------------
# Cleanup function — runs when this script exits (Ctrl+C, kill, error)
# ---------------------------------------------------------------------------
PIDS=()

cleanup() {
    echo ""
    echo -e "${YELLOW}  Shutting down all services...${RESET}"

    # Kill all tracked PIDs
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done

    # Give processes a moment to exit gracefully
    sleep 1

    # Force kill anything still running
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -KILL "$pid" 2>/dev/null || true
        fi
    done

    # Clean up PID files
    rm -f "$PID_DIR"/*.pid

    echo -e "${GREEN}  All services stopped. Goodbye!${RESET}"
    echo ""
    exit 0
}

trap cleanup INT TERM EXIT

# ---------------------------------------------------------------------------
# Log directory
# ---------------------------------------------------------------------------
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"

# ---------------------------------------------------------------------------
# Start: Go Backend
# ---------------------------------------------------------------------------
print_header "Starting Backend (Go)"

BACKEND_LOG="$LOG_DIR/backend.log"

(
    cd "$SCRIPT_DIR/backend"
    "$SCRIPT_DIR/bin/businessos-server" >> "$BACKEND_LOG" 2>&1
) &
BACKEND_PID=$!
PIDS+=("$BACKEND_PID")
echo "$BACKEND_PID" > "$PID_DIR/backend.pid"

# Wait up to 15 seconds for the backend to come up
if wait_for_port 8001 "backend"; then
    print_ok "Backend running on http://localhost:8001  (log: logs/backend.log)"
else
    print_error "Backend failed to start within 15 seconds."
    echo ""
    echo -e "  Check the log for details:"
    echo -e "  ${BOLD}cat logs/backend.log${RESET}"
    echo ""
    # Print last 20 lines of log
    if [ -f "$BACKEND_LOG" ]; then
        echo -e "${YELLOW}  --- Last 20 lines of backend.log ---${RESET}"
        tail -20 "$BACKEND_LOG" | sed 's/^/  /'
        echo -e "${YELLOW}  ------------------------------------${RESET}"
    fi
    echo ""
    print_warn "Continuing with frontend only. Fix the backend issue and restart."
fi

# ---------------------------------------------------------------------------
# Start: SvelteKit Frontend
# ---------------------------------------------------------------------------
print_header "Starting Frontend (SvelteKit)"

FRONTEND_LOG="$LOG_DIR/frontend.log"

(
    cd "$SCRIPT_DIR/frontend"
    npm run dev >> "$FRONTEND_LOG" 2>&1
) &
FRONTEND_PID=$!
PIDS+=("$FRONTEND_PID")
echo "$FRONTEND_PID" > "$PID_DIR/frontend.pid"

# Wait up to 20 seconds for the dev server
if wait_for_port 5173 "frontend"; then
    print_ok "Frontend running on http://localhost:5173  (log: logs/frontend.log)"
else
    print_error "Frontend dev server failed to start within 20 seconds."
    echo ""
    echo -e "  Check the log for details:"
    echo -e "  ${BOLD}cat logs/frontend.log${RESET}"
    echo ""
fi

# ---------------------------------------------------------------------------
# Start: OSA Agent (optional, only if configured)
# ---------------------------------------------------------------------------
OSA_ENABLED=$(grep -E '^OSA_ENABLED=' "$SCRIPT_DIR/backend/.env" 2>/dev/null | cut -d= -f2 | tr -d '"' | tr -d "'" || echo "false")

if [[ "${OSA_ENABLED,,}" == "true" ]]; then
    print_header "Starting OSA Agent (Elixir)"

    OSA_DIR="$SCRIPT_DIR/osa-agent"
    OSA_LOG="$LOG_DIR/osa.log"

    if [ -d "$OSA_DIR" ] && [ -f "$OSA_DIR/mix.exs" ]; then
        if command_exists mix; then
            (
                cd "$OSA_DIR"
                mix phx.server >> "$OSA_LOG" 2>&1
            ) &
            OSA_PID=$!
            PIDS+=("$OSA_PID")
            echo "$OSA_PID" > "$PID_DIR/osa.pid"

            if wait_for_port 8089 "osa"; then
                print_ok "OSA Agent running on http://localhost:8089  (log: logs/osa.log)"
            else
                print_warn "OSA Agent did not start. Check logs/osa.log"
            fi
        else
            print_warn "OSA_ENABLED=true but Elixir/mix not found. Skipping OSA Agent."
            print_warn "Install Elixir or set OSA_ENABLED=false in backend/.env"
        fi
    else
        print_warn "OSA_ENABLED=true but osa-agent/ directory not found. Skipping."
    fi
fi

# ---------------------------------------------------------------------------
# Open browser
# ---------------------------------------------------------------------------
if [ "$OPEN_BROWSER" = true ] && port_in_use 5173; then
    sleep 1
    case "$(uname -s)" in
        Darwin) open "http://localhost:5173" 2>/dev/null || true ;;
        Linux)
            if command_exists xdg-open; then
                xdg-open "http://localhost:5173" 2>/dev/null || true
            fi
            ;;
    esac
fi

# ---------------------------------------------------------------------------
# Status summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${BOLD}${GREEN}"
echo "  ╔══════════════════════════════════════════════════╗"
echo "  ║         All Services Running                     ║"
echo "  ╚══════════════════════════════════════════════════╝"
echo -e "${RESET}"
echo -e "  ${CYAN}Open in browser:${RESET}  ${BOLD}http://localhost:5173${RESET}"
echo -e "  ${CYAN}API endpoint:${RESET}    ${BOLD}http://localhost:8001${RESET}"
echo -e "  ${CYAN}API health:${RESET}      ${BOLD}http://localhost:8001/health${RESET}"
echo ""
echo -e "  ${YELLOW}Logs:${RESET}"
echo -e "    Backend:  logs/backend.log"
echo -e "    Frontend: logs/frontend.log"
echo ""
echo -e "  ${YELLOW}Press Ctrl+C to stop all services.${RESET}"
echo ""

# ---------------------------------------------------------------------------
# Keep this script alive — monitors child processes
# ---------------------------------------------------------------------------
# Poll every 5 seconds and report if any service crashes unexpectedly
while true; do
    sleep 5

    CRASHED=false
    for pid in "${PIDS[@]}"; do
        if ! kill -0 "$pid" 2>/dev/null; then
            CRASHED=true
            break
        fi
    done

    if [ "$CRASHED" = true ]; then
        echo ""
        echo -e "${RED}${BOLD}  A service has crashed unexpectedly.${RESET}"
        echo -e "  Check the log files in ${BOLD}logs/${RESET} for details."
        echo -e "  Stopping all services..."
        cleanup
        break
    fi
done
