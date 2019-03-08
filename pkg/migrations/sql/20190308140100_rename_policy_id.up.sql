ALTER TABLE events
  RENAME COLUMN policy_id TO community_id;

ALTER INDEX events_policy_id_idx RENAME TO events_community_id_idx;