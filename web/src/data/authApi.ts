export type AuthUser = {
  id: string;
  email: string;
  role: 'user' | 'editor' | 'admin';
};

export type AuthState = {
  authenticated: boolean;
  user?: AuthUser;
};

async function fetchAuthJson<T>(fetcher: typeof fetch, input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set('Accept', 'application/json');

  const response = await fetcher(input, {
    ...init,
    headers,
  });

  if (!response.ok) {
    throw new Error(`Auth API request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export async function fetchAuthSession(fetcher: typeof fetch = fetch): Promise<AuthState> {
  return fetchAuthJson<AuthState>(fetcher, '/api/auth/session');
}

export async function logout(fetcher: typeof fetch = fetch): Promise<AuthState> {
  return fetchAuthJson<AuthState>(fetcher, '/api/auth/logout', { method: 'POST', headers: { 'X-DSA-Slate-CSRF': '1' } });
}
