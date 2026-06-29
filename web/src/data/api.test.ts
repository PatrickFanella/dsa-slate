import { describe, expect, it, vi } from 'vitest';
import { californiaDirectoryData } from './california';
import {
  fetchBallotMeasures,
  fetchCandidates,
  fetchChapters,
  fetchDirectoryData,
  fetchElections,
  fetchEndorsements,
  fetchMeasurePositions,
  fetchRaces,
  fetchVoterGuides,
} from './api';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('fetchDirectoryData', () => {
  it('returns validated directory data from the API', async () => {
    const fetcher = vi.fn(async () => jsonResponse(californiaDirectoryData));

    const data = await fetchDirectoryData(fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/directory', { headers: { Accept: 'application/json' } });
    expect(data.chapters.map((chapter) => chapter.id)).toContain('dsa-san-diego');
  });

  it('throws a clear error for non-2xx responses', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ error: 'internal_server_error' }, 500));

    await expect(fetchDirectoryData(fetcher, '/api/directory')).rejects.toThrow(
      'Directory API request failed with status 500',
    );
  });

  it('throws validation errors for invalid directory data', async () => {
    const invalidData = {
      ...californiaDirectoryData,
      endorsements: [{ ...californiaDirectoryData.endorsements[0], sourceUrl: '' }],
    };
    const fetcher = vi.fn(async () => jsonResponse(invalidData));

    await expect(fetchDirectoryData(fetcher, '/api/directory')).rejects.toThrow(
      'Directory API data validation failed:\n',
    );
  });

  it('throws a clear error when the API returns no public records', async () => {
    const emptyData = {
      chapters: [] as const,
      elections: [] as const,
      races: [] as const,
      candidates: [] as const,
      voterGuides: [] as const,
      endorsements: [] as const,
      ballotMeasures: [] as const,
      measurePositions: [] as const,
    };
    const fetcher = vi.fn(async () => jsonResponse(emptyData));

    await expect(fetchDirectoryData(fetcher, '/api/directory')).rejects.toThrow(
      'Directory API returned no public records',
    );
  });
});

describe('collection fetch helpers', () => {
  it('call the expected collection endpoints', async () => {
    const fetcher = vi.fn(async () => jsonResponse([]));

    await fetchChapters(fetcher);
    await fetchElections(fetcher);
    await fetchCandidates(fetcher);
    await fetchBallotMeasures({ electionId: 'election-1' }, fetcher);
    await fetchMeasurePositions({ chapterId: 'chapter-1', measureId: 'measure-1' }, fetcher);
    await fetchVoterGuides({ chapterId: 'chapter-1', electionId: 'election-1' }, fetcher);
    await fetchEndorsements({ chapterId: 'chapter-1', electionId: 'election-1', raceId: 'race-1' }, fetcher);

    expect(fetcher).toHaveBeenNthCalledWith(1, '/api/chapters', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(2, '/api/elections', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(3, '/api/candidates', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(4, '/api/ballot-measures?electionId=election-1', {
      headers: { Accept: 'application/json' },
    });
    expect(fetcher).toHaveBeenNthCalledWith(5, '/api/measure-positions?chapterId=chapter-1&measureId=measure-1', {
      headers: { Accept: 'application/json' },
    });
    expect(fetcher).toHaveBeenNthCalledWith(6, '/api/voter-guides?chapterId=chapter-1&electionId=election-1', {
      headers: { Accept: 'application/json' },
    });
    expect(fetcher).toHaveBeenNthCalledWith(7, '/api/endorsements?chapterId=chapter-1&electionId=election-1&raceId=race-1', {
      headers: { Accept: 'application/json' },
    });
  });

  it('encodes query params and omits empty values', async () => {
    const fetcher = vi.fn(async () => jsonResponse([]));

    await fetchRaces({ electionId: '2026 ca primary', jurisdiction: '', level: 'statewide & local' }, fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/races?electionId=2026+ca+primary&level=statewide+%26+local', {
      headers: { Accept: 'application/json' },
    });
  });

  it('throws a clear error for non-2xx collection responses', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ error: 'nope' }, 503));

    await expect(fetchRaces({ electionId: 'election-1' }, fetcher)).rejects.toThrow(
      'Collection API request failed with status 503',
    );
  });
});
