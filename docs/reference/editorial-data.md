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
