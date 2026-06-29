# Data Quality Admin Prep Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add a deterministic data-quality reporting slice that catches weak or stale source-linked records before any write/admin API work begins.

**Architecture:** Keep the public API read-only. Add a backend `directory` quality package helper that evaluates `DirectoryData`, then expose it through a small CLI command that reads the current Postgres-backed directory after `make seed`. Keep frontend validation intact and document the manual editorial workflow so future admin/write features have clear acceptance criteria.

**Tech Stack:** Go 1.x backend, pgxpool/Postgres, existing SQL migrations/seeds, Vite React TypeScript frontend validation, Makefile verification.

---

## Scope boundaries

### In scope

- Deterministic data-quality findings for source-linked electoral records.
- A CLI command for local/editorial review against seeded Postgres data.
- Tests for missing source context, stale source dates, weak confidence labels, and broken editorial coverage.
- Documentation for a manual add/update record workflow.
- Seed fixture expansion only if needed to prove the checker catches multiple record classes.

### Out of scope

- No write/admin API endpoints.
- No auth, sessions, roles, or admin UI.
- No raw voter addresses, address lookup, or ZIP lookup.
- No live link crawler/scraper.
- No migration for editorial status fields unless a later slice needs persisted review state.

## File map

- Create: `backend/internal/directory/quality.go` — pure quality checks over `DirectoryData`.
- Create: `backend/internal/directory/quality_test.go` — unit tests for quality findings and clean seed-like data.
- Create: `backend/cmd/dataquality/main.go` — CLI entrypoint that loads Postgres directory data and prints a stable report.
- Modify: `Makefile` — add `data-quality` target that runs the CLI against local seeded Postgres.
- Modify: `docs/reference/editorial-data.md` — document quality findings and manual record workflow.
- Modify: `README.md` — mention the data-quality command beside API/dev workflow.
- Optional modify: `seeds/california.sql` and `web/src/data/california.ts` — only to add safe, source-linked fixture coverage if tests reveal the current fixture is too narrow.

---

## Task 1: Add pure backend quality checks

**Files:**
- Create: `backend/internal/directory/quality.go`
- Create: `backend/internal/directory/quality_test.go`

- [ ] **Step 1: Write failing tests for quality findings**

Create `backend/internal/directory/quality_test.go`:

```go
package directory

import (
	"strings"
	"testing"
)

func TestQualityReportPassesCleanSourceLinkedData(t *testing.T) {
	data := DirectoryData{
		Chapters: []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", WebsiteURL: "https://dsasandiego.org/", Active: true}},
		Elections: []Election{{ID: "2026-ca-primary", Name: "2026 California Primary", ElectionDate: "2026-06-02", Jurisdiction: "CA"}},
		Races: []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", District: "CA-48", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates: []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		VoterGuides: []VoterGuide{{ID: "dsa-san-diego-2026-primary-guide", Title: "DSA San Diego 2026 Primary Guide", ChapterID: "dsa-san-diego", ElectionID: "2026-ca-primary", SourceURL: "https://dsasandiego.org/guide", Confidence: "official_source", PublishedAt: "2026-05-05"}},
		Endorsements: []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", VoterGuideID: "dsa-san-diego-2026-primary-guide", EndorsementType: "recommended", SourceURL: "https://dsasandiego.org/guide", SourceDate: "2026-05-05", Confidence: "official_source", Notes: "Preserve recommendation language."}},
		BallotMeasures: []BallotMeasure{{ID: "san-diego-measure-a-2026-primary", ElectionID: "2026-ca-primary", Title: "San Diego Measure A", Jurisdiction: "San Diego, CA", MeasureNumber: "A"}},
		MeasurePositions: []MeasurePosition{{ID: "san-diego-measure-a-dsa-sd-yes-2026-primary", MeasureID: "san-diego-measure-a-2026-primary", ChapterID: "dsa-san-diego", Position: "yes", SourceURL: "https://dsasandiego.org/guide", Confidence: "official_source"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})

	if report.HasFailures() {
		t.Fatalf("expected no failed findings, got:\n%s", report.String())
	}
}

func TestQualityReportFlagsMissingAndWeakSourceContext(t *testing.T) {
	data := DirectoryData{
		Chapters: []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", Active: true}},
		Elections: []Election{{ID: "2026-ca-primary", Name: "2026 California Primary"}},
		Races: []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates: []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		VoterGuides: []VoterGuide{{ID: "dsa-san-diego-2026-primary-guide", Title: "Guide", ChapterID: "dsa-san-diego", SourceURL: "", Confidence: "needs_review"}},
		Endorsements: []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", EndorsementType: "endorsed", SourceURL: "", Confidence: "needs_review"}},
		MeasurePositions: []MeasurePosition{{ID: "measure-position", MeasureID: "missing-measure", ChapterID: "dsa-san-diego", Position: "yes", SourceURL: "", Confidence: "needs_review"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})
	output := report.String()

	for _, want := range []string{
		"FAIL voterGuides[dsa-san-diego-2026-primary-guide] missing sourceUrl",
		"WARN voterGuides[dsa-san-diego-2026-primary-guide] confidence needs_review requires manual review",
		"FAIL endorsements[contreras-dsa-sd-2026-primary] missing sourceUrl",
		"WARN endorsements[contreras-dsa-sd-2026-primary] missing sourceDate",
		"FAIL measurePositions[measure-position] missing sourceUrl",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected report to contain %q, got:\n%s", want, output)
		}
	}
}

func TestQualityReportFlagsStaleSourceDates(t *testing.T) {
	data := DirectoryData{
		Chapters: []Chapter{{ID: "dsa-san-diego", Name: "DSA San Diego", State: "CA", Active: true}},
		Elections: []Election{{ID: "2026-ca-primary", Name: "2026 California Primary"}},
		Races: []Race{{ID: "ca-48-us-house-2026-primary", ElectionID: "2026-ca-primary", Office: "U.S. House", Level: "federal", Jurisdiction: "CA-48"}},
		Candidates: []Candidate{{ID: "corinna-contreras", Name: "Corinna Contreras"}},
		Endorsements: []Endorsement{{ID: "contreras-dsa-sd-2026-primary", CandidateID: "corinna-contreras", RaceID: "ca-48-us-house-2026-primary", ChapterID: "dsa-san-diego", EndorsementType: "recommended", SourceURL: "https://dsasandiego.org/guide", SourceDate: "2025-01-01", Confidence: "official_source"}},
	}

	report := EvaluateQuality(data, QualityOptions{ReferenceDate: "2026-06-28", StaleAfterDays: 180})

	if !strings.Contains(report.String(), "WARN endorsements[contreras-dsa-sd-2026-primary] sourceDate is older than 180 days") {
		t.Fatalf("expected stale source date warning, got:\n%s", report.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd backend && go test ./internal/directory -run Quality -count=1
```

