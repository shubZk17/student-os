.PHONY: help dev-backend dev-frontend dev docker-up docker-down seed-dev test build \
       finch-up finch-down finch-build finch-logs localstack-init sam-invoke worker

help:
	@echo "StudentOS Development Commands:"
	@echo ""
	@echo "  ── Quick Start (Build It) ──────────────────────────────────────"
	@echo "  make finch-up        Start entire stack (Finch or Docker)"
	@echo "  make finch-down      Stop entire stack"
	@echo "  make finch-build     Rebuild all containers"
	@echo "  make finch-logs      Tail logs from all services"
	@echo ""
	@echo "  ── Individual Services ─────────────────────────────────────────"
	@echo "  make docker-up       Start PostgreSQL only"
	@echo "  make docker-down     Stop PostgreSQL container"
	@echo "  make dev-backend     Run Go backend (no container)"
	@echo "  make dev-frontend    Run React Vite dev server (no container)"
	@echo ""
	@echo "  ── Data & Ingestion ────────────────────────────────────────────"
	@echo "  make seed-dev        Load demo data (start backend once first)"
	@echo "  make worker          Run ingestion worker manually"
	@echo "  make sam-invoke      Run ingestion via SAM CLI (requires SAM)"
	@echo ""
	@echo "  ── LocalStack ─────────────────────────────────────────────────"
	@echo "  make localstack-init  Re-run LocalStack initialization"
	@echo "  make localstack-s3    List S3 buckets in LocalStack"
	@echo ""
	@echo "  ── CI ──────────────────────────────────────────────────────────"
	@echo "  make test            Run backend unit tests"
	@echo "  make build           Build frontend and backend for production"

# ── Full Stack (Finch / Docker Compose) ────────────────────────────────────
# Works with both `finch compose` and `docker compose`.
# Detects which is available and uses it.
COMPOSE := $(shell command -v finch 2>/dev/null && echo "finch compose" || echo "docker compose")

finch-up:
	$(COMPOSE) up -d
	@echo ""
	@echo "StudentOS is starting..."
	@echo "  Frontend:   http://localhost:3000"
	@echo "  Backend:    http://localhost:8080"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  LocalStack: http://localhost:4566"
	@echo "  PostgreSQL: localhost:5432"

finch-down:
	$(COMPOSE) down

finch-build:
	$(COMPOSE) build

finch-logs:
	$(COMPOSE) logs -f

# ── PostgreSQL Only ────────────────────────────────────────────────────────
docker-up:
	$(COMPOSE) up -d postgres

docker-down:
	$(COMPOSE) down

# ── Development Seeds ──────────────────────────────────────────────────────
seed-dev:
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U studentos -d studentos_db < seeds/dev_seed.sql

# ── Bare-Metal Development (no containers for app, just DB) ────────────────
dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

worker:
	cd backend && go run cmd/worker/main.go

# ── SAM CLI ────────────────────────────────────────────────────────────────
sam-invoke:
	cd sam && sam local invoke IngestFunction --docker-network student-os_default

# ── LocalStack Utilities ──────────────────────────────────────────────────
localstack-init:
	$(COMPOSE) exec localstack bash /etc/localstack/init/ready.d/init-aws.sh

localstack-s3:
	aws --endpoint-url=http://localhost:4566 s3 ls

# ── CI & Build ─────────────────────────────────────────────────────────────
test:
	cd backend && go test ./...

build:
	cd backend && go build -o bin/server cmd/server/main.go
	cd frontend && npm run build
