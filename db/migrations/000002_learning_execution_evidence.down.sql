DROP TABLE IF EXISTS learning_outbox;
DROP TABLE IF EXISTS evidence_records;
DROP TABLE IF EXISTS evaluation_batches;
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS assistance_events;
DROP TABLE IF EXISTS attempt_executions;
DROP TABLE IF EXISTS attempt_submissions;
ALTER TABLE learning_attempts DROP CONSTRAINT IF EXISTS learning_attempts_id_learner_unique;
