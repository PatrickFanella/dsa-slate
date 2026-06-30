import type { ReactNode } from 'react';
import { SourceBadge } from './SourceBadge';
import type { EndorsementCardData, MeasureCardData, VoterGuideCardData } from '../data';

type SelectableCardProps = {
  selected?: boolean;
  onSelect?: () => void;
};

export type SaveActionProps = {
  authenticated?: boolean;
  saved?: boolean;
  saving?: boolean;
  onToggleSaved?: () => void;
};

function cardClassName(selected: boolean | undefined): string {
  return [
    'rounded-none border-2 bg-surface p-5 text-left shadow-[4px_4px_0_var(--theme-shadow)] transition',
    selected ? 'border-primary shadow-[7px_7px_0_var(--theme-shadow)]' : 'border-border hover:-translate-y-0.5 hover:border-border-strong',
  ].join(' ');
}

function SelectionButton({ children, onSelect, selected }: SelectableCardProps & { children: ReactNode }) {
  return (
    <button
      aria-pressed={selected ? 'true' : 'false'}
      className="block w-full rounded-none text-left focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      type="button"
      onClick={onSelect}
    >
      {selected ? <span className="mb-2 inline-block text-xs font-black uppercase tracking-wide text-primary">Selected</span> : null}
      {children}
    </button>
  );
}

function SaveButton({ authenticated, saved, saving, onToggleSaved }: SaveActionProps) {
  if (!onToggleSaved) {
    return null;
  }

  if (!authenticated) {
    return (
      <a
        className="mt-4 block w-full rounded-none border-2 border-border-strong bg-surface-muted px-4 py-2 text-center text-sm font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        href="/api/auth/google/start"
      >
        Sign in to save
      </a>
    );
  }

  return (
    <button
      className="mt-4 w-full rounded-none border-2 border-border-strong bg-surface-muted px-4 py-2 text-sm font-black text-foreground transition hover:bg-surface-accent disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      disabled={saving}
      type="button"
      onClick={onToggleSaved}
    >
      {saving ? (saved ? 'Removing…' : 'Saving…') : saved ? 'Remove from my slate' : 'Save to my slate'}
    </button>
  );
}

export function EndorsementCard({
  authenticated,
  item,
  onSelect,
  onToggleSaved,
  saved,
  saving,
  selected,
}: { item: EndorsementCardData } & SelectableCardProps & SaveActionProps) {
  const raceLine = [item.office, item.district, item.electionName].filter(Boolean).join(' · ');

  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-2xl font-black text-foreground">{item.candidateName}</h3>
        <p className="mt-2 text-sm leading-6 text-muted">{raceLine}</p>
      </SelectionButton>
      <SourceBadge
        confidence={item.confidence}
        endorsementType={item.endorsementType}
        sourceDate={item.sourceDate}
        sourceUrl={item.sourceUrl}
      />
      <SaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
    </article>
  );
}

export function MeasureCard({
  authenticated,
  item,
  onSelect,
  onToggleSaved,
  saved,
  saving,
  selected,
}: { item: MeasureCardData } & SelectableCardProps & SaveActionProps) {
  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-2xl font-black text-foreground">{item.title}</h3>
        <p className="mt-2 text-sm leading-6 text-muted">
          {item.jurisdiction} · {item.electionName} · Position: {item.position.toUpperCase()}
        </p>
      </SelectionButton>
      <SourceBadge confidence={item.confidence} sourceUrl={item.sourceUrl} />
      <SaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
    </article>
  );
}

export function VoterGuideCard({
  authenticated,
  item,
  onSelect,
  onToggleSaved,
  saved,
  saving,
  selected,
}: { item: VoterGuideCardData } & SelectableCardProps & SaveActionProps) {
  return (
    <article className={cardClassName(selected)} data-selected={selected ? 'true' : 'false'}>
      <SelectionButton onSelect={onSelect} selected={selected}>
        <p className="text-sm font-black text-primary">{item.chapterName}</p>
        <h3 className="mt-2 text-xl font-black text-foreground">{item.title}</h3>
        {item.electionName ? <p className="mt-2 text-sm leading-6 text-muted">{item.electionName}</p> : null}
        {item.summary ? <p className="mt-3 text-sm leading-6 text-muted">{item.summary}</p> : null}
      </SelectionButton>
      <SourceBadge confidence={item.confidence} sourceDate={item.publishedAt} sourceUrl={item.sourceUrl} />
      <SaveButton authenticated={authenticated} onToggleSaved={onToggleSaved} saved={saved} saving={saving} />
    </article>
  );
}
