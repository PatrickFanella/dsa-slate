export type CivicLookupRequest = {
  address: string;
  electionId?: string;
};

export type CivicAddress = {
  locationName?: string;
  line1?: string;
  line2?: string;
  line3?: string;
  city?: string;
  state?: string;
  zip?: string;
};

export type CivicElection = {
  id?: string;
  name?: string;
  electionDay?: string;
};

export type CivicSource = {
  name?: string;
  official: boolean;
};

export type CivicVotingLocation = {
  address: CivicAddress;
  pollingHours?: string;
  startDate?: string;
  endDate?: string;
  sources: CivicSource[];
};

export type CivicContest = {
  id?: string;
  type?: string;
  office?: string;
  ballotTitle?: string;
  district?: string;
  candidates: string[];
};

export type CivicStateElectionOffice = {
  name?: string;
  electionInfoUrl?: string;
  registrationUrl?: string;
  registrationConfirmationUrl?: string;
  absenteeVotingInfoUrl?: string;
  ballotInfoUrl?: string;
};

export type CivicFallbackLinks = {
  registerToVoteUrl: string;
  nassRegistrationStatusUrl: string;
  nassPollingPlaceUrl: string;
};

export type CivicLookupResponse = {
  status: 'success';
  source: string;
  disclaimer: string;
  normalizedInput: CivicAddress;
  election: CivicElection;
  otherElections: CivicElection[];
  pollingLocations: CivicVotingLocation[];
  earlyVoteSites: CivicVotingLocation[];
  dropOffLocations: CivicVotingLocation[];
  contests: CivicContest[];
  stateElectionOffice: CivicStateElectionOffice;
  fallbackLinks: CivicFallbackLinks;
};

export type CivicElectionsResponse = {
  source: string;
  elections: CivicElection[];
};

export type CivicLookupErrorCode = 'civic_not_configured' | 'validation_failed' | 'civic_lookup_failed';

export class CivicLookupApiError extends Error {
  code: CivicLookupErrorCode;
  status: number;
  messages: string[];

  constructor(status: number, code: CivicLookupErrorCode, messages: string[] = []) {
    super(code);
    this.name = 'CivicLookupApiError';
    this.code = code;
    this.status = status;
    this.messages = messages;
  }
}

export const civicRegistrationDisclaimer =
  'This lookup does not confirm that you are registered to vote. Use your state election office registration-status link to confirm.';

type CivicFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

function compactText(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed === '' ? undefined : trimmed;
}

async function readResponseBody(response: Response): Promise<unknown> {
  const contentType = response.headers.get('Content-Type') ?? '';

  if (!contentType.includes('application/json')) {
    return undefined;
  }

  try {
    return await response.json();
  } catch {
    return undefined;
  }
}

function parseErrorMessages(body: unknown): string[] {
  if (body && typeof body === 'object') {
    if (Array.isArray((body as { messages?: unknown }).messages)) {
      return (body as { messages: unknown[] }).messages.filter((message): message is string => typeof message === 'string');
    }

    if (typeof (body as { message?: unknown }).message === 'string') {
      return [(body as { message: string }).message];
    }
  }

  return [];
}

function parseErrorCode(body: unknown): CivicLookupErrorCode {
  if (body && typeof body === 'object') {
    const error = (body as { error?: unknown }).error;

    if (error === 'civic_not_configured' || error === 'validation_failed' || error === 'civic_lookup_failed') {
      return error;
    }
  }

  return 'civic_lookup_failed';
}

function parseCivicErrorBody(body: unknown): { code: CivicLookupErrorCode; messages: string[] } {
  return { code: parseErrorCode(body), messages: parseErrorMessages(body) };
}

function isCivicLookupResponse(body: unknown): body is CivicLookupResponse {
  return Boolean(body && typeof body === 'object' && (body as { status?: unknown }).status === 'success');
}

function isCivicElectionsResponse(body: unknown): body is CivicElectionsResponse {
  return Boolean(body && typeof body === 'object' && Array.isArray((body as { elections?: unknown }).elections));
}

export async function lookupCivicInformation(
  request: CivicLookupRequest,
  fetcher: CivicFetch = fetch,
): Promise<CivicLookupResponse> {
  const payload: CivicLookupRequest = {
    address: request.address.trim(),
    ...(compactText(request.electionId ?? '') ? { electionId: compactText(request.electionId ?? '') } : {}),
  };

  const response = await fetcher('/api/civic/lookup', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  const body = await readResponseBody(response);

  if (!response.ok) {
    const { code, messages } = parseCivicErrorBody(body);
    throw new CivicLookupApiError(response.status, code, messages);
  }

  if (!isCivicLookupResponse(body)) {
    throw new CivicLookupApiError(502, 'civic_lookup_failed', ['Civic lookup returned an unexpected response.']);
  }

  return body;
}

export async function fetchCivicElections(fetcher: CivicFetch = fetch): Promise<CivicElectionsResponse> {
  const response = await fetcher('/api/civic/elections', {
    headers: { Accept: 'application/json' },
  });

  const body = await readResponseBody(response);

  if (!response.ok) {
    const { code, messages } = parseCivicErrorBody(body);
    throw new CivicLookupApiError(response.status, code, messages);
  }

  if (!isCivicElectionsResponse(body)) {
    throw new CivicLookupApiError(502, 'civic_lookup_failed', ['Civic elections returned an unexpected response.']);
  }

  return body;
}
