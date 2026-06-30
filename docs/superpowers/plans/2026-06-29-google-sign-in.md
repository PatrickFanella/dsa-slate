# Google Sign-In Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add non-intrusive Google OIDC sign-in that creates local DSA Slate users and opaque app sessions without storing passwords.

**Architecture:** Use Google Authorization Code flow with `openid email profile`, `state`, `nonce`, and PKCE. Google verifies identity only; DSA Slate keeps local `users`, `user_oauth_identities`, `user_sessions`, and role authorization. The backend starts/callbacks the OAuth flow, verifies ID tokens, creates a local HTTP-only `dsa_slate_session` cookie, and the frontend only links/buttons to those backend routes.

**Tech Stack:** Go backend, PostgreSQL, `golang.org/x/oauth2`, `google.golang.org/api/idtoken`, existing `backend/internal/auth` session foundation, Vite React TypeScript frontend.

---

## Requirements summary

User asked for a non-intrusive auth service. Decision: use Google sign-in/OIDC first because Google credentials already exist and this avoids password storage/reset flows. Authorization remains local in DSA Slate roles.

## Scope boundaries

### In scope

- Google OIDC sign-in start/callback endpoints.
- OAuth state, nonce, and PKCE verifier temporary cookies.
- ID token verification.
- Local user creation/update.
- Local OAuth identity table keyed by Google `sub`.
- Local opaque `dsa_slate_session` cookie.
- Logout endpoint that deletes session and clears cookie.
- Minimal frontend sign-in/sign-out/session status UI.
- Docs/env updates.

### Explicitly out of scope

- No email/password.
- No magic links.
- No address storage.
- No saved slate yet.
- No admin route migration to role-gated auth in this slice.
- No refresh tokens or offline Google access.
- No Google API scopes beyond `openid email profile`.

## File map

- Modify: `migrations/0001_init.sql` — add `user_oauth_identities` table.
- Modify: `backend/go.mod`, `backend/go.sum` — add OAuth/OIDC deps.
- Modify: `backend/internal/app/app.go` — add Google OAuth config helpers and public app URL helper.
- Modify: `.env.example` — add redirect/app URL placeholders if missing.
- Modify: `backend/internal/auth/session.go` — extend store interface for OAuth identity helpers.
- Modify: `backend/internal/auth/postgres.go` — implement OAuth identity methods.
- Create: `backend/internal/auth/google.go` — OAuth config, temp cookies, ID token verifier interface, callback exchange helpers.
- Create: `backend/internal/auth/google_test.go` — state/nonce/PKCE/callback unit tests.
- Modify: `backend/internal/httpapi/server.go` — add auth routes.
- Modify: `backend/internal/httpapi/server_test.go` — route tests with fake Google auth service.
- Create: `web/src/data/authApi.ts` — frontend session/logout helper.
- Create: `web/src/data/authApi.test.ts` — helper tests.
- Modify: `web/src/App.tsx` — add sign-in/session status shell.
- Modify: `web/src/App.test.tsx` — public render assertions.
- Modify: `docs/reference/api.md` — document auth endpoints.
- Modify: `docs/product/private-alpha-architecture.md` — mark Google sign-in slice.
- Modify: `README.md` — local config and redirect URI instructions.

---

## API contract

### Start Google sign-in

```http
GET /api/auth/google/start
```

Behavior:

- requires `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, and `APP_BASE_URL`
- sets temporary HTTP-only cookies for state, nonce, and PKCE verifier
- redirects to Google auth endpoint

### Google callback

```http
GET /api/auth/google/callback?state=<state>&code=<code>
```

Behavior:

- validates state cookie
- exchanges code using PKCE verifier
- verifies ID token audience and nonce
- upserts local user and `user_oauth_identities`
- creates `dsa_slate_session`
- redirects to `/`

### Logout

```http
POST /api/auth/logout
```

Behavior:

- deletes session if cookie exists
- clears `dsa_slate_session`
- returns `{authenticated:false}`

### Session

Already implemented:

```http
GET /api/auth/session
```

---

## Task 1: Add OAuth identity schema and store methods

**Files:**
- Modify: `migrations/0001_init.sql`
- Modify: `backend/internal/auth/session.go`
- Modify: `backend/internal/auth/postgres.go`
- Modify: `backend/internal/auth/auth_test.go`

- [ ] **Step 1: Add migration table**

Add after `user_sessions` in `migrations/0001_init.sql`:

```sql
CREATE TABLE IF NOT EXISTS user_oauth_identities (
  provider text NOT NULL,
  provider_subject text NOT NULL,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (provider, provider_subject)
);

