# DSA Slate API

The public API is read-only JSON for source-linked electoral guide data. Admin draft endpoints are token-protected and local-only.

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

## Anonymous Civic lookup endpoint

`POST /api/civic/lookup` is available when `GOOGLE_CIVIC_API_KEY` is configured. It returns polling-place and registration links from Google Civic Information API without storing raw addresses and without confirming voter registration status.

See `docs/reference/civic-api.md` for the privacy and response contract.

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

Anonymous Civic lookup errors:

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
