# Card Detail UX Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add a polished, accessible detail panel so users can inspect each endorsement, ballot measure, or voter guide with source context after browsing/filtering live API data.

**Architecture:** Keep the current API, browse filters, and text search unchanged. Enrich the existing joined card DTOs in `web/src/data/index.ts`, make directory cards selectable without nesting source links inside buttons, and render a reusable `DetailPanel` alongside the lists. The app owns one selected-card key and clears it safely when filters remove the selected record.

**Tech Stack:** Vite React TypeScript, Tailwind v4 semantic utilities, Vitest server-render smoke tests, existing source-linked directory schema.

---

## Current State

- `web/src/App.tsx` renders three card sections from `buildDirectoryCards(directoryState.data)`.
- Structured browse filters are API-backed through `useDirectoryData(fetch, browseFilters)`.
- Text search remains client-side through `filterDirectoryCards(directoryCards, query)`.
- `web/src/components/DirectoryCard.tsx` renders static `<article>` cards with `SourceBadge` links.
- `web/src/data/index.ts` joins raw records into `EndorsementCardData`, `MeasureCardData`, and `VoterGuideCardData`, but drops IDs and richer source context needed for a detail panel.

## UX Rules

- Use a persistent inline panel, not a modal.
- Desktop/tablet: show detail panel in a two-column layout beside the lists.
- Mobile: the panel stacks above the sections after the filter controls.
- Cards use a single selectable surface and a separate source link; avoid nested interactive elements.
- Selected state must not rely on color alone: use `aria-pressed`, `border-primary`, a visible label, and offset shadow changes.
- Detail region must be labeled and use `aria-live="polite"` so selection changes are announced.
- Preserve source-link visibility and focus rings.

## File Map

- Modify: `web/src/data/index.ts`
  - Add `DirectoryCardKind`, `DirectoryCardKey`, richer fields, and `allDirectoryCards()` helper.
- Modify: `web/src/data/schema.test.ts`
  - Assert enriched joined fields are present for seed data.
- Modify: `web/src/components/DirectoryCard.tsx`
  - Add selectable card props and selected state styling.
- Create: `web/src/components/DetailPanel.tsx`
  - Render selected card details for endorsements, measures, and voter guides.
- Modify: `web/src/App.tsx`
  - Track selected card key, pass selection props to cards, render panel, and clear stale selections.
- Modify: `web/src/App.test.tsx`
  - Add smoke assertions for detail panel, selectable card state, and source metadata.
- Modify: `docs/reference/design.md`
  - Document selected-card/detail-panel styling rules.

## Task 1: Enrich Joined Card Data

**Files:**
- Modify: `web/src/data/index.ts`
- Test: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add richer card types**

In `web/src/data/index.ts`, replace the card type block at the top with this expanded version:

```ts
export type DirectoryCardKind = 'endorsement' | 'measure' | 'voterGuide';
export type DirectoryCardKey = `${DirectoryCardKind}:${string}`;

export type EndorsementCardData = {
  kind: 'endorsement';
  key: DirectoryCardKey;
  id: string;
  candidateId: string;
  candidateName: string;
  candidateParty?: string;
  candidateCampaignUrl?: string;
  raceId: string;
  office: string;
  district?: string;
  raceLevel: DirectoryData['races'][number]['level'];
  raceJurisdiction: string;
  electionId: string;
  electionName: string;
  electionDate?: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  voterGuideId?: string;
  endorsementType: EndorsementType;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  sourceDate?: string;
  notes?: string;
};

export type MeasureCardData = {
  kind: 'measure';
  key: DirectoryCardKey;
  id: string;
  measureId: string;
  title: string;
  jurisdiction: string;
  measureNumber?: string;
  summary?: string;
  electionId: string;
  electionName: string;
  electionDate?: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  position: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
};

export type VoterGuideCardData = {
  kind: 'voterGuide';
  key: DirectoryCardKey;
  id: string;
  title: string;
  chapterId: string;
  chapterName: string;
  chapterWebsiteUrl?: string;
  electionId?: string;
  electionName?: string;
  electionDate?: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  publishedAt?: string;
  summary?: string;
};
```

- [ ] **Step 2: Add a helper to flatten cards**

In `web/src/data/index.ts`, add below `buildDirectoryCards`:

```ts
export function allDirectoryCards(cards: DirectoryCards): Array<EndorsementCardData | MeasureCardData | VoterGuideCardData> {
  return [...cards.voterGuides, ...cards.endorsements, ...cards.measures];
}
```

- [ ] **Step 3: Populate enriched fields in endorsement cards**

In `buildEndorsementCards`, return this object:

```ts
return {
  kind: 'endorsement',
  key: `endorsement:${endorsement.id}`,
  id: endorsement.id,
  candidateId: endorsement.candidateId,
  candidateName: candidate.name,
  candidateParty: candidate.party,
  candidateCampaignUrl: candidate.campaignUrl,
  raceId: endorsement.raceId,
  office: race.office,
  district: race.district,
  raceLevel: race.level,
  raceJurisdiction: race.jurisdiction,
  electionId: race.electionId,
  electionName: election.name,
  electionDate: election.electionDate,
  chapterId: endorsement.chapterId,
  chapterName: chapter.name,
  chapterWebsiteUrl: chapter.websiteUrl,
  voterGuideId: endorsement.voterGuideId,
  endorsementType: endorsement.endorsementType,
  confidence: endorsement.confidence,
  sourceUrl: endorsement.sourceUrl,
  sourceDate: endorsement.sourceDate,
  notes: endorsement.notes,
};
```

- [ ] **Step 4: Populate enriched fields in measure cards**

In `buildMeasureCards`, return this object:

```ts
return {
  kind: 'measure',
  key: `measure:${position.id}`,
  id: position.id,
  measureId: position.measureId,
  title: measure.title,
  jurisdiction: measure.jurisdiction,
  measureNumber: measure.measureNumber,
  summary: measure.summary,
  electionId: measure.electionId,
  electionName: election.name,
  electionDate: election.electionDate,
  chapterId: position.chapterId,
  chapterName: chapter.name,
  chapterWebsiteUrl: chapter.websiteUrl,
  position: position.position,
  confidence: position.confidence,
  sourceUrl: position.sourceUrl,
};
```

- [ ] **Step 5: Populate enriched fields in voter guide cards**

In `buildVoterGuideCards`, return this object:

```ts
return {
  kind: 'voterGuide',
  key: `voterGuide:${guide.id}`,
  id: guide.id,
  title: guide.title,
  chapterId: guide.chapterId,
  chapterName: chapter.name,
  chapterWebsiteUrl: chapter.websiteUrl,
  electionId: guide.electionId,
  electionName: election?.name,
  electionDate: election?.electionDate,
  confidence: guide.confidence,
  sourceUrl: guide.sourceUrl,
  publishedAt: guide.publishedAt,
  summary: guide.summary,
};
```

- [ ] **Step 6: Add enriched data assertions**

In `web/src/data/schema.test.ts`, extend the `builds joined endorsement, measure, and guide cards` test with:

```ts
expect(getEndorsementCards()).toEqual(
  expect.arrayContaining([
    expect.objectContaining({
      kind: 'endorsement',
      key: 'endorsement:contreras-dsa-sd-2026-primary',
      candidateId: 'corinna-contreras',
      chapterId: 'dsa-san-diego',
      raceId: 'ca-48-us-house-2026-primary',
      notes: expect.stringContaining('Seeded from DSA San Diego'),
    }),
  ]),
);
expect(getMeasureCards()).toEqual(
  expect.arrayContaining([
    expect.objectContaining({
      kind: 'measure',
      key: 'measure:san-diego-measure-a-dsa-sd-yes-2026-primary',
      measureId: 'san-diego-measure-a-2026-primary',
      measureNumber: 'A',
      summary: expect.stringContaining('Local ballot measure'),
    }),
  ]),
);
expect(getVoterGuideCards()).toEqual(
  expect.arrayContaining([
    expect.objectContaining({
      kind: 'voterGuide',
      key: 'voterGuide:dsa-san-diego-2026-primary-guide',
      chapterId: 'dsa-san-diego',
      electionId: '2026-ca-primary',
      publishedAt: '2026-05-05',
    }),
  ]),
);
```

Keep the existing broad assertions in that test; do not remove them.

- [ ] **Step 7: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

## Task 2: Make Directory Cards Selectable

**Files:**
- Modify: `web/src/components/DirectoryCard.tsx`
- Test: `web/src/App.test.tsx`

