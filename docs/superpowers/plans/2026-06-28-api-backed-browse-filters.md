# API-Backed Browse Filters Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Make structured browse filters use the existing collection API filters while preserving the current text search, card UI, and static fallback behavior.

**Architecture:** Keep `/api/directory` as a compatibility helper and keep text search client-side. Add a typed browse-filter loader that calls the collection endpoints with supported query params, assembles a referentially valid `DirectoryData`, and prunes locally only where the backend does not yet expose a matching filter. The React app owns filter state and passes it to `useDirectoryData`, while card rendering continues through `buildDirectoryCards` and `filterDirectoryCards`.

**Tech Stack:** Vite React TypeScript, Vitest, existing Go read API collection filters.

---

## Current State

- `web/src/data/api.ts` has collection helpers and `fetchDirectoryCollections(fetcher)`.
- `web/src/data/useDirectoryData.ts` loads collection-backed `DirectoryData` with static fallback semantics.
- `web/src/App.tsx` has one visible filter: free-text `query`.
- `web/src/data/filter.ts` performs only client-side text search over built cards.
- Backend-supported structured filters are:
  - `/api/races?electionId&jurisdiction&level`
  - `/api/voter-guides?chapterId&electionId`
  - `/api/endorsements?chapterId&electionId&raceId`
  - `/api/ballot-measures?electionId`
  - `/api/measure-positions?chapterId&measureId`

## File Map

- Modify: `web/src/data/api.ts`
  - Add `BrowseFilters` type.
  - Add `fetchBrowseDirectoryData(filters, fetcher)` orchestration helper.
  - Keep `fetchDirectoryCollections` as the unfiltered default path.
- Modify: `web/src/data/useDirectoryData.ts`
  - Accept optional browse filters.
  - Pass filters to the browse loader.
  - Preserve `loading | api | fallback` status behavior.
- Modify: `web/src/App.tsx`
  - Add controlled `<select>` filters for chapter, election, and race level.
  - Keep text search as the final client-side pass.
  - Keep no-address-copy and result counts.
- Modify: `web/src/data/api.test.ts`
  - Cover query routing, local pruning, validation, and empty-data behavior.
- Modify: `web/src/data/useDirectoryData.test.ts`
  - Cover helper-level state behavior if the hook signature changes exported helper behavior.
- Modify: `web/src/App.test.tsx`
  - Smoke-test new browse controls and existing homepage content.
- Modify: `docs/reference/api.md`
  - Document that the frontend uses server-supported structured filters and keeps text search client-side.

## Task 1: Add a Browse Filter API Adapter

**Files:**
- Modify: `web/src/data/api.ts`
- Test: `web/src/data/api.test.ts`

- [ ] **Step 1: Add the failing test for endpoint routing**

Add this test inside `describe('fetchDirectoryCollections', ...)` or a new `describe('fetchBrowseDirectoryData', ...)` block in `web/src/data/api.test.ts`:

```ts
it('routes supported browse filters to collection endpoints', async () => {
  const fetcher = vi.fn(async (input: RequestInfo | URL) => {
    switch (String(input)) {
      case '/api/chapters':
        return jsonResponse(californiaDirectoryData.chapters);
      case '/api/elections':
        return jsonResponse(californiaDirectoryData.elections);
      case '/api/races?electionId=2026-ca-primary&level=federal':
        return jsonResponse(californiaDirectoryData.races.filter((race) => race.electionId === '2026-ca-primary' && race.level === 'federal'));
      case '/api/candidates':
        return jsonResponse(californiaDirectoryData.candidates);
      case '/api/voter-guides?chapterId=dsa-san-diego&electionId=2026-ca-primary':
        return jsonResponse(californiaDirectoryData.voterGuides.filter((guide) => guide.chapterId === 'dsa-san-diego' && guide.electionId === '2026-ca-primary'));
      case '/api/endorsements?chapterId=dsa-san-diego&electionId=2026-ca-primary':
        return jsonResponse(californiaDirectoryData.endorsements.filter((endorsement) => endorsement.chapterId === 'dsa-san-diego'));
      case '/api/ballot-measures?electionId=2026-ca-primary':
        return jsonResponse(californiaDirectoryData.ballotMeasures.filter((measure) => measure.electionId === '2026-ca-primary'));
      case '/api/measure-positions?chapterId=dsa-san-diego':
        return jsonResponse(californiaDirectoryData.measurePositions.filter((position) => position.chapterId === 'dsa-san-diego'));
      default:
        return jsonResponse({ error: 'unexpected endpoint' }, 404);
    }
  });

  await fetchBrowseDirectoryData(
    { chapterId: 'dsa-san-diego', electionId: '2026-ca-primary', raceLevel: 'federal' },
    fetcher,
  );

  expect(fetcher).toHaveBeenCalledWith('/api/races?electionId=2026-ca-primary&level=federal', {
    headers: { Accept: 'application/json' },
  });
  expect(fetcher).toHaveBeenCalledWith('/api/voter-guides?chapterId=dsa-san-diego&electionId=2026-ca-primary', {
    headers: { Accept: 'application/json' },
  });
  expect(fetcher).toHaveBeenCalledWith('/api/endorsements?chapterId=dsa-san-diego&electionId=2026-ca-primary', {
    headers: { Accept: 'application/json' },
  });
  expect(fetcher).toHaveBeenCalledWith('/api/ballot-measures?electionId=2026-ca-primary', {
    headers: { Accept: 'application/json' },
  });
  expect(fetcher).toHaveBeenCalledWith('/api/measure-positions?chapterId=dsa-san-diego', {
    headers: { Accept: 'application/json' },
  });
});
```

