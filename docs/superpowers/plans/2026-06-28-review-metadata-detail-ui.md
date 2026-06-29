# Review Metadata Detail UI Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Show editorial review metadata in the source-linked detail panel so users can see when a record was manually reviewed and why source dates may be absent.

**Architecture:** The API and static seed already expose optional `reviewedAt` and `reviewNotes` on source-bearing records. This slice carries those fields through frontend card DTOs and renders them in `DetailPanel` only, keeping browse cards uncluttered. No backend/API/schema changes are needed.

**Tech Stack:** Vite React TypeScript, existing joined card DTOs in `web/src/data/index.ts`, server-rendered Vitest component assertions, Tailwind v4 semantic utilities.

---

## Scope boundaries

### In scope

- Add `reviewedAt` / `reviewNotes` to `EndorsementCardData`, `MeasureCardData`, and `VoterGuideCardData`.
- Populate those fields from `DirectoryData` in `buildEndorsementCards`, `buildMeasureCards`, and `buildVoterGuideCards`.
- Render review metadata in `DetailPanel` for endorsements, ballot measure positions, and voter guides.
- Add tests proving DSA-LA/CA DSA review metadata appears in the detail panel.
- Document the UI rule in `docs/reference/design.md`.

### Out of scope

- No backend, API, migration, seed, or TypeScript schema changes.
- No admin/write API.
- No new routes/pages/modals.
- No link crawler or source truth verification.
- No card-face clutter; review metadata belongs in the detail panel.

## File map

- Modify: `web/src/data/index.ts` — add review metadata fields to joined card DTOs.
- Modify: `web/src/components/DetailPanel.tsx` — render review metadata rows and a short source review note.
- Modify: `web/src/App.test.tsx` — add component assertions for review metadata.
- Modify: `web/src/data/schema.test.ts` — lock DTO propagation from static seed data.
- Modify: `docs/reference/design.md` — document where review metadata appears.

---

## Task 1: Carry review metadata into card DTOs

**Files:**
- Modify: `web/src/data/index.ts`
- Modify: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add failing DTO propagation tests**

In `web/src/data/schema.test.ts`, inside `describe('california seed data', ...)`, add assertions to the existing `builds joined endorsement, measure, and guide cards` test.

For the DSA-LA endorsement object, add:

```ts
reviewedAt: '2026-06-28',
reviewNotes: expect.stringContaining('Manual review confirmed DSA-LA voter guide'),
```

For the DSA-LA voter guide object, add:

```ts
reviewedAt: '2026-06-28',
reviewNotes: expect.stringContaining('no visible publication date'),
```

For the San Diego measure object, add:

```ts
reviewedAt: '2026-06-28',
reviewNotes: expect.stringContaining('Manual review confirmed source URL'),
```

- [ ] **Step 2: Run the focused test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: FAIL because joined DTOs do not yet include `reviewedAt` / `reviewNotes`.

- [ ] **Step 3: Add review fields to card types**

In `web/src/data/index.ts`, add these optional fields to `EndorsementCardData`, `MeasureCardData`, and `VoterGuideCardData`:

```ts
reviewedAt?: string;
reviewNotes?: string;
```

- [ ] **Step 4: Populate review fields in card builders**

In `buildEndorsementCards`, add:

```ts
reviewedAt: endorsement.reviewedAt,
reviewNotes: endorsement.reviewNotes,
```

In `buildMeasureCards`, add:

```ts
reviewedAt: position.reviewedAt,
reviewNotes: position.reviewNotes,
```

In `buildVoterGuideCards`, add:

