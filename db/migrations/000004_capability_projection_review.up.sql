CREATE TABLE capability_snapshots (
    learner_id uuid NOT NULL REFERENCES learners(id) ON DELETE RESTRICT,
    capability_id text NOT NULL,
    capability_version integer NOT NULL CHECK (capability_version > 0),
    capability_hash text NOT NULL CHECK (capability_hash ~ '^[0-9a-f]{64}$'),
    acquisition_state text NOT NULL CHECK (acquisition_state IN ('not_started','exploring','practiced','verified','stable')),
    independence_state text NOT NULL CHECK (independence_state IN ('unverified','guided','ai_assisted','hinted','referenced','independent')),
    transfer_state text NOT NULL CHECK (transfer_state IN ('unverified','same_context','variant','new_project')),
    retention_base_state text NOT NULL CHECK (retention_base_state IN ('fresh','rusty')),
    last_evidence_at timestamptz,
    last_independent_at timestamptz,
    next_review_at timestamptz,
    projection_version integer NOT NULL CHECK (projection_version > 0),
    projected_at timestamptz NOT NULL,
    PRIMARY KEY (learner_id, capability_id, capability_version)
);

CREATE INDEX capability_snapshots_next_review_idx
    ON capability_snapshots (next_review_at, learner_id)
    WHERE next_review_at IS NOT NULL;

CREATE TABLE review_items (
    id uuid PRIMARY KEY,
    learner_id uuid NOT NULL REFERENCES learners(id) ON DELETE RESTRICT,
    capability_id text NOT NULL,
    capability_version integer NOT NULL CHECK (capability_version > 0),
    source_evidence_id uuid NOT NULL REFERENCES evidence_records(id) ON DELETE RESTRICT,
    release_id text NOT NULL REFERENCES definition_releases(release_id) ON DELETE RESTRICT,
    activity_id text NOT NULL,
    activity_version integer NOT NULL CHECK (activity_version > 0),
    activity_hash text NOT NULL CHECK (activity_hash ~ '^[0-9a-f]{64}$'),
    review_group_key text NOT NULL,
    due_at timestamptz NOT NULL,
    priority integer NOT NULL DEFAULT 0,
    reason text NOT NULL CHECK (reason IN ('first_review','maintenance','remediation','review_incomplete')),
    status text NOT NULL CHECK (status IN ('open','claimed','completed','replaced')),
    claimed_attempt_id uuid REFERENCES learning_attempts(id) ON DELETE RESTRICT,
    evaluation_batch_id uuid REFERENCES evaluation_batches(id) ON DELETE RESTRICT,
    policy_version integer NOT NULL CHECK (policy_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    replaced_at timestamptz,
    CHECK (updated_at >= created_at),
    CHECK ((status = 'claimed') = (claimed_attempt_id IS NOT NULL AND completed_at IS NULL AND replaced_at IS NULL)),
    CHECK ((status = 'completed') = (completed_at IS NOT NULL AND evaluation_batch_id IS NOT NULL)),
    CHECK ((status = 'replaced') = (replaced_at IS NOT NULL)),
    UNIQUE (learner_id, capability_id, capability_version, source_evidence_id, policy_version)
);

CREATE UNIQUE INDEX review_items_one_active_policy_idx
    ON review_items (learner_id, capability_id, capability_version, policy_version)
    WHERE status IN ('open','claimed');
CREATE INDEX review_items_due_idx
    ON review_items (learner_id, due_at, priority DESC)
    WHERE status IN ('open','claimed');
CREATE INDEX review_items_group_idx
    ON review_items (learner_id, review_group_key, status);

CREATE TABLE attempt_review_items (
    attempt_id uuid NOT NULL REFERENCES learning_attempts(id) ON DELETE RESTRICT,
    review_item_id uuid NOT NULL UNIQUE REFERENCES review_items(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (attempt_id, review_item_id)
);

ALTER TABLE learning_outbox
    ADD COLUMN consumer text,
    ADD COLUMN consumer_version integer,
    ADD COLUMN last_error text;

CREATE INDEX learning_outbox_topic_status_available_idx
    ON learning_outbox (topic, status, available_at, created_at, id);
