ALTER TABLE saved_slate_items
  ADD COLUMN IF NOT EXISTS record_type text,
  ADD COLUMN IF NOT EXISTS record_id text;

UPDATE saved_slate_items
SET record_type = 'endorsement',
    record_id = endorsement_id
WHERE record_type IS NULL
  AND record_id IS NULL
  AND endorsement_id IS NOT NULL;

ALTER TABLE saved_slate_items
  ALTER COLUMN record_type SET NOT NULL,
  ALTER COLUMN record_id SET NOT NULL;

ALTER TABLE saved_slate_items
  DROP CONSTRAINT IF EXISTS saved_slate_items_endorsement_id_fkey,
  DROP CONSTRAINT IF EXISTS saved_slate_items_user_id_endorsement_id_key;

DROP INDEX IF EXISTS saved_slate_items_endorsement_id_idx;

ALTER TABLE saved_slate_items
  DROP COLUMN IF EXISTS endorsement_id;

ALTER TABLE saved_slate_items
  ADD CONSTRAINT saved_slate_items_record_type_check CHECK (record_type IN ('endorsement', 'measure_position', 'voter_guide')),
  ADD CONSTRAINT saved_slate_items_user_record_key UNIQUE (user_id, record_type, record_id);

CREATE INDEX IF NOT EXISTS saved_slate_items_record_idx ON saved_slate_items(record_type, record_id);
