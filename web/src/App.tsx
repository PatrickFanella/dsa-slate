import { EndorsementCard, MeasureCard, VoterGuideCard } from './components/DirectoryCard';
import { buildDirectoryCards, directoryData } from './data';
import type { ReactNode } from 'react';

const directoryCards = buildDirectoryCards(directoryData);
const endorsementCards = directoryCards.endorsements;
const measureCards = directoryCards.measures;
const voterGuideCards = directoryCards.voterGuides;

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
          <div className="mt-8 grid gap-3 rounded-2xl border border-red-800 bg-red-900/40 p-4 md:grid-cols-[1fr_auto]">
            <label className="sr-only" htmlFor="slate-search">
              Search candidates, chapters, offices, elections, or voter guides
            </label>
            <input
              id="slate-search"
              className="cursor-not-allowed rounded-xl border border-red-700 bg-red-950 px-4 py-3 text-rose-50 opacity-70 placeholder:text-rose-200/60"
              disabled
              placeholder="Browse-only MVP: search comes later"
              type="search"
            />
            <button
              className="cursor-not-allowed rounded-xl bg-rose-100 px-6 py-3 font-bold text-red-950 opacity-70 transition"
              disabled
              type="button"
            >
              Coming later
            </button>
          </div>
          <p className="mt-3 text-sm leading-6 text-rose-100/70">
            Browse-only MVP: search is disabled until the directory workflow is ready.
          </p>
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

function DirectorySection({ children, description, title }: { children: ReactNode; description: string; title: string }) {
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
