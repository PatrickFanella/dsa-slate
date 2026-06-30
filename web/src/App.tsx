import { AdminPanel } from './components/AdminPanel';
import { DetailPanel } from './components/DetailPanel';
import { EndorsementCard, MeasureCard, VoterGuideCard } from './components/DirectoryCard';
import { MethodologyPanel } from './components/MethodologyPanel';
import { allDirectoryCards, buildDirectoryCards, type DirectoryCardKey } from './data';
import { filterDirectoryCards } from './data/filter';
import type { BrowseFilters } from './data/api';
import { useDirectoryData, type DirectoryLoadStatus } from './data/useDirectoryData';
import type { ReactNode } from 'react';
import { useEffect, useMemo, useState } from 'react';

type ThemeMode = 'light' | 'dark';

const THEME_STORAGE_KEY = 'dsa-slate-theme';

function isAdminRoute(): boolean {
  return typeof window !== 'undefined' && window.location.pathname === '/admin';
}

function readStoredTheme(): ThemeMode | null {
  try {
    const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY);

    return storedTheme === 'dark' || storedTheme === 'light' ? storedTheme : null;
  } catch {
    return null;
  }
}

function writeStoredTheme(theme: ThemeMode) {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    // Ignore storage failures and keep the UI functional.
  }
}

function getInitialTheme(): ThemeMode {
  if (typeof window === 'undefined') {
    return 'light';
  }

  const storedTheme = readStoredTheme();

  return storedTheme ?? 'light';
}

function PublicHomePage() {
  const [query, setQuery] = useState('');
  const [chapterId, setChapterId] = useState('');
  const [electionId, setElectionId] = useState('');
  const [raceLevel, setRaceLevel] = useState('');
  const [selectedCardKey, setSelectedCardKey] = useState<DirectoryCardKey | undefined>(undefined);
  const [theme, setTheme] = useState<ThemeMode>(() => getInitialTheme());
  const browseFilters = useMemo<BrowseFilters>(
    () => ({
      chapterId: chapterId || undefined,
      electionId: electionId || undefined,
      raceLevel: (raceLevel || undefined) as BrowseFilters['raceLevel'],
    }),
    [chapterId, electionId, raceLevel],
  );
  const directoryState = useDirectoryData(fetch, browseFilters);

  const directoryCards = useMemo(() => buildDirectoryCards(directoryState.data), [directoryState.data]);
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

  const filteredCards = useMemo(() => filterDirectoryCards(directoryCards, query), [directoryCards, query]);
  const filteredCardList = useMemo(() => allDirectoryCards(filteredCards), [filteredCards]);
  const selectedCard = useMemo(
    () => filteredCardList.find((card) => card.key === selectedCardKey),
    [filteredCardList, selectedCardKey],
  );
  const endorsementCards = filteredCards.endorsements;
  const measureCards = filteredCards.measures;
  const voterGuideCards = filteredCards.voterGuides;
  const visibleResultCount = endorsementCards.length + measureCards.length + voterGuideCards.length;
  const hasResults = visibleResultCount > 0;
  const noVisibleResults = !hasResults;

  useEffect(() => {
    if (selectedCardKey && !selectedCard) {
      setSelectedCardKey(undefined);
    }
  }, [selectedCard, selectedCardKey]);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    writeStoredTheme(theme);
  }, [theme]);

  const nextTheme = theme === 'light' ? 'dark' : 'light';

  return (
    <main className="min-h-screen bg-background px-6 py-10 text-foreground">
      <section className="mx-auto flex max-w-6xl flex-col gap-10">
        <header className="rounded-none border-2 border-border-strong bg-surface p-8 shadow-[10px_10px_0_var(--theme-shadow)] md:p-12">
          <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-4xl">
              <p className="text-sm font-black uppercase tracking-[0.28em] text-primary">DSA Slate</p>
              <h1 className="mt-4 font-display text-4xl font-black tracking-tight text-foreground md:text-6xl">
                Find your DSA slate
              </h1>
              <p className="mt-5 max-w-3xl text-lg leading-8 text-muted">
                A source-linked election guide for DSA endorsements, recommendations, ballot measures, and voter guides
                from national, state, and local chapters.
              </p>
            </div>
            <button
              aria-label={`Switch to ${nextTheme} theme`}
              className="inline-flex flex-col items-start gap-1 rounded-none border-2 border-border-strong bg-surface-muted px-4 py-3 text-left shadow-[4px_4px_0_var(--theme-shadow)] transition hover:-translate-y-0.5 hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              type="button"
              onClick={() => setTheme(nextTheme)}
            >
              <span className="text-[0.65rem] font-black uppercase tracking-[0.28em] text-muted">Theme</span>
              <span className="text-sm font-black text-foreground">{theme === 'light' ? 'Light' : 'Dark'} theme</span>
            </button>
          </div>
          <div className="mt-6 flex flex-wrap gap-3">
            <a
              className="rounded-none border-2 border-primary bg-primary px-5 py-3 font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              href="#methodology"
            >
              Methodology & disclaimer
            </a>
          </div>
          <div className="mt-8 grid gap-3 rounded-none border-2 border-border bg-surface-muted p-4 md:grid-cols-[1fr_auto]">
            <label className="sr-only" htmlFor="slate-search">
              Text search candidates, chapters, offices, elections, or voter guides
            </label>
            <input
              id="slate-search"
              className="rounded-none border-2 border-border-strong bg-surface px-4 py-3 text-foreground placeholder:text-muted transition focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search this source-linked directory"
              type="search"
              value={query}
            />
            <button
              className="rounded-none border-2 border-border-strong bg-surface px-6 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              type="button"
              onClick={() => {
                setQuery('');
                setChapterId('');
                setElectionId('');
                setRaceLevel('');
                setSelectedCardKey(undefined);
              }}
            >
              Clear all
            </button>
          </div>
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
          <div className="mt-3 flex flex-col gap-2 text-sm leading-6 text-muted md:flex-row md:items-center md:justify-between">
            <p>Browse filters use public API fields. Text search stays local. Raw addresses are not stored.</p>
            <p aria-live="polite">{visibleResultCount} visible {visibleResultCount === 1 ? 'result' : 'results'}</p>
          </div>
          <DirectoryDataStatus errorMessage={directoryState.errorMessage} status={directoryState.status} />
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <StatCard label="Chapters" value={directoryState.data.chapters.length.toString()} />
          <StatCard label="Candidate records" value={endorsementCards.length.toString()} />
          <StatCard label="Voter guides" value={voterGuideCards.length.toString()} />
        </section>

        <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem] lg:items-start">
          <div className="lg:order-2 lg:sticky lg:top-6">
            <DetailPanel item={selectedCard} />
          </div>
          <div className="flex flex-col gap-10 lg:order-1">
            {noVisibleResults ? (
              <EmptyStateCard message="No records match this text search. Try a different name, chapter, office, or election." />
            ) : null}

            <DirectorySection
              count={voterGuideCards.length}
              title="Voter guides"
              description="Initial source-linked seed set for the directory MVP."
            >
              {voterGuideCards.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2">
                  {voterGuideCards.map((item) => (
                    <VoterGuideCard
                      item={item}
                      key={item.id}
                      selected={selectedCardKey === item.key}
                      onSelect={() => setSelectedCardKey(item.key)}
                    />
                  ))}
                </div>
              ) : hasResults ? (
                <EmptyStateCard message="No voter guides match this search." />
              ) : null}
            </DirectorySection>

            <DirectorySection
              count={endorsementCards.length}
              title="Candidate endorsements & recommendations"
              description="Recommendation and endorsement wording is preserved exactly as recorded from the source."
            >
              {endorsementCards.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2">
                  {endorsementCards.map((item) => (
                    <EndorsementCard
                      item={item}
                      key={item.id}
                      selected={selectedCardKey === item.key}
                      onSelect={() => setSelectedCardKey(item.key)}
                    />
                  ))}
                </div>
              ) : hasResults ? (
                <EmptyStateCard message="No candidate endorsements or recommendations match this search." />
              ) : null}
            </DirectorySection>

            <DirectorySection
              count={measureCards.length}
              title="Ballot measures"
              description="Measure positions stay linked to the original chapter source."
            >
              {measureCards.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2">
                  {measureCards.map((item) => (
                    <MeasureCard
                      item={item}
                      key={item.id}
                      selected={selectedCardKey === item.key}
                      onSelect={() => setSelectedCardKey(item.key)}
                    />
                  ))}
                </div>
              ) : hasResults ? (
                <EmptyStateCard message="No ballot measures match this search." />
              ) : null}
            </DirectorySection>
          </div>
        </section>

        <MethodologyPanel />
      </section>
    </main>
  );
}

