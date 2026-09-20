# StudentOS — AWS Ship It Migration Plan

**Status: PLAN ONLY. Nothing has been deployed to AWS.**

No AWS resource has been created, and no AWS account has been verified, because the
tooling required to do so is not present on this machine (see
[Blockers](#0-blockers--why-nothing-is-deployed-yet)). Every AWS item in this document is
**PLANNED**, not implemented.

Audit findings below were read directly from the source at commit `1356259`. Where the
existing documentation disagrees with the code, the code is treated as the truth and the
discrepancy is called out.

---

## 0. Blockers — why nothing is deployed yet

| Requirement | Needed for | Present? |
| --- | --- | --- |
| AWS CLI (`aws`) | §7 account verification, all provisioning | **No** — not installed, not on PATH |
| AWS credentials | Everything | **No** — no `~/.aws`, no `AWS_*` env vars |
| Docker | §12 local container verification, ECR image for App Runner | **No** — not installed |
| AWS SAM or CDK | §26 infrastructure as code | **No** — neither installed |

Spec §7 says *"Do NOT proceed if the CLI is authenticated against an unexpected AWS
account."* I cannot run `aws sts get-caller-identity` at all, so I cannot confirm the
account holding the $100 credit. Provisioning is therefore not attempted.

**To unblock**, on this machine:

```bash
# 1. AWS CLI v2
winget install Amazon.AWSCLI

# 2. Credentials (interactive — run this yourself, never paste keys into chat)
aws configure sso          # preferred
# or: aws configure

# 3. Verify the account holding the credit
aws sts get-caller-identity

# 4. Container tooling for the App Runner image
winget install Docker.DockerDesktop

# 5. IaC (pick ONE — this plan assumes CDK)
npm install -g aws-cdk
```

---

## 1. Current architecture (verified against source)

```text
Browser
   │
   ├── Vercel (static React/Vite SPA)        https://student-os-go-solo1.vercel.app
   └── Render static site (second copy)      https://studentos-web.onrender.com
              │
              │ HTTPS, Bearer JWT
              ▼
   Render Web Service (Docker, Go/Gin)       https://studentos-api-0yqr.onrender.com
              │
              ▼
   Render PostgreSQL 16 (free plan)
              │
   (migrations applied by the API on startup)

   Ingestion worker (backend/cmd/worker) — NOT DEPLOYED, run by hand only
```

| Component | Detail |
| --- | --- |
| Frontend | React 18 + Vite 5 + TypeScript + Tailwind. Build `tsc && vite build` → `dist/`. API base from `VITE_API_URL`, baked at build time. |
| Backend | Go 1.23, Gin, `pgx/v5`. Entrypoint `backend/cmd/server`. Port from `PORT` (default 8080). Graceful shutdown on SIGINT/SIGTERM present. |
| Health | `GET /health` → `{"database":"connected","status":"healthy"}`. Verified live. |
| Migrations | Embedded SQL, applied automatically on server start (`backend/migrations`). 3 migrations. |
| Auth | Local. Argon2id password hashing, HS256 JWT access tokens signed with `JWT_SECRET`, opaque UUID refresh tokens persisted in Postgres and single-use. |
| CORS | `ALLOWED_ORIGINS` comma-separated allowlist. Config refuses `*` when `ENV=production`. |
| Ingestion | `backend/cmd/worker` — a **run-to-completion batch process**, not a server. Fetches Greenhouse/Lever/Ashby, normalises, dedupes, extracts skills, closes stale postings, exits. |
| Rate limiting | In-process, keyed on client IP via `TRUSTED_PLATFORM` header. |

### 1.1 Two findings that change what we can honestly claim

**Finding A — semantic/vector matching is a placeholder.**

The schema does create the extension and the columns:

```sql
CREATE EXTENSION IF NOT EXISTS "vector";   -- 000001_init_schema.up.sql:5
embedding vector(384),                      -- lines 50 and 93
```

But **no Go code ever reads or writes an embedding.** The only call site of the matching
engine passes semantic similarity as a literal zero:

```go
// backend/internal/jobs/handler.go:99
}, 0.0)
```

and the engine then redistributes that weight across other deterministic signals:

```go
// backend/internal/matching/engine.go:203-207
semScore := semanticSimilarity
if semScore <= 0 {
    semScore = (skillScore + roleScore + locScore) / 3.0
}
```

So the advertised "Semantic 10%" is, in practice, **10% extra weight on skill/role/location**.
No embeddings are generated, stored, or queried. `README.md` currently claims *"PostgreSQL
with `pgvector` extension for semantic embeddings"*, which overstates this.

Per spec §14 and §34 this must be documented as **PLANNED, not implemented**, and we must
not spend credit on embedding infrastructure during the migration. The 40/20/15/10/5/10
weights stay exactly as they are (§33) — only the documentation changes.

**Finding B — S3 has no current use.**

There is no upload endpoint anywhere in the backend: no `multipart`, no `FormFile`, no
upload handler. `resume_url` and `portfolio_url` are plain text columns holding URLs the
student types in. Per §19 ("do not add S3 simply because it is an AWS Ship It service"),
**S3 is deferred** and not provisioned.

---

## 2. Target architecture

```text
Browser
   │
   ▼
AWS Amplify Hosting ── React SPA (HTTPS, SPA rewrite)
   │
   │ HTTPS, Bearer JWT
   ▼
AWS App Runner ── Go API container (public endpoint, autoscale min 1)
   │
   ├────────────▶ Amazon RDS PostgreSQL 16 (db.t4g.micro, single-AZ, public, SG-locked)
   │
   └────────────▶ Amazon CloudWatch Logs (App Runner ships logs natively)

EventBridge Scheduler ──(HTTPS + secret header)──▶ App Runner POST /internal/ingest
                                                            │
                                                            ▼
                                                   existing ingestion code
                                                            │
                                                            ▼
                                                          RDS
```

No VPC connector, no NAT Gateway, no ALB, no ECS cluster, no Lambda.

---

## 3. Migration table

Every row is checked against the repository, and marked with what it costs us in work.

| Current | Target | Status | Notes |
| --- | --- | --- | --- |
| Vercel + Render static frontend | AWS Amplify Hosting | **CONFIG DONE**, not deployed | `amplify.yml` added (monorepo appRoot `frontend`). Still needs, in the console: `VITE_API_URL` and the `/<*>` to `/index.html` 200 rewrite. App not created. |
| Render Web Service (Docker) | AWS App Runner | **CONFIG DONE**, not deployed | `backend/apprunner.yaml` added (source-based, no Docker/ECR needed). Dockerfile kept as the image-based alternative. Service not created. |
| Render PostgreSQL | Amazon RDS PostgreSQL 16 | PLANNED | Schema unchanged. Migrations self-apply on boot. |
| `CREATE EXTENSION vector` | pgvector on RDS | PLANNED | RDS Postgres 16 supports pgvector. Extension will be enabled so migration 1 succeeds — but see Finding A: **no embeddings are used.** Enabling it is required for the migration to run, not evidence of semantic search. |
| Local Argon2id + HS256 JWT | Amazon Cognito | **DEFERRED — see §5** | High risk, low hackathon value. Recommendation: keep existing auth. |
| `resume_url` text field | Amazon S3 | **DEFERRED** | No upload feature exists (Finding B). |
| Ingestion cron (not deployed) | EventBridge Scheduler → App Runner route | **CODE DONE**, not deployed | `internal/ingest` package + authenticated `POST /internal/ingest`. Unit-tested; the AWS schedule does not exist yet. |
| Render logs | Amazon CloudWatch Logs | PLANNED | Automatic with App Runner. Existing request-ID/status/latency log lines are preserved as-is. |
| — | AWS Budgets ($10/$25/$50) | PLANNED | §29. Create before RDS/App Runner. |

---

## 4. Cost — what I can and cannot tell you

**I have not checked current AWS pricing**, because that requires the account and the
pricing API. Spec §40.8 says not to invent estimates, so I am not putting a dollar figure
on this plan.

What I can state as a structural risk, to be confirmed before provisioning:

- **App Runner is the main credit risk.** It bills for provisioned container memory
  continuously whenever a service exists, including while idle, and it has historically had
  no always-free tier. With `min instances = 1` this is a standing charge for the whole
  hackathon period. Confirm current pricing before creating the service, and consider
  pausing the service between demos.
- **RDS free tier is account-age dependent.** 750 hours of `db.t4g.micro` is a
  *new-account, first-12-months* offer. If this account is older, RDS bills hourly.
  Confirm against the account's free-tier status before creating the instance.
- **Amplify, CloudWatch, EventBridge Scheduler, Cognito** are expected to be negligible at
  demo traffic, but "expected" is not "verified".
- **NAT Gateway is the classic credit-killer** (~$0.045/hr plus data processing, always on).
  This plan avoids it entirely by keeping App Runner on its default public egress and RDS
  publicly addressable but locked by security group.

**Action before any provisioning:** create the AWS Budget first (§29), then check
free-tier status in the Billing console, then provision RDS, then App Runner.

---

## 5. Authentication: recommendation is to NOT migrate to Cognito

Spec §17 says *"Do not blindly replace working authentication."* Here is why I think
Cognito is the wrong trade for this project, stated plainly so you can overrule it.

**The cost of migrating:**

1. **Every table foreign-keys to `users.id`** (an internal UUID). Cognito issues tokens
   whose `sub` is a Cognito-owned identifier. You must either keep a `users` row per
   Cognito user and map `sub → users.id` on every request (so you still run a users
   table — Cognito buys you little), or rewrite every FK.
2. **Existing Argon2id password hashes cannot be imported into Cognito.** Cognito's bulk
   import does not accept external hashes; the only path is a Migrate-User Lambda trigger
   that re-verifies against the old store on first login — which means adding Lambda, and
   keeping the old auth code alive anyway.
3. **Token validation changes** from HS256-with-shared-secret to RS256-against-JWKS, so
   `middleware/auth.go` and the refresh flow are rewritten.
4. **The refresh-token table and its single-use rotation logic** get replaced by Cognito's
   own refresh handling — a behaviour change in the security-sensitive path.
5. The frontend auth context, login, register and token-refresh paths all change.

**What it buys for a hackathon demo:** a managed user pool, and the ability to say
"Cognito" on the slide.

**Recommendation:** keep the existing auth for the submission and document it honestly as
"custom JWT auth, Cognito evaluated and deferred". It already does the things that matter
(Argon2id, short-lived access tokens, single-use refresh rotation, per-user scoping).

**If you want Cognito anyway**, say so and I will do it as a separate, self-contained
phase with its own rollback — but it should not be bundled with the infrastructure move,
because if the deployment breaks you won't know which change did it.

---

## 6. Ingestion: how EventBridge actually drives a batch process

`backend/cmd/worker` runs to completion and exits. EventBridge cannot "run" it — EventBridge
invokes a *target*. The options:

| Option | Verdict |
| --- | --- |
| EventBridge → Lambda running the worker | Rejected. 15-min hard ceiling vs the worker's own 15-min context; needs a separate build artifact and a VPC path to RDS. |
| EventBridge → ECS/Fargate task | Rejected by §6 (no ECS/Fargate). Also needs networking we're avoiding. |
| **EventBridge Scheduler → App Runner HTTPS route** | **Chosen.** No new compute, no new networking, reuses the running service. |

**Required code change:** add `POST /internal/ingest` to the existing API, guarded by a
shared secret header (`X-Ingest-Token`, value from env), which runs the existing ingestion
routine in a goroutine and returns 202 immediately so the scheduler does not wait.

This is a genuine change to `cmd/server` and a refactor to make the worker logic callable
as a package rather than only from `main()`. It carries the risk that a long ingestion run
competes for resources with live API traffic on the same small App Runner instance —
acceptable at hackathon scale, worth noting.

Schedule: `rate(6 hours)`, matching the existing documented cadence.

---

## 7. Migration risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| **App Runner idle billing drains the $100** | High | Budget alarms first. Consider pausing the service when not demoing. Confirm pricing before creating. |
| **RDS not free-tier-eligible on this account** | High | Check billing console before creating the instance. |
| **Accidental NAT Gateway** via a VPC connector | High | Do not attach App Runner to a VPC. Keep RDS public + SG-restricted. Explicitly verify no NAT exists after provisioning. |
| **RDS publicly accessible** | Security | Security group must allow only App Runner egress. App Runner's public egress IPs are not static, so this needs care — if it cannot be locked down tightly, revisit. **This is the weakest point of the lean design and must be verified, not assumed.** |
| **pgvector unavailable** → migration 1 fails → health check fails → deploy rolls back | High | Enable the extension on RDS *before* first boot. Verify `SELECT * FROM pg_extension`. |
| **CORS** — Amplify domain unknown until first deploy | Medium | Deploy Amplify first, read the domain, then set `ALLOWED_ORIGINS`. Config rejects `*` in production by design. |
| **`VITE_API_URL` is baked at build time** | Medium | Amplify must rebuild after the App Runner URL exists. Ordering: App Runner → set Amplify env → rebuild. |
| **Rate limiter breaks behind App Runner** | High | `c.ClientIP()` resolves to App Runner's proxy, making the 30/5min credential limit global rather than per-client. `TRUSTED_PLATFORM=X-Forwarded-For` is not a fix (client-forgeable). Options recorded in `backend/apprunner.yaml`; **must be decided before launch.** |
| **Guessing service URLs** | Medium | Already bit us twice this project: `studentos-api.onrender.com` and `student-os.vercel.app` are *other people's* live sites. Always read the real URL from the console; never assume the name. |
| **Data loss** | High | Render's Postgres holds real data. Export with `pg_dump` before cutover; do not delete Render resources until AWS is verified. |
| **Secrets in git** | High | App Runner env vars / Secrets Manager only. `.env.example` keeps placeholders. |
| **IAM permissions insufficient** | Medium | Verify with `get-caller-identity` and a dry run before the long path. |

---

## 8. Rollback

The existing deployment is **left running and untouched** throughout. Rollback is therefore
cheap at every stage:

1. **Frontend** — Vercel and the Render static site stay live. Rolling back = keep using
   `https://student-os-go-solo1.vercel.app`. Nothing to undo.
2. **Backend** — Render service `studentos-api-0yqr.onrender.com` stays live. Rolling back =
   point `VITE_API_URL` back at it and redeploy the frontend.
3. **Database** — Render Postgres remains the system of record until AWS is verified
   end-to-end. The RDS copy is loaded from a `pg_dump`; the original is not dropped.
4. **Code** — all work is on `feat/aws-ship-it`. `main` is unchanged and currently deployed.
   Rolling back = do not merge the branch.
5. **AWS teardown** — all resources carry `Project=StudentOS`, `Environment=hackathon` tags
   so they can be found and deleted. Delete order: EventBridge schedule → App Runner →
   Amplify → RDS (final snapshot) → budgets.

**The Render free Postgres expires ~30 days after creation (created 2026-09-20).** That is
the real deadline on this rollback path, not a safety net forever.

---

## 9. Execution order (once unblocked)

| # | Step | Verified by |
| --- | --- | --- |
| 1 | `aws sts get-caller-identity`, record account ID | Output matches the credit account |
| 2 | Create AWS Budgets: $10 / $25 / $50 | Budget visible in console |
| 3 | Check RDS free-tier eligibility | Billing console |
| 4 | `docker build` + `docker run` the backend locally | `/health` returns healthy against local PG |
| 5 | Create RDS, enable pgvector | `SELECT extname FROM pg_extension` includes `vector` |
| 6 | `pg_dump` Render → restore to RDS | Row counts match per table |
| 7 | Push image to ECR, create App Runner service | `/health` on the real App Runner URL |
| 8 | Confirm **no NAT Gateway exists** | `aws ec2 describe-nat-gateways` returns empty |
| 9 | Create Amplify app, set `VITE_API_URL`, build | SPA loads, deep link refresh works |
| 10 | Set `ALLOWED_ORIGINS` to the Amplify domain, redeploy API | Browser request succeeds, no CORS error |
| 11 | Add `/internal/ingest`, create EventBridge schedule | Opportunity row count increases after a run |
| 12 | End-to-end acceptance test (§32) | Each step of the chain passes individually |
| 13 | Write `AWS_DEPLOYMENT_REPORT.md` from observed results | — |

Documentation (`README.md`, `ARCHITECTURE.md`) is updated at step 13 from what was actually
observed — not before, so it cannot describe something that did not happen.