Expected now: TypeScript fails because `fetchBrowseDirectoryData` does not exist.

- [ ] **Step 2: Add the failing test for local pruning**

Add this test to prove the adapter keeps assembled data referentially valid when the backend cannot filter by race level directly on endorsements:

```ts
it('prunes endorsements and measure positions that reference filtered-out records', async () => {
  const includedRace = californiaDirectoryData.races[0];
  const excludedRace = californiaDirectoryData.races.find((race) => race.id !== includedRace?.id);
  const includedMeasure = californiaDirectoryData.ballotMeasures[0];
  const excludedMeasure = { ...includedMeasure, id: 'filtered-out-measure' };

  if (!includedRace || !excludedRace || !excludedMeasure) {
    throw new Error('test fixture must include multiple races and measures');
  }

  const fetcher = vi.fn(async (input: RequestInfo | URL) => {
    switch (String(input)) {
      case '/api/chapters':
        return jsonResponse(californiaDirectoryData.chapters);
      case '/api/elections':
        return jsonResponse(californiaDirectoryData.elections);
      case '/api/races?level=federal':
        return jsonResponse([includedRace]);
      case '/api/candidates':
        return jsonResponse(californiaDirectoryData.candidates);
      case '/api/voter-guides':
        return jsonResponse(californiaDirectoryData.voterGuides);
      case '/api/endorsements':
        return jsonResponse([
          { ...californiaDirectoryData.endorsements[0], raceId: includedRace.id },
          { ...californiaDirectoryData.endorsements[0], id: 'filtered-out-endorsement', raceId: excludedRace.id },
        ]);
      case '/api/ballot-measures':
        return jsonResponse([includedMeasure]);
      case '/api/measure-positions':
        return jsonResponse([
          { ...californiaDirectoryData.measurePositions[0], measureId: includedMeasure.id },
          { ...californiaDirectoryData.measurePositions[0], id: 'filtered-out-position', measureId: excludedMeasure.id },
        ]);
      default:
        return jsonResponse({ error: 'unexpected endpoint' }, 404);
    }
  });

  const data = await fetchBrowseDirectoryData({ raceLevel: 'federal' }, fetcher);

  expect(data.races).toEqual([includedRace]);
  expect(data.endorsements.map((endorsement) => endorsement.raceId)).toEqual([includedRace.id]);
  expect(data.ballotMeasures).toEqual([includedMeasure]);
  expect(data.measurePositions.map((position) => position.measureId)).toEqual([includedMeasure.id]);
});
```

- [ ] **Step 3: Run the focused failing tests**

Run:

```bash
pnpm --dir web run test -- src/data/api.test.ts
```

Expected: FAIL with `fetchBrowseDirectoryData` missing.

- [ ] **Step 4: Implement `BrowseFilters` and the adapter**

In `web/src/data/api.ts`, add this near the existing `QueryValue` type:

```ts
export type BrowseFilters = {
  chapterId?: string;
  electionId?: string;
  raceLevel?: DirectoryData['races'][number]['level'];
};
```

Add these helpers below `fetchDirectoryCollections`:

```ts
function compactFilters<T extends Record<string, string | undefined>>(filters: T): T {
  return Object.fromEntries(Object.entries(filters).filter(([, value]) => value !== undefined && value !== '')) as T;
}

function filterToIds<T extends { id: string }>(records: T[]): Set<string> {
  return new Set(records.map((record) => record.id));
}

export async function fetchBrowseDirectoryData(
  filters: BrowseFilters = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData> {
  const raceFilters = compactFilters({ electionId: filters.electionId, level: filters.raceLevel });
  const chapterElectionFilters = compactFilters({ chapterId: filters.chapterId, electionId: filters.electionId });
  const electionFilters = compactFilters({ electionId: filters.electionId });
  const chapterFilters = compactFilters({ chapterId: filters.chapterId });

  const [chapters, elections, races, candidates, voterGuides, rawEndorsements, ballotMeasures, rawMeasurePositions] =
    await Promise.all([
      fetchChapters(fetcher),
      fetchElections(fetcher),
      fetchRaces(raceFilters, fetcher),
      fetchCandidates(fetcher),
      fetchVoterGuides(chapterElectionFilters, fetcher),
      fetchEndorsements(chapterElectionFilters, fetcher),
      fetchBallotMeasures(electionFilters, fetcher),
      fetchMeasurePositions(chapterFilters, fetcher),
    ]);

  const raceIds = filterToIds(races);
  const measureIds = filterToIds(ballotMeasures);

  const endorsements = rawEndorsements.filter((endorsement) => raceIds.has(endorsement.raceId));
  const measurePositions = rawMeasurePositions.filter((position) => measureIds.has(position.measureId));

  return validateDirectoryApiData(
    {
      chapters,
      elections,
      races,
      candidates,
      voterGuides,
      endorsements,
      ballotMeasures,
      measurePositions,
    },
    'Directory API',
  );
}
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/data/api.test.ts
```

Expected: PASS.

## Task 2: Wire Browse Filters into the Data Hook

**Files:**
- Modify: `web/src/data/useDirectoryData.ts`
- Test: `web/src/data/useDirectoryData.test.ts`

- [ ] **Step 1: Add a hook input type**

In `web/src/data/useDirectoryData.ts`, change the import:

```ts
import { fetchBrowseDirectoryData, type BrowseFilters, type DirectoryFetch } from './api';
```

Change the hook signature:

```ts
export function useDirectoryData(fetcher: DirectoryFetch = fetch, filters: BrowseFilters = {}): DirectoryLoadState {
```

Change the fetch call:

```ts
fetchBrowseDirectoryData(filters, fetcher)
```

Change the effect dependency list:

```ts
}, [fetcher, filters]);
```

- [ ] **Step 2: Prevent accidental infinite refetching**

Because objects in React dependency arrays must be stable, the caller must memoize `filters`. Add this comment above the hook signature:

```ts
// Callers must pass a memoized filters object; otherwise React will refetch on every render.
```

- [ ] **Step 3: Run existing data hook tests**

Run:

```bash
pnpm --dir web run test -- src/data/useDirectoryData.test.ts
```

Expected: PASS. The existing helper tests do not need to mount the hook.

## Task 3: Add Structured Browse Controls to the App

**Files:**
- Modify: `web/src/App.tsx`
- Test: `web/src/App.test.tsx`

- [ ] **Step 1: Add filter state and memoized API filters**

In `web/src/App.tsx`, add the import:

```ts
import type { BrowseFilters } from './data/api';
```

Inside `App`, add state after `query`:

```ts
const [chapterId, setChapterId] = useState('');
const [electionId, setElectionId] = useState('');
const [raceLevel, setRaceLevel] = useState('');
```

Add this memo before `useDirectoryData`:

```ts
const browseFilters = useMemo<BrowseFilters>(
  () => ({
    chapterId: chapterId || undefined,
    electionId: electionId || undefined,
    raceLevel: (raceLevel || undefined) as BrowseFilters['raceLevel'],
  }),
  [chapterId, electionId, raceLevel],
);
```

Change the hook call:

```ts
const directoryState = useDirectoryData(fetch, browseFilters);
```

- [ ] **Step 2: Add option lists**

Below the existing `directoryCards` memo, add:

```ts
const chapterOptions = useMemo(
  () => [...directoryState.data.chapters].sort((left, right) => left.name.localeCompare(right.name)),
  [directoryState.data.chapters],
);
const electionOptions = useMemo(
  () => [...directoryState.data.elections].sort((left, right) => left.name.localeCompare(right.name)),
  [directoryState.data.elections],
);
const raceLevelOptions = [
  { value: 'federal', label: 'Federal' },
  { value: 'statewide', label: 'Statewide' },
  { value: 'state_legislature', label: 'State legislature' },
  { value: 'local', label: 'Local' },
  { value: 'unknown', label: 'Unknown' },
] as const;
```

- [ ] **Step 3: Add controls below the text search input**

Insert this JSX below the current text search `<div className="mt-8 grid ...">...</div>` block:

```tsx
<div className="mt-4 grid gap-3 rounded-none border-2 border-border bg-surface-muted p-4 md:grid-cols-3">
  <label className="flex flex-col gap-2 text-sm font-bold text-foreground" htmlFor="chapter-filter">
    Chapter
    <select
      id="chapter-filter"
      className="rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      value={chapterId}
      onChange={(event) => setChapterId(event.target.value)}
    >
      <option value="">All chapters</option>
      {chapterOptions.map((chapter) => (
        <option key={chapter.id} value={chapter.id}>
          {chapter.name}
        </option>
      ))}
    </select>
  </label>

  <label className="flex flex-col gap-2 text-sm font-bold text-foreground" htmlFor="election-filter">
    Election
    <select
      id="election-filter"
      className="rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      value={electionId}
      onChange={(event) => setElectionId(event.target.value)}
    >
      <option value="">All elections</option>
      {electionOptions.map((election) => (
        <option key={election.id} value={election.id}>
          {election.name}
        </option>
      ))}
    </select>
  </label>

  <label className="flex flex-col gap-2 text-sm font-bold text-foreground" htmlFor="race-level-filter">
    Race level
    <select
      id="race-level-filter"
      className="rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      value={raceLevel}
      onChange={(event) => setRaceLevel(event.target.value)}
    >
      <option value="">All levels</option>
      {raceLevelOptions.map((level) => (
        <option key={level.value} value={level.value}>
          {level.label}
        </option>
      ))}
    </select>
  </label>
</div>
```

- [ ] **Step 4: Add a clear-all button**

Change the existing Clear button `onClick` handler to clear all filters:

```tsx
onClick={() => {
  setQuery('');
  setChapterId('');
  setElectionId('');
  setRaceLevel('');
}}
```

Change the button text from `Clear` to:

```tsx
Clear all
```

- [ ] **Step 5: Update the browse copy**

Change the paragraph text at `web/src/App.tsx:117` from:

```tsx
<p>Text search only — not address lookup. Raw addresses are not stored.</p>
```

to:

```tsx
<p>Browse filters use public API fields. Text search stays local. Raw addresses are not stored.</p>
```

- [ ] **Step 6: Update the App smoke test**

In `web/src/App.test.tsx`, add assertions:

```ts
expect(screen.getByLabelText('Chapter')).toBeTruthy();
expect(screen.getByLabelText('Election')).toBeTruthy();
expect(screen.getByLabelText('Race level')).toBeTruthy();
expect(screen.getByRole('button', { name: /clear all/i })).toBeTruthy();
```

- [ ] **Step 7: Run App tests**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: PASS.

## Task 4: Document API-Backed Browse Semantics

**Files:**
- Modify: `docs/reference/api.md`
- Optionally modify: `README.md`

- [ ] **Step 1: Add frontend browse note**

Append this section to `docs/reference/api.md`:

```md
## Frontend browse behavior

The React app uses collection endpoints for structured browse filters:

- chapter filters map to voter guides, endorsements, and measure positions.
- election filters map to races, voter guides, endorsements, and ballot measures.
- race-level filters map to races; the frontend prunes endorsements whose races are filtered out because endorsements do not currently expose a direct `level` query parameter.
- free-text search remains client-side over rendered card fields.

The app never sends or stores raw voter addresses.
```

- [ ] **Step 2: Run documentation-safe verification**

Run:

```bash
make verify
```

Expected: PASS.

## Task 5: Final Verification and Commit

**Files:**
- All files changed by Tasks 1-4.

- [ ] **Step 1: Run full verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 2: Inspect the diff**

Run:

```bash
git status --short
git diff --stat
git diff -- web/src/data/api.ts web/src/data/useDirectoryData.ts web/src/App.tsx web/src/data/api.test.ts web/src/App.test.tsx docs/reference/api.md
```

Expected: only the planned files changed.

- [ ] **Step 3: Commit**

Run:

```bash
git add web/src/data/api.ts web/src/data/useDirectoryData.ts web/src/App.tsx web/src/data/api.test.ts web/src/data/useDirectoryData.test.ts web/src/App.test.tsx docs/reference/api.md
git commit -m "feat: add api-backed browse filters"
```

## Acceptance Criteria

- The app exposes chapter, election, and race-level browse filters.
- Structured filters use existing API collection query params where supported.
- Text search remains client-side and works after structured filtering.
- Static fallback behavior remains intact if API loading fails.
- No backend API changes are required for this slice.
- `make verify` passes.

## Known Follow-Ups

- Add backend text search only if data volume grows enough to require it.
- Add direct endorsement `level` filtering only if frontend pruning becomes inefficient.
- Add direct chapter filtering for ballot measures only if the product needs measure lists independent of measure positions.
