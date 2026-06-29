# Draft Publish CRUD Foundation Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add a safe admin-write foundation for chapters and endorsements using draft/publish state, review metadata, and an audit trail before building user-facing forms.

**Architecture:** Keep existing public read endpoints stable and publish-filtered. Add separate admin write tables/endpoints for draft records, then promote reviewed drafts into the public read tables. The first slice uses a local admin token from environment config instead of full user accounts, because accounts/address storage require a later privacy/security phase.

**Tech Stack:** Go HTTP API, PostgreSQL, pgxpool, React TypeScript later, Makefile verification, existing data-quality checks.

---

## Requirements summary

User wants forms to add chapters and endorsements, but prefers working toward CRUD on the API first. Approved direction: **draft/publish writes**.

## Scope boundaries

### In scope for this foundation

- Admin-only create/update/list/get draft records for:
  - chapters
  - endorsements
- Draft states:
  - `draft`
  - `needs_review`
  - `published`
  - `archived`
- Publish operation that writes reviewed drafts into existing public tables.
- Minimal admin protection with `ADMIN_API_TOKEN` bearer token.
- Audit rows for create/update/publish/archive attempts.
- Server-side validation for required fields, enum values, source URLs, and references.
- Docs for future form integration.

### Explicitly out of scope

- No public submissions.
- No user accounts yet.
- No raw voter address storage.
- No Google sign-in/OAuth flow.
- No React forms in this slice.
- No delete endpoint; use `archived` instead.
- No automated scraping/importing endorsements.

## File map

- Modify: `migrations/0001_init.sql` — add draft/audit tables and status constraints.
- Create: `backend/internal/admin/store.go` — admin DTOs/store interface.
- Create: `backend/internal/admin/postgres.go` — Postgres writes/publish operations.
- Create: `backend/internal/admin/validation.go` — request validation helpers.
- Create: `backend/internal/admin/postgres_test.go` — DB-free validation tests or pgxpool integration tests if existing harness appears.
- Modify: `backend/internal/httpapi/server.go` — route admin endpoints behind bearer token.
- Modify: `backend/internal/httpapi/server_test.go` — admin auth/validation route tests.
- Modify: `backend/internal/app/app.go` — wire admin token and store into HTTP server if needed.
- Modify: `.env.example` — add `ADMIN_API_TOKEN=` placeholder.
- Modify: `docs/reference/api.md` — document admin draft endpoints separately from public API.
- Modify: `docs/reference/editorial-data.md` — document draft/publish workflow.
- Modify: `README.md` — note local admin token and no public submissions.

---

## Data model contract

### New table: `admin_chapter_drafts`

```sql
CREATE TABLE IF NOT EXISTS admin_chapter_drafts (
  id text PRIMARY KEY,
  chapter_id text NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  name text NOT NULL,
  state text NOT NULL,
  website_url text,
  official_dsa_url text,
  coverage_notes text,
  group_email text,
  linktree_url text,
  facebook_url text,
  twitter_url text,
  bluesky_url text,
  instagram_url text,
  blog_url text,
  event_calendar_url text,
  email_list_url text,
  review_notes text,
  reviewed_at date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT admin_chapter_drafts_status_check CHECK (status IN ('draft', 'needs_review', 'published', 'archived'))
);
```

### New table: `admin_endorsement_drafts`

```sql
CREATE TABLE IF NOT EXISTS admin_endorsement_drafts (
  id text PRIMARY KEY,
  endorsement_id text NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  candidate_id text NOT NULL,
  race_id text NOT NULL,
  chapter_id text NOT NULL,
  voter_guide_id text,
  endorsement_type text NOT NULL,
  source_url text NOT NULL,
  source_date date,
  confidence text NOT NULL DEFAULT 'needs_review',
  notes text,
  review_notes text,
  reviewed_at date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT admin_endorsement_drafts_status_check CHECK (status IN ('draft', 'needs_review', 'published', 'archived')),
  CONSTRAINT admin_endorsement_drafts_type_check CHECK (endorsement_type IN ('endorsed', 'recommended', 'opposed', 'no_recommendation', 'write_in', 'coalition', 'unknown')),
  CONSTRAINT admin_endorsement_drafts_confidence_check CHECK (confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review'))
);
```