- [ ] **Step 1: Replace card components with selectable versions**

Replace `web/src/components/DirectoryCard.tsx` with:

```tsx
import { SourceBadge } from './SourceBadge';
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';
import type { ReactNode } from 'react';

type SelectableCardProps = {
  selected?: boolean;
  onSelect?: () => void;
};

function cardClassName(selected: boolean | undefined): string {
  return [
    'rounded-none border-2 bg-surface p-5 text-left shadow-[4px_4px_0_var(--theme-shadow)] transition',
    selected ? 'border-primary shadow-[7px_7px_0_var(--theme-shadow)]' : 'border-border hover:-translate-y-0.5 hover:border-border-strong',
  ].join(' ');
}

function SelectionButton({ children, onSelect, selected }: SelectableCardProps & { children: ReactNode }) {
  return (
    <button
      aria-pressed={selected ? 'true' : 'false'}
      className="block w-full rounded-none text-left focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      type="button"
      onClick={onSelect}
    >
      {selected ? (
        <span className="mb-2 inline-block text-xs font-black uppercase tracking-wide text-primary">Selected</span>
      ) : null}
      {children}
    </button>
  );
}

export function EndorsementCard({ item, onSelect, selected }: { item: EndorsementCardData } & SelectableCardProps) {
  const raceLine = [item.office, item.district, item.electionName].filter(Boolean).join(' · ');

  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-2xl font-black text-foreground">{item.candidateName}</h3>
        <p className="mt-2 text-sm leading-6 text-muted">{raceLine}</p>
      </SelectionButton>
      <SourceBadge
        confidence={item.confidence}
        endorsementType={item.endorsementType}
        sourceDate={item.sourceDate}
        sourceUrl={item.sourceUrl}
      />
    </article>
  );
}

export function MeasureCard({ item, onSelect, selected }: { item: MeasureCardData } & SelectableCardProps) {
  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-2xl font-black text-foreground">{item.title}</h3>
        <p className="mt-2 text-sm leading-6 text-muted">
          {item.jurisdiction} · {item.electionName} · Position: {item.position.toUpperCase()}
        </p>
      </SelectionButton>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
    </article>
  );
}

export function VoterGuideCard({ item, onSelect, selected }: { item: VoterGuideCardData } & SelectableCardProps) {
  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-xl font-black text-foreground">{item.title}</h3>
        {item.electionName ? <p className="mt-2 text-sm leading-6 text-muted">{item.electionName}</p> : null}
        {item.summary ? <p className="mt-3 text-sm leading-6 text-muted">{item.summary}</p> : null}
      </SelectionButton>
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
    </article>
  );
}
```

- [ ] **Step 2: Preserve visual content**

Do not change card text except for the optional `Selected` label. Existing homepage assertions for `Corinna Contreras`, `Recommended`, `Official source`, `San Diego Measure A`, and `Original source` must still pass.

- [ ] **Step 3: Run App tests**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: PASS after `App.tsx` is updated in Task 4; if run before Task 4, TypeScript may still pass but selected state will not be exercised.

## Task 3: Create the Detail Panel Component

**Files:**
- Create: `web/src/components/DetailPanel.tsx`
- Test: `web/src/App.test.tsx`

- [ ] **Step 1: Create the component file**

Create `web/src/components/DetailPanel.tsx`:

