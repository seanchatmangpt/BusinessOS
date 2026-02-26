# Database Setup

BusinessOS uses PostgreSQL 15+ as its primary database and Redis 7+ for sessions, caching, and real-time pub/sub. This guide covers both local installation and the hosted Supabase option.

---

## PostgreSQL 15+

### macOS

```bash
brew install postgresql@15
brew services start postgresql@15
```

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y postgresql-15 postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### Windows

Download and run the installer from [postgresql.org/download/windows](https://www.postgresql.org/download/windows/).

---

## Create the Database

```bash
# Connect as the postgres superuser
psql -U postgres

# Inside psql:
CREATE DATABASE business_os;
\c business_os
```

---

## Enable the pgvector Extension

BusinessOS uses `pgvector` to store and search vector embeddings for its AI memory and RAG (retrieval-augmented generation) features.

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

Verify it is installed:

```sql
SELECT * FROM pg_extension WHERE extname = 'vector';
```

> If you see `ERROR: could not open extension control file`, you need to install the `pgvector` package first.

### Installing pgvector

**macOS (Homebrew):**
```bash
brew install pgvector
```

**Ubuntu / Debian (from source):**
```bash
sudo apt install postgresql-server-dev-15
git clone --branch v0.7.0 https://github.com/pgvector/pgvector.git
cd pgvector
make
sudo make install
```

---

## Run Migrations

All migrations live in `backend/internal/database/migrations/` and are named with a numeric prefix (`001_`, `002_`, etc.). They must be applied in order.

**Apply all migrations in sequence:**

```bash
DATABASE_URL="postgres://postgres:yourpassword@localhost:5432/business_os?sslmode=disable"

for f in $(ls backend/internal/database/migrations/*.sql | sort); do
  echo "Applying $f..."
  psql "$DATABASE_URL" -f "$f"
done
```

**Apply a single migration:**

```bash
psql "$DATABASE_URL" -f backend/internal/database/migrations/001_initial_schema.sql
```

> Migrations are append-only. Never modify an existing migration file — add a new one instead.

---

## Redis 7+

Redis is used for:
- Session storage (JWT session data)
- Query result caching
- Real-time pub/sub for SSE streaming

### macOS

```bash
brew install redis
brew services start redis
```

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

### Set a Redis Password

BusinessOS requires Redis to run with a password. Edit `/etc/redis/redis.conf`:

```
requirepass yourredispassword
```

Or pass it as a command-line argument:

```bash
redis-server --requirepass yourredispassword
```

The Docker Compose setup in `backend/docker-compose.yml` handles this automatically using the `REDIS_PASSWORD` environment variable.

---

## Environment Variables

Set these in `backend/.env`:

```env
# PostgreSQL
DATABASE_URL=postgres://postgres:yourpassword@localhost:5432/business_os?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=yourredispassword
REDIS_KEY_HMAC_SECRET=   # generate: openssl rand -base64 32
REDIS_TLS_ENABLED=false  # set to true in production
```

---

## Supabase (Hosted PostgreSQL)

[Supabase](https://supabase.com) provides hosted PostgreSQL with `pgvector` pre-installed and is a convenient alternative to running a local database.

### Setup

1. Create a project at [supabase.com](https://supabase.com).
2. In your project dashboard go to **Settings > Database**.
3. Under **Connection string**, select **Connection pooling** with **Transaction mode**.
4. Copy the connection string and replace `[YOUR-PASSWORD]` with your database password.

```env
DATABASE_URL=postgres://postgres.PROJECT_ID:PASSWORD@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true
```

5. For migration scripts that need a direct connection (bypassing the pooler), also set:

```env
SUPABASE_DIRECT_HOST=db.YOUR-PROJECT-ID.supabase.co:5432
```

### Enable pgvector on Supabase

In the Supabase SQL editor, run:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

pgvector is available on all Supabase projects.

---

## Troubleshooting

### "could not connect to server: Connection refused"

The PostgreSQL service is not running.

```bash
# macOS
brew services restart postgresql@15

# Linux
sudo systemctl restart postgresql
```

### "role does not exist"

The `postgres` role is missing. Create it:

```bash
createuser -s postgres
```

### "extension vector does not exist"

pgvector is not installed. See [Installing pgvector](#installing-pgvector) above.

### "FATAL: password authentication failed"

The password in `DATABASE_URL` does not match the database user's password. Reset it:

```bash
psql -U postgres -c "ALTER USER postgres PASSWORD 'yournewpassword';"
```

### Redis "WRONGPASS invalid username-password pair"

The `REDIS_PASSWORD` in your `.env` does not match the Redis server configuration. Verify:

```bash
redis-cli -a yourpassword ping
# Expected: PONG
```

### Migrations fail partway through

If a migration fails, fix the issue and re-run only the failed migration. Already-applied migrations are idempotent for schema objects but not always for data inserts, so check the specific error message before re-running.
