import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import App from './App';
import { DetailPanel } from './components/DetailPanel';
import { VoterGuideCard } from './components/DirectoryCard';
import { getEndorsementCards, getMeasureCards, getVoterGuideCards } from './data';

function setPathname(pathname: string): () => void {
  const originalWindow = (globalThis as typeof globalThis & { window?: Window }).window;
  const fakeWindow = {
    location: { pathname },
    history: { pushState: (_state: unknown, _title: string, nextPathname: string) => { fakeWindow.location.pathname = nextPathname; } },
  } as typeof window;

  Object.defineProperty(globalThis, 'window', {
    configurable: true,
    value: fakeWindow,
    writable: true,
  });

  return () => {
    if (originalWindow) {
      Object.defineProperty(globalThis, 'window', {
        configurable: true,
        value: originalWindow,
        writable: true,
      });
      return;
    }

    Reflect.deleteProperty(globalThis as typeof globalThis & { window?: Window }, 'window');
  };
}

describe('App', () => {
  it('renders the admin drafting console on /admin', () => {
    const restore = setPathname('/admin');

    try {
      const html = renderToString(<App />);

      expect(html).toContain('DSA Slate Admin');
      expect(html).toContain('Editorial draft console');
      expect(html).toContain('Token-protected local drafting');
      expect(html).toContain('Admin bearer token');
      expect(html).toContain('type="password"');
      expect(html).toContain('Chapter drafts');
      expect(html).toContain('Endorsement drafts');
      expect(html).toContain('Save draft');
      expect(html).toContain('List drafts');
      expect(html).toContain('Load by ID');
      expect(html).toContain('Publish');
      expect(html).toContain('Archive');
      expect(html).toContain('Authorization: Bearer &lt;token&gt;');
      expect(html).toContain('not a public submission form');
      expect(html).not.toContain('Find your DSA slate');
    } finally {
      restore();
    }
  });

  it('renders the source-linked directory homepage', () => {
    const html = renderToString(<App />);

    expect(html).toContain('Find your DSA slate');
    expect(html).toContain('bg-background');
    expect(html).toContain('text-foreground');
    expect(html).toContain('rounded-none');
    expect(html).toContain('shadow-[10px_10px_0_var(--theme-shadow)]');
    expect(html).toContain('bg-primary');
    expect(html).toContain('Switch to dark theme');
    expect(html).toMatch(/Light.*theme/);
    expect(html).toContain('How this directory is built');
    expect(html).toContain('Voter guides');
    expect(html).toContain('Search this source-linked directory');
    expect(html).toContain('Browse filters use public API fields. Text search stays local. Raw addresses are not stored.');
    expect(html).toContain('Chapter');
    expect(html).toContain('Election');
    expect(html).toContain('Race level');
    expect(html).toContain('Clear all');
    expect(html).toContain('Record detail');
    expect(html).toContain('Select a record');
    expect(html).toContain('Choose a voter guide, endorsement, or ballot measure to inspect source-linked details.');
    expect(html).toContain('Loading live directory data');
    expect(html).toContain('Static fallback is visible while the API loads.');
    expect(html).toContain('Chapters');
    expect(html).toContain('No raw voter addresses are stored.');
    expect(html).toContain('The public label stays as the original source used it.');
    expect(html).toContain('DSA Slate links to official election information and does not replace official voter or election sources.');
    expect(html).toContain('aria-live="polite"');
    expect(html).not.toContain('disabled');
    expect(html).toContain('Corinna Contreras');
    expect(html).toContain('Recommended');
    expect(html).toContain('Official source');
    expect(html).toContain('official_social_source');
    expect(html).toContain('campaign_source');
    expect(html).toContain('press_source');
    expect(html).toContain('needs_review');
    expect(html).toContain('San Diego Measure A');
    expect(html).toContain('Original source');
    expect(html).toContain('aria-pressed="false"');
  });

  it('shows confidence on voter guide cards', () => {
    const html = renderToString(<VoterGuideCard item={getVoterGuideCards()[0]} />);

    expect(html).toContain('Official source');
    expect(html).toContain('Original source');
  });

  it('shows selected record details with source metadata', () => {
    const html = renderToString(<DetailPanel item={getEndorsementCards()[0]} />);

    expect(html).toContain('Candidate endorsement');
    expect(html).toContain('Corinna Contreras');
    expect(html).toContain('Source label');
    expect(html).toContain('Race jurisdiction');
    expect(html).toContain('CA-48');
    expect(html).toContain('Notes');
    expect(html).toContain('Seeded from DSA San Diego');
    expect(html).toContain('Original source');
  });

  it('shows editorial review metadata in selected record details', () => {
    const dsaLaEndorsement = getEndorsementCards().find(
      (card) => card.id === 'hugo-soto-martinez-dsa-la-2026-primary',
    );
    const caGuide = getVoterGuideCards().find((card) => card.id === 'dsa-la-2026-primary-guide');
    const measure = getMeasureCards()[0];

    expect(dsaLaEndorsement).toBeDefined();
    expect(caGuide).toBeDefined();

    const endorsementHtml = renderToString(<DetailPanel item={dsaLaEndorsement!} />);
    expect(endorsementHtml).toContain('Editorial review');
    expect(endorsementHtml).toContain('Reviewed on');
    expect(endorsementHtml).toContain('2026-06-28');
    expect(endorsementHtml).toContain('Manual review confirmed DSA-LA voter guide');

    const guideHtml = renderToString(<DetailPanel item={caGuide!} />);
    expect(guideHtml).toContain('No visible publication date in reviewed source.');
    expect(guideHtml).toContain('do not invent publishedAt');

    const measureHtml = renderToString(<DetailPanel item={measure!} />);
    expect(measureHtml).toContain('Manual review confirmed source URL');
  });
});
