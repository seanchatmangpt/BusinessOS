#!/bin/bash

# BusinessOS Database Migration Script
# Usage: ./scripts/migrate.sh [command] [args]
# Commands:
#   up [count]      - Apply all pending migrations (or N migrations)
#   down [count]    - Rollback last N migrations (default: 1)
#   status          - Show migration status
#   version         - Show current schema version
#   verify          - Verify checksums of applied migrations
#   help            - Show this help message

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"
BACKEND_DIR="$PROJECT_ROOT/desktop/backend-go"

# Logger functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Check prerequisites
check_prerequisites() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi

    if [ ! -d "$BACKEND_DIR" ]; then
        log_error "Backend directory not found: $BACKEND_DIR"
        exit 1
    fi

    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        log_warn ".env file not found at $PROJECT_ROOT/.env"
        log_info "Migration commands require DATABASE_URL environment variable"
    fi

    if ! command -v go &> /dev/null; then
        log_error "Go compiler not found. Please install Go 1.24+"
        exit 1
    fi

    log_success "Prerequisites verified"
}

# Show help
show_help() {
    cat << 'EOF'
BusinessOS Database Migration Tool

USAGE:
    ./scripts/migrate.sh [command] [args]

COMMANDS:
    up [n]          Apply all pending migrations (or first N pending migrations)
    down [n]        Rollback last N applied migrations (default: 1)
    status          Show current migration status and list all migrations
    version         Show current schema version
    verify          Verify checksums of all applied migrations
    help            Show this help message

ENVIRONMENT VARIABLES:
    DATABASE_URL    PostgreSQL connection string (required for migrate commands)
                    Format: postgres://user:pass@host:port/database

EXAMPLES:
    # Apply all pending migrations
    ./scripts/migrate.sh up

    # Rollback last 3 migrations
    ./scripts/migrate.sh down 3

    # Check current status
    ./scripts/migrate.sh status

    # Verify applied migrations haven't been modified
    ./scripts/migrate.sh verify

MIGRATION FILES:
    Location: ./migrations/
    Naming convention:
        - NNN_name_description.sql      (apply migration)
        - rollback_NNN_name.sql         (rollback migration)

    Where NNN is a 3-digit version number (001, 002, 003, etc.)

BEST PRACTICES:
    1. Write rollback_*.sql before applying migration
    2. Test rollback in staging before production
    3. Never modify a migration after applying to production
    4. Always backup database before running migrations
    5. Schedule large migrations during low-traffic windows
    6. Monitor database during migration execution

TROUBLESHOOTING:
    - Checksum mismatch: Don't modify migration files after applying
    - Connection refused: Check DATABASE_URL and PostgreSQL service status
    - Permission denied: Ensure migration files are readable
    - Rollback not found: Create rollback_*.sql before applying migration

For more information, see ./BusinessOS/docs/migration-guide.md

EOF
}

