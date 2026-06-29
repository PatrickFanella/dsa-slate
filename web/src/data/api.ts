import { validateDirectoryData, type DirectoryData } from './schema';

export type DirectoryFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;
export type BrowseFilters = {
  chapterId?: string;
  electionId?: string;
  raceLevel?: DirectoryData['races'][number]['level'];
};

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

function compactFilters<T extends Record<string, QueryValue>>(filters: T): T {
  return Object.fromEntries(Object.entries(filters).filter(([, value]) => value !== undefined && value !== '')) as T;
}

function recordIdSet<T extends { id: string }>(records: T[]): Set<string> {
  return new Set(records.map((record) => record.id));
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

export async function fetchBrowseDirectoryData(
  filters: BrowseFilters = {},
  fetcher: DirectoryFetch = fetch,
): Promise<DirectoryData> {
  const raceFilters = compactFilters({ electionId: filters.electionId, level: filters.raceLevel });
  const chapterElectionFilters = compactFilters({ chapterId: filters.chapterId, electionId: filters.electionId });
  const electionFilters = compactFilters({ electionId: filters.electionId });
  const chapterFilters = compactFilters({ chapterId: filters.chapterId });

  const [chapters, elections, races, candidates, voterGuides, endorsements, ballotMeasures, measurePositions] =
    await Promise.all([
      fetchChapters(fetcher),
      fetchElections(fetcher),
      fetchRaces(raceFilters, fetcher),
      fetchCandidates(fetcher),
      fetchVoterGuides(chapterElectionFilters, fetcher),
      fetchEndorsements(chapterElectionFilters, fetcher),
      fetchBallotMeasures(electionFilters, fetcher),
      fetchMeasurePositions(chapterFilters, fetcher),
    ]);

  const raceIds = recordIdSet(races);
  const ballotMeasureIds = recordIdSet(ballotMeasures);

  const filteredEndorsements = filters.raceLevel
    ? endorsements.filter((endorsement) => raceIds.has(endorsement.raceId))
    : endorsements;
  const filteredMeasurePositions = filters.electionId
    ? measurePositions.filter((position) => ballotMeasureIds.has(position.measureId))
    : measurePositions;

  return validateDirectoryApiData(
    {
      chapters,
      elections,
      races,
      candidates,
      voterGuides,
      endorsements: filteredEndorsements,
      ballotMeasures,
      measurePositions: filteredMeasurePositions,
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
