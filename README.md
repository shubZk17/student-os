# StudentOS

> **One platform to discover opportunities, manage applications, and build your career — running entirely on your machine.**

StudentOS is a personal opportunity discovery and career management platform for college students. It replaces fragmented internship hunting across LinkedIn, campus portals, and WhatsApp groups with a single intelligent system that matches you to real opportunities from Greenhouse, Lever, and Ashby job boards.

**Build It Track** — WeMakeDevs First Commit. Open source, local-first, reproducible.

---

## The Problem

Students waste hours every week:
- Scrolling through irrelevant job postings on multiple platforms
- Losing track of applications across email, spreadsheets, and bookmarks
- Missing deadlines because there's no central place to manage the pipeline
- Not understanding _why_ certain opportunities are a good fit

## The Solution

StudentOS combines:

| Feature | What it does |
|---|---|
| **Smart Profile** | Academic info, skills, preferences, and a profile strength score |
| **Opportunity Discovery** | Real internships and jobs pulled from Greenhouse, Lever, and Ashby boards |
| **Explainable Matching** | Multi-signal scoring with transparent "Why am I seeing this?" explanations |
| **Application Tracker** | Kanban pipeline: Saved → Applied → Assessment → Interview → Offer/Rejected |
| **Project Portfolio** | Track projects and link them to skills that boost your match scores |
| **Command Center** | Dashboard with upcoming interviews, deadlines, and high-match alerts |

---

## Build It Architecture

```
                         STUDENT
                            │
                            ▼
                    ┌───────────────┐
                    │ React 18 SPA  │
                    │ Vite + TS     │  ← localhost:3000
                    └───────┬───────┘
                            │ /api proxy
                            ▼
                    ┌───────────────┐
                    │ Go REST API   │
                    │ Gin + pgx     │  ← localhost:8080
                    └───────┬───────┘
                            │
                 ┌──────────┴──────────┐
                 │                     │
                 ▼                     ▼
        ┌─────────────────┐    ┌─────────────────┐
        │ PostgreSQL 16   │    │   LocalStack    │
        │ + pgvector      │    │ S3 + EventBridge│  ← localhost:4566
        │ localhost:5432   │    └─────────────────┘
        └─────────────────┘

        ┌──────────────────────────────────────────┐
        │              FINCH / DOCKER              │
        │       Local container environment        │
        └──────────────────────────────────────────┘

        ┌──────────────────────────────────────────┐
        │              SAM CLI                     │
        │       Local serverless development       │
        └──────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology | Purpose |
|---|---|---|
| Frontend | React 18, TypeScript, Vite 5, Tailwind CSS | Student-facing SPA |
| Backend | Go 1.23, Gin, pgx/v5 | REST API, matching engine, ingestion |
| Database | PostgreSQL 16 + pgvector | Relational data, vector extension (schema-ready) |
| Containers | **Finch** (or Docker) | Local container runtime and builds |
| AWS Emulation | **LocalStack** | S3 bucket, EventBridge scheduling |
| Serverless | **SAM CLI** | Local Lambda invocation for ingestion |
| Auth | Argon2id + JWT (HS256) | Password hashing, access/refresh tokens |

---

## Quick Start

### Prerequisites

- [Finch](https://github.com/runfinch/finch) (or Docker + Docker Compose)
- Git

### One-Command Setup

```bash
# 1. Clone
git clone https://github.com/shubZk17/student-os.git
cd student-os

# 2. Configure
cp .env.example .env

# 3. Launch everything
finch compose up -d          # or: docker compose up -d

# 4. Wait for health check
curl http://localhost:8080/health
# → {"status":"healthy","database":"connected"}
```

**That's it.** Open [http://localhost:3000](http://localhost:3000) in your browser.

### What Just Started

| Service | URL | Container |
|---|---|---|
| Frontend | http://localhost:3000 | `studentos_frontend` |
| Backend API | http://localhost:8080 | `studentos_backend` |
| Health Check | http://localhost:8080/health | — |
| PostgreSQL | localhost:5432 | `studentos_postgres` |
| LocalStack | http://localhost:4566 | `studentos_localstack` |

### Load Demo Data

```bash
# After the stack is running:
make seed-dev
```

### Pull Real Opportunities

```bash
# Run the ingestion worker (fetches from live Greenhouse/Lever/Ashby boards):
finch compose exec backend /app/worker
```

---

## Local Development (No Containers)

For faster iteration, run the backend and frontend directly:

```bash
# Terminal 1: Start PostgreSQL only
make docker-up               # or: finch compose up -d postgres

# Terminal 2: Backend
cd backend && go run cmd/server/main.go

