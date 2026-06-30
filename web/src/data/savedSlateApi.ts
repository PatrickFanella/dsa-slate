export type SavedSlateItem = {
  id: string;
  userId: string;
  endorsementId: string;
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

export async function saveSlateItem(endorsementId: string, fetcher: typeof fetch = fetch): Promise<SavedSlateItem> {
  return requestSavedSlateJson<SavedSlateItem>(fetcher, '/api/me/slate/items', {
    method: 'PUT',
    body: JSON.stringify({ endorsementId }),
  });
}

export async function deleteSlateItem(itemId: string, fetcher: typeof fetch = fetch): Promise<{ ok: boolean }> {
  return requestSavedSlateJson<{ ok: boolean }>(fetcher, `/api/me/slate/items/${encodeURIComponent(itemId)}`, {
    method: 'DELETE',
  });
}
