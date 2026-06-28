import { describe, expect, it } from 'vitest';
import { buildDirectoryCards, directoryData } from './index';
import { filterDirectoryCards } from './filter';

describe('filterDirectoryCards', () => {
  it('returns the original card arrays for blank queries', () => {
    const cards = buildDirectoryCards(directoryData);

    const filtered = filterDirectoryCards(cards, '   ');

    expect(filtered).toBe(cards);
    expect(filtered.endorsements).toBe(cards.endorsements);
    expect(filtered.measures).toBe(cards.measures);
    expect(filtered.voterGuides).toBe(cards.voterGuides);
  });

  it('filters endorsements by candidate text', () => {
    const cards = buildDirectoryCards(directoryData);

    const filtered = filterDirectoryCards(cards, '  CORINNA   CONTRERAS  ');

    expect(filtered.endorsements).toHaveLength(1);
    expect(filtered.endorsements[0]?.candidateName).toBe('Corinna Contreras');
    expect(filtered.measures).toHaveLength(0);
    expect(filtered.voterGuides).toHaveLength(0);
  });

  it('filters measures by title text', () => {
    const cards = buildDirectoryCards(directoryData);

    const filtered = filterDirectoryCards(cards, '  san   diego   measure   a  ');

    expect(filtered.endorsements).toHaveLength(0);
    expect(filtered.measures).toHaveLength(1);
    expect(filtered.measures[0]?.title).toBe('San Diego Measure A');
    expect(filtered.voterGuides).toHaveLength(0);
  });

  it('filters voter guides by summary text', () => {
    const cards = buildDirectoryCards(directoryData);

    const filtered = filterDirectoryCards(cards, '  local   voter   guide  ');

    expect(filtered.endorsements).toHaveLength(0);
    expect(filtered.measures).toHaveLength(0);
    expect(filtered.voterGuides).toHaveLength(1);
    expect(filtered.voterGuides[0]?.summary).toContain('Local voter guide with federal, statewide, local, and ballot measure recommendations.');
  });

  it('returns empty card arrays when nothing matches', () => {
    const cards = buildDirectoryCards(directoryData);

    const filtered = filterDirectoryCards(cards, 'zebra');

    expect(filtered.endorsements).toHaveLength(0);
    expect(filtered.measures).toHaveLength(0);
    expect(filtered.voterGuides).toHaveLength(0);
  });
});
