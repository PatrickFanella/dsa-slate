import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import App from './App';
import { VoterGuideCard } from './components/DirectoryCard';
import { getVoterGuideCards } from './data';

describe('App', () => {
  it('renders the trusted directory homepage', () => {
    const html = renderToString(<App />);

    expect(html).toContain('Find your DSA slate');
    expect(html).toContain('Voter guides');
    expect(html).toContain('Search the trusted directory');
    expect(html).toContain('Text search only — not address lookup. Raw addresses are not stored.');
    expect(html).toContain('aria-live="polite"');
    expect(html).not.toContain('disabled');
    expect(html).toContain('Corinna Contreras');
    expect(html).toContain('Recommended');
    expect(html).toContain('Official source');
    expect(html).toContain('San Diego Measure A');
    expect(html).toContain('Original source');
  });

  it('shows confidence on voter guide cards', () => {
    const html = renderToString(<VoterGuideCard item={getVoterGuideCards()[0]} />);

    expect(html).toContain('Official source');
    expect(html).toContain('Original source');
  });
});
