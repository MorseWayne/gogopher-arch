ALTER TABLE attempt_submissions
    ADD COLUMN explanation text NOT NULL DEFAULT ''
    CHECK (char_length(explanation) <= 4000);
