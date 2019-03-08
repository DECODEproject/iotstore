ALTER TABLE events
  RENAME COLUMN community_id TO policy_id;

ALTER INDEX events_community_id_idx RENAME TO events_policy_id_idx;