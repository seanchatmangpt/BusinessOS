# BusinessOS Documentation

## Structure

```
docs/
├── status/          # Status reports and completion summaries
├── guides/          # How-to guides and tutorials  
├── architecture/    # Architecture decisions and design docs
├── features/        # Feature-specific documentation
└── README.md        # This file
```

## Quick Links

### Background Jobs System (NEW)
- [MASTER DOCUMENTATION](BACKGROUND_JOBS_COMPLETE_DOCUMENTATION.md) - Complete guide (58KB, 15,000 words)
- [Quick Start](BACKGROUND_JOBS_QUICKSTART.md) - Get started in 5 minutes
- [Final Summary](FINAL_SUMMARY.md) - Portuguese executive summary
- [Delivery Checklist](DELIVERY_CHECKLIST.md) - Complete verification

### Implementation Status
- [Linear Issues Implementation](LINEAR_ISSUES_IMPLEMENTATION.md) - Complete Q1 implementation
- [Latest Status](status/) - Recent completion reports

### API Reference
- [API Endpoints](API_ENDPOINTS_REFERENCE.md)
- [Testing Guide](TESTING_GUIDE.md)

### Architecture
- [Database Schema](database_troubleshooting.md)
- [Workspace Implementation](WORKSPACE_IMPLEMENTATION_TECHNICAL_SUMMARY.md)
- [Memory Hierarchy](workspace_memory_ui_guide.md)

## For Developers

### Claude Code Workflow
- [**CLAUDE.md (root)**](../CLAUDE.md) - Guia completo do workflow com subagents
- [**ADVANCED_TASKMANAGER.md**](ADVANCED_TASKMANAGER.md) - Sistema completo: Microtasks, Milestones, Feedback Loop
- [**TASKMANAGER_EXAMPLES.md**](TASKMANAGER_EXAMPLES.md) - TaskManager automático em ação
- [**WORKFLOW_EXAMPLE.md**](WORKFLOW_EXAMPLE.md) - Exemplo prático de decomposição em subtasks

**Starting development:**
```bash
# From project root
./dev.sh
```

**Running tests:**
```bash
cd desktop/backend-go
./scripts/tests/test_all_endpoints.sh
```

**Database migrations:**
```bash
go run scripts/migrations/run_q1_migrations.go
```

## Documentation Standards

1. **File naming:**
   - Use snake_case: `feature_name_guide.md`
   - Include date for status reports: `2026-01-06_status.md`

2. **Location:**
   - Root: Only README.md and CLAUDE.md
   - Everything else: organized in docs/

3. **Keep updated:**
   - Update docs when features change
   - Archive old status reports in status/