export default function App() {
  return isAdminRoute() ? <AdminPanel /> : <PublicHomePage />;
}

function DirectorySection({
  children,
  count,
  description,
  title,
}: {
  children: ReactNode;
  count: number;
  description: string;
  title: string;
}) {
  return (
    <section className="rounded-none border-2 border-border-strong bg-surface p-6 shadow-[6px_6px_0_var(--theme-shadow)] md:p-8">
      <div className="mb-5">
        <h2 className="text-2xl font-black text-foreground">
          {title} <span className="text-muted">({count})</span>
        </h2>
        <p className="mt-2 text-sm leading-6 text-muted">{description}</p>
      </div>
      {children}
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <article className="rounded-none border-2 border-border bg-surface-muted p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <p className="text-sm font-black uppercase tracking-wide text-primary">{label}</p>
      <p className="mt-2 text-4xl font-black text-foreground">{value}</p>
    </article>
  );
}

function EmptyStateCard({ message }: { message: string }) {
  return (
    <article className="rounded-none border-2 border-dashed border-border-strong bg-surface-muted p-6 text-muted">
      <h3 className="text-lg font-black text-foreground">No matches yet</h3>
      <p className="mt-2 text-sm leading-6">{message}</p>
    </article>
  );
}

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
    <p className="mt-2 rounded-none border-2 border-border bg-surface-muted px-4 py-3 text-sm leading-6 text-muted">
      <span className="font-bold text-foreground">{copy[status].label}.</span> {copy[status].detail}
    </p>
  );
}