Expected: FAIL with `undefined: EvaluateQuality`.

- [ ] **Step 3: Implement minimal quality report types and checks**

Create `backend/internal/directory/quality.go`:

```go
package directory

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type QualitySeverity string

const (
	QualityFail QualitySeverity = "FAIL"
	QualityWarn QualitySeverity = "WARN"
)

type QualityOptions struct {
	ReferenceDate  string
	StaleAfterDays int
}

type QualityFinding struct {
	Severity QualitySeverity
	Record   string
	Message  string
}

type QualityReport struct {
	Findings []QualityFinding
}

func (r QualityReport) HasFailures() bool {
	for _, finding := range r.Findings {
		if finding.Severity == QualityFail {
			return true
		}
	}
	return false
}

func (r QualityReport) String() string {
	if len(r.Findings) == 0 {
		return "PASS data quality checks\n"
	}

	findings := append([]QualityFinding(nil), r.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity < findings[j].Severity
		}
		if findings[i].Record != findings[j].Record {
			return findings[i].Record < findings[j].Record
		}
		return findings[i].Message < findings[j].Message
	})

	var builder strings.Builder
	for _, finding := range findings {
		builder.WriteString(fmt.Sprintf("%s %s %s\n", finding.Severity, finding.Record, finding.Message))
	}
	return builder.String()
}

func EvaluateQuality(data DirectoryData, options QualityOptions) QualityReport {
	report := QualityReport{Findings: make([]QualityFinding, 0)}
	data = NormalizeDirectoryData(data)

	staleAfterDays := options.StaleAfterDays
	if staleAfterDays <= 0 {
		staleAfterDays = 180
	}

	referenceDate := time.Now().UTC()
	if options.ReferenceDate != "" {
		if parsed, err := time.Parse("2006-01-02", options.ReferenceDate); err == nil {
			referenceDate = parsed
		} else {
			report.add(QualityWarn, "qualityOptions", "referenceDate must be YYYY-MM-DD")
		}
	}

	for _, guide := range data.VoterGuides {
		record := fmt.Sprintf("voterGuides[%s]", guide.ID)
		report.requireSourceURL(record, guide.SourceURL)
		report.warnWeakConfidence(record, guide.Confidence)
		if guide.PublishedAt == "" {
			report.add(QualityWarn, record, "missing publishedAt")
		} else {
			report.warnIfStale(record, "publishedAt", guide.PublishedAt, referenceDate, staleAfterDays)
		}
	}

	for _, endorsement := range data.Endorsements {
		record := fmt.Sprintf("endorsements[%s]", endorsement.ID)
		report.requireSourceURL(record, endorsement.SourceURL)
		report.warnWeakConfidence(record, endorsement.Confidence)
		if endorsement.SourceDate == "" {
			report.add(QualityWarn, record, "missing sourceDate")
		} else {
			report.warnIfStale(record, "sourceDate", endorsement.SourceDate, referenceDate, staleAfterDays)
		}
		if endorsement.EndorsementType == "unknown" {
			report.add(QualityWarn, record, "endorsementType unknown requires manual review")
		}
	}

	for _, position := range data.MeasurePositions {
		record := fmt.Sprintf("measurePositions[%s]", position.ID)
		report.requireSourceURL(record, position.SourceURL)
		report.warnWeakConfidence(record, position.Confidence)
		if position.Position == "unknown" {
			report.add(QualityWarn, record, "position unknown requires manual review")
		}
	}

	return report
}

func (r *QualityReport) add(severity QualitySeverity, record string, message string) {
	r.Findings = append(r.Findings, QualityFinding{Severity: severity, Record: record, Message: message})
}

func (r *QualityReport) requireSourceURL(record string, sourceURL string) {
	if strings.TrimSpace(sourceURL) == "" {
		r.add(QualityFail, record, "missing sourceUrl")
	}
}

func (r *QualityReport) warnWeakConfidence(record string, confidence string) {
	switch confidence {
	case "needs_review", "campaign_source", "press_source":
		r.add(QualityWarn, record, fmt.Sprintf("confidence %s requires manual review", confidence))
	}
}

func (r *QualityReport) warnIfStale(record string, fieldName string, value string, referenceDate time.Time, staleAfterDays int) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		r.add(QualityWarn, record, fmt.Sprintf("%s must be YYYY-MM-DD", fieldName))
		return
	}

	if referenceDate.Sub(date) > time.Duration(staleAfterDays)*24*time.Hour {
		r.add(QualityWarn, record, fmt.Sprintf("%s is older than %d days", fieldName, staleAfterDays))
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd backend && go test ./internal/directory -run Quality -count=1
```

