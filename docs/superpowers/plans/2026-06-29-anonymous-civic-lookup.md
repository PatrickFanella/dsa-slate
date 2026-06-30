# Anonymous Civic Lookup Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Add an anonymous `POST /api/civic/lookup` endpoint that returns polling-place and registration-link information from Google Civic without storing raw addresses.

**Architecture:** Add a small backend Civic client package with a testable HTTP client interface, map Google Civic `voterInfoQuery` responses into a privacy-safe DSA Slate response shape, then expose it through the existing HTTP API. Keep all lookup data transient; no DB tables, user accounts, saved addresses, or frontend profile persistence in this slice.

**Tech Stack:** Go backend, standard-library `net/http`, Google Civic Information API v2, existing HTTP server tests, `.env.example` config, Vite frontend only for an optional lightweight lookup panel if time permits.

---

## Requirements summary

User asked whether an API can check voter registration status, polling place, and registration links. Findings:

- Google Civic can return polling locations, early vote sites, drop-off locations, election officials, contests, registration URLs, and registration-confirmation URLs.
- Google Civic cannot reliably prove that a person is registered to vote.
- NASS/Vote.gov are good official fallback links for registration and registration-status confirmation.

Therefore this slice must describe itself as **voting information lookup**, not “registration check.”

## Scope boundaries

### In scope

- Backend endpoint: `POST /api/civic/lookup`.
- Server-side Google Civic API key use via `GOOGLE_CIVIC_API_KEY`.
- Request body with `address` and optional `electionId`.
- Response with normalized input, election, polling locations, early vote sites, drop-off locations, election office links, registration links, and a clear registration-status disclaimer.
- Tests using a fake Civic HTTP client; no live Google calls in tests.
- Docs updates explaining limitations and no raw address storage.

### Explicitly out of scope

- No raw address storage.
- No user accounts or saved profiles.
- No encrypted address-at-rest design yet.
- No frontend saved slate/profile integration.
- No Google OAuth/sign-in.
- No claim that the API confirms registration status.
- No dependence on live Google API in `make verify`.

## File map

- Create: `backend/internal/civic/store.go` — DTOs and `LookupService` interface.
- Create: `backend/internal/civic/google.go` — Google Civic client and response mapping.
- Create: `backend/internal/civic/google_test.go` — client mapping/URL/privacy tests with fake HTTP transport.
- Modify: `backend/internal/app/app.go` — add `GoogleCivicAPIKey()` helper.
- Modify: `backend/cmd/app/main.go` — wire Civic service when the key is configured.
- Modify: `backend/internal/httpapi/server.go` — add `POST /api/civic/lookup` route and handler.
- Modify: `backend/internal/httpapi/server_test.go` — route tests for missing config, validation, and successful lookup with fake service.
- Modify: `docs/reference/civic-api.md` — update implemented contract and disclaimer.
- Modify: `docs/reference/api.md` — move endpoint from planned to available when configured.
- Modify: `README.md` — document local env and privacy behavior.
- Optional modify: `web/src/App.tsx` and tests — only if implementing a small public lookup panel in the same slice. Recommended to defer UI to the next slice.

---

## Public contract

### Request

```http
POST /api/civic/lookup
Content-Type: application/json
```

```json
{
  "address": "123 Main St, San Diego, CA",
  "electionId": "2000"
}
```

`electionId` is optional. If absent and Google returns `otherElections`, the frontend can ask the user to pick one and submit a second request.

### Success response

```json
{
  "status": "success",
  "source": "google_civic_information_api",
  "disclaimer": "This lookup does not confirm that you are registered to vote. Use your state election office registration-status link to confirm.",
  "normalizedInput": {
    "line1": "123 MAIN ST",
    "city": "SAN DIEGO",
    "state": "CA",
    "zip": "92101"
  },
  "election": {
    "id": "2000",
    "name": "VIP Test Election",
    "electionDay": "2026-06-02"
  },
  "otherElections": [],
  "pollingLocations": [],
  "earlyVoteSites": [],
  "dropOffLocations": [],
  "contests": [],
  "stateElectionOffice": {
    "name": "California Secretary of State",
    "electionInfoUrl": "https://www.sos.ca.gov/elections/",
    "registrationUrl": "https://registertovote.ca.gov/",
    "registrationConfirmationUrl": "https://voterstatus.sos.ca.gov/"
  },
  "fallbackLinks": {
    "registerToVoteUrl": "https://vote.gov/",
    "nassRegistrationStatusUrl": "https://www.nass.org/can-I-vote/voter-registration-status",
    "nassPollingPlaceUrl": "https://www.nass.org/can-i-vote/find-your-polling-place"
  }
}
```

### Errors

Missing API key:

```json
{"error":"civic_not_configured"}
```

