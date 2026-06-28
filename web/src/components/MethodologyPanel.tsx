import type { ReactNode } from 'react';

export function MethodologyPanel() {
  return (
    <section
      id="methodology"
      className="rounded-3xl border border-red-800/70 bg-red-950/50 p-6 md:p-8"
    >
      <div className="max-w-4xl">
        <p className="text-sm font-semibold uppercase tracking-[0.35em] text-rose-300">Methodology & disclaimer</p>
        <h2 className="mt-3 text-3xl font-black text-white md:text-4xl">How this directory is built</h2>
        <p className="mt-4 text-base leading-7 text-rose-100/80">
          DSA Slate uses manual, source-linked records. Each public record is recorded from a public chapter, state,
          national, press, campaign, or verified social source and then reviewed by hand before it is published.
        </p>

        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <MethodologyItem title="How records are sourced">
            We publish source-linked manual records from official DSA pages, verified DSA social accounts, and other
            documented public sources. The goal is to keep the source-linked directory readable while preserving the
            original source.
          </MethodologyItem>
          <MethodologyItem title="Confidence labels">
            <span className="font-semibold text-white">official_source</span> means recorded from an official DSA
            website or page. <span className="font-semibold text-white">official_social_source</span> means recorded
            from an official DSA social account. <span className="font-semibold text-white">campaign_source</span>{' '}
            means recorded from a candidate or campaign claim and flagged for manual review.{' '}
            <span className="font-semibold text-white">press_source</span> means recorded from a press report and
            flagged for manual review. <span className="font-semibold text-white">needs_review</span> means the source
            is unclear, incomplete, or still needs human confirmation.
          </MethodologyItem>
          <MethodologyItem title="Endorsement wording">
            We do not collapse <span className="font-semibold text-white">recommended</span> into{' '}
            <span className="font-semibold text-white">endorsed</span>. The public label stays as the original source
            used it.
          </MethodologyItem>
          <MethodologyItem title="Privacy">
            No raw voter addresses are stored. The first public slice is browse and text search only; it does not do
            live address lookup.
          </MethodologyItem>
          <MethodologyItem title="Election information disclaimer">
            DSA Slate links to official election information and does not replace official voter or election sources. It
            is not official election administration or legal advice.
          </MethodologyItem>
        </div>
      </div>
    </section>
  );
}

function MethodologyItem({ children, title }: { children: ReactNode; title: string }) {
  return (
    <article className="rounded-2xl border border-red-800 bg-red-900/30 p-5">
      <h3 className="text-lg font-black text-white">{title}</h3>
      <p className="mt-2 text-sm leading-6 text-rose-100/75">{children}</p>
    </article>
  );
}
