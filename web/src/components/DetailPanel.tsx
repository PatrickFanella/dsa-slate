import { SourceBadge } from './SourceBadge';
import type { SaveActionProps } from './DirectoryCard';
import { endorsementTypeLabel } from '../data/schema';
import type { DirectoryCardData, EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';

type DetailItem = DirectoryCardData;

type DetailPanelProps = {
  item?: DetailItem;
} & SaveActionProps;

function DetailRow({ label, value }: { label: string; value?: string }) {
  if (!value) {
    return null;
  }

  return (
    <div className="border-t-2 border-border py-3">
      <dt className="text-xs font-black uppercase tracking-wide text-muted">{label}</dt>
      <dd className="mt-1 text-sm font-bold leading-6 text-foreground">{value}</dd>
    </div>
  );
}

function DetailLink({ href, label }: { href?: string; label: string }) {
  if (!href) {
    return null;
  }

  return (
    <a
      className="text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      href={href}
    >
      {label}
    </a>
  );
}

function ReviewMetadata({ reviewedAt, reviewNotes }: { reviewedAt?: string; reviewNotes?: string }) {
  if (!reviewedAt && !reviewNotes) {
    return null;
  }

  return (
    <section className="mt-5 border-2 border-border bg-surface-muted p-4">
      <p className="text-xs font-black uppercase tracking-wide text-primary">Editorial review</p>
      <dl className="mt-2">
        <DetailRow label="Reviewed on" value={reviewedAt} />
        <DetailRow label="Review notes" value={reviewNotes} />
      </dl>
    </section>
  );
}

function DetailSaveButton({ authenticated, onToggleSaved, saved, saving }: SaveActionProps) {
  if (!onToggleSaved) {
    return null;
  }

  if (!authenticated) {
    return (
      <a
        className="mt-5 block w-full rounded-none border-2 border-primary bg-primary px-4 py-3 text-center font-black text-primary-foreground transition hover:bg-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        href="/api/auth/google/start"
      >
        Sign in to save this record
      </a>
    );
  }

  return (
    <button
      className="mt-5 w-full rounded-none border-2 border-border-strong bg-primary px-4 py-3 font-black text-primary-foreground transition hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      disabled={saving}
      type="button"
      onClick={onToggleSaved}
    >
      {saving ? (saved ? 'Removing…' : 'Saving…') : saved ? 'Remove from my slate' : 'Save to my slate'}
    </button>
  );
}

export function DetailPanel({ authenticated, item, onToggleSaved, saved, saving }: DetailPanelProps) {
  if (!item) {
    return (
      <aside
        aria-label="Record detail"
        aria-live="polite"
        className="rounded-none border-2 border-dashed border-border-strong bg-surface-muted p-6 text-muted shadow-[6px_6px_0_var(--theme-shadow)]"
      >
        <p className="text-xs font-black uppercase tracking-wide text-primary">Record detail</p>
        <h2 className="mt-3 text-2xl font-black text-foreground">Select a record</h2>
        <p className="mt-3 text-sm leading-6">
          Choose a voter guide, endorsement, or ballot measure to inspect source-linked details.
        </p>
      </aside>
    );
  }

  const title = item.kind === 'endorsement' ? item.candidateName : item.title;
  const kicker = item.kind === 'endorsement' ? 'Candidate endorsement' : item.kind === 'measure' ? 'Ballot measure' : 'Voter guide';

  return (
    <aside
      aria-label="Record detail"
      aria-live="polite"
      className="rounded-none border-2 border-border-strong bg-surface p-6 shadow-[8px_8px_0_var(--theme-shadow)]"
    >
      <p className="text-xs font-black uppercase tracking-wide text-primary">{kicker}</p>
      <h2 className="mt-3 text-3xl font-black text-foreground">{title}</h2>
      <p className="mt-3 text-sm leading-6 text-muted">Recorded from source-linked public DSA electoral material.</p>

      {item.kind === 'endorsement' ? (
        <EndorsementDetail
          authenticated={authenticated}
          item={item}
          onToggleSaved={onToggleSaved}
          saved={saved}
          saving={saving}
        />
      ) : null}
      {item.kind === 'measure' ? (
        <MeasureDetail authenticated={authenticated} item={item} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
      ) : null}
      {item.kind === 'voterGuide' ? (
        <VoterGuideDetail authenticated={authenticated} item={item} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
      ) : null}
    </aside>
  );
}

function EndorsementDetail({
  authenticated,
  item,
  onToggleSaved,
  saved,
  saving,
}: { item: EndorsementCardData } & SaveActionProps) {
  return (
    <>
      <SourceBadge
        confidence={item.confidence}
        endorsementType={item.endorsementType}
        sourceDate={item.sourceDate}
        sourceUrl={item.sourceUrl}
      />
      <dl className="mt-5">
        <DetailRow label="Source label" value={endorsementTypeLabel(item.endorsementType)} />
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Office" value={[item.office, item.district].filter(Boolean).join(' · ')} />
        <DetailRow label="Race level" value={item.raceLevel.split('_').join(' ')} />
        <DetailRow label="Race jurisdiction" value={item.raceJurisdiction} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Candidate party" value={item.candidateParty} />
        <DetailRow label="Notes" value={item.notes} />
      </dl>
      <ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
      <DetailSaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
      <div className="mt-4 flex flex-wrap gap-4 text-sm font-bold">
        <DetailLink href={item.candidateCampaignUrl} label="Candidate site" />
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}

function MeasureDetail({ authenticated, item, onToggleSaved, saved, saving }: { item: MeasureCardData } & SaveActionProps) {
  return (
    <>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
      <dl className="mt-5">
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Position" value={item.position.toUpperCase()} />
        <DetailRow label="Measure number" value={item.measureNumber} />
        <DetailRow label="Jurisdiction" value={item.jurisdiction} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Summary" value={item.summary} />
      </dl>
      <ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
      <DetailSaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
      <div className="mt-4 text-sm font-bold">
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}

function VoterGuideDetail({ authenticated, item, onToggleSaved, saved, saving }: { item: VoterGuideCardData } & SaveActionProps) {
  return (
    <>
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
      <dl className="mt-5">
        <DetailRow label="Chapter" value={item.chapterName} />
        <DetailRow label="Election" value={[item.electionName, item.electionDate].filter(Boolean).join(' · ')} />
        <DetailRow label="Published" value={item.publishedAt} />
        <DetailRow
          label="Publication date"
          value={!item.publishedAt && item.reviewedAt ? 'No visible publication date in reviewed source.' : undefined}
        />
        <DetailRow label="Summary" value={item.summary} />
      </dl>
      <ReviewMetadata reviewedAt={item.reviewedAt} reviewNotes={item.reviewNotes} />
      <DetailSaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
      <div className="mt-4 text-sm font-bold">
        <DetailLink href={item.chapterWebsiteUrl} label="Chapter site" />
      </div>
    </>
  );
}
