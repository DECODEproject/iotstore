ALTER TABLE events
  RENAME COLUMN policy_id TO public_key;

ALTER INDEX events_policy_id_idx RENAME TO events_public_key_idx;