# Terminal 3: Frontend (with hot reload)
cd frontend && npm install && npm run dev
```

Frontend is available at http://localhost:5173, with API proxy to http://localhost:8080.

---

## Explainable Matching

Every opportunity is scored against your profile using a transparent, deterministic formula:

| Signal | Weight | What it measures |
|---|---|---|
| Skill Match | 40% | Overlap between your skills and the opportunity's requirements |
| Role Fit | 20% | Fuzzy match against your target roles |
| Eligibility | 15% | Degree, graduation year, CGPA compatibility |
| Location | 10% | Match against your preferred locations |
| Project Relevance | 5% | Skills from your portfolio projects |
| Semantic | 10% | _Schema-ready, not yet implemented_ — redistributed across above signals |

Each match includes verified positive signals (e.g., "✓ Python in profile") and gap analysis (e.g., "⚠ Docker required but not in profile").

---

## Build It Technologies

### Finch — Container Runtime

[Finch](https://github.com/runfinch/finch) provides the local container environment. All services (frontend, backend, PostgreSQL, LocalStack) run as containers orchestrated by Compose. The Compose file is fully compatible with both Finch and Docker.

### LocalStack — AWS-Compatible Local Services

[LocalStack](https://localstack.cloud) emulates AWS services locally. StudentOS uses:
- **S3**: Bucket `studentos-assets` for future resume/portfolio uploads
- **EventBridge**: Scheduled rule for 6-hour ingestion cadence

No AWS account or credentials required. See `localstack/init-aws.sh`.

### SAM CLI — Serverless Development

[AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/) enables local invocation of the ingestion worker as a Lambda function:

```bash
cd sam && sam local invoke IngestFunction
```

The main Go API is **not** converted to Lambda — it remains a conventional server.

---

## Project Structure

```text
student-os/
├── .env.example              # Environment variables (no secrets)
├── .gitignore
├── Makefile                  # Developer commands (finch-up, seed-dev, etc.)
├── docker-compose.yml        # Full local stack (Finch / Docker)
├── BUILD_IT.md               # Build It track documentation
├── README.md                 # This file
├── render.yaml               # Optional: Render cloud deployment
├── amplify.yml               # Historical: AWS Amplify (never deployed)
├── AWS_MIGRATION_PLAN.md     # Historical: Ship It plan (superseded)
├── localstack/
│   └── init-aws.sh           # LocalStack S3 + EventBridge setup
├── sam/
│   ├── template.yaml         # SAM template for ingestion Lambda
│   └── samconfig.toml        # SAM local configuration
├── seeds/
│   └── dev_seed.sql          # Demo data (dev only)
├── backend/
│   ├── Dockerfile            # Multi-stage Go build
│   ├── apprunner.yaml        # Historical: App Runner (never deployed)
│   ├── cmd/
│   │   ├── server/           # API server entrypoint
│   │   └── worker/           # Ingestion worker entrypoint
│   ├── internal/
│   │   ├── auth/             # Argon2id, JWT, auth handlers
│   │   ├── users/            # Student profiles & skills
│   │   ├── jobs/             # Opportunity models & provider adapters
│   │   ├── matching/         # Explainable matching engine & tests
│   │   ├── applications/     # Kanban application tracking
│   │   ├── projects/         # Portfolio & project skills
│   │   ├── ingest/           # Job-board ingestion (CLI + API)
│   │   ├── notifications/    # In-app notifications
│   │   ├── dashboard/        # Command-center metrics
│   │   ├── database/         # Connection pooling
│   │   └── middleware/       # Auth, CORS, rate-limiting, logging
│   └── migrations/           # Embedded SQL migrations (auto-applied)
└── frontend/
    ├── Dockerfile            # Multi-stage Node + nginx
    ├── nginx.conf            # SPA routing + API proxy
    ├── vercel.json           # Optional: Vercel deployment
    └── src/
        ├── api/              # Typed REST API client
        ├── components/       # UI components & layouts
        ├── context/          # Auth context provider
        ├── pages/            # Views (Dashboard, Jobs, Kanban, Portfolio, Profile)
        └── types/            # TypeScript data models
```

---

## Current Limitations

An honest list of what is **not** true of this system today:

- **Semantic / vector matching is not implemented.** The schema declares `vector(384)` columns and the pgvector extension, but no code generates, stores, or queries embeddings. The 10% semantic weight is redistributed across deterministic signals.
- **No file uploads.** `resume_url` and `portfolio_url` are text fields for links the student types in. The S3 bucket in LocalStack is provisioned for future use but no upload endpoint exists.
- **Job ingestion is on-demand, not scheduled.** The ingestion code works and is run manually (`make worker`). EventBridge scheduling is configured in LocalStack as architectural demonstration, but the backend does not poll EventBridge.
- **Notifications are read-only.** Created by the backend and can be marked read; no push, email, or digest delivery.
- **Interview scheduling is minimal.** An application carries an `interview_date` field, but there is no calendar integration.
- **The AWS Ship It migration was attempted and did not succeed.** `AWS_MIGRATION_PLAN.md` documents the plan. No AWS resources were created.

---

## Optional Cloud Deployment

The Build It canonical environment is **local**. For public demonstration, optional deployments exist:

| Platform | URL | Status |
|---|---|---|
| Frontend (Vercel) | https://student-os-go-solo1.vercel.app | Live |
| Frontend (Render) | https://studentos-web.onrender.com | Live |
| API (Render) | https://studentos-api-0yqr.onrender.com | Live (sleeps when idle) |

These are **not** the Build It infrastructure. They are preserved for convenience.

---

## Testing

```bash
# Backend unit tests
make test

# Frontend type-check + build
cd frontend && npm run build

# Full CI (runs on every push via GitHub Actions)
# See .github/workflows/ci.yml
```

---

## License

Apache-2.0 or MIT License.
