ALTER TABLE events
  ADD COLUMN device_token TEXT NOT NULL;

CREATE INDEX IF NOT EXISTS events_device_token_idx
  ON events (device_token);