DELETE FROM learning_outbox WHERE status = 'failed';

ALTER TABLE learning_outbox
    DROP CONSTRAINT learning_outbox_failed_at_check,
    DROP COLUMN failed_at,
    DROP CONSTRAINT learning_outbox_status_check;

ALTER TABLE learning_outbox
    ADD CONSTRAINT learning_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'completed'));
