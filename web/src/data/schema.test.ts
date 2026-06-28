import { describe, expect, it } from 'vitest';
import { californiaDirectoryData } from './california';
import { buildDirectoryCards, getEndorsementCards, getMeasureCards, getVoterGuideCards } from './index';
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

  it('reports duplicate IDs by collection', () => {
    const data: DirectoryData = {
      chapters: [
        { id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' },
        { id: 'dsa-san-diego', name: 'Duplicate DSA San Diego', state: 'CA' },
      ],
      elections: [],
      races: [],
      candidates: [],
      voterGuides: [],
      endorsements: [],
      ballotMeasures: [],
      measurePositions: [],
    };

    expect(validateDirectoryData(data)).toContain('chapters[dsa-san-diego] duplicate id');
  });

  it('reports invalid URLs and non-ISO dates', () => {
    const data = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA', websiteUrl: 'not-a-url' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary', electionDate: '06/02/2026' }],
      races: [],
      candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras', campaignUrl: 'corinna.example' }],
      voterGuides: [
        {
          id: 'dsa-san-diego-2026-primary-guide',
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          sourceUrl: 'dsasandiego.org/guide',
          confidence: 'official_source',
          publishedAt: 'May 5, 2026',
        },
      ],
      endorsements: [],
      ballotMeasures: [],
      measurePositions: [],
    } as unknown as DirectoryData;

    expect(validateDirectoryData(data)).toEqual(
      expect.arrayContaining([
        'chapters[dsa-san-diego].websiteUrl must be an absolute http(s) URL',
        'elections[2026-ca-primary].electionDate must be YYYY-MM-DD',
        'candidates[corinna-contreras].campaignUrl must be an absolute http(s) URL',
        'voterGuides[dsa-san-diego-2026-primary-guide].sourceUrl must be an absolute http(s) URL',
        'voterGuides[dsa-san-diego-2026-primary-guide].publishedAt must be YYYY-MM-DD',
      ]),
    );
  });

  it('rejects impossible calendar dates in YYYY-MM-DD format', () => {
    const data: DirectoryData = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary', electionDate: '2026-02-31' }],
      races: [],
      candidates: [],
      voterGuides: [
        {
          id: 'dsa-san-diego-2026-primary-guide',
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          chapterId: 'dsa-san-diego',
          sourceUrl: 'https://dsasandiego.org/guide',
          confidence: 'official_source',
          publishedAt: '2026-04-31',
        },
      ],
      endorsements: [],
      ballotMeasures: [],
      measurePositions: [],
    };

    expect(validateDirectoryData(data)).toEqual(
      expect.arrayContaining([
        'elections[2026-ca-primary].electionDate must be YYYY-MM-DD',
        'voterGuides[dsa-san-diego-2026-primary-guide].publishedAt must be YYYY-MM-DD',
      ]),
    );
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

  it('reports broken record references before cards are built', () => {
    const data: DirectoryData = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
      races: [
        {
          id: 'ca-48-us-house-2026-primary',
          electionId: 'missing-election',
          office: 'U.S. House',
          level: 'federal',
          jurisdiction: 'CA-48',
        },
      ],
      candidates: [{ id: 'corinna-contreras', name: 'Corinna Contreras' }],
      voterGuides: [
        {
          id: 'dsa-san-diego-2026-primary-guide',
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          chapterId: 'missing-chapter',
          electionId: 'missing-election',
          sourceUrl: 'https://dsasandiego.org/guide',
          confidence: 'official_source',
        },
      ],
      endorsements: [
        {
          id: 'contreras-dsa-sd-2026-primary',
          candidateId: 'missing-candidate',
          raceId: 'missing-race',
          chapterId: 'missing-chapter',
          voterGuideId: 'missing-guide',
          endorsementType: 'recommended',
          sourceUrl: 'https://dsasandiego.org/guide',
          confidence: 'official_source',
        },
      ],
      ballotMeasures: [
        {
          id: 'san-diego-measure-a-2026-primary',
          electionId: 'missing-election',
          title: 'San Diego Measure A',
          jurisdiction: 'San Diego, CA',
        },
      ],
      measurePositions: [
        {
          id: 'san-diego-measure-a-dsa-sd-yes-2026-primary',
          measureId: 'missing-measure',
          chapterId: 'missing-chapter',
          position: 'yes',
          sourceUrl: 'https://dsasandiego.org/guide',
          confidence: 'official_source',
        },
      ],
    };

    expect(validateDirectoryData(data)).toEqual(
      expect.arrayContaining([
        'races[ca-48-us-house-2026-primary].electionId references missing elections[missing-election]',
        'voterGuides[dsa-san-diego-2026-primary-guide].chapterId references missing chapters[missing-chapter]',
        'endorsements[contreras-dsa-sd-2026-primary].candidateId references missing candidates[missing-candidate]',
        'endorsements[contreras-dsa-sd-2026-primary].raceId references missing races[missing-race]',
        'endorsements[contreras-dsa-sd-2026-primary].chapterId references missing chapters[missing-chapter]',
        'endorsements[contreras-dsa-sd-2026-primary].voterGuideId references missing voterGuides[missing-guide]',
        'ballotMeasures[san-diego-measure-a-2026-primary].electionId references missing elections[missing-election]',
        'measurePositions[san-diego-measure-a-dsa-sd-yes-2026-primary].measureId references missing ballotMeasures[missing-measure]',
        'measurePositions[san-diego-measure-a-dsa-sd-yes-2026-primary].chapterId references missing chapters[missing-chapter]',
      ]),
    );
  });

  it('requires voter guides to include a chapterId before validating chapter references', () => {
    const data = {
      chapters: [{ id: 'dsa-san-diego', name: 'DSA San Diego', state: 'CA' }],
      elections: [{ id: '2026-ca-primary', name: '2026 California Primary' }],
      races: [],
      candidates: [],
      voterGuides: [
        {
          id: 'dsa-san-diego-2026-primary-guide',
          title: "DSA San Diego's June 2026 Primary Voter Guide",
          sourceUrl: 'https://dsasandiego.org/guide',
          confidence: 'official_source',
        },
      ],
      endorsements: [],
      ballotMeasures: [],
      measurePositions: [],
    } as unknown as DirectoryData;

    expect(validateDirectoryData(data)).toContain('voterGuides[dsa-san-diego-2026-primary-guide] must include chapterId');
  });

  it('fails card building with validation errors instead of render-time joins', () => {
    const data: DirectoryData = {
      chapters: [],
      elections: [],
      races: [],
      candidates: [],
      voterGuides: [],
      endorsements: [
        {
          id: 'broken-endorsement',
          candidateId: 'missing-candidate',
          raceId: 'missing-race',
          chapterId: 'missing-chapter',
          endorsementType: 'recommended',
          sourceUrl: 'https://example.org/source',
          confidence: 'needs_review',
        },
      ],
      ballotMeasures: [],
      measurePositions: [],
    };

    expect(() => buildDirectoryCards(data)).toThrow(
      'Directory data validation failed:\nendorsements[broken-endorsement].candidateId references missing candidates[missing-candidate]',
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
