# DSA Slate Private Alpha PRD

## Product vision

DSA Slate is a voter guide app that helps people find, understand, save, and act on DSA-endorsed candidates, ballot measure positions, and chapter voter guides.

The private alpha should prove that DSA Slate can combine three things safely:

1. source-linked DSA electoral data,
2. official civic lookup data,
3. user-specific saved slate workflows.

The product must not imply that DSA Slate is an election authority. It links to official voter registration, polling place, and election office sources and preserves DSA source wording without upgrading recommendations into endorsements.

## Current foundation

Already built:

- public read API for chapters, elections, races, candidates, voter guides, endorsements, ballot measures, and measure positions
- chapter directory seed imported into Postgres
- California cluster electoral seed data
- source-linked detail panel
- review metadata in details
- data-quality command
- session-gated admin draft console for chapter and endorsement drafts
- anonymous Civic lookup endpoint and public lookup panel
- Civic supported-election dropdown
- Google sign-in with local user sessions
- saved-slate API and UI for public endorsement, measure-position, and voter-guide references
- private-alpha launch runbook with tester flow and pass/fail checks
- no raw address storage in current runtime

## Private alpha goals

The private alpha should answer these questions:

1. Can an invited user create an account and maintain a useful saved slate?
2. Can the app connect a voter’s address-derived civic context to DSA records without unsafe address handling?
3. Can editors add and publish trusted records without direct database edits?
4. Can we explain source confidence, review status, and official-vs-DSA data clearly enough for real users?

## Primary users and roles

### Anonymous voter

Can:

- browse DSA records
- search/filter the directory
- use anonymous Civic lookup
- follow registration/polling-place links

Cannot:

- save a slate
- store profile data
- submit public endorsement suggestions in private alpha

### Signed-in voter

Can:

- save candidates, ballot measure positions, and voter guides to “My Slate”
- save non-sensitive preferences such as preferred chapter, state, and election
- run Civic lookup from a signed-in session
- eventually opt into a full profile after privacy work is complete

Private alpha starts with saved slate and preferences. Raw saved addresses are gated behind a separate privacy/security milestone.

### Editor

Can:

- use admin draft workflows
- create/update chapter and endorsement drafts
- publish reviewed drafts
- archive drafts
- inspect source/review metadata

### Admin

Can:

- do everything an editor can
- manage users and roles once auth exists
- manage future source registry/scraper/public-submission settings

### Public submitter

Post-alpha role. Public submissions are not in private alpha.

## Core workflows

### Find my slate

1. User lands on DSA Slate.
2. User searches/browses by chapter, election, and race level.
3. User can optionally use Civic lookup to find official election context.
4. User opens detail panels to verify source, confidence, and review metadata.
5. User saves relevant records if signed in.
6. User can return to the directory, My Slate, or Admin through primary navigation.

### Save my slate

1. User signs in.
2. User saves endorsements, measure positions, and voter guides.
3. User sees a “My Slate” view.
4. User can remove saved items.
5. User can export/print/share a non-sensitive slate summary.

### Manage profile

Private alpha profile fields:

- display name or email-derived label
- preferred chapter
- preferred state
- followed elections
- theme preference if useful

Deferred profile fields:

- raw address
- phone number
- district history

### Anonymous Civic lookup

1. User loads supported elections.
2. User optionally selects a Google Civic election.
3. User enters an address.
4. Backend calls Google Civic.
5. App shows official election/polling/registration links.
6. App clearly says lookup does not confirm registration.
7. App does not persist the address.

### Future opt-in address profile

Only after privacy/security design:

1. User explicitly opts into storing address.
2. Address is encrypted at rest.
3. Logs redact address fields.
4. User can delete/export address data.
5. Admin access is restricted and audited.

### Review and publish electoral data

1. Editor opens `/admin`.
2. Editor uses an authenticated `editor` or `admin` session.
3. Editor creates a draft.
4. Draft is validated.
5. Editor publishes reviewed draft.
6. Public read tables update.
7. Data quality checks remain green.

## Functional requirements

### Authentication

Private alpha must support authenticated users.

Desired long-term methods:

- email/password
- Google sign-in

Implementation recommendation:

1. implement one auth method first,
2. add role-based sessions,
3. protect admin console with roles,
4. add second auth method after the session model is stable.

### Saved slate

Private alpha must support saving:

- endorsement IDs
- measure position IDs
- voter guide IDs
- optional notes or labels later

Saved slate records must reference public IDs, not duplicate source text.
Current implementation saves endorsement, measure-position, and voter-guide IDs.

### Civic integration

Private alpha must support:

- list supported Google Civic elections
- anonymous address lookup
- selected election ID lookup
- official registration-status links
- official registration links
- polling locations when available
- early vote sites when available
- drop-off locations when available

Private alpha must not claim to verify voter registration status.

### Admin/editorial

Private alpha must support:

- admin/editor role access
- chapter draft publish workflow
- endorsement draft publish workflow
- audit log for admin write actions
- source/review metadata

Post-alpha should add:

- public submissions
- source registry
- scraper-assisted draft creation
- richer review queues

## Non-functional requirements

### Privacy

- No raw address storage until privacy/security design is complete.
- No raw addresses in logs.
- No public claims that Civic lookup confirms registration.
- Store only references to public records in saved slate.

### Security

- Passwords must be hashed with a modern password hashing scheme if email/password ships.
- Sessions must be HTTP-only and same-site.
- Admin routes must require authenticated `editor` or `admin` role.
- Cookie-authenticated writes must require same-origin CSRF protection.
- Audit admin write actions.

### Data integrity

- Public records must remain source-linked.
- Publish workflow must preserve exact DSA wording.
- Data-quality checks must pass before launch candidate builds.
- Weak confidence labels must remain visible.

### Accessibility

- Forms must have labels and visible focus states.
- Detail and status changes should use polite live regions where appropriate.
- Civic/registration disclaimers must be visible, not hidden behind accordions.

## Private alpha acceptance criteria

- Invited user can sign in.
- Signed-in user can save and remove at least one endorsement.
- Signed-in user can save and remove at least one measure position and one voter guide.
- Signed-in user can view “My Slate.”
- Anonymous user can still browse without an account.
- Civic lookup works with supported-election dropdown.
- Admin/editor can publish chapter and endorsement drafts.
- Admin console no longer depends on a shared local bearer token once auth roles land.
- `make verify` passes.
- `make data-quality` passes.
- Documentation states that registration status is not confirmed by DSA Slate.
- Private-alpha launch runbook covers anonymous browse, Civic lookup, saved slate, editor/admin flow, negative role checks, and cleanup.

## Post-alpha scope

- raw saved addresses
- address encryption at rest
- delete/export account data
- public submissions
- scraper-assisted source discovery
- source registry UI
- Google Places address autocomplete
- campaign operations tools
- launch/deployment hardening

## Recommended next implementation sequence

1. Auth/session foundation.
2. Role-gated admin console.
3. Saved slate data model/API.
4. Saved slate UI.
5. Private alpha polish pass.
6. Privacy/security design for opt-in saved addresses.
