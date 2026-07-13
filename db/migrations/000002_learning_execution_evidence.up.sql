ALTER TABLE learning_attempts
    ADD CONSTRAINT learning_attempts_id_learner_unique UNIQUE (id, learner_id);

CREATE TABLE attempt_submissions (
    id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL UNIQUE,
    learner_id uuid NOT NULL,
    submission_key text NOT NULL CHECK (length(submission_key) BETWEEN 1 AND 200),
    request_fingerprint text NOT NULL CHECK (request_fingerprint ~ '^[0-9a-f]{64}$'),
    workspace jsonb NOT NULL CHECK (jsonb_typeof(workspace) = 'object'),
    workspace_revision bigint NOT NULL CHECK (workspace_revision >= 0),
    workspace_hash text NOT NULL CHECK (workspace_hash ~ '^[0-9a-f]{64}$'),
    rule_set_hash text NOT NULL CHECK (rule_set_hash ~ '^[0-9a-f]{64}$'),
    assistance_cutoff_seq bigint NOT NULL DEFAULT 0 CHECK (assistance_cutoff_seq >= 0),
    status text NOT NULL CHECK (status IN ('frozen', 'executing', 'evaluated', 'infra_failed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    evaluated_at timestamptz,
    CHECK ((status = 'evaluated') = (evaluated_at IS NOT NULL)),
    UNIQUE (id, attempt_id),
    FOREIGN KEY (attempt_id, learner_id) REFERENCES learning_attempts(id, learner_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX attempt_submissions_attempt_key_idx
    ON attempt_submissions (attempt_id, submission_key);

CREATE TABLE attempt_executions (
    id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL REFERENCES learning_attempts(id) ON DELETE RESTRICT,
    submission_id uuid,
    action text NOT NULL CHECK (action IN ('build', 'test', 'vet', 'submit')),
    sequence integer NOT NULL DEFAULT 0 CHECK (sequence >= 0),
    request_key text NOT NULL CHECK (length(request_key) BETWEEN 1 AND 200),
    request_fingerprint text NOT NULL CHECK (request_fingerprint ~ '^[0-9a-f]{64}$'),
    release_id text NOT NULL REFERENCES definition_releases(release_id) ON DELETE RESTRICT,
    task_id text NOT NULL,
    task_version integer NOT NULL CHECK (task_version > 0),
    task_hash text NOT NULL CHECK (task_hash ~ '^[0-9a-f]{64}$'),
    workspace_revision bigint NOT NULL CHECK (workspace_revision >= 0),
    workspace_hash text NOT NULL CHECK (workspace_hash ~ '^[0-9a-f]{64}$'),
    spec jsonb NOT NULL CHECK (jsonb_typeof(spec) = 'object'),
    status text NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'user_failed', 'infra_failed')),
    result jsonb,
    claim_count integer NOT NULL DEFAULT 0 CHECK (claim_count >= 0),
    lease_owner text,
    lease_expires_at timestamptz,
    lease_heartbeat_at timestamptz,
    started_at timestamptz,
    finished_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((action = 'submit') = (submission_id IS NOT NULL)),
    CHECK ((submission_id IS NULL AND sequence = 0) OR submission_id IS NOT NULL),
    CHECK (
        (status = 'running' AND lease_owner IS NOT NULL AND lease_expires_at IS NOT NULL AND lease_heartbeat_at IS NOT NULL)
        OR
        (status <> 'running' AND lease_owner IS NULL AND lease_expires_at IS NULL AND lease_heartbeat_at IS NULL)
    ),
    CHECK ((status IN ('succeeded', 'user_failed', 'infra_failed')) = (result IS NOT NULL AND finished_at IS NOT NULL)),
    CHECK (started_at IS NULL OR started_at >= created_at),
    CHECK (finished_at IS NULL OR finished_at >= created_at),
    CHECK (updated_at >= created_at),
    UNIQUE (id, submission_id),
    FOREIGN KEY (submission_id, attempt_id) REFERENCES attempt_submissions(id, attempt_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX attempt_executions_normal_request_idx
    ON attempt_executions (attempt_id, request_key)
    WHERE submission_id IS NULL;
CREATE UNIQUE INDEX attempt_executions_submit_sequence_idx
    ON attempt_executions (submission_id, sequence)
    WHERE submission_id IS NOT NULL;
CREATE INDEX attempt_executions_queue_idx
    ON attempt_executions (created_at, id)
    WHERE status = 'queued';
CREATE INDEX attempt_executions_expired_lease_idx
    ON attempt_executions (lease_expires_at, id)
    WHERE status = 'running';

CREATE TABLE assistance_events (
    id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL,
    learner_id uuid NOT NULL,
    event_key text NOT NULL CHECK (length(event_key) BETWEEN 1 AND 200),
    event_seq bigint NOT NULL CHECK (event_seq > 0),
    event_type text NOT NULL CHECK (event_type IN ('hint_revealed', 'reference_opened', 'solution_viewed', 'ai_declared')),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(payload) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (attempt_id, event_key),
    UNIQUE (attempt_id, event_seq),
    FOREIGN KEY (attempt_id, learner_id) REFERENCES learning_attempts(id, learner_id) ON DELETE RESTRICT
);

CREATE INDEX assistance_events_attempt_idx
    ON assistance_events (attempt_id, event_seq);

CREATE TABLE artifacts (
    id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL REFERENCES learning_attempts(id) ON DELETE RESTRICT,
    submission_id uuid,
    kind text NOT NULL CHECK (kind IN ('workspace', 'diff', 'explanation', 'test_report')),
    content jsonb NOT NULL,
    content_bytes integer NOT NULL CHECK (content_bytes >= 0 AND content_bytes <= 4194304),
    content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (submission_id, attempt_id) REFERENCES attempt_submissions(id, attempt_id) ON DELETE RESTRICT
);

CREATE INDEX artifacts_attempt_idx ON artifacts (attempt_id, created_at);

CREATE TABLE evaluation_batches (
    id uuid PRIMARY KEY,
    submission_id uuid NOT NULL REFERENCES attempt_submissions(id) ON DELETE RESTRICT,
    execution_id uuid NOT NULL,
    rule_set_hash text NOT NULL CHECK (rule_set_hash ~ '^[0-9a-f]{64}$'),
    rule_results jsonb NOT NULL CHECK (jsonb_typeof(rule_results) = 'array'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (submission_id, rule_set_hash),
    FOREIGN KEY (execution_id, submission_id) REFERENCES attempt_executions(id, submission_id) ON DELETE RESTRICT
);

CREATE TABLE evidence_records (
    id uuid PRIMARY KEY,
    evaluation_batch_id uuid NOT NULL REFERENCES evaluation_batches(id) ON DELETE RESTRICT,
    learner_id uuid NOT NULL,
    capability_id text NOT NULL,
    capability_version integer NOT NULL CHECK (capability_version > 0),
    capability_hash text NOT NULL CHECK (capability_hash ~ '^[0-9a-f]{64}$'),
    attempt_id uuid NOT NULL REFERENCES learning_attempts(id) ON DELETE RESTRICT,
    activity_id text NOT NULL,
    artifact_id uuid REFERENCES artifacts(id) ON DELETE RESTRICT,
    evidence_rule_id text NOT NULL,
    evidence_type text NOT NULL CHECK (evidence_type IN ('implement', 'test', 'diagnose')),
    result text NOT NULL CHECK (result IN ('passed', 'failed')),
    independence text NOT NULL CHECK (independence IN ('guided', 'hinted', 'referenced', 'ai_assisted', 'independent')),
    context_level text NOT NULL CHECK (context_level IN ('same_context', 'variant')),
    evaluator text NOT NULL CHECK (evaluator = 'deterministic'),
    rule_version integer NOT NULL CHECK (rule_version > 0),
    reason text NOT NULL,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (evaluation_batch_id, capability_id, capability_version, evidence_rule_id, evidence_type),
    FOREIGN KEY (attempt_id, learner_id) REFERENCES learning_attempts(id, learner_id) ON DELETE RESTRICT
);

CREATE INDEX evidence_records_learner_capability_idx
    ON evidence_records (learner_id, capability_id, capability_version, occurred_at DESC);

CREATE TABLE learning_outbox (
    id uuid PRIMARY KEY,
    topic text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    idempotency_key text NOT NULL UNIQUE,
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed')),
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    CHECK (
        (status = 'processing' AND lease_owner IS NOT NULL AND lease_expires_at IS NOT NULL)
        OR
        (status <> 'processing' AND lease_owner IS NULL AND lease_expires_at IS NULL)
    ),
    CHECK ((status = 'completed') = (completed_at IS NOT NULL))
);

CREATE INDEX learning_outbox_pending_idx
    ON learning_outbox (available_at, created_at)
    WHERE status = 'pending';
