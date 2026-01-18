#!/bin/bash

# Extract metadata for all .md files
# Output: CSV with metadata

cd /Users/rhl/Desktop/BusinessOS2

echo "Path,Created,Author,LastModified,LastAuthor,Lines" > /tmp/doc_inventory.csv

find . -name "*.md" -type f | grep -v node_modules | grep -v .git | grep -v venv | while read -r file; do
    # Skip if file doesn't exist (race condition)
    if [ ! -f "$file" ]; then
        continue
    fi

    # Get creation date and author (first commit that added this file)
    creation_info=$(git log --diff-filter=A --format="%ad|%an" --date=short -- "$file" 2>/dev/null | tail -1)

    # Get last modification date and author
    last_mod=$(git log -1 --format="%ad|%an" --date=short -- "$file" 2>/dev/null)

    # Count lines
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")

    # If git info not available (untracked file)
    if [ -z "$creation_info" ]; then
        creation_info="untracked|unknown"
    fi

    if [ -z "$last_mod" ]; then
        last_mod="untracked|unknown"
    fi

    # Parse creation info
    created_date=$(echo "$creation_info" | cut -d'|' -f1)
    created_author=$(echo "$creation_info" | cut -d'|' -f2)

    # Parse last mod info
    modified_date=$(echo "$last_mod" | cut -d'|' -f1)
    modified_author=$(echo "$last_mod" | cut -d'|' -f2)

    # Escape commas in file path and authors
    file_clean=$(echo "$file" | sed 's/,/;/g')
    created_author_clean=$(echo "$created_author" | sed 's/,/;/g')
    modified_author_clean=$(echo "$modified_author" | sed 's/,/;/g')

    echo "$file_clean,$created_date,$created_author_clean,$modified_date,$modified_author_clean,$lines" >> /tmp/doc_inventory.csv
done

echo "Metadata extracted to /tmp/doc_inventory.csv"
