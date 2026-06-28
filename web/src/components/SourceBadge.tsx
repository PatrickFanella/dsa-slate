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