### New table: `admin_audit_events`

```sql
CREATE TABLE IF NOT EXISTS admin_audit_events (
  id bigserial PRIMARY KEY,
  actor text NOT NULL,
  action text NOT NULL,
  record_type text NOT NULL,
  record_id text NOT NULL,
  message text,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

---

## API contract

All admin endpoints require:

```http
Authorization: Bearer <ADMIN_API_TOKEN>
```

### Chapter drafts

- `GET /api/admin/chapter-drafts`
- `POST /api/admin/chapter-drafts`
- `GET /api/admin/chapter-drafts/{id}`
- `PUT /api/admin/chapter-drafts/{id}`
- `POST /api/admin/chapter-drafts/{id}/publish`
- `POST /api/admin/chapter-drafts/{id}/archive`

### Endorsement drafts

- `GET /api/admin/endorsement-drafts`
- `POST /api/admin/endorsement-drafts`
- `GET /api/admin/endorsement-drafts/{id}`
- `PUT /api/admin/endorsement-drafts/{id}`
- `POST /api/admin/endorsement-drafts/{id}/publish`
- `POST /api/admin/endorsement-drafts/{id}/archive`

### Error shape

```json
{"error":"unauthorized"}
```

```json
{"error":"validation_failed","messages":["sourceUrl must be an absolute http(s) URL"]}
```

---

## Task 1: Add migration tables

**Files:**
- Modify: `migrations/0001_init.sql`

- [ ] **Step 1: Add draft/audit tables**

Append the SQL from “Data model contract” to `migrations/0001_init.sql` after existing table/alter sections.

- [ ] **Step 2: Verify migration applies**

Run:

```bash
make migrate
```

Expected: migration succeeds; repeated runs are idempotent.

---

## Task 2: Add admin DTOs and validation

**Files:**
- Create: `backend/internal/admin/store.go`
- Create: `backend/internal/admin/validation.go`
- Create: `backend/internal/admin/validation_test.go`

- [ ] **Step 1: Create DTOs and interface**

Create `backend/internal/admin/store.go` with:

```go
package admin

import "context"

type DraftStatus string

const (
	StatusDraft      DraftStatus = "draft"
	StatusNeedsReview DraftStatus = "needs_review"
	StatusPublished  DraftStatus = "published"
	StatusArchived   DraftStatus = "archived"
)

type ChapterDraft struct {
	ID               string      `json:"id"`
	ChapterID        string      `json:"chapterId"`
	Status           DraftStatus `json:"status"`
	Name             string      `json:"name"`
	State            string      `json:"state"`
	WebsiteURL       string      `json:"websiteUrl,omitempty"`
	OfficialDSAURL   string      `json:"officialDsaUrl,omitempty"`
	CoverageNotes    string      `json:"coverageNotes,omitempty"`
	GroupEmail       string      `json:"groupEmail,omitempty"`
	ReviewNotes      string      `json:"reviewNotes,omitempty"`
	ReviewedAt       string      `json:"reviewedAt,omitempty"`
}

type EndorsementDraft struct {
	ID              string      `json:"id"`
	EndorsementID   string      `json:"endorsementId"`
	Status          DraftStatus `json:"status"`
	CandidateID     string      `json:"candidateId"`
	RaceID          string      `json:"raceId"`
	ChapterID       string      `json:"chapterId"`
	VoterGuideID    string      `json:"voterGuideId,omitempty"`
	EndorsementType string      `json:"endorsementType"`
	SourceURL       string      `json:"sourceUrl"`
	SourceDate      string      `json:"sourceDate,omitempty"`
	Confidence      string      `json:"confidence"`
	Notes           string      `json:"notes,omitempty"`
	ReviewNotes     string      `json:"reviewNotes,omitempty"`
	ReviewedAt      string      `json:"reviewedAt,omitempty"`
}

