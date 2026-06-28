# Editorial Data Reference

DSA Slate starts with manually reviewed, source-linked seed data. The first trusted-directory MVP uses the California cluster and checked-in seed files under `web/src/data/`.

## Required fields for public records

Every public endorsement, recommendation, voter guide, and measure position must include:

- original source URL
- endorsing body or chapter
- exact endorsement/position type
- confidence label
- election or jurisdiction context when known

## Endorsement wording rule

Do not collapse `recommended` into `endorsed`. If a chapter guide says “recommend,” publish `recommended`. If a chapter guide says “endorse,” publish `endorsed`.

## Confidence labels

- `official_source` — DSA national, state, or chapter website.
- `official_social_source` — verified DSA body social account.
- `campaign_source` — candidate claims endorsement; needs confirmation.
- `press_source` — news report; needs official confirmation where possible.
- `needs_review` — unclear, incomplete, or unverified source.

## First-slice privacy rule

The trusted-directory MVP does not store voter addresses. Search is static text search or browsing only. Future ZIP/address lookup requires a separate privacy review before implementation.

## Seed review checklist

- Confirm the source page is public.
- Confirm the source is an official DSA national, state, or chapter source when using `official_source`.
- Preserve exact endorsement wording.
- Add source date when visible.
- Add notes for ambiguity rather than upgrading confidence.
- Run `make verify` before merging seed changes.

## Automated seed validation

`web/src/data/schema.ts` enforces local metadata integrity for checked-in seed data.

The validator rejects:

- duplicate IDs within any collection
- missing `sourceUrl` on voter guides, endorsements, and measure positions
- voter guides without `chapterId` context
- non-HTTP(S) source, website, official, campaign, or photo URLs
- non-ISO dates; use `YYYY-MM-DD`
- references to missing chapters, elections, races, candidates, voter guides, or ballot measures

The validator does not prove that an endorsement is true, current, or legally usable. Human review still must confirm that the source page is public, official enough for its confidence label, and quoted without upgrading “recommended” to “endorsed.”
