import { SourceBadge } from './SourceBadge';
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';

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
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
    </article>
  );
}
