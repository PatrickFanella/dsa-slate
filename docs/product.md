# DSA Slate Product Requirements

DSA Slate is a public, source-linked election guide for DSA electoral work.

The first public slice pairs the source-linked directory with a methodology/disclaimer section so users can see how records are sourced, labeled, and limited.

## Core promise

> Find every DSA-endorsed candidate, ballot measure, and voter guide — national, state, and local — in one searchable place.

The product answers: “What is the DSA slate where I live, what election is it for, who endorsed it, and where is the original source?”

## MVP goals

- Build a source-linked public directory of DSA endorsements and voter guides.
- Show source URLs, source dates, and confidence labels for every item.
- Distinguish `endorsed` from `recommended` and other endorsement types.
- Explain public methodology, confidence labels, privacy limits, and election-information disclaimers.
- Support national, state, chapter, candidate, race, measure, and voter guide pages.

## Approved first source-linked slice

- Seed the California cluster as static public cards from source-linked seed data.
- Cover chapters, candidate recommendations, ballot measures, and voter guides.
- Preserve source URLs, source dates, and confidence labels in the public UI.
- Render a public methodology/disclaimer slice that states no raw voter addresses are stored and that DSA Slate links to official election information instead of replacing it.
- Keep the MVP browse-only for now; no submission, admin, or lookup workflows.

## MVP non-goals

- No live election-night results.
- No automated nationwide scraping.
- No raw voter address storage.
- No AI-generated voter recommendations.
- No campaign coordination or paid promotion without legal review.
- No copy that implies legal advice, official election administration, or live address lookup.

## Required confidence labels

- `official_source` — recorded from a DSA national, chapter, or state website.
- `official_social_source` — recorded from a verified chapter/body social account.
- `campaign_source` — recorded from a candidate claim; manual review required.
- `press_source` — recorded from a news article; manual review required.
- `needs_review` — unclear or unverified source.

## Initial launch scope

- Homepage
- National endorsements page
- State and chapter browsing
- Candidate and ballot measure cards
- Voter guide links
- Methodology/disclaimer page

Address lookup and results tracking come after the editorial/data workflow is solid.

## Trusted Directory MVP plan decisions

- First launch proves a source-linked directory, not location lookup.
- Initial seed data focuses on the California cluster.
- Editorial workflow is manual seed files plus documented review rules.
- First implementation slice renders static public cards from seed data.
- Address lookup, admin UI, CMS integration, public submission forms, and results tracking are later phases.
