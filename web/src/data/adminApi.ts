import { confidenceLevels, endorsementTypes } from './schema';

export type AdminDraftStatus = 'draft' | 'needs_review' | 'published' | 'archived';

export const adminDraftStatuses: AdminDraftStatus[] = ['draft', 'needs_review', 'published', 'archived'];

export type AdminChapterDraft = {
  id: string;
  chapterId: string;
  status: AdminDraftStatus;
  name: string;
  state: string;
  websiteUrl?: string;
  groupEmail?: string;
  reviewNotes?: string;
  reviewedAt?: string;
};

export type AdminChapterDraftForm = {
  id: string;
  chapterId: string;
  status: AdminDraftStatus;
  name: string;
  state: string;
  websiteUrl: string;
  groupEmail: string;
  reviewNotes: string;
  reviewedAt: string;
};

export type AdminEndorsementDraft = {
  id: string;
  endorsementId: string;
  status: AdminDraftStatus;
  candidateId: string;
  raceId: string;
  chapterId: string;
  voterGuideId?: string;
  endorsementType: (typeof endorsementTypes)[number];
  sourceUrl: string;
  sourceDate?: string;
  confidence: (typeof confidenceLevels)[number];
  notes?: string;
  reviewNotes?: string;
  reviewedAt?: string;
};

export type AdminEndorsementDraftForm = {
  id: string;
  endorsementId: string;
  status: AdminDraftStatus;
  candidateId: string;
  raceId: string;
  chapterId: string;
  voterGuideId: string;
  endorsementType: (typeof endorsementTypes)[number];
  sourceUrl: string;
  sourceDate: string;
  confidence: (typeof confidenceLevels)[number];
  notes: string;
  reviewNotes: string;
  reviewedAt: string;
};

export class AdminApiError extends Error {
  status: number;
  error: string;
  messages: string[];

  constructor(status: number, error: string, messages: string[] = []) {
    super(error);
    this.name = 'AdminApiError';
    this.status = status;
    this.error = error;
    this.messages = messages;
  }
}

type AdminFetch = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

function blankToUndefined(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed === '' ? undefined : trimmed;
}

