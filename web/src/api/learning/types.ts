export type APIVersion = 'v1'

export interface LearningErrorDetail {
  code: string
  message: string
}

export interface LearningErrorResponse {
  error: LearningErrorDetail
  current_revision?: number
  current_hash?: string
  execution_id?: string
  submission_id?: string
  event_id?: string
}

export interface SessionResponse {
  api_version: APIVersion
  learner: { id: string }
  session: { expires_at: string }
}

export interface VersionedDefinitionRef {
  id: string
  version: number
}

export interface AssistancePolicy {
  hints: boolean
  references: boolean
  solution: boolean
  ai_declaration: boolean
}

export interface Activity {
  id: string
  version: number
  content_hash: string
  rule_set_hash: string
  title: string
  kind: string
  mode: string
  capability_refs: VersionedDefinitionRef[]
  task_ref: VersionedDefinitionRef
  content_ref?: string
  assistance_policy: AssistancePolicy
}

export interface ActivityResponse {
  api_version: APIVersion
  release_id: string
  activity: Activity
  task: Task
}

export interface TaskHintSummary {
  id: string
  level: number
  title: string
}

export interface Task {
  id: string
  version: number
  content_hash: string
  bundle_hash: string
  language: string
  workspace_root: string
  editable_paths: string[]
  readonly_paths: string[]
  visible_tests: string[]
  allowed_actions: ExecutionAction[]
  hints: TaskHintSummary[]
  solution?: string
  limits: {
    max_files: number
    max_file_bytes: number
    max_total_bytes: number
  }
  readme: string
}

export interface RequiredEvidence {
  type: string
  independence: string
  context: string
  rule_ids: string[]
}

export interface Capability {
  id: string
  version: number
  content_hash: string
  name: string
  description: string
  milestone: string
  domain: string
  prerequisites: {
    hard: VersionedDefinitionRef[]
    recommended: VersionedDefinitionRef[]
  }
  required_evidence: RequiredEvidence[]
  review_policy: {
    first_review_after_days: number
    success_interval_days: number
    failure_remediation_after_days: number
  }
  resource_refs: string[]
}

export type AcquisitionState = 'not_started' | 'exploring' | 'practiced' | 'verified' | 'stable'
export type IndependenceState = 'unverified' | 'guided' | 'ai_assisted' | 'hinted' | 'referenced' | 'independent'
export type TransferState = 'unverified' | 'same_context' | 'variant' | 'new_project'
export type RetentionState = 'fresh' | 'due' | 'rusty'

export interface CapabilitySnapshot {
  learner_id: string
  capability_id: string
  capability_version: number
  capability_hash: string
  projection_version: number
  projected_at: string
  acquisition_state: AcquisitionState
  independence_state: IndependenceState
  transfer_state: TransferState
  retention_base_state: 'fresh' | 'rusty'
  retention_state: RetentionState
  last_evidence_at?: string
  last_independent_at?: string
  next_review_at?: string
}

export interface EvidenceSummary {
  id: string
  evaluation_batch_id: string
  attempt_id: string
  activity_id: string
  activity_mode: string
  rule_id: string
  evidence_type: string
  result: RuleStatus
  independence: IndependenceState
  context: TransferState
  reason: string
  occurred_at: string
}

export interface CapabilityResponse {
  api_version: APIVersion
  release_id: string
  capability: Capability
  snapshot: CapabilitySnapshot | null
  recent_evidence: EvidenceSummary[]
  source: {
    definition: 'release_bundle'
    snapshot: 'capability_snapshots'
    evidence: 'evidence_records'
    retention: 'derived_at_read'
    as_of: string
    clock: 'server'
  }
}

export interface ReviewItem {
  id: string
  release_id: string
  capability_id: string
  capability_version: number
  group_key: string
  due_at: string
  priority: number
  reason: string
  status: 'open' | 'claimed'
  claimed_attempt_id?: string
}

export interface PrerequisiteStatus {
  id: string
  required_version: number
  satisfied: boolean
  satisfied_version?: number
}

export interface NextRecommendation {
  kind: 'acquisition' | 'review'
  reason: 'acquisition_path' | 'continue_attempt' | 'due_review' | 'claimed_review'
  activity: Activity
  target_capability?: VersionedDefinitionRef
  review_item?: ReviewItem
  open_attempt?: {
    id: string
    release_id: string
    status: 'active' | 'submitted' | 'submit_infra_failed'
    updated_at: string
  }
  hard_prerequisites: PrerequisiteStatus[]
  recommended_prerequisites: PrerequisiteStatus[]
}

