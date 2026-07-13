ALTER TABLE review_items
    DROP CONSTRAINT IF EXISTS review_items_reason_check,
    DROP CONSTRAINT IF EXISTS review_items_predecessor_unique,
    DROP CONSTRAINT IF EXISTS review_items_completed_outcome_check,
    DROP CONSTRAINT IF EXISTS review_items_source_check;

UPDATE review_items SET predecessor_review_item_id = NULL;

DELETE FROM attempt_review_items
WHERE review_item_id IN (
    SELECT id FROM review_items WHERE source_evidence_id IS NULL OR reason = 'remediation_review'
);
DELETE FROM review_items
WHERE source_evidence_id IS NULL OR reason = 'remediation_review';

ALTER TABLE review_items
    ADD CONSTRAINT review_items_reason_check
        CHECK (reason IN ('first_review','maintenance','remediation','review_incomplete')),
    DROP COLUMN IF EXISTS outcome,
    DROP COLUMN IF EXISTS predecessor_review_item_id,
    ALTER COLUMN source_evidence_id SET NOT NULL;
