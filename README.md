# DSA Slate

Public, source-linked election guide for DSA electoral work.

DSA Slate helps people find DSA-endorsed candidates, ballot measures, and voter guides across national, state, and local elections. The project starts as a verified directory with source links before adding address-based lookup or results automation.

## Product direction

Core promise:

> Find every DSA-endorsed candidate, ballot measure, and voter guide — national, state, and local — in one searchable place.

First trusted-directory slice:

- homepage and static public directory browsing
- California-cluster seed data
- candidate recommendation, ballot measure, and voter guide cards
- source URLs, source dates, and confidence labels on every public item
- manual seed-file workflow with documented editorial rules
- no raw voter address storage

Out of scope for the first launch: address lookup, admin UI, public submissions, live election-night results, campaign coordination, address storage, and automated nationwide scraping.

## Stack

- Go backend
- Vite React TypeScript frontend
- PostgreSQL
- Docker Compose
- Makefile-based verification
- Open Pilot issue and PR conventions

## Quick start

```bash
make verify
```

Useful local commands:

```bash
make db-up          # start Postgres
make migrate        # apply SQL migrations
make seed           # apply the California seed fixture
make dev-web        # start Vite locally; /api proxies to http://localhost:8080
make dev-api        # run the Go API locally against compose Postgres
make api-smoke      # curl /healthz and /api/directory on localhost
make dev            # start DB, migrate, then run Vite
make dev-full       # seed the DB, then print the API + web commands to run in separate terminals
make containers-up  # build and run Postgres, API, and web containers
make logs           # follow compose logs
```

For the API-backed frontend flow, run `make seed` or `make dev-full` first so the database has directory data, then start the API with `make dev-api` and run `make dev-web` in another terminal. The Vite dev server proxies `/api` to `http://localhost:8080`, so the frontend can read live directory data while keeping the checked-in California seed as fallback.

## Development / API

Read API contract: `docs/reference/api.md`.
Design reference: `docs/reference/design.md`.
The frontend uses official DSA palette tokens mapped into Tailwind v4 semantic utilities in `web/src/styles.css`.
The public UI includes a light/dark theme toggle driven by `data-theme` on `document.documentElement`.

Recommended feature order: stabilize the read API and seed data first, then build frontend screens against that contract. Avoid frontend-only data shapes that diverge from the Go API DTOs.

## Open Pilot

This repository keeps the base Open Pilot issue template and PR template.
Bootstrap labels after creating the repository:

```bash
open-pilot labels bootstrap subculture-collective/dsa-slate
```

## License

GPL-3.0-only. See `LICENSE`.
