# Documentation Migration Guide

**For team members adapting to the new structure**

Date: 2026-01-19

---

## What Changed?

We reorganized 209 docs from a flat/mixed structure into a **4-tier relevance-based system**.

New structure:
- `0-START-HERE/` - 9 essential entry points
- `1-active/` - 37 recent docs (last 30 days)
- `2-reference/` - 48 timeless docs
- `3-completed/` - 39 historical docs
- `archive/` - 65 deprecated docs

---

## Finding Your Old Docs

### Quick Lookup

| Old Location | New Location |
|-------------|-------------|
| `docs/README.md` | `docs/0-START-HERE/README.md` |
| `docs/START_HERE.md` | `docs/0-START-HERE/START_HERE.md` |
| `docs/TEAM_START_HERE.md` | `docs/0-START-HERE/TEAM_START_HERE.md` |
| `docs/CODEBASE_CLEANUP_*.md` | `docs/1-active/reports/2026-01-19-*.md` |
| `docs/OSA_BUILD_*.md` | `docs/1-active/features/onboarding/2026-01-*.md` |
| `docs/AGENT_*.md` | `docs/1-active/features/agents/*.md` |
| `docs/architecture/` | `docs/2-reference/architecture/` |
| `docs/development/` | `docs/2-reference/development/` |
| `docs/planning/` | `docs/3-completed/2026-Q1/planning/` |

### Use Git to Find Moves

```bash
git log --follow --name-status -- docs/YOUR_FILE.md
```

---

## New Naming Conventions

### Reports Use Date Prefixes

Format: `YYYY-MM-DD-description.md`

Example: `2026-01-19-codebase-cleanup-report.md`

---

## Where to Put New Docs?

- Entry point? → `0-START-HERE/`
- Current work (30 days)? → `1-active/`
- Timeless guide? → `2-reference/`
- Completed work? → `3-completed/YYYY-QX/`
- Deprecated? → `archive/`

---

## Document Lifecycle

1. **Create** in appropriate directory
2. **Active** in `1-active/` (30 days)
3. **Complete** move to `3-completed/YYYY-QX/`
4. **Archive** after 1 year or when deprecated

---

## Checklist for New Docs

- [ ] Correct directory
- [ ] Date prefix (if report)
- [ ] Grouped with related docs
- [ ] Last Updated date at bottom
- [ ] Update directory README if needed

---

## Where to Find Things Now

- Start here: `0-START-HERE/TEAM_START_HERE.md`
- Recent reports: `1-active/reports/`
- Onboarding: `1-active/features/onboarding/`
- Agents: `1-active/features/agents/`
- Architecture: `2-reference/architecture/`
- Dev setup: `2-reference/development/`
- Past work: `3-completed/2026-Q1/`

---

**Need Help?** Check INDEX.md or ask in team chat

**Migration Date:** 2026-01-19
**Git History:** Fully preserved (152 tracked moves)
