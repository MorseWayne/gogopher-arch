ALTER TABLE learning_outbox
    DROP CONSTRAINT learning_outbox_status_check;

ALTER TABLE learning_outbox
    ADD CONSTRAINT learning_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    ADD COLUMN failed_at timestamptz,
    ADD CONSTRAINT learning_outbox_failed_at_check
        CHECK ((status = 'failed') = (failed_at IS NOT NULL));
