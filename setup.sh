#!/usr/bin/env bash
# =============================================================================
# BusinessOS Setup Script
# =============================================================================
# One command to set up everything from scratch.
# Run this after cloning the repo:
#
#   ./setup.sh           Full setup (recommended)
#   ./setup.sh --quick   Minimal setup (frontend + backend + SQLite, no extras)
#
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
    echo -e "${BOLD}${BLUE}============================================${RESET}"
    echo -e "${BOLD}${BLUE}  $1${RESET}"
    echo -e "${BOLD}${BLUE}============================================${RESET}"
    echo ""
}

print_step() {
    echo -e "${CYAN}  -->${RESET} $1"
}

print_ok() {
    echo -e "${GREEN}  [OK]${RESET} $1"
}

print_warn() {
    echo -e "${YELLOW}  [WARN]${RESET} $1"
}

print_error() {
    echo -e "${RED}  [ERROR]${RESET} $1"
}

print_fatal() {
    echo ""
    echo -e "${RED}${BOLD}FATAL: $1${RESET}"
    echo ""
    echo -e "${YELLOW}What to do:${RESET} $2"
    echo ""
    exit 1
}

# Show a progress spinner while a command runs
spin() {
    local pid=$!
    local delay=0.1
    local frames=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
    local i=0
    while kill -0 "$pid" 2>/dev/null; do
        printf "\r  ${CYAN}%s${RESET} %s" "${frames[$((i % 10))]}" "$1"
        i=$((i + 1))
        sleep "$delay"
    done
    printf "\r  ${GREEN}[DONE]${RESET} %s\n" "$1"
}

command_exists() {
    command -v "$1" &>/dev/null
}

# ---------------------------------------------------------------------------
# Script root — always relative to this file, not pwd
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---------------------------------------------------------------------------
# Parse flags
# ---------------------------------------------------------------------------
QUICK_MODE=false
SKIP_POSTGRES=false
SKIP_REDIS=false
SKIP_OSA=true   # OSA (Elixir agent) is opt-in

for arg in "$@"; do
    case "$arg" in
        --quick)   QUICK_MODE=true ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --quick   Minimal setup: frontend + backend + SQLite, skips optional services"
            echo "  --help    Show this message"
            exit 0
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Detect operating system
# ---------------------------------------------------------------------------
detect_os() {
    case "$(uname -s)" in
        Darwin) OS="macos" ;;
        Linux)
            if grep -qi microsoft /proc/version 2>/dev/null; then
                OS="wsl"
            else
                OS="linux"
            fi
            ;;
        *)
            print_fatal "Unsupported operating system: $(uname -s)" \
                "BusinessOS supports macOS, Linux, and Windows via WSL2."
            ;;
    esac
    print_ok "Operating system detected: ${BOLD}$OS${RESET}"
}

# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------
clear
echo ""
echo -e "${BOLD}${BLUE}"
echo "  ╔══════════════════════════════════════════════════╗"
echo "  ║          BusinessOS  —  Setup Wizard             ║"
echo "  ╚══════════════════════════════════════════════════╝"
echo -e "${RESET}"
echo -e "  This script will set up everything you need to run BusinessOS."
echo -e "  It usually takes ${BOLD}3–8 minutes${RESET} on a first run."
echo ""

if [ "$QUICK_MODE" = true ]; then
    echo -e "  ${YELLOW}Mode: QUICK (SQLite, no optional services)${RESET}"
else
    echo -e "  ${CYAN}Mode: FULL (PostgreSQL + Redis)${RESET}"
fi
echo ""

# ---------------------------------------------------------------------------
# Pre-flight: make sure we're in the right directory
# ---------------------------------------------------------------------------
if [ ! -f "$SCRIPT_DIR/backend/go.mod" ] || [ ! -d "$SCRIPT_DIR/frontend" ]; then
    print_fatal "This does not look like the BusinessOS root directory." \
        "Run this script from the root of the cloned repository:\n  cd BusinessOS && ./setup.sh"
