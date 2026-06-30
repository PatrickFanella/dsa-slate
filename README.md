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
- authenticated saved-slate API for saving public record references
- no raw voter address storage

Out of scope for the first launch: address lookup, admin UI, public submissions, live election-night results, campaign coordination, address storage, and automated nationwide scraping.

Editorial/admin writes require a signed-in `editor` or `admin`; the draft console remains at `/admin`.
Google OIDC sign-in is available for local user sessions; open the home page and click “Sign in with Google”. `ADMIN_API_TOKEN` is only a temporary local/server fallback when the auth store is not wired, not browser UI auth.

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
make seed           # apply the chapter directory and California seed fixtures
make dev-web        # start Vite locally; /api proxies to http://localhost:8080
make dev-api        # run the Go API locally against compose Postgres
make api-smoke      # curl /healthz and /api/directory on localhost
make dev            # start DB, migrate, then run Vite
make dev-full       # seed the DB, then print the API + web commands to run in separate terminals
make containers-up  # build and run Postgres, API, and web containers
make logs           # follow compose logs
```

For the API-backed frontend flow, run `make seed` or `make dev-full` first so the database has directory data, then start the API with `make dev-api` and run `make dev-web` in another terminal. The Vite dev server proxies `/api` to `http://localhost:8080`, so the frontend can read live directory data while keeping the checked-in California seed as fallback.

Admin draft endpoints are session-gated for authenticated editors/admins and local/editorial-only; they are not public submission forms.

## Development / API

Read API contract: `docs/reference/api.md`.
Design reference: `docs/reference/design.md`.
Civic API/privacy reference: `docs/reference/civic-api.md`.
Private alpha PRD: `docs/product/private-alpha-prd.md`.
Private alpha architecture: `docs/product/private-alpha-architecture.md`.
The frontend uses official DSA palette tokens mapped into Tailwind v4 semantic utilities in `web/src/styles.css`.
The public UI includes a light/dark theme toggle driven by `data-theme` on `document.documentElement`.

Recommended feature order: stabilize the read API and seed data first, then build frontend screens against that contract. Avoid frontend-only data shapes that diverge from the Go API DTOs.

- Run `make data-quality` after seed edits to check source-linked records against local Postgres before proposing admin/write API work.
- `make seed` loads the imported DSA chapter directory from `seeds/chapters.sql`, then applies California electoral records from `seeds/california.sql`; runtime code reads Postgres, not the original CSV import source.
- Set `GOOGLE_CIVIC_API_KEY` in `.env` to enable anonymous `/api/civic/lookup`; `make dev-api` and Docker Compose pass it to the Go API without committing the key. The endpoint does not store raw addresses and does not confirm voter registration status.
- The public homepage includes a Civic lookup panel with a supported-election dropdown for anonymous address-based election info; the submitted address stays in browser component state only.
- For local Google sign-in, configure `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_OAUTH_REDIRECT_URL`, and `APP_BASE_URL`. In Google Cloud Console, add this redirect URI: `http://localhost:5173/api/auth/google/callback`.
- Google sign-in creates local `user` accounts. Existing `editor`/`admin` emails must have an explicit Google OAuth identity link before they can inherit privileged roles.
- Signed-in users can save public endorsement, measure-position, and voter-guide references through the saved-slate API. The app stores user IDs and public record IDs, not raw voter addresses.

## Open Pilot

This repository keeps the base Open Pilot issue template and PR template.
Bootstrap labels after creating the repository:

```bash
open-pilot labels bootstrap subculture-collective/dsa-slate
```

## License

GPL-3.0-only. See `LICENSE`.
