# Source/Data Hardening Implementation Plan

Status: implemented in this session with review fixes for strict dates, required voterGuide chapterId, and validation-safe card helpers.

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Strengthen the trusted-directory seed-data contract so bad source data fails tests before it reaches the public UI.

**Architecture:** Keep validation local to the static TypeScript data layer for this slice. `web/src/data/schema.ts` remains the source of record for types, controlled labels, and `validateDirectoryData`; `web/src/data/index.ts` should validate before building cards so UI helpers no longer discover broken references at render time. Documentation explains what the validator enforces and what still requires human editorial review.

**Tech Stack:** Vite React TypeScript, Vitest, checked-in TypeScript seed data, Markdown docs, existing `make verify` workflow.

---

## Scope

Implement source/data hardening only:

- duplicate ID checks per collection
- required source URL checks for guides, endorsements, and measure positions
- URL syntax checks for source and official/campaign URLs
- referential integrity checks for all `*Id` fields used by cards and SQL contract
- ISO date checks for `electionDate`, `publishedAt`, and `sourceDate`
- source-date guidance for public endorsement records
- card helper preflight validation so broken data fails with one validation error block
- docs and tests for the enforced contract

Out of scope:

- live source fetching or link health checks
- scraping
- database-backed data loading
- admin/editor UI
- public submission flow
- address or ZIP lookup
- validating the truth of a source claim beyond local metadata rules

## Current files

- Modify `web/src/data/schema.ts` — expand validation helpers.
- Modify `web/src/data/schema.test.ts` — add focused validation tests.
- Modify `web/src/data/index.ts` — validate directory data before card joins.
- Modify `docs/reference/editorial-data.md` — document validator rules and human review limits.
- Optionally modify `README.md` only if a new command or workflow is introduced. This plan does not require one.

---

## Task 1: Expand data validation helpers

**Files:**
- Modify: `web/src/data/schema.ts`
- Test: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add failing tests for duplicate IDs**

Add this test inside `describe('directory schema helpers', ...)` in `web/src/data/schema.test.ts`:

```ts
it('reports duplicate IDs by collection', () => {
  const data: DirectoryData = {
    chapters: [
      { id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' },
      { id: 'dsa-san-diego', name: 'Duplicate DSA San Diego', state: 'CA' },
    ],
    elections: [],
    races: [],
    candidates: [],
    voterGuides: [],
    endorsements: [],
    ballotMeasures: [],
    measurePositions: [],
  };

  expect(validateDirectoryData(data)).toContain('chapters[dsa-san-diego] duplicate id');
});
```

- [ ] **Step 2: Run test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: FAIL because duplicate IDs are not yet checked.

- [ ] **Step 3: Add duplicate ID validation**

In `web/src/data/schema.ts`, add this helper above `validateDirectoryData`:

```ts
function validateUniqueIds(collectionName: string, records: { id: string }[], errors: string[]) {
  const seen = new Set<string>();

  for (const record of records) {
    if (seen.has(record.id)) {
      errors.push(`${collectionName}[${record.id}] duplicate id`);
      continue;
    }

    seen.add(record.id);
  }
}
```

Then call it at the top of `validateDirectoryData`:

```ts
validateUniqueIds('chapters', data.chapters, errors);
validateUniqueIds('elections', data.elections, errors);
validateUniqueIds('races', data.races, errors);
validateUniqueIds('candidates', data.candidates, errors);
validateUniqueIds('voterGuides', data.voterGuides, errors);
validateUniqueIds('endorsements', data.endorsements, errors);
validateUniqueIds('ballotMeasures', data.ballotMeasures, errors);
validateUniqueIds('measurePositions', data.measurePositions, errors);
```

- [ ] **Step 4: Run test and confirm pass**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

---

## Task 2: Validate URLs and dates

**Files:**
- Modify: `web/src/data/schema.ts`
- Test: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add failing tests for invalid URLs and dates**

Add this test inside `describe('directory schema helpers', ...)`:

```ts
it('reports invalid URLs and non-ISO dates', () => {
  const data: DirectoryData = {
    chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA', websiteUrl: 'not-a-url' }],
    elections: [{ id: '2026-ca-primary', name: '2026 California Primary', electionDate: '06/02/2026' }],
    races: [],
    candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras', campaignUrl: 'corinna.example' }],
    voterGuides: [
      {
        id: 'dsa-san-diego-2026-primary-guide',
        title: "DSA San Diego's June 2026 Primary Voter Guide",
        sourceUrl: 'dsasandiego.org/guide',
        confidence: 'official_source',
        publishedAt: 'May 5, 2026',
      },
    ],
    endorsements: [],
    ballotMeasures: [],
    measurePositions: [],
  };

  expect(validateDirectoryData(data)).toEqual(
    expect.arrayContaining([
      'chapters[dsa-san-diego].websiteUrl must be an absolute http(s) URL',
      'elections[2026-ca-primary].electionDate must be YYYY-MM-DD',
      'candidates[corinna-contreras].campaignUrl must be an absolute http(s) URL',
      'voterGuides[dsa-san-diego-2026-primary-guide].sourceUrl must be an absolute http(s) URL',
      'voterGuides[dsa-san-diego-2026-primary-guide].publishedAt must be YYYY-MM-DD',
    ]),
  );
});
```

- [ ] **Step 2: Run test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: FAIL because URL/date syntax is not yet checked.

- [ ] **Step 3: Add URL/date helpers and validation calls**

In `web/src/data/schema.ts`, add helpers above `validateDirectoryData`:

```ts
function isAbsoluteHttpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

function isIsoDate(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) && !Number.isNaN(Date.parse(`${value}T00:00:00Z`));
}

function validateOptionalUrl(recordName: string, fieldName: string, value: string | undefined, errors: string[]) {
  if (value && !isAbsoluteHttpUrl(value)) {
    errors.push(`${recordName}.${fieldName} must be an absolute http(s) URL`);
  }
}

function validateRequiredUrl(recordName: string, fieldName: string, value: string, errors: string[]) {
  if (!value.trim()) {
    errors.push(`${recordName} must include ${fieldName}`);
    return;
  }

  if (!isAbsoluteHttpUrl(value)) {
    errors.push(`${recordName}.${fieldName} must be an absolute http(s) URL`);
  }
}

function validateOptionalIsoDate(recordName: string, fieldName: string, value: string | undefined, errors: string[]) {
  if (value && !isIsoDate(value)) {
    errors.push(`${recordName}.${fieldName} must be YYYY-MM-DD`);
  }
}
```

Then add validation calls in `validateDirectoryData`:

```ts
for (const chapter of data.chapters) {
  validateOptionalUrl(`chapters[${chapter.id}]`, 'websiteUrl', chapter.websiteUrl, errors);
  validateOptionalUrl(`chapters[${chapter.id}]`, 'officialDsaUrl', chapter.officialDsaUrl, errors);
}

for (const election of data.elections) {
  validateOptionalIsoDate(`elections[${election.id}]`, 'electionDate', election.electionDate, errors);
  validateOptionalUrl(`elections[${election.id}]`, 'officialUrl', election.officialUrl, errors);
}

for (const candidate of data.candidates) {
  validateOptionalUrl(`candidates[${candidate.id}]`, 'campaignUrl', candidate.campaignUrl, errors);
  validateOptionalUrl(`candidates[${candidate.id}]`, 'photoUrl', candidate.photoUrl, errors);
}
```

Replace existing source URL checks with `validateRequiredUrl` and add date checks:

```ts
for (const guide of data.voterGuides) {
  validateRequiredUrl(`voterGuides[${guide.id}]`, 'sourceUrl', guide.sourceUrl, errors);
  validateOptionalIsoDate(`voterGuides[${guide.id}]`, 'publishedAt', guide.publishedAt, errors);
}

for (const endorsement of data.endorsements) {
  validateRequiredUrl(`endorsements[${endorsement.id}]`, 'sourceUrl', endorsement.sourceUrl, errors);
  validateOptionalIsoDate(`endorsements[${endorsement.id}]`, 'sourceDate', endorsement.sourceDate, errors);
}

for (const position of data.measurePositions) {
  validateRequiredUrl(`measurePositions[${position.id}]`, 'sourceUrl', position.sourceUrl, errors);
}
```

