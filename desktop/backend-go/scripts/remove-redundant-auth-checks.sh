#!/bin/bash
set -e

# Phase 2A: Remove Redundant Auth Checks
# This script removes 392 redundant nil checks in handlers where AuthMiddleware already validated auth
#
# What it does:
# 1. Finds all handler files
# 2. Removes the 4-line pattern:
#      user := middleware.GetCurrentUser(c)
#      if user == nil {
#          c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
#          return
#      }
# 3. Replaces with:
#      user := middleware.GetCurrentUser(c)
#      // Auth guaranteed by middleware - user cannot be nil here

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║ Phase 2A: Removing Redundant Auth Checks                    ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║ Target: 392 duplicate auth checks across 53 handler files   ║"
echo "║ Pattern: if user == nil { return 401 }                      ║"
echo "║ Reason: AuthMiddleware already validates - check is         ║"
echo "║         unreachable code                                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

cd "$(dirname "$0")/.." || exit 1

# Counters
FILES_MODIFIED=0
CHECKS_REMOVED=0

# Create backup directory
BACKUP_DIR="./backups/phase2a-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "📦 Backup directory: $BACKUP_DIR"
echo ""

# Find all handler files (excluding tests)
HANDLER_FILES=$(find internal/handlers -name "*.go" -type f ! -name "*_test.go")

for file in $HANDLER_FILES; do
    # Skip if file doesn't contain the pattern
    if ! grep -q 'user := middleware\.GetCurrentUser(c)' "$file"; then
        continue
    fi

    # Create backup
    cp "$file" "$BACKUP_DIR/$(basename "$file")"

    # Count occurrences before
    BEFORE=$(grep -c 'if user == nil {' "$file" || true)

    if [ "$BEFORE" -eq 0 ]; then
        continue
    fi

    echo "🔧 Processing: $file"
    echo "   Checks found: $BEFORE"

    # Create temporary file for processing
    TEMP_FILE=$(mktemp)

    # Use awk to remove the redundant check pattern
    awk '
    /user := middleware\.GetCurrentUser\(c\)/ {
        print
        getline
        # Check if next line is "if user == nil {"
        if ($0 ~ /^\s*if user == nil \{/) {
            # Skip the next 3 lines (the if block)
            getline  # Skip: c.JSON(...)
            getline  # Skip: return
            getline  # Skip: }
            # Add comment
            print "\t// Auth guaranteed by middleware - user cannot be nil here"
        } else {
            # Not the pattern, print the line
            print
        }
        next
    }
    { print }
    ' "$file" > "$TEMP_FILE"

    # Replace original file
    mv "$TEMP_FILE" "$file"

    # Count occurrences after
    AFTER=$(grep -c 'if user == nil {' "$file" || true)
    REMOVED=$((BEFORE - AFTER))

    if [ "$REMOVED" -gt 0 ]; then
        echo "   ✅ Removed: $REMOVED checks"
        FILES_MODIFIED=$((FILES_MODIFIED + 1))
        CHECKS_REMOVED=$((CHECKS_REMOVED + REMOVED))
    else
        echo "   ⚠️  No changes (pattern may be different)"
    fi
    echo ""
done

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║ MIGRATION COMPLETE                                           ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║ Files modified:      $FILES_MODIFIED                                    ║"
echo "║ Checks removed:      $CHECKS_REMOVED                                   ║"
echo "║ Expected:            392                                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

if [ "$CHECKS_REMOVED" -lt 392 ]; then
    echo "⚠️  WARNING: Removed fewer checks than expected"
    echo "   Some checks may have different formatting"
    echo "   Review manually with: git diff internal/handlers"
fi

echo "📝 Next steps:"
echo "   1. Review changes: git diff internal/handlers"
echo "   2. Build backend: go build ./cmd/server"
echo "   3. Run tests: go test ./internal/handlers/..."
echo "   4. Commit: git add internal/handlers && git commit -m 'refactor: remove 392 redundant auth checks (Phase 2A)'"
echo ""
echo "   To restore backups: cp $BACKUP_DIR/* internal/handlers/"