fi

detect_os

# ---------------------------------------------------------------------------
# SECTION 1: Node.js
# ---------------------------------------------------------------------------
print_header "Checking Node.js"

install_node_via_nvm() {
    print_step "Installing Node.js 20 via nvm..."

    # Install nvm if not present
    if [ ! -f "$HOME/.nvm/nvm.sh" ]; then
        print_step "Downloading nvm (Node Version Manager)..."
        curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash 2>&1 | tail -3 &
        spin "Installing nvm"
    fi

    # Source nvm
    export NVM_DIR="$HOME/.nvm"
    # shellcheck disable=SC1091
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

    nvm install 20 2>&1 | tail -3 &
    spin "Installing Node.js 20"
    nvm use 20 &>/dev/null
    nvm alias default 20 &>/dev/null
}

if command_exists node; then
    NODE_VERSION=$(node --version | sed 's/v//' | cut -d. -f1)
    if [ "$NODE_VERSION" -lt 20 ]; then
        print_warn "Node.js $NODE_VERSION found — version 20+ is required."
        install_node_via_nvm
    else
        print_ok "Node.js $(node --version) — good to go"
    fi
else
    print_warn "Node.js not found — installing via nvm."
    install_node_via_nvm
fi

# Verify node is now available
if ! command_exists node; then
    print_fatal "Node.js installation failed." \
        "Visit https://nodejs.org and install Node.js 20 manually, then re-run this script."
fi

# ---------------------------------------------------------------------------
# SECTION 2: Go
# ---------------------------------------------------------------------------
print_header "Checking Go"

install_go_macos() {
    if ! command_exists brew; then
        print_fatal "Homebrew not found — cannot auto-install Go." \
            "Install Homebrew first: https://brew.sh\nThen install Go: brew install go"
    fi
    print_step "Installing Go via Homebrew (this may take a minute)..."
    brew install go 2>&1 | tail -3 &
    spin "Installing Go"
}

install_go_linux() {
    GO_VERSION="1.24.1"
    print_step "Downloading Go $GO_VERSION..."
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  GO_ARCH="amd64" ;;
        aarch64) GO_ARCH="arm64" ;;
        *)
            print_fatal "Unsupported CPU architecture: $ARCH" \
                "Download Go manually from https://go.dev/dl/"
            ;;
    esac

    GO_TAR="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    curl -fsSL "https://go.dev/dl/${GO_TAR}" -o "/tmp/${GO_TAR}" &
    spin "Downloading Go $GO_VERSION"

    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${GO_TAR}" &
    spin "Extracting Go"
    rm -f "/tmp/${GO_TAR}"

    # Add to PATH for this session
    export PATH="/usr/local/go/bin:$PATH"

    # Add to shell profile if not already there
    PROFILE_FILE="$HOME/.profile"
    if [ -f "$HOME/.bashrc" ]; then PROFILE_FILE="$HOME/.bashrc"; fi
    if [ -f "$HOME/.zshrc" ]; then PROFILE_FILE="$HOME/.zshrc"; fi

    if ! grep -q '/usr/local/go/bin' "$PROFILE_FILE" 2>/dev/null; then
        echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$PROFILE_FILE"
        print_ok "Added Go to PATH in $PROFILE_FILE"
    fi
}

check_go_version() {
    if command_exists go; then
        GO_MAJOR=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1 | sed 's/go//' | cut -d. -f1)
        GO_MINOR=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1 | sed 's/go//' | cut -d. -f2)
        # Need Go 1.24+
        if [ "$GO_MAJOR" -gt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -ge 24 ]; }; then
            return 0  # version OK
        fi
        return 1  # version too old
    fi
    return 1  # not installed
}

if check_go_version; then
    print_ok "Go $(go version | awk '{print $3}') — good to go"
