export type SavedSlateRecordType = 'endorsement' | 'measure_position' | 'voter_guide';

export type SavedSlateItem = {
  id: string;
  userId: string;
  recordType: SavedSlateRecordType;
  recordId: string;
  createdAt: string;
  updatedAt: string;
};

export type SavedSlate = {
  items: SavedSlateItem[];
};

async function requestSavedSlateJson<T>(fetcher: typeof fetch, input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set('Accept', 'application/json');
  if (init?.body) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetcher(input, { ...init, headers });
  if (!response.ok) {
    throw new Error(`Saved slate API request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export async function fetchSavedSlate(fetcher: typeof fetch = fetch): Promise<SavedSlate> {
  return requestSavedSlateJson<SavedSlate>(fetcher, '/api/me/slate');
}

export async function saveSlateItem(recordType: SavedSlateRecordType, recordId: string, fetcher: typeof fetch = fetch): Promise<SavedSlateItem> {
  return requestSavedSlateJson<SavedSlateItem>(fetcher, '/api/me/slate/items', {
    method: 'PUT',
    headers: { 'X-DSA-Slate-CSRF': '1' },
    body: JSON.stringify({ recordType, recordId }),
  });
}

export async function deleteSlateItem(itemId: string, fetcher: typeof fetch = fetch): Promise<{ ok: boolean }> {
  return requestSavedSlateJson<{ ok: boolean }>(fetcher, `/api/me/slate/items/${encodeURIComponent(itemId)}`, {
    method: 'DELETE',
    headers: { 'X-DSA-Slate-CSRF': '1' },
  });
}
