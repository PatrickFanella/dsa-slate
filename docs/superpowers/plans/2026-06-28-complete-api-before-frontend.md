# Complete API Before Frontend Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Finish a thin, stable, source-linked read API that can power the frontend before deeper frontend feature work or visual polish.

**Architecture:** Keep one PostgreSQL-backed Go API as the source of truth, with the existing aggregate `GET /api/directory` preserved for the frontend. Add resource-specific read endpoints and small query filters only where they reduce frontend coupling; do not add writes, auth, raw voter addresses, or enterprise API machinery yet.

**Tech Stack:** Go standard library HTTP server, pgx/PostgreSQL, SQL migrations and seed fixtures, Vite React TypeScript client, Go tests, Vitest, Docker Compose, `make verify`.

---

## Scope

This plan builds a **complete vertical-slice read API**, not a full admin platform.

In scope:

- Keep `GET /healthz` and `GET /api/directory` working.
- Add individual collection endpoints for chapters, elections, races, candidates, voter guides, endorsements, ballot measures, and measure positions.
- Add simple, documented query filters that match current UI needs.
- Add backend contract tests for route shape, empty arrays, filters, method handling, and JSON errors.
- Add a lightweight API contract document for frontend work.
- Add frontend API client functions/tests for the finalized endpoints without redesigning UI.
- Verify seed data still passes source-link/confidence expectations.

Out of scope:

- Admin write API.
- Auth/session/RBAC.
- Public submissions.
- Address, voter, precinct, or district lookup.
- Pagination unless seed data grows large enough to require it.
- Major frontend redesign or polish.

## Current baseline

Existing useful pieces:

- `backend/internal/httpapi/server.go` already serves `GET /healthz` and `GET /api/directory`.
- `backend/internal/directory/store.go` defines the aggregate DTOs.
- `backend/internal/directory/postgres.go` reads all existing tables.
- `web/src/data/api.ts` already validates `GET /api/directory` responses.
- `migrations/0001_init.sql` and `seeds/california.sql` already model the main public data.

The remaining risk is that the frontend only has one aggregate API shape. Before building more UI, stabilize the backend contract and filtering boundaries.

## Proposed API contract

Keep aggregate endpoint:

```http
GET /api/directory
```

Add collection endpoints:

```http
GET /api/chapters
GET /api/elections
GET /api/races?electionId=2026-ca-primary&jurisdiction=California&level=statewide
GET /api/candidates
GET /api/voter-guides?chapterId=dsa-la&electionId=2026-ca-primary
GET /api/endorsements?chapterId=dsa-la&electionId=2026-ca-primary&raceId=ca-governor-2026
GET /api/ballot-measures?electionId=2026-ca-primary
GET /api/measure-positions?chapterId=dsa-la&measureId=ca-prop-example
```

Error shape:

```json
{"error":"not_found"}
```

```json
{"error":"invalid_query","message":"level must be one of federal, statewide, state_legislature, local, unknown"}
```

## Files

- Modify `backend/internal/directory/store.go` — extend `Store` with collection methods and filter structs.
- Modify `backend/internal/directory/memory.go` — implement filtered collection methods for tests.
- Modify `backend/internal/directory/postgres.go` — implement collection query methods and SQL filters.
- Modify `backend/internal/httpapi/server.go` — route collection endpoints, parse query filters, write consistent errors.
- Modify `backend/internal/httpapi/server_test.go` — add endpoint/filter/error contract tests.
- Create `backend/internal/directory/filter_test.go` — test memory-store filter semantics separately from HTTP.
- Modify `web/src/data/api.ts` — add typed fetch helpers for individual collection endpoints.
- Modify `web/src/data/api.test.ts` — test helper URLs, validation, and error behavior.
- Create `docs/reference/api.md` — document stable read API contract and non-goals.
- Modify `README.md` — link API contract and clarify recommended backend-first workflow.

## Tasks

### Task 1: Lock the backend read contract in docs

**Files:**

- Create: `docs/reference/api.md`
- Modify: `README.md`

- [ ] **Step 1: Create the API contract document**

Create `docs/reference/api.md` with:

