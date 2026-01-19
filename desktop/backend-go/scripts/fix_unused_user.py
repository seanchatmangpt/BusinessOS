#!/usr/bin/env python3
"""Fix unused user variables after removing redundant auth checks."""

import re
from pathlib import Path

def fix_file(filepath):
    """Replace 'user := GetCurrentUser' with '_ = GetCurrentUser' if user is never used."""
    with open(filepath, 'r') as f:
        lines = f.readlines()

    modified = False
    for i, line in enumerate(lines):
        # If line has: user := middleware.GetCurrentUser(c)
        if re.match(r'\s*user := middleware\.GetCurrentUser\(c\)', line):
            # Check if user is used later in the file (after the auth comment)
            rest_of_file = ''.join(lines[i+2:])  # Skip current line and comment line

            # Check if 'user' appears anywhere after (as variable, not in comment)
            user_pattern = r'\buser\b'
            if not re.search(user_pattern, rest_of_file):
                # User is never used - replace with blank identifier
                lines[i] = line.replace('user :=', '_ =')
                modified = True

    if modified:
        with open(filepath, 'w') as f:
            f.writelines(lines)
        return True
    return False

def main():
    handlers_dir = Path('internal/handlers')

    print("Fixing unused user variables...")
    fixed = 0

    for filepath in sorted(handlers_dir.glob('*.go')):
        if filepath.name.endswith('_test.go'):
            continue

        if fix_file(filepath):
            print(f"✅ {filepath.name}")
            fixed += 1

    print(f"\n✅ Fixed {fixed} files")

if __name__ == '__main__':
    main()
