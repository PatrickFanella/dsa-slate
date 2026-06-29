import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import App from './App';
import { VoterGuideCard } from './components/DirectoryCard';
import { getVoterGuideCards } from './data';

describe('App', () => {
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
    expect(html).toContain('Text search only — not address lookup. Raw addresses are not stored.');
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
  });

  it('shows confidence on voter guide cards', () => {
    const html = renderToString(<VoterGuideCard item={getVoterGuideCards()[0]} />);

    expect(html).toContain('Official source');
    expect(html).toContain('Original source');
  });
});
