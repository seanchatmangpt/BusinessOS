# =============================================================================
# BusinessOS — Makefile
# =============================================================================
# Simple shortcuts for common tasks.
# Run `make help` to see all available targets.
#
# Usage:
#   make setup      Set up everything from scratch
#   make start      Start all services
#   make stop       Stop all services
#   make dev        Start in development mode (same as start)
#   make build      Build production binaries
#   make desktop    Build the Electron desktop app
#   make test       Run all tests
#   make clean      Remove build artifacts
# =============================================================================

# Project root — always the directory containing this Makefile
PROJECT_ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Go binary output directory
BIN_DIR      := $(PROJECT_ROOT)bin

# Backend source
BACKEND_DIR  := $(PROJECT_ROOT)backend

# Frontend source
FRONTEND_DIR := $(PROJECT_ROOT)frontend

# Desktop app source
DESKTOP_DIR  := $(PROJECT_ROOT)desktop-app

# Build version (from git tag if available)
VERSION      := $(shell git describe --tags --always 2>/dev/null || echo "dev")

# Go ldflags for smaller, version-stamped binaries
GO_LDFLAGS   := -ldflags="-s -w -X main.Version=$(VERSION)"

# Detect OS for open-browser command
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  OPEN_CMD := open
else
  OPEN_CMD := xdg-open
endif

# Make all targets phony (not actual files)
.PHONY: help setup start stop dev build build-backend build-frontend \
        desktop test test-backend test-frontend test-e2e clean \
        logs status check-env

# ---------------------------------------------------------------------------
# Default target: show help
# ---------------------------------------------------------------------------
help:
	@echo ""
	@echo "  BusinessOS — Available Commands"
	@echo "  ─────────────────────────────────────────────"
	@echo ""
	@echo "  GETTING STARTED:"
	@echo "    make setup       Set up everything (run this first!)"
	@echo "    make start       Start all services"
	@echo "    make stop        Stop all services"
	@echo ""
	@echo "  DEVELOPMENT:"
	@echo "    make dev         Start in development mode (hot reload)"
	@echo "    make build       Build production binaries"
	@echo "    make build-backend   Build only the Go backend"
	@echo "    make build-frontend  Build only the SvelteKit frontend"
	@echo "    make desktop     Build the Electron desktop app"
	@echo ""
	@echo "  TESTING:"
	@echo "    make test        Run all tests"
	@echo "    make test-backend   Run Go backend tests"
	@echo "    make test-frontend  Run SvelteKit frontend tests"
	@echo "    make test-e2e    Run end-to-end Playwright tests"
	@echo ""
	@echo "  UTILITIES:"
	@echo "    make status      Show status of running services"
	@echo "    make logs        Tail all service logs"
	@echo "    make clean       Remove build artifacts"
	@echo "    make check-env   Verify .env is configured correctly"
	@echo ""
	@echo "  AUTOSTART:"
	@echo "    make autostart         Install desktop auto-start on login"
	@echo "    make autostart-remove  Remove desktop auto-start"
	@echo ""

# ---------------------------------------------------------------------------
# Setup — run this once after cloning
# ---------------------------------------------------------------------------
setup:
	@echo ""
	@echo "  Running BusinessOS setup..."
	@echo ""
	@chmod +x "$(PROJECT_ROOT)setup.sh"
	@"$(PROJECT_ROOT)setup.sh"

setup-quick:
	@echo ""
	@echo "  Running BusinessOS quick setup (SQLite, no optional services)..."
	@echo ""
	@chmod +x "$(PROJECT_ROOT)setup.sh"
	@"$(PROJECT_ROOT)setup.sh" --quick

# ---------------------------------------------------------------------------
# Start / Stop — run all services
# ---------------------------------------------------------------------------
start:
	@chmod +x "$(PROJECT_ROOT)start.sh"
	@"$(PROJECT_ROOT)start.sh"

stop:
	@chmod +x "$(PROJECT_ROOT)stop.sh"
	@"$(PROJECT_ROOT)stop.sh"

# Development mode is the same as start (hot reload is always on in dev)
dev: start

# ---------------------------------------------------------------------------
# Build — compile production artifacts
# ---------------------------------------------------------------------------
build: build-backend build-frontend
	@echo ""
	@echo "  Build complete."
	@echo "  Backend binary:    $(BIN_DIR)/businessos-server"
	@echo "  Frontend dist:     $(FRONTEND_DIR)/build/"
	@echo ""

build-backend:
	@echo ""
	@echo "  Building Go backend (version: $(VERSION))..."
	@mkdir -p "$(BIN_DIR)"
	@cd "$(BACKEND_DIR)" && \
		go build $(GO_LDFLAGS) \
		-o "$(BIN_DIR)/businessos-server" \
		./cmd/server/
	@echo "  Backend binary ready: $(BIN_DIR)/businessos-server"
	@echo ""

build-frontend:
	@echo ""
	@echo "  Building SvelteKit frontend..."
	@cd "$(FRONTEND_DIR)" && npm run build
	@echo "  Frontend build ready: $(FRONTEND_DIR)/build/"
	@echo ""

# ---------------------------------------------------------------------------
# Desktop App — Electron build
# ---------------------------------------------------------------------------
desktop:
	@echo ""
	@echo "  Building Electron desktop app..."
	@cd "$(DESKTOP_DIR)" && npm run make
	@echo ""
	@echo "  Desktop app built in: $(DESKTOP_DIR)/out/"
	@echo ""

desktop-start:
	@echo ""
	@echo "  Starting Electron app (dev mode)..."
	@cd "$(DESKTOP_DIR)" && npm run start

# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
test: test-backend test-frontend
	@echo ""
	@echo "  All tests complete."
	@echo ""

test-backend:
	@echo ""
	@echo "  Running Go backend tests..."
	@cd "$(BACKEND_DIR)" && go test -v -race ./...
	@echo ""

test-frontend:
	@echo ""
	@echo "  Running SvelteKit frontend tests (vitest)..."
	@cd "$(FRONTEND_DIR)" && npm run test
	@echo ""

test-e2e:
	@echo ""
	@echo "  Running end-to-end tests (Playwright)..."
	@echo "  Make sure the app is running first: make start"
	@echo ""
	@cd "$(FRONTEND_DIR)" && npm run test:e2e
	@echo ""

test-coverage:
	@echo ""
	@echo "  Running tests with coverage report..."
	@cd "$(FRONTEND_DIR)" && npm run test:coverage
	@echo ""

# ---------------------------------------------------------------------------
# Status — check what's running
# ---------------------------------------------------------------------------
status:
	@echo ""
	@echo "  BusinessOS Service Status"
	@echo "  ─────────────────────────────────────"
	@echo ""
	@if nc -z localhost 8001 2>/dev/null; then \
		echo "  Backend  (port 8001):   RUNNING"; \
	else \
		echo "  Backend  (port 8001):   STOPPED"; \
	fi
	@if nc -z localhost 5173 2>/dev/null; then \
		echo "  Frontend (port 5173):   RUNNING"; \
	else \
		echo "  Frontend (port 5173):   STOPPED"; \
	fi
	@if nc -z localhost 8089 2>/dev/null; then \
		echo "  OSA Agent (port 8089):  RUNNING"; \
	else \
		echo "  OSA Agent (port 8089):  STOPPED (optional)"; \
	fi
	@if nc -z localhost 5432 2>/dev/null; then \
		echo "  PostgreSQL (port 5432): RUNNING"; \
	else \
		echo "  PostgreSQL (port 5432): STOPPED"; \
	fi
	@if nc -z localhost 6379 2>/dev/null; then \
		echo "  Redis (port 6379):      RUNNING"; \
	else \
		echo "  Redis (port 6379):      STOPPED (optional, in-memory fallback active)"; \
	fi
	@echo ""

# ---------------------------------------------------------------------------
# Logs — tail all service logs
# ---------------------------------------------------------------------------
logs:
	@echo ""
	@echo "  Tailing all service logs (Ctrl+C to stop)..."
	@echo ""
	@mkdir -p "$(PROJECT_ROOT)logs"
	@tail -f "$(PROJECT_ROOT)logs/backend.log" \
	          "$(PROJECT_ROOT)logs/frontend.log" \
	          2>/dev/null || echo "  No log files found. Start the app first with: make start"

logs-backend:
	@tail -f "$(PROJECT_ROOT)logs/backend.log" 2>/dev/null || echo "Backend log not found."

logs-frontend:
	@tail -f "$(PROJECT_ROOT)logs/frontend.log" 2>/dev/null || echo "Frontend log not found."

# ---------------------------------------------------------------------------
# Environment check
# ---------------------------------------------------------------------------
check-env:
	@echo ""
	@echo "  Checking environment configuration..."
	@echo ""
	@if [ ! -f "$(BACKEND_DIR)/.env" ]; then \
		echo "  ERROR: backend/.env not found. Run: make setup"; \
		exit 1; \
	fi
	@echo "  backend/.env found"
	@if grep -q "CHANGE_ME" "$(BACKEND_DIR)/.env" 2>/dev/null; then \
		echo "  WARN: backend/.env contains 'CHANGE_ME' placeholder values."; \
		echo "        Edit backend/.env and fill in the required values."; \
	else \
		echo "  .env looks configured"; \
	fi
	@if [ ! -f "$(BIN_DIR)/businessos-server" ]; then \
		echo "  WARN: Backend binary not found. Run: make build-backend"; \
	else \
		echo "  Backend binary found: $(BIN_DIR)/businessos-server"; \
	fi
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "  WARN: frontend/node_modules not found. Run: make setup"; \
	else \
		echo "  Frontend node_modules found"; \
	fi
	@echo ""
	@echo "  To start the app, run: make start"
	@echo ""

# ---------------------------------------------------------------------------
# Auto-start (desktop app)
# ---------------------------------------------------------------------------
autostart:
	@chmod +x "$(DESKTOP_DIR)/scripts/install-autostart.sh"
	@"$(DESKTOP_DIR)/scripts/install-autostart.sh"

autostart-remove:
	@chmod +x "$(DESKTOP_DIR)/scripts/install-autostart.sh"
	@"$(DESKTOP_DIR)/scripts/install-autostart.sh" --uninstall

autostart-status:
	@chmod +x "$(DESKTOP_DIR)/scripts/install-autostart.sh"
	@"$(DESKTOP_DIR)/scripts/install-autostart.sh" --status

# ---------------------------------------------------------------------------
# Clean — remove build artifacts
# ---------------------------------------------------------------------------
clean:
	@echo ""
	@echo "  Cleaning build artifacts..."
	@rm -rf "$(BIN_DIR)/businessos-server"
	@rm -rf "$(FRONTEND_DIR)/build" "$(FRONTEND_DIR)/.svelte-kit"
	@rm -rf "$(DESKTOP_DIR)/out" "$(DESKTOP_DIR)/.vite"
	@rm -rf "$(PROJECT_ROOT)logs"
	@rm -rf "$(PROJECT_ROOT).pids"
	@echo "  Clean complete."
	@echo ""

clean-all: clean
	@echo "  Removing node_modules (this will require npm install again)..."
	@rm -rf "$(FRONTEND_DIR)/node_modules"
	@rm -rf "$(DESKTOP_DIR)/node_modules"
	@echo "  Done. Run 'make setup' to reinstall."
	@echo ""
