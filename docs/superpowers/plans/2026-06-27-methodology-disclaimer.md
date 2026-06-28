# Methodology/Disclaimer Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add public methodology and disclaimer content to explain how DSA Slate sources, labels, and limits directory records.

**Architecture:** Keep this as a static frontend section and reusable component for now, with matching docs updates. No routing library or backend changes in this slice. Link to the section from the homepage hero so users can quickly understand confidence labels, endorsement wording, privacy, and election-information limits.

**Tech Stack:** Vite React TypeScript, existing static seed data UI, Vitest server-render tests, Markdown docs.

---

## Scope

- Add a methodology/disclaimer section to the public homepage.
- Explain source rules, confidence labels, endorsement-vs-recommendation wording, privacy/no raw addresses, and election info disclaimer.
- Link from hero/header to the methodology section.
- Keep copy concise and non-legalistic.
- Update tests and product/editorial docs.

Out of scope:

- Legal review automation
- Real routing/detail pages
- Backend/API changes
- Public submission forms

## Tasks

### Task 1: Methodology component

**Files:**
- Create: `web/src/components/MethodologyPanel.tsx`
- Modify: `web/src/App.tsx`

- [ ] Create a `MethodologyPanel` component with sections: “How records are sourced,” “Confidence labels,” “Endorsement wording,” “Privacy,” and “Election information disclaimer.”
- [ ] Add it near the bottom of the homepage with `id="methodology"`.
- [ ] Add a hero link/button to `#methodology`.

### Task 2: Tests

**Files:**
- Modify: `web/src/App.test.tsx`

- [ ] Assert homepage renders “Methodology & disclaimer”.
- [ ] Assert homepage renders `No raw voter addresses are stored` or equivalent.
- [ ] Assert homepage renders confidence-label explanation and endorsement/recommendation wording.

### Task 3: Docs

**Files:**
- Modify: `docs/reference/editorial-data.md`
- Modify: `docs/product.md`

- [ ] Note that the public methodology/disclaimer is part of the first launch slice.
- [ ] Keep docs aligned with UI copy: source-linked, confidence-labeled, not official voter info, no raw address storage.

### Task 4: Verify and review

- [ ] Run `pnpm --dir web run test`.
- [ ] Run `pnpm --dir web run format`.
- [ ] Run `pnpm --dir web run lint`.
- [ ] Run `make verify`.
- [ ] Review that copy does not imply legal advice, official election administration, or live address lookup.

## Acceptance criteria

- Public homepage includes methodology/disclaimer content.
- Hero links to methodology section.
- Confidence labels and endorsed/recommended distinction are explained.
- Privacy copy says raw voter addresses are not stored.
- Disclaimer says DSA Slate links to official election information rather than replacing it.
- `make verify` passes.
