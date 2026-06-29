import { SourceBadge } from './SourceBadge';
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';

export function EndorsementCard({ item }: { item: EndorsementCardData }) {
  const raceLine = [item.office, item.district, item.electionName].filter(Boolean).join(' · ');

  return (
    <article className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <p className="text-sm font-black text-primary">{item.chapterName}</p>
      <h3 className="mt-2 text-2xl font-black text-foreground">{item.candidateName}</h3>
      <p className="mt-2 text-sm leading-6 text-muted">{raceLine}</p>
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
    <article className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <p className="text-sm font-black text-primary">{item.chapterName}</p>
      <h3 className="mt-2 text-2xl font-black text-foreground">{item.title}</h3>
      <p className="mt-2 text-sm leading-6 text-muted">
        {item.jurisdiction} · {item.electionName} · Position: {item.position.toUpperCase()}
      </p>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
    </article>
  );
}

export function VoterGuideCard({ item }: { item: VoterGuideCardData }) {
  return (
    <article className="rounded-none border-2 border-border bg-surface p-5 shadow-[4px_4px_0_var(--theme-shadow)]">
      <p className="text-sm font-black text-primary">{item.chapterName}</p>
      <h3 className="mt-2 text-xl font-black text-foreground">{item.title}</h3>
      {item.electionName ? <p className="mt-2 text-sm leading-6 text-muted">{item.electionName}</p> : null}
      {item.summary ? <p className="mt-3 text-sm leading-6 text-muted">{item.summary}</p> : null}
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
    </article>
  );
}
