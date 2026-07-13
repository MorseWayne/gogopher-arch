DROP INDEX IF EXISTS learning_outbox_topic_status_available_idx;
ALTER TABLE learning_outbox
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS consumer_version,
    DROP COLUMN IF EXISTS consumer;
DROP TABLE IF EXISTS attempt_review_items;
DROP TABLE IF EXISTS review_items;
DROP TABLE IF EXISTS capability_snapshots;
