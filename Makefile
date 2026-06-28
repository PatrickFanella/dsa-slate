SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

COMPOSE := docker compose
WEB := pnpm --dir web

.PHONY: help verify deps fmt lint test build compose-config open-pilot-check \
	containers-up containers-down containers-clean logs db-up db-shell migrate db-reset seed dev-web dev-api api-smoke dev

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

db-shell: db-up ## Open psql in the Postgres container
	$(COMPOSE) exec postgres psql -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}"

migrate: db-up ## Apply SQL migrations to local Postgres
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}" < migrations/0001_init.sql

db-reset: ## Reset local Postgres volume and rerun migrations
	$(COMPOSE) down --volumes --remove-orphans
	$(MAKE) migrate

seed: migrate ## Apply the California seed fixture
	$(COMPOSE) exec -T postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-dsa_slate}" -d "$${POSTGRES_DB:-dsa_slate}" < seeds/california.sql

dev-web: deps ## Start local Vite frontend dev server
	$(WEB) run dev

dev-api: db-up ## Run the Go backend locally against compose Postgres
	@published_port="$$( $(COMPOSE) port postgres 5432 )"; \
	postgres_port="$${published_port##*:}"; \
	DATABASE_URL="$${DATABASE_URL:-postgres://$${POSTGRES_USER:-dsa_slate}:$${POSTGRES_PASSWORD:-change-me}@localhost:$${POSTGRES_PORT:-$${postgres_port:-5432}}/$${POSTGRES_DB:-dsa_slate}?sslmode=disable}"; \
	export DATABASE_URL; \
	cd backend && go run ./cmd/app

api-smoke: ## Curl the local API health and directory endpoints
	@curl -fsS "http://localhost:$${API_PORT:-8080}/healthz" >/dev/null
	@curl -fsS "http://localhost:$${API_PORT:-8080}/api/directory" >/dev/null

dev: db-up migrate dev-web ## Start database, run migrations, then start frontend

open-pilot-check:
	@test -f .gitea/ISSUE_TEMPLATE/open-pilot-task.yaml
	@grep -q 'Test Command' .gitea/ISSUE_TEMPLATE/open-pilot-task.yaml
