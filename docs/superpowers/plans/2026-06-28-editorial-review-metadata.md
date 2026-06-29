# Editorial Review Metadata Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add explicit editorial review metadata so source-linked records can document manual review and “no visible source date” exceptions without inventing publication dates.

**Architecture:** Extend only public source-bearing records (`voter_guides`, `endorsements`, `measure_positions`) with nullable `reviewed_at` and `review_notes`. Carry those fields through Go DTOs, Postgres reads, API JSON, TypeScript schema/static fallback, and the data-quality checker. Keep the API read-only and keep write/admin UI out of scope.

**Tech Stack:** Go backend, PostgreSQL migrations/seeds, pgxpool read store, Vite React TypeScript validation, Makefile verification.

---

## Why this slice now

`make data-quality` currently passes with one warning:

```text
WARN voterGuides[california-dsa-2026-primary-guide] missing publishedAt
```

The source page at `https://www.californiadsa.org/voterguide` is live and source-linked, but no visible publication date was found during planning. The next slice should not invent `publishedAt`. Instead, it should let editors document that the record was manually reviewed and that no visible source date was present.

## Scope boundaries

### In scope

- Add `reviewed_at` and `review_notes` to voter guides, endorsements, and measure positions.
- Show those fields in read API JSON as `reviewedAt` and `reviewNotes`.
- Add TypeScript schema validation for optional ISO `reviewedAt`.
- Update California seed/static fallback with review metadata for the CA DSA statewide guide.
- Teach the data-quality report to suppress missing date warnings only when review metadata documents the exception.

### Out of scope

- No write/admin API endpoints.
- No auth, roles, sessions, or admin UI.
- No raw voter addresses or voter lookup.
- No crawler/scraper/link checker.
- No source truth verification beyond human-written review metadata.

## File map

- Modify: `migrations/0001_init.sql` — add idempotent columns with `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`.
- Modify: `seeds/california.sql` — add review metadata values to source-bearing seed records.
- Modify: `backend/internal/directory/store.go` — add DTO fields.
- Modify: `backend/internal/directory/postgres.go` — select/scan review fields.
- Modify: `backend/internal/httpapi/server_test.go` — assert camelCase review JSON.
- Modify: `backend/internal/directory/quality.go` — use review metadata to document missing source dates.
- Modify: `backend/internal/directory/quality_test.go` — lock warning suppression and malformed review-date warnings.
- Modify: `web/src/data/schema.ts` — add fields and validate `reviewedAt` dates.
- Modify: `web/src/data/california.ts` — mirror SQL seed review metadata.
- Modify: `web/src/data/schema.test.ts` — lock static fallback schema and reviewedAt validation.
- Modify: `docs/reference/editorial-data.md` — document review metadata semantics.

---

## Task 1: Extend source-bearing DTOs and frontend schema

**Files:**
- Modify: `backend/internal/directory/store.go`
- Modify: `web/src/data/schema.ts`
- Modify: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add Go DTO fields**

In `backend/internal/directory/store.go`, update the three source-bearing structs:

```go
type VoterGuide struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ChapterID   string `json:"chapterId"`
	ElectionID  string `json:"electionId,omitempty"`
	SourceURL   string `json:"sourceUrl"`
	Confidence  string `json:"confidence"`
	PublishedAt string `json:"publishedAt,omitempty"`
	Summary     string `json:"summary,omitempty"`
	ReviewedAt  string `json:"reviewedAt,omitempty"`
	ReviewNotes string `json:"reviewNotes,omitempty"`
}

type Endorsement struct {
	ID              string `json:"id"`
	CandidateID     string `json:"candidateId"`
	RaceID          string `json:"raceId"`
	ChapterID       string `json:"chapterId"`
	VoterGuideID    string `json:"voterGuideId,omitempty"`
	EndorsementType string `json:"endorsementType"`
	SourceURL       string `json:"sourceUrl"`
	SourceDate      string `json:"sourceDate,omitempty"`
	Confidence      string `json:"confidence"`
	Notes           string `json:"notes,omitempty"`
	ReviewedAt      string `json:"reviewedAt,omitempty"`
	ReviewNotes     string `json:"reviewNotes,omitempty"`
}

type MeasurePosition struct {
	ID          string `json:"id"`
	MeasureID   string `json:"measureId"`
	ChapterID   string `json:"chapterId"`
	Position    string `json:"position"`
	SourceURL   string `json:"sourceUrl"`
	Confidence  string `json:"confidence"`
	ReviewedAt  string `json:"reviewedAt,omitempty"`
	ReviewNotes string `json:"reviewNotes,omitempty"`
}
```

