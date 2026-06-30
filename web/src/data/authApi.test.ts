import { describe, expect, it, vi } from 'vitest';
import { fetchAuthSession, logout } from './authApi';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('authApi', () => {
  it('fetches the auth session with JSON headers', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ authenticated: false }));

    const state = await fetchAuthSession(fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/auth/session', expect.objectContaining({ headers: expect.any(Headers) }));
    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect([...new Headers(options?.headers).entries()]).toEqual([['accept', 'application/json']]);
    expect(state).toEqual({ authenticated: false });
  });

  it('logs out through the auth endpoint', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ authenticated: false }));

    const state = await logout(fetcher);

    expect(fetcher).toHaveBeenCalledWith('/api/auth/logout', expect.objectContaining({ headers: expect.any(Headers), method: 'POST' }));
    const [, options] = fetcher.mock.calls[0] as unknown as [RequestInfo | URL, RequestInit?];
    expect([...new Headers(options?.headers).entries()]).toEqual([['accept', 'application/json']]);
    expect(state).toEqual({ authenticated: false });
  });

  it('throws for non-2xx auth responses', async () => {
    const fetcher = vi.fn(async () => jsonResponse({ error: 'nope' }, 503));

    await expect(fetchAuthSession(fetcher)).rejects.toThrow('Auth API request failed with status 503');
  });
});
