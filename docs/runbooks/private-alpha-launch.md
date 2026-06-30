# Private Alpha Launch Runbook

## Purpose

Use this runbook to decide whether a DSA Slate private-alpha build is ready for invited testers. The alpha validates the source-linked directory, Google sign-in, saved slate, anonymous Civic lookup, and authenticated editor workflow without storing raw voter addresses.

## Scope

In scope:

- public directory browsing and record detail review
- anonymous Civic lookup against supported Google Civic elections
- Google sign-in and sign-out
- signed-in saved slate for endorsements, measure positions, and voter guides
- `/admin` chapter and endorsement draft workflows for `editor`/`admin` users
- data-quality and full verification checks

Out of scope:

- public launch traffic
- raw saved addresses or voter profiles containing addresses
- public submissions
- email/password auth
- scraper-assisted imports
- election-night results automation

## Preconditions

- Local `.env` is present and uncommitted.
- `APP_BASE_URL` matches the web origin used for testing.
- `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, and `GOOGLE_OAUTH_REDIRECT_URL` are configured for Google sign-in.
- `GOOGLE_CIVIC_API_KEY` is configured if Civic lookup is part of the test pass.
- The database is reachable and seeded with source-linked records.
- At least one invited tester account can sign in as a normal `user`.
- At least one editorial account has role `editor` or `admin` and an explicitly linked Google OAuth identity.
- No test plan depends on storing raw voter addresses.

## Launch checklist

1. Pull the intended branch and inspect local state.
2. Run database setup:

   ```bash
   make db-up
   make migrate
   make seed
   ```

3. Run acceptance checks:

   ```bash
   make data-quality
   make verify
   ```

4. Start the app locally:

   ```bash
   make dev-api
   make dev-web
   ```

5. Open the web app at `APP_BASE_URL`.
6. Complete the tester flow below.
7. Record blockers with reproduction steps and affected role.
8. Do not invite testers until all critical blockers are fixed and verification is green.

## Tester flow

### Anonymous browse

1. Open `/` in a fresh browser session.
2. Confirm the page shows the source-linked directory and methodology/disclaimer copy.
3. Search for a candidate, chapter, office, election, or voter guide.
4. Use chapter, election, and race-level filters.
5. Open record details for:
   - one endorsement,
   - one ballot-measure position,
   - one voter guide.
6. Confirm every public record shows source/confidence context and does not claim to be an election authority.
7. Confirm save actions are sign-in CTAs, not dead disabled controls.

Expected result: anonymous users can browse and inspect source-linked records but cannot save records without signing in.

### Anonymous Civic lookup

1. On `/`, open the Civic lookup panel.
2. Load supported elections.
3. Select an election where possible.
4. Enter a test address.
5. Submit lookup.
6. Confirm the page states that DSA Slate does not confirm registration status.
7. Confirm the address is not shown as saved profile data after refresh.

Expected result: lookup returns official election links when Civic data is available, keeps no raw saved address, and preserves the registration-status disclaimer.

### Signed-in saved slate

1. Click “Sign in with Google”.
2. Confirm the signed-in session banner shows the tester email and role.
3. Save one endorsement, one ballot-measure position, and one voter guide.
4. Open `/me/slate`.
5. Confirm saved records appear and the page says only public record references are stored.
6. Remove one saved record.
7. Sign out or switch account.
8. Confirm `/me/slate` asks the visitor to sign in again.

Expected result: saved slate persists by signed-in user, stores public record references only, and supports removal and sign-out.

### Editor/admin draft flow

1. Sign in as an account with role `editor` or `admin`.
2. Open `/admin`.
3. Confirm the signed-in session panel shows the editor/admin role.
4. Create or load a chapter draft.
5. Create or load an endorsement draft.
6. Publish only source-reviewed test data.
7. Confirm validation failures are visible when required fields are omitted.
8. Confirm a normal `user` role is blocked from editing.

Expected result: editor/admin users can draft and publish through the session-gated admin UI; normal users and anonymous visitors cannot.

### Negative and privacy checks

- Anonymous `/admin` access shows a sign-in gate.
- Normal signed-in `user` access to `/admin` shows “Editor or admin role required.”
- Browser UI does not ask for an admin bearer token.
- Saved slate pages do not display or store raw voter addresses.
- Civic lookup copy says registration status is not confirmed.
- Cookie-authenticated writes continue to work only through same-origin app requests.

## Pass/fail criteria

Pass when all are true:

- `make verify` passes.
- `make data-quality` passes after seed or migration changes.
- Anonymous browse, Civic lookup, saved slate, and editor/admin tester flows pass.
- Privacy copy is visible in the Civic lookup and saved-slate flows.
- No raw voter address storage was introduced.
- No shared bearer token is exposed in the browser admin UI.

Fail and stop tester invites when any are true:

- source-linked public data is missing source URLs or confidence labels,
- admin/editor writes are available to anonymous or normal users,
- saved slate stores or exposes raw addresses,
- Civic lookup implies registration status is confirmed,
- verification or data-quality checks fail.

## Cleanup and rollback

- Remove test saved slate records from test accounts if needed.
- Archive accidental admin drafts rather than editing public seed data directly.
- Re-run `make seed` against local development data to restore known fixtures.
- If a launched alpha build is unsafe, roll back to the last commit where `make verify` and tester flow passed.

## References

- PRD: `docs/product/private-alpha-prd.md`
- Architecture plan: `docs/product/private-alpha-architecture.md`
- API contract: `docs/reference/api.md`
- Civic privacy reference: `docs/reference/civic-api.md`
- Editorial data rules: `docs/reference/editorial-data.md`
