CREATE TABLE definition_releases (
    release_id text PRIMARY KEY,
    schema_version integer NOT NULL CHECK (schema_version > 0),
    manifest jsonb NOT NULL CHECK (jsonb_typeof(manifest) = 'object'),
    bundle_hash text NOT NULL CHECK (bundle_hash ~ '^[0-9a-f]{64}$'),
    status text NOT NULL CHECK (status IN ('draft', 'current', 'superseded', 'withdrawn')),
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    CHECK ((status = 'draft' AND published_at IS NULL) OR (status <> 'draft' AND published_at IS NOT NULL))
);

CREATE UNIQUE INDEX definition_releases_one_current_idx
    ON definition_releases ((status))
    WHERE status = 'current';

CREATE TABLE definition_versions (
    kind text NOT NULL CHECK (kind IN ('capability', 'activity', 'task')),
    definition_id text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    content_hash text NOT NULL CHECK (content_hash ~ '^[0-9a-f]{64}$'),
    bundle_hash text NOT NULL CHECK (bundle_hash ~ '^[0-9a-f]{64}$'),
    release_id text NOT NULL REFERENCES definition_releases(release_id) ON DELETE RESTRICT,
    definition jsonb NOT NULL CHECK (jsonb_typeof(definition) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (kind, definition_id, version),
    UNIQUE (release_id, kind, definition_id, version)
);

CREATE INDEX definition_versions_release_idx ON definition_versions (release_id);

CREATE TABLE learners (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE learner_sessions (
    id uuid PRIMARY KEY,
    learner_id uuid NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE CHECK (token_hash ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    last_used_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CHECK (expires_at > created_at),
    CHECK (last_used_at >= created_at)
);

CREATE INDEX learner_sessions_learner_idx ON learner_sessions (learner_id);
CREATE INDEX learner_sessions_active_expiry_idx
    ON learner_sessions (expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE learning_attempts (
    id uuid PRIMARY KEY,
    learner_id uuid NOT NULL REFERENCES learners(id) ON DELETE RESTRICT,
    release_id text NOT NULL REFERENCES definition_releases(release_id) ON DELETE RESTRICT,
    activity_id text NOT NULL,
    activity_version integer NOT NULL CHECK (activity_version > 0),
    activity_hash text NOT NULL CHECK (activity_hash ~ '^[0-9a-f]{64}$'),
    task_id text NOT NULL,
    task_version integer NOT NULL CHECK (task_version > 0),
    task_hash text NOT NULL CHECK (task_hash ~ '^[0-9a-f]{64}$'),
    capability_refs jsonb NOT NULL CHECK (jsonb_typeof(capability_refs) = 'array'),
    mode text NOT NULL CHECK (mode IN ('guided', 'practice', 'assessment', 'review')),
    status text NOT NULL CHECK (status IN ('active', 'submitted', 'submit_infra_failed', 'completed')),
    workspace jsonb NOT NULL CHECK (jsonb_typeof(workspace) = 'object'),
    workspace_revision bigint NOT NULL DEFAULT 0 CHECK (workspace_revision >= 0),
    workspace_hash text NOT NULL CHECK (workspace_hash ~ '^[0-9a-f]{64}$'),
    started_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    submitted_at timestamptz,
    completed_at timestamptz,
    CHECK (updated_at >= started_at),
    CHECK (submitted_at IS NULL OR submitted_at >= started_at),
    CHECK (completed_at IS NULL OR completed_at >= started_at),
    CHECK (status = 'active' OR submitted_at IS NOT NULL),
    CHECK (status <> 'completed' OR completed_at IS NOT NULL)
);

CREATE INDEX learning_attempts_learner_status_idx
    ON learning_attempts (learner_id, status, updated_at DESC);
CREATE INDEX learning_attempts_definition_idx
    ON learning_attempts (release_id, activity_id, activity_version);
