#!/bin/bash

# BusinessOS Complete Startup Script
# Starts all required services for development environment
# Usage: ./startup.sh [option]
# Options: all, docker, local, stop, status, clean

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"
FRONTEND_DIR="$SCRIPT_DIR/frontend"
LOG_DIR="$SCRIPT_DIR/.startup-logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create log directory
mkdir -p "$LOG_DIR"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing=0

    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL not found. Install with: brew install postgresql@16"
        missing=1
    else
        log_success "PostgreSQL found: $(psql --version)"
    fi

    if ! command -v go &> /dev/null; then
        log_error "Go not found. Install with: brew install go"
        missing=1
    else
        log_success "Go found: $(go version)"
    fi

    if ! command -v node &> /dev/null; then
        log_error "Node.js not found. Install with: brew install node"
        missing=1
    else
        log_success "Node.js found: $(node --version)"
    fi

    if ! command -v npm &> /dev/null; then
        log_error "npm not found. Should be installed with Node.js"
        missing=1
    else
        log_success "npm found: $(npm --version)"
    fi

    if [ $missing -eq 1 ]; then
        log_error "Missing prerequisites. Please install and try again."
        exit 1
    fi
}

# Start services using local Homebrew installations
start_local() {
    log_info "Starting services with local Homebrew installations..."

    # Start PostgreSQL
    log_info "Starting PostgreSQL service..."
    if brew services list | grep -q "postgresql@16.*started"; then
        log_warn "PostgreSQL already running"
    else
        brew services start postgresql@16
        log_success "PostgreSQL started"
        sleep 2
    fi

    # Wait for PostgreSQL to be ready
    log_info "Waiting for PostgreSQL to be ready..."
    local retry=0
    while ! psql -U postgres -h localhost -c "SELECT 1" &>/dev/null; do
        if [ $retry -ge 30 ]; then
            log_error "PostgreSQL failed to start"
            exit 1
        fi
        sleep 1
        ((retry++))
    done
    log_success "PostgreSQL is ready"

    # Initialize database
    log_info "Initializing database..."
    if ! psql -U postgres -h localhost -c "SELECT datname FROM pg_database WHERE datname='businessos'" &>/dev/null | grep -q businessos; then
        psql -U postgres -h localhost << EOF
CREATE DATABASE businessos;
CREATE USER businessos_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE businessos TO businessos_user;
EOF
        log_success "Database created"
    else
        log_warn "Database already exists"
    fi

    # Initialize schema
    if [ -f "$BACKEND_DIR/internal/database/init.sql" ]; then
        if psql -U postgres -h localhost -d businessos -c "\dt" 2>/dev/null | grep -q "user"; then
            log_warn "Schema already initialized"
        else
            psql -U postgres -h localhost -d businessos -f "$BACKEND_DIR/internal/database/init.sql"
            log_success "Schema initialized"
        fi
    fi

    # Verify connection
    if psql -U postgres -h localhost -d businessos -c "SELECT version();" &>/dev/null; then
        log_success "Database connection verified"
    else
        log_error "Failed to connect to database"
        exit 1
    fi
}

# Start services using Docker Compose
start_docker() {
    log_info "Starting services with Docker Compose..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker Desktop"
        exit 1
    fi

    cd "$SCRIPT_DIR"

    log_info "Starting Docker containers..."
    docker-compose up -d

    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    local retry=0
    while [ $retry -lt 60 ]; do
        if docker-compose ps | grep -q "postgres.*healthy" && docker-compose ps | grep -q "redis.*healthy"; then
            log_success "All services are healthy"
            break
        fi
        sleep 1
        ((retry++))
    done

    if [ $retry -ge 60 ]; then
        log_error "Services failed to become healthy"
        docker-compose logs
        exit 1
    fi

    # Verify database connection
    log_info "Verifying database connection..."
    if docker-compose exec -T postgres psql -U postgres -d businessos -c "SELECT version();" &>/dev/null; then
        log_success "Database connection verified"
    else
        log_error "Failed to connect to database"
        exit 1
    fi
}

# Start backend service
start_backend() {
    log_info "Starting Go backend..."

    cd "$BACKEND_DIR"

    # Check for .env file
    if [ ! -f ".env" ]; then
        log_warn ".env file not found. Creating from defaults..."
        cat > .env << EOF
DATABASE_URL=postgres://postgres:password@localhost:5432/businessos?sslmode=disable
SERVER_PORT=8001
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:latest
EOF
    fi

    # Download dependencies
    log_info "Downloading Go dependencies..."
    go mod download
    go mod tidy

    export ALLOWED_ORIGINS="http://localhost:5173,http://localhost:5174,http://localhost:3000"
    export SERVER_PORT="8001"
    log_info "Environment variables set (using port 8001)"

    # Start backend in background
    log_info "Starting server on port 8001..."
    go run cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
    local backend_pid=$!
    echo $backend_pid > "$LOG_DIR/backend.pid"

    # Wait for backend to start
    sleep 3

    if kill -0 $backend_pid 2>/dev/null; then
        log_success "Backend started (PID: $backend_pid)"

        # Verify health endpoint
        if curl -s http://localhost:8001/health | grep -q "healthy"; then
            log_success "Backend is healthy"
        else
            log_warn "Backend health check not yet responding"
        fi
    else
        log_error "Backend failed to start. Check logs:"
        cat "$LOG_DIR/backend.log"
        exit 1
    fi
}