- [ ] **Step 2: Add TypeScript fields**

In `web/src/data/schema.ts`, update the matching types:

```ts
export type VoterGuide = {
  id: string;
  title: string;
  chapterId: string;
  electionId?: string;
  sourceUrl: string;
  confidence: ConfidenceLevel;
  publishedAt?: string;
  summary?: string;
  reviewedAt?: string;
  reviewNotes?: string;
};

export type Endorsement = {
  id: string;
  candidateId: string;
  raceId: string;
  chapterId: string;
  voterGuideId?: string;
  endorsementType: EndorsementType;
  sourceUrl: string;
  sourceDate?: string;
  confidence: ConfidenceLevel;
  notes?: string;
  reviewedAt?: string;
  reviewNotes?: string;
};

export type MeasurePosition = {
  id: string;
  measureId: string;
  chapterId: string;
  position: 'yes' | 'no' | 'no_recommendation' | 'opposed' | 'unknown';
  sourceUrl: string;
  confidence: ConfidenceLevel;
  reviewedAt?: string;
  reviewNotes?: string;
};
```

Then add review-date validation inside `validateDirectoryData`:

```ts
validateOptionalIsoDate(`voterGuides[${guide.id}]`, 'reviewedAt', guide.reviewedAt, errors);
validateOptionalIsoDate(`endorsements[${endorsement.id}]`, 'reviewedAt', endorsement.reviewedAt, errors);
validateOptionalIsoDate(`measurePositions[${position.id}]`, 'reviewedAt', position.reviewedAt, errors);
```

Place those calls in the existing loops for voter guides, endorsements, and measure positions.

- [ ] **Step 3: Add failing frontend validation test**

In `web/src/data/schema.test.ts`, append this test to `describe('directory schema helpers', ...)`:

```ts
it('validates optional editorial reviewedAt dates', () => {
  const data: DirectoryData = {
    chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
    elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
    races: [
      {
        id: 'ca-48-us-house-2026-primary',
        electionId: '2026-ca-primary',
        office: 'U.S. House',
        level: 'federal',
        jurisdiction: 'CA-48',
      },
    ],
    candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras' }],
    voterGuides: [
      {
        id: 'guide-1',
        title: 'Guide',
        chapterId: 'dsa-san-diego',
        sourceUrl: 'https://example.org/guide',
        confidence: 'official_source',
        reviewedAt: 'June 28, 2026',
      },
    ],
    endorsements: [
      {
        id: 'endorsement-1',
        candidateId: 'corinna-contreras',
        raceId: 'ca-48-us-house-2026-primary',
        chapterId: 'dsa-san-diego',
        endorsementType: 'recommended',
        sourceUrl: 'https://example.org/guide',
        confidence: 'official_source',
        reviewedAt: '2026-02-31',
      },
    ],
    ballotMeasures: [{ id: 'measure-a', electionId: '2026-ca-primary', title: 'Measure A', jurisdiction: 'San Diego, CA' }],
    measurePositions: [
      {
        id: 'measure-position-1',
        measureId: 'measure-a',
        chapterId: 'dsa-san-diego',
        position: 'yes',
        sourceUrl: 'https://example.org/guide',
        confidence: 'official_source',
        reviewedAt: 'not-a-date',
      },
    ],
  };

  expect(validateDirectoryData(data)).toEqual(
    expect.arrayContaining([
      'voterGuides[guide-1].reviewedAt must be YYYY-MM-DD',
      'endorsements[endorsement-1].reviewedAt must be YYYY-MM-DD',
      'measurePositions[measure-position-1].reviewedAt must be YYYY-MM-DD',
    ]),
  );
});
```

