# Deployment Guide

This guide covers deploying BusinessOS to production using Docker, Docker Compose for local environments, and Google Cloud Run for cloud deployments.

---

## Prerequisites

- Docker 24+ and Docker Compose v2
- A production PostgreSQL 15+ database with the `pgvector` extension (or Supabase)
- A production Redis 7+ instance
- All required environment variables configured (see [docs/integrations/ENVIRONMENT.md](../integrations/ENVIRONMENT.md))

---

## Docker Build

The `backend/` directory contains a `Dockerfile` for building the Go backend as a container image.

### Build the Backend Image

```bash
cd backend
docker build -t businessos-backend:latest .
```

### Run the Backend Container

```bash
docker run -d \
  --name businessos-backend \
  -p 8001:8001 \
  --env-file .env \
  businessos-backend:latest
```

Verify the container is running and healthy:

```bash
curl http://localhost:8001/health
# Expected: {"status":"ok"}
```

---

## Docker Compose (Local Development)

The `backend/docker-compose.yml` file starts PostgreSQL and Redis for local development. The backend itself runs as a native Go process outside of Docker to support hot-reloading.

### Start Dependencies

```bash
cd backend

# Copy and configure environment variables
cp .env.example .env
# Edit .env: add POSTGRES_PASSWORD and REDIS_PASSWORD at minimum

docker compose up -d
```

This starts:
- PostgreSQL 16 on port `5433` (container name: `businessos-postgres`)
- Redis 7 on port `6379` (container name: `businessos-redis`)

### Stop Dependencies

```bash
docker compose down
```

To also remove the stored data volumes:

```bash
docker compose down -v
```

### Verify Services Are Running

```bash
# PostgreSQL
docker exec businessos-postgres pg_isready -U postgres
# Expected: /var/run/postgresql:5432 - accepting connections

# Redis
docker exec businessos-redis redis-cli -a "$REDIS_PASSWORD" ping
# Expected: PONG
```

---

## Building the Workspace Docker Image

The terminal sandbox and BUILD mode code generation use an isolated Docker container called the **workspace image**. This image must be built separately before the terminal feature will work.

```bash
cd backend/docker/workspace
docker build -t businessos-workspace:latest .
```

Verify the image is available:

```bash
docker images businessos-workspace:latest
```

The workspace image is based on Alpine Linux and includes common development tools (Node.js, Python, Git, curl). It is approximately 270 MB.

> If the terminal shows "image not found" errors, rebuild the workspace image using the command above.

---

## Cloud Run Deployment (GCP)

