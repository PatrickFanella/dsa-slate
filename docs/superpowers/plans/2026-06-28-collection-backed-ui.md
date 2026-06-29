# Collection-Backed UI API Plan

> Lightweight plan for the next API slice. Execute in small batches and verify with `make verify` before committing.

**Goal:** Make the React app load directory data from the existing collection endpoints instead of relying on the aggregate `/api/directory` endpoint.

**Architecture:** Keep the current UI and card-building code unchanged. Add one frontend adapter that fetches `/api/chapters`, `/api/elections`, `/api/races`, `/api/candidates`, `/api/voter-guides`, `/api/endorsements`, `/api/ballot-measures`, and `/api/measure-positions`, assembles a `DirectoryData` object, validates it with the existing validator, and preserves the static fallback path when API loading fails.

**Tech Stack:** Vite React TypeScript, existing browser `fetch`, Vitest, Go backend read API.

---

## Scope

In scope:
- Add `fetchDirectoryCollections(fetcher?: DirectoryFetch): Promise<DirectoryData>` in `web/src/data/api.ts`.
- Reuse existing collection helpers and `validateDirectoryData`.
- Reject completely empty collection responses with the same public-records error used by `fetchDirectoryData`.
- Switch `useDirectoryData` to call the collection-backed loader.
- Add tests proving the loader calls collection endpoints, validates assembled data, handles empty data, and preserves fallback state.

Out of scope:
- Backend endpoint changes.
- Server-side filtered browse controls.
- Pagination, caching, abort signals, or React Query.
- UI layout/card changes.

## Implementation Steps

1. In `web/src/data/api.ts`, extract the existing empty-directory check into a shared helper.
2. Add `fetchDirectoryCollections(fetcher = fetch)` that runs the eight collection helper requests, assembles `DirectoryData`, validates it, checks for empty public records, and returns the data.
3. Update `web/src/data/useDirectoryData.ts` to import and call `fetchDirectoryCollections` instead of `fetchDirectoryData`.
4. Extend `web/src/data/api.test.ts` with tests for successful collection assembly, validation failure, and empty collection failure.
5. Keep `web/src/data/useDirectoryData.test.ts` helper tests intact; add direct helper coverage only if needed.
6. Run `pnpm --dir web run test`, then `make verify`.

## Acceptance Criteria

- Runtime UI no longer calls `/api/directory` through `useDirectoryData`.
- Existing `/api/directory` helper remains available and tested as a compatibility path.
- Collection loading preserves source-linked data validation and static fallback behavior.
- All frontend and repo verification checks pass.
