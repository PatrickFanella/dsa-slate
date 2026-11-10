# DSA Slate

DSA Slate is a reserved project repository. It contains only the starting
scaffold: a Go backend, a Vite React TypeScript frontend, a PostgreSQL schema
and Docker Compose files. No slate or organizing features are built yet.

## What is here today

- **Backend** (`backend/`): a Go 1.26 module whose `cmd/app` program prints a
  greeting and exits. It does not serve an HTTP API yet.
- **Frontend** (`web/`): a React 19 page built with Vite and Tailwind CSS that
  shows a placeholder card.
- **Database** (`migrations/0001_init.sql`): one `app_metadata` key-value table.
- **Compose stack** (`docker-compose.yml`): PostgreSQL 17, the backend image
  and the frontend served by nginx.

## Quick start

Requirements: Go 1.26, Node.js with pnpm 10, and Docker with Compose.

```bash
git clone https://github.com/subculture-collective/dsa-slate.git
cd dsa-slate
make deps                          # install frontend dependencies
pnpm --dir web run dev             # frontend dev server on port 5173
cd backend && go run ./cmd/app     # run the backend program
```

Run every check:

```bash
make verify
```

`make verify` installs dependencies, checks Go formatting and frontend types,
runs `go vet`, ESLint, Go tests and Vitest, builds both parts and validates the
Compose file. `make lint`, `make test` and `make build` run those steps on
their own.

To start the containers:

```bash
cp .env.example .env
docker compose up --build
```

## Configuration

`.env.example` lists the settings Compose reads:

- `PROJECT_NAME`
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_PORT`

`docker-compose.yml` also accepts `API_PORT` (default 8080) and `WEB_PORT`
(default 5173). Set your own database password; the Compose default is a
placeholder.

## Deploy

`backend/Dockerfile` builds a static Go binary on Alpine, and `web/Dockerfile`
builds the frontend and serves it from nginx on port 80. There is nothing to
deploy beyond the scaffold yet.

## Contributing

- Bugs and feature requests: [GitHub issues](https://github.com/subculture-collective/dsa-slate/issues).
- Code: open a pull request on GitHub against `main`. CI runs `make verify` on
  every pull request.

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

GNU General Public License, version 3 only (`GPL-3.0-only`). See
[LICENSE](LICENSE).
