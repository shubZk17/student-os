# StudentOS

> **One place to understand your college life, discover opportunities, manage your projects, and decide what to do next.**

StudentOS is a personal operating system and opportunity discovery platform tailored for college and university students. It replaces the fragmented noise of college ERP portals, WhatsApp broadcast groups, scattered LinkedIn postings, and unorganized email threads with an intelligent, personalized, and explainable opportunity and execution layer.

---

## Key Features

* **Personalized Opportunity Discovery:** Multi-signal matching engine ranks internships and jobs according to your profile, graduation year, target roles, location preferences, and verified skills.
* **Explainable Matching ("Why am I seeing this?"):** Every match returns positive reasons (e.g., `✓ Python in profile`, `✓ Preferred location: Bangalore`) and actionable gap analysis (e.g., `⚠ Docker required but not in profile`).
* **Application Tracker:** Kanban-style application management (`Saved` ➔ `Applied` ➔ `Assessment` ➔ `Interview` ➔ `Offer` / `Rejected`), with notes and an interview date per application.
* **Projects & Portfolio Showcase:** Maintain portfolio projects linked directly to verified skills that dynamically boost opportunity recommendation scores.
* **Personalized Dashboard:** A single daily command center highlighting upcoming interview schedules, urgent deadlines, new high matches, and profile completion strength.
* **Structured Job Ingestion:** Pulls student roles (internships, new-grad, early-career) from public Greenhouse, Lever and Ashby job boards, tags skills, and closes postings that disappear. Runs on demand — see [Current limitations](#current-limitations).
* **Cloud Deployable:** One-click Render Blueprint (`render.yaml`), with zero LLM or paid search dependencies — matching is deterministic and runs in-process.

---

## System Architecture

```
                       [ React 18 SPA (Vite + TypeScript + Tailwind) ]
                                            │
                                            │ REST APIs / JWT Auth
                                            ▼
                           [ Go Backend Server (Gin, pgx) ]
                ┌───────────────────────────┼───────────────────────────┐
                ▼                           ▼                           ▼
        [ Auth & Profile ]         [ Matching Engine ]         [ Ingestion Worker ]
                │                           │                           │
                └───────────────────────────┼───────────────────────────┘
                                            │
                                            ▼
                              [ PostgreSQL 16 ]
```

---

## Tech Stack

* **Backend:** Go 1.22, Gin HTTP router, `pgx/v5` PostgreSQL driver, Argon2id, JWT.
* **Frontend:** React 18, TypeScript, Vite, Tailwind CSS, Lucide React, React Router.
* **Database:** PostgreSQL 16. The schema declares the `vector` extension and two
  `embedding` columns, but nothing writes to them yet — see [Current limitations](#current-limitations).
* **Infrastructure:** Render (`render.yaml`): managed PostgreSQL, Docker web service, static site.

---

## Project Structure

```text
Student OS/
├── .env.example              # Environment variables template
├── .gitignore                # Git ignore rules
├── Makefile                  # Developer ergonomics
├── docker-compose.yml        # Local PostgreSQL + pgvector
├── README.md                 # Product documentation
├── docs/
│   ├── FEATURES.md           # What the product does today
│   ├── PRD.md                # Complete Product Requirements Document
│   └── ARCHITECTURE.md       # Architectural specifications
├── render.yaml               # Render Blueprint (production deployment)
├── seeds/
│   └── dev_seed.sql          # Demo data, local development only
├── backend/
│   ├── Dockerfile            # API server + worker image
│   ├── migrations/           # Embedded SQL migrations, applied on server start
│   ├── cmd/
│   │   ├── server/           # API server entrypoint
│   │   └── worker/           # Scheduled ingestion worker
│   ├── internal/
│   │   ├── auth/             # Argon2id, JWT, auth handlers
│   │   ├── users/            # Student profiles & skills
│   │   ├── jobs/             # Job models & provider adapters
│   │   ├── matching/         # Explainable matching engine & tests
│   │   ├── applications/     # Kanban application tracking
│   │   ├── projects/         # Portfolio & project skills
│   │   ├── ingest/           # Job-board ingestion, callable from CLI or API
│   │   ├── notifications/    # In-app notifications
│   │   ├── dashboard/        # Command-center metrics
│   │   ├── database/         # Connection pooling
│   │   └── middleware/       # Auth, CORS, logging
│   └── go.mod
└── frontend/
    ├── src/
    │   ├── api/              # Typed REST API clients
    │   ├── components/       # UI components & layouts
    │   ├── context/          # Authentication context
    │   ├── pages/            # Views (Dashboard, Jobs, Kanban, Portfolio, Profile)
    │   ├── types/            # TypeScript data models
    │   ├── App.tsx           # Router & navigation
    │   └── main.tsx
    ├── package.json
    ├── vite.config.ts
    └── tailwind.config.js
```

---

## Quickstart (Local Development)

### 1. Prerequisites
* Go 1.22+
* Node.js 20+ / 22+ & npm
* Docker & Docker Compose (or local PostgreSQL)

### 2. Configure Environment
```bash
cp .env.example .env
```

### 3. Start Database
```bash
docker compose up -d
```

### 4. Run Backend
```bash
cd backend
go run ./cmd/server
```
The Go API server will start on `http://localhost:8080` and apply any pending migrations.

Optional: load demo opportunities with `make seed-dev`, and pull real postings with `go run ./cmd/worker`.

### 5. Run Frontend
```bash
cd frontend
npm install
npm run dev
```
The React SPA will be available at `http://localhost:5173`.

---

## Explainable Matching Formula

$$S = (0.40 \cdot S_{\text{skill}}) + (0.20 \cdot S_{\text{role}}) + (0.15 \cdot S_{\text{elig}}) + (0.10 \cdot S_{\text{loc}}) + (0.05 \cdot S_{\text{proj}}) + (0.10 \cdot S_{\text{sem}})$$

1. **Stage 1 (Deterministic Filtering):** Filter out expired postings, conflicting degree requirements, and incompatible graduation batches.
2. **Stage 2 (Multi-Signal Scoring):** Compute weighted overlap of candidate skills, target roles, locations, and portfolio relevance.
3. **Stage 3 (Explainability):** Generate verified bullet points explaining positive signals and highlight missing requirements.

> **On the semantic term.** `S_sem` is wired into the formula but no embeddings are
> generated today, so it is always supplied as `0`. The engine then substitutes the mean
> of the skill, role and location scores, which means that 10% currently acts as extra
> weight on those three deterministic signals rather than as semantic similarity.
> The weights are left as-is so the behaviour is not silently changed.

---

## Live environments

| What | URL |
| --- | --- |
| Frontend (Vercel, primary) | https://student-os-go-solo1.vercel.app |
| Frontend (Render static) | https://studentos-web.onrender.com |
| API (Render) | https://studentos-api-0yqr.onrender.com |

Both frontends and the API redeploy automatically on every push to `main`.
`.github/workflows/ci.yml` runs `go vet`/`go test`/`go build` and the frontend
build on every push and pull request.

## Deployment (Render)

`render.yaml` defines managed Postgres (private network only), the API (Docker)
and the static frontend.

1. Push the repo to GitHub, then in Render choose **New → Blueprint** and select it.
2. When prompted, set:
   * `ALLOWED_ORIGINS` on `studentos-api`: the frontend URLs, comma-separated
   * `VITE_API_URL` on `studentos-web`: the API URL
3. Deploy. The API applies migrations on startup; a failed migration fails the health check,
   so the previous version keeps serving. `JWT_SECRET` is generated by Render.

The ingestion cron (`backend/cmd/worker`) is **not** deployed: Render has no free cron
tier. Restore the `type: cron` service in `render.yaml` on a paid plan, or run it by hand
with `make worker`, to populate opportunities.

## Deployment (Vercel)

The Vercel project builds only `frontend/` (Root Directory = `frontend`). `frontend/vercel.json`
adds the SPA rewrite that react-router needs and the security headers. Set `VITE_API_URL`
in the project's environment variables to the API URL.

Demo data (`seeds/`) is never deployed. Job boards can be changed with the
`GREENHOUSE_BOARDS`, `LEVER_COMPANIES` and `ASHBY_BOARDS` env vars on the worker.

---

## Current limitations

An honest list of what is **not** true of the deployed system today.

* **Job ingestion is not scheduled in production.** The ingestion code
  (`backend/cmd/worker`, `backend/internal/ingest`) works and is run on demand with
  `make worker`. It is not on a timer, because Render has no free cron tier, so the
  opportunity set does not refresh by itself. The 6-hour cadence is the intended
  schedule, not the running one.
* **Semantic / vector matching is not implemented.** The schema declares the `vector`
  extension and `embedding vector(384)` columns, but no code generates, stores or
  queries an embedding. See the note under the matching formula.
* **No file uploads.** `resume_url` and `portfolio_url` are text fields for links the
  student pastes in; nothing is stored server-side, and there is no object storage.
* **Free-tier hosting caveats.** The API sleeps when idle, so the first request after a
  quiet period is slow. The Render free PostgreSQL instance expires 30 days after
  creation (created 2026-09-20).
* **Notifications are read-only in practice.** They are created by the backend and can
  be marked read; there is no push, email or digest delivery.
* **Interview scheduling is minimal.** An application can carry an `interview_date`, but
  there is no calendar integration or reminder delivery.
* **AWS migration is planned, not done.** `AWS_MIGRATION_PLAN.md` on the
  `feat/aws-ship-it` branch documents the target AWS architecture and the code prepared
  for it. No AWS resources exist.

---

## License

Apache-2.0 or MIT License.
