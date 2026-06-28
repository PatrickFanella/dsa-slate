import { EndorsementCard, MeasureCard, VoterGuideCard } from './components/DirectoryCard';
import { buildDirectoryCards, directoryData } from './data';
import { filterDirectoryCards } from './data/filter';
import type { ReactNode } from 'react';
import { useMemo, useState } from 'react';

const directoryCards = buildDirectoryCards(directoryData);

export default function App() {
  const [query, setQuery] = useState('');

  const filteredCards = useMemo(() => filterDirectoryCards(directoryCards, query), [query]);
  const endorsementCards = filteredCards.endorsements;
  const measureCards = filteredCards.measures;
  const voterGuideCards = filteredCards.voterGuides;
  const visibleResultCount = endorsementCards.length + measureCards.length + voterGuideCards.length;
  const hasResults = visibleResultCount > 0;
  const noVisibleResults = !hasResults;

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
          <div className="mt-8 grid gap-3 rounded-2xl border border-red-800 bg-red-900/40 p-4 md:grid-cols-[1fr_auto]">
            <label className="sr-only" htmlFor="slate-search">
              Text search candidates, chapters, offices, elections, or voter guides
            </label>
            <input
              id="slate-search"
              className="rounded-xl border border-red-700 bg-red-950 px-4 py-3 text-rose-50 placeholder:text-rose-200/60 transition focus-visible:border-rose-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-200 focus-visible:ring-offset-2 focus-visible:ring-offset-red-950"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search the trusted directory"
              type="search"
              value={query}
            />
            <button
              className="rounded-xl bg-rose-100 px-6 py-3 font-bold text-red-950 transition hover:bg-rose-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-200 focus-visible:ring-offset-2 focus-visible:ring-offset-red-950"
              type="button"
              onClick={() => setQuery('')}
            >
              Clear
            </button>
          </div>
          <div className="mt-3 flex flex-col gap-2 text-sm leading-6 text-rose-100/70 md:flex-row md:items-center md:justify-between">
            <p>Text search only — not address lookup. Raw addresses are not stored.</p>
            <p aria-live="polite">{visibleResultCount} visible {visibleResultCount === 1 ? 'result' : 'results'}</p>
          </div>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <StatCard label="Seed chapters" value={directoryData.chapters.length.toString()} />
          <StatCard label="Candidate records" value={endorsementCards.length.toString()} />
          <StatCard label="Voter guides" value={voterGuideCards.length.toString()} />
        </section>

        {noVisibleResults ? (
          <EmptyStateCard message="No records match this text search. Try a different name, chapter, office, or election." />
        ) : null}

        <DirectorySection
          count={voterGuideCards.length}
          title="Voter guides"
          description="Initial source-linked seed set for the trusted directory MVP."
        >
          {voterGuideCards.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2">
              {voterGuideCards.map((item) => (
                <VoterGuideCard item={item} key={item.id} />
              ))}
            </div>
          ) : hasResults ? (
            <EmptyStateCard message="No voter guides match this search." />
          ) : null}
        </DirectorySection>

        <DirectorySection
          count={endorsementCards.length}
          title="Candidate recommendations"
          description="Recommendation and endorsement wording is preserved exactly in the public label."
        >
          {endorsementCards.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2">
              {endorsementCards.map((item) => (
                <EndorsementCard item={item} key={item.id} />
              ))}
            </div>
          ) : hasResults ? (
            <EmptyStateCard message="No candidate recommendations match this search." />
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
                <MeasureCard item={item} key={item.id} />
              ))}
            </div>
          ) : hasResults ? (
            <EmptyStateCard message="No ballot measures match this search." />
          ) : null}
        </DirectorySection>
      </section>
    </main>
  );
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
    <section className="rounded-3xl border border-red-800/70 bg-red-950/50 p-6 md:p-8">
      <div className="mb-5">
        <h2 className="text-2xl font-black text-white">
          {title} <span className="text-rose-200/70">({count})</span>
        </h2>
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

function EmptyStateCard({ message }: { message: string }) {
  return (
    <article className="rounded-2xl border border-dashed border-red-700 bg-red-950/40 p-6 text-rose-100/80">
      <h3 className="text-lg font-black text-white">No matches yet</h3>
      <p className="mt-2 text-sm leading-6">{message}</p>
    </article>
  );
}
