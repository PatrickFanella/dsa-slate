import { describe, expect, it } from 'vitest';
import { californiaDirectoryData } from './california';
import { getEndorsementCards, getMeasureCards, getVoterGuideCards } from './index';
import {
  confidenceLabel,
  endorsementTypeLabel,
  validateDirectoryData,
  type DirectoryData,
} from './schema';

describe('directory schema helpers', () => {
  it('labels endorsement types without collapsing recommended into endorsed', () => {
    expect(endorsementTypeLabel('endorsed')).toBe('Endorsed');
    expect(endorsementTypeLabel('recommended')).toBe('Recommended');
  });

  it('labels confidence levels clearly', () => {
    expect(confidenceLabel('official_source')).toBe('Official source');
    expect(confidenceLabel('needs_review')).toBe('Needs review');
  });

  it('requires source URLs for endorsement and measure records', () => {
    const data: DirectoryData = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
      candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras' }],
      races: [
        {
          id: 'ca-48-us-house-2026-primary',
          electionId: '2026-ca-primary',
          office: 'U.S. House',
          district: 'CA-48',
          level: 'federal',
          jurisdiction: 'CA',
        },
      ],
      voterGuides: [],
      endorsements: [
        {
          id: 'contreras-dsa-sd-2026-primary',
          candidateId: 'corinna-contreras',
          raceId: 'ca-48-us-house-2026-primary',
          chapterId: 'dsa-san-diego',
          endorsementType: 'recommended',
          sourceUrl: '',
          confidence: 'needs_review',
        },
      ],
      ballotMeasures: [],
      measurePositions: [],
    };

    expect(validateDirectoryData(data)).toContain(
      'endorsements[contreras-dsa-sd-2026-primary] must include sourceUrl',
    );
  });

  it('requires source URLs for voter guides and measure positions', () => {
    const data: DirectoryData = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
      candidates: [],
      races: [],
      voterGuides: [
        {
          id: 'dsa-san-diego-2026-primary-guide',
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          chapterId: 'dsa-san-diego',
          electionId: '2026-ca-primary',
          sourceUrl: '',
          confidence: 'official_source',
        },
      ],
      endorsements: [],
      ballotMeasures: [{ id: 'san-diego-measure-a-2026-primary', electionId: '2026-ca-primary', title: 'San Diego Measure A', jurisdiction: 'San Diego, CA' }],
      measurePositions: [
        {
          id: 'san-diego-measure-a-dsa-sd-yes-2026-primary',
          measureId: 'san-diego-measure-a-2026-primary',
          chapterId: 'dsa-san-diego',
          position: 'yes',
          sourceUrl: '',
          confidence: 'official_source',
        },
      ],
    };

    expect(validateDirectoryData(data)).toEqual(
      expect.arrayContaining([
        'voterGuides[dsa-san-diego-2026-primary-guide] must include sourceUrl',
        'measurePositions[san-diego-measure-a-dsa-sd-yes-2026-primary] must include sourceUrl',
      ]),
    );
  });
});

describe('california seed data', () => {
  it('has no source validation errors', () => {
    expect(validateDirectoryData(californiaDirectoryData)).toEqual([]);
  });

  it('includes California cluster chapter and guide records', () => {
    expect(californiaDirectoryData.chapters.map((chapter) => chapter.id)).toContain('california-dsa');
    expect(californiaDirectoryData.chapters.map((chapter) => chapter.id)).toContain('dsa-san-diego');
    expect(californiaDirectoryData.voterGuides.map((guide) => guide.id)).toContain('california-dsa-2026-primary-guide');
  });

  it('builds joined endorsement, measure, and guide cards', () => {
    expect(getEndorsementCards()).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ candidateName: 'Corinna Contreras', endorsementType: 'recommended' }),
      ]),
    );
    expect(getMeasureCards()).toEqual(
      expect.arrayContaining([expect.objectContaining({ title: 'San Diego Measure A', position: 'yes' })]),
    );
    expect(getVoterGuideCards()).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          confidence: 'official_source',
        }),
      ]),
    );
  });
});