```markdown
# DSA Slate Read API

The API is a public, read-only JSON API for source-linked electoral guide data.

## Non-goals

- No write/admin endpoints yet.
- No authentication yet.
- No raw voter addresses or address lookup.
- No personally identifying voter data.

## Health

`GET /healthz`

```json
{"ok":true,"service":"dsa-slate"}
```

## Aggregate directory

`GET /api/directory`

Returns all public directory collections:

- `chapters`
- `elections`
- `races`
- `candidates`
- `voterGuides`
- `endorsements`
- `ballotMeasures`
- `measurePositions`

All empty collections are JSON arrays, not `null`.

## Collection endpoints

- `GET /api/chapters`
- `GET /api/elections`
- `GET /api/races?electionId=<id>&jurisdiction=<name>&level=<level>`
- `GET /api/candidates`
- `GET /api/voter-guides?chapterId=<id>&electionId=<id>`
- `GET /api/endorsements?chapterId=<id>&electionId=<id>&raceId=<id>`
- `GET /api/ballot-measures?electionId=<id>`
- `GET /api/measure-positions?chapterId=<id>&measureId=<id>`

## Error shape

```json
{"error":"not_found"}
```

```json
{"error":"invalid_query","message":"level must be one of federal, statewide, state_legislature, local, unknown"}
```

## Source-link requirements

Endorsements, voter guides, and measure positions must include `sourceUrl` and `confidence`.
Confidence values are:

- `official_source`
- `official_social_source`
- `campaign_source`
- `press_source`
- `needs_review`
```

- [ ] **Step 2: Link the API contract from README**

Add this under the README development/API section:

```markdown
Read API contract: `docs/reference/api.md`.

Recommended feature order: stabilize the read API and seed data first, then build frontend screens against that contract. Avoid frontend-only data shapes that diverge from the Go API DTOs.
```

- [ ] **Step 3: Verify docs render as plain Markdown**

Run:

```bash
make verify
```

Expected: either PASS, or only pre-existing unrelated failures. If it fails because the new Markdown broke formatting/linting, fix the Markdown before continuing.

### Task 2: Extend the store interface with collection methods

**Files:**

- Modify: `backend/internal/directory/store.go`
- Modify: `backend/internal/directory/memory.go`
- Create: `backend/internal/directory/filter_test.go`

- [ ] **Step 1: Add filter structs and collection methods**

In `backend/internal/directory/store.go`, change the interface from:

```go
type Store interface {
	Directory(ctx context.Context) (DirectoryData, error)
}
```

to:

```go
type RaceFilter struct {
	ElectionID   string
	Jurisdiction string
	Level        string
}

type VoterGuideFilter struct {
	ChapterID  string
	ElectionID string
}

type EndorsementFilter struct {
	ChapterID  string
	ElectionID string
	RaceID     string
}

type BallotMeasureFilter struct {
	ElectionID string
}

type MeasurePositionFilter struct {
	ChapterID string
	MeasureID string
}

type Store interface {
	Directory(ctx context.Context) (DirectoryData, error)
	Chapters(ctx context.Context) ([]Chapter, error)
	Elections(ctx context.Context) ([]Election, error)
	Races(ctx context.Context, filter RaceFilter) ([]Race, error)
	Candidates(ctx context.Context) ([]Candidate, error)
	VoterGuides(ctx context.Context, filter VoterGuideFilter) ([]VoterGuide, error)
	Endorsements(ctx context.Context, filter EndorsementFilter) ([]Endorsement, error)
	BallotMeasures(ctx context.Context, filter BallotMeasureFilter) ([]BallotMeasure, error)
	MeasurePositions(ctx context.Context, filter MeasurePositionFilter) ([]MeasurePosition, error)
}
```

- [ ] **Step 2: Implement memory-store methods**

In `backend/internal/directory/memory.go`, add methods that filter the normalized in-memory data. Use exact equality and ignore empty filter fields.

Example for races:

