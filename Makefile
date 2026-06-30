SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

COMPOSE := docker compose
WEB := pnpm --dir web

.PHONY: help verify deps fmt lint test build compose-config open-pilot-check \
	containers-up containers-down containers-clean logs db-up db-sync-password db-shell migrate db-reset seed-chapters seed data-quality dev-web dev-api api-smoke dev dev-full

help:
	@awk 'BEGIN {FS = ":.*##"; printf "Commands:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps:
	$(WEB) install --frozen-lockfile

verify: deps fmt lint test build compose-config open-pilot-check ## Run all checks

fmt:
	@git ls-files --cached --others --exclude-standard 'backend/*.go' | xargs -r gofmt -w
	$(WEB) run format

lint:
	cd backend && go vet ./...
	$(WEB) run lint

test:
	cd backend && go test ./...
	$(WEB) run test

build:
	mkdir -p bin
	cd backend && go build -o ../bin/dsa-slate ./cmd/app
	$(WEB) run build

compose-config:
	$(COMPOSE) config --quiet

containers-up: ## Build and start Postgres, API, and web containers
	$(COMPOSE) up -d --build

containers-down: ## Stop compose containers and keep volumes
	$(COMPOSE) down

containers-clean: ## Stop compose containers and remove volumes
	$(COMPOSE) down --volumes --remove-orphans

logs: ## Follow compose logs
	$(COMPOSE) logs -f

db-up: ## Start Postgres and wait until healthy
	$(COMPOSE) up -d --wait postgres

db-sync-password: db-up ## Sync local Postgres role password from .env/compose env without printing it
	python3 scripts/sync_postgres_password_sql.py | $(COMPOSE) exec -T postgres sh -lc 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" >/dev/null'

db-shell: db-up ## Open psql in the Postgres container
	$(COMPOSE) exec postgres psql -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}"

migrate: db-up ## Apply SQL migrations to local Postgres
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}" < migrations/0001_init.sql

db-reset: ## Reset local Postgres volume and rerun migrations
	$(COMPOSE) down --volumes --remove-orphans
	$(MAKE) migrate

seed-chapters: migrate ## Apply the one-time imported DSA chapter directory seed
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}" < seeds/chapters.sql

seed: migrate seed-chapters ## Apply the chapter directory and California electoral seed fixtures
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}" < seeds/california.sql

data-quality: seed db-sync-password ## Run source-linked data quality checks against local seeded Postgres
	@published_port="$$( $(COMPOSE) port postgres 5432 )"; \
	postgres_port="$${published_port##*:}"; \
	DATABASE_HOST="$${DATABASE_HOST:-localhost}"; \
	DATABASE_PORT="$${DATABASE_PORT:-$${POSTGRES_PORT:-$${postgres_port:-5432}}}"; \
	QUALITY_REFERENCE_DATE="$${QUALITY_REFERENCE_DATE:-2026-06-28}"; \
	export DATABASE_HOST DATABASE_PORT POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB QUALITY_REFERENCE_DATE; \
	cd backend && go run ./cmd/dataquality

dev-web: deps ## Start local Vite frontend dev server
	$(WEB) run dev

dev-api: db-sync-password ## Run the Go backend locally against compose Postgres
	@published_port="$$( $(COMPOSE) port postgres 5432 )"; \
	if [ -f .env ]; then set -a; source .env; set +a; fi; \
	postgres_port="$${published_port##*:}"; \
	DATABASE_HOST="$${DATABASE_HOST:-localhost}"; \
	DATABASE_PORT="$${DATABASE_PORT:-$${POSTGRES_PORT:-$${postgres_port:-5432}}}"; \
	export DATABASE_HOST DATABASE_PORT POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB GOOGLE_CIVIC_API_KEY ADMIN_API_TOKEN; \
	cd backend && go run ./cmd/app

api-smoke: ## Curl the local API health and directory endpoints
	@curl -fsS "http://localhost:$${API_PORT:-8080}/healthz" >/dev/null
	@curl -fsS "http://localhost:$${API_PORT:-8080}/api/directory" >/dev/null

dev: db-up migrate dev-web ## Start database, run migrations, then start frontend

dev-full: seed ## Prepare DB seed, then print commands for API and web dev servers
	@printf '%s\n' 'make dev-api'
	@printf '%s\n' 'make dev-web'

open-pilot-check:
	@test -f .gitea/ISSUE_TEMPLATE/open-pilot-task.yaml
	@grep -q 'Test Command' .gitea/ISSUE_TEMPLATE/open-pilot-task.yaml
