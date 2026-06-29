CREATE TABLE IF NOT EXISTS chapters (
  id text PRIMARY KEY,
  name text NOT NULL,
  state text NOT NULL,
  website_url text,
  official_dsa_url text,
  coverage_notes text,
  status text,
  group_email text,
  linktree_url text,
  facebook_url text,
  twitter_url text,
  bluesky_url text,
  instagram_url text,
  blog_url text,
  event_calendar_url text,
  email_list_url text,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS elections (
  id text PRIMARY KEY,
  name text NOT NULL,
  election_date date,
  election_type text,
  jurisdiction text,
  official_url text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS races (
  id text PRIMARY KEY,
  election_id text NOT NULL REFERENCES elections(id),
  office text NOT NULL,
  district text,
  level text NOT NULL,
  jurisdiction text NOT NULL
);

CREATE TABLE IF NOT EXISTS candidates (
  id text PRIMARY KEY,
  name text NOT NULL,
  campaign_url text,
  party text,
  incumbent boolean,
  photo_url text
);

CREATE TABLE IF NOT EXISTS voter_guides (
  id text PRIMARY KEY,
  title text NOT NULL,
  chapter_id text NOT NULL REFERENCES chapters(id),
  election_id text REFERENCES elections(id),
  source_url text NOT NULL,
  confidence text NOT NULL DEFAULT 'needs_review',
  published_at date,
  summary text,
  reviewed_at date,
  review_notes text,
  CONSTRAINT voter_guides_confidence_check CHECK (
    confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review')
  )
);

ALTER TABLE voter_guides
  ADD COLUMN IF NOT EXISTS confidence text NOT NULL DEFAULT 'needs_review';

CREATE TABLE IF NOT EXISTS endorsements (
  id text PRIMARY KEY,
  candidate_id text NOT NULL REFERENCES candidates(id),
  race_id text NOT NULL REFERENCES races(id),
  chapter_id text NOT NULL REFERENCES chapters(id),
  voter_guide_id text REFERENCES voter_guides(id),
  endorsement_type text NOT NULL,
  source_url text NOT NULL,
  source_date date,
  confidence text NOT NULL DEFAULT 'needs_review',
  notes text,
  reviewed_at date,
  review_notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT endorsements_type_check CHECK (
    endorsement_type IN ('endorsed', 'recommended', 'opposed', 'no_recommendation', 'write_in', 'coalition', 'unknown')
  ),
  CONSTRAINT endorsements_confidence_check CHECK (
    confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review')
  )
);

CREATE TABLE IF NOT EXISTS ballot_measures (
  id text PRIMARY KEY,
  election_id text NOT NULL REFERENCES elections(id),
  title text NOT NULL,
  jurisdiction text NOT NULL,
  measure_number text,
  summary text
);

CREATE TABLE IF NOT EXISTS measure_positions (
  id text PRIMARY KEY,
  measure_id text NOT NULL REFERENCES ballot_measures(id),
  chapter_id text NOT NULL REFERENCES chapters(id),
  position text NOT NULL,
  source_url text NOT NULL,
  confidence text NOT NULL DEFAULT 'needs_review',
  reviewed_at date,
  review_notes text,
  CONSTRAINT measure_positions_position_check CHECK (position IN ('yes', 'no', 'no_recommendation', 'opposed', 'unknown')),
  CONSTRAINT measure_positions_confidence_check CHECK (
    confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review')
  )
);

ALTER TABLE chapters
  ALTER COLUMN state SET NOT NULL;

ALTER TABLE chapters
  ADD COLUMN IF NOT EXISTS status text,
  ADD COLUMN IF NOT EXISTS group_email text,
  ADD COLUMN IF NOT EXISTS linktree_url text,
  ADD COLUMN IF NOT EXISTS facebook_url text,
  ADD COLUMN IF NOT EXISTS twitter_url text,
  ADD COLUMN IF NOT EXISTS bluesky_url text,
  ADD COLUMN IF NOT EXISTS instagram_url text,
  ADD COLUMN IF NOT EXISTS blog_url text,
  ADD COLUMN IF NOT EXISTS event_calendar_url text,
  ADD COLUMN IF NOT EXISTS email_list_url text;

ALTER TABLE races
  ALTER COLUMN election_id SET NOT NULL,
  ALTER COLUMN level SET NOT NULL,
  ALTER COLUMN jurisdiction SET NOT NULL;

ALTER TABLE voter_guides
  ALTER COLUMN chapter_id SET NOT NULL;

ALTER TABLE voter_guides
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;

ALTER TABLE endorsements
  ALTER COLUMN candidate_id SET NOT NULL,
  ALTER COLUMN race_id SET NOT NULL,
  ALTER COLUMN chapter_id SET NOT NULL;

ALTER TABLE endorsements
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;

ALTER TABLE ballot_measures
  ALTER COLUMN election_id SET NOT NULL,
  ALTER COLUMN jurisdiction SET NOT NULL;

ALTER TABLE measure_positions
  ALTER COLUMN measure_id SET NOT NULL,
  ALTER COLUMN chapter_id SET NOT NULL;

ALTER TABLE measure_positions
  ADD COLUMN IF NOT EXISTS reviewed_at date,
  ADD COLUMN IF NOT EXISTS review_notes text;
