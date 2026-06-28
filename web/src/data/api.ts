import { validateDirectoryData, type DirectoryData } from './schema';

export type DirectoryFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

export async function fetchDirectoryData(
  fetcher: DirectoryFetch = fetch,
  endpoint = '/api/directory',
): Promise<DirectoryData> {
  const response = await fetcher(endpoint, { headers: { Accept: 'application/json' } });

  if (!response.ok) {
    throw new Error(`Directory API request failed with status ${response.status}`);
  }

  const data = (await response.json()) as DirectoryData;
  const errors = validateDirectoryData(data);

  if (errors.length > 0) {
    throw new Error(`Directory API data validation failed:\n${errors.join('\n')}`);
  }

  const isCompletelyEmpty =
    data.chapters.length === 0 &&
    data.elections.length === 0 &&
    data.races.length === 0 &&
    data.candidates.length === 0 &&
    data.voterGuides.length === 0 &&
    data.endorsements.length === 0 &&
    data.ballotMeasures.length === 0 &&
    data.measurePositions.length === 0;

  if (isCompletelyEmpty) {
    throw new Error('Directory API returned no public records');
  }

  return data;
}
