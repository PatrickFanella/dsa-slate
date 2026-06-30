CREATE TABLE IF NOT EXISTS saved_slate_items (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  endorsement_id text NOT NULL REFERENCES endorsements(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, endorsement_id)
);

CREATE INDEX IF NOT EXISTS saved_slate_items_user_id_idx ON saved_slate_items(user_id);
CREATE INDEX IF NOT EXISTS saved_slate_items_endorsement_id_idx ON saved_slate_items(endorsement_id);
