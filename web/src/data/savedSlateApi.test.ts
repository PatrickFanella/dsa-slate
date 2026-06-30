import { describe, expect, it, vi } from 'vitest';
import { deleteSlateItem, fetchSavedSlate, saveSlateItem } from './savedSlateApi';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('savedSlateApi', () => {
  it('fetches the current saved slate', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ items: [] }));

    const slate = await fetchSavedSlate(fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/me/slate', expect.objectContaining({ headers: expect.any(Headers) }));
    expect(slate).toEqual({ items: [] });
  });

  it('saves an endorsement reference', async () => {
    const body = { id: 'item-1', userId: 'user-1', recordType: 'endorsement', recordId: 'endorsement-1', createdAt: '2026-06-30T00:00:00Z', updatedAt: '2026-06-30T00:00:00Z' };
    const fetcher = vi.fn(async () => jsonResponse(body));

    const item = await saveSlateItem('endorsement', 'endorsement-1', fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/me/slate/items', expect.objectContaining({ method: 'PUT', headers: expect.any(Headers) }));
    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect(options?.body).toBe(JSON.stringify({ recordType: 'endorsement', recordId: 'endorsement-1' }));
    expect([...new Headers(options?.headers).entries()]).toEqual([
      ['accept', 'application/json'],
      ['content-type', 'application/json'],
      ['x-dsa-slate-csrf', '1'],
    ]);
    expect(item).toEqual(body);
  });

  it('saves measure position and voter guide references', async () => {
    const fetcher = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      const body = JSON.parse(String(init?.body)) as { recordType: string; recordId: string };
      return jsonResponse({ id: `item-${body.recordId}`, userId: 'user-1', ...body, createdAt: '2026-06-30T00:00:00Z', updatedAt: '2026-06-30T00:00:00Z' });
    });

    await expect(saveSlateItem('measure_position', 'measure-position-1', fetcher)).resolves.toMatchObject({ recordType: 'measure_position', recordId: 'measure-position-1' });
    await expect(saveSlateItem('voter_guide', 'voter-guide-1', fetcher)).resolves.toMatchObject({ recordType: 'voter_guide', recordId: 'voter-guide-1' });
  });

  it('deletes a saved slate item', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ ok: true }));

    const result = await deleteSlateItem('item/1', fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/me/slate/items/item%2F1', expect.objectContaining({ method: 'DELETE' }));
    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect(new Headers(options?.headers).get('X-DSA-Slate-CSRF')).toBe('1');
    expect(result).toEqual({ ok: true });
  });

  it('throws for non-2xx responses', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ error: 'unauthorized' }, 401));

    await expect(fetchSavedSlate(fetcher)).rejects.toThrow('Saved slate API request failed with status 401');
  });
});