else
    if command_exists go; then
        print_warn "Go $(go version | awk '{print $3}') found — version 1.24+ is required."
    else
        print_warn "Go not found."
    fi

    case "$OS" in
        macos)  install_go_macos ;;
        linux|wsl) install_go_linux ;;
    esac

    if ! check_go_version; then
        print_fatal "Go installation failed." \
            "Visit https://go.dev/dl/ and install Go 1.24+ manually, then re-run this script."
    fi
    print_ok "Go $(go version | awk '{print $3}') installed successfully"
fi

# ---------------------------------------------------------------------------
# SECTION 3: Database — PostgreSQL or SQLite
# ---------------------------------------------------------------------------
print_header "Database Setup"

USE_SQLITE=false

if [ "$QUICK_MODE" = true ]; then
    USE_SQLITE=true
    print_warn "Quick mode: using SQLite (zero config, great for development)."
else
    # Check if PostgreSQL is running
    PG_RUNNING=false
    if command_exists pg_isready && pg_isready -q 2>/dev/null; then
        PG_RUNNING=true
        print_ok "PostgreSQL is running."
    elif command_exists psql; then
        print_warn "PostgreSQL is installed but not running."
    else
        print_warn "PostgreSQL not found."
        echo ""
        echo -e "  BusinessOS can run with ${BOLD}SQLite${RESET} (easiest) or ${BOLD}PostgreSQL${RESET} (recommended for real use)."
        echo ""
        printf "  Use SQLite for now? [Y/n]: "
        read -r PG_CHOICE
        PG_CHOICE="${PG_CHOICE:-Y}"
        if [[ "$PG_CHOICE" =~ ^[Yy]$ ]]; then
            USE_SQLITE=true
        else
            # Try to install PostgreSQL
            case "$OS" in
                macos)
                    if command_exists brew; then
                        print_step "Installing PostgreSQL 16 via Homebrew..."
                        brew install postgresql@16 2>&1 | tail -3 &
                        spin "Installing PostgreSQL"
                        brew services start postgresql@16 &>/dev/null
                        export PATH="/opt/homebrew/opt/postgresql@16/bin:$PATH"
                        PG_RUNNING=true
                    else
                        print_fatal "Cannot install PostgreSQL (Homebrew not found)." \
                            "Install Homebrew (https://brew.sh) then run: brew install postgresql@16 && brew services start postgresql@16"
                    fi
                    ;;
                linux|wsl)
                    print_step "Installing PostgreSQL via apt..."
                    sudo apt-get update -qq
                    sudo apt-get install -y -qq postgresql postgresql-client 2>&1 | tail -3 &
                    spin "Installing PostgreSQL"
                    sudo service postgresql start &>/dev/null || true
                    PG_RUNNING=true
                    ;;
            esac
        fi
    fi

    if [ "$USE_SQLITE" = false ] && [ "$PG_RUNNING" = true ]; then
        # Set up the database and user
        print_step "Creating database 'business_os'..."
        DB_NAME="business_os"
        DB_USER="businessos"
        DB_PASS="businessos_dev"

        # Try to create user and database (ignore errors if they already exist)
        if [ "$OS" = "macos" ]; then
            psql postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';" 2>/dev/null || true
            psql postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || true
            psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null || true
        else
            sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';" 2>/dev/null || true
            sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || true
            sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null || true
        fi
        print_ok "Database '$DB_NAME' ready (user: $DB_USER)"
    fi
fi

# ---------------------------------------------------------------------------
# SECTION 4: Redis (optional)
# ---------------------------------------------------------------------------
if [ "$QUICK_MODE" = false ]; then
    print_header "Redis (Optional)"

    REDIS_AVAILABLE=false
    if command_exists redis-cli && redis-cli ping &>/dev/null; then
        REDIS_AVAILABLE=true
        print_ok "Redis is running — will use for caching."
    else
        print_warn "Redis not found or not running — the app will use an in-memory fallback."
        print_warn "For production use, install Redis: https://redis.io/docs/getting-started/"
        SKIP_REDIS=true
    fi
