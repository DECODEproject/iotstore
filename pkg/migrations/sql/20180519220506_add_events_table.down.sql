DROP INDEX IF EXISTS events_public_key_index CASCADE;
DROP INDEX IF EXISTS events_user_id_index CASCADE;
DROP INDEX IF EXISTS events_recorded_at_index CASCADE;

DROP TABLE IF EXISTS events CASCADE;