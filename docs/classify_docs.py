#!/usr/bin/env python3
"""
Classify documentation files by type, category, relevance, and status
"""

import csv
import re
from datetime import datetime, timedelta
from pathlib import Path

def classify_type(path):
    """Determine document type from path and name"""
    path_lower = path.lower()

    # Template files
    if 'template' in path_lower:
        return 'template'

    # Test/verification files
    if any(x in path_lower for x in ['test', 'verification', '_complete.md', 'results']):
        return 'test-report'

    # Architecture/design docs
    if any(x in path_lower for x in ['architecture', 'design', 'diagram', 'system']):
        return 'architecture'

    # API docs
    if 'api' in path_lower and 'reference' in path_lower:
        return 'api-reference'
    if 'api' in path_lower:
        return 'api-guide'

    # Implementation/feature docs
    if any(x in path_lower for x in ['implementation', 'feature', 'integration']):
        return 'implementation'

    # Guides
    if any(x in path_lower for x in ['guide', 'tutorial', 'quickstart', 'getting_started', 'how-to']):
        return 'guide'

    # Status reports
    if any(x in path_lower for x in ['status', 'summary', 'report', 'assessment', 'audit']):
        return 'report'

    # Planning docs
    if any(x in path_lower for x in ['plan', 'roadmap', 'phase', 'sprint']):
        return 'planning'

    # Analysis docs
    if any(x in path_lower for x in ['analysis', 'research', 'review']):
        return 'analysis'

    # Changelog/release notes
    if any(x in path_lower for x in ['changelog', 'release']):
        return 'changelog'

    # Security
    if 'security' in path_lower:
        return 'security'

    # Task/todo docs
    if any(x in path_lower for x in ['task', 'todo', 'checklist']):
        return 'task-list'

    # ADRs
    if 'adr' in path_lower or 'decision' in path_lower:
        return 'adr'

    # Skills
    if 'skill' in path_lower:
        return 'skill'

    # README files
    if 'readme' in path_lower:
        return 'readme'

    # Reference docs
    if 'reference' in path_lower or 'cheatsheet' in path_lower:
        return 'reference'

    return 'documentation'

def classify_category(path):
    """Determine document category"""
    path_lower = path.lower()

    # Frontend
    if any(x in path_lower for x in ['frontend', 'svelte', 'react', 'gesture', 'ui', 'components', 'onboarding', '3d', 'desktop']):
        return 'frontend'

    # Backend
    if any(x in path_lower for x in ['backend', 'api', 'handler', 'service', 'repository', 'golang', 'go']):
        return 'backend'

    # Database
    if any(x in path_lower for x in ['database', 'supabase', 'postgres', 'migration', 'schema']):
        return 'database'

    # Voice/Audio
    if any(x in path_lower for x in ['voice', 'audio', 'livekit', 'vad', 'microphone', 'deepgram']):
        return 'voice'

    # Integrations
    if any(x in path_lower for x in ['integration', 'google', 'oauth', 'sync', 'webhook']):
        return 'integrations'

    # Agents/AI
    if any(x in path_lower for x in ['agent', 'osa', 'claude', 'ai', 'thinking', 'rag']):
        return 'agents'

    # Infrastructure/DevOps
    if any(x in path_lower for x in ['docker', 'deployment', 'cloud', 'infrastructure', 'container', 'gcp']):
        return 'infrastructure'

    # Security
    if 'security' in path_lower:
        return 'security'

    # Testing
    if any(x in path_lower for x in ['test', 'testing']):
        return 'testing'

    # Workspace/team features
    if any(x in path_lower for x in ['workspace', 'team', 'invite', 'permission']):
        return 'workspace'

    # Project management
    if any(x in path_lower for x in ['project', 'task', 'notification', 'dashboard']):
        return 'project-mgmt'

    # Architecture/planning
    if any(x in path_lower for x in ['architecture', 'planning', 'vision', 'roadmap']):
        return 'architecture'

    # Skills
    if 'skill' in path_lower:
        return 'skills'

    return 'general'