fi

# ---------------------------------------------------------------------------
# SECTION 5: Elixir / OSA Agent (optional, interactive)
# ---------------------------------------------------------------------------
if [ "$QUICK_MODE" = false ]; then
    print_header "OSA Agent (Optional)"
    echo -e "  The OSA Agent is an optional Elixir-based AI orchestration layer."
    echo -e "  ${YELLOW}You can skip this and add it later.${RESET}"
    echo ""
    printf "  Install OSA Agent (requires Elixir 1.19+ / Erlang OTP 28+)? [y/N]: "
    read -r OSA_CHOICE
    OSA_CHOICE="${OSA_CHOICE:-N}"
    if [[ "$OSA_CHOICE" =~ ^[Yy]$ ]]; then
        SKIP_OSA=false
        if ! command_exists elixir; then
            print_warn "Elixir not found."
            case "$OS" in
                macos)
                    if command_exists brew; then
                        print_step "Installing Elixir via Homebrew..."
                        brew install elixir 2>&1 | tail -3 &
                        spin "Installing Elixir"
                    else
                        print_warn "Cannot auto-install Elixir (Homebrew not found)."
                        print_warn "Visit https://elixir-lang.org/install.html to install manually."
                        SKIP_OSA=true
                    fi
                    ;;
                linux|wsl)
                    print_step "Installing Elixir via apt..."
                    sudo apt-get update -qq
                    sudo apt-get install -y -qq elixir 2>&1 | tail -3 &
                    spin "Installing Elixir"
                    ;;
            esac
        else
            print_ok "Elixir $(elixir --version | head -1) found."
        fi
    else
        print_ok "Skipping OSA Agent — you can set it up later with: ./setup.sh"
    fi
fi

# ---------------------------------------------------------------------------
# SECTION 6: Copy .env file
# ---------------------------------------------------------------------------
print_header "Environment Configuration"

ENV_FILE="$SCRIPT_DIR/backend/.env"
ENV_EXAMPLE="$SCRIPT_DIR/backend/.env.example"

if [ ! -f "$ENV_FILE" ]; then
    if [ -f "$ENV_EXAMPLE" ]; then
        print_step "Copying .env.example → .env"
        cp "$ENV_EXAMPLE" "$ENV_FILE"
    else
        print_step "Creating .env with sensible defaults..."
        # Generate a random secret key (no external deps needed)
        SECRET_KEY=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#$%^&*' </dev/urandom 2>/dev/null | head -c 48 || echo "dev-secret-change-me-in-production-$(date +%s)")

        if [ "$USE_SQLITE" = true ]; then
            DB_URL="sqlite://./businessos.db"
            DB_REQUIRED="false"
        else
            DB_URL="postgres://businessos:businessos_dev@localhost:5432/business_os"
            DB_REQUIRED="true"
        fi

        if [ "$SKIP_REDIS" = true ] || [ "$QUICK_MODE" = true ]; then
            REDIS_URL_VAL="redis://localhost:6379/0"
        else
            REDIS_URL_VAL="redis://localhost:6379/0"
        fi

        cat > "$ENV_FILE" << EOF
# BusinessOS Backend Configuration
# Generated by setup.sh on $(date)
# Edit this file to configure your installation.

# ── Environment ─────────────────────────────────────────────────────────────
ENVIRONMENT=development

# ── Server ──────────────────────────────────────────────────────────────────
SERVER_PORT=8001
BASE_URL=http://localhost:8001

# ── Security ─────────────────────────────────────────────────────────────────
# Auto-generated — safe for local dev. MUST change before deploying to prod.
SECRET_KEY=${SECRET_KEY}
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=1440

# ── Database ─────────────────────────────────────────────────────────────────
DATABASE_URL=${DB_URL}
DATABASE_REQUIRED=${DB_REQUIRED}

