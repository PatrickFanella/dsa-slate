import {
  confidenceLabel,
  endorsementTypeLabel,
  type ConfidenceLevel,
  type EndorsementType,
} from '../data/schema';

type SourceBadgeProps = {
  confidence: ConfidenceLevel;
  endorsementType?: EndorsementType;
  sourceUrl: string;
  sourceDate?: string;
};

export function SourceBadge({ confidence, endorsementType, sourceUrl, sourceDate }: SourceBadgeProps) {
  return (
    <div className="mt-4 flex flex-wrap items-center gap-2 text-xs font-black uppercase tracking-wide">
      {endorsementType ? (
        <span className="rounded-none border-2 border-primary bg-primary px-3 py-1 text-primary-foreground shadow-[2px_2px_0_var(--theme-shadow)]">
          {endorsementTypeLabel(endorsementType)}
        </span>
      ) : null}
      <span className="rounded-none border-2 border-border-strong bg-surface px-3 py-1 text-foreground shadow-[2px_2px_0_var(--theme-shadow)]">
        {confidenceLabel(confidence)}
      </span>
      {sourceDate ? <span className="text-muted">Source date: {sourceDate}</span> : null}
      <a
        className="text-link underline decoration-2 underline-offset-4 hover:text-primary-hover focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        href={sourceUrl}
      >
        Original source
      </a>
    </div>
  );
}