```go
func (s *MemoryStore) Races(ctx context.Context, filter RaceFilter) ([]Race, error) {
	data := NormalizeDirectoryData(s.Data)
	races := make([]Race, 0)
	for _, race := range data.Races {
		if filter.ElectionID != "" && race.ElectionID != filter.ElectionID {
			continue
		}
		if filter.Jurisdiction != "" && race.Jurisdiction != filter.Jurisdiction {
			continue
		}
		if filter.Level != "" && race.Level != filter.Level {
			continue
		}
		races = append(races, race)
	}
	return races, nil
}
```

Implement the remaining methods with the same pattern.

- [ ] **Step 3: Write memory filter tests**

Create `backend/internal/directory/filter_test.go` with focused tests like:

```go
package directory

import (
	"context"
	"testing"
)

func TestMemoryStoreFiltersRaces(t *testing.T) {
	store := NewMemoryStore(DirectoryData{Races: []Race{
		{ID: "race-1", ElectionID: "election-1", Office: "Mayor", Level: "local", Jurisdiction: "City"},
		{ID: "race-2", ElectionID: "election-2", Office: "Governor", Level: "statewide", Jurisdiction: "California"},
	}})

	got, err := store.Races(context.Background(), RaceFilter{ElectionID: "election-2", Level: "statewide"})
	if err != nil {
		t.Fatalf("Races returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "race-2" {
		t.Fatalf("Races = %#v, want race-2", got)
	}
}
```

- [ ] **Step 4: Run backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

### Task 3: Add HTTP collection routes and query validation

**Files:**

- Modify: `backend/internal/httpapi/server.go`
- Modify: `backend/internal/httpapi/server_test.go`

- [ ] **Step 1: Add route cases**

In `Server.ServeHTTP`, add `GET` routes for:

```go
case "/api/chapters":
case "/api/elections":
case "/api/races":
case "/api/candidates":
case "/api/voter-guides":
case "/api/endorsements":
case "/api/ballot-measures":
case "/api/measure-positions":
```

Each route should reject non-GET with the existing `methodNotAllowed` helper.

- [ ] **Step 2: Add invalid query error helper**

Add this helper to `server.go`:

```go
func (s *Server) invalidQuery(w http.ResponseWriter, message string) {
	s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_query", "message": message})
}
```

- [ ] **Step 3: Validate race level values**

Add:

```go
func validRaceLevel(level string) bool {
	switch level {
	case "", "federal", "statewide", "state_legislature", "local", "unknown":
		return true
	default:
		return false
	}
}
```

Use it for `/api/races?level=...` and return `400` with `invalid_query` if invalid.

- [ ] **Step 4: Add endpoint handler tests**

Extend `server_test.go` with tests that assert:

- `GET /api/chapters` returns a JSON array.
- `GET /api/races?electionId=election-2&level=statewide` returns only matching races.
- `GET /api/races?level=bad` returns `400` and `{"error":"invalid_query"}`.
- `POST /api/chapters` returns `405`.

Use `directory.NewMemoryStore(...)` fixtures, not Postgres.

- [ ] **Step 5: Run backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

### Task 4: Implement PostgreSQL collection methods

**Files:**

- Modify: `backend/internal/directory/postgres.go`

- [ ] **Step 1: Export existing private query methods through interface methods**

The file already has private methods such as `chapters(ctx)`, `elections(ctx)`, and similar. Add exported wrappers:

```go
func (s *PostgresStore) Chapters(ctx context.Context) ([]Chapter, error) {
	return s.chapters(ctx)
}

func (s *PostgresStore) Elections(ctx context.Context) ([]Election, error) {
	return s.elections(ctx)
}

func (s *PostgresStore) Candidates(ctx context.Context) ([]Candidate, error) {
	return s.candidates(ctx)
}
```

- [ ] **Step 2: Add filtered SQL methods**

Implement filtered methods for races, voter guides, endorsements, ballot measures, and measure positions. Keep deterministic `ORDER BY id`.

Preferred simple SQL pattern:

```sql
WHERE ($1 = '' OR election_id = $1)
  AND ($2 = '' OR jurisdiction = $2)
  AND ($3 = '' OR level = $3)
ORDER BY id
```

For endorsements filtered by `electionId`, join through `races`:

