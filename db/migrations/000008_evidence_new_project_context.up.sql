ALTER TABLE evidence_records
    DROP CONSTRAINT evidence_records_context_level_check,
    ADD CONSTRAINT evidence_records_context_level_check
        CHECK (context_level IN ('same_context', 'variant', 'new_project'));