# Apply migrations
migrate_up() {
    local count=${1:-999}

    log_info "Applying migrations..."
    log_info "Migrations directory: $MIGRATIONS_DIR"

    # List migration files
    local mig_count=0
    for f in "$MIGRATIONS_DIR"/*.sql; do
        if [[ ! "$f" =~ rollback_ ]]; then
            ((mig_count++))
        fi
    done

    if [ "$mig_count" -eq 0 ]; then
        log_warn "No migration files found in $MIGRATIONS_DIR"
        return 0
    fi

    log_info "Found $mig_count migration file(s)"

    # List migrations
    echo ""
    log_info "Migrations to apply:"
    for f in "$MIGRATIONS_DIR"/*_*.sql; do
        if [[ ! "$f" =~ rollback_ ]]; then
            local basename=$(basename "$f")
            echo "  - $basename"
        fi
    done
    echo ""

    # Check if we should use Go migration runner
    if command -v go &> /dev/null && [ -f "$BACKEND_DIR/internal/database/migrate.go" ]; then
        log_info "Using Go migration runner..."
        cd "$PROJECT_ROOT"
        go run "$BACKEND_DIR/cmd/migrate/main.go" --action=up --count="$count" || {
            log_error "Migration failed"
            exit 1
        }
    else
        log_warn "Go migration runner not available, showing manual procedure"
        log_info "To enable automated migrations:"
        log_info "  1. Implement Go migration runner (internal/database/migrate.go)"
        log_info "  2. Create cmd/migrate/main.go entry point"
        log_info "  3. Run: go run ./cmd/migrate --action=up"
        exit 1
    fi

    log_success "Migrations applied successfully"
}

# Rollback migrations
migrate_down() {
    local count=${1:-1}

    if [ "$count" -le 0 ]; then
        log_error "Rollback count must be > 0"
        exit 1
    fi

    log_warn "Rolling back $count migration(s)..."
    log_warn "This operation will DELETE data. Ensure you have a backup!"

    # Confirmation
    echo -n "Are you sure? (type 'yes' to confirm): "
    read -r response

    if [ "$response" != "yes" ]; then
        log_info "Rollback cancelled"
        exit 0
    fi

    # Check for rollback files
    local rollback_count=0
    for f in "$MIGRATIONS_DIR"/rollback_*.sql; do
        if [ -f "$f" ]; then
            ((rollback_count++))
        fi
    done

    if [ "$rollback_count" -lt "$count" ]; then
        log_error "Only $rollback_count rollback file(s) available, but $count requested"
        exit 1
    fi

    log_info "Using Go migration runner..."
    cd "$PROJECT_ROOT"
    go run "$BACKEND_DIR/cmd/migrate/main.go" --action=down --count="$count" || {
        log_error "Rollback failed"
        exit 1
    }

    log_success "Migrations rolled back successfully"
}

# Show migration status
show_status() {
    log_info "Migration Status"
    echo ""

    log_info "Applied migrations:"
    if command -v go &> /dev/null && [ -f "$BACKEND_DIR/internal/database/migrate.go" ]; then
        cd "$PROJECT_ROOT"
        go run "$BACKEND_DIR/cmd/migrate/main.go" --action=status || {
            log_warn "Could not retrieve migration status from database"
        }
    else
        log_warn "Go migration runner not available"
    fi

    echo ""
    log_info "Available migrations on disk:"
    for f in "$MIGRATIONS_DIR"/*_*.sql; do
        if [[ ! "$f" =~ rollback_ ]]; then
            local basename=$(basename "$f")
            local size=$(du -h "$f" | cut -f1)
            echo "  - $basename [$size]"
        fi
    done

    echo ""
    log_info "Available rollbacks:"
    local count=0
    for f in "$MIGRATIONS_DIR"/rollback_*.sql; do
        if [ -f "$f" ]; then
            local basename=$(basename "$f")
            echo "  - $basename"
            ((count++))
        fi
    done

    if [ "$count" -eq 0 ]; then
        log_warn "No rollback files found (migrations cannot be rolled back)"
    fi
}

# Show current version
show_version() {
    log_info "Retrieving current schema version..."

    if command -v go &> /dev/null && [ -f "$BACKEND_DIR/internal/database/migrate.go" ]; then
        cd "$PROJECT_ROOT"
        go run "$BACKEND_DIR/cmd/migrate/main.go" --action=version || {
            log_error "Could not retrieve version"
            exit 1
        }
    else
        log_warn "Go migration runner not available"
        exit 1
    fi
}

# Verify migration checksums
verify_checksums() {
    log_info "Verifying migration checksums..."

    if command -v go &> /dev/null && [ -f "$BACKEND_DIR/internal/database/migrate.go" ]; then
        cd "$PROJECT_ROOT"
        go run "$BACKEND_DIR/cmd/migrate/main.go" --action=verify || {
            log_error "Checksum verification failed"
            exit 1
        }
    else
        log_warn "Go migration runner not available"
        exit 1
    fi
}

# Main entry point
main() {
    local command=${1:-help}
    shift || true

    case "$command" in
        up)
            check_prerequisites
            migrate_up "$@"
            ;;
        down)
            check_prerequisites
            migrate_down "$@"
            ;;
        status)
            check_prerequisites
            show_status
            ;;
        version)
            check_prerequisites
            show_version
            ;;
        verify)
            check_prerequisites
            verify_checksums
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