def classify_relevance(created_date, modified_date):
    """Determine relevance based on dates"""
    try:
        if created_date == 'untracked' or modified_date == 'untracked':
            return 'active'

        mod_date = datetime.strptime(modified_date, '%Y-%m-%d')
        today = datetime.now()

        # Recent: last 2 weeks
        if mod_date >= today - timedelta(days=14):
            return 'recent'

        # Active: last month
        if mod_date >= today - timedelta(days=30):
            return 'active'

        # Historical: older
        return 'historical'
    except:
        return 'unknown'

def classify_part_of(path, category):
    """Determine which feature/system this belongs to"""
    path_lower = path.lower()

    # OSA Build
    if 'osa_build' in path_lower or 'osa-build' in path_lower:
        return 'OSA Build'

    # Onboarding
    if 'onboarding' in path_lower:
        return 'Onboarding'

    # Voice System
    if any(x in path_lower for x in ['voice', 'livekit', 'vad', 'deepgram']):
        return 'Voice System'

    # Workspace
    if 'workspace' in path_lower:
        return 'Workspace'

    # Custom Agents
    if 'custom_agent' in path_lower or 'agent_system' in path_lower:
        return 'Custom Agents'

    # Integrations
    if 'integration' in path_lower:
        return 'Integrations'

    # 3D Desktop
    if '3d' in path_lower or 'desktop3d' in path_lower:
        return '3D Desktop'

    # Gesture System
    if 'gesture' in path_lower:
        return 'Gesture System'

    # Background Jobs
    if 'background_job' in path_lower or 'job_handler' in path_lower:
        return 'Background Jobs'

    # Sync Engine
    if 'sync' in path_lower or 'nats' in path_lower or 'outbox' in path_lower:
        return 'Sync Engine'

    # Notifications
    if 'notification' in path_lower:
        return 'Notifications'

    # Dashboard
    if 'dashboard' in path_lower:
        return 'Dashboard'

    # Skills
    if 'skill' in path_lower:
        return 'Skills'

    # API
    if '/api/' in path_lower:
        return 'API'

    # Database
    if '/database/' in path_lower:
        return 'Database'

    # Security
    if 'security' in path_lower:
        return 'Security'

    # Testing
    if '/test/' in path_lower:
        return 'Testing'

    # Documentation system
    if '/team-review/' in path_lower or 'documentation_map' in path_lower:
        return 'Documentation'

    # Claude Code workflow
    if '.claude' in path_lower:
        return 'Claude Code'

    return category.title()

def classify_status(path, relevance, doc_type):
    """Determine status of the document"""
    path_lower = path.lower()

    # Archived explicitly
    if '/archive/' in path_lower:
        return 'archived'

    # Templates are always active
    if doc_type == 'template':
        return 'active'

    # Complete/deprecated markers
    if any(x in path_lower for x in ['_complete', '_summary', 'deprecated']):
        return 'complete'

    # Superseded markers
    if any(x in path_lower for x in ['_v2', '_old']):
        return 'superseded'

    # Recent docs are active
    if relevance == 'recent':
        return 'active'

    # Active docs are maintenance
    if relevance == 'active':
        return 'active'

    # Historical are reference
    if relevance == 'historical':
        return 'reference'

    return 'active'

def main():
    # Read the CSV
    input_file = '/tmp/doc_inventory.csv'
    output_file = '/Users/rhl/Desktop/BusinessOS2/docs/DOCUMENTATION_INVENTORY.csv'

    rows = []
    with open(input_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            path = row['Path']
            created = row['Created']
            modified = row['LastModified']

            # Classify
            doc_type = classify_type(path)
            category = classify_category(path)
            relevance = classify_relevance(created, modified)
            part_of = classify_part_of(path, category)
            status = classify_status(path, relevance, doc_type)

            rows.append({
                'Path': path,
                'Created': created,
                'Author': row['Author'],
                'LastModified': modified,
                'LastAuthor': row['LastAuthor'],
                'Type': doc_type,
                'Category': category,
                'Relevance': relevance,
                'PartOf': part_of,
                'Status': status,
                'Lines': row['Lines']
            })

    # Write enriched CSV
    with open(output_file, 'w', newline='') as f:
        fieldnames = ['Path', 'Created', 'Author', 'LastModified', 'LastAuthor',
                     'Type', 'Category', 'Relevance', 'PartOf', 'Status', 'Lines']
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    print(f"✅ Classified {len(rows)} documents")
    print(f"📊 Output: {output_file}")

if __name__ == '__main__':
    main()