```ts
reviewedAt: guide.reviewedAt,
reviewNotes: guide.reviewNotes,
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

---

## Task 2: Render review metadata in the detail panel

**Files:**
- Modify: `web/src/components/DetailPanel.tsx`
- Modify: `web/src/App.test.tsx`

- [ ] **Step 1: Add failing detail panel tests**

In `web/src/App.test.tsx`, update imports:

```ts
import { getEndorsementCards, getMeasureCards, getVoterGuideCards } from './data';
```

Add this test after `shows selected record details with source metadata`:

```tsx
it('shows editorial review metadata in selected record details', () => {
  const dsaLaEndorsement = getEndorsementCards().find((card) => card.id === 'hugo-soto-martinez-dsa-la-2026-primary');
  const caGuide = getVoterGuideCards().find((card) => card.id === 'california-dsa-2026-primary-guide');
  const measure = getMeasureCards()[0];

  expect(dsaLaEndorsement).toBeDefined();
  expect(caGuide).toBeDefined();

  const endorsementHtml = renderToString(<DetailPanel item={dsaLaEndorsement} />);
  expect(endorsementHtml).toContain('Editorial review');
  expect(endorsementHtml).toContain('Reviewed on');
  expect(endorsementHtml).toContain('2026-06-28');
  expect(endorsementHtml).toContain('Manual review confirmed DSA-LA voter guide');

  const guideHtml = renderToString(<DetailPanel item={caGuide} />);
  expect(guideHtml).toContain('No visible publication date');
  expect(guideHtml).toContain('do not invent publishedAt');

  const measureHtml = renderToString(<DetailPanel item={measure} />);
  expect(measureHtml).toContain('Manual review confirmed source URL');
});
```

- [ ] **Step 2: Run the focused test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: FAIL because the detail panel does not yet render review metadata.

- [ ] **Step 3: Add a review metadata component**

In `web/src/components/DetailPanel.tsx`, add below `DetailLink`:

```tsx
function ReviewMetadata({ reviewedAt, reviewNotes }: { reviewedAt?: string; reviewNotes?: string }) {
  if (!reviewedAt && !reviewNotes) {
    return null;
  }

  return (
    <section className="mt-5 border-2 border-border bg-surface-muted p-4">
      <p className="text-xs font-black uppercase tracking-wide text-primary">Editorial review</p>
      <dl className="mt-2">
        <DetailRow label="Reviewed on" value={reviewedAt} />
        <DetailRow label="Review notes" value={reviewNotes} />
      </dl>
    </section>
  );
}
```

- [ ] **Step 4: Render review metadata for all detail types**

In `EndorsementDetail`, render after the `<dl>`:

```tsx
<ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
```

In `MeasureDetail`, render after the `<dl>`:

```tsx
<ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
```

In `VoterGuideDetail`, render after the `<dl>`:

```tsx
<ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
```

- [ ] **Step 5: Improve missing publication date copy for reviewed voter guides**

In `VoterGuideDetail`, after the `<DetailRow label="Published" ... />`, add:

```tsx
<DetailRow
  label="Publication date"
  value={!item.publishedAt && item.reviewedAt ? 'No visible publication date in reviewed source.' : undefined}
/>
```

This ensures users see why `Published` is absent for CA DSA / DSA-LA guides.

- [ ] **Step 6: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: PASS.

---

## Task 3: Document the UI rule and verify

**Files:**
- Modify: `docs/reference/design.md`

- [ ] **Step 1: Add the design note**

Add this section near the detail panel guidance in `docs/reference/design.md`:

```markdown
## Editorial review metadata in details

Review metadata appears only in the selected record detail panel, not on card faces. Use `reviewedAt` and `reviewNotes` to explain manual source review, especially when a valid source has no visible publication or source date. Keep source links, confidence labels, and original endorsement wording more prominent than review notes.
```

- [ ] **Step 2: Run full verification**

Run:

```bash
make verify
```

Expected: PASS.

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
git add web/src/data/index.ts web/src/components/DetailPanel.tsx web/src/App.test.tsx web/src/data/schema.test.ts docs/reference/design.md docs/superpowers/plans/2026-06-28-review-metadata-detail-ui.md
git commit -m "feat: show review metadata in details"
```

Expected: commit succeeds.

---

## Self-review

- Spec coverage: makes review metadata visible in details for all source-bearing record kinds.
- Scope control: no backend/API/schema/seed changes; optional metadata is already available.
- UX: cards remain uncluttered; review context appears in the detail panel where users inspect source context.
- Verification: focused DTO and detail tests plus `make verify` and `git diff --check` are required.
