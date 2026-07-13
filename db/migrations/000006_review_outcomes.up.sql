ALTER TABLE review_items
    ALTER COLUMN source_evidence_id DROP NOT NULL,
    ADD COLUMN predecessor_review_item_id uuid REFERENCES review_items(id) ON DELETE RESTRICT,
    ADD COLUMN outcome text CHECK (outcome IN ('passed','failed','incomplete'));

UPDATE review_items r
SET outcome = CASE
    WHEN EXISTS (
        SELECT 1 FROM evidence_records e
        WHERE e.evaluation_batch_id = r.evaluation_batch_id
          AND e.capability_id = r.capability_id
          AND e.capability_version = r.capability_version
          AND e.result = 'failed'
    ) THEN 'failed'
    WHEN EXISTS (
        SELECT 1 FROM evidence_records e
        WHERE e.evaluation_batch_id = r.evaluation_batch_id
          AND e.capability_id = r.capability_id
          AND e.capability_version = r.capability_version
    ) THEN 'passed'
    ELSE 'incomplete'
END
WHERE r.status = 'completed';

ALTER TABLE review_items
    ADD CONSTRAINT review_items_source_check
        CHECK (source_evidence_id IS NOT NULL OR predecessor_review_item_id IS NOT NULL),
    ADD CONSTRAINT review_items_completed_outcome_check
        CHECK (status <> 'completed' OR outcome IS NOT NULL),
    ADD CONSTRAINT review_items_predecessor_unique UNIQUE (predecessor_review_item_id),
    DROP CONSTRAINT review_items_reason_check,
    ADD CONSTRAINT review_items_reason_check
        CHECK (reason IN ('first_review','maintenance','remediation','remediation_review','review_incomplete'));
