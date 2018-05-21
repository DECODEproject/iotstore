CREATE TABLE IF NOT EXISTS events (
  id SERIAL PRIMARY KEY,
  public_key TEXT NOT NULL,
  user_uid TEXT NOT NULL,
  recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  data BYTEA
);

CREATE INDEX IF NOT EXISTS events_public_key_idx
  ON events (public_key);

CREATE INDEX IF NOT EXISTS events_recorded_at_user_uid_idx
  ON events (recorded_at, user_uid);

CREATE INDEX IF NOT EXISTS events_user_uid_idx
  ON events (user_uid);