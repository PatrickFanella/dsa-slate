import { describe, expect, it, vi } from 'vitest';
import { californiaDirectoryData } from './california';
import { fetchDirectoryData } from './api';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('fetchDirectoryData', () => {
  it('returns validated directory data from the API', async () => {
    const fetcher = vi.fn(async () => jsonResponse(californiaDirectoryData));

    const data = await fetchDirectoryData(fetcher, '/api/directory');

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
