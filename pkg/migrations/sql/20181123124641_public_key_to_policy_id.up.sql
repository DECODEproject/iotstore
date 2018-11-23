ALTER TABLE events
  RENAME COLUMN public_key TO policy_id;

ALTER INDEX events_public_key_idx RENAME TO events_policy_id_idx;