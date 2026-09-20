# StudentOS — Build It Track

> **WeMakeDevs First Commit — Build It: Open source, on your machine.**

This document explains how StudentOS uses the Build It track technologies to create a fully reproducible, local-first development and demonstration environment. No AWS account is required.

---

## Why Build It

StudentOS was originally developed with Vercel (frontend) and Render (backend + PostgreSQL) for cloud hosting. An AWS Ship It migration was planned but did not succeed — the required tooling (AWS CLI, credentials, Docker) was not available on the development machine.

The Build It track is a better fit for StudentOS because:

1. **The product works locally.** The Go backend, React frontend, and PostgreSQL database have no hard dependency on any cloud provider.
2. **Reproducibility matters more than hosting.** A hackathon judge should be able to clone the repo and run the entire platform in minutes.
3. **AWS-compatible architecture can be demonstrated locally.** Using Finch + LocalStack + SAM CLI, we show how StudentOS _could_ run on AWS — without spending credits or requiring an account.

---

## Technologies Used

Only technologies that are **actually configured and functional** in this repository:

| Technology | Category | How it's used |
|---|---|---|
| **Finch** | Containers | Local container runtime. Builds and runs all services via `finch compose`. |
| **LocalStack** | AWS Emulation | Emulates S3 (asset storage) and EventBridge (ingestion scheduling) locally. |
| **SAM CLI** | Serverless | Local invocation of the ingestion worker as a container-based Lambda. |
| **Cedar** | Auth / policy | Authorization policy for student-owned records, enforced in the API via `cedar-go`. See [Cedar](#cedar--authorization-policy) below. |
| **PostgreSQL 16** | Database | Primary data store with pgvector extension. |
| **pgvector** | Search | Vector extension installed (schema-ready for future semantic matching). |
| **React 18** | Frontend | TypeScript SPA with Vite and Tailwind CSS. |
| **Go 1.23** | Backend | REST API with Gin, Argon2id auth, and pgx driver. |

### Verification status

Being explicit, because it matters more than the list above:

| Technology | Status |
|---|---|
| **Cedar** | **Verified.** `go test ./internal/authz/...` passes, including cross-student denial. |
| Go backend, React frontend, PostgreSQL | **Verified.** Build, unit tests, and the live deployment all pass. |
| Finch, LocalStack, SAM CLI | **Configured, not yet run.** No container runtime was available on the development machine (Windows with no WSL distribution), so `finch compose up` has not been executed end to end. The Compose, LocalStack and SAM files are written but unproven. |

### Technologies Evaluated and Not Used

| Technology | Reason |
|---|---|
| OpenSearch | PostgreSQL search is adequate at current scale. |
| Strands Agents SDK | No AI agent feature needed for core product. |
| Firecracker | Requires Linux host; no user-facing benefit. |
| EKS / EKS Anywhere | Kubernetes is overkill for 3-4 containers. |
| Cognito | Belongs to Ship It track. Existing Argon2id auth works. |
| Corretto | Standard Go and Node runtimes are sufficient. |

---

## Architecture

```
                         STUDENT
                            │
                            ▼
                    ┌───────────────┐
                    │ React 18 SPA  │    Port 3000
                    │ Vite + TS +   │    nginx serves built assets
                    │ Tailwind CSS  │    proxies /api → backend
                    └───────┬───────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ Go REST API   │    Port 8080
                    │ Gin + pgx/v5  │    12 internal packages
                    │ Argon2id JWT  │    Health: GET /health
                    └───────┬───────┘
                            │
                 ┌──────────┴──────────┐
                 │                     │
                 ▼                     ▼
        ┌─────────────────┐    ┌─────────────────┐
        │ PostgreSQL 16   │    │   LocalStack    │
        │ + pgvector      │    │                 │
        │ Port 5432       │    │ Port 4566       │
        │                 │    │                 │
        │ Tables:         │    │ Services:       │
        │  users          │    │  S3             │
        │  student_profiles│   │  EventBridge    │
        │  skills         │    │                 │
        │  opportunities  │    └────────┬────────┘
        │  applications   │             │
        │  projects       │    ┌────────┴────────┐
        │  notifications  │    │                 │
        │  refresh_tokens │    ▼                 ▼
        └─────────────────┘   S3 Bucket     EventBridge
                              studentos-    6-hour schedule
                              assets        for ingestion

     ┌─────────────────────────────────────────────────┐
     │                    FINCH                        │
     │  Container runtime for all services             │
     │  finch compose up -d                            │
     └─────────────────────────────────────────────────┘

     ┌─────────────────────────────────────────────────┐
     │                   SAM CLI                       │
     │  sam local invoke IngestFunction                │
     │  Container-based Lambda for ingestion worker    │
     └─────────────────────────────────────────────────┘
```

---

## Installation

### Prerequisites

| Tool | Required | Install |
|---|---|---|
| Finch | Yes (or Docker) | https://github.com/runfinch/finch |
| Git | Yes | https://git-scm.com |
| SAM CLI | Optional | https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html |
| AWS CLI | Optional | Only for `awslocal` commands against LocalStack |
| Go 1.22+ | Only for bare-metal dev | https://go.dev |
| Node.js 22+ | Only for bare-metal dev | https://nodejs.org |

### Setup

```bash
# Clone the repository
git clone https://github.com/shubZk17/student-os.git
cd student-os

# Copy environment configuration
cp .env.example .env

# Build and start all services
finch compose up -d

# Verify health
curl http://localhost:8080/health
# → {"status":"healthy","database":"connected"}

# (Optional) Load demo data
make seed-dev

# (Optional) Pull real opportunities from job boards
finch compose exec backend /app/worker
```

---

## Running Locally

### Full Stack (Containers)

```bash
make finch-up       # Start everything
make finch-down     # Stop everything
make finch-logs     # Tail all logs
make finch-build    # Rebuild containers
```

### Bare-Metal (Faster Iteration)

```bash
make docker-up      # Start PostgreSQL only
make dev-backend    # Go API with hot reload
make dev-frontend   # Vite dev server with hot reload
```

---

## Testing

### Backend

```bash
make test
# Runs: go test ./... in backend/
```

### Frontend

```bash
cd frontend && npm run build
# Type-checks and builds the production bundle
```

### Infrastructure

```bash
# Health check
curl http://localhost:8080/health

# LocalStack S3
aws --endpoint-url=http://localhost:4566 s3 ls

# LocalStack EventBridge
aws --endpoint-url=http://localhost:4566 events list-rules
```

### SAM CLI

```bash
cd sam && sam local invoke IngestFunction
```

---

## Cedar — authorization policy

[Cedar](https://www.cedarpolicy.com/) is AWS's open-source authorization language. In
StudentOS it answers one question: *may this student act on this record?*

**The problem it solves.** Every student-owned table is scoped in SQL with
`AND user_id = $n`, repeated across nine handlers. That is secure, but the policy
exists only as a predicate copied into each query — there is no single place to read
what the rules are, and no way to test them without a database.

**What changed.** The rules now live in
[`backend/internal/authz/policies.cedar`](backend/internal/authz/policies.cedar), a
readable policy file, evaluated by `cedar-go` (pure Go, no AWS account, no service):

```cedar
permit (
    principal,
    action in [Action::"viewApplication", Action::"updateApplication",
               Action::"deleteApplication"],
    resource
)
when { resource.owner == principal };
```

**Enforcement is layered, not replaced.** The Cedar check runs *before* the query; the
SQL ownership predicate is deliberately kept. If a handler ever forgets to ask Cedar,
the database still refuses. Cedar is the auditable statement of intent; SQL is the
backstop.

**Verified:**

```bash
cd backend && go test ./internal/authz/... -v
```

The tests assert that an owner is allowed, a different student is denied on every
action (the IDOR cases), an action with no policy is denied, a record with no owner is
denied, and an empty principal is denied. Cedar is deny-by-default, so anything the
policy does not explicitly permit is refused.

Currently wired into the application-tracker endpoints
(`PATCH`/`DELETE /api/v1/applications/:id`). Projects and notifications still rely on
SQL scoping alone; extending them is mechanical and uses the same `authz.Can` call.

---

## AWS Compatibility

StudentOS demonstrates AWS-compatible architecture without requiring an AWS account:

| AWS Service | Local Equivalent | How |
|---|---|---|
| S3 | LocalStack S3 | `aws --endpoint-url=http://localhost:4566 s3 ls` |
| EventBridge | LocalStack EventBridge | Scheduled rule for ingestion |
| Lambda | SAM CLI local invoke | Container-based Lambda for worker |
| RDS PostgreSQL | PostgreSQL container | pgvector/pgvector:pg16 image |

The LocalStack initialization script (`localstack/init-aws.sh`) creates these resources automatically when the stack starts.

### What would change for real AWS deployment

1. Replace LocalStack endpoint URLs with real AWS endpoints
2. Add IAM roles and security groups
3. Replace container PostgreSQL with RDS
4. Deploy frontend to Amplify or CloudFront
5. Deploy backend to App Runner or ECS
6. Replace local EventBridge with AWS EventBridge Scheduler

The application code would not change — only the infrastructure configuration.

---

## No AWS Account Required

The entire Build It environment runs locally. The `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` in `.env.example` are set to `test` — these are LocalStack's default test credentials, not real AWS credentials.

No AWS account, no credit card, no cloud spending.

---

## Data Flow

```
Student → Register/Login → JWT tokens stored in localStorage
                              │
                              ▼
Student → Update Profile → Go API → PostgreSQL (student_profiles, student_skills)
                              │
                              ▼
Student → Browse Opportunities → Go API → PostgreSQL (opportunities)
                              │
                              ▼
                     Matching Engine
                    (40% skill, 20% role, 15% eligibility,
                     10% location, 5% project, 10% semantic*)
                              │
                              ▼
                     Ranked opportunities with
                     explainable match reasons
                              │
                              ▼
Student → Save/Apply → Go API → PostgreSQL (applications)
                              │
                              ▼
Student → Dashboard → Upcoming interviews, deadlines, match alerts


* Semantic matching is schema-ready but not implemented.
  The 10% weight is redistributed across skill, role, and location signals.
```

### Ingestion Flow

```
Manual trigger (make worker)
   or API route (POST /internal/ingest with INGEST_TOKEN)
   or SAM invoke (sam local invoke IngestFunction)
                    │
                    ▼
           ┌───────────────┐
           │  Job Providers │
           │  Greenhouse    │
           │  Lever         │
           │  Ashby         │
           └───────┬───────┘
                    │
                    ▼
           Fetch public job boards
                    │
                    ▼
           Filter student roles
           (intern, co-op, new grad, entry-level)
                    │
                    ▼
           Extract skills from known skill list
                    │
                    ▼
           Upsert into PostgreSQL
           (ON CONFLICT → update)
                    │
                    ▼
           Close stale postings
           (no longer listed on the board)
```