```tsx
import { SourceBadge } from './SourceBadge';
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';
import { endorsementTypeLabel } from '../data/schema';

type DetailItem = EndorsementCardData | MeasureCardData | VoterGuideCardData;

type DetailPanelProps = {
  item?: DetailItem;
};

function DetailRow({ label, value }: { label: string; value?: string }) {
  if (!value) {
    return null;
  }

  return (
    <div className="border-t-2 border-border py-3">
      <dt className="text-xs font-black uppercase tracking-wide text-muted">{label}</dt>
      <dd className="mt-1 text-sm font-bold leading-6 text-foreground">{value}</dd>
    </div>
  );
}

function DetailLink({ href, label }: { href?: string; label: string }) {
  if (!href) {
    return null;
  }

  return (
    <a
      className="text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      href={href}
    >
      {label}
    </a>
  );
}

export function DetailPanel({ item }: DetailPanelProps) {
  if (!item) {
    return (
      <aside
        aria-label="Record detail"
        className="rounded-none border-2 border-dashed border-border-strong bg-surface-muted p-6 text-muted shadow-[6px_6px_0_var(--theme-shadow)]"
      >
        <p className="text-xs font-black uppercase tracking-wide text-primary">Record detail</p>
        <h2 className="mt-3 text-2xl font-black text-foreground">Select a record</h2>
        <p className="mt-3 text-sm leading-6">Choose a voter guide, endorsement, or ballot measure to inspect source-linked details.</p>
      </aside>
    );
  }

  const title = item.kind === 'endorsement' ? item.candidateName : item.title;
  const kicker = item.kind === 'endorsement' ? 'Candidate endorsement' : item.kind === 'measure' ? 'Ballot measure' : 'Voter guide';

  return (
    <aside
      aria-label="Record detail"
      aria-live="polite"
      className="rounded-none border-2 border-border-strong bg-surface p-6 shadow-[8px_8px_0_var(--theme-shadow)]"
    >
      <p className="text-xs font-black uppercase tracking-wide text-primary">{kicker}</p>
      <h2 className="mt-3 text-3xl font-black text-foreground">{title}</h2>
      <p className="mt-3 text-sm leading-6 text-muted">Recorded from source-linked public DSA electoral material.</p>

      {item.kind === 'endorsement' ? <EndorsementDetail item={item} /> : null}
      {item.kind === 'measure' ? <MeasureDetail item={item} /> : null}
      {item.kind === 'voterGuide' ? <VoterGuideDetail item={item} /> : null}
    </aside>
  );
}

function EndorsementDetail({ item }: { item: EndorsementCardData }) {
  return (
    <>
      <SourceBadge
        confidence={item.confidence}
        endorsementType={item.endorsementType}
        sourceDate={item.sourceDate}
        sourceUrl={item.sourceUrl}
      />
      <dl className="mt-5">
        <DetailRow label="Source label" value={endorsementTypeLabel(item.endorsementType)} />
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Office" value={[item.office, item.district].filter(Boolean).join(' · ')} />
        <DetailRow label="Race level" value={item.raceLevel.split('_').join(' ')} />
        <DetailRow label="Race jurisdiction" value={item.raceJurisdiction} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Candidate party" value={item.candidateParty} />
        <DetailRow label="Notes" value={item.notes} />
      </dl>
      <div className="mt-4 flex flex-wrap gap-4 text-sm font-bold">
        <DetailLink href={item.candidateCampaignUrl} label="Candidate site" />
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}

function MeasureDetail({ item }: { item: MeasureCardData }) {
  return (
    <>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
      <dl className="mt-5">
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Position" value={item.position.toUpperCase()} />
        <DetailRow label="Measure number" value={item.measureNumber} />
        <DetailRow label="Jurisdiction" value={item.jurisdiction} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Summary" value={item.summary} />
      </dl>
      <div className="mt-4 text-sm font-bold">
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}

function VoterGuideDetail({ item }: { item: VoterGuideCardData }) {
  return (
    <>
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
      <dl className="mt-5">
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Published" value={item.publishedAt} />
        <DetailRow label="Summary" value={item.summary} />
      </dl>
      <div className="mt-4 text-sm font-bold">
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}
```

- [ ] **Step 2: Run typecheck**

Run:

```bash
pnpm --dir web run format
```

Expected: PASS after imports are wired in Task 4. If run immediately, this file should still typecheck because it is not imported yet.

## Task 4: Wire Selection into the App

**Files:**
- Modify: `web/src/App.tsx`
- Test: `web/src/App.test.tsx`

- [ ] **Step 1: Import detail helpers**

In `web/src/App.tsx`, change the data import:

```ts
import { allDirectoryCards, buildDirectoryCards, type DirectoryCardKey } from './data';
```

Add:

```ts
import { DetailPanel } from './components/DetailPanel';
```

- [ ] **Step 2: Add selected-card state**

Inside `App`, add:

```ts
const [selectedCardKey, setSelectedCardKey] = useState<DirectoryCardKey | undefined>();
```

- [ ] **Step 3: Derive selected item and clear stale selection**

After `filteredCards` is created, add:

