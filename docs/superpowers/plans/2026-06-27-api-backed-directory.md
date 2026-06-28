# API-Backed Directory Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Load the public DSA Slate directory from `/api/directory` when available, while preserving the checked-in California static data as the safe fallback.

**Architecture:** Keep the frontend contract centered on `DirectoryData` and `buildDirectoryCards`. Add a tiny API client that fetches JSON, validates it with existing source/data checks, and returns typed data. Add a React hook that starts from static data, upgrades to API data after mount, and keeps the static directory visible if the API fails.

**Tech Stack:** React 19, Vite, TypeScript, Vitest, Go standard-library API, Postgres seed fixture, Docker Compose.

**Status:** Implemented in this session through Task 5. Review found no blockers; follow-up runtime enum/shape hardening is noted as future work. `make verify` passed with 27 frontend tests.

---

## Scope

### In scope

- Fetch `/api/directory` from the frontend.
- Validate API JSON with the same `validateDirectoryData` rules used for static seed data.
- Render API data when the request succeeds.
- Keep rendering static checked-in California data if the API is unavailable, invalid, or returns a non-2xx response.
- Add visible but low-drama data-source status copy: loading, live API data, static fallback.
- Add Vite dev proxy for `/api` to the local Go API.
- Add tests for API success, fallback behavior, validation failure, and SSR/static initial render.

### Out of scope

- Server-side rendering.
- URL query params or routing.
- Address/ZIP lookup.
- Persisting raw voter addresses.
- Auth, admin UI, public submissions, scraping, or live results.
- Replacing the static seed immediately; static remains the launch fallback.

## File Structure

- Create `web/src/data/api.ts` — small fetch/validation client for `/api/directory`.
- Create `web/src/data/api.test.ts` — unit tests for successful fetch, HTTP failure, invalid JSON shape, and validation failure.
- Create `web/src/data/useDirectoryData.ts` — React hook that exposes `{ data, status, errorMessage }` while falling back to static data.
- Create `web/src/data/useDirectoryData.test.ts` — tests the hook's pure state helpers and fetch transition functions without adding a DOM test dependency.
- Modify `web/src/App.tsx` — replace module-level static cards with hook-driven cards and add a data-source status line.
- Modify `web/src/App.test.tsx` — assert SSR still renders static directory and fallback/status copy.
- Modify `web/vite.config.ts` — proxy `/api` to local Go API for `make dev-web`.
- Modify `README.md` — document API-backed frontend local flow.
- Modify `Makefile` — add an optional `dev-full` target for running seeded API + frontend in separate terminal guidance, without making `make dev` surprising.

## Task 1: API client and validation

**Files:**
- Create: `web/src/data/api.ts`
- Create: `web/src/data/api.test.ts`

- [ ] **Step 1: Add failing tests for the API client**

Create `web/src/data/api.test.ts`:

```ts
import { describe, expect, it, vi } from 'vitest';
import { californiaDirectoryData } from './california';
import { fetchDirectoryData } from './api';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('fetchDirectoryData', () => {
  it('returns validated directory data from the API', async () => {
    const fetcher = vi.fn(async () => jsonResponse(californiaDirectoryData));

    const data = await fetchDirectoryData(fetcher, '/api/directory');

    expect(fetcher).toHaveBeenCalledWith('/api/directory', { headers: { Accept: 'application/json' } });
    expect(data.chapters.map((chapter) => chapter.id)).toContain('dsa-san-diego');
  });

  it('throws a clear error for non-2xx responses', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ error: 'internal_server_error' }, 500));

    await expect(fetchDirectoryData(fetcher, '/api/directory')).rejects.toThrow('Directory API request failed with status 500');
  });

  it('throws validation errors for invalid directory data', async () => {
    const invalidData = {
      ...californiaDirectoryData,
      endorsements: [{ ...californiaDirectoryData.endorsements[0], sourceUrl: '' }],
    };
    const fetcher = vi.fn(async () => jsonResponse(invalidData));

    await expect(fetchDirectoryData(fetcher, '/api/directory')).rejects.toThrow('Directory API data validation failed');
  });
});
```

- [ ] **Step 2: Run the focused failing test**

Run:

```bash
pnpm --dir web run test -- src/data/api.test.ts
```

Expected: FAIL because `web/src/data/api.ts` does not exist.

- [ ] **Step 3: Implement the API client**

Create `web/src/data/api.ts`:

```ts
import { validateDirectoryData, type DirectoryData } from './schema';

export type DirectoryFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

export async function fetchDirectoryData(
  fetcher: DirectoryFetch = fetch,
  endpoint = '/api/directory',
): Promise<DirectoryData> {
  const response = await fetcher(endpoint, { headers: { Accept: 'application/json' } });

  if (!response.ok) {
    throw new Error(`Directory API request failed with status ${response.status}`);
  }

  const data = (await response.json()) as DirectoryData;
  const errors = validateDirectoryData(data);

  if (errors.length > 0) {
    throw new Error(`Directory API data validation failed:\n${errors.join('\n')}`);
  }

  return data;
}
```

