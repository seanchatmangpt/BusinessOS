# Welcome to BusinessOS

**Your complete guide to the project**

**Last Updated:** 2026-01-19 | **Total Team Members:** 5+ | **Active Phase:** Q1 2026

---

## 🎯 Choose Your Path

### I'm New to the Project
**Time needed:** 30 minutes

1. Read this page (5 min)
2. Read [Project Overview](#project-overview) (10 min)
3. Read [TEAM_START_HERE.md](../TEAM_START_HERE.md) (10 min)
4. Explore [Key Documentation](#key-documentation) (5 min)

**Then:** Pick your role below

---

### I'm a Frontend Developer
**Time needed:** 20 minutes

1. Review [Frontend Development Guide](#frontend-development)
2. Check [Component Patterns](#component-patterns)
3. Read [Onboarding Flow](#onboarding-system) (what you'll be building)
4. Clone and explore [frontend/](../../frontend/) directory

**Key files:**
- `/frontend/src/routes/` - All page components
- `/frontend/src/components/` - Reusable components
- `/frontend/src/lib/stores/` - State management

---

### I'm a Backend Developer (Go)
**Time needed:** 20 minutes

1. Review [Backend Development Guide](#backend-development)
2. Check [API Architecture](#backend-development)
3. Read [Onboarding Flow](#onboarding-system) (what backend enables)
4. Clone and explore [desktop/backend-go/](../../desktop/backend-go/) directory

**Key files:**
- `/desktop/backend-go/internal/handler/` - API endpoints
- `/desktop/backend-go/internal/service/` - Business logic
- `/desktop/backend-go/internal/repository/` - Database access

---

### I'm a Designer or PM
**Time needed:** 15 minutes

1. Read [Product Overview](#product-overview)
2. Review [Current Phase](#current-phase)
3. Check [Onboarding System](#onboarding-system)
4. Read [DELIVERY_CHECKLIST.md](../DELIVERY_CHECKLIST.md)

---

### I'm Deploying or Managing Infrastructure
**Time needed:** 15 minutes

1. Read [Deployment Guide](#deployment)
2. Review [Production Readiness](#production-readiness)
3. Check [Security Checklist](#security)
4. Read deployment docs in [deployment/](../deployment/)

---

## 📍 Project Overview

**BusinessOS** is an AI-powered business operating system that helps teams collaborate, manage projects, and build AI agents through a conversational interface.

### What We're Building

A **meta-platform generator** that lets users:
- Build custom operating systems through conversation
- Collaborate with AI agents in real-time
- Manage projects, tasks, and teams
- Integrate with external tools and services
- Scale from startup to enterprise

### Core Features (In Active Development)

| Feature | Status | Documentation |
|---------|--------|---------------|
| **Onboarding System** | Complete | [OSA_BUILD_ONBOARDING_FLOW.md](../OSA_BUILD_ONBOARDING_FLOW.md) |
| **Google OAuth** | Complete | [OSA_BUILD_ONBOARDING_FLOW.md](../OSA_BUILD_ONBOARDING_FLOW.md) |
| **Agent Architecture** | Complete | [OSA_BUILD_DEEP_IMPLEMENTATION_PLAN.md](../OSA_BUILD_DEEP_IMPLEMENTATION_PLAN.md) |
| **Voice System** | Complete | [features/voice-system/](../features/voice-system/) |
| **Social Features** | In Progress | [OSA_BUILD_SOCIAL_SYSTEM.md](../OSA_BUILD_SOCIAL_SYSTEM.md) |
| **Testing Suite** | Comprehensive | [OSA_BUILD_TESTING_PLAN.md](../OSA_BUILD_TESTING_PLAN.md) |

### Tech Stack

| Layer | Technology | Key Info |
|-------|-----------|----------|
| **Frontend** | SvelteKit + Svelte 5 + TypeScript | `/frontend/` |
| **Backend** | Go 1.24.1 + Chi router | `/desktop/backend-go/` |
| **Database** | PostgreSQL + Redis | Cloud-hosted |
| **Deployment** | Docker + GCP Cloud Run | Containerized |
| **Integration** | 20+ MCP tool servers | Extensible |

---

## 🚀 Current Phase (Q1 2026)

### What's Happening Now

**Phase 4: Complete Beta Platform**

Focus:
- Onboarding system (OAuth + Gmail + voice)
- Agent architecture refinement
- Social features (teams, sharing)
- Testing and QA
- Production readiness

### Key Milestones

- ✅ Backend foundation (API, database)
- ✅ Onboarding system (OAuth, voice)
- ✅ Agent architecture (validation, memory)
- 🔄 Social features (in progress)
- ⏳ Production launch (planned)

---

## 📚 Key Documentation

### Essential Reading

1. **[INDEX.md](../INDEX.md)** - Master index of all 1,761+ docs
2. **[TEAM_START_HERE.md](../TEAM_START_HERE.md)** - Team orientation
3. **[OSA_BUILD_ONBOARDING_FLOW.md](../OSA_BUILD_ONBOARDING_FLOW.md)** - Current major feature
4. **[RECENT_CHANGES.md](../RECENT_CHANGES.md)** - What's changed lately

### By Role

#### Frontend Developers
- [FORM_COMPONENTS_USAGE_GUIDE.md](../FORM_COMPONENTS_USAGE_GUIDE.md) - Component system
- [FORM_PATTERNS_INDEX.md](../FORM_PATTERNS_INDEX.md) - Common patterns
- [development/frontend/](../development/frontend/) - All frontend guides
- [features/onboarding/](../features/onboarding/) - Feature you're building

#### Backend Developers (Go)
- [development/backend/](../development/backend/) - Backend guides
- [BUSINESSOS_TABLES_RECOMMENDATIONS.md](../BUSINESSOS_TABLES_RECOMMENDATIONS.md) - Database schema
- [OSA_BUILD_DEEP_IMPLEMENTATION_PLAN.md](../OSA_BUILD_DEEP_IMPLEMENTATION_PLAN.md) - Architecture
- [features/onboarding/](../features/onboarding/) - Feature documentation

#### DevOps/Infrastructure
- [deployment/](../deployment/) - Deployment docs
- [PRODUCTION_READINESS_ASSESSMENT.md](../PRODUCTION_READINESS_ASSESSMENT.md) - Prod checklist
- [SECURITY_CLEANUP_REPORT.md](../SECURITY_CLEANUP_REPORT.md) - Security status
- [development/backend/](../development/backend/) - Backend infrastructure

#### PMs/Stakeholders
- [EXECUTIVE_SUMMARY.md](../EXECUTIVE_SUMMARY.md) - High-level overview
- [DELIVERY_CHECKLIST.md](../DELIVERY_CHECKLIST.md) - What's shipped
- [OSA_BUILD_USER_FLOW_GUIDE.md](../OSA_BUILD_USER_FLOW_GUIDE.md) - User workflows

---

## 🎓 Quick Guides

### Frontend Development

**Setup (first time):**
```bash
cd frontend
npm install
npm run dev
```

**Key directories:**
- `src/routes/` - Page components (SvelteKit)
- `src/components/` - Reusable components
- `src/lib/stores/` - State management
- `src/lib/api/` - API client

**Key patterns:**
- Use SvelteKit's `+page.svelte` for pages
- Use stores for global state
- Use form actions for mutations
- TypeScript strict mode required

---

### Backend Development (Go)

**Setup (first time):**
```bash
cd desktop/backend-go
go install
./scripts/run.sh  # or manual setup
```

**Key directories:**
- `internal/handler/` - HTTP handlers
- `internal/service/` - Business logic
- `internal/repository/` - Database queries
- `scripts/` - Utilities and migrations

**Key patterns:**
- Handler → Service → Repository (layered)
- Context propagation in all functions
- Error handling (no panics)
- Structured logging with slog

---

### Testing

**Frontend tests:**
```bash
cd frontend
npm run test
```

**Backend tests:**
```bash
cd desktop/backend-go
go test ./...
```

**Integration tests:** See [TESTING_OSA_ONBOARDING.md](../TESTING_OSA_ONBOARDING.md)

---

### Deployment

**Local testing:**
```bash
# Frontend
npm run preview

# Backend
./scripts/run.sh
```

**Staging deployment:** See [deployment/](../deployment/)

**Production:** Review [PRODUCTION_READINESS_ASSESSMENT.md](../PRODUCTION_READINESS_ASSESSMENT.md) first

---

## 🏗️ Component Patterns

### Forms

**Pattern:** SvelteKit form actions + Zod validation + TypeScript

```svelte
<form method="POST" action="?/submit">
  <input name="email" type="email" required />
  <button type="submit">Submit</button>
</form>
```

**Details:** [FORM_PATTERNS_INDEX.md](../FORM_PATTERNS_INDEX.md)

### API Calls

**Pattern:** Go handler → TypeScript fetch → Frontend store

**Backend (Go):**
```go
// Handler → Service → Repository
func (h *Handler) CreateUser(c *gin.Context) {
  // Validation
  // Call service
  // Return JSON
}
```

**Frontend (TypeScript):**
```ts
const response = await fetch('/api/users', {
  method: 'POST',
  body: JSON.stringify(data)
})
```

---

## 🔒 Security Checklist

Before merging code:

- [ ] No credentials in code
- [ ] Input validation on backend
- [ ] Auth token validation
- [ ] CORS headers correct
- [ ] SQL injection prevention
- [ ] Rate limiting implemented

See [SECURITY_CLEANUP_REPORT.md](../SECURITY_CLEANUP_REPORT.md) for details

---

## 🚀 Onboarding System

The feature everyone's working on right now.

### What It Does

1. **Welcome flow** - New user signup
2. **OAuth** - Google sign-in
3. **Setup** - Email, profile, preferences
4. **Agent intro** - Meet the AI
5. **First task** - Interactive tutorial

### Documentation

- **User flow:** [OSA_BUILD_USER_FLOW_GUIDE.md](../OSA_BUILD_USER_FLOW_GUIDE.md) (5 min read)
- **Full spec:** [OSA_BUILD_ONBOARDING_FLOW.md](../OSA_BUILD_ONBOARDING_FLOW.md) (30 min read)
- **Testing:** [TESTING_OSA_ONBOARDING.md](../TESTING_OSA_ONBOARDING.md) (test scenarios)
- **Implementation:** [features/onboarding/](../features/onboarding/) (detailed guides)

---

## 👥 Team

| Person | Role | Notes |
|--------|------|-------|
| Roberto | Architecture, Frontend | Lead |
| Pedro | Backend (Go) | Main backend developer |
| Nick | DevOps, Terminal | Deployment & infrastructure |
| Abdul | E2B Integration | Sandbox/execution |
| Nejd/Javaris | Frontend | Component development |

---

## 📋 First Week Tasks

### Day 1
- [ ] Clone repo and setup environment
- [ ] Read this README
- [ ] Read TEAM_START_HERE.md
- [ ] Run `npm install` (frontend) or `go mod download` (backend)

### Day 2-3
- [ ] Follow your role's quick guide above
- [ ] Read related feature docs
- [ ] Make first code change

### Day 4-5
- [ ] Create first PR
- [ ] Get code review
- [ ] Deploy to staging

### Week 2+
- [ ] Own a component/feature
- [ ] Contribute to team deliverables

---

## ❓ FAQ

**Q: Where's the feature I need to build?**
A: Check [features/](../features/) or search in [INDEX.md](../INDEX.md)

**Q: How do I deploy?**
A: See [deployment/](../deployment/) - but ask Nick first for GCP access

**Q: I found a bug, where do I report it?**
A: Create Linear issue and ping the relevant person

**Q: Can I refactor this code?**
A: Sure! But discuss first in #engineering or file an issue

**Q: How do I understand the architecture?**
A: Start with [architecture/](../architecture/) and [adrs/](../adrs/)

**Q: Where's the database schema?**
A: [BUSINESSOS_TABLES_RECOMMENDATIONS.md](../BUSINESSOS_TABLES_RECOMMENDATIONS.md)

---

## 🔗 Quick Links

| Purpose | Link |
|---------|------|
| Master index | [INDEX.md](../INDEX.md) |
| Team guide | [TEAM_START_HERE.md](../TEAM_START_HERE.md) |
| Recent work | [RECENT_CHANGES.md](../RECENT_CHANGES.md) |
| All docs | [DOCUMENTATION_INVENTORY.md](../DOCUMENTATION_INVENTORY.md) |
| Frontend docs | [development/frontend/](../development/frontend/) |
| Backend docs | [development/backend/](../development/backend/) |
| Deployment | [deployment/](../deployment/) |
| Architecture | [architecture/](../architecture/) |
| Features | [features/](../features/) |
| Onboarding feature | [OSA_BUILD_ONBOARDING_FLOW.md](../OSA_BUILD_ONBOARDING_FLOW.md) |

---

## 📞 Getting Help

1. **For project questions:** Check [INDEX.md](../INDEX.md)
2. **For technical issues:** Check [RECENT_CHANGES.md](../RECENT_CHANGES.md)
3. **For debugging:** Check [development/](../development/) or archive/troubleshooting/
4. **For team questions:** Ask in Linear or team-review/
5. **For ops questions:** Ask Nick (DevOps)

---

## ✅ Next Steps

1. **Read:** This page + TEAM_START_HERE.md (30 min total)
2. **Setup:** Clone repo and install dependencies (15 min)
3. **Explore:** Open the codebase, run dev server (15 min)
4. **Contribute:** Pick a task from your team lead (asap)

**Questions?** Check TEAM_START_HERE.md or ask in #engineering

---

**Welcome aboard!**

**Last Updated:** 2026-01-19
**Maintained By:** Roberto
