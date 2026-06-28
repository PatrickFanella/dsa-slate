# Trusted Directory MVP Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Build a static, source-linked DSA Slate trusted directory MVP using California-cluster seed data.

**Architecture:** Keep the first slice simple: checked-in TypeScript seed data drives static React pages/components, while the SQL migration remains the future database contract. No admin UI, address lookup, or external scraping in this slice. The frontend validates and renders source metadata, endorsement type, and confidence labels so sourcing rules are visible from day one.

**Tech Stack:** Go backend placeholder, Vite React TypeScript frontend, Tailwind CSS, Vitest, PostgreSQL migration SQL, Docker Compose, Makefile verification.

---

## Interview decisions

- First launch goal: trusted source-linked directory.
- Initial seed cluster: California DSA, DSA San Diego, DSA-LA, and nearby California examples.
- Editorial workflow: manual seed files and review conventions; no admin UI yet.
- First slice done criteria: static browsable MVP rendered from source-linked seed data.
- Planning route: full plan + review gate.
- Persistence mode: hybrid; this repo plan is the reviewable artifact.

## Complexity assessment

| Dimension | Rating | Reason |
| --- | --- | --- |
| Logic depth | Medium | Mostly rendering/filtering and validation, but source-confidence semantics matter. |
| Contract sensitivity | High | Endorsement types, confidence labels, and source fields become public data contracts. |
| Context span | Medium | Several coordinated files across docs, seed data, UI, tests, and schema. |
| Discovery need | Medium | Seed entries require source review, but this plan uses a bounded California cluster. |
| Failure cost | High | Mislabeling endorsements or implying official status would undermine trust. |
| Concern coupling | Medium | Data, editorial policy, public UI, and privacy guardrails must stay aligned. |

## File structure

- Create `web/src/data/schema.ts` — TypeScript types, enums, labels, and validation helpers for directory records.
- Create `web/src/data/california.ts` — manually reviewed California-cluster seed data.
- Create `web/src/data/index.ts` — normalized exports and lightweight query helpers for pages/components.
- Create `web/src/components/DirectoryCard.tsx` — reusable cards for candidates, measures, and voter guides.
- Create `web/src/components/SourceBadge.tsx` — visible source/confidence/endorsement metadata.
- Modify `web/src/App.tsx` — render static MVP homepage, directory sections, and seed-driven cards.
- Modify `web/src/App.test.tsx` — assert seed-driven public contract: source links, labels, and California cluster content render.
- Create `web/src/data/schema.test.ts` — unit tests for validation helpers and controlled vocabulary.
- Modify `docs/product.md` — add a “Trusted Directory MVP Plan Decisions” section summarizing approved scope.
- Create `docs/reference/editorial-data.md` — manual seed/source review rules for the first launch.
- Modify `migrations/0001_init.sql` only if implementation reveals a mismatch between SQL constraints and TypeScript contract.

## Controlled vocabulary

Use these exact public values in seed data and UI.

```ts
export const endorsementTypes = [
  'endorsed',
  'recommended',
  'opposed',
  'no_recommendation',
  'write_in',
  'coalition',
  'unknown',
] as const;

export const confidenceLevels = [
  'official_source',
  'official_social_source',
  'campaign_source',
  'press_source',
  'needs_review',
] as const;
```

---

## Task 1: Add directory data contract and validation

**Files:**
- Create: `web/src/data/schema.ts`
- Create: `web/src/data/schema.test.ts`

- [ ] **Step 1: Write failing tests for labels and validation**

Create `web/src/data/schema.test.ts`:

