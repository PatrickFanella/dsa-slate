# Civic Elections Dropdown Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Replace the free-text Civic election ID field with a dropdown backed by Google Civic `/elections`, so voters can select a supported election before address lookup.

**Architecture:** Extend the existing backend Civic client/service with a read-only `Elections` method and expose `GET /api/civic/elections` when `GOOGLE_CIVIC_API_KEY` is configured. Add frontend Civic API helpers and update `CivicLookupPanel` to load elections, render a dropdown, and submit the selected Google election ID with the address. Keep address autocomplete and address persistence out of scope.

**Tech Stack:** Go backend, Google Civic Information API v2 `/elections`, Vite React TypeScript, existing Civic lookup panel, existing no-address-storage privacy boundary.

---

## Requirements summary

Current Civic lookup can return `Election unknown` when Google cannot infer an election from address alone. Google’s recommended workflow is to list available elections, then call voter info with a selected `electionId` when needed. The current UI exposes a free-text “Election ID” field; this should become a dropdown of Google-supported elections.

## Scope boundaries

### In scope

- Backend `GET /api/civic/elections` endpoint.
- Google Civic `/elections` client call using server-side `GOOGLE_CIVIC_API_KEY`.
- Frontend helper to load elections.
- Civic panel dropdown for available elections.
- Submit selected election ID with address lookup.
- Error/empty states for unconfigured Civic API and no elections returned.
- Tests using fake Google/service responses; no live Google calls in tests.

### Explicitly out of scope

- No address autocomplete.
- No Google Places API.
- No address storage.
- No user accounts or saved profile changes.
- No election caching table.
- No cron/sync job.
- No promise that every user address has a current election.

## File map

- Modify: `backend/internal/civic/store.go` — add elections DTOs and service method.
- Modify: `backend/internal/civic/google.go` — add `/elections` client call and mapping.
- Modify: `backend/internal/civic/google_test.go` — add fake HTTP tests for elections URL/mapping/errors.
- Modify: `backend/internal/httpapi/server.go` — add `GET /api/civic/elections` route.
- Modify: `backend/internal/httpapi/server_test.go` — add unconfigured/success/error route tests.
- Modify: `web/src/data/civicApi.ts` — add election list helper/types.
- Modify: `web/src/data/civicApi.test.ts` — add frontend helper tests.
- Modify: `web/src/components/CivicLookupPanel.tsx` — replace free-text election ID input with dropdown.
- Modify: `web/src/App.test.tsx` — update homepage assertions.
- Modify: `docs/reference/civic-api.md` — document elections endpoint and dropdown workflow.
- Modify: `docs/reference/api.md` — document endpoint and errors.
- Modify: `README.md` — mention election dropdown when key is configured.

---

## Public contract

### Request

```http
GET /api/civic/elections
```

### Success response

```json
{
  "source": "google_civic_information_api",
  "elections": [
    {"id":"2000","name":"VIP Test Election","electionDay":"2031-12-06"},
    {"id":"9438","name":"Colorado Primary Election","electionDay":"2026-06-30"}
  ]
}
```

### Errors

Missing API key:

```json
{"error":"civic_not_configured"}
```

External lookup failure:

```json
{"error":"civic_lookup_failed","message":"Google Civic elections lookup failed"}
```

---

## Task 1: Add Civic elections service/client support

**Files:**
- Modify: `backend/internal/civic/store.go`
- Modify: `backend/internal/civic/google.go`
- Modify: `backend/internal/civic/google_test.go`

- [ ] **Step 1: Add DTOs and service method**

In `backend/internal/civic/store.go`, add:

```go
type ElectionsResponse struct {
	Source    string     `json:"source"`
	Elections []Election `json:"elections"`
}
```

Extend `LookupService`:

```go
type LookupService interface {
	Lookup(ctx context.Context, request LookupRequest) (LookupResponse, error)
	Elections(ctx context.Context) (ElectionsResponse, error)
}
```

- [ ] **Step 2: Add failing Google client tests**

In `backend/internal/civic/google_test.go`, add:

```go
func TestGoogleClientElections(t *testing.T) {
	const responseBody = `{"elections":[{"id":"2000","name":"VIP Test Election","electionDay":"2031-12-06"},{"id":"9438","name":"Colorado Primary Election","electionDay":"2026-06-30"}]}`
	var gotRequest *http.Request
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotRequest = r.Clone(r.Context())
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(responseBody)), Header: make(http.Header), Request: r}, nil
	})
	client := NewGoogleClient("fake-key", &http.Client{Transport: transport})

	got, err := client.Elections(context.Background())
	if err != nil {
		t.Fatalf("Elections() error = %v", err)
	}
	if gotRequest.URL.Path != "/civicinfo/v2/elections" {
		t.Fatalf("path = %q, want /civicinfo/v2/elections", gotRequest.URL.Path)
	}
	if gotRequest.URL.Query().Get("key") != "fake-key" {
		t.Fatalf("key query = %q, want fake-key", gotRequest.URL.Query().Get("key"))
	}
	if got.Source != SourceGoogleCivicInformationAPI {
		t.Fatalf("source = %q, want google source", got.Source)
	}
	if len(got.Elections) != 2 || got.Elections[1].ID != "9438" || got.Elections[1].Name != "Colorado Primary Election" {
		t.Fatalf("elections = %#v, want mapped elections", got.Elections)
	}
}

func TestGoogleClientElectionsReturnsUpstreamErrorMessage(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusForbidden, Status: "403 Forbidden", Body: io.NopCloser(strings.NewReader(`{"error":{"code":403,"message":"API key not authorized"}}`)), Header: make(http.Header), Request: r}, nil
	})
	client := NewGoogleClient("fake-key", &http.Client{Transport: transport})

	_, err := client.Elections(context.Background())
	if err == nil {
		t.Fatal("expected upstream error")
	}
	upstreamErr, ok := err.(UpstreamError)
	if !ok {
		t.Fatalf("err = %T, want UpstreamError", err)
	}
	if upstreamErr.StatusCode != http.StatusForbidden || upstreamErr.Message != "API key not authorized" {
		t.Fatalf("upstream error = %#v", upstreamErr)
	}
}
```

- [ ] **Step 3: Implement Google elections call**

In `backend/internal/civic/google.go`, add:

```go
const googleCivicElectionsURL = "https://www.googleapis.com/civicinfo/v2/elections"
```

Add response type:

```go
type googleElectionsResponse struct {
	Elections []googleElection `json:"elections"`
}
```

Add method:

```go
func (c *GoogleClient) Elections(ctx context.Context) (ElectionsResponse, error) {
	if c.apiKey == "" {
		return ElectionsResponse{}, errors.New("google civic api key is required")
	}

	electionsURL, err := url.Parse(googleCivicElectionsURL)
	if err != nil {
		return ElectionsResponse{}, fmt.Errorf("parse google civic elections url: %w", err)
	}
	query := electionsURL.Query()
	query.Set("key", c.apiKey)
	electionsURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, electionsURL.String(), nil)
	if err != nil {
		return ElectionsResponse{}, fmt.Errorf("build google civic elections request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ElectionsResponse{}, fmt.Errorf("do google civic elections request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ElectionsResponse{}, UpstreamError{StatusCode: resp.StatusCode, Message: readGoogleErrorMessage(resp)}
	}

	var payload googleElectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ElectionsResponse{}, fmt.Errorf("decode google civic elections response: %w", err)
	}

	return ElectionsResponse{Source: SourceGoogleCivicInformationAPI, Elections: mapElections(payload.Elections)}, nil
}
```

- [ ] **Step 4: Run Civic tests**

```bash
cd backend && go test ./internal/civic -count=1
```

Expected: PASS.

---

## Task 2: Add backend `GET /api/civic/elections`

**Files:**
- Modify: `backend/internal/httpapi/server.go`
- Modify: `backend/internal/httpapi/server_test.go`

- [ ] **Step 1: Add fake service support in tests**

Update `testCivicService` in `server_test.go` with:

```go
electionsResponse civic.ElectionsResponse
electionsErr      error
```

Add method:

```go
func (s *testCivicService) Elections(ctx context.Context) (civic.ElectionsResponse, error) {
	if s.electionsErr != nil {
		return civic.ElectionsResponse{}, s.electionsErr
	}
	return s.electionsResponse, nil
}
```

- [ ] **Step 2: Add route tests**

In `server_test.go`, add:

```go
func TestCivicElectionsRoute(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", &testCivicService{electionsResponse: civic.ElectionsResponse{Source: civic.SourceGoogleCivicInformationAPI, Elections: []civic.Election{{ID: "2000", Name: "VIP Test Election", ElectionDay: "2031-12-06"}}}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/civic/elections", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got civic.ElectionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Elections) != 1 || got.Elections[0].ID != "2000" {
		t.Fatalf("response = %#v, want one election", got)
	}
}

func TestCivicElectionsNotConfigured(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/civic/elections", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestCivicElectionsMethodNotAllowed(t *testing.T) {
	handler := NewServerWithServices(directory.NewMemoryStore(directory.DirectoryData{}), nil, "", &testCivicService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/civic/elections", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", got)
	}
}
```