BusinessOS can be deployed to [Google Cloud Run](https://cloud.google.com/run) as a stateless container. Cloud Run handles scaling, HTTPS termination, and regional deployment automatically.

### Prerequisites

- Google Cloud SDK (`gcloud`) installed and authenticated
- A GCP project with Cloud Run and Container Registry (or Artifact Registry) enabled
- A Cloud SQL instance (PostgreSQL 15+ with pgvector) or Supabase connection string
- A Redis instance (Memorystore or external provider)

### 1. Build and Push the Container Image

```bash
# Set your GCP project ID
export PROJECT_ID=your-gcp-project-id
export REGION=us-central1
export IMAGE=gcr.io/$PROJECT_ID/businessos-backend:latest

# Build and push
docker build -t $IMAGE backend/
docker push $IMAGE
```

Or use Cloud Build to build remotely:

```bash
cd backend
gcloud builds submit --tag $IMAGE
```

### 2. Deploy to Cloud Run

```bash
gcloud run deploy businessos-backend \
  --image $IMAGE \
  --region $REGION \
  --platform managed \
  --allow-unauthenticated \
  --port 8001 \
  --memory 512Mi \
  --cpu 1 \
  --set-env-vars "ENVIRONMENT=production" \
  --set-env-vars "SERVER_PORT=8001" \
  --set-secrets "DATABASE_URL=businessos-database-url:latest" \
  --set-secrets "REDIS_URL=businessos-redis-url:latest" \
  --set-secrets "SECRET_KEY=businessos-secret-key:latest" \
  --set-secrets "TOKEN_ENCRYPTION_KEY=businessos-token-key:latest" \
  --set-secrets "REDIS_PASSWORD=businessos-redis-password:latest"
```

> Store sensitive environment variables in [Google Secret Manager](https://cloud.google.com/secret-manager) and reference them with `--set-secrets`. Never pass secrets directly on the command line.

### 3. Configure the Frontend

After deployment, Cloud Run assigns a service URL like `https://businessos-backend-xyz.run.app`. Update the frontend's API base URL to point to this address.

In `frontend/.env.production`:

```env
PUBLIC_API_BASE_URL=https://businessos-backend-xyz.run.app
```

Rebuild and redeploy the frontend:

```bash
cd frontend
npm run build
```

Deploy the frontend to your hosting provider (Cloud Run, Vercel, Netlify, or a static CDN).

---

## Environment Configuration for Production

When deploying to production, update these variables from their development defaults:

```env
# Set to production
ENVIRONMENT=production

# Use your production domain
APP_URL=https://app.yourdomain.com
ALLOWED_ORIGINS=https://app.yourdomain.com

# OAuth redirect URIs must match your production domain
GOOGLE_REDIRECT_URI=https://api.yourdomain.com/api/auth/google/callback
MICROSOFT_REDIRECT_URI=https://api.yourdomain.com/api/auth/microsoft/callback
# ... repeat for all other integrations

# Enable Redis TLS for production Redis instances
REDIS_TLS_ENABLED=true

# Use a strong, randomly generated secret key
SECRET_KEY=<64-byte random value>
TOKEN_ENCRYPTION_KEY=<32-byte random value>
```

---

## Health Check

Cloud Run and other container orchestrators use the health check endpoint to determine if the service is ready to receive traffic:

```
GET /health
```

**Response when healthy:**
```json
{"status":"ok"}
```

The health check does not require authentication. Configure your load balancer or Cloud Run health check to poll this endpoint.

**Cloud Run health check configuration:**

Cloud Run automatically uses the root path `/` for liveness checks. To use `/health` instead, configure a startup probe:

```yaml
# In the Cloud Run service YAML
startupProbe:
  httpGet:
    path: /health
    port: 8001
  initialDelaySeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

---

## Running Migrations in Production

Migrations must be applied before starting the backend. In a CI/CD pipeline, run migrations as a separate step before deploying the new container image:

```bash
# Using a temporary Cloud Run job or Cloud Build step:
for f in $(ls backend/internal/database/migrations/*.sql | sort); do
  psql "$DATABASE_URL" -f "$f"
done
```

For Supabase, use the Supabase CLI to manage migrations:

```bash
supabase db push
```

---

## Troubleshooting

### Container Exits Immediately

Check the container logs for startup errors:

```bash
docker logs businessos-backend
```

Common causes:
- Missing required environment variables (`DATABASE_URL`, `SECRET_KEY`, `TOKEN_ENCRYPTION_KEY`)
- Database is not reachable from the container network
- Port conflict — another process is already using port `8001`

### "relation does not exist" Database Errors

Migrations have not been applied. Run the migration script against the production database before deploying.

### OAuth Redirects Fail in Production

The redirect URI registered in each OAuth provider's developer console must exactly match the `*_REDIRECT_URI` environment variables. Ensure both are set to the production domain.

### Workspace Docker Image Not Found

The terminal sandbox requires the `businessos-workspace:latest` image to be available on the Docker host. Build it on the production host:

```bash
cd backend/docker/workspace
docker build -t businessos-workspace:latest .
```

For Cloud Run, the sandbox feature requires the container to have access to a Docker daemon, which Cloud Run does not provide. Use an alternative compute option (Cloud Run on GKE, Compute Engine, or a VM-based deployment) if terminal functionality is required in production.

### Redis Connection Refused

Verify the `REDIS_URL` points to the correct host and port, and that `REDIS_PASSWORD` matches the password set on the Redis server. For production Redis instances, ensure `REDIS_TLS_ENABLED=true` if the instance requires TLS.