```ts
import { describe, expect, it } from 'vitest';
import {
  confidenceLabel,
  endorsementTypeLabel,
  validateDirectoryData,
  type DirectoryData,
} from './schema';

describe('directory schema helpers', () => {
  it('labels endorsement types without collapsing recommended into endorsed', () => {
    expect(endorsementTypeLabel('endorsed')).toBe('Endorsed');
    expect(endorsementTypeLabel('recommended')).toBe('Recommended');
  });

  it('labels confidence levels clearly', () => {
    expect(confidenceLabel('official_source')).toBe('Official source');
    expect(confidenceLabel('needs_review')).toBe('Needs review');
  });

  it('requires source URLs for endorsement and measure records', () => {
    const data: DirectoryData = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
      candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras' }],
      races: [
        {
          id: 'ca-48-us-house-2026-primary',
          electionId: '2026-ca-primary',
          office: 'U.S. House',
          district: 'CA-48',
          level: 'federal',
          jurisdiction: 'CA',
        },
      ],
      voterGuides: [],
      endorsements: [
        {
          id: 'contreras-dsa-sd-2026-primary',
          candidateId: 'corinna-contreras',
          raceId: 'ca-48-us-house-2026-primary',
          chapterId: 'dsa-san-diego',
          endorsementType: 'recommended',
          sourceUrl: '',
          confidence: 'needs_review',
        },
      ],
      ballotMeasures: [],
      measurePositions: [],
    };

    expect(validateDirectoryData(data)).toContain(
      'endorsements[contreras-dsa-sd-2026-primary] must include sourceUrl',
    );
  });
});
```

- [ ] **Step 2: Run the focused failing test**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: fail because `web/src/data/schema.ts` does not exist.

- [ ] **Step 3: Implement the schema and validation helpers**

Create `web/src/data/schema.ts`:

```ts
export const endorsementTypes = [
  'endorsed',
  'recommended',
  'opposed',
  'no_recommendation',
  'write_in',
  'coalition',
  'unknown',
] as const;

export const confidenceLevels = [
  'official_source',
  'official_social_source',
  'campaign_source',
  'press_source',
  'needs_review',
] as const;

export type EndorsementType = (typeof endorsementTypes)[number];
export type ConfidenceLevel = (typeof confidenceLevels)[number];

export type Chapter = {
  id: string;
  name: string;
  state: string;
  websiteUrl?: string;
  officialDsaUrl?: string;
  coverageNotes?: string;
  active?: boolean;
};

export type Election = {
  id: string;
  name: string;
  electionDate?: string;
  electionType?: string;
  jurisdiction?: string;
  officialUrl?: string;
};

export type Race = {
  id: string;
  electionId: string;
  office: string;
  district?: string;
  level: 'federal' | 'statewide' | 'state_legislature' | 'local' | 'unknown';
  jurisdiction: string;
};

export type Candidate = {
  id: string;
  name: string;
  campaignUrl?: string;
  party?: string;
  incumbent?: boolean;
  photoUrl?: string;
};

export type VoterGuide = {
  id: string;
  title: string;
  chapterId?: string;
  electionId?: string;
  sourceUrl: string;
  publishedAt?: string;
  summary?: string;
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
};

export type BallotMeasure = {
  id: string;
  electionId: string;
  title: string;
  jurisdiction: string;
  measureNumber?: string;
  summary?: string;
};

export type MeasurePosition = {
  id: string;
  measureId: string;
  chapterId: string;
  position: 'yes' | 'no' | 'no_recommendation' | 'opposed' | 'unknown';
  sourceUrl: string;
  confidence: ConfidenceLevel;
};

export type DirectoryData = {
  chapters: Chapter[];
  elections: Election[];
  races: Race[];
  candidates: Candidate[];
  voterGuides: VoterGuide[];
  endorsements: Endorsement[];
  ballotMeasures: BallotMeasure[];
  measurePositions: MeasurePosition[];
};

export function endorsementTypeLabel(type: EndorsementType): string {
  const labels: Record<EndorsementType, string> = {
    endorsed: 'Endorsed',
    recommended: 'Recommended',
    opposed: 'Opposed',
    no_recommendation: 'No recommendation',
    write_in: 'Write-in',
    coalition: 'Coalition',
    unknown: 'Unknown',
  };

  return labels[type];
}

export function confidenceLabel(level: ConfidenceLevel): string {
  const labels: Record<ConfidenceLevel, string> = {
    official_source: 'Official source',
    official_social_source: 'Official social source',
    campaign_source: 'Campaign source',
    press_source: 'Press source',
    needs_review: 'Needs review',
  };

  return labels[level];
}

export function validateDirectoryData(data: DirectoryData): string[] {
  const errors: string[] = [];

  for (const guide of data.voterGuides) {
    if (!guide.sourceUrl.trim()) {
      errors.push(`voterGuides[${guide.id}] must include sourceUrl`);
    }
  }

  for (const endorsement of data.endorsements) {
    if (!endorsement.sourceUrl.trim()) {
      errors.push(`endorsements[${endorsement.id}] must include sourceUrl`);
    }
  }

  for (const position of data.measurePositions) {
    if (!position.sourceUrl.trim()) {
      errors.push(`measurePositions[${position.id}] must include sourceUrl`);
    }
  }

  return errors;
}
```