- [ ] **Step 4: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected after implementation: PASS.

---

## Task 2: Add migration and seed/static review metadata

**Files:**
- Modify: `migrations/0001_init.sql`
- Modify: `seeds/california.sql`
- Modify: `web/src/data/california.ts`

- [ ] **Step 1: Add idempotent schema columns**

In `migrations/0001_init.sql`, add these columns to the `CREATE TABLE IF NOT EXISTS` definitions:

```sql
-- voter_guides table
reviewed_at date,
review_notes text,

-- endorsements table
reviewed_at date,
review_notes text,

-- measure_positions table
reviewed_at date,
review_notes text,
```

Also add idempotent alters after the existing `ALTER TABLE voter_guides ADD COLUMN IF NOT EXISTS confidence...` block:

```sql
ALTER TABLE voter_guides
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;

ALTER TABLE endorsements
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;

ALTER TABLE measure_positions
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;
```

- [ ] **Step 2: Update SQL seed values**

In `seeds/california.sql`, expand the voter guide insert columns:

```sql
INSERT INTO voter_guides (id, title, chapter_id, election_id, source_url, confidence, published_at, summary, reviewed_at, review_notes)
```

Set these values for `california-dsa-2026-primary-guide`:

```sql
DATE '2026-06-28',
'Manual review found no visible publication date on the public source page; do not invent published_at.'
```

Set these values for `dsa-san-diego-2026-primary-guide`:

```sql
DATE '2026-06-28',
'Manual review confirmed visible guide date and public source URL.'
```

Update the `ON CONFLICT` block:

```sql
reviewed_at = EXCLUDED.reviewed_at,
review_notes = EXCLUDED.review_notes;
```

For `endorsements`, expand columns and values with:

```sql
reviewed_at,
review_notes
```

Use for both current endorsements:

```sql
DATE '2026-06-28',
'Manual review confirmed source URL, source date, chapter, candidate, race, and recommendation wording.'
```

Update the endorsement `ON CONFLICT` block with `reviewed_at` and `review_notes` assignments.

For `measure_positions`, expand columns and values with:

```sql
reviewed_at,
review_notes
```

Use:

```sql
DATE '2026-06-28',
'Manual review confirmed source URL, chapter, ballot measure, and position wording.'
```

Update the measure-position `ON CONFLICT` block with `reviewed_at` and `review_notes` assignments.

- [ ] **Step 3: Mirror static fallback data**

In `web/src/data/california.ts`, add matching fields.

For `california-dsa-2026-primary-guide`:

```ts
reviewedAt: '2026-06-28',
reviewNotes: 'Manual review found no visible publication date on the public source page; do not invent publishedAt.',
```

For `dsa-san-diego-2026-primary-guide`:

```ts
reviewedAt: '2026-06-28',
reviewNotes: 'Manual review confirmed visible guide date and public source URL.',
```

For both endorsements:

```ts
reviewedAt: '2026-06-28',
reviewNotes: 'Manual review confirmed source URL, source date, chapter, candidate, race, and recommendation wording.',
```

For the measure position:

```ts
reviewedAt: '2026-06-28',
reviewNotes: 'Manual review confirmed source URL, chapter, ballot measure, and position wording.',
```

- [ ] **Step 4: Run frontend seed validation**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

---

## Task 3: Carry review metadata through Postgres reads and API JSON

**Files:**
- Modify: `backend/internal/directory/postgres.go`
- Modify: `backend/internal/httpapi/server_test.go`

- [ ] **Step 1: Update Postgres voter guide scan**

In `backend/internal/directory/postgres.go`, change the voter guide query to:

```go
SELECT id, title, chapter_id, election_id, source_url, confidence, published_at, summary, reviewed_at, review_notes
FROM voter_guides
```

Add scan variables:

```go
reviewedAt  sql.NullTime
reviewNotes sql.NullString
```

Scan them after `summary`, then populate:

```go
ReviewedAt:  optionalDate(reviewedAt),
ReviewNotes: optionalString(reviewNotes),
```