# ── Redis ────────────────────────────────────────────────────────────────────
# If Redis isn't running, the app will use an in-memory fallback automatically.
REDIS_URL=${REDIS_URL_VAL}
REDIS_PASSWORD=
REDIS_TLS_ENABLED=false
# SECURITY: Set strong values in production (min 32 chars each)
REDIS_KEY_HMAC_SECRET=
TOKEN_ENCRYPTION_KEY=

# ── AI Provider ──────────────────────────────────────────────────────────────
# Options: ollama_local | ollama_cloud | anthropic | groq | openai
# ollama_local: Free, runs models on your machine (requires Ollama installed)
# ollama_cloud: Uses api.ollama.com (requires API key)
# anthropic:    Uses Claude (requires ANTHROPIC_API_KEY)
# groq:         Fast inference (requires GROQ_API_KEY)
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b

# Fill in the key for your chosen provider:
OLLAMA_CLOUD_API_KEY=
OLLAMA_CLOUD_MODEL=llama3.2
ANTHROPIC_API_KEY=
ANTHROPIC_MODEL=claude-sonnet-4-20250514
GROQ_API_KEY=
GROQ_MODEL=llama-3.3-70b-versatile
OPENAI_API_KEY=
OPENAI_MODEL=gpt-4o-mini
XAI_API_KEY=
XAI_MODEL=grok-beta
XAI_BASE_URL=https://api.x.ai/v1

# ── CORS ─────────────────────────────────────────────────────────────────────
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:3000,app://localhost

# ── OSA Agent (Optional) ─────────────────────────────────────────────────────
OSA_ENABLED=false
OSA_BASE_URL=http://localhost:8089
OSA_SHARED_SECRET=
OSA_TIMEOUT=30
OSA_MAX_RETRIES=3
OSA_RETRY_DELAY=2

# ── OAuth Integrations (all optional) ────────────────────────────────────────
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=http://localhost:8001/api/auth/google/callback/login
GOOGLE_INTEGRATION_REDIRECT_URI=http://localhost:8001/api/integrations/google/callback

SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=http://localhost:8001/api/integrations/slack/callback

NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=http://localhost:8001/api/integrations/notion/callback

HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
HUBSPOT_REDIRECT_URI=http://localhost:8001/api/integrations/hubspot/callback

LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
LINEAR_REDIRECT_URI=http://localhost:8001/api/integrations/linear/callback
LINEAR_WEBHOOK_SECRET=

CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
CLICKUP_REDIRECT_URI=http://localhost:8001/api/integrations/clickup/callback

AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
AIRTABLE_REDIRECT_URI=http://localhost:8001/api/integrations/airtable/callback

MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=http://localhost:8001/api/integrations/microsoft/callback

# ── Web Search (all optional, DuckDuckGo used as free fallback) ──────────────
BRAVE_SEARCH_API_KEY=
SERPER_API_KEY=
TAVILY_API_KEY=
SEARCH_PROVIDER=auto

# ── Web Push Notifications (optional) ────────────────────────────────────────
# Generate: npx web-push generate-vapid-keys
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_CONTACT=mailto:admin@businessos.app

# ── Background Jobs (disabled by default) ────────────────────────────────────
CONVERSATION_SUMMARY_JOB_ENABLED=false
BEHAVIOR_PATTERNS_JOB_ENABLED=false
APP_PROFILER_SYNC_JOB_ENABLED=false

# ── Internal API Security ─────────────────────────────────────────────────────
INTERNAL_API_SECRET=
INTERNAL_ALLOWED_IPS=

# ── Other ────────────────────────────────────────────────────────────────────
SUPERMEMORY_API_KEY=
WEBHOOK_SIGNING_SECRET=
NODE_ID=businessos
NATS_URL=nats://localhost:4222
NATS_ENABLED=false
ENABLE_LOCAL_MODELS=true

