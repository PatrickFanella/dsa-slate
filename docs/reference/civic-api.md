# Google Civic API Reference

DSA Slate may use the Google Civic Information API for anonymous election and district lookup. This integration must preserve the project privacy rule: do not store raw voter addresses unless a future opt-in account feature has a privacy/security review.

## Current phase: anonymous lookup only

Use only `GOOGLE_CIVIC_API_KEY` for server-side calls to Google Civic Information API endpoints.

Expected behavior:

1. User enters an address in the browser.
2. Browser sends it to a backend lookup endpoint over HTTPS.
3. Backend calls Google Civic API with the configured API key.
4. Backend returns normalized election/district/official context.
5. Backend does not persist the raw address.
6. Logs must not include the raw address.

The OAuth client ID/secret are not required for anonymous Civic API lookup. Keep OAuth credentials reserved for a later account/sign-in slice.

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

## Proposed anonymous endpoint

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
  "normalizedInput": {
    "state": "CA",
    "city": "San Diego"
  },
  "elections": [],
  "districts": [],
  "officials": [],
  "source": "google_civic_information_api"
}
```

Do not include the full submitted address in the response unless the user needs to confirm it in the same request flow. Never write it to database-backed user or audit records in the anonymous phase.

## Environment variables

```env
GOOGLE_CIVIC_API_KEY=
GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
```

`GOOGLE_CIVIC_API_KEY` is the only variable needed for the first lookup implementation. OAuth variables are placeholders for a later account feature.
