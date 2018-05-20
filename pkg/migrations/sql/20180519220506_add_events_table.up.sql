CREATE TABLE IF NOT EXISTS events (
  id SERIAL PRIMARY KEY,
  public_key TEXT NOT NULL,
  user_id TEXT NOT NULL,
  recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  data JSONB
);

CREATE INDEX IF NOT EXISTS events_public_key_index
  ON events (public_key);

CREATE INDEX IF NOT EXISTS events_user_id_index
  ON events (user_id);

CREATE INDEX IF NOT EXISTS events_recorded_at
  ON events (recorded_at);