- [ ] **Step 4: Run test and confirm pass**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

---

## Task 3: Validate referential integrity

**Files:**
- Modify: `web/src/data/schema.ts`
- Test: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add failing tests for broken references**

Add this test inside `describe('directory schema helpers', ...)`:

```ts
it('reports broken record references before cards are built', () => {
  const data: DirectoryData = {
    chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
    elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
    races: [{ id: 'ca-48-us-house-2026-primary', electionId: 'missing-election', office: 'U.S. House', level: 'federal', jurisdiction: 'CA-48' }],
    candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras' }],
    voterGuides: [
      {
        id: 'dsa-san-diego-2026-primary-guide',
        title: "DSA San Diego's June 2026 Primary Voter Guide",
        chapterId: 'missing-chapter',
        electionId: 'missing-election',
        sourceUrl: 'https://dsasandiego.org/guide',
        confidence: 'official_source',
      },
    ],
    endorsements: [
      {
        id: 'contreras-dsa-sd-2026-primary',
        candidateId: 'missing-candidate',
        raceId: 'missing-race',
        chapterId: 'missing-chapter',
        voterGuideId: 'missing-guide',
        endorsementType: 'recommended',
        sourceUrl: 'https://dsasandiego.org/guide',
        confidence: 'official_source',
      },
    ],
    ballotMeasures: [{ id: 'san-diego-measure-a-2026-primary', electionId: 'missing-election', title: 'San Diego Measure A', jurisdiction: 'San Diego, CA' }],
    measurePositions: [
      {
        id: 'san-diego-measure-a-dsa-sd-yes-2026-primary',
        measureId: 'missing-measure',
        chapterId: 'missing-chapter',
        position: 'yes',
        sourceUrl: 'https://dsasandiego.org/guide',
        confidence: 'official_source',
      },
    ],
  };

  expect(validateDirectoryData(data)).toEqual(
    expect.arrayContaining([
      'races[ca-48-us-house-2026-primary].electionId references missing elections[missing-election]',
      'voterGuides[dsa-san-diego-2026-primary-guide].chapterId references missing chapters[missing-chapter]',
      'endorsements[contreras-dsa-sd-2026-primary].candidateId references missing candidates[missing-candidate]',
      'endorsements[contreras-dsa-sd-2026-primary].voterGuideId references missing voterGuides[missing-guide]',
      'ballotMeasures[san-diego-measure-a-2026-primary].electionId references missing elections[missing-election]',
      'measurePositions[san-diego-measure-a-dsa-sd-yes-2026-primary].measureId references missing ballotMeasures[missing-measure]',
    ]),
  );
});
```

- [ ] **Step 2: Run test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: FAIL because broken references are not yet checked.

- [ ] **Step 3: Add reference validation helpers and calls**

In `web/src/data/schema.ts`, add helpers above `validateDirectoryData`:

```ts
function idSet(records: { id: string }[]): Set<string> {
  return new Set(records.map((record) => record.id));
}

function validateRequiredReference(
  recordName: string,
  fieldName: string,
  value: string,
  targetCollectionName: string,
  targetIds: Set<string>,
  errors: string[],
) {
  if (!targetIds.has(value)) {
    errors.push(`${recordName}.${fieldName} references missing ${targetCollectionName}[${value}]`);
  }
}

function validateOptionalReference(
  recordName: string,
  fieldName: string,
  value: string | undefined,
  targetCollectionName: string,
  targetIds: Set<string>,
  errors: string[],
) {
  if (value && !targetIds.has(value)) {
    errors.push(`${recordName}.${fieldName} references missing ${targetCollectionName}[${value}]`);
  }
}
```

Inside `validateDirectoryData`, after duplicate ID checks, create ID sets:

```ts
const chapterIds = idSet(data.chapters);
const electionIds = idSet(data.elections);
const raceIds = idSet(data.races);
const candidateIds = idSet(data.candidates);
const voterGuideIds = idSet(data.voterGuides);
const ballotMeasureIds = idSet(data.ballotMeasures);
```

Then add reference checks inside existing loops:

```ts
for (const race of data.races) {
  validateRequiredReference(`races[${race.id}]`, 'electionId', race.electionId, 'elections', electionIds, errors);
}

for (const guide of data.voterGuides) {
  validateOptionalReference(`voterGuides[${guide.id}]`, 'chapterId', guide.chapterId, 'chapters', chapterIds, errors);
  validateOptionalReference(`voterGuides[${guide.id}]`, 'electionId', guide.electionId, 'elections', electionIds, errors);
}

for (const endorsement of data.endorsements) {
  validateRequiredReference(`endorsements[${endorsement.id}]`, 'candidateId', endorsement.candidateId, 'candidates', candidateIds, errors);
  validateRequiredReference(`endorsements[${endorsement.id}]`, 'raceId', endorsement.raceId, 'races', raceIds, errors);
  validateRequiredReference(`endorsements[${endorsement.id}]`, 'chapterId', endorsement.chapterId, 'chapters', chapterIds, errors);
  validateOptionalReference(`endorsements[${endorsement.id}]`, 'voterGuideId', endorsement.voterGuideId, 'voterGuides', voterGuideIds, errors);
}

for (const measure of data.ballotMeasures) {
  validateRequiredReference(`ballotMeasures[${measure.id}]`, 'electionId', measure.electionId, 'elections', electionIds, errors);
}

for (const position of data.measurePositions) {
  validateRequiredReference(`measurePositions[${position.id}]`, 'measureId', position.measureId, 'ballotMeasures', ballotMeasureIds, errors);
  validateRequiredReference(`measurePositions[${position.id}]`, 'chapterId', position.chapterId, 'chapters', chapterIds, errors);
}
```

- [ ] **Step 4: Run test and confirm pass**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: PASS.

---

## Task 4: Fail card helpers before rendering broken data

**Files:**
- Modify: `web/src/data/index.ts`
- Test: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add a failing test for validation preflight**

Modify the import from `./index` in `web/src/data/schema.test.ts` to include `buildDirectoryCards`:

```ts
import { buildDirectoryCards, getEndorsementCards, getMeasureCards, getVoterGuideCards } from './index';
```

Add this test inside `describe('directory schema helpers', ...)`:

```ts
it('fails card building with validation errors instead of render-time joins', () => {
  const data: DirectoryData = {
    chapters: [],
    elections: [],
    races: [],
    candidates: [],
    voterGuides: [],
    endorsements: [
      {
        id: 'broken-endorsement',
        candidateId: 'missing-candidate',
        raceId: 'missing-race',
        chapterId: 'missing-chapter',
        endorsementType: 'recommended',
        sourceUrl: 'https://example.org/source',
        confidence: 'needs_review',
      },
    ],
    ballotMeasures: [],
    measurePositions: [],
  };

  expect(() => buildDirectoryCards(data)).toThrow(
    'Directory data validation failed:\nendorsements[broken-endorsement].candidateId references missing candidates[missing-candidate]',
  );
});
```

- [ ] **Step 2: Run test and confirm failure**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: FAIL because `buildDirectoryCards` does not exist.

- [ ] **Step 3: Add card-building preflight API**

In `web/src/data/index.ts`, update imports:

```ts
import { californiaDirectoryData } from './california';
import { validateDirectoryData, type ConfidenceLevel, type DirectoryData, type EndorsementType } from './schema';
```

Add this exported type after card data types:

```ts
export type DirectoryCards = {
  endorsements: EndorsementCardData[];
  measures: MeasureCardData[];
  voterGuides: VoterGuideCardData[];
};
```

Add this function before the existing `get*Cards` exports:

```ts
export function assertValidDirectoryData(data: DirectoryData) {
  const errors = validateDirectoryData(data);

  if (errors.length > 0) {
    throw new Error(`Directory data validation failed:\n${errors.join('\n')}`);
  }
}

export function buildDirectoryCards(data: DirectoryData): DirectoryCards {
  assertValidDirectoryData(data);

  return {
    endorsements: getEndorsementCards(data),
    measures: getMeasureCards(data),
    voterGuides: getVoterGuideCards(data),
  };
}
```

Change existing helper signatures to accept a defaulted data argument:

```ts
export function getEndorsementCards(data: DirectoryData = directoryData): EndorsementCardData[] {
  return data.endorsements.map((endorsement) => {
    const candidate = data.candidates.find((item) => item.id === endorsement.candidateId);
    const race = data.races.find((item) => item.id === endorsement.raceId);
    const election = data.elections.find((item) => item.id === race?.electionId);
    const chapter = data.chapters.find((item) => item.id === endorsement.chapterId);
    // keep existing return body
  });
}
```

Apply the same `data: DirectoryData = directoryData` pattern to `getMeasureCards` and `getVoterGuideCards`, replacing `directoryData` lookups with `data`.

Optional but recommended: in `App.tsx`, replace separate helper calls with:

```ts
const directoryCards = buildDirectoryCards(directoryData);
const endorsementCards = directoryCards.endorsements;
const measureCards = directoryCards.measures;
const voterGuideCards = directoryCards.voterGuides;
```

- [ ] **Step 4: Run tests and confirm pass**

Run:

```bash
pnpm --dir web run test
```

Expected: PASS.

---

## Task 5: Document the enforced seed contract

**Files:**
- Modify: `docs/reference/editorial-data.md`
- Optionally modify: `docs/product.md` only if wording contradicts this contract.

- [ ] **Step 1: Add validator contract docs**

Append this section to `docs/reference/editorial-data.md`:

```md
## Automated seed validation

`web/src/data/schema.ts` enforces local metadata integrity for checked-in seed data.

The validator rejects:

- duplicate IDs within any collection
- missing `sourceUrl` on voter guides, endorsements, and measure positions
- non-HTTP(S) source, website, official, campaign, or photo URLs
- non-ISO dates; use `YYYY-MM-DD`
- references to missing chapters, elections, races, candidates, voter guides, or ballot measures

The validator does not prove that an endorsement is true, current, or legally usable. Human review still must confirm that the source page is public, official enough for its confidence label, and quoted without upgrading “recommended” to “endorsed.”
```

- [ ] **Step 2: Run docs-adjacent checks**

Run:

```bash
test -f docs/reference/editorial-data.md
pnpm --dir web run format
```

Expected: PASS.

---

## Task 6: Final verification and review

**Files:**
- Review: `web/src/data/schema.ts`
- Review: `web/src/data/schema.test.ts`
- Review: `web/src/data/index.ts`
- Review: `docs/reference/editorial-data.md`

- [ ] **Step 1: Run full verification**

Run:

```bash
make verify
```

Expected: PASS.

- [ ] **Step 2: Run focused source-contract review**

Review these invariants manually or via `review-quality`/`oracle`:

- every validation error is deterministic and names the failing record
- `californiaDirectoryData` returns `[]` from `validateDirectoryData`
- public card helpers cannot silently render broken references
- source URL checks only require absolute `http`/`https`, not live network access
- docs explain automated validation versus human editorial review

- [ ] **Step 3: Commit after approval**

Only after user approval:

```bash
git add web/src/data/schema.ts web/src/data/schema.test.ts web/src/data/index.ts docs/reference/editorial-data.md
git commit -m "test: harden directory seed data validation"
```

---

## Acceptance criteria

- `validateDirectoryData(californiaDirectoryData)` returns `[]`.
- Duplicate IDs produce collection-specific errors.
- Broken foreign-key-style references produce record-specific errors.
- Invalid source/website/campaign/photo URLs produce record-specific errors.
- Non-ISO dates produce record-specific errors.
- Card building fails from validation errors before render-time join errors.
- Editorial docs state what automated validation enforces and what still requires human review.
- `make verify` passes.