CREATE INDEX IF NOT EXISTS user_oauth_identities_user_id_idx ON user_oauth_identities(user_id);
```

- [ ] **Step 2: Add DTO and store methods**

In `backend/internal/auth/session.go`, add:

```go
type OAuthIdentity struct {
	Provider        string    `json:"provider"`
	ProviderSubject string    `json:"providerSubject"`
	UserID          string    `json:"userId"`
	CreatedAt       time.Time `json:"createdAt"`
}
```

Extend `Store`:

```go
UpsertOAuthIdentity(context.Context, OAuthIdentity) (OAuthIdentity, error)
GetOAuthIdentity(context.Context, string, string) (OAuthIdentity, error)
```

- [ ] **Step 3: Implement Postgres methods**

In `backend/internal/auth/postgres.go`, add `UpsertOAuthIdentity`, `GetOAuthIdentity`, and `scanOAuthIdentity` using `provider`, `provider_subject`, `user_id`, `created_at`.

- [ ] **Step 4: Update memory test store**

In `backend/internal/auth/auth_test.go`, add no-op/in-memory implementations to satisfy the interface.

- [ ] **Step 5: Run auth tests**

```bash
cd backend && go test ./internal/auth -count=1
```

Expected: PASS.

---

## Task 2: Add Google OAuth config and unit-tested helpers

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`
- Create: `backend/internal/auth/google.go`
- Create: `backend/internal/auth/google_test.go`

- [ ] **Step 1: Add deps**

Run:

```bash
cd backend && go get golang.org/x/oauth2 google.golang.org/api/idtoken
```

- [ ] **Step 2: Add Google auth helper types**

Create `backend/internal/auth/google.go` with:

- `GoogleOAuthConfig`
- `GoogleUserInfo`
- `GoogleTokenVerifier` interface
- temp cookie names:
  - `dsa_google_oauth_state`
  - `dsa_google_oauth_nonce`
  - `dsa_google_oauth_pkce`
- `NewGoogleOAuthConfig(...)`
- `BuildGoogleAuthURL(...)`
- `NewTemporaryOAuthCookie(...)`
- `ClearTemporaryOAuthCookie(...)`
- `VerifyGoogleIDToken(...)` wrapper that uses `idtoken.Validate`

Generate state/nonce/PKCE with `GenerateSessionToken()` or a dedicated crypto-random helper.

- [ ] **Step 3: Add helper tests**

Create `backend/internal/auth/google_test.go` with tests that assert:

- auth URL contains `openid email profile`
- auth URL contains state and nonce
- auth URL uses PKCE S256 challenge
- temp cookies are HTTP-only and SameSite=Lax
- user ID generated from Google subject is stable, e.g. `google:<sub>` or another deterministic local ID strategy

- [ ] **Step 4: Run auth tests**

```bash
cd backend && go test ./internal/auth -count=1
```

Expected: PASS.

---

## Task 3: Add auth app config

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `.env.example`

- [ ] **Step 1: Add config helpers**

In `backend/internal/app/app.go`, add:

```go
func AppBaseURL() string {
	return envOrDefault("APP_BASE_URL", "http://localhost:5173")
}

func GoogleOAuthClientID() string {
	return os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
}

func GoogleOAuthClientSecret() string {
	return os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
}

func GoogleOAuthRedirectURL() string {
	return envOrDefault("GOOGLE_OAUTH_REDIRECT_URL", strings.TrimRight(AppBaseURL(), "/")+"/api/auth/google/callback")
}
```

Add `strings` import if not already present.

- [ ] **Step 2: Update `.env.example`**

Ensure these placeholders exist:

```env
APP_BASE_URL=http://localhost:5173
GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:5173/api/auth/google/callback
```

- [ ] **Step 3: Run app tests**

```bash
cd backend && go test ./internal/app -count=1
```

Expected: PASS.

---

## Task 4: Add backend auth routes with fakeable service

**Files:**
- Modify: `backend/internal/httpapi/server.go`
- Modify: `backend/internal/httpapi/server_test.go`
- Modify: `backend/cmd/app/main.go`

- [ ] **Step 1: Add auth service abstraction**

Keep route code testable by adding an interface in `server.go` or `auth/google.go`:

```go
type GoogleSignInService interface {
	Start(http.ResponseWriter, *http.Request)
	Callback(http.ResponseWriter, *http.Request) error
}
```

If implemented in `auth`, create a concrete service that has:

- OAuth config
- auth store
- session duration
- secure cookie flag

- [ ] **Step 2: Add routes**

In `ServeHTTP`:

- `GET /api/auth/google/start`
- `GET /api/auth/google/callback`
- `POST /api/auth/logout`

Keep `GET /api/auth/session` as-is.

- [ ] **Step 3: Callback behavior**

Callback must:

1. validate state cookie against query `state`
2. exchange code using PKCE verifier
3. verify ID token and nonce
4. upsert user with role default `user`
5. upsert OAuth identity provider `google`
6. create local session
7. set `dsa_slate_session`
8. clear temp OAuth cookies
9. redirect to `/`

- [ ] **Step 4: Logout behavior**

Logout must:

1. delete session if present
2. clear session cookie
3. return `auth.State{}`

- [ ] **Step 5: Route tests**

In `server_test.go`, add tests for:

- start route redirects when service configured
- start route returns `auth_not_configured` when missing config
- callback rejects missing/wrong state
- callback sets session cookie with fake verifier/exchanger
- logout clears cookie and deletes session

- [ ] **Step 6: Wire app startup**

In `backend/cmd/app/main.go`, create Google sign-in service only when client ID/secret are set. If not set, auth session remains available but Google start returns not configured.

- [ ] **Step 7: Run HTTP tests**

```bash
cd backend && go test ./internal/httpapi -count=1
```

Expected: PASS.

---

## Task 5: Add frontend session/sign-in shell

**Files:**
- Create: `web/src/data/authApi.ts`
- Create: `web/src/data/authApi.test.ts`
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.test.tsx`

- [ ] **Step 1: Add frontend auth helpers**

Create `web/src/data/authApi.ts`:

```ts
export type AuthUser = { id: string; email: string; role: 'user' | 'editor' | 'admin' };
export type AuthState = { authenticated: boolean; user?: AuthUser };

export async function fetchAuthSession(fetcher = fetch): Promise<AuthState> {
  const response = await fetcher('/api/auth/session', { headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error('auth session failed');
  return response.json();
}

export async function logout(fetcher = fetch): Promise<AuthState> {
  const response = await fetcher('/api/auth/logout', { method: 'POST', headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error('logout failed');
  return response.json();
}
```

- [ ] **Step 2: Add helper tests**

Test `/api/auth/session` and `/api/auth/logout` calls.

- [ ] **Step 3: Add UI shell**

In `App.tsx`, add a compact auth status/sign-in area in the header:

- “Sign in with Google” link to `/api/auth/google/start`
- after later session fetch: show user email and sign out button

For this slice, it is acceptable to render the sign-in link without fully reactive session fetch if that keeps scope small. If session fetch is added, keep it client-only and do not break SSR tests.

- [ ] **Step 4: Update tests**

Assert homepage contains “Sign in with Google”.

- [ ] **Step 5: Run frontend tests**

```bash
pnpm --dir web run test -- src/data/authApi.test.ts src/App.test.tsx
```

Expected: PASS.

---

## Task 6: Docs, verify, commit

**Files:**
- Modify: `docs/reference/api.md`
- Modify: `docs/product/private-alpha-architecture.md`
- Modify: `README.md`

- [ ] **Step 1: Document Google redirect setup**

Add to README:

```markdown
For local Google sign-in, configure this redirect URI in Google Cloud Console:

`http://localhost:5173/api/auth/google/callback`
```

Because Vite proxies `/api` to the Go backend, the browser-facing redirect stays on port 5173 in development.

- [ ] **Step 2: Document endpoints**

In `docs/reference/api.md`, document:

- `GET /api/auth/google/start`
- `GET /api/auth/google/callback`
- `POST /api/auth/logout`
- existing `GET /api/auth/session`

- [ ] **Step 3: Full verification**

```bash
make verify
git diff --check
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add .env.example backend/go.mod backend/go.sum backend/cmd/app/main.go backend/internal/app/app.go backend/internal/auth backend/internal/httpapi/server.go backend/internal/httpapi/server_test.go migrations/0001_init.sql web/src/data/authApi.ts web/src/data/authApi.test.ts web/src/App.tsx web/src/App.test.tsx docs/reference/api.md docs/product/private-alpha-architecture.md README.md docs/superpowers/plans/2026-06-29-google-sign-in.md
git commit -m "feat: add google sign-in"
```

---

## Self-review

- Spec coverage: adds non-intrusive Google sign-in with local sessions and roles.
- Security: uses state, nonce, PKCE, verified ID token, HTTP-only local session cookie.
- Scope control: no password auth, no saved slate, no address storage, no admin role migration yet.
- Verification: backend fake-service tests, auth helper tests, frontend tests, `make verify`, and `git diff --check` are required.
