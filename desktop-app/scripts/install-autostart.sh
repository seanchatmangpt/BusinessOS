#!/usr/bin/env bash
# =============================================================================
# BusinessOS Desktop — Auto-start Installer
# =============================================================================
# Makes the BusinessOS desktop app launch automatically when you log in.
#
# Usage:
#   ./install-autostart.sh              Install auto-start
#   ./install-autostart.sh --uninstall  Remove auto-start
#   ./install-autostart.sh --status     Check if auto-start is installed
#
# Supports: macOS (LaunchAgent) and Linux/WSL (XDG autostart .desktop file)
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

print_ok()   { echo -e "  ${GREEN}[OK]${RESET}    $1"; }
print_warn() { echo -e "  ${YELLOW}[WARN]${RESET}  $1"; }
print_info() { echo -e "  ${CYAN}[INFO]${RESET}  $1"; }
print_error(){ echo -e "  ${RED}[ERROR]${RESET} $1"; }

print_fatal() {
    echo ""
    echo -e "${RED}${BOLD}Error: $1${RESET}"
    echo -e "${YELLOW}What to do: $2${RESET}"
    echo ""
    exit 1
}

command_exists() { command -v "$1" &>/dev/null; }

# ---------------------------------------------------------------------------
# Detect OS
# ---------------------------------------------------------------------------
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
        print_fatal "Unsupported OS: $(uname -s)" \
            "Auto-start is supported on macOS and Linux. On Windows, use Task Scheduler."
        ;;
esac

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
ACTION="install"

for arg in "$@"; do
    case "$arg" in
        --uninstall) ACTION="uninstall" ;;
        --status)    ACTION="status" ;;
        --help|-h)
            echo "Usage: $0 [--uninstall | --status]"
            echo ""
            echo "  (no flag)     Install auto-start on login"
            echo "  --uninstall   Remove auto-start"
            echo "  --status      Check whether auto-start is installed"
            exit 0
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Locate the BusinessOS app / executable
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# desktop-app/scripts/ → desktop-app/ → project-root/
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DESKTOP_APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Bundle ID used in plist file name (must match forge.config.ts)
BUNDLE_ID="com.businessos.desktop"
APP_NAME="BusinessOS"

# ---------------------------------------------------------------------------
# Find the packaged app or dev electron binary
# ---------------------------------------------------------------------------
find_app_executable() {
    local app_path=""

    case "$OS" in
        macos)
            # Check for packaged .app in standard locations
            for candidate in \
                "$DESKTOP_APP_DIR/out/${APP_NAME}.app/Contents/MacOS/${APP_NAME}" \
                "$DESKTOP_APP_DIR/out/${APP_NAME}-darwin-arm64/${APP_NAME}.app/Contents/MacOS/${APP_NAME}" \
                "$DESKTOP_APP_DIR/out/${APP_NAME}-darwin-x64/${APP_NAME}.app/Contents/MacOS/${APP_NAME}" \
                "/Applications/${APP_NAME}.app/Contents/MacOS/${APP_NAME}"; do
                if [ -f "$candidate" ]; then
                    app_path="$candidate"
                    break
                fi
            done

            # Fall back to electron-forge dev binary
            if [ -z "$app_path" ]; then
                if [ -f "$DESKTOP_APP_DIR/node_modules/.bin/electron" ]; then
                    app_path="$DESKTOP_APP_DIR/node_modules/.bin/electron"
                fi
            fi
            ;;

        linux|wsl)
            # Check for packaged binary
            for candidate in \
                "$DESKTOP_APP_DIR/out/${APP_NAME}-linux-x64/${APP_NAME}" \
                "$DESKTOP_APP_DIR/out/make/deb/x64/${APP_NAME,,}"_*_amd64.deb \
                "/usr/bin/${APP_NAME,,}" \
                "/usr/local/bin/${APP_NAME,,}"; do
                if [ -f "$candidate" ]; then
                    app_path="$candidate"
                    break
                fi
            done

            # Fall back to electron dev binary
            if [ -z "$app_path" ]; then
                if [ -f "$DESKTOP_APP_DIR/node_modules/.bin/electron" ]; then
                    app_path="$DESKTOP_APP_DIR/node_modules/.bin/electron"
                fi
            fi
            ;;
    esac

    echo "$app_path"
}

