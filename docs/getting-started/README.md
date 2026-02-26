# Getting Started with BusinessOS

BusinessOS is an open-source AI-powered business operating system. This guide walks you through setting up a local development environment and getting to your first running instance.

---

## Prerequisites

| Tool | Minimum Version | Notes |
|------|----------------|-------|
| Go | 1.24+ | [go.dev/dl](https://go.dev/dl) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org) |
| PostgreSQL | 15+ | With the `pgvector` extension |
| Redis | 7+ | Used for sessions and caching |
| Docker | Latest | Required for the terminal sandbox feature |

> If you prefer a fully containerized setup, skip to [Docker Compose option](#docker-compose-option) below.

---

## Step-by-Step Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/businessos.git
cd businessos
```

### 2. Configure Environment Variables

```bash
cp backend/.env.example backend/.env
```

Open `backend/.env` and fill in the required values. At minimum you need:

```env
DATABASE_URL=postgres://postgres:yourpassword@localhost:5432/business_os?sslmode=disable
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=yourredispassword
SECRET_KEY=                  # generate: openssl rand -base64 64
TOKEN_ENCRYPTION_KEY=        # generate: openssl rand -base64 32
```

See [docs/integrations/ENVIRONMENT.md](../integrations/ENVIRONMENT.md) for the complete variable reference.

### 3. Set Up the Database

Install PostgreSQL 15+ and enable the `pgvector` extension:

```bash
psql -U postgres -c "CREATE DATABASE business_os;"
psql -U postgres -d business_os -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

Then run the migrations in order:

```bash
for f in backend/internal/database/migrations/*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

See [docs/getting-started/DATABASE.md](DATABASE.md) for full database setup instructions including the Supabase hosted option.

### 4. Start the Backend

```bash
cd backend
go run ./cmd/server
```

The backend starts on port `8001` by default. Verify it is running:

```bash
curl http://localhost:8001/health
# Expected: {"status":"ok"}
```

### 5. Install Frontend Dependencies and Start Dev Server

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server starts at `http://localhost:5173`.

---

## Docker Compose Option

The repository includes a `docker-compose.yml` in the `backend/` directory that spins up PostgreSQL and Redis for local development. This is the fastest way to get dependencies running without a local database installation.

```bash
# From the backend/ directory
cp .env.example .env
# Edit .env and add POSTGRES_PASSWORD and REDIS_PASSWORD

docker compose up -d
```

Then run the backend and frontend as described in steps 4 and 5 above.

> The Docker Compose file exposes PostgreSQL on port `5433` (to avoid conflicts with a locally installed instance) and Redis on port `6379`.

---

## First Login and Onboarding

1. Open `http://localhost:5173` in your browser.
2. Click **Register** and create an account with your email and password.
3. After verifying your email, you are redirected to the onboarding flow.
4. The onboarding collects workspace information and guides you through connecting your first integration.
5. After onboarding completes you land on the main dashboard.

> To skip email verification during development, check the `RESEND_API_KEY` setting — if it is blank the backend still creates the account but does not send a verification email.

---

## Connecting Your First Integration (Google)

BusinessOS supports OAuth connections to external services. To connect Google:

1. Go to **Settings > Integrations** in the sidebar.
2. Click **Connect** next to Google.
3. You are redirected to Google's OAuth consent screen.
4. Grant the requested permissions.
5. You are redirected back to BusinessOS with your Google account connected.

For this to work, you must have set the following environment variables in `backend/.env`:

```env
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/auth/google/callback
```

See [docs/integrations/README.md](../integrations/README.md) for setup instructions for all 9 supported integrations.

---

## Verifying Your Setup

Run these checks to confirm everything is working:

```bash
# Backend compiles cleanly
cd backend && go build -o /dev/null ./cmd/server

# All tests pass
cd backend && go test ./...

# Frontend builds without errors
cd frontend && npm run build && npm run check
```

---

## What Next

- [Architecture Overview](../architecture/README.md) — understand how the system is structured
- [Module Overview](../modules/README.md) — explore the 15 built-in modules
- [API Reference](../api/README.md) — start building with the REST API
- [Deployment Guide](../deployment/README.md) — deploy to production
