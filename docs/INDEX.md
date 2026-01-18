# BusinessOS2 Documentation Index

**Reorganized by relevance and recency - 2026-01-19**

Total Documentation: 209 markdown files

---

## 🎯 Quick Start

**New to the project?** Start here:

1. Read [`0-START-HERE/TEAM_START_HERE.md`](0-START-HERE/TEAM_START_HERE.md)
2. Check [`0-START-HERE/RECENT_CHANGES.md`](0-START-HERE/RECENT_CHANGES.md)
3. Review [`0-START-HERE/EXECUTIVE_SUMMARY.md`](0-START-HERE/EXECUTIVE_SUMMARY.md)

---

## 📂 Directory Structure

```
docs/
├── 0-START-HERE/       9 docs  - Essential entry points
├── 1-active/          37 docs  - Last 30 days work
│   ├── reports/              (10 recent reports)
│   ├── features/             (onboarding, agents)
│   ├── plans/                (Q1 plans)
│   └── team-review/          (PR reviews, sprints)
├── 2-reference/       48 docs  - Timeless reference
│   ├── architecture/         (7 architecture docs)
│   ├── development/          (9 dev guides)
│   ├── deployment/           (3 deployment guides)
│   ├── guides/               (8 how-to guides)
│   ├── integrations/         (3 integration docs)
│   ├── database/             (2 schema docs)
│   ├── research/             (5 research docs)
│   └── [6 more subdirectories]
├── 3-completed/       39 docs  - Historical work
│   ├── 2026-Q1/              (planning, features, implementation)
│   ├── 2025-Q4/              (SORX 2.0, older work)
│   └── 2025-Q3/              (historical)
└── archive/           65 docs  - Deprecated/old
    └── [well-organized historical archive]
```

---

## 📋 Navigation by Need

### I want to...

**Get started on the project**
→ [`0-START-HERE/`](0-START-HERE/)

**See what's happening now**
→ [`1-active/`](1-active/)

**Learn how the system works**
→ [`2-reference/architecture/`](2-reference/architecture/)

**Set up my development environment**
→ [`2-reference/development/`](2-reference/development/)

**Deploy to production**
→ [`2-reference/deployment/`](2-reference/deployment/)

**Find a how-to guide**
→ [`2-reference/guides/`](2-reference/guides/)

**Understand past decisions**
→ [`3-completed/`](3-completed/) or [`2-reference/adrs/`](2-reference/adrs/)

**Review recent reports**
→ [`1-active/reports/`](1-active/reports/)

**Work on onboarding feature**
→ [`1-active/features/onboarding/`](1-active/features/onboarding/)

**Work on agent features**
→ [`1-active/features/agents/`](1-active/features/agents/)

---

## 🔍 Key Documents

### Entry Points
- [Team Start Here](0-START-HERE/TEAM_START_HERE.md) - Main team entry point
- [Recent Changes](0-START-HERE/RECENT_CHANGES.md) - Latest updates
- [Executive Summary](0-START-HERE/EXECUTIVE_SUMMARY.md) - High-level overview

### Architecture
- [Architecture Overview](2-reference/architecture/ARCHITECTURE.md)
- [MIOSA Vision Part 1](2-reference/architecture/MIOSA_VISION_PART1.md)
- [MIOSA Vision Part 2](2-reference/architecture/MIOSA_VISION_PART2.md)

### Development
- [Developer Quickstart](2-reference/development/DEVELOPER_QUICKSTART.md)
- [Frontend Guide](2-reference/development/FRONTEND.md)
- [Backend Guide](2-reference/development/BACKEND.md)

### Deployment
- [Deployment Guide](2-reference/deployment/DEPLOYMENT_GUIDE.md)
- [Cloud Infrastructure](2-reference/deployment/CLOUD-INFRASTRUCTURE.md)

### Current Work
- [Q1 2026 Sprint Plan](1-active/plans/2026-q1-sprint-plan.md)
- [Onboarding Flow](1-active/features/onboarding/2026-01-osa-build-onboarding-flow.md)
- [Agent Validation](1-active/features/agents/AGENT_VALIDATION.md)