- [ ] **Step 2: Update Postgres endorsement scan**

Change the endorsement query to:

```go
SELECT e.id, e.candidate_id, e.race_id, e.chapter_id, e.voter_guide_id, e.endorsement_type, e.source_url, e.source_date, e.confidence, e.notes, e.reviewed_at, e.review_notes
FROM endorsements e
```

Add scan variables:

```go
reviewedAt  sql.NullTime
reviewNotes sql.NullString
```

Scan them after `notes`, then populate:

```go
ReviewedAt:  optionalDate(reviewedAt),
ReviewNotes: optionalString(reviewNotes),
```

- [ ] **Step 3: Update Postgres measure-position scan**

Change the measure-position query to:

```go
SELECT id, measure_id, chapter_id, position, source_url, confidence, reviewed_at, review_notes
FROM measure_positions
```

Add scan variables:

```go
reviewedAt  sql.NullTime
reviewNotes sql.NullString
```

Scan them after `confidence`, then populate:

```go
ReviewedAt:  optionalDate(reviewedAt),
ReviewNotes: optionalString(reviewNotes),
```

- [ ] **Step 4: Add API camelCase assertion**

In `backend/internal/httpapi/server_test.go`, update the voter guide fixture in `TestDirectoryUsesCamelCaseFieldNames`:

```go
ReviewedAt:  "2026-06-28",
ReviewNotes: "Manual review documented no visible source date.",
```

Then add assertions after the existing `sourceUrl` check:

```go
if got := guide["reviewedAt"]; got != "2026-06-28" {
	t.Fatalf("reviewedAt = %#v, want 2026-06-28", got)
}
if _, ok := guide["reviewed_at"]; ok {
	t.Fatalf("unexpected snake_case reviewed_at in voter guide: %#v", guide)
}
```

- [ ] **Step 5: Run backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

---

## Task 4: Update quality report semantics

**Files:**
- Modify: `backend/internal/directory/quality.go`
- Modify: `backend/internal/directory/quality_test.go`

- [ ] **Step 1: Add tests for documented date exceptions**

Append to `backend/internal/directory/quality_test.go`:

```go
func TestQualityReportAcceptsDocumentedMissingPublishedAt(t *testing.T) {
	data := DirectoryData{
		Chapters: []Chapter{{ID: "california-dsa", Name: "California DSA", State: "CA", Active: true}},
		VoterGuides: []VoterGuide{{
			ID:          "california-dsa-2026-primary-guide",
			Title:       "California DSA 2026 Primary Voter Guide",
			ChapterID:   "california-dsa",
			SourceURL:   "https://www.californiadsa.org/voterguide",
			Confidence:  "official_source",
			ReviewedAt:  "2026-06-28",
			ReviewNotes: "Manual review found no visible publication date on the public source page; do not invent publishedAt.",
		}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})

	if got := report.String(); got != "PASS data quality checks\n" {
		t.Fatalf("expected documented missing publishedAt to pass, got:\n%s", got)
	}
}

func TestQualityReportFlagsBadReviewMetadata(t *testing.T) {
	data := DirectoryData{
		Chapters: []Chapter{{ID: "california-dsa", Name: "California DSA", State: "CA", Active: true}},
		VoterGuides: []VoterGuide{{
			ID:          "california-dsa-2026-primary-guide",
			Title:       "California DSA 2026 Primary Voter Guide",
			ChapterID:   "california-dsa",
			SourceURL:   "https://www.californiadsa.org/voterguide",
			Confidence:  "official_source",
			ReviewedAt:  "06/28/2026",
			ReviewNotes: "No visible date.",
		}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})
	output := report.String()

	for _, want := range []string{
		"WARN voterGuides[california-dsa-2026-primary-guide] reviewedAt must be YYYY-MM-DD",
		"WARN voterGuides[california-dsa-2026-primary-guide] missing publishedAt",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected report to contain %q, got:\n%s", want, output)
		}
	}
}
```

- [ ] **Step 2: Add review metadata helpers**

In `backend/internal/directory/quality.go`, add:

```go
func (r *QualityReport) validateReviewMetadata(record string, reviewedAt string, reviewNotes string) bool {
	if strings.TrimSpace(reviewedAt) == "" && strings.TrimSpace(reviewNotes) == "" {
		return false
	}
	if strings.TrimSpace(reviewedAt) == "" {
		r.add(QualityWarn, record, "missing reviewedAt for reviewNotes")
		return false
	}
	if _, err := time.Parse("2006-01-02", reviewedAt); err != nil {
		r.add(QualityWarn, record, "reviewedAt must be YYYY-MM-DD")
		return false
	}
	if strings.TrimSpace(reviewNotes) == "" {
		r.add(QualityWarn, record, "missing reviewNotes for reviewedAt")
		return false
	}
	return true
}
```

- [ ] **Step 3: Use review metadata for missing date warnings**

In the voter guide loop, before `PublishedAt` handling, add:

```go
hasReviewMetadata := report.validateReviewMetadata(record, guide.ReviewedAt, guide.ReviewNotes)
```

Then change missing `PublishedAt` handling to:

```go
if guide.PublishedAt == "" {
	if !hasReviewMetadata {
		report.add(QualityWarn, record, "missing publishedAt")
	}
} else {
	report.warnIfStale(record, "publishedAt", guide.PublishedAt, referenceDate, staleAfterDays)
}
```

In the endorsement loop, add:

```go
hasReviewMetadata := report.validateReviewMetadata(record, endorsement.ReviewedAt, endorsement.ReviewNotes)
```

Then change missing `SourceDate` handling to:

```go
if endorsement.SourceDate == "" {
	if !hasReviewMetadata {
		report.add(QualityWarn, record, "missing sourceDate")
	}
} else {
	report.warnIfStale(record, "sourceDate", endorsement.SourceDate, referenceDate, staleAfterDays)
}
```

In the measure position loop, call:

```go
report.validateReviewMetadata(record, position.ReviewedAt, position.ReviewNotes)
```

- [ ] **Step 4: Run quality tests**

Run:

```bash
cd backend && go test ./internal/directory -run Quality -count=1
```

Expected: PASS.

---

## Task 5: Update docs and verify the whole repo

**Files:**
- Modify: `docs/reference/editorial-data.md`

- [ ] **Step 1: Document review metadata semantics**

Add this section to `docs/reference/editorial-data.md` after “Manual add/update workflow”:

```markdown
## Editorial review metadata

Source-bearing public records may include:

- `reviewedAt` — the date an editor manually reviewed the source, formatted `YYYY-MM-DD`.
- `reviewNotes` — a short explanation of what was checked or why source metadata is absent.

Use review metadata when a public source is valid but lacks a visible publication/source date. Do not invent `publishedAt` or `sourceDate`. A documented review can suppress missing-date data-quality warnings, but it does not prove the record is true or current.
```

- [ ] **Step 2: Run full verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 3: Run data-quality**

Run:

```bash
make data-quality
```

Expected:

```text
PASS data quality checks
```

- [ ] **Step 4: Check diff hygiene**

Run:

```bash
git diff --check
git status --short
```

Expected: no whitespace errors; only intended files modified.

- [ ] **Step 5: Commit**

Run:

```bash
git add migrations/0001_init.sql seeds/california.sql backend/internal/directory/store.go backend/internal/directory/postgres.go backend/internal/directory/quality.go backend/internal/directory/quality_test.go backend/internal/httpapi/server_test.go web/src/data/schema.ts web/src/data/california.ts web/src/data/schema.test.ts docs/reference/editorial-data.md docs/superpowers/plans/2026-06-28-editorial-review-metadata.md
git commit -m "feat: add editorial review metadata"
```

Expected: commit succeeds.

---

## Self-review

- Spec coverage: resolves the current missing `publishedAt` warning without falsifying source data.
- Contract impact: extends read DTOs and JSON only with optional fields; existing clients remain compatible.
- Source-link integrity: keeps source URLs required and uses review notes for ambiguity.
- Privacy: no address, ZIP, voter-file, auth, admin, or submission data is added.
- Verification: `make verify`, `make data-quality`, focused Go quality tests, and frontend schema tests are required.