APP_EXEC=$(find_app_executable)

if [ -z "$APP_EXEC" ]; then
    if [ "$ACTION" != "uninstall" ] && [ "$ACTION" != "status" ]; then
        echo -e "${YELLOW}"
        echo "  The packaged BusinessOS app was not found."
        echo "  Auto-start will be configured for the development version (npm start)."
        echo -e "${RESET}"
        # Use npm start as the launch command
        USE_NPM_START=true
    fi
else
    USE_NPM_START=false
fi

# ---------------------------------------------------------------------------
# macOS: LaunchAgent plist
# ---------------------------------------------------------------------------
LAUNCHAGENTS_DIR="$HOME/Library/LaunchAgents"
PLIST_FILE="$LAUNCHAGENTS_DIR/${BUNDLE_ID}.plist"

install_macos() {
    mkdir -p "$LAUNCHAGENTS_DIR"

    if [ "${USE_NPM_START:-false}" = true ]; then
        # Launch via npm run start (for development)
        NPM_BIN=$(command -v npm)
        cat > "$PLIST_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Auto-start label — must be unique on the system -->
    <key>Label</key>
    <string>${BUNDLE_ID}</string>

    <!-- Command to run on login -->
    <key>ProgramArguments</key>
    <array>
        <string>${NPM_BIN}</string>
        <string>run</string>
        <string>start</string>
    </array>

    <!-- Working directory -->
    <key>WorkingDirectory</key>
    <string>${DESKTOP_APP_DIR}</string>

    <!-- Run at login -->
    <key>RunAtLoad</key>
    <true/>

    <!-- Keep alive if it crashes (optional — remove this block to disable) -->
    <key>KeepAlive</key>
    <false/>

    <!-- Redirect stdout/stderr to log files -->
    <key>StandardOutPath</key>
    <string>${PROJECT_ROOT}/logs/desktop-stdout.log</string>
    <key>StandardErrorPath</key>
    <string>${PROJECT_ROOT}/logs/desktop-stderr.log</string>

    <!-- Delay before starting (seconds) — gives other login items time to load -->
    <key>StartInterval</key>
    <integer>5</integer>

    <!-- Environment variables passed to the process -->
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin</string>
        <key>HOME</key>
        <string>${HOME}</string>
    </dict>
</dict>
</plist>
EOF
    else
        # Launch the packaged binary directly
        cat > "$PLIST_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${BUNDLE_ID}</string>

    <key>ProgramArguments</key>
    <array>
        <string>${APP_EXEC}</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <false/>

    <key>StandardOutPath</key>
    <string>${PROJECT_ROOT}/logs/desktop-stdout.log</string>
    <key>StandardErrorPath</key>
    <string>${PROJECT_ROOT}/logs/desktop-stderr.log</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin</string>
        <key>HOME</key>
        <string>${HOME}</string>
    </dict>
</dict>
</plist>
EOF
    fi

    # Register with launchctl so it starts immediately (not just on next login)
    launchctl load -w "$PLIST_FILE" 2>/dev/null || true

    print_ok "LaunchAgent installed: $PLIST_FILE"
    print_ok "BusinessOS will now start automatically when you log in."
    echo ""
    print_info "To start it right now (without logging out):"
    echo -e "    ${BOLD}launchctl start ${BUNDLE_ID}${RESET}"
    echo ""
    print_info "To stop it right now:"
    echo -e "    ${BOLD}launchctl stop ${BUNDLE_ID}${RESET}"
    echo ""
    print_info "To remove auto-start:"
    echo -e "    ${BOLD}$0 --uninstall${RESET}"
}

uninstall_macos() {
    if [ ! -f "$PLIST_FILE" ]; then
        print_warn "LaunchAgent not found at $PLIST_FILE — already uninstalled?"
        return
    fi

    # Unload before removing
    launchctl unload -w "$PLIST_FILE" 2>/dev/null || true
    rm -f "$PLIST_FILE"
    print_ok "LaunchAgent removed. BusinessOS will no longer auto-start on login."
}

