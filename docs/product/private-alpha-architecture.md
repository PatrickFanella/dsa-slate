# Private Alpha Architecture Plan

## Architecture goal

Evolve DSA Slate from a public source-linked directory into a private-alpha voter guide app with accounts, role-gated editorial workflows, saved slates, and privacy-safe Civic lookup.

The architecture should preserve the current rule: no raw address storage until there is a dedicated privacy/security design.

## Existing system

### Backend

- Go HTTP API.
- PostgreSQL via pgx.
- Public read tables for directory data.
- Admin draft tables for chapter and endorsement drafts.
- Data-quality command over public read model.
- Google Civic client for elections and anonymous voter-info lookup.

### Frontend

- Vite React TypeScript.
- Public directory homepage.
- Detail panel for source metadata.
- Admin draft console at `/admin`.
- Civic lookup panel with supported-election dropdown.

### Operations

- Docker Compose for Postgres/API/web.
- `.env` for local secrets.
- `make verify` as primary verification.
- `make data-quality` for electoral data validation.

## Target private-alpha modules

### Auth module

New package:

- `backend/internal/auth`

Responsibilities:

- user records
- password hash or external identity records
- session creation/validation
- role checks
- CSRF token support for cookie-authenticated writes

Private-alpha slice 1 implemented:

- `users` and `user_sessions` schema in `migrations/0001_init.sql`
- `backend/internal/auth` package with role/session DTOs, session token generation, and Postgres helpers
- `GET /api/auth/session` anonymous/current-session read path
- Google OIDC sign-in slice with `user_oauth_identities`, state/nonce/PKCE, verified ID tokens, local sessions, and logout
- existing `editor`/`admin` accounts require explicit OAuth identity links before Google sign-in can use those privileged roles

### User/profile module

New package:

- `backend/internal/users`

Responsibilities:

- user profile read/update
- non-sensitive preferences
- preferred chapter/state/election

Do not add raw address fields in the first private-alpha implementation.

### Saved slate module

New package:

- `backend/internal/slate`

Responsibilities:

- save public record references for a user
- list saved records joined with public read data
- remove saved records
- group records by election/chapter/type

Saved slate should reference public IDs:

- `endorsement_id`
- `measure_position_id`
- `voter_guide_id`

It should not duplicate source fields.

### Admin/editorial module

Existing package:

- `backend/internal/admin`

Private-alpha changes:

- replace bearer-token access with authenticated role checks
- keep draft/publish behavior
- keep audit events
- add actor as authenticated user ID/email instead of static `admin`

### Civic module

Existing package:

- `backend/internal/civic`

Private-alpha changes:

- keep anonymous lookup
- optionally associate derived non-sensitive election/district IDs with a saved slate later
- do not store raw submitted address

## Database roadmap

### Auth/session tables

Proposed first tables:

```sql
CREATE TABLE users (
  id text PRIMARY KEY,
  email text NOT NULL UNIQUE,
  display_name text,
  role text NOT NULL DEFAULT 'user',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT users_role_check CHECK (role IN ('user', 'editor', 'admin'))
);

CREATE TABLE user_password_credentials (
  user_id text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_sessions (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

If Google sign-in is added:

```sql
CREATE TABLE user_oauth_identities (
  provider text NOT NULL,
  provider_subject text NOT NULL,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (provider, provider_subject)
);
```

### Profile preferences

```sql
CREATE TABLE user_preferences (
  user_id text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  preferred_chapter_id text REFERENCES chapters(id),
  preferred_state text,
  preferred_election_id text REFERENCES elections(id),
  theme text,
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### Saved slate

```sql
CREATE TABLE saved_slate_items (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  record_type text NOT NULL,
  record_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT saved_slate_items_type_check CHECK (record_type IN ('endorsement', 'measure_position', 'voter_guide')),
  UNIQUE (user_id, record_type, record_id)
);
```

No raw address table is part of private-alpha foundation.

## API roadmap

### Auth endpoints

- `GET /api/auth/google/start`
- `GET /api/auth/google/callback`
- `POST /api/auth/logout`
- `GET /api/auth/session`

### User endpoints

- `GET /api/me`
- `PUT /api/me/preferences`

### Saved slate endpoints

- `GET /api/me/slate`
- `POST /api/me/slate/items`
- `DELETE /api/me/slate/items/{id}`

### Admin endpoints

Existing `/api/admin/...` endpoints remain, but auth changes:

- remove shared bearer-token dependency
- require authenticated `editor` or `admin`
- audit actor from session user

### Civic endpoints

Existing endpoints remain anonymous:

- `GET /api/civic/elections`
- `POST /api/civic/lookup`

Future saved-profile endpoint must not be added until address privacy design exists.

## Frontend roadmap

### App shell

Before adding many more routes, add a minimal route switch for:

- `/` public voter guide
- `/admin` editor console
- `/login` auth
- `/me/slate` saved slate

Router dependency is optional; current pathname switch can work for one more slice, but a router becomes useful once auth pages land.

### Auth UI

- login/register form
- session status
- logout button
- role-aware navigation

### Saved slate UI

- save button on detail panel
- saved state indicator on cards/details
- “My Slate” page
- remove saved item
- grouped print-friendly slate view

### Admin UI

- replace token input with session-aware access
- show current role/user
- keep existing draft forms

### Civic UI

- keep no-storage copy visible
- keep supported-election dropdown
- do not add address autocomplete until Google Places/privacy review is approved

## Security plan

### Session security

- HTTP-only session cookie.
- SameSite=Lax or Strict.
- Secure flag outside local development.
- Session expiration.
- Logout invalidates session.

### CSRF

Cookie-authenticated writes require CSRF protection.

Protected write groups:

- `/api/admin/...`
- `/api/me/...`
- auth logout if cookie based

### Passwords

If email/password ships:

- use a modern password hashing package
- never store plaintext passwords
- add password reset only as a separate slice

### Address privacy

Before storing raw addresses, create a separate threat model covering:

- encryption key management
- log redaction
- database backups
- admin support access
- deletion/export
- retention
- breach response

## Implementation sequence

### Slice 1: Auth schema and session foundation

- users table
- sessions table
- auth package
- session middleware
- `GET /api/auth/session`
- tests

### Slice 2: First sign-in method

- choose email/password or magic link
- login/register/logout endpoints
- minimal login UI
- tests

### Slice 3: Role-gated admin

- replace admin token UI
- protect admin API by role
- update audit actor
- tests

### Slice 4: Saved slate backend

- saved slate table
- API endpoints
- join saved records against public directory data
- tests

### Slice 5: Saved slate UI

- save/remove buttons
- My Slate page
- empty states
- tests

### Slice 6: Private alpha polish

- navigation
- copy review
- runbook
- launch checklist

### Slice 7: Address privacy design

- no code that stores raw addresses until this is complete

## Architecture decisions to make before Slice 1

1. First auth method: email/password, magic link, or Google sign-in.
2. Session storage: DB-backed opaque session ID is recommended.
3. Router: keep manual pathname switch for one more slice or add React Router before auth pages.
4. User IDs: generated string IDs, not serial integers, are recommended.
5. Password hash package if email/password is selected.

## Verification expectations

Every private-alpha slice must run:

```bash
make verify
```

Data-affecting slices must also run:

```bash
make data-quality
```

Auth/security slices must include negative tests for unauthorized and wrong-role access.
