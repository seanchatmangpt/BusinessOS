#!/usr/bin/env python3
"""
Phase 2A: Remove Redundant Auth Checks
Removes 392 duplicate nil checks where AuthMiddleware already validated auth.
"""

import os
import re
import sys
from pathlib import Path

def process_file(filepath):
    """Remove redundant auth checks from a single file."""
    with open(filepath, 'r') as f:
        content = f.read()

    original_content = content

    # Pattern: user := middleware.GetCurrentUser(c)
    #          if user == nil {
    #              c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
    #              return
    #          }

    # Match with varying whitespace (tabs or spaces)
    pattern = re.compile(
        r'(\s*)user := middleware\.GetCurrentUser\(c\)\n'
        r'\s*if user == nil \{\n'
        r'\s*c\.JSON\(http\.StatusUnauthorized, gin\.H\{"error": "Not authenticated"\}\)\n'
        r'\s*return\n'
        r'\s*\}',
        re.MULTILINE
    )

    # Replacement: keep the GetCurrentUser line, add comment
    replacement = r'\1user := middleware.GetCurrentUser(c)\n\1// Auth guaranteed by middleware - user cannot be nil here'

    content = pattern.sub(replacement, content)

    # Count changes
    changes = len(pattern.findall(original_content))

    if content != original_content:
        with open(filepath, 'w') as f:
            f.write(content)

    return changes

def main():
    """Process all handler files."""
    handlers_dir = Path('internal/handlers')

    print("╔══════════════════════════════════════════════════════════════╗")
    print("║ Phase 2A: Removing Redundant Auth Checks                    ║")
    print("╠══════════════════════════════════════════════════════════════╣")
    print("║ Target: 392 duplicate auth checks across 53 handler files   ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print()

    total_changes = 0
    files_modified = 0

    # Find all .go files (excluding tests)
    handler_files = sorted(handlers_dir.glob('*.go'))
    handler_files = [f for f in handler_files if not f.name.endswith('_test.go')]

    for filepath in handler_files:
        changes = process_file(filepath)

        if changes > 0:
            print(f"✅ {filepath.name}: Removed {changes} redundant checks")
            files_modified += 1
            total_changes += changes

    print()
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║ MIGRATION COMPLETE                                           ║")
    print("╠══════════════════════════════════════════════════════════════╣")
    print(f"║ Files modified:      {files_modified:<40} ║")
    print(f"║ Checks removed:      {total_changes:<40} ║")
    print(f"║ Expected:            392                                     ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print()

    if total_changes < 392:
        print(f"⚠️  Note: Removed {total_changes}/392 checks")
        print("   Some checks may have already been removed or have different formatting")

    print("📝 Next steps:")
    print("   1. Review changes: git diff internal/handlers")
    print("   2. Build backend: go build ./cmd/server")
    print("   3. Run tests: go test ./internal/handlers/...")
    print("   4. Commit changes")

if __name__ == '__main__':
    main()