Validation failure:

```json
{"error":"validation_failed","messages":["address is required"]}
```

External lookup failure:

```json
{"error":"civic_lookup_failed","message":"Google Civic lookup failed"}
```

---

## Task 1: Add Civic DTOs and Google client mapping

**Files:**
- Create: `backend/internal/civic/store.go`
- Create: `backend/internal/civic/google.go`
- Create: `backend/internal/civic/google_test.go`

- [ ] **Step 1: Create DTOs/interface**

Create `backend/internal/civic/store.go`:

```go
package civic

import "context"

const RegistrationDisclaimer = "This lookup does not confirm that you are registered to vote. Use your state election office registration-status link to confirm."

type LookupRequest struct {
	Address    string `json:"address"`
	ElectionID string `json:"electionId,omitempty"`
}

type LookupResponse struct {
	Status              string              `json:"status"`
	Source              string              `json:"source"`
	Disclaimer          string              `json:"disclaimer"`
	NormalizedInput     Address             `json:"normalizedInput,omitempty"`
	Election            Election            `json:"election,omitempty"`
	OtherElections      []Election          `json:"otherElections"`
	PollingLocations    []VotingLocation    `json:"pollingLocations"`
	EarlyVoteSites      []VotingLocation    `json:"earlyVoteSites"`
	DropOffLocations    []VotingLocation    `json:"dropOffLocations"`
	Contests            []Contest           `json:"contests"`
	StateElectionOffice StateElectionOffice `json:"stateElectionOffice,omitempty"`
	FallbackLinks       FallbackLinks       `json:"fallbackLinks"`
}

type Address struct {
	LocationName string `json:"locationName,omitempty"`
	Line1        string `json:"line1,omitempty"`
	Line2        string `json:"line2,omitempty"`
	Line3        string `json:"line3,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Zip          string `json:"zip,omitempty"`
}

type Election struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	ElectionDay string `json:"electionDay,omitempty"`
}

type VotingLocation struct {
	Address      Address  `json:"address"`
	PollingHours string   `json:"pollingHours,omitempty"`
	StartDate    string   `json:"startDate,omitempty"`
	EndDate      string   `json:"endDate,omitempty"`
	Sources      []Source `json:"sources"`
}

type Source struct {
	Name     string `json:"name,omitempty"`
	Official bool   `json:"official"`
}

type Contest struct {
	ID          string   `json:"id,omitempty"`
	Type        string   `json:"type,omitempty"`
	Office      string   `json:"office,omitempty"`
	BallotTitle string   `json:"ballotTitle,omitempty"`
	District    string   `json:"district,omitempty"`
	Candidates  []string `json:"candidates"`
}

type StateElectionOffice struct {
	Name                            string `json:"name,omitempty"`
	ElectionInfoURL                 string `json:"electionInfoUrl,omitempty"`
	RegistrationURL                 string `json:"registrationUrl,omitempty"`
	RegistrationConfirmationURL     string `json:"registrationConfirmationUrl,omitempty"`
	AbsenteeVotingInfoURL           string `json:"absenteeVotingInfoUrl,omitempty"`
	BallotInfoURL                   string `json:"ballotInfoUrl,omitempty"`
}

type FallbackLinks struct {
	RegisterToVoteURL          string `json:"registerToVoteUrl"`
	NASSRegistrationStatusURL  string `json:"nassRegistrationStatusUrl"`
	NASSPollingPlaceURL        string `json:"nassPollingPlaceUrl"`
}

type LookupService interface {
	Lookup(ctx context.Context, request LookupRequest) (LookupResponse, error)
}
```

- [ ] **Step 2: Write failing mapping tests**

Create `backend/internal/civic/google_test.go` with tests that:

- call the client with address `123 Main St, San Diego, CA`
- verify the outgoing URL is `/civicinfo/v2/voterinfo`
- verify `address`, `key`, `officialOnly=true`, and optional `electionId` query params
- return a fake Google JSON response with polling location and registration links
- assert `LookupResponse.Disclaimer` contains `does not confirm`
- assert no raw request address is copied into a custom persisted/audit field

- [ ] **Step 3: Implement Google client**

Create `backend/internal/civic/google.go` with:

- `type HTTPClient interface { Do(*http.Request) (*http.Response, error) }`
- `func NewGoogleClient(apiKey string, httpClient HTTPClient) *GoogleClient`
- `func (c *GoogleClient) Lookup(ctx context.Context, request LookupRequest) (LookupResponse, error)`
- URL: `https://www.googleapis.com/civicinfo/v2/voterinfo`
- Query params: `address`, `key`, `officialOnly=true`, optional `electionId`
- Return validation error for blank address before external call
- Map only the response fields in this plan.

