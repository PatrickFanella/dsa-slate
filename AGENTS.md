# Agent Instructions

This repository is DSA Slate: a Go full-stack application with Vite React TypeScript, Postgres, and Docker for a source-linked DSA electoral guide.

## Priorities

1. Keep the application runnable and minimal.
2. Preserve Open Pilot issue and PR templates.
3. Keep generated caches, secrets, and local data out of git.
4. Keep endorsement data source-linked with clear confidence labels.
5. Do not store raw voter addresses unless a future feature explicitly adds opt-in handling and a privacy review.
6. Update this file and `README.md` when stack conventions change.

## Verification

Run:

```bash
make verify
```

Use narrower commands only when an issue explicitly asks for a smaller check.
