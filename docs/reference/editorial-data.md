# Editorial Data Reference

DSA Slate starts with manually reviewed, source-linked seed data. The first source-linked MVP uses the California cluster and checked-in seed files under `web/src/data/`. The public homepage also includes a methodology/disclaimer slice that explains these source rules to users.

## Required fields for public records

Every public endorsement, recommendation, voter guide, and measure position must include:

- original source URL
- endorsing body or chapter
- exact endorsement/position type
- confidence label
- election or jurisdiction context when known

Public UI copy keeps records source-linked and manual, rather than implying automated verification.

## Endorsement wording rule

Do not collapse `recommended` into `endorsed`. If a chapter guide says “recommend,” publish `recommended`. If a chapter guide says “endorse,” publish `endorsed`.

## Confidence labels

- `official_source` — recorded from an official DSA national, state, or chapter website.
- `official_social_source` — recorded from a verified DSA body social account.
- `campaign_source` — recorded from a candidate or campaign claim; manual review required.
- `press_source` — recorded from a press report; manual review required.
- `needs_review` — unclear, incomplete, or unverified source.

## First-slice privacy rule

The source-linked MVP does not store raw voter addresses. Search is static text search or browsing only. Future ZIP/address lookup requires a separate privacy review before implementation.

## Seed review checklist

- Confirm the source page is public.
- Confirm the source is an official DSA national, state, or chapter source when using `official_source`.
- Preserve exact endorsement wording.
- Add source date when visible.
- Add notes for ambiguity rather than upgrading confidence.
- Run `make verify` before merging seed changes.

## Manual add/update workflow

Until a reviewed admin API exists, all public records are added through reviewed seed changes:

1. Capture the public source URL before editing data.
2. Confirm the endorsing body, chapter, election, race, candidate, or measure context.
3. Preserve exact wording: `recommended` stays `recommended`; do not upgrade it to `endorsed`.
4. Use `needs_review` when source authority, date, or wording is unclear.
5. Update both checked-in seed surfaces when needed: `seeds/california.sql` for Postgres and `web/src/data/california.ts` for static fallback.
6. Run `make verify`.
7. Run `make data-quality` when local Docker/Postgres is available.

The data-quality command is a metadata check, not a truth oracle. Human review still decides whether a source is official enough and whether a record should be published.

## Chapter directory import

The national chapter directory was imported once from a CSV snapshot into `seeds/chapters.sql`. Do not add runtime CSV parsing. Future chapter updates should edit database seed rows or a reviewed admin workflow so Postgres remains the source the API reads.

## Editorial review metadata

Source-bearing public records may include:

- `reviewedAt` — the date an editor manually reviewed the source, formatted `YYYY-MM-DD`.
- `reviewNotes` — a short explanation of what was checked or why source metadata is absent.

Use review metadata when a public source is valid but lacks a visible publication/source date. Do not invent `publishedAt` or `sourceDate`. A documented review can suppress missing-date data-quality warnings, but it does not prove the record is true or current.

## Automated seed validation

`web/src/data/schema.ts` enforces local metadata integrity for checked-in seed data.

The validator rejects:

- duplicate IDs within any collection
- missing `sourceUrl` on voter guides, endorsements, and measure positions
- voter guides without `chapterId` context
- non-HTTP(S) source, website, official, campaign, or photo URLs
- non-ISO dates; use `YYYY-MM-DD`
- references to missing chapters, elections, races, candidates, voter guides, or ballot measures

The validator does not prove that an endorsement is true, current, or legally usable. Human review still must confirm that the source page is public, official enough for its confidence label, and quoted without upgrading “recommended” to “endorsed”.

## Backend data-quality report

`make data-quality` seeds local Postgres and runs `backend/cmd/dataquality` against the read model.

The report fails on missing source URLs for public voter-guide, endorsement, and measure-position records. It warns on stale dates, missing dates, weak confidence labels, and unknown public positions. A documented editorial review can suppress missing-date warnings for source-bearing records when the visible source date is absent. Warnings are allowed during early seed expansion, but they should be resolved or explained before an admin/write API is introduced.