- [ ] **Step 4: Run Civic tests**

```bash
cd backend && go test ./internal/civic -count=1
```

Expected: PASS.

---

## Task 2: Add HTTP route with fake-service tests

**Files:**
- Modify: `backend/internal/httpapi/server.go`
- Modify: `backend/internal/httpapi/server_test.go`

- [ ] **Step 1: Extend server with optional Civic service**

Add `civic.LookupService` to `Server` and a constructor:

```go
func NewServerWithServices(store directory.Store, adminStore admin.Store, adminToken string, civicService civic.LookupService) http.Handler
```

Keep `NewServer` and `NewServerWithAdmin` working.

- [ ] **Step 2: Add route tests**

In `server_test.go`, add tests for:

- `GET /api/civic/lookup` returns method-not-allowed.
- `POST /api/civic/lookup` with no configured service returns `503 {"error":"civic_not_configured"}`.
- Blank address returns `400 {"error":"validation_failed","messages":["address is required"]}`.
- Valid address calls fake service and returns the disclaimer and polling locations.

- [ ] **Step 3: Implement handler**

Add to `ServeHTTP`:

```go
case "/api/civic/lookup":
	if r.Method != http.MethodPost {
		s.methodNotAllowedAllow(w, http.MethodPost)
		return
	}
	s.civicLookup(w, r)
```

Handler behavior:

- decode strict JSON into `civic.LookupRequest`
- trim/validate address
- call service
- return JSON
- do not log or store the address

- [ ] **Step 4: Run HTTP tests**

```bash
cd backend && go test ./internal/httpapi -count=1
```

Expected: PASS.

---

## Task 3: Wire config and app startup

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/cmd/app/main.go`
- Modify: `.env.example`

- [ ] **Step 1: Add config helper**

In `backend/internal/app/app.go`:

```go
func GoogleCivicAPIKey() string {
	return os.Getenv("GOOGLE_CIVIC_API_KEY")
}
```

- [ ] **Step 2: Wire service when key exists**

In `backend/cmd/app/main.go`, create a Civic service only when `GOOGLE_CIVIC_API_KEY` is non-empty:

```go
civicAPIKey := appcfg.GoogleCivicAPIKey()
var civicService civic.LookupService
if civicAPIKey != "" {
	civicService = civic.NewGoogleClient(civicAPIKey, http.DefaultClient)
}
```

Pass it into the new server constructor while preserving admin wiring.

- [ ] **Step 3: Confirm `.env.example` already has key**

No real key goes into git. `.env.example` should keep:

```env
GOOGLE_CIVIC_API_KEY=
```

- [ ] **Step 4: Run backend tests**

```bash
cd backend && go test ./...
```

Expected: PASS.

---

## Task 4: Update docs and verify

**Files:**
- Modify: `docs/reference/civic-api.md`
- Modify: `docs/reference/api.md`
- Modify: `README.md`

- [ ] **Step 1: Update Civic docs**

Revise `docs/reference/civic-api.md` to say the endpoint is implemented when `GOOGLE_CIVIC_API_KEY` is configured and that it returns polling/registration links but does not confirm registration.

- [ ] **Step 2: Update API docs**

Move `POST /api/civic/lookup` from “planned” to an available optional endpoint. Include request/response/error examples.

- [ ] **Step 3: Update README**

Add a bullet under Development/API:

```markdown
- Set `GOOGLE_CIVIC_API_KEY` to enable anonymous `/api/civic/lookup`; the endpoint does not store raw addresses and does not confirm voter registration status.
```

- [ ] **Step 4: Full verification**

```bash
make verify
git diff --check
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/civic backend/internal/httpapi/server.go backend/internal/httpapi/server_test.go backend/internal/app/app.go backend/cmd/app/main.go docs/reference/civic-api.md docs/reference/api.md README.md docs/superpowers/plans/2026-06-29-anonymous-civic-lookup.md
git commit -m "feat: add anonymous civic lookup endpoint"
```

---

## Follow-up slice: Civic lookup UI

After the backend endpoint lands, add a public UI panel that:

- accepts an address without saving it
- displays polling locations, early vote sites, drop boxes, election office links, registration-status link, and register-to-vote link
- clearly says “This does not confirm registration”
- optionally supports `otherElections` selection and a second lookup by `electionId`

Keep saved profiles/address storage out of that follow-up until the private-alpha PRD and privacy/security design are complete.

---

## Self-review

- Spec coverage: answers polling place + register/check-registration links while avoiding false registration confirmation.
- Privacy: no raw address storage, no DB writes, no logging requirement added.
- Scope control: backend endpoint only; public UI and saved profiles are deferred.
- Verification: fake-service/client tests, backend tests, full `make verify`, and `git diff --check` are required.