type Store interface {
	ListChapterDrafts(ctx context.Context) ([]ChapterDraft, error)
	UpsertChapterDraft(ctx context.Context, draft ChapterDraft, actor string) (ChapterDraft, error)
	PublishChapterDraft(ctx context.Context, id string, actor string) error
	ArchiveChapterDraft(ctx context.Context, id string, actor string) error
	ListEndorsementDrafts(ctx context.Context) ([]EndorsementDraft, error)
	UpsertEndorsementDraft(ctx context.Context, draft EndorsementDraft, actor string) (EndorsementDraft, error)
	PublishEndorsementDraft(ctx context.Context, id string, actor string) error
	ArchiveEndorsementDraft(ctx context.Context, id string, actor string) error
}
```

- [ ] **Step 2: Add validation tests**

Create `backend/internal/admin/validation_test.go` with tests for missing chapter fields, invalid URL, invalid status, invalid endorsement type, invalid confidence, and missing source URL.

- [ ] **Step 3: Implement validation**

Create `backend/internal/admin/validation.go` with `ValidateChapterDraft(draft ChapterDraft) []string` and `ValidateEndorsementDraft(draft EndorsementDraft) []string`. Reuse only standard library helpers: `net/url`, `strings`, and `time`.

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/admin -count=1
```

Expected: PASS.

---

## Task 3: Implement Postgres admin store

**Files:**
- Create: `backend/internal/admin/postgres.go`
- Create: `backend/internal/admin/postgres_test.go`

- [ ] **Step 1: Implement draft upserts and list methods**

Create `NewPostgresStore(pool *pgxpool.Pool) *PostgresStore` and methods matching `admin.Store`.

- [ ] **Step 2: Implement publish methods**

Publishing chapter drafts upserts into `chapters`. Publishing endorsement drafts upserts into `endorsements` and requires referenced `candidate_id`, `race_id`, and `chapter_id` to already exist.

- [ ] **Step 3: Add audit writes**

Every create/update/publish/archive writes one row to `admin_audit_events` with actor, action, record type, record ID, and a short message.

- [ ] **Step 4: Run backend tests**

```bash
cd backend && go test ./...
```

Expected: PASS.

---

## Task 4: Add admin HTTP routing behind bearer auth

**Files:**
- Modify: `backend/internal/httpapi/server.go`
- Modify: `backend/internal/httpapi/server_test.go`
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: Extend server constructor**

Add optional admin store/token fields while preserving existing `NewServer(directory.Store)` tests. Use a second constructor like:

```go
func NewServerWithAdmin(store directory.Store, adminStore admin.Store, adminToken string) *Server
```

- [ ] **Step 2: Add bearer auth helper**

Return `401 {"error":"unauthorized"}` when token is missing or wrong.

- [ ] **Step 3: Add route tests**

In `server_test.go`, add tests for:

- missing admin token returns 401
- wrong token returns 401
- valid token can create a chapter draft
- invalid payload returns `validation_failed`

- [ ] **Step 4: Wire app config**

In `backend/internal/app/app.go`, read `ADMIN_API_TOKEN` from environment and pass it to the admin-enabled server constructor when non-empty.

- [ ] **Step 5: Run backend tests**

```bash
cd backend && go test ./...
```

Expected: PASS.

---

## Task 5: Docs/config verification

**Files:**
- Modify: `.env.example`
- Modify: `docs/reference/api.md`
- Modify: `docs/reference/editorial-data.md`
- Modify: `README.md`

- [ ] **Step 1: Add env placeholder**

Add to `.env.example`:

```env
ADMIN_API_TOKEN=
```

- [ ] **Step 2: Document admin endpoints**

Document that admin endpoints are local/admin-only, token-protected, draft/publish gated, and not public submissions.

- [ ] **Step 3: Run full verification**

```bash
make verify
make data-quality
```

Expected: both pass.

- [ ] **Step 4: Commit**

```bash
git add migrations/0001_init.sql backend/internal/admin backend/internal/httpapi/server.go backend/internal/httpapi/server_test.go backend/internal/app/app.go .env.example docs/reference/api.md docs/reference/editorial-data.md README.md docs/superpowers/plans/2026-06-29-draft-publish-crud-foundation.md
git commit -m "feat: add draft publish admin api foundation"
```

---

## Self-review

- Spec coverage: supports future chapter/endorsement forms through a safe write API foundation.
- Security: token-gated admin endpoints, no public submissions, no raw addresses.
- Data integrity: draft/publish workflow prevents immediate public mutation from unreviewed input.
- YAGNI: no user accounts, no OAuth, no React forms, no scraper.
- Verification: backend tests, full `make verify`, and `make data-quality` are required.
