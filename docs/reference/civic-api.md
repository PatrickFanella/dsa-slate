# Google Civic API Reference

DSA Slate uses the Google Civic Information API for anonymous voting-information lookup. This integration must preserve the project privacy rule: do not store raw voter addresses unless a future opt-in account feature has a privacy/security review.

## Current phase: anonymous lookup only

Use only `GOOGLE_CIVIC_API_KEY` for server-side calls to Google Civic Information API endpoints.

The public homepage includes an anonymous lookup panel that posts address input directly to this endpoint and keeps the address in browser component state only.

Expected behavior:

1. User enters an address in the browser.
2. Browser sends it to a backend lookup endpoint over HTTPS.
3. Backend calls Google Civic API with the configured API key.
4. Backend returns normalized election, polling-place, and election-office context.
5. Backend does not persist the raw address.
6. Logs must not include the raw address.

The endpoint does **not** confirm whether a user is registered to vote. It provides official registration-status and polling-place links instead.

The OAuth client ID/secret are not required for anonymous Civic API lookup. Keep OAuth credentials reserved for a later account/sign-in slice.

## Supported elections

`GET /api/civic/elections` lists elections currently supported by Google Civic. The public lookup panel uses this endpoint to populate the election dropdown. Selecting one sends its Google `electionId` with the address lookup and can avoid `Election unknown` responses when Google cannot infer an election from address alone.

## Future phases

### Accounts without address storage

Add accounts only for non-sensitive preferences first:

- preferred chapter
- state
- theme
- followed elections

Do not store raw addresses in this phase.

### Opt-in saved address

Before storing any raw address, complete a privacy/security review covering:

- explicit opt-in copy
- address encryption at rest
- log redaction
- delete/export controls
- retention policy
- admin access controls
- incident response expectations

## Implemented anonymous endpoint

```http
POST /api/civic/lookup
Content-Type: application/json
```

Request:

```json
{"address":"123 Main St, San Diego, CA"}
```

Response shape:

```json
{
  "status": "success",
  "source": "google_civic_information_api",
  "disclaimer": "This lookup does not confirm that you are registered to vote. Use your state election office registration-status link to confirm.",
  "normalizedInput": {
    "state": "CA",
    "city": "San Diego"
  },
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

Do not include the full submitted address in the response unless the user needs to confirm it in the same request flow. Never write it to database-backed user or audit records in the anonymous phase.

## Supported elections

`GET /api/civic/elections` lists elections currently supported by Google Civic. The public lookup panel uses this endpoint to populate the election dropdown. Selecting one sends its Google `electionId` with the address lookup and can avoid `Election unknown` responses when Google cannot infer an election from address alone.

## Environment variables

```env
GOOGLE_CIVIC_API_KEY=
GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
```

`GOOGLE_CIVIC_API_KEY` is the only variable needed for the first lookup implementation. OAuth variables are placeholders for a later account feature.