Expected: PASS.

---

## Task 2: Add the data-quality CLI and Makefile target

**Files:**
- Create: `backend/cmd/dataquality/main.go`
- Modify: `Makefile`

- [ ] **Step 1: Write the CLI entrypoint**

Create `backend/cmd/dataquality/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/dsa-slate/internal/directory"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	store := directory.NewPostgresStore(pool)
	data, err := store.Directory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load directory: %v\n", err)
		os.Exit(2)
	}

	referenceDate := os.Getenv("QUALITY_REFERENCE_DATE")
	if referenceDate == "" {
		referenceDate = time.Now().UTC().Format("2006-01-02")
	}

	report := directory.EvaluateQuality(data, directory.QualityOptions{ReferenceDate: referenceDate, StaleAfterDays: 180})
	fmt.Print(report.String())
	if report.HasFailures() {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Add a root Makefile target**

Modify `Makefile`:

```make
.PHONY: help verify deps fmt lint test build compose-config open-pilot-check \
	containers-up containers-down containers-clean logs db-up db-shell migrate db-reset seed data-quality dev-web dev-api api-smoke dev dev-full
```

Add after `seed`:

```make
data-quality: seed ## Run source-linked data quality checks against local seeded Postgres
	@published_port="$$($(COMPOSE) port postgres 5432)"; \
	postgres_port="$${published_port##*:}"; \
	DATABASE_URL="$${DATABASE_URL:-postgres://$${POSTGRES_USER:-dsa_slate}:$${POSTGRES_PASSWORD:-change-me}@localhost:$${POSTGRES_PORT:-$${postgres_port:-5432}}/$${POSTGRES_DB:-dsa_slate}?sslmode=disable}"; \
	QUALITY_REFERENCE_DATE="$${QUALITY_REFERENCE_DATE:-2026-06-28}"; \
	export DATABASE_URL QUALITY_REFERENCE_DATE; \
	cd backend && go run ./cmd/dataquality
