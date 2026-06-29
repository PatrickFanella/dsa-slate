import { describe, expect, it, vi } from 'vitest';
import { californiaDirectoryData } from './california';
import {
  fetchBallotMeasures,
  fetchCandidates,
  fetchChapters,
  fetchDirectoryData,
  fetchDirectoryCollections,
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

describe('fetchDirectoryCollections', () => {
  it('assembles and returns validated directory data from collection endpoints', async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      switch (String(input)) {
        case '/api/chapters':
          return jsonResponse(californiaDirectoryData.chapters);
        case '/api/elections':
          return jsonResponse(californiaDirectoryData.elections);
        case '/api/races':
          return jsonResponse(californiaDirectoryData.races);
        case '/api/candidates':
          return jsonResponse(californiaDirectoryData.candidates);
        case '/api/voter-guides':
          return jsonResponse(californiaDirectoryData.voterGuides);
        case '/api/endorsements':
          return jsonResponse(californiaDirectoryData.endorsements);
        case '/api/ballot-measures':
          return jsonResponse(californiaDirectoryData.ballotMeasures);
        case '/api/measure-positions':
          return jsonResponse(californiaDirectoryData.measurePositions);
        default:
          return jsonResponse({ error: 'unexpected endpoint' }, 404);
      }
    });

    const data = await fetchDirectoryCollections(fetcher);

    expect(data).toEqual(californiaDirectoryData);
  });

  it('calls every collection endpoint with JSON headers', async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      switch (String(input)) {
        case '/api/chapters':
          return jsonResponse(californiaDirectoryData.chapters);
        case '/api/elections':
          return jsonResponse(californiaDirectoryData.elections);
        case '/api/races':
          return jsonResponse(californiaDirectoryData.races);
        case '/api/candidates':
          return jsonResponse(californiaDirectoryData.candidates);
        case '/api/voter-guides':
          return jsonResponse(californiaDirectoryData.voterGuides);
        case '/api/endorsements':
          return jsonResponse(californiaDirectoryData.endorsements);
        case '/api/ballot-measures':
          return jsonResponse(californiaDirectoryData.ballotMeasures);
        case '/api/measure-positions':
          return jsonResponse(californiaDirectoryData.measurePositions);
        default:
          return jsonResponse({ error: 'unexpected endpoint' }, 404);
      }
    });

    await fetchDirectoryCollections(fetcher);

    expect(fetcher).toHaveBeenNthCalledWith(1, '/api/chapters', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(2, '/api/elections', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(3, '/api/races', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(4, '/api/candidates', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(5, '/api/voter-guides', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(6, '/api/endorsements', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(7, '/api/ballot-measures', { headers: { Accept: 'application/json' } });
    expect(fetcher).toHaveBeenNthCalledWith(8, '/api/measure-positions', { headers: { Accept: 'application/json' } });
  });

  it('throws validation errors for invalid assembled collection data', async () => {
    const invalidData = {
      ...californiaDirectoryData,
      endorsements: [{ ...californiaDirectoryData.endorsements[0], sourceUrl: '' }],
    };

    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      switch (String(input)) {
        case '/api/chapters':
          return jsonResponse(invalidData.chapters);
        case '/api/elections':
          return jsonResponse(invalidData.elections);
        case '/api/races':
          return jsonResponse(invalidData.races);
        case '/api/candidates':
          return jsonResponse(invalidData.candidates);
        case '/api/voter-guides':
          return jsonResponse(invalidData.voterGuides);
        case '/api/endorsements':
          return jsonResponse(invalidData.endorsements);
        case '/api/ballot-measures':
          return jsonResponse(invalidData.ballotMeasures);
        case '/api/measure-positions':
          return jsonResponse(invalidData.measurePositions);
        default:
          return jsonResponse({ error: 'unexpected endpoint' }, 404);
      }
    });

    await expect(fetchDirectoryCollections(fetcher)).rejects.toThrow('Directory API data validation failed:\n');
  });

  it('throws a clear error when every collection is empty', async () => {
    const fetcher = vi.fn(async () => jsonResponse([]));

    await expect(fetchDirectoryCollections(fetcher)).rejects.toThrow('Directory API returned no public records');
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