---

## 📊 Recent Reports (Jan 2026)

All reports use date prefix: `YYYY-MM-DD-description.md`

- [Codebase Cleanup Master Report](1-active/reports/2026-01-19-codebase-cleanup-master-report.md)
- [File Cleanup Report](1-active/reports/2026-01-19-file-cleanup-report.md)
- [Security Cleanup Report](1-active/reports/2026-01-19-security-cleanup-report.md)
- [Quality Report](1-active/reports/2026-01-19-quality-report.md)
- [Master Test Report](1-active/reports/2026-01-19-master-test-report.md)
- [PR Review](1-active/reports/2026-01-19-pr-review.md)

[See all reports →](1-active/reports/)

---

## 🗂️ Organization Principles

### 0-START-HERE
**Purpose:** Essential entry points for new team members

**Retention:** Permanent (keep minimal, max 10 docs)

### 1-active
**Purpose:** Current work from last 30 days

**Retention:** 30 days → then move to `3-completed/`

**Naming:** Reports use `YYYY-MM-DD-` prefix

### 2-reference
**Purpose:** Timeless documentation that doesn't age

**Retention:** Permanent (update when systems change)

**Content:** Architecture, guides, patterns, ADRs

### 3-completed
**Purpose:** Historical completed work

**Organization:** By quarter (YYYY-QX)

**Retention:** 1 year → then archive if deprecated

### archive
**Purpose:** Deprecated, superseded, or old documentation

**Retention:** Indefinite for historical reference

---

## 📅 Maintenance Schedule

### Weekly
- Review `1-active/` for docs >30 days old
- Move aged docs to appropriate `3-completed/YYYY-QX/`

### Monthly
- Update `0-START-HERE/RECENT_CHANGES.md`
- Review new docs for proper categorization
- Ensure date prefixes on new reports

### Quarterly
- Review `3-completed/` for very old docs
- Archive deprecated docs to `archive/`
- Update this INDEX.md

### Annually
- Review entire documentation structure
- Update organization principles if needed
- Major cleanup of `archive/`

---

## 🛠️ Contributing Docs

### Adding New Documentation

1. **Choose the right location:**
   - Entry point? → `0-START-HERE/`
   - Current work? → `1-active/`
   - Timeless guide? → `2-reference/`
   - Completed work? → `3-completed/YYYY-QX/`

2. **Follow naming conventions:**
   - Reports: `YYYY-MM-DD-description.md`
   - Features: Group in subdirectories
   - ADRs: `ADR-XXX-description.md`

3. **Add README if creating new directory** (≥5 files)

4. **Link from this INDEX.md** if important

### Updating Existing Docs

1. **Update the content**
2. **Update "Last Updated" date** at bottom
3. **If in `2-reference/`**, ensure it stays timeless
4. **If in `1-active/`**, check if it's >30 days old

---

## 📖 Additional Resources

- [Reorganization Report](REORGANIZATION_BY_RELEVANCE_REPORT.md) - Full details on reorganization
- [Team Review Documentation](1-active/team-review/README.md) - PR reviews, sprint summaries
- [Claude Code Quickstart](0-START-HERE/CLAUDE_CODE_QUICKSTART.md) - Using Claude Code

---

## 📞 Help & Support

**Questions about documentation?**
- Check the README in each directory
- Review [Team Start Here](0-START-HERE/TEAM_START_HERE.md)
- Ask in team chat

**Found outdated docs?**
- Move to appropriate location
- Update or archive as needed
- Follow retention policies above

**Need to add docs?**
- Follow contributing guidelines above
- Use proper naming conventions
- Add to correct directory

---

**Documentation Structure Version:** 2.0
**Last Major Reorganization:** 2026-01-19
**Total Documents:** 209
**Organization Scheme:** Relevance-based with time retention
**Maintained by:** BusinessOS Team

**Related:**
- [CLAUDE.md](../CLAUDE.md) - Project instructions
- [README.md](../README.md) - Project README
- [TASKS.md](../TASKS.md) - Task tracking
