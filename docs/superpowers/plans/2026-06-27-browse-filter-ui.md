# Browse/Filter UI Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the disabled browse-only search affordance with working client-side filtering across the static trusted-directory cards.

**Architecture:** Keep filtering entirely client-side for this slice. Reuse validated card data from `buildDirectoryCards(directoryData)`, derive a search index in React, and filter sections without changing seed data or adding routing. Empty states should make clear that this is text filtering, not address lookup.

**Tech Stack:** Vite React TypeScript, existing React components, Vitest server rendering tests, static TypeScript seed data.

---

## Scope

- Add a working text search field for candidates, voter guides, measures, chapters, offices, districts, elections, jurisdictions, endorsement/position labels, and confidence labels.
- Show filtered counts in section headings or helper text.
- Preserve all source links and confidence labels.
- Keep address/ZIP lookup explicitly out of scope.
- Add tests for matching candidate text, matching ballot measure text, and no-results copy.

Out of scope:

- URL query params
- fuzzy search library
- address or ZIP lookup
- backend/API search
- routing/detail pages

## Files

- Modify `web/src/App.tsx` — add React state, search input, filtering helpers, and empty states.
- Modify `web/src/App.test.tsx` — switch from render-only tests to interaction-capable React tests if existing tooling supports it, or test exported pure filter helper.
- Optionally create `web/src/data/filter.ts` — pure filtering helpers if App would become too large.
- Optionally create `web/src/data/filter.test.ts` — focused unit tests for search matching.

## Task 1: Add pure filter helpers

**Files:**
- Create: `web/src/data/filter.ts`
- Create: `web/src/data/filter.test.ts`

- [ ] Create `filterDirectoryCards(cards, query)` that returns `{ endorsements, measures, voterGuides }`.
- [ ] Normalize query and haystack with lowercase whitespace trimming.
- [ ] Match endorsements on candidate, office, district, election, chapter, endorsement type, confidence, and source date.
- [ ] Match measures on title, jurisdiction, election, chapter, position, and confidence.
- [ ] Match voter guides on title, chapter, election, confidence, published date, and summary.
- [ ] Add tests for candidate, measure, guide, and no-match filtering.

## Task 2: Wire filter UI into homepage

**Files:**
- Modify: `web/src/App.tsx`

- [ ] Replace disabled search UI with controlled input.
- [ ] Use `filterDirectoryCards(directoryCards, query)` for displayed cards.
- [ ] Add helper copy: “Text search only — address lookup comes later and will not store raw addresses.”
- [ ] Show total visible result count.
- [ ] Show an empty-state card when no records match.

## Task 3: Update public rendering tests

**Files:**
- Modify: `web/src/App.test.tsx`
- Test: `web/src/data/filter.test.ts`

- [ ] Keep existing source/confidence render assertions.
- [ ] Assert homepage now renders an enabled search input with text-search copy.
- [ ] Assert pure filter tests cover candidate and measure results.
- [ ] Avoid browser-only test dependencies unless already installed.

## Task 4: Verify and review

- [ ] Run `pnpm --dir web run test`.
- [ ] Run `pnpm --dir web run format`.
- [ ] Run `pnpm --dir web run lint`.
- [ ] Run `make verify`.
- [ ] Review that UI copy never implies address lookup is live.

## Acceptance criteria

- Homepage search input is enabled and filters static cards.
- Candidate, measure, and voter-guide matches work from pure tests.
- Empty state appears for no matches.
- Source/confidence labels still render.
- Address lookup remains clearly later-phase.
- `make verify` passes.
