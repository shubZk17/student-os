# StudentOS

> **One platform to discover opportunities, track applications, and build your career.**

StudentOS is a personal opportunity discovery and career management platform for college
students. It replaces fragmented internship hunting across LinkedIn, campus portals, and
WhatsApp groups with a single system that pulls real openings from company job boards,
scores them against your profile, and tells you *why* each one fits.

---

## The Problem

Students waste hours every week:

- Scrolling through irrelevant job postings on multiple platforms
- Losing track of applications across email, spreadsheets, and bookmarks
- Missing deadlines because there's no central place to manage the pipeline
- Not understanding _why_ certain opportunities are a good fit

## What StudentOS Does

| Feature | What it does |
|---|---|
| **Smart Profile** | Academic record, skills, target roles, location and work preferences, with a profile strength score that shows what's still missing |
| **Opportunity Discovery** | Real internships and jobs pulled automatically from Greenhouse, Lever, and Ashby company boards — deduplicated, normalized, and filtered to student-appropriate roles |
| **Explainable Matching** | Every opening is scored against your profile, with a transparent breakdown of which signals matched and which requirements you don't meet yet |
| **Application Tracker** | Kanban pipeline: Saved → Applied → Assessment → Interview → Offer / Rejected, with interview dates, follow-up reminders, and per-application notes |
| **Project Portfolio** | Record your projects and link them to skills — project-validated skills feed directly back into your match scores |
| **Command Center** | Dashboard showing active applications, upcoming interviews, and deadlines closing within the week |

---

## Features in Detail

### Opportunity Discovery

An ingestion worker fetches live postings directly from public company job boards:

- **Sources**: Greenhouse, Lever, and Ashby board APIs
- **Filtering**: keeps intern, co-op, new-grad and entry-level roles; discards senior postings
- **Normalization**: company, title, location, remote flag, salary and deadline are mapped into one consistent shape regardless of source
- **Skill extraction**: requirements are matched against a reference skill list and stored as structured tags, not free text
- **Deduplication**: postings upsert on their source ID, so re-running the worker updates rather than duplicates
- **Stale closing**: postings that disappear from a board are automatically marked closed

Run it on demand:

```bash
make worker
```

### Explainable Matching

Every opportunity is scored against your profile using a transparent, deterministic
formula — no black box, no "trust the algorithm."

**Stage 1 — hard filters.** Closed postings, passed deadlines, roles you've already
applied to, and degree or graduation-year mismatches are removed outright.

**Stage 2 — weighted scoring:**

| Signal | Weight | What it measures |
|---|---|---|
| Skill Match | 40% | Overlap between your skills and the opportunity's requirements |
| Role Fit | 20% | Fuzzy match against your target roles |
| Eligibility | 15% | Degree, graduation year, CGPA compatibility |
| Location | 10% | Match against your preferred locations and remote preference |
| Project Relevance | 5% | Skills demonstrated in your portfolio projects |
| Semantic | 10% | _Not yet implemented_ — redistributed across the signals above |

Each result carries its own reasoning: verified positive signals ("Verified skill: Python",
"Meets CGPA requirement (8.2 >= 7.5)") alongside a gap analysis ("Docker required but not
in profile"). You always know why something ranked where it did, and what would move it up.

### Application Tracking

Save an opportunity and it enters your pipeline. Move it through six stages, attach
interview dates and notes, and the dashboard keeps count of what's active, what's coming
up, and which saved postings are about to expire.

### Portfolio

Projects aren't decoration — each one links to the skills it demonstrates, and those
skills feed the matching engine. Adding a project that uses React measurably raises your
score against every React opening.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18, TypeScript, Vite 5, Tailwind CSS |
| Backend | Go 1.23, Gin, pgx/v5 |
| Database | PostgreSQL 16 (+ pgvector extension, schema-ready) |
| Authentication | Argon2id password hashing, JWT access tokens, rotating refresh tokens |
| Authorization | [Cedar](https://www.cedarpolicy.com/) policy engine (`cedar-go`) |
| Containers | Docker / Finch Compose |

```
                    STUDENT
                       │
                       ▼
              ┌────────────────┐
              │  React 18 SPA  │   localhost:3000
              │   Vite + TS    │
              └────────┬───────┘
                       │  /api
                       ▼
              ┌────────────────┐
              │   Go REST API  │   localhost:8080
              │   Gin + pgx    │   matching · auth · ingestion
              └────────┬───────┘
                       │
                       ▼
              ┌────────────────┐        ┌──────────────────┐
              │  PostgreSQL 16 │        │ Greenhouse/Lever │
              │                │◀───────│   /Ashby boards  │
              └────────────────┘ ingest └──────────────────┘
```

---

## Quick Start

### Prerequisites

- Docker + Docker Compose (or [Finch](https://github.com/runfinch/finch))
- Git

```bash
git clone https://github.com/shubZk17/student-os.git
cd student-os
cp .env.example .env
docker compose up -d          # or: finch compose up -d

curl http://localhost:8080/health
# → {"status":"healthy","database":"connected"}
```

Open [http://localhost:3000](http://localhost:3000) and register an account.

> **Note:** the container path is written but has not been run end to end — no container
> runtime was available on the development machine. If it fails, the local path below is
> the one that is known to work.

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Health check | http://localhost:8080/health |
| PostgreSQL | localhost:5432 |

### Load Demo Data

```bash
make seed-dev       # sample opportunities, so the app isn't empty on first run
```

### Pull Real Opportunities

```bash
make worker         # fetches live postings from Greenhouse, Lever and Ashby
```

---

## Local Development (No Containers)

```bash
# Terminal 1 — PostgreSQL only
make docker-up

# Terminal 2 — Backend (migrations apply automatically on boot)
cd backend && go run cmd/server/main.go

# Terminal 3 — Frontend with hot reload
cd frontend && npm install && npm run dev
```

Frontend runs at http://localhost:5173 and proxies the API to http://localhost:8080.

---

## API

All routes are prefixed `/api/v1`. Protected routes take `Authorization: Bearer <token>`.

| Method | Route | Auth | Purpose |
|---|---|---|---|
| POST | `/auth/register` | — | Create an account (also creates the profile) |
| POST | `/auth/login` | — | Exchange credentials for a token pair |
| POST | `/auth/refresh` | — | Rotate a refresh token |
| POST | `/auth/logout` | — | Revoke a refresh token |
| GET | `/opportunities` | — | Search, filter and paginate openings |
| GET | `/opportunities/:id` | — | Single opportunity detail |
| GET | `/recommendations` | ✓ | Ranked, explained matches for you |
| GET · PUT | `/profile` | ✓ | Read / update profile and skills |
| GET · POST | `/applications` | ✓ | List / add to the tracker |
| PATCH · DELETE | `/applications/:id` | ✓ | Move stage / remove |
| GET · POST | `/projects` | ✓ | List / create portfolio projects |
| DELETE | `/projects/:id` | ✓ | Remove a project |
| GET | `/notifications` | ✓ | List notifications and unread count |
| PATCH | `/notifications/:id/read` | ✓ | Mark one read |
| GET | `/dashboard/summary` | ✓ | Command-center metrics |

### Security

- Passwords hashed with Argon2id (64 MB, 3 iterations), with bounded concurrency so a login burst can't exhaust memory
- Short-lived access tokens with single-use rotating refresh tokens
- Every query parameterized; every student-owned record scoped by owner in SQL
- Ownership rules additionally declared in a readable [Cedar policy](backend/internal/authz/policies.cedar) and checked before the query runs — the SQL predicate stays as a backstop
- Rate limiting on credential routes, strict CORS allowlist, 1 MiB request cap, and production config validation that refuses to boot with development secrets

---

## Project Structure

```text
student-os/
├── docker-compose.yml        # Full local stack
├── Makefile                  # Developer commands
├── .env.example              # Environment variables (no secrets)
├── seeds/dev_seed.sql        # Demo data (dev only)
├── backend/
│   ├── cmd/
│   │   ├── server/           # API server entrypoint
│   │   └── worker/           # Ingestion worker entrypoint
│   ├── internal/
│   │   ├── auth/             # Argon2id, JWT, auth handlers
│   │   ├── authz/            # Cedar authorization policy
│   │   ├── users/            # Student profiles & skills
│   │   ├── jobs/             # Opportunities & job-board adapters
│   │   ├── matching/         # Explainable matching engine
│   │   ├── applications/     # Kanban application tracking
│   │   ├── projects/         # Portfolio & project skills
│   │   ├── ingest/           # Job-board ingestion (CLI + API)
│   │   ├── notifications/    # In-app notifications
│   │   ├── dashboard/        # Command-center metrics
│   │   ├── database/         # Connection pooling
│   │   └── middleware/       # Auth, CORS, rate-limiting, logging
│   └── migrations/           # Embedded SQL migrations (auto-applied)
└── frontend/
    ├── nginx.conf            # SPA routing + API proxy
    └── src/
        ├── api/              # Typed REST client
        ├── components/       # UI components & layouts
        ├── context/          # Auth context provider
        ├── pages/            # Dashboard, Opportunities, Applications, Projects, Profile
        └── types/            # TypeScript data models
```

---

## Current Limitations

An honest list of what is **not** true of this system today:

- **Semantic / vector matching is not implemented.** The schema declares `vector(384)` columns and the pgvector extension, but no code generates, stores, or queries embeddings. The 10% semantic weight is redistributed across the deterministic signals.
- **No file uploads.** `resume_url` and `portfolio_url` are text fields for links you type in.
- **Job ingestion is on-demand, not scheduled.** The ingestion code works and is run manually (`make worker`) or through an authenticated internal route; nothing triggers it on a timer yet.
- **Notifications have no producer.** The table, the API and the bell UI all work, but nothing in the codebase ever inserts a notification, so the bell is empty.
- **Interview scheduling is minimal.** An application carries an `interview_date` field, but there is no calendar integration.

---

## Testing

```bash
make test                          # backend unit tests
cd frontend && npm run build       # type-check and production build
```

Both run in CI on every push — see `.github/workflows/ci.yml`.

---

## Deployment

The canonical environment is local. Optional deployment configs are included:

| Platform | Config | Status |
|---|---|---|
| Vercel (frontend) | `frontend/vercel.json` | Live — https://student-os-go-solo1.vercel.app |
| Render (frontend + API + Postgres) | `render.yaml` | Live — API sleeps when idle |

`amplify.yml` and `backend/apprunner.yaml` are historical artifacts of a cloud migration
that was planned but never deployed.

---

## License

Apache-2.0 or MIT License.