- [ ] **Step 3: Implement route**

In `ServeHTTP`, add before `/api/civic/lookup`:

```go
case "/api/civic/elections":
	if r.Method != http.MethodGet {
		s.methodNotAllowedAllow(w, http.MethodGet)
		return
	}
	s.civicElections(w, r)
```

Add handler:

```go
func (s *Server) civicElections(w http.ResponseWriter, r *http.Request) {
	if s.civicService == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "civic_not_configured"})
		return
	}

	response, err := s.civicService.Elections(r.Context())
	if err != nil {
		statusCode := http.StatusBadGateway
		message := "Google Civic elections lookup failed"
		var upstreamErr civic.UpstreamError
		if errors.As(err, &upstreamErr) {
			message = upstreamErr.Message
			if upstreamErr.StatusCode >= http.StatusBadRequest && upstreamErr.StatusCode < http.StatusInternalServerError {
				statusCode = http.StatusBadRequest
			}
		}
		s.writeJSON(w, statusCode, map[string]string{"error": "civic_lookup_failed", "message": message})
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}
```

- [ ] **Step 4: Run HTTP tests**

```bash
cd backend && go test ./internal/httpapi -count=1
```

Expected: PASS.

---

## Task 3: Add frontend election list helper

**Files:**
- Modify: `web/src/data/civicApi.ts`
- Modify: `web/src/data/civicApi.test.ts`

- [ ] **Step 1: Add failing helper tests**

In `web/src/data/civicApi.test.ts`, add:

```ts
it('loads Civic elections', async () => {
  const fetcher = vi.fn(async () =>
    jsonResponse({
      source: 'google_civic_information_api',
      elections: [{ id: '2000', name: 'VIP Test Election', electionDay: '2031-12-06' }],
    }),
  );

  const response = await fetchCivicElections(fetcher);

  expect(fetcher).toHaveBeenCalledWith('/api/civic/elections', {
    headers: { Accept: 'application/json' },
  });
  expect(response.elections[0].id).toBe('2000');
});

it('throws typed errors when Civic elections are not configured', async () => {
  const fetcher = vi.fn(async () => jsonResponse({ error: 'civic_not_configured' }, 503));

  await expect(fetchCivicElections(fetcher)).rejects.toMatchObject({
    code: 'civic_not_configured',
    status: 503,
  });
});
```

Update imports:

```ts
import { CivicLookupApiError, fetchCivicElections, lookupCivicInformation } from './civicApi';
```

- [ ] **Step 2: Add frontend types/helper**

In `web/src/data/civicApi.ts`, add:

```ts
export type CivicElectionsResponse = {
  source: string;
  elections: CivicElection[];
};
```

Add validator:

```ts
function isCivicElectionsResponse(body: unknown): body is CivicElectionsResponse {
  return Boolean(body && typeof body === 'object' && Array.isArray((body as { elections?: unknown }).elections));
}
```

Add helper:

```ts
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
```

- [ ] **Step 3: Run frontend helper tests**

```bash
pnpm --dir web run test -- src/data/civicApi.test.ts
```

Expected: PASS.

---

## Task 4: Replace free-text election ID with dropdown

**Files:**
- Modify: `web/src/components/CivicLookupPanel.tsx`
- Modify: `web/src/App.test.tsx`

- [ ] **Step 1: Add render assertions**

In `web/src/App.test.tsx`, replace the old `Election ID` assertion with:

```ts
expect(html).toContain('Civic election');
expect(html).toContain('Load supported elections');
expect(html).toContain('Choose an election from Google Civic before looking up an address when possible.');
expect(html).not.toContain('Enter an optional election ID');
```

- [ ] **Step 2: Update component imports/state**

In `CivicLookupPanel.tsx`, update imports:

```ts
import {
  CivicLookupApiError,
  civicRegistrationDisclaimer,
  fetchCivicElections,
  lookupCivicInformation,
  type CivicAddress,
  type CivicElection,
  type CivicLookupResponse,
  type CivicVotingLocation,
} from '../data/civicApi';
```

Add state:

```ts
const [elections, setElections] = useState<CivicElection[]>([]);
const [electionsLoaded, setElectionsLoaded] = useState(false);
```

- [ ] **Step 3: Add load elections action**

Inside `CivicLookupPanel`, add:

```ts
async function loadElections() {
  setNotice({ kind: 'info', title: 'Loading supported Civic elections…' });
  try {
    const nextElections = await fetchCivicElections();
    setElections(nextElections.elections);
    setElectionsLoaded(true);
    setNotice({
      kind: 'success',
      title: nextElections.elections.length > 0 ? 'Supported elections loaded' : 'No supported elections returned',
      messages:
        nextElections.elections.length > 0
          ? ['Choose an election from Google Civic before looking up an address when possible.']
          : ['Google Civic did not return any currently supported elections. You can still try an address lookup without selecting one.'],
    });
  } catch (error) {
    if (error instanceof CivicLookupApiError) {
      setNotice({
        kind: 'error',
        title: error.code === 'civic_not_configured' ? 'Civic lookup not configured' : 'Election list failed',
        messages: error.messages.length > 0 ? error.messages : ['Unable to load supported Civic elections.'],
      });
      return;
    }
    setNotice({ kind: 'error', title: 'Election list failed', messages: ['Unable to load supported Civic elections.'] });
  }
}
```

- [ ] **Step 4: Replace election text input with dropdown + button**

Replace the `FieldShell` for `civic-election-id` with:

```tsx
<div className="flex flex-col gap-2 text-sm font-bold text-foreground">
  <span className="text-xs font-black uppercase tracking-[0.24em] text-primary">Civic election</span>
  <div className="grid gap-2 md:grid-cols-[1fr_auto]">
    <select
      id="civic-election-id"
      className="rounded-none border-2 border-border-strong bg-surface px-3 py-3 text-foreground focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      value={electionId}
      onChange={(event) => setElectionId(event.target.value)}
    >
      <option value="">Let Civic choose</option>
      {elections.map((election) => (
        <option key={election.id ?? `${election.name}-${election.electionDay}`} value={election.id ?? ''}>
          {[election.name, election.electionDay, election.id ? `ID ${election.id}` : undefined].filter(Boolean).join(' · ')}
        </option>
      ))}
    </select>
    <button
      className="rounded-none border-2 border-border-strong bg-surface px-4 py-3 font-black text-foreground transition hover:bg-surface-accent focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      type="button"
      onClick={loadElections}
    >
      Load supported elections
    </button>
  </div>
  <span className="text-xs font-normal leading-5 text-muted">
    Choose an election from Google Civic before looking up an address when possible.
    {electionsLoaded ? ` ${elections.length} supported elections loaded.` : ''}
  </span>
</div>
```

Update initial notice message to:

```ts
'Load supported elections, then choose one before looking up an address when possible.',
```

- [ ] **Step 5: Run App tests**

```bash
pnpm --dir web run test -- src/App.test.tsx
```

Expected: PASS.

---

## Task 5: Update docs and verify

**Files:**
- Modify: `docs/reference/civic-api.md`
- Modify: `docs/reference/api.md`
- Modify: `README.md`

- [ ] **Step 1: Update Civic reference**

Add to `docs/reference/civic-api.md` after the endpoint section:

```markdown
## Supported elections

`GET /api/civic/elections` lists elections currently supported by Google Civic. The public lookup panel uses this endpoint to populate the election dropdown. Selecting one sends its Google `electionId` with the address lookup and can avoid `Election unknown` responses when Google cannot infer an election from address alone.
```

- [ ] **Step 2: Update API docs**

In `docs/reference/api.md`, add:

```markdown
- `GET /api/civic/elections` — lists Google Civic-supported elections when `GOOGLE_CIVIC_API_KEY` is configured.
```

- [ ] **Step 3: Update README**

Update the Civic bullet to mention the dropdown:

```markdown
- The public homepage includes a Civic lookup panel with a supported-election dropdown; submitted addresses stay in browser component state only.
```

- [ ] **Step 4: Full verification**

```bash
make verify
git diff --check
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/civic backend/internal/httpapi/server.go backend/internal/httpapi/server_test.go web/src/data/civicApi.ts web/src/data/civicApi.test.ts web/src/components/CivicLookupPanel.tsx web/src/App.test.tsx docs/reference/civic-api.md docs/reference/api.md README.md docs/superpowers/plans/2026-06-29-civic-elections-dropdown.md
git commit -m "feat: add civic elections dropdown"
```

---

## Self-review

- Spec coverage: replaces manual election ID entry with Google-supported election selection.
- Privacy: no address autocomplete, no Google Places, no address storage, no accounts.
- UX: keeps manual address input but makes election selection discoverable.
- Verification: backend fake-client tests, frontend helper/component tests, `make verify`, and `git diff --check` are required.