- [ ] **Step 4: Run focused test and typecheck**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
pnpm --dir web run format
```

Expected: both pass.

- [ ] **Step 5: Commit**

```bash
git add web/src/data/schema.ts web/src/data/schema.test.ts
git commit -m "feat: add directory data contract"
```

---

## Task 2: Add California-cluster seed data and query helpers

**Files:**
- Create: `web/src/data/california.ts`
- Create: `web/src/data/index.ts`
- Modify: `web/src/data/schema.test.ts`

- [ ] **Step 1: Add tests for valid seed data and helper joins**

Append to `web/src/data/schema.test.ts`:

```ts
import { californiaDirectoryData } from './california';
import { getEndorsementCards, getMeasureCards, getVoterGuideCards } from './index';

describe('california seed data', () => {
  it('has no source validation errors', () => {
    expect(validateDirectoryData(californiaDirectoryData)).toEqual([]);
  });

  it('includes California cluster chapter and guide records', () => {
    expect(californiaDirectoryData.chapters.map((chapter) => chapter.id)).toContain('california-dsa');
    expect(californiaDirectoryData.chapters.map((chapter) => chapter.id)).toContain('dsa-san-diego');
    expect(californiaDirectoryData.voterGuides.map((guide) => guide.id)).toContain('california-dsa-2026-primary-guide');
  });

  it('builds joined endorsement, measure, and guide cards', () => {
    expect(getEndorsementCards()).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ candidateName: 'Corinna Contreras', endorsementType: 'recommended' }),
      ]),
    );
    expect(getMeasureCards()).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ title: 'San Diego Measure A', position: 'yes' }),
      ]),
    );
    expect(getVoterGuideCards()).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ title: "DSA San Diego's June 2026 Primary Voter Guide" }),
      ]),
    );
  });
});
```

- [ ] **Step 2: Run the focused failing test**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
```

Expected: fail because `california.ts` and `index.ts` do not exist.

- [ ] **Step 3: Create California seed data**

Create `web/src/data/california.ts`:

```ts
import type { DirectoryData } from './schema';

export const californiaDirectoryData: DirectoryData = {
  chapters: [
    {
      id: 'california-dsa',
      name: 'California DSA',
      state: 'CA',
      websiteUrl: 'https://www.californiadsa.org/',
      coverageNotes: 'Statewide DSA voter guide and links to local California chapter guides.',
      active: true,
    },
    {
      id: 'dsa-san-diego',
      name: 'DSA San Diego',
      state: 'CA',
      websiteUrl: 'https://dsasandiego.org/',
      coverageNotes: 'San Diego County and local voter guide coverage.',
      active: true,
    },
    {
      id: 'dsa-los-angeles',
      name: 'DSA-LA',
      state: 'CA',
      websiteUrl: 'https://dsa-la.org/',
      coverageNotes: 'Los Angeles area chapter included as a California-cluster seed source.',
      active: true,
    },
  ],
  elections: [
    {
      id: '2026-ca-primary',
      name: '2026 California Primary',
      electionDate: '2026-06-02',
      electionType: 'primary',
      jurisdiction: 'CA',
    },
  ],
  races: [
    {
      id: 'ca-48-us-house-2026-primary',
      electionId: '2026-ca-primary',
      office: 'U.S. House',
      district: 'CA-48',
      level: 'federal',
      jurisdiction: 'CA-48',
    },
    {
      id: 'ca-50-us-house-2026-primary',
      electionId: '2026-ca-primary',
      office: 'U.S. House',
      district: 'CA-50',
      level: 'federal',
      jurisdiction: 'CA-50',
    },
  ],
  candidates: [
    {
      id: 'corinna-contreras',
      name: 'Corinna Contreras',
      party: 'Democrat',
    },
    {
      id: 'aishwarya-sparky-mitra',
      name: 'Aishwarya “Sparky” Mitra',
      party: 'Democrat',
    },
  ],
  voterGuides: [
    {
      id: 'california-dsa-2026-primary-guide',
      title: 'California DSA 2026 Primary Voter Guide',
      chapterId: 'california-dsa',
      electionId: '2026-ca-primary',
      sourceUrl: 'https://www.californiadsa.org/voterguide',
      summary: 'Statewide guide and local chapter guide hub for California DSA electoral recommendations.',
    },
    {
      id: 'dsa-san-diego-2026-primary-guide',
      title: "DSA San Diego's June 2026 Primary Voter Guide",
      chapterId: 'dsa-san-diego',
      electionId: '2026-ca-primary',
      sourceUrl: 'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
      publishedAt: '2026-05-05',
      summary: 'Local voter guide with federal, statewide, local, and ballot measure recommendations.',
    },
  ],
  endorsements: [
    {
      id: 'contreras-dsa-sd-2026-primary',
      candidateId: 'corinna-contreras',
      raceId: 'ca-48-us-house-2026-primary',
      chapterId: 'dsa-san-diego',
      voterGuideId: 'dsa-san-diego-2026-primary-guide',
      endorsementType: 'recommended',
      sourceUrl: 'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
      sourceDate: '2026-05-05',
      confidence: 'official_source',
      notes: 'Seeded from DSA San Diego voter guide wording; preserve recommendation language.',
    },
    {
      id: 'mitra-dsa-sd-2026-primary',
      candidateId: 'aishwarya-sparky-mitra',
      raceId: 'ca-50-us-house-2026-primary',
      chapterId: 'dsa-san-diego',
      voterGuideId: 'dsa-san-diego-2026-primary-guide',
      endorsementType: 'recommended',
      sourceUrl: 'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
      sourceDate: '2026-05-05',
      confidence: 'official_source',
      notes: 'Seeded from DSA San Diego voter guide wording; preserve recommendation language.',
    },
  ],
  ballotMeasures: [
    {
      id: 'san-diego-measure-a-2026-primary',
      electionId: '2026-ca-primary',
      title: 'San Diego Measure A',
      jurisdiction: 'San Diego, CA',
      measureNumber: 'A',
      summary: 'Local ballot measure included in DSA San Diego voter guide seed data.',
    },
  ],
  measurePositions: [
    {
      id: 'san-diego-measure-a-dsa-sd-yes-2026-primary',
      measureId: 'san-diego-measure-a-2026-primary',
      chapterId: 'dsa-san-diego',
      position: 'yes',
      sourceUrl: 'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
      confidence: 'official_source',
    },
  ],
};
```

- [ ] **Step 4: Create query helpers**

Create `web/src/data/index.ts`:

```ts
import { californiaDirectoryData } from './california';
import type { ConfidenceLevel, EndorsementType } from './schema';

export const directoryData = californiaDirectoryData;

export type EndorsementCardData = {
  id: string;
  candidateName: string;
  office: string;
  district?: string;
  electionName: string;
  chapterName: string;
  endorsementType: EndorsementType;
  confidence: ConfidenceLevel;
  sourceUrl: string;
  sourceDate?: string;
};

export type MeasureCardData = {
  id: string;
  title: string;
  jurisdiction: string;
  electionName: string;
  chapterName: string;
  position: string;
  confidence: ConfidenceLevel;
  sourceUrl: string;
};

export type VoterGuideCardData = {
  id: string;
  title: string;
  chapterName: string;
  electionName?: string;
  sourceUrl: string;
  publishedAt?: string;
  summary?: string;
};

export function getEndorsementCards(): EndorsementCardData[] {
  return directoryData.endorsements.map((endorsement) => {
    const candidate = directoryData.candidates.find((item) => item.id === endorsement.candidateId);
    const race = directoryData.races.find((item) => item.id === endorsement.raceId);
    const election = directoryData.elections.find((item) => item.id === race?.electionId);
    const chapter = directoryData.chapters.find((item) => item.id === endorsement.chapterId);

    if (!candidate || !race || !election || !chapter) {
      throw new Error(`Unable to build endorsement card ${endorsement.id}`);
    }

    return {
      id: endorsement.id,
      candidateName: candidate.name,
      office: race.office,
      district: race.district,
      electionName: election.name,
      chapterName: chapter.name,
      endorsementType: endorsement.endorsementType,
      confidence: endorsement.confidence,
      sourceUrl: endorsement.sourceUrl,
      sourceDate: endorsement.sourceDate,
    };
  });
}

export function getMeasureCards(): MeasureCardData[] {
  return directoryData.measurePositions.map((position) => {
    const measure = directoryData.ballotMeasures.find((item) => item.id === position.measureId);
    const election = directoryData.elections.find((item) => item.id === measure?.electionId);
    const chapter = directoryData.chapters.find((item) => item.id === position.chapterId);

    if (!measure || !election || !chapter) {
      throw new Error(`Unable to build measure card ${position.id}`);
    }

    return {
      id: position.id,
      title: measure.title,
      jurisdiction: measure.jurisdiction,
      electionName: election.name,
      chapterName: chapter.name,
      position: position.position,
      confidence: position.confidence,
      sourceUrl: position.sourceUrl,
    };
  });
}

export function getVoterGuideCards(): VoterGuideCardData[] {
  return directoryData.voterGuides.map((guide) => {
    const chapter = directoryData.chapters.find((item) => item.id === guide.chapterId);
    const election = directoryData.elections.find((item) => item.id === guide.electionId);

    if (!chapter) {
      throw new Error(`Unable to build voter guide card ${guide.id}`);
    }

    return {
      id: guide.id,
      title: guide.title,
      chapterName: chapter.name,
      electionName: election?.name,
      sourceUrl: guide.sourceUrl,
      publishedAt: guide.publishedAt,
      summary: guide.summary,
    };
  });
}
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
pnpm --dir web run test -- src/data/schema.test.ts
pnpm --dir web run format
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add web/src/data/california.ts web/src/data/index.ts web/src/data/schema.test.ts
git commit -m "feat: seed california directory data"
```

---

## Task 3: Add source-aware directory components

**Files:**
- Create: `web/src/components/SourceBadge.tsx`
- Create: `web/src/components/DirectoryCard.tsx`
- Modify: `web/src/App.test.tsx`

- [ ] **Step 1: Expand app test to require source-aware UI text**

Replace `web/src/App.test.tsx` with:

```tsx
import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import App from './App';

describe('App', () => {
  it('renders the trusted directory homepage', () => {
    const html = renderToString(<App />);

    expect(html).toContain('Find your DSA slate');
    expect(html).toContain('California cluster');
    expect(html).toContain('Corinna Contreras');
    expect(html).toContain('Recommended');
    expect(html).toContain('Official source');
    expect(html).toContain('San Diego Measure A');
    expect(html).toContain('Original source');
  });
});
```

- [ ] **Step 2: Run failing app test**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: fail because the app does not render seed-driven cards yet.

- [ ] **Step 3: Add source badge component**

Create `web/src/components/SourceBadge.tsx`:

```tsx
import { confidenceLabel, endorsementTypeLabel, type ConfidenceLevel, type EndorsementType } from '../data/schema';

type SourceBadgeProps = {
  confidence: ConfidenceLevel;
  endorsementType?: EndorsementType;
  sourceUrl: string;
  sourceDate?: string;
};

export function SourceBadge({ confidence, endorsementType, sourceUrl, sourceDate }: SourceBadgeProps) {
  return (
    <div className="mt-4 flex flex-wrap items-center gap-2 text-xs font-semibold uppercase tracking-wide">
      {endorsementType ? (
        <span className="rounded-full bg-rose-100 px-3 py-1 text-red-950">
          {endorsementTypeLabel(endorsementType)}
        </span>
      ) : null}
      <span className="rounded-full border border-red-600 px-3 py-1 text-rose-100">
        {confidenceLabel(confidence)}
      </span>
      {sourceDate ? <span className="text-rose-200/70">Source date: {sourceDate}</span> : null}
      <a className="text-rose-100 underline decoration-rose-300 underline-offset-4" href={sourceUrl}>
        Original source
      </a>
    </div>
  );
}
```

- [ ] **Step 4: Add directory card components**

Create `web/src/components/DirectoryCard.tsx`:

```tsx
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';
import { SourceBadge } from './SourceBadge';

export function EndorsementCard({ item }: { item: EndorsementCardData }) {
  const raceLine = [item.office, item.district, item.electionName].filter(Boolean).join(' · ');

  return (
    <article className="rounded-2xl border border-red-800 bg-red-900/30 p-5">
      <p className="text-sm font-semibold text-rose-200">{item.chapterName}</p>
      <h3 className="mt-2 text-2xl font-black text-white">{item.candidateName}</h3>
      <p className="mt-2 text-sm text-rose-100/80">{raceLine}</p>
      <SourceBadge
        confidence={item.confidence}
        endorsementType={item.endorsementType}
        sourceDate={item.sourceDate}
        sourceUrl={item.sourceUrl}
      />
    </article>
  );
}

export function MeasureCard({ item }: { item: MeasureCardData }) {
  return (
    <article className="rounded-2xl border border-red-800 bg-red-900/30 p-5">
      <p className="text-sm font-semibold text-rose-200">{item.chapterName}</p>
      <h3 className="mt-2 text-2xl font-black text-white">{item.title}</h3>
      <p className="mt-2 text-sm text-rose-100/80">
        {item.jurisdiction} · {item.electionName} · Position: {item.position.toUpperCase()}
      </p>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
    </article>
  );
}

export function VoterGuideCard({ item }: { item: VoterGuideCardData }) {
  return (
    <article className="rounded-2xl border border-red-800 bg-red-900/30 p-5">
      <p className="text-sm font-semibold text-rose-200">{item.chapterName}</p>
      <h3 className="mt-2 text-xl font-black text-white">{item.title}</h3>
      {item.electionName ? <p className="mt-2 text-sm text-rose-100/80">{item.electionName}</p> : null}
      {item.summary ? <p className="mt-3 text-sm leading-6 text-rose-100/70">{item.summary}</p> : null}
      <div className="mt-4 text-xs font-semibold uppercase tracking-wide">
        {item.publishedAt ? <span className="mr-3 text-rose-200/70">Published: {item.publishedAt}</span> : null}
        <a className="text-rose-100 underline decoration-rose-300 underline-offset-4" href={item.sourceUrl}>
          Original source
        </a>
      </div>
    </article>
  );
}
```

- [ ] **Step 5: Commit component groundwork after tests still fail**

Run:

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: still fail until `App.tsx` uses these components.

Commit:

```bash
git add web/src/components/SourceBadge.tsx web/src/components/DirectoryCard.tsx web/src/App.test.tsx
git commit -m "feat: add source-aware directory cards"
```

---

## Task 4: Render static trusted-directory MVP homepage

**Files:**
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Replace homepage with seed-driven directory view**

Replace `web/src/App.tsx` with:

```tsx
import { EndorsementCard, MeasureCard, VoterGuideCard } from './components/DirectoryCard';
import { directoryData, getEndorsementCards, getMeasureCards, getVoterGuideCards } from './data';

const endorsementCards = getEndorsementCards();
const measureCards = getMeasureCards();
const voterGuideCards = getVoterGuideCards();

export default function App() {
  return (
    <main className="min-h-screen bg-red-950 px-6 py-10 text-rose-50">
      <section className="mx-auto flex max-w-6xl flex-col gap-10">
        <header className="rounded-3xl border border-red-800/70 bg-red-950/60 p-8 shadow-2xl shadow-red-950/50 md:p-12">
          <p className="text-sm font-semibold uppercase tracking-[0.35em] text-rose-300">DSA Slate</p>
          <h1 className="mt-4 text-4xl font-black tracking-tight text-white md:text-6xl">Find your DSA slate</h1>
          <p className="mt-5 max-w-3xl text-lg leading-8 text-rose-100/85">
            A source-linked election guide for DSA endorsements, recommendations, ballot measures, and voter guides
            from national, state, and local chapters.
          </p>
          <form className="mt-8 grid gap-3 rounded-2xl border border-red-800 bg-red-900/40 p-4 md:grid-cols-[1fr_auto]">
            <label className="sr-only" htmlFor="slate-search">
              Search candidates, chapters, offices, elections, or voter guides
            </label>
            <input
              id="slate-search"
              className="rounded-xl border border-red-700 bg-red-950 px-4 py-3 text-rose-50 placeholder:text-rose-200/60"
              placeholder="Search candidates, chapters, offices, elections..."
              type="search"
            />
            <button className="rounded-xl bg-rose-100 px-6 py-3 font-bold text-red-950 transition hover:bg-white" type="submit">
              Search
            </button>
          </form>
          <p className="mt-4 text-sm leading-6 text-rose-100/70">
            MVP note: this slice is the California cluster trusted directory. Address lookup comes later and must not
            store raw voter addresses.
          </p>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <StatCard label="Seed chapters" value={directoryData.chapters.length.toString()} />
          <StatCard label="Candidate records" value={endorsementCards.length.toString()} />
          <StatCard label="Voter guides" value={voterGuideCards.length.toString()} />
        </section>

        <DirectorySection title="California cluster" description="Initial source-linked seed set for the trusted directory MVP.">
          <div className="grid gap-4 md:grid-cols-2">
            {voterGuideCards.map((item) => (
              <VoterGuideCard item={item} key={item.id} />
            ))}
          </div>
        </DirectorySection>

        <DirectorySection title="Candidate recommendations" description="Recommendation and endorsement wording is preserved exactly in the public label.">
          <div className="grid gap-4 md:grid-cols-2">
            {endorsementCards.map((item) => (
              <EndorsementCard item={item} key={item.id} />
            ))}
          </div>
        </DirectorySection>

        <DirectorySection title="Ballot measures" description="Measure positions stay linked to the original chapter source.">
          <div className="grid gap-4 md:grid-cols-2">
            {measureCards.map((item) => (
              <MeasureCard item={item} key={item.id} />
            ))}
          </div>
        </DirectorySection>
      </section>
    </main>
  );
}

function DirectorySection({ children, description, title }: { children: React.ReactNode; description: string; title: string }) {
  return (
    <section className="rounded-3xl border border-red-800/70 bg-red-950/50 p-6 md:p-8">
      <div className="mb-5">
        <h2 className="text-2xl font-black text-white">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-rose-100/70">{description}</p>
      </div>
      {children}
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <article className="rounded-2xl border border-red-800 bg-red-900/30 p-5">
      <p className="text-sm font-semibold uppercase tracking-wide text-rose-200/80">{label}</p>
      <p className="mt-2 text-4xl font-black text-white">{value}</p>
    </article>
  );
}
```

- [ ] **Step 2: Run frontend tests and typecheck**

Run:

```bash
pnpm --dir web run test
pnpm --dir web run format
pnpm --dir web run lint
```

Expected: pass.

- [ ] **Step 3: Commit homepage rendering**

```bash
git add web/src/App.tsx
git commit -m "feat: render trusted directory homepage"
```

---

## Task 5: Document editorial seed rules and update PRD decisions

**Files:**
- Create: `docs/reference/editorial-data.md`
- Modify: `docs/product.md`

- [ ] **Step 1: Create editorial data reference**

Create `docs/reference/editorial-data.md`:

```markdown
# Editorial Data Reference

DSA Slate starts with manually reviewed, source-linked seed data. The first trusted-directory MVP uses the California cluster and checked-in seed files under `web/src/data/`.

## Required fields for public records

Every public endorsement, recommendation, voter guide, and measure position must include:

- original source URL
- endorsing body or chapter
- exact endorsement/position type
- confidence label
- election or jurisdiction context when known

## Endorsement wording rule

Do not collapse `recommended` into `endorsed`. If a chapter guide says “recommend,” publish `recommended`. If a chapter guide says “endorse,” publish `endorsed`.

## Confidence labels

- `official_source` — DSA national, state, or chapter website.
- `official_social_source` — verified DSA body social account.
- `campaign_source` — candidate claims endorsement; needs confirmation.
- `press_source` — news report; needs official confirmation where possible.
- `needs_review` — unclear, incomplete, or unverified source.

## First-slice privacy rule

The trusted-directory MVP does not store voter addresses. Search is static text search or browsing only. Future ZIP/address lookup requires a separate privacy review before implementation.

## Seed review checklist

- Confirm the source page is public.
- Confirm the source is an official DSA national, state, or chapter source when using `official_source`.
- Preserve exact endorsement wording.
- Add source date when visible.
- Add notes for ambiguity rather than upgrading confidence.
- Run `make verify` before merging seed changes.
```

- [ ] **Step 2: Add plan decisions to product PRD**

Append to `docs/product.md`:

```markdown

## Trusted Directory MVP plan decisions

- First launch proves a trusted source-linked directory, not location lookup.
- Initial seed data focuses on the California cluster.
- Editorial workflow is manual seed files plus documented review rules.
- First implementation slice renders static public cards from seed data.
- Address lookup, admin UI, CMS integration, public submission forms, and results tracking are later phases.
```

- [ ] **Step 3: Commit docs**

```bash
git add docs/reference/editorial-data.md docs/product.md
git commit -m "docs: define trusted directory editorial rules"
```

---

## Task 6: Full verification and review gate

**Files:**
- Review all changed files from Tasks 1–5.

- [ ] **Step 1: Run full verification**

Run:

```bash
make verify
```

Expected: pass. Note that pnpm may warn about ignored `esbuild` build scripts; that warning is acceptable if all checks pass.

- [ ] **Step 2: Inspect remaining template or placeholder strings**

Run:

```bash
rg "project-template|SUBCULT template|template heading|TODO|TBD|fill in" .
```

Expected: no matches in project-authored files except acceptable dependency lockfile strings containing package names.

- [ ] **Step 3: Review source-label behavior manually**

Run:

```bash
pnpm --dir web run build
```

Expected: build succeeds and the rendered app includes:

- “Find your DSA slate”
- “California cluster”
- “Corinna Contreras”
- “Recommended”
- “Official source”
- “San Diego Measure A”
- “Original source” links

- [ ] **Step 4: Stop for human review**

Do not continue to admin UI, location lookup, route splitting, scraping, or DB-backed APIs until the owner reviews this slice and approves the next scope.

---

## Out of scope for this plan

- Address or ZIP lookup.
- Raw voter address storage.
- Admin UI or CMS integration.
- Public submission/correction form.
- Automated scraping or source refresh jobs.
- Live election results.
- Campaign coordination, paid communications, or legal/compliance automation.

## Acceptance criteria

- The repository contains a typed directory data contract for endorsements, confidence labels, chapters, elections, races, candidates, voter guides, measures, and positions.
- California-cluster seed data validates with no missing source URLs.
- The homepage renders seed-driven voter guide, candidate, and measure cards.
- Public UI distinguishes `recommended` from `endorsed`.
- Public UI shows original source links and confidence labels.
- Editorial seed rules are documented.
- `make verify` passes.

## Self-review notes

- Spec coverage: the plan covers the approved trusted-directory MVP, California seed cluster, manual seed workflow, static browsable cards, source links, confidence labels, and privacy guardrail.
- Placeholder scan: no intentionally unresolved placeholders are used in implementation steps.
- Type consistency: `EndorsementType`, `ConfidenceLevel`, seed data, card helpers, and component props use the same names across tasks.
