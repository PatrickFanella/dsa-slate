CREATE TABLE IF NOT EXISTS chapters (
  id text PRIMARY KEY,
  name text NOT NULL,
  state text,
  website_url text,
  official_dsa_url text,
  coverage_notes text,
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
  election_id text REFERENCES elections(id),
  office text NOT NULL,
  district text,
  level text,
  jurisdiction text
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
  chapter_id text REFERENCES chapters(id),
  election_id text REFERENCES elections(id),
  source_url text NOT NULL,
  confidence text NOT NULL DEFAULT 'needs_review',
  published_at date,
  summary text,
  CONSTRAINT voter_guides_confidence_check CHECK (
    confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review')
  )
);

ALTER TABLE voter_guides
  ADD COLUMN IF NOT EXISTS confidence text NOT NULL DEFAULT 'needs_review';

CREATE TABLE IF NOT EXISTS endorsements (
  id text PRIMARY KEY,
  candidate_id text REFERENCES candidates(id),
  race_id text REFERENCES races(id),
  chapter_id text REFERENCES chapters(id),
  voter_guide_id text REFERENCES voter_guides(id),
  endorsement_type text NOT NULL,
  source_url text NOT NULL,
  source_date date,
  confidence text NOT NULL DEFAULT 'needs_review',
  notes text,
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
  election_id text REFERENCES elections(id),
  title text NOT NULL,
  jurisdiction text,
  measure_number text,
  summary text
);

CREATE TABLE IF NOT EXISTS measure_positions (
  id text PRIMARY KEY,
  measure_id text REFERENCES ballot_measures(id),
  chapter_id text REFERENCES chapters(id),
  position text NOT NULL,
  source_url text NOT NULL,
  confidence text NOT NULL DEFAULT 'needs_review',
  CONSTRAINT measure_positions_position_check CHECK (position IN ('yes', 'no', 'no_recommendation', 'opposed', 'unknown')),
  CONSTRAINT measure_positions_confidence_check CHECK (
    confidence IN ('official_source', 'official_social_source', 'campaign_source', 'press_source', 'needs_review')
  )
);
