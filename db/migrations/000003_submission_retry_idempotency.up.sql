CREATE UNIQUE INDEX attempt_executions_submission_request_idx
    ON attempt_executions (submission_id, request_key)
    WHERE submission_id IS NOT NULL;
