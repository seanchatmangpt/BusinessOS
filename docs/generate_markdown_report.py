#!/usr/bin/env python3
"""
Generate comprehensive markdown documentation inventory
"""

import csv
from collections import defaultdict
from datetime import datetime

def load_data(csv_file):
    """Load CSV data"""
    rows = []
    with open(csv_file, 'r') as f:
        reader = csv.DictReader(f)
        rows = list(reader)
    return rows

def generate_markdown(data):
    """Generate markdown report"""

    # Statistics
    total_docs = len(data)
    total_lines = sum(int(row['Lines']) for row in data)

    # Group by different dimensions
    by_relevance = defaultdict(list)
    by_category = defaultdict(list)
    by_type = defaultdict(list)
    by_status = defaultdict(list)
    by_part_of = defaultdict(list)
    by_author = defaultdict(int)

    for row in data:
        by_relevance[row['Relevance']].append(row)
        by_category[row['Category']].append(row)
        by_type[row['Type']].append(row)
        by_status[row['Status']].append(row)
        by_part_of[row['PartOf']].append(row)
        by_author[row['Author']] += 1

    # Start markdown
    md = []
    md.append("# BusinessOS Documentation Inventory\n")
    md.append(f"**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
    md.append(f"**Total Documents:** {total_docs}")
    md.append(f"**Total Lines:** {total_lines:,}\n")

    # Table of Contents
    md.append("## Table of Contents\n")
    md.append("1. [Executive Summary](#executive-summary)")
    md.append("2. [By Relevance](#by-relevance)")
    md.append("3. [By Category](#by-category)")
    md.append("4. [By Type](#by-type)")
    md.append("5. [By Feature/System](#by-featuresystem)")
    md.append("6. [By Status](#by-status)")
    md.append("7. [By Author](#by-author)")
    md.append("8. [Complete Inventory](#complete-inventory)\n")

    # Executive Summary
    md.append("## Executive Summary\n")
    md.append("### Relevance Distribution\n")
    for relevance in ['recent', 'active', 'historical', 'unknown']:
        count = len(by_relevance.get(relevance, []))
        pct = (count / total_docs * 100) if total_docs > 0 else 0
        md.append(f"- **{relevance.title()}:** {count} docs ({pct:.1f}%)")

    md.append("\n### Category Distribution\n")
    sorted_cats = sorted(by_category.items(), key=lambda x: len(x[1]), reverse=True)
    for cat, docs in sorted_cats[:10]:
        count = len(docs)
        pct = (count / total_docs * 100) if total_docs > 0 else 0
        md.append(f"- **{cat}:** {count} docs ({pct:.1f}%)")

    md.append("\n### Type Distribution\n")
    sorted_types = sorted(by_type.items(), key=lambda x: len(x[1]), reverse=True)
    for doc_type, docs in sorted_types[:10]:
        count = len(docs)
        pct = (count / total_docs * 100) if total_docs > 0 else 0
        md.append(f"- **{doc_type}:** {count} docs ({pct:.1f}%)")

    # By Relevance
    md.append("\n---\n\n## By Relevance\n")

    for relevance in ['recent', 'active', 'historical']:
        docs = by_relevance.get(relevance, [])
        if not docs:
            continue

        md.append(f"\n### {relevance.upper()} ({len(docs)} docs)\n")
        md.append("Documents modified in the last 2 weeks (recent), last month (active), or older (historical).\n")

        # Group by category within relevance
        by_cat = defaultdict(list)
        for doc in docs:
            by_cat[doc['Category']].append(doc)

        for cat in sorted(by_cat.keys()):
            cat_docs = by_cat[cat]
            md.append(f"\n#### {cat.title()} ({len(cat_docs)})\n")
            md.append("| File | Type | Part Of | Status | Lines | Last Modified |")
            md.append("|------|------|---------|--------|-------|---------------|")

            for doc in sorted(cat_docs, key=lambda x: x['LastModified'], reverse=True)[:20]:
                path = doc['Path'].replace('./','')
                # Truncate long paths
                if len(path) > 60:
                    path = '...' + path[-57:]
                md.append(f"| `{path}` | {doc['Type']} | {doc['PartOf']} | {doc['Status']} | {doc['Lines']} | {doc['LastModified']} |")

    # By Category
    md.append("\n---\n\n## By Category\n")

    sorted_cats = sorted(by_category.items(), key=lambda x: len(x[1]), reverse=True)
    for cat, docs in sorted_cats:
        md.append(f"\n### {cat.upper()} ({len(docs)} docs)\n")

        # Group by type within category
        by_t = defaultdict(list)
        for doc in docs:
            by_t[doc['Type']].append(doc)

        for doc_type in sorted(by_t.keys()):
            type_docs = by_t[doc_type]
            if len(type_docs) == 0:
                continue

            md.append(f"\n#### {doc_type.replace('-', ' ').title()} ({len(type_docs)})\n")
            md.append("| File | Part Of | Status | Lines | Last Modified |")
            md.append("|------|---------|--------|-------|---------------|")

            for doc in sorted(type_docs, key=lambda x: x['LastModified'], reverse=True)[:15]:
                path = doc['Path'].replace('./','')
                if len(path) > 60:
                    path = '...' + path[-57:]
                md.append(f"| `{path}` | {doc['PartOf']} | {doc['Status']} | {doc['Lines']} | {doc['LastModified']} |")

    # By Type
    md.append("\n---\n\n## By Type\n")

    sorted_types = sorted(by_type.items(), key=lambda x: len(x[1]), reverse=True)
    for doc_type, docs in sorted_types:
        if len(docs) == 0:
            continue

        md.append(f"\n### {doc_type.replace('-', ' ').upper()} ({len(docs)} docs)\n")
        md.append("| File | Category | Part Of | Status | Lines | Last Modified |")
        md.append("|------|----------|---------|--------|-------|---------------|")

        for doc in sorted(docs, key=lambda x: x['LastModified'], reverse=True)[:20]:
            path = doc['Path'].replace('./','')
            if len(path) > 55:
                path = '...' + path[-52:]
            md.append(f"| `{path}` | {doc['Category']} | {doc['PartOf']} | {doc['Status']} | {doc['Lines']} | {doc['LastModified']} |")

    # By Feature/System
    md.append("\n---\n\n## By Feature/System\n")

    sorted_parts = sorted(by_part_of.items(), key=lambda x: len(x[1]), reverse=True)
    for part, docs in sorted_parts:
        if len(docs) < 3:  # Skip small groups
            continue

        md.append(f"\n### {part} ({len(docs)} docs)\n")
        md.append("| File | Type | Category | Status | Lines | Last Modified |")
        md.append("|------|------|----------|--------|-------|---------------|")

        for doc in sorted(docs, key=lambda x: x['LastModified'], reverse=True)[:15]:
            path = doc['Path'].replace('./','')
            if len(path) > 55:
                path = '...' + path[-52:]
            md.append(f"| `{path}` | {doc['Type']} | {doc['Category']} | {doc['Status']} | {doc['Lines']} | {doc['LastModified']} |")

    # By Status
    md.append("\n---\n\n## By Status\n")

    for status in ['active', 'complete', 'reference', 'archived', 'superseded']:
        docs = by_status.get(status, [])
        if not docs:
            continue

        md.append(f"\n### {status.upper()} ({len(docs)} docs)\n")
        md.append("| File | Type | Category | Part Of | Lines | Last Modified |")
        md.append("|------|------|----------|---------|-------|---------------|")

        for doc in sorted(docs, key=lambda x: x['LastModified'], reverse=True)[:20]:
            path = doc['Path'].replace('./','')
            if len(path) > 50:
                path = '...' + path[-47:]
            md.append(f"| `{path}` | {doc['Type']} | {doc['Category']} | {doc['PartOf']} | {doc['Lines']} | {doc['LastModified']} |")

    # By Author
    md.append("\n---\n\n## By Author\n")
    md.append("| Author | Documents | % of Total |")
    md.append("|--------|-----------|------------|")

    sorted_authors = sorted(by_author.items(), key=lambda x: x[1], reverse=True)
    for author, count in sorted_authors:
        if author == 'unknown':
            continue
        pct = (count / total_docs * 100) if total_docs > 0 else 0
        md.append(f"| {author} | {count} | {pct:.1f}% |")

    # Complete inventory (last 100 most recently modified)
    md.append("\n---\n\n## Complete Inventory\n")
    md.append("### Most Recently Modified (Last 100 files)\n")
    md.append("| File | Type | Category | Part Of | Status | Lines | Modified | Author |")
    md.append("|------|------|----------|---------|--------|-------|----------|--------|")

    all_sorted = sorted(data, key=lambda x: x['LastModified'], reverse=True)
    for doc in all_sorted[:100]:
        path = doc['Path'].replace('./','')
        if len(path) > 45:
            path = '...' + path[-42:]
        author = doc['LastAuthor']
        if len(author) > 20:
            author = author[:17] + '...'
        md.append(f"| `{path}` | {doc['Type']} | {doc['Category']} | {doc['PartOf']} | {doc['Status']} | {doc['Lines']} | {doc['LastModified']} | {author} |")

    return '\n'.join(md)

def main():
    csv_file = '/Users/rhl/Desktop/BusinessOS2/docs/DOCUMENTATION_INVENTORY.csv'
    output_file = '/Users/rhl/Desktop/BusinessOS2/docs/DOCUMENTATION_INVENTORY.md'

    data = load_data(csv_file)
    markdown = generate_markdown(data)

    with open(output_file, 'w') as f:
        f.write(markdown)

    print(f"✅ Generated markdown report")
    print(f"📄 Output: {output_file}")
    print(f"📊 {len(data)} documents analyzed")

if __name__ == '__main__':
    main()
