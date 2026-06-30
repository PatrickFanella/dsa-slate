# DSA Slate API

The public API is read-only JSON for source-linked electoral guide data. Admin draft endpoints are token-protected and local-only.

Private-alpha session foundation is implemented via `GET /api/auth/session` and the HTTP-only `dsa_slate_session` cookie name. Google OIDC sign-in now sits on top of that session model.

## Non-goals

- No public write endpoints.
- No raw voter addresses or address lookup.
- No personally identifying voter data.

## Health

`GET /healthz`

```json
{"ok":true,"service":"dsa-slate"}
```

## Aggregate directory

`GET /api/directory`

Returns all public directory collections:

- `chapters`
- `elections`
- `races`
- `candidates`
- `voterGuides`
- `endorsements`
- `ballotMeasures`
- `measurePositions`

All empty collections are JSON arrays, not `null`.

## Collection endpoints

- `GET /api/chapters` — all known DSA chapters, including optional public contact and social fields when present.
- `GET /api/elections`
- `GET /api/races?electionId=<id>&jurisdiction=<name>&level=<level>`
- `GET /api/candidates`
- `GET /api/voter-guides?chapterId=<id>&electionId=<id>`
- `GET /api/endorsements?chapterId=<id>&electionId=<id>&raceId=<id>`
- `GET /api/ballot-measures?electionId=<id>`
- `GET /api/measure-positions?chapterId=<id>&measureId=<id>`
- `GET /api/civic/elections` — lists Google Civic-supported elections when `GOOGLE_CIVIC_API_KEY` is configured.

## Anonymous Civic lookup endpoint

`POST /api/civic/lookup` is available when `GOOGLE_CIVIC_API_KEY` is configured. It returns polling-place and registration links from Google Civic Information API without storing raw addresses and without confirming voter registration status.

See `docs/reference/civic-api.md` for the privacy and response contract.

## Auth session foundation

`GET /api/auth/session`

- Returns anonymous state when the `dsa_slate_session` cookie is absent, invalid, or expired.
- Returns `{authenticated:true,user:{id,email,role}}` when a valid session exists.

## Google OIDC sign-in

`GET /api/auth/google/start`

- Requires `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, and `APP_BASE_URL`/`GOOGLE_OAUTH_REDIRECT_URL`.
- Sets temporary HTTP-only `dsa_google_oauth_state`, `dsa_google_oauth_nonce`, and `dsa_google_oauth_pkce` cookies.
- Redirects to Google with `openid email profile` only.

`GET /api/auth/google/callback`

- Validates the `state` query against the temporary cookie.
- Exchanges the authorization code with PKCE.
- Verifies the ID token and nonce.
- Creates/updates the local user, `user_oauth_identities`, and `dsa_slate_session`.
- New Google identities may auto-link to existing local `user` accounts by verified email, but not to existing `editor` or `admin` accounts.
- Clears the temporary OAuth cookies and redirects to `/`.

`POST /api/auth/logout`

- Deletes the current session when present.
- Clears the `dsa_slate_session` cookie.
- Returns `{authenticated:false}`.

## Admin draft endpoints

All admin endpoints require:

```http
Authorization: Bearer <ADMIN_API_TOKEN>
```

They are intended for local/editorial use only and operate on draft records before publishing into the public read tables.
The frontend draft console is available at `/admin` and should be used with a session-only bearer token from `ADMIN_API_TOKEN`.

- `GET /api/admin/chapter-drafts`
- `POST /api/admin/chapter-drafts`
- `GET /api/admin/chapter-drafts/{id}`
- `PUT /api/admin/chapter-drafts/{id}`
- `POST /api/admin/chapter-drafts/{id}/publish`
- `POST /api/admin/chapter-drafts/{id}/archive`
- `GET /api/admin/endorsement-drafts`
- `POST /api/admin/endorsement-drafts`
- `GET /api/admin/endorsement-drafts/{id}`
- `PUT /api/admin/endorsement-drafts/{id}`
- `POST /api/admin/endorsement-drafts/{id}/publish`
- `POST /api/admin/endorsement-drafts/{id}/archive`

Unauthorized requests return:

```json
{"error":"unauthorized"}
```

Validation errors return:

```json
{"error":"validation_failed","messages":["..."]}
```

## Error shape

```json
{"error":"not_found"}
```

```json
{"error":"invalid_query","message":"level must be one of federal, statewide, state_legislature, local, unknown"}
```

Anonymous Civic lookup errors (and the supported-election list endpoint):

```json
{"error":"civic_not_configured"}
```

```json
{"error":"validation_failed","messages":["address is required"]}
```

```json
{"error":"civic_lookup_failed","message":"Google Civic lookup failed"}
```

## Source-link requirements

Endorsements, voter guides, and measure positions must include `sourceUrl` and `confidence`.
Confidence values are:

- `official_source`
- `official_social_source`
- `campaign_source`
- `press_source`
- `needs_review`

## Frontend browse behavior

The React app uses collection endpoints for structured browse filters:

- chapter filters map to voter guides, endorsements, and measure positions.
- election filters map to races, voter guides, endorsements, and ballot measures.
- race-level filters map to races; the frontend prunes endorsements whose races are filtered out because endorsements do not currently expose a direct `level` query parameter.
- free-text search remains client-side over rendered card fields.

The app never sends or stores raw voter addresses.
