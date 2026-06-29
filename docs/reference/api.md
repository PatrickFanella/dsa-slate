# DSA Slate Read API

The API is a public, read-only JSON API for source-linked electoral guide data.

## Non-goals

- No write/admin endpoints yet.
- No authentication yet.
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

- `GET /api/chapters`
- `GET /api/elections`
- `GET /api/races?electionId=<id>&jurisdiction=<name>&level=<level>`
- `GET /api/candidates`
- `GET /api/voter-guides?chapterId=<id>&electionId=<id>`
- `GET /api/endorsements?chapterId=<id>&electionId=<id>&raceId=<id>`
- `GET /api/ballot-measures?electionId=<id>`
- `GET /api/measure-positions?chapterId=<id>&measureId=<id>`

## Error shape

```json
{"error":"not_found"}
```

```json
{"error":"invalid_query","message":"level must be one of federal, statewide, state_legislature, local, unknown"}
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