```

- [ ] **Step 3: Run CLI against seeded database**

Run:

```bash
make data-quality
```

Expected: exits 0 and prints either `PASS data quality checks` or stable `WARN ...` lines only. It must not print `FAIL ...` for the checked-in seed fixture.

- [ ] **Step 4: Run backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

---

## Task 3: Extend frontend schema tests for editorial parity

**Files:**
- Modify: `web/src/data/schema.test.ts`
- Optional modify: `web/src/data/schema.ts`

- [ ] **Step 1: Add tests that lock current frontend validator guarantees**

Append to `describe('directory schema helpers', ...)` in `web/src/data/schema.test.ts`:

```ts
it('keeps source-linked public records mandatory before UI card joins', () => {
  const data: DirectoryData = {
    chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
    elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
    races: [],
    candidates: [],
    voterGuides: [
      {
        id: 'guide-without-source',
        title: 'Guide without source',
        chapterId: 'dsa-san-diego',
        electionId: '2026-ca-primary',
        sourceUrl: '',
        confidence: 'needs_review',
      },
    ],
    endorsements: [],
    ballotMeasures: [{ id: 'measure-a', electionId: '2026-ca-primary', title: 'Measure A', jurisdiction: 'San Diego, CA' }],
    measurePositions: [
      {
        id: 'measure-position-without-source',
        measureId: 'measure-a',
        chapterId: 'dsa-san-diego',
        position: 'unknown',
        sourceUrl: '',
        confidence: 'needs_review',
      },
    ],
  };

  expect(validateDirectoryData(data)).toEqual(
    expect.arrayContaining([
      'voterGuides[guide-without-source] must include sourceUrl',
      'measurePositions[measure-position-without-source] must include sourceUrl',
    ]),
  );
});
```

- [ ] **Step 2: Run focused frontend tests**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS. If the test fails, adjust only `web/src/data/schema.ts` source-url validation; do not loosen existing validation.

---

## Task 4: Document manual admin-prep workflow

**Files:**
- Modify: `docs/reference/editorial-data.md`
- Modify: `README.md`

- [ ] **Step 1: Update editorial reference**

Add this section to `docs/reference/editorial-data.md` after “Seed review checklist”:

```markdown
## Manual add/update workflow

Until a reviewed admin API exists, all public records are added through reviewed seed changes:

1. Capture the public source URL before editing data.
2. Confirm the endorsing body, chapter, election, race, candidate, or measure context.
3. Preserve exact wording: `recommended` stays `recommended`; do not upgrade it to `endorsed`.
4. Use `needs_review` when source authority, date, or wording is unclear.
5. Update both checked-in seed surfaces when needed: `seeds/california.sql` for Postgres and `web/src/data/california.ts` for static fallback.
6. Run `make verify`.
7. Run `make data-quality` when local Docker/Postgres is available.

The data-quality command is a metadata check, not a truth oracle. Human review still decides whether a source is official enough and whether a record should be published.
```

Add this section near “Automated seed validation”:

```markdown
## Backend data-quality report

`make data-quality` seeds local Postgres and runs `backend/cmd/dataquality` against the read model.

The report fails on missing source URLs for public voter-guide, endorsement, and measure-position records. It warns on stale dates, missing dates, weak confidence labels, and unknown public positions. Warnings are allowed during early seed expansion, but they should be resolved or explained before an admin/write API is introduced.
```

- [ ] **Step 2: Update README**

Add a short bullet under the Development/API section in `README.md`:

```markdown
- Run `make data-quality` after seed edits to check source-linked records against local Postgres before proposing admin/write API work.
```

- [ ] **Step 3: Verify docs render as plain Markdown**

Run:

```bash
git diff -- docs/reference/editorial-data.md README.md
```

Expected: diff only contains the new workflow/command notes.

---

## Task 5: Full verification and commit

**Files:**
- All files touched above.

- [ ] **Step 1: Run repository verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 2: Run data-quality command**

Run:

```bash
make data-quality
```

Expected: PASS or warnings only; no `FAIL` lines. If it fails due to missing Docker/Postgres, record the environment failure and ensure `make verify` passed.

- [ ] **Step 3: Check diff hygiene**

Run:

```bash
git diff --check
git status --short
```

Expected: no whitespace errors; only intended files modified.

- [ ] **Step 4: Commit**

Run:

```bash
git add backend/internal/directory/quality.go backend/internal/directory/quality_test.go backend/cmd/dataquality/main.go Makefile docs/reference/editorial-data.md README.md web/src/data/schema.test.ts docs/superpowers/plans/2026-06-28-data-quality-admin-prep.md
git commit -m "feat: add data quality report"
```

Expected: commit succeeds.

---

## Self-review

- Spec coverage: covers a deterministic data quality report, local command, docs/manual workflow, and keeps write/admin API out of scope.
- Source-link rules: public records still require source URLs; weak confidence and stale/missing dates are visible warnings.
- Privacy: no raw addresses, ZIP lookup, voter files, auth, or admin UI are introduced.
- Verification: focused Go tests, focused frontend schema test, `make verify`, `make data-quality`, and `git diff --check` are required.
