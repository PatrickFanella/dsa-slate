import { validateDirectoryData, type DirectoryData } from './schema';

export type DirectoryFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

type QueryValue = string | undefined;

function buildQueryString(params: Record<string, QueryValue>): string {
  const searchParams = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') {
      continue;
    }

    searchParams.set(key, value);
  }

  const query = searchParams.toString();
  return query ? `?${query}` : '';
}

async function fetchCollectionData<T>(
  fetcher: DirectoryFetch,
  endpoint: string,
): Promise<T> {
  const response = await fetcher(endpoint, { headers: { Accept: 'application/json' } });

  if (!response.ok) {
    throw new Error(`Collection API request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

function isDirectoryDataEmpty(data: DirectoryData): boolean {
  return (
    data.chapters.length === 0 &&
    data.elections.length === 0 &&
    data.races.length === 0 &&
    data.candidates.length === 0 &&
    data.voterGuides.length === 0 &&
    data.endorsements.length === 0 &&
    data.ballotMeasures.length === 0 &&
    data.measurePositions.length === 0
  );
}

function validateDirectoryApiData(data: DirectoryData, sourceLabel: string): DirectoryData {
  const errors = validateDirectoryData(data);

  if (errors.length > 0) {
    throw new Error(`${sourceLabel} data validation failed:\n${errors.join('\n')}`);
  }

  if (isDirectoryDataEmpty(data)) {
    throw new Error('Directory API returned no public records');
  }

  return data;
}

export async function fetchDirectoryData(
  fetcher: DirectoryFetch = fetch,
  endpoint = '/api/directory',
): Promise<DirectoryData> {
  const response = await fetcher(endpoint, { headers: { Accept: 'application/json' } });

  if (!response.ok) {
    throw new Error(`Directory API request failed with status ${response.status}`);
  }

  const data = (await response.json()) as DirectoryData;
  return validateDirectoryApiData(data, 'Directory API');
}

export async function fetchDirectoryCollections(fetcher: DirectoryFetch = fetch): Promise<DirectoryData> {
  const [chapters, elections, races, candidates, voterGuides, endorsements, ballotMeasures, measurePositions] =
    await Promise.all([
      fetchChapters(fetcher),
      fetchElections(fetcher),
      fetchRaces({}, fetcher),
      fetchCandidates(fetcher),
      fetchVoterGuides({}, fetcher),
      fetchEndorsements({}, fetcher),
      fetchBallotMeasures({}, fetcher),
      fetchMeasurePositions({}, fetcher),
    ]);

  return validateDirectoryApiData(
    {
      chapters,
      elections,
      races,
      candidates,
      voterGuides,
      endorsements,
      ballotMeasures,
      measurePositions,
    },
    'Directory API',
  );
}

export async function fetchChapters(fetcher: DirectoryFetch = fetch): Promise<DirectoryData['chapters']> {
  return fetchCollectionData<DirectoryData['chapters']>(fetcher, '/api/chapters');
}

export async function fetchElections(fetcher: DirectoryFetch = fetch): Promise<DirectoryData['elections']> {
  return fetchCollectionData<DirectoryData['elections']>(fetcher, '/api/elections');
}

export async function fetchRaces(
  filters: { electionId?: string; jurisdiction?: string; level?: string } = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData['races']> {
  return fetchCollectionData<DirectoryData['races']>(
    fetcher,
    `/api/races${buildQueryString({
      electionId: filters.electionId,
      jurisdiction: filters.jurisdiction,
      level: filters.level,
    })}`,
  );
}

export async function fetchCandidates(fetcher: DirectoryFetch = fetch): Promise<DirectoryData['candidates']> {
  return fetchCollectionData<DirectoryData['candidates']>(fetcher, '/api/candidates');
}

export async function fetchVoterGuides(
  filters: { chapterId?: string; electionId?: string } = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData['voterGuides']> {
  return fetchCollectionData<DirectoryData['voterGuides']>(
    fetcher,
    `/api/voter-guides${buildQueryString({
      chapterId: filters.chapterId,
      electionId: filters.electionId,
    })}`,
  );
}

export async function fetchEndorsements(
  filters: { chapterId?: string; electionId?: string; raceId?: string } = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData['endorsements']> {
  return fetchCollectionData<DirectoryData['endorsements']>(
    fetcher,
    `/api/endorsements${buildQueryString({
      chapterId: filters.chapterId,
      electionId: filters.electionId,
      raceId: filters.raceId,
    })}`,
  );
}

export async function fetchBallotMeasures(
  filters: { electionId?: string } = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData['ballotMeasures']> {
  return fetchCollectionData<DirectoryData['ballotMeasures']>(
    fetcher,
    `/api/ballot-measures${buildQueryString({
      electionId: filters.electionId,
    })}`,
  );
}

export async function fetchMeasurePositions(
  filters: { chapterId?: string; measureId?: string } = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData['measurePositions']> {
  return fetchCollectionData<DirectoryData['measurePositions']>(
    fetcher,
    `/api/measure-positions${buildQueryString({
      chapterId: filters.chapterId,
      measureId: filters.measureId,
    })}`,
  );
}
