BEGIN;

INSERT INTO chapters (id, name, state, website_url, official_dsa_url, coverage_notes, active)
VALUES
  (
    'california-dsa',
    'California DSA',
    'CA',
    'https://www.californiadsa.org/',
    NULL,
    'Statewide DSA voter guide and links to local California chapter guides.',
    TRUE
  ),
  (
    'dsa-san-diego',
    'DSA San Diego',
    'CA',
    'https://dsasandiego.org/',
    NULL,
    'San Diego County and local voter guide coverage.',
    TRUE
  ),
  (
    'dsa-los-angeles',
    'DSA-LA',
    'CA',
    'https://dsa-la.org/',
    NULL,
    'Los Angeles area chapter included as a California-cluster seed source.',
    TRUE
  )
ON CONFLICT (id) DO UPDATE
SET
  name = EXCLUDED.name,
  state = EXCLUDED.state,
  website_url = EXCLUDED.website_url,
  official_dsa_url = EXCLUDED.official_dsa_url,
  coverage_notes = EXCLUDED.coverage_notes,
  active = EXCLUDED.active,
  updated_at = now();

INSERT INTO elections (id, name, election_date, election_type, jurisdiction, official_url)
VALUES
  (
    '2026-ca-primary',
    '2026 California Primary',
    DATE '2026-06-02',
    'primary',
    'CA',
    NULL
  )
ON CONFLICT (id) DO UPDATE
SET
  name = EXCLUDED.name,
  election_date = EXCLUDED.election_date,
  election_type = EXCLUDED.election_type,
  jurisdiction = EXCLUDED.jurisdiction,
  official_url = EXCLUDED.official_url;

INSERT INTO races (id, election_id, office, district, level, jurisdiction)
VALUES
  ('ca-48-us-house-2026-primary', '2026-ca-primary', 'U.S. House', 'CA-48', 'federal', 'CA-48'),
  ('ca-50-us-house-2026-primary', '2026-ca-primary', 'U.S. House', 'CA-50', 'federal', 'CA-50')
ON CONFLICT (id) DO UPDATE
SET
  election_id = EXCLUDED.election_id,
  office = EXCLUDED.office,
  district = EXCLUDED.district,
  level = EXCLUDED.level,
  jurisdiction = EXCLUDED.jurisdiction;

INSERT INTO candidates (id, name, campaign_url, party, incumbent, photo_url)
VALUES
  ('corinna-contreras', 'Corinna Contreras', NULL, 'Democrat', NULL, NULL),
  ('aishwarya-sparky-mitra', 'Aishwarya “Sparky” Mitra', NULL, 'Democrat', NULL, NULL)
ON CONFLICT (id) DO UPDATE
SET
  name = EXCLUDED.name,
  campaign_url = EXCLUDED.campaign_url,
  party = EXCLUDED.party,
  incumbent = EXCLUDED.incumbent,
  photo_url = EXCLUDED.photo_url;

INSERT INTO voter_guides (id, title, chapter_id, election_id, source_url, confidence, published_at, summary)
VALUES
  (
    'california-dsa-2026-primary-guide',
    'California DSA 2026 Primary Voter Guide',
    'california-dsa',
    '2026-ca-primary',
    'https://www.californiadsa.org/voterguide',
    'official_source',
    NULL,
    'Statewide guide and local chapter guide hub for California DSA electoral recommendations.'
  ),
  (
    'dsa-san-diego-2026-primary-guide',
    'DSA San Diego''s June 2026 Primary Voter Guide',
    'dsa-san-diego',
    '2026-ca-primary',
    'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
    'official_source',
    DATE '2026-05-05',
    'Local voter guide with federal, statewide, local, and ballot measure recommendations.'
  )
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  chapter_id = EXCLUDED.chapter_id,
  election_id = EXCLUDED.election_id,
  source_url = EXCLUDED.source_url,
  confidence = EXCLUDED.confidence,
  published_at = EXCLUDED.published_at,
  summary = EXCLUDED.summary;

INSERT INTO ballot_measures (id, election_id, title, jurisdiction, measure_number, summary)
VALUES
  (
    'san-diego-measure-a-2026-primary',
    '2026-ca-primary',
    'San Diego Measure A',
    'San Diego, CA',
    'A',
    'Local ballot measure included in DSA San Diego voter guide seed data.'
  )
ON CONFLICT (id) DO UPDATE
SET
  election_id = EXCLUDED.election_id,
  title = EXCLUDED.title,
  jurisdiction = EXCLUDED.jurisdiction,
  measure_number = EXCLUDED.measure_number,
  summary = EXCLUDED.summary;

INSERT INTO endorsements (id, candidate_id, race_id, chapter_id, voter_guide_id, endorsement_type, source_url, source_date, confidence, notes)
VALUES
  (
    'contreras-dsa-sd-2026-primary',
    'corinna-contreras',
    'ca-48-us-house-2026-primary',
    'dsa-san-diego',
    'dsa-san-diego-2026-primary-guide',
    'recommended',
    'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
    DATE '2026-05-05',
    'official_source',
    'Seeded from DSA San Diego voter guide wording; preserve recommendation language.'
  ),
  (
    'mitra-dsa-sd-2026-primary',
    'aishwarya-sparky-mitra',
    'ca-50-us-house-2026-primary',
    'dsa-san-diego',
    'dsa-san-diego-2026-primary-guide',
    'recommended',
    'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
    DATE '2026-05-05',
    'official_source',
    'Seeded from DSA San Diego voter guide wording; preserve recommendation language.'
  )
ON CONFLICT (id) DO UPDATE
SET
  candidate_id = EXCLUDED.candidate_id,
  race_id = EXCLUDED.race_id,
  chapter_id = EXCLUDED.chapter_id,
  voter_guide_id = EXCLUDED.voter_guide_id,
  endorsement_type = EXCLUDED.endorsement_type,
  source_url = EXCLUDED.source_url,
  source_date = EXCLUDED.source_date,
  confidence = EXCLUDED.confidence,
  notes = EXCLUDED.notes,
  updated_at = now();

INSERT INTO measure_positions (id, measure_id, chapter_id, position, source_url, confidence)
VALUES
  (
    'san-diego-measure-a-dsa-sd-yes-2026-primary',
    'san-diego-measure-a-2026-primary',
    'dsa-san-diego',
    'yes',
    'https://dsasandiego.org/dsa-san-diegos-june-2026-primary-voter-guide/',
    'official_source'
  )
ON CONFLICT (id) DO UPDATE
SET
  measure_id = EXCLUDED.measure_id,
  chapter_id = EXCLUDED.chapter_id,
  position = EXCLUDED.position,
  source_url = EXCLUDED.source_url,
  confidence = EXCLUDED.confidence;

COMMIT;