# ── Sandbox Containers ────────────────────────────────────────────────────────
SANDBOX_PORT_MIN=9000
SANDBOX_PORT_MAX=9999
SANDBOX_MAX_PER_USER=5
EOF
    fi
    print_ok ".env created at backend/.env"
else
    print_ok ".env already exists — skipping (delete it to regenerate)"
fi

# Also create a frontend .env if it doesn't exist
FRONTEND_ENV="$SCRIPT_DIR/frontend/.env"
if [ ! -f "$FRONTEND_ENV" ]; then
    cat > "$FRONTEND_ENV" << 'EOF'
# Frontend environment (SvelteKit)
# PUBLIC_ variables are exposed to the browser.
PUBLIC_API_URL=http://localhost:8001
PUBLIC_APP_VERSION=dev
EOF
    print_ok "frontend/.env created"
fi

# ---------------------------------------------------------------------------
# SECTION 7: Install frontend dependencies
# ---------------------------------------------------------------------------
print_header "Frontend Dependencies (npm install)"

print_step "Installing frontend packages..."
(cd "$SCRIPT_DIR/frontend" && npm install --silent) &
spin "npm install in frontend/"

print_ok "Frontend dependencies installed"

# ---------------------------------------------------------------------------
# SECTION 8: Install desktop-app dependencies (if present)
# ---------------------------------------------------------------------------
if [ -d "$SCRIPT_DIR/desktop-app" ] && [ -f "$SCRIPT_DIR/desktop-app/package.json" ]; then
    print_header "Desktop App Dependencies"
    print_step "Installing desktop-app packages..."
    (cd "$SCRIPT_DIR/desktop-app" && npm install --silent) &
    spin "npm install in desktop-app/"
    print_ok "Desktop app dependencies installed"
fi

# ---------------------------------------------------------------------------
# SECTION 9: Go dependencies
# ---------------------------------------------------------------------------
print_header "Backend Dependencies (go mod download)"

print_step "Downloading Go modules..."
(cd "$SCRIPT_DIR/backend" && go mod download 2>&1) &
spin "go mod download"

print_ok "Go modules downloaded"

# ---------------------------------------------------------------------------
# SECTION 10: Database migrations
# ---------------------------------------------------------------------------
print_header "Database Migrations"

if [ "$USE_SQLITE" = true ]; then
    print_warn "SQLite mode: skipping PostgreSQL migrations."
    print_warn "The backend will create its schema automatically on first run."
else
    # Check if migrate tool is available or use built-in migration runner
    MIGRATIONS_DIR="$SCRIPT_DIR/backend/migrations"
    if [ -d "$MIGRATIONS_DIR" ] && [ "$(ls -A "$MIGRATIONS_DIR" 2>/dev/null)" ]; then
        print_step "Running database migrations..."

        # Check if golang-migrate is available
        if command_exists migrate; then
            (migrate -path "$MIGRATIONS_DIR" -database "postgres://businessos:businessos_dev@localhost:5432/business_os?sslmode=disable" up 2>&1) &
            spin "Running migrations"
        else
            # Use psql directly to apply migration files in order
            if command_exists psql; then
                for SQL_FILE in $(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort); do
                    FNAME=$(basename "$SQL_FILE")
                    print_step "Applying $FNAME..."
                    if [ "$OS" = "macos" ]; then
                        psql "postgres://businessos:businessos_dev@localhost:5432/business_os" \
                            -f "$SQL_FILE" &>/dev/null || true
                    else
                        sudo -u postgres psql "business_os" -f "$SQL_FILE" &>/dev/null || true
                    fi
                done
                print_ok "Migrations applied"
            else
                print_warn "psql not found — migrations skipped. Run them manually after starting PostgreSQL."
            fi
        fi
    else
        print_warn "No migration files found — the backend will initialize the schema on first start."
    fi
fi