```ts
const filteredCardList = useMemo(() => allDirectoryCards(filteredCards), [filteredCards]);
const selectedCard = useMemo(
  () => filteredCardList.find((card) => card.key === selectedCardKey),
  [filteredCardList, selectedCardKey],
);

useEffect(() => {
  if (selectedCardKey && !selectedCard) {
    setSelectedCardKey(undefined);
  }
}, [selectedCard, selectedCardKey]);
```

- [ ] **Step 4: Add the detail panel layout**

Replace the three `DirectorySection` blocks with a two-column wrapper:

```tsx
<section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem] lg:items-start">
  <div className="flex flex-col gap-10">
    {/* existing noVisibleResults and DirectorySection blocks move here */}
  </div>
  <div className="lg:sticky lg:top-6">
    <DetailPanel item={selectedCard} />
  </div>
</section>
```

Keep `MethodologyPanel` after this wrapper so methodology stays below the browse/detail experience.

- [ ] **Step 5: Pass selection props to each card**

Change each card render call.

For voter guides:

```tsx
<VoterGuideCard
  item={item}
  key={item.id}
  selected={selectedCardKey === item.key}
  onSelect={() => setSelectedCardKey(item.key)}
/>
```

For endorsements:

```tsx
<EndorsementCard
  item={item}
  key={item.id}
  selected={selectedCardKey === item.key}
  onSelect={() => setSelectedCardKey(item.key)}
/>
```

For measures:

```tsx
<MeasureCard
  item={item}
  key={item.id}
  selected={selectedCardKey === item.key}
  onSelect={() => setSelectedCardKey(item.key)}
/>
```

- [ ] **Step 6: Clear selection with filters**

In the `Clear all` button handler, add:

```ts
setSelectedCardKey(undefined);
```

- [ ] **Step 7: Add App smoke assertions**

In `web/src/App.test.tsx`, add these assertions to the homepage render test:

```ts
expect(html).toContain('Record detail');
expect(html).toContain('Select a record');
expect(html).toContain('Choose a voter guide, endorsement, or ballot measure to inspect source-linked details.');
expect(html).toContain('aria-pressed="false"');
```

- [ ] **Step 8: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx src/data/schema.test.ts
```

Expected: PASS.

## Task 5: Document Detail UX Rules

**Files:**
- Modify: `docs/reference/design.md`

- [ ] **Step 1: Add detail panel section**

Append this section to `docs/reference/design.md`:

```md
## Directory detail panel

Directory cards may be selectable, but source links must remain separate clickable elements.

- Use `aria-pressed` on the selection button.
- Show selected state with a label plus border/shadow changes, not color alone.
- Keep the detail panel inline; do not use a modal for browsing records.
- The panel must include the original source link, confidence label, chapter, election/race context when available, and any notes or summaries in the data.
- The detail region should be labeled and announce changes with `aria-live="polite"`.
```

- [ ] **Step 2: Run docs-safe verification**

Run:

```bash
make verify
```

Expected: PASS.

## Task 6: Final Verification and Commit

**Files:**
- All files changed by Tasks 1-5.

- [ ] **Step 1: Run full verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 2: Inspect diff**

Run:

```bash
git status --short
git diff --stat
git diff -- web/src/data/index.ts web/src/data/schema.test.ts web/src/components/DirectoryCard.tsx web/src/components/DetailPanel.tsx web/src/App.tsx web/src/App.test.tsx docs/reference/design.md
```

Expected: only planned files changed.

- [ ] **Step 3: Commit**

Run:

```bash
git add web/src/data/index.ts web/src/data/schema.test.ts web/src/components/DirectoryCard.tsx web/src/components/DetailPanel.tsx web/src/App.tsx web/src/App.test.tsx docs/reference/design.md
git commit -m "feat: add source-linked detail panel"
```

## Acceptance Criteria

- Users can select any voter guide, endorsement, or ballot measure card.
- A detail panel shows source-linked metadata for the selected card.
- Source links remain real links and are not nested inside card buttons.
- Selected state is keyboard-accessible and visible without relying only on color.
- Browse filters and text search still work and clear stale selections.
- No backend/API changes are required.
- `make verify` passes.

## Known Follow-Ups

- Add route-backed detail URLs after the inline panel proves useful.
- Add keyboard arrow navigation between cards if browsing volume grows.
- Add richer editorial provenance fields in a later data-quality/admin slice.