```sql
SELECT e.id, e.candidate_id, e.race_id, e.chapter_id, e.voter_guide_id,
       e.endorsement_type, e.source_url, e.source_date, e.confidence, e.notes
FROM endorsements e
JOIN races r ON r.id = e.race_id
WHERE ($1 = '' OR e.chapter_id = $1)
  AND ($2 = '' OR r.election_id = $2)
  AND ($3 = '' OR e.race_id = $3)
ORDER BY e.id
```

- [ ] **Step 3: Keep required-field behavior strict**

Do not use `COALESCE` to hide broken seed/schema data. Continue using the existing `requiredString`/optional helper pattern so broken required fields fail loudly.

- [ ] **Step 4: Run backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

### Task 5: Add frontend API helpers without UI churn

**Files:**

- Modify: `web/src/data/api.ts`
- Modify: `web/src/data/api.test.ts`

- [ ] **Step 1: Add collection fetch helpers**

In `web/src/data/api.ts`, keep `fetchDirectoryData` unchanged and add typed helpers. Use a local URL builder so query params stay encoded.

```ts
function withQuery(endpoint: string, params: Record<string, string | undefined>): string {
  const search = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (value) {
      search.set(key, value);
    }
  }

  const query = search.toString();
  return query ? `${endpoint}?${query}` : endpoint;
}
```

Add helpers such as:

```ts
export async function fetchRaces(
  fetcher: DirectoryFetch = fetch,
  filters: { electionId?: string; jurisdiction?: string; level?: string } = {},
) {
  const response = await fetcher(withQuery('/api/races', filters), { headers: { Accept: 'application/json' } });
  if (!response.ok) {
    throw new Error(`Races API request failed with status ${response.status}`);
  }
  return (await response.json()) as DirectoryData['races'];
}
```

Add equivalent helpers for the other collection endpoints.

- [ ] **Step 2: Add frontend helper tests**

In `web/src/data/api.test.ts`, add tests that assert:

- `fetchRaces(fakeFetch, { electionId: 'election 1', level: 'local' })` calls `/api/races?electionId=election+1&level=local`.
- non-OK collection responses throw a useful error.
- `fetchDirectoryData` behavior remains unchanged.

- [ ] **Step 3: Run web tests**

Run:

```bash
pnpm --dir web run test
```

Expected: PASS.

### Task 6: End-to-end verification and handoff to frontend

**Files:**

- Modify only if verification finds issues: `README.md`, `docs/reference/api.md`, or tests.

- [ ] **Step 1: Run full repo verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 2: Smoke test the API with seed data**

Run:

```bash
make seed
make dev-api
```

In a second shell:

```bash
make api-smoke
curl -fsS 'http://localhost:8080/api/races?level=statewide'
curl -fsS 'http://localhost:8080/api/endorsements?chapterId=dsa-la'
```

Expected: successful JSON responses; empty result sets are `[]`, not `null`.

- [ ] **Step 3: Confirm privacy boundary**

Search the changed API, docs, and seed files for address-like fields:

```bash
rg -n "address|street|zip|zipcode|precinct|voter" backend web/src/data docs/reference seeds migrations
```

Expected: no raw voter address storage or lookup introduced. If a documentation mention appears, confirm it is a non-goal/privacy warning.

- [ ] **Step 4: Frontend handoff decision**

Once this passes, frontend work can proceed against either:

1. `GET /api/directory` for simple aggregate page rendering, or
2. collection helpers for screens that need server-side filtering.

Do not add frontend-only data fields unless they are also added to `docs/reference/api.md`, Go DTOs, seed data, and TypeScript schema.

## Acceptance criteria

- `make verify` passes.
- `docs/reference/api.md` documents every public read route.
- Existing `GET /api/directory` behavior is preserved.
- Collection endpoints return frontend-compatible camelCase JSON.
- Filtered endpoints support current UI-level filtering without introducing writes/auth/address lookup.
- API errors use consistent JSON shapes.
- Frontend has typed fetch helpers and tests for the finalized API contract.
- No raw voter addresses or address-like user data are stored or exposed.