- [ ] **Step 4: Verify the API client tests pass**

Run:

```bash
pnpm --dir web run test -- src/data/api.test.ts
pnpm --dir web run format
```

Expected: PASS.

## Task 2: Directory loading state and hook

**Files:**
- Create: `web/src/data/useDirectoryData.ts`
- Create: `web/src/data/useDirectoryData.test.ts`

- [ ] **Step 1: Add tests for fallback state transitions**

Create `web/src/data/useDirectoryData.test.ts`:

```ts
import { describe, expect, it } from 'vitest';
import { californiaDirectoryData } from './california';
import { directoryLoadFailed, directoryLoadStarted, directoryLoadSucceeded, initialDirectoryState } from './useDirectoryData';

describe('directory loading state helpers', () => {
  it('starts with static data while the API request is pending', () => {
    const state = initialDirectoryState(californiaDirectoryData);

    expect(state.status).toBe('loading');
    expect(state.data).toBe(californiaDirectoryData);
    expect(state.errorMessage).toBeUndefined();
  });

  it('uses API data after a successful load', () => {
    const state = directoryLoadSucceeded(initialDirectoryState(californiaDirectoryData), californiaDirectoryData);

    expect(state.status).toBe('api');
    expect(state.data).toBe(californiaDirectoryData);
  });

  it('keeps the previous static data after a failed load', () => {
    const pending = directoryLoadStarted(initialDirectoryState(californiaDirectoryData));
    const failed = directoryLoadFailed(pending, new Error('network unavailable'));

    expect(failed.status).toBe('fallback');
    expect(failed.data).toBe(californiaDirectoryData);
    expect(failed.errorMessage).toBe('network unavailable');
  });
});
```

- [ ] **Step 2: Run the focused failing test**

Run:

```bash
pnpm --dir web run test -- src/data/useDirectoryData.test.ts
```

Expected: FAIL because `web/src/data/useDirectoryData.ts` does not exist.

- [ ] **Step 3: Implement state helpers and hook**

Create `web/src/data/useDirectoryData.ts`:

```ts
import { useEffect, useState } from 'react';
import { directoryData as staticDirectoryData } from './index';
import { fetchDirectoryData, type DirectoryFetch } from './api';
import type { DirectoryData } from './schema';

export type DirectoryLoadStatus = 'loading' | 'api' | 'fallback';

export type DirectoryLoadState = {
  data: DirectoryData;
  status: DirectoryLoadStatus;
  errorMessage?: string;
};

export function initialDirectoryState(data: DirectoryData = staticDirectoryData): DirectoryLoadState {
  return { data, status: 'loading' };
}

export function directoryLoadStarted(state: DirectoryLoadState): DirectoryLoadState {
  return { data: state.data, status: 'loading' };
}

export function directoryLoadSucceeded(state: DirectoryLoadState, data: DirectoryData): DirectoryLoadState {
  return { ...state, data, status: 'api', errorMessage: undefined };
}

export function directoryLoadFailed(state: DirectoryLoadState, error: unknown): DirectoryLoadState {
  const message = error instanceof Error ? error.message : 'Directory API request failed';
  return { data: state.data, status: 'fallback', errorMessage: message };
}

export function useDirectoryData(fetcher: DirectoryFetch = fetch): DirectoryLoadState {
  const [state, setState] = useState<DirectoryLoadState>(() => initialDirectoryState());

  useEffect(() => {
    let active = true;

    setState((current) => directoryLoadStarted(current));

    fetchDirectoryData(fetcher)
      .then((data) => {
        if (active) {
          setState((current) => directoryLoadSucceeded(current, data));
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setState((current) => directoryLoadFailed(current, error));
        }
      });

    return () => {
      active = false;
    };
  }, [fetcher]);

  return state;
}
```

- [ ] **Step 4: Verify state helper tests pass**

Run:

```bash
pnpm --dir web run test -- src/data/useDirectoryData.test.ts
pnpm --dir web run format
```

Expected: PASS.

## Task 3: Wire API-backed data into the homepage

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.test.tsx`

- [ ] **Step 1: Update App tests for static initial render and data-source copy**

Modify `web/src/App.test.tsx` to include these assertions in the first test:

```ts
expect(html).toContain('Loading live directory data');
expect(html).toContain('Static fallback is visible while the API loads.');
```

Expected: the existing SSR test still checks `Corinna Contreras`, source labels, and no disabled fake search.

- [ ] **Step 2: Run the focused failing test**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: FAIL because App does not render data-source status copy yet.

- [ ] **Step 3: Replace module-level static cards with hook-driven cards**

Modify the imports and top of `web/src/App.tsx`:

```tsx
import { EndorsementCard, MeasureCard, VoterGuideCard } from './components/DirectoryCard';
import { MethodologyPanel } from './components/MethodologyPanel';
import { buildDirectoryCards } from './data';
import { filterDirectoryCards } from './data/filter';
import { useDirectoryData, type DirectoryLoadStatus } from './data/useDirectoryData';
import type { ReactNode } from 'react';
import { useMemo, useState } from 'react';

