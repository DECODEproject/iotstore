DROP INDEX IF EXISTS events_public_key_idx CASCADE;
DROP INDEX IF EXISTS events_recorded_at_user_uid_idx CASCADE;
DROP INDEX IF EXISTS events_user_uid_idx CASCADE;

DROP TABLE IF EXISTS events CASCADE;