# Start frontend service
start_frontend() {
    log_info "Starting SvelteKit frontend..."

    cd "$FRONTEND_DIR"

    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        log_info "Installing npm dependencies..."
        npm install
    fi

    # Start frontend in background
    log_info "Starting frontend on port 5173..."
    npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    local frontend_pid=$!
    echo $frontend_pid > "$LOG_DIR/frontend.pid"

    sleep 3

    if kill -0 $frontend_pid 2>/dev/null; then
        log_success "Frontend started (PID: $frontend_pid)"
    else
        log_error "Frontend failed to start. Check logs:"
        cat "$LOG_DIR/frontend.log"
        exit 1
    fi
}

# Stop all services
stop_services() {
    log_info "Stopping all services..."

    # Stop backend
    if [ -f "$LOG_DIR/backend.pid" ]; then
        local pid=$(cat "$LOG_DIR/backend.pid")
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "Backend stopped (PID: $pid)"
        fi
        rm "$LOG_DIR/backend.pid"
    fi

    # Stop frontend
    if [ -f "$LOG_DIR/frontend.pid" ]; then
        local pid=$(cat "$LOG_DIR/frontend.pid")
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "Frontend stopped (PID: $pid)"
        fi
        rm "$LOG_DIR/frontend.pid"
    fi

    # Stop Docker services if running
    if docker-compose ps &>/dev/null; then
        log_info "Stopping Docker services..."
        docker-compose down
        log_success "Docker services stopped"
    fi
}

# Show service status
show_status() {
    log_info "Service Status:"
    echo ""

    # PostgreSQL status
    if brew services list | grep -q "postgresql.*started"; then
        log_success "PostgreSQL: Running"
        if psql -U postgres -h localhost -d businessos -c "SELECT 1" &>/dev/null; then
            log_success "  Database connection: OK"
        else
            log_error "  Database connection: FAILED"
        fi
    else
        log_warn "PostgreSQL: Stopped"
    fi

    # Backend status
    if [ -f "$LOG_DIR/backend.pid" ]; then
        local pid=$(cat "$LOG_DIR/backend.pid")
        if kill -0 $pid 2>/dev/null; then
            log_success "Backend: Running (PID: $pid)"
            if curl -s http://localhost:8001/health | grep -q "healthy"; then
                log_success "  Health check: OK"
            else
                log_warn "  Health check: Not responding"
            fi
        else
            log_error "Backend: Stopped (stale PID)"
            rm "$LOG_DIR/backend.pid"
        fi
    else
        log_warn "Backend: Not running"
    fi

    # Frontend status
    if [ -f "$LOG_DIR/frontend.pid" ]; then
        local pid=$(cat "$LOG_DIR/frontend.pid")
        if kill -0 $pid 2>/dev/null; then
            log_success "Frontend: Running (PID: $pid)"
        else
            log_error "Frontend: Stopped (stale PID)"
            rm "$LOG_DIR/frontend.pid"
        fi
    else
        log_warn "Frontend: Not running"
    fi

    echo ""
    echo "Endpoints:"
    echo "  Database:  postgres://postgres@localhost:5432/businessos"
    echo "  Backend:   http://localhost:8001"
    echo "  Frontend:  http://localhost:5173"
    echo "  Redis:     localhost:6379 (optional)"
}

# Clean logs and state
clean_logs() {
    log_info "Cleaning logs and state files..."
    rm -rf "$LOG_DIR"
    mkdir -p "$LOG_DIR"
    log_success "Cleaned"
}

# Main
main() {
    local mode="${1:-all}"

    case "$mode" in
        all)
            log_info "=== BusinessOS Full Startup ==="
            check_prerequisites
            start_local
            start_backend
            start_frontend
            echo ""
            show_status
            echo ""
            log_success "All services started successfully!"
            log_info "Access the application at: http://localhost:5173"
            ;;
        docker)
            log_info "=== BusinessOS Docker Startup ==="
            start_docker
            start_backend
            start_frontend
            echo ""
            show_status
            echo ""
            log_success "All services started successfully!"
            ;;
        local)
            log_info "=== BusinessOS Local Startup ==="
            check_prerequisites
            start_local
            echo ""
            log_success "Local services started!"
            log_info "Now start backend and frontend manually:"
            log_info "  1. cd $BACKEND_DIR && go run cmd/server/main.go"
            log_info "  2. cd $FRONTEND_DIR && npm run dev"
            ;;
        stop)
            stop_services
            ;;
        status)
            show_status
            ;;
        clean)
            clean_logs
            ;;
        *)
            echo "Usage: $0 [option]"
            echo "Options:"
            echo "  all     - Start all services (database + backend + frontend)"
            echo "  docker  - Start using Docker Compose (requires Docker)"
            echo "  local   - Start only database services (run backend/frontend manually)"
            echo "  stop    - Stop all running services"
            echo "  status  - Show service status"
            echo "  clean   - Clean logs and state files"
            exit 1
            ;;
    esac
}

main "$@"