export default function App() {
  const [query, setQuery] = useState('');
  const directoryState = useDirectoryData();

  const directoryCards = useMemo(() => buildDirectoryCards(directoryState.data), [directoryState.data]);
  const filteredCards = useMemo(() => filterDirectoryCards(directoryCards, query), [directoryCards, query]);
```

Replace `directoryData` stat use with `directoryState.data`:

```tsx
<StatCard label="Seed chapters" value={directoryState.data.chapters.length.toString()} />
```

Add this line below the existing text-search helper block:

```tsx
<DirectoryDataStatus errorMessage={directoryState.errorMessage} status={directoryState.status} />
```

Add this component near the bottom of `web/src/App.tsx`:

```tsx
function DirectoryDataStatus({ errorMessage, status }: { errorMessage?: string; status: DirectoryLoadStatus }) {
  const copy: Record<DirectoryLoadStatus, { label: string; detail: string }> = {
    loading: {
      label: 'Loading live directory data',
      detail: 'Static fallback is visible while the API loads.',
    },
    api: {
      label: 'Live API data',
      detail: 'Records are loaded from the DSA Slate read API.',
    },
    fallback: {
      label: 'Static fallback data',
      detail: errorMessage
        ? `The API did not load, so the checked-in source-linked seed is shown. ${errorMessage}`
        : 'The API did not load, so the checked-in source-linked seed is shown.',
    },
  };

  return (
    <p className="mt-2 rounded-xl border border-red-800 bg-red-900/30 px-4 py-3 text-sm leading-6 text-rose-100/75">
      <span className="font-bold text-rose-50">{copy[status].label}.</span> {copy[status].detail}
    </p>
  );
}
```

- [ ] **Step 4: Verify homepage tests pass**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
pnpm --dir web run format
pnpm --dir web run lint
```

Expected: PASS.

## Task 4: Vite dev proxy and local workflow docs

**Files:**
- Modify: `web/vite.config.ts`
- Modify: `README.md`
- Modify: `Makefile`

- [ ] **Step 1: Add Vite proxy configuration**

Modify `web/vite.config.ts`:

```ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
```

- [ ] **Step 2: Add a non-surprising full-stack dev helper**

Modify `Makefile` `.PHONY` line to include `dev-full` and add:

```make
dev-full: seed ## Prepare DB seed, then print commands for API and web dev servers
	@printf '%s\n' 'Run these in two terminals:'
	@printf '%s\n' '  make dev-api'
	@printf '%s\n' '  make dev-web'
```

Keep `make dev` unchanged so it does not unexpectedly start the API or seed data beyond its current behavior.

- [ ] **Step 3: Document the API-backed frontend flow**

Add to `README.md` local commands section:

````md
### API-backed frontend

The frontend starts with checked-in California seed data and upgrades to `/api/directory` after the page loads. If the API is unavailable or returns invalid directory data, the page keeps showing the static fallback.

```bash
make dev-full   # prepares local Postgres and prints the two dev commands
make dev-api    # terminal 1: API on :8080
make dev-web    # terminal 2: Vite on :5173, proxying /api to :8080
```
````

- [ ] **Step 4: Verify config/docs changes**

Run:

```bash
make help
pnpm --dir web run format
pnpm --dir web run lint
```

Expected: PASS, and `make help` lists `dev-full`.

## Task 5: Full verification and review gate

**Files:**
- No new files unless review fixes are needed.

- [ ] **Step 1: Run frontend and full repo verification**

Run:

```bash
pnpm --dir web run test
make verify
```

Expected: PASS.

- [ ] **Step 2: Run API-backed local smoke when Docker is available**

Run:

```bash
make seed
```

Then in terminal 1:

```bash
make dev-api
```

Then in terminal 2:

```bash
make api-smoke
```

Expected: PASS. Manual browser check at `http://localhost:5173` should show `Live API data` when `make dev-web` is also running.

- [ ] **Step 3: Review acceptance checklist**

Confirm:

- Static SSR render still shows public directory content.
- Browser runtime fetches `/api/directory`.
- API success swaps to API data and shows `Live API data`.
- API failure keeps static data visible and shows `Static fallback data`.
- Invalid API data is rejected by `validateDirectoryData`.
- No raw address fields are introduced.
- Source URLs, confidence labels, endorsement/recommendation wording, and methodology copy remain visible.

## Execution Choice

Recommended execution path: **Subagent-driven**, one fresh @fixer per task, with orchestrator verification after Task 3 and Task 5.

Commit suggestion after successful verification:

```bash
git add Makefile README.md web/src/App.tsx web/src/App.test.tsx web/src/data/api.ts web/src/data/api.test.ts web/src/data/useDirectoryData.ts web/src/data/useDirectoryData.test.ts web/vite.config.ts docs/superpowers/plans/2026-06-27-api-backed-directory.md
git commit -m "feat: load directory from api with static fallback"
```