status_macos() {
    if [ -f "$PLIST_FILE" ]; then
        echo -e "  ${GREEN}[INSTALLED]${RESET} Auto-start is enabled."
        echo -e "              Plist: $PLIST_FILE"
        # Show launchctl status
        LCTL_STATUS=$(launchctl list "${BUNDLE_ID}" 2>/dev/null || echo "not loaded")
        echo -e "              Status: $LCTL_STATUS"
    else
        echo -e "  ${YELLOW}[NOT INSTALLED]${RESET} Auto-start is not configured."
        echo ""
        echo -e "  Run ${BOLD}$0${RESET} to enable it."
    fi
}

# ---------------------------------------------------------------------------
# Linux / WSL: XDG Autostart .desktop file
# ---------------------------------------------------------------------------
AUTOSTART_DIR="$HOME/.config/autostart"
DESKTOP_FILE="$AUTOSTART_DIR/${BUNDLE_ID}.desktop"

install_linux() {
    if [ "$OS" = "wsl" ]; then
        print_warn "WSL detected — XDG autostart may not work inside WSL."
        print_warn "For Windows auto-start, add a shortcut to:"
        print_warn "  %APPDATA%\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
        echo ""
    fi

    mkdir -p "$AUTOSTART_DIR"
    mkdir -p "$PROJECT_ROOT/logs"

    if [ "${USE_NPM_START:-false}" = true ]; then
        NPM_BIN=$(command -v npm)
        EXEC_LINE="bash -c 'cd ${DESKTOP_APP_DIR} && ${NPM_BIN} run start >> ${PROJECT_ROOT}/logs/desktop.log 2>&1'"
    else
        EXEC_LINE="${APP_EXEC}"
    fi

    cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
# BusinessOS auto-start entry
# Generated by install-autostart.sh on $(date)
Type=Application
Name=${APP_NAME}
Comment=BusinessOS — Your business command center
Exec=${EXEC_LINE}
Icon=${DESKTOP_APP_DIR}/resources/icons/icon.png
Terminal=false
Categories=Office;Business;
StartupNotify=true
Hidden=false
X-GNOME-Autostart-enabled=true
X-GNOME-Autostart-Delay=5
EOF

    chmod +x "$DESKTOP_FILE"

    print_ok "Autostart entry installed: $DESKTOP_FILE"
    print_ok "BusinessOS will now start automatically when you log in."
    echo ""
    print_info "To remove auto-start:"
    echo -e "    ${BOLD}$0 --uninstall${RESET}"
}

uninstall_linux() {
    if [ ! -f "$DESKTOP_FILE" ]; then
        print_warn ".desktop file not found at $DESKTOP_FILE — already uninstalled?"
        return
    fi

    rm -f "$DESKTOP_FILE"
    print_ok ".desktop file removed. BusinessOS will no longer auto-start on login."
}

status_linux() {
    if [ -f "$DESKTOP_FILE" ]; then
        echo -e "  ${GREEN}[INSTALLED]${RESET} Auto-start is enabled."
        echo -e "              File: $DESKTOP_FILE"
    else
        echo -e "  ${YELLOW}[NOT INSTALLED]${RESET} Auto-start is not configured."
        echo ""
        echo -e "  Run ${BOLD}$0${RESET} to enable it."
    fi
}

# ---------------------------------------------------------------------------
# Dispatch
# ---------------------------------------------------------------------------
echo ""
echo -e "${BOLD}${BLUE}  BusinessOS Desktop — Auto-start ${ACTION^}${RESET}"
echo -e "${BLUE}  ──────────────────────────────────────────${RESET}"
echo ""

case "$ACTION" in
    install)
        case "$OS" in
            macos)        install_macos ;;
            linux|wsl)    install_linux ;;
        esac
        ;;
    uninstall)
        case "$OS" in
            macos)        uninstall_macos ;;
            linux|wsl)    uninstall_linux ;;
        esac
        ;;
    status)
        case "$OS" in
            macos)        status_macos ;;
            linux|wsl)    status_linux ;;
        esac
        ;;
esac

echo ""
