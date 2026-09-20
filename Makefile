.PHONY: help dev-backend dev-frontend dev docker-up docker-down seed-dev test build

help:
	@echo "StudentOS Development Commands:"
	@echo "  make dev-backend   - Run Go backend server"
	@echo "  make dev-frontend  - Run React Vite dev server"
	@echo "  make docker-up     - Start PostgreSQL with pgvector"
	@echo "  make docker-down   - Stop PostgreSQL container"
	@echo "  make seed-dev      - Load demo data (dev only; start the backend once first)"
	@echo "  make test          - Run backend automated unit tests"
	@echo "  make build         - Build frontend and backend for production"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

seed-dev:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U studentos -d studentos_db < seeds/dev_seed.sql

dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm.cmd run dev

test:
	cd backend && go test ./...

build:
	cd backend && go build -o bin/server cmd/server/main.go
	cd frontend && npm.cmd run build