export interface NextResponse {
  api_version: APIVersion
  recommendation: NextRecommendation | null
  source: {
    release_id: string
    state: 'server_learning_state'
    as_of: string
    clock: 'server' | 'test_override'
  }
}

export type AttemptStatus = 'active' | 'submitted' | 'submit_infra_failed' | 'completed'
export type ExecutionAction = 'build' | 'test' | 'vet' | 'submit'
export type ExecutionStatus = 'queued' | 'running' | 'succeeded' | 'user_failed' | 'infra_failed'
export type RuleStatus = 'passed' | 'failed' | 'not_evaluated'
export type ExecutionStage = 'build' | 'visible_test' | 'vet' | 'held_out_test' | 'ast' | 'explanation'

export interface TestEvent {
  action: string
  package: string
  test?: string
  elapsed?: number
}

export interface ExecutionStageResult {
  stage: ExecutionStage
  status: 'passed' | 'failed'
  exit_code: number
  stdout?: string
  stderr?: string
  duration_ms: number
  timed_out: boolean
  output_truncated: boolean
  public_summary?: string
  test_events?: TestEvent[]
}

export interface Execution {
  api_version: APIVersion
  id: string
  attempt_id: string
  submission_id?: string
  action: ExecutionAction
  sequence: number
  status: ExecutionStatus
  workspace_revision: number
  workspace_hash: string
  stages: ExecutionStageResult[]
  failure?: { code: string; message: string }
  started_at?: string
  finished_at?: string
  created_at: string
  updated_at: string
}

export interface Submission {
  id: string
  attempt_id?: string
  submission_key?: string
  workspace_revision: number
  workspace_hash: string
  rule_set_hash: string
  assistance_cutoff_seq: number
  explanation: string
  status: 'frozen' | 'executing' | 'evaluated' | 'infra_failed'
  latest_execution_id: string
  latest_execution_sequence: number
  latest_execution_status: ExecutionStatus
  created_at: string
  evaluated_at?: string
}

export interface RuleResult {
  rule_id: string
  status: RuleStatus
  stage: ExecutionStage
  package?: string
  test?: string
  analyzer?: string
  summary: string
  execution_id: string
}

export interface Evidence {
  id: string
  evaluation_batch_id: string
  capability_id: string
  capability_version: number
  evidence_rule_id: string
  evidence_type: string
  result: RuleStatus
  independence: Exclude<IndependenceState, 'unverified'>
  context_level: string
  evaluator: string
  reason: string
  occurred_at: string
}

export interface AttemptResponse {
  api_version: APIVersion
  id: string
  release_id: string
  activity_id: string
  activity_version: number
  activity_hash: string
  task_id: string
  task_version: number
  task_hash: string
  capability_refs: VersionedDefinitionRef[]
  mode: string
  status: AttemptStatus
  workspace: Record<string, string>
  workspace_revision: number
  workspace_hash: string
  assistance: {
    level: Exclude<IndependenceState, 'unverified'>
    events: AssistanceEvent[]
  }
  submission?: Submission
  executions: Execution[]
  rule_results: RuleResult[]
  evidence: Evidence[]
}

export interface CreateAttemptRequest {
  activity_id: string
  activity_version: number
}

export interface SaveWorkspaceRequest {
  base_revision: number
  files: Record<string, string>
}

export interface ExecuteAttemptRequest {
  request_key: string
  action: Exclude<ExecutionAction, 'submit'>
  workspace_revision: number
  workspace_hash: string
}

export interface SubmitAttemptRequest {
  submission_key: string
  workspace_revision: number
  workspace_hash: string
  explanation: string
}

export interface SubmissionResponse {
  api_version: APIVersion
  submission: Submission
  execution: {
    id: string
    sequence: number
    status: ExecutionStatus
  }
}

export type AssistanceEventType = 'reference_opened' | 'solution_viewed' | 'ai_declared'

export interface AssistanceEvent {
  id: string
  attempt_id: string
  event_key: string
  event_seq: number
  event_type: AssistanceEventType | 'hint_revealed'
  payload: unknown
  created_at: string
}

export interface AssistanceEventResponse {
  api_version: APIVersion
  event: AssistanceEvent
}

export interface HintResponse extends AssistanceEventResponse {
  hint: { id: string; title: string; body: string }
}