function adminHeaders(hasBody: boolean): HeadersInit {
  return {
    Accept: 'application/json',
    ...(hasBody ? { 'Content-Type': 'application/json' } : {}),
  };
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

function parseAdminErrorBody(body: unknown): { error: string; messages: string[] } {
  if (body && typeof body === 'object') {
    const error = typeof (body as { error?: unknown }).error === 'string' ? (body as { error: string }).error : 'request_failed';
    const messages = Array.isArray((body as { messages?: unknown }).messages)
      ? (body as { messages: unknown[] }).messages.filter((message): message is string => typeof message === 'string')
      : typeof (body as { message?: unknown }).message === 'string'
        ? [(body as { message: string }).message]
        : [];

    return { error, messages };
  }

  return { error: 'request_failed', messages: [] };
}

async function requestAdminJson<T>(
  fetcher: AdminFetch,
  input: RequestInfo | URL,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetcher(input, {
    ...init,
    headers: {
      ...adminHeaders(Boolean(init.body)),
      ...(init.headers ?? {}),
    },
  });

  const body = await readResponseBody(response);

  if (!response.ok) {
    const { error, messages } = parseAdminErrorBody(body);
    throw new AdminApiError(response.status, error, messages);
  }

  return body as T;
}

function toChapterDraftPayload(draft: AdminChapterDraftForm): AdminChapterDraft {
  return {
    id: draft.id.trim(),
    chapterId: draft.chapterId.trim(),
    status: draft.status,
    name: draft.name.trim(),
    state: draft.state.trim(),
    websiteUrl: blankToUndefined(draft.websiteUrl),
    groupEmail: blankToUndefined(draft.groupEmail),
    reviewNotes: blankToUndefined(draft.reviewNotes),
    reviewedAt: blankToUndefined(draft.reviewedAt),
  };
}

function toEndorsementDraftPayload(draft: AdminEndorsementDraftForm): AdminEndorsementDraft {
  return {
    id: draft.id.trim(),
    endorsementId: draft.endorsementId.trim(),
    status: draft.status,
    candidateId: draft.candidateId.trim(),
    raceId: draft.raceId.trim(),
    chapterId: draft.chapterId.trim(),
    voterGuideId: blankToUndefined(draft.voterGuideId),
    endorsementType: draft.endorsementType,
    sourceUrl: draft.sourceUrl.trim(),
    sourceDate: blankToUndefined(draft.sourceDate),
    confidence: draft.confidence,
    notes: blankToUndefined(draft.notes),
    reviewNotes: blankToUndefined(draft.reviewNotes),
    reviewedAt: blankToUndefined(draft.reviewedAt),
  };
}

export function emptyAdminChapterDraftForm(): AdminChapterDraftForm {
  return {
    id: '',
    chapterId: '',
    status: 'draft',
    name: '',
    state: '',
    websiteUrl: '',
    groupEmail: '',
    reviewNotes: '',
    reviewedAt: '',
  };
}

export function emptyAdminEndorsementDraftForm(): AdminEndorsementDraftForm {
  return {
    id: '',
    endorsementId: '',
    status: 'draft',
    candidateId: '',
    raceId: '',
    chapterId: '',
    voterGuideId: '',
    endorsementType: 'endorsed',
    sourceUrl: '',
    sourceDate: '',
    confidence: 'official_source',
    notes: '',
    reviewNotes: '',
    reviewedAt: '',
  };
}

export function adminChapterDraftToForm(draft: Partial<AdminChapterDraft> = {}): AdminChapterDraftForm {
  return {
    id: draft.id ?? '',
    chapterId: draft.chapterId ?? '',
    status: draft.status ?? 'draft',
    name: draft.name ?? '',
    state: draft.state ?? '',
    websiteUrl: draft.websiteUrl ?? '',
    groupEmail: draft.groupEmail ?? '',
    reviewNotes: draft.reviewNotes ?? '',
    reviewedAt: draft.reviewedAt ?? '',
  };
}

export function adminEndorsementDraftToForm(draft: Partial<AdminEndorsementDraft> = {}): AdminEndorsementDraftForm {
  return {
    id: draft.id ?? '',
    endorsementId: draft.endorsementId ?? '',
    status: draft.status ?? 'draft',
    candidateId: draft.candidateId ?? '',
    raceId: draft.raceId ?? '',
    chapterId: draft.chapterId ?? '',
    voterGuideId: draft.voterGuideId ?? '',
    endorsementType: draft.endorsementType ?? 'endorsed',
    sourceUrl: draft.sourceUrl ?? '',
    sourceDate: draft.sourceDate ?? '',
    confidence: draft.confidence ?? 'official_source',
    notes: draft.notes ?? '',
    reviewNotes: draft.reviewNotes ?? '',
    reviewedAt: draft.reviewedAt ?? '',
  };
}

export async function listAdminChapterDrafts(fetcher: AdminFetch = fetch): Promise<AdminChapterDraft[]> {
  return requestAdminJson<AdminChapterDraft[]>(fetcher, '/api/admin/chapter-drafts');
}

export async function loadAdminChapterDraft(
  id: string,
  fetcher: AdminFetch = fetch,
): Promise<AdminChapterDraft> {
  return requestAdminJson<AdminChapterDraft>(fetcher, `/api/admin/chapter-drafts/${encodeURIComponent(id)}`);
}

export async function saveAdminChapterDraft(
  draft: AdminChapterDraftForm,
  fetcher: AdminFetch = fetch,
): Promise<AdminChapterDraft> {
  return requestAdminJson<AdminChapterDraft>(fetcher, '/api/admin/chapter-drafts', {
    method: 'POST',
    body: JSON.stringify(toChapterDraftPayload(draft)),
  });
}

export async function publishAdminChapterDraft(id: string, fetcher: AdminFetch = fetch): Promise<void> {
  await requestAdminJson(fetcher, `/api/admin/chapter-drafts/${encodeURIComponent(id)}/publish`, {
    method: 'POST',
  });
}

export async function archiveAdminChapterDraft(id: string, fetcher: AdminFetch = fetch): Promise<void> {
  await requestAdminJson(fetcher, `/api/admin/chapter-drafts/${encodeURIComponent(id)}/archive`, {
    method: 'POST',
  });
}

export async function listAdminEndorsementDrafts(fetcher: AdminFetch = fetch): Promise<AdminEndorsementDraft[]> {
  return requestAdminJson<AdminEndorsementDraft[]>(fetcher, '/api/admin/endorsement-drafts');
}

export async function loadAdminEndorsementDraft(
  id: string,
  fetcher: AdminFetch = fetch,
): Promise<AdminEndorsementDraft> {
  return requestAdminJson<AdminEndorsementDraft>(fetcher, `/api/admin/endorsement-drafts/${encodeURIComponent(id)}`);
}

export async function saveAdminEndorsementDraft(
  draft: AdminEndorsementDraftForm,
  fetcher: AdminFetch = fetch,
): Promise<AdminEndorsementDraft> {
  return requestAdminJson<AdminEndorsementDraft>(fetcher, '/api/admin/endorsement-drafts', {
    method: 'POST',
    body: JSON.stringify(toEndorsementDraftPayload(draft)),
  });
}

export async function publishAdminEndorsementDraft(id: string, fetcher: AdminFetch = fetch): Promise<void> {
  await requestAdminJson(fetcher, `/api/admin/endorsement-drafts/${encodeURIComponent(id)}/publish`, {
    method: 'POST',
  });
}

export async function archiveAdminEndorsementDraft(id: string, fetcher: AdminFetch = fetch): Promise<void> {
  await requestAdminJson(fetcher, `/api/admin/endorsement-drafts/${encodeURIComponent(id)}/archive`, {
    method: 'POST',
  });
}