# ---------------------------------------------------------------------------
# SECTION 11: Build Go backend binary
# ---------------------------------------------------------------------------
print_header "Building Backend"

print_step "Compiling Go backend..."
BIN_DIR="$SCRIPT_DIR/bin"
mkdir -p "$BIN_DIR"

(
    cd "$SCRIPT_DIR/backend"
    go build \
        -ldflags="-s -w -X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" \
        -o "$BIN_DIR/businessos-server" \
        ./cmd/server/
) &
spin "go build ./cmd/server/"

if [ ! -f "$BIN_DIR/businessos-server" ]; then
    print_fatal "Backend build failed." \
        "Run 'cd backend && go build ./cmd/server/' to see the full error, then fix it and re-run setup.sh."
fi

print_ok "Backend binary built at bin/businessos-server"

# ---------------------------------------------------------------------------
# SECTION 12: Make helper scripts executable
# ---------------------------------------------------------------------------
print_header "Finalizing"

for SCRIPT in "$SCRIPT_DIR/start.sh" "$SCRIPT_DIR/stop.sh"; do
    if [ -f "$SCRIPT" ]; then
        chmod +x "$SCRIPT"
        print_ok "$(basename "$SCRIPT") is executable"
    fi
done

if [ -f "$SCRIPT_DIR/desktop-app/scripts/install-autostart.sh" ]; then
    chmod +x "$SCRIPT_DIR/desktop-app/scripts/install-autostart.sh"
    print_ok "desktop-app/scripts/install-autostart.sh is executable"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${BOLD}${GREEN}"
echo "  ╔══════════════════════════════════════════════════╗"
echo "  ║          Setup Complete!                         ║"
echo "  ╚══════════════════════════════════════════════════╝"
echo -e "${RESET}"
echo -e "  ${BOLD}What was set up:${RESET}"
echo -e "    ${GREEN}✓${RESET} Node.js $(node --version)"
echo -e "    ${GREEN}✓${RESET} Go $(go version | awk '{print $3}')"
if [ "$USE_SQLITE" = true ]; then
    echo -e "    ${GREEN}✓${RESET} Database: SQLite (file-based, zero config)"
else
    echo -e "    ${GREEN}✓${RESET} Database: PostgreSQL"
fi
if [ "$SKIP_REDIS" = false ] && [ "$QUICK_MODE" = false ]; then
    echo -e "    ${GREEN}✓${RESET} Redis: connected"
else
    echo -e "    ${YELLOW}~${RESET} Redis: in-memory fallback (install Redis for production)"
fi
echo -e "    ${GREEN}✓${RESET} Backend binary: bin/businessos-server"
echo -e "    ${GREEN}✓${RESET} Frontend packages: frontend/node_modules"
echo -e "    ${GREEN}✓${RESET} Environment: backend/.env"
echo ""
echo -e "  ${BOLD}Next steps:${RESET}"
echo ""
echo -e "  1. ${CYAN}Start everything:${RESET}"
echo -e "     ${BOLD}./start.sh${RESET}"
echo ""
echo -e "  2. ${CYAN}Open your browser:${RESET}"
echo -e "     ${BOLD}http://localhost:5173${RESET}"
echo ""
echo -e "  3. ${CYAN}Configure your AI provider${RESET} (optional but recommended):"
echo -e "     Edit ${BOLD}backend/.env${RESET} and set one of:"
echo -e "     • ANTHROPIC_API_KEY (for Claude)"
echo -e "     • GROQ_API_KEY (for fast free inference)"
echo -e "     • Or install Ollama for 100% local AI: ${BOLD}https://ollama.ai${RESET}"
echo ""
echo -e "  4. ${CYAN}Auto-start on login${RESET} (for the desktop app):"
echo -e "     ${BOLD}./desktop-app/scripts/install-autostart.sh${RESET}"
echo ""
echo -e "  ${YELLOW}Tip:${RESET} Run ${BOLD}make help${RESET} to see all available commands."
echo